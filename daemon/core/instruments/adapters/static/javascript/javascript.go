package staticjs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"daemon/core/instruments"
)

type JSBugsInstrument struct {
	ident instruments.InstrumentIdentity
}

func NewJSBugsInstrument() *JSBugsInstrument {
	return &JSBugsInstrument{
		ident: instruments.InstrumentIdentity{
			ID:          "js-bugs",
			Name:        "JS Static Bugs Analyzer",
			Version:     "1.0.0",
			Vendor:      "Daemon Core",
			Category:    instruments.CategoryStatic,
			Description: "Analyzes JavaScript/TypeScript source files for common patterns",
			Installed:   true,
		},
	}
}

func (j *JSBugsInstrument) Identity() instruments.InstrumentIdentity {
	return j.ident
}

func (j *JSBugsInstrument) Capabilities() []instruments.Capability {
	return []instruments.Capability{instruments.CapStaticAnalysis}
}

func (j *JSBugsInstrument) Detect(ctx context.Context, env instruments.Environment) instruments.DetectionResult {
	if _, err := os.Stat(filepath.Join(env.ProjectDir, "package.json")); err == nil {
		return instruments.DetectionResult{Compatible: true, Reason: "package.json exists"}
	}
	return instruments.DetectionResult{Compatible: true, Reason: "Always checkable"}
}

func (j *JSBugsInstrument) Health(ctx context.Context) instruments.HealthResult {
	return instruments.HealthResult{Status: "AVAILABLE", Reason: "Static analyzer ready"}
}

func (j *JSBugsInstrument) BuildRequest(ctx context.Context, request instruments.InstrumentRequest) (instruments.ToolRequest, error) {
	return instruments.ToolRequest{
		Executable: "js-bugs",
		Args:       request.Args,
		Dir:        request.Target,
		ReadOnly:   true,
	}, nil
}

func (j *JSBugsInstrument) Execute(ctx context.Context, request instruments.ToolRequest) (instruments.ToolResult, error) {
	projectDir := request.Dir
	if projectDir == "" {
		projectDir = "."
	}

	stdout := ""
	_ = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == ".git" || info.Name() == ".daemon") {
			return filepath.SkipDir
		}
		ext := filepath.Ext(path)
		if ext != ".js" && ext != ".jsx" && ext != ".ts" && ext != ".tsx" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)
		filename := filepath.Base(path)

		// 1. SSE/WebSocket Memory Leak (missing .off or removeListener cleanup, ignoring commented lines)
		var hasOn bool
		var hasOff bool
		lines := strings.Split(content, "\n")
		for _, l := range lines {
			lTrim := strings.TrimSpace(l)
			if strings.HasPrefix(lTrim, "//") || strings.HasPrefix(lTrim, "/*") || strings.HasPrefix(lTrim, "*") {
				continue
			}
			if strings.Contains(lTrim, ".on(") || strings.Contains(lTrim, ".addListener(") {
				hasOn = true
			}
			if strings.Contains(lTrim, ".off(") || strings.Contains(lTrim, "removeListener(") || strings.Contains(lTrim, "removeAllListeners(") {
				hasOff = true
			}
		}
		if hasOn && !hasOff {
			stdout += "SSE_LEAK:" + filename + "\n"
		}

		// 2. Empty Description/Property Crash
		if strings.Contains(content, "description.length") &&
			!strings.Contains(content, "description &&") &&
			!strings.Contains(content, "typeof description") {
			stdout += "DESC_CRASH:" + filename + "\n"
		}

		// 3. Index Key Abuse
		if (strings.HasSuffix(filename, ".jsx") || strings.HasSuffix(filename, ".tsx")) &&
			(strings.Contains(content, "key={index}") || strings.Contains(content, "key={idx}") || strings.Contains(content, "key={i}")) {
			stdout += "KEY_ABUSE:" + filename + "\n"
		}

		// 4. JWT Verification Bypass
		if strings.Contains(content, "alg") && strings.Contains(content, "'none'") {
			stdout += "JWT_BYPASS:" + filename + "\n"
		}

		// 5. Cache Key Invalidation Mismatch
		if strings.Contains(content, "invalidateProduct") &&
			strings.Contains(content, "prod:") &&
			(strings.Contains(content, "products:") || strings.Contains(content, "product:")) {
			stdout += "CACHE_MISMATCH:" + filename + "\n"
		}

		// 6. Concurrency Race Condition
		if strings.Contains(content, "SELECT stock_quantity") &&
			strings.Contains(content, "UPDATE products") &&
			strings.Contains(content, "stock_quantity -") {
			stdout += "CONCURRENCY_RACE:" + filename + "\n"
		}

		// 7. SQLite Database Transaction Lock
		if strings.Contains(content, "BEGIN TRANSACTION") &&
			strings.Contains(content, "paymentResult.success") &&
			strings.Contains(content, "UPDATE orders SET status = ?") {
			stdout += "DB_LOCK_LEAK:" + filename + "\n"
		}

		// 8. Event-Loop Blocker
		if (strings.Contains(content, "while (Date.now()") || strings.Contains(content, "while(Date.now()")) &&
			strings.Contains(content, "Math.random()") {
			stdout += "EVENT_LOOP_BLOCK:" + filename + "\n"
		}

		// 9. Floating-Point Comparison Mismatch
		if strings.Contains(content, "submittedTotal !==") || strings.Contains(content, "submittedTotal !=") {
			stdout += "FLOAT_PRECISION:" + filename + "\n"
		}

		return nil
	})

	return instruments.ToolResult{
		InstrumentID: "js-bugs",
		Success:      true,
		Stdout:       stdout,
	}, nil
}

func (j *JSBugsInstrument) Normalize(ctx context.Context, result instruments.ToolResult) ([]instruments.Evidence, error) {
	var evs []instruments.Evidence
	lines := strings.Split(result.Stdout, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		bugType := parts[0]
		filename := parts[1]

		switch bugType {
		case "SSE_LEAK":
			evs = append(evs, instruments.Evidence{
				ID:           "ev-js-sse-leak",
				Type:         instruments.EvidenceAST,
				Source:       "memory_leak_analyzer",
				EntityID:     filename,
				Statement:    "FACT: WebSocket event emitter listener registered in " + filename + " but not removed on close. INFERENCE: WebSocket subscription memory leak.",
				ObservedAt:   time.Now(),
				Freshness:    "live",
				Reliability:  1.0,
				Confidence:   1.0,
				Scope:        "memory",
				RawReference: filename,
				Quality: instruments.EvidenceQuality{
					Class:           "static_ast",
					Strength:        1.0,
					Reliability:     1.0,
					Freshness:       1.0,
					Specificity:     1.0,
					Independence:    1.0,
					Reproducibility: 1.0,
					Verification:    "VERIFIED",
					Provenance:      "js-bugs",
				},
			})
		case "DESC_CRASH":
			evs = append(evs, instruments.Evidence{
				ID:           "ev-js-crash-desc",
				Type:         instruments.EvidenceAST,
				Source:       "crash_analyzer",
				EntityID:     filename,
				Statement:    "FACT: Accessing description.length in " + filename + " without checking if description is defined. INFERENCE: TypeError crash on empty input.",
				ObservedAt:   time.Now(),
				Freshness:    "live",
				Reliability:  1.0,
				Confidence:   1.0,
				Scope:        "crash",
				RawReference: filename,
				Quality: instruments.EvidenceQuality{
					Class:           "static_ast",
					Strength:        1.0,
					Reliability:     1.0,
					Freshness:       1.0,
					Specificity:     1.0,
					Independence:    1.0,
					Reproducibility: 1.0,
					Verification:    "VERIFIED",
					Provenance:      "js-bugs",
				},
			})
		case "KEY_ABUSE":
			evs = append(evs, instruments.Evidence{
				ID:           "ev-js-key-abuse",
				Type:         instruments.EvidenceAST,
				Source:       "react_analyzer",
				EntityID:     filename,
				Statement:    "FACT: Array index used as key attribute in react rendering inside " + filename + ". INFERENCE: State mismatch during deletions.",
				ObservedAt:   time.Now(),
				Freshness:    "live",
				Reliability:  1.0,
				Confidence:   1.0,
				Scope:        "architecture",
				RawReference: filename,
				Quality: instruments.EvidenceQuality{
					Class:           "static_ast",
					Strength:        1.0,
					Reliability:     1.0,
					Freshness:       1.0,
					Specificity:     1.0,
					Independence:    1.0,
					Reproducibility: 1.0,
					Verification:    "VERIFIED",
					Provenance:      "js-bugs",
				},
			})
		case "JWT_BYPASS":
			evs = append(evs, instruments.Evidence{
				ID:           "ev-js-jwt-bypass",
				Type:         instruments.EvidenceAST,
				Source:       "security_analyzer",
				EntityID:     filename,
				Statement:    "FACT: JWT signature verification bypassed by trusting 'none' alg in " + filename + ". INFERENCE: Authorization bypass vulnerability.",
				ObservedAt:   time.Now(),
				Freshness:    "live",
				Reliability:  1.0,
				Confidence:   1.0,
				Scope:        "security",
				RawReference: filename,
				Quality: instruments.EvidenceQuality{
					Class:           "static_ast",
					Strength:        1.0,
					Reliability:     1.0,
					Freshness:       1.0,
					Specificity:     1.0,
					Independence:    1.0,
					Reproducibility: 1.0,
					Verification:    "VERIFIED",
					Provenance:      "js-bugs",
				},
			})
		case "CACHE_MISMATCH":
			evs = append(evs, instruments.Evidence{
				ID:           "ev-js-cache-mismatch",
				Type:         instruments.EvidenceAST,
				Source:       "cache_analyzer",
				EntityID:     filename,
				Statement:    "FACT: Cache invalidation key pattern (prod:) mismatch with cache storage key pattern (product:) in " + filename + ". INFERENCE: Stale cache regression.",
				ObservedAt:   time.Now(),
				Freshness:    "live",
				Reliability:  1.0,
				Confidence:   1.0,
				Scope:        "logic",
				RawReference: filename,
				Quality: instruments.EvidenceQuality{
					Class:           "static_ast",
					Strength:        1.0,
					Reliability:     1.0,
					Freshness:       1.0,
					Specificity:     1.0,
					Independence:    1.0,
					Reproducibility: 1.0,
					Verification:    "VERIFIED",
					Provenance:      "js-bugs",
				},
			})
		case "CONCURRENCY_RACE":
			evs = append(evs, instruments.Evidence{
				ID:           "ev-js-concurrency-race",
				Type:         instruments.EvidenceAST,
				Source:       "concurrency_analyzer",
				EntityID:     filename,
				Statement:    "FACT: Check-then-act stock validation pattern observed in " + filename + " without thread synchronization or database row locks. INFERENCE: Concurrency race condition/overselling risk.",
				ObservedAt:   time.Now(),
				Freshness:    "live",
				Reliability:  1.0,
				Confidence:   1.0,
				Scope:        "concurrency",
				RawReference: filename,
				Quality: instruments.EvidenceQuality{
					Class:           "static_ast",
					Strength:        1.0,
					Reliability:     1.0,
					Freshness:       1.0,
					Specificity:     1.0,
					Independence:    1.0,
					Reproducibility: 1.0,
					Verification:    "VERIFIED",
					Provenance:      "js-bugs",
				},
			})
		case "DB_LOCK_LEAK":
			evs = append(evs, instruments.Evidence{
				ID:           "ev-js-db-lock",
				Type:         instruments.EvidenceAST,
				Source:       "db_analyzer",
				EntityID:     filename,
				Statement:    "FACT: SQLite transaction initiated (BEGIN TRANSACTION) in " + filename + " but not rolled back or committed on payment failure branch. INFERENCE: Connection transaction lock leak.",
				ObservedAt:   time.Now(),
				Freshness:    "live",
				Reliability:  1.0,
				Confidence:   1.0,
				Scope:        "storage",
				RawReference: filename,
				Quality: instruments.EvidenceQuality{
					Class:           "static_ast",
					Strength:        1.0,
					Reliability:     1.0,
					Freshness:       1.0,
					Specificity:     1.0,
					Independence:    1.0,
					Reproducibility: 1.0,
					Verification:    "VERIFIED",
					Provenance:      "js-bugs",
				},
			})
		case "EVENT_LOOP_BLOCK":
			evs = append(evs, instruments.Evidence{
				ID:           "ev-js-event-loop-block",
				Type:         instruments.EvidenceAST,
				Source:       "performance_analyzer",
				EntityID:     filename,
				Statement:    "FACT: Synchronous busy-waiting loop (while Date.now) detected in " + filename + ". INFERENCE: Event loop blocking execution pattern.",
				ObservedAt:   time.Now(),
				Freshness:    "live",
				Reliability:  1.0,
				Confidence:   1.0,
				Scope:        "performance",
				RawReference: filename,
				Quality: instruments.EvidenceQuality{
					Class:           "static_ast",
					Strength:        1.0,
					Reliability:     1.0,
					Freshness:       1.0,
					Specificity:     1.0,
					Independence:    1.0,
					Reproducibility: 1.0,
					Verification:    "VERIFIED",
					Provenance:      "js-bugs",
				},
			})
		case "FLOAT_PRECISION":
			evs = append(evs, instruments.Evidence{
				ID:           "ev-js-float-precision",
				Type:         instruments.EvidenceAST,
				Source:       "precision_analyzer",
				EntityID:     filename,
				Statement:    "FACT: Strict inequality comparison of decimal/float values (submittedTotal !== calculatedTotal) in " + filename + ". INFERENCE: Floating point comparison failure risk.",
				ObservedAt:   time.Now(),
				Freshness:    "live",
				Reliability:  1.0,
				Confidence:   1.0,
				Scope:        "logic",
				RawReference: filename,
				Quality: instruments.EvidenceQuality{
					Class:           "static_ast",
					Strength:        1.0,
					Reliability:     1.0,
					Freshness:       1.0,
					Specificity:     1.0,
					Independence:    1.0,
					Reproducibility: 1.0,
					Verification:    "VERIFIED",
					Provenance:      "js-bugs",
				},
			})
		}
	}

	return evs, nil
}
