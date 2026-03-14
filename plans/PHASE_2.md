# Phase 2: CLI Skeleton + Serve Command

## Goal
After this phase: a working `go-llama-cpp` binary with `serve`, `init`, and stub
commands. `serve` actually launches `llama-server` in the background.

## Prerequisites
- Phase 1 complete. `internal/config`, `internal/homedir`, `internal/ui` exist and tests pass.

---

## Step 1: Create `cmd/go-llama-cpp/main.go` — Root cobra command

```go
package main 
```

Set up the root cobra command:

    Use: go-llama-cpp
    Short: A lightweight process manager for llama.cpp
    No Run function on root (it just prints help)
    Add a persistent flag: --home (string) to override GO_LLAMA_HOME
    In PersistentPreRunE: if --home is set, os.Setenv("GO_LLAMA_HOME", value)

Add subcommands (each in its own file in cmd/go-llama-cpp/):

## Step 2: Create cmd/go-llama-cpp/cmd_init.go — init command

    Use: init
    Short: Initialize the go-llama.cpp home directory
    RunE function:
        Call homedir.Resolve()
        Call layout.EnsureExists()
        Print ui.Info("Initialized home directory at %s", layout.Root)

Test manually: go run ./cmd/go-llama-cpp init

## Step 3: Create internal/server/server.go — Server orchestrator

```go
// Package server manages llama-server child processes.
package server
```

Types

```go
// Instance represents a running llama-server process.
type Instance struct {
    ModelName string
    Cmd       *exec.Cmd
    Port      int
    LogFile   string
    PID       int
}

// Options holds everything needed to start a server.
type Options struct {
    Config    *config.Config
    ModelPath string   // full path to .gguf file
    LogDir    string   // directory for log files
}
```

FindBinary() (string, error)

Use exec.LookPath("llama-server").
Return a clear error if not found: "llama-server not found in $PATH. Install llama.cpp first."
Start(opts Options) (*Instance, error)

    Call FindBinary() to get the binary path.
    Call config.ToArgs(opts.Config) to build the argument list.
    Prepend --model <opts.ModelPath> if opts.Config.Model is empty. (If Config already has Model set, use that.)
    Determine port: if Config.Port is set, use it. Otherwise default to 8080.
    Create log file at opts.LogDir/<modelname>_<port>.log (use os.Create).
    Build exec.Cmd:
        Set Cmd.Stdout and Cmd.Stderr both to the log file.
        Set Cmd.SysProcAttr to detach the process:
            On Linux/Mac: &syscall.SysProcAttr{Setsid: true}
            Put this behind a build tag. Create two files:
                internal/server/proc_unix.go with //go:build !windows
                internal/server/proc_windows.go with //go:build windows
                Each exports a function sysProcAttr() *syscall.SysProcAttr
                Unix returns &syscall.SysProcAttr{Setsid: true}
                Windows returns &syscall.SysProcAttr{CreationFlags: 0x00000008} (DETACHED_PROCESS)
    Call Cmd.Start() (NOT Cmd.Run() — we don't want to block).
    Return &Instance{ModelName: modelName, Cmd: cmd, Port: port, LogFile: logPath, PID: cmd.Process.Pid}.

Stop(pid int) error

    Find process by PID: os.FindProcess(pid).
    Send SIGTERM (Unix) or Process.Kill() (Windows). Use build tags again:
        func stopProcess(p *os.Process) error in proc_unix.go: p.Signal(syscall.SIGTERM)
        func stopProcess(p *os.Process) error in proc_windows.go: p.Kill()
    Return error if any.

Tests in internal/server/server_test.go:

You CANNOT test with a real llama-server in unit tests. Test what you can:

    TestToArgs_integration: Create a config, call ToArgs, verify the argument list is correct. (This is really a config test but validates the integration.)
    TestFindBinary_notInPath: Set PATH to empty temp dir, verify error.
    Do NOT write a test for Start/Stop — those require the actual binary. Add a comment: // Integration tests for Start/Stop require llama-server in PATH.

Run: go build ./... && go test ./... -v

## Step 4: Create cmd/go-llama-cpp/cmd_serve.go — serve command



    Use: serve <model>

    Short: Start a llama-server instance

    Args: cobra.ExactArgs(1) — the argument is the model name

    Flags (all optional):
        --config (string): name of a saved preset
        --override (bool): allow flag overrides when using --config
        --port (int): port to bind
        --ctx-size (int): context size
        --n-gpu-layers (int): GPU layers
        --temp (float64): temperature
        --threads (int): thread count
        --flash-attn (bool): enable flash attention
        --no-mmap (bool): disable mmap
        --cache-type-k (string): K cache type
        --cache-type-v (string): V cache type
        --parallel (int): parallel sequences

    RunE logic:
        modelName := args[0]
        Resolve home directory, ensure it exists.
        Construct model path: layout.Models + "/" + modelName + ".gguf"
            BUT: validate modelName has no path separators or ..
            If model file doesn't exist, print error and return.
        If --config flag is set: a. Load the config from the store. b. Check if ANY other llama-server flags were passed on CLI (check cmd.Flags().Changed("port") etc.) c. If flags were passed AND --override is NOT set:
            ui.Warn("Ignoring CLI flags because --config is set. Use --override to apply them.") d. If flags were passed AND --override IS set:
            Merge CLI flags into the loaded config (CLI wins).
            (Wizard prompt for saving will be added in Phase 4.)
        If --config flag is NOT set: a. Build a Config from CLI flags. b. (Wizard prompt for saving will be added in Phase 4.)
        Call server.Start(opts).
        Print: ui.Info("Server started for %s on port %d (PID %d)", modelName, instance.Port, instance.PID)
        Print: ui.Info("Logs: %s", instance.LogFile)

Helper: buildConfigFromFlags(cmd *cobra.Command, modelName string) *config.Config

Create a new Config. For each flag, check cmd.Flags().Changed("flag-name").
If changed, set the corresponding Config field. If not changed, leave it nil.

This helper is used in both the "no config" and "override" paths.
Helper: mergeConfigWithFlags(base *config.Config, overrides *config.Config) *config.Config

For each field in overrides: if non-nil, copy it to base. Return base.
Helper: cliOverridesExist(cmd *cobra.Command) bool

Return true if ANY of the llama-server-related flags were explicitly set by the user.
Check each with cmd.Flags().Changed(...).

## Step 5: Verify the binary works end-to-end (manual)

```bash
go build -o go-llama-cpp ./cmd/go-llama-cpp

# Initialize
./go-llama-cpp init

# Try serve (will fail if no model file, that's OK — test the error message)
./go-llama-cpp serve test-model
# Expected: [ERROR] Model file not found: /home/you/.go-llama.cpp/models/test-model.gguf

# If you have a real model and llama-server installed, test:
# cp /path/to/model.gguf ~/.go-llama.cpp/models/MyModel.gguf
# ./go-llama-cpp serve MyModel --n-gpu-layers 99 --ctx-size 4096
# Expected: [INFO] Server started for MyModel on port 8080 (PID XXXXX)
```

Run: go build ./... && go vet ./... && go test ./... -v

---

## Done Criteria
- [ ] `go build -o go-llama-cpp ./cmd/go-llama-cpp` produces a binary
- [ ] `./go-llama-cpp --help` shows serve, init commands
- [ ] `./go-llama-cpp init` creates `~/.go-llama.cpp/` structure
- [ ] `./go-llama-cpp serve nonexistent` gives a clear model-not-found error
- [ ] All existing tests still pass
