// Package projector consumes domain events from the bus and maintains the read
// model (eventual consistency). It is idempotent (safe under at-least-once
// delivery) and can rebuild the read model from the durable write side on boot.
package projector

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"url-shortener/internal/application/port"
	"url-shortener/internal/domain/link"
	"url-shortener/internal/infrastructure/event"
)

// Projector applies events to the read model.
type Projector struct {
	readModel port.ReadModelWriter
	analytics port.AnalyticsStore
	links     port.LinkReader
	log       *zap.Logger
}

// New constructs a projector.
func New(readModel port.ReadModelWriter, analytics port.AnalyticsStore, links port.LinkReader, log *zap.Logger) *Projector {
	return &Projector{readModel: readModel, analytics: analytics, links: links, log: log}
}

// Bootstrap rebuilds the read model from the durable write side. This is what
// makes the read store disposable: if Redis is wiped, the read model is fully
// reconstructed from Postgres on startup.
func (p *Projector) Bootstrap(ctx context.Context) error {
	links, err := p.links.ListAll(ctx)
	if err != nil {
		return err
	}
	for _, l := range links {
		if l.Disabled {
			if err := p.readModel.DeleteLink(ctx, l.ShortCode); err != nil {
				p.log.Warn("bootstrap: delete disabled link failed", zap.Error(err), zap.String("short", l.ShortCode))
			}
			continue
		}
		if err := p.readModel.SaveLink(ctx, l.ShortCode, l.LongURL); err != nil {
			p.log.Warn("bootstrap: save link failed", zap.Error(err), zap.String("short", l.ShortCode))
		}
	}

	counts, err := p.analytics.ListAll(ctx)
	if err != nil {
		return err
	}
	for _, c := range counts {
		if err := p.readModel.SetClicks(ctx, c.ShortCode, c.Count); err != nil {
			p.log.Warn("bootstrap: set clicks failed", zap.Error(err), zap.String("short", c.ShortCode))
		}
	}

	p.log.Info("read model bootstrap complete", zap.Int("links", len(links)), zap.Int("analytics", len(counts)))
	return nil
}

// Handle dispatches an envelope to the right projection. A returned error causes
// the bus to redeliver the message.
func (p *Projector) Handle(env event.Envelope) error {
	ctx := context.Background()

	var payload event.Payload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		p.log.Error("projector: bad payload", zap.Error(err), zap.String("event_id", env.EventID))
		return nil // poison payload: ack to avoid infinite redelivery
	}

	switch env.EventType {
	case link.EventTypeLinkCreated:
		return p.onCreated(ctx, payload)
	case link.EventTypeLinkDisabled:
		return p.onDisabled(ctx, payload)
	case link.EventTypeLinkClicked:
		return p.onClicked(ctx, env.EventID, payload)
	default:
		p.log.Warn("projector: unknown event type", zap.String("type", env.EventType))
		return nil
	}
}

// onCreated/onDisabled are naturally idempotent (SET/DEL), so no dedup is needed.
func (p *Projector) onCreated(ctx context.Context, payload event.Payload) error {
	if err := p.readModel.SaveLink(ctx, payload.ShortCode, payload.LongURL); err != nil {
		return err
	}
	p.log.Info("projected link.created", zap.String("short", payload.ShortCode))
	return nil
}

func (p *Projector) onDisabled(ctx context.Context, payload event.Payload) error {
	if err := p.readModel.DeleteLink(ctx, payload.ShortCode); err != nil {
		return err
	}
	p.log.Info("projected link.disabled", zap.String("short", payload.ShortCode))
	return nil
}

// onClicked applies the click to the durable counter exactly once (via the
// analytics store's dedup), then mirrors the authoritative total to Redis.
func (p *Projector) onClicked(ctx context.Context, eventID string, payload event.Payload) error {
	id, err := uuid.Parse(eventID)
	if err != nil {
		p.log.Error("projector: bad event id", zap.Error(err), zap.String("event_id", eventID))
		return nil
	}

	total, err := p.analytics.RecordClick(ctx, id, payload.ShortCode, time.Now().UTC())
	if err != nil {
		return err
	}
	if err := p.readModel.SetClicks(ctx, payload.ShortCode, total); err != nil {
		return err
	}
	p.log.Info("projected link.clicked", zap.String("short", payload.ShortCode), zap.Int64("total", total))
	return nil
}
