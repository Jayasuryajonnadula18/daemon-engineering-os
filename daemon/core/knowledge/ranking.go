package knowledge

import (
	"sort"
	"strings"

	"daemon/core/memory"
)

// Ranker enforces mandatory Knowledge Ranking order:
// Personal > Project > Organization > Generic
type Ranker struct{}

// NewRanker instantiates a new Knowledge Ranker.
func NewRanker() *Ranker {
	return &Ranker{}
}

// ScopeWeight returns numeric precedence for scope levels.
func ScopeWeight(scope string) int {
	switch strings.ToLower(scope) {
	case "personal":
		return 400
	case "project":
		return 300
	case "organization", "org":
		return 200
	case "generic":
		return 100
	default:
		return 50
	}
}

// RankRecords sorts knowledge records by ScopeWeight (descending), then Confidence (descending).
func (r *Ranker) RankRecords(records []memory.KnowledgeRecord) []memory.KnowledgeRecord {
	sorted := make([]memory.KnowledgeRecord, len(records))
	copy(sorted, records)

	sort.SliceStable(sorted, func(i, j int) bool {
		wI := ScopeWeight(sorted[i].Scope)
		wJ := ScopeWeight(sorted[j].Scope)
		if wI != wJ {
			return wI > wJ
		}
		return sorted[i].Confidence > sorted[j].Confidence
	})

	return sorted
}
