package projector

import (
	"context"
	"encoding/json"
	"url-shortener/internal/domain"
	"url-shortener/internal/repository"

	"go.uber.org/zap"
)

type LinkProjector struct {
	readRepo  *repository.ReadRepository
	writeRepo *repository.LinkRepository
	logger    *zap.Logger
}

func NewLinkProjector(readRepo *repository.ReadRepository, writeRepo *repository.LinkRepository, log *zap.Logger) *LinkProjector {
	return &LinkProjector{
		readRepo:  readRepo,
		writeRepo: writeRepo,
		logger:    log,
	}
}

func (p *LinkProjector) Bootstrap(ctx context.Context) error {
	links, err := p.writeRepo.ListAll(ctx)
	if err != nil {
		return err
	}

	for _, link := range links {
		if link.Disabled {
			err = p.readRepo.DeleteLink(ctx, link.ShortCode)
			if err != nil {
				p.logger.Warn("bootstrap delete disabled link failed", zap.Error(err), zap.String("short", link.ShortCode))
			}
			continue
		}

		err = p.readRepo.SaveLink(ctx, link.ShortCode, link.LongURL)
		if err != nil {
			p.logger.Warn("bootstrap save link failed", zap.Error(err), zap.String("short", link.ShortCode))
		}
	}

	p.logger.Info("bootstrap read model complete", zap.Int("links_total", len(links)))
	return nil
}

func (p *LinkProjector) HandleLinkCreated(data []byte) {
	var e domain.LinkCreatedEvent
	err := json.Unmarshal(data, &e)
	if err != nil {
		return
	}

	err = p.readRepo.SaveLink(context.Background(), e.ShortCode, e.LongURL)
	if err != nil {
		p.logger.Warn("failed to save link projection", zap.Error(err), zap.String("short", e.ShortCode))
		return
	}

	p.logger.Info("link.created event processed", zap.String("short", e.ShortCode))
}

func (p *LinkProjector) HandleLinkDisabled(data []byte) {
	var e domain.LinkDisabledEvent
	err := json.Unmarshal(data, &e)
	if err != nil {
		return
	}

	err = p.readRepo.DeleteLink(context.Background(), e.ShortCode)
	if err != nil {
		p.logger.Warn("failed to delete link projection", zap.Error(err), zap.String("short", e.ShortCode))
		return
	}

	p.logger.Info("link.disabled event processed", zap.String("short", e.ShortCode))
}

func (p *LinkProjector) HandleClick(data []byte) {
	var e domain.LinkClickedEvent

	err := json.Unmarshal(data, &e)
	if err != nil {
		return
	}
	err = p.readRepo.IncClicks(context.Background(), e.ShortCode)
	if err != nil {
		p.logger.Warn("Failed to increment click counter", zap.Error(err))
		return
	}
	p.logger.Info("click event processed", zap.String("short", e.ShortCode))
}
