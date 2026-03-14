# Phase 3: Process Registry + List/Kill Commands

## Goal
After this phase: running instances are tracked in a JSON file. `list` shows
them with live PID status. `kill` stops a server by model name. Stale entries
are cleaned automatically.

## Prerequisites
- Phase 2 complete. `serve` command launches llama-server and returns a PID.

---

## Step 1: Create `internal/registry/registry.go`

```go
// Package registry tracks running llama-server instances via a JSON file.
package registry
```
### Types

```go
// Entry represents a tracked server instance.
type Entry struct {
    ModelName string `json:"model_name"`
    PID       int    `json:"pid"`
    Port      int    `json:"port"`
    LogFile   string `json:"log_file"`
    StartedAt string `json:"started_at"` // RFC3339 timestamp
}

// Registry manages the state file.
type Registry struct {
    path string // full path to running.json
    mu   sync.Mutex
}
```

### `New(statePath string) *Registry`
- `path` = `statePath + "/running.json"`

### `(r *Registry) Add(e Entry) error`
1. Lock the mutex.
2. Load current entries from file (call internal `load()` method).
3. Check if an entry with the same ModelName AND Port already exists.
   If yes, return error: `"server for %s already running on port %d (PID %d)"`
4. Append the new entry.
5. Write back to file (call internal `save()` method).

### `(r *Registry) Remove(modelName string, port int) error`
1. Lock mutex.
2. Load entries.
3. Filter out the matching entry.
4. Save.
5. If no entry was removed, return `ErrNotFound`.

### `(r *Registry) RemoveByPID(pid int) error`
Same as Remove but matches by PID instead.

### `(r *Registry) List() ([]Entry, error)`
1. Lock mutex.
2. Load entries.
3. For each entry, check if PID is still alive (call `isAlive(pid int) bool`).
4. Remove dead entries from the list (auto-cleanup).
5. Save the cleaned list.
6. Return it.

### `(r *Registry) FindByModel(modelName string) ([]Entry, error)`
1. Call List() (which auto-cleans).
2. Filter entries by ModelName.

### Internal helpers

#### `(r *Registry) load() ([]Entry, error)`
- Read file. If `os.IsNotExist`, return empty slice (not an error).
- Unmarshal JSON.

#### `(r *Registry) save(entries []Entry) error`
- Marshal with indent.
- Write to file with 0644 perms.

#### `isAlive(pid int) bool`
- Use `os.FindProcess(pid)` then send signal 0 (Unix) to check.
- Build-tagged implementations:
  - Unix (`proc_unix.go` in registry package): `process.Signal(syscall.Signal(0))` — nil error means alive.
  - Windows (`proc_windows.go`): Use `os.FindProcess(pid)` — on Windows this always succeeds,
    so instead open a handle with `golang.org/x/sys/windows` or just assume alive
    (document the limitation). Keep it simple: just return true and add a TODO comment.

### Tests in `internal/registry/registry_test.go`:

- `TestRegistry_AddAndList`: add 2 entries, list, verify both present
- `TestRegistry_Add_duplicate`: add same model+port twice, expect error
- `TestRegistry_Remove`: add then remove, verify List returns empty
- `TestRegistry_RemoveByPID`: add entry, remove by PID, verify gone
- `TestRegistry_Remove_notFound`: remove nonexistent, check `errors.Is(err, ErrNotFound)`
- `TestRegistry_List_cleansDead`: add entry with a PID that doesn't exist
  (use a very high PID like 9999999), call List, verify it's auto-removed
- `TestRegistry_FindByModel`: add 2 entries for different models, find by name, verify correct one
- `TestRegistry_persistence`: create registry, add entry, create NEW registry instance
  pointing at same file, call List, verify entry survived

All tests use `t.TempDir()` for the state directory.

Define a sentinel error:
```go
var ErrNotFound = errors.New("registry entry not found")
var ErrAlreadyRunning = errors.New("server already running")
```

Run: `go build ./... && go test ./internal/registry/ -v`

---

## Step 2: Wire registry into `serve` command

Edit `cmd/go-llama-cpp/cmd_serve.go`:

After `server.Start(opts)` succeeds:
1. Create a `registry.New(layout.State)`.
2. Create an `Entry` with ModelName, PID, Port, LogFile, StartedAt = `time.Now().Format(time.RFC3339)`.
3. Call `registry.Add(entry)`.
4. If Add returns `ErrAlreadyRunning`, call `server.Stop(instance.PID)` to clean up the
   process we just started, then return the error to the user.

Run: `go build ./... && go vet ./...`

---

## Step 3: Create `cmd/go-llama-cpp/cmd_list.go` — list command

- Use: `list`
- Short: `Show running llama-server instances`
- Aliases: `[]string{"ls", "ps"}`
- RunE:
  1. Resolve home directory.
  2. Create registry.
  3. Call `registry.List()`.
  4. If empty: `ui.Info("No running servers.")`
  5. Otherwise: print a formatted table to stdout.

Table format:
```
MODEL                PORT   STATUS    PID     STARTED
Qwen3.5-9B-Q4_K_M   8080   Running   12345   2025-06-15T10:30:00Z
Llama-3-8B-Q5_K_M   8081   Running   12346   2025-06-15T11:00:00Z
```

Use `fmt.Fprintf` with `text/tabwriter` from stdlib for aligned columns.
Status is always "Running" because List() already cleaned dead entries.

---

## Step 4: Create `cmd/go-llama-cpp/cmd_kill.go` — kill command

- Use: `kill <model>`
- Short: `Stop a running llama-server instance`
- Args: `cobra.ExactArgs(1)`
- Flags:
  - `--port` (int, optional): if the same model runs on multiple ports, specify which one
  - `--all` (bool): kill ALL instances of this model
- RunE:
  1. Resolve home, create registry.
  2. `modelName := args[0]`
  3. Call `registry.FindByModel(modelName)`.
  4. If no entries found: `ui.Error("No running server found for %s", modelName)` and return error.
  5. If multiple entries and neither `--port` nor `--all` is set:
     - Print all matching entries as a table.
     - `ui.Error("Multiple instances found. Use --port <port> or --all to specify.")`
     - Return error.
  6. If `--all` is set: iterate all matching entries, stop each.
  7. Otherwise: find the one matching `--port` (or the only one).
  8. For each entry to kill:
     a. `server.Stop(entry.PID)`
     b. `registry.Remove(entry.ModelName, entry.Port)`
     c. `ui.Info("Stopped server for %s on port %d (PID %d)", entry.ModelName, entry.Port, entry.PID)`

---

## Step 5: Create `cmd/go-llama-cpp/cmd_logs.go` — bonus: logs command

This is small and very useful for a backgrounded process. Add it now.

- Use: `logs <model>`
- Short: `Show logs for a running server`
- Flags:
  - `--port` (int, optional): same as kill
  - `--follow` / `-f` (bool): tail the log file
  - `--lines` / `-n` (int, default 50): number of lines to show
- RunE:
  1. Resolve home, create registry.
  2. Find entry by model (same disambiguation as kill).
  3. If `--follow`: use `exec.Command("tail", "-f", entry.LogFile)` on Unix.
     Pipe stdout/stderr to os.Stdout/os.Stderr. Call `cmd.Run()` (blocks until ctrl-c).
     On Windows: fall back to polling the file every 500ms in a goroutine (or just
     print a warning that --follow isn't supported on Windows and print the last N lines).
  4. If not follow: read the log file, print last N lines.

### Helper: `lastNLines(filePath string, n int) (string, error)`
- Read the file, split by `\n`, return the last `n` lines joined.
- If file has fewer than `n` lines, return everything.

Write a test for `lastNLines` in `cmd/go-llama-cpp/logs_test.go`:
- Write a temp file with 100 lines, call lastNLines(path, 10), verify 10 lines.
- File with 3 lines, lastNLines(path, 10), verify 3 lines.

Run: `go build ./... && go vet ./... && go test ./... -v`

---

## Done Criteria
- [ ] `./go-llama-cpp list` works (shows empty table or running servers)
- [ ] `./go-llama-cpp kill <model>` stops a server and removes it from registry
- [ ] `./go-llama-cpp logs <model>` shows log output
- [ ] Registry auto-cleans dead PIDs on every `list` call
- [ ] All tests pass
- [ ] If you start a server with `serve`, it appears in `list`, and `kill` stops it

