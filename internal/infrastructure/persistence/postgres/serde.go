package postgres

import (
	"encoding/json"
	"fmt"

	"url-shortener/internal/domain/link"
	"url-shortener/internal/domain/shared"
)

// toOutboxRow converts a domain event into an outbox row. The payload type switch
// is an adapter concern: the domain stays free of serialization details.
func toOutboxRow(e shared.DomainEvent) (outboxModel, error) {
	var payload []byte
	var err error

	switch ev := e.(type) {
	case link.LinkCreated:
		payload, err = json.Marshal(eventPayload{ShortCode: ev.ShortCode, LongURL: ev.LongURL})
	case link.LinkDisabled:
		payload, err = json.Marshal(eventPayload{ShortCode: ev.ShortCode})
	case link.LinkClicked:
		payload, err = json.Marshal(eventPayload{ShortCode: ev.ShortCode})
	default:
		return outboxModel{}, fmt.Errorf("postgres: unknown domain event type %T", e)
	}
	if err != nil {
		return outboxModel{}, err
	}

	return outboxModel{
		ID:          e.EventID(),
		AggregateID: e.AggregateID(),
		EventType:   e.EventType(),
		Payload:     payload,
		OccurredAt:  e.OccurredAt(),
	}, nil
}

// eventPayload mirrors event.Payload but is kept local to avoid leaking the
// infrastructure/event package into the persistence serializer.
type eventPayload struct {
	ShortCode string `json:"short_code"`
	LongURL   string `json:"long_url,omitempty"`
}
