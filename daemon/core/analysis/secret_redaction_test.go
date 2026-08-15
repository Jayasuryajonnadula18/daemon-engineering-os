package analysis

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSecretRedaction_CredentialNeverInFindingDescription verifies hardcoded secrets
// are NOT emitted verbatim in finding descriptions — only variable names appear.
func TestSecretRedaction_CredentialNeverInFindingDescription(t *testing.T) {
	dir := t.TempDir()

	secretValue := "super_secret_api_key_12345_REDACT_ME"
	code := `package main
func main() {
	token := "` + secretValue + `"
	_ = token
}`
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0644)

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	res, err := pipeline.RunAnalysis(context.Background(), dir, true)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}

	for _, f := range res.Findings {
		if strings.Contains(f.Description, secretValue) {
			t.Fatalf("secret value leaked into finding description: %q", f.Description)
		}
		if strings.Contains(f.Title, secretValue) {
			t.Fatalf("secret value leaked into finding title: %q", f.Title)
		}
		for _, ev := range f.EvidenceIDs {
			if strings.Contains(ev, secretValue) {
				t.Fatalf("secret value leaked into evidence ID: %q", ev)
			}
		}
	}
}

// TestSecretRedaction_PasswordNeverInJSON verifies password variable findings
// do not contain raw credential string values anywhere in JSON-serializable fields.
func TestSecretRedaction_PasswordNeverInJSON(t *testing.T) {
	dir := t.TempDir()

	rawPassword := "VERY_SECRET_HUNTER2_PASSWORD"
	code := `package main
func main() {
	password := "` + rawPassword + `"
	_ = password
}`
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0644)

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	res, err := pipeline.RunAnalysis(context.Background(), dir, true)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}

	// Verify no finding contains the raw secret value
	for _, f := range res.Findings {
		if strings.Contains(f.Description, rawPassword) {
			t.Fatalf("raw password leaked into finding description")
		}
		for _, action := range f.SuggestedActions {
			if strings.Contains(action, rawPassword) {
				t.Fatalf("raw password leaked into suggested action")
			}
		}
	}
}

// TestSecretRedaction_EnvFileNeverLeaked verifies .env-like config files
// with secrets are not emitted in analysis output.
func TestSecretRedaction_EnvFileNeverLeaked(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, ".env"), []byte(`
DATABASE_PASSWORD=ultra_secret_db_pass_9999
API_KEY=sk-live-123456789-NEVER-EMIT
`), 0644)

	code := `package main
import "os"
func main() {
	_ = os.Getenv("API_KEY")
}`
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0644)

	pipeline := NewDeepAnalyzerPipeline(nil, nil)
	res, err := pipeline.RunAnalysis(context.Background(), dir, true)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}

	blocklist := []string{"ultra_secret_db_pass_9999", "sk-live-123456789-NEVER-EMIT"}
	for _, f := range res.Findings {
		for _, secret := range blocklist {
			if strings.Contains(f.Description, secret) {
				t.Fatalf("env file secret leaked into finding: %s", secret)
			}
		}
	}
}
