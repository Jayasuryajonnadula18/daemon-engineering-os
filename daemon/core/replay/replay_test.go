package replay

import (
	"testing"
	"time"

	"daemon/core/events"
)

func TestReplayEngine(t *testing.T) {
	eb := events.NewMemoryEventBus()
	re := NewReplayEngine(eb, nil)

	eb.Publish(events.Event{
		Type:      "DeploymentExecuted",
		Payload:   "Payments deployment completed",
		Timestamp: time.Now(),
	})

	eventsList, err := re.ReplaySession(24*time.Hour, "", "deploy")
	if err != nil {
		t.Fatalf("unexpected replay error: %v", err)
	}

	if len(eventsList) != 1 {
		t.Fatalf("expected 1 replay event, got %d", len(eventsList))
	}

	if eventsList[0].Title != "Deployment Executed" {
		t.Fatalf("expected Deployment Executed title, got %s", eventsList[0].Title)
	}
}
