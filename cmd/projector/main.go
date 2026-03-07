package main

import (
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

	redis := db.NewRedis(cfg.RedisAddr)

	nats, err := messaging.New(cfg.NatsURL)
	if err != nil {
		panic(err)
	}

	readRepo := repository.NewReadRepository(redis)

	p := projector.NewLinkProjector(readRepo, log)

	nats.Subscribe("link.clicked", p.HandleClick)

	select {}
}
