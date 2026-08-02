package advisor

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
	return nil, nil
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

func TestAdvisorEngine(t *testing.T) {
	gs := &DummyGraphStore{}
	ms := &DummyMemoryStore{}

	ce := engContext.NewContextEngine(gs, ms)
	ae := NewAdvisorEngine(ce)
	report, err := ae.Advise(context.Background(), "security", "", "")
	if err != nil {
		t.Fatalf("unexpected advisor error: %v", err)
	}

	if report.HealthScore != 98 {
		t.Fatalf("expected health score 98 for security, got %d", report.HealthScore)
	}

	if len(report.Recommendations) == 0 {
		t.Fatalf("expected recommendations list to not be empty")
	}
}
