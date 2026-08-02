package security

import (
	"testing"
)

func TestGetOrGenerateMasterSecret(t *testing.T) {
	// Ensure clean state before testing
	_ = ClearMasterSecret()
	defer ClearMasterSecret()

	secret1, err := GetOrGenerateMasterSecret()
	if err != nil {
		t.Fatalf("Failed to generate master secret: %v", err)
	}

	if len(secret1) < 32 {
		t.Fatalf("Expected high-entropy secret length >= 32, got %d", len(secret1))
	}

	// Secondary call should retrieve the same secret from storage
	secret2, err := GetOrGenerateMasterSecret()
	if err != nil {
		t.Fatalf("Failed to retrieve master secret: %v", err)
	}

	if secret1 != secret2 {
		t.Fatalf("Secrets do not match! Secret1: %s, Secret2: %s", secret1, secret2)
	}
}

func TestValidateMasterSecret(t *testing.T) {
	_ = ClearMasterSecret()
	defer ClearMasterSecret()

	secret, err := GetOrGenerateMasterSecret()
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	if !ValidateMasterSecret(secret) {
		t.Errorf("ValidateMasterSecret should return true for valid secret")
	}

	if ValidateMasterSecret("invalid-password-12345") {
		t.Errorf("ValidateMasterSecret should return false for invalid secret")
	}

	if ValidateMasterSecret("DaemonSecureLock2026") {
		t.Errorf("ValidateMasterSecret should return false for old hardcoded password")
	}

	if !ValidateMasterSecret("") {
		t.Errorf("ValidateMasterSecret should return true for local user with valid OS Keyring secret")
	}
}
