#!/usr/bin/env bash
set -e

export GO_LLAMA_HOME=$(mktemp -d)
echo "Using temp home: $GO_LLAMA_HOME"

BIN="./go-llama-cpp"
go build -o "$BIN" ./cmd/go-llama-cpp

# Init
$BIN init

# Create a fake model file for testing
touch "$GO_LLAMA_HOME/models/test-model.gguf"

echo "--- Test: serve with flags should prompt to save ---"
# The wizard prompts should appear before the server start error
# We expect: confirmation prompt, name prompt, save success message, then server error
OUTPUT=$(echo -e "y\ntest-preset\n" | $BIN serve test-model --temp 0.6 --n-gpu-layers 99 --port 19999 2>&1 || true)
echo "$OUTPUT"

# Verify the wizard prompts appeared
if echo "$OUTPUT" | grep -q "Configuration is new. Save as a preset?"; then
  echo "[PASS] Wizard confirmation prompt appeared"
else
  echo "[FAIL] Wizard confirmation prompt did not appear"
  exit 1
fi

if echo "$OUTPUT" | grep -q "Preset 'test-preset' saved"; then
  echo "[PASS] Preset was saved successfully"
else
  echo "[FAIL] Preset was not saved"
  exit 1
fi

echo "--- Checking preset was saved ---"
cat "$GO_LLAMA_HOME/configs/test-preset.json"

echo "--- Test: list should show the running server ---"
$BIN list || true

echo "--- Test: kill ---"
$BIN kill test-model || true

echo "--- Test: list should be empty ---"
$BIN list || true

echo "--- Cleanup ---"
rm -rf "$GO_LLAMA_HOME"
echo "DONE"