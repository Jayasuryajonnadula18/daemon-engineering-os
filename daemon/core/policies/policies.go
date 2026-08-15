package policies

import (
	"context"
	"strings"

	"daemon/core/capabilities"
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
	targetLower := strings.ToLower(target)

	// Hard Ceiling 1: Force push is NON-NEGOTIABLE DENIED / MANDATORY CONFIRM
	if strings.Contains(actionLower, "force_push") || strings.Contains(actionLower, "push_force") || strings.Contains(actionLower, "push -f") {
		return DecDeny, nil
	}

	// Hard Ceiling 2: Production writes DENIED
	if (targetLower == "production" || targetLower == "prod" || targetLower == "staging") && (strings.Contains(actionLower, "write") || strings.Contains(actionLower, "delete") || strings.Contains(actionLower, "deploy")) {
		return DecDeny, nil
	}

	// Hard Ceiling 3: Secret / token rotation requires mandatory confirmation
	if strings.Contains(actionLower, "secret_rotate") || strings.Contains(actionLower, "token_rotate") || strings.Contains(actionLower, "rotate_secret") {
		return DecConfirm, nil
	}

	if strings.Contains(actionLower, "delete") || strings.Contains(actionLower, "kill") || strings.Contains(actionLower, "overwrite") || strings.Contains(actionLower, "fix") {
		return DecConfirm, nil
	}

	return DecAllow, nil
}

// EvaluateCapability applies hard safety ceilings to capability execution.
func (p *MemoryPolicyEngine) EvaluateCapability(ctx context.Context, capabilityName string, target string, risk string) (PolicyDecision, error) {
	if p.readOnly {
		return DecReadOnly, nil
	}

	riskLower := strings.ToLower(risk)
	targetLower := strings.ToLower(target)

	// Hard Ceiling 4: High risk actions targeting production CANNOT be auto-approved
	if (targetLower == "production" || targetLower == "prod") && riskLower == string(capabilities.RiskHigh) {
		return DecDeny, nil
	}

	if (targetLower == "production" || targetLower == "prod") && riskLower == string(capabilities.RiskMedium) {
		return DecConfirm, nil
	}

	if riskLower == string(capabilities.RiskHigh) {
		return DecConfirm, nil
	}

	return p.Evaluate(ctx, capabilityName, target)
}
