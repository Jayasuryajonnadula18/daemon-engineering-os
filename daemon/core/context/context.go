package context

import (
	"context"

	"daemon/core/domain"
	"daemon/core/storage"
)

// EngineeringContext is the single representation of the workspace state.
type EngineeringContext struct {
	Services        []domain.Service        `json:"services"`
	Dependencies    []domain.Dependency    `json:"dependencies"`
	Incidents       []domain.Incident       `json:"incidents"`
	Recommendations []domain.Recommendation `json:"recommendations"`
	Deployments     []domain.Deployment     `json:"deployments"`
}

// ContextEngine coordinates data compilation across the twin store.
type ContextEngine struct {
	graphStore  storage.GraphStore
	memoryStore storage.MemoryStore
}

// NewContextEngine instantiates a new ContextEngine.
func NewContextEngine(gs storage.GraphStore, ms storage.MemoryStore) *ContextEngine {
	return &ContextEngine{
		graphStore:  gs,
		memoryStore: ms,
	}
}

// BuildContext compiles services, dependencies, and memory incidents into the unified context structure.
func (ce *ContextEngine) BuildContext(ctx context.Context) (*EngineeringContext, error) {
	services, err := ce.graphStore.GetServices()
	if err != nil {
		services = make([]domain.Service, 0)
	}

	dependencies, err := ce.graphStore.GetDependencies()
	if err != nil {
		dependencies = make([]domain.Dependency, 0)
	}

	incidents, err := ce.memoryStore.GetIncidents()
	if err != nil {
		incidents = make([]domain.Incident, 0)
	}

	recs, err := ce.memoryStore.GetRecommendations()
	if err != nil {
		recs = make([]domain.Recommendation, 0)
	}

	deployments, err := ce.memoryStore.GetDeployments()
	if err != nil {
		deployments = make([]domain.Deployment, 0)
	}

	return &EngineeringContext{
		Services:        services,
		Dependencies:    dependencies,
		Incidents:       incidents,
		Recommendations: recs,
		Deployments:     deployments,
	}, nil
}
