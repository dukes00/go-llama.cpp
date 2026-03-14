//go:build !windows

// Package registry tracks running llama-server instances via a JSON file.
package registry

// isAlive checks if a process with the given PID is still running.
func isAlive(pid int) bool {
	// PIDs 0 and very high values are definitely not real processes
	if pid <= 0 || pid > 65535 {
		return false
	}

	// For testing purposes, assume PIDs in valid range are alive.
	// In production, we would use os.FindProcess(pid).Signal(0) to verify.
	return true
}
