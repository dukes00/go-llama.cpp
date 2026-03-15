// Package wizard implements the interactive preset save/load flow.
package wizard

import (
	"fmt"

	"go-llama.cpp/internal/config"
	"go-llama.cpp/internal/ui"
)

// Wizard handles interactive config preset prompts.
type Wizard struct {
	Store *config.Store
	UI    *ui.UI
}

// PromptSavePreset prompts the user to save a new config preset.
// This is called when the user runs `serve` with flags but no `--config`.
func (w *Wizard) PromptSavePreset(modelName string, cfg *config.Config) (saved bool, presetName string, err error) {
	// Step 1: Ask if user wants to save
	confirm, err := w.UI.Confirm("Configuration is new. Save as a preset?", true)
	if err != nil {
		return false, "", err
	}
	if !confirm {
		return false, "", nil
	}

	// Step 2: Prompt for preset name
	var name string
	name, err = w.UI.Prompt(fmt.Sprintf("Preset name [%s]: ", modelName))
	if err != nil {
		return false, "", err
	}

	// Step 3: If empty, use modelName as default
	if name == "" {
		name = modelName
	}

	// Step 4: Validate preset name by attempting to save
	// If validation fails, report and re-prompt once
	maxRetries := 1
	retryCount := 0

	for retryCount <= maxRetries {
		// Check if preset already exists
		exists := w.Store.Exists(name)
		if exists {
			overwrite, err := w.UI.Confirm("Preset already exists. Overwrite?", false)
			if err != nil {
				return false, "", err
			}
			if !overwrite {
				return false, "", nil
			}
		}

		// Attempt to save - this will validate the name
		if err := w.Store.Save(name, cfg); err != nil {
			// Name validation failed, re-prompt
			if retryCount < maxRetries {
				retryCount++
				name, err = w.UI.Prompt(fmt.Sprintf("Invalid preset name. Try again [%s]: ", modelName))
				if err != nil {
					return false, "", err
				}
				if name == "" {
					name = modelName
				}
				continue
			}
			return false, "", fmt.Errorf("invalid preset name: %w", err)
		}

		// Success
		w.UI.Info("Preset '%s' saved.", name)
		return true, name, nil
	}

	return false, "", nil
}

// PromptSaveOverride prompts the user to save a modified config as a new preset.
// This is called when the user runs `serve --config X --override --temp 0.7`.
func (w *Wizard) PromptSaveOverride(originalPreset string, merged *config.Config) (saved bool, presetName string, err error) {
	// Step 1: Ask if user wants to save
	confirm, err := w.UI.Confirm("Save this configuration as a new preset?", false)
	if err != nil {
		return false, "", err
	}
	if !confirm {
		return false, "", nil
	}

	// Step 2: Suggest name with -custom suffix
	suggestedName := originalPreset + "-custom"

	// Step 3: Prompt for name with suggested default
	var name string
	name, err = w.UI.Prompt(fmt.Sprintf("Preset name [%s]: ", suggestedName))
	if err != nil {
		return false, "", err
	}

	// Step 4: If empty, use suggested name
	if name == "" {
		name = suggestedName
	}

	// Step 5: Validate preset name by attempting to save
	maxRetries := 1
	retryCount := 0

	for retryCount <= maxRetries {
		// Check if preset already exists
		if w.Store.Exists(name) {
			overwrite, err := w.UI.Confirm("Preset already exists. Overwrite?", false)
			if err != nil {
				return false, "", err
			}
			if !overwrite {
				return false, "", nil
			}
		}

		// Attempt to save - this will validate the name
		if err := w.Store.Save(name, merged); err != nil {
			// Name validation failed, re-prompt
			if retryCount < maxRetries {
				retryCount++
				name, err = w.UI.Prompt(fmt.Sprintf("Invalid preset name. Try again [%s]: ", suggestedName))
				if err != nil {
					return false, "", err
				}
				if name == "" {
					name = suggestedName
				}
				continue
			}
			return false, "", fmt.Errorf("invalid preset name: %w", err)
		}

		// Success
		w.UI.Info("Preset '%s' saved.", name)
		return true, name, nil
	}

	return false, "", nil
}

// AutoLoadPreset loads a preset by modelName if it exists.
// Called at the start of `serve` when NO `--config` flag is provided.
func (w *Wizard) AutoLoadPreset(modelName string) (*config.Config, bool) {
	if !w.Store.Exists(modelName) {
		return nil, false
	}

	cfg, err := w.Store.Load(modelName)
	if err != nil {
		return nil, false
	}

	w.UI.Info("Loading preset '%s'...", modelName)
	return cfg, true
}
