package registry

import (
	"os"
	"testing"
	"time"
)

func TestRegistry_AddAndList(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create registry
	reg, err := New(dir)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	// Add entries
	entry1 := Entry{
		ModelName: "test-model-1",
		PID:       os.Getpid(),
		Port:      8080,
		LogFile:   "/tmp/test.log",
		StartedAt: time.Now().Format(time.RFC3339),
	}
	if err := reg.Add(entry1); err != nil {
		t.Fatalf("Add() = %v", err)
	}

	entry2 := Entry{
		ModelName: "test-model-2",
		PID:       os.Getpid(),
		Port:      8081,
		LogFile:   "/tmp/test2.log",
		StartedAt: time.Now().Format(time.RFC3339),
	}
	if err := reg.Add(entry2); err != nil {
		t.Fatalf("Add() = %v", err)
	}

	// List entries
	entries, err := reg.List()
	if err != nil {
		t.Fatalf("List() = %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("List() = %d entries, want 2", len(entries))
	}
}

func TestRegistry_Add_duplicate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	reg, err := New(dir)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	entry := Entry{
		ModelName: "test-model",
		PID:       12345,
		Port:      8080,
		LogFile:   "/tmp/test.log",
		StartedAt: time.Now().Format(time.RFC3339),
	}

	if err := reg.Add(entry); err != nil {
		t.Fatalf("First Add() = %v", err)
	}

	// Try to add duplicate
	if err := reg.Add(entry); err != ErrAlreadyRunning {
		t.Errorf("Second Add() = %v, want ErrAlreadyRunning", err)
	}
}

func TestRegistry_Remove(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	reg, err := New(dir)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	entry := Entry{
		ModelName: "test-model",
		PID:       12345,
		Port:      8080,
		LogFile:   "/tmp/test.log",
		StartedAt: time.Now().Format(time.RFC3339),
	}

	if err := reg.Add(entry); err != nil {
		t.Fatalf("Add() = %v", err)
	}

	// Remove entry
	if err := reg.Remove("test-model", 8080); err != nil {
		t.Fatalf("Remove() = %v", err)
	}

	// Verify removed
	entries, err := reg.List()
	if err != nil {
		t.Fatalf("List() = %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("List() = %d entries, want 0", len(entries))
	}
}

func TestRegistry_RemoveByPID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	reg, err := New(dir)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	entry := Entry{
		ModelName: "test-model",
		PID:       12345,
		Port:      8080,
		LogFile:   "/tmp/test.log",
		StartedAt: time.Now().Format(time.RFC3339),
	}

	if err := reg.Add(entry); err != nil {
		t.Fatalf("Add() = %v", err)
	}

	// Remove by PID
	if err := reg.RemoveByPID(12345); err != nil {
		t.Fatalf("RemoveByPID() = %v", err)
	}

	// Verify removed
	entries, err := reg.List()
	if err != nil {
		t.Fatalf("List() = %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("List() = %d entries, want 0", len(entries))
	}
}

func TestRegistry_Remove_notFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	reg, err := New(dir)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	// Try to remove non-existent entry
	err = reg.Remove("nonexistent", 8080)
	if !errorsIs(err, ErrNotFound) {
		t.Errorf("Remove() = %v, want ErrNotFound", err)
	}
}

func TestRegistry_List_cleansDead(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	reg, err := New(dir)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	// Add entry with dead PID (very high PID that doesn't exist)
	entry := Entry{
		ModelName: "test-model",
		PID:       9999999,
		Port:      8080,
		LogFile:   "/tmp/test.log",
		StartedAt: time.Now().Format(time.RFC3339),
	}

	if err := reg.Add(entry); err != nil {
		t.Fatalf("Add() = %v", err)
	}

	// List should auto-clean dead entries
	entries, err := reg.List()
	if err != nil {
		t.Fatalf("List() = %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("List() = %d entries, want 0 (dead PID should be cleaned)", len(entries))
	}
}

func TestRegistry_FindByModel(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	reg, err := New(dir)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	// Add entries for different models
	entry1 := Entry{
		ModelName: "model-a",
		PID:       os.Getpid(),
		Port:      8080,
		LogFile:   "/tmp/test.log",
		StartedAt: time.Now().Format(time.RFC3339),
	}
	if err := reg.Add(entry1); err != nil {
		t.Fatalf("Add() = %v", err)
	}

	entry2 := Entry{
		ModelName: "model-b",
		PID:       os.Getpid(),
		Port:      8081,
		LogFile:   "/tmp/test2.log",
		StartedAt: time.Now().Format(time.RFC3339),
	}
	if err := reg.Add(entry2); err != nil {
		t.Fatalf("Add() = %v", err)
	}

	// Find by model name
	entries, err := reg.FindByModel("model-a")
	if err != nil {
		t.Fatalf("FindByModel() = %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("FindByModel() = %d entries, want 1", len(entries))
	}
	if entries[0].ModelName != "model-a" {
		t.Errorf("FindByModel() = %s, want model-a", entries[0].ModelName)
	}
}

func TestRegistry_persistence(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create first registry and add entry
	reg1, err := New(dir)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	entry := Entry{
		ModelName: "test-model",
		PID:       os.Getpid(),
		Port:      8080,
		LogFile:   "/tmp/test.log",
		StartedAt: time.Now().Format(time.RFC3339),
	}

	if err := reg1.Add(entry); err != nil {
		t.Fatalf("Add() = %v", err)
	}

	// Create new registry instance pointing to same file
	reg2, err := New(dir)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	// List should show the entry
	entries, err := reg2.List()
	if err != nil {
		t.Fatalf("List() = %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("List() = %d entries, want 1", len(entries))
	}
}

// errorsIs is a helper to check if error matches sentinel error.
func errorsIs(err error, target error) bool {
	return err == target
}
