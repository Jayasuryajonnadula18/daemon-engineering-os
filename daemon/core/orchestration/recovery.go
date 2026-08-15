package orchestration

import (
	"fmt"
)

type FailureClass string

const (
	FailureTransient          FailureClass = "TRANSIENT"
	FailurePermanent          FailureClass = "PERMANENT"
	FailurePolicyDenied       FailureClass = "POLICY_DENIED"
	FailureVerificationFailed FailureClass = "VERIFICATION_FAILED"
	FailureNonReversible      FailureClass = "NON_REVERSIBLE"
)

type RecoveryStrategy struct {
	Class        FailureClass `json:"class"`
	CanRetry     bool         `json:"can_retry"`
	CanRollback  bool         `json:"can_rollback"`
	Compensation string       `json:"compensation"`
}

type RecoveryEngine struct{}

func NewRecoveryEngine() *RecoveryEngine {
	return &RecoveryEngine{}
}

// EvaluateRecovery determines whether a node failure can be retried, rolled back, or requires MANUAL_INTERVENTION.
func (re *RecoveryEngine) EvaluateRecovery(node DAGNode, err error) RecoveryStrategy {
	if !node.Reversible {
		return RecoveryStrategy{
			Class:        FailureNonReversible,
			CanRetry:     false,
			CanRollback:  false,
			Compensation: fmt.Sprintf("Manual intervention required: Action '%s' on target '%s' is non-reversible", node.CapabilityName, node.Inputs["target"]),
		}
	}

	return RecoveryStrategy{
		Class:        FailureVerificationFailed,
		CanRetry:     true,
		CanRollback:  true,
		Compensation: "Automatic rollback to previous verified checkpoint",
	}
}
