package query

import (
	"context"

	"url-shortener/internal/application/port"
)

// GetStats is the query for a link's click analytics.
type GetStats struct {
	ShortCode string
}

// StatsResult is the analytics view returned to the caller.
type StatsResult struct {
	ShortCode string
	Clicks    int64
}

// GetStatsHandler executes the GetStats use case.
type GetStatsHandler struct {
	read port.ReadModel
}

// NewGetStatsHandler wires the handler with its dependencies.
func NewGetStatsHandler(read port.ReadModel) *GetStatsHandler {
	return &GetStatsHandler{read: read}
}

// Handle reads the click counter from the fast read model.
func (h *GetStatsHandler) Handle(ctx context.Context, q GetStats) (StatsResult, error) {
	clicks, err := h.read.Clicks(ctx, q.ShortCode)
	if err != nil {
		return StatsResult{}, err
	}
	return StatsResult{ShortCode: q.ShortCode, Clicks: clicks}, nil
}
