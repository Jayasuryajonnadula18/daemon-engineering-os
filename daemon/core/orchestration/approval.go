package orchestration

import (
	"fmt"
	"time"
)

type ApprovalRequest struct {
	ID          string         `json:"id"`
	ExecutionID string         `json:"execution_id"`
	NodeID      string         `json:"node_id"`
	Reason      string         `json:"reason"`
	RiskLevel   RiskLevel      `json:"risk_level"`
	Impact      ImpactAnalysis `json:"impact"`
	RequestedAt time.Time      `json:"requested_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
}

type ApprovalManager struct {
	requests map[string]*ApprovalRequest
}

func NewApprovalManager() *ApprovalManager {
	return &ApprovalManager{
		requests: make(map[string]*ApprovalRequest),
	}
}

func (am *ApprovalManager) CreateRequest(execID string, nodeID string, risk RiskLevel, impact ImpactAnalysis, reason string) *ApprovalRequest {
	req := &ApprovalRequest{
		ID:          fmt.Sprintf("appr-%d", time.Now().UnixNano()),
		ExecutionID: execID,
		NodeID:      nodeID,
		Reason:      reason,
		RiskLevel:   risk,
		Impact:      impact,
		RequestedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	am.requests[req.ID] = req
	return req
}
