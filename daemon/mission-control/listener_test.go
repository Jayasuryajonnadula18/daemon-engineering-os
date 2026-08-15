package missioncontrol

import (
	"context"
	"testing"
	"time"

	"daemon/core/events"
)

func TestAmbientListener_IngestEventAuthAndPublish(t *testing.T) {
	eb := events.NewMemoryEventBus()
	listener := NewAmbientListener(eb, "secret-token-123")

	evt := IDETelemetryEvent{
		Source:    "vscode-extension",
		EventType: "file_saved",
		FilePath:  "main.go",
		Timestamp: time.Now(),
	}

	// 1. Unauthorized Attempt
	err := listener.IngestEvent(context.Background(), "invalid-token", evt)
	if err == nil {
		t.Fatalf("expected UNAUTHORIZED error on invalid token")
	}

	// 2. Authorized Attempt
	err = listener.IngestEvent(context.Background(), "secret-token-123", evt)
	if err != nil {
		t.Fatalf("unexpected error on authorized ingestion: %v", err)
	}

	timeline := eb.GetTimeline()
	if len(timeline) != 1 {
		t.Fatalf("expected 1 event in event bus timeline, got %d", len(timeline))
	}
	if timeline[0].Type != "file_saved" {
		t.Fatalf("event type mismatch: %s", timeline[0].Type)
	}
}
