package helpers

import (
	"fmt"
	"log"
)

func Register(recipient string, recipients map[string]chan string) error {

	if recipient == "" {
		return fmt.Errorf("no recipient supplied to register")
	}
	if _, exists := recipients[recipient]; !exists {
		recipients[recipient] = make(chan string, 2)
		log.Printf("Registered %v", recipient)
	}

	return nil
}
