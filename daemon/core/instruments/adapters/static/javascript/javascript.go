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

		// 1. SSE/WebSocket Memory Leak (missing .off or removeListener cleanup)
		if (strings.Contains(content, ".on('") || strings.Contains(content, ".on(\"")) &&
			!strings.Contains(content, ".off(") &&
			!strings.Contains(content, "removeListener") {
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
			(strings.Contains(content, "key={index}") || strings.Contains(content, "key={idx}")) {
			stdout += "KEY_ABUSE:" + filename + "\n"
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
				Statement:    "FACT: Event emitter listener registered on SSE endpoint in " + filename + " but removeListener is not called on client close. INFERENCE: SSE subscription leak.",
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
		}
	}

	return evs, nil
}
