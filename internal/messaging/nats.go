package messaging

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type Client struct {
	conn *nats.Conn
}

func New(url string) (*Client, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Publish(subject string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.conn.Publish(subject, data)
}

func (c *Client) Subscribe(subject string, handler func(data []byte)) error {
	_, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	return err
}
