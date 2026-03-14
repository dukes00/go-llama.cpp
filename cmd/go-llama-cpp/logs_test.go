package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestLastNLines(t *testing.T) {
	t.Run("returns last n lines from file with more lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := tmpDir + "/test.txt"

		// Write 100 lines
		var content strings.Builder
		for i := 0; i < 100; i++ {
			content.WriteString(fmt.Sprintf("line %d\n", i))
		}
		if err := os.WriteFile(path, []byte(content.String()), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := lastNLines(path, 10)
		if err != nil {
			t.Fatal(err)
		}

		lines := strings.Split(result, "\n")
		if len(lines) != 10 {
			t.Errorf("expected 10 lines, got %d", len(lines))
		}

		// Check first line is line 91
		if !strings.HasPrefix(lines[0], "line 9") {
			t.Errorf("expected line 91+, got %q", lines[0])
		}
	})

	t.Run("returns all lines when file has fewer lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := tmpDir + "/test.txt"

		// Write 3 lines
		var content strings.Builder
		for i := 0; i < 3; i++ {
			content.WriteString(fmt.Sprintf("line %d\n", i))
		}
		if err := os.WriteFile(path, []byte(content.String()), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := lastNLines(path, 10)
		if err != nil {
			t.Fatal(err)
		}

		lines := strings.Split(result, "\n")
		if len(lines) != 3 {
			t.Errorf("expected 3 lines, got %d", len(lines))
		}
	})
}
