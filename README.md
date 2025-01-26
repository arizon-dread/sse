# Server-sent Events api

## Introduction
This is an api that accepts messages on `POST` to `/message`. Clients can start listening at `/events/{client-id}`.  

The messages can either be simply handled by a single instance that uses buffered channels in memory to temporarily cache up to 10 messages while there's no consumer listening to events, or you can set a cache.url setting that points to a redis-like instance (redis of valkey or a similar project that satisfies the same api structure) to use the cache to store consumers and keep track on what messages they have received. The Streams api is used to funnel messages from the producing REST endpoint to the consuming SSE endpoint. Messages are cached inside redis/valkey streams if the consumer is disconnected.  

# Steps to run
- run `go run cmd/sse/main.go` to start the program
- register a client `curl -X POST localhost:8080/register/client`
- start subscribing `curl localhost:8080/events/client -vN` the `N` flag prevents curl from buffering the responses.
- send a message: `curl  http://localhost:8080/message -d '{"recipient": "client", "message": "hello world"}'` from a different terminal
- the message should be received in the terminal that is subscribing.
- the message is just a string and could be a base64 encoded value, a json payload (url-encoded or with proper escape characters in place), xml or just a plain string.