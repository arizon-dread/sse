package main

import (
	"log"
	"net/http"

	"github.com/arizon-dread/sse/api"
	"github.com/arizon-dread/sse/internal/config"
	"github.com/arizon-dread/sse/pkg/http/middlewares/headers"
)

func main() {
	cfg := config.Get()
	backingCacheText := "instance in-memory cache, not suitable for multi instance use"
	if len(cfg.Cache.Url) > 0 {
		backingCacheText = "valkey/redis cache and stream"
	}
	log.Printf("Starting Server-sent Events micro service api with %v\n", backingCacheText)
	mux := http.NewServeMux()
	events := http.HandlerFunc(api.Events)
	mux.Handle("GET /events/{recipient}", headers.SseHeadersMiddleware(events))
	mux.HandleFunc("POST /message", api.ForwardMsg)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
