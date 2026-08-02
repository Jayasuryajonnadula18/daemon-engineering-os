package security

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	ServiceName = "DaemonEngineeringOS"
	SecretUser  = "master_secret"
)

// GenerateRandomSecret generates a 32-byte (64 hex char) cryptographically secure random token.
func GenerateRandomSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetOrGenerateMasterSecret retrieves the master secret from OS keyring (or fallback file).
// If no secret exists, it generates a new secure random secret and persists it.
func GetOrGenerateMasterSecret() (string, error) {
	// 1. Try OS Keyring (Windows Credential Manager / macOS Keychain / Linux Secret Service)
	secret, err := keyring.Get(ServiceName, SecretUser)
	if err == nil && secret != "" {
		return strings.TrimSpace(secret), nil
	}

	// 2. Try file fallback in ~/.daemon/.master_secret
	home, homeErr := os.UserHomeDir()
	var fallbackPath string
	if homeErr == nil {
		fallbackPath = filepath.Join(home, ".daemon", ".master_secret")
		if data, err := os.ReadFile(fallbackPath); err == nil {
			trimmed := strings.TrimSpace(string(data))
			if trimmed != "" {
				// Attempt to migrate to OS Keyring
				_ = keyring.Set(ServiceName, SecretUser, trimmed)
				return trimmed, nil
			}
		}
	}

	// 3. No secret exists — generate a brand new random secret
	newSecret, err := GenerateRandomSecret()
	if err != nil {
		return "", err
	}

	// Try saving to OS Keyring first
	keyringErr := keyring.Set(ServiceName, SecretUser, newSecret)
	if keyringErr != nil && fallbackPath != "" {
		// Fallback to local file with restricted permissions (0600)
		_ = os.MkdirAll(filepath.Dir(fallbackPath), 0700)
		_ = os.WriteFile(fallbackPath, []byte(newSecret), 0600)
	}

	return newSecret, nil
}

// ValidateMasterSecret checks if the provided secret matches the registered master secret.
// If DAEMON_PASSWORD is set, it checks against the keyring/registered secret.
func ValidateMasterSecret(provided string) bool {
	if provided == "" {
		return false
	}
	expected, err := GetOrGenerateMasterSecret()
	if err != nil || expected == "" {
		return false
	}
	return strings.TrimSpace(provided) == expected
}

// ClearMasterSecret removes the master secret from OS keyring and fallback file (used in tests).
func ClearMasterSecret() error {
	_ = keyring.Delete(ServiceName, SecretUser)
	home, err := os.UserHomeDir()
	if err == nil {
		fallbackPath := filepath.Join(home, ".daemon", ".master_secret")
		_ = os.Remove(fallbackPath)
	}
	return nil
}
