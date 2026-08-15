package integration

import (
	"context"
	"fmt"
	"time"

	"daemon/core/datamine"
	"daemon/core/instruments"
)

type DatamineAdapter struct{}

func NewDatamineAdapter() *DatamineAdapter {
	return &DatamineAdapter{}
}

// GatherGitEvidence queries the git log via Datamine and registers commits as evidence
func (da *DatamineAdapter) GatherGitEvidence(ctx context.Context, repoPath string, maxCount int) ([]instruments.Evidence, error) {
	report, err := datamine.MineCommits(ctx, repoPath, maxCount)
	if err != nil {
		return nil, err
	}

	var list []instruments.Evidence
	for cat, count := range report.CategoryCounts {
		list = append(list, instruments.Evidence{
			ID:           fmt.Sprintf("ev-git-datamine-%s", cat),
			Type:         instruments.EvidenceDatamine,
			Source:       "git_datamine",
			EntityID:     repoPath,
			Statement:    fmt.Sprintf("Mined %d commit(s) of category: %s", count, cat),
			ObservedAt:   time.Now(),
			Freshness:    "live",
			Reliability:  0.8,
			Confidence:   0.8,
			Scope:        "repository",
			Quality: instruments.EvidenceQuality{
				Class:           "git_datamine",
				Strength:        0.8,
				Reliability:     0.8,
				Freshness:       1.0,
				Specificity:     0.8,
				Independence:    1.0,
				Reproducibility: 1.0,
				Verification:    "VERIFIED",
				Provenance:      "git_datamine",
			},
		})
	}

	return list, nil
}
