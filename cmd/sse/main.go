package main

import (
	"log"
	"net/http"

	"github.com/arizon-dread/sse/api"
	"github.com/arizon-dread/sse/pkg/http/middlewares/headers"
)

func main() {
	log.Println("Starting simple Server-sent Events lab")
	mux := http.NewServeMux()
	events := http.HandlerFunc(api.Events)
	mux.Handle("GET /events/{recipient}", headers.SseHeadersMiddleware(events))
	mux.HandleFunc("POST /message", api.ForwardMsg)
	mux.HandleFunc("POST /register/{recipient}", api.Register)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
