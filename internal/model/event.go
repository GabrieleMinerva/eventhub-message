package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

func Parse(raw []byte) (Event, error) {
	var ev Event
	if err := json.Unmarshal(raw, &ev); err != nil {
		return Event{}, fmt.Errorf("invalid event json: %w", err)
	}
	if ev.ID == "" {
		return Event{}, fmt.Errorf("event id is required")
	}
	if ev.Type == "" {
		return Event{}, fmt.Errorf("event type is required")
	}
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}
	if ev.Payload == nil {
		ev.Payload = map[string]interface{}{}
	}
	return ev, nil
}
