// Package main is the entry point for go-llama-cpp CLI.
package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go-llama.cpp/internal/homedir"
	"go-llama.cpp/internal/registry"
	"go-llama.cpp/internal/server"
	"go-llama.cpp/internal/ui"
)

var (
	killPortFlag int
	killAllFlag  bool
)

// cmdKill implements the kill subcommand.
func cmdKill() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kill <model>",
		Short: "Stop a running llama-server instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelName := args[0]

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

			// Find entries for this model
			entries, err := reg.FindByModel(modelName)
			if err != nil {
				return fmt.Errorf("finding server: %w", err)
			}

			if len(entries) == 0 {
				ui.Error("No running server found for %s", modelName)
				return fmt.Errorf("no server found")
			}

			// Handle multiple entries
			if len(entries) > 1 {
				if !killAllFlag && killPortFlag == 0 {
					// Print all matching entries as a table
					w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
					fmt.Fprintln(w, "MODEL\t\tPORT\tSTATUS\tPID\t\tSTARTED")
					fmt.Fprintln(w, "------------------------------------------------------------")
					for _, e := range entries {
						fmt.Fprintf(w, "%s\t%d\t%s\t%d\t%s\n",
							e.ModelName, e.Port, "Running", e.PID, e.StartedAt)
					}
					w.Flush()

					ui.Error("Multiple instances found. Use --port <port> or --all to specify.")
					return fmt.Errorf("multiple instances found")
				}
			}

			// Determine which entries to kill
			var toKill []registry.Entry
			for _, e := range entries {
				if killAllFlag {
					toKill = append(toKill, e)
				} else if killPortFlag != 0 && e.Port == killPortFlag {
					toKill = append(toKill, e)
				} else if len(entries) == 1 {
					// Only one entry and no port filter
					toKill = append(toKill, e)
				}
			}

			// Kill each entry
			for _, e := range toKill {
				// Stop the process
				if err := server.Stop(e.PID); err != nil {
					ui.Warn("Failed to stop server: %v", err)
				}

				// Remove from registry
				if err := reg.Remove(e.ModelName, e.Port); err != nil {
					ui.Warn("Failed to remove from registry: %v", err)
				}

				ui.Info("Stopped server for %s on port %d (PID %d)", e.ModelName, e.Port, e.PID)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&killPortFlag, "port", 0, "Port to kill (use with model name)")
	cmd.Flags().BoolVar(&killAllFlag, "all", false, "Kill all instances of this model")

	return cmd
}
