package client

import (
	"time"

	"github.com/chimera-rpg/go-common/network"
)

// Message is a container for a received network message.
type Message struct {
	Received time.Time
	Message  network.CommandMessage
}

// HandleMessageCommand received network.CommandMessage types and adds it to the client's message history.
func (c *Client) HandleMessageCommand(m network.CommandMessage) {
	c.MessageHistory = append(c.MessageHistory, Message{
		Received: time.Now(),
		Message:  m,
	})
}
