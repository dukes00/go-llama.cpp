package homedir

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolve_default(t *testing.T) {
	// Save original env var
	orig := os.Getenv("GO_LLAMA_HOME")
	defer os.Setenv("GO_LLAMA_HOME", orig)

	// Unset the env var
	os.Unsetenv("GO_LLAMA_HOME")

	layout, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	// Verify Root ends with .go-llama.cpp
	if !strings.HasSuffix(layout.Root, ".go-llama.cpp") {
		t.Errorf("Root = %q, want suffix .go-llama.cpp", layout.Root)
	}

	// Verify sub-paths are correct
	expectedModels := filepath.Join(layout.Root, "models")
	if layout.Models != expectedModels {
		t.Errorf("Models = %q, want %q", layout.Models, expectedModels)
	}

	expectedConfigs := filepath.Join(layout.Root, "configs")
	if layout.Configs != expectedConfigs {
		t.Errorf("Configs = %q, want %q", layout.Configs, expectedConfigs)
	}

	expectedState := filepath.Join(layout.Root, "state")
	if layout.State != expectedState {
		t.Errorf("State = %q, want %q", layout.State, expectedState)
	}

	expectedLogs := filepath.Join(layout.Root, "logs")
	if layout.Logs != expectedLogs {
		t.Errorf("Logs = %q, want %q", layout.Logs, expectedLogs)
	}

	expectedBin := filepath.Join(layout.Root, "bin")
	if layout.Bin != expectedBin {
		t.Errorf("Bin = %q, want %q", layout.Bin, expectedBin)
	}
}

func TestResolve_envOverride(t *testing.T) {
	// Save original env var
	orig := os.Getenv("GO_LLAMA_HOME")
	defer os.Setenv("GO_LLAMA_HOME", orig)

	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "go-llama-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp() returned error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set the env var
	os.Setenv("GO_LLAMA_HOME", tmpDir)

	layout, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	// Verify Root matches the temp directory
	if layout.Root != tmpDir {
		t.Errorf("Root = %q, want %q", layout.Root, tmpDir)
	}
}

func TestEnsureExists(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "go-llama-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp() returned error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set the env var
	os.Setenv("GO_LLAMA_HOME", tmpDir)

	// Resolve the layout
	layout, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	// Ensure directories exist
	if err := layout.EnsureExists(); err != nil {
		t.Fatalf("EnsureExists() returned error: %v", err)
	}

	// Verify all directories exist
	for _, dir := range []string{layout.Models, layout.Configs, layout.State, layout.Logs, layout.Bin} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Directory %q does not exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Errorf("%q is not a directory", dir)
		}
	}
}

func TestValidate_nonexistent(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "go-llama-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp() returned error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set the env var to a non-existent path (tmpDir + "/nonexistent")
	os.Setenv("GO_LLAMA_HOME", filepath.Join(tmpDir, "nonexistent"))

	// Resolve the layout
	layout, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	// Validate should fail because the directory doesn't exist
	err = layout.Validate()
	if err == nil {
		t.Errorf("Validate() should return error for non-existent root")
	}

	var notFoundErr *ErrRootNotFound
	if !errors.As(err, &notFoundErr) {
		t.Errorf("Expected ErrRootNotFound, got %T: %v", err, err)
	}
}

func TestValidate_invalid(t *testing.T) {
	// Create a temp file (not a directory)
	tmpFile, err := os.CreateTemp("", "go-llama-test-*")
	if err != nil {
		t.Fatalf("CreateTemp() returned error: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Set the env var to the file path
	os.Setenv("GO_LLAMA_HOME", tmpFile.Name())

	// Resolve the layout
	layout, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	// Validate should fail because the path is not a directory
	err = layout.Validate()
	if err == nil {
		t.Errorf("Validate() should return error for non-directory root")
	}

	var invalidErr *ErrRootInvalid
	if !errors.As(err, &invalidErr) {
		t.Errorf("Expected ErrRootInvalid, got %T: %v", err, err)
	}
}
