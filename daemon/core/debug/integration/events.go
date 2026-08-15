package integration

import (
	"fmt"
	"strings"

	"daemon/core/events"
	"daemon/core/instruments"
)

type EventsAdapter struct {
	bus events.EventBus
}

func NewEventsAdapter(bus events.EventBus) *EventsAdapter {
	return &EventsAdapter{bus: bus}
}

// GatherEventsEvidence scans the EventBus timeline for failures and alerts
func (ea *EventsAdapter) GatherEventsEvidence() ([]instruments.Evidence, error) {
	if ea.bus == nil {
		return nil, nil
	}

	timeline := ea.bus.GetTimeline()
	var list []instruments.Evidence

	for i, e := range timeline {
		// Only capture system alerts, errors, or restarts
		if containsAny(e.Type, "fail", "alert", "error", "restart", "incident") {
			list = append(list, instruments.Evidence{
				ID:           fmt.Sprintf("ev-event-alert-%d", i),
				Type:         instruments.EvidenceEvent,
				Source:       "event_bus",
				EntityID:     e.EntityID,
				Statement:    fmt.Sprintf("Timeline alert of type '%s' from source '%s'", e.Type, e.Source),
				ObservedAt:   e.Timestamp,
				Freshness:    "live",
				Reliability:  0.85,
				Confidence:   0.85,
				Scope:        "events",
				Quality: instruments.EvidenceQuality{
					Class:           "event_bus",
					Strength:        0.85,
					Reliability:     0.85,
					Freshness:       1.0,
					Specificity:     0.85,
					Independence:    1.0,
					Reproducibility: 1.0,
					Verification:    "VERIFIED",
					Provenance:      "event_bus",
				},
			})
		}
	}

	return list, nil
}

func containsAny(s string, keywords ...string) bool {
	lower := strings.ToLower(s)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
