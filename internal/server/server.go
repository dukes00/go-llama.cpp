// Package server manages llama-server child processes.
package server

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"go-llama.cpp/internal/config"
)

var ErrBinaryNotFound = errors.New("llama-server not found in $PATH. Install llama.cpp first.")

// Instance represents a running llama-server process.
type Instance struct {
	ModelName string
	Cmd       *exec.Cmd
	Port      int
	LogFile   string
	PID       int
}

// Options holds everything needed to start a server.
type Options struct {
	Config    *config.Config
	ModelPath string // full path to .gguf file
	LogDir    string // directory for log files
}

// FindBinary locates the llama-server binary in PATH.
func FindBinary() (string, error) {
	return exec.LookPath("llama-server")
}

// Start launches a llama-server instance in the background.
func Start(opts Options) (*Instance, error) {
	binaryPath, err := FindBinary()
	if err != nil {
		return nil, ErrBinaryNotFound
	}

	// Build a shallow copy so we can set the model path without mutating the caller's config.
	cfgCopy := *opts.Config
	cfgCopy.Model = opts.ModelPath
	args := config.ToArgs(&cfgCopy)

	modelName := opts.Config.Model
	if modelName == "" && opts.ModelPath != "" {
		modelName = filepath.Base(opts.ModelPath)
	}
	if modelName == "" {
		return nil, errors.New("model name is required")
	}

	port := 8080
	if opts.Config.Port != nil {
		port = *opts.Config.Port
	}

	logPath := filepath.Join(opts.LogDir, modelName+"_"+strconv.Itoa(port)+".log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	sysProcAttr := sysProcAttr()
	cmd.SysProcAttr = sysProcAttr

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return nil, err
	}
	logFile.Close() // child has inherited the fd; parent no longer needs it

	return &Instance{
		ModelName: modelName,
		Cmd:       cmd,
		Port:      port,
		LogFile:   logPath,
		PID:       cmd.Process.Pid,
	}, nil
}

// Stop terminates a llama-server process by PID.
func Stop(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	err = stopProcess(process)
	if err != nil {
		return err
	}

	return nil
}
