package debug

import (
	"context"
	"fmt"

	"daemon/core/instruments"
)

type Evidence = instruments.Evidence
type EvidenceType = instruments.EvidenceType

const (
	EvidenceGit           = instruments.EvidenceGit
	EvidenceCode          = instruments.EvidenceCode
	EvidenceSourceCode    = instruments.EvidenceSourceCode
	EvidenceAST           = instruments.EvidenceAST
	EvidenceBuild         = instruments.EvidenceBuild
	EvidenceTest          = instruments.EvidenceTest
	EvidenceLog           = instruments.EvidenceLog
	EvidenceMetric        = instruments.EvidenceMetric
	EvidenceHistory       = instruments.EvidenceHistory
	EvidenceDatamine      = instruments.EvidenceDatamine
	EvidenceModel         = instruments.EvidenceModel
	EvidenceRuntime       = instruments.EvidenceRuntime
	EvidenceProcess       = instruments.EvidenceProcess
	EvidencePort          = instruments.EvidencePort
	EvidenceDependency    = instruments.EvidenceDependency
	EvidenceConfiguration = instruments.EvidenceConfiguration
	EvidenceKG            = instruments.EvidenceKG
	EvidenceTwin          = instruments.EvidenceTwin
	EvidenceEvent         = instruments.EvidenceEvent
	EvidenceWorkflow      = instruments.EvidenceWorkflow
)

func RedactSecrets(input string) string {
	redacted, _ := instruments.RedactSecrets(input)
	return redacted
}

type ReasoningEngine interface {
	GenerateHypotheses(ctx context.Context, problem string, evidenceList []instruments.Evidence) ([]Hypothesis, error)
	ProposeExperiments(ctx context.Context, hypotheses []Hypothesis, evidence []instruments.Evidence) ([]ExperimentPlan, error)
	ChallengeHypothesis(ctx context.Context, leading Hypothesis, alternatives []Hypothesis) (HypothesisChallenge, error)
	ExplainConclusion(ctx context.Context, hypothesis Hypothesis) (string, error)
}

type LocalDeterministicReasoningEngine struct{}

func NewLocalDeterministicReasoningEngine() *LocalDeterministicReasoningEngine {
	return &LocalDeterministicReasoningEngine{}
}

func (d *LocalDeterministicReasoningEngine) GenerateHypotheses(ctx context.Context, problem string, evidenceList []instruments.Evidence) ([]Hypothesis, error) {
	return GenerateDeterministicHypotheses(problem, evidenceList), nil
}

func (d *LocalDeterministicReasoningEngine) ProposeExperiments(ctx context.Context, hypotheses []Hypothesis, evidence []instruments.Evidence) ([]ExperimentPlan, error) {
	var plans []ExperimentPlan
	for _, hyp := range hypotheses {
		if hyp.Conclusion != ConclusionInconclusive {
			continue
		}
		switch hyp.ID {
		case "hyp-build-syntax":
			plans = append(plans, ExperimentPlan{
				ID:             "exp-build-check",
				Capability:     instruments.CapBuild,
				HypothesisIDs:  []string{hyp.ID},
				Rationale:      "Deterministic build syntax evaluation",
				CostLevel:      CostLow,
				Discrimination: 0.95,
				CPUPercent:     5,
				RAMMegabytes:   50,
			})
		case "hyp-test-regression":
			plans = append(plans, ExperimentPlan{
				ID:             "exp-test-run",
				Capability:     instruments.CapUnitTesting,
				HypothesisIDs:  []string{hyp.ID},
				Rationale:      "Deterministic regression check",
				CostLevel:      CostMedium,
				Discrimination: 0.90,
				CPUPercent:     15,
				RAMMegabytes:   100,
			})
		case "hyp-leak-http-body":
			plans = append(plans, ExperimentPlan{
				ID:             "exp-static-memory-leak",
				Capability:     instruments.CapStaticAnalysis,
				HypothesisIDs:  []string{hyp.ID},
				Rationale:      "Deterministic memory leak scanner",
				CostLevel:      CostLow,
				Discrimination: 0.85,
				CPUPercent:     5,
				RAMMegabytes:   30,
			})
		case "hyp-generic-regression":
			plans = append(plans, ExperimentPlan{
				ID:             "exp-static-analysis",
				Capability:     instruments.CapStaticAnalysis,
				HypothesisIDs:  []string{hyp.ID},
				Rationale:      "Deterministic static code sweep",
				CostLevel:      CostLow,
				Discrimination: 0.75,
				CPUPercent:     5,
				RAMMegabytes:   30,
			})
		}
	}
	return plans, nil
}

func (d *LocalDeterministicReasoningEngine) ChallengeHypothesis(ctx context.Context, leading Hypothesis, alternatives []Hypothesis) (HypothesisChallenge, error) {
	var alts []string
	for _, a := range alternatives {
		alts = append(alts, a.ID)
	}
	return HypothesisChallenge{
		LeadingHypothesis: leading.ID,
		Alternatives:      alts,
		Supporting:        leading.SupportingEvidence,
		Contradicting:     leading.ContradictingEvidence,
		Missing:           []string{},
		FalsificationPlan: []ExperimentPlan{},
	}, nil
}

func (d *LocalDeterministicReasoningEngine) ExplainConclusion(ctx context.Context, hypothesis Hypothesis) (string, error) {
	return fmt.Sprintf("Deterministic validation: hypothesis %s verified with conclusion %s", hypothesis.ID, hypothesis.Conclusion), nil
}
