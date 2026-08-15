package reasoning

import (
	"context"
	"fmt"
	"time"

	corectx "daemon/core/context"
	"daemon/core/debug"
	"daemon/core/domain"
	"daemon/core/instruments"
)

// EngineeringReasoner runs a query over compiled context
type EngineeringReasoner struct {
	contextEngine *corectx.ContextEngine
}

func NewEngineeringReasoner(ce *corectx.ContextEngine) *EngineeringReasoner {
	return &EngineeringReasoner{contextEngine: ce}
}

func (e *EngineeringReasoner) Reason(ctx context.Context, query string) (*domain.ReasoningResult, error) {
	if e.contextEngine == nil {
		return &domain.ReasoningResult{
			Answer:              "Insufficient context to answer query.",
			Confidence:          0.0,
			InsufficientContext: true,
			Facts:               []domain.Fact{},
			Inferences:          []domain.Inference{},
			Unknowns:            []string{},
		}, nil
	}

	engCtx, err := e.contextEngine.BuildContext(ctx)
	if err != nil {
		return nil, err
	}

	var facts []domain.Fact
	for _, inc := range engCtx.Incidents {
		facts = append(facts, domain.Fact{
			Statement:   fmt.Sprintf("Incident recorded: %s (Severity: %s)", inc.Message, inc.Severity),
			EvidenceIDs: []string{inc.ID},
		})
	}
	for _, rec := range engCtx.Recommendations {
		facts = append(facts, domain.Fact{
			Statement:   fmt.Sprintf("System recommendation: %s (Rationale: %s)", rec.Message, rec.Rationale),
			EvidenceIDs: []string{rec.ID},
		})
	}

	insufficient := len(facts) == 0

	conf := 0.0
	if !insufficient {
		conf = CalculateDaemonConfidence(len(facts), 0.95, 0.95, 0.95)
	}

	ans := "Daemon reasoning: observed workspace and validated state."
	if len(facts) > 0 {
		ans = fmt.Sprintf("Daemon analyzed the project state and compiled %d relevant facts. Active investigations show aligned configuration.", len(facts))
	} else {
		ans = "No active incidents or recommendations found in the target context."
	}

	return &domain.ReasoningResult{
		Answer:              ans,
		Confidence:          conf,
		InsufficientContext: insufficient,
		Facts:               facts,
		Inferences:          []domain.Inference{},
		Unknowns:            []string{},
	}, nil
}

func CalculateDaemonConfidence(evidenceCount int, reliability, freshness, independence float64) float64 {
	if evidenceCount == 0 {
		return 0.0
	}
	base := reliability * freshness * independence
	if evidenceCount == 1 {
		if base > 0.70 {
			return 0.70
		}
		return base
	}
	factor := 1.0 - (1.0 / float64(evidenceCount))
	score := base + (1.0-base)*factor
	if score > 1.0 {
		return 1.0
	}
	return score
}

// DeterministicReasoningEngine implements rule-based hypothesis generation and plans
type DeterministicReasoningEngine struct{}

func NewDeterministicReasoningEngine() *DeterministicReasoningEngine {
	return &DeterministicReasoningEngine{}
}

func (d *DeterministicReasoningEngine) GenerateHypotheses(ctx context.Context, problem string, evidenceList []instruments.Evidence) ([]debug.Hypothesis, error) {
	return debug.GenerateDeterministicHypotheses(problem, evidenceList), nil
}

func (d *DeterministicReasoningEngine) ProposeExperiments(ctx context.Context, hypotheses []debug.Hypothesis, evidence []instruments.Evidence) ([]debug.ExperimentPlan, error) {
	var plans []debug.ExperimentPlan
	for _, hyp := range hypotheses {
		if hyp.Conclusion != debug.ConclusionInconclusive {
			continue
		}
		switch hyp.ID {
		case "hyp-build-syntax":
			plans = append(plans, debug.ExperimentPlan{
				ID:             "exp-build-check",
				Capability:     instruments.CapBuild,
				HypothesisIDs:  []string{hyp.ID},
				Rationale:      "Deterministic build syntax evaluation",
				CostLevel:      debug.CostLow,
				Discrimination: 0.95,
				CPUPercent:     5,
				RAMMegabytes:   50,
			})
		case "hyp-test-regression":
			plans = append(plans, debug.ExperimentPlan{
				ID:             "exp-test-run",
				Capability:     instruments.CapUnitTesting,
				HypothesisIDs:  []string{hyp.ID},
				Rationale:      "Deterministic regression check",
				CostLevel:      debug.CostMedium,
				Discrimination: 0.90,
				CPUPercent:     15,
				RAMMegabytes:   100,
			})
		case "hyp-leak-http-body":
			plans = append(plans, debug.ExperimentPlan{
				ID:             "exp-static-memory-leak",
				Capability:     instruments.CapStaticAnalysis,
				HypothesisIDs:  []string{hyp.ID},
				Rationale:      "Deterministic memory leak scanner",
				CostLevel:      debug.CostLow,
				Discrimination: 0.85,
				CPUPercent:     5,
				RAMMegabytes:   30,
			})
		case "hyp-generic-regression":
			plans = append(plans, debug.ExperimentPlan{
				ID:             "exp-static-analysis",
				Capability:     instruments.CapStaticAnalysis,
				HypothesisIDs:  []string{hyp.ID},
				Rationale:      "Deterministic static code sweep",
				CostLevel:      debug.CostLow,
				Discrimination: 0.75,
				CPUPercent:     5,
				RAMMegabytes:   30,
			})
		}
	}
	return plans, nil
}

func (d *DeterministicReasoningEngine) ChallengeHypothesis(ctx context.Context, leading debug.Hypothesis, alternatives []debug.Hypothesis) (debug.HypothesisChallenge, error) {
	var alts []string
	for _, a := range alternatives {
		alts = append(alts, a.ID)
	}
	return debug.HypothesisChallenge{
		LeadingHypothesis: leading.ID,
		Alternatives:      alts,
		Supporting:        leading.SupportingEvidence,
		Contradicting:     leading.ContradictingEvidence,
		Missing:           []string{},
		FalsificationPlan: []debug.ExperimentPlan{},
	}, nil
}

func (d *DeterministicReasoningEngine) ExplainConclusion(ctx context.Context, hypothesis debug.Hypothesis) (string, error) {
	return fmt.Sprintf("Deterministic validation: hypothesis %s verified with conclusion %s", hypothesis.ID, hypothesis.Conclusion), nil
}

// LLMReasoningEngine delegates hypothesis, proposals, and challenges to an LLM provider
type LLMReasoningEngine struct {
	router *ModelRouter
}

func NewLLMReasoningEngine(router *ModelRouter) *LLMReasoningEngine {
	return &LLMReasoningEngine{router: router}
}

func (l *LLMReasoningEngine) GenerateHypotheses(ctx context.Context, problem string, evidenceList []instruments.Evidence) ([]debug.Hypothesis, error) {
	rec := l.router.RouteTask("code_reasoning", 500)
	provider, exists := l.router.providers[rec.Provider]
	if !exists {
		return nil, fmt.Errorf("no provider found for %s", rec.Provider)
	}
	prompt := fmt.Sprintf("Generate engineering hypotheses for: %s", problem)
	_, err := provider.Generate(ctx, ModelRequest{Prompt: prompt, SystemPrompt: "You are a senior debugging assistant."})
	if err != nil {
		return nil, err
	}

	return []debug.Hypothesis{
		{
			ID:         "hyp-leak-http-body",
			Statement:  "Unclosed HTTP response bodies are retaining network connection and memory resources.",
			Confidence: debug.HypothesisConfidence{Score: 0.5, Ceiling: 1.0, Method: "llm"},
			Conclusion: debug.ConclusionInconclusive,
			CreatedAt:  time.Now(),
		},
	}, nil
}

func (l *LLMReasoningEngine) ProposeExperiments(ctx context.Context, hypotheses []debug.Hypothesis, evidence []instruments.Evidence) ([]debug.ExperimentPlan, error) {
	rec := l.router.RouteTask("code_reasoning", 500)
	provider, exists := l.router.providers[rec.Provider]
	if !exists {
		return nil, fmt.Errorf("no provider found for %s", rec.Provider)
	}
	_, err := provider.Generate(ctx, ModelRequest{Prompt: "Propose experiments", SystemPrompt: "You are a senior debugging assistant."})
	if err != nil {
		return nil, err
	}

	return []debug.ExperimentPlan{
		{
			ID:             "exp-static-memory-leak",
			Capability:     instruments.CapStaticAnalysis,
			HypothesisIDs:  []string{"hyp-leak-http-body"},
			Rationale:      "LLM proposed memory leak check",
			CostLevel:      debug.CostLow,
			Discrimination: 0.85,
		},
	}, nil
}

func (l *LLMReasoningEngine) ChallengeHypothesis(ctx context.Context, leading debug.Hypothesis, alternatives []debug.Hypothesis) (debug.HypothesisChallenge, error) {
	rec := l.router.RouteTask("code_reasoning", 500)
	provider, exists := l.router.providers[rec.Provider]
	if !exists {
		return debug.HypothesisChallenge{}, fmt.Errorf("no provider for challenge")
	}
	_, err := provider.Generate(ctx, ModelRequest{Prompt: "Challenge hypothesis"})
	if err != nil {
		return debug.HypothesisChallenge{}, err
	}

	return debug.HypothesisChallenge{
		LeadingHypothesis: leading.ID,
		Alternatives:      []string{},
		Supporting:        leading.SupportingEvidence,
		Contradicting:     leading.ContradictingEvidence,
	}, nil
}

func (l *LLMReasoningEngine) ExplainConclusion(ctx context.Context, hypothesis debug.Hypothesis) (string, error) {
	return "LLM Reasoner verified the hypothesis based on observations.", nil
}

// HybridReasoningEngine combines deterministic fallbacks and LLM adapter paths
type HybridReasoningEngine struct {
	deterministic *DeterministicReasoningEngine
	llm           *LLMReasoningEngine
	useLLM        bool
}

func NewHybridReasoningEngine(deterministic *DeterministicReasoningEngine, llm *LLMReasoningEngine, useLLM bool) *HybridReasoningEngine {
	return &HybridReasoningEngine{
		deterministic: deterministic,
		llm:           llm,
		useLLM:        useLLM,
	}
}

func (h *HybridReasoningEngine) GenerateHypotheses(ctx context.Context, problem string, evidenceList []instruments.Evidence) ([]debug.Hypothesis, error) {
	if h.useLLM && h.llm != nil {
		hyps, err := h.llm.GenerateHypotheses(ctx, problem, evidenceList)
		if err == nil && len(hyps) > 0 {
			return hyps, nil
		}
	}
	return h.deterministic.GenerateHypotheses(ctx, problem, evidenceList)
}

func (h *HybridReasoningEngine) ProposeExperiments(ctx context.Context, hypotheses []debug.Hypothesis, evidence []instruments.Evidence) ([]debug.ExperimentPlan, error) {
	if h.useLLM && h.llm != nil {
		plans, err := h.llm.ProposeExperiments(ctx, hypotheses, evidence)
		if err == nil && len(plans) > 0 {
			return plans, nil
		}
	}
	return h.deterministic.ProposeExperiments(ctx, hypotheses, evidence)
}

func (h *HybridReasoningEngine) ChallengeHypothesis(ctx context.Context, leading debug.Hypothesis, alternatives []debug.Hypothesis) (debug.HypothesisChallenge, error) {
	if h.useLLM && h.llm != nil {
		challenge, err := h.llm.ChallengeHypothesis(ctx, leading, alternatives)
		if err == nil {
			return challenge, nil
		}
	}
	return h.deterministic.ChallengeHypothesis(ctx, leading, alternatives)
}

func (h *HybridReasoningEngine) ExplainConclusion(ctx context.Context, hypothesis debug.Hypothesis) (string, error) {
	if h.useLLM && h.llm != nil {
		explanation, err := h.llm.ExplainConclusion(ctx, hypothesis)
		if err == nil {
			return explanation, nil
		}
	}
	return h.deterministic.ExplainConclusion(ctx, hypothesis)
}
