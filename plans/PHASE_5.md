# Phase 5: Model Download Manager

## Goal
After this phase: `go-llama-cpp download` fetches GGUF files from Hugging Face.
`serve` can auto-download if a model isn't present locally.

## Prerequisites
- Phase 1 (homedir, config) complete. Other phases not strictly required but should be done.

---

## Step 1: Create `internal/model/model.go`

```go
// Package model handles GGUF model file discovery and downloading.
package model
```

### Types

```go
// Manager handles model file operations.
type Manager struct {
    ModelsDir string // path to the models directory
}

// DownloadProgress is sent through a channel during download.
type DownloadProgress struct {
    BytesDownloaded int64
    TotalBytes      int64 // -1 if unknown
    Percent         float64
    Done            bool
    Err             error
}
```

### `(m *Manager) Resolve(name string) (string, error)`

Given a model name (e.g., "Qwen3.5-9B-Q4_K_M"), return the full path to the .gguf file.

1. Validate name: no path separators, no `..`, no empty string.
   Use `filepath.Base(name) != name` as a quick check — if they differ, the name has path components.
2. Try these in order:
   a. `m.ModelsDir/<name>` (exact filename — user might include .gguf)
   b. `m.ModelsDir/<name>.gguf`
3. Return the first one that exists.
4. If neither exists, return `ErrModelNotFound`.

```go
var ErrModelNotFound = errors.New("model not found")
```

### `(m *Manager) List() ([]ModelInfo, error)`

```go
type ModelInfo struct {
    Name     string    // filename without .gguf
    FileName string    // full filename
    Path     string    // absolute path
    Size     int64     // file size in bytes
    ModTime  time.Time // last modified
}
```

1. ReadDir on ModelsDir.
2. Filter for files ending in `.gguf`.
3. Stat each, populate ModelInfo.
4. Return sorted by Name.

### `(m *Manager) Download(url string, fileName string, progress chan<- DownloadProgress) error`

This is the core downloader.

1. Validate fileName (same rules as Resolve).
2. Determine dest path: `m.ModelsDir/<fileName>`. If it doesn't end in .gguf, append it.
3. Check if file already exists. If so, return error: `"model %s already exists. Delete it first or use a different name."`.
4. Create a temp file in ModelsDir: `<fileName>.download.tmp`.
5. Send HTTP GET to url.
6. Check response status. If not 200, return error with status code and body preview.
7. Read `Content-Length` header for total size. If missing, TotalBytes = -1.
8. Copy response body to temp file using `io.TeeReader` or a counting writer.
9. Every 1MB downloaded (or every 500ms, whichever comes first), send a `DownloadProgress` to the channel.
10. When done:
    a. Close temp file.
    b. Rename temp file to final path (`os.Rename`).
    c. Send final `DownloadProgress{Done: true, Percent: 100}`.
    d. Close channel.
11. If error at any point:
    a. Remove temp file.
    b. Send `DownloadProgress{Err: err}`.
    c. Close channel.
    d. Return error.

### Counting Writer helper:

```go
// countingWriter wraps an io.Writer and counts bytes written.
type countingWriter struct {
    w     io.Writer
    count int64
}

func (cw *countingWriter) Write(p []byte) (int, error) {
    n, err := cw.w.Write(p)
    cw.count += int64(n)
    return n, err
}
```

### `ParseHuggingFaceURL(input string) (downloadURL string, fileName string, err error)`

Users will pass things like:
- Full URL: `https://huggingface.co/Qwen/Qwen3-8B-GGUF/resolve/main/qwen3-8b-q4_k_m.gguf`
- Short form: `Qwen/Qwen3-8B-GGUF/qwen3-8b-q4_k_m.gguf`

Parse these:
1. If input starts with `http://` or `https://`: validate it's a huggingface.co URL
   (or any URL — we should support arbitrary URLs too). Return as-is.
   Extract fileName from the last path segment.
2. If input matches `<org>/<repo>/<filename>`:
   Construct: `https://huggingface.co/<org>/<repo>/resolve/main/<filename>`
   Use filename as-is.
3. Otherwise: return error with example usage.

---

## Step 2: Tests in `internal/model/model_test.go`

### Resolve tests:
- `TestResolve_exactMatch`: create `t.TempDir()/mymodel.gguf`, resolve "mymodel.gguf" → found
- `TestResolve_withoutExtension`: create `mymodel.gguf`, resolve "mymodel" → found
- `TestResolve_notFound`: resolve "nonexistent" → ErrModelNotFound
- `TestResolve_pathTraversal`: resolve "../etc/passwd" → error (not ErrModelNotFound, a validation error)

### List tests:
- `TestList_empty`: empty dir → empty slice
- `TestList_multipleModels`: create 3 .gguf files and 1 .txt file → only 3 returned
- `TestList_sorted`: verify alphabetical order

### ParseHuggingFaceURL tests:
- `TestParseHF_fullURL`: full HF URL → returns same URL, extracted filename
- `TestParseHF_shortForm`: `Qwen/Qwen3-8B-GGUF/qwen3-8b-q4_k_m.gguf` → correct URL
- `TestParseHF_arbitraryURL`: `https://example.com/model.gguf` → returns as-is
- `TestParseHF_invalid`: `"just-a-name"` → error

### Download tests:
Use `net/http/httptest` to create a test server that serves a small fake file.

- `TestDownload_success`: serve a 1KB file, download it, verify file exists and content matches
- `TestDownload_progress`: same but read from progress channel, verify we get updates
- `TestDownload_alreadyExists`: pre-create file, attempt download → error
- `TestDownload_serverError`: test server returns 404 → error, no temp file left behind
- `TestDownload_interrupted`: test server closes connection mid-transfer →
  error, temp file cleaned up

Run: `go build ./... && go test ./internal/model/ -v`

---

## Step 3: Create `cmd/go-llama-cpp/cmd_download.go`

- Use: `download <source>`
- Short: `Download a GGUF model file`
- Args: `cobra.ExactArgs(1)` — the source (URL or short form)
- Flags:
  - `--name` (string, optional): override the saved filename
- RunE:
  1. Resolve home, ensure exists.
  2. Parse source with `model.ParseHuggingFaceURL(args[0])`.
  3. If `--name` is set, use that as fileName instead.
  4. Create `model.Manager{ModelsDir: layout.Models}`.
  5. Create progress channel: `make(chan model.DownloadProgress)`.
  6. Start download in a goroutine.
  7. In the main goroutine, read from the progress channel and print a progress bar:
     ```
     [INFO] Downloading qwen3-8b-q4_k_m.gguf...
     [=============================>          ] 73% (3.2 GB / 4.4 GB)
     ```
  8. When done: `ui.Info("Downloaded %s to %s", fileName, destPath)`.
  9. If error: clean error message.

### Progress bar rendering:

Create a helper function:
```go
func renderProgressBar(p model.DownloadProgress) string
```
- Bar width: 40 characters.
- If TotalBytes == -1: show only bytes downloaded (no percentage, no bar).
- Use `\r` to overwrite the line (print to os.Stdout directly, not through ui.Info).
- When Done, print a final newline.

---

## Step 4: Create `cmd/go-llama-cpp/cmd_models.go` — list models

- Use: `models`
- Short: `List downloaded models`
- Aliases: `[]string{"model-list"}`
- RunE:
  1. Resolve home.
  2. Create model.Manager.
  3. Call List().
  4. Print a table:
     ```
     NAME                     SIZE      MODIFIED
     Qwen3.5-9B-Q4_K_M       5.4 GB    2025-06-15
     Llama-3-8B-Q5_K_M        5.8 GB    2025-06-14
     ```
  5. Use `text/tabwriter` for alignment.
  6. Format size with a helper: `formatBytes(n int64) string` → "5.4 GB", "850 MB", etc.

Write a test for `formatBytes`:
- 0 → "0 B"
- 1023 → "1023 B"
- 1024 → "1.0 KB"
- 5_400_000_000 → "5.0 GB"

---

## Step 5: Wire auto-download into `serve` (optional prompt)

Edit `cmd/go-llama-cpp/cmd_serve.go`:

When the model file is not found (currently we just error), add:

```
[WARN] Model 'X' not found in /home/user/.go-llama.cpp/models/
[INFO] Would you like to search Hugging Face for it? (y/n) [n]
```

If yes:
- This is a stretch goal. For now, just print:
  ```
  [INFO] Download it with: go-llama-cpp download <org>/<repo>/<filename.gguf>
  ```
- Do NOT implement automatic HF search in this phase. Just provide the hint.

---

## Step 6: Verify

```bash
go build -o go-llama-cpp ./cmd/go-llama-cpp

# List models (should be empty or show existing)
./go-llama-cpp models

# Download a small test model (use a real small GGUF for testing)
# Example: a tiny test GGUF from HF
./go-llama-cpp download https://huggingface.co/ggml-org/tiny-llamas/resolve/main/stories260K.gguf

# Verify it shows up
./go-llama-cpp models
```

Run: `go build ./... && go vet ./... && go test ./... -v`

---

## Done Criteria
- [ ] `download <url>` fetches a file with progress bar
- [ ] `download <org>/<repo>/<file>` constructs the HF URL correctly
- [ ] `models` lists downloaded .gguf files with size and date
- [ ] Downloaded file lands in `~/.go-llama.cpp/models/` with correct name
- [ ] Interrupted downloads clean up temp files
- [ ] All tests pass including httptest-based download tests