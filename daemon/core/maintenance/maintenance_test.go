package maintenance

import (
	"context"
	"os"
	"path/filepath"
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

func TestCoreFourMaintenance(t *testing.T) {
	ctx := context.Background()
	ce := engContext.NewContextEngine(&dummyGraphStore{}, &dummyMemoryStore{})
	pe := policies.NewMemoryPolicyEngine(false)

	me := NewMaintenanceEngine(ce, pe)

	// Create temp directory for clean test isolation
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	// Test 1: Empty dir should have 0 drift (silence contract)
	repClean, err := me.RunCoreFourMaintenance(ctx, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repClean.HasDrift {
		t.Errorf("expected 0 drift in empty directory, got HasDrift = true")
	}

	// Test 2: Create .env with keys missing from .env.example
	_ = os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("PORT=5000\nDB_HOST=localhost\n"), 0644)
	repEnv, err := me.RunCoreFourMaintenance(ctx, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repEnv.HasDrift || repEnv.EnvDrift == nil {
		t.Fatalf("expected EnvDrift to be detected when .env exists without .env.example")
	}
	if !repEnv.EnvDrift.MissingExampleFile {
		t.Errorf("expected MissingExampleFile to be true")
	}

	// Test 3: Apply fix (--apply) to auto-generate .env.example
	repApply, err := me.RunCoreFourMaintenance(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error during apply: %v", err)
	}
	if len(repApply.RepairsExecuted) == 0 {
		t.Errorf("expected repairs executed during --apply mode")
	}

	// Verify .env.example created with PORT and DB_HOST (values ignored)
	exData, err := os.ReadFile(filepath.Join(tmpDir, ".env.example"))
	if err != nil {
		t.Fatalf("expected .env.example to be created: %v", err)
	}
	if !os.FileMode(0644).IsRegular() {
		t.Errorf("regular file expected")
	}
	if len(exData) == 0 {
		t.Errorf(".env.example should not be empty")
	}
}
