package reasoning

import (
	"context"
	"testing"

	engContext "daemon/core/context"
	"daemon/core/domain"
	"daemon/core/storage"
)

type OrchDummyGraphStore struct {
	storage.GraphStore
}

func (d *OrchDummyGraphStore) GetServices() ([]domain.Service, error) {
	return []domain.Service{
		{Name: "orders-api", Status: "Running"},
	}, nil
}
func (d *OrchDummyGraphStore) GetDependencies() ([]domain.Dependency, error) {
	return nil, nil
}

type OrchDummyMemoryStore struct {
	storage.MemoryStore
}

func (d *OrchDummyMemoryStore) GetIncidents() ([]domain.Incident, error) {
	return nil, nil
}
func (d *OrchDummyMemoryStore) GetRecommendations() ([]domain.Recommendation, error) {
	return nil, nil
}
func (d *OrchDummyMemoryStore) GetDeployments() ([]domain.Deployment, error) {
	return nil, nil
}

func TestEngineeringOrchestrator(t *testing.T) {
	gs := &OrchDummyGraphStore{}
	ms := &OrchDummyMemoryStore{}

	ce := engContext.NewContextEngine(gs, ms)
	cb := NewContextBuilder(ce)
	mr := NewModelRouter(false)
	orch := NewEngineeringOrchestrator(cb, mr)

	plan, err := orch.Orchestrate(context.Background(), "Deploy orders-api service to production")
	if err != nil {
		t.Fatalf("unexpected orchestration error: %v", err)
	}

	if plan.Domain != "Deployment" {
		t.Errorf("expected Deployment domain, got %s", plan.Domain)
	}

	if !plan.RequiresApproval {
		t.Errorf("expected deployment orchestration to require approval")
	}

	// Verify Dynamic DAG structure compilation from dummy services
	if len(plan.Graph.Nodes) == 0 {
		t.Errorf("expected non-zero dynamic nodes in deployment DAG, got 0")
	}

	if len(plan.Graph.Edges) == 0 {
		t.Errorf("expected non-zero dynamic edges in deployment DAG, got 0")
	}
}
