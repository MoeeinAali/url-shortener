package main

import (
	"url-shortener/internal/config"
	"url-shortener/internal/db"
	"url-shortener/internal/handlers"
	"url-shortener/internal/logger"
	"url-shortener/internal/messaging"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	err = db.RunMigrations(pg)
	if err != nil {
		log.Fatal("migration failed", zap.Error(err))
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

	commandHandler := handlers.NewCommandHandler(commandService)
	queryHandler := handlers.NewQueryHandler(queryService)

	r := gin.Default()

	// command routes
	r.POST("/links", commandHandler.CreateLink)
	r.POST("/links/:short/disable", commandHandler.DisableLink)

	// query routes
	r.GET("/:short", queryHandler.Redirect)
	r.GET("/links/:short/stats", queryHandler.Stats)

	err = r.Run(":8080")
	if err != nil {
		panic(err)
		return
	}
}
