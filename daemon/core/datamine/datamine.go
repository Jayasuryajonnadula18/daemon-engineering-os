package datamine

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type CommitCategory string

const (
	CategoryFeature  CommitCategory = "feature"
	CategoryFix      CommitCategory = "fix"
	CategoryRefactor CommitCategory = "refactor"
	CategoryDocs     CommitCategory = "docs"
	CategoryChore    CommitCategory = "chore"
)

type ClassificationReport struct {
	TotalCommits     int                    `json:"total_commits"`
	CategoryCounts   map[CommitCategory]int `json:"category_counts"`
	ExcludedCount    int                    `json:"excluded_count"`
	ExcludedLog      []string               `json:"excluded_log"`
	IsBalancedSample bool                   `json:"is_balanced_sample"`
}

// ClassifyCommitSubject categorizes a commit message by its subject line.
// Returns empty string if unmatched, ensuring zero fallback to generic "testing" category.
func ClassifyCommitSubject(subject string) CommitCategory {
	lower := strings.ToLower(subject)
	if strings.HasPrefix(lower, "feat") || strings.Contains(lower, "feature") || strings.Contains(lower, "add ") {
		return CategoryFeature
	}
	if strings.HasPrefix(lower, "fix") || strings.Contains(lower, "bug") || strings.Contains(lower, "repair") || strings.Contains(lower, "patch") {
		return CategoryFix
	}
	if strings.HasPrefix(lower, "refactor") || strings.Contains(lower, "clean") || strings.Contains(lower, "restructure") {
		return CategoryRefactor
	}
	if strings.HasPrefix(lower, "docs") || strings.Contains(lower, "readme") || strings.Contains(lower, "doc") {
		return CategoryDocs
	}
	if strings.HasPrefix(lower, "chore") || strings.Contains(lower, "bump") || strings.Contains(lower, "build") {
		return CategoryChore
	}

	// Strictly NO default fallback to "testing". Return empty string for unmatched commits.
	return ""
}

// MineCommits parses commit history from a Git repository path and classifies subjects.
func MineCommits(ctx context.Context, repoPath string, maxCount int) (*ClassificationReport, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "log", fmt.Sprintf("-n%d", maxCount), "--pretty=format:%H|%s")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch git log: %w", err)
	}

	report := &ClassificationReport{
		CategoryCounts: make(map[CommitCategory]int),
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) < 2 {
			continue
		}
		report.TotalCommits++
		subject := parts[1]
		cat := ClassifyCommitSubject(subject)
		if cat != "" {
			report.CategoryCounts[cat]++
		} else {
			report.ExcludedCount++
			report.ExcludedLog = append(report.ExcludedLog, fmt.Sprintf("excluded — no category match: %s (%s)", parts[0][:7], subject))
		}
	}

	// Calculate balance (no category exceeds 3x smallest count among non-zero categories)
	minCount := 999999
	maxCatCount := 0
	for _, count := range report.CategoryCounts {
		if count < minCount {
			minCount = count
		}
		if count > maxCatCount {
			maxCatCount = count
		}
	}
	if minCount > 0 && maxCatCount <= minCount*3 {
		report.IsBalancedSample = true
	}

	return report, nil
}
