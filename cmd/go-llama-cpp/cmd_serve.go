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
	"github.com/spf13/pflag"
	"go-llama.cpp/internal/config"
	"go-llama.cpp/internal/homedir"
	"go-llama.cpp/internal/registry"
	"go-llama.cpp/internal/server"
	"go-llama.cpp/internal/ui"
	"go-llama.cpp/internal/wizard"
)

// ── Flag variables ─────────────────────────────────────────────────────────
// One var per llama-server flag, grouped by category.

var (
	// Preset management
	configFlag   string
	overrideFlag bool

	// ── CPU / threading ─────────────────────────────────────────────────────
	threadsFlag        int
	threadsBatchFlag   int
	cpuMaskFlag        string
	cpuRangeFlag       string
	cpuStrictFlag      int
	prioFlag           int
	pollFlag           int
	cpuMaskBatchFlag   string
	cpuRangeBatchFlag  string
	cpuStrictBatchFlag int
	prioBatchFlag      int

	// ── Context / generation ────────────────────────────────────────────────
	ctxSizeFlag    int
	nPredictFlag   int
	batchSizeFlag  int
	ubatchSizeFlag int
	keepFlag       int
	swaFullFlag    bool

	// ── Attention ───────────────────────────────────────────────────────────
	flashAttnFlag     bool
	verbosePromptFlag bool
	escapeFlag        bool
	perfFlag          bool

	// ── RoPE ────────────────────────────────────────────────────────────────
	ropeScalingFlag    string
	ropeScaleFlag      float64
	ropeFreqBaseFlag   float64
	ropeFreqScaleFlag  float64
	yarnOrigCtxFlag    int
	yarnExtFactorFlag  float64
	yarnAttnFactorFlag float64
	yarnBetaSlowFlag   float64
	yarnBetaFastFlag   float64

	// ── KV cache ────────────────────────────────────────────────────────────
	cacheTypeKFlag      string
	cacheTypeVFlag      string
	cacheTypeKDraftFlag string
	cacheTypeVDraftFlag string
	defragTholdFlag     float64
	kvOffloadFlag       bool
	noKVOffloadFlag     bool
	kvUnifiedFlag       bool
	noKVUnifiedFlag     bool

	// ── Memory ──────────────────────────────────────────────────────────────
	noMmapFlag     bool
	mlockFlag      bool
	repackFlag     bool
	noRepackFlag   bool
	noHostFlag     bool
	directIOFlag   bool
	noDirectIOFlag bool

	// ── GPU / offloading ────────────────────────────────────────────────────
	nGpuLayersFlag     int
	deviceFlag         string
	splitModeFlag      string
	tensorSplitFlag    string
	mainGpuFlag        int
	numaFlag           string
	overrideTensorFlag string
	cpuMoeFlag         bool
	nCpuMoeFlag        int
	opOffloadFlag      bool
	noOpOffloadFlag    bool
	fitFlag            string
	fitTargetFlag      string
	fitCtxFlag         int

	// ── Validation ──────────────────────────────────────────────────────────
	checkTensorsFlag bool
	overrideKVFlag   string

	// ── LoRA / adapters ─────────────────────────────────────────────────────
	loraFlag                 string
	loraScaledFlag           string
	controlVectorFlag        string
	controlVectorScaledFlag  string
	loraInitWithoutApplyFlag bool

	// ── Model source ────────────────────────────────────────────────────────
	modelUrlFlag    string
	dockerRepoFlag  string
	hfRepoFlag      string
	hfRepoDraftFlag string
	hfFileFlag      string
	hfRepoVFlag     string
	hfFileVFlag     string
	hfTokenFlag     string
	offlineFlag     bool

	// ── Logging ─────────────────────────────────────────────────────────────
	logDisableFlag    bool
	logFileFlag       string
	logColorsFlag     string
	verboseFlag       bool
	verbosityFlag     int
	logPrefixFlag     bool
	logTimestampsFlag bool

	// ── Sampling ────────────────────────────────────────────────────────────
	samplersFlag           string
	seedFlag               int
	samplerSeqFlag         string
	ignoreEosFlag          bool
	tempFlag               float64
	topKFlag               int
	topPFlag               float64
	minPFlag               float64
	topNSigmaFlag          float64
	xtcProbabilityFlag     float64
	xtcThresholdFlag       float64
	typicalPFlag           float64
	repeatLastNFlag        int
	repetitionPenaltyFlag  float64
	presencePenaltyFlag    float64
	frequencyPenaltyFlag   float64
	dryMultiplierFlag      float64
	dryBaseFlag            float64
	dryAllowedLengthFlag   int
	dryPenaltyLastNFlag    int
	drySequenceBreakerFlag string
	adaptiveTargetFlag     float64
	adaptiveDecayFlag      float64
	dynaTempRangeFlag      float64
	dynaTempExpFlag        float64
	mirostatFlag           int
	mirostatLrFlag         float64
	mirostatEntFlag        float64
	grammarFlag            string
	grammarFileFlag        string
	jsonSchemaFlag         string
	jsonSchemaFileFlag     string
	backendSamplingFlag    bool

	// ── Server – network ────────────────────────────────────────────────────
	portFlag       int
	hostFlag       string
	staticPathFlag string
	apiPrefixFlag  string
	aliasFlag      string
	tagsFlag       string

	// ── Server – auth / TLS ─────────────────────────────────────────────────
	apiKeyFlag      string
	apiKeyFileFlag  string
	sslKeyFileFlag  string
	sslCertFileFlag string

	// ── Server – timeouts ───────────────────────────────────────────────────
	timeoutFlag     int
	threadsHttpFlag int

	// ── Server – features ───────────────────────────────────────────────────
	embeddingsFlag bool
	rerankingFlag  bool
	metricsFlag    bool
	propsFlag      bool
	slotsFlag      bool
	noSlotsFlag    bool
	poolingFlag    string

	// ── Server – slots / batching ────────────────────────────────────────────
	parallelFlag             int
	contBatchingFlag         bool
	noContBatchingFlag       bool
	slotSavePathFlag         string
	slotPromptSimilarityFlag float64
	sleepIdleSecondsFlag     int

	// ── Server – context cache ───────────────────────────────────────────────
	cachePromptFlag            bool
	noCachePromptFlag          bool
	cacheReuseFlag             int
	ctxCheckpointsFlag         int
	checkpointEveryNTokensFlag int
	cacheRamFlag               int
	contextShiftFlag           bool
	noContextShiftFlag         bool

	// ── Server – chat / template ─────────────────────────────────────────────
	warmupFlag             bool
	noWarmupFlag           bool
	spmInfillFlag          bool
	chatTemplateFlag       string
	chatTemplateFileFlag   string
	chatTemplateKwargsFlag string
	jinjaFlag              bool
	noJinjaFlag            bool
	prefillAssistantFlag   bool
	noPrefillAssistantFlag bool

	// ── Server – reasoning ───────────────────────────────────────────────────
	reasoningFormatFlag        string
	reasoningFlag              string
	reasoningBudgetFlag        int
	reasoningBudgetMessageFlag string

	// ── Server – Web UI ──────────────────────────────────────────────────────
	webuiFlag           bool
	noWebuiFlag         bool
	webuiConfigFlag     string
	webuiConfigFileFlag string
	webuiMcpProxyFlag   bool
	noWebuiMcpProxyFlag bool

	// ── Server – multimodal ──────────────────────────────────────────────────
	mmprojFlag          string
	mmprojUrlFlag       string
	noMmprojFlag        bool
	mmprojOffloadFlag   bool
	noMmprojOffloadFlag bool
	imageMinTokensFlag  int
	imageMaxTokensFlag  int
	mediaPathFlag       string

	// ── Server – lookup cache ────────────────────────────────────────────────
	lookupCacheStaticFlag  string
	lookupCacheDynamicFlag string

	// ── Server – router ──────────────────────────────────────────────────────
	modelsDirFlag        string
	modelsPresetFlag     string
	modelsMaxFlag        int
	modelsAutoloadFlag   bool
	noModelsAutoloadFlag bool

	// ── Speculative decoding ─────────────────────────────────────────────────
	modelDraftFlag          string
	threadsDraftFlag        int
	threadsBatchDraftFlag   int
	draftMaxFlag            int
	draftMinFlag            int
	draftPMinFlag           float64
	ctxSizeDraftFlag        int
	deviceDraftFlag         string
	nGpuLayersDraftFlag     int
	overrideTensorDraftFlag string
	cpuMoeDraftFlag         bool
	nCpuMoeDraftFlag        int
	specTypeFlag            string
	specNgramSizeNFlag      int
	specNgramSizeMFlag      int
	specNgramMinHitsFlag    int

	// ── TTS / vocoder ─────────────────────────────────────────────────────────
	modelVocoderFlag      string
	ttsUseGuideTokensFlag bool
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

			configStore := &config.Store{Dir: layout.Configs}
			wz := &wizard.Wizard{Store: configStore, UI: ui.Default}

			if configFlag != "" {
				cfg, err = configStore.Load(configFlag)
				if err != nil {
					return fmt.Errorf("loading config %q: %w", configFlag, err)
				}

				if cliOverridesExist(cmd) {
					if !overrideFlag {
						ui.Warn("Ignoring CLI flags because --config is set. Use --override to apply them.")
					} else {
						overrides := buildConfigFromFlags(cmd, modelName)
						cfg = mergeConfigWithFlags(cfg, overrides)

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
				if loadedCfg, found := wz.AutoLoadPreset(modelName); found {
					cfg = loadedCfg
					if cliOverridesExist(cmd) {
						cfg = buildConfigFromFlags(cmd, modelName)
						ui.Info("Preset '%s' exists but CLI flags provided. Using CLI flags.", modelName)
					}
				} else {
					cfg = buildConfigFromFlags(cmd, modelName)

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

	// ── Preset management ────────────────────────────────────────────────────
	cmd.Flags().StringVar(&configFlag, "config", "", "Name of a saved preset")
	cmd.Flags().BoolVar(&overrideFlag, "override", false, "Allow flag overrides when using --config")

	// ── CPU / threading ──────────────────────────────────────────────────────
	cmd.Flags().IntVar(&threadsFlag, "threads", 0, "CPU threads for generation (default: -1)")
	cmd.Flags().IntVar(&threadsBatchFlag, "threads-batch", 0, "CPU threads for batch processing")
	cmd.Flags().StringVar(&cpuMaskFlag, "cpu-mask", "", "CPU affinity mask (hex)")
	cmd.Flags().StringVar(&cpuRangeFlag, "cpu-range", "", "CPU affinity range (lo-hi)")
	cmd.Flags().IntVar(&cpuStrictFlag, "cpu-strict", 0, "Strict CPU placement (0 or 1)")
	cmd.Flags().IntVar(&prioFlag, "prio", 0, "Process priority: -1=low, 0=normal, 1=medium, 2=high, 3=realtime")
	cmd.Flags().IntVar(&pollFlag, "poll", 0, "Polling level 0-100 (0=no polling)")
	cmd.Flags().StringVar(&cpuMaskBatchFlag, "cpu-mask-batch", "", "CPU affinity mask for batch")
	cmd.Flags().StringVar(&cpuRangeBatchFlag, "cpu-range-batch", "", "CPU affinity range for batch")
	cmd.Flags().IntVar(&cpuStrictBatchFlag, "cpu-strict-batch", 0, "Strict CPU placement for batch (0 or 1)")
	cmd.Flags().IntVar(&prioBatchFlag, "prio-batch", 0, "Thread priority for batch")

	// ── Context / generation ─────────────────────────────────────────────────
	cmd.Flags().IntVar(&ctxSizeFlag, "ctx-size", 0, "Context size (0=from model)")
	cmd.Flags().IntVar(&nPredictFlag, "n-predict", 0, "Tokens to predict (-1=infinity)")
	cmd.Flags().IntVar(&batchSizeFlag, "batch-size", 0, "Logical max batch size (default: 2048)")
	cmd.Flags().IntVar(&ubatchSizeFlag, "ubatch-size", 0, "Physical max batch size (default: 512)")
	cmd.Flags().IntVar(&keepFlag, "keep", 0, "Tokens to keep from initial prompt (0=none, -1=all)")
	cmd.Flags().BoolVar(&swaFullFlag, "swa-full", false, "Use full-size SWA cache")

	// ── Attention ────────────────────────────────────────────────────────────
	cmd.Flags().BoolVar(&flashAttnFlag, "flash-attn", false, "Enable flash attention")
	cmd.Flags().BoolVar(&verbosePromptFlag, "verbose-prompt", false, "Print verbose prompt before generation")
	cmd.Flags().BoolVar(&escapeFlag, "escape", false, "Process escape sequences in prompts")
	cmd.Flags().BoolVar(&perfFlag, "perf", false, "Enable internal performance timings")

	// ── RoPE ─────────────────────────────────────────────────────────────────
	cmd.Flags().StringVar(&ropeScalingFlag, "rope-scaling", "", "RoPE scaling method: none, linear, yarn")
	cmd.Flags().Float64Var(&ropeScaleFlag, "rope-scale", 0, "RoPE context scaling factor")
	cmd.Flags().Float64Var(&ropeFreqBaseFlag, "rope-freq-base", 0, "RoPE base frequency")
	cmd.Flags().Float64Var(&ropeFreqScaleFlag, "rope-freq-scale", 0, "RoPE frequency scaling factor")
	cmd.Flags().IntVar(&yarnOrigCtxFlag, "yarn-orig-ctx", 0, "YaRN: original context size")
	cmd.Flags().Float64Var(&yarnExtFactorFlag, "yarn-ext-factor", 0, "YaRN: extrapolation mix factor")
	cmd.Flags().Float64Var(&yarnAttnFactorFlag, "yarn-attn-factor", 0, "YaRN: attention scale factor")
	cmd.Flags().Float64Var(&yarnBetaSlowFlag, "yarn-beta-slow", 0, "YaRN: high correction dim")
	cmd.Flags().Float64Var(&yarnBetaFastFlag, "yarn-beta-fast", 0, "YaRN: low correction dim")

	// ── KV cache ─────────────────────────────────────────────────────────────
	cmd.Flags().StringVar(&cacheTypeKFlag, "cache-type-k", "", "K cache type: f32/f16/bf16/q8_0/q4_0/q4_1/iq4_nl/q5_0/q5_1")
	cmd.Flags().StringVar(&cacheTypeVFlag, "cache-type-v", "", "V cache type: f32/f16/bf16/q8_0/q4_0/q4_1/iq4_nl/q5_0/q5_1")
	cmd.Flags().StringVar(&cacheTypeKDraftFlag, "cache-type-k-draft", "", "K cache type for draft model")
	cmd.Flags().StringVar(&cacheTypeVDraftFlag, "cache-type-v-draft", "", "V cache type for draft model")
	cmd.Flags().Float64Var(&defragTholdFlag, "defrag-thold", 0, "KV cache defragmentation threshold")
	cmd.Flags().BoolVar(&kvOffloadFlag, "kv-offload", false, "Enable KV cache offloading")
	cmd.Flags().BoolVar(&noKVOffloadFlag, "no-kv-offload", false, "Disable KV cache offloading")
	cmd.Flags().BoolVar(&kvUnifiedFlag, "kv-unified", false, "Use unified KV buffer")
	cmd.Flags().BoolVar(&noKVUnifiedFlag, "no-kv-unified", false, "Disable unified KV buffer")

	// ── Memory ───────────────────────────────────────────────────────────────
	cmd.Flags().BoolVar(&noMmapFlag, "no-mmap", false, "Disable memory-mapped model loading")
	cmd.Flags().BoolVar(&mlockFlag, "mlock", false, "Force model to stay in RAM (no swap)")
	cmd.Flags().BoolVar(&repackFlag, "repack", false, "Enable weight repacking")
	cmd.Flags().BoolVar(&noRepackFlag, "no-repack", false, "Disable weight repacking")
	cmd.Flags().BoolVar(&noHostFlag, "no-host", false, "Bypass host buffer")
	cmd.Flags().BoolVar(&directIOFlag, "direct-io", false, "Enable DirectIO")
	cmd.Flags().BoolVar(&noDirectIOFlag, "no-direct-io", false, "Disable DirectIO")

	// ── GPU / offloading ─────────────────────────────────────────────────────
	cmd.Flags().IntVar(&nGpuLayersFlag, "n-gpu-layers", 0, "Layers to offload to GPU (auto/all or exact number)")
	cmd.Flags().StringVar(&deviceFlag, "device", "", "Comma-separated list of devices for offloading")
	cmd.Flags().StringVar(&splitModeFlag, "split-mode", "", "Multi-GPU split mode: none, layer, row")
	cmd.Flags().StringVar(&tensorSplitFlag, "tensor-split", "", "Fraction per GPU, e.g. 3,1")
	cmd.Flags().IntVar(&mainGpuFlag, "main-gpu", 0, "Primary GPU index")
	cmd.Flags().StringVar(&numaFlag, "numa", "", "NUMA strategy: distribute, isolate, numactl")
	cmd.Flags().StringVar(&overrideTensorFlag, "override-tensor", "", "Override tensor buffer type pattern=type,...")
	cmd.Flags().BoolVar(&cpuMoeFlag, "cpu-moe", false, "Keep all MoE weights on CPU")
	cmd.Flags().IntVar(&nCpuMoeFlag, "n-cpu-moe", 0, "MoE CPU layers count")
	cmd.Flags().BoolVar(&opOffloadFlag, "op-offload", false, "Offload host tensor ops to device")
	cmd.Flags().BoolVar(&noOpOffloadFlag, "no-op-offload", false, "Disable op offload")
	cmd.Flags().StringVar(&fitFlag, "fit", "", "Adjust args to fit device memory: on, off")
	cmd.Flags().StringVar(&fitTargetFlag, "fit-target", "", "Target margin per device for --fit (MiB, comma-separated)")
	cmd.Flags().IntVar(&fitCtxFlag, "fit-ctx", 0, "Minimum ctx size for --fit")

	// ── Validation ───────────────────────────────────────────────────────────
	cmd.Flags().BoolVar(&checkTensorsFlag, "check-tensors", false, "Check model tensor data for invalid values")
	cmd.Flags().StringVar(&overrideKVFlag, "override-kv", "", "Override model metadata, e.g. key=type:value,...")

	// ── LoRA / adapters ──────────────────────────────────────────────────────
	cmd.Flags().StringVar(&loraFlag, "lora", "", "LoRA adapter path(s), comma-separated")
	cmd.Flags().StringVar(&loraScaledFlag, "lora-scaled", "", "LoRA adapter with scale: FNAME:SCALE,...")
	cmd.Flags().StringVar(&controlVectorFlag, "control-vector", "", "Control vector path(s), comma-separated")
	cmd.Flags().StringVar(&controlVectorScaledFlag, "control-vector-scaled", "", "Control vector with scale: FNAME:SCALE,...")
	cmd.Flags().BoolVar(&loraInitWithoutApplyFlag, "lora-init-without-apply", false, "Load LoRA without applying (apply via POST /lora-adapters)")

	// ── Model source ─────────────────────────────────────────────────────────
	cmd.Flags().StringVar(&modelUrlFlag, "model-url", "", "Model download URL")
	cmd.Flags().StringVar(&dockerRepoFlag, "docker-repo", "", "Docker Hub model repository")
	cmd.Flags().StringVar(&hfRepoFlag, "hf-repo", "", "Hugging Face repo: user/model[:quant]")
	cmd.Flags().StringVar(&hfRepoDraftFlag, "hf-repo-draft", "", "HF repo for draft model")
	cmd.Flags().StringVar(&hfFileFlag, "hf-file", "", "Hugging Face model file override")
	cmd.Flags().StringVar(&hfRepoVFlag, "hf-repo-v", "", "HF repo for vocoder model")
	cmd.Flags().StringVar(&hfFileVFlag, "hf-file-v", "", "HF file for vocoder model")
	cmd.Flags().StringVar(&hfTokenFlag, "hf-token", "", "Hugging Face access token")
	cmd.Flags().BoolVar(&offlineFlag, "offline", false, "Offline mode: force cache, no network")

	// ── Logging ──────────────────────────────────────────────────────────────
	cmd.Flags().BoolVar(&logDisableFlag, "log-disable", false, "Disable logging")
	cmd.Flags().StringVar(&logFileFlag, "log-file", "", "Log to file")
	cmd.Flags().StringVar(&logColorsFlag, "log-colors", "", "Colored logging: on, off, auto")
	cmd.Flags().BoolVar(&verboseFlag, "verbose", false, "Set verbosity to infinity (log all)")
	cmd.Flags().IntVar(&verbosityFlag, "verbosity", 0, "Verbosity threshold (0=generic, 1=error, 2=warn, 3=info, 4=debug)")
	cmd.Flags().BoolVar(&logPrefixFlag, "log-prefix", false, "Enable prefix in log messages")
	cmd.Flags().BoolVar(&logTimestampsFlag, "log-timestamps", false, "Enable timestamps in log messages")

	// ── Sampling ─────────────────────────────────────────────────────────────
	cmd.Flags().StringVar(&samplersFlag, "samplers", "", "Sampler chain, semicolon-separated")
	cmd.Flags().IntVar(&seedFlag, "seed", 0, "RNG seed (-1=random)")
	cmd.Flags().StringVar(&samplerSeqFlag, "sampler-seq", "", "Simplified sampler sequence")
	cmd.Flags().BoolVar(&ignoreEosFlag, "ignore-eos", false, "Ignore end-of-stream token")
	cmd.Flags().Float64Var(&tempFlag, "temp", 0, "Temperature (default: 0.80)")
	cmd.Flags().IntVar(&topKFlag, "top-k", 0, "Top-K sampling (0=disabled)")
	cmd.Flags().Float64Var(&topPFlag, "top-p", 0, "Top-P sampling (1.0=disabled)")
	cmd.Flags().Float64Var(&minPFlag, "min-p", 0, "Min-P sampling (0.0=disabled)")
	cmd.Flags().Float64Var(&topNSigmaFlag, "top-n-sigma", 0, "Top-N-Sigma sampling (-1.0=disabled)")
	cmd.Flags().Float64Var(&xtcProbabilityFlag, "xtc-probability", 0, "XTC probability (0.0=disabled)")
	cmd.Flags().Float64Var(&xtcThresholdFlag, "xtc-threshold", 0, "XTC threshold (1.0=disabled)")
	cmd.Flags().Float64Var(&typicalPFlag, "typical", 0, "Locally typical sampling P (1.0=disabled)")
	cmd.Flags().IntVar(&repeatLastNFlag, "repeat-last-n", 0, "Last N tokens for repeat penalty (0=disabled, -1=ctx)")
	cmd.Flags().Float64Var(&repetitionPenaltyFlag, "repetition-penalty", 0, "Repeat penalty (1.0=disabled)")
	cmd.Flags().Float64Var(&presencePenaltyFlag, "presence-penalty", 0, "Presence penalty (0.0=disabled)")
	cmd.Flags().Float64Var(&frequencyPenaltyFlag, "frequency-penalty", 0, "Frequency penalty (0.0=disabled)")
	cmd.Flags().Float64Var(&dryMultiplierFlag, "dry-multiplier", 0, "DRY sampling multiplier (0.0=disabled)")
	cmd.Flags().Float64Var(&dryBaseFlag, "dry-base", 0, "DRY base value (default: 1.75)")
	cmd.Flags().IntVar(&dryAllowedLengthFlag, "dry-allowed-length", 0, "DRY allowed length (default: 2)")
	cmd.Flags().IntVar(&dryPenaltyLastNFlag, "dry-penalty-last-n", 0, "DRY penalty for last N tokens (-1=ctx)")
	cmd.Flags().StringVar(&drySequenceBreakerFlag, "dry-sequence-breaker", "", "DRY sequence breaker string")
	cmd.Flags().Float64Var(&adaptiveTargetFlag, "adaptive-target", 0, "Adaptive-P target probability (-1=disabled)")
	cmd.Flags().Float64Var(&adaptiveDecayFlag, "adaptive-decay", 0, "Adaptive-P decay rate (default: 0.90)")
	cmd.Flags().Float64Var(&dynaTempRangeFlag, "dynatemp-range", 0, "Dynamic temperature range (0.0=disabled)")
	cmd.Flags().Float64Var(&dynaTempExpFlag, "dynatemp-exp", 0, "Dynamic temperature exponent (default: 1.0)")
	cmd.Flags().IntVar(&mirostatFlag, "mirostat", 0, "Mirostat sampling (0=disabled, 1=v1, 2=v2)")
	cmd.Flags().Float64Var(&mirostatLrFlag, "mirostat-lr", 0, "Mirostat learning rate eta (default: 0.1)")
	cmd.Flags().Float64Var(&mirostatEntFlag, "mirostat-ent", 0, "Mirostat target entropy tau (default: 5.0)")
	cmd.Flags().StringVar(&grammarFlag, "grammar", "", "BNF-like grammar to constrain output")
	cmd.Flags().StringVar(&grammarFileFlag, "grammar-file", "", "File with grammar definition")
	cmd.Flags().StringVar(&jsonSchemaFlag, "json-schema", "", "JSON schema to constrain output")
	cmd.Flags().StringVar(&jsonSchemaFileFlag, "json-schema-file", "", "File with JSON schema")
	cmd.Flags().BoolVar(&backendSamplingFlag, "backend-sampling", false, "Enable backend sampling (experimental)")

	// ── Server – network ─────────────────────────────────────────────────────
	cmd.Flags().IntVar(&portFlag, "port", 0, "Port to listen on (default: 8080)")
	cmd.Flags().StringVar(&hostFlag, "host", "", "IP address to listen on (default: 127.0.0.1)")
	cmd.Flags().StringVar(&staticPathFlag, "static-path", "", "Path to serve static files from")
	cmd.Flags().StringVar(&apiPrefixFlag, "api-prefix", "", "API prefix path (without trailing slash)")
	cmd.Flags().StringVar(&aliasFlag, "alias", "", "Model name aliases, comma-separated")
	cmd.Flags().StringVar(&tagsFlag, "tags", "", "Model tags, comma-separated")

	// ── Server – auth / TLS ──────────────────────────────────────────────────
	cmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "API key(s), comma-separated")
	cmd.Flags().StringVar(&apiKeyFileFlag, "api-key-file", "", "File containing API keys")
	cmd.Flags().StringVar(&sslKeyFileFlag, "ssl-key-file", "", "PEM-encoded SSL private key file")
	cmd.Flags().StringVar(&sslCertFileFlag, "ssl-cert-file", "", "PEM-encoded SSL certificate file")

	// ── Server – timeouts ────────────────────────────────────────────────────
	cmd.Flags().IntVar(&timeoutFlag, "timeout", 0, "Server read/write timeout in seconds (default: 600)")
	cmd.Flags().IntVar(&threadsHttpFlag, "threads-http", 0, "Threads for HTTP request processing")

	// ── Server – features ────────────────────────────────────────────────────
	cmd.Flags().BoolVar(&embeddingsFlag, "embeddings", false, "Restrict to embedding use case only")
	cmd.Flags().BoolVar(&rerankingFlag, "reranking", false, "Enable reranking endpoint")
	cmd.Flags().BoolVar(&metricsFlag, "metrics", false, "Enable Prometheus metrics endpoint")
	cmd.Flags().BoolVar(&propsFlag, "props", false, "Enable POST /props endpoint")
	cmd.Flags().BoolVar(&slotsFlag, "slots", false, "Enable slots monitoring endpoint")
	cmd.Flags().BoolVar(&noSlotsFlag, "no-slots", false, "Disable slots monitoring endpoint")
	cmd.Flags().StringVar(&poolingFlag, "pooling", "", "Pooling type for embeddings: none, mean, cls, last, rank")

	// ── Server – slots / batching ─────────────────────────────────────────────
	cmd.Flags().IntVar(&parallelFlag, "parallel", 0, "Number of server slots (default: auto)")
	cmd.Flags().BoolVar(&contBatchingFlag, "cont-batching", false, "Enable continuous batching")
	cmd.Flags().BoolVar(&noContBatchingFlag, "no-cont-batching", false, "Disable continuous batching")
	cmd.Flags().StringVar(&slotSavePathFlag, "slot-save-path", "", "Path to save slot KV cache")
	cmd.Flags().Float64Var(&slotPromptSimilarityFlag, "slot-prompt-similarity", 0, "Min prompt similarity for slot reuse (default: 0.10)")
	cmd.Flags().IntVar(&sleepIdleSecondsFlag, "sleep-idle-seconds", 0, "Seconds idle before server sleeps (-1=disabled)")

	// ── Server – context cache ────────────────────────────────────────────────
	cmd.Flags().BoolVar(&cachePromptFlag, "cache-prompt", false, "Enable prompt caching")
	cmd.Flags().BoolVar(&noCachePromptFlag, "no-cache-prompt", false, "Disable prompt caching")
	cmd.Flags().IntVar(&cacheReuseFlag, "cache-reuse", 0, "Min chunk size for KV cache reuse")
	cmd.Flags().IntVar(&ctxCheckpointsFlag, "ctx-checkpoints", 0, "Max context checkpoints per slot (default: 32)")
	cmd.Flags().IntVar(&checkpointEveryNTokensFlag, "checkpoint-every-n-tokens", 0, "Create checkpoint every N tokens during prefill")
	cmd.Flags().IntVar(&cacheRamFlag, "cache-ram", 0, "Max cache size in MiB (-1=no limit, 0=disable)")
	cmd.Flags().BoolVar(&contextShiftFlag, "context-shift", false, "Enable context shift for infinite generation")
	cmd.Flags().BoolVar(&noContextShiftFlag, "no-context-shift", false, "Disable context shift")

	// ── Server – chat / template ──────────────────────────────────────────────
	cmd.Flags().BoolVar(&warmupFlag, "warmup", false, "Enable warmup run")
	cmd.Flags().BoolVar(&noWarmupFlag, "no-warmup", false, "Disable warmup run")
	cmd.Flags().BoolVar(&spmInfillFlag, "spm-infill", false, "Use Suffix/Prefix/Middle pattern for infill")
	cmd.Flags().StringVar(&chatTemplateFlag, "chat-template", "", "Custom Jinja chat template")
	cmd.Flags().StringVar(&chatTemplateFileFlag, "chat-template-file", "", "File with custom Jinja chat template")
	cmd.Flags().StringVar(&chatTemplateKwargsFlag, "chat-template-kwargs", "", "JSON params for template parser")
	cmd.Flags().BoolVar(&jinjaFlag, "jinja", false, "Use Jinja template engine for chat")
	cmd.Flags().BoolVar(&noJinjaFlag, "no-jinja", false, "Disable Jinja template engine")
	cmd.Flags().BoolVar(&prefillAssistantFlag, "prefill-assistant", false, "Prefill assistant response when last message is assistant")
	cmd.Flags().BoolVar(&noPrefillAssistantFlag, "no-prefill-assistant", false, "Disable prefill assistant")

	// ── Server – reasoning ────────────────────────────────────────────────────
	cmd.Flags().StringVar(&reasoningFormatFlag, "reasoning-format", "", "Reasoning format: none, deepseek, deepseek-legacy (default: auto)")
	cmd.Flags().StringVar(&reasoningFlag, "reasoning", "", "Use reasoning/thinking: on, off, auto")
	cmd.Flags().IntVar(&reasoningBudgetFlag, "reasoning-budget", 0, "Token budget for thinking (-1=unlimited, 0=immediate end, N=budget)")
	cmd.Flags().StringVar(&reasoningBudgetMessageFlag, "reasoning-budget-message", "", "Message injected before end-of-thinking tag")

	// ── Server – Web UI ───────────────────────────────────────────────────────
	cmd.Flags().BoolVar(&webuiFlag, "webui", false, "Enable Web UI")
	cmd.Flags().BoolVar(&noWebuiFlag, "no-webui", false, "Disable Web UI")
	cmd.Flags().StringVar(&webuiConfigFlag, "webui-config", "", "JSON for default WebUI settings")
	cmd.Flags().StringVar(&webuiConfigFileFlag, "webui-config-file", "", "JSON file for default WebUI settings")
	cmd.Flags().BoolVar(&webuiMcpProxyFlag, "webui-mcp-proxy", false, "Enable MCP CORS proxy (experimental)")
	cmd.Flags().BoolVar(&noWebuiMcpProxyFlag, "no-webui-mcp-proxy", false, "Disable MCP CORS proxy")

	// ── Server – multimodal ───────────────────────────────────────────────────
	cmd.Flags().StringVar(&mmprojFlag, "mmproj", "", "Path to multimodal projector file")
	cmd.Flags().StringVar(&mmprojUrlFlag, "mmproj-url", "", "URL to multimodal projector file")
	cmd.Flags().BoolVar(&noMmprojFlag, "no-mmproj", false, "Disable multimodal projector auto-load")
	cmd.Flags().BoolVar(&mmprojOffloadFlag, "mmproj-offload", false, "Enable GPU offloading for mmproj")
	cmd.Flags().BoolVar(&noMmprojOffloadFlag, "no-mmproj-offload", false, "Disable GPU offloading for mmproj")
	cmd.Flags().IntVar(&imageMinTokensFlag, "image-min-tokens", 0, "Min tokens per image for dynamic resolution")
	cmd.Flags().IntVar(&imageMaxTokensFlag, "image-max-tokens", 0, "Max tokens per image for dynamic resolution")
	cmd.Flags().StringVar(&mediaPathFlag, "media-path", "", "Directory for local media files (file:// access)")

	// ── Server – lookup cache ─────────────────────────────────────────────────
	cmd.Flags().StringVar(&lookupCacheStaticFlag, "lookup-cache-static", "", "Path to static lookup cache")
	cmd.Flags().StringVar(&lookupCacheDynamicFlag, "lookup-cache-dynamic", "", "Path to dynamic lookup cache")

	// ── Server – router ───────────────────────────────────────────────────────
	cmd.Flags().StringVar(&modelsDirFlag, "models-dir", "", "Directory containing models for router")
	cmd.Flags().StringVar(&modelsPresetFlag, "models-preset", "", "INI file with model presets for router")
	cmd.Flags().IntVar(&modelsMaxFlag, "models-max", 0, "Max simultaneous models for router (0=unlimited)")
	cmd.Flags().BoolVar(&modelsAutoloadFlag, "models-autoload", false, "Auto-load models in router")
	cmd.Flags().BoolVar(&noModelsAutoloadFlag, "no-models-autoload", false, "Disable auto-load in router")

	// ── Speculative decoding ──────────────────────────────────────────────────
	cmd.Flags().StringVar(&modelDraftFlag, "model-draft", "", "Draft model path for speculative decoding")
	cmd.Flags().IntVar(&threadsDraftFlag, "threads-draft", 0, "Threads for draft generation")
	cmd.Flags().IntVar(&threadsBatchDraftFlag, "threads-batch-draft", 0, "Threads for draft batch processing")
	cmd.Flags().IntVar(&draftMaxFlag, "draft", 0, "Max draft tokens (default: 16)")
	cmd.Flags().IntVar(&draftMinFlag, "draft-min", 0, "Min draft tokens (default: 0)")
	cmd.Flags().Float64Var(&draftPMinFlag, "draft-p-min", 0, "Min speculative decoding probability (default: 0.75)")
	cmd.Flags().IntVar(&ctxSizeDraftFlag, "ctx-size-draft", 0, "Context size for draft model (0=from model)")
	cmd.Flags().StringVar(&deviceDraftFlag, "device-draft", "", "Devices for draft model offloading")
	cmd.Flags().IntVar(&nGpuLayersDraftFlag, "gpu-layers-draft", 0, "GPU layers for draft model")
	cmd.Flags().StringVar(&overrideTensorDraftFlag, "override-tensor-draft", "", "Override tensor buffer type for draft model")
	cmd.Flags().BoolVar(&cpuMoeDraftFlag, "cpu-moe-draft", false, "Keep all MoE weights on CPU for draft model")
	cmd.Flags().IntVar(&nCpuMoeDraftFlag, "n-cpu-moe-draft", 0, "MoE CPU layers for draft model")
	cmd.Flags().StringVar(&specTypeFlag, "spec-type", "", "Speculative decoding type when no draft model")
	cmd.Flags().IntVar(&specNgramSizeNFlag, "spec-ngram-size-n", 0, "Ngram size N for speculative decoding")
	cmd.Flags().IntVar(&specNgramSizeMFlag, "spec-ngram-size-m", 0, "Ngram size M for speculative decoding")
	cmd.Flags().IntVar(&specNgramMinHitsFlag, "spec-ngram-min-hits", 0, "Min hits for ngram-map speculative decoding")

	// ── TTS / vocoder ─────────────────────────────────────────────────────────
	cmd.Flags().StringVar(&modelVocoderFlag, "model-vocoder", "", "Vocoder model path for audio generation")
	cmd.Flags().BoolVar(&ttsUseGuideTokensFlag, "tts-use-guide-tokens", false, "Use guide tokens to improve TTS word recall")

	return cmd
}

// buildConfigFromFlags creates a Config from CLI flags that were explicitly set.
func buildConfigFromFlags(cmd *cobra.Command, modelName string) *config.Config {
	cfg := &config.Config{Model: modelName}
	f := cmd.Flags()
	ch := f.Changed

	// ── CPU / threading ───────────────────────────────────────────────────────
	if ch("threads") {
		cfg.Threads = &threadsFlag
	}
	if ch("threads-batch") {
		cfg.ThreadsBatch = &threadsBatchFlag
	}
	if ch("cpu-mask") {
		cfg.CPUMask = &cpuMaskFlag
	}
	if ch("cpu-range") {
		cfg.CPURange = &cpuRangeFlag
	}
	if ch("cpu-strict") {
		cfg.CPUStrict = &cpuStrictFlag
	}
	if ch("prio") {
		cfg.Prio = &prioFlag
	}
	if ch("poll") {
		cfg.Poll = &pollFlag
	}
	if ch("cpu-mask-batch") {
		cfg.CPUMaskBatch = &cpuMaskBatchFlag
	}
	if ch("cpu-range-batch") {
		cfg.CPURangeBatch = &cpuRangeBatchFlag
	}
	if ch("cpu-strict-batch") {
		cfg.CPUStrictBatch = &cpuStrictBatchFlag
	}
	if ch("prio-batch") {
		cfg.PrioBatch = &prioBatchFlag
	}

	// ── Context / generation ──────────────────────────────────────────────────
	if ch("ctx-size") {
		cfg.CtxSize = &ctxSizeFlag
	}
	if ch("n-predict") {
		cfg.NPredict = &nPredictFlag
	}
	if ch("batch-size") {
		cfg.BatchSize = &batchSizeFlag
	}
	if ch("ubatch-size") {
		cfg.UbatchSize = &ubatchSizeFlag
	}
	if ch("keep") {
		cfg.Keep = &keepFlag
	}
	if ch("swa-full") {
		cfg.SwaFull = &swaFullFlag
	}

	// ── Attention ─────────────────────────────────────────────────────────────
	if ch("flash-attn") {
		cfg.FlashAttn = &flashAttnFlag
	}
	if ch("verbose-prompt") {
		cfg.VerbosePrompt = &verbosePromptFlag
	}
	if ch("escape") {
		cfg.Escape = &escapeFlag
	}
	if ch("perf") {
		cfg.Perf = &perfFlag
	}

	// ── RoPE ──────────────────────────────────────────────────────────────────
	if ch("rope-scaling") {
		cfg.RopeScaling = &ropeScalingFlag
	}
	if ch("rope-scale") {
		cfg.RopeScale = &ropeScaleFlag
	}
	if ch("rope-freq-base") {
		cfg.RopeFreqBase = &ropeFreqBaseFlag
	}
	if ch("rope-freq-scale") {
		cfg.RopeFreqScale = &ropeFreqScaleFlag
	}
	if ch("yarn-orig-ctx") {
		cfg.YarnOrigCtx = &yarnOrigCtxFlag
	}
	if ch("yarn-ext-factor") {
		cfg.YarnExtFactor = &yarnExtFactorFlag
	}
	if ch("yarn-attn-factor") {
		cfg.YarnAttnFactor = &yarnAttnFactorFlag
	}
	if ch("yarn-beta-slow") {
		cfg.YarnBetaSlow = &yarnBetaSlowFlag
	}
	if ch("yarn-beta-fast") {
		cfg.YarnBetaFast = &yarnBetaFastFlag
	}

	// ── KV cache ──────────────────────────────────────────────────────────────
	if ch("cache-type-k") {
		cfg.CacheTypeK = &cacheTypeKFlag
	}
	if ch("cache-type-v") {
		cfg.CacheTypeV = &cacheTypeVFlag
	}
	if ch("cache-type-k-draft") {
		cfg.CacheTypeKDraft = &cacheTypeKDraftFlag
	}
	if ch("cache-type-v-draft") {
		cfg.CacheTypeVDraft = &cacheTypeVDraftFlag
	}
	if ch("defrag-thold") {
		cfg.DefragThold = &defragTholdFlag
	}
	if ch("kv-offload") {
		cfg.KVOffload = config.Ptr(true)
	}
	if ch("no-kv-offload") {
		cfg.KVOffload = config.Ptr(false)
	}
	if ch("kv-unified") {
		cfg.KVUnified = config.Ptr(true)
	}
	if ch("no-kv-unified") {
		cfg.KVUnified = config.Ptr(false)
	}

	// ── Memory ────────────────────────────────────────────────────────────────
	if ch("no-mmap") {
		cfg.NoMmap = &noMmapFlag
	}
	if ch("mlock") {
		cfg.Mlock = &mlockFlag
	}
	if ch("repack") {
		cfg.Repack = config.Ptr(true)
	}
	if ch("no-repack") {
		cfg.Repack = config.Ptr(false)
	}
	if ch("no-host") {
		cfg.NoHost = &noHostFlag
	}
	if ch("direct-io") {
		cfg.DirectIO = config.Ptr(true)
	}
	if ch("no-direct-io") {
		cfg.DirectIO = config.Ptr(false)
	}

	// ── GPU / offloading ──────────────────────────────────────────────────────
	if ch("n-gpu-layers") {
		cfg.NGPULayers = &nGpuLayersFlag
	}
	if ch("device") {
		cfg.Device = &deviceFlag
	}
	if ch("split-mode") {
		cfg.SplitMode = &splitModeFlag
	}
	if ch("tensor-split") {
		cfg.TensorSplit = &tensorSplitFlag
	}
	if ch("main-gpu") {
		cfg.MainGpu = &mainGpuFlag
	}
	if ch("numa") {
		cfg.Numa = &numaFlag
	}
	if ch("override-tensor") {
		cfg.OverrideTensor = &overrideTensorFlag
	}
	if ch("cpu-moe") {
		cfg.CPUMoe = &cpuMoeFlag
	}
	if ch("n-cpu-moe") {
		cfg.NCpuMoe = &nCpuMoeFlag
	}
	if ch("op-offload") {
		cfg.OpOffload = config.Ptr(true)
	}
	if ch("no-op-offload") {
		cfg.OpOffload = config.Ptr(false)
	}
	if ch("fit") {
		cfg.Fit = &fitFlag
	}
	if ch("fit-target") {
		cfg.FitTarget = &fitTargetFlag
	}
	if ch("fit-ctx") {
		cfg.FitCtx = &fitCtxFlag
	}

	// ── Validation ────────────────────────────────────────────────────────────
	if ch("check-tensors") {
		cfg.CheckTensors = &checkTensorsFlag
	}
	if ch("override-kv") {
		cfg.OverrideKV = &overrideKVFlag
	}

	// ── LoRA / adapters ───────────────────────────────────────────────────────
	if ch("lora") {
		cfg.Lora = &loraFlag
	}
	if ch("lora-scaled") {
		cfg.LoraScaled = &loraScaledFlag
	}
	if ch("control-vector") {
		cfg.ControlVector = &controlVectorFlag
	}
	if ch("control-vector-scaled") {
		cfg.ControlVectorScaled = &controlVectorScaledFlag
	}
	if ch("lora-init-without-apply") {
		cfg.LoraInitWithoutApply = &loraInitWithoutApplyFlag
	}

	// ── Model source ──────────────────────────────────────────────────────────
	if ch("model-url") {
		cfg.ModelUrl = &modelUrlFlag
	}
	if ch("docker-repo") {
		cfg.DockerRepo = &dockerRepoFlag
	}
	if ch("hf-repo") {
		cfg.HfRepo = &hfRepoFlag
	}
	if ch("hf-repo-draft") {
		cfg.HfRepoDraft = &hfRepoDraftFlag
	}
	if ch("hf-file") {
		cfg.HfFile = &hfFileFlag
	}
	if ch("hf-repo-v") {
		cfg.HfRepoV = &hfRepoVFlag
	}
	if ch("hf-file-v") {
		cfg.HfFileV = &hfFileVFlag
	}
	if ch("hf-token") {
		cfg.HfToken = &hfTokenFlag
	}
	if ch("offline") {
		cfg.Offline = &offlineFlag
	}

	// ── Logging ───────────────────────────────────────────────────────────────
	if ch("log-disable") {
		cfg.LogDisable = &logDisableFlag
	}
	if ch("log-file") {
		cfg.LogFile = &logFileFlag
	}
	if ch("log-colors") {
		cfg.LogColors = &logColorsFlag
	}
	if ch("verbose") {
		cfg.Verbose = &verboseFlag
	}
	if ch("verbosity") {
		cfg.Verbosity = &verbosityFlag
	}
	if ch("log-prefix") {
		cfg.LogPrefix = &logPrefixFlag
	}
	if ch("log-timestamps") {
		cfg.LogTimestamps = &logTimestampsFlag
	}

	// ── Sampling ──────────────────────────────────────────────────────────────
	if ch("samplers") {
		cfg.Samplers = &samplersFlag
	}
	if ch("seed") {
		cfg.Seed = &seedFlag
	}
	if ch("sampler-seq") {
		cfg.SamplerSeq = &samplerSeqFlag
	}
	if ch("ignore-eos") {
		cfg.IgnoreEos = &ignoreEosFlag
	}
	if ch("temp") {
		cfg.Temp = &tempFlag
	}
	if ch("top-k") {
		cfg.TopK = &topKFlag
	}
	if ch("top-p") {
		cfg.TopP = &topPFlag
	}
	if ch("min-p") {
		cfg.MinP = &minPFlag
	}
	if ch("top-n-sigma") {
		cfg.TopNSigma = &topNSigmaFlag
	}
	if ch("xtc-probability") {
		cfg.XtcProbability = &xtcProbabilityFlag
	}
	if ch("xtc-threshold") {
		cfg.XtcThreshold = &xtcThresholdFlag
	}
	if ch("typical") {
		cfg.TypicalP = &typicalPFlag
	}
	if ch("repeat-last-n") {
		cfg.RepeatLastN = &repeatLastNFlag
	}
	if ch("repetition-penalty") {
		cfg.RepetitionPenalty = &repetitionPenaltyFlag
	}
	if ch("presence-penalty") {
		cfg.PresencePenalty = &presencePenaltyFlag
	}
	if ch("frequency-penalty") {
		cfg.FrequencyPenalty = &frequencyPenaltyFlag
	}
	if ch("dry-multiplier") {
		cfg.DryMultiplier = &dryMultiplierFlag
	}
	if ch("dry-base") {
		cfg.DryBase = &dryBaseFlag
	}
	if ch("dry-allowed-length") {
		cfg.DryAllowedLength = &dryAllowedLengthFlag
	}
	if ch("dry-penalty-last-n") {
		cfg.DryPenaltyLastN = &dryPenaltyLastNFlag
	}
	if ch("dry-sequence-breaker") {
		cfg.DrySequenceBreaker = &drySequenceBreakerFlag
	}
	if ch("adaptive-target") {
		cfg.AdaptiveTarget = &adaptiveTargetFlag
	}
	if ch("adaptive-decay") {
		cfg.AdaptiveDecay = &adaptiveDecayFlag
	}
	if ch("dynatemp-range") {
		cfg.DynaTempRange = &dynaTempRangeFlag
	}
	if ch("dynatemp-exp") {
		cfg.DynaTempExp = &dynaTempExpFlag
	}
	if ch("mirostat") {
		cfg.Mirostat = &mirostatFlag
	}
	if ch("mirostat-lr") {
		cfg.MirostatLr = &mirostatLrFlag
	}
	if ch("mirostat-ent") {
		cfg.MirostatEnt = &mirostatEntFlag
	}
	if ch("grammar") {
		cfg.Grammar = &grammarFlag
	}
	if ch("grammar-file") {
		cfg.GrammarFile = &grammarFileFlag
	}
	if ch("json-schema") {
		cfg.JsonSchema = &jsonSchemaFlag
	}
	if ch("json-schema-file") {
		cfg.JsonSchemaFile = &jsonSchemaFileFlag
	}
	if ch("backend-sampling") {
		cfg.BackendSampling = &backendSamplingFlag
	}

	// ── Server – network ──────────────────────────────────────────────────────
	if ch("port") {
		cfg.Port = &portFlag
	}
	if ch("host") {
		cfg.Host = &hostFlag
	}
	if ch("static-path") {
		cfg.StaticPath = &staticPathFlag
	}
	if ch("api-prefix") {
		cfg.ApiPrefix = &apiPrefixFlag
	}
	if ch("alias") {
		cfg.Alias = &aliasFlag
	}
	if ch("tags") {
		cfg.Tags = &tagsFlag
	}

	// ── Server – auth / TLS ───────────────────────────────────────────────────
	if ch("api-key") {
		cfg.ApiKey = &apiKeyFlag
	}
	if ch("api-key-file") {
		cfg.ApiKeyFile = &apiKeyFileFlag
	}
	if ch("ssl-key-file") {
		cfg.SslKeyFile = &sslKeyFileFlag
	}
	if ch("ssl-cert-file") {
		cfg.SslCertFile = &sslCertFileFlag
	}

	// ── Server – timeouts ─────────────────────────────────────────────────────
	if ch("timeout") {
		cfg.Timeout = &timeoutFlag
	}
	if ch("threads-http") {
		cfg.ThreadsHttp = &threadsHttpFlag
	}

	// ── Server – features ─────────────────────────────────────────────────────
	if ch("embeddings") {
		cfg.Embeddings = &embeddingsFlag
	}
	if ch("reranking") {
		cfg.Reranking = &rerankingFlag
	}
	if ch("metrics") {
		cfg.Metrics = &metricsFlag
	}
	if ch("props") {
		cfg.Props = &propsFlag
	}
	if ch("slots") {
		cfg.Slots = config.Ptr(true)
	}
	if ch("no-slots") {
		cfg.Slots = config.Ptr(false)
	}
	if ch("pooling") {
		cfg.Pooling = &poolingFlag
	}

	// ── Server – slots / batching ──────────────────────────────────────────────
	if ch("parallel") {
		cfg.NParallel = &parallelFlag
	}
	if ch("cont-batching") {
		cfg.ContBatching = config.Ptr(true)
	}
	if ch("no-cont-batching") {
		cfg.ContBatching = config.Ptr(false)
	}
	if ch("slot-save-path") {
		cfg.SlotSavePath = &slotSavePathFlag
	}
	if ch("slot-prompt-similarity") {
		cfg.SlotPromptSimilarity = &slotPromptSimilarityFlag
	}
	if ch("sleep-idle-seconds") {
		cfg.SleepIdleSeconds = &sleepIdleSecondsFlag
	}

	// ── Server – context cache ─────────────────────────────────────────────────
	if ch("cache-prompt") {
		cfg.CachePrompt = config.Ptr(true)
	}
	if ch("no-cache-prompt") {
		cfg.CachePrompt = config.Ptr(false)
	}
	if ch("cache-reuse") {
		cfg.CacheReuse = &cacheReuseFlag
	}
	if ch("ctx-checkpoints") {
		cfg.CtxCheckpoints = &ctxCheckpointsFlag
	}
	if ch("checkpoint-every-n-tokens") {
		cfg.CheckpointEveryNTokens = &checkpointEveryNTokensFlag
	}
	if ch("cache-ram") {
		cfg.CacheRam = &cacheRamFlag
	}
	if ch("context-shift") {
		cfg.ContextShift = config.Ptr(true)
	}
	if ch("no-context-shift") {
		cfg.ContextShift = config.Ptr(false)
	}

	// ── Server – chat / template ───────────────────────────────────────────────
	if ch("warmup") {
		cfg.Warmup = config.Ptr(true)
	}
	if ch("no-warmup") {
		cfg.Warmup = config.Ptr(false)
	}
	if ch("spm-infill") {
		cfg.SpmInfill = &spmInfillFlag
	}
	if ch("chat-template") {
		cfg.ChatTemplate = &chatTemplateFlag
	}
	if ch("chat-template-file") {
		cfg.ChatTemplateFile = &chatTemplateFileFlag
	}
	if ch("chat-template-kwargs") {
		cfg.ChatTemplateKwargs = &chatTemplateKwargsFlag
	}
	if ch("jinja") {
		cfg.Jinja = config.Ptr(true)
	}
	if ch("no-jinja") {
		cfg.Jinja = config.Ptr(false)
	}
	if ch("prefill-assistant") {
		cfg.PrefillAssistant = config.Ptr(true)
	}
	if ch("no-prefill-assistant") {
		cfg.PrefillAssistant = config.Ptr(false)
	}

	// ── Server – reasoning ────────────────────────────────────────────────────
	if ch("reasoning-format") {
		cfg.ReasoningFormat = &reasoningFormatFlag
	}
	if ch("reasoning") {
		cfg.Reasoning = &reasoningFlag
	}
	if ch("reasoning-budget") {
		cfg.ReasoningBudget = &reasoningBudgetFlag
	}
	if ch("reasoning-budget-message") {
		cfg.ReasoningBudgetMessage = &reasoningBudgetMessageFlag
	}

	// ── Server – Web UI ───────────────────────────────────────────────────────
	if ch("webui") {
		cfg.WebUI = config.Ptr(true)
	}
	if ch("no-webui") {
		cfg.WebUI = config.Ptr(false)
	}
	if ch("webui-config") {
		cfg.WebUIConfig = &webuiConfigFlag
	}
	if ch("webui-config-file") {
		cfg.WebUIConfigFile = &webuiConfigFileFlag
	}
	if ch("webui-mcp-proxy") {
		cfg.WebUIMcpProxy = config.Ptr(true)
	}
	if ch("no-webui-mcp-proxy") {
		cfg.WebUIMcpProxy = config.Ptr(false)
	}

	// ── Server – multimodal ───────────────────────────────────────────────────
	if ch("mmproj") {
		cfg.Mmproj = &mmprojFlag
	}
	if ch("mmproj-url") {
		cfg.MmprojUrl = &mmprojUrlFlag
	}
	if ch("no-mmproj") {
		cfg.MmprojAuto = config.Ptr(false)
	}
	if ch("mmproj-offload") {
		cfg.MmprojOffload = config.Ptr(true)
	}
	if ch("no-mmproj-offload") {
		cfg.MmprojOffload = config.Ptr(false)
	}
	if ch("image-min-tokens") {
		cfg.ImageMinTokens = &imageMinTokensFlag
	}
	if ch("image-max-tokens") {
		cfg.ImageMaxTokens = &imageMaxTokensFlag
	}
	if ch("media-path") {
		cfg.MediaPath = &mediaPathFlag
	}

	// ── Server – lookup cache ─────────────────────────────────────────────────
	if ch("lookup-cache-static") {
		cfg.LookupCacheStatic = &lookupCacheStaticFlag
	}
	if ch("lookup-cache-dynamic") {
		cfg.LookupCacheDynamic = &lookupCacheDynamicFlag
	}

	// ── Server – router ───────────────────────────────────────────────────────
	if ch("models-dir") {
		cfg.ModelsDir = &modelsDirFlag
	}
	if ch("models-preset") {
		cfg.ModelsPreset = &modelsPresetFlag
	}
	if ch("models-max") {
		cfg.ModelsMax = &modelsMaxFlag
	}
	if ch("models-autoload") {
		cfg.ModelsAutoload = config.Ptr(true)
	}
	if ch("no-models-autoload") {
		cfg.ModelsAutoload = config.Ptr(false)
	}

	// ── Speculative decoding ──────────────────────────────────────────────────
	if ch("model-draft") {
		cfg.ModelDraft = &modelDraftFlag
	}
	if ch("threads-draft") {
		cfg.ThreadsDraft = &threadsDraftFlag
	}
	if ch("threads-batch-draft") {
		cfg.ThreadsBatchDraft = &threadsBatchDraftFlag
	}
	if ch("draft") {
		cfg.DraftMax = &draftMaxFlag
	}
	if ch("draft-min") {
		cfg.DraftMin = &draftMinFlag
	}
	if ch("draft-p-min") {
		cfg.DraftPMin = &draftPMinFlag
	}
	if ch("ctx-size-draft") {
		cfg.CtxSizeDraft = &ctxSizeDraftFlag
	}
	if ch("device-draft") {
		cfg.DeviceDraft = &deviceDraftFlag
	}
	if ch("gpu-layers-draft") {
		cfg.NGpuLayersDraft = &nGpuLayersDraftFlag
	}
	if ch("override-tensor-draft") {
		cfg.OverrideTensorDraft = &overrideTensorDraftFlag
	}
	if ch("cpu-moe-draft") {
		cfg.CPUMoeDraft = &cpuMoeDraftFlag
	}
	if ch("n-cpu-moe-draft") {
		cfg.NCpuMoeDraft = &nCpuMoeDraftFlag
	}
	if ch("spec-type") {
		cfg.SpecType = &specTypeFlag
	}
	if ch("spec-ngram-size-n") {
		cfg.SpecNgramSizeN = &specNgramSizeNFlag
	}
	if ch("spec-ngram-size-m") {
		cfg.SpecNgramSizeM = &specNgramSizeMFlag
	}
	if ch("spec-ngram-min-hits") {
		cfg.SpecNgramMinHits = &specNgramMinHitsFlag
	}

	// ── TTS / vocoder ─────────────────────────────────────────────────────────
	if ch("model-vocoder") {
		cfg.ModelVocoder = &modelVocoderFlag
	}
	if ch("tts-use-guide-tokens") {
		cfg.TtsUseGuideTokens = &ttsUseGuideTokensFlag
	}

	return cfg
}

// mergeConfigWithFlags merges non-nil fields from overrides into base.
func mergeConfigWithFlags(base *config.Config, overrides *config.Config) *config.Config {
	if overrides == nil {
		return base
	}
	if overrides.Model != "" {
		base.Model = overrides.Model
	}
	if overrides.Threads != nil {
		base.Threads = overrides.Threads
	}
	if overrides.ThreadsBatch != nil {
		base.ThreadsBatch = overrides.ThreadsBatch
	}
	if overrides.CPUMask != nil {
		base.CPUMask = overrides.CPUMask
	}
	if overrides.CPURange != nil {
		base.CPURange = overrides.CPURange
	}
	if overrides.CPUStrict != nil {
		base.CPUStrict = overrides.CPUStrict
	}
	if overrides.Prio != nil {
		base.Prio = overrides.Prio
	}
	if overrides.Poll != nil {
		base.Poll = overrides.Poll
	}
	if overrides.CPUMaskBatch != nil {
		base.CPUMaskBatch = overrides.CPUMaskBatch
	}
	if overrides.CPURangeBatch != nil {
		base.CPURangeBatch = overrides.CPURangeBatch
	}
	if overrides.CPUStrictBatch != nil {
		base.CPUStrictBatch = overrides.CPUStrictBatch
	}
	if overrides.PrioBatch != nil {
		base.PrioBatch = overrides.PrioBatch
	}
	if overrides.CtxSize != nil {
		base.CtxSize = overrides.CtxSize
	}
	if overrides.NPredict != nil {
		base.NPredict = overrides.NPredict
	}
	if overrides.BatchSize != nil {
		base.BatchSize = overrides.BatchSize
	}
	if overrides.UbatchSize != nil {
		base.UbatchSize = overrides.UbatchSize
	}
	if overrides.Keep != nil {
		base.Keep = overrides.Keep
	}
	if overrides.SwaFull != nil {
		base.SwaFull = overrides.SwaFull
	}
	if overrides.FlashAttn != nil {
		base.FlashAttn = overrides.FlashAttn
	}
	if overrides.VerbosePrompt != nil {
		base.VerbosePrompt = overrides.VerbosePrompt
	}
	if overrides.Escape != nil {
		base.Escape = overrides.Escape
	}
	if overrides.Perf != nil {
		base.Perf = overrides.Perf
	}
	if overrides.RopeScaling != nil {
		base.RopeScaling = overrides.RopeScaling
	}
	if overrides.RopeScale != nil {
		base.RopeScale = overrides.RopeScale
	}
	if overrides.RopeFreqBase != nil {
		base.RopeFreqBase = overrides.RopeFreqBase
	}
	if overrides.RopeFreqScale != nil {
		base.RopeFreqScale = overrides.RopeFreqScale
	}
	if overrides.YarnOrigCtx != nil {
		base.YarnOrigCtx = overrides.YarnOrigCtx
	}
	if overrides.YarnExtFactor != nil {
		base.YarnExtFactor = overrides.YarnExtFactor
	}
	if overrides.YarnAttnFactor != nil {
		base.YarnAttnFactor = overrides.YarnAttnFactor
	}
	if overrides.YarnBetaSlow != nil {
		base.YarnBetaSlow = overrides.YarnBetaSlow
	}
	if overrides.YarnBetaFast != nil {
		base.YarnBetaFast = overrides.YarnBetaFast
	}
	if overrides.CacheTypeK != nil {
		base.CacheTypeK = overrides.CacheTypeK
	}
	if overrides.CacheTypeV != nil {
		base.CacheTypeV = overrides.CacheTypeV
	}
	if overrides.CacheTypeKDraft != nil {
		base.CacheTypeKDraft = overrides.CacheTypeKDraft
	}
	if overrides.CacheTypeVDraft != nil {
		base.CacheTypeVDraft = overrides.CacheTypeVDraft
	}
	if overrides.DefragThold != nil {
		base.DefragThold = overrides.DefragThold
	}
	if overrides.KVOffload != nil {
		base.KVOffload = overrides.KVOffload
	}
	if overrides.KVUnified != nil {
		base.KVUnified = overrides.KVUnified
	}
	if overrides.NoMmap != nil {
		base.NoMmap = overrides.NoMmap
	}
	if overrides.Mlock != nil {
		base.Mlock = overrides.Mlock
	}
	if overrides.Repack != nil {
		base.Repack = overrides.Repack
	}
	if overrides.NoHost != nil {
		base.NoHost = overrides.NoHost
	}
	if overrides.DirectIO != nil {
		base.DirectIO = overrides.DirectIO
	}
	if overrides.NGPULayers != nil {
		base.NGPULayers = overrides.NGPULayers
	}
	if overrides.Device != nil {
		base.Device = overrides.Device
	}
	if overrides.SplitMode != nil {
		base.SplitMode = overrides.SplitMode
	}
	if overrides.TensorSplit != nil {
		base.TensorSplit = overrides.TensorSplit
	}
	if overrides.MainGpu != nil {
		base.MainGpu = overrides.MainGpu
	}
	if overrides.Numa != nil {
		base.Numa = overrides.Numa
	}
	if overrides.OverrideTensor != nil {
		base.OverrideTensor = overrides.OverrideTensor
	}
	if overrides.CPUMoe != nil {
		base.CPUMoe = overrides.CPUMoe
	}
	if overrides.NCpuMoe != nil {
		base.NCpuMoe = overrides.NCpuMoe
	}
	if overrides.OpOffload != nil {
		base.OpOffload = overrides.OpOffload
	}
	if overrides.Fit != nil {
		base.Fit = overrides.Fit
	}
	if overrides.FitTarget != nil {
		base.FitTarget = overrides.FitTarget
	}
	if overrides.FitCtx != nil {
		base.FitCtx = overrides.FitCtx
	}
	if overrides.CheckTensors != nil {
		base.CheckTensors = overrides.CheckTensors
	}
	if overrides.OverrideKV != nil {
		base.OverrideKV = overrides.OverrideKV
	}
	if overrides.Lora != nil {
		base.Lora = overrides.Lora
	}
	if overrides.LoraScaled != nil {
		base.LoraScaled = overrides.LoraScaled
	}
	if overrides.ControlVector != nil {
		base.ControlVector = overrides.ControlVector
	}
	if overrides.ControlVectorScaled != nil {
		base.ControlVectorScaled = overrides.ControlVectorScaled
	}
	if overrides.LoraInitWithoutApply != nil {
		base.LoraInitWithoutApply = overrides.LoraInitWithoutApply
	}
	if overrides.ModelUrl != nil {
		base.ModelUrl = overrides.ModelUrl
	}
	if overrides.DockerRepo != nil {
		base.DockerRepo = overrides.DockerRepo
	}
	if overrides.HfRepo != nil {
		base.HfRepo = overrides.HfRepo
	}
	if overrides.HfRepoDraft != nil {
		base.HfRepoDraft = overrides.HfRepoDraft
	}
	if overrides.HfFile != nil {
		base.HfFile = overrides.HfFile
	}
	if overrides.HfRepoV != nil {
		base.HfRepoV = overrides.HfRepoV
	}
	if overrides.HfFileV != nil {
		base.HfFileV = overrides.HfFileV
	}
	if overrides.HfToken != nil {
		base.HfToken = overrides.HfToken
	}
	if overrides.Offline != nil {
		base.Offline = overrides.Offline
	}
	if overrides.LogDisable != nil {
		base.LogDisable = overrides.LogDisable
	}
	if overrides.LogFile != nil {
		base.LogFile = overrides.LogFile
	}
	if overrides.LogColors != nil {
		base.LogColors = overrides.LogColors
	}
	if overrides.Verbose != nil {
		base.Verbose = overrides.Verbose
	}
	if overrides.Verbosity != nil {
		base.Verbosity = overrides.Verbosity
	}
	if overrides.LogPrefix != nil {
		base.LogPrefix = overrides.LogPrefix
	}
	if overrides.LogTimestamps != nil {
		base.LogTimestamps = overrides.LogTimestamps
	}
	if overrides.Samplers != nil {
		base.Samplers = overrides.Samplers
	}
	if overrides.Seed != nil {
		base.Seed = overrides.Seed
	}
	if overrides.SamplerSeq != nil {
		base.SamplerSeq = overrides.SamplerSeq
	}
	if overrides.IgnoreEos != nil {
		base.IgnoreEos = overrides.IgnoreEos
	}
	if overrides.Temp != nil {
		base.Temp = overrides.Temp
	}
	if overrides.TopK != nil {
		base.TopK = overrides.TopK
	}
	if overrides.TopP != nil {
		base.TopP = overrides.TopP
	}
	if overrides.MinP != nil {
		base.MinP = overrides.MinP
	}
	if overrides.TopNSigma != nil {
		base.TopNSigma = overrides.TopNSigma
	}
	if overrides.XtcProbability != nil {
		base.XtcProbability = overrides.XtcProbability
	}
	if overrides.XtcThreshold != nil {
		base.XtcThreshold = overrides.XtcThreshold
	}
	if overrides.TypicalP != nil {
		base.TypicalP = overrides.TypicalP
	}
	if overrides.RepeatLastN != nil {
		base.RepeatLastN = overrides.RepeatLastN
	}
	if overrides.RepetitionPenalty != nil {
		base.RepetitionPenalty = overrides.RepetitionPenalty
	}
	if overrides.PresencePenalty != nil {
		base.PresencePenalty = overrides.PresencePenalty
	}
	if overrides.FrequencyPenalty != nil {
		base.FrequencyPenalty = overrides.FrequencyPenalty
	}
	if overrides.DryMultiplier != nil {
		base.DryMultiplier = overrides.DryMultiplier
	}
	if overrides.DryBase != nil {
		base.DryBase = overrides.DryBase
	}
	if overrides.DryAllowedLength != nil {
		base.DryAllowedLength = overrides.DryAllowedLength
	}
	if overrides.DryPenaltyLastN != nil {
		base.DryPenaltyLastN = overrides.DryPenaltyLastN
	}
	if overrides.DrySequenceBreaker != nil {
		base.DrySequenceBreaker = overrides.DrySequenceBreaker
	}
	if overrides.AdaptiveTarget != nil {
		base.AdaptiveTarget = overrides.AdaptiveTarget
	}
	if overrides.AdaptiveDecay != nil {
		base.AdaptiveDecay = overrides.AdaptiveDecay
	}
	if overrides.DynaTempRange != nil {
		base.DynaTempRange = overrides.DynaTempRange
	}
	if overrides.DynaTempExp != nil {
		base.DynaTempExp = overrides.DynaTempExp
	}
	if overrides.Mirostat != nil {
		base.Mirostat = overrides.Mirostat
	}
	if overrides.MirostatLr != nil {
		base.MirostatLr = overrides.MirostatLr
	}
	if overrides.MirostatEnt != nil {
		base.MirostatEnt = overrides.MirostatEnt
	}
	if overrides.Grammar != nil {
		base.Grammar = overrides.Grammar
	}
	if overrides.GrammarFile != nil {
		base.GrammarFile = overrides.GrammarFile
	}
	if overrides.JsonSchema != nil {
		base.JsonSchema = overrides.JsonSchema
	}
	if overrides.JsonSchemaFile != nil {
		base.JsonSchemaFile = overrides.JsonSchemaFile
	}
	if overrides.BackendSampling != nil {
		base.BackendSampling = overrides.BackendSampling
	}
	if overrides.Host != nil {
		base.Host = overrides.Host
	}
	if overrides.Port != nil {
		base.Port = overrides.Port
	}
	if overrides.StaticPath != nil {
		base.StaticPath = overrides.StaticPath
	}
	if overrides.ApiPrefix != nil {
		base.ApiPrefix = overrides.ApiPrefix
	}
	if overrides.Alias != nil {
		base.Alias = overrides.Alias
	}
	if overrides.Tags != nil {
		base.Tags = overrides.Tags
	}
	if overrides.ApiKey != nil {
		base.ApiKey = overrides.ApiKey
	}
	if overrides.ApiKeyFile != nil {
		base.ApiKeyFile = overrides.ApiKeyFile
	}
	if overrides.SslKeyFile != nil {
		base.SslKeyFile = overrides.SslKeyFile
	}
	if overrides.SslCertFile != nil {
		base.SslCertFile = overrides.SslCertFile
	}
	if overrides.Timeout != nil {
		base.Timeout = overrides.Timeout
	}
	if overrides.ThreadsHttp != nil {
		base.ThreadsHttp = overrides.ThreadsHttp
	}
	if overrides.Embeddings != nil {
		base.Embeddings = overrides.Embeddings
	}
	if overrides.Reranking != nil {
		base.Reranking = overrides.Reranking
	}
	if overrides.Metrics != nil {
		base.Metrics = overrides.Metrics
	}
	if overrides.Props != nil {
		base.Props = overrides.Props
	}
	if overrides.Slots != nil {
		base.Slots = overrides.Slots
	}
	if overrides.Pooling != nil {
		base.Pooling = overrides.Pooling
	}
	if overrides.NParallel != nil {
		base.NParallel = overrides.NParallel
	}
	if overrides.ContBatching != nil {
		base.ContBatching = overrides.ContBatching
	}
	if overrides.SlotSavePath != nil {
		base.SlotSavePath = overrides.SlotSavePath
	}
	if overrides.SlotPromptSimilarity != nil {
		base.SlotPromptSimilarity = overrides.SlotPromptSimilarity
	}
	if overrides.SleepIdleSeconds != nil {
		base.SleepIdleSeconds = overrides.SleepIdleSeconds
	}
	if overrides.CachePrompt != nil {
		base.CachePrompt = overrides.CachePrompt
	}
	if overrides.CacheReuse != nil {
		base.CacheReuse = overrides.CacheReuse
	}
	if overrides.CtxCheckpoints != nil {
		base.CtxCheckpoints = overrides.CtxCheckpoints
	}
	if overrides.CheckpointEveryNTokens != nil {
		base.CheckpointEveryNTokens = overrides.CheckpointEveryNTokens
	}
	if overrides.CacheRam != nil {
		base.CacheRam = overrides.CacheRam
	}
	if overrides.ContextShift != nil {
		base.ContextShift = overrides.ContextShift
	}
	if overrides.Warmup != nil {
		base.Warmup = overrides.Warmup
	}
	if overrides.SpmInfill != nil {
		base.SpmInfill = overrides.SpmInfill
	}
	if overrides.ChatTemplate != nil {
		base.ChatTemplate = overrides.ChatTemplate
	}
	if overrides.ChatTemplateFile != nil {
		base.ChatTemplateFile = overrides.ChatTemplateFile
	}
	if overrides.ChatTemplateKwargs != nil {
		base.ChatTemplateKwargs = overrides.ChatTemplateKwargs
	}
	if overrides.Jinja != nil {
		base.Jinja = overrides.Jinja
	}
	if overrides.PrefillAssistant != nil {
		base.PrefillAssistant = overrides.PrefillAssistant
	}
	if overrides.ReasoningFormat != nil {
		base.ReasoningFormat = overrides.ReasoningFormat
	}
	if overrides.Reasoning != nil {
		base.Reasoning = overrides.Reasoning
	}
	if overrides.ReasoningBudget != nil {
		base.ReasoningBudget = overrides.ReasoningBudget
	}
	if overrides.ReasoningBudgetMessage != nil {
		base.ReasoningBudgetMessage = overrides.ReasoningBudgetMessage
	}
	if overrides.WebUI != nil {
		base.WebUI = overrides.WebUI
	}
	if overrides.WebUIConfig != nil {
		base.WebUIConfig = overrides.WebUIConfig
	}
	if overrides.WebUIConfigFile != nil {
		base.WebUIConfigFile = overrides.WebUIConfigFile
	}
	if overrides.WebUIMcpProxy != nil {
		base.WebUIMcpProxy = overrides.WebUIMcpProxy
	}
	if overrides.Mmproj != nil {
		base.Mmproj = overrides.Mmproj
	}
	if overrides.MmprojUrl != nil {
		base.MmprojUrl = overrides.MmprojUrl
	}
	if overrides.MmprojAuto != nil {
		base.MmprojAuto = overrides.MmprojAuto
	}
	if overrides.MmprojOffload != nil {
		base.MmprojOffload = overrides.MmprojOffload
	}
	if overrides.ImageMinTokens != nil {
		base.ImageMinTokens = overrides.ImageMinTokens
	}
	if overrides.ImageMaxTokens != nil {
		base.ImageMaxTokens = overrides.ImageMaxTokens
	}
	if overrides.MediaPath != nil {
		base.MediaPath = overrides.MediaPath
	}
	if overrides.LookupCacheStatic != nil {
		base.LookupCacheStatic = overrides.LookupCacheStatic
	}
	if overrides.LookupCacheDynamic != nil {
		base.LookupCacheDynamic = overrides.LookupCacheDynamic
	}
	if overrides.ModelsDir != nil {
		base.ModelsDir = overrides.ModelsDir
	}
	if overrides.ModelsPreset != nil {
		base.ModelsPreset = overrides.ModelsPreset
	}
	if overrides.ModelsMax != nil {
		base.ModelsMax = overrides.ModelsMax
	}
	if overrides.ModelsAutoload != nil {
		base.ModelsAutoload = overrides.ModelsAutoload
	}
	if overrides.ModelDraft != nil {
		base.ModelDraft = overrides.ModelDraft
	}
	if overrides.ThreadsDraft != nil {
		base.ThreadsDraft = overrides.ThreadsDraft
	}
	if overrides.ThreadsBatchDraft != nil {
		base.ThreadsBatchDraft = overrides.ThreadsBatchDraft
	}
	if overrides.DraftMax != nil {
		base.DraftMax = overrides.DraftMax
	}
	if overrides.DraftMin != nil {
		base.DraftMin = overrides.DraftMin
	}
	if overrides.DraftPMin != nil {
		base.DraftPMin = overrides.DraftPMin
	}
	if overrides.CtxSizeDraft != nil {
		base.CtxSizeDraft = overrides.CtxSizeDraft
	}
	if overrides.DeviceDraft != nil {
		base.DeviceDraft = overrides.DeviceDraft
	}
	if overrides.NGpuLayersDraft != nil {
		base.NGpuLayersDraft = overrides.NGpuLayersDraft
	}
	if overrides.OverrideTensorDraft != nil {
		base.OverrideTensorDraft = overrides.OverrideTensorDraft
	}
	if overrides.CPUMoeDraft != nil {
		base.CPUMoeDraft = overrides.CPUMoeDraft
	}
	if overrides.NCpuMoeDraft != nil {
		base.NCpuMoeDraft = overrides.NCpuMoeDraft
	}
	if overrides.SpecType != nil {
		base.SpecType = overrides.SpecType
	}
	if overrides.SpecNgramSizeN != nil {
		base.SpecNgramSizeN = overrides.SpecNgramSizeN
	}
	if overrides.SpecNgramSizeM != nil {
		base.SpecNgramSizeM = overrides.SpecNgramSizeM
	}
	if overrides.SpecNgramMinHits != nil {
		base.SpecNgramMinHits = overrides.SpecNgramMinHits
	}
	if overrides.ModelVocoder != nil {
		base.ModelVocoder = overrides.ModelVocoder
	}
	if overrides.TtsUseGuideTokens != nil {
		base.TtsUseGuideTokens = overrides.TtsUseGuideTokens
	}
	if overrides.Extra != nil {
		base.Extra = overrides.Extra
	}
	return base
}

// cliOverridesExist returns true if any llama-server flag was explicitly set.
// Uses cobra's Visit (only visits changed flags) so we never need to maintain a manual list.
func cliOverridesExist(cmd *cobra.Command) bool {
	changed := false
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Name != "config" && f.Name != "override" {
			changed = true
		}
	})
	return changed
}
