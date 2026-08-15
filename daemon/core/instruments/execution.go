package instruments

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"daemon/core/policies"
	"daemon/core/resource"
)

type ExecutionValidator struct{}

func NewExecutionValidator() *ExecutionValidator {
	return &ExecutionValidator{}
}

func (ev *ExecutionValidator) ValidateRequest(req ToolRequest, workspaceDir string) error {
	if req.Executable == "" {
		return fmt.Errorf("executable is required")
	}

	// 1. Path Traversal & Workspace Boundary check
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err == nil && req.Dir != "" {
		absDir, err := filepath.Abs(req.Dir)
		if err == nil {
			if !strings.HasPrefix(absDir, absWorkspace) {
				return fmt.Errorf("path boundary violation: execution directory %s outside workspace %s", absDir, absWorkspace)
			}
		}
	}

	// 2. Shell Metacharacter validation (reject injection metacharacters)
	metacharacters := []string{";", "&&", "||", "|", "`", "$", "(", ")", ">", "<", "\n", "\r"}
	for _, arg := range req.Args {
		for _, char := range metacharacters {
			if strings.Contains(arg, char) {
				return fmt.Errorf("security check failure: argument contains shell metacharacter '%s'", char)
			}
		}
	}

	return nil
}

type InstrumentExecutor struct {
	policy    policies.PolicyEngine
	governor  *resource.ResourceGovernor
	validator *ExecutionValidator
}

func NewInstrumentExecutor(p policies.PolicyEngine, g *resource.ResourceGovernor) *InstrumentExecutor {
	return &InstrumentExecutor{
		policy:    p,
		governor:  g,
		validator: NewExecutionValidator(),
	}
}

// ExecuteRequest validates security policies, resource allocation, and runs the safe subprocess.
func (ie *InstrumentExecutor) ExecuteRequest(ctx context.Context, cap Capability, req ToolRequest, workspaceDir string, target string, userRequested bool) (ToolResult, error) {
	// 1. Static Validation
	if err := ie.validator.ValidateRequest(req, workspaceDir); err != nil {
		return ToolResult{Success: false, Stderr: err.Error()}, err
	}

	// 2. Safety Policy evaluation
	if ie.policy != nil {
		risk := "low"
		if !req.ReadOnly {
			risk = "high"
		}
		decision, err := ie.policy.Evaluate(ctx, string(cap), target)
		if err != nil {
			return ToolResult{}, err
		}
		// Also evaluate under Capability safety policies
		if memPolicy, ok := ie.policy.(*policies.MemoryPolicyEngine); ok {
			decision, err = memPolicy.EvaluateCapability(ctx, string(cap), target, risk)
			if err != nil {
				return ToolResult{}, err
			}
		}

		if decision == policies.DecDeny {
			return ToolResult{
				Success: false,
				Stderr:  "Policy Engine Denied Capability Execution",
			}, fmt.Errorf("unauthorized tool execution by Policy Engine")
		}
	}

	// 3. Resource Governor evaluation
	if ie.governor != nil {
		decision := ie.governor.Evaluate(req.Executable, userRequested)
		if decision.Decision == resource.DecisionDefer {
			return ToolResult{
				Success: false,
				Stderr:  "Resource execution deferred: " + decision.Reason,
				Metadata: map[string]string{
					"governor_status": "deferred",
					"reason":          decision.Reason,
				},
			}, nil
		}
	}

	// 4. Executing the safe command
	startTime := time.Now()
	cmd := exec.CommandContext(ctx, req.Executable, req.Args...)
	if req.Dir != "" {
		cmd.Dir = req.Dir
	} else {
		cmd.Dir = workspaceDir
	}

	if len(req.Env) > 0 {
		cmd.Env = req.Env
	}

	out, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	stdoutClean, _ := RedactSecrets(string(out))

	return ToolResult{
		InstrumentID: req.Executable,
		ExitCode:     exitCode,
		Duration:     duration,
		Success:      err == nil,
		Stdout:       stdoutClean,
		Metadata:     req.Metadata,
	}, nil
}
