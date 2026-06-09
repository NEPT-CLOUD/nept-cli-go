package cmd

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/config"
)

func TestLogoutCmd(t *testing.T) {
	// Mock keyring for testing
	keyring.MockInit()

	t.Run("logout when not logged in", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)

		appContainer := &app.App{
			Config: &config.Config{
				Environment: "test",
				Format:      "text",
				Verbose:     false,
			},
			Logger: slog.New(slog.NewTextHandler(errBuf, nil)),
			Out:    outBuf,
			ErrOut: errBuf,
		}

		cmd := NewLogoutCmd(appContainer)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		got := strings.TrimSpace(outBuf.String())
		expected := "You are not currently logged in."
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("logout when logged in", func(t *testing.T) {
		// Set credentials in mock keyring
		_ = keyring.Set("nept", "api-key", "test-key")
		_ = keyring.Set("nept", "user-id", "test-user")

		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)

		appContainer := &app.App{
			Config: &config.Config{
				Environment: "test",
				Format:      "text",
				Verbose:     false,
			},
			Logger: slog.New(slog.NewTextHandler(errBuf, nil)),
			Out:    outBuf,
			ErrOut: errBuf,
		}

		cmd := NewLogoutCmd(appContainer)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		got := strings.TrimSpace(outBuf.String())
		expected := "Logout successful. Credentials cleared from keychain."
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}

		// Verify credentials are deleted
		_, errKey := keyring.Get("nept", "api-key")
		_, errUser := keyring.Get("nept", "user-id")
		if errKey == nil {
			t.Error("expected api-key to be deleted, but it was found")
		}
		if errUser == nil {
			t.Error("expected user-id to be deleted, but it was found")
		}
	})

	t.Run("logout in JSON format", func(t *testing.T) {
		_ = keyring.Set("nept", "api-key", "test-key")
		_ = keyring.Set("nept", "user-id", "test-user")

		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)

		appContainer := &app.App{
			Config: &config.Config{
				Environment: "test",
				Format:      "json",
				Verbose:     false,
			},
			Logger: slog.New(slog.NewJSONHandler(errBuf, nil)),
			Out:    outBuf,
			ErrOut: errBuf,
		}

		cmd := NewLogoutCmd(appContainer)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		var resp struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse JSON response: %v", err)
		}

		if resp.Status != "success" {
			t.Errorf("expected status to be 'success', got %q", resp.Status)
		}
		if resp.Message != "Credentials cleared from keychain." {
			t.Errorf("expected message to be 'Credentials cleared from keychain.', got %q", resp.Message)
		}
	})
}
