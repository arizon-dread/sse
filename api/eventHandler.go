package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/arizon-dread/sse/internal/helpers"
	"github.com/arizon-dread/sse/internal/model"
)

var recipients = make(map[string]chan string, 0)

func Events(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipient := r.PathValue("recipient")
	if err := helpers.Register(recipient, recipients); err != nil {
		log.Printf("error registering client: '%v', error: %v", recipient, err)
		w.WriteHeader(404)
		w.Write([]byte("no client supplied"))
	}

	defer func() {
		if _, exists := recipients[recipient]; exists {
			log.Printf("unregistering client %v", recipient)
			w.Write([]byte("event: unregistering client\n\n"))
			close(recipients[recipient])
			delete(recipients, recipient)
			ctx.Done()
		}

	}()
	log.Printf("client %v is waiting for messages", recipient)
	for {
		select {
		case <-ctx.Done():
			return
		case res, ok := <-recipients[recipient]:
			if ok {
				fmt.Fprintf(w, "data: %s\n\n", res)
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			} else {
				ctx.Done()
				return
			}

		}

	}

}

func ForwardMsg(w http.ResponseWriter, r *http.Request) {
	var msg model.Message
	reqBody, err := io.ReadAll(r.Body)
	//no body sent in post, return bad request
	if err != nil {
		http.Error(w, "body unreadable", http.StatusBadRequest)
		return
	}
	//body doesn't match model, return 400
	err = json.Unmarshal(reqBody, &msg)
	if err != nil {
		http.Error(w, "unable to unmarshal message", http.StatusBadRequest)
		return
	}
	log.Printf("received message for %v", msg.Recipient)
	if err = helpers.Register(msg.Recipient, recipients); err != nil {
		returnConflict(w)
		return
	}
	if ch, exists := recipients[msg.Recipient]; exists {
		log.Printf("forwarding message to %v", msg.Recipient)
		select {
		case ch <- msg.Message:
		default:
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("cache is full, no client is listening."))
		}

	} else {
		returnConflict(w)
	}

}

func returnConflict(w http.ResponseWriter) {
	w.WriteHeader(http.StatusConflict)
	w.Write([]byte("supplied client is not registered."))
}
