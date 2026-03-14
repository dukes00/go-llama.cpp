// Package main is the entry point for go-llama-cpp CLI.
package main

import (
	"os"

	"github.com/spf13/cobra"
)

var homeDir string

func main() {
	rootCmd := &cobra.Command{
		Use:   "go-llama-cpp",
		Short: "A lightweight process manager for llama.cpp",
		Long:  "A lightweight process manager for llama.cpp. Think of it as 'podman for llama.cpp'.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if homeDir != "" {
				os.Setenv("GO_LLAMA_HOME", homeDir)
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&homeDir, "home", "", "Override GO_LLAMA_HOME directory")

	rootCmd.AddCommand(cmdInit())
	rootCmd.AddCommand(cmdServe())
	rootCmd.AddCommand(cmdList())
	rootCmd.AddCommand(cmdKill())
	rootCmd.AddCommand(cmdLogs())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
