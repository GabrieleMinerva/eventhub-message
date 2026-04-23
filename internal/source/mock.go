package source

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/example/go-worker-service/internal/worker"
)

type MockSource struct {
	ticker *time.Ticker
	count  int64
}

func NewMockSource(interval time.Duration) *MockSource {
	return &MockSource{ticker: time.NewTicker(interval)}
}

func (m *MockSource) Receive(ctx context.Context) (worker.Message, error) {
	select {
	case <-ctx.Done():
		return worker.Message{}, ctx.Err()
	case <-m.ticker.C:
		m.count++
		body, _ := json.Marshal(map[string]any{
			"id":        fmt.Sprintf("mock-%d", m.count),
			"type":      "mock.generated",
			"timestamp": time.Now().UTC(),
			"payload": map[string]any{
				"sequence": m.count,
				"source":   "mock",
			},
		})
		return worker.Message{Body: body, Partition: "mock", SequenceID: m.count}, nil
	}
}

func (m *MockSource) Close(context.Context) error {
	m.ticker.Stop()
	return nil
}
