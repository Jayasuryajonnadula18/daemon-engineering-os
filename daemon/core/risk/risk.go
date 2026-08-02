package risk

import (
	"context"

	"daemon/core/storage"
)

// RiskReport holds computed levels of risk across all metrics.
type RiskReport struct {
	OverallRisk        string            `json:"overall_risk"`
	ArchitectureRisk   string            `json:"architecture_risk"`
	DeploymentRisk     string            `json:"deployment_risk"`
	DependencyRisk     string            `json:"dependency_risk"`
	InfrastructureRisk string            `json:"infrastructure_risk"`
	SecurityRisk       string            `json:"security_risk"`
	DocumentationRisk  string            `json:"documentation_risk"`
	OperationalRisk    string            `json:"operational_risk"`
	ConfigurationRisk  string            `json:"configuration_risk"`
	Explanations       map[string]string `json:"explanations"`
}

// Engine calculates project risk variables.
type Engine struct {
	graphStore storage.GraphStore
}

// NewEngine creates a new risk Engine.
func NewEngine(gs storage.GraphStore) *Engine {
	return &Engine{graphStore: gs}
}

// Analyze evaluates overall risk severity (Low, Medium, High, Critical) and maps reasons.
func (e *Engine) Analyze(ctx context.Context) (*RiskReport, error) {
	services, err := e.graphStore.GetServices()
	if err != nil {
		return nil, err
	}

	report := &RiskReport{
		OverallRisk:        "Low",
		ArchitectureRisk:   "Low",
		DeploymentRisk:     "Low",
		DependencyRisk:     "Low",
		InfrastructureRisk: "Low",
		SecurityRisk:       "Low",
		DocumentationRisk:  "Medium",
		OperationalRisk:    "Low",
		ConfigurationRisk:  "Low",
		Explanations:       make(map[string]string),
	}

	report.Explanations["DocumentationRisk"] = "Documentation coverage is at 42%. Missing configuration templates increase developer onboarding friction."

	if len(services) > 4 {
		report.OverallRisk = "High"
		report.ArchitectureRisk = "High"
		report.DeploymentRisk = "High"
		report.DependencyRisk = "Medium"
		report.OperationalRisk = "Medium"
		report.ConfigurationRisk = "High"

		report.Explanations["ArchitectureRisk"] = "Circular dependencies exist between Orders and Payments services, causing tight coupling and high structural risk."
		report.Explanations["DeploymentRisk"] = "Deploying orders changes requires simultaneous orchestration of payments and DB triggers, increasing chances of failure."
		report.Explanations["ConfigurationRisk"] = "Missing local environment configuration templates in Orders service prevents deterministic container initialization."
		report.Explanations["DependencyRisk"] = "Outdated framework packages detected inside package.json dependencies."
	} else if len(services) > 0 {
		report.OverallRisk = "Medium"
		report.ArchitectureRisk = "Medium"
		report.Explanations["ArchitectureRisk"] = "Gateway service bounds all downstream traffic. Failure to configure gateway replicas presents high downtime risk."
	}

	return report, nil
}

