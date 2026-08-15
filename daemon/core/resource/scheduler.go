package resource

import (
	"context"
	"fmt"
	"sync"
)

type TaskPriority string

const (
	PriorityCritical     TaskPriority = "CRITICAL"
	PriorityImportant    TaskPriority = "IMPORTANT"
	PriorityBackground   TaskPriority = "BACKGROUND"
	PriorityOpportunistic TaskPriority = "OPPORTUNISTIC"
)

type ScheduledTask struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Priority TaskPriority `json:"priority"`
	Fn       func(ctx context.Context) error
}

type AdaptiveScheduler struct {
	mu       sync.RWMutex
	governor *ResourceGovernor
	tasks    map[string]ScheduledTask
}

func NewAdaptiveScheduler(gov *ResourceGovernor) *AdaptiveScheduler {
	return &AdaptiveScheduler{
		governor: gov,
		tasks:    make(map[string]ScheduledTask),
	}
}

func (s *AdaptiveScheduler) RegisterTask(task ScheduledTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
}

func (s *AdaptiveScheduler) ExecuteTask(ctx context.Context, taskID string, isUserRequested bool) (GovernorDecision, error) {
	s.mu.RLock()
	task, exists := s.tasks[taskID]
	s.mu.RUnlock()

	if !exists {
		return GovernorDecision{}, fmt.Errorf("task ID %s not registered", taskID)
	}

	decision := s.governor.Evaluate(task.Name, isUserRequested)

	// Critical tasks always execute regardless of background pressure
	if task.Priority == PriorityCritical {
		decision.Decision = DecisionExecute
		decision.Reason = "Task priority CRITICAL: bypasses background pressure deferral."
	}

	if decision.Decision == DecisionDefer || decision.Decision == DecisionPause {
		if task.Priority == PriorityBackground || task.Priority == PriorityOpportunistic {
			return decision, fmt.Errorf("TASK_DEFERRED: %s", decision.Reason)
		}
	}

	if task.Fn != nil {
		err := task.Fn(ctx)
		return decision, err
	}

	return decision, nil
}
