package main

import (
	"context"
	"url-shortener/internal/config"
	"url-shortener/internal/db"
	"url-shortener/internal/logger"
	"url-shortener/internal/messaging"
	"url-shortener/internal/projector"
	"url-shortener/internal/repository"
)

func main() {
	cfg := config.Load()

	log, err := logger.New()
	if err != nil {
		panic(err)
	}

	pg, err := db.NewPostgres(cfg.PostgresDSN)
	if err != nil {
		panic(err)
	}

	redis := db.NewRedis(cfg.RedisAddr)

	nats, err := messaging.New(cfg.NatsURL)
	if err != nil {
		panic(err)
	}

	readRepo := repository.NewReadRepository(redis)
	writeRepo := repository.NewLinkRepository(pg)

	p := projector.NewLinkProjector(readRepo, writeRepo, log)

	err = p.Bootstrap(context.Background())
	if err != nil {
		log.Warn("bootstrap read model failed")
	}

	nats.Subscribe("link.created", p.HandleLinkCreated)
	nats.Subscribe("link.disabled", p.HandleLinkDisabled)
	nats.Subscribe("link.clicked", p.HandleClick)

	select {}
}
