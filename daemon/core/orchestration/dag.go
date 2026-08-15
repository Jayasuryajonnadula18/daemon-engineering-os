package orchestration

import (
	"time"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

type NodeStatus string

const (
	NodePending            NodeStatus = "PENDING"
	NodeWaiting            NodeStatus = "WAITING"
	NodeApproved           NodeStatus = "APPROVED"
	NodeRunning            NodeStatus = "RUNNING"
	NodeExecutedUnverified NodeStatus = "EXECUTED_UNVERIFIED"
	NodeVerifying          NodeStatus = "VERIFYING"
	NodeVerified           NodeStatus = "VERIFIED"
	NodeFailed             NodeStatus = "FAILED"
	NodeRolledBack         NodeStatus = "ROLLED_BACK"
	NodeSkipped            NodeStatus = "SKIPPED"
)

type DAGState string

const (
	StateDraft              DAGState = "DRAFT"
	StateCompiled           DAGState = "COMPILED"
	StateValidated          DAGState = "VALIDATED"
	StateStale              DAGState = "STALE"
	StateImpactAnalyzed     DAGState = "IMPACT_ANALYZED"
	StateAwaitingApproval   DAGState = "AWAITING_APPROVAL"
	StateApproved           DAGState = "APPROVED"
	StateRunning            DAGState = "RUNNING"
	StateCancelling         DAGState = "CANCELLING"
	StateVerifying          DAGState = "VERIFYING"
	StateCompleted          DAGState = "COMPLETED"
	StateRecovering         DAGState = "RECOVERING"
	StateRolledBack         DAGState = "ROLLED_BACK"
	StateCompensated        DAGState = "COMPENSATED"
	StateManualIntervention DAGState = "MANUAL_INTERVENTION"
)

type ResourceLock struct {
	ResourceID string `json:"resource_id"`
	Mode       string `json:"mode"` // "READ", "WRITE"
}

type PlanFreshnessHeader struct {
	PlanHash          string    `json:"plan_hash"`
	TwinVersion       string    `json:"twin_version"`
	PolicyVersion     string    `json:"policy_version"`
	CapabilityVersion string    `json:"capability_version"`
	ContextVersion    string    `json:"context_version"`
	CreatedAt         time.Time `json:"created_at"`
}

type DAGNode struct {
	ID             string            `json:"id"`
	CapabilityName string            `json:"capability_name"`
	Inputs         map[string]string `json:"inputs"`
	Parents        []string          `json:"parents"`
	Locks          []ResourceLock    `json:"locks"`
	RiskLevel      RiskLevel         `json:"risk_level"`
	Status         NodeStatus        `json:"status"`
	Reversible     bool              `json:"reversible"`
	Idempotent     bool              `json:"idempotent"`
}

type ExecutionDAG struct {
	ID         string              `json:"id"`
	Intent     string              `json:"intent"`
	Freshness  PlanFreshnessHeader `json:"freshness"`
	State      DAGState            `json:"state"`
	Nodes      []DAGNode           `json:"nodes"`
	TotalNodes int                 `json:"total_nodes"`
}
