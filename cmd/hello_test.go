package cmd

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/config"
)

func TestHelloCmd(t *testing.T) {
	t.Run("default greeting", func(t *testing.T) {
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

		cmd := NewHelloCmd(appContainer)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{}) // uses default flag values

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		got := strings.TrimSpace(outBuf.String())
		expected := "Hello, world!"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("greeting with flag and uppercase", func(t *testing.T) {
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

		cmd := NewHelloCmd(appContainer)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--name", "Alice", "--uppercase"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		got := strings.TrimSpace(outBuf.String())
		expected := "HELLO, ALICE!"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("greeting in JSON format", func(t *testing.T) {
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

		cmd := NewHelloCmd(appContainer)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--name", "Bob"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		var resp HelloResponse
		if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse JSON response: %v", err)
		}

		if resp.Greeting != "Hello, Bob!" {
			t.Errorf("expected greeting to be 'Hello, Bob!', got %q", resp.Greeting)
		}
		if resp.Name != "Bob" {
			t.Errorf("expected name to be 'Bob', got %q", resp.Name)
		}
	})
}
