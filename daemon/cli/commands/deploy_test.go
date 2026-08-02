package commands

import (
	"testing"
)

func TestIsProductionEnvironment(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{"production", true},
		{"prod", true},
		{"staging", true},
		{"live", true},
		{"PROD-US-EAST", true},
		{"development", false},
		{"dev", false},
		{"local", false},
		{"test", false},
	}

	for _, tt := range tests {
		result := IsProductionEnvironment(tt.env)
		if result != tt.expected {
			t.Errorf("IsProductionEnvironment(%q) = %v; want %v", tt.env, result, tt.expected)
		}
	}
}
