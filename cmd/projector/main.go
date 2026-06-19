// Command projector consumes domain events from JetStream and maintains the read
// model (the Projector/Consumer of the CQRS architecture).
package main

import (
	"context"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"url-shortener/internal/infrastructure/config"
	"url-shortener/internal/infrastructure/logger"
	"url-shortener/internal/infrastructure/messaging/jetstream"
	"url-shortener/internal/infrastructure/persistence/postgres"
	"url-shortener/internal/infrastructure/persistence/redis"
	"url-shortener/internal/infrastructure/projector"
)

const durableConsumerName = "link-projector"

func main() {
	cfg := config.Load()

	log, err := logger.New()
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	db, err := postgres.Open(cfg.PostgresDSN)
	if err != nil {
		log.Fatal("connect postgres", zap.Error(err))
	}
	if err := postgres.AutoMigrate(db); err != nil {
		log.Fatal("migrate postgres", zap.Error(err))
	}

	rdb := redis.Open(cfg.RedisAddr, cfg.RedisPassword)
	defer func() { _ = rdb.Close() }()

	js, err := jetstream.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal("connect jetstream", zap.Error(err))
	}
	defer js.Close()
	if err := js.EnsureStream(); err != nil {
		log.Fatal("ensure stream", zap.Error(err))
	}

	readModel := redis.NewReadModel(rdb)
	analytics := postgres.NewAnalyticsStore(db)
	linkRepo := postgres.NewLinkRepository(db)

	p := projector.New(readModel, analytics, linkRepo, log)

	// Rebuild the read model from the durable write side before consuming.
	if err := p.Bootstrap(context.Background()); err != nil {
		log.Warn("read model bootstrap failed", zap.Error(err))
	}

	sub, err := js.Subscribe(durableConsumerName, p.Handle)
	if err != nil {
		log.Fatal("subscribe jetstream", zap.Error(err))
	}
	defer func() { _ = sub.Unsubscribe() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("projector started")
	<-ctx.Done()
	log.Info("projector stopped")
}
