// Package server manages llama-server child processes.
package server

import (
	"os"
	"testing"

	"go-llama.cpp/internal/config"
)

// TestToArgs_integration verifies that ToArgs produces the correct argument list.
func TestToArgs_integration(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected []string
	}{
		{
			name: "full config",
			cfg: &config.Config{
				Model:      "test.gguf",
				Host:       config.Ptr("http://localhost"),
				Port:       config.Ptr(8080),
				CtxSize:    config.Ptr(4096),
				NGPULayers: config.Ptr(35),
				Temp:       config.Ptr(0.7),
				Threads:    config.Ptr(8),
				FlashAttn:  config.Ptr(true),
				NoMmap:     config.Ptr(false),
				CacheTypeK: config.Ptr("fp16"),
				CacheTypeV: config.Ptr("fp16"),
				NParallel:  config.Ptr(2),
				Extra:      map[string]string{"verbose": "true"},
			},
			expected: []string{
				"--model", "test.gguf",
				"--host", "http://localhost",
				"--port", "8080",
				"--ctx-size", "4096",
				"--n-gpu-layers", "35",
				"--temp", "0.7",
				"--threads", "8",
				"--flash-attn",
				"--cache-type-k", "fp16",
				"--cache-type-v", "fp16",
				"--parallel", "2",
				"--verbose", "true",
			},
		},
		{
			name: "minimal config",
			cfg: &config.Config{
				Model: "test.gguf",
			},
			expected: []string{
				"--model", "test.gguf",
			},
		},
		{
			name: "partial config",
			cfg: &config.Config{
				Model:     "test.gguf",
				Port:      config.Ptr(9000),
				FlashAttn: config.Ptr(true),
			},
			expected: []string{
				"--model", "test.gguf",
				"--port", "9000",
				"--flash-attn",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.ToArgs(tt.cfg)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d arguments, got %d", len(tt.expected), len(result))
			}
			for i, exp := range tt.expected {
				if result[i] != exp {
					t.Errorf("argument %d: expected %q, got %q", i, exp, result[i])
				}
			}
		})
	}
}

// TestFindBinary_notInPath verifies that FindBinary returns an error when llama-server is not in PATH.
func TestFindBinary_notInPath(t *testing.T) {
	// Create a temp directory with no executables
	tmpDir, err := os.MkdirTemp("", "test-path-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	// Save original PATH
	oldPath := os.Getenv("PATH")

	// Set PATH to empty temp dir
	os.Setenv("PATH", tmpDir)

	// Try to find llama-server - should fail
	_, err = FindBinary()
	if err == nil {
		t.Error("expected error when llama-server not in PATH, got nil")
	}

	// Restore original PATH
	os.Setenv("PATH", oldPath)
}
