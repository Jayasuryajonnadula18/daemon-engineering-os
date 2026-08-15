package reasoning

import (
	"sync"
	"time"
)

// ModelMetric logs observable telemetry for model reasoning calls.
type ModelMetric struct {
	ID           string        `json:"id"`
	Provider     string        `json:"provider"`
	Model        string        `json:"model"`
	Task         string        `json:"task"`
	LatencyMs    int64         `json:"latency_ms"`
	InputTokens  int           `json:"input_tokens"`
	OutputTokens int           `json:"output_tokens"`
	FallbackUsed bool          `json:"fallback_used"`
	ContextSize  int           `json:"context_size"`
	Success      bool          `json:"success"`
	Timestamp    time.Time     `json:"timestamp"`
}

type ModelTelemetry struct {
	mu      sync.RWMutex
	metrics []ModelMetric
}

func NewModelTelemetry() *ModelTelemetry {
	return &ModelTelemetry{metrics: make([]ModelMetric, 0)}
}

func (mt *ModelTelemetry) Record(metric ModelMetric) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	metric.Timestamp = time.Now()
	mt.metrics = append(mt.metrics, metric)
}

func (mt *ModelTelemetry) GetMetrics() []ModelMetric {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	res := make([]ModelMetric, len(mt.metrics))
	copy(res, mt.metrics)
	return res
}
