package knowledge

import (
	"testing"

	"daemon/core/memory"
)

func TestKnowledgeRanking_PersonalOutranksGeneric(t *testing.T) {
	ranker := NewRanker()

	genericRec := memory.KnowledgeRecord{
		ID:             "gen-1",
		ErrorSignature: "ENV_MISSING",
		Fix:            "Generic fix recommendation",
		Scope:          "generic",
		Confidence:     0.95,
	}

	personalRec := memory.KnowledgeRecord{
		ID:             "pers-1",
		ErrorSignature: "ENV_MISSING",
		Fix:            "Personal tested fix",
		Scope:          "personal",
		Confidence:     0.90, // Even with slightly lower confidence, personal outranks generic
	}

	records := []memory.KnowledgeRecord{genericRec, personalRec}
	ranked := ranker.RankRecords(records)

	if len(ranked) != 2 {
		t.Fatalf("expected 2 ranked records, got %d", len(ranked))
	}

	if ranked[0].ID != "pers-1" {
		t.Fatalf("expected Personal record 'pers-1' to be ranked #1, got '%s'", ranked[0].ID)
	}
	if ranked[1].ID != "gen-1" {
		t.Fatalf("expected Generic record 'gen-1' to be ranked #2, got '%s'", ranked[1].ID)
	}
}
