package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/config"
)

func TestPrintResult(t *testing.T) {
	t.Run("text format", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		app := &App{
			Config: &config.Config{Format: "text"},
			Out:    outBuf,
		}

		err := app.PrintResult("Hello world\n", map[string]string{"msg": "Hello world"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := outBuf.String()
		expected := "Hello world\n"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("json format", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		app := &App{
			Config: &config.Config{Format: "json"},
			Out:    outBuf,
		}

		data := map[string]string{"msg": "Hello world"}
		err := app.PrintResult("Hello world", data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var got map[string]string
		if err := json.Unmarshal(outBuf.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if got["msg"] != "Hello world" {
			t.Errorf("expected %q, got %q", "Hello world", got["msg"])
		}
	})
}

func TestPrintErr(t *testing.T) {
	t.Run("text format error", func(t *testing.T) {
		errBuf := new(bytes.Buffer)
		app := &App{
			Config: &config.Config{Format: "text"},
			ErrOut: errBuf,
		}

		testErr := errors.New("something went wrong")
		app.PrintErr("SOME_ERROR", testErr)

		got := errBuf.String()
		expected := "Error: something went wrong\n"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("json format error", func(t *testing.T) {
		errBuf := new(bytes.Buffer)
		app := &App{
			Config: &config.Config{Format: "json"},
			ErrOut: errBuf,
		}

		testErr := errors.New("something went wrong")
		app.PrintErr("SOME_ERROR", testErr)

		type ErrorDetail struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		type ErrorResponse struct {
			Status string      `json:"status"`
			Error  ErrorDetail `json:"error"`
		}

		var resp ErrorResponse
		if err := json.Unmarshal(errBuf.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if resp.Status != "error" {
			t.Errorf("expected status 'error', got %q", resp.Status)
		}
		if resp.Error.Code != "SOME_ERROR" {
			t.Errorf("expected code 'SOME_ERROR', got %q", resp.Error.Code)
		}
		if resp.Error.Message != "something went wrong" {
			t.Errorf("expected message 'something went wrong', got %q", resp.Error.Message)
		}
	})

	t.Run("fallback JSON error with os.Args", func(t *testing.T) {
		// Mock os.Args
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"nept", "hello", "--format", "json"}

		errBuf := new(bytes.Buffer)
		app := &App{
			Config: nil, // Config is nil
			ErrOut: errBuf,
		}

		testErr := errors.New("command parsing failed")
		app.PrintErr("PARSE_FAILED", testErr)

		type ErrorDetail struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		type ErrorResponse struct {
			Status string      `json:"status"`
			Error  ErrorDetail `json:"error"`
		}

		var resp ErrorResponse
		if err := json.Unmarshal(errBuf.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal fallback JSON: %v", err)
		}

		if resp.Error.Code != "PARSE_FAILED" {
			t.Errorf("expected code 'PARSE_FAILED', got %q", resp.Error.Code)
		}
	})
}
