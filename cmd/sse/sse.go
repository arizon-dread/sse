package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arizon-dread/sse/api"
	"github.com/arizon-dread/sse/internal/config"
	"github.com/arizon-dread/sse/internal/model"
	"github.com/arizon-dread/sse/pkg/http/middlewares/headers"
	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/go-fuego/fuego/param"
)

func main() {
	cfg := config.Get()
	backingCacheText := "instance in-memory cache, not suitable for multi instance use"
	if len(cfg.Cache.Url) > 0 {
		backingCacheText = "valkey/redis cache and stream"
	}
	log.Printf("Starting Server-sent Events micro service api with %v\n on port %v", backingCacheText, cfg.ApiPort)
	s := fuego.NewServer(
		fuego.WithAddr(fmt.Sprintf("localhost:%v", cfg.ApiPort)),
	)

	//mux := http.NewServeMux()
	events := http.HandlerFunc(api.Events)
	headersMiddleware := fuego.Group(s, "/events")
	fuego.Use(headersMiddleware, headers.SseHeadersMiddleware)
	fuego.GetStd(headersMiddleware, "/{recipient}", events,
		option.Summary("Stream events for a recipient"),
		option.AddResponse(200, "text/event-stream", fuego.Response{
			Type:         "",
			ContentTypes: []string{"text/event-stream"},
		}))
	fuego.GetStd(s, "/lastRead/{recipient}", api.LastRead,
		option.Summary("Get last read message for a recipient"),
		option.Query("recipient", "The identifier for the recipient", param.Required()),

		option.AddResponse(200, "text/plain", fuego.Response{
			Type:         "",
			ContentTypes: []string{"text/plain"},
		}))

	msgBody := fuego.RequestBody{
		Type:         model.Message{},
		ContentTypes: []string{"application/json"},
	}
	fuego.PostStd(s, "/message", api.ForwardMsg,
		option.Summary("Send a message to a recipient"),
		option.RequestBody(msgBody),
		option.AddResponse(200, "text/plain", fuego.Response{
			Type:         "",
			ContentTypes: []string{"text/plain"},
		}),
	)
	//mux.Handle("GET /events/{recipient}", headers.SseHeadersMiddleware(events))
	//mux.HandleFunc("GET /lastRead/{recipient}", api.LastRead)
	//mux.HandleFunc("POST /message", api.ForwardMsg)
	//log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", cfg.ApiPort), mux))
	log.Fatal(s.Run())
}
