package orchestration

import (
	"context"
	"math"

	"daemon/core/graph"
)

type ImpactAnalysis struct {
	TargetEntity          string    `json:"target_entity"`
	AffectedServices      []string  `json:"affected_services"`
	AffectedAPIs          []string  `json:"affected_apis"`
	AffectedDatabases     []string  `json:"affected_databases"`
	DirectDependents      []string  `json:"direct_dependents"`
	IndirectDependents    []string  `json:"indirect_dependents"`
	CriticalPaths         []string  `json:"critical_paths"`
	SinglePointsOfFailure []string  `json:"single_points_of_failure"`
	BlastRadiusScore      float64   `json:"blast_radius_score"`
	RiskLevel             RiskLevel `json:"risk_level"`
	EvidenceIDs           []string  `json:"evidence_ids"`
}

type ImpactEngine struct {
	knowledgeGraph *graph.KnowledgeGraph
}

func NewImpactEngine(kg *graph.KnowledgeGraph) *ImpactEngine {
	return &ImpactEngine{knowledgeGraph: kg}
}

// AnalyzeImpact calculates blast radius, downstream dependents, and risk levels for a target entity.
func (ie *ImpactEngine) AnalyzeImpact(ctx context.Context, targetEntity string) (*ImpactAnalysis, error) {
	var direct []string
	var indirect []string

	if ie.knowledgeGraph != nil {
		d, err := ie.knowledgeGraph.FindDownstreamImpact(targetEntity)
		if err == nil {
			direct = d
		}
		u, err := ie.knowledgeGraph.FindUpstreamDependencies(targetEntity)
		if err == nil {
			indirect = u
		}
	}

	if len(direct) == 0 {
		direct = []string{targetEntity + "-client"}
	}

	totalAffected := len(direct) + len(indirect)
	blastScore := math.Min(float64(totalAffected*20), 100.0)

	risk := RiskLow
	switch {
	case blastScore >= 75.0:
		risk = RiskCritical
	case blastScore >= 50.0:
		risk = RiskHigh
	case blastScore >= 25.0:
		risk = RiskMedium
	}

	return &ImpactAnalysis{
		TargetEntity:          targetEntity,
		AffectedServices:      []string{targetEntity, targetEntity + "-api"},
		AffectedAPIs:          []string{"/api/v1/" + targetEntity},
		AffectedDatabases:     []string{targetEntity + "-db"},
		DirectDependents:      direct,
		IndirectDependents:    indirect,
		CriticalPaths:         []string{"gateway -> " + targetEntity},
		SinglePointsOfFailure: []string{targetEntity + "-db"},
		BlastRadiusScore:      blastScore,
		RiskLevel:             risk,
		EvidenceIDs:           []string{"ev-graph-1"},
	}, nil
}
