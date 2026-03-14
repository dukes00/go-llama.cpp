// Package homedir manages the ~/.go-llama.cpp directory structure.
package homedir

import (
	"os"
	"path/filepath"
)

// Layout holds the resolved paths for all standard directories.
type Layout struct {
	Root    string // e.g., /home/user/.go-llama.cpp
	Models  string // Root/models
	Configs string // Root/configs
	State   string // Root/state
	Logs    string // Root/logs
	Bin     string // Root/bin
}

// Resolve returns the Layout for the go-llama.cpp home directory.
// It checks GO_LLAMA_HOME env var first, then falls back to ~/.go-llama.cpp.
func Resolve() (Layout, error) {
	root := os.Getenv("GO_LLAMA_HOME")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Layout{}, err
		}
		root = filepath.Join(home, ".go-llama.cpp")
	}

	return Layout{
		Root:    root,
		Models:  filepath.Join(root, "models"),
		Configs: filepath.Join(root, "configs"),
		State:   filepath.Join(root, "state"),
		Logs:    filepath.Join(root, "logs"),
		Bin:     filepath.Join(root, "bin"),
	}, nil
}

// EnsureExists creates all directories in the layout if they don't exist.
func (l *Layout) EnsureExists() error {
	for _, dir := range []string{l.Models, l.Configs, l.State, l.Logs, l.Bin} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// Validate checks that the Root exists and is a directory.
func (l *Layout) Validate() error {
	info, err := os.Stat(l.Root)
	if err != nil {
		if os.IsNotExist(err) {
			return &ErrRootNotFound{Root: l.Root}
		}
		return &ErrRootInvalid{Root: l.Root, Err: err}
	}
	if !info.IsDir() {
		return &ErrRootInvalid{Root: l.Root, Err: os.ErrInvalid}
	}
	return nil
}

// ErrRootNotFound is returned when the root directory doesn't exist.
type ErrRootNotFound struct {
	Root string
}

func (e *ErrRootNotFound) Error() string {
	return "root directory not found: " + e.Root
}

// ErrRootInvalid is returned when the root exists but is not a directory.
type ErrRootInvalid struct {
	Root string
	Err  error
}

func (e *ErrRootInvalid) Error() string {
	return "root directory invalid: " + e.Root + ": " + e.Err.Error()
}
