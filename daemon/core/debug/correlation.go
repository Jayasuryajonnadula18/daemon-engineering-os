package debug

import "strings"

type CorrelationType string

const (
	CorrelationAgreement     CorrelationType = "AGREEMENT"
	CorrelationContradiction CorrelationType = "CONTRADICTION"
	CorrelationDuplicate     CorrelationType = "DUPLICATE"
	CorrelationDerived       CorrelationType = "DERIVED_FROM_SAME_SOURCE"
	CorrelationIndependent   CorrelationType = "INDEPENDENT_CONFIRMATION"
	CorrelationStale         CorrelationType = "STALE"
)

type CorrelatedGroup struct {
	EvidenceIDs []string
	Type        CorrelationType
	Summary     string
}

// Correlate evaluates a set of evidence to find agreements, contradictions, duplicates, and independent confirmations.
func Correlate(evidence []Evidence) []CorrelatedGroup {
	var groups []CorrelatedGroup
	if len(evidence) < 2 {
		return groups
	}

	for i := 0; i < len(evidence); i++ {
		for j := i + 1; j < len(evidence); j++ {
			e1 := evidence[i]
			e2 := evidence[j]

			// 1. Check for duplicates/derived
			if e1.Source == e2.Source && e1.EntityID == e2.EntityID {
				groups = append(groups, CorrelatedGroup{
					EvidenceIDs: []string{e1.ID, e2.ID},
					Type:        CorrelationDerived,
					Summary:     "Evidence items are derived from the same source component change.",
				})
				continue
			}

			// 2. Check for Independent Confirmation
			if e1.Type != e2.Type && e1.Scope == e2.Scope && e1.EntityID == e2.EntityID {
				groups = append(groups, CorrelatedGroup{
					EvidenceIDs: []string{e1.ID, e2.ID},
					Type:        CorrelationIndependent,
					Summary:     "Independent confirmation across code analysis and runtime observations.",
				})
				continue
			}

			// 3. Check for Contradiction (e.g. compiler status or test outcome opposite statements)
			isE1Failed := containsAny(e1.Statement, "fail", "error", "leak", "broken")
			isE2Failed := containsAny(e2.Statement, "fail", "error", "leak", "broken")
			isE1Passed := containsAny(e1.Statement, "pass", "success", "clean", "ok")
			isE2Passed := containsAny(e2.Statement, "pass", "success", "clean", "ok")

			if (isE1Failed && isE2Passed) || (isE1Passed && isE2Failed) {
				groups = append(groups, CorrelatedGroup{
					EvidenceIDs: []string{e1.ID, e2.ID},
					Type:        CorrelationContradiction,
					Summary:     "Contradicting statements between evidence sources.",
				})
			} else if (isE1Failed && isE2Failed) || (isE1Passed && isE2Passed) {
				groups = append(groups, CorrelatedGroup{
					EvidenceIDs: []string{e1.ID, e2.ID},
					Type:        CorrelationAgreement,
					Summary:     "Supporting agreement in evidence outcomes.",
				})
			}
		}
	}

	return groups
}

func containsAny(s string, keywords ...string) bool {
	lower := strings.ToLower(s)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
