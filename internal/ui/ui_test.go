package ui

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestConfirmParsing(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		defaultYes     bool
		expectedResult bool
	}{
		{"input y", "y\n", false, true},
		{"input n", "n\n", false, false},
		{"input empty with defaultYes=true", "\n", true, true},
		{"input empty with defaultYes=false", "\n", false, false},
		{"input Y", "Y\n", false, true},
		{"input YES", "YES\n", false, true},
		{"input yes", "yes\n", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputReader := io.NopCloser(strings.NewReader(tt.input))
			ui := &UI{
				In:  inputReader,
				Out: &bytes.Buffer{},
				Err: &bytes.Buffer{},
			}
			result, err := ui.Confirm("Test question", tt.defaultYes)
			if err != nil {
				t.Fatalf("Confirm returned error: %v", err)
			}
			if result != tt.expectedResult {
				t.Errorf("Confirm() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestConfirmWithDefault(t *testing.T) {
	// Test that empty input returns defaultYes
	t.Run("empty input with defaultYes=true", func(t *testing.T) {
		inputReader := io.NopCloser(strings.NewReader("\n"))
		ui := &UI{
			In:  inputReader,
			Out: &bytes.Buffer{},
			Err: &bytes.Buffer{},
		}
		result, err := ui.Confirm("Test question", true)
		if err != nil {
			t.Fatalf("Confirm returned error: %v", err)
		}
		if !result {
			t.Errorf("Confirm() = %v, want true", result)
		}
	})

	t.Run("empty input with defaultYes=false", func(t *testing.T) {
		inputReader := io.NopCloser(strings.NewReader("\n"))
		ui := &UI{
			In:  inputReader,
			Out: &bytes.Buffer{},
			Err: &bytes.Buffer{},
		}
		result, err := ui.Confirm("Test question", false)
		if err != nil {
			t.Fatalf("Confirm returned error: %v", err)
		}
		if result {
			t.Errorf("Confirm() = %v, want false", result)
		}
	})
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	ui := &UI{
		Out: &buf,
		Err: &buf,
	}
	ui.Info("Hello %s", "World")
	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Expected [INFO] prefix, got %q", output)
	}
	if !strings.Contains(output, "Hello World") {
		t.Errorf("Expected message content, got %q", output)
	}
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	ui := &UI{
		Out: &buf,
		Err: &buf,
	}
	ui.Warn("Warning message")
	output := buf.String()
	if !strings.Contains(output, "[WARN]") {
		t.Errorf("Expected [WARN] prefix, got %q", output)
	}
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	ui := &UI{
		Out: &buf,
		Err: &buf,
	}
	ui.Error("Error message")
	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("Expected [ERROR] prefix, got %q", output)
	}
}

func TestPrompt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple input", "hello world\n", "hello world"},
		{"input with spaces", "  test  \n", "test"},
		{"empty input", "\n", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := &UI{
				In:  io.NopCloser(strings.NewReader(tt.input)),
				Out: &bytes.Buffer{},
				Err: &bytes.Buffer{},
			}
			result, err := ui.Prompt("Question: ")
			if err != nil {
				t.Fatalf("Prompt returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Prompt() = %q, want %q", result, tt.expected)
			}
		})
	}
}
