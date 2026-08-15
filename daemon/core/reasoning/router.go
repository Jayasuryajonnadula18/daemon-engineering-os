package reasoning

import (
	"context"
	"strings"
)

type TaskType string

const (
	TaskClassification      TaskType = "classification"
	TaskCodeReasoning       TaskType = "code_reasoning"
	TaskArchitectureAnalysis TaskType = "architecture_analysis"
	TaskLongContext         TaskType = "long_context"
	TaskPrivacySensitive    TaskType = "privacy_sensitive"
	TaskDeterministicCheck  TaskType = "deterministic_check"
)

// ModelCapability specifies capabilities advertised by a specific LLM model.
type ModelCapability struct {
	Provider         string `json:"provider"`
	Model            string `json:"model"`
	ContextWindow    int    `json:"context_window"`
	StructuredOutput bool   `json:"structured_output"`
	CodeReasoning    bool   `json:"code_reasoning"`
	LongContext      bool   `json:"long_context"`
	LocalPrivate     bool   `json:"local_private"`
	CostTier         string `json:"cost_tier"`
	LatencyEstimate  string `json:"latency_estimate"`
}

type ModelRequest struct {
	Prompt       string
	SystemPrompt string
}

type ModelResponse struct {
	Text string
}

// ModelProvider defines the interface implemented by distinct LLM backends.
type ModelProvider interface {
	Name() string
	Available(ctx context.Context) bool
	Generate(ctx context.Context, request ModelRequest) (ModelResponse, error)
	// Kept for backward compatibility
	GenerateOld(ctx context.Context, prompt string, systemPrompt string) (string, error)
}

// Distinct Provider Structs
type OllamaProvider struct{}
func (p *OllamaProvider) Name() string { return "Ollama" }
func (p *OllamaProvider) Available(ctx context.Context) bool { return true }
func (p *OllamaProvider) Generate(ctx context.Context, request ModelRequest) (ModelResponse, error) {
	return ModelResponse{Text: "Local reasoning output"}, nil
}
func (p *OllamaProvider) GenerateOld(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	return "Local reasoning output", nil
}
func (p *OllamaProvider) Capability() ModelCapability {
	return ModelCapability{
		Provider: "Ollama", Model: "qwen2.5-coder:7b", ContextWindow: 32000,
		StructuredOutput: true, CodeReasoning: true, LocalPrivate: true, CostTier: "free", LatencyEstimate: "fast",
	}
}

type AnthropicProvider struct{}
func (p *AnthropicProvider) Name() string { return "Anthropic" }
func (p *AnthropicProvider) Available(ctx context.Context) bool { return true }
func (p *AnthropicProvider) Generate(ctx context.Context, request ModelRequest) (ModelResponse, error) {
	return ModelResponse{Text: "Anthropic reasoning output"}, nil
}
func (p *AnthropicProvider) GenerateOld(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	return "Anthropic reasoning output", nil
}
func (p *AnthropicProvider) Capability() ModelCapability {
	return ModelCapability{
		Provider: "Anthropic", Model: "claude-3-5-sonnet", ContextWindow: 200000,
		StructuredOutput: true, CodeReasoning: true, LongContext: true, CostTier: "high", LatencyEstimate: "medium",
	}
}

type OpenAIProvider struct{}
func (p *OpenAIProvider) Name() string { return "OpenAI" }
func (p *OpenAIProvider) Available(ctx context.Context) bool { return true }
func (p *OpenAIProvider) Generate(ctx context.Context, request ModelRequest) (ModelResponse, error) {
	return ModelResponse{Text: "OpenAI reasoning output"}, nil
}
func (p *OpenAIProvider) GenerateOld(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	return "OpenAI reasoning output", nil
}
func (p *OpenAIProvider) Capability() ModelCapability {
	return ModelCapability{
		Provider: "OpenAI", Model: "gpt-4o", ContextWindow: 128000,
		StructuredOutput: true, CodeReasoning: true, CostTier: "medium", LatencyEstimate: "fast",
	}
}

type GeminiProvider struct{}
func (p *GeminiProvider) Name() string { return "Gemini" }
func (p *GeminiProvider) Available(ctx context.Context) bool { return true }
func (p *GeminiProvider) Generate(ctx context.Context, request ModelRequest) (ModelResponse, error) {
	return ModelResponse{Text: "Gemini reasoning output"}, nil
}
func (p *GeminiProvider) GenerateOld(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	return "Gemini reasoning output", nil
}
func (p *GeminiProvider) Capability() ModelCapability {
	return ModelCapability{
		Provider: "Gemini", Model: "gemini-1-5-pro", ContextWindow: 1000000,
		StructuredOutput: true, LongContext: true, CostTier: "low", LatencyEstimate: "fast",
	}
}

type DeepSeekProvider struct{}
func (p *DeepSeekProvider) Name() string { return "DeepSeek" }
func (p *DeepSeekProvider) Available(ctx context.Context) bool { return true }
func (p *DeepSeekProvider) Generate(ctx context.Context, request ModelRequest) (ModelResponse, error) {
	return ModelResponse{Text: "DeepSeek reasoning output"}, nil
}
func (p *DeepSeekProvider) GenerateOld(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	return "DeepSeek reasoning output", nil
}
func (p *DeepSeekProvider) Capability() ModelCapability {
	return ModelCapability{
		Provider: "DeepSeek", Model: "deepseek-v3", ContextWindow: 64000,
		StructuredOutput: true, CodeReasoning: true, CostTier: "low", LatencyEstimate: "medium",
	}
}

// ModelRecommendation contains selected model characteristics.
type ModelRecommendation struct {
	ModelName    string          `json:"model_name"`
	LatencyMs    int             `json:"latency_ms"`
	CostPerToken float64         `json:"cost_per_token"`
	Provider     string          `json:"provider"`
	Offline      bool            `json:"offline"`
	Capability   ModelCapability `json:"capability"`
}

// ModelRouter selects the most appropriate provider-independent model dynamically.
type ModelRouter struct {
	offlineMode bool
	providers   map[string]ModelProvider
}

// NewModelRouter instantiates a new ModelRouter.
func NewModelRouter(offline bool) *ModelRouter {
	mr := &ModelRouter{
		offlineMode: offline,
		providers:   make(map[string]ModelProvider),
	}
	mr.RegisterProvider(&OllamaProvider{})
	mr.RegisterProvider(&AnthropicProvider{})
	mr.RegisterProvider(&OpenAIProvider{})
	mr.RegisterProvider(&GeminiProvider{})
	mr.RegisterProvider(&DeepSeekProvider{})
	return mr
}

func (mr *ModelRouter) RegisterProvider(p ModelProvider) {
	mr.providers[p.Name()] = p
}

// RouteTask determines the best model based on task categories and context size.
func (mr *ModelRouter) RouteTask(taskType string, contextLength int) ModelRecommendation {
	tType := TaskType(strings.ToLower(taskType))

	// Get first available provider checking Availability
	var selected ModelProvider
	if mr.offlineMode || tType == TaskPrivacySensitive {
		selected = mr.providers["Ollama"]
	} else {
		switch {
		case tType == TaskArchitectureAnalysis || contextLength > 100000:
			selected = mr.providers["Gemini"]
		case tType == TaskCodeReasoning:
			selected = mr.providers["Anthropic"]
		default:
			selected = mr.providers["OpenAI"]
		}
	}

	// Fallback mechanism: if selected is not available, check other providers
	ctx := context.Background()
	if selected == nil || !selected.Available(ctx) {
		for _, prov := range mr.providers {
			if prov.Available(ctx) {
				selected = prov
				break
			}
		}
	}

	// Default fallback to Ollama if all fail
	if selected == nil {
		selected = mr.providers["Ollama"]
	}

	var cap ModelCapability
	switch selected.Name() {
	case "Ollama":
		cap = mr.providers["Ollama"].(*OllamaProvider).Capability()
		return ModelRecommendation{ModelName: cap.Model, LatencyMs: 150, CostPerToken: 0.0, Provider: "Ollama", Offline: true, Capability: cap}
	case "Gemini":
		cap = mr.providers["Gemini"].(*GeminiProvider).Capability()
		return ModelRecommendation{ModelName: cap.Model, LatencyMs: 280, CostPerToken: 0.000007, Provider: "Gemini", Offline: false, Capability: cap}
	case "Anthropic":
		cap = mr.providers["Anthropic"].(*AnthropicProvider).Capability()
		return ModelRecommendation{ModelName: cap.Model, LatencyMs: 450, CostPerToken: 0.000015, Provider: "Anthropic", Offline: false, Capability: cap}
	case "DeepSeek":
		cap = mr.providers["DeepSeek"].(*DeepSeekProvider).Capability()
		return ModelRecommendation{ModelName: cap.Model, LatencyMs: 450, CostPerToken: 0.000002, Provider: "DeepSeek", Offline: false, Capability: cap}
	default:
		cap = mr.providers["OpenAI"].(*OpenAIProvider).Capability()
		return ModelRecommendation{ModelName: cap.Model, LatencyMs: 190, CostPerToken: 0.000002, Provider: "OpenAI", Offline: false, Capability: cap}
	}
}

// Generate routes and executes LLM generation using the selected provider
func (mr *ModelRouter) Generate(ctx context.Context, taskType string, prompt string, systemPrompt string) (string, error) {
	rec := mr.RouteTask(taskType, len(prompt))
	provider := mr.providers[rec.Provider]
	resp, err := provider.Generate(ctx, ModelRequest{Prompt: prompt, SystemPrompt: systemPrompt})
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}
