package handlers

import (
	"fmt"
	"log"

	"github.com/arizon-dread/sse/internal/config"
)

type MsgHandler interface {
	Pub(msg string) error
	Sub(chan string) error
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
		return CacheMsgHandler{Name: rcpt, Ch: registerCacheMemRcpt(rcpt)}, nil
	} else {
		return InMemMsgHandler{Name: rcpt, Ch: registerMemRcpt(rcpt)}, nil
	}
}

func registerMemRcpt(rcpt string) chan string {
	if _, exists := recipients[rcpt]; !exists {
		recipients[rcpt] = make(chan string, 10)
		log.Printf("Registered %v", rcpt)
		return recipients[rcpt]
	}
	return recipients[rcpt]
}

func registerCacheMemRcpt(rcpt string) chan string {

}
