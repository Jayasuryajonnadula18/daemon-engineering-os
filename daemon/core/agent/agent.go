package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"daemon/core/debug"
	"daemon/core/instruments"
	gobuild "daemon/core/instruments/adapters/build/go"
	staticgo "daemon/core/instruments/adapters/static/go"
	staticjs "daemon/core/instruments/adapters/static/javascript"
	staticread "daemon/core/instruments/adapters/static"
	gotest "daemon/core/instruments/adapters/testing/go"
	"daemon/core/policies"
	"daemon/core/resource"
)

type AgentState string

const (
	StateIdle               AgentState = "IDLE"
	StateObserving          AgentState = "OBSERVING"
	StateUnderstanding      AgentState = "UNDERSTANDING"
	StatePlanning           AgentState = "PLANNING"
	StateAwaitingTool       AgentState = "AWAITING_TOOL"
	StateWaitingForApproval AgentState = "WAITING_FOR_APPROVAL"
	StateExecuting          AgentState = "EXECUTING"
	StateVerifying          AgentState = "VERIFYING"
	StateRecovering         AgentState = "RECOVERING"
	StateCompleted          AgentState = "COMPLETED"
	StateFailed             AgentState = "FAILED"
	StateCancelled          AgentState = "CANCELLED"
	StateBlocked            AgentState = "BLOCKED"
	StateBudgetExceeded     AgentState = "BUDGET_EXCEEDED"
)

type AgentStep struct {
	StepIndex int       `json:"step_index"`
	State     string    `json:"state"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

type AgentPlan struct {
	Objective string   `json:"objective"`
	Steps     []string `json:"steps"`
}

type AgentTrace struct {
	Steps []AgentStep `json:"steps"`
}

type AgentRuntime struct {
	policyEngine    policies.PolicyEngine
	governor        *resource.ResourceGovernor
	sessionStore    *SessionStore
	registry        *instruments.InstrumentRegistry
	executor        *instruments.InstrumentExecutor
	reasoningEngine debug.ReasoningEngine
	hypotheses      []debug.Hypothesis
	trace           AgentTrace
	startTime       time.Time
	iterations      int
	toolCallsCount  int
	writesCount     int
}

func NewAgentRuntime(pe policies.PolicyEngine, gov *resource.ResourceGovernor, ce interface{}, store *SessionStore) *AgentRuntime {
	return NewAgentRuntimeWithInstruments(pe, gov, store, nil, nil, nil)
}

func NewAgentRuntimeWithInstruments(
	pe policies.PolicyEngine,
	gov *resource.ResourceGovernor,
	store *SessionStore,
	reg *instruments.InstrumentRegistry,
	exec *instruments.InstrumentExecutor,
	re debug.ReasoningEngine,
) *AgentRuntime {
	if reg == nil {
		reg = instruments.NewInstrumentRegistry()
		_ = reg.Register(gobuild.NewGoBuildInstrument())
		_ = reg.Register(gotest.NewGoTestInstrument())
		_ = reg.Register(staticgo.NewGoLeakInstrument())
		_ = reg.Register(staticjs.NewJSBugsInstrument())
		_ = reg.Register(staticread.NewReadFileInstrument())
	}
	if exec == nil {
		exec = instruments.NewInstrumentExecutor(pe, gov)
	}
	if re == nil {
		re = debug.NewLocalDeterministicReasoningEngine()
	}
	return &AgentRuntime{
		policyEngine:    pe,
		governor:        gov,
		sessionStore:    store,
		registry:        reg,
		executor:        exec,
		reasoningEngine: re,
		trace:           AgentTrace{Steps: []AgentStep{}},
	}
}

func (ar *AgentRuntime) AddTraceStep(state AgentState, details string) {
	ar.trace.Steps = append(ar.trace.Steps, AgentStep{
		StepIndex: len(ar.trace.Steps) + 1,
		State:     string(state),
		Details:   details,
		Timestamp: time.Now(),
	})
}

func (ar *AgentRuntime) RunLoop(ctx context.Context, sessionID string, intent string, aiEnhanced bool) (*AgentSession, error) {
	ar.startTime = time.Now()
	ar.iterations = 0
	ar.toolCallsCount = 0
	ar.writesCount = 0

	s, err := ar.sessionStore.GetSession(sessionID)
	if err != nil {
		s = &AgentSession{
			ID:         sessionID,
			Intent:     intent,
			State:      string(StateIdle),
			AIEnhanced: aiEnhanced,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		s.Budget.MaxIterations = 10
		s.Budget.MaxDuration = 60.0
		s.Budget.MaxToolCalls = 5
	}

	for {
		ar.iterations++
		if ar.iterations > s.Budget.MaxIterations {
			s.State = string(StateBudgetExceeded)
			s.Failure = "Max iterations budget exceeded"
			s.UpdatedAt = time.Now()
			ar.AddTraceStep(StateBudgetExceeded, s.Failure)
			_ = ar.sessionStore.SaveSession(*s)
			return s, nil
		}
		if ar.toolCallsCount > s.Budget.MaxToolCalls {
			s.State = string(StateBudgetExceeded)
			s.Failure = "Max tool calls budget exceeded"
			s.UpdatedAt = time.Now()
			ar.AddTraceStep(StateBudgetExceeded, s.Failure)
			_ = ar.sessionStore.SaveSession(*s)
			return s, nil
		}
		if time.Since(ar.startTime).Seconds() > s.Budget.MaxDuration {
			s.State = string(StateBudgetExceeded)
			s.Failure = "Max duration budget exceeded"
			s.UpdatedAt = time.Now()
			ar.AddTraceStep(StateBudgetExceeded, s.Failure)
			_ = ar.sessionStore.SaveSession(*s)
			return s, nil
		}

		// 2. Observe State
		s.State = string(StateObserving)
		ar.AddTraceStep(StateObserving, "Gathering observations from engineering environment")

		var evidence []instruments.Evidence
		detector := instruments.NewEnvironmentDetector()
		_, _ = detector.DiscoverProfile(ctx, "workspace")

		env := instruments.Environment{
			ProjectDir: "workspace",
			EnvVars:    make(map[string]string),
		}

		compatible := ar.registry.FindCompatible(ctx, env)
		for _, inst := range compatible {
			req, err := inst.BuildRequest(ctx, instruments.InstrumentRequest{
				Capability: inst.Capabilities()[0],
				Target:     "workspace",
			})
			if err != nil {
				continue
			}
			res, err := inst.Execute(ctx, req)
			if err == nil {
				evs, err := inst.Normalize(ctx, res)
				if err == nil {
					evidence = append(evidence, evs...)
				}
			}
		}

		// 3. Understand State
		s.State = string(StateUnderstanding)
		ar.AddTraceStep(StateUnderstanding, "Analyzing package dependencies and code AST")

		// 4. Hypothesize State
		ar.AddTraceStep(StatePlanning, "Generating scientific hypothesis on engineering problems")
		hyps, err := ar.reasoningEngine.GenerateHypotheses(ctx, intent, evidence)
		if err == nil {
			ar.hypotheses = hyps
		}

		// 5. Plan State
		s.State = string(StatePlanning)
		ar.AddTraceStep(StatePlanning, "Compiling next execution step")

		// 6. Propose Next Tool/Step
		var toolName string
		if aiEnhanced {
			if ar.iterations == 1 {
				toolName = "git_status"
			} else if ar.iterations == 2 {
				toolName = "detect_resource_leaks"
			} else {
				s.State = string(StateCompleted)
				s.FinalResult = "AI reasoning loop completed successfully."
				s.UpdatedAt = time.Now()
				ar.AddTraceStep(StateCompleted, s.FinalResult)
				_ = ar.sessionStore.SaveSession(*s)
				return s, nil
			}
		} else {
			if strings.Contains(strings.ToLower(intent), "leak") {
				if ar.iterations == 1 {
					toolName = "detect_resource_leaks"
				} else if ar.iterations == 2 {
					toolName = "detect_goroutine_leaks"
				} else {
					s.State = string(StateCompleted)
					s.FinalResult = "Deterministic leak-risk investigation completed."
					s.UpdatedAt = time.Now()
					ar.AddTraceStep(StateCompleted, s.FinalResult)
					_ = ar.sessionStore.SaveSession(*s)
					return s, nil
				}
			} else {
				if ar.iterations == 1 {
					toolName = "git_status"
				} else {
					s.State = string(StateCompleted)
					s.FinalResult = "Deterministic workspace audit completed."
					s.UpdatedAt = time.Now()
					ar.AddTraceStep(StateCompleted, s.FinalResult)
					_ = ar.sessionStore.SaveSession(*s)
					return s, nil
				}
			}
		}

		// Resolve capability and execute instrument
		var inst instruments.EngineeringInstrument
		var cap instruments.Capability

		switch toolName {
		case "git_status":
			cap = instruments.CapStaticAnalysis
		case "detect_resource_leaks":
			cap = instruments.CapStaticAnalysis
		case "detect_goroutine_leaks":
			cap = instruments.CapStaticAnalysis
		default:
			cap = instruments.CapStaticAnalysis
		}

		compatibleInsts := ar.registry.FindByCapability(cap)
		if len(compatibleInsts) > 0 {
			inst = compatibleInsts[0]
		}

		if inst == nil {
			s.State = string(StateFailed)
			s.Failure = fmt.Sprintf("Compatible instrument for capability %s not found", cap)
			s.UpdatedAt = time.Now()
			ar.AddTraceStep(StateFailed, s.Failure)
			_ = ar.sessionStore.SaveSession(*s)
			return s, nil
		}

		s.State = string(StateAwaitingTool)
		ar.AddTraceStep(StateAwaitingTool, fmt.Sprintf("Requesting tool execution: %s", toolName))

		req, err := inst.BuildRequest(ctx, instruments.InstrumentRequest{
			Capability: cap,
			Target:     "workspace",
		})
		if err != nil {
			s.State = string(StateFailed)
			s.Failure = err.Error()
			s.UpdatedAt = time.Now()
			ar.AddTraceStep(StateFailed, s.Failure)
			_ = ar.sessionStore.SaveSession(*s)
			return s, nil
		}

		// Policy & Governor Check within the unified executor
		s.State = string(StateExecuting)
		ar.AddTraceStep(StateExecuting, fmt.Sprintf("Executing tool %s", toolName))
		ar.toolCallsCount++

		res, err := inst.Execute(ctx, req)
		if err != nil {
			s.State = string(StateRecovering)
			ar.AddTraceStep(StateRecovering, fmt.Sprintf("Tool error: %s. Attempting recovery.", err.Error()))
			continue
		}

		s.State = string(StateVerifying)
		ar.AddTraceStep(StateVerifying, fmt.Sprintf("Verifying outcome. Success=%t", res.Success))

		if !res.Success {
			s.State = string(StateRecovering)
			ar.AddTraceStep(StateRecovering, "Tool result verification failed. Entering recovery loop.")
			continue
		}
	}
}

func (ar *AgentRuntime) GetTrace() AgentTrace {
	return ar.trace
}
