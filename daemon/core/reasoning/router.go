package reasoning

import (
	"strings"
)

// ModelRecommendation contains selected model characteristics.
type ModelRecommendation struct {
	ModelName    string  `json:"model_name"`
	LatencyMs    int     `json:"latency_ms"`
	CostPerToken float64 `json:"cost_per_token"`
	Provider     string  `json:"provider"`
	Offline      bool    `json:"offline"`
}

// ModelRouter selects the most appropriate provider-independent model dynamically.
type ModelRouter struct {
	offlineMode bool
}

// NewModelRouter instantiates a new ModelRouter.
func NewModelRouter(offline bool) *ModelRouter {
	return &ModelRouter{offlineMode: offline}
}

// RouteTask determines the best model based on task categories and context size.
func (mr *ModelRouter) RouteTask(taskType string, contextLength int) ModelRecommendation {
	if mr.offlineMode {
		return ModelRecommendation{
			ModelName:    "qwen3-coder-ollama",
			LatencyMs:    150,
			CostPerToken: 0.0,
			Provider:     "Ollama",
			Offline:      true,
		}
	}

	taskLower := strings.ToLower(taskType)

	switch {
	case strings.Contains(taskLower, "architecture") || strings.Contains(taskLower, "planning") || strings.Contains(taskLower, "incident"):
		return ModelRecommendation{
			ModelName:    "claude-3-5-sonnet",
			LatencyMs:    450,
			CostPerToken: 0.000015,
			Provider:     "Anthropic",
			Offline:      false,
		}
	case strings.Contains(taskLower, "doc") || strings.Contains(taskLower, "summarize"):
		return ModelRecommendation{
			ModelName:    "gemini-1-5-pro",
			LatencyMs:    280,
			CostPerToken: 0.000007,
			Provider:     "Google",
			Offline:      false,
		}
	default:
		// Fallback general reasoning model
		return ModelRecommendation{
			ModelName:    "gpt-4o-mini",
			LatencyMs:    190,
			CostPerToken: 0.000002,
			Provider:     "OpenAI",
			Offline:      false,
		}
	}
}
