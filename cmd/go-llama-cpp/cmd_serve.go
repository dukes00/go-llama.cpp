// Package main is the entry point for go-llama-cpp CLI.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go-llama.cpp/internal/config"
	"go-llama.cpp/internal/homedir"
	"go-llama.cpp/internal/registry"
	"go-llama.cpp/internal/server"
	"go-llama.cpp/internal/ui"
	"go-llama.cpp/internal/wizard"
)

var (
	configFlag     string
	overrideFlag   bool
	portFlag       int
	ctxSizeFlag    int
	nGpuLayersFlag int
	tempFlag       float64
	threadsFlag    int
	flashAttnFlag  bool
	noMmapFlag     bool
	cacheTypeKFlag string
	cacheTypeVFlag string
	parallelFlag   int
)

// cmdServe implements the serve subcommand.
func cmdServe() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve <model>",
		Short: "Start a llama-server instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelName := args[0]

			// Resolve home directory
			layout, err := homedir.Resolve()
			if err != nil {
				return fmt.Errorf("resolving home directory: %w", err)
			}

			// Construct model path
			modelPath := filepath.Join(layout.Models, modelName+".gguf")

			// Validate model name (no path separators or ..)
			if strings.Contains(modelName, "/") || strings.Contains(modelName, "\\") || strings.Contains(modelName, "..") {
				return errors.New("model name must not contain /, \\, or ..")
			}

			// Check if model file exists
			if _, err := os.Stat(modelPath); os.IsNotExist(err) {
				return fmt.Errorf("model file not found: %s", modelPath)
			}

			// Build config
			var cfg *config.Config

			// Create wizard for interactive preset save/load
			configStore := &config.Store{Dir: layout.Configs}
			wz := &wizard.Wizard{Store: configStore, UI: ui.Default}

			if configFlag != "" {
				// Load from config file
				cfg, err = configStore.Load(configFlag)
				if err != nil {
					return fmt.Errorf("loading config %q: %w", configFlag, err)
				}

				// Check if any CLI flags were set
				if cliOverridesExist(cmd) {
					if !overrideFlag {
						// --config set but no --override: warn and ignore flags
						ui.Warn("Ignoring CLI flags because --config is set. Use --override to apply them.")
					} else {
						// --config --override: merge flags into config
						overrides := buildConfigFromFlags(cmd, modelName)
						cfg = mergeConfigWithFlags(cfg, overrides)

						// Prompt to save the modified config as a new preset
						saved, presetName, err := wz.PromptSaveOverride(configFlag, cfg)
						if err != nil {
							return fmt.Errorf("prompting to save override: %w", err)
						}
						if saved {
							ui.Info("Override saved as preset '%s'.", presetName)
						}
					}
				}
			} else {
				// No --config flag: check for auto-load or build from flags
				// Step 1: Try to auto-load preset by model name
				if loadedCfg, found := wz.AutoLoadPreset(modelName); found {
					cfg = loadedCfg
					// Auto-loaded preset - check if CLI flags were provided
					if cliOverridesExist(cmd) {
						// CLI flags override the preset - build config from flags
						// Do NOT prompt to save (explicit override)
						cfg = buildConfigFromFlags(cmd, modelName)
						ui.Info("Preset '%s' exists but CLI flags provided. Using CLI flags.", modelName)
					}
					// If no CLI flags, use the loaded config as-is
				} else {
					// No preset found or not loaded - build config from flags
					cfg = buildConfigFromFlags(cmd, modelName)

					// Step 2: If any flags were set (not just bare `serve model`), prompt to save
					if cliOverridesExist(cmd) {
						saved, presetName, err := wz.PromptSavePreset(modelName, cfg)
						if err != nil {
							return fmt.Errorf("prompting to save preset: %w", err)
						}
						if saved {
							ui.Info("Preset '%s' saved.", presetName)
						}
					}
				}
			}

			// Start server
			opts := server.Options{
				Config:    cfg,
				ModelPath: modelPath,
				LogDir:    layout.Logs,
			}

			instance, err := server.Start(opts)
			if err != nil {
				return fmt.Errorf("starting server: %w", err)
			}

			ui.Info("Server started for %s on port %d (PID %d)", modelName, instance.Port, instance.PID)
			ui.Info("Logs: %s", instance.LogFile)

			// Register the running server
			reg, err := registry.New(layout.State)
			if err != nil {
				ui.Warn("Failed to register server: %v", err)
				return nil
			}

			entry := registry.Entry{
				ModelName: modelName,
				PID:       instance.PID,
				Port:      instance.Port,
				LogFile:   instance.LogFile,
				StartedAt: time.Now().Format(time.RFC3339),
			}

			if err := reg.Add(entry); err != nil {
				if errors.Is(err, registry.ErrAlreadyRunning) {
					ui.Warn("Server for %s already running on port %d (PID %d)", modelName, instance.Port, instance.PID)
					// Stop the process we just started
					if err := server.Stop(instance.PID); err != nil {
						ui.Warn("Failed to stop duplicate server: %v", err)
					}
					return err
				}
				return fmt.Errorf("registering server: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&configFlag, "config", "", "Name of a saved preset")
	cmd.Flags().BoolVar(&overrideFlag, "override", false, "Allow flag overrides when using --config")
	cmd.Flags().IntVar(&portFlag, "port", 0, "Port to bind")
	cmd.Flags().IntVar(&ctxSizeFlag, "ctx-size", 0, "Context size")
	cmd.Flags().IntVar(&nGpuLayersFlag, "n-gpu-layers", 0, "GPU layers")
	cmd.Flags().Float64Var(&tempFlag, "temp", 0, "Temperature")
	cmd.Flags().IntVar(&threadsFlag, "threads", 0, "Thread count")
	cmd.Flags().BoolVar(&flashAttnFlag, "flash-attn", false, "Enable flash attention")
	cmd.Flags().BoolVar(&noMmapFlag, "no-mmap", false, "Disable mmap")
	cmd.Flags().StringVar(&cacheTypeKFlag, "cache-type-k", "", "K cache type")
	cmd.Flags().StringVar(&cacheTypeVFlag, "cache-type-v", "", "V cache type")
	cmd.Flags().IntVar(&parallelFlag, "parallel", 0, "Parallel sequences")

	return cmd
}

// buildConfigFromFlags creates a Config from CLI flags.
func buildConfigFromFlags(cmd *cobra.Command, modelName string) *config.Config {
	cfg := &config.Config{
		Model: modelName,
	}

	if cmd.Flags().Changed("port") {
		cfg.Port = &portFlag
	}
	if cmd.Flags().Changed("ctx-size") {
		cfg.CtxSize = &ctxSizeFlag
	}
	if cmd.Flags().Changed("n-gpu-layers") {
		cfg.NGPULayers = &nGpuLayersFlag
	}
	if cmd.Flags().Changed("temp") {
		cfg.Temp = &tempFlag
	}
	if cmd.Flags().Changed("threads") {
		cfg.Threads = &threadsFlag
	}
	if cmd.Flags().Changed("flash-attn") {
		cfg.FlashAttn = &flashAttnFlag
	}
	if cmd.Flags().Changed("no-mmap") {
		cfg.NoMmap = &noMmapFlag
	}
	if cmd.Flags().Changed("cache-type-k") {
		cfg.CacheTypeK = &cacheTypeKFlag
	}
	if cmd.Flags().Changed("cache-type-v") {
		cfg.CacheTypeV = &cacheTypeVFlag
	}
	if cmd.Flags().Changed("parallel") {
		cfg.NParallel = &parallelFlag
	}

	return cfg
}

// mergeConfigWithFlags merges overrides into base config.
func mergeConfigWithFlags(base *config.Config, overrides *config.Config) *config.Config {
	if overrides == nil {
		return base
	}

	if overrides.Model != "" {
		base.Model = overrides.Model
	}
	if overrides.Port != nil {
		base.Port = overrides.Port
	}
	if overrides.CtxSize != nil {
		base.CtxSize = overrides.CtxSize
	}
	if overrides.NGPULayers != nil {
		base.NGPULayers = overrides.NGPULayers
	}
	if overrides.Temp != nil {
		base.Temp = overrides.Temp
	}
	if overrides.Threads != nil {
		base.Threads = overrides.Threads
	}
	if overrides.FlashAttn != nil {
		base.FlashAttn = overrides.FlashAttn
	}
	if overrides.NoMmap != nil {
		base.NoMmap = overrides.NoMmap
	}
	if overrides.CacheTypeK != nil {
		base.CacheTypeK = overrides.CacheTypeK
	}
	if overrides.CacheTypeV != nil {
		base.CacheTypeV = overrides.CacheTypeV
	}
	if overrides.NParallel != nil {
		base.NParallel = overrides.NParallel
	}

	return base
}

// cliOverridesExist returns true if any llama-server flags were explicitly set.
func cliOverridesExist(cmd *cobra.Command) bool {
	return cmd.Flags().Changed("port") ||
		cmd.Flags().Changed("ctx-size") ||
		cmd.Flags().Changed("n-gpu-layers") ||
		cmd.Flags().Changed("temp") ||
		cmd.Flags().Changed("threads") ||
		cmd.Flags().Changed("flash-attn") ||
		cmd.Flags().Changed("no-mmap") ||
		cmd.Flags().Changed("cache-type-k") ||
		cmd.Flags().Changed("cache-type-v") ||
		cmd.Flags().Changed("parallel")
}
