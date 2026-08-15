package testing

import (
	"context"
	"fmt"
	"time"

	"daemon/core/agent"
)

type AgentTestCase struct {
	ID                 string
	Name               string
	Intent             string
	ExpectedState      agent.AgentState
	ExpectedMinSteps   int
	ExpectsSuccess     bool
	InjectLLMFailure   bool
	InjectPolicyDeny   bool
	InjectGovDefer     bool
}

type AgentScenario struct {
	ID          string
	Description string
	Workspace   string
	IncidentID  string
}

type AgentFixture struct {
	ID        string
	Path      string
	Content   string
}

type AgentEvaluation struct {
	TaskSuccess        bool
	EvidenceQuality    float64
	ToolEfficiency     float64
	HallucinationRate  float64
	PolicyViolations   int
	Latency            time.Duration
	Cost               float64
}

type AgentTestingLab struct {
	runtime *agent.AgentRuntime
}

func NewAgentTestingLab(rt *agent.AgentRuntime) *AgentTestingLab {
	return &AgentTestingLab{runtime: rt}
}

func (lab *AgentTestingLab) RunTest(ctx context.Context, tc AgentTestCase) (AgentEvaluation, error) {
	start := time.Now()
	// Execute agent run loop
	sessionID := fmt.Sprintf("test-sess-%s", tc.ID)
	
	// Invoke run loop with deterministic/offline agent execution
	sess, err := lab.runtime.RunLoop(ctx, sessionID, tc.Intent, false)
	if err != nil {
		return AgentEvaluation{TaskSuccess: false}, err
	}

	success := (sess.State == string(tc.ExpectedState)) && (sess.Failure == "")
	if tc.ExpectsSuccess && sess.State == string(agent.StateFailed) {
		success = false
	}

	return AgentEvaluation{
		TaskSuccess:      success,
		EvidenceQuality:  1.0,
		ToolEfficiency:   1.0,
		Latency:          time.Since(start),
		PolicyViolations: 0,
	}, nil
}
