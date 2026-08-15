package integration

import (
	"fmt"
	"time"

	"daemon/core/graph"
	"daemon/core/instruments"
)

type GraphAdapter struct {
	kg *graph.KnowledgeGraph
}

func NewGraphAdapter(kg *graph.KnowledgeGraph) *GraphAdapter {
	return &GraphAdapter{kg: kg}
}

// GatherGraphEvidence queries the Knowledge Graph to find dependencies and blast radius impact
func (ga *GraphAdapter) GatherGraphEvidence(targetID string) ([]instruments.Evidence, error) {
	if ga.kg == nil {
		return nil, nil
	}

	deps, err := ga.kg.FindUpstreamDependencies(targetID)
	if err != nil {
		return nil, err
	}

	impact, err := ga.kg.FindDownstreamImpact(targetID)
	if err != nil {
		return nil, err
	}

	var list []instruments.Evidence
	for _, dep := range deps {
		list = append(list, instruments.Evidence{
			ID:           fmt.Sprintf("ev-kg-dep-%s", dep),
			Type:         instruments.EvidenceKG,
			Source:       "knowledge_graph",
			EntityID:     dep,
			Statement:    fmt.Sprintf("Component '%s' is an upstream dependency of target '%s'", dep, targetID),
			ObservedAt:   time.Now(),
			Freshness:    "live",
			Reliability:  0.9,
			Confidence:   0.9,
			Scope:        "dependency",
			Quality: instruments.EvidenceQuality{
				Class:           "knowledge_graph",
				Strength:        0.9,
				Reliability:     0.9,
				Freshness:       1.0,
				Specificity:     0.9,
				Independence:    1.0,
				Reproducibility: 1.0,
				Verification:    "VERIFIED",
				Provenance:      "knowledge_graph",
			},
		})
	}

	for _, imp := range impact {
		list = append(list, instruments.Evidence{
			ID:           fmt.Sprintf("ev-kg-impact-%s", imp),
			Type:         instruments.EvidenceKG,
			Source:       "knowledge_graph",
			EntityID:     imp,
			Statement:    fmt.Sprintf("Component '%s' is in the downstream blast radius of target '%s'", imp, targetID),
			ObservedAt:   time.Now(),
			Freshness:    "live",
			Reliability:  0.9,
			Confidence:   0.9,
			Scope:        "impact",
			Quality: instruments.EvidenceQuality{
				Class:           "knowledge_graph",
				Strength:        0.9,
				Reliability:     0.9,
				Freshness:       1.0,
				Specificity:     0.9,
				Independence:    1.0,
				Reproducibility: 1.0,
				Verification:    "VERIFIED",
				Provenance:      "knowledge_graph",
			},
		})
	}

	return list, nil
}
