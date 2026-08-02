package integrations

import (
	"context"
	"strings"

	"daemon/core/storage"
)

// ChangeReport summarizes the affected entities of a file modification.
type ChangeReport struct {
	FilesChanged     []string `json:"files_changed"`
	AffectedServices []string `json:"affected_services"`
	AffectedAPIs     []string `json:"affected_apis"`
	DbImpact         string   `json:"db_impact"`
	DocsToUpdate     []string `json:"docs_to_update"`
	RiskSeverity     string   `json:"risk_severity"`
	SuggestedTests   []string `json:"suggested_tests"`
}

// Engine acts as the coordinator tracking filesystem event implications.
type Engine struct {
	graphStore storage.GraphStore
}

// NewEngine instantiates a new coordination Engine.
func NewEngine(gs storage.GraphStore) *Engine {
	return &Engine{graphStore: gs}
}

// AnalyzeChange evaluates files and returns structural affected reports.
func (e *Engine) AnalyzeChange(ctx context.Context, filepath string) (*ChangeReport, error) {
	report := &ChangeReport{
		FilesChanged:     []string{filepath},
		AffectedServices: []string{"None"},
		AffectedAPIs:     []string{"None"},
		DbImpact:         "None (no database schema files modified)",
		DocsToUpdate:     []string{"README.md (general workspace verification)"},
		RiskSeverity:     "Low",
		SuggestedTests:   []string{"Integration test suite verification"},
	}

	filepathLower := strings.ToLower(filepath)

	if strings.Contains(filepathLower, "payments") {
		report.AffectedServices = []string{"Payments Service API", "Orders Service API (coupled dependents)"}
		report.AffectedAPIs = []string{"POST /api/payments/checkout"}
		report.RiskSeverity = "High"
		report.SuggestedTests = []string{"payments-api unit tests", "orders-sync validation checks"}
	} else if strings.Contains(filepathLower, "db") || strings.Contains(filepathLower, "migration") {
		report.AffectedServices = []string{"PostgreSQL database"}
		report.DbImpact = "Database schema migration detected. Migrations need execution."
		report.RiskSeverity = "Medium"
		report.SuggestedTests = []string{"database connectivity validation"}
		report.DocsToUpdate = []string{"Environment Documentation (database schema variables)"}
	} else if strings.Contains(filepathLower, "auth") {
		report.AffectedServices = []string{"Authentication Service"}
		report.AffectedAPIs = []string{"GET /api/auth/session", "POST /api/auth/token"}
		report.RiskSeverity = "Critical"
		report.SuggestedTests = []string{"auth-service session tests"}
	}

	return report, nil
}

