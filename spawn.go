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
// Returns the new session ID, or "" on timeout.
func discoverNewHappierSession(existing map[string]bool, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for id := range listHappierSessionIDs() {
			if !existing[id] {
				return id
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return ""
}

// profileToHappierArgs returns the --profile flag args for happier, or nil if profile is empty.
// Isolates all happier --profile CLI vocabulary to this single function.
func profileToHappierArgs(profile string) []string {
	if profile == "" {
		return nil
	}
	return []string{"--profile", profile}
}

// happierAvailable returns true if the happier binary is in PATH.
func happierAvailable() bool {
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
func spawnSession(ctx context.Context, name, directory string, contextID, contextName, mission *string, model, profile string, worktreePath, gitBranch, repoPath *string, envOverrides map[string]string) (*Session, error) {
	s, err := createSession(ctx, name, directory, contextID, contextName, mission, false, model, profile, worktreePath, gitBranch, repoPath)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	var sessionCmd []string
	var preSpawnIDs map[string]bool
	var claudeUUID string
	useHappier := happierAvailable()

	if useHappier {
		preSpawnIDs = listHappierSessionIDs()
		sessionCmd = []string{"happier", "--yolo"}
		sessionCmd = append(sessionCmd, profileToHappierArgs(profile)...)
		sessionCmd = append(sessionCmd, "--model", effectiveModel(model))
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
		// Discover the happier session ID and set the session title in the mobile app.
		if happierID := discoverNewHappierSession(preSpawnIDs, 10*time.Second); happierID != "" {
			if name != "" {
				exec.Command("happier", "session", "set-title", happierID, name).Run()
			}
			if err := updateClaudeSessionID(ctx, s.ID, happierID); err != nil {
				log.Printf("spawn: failed to store happier session ID for %s: %v", s.ID, err)
			}
			s.ClaudeSessionID = &happierID
			log.Printf("spawn: created session %q (tmux=%s, dir=%s, happier=%s, model=%s, profile=%s)", name, s.TmuxSession, directory, happierID, model, profile)
		} else {
			log.Printf("spawn: created session %q (tmux=%s, dir=%s, model=%s, profile=%s) — happier ID discovery timed out", name, s.TmuxSession, directory, model, profile)
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
