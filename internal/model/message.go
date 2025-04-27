package model

type Message struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}
