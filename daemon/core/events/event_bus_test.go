package events

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryEventBus_PubSubAndTimeline(t *testing.T) {
	bus := NewMemoryEventBus()

	var counter int32
	bus.Subscribe("file.changed", func(e Event) {
		atomic.AddInt32(&counter, 1)
	})

	evt := Event{
		ID:        "evt-1",
		Type:      "file.changed",
		Source:    "filesystem",
		EntityID:  ".env",
		Timestamp: time.Now(),
		Payload:   map[string]any{"path": ".env"},
	}

	bus.Publish(evt)

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&counter) != 1 {
		t.Fatalf("expected subscriber callback to run 1 time, got %d", counter)
	}

	timeline := bus.GetTimeline()
	if len(timeline) != 1 {
		t.Fatalf("expected timeline to have 1 event, got %d", len(timeline))
	}
}
