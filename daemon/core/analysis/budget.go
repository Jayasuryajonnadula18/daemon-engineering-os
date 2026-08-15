package analysis

import (
	"encoding/json"
	"os"
	"time"
)

type AnalysisBudget struct {
	MaxDuration      time.Duration // default 5 minutes
	MaxFiles         int           // default 10000
	MaxBytes         int64         // default 500MB
	MaxMemoryMB      int64         // default 1024MB
	MaxBackgroundCPU float64       // default 0.85
	MaxConcurrency   int           // default 4
}

type AnalysisStatus string

const (
	StatusCompleted AnalysisStatus = "COMPLETED"
	StatusPartial   AnalysisStatus = "PARTIAL"
	StatusDeferred  AnalysisStatus = "DEFERRED"
	StatusCancelled AnalysisStatus = "CANCELLED"
)

func DefaultAnalysisBudget() AnalysisBudget {
	return AnalysisBudget{
		MaxDuration:      5 * time.Minute,
		MaxFiles:         10000,
		MaxBytes:         500 * 1024 * 1024,
		MaxMemoryMB:      1024,
		MaxBackgroundCPU: 0.85,
		MaxConcurrency:   4,
	}
}

// LoadAnalysisBudget loads budget settings from a JSON file, or returns defaults if missing/unparseable.
func LoadAnalysisBudget(configPath string) AnalysisBudget {
	budget := DefaultAnalysisBudget()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return budget
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return budget
	}

	// Helper to extract nested or flat properties
	getVal := func(keys ...string) (any, bool) {
		// First try flat key like "analysis.max_duration"
		for _, key := range keys {
			if val, ok := raw[key]; ok {
				return val, true
			}
		}
		// Then try nested under "analysis" -> "max_duration"
		if analysisVal, ok := raw["analysis"]; ok {
			if nestedMap, ok := analysisVal.(map[string]any); ok {
				for _, key := range keys {
					// convert flat to short key (e.g. "analysis.max_duration" -> "max_duration")
					shortKey := key
					if idx := len("analysis."); idx < len(key) && key[:idx] == "analysis." {
						shortKey = key[idx:]
					}
					if val, ok := nestedMap[shortKey]; ok {
						return val, true
					}
				}
			}
		}
		return nil, false
	}

	// 1. Duration
	if val, ok := getVal("analysis.max_duration"); ok {
		if strVal, ok := val.(string); ok {
			if dur, err := time.ParseDuration(strVal); err == nil {
				budget.MaxDuration = dur
			}
		} else if numVal, ok := val.(float64); ok {
			budget.MaxDuration = time.Duration(numVal) * time.Second
		}
	}

	// 2. MaxFiles
	if val, ok := getVal("analysis.max_files"); ok {
		if numVal, ok := val.(float64); ok {
			budget.MaxFiles = int(numVal)
		}
	}

	// 3. MaxBytes
	if val, ok := getVal("analysis.max_bytes"); ok {
		if numVal, ok := val.(float64); ok {
			budget.MaxBytes = int64(numVal)
		}
	}

	// 4. MaxMemoryMB
	if val, ok := getVal("analysis.max_memory_mb"); ok {
		if numVal, ok := val.(float64); ok {
			budget.MaxMemoryMB = int64(numVal)
		}
	}

	// 5. MaxBackgroundCPU
	if val, ok := getVal("analysis.max_background_cpu"); ok {
		if numVal, ok := val.(float64); ok {
			budget.MaxBackgroundCPU = numVal
		}
	}

	// 6. MaxConcurrency
	if val, ok := getVal("analysis.max_concurrency"); ok {
		if numVal, ok := val.(float64); ok {
			budget.MaxConcurrency = int(numVal)
		}
	}

	return budget
}
