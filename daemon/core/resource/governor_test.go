package resource

import (
	"context"
	"testing"
)

func TestResourceGovernor_EvaluationAndHysteresis(t *testing.T) {
	profiler := NewProfiler()
	cfg := DefaultResourceConfig()
	gov := NewResourceGovernor(profiler, cfg)

	// 1. Normal Load (25% CPU) -> EXECUTE
	dec1 := gov.Evaluate("background_index", false)
	if dec1.Decision != DecisionExecute {
		t.Fatalf("expected EXECUTE under 25%% CPU, got %s", dec1.Decision)
	}

	// 2. High Load (90% CPU) -> DEFER
	profiler.SetSimulatedLoad(0.90, 8192)
	dec2 := gov.Evaluate("background_index", false)
	if dec2.Decision != DecisionDefer {
		t.Fatalf("expected DEFER under 90%% CPU, got %s", dec2.Decision)
	}

	// 3. Fluctuation (75% CPU) -> Still DEFER due to hysteresis (resume threshold is 70%)
	profiler.SetSimulatedLoad(0.75, 8192)
	dec3 := gov.Evaluate("background_index", false)
	if dec3.Decision != DecisionDefer {
		t.Fatalf("expected DEFER during hysteresis at 75%% CPU, got %s", dec3.Decision)
	}

	// 4. Cool Down (65% CPU) -> RESUME
	profiler.SetSimulatedLoad(0.65, 8192)
	dec4 := gov.Evaluate("background_index", false)
	if dec4.Decision != DecisionResume {
		t.Fatalf("expected RESUME at 65%% CPU, got %s", dec4.Decision)
	}

	// 5. User-Requested Work Override
	profiler.SetSimulatedLoad(0.95, 8192)
	decUser := gov.Evaluate("user_deploy", true)
	if decUser.Decision != DecisionExecute {
		t.Fatalf("expected EXECUTE for user-requested work even under 95%% CPU, got %s", decUser.Decision)
	}
}

func TestAdaptiveScheduler_PriorityGating(t *testing.T) {
	profiler := NewProfiler()
	cfg := DefaultResourceConfig()
	gov := NewResourceGovernor(profiler, cfg)
	sched := NewAdaptiveScheduler(gov)

	executed := false
	sched.RegisterTask(ScheduledTask{
		ID:       "task-bg",
		Name:     "background_indexing",
		Priority: PriorityBackground,
		Fn: func(ctx context.Context) error {
			executed = true
			return nil
		},
	})

	// Simulate High CPU
	profiler.SetSimulatedLoad(0.92, 8192)
	dec, err := sched.ExecuteTask(context.Background(), "task-bg", false)
	if err == nil {
		t.Fatalf("expected error on deferred background task")
	}
	if dec.Decision != DecisionDefer {
		t.Fatalf("expected decision DEFER, got %s", dec.Decision)
	}
	if executed {
		t.Fatalf("background task should not have executed under 92%% CPU load")
	}
}
