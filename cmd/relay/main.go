// Command relay is the outbox publisher: it forwards committed domain events from
// the Postgres outbox to NATS JetStream (the publishing half of the Transactional
// Outbox pattern).
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
	"url-shortener/internal/infrastructure/relay"
)

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

	js, err := jetstream.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal("connect jetstream", zap.Error(err))
	}
	defer js.Close()
	if err := js.EnsureStream(); err != nil {
		log.Fatal("ensure stream", zap.Error(err))
	}

	outbox := postgres.NewOutboxStore(db)
	r := relay.New(outbox, js, cfg.RelayBatchSize, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("relay started", zap.Duration("interval", cfg.RelayPollInterval))
	r.Run(ctx, cfg.RelayPollInterval)
	log.Info("relay stopped")
}
