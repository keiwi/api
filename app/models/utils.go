package models

import (
	"aahframework.org/aah.v0"
	"aahframework.org/log.v0"
	"github.com/nats-io/go-nats"
)

var (
	Conn *nats.Conn
)

type Response struct {
	// MessageJSON - json data for outputting
	Success bool        `json:"success"` // Wether an error occured or not
	Message string      `json:"message"` // The message
	Data    interface{} `json:"data"`    // Extra data, generally it will contain a struct
}

// TODO: Consider about being more specific about editing requests rather then going with this option.
type EditRequest struct {
	ID     string      `json:"id"`
	Option string      `json:"option"`
	Value  interface{} `json:"value"`
}

func ConnectNats(_ *aah.Event) {
	conn, err := nats.Connect(aah.AppConfig().StringDefault("nats.url", nats.DefaultURL))
	if err != nil {
		log.Fatalf("error when connecting to nats: %v", err)
	}

	Conn = conn
}

func DisconnectNats(_ *aah.Event) {
	Conn.Close()
}
