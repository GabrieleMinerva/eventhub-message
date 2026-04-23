package source

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/example/go-worker-service/internal/worker"
)

// EventHubSource is a lightweight adapter used for local/demo environments where
// direct Azure SDK access is not available. It reads JSON events from a file path
// (newline-delimited JSON) and emits them as worker messages in a loop.
type EventHubSource struct {
	messages [][]byte
	idx      int
}

func NewEventHubSource(_ string, _ string, _ string, _ time.Duration) (*EventHubSource, error) {
	path := os.Getenv("EVENTHUB_EVENTS_FILE")
	if path == "" {
		return nil, fmt.Errorf("EVENTHUB_EVENTS_FILE is required in this environment for eventhub mode")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read EVENTHUB_EVENTS_FILE: %w", err)
	}
	lines := splitLines(raw)
	if len(lines) == 0 {
		return nil, fmt.Errorf("EVENTHUB_EVENTS_FILE has no events")
	}
	for i, line := range lines {
		if !json.Valid(line) {
			return nil, fmt.Errorf("invalid json at line %d", i+1)
		}
	}
	return &EventHubSource{messages: lines}, nil
}

func (s *EventHubSource) Receive(ctx context.Context) (worker.Message, error) {
	if len(s.messages) == 0 {
		<-ctx.Done()
		return worker.Message{}, ctx.Err()
	}
	if s.idx >= len(s.messages) {
		s.idx = 0
	}
	msg := worker.Message{Body: s.messages[s.idx], Partition: "eventhub-file", SequenceID: int64(s.idx + 1)}
	s.idx++
	select {
	case <-ctx.Done():
		return worker.Message{}, ctx.Err()
	case <-time.After(250 * time.Millisecond):
		return msg, nil
	}
}

func (s *EventHubSource) Close(context.Context) error { return nil }

func splitLines(raw []byte) [][]byte {
	parts := [][]byte{}
	start := 0
	for i, b := range raw {
		if b == '\n' {
			if i > start {
				parts = append(parts, raw[start:i])
			}
			start = i + 1
		}
	}
	if start < len(raw) {
		parts = append(parts, raw[start:])
	}
	return parts
}
