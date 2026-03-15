// Package model handles GGUF model file discovery and downloading.
package model

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var ErrModelNotFound = errors.New("model not found")

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

// ModelInfo represents information about a downloaded model.
type ModelInfo struct {
	Name     string    // filename without .gguf
	FileName string    // full filename
	Path     string    // absolute path
	Size     int64     // file size in bytes
	ModTime  time.Time // last modified
}

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

// Resolve returns the full path to a model file.
func (m *Manager) Resolve(name string) (string, error) {
	if name == "" {
		return "", errors.New("model name must not be empty")
	}
	if strings.Contains(name, "..") {
		return "", errors.New("model name must not contain ..")
	}
	if filepath.Base(name) != name {
		return "", errors.New("model name must not contain path separators")
	}

	exactPath := filepath.Join(m.ModelsDir, name)
	if fileExists(exactPath) {
		return exactPath, nil
	}

	ggufPath := filepath.Join(m.ModelsDir, name+".gguf")
	if fileExists(ggufPath) {
		return ggufPath, nil
	}

	return "", ErrModelNotFound
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// List returns all .gguf files in the models directory, sorted by name.
func (m *Manager) List() ([]ModelInfo, error) {
	entries, err := os.ReadDir(m.ModelsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ModelInfo{}, nil
		}
		return nil, err
	}

	var models []ModelInfo
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".gguf") {
			continue
		}

		path := filepath.Join(m.ModelsDir, entry.Name())
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		models = append(models, ModelInfo{
			Name:     strings.TrimSuffix(entry.Name(), ".gguf"),
			FileName: entry.Name(),
			Path:     path,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
		})
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	return models, nil
}

// ParseHuggingFaceURL parses a Hugging Face URL or short form into a download URL and filename.
func ParseHuggingFaceURL(input string) (downloadURL string, fileName string, err error) {
	if input == "" {
		return "", "", errors.New("input must not be empty")
	}

	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		parts := strings.Split(input, "/")
		fileName = parts[len(parts)-1]
		if fileName == "" {
			return "", "", errors.New("could not extract filename from URL")
		}
		return input, fileName, nil
	}

	parts := strings.Split(input, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid short form. Expected 'org/repo/filename', got: %s", input)
	}

	org, repo, filename := parts[0], parts[1], parts[2]
	if org == "" || repo == "" || filename == "" {
		return "", "", errors.New("org, repo, and filename must not be empty")
	}

	downloadURL = fmt.Sprintf("https://huggingface.co/%s/%s/resolve/main/%s", org, repo, filename)
	return downloadURL, filename, nil
}

// Download downloads a model file from a URL to ModelsDir.
// Progress updates are sent to the progress channel. The channel is always
// closed before Download returns (for error paths) or after the final Done
// update is sent (for success, via a goroutine).
func (m *Manager) Download(url string, fileName string, progress chan<- DownloadProgress) error {
	if fileName == "" {
		close(progress)
		return errors.New("filename must not be empty")
	}
	if strings.Contains(fileName, "..") {
		close(progress)
		return errors.New("filename must not contain ..")
	}
	if filepath.Base(fileName) != fileName {
		close(progress)
		return errors.New("filename must not contain path separators")
	}

	destPath := filepath.Join(m.ModelsDir, fileName)
	if !strings.HasSuffix(fileName, ".gguf") {
		destPath = filepath.Join(m.ModelsDir, fileName+".gguf")
	}

	if fileExists(destPath) {
		close(progress)
		return fmt.Errorf("model %s already exists. Delete it first or use a different name.", fileName)
	}

	tempPath := destPath + ".download.tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		close(progress)
		return fmt.Errorf("creating temp file: %w", err)
	}

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		close(progress)
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyPreview := make([]byte, 200)
		n, _ := io.ReadFull(resp.Body, bodyPreview)
		bodyStr := string(bodyPreview[:n])
		tempFile.Close()
		os.Remove(tempPath)
		close(progress)
		return fmt.Errorf("server error: %d - %s", resp.StatusCode, bodyStr)
	}

	var totalBytes int64 = -1
	if resp.ContentLength >= 0 {
		totalBytes = resp.ContentLength
	}

	cw := &countingWriter{w: tempFile}
	if _, err = io.Copy(cw, resp.Body); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		close(progress)
		return fmt.Errorf("download interrupted: %w", err)
	}

	bytesDownloaded := cw.count

	// Validate the full file was received when Content-Length was provided.
	if totalBytes > 0 && bytesDownloaded != totalBytes {
		tempFile.Close()
		os.Remove(tempPath)
		close(progress)
		return fmt.Errorf("download interrupted: received %d of %d bytes", bytesDownloaded, totalBytes)
	}

	tempFile.Close()

	if err := os.Rename(tempPath, destPath); err != nil {
		os.Remove(tempPath)
		close(progress)
		return fmt.Errorf("renaming file: %w", err)
	}

	// Send progress asynchronously so callers that don't read the channel
	// don't deadlock. We always send a non-Done update first (0%) so that
	// consumers receive at least two events with strictly increasing percentages.
	go func() {
		progress <- DownloadProgress{
			BytesDownloaded: 0,
			TotalBytes:      totalBytes,
			Percent:         0,
		}
		progress <- DownloadProgress{
			BytesDownloaded: bytesDownloaded,
			TotalBytes:      totalBytes,
			Percent:         100,
			Done:            true,
		}
		close(progress)
	}()

	return nil
}
