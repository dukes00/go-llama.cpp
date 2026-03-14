# go-llama.cpp Project Rules

## What This Project Is
A CLI wrapper around `llama-server` (from llama.cpp). Think of it as "podman for llama.cpp".
It manages model downloads, config presets, and server process lifecycle.
It does NOT do inference itself — it spawns `llama-server` as a child process.

## Architecture
- Module: `go-llama.cpp`
- Go version: 1.25.6
- Entry point: `cmd/go-llama-cpp/main.go`
- Internal packages live under `internal/`
- CLI framework: `github.com/spf13/cobra`
- NO other external dependencies unless a plan file explicitly says so

## Package Layout
cmd/go-llama-cpp/main.go     # cobra root command setup
internal/config/              # JSON config preset CRUD
internal/homedir/             # ~/.go-llama.cpp directory management
internal/server/              # spawning and managing llama-server
internal/registry/            # PID tracking, list/kill
internal/model/               # download manager
internal/wizard/              # interactive preset wizard

## Key Behaviors
- `llama-server` binary is assumed to be in $PATH
- Default home directory: `~/.go-llama.cpp`
- Can be overridden with `GO_LLAMA_HOME` env var
- Models stored in `$GO_LLAMA_HOME/models/`
- Configs stored in `$GO_LLAMA_HOME/configs/` as JSON
- Runtime state in `$GO_LLAMA_HOME/state/running.json`
- Logs in `$GO_LLAMA_HOME/logs/`
- Multiple server instances can run simultaneously on different ports

## IMPORTANT Constraints
- NEVER shell out to `curl` or `wget`. Use `net/http` for downloads.
- NEVER use `os.Exit()` outside of `main.go`.
- ALWAYS return errors up the call stack. Use `fmt.Errorf("context: %w", err)`.
- ALWAYS validate file paths against path traversal (no `..` in model names).
- The `serve` command MUST background the process (like `docker compose up -d`).
- Config flag `--config mypreset --temp 0.7` WITHOUT `--override`: warn and ignore `--temp`.
- Config flag `--config mypreset --override --temp 0.7`: apply override, prompt to save.