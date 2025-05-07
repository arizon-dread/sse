package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

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
	go func() {
		err = handler.Receive(ctx, subChan, cancel)
		if err != nil {
			if !strings.Contains(err.Error(), "redis: nil") {
				cancel()
				log.Printf("error encountered when starting to receive from the backing data structure, %v\n", err)
			}
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

				matched, err := regexp.MatchString(`^[a-zA-Z]+: .*\n{2}$`, res)
				if err != nil {
					log.Printf("error matching message: %v", err)
					continue
				}
				if matched {
					_, err := fmt.Fprintf(w, "%s", res)
					log.Printf(`sent raw string to client: "%s"`, res)
					if err != nil {
						log.Printf("failed to write %s to client, err: %v", res, err)
					}
				} else {
					fmt.Fprintf(w, "data: %s\n\n", res)
					log.Printf(`sent prepended data string to client: "data: %s"`, res)
				}
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
				handler.SetLastRead(time.Now())
			} else {
				ctx.Done()
				cancel()
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

func LastRead(w http.ResponseWriter, r *http.Request) {
	recipient := r.PathValue("recipient")
	handler, err := handlers.Register(recipient)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("recipient %v not found", recipient)))
		return
	}
	dt := handler.GetLastRead()
	if dt == nil {
		w.WriteHeader(http.StatusTooEarly)
		w.Write([]byte(fmt.Sprintf("recipient has not read any messages yet")))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(string(dt.Format("2006-01-02T15:04:05"))))
}
