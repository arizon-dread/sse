package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/arizon-dread/sse/internal/model"
)

var recipients = make(map[string]chan string, 0)

func Register(w http.ResponseWriter, r *http.Request) {
	recipient := r.PathValue("recipient")
	if recipient == "" {
		http.Error(w, "no recipient supplied", http.StatusBadRequest)
		return
	}
	if _, exists := recipients[recipient]; !exists {
		recipients[recipient] = make(chan string)
		log.Printf("Registered %v", recipient)
	}

}

func Events(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipient := r.PathValue("recipient")
	if recipient == "" {
		http.Error(w, "no recipient supplied", http.StatusBadRequest)
		return
	}
	log.Printf("client %v is waiting for messages", recipient)
	defer func() {
		if _, exists := recipients[recipient]; exists {
			log.Printf("unregistering client %v", recipient)
			close(recipients[recipient])
			delete(recipients, recipient)
		}

	}()
	for {
		select {
		case <-ctx.Done():
			return
		case res, ok := <-recipients[recipient]:
			if ok {
				fmt.Fprintf(w, "%s\n", res)
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
	if ch, exists := recipients[msg.Recipient]; exists {
		log.Printf("forwarding message to %v", msg.Recipient)
		ch <- msg.Message

	}
}
