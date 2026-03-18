package config

import (
	"errors"
	"strings"
	"testing"
)

func TestToArgs_full(t *testing.T) {
	host := "0.0.0.0"
	port := 8080
	ctxSize := 4096
	ngl := 35
	temp := 0.7
	threads := 8
	flashAttn := "on"
	noMmap := false
	contBatching := true
	nParallel := 4
	cacheTypeK := "fp16"
	cacheTypeV := "fp16"

	cfg := &Config{
		Model:        "models/qwen2.5-7b-instruct-q4_0.gguf",
		Host:         &host,
		Port:         &port,
		CtxSize:      &ctxSize,
		NGPULayers:   &ngl,
		Temp:         &temp,
		Threads:      &threads,
		FlashAttn:    &flashAttn,
		NoMmap:       &noMmap,
		ContBatching: &contBatching,
		NParallel:    &nParallel,
		CacheTypeK:   &cacheTypeK,
		CacheTypeV:   &cacheTypeV,
		Extra:        map[string]string{"extra-flag": "extra-value"},
	}

	args := ToArgs(cfg)

	// Check that all expected args are present
	// Order matches ToArgs implementation: Model, Host, Port, CtxSize, NGPULayers, Temp, Threads,
	// FlashAttn, NoMmap (only if true), ContBatching, CacheTypeK, CacheTypeV, NParallel, then Extra
	// Note: noMmap is false, so --no-mmap is NOT included
	expectedArgs := []string{
		"--model", "models/qwen2.5-7b-instruct-q4_0.gguf",
		"--host", "0.0.0.0",
		"--port", "8080",
		"--ctx-size", "4096",
		"--n-gpu-layers", "35",
		"--temp", "0.7",
		"--threads", "8",
		"--flash-attn", "on",
		"--cont-batching",
		"--cache-type-k", "fp16",
		"--cache-type-v", "fp16",
		"--parallel", "4",
		"--extra-flag", "extra-value",
	}

	if len(args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d: %v", len(expectedArgs), len(args), args)
	}

	// Compare slices directly - ToArgs returns args in deterministic order
	for i := 0; i < len(expectedArgs); i += 2 {
		if args[i] != expectedArgs[i] {
			t.Errorf("Expected arg %d to be %q, got %q", i, expectedArgs[i], args[i])
		}
		if i+1 < len(args) && args[i+1] != expectedArgs[i+1] {
			t.Errorf("Expected arg %d to be %q, got %q", i+1, expectedArgs[i+1], args[i+1])
		}
	}
}

func TestToArgs_partial(t *testing.T) {
	cfg := &Config{
		Model: "models/test.gguf",
		Temp:  Ptr(0.8),
	}

	args := ToArgs(cfg)

	if len(args) != 4 {
		t.Errorf("Expected 4 args, got %d: %v", len(args), args)
	}

	if args[0] != "--model" || args[1] != "models/test.gguf" {
		t.Errorf("Expected --model models/test.gguf, got %v", args[:2])
	}

	if args[2] != "--temp" || args[3] != "0.8" {
		t.Errorf("Expected --temp 0.8, got %v", args[2:4])
	}
}

func TestToArgs_flashAttn(t *testing.T) {
	tests := []struct {
		name      string
		flashAttn string
		expected  string
	}{
		{"FlashAttn on", "on", "on"},
		{"FlashAttn off", "off", "off"},
		{"FlashAttn auto", "auto", "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Model:     "models/test.gguf",
				FlashAttn: &tt.flashAttn,
			}
			args := ToArgs(cfg)
			found := false
			for i := 0; i < len(args); i += 2 {
				if args[i] == "--flash-attn" && args[i+1] == tt.expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected --flash-attn %s when FlashAttn=%q, got %v", tt.expected, tt.flashAttn, args)
			}
		})
	}
}

func TestToArgs_booleans(t *testing.T) {
	// Test NoMmap=true
	noMmapTrue := true
	cfg1 := &Config{
		Model:  "models/test.gguf",
		NoMmap: &noMmapTrue,
	}
	args1 := ToArgs(cfg1)
	found := false
	for i := 0; i < len(args1); i += 2 {
		if args1[i] == "--no-mmap" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected --no-mmap when NoMmap=true, got %v", args1)
	}

	// Test NoMmap=false
	noMmapFalse := false
	cfg2 := &Config{
		Model:  "models/test.gguf",
		NoMmap: &noMmapFalse,
	}
	args2 := ToArgs(cfg2)
	found = false
	for i := 0; i < len(args2); i += 2 {
		if args2[i] == "--no-mmap" {
			found = true
			break
		}
	}
	if found {
		t.Errorf("Did not expect --no-mmap when NoMmap=false, got %v", args2)
	}

	// Test ContBatching=true
	contBatchingTrue := true
	cfg3 := &Config{
		Model:        "models/test.gguf",
		ContBatching: &contBatchingTrue,
	}
	args3 := ToArgs(cfg3)
	found = false
	for i := 0; i < len(args3); i += 2 {
		if args3[i] == "--cont-batching" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected --cont-batching when ContBatching=true, got %v", args3)
	}

	// Test ContBatching=false
	contBatchingFalse := false
	cfg4 := &Config{
		Model:        "models/test.gguf",
		ContBatching: &contBatchingFalse,
	}
	args4 := ToArgs(cfg4)
	found = false
	for i := 0; i < len(args4); i += 2 {
		if args4[i] == "--cont-batching" {
			found = true
			break
		}
	}
	if found {
		t.Errorf("Did not expect --cont-batching when ContBatching=false, got %v", args4)
	}
}

func TestStore_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{Dir: tmpDir}

	cfg := &Config{
		Model: "models/test.gguf",
		Temp:  Ptr(0.7),
		Host:  Ptr("127.0.0.1"),
		Port:  Ptr(8080),
	}

	if err := store.Save("test", cfg); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	loaded, err := store.Load("test")
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if loaded.Model != cfg.Model {
		t.Errorf("Model: expected %q, got %q", cfg.Model, loaded.Model)
	}
	if loaded.Temp == nil || *loaded.Temp != *cfg.Temp {
		t.Errorf("Temp: expected %v, got %v", *cfg.Temp, *loaded.Temp)
	}
	if loaded.Host == nil || *loaded.Host != *cfg.Host {
		t.Errorf("Host: expected %q, got %q", *cfg.Host, *loaded.Host)
	}
	if loaded.Port == nil || *loaded.Port != *cfg.Port {
		t.Errorf("Port: expected %v, got %v", *cfg.Port, *loaded.Port)
	}
}

func TestStore_Load_notFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{Dir: tmpDir}

	_, err := store.Load("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent config")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestStore_List(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{Dir: tmpDir}

	// Create some config files
	cfg1 := &Config{Model: "models/test1.gguf"}
	cfg2 := &Config{Model: "models/test2.gguf"}
	cfg3 := &Config{Model: "models/test3.gguf"}

	store.Save("config1", cfg1)
	store.Save("config2", cfg2)
	store.Save("config3", cfg3)

	names, err := store.List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}

	if len(names) != 3 {
		t.Errorf("Expected 3 configs, got %d: %v", len(names), names)
	}

	// Check that all names are present
	for _, name := range []string{"config1", "config2", "config3"} {
		found := false
		for _, n := range names {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected config %q in list", name)
		}
	}
}

func TestStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{Dir: tmpDir}

	cfg := &Config{Model: "models/test.gguf"}
	store.Save("test", cfg)

	if !store.Exists("test") {
		t.Fatal("Expected config to exist before delete")
	}

	if err := store.Delete("test"); err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	if store.Exists("test") {
		t.Error("Expected config to not exist after delete")
	}
}

func TestStore_Save_invalidName(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{Dir: tmpDir}

	cfg := &Config{Model: "models/test.gguf"}

	tests := []struct {
		name     string
		expected string
	}{
		{"", "config name must not be empty"},
		{"../traversal", "config name must not contain"},
		{"path/to/config", "config name must not contain"},
		{"config/with/slashes", "config name must not contain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Save(tt.name, cfg)
			if err == nil {
				t.Errorf("Expected error for invalid name %q", tt.name)
			}
			if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("Expected error containing %q, got %q", tt.expected, err.Error())
			}
		})
	}
}
