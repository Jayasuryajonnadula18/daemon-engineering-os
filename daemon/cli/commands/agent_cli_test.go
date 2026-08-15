package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"daemon/core/security"
)

func TestCLI_AgentAndToolsCommands(t *testing.T) {
	tmp := t.TempDir()
	binPath := filepath.Join(tmp, "daemon.exe")

	secret, err := security.GetOrGenerateMasterSecret()
	if err != nil || secret == "" {
		t.Fatalf("failed to retrieve or generate master secret: %v", err)
	}

	// Compile daemon.exe
	cmdBuild := exec.Command("go", "build", "-o", binPath, "../../main.go")
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("failed to compile daemon binary: %v\nOutput: %s", err, string(out))
	}

	projectDir := filepath.Join(tmp, "workspace")
	_ = os.MkdirAll(projectDir, 0755)
	_ = os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("package main\nfunc main() {}"), 0644)
	_ = os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte("module workspace\ngo 1.20"), 0644)

	// Helper helper to run CLI with correct env
	runCLI := func(args ...string) (string, error) {
		cmd := exec.Command(binPath, args...)
		cmd.Dir = projectDir
		cmd.Env = append(os.Environ(), "DAEMON_PASSWORD="+secret)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			return stdout.String(), fmt.Errorf("error: %v, stderr: %s, stdout: %s", err, stderr.String(), stdout.String())
		}
		return stdout.String(), nil
	}

	// 1. daemon tools list --json
	t.Run("tools list json", func(t *testing.T) {
		out, err := runCLI("tools", "list", "--json")
		if err != nil {
			t.Fatalf("tools list failed: %v", err)
		}
		var resp struct {
			Success bool `json:"success"`
			Data    []map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("invalid json output: %v, output: %s", err, out)
		}
		if !resp.Success || len(resp.Data) == 0 {
			t.Errorf("expected successful non-empty tools list, got: %s", out)
		}
	})

	// 2. daemon tools inspect
	t.Run("tools inspect json", func(t *testing.T) {
		out, err := runCLI("tools", "inspect", "read_file", "--json")
		if err != nil {
			t.Fatalf("tools inspect failed: %v", err)
		}
		if !strings.Contains(out, `"name": "read_file"`) {
			t.Errorf("expected tool metadata for read_file, got: %s", out)
		}
	})

	// 3. daemon agent run dry-run
	t.Run("agent run dry-run", func(t *testing.T) {
		out, err := runCLI("agent", "run", "inspect workspace", "--dry-run")
		if err != nil {
			t.Fatalf("agent run dry-run failed: %v", err)
		}
		if !strings.Contains(out, "Dry-run requested. Session") {
			t.Errorf("expected dry-run response, got: %s", out)
		}
	})

	// 4. daemon agent run
	var sessionID string
	t.Run("agent run live", func(t *testing.T) {
		out, err := runCLI("agent", "run", "inspect workspace", "--json")
		if err != nil {
			t.Fatalf("agent run live failed: %v", err)
		}
		var resp struct {
			Success bool `json:"success"`
			Data    map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("invalid json: %v, output: %s", err, out)
		}
		if !resp.Success {
			t.Fatalf("agent run reported failure: %s", out)
		}
		sessionID = resp.Data["id"].(string)
		if sessionID == "" {
			t.Fatalf("expected non-empty session ID, got: %s", out)
		}
	})

	// 5. daemon agent list
	t.Run("agent list", func(t *testing.T) {
		out, err := runCLI("agent", "list", "--json")
		if err != nil {
			t.Fatalf("agent list failed: %v", err)
		}
		if !strings.Contains(out, sessionID) {
			t.Errorf("expected list to contain session %s, got: %s", sessionID, out)
		}
	})

	// 6. daemon agent inspect
	t.Run("agent inspect", func(t *testing.T) {
		out, err := runCLI("agent", "inspect", sessionID, "--json")
		if err != nil {
			t.Fatalf("agent inspect failed: %v", err)
		}
		if !strings.Contains(out, sessionID) {
			t.Errorf("expected inspect response to contain session ID, got: %s", out)
		}
	})

	// 7. daemon agent cancel
	t.Run("agent cancel", func(t *testing.T) {
		out, err := runCLI("agent", "cancel", sessionID, "--json")
		if err != nil {
			t.Fatalf("agent cancel failed: %v", err)
		}
		if !strings.Contains(out, "StateCancelled") && !strings.Contains(out, "CANCELLED") {
			t.Errorf("expected cancel confirmation state, got: %s", out)
		}
	})

	// 8. daemon session list
	t.Run("session list", func(t *testing.T) {
		out, err := runCLI("session", "list", "--json")
		if err != nil {
			t.Fatalf("session list failed: %v", err)
		}
		if !strings.Contains(out, sessionID) {
			t.Errorf("expected session list to contain session %s, got: %s", sessionID, out)
		}
	})

	// 9. daemon session inspect
	t.Run("session inspect", func(t *testing.T) {
		out, err := runCLI("session", "inspect", sessionID, "--json")
		if err != nil {
			t.Fatalf("session inspect failed: %v", err)
		}
		if !strings.Contains(out, sessionID) {
			t.Errorf("expected session inspect to contain session ID, got: %s", out)
		}
	})
}
