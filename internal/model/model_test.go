package model

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestResolve_exactMatch tests resolving a model that exists with exact filename.
func TestResolve_exactMatch(t *testing.T) {
	dir := t.TempDir()
	testFile := dir + "/mymodel.gguf"
	if err := os.WriteFile(testFile, []byte("fake model"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := Manager{ModelsDir: dir}
	path, err := m.Resolve("mymodel.gguf")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != testFile {
		t.Errorf("expected %q, got %q", testFile, path)
	}
}

// TestResolve_withoutExtension tests resolving a model without extension.
func TestResolve_withoutExtension(t *testing.T) {
	dir := t.TempDir()
	testFile := dir + "/mymodel.gguf"
	if err := os.WriteFile(testFile, []byte("fake model"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := Manager{ModelsDir: dir}
	path, err := m.Resolve("mymodel")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != testFile {
		t.Errorf("expected %q, got %q", testFile, path)
	}
}

// TestResolve_notFound tests resolving a model that doesn't exist.
func TestResolve_notFound(t *testing.T) {
	dir := t.TempDir()
	m := Manager{ModelsDir: dir}
	_, err := m.Resolve("nonexistent")
	if !IsModelNotFound(err) {
		t.Errorf("expected ErrModelNotFound, got %v", err)
	}
}

// TestResolve_pathTraversal tests that path traversal is rejected.
func TestResolve_pathTraversal(t *testing.T) {
	dir := t.TempDir()
	m := Manager{ModelsDir: dir}
	_, err := m.Resolve("../etc/passwd")
	if err == nil || !strings.Contains(err.Error(), "..") {
		t.Errorf("expected validation error for path traversal, got %v", err)
	}
}

// IsModelNotFound checks if an error is ErrModelNotFound.
func IsModelNotFound(err error) bool {
	return err == ErrModelNotFound
}

// TestList_empty tests listing models in an empty directory.
func TestList_empty(t *testing.T) {
	dir := t.TempDir()
	m := Manager{ModelsDir: dir}
	models, err := m.List()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(models) != 0 {
		t.Errorf("expected empty list, got %d items", len(models))
	}
}

// TestList_multipleModels tests listing multiple models.
func TestList_multipleModels(t *testing.T) {
	dir := t.TempDir()

	// Create 3 .gguf files and 1 .txt file
	files := []string{
		dir + "/model1.gguf",
		dir + "/model2.gguf",
		dir + "/model3.gguf",
		dir + "/notamodel.txt",
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("fake model"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	m := Manager{ModelsDir: dir}
	models, err := m.List()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(models) != 3 {
		t.Errorf("expected 3 models, got %d", len(models))
	}
}

// TestList_sorted tests that models are returned in alphabetical order.
func TestList_sorted(t *testing.T) {
	dir := t.TempDir()

	files := []string{
		dir + "/zebra.gguf",
		dir + "/alpha.gguf",
		dir + "/middle.gguf",
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("fake model"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	m := Manager{ModelsDir: dir}
	models, err := m.List()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for i := 1; i < len(models); i++ {
		if models[i-1].Name >= models[i].Name {
			t.Errorf("models not sorted: %q >= %q", models[i-1].Name, models[i].Name)
		}
	}
}

// TestParseHF_fullURL tests parsing a full Hugging Face URL.
func TestParseHF_fullURL(t *testing.T) {
	url := "https://huggingface.co/Qwen/Qwen3-8B-GGUF/resolve/main/qwen3-8b-q4_k_m.gguf"
	downloadURL, fileName, err := ParseHuggingFaceURL(url)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if downloadURL != url {
		t.Errorf("expected %q, got %q", url, downloadURL)
	}
	if fileName != "qwen3-8b-q4_k_m.gguf" {
		t.Errorf("expected %q, got %q", "qwen3-8b-q4_k_m.gguf", fileName)
	}
}

// TestParseHF_shortForm tests parsing a short form URL.
func TestParseHF_shortForm(t *testing.T) {
	input := "Qwen/Qwen3-8B-GGUF/qwen3-8b-q4_k_m.gguf"
	downloadURL, fileName, err := ParseHuggingFaceURL(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expectedURL := "https://huggingface.co/Qwen/Qwen3-8B-GGUF/resolve/main/qwen3-8b-q4_k_m.gguf"
	if downloadURL != expectedURL {
		t.Errorf("expected %q, got %q", expectedURL, downloadURL)
	}
	if fileName != "qwen3-8b-q4_k_m.gguf" {
		t.Errorf("expected %q, got %q", "qwen3-8b-q4_k_m.gguf", fileName)
	}
}

// TestParseHF_arbitraryURL tests parsing an arbitrary URL.
func TestParseHF_arbitraryURL(t *testing.T) {
	url := "https://example.com/model.gguf"
	downloadURL, fileName, err := ParseHuggingFaceURL(url)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if downloadURL != url {
		t.Errorf("expected %q, got %q", url, downloadURL)
	}
	if fileName != "model.gguf" {
		t.Errorf("expected %q, got %q", "model.gguf", fileName)
	}
}

// TestParseHF_invalid tests parsing an invalid input.
func TestParseHF_invalid(t *testing.T) {
	_, _, err := ParseHuggingFaceURL("just-a-name")
	if err == nil {
		t.Error("expected error for invalid input")
	}
}

// TestDownload_success tests successful download.
func TestDownload_success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1024")
		w.Write(make([]byte, 1024))
	}))
	defer server.Close()

	dir := t.TempDir()
	m := Manager{ModelsDir: dir}

	progress := make(chan DownloadProgress)
	err := m.Download(server.URL, "test-model", progress)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify file exists
	path := dir + "/test-model.gguf"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist at %q", path)
	}

	// Verify content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected no error reading file, got %v", err)
	}
	if len(content) != 1024 {
		t.Errorf("expected 1024 bytes, got %d", len(content))
	}
}

// TestDownload_progress tests that progress updates are sent.
func TestDownload_progress(t *testing.T) {
	var receivedProgress []DownloadProgress

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1024")
		w.Write(make([]byte, 1024))
	}))
	defer server.Close()

	dir := t.TempDir()
	m := Manager{ModelsDir: dir}

	progress := make(chan DownloadProgress)
	go func() {
		err := m.Download(server.URL, "test-model", progress)
		if err != nil {
			t.Logf("download error: %v", err)
		}
	}()

	// Wait for download to complete
	done := make(chan struct{})
	go func() {
		for p := range progress {
			receivedProgress = append(receivedProgress, p)
			if p.Done {
				close(done)
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("download timed out")
	}

	// Verify we received progress updates
	if len(receivedProgress) < 2 {
		t.Errorf("expected at least 2 progress updates, got %d", len(receivedProgress))
	}

	// Verify final progress
	var finalProgress DownloadProgress
	for _, p := range receivedProgress {
		if p.Done {
			finalProgress = p
			break
		}
	}
	if !finalProgress.Done {
		t.Error("expected final progress to have Done=true")
	}
}

// TestDownload_alreadyExists tests that downloading an existing file fails.
func TestDownload_alreadyExists(t *testing.T) {
	dir := t.TempDir()
	testFile := dir + "/test-model.gguf"
	if err := os.WriteFile(testFile, []byte("fake model"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := Manager{ModelsDir: dir}
	progress := make(chan DownloadProgress)
	err := m.Download("https://example.com/model.gguf", "test-model", progress)
	if err == nil {
		t.Error("expected error for existing file")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got %v", err)
	}

	// Verify channel was closed
	select {
	case <-progress:
		// Channel received a message, which is expected
	default:
		t.Error("expected progress channel to be closed")
	}
}

// TestDownload_serverError tests that server errors are handled correctly.
func TestDownload_serverError(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	dir := t.TempDir()
	m := Manager{ModelsDir: dir}

	progress := make(chan DownloadProgress)
	err := m.Download(server.URL, "test-model", progress)
	if err == nil {
		t.Error("expected error for 404 response")
	}

	// Verify no temp file was left behind
	_, err = os.Stat(dir + "/test-model.gguf.download.tmp")
	if err == nil {
		t.Error("expected temp file to be removed on error")
	}

	// Verify channel was closed
	select {
	case <-progress:
		// Channel received a message, which is expected
	default:
		t.Error("expected progress channel to be closed")
	}
}

// TestDownload_interrupted tests that interrupted downloads clean up temp files.
func TestDownload_interrupted(t *testing.T) {
	// Create test server that closes connection mid-transfer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10000")
		w.Write(make([]byte, 500)) // Send only half
	}))
	defer server.Close()

	dir := t.TempDir()
	m := Manager{ModelsDir: dir}

	progress := make(chan DownloadProgress)
	err := m.Download(server.URL, "test-model", progress)
	if err == nil {
		t.Error("expected error for interrupted download")
	}

	// Verify no temp file was left behind
	_, err = os.Stat(dir + "/test-model.gguf.download.tmp")
	if err == nil {
		t.Error("expected temp file to be removed on error")
	}

	// Verify channel was closed
	select {
	case <-progress:
		// Channel received a message, which is expected
	default:
		t.Error("expected progress channel to be closed")
	}
}

// TestDownload_progress_updates tests that progress updates are sent at correct intervals.
func TestDownload_progress_updates(t *testing.T) {
	var receivedProgress []DownloadProgress

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "2048")
		w.Write(make([]byte, 2048))
	}))
	defer server.Close()

	dir := t.TempDir()
	m := Manager{ModelsDir: dir}

	progress := make(chan DownloadProgress)
	go func() {
		err := m.Download(server.URL, "test-model", progress)
		if err != nil {
			t.Logf("download error: %v", err)
		}
	}()

	// Wait for download to complete
	done := make(chan struct{})
	go func() {
		for p := range progress {
			receivedProgress = append(receivedProgress, p)
			if p.Done {
				close(done)
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("download timed out")
	}

	// Verify we received progress updates
	if len(receivedProgress) < 2 {
		t.Errorf("expected at least 2 progress updates, got %d", len(receivedProgress))
	}

	// Verify percentages are increasing
	for i := 1; i < len(receivedProgress); i++ {
		if receivedProgress[i].Percent <= receivedProgress[i-1].Percent {
			t.Errorf("percentages not increasing: %f <= %f", receivedProgress[i].Percent, receivedProgress[i-1].Percent)
		}
	}
}
