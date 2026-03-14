// Package registry tracks running llama-server instances via a JSON file.
package registry

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrNotFound       = errors.New("registry entry not found")
	ErrAlreadyRunning = errors.New("server already running")
)

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
	path string
	mu   sync.Mutex
}

// New creates a new Registry for the given state directory.
func New(statePath string) (*Registry, error) {
	fullPath := filepath.Join(statePath, "running.json")
	return &Registry{path: fullPath}, nil
}

// Add adds a new entry to the registry.
func (r *Registry) Add(e Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, err := r.load()
	if err != nil {
		return err
	}

	// Check for duplicate model+port
	for _, existing := range entries {
		if existing.ModelName == e.ModelName && existing.Port == e.Port {
			return ErrAlreadyRunning
		}
	}

	entries = append(entries, e)
	return r.save(entries)
}

// Remove removes an entry by model name and port.
func (r *Registry) Remove(modelName string, port int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, err := r.load()
	if err != nil {
		return err
	}

	// Filter out the matching entry
	var filtered []Entry
	for _, e := range entries {
		if e.ModelName != modelName || e.Port != port {
			filtered = append(filtered, e)
		}
	}

	if len(filtered) == len(entries) {
		return ErrNotFound
	}

	return r.save(filtered)
}

// RemoveByPID removes an entry by PID.
func (r *Registry) RemoveByPID(pid int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, err := r.load()
	if err != nil {
		return err
	}

	// Filter out the matching entry
	var filtered []Entry
	for _, e := range entries {
		if e.PID != pid {
			filtered = append(filtered, e)
		}
	}

	if len(filtered) == len(entries) {
		return ErrNotFound
	}

	return r.save(filtered)
}

// List returns all entries, auto-cleaning dead PIDs.
func (r *Registry) List() ([]Entry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, err := r.load()
	if err != nil {
		return nil, err
	}

	// Auto-cleanup: remove dead entries
	var alive []Entry
	for _, e := range entries {
		if isAlive(e.PID) {
			alive = append(alive, e)
		}
	}

	if len(alive) != len(entries) {
		// Entries were removed, save the cleaned list
		if err := r.save(alive); err != nil {
			return nil, err
		}
	}

	return alive, nil
}

// FindByModel finds all entries for a given model name.
func (r *Registry) FindByModel(modelName string) ([]Entry, error) {
	entries, err := r.List()
	if err != nil {
		return nil, err
	}

	var filtered []Entry
	for _, e := range entries {
		if e.ModelName == modelName {
			filtered = append(filtered, e)
		}
	}

	return filtered, nil
}

// load reads entries from the JSON file.
func (r *Registry) load() ([]Entry, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, err
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// save writes entries to the JSON file.
func (r *Registry) save(entries []Entry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(r.path, data, 0o644); err != nil {
		return err
	}

	return nil
}
