package debug

import (
	"fmt"
	"hash/fnv"
	"strings"

	"daemon/core/instruments"
	"daemon/core/resource"
)

// CostLevel classifies the expected resource cost of an experiment.
type CostLevel string

const (
	CostLow    CostLevel = "LOW"
	CostMedium CostLevel = "MEDIUM"
	CostHigh   CostLevel = "HIGH"
)

// FingerprintContext captures all the relevant project/twin state to identify
// if an experiment has already been run on the exact same project revision.
type FingerprintContext struct {
	Capability          instruments.Capability `json:"capability"`
	Target              string                 `json:"target"`
	Parameters          []string               `json:"parameters"`
	RelevantFilesDigest string                 `json:"relevant_files_digest"`
	DependencyDigest    string                 `json:"dependency_digest"`
	TwinVersion         string                 `json:"twin_version"`
	GitRevision         string                 `json:"git_revision"`
}

// ExperimentPlan is the unit of work selected by ExperimentSelector.
// It specifies WHAT should be tested and WHY — not which tool to use.
// Instrument selection is a downstream concern.
type ExperimentPlan struct {
	// ID is a stable identifier for the experiment type (not the instance).
	ID string `json:"id"`

	// Capability is the abstract engineering capability required to run this experiment.
	// The InstrumentSelector uses this to find a compatible installed instrument.
	Capability instruments.Capability `json:"capability"`

	// HypothesisIDs lists the active hypotheses this experiment discriminates between.
	HypothesisIDs []string `json:"hypothesis_ids"`

	// Rationale explains why this experiment was selected.
	Rationale string `json:"rationale"`

	// CostLevel is the expected resource cost bucket.
	CostLevel CostLevel `json:"cost_level"`

	// Discrimination is the expected information gain for distinguishing between
	// active hypotheses (0.0–1.0). Higher = more discriminating.
	Discrimination float64 `json:"discrimination"`

	// CPUPercent and RAMMegabytes are the expected hardware load.
	CPUPercent   float64 `json:"cpu_percent"`
	RAMMegabytes float64 `json:"ram_megabytes"`

	// Fingerprint is a stable deduplication key computed from:
	//   hash(capability + target + relevant parameters + file modtimes + git revisions)
	// If an experiment with the same fingerprint was already executed and the
	// evidence is still fresh, the selector will skip it and reuse the evidence.
	Fingerprint string `json:"fingerprint"`
}

// hypothesisExperimentTemplates maps each known hypothesis ID to its most
// discriminating experiment. These are templates — Fingerprint and HypothesisIDs
// are populated at selection time.
var hypothesisExperimentTemplates = map[string]ExperimentPlan{
	"hyp-build-syntax": {
		ID:             "exp-build-check",
		Capability:     instruments.CapBuild,
		Rationale:      "Compiler errors directly verify or eliminate build failure hypothesis with 100% precision",
		CostLevel:      CostLow,
		Discrimination: 0.95,
		CPUPercent:     5.0,
		RAMMegabytes:   50.0,
	},
	"hyp-test-regression": {
		ID:             "exp-test-run",
		Capability:     instruments.CapUnitTesting,
		Rationale:      "Running existing tests directly verifies or eliminates regression hypothesis",
		CostLevel:      CostMedium,
		Discrimination: 0.90,
		CPUPercent:     15.0,
		RAMMegabytes:   100.0,
	},
	"hyp-leak-http-body": {
		ID:             "exp-static-memory-leak",
		Capability:     instruments.CapStaticAnalysis,
		Rationale:      "Static analysis reveals unclosed resource lifecycle patterns with high specificity",
		CostLevel:      CostLow,
		Discrimination: 0.85,
		CPUPercent:     5.0,
		RAMMegabytes:   30.0,
	},
	"hyp-leak-goroutine": {
		ID:             "exp-goroutine-analysis",
		Capability:     instruments.CapGoroutineAnalysis,
		Rationale:      "Goroutine analysis reveals leaking or stalled goroutines in live process",
		CostLevel:      CostMedium,
		Discrimination: 0.85,
		CPUPercent:     10.0,
		RAMMegabytes:   50.0,
	},
	"hyp-generic-regression": {
		ID:             "exp-static-analysis",
		Capability:     instruments.CapStaticAnalysis,
		Rationale:      "Static analysis identifies code antipatterns and regression indicators",
		CostLevel:      CostLow,
		Discrimination: 0.75,
		CPUPercent:     5.0,
		RAMMegabytes:   30.0,
	},
}

// computeFingerprint creates a stable, collision-resistant deduplication key for an experiment.
// The fingerprint encodes the experiment type, target, and any relevant parameters so that
// re-running the same experiment on the same target is skipped when evidence is still fresh.
func computeFingerprint(ctx FingerprintContext) string {
	h := fnv.New64a()
	params := ""
	if len(ctx.Parameters) > 0 {
		params = strings.Join(ctx.Parameters, ",")
	}
	h.Write([]byte(fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s",
		ctx.Capability, ctx.Target, params,
		ctx.RelevantFilesDigest, ctx.DependencyDigest,
		ctx.TwinVersion, ctx.GitRevision)))
	return fmt.Sprintf("%016x", h.Sum64())
}


// ExperimentSelector selects the highest-value, resource-feasible, non-duplicate
// experiment from the current active hypothesis set.
//
// Design contract:
//   - The selector asks "what experiment gives maximum information per unit of cost?"
//   - It NEVER asks "what tool should we use?" — that is InstrumentSelector's job.
//   - It consults the Resource Governor BEFORE selection, not at execution time.
//   - It enforces deduplication via ExperimentPlan.Fingerprint.
type ExperimentSelector struct {
	governor *resource.ResourceGovernor
}

// NewExperimentSelector creates an ExperimentSelector. governor may be nil (used in tests).
func NewExperimentSelector(governor *resource.ResourceGovernor) *ExperimentSelector {
	return &ExperimentSelector{governor: governor}
}

// SelectNextExperiment selects the best experiment to run next given:
//   - hypotheses: the current active hypothesis set (with confidence and conclusion state)
//   - executedFingerprints: fingerprints of already-executed experiments (for deduplication)
//   - fctx: the base fingerprint context of the environment (Target, GitRevision, etc.)
//
// Resource Governor is consulted before selection. If the system is CONSTRAINED,
// HIGH-cost experiments are deferred — preventing surprises at execution time.
//
// Returns nil, err when no feasible experiments remain (budget exhausted or all
// hypotheses are already eliminated).
func (es *ExperimentSelector) SelectNextExperiment(
	hypotheses []Hypothesis,
	executedFingerprints map[string]bool,
	fctx FingerprintContext,
) (*ExperimentPlan, error) {
	// Consult the Resource Governor BEFORE candidate selection.
	// This prevents discovering resource constraints at execution time.
	constrained := false
	if es.governor != nil {
		decision := es.governor.Evaluate("experiment-pre-selection", false)
		constrained = decision.Tier == resource.TierConstrained
	}

	var candidates []ExperimentPlan

	for _, hyp := range hypotheses {
		// Skip hypotheses that have been definitively eliminated by contradicting evidence.
		// INCONCLUSIVE hypotheses are still active — they need more evidence.
		if hyp.Conclusion == ConclusionContradicted || hyp.Conclusion == ConclusionRefuted {
			continue
		}

		tmpl, ok := hypothesisExperimentTemplates[hyp.ID]
		if !ok {
			continue // No experiment template for this hypothesis
		}

		plan := tmpl
		plan.HypothesisIDs = []string{hyp.ID}
		
		// Fill capability and calculate fingerprint dynamically
		fctx.Capability = plan.Capability
		plan.Fingerprint = computeFingerprint(fctx)

		// Deduplication: skip experiments already executed in this investigation.
		if executedFingerprints[plan.Fingerprint] {
			continue
		}

		// Resource governance: defer HIGH-cost experiments when the system is constrained.
		// This directly implements the "respect the user's hardware" principle.
		if constrained && plan.CostLevel == CostHigh {
			continue
		}

		candidates = append(candidates, plan)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no feasible experiments remain for current hypotheses and resource constraints")
	}

	// Rank candidates by Discrimination descending.
	// The experiment that best distinguishes between active hypotheses is selected first.
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].Discrimination > candidates[i].Discrimination {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	best := candidates[0]
	return &best, nil
}

