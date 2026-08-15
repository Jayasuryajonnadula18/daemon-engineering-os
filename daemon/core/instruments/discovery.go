package instruments

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

type EnvironmentDetector struct{}

func NewEnvironmentDetector() *EnvironmentDetector {
	return &EnvironmentDetector{}
}

// DiscoverProfile scans the workspace statically and returns a ProjectCapabilityProfile without executing external tools.
func (ed *EnvironmentDetector) DiscoverProfile(ctx context.Context, projectDir string) (*ProjectCapabilityProfile, error) {
	profile := NewProjectCapabilityProfile()

	if projectDir == "" {
		return profile, nil
	}

	// 1. Static manifest inspection
	hasGoMod := false
	if _, err := os.Stat(filepath.Join(projectDir, "go.mod")); err == nil {
		profile.Languages = append(profile.Languages, "Go")
		profile.BuildSystems = append(profile.BuildSystems, "go build")
		hasGoMod = true
	}

	hasPackageJson := false
	if _, err := os.Stat(filepath.Join(projectDir, "package.json")); err == nil {
		profile.Languages = append(profile.Languages, "JavaScript", "TypeScript")
		profile.BuildSystems = append(profile.BuildSystems, "npm", "yarn", "pnpm")
		hasPackageJson = true
	}

	hasCargo := false
	if _, err := os.Stat(filepath.Join(projectDir, "Cargo.toml")); err == nil {
		profile.Languages = append(profile.Languages, "Rust")
		profile.BuildSystems = append(profile.BuildSystems, "cargo")
		hasCargo = true
	}

	hasPyProject := false
	if _, err := os.Stat(filepath.Join(projectDir, "pyproject.toml")); err == nil {
		profile.Languages = append(profile.Languages, "Python")
		profile.BuildSystems = append(profile.BuildSystems, "pip")
		hasPyProject = true
	}

	// 2. Extension scanning (if profile is empty or to complement it)
	if !hasGoMod && !hasPackageJson && !hasCargo && !hasPyProject {
		_ = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() && (info.Name() == ".git" || info.Name() == ".daemon" || info.Name() == "node_modules") {
				return filepath.SkipDir
			}
			if !info.IsDir() {
				ext := filepath.Ext(path)
				switch ext {
				case ".go":
					profile.Languages = append(profile.Languages, "Go")
				case ".js", ".jsx", ".ts", ".tsx":
					profile.Languages = append(profile.Languages, "JavaScript", "TypeScript")
				case ".rs":
					profile.Languages = append(profile.Languages, "Rust")
				case ".py":
					profile.Languages = append(profile.Languages, "Python")
				}
			}
			return nil
		})
	}

	// Deduplicate language arrays
	profile.Languages = uniqueStrings(profile.Languages)
	profile.BuildSystems = uniqueStrings(profile.BuildSystems)

	return profile, nil
}

// IsBinaryInstalled checks if a binary exists in the system path without executing it.
func IsBinaryInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func uniqueStrings(arr []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range arr {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
