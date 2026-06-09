package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFramework(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "nept-fw-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case: Go mod file
	goDir := filepath.Join(tempDir, "go-project")
	_ = os.Mkdir(goDir, 0755)
	_ = os.WriteFile(filepath.Join(goDir, "go.mod"), []byte("module test"), 0644)
	fw := DetectFramework(goDir)
	if fw.Framework != "Go" {
		t.Errorf("expected Go, got %s", fw.Framework)
	}

	// Test case: Package.json with Next dependency
	nodeDir := filepath.Join(tempDir, "node-project")
	_ = os.Mkdir(nodeDir, 0755)
	pkgJSON := `{
		"dependencies": {
			"next": "^14.0.0"
		}
	}`
	_ = os.WriteFile(filepath.Join(nodeDir, "package.json"), []byte(pkgJSON), 0644)
	fw = DetectFramework(nodeDir)
	if fw.Framework != "Next.js" {
		t.Errorf("expected Next.js, got %s", fw.Framework)
	}

	// Test case: Cargo.toml
	rustDir := filepath.Join(tempDir, "rust-project")
	_ = os.Mkdir(rustDir, 0755)
	_ = os.WriteFile(filepath.Join(rustDir, "Cargo.toml"), []byte("[package]"), 0644)
	fw = DetectFramework(rustDir)
	if fw.Framework != "Rust" {
		t.Errorf("expected Rust, got %s", fw.Framework)
	}
}
