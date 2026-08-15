package integration

import (
	"fmt"
	"time"

	"daemon/core/analysis"
	"daemon/core/instruments"
)

type FindingsAdapter struct {
	pipeline *analysis.DeepAnalyzerPipeline
}

func NewFindingsAdapter(pipeline *analysis.DeepAnalyzerPipeline) *FindingsAdapter {
	return &FindingsAdapter{pipeline: pipeline}
}

// GatherFindingsEvidence runs a cached or fast analysis scan to list findings as evidence
func (fa *FindingsAdapter) GatherFindingsEvidence(projectDir string) ([]instruments.Evidence, error) {
	if fa.pipeline == nil {
		return nil, nil
	}

	// We can run a quick analysis check (using deep=false)
	res, err := fa.pipeline.RunAnalysis(nil, projectDir, false)
	if err != nil {
		return nil, err
	}

	var list []instruments.Evidence
	for _, f := range res.Findings {
		list = append(list, instruments.Evidence{
			ID:           fmt.Sprintf("ev-finding-%s", f.ID),
			Type:         instruments.EvidenceHistory,
			Source:       "finding_ledger",
			EntityID:     f.ID,
			Statement:    fmt.Sprintf("Finding '%s' [%s]: %s", f.Title, f.Severity, f.Description),
			ObservedAt:   time.Now(),
			Freshness:    "live",
			Reliability:  0.9,
			Confidence:   f.Confidence,
			Scope:        string(f.Category),
			Quality: instruments.EvidenceQuality{
				Class:           "finding_ledger",
				Strength:        0.9,
				Reliability:     0.9,
				Freshness:       1.0,
				Specificity:     0.9,
				Independence:    1.0,
				Reproducibility: 1.0,
				Verification:    "VERIFIED",
				Provenance:      "finding_ledger",
			},
		})
	}

	return list, nil
}
