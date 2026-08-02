package integrations

import (
	"context"
	"errors"
	"testing"
	"time"

	"daemon/core/storage"
)

type DummyGraphStore struct {
	storage.GraphStore
}

func (d *DummyGraphStore) AddNode(nodeType, id, name string, properties map[string]string) error {
	return nil
}
func (d *DummyGraphStore) AddEdge(fromType, fromID, toType, toID, relation string) error {
	return nil
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	if cb.state != StateClosed {
		t.Fatalf("expected state Closed, got %s", cb.state)
	}

	_ = cb.Execute(func() error { return errors.New("fail") })
	_ = cb.Execute(func() error { return errors.New("fail") })

	if cb.state != StateOpen {
		t.Fatalf("expected breaker to be Open, got %s", cb.state)
	}

	err := cb.Execute(func() error { return nil })
	if err == nil || err.Error() != "circuit breaker is open; request blocked" {
		t.Fatalf("expected block error, got %v", err)
	}
}

func TestRetryWithBackoff(t *testing.T) {
	ctx := context.Background()
	counter := 0
	err := RetryWithBackoff(ctx, 3, 10*time.Millisecond, func() error {
		counter++
		if counter < 2 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected successful retry execution, got %v", err)
	}
	if counter != 2 {
		t.Fatalf("expected counter to be 2, got %d", counter)
	}
}
