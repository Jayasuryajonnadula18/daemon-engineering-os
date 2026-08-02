package automation

import (
	"errors"
	"strings"
)

// TemplateConfig holds the template details for project setup automation.
type TemplateConfig struct {
	Name            string   `json:"name"`
	SupportedPacks  []string `json:"supported_packs"`
	DefaultProfiles []string `json:"default_profiles"`
	SetupRecipe     string   `json:"setup_recipe"`
}

// GetTemplate retrieves a template config by name.
func GetTemplate(name string) (*TemplateConfig, error) {
	switch strings.ToLower(name) {
	case "nextjs", "next.js saas":
		return &TemplateConfig{
			Name:            "Next.js SaaS Blueprint",
			SupportedPacks:  []string{"Node", "Git", "Docker Compose", "PostgreSQL"},
			DefaultProfiles: []string{"Frontend", "Backend", "Full Stack"},
			SetupRecipe:     "Validate packages, install node modules, verify environment variables",
		}, nil
	case "fastapi", "python fastapi":
		return &TemplateConfig{
			Name:            "Python FastAPI Microservices",
			SupportedPacks:  []string{"Python", "Docker", "PostgreSQL", "Redis"},
			DefaultProfiles: []string{"Backend", "Databases"},
			SetupRecipe:     "Configure virtualenv, install pip packages, setup Postgres schemas",
		}, nil
	case "go", "go microservices":
		return &TemplateConfig{
			Name:            "Go Microservices Hub",
			SupportedPacks:  []string{"Go", "Docker", "Kubernetes", "Redis"},
			DefaultProfiles: []string{"Backend", "Infrastructure"},
			SetupRecipe:     "Run go mod tidy, compile local dev binaries, verify Docker containers health",
		}, nil
	default:
		return nil, errors.New("unknown blueprint template specification")
	}
}

