# Serverside Event Streaming lab

# Steps to run
- run `go run cmd/sse/main.go` to start the program
- register a client `curl -X POST localhost:8080/register/client`
- start subscribing `curl localhost:8080/events/client -vN` the `N` flag prevents curl from buffering the responses.
- send a message: `curl  http://localhost:8080/message -d '{"recipient": "client", "message": "hello world"}'` from a different terminal
- the message should be received in the terminal that is subscribing.
