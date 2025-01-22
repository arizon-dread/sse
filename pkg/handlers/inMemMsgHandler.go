package handlers

import "log"

type InMemMsgHandler struct {
	Name string
	Ch   chan string
}

func (immh InMemMsgHandler) Send(msg string) error {
	recipients[immh.Name] <- msg
	return nil
}
func (immh InMemMsgHandler) Receive(ch chan string) error {
	if immh.Exists() {
		return nil
	}
	recipients[immh.Name] = ch
	return nil
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
