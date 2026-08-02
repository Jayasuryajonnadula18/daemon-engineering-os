package runtime

import (
	"daemon/core/policies"
	"daemon/core/reasoning"
	"daemon/core/workflow"
	"daemon/core/events"
	"daemon/sdk/plugin"
	"daemon/core/storage"
)

// Container manages dependency injection across all engines and interfaces.
type Container interface {
	ResolveGraphStore() storage.GraphStore
	ResolveMemoryStore() storage.MemoryStore
	ResolveWorkflowEngine() workflow.Engine
	ResolveReasoningEngine() reasoning.Engine
	ResolveCapabilityRegistry() plugin.CapabilityRegistry
	ResolvePolicyEngine() policies.PolicyEngine
	ResolveScheduler() Scheduler
	ResolveEventBus() events.EventBus
	ResolveResourceManager() ResourceManager
}

// DefaultContainer is the standard concrete DI container implementation.
type DefaultContainer struct {
	graphStore         storage.GraphStore
	memoryStore        storage.MemoryStore
	workflowEngine     workflow.Engine
	reasoningEngine    reasoning.Engine
	capabilityRegistry plugin.CapabilityRegistry
	policyEngine       policies.PolicyEngine
	scheduler          Scheduler
	eventBus           events.EventBus
	resourceManager    ResourceManager
}

// NewDefaultContainer instantiates a new DefaultContainer.
func NewDefaultContainer(
	graphStore storage.GraphStore,
	memoryStore storage.MemoryStore,
	workflowEngine workflow.Engine,
	reasoningEngine reasoning.Engine,
	capabilityRegistry plugin.CapabilityRegistry,
	policyEngine policies.PolicyEngine,
	scheduler Scheduler,
	eventBus events.EventBus,
	resourceManager ResourceManager,
) *DefaultContainer {
	return &DefaultContainer{
		graphStore:         graphStore,
		memoryStore:        memoryStore,
		workflowEngine:     workflowEngine,
		reasoningEngine:    reasoningEngine,
		capabilityRegistry: capabilityRegistry,
		policyEngine:       policyEngine,
		scheduler:          scheduler,
		eventBus:           eventBus,
		resourceManager:    resourceManager,
	}
}

// ResolveGraphStore resolves the graph database layer.
func (c *DefaultContainer) ResolveGraphStore() storage.GraphStore {
	return c.graphStore
}

// ResolveMemoryStore resolves the engineering memory store.
func (c *DefaultContainer) ResolveMemoryStore() storage.MemoryStore {
	return c.memoryStore
}

// ResolveWorkflowEngine resolves the workflow engine.
func (c *DefaultContainer) ResolveWorkflowEngine() workflow.Engine {
	return c.workflowEngine
}

// ResolveReasoningEngine resolves the AI reasoning engine.
func (c *DefaultContainer) ResolveReasoningEngine() reasoning.Engine {
	return c.reasoningEngine
}

// ResolveCapabilityRegistry resolves the capability plugin registry.
func (c *DefaultContainer) ResolveCapabilityRegistry() plugin.CapabilityRegistry {
	return c.capabilityRegistry
}

// ResolvePolicyEngine resolves the actions policy engine.
func (c *DefaultContainer) ResolvePolicyEngine() policies.PolicyEngine {
	return c.policyEngine
}

// ResolveScheduler resolves the scheduled task runner.
func (c *DefaultContainer) ResolveScheduler() Scheduler {
	return c.scheduler
}

// ResolveEventBus resolves the system event bus.
func (c *DefaultContainer) ResolveEventBus() events.EventBus {
	return c.eventBus
}

// ResolveResourceManager resolves the resources manager.
func (c *DefaultContainer) ResolveResourceManager() ResourceManager {
	return c.resourceManager
}

