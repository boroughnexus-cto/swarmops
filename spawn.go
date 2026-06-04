package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// uuidV4Pattern validates a canonical UUID v4 string.
var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// defaultSessionModelFallback is the model used when neither the session record
// nor SWARMOPS_DEFAULT_MODEL specifies one. Claude otherwise picks its own
// default, which has been Opus-1M — too expensive for routine sessions.
const defaultSessionModelFallback = "sonnet"

// effectiveModel resolves the model string passed to claude --model.
// Returns the explicit model if set; otherwise SWARMOPS_DEFAULT_MODEL; otherwise
// defaultSessionModelFallback ("sonnet"). Always returns a non-empty string.
func effectiveModel(model string) string {
	if model != "" {
		return model
	}
	if env := os.Getenv("SWARMOPS_DEFAULT_MODEL"); env != "" {
		return env
	}
	return defaultSessionModelFallback
}

// generateUUID produces a random UUID v4 string using crypto/rand (no external dependency).
func generateUUID() string {
	var b [16]byte
	rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// isValidUUID checks whether s is a valid UUID v4 in canonical lowercase hex form.
func isValidUUID(s string) bool {
	return uuidV4Pattern.MatchString(s)
}

// remoteControlArgs returns the claude flags that enable native Remote Control,
// named after the SwarmOps session so it's identifiable in the Remote Control
// UI. An empty name, or one starting with "-" (which the optional-value flag
// would mistake for the next flag), falls back to bare --remote-control.
func remoteControlArgs(name string) []string {
	// Kill switch: if Claude's Remote Control ever needs account/config we don't
	// have, set SWARMOPS_DISABLE_REMOTE_CONTROL=1 and restart to spawn without it
	// (no code revert needed).
	if os.Getenv("SWARMOPS_DISABLE_REMOTE_CONTROL") != "" {
		return nil
	}
	name = strings.TrimSpace(name)
	if name == "" || strings.HasPrefix(name, "-") {
		return []string{"--remote-control"}
	}
	return []string{"--remote-control", name}
}

// spawnSession creates a new tmux session and launches claude inside it.
// Sessions are launched with a persisted --session-id so the conversation can
// be resumed (claude --resume <id>) after a swarmops restart, and are controlled
// natively over tmux (send-keys / capture-pane) — no external wrapper.
// An optional model string selects the claude model.
// envOverrides injects extra environment variables into the tmux session via -e flags.
// Pass nil to inherit the swarmops process environment unchanged (default behavior).
// (profile is retained on the session record for backward compat but unused.)
func spawnSession(ctx context.Context, name, directory string, mission *string, model, profile string, envOverrides map[string]string) (*Session, error) {
	// If the caller specified ANTHROPIC_MODEL via env_overrides but no
	// explicit `model` arg, hoist it onto the session record. Without this,
	// LiteLLM-routed sessions (spawned with env_overrides only) end up with
	// Model="" in the DB — restore.go then falls back to effectiveModel("")
	// which is "sonnet", not the LiteLLM target. See bug: [dseek] sessions
	// failing to come back after a swarmops restart.
	if model == "" && envOverrides != nil {
		if m, ok := envOverrides["ANTHROPIC_MODEL"]; ok && m != "" {
			model = m
		}
	}

	// Default: route through the quota proxy on localhost:8082 to capture Anthropic
	// rate-limit headers (session / weekly utilization). LiteLLM-routed sessions
	// ([gpt]/[dseek] via litellmEnvOverrides) override this with the LiteLLM URL.
	const quotaProxyURL = "http://localhost:8082"
	if envOverrides == nil {
		envOverrides = map[string]string{
			"ANTHROPIC_BASE_URL": quotaProxyURL,
		}
	} else if _, hasBaseURL := envOverrides["ANTHROPIC_BASE_URL"]; !hasBaseURL {
		// Profile-backed sessions (e.g. "deepseek"/"openai" profiles that don't
		// explicitly set ANTHROPIC_BASE_URL) also route via proxy by default.
		envOverrides["ANTHROPIC_BASE_URL"] = quotaProxyURL
	}

	// Pop the magic MCP-restrict key out of envOverrides so it never reaches
	// the spawned process. Resolution happens after createSession because we
	// need the session id to name the per-session config file.
	var mcpRestrict []string
	if envOverrides != nil {
		if v, ok := envOverrides[swarmopsMCPRestrictKey]; ok {
			for _, name := range strings.Split(v, ",") {
				name = strings.TrimSpace(name)
				if name != "" {
					mcpRestrict = append(mcpRestrict, name)
				}
			}
			delete(envOverrides, swarmopsMCPRestrictKey)
		}
	}

	s, err := createSession(ctx, name, directory, mission, false, model, profile)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	var restrictedMCPPath string
	if len(mcpRestrict) > 0 {
		var unknown []string
		var rerr error
		restrictedMCPPath, unknown, rerr = applyMCPRestriction(s.ID, mcpRestrict)
		if rerr != nil {
			deleteSession(ctx, s.ID)
			return nil, fmt.Errorf("mcp restriction: %w", rerr)
		}
		if len(unknown) > 0 {
			log.Printf("spawn: session %s: ignored unknown MCP servers: %s", s.ID, strings.Join(unknown, ","))
		}
	}

	// Launch claude directly with a persisted --session-id so the conversation
	// can be resumed (claude --resume <id>) after a swarmops restart. Remote
	// Control is enabled (named after the session) on every spawn.
	claudeUUID := generateUUID()
	sessionCmd := []string{"claude"}
	sessionCmd = append(sessionCmd, remoteControlArgs(name)...)
	sessionCmd = append(sessionCmd, "--session-id", claudeUUID, "--dangerously-skip-permissions")
	if restrictedMCPPath != "" {
		sessionCmd = append(sessionCmd, "--strict-mcp-config", "--mcp-config", restrictedMCPPath)
	}
	if model != "" {
		sessionCmd = append(sessionCmd, "--model", model)
	}

	tmuxArgs := []string{"new-session", "-d",
		"-s", s.TmuxSession,
		"-c", directory,
		"-x", "200", "-y", "50",
	}
	for k, v := range envOverrides {
		tmuxArgs = append(tmuxArgs, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	tmuxArgs = append(tmuxArgs, "--")
	tmuxArgs = append(tmuxArgs, sessionCmd...)
	if out, err := exec.Command("tmux", tmuxArgs...).CombinedOutput(); err != nil {
		deleteSession(ctx, s.ID)
		return nil, fmt.Errorf("tmux new-session: %v: %s", err, out)
	}

	if err := updateClaudeSessionID(ctx, s.ID, claudeUUID); err != nil {
		log.Printf("spawn: failed to store claude session ID for %s: %v", s.ID, err)
	}
	s.ClaudeSessionID = &claudeUUID
	log.Printf("spawn: created session %q (tmux=%s, dir=%s, claude=%s, model=%s)", name, s.TmuxSession, directory, claudeUUID, model)

	return s, nil
}
