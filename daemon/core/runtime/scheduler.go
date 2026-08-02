package runtime

import (
	"context"
	"sync"
	"time"
)

// Job represents a routine scheduled task.
type Job struct {
	ID       string
	Interval time.Duration
	Run      func(ctx context.Context) error
}

// Scheduler triggers periodic workflows in the background.
type Scheduler interface {
	Schedule(job Job) error
	Start(ctx context.Context) error
	Stop() error
}

// MemoryScheduler implements basic periodic execution intervals.
type MemoryScheduler struct {
	mu     sync.RWMutex
	jobs   []Job
	stopCh chan struct{}
}

// NewMemoryScheduler instantiates a new MemoryScheduler.
func NewMemoryScheduler() *MemoryScheduler {
	return &MemoryScheduler{
		jobs:   make([]Job, 0),
		stopCh: make(chan struct{}),
	}
}

// Schedule queues a background job.
func (s *MemoryScheduler) Schedule(job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, job)
	return nil
}

// Start spawns background runner routines for all queued jobs.
func (s *MemoryScheduler) Start(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, job := range s.jobs {
		go func(j Job) {
			ticker := time.NewTicker(j.Interval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					_ = j.Run(ctx)
				case <-s.stopCh:
					return
				case <-ctx.Done():
					return
				}
			}
		}(job)
	}

	return nil
}

// Stop terminates all scheduled tickers.
func (s *MemoryScheduler) Stop() error {
	close(s.stopCh)
	return nil
}

