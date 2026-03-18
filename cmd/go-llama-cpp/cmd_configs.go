package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go-llama.cpp/internal/config"
	"go-llama.cpp/internal/homedir"
	"go-llama.cpp/internal/ui"
)

func cmdConfigs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configs",
		Short: "Manage saved configuration presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigsList()
		},
	}

	cmd.AddCommand(cmdConfigsList())
	cmd.AddCommand(cmdConfigsShow())
	cmd.AddCommand(cmdConfigsNew())
	cmd.AddCommand(cmdConfigsEdit())
	cmd.AddCommand(cmdConfigsDelete())

	return cmd
}

func cmdConfigsList() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all saved presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigsList()
		},
	}
}

func runConfigsList() error {
	layout, err := homedir.Resolve()
	if err != nil {
		return err
	}
	store := &config.Store{Dir: layout.Configs}

	names, err := store.List()
	if err != nil {
		return err
	}

	if len(names) == 0 {
		ui.Info("No presets saved. Use 'go-llama-cpp configs new' to create one.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tMODEL\tPORT\tCTX\tGPU LAYERS")
	for _, name := range names {
		cfg, err := store.Load(name)
		if err != nil {
			fmt.Fprintf(w, "%s\t<error loading>\t\t\t\n", name)
			continue
		}
		model := cfg.Model
		if model == "" {
			model = "-"
		}
		port := colInt(cfg.Port)
		ctx := colInt(cfg.CtxSize)
		gpu := colInt(cfg.NGPULayers)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, model, port, ctx, gpu)
	}
	return w.Flush()
}

func cmdConfigsShow() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show all settings for a preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			layout, err := homedir.Resolve()
			if err != nil {
				return err
			}
			store := &config.Store{Dir: layout.Configs}

			cfg, err := store.Load(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Preset: %s\n\n", args[0])
			cliArgs := config.ToArgs(cfg)
			for i := 0; i < len(cliArgs); i++ {
				if i+1 < len(cliArgs) && !strings.HasPrefix(cliArgs[i+1], "--") {
					fmt.Printf("  %s %s\n", cliArgs[i], cliArgs[i+1])
					i++
				} else {
					fmt.Printf("  %s\n", cliArgs[i])
				}
			}
			return nil
		},
	}
}

func cmdConfigsNew() *cobra.Command {
	return &cobra.Command{
		Use:   "new [name]",
		Short: "Create a new preset interactively",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			layout, err := homedir.Resolve()
			if err != nil {
				return err
			}
			store := &config.Store{Dir: layout.Configs}

			u := ui.Default
			var name string
			if len(args) == 1 {
				name = args[0]
			} else {
				name, err = u.Prompt("Preset name: ")
				if err != nil {
					return err
				}
			}
			if name == "" {
				return fmt.Errorf("preset name must not be empty")
			}

			if store.Exists(name) {
				overwrite, err := u.Confirm(fmt.Sprintf("Preset %q already exists. Overwrite?", name), false)
				if err != nil {
					return err
				}
				if !overwrite {
					return nil
				}
			}

			cfg := &config.Config{}
			if err := promptConfigFields(u, cfg); err != nil {
				return err
			}

			if err := store.Save(name, cfg); err != nil {
				return err
			}
			u.Info("Preset %q saved.", name)
			return nil
		},
	}
}

func cmdConfigsEdit() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit an existing preset interactively",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			layout, err := homedir.Resolve()
			if err != nil {
				return err
			}
			store := &config.Store{Dir: layout.Configs}

			cfg, err := store.Load(args[0])
			if err != nil {
				return err
			}

			u := ui.Default
			fmt.Println("Editing preset: " + args[0])
			fmt.Println("Press Enter to keep current value. Type - to clear an optional field.")
			fmt.Println()

			if err := promptConfigFields(u, cfg); err != nil {
				return err
			}

			if err := store.Save(args[0], cfg); err != nil {
				return err
			}
			u.Info("Preset %q updated.", args[0])
			return nil
		},
	}
}

func cmdConfigsDelete() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a saved preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			layout, err := homedir.Resolve()
			if err != nil {
				return err
			}
			store := &config.Store{Dir: layout.Configs}

			u := ui.Default
			confirm, err := u.Confirm(fmt.Sprintf("Delete preset %q?", args[0]), false)
			if err != nil {
				return err
			}
			if !confirm {
				return nil
			}

			if err := store.Delete(args[0]); err != nil {
				return err
			}
			u.Info("Preset %q deleted.", args[0])
			return nil
		},
	}
}

// promptConfigFields walks the user through the key llama-server fields,
// prompting for each with the current value as default.
// Enter keeps the current value; "-" clears an optional field.
func promptConfigFields(u *ui.UI, cfg *config.Config) error {
	var err error

	fmt.Println("── Core ─────────────────────────────────────────")
	cfg.Model, err = promptString(u, "--model", cfg.Model, true)
	if err != nil {
		return err
	}
	cfg.Host, err = promptOptStr(u, "--host", cfg.Host)
	if err != nil {
		return err
	}
	cfg.Port, err = promptOptInt(u, "--port", cfg.Port)
	if err != nil {
		return err
	}

	fmt.Println("── Context / generation ─────────────────────────")
	cfg.CtxSize, err = promptOptInt(u, "--ctx-size", cfg.CtxSize)
	if err != nil {
		return err
	}
	cfg.NPredict, err = promptOptInt(u, "--n-predict", cfg.NPredict)
	if err != nil {
		return err
	}

	fmt.Println("── GPU / offloading ──────────────────────────────")
	cfg.NGPULayers, err = promptOptInt(u, "--n-gpu-layers", cfg.NGPULayers)
	if err != nil {
		return err
	}

	fmt.Println("── Sampling ──────────────────────────────────────")
	cfg.Temp, err = promptOptFloat(u, "--temp", cfg.Temp)
	if err != nil {
		return err
	}

	fmt.Println("── CPU / threading ───────────────────────────────")
	cfg.Threads, err = promptOptInt(u, "--threads", cfg.Threads)
	if err != nil {
		return err
	}

	fmt.Println("── Attention / memory ────────────────────────────")
	cfg.FlashAttn, err = promptOptStr(u, "--flash-attn (on/off/auto)", cfg.FlashAttn)
	if err != nil {
		return err
	}
	cfg.NoMmap, err = promptOptBool(u, "--no-mmap", cfg.NoMmap)
	if err != nil {
		return err
	}

	fmt.Println("── KV cache ──────────────────────────────────────")
	cfg.CacheTypeK, err = promptOptStr(u, "--cache-type-k", cfg.CacheTypeK)
	if err != nil {
		return err
	}
	cfg.CacheTypeV, err = promptOptStr(u, "--cache-type-v", cfg.CacheTypeV)
	if err != nil {
		return err
	}

	fmt.Println("── Server / batching ─────────────────────────────")
	cfg.NParallel, err = promptOptInt(u, "--parallel", cfg.NParallel)
	if err != nil {
		return err
	}
	cfg.ContBatching, err = promptOptBool(u, "--cont-batching", cfg.ContBatching)
	if err != nil {
		return err
	}

	return nil
}

// ── prompt helpers ─────────────────────────────────────────────────────────────

// promptString prompts for a required string field. Empty input is only accepted
// if required is false.
func promptString(u *ui.UI, label, cur string, required bool) (string, error) {
	for {
		val, err := u.Prompt(fmt.Sprintf("  %s [%s]: ", label, cur))
		if err != nil {
			return "", err
		}
		if val == "" {
			if cur != "" || !required {
				return cur, nil
			}
			fmt.Println("  (required)")
			continue
		}
		return val, nil
	}
}

// promptOptStr prompts for an optional string field.
// Enter keeps current; "-" clears.
func promptOptStr(u *ui.UI, label string, cur *string) (*string, error) {
	display := ""
	if cur != nil {
		display = *cur
	}
	val, err := u.Prompt(fmt.Sprintf("  %s [%s]: ", label, display))
	if err != nil {
		return nil, err
	}
	if val == "" {
		return cur, nil
	}
	if val == "-" {
		return nil, nil
	}
	return &val, nil
}

// promptOptInt prompts for an optional int field.
// Enter keeps current; "-" clears.
func promptOptInt(u *ui.UI, label string, cur *int) (*int, error) {
	display := ""
	if cur != nil {
		display = strconv.Itoa(*cur)
	}
	for {
		val, err := u.Prompt(fmt.Sprintf("  %s [%s]: ", label, display))
		if err != nil {
			return nil, err
		}
		if val == "" {
			return cur, nil
		}
		if val == "-" {
			return nil, nil
		}
		n, err := strconv.Atoi(val)
		if err != nil {
			fmt.Printf("  (expected integer, got %q)\n", val)
			continue
		}
		return &n, nil
	}
}

// promptOptFloat prompts for an optional float64 field.
// Enter keeps current; "-" clears.
func promptOptFloat(u *ui.UI, label string, cur *float64) (*float64, error) {
	display := ""
	if cur != nil {
		display = strconv.FormatFloat(*cur, 'f', -1, 64)
	}
	for {
		val, err := u.Prompt(fmt.Sprintf("  %s [%s]: ", label, display))
		if err != nil {
			return nil, err
		}
		if val == "" {
			return cur, nil
		}
		if val == "-" {
			return nil, nil
		}
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			fmt.Printf("  (expected number, got %q)\n", val)
			continue
		}
		return &f, nil
	}
}

// promptOptBool prompts for an optional bool field (y/n).
// Enter keeps current; "-" clears.
func promptOptBool(u *ui.UI, label string, cur *bool) (*bool, error) {
	display := ""
	if cur != nil {
		if *cur {
			display = "y"
		} else {
			display = "n"
		}
	}
	for {
		val, err := u.Prompt(fmt.Sprintf("  %s (y/n) [%s]: ", label, display))
		if err != nil {
			return nil, err
		}
		if val == "" {
			return cur, nil
		}
		if val == "-" {
			return nil, nil
		}
		switch strings.ToLower(val) {
		case "y", "yes", "true", "1":
			b := true
			return &b, nil
		case "n", "no", "false", "0":
			b := false
			return &b, nil
		default:
			fmt.Printf("  (expected y or n, got %q)\n", val)
		}
	}
}

// colInt formats a *int as a string for table columns, returning "-" when nil.
func colInt(v *int) string {
	if v == nil {
		return "-"
	}
	return strconv.Itoa(*v)
}
