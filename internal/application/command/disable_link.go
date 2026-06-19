package command

import (
	"context"

	"go.uber.org/zap"

	"url-shortener/internal/domain/link"
)

// DisableLink is the command to deactivate a link.
type DisableLink struct {
	ShortCode string
}

// DisableLinkHandler executes the DisableLink use case.
type DisableLinkHandler struct {
	repo link.Repository
	log  *zap.Logger
}

// NewDisableLinkHandler wires the handler with its dependencies.
func NewDisableLinkHandler(repo link.Repository, log *zap.Logger) *DisableLinkHandler {
	return &DisableLinkHandler{repo: repo, log: log}
}

// Handle loads the aggregate, applies the Disable behavior (which enforces the
// invariant and raises LinkDisabled) and persists state + event atomically.
func (h *DisableLinkHandler) Handle(ctx context.Context, cmd DisableLink) error {
	code, err := link.NewShortCode(cmd.ShortCode)
	if err != nil {
		return err
	}

	aggregate, err := h.repo.FindByShortCode(ctx, code)
	if err != nil {
		return err
	}

	if err := aggregate.Disable(); err != nil {
		return err
	}

	if err := h.repo.Save(ctx, aggregate); err != nil {
		return err
	}

	h.log.Info("link disabled", zap.String("short", cmd.ShortCode))
	return nil
}
