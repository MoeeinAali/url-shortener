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
		return nil, err
	}

	err = s.nats.Publish("link.created", domain.LinkCreatedEvent{
		ShortCode: short,
		LongURL:   url,
	})
	if err != nil {
		s.logger.Warn("publish link.created failed", zap.Error(err), zap.String("short", short))
	}

	s.logger.Info("create link successfully", zap.String("short", short))

	return link, nil
}

func (s *CommandService) DisableLink(cta context.Context, short string) error {
	err := s.repo.Disable(cta, short)
	if err != nil {
		return err
	}

	err = s.nats.Publish("link.disabled", domain.LinkDisabledEvent{ShortCode: short})
	if err != nil {
		s.logger.Warn("publish link.disabled failed", zap.Error(err), zap.String("short", short))
	}

	return nil
}
