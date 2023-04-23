package sse

import "log"

// Client represents a web browser connection.
type Client struct {
	lastEventID,
	name string
	channel string
	send    chan *Message
}

func newClient(lastEventID, name string, channel string) *Client {
	return &Client{
		lastEventID,
		name,
		channel,
		make(chan *Message),
	}
}

// SendMessage sends a message to client.
func (c *Client) SendMessage(message *Message) {
	c.lastEventID = message.id
	c.send <- message
}

// Disconnect disconnects client.
func (c *Client) Disconnect() {
	log.Println("sse.Client.Disconnect: not implemented")
}

// Name returns client's name.
func (c *Client) Name() string {
	return c.name
}

// Channel returns the channel where this client is subscribed to.
func (c *Client) Channel() string {
	return c.channel
}

// LastEventID returns the ID of the last message sent.
func (c *Client) LastEventID() string {
	return c.lastEventID
}
