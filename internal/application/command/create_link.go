// Package command holds the write-side use cases (the "C" in CQRS). Each command
// is a plain data struct and each handler orchestrates the domain to fulfil it.
package command

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"url-shortener/internal/domain/link"
)

// ErrShortCodeExhausted is returned when a unique short code could not be
// generated within the allowed number of attempts.
var ErrShortCodeExhausted = errors.New("could not generate a unique short code")

// ShortCodeGenerator is the port for producing short codes (satisfied by
// link.Generator). Injecting it keeps the handler deterministic in tests.
type ShortCodeGenerator interface {
	Generate() (link.ShortCode, error)
}

// CreateLink is the command to shorten a URL.
type CreateLink struct {
	URL string
}

// CreateLinkResult is the outcome of a successful CreateLink.
type CreateLinkResult struct {
	ShortCode string
	LongURL   string
}

// CreateLinkHandler executes the CreateLink use case.
type CreateLinkHandler struct {
	repo link.Repository
	gen  ShortCodeGenerator
	log  *zap.Logger
}

// NewCreateLinkHandler wires the handler with its dependencies.
func NewCreateLinkHandler(repo link.Repository, gen ShortCodeGenerator, log *zap.Logger) *CreateLinkHandler {
	return &CreateLinkHandler{repo: repo, gen: gen, log: log}
}

const maxShortCodeAttempts = 5

// Handle validates the URL, allocates a collision-free short code and persists a
// new Link aggregate (state + LinkCreated event) atomically via the repository.
func (h *CreateLinkHandler) Handle(ctx context.Context, cmd CreateLink) (CreateLinkResult, error) {
	longURL, err := link.NewURL(cmd.URL)
	if err != nil {
		return CreateLinkResult{}, err
	}

	for attempt := 0; attempt < maxShortCodeAttempts; attempt++ {
		code, err := h.gen.Generate()
		if err != nil {
			return CreateLinkResult{}, err
		}

		exists, err := h.repo.ExistsByShortCode(ctx, code)
		if err != nil {
			return CreateLinkResult{}, err
		}
		if exists {
			continue
		}

		aggregate, err := link.NewLink(code, longURL)
		if err != nil {
			return CreateLinkResult{}, err
		}
		if err := h.repo.Save(ctx, aggregate); err != nil {
			// A concurrent writer may have taken the code between the existence
			// check and the insert; retry with a fresh code.
			h.log.Warn("save link failed, retrying", zap.Error(err), zap.String("short", code.String()))
			continue
		}

		h.log.Info("link created", zap.String("short", code.String()))
		return CreateLinkResult{ShortCode: code.String(), LongURL: longURL.String()}, nil
	}

	return CreateLinkResult{}, ErrShortCodeExhausted
}
