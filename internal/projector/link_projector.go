package projector

import (
	"context"
	"encoding/json"
	"url-shortener/internal/domain"
	"url-shortener/internal/repository"

	"go.uber.org/zap"
)

type LinkProjector struct {
	repo   *repository.ReadRepository
	logger *zap.Logger
}

func NewLinkProjector(r *repository.ReadRepository, log *zap.Logger) *LinkProjector {
	return &LinkProjector{
		repo:   r,
		logger: log,
	}
}

func (p *LinkProjector) HandleClick(data []byte) {
	var e domain.LinkClickedEvent

	err := json.Unmarshal(data, &e)
	if err != nil {
		return
	}
	err = p.repo.IncClicks(context.Background(), e.ShortCode)
	if err != nil {
		p.logger.Warn("Failed to increment click counter", zap.Error(err))
		return
	}
	p.logger.Info("click event processed", zap.String("short", e.ShortCode))
}
