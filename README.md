# go-llama-cpp

A lightweight process manager for [llama.cpp](https://github.com/ggerganov/llama.cpp). Think of it as `podman` for `llama-server` — start, stop, and monitor model servers from the command line without babysitting processes manually.

> Vibe coded project: ~80% [OmniCoder-9B](https://huggingface.co/OmniCoder) @ Cline, ~20% Claude Code.

---

## What it does

- Download GGUF model files from Hugging Face (or any URL)
- Start `llama-server` instances with full flag coverage
- Track running servers by model name (persisted registry)
- List, tail logs, and kill servers by name
- Save and reload server presets (configs)
- Interactive preset wizard on first `serve`

## Install

```bash
go install go-llama.cpp/cmd/go-llama-cpp@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/go-llama-cpp
cd go-llama-cpp
go build -o go-llama-cpp ./cmd/go-llama-cpp
```

Requires `llama-server` somewhere on your `$PATH` (or set via `GO_LLAMA_HOME`).

## Quick start

```bash
# Initialize home directory (~/.go-llama.cpp)
go-llama-cpp init

# Download a model
go-llama-cpp download Qwen/Qwen3-8B-GGUF/qwen3-8b-q4_k_m.gguf
# or a full URL:
go-llama-cpp download https://huggingface.co/ggml-org/tiny-llamas/resolve/main/stories260K.gguf

# List downloaded models
go-llama-cpp models

# Start a server
go-llama-cpp serve qwen3-8b-q4_k_m --port 8080 --n-gpu-layers 35

# List running servers
go-llama-cpp list

# Tail logs
go-llama-cpp logs qwen3-8b-q4_k_m

# Stop a server
go-llama-cpp kill qwen3-8b-q4_k_m
```

## Commands

| Command | Description |
|---|---|
| `init` | Create `~/.go-llama.cpp/` directory structure |
| `serve <model>` | Start a `llama-server` instance |
| `list` | Show running server instances |
| `kill <model>` | Stop a running server |
| `logs <model>` | Tail server logs |
| `download <source>` | Download a GGUF file |
| `models` | List downloaded models |

### `serve` flags

Covers the full `llama-server` flag set: context size, GPU layers, sampling parameters, LoRA adapters, speculative decoding, RoPE settings, cache types, and more. Run `go-llama-cpp serve --help` for the full list.

Presets can be saved and reloaded:

```bash
# Save current flags as a preset
go-llama-cpp serve qwen3-8b --n-gpu-layers 35 --ctx-size 8192
# (prompted to save as "qwen3" on first run)

# Reload preset
go-llama-cpp serve qwen3-8b --config qwen3
```

### `download` flags

| Flag | Description |
|---|---|
| `--name` | Override the saved filename |

Source can be a full URL or short `org/repo/filename` form (resolved to `https://huggingface.co/<org>/<repo>/resolve/main/<filename>`).

## Home directory

Default: `~/.go-llama.cpp/`
Override: `GO_LLAMA_HOME` env var or `--home` flag.

```
~/.go-llama.cpp/
  models/    # downloaded GGUF files
  configs/   # saved presets (JSON)
  logs/      # per-server log files
  state/     # process registry
  bin/       # optional: place llama-server here
```

## License

MIT
