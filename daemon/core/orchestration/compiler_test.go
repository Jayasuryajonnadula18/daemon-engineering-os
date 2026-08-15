package orchestration

import (
	"testing"
)

func TestDAGCompiler_CompileAndValidateAcyclic(t *testing.T) {
	compiler := NewDAGCompiler()
	intent := ExecutionIntent{
		Objective: "restart orders service",
		Targets:   []string{"service-orders"},
	}

	dag, err := compiler.Compile(intent)
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	if dag.TotalNodes != 3 {
		t.Fatalf("expected 3 nodes, got %d", dag.TotalNodes)
	}
	if dag.State != StateCompiled {
		t.Fatalf("expected StateCompiled, got %s", dag.State)
	}

	if !compiler.ValidateFreshness(dag, "v1.0.0") {
		t.Fatalf("expected freshness validation to pass")
	}
	if compiler.ValidateFreshness(dag, "v2.0.0-mutated") {
		t.Fatalf("expected freshness validation to fail on mutated twin version")
	}
}

func TestWaveScheduler_LockMatrix(t *testing.T) {
	sched := NewWaveScheduler()

	readLock := []ResourceLock{{ResourceID: "db", Mode: "READ"}}
	writeLock := []ResourceLock{{ResourceID: "db", Mode: "WRITE"}}

	if !sched.AcquireLocks(readLock) {
		t.Fatalf("failed to acquire initial read lock")
	}

	// Another READ lock should be allowed
	if !sched.CanAcquireLocks(readLock) {
		t.Fatalf("expected READ+READ to be allowed in lock matrix")
	}

	// WRITE lock should conflict with active READ lock
	if sched.CanAcquireLocks(writeLock) {
		t.Fatalf("expected WRITE lock to conflict with active READ lock")
	}

	sched.ReleaseLocks(readLock)

	// Now WRITE lock can be acquired
	if !sched.AcquireLocks(writeLock) {
		t.Fatalf("failed to acquire WRITE lock after release")
	}
}
