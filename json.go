package main

import (
	"encoding/json"
	"fmt"
	"io"
)

// Error codes classify runtime and usage failures in JSON responses.
const (
	errorCodeRuntime = "runtime_error"
	errorCodeUsage   = "usage_error"
)

// errorPayload wraps a JSON error response body.
type errorPayload struct {
	Error errorBody `json:"error"`
}

// errorBody describes a machine-readable CLI error.
type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// writeErrorJSON writes a structured error payload.
func writeErrorJSON(w io.Writer, code, message string, pretty bool) error {
	payload := errorPayload{
		Error: errorBody{
			Code:    code,
			Message: message,
		},
	}

	return writeJSON(w, payload, pretty)
}

// writeJSON writes a JSON payload with a trailing newline.
func writeJSON(w io.Writer, payload any, pretty bool) error {
	jsonBytes, err := marshalJSON(payload, pretty)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	if _, err = w.Write(jsonBytes); err != nil {
		return fmt.Errorf("write json: %w", err)
	}

	return nil
}

// marshalJSON encodes a payload as compact or pretty JSON.
func marshalJSON(payload any, pretty bool) ([]byte, error) {
	if pretty {
		jsonBytes, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return nil, err
		}

		return append(jsonBytes, '\n'), nil
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return append(jsonBytes, '\n'), nil
}
