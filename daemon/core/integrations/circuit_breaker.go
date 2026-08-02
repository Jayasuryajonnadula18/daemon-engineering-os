package integrations

import (
	"context"
	"errors"
	"math"
	"time"
)

// CircuitState represents the stability state of an integration link.
type CircuitState string

const (
	StateClosed   CircuitState = "Closed"
	StateOpen     CircuitState = "Open"
	StateHalfOpen CircuitState = "HalfOpen"
)

// CircuitBreaker wraps execution flows to prevent cascade failures.
type CircuitBreaker struct {
	state        CircuitState
	failureCount int
	threshold    int
	cooldown     time.Duration
	lastFailure  time.Time
}

// NewCircuitBreaker builds a CircuitBreaker.
func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:     StateClosed,
		threshold: threshold,
		cooldown:  cooldown,
	}
}

// Execute wraps a target function call.
func (cb *CircuitBreaker) Execute(f func() error) error {
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.cooldown {
			cb.state = StateHalfOpen
		} else {
			return errors.New("circuit breaker is open; request blocked")
		}
	}

	err := f()
	if err != nil {
		cb.failureCount++
		cb.lastFailure = time.Now()
		if cb.failureCount >= cb.threshold {
			cb.state = StateOpen
		}
		return err
	}

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failureCount = 0
	}
	return nil
}

// RetryWithBackoff executes a function with exponential backoff intervals.
func RetryWithBackoff(ctx context.Context, maxRetries int, baseDelay time.Duration, f func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err = f()
		if err == nil {
			return nil
		}

		delay := time.Duration(float64(baseDelay) * math.Pow(1.5, float64(i)))
		// Ensure sleep is context-sensitive
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return err
}
