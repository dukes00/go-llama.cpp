package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go-llama.cpp/internal/homedir"
	"go-llama.cpp/internal/model"
	"go-llama.cpp/internal/ui"
)

func cmdDownload() *cobra.Command {
	var nameFlag string

	cmd := &cobra.Command{
		Use:   "download <source>",
		Short: "Download a GGUF model file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			layout, err := homedir.Resolve()
			if err != nil {
				return err
			}
			if err := layout.EnsureExists(); err != nil {
				return err
			}

			downloadURL, fileName, err := model.ParseHuggingFaceURL(args[0])
			if err != nil {
				return err
			}

			if nameFlag != "" {
				fileName = nameFlag
			}

			mgr := model.Manager{ModelsDir: layout.Models}
			progress := make(chan model.DownloadProgress)

			ui.Info("Downloading %s...", fileName)

			var dlErr error
			go func() {
				dlErr = mgr.Download(downloadURL, fileName, progress)
			}()

			for p := range progress {
				if p.Err != nil {
					return p.Err
				}
				fmt.Fprint(os.Stdout, renderProgressBar(p))
				if p.Done {
					fmt.Fprintln(os.Stdout)
				}
			}

			if dlErr != nil {
				return dlErr
			}

			destPath, _ := mgr.Resolve(fileName)
			ui.Info("Downloaded %s to %s", fileName, destPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "Override the saved filename")
	return cmd
}

// renderProgressBar returns a progress bar string for the given progress update.
func renderProgressBar(p model.DownloadProgress) string {
	const barWidth = 40

	if p.TotalBytes < 0 {
		return fmt.Sprintf("\r  %s downloaded", formatBytes(p.BytesDownloaded))
	}

	filled := int(float64(barWidth) * p.Percent / 100)
	if filled > barWidth {
		filled = barWidth
	}

	bar := make([]byte, barWidth)
	for i := range bar {
		if i < filled {
			bar[i] = '='
		} else {
			bar[i] = ' '
		}
	}
	if filled < barWidth {
		bar[filled] = '>'
	}

	return fmt.Sprintf("\r[%s] %.0f%% (%s / %s)",
		string(bar),
		p.Percent,
		formatBytes(p.BytesDownloaded),
		formatBytes(p.TotalBytes),
	)
}
