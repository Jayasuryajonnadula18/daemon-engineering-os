package intelligence

import (
	"path/filepath"
	"testing"
)

func TestIntelligenceStateStore_SaveAndRetrieve(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_intel.db")

	store, err := NewIntelligenceStateStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create IntelligenceStateStore: %v", err)
	}
	defer store.Close()

	opp := AutomationOpportunity{
		PatternID:        "pat-101",
		Sequence:         []string{"docker restart", "check logs", "restart api"},
		OccurrencesCount: 12,
		AverageDuration:  "11m",
		Confidence:       0.96,
		OpportunityScore: "HIGH",
	}

	if err := store.SaveOpportunity(opp); err != nil {
		t.Fatalf("failed to save opportunity: %v", err)
	}

	list, err := store.GetOpportunities()
	if err != nil {
		t.Fatalf("failed to retrieve opportunities: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 opportunity record, got %d", len(list))
	}
	if list[0].PatternID != "pat-101" || list[0].OpportunityScore != "HIGH" {
		t.Fatalf("retrieved pattern mismatch: %v", list[0])
	}
}
