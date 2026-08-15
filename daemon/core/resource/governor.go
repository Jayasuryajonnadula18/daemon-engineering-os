package resource

import (
	"fmt"
	"sync"
	"time"
)

type ResourceTier string

const (
	TierConstrained     ResourceTier = "CONSTRAINED"
	TierBalanced        ResourceTier = "BALANCED"
	TierHighPerformance ResourceTier = "HIGH_PERFORMANCE"
)

type ResourceDecision string

const (
	DecisionExecute ResourceDecision = "EXECUTE"
	DecisionDefer   ResourceDecision = "DEFER"
	DecisionPause   ResourceDecision = "PAUSE"
	DecisionResume  ResourceDecision = "RESUME"
)

type GovernorDecision struct {
	Task          string           `json:"task"`
	Decision      ResourceDecision `json:"decision"`
	Reason        string           `json:"reason"`
	Tier          ResourceTier     `json:"tier"`
	UserRequested bool             `json:"user_requested"`
	Metrics       HardwareMetrics  `json:"metrics"`
	Budget        ResourceConfig   `json:"budget"`
	Timestamp     time.Time        `json:"timestamp"`
}

type ResourceGovernor struct {
	mu                sync.RWMutex
	profiler          *Profiler
	config            ResourceConfig
	currentlyDeferred bool
	overrideMetrics   *HardwareMetrics // used in tests to simulate pressure
}

func NewResourceGovernor(p *Profiler, cfg ResourceConfig) *ResourceGovernor {
	if p == nil {
		p = NewProfiler()
	}
	return &ResourceGovernor{
		profiler: p,
		config:   cfg,
	}
}

// SetOverrideMetrics injects synthetic hardware metrics for testing.
// Pass nil to clear the override and resume live profiler readings.
func (rg *ResourceGovernor) SetOverrideMetrics(m *HardwareMetrics) {
	rg.mu.Lock()
	defer rg.mu.Unlock()
	rg.overrideMetrics = m
}

func (rg *ResourceGovernor) CalculateTier(m HardwareMetrics) ResourceTier {
	if m.CPUUtilization >= rg.config.CPUDeferPercent || m.AvailableMemoryMB < rg.config.MemoryDeferMB {
		return TierConstrained
	}
	if m.CPUCores >= 16 && m.TotalMemoryMB >= 32768 && m.GPUAvailable {
		return TierHighPerformance
	}
	return TierBalanced
}

func (rg *ResourceGovernor) Evaluate(taskName string, isUserRequested bool) GovernorDecision {
	rg.mu.Lock()
	defer rg.mu.Unlock()

	var metrics HardwareMetrics
	if rg.overrideMetrics != nil {
		metrics = *rg.overrideMetrics
	} else {
		metrics = rg.profiler.Refresh()
	}
	tier := rg.CalculateTier(metrics)

	// User-requested work is NEVER silently blocked
	if isUserRequested {
		reason := "User explicit action granted execution."
		if metrics.CPUUtilization >= rg.config.CPUDeferPercent {
			reason = fmt.Sprintf("User explicit action granted. WARNING: Host CPU utilization (%.0f%%) is high.", metrics.CPUUtilization*100)
		}
		return GovernorDecision{
			Task:          taskName,
			Decision:      DecisionExecute,
			Reason:        reason,
			Tier:          tier,
			UserRequested: true,
			Metrics:       metrics,
			Budget:        rg.config,
			Timestamp:     time.Now(),
		}
	}

	// Background work evaluation with Hysteresis
	if rg.currentlyDeferred {
		if metrics.CPUUtilization <= rg.config.CPUResumePercent && metrics.AvailableMemoryMB >= rg.config.MemoryDeferMB {
			rg.currentlyDeferred = false
			return GovernorDecision{
				Task:          taskName,
				Decision:      DecisionResume,
				Reason:        fmt.Sprintf("Host CPU utilization (%.0f%%) cooled down below resume threshold (%.0f%%). Background work resumed.", metrics.CPUUtilization*100, rg.config.CPUResumePercent*100),
				Tier:          tier,
				UserRequested: false,
				Metrics:       metrics,
				Budget:        rg.config,
				Timestamp:     time.Now(),
			}
		}
		return GovernorDecision{
			Task:          taskName,
			Decision:      DecisionDefer,
			Reason:        fmt.Sprintf("Host CPU utilization (%.0f%%) exceeds resume threshold (%.0f%%). Background work remains deferred.", metrics.CPUUtilization*100, rg.config.CPUResumePercent*100),
			Tier:          tier,
			UserRequested: false,
			Metrics:       metrics,
			Budget:        rg.config,
			Timestamp:     time.Now(),
		}
	}

	if metrics.CPUUtilization >= rg.config.CPUDeferPercent || metrics.AvailableMemoryMB < rg.config.MemoryDeferMB {
		rg.currentlyDeferred = true
		return GovernorDecision{
			Task:          taskName,
			Decision:      DecisionDefer,
			Reason:        fmt.Sprintf("Host CPU utilization (%.0f%%) exceeds defer threshold (%.0f%%). Background work deferred to protect hardware.", metrics.CPUUtilization*100, rg.config.CPUDeferPercent*100),
			Tier:          tier,
			UserRequested: false,
			Metrics:       metrics,
			Budget:        rg.config,
			Timestamp:     time.Now(),
		}
	}

	return GovernorDecision{
		Task:          taskName,
		Decision:      DecisionExecute,
		Reason:        "Host resource utilization within safe thresholds.",
		Tier:          tier,
		UserRequested: false,
		Metrics:       metrics,
		Budget:        rg.config,
		Timestamp:     time.Now(),
	}
}

func (rg *ResourceGovernor) SelectModelTier() string {
	rg.mu.RLock()
	defer rg.mu.RUnlock()

	metrics := rg.profiler.GetMetrics()
	tier := rg.CalculateTier(metrics)

	switch tier {
	case TierHighPerformance:
		return "large"
	case TierConstrained:
		return "lightweight"
	default:
		return "medium"
	}
}
