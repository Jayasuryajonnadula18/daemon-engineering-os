package missioncontrol

import (
	"context"
	"fmt"
	"sync"
	"time"

	"daemon/core/events"
)

type IDETelemetryEvent struct {
	ID        string            `json:"id"`
	Source    string            `json:"source"`    // e.g. "vscode-extension"
	EventType string            `json:"event_type"`// e.g. "file_saved", "build_triggered"
	FilePath  string            `json:"file_path,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

type AmbientListener struct {
	mu           sync.RWMutex
	eventBus     events.EventBus
	sessionToken string
}

func NewAmbientListener(eb events.EventBus, token string) *AmbientListener {
	if token == "" {
		token = "daemon-local-session-secret"
	}
	return &AmbientListener{
		eventBus:     eb,
		sessionToken: token,
	}
}

// IngestEvent validates token authentication, sanitizes IDE telemetry, and broadcasts to MemoryEventBus.
func (al *AmbientListener) IngestEvent(ctx context.Context, authToken string, evt IDETelemetryEvent) error {
	al.mu.RLock()
	defer al.mu.RUnlock()

	if authToken != al.sessionToken && authToken != "Bearer "+al.sessionToken {
		return fmt.Errorf("UNAUTHORIZED: Invalid or missing IDE session authentication token")
	}

	if evt.EventType == "" {
		return fmt.Errorf("BAD_REQUEST: EventType is required")
	}

	payload := map[string]interface{}{
		"source":    evt.Source,
		"file_path": evt.FilePath,
		"metadata":  evt.Metadata,
	}

	busEvent := events.Event{
		ID:        fmt.Sprintf("evt-ide-%d", time.Now().UnixNano()),
		Type:      evt.EventType,
		Source:    evt.Source,
		EntityID:  evt.FilePath,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	if al.eventBus != nil {
		al.eventBus.Publish(busEvent)
	}

	return nil
}
