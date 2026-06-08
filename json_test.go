package main

import (
	"bytes"
	"errors"
	"testing"
)

type successPayload struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestWriteJSONCompact(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	err := writeJSON(&stdout, successPayload{Name: "alpha", Count: 2}, false)
	if err != nil {
		t.Fatalf("writeJSON returned error: %v", err)
	}

	want := "{\"name\":\"alpha\",\"count\":2}\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestWriteJSONPretty(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	err := writeJSON(&stdout, successPayload{Name: "alpha", Count: 2}, true)
	if err != nil {
		t.Fatalf("writeJSON returned error: %v", err)
	}

	want := "{\n  \"name\": \"alpha\",\n  \"count\": 2\n}\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestWriteErrorJSONCompact(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	err := writeErrorJSON(&stderr, errorCodeRuntime, "boom", false)
	if err != nil {
		t.Fatalf("writeErrorJSON returned error: %v", err)
	}

	want := "{\"error\":{\"code\":\"runtime_error\",\"message\":\"boom\"}}\n"
	if stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
	}
}

func TestWriteErrorJSONPretty(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	err := writeErrorJSON(&stderr, errorCodeUsage, "missing command", true)
	if err != nil {
		t.Fatalf("writeErrorJSON returned error: %v", err)
	}

	want := "{\n  \"error\": {\n    \"code\": \"usage_error\",\n    \"message\": \"missing command\"\n  }\n}\n"
	if stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
	}
}

func TestWriteJSONWriteError(t *testing.T) {
	t.Parallel()

	err := writeJSON(errWriter{}, successPayload{Name: "alpha", Count: 2}, false)
	if err == nil {
		t.Fatal("writeJSON error = nil, want non-nil")
	}
}

type errWriter struct{}

func (errWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}
