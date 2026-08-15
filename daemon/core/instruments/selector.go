package instruments

import (
	"context"
	"fmt"
)

type InvestigationRequest struct {
	Capability        Capability        `json:"capability"`
	Language          string            `json:"language"`
	Runtime           string            `json:"runtime"`
	MaxCost           float64           `json:"max_cost"`
	MaxCPUPercent     float64           `json:"max_cpu_percent"`
	LocalOnly         bool              `json:"local_only"`
	Metadata          map[string]string `json:"metadata"`
}

type InstrumentPlan struct {
	Capability        Capability           `json:"capability"`
	Selected          string               `json:"selected"`
	SelectionReason   []string             `json:"selection_reason"`
	Alternatives      []string             `json:"alternatives"`
	ExecutionRequired bool                 `json:"execution_required"`
	ExpectedCost      float64              `json:"expected_cost"`
	Identity          *InstrumentIdentity  `json:"identity,omitempty"`
}

type InstrumentSelector struct {
	registry *InstrumentRegistry
}

func NewInstrumentSelector(registry *InstrumentRegistry) *InstrumentSelector {
	return &InstrumentSelector{registry: registry}
}

// SelectInstrument chooses the best instrument matching the criteria.
func (s *InstrumentSelector) SelectInstrument(ctx context.Context, req InvestigationRequest, env Environment) (InstrumentPlan, error) {
	plan := InstrumentPlan{
		Capability:        req.Capability,
		Alternatives:      []string{},
		ExecutionRequired: false,
	}

	// 1. Get all instruments matching capability
	candidates := s.registry.FindByCapability(req.Capability)
	if len(candidates) == 0 {
		return plan, fmt.Errorf("no instruments registered for capability %s", req.Capability)
	}

	type ScoredCandidate struct {
		inst   EngineeringInstrument
		score  float64
		reasons []string
	}

	var scored []ScoredCandidate

	// 2. Score candidates
	for _, inst := range candidates {
		identity := inst.Identity()
		det := inst.Detect(ctx, env)

		// Hard constraint: Project must be compatible
		if !det.Compatible {
			continue
		}

		score := 0.0
		var reasons []string

		reasons = append(reasons, "project_runtime_compatible")
		reasons = append(reasons, "adapter_available")

		if identity.Installed {
			score += 10.0
			reasons = append(reasons, "tool_discovered")
			reasons = append(reasons, "tool_installed")
		} else {
			reasons = append(reasons, "tool_not_installed")
		}

		// Simple cost penalty
		cost := 1.0 // default
		score += (10.0 - cost)
		if cost < 3.0 {
			reasons = append(reasons, "low_resource_cost")
		}

		scored = append(scored, ScoredCandidate{
			inst:    inst,
			score:   score,
			reasons: reasons,
		})
	}

	if len(scored) == 0 {
		return plan, fmt.Errorf("no compatible instruments found for capability %s in current workspace", req.Capability)
	}

	// 3. Sort candidates by score descending
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	best := scored[0]
	bestIdent := best.inst.Identity()
	plan.Selected = bestIdent.Name
	plan.SelectionReason = best.reasons
	plan.ExpectedCost = 1.0
	plan.Identity = &bestIdent

	// Populate alternatives
	for i := 1; i < len(scored); i++ {
		plan.Alternatives = append(plan.Alternatives, scored[i].inst.Identity().Name)
	}

	return plan, nil
}
