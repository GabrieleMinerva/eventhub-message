package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/example/go-worker-service/internal/model"
)

type Message struct {
	Body       []byte
	Partition  string
	SequenceID int64
	Ack        func(context.Context) error
}

type Source interface {
	Receive(context.Context) (Message, error)
	Close(context.Context) error
}

type Processor interface {
	Process(context.Context, model.Event) error
}

type Stats struct {
	Processed atomic.Uint64
	Failed    atomic.Uint64
	Skipped   atomic.Uint64
}

type Service struct {
	logger              *slog.Logger
	source              Source
	processor           Processor
	stats               *Stats
	receiveErrorBackoff time.Duration
}

func New(logger *slog.Logger, source Source, processor Processor, stats *Stats, receiveErrorBackoff time.Duration) *Service {
	if stats == nil {
		stats = &Stats{}
	}
	if receiveErrorBackoff <= 0 {
		receiveErrorBackoff = 250 * time.Millisecond
	}
	return &Service{logger: logger, source: source, processor: processor, stats: stats, receiveErrorBackoff: receiveErrorBackoff}
}

func (s *Service) Run(ctx context.Context) error {
	s.logger.Info("worker started")
	defer s.logger.Info("worker stopped")

	for {
		msg, err := s.source.Receive(ctx)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, context.Canceled) {
				return nil
			}
			s.stats.Failed.Add(1)
			s.logger.Error("failed to receive message", "error", err)
			if !sleepWithContext(ctx, s.receiveErrorBackoff) {
				return nil
			}
			continue
		}

		ev, err := model.Parse(msg.Body)
		if err != nil {
			s.stats.Failed.Add(1)
			s.logger.Warn("dropping invalid event", "error", err, "partition", msg.Partition, "sequence", msg.SequenceID)
			continue
		}

		err = s.processor.Process(ctx, ev)
		if err != nil {
			if errors.Is(err, ErrSkipEvent) {
				s.stats.Skipped.Add(1)
				s.logger.Info("skipped duplicate event", "event_id", ev.ID)
				continue
			}
			s.stats.Failed.Add(1)
			s.logger.Error("processor failed", "error", err, "event_id", ev.ID)
			continue
		}

		if msg.Ack != nil {
			if err := msg.Ack(ctx); err != nil {
				s.stats.Failed.Add(1)
				s.logger.Error("ack failed", "error", err, "event_id", ev.ID)
				continue
			}
		}

		s.stats.Processed.Add(1)
		s.logger.Info("event processed", "event_id", ev.ID, "type", ev.Type)
	}
}

func (s *Service) Shutdown(ctx context.Context) error {
	if err := s.source.Close(ctx); err != nil {
		return fmt.Errorf("source close: %w", err)
	}
	return nil
}

func (s *Service) Stats() (processed, failed, skipped uint64) {
	return s.stats.Processed.Load(), s.stats.Failed.Load(), s.stats.Skipped.Load()
}

type LogProcessor struct {
	logger *slog.Logger
}

func NewLogProcessor(logger *slog.Logger) *LogProcessor {
	return &LogProcessor{logger: logger}
}

func (p *LogProcessor) Process(_ context.Context, ev model.Event) error {
	p.logger.Info("processing payload", "event_id", ev.ID, "payload", ev.Payload)
	return nil
}

var ErrSkipEvent = errors.New("skip event")

type DedupProcessor struct {
	next       Processor
	mu         sync.Mutex
	seen       map[string]time.Time
	ttl        time.Duration
	maxEntries int
}

func NewDedupProcessor(next Processor, ttl time.Duration, maxEntries int) *DedupProcessor {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if maxEntries <= 0 {
		maxEntries = 10000
	}
	return &DedupProcessor{next: next, seen: make(map[string]time.Time, maxEntries), ttl: ttl, maxEntries: maxEntries}
}

func (d *DedupProcessor) Process(ctx context.Context, ev model.Event) error {
	now := time.Now()
	d.mu.Lock()
	d.gc(now)
	if exp, ok := d.seen[ev.ID]; ok && exp.After(now) {
		d.mu.Unlock()
		return ErrSkipEvent
	}
	d.seen[ev.ID] = now.Add(d.ttl)
	d.mu.Unlock()
	return d.next.Process(ctx, ev)
}

func (d *DedupProcessor) gc(now time.Time) {
	if len(d.seen) < d.maxEntries {
		return
	}
	for k, exp := range d.seen {
		if exp.Before(now) {
			delete(d.seen, k)
		}
	}
	if len(d.seen) > d.maxEntries {
		for k := range d.seen {
			delete(d.seen, k)
			if len(d.seen) <= d.maxEntries {
				break
			}
		}
	}
}

func sleepWithContext(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
