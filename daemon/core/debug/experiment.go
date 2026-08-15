package debug

import "context"

type Experiment interface {
	Name() string
	Cost() float64
	ExpectedInformationGain() float64
	DiscriminationPower() float64
	ResourceImpact() float64
	RequiredCapability() string
	Risk() string
	ApplicableTo() []string
	Execute(ctx context.Context) (ExperimentResult, error)
}

type ExperimentResult struct {
	ExperimentName string     `json:"experiment_name"`
	Success        bool       `json:"success"`
	Output         string     `json:"output"`
	EvidenceIDs    []string   `json:"evidence_ids"`
	Conclusion     Conclusion `json:"conclusion"`
	DurationMs     int64      `json:"duration_ms"`
}

// SelectNextExperiment chooses the experiment that maximizes discrimination power and minimizes cost/impact
func SelectNextExperiment(experiments []Experiment, activeCPU float64) Experiment {
	var best Experiment
	bestScore := -1.0

	for _, exp := range experiments {
		// If CPU is critically high, skip high-resource impact experiments (Governor Gating)
		if activeCPU > 90.0 && exp.ResourceImpact() > 5.0 {
			continue
		}

		// Score: DiscriminationPower * ExpectedInformationGain / Cost
		cost := exp.Cost()
		if cost <= 0 {
			cost = 0.1
		}
		score := (exp.DiscriminationPower() * exp.ExpectedInformationGain()) / cost
		if score > bestScore {
			bestScore = score
			best = exp
		}
	}

	return best
}
