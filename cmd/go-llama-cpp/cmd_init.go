// Package main is the entry point for go-llama-cpp CLI.
package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go-llama.cpp/internal/homedir"
	"go-llama.cpp/internal/ui"
)

// cmdInit implements the init subcommand.
func cmdInit() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the go-llama.cpp home directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			layout, err := homedir.Resolve()
			if err != nil {
				return fmt.Errorf("resolving home directory: %w", err)
			}

			if err := layout.EnsureExists(); err != nil {
				return fmt.Errorf("creating home directory: %w", err)
			}

			ui.Info("Initialized home directory at %s", layout.Root)
			return nil
		},
	}
}
