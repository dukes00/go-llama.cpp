// Package config manages JSON configuration presets for llama-server.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var ErrNotFound = errors.New("config not found")

// Config represents a llama-server configuration preset.
// Fields map 1:1 to llama-server CLI flags; nil means "use server default".
type Config struct {
	// Core
	Model string `json:"model"`

	// ── CPU / threading ──────────────────────────────────────────────────────
	Threads        *int    `json:"threads,omitempty"`
	ThreadsBatch   *int    `json:"threads_batch,omitempty"`
	CPUMask        *string `json:"cpu_mask,omitempty"`
	CPURange       *string `json:"cpu_range,omitempty"`
	CPUStrict      *int    `json:"cpu_strict,omitempty"`
	Prio           *int    `json:"prio,omitempty"`
	Poll           *int    `json:"poll,omitempty"`
	CPUMaskBatch   *string `json:"cpu_mask_batch,omitempty"`
	CPURangeBatch  *string `json:"cpu_range_batch,omitempty"`
	CPUStrictBatch *int    `json:"cpu_strict_batch,omitempty"`
	PrioBatch      *int    `json:"prio_batch,omitempty"`

	// ── Context / generation ─────────────────────────────────────────────────
	CtxSize    *int  `json:"ctx_size,omitempty"`
	NPredict   *int  `json:"n_predict,omitempty"`
	BatchSize  *int  `json:"batch_size,omitempty"`
	UbatchSize *int  `json:"ubatch_size,omitempty"`
	Keep       *int  `json:"keep,omitempty"`
	SwaFull    *bool `json:"swa_full,omitempty"`

	// ── Attention ────────────────────────────────────────────────────────────
	// FlashAttn: accept "on", "off", or "auto"
	FlashAttn     *string `json:"flash_attn,omitempty"`
	VerbosePrompt *bool   `json:"verbose_prompt,omitempty"`
	Escape        *bool   `json:"escape,omitempty"`
	Perf          *bool   `json:"perf,omitempty"`

	// ── RoPE ─────────────────────────────────────────────────────────────────
	RopeScaling    *string  `json:"rope_scaling,omitempty"`
	RopeScale      *float64 `json:"rope_scale,omitempty"`
	RopeFreqBase   *float64 `json:"rope_freq_base,omitempty"`
	RopeFreqScale  *float64 `json:"rope_freq_scale,omitempty"`
	YarnOrigCtx    *int     `json:"yarn_orig_ctx,omitempty"`
	YarnExtFactor  *float64 `json:"yarn_ext_factor,omitempty"`
	YarnAttnFactor *float64 `json:"yarn_attn_factor,omitempty"`
	YarnBetaSlow   *float64 `json:"yarn_beta_slow,omitempty"`
	YarnBetaFast   *float64 `json:"yarn_beta_fast,omitempty"`

	// ── KV cache ─────────────────────────────────────────────────────────────
	CacheTypeK      *string  `json:"cache_type_k,omitempty"`
	CacheTypeV      *string  `json:"cache_type_v,omitempty"`
	CacheTypeKDraft *string  `json:"cache_type_k_draft,omitempty"`
	CacheTypeVDraft *string  `json:"cache_type_v_draft,omitempty"`
	DefragThold     *float64 `json:"defrag_thold,omitempty"`
	// KVOffload: toggle; true→--kv-offload, false→--no-kv-offload
	KVOffload *bool `json:"kv_offload,omitempty"`
	// KVUnified: toggle; true→--kv-unified, false→--no-kv-unified
	KVUnified *bool `json:"kv_unified,omitempty"`

	// ── Memory ───────────────────────────────────────────────────────────────
	// NoMmap: emit --no-mmap only when true
	NoMmap *bool `json:"no_mmap,omitempty"`
	Mlock  *bool `json:"mlock,omitempty"`
	// Repack: toggle; true→--repack, false→--no-repack
	Repack   *bool `json:"repack,omitempty"`
	NoHost   *bool `json:"no_host,omitempty"`
	DirectIO *bool `json:"direct_io,omitempty"`

	// ── GPU / offloading ─────────────────────────────────────────────────────
	NGPULayers     *int    `json:"n_gpu_layers,omitempty"`
	Device         *string `json:"device,omitempty"`
	SplitMode      *string `json:"split_mode,omitempty"`
	TensorSplit    *string `json:"tensor_split,omitempty"`
	MainGpu        *int    `json:"main_gpu,omitempty"`
	Numa           *string `json:"numa,omitempty"`
	OverrideTensor *string `json:"override_tensor,omitempty"`
	CPUMoe         *bool   `json:"cpu_moe,omitempty"`
	NCpuMoe        *int    `json:"n_cpu_moe,omitempty"`
	// OpOffload: toggle; true→--op-offload, false→--no-op-offload
	OpOffload *bool   `json:"op_offload,omitempty"`
	Fit       *string `json:"fit,omitempty"`
	FitTarget *string `json:"fit_target,omitempty"`
	FitCtx    *int    `json:"fit_ctx,omitempty"`

	// ── Validation ───────────────────────────────────────────────────────────
	CheckTensors *bool   `json:"check_tensors,omitempty"`
	OverrideKV   *string `json:"override_kv,omitempty"`

	// ── LoRA / adapters ──────────────────────────────────────────────────────
	Lora                 *string `json:"lora,omitempty"`
	LoraScaled           *string `json:"lora_scaled,omitempty"`
	ControlVector        *string `json:"control_vector,omitempty"`
	ControlVectorScaled  *string `json:"control_vector_scaled,omitempty"`
	LoraInitWithoutApply *bool   `json:"lora_init_without_apply,omitempty"`

	// ── Model source ─────────────────────────────────────────────────────────
	ModelUrl    *string `json:"model_url,omitempty"`
	DockerRepo  *string `json:"docker_repo,omitempty"`
	HfRepo      *string `json:"hf_repo,omitempty"`
	HfRepoDraft *string `json:"hf_repo_draft,omitempty"`
	HfFile      *string `json:"hf_file,omitempty"`
	HfRepoV     *string `json:"hf_repo_v,omitempty"`
	HfFileV     *string `json:"hf_file_v,omitempty"`
	HfToken     *string `json:"hf_token,omitempty"`
	Offline     *bool   `json:"offline,omitempty"`

	// ── Logging ──────────────────────────────────────────────────────────────
	LogDisable    *bool   `json:"log_disable,omitempty"`
	LogFile       *string `json:"log_file,omitempty"`
	LogColors     *string `json:"log_colors,omitempty"`
	Verbose       *bool   `json:"verbose,omitempty"`
	Verbosity     *int    `json:"verbosity,omitempty"`
	LogPrefix     *bool   `json:"log_prefix,omitempty"`
	LogTimestamps *bool   `json:"log_timestamps,omitempty"`

	// ── Sampling ─────────────────────────────────────────────────────────────
	Samplers           *string  `json:"samplers,omitempty"`
	Seed               *int     `json:"seed,omitempty"`
	SamplerSeq         *string  `json:"sampler_seq,omitempty"`
	IgnoreEos          *bool    `json:"ignore_eos,omitempty"`
	Temp               *float64 `json:"temp,omitempty"`
	TopK               *int     `json:"top_k,omitempty"`
	TopP               *float64 `json:"top_p,omitempty"`
	MinP               *float64 `json:"min_p,omitempty"`
	TopNSigma          *float64 `json:"top_n_sigma,omitempty"`
	XtcProbability     *float64 `json:"xtc_probability,omitempty"`
	XtcThreshold       *float64 `json:"xtc_threshold,omitempty"`
	TypicalP           *float64 `json:"typical_p,omitempty"`
	RepeatLastN        *int     `json:"repeat_last_n,omitempty"`
	RepetitionPenalty  *float64 `json:"repetition_penalty,omitempty"`
	PresencePenalty    *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty   *float64 `json:"frequency_penalty,omitempty"`
	DryMultiplier      *float64 `json:"dry_multiplier,omitempty"`
	DryBase            *float64 `json:"dry_base,omitempty"`
	DryAllowedLength   *int     `json:"dry_allowed_length,omitempty"`
	DryPenaltyLastN    *int     `json:"dry_penalty_last_n,omitempty"`
	DrySequenceBreaker *string  `json:"dry_sequence_breaker,omitempty"`
	AdaptiveTarget     *float64 `json:"adaptive_target,omitempty"`
	AdaptiveDecay      *float64 `json:"adaptive_decay,omitempty"`
	DynaTempRange      *float64 `json:"dynatemp_range,omitempty"`
	DynaTempExp        *float64 `json:"dynatemp_exp,omitempty"`
	Mirostat           *int     `json:"mirostat,omitempty"`
	MirostatLr         *float64 `json:"mirostat_lr,omitempty"`
	MirostatEnt        *float64 `json:"mirostat_ent,omitempty"`
	Grammar            *string  `json:"grammar,omitempty"`
	GrammarFile        *string  `json:"grammar_file,omitempty"`
	JsonSchema         *string  `json:"json_schema,omitempty"`
	JsonSchemaFile     *string  `json:"json_schema_file,omitempty"`
	BackendSampling    *bool    `json:"backend_sampling,omitempty"`

	// ── Server – network ─────────────────────────────────────────────────────
	Host       *string `json:"host,omitempty"`
	Port       *int    `json:"port,omitempty"`
	StaticPath *string `json:"static_path,omitempty"`
	ApiPrefix  *string `json:"api_prefix,omitempty"`
	Alias      *string `json:"alias,omitempty"`
	Tags       *string `json:"tags,omitempty"`

	// ── Server – auth / TLS ──────────────────────────────────────────────────
	ApiKey      *string `json:"api_key,omitempty"`
	ApiKeyFile  *string `json:"api_key_file,omitempty"`
	SslKeyFile  *string `json:"ssl_key_file,omitempty"`
	SslCertFile *string `json:"ssl_cert_file,omitempty"`

	// ── Server – timeouts / threads ──────────────────────────────────────────
	Timeout     *int `json:"timeout,omitempty"`
	ThreadsHttp *int `json:"threads_http,omitempty"`

	// ── Server – features ────────────────────────────────────────────────────
	Embeddings *bool `json:"embeddings,omitempty"`
	Reranking  *bool `json:"reranking,omitempty"`
	Metrics    *bool `json:"metrics,omitempty"`
	Props      *bool `json:"props,omitempty"`
	// Slots: toggle; true→--slots, false→--no-slots
	Slots   *bool   `json:"slots,omitempty"`
	Pooling *string `json:"pooling,omitempty"`

	// ── Server – slots / batching ────────────────────────────────────────────
	NParallel *int `json:"n_parallel,omitempty"`
	// ContBatching: toggle; true→--cont-batching, false→--no-cont-batching
	ContBatching         *bool    `json:"cont_batching,omitempty"`
	SlotSavePath         *string  `json:"slot_save_path,omitempty"`
	SlotPromptSimilarity *float64 `json:"slot_prompt_similarity,omitempty"`
	SleepIdleSeconds     *int     `json:"sleep_idle_seconds,omitempty"`

	// ── Server – context cache ───────────────────────────────────────────────
	// CachePrompt: toggle; true→--cache-prompt, false→--no-cache-prompt
	CachePrompt            *bool `json:"cache_prompt,omitempty"`
	CacheReuse             *int  `json:"cache_reuse,omitempty"`
	CtxCheckpoints         *int  `json:"ctx_checkpoints,omitempty"`
	CheckpointEveryNTokens *int  `json:"checkpoint_every_n_tokens,omitempty"`
	CacheRam               *int  `json:"cache_ram,omitempty"`
	// ContextShift: toggle; true→--context-shift, false→--no-context-shift
	ContextShift *bool `json:"context_shift,omitempty"`

	// ── Server – chat / template ─────────────────────────────────────────────
	// Warmup: toggle; true→--warmup, false→--no-warmup
	Warmup             *bool   `json:"warmup,omitempty"`
	SpmInfill          *bool   `json:"spm_infill,omitempty"`
	ChatTemplate       *string `json:"chat_template,omitempty"`
	ChatTemplateFile   *string `json:"chat_template_file,omitempty"`
	ChatTemplateKwargs *string `json:"chat_template_kwargs,omitempty"`
	// Jinja: toggle; true→--jinja, false→--no-jinja
	Jinja *bool `json:"jinja,omitempty"`
	// PrefillAssistant: toggle; true→--prefill-assistant, false→--no-prefill-assistant
	PrefillAssistant *bool `json:"prefill_assistant,omitempty"`

	// ── Server – reasoning ───────────────────────────────────────────────────
	ReasoningFormat        *string `json:"reasoning_format,omitempty"`
	Reasoning              *string `json:"reasoning,omitempty"`
	ReasoningBudget        *int    `json:"reasoning_budget,omitempty"`
	ReasoningBudgetMessage *string `json:"reasoning_budget_message,omitempty"`

	// ── Server – Web UI ──────────────────────────────────────────────────────
	// WebUI: toggle; true→--webui, false→--no-webui
	WebUI           *bool   `json:"webui,omitempty"`
	WebUIConfig     *string `json:"webui_config,omitempty"`
	WebUIConfigFile *string `json:"webui_config_file,omitempty"`
	WebUIMcpProxy   *bool   `json:"webui_mcp_proxy,omitempty"`

	// ── Server – multimodal ──────────────────────────────────────────────────
	Mmproj         *string `json:"mmproj,omitempty"`
	MmprojUrl      *string `json:"mmproj_url,omitempty"`
	MmprojAuto     *bool   `json:"mmproj_auto,omitempty"`
	MmprojOffload  *bool   `json:"mmproj_offload,omitempty"`
	ImageMinTokens *int    `json:"image_min_tokens,omitempty"`
	ImageMaxTokens *int    `json:"image_max_tokens,omitempty"`
	MediaPath      *string `json:"media_path,omitempty"`

	// ── Server – lookup cache ────────────────────────────────────────────────
	LookupCacheStatic  *string `json:"lookup_cache_static,omitempty"`
	LookupCacheDynamic *string `json:"lookup_cache_dynamic,omitempty"`

	// ── Server – router ──────────────────────────────────────────────────────
	ModelsDir    *string `json:"models_dir,omitempty"`
	ModelsPreset *string `json:"models_preset,omitempty"`
	ModelsMax    *int    `json:"models_max,omitempty"`
	// ModelsAutoload: toggle; true→--models-autoload, false→--no-models-autoload
	ModelsAutoload *bool `json:"models_autoload,omitempty"`

	// ── Speculative decoding ─────────────────────────────────────────────────
	ModelDraft          *string  `json:"model_draft,omitempty"`
	ThreadsDraft        *int     `json:"threads_draft,omitempty"`
	ThreadsBatchDraft   *int     `json:"threads_batch_draft,omitempty"`
	DraftMax            *int     `json:"draft_max,omitempty"`
	DraftMin            *int     `json:"draft_min,omitempty"`
	DraftPMin           *float64 `json:"draft_p_min,omitempty"`
	CtxSizeDraft        *int     `json:"ctx_size_draft,omitempty"`
	DeviceDraft         *string  `json:"device_draft,omitempty"`
	NGpuLayersDraft     *int     `json:"n_gpu_layers_draft,omitempty"`
	OverrideTensorDraft *string  `json:"override_tensor_draft,omitempty"`
	CPUMoeDraft         *bool    `json:"cpu_moe_draft,omitempty"`
	NCpuMoeDraft        *int     `json:"n_cpu_moe_draft,omitempty"`
	SpecType            *string  `json:"spec_type,omitempty"`
	SpecNgramSizeN      *int     `json:"spec_ngram_size_n,omitempty"`
	SpecNgramSizeM      *int     `json:"spec_ngram_size_m,omitempty"`
	SpecNgramMinHits    *int     `json:"spec_ngram_min_hits,omitempty"`

	// ── TTS / vocoder ────────────────────────────────────────────────────────
	ModelVocoder      *string `json:"model_vocoder,omitempty"`
	TtsUseGuideTokens *bool   `json:"tts_use_guide_tokens,omitempty"`

	// ── Extra passthrough ────────────────────────────────────────────────────
	// Arbitrary --key value pairs appended last.
	Extra map[string]string `json:"extra,omitempty"`
}

// Ptr creates a pointer to a value.
func Ptr[T any](v T) *T {
	return &v
}

// ToArgs converts a Config to a slice of llama-server CLI arguments.
func ToArgs(c *Config) []string {
	var args []string

	// ── Core ─────────────────────────────────────────────────────────────────
	args = append(args, "--model", c.Model)
	args = appendStr(args, "host", c.Host)
	args = appendInt(args, "port", c.Port)
	args = appendInt(args, "ctx-size", c.CtxSize)
	args = appendInt(args, "n-gpu-layers", c.NGPULayers)
	args = appendFloat(args, "temp", c.Temp)
	args = appendInt(args, "threads", c.Threads)

	// FlashAttn: accept "on", "off", or "auto"
	args = appendStr(args, "flash-attn", c.FlashAttn)
	if c.NoMmap != nil && *c.NoMmap {
		args = append(args, "--no-mmap")
	}
	args = appendToggle(args, "cont-batching", c.ContBatching)
	args = appendStr(args, "cache-type-k", c.CacheTypeK)
	args = appendStr(args, "cache-type-v", c.CacheTypeV)
	args = appendInt(args, "parallel", c.NParallel)

	// ── CPU / threading ───────────────────────────────────────────────────────
	args = appendInt(args, "threads-batch", c.ThreadsBatch)
	args = appendStr(args, "cpu-mask", c.CPUMask)
	args = appendStr(args, "cpu-range", c.CPURange)
	args = appendInt(args, "cpu-strict", c.CPUStrict)
	args = appendInt(args, "prio", c.Prio)
	args = appendInt(args, "poll", c.Poll)
	args = appendStr(args, "cpu-mask-batch", c.CPUMaskBatch)
	args = appendStr(args, "cpu-range-batch", c.CPURangeBatch)
	args = appendInt(args, "cpu-strict-batch", c.CPUStrictBatch)
	args = appendInt(args, "prio-batch", c.PrioBatch)

	// ── Context / generation ──────────────────────────────────────────────────
	args = appendInt(args, "n-predict", c.NPredict)
	args = appendInt(args, "batch-size", c.BatchSize)
	args = appendInt(args, "ubatch-size", c.UbatchSize)
	args = appendInt(args, "keep", c.Keep)
	args = appendBool(args, "swa-full", c.SwaFull)

	// ── Attention ─────────────────────────────────────────────────────────────
	args = appendBool(args, "verbose-prompt", c.VerbosePrompt)
	args = appendBool(args, "escape", c.Escape)
	args = appendBool(args, "perf", c.Perf)

	// ── RoPE ──────────────────────────────────────────────────────────────────
	args = appendStr(args, "rope-scaling", c.RopeScaling)
	args = appendFloat(args, "rope-scale", c.RopeScale)
	args = appendFloat(args, "rope-freq-base", c.RopeFreqBase)
	args = appendFloat(args, "rope-freq-scale", c.RopeFreqScale)
	args = appendInt(args, "yarn-orig-ctx", c.YarnOrigCtx)
	args = appendFloat(args, "yarn-ext-factor", c.YarnExtFactor)
	args = appendFloat(args, "yarn-attn-factor", c.YarnAttnFactor)
	args = appendFloat(args, "yarn-beta-slow", c.YarnBetaSlow)
	args = appendFloat(args, "yarn-beta-fast", c.YarnBetaFast)

	// ── KV cache ──────────────────────────────────────────────────────────────
	args = appendStr(args, "cache-type-k-draft", c.CacheTypeKDraft)
	args = appendStr(args, "cache-type-v-draft", c.CacheTypeVDraft)
	args = appendFloat(args, "defrag-thold", c.DefragThold)
	args = appendToggle(args, "kv-offload", c.KVOffload)
	args = appendToggle(args, "kv-unified", c.KVUnified)

	// ── Memory ────────────────────────────────────────────────────────────────
	args = appendBool(args, "mlock", c.Mlock)
	args = appendToggle(args, "repack", c.Repack)
	args = appendBool(args, "no-host", c.NoHost)
	args = appendToggle(args, "direct-io", c.DirectIO)

	// ── GPU / offloading ──────────────────────────────────────────────────────
	args = appendStr(args, "device", c.Device)
	args = appendStr(args, "split-mode", c.SplitMode)
	args = appendStr(args, "tensor-split", c.TensorSplit)
	args = appendInt(args, "main-gpu", c.MainGpu)
	args = appendStr(args, "numa", c.Numa)
	args = appendStr(args, "override-tensor", c.OverrideTensor)
	args = appendBool(args, "cpu-moe", c.CPUMoe)
	args = appendInt(args, "n-cpu-moe", c.NCpuMoe)
	args = appendToggle(args, "op-offload", c.OpOffload)
	args = appendStr(args, "fit", c.Fit)
	args = appendStr(args, "fit-target", c.FitTarget)
	args = appendInt(args, "fit-ctx", c.FitCtx)

	// ── Validation ────────────────────────────────────────────────────────────
	args = appendBool(args, "check-tensors", c.CheckTensors)
	args = appendStr(args, "override-kv", c.OverrideKV)

	// ── LoRA / adapters ───────────────────────────────────────────────────────
	args = appendStr(args, "lora", c.Lora)
	args = appendStr(args, "lora-scaled", c.LoraScaled)
	args = appendStr(args, "control-vector", c.ControlVector)
	args = appendStr(args, "control-vector-scaled", c.ControlVectorScaled)
	args = appendBool(args, "lora-init-without-apply", c.LoraInitWithoutApply)

	// ── Model source ──────────────────────────────────────────────────────────
	args = appendStr(args, "model-url", c.ModelUrl)
	args = appendStr(args, "docker-repo", c.DockerRepo)
	args = appendStr(args, "hf-repo", c.HfRepo)
	args = appendStr(args, "hf-repo-draft", c.HfRepoDraft)
	args = appendStr(args, "hf-file", c.HfFile)
	args = appendStr(args, "hf-repo-v", c.HfRepoV)
	args = appendStr(args, "hf-file-v", c.HfFileV)
	args = appendStr(args, "hf-token", c.HfToken)
	args = appendBool(args, "offline", c.Offline)

	// ── Logging ───────────────────────────────────────────────────────────────
	args = appendBool(args, "log-disable", c.LogDisable)
	args = appendStr(args, "log-file", c.LogFile)
	args = appendStr(args, "log-colors", c.LogColors)
	args = appendBool(args, "verbose", c.Verbose)
	args = appendInt(args, "verbosity", c.Verbosity)
	args = appendBool(args, "log-prefix", c.LogPrefix)
	args = appendBool(args, "log-timestamps", c.LogTimestamps)

	// ── Sampling ──────────────────────────────────────────────────────────────
	args = appendStr(args, "samplers", c.Samplers)
	args = appendInt(args, "seed", c.Seed)
	args = appendStr(args, "sampler-seq", c.SamplerSeq)
	args = appendBool(args, "ignore-eos", c.IgnoreEos)
	args = appendInt(args, "top-k", c.TopK)
	args = appendFloat(args, "top-p", c.TopP)
	args = appendFloat(args, "min-p", c.MinP)
	args = appendFloat(args, "top-n-sigma", c.TopNSigma)
	args = appendFloat(args, "xtc-probability", c.XtcProbability)
	args = appendFloat(args, "xtc-threshold", c.XtcThreshold)
	args = appendFloat(args, "typical", c.TypicalP)
	args = appendInt(args, "repeat-last-n", c.RepeatLastN)
	args = appendFloat(args, "repeat-penalty", c.RepetitionPenalty)
	args = appendFloat(args, "presence-penalty", c.PresencePenalty)
	args = appendFloat(args, "frequency-penalty", c.FrequencyPenalty)
	args = appendFloat(args, "dry-multiplier", c.DryMultiplier)
	args = appendFloat(args, "dry-base", c.DryBase)
	args = appendInt(args, "dry-allowed-length", c.DryAllowedLength)
	args = appendInt(args, "dry-penalty-last-n", c.DryPenaltyLastN)
	args = appendStr(args, "dry-sequence-breaker", c.DrySequenceBreaker)
	args = appendFloat(args, "adaptive-target", c.AdaptiveTarget)
	args = appendFloat(args, "adaptive-decay", c.AdaptiveDecay)
	args = appendFloat(args, "dynatemp-range", c.DynaTempRange)
	args = appendFloat(args, "dynatemp-exp", c.DynaTempExp)
	args = appendInt(args, "mirostat", c.Mirostat)
	args = appendFloat(args, "mirostat-lr", c.MirostatLr)
	args = appendFloat(args, "mirostat-ent", c.MirostatEnt)
	args = appendStr(args, "grammar", c.Grammar)
	args = appendStr(args, "grammar-file", c.GrammarFile)
	args = appendStr(args, "json-schema", c.JsonSchema)
	args = appendStr(args, "json-schema-file", c.JsonSchemaFile)
	args = appendBool(args, "backend-sampling", c.BackendSampling)

	// ── Server – network ──────────────────────────────────────────────────────
	args = appendStr(args, "path", c.StaticPath)
	args = appendStr(args, "api-prefix", c.ApiPrefix)
	args = appendStr(args, "alias", c.Alias)
	args = appendStr(args, "tags", c.Tags)

	// ── Server – auth / TLS ───────────────────────────────────────────────────
	args = appendStr(args, "api-key", c.ApiKey)
	args = appendStr(args, "api-key-file", c.ApiKeyFile)
	args = appendStr(args, "ssl-key-file", c.SslKeyFile)
	args = appendStr(args, "ssl-cert-file", c.SslCertFile)

	// ── Server – timeouts ─────────────────────────────────────────────────────
	args = appendInt(args, "timeout", c.Timeout)
	args = appendInt(args, "threads-http", c.ThreadsHttp)

	// ── Server – features ─────────────────────────────────────────────────────
	args = appendBool(args, "embeddings", c.Embeddings)
	args = appendBool(args, "reranking", c.Reranking)
	args = appendBool(args, "metrics", c.Metrics)
	args = appendBool(args, "props", c.Props)
	args = appendToggle(args, "slots", c.Slots)
	args = appendStr(args, "pooling", c.Pooling)

	// ── Server – slots / batching ─────────────────────────────────────────────
	args = appendStr(args, "slot-save-path", c.SlotSavePath)
	args = appendFloat(args, "slot-prompt-similarity", c.SlotPromptSimilarity)
	args = appendInt(args, "sleep-idle-seconds", c.SleepIdleSeconds)

	// ── Server – context cache ────────────────────────────────────────────────
	args = appendToggle(args, "cache-prompt", c.CachePrompt)
	args = appendInt(args, "cache-reuse", c.CacheReuse)
	args = appendInt(args, "ctx-checkpoints", c.CtxCheckpoints)
	args = appendInt(args, "checkpoint-every-n-tokens", c.CheckpointEveryNTokens)
	args = appendInt(args, "cache-ram", c.CacheRam)
	args = appendToggle(args, "context-shift", c.ContextShift)

	// ── Server – chat / template ──────────────────────────────────────────────
	args = appendToggle(args, "warmup", c.Warmup)
	args = appendBool(args, "spm-infill", c.SpmInfill)
	args = appendStr(args, "chat-template", c.ChatTemplate)
	args = appendStr(args, "chat-template-file", c.ChatTemplateFile)
	args = appendStr(args, "chat-template-kwargs", c.ChatTemplateKwargs)
	args = appendToggle(args, "jinja", c.Jinja)
	args = appendToggle(args, "prefill-assistant", c.PrefillAssistant)

	// ── Server – reasoning ────────────────────────────────────────────────────
	args = appendStr(args, "reasoning-format", c.ReasoningFormat)
	args = appendStr(args, "reasoning", c.Reasoning)
	args = appendInt(args, "reasoning-budget", c.ReasoningBudget)
	args = appendStr(args, "reasoning-budget-message", c.ReasoningBudgetMessage)

	// ── Server – Web UI ───────────────────────────────────────────────────────
	args = appendToggle(args, "webui", c.WebUI)
	args = appendStr(args, "webui-config", c.WebUIConfig)
	args = appendStr(args, "webui-config-file", c.WebUIConfigFile)
	args = appendToggle(args, "webui-mcp-proxy", c.WebUIMcpProxy)

	// ── Server – multimodal ───────────────────────────────────────────────────
	args = appendStr(args, "mmproj", c.Mmproj)
	args = appendStr(args, "mmproj-url", c.MmprojUrl)
	args = appendToggle(args, "mmproj-offload", c.MmprojOffload)
	if c.MmprojAuto != nil && !*c.MmprojAuto {
		args = append(args, "--no-mmproj")
	}
	args = appendInt(args, "image-min-tokens", c.ImageMinTokens)
	args = appendInt(args, "image-max-tokens", c.ImageMaxTokens)
	args = appendStr(args, "media-path", c.MediaPath)

	// ── Server – lookup cache ─────────────────────────────────────────────────
	args = appendStr(args, "lookup-cache-static", c.LookupCacheStatic)
	args = appendStr(args, "lookup-cache-dynamic", c.LookupCacheDynamic)

	// ── Server – router ───────────────────────────────────────────────────────
	args = appendStr(args, "models-dir", c.ModelsDir)
	args = appendStr(args, "models-preset", c.ModelsPreset)
	args = appendInt(args, "models-max", c.ModelsMax)
	args = appendToggle(args, "models-autoload", c.ModelsAutoload)

	// ── Speculative decoding ──────────────────────────────────────────────────
	args = appendStr(args, "model-draft", c.ModelDraft)
	args = appendInt(args, "threads-draft", c.ThreadsDraft)
	args = appendInt(args, "threads-batch-draft", c.ThreadsBatchDraft)
	args = appendInt(args, "draft", c.DraftMax)
	args = appendInt(args, "draft-min", c.DraftMin)
	args = appendFloat(args, "draft-p-min", c.DraftPMin)
	args = appendInt(args, "ctx-size-draft", c.CtxSizeDraft)
	args = appendStr(args, "device-draft", c.DeviceDraft)
	args = appendInt(args, "gpu-layers-draft", c.NGpuLayersDraft)
	args = appendStr(args, "override-tensor-draft", c.OverrideTensorDraft)
	args = appendBool(args, "cpu-moe-draft", c.CPUMoeDraft)
	args = appendInt(args, "n-cpu-moe-draft", c.NCpuMoeDraft)
	args = appendStr(args, "spec-type", c.SpecType)
	args = appendInt(args, "spec-ngram-size-n", c.SpecNgramSizeN)
	args = appendInt(args, "spec-ngram-size-m", c.SpecNgramSizeM)
	args = appendInt(args, "spec-ngram-min-hits", c.SpecNgramMinHits)

	// ── TTS / vocoder ─────────────────────────────────────────────────────────
	args = appendStr(args, "model-vocoder", c.ModelVocoder)
	args = appendBool(args, "tts-use-guide-tokens", c.TtsUseGuideTokens)

	// ── Extra passthrough ─────────────────────────────────────────────────────
	if c.Extra != nil {
		keys := make([]string, 0, len(c.Extra))
		for k := range c.Extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			args = append(args, "--"+k, c.Extra[k])
		}
	}

	return args
}

// ── ToArgs helpers ────────────────────────────────────────────────────────────

func appendStr(args []string, flag string, val *string) []string {
	if val != nil && *val != "" {
		return append(args, "--"+flag, *val)
	}
	return args
}

func appendInt(args []string, flag string, val *int) []string {
	if val != nil {
		return append(args, "--"+flag, strconv.Itoa(*val))
	}
	return args
}

func appendFloat(args []string, flag string, val *float64) []string {
	if val != nil {
		return append(args, "--"+flag, strconv.FormatFloat(*val, 'f', -1, 64))
	}
	return args
}

// appendBool emits --flag only when val is non-nil and true.
func appendBool(args []string, flag string, val *bool) []string {
	if val != nil && *val {
		return append(args, "--"+flag)
	}
	return args
}

// appendToggle emits --flag when true, --no-flag when false, nothing when nil.
func appendToggle(args []string, flag string, val *bool) []string {
	if val == nil {
		return args
	}
	if *val {
		return append(args, "--"+flag)
	}
	return append(args, "--no-"+flag)
}

// ── Store ─────────────────────────────────────────────────────────────────────

// Store manages reading and writing config files from a directory.
type Store struct {
	Dir string // path to the configs directory
}

// Save writes a config to a JSON file.
func (s *Store) Save(name string, cfg *Config) error {
	if name == "" {
		return errors.New("config name must not be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return errors.New("config name must not contain /, \\, or ..")
	}

	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	path := filepath.Join(s.Dir, name+".json")
	return os.WriteFile(path, data, 0o644)
}

// Load reads a config from a JSON file.
func (s *Store) Load(name string) (*Config, error) {
	if name == "" {
		return nil, errors.New("config name must not be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return nil, errors.New("config name must not contain /, \\, or ..")
	}

	path := filepath.Join(s.Dir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Exists checks if a config file exists.
func (s *Store) Exists(name string) bool {
	path := filepath.Join(s.Dir, name+".json")
	_, err := os.Stat(path)
	return err == nil
}

// List returns all config names.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			names = append(names, strings.TrimSuffix(entry.Name(), ".json"))
		}
	}
	return names, nil
}

// Delete removes a config file.
func (s *Store) Delete(name string) error {
	if name == "" {
		return errors.New("config name must not be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return errors.New("config name must not contain /, \\, or ..")
	}

	path := filepath.Join(s.Dir, name+".json")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
