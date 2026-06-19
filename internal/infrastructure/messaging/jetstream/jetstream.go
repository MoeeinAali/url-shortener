// Package jetstream is the messaging adapter built on NATS JetStream. JetStream
// gives durable, replayable streams with at-least-once delivery and server-side
// de-duplication — the backbone of loss-free eventual consistency.
package jetstream

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"

	"url-shortener/internal/infrastructure/event"
)

const (
	streamName     = "LINKS"
	streamSubjects = "link.>"
	// dedupWindow is JetStream's de-duplication window keyed by Nats-Msg-Id.
	dedupWindow = 2 * time.Minute
)

// Conn wraps a NATS connection and its JetStream context.
type Conn struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

// connectAttempts bounds how long Connect waits for NATS to become reachable
// (so services can start before the broker in container orchestration).
const connectAttempts = 30

// Connect dials NATS and obtains a JetStream context with resilient reconnects.
// It retries the initial dial so startup ordering does not matter.
func Connect(url string) (*Conn, error) {
	var (
		nc  *nats.Conn
		err error
	)
	for attempt := 1; attempt <= connectAttempts; attempt++ {
		nc, err = nats.Connect(url,
			nats.MaxReconnects(-1),
			nats.ReconnectWait(time.Second),
		)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}
	return &Conn{nc: nc, js: js}, nil
}

// EnsureStream creates the durable, file-backed stream if it does not exist.
func (c *Conn) EnsureStream() error {
	if _, err := c.js.StreamInfo(streamName); err == nil {
		return nil
	}
	_, err := c.js.AddStream(&nats.StreamConfig{
		Name:       streamName,
		Subjects:   []string{streamSubjects},
		Storage:    nats.FileStorage,
		Retention:  nats.LimitsPolicy,
		Duplicates: dedupWindow,
	})
	return err
}

// Publish sends an envelope to the bus. The msgID enables JetStream dedup.
func (c *Conn) Publish(subject string, data []byte, msgID string) error {
	_, err := c.js.Publish(subject, data, nats.MsgId(msgID))
	return err
}

// Subscribe creates a durable, explicitly-acked push subscription over all link
// events. The handler runs per message; returning an error triggers a redelivery.
func (c *Conn) Subscribe(durable string, handler func(event.Envelope) error) (*nats.Subscription, error) {
	return c.js.Subscribe(streamSubjects, func(m *nats.Msg) {
		var env event.Envelope
		if err := json.Unmarshal(m.Data, &env); err != nil {
			// Poison message: terminate so it is not redelivered forever.
			_ = m.Term()
			return
		}
		if err := handler(env); err != nil {
			_ = m.Nak()
			return
		}
		_ = m.Ack()
	},
		nats.Durable(durable),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.DeliverAll(),
		nats.AckWait(30*time.Second),
	)
}

// Close drains and closes the connection.
func (c *Conn) Close() {
	if c.nc != nil {
		_ = c.nc.Drain()
	}
}
