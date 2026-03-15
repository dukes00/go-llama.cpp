package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go-llama.cpp/internal/homedir"
	"go-llama.cpp/internal/model"
)

func cmdModels() *cobra.Command {
	return &cobra.Command{
		Use:     "models",
		Short:   "List downloaded models",
		Aliases: []string{"model-list"},
		RunE: func(cmd *cobra.Command, args []string) error {
			layout, err := homedir.Resolve()
			if err != nil {
				return err
			}

			mgr := model.Manager{ModelsDir: layout.Models}
			models, err := mgr.List()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tSIZE\tMODIFIED")
			for _, m := range models {
				fmt.Fprintf(w, "%s\t%s\t%s\n",
					m.Name,
					formatBytes(m.Size),
					m.ModTime.Format("2006-01-02"),
				)
			}
			return w.Flush()
		},
	}
}

// formatBytes formats a byte count as a human-readable string.
func formatBytes(n int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)
	switch {
	case n >= TB:
		return fmt.Sprintf("%.1f TB", float64(n)/TB)
	case n >= GB:
		return fmt.Sprintf("%.1f GB", float64(n)/GB)
	case n >= MB:
		return fmt.Sprintf("%.1f MB", float64(n)/MB)
	case n >= KB:
		return fmt.Sprintf("%.1f KB", float64(n)/KB)
	default:
		return fmt.Sprintf("%d B", n)
	}
}
