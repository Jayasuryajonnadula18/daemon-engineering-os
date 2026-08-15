package workflow

import (
	"context"
	"errors"
	"sync"
	"time"

	"daemon/core/domain"
	"daemon/core/events"
)

// Step represents a named execution logic closure.
type Step struct {
	Name string
	Run  func(ctx context.Context) error
}

// Engine defines operations for initiating and executing workflows.
type Engine interface {
	CreateWorkflow(name string, steps []Step) (*domain.Workflow, error)
	Execute(ctx context.Context, id string) error
	GetWorkflow(id string) (*domain.Workflow, error)
}

// MemoryWorkflowEngine manages workflows in memory and broadcasts progression events.
type MemoryWorkflowEngine struct {
	mu        sync.RWMutex
	workflows map[string]*domain.Workflow
	steps     map[string][]Step
	eventBus  events.EventBus
}

// NewMemoryWorkflowEngine initializes a new MemoryWorkflowEngine.
func NewMemoryWorkflowEngine(eb events.EventBus) *MemoryWorkflowEngine {
	return &MemoryWorkflowEngine{
		workflows: make(map[string]*domain.Workflow),
		steps:     make(map[string][]Step),
		eventBus:  eb,
	}
}

// CreateWorkflow registers a new workflow with standard pending status.
func (e *MemoryWorkflowEngine) CreateWorkflow(name string, steps []Step) (*domain.Workflow, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	id := name + "-" + time.Now().Format("20060102-150405.000")
	stepNames := make([]string, len(steps))
	for i, s := range steps {
		stepNames[i] = s.Name
	}

	wf := &domain.Workflow{
		ID:        id,
		Name:      name,
		Status:    "pending",
		Steps:     stepNames,
		StartedAt: time.Time{},
		EndedAt:   time.Time{},
	}

	e.workflows[id] = wf
	e.steps[id] = steps

	return wf, nil
}

// Execute runs all steps sequentially, broadcasting start, step, and end updates.
func (e *MemoryWorkflowEngine) Execute(ctx context.Context, id string) error {
	e.mu.RLock()
	wf, ok := e.workflows[id]
	steps, stepOk := e.steps[id]
	e.mu.RUnlock()

	if !ok || !stepOk {
		return errors.New("workflow not found")
	}

	e.mu.Lock()
	wf.Status = "running"
	wf.StartedAt = time.Now()
	e.mu.Unlock()

	e.eventBus.Publish(events.Event{
		Type:      "WorkflowStarted",
		Payload:   map[string]any{"workflow_id": wf.ID, "name": wf.Name, "status": wf.Status},
		Timestamp: time.Now(),
	})

	var runErr error
	for _, step := range steps {
		if err := step.Run(ctx); err != nil {
			runErr = err
			break
		}
	}

	e.mu.Lock()
	if runErr != nil {
		wf.Status = "failed"
	} else {
		wf.Status = "succeeded"
	}
	wf.EndedAt = time.Now()
	e.mu.Unlock()

	e.eventBus.Publish(events.Event{
		Type:      "WorkflowCompleted",
		Payload:   map[string]any{"workflow_id": wf.ID, "name": wf.Name, "status": wf.Status},
		Timestamp: time.Now(),
	})

	return runErr
}

// GetWorkflow returns a workflow record by ID.
func (e *MemoryWorkflowEngine) GetWorkflow(id string) (*domain.Workflow, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	wf, ok := e.workflows[id]
	if !ok {
		return nil, errors.New("workflow not found")
	}
	return wf, nil
}

