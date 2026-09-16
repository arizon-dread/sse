FROM docker.io/golang:1.25-alpine AS go-mod-tidy
WORKDIR /go/src
RUN apk update && apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod tidy

FROM docker.io/golang:1.25-alpine AS build
WORKDIR /go/src
COPY --from=go-mod-tidy /go/src/* ./
COPY . ./ 
RUN apk update && apk add --no-cache git && mkdir -p /go/bin
RUN go build -v -o /go/bin ./...

FROM docker.io/alpine:3.23 AS final
RUN mkdir -p /go/bin
WORKDIR /go/bin
COPY --from=build /go/bin/sse .
ENTRYPOINT ["./sse"]
