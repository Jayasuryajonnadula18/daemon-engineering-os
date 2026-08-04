package commands

import (
	"encoding/json"
	"fmt"
	"os"

	engContext "daemon/core/context"
	"daemon/core/maintenance"
	"daemon/core/policies"

	"github.com/spf13/cobra"
)

var (
	maintainApplyFlag  bool
	maintainFixFlag    bool
	maintainDryRunFlag bool
	maintainJsonFlag   bool
	maintainCiFlag     bool
)

var maintainCmd = &cobra.Command{
	Use:     "maintain",
	Aliases: []string{"care", "health"},
	Short:   "Perform Core Four Workspace Maintenance & Health Analysis",
	Long: `Daemon Maintenance Engine (Pillar 24) — Core Four & Full Catalog Checks:
  1. Missing / Misconfigured .env vs .env.example (Key presence strictly, zero values)
  2. Stale / Drifted Dependencies (Lockfiles, Python venvs, conflicting lockfiles)
  3. Dangling Docker State (Exited containers >24h, unreferenced images >10MB)
  4. Broken Symlinks & Uncommitted Conflict Markers (Repo symlinks and merge markers)

Operating Modes:
  • On-Demand Mode: 'daemon maintain'
  • CI Gate Mode:    'daemon maintain --ci' (exits with non-zero code on drift)
  • Apply Repairs:   'daemon maintain --apply'`,
	Run: func(cmd *cobra.Command, args []string) {
		applyFix := maintainApplyFlag || maintainFixFlag

		pe := policies.NewMemoryPolicyEngine(maintainDryRunFlag)
		ce := engContext.NewContextEngine(rt.Container.ResolveGraphStore(), rt.Container.ResolveMemoryStore())
		me := maintenance.NewMaintenanceEngine(ce, pe)

		report, err := me.RunCoreFourMaintenance(cmd.Context(), applyFix)
		if err != nil {
			fmt.Printf("❌ Maintenance Error: %v\n", err)
			if maintainCiFlag {
				os.Exit(1)
			}
			return
		}

		if maintainJsonFlag {
			data, jsonErr := json.MarshalIndent(report, "", "  ")
			if jsonErr != nil {
				fmt.Printf("❌ JSON Error: %v\n", jsonErr)
				if maintainCiFlag {
					os.Exit(1)
				}
				return
			}
			fmt.Println(string(data))
			if maintainCiFlag && report.HasDrift {
				os.Exit(1)
			}
			return
		}

		// SILENCE CONTRACT: If no drift exists, output single clean line
		if !report.HasDrift && len(report.RepairsExecuted) == 0 {
			fmt.Println("✔ Workspace maintained — 0 drift incidents detected.")
			return
		}

		fmt.Println("================================================================================")
		fmt.Println("🛠️  DAEMON MAINTENANCE ENGINE — WORKSPACE AUDIT")
		fmt.Println("================================================================================")
		fmt.Printf("Checked Directory: %s\n\n", report.CheckedDir)

		// 1. Check: Environment Drift
		if report.EnvDrift != nil {
			fmt.Println("1️⃣  ENVIRONMENT CONFIGURATION CHECK (.env vs .env.example)")
			fmt.Println("   • What was checked: Key presence between .env and .env.example (values ignored)")

			if report.EnvDrift.MissingEnvFile {
				fmt.Println("   • What was found:   ⚠️ .env file is missing!")
				if len(report.EnvDrift.MissingKeysInEnv) > 0 {
					fmt.Printf("                       Missing keys required by .env.example: %v\n", report.EnvDrift.MissingKeysInEnv)
				}
			} else if report.EnvDrift.MissingExampleFile {
				fmt.Println("   • What was found:   ⚠️ .env.example template file is missing!")
				if len(report.EnvDrift.MissingKeysInEx) > 0 {
					fmt.Printf("                       Keys in .env missing from template: %v\n", report.EnvDrift.MissingKeysInEx)
				}
			} else {
				if len(report.EnvDrift.MissingKeysInEnv) > 0 {
					fmt.Printf("   • What was found:   ⚠️ Keys present in .env.example but missing in .env: %v\n", report.EnvDrift.MissingKeysInEnv)
				}
				if len(report.EnvDrift.MissingKeysInEx) > 0 {
					fmt.Printf("   • What was found:   ⚠️ Keys present in .env but missing in .env.example: %v\n", report.EnvDrift.MissingKeysInEx)
				}
			}
			for _, note := range report.EnvDrift.MultiEnvNotes {
				fmt.Printf("   • Note:             %s\n", note)
			}
			fmt.Println("   • Safety/Reversible: Safe & Reversible (Generates/updates key templates without touching values)")
			fmt.Println()
		}

		// 2. Check: Dependency Drift
		if len(report.DepDrift) > 0 {
			fmt.Println("2️⃣  DEPENDENCY DRIFT CHECK (Lockfiles & Environment Integrity)")
			for _, dep := range report.DepDrift {
				fmt.Printf("   • What was checked: %s, %s, %s\n", dep.ManifestFile, dep.LockFile, dep.InstallDir)
				fmt.Printf("   • What was found:   ⚠️ Dependency Drift: %s\n", dep.DriftType)
				if !dep.LockTime.IsZero() {
					fmt.Printf("                       Lockfile mod time: %s\n", dep.LockTime.Format("2006-01-02 15:04:05"))
				}
				if !dep.InstallTime.IsZero() {
					fmt.Printf("                       Install mod time:  %s\n", dep.InstallTime.Format("2006-01-02 15:04:05"))
				}
				fmt.Printf("   • Suggested Fix:    Run '%s'\n", dep.SuggestCmd)
				fmt.Println("   • Guardrail:        Daemon does not auto-install dependencies. Please run the suggested command.")
			}
			fmt.Println()
		}

		// 3. Check: Dangling Docker State
		if len(report.DockerDangling) > 0 {
			fmt.Println("3️⃣  DANGLING DOCKER STATE CHECK (Exited Containers >24h & Dangling Images)")
			fmt.Println("   • What was checked: Docker host daemon status (status=exited, dangling=true)")
			fmt.Println("   • Listed Inventory:")
			for _, item := range report.DockerDangling {
				if item.Type == "container" {
					fmt.Printf("     - Exited Container: ID [%s] Name: %s (Age: %s)\n", item.ID, item.Name, item.Age)
				} else {
					fmt.Printf("     - Dangling Image:     ID [%s] Size: %s\n", item.ID, item.Size)
				}
			}
			fmt.Println("   • Safety/Reversible: Irreversible once pruned. Requires explicit '--apply' flag.")
			if !applyFix {
				fmt.Println("   • Action Required:   Run 'daemon maintain --apply' to remove this inventory.")
			}
			fmt.Println()
		}

		// 4. Check: Broken Symlinks
		if len(report.BrokenSymlinks) > 0 {
			fmt.Println("4️⃣  BROKEN SYMLINKS & DEAD REFERENCES CHECK")
			fmt.Println("   • What was checked: Repository symlink targets via os.Lstat")
			for _, sym := range report.BrokenSymlinks {
				fmt.Printf("   • What was found:   ⚠️ Broken Symlink: '%s' ──> '%s' (%s)\n", sym.Path, sym.Target, sym.Reason)
			}
			fmt.Println("   • Safety/Reversible: Reversible by restoring target or recreating symlink.")
			fmt.Println()
		}

		// 5. Check: Conflict Markers
		if len(report.ConflictMarkers) > 0 {
			fmt.Println("5️⃣  UNCOMMITTED MERGE CONFLICT MARKERS CHECK")
			fmt.Println("   • What was checked: Tracked repository text files")
			for _, cm := range report.ConflictMarkers {
				fmt.Printf("   • What was found:   ⚠️ Conflict Marker in '%s' at line %d: %s\n", cm.Path, cm.LineNumber, cm.Marker)
			}
			fmt.Println()
		}

		// 6. Check: SSH Tunnels & Port Collisions
		if len(report.SSHTunnels) > 0 {
			fmt.Println("6️⃣  ACTIVE SSH TUNNEL & PORT COLLISION CHECK")
			fmt.Println("   • What was checked: Local port bindings (netstat -ano / tasklist)")
			for _, tun := range report.SSHTunnels {
				fmt.Printf("   • What was found:   ⚠️ Active SSH Tunnel: %s\n", tun.Details)
				fmt.Printf("   • Suggested Fix:    Run '%s'\n", tun.SuggestCmd)
			}
			fmt.Println()
		}

		// Repairs Executed Output
		if len(report.RepairsExecuted) > 0 {
			fmt.Println("⚡ REPAIRS EXECUTED (--apply mode):")
			for _, r := range report.RepairsExecuted {
				fmt.Printf("  %s\n", r)
			}
			fmt.Println()
		}

		fmt.Println("================================================================================")

		if maintainCiFlag && report.HasDrift {
			os.Exit(1)
		}
	},
}

func init() {
	maintainCmd.Flags().BoolVar(&maintainApplyFlag, "apply", false, "Execute approved maintenance repairs and Docker pruning")
	maintainCmd.Flags().BoolVar(&maintainFixFlag, "fix", false, "Alias for --apply")
	maintainCmd.Flags().BoolVar(&maintainDryRunFlag, "dry-run", false, "Preview maintenance operations without making changes")
	maintainCmd.Flags().BoolVar(&maintainJsonFlag, "json", false, "Output machine-readable JSON format")
	maintainCmd.Flags().BoolVar(&maintainCiFlag, "ci", false, "CI gate mode: exits with non-zero status code if drift detected")

	rootCmd.AddCommand(maintainCmd)
}
