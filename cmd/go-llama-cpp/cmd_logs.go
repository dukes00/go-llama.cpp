// Package main is the entry point for go-llama-cpp CLI.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go-llama.cpp/internal/homedir"
	"go-llama.cpp/internal/registry"
	"go-llama.cpp/internal/ui"
)

var (
	logsPortFlag   int
	logsFollowFlag bool
	logsLinesFlag  int
)

// lastNLines reads the last n lines from a file.
func lastNLines(filePath string, n int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read all lines
	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	// Return last n lines
	if len(lines) <= n {
		return joinLines(lines), nil
	}

	return joinLines(lines[len(lines)-n:]), nil
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

// cmdLogs implements the logs subcommand.
func cmdLogs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <model>",
		Short: "Show logs for a running server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

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

			// Find entry by model
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
				if !logsFollowFlag && logsPortFlag == 0 {
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

			// Determine which entry to show logs for
			var entry *registry.Entry
			for _, e := range entries {
				if logsFollowFlag {
					if logsPortFlag != 0 && e.Port == logsPortFlag {
						entry = &e
						break
					}
				} else {
					if logsPortFlag != 0 && e.Port == logsPortFlag {
						entry = &e
						break
					}
					if entry == nil {
						entry = &e
					}
				}
			}

			if entry == nil {
				ui.Error("No running server found for %s", modelName)
				return fmt.Errorf("no server found")
			}

			// Set default lines if not specified
			if logsLinesFlag == 0 {
				logsLinesFlag = 50
			}

			// Show logs
			if logsFollowFlag {
				// On Unix, use tail -f
				cmd := exec.Command("tail", "-f", entry.LogFile)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					ui.Error("Failed to follow logs: %v", err)
					return err
				}
			} else {
				// Show last N lines
				content, err := lastNLines(entry.LogFile, logsLinesFlag)
				if err != nil {
					ui.Error("Failed to read log file: %v", err)
					return err
				}
				if content != "" {
					fmt.Println(content)
				} else {
					ui.Info("No logs available yet.")
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&logsPortFlag, "port", 0, "Port to show logs for")
	cmd.Flags().BoolVarP(&logsFollowFlag, "follow", "f", false, "Tail the log file")
	cmd.Flags().IntVarP(&logsLinesFlag, "lines", "n", 50, "Number of lines to show")

	return cmd
}
