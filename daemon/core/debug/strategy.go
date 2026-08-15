package debug

import "strings"

type InvestigationStrategy string

const (
	StrategyRegression   InvestigationStrategy = "REGRESSION_STRATEGY"
	StrategyMemory       InvestigationStrategy = "MEMORY_LEAK_STRATEGY"
	StrategyCrash        InvestigationStrategy = "CRASH_STRATEGY"
	StrategyPerformance  InvestigationStrategy = "PERFORMANCE_STRATEGY"
	StrategyTestFailure  InvestigationStrategy = "TEST_FAILURE_STRATEGY"
	StrategyBuildFailure InvestigationStrategy = "BUILD_FAILURE_STRATEGY"
	StrategyRuntime      InvestigationStrategy = "RUNTIME_FAILURE_STRATEGY"
	StrategyConfig       InvestigationStrategy = "CONFIGURATION_STRATEGY"
	StrategyGeneric      InvestigationStrategy = "GENERIC_STRATEGY"
)

type StrategyPlanner struct{}

func NewStrategyPlanner() *StrategyPlanner {
	return &StrategyPlanner{}
}

// PlanStrategy determines the strategy dynamically based on problem description and gathered evidence
func (sp *StrategyPlanner) PlanStrategy(problem string, evidence []Evidence) InvestigationStrategy {
	lower := strings.ToLower(problem)

	// Check intent classification patterns
	if strings.Contains(lower, "leak") || strings.Contains(lower, "memory") || strings.Contains(lower, "growth") {
		return StrategyMemory
	}
	if strings.Contains(lower, "crash") || strings.Contains(lower, "panic") || strings.Contains(lower, "exception") {
		return StrategyCrash
	}
	if strings.Contains(lower, "regression") || strings.Contains(lower, "broke") || strings.Contains(lower, "last change") || strings.Contains(lower, "failing after") {
		return StrategyRegression
	}
	if strings.Contains(lower, "build") || strings.Contains(lower, "compile") {
		return StrategyBuildFailure
	}
	if strings.Contains(lower, "test") || strings.Contains(lower, "fail") {
		return StrategyTestFailure
	}
	if strings.Contains(lower, "performance") || strings.Contains(lower, "slow") || strings.Contains(lower, "latency") {
		return StrategyPerformance
	}
	if strings.Contains(lower, "config") || strings.Contains(lower, "env") || strings.Contains(lower, "port") {
		return StrategyConfig
	}
	if strings.Contains(lower, "why") || strings.Contains(lower, "500") || strings.Contains(lower, "restart") {
		return StrategyRuntime
	}

	// Dynamic evidence-based classification
	for _, ev := range evidence {
		if ev.Type == EvidenceLog && strings.Contains(strings.ToLower(ev.Statement), "panic") {
			return StrategyCrash
		}
		if ev.Type == EvidenceBuild && strings.Contains(strings.ToLower(ev.Statement), "error") {
			return StrategyBuildFailure
		}
	}

	return StrategyGeneric
}
