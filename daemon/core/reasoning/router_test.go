package reasoning

import (
	"testing"
)

func TestModelRouter(t *testing.T) {
	mrOnline := NewModelRouter(false)
	mrOffline := NewModelRouter(true)

	recOffline := mrOffline.RouteTask("planning", 100)
	if recOffline.ModelName != "qwen3-coder-ollama" || !recOffline.Offline {
		t.Errorf("expected offline Ollama model, got %s", recOffline.ModelName)
	}

	recOnline := mrOnline.RouteTask("planning", 2000)
	if recOnline.ModelName != "claude-3-5-sonnet" || recOnline.Provider != "Anthropic" {
		t.Errorf("expected Sonnet model, got %s", recOnline.ModelName)
	}

	recDoc := mrOnline.RouteTask("documentation", 500)
	if recDoc.ModelName != "gemini-1-5-pro" {
		t.Errorf("expected Gemini model, got %s", recDoc.ModelName)
	}
}
