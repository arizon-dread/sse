package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/arizon-dread/sse/internal/model"
	"github.com/arizon-dread/sse/pkg/handlers"
	"golang.org/x/net/context"
)

func Events(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	recipient := r.PathValue("recipient")
	handler, err := handlers.Register(recipient)
	if err != nil {
		log.Printf("error registering client: '%v', error: %v", recipient, err)
		w.WriteHeader(404)
		w.Write([]byte("no client supplied"))
	}

	defer func() {
		handler.Unregister()

	}()
	subChan := handler.GetChannel()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = handler.Receive(ctx, subChan)
		if err != nil {
			cancel()
			log.Printf("error encountered when starting to receive from the backing data structure, %v\n", err)
		}
	}()
	log.Printf("client %v is waiting for messages", recipient)
	for {
		select {
		case <-ctx.Done():
			cancel()
			return
		case res, ok := <-subChan:
			if ok {
				fmt.Fprintf(w, "data: %s\n\n", res)
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			} else {
				ctx.Done()
				cancel()
				wg.Wait()
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
	handler, err := handlers.Register(msg.Recipient)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = handler.Send(msg.Message)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("cache is full, no client is listening."))
		return
	}
	w.WriteHeader(http.StatusCreated)

}

func returnConflict(w http.ResponseWriter) {
	w.WriteHeader(http.StatusConflict)
	w.Write([]byte("supplied client is not registered."))
}
