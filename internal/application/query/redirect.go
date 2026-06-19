// Package query holds the read-side use cases (the "Q" in CQRS). Queries never
// mutate aggregate state; they serve from the fast read model.
package query

import (
	"context"

	"go.uber.org/zap"

	"url-shortener/internal/application/command"
	"url-shortener/internal/application/port"
)

// Redirect is the query to resolve a short code to its destination.
type Redirect struct {
	ShortCode string
}

// RedirectHandler executes the Redirect use case. Resolving the URL is a pure
// read from the fast store; recording the click is a write-side side effect
// dispatched as a command. A failure to record a click never fails the redirect.
type RedirectHandler struct {
	read        port.ReadModel
	recordClick *command.RecordClickHandler
	log         *zap.Logger
}

// NewRedirectHandler wires the handler with its dependencies.
func NewRedirectHandler(read port.ReadModel, recordClick *command.RecordClickHandler, log *zap.Logger) *RedirectHandler {
	return &RedirectHandler{read: read, recordClick: recordClick, log: log}
}

// Handle resolves the destination and records a click (best effort).
func (h *RedirectHandler) Handle(ctx context.Context, q Redirect) (string, error) {
	longURL, err := h.read.LongURL(ctx, q.ShortCode)
	if err != nil {
		return "", err
	}

	if err := h.recordClick.Handle(ctx, command.RecordClick{ShortCode: q.ShortCode}); err != nil {
		// Availability over a single lost click: log and still redirect.
		h.log.Warn("failed to record click", zap.Error(err), zap.String("short", q.ShortCode))
	}

	return longURL, nil
}
