package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SwarmOps supports launching a session with only a subset of MCP servers.
// When spawnSession is called with a non-empty mcp_servers list we:
//
//  1. Resolve each requested name against the swarmops MCP catalogue
//     (~/.swarmops/mcp-config.json — 35 servers).
//  2. Write a session-scoped JSON file containing ONLY those servers.
//  3. Launch claude with `--strict-mcp-config --mcp-config <file>` so it loads
//     exactly that subset (see spawnSession).
//
// Result: claude boots with just the subset; the user-level ~/.claude.json
// (with all ~50 servers) is ignored thanks to --strict-mcp-config. For
// [dseek]/[gpt] sessions this cuts the system prompt + tools payload from
// ~300k tokens to whatever the subset's tool definitions add up to.

const (
	swarmopsMCPCatalogue = "/home/sbarker/.swarmops/mcp-config.json"
	swarmopsAgentsDir    = "/home/sbarker/.swarmops/agents"

	// swarmopsMCPRestrictKey is a magic envOverrides key the MCP write tools
	// use to pass a comma-separated list of MCP server names into spawnSession
	// without growing the signature of every caller. spawnSession pops it out,
	// resolves the subset against the catalogue, and writes the per-session
	// restricted-mcp-config.json. The key is stripped before the env is
	// passed to tmux -e — it must not appear in the spawned process.
	swarmopsMCPRestrictKey = "__SWOP_MCP_RESTRICT__"
)

// mcpServerCatalogue holds the parsed mcp-config.json structure. Each value
// is the raw JSON for one server (transport, url, auth, etc.) — we don't
// need to interpret it, just route the requested subset through verbatim.
type mcpServerCatalogue struct {
	MCPServers map[string]json.RawMessage `json:"mcpServers"`
}

// loadMCPCatalogue reads the swarmops MCP catalogue. Returns an error
// describing the failure if the file is missing or malformed; callers
// should treat that as "fall back to unrestricted spawn" so a config
// issue doesn't break session creation entirely.
func loadMCPCatalogue() (*mcpServerCatalogue, error) {
	data, err := os.ReadFile(swarmopsMCPCatalogue)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", swarmopsMCPCatalogue, err)
	}
	var cat mcpServerCatalogue
	if err := json.Unmarshal(data, &cat); err != nil {
		return nil, fmt.Errorf("parse %s: %w", swarmopsMCPCatalogue, err)
	}
	return &cat, nil
}

// writeRestrictedMCPConfig writes a per-session subset of the swarmops
// MCP catalogue to disk and returns its absolute path. The file lives
// under ~/.swarmops/agents/<sessionID>/ alongside the agent's other
// scratch state.
//
// Unknown server names in `requested` are reported as a soft error
// (returned alongside the file path) so the caller can surface them to
// the user, but the spawn still proceeds with whatever was matched.
// An empty matched set is treated as a hard error — better to fail
// loudly than to spawn a session with zero tools.
func writeRestrictedMCPConfig(sessionID string, requested []string) (string, []string, error) {
	if len(requested) == 0 {
		return "", nil, nil // caller treats "" as "no restriction"
	}
	cat, err := loadMCPCatalogue()
	if err != nil {
		return "", nil, err
	}
	subset := map[string]json.RawMessage{}
	var unknown []string
	for _, name := range requested {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if cfg, ok := cat.MCPServers[name]; ok {
			subset[name] = cfg
		} else {
			unknown = append(unknown, name)
		}
	}
	if len(subset) == 0 {
		return "", unknown, fmt.Errorf("no requested MCP servers matched the catalogue at %s", swarmopsMCPCatalogue)
	}
	dir := filepath.Join(swarmopsAgentsDir, sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", unknown, fmt.Errorf("mkdir %s: %w", dir, err)
	}
	out := filepath.Join(dir, "restricted-mcp-config.json")
	body, err := json.MarshalIndent(mcpServerCatalogue{MCPServers: subset}, "", "  ")
	if err != nil {
		return "", unknown, fmt.Errorf("marshal restricted config: %w", err)
	}
	if err := os.WriteFile(out, body, 0o644); err != nil {
		return "", unknown, fmt.Errorf("write %s: %w", out, err)
	}
	return out, unknown, nil
}

// stashMCPServers pulls the `mcp_servers` arg out of an MCP tool's args
// map and encodes it into envOverrides under the magic swarmopsMCPRestrictKey
// so it threads through to spawnSession without changing the signature of
// every layer in between. Returns envOverrides (possibly newly-allocated).
//
// Accepts either []interface{} (Go's typical MCP unmarshal of a JSON array)
// or []string. Empty / missing / all-blank input is a no-op.
func stashMCPServers(args map[string]interface{}, envOverrides map[string]string) map[string]string {
	raw, ok := args["mcp_servers"]
	if !ok || raw == nil {
		return envOverrides
	}
	var names []string
	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					names = append(names, trimmed)
				}
			}
		}
	case []string:
		for _, s := range v {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				names = append(names, trimmed)
			}
		}
	}
	if len(names) == 0 {
		return envOverrides
	}
	if envOverrides == nil {
		envOverrides = map[string]string{}
	}
	envOverrides[swarmopsMCPRestrictKey] = strings.Join(names, ",")
	return envOverrides
}

// applyMCPRestriction writes the per-session restricted MCP config and returns
// its path so spawnSession can pass `--strict-mcp-config --mcp-config <path>`
// to claude directly. Called from spawn.go after the session row exists (we
// need the session id for the file path). Returns "" when there's nothing to
// restrict.
// restrictedMCPConfigPath returns the path to a session's restricted MCP config
// if one was written at spawn (i.e. the session was created with an mcp_servers
// subset), or "" otherwise. Used on restore/resume so the subset is re-applied
// via --strict-mcp-config instead of the session coming back with all servers.
func restrictedMCPConfigPath(sessionID string) string {
	p := filepath.Join(swarmopsAgentsDir, sessionID, "restricted-mcp-config.json")
	if _, err := os.Stat(p); err != nil {
		return ""
	}
	return p
}

func applyMCPRestriction(sessionID string, requested []string) (string, []string, error) {
	if len(requested) == 0 {
		return "", nil, nil
	}
	path, unknown, err := writeRestrictedMCPConfig(sessionID, requested)
	if err != nil {
		return "", unknown, err
	}
	return path, unknown, nil
}
