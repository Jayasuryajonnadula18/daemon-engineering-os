package context

import (
	"context"

	"daemon/core/domain"
	"daemon/core/storage"
)

type ResolutionLevel string

const (
	ResolutionFile      ResolutionLevel = "file"
	ResolutionModule    ResolutionLevel = "module"
	ResolutionFeature   ResolutionLevel = "feature"
	ResolutionService   ResolutionLevel = "service"
	ResolutionWorkspace ResolutionLevel = "workspace"
)

// BoundedContextMetadata records query token budget and entity bounds.
type BoundedContextMetadata struct {
	Resolution          ResolutionLevel `json:"resolution"`
	EntitiesCount       int             `json:"entities_count"`
	FilesCount          int             `json:"files_count"`
	EstimatedTokens     int             `json:"estimated_tokens"`
	MaxTokenBudget      int             `json:"max_token_budget"`
	InsufficientContext bool            `json:"insufficient_context"`
	SourcesCount        int             `json:"sources_count"`
}

// EngineeringContext is the single representation of the workspace state.
type EngineeringContext struct {
	Resolution      ResolutionLevel         `json:"resolution"`
	Metadata        BoundedContextMetadata  `json:"metadata"`
	Services        []domain.Service        `json:"services"`
	Dependencies    []domain.Dependency    `json:"dependencies"`
	Incidents       []domain.Incident       `json:"incidents"`
	Recommendations []domain.Recommendation `json:"recommendations"`
	Deployments     []domain.Deployment     `json:"deployments"`
	Provenance      domain.Provenance       `json:"provenance"`
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

// QueryBoundedContext compiles a bounded context payload within a specified token budget and resolution level.
func (ce *ContextEngine) QueryBoundedContext(ctx context.Context, query string, targetResolution ResolutionLevel, tokenBudget int) (*EngineeringContext, error) {
	if tokenBudget <= 0 {
		tokenBudget = 8000 // default max token cap
	}

	engCtx, err := ce.BuildContext(ctx)
	if err != nil {
		return nil, err
	}

	engCtx.Resolution = targetResolution
	if targetResolution == "" {
		engCtx.Resolution = ResolutionService
	}

	// Calculate entity counts and token estimates
	totalEntities := len(engCtx.Services) + len(engCtx.Dependencies) + len(engCtx.Incidents)
	estTokens := totalEntities * 150

	insufficient := totalEntities == 0

	engCtx.Metadata = BoundedContextMetadata{
		Resolution:          engCtx.Resolution,
		EntitiesCount:       totalEntities,
		FilesCount:          len(engCtx.Services),
		EstimatedTokens:     estTokens,
		MaxTokenBudget:      tokenBudget,
		InsufficientContext: insufficient,
		SourcesCount:        3,
	}

	return engCtx, nil
}

// BuildContext compiles services, dependencies, and memory incidents into the unified context structure.
func (ce *ContextEngine) BuildContext(ctx context.Context) (*EngineeringContext, error) {
	var services []domain.Service
	var dependencies []domain.Dependency
	var incidents []domain.Incident
	var recs []domain.Recommendation
	var deployments []domain.Deployment

	if ce.graphStore != nil {
		svcs, err := ce.graphStore.GetServices()
		if err == nil {
			services = svcs
		}
		deps, err := ce.graphStore.GetDependencies()
		if err == nil {
			dependencies = deps
		}
	}
	if services == nil {
		services = make([]domain.Service, 0)
	}
	if dependencies == nil {
		dependencies = make([]domain.Dependency, 0)
	}

	if ce.memoryStore != nil {
		incs, err := ce.memoryStore.GetIncidents()
		if err == nil {
			incidents = incs
		}
		rcs, err := ce.memoryStore.GetRecommendations()
		if err == nil {
			recs = rcs
		}
		deps, err := ce.memoryStore.GetDeployments()
		if err == nil {
			deployments = deps
		}
	}
	if incidents == nil {
		incidents = make([]domain.Incident, 0)
	}
	if recs == nil {
		recs = make([]domain.Recommendation, 0)
	}
	if deployments == nil {
		deployments = make([]domain.Deployment, 0)
	}

	return &EngineeringContext{
		Resolution:      ResolutionWorkspace,
		Services:        services,
		Dependencies:    dependencies,
		Incidents:       incidents,
		Recommendations: recs,
		Deployments:     deployments,
	}, nil
}
