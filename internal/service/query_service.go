package service

import (
	"context"
	"url-shortener/internal/domain"
	"url-shortener/internal/messaging"
	"url-shortener/internal/repository"

	"go.uber.org/zap"
)

type QueryService struct {
	repo   *repository.ReadRepository
	nats   *messaging.Client
	logger *zap.Logger
}

func NewQueryService(r *repository.ReadRepository, n *messaging.Client, log *zap.Logger) *QueryService {
	return &QueryService{
		repo:   r,
		nats:   n,
		logger: log,
	}
}

func (s *QueryService) Redirect(ctx context.Context, short string) (string, error) {
	url, err := s.repo.GetLink(ctx, short)
	if err != nil {
		return "", err
	}

	err = s.nats.Publish("link.clicked", domain.LinkClickedEvent{ShortCode: short})

	if err != nil {
		return "", err
	}

	return url, nil
}

func (s *QueryService) Stats(ctx context.Context, short string) (int64, error) {
	return s.repo.GetClicks(ctx, short)
}
