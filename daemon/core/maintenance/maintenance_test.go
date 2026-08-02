package maintenance

import (
	"context"
	"strings"
	"testing"

	engContext "daemon/core/context"
	"daemon/core/domain"
	"daemon/core/policies"
	"daemon/core/storage"
)

type dummyGraphStore struct {
	storage.GraphStore
}

func (d *dummyGraphStore) GetServices() ([]domain.Service, error) {
	return []domain.Service{
		{ID: "srv-1", Name: "orders-api"},
		{ID: "srv-2", Name: "payments-api"},
	}, nil
}
func (d *dummyGraphStore) GetDependencies() ([]domain.Dependency, error) {
	return nil, nil
}

type dummyMemoryStore struct {
	storage.MemoryStore
}

func (d *dummyMemoryStore) GetIncidents() ([]domain.Incident, error) {
	return nil, nil
}
func (d *dummyMemoryStore) GetRecommendations() ([]domain.Recommendation, error) {
	return nil, nil
}
func (d *dummyMemoryStore) GetDeployments() ([]domain.Deployment, error) {
	return nil, nil
}

func TestMaintenanceEngine_RunMaintenance(t *testing.T) {
	ctx := context.Background()
	ce := engContext.NewContextEngine(&dummyGraphStore{}, &dummyMemoryStore{})
	pe := policies.NewMemoryPolicyEngine(false) // allow mutation for auto-fix test

	me := NewMaintenanceEngine(ce, pe)

	// Test category "containers"
	rep, err := me.RunMaintenance(ctx, "containers", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Category != "containers" {
		t.Errorf("expected category containers, got %s", rep.Category)
	}
	if rep.ConfidenceScore != 96 {
		t.Errorf("expected confidence score 96, got %d", rep.ConfidenceScore)
	}
	if len(rep.Evidence) == 0 {
		t.Errorf("expected evidence items, got none")
	}

	// Test auto-fix self healing execution
	repFix, err := me.RunMaintenance(ctx, "all", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(repFix.Status, "Repaired") {
		t.Errorf("expected Repaired status, got %s", repFix.Status)
	}
}
