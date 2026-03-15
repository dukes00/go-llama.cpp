// Package ui provides simple formatted output for CLI user interaction.
package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Default is the default UI instance using os.Stdin/Stdout/Stderr.
var Default = &UI{
	In:  os.Stdin,
	Out: os.Stdout,
	Err: os.Stderr,
}

// UI holds the I/O streams for user interaction.
type UI struct {
	In     io.Reader
	Out    io.Writer
	Err    io.Writer
	reader *bufio.Reader
}

// getReader returns a persistent bufio.Reader for In, creating it lazily.
// Reusing the same reader across calls preserves buffered data between prompts.
func (u *UI) getReader() *bufio.Reader {
	if u.reader == nil {
		u.reader = bufio.NewReader(u.In)
	}
	return u.reader
}

// Info prints [INFO] <message> to stdout.
func (u *UI) Info(format string, args ...any) {
	fmt.Fprintf(u.Out, "[INFO] "+format+"\n", args...)
}

// Warn prints [WARN] <message> to stderr.
func (u *UI) Warn(format string, args ...any) {
	fmt.Fprintf(u.Err, "[WARN] "+format+"\n", args...)
}

// Error prints [ERROR] <message> to stderr.
func (u *UI) Error(format string, args ...any) {
	fmt.Fprintf(u.Err, "[ERROR] "+format+"\n", args...)
}

// Prompt prints question to stdout, reads one line from stdin, returns trimmed input.
func (u *UI) Prompt(question string) (string, error) {
	fmt.Fprint(u.Out, question)
	input, err := u.getReader().ReadString('\n')
	if err != nil {
		return strings.TrimSpace(input), err
	}
	return strings.TrimSpace(input), nil
}

// Confirm prints question (y/n) with default, calls Prompt, returns bool.
func (u *UI) Confirm(question string, defaultYes bool) (bool, error) {
	if defaultYes {
		fmt.Fprint(u.Out, question+" [Y/n]:")
	} else {
		fmt.Fprint(u.Out, question+" [y/N]:")
	}
	input, err := u.Prompt("")
	if err != nil {
		return false, err
	}
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return defaultYes, nil
	}
	return input == "y" || input == "yes", nil
}

// Info prints [INFO] <message> to stdout.
func Info(format string, args ...any) {
	Default.Info(format, args...)
}

// Warn prints [WARN] <message> to stderr.
func Warn(format string, args ...any) {
	Default.Warn(format, args...)
}

// Error prints [ERROR] <message> to stderr.
func Error(format string, args ...any) {
	Default.Error(format, args...)
}

// Prompt prints question to stdout, reads one line from stdin, returns trimmed input.
func Prompt(question string) (string, error) {
	return Default.Prompt(question)
}

// Confirm prints question (y/n) with default, calls Prompt, returns bool.
func Confirm(question string, defaultYes bool) (bool, error) {
	return Default.Confirm(question, defaultYes)
}
