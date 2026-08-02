package policies

import (
	"context"
	"strings"
)

// PolicyDecision represents the access permission result.
type PolicyDecision string

const (
	DecAllow    PolicyDecision = "allow"
	DecDeny     PolicyDecision = "deny"
	DecConfirm  PolicyDecision = "confirm"
	DecReadOnly PolicyDecision = "read_only"
)

// PolicyEngine validates if runtime execution actions conform to safety policies.
type PolicyEngine interface {
	Evaluate(ctx context.Context, action string, target string) (PolicyDecision, error)
}

// MemoryPolicyEngine stores basic configuration constraints in-memory.
type MemoryPolicyEngine struct {
	readOnly bool
}

// NewMemoryPolicyEngine instantiates a new MemoryPolicyEngine.
func NewMemoryPolicyEngine(readOnly bool) *MemoryPolicyEngine {
	return &MemoryPolicyEngine{readOnly: readOnly}
}

// Evaluate checks if the action requires confirmation, is denied, or is allowed.
func (p *MemoryPolicyEngine) Evaluate(ctx context.Context, action string, target string) (PolicyDecision, error) {
	if p.readOnly {
		return DecReadOnly, nil
	}

	actionLower := strings.ToLower(action)
	if strings.Contains(actionLower, "delete") || strings.Contains(actionLower, "kill") || strings.Contains(actionLower, "overwrite") || strings.Contains(actionLower, "fix") {
		return DecConfirm, nil
	}

	return DecAllow, nil
}
