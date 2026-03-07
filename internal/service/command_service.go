package service

import (
	"context"
	"url-shortener/internal/domain"
	"url-shortener/internal/messaging"
	"url-shortener/internal/repository"
	"url-shortener/pkg/shortener"

	"go.uber.org/zap"
)

type CommandService struct {
	repo   *repository.LinkRepository
	logger *zap.Logger
	nats   *messaging.Client
}

func NewCommandService(repo *repository.LinkRepository, n *messaging.Client, logger *zap.Logger) *CommandService {
	return &CommandService{
		repo:   repo,
		logger: logger,
		nats:   n,
	}
}

func (s *CommandService) CreateLink(ctx context.Context, url string) (*domain.Link, error) {
	short := shortener.Generate()
	link := domain.NewLink(short, url)

	err := s.repo.Create(ctx, link)
	if err != nil {
		s.logger.Error("create link failed", zap.Error(err))
	}

	s.logger.Info("create link successfully", zap.String("short", short))

	return link, nil
}

func (s *CommandService) DisableLink(cta context.Context, short string) error {
	return s.repo.Disable(cta, short)
}
