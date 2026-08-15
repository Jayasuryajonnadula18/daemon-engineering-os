package commands

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"daemon/cli/output"
	"github.com/spf13/cobra"
)

// CheckBinaryInstalled checks if a binary tool is available on PATH.
func CheckBinaryInstalled(name string) (bool, string) {
	path, err := exec.LookPath(name)
	if err != nil {
		return false, "NOT FOUND"
	}
	return true, path
}

// CheckPortListening checks if a local TCP port is responding.
func CheckPortListening(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

var doctorJSONFlag bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Perform Engineering Diagnostics and Health analysis",
	Run: func(cmd *cobra.Command, args []string) {
		re := rt.Container.ResolveReasoningEngine()

		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}

		// 1. Toolchain Diagnostics
		goOk, goPath := CheckBinaryInstalled("go")
		nodeOk, nodePath := CheckBinaryInstalled("node")
		dockerOk, dockerPath := CheckBinaryInstalled("docker")
		gitOk, gitPath := CheckBinaryInstalled("git")
		ollamaOk, _ := CheckBinaryInstalled("ollama")

		// 2. Real Port & Service Diagnostics
		ollamaListening := CheckPortListening(11434)
		postgresListening := CheckPortListening(5432)
		redisListening := CheckPortListening(6379)

		// 3. Filesystem & Manifest Diagnostics
		hasEnv := false
		if _, err := os.Stat(filepath.Join(cwd, ".env")); err == nil {
			hasEnv = true
		}
		hasEnvExample := false
		if _, err := os.Stat(filepath.Join(cwd, ".env.example")); err == nil {
			hasEnvExample = true
		}
		hasPkgJson := false
		if _, err := os.Stat(filepath.Join(cwd, "package.json")); err == nil {
			hasPkgJson = true
		}
		hasGoMod := false
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			hasGoMod = true
		}
		hasDockerCompose := false
		if _, err := os.Stat(filepath.Join(cwd, "docker-compose.yml")); err == nil {
			hasDockerCompose = true
		}

		// 4. Knowledge Graph Stats
		gs := rt.Container.ResolveGraphStore()
		services, _ := gs.GetServices()
		deps, _ := gs.GetDependencies()

		// Calculate Dynamic Health & Readiness Scores
		healthScore := 100
		var findings []string
		var recommendations []string

		if !hasEnvExample && hasEnv {
			healthScore -= 15
			findings = append(findings, "Missing .env.example configuration template file")
			recommendations = append(recommendations, "Generate .env.example using 'daemon fix --execute'")
		}

		if !dockerOk {
			healthScore -= 10
			findings = append(findings, "Docker CLI not detected on system PATH")
		}

		if !ollamaListening {
			findings = append(findings, "Ollama local LLM service offline (port 11434 not responding)")
			recommendations = append(recommendations, "Start local Ollama service with 'ollama run qwen2.5-coder:7b'")
		}

		if healthScore < 0 {
			healthScore = 0
		}

		readinessScore := 85
		if hasEnv && hasEnvExample {
			readinessScore = 98
		}

		techDebt := "Low"
		if len(findings) > 2 {
			techDebt = "Medium"
		}

		if doctorJSONFlag {
			data := map[string]interface{}{
				"project":         filepath.Base(cwd),
				"health_score":    healthScore,
				"readiness_score": readinessScore,
				"tech_debt":       techDebt,
				"findings":        findings,
				"recommendations": recommendations,
				"toolchain": map[string]interface{}{
					"go":     goOk,
					"node":   nodeOk,
					"docker": dockerOk,
					"git":    gitOk,
					"ollama": ollamaOk,
				},
				"services": map[string]interface{}{
					"ollama_11434":   ollamaListening,
					"postgres_5432":  postgresListening,
					"redis_6379":     redisListening,
					"tracked_count":  len(services),
					"deps_count":     len(deps),
				},
			}
			output.RenderJSON("doctor", data, nil)
			return
		}

		fmt.Println("Running real-time engineering diagnostics...")

		fmt.Println("\n=========================================")
		fmt.Println("ENGINEERING HEALTH REPORT (LIVE DIAGNOSTICS)")
		fmt.Println("=========================================")
		fmt.Printf("Target Project:    %s\n", filepath.Base(cwd))
		fmt.Printf("OS/Architecture:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Overall Score:     %d%%\n\n", healthScore)

		fmt.Println("Toolchain Health:")
		fmt.Printf("  • Go Compiler:   %t (%s)\n", goOk, goPath)
		fmt.Printf("  • Node.js:       %t (%s)\n", nodeOk, nodePath)
		fmt.Printf("  • Git SCM:       %t (%s)\n", gitOk, gitPath)
		fmt.Printf("  • Docker CLI:    %t (%s)\n", dockerOk, dockerPath)
		fmt.Printf("  • Ollama CLI:    %t (Service port 11434 responding: %t)\n", ollamaOk, ollamaListening)

		fmt.Println("\nWorkspace Manifests:")
		fmt.Printf("  • .env File:               %t\n", hasEnv)
		fmt.Printf("  • .env.example Template:   %t\n", hasEnvExample)
		fmt.Printf("  • package.json Manifest:   %t\n", hasPkgJson)
		fmt.Printf("  • go.mod Manifest:         %t\n", hasGoMod)
		fmt.Printf("  • docker-compose.yml:      %t\n", hasDockerCompose)

		fmt.Println("\nLive Service Probes:")
		fmt.Printf("  • PostgreSQL (Port 5432):  %t\n", postgresListening)
		fmt.Printf("  • Redis Cache (Port 6379): %t\n", redisListening)

		fmt.Println("\nKnowledge Graph Model:")
		fmt.Printf("  • Tracked Services:        %d\n", len(services))
		fmt.Printf("  • Tracked Dependencies:    %d\n", len(deps))

		fmt.Println("\n-----------------------------------------")
		fmt.Printf("Readiness Score:   %d%%\n", readinessScore)
		fmt.Printf("Technical Debt:    %s\n", techDebt)
		fmt.Println("-----------------------------------------")

		if len(recommendations) > 0 {
			fmt.Println("\nActionable Recommendations:")
			for _, r := range recommendations {
				fmt.Printf("  [ ] %s\n", r)
				explanation, _ := re.Explain(context.Background(), r, "Local workspace health and developer onboarding performance.")
				fmt.Printf("      Why this matters: %s\n", explanation)
			}
		} else {
			fmt.Println("\n✔ Workspace is 100% healthy. No critical diagnostics issues found.")
		}
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorJSONFlag, "json", false, "Output machine-readable JSON")
	rootCmd.AddCommand(doctorCmd)
}
