package orchestration

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type ExecutionIntent struct {
	Objective   string   `json:"objective"`
	Targets     []string `json:"targets"`
	Constraints []string `json:"constraints"`
	Preferences []string `json:"preferences"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type DAGCompiler struct{}

func NewDAGCompiler() *DAGCompiler {
	return &DAGCompiler{}
}

// Compile translates ExecutionIntent into a dependency-ordered ExecutionDAG with plan freshness headers.
func (c *DAGCompiler) Compile(intent ExecutionIntent) (*ExecutionDAG, error) {
	nodes := []DAGNode{
		{
			ID:             "node-1-validate",
			CapabilityName: "workspace_verification",
			Inputs:         map[string]string{"target": "all"},
			Parents:        []string{},
			Locks:          []ResourceLock{{ResourceID: "workspace", Mode: "READ"}},
			RiskLevel:      RiskLow,
			Status:         NodePending,
			Reversible:     true,
			Idempotent:     true,
		},
		{
			ID:             "node-2-build",
			CapabilityName: "docker_build",
			Inputs:         map[string]string{"target": "service-orders"},
			Parents:        []string{"node-1-validate"},
			Locks:          []ResourceLock{{ResourceID: "docker-compose", Mode: "WRITE"}},
			RiskLevel:      RiskMedium,
			Status:         NodePending,
			Reversible:     true,
			Idempotent:     true,
		},
		{
			ID:             "node-3-deploy",
			CapabilityName: "service_deploy",
			Inputs:         map[string]string{"target": "service-orders"},
			Parents:        []string{"node-2-build"},
			Locks:          []ResourceLock{{ResourceID: "docker-compose", Mode: "WRITE"}},
			RiskLevel:      RiskHigh,
			Status:         NodePending,
			Reversible:     true,
			Idempotent:     false,
		},
	}

	freshness := PlanFreshnessHeader{
		PlanHash:          c.computeHash(intent, nodes),
		TwinVersion:       "v1.0.0",
		PolicyVersion:     "v1.0.0",
		CapabilityVersion: "v1.0.0",
		ContextVersion:    "v1.0.0",
		CreatedAt:         time.Now(),
	}

	dag := &ExecutionDAG{
		ID:         fmt.Sprintf("dag-%d", time.Now().UnixNano()),
		Intent:     intent.Objective,
		Freshness:  freshness,
		State:      StateCompiled,
		Nodes:      nodes,
		TotalNodes: len(nodes),
	}

	if err := c.validateAcyclic(dag); err != nil {
		return nil, err
	}

	return dag, nil
}

func (c *DAGCompiler) computeHash(intent ExecutionIntent, nodes []DAGNode) string {
	h := sha256.New()
	h.Write([]byte(intent.Objective))
	for _, n := range nodes {
		h.Write([]byte(n.ID + n.CapabilityName))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (c *DAGCompiler) validateAcyclic(dag *ExecutionDAG) error {
	visited := make(map[string]int) // 0: unvisited, 1: visiting, 2: visited
	nodeMap := make(map[string]DAGNode)
	for _, n := range dag.Nodes {
		nodeMap[n.ID] = n
	}

	var dfs func(id string) error
	dfs = func(id string) error {
		visited[id] = 1
		node := nodeMap[id]
		for _, pID := range node.Parents {
			if visited[pID] == 1 {
				return fmt.Errorf("cycle detected in execution DAG between node %s and parent %s", id, pID)
			}
			if visited[pID] == 0 {
				if err := dfs(pID); err != nil {
					return err
				}
			}
		}
		visited[id] = 2
		return nil
	}

	for _, n := range dag.Nodes {
		if visited[n.ID] == 0 {
			if err := dfs(n.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateFreshness compares plan freshness headers against current twin versions.
func (c *DAGCompiler) ValidateFreshness(dag *ExecutionDAG, currentTwinVersion string) bool {
	return dag.Freshness.TwinVersion == currentTwinVersion
}
