package context

import (
	"context"
	"testing"
)

func TestContextEngine_QueryBoundedContext(t *testing.T) {
	ce := NewContextEngine(nil, nil)
	engCtx, err := ce.QueryBoundedContext(context.Background(), "why is auth failing", ResolutionService, 4000)
	if err != nil {
		t.Fatalf("QueryBoundedContext failed: %v", err)
	}

	if engCtx.Resolution != ResolutionService {
		t.Errorf("expected resolution level %s, got %s", ResolutionService, engCtx.Resolution)
	}

	if !engCtx.Metadata.InsufficientContext {
		t.Errorf("expected InsufficientContext to be true when no entities are present")
	}

	if engCtx.Metadata.MaxTokenBudget != 4000 {
		t.Errorf("expected max token budget 4000, got %d", engCtx.Metadata.MaxTokenBudget)
	}
}
