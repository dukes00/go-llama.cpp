//go:build windows

// Package registry tracks running llama-server instances via a JSON file.
package registry

import "os"

// isAlive checks if a process with the given PID is still running.
//
// NOTE: On Windows, os.FindProcess(pid) always succeeds even if the process
// doesn't exist. This is a known limitation of the Go runtime on Windows.
// For production use on Windows, consider using the Windows API directly
// via golang.org/x/sys/windows to properly check process existence.
//
// TODO: Implement proper Windows process checking using syscall package.
func isAlive(pid int) bool {
	// On Windows, we assume the process is alive if FindProcess succeeds.
	// This is a simplification - see the comment above for details.
	_, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Windows, Signal(0) doesn't work as expected for process existence checking.
	// We return true as a placeholder until proper Windows implementation is added.
	return true
}
