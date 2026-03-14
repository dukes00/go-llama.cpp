# Phase 4: Interactive Preset Wizard + Override Logic

## Goal
After this phase: when a user runs `serve` without `--config` and with unique flags,
the tool prompts to save a preset. The `--config` + `--override` flow works with
a save prompt. Saved presets auto-load on future runs.

## Prerequisites
- Phase 1 (config store), Phase 2 (serve command), Phase 3 (registry) all complete.

---

## Step 1: Create `internal/wizard/wizard.go`

```go
// Package wizard implements the interactive preset save/load flow.
package wizard
```

### Types

```go
// Wizard handles interactive config preset prompts.
type Wizard struct {
    Store *config.Store
    UI    *ui.UI
}
```

### `(w *Wizard) PromptSavePreset(modelName string, cfg *config.Config) (saved bool, presetName string, err error)`

This is called when the user runs `serve` with flags but no `--config`.

1. Call `w.UI.Confirm("Configuration is new. Save as a preset?", true)`.
2. If user says no: return `false, "", nil`.
3. If yes:
   a. Call `w.UI.Prompt(fmt.Sprintf("Preset name [%s]: ", modelName))`.
   b. If input is empty, use `modelName` as the preset name.
   c. Validate the preset name (no special chars — reuse config store's validation
      by attempting to save; if it errors on the name, report it and re-prompt once).
   d. Check if preset already exists: `w.Store.Exists(name)`.
   e. If exists: `w.UI.Confirm("Preset already exists. Overwrite?", false)`.
      If no: return `false, "", nil`.
   f. Call `w.Store.Save(name, cfg)`.
   g. `w.UI.Info("Preset '%s' saved.", name)`
   h. Return `true, name, nil`.

### `(w *Wizard) PromptSaveOverride(originalPreset string, merged *config.Config) (saved bool, presetName string, err error)`

This is called when the user runs `serve --config X --override --temp 0.7`.

1. `w.UI.Confirm("Save this configuration as a new preset?", false)`.
2. If no: return `false, "", nil`.
3. If yes:
   a. Suggest name: `originalPreset + "-custom"`.
   b. Prompt for name with that default.
   c. Same exists/overwrite logic as above.
   d. Save and return.

### `(w *Wizard) AutoLoadPreset(modelName string) (*config.Config, bool)`

Called at the start of `serve` when NO `--config` flag is provided.

1. Check `w.Store.Exists(modelName)`.
2. If yes: load it. `w.UI.Info("Loading preset '%s'...", modelName)`. Return config, true.
3. If no: return nil, false.

---

## Step 2: Tests in `internal/wizard/wizard_test.go`

For all tests, create a `UI` with `bytes.Buffer` for In/Out/Err to simulate user input.

### Helper to build a test wizard:
```go
func testWizard(t *testing.T, input string) (*Wizard, *bytes.Buffer, *bytes.Buffer) {
    t.Helper()
    dir := t.TempDir()
    store := &config.Store{Dir: dir}
    out := &bytes.Buffer{}
    errOut := &bytes.Buffer{}
    u := &ui.UI{
        In:  strings.NewReader(input),
        Out: out,
        Err: errOut,
    }
    return &Wizard{Store: store, UI: u}, out, errOut
}
```

### Tests:

- `TestPromptSavePreset_yes_defaultName`:
  Input: `"y\n\n"` (yes to save, empty for name → uses model name).
  Verify: config file exists with model name, returns saved=true.

- `TestPromptSavePreset_yes_customName`:
  Input: `"y\nmy-preset\n"`.
  Verify: `my-preset.json` exists.

- `TestPromptSavePreset_no`:
  Input: `"n\n"`.
  Verify: no file created, returns saved=false.

- `TestPromptSavePreset_overwriteExisting`:
  Pre-create a preset. Input: `"y\nexisting-name\ny\n"` (yes save, name that exists, yes overwrite).
  Verify: file updated.

- `TestPromptSavePreset_declineOverwrite`:
  Pre-create a preset. Input: `"y\nexisting-name\nn\n"`.
  Verify: returns saved=false, original file unchanged.

- `TestAutoLoadPreset_exists`:
  Pre-create a preset named "mymodel". Call AutoLoadPreset("mymodel").
  Verify: returns config, true.

- `TestAutoLoadPreset_notExists`:
  Call AutoLoadPreset("nonexistent"). Verify: returns nil, false.

- `TestPromptSaveOverride_yes`:
  Input: `"y\n\n"`. Verify: saved with default "-custom" suffix.

- `TestPromptSaveOverride_no`:
  Input: `"n\n"`. Verify: nothing saved.

Run: `go build ./... && go test ./internal/wizard/ -v`

---

## Step 3: Wire wizard into `cmd/go-llama-cpp/cmd_serve.go`

Modify the serve RunE. The updated flow:

```
1. Parse modelName from args[0]
2. Resolve homedir, ensure exists
3. Validate model path exists

4. Create wizard: &wizard.Wizard{Store: configStore, UI: ui.Default}

5. IF --config is set:
   a. Load config from store
   b. IF cliOverridesExist(cmd):
      - IF --override is set:
        - Build overrides from flags
        - Merge into loaded config
        - Call wizard.PromptSaveOverride(configName, merged)
      - ELSE:
        - ui.Warn for each changed flag: "Ignoring --%s (using preset '%s'). Use --override to apply."
   c. Use the loaded (possibly merged) config

6. ELSE (no --config):
   a. Try wizard.AutoLoadPreset(modelName)
   b. IF preset found AND no CLI flags set:
      - Use the loaded config (auto-load path)
   c. IF preset found AND CLI flags ARE set:
      - ui.Info("Preset '%s' exists but CLI flags provided. Using CLI flags.", modelName)
      - Build config from flags
      - (Do NOT prompt to save — they're explicitly overriding)
   d. IF preset NOT found:
      - Build config from flags
      - IF at least one flag was set (not just bare `serve model`):
        - Call wizard.PromptSavePreset(modelName, cfg)

7. Start server with final config
8. Register in registry
9. Print success
```

This is the most complex step. Take it slow and implement it as a clearly-commented
sequence with early returns.

---

## Step 4: Integration test — manual verification script

Create `scripts/test_wizard.sh` (not for CI, just a helper):

```bash
#!/usr/bin/env bash
set -e

export GO_LLAMA_HOME=$(mktemp -d)
echo "Using temp home: $GO_LLAMA_HOME"

BIN="./go-llama-cpp"
go build -o "$BIN" ./cmd/go-llama-cpp

# Init
$BIN init

# Create a fake model file for testing
touch "$GO_LLAMA_HOME/models/test-model.gguf"

echo "--- Test: serve with flags should prompt to save ---"
echo -e "y\ntest-preset\n" | $BIN serve test-model --temp 0.6 --n-gpu-layers 99 --port 19999

echo "--- Checking preset was saved ---"
cat "$GO_LLAMA_HOME/configs/test-preset.json"

echo "--- Test: list should show the running server ---"
$BIN list

echo "--- Test: kill ---"
$BIN kill test-model

echo "--- Test: list should be empty ---"
$BIN list

echo "--- Cleanup ---"
rm -rf "$GO_LLAMA_HOME"
echo "DONE"
```

Mark it executable: `chmod +x scripts/test_wizard.sh`

NOTE: This script will only fully work if `llama-server` is in PATH.
Without it, the serve will fail at process start — that's fine, the wizard
prompts should still trigger before the start call. Restructure if needed:
move the wizard prompt BEFORE the server.Start() call so it works even in testing.

---

## Done Criteria
- [ ] `serve model --temp 0.6` (no preset exists) → prompts to save → saves JSON file
- [ ] `serve model` (preset exists) → auto-loads preset, no prompt
- [ ] `serve model --config X` → loads X, ignores other flags with warning
- [ ] `serve model --config X --override --temp 0.7` → merges, prompts to save as new preset
- [ ] All unit tests pass
- [ ] Wizard tests cover all user input paths (yes/no/default/overwrite)