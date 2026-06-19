// Command api runs the REST API: both the command (write) and query (read) sides
// of the CQRS system.
package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"url-shortener/internal/application/command"
	"url-shortener/internal/application/query"
	"url-shortener/internal/domain/link"
	"url-shortener/internal/infrastructure/config"
	"url-shortener/internal/infrastructure/logger"
	"url-shortener/internal/infrastructure/persistence/postgres"
	"url-shortener/internal/infrastructure/persistence/redis"
	transport "url-shortener/internal/interfaces/http"
)

func main() {
	cfg := config.Load()

	log, err := logger.New()
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	// Write store (Postgres) + migrations.
	db, err := postgres.Open(cfg.PostgresDSN)
	if err != nil {
		log.Fatal("connect postgres", zap.Error(err))
	}
	if err := postgres.AutoMigrate(db); err != nil {
		log.Fatal("migrate postgres", zap.Error(err))
	}

	// Read store (Redis).
	rdb := redis.Open(cfg.RedisAddr, cfg.RedisPassword)
	defer func() { _ = rdb.Close() }()

	// Adapters.
	linkRepo := postgres.NewLinkRepository(db)
	outbox := postgres.NewOutboxStore(db)
	readModel := redis.NewReadModel(rdb)

	// Use cases.
	createHandler := command.NewCreateLinkHandler(linkRepo, link.Generator{}, log)
	disableHandler := command.NewDisableLinkHandler(linkRepo, log)
	recordClickHandler := command.NewRecordClickHandler(outbox)
	redirectHandler := query.NewRedirectHandler(readModel, recordClickHandler, log)
	statsHandler := query.NewGetStatsHandler(readModel)

	// Transport.
	cmdHandler := transport.NewCommandHandler(createHandler, disableHandler, cfg.BaseURL)
	qryHandler := transport.NewQueryHandler(redirectHandler, statsHandler)
	router := transport.NewRouter(cmdHandler, qryHandler)

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: router}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("api listening", zap.String("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("http server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info("api shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
	}
}
