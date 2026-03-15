// Package wizard implements the interactive preset save/load flow.
package wizard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-llama.cpp/internal/config"
	"go-llama.cpp/internal/ui"
)

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

func TestPromptSavePreset_yes_defaultName(t *testing.T) {
	w, _, _ := testWizard(t, "y\n\n")
	saved, name, err := w.PromptSavePreset("test-model", &config.Config{Model: "test-model"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !saved {
		t.Fatal("expected saved=true")
	}
	if name != "test-model" {
		t.Errorf("expected name='test-model', got '%s'", name)
	}

	// Verify file was created
	path := filepath.Join(w.Store.Dir, "test-model.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected config file to exist")
	}
}

func TestPromptSavePreset_yes_customName(t *testing.T) {
	w, _, _ := testWizard(t, "y\nmy-preset\n")
	saved, name, err := w.PromptSavePreset("test-model", &config.Config{Model: "test-model"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !saved {
		t.Fatal("expected saved=true")
	}
	if name != "my-preset" {
		t.Errorf("expected name='my-preset', got '%s'", name)
	}

	// Verify file was created
	path := filepath.Join(w.Store.Dir, "my-preset.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected config file to exist")
	}
}

func TestPromptSavePreset_no(t *testing.T) {
	w, _, _ := testWizard(t, "n\n")
	saved, name, err := w.PromptSavePreset("test-model", &config.Config{Model: "test-model"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved {
		t.Fatal("expected saved=false")
	}
	if name != "" {
		t.Errorf("expected empty name, got '%s'", name)
	}

	// Verify no file was created
	path := filepath.Join(w.Store.Dir, "test-model.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("expected no config file to exist")
	}
}

func TestPromptSavePreset_overwriteExisting(t *testing.T) {
	dir := t.TempDir()
	store := &config.Store{Dir: dir}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	u := &ui.UI{
		In:  strings.NewReader("y\nexisting-name\ny\n"),
		Out: out,
		Err: errOut,
	}

	// Pre-create a preset
	existingCfg := &config.Config{Model: "existing", Temp: config.Ptr(0.5)}
	if err := store.Save("existing-name", existingCfg); err != nil {
		t.Fatalf("failed to create existing preset: %v", err)
	}

	w := &Wizard{Store: store, UI: u}
	saved, name, err := w.PromptSavePreset("test-model", &config.Config{Model: "test-model", Temp: config.Ptr(0.6)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !saved {
		t.Fatal("expected saved=true")
	}
	if name != "existing-name" {
		t.Errorf("expected name='existing-name', got '%s'", name)
	}

	// Verify file was updated
	path := filepath.Join(dir, "existing-name.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !strings.Contains(string(data), `"temp":0.6`) {
		t.Fatal("expected updated config with new temp value")
	}
}

func TestPromptSavePreset_declineOverwrite(t *testing.T) {
	dir := t.TempDir()
	store := &config.Store{Dir: dir}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	u := &ui.UI{
		In:  strings.NewReader("y\nexisting-name\nn\n"),
		Out: out,
		Err: errOut,
	}

	// Pre-create a preset
	existingCfg := &config.Config{Model: "existing", Temp: config.Ptr(0.5)}
	if err := store.Save("existing-name", existingCfg); err != nil {
		t.Fatalf("failed to create existing preset: %v", err)
	}

	w := &Wizard{Store: store, UI: u}
	saved, name, err := w.PromptSavePreset("test-model", &config.Config{Model: "test-model", Temp: config.Ptr(0.6)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved {
		t.Fatal("expected saved=false")
	}
	if name != "" {
		t.Errorf("expected empty name, got '%s'", name)
	}

	// Verify original file unchanged
	path := filepath.Join(dir, "existing-name.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !strings.Contains(string(data), `"temp":0.5`) {
		t.Fatal("expected original config to be unchanged")
	}
}

func TestAutoLoadPreset_exists(t *testing.T) {
	dir := t.TempDir()
	store := &config.Store{Dir: dir}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	u := &ui.UI{
		Out: out,
		Err: errOut,
	}

	// Pre-create a preset
	existingCfg := &config.Config{Model: "mymodel", Temp: config.Ptr(0.5)}
	if err := store.Save("mymodel", existingCfg); err != nil {
		t.Fatalf("failed to create preset: %v", err)
	}

	w := &Wizard{Store: store, UI: u}
	cfg, found := w.AutoLoadPreset("mymodel")
	if !found {
		t.Fatal("expected preset to be found")
	}
	if cfg.Temp == nil || *cfg.Temp != 0.5 {
		t.Errorf("expected temp=0.5, got %v", cfg.Temp)
	}
}

func TestAutoLoadPreset_notExists(t *testing.T) {
	dir := t.TempDir()
	store := &config.Store{Dir: dir}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	u := &ui.UI{
		Out: out,
		Err: errOut,
	}

	w := &Wizard{Store: store, UI: u}
	cfg, found := w.AutoLoadPreset("nonexistent")
	if found {
		t.Fatal("expected preset not to be found")
	}
	if cfg != nil {
		t.Fatal("expected nil config")
	}
}

func TestPromptSaveOverride_yes(t *testing.T) {
	w, _, _ := testWizard(t, "y\n\n")
	saved, name, err := w.PromptSaveOverride("original-preset", &config.Config{Model: "test-model", Temp: config.Ptr(0.7)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !saved {
		t.Fatal("expected saved=true")
	}
	if name != "original-preset-custom" {
		t.Errorf("expected name='original-preset-custom', got '%s'", name)
	}

	// Verify file was created
	path := filepath.Join(w.Store.Dir, "original-preset-custom.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected config file to exist")
	}
}

func TestPromptSaveOverride_no(t *testing.T) {
	w, _, _ := testWizard(t, "n\n")
	saved, name, err := w.PromptSaveOverride("original-preset", &config.Config{Model: "test-model", Temp: config.Ptr(0.7)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved {
		t.Fatal("expected saved=false")
	}
	if name != "" {
		t.Errorf("expected empty name, got '%s'", name)
	}
}
