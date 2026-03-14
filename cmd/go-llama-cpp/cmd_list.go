// Package main is the entry point for go-llama-cpp CLI.
package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go-llama.cpp/internal/homedir"
	"go-llama.cpp/internal/registry"
	"go-llama.cpp/internal/ui"
)

// cmdList implements the list subcommand.
func cmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Show running llama-server instances",
		Aliases: []string{"ls", "ps"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve home directory
			layout, err := homedir.Resolve()
			if err != nil {
				return fmt.Errorf("resolving home directory: %w", err)
			}

			// Create registry
			reg, err := registry.New(layout.State)
			if err != nil {
				return fmt.Errorf("creating registry: %w", err)
			}

			// Get list of running servers
			entries, err := reg.List()
			if err != nil {
				return fmt.Errorf("listing servers: %w", err)
			}

			if len(entries) == 0 {
				ui.Info("No running servers.")
				return nil
			}

			// Print formatted table
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "MODEL\t\tPORT\tSTATUS\tPID\t\tSTARTED")
			fmt.Fprintln(w, "------------------------------------------------------------")
			for _, e := range entries {
				fmt.Fprintf(w, "%s\t%d\t%s\t%d\t%s\n",
					e.ModelName, e.Port, "Running", e.PID, e.StartedAt)
			}
			w.Flush()

			return nil
		},
	}

	return cmd
}
