package model

import "testing"

func TestParse(t *testing.T) {
	raw := []byte(`{"id":"e-1","type":"user.created","payload":{"k":"v"}}`)
	ev, err := Parse(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ev.ID != "e-1" || ev.Type != "user.created" {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestParseMissingID(t *testing.T) {
	_, err := Parse([]byte(`{"type":"a"}`))
	if err == nil {
		t.Fatalf("expected error")
	}
}
