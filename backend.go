package main

import (
	"os"
	"strings"
)

// LiteLLM constants. The URL and key are env-driven so they can be rotated
// without a rebuild; defaults preserve the historical hardcoded values that
// were embedded in the MCP tool descriptions before this file existed.
const (
	defaultLiteLLMURL = "http://10.0.0.2:4000"
	defaultLiteLLMKey = "sk-35439ddea8690f7c89be8497e2f43e318d4890123d288cca"

	// LiteLLM model IDs surfaced via the TUI backend picker.
	litellmModelGPT55     = "chatgptsub-gpt-5.5"
	litellmModelDeepseek4 = "or-deepseek-v4-pro"
)

// litellmURL returns the configured LiteLLM proxy URL.
func litellmURL() string {
	if v := strings.TrimSpace(os.Getenv("SWARMOPS_LITELLM_URL")); v != "" {
		return v
	}
	return defaultLiteLLMURL
}

// litellmAPIKey returns the LiteLLM proxy API key.
func litellmAPIKey() string {
	if v := strings.TrimSpace(os.Getenv("SWARMOPS_LITELLM_API_KEY")); v != "" {
		return v
	}
	return defaultLiteLLMKey
}

// litellmEnvOverrides returns the env vars needed to route a Claude Code
// session through LiteLLM with the given model. Used when the TUI selects a
// LiteLLM-backed entry from the picker.
func litellmEnvOverrides(model string) map[string]string {
	env := map[string]string{
		"ANTHROPIC_BASE_URL": litellmURL(),
		"ANTHROPIC_API_KEY":  litellmAPIKey(),
	}
	if model != "" {
		env["ANTHROPIC_MODEL"] = model
	}
	return env
}

// autoPrefixSessionName tags a session name with [gpt] or [dseek] when the
// session routes through LiteLLM, so the source backend is visible in lists
// and tmux titles. Idempotent: names that already start with "[" are left
// alone. Empty names pass through unchanged so auto-generated IDs are used.
func autoPrefixSessionName(name, model string, env map[string]string) string {
	if name == "" || strings.HasPrefix(name, "[") {
		return name
	}
	if _, ok := env["ANTHROPIC_BASE_URL"]; !ok {
		return name
	}
	prefix := "[gpt] "
	if strings.HasPrefix(model, "or-deepseek") {
		prefix = "[dseek] "
	}
	return prefix + name
}
