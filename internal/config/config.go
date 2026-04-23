package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Mode                 string
	LogLevel             slog.Level
	HTTPPort             int
	EventHubConnString   string
	EventHubName         string
	ConsumerGroup        string
	EventHubEventsFile   string
	CheckpointInterval   time.Duration
	MockTickInterval     time.Duration
	ReceiveErrorBackoff  time.Duration
	DedupWindow          time.Duration
	DedupMaxEntries      int
}

func Load() (Config, error) {
	cfg := Config{
		Mode:                getEnv("WORKER_MODE", "mock"),
		ConsumerGroup:       getEnv("EVENTHUB_CONSUMER_GROUP", "$Default"),
		CheckpointInterval:  getEnvDuration("CHECKPOINT_INTERVAL", 30*time.Second),
		MockTickInterval:    getEnvDuration("MOCK_TICK_INTERVAL", 5*time.Second),
		ReceiveErrorBackoff: getEnvDuration("RECEIVE_ERROR_BACKOFF", 250*time.Millisecond),
		DedupWindow:         getEnvDuration("DEDUP_WINDOW", 10*time.Minute),
		DedupMaxEntries:     getEnvInt("DEDUP_MAX_ENTRIES", 10000),
	}

	level, err := parseLogLevel(getEnv("LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}
	cfg.LogLevel = level

	port, err := strconv.Atoi(getEnv("HTTP_PORT", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}
	cfg.HTTPPort = port

	cfg.EventHubConnString = os.Getenv("EVENTHUB_CONNECTION_STRING")
	cfg.EventHubName = os.Getenv("EVENTHUB_NAME")
	cfg.EventHubEventsFile = os.Getenv("EVENTHUB_EVENTS_FILE")

	if cfg.Mode == "eventhub" {
		if cfg.EventHubConnString == "" {
			return Config{}, fmt.Errorf("EVENTHUB_CONNECTION_STRING is required in eventhub mode")
		}
		if cfg.EventHubName == "" {
			return Config{}, fmt.Errorf("EVENTHUB_NAME is required in eventhub mode")
		}
		if cfg.EventHubEventsFile == "" {
			return Config{}, fmt.Errorf("EVENTHUB_EVENTS_FILE is required in eventhub mode in this environment")
		}
	}

	if cfg.Mode != "eventhub" && cfg.Mode != "mock" {
		return Config{}, fmt.Errorf("WORKER_MODE must be 'eventhub' or 'mock', got: %s", cfg.Mode)
	}

	return cfg, nil
}

func parseLogLevel(raw string) (slog.Level, error) {
	switch raw {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid LOG_LEVEL: %s (supported: debug|info|warn|error)", raw)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
