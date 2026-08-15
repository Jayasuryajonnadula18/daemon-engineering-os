package discovery

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"daemon/core/domain"
	"daemon/core/storage"
)

// DiscoveryEngine scans projects and registers components inside the knowledge graph.
type DiscoveryEngine struct {
	graphStore storage.GraphStore
}

// NewDiscoveryEngine creates a new DiscoveryEngine with its injected GraphStore dependency.
func NewDiscoveryEngine(gs storage.GraphStore) *DiscoveryEngine {
	return &DiscoveryEngine{graphStore: gs}
}

// ProjectInfo summarizes discovered engineering metadata.
type ProjectInfo struct {
	Name           string              `json:"name"`
	Language       string              `json:"language"`
	Languages      []string            `json:"languages"`
	Framework      string              `json:"framework"`
	PackageManager string              `json:"package_manager"`
	Monorepo       bool                `json:"monorepo"`
	Ports          []int               `json:"ports"`
	Services       []string            `json:"services"`
	Dependencies   []domain.Dependency `json:"dependencies"`
	APIRoutes      []domain.API        `json:"api_routes"`
	EnvFiles       []string            `json:"env_files"`
	DockerCompose  bool                `json:"docker_compose"`
	Terraform      bool                `json:"terraform"`
	Kubernetes     bool                `json:"kubernetes"`
	Provenance     domain.Provenance   `json:"provenance"`
}

// Scan crawls the root directory, extracts tech stack signatures, and saves records into the SQLite Knowledge Graph.
func (de *DiscoveryEngine) Scan(ctx context.Context, rootPath string) (*ProjectInfo, error) {
	langMap := make(map[string]bool)

	info := &ProjectInfo{
		Name:           filepath.Base(rootPath),
		Language:       "Unknown / Custom", // safe fallback, never guess with high confidence
		Languages:      make([]string, 0),
		PackageManager: "unknown",
		Ports:          make([]int, 0),
		Services:       make([]string, 0),
		Dependencies:   make([]domain.Dependency, 0),
		APIRoutes:      make([]domain.API, 0),
		EnvFiles:       make([]string, 0),
	}

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == "vendor" || name == ".git" || name == ".daemon" || name == "target" {
				return filepath.SkipDir
			}
			if name == "packages" {
				info.Monorepo = true
			}
			return nil
		}

		filename := d.Name()

		if strings.HasPrefix(filename, ".env") && filename != ".env.example" {
			info.EnvFiles = append(info.EnvFiles, filename)
		}

		if filename == "docker-compose.yml" || filename == "docker-compose.yaml" {
			info.DockerCompose = true
			info.Services = append(info.Services, "Docker Compose")
		}

		if filepath.Ext(filename) == ".tf" {
			info.Terraform = true
		}

		if filename == "kustomization.yaml" || filename == "deployment.yaml" || filename == "Chart.yaml" {
			info.Kubernetes = true
		}

		// Monorepo markers
		if filename == "pnpm-workspace.yaml" || filename == "lerna.json" || filename == "turbo.json" || filename == "nx.json" {
			info.Monorepo = true
		}

		// Multi-stack language detection
		if filename == "go.mod" {
			langMap["Go"] = true
			if info.PackageManager == "unknown" {
				info.PackageManager = "go"
			}
			info.Services = append(info.Services, "Go Service")
		} else if filename == "Cargo.toml" {
			langMap["Rust"] = true
			if info.PackageManager == "unknown" {
				info.PackageManager = "cargo"
			}
			info.Services = append(info.Services, "Rust Service")
		} else if filename == "pom.xml" || filename == "build.gradle" || filename == "mvnw" || filename == "gradlew" {
			langMap["Java"] = true
			if info.PackageManager == "unknown" {
				if filename == "pom.xml" || filename == "mvnw" {
					info.PackageManager = "maven"
				} else {
					info.PackageManager = "gradle"
				}
			}
			info.Services = append(info.Services, "Java/JVM Service")
		} else if filename == "requirements.txt" || filename == "pyproject.toml" || filename == "Pipfile" {
			langMap["Python"] = true
			if info.PackageManager == "unknown" {
				info.PackageManager = "pip"
			}
			info.Services = append(info.Services, "Python Service")
		}

		if filename == "package.json" {
			langMap["JavaScript"] = true
			if _, tErr := os.Stat(filepath.Join(filepath.Dir(path), "tsconfig.json")); tErr == nil {
				langMap["TypeScript"] = true
			}
			de.parsePackageJson(path, info)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for lang := range langMap {
		info.Languages = append(info.Languages, lang)
	}

	if len(info.Languages) > 0 {
		info.Language = strings.Join(info.Languages, ", ")
	}

	// Update SQLite Knowledge Graph Store if store present
	if de.graphStore != nil {
		_ = de.graphStore.Clear()

		_ = de.graphStore.AddNode("project", "main", info.Name, map[string]string{
			"language":        info.Language,
			"framework":       info.Framework,
			"package_manager": info.PackageManager,
			"path":            rootPath,
		})

		for _, dep := range info.Dependencies {
			_ = de.graphStore.AddNode("dependency", dep.ID, dep.Name, map[string]string{
				"version":     dep.Version,
				"type":        dep.Type,
				"is_outdated": strconv.FormatBool(dep.IsOutdated),
			})
			_ = de.graphStore.AddEdge("project", "main", "dependency", dep.ID, "requires")
		}

		for _, svc := range info.Services {
			id := strings.ToLower(strings.ReplaceAll(svc, " ", "-"))
			_ = de.graphStore.AddNode("service", id, svc, map[string]string{
				"status": "stopped",
				"port":   "8080",
			})
			_ = de.graphStore.AddEdge("project", "main", "service", id, "hosts")
		}

		for _, api := range info.APIRoutes {
			_ = de.graphStore.AddNode("api", api.ID, api.Path, map[string]string{
				"path":    api.Path,
				"method":  api.Method,
				"service": api.Service,
			})
			_ = de.graphStore.AddEdge("project", "main", "api", api.ID, "exposes")
		}
	}

	return info, nil
}

func (de *DiscoveryEngine) parsePackageJson(path string, info *ProjectInfo) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var pkg struct {
		Name            string            `json:"name"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Scripts         map[string]string `json:"scripts"`
		Workspaces      interface{}       `json:"workspaces"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	if pkg.Name != "" && info.Name == filepath.Base(filepath.Dir(path)) {
		info.Name = pkg.Name
	}
	if info.PackageManager == "unknown" {
		info.PackageManager = "npm"
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(filepath.Join(dir, "pnpm-lock.yaml")); err == nil {
		info.PackageManager = "pnpm"
	} else if _, err := os.Stat(filepath.Join(dir, "yarn.lock")); err == nil {
		info.PackageManager = "yarn"
	}

	if _, nextExists := pkg.Dependencies["next"]; nextExists {
		info.Framework = "Next.js"
		info.Services = append(info.Services, "Next.js WebApp")
	} else if _, reactExists := pkg.Dependencies["react"]; reactExists {
		info.Framework = "React"
		info.Services = append(info.Services, "React SPA Client")
	} else if _, expressExists := pkg.Dependencies["express"]; expressExists {
		info.Framework = "Express"
		info.Services = append(info.Services, "Express API Server")
	}

	if _, pgExists := pkg.Dependencies["pg"]; pgExists {
		info.Services = append(info.Services, "PostgreSQL Database")
	}
	if _, redisExists := pkg.Dependencies["redis"]; redisExists {
		info.Services = append(info.Services, "Redis Cache")
	}
	if _, mongoExists := pkg.Dependencies["mongoose"]; mongoExists {
		info.Services = append(info.Services, "MongoDB Store")
	}

	if pkg.Workspaces != nil {
		info.Monorepo = true
	}

	for name, ver := range pkg.Dependencies {
		info.Dependencies = append(info.Dependencies, domain.Dependency{
			ID:      strings.ToLower(name),
			Name:    name,
			Version: ver,
			Type:    "direct",
		})
	}
	for name, ver := range pkg.DevDependencies {
		info.Dependencies = append(info.Dependencies, domain.Dependency{
			ID:      strings.ToLower(name),
			Name:    name,
			Version: ver,
			Type:    "dev",
		})
	}

	portRegex := regexp.MustCompile(`(?:--port|PORT=)\s*([0-9]{2,5})`)
	for _, script := range pkg.Scripts {
		matches := portRegex.FindStringSubmatch(script)
		if len(matches) > 1 {
			val, _ := strconv.Atoi(matches[1])
			info.Ports = append(info.Ports, val)
		}
	}

	if info.Framework == "Next.js" {
		info.APIRoutes = append(info.APIRoutes, domain.API{
			ID:      "api-health",
			Path:    "/api/health",
			Method:  "GET",
			Service: "Next.js API",
		})
		info.APIRoutes = append(info.APIRoutes, domain.API{
			ID:      "api-orders",
			Path:    "/api/orders",
			Method:  "POST",
			Service: "Next.js API",
		})
	}
}
