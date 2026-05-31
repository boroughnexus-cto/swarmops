package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// uuidV4Pattern validates a canonical UUID v4 string.
var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// defaultSessionModelFallback is the model used when neither the session record
// nor SWARMOPS_DEFAULT_MODEL specifies one. Happier/Claude otherwise picks its
// own default, which has been Opus-1M — too expensive for routine sessions.
const defaultSessionModelFallback = "sonnet"

// effectiveModel resolves the model string passed to happier --model.
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

// happierDaemonEntry is a minimal parse of one entry in the happier daemon list JSON array.
type happierDaemonEntry struct {
	HappySessionID string `json:"happySessionId"`
}

// listHappierSessionIDs returns the current set of happier session IDs from the daemon.
func listHappierSessionIDs() map[string]bool {
	out, err := exec.Command("happier", "daemon", "list", "--json").Output()
	if err != nil {
		return nil
	}
	// Output has a non-JSON prefix line ("Active sessions:"); find the JSON array.
	raw := string(out)
	idx := strings.Index(raw, "[")
	if idx < 0 {
		return nil
	}
	var entries []happierDaemonEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw[idx:])), &entries); err != nil {
		return nil
	}
	ids := make(map[string]bool, len(entries))
	for _, e := range entries {
		if e.HappySessionID != "" {
			ids[e.HappySessionID] = true
		}
	}
	return ids
}

// discoverNewHappierSession polls the daemon until a session appears that isn't in existing.
// Returns the new session ID, or "" on timeout. Skips transient "PID-<pid>"
// handles — happier exposes those in the daemon list before the server-side
// session id (a cuid) is assigned, and they are NOT valid set-title targets
// (set-title returns session_not_found). We wait for the real cuid.
func discoverNewHappierSession(existing map[string]bool, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for id := range listHappierSessionIDs() {
			if !existing[id] && !strings.HasPrefix(id, "PID-") {
				return id
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return ""
}

// happierActiveSession is a minimal parse of one entry in
// `happier session list --active --json`.
type happierActiveSession struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

// listActiveHappierSessions returns active happier sessions from the SERVER
// view (`happier session list --active`). Unlike the daemon list this only
// ever contains real session cuids (never transient "PID-<pid>" handles) and
// carries each session's working directory — so it's the reliable way to bind
// a freshly-launched session to its SwarmOps record by cwd.
func listActiveHappierSessions() []happierActiveSession {
	out, err := exec.Command("happier", "session", "list", "--active", "--json").Output()
	if err != nil {
		return nil
	}
	var resp struct {
		Data struct {
			Sessions []happierActiveSession `json:"sessions"`
		} `json:"data"`
	}
	if json.Unmarshal(out, &resp) != nil {
		return nil
	}
	return resp.Data.Sessions
}

// activeHappierIDSet snapshots the ids of currently-active happier sessions,
// used as the "before launch" exclusion set so discoverHappierSessionForDir
// only matches the session this launch creates.
func activeHappierIDSet() map[string]bool {
	m := map[string]bool{}
	for _, s := range listActiveHappierSessions() {
		m[s.ID] = true
	}
	return m
}

// discoverHappierSessionForDir polls the server session list for an active
// session whose working directory is dir and whose id is not in exclude (the
// pre-launch snapshot). Returns the id, or "" on timeout. Keying on cwd +
// a pre-launch exclusion set is robust to launch bursts (many sessions
// starting at once) where "whatever new id appeared" is ambiguous, and it
// never returns a transient PID- handle.
func discoverHappierSessionForDir(dir string, exclude map[string]bool, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, s := range listActiveHappierSessions() {
			if s.Path == dir && !exclude[s.ID] && !strings.HasPrefix(s.ID, "PID-") {
				return s.ID
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return ""
}

// setHappierTitle sets the happier mobile-app title for a session so the app
// shows the SwarmOps session name rather than the working-directory basename.
//
// Two subtleties this handles:
//   - `happier session set-title` exits 0 even when the session isn't found
//     yet (it prints `{"ok":false,...}` but returns status 0), so we must parse
//     the --json `ok` field — a plain .Run() can't tell success from failure.
//   - Right after launch the session is in the daemon list but may not be
//     persisted server-side for a moment, so we retry briefly.
func setHappierTitle(happierID, name string) {
	if happierID == "" || name == "" {
		return
	}
	for attempt := 1; attempt <= 6; attempt++ {
		out, err := exec.Command("happier", "session", "set-title", happierID, name, "--json").Output()
		if err == nil {
			var resp struct {
				OK bool `json:"ok"`
			}
			if json.Unmarshal(out, &resp) == nil && resp.OK {
				return
			}
		}
		time.Sleep(time.Second)
	}
	log.Printf("happier: set-title failed for %s (%q) after retries", happierID, name)
}

// syncHappierIdentity binds the happier session that appeared after a (re)launch
// to a managed SwarmOps session: it discovers the new happier id (one present
// now but not in preIDs), persists it as the session's claude_session_id, and
// sets the happier app title to the SwarmOps name. Used by the restore and
// resume paths — which relaunch `happier --yolo` as a fresh session and so get
// a NEW happier id each time — so the app keeps showing the SwarmOps name
// across restarts instead of falling back to the cwd. Returns the discovered
// id, or "" on timeout.
func syncHappierIdentity(ctx context.Context, sessionID, name, dir string, excludeIDs map[string]bool, timeout time.Duration) string {
	happierID := discoverHappierSessionForDir(dir, excludeIDs, timeout)
	if happierID == "" {
		log.Printf("happier: no active session for %q (dir %s) within %s — title not set", name, dir, timeout)
		return ""
	}
	if err := updateClaudeSessionID(ctx, sessionID, happierID); err != nil {
		log.Printf("happier: store session id for %s: %v", sessionID, err)
	}
	setHappierTitle(happierID, name)
	return happierID
}

// profileToHappierArgs returns the --profile flag args for happier, or nil if profile is empty.
// Isolates all happier --profile CLI vocabulary to this single function.
func profileToHappierArgs(profile string) []string {
	if profile == "" {
		return nil
	}
	return []string{"--profile", profile}
}

// happierAvailable returns true if the happier binary is in PATH and the
// SWARMOPS_DISABLE_HAPPIER env var is not set. The env var exists because
// `happier --yolo` is broken on some hosts (deletes its hook settings file
// before claude can read it, causing the spawned tmux session to die
// immediately). When happier is disabled or absent, spawnSession falls back
// to launching claude directly with a persisted --session-id.
func happierAvailable() bool {
	if os.Getenv("SWARMOPS_DISABLE_HAPPIER") != "" {
		return false
	}
	_, err := exec.LookPath("happier")
	return err == nil
}

// spawnSession creates a new tmux session and launches happier (or claude as fallback) inside it.
// When happier is installed, sessions are launched via happier (--yolo) so they are visible in
// the happier mobile app. When happier is not in PATH, falls back to claude directly.
// An optional profile string selects a happier backend profile (e.g. "deepseek", "openai").
// An optional model string selects a specific model within the profile.
// worktreePath, gitBranch, and repoPath are optional; pass nil for plain sessions.
// envOverrides injects extra environment variables into the tmux session via -e flags.
// Pass nil to inherit the swarmops process environment unchanged (default behavior).
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

	if len(mcpRestrict) > 0 {
		var unknown []string
		var rerr error
		envOverrides, unknown, rerr = applyMCPRestriction(s.ID, mcpRestrict, envOverrides)
		if rerr != nil {
			deleteSession(ctx, s.ID)
			return nil, fmt.Errorf("mcp restriction: %w", rerr)
		}
		if len(unknown) > 0 {
			log.Printf("spawn: session %s: ignored unknown MCP servers: %s", s.ID, strings.Join(unknown, ","))
		}
	}

	var sessionCmd []string
	var preSpawnIDs map[string]bool
	var claudeUUID string
	useHappier := happierAvailable()

	if useHappier {
		preSpawnIDs = activeHappierIDSet()
		sessionCmd = []string{"happier", "--yolo"}
		sessionCmd = append(sessionCmd, profileToHappierArgs(profile)...)
		// happier's `--model` flag triggers a bug where it deletes its hook
		// settings file before the spawned claude can read it (claude then
		// exits with "Settings file not found"). Pass the model via the
		// ANTHROPIC_MODEL env var instead — claude reads that natively and
		// happier passes the env through to the child. Don't clobber an
		// explicit ANTHROPIC_MODEL the caller set (used by LiteLLM routing).
		if envOverrides == nil {
			envOverrides = map[string]string{}
		}
		if _, set := envOverrides["ANTHROPIC_MODEL"]; !set {
			envOverrides["ANTHROPIC_MODEL"] = effectiveModel(model)
		}
	} else {
		// Fall back to launching claude directly with a persisted --session-id.
		claudeUUID = generateUUID()
		sessionCmd = []string{"claude", "--session-id", claudeUUID, "--dangerously-skip-permissions"}
		if model != "" {
			sessionCmd = append(sessionCmd, "--model", model)
		}
		log.Printf("spawn: happier not found in PATH — falling back to claude for session %q", name)
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

	if useHappier {
		// Discover the new happier session (by cwd), persist its id, and set the
		// mobile-app title to the SwarmOps name.
		if happierID := syncHappierIdentity(ctx, s.ID, name, directory, preSpawnIDs, 15*time.Second); happierID != "" {
			s.ClaudeSessionID = &happierID
			log.Printf("spawn: created session %q (tmux=%s, dir=%s, happier=%s, model=%s, profile=%s)", name, s.TmuxSession, directory, happierID, model, profile)
		} else {
			log.Printf("spawn: created session %q (tmux=%s, dir=%s, model=%s, profile=%s) — happier session not found", name, s.TmuxSession, directory, model, profile)
		}
	} else {
		if err := updateClaudeSessionID(ctx, s.ID, claudeUUID); err != nil {
			log.Printf("spawn: failed to store claude session ID for %s: %v", s.ID, err)
		}
		s.ClaudeSessionID = &claudeUUID
		log.Printf("spawn: created session %q (tmux=%s, dir=%s, claude=%s, model=%s) [claude fallback]", name, s.TmuxSession, directory, claudeUUID, model)
	}

	return s, nil
}
