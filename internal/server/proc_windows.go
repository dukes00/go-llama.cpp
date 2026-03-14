//go:build windows

// Package server manages llama-server child processes.
package server

import (
	"os"
	"syscall"
)

// sysProcAttr returns platform-specific attributes to detach the process.
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: 0x00000008} // DETACHED_PROCESS
}

// stopProcess terminates a process on Windows.
func stopProcess(p *os.Process) error {
	return p.Kill()
}
