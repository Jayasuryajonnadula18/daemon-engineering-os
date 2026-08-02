package events

import (
	"sync"
	"time"
)

// Event represents a system event inside the Daemon Runtime.
type Event struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// EventBus provides pub/sub and timeline retrieval for internal events.
type EventBus interface {
	Publish(event Event)
	Subscribe(eventType string, handler func(Event))
	GetTimeline() []Event
}

// MemoryEventBus is a thread-safe, in-memory implementation of the EventBus interface.
type MemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]func(Event)
	timeline    []Event
}

// NewMemoryEventBus instantiates a new MemoryEventBus.
func NewMemoryEventBus() *MemoryEventBus {
	return &MemoryEventBus{
		subscribers: make(map[string][]func(Event)),
		timeline:    make([]Event, 0),
	}
}

// Publish writes the event to the timeline log and triggers all registered event handlers.
func (b *MemoryEventBus) Publish(event Event) {
	b.mu.Lock()
	b.timeline = append(b.timeline, event)
	b.mu.Unlock()

	b.mu.RLock()
	defer b.mu.RUnlock()

	// Notify explicit handlers
	if handlers, ok := b.subscribers[event.Type]; ok {
		for _, handler := range handlers {
			go handler(event)
		}
	}

	// Notify wildcard handlers
	if handlers, ok := b.subscribers["*"]; ok {
		for _, handler := range handlers {
			go handler(event)
		}
	}
}

// Subscribe registers a callback to trigger on specific events (or '*' for all events).
func (b *MemoryEventBus) Subscribe(eventType string, handler func(Event)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

// GetTimeline returns a copy of all recorded timeline events.
func (b *MemoryEventBus) GetTimeline() []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()
	res := make([]Event, len(b.timeline))
	copy(res, b.timeline)
	return res
}

