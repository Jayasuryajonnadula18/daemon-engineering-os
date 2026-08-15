package analysis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"daemon/core/domain"
	"daemon/core/graph"
	"daemon/core/resource"
)

const AnalyzerVersion = "1.0.0"

type Reasoner interface {
	Reason(ctx context.Context, query string) (*domain.ReasoningResult, error)
}

type DeepAnalyzerPipeline struct {
	governor            *resource.ResourceGovernor
	reasoner            Reasoner
	lifetimeAnalyzer    *ResourceLifetimeAnalyzer
	correctnessAnalyzer *CorrectnessAnalyzer
	concurrencyAnalyzer *ConcurrencyAnalyzer
	securityAnalyzer    *SecurityAnalyzer
	testIntel           *TestIntelligence
	budget              *AnalysisBudget
	cache               *AnalysisCache
	changedOnly         bool
	graph               *graph.KnowledgeGraph
}

func (p *DeepAnalyzerPipeline) SetCache(c *AnalysisCache) {
	p.cache = c
}

func (p *DeepAnalyzerPipeline) SetChangedOnly(b bool) {
	p.changedOnly = b
}

func (p *DeepAnalyzerPipeline) SetGraph(g *graph.KnowledgeGraph) {
	p.graph = g
}

func NewDeepAnalyzerPipeline(gov *resource.ResourceGovernor, r Reasoner) *DeepAnalyzerPipeline {
	if gov == nil {
		gov = resource.NewResourceGovernor(nil, resource.DefaultResourceConfig())
	}
	return &DeepAnalyzerPipeline{
		governor:            gov,
		reasoner:            r,
		lifetimeAnalyzer:    NewResourceLifetimeAnalyzer(),
		correctnessAnalyzer: NewCorrectnessAnalyzer(),
		concurrencyAnalyzer: NewConcurrencyAnalyzer(),
		securityAnalyzer:    NewSecurityAnalyzer(),
		testIntel:           NewTestIntelligence(),
	}
}

type GoFileInfo struct {
	Path        string
	PackageName string
	Imports     []string
}

func parseGoFileInfo(path string) (*GoFileInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var imports []string
	for _, imp := range f.Imports {
		if imp.Path != nil {
			val := strings.Trim(imp.Path.Value, `"`)
			imports = append(imports, val)
		}
	}
	return &GoFileInfo{
		Path:        path,
		PackageName: f.Name.Name,
		Imports:     imports,
	}, nil
}

func (p *DeepAnalyzerPipeline) RunAnalysis(ctx context.Context, projectDir string, deep bool) (*AnalysisResult, error) {
	startTime := time.Now()

	// 1. Consult Resource Governor
	dec := p.governor.Evaluate("deep_analysis", false)
	if dec.Decision == resource.DecisionDefer {
		return &AnalysisResult{
			Findings:        nil,
			Recommendations: []string{"Deep analysis deferred due to host resource pressure. Will resume when system cools down."},
			Confidence:      1.0,
			AIEnhanced:      false,
			AnalyzerStatus: []AnalyzerStatus{
				{Name: "ResourceGovernor", Available: true, RunTime: "0ms", Message: dec.Reason, Timestamp: time.Now()},
			},
			Timestamp: time.Now(),
			Status:    StatusDeferred,
		}, nil
	}

	budget := DefaultAnalysisBudget()
	if p.budget != nil {
		budget = *p.budget
	}

	// Read config file if it exists to support configurable budgets
	configPath := filepath.Join(projectDir, ".daemon", "config.json")
	if _, err := os.Stat(configPath); err == nil {
		budget = LoadAnalysisBudget(configPath)
	}

	// Check max background CPU budget
	if dec.Metrics.CPUUtilization > budget.MaxBackgroundCPU {
		return &AnalysisResult{
			Findings:        nil,
			Recommendations: []string{"Background analysis deferred: CPU budget exceeded."},
			Confidence:      1.0,
			AIEnhanced:      false,
			AnalyzerStatus: []AnalyzerStatus{
				{Name: "ResourceGovernor", Available: true, RunTime: "0ms", Message: fmt.Sprintf("CPU utilization %.2f exceeds budget %.2f", dec.Metrics.CPUUtilization, budget.MaxBackgroundCPU), Timestamp: time.Now()},
			},
			Timestamp:    time.Now(),
			Status:       StatusDeferred,
			StatusReason: "analysis_cpu_budget_exceeded",
		}, nil
	}

	var status []AnalyzerStatus
	var findings []Finding
	fileCount := 0
	cacheHits := 0
	cacheMisses := 0

	// 2. Discover all files
	var allGoFiles []string
	_ = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.Contains(path, "vendor") && !strings.Contains(path, ".daemon") {
			allGoFiles = append(allGoFiles, path)
		}
		return nil
	})

	totalFiles := len(allGoFiles)

	// Fetch dependencies and Twin version metadata
	dependencyState := "default"
	if data, err := os.ReadFile(filepath.Join(projectDir, "go.mod")); err == nil {
		h := sha256.Sum256(data)
		dependencyState = hex.EncodeToString(h[:])
	}

	twinVersion := "default"
	if p.cache != nil {
		var nodesCount, edgesCount int
		_ = p.cache.db.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&nodesCount)
		_ = p.cache.db.QueryRow("SELECT COUNT(*) FROM edges").Scan(&edgesCount)
		twinVersion = fmt.Sprintf("nodes:%d,edges:%d", nodesCount, edgesCount)
	}

	// Parse basic package/import info for dependency-aware analysis
	infoMap := make(map[string]*GoFileInfo)
	for _, path := range allGoFiles {
		if info, err := parseGoFileInfo(path); err == nil {
			infoMap[path] = info
		}
	}

	// 3. Determine if files are changed or stale
	isChangedMap := make(map[string]bool)
	for _, path := range allGoFiles {
		if p.cache != nil {
			stale, _, _, _ := p.cache.IsStale(path, AnalyzerVersion, dependencyState, twinVersion)
			if stale {
				isChangedMap[path] = true
			}
		} else {
			isChangedMap[path] = true
		}
	}

	// Trace package-level and knowledge graph-level impacted scope
	impacted := make(map[string]bool)
	var queue []string
	for _, path := range allGoFiles {
		if isChangedMap[path] {
			impacted[path] = true
			queue = append(queue, path)

			// Also query the Knowledge Graph to find downstream impacted entities/files
			if p.graph != nil {
				if downstream, err := p.graph.FindDownstreamImpact(path); err == nil {
					for _, d := range downstream {
						// Map downstream entity name back to a matching Go file path
						for _, otherPath := range allGoFiles {
							if strings.Contains(strings.ToLower(otherPath), strings.ToLower(d)) {
								if !impacted[otherPath] {
									impacted[otherPath] = true
									queue = append(queue, otherPath)
								}
							}
						}
					}
				}
			}
		}
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		currInfo := infoMap[curr]
		if currInfo == nil {
			continue
		}

		for _, other := range allGoFiles {
			if impacted[other] {
				continue
			}
			otherInfo := infoMap[other]
			if otherInfo == nil {
				continue
			}
			for _, imp := range otherInfo.Imports {
				if strings.HasSuffix(imp, "/"+currInfo.PackageName) || imp == currInfo.PackageName {
					impacted[other] = true
					queue = append(queue, other)
					break
				}
			}
		}
	}

	// Filter files based on changedOnly or full scan
	var filesToAnalyze []string
	var cachedResults []Finding
	for _, path := range allGoFiles {
		if p.changedOnly && !impacted[path] {
			// If changedOnly is active, skip analyzing unimpacted files entirely
			// and load their cached findings if any.
			if p.cache != nil {
				_, _, cached, _ := p.cache.IsStale(path, AnalyzerVersion, dependencyState, twinVersion)
				cachedResults = append(cachedResults, cached...)
				cacheHits++
			}
			continue
		}

		if p.cache != nil && !impacted[path] {
			stale, _, cached, _ := p.cache.IsStale(path, AnalyzerVersion, dependencyState, twinVersion)
			if !stale {
				cachedResults = append(cachedResults, cached...)
				cacheHits++
				continue
			}
		}

		filesToAnalyze = append(filesToAnalyze, path)
	}

	findings = append(findings, cachedResults...)

	// 4. Run analysis on the subset of files to analyze with Concurrency limits and Budgets
	budgetExceeded := false
	var statusReason string
	var bytesRead int64

	type fileResult struct {
		path     string
		findings []Finding
		hash     string
	}

	tasks := make(chan string, len(filesToAnalyze))
	resultsChan := make(chan fileResult, len(filesToAnalyze))

	concurrency := budget.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > len(filesToAnalyze) {
		concurrency = len(filesToAnalyze)
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range tasks {
				var fnds []Finding
				lf, _ := p.lifetimeAnalyzer.AnalyzeGoFile(path)
				fnds = append(fnds, lf...)
				cf, _ := p.correctnessAnalyzer.AnalyzeGoFile(path)
				fnds = append(fnds, cf...)
				cc, _ := p.concurrencyAnalyzer.AnalyzeGoFile(path)
				fnds = append(fnds, cc...)
				sf, _ := p.securityAnalyzer.AnalyzeGoFile(path)
				fnds = append(fnds, sf...)

				var currentHash string
				if p.cache != nil {
					if f, err := os.Open(path); err == nil {
						h := sha256.New()
						_, _ = io.Copy(h, f)
						currentHash = hex.EncodeToString(h.Sum(nil))
						f.Close()
					}
				}

				resultsChan <- fileResult{
					path:     path,
					findings: fnds,
					hash:     currentHash,
				}
			}
		}()
	}

	// Feed tasks and check budgets dynamically
	for _, path := range filesToAnalyze {
		select {
		case <-ctx.Done():
			budgetExceeded = true
			statusReason = "context_cancelled"
			break
		default:
		}

		if budgetExceeded {
			break
		}

		// Check MaxDuration budget
		if time.Since(startTime) > budget.MaxDuration {
			budgetExceeded = true
			statusReason = "analysis_time_budget_exceeded"
			break
		}

		// Check MaxFiles budget
		if fileCount >= budget.MaxFiles {
			budgetExceeded = true
			statusReason = "analysis_files_budget_exceeded"
			break
		}

		// Check MaxBytes budget
		if info, err := os.Stat(path); err == nil {
			bytesRead += info.Size()
			if bytesRead > budget.MaxBytes {
				budgetExceeded = true
				statusReason = "analysis_size_budget_exceeded"
				break
			}
		}

		// Check MaxMemoryMB budget
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		allocMB := int64(m.Alloc) / 1024 / 1024
		if allocMB > budget.MaxMemoryMB {
			budgetExceeded = true
			statusReason = "analysis_memory_budget_exceeded"
			break
		}

		tasks <- path
		fileCount++
	}

	close(tasks)
	wg.Wait()
	close(resultsChan)

	// Collect findings and store cache entries
	for res := range resultsChan {
		findings = append(findings, res.findings...)
		cacheMisses++

		if p.cache != nil {
			findingsJSON, _ := json.Marshal(res.findings)
			_ = p.cache.SetEntry(CacheEntry{
				FilePath:          res.path,
				FileHash:          res.hash,
				ASTHash:           res.hash,
				AnalyzerVersion:   AnalyzerVersion,
				DependencyState:   dependencyState,
				TwinVersion:       twinVersion,
				FindingCount:      len(res.findings),
				FindingsGenerated: string(findingsJSON),
				AnalyzerStatus:    "OK",
				AnalyzedAt:        time.Now(),
				Stale:             false,
			})
		}
	}

	totalDuration := time.Since(startTime).String()

	if budgetExceeded {
		status = append(status, AnalyzerStatus{Name: "PipelineOrchestrator", Available: true, RunTime: totalDuration, Message: "Analysis partial due to budget limit", Timestamp: time.Now()})
		filesRemaining := totalFiles - fileCount - cacheHits
		if filesRemaining < 0 {
			filesRemaining = 0
		}
		return &AnalysisResult{
			Findings:        findings,
			Recommendations: []string{"Analysis terminated early: " + statusReason},
			Confidence:      0.50,
			AIEnhanced:      false,
			AnalyzerStatus:  status,
			AnalyzedFiles:   fileCount,
			CacheHits:       cacheHits,
			CacheMisses:     cacheMisses,
			Timestamp:       time.Now(),
			Status:          StatusPartial,
			StatusReason:    statusReason,
			FilesRemaining:  filesRemaining,
		}, nil
	}

	// 5. Dual-Path LLM Enhancement Check
	aiEnhanced := false
	var recommendations []string

	if p.reasoner != nil {
		res, rErr := p.reasoner.Reason(ctx, "summarize deep engineering analysis findings")
		if rErr == nil && !res.InsufficientContext {
			aiEnhanced = true
			recommendations = append(recommendations, res.Answer)
		}
	}

	if !aiEnhanced {
		status = append(status, AnalyzerStatus{Name: "LLMReasoningEngine", Available: false, RunTime: "0ms", Message: "AI enhancement unavailable. Running 100% deterministic analysis engine.", Timestamp: time.Now()})
		if len(findings) == 0 {
			recommendations = append(recommendations, "Zero critical defects or resource leaks detected across analyzed modules.")
		} else {
			recommendations = append(recommendations, fmt.Sprintf("Deterministic engine identified %d potential engineering defects across %d source files.", len(findings), fileCount+cacheHits))
		}
	}

	status = append(status, AnalyzerStatus{Name: "PipelineOrchestrator", Available: true, RunTime: totalDuration, Message: "Analysis pipeline complete", Timestamp: time.Now()})

	return &AnalysisResult{
		Findings:        findings,
		Evidence:        nil,
		Recommendations: recommendations,
		Confidence:      0.90,
		AIEnhanced:      aiEnhanced,
		AnalyzerStatus:  status,
		AnalyzedFiles:   fileCount + cacheHits,
		CacheHits:       cacheHits,
		CacheMisses:     cacheMisses,
		Timestamp:       time.Now(),
		Status:          StatusCompleted,
	}, nil
}

func (p *DeepAnalyzerPipeline) SetBudget(b AnalysisBudget) {
	p.budget = &b
}

func (p *DeepAnalyzerPipeline) GetTestIntelligence() *TestIntelligence {
	return p.testIntel
}
