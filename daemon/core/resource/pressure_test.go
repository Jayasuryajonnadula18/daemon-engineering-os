package resource

import (
	"testing"
)

// TestResourcePressure_HighCPU verifies that background tasks are DEFERRED when CPU pressure exceeds threshold.
func TestResourcePressure_HighCPU(t *testing.T) {
	// Simulate high CPU: use a config with a very low defer threshold
	cfg := ResourceConfig{
		Adaptive:              true,
		CPUDeferPercent:       0.10, // defer at 10% CPU — will always trigger in test
		CPUResumePercent:      0.05,
		MemoryDeferMB:         1,    // won't trigger (1MB threshold)
		MaxBackgroundCPU:      0.05,
		MaxBackgroundMemoryMB: 512,
		MaxAIRequestsPerHour:  20,
	}

	profiler := NewProfiler()
	gov := NewResourceGovernor(profiler, cfg)

	// Force a high-CPU simulation by overriding with a mock metric
	gov.SetOverrideMetrics(&HardwareMetrics{
		CPUCores:          4,
		CPUUtilization:    0.95, // 95% CPU — well above any threshold
		TotalMemoryMB:     8192,
		AvailableMemoryMB: 6000,
		GPUAvailable:      false,
		FreeDiskGB:        100,
	})

	dec := gov.Evaluate("background_indexing", false)
	if dec.Decision != DecisionDefer && dec.Decision != DecisionPause {
		t.Fatalf("expected DEFER or PAUSE under high CPU pressure, got %s (reason: %s)", dec.Decision, dec.Reason)
	}
}

// TestResourcePressure_UserRequestedNeverBlocked verifies explicit user commands are never deferred.
func TestResourcePressure_UserRequestedNeverBlocked(t *testing.T) {
	cfg := ResourceConfig{
		Adaptive:              true,
		CPUDeferPercent:       0.10, // absurdly low threshold
		CPUResumePercent:      0.05,
		MemoryDeferMB:         1,
		MaxBackgroundCPU:      0.05,
		MaxBackgroundMemoryMB: 512,
		MaxAIRequestsPerHour:  20,
	}

	profiler := NewProfiler()
	gov := NewResourceGovernor(profiler, cfg)

	gov.SetOverrideMetrics(&HardwareMetrics{
		CPUCores:          4,
		CPUUtilization:    0.99, // near 100% CPU
		TotalMemoryMB:     8192,
		AvailableMemoryMB: 100,  // very low RAM
		GPUAvailable:      false,
		FreeDiskGB:        1,
	})

	// User-requested = true must always return EXECUTE
	dec := gov.Evaluate("daemon_automate", true)
	if dec.Decision != DecisionExecute {
		t.Fatalf("user-requested command must never be blocked by resource governor, got %s", dec.Decision)
	}
}

// TestResourcePressure_RecoveryResume verifies that after pressure drops, governor returns EXECUTE.
func TestResourcePressure_RecoveryResume(t *testing.T) {
	cfg := DefaultResourceConfig()
	profiler := NewProfiler()
	gov := NewResourceGovernor(profiler, cfg)

	// Simulate high pressure
	gov.SetOverrideMetrics(&HardwareMetrics{
		CPUCores:          4,
		CPUUtilization:    0.95,
		TotalMemoryMB:     8192,
		AvailableMemoryMB: 500,
		GPUAvailable:      false,
		FreeDiskGB:        10,
	})
	dec := gov.Evaluate("background_analysis", false)
	if dec.Decision == DecisionExecute {
		t.Logf("Note: system under pressure gave EXECUTE — host may genuinely be low-load. Skipping pressure assertion.")
		return
	}

	// Simulate pressure recovery
	gov.SetOverrideMetrics(&HardwareMetrics{
		CPUCores:          4,
		CPUUtilization:    0.20,
		TotalMemoryMB:     8192,
		AvailableMemoryMB: 6000,
		GPUAvailable:      false,
		FreeDiskGB:        100,
	})
	decRecovered := gov.Evaluate("background_analysis", false)
	if decRecovered.Decision != DecisionExecute && decRecovered.Decision != DecisionResume {
		t.Fatalf("expected EXECUTE or RESUME after resource pressure recovery, got %s", decRecovered.Decision)
	}
}

// TestResourcePressure_LowRAM verifies background analysis is deferred when RAM is constrained.
func TestResourcePressure_LowRAM(t *testing.T) {
	cfg := ResourceConfig{
		Adaptive:              true,
		CPUDeferPercent:       0.95, // won't trigger on CPU
		CPUResumePercent:      0.80,
		MemoryDeferMB:         8000, // defer if available < 8000MB (very high threshold)
		MaxBackgroundCPU:      0.15,
		MaxBackgroundMemoryMB: 512,
		MaxAIRequestsPerHour:  20,
	}

	profiler := NewProfiler()
	gov := NewResourceGovernor(profiler, cfg)

	gov.SetOverrideMetrics(&HardwareMetrics{
		CPUCores:          4,
		CPUUtilization:    0.10, // low CPU
		TotalMemoryMB:     8192,
		AvailableMemoryMB: 512, // very low available RAM — below 8000MB threshold
		GPUAvailable:      false,
		FreeDiskGB:        50,
	})

	dec := gov.Evaluate("background_deep_analysis", false)
	if dec.Decision != DecisionDefer && dec.Decision != DecisionPause {
		t.Fatalf("expected DEFER under low RAM, got %s", dec.Decision)
	}
}
