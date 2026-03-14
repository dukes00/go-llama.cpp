# Phase 1: Foundation — Types, Home Directory, Config Store

## Goal
After this phase, we have: the Go module initialized, core types defined,
home directory creation/validation, and full CRUD for JSON config presets with tests.

## Prerequisites
- None. This is the first phase.
- `go.mod` already exists with `module go-llama.cpp` and `go 1.25.6`.

---

## Step 1: Create `go.mod` and install cobra

Run these commands:
```bash
mkdir -p cmd/go-llama-cpp internal/config internal/homedir internal/ui
go mod init go-llama.cpp
go get github.com/spf13/cobra@latest
```

---
## Step 2: Create internal/ui/ui.go — User output helpers

This is a tiny package for consistent CLI output. NO external dependencies.

```go
// Package ui provides simple formatted output for CLI user interaction.
package ui
```

Define these functions:

    Info(format string, args ...any) — prints [INFO] <message> to stdout
    Warn(format string, args ...any) — prints [WARN] <message> to stderr
    Error(format string, args ...any) — prints [ERROR] <message> to stderr
    Prompt(question string) (string, error) — prints question to stdout, reads one line from stdin, returns trimmed input
    Confirm(question string, defaultYes bool) (bool, error) — prints question (y/n) [default], calls Prompt, returns bool

Use fmt.Fprintf with os.Stdout / os.Stderr. For Prompt, use bufio.NewReader(os.Stdin).ReadString('\n').

Write tests for Confirm parsing logic in internal/ui/ui_test.go:

    Test input "y" → true
    Test input "n" → false
    Test input "" with defaultYes=true → true
    Test input "" with defaultYes=false → false
    Test input "Y" → true (case insensitive)

To test Prompt/Confirm, accept an io.Reader and io.Writer in the test by making the
functions methods on a UI struct:

```go
type UI struct {
    In  io.Reader
    Out io.Writer
    Err io.Writer
}
```

Provide a package-level Default variable initialized with os.Stdin/Stdout/Stderr.
The package-level functions (Info, Warn, etc.) delegate to Default.

Run: go build ./... && go test ./internal/ui/ -v
## Step 3: Create internal/homedir/homedir.go — Home directory management

```go
// Package homedir manages the ~/.go-llama.cpp directory structure.
package homedir
```

Define a struct:

```go
// Layout holds the resolved paths for all standard directories.
type Layout struct {
    Root    string // e.g., /home/user/.go-llama.cpp
    Models  string // Root/models
    Configs string // Root/configs
    State   string // Root/state
    Logs    string // Root/logs
    Bin     string // Root/bin
}
```

Functions to implement:
Resolve() (Layout, error)

    Check GO_LLAMA_HOME env var. If set, use that as Root.
    Otherwise, use os.UserHomeDir() + /.go-llama.cpp.
    Populate all sub-paths by joining Root with the dir names.
    Return Layout. Do NOT create directories here.

(l Layout) EnsureExists() error

    For each directory in the layout (Models, Configs, State, Logs, Bin): call os.MkdirAll(dir, 0755).
    Return first error encountered.

(l Layout) Validate() error

    Check that Root exists and is a directory.
    Return a descriptive error if not.

Write tests in internal/homedir/homedir_test.go:

    TestResolve_default: unset GO_LLAMA_HOME, verify Root ends with .go-llama.cpp
    TestResolve_envOverride: set GO_LLAMA_HOME to t.TempDir(), verify Root matches
    TestEnsureExists: resolve to t.TempDir(), call EnsureExists, verify all dirs exist
    TestValidate_nonexistent: verify error when root doesn't exist

Run: go build ./... && go test ./internal/homedir/ -v
## Step 4: Create internal/config/config.go — Config types and store

```go
// Package config manages JSON configuration presets for llama-server.
package config
```

Config struct

This struct maps to llama-server CLI flags. Use pointer types so we can distinguish
"not set" from "zero value". The JSON tags MUST match llama-server flag names.

```go
// Config represents a llama-server configuration preset.
type Config struct {
    // Model is the path or name of the GGUF model file.
    Model string `json:"model"`

    // Host is the address to bind to (default: 127.0.0.1).
    Host *string `json:"host,omitempty"`

    // Port is the port to listen on (default: 8080).
    Port *int `json:"port,omitempty"`

    // CtxSize is the context size in tokens (--ctx-size / -c).
    CtxSize *int `json:"ctx_size,omitempty"`

    // NGPULayers is the number of layers to offload to GPU (--n-gpu-layers / -ngl).
    NGPULayers *int `json:"n_gpu_layers,omitempty"`

    // Temp is the sampling temperature (--temp).
    Temp *float64 `json:"temp,omitempty"`

    // Threads is the number of threads (--threads / -t).
    Threads *int `json:"threads,omitempty"`

    // FlashAttn enables flash attention (--flash-attn / -fa).
    FlashAttn *bool `json:"flash_attn,omitempty"`

    // NoMmap disables memory mapping (--no-mmap).
    NoMmap *bool `json:"no_mmap,omitempty"`

    // CacheTypeK is the KV cache type for keys (--cache-type-k).
    CacheTypeK *string `json:"cache_type_k,omitempty"`

    // CacheTypeV is the KV cache type for values (--cache-type-v).
    CacheTypeV *string `json:"cache_type_v,omitempty"`

    // ContBatching enables continuous batching (--cont-batching).
    ContBatching *bool `json:"cont_batching,omitempty"`

    // NParallel is the number of parallel sequences (--parallel / -np).
    NParallel *int `json:"n_parallel,omitempty"`

    // Extra holds any additional flags not covered above.
    // Keys are flag names (without --), values are string representations.
    Extra map[string]string `json:"extra,omitempty"`
}
```
Helper functions for creating pointer values:

```go
func Ptr[T any](v T) *T { return &v }
```

ToArgs(c *Config) []string

Convert a Config to a slice of llama-server CLI arguments.
Rules:

    Skip nil pointer fields (they are "not set").
    Model → --model <value>
    Host → --host <value>
    Port → --port <value>
    CtxSize → --ctx-size <value>
    NGPULayers → --n-gpu-layers <value>
    Temp → --temp <value>
    Threads → --threads <value>
    FlashAttn → if true: --flash-attn, if false: omit
    NoMmap → if true: --no-mmap, if false: omit
    CacheTypeK → --cache-type-k <value>
    CacheTypeV → --cache-type-v <value>
    ContBatching → if true: --cont-batching, if false: omit
    NParallel → --parallel <value>
    For Extra: iterate sorted keys, emit --<key> <value> for each

Use strconv.Itoa for ints, strconv.FormatFloat(v, 'f', -1, 64) for floats.
Store struct

```go
// Store manages reading and writing config files from a directory.
type Store struct {
    Dir string // path to the configs directory
}
```

Methods:
(s *Store) Save(name string, cfg *Config) error

    Validate name: must be non-empty, must not contain /, \, or ...
    Marshal cfg to JSON with json.MarshalIndent(cfg, "", " ").
    Write to s.Dir/<name>.json using os.WriteFile with perm 0644.

(s *Store) Load(name string) (*Config, error)

    Validate name (same rules).
    Read s.Dir/<name>.json.
    Unmarshal into *Config.
    Return it.

(s *Store) Exists(name string) bool

    Stat the file. Return true if it exists.

(s *Store) List() ([]string, error)

    ReadDir, filter for .json files, return names without extension.

(s *Store) Delete(name string) error

    Validate name, remove file.

Define a sentinel error:

```go
var ErrNotFound = errors.New("config not found")
```

Load should return ErrNotFound when the file doesn't exist (wrap with %w).
Tests in internal/config/config_test.go:

    TestToArgs_full: create a Config with all fields set, verify args list
    TestToArgs_partial: Config with only Model and Temp, verify only those args
    TestToArgs_booleans: FlashAttn=true emits --flash-attn, FlashAttn=false omits it
    TestStore_SaveAndLoad: save a config, load it, compare fields
    TestStore_Load_notFound: load nonexistent, check errors.Is(err, ErrNotFound)
    TestStore_List: save 3 configs, list, verify all 3 names
    TestStore_Delete: save then delete, verify Exists returns false
    TestStore_Save_invalidName: names with ".." or "/" should error

All tests use t.TempDir() as the store directory.

Run: go build ./... && go test ./internal/config/ -v
## Step 5: Verify everything compiles together

Run:

bash
go build ./...
go vet ./...
go test ./... -v

All must pass with zero errors. If there are unused imports or variables, fix them.

## Done Criteria
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` clean
- [ ] `go test ./... -v` — all tests pass
- [ ] Files exist: `internal/ui/ui.go`, `internal/homedir/homedir.go`, `internal/config/config.go`
- [ ] Each has a corresponding `_test.go`
