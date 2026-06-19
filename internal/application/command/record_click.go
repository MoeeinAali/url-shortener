package command

import (
	"context"

	"url-shortener/internal/application/port"
	"url-shortener/internal/domain/link"
)

// RecordClick is the command emitted on every redirect. It does not load the
// aggregate (the hot path stays cheap); it durably appends a LinkClicked event
// to the outbox so the click is never lost.
type RecordClick struct {
	ShortCode string
}

// RecordClickHandler executes the RecordClick use case.
type RecordClickHandler struct {
	outbox port.Outbox
}

// NewRecordClickHandler wires the handler with its dependencies.
func NewRecordClickHandler(outbox port.Outbox) *RecordClickHandler {
	return &RecordClickHandler{outbox: outbox}
}

// Handle appends a LinkClicked domain event to the outbox.
func (h *RecordClickHandler) Handle(ctx context.Context, cmd RecordClick) error {
	code, err := link.NewShortCode(cmd.ShortCode)
	if err != nil {
		return err
	}
	return h.outbox.Append(ctx, link.NewLinkClicked(code))
}
