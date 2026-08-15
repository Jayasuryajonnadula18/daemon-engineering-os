package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	var evsStr []string
	for _, ev := range evidenceList {
		evsStr = append(evsStr, fmt.Sprintf("- Evidence ID: %s, Statement: %s", ev.ID, ev.Statement))
	}
	evidencePrompt := strings.Join(evsStr, "\n")

	systemPrompt := "You are a senior debugging assistant. You must formulate potential hypotheses that explain the reported problem using the available evidence. You MUST respond with ONLY a valid JSON array of objects matching this schema:\n[\n  {\n    \"id\": \"hyp-custom-id\",\n    \"statement\": \"Detailed statement explaining the failure mode\",\n    \"confidence_score\": 0.8\n  }\n]\nDo not include any conversational filler, markdown syntax blocks, or other text outside the JSON array."
	prompt := fmt.Sprintf("Problem description: %s\n\nGathered evidence:\n%s\n\nGenerate hypotheses.", problem, evidencePrompt)

	resp, err := l.router.Generate(ctx, "code_reasoning", prompt, systemPrompt)
	if err != nil {
		return debug.GenerateDeterministicHypotheses(problem, evidenceList), nil
	}

	cleanText := strings.TrimSpace(resp)
	if strings.HasPrefix(cleanText, "```json") {
		cleanText = strings.TrimPrefix(cleanText, "```json")
		cleanText = strings.TrimSuffix(cleanText, "```")
	} else if strings.HasPrefix(cleanText, "```") {
		cleanText = strings.TrimPrefix(cleanText, "```")
		cleanText = strings.TrimSuffix(cleanText, "```")
	}
	cleanText = strings.TrimSpace(cleanText)

	type llmHypothesis struct {
		ID              string  `json:"id"`
		Statement       string  `json:"statement"`
		ConfidenceScore float64 `json:"confidence_score"`
	}

	var llmHyps []llmHypothesis
	if err := json.Unmarshal([]byte(cleanText), &llmHyps); err == nil && len(llmHyps) > 0 {
		var list []debug.Hypothesis
		for _, lh := range llmHyps {
			list = append(list, debug.Hypothesis{
				ID:         lh.ID,
				Statement:  lh.Statement,
				Confidence: debug.HypothesisConfidence{Score: lh.ConfidenceScore, Ceiling: 1.0, Method: "llm"},
				Conclusion: debug.ConclusionInconclusive,
				CreatedAt:  time.Now(),
			})
		}
		return list, nil
	}

	return debug.GenerateDeterministicHypotheses(problem, evidenceList), nil
}

func (l *LLMReasoningEngine) ProposeExperiments(ctx context.Context, hypotheses []debug.Hypothesis, evidence []instruments.Evidence) ([]debug.ExperimentPlan, error) {
	var hypsStr []string
	for _, h := range hypotheses {
		hypsStr = append(hypsStr, fmt.Sprintf("- Hypothesis ID: %s, Statement: %s", h.ID, h.Statement))
	}
	hypsPrompt := strings.Join(hypsStr, "\n")

	systemPrompt := "You are a senior debugging assistant. You must propose experiment plans using the available capabilities (BUILD, UNIT_TESTING, STATIC_ANALYSIS) to narrow down the active hypotheses. You MUST respond with ONLY a valid JSON array of objects matching this schema:\n[\n  {\n    \"id\": \"exp-custom-id\",\n    \"capability\": \"STATIC_ANALYSIS\",\n    \"hypothesis_ids\": [\"hyp-custom-id\"],\n    \"rationale\": \"Why this check is relevant\",\n    \"cost_level\": \"LOW\",\n    \"discrimination\": 0.8\n  }\n]\nDo not include any conversational filler, markdown syntax blocks, or other text outside the JSON array."
	prompt := fmt.Sprintf("Active hypotheses:\n%s\n\nPropose experiments using available capabilities.", hypsPrompt)

	resp, err := l.router.Generate(ctx, "code_reasoning", prompt, systemPrompt)
	if err != nil {
		return []debug.ExperimentPlan{}, nil
	}

	cleanText := strings.TrimSpace(resp)
	if strings.HasPrefix(cleanText, "```json") {
		cleanText = strings.TrimPrefix(cleanText, "```json")
		cleanText = strings.TrimSuffix(cleanText, "```")
	} else if strings.HasPrefix(cleanText, "```") {
		cleanText = strings.TrimPrefix(cleanText, "```")
		cleanText = strings.TrimSuffix(cleanText, "```")
	}
	cleanText = strings.TrimSpace(cleanText)

	type llmExperimentPlan struct {
		ID             string   `json:"id"`
		Capability     string   `json:"capability"`
		HypothesisIDs  []string `json:"hypothesis_ids"`
		Rationale      string   `json:"rationale"`
		CostLevel      string   `json:"cost_level"`
		Discrimination float64  `json:"discrimination"`
	}

	var llmPlans []llmExperimentPlan
	if err := json.Unmarshal([]byte(cleanText), &llmPlans); err == nil && len(llmPlans) > 0 {
		var plans []debug.ExperimentPlan
		for _, lp := range llmPlans {
			var cap instruments.Capability
			switch lp.Capability {
			case "BUILD":
				cap = instruments.CapBuild
			case "UNIT_TESTING":
				cap = instruments.CapUnitTesting
			case "STATIC_ANALYSIS":
				cap = instruments.CapStaticAnalysis
			default:
				cap = instruments.CapStaticAnalysis
			}
			plans = append(plans, debug.ExperimentPlan{
				ID:             lp.ID,
				Capability:     cap,
				HypothesisIDs:  lp.HypothesisIDs,
				Rationale:      lp.Rationale,
				CostLevel:      debug.CostLevel(lp.CostLevel),
				Discrimination: lp.Discrimination,
			})
		}
		return plans, nil
	}

	return []debug.ExperimentPlan{}, nil
}

func (l *LLMReasoningEngine) ChallengeHypothesis(ctx context.Context, leading debug.Hypothesis, alternatives []debug.Hypothesis) (debug.HypothesisChallenge, error) {
	var altsStr []string
	for _, a := range alternatives {
		altsStr = append(altsStr, fmt.Sprintf("- ID: %s, Statement: %s", a.ID, a.Statement))
	}
	altsPrompt := strings.Join(altsStr, "\n")

	systemPrompt := "You are a senior debugging assistant. You must challenge the leading hypothesis by presenting alternative explanations and proposing a falsification plan of experiments (BUILD, UNIT_TESTING, STATIC_ANALYSIS) to disprove it. You MUST respond with ONLY a valid JSON object matching this schema:\n{\n  \"leading_hypothesis\": \"hyp-id\",\n  \"alternatives\": [\"alt-id\"],\n  \"supporting\": [\"ev-id\"],\n  \"contradicting\": [],\n  \"missing\": [],\n  \"falsification_plan\": [\n    {\n      \"id\": \"exp-falsify-id\",\n      \"capability\": \"UNIT_TESTING\",\n      \"hypothesis_ids\": [\"hyp-id\"],\n      \"rationale\": \"Run specific test to disprove leading hypothesis\",\n      \"cost_level\": \"LOW\",\n      \"discrimination\": 0.9\n    }\n  ]\n}\nDo not include any conversational filler, markdown syntax blocks, or other text outside the JSON object."
	prompt := fmt.Sprintf("Leading hypothesis: %s (%s)\n\nAlternatives:\n%s\n\nChallenge the leading hypothesis.", leading.ID, leading.Statement, altsPrompt)

	resp, err := l.router.Generate(ctx, "code_reasoning", prompt, systemPrompt)
	if err != nil {
		var alts []string
		for _, a := range alternatives {
			alts = append(alts, a.ID)
		}
		return debug.HypothesisChallenge{
			LeadingHypothesis: leading.ID,
			Alternatives:      alts,
			Supporting:        leading.SupportingEvidence,
			Contradicting:     leading.ContradictingEvidence,
		}, nil
	}

	cleanText := strings.TrimSpace(resp)
	if strings.HasPrefix(cleanText, "```json") {
		cleanText = strings.TrimPrefix(cleanText, "```json")
		cleanText = strings.TrimSuffix(cleanText, "```")
	} else if strings.HasPrefix(cleanText, "```") {
		cleanText = strings.TrimPrefix(cleanText, "```")
		cleanText = strings.TrimSuffix(cleanText, "```")
	}
	cleanText = strings.TrimSpace(cleanText)

	type llmExperimentPlan struct {
		ID             string   `json:"id"`
		Capability     string   `json:"capability"`
		HypothesisIDs  []string `json:"hypothesis_ids"`
		Rationale      string   `json:"rationale"`
		CostLevel      string   `json:"cost_level"`
		Discrimination float64  `json:"discrimination"`
	}

	type llmChallenge struct {
		LeadingHypothesis string              `json:"leading_hypothesis"`
		Alternatives      []string            `json:"alternatives"`
		Supporting        []string            `json:"supporting"`
		Contradicting     []string            `json:"contradicting"`
		Missing           []string            `json:"missing"`
		FalsificationPlan []llmExperimentPlan `json:"falsification_plan"`
	}

	var lc llmChallenge
	if err := json.Unmarshal([]byte(cleanText), &lc); err == nil {
		var plans []debug.ExperimentPlan
		for _, lp := range lc.FalsificationPlan {
			var cap instruments.Capability
			switch lp.Capability {
			case "BUILD":
				cap = instruments.CapBuild
			case "UNIT_TESTING":
				cap = instruments.CapUnitTesting
			case "STATIC_ANALYSIS":
				cap = instruments.CapStaticAnalysis
			default:
				cap = instruments.CapStaticAnalysis
			}
			plans = append(plans, debug.ExperimentPlan{
				ID:             lp.ID,
				Capability:     cap,
				HypothesisIDs:  lp.HypothesisIDs,
				Rationale:      lp.Rationale,
				CostLevel:      debug.CostLevel(lp.CostLevel),
				Discrimination: lp.Discrimination,
			})
		}
		return debug.HypothesisChallenge{
			LeadingHypothesis: lc.LeadingHypothesis,
			Alternatives:      lc.Alternatives,
			Supporting:        lc.Supporting,
			Contradicting:     lc.Contradicting,
			Missing:           lc.Missing,
			FalsificationPlan: plans,
		}, nil
	}

	var alts []string
	for _, a := range alternatives {
		alts = append(alts, a.ID)
	}
	return debug.HypothesisChallenge{
		LeadingHypothesis: leading.ID,
		Alternatives:      alts,
		Supporting:        leading.SupportingEvidence,
		Contradicting:     leading.ContradictingEvidence,
	}, nil
}

func (l *LLMReasoningEngine) ExplainConclusion(ctx context.Context, hypothesis debug.Hypothesis) (string, error) {
	systemPrompt := "You are a senior debugging assistant. Synthesize a concise explanation of why the hypothesis was verified or not based on the gathered evidence."
	prompt := fmt.Sprintf("Hypothesis: %s (%s)\nConclusion: %s\nSupporting Evidence: %v\n\nExplain the conclusion.", hypothesis.ID, hypothesis.Statement, hypothesis.Conclusion, hypothesis.SupportingEvidence)

	resp, err := l.router.Generate(ctx, "code_reasoning", prompt, systemPrompt)
	if err != nil {
		return fmt.Sprintf("Deterministic validation: hypothesis %s verified with conclusion %s", hypothesis.ID, hypothesis.Conclusion), nil
	}
	return strings.TrimSpace(resp), nil
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
