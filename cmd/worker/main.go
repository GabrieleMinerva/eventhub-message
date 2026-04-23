package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/go-worker-service/internal/config"
	"github.com/example/go-worker-service/internal/server"
	"github.com/example/go-worker-service/internal/source"
	"github.com/example/go-worker-service/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	var src worker.Source
	switch cfg.Mode {
	case "eventhub":
		src, err = source.NewEventHubSource(cfg.EventHubConnString, cfg.EventHubName, cfg.ConsumerGroup, cfg.CheckpointInterval)
		if err != nil {
			logger.Error("cannot initialize eventhub source", "error", err)
			os.Exit(1)
		}
	default:
		src = source.NewMockSource(cfg.MockTickInterval)
	}

	stats := &worker.Stats{}
	processor := worker.NewDedupProcessor(worker.NewLogProcessor(logger), cfg.DedupWindow, cfg.DedupMaxEntries)
	svc := worker.New(logger, src, processor, stats, cfg.ReceiveErrorBackoff)
	httpServer := server.New(cfg.HTTPPort, svc)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Error("http server failed", "error", err)
			stop()
		}
	}()

	if err := svc.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("worker failed", "error", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = svc.Shutdown(shutdownCtx)
	_ = httpServer.Shutdown(shutdownCtx)
}
