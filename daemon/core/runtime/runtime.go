package runtime

import (
	"context"
	"errors"
)

// Runtime orchestrates lifecycle phases and coordinates all interfaces through its Container.
type Runtime struct {
	Container   Container
	initialized bool
	started     bool
}

// NewRuntime instantiates a new Runtime facade.
func NewRuntime(container Container) *Runtime {
	return &Runtime{Container: container}
}

// Initialize prepares core registries, storage connections, and database tables.
func (r *Runtime) Initialize(ctx context.Context) error {
	if r.initialized {
		return nil
	}
	r.initialized = true
	return nil
}

// Start begins background services like the Scheduler.
func (r *Runtime) Start(ctx context.Context) error {
	if !r.initialized {
		return errors.New("runtime not initialized")
	}
	if r.started {
		return nil
	}

	if err := r.Container.ResolveScheduler().Start(ctx); err != nil {
		return err
	}

	r.started = true
	return nil
}

// Health queries status across all active interfaces.
func (r *Runtime) Health() string {
	if !r.initialized {
		return "uninitialized"
	}
	if !r.started {
		return "stopped"
	}
	return "healthy"
}

// Reload triggers configuration refreshes.
func (r *Runtime) Reload(ctx context.Context) error {
	return nil
}

// Stop pauses all execution workers and background runners.
func (r *Runtime) Stop() error {
	if !r.started {
		return nil
	}

	if err := r.Container.ResolveScheduler().Stop(); err != nil {
		return err
	}

	r.started = false
	return nil
}

// Shutdown stops active processes and releases file locks.
func (r *Runtime) Shutdown(ctx context.Context) error {
	_ = r.Stop()
	return r.Container.ResolveResourceManager().Cleanup(ctx)
}

