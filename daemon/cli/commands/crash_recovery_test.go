package commands

import (
	"bytes"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"daemon/core/security"
	_ "modernc.org/sqlite"
)

// TestCLI_TrueProcessCrashRecovery compiles, runs daemon.exe automate, kills it mid-execution,
// and verifies that resuming is handled safely depending on the node's idempotency.
func TestCLI_TrueProcessCrashRecovery(t *testing.T) {
	// 1. Find or build daemon.exe
	tmp := t.TempDir()
	binPath := filepath.Join(tmp, "daemon.exe")

	secret, err := security.GetOrGenerateMasterSecret()
	if err != nil || secret == "" {
		t.Fatalf("failed to retrieve or generate master secret for testing: %v", err)
	}

	// Compile the CLI binary
	cmdBuild := exec.Command("go", "build", "-o", binPath, "../../main.go")
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("failed to compile daemon binary: %v\nOutput: %s", err, string(out))
	}

	// 2. Set up a disposable project workspace
	projectDir := filepath.Join(tmp, "workspace")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	// Create a mock Go file in workspace so other features work if needed
	err = os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("package main\nfunc main() {}"), 0644)
	if err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	// 3. Test Case A: Interrupt during an IDEMPOTENT node ("node-1-validate")
	// Since node-1 is the first node and takes ~20ms, if we kill it almost immediately (say, after 5ms),
	// it will be in the NodeRunning state in the checkpoint DB.
	// Since it is idempotent, the resume command should succeed.
	
	// Create the process
	cmd := exec.Command(binPath, "automate", "--json")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "DAEMON_PASSWORD="+secret)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start execution
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon process: %v", err)
	}

	// Poll until the database file is created and has the node_checkpoints table, meaning it started writing
	dbPath := filepath.Join(projectDir, ".daemon", "daemon.db")
	var polledOk bool
	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		db, err := sql.Open("sqlite", dbPath)
		if err == nil {
			_, _ = db.Exec("PRAGMA busy_timeout=5000;")
			_, _ = db.Exec("PRAGMA journal_mode=WAL;")
			func() {
				defer db.Close()
				var count int
				err = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='node_checkpoints'").Scan(&count)
				if err == nil && count > 0 {
					var rowCount int
					err = db.QueryRow("SELECT count(*) FROM node_checkpoints").Scan(&rowCount)
					if err == nil && rowCount > 0 {
						polledOk = true
					}
				}
			}()
			if polledOk {
				break
			}
		}
	}

	if !polledOk {
		t.Fatalf("timed out waiting for daemon database to initialize and start writing checkpoints.\nStdout: %s\nStderr: %s", stdout.String(), stderr.String())
	}

	// Hard kill the process mid-execution
	if err := cmd.Process.Kill(); err != nil {
		t.Fatalf("failed to kill daemon process: %v", err)
	}

	// Wait for the process to exit completely to release OS file locks
	_ = cmd.Wait()
	time.Sleep(100 * time.Millisecond) // extra buffer for file lock release on Windows

	// Verify SQLite database remains valid and has the checkpoint
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	var executionID, nodeID, status string
	row := db.QueryRow("SELECT execution_id, node_id, status FROM node_checkpoints LIMIT 1")
	if err := row.Scan(&executionID, &nodeID, &status); err != nil {
		db.Close()
		t.Fatalf("failed to query checkpoint: %v (DB might be corrupted or empty)", err)
	}
	db.Close()

	t.Logf("Found checkpoint: ExecID=%s, NodeID=%s, Status=%s", executionID, nodeID, status)

	// Now resume the execution. Since node-1 is idempotent, resume should succeed.
	cmdResume := exec.Command(binPath, "automate", "--resume", executionID, "--json")
	cmdResume.Dir = projectDir
	cmdResume.Env = append(os.Environ(), "DAEMON_PASSWORD="+secret)
	outResume, err := cmdResume.CombinedOutput()
	if err != nil {
		t.Fatalf("resume failed: %v\nOutput: %s", err, string(outResume))
	}

	t.Logf("Resume Output:\n%s", string(outResume))
	if !strings.Contains(string(outResume), `"final_state": "COMPLETED"`) {
		t.Errorf("expected resume to succeed and complete the execution for idempotent node")
	}

	// 4. Test Case B: Simulate an incomplete NON-IDEMPOTENT node
	// We will manually pre-seed the checkpoint database with a non-idempotent node ("node-3-deploy")
	// in RUNNING state to simulate a crash during that node, and verify that resuming triggers MANUAL_INTERVENTION.
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open db for seeding: %v", err)
	}
	defer db.Close()

	// Seed node-1 and node-2 as verified
	_, _ = db.Exec(`INSERT OR REPLACE INTO node_checkpoints 
		(execution_id, dag_id, node_id, attempt, status, started_at) 
		VALUES (?, ?, ?, ?, ?, ?)`, 
		"exec-non-idempotent-crash", "dag-test", "node-1-validate", 1, "VERIFIED", time.Now().Format(time.RFC3339))
	
	_, _ = db.Exec(`INSERT OR REPLACE INTO node_checkpoints 
		(execution_id, dag_id, node_id, attempt, status, started_at) 
		VALUES (?, ?, ?, ?, ?, ?)`, 
		"exec-non-idempotent-crash", "dag-test", "node-2-build", 1, "VERIFIED", time.Now().Format(time.RFC3339))

	// Seed node-3 (non-idempotent) as RUNNING (incomplete/crashed)
	_, err = db.Exec(`INSERT OR REPLACE INTO node_checkpoints 
		(execution_id, dag_id, node_id, attempt, status, started_at) 
		VALUES (?, ?, ?, ?, ?, ?)`, 
		"exec-non-idempotent-crash", "dag-test", "node-3-deploy", 1, "RUNNING", time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("failed to seed non-idempotent checkpoint: %v", err)
	}

	// Execute resume on this non-idempotent execution
	cmdResumeNonIdem := exec.Command(binPath, "automate", "--resume", "exec-non-idempotent-crash", "--json")
	cmdResumeNonIdem.Dir = projectDir
	cmdResumeNonIdem.Env = append(os.Environ(), "DAEMON_PASSWORD="+secret)
	outResumeNonIdem, err := cmdResumeNonIdem.CombinedOutput()
	if err != nil {
		t.Fatalf("resume command errored: %v\nOutput: %s", err, string(outResumeNonIdem))
	}

	t.Logf("Non-Idempotent Resume Output:\n%s", string(outResumeNonIdem))
	if !strings.Contains(string(outResumeNonIdem), `"final_state": "MANUAL_INTERVENTION"`) {
		t.Errorf("expected resume to trigger MANUAL_INTERVENTION for non-idempotent incomplete node")
	}
}
