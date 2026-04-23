package worker

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/example/go-worker-service/internal/model"
)

type fakeSource struct {
	messages []Message
	idx      int
}

func (f *fakeSource) Receive(ctx context.Context) (Message, error) {
	if f.idx >= len(f.messages) {
		<-ctx.Done()
		return Message{}, ctx.Err()
	}
	m := f.messages[f.idx]
	f.idx++
	return m, nil
}
func (f *fakeSource) Close(context.Context) error { return nil }

type fakeProcessor struct{ fail bool }

func (p fakeProcessor) Process(_ context.Context, _ model.Event) error {
	if p.fail {
		return errors.New("boom")
	}
	return nil
}

func TestServiceRunProcessesMessage(t *testing.T) {
	src := &fakeSource{messages: []Message{{Body: []byte(`{"id":"1","type":"t"}`)}}}
	stats := &Stats{}
	svc := New(slog.Default(), src, fakeProcessor{}, stats, 5*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_ = svc.Run(ctx)

	processed, _, _ := svc.Stats()
	if processed == 0 {
		t.Fatal("expected at least one processed event")
	}
}


func TestDedupProcessorSkipsDuplicate(t *testing.T) {
	base := fakeProcessor{}
	dp := NewDedupProcessor(base, time.Minute, 100)
	ev := model.Event{ID: "same", Type: "x"}
	if err := dp.Process(context.Background(), ev); err != nil {
		t.Fatalf("unexpected first process error: %v", err)
	}
	if err := dp.Process(context.Background(), ev); !errors.Is(err, ErrSkipEvent) {
		t.Fatalf("expected ErrSkipEvent, got %v", err)
	}
}
