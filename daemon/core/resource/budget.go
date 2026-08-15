package resource

type ResourceConfig struct {
	Adaptive              bool    `json:"adaptive"`
	CPUDeferPercent       float64 `json:"cpu_defer_percent"`
	CPUResumePercent      float64 `json:"cpu_resume_percent"`
	MemoryDeferMB         uint64  `json:"memory_defer_mb"`
	MaxBackgroundCPU      float64 `json:"max_background_cpu"`
	MaxBackgroundMemoryMB uint64  `json:"max_background_memory_mb"`
	MaxAIRequestsPerHour  int     `json:"max_ai_requests_per_hour"`
}

func DefaultResourceConfig() ResourceConfig {
	return ResourceConfig{
		Adaptive:              true,
		CPUDeferPercent:       0.85, // 85%
		CPUResumePercent:      0.70, // 70% hysteresis cooldown
		MemoryDeferMB:         1024, // 1GB
		MaxBackgroundCPU:      0.15, // 15%
		MaxBackgroundMemoryMB: 1024,
		MaxAIRequestsPerHour:  20,
	}
}
