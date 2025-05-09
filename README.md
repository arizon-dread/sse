# Server-sent Events api

## Introduction
This is an api that accepts messages on `POST` to `/message`. Clients can start listening at `/events/{client-id}`. You can get the timestamp of the last message that the client read at the `/lastRead/{client-id}` endpoint.  

The messages can either be simply handled by a single instance that uses buffered channels in memory to temporarily cache up to 10 messages while there's no consumer listening to events, or you can set a cache.url setting that points to a redis-like instance (redis of valkey or a similar project that satisfies the same api structure) to use the cache to store consumers and keep track on what messages they have received. The Streams api is used to funnel messages from the producing REST endpoint to the consuming SSE endpoint. Messages are cached inside redis/valkey streams if the consumer is disconnected.  

# Steps to run
- run `go run cmd/sse/sse.go` to start the program
- start subscribing `curl localhost:8080/events/client -vN` the `N` flag prevents curl from buffering the responses.
- send a message: `curl  http://localhost:8080/message -d '{"recipient": "client", "message": "hello world"}'` from a different terminal
- the message should be received in the terminal that is subscribing.
- the message is just a string and could be a base64 encoded value, a json payload (url-encoded or with proper escape characters in place), xml or just a plain string.

# Production runtime
The SSE api as a stand alone application caches the messages in channels inside the go runtime and is limited to a single instance and a small amount of cached messages. There is a configuration posibility where you can use a redis like cache and stream instance (that satisfies the API's for caching and streaming) to back the sse application, thus opening the posibility of a more robust and persistent message handling together with horizontal scaling posibilities. 

Any configuration or license considerations of the choice of the cache and stream implementation is out of scope for SSE unless it is related to the connection config.
