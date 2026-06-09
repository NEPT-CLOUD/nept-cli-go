package utils

import (
	"os"
	"testing"
)

func TestGetNodeEnv(t *testing.T) {
	// Backup existing env
	original := os.Getenv("NODE_ENV")
	defer os.Setenv("NODE_ENV", original)

	// Test default when NODE_ENV is unset/empty
	os.Setenv("NODE_ENV", "")
	if got := GetNodeEnv(); got != "pro" {
		t.Errorf("GetNodeEnv() = %q, want %q", got, "pro")
	}

	// Test when NODE_ENV is set to "production"
	os.Setenv("NODE_ENV", "production")
	if got := GetNodeEnv(); got != "production" {
		t.Errorf("GetNodeEnv() = %q, want %q", got, "production")
	}
}
