package reasoning

import (
	"context"
	"strings"

	engContext "daemon/core/context"
	"daemon/core/domain"
)

// ContextBuilder optimizes prompt context windows based on developer intent.
type ContextBuilder struct {
	contextEngine *engContext.ContextEngine
}

// NewContextBuilder instantiates a new ContextBuilder.
func NewContextBuilder(ce *engContext.ContextEngine) *ContextBuilder {
	return &ContextBuilder{contextEngine: ce}
}

// BuildOptimizedContext gathers and prunes the unified context relevant to the user request.
func (cb *ContextBuilder) BuildOptimizedContext(ctx context.Context, intent string) (*engContext.EngineeringContext, error) {
	fullCtx, err := cb.contextEngine.BuildContext(ctx)
	if err != nil {
		return nil, err
	}

	intentLower := strings.ToLower(intent)

	// If no intent keyword filter matches, return full context
	var filteredServices []domain.Service
	var filteredIncidents []domain.Incident
	var filteredRecs []domain.Recommendation

	// Search for key target systems in intent (e.g. "orders", "payments", "auth")
	hasFilters := false
	targets := []string{"orders", "payments", "auth", "gateway"}
	for _, t := range targets {
		if strings.Contains(intentLower, t) {
			hasFilters = true
		}
	}

	if !hasFilters {
		return fullCtx, nil
	}

	for _, s := range fullCtx.Services {
		name := strings.ToLower(s.Name)
		for _, t := range targets {
			if strings.Contains(intentLower, t) && strings.Contains(name, t) {
				filteredServices = append(filteredServices, s)
			}
		}
	}

	for _, inc := range fullCtx.Incidents {
		msg := strings.ToLower(inc.Message)
		for _, t := range targets {
			if strings.Contains(intentLower, t) && strings.Contains(msg, t) {
				filteredIncidents = append(filteredIncidents, inc)
			}
		}
	}

	for _, rec := range fullCtx.Recommendations {
		msg := strings.ToLower(rec.Message)
		for _, t := range targets {
			if strings.Contains(intentLower, t) && strings.Contains(msg, t) {
				filteredRecs = append(filteredRecs, rec)
			}
		}
	}

	// Fallback to avoid returning completely empty arrays if filter removed all
	if len(filteredServices) == 0 {
		filteredServices = fullCtx.Services
	}
	if len(filteredIncidents) == 0 {
		filteredIncidents = fullCtx.Incidents
	}
	if len(filteredRecs) == 0 {
		filteredRecs = fullCtx.Recommendations
	}

	return &engContext.EngineeringContext{
		Services:        filteredServices,
		Dependencies:    fullCtx.Dependencies, // keep dependencies unchanged
		Incidents:       filteredIncidents,
		Recommendations: filteredRecs,
		Deployments:     fullCtx.Deployments,
	}, nil
}
