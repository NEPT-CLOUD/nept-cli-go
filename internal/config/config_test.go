package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test Default Configuration
	t.Run("default configurations", func(t *testing.T) {
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}
		if cfg.Environment != "production" {
			t.Errorf("expected Environment to be 'production', got %q", cfg.Environment)
		}
		if cfg.Format != "text" {
			t.Errorf("expected Format to be 'text', got %q", cfg.Format)
		}
		if cfg.APIKey != "" {
			t.Errorf("expected APIKey to be empty, got %q", cfg.APIKey)
		}
		if cfg.Verbose {
			t.Errorf("expected Verbose to be false, got true")
		}
	})

	// Test Environment overrides
	t.Run("env variables override", func(t *testing.T) {
		t.Setenv("NEPT_API_KEY", "env_secret_key")
		t.Setenv("NEPT_ENVIRONMENT", "staging")
		t.Setenv("NEPT_VERBOSE", "true")
		t.Setenv("NEPT_FORMAT", "json")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}
		if cfg.APIKey != "env_secret_key" {
			t.Errorf("expected APIKey to be 'env_secret_key', got %q", cfg.APIKey)
		}
		if cfg.Environment != "staging" {
			t.Errorf("expected Environment to be 'staging', got %q", cfg.Environment)
		}
		if !cfg.Verbose {
			t.Errorf("expected Verbose to be true, got false")
		}
		if cfg.Format != "json" {
			t.Errorf("expected Format to be 'json', got %q", cfg.Format)
		}
	})

	// Test Config File load
	t.Run("config file load", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.yaml")

		content := `
environment: development
api_key: file_secret_key
verbose: true
format: json
`
		if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create temp config file: %v", err)
		}

		cfg, err := Load(configFile)
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}
		if cfg.APIKey != "file_secret_key" {
			t.Errorf("expected APIKey to be 'file_secret_key', got %q", cfg.APIKey)
		}
		if cfg.Environment != "development" {
			t.Errorf("expected Environment to be 'development', got %q", cfg.Environment)
		}
		if !cfg.Verbose {
			t.Errorf("expected Verbose to be true, got false")
		}
		if cfg.Format != "json" {
			t.Errorf("expected Format to be 'json', got %q", cfg.Format)
		}
	})
}

func TestSaveDefault(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".nept.yaml")

	err := SaveDefault(configFile)
	if err != nil {
		t.Fatalf("unexpected error saving default config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatalf("expected config file to be created at %s, but it wasn't", configFile)
	}

	// Load and verify saved config
	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("failed to load saved configuration: %v", err)
	}

	if cfg.APIKey != "your_api_key_here" {
		t.Errorf("expected APIKey to be 'your_api_key_here', got %q", cfg.APIKey)
	}
}
