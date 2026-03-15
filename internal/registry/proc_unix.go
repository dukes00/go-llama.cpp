//go:build !windows

// Package registry tracks running llama-server instances via a JSON file.
package registry

import (
	"os"
	"syscall"
)

// isAlive checks if a process with the given PID is still running.
func isAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}
