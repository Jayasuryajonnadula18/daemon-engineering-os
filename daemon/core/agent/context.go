package agent

import (
	"context"
	"fmt"

	corectx "daemon/core/context"
)

type AgentContextAdapter struct {
	engine *corectx.ContextEngine
}

func NewAgentContextAdapter(engine *corectx.ContextEngine) *AgentContextAdapter {
	return &AgentContextAdapter{engine: engine}
}

type ContextQuery struct {
	Intent     string
	Resolution string // e.g. "file", "symbol", "module", "service", "workspace"
	Budget     int
}

func (aca *AgentContextAdapter) ResolveContext(ctx context.Context, q ContextQuery) (*corectx.EngineeringContext, error) {
	if aca.engine == nil {
		return nil, fmt.Errorf("context engine is not initialized")
	}

	var resLevel corectx.ResolutionLevel
	switch q.Resolution {
	case "file":
		resLevel = corectx.ResolutionFile
	case "module":
		resLevel = corectx.ResolutionModule
	case "feature":
		resLevel = corectx.ResolutionFeature
	case "service":
		resLevel = corectx.ResolutionService
	case "workspace":
		resLevel = corectx.ResolutionWorkspace
	default:
		resLevel = corectx.ResolutionWorkspace
	}

	return aca.engine.QueryBoundedContext(ctx, q.Intent, resLevel, q.Budget)
}
