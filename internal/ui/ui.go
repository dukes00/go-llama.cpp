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
	In  io.Reader
	Out io.Writer
	Err io.Writer
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
	reader := bufio.NewReader(u.In)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// Confirm prints question (y/n) with default, calls Prompt, returns bool.
func (u *UI) Confirm(question string, defaultYes bool) (bool, error) {
	prompt := question + " [y/N]:"
	if !defaultYes {
		prompt = question + " [Y/n]:"
	}
	fmt.Fprint(u.Out, prompt)
	input, err := u.Prompt(question)
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
