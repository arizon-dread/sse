package handlers

import (
	"context"
	"log"
	"time"
)

type InMemMsgHandler struct {
	Name string
	Ch   chan string
}

func (immh InMemMsgHandler) Send(msg string) error {
	recipients[immh.Name] <- msg
	return nil
}
func (immh InMemMsgHandler) Receive(ctx context.Context, ch chan string, cancel context.CancelFunc) error {
	if immh.Exists() {
		return nil
	}
	recipients[immh.Name] = ch
	return nil
}
func (immh InMemMsgHandler) GetLastRead() *time.Time {
	lastRead, ok := InMemRecipientsLastRead[immh.Name]
	if !ok {
		return nil
	}
	return &lastRead
}
func (immh InMemMsgHandler) SetLastRead(d time.Time) {
	now := time.Now()
	InMemRecipientsLastRead[immh.Name] = now
}
func (immh InMemMsgHandler) Exists() bool {
	if _, exists := recipients[immh.Name]; exists {
		return true
	}
	return false
}
func (immh InMemMsgHandler) Unregister() {
	log.Printf("unregistering client %v\n", immh.Name)
	delete(recipients, immh.Name)
	close(immh.Ch)
}

func (immh InMemMsgHandler) GetName() string {
	return immh.Name
}

func (immh InMemMsgHandler) GetChannel() chan string {
	return immh.Ch
}
