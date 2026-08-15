package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSONResponse wraps machine-readable CLI JSON responses.
type JSONResponse struct {
	Success bool        `json:"success"`
	Command string      `json:"command"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// RenderJSON outputs a structured JSON payload to stdout and exits cleanly.
func RenderJSON(command string, data interface{}, err error) {
	resp := JSONResponse{
		Success: err == nil,
		Command: command,
		Data:    data,
	}
	if err != nil {
		resp.Error = err.Error()
	}

	bytes, marshalErr := json.MarshalIndent(resp, "", "  ")
	if marshalErr != nil {
		fmt.Fprintf(os.Stderr, "failed to serialize JSON output: %v\n", marshalErr)
		os.Exit(1)
	}

	fmt.Println(string(bytes))
}
