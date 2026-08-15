package integration

import (
	"context"
	"fmt"
	"time"

	"daemon/core/instruments"
	"daemon/core/twin"
)

type TwinAdapter struct {
	model *twin.TwinModel
}

func NewTwinAdapter(model *twin.TwinModel) *TwinAdapter {
	return &TwinAdapter{model: model}
}

// GatherTwinEvidence searches the twin model for matches to localize the problem area
func (ta *TwinAdapter) GatherTwinEvidence(ctx context.Context, query string) ([]instruments.Evidence, error) {
	if ta.model == nil {
		return nil, nil
	}

	results, err := ta.model.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	var list []instruments.Evidence
	for i, r := range results {
		list = append(list, instruments.Evidence{
			ID:           fmt.Sprintf("ev-twin-match-%d", i),
			Type:         instruments.EvidenceTwin,
			Source:       "twin_model",
			EntityID:     r.ID,
			Statement:    fmt.Sprintf("Matched twin entity '%s' (%s): %s", r.Name, r.Type, r.Context),
			ObservedAt:   time.Now(),
			Freshness:    "live",
			Reliability:  0.85,
			Confidence:   0.85,
			Scope:        r.Type,
			Quality: instruments.EvidenceQuality{
				Class:           "twin_model",
				Strength:        0.85,
				Reliability:     0.85,
				Freshness:       1.0,
				Specificity:     0.85,
				Independence:    1.0,
				Reproducibility: 1.0,
				Verification:    "VERIFIED",
				Provenance:      "twin_model",
			},
		})
	}

	return list, nil
}
