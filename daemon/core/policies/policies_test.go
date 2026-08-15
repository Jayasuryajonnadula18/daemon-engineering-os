package policies

import (
	"context"
	"testing"

	"daemon/core/capabilities"
)

func TestPolicyEngine_HardCeilings(t *testing.T) {
	pe := NewMemoryPolicyEngine(false)
	ctx := context.Background()

	// Ceiling 1: Force push is DENIED
	dec, err := pe.Evaluate(ctx, "git_force_push", "main")
	if err != nil || dec != DecDeny {
		t.Fatalf("ceiling 1 failed: expected force_push to be DENIED, got dec=%s err=%v", dec, err)
	}

	// Ceiling 2: Production writes DENIED
	dec, err = pe.Evaluate(ctx, "write_config", "production")
	if err != nil || dec != DecDeny {
		t.Fatalf("ceiling 2 failed: expected production write to be DENIED, got dec=%s err=%v", dec, err)
	}

	// Ceiling 3: Secret rotation requires confirmation
	dec, err = pe.Evaluate(ctx, "secret_rotate", "local")
	if err != nil || dec != DecConfirm {
		t.Fatalf("ceiling 3 failed: expected secret rotation to require CONFIRM, got dec=%s err=%v", dec, err)
	}

	// Ceiling 4: High risk action in production DENIED
	dec, err = pe.EvaluateCapability(ctx, "deploy_production_cluster", "production", string(capabilities.RiskHigh))
	if err != nil || dec != DecDeny {
		t.Fatalf("ceiling 4 failed: expected high-risk production action to be DENIED, got dec=%s err=%v", dec, err)
	}
}
