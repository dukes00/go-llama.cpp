// Package config manages JSON configuration presets for llama-server.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var ErrNotFound = errors.New("config not found")

// Config represents a llama-server configuration preset.
type Config struct {
	Model        string            `json:"model"`
	Host         *string           `json:"host,omitempty"`
	Port         *int              `json:"port,omitempty"`
	CtxSize      *int              `json:"ctx_size,omitempty"`
	NGPULayers   *int              `json:"n_gpu_layers,omitempty"`
	Temp         *float64          `json:"temp,omitempty"`
	Threads      *int              `json:"threads,omitempty"`
	FlashAttn    *bool             `json:"flash_attn,omitempty"`
	NoMmap       *bool             `json:"no_mmap,omitempty"`
	CacheTypeK   *string           `json:"cache_type_k,omitempty"`
	CacheTypeV   *string           `json:"cache_type_v,omitempty"`
	ContBatching *bool             `json:"cont_batching,omitempty"`
	NParallel    *int              `json:"n_parallel,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

// Ptr creates a pointer to a value.
func Ptr[T any](v T) *T {
	return &v
}

// ToArgs converts a Config to a slice of llama-server CLI arguments.
func ToArgs(c *Config) []string {
	var args []string

	// Model is required, always include it
	args = append(args, "--model", c.Model)

	// Skip nil pointer fields (they are "not set")
	if c.Host != nil {
		args = append(args, "--host", *c.Host)
	}
	if c.Port != nil {
		args = append(args, "--port", strconv.Itoa(*c.Port))
	}
	if c.CtxSize != nil {
		args = append(args, "--ctx-size", strconv.Itoa(*c.CtxSize))
	}
	if c.NGPULayers != nil {
		args = append(args, "--n-gpu-layers", strconv.Itoa(*c.NGPULayers))
	}
	if c.Temp != nil {
		args = append(args, "--temp", strconv.FormatFloat(*c.Temp, 'f', -1, 64))
	}
	if c.Threads != nil {
		args = append(args, "--threads", strconv.Itoa(*c.Threads))
	}

	// Boolean flags - only include if true
	if c.FlashAttn != nil && *c.FlashAttn {
		args = append(args, "--flash-attn")
	}
	if c.NoMmap != nil && *c.NoMmap {
		args = append(args, "--no-mmap")
	}
	if c.ContBatching != nil && *c.ContBatching {
		args = append(args, "--cont-batching")
	}

	// String pointer fields
	if c.CacheTypeK != nil {
		args = append(args, "--cache-type-k", *c.CacheTypeK)
	}
	if c.CacheTypeV != nil {
		args = append(args, "--cache-type-v", *c.CacheTypeV)
	}

	// Integer pointer fields
	if c.NParallel != nil {
		args = append(args, "--parallel", strconv.Itoa(*c.NParallel))
	}

	// Extra flags - iterate sorted keys
	if c.Extra != nil {
		keys := make([]string, 0, len(c.Extra))
		for k := range c.Extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			args = append(args, "--"+k, c.Extra[k])
		}
	}

	return args
}

// Store manages reading and writing config files from a directory.
type Store struct {
	Dir string // path to the configs directory
}

// Save writes a config to a JSON file.
func (s *Store) Save(name string, cfg *Config) error {
	// Validate name
	if name == "" {
		return errors.New("config name must not be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return errors.New("config name must not contain /, \\, or ..")
	}

	// Ensure directory exists
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}

	// Marshal to JSON
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	// Write to file
	path := filepath.Join(s.Dir, name+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}

	return nil
}

// Load reads a config from a JSON file.
func (s *Store) Load(name string) (*Config, error) {
	// Validate name
	if name == "" {
		return nil, errors.New("config name must not be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return nil, errors.New("config name must not contain /, \\, or ..")
	}

	path := filepath.Join(s.Dir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Exists checks if a config file exists.
func (s *Store) Exists(name string) bool {
	path := filepath.Join(s.Dir, name+".json")
	_, err := os.Stat(path)
	return err == nil
}

// List returns all config names.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			names = append(names, strings.TrimSuffix(entry.Name(), ".json"))
		}
	}
	return names, nil
}

// Delete removes a config file.
func (s *Store) Delete(name string) error {
	// Validate name
	if name == "" {
		return errors.New("config name must not be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return errors.New("config name must not contain /, \\, or ..")
	}

	path := filepath.Join(s.Dir, name+".json")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}

	return nil
}
