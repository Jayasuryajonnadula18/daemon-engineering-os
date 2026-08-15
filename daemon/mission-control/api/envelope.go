package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type APIMetadata struct {
	RequestID     string    `json:"request_id"`
	Timestamp     time.Time `json:"timestamp"`
	SchemaVersion string    `json:"schema_version"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIResponse[T any] struct {
	Success bool        `json:"success"`
	Command string      `json:"command"`
	Data    T           `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    APIMetadata `json:"meta"`
}

type ControlRequest struct {
	ExecutionID string `json:"execution_id"`
	NodeID      string `json:"node_id,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type StreamEvent struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

func RenderEnvelope[T any](w http.ResponseWriter, command string, data T, err error) {
	w.Header().Set("Content-Type", "application/json")

	meta := APIMetadata{
		RequestID:     fmt.Sprintf("req-%d", time.Now().UnixNano()),
		Timestamp:     time.Now(),
		SchemaVersion: "v1.2",
	}

	resp := APIResponse[T]{
		Success: err == nil,
		Command: command,
		Data:    data,
		Meta:    meta,
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = &APIError{
			Code:    "ERROR",
			Message: err.Error(),
		}
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(resp)
}
