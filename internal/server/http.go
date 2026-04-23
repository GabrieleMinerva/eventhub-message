package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type StatsReader interface {
	Stats() (processed, failed, skipped uint64)
}

type Server struct {
	httpServer *http.Server
}

func New(port int, stats StatsReader) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		processed, failed, skipped := stats.Stats()
		status := http.StatusOK
		if failed > processed+skipped {
			status = http.StatusServiceUnavailable
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    map[bool]string{true: "degraded", false: "ready"}[status != http.StatusOK],
			"processed": processed,
			"failed":    failed,
			"skipped":   skipped,
		})
	})

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
