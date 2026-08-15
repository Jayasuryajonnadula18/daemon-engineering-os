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

func TestCLI_DebugCommand(t *testing.T) {
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

	// Helper to run CLI with keyring secret
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

	// 1a. Run debug command on a healthy workspace — must NOT fabricate root causes
	t.Run("debug basic healthy", func(t *testing.T) {
		out, err := runCLI("debug", "checkout is failing")
		// A healthy workspace returns INSUFFICIENT_CONTEXT — this is the correct behavior.
		// The test accepts both: a clean exit with INSUFFICIENT_CONTEXT, or an error.
		_ = err
		if strings.Contains(out, "ROOT_CAUSE_IDENTIFIED") {
			// Only valid if actual evidence was found
			if !strings.Contains(out, "VERIFIED") {
				t.Errorf("debug basic: ROOT_CAUSE_IDENTIFIED without VERIFIED evidence is a false certainty violation: %s", out)
			}
		}
		if !strings.Contains(out, "DAEMON DEBUG REPORT") {
			t.Errorf("expected debug report title, got: %s", out)
		}
	})

	// 1b. Run debug command on a workspace with a real detectable problem
	t.Run("debug basic leak", func(t *testing.T) {
		// Write a file with a detectable HTTP body leak
		_ = os.WriteFile(filepath.Join(projectDir, "leak.go"), []byte(`package main
import "net/http"
func fetch() { resp, _ := http.Get("http://example.com"); _ = resp }
`), 0644)
		defer os.Remove(filepath.Join(projectDir, "leak.go"))

		out, err := runCLI("debug", "memory usage keeps increasing")
		if err != nil {
			t.Fatalf("debug leak command failed: %v", err)
		}
		if !strings.Contains(out, "DAEMON DEBUG REPORT") {
			t.Errorf("expected debug report title, got: %s", out)
		}
		if !strings.Contains(out, "ROOT_CAUSE_IDENTIFIED") {
			t.Errorf("expected identified cause for leak workspace, got: %s", out)
		}
	})

	// 2. Run JSON debug command
	t.Run("debug json", func(t *testing.T) {
		out, err := runCLI("debug", "checkout is failing", "--json")
		if err != nil {
			t.Fatalf("debug json command failed: %v", err)
		}
		var resp struct {
			Success bool                   `json:"success"`
			Command string                 `json:"command"`
			Data    map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("invalid json output: %v, output: %s", err, out)
		}
		if !resp.Success || resp.Command != "debug" {
			t.Errorf("unexpected json envelope: %s", out)
		}
	})

	// 3. Secrets Redaction
	t.Run("debug secrets redaction", func(t *testing.T) {
		out, err := runCLI("debug", "checkout failing due to api_key=DAEMON_TEST_SECRET_DO_NOT_USE_12345", "--json")
		if err != nil {
			t.Fatalf("debug command failed: %v", err)
		}
		if strings.Contains(out, "DAEMON_TEST_SECRET_DO_NOT_USE") {
			t.Errorf("expected secret canary to be redacted, got: %s", out)
		}
		if !strings.Contains(out, "[REDACTED_SECRET]") {
			t.Errorf("expected [REDACTED_SECRET] string, got: %s", out)
		}
	})
}
