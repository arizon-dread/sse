package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arizon-dread/sse/internal/config"
	"github.com/arizon-dread/sse/internal/helpers"
)

type MsgHandler interface {
	GetName() string
	GetChannel() chan string
	Send(msg string) error
	Receive(context.Context, chan string) error
	Exists() bool
	Unregister()
}

var recipients = make(map[string]chan string, 0)

func Register(rcpt string) (MsgHandler, error) {

	if rcpt == "" {
		return nil, fmt.Errorf("no recipient supplied to register")
	}
	cfg := config.Get()

	if len(cfg.Cache.Url) > 0 {
		return CacheMsgHandler{Name: rcpt, Ch: registerCacheRcpt(rcpt)}, nil
	} else {
		return InMemMsgHandler{Name: rcpt, Ch: registerMemRcpt(rcpt)}, nil
	}
}

func registerMemRcpt(rcpt string) chan string {
	if _, exists := recipients[rcpt]; !exists {
		recipients[rcpt] = make(chan string, 10)
		log.Printf("Registered %v\n", rcpt)
		return recipients[rcpt]
	}
	return recipients[rcpt]
}

func registerCacheRcpt(rcpt string) chan string {
	rdb, err := helpers.GetCacheConn()
	resp := rdb.Get(context.Background(), "receiver-"+rcpt)
	res, _ := resp.Result()
	if res == "" {
		rdb.Set(context.Background(), "receiver-"+rcpt, "$", time.Hour*24)

	}
	ch := make(chan string, 10)
	log.Printf("Registered %v\n", rcpt)
	if err != nil {
		log.Printf("failed getting cache connection")
	}

	log.Printf("Starting to read from %v\n", res)
	return ch
}
