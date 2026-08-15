package integration

import (
	"fmt"
	"strings"
	"time"

	"daemon/core/evolution"
	"daemon/core/instruments"
)

type HistoryAdapter struct {
	ledger *evolution.FixLedger
}

func NewHistoryAdapter(ledger *evolution.FixLedger) *HistoryAdapter {
	return &HistoryAdapter{ledger: ledger}
}

// GatherHistoryEvidence checks the FixLedger for historically successful fixes
func (ha *HistoryAdapter) GatherHistoryEvidence(errorSignature string) ([]instruments.Evidence, error) {
	if ha.ledger == nil {
		return nil, nil
	}

	entries, err := ha.ledger.GetEntries()
	if err != nil {
		return nil, err
	}

	var list []instruments.Evidence
	lowerSig := strings.ToLower(errorSignature)
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.ErrorSignature), lowerSig) || strings.Contains(strings.ToLower(entry.RootCause), lowerSig) {
			list = append(list, instruments.Evidence{
				ID:           fmt.Sprintf("ev-hist-fix-%s", entry.ActionID),
				Type:         instruments.EvidenceHistory,
				Source:       "fix_ledger",
				EntityID:     entry.ActionID,
				Statement:    fmt.Sprintf("Historically resolved incident with fix: %s (Root cause: %s, Success: %s)", entry.FixSummary, entry.RootCause, entry.VerificationResult),
				ObservedAt:   time.Now(),
				Freshness:    "historical",
				Reliability:  0.8,
				Confidence:   0.8,
				Scope:        "history",
				Quality: instruments.EvidenceQuality{
					Class:           "fix_ledger",
					Strength:        0.8,
					Reliability:     0.8,
					Freshness:       0.5,
					Specificity:     0.8,
					Independence:    1.0,
					Reproducibility: 1.0,
					Verification:    "VERIFIED",
					Provenance:      "fix_ledger",
				},
			})
		}
	}

	return list, nil
}
