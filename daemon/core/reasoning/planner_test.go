package reasoning

import (
	"context"
	"testing"

	engContext "daemon/core/context"
	"daemon/core/domain"
	"daemon/core/storage"
)

type DummyGraphStore struct {
	storage.GraphStore
}

func (d *DummyGraphStore) GetServices() ([]domain.Service, error) {
	return []domain.Service{
		{Name: "orders-api", Status: "Running"},
	}, nil
}
func (d *DummyGraphStore) GetDependencies() ([]domain.Dependency, error) {
	return nil, nil
}

type DummyMemoryStore struct {
	storage.MemoryStore
}

func (d *DummyMemoryStore) GetIncidents() ([]domain.Incident, error) {
	return nil, nil
}
func (d *DummyMemoryStore) GetRecommendations() ([]domain.Recommendation, error) {
	return nil, nil
}
func (d *DummyMemoryStore) GetDeployments() ([]domain.Deployment, error) {
	return nil, nil
}

func TestEngineeringPlanner(t *testing.T) {
	gs := &DummyGraphStore{}
	ms := &DummyMemoryStore{}

	ce := engContext.NewContextEngine(gs, ms)
	cb := NewContextBuilder(ce)
	mr := NewModelRouter(false)
	planner := NewEngineeringPlanner(cb, mr)

	plan, err := planner.GeneratePlan(context.Background(), "Deploy orders service")
	if err != nil {
		t.Fatalf("unexpected planner error: %v", err)
	}

	if !plan.RequiresApproval {
		t.Errorf("expected deployment planner to require approval")
	}

	if len(plan.Steps) == 0 {
		t.Errorf("expected steps to be generated")
	}
}
