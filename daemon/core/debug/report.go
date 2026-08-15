package debug

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RenderStdout produces a concise, developer-oriented report
func RenderStdout(inv *Investigation) string {
	var sb strings.Builder
	sb.WriteString("DAEMON DEBUG REPORT\n")
	sb.WriteString("==================================================\n")
	sb.WriteString(fmt.Sprintf("Problem:       %s\n", inv.Problem))
	sb.WriteString(fmt.Sprintf("Status:        %s\n", inv.Status))
	if inv.Reason != "" {
		sb.WriteString(fmt.Sprintf("Reason:        %s\n", inv.Reason))
	}
	sb.WriteString(fmt.Sprintf("Time:          %.2fs\n", float64(inv.DurationMs)/1000.0))
	sb.WriteString(fmt.Sprintf("AI Enhanced:   %t\n", inv.AIEnhanced))
	sb.WriteString(fmt.Sprintf("Confidence:    %.0f%%\n\n", inv.Confidence*100))

	if len(inv.RootCauses) > 0 {
		sb.WriteString("LIKELY ROOT CAUSES\n")
		sb.WriteString("--------------------------------------------------\n")
		for i, rc := range inv.RootCauses {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rc.Statement))
			sb.WriteString(fmt.Sprintf("   Confidence:  %.0f%%\n", rc.Confidence*100))
			sb.WriteString(fmt.Sprintf("   Verification: %s\n", rc.VerificationStatus))
			sb.WriteString(fmt.Sprintf("   EvidenceIDs:  %v\n\n", rc.EvidenceIDs))
		}
	} else {
		sb.WriteString("No root cause could be definitively identified from available evidence.\n")
	}

	if len(inv.Recommendations) > 0 {
		sb.WriteString("RECOMMENDED ACTION\n")
		sb.WriteString("--------------------------------------------------\n")
		for _, rec := range inv.Recommendations {
			sb.WriteString(fmt.Sprintf("- %s\n", rec))
		}
		sb.WriteString("\n")
	}

	if len(inv.Logs) > 0 {
		sb.WriteString("PROGRESSION LOGS\n")
		sb.WriteString("--------------------------------------------------\n")
		for _, log := range inv.Logs {
			sb.WriteString(fmt.Sprintf("%s\n", log))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderJSON returns a strict JSON representation with no logs or ANSI escape codes
func RenderJSON(inv *Investigation, err error) (string, error) {
	envelope := map[string]interface{}{
		"success": err == nil,
		"command": "debug",
		"data":    inv,
	}
	if err != nil {
		envelope["error"] = err.Error()
	}

	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
