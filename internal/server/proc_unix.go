//go:build !windows

// Package server manages llama-server child processes.
package server

import (
	"os"
	"syscall"
)

// sysProcAttr returns platform-specific attributes to detach the process.
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

// stopProcess terminates a process on Unix systems.
func stopProcess(p *os.Process) error {
	return p.Signal(syscall.SIGTERM)
}
