package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/config"
)

func TestSchemaCmd(t *testing.T) {
	t.Run("text tree format", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)

		appContainer := &app.App{
			Config: &config.Config{
				Environment: "test",
				Format:      "text",
			},
			Out:    outBuf,
			ErrOut: errBuf,
		}

		rootCmd := NewRootCmd(appContainer)
		cmd := NewSchemaCmd(appContainer, rootCmd)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		got := outBuf.String()
		
		// The output should contain command descriptions
		if !strings.Contains(got, "nept - nept is a scalable command-line tool") {
			t.Errorf("expected text schema to contain root command description, got:\n%s", got)
		}
		if !strings.Contains(got, "hello - Say hello to someone") {
			t.Errorf("expected text schema to contain hello command description, got:\n%s", got)
		}
		if !strings.Contains(got, "version - Print CLI build and version information") {
			t.Errorf("expected text schema to contain version command description, got:\n%s", got)
		}
	})

	t.Run("json format schema", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)

		appContainer := &app.App{
			Config: &config.Config{
				Environment: "test",
				Format:      "json",
			},
			Out:    outBuf,
			ErrOut: errBuf,
		}

		rootCmd := NewRootCmd(appContainer)
		cmd := NewSchemaCmd(appContainer, rootCmd)
		cmd.SetOut(outBuf)
		cmd.SetErr(errBuf)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		var schema CommandSchema
		if err := json.Unmarshal(outBuf.Bytes(), &schema); err != nil {
			t.Fatalf("failed to parse schema JSON output: %v, raw output: %q", err, outBuf.String())
		}

		if schema.Name != "nept" {
			t.Errorf("expected root name 'nept', got %q", schema.Name)
		}

		// Check subcommands are parsed
		foundHello := false
		foundVersion := false
		for _, sub := range schema.Subcommands {
			if sub.Name == "hello" {
				foundHello = true
				// Check for hello flags
				foundNameFlag := false
				for _, flag := range sub.Flags {
					if flag.Name == "name" {
						foundNameFlag = true
						if flag.Default != "world" {
							t.Errorf("expected default name 'world', got %q", flag.Default)
						}
					}
				}
				if !foundNameFlag {
					t.Error("expected to find name flag in hello subcommand")
				}
			}
			if sub.Name == "version" {
				foundVersion = true
			}
		}

		if !foundHello {
			t.Error("expected to find 'hello' subcommand in schema")
		}
		if !foundVersion {
			t.Error("expected to find 'version' subcommand in schema")
		}
	})
}
