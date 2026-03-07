package main

import (
	"url-shortener/internal/config"
	"url-shortener/internal/db"
	"url-shortener/internal/logger"
	"url-shortener/internal/messaging"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
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

	writeRepo := repository.NewLinkRepository(pg)
	readRepo := repository.NewReadRepository(redis)

	commandService := service.NewCommandService(writeRepo, nats, log)
	queryService := service.NewQueryService(readRepo, nats, log)

	_ = commandService
	_ = queryService

	r := gin.Default()

	r.Run(":8080")
}
