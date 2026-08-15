package reasoning

import (
	"testing"
)

func TestModelRouter_TaskAwareSelection(t *testing.T) {
	mr := NewModelRouter(false)

	recCode := mr.RouteTask("code_reasoning", 5000)
	if recCode.Provider != "Anthropic" {
		t.Fatalf("expected Anthropic provider for code_reasoning, got %s", recCode.Provider)
	}
	if !recCode.Capability.CodeReasoning {
		t.Fatalf("expected CodeReasoning capability to be true")
	}

	recArch := mr.RouteTask("architecture_analysis", 150000)
	if recArch.Provider != "Gemini" {
		t.Fatalf("expected Gemini provider for long-context architecture analysis, got %s", recArch.Provider)
	}

	recPrivacy := mr.RouteTask("privacy_sensitive", 1000)
	if recPrivacy.Provider != "Ollama" || !recPrivacy.Offline {
		t.Fatalf("expected Ollama offline provider for privacy sensitive task, got %s", recPrivacy.Provider)
	}
}
