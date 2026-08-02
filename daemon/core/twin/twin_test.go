package twin

import (
	"context"
	"testing"

	"daemon/core/domain"
	"daemon/core/storage"
)

type MockGraphStore struct {
	storage.GraphStore
	services []domain.Service
	deps     []domain.Dependency
	modules  []domain.Module
}

func (m *MockGraphStore) GetServices() ([]domain.Service, error) {
	return m.services, nil
}
func (m *MockGraphStore) GetDependencies() ([]domain.Dependency, error) {
	return m.deps, nil
}
func (m *MockGraphStore) GetAPIs() ([]domain.API, error) {
	return nil, nil
}
func (m *MockGraphStore) GetNodes(nodeType string) ([]domain.Module, error) {
	if nodeType == "Container" {
		return m.modules, nil
	}
	return nil, nil
}

func TestTwinSearch(t *testing.T) {
	gs := &MockGraphStore{
		services: []domain.Service{
			{ID: "auth", Name: "Auth Service", Port: 5001},
		},
		deps: []domain.Dependency{
			{ID: "lodash", Name: "lodash", Version: "4.17.21"},
		},
		modules: []domain.Module{
			{ID: "cont-pay", Name: "payments-api", Type: "Container"},
		},
	}

	model := NewTwinModel(gs)
	res, err := model.Search(context.Background(), "payment")
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(res))
	}
	if res[0].Name != "payments-api" {
		t.Fatalf("expected search match payments-api, got %s", res[0].Name)
	}
}
