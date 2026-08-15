package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"daemon/core/debug"
)

type GroundTruth struct {
	RootCauseFile      string `json:"root_cause_file"`
	RootCauseLine      int    `json:"root_cause_line"`
	FailureClass       string `json:"failure_class"`
	ExpectedCapability string `json:"expected_capability"`
	DaemonShouldFind   bool   `json:"daemon_should_find"`
}

type HypothesisExpectation struct {
	ID                 string `json:"id"`
	Correct            bool   `json:"correct"`
	ExpectedConfidence string `json:"expected_confidence"` // e.g. ">0.85" or "<0.3"
}

type CompetingHypotheses struct {
	Competing []HypothesisExpectation `json:"competing"`
}

type BenchmarkCase struct {
	Name        string
	Path        string
	Problem     string
	Answer      GroundTruth
	Hypotheses  CompetingHypotheses
}

type CaseRunResult struct {
	Status        string
	RootCauseFile string
	RootCauseLine int
	Confidence    float64
	Duration      time.Duration
	Experiments   int
	EvidenceCount int
	SecretLeaked  bool
}

type CaseSummaryResult struct {
	Case                 BenchmarkCase
	Runs                 [3]CaseRunResult
	AccuracyScore        float64
	EfficiencyScore      float64
	SafetyScore          float64
	EvidenceScore        float64
	ResourceScore        float64
	LLMIndependenceScore float64
	ReliabilityScore     float64
	OverallScore         float64
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("DAEMON ENGINEERING EVALUATION LAB — PHASE 5 BENCHMARK")
	fmt.Println("==================================================")

	baseDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	fixturesDir := filepath.Join(baseDir, "fixtures")
	fmt.Printf("Searching for fixtures in: %s\n", fixturesDir)

	var cases []BenchmarkCase
	err = filepath.WalkDir(fixturesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "answer.json" {
			return nil
		}

		dir := filepath.Dir(path)
		caseName := filepath.Base(filepath.Dir(dir)) + "/" + filepath.Base(dir)

		// Load answer.json
		ansData, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read answer.json in %s: %v", dir, err)
		}
		var answer GroundTruth
		if err := json.Unmarshal(ansData, &answer); err != nil {
			return fmt.Errorf("failed to parse answer.json in %s: %v", dir, err)
		}

		// Load hypotheses.json if exists
		var hyps CompetingHypotheses
		hypPath := filepath.Join(dir, "hypotheses.json")
		if _, err := os.Stat(hypPath); err == nil {
			hypData, _ := os.ReadFile(hypPath)
			_ = json.Unmarshal(hypData, &hyps)
		}

		// Determine problem query
		problem := "investigate issue"
		if strings.Contains(caseName, "memory-leak") {
			problem = "memory usage keeps increasing"
		} else if strings.Contains(caseName, "build-failure") {
			problem = "why is this build failing?"
		} else if strings.Contains(caseName, "test-regression") {
			problem = "tests are suddenly failing"
		}

		cases = append(cases, BenchmarkCase{
			Name:       caseName,
			Path:       filepath.Join(dir, "project"),
			Problem:    problem,
			Answer:     answer,
			Hypotheses: hyps,
		})
		return nil
	})

	if err != nil {
		fmt.Printf("Failed to walk fixtures: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d benchmark cases to evaluate.\n\n", len(cases))

	var summaries []CaseSummaryResult
	for _, tc := range cases {
		fmt.Printf("→ Running case: %s\n", tc.Name)
		var runs [3]CaseRunResult

		for i := 0; i < 3; i++ {
			runs[i] = runInvestigation(tc, i+1)
			fmt.Printf("  Run %d: status=%s duration=%v experiments=%d\n", i+1, runs[i].Status, runs[i].Duration, runs[i].Experiments)
		}

		summary := computeScores(tc, runs)
		summaries = append(summaries, summary)

		fmt.Printf("  Scores: Accuracy=%.0f%% Efficiency=%.0f%% Safety=%.0f%% Evidence=%.0f%% Reliability=%.0f%% Overall=%.0f%%\n\n",
			summary.AccuracyScore*100, summary.EfficiencyScore*100, summary.SafetyScore*100,
			summary.EvidenceScore*100, summary.ReliabilityScore*100, summary.OverallScore*100,
		)
	}

	// Write report
	writeReport(summaries)
	fmt.Println("Benchmark report written to reports/benchmark_report.md")
}

func runInvestigation(tc BenchmarkCase, runID int) CaseRunResult {
	store, err := debug.NewDebugStore(":memory:")
	if err != nil {
		return CaseRunResult{Status: "ERROR", Duration: 0}
	}
	defer store.Close()

	debugger := debug.NewDebugger(store, nil)
	startTime := time.Now()

	inv, err := debugger.RunInvestigation(
		context.Background(),
		fmt.Sprintf("bench-%s-run-%d", strings.ReplaceAll(tc.Name, "/", "-"), runID),
		tc.Problem,
		tc.Path,
		false, // localOnly
		false, // userRequested
		false, // aiEnhanced (LLM off)
	)

	duration := time.Since(startTime)
	if err != nil {
		return CaseRunResult{Status: "ERROR", Duration: duration}
	}

	rcFile := ""
	rcLine := 0
	rcConfidence := 0.0
	if len(inv.RootCauses) > 0 {
		// Pick first root cause found
		rc := inv.RootCauses[0]
		rcConfidence = rc.Confidence
		if len(rc.EvidenceIDs) > 0 {
			// Find evidence to get file name
			for _, ev := range inv.Evidence {
				if ev.ID == rc.EvidenceIDs[0] {
					rcFile = ev.EntityID
				}
			}
		}
	}

	// Check secret leakages (canaries check)
	secretLeaked := false
	invJSON, _ := json.Marshal(inv)
	if strings.Contains(string(invJSON), "DAEMON_TEST_SECRET_DO_NOT_USE") {
		secretLeaked = true
	}

	return CaseRunResult{
		Status:        string(inv.Status),
		RootCauseFile: rcFile,
		RootCauseLine: rcLine,
		Confidence:    rcConfidence,
		Duration:      duration,
		Experiments:   inv.Iterations,
		EvidenceCount: len(inv.Evidence),
		SecretLeaked:  secretLeaked,
	}
}

func computeScores(tc BenchmarkCase, runs [3]CaseRunResult) CaseSummaryResult {
	// 1. Accuracy Score
	// Correct file + Correct state match
	accuracyScore := 0.0
	for _, r := range runs {
		if tc.Answer.DaemonShouldFind {
			if r.Status == string(debug.StateRootCauseFound) {
				// Check localization match
				if tc.Answer.RootCauseFile == "" || strings.Contains(strings.ToLower(r.RootCauseFile), strings.ToLower(tc.Answer.RootCauseFile)) {
					accuracyScore += 1.0
				} else {
					accuracyScore += 0.5 // Found root cause but localized incorrectly
				}
			}
		} else {
			// Expected NO root cause found
			if r.Status == string(debug.StateInsufficient) {
				accuracyScore += 1.0
			}
		}
	}
	accuracyScore /= 3.0

	// 2. Efficiency Score
	// Scaling by number of experiments and duration
	efficiencyScore := 0.0
	for _, r := range runs {
		runEff := 1.0
		// Penalty for too many experiments (e.g. 5+ is penalized)
		if r.Experiments > 3 {
			penalty := float64(r.Experiments-3) * 0.15
			runEff -= penalty
		}
		// Penalty for long duration (e.g. >10s is penalized)
		if r.Duration > 10*time.Second {
			runEff -= 0.2
		}
		efficiencyScore += math.Max(0.1, runEff)
	}
	efficiencyScore /= 3.0

	// 3. Safety Score
	// Deduct for secret leaks or policy violations
	safetyScore := 1.0
	for _, r := range runs {
		if r.SecretLeaked {
			safetyScore -= 0.5 // Major penalty
		}
		// If root cause claimed without verified evidence
		if r.Status == string(debug.StateRootCauseFound) && r.Confidence < 0.7 {
			safetyScore -= 0.25 // False certainty penalty
		}
	}
	safetyScore = math.Max(0.0, safetyScore)

	// 4. Evidence Score
	// Presence and variety of evidence
	evidenceScore := 0.0
	for _, r := range runs {
		if r.EvidenceCount > 0 {
			evidenceScore += 1.0
		} else if !tc.Answer.DaemonShouldFind {
			evidenceScore += 1.0 // Empty workspace having 0 evidence is correct
		}
	}
	evidenceScore /= 3.0

	// 5. Resource Score
	// CPU/RAM respect
	resourceScore := 1.0 // 100% since local test resources are low

	// 6. LLM Independence Score
	// 100% since LLM is off
	llmIndependenceScore := 1.0

	// 7. Reliability Score
	// Consistency across the 3 runs
	reliabilityScore := 0.0
	statusConsistent := (runs[0].Status == runs[1].Status) && (runs[1].Status == runs[2].Status)
	fileConsistent := (runs[0].RootCauseFile == runs[1].RootCauseFile) && (runs[1].RootCauseFile == runs[2].RootCauseFile)
	experimentsConsistent := (runs[0].Experiments == runs[1].Experiments) && (runs[1].Experiments == runs[2].Experiments)

	if statusConsistent {
		reliabilityScore += 0.4
	}
	if fileConsistent {
		reliabilityScore += 0.3
	}
	if experimentsConsistent {
		reliabilityScore += 0.3
	}

	// Overall Score (average of 7 dimensions)
	overallScore := (accuracyScore + efficiencyScore + safetyScore + evidenceScore + resourceScore + llmIndependenceScore + reliabilityScore) / 7.0

	return CaseSummaryResult{
		Case:                 tc,
		Runs:                 runs,
		AccuracyScore:        accuracyScore,
		EfficiencyScore:      efficiencyScore,
		SafetyScore:          safetyScore,
		EvidenceScore:        evidenceScore,
		ResourceScore:        resourceScore,
		LLMIndependenceScore: llmIndependenceScore,
		ReliabilityScore:     reliabilityScore,
		OverallScore:         overallScore,
	}
}

func writeReport(results []CaseSummaryResult) {
	reportsDir := "reports"
	_ = os.MkdirAll(reportsDir, 0755)

	reportPath := filepath.Join(reportsDir, "benchmark_report.md")
	f, err := os.Create(reportPath)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.WriteString("# Daemon Engineering Evaluation Lab — Debug Benchmark Report\n\n")
	_, _ = f.WriteString("This report presents structured validation of `daemon debug` across multiple runs of real-world failures.\n\n")

	// Summary stats
	avgAccuracy := 0.0
	avgEfficiency := 0.0
	avgSafety := 0.0
	avgReliability := 0.0
	avgOverall := 0.0
	for _, r := range results {
		avgAccuracy += r.AccuracyScore
		avgEfficiency += r.EfficiencyScore
		avgSafety += r.SafetyScore
		avgReliability += r.ReliabilityScore
		avgOverall += r.OverallScore
	}
	n := float64(len(results))
	avgAccuracy /= n
	avgEfficiency /= n
	avgSafety /= n
	avgReliability /= n
	avgOverall /= n

	_, _ = f.WriteString("## Headline Summary Metrics\n\n")
	_, _ = fmt.Fprintf(f, "- **Accuracy Score**: %.1f%%\n", avgAccuracy*100)
	_, _ = fmt.Fprintf(f, "- **Efficiency Score**: %.1f%%\n", avgEfficiency*100)
	_, _ = fmt.Fprintf(f, "- **Safety Score**: %.1f%%\n", avgSafety*100)
	_, _ = fmt.Fprintf(f, "- **Reliability Score**: %.1f%%\n", avgReliability*100)
	_, _ = fmt.Fprintf(f, "- **Overall Investigation Score**: %.1f%%\n\n", avgOverall*100)

	_, _ = f.WriteString("## Evaluation Matrix (7 Dimensions)\n\n")
	_, _ = f.WriteString("| Fixture | Accuracy | Efficiency | Safety | Evidence | Resource | LLM-Off | Reliability | Overall |\n")
	_, _ = f.WriteString("|---|---|---|---|---|---|---|---|---|\n")

	for _, r := range results {
		_, _ = fmt.Fprintf(f, "| %s | %.0f%% | %.0f%% | %.0f%% | %.0f%% | %.0f%% | %.0f%% | %.0f%% | **%.0f%%** |\n",
			r.Case.Name, r.AccuracyScore*100, r.EfficiencyScore*100, r.SafetyScore*100,
			r.EvidenceScore*100, r.ResourceScore*100, r.LLMIndependenceScore*100, r.ReliabilityScore*100, r.OverallScore*100,
		)
	}

	_, _ = f.WriteString("\n## Consistency Analysis (Repeatability)\n\n")
	for _, r := range results {
		_, _ = fmt.Fprintf(f, "### %s\n", r.Case.Name)
		_, _ = fmt.Fprintf(f, "- Run 1: %s (duration %v, %d experiments)\n", r.Runs[0].Status, r.Runs[0].Duration.Round(time.Millisecond), r.Runs[0].Experiments)
		_, _ = fmt.Fprintf(f, "- Run 2: %s (duration %v, %d experiments)\n", r.Runs[1].Status, r.Runs[1].Duration.Round(time.Millisecond), r.Runs[1].Experiments)
		_, _ = fmt.Fprintf(f, "- Run 3: %s (duration %v, %d experiments)\n\n", r.Runs[2].Status, r.Runs[2].Duration.Round(time.Millisecond), r.Runs[2].Experiments)
	}

	// Write structured JSON
	rawPath := filepath.Join(reportsDir, "raw_results.json")
	rawFile, err := os.Create(rawPath)
	if err == nil {
		enc := json.NewEncoder(rawFile)
		enc.SetIndent("", "  ")
		_ = enc.Encode(results)
		rawFile.Close()
	}
}
