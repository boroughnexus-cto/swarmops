package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	xansi "github.com/charmbracelet/x/ansi"
)

// Session represents a managed Claude Code tmux session.
type Session struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	TmuxSession     string  `json:"tmux_session"`
	Directory       string  `json:"directory"`
	Mission         *string `json:"mission,omitempty"`
	ClaudeSessionID *string `json:"claude_session_id,omitempty"`
	Model           string  `json:"model,omitempty"` // "" = default claude model
	Hidden          bool    `json:"hidden"`
	Status          string  `json:"status"`
	CreatedAt       int64   `json:"created_at"`
	UpdatedAt       int64   `json:"updated_at"`
	// Agent worktree fields (nil for plain swop_run_task sessions).
	WorktreePath *string `json:"worktree_path,omitempty"`
	GitBranch    *string `json:"git_branch,omitempty"`
	RepoPath     *string `json:"repo_path,omitempty"`
}

// ManagedSessionEvent is a single audit trail entry for a managed session.
type ManagedSessionEvent struct {
	ID        int64  `json:"id"`
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
	EventType string `json:"event_type"` // created | stopped | deleted | renamed | mission_set
	Details   string `json:"details,omitempty"`
	Timestamp int64  `json:"ts"`
}

// recordSessionEvent writes an audit event. Safe to call with nil database.
func recordSessionEvent(sessionID, name, eventType, details string) {
	if database == nil {
		return
	}
	_, err := database.Exec(
		`INSERT INTO managed_session_events (session_id, name, event_type, details, ts) VALUES (?, ?, ?, ?, ?)`,
		sessionID, name, eventType, details, time.Now().Unix(),
	)
	if err != nil {
		log.Printf("audit: record event %s/%s: %v", sessionID, eventType, err)
	}
}

// listAuditEvents returns recent session audit events, newest first.
func listAuditEvents(ctx context.Context, limit int) ([]ManagedSessionEvent, error) {
	if database == nil {
		return nil, nil
	}
	rows, err := database.QueryContext(ctx,
		`SELECT id, session_id, name, event_type, COALESCE(details,''), ts
		 FROM managed_session_events ORDER BY ts DESC, id DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []ManagedSessionEvent
	for rows.Next() {
		var e ManagedSessionEvent
		if err := rows.Scan(&e.ID, &e.SessionID, &e.Name, &e.EventType, &e.Details, &e.Timestamp); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func generateID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func createSession(ctx context.Context, name, directory string, mission *string, hidden bool, model string) (*Session, error) {
	id := generateID()
	tmuxName := "sw-" + id
	now := time.Now().Unix()

	_, err := database.ExecContext(ctx,
		`INSERT INTO managed_sessions (id, name, tmux_session, directory, mission, model, hidden, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, name, tmuxName, directory, mission, nullableString(model), boolToInt(hidden), "running", now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}

	recordSessionEvent(id, name, "created", fmt.Sprintf(`{"directory":%q}`, directory))
	fireWebhook("session_created", map[string]interface{}{
		"id": id, "name": name, "directory": directory,
	})
	return &Session{
		ID:          id,
		Name:        name,
		TmuxSession: tmuxName,
		Directory:   directory,
		Mission:     mission,
		Model:       model,
		Hidden:      hidden,
		Status:      "running",
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// nullableStringPtr returns nil if p is nil or points to an empty string.
func nullableStringPtr(p *string) interface{} {
	if p == nil || *p == "" {
		return nil
	}
	return *p
}

// nullableString returns nil if s is empty (for SQL nullable columns).
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func listSessions(ctx context.Context) ([]Session, error) {
	rows, err := database.QueryContext(ctx,
		`SELECT id, name, tmux_session, directory, mission, claude_session_id, COALESCE(model,''), hidden, status, created_at, updated_at, worktree_path, git_branch, repo_path
		 FROM managed_sessions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		var mission, claudeID, worktreePath, gitBranch, repoPath sql.NullString
		var hiddenInt int
		if err := rows.Scan(&s.ID, &s.Name, &s.TmuxSession, &s.Directory, &mission, &claudeID, &s.Model, &hiddenInt, &s.Status, &s.CreatedAt, &s.UpdatedAt, &worktreePath, &gitBranch, &repoPath); err != nil {
			return nil, err
		}
		if mission.Valid {
			s.Mission = &mission.String
		}
		if claudeID.Valid {
			s.ClaudeSessionID = &claudeID.String
		}
		if worktreePath.Valid {
			s.WorktreePath = &worktreePath.String
		}
		if gitBranch.Valid {
			s.GitBranch = &gitBranch.String
		}
		if repoPath.Valid {
			s.RepoPath = &repoPath.String
		}
		s.Hidden = hiddenInt != 0
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func getSession(ctx context.Context, id string) (*Session, error) {
	var s Session
	var mission, claudeID, worktreePath, gitBranch, repoPath sql.NullString
	var hiddenInt int
	err := database.QueryRowContext(ctx,
		`SELECT id, name, tmux_session, directory, mission, claude_session_id, COALESCE(model,''), hidden, status, created_at, updated_at, worktree_path, git_branch, repo_path
		 FROM managed_sessions WHERE id = ?`, id,
	).Scan(&s.ID, &s.Name, &s.TmuxSession, &s.Directory, &mission, &claudeID, &s.Model, &hiddenInt, &s.Status, &s.CreatedAt, &s.UpdatedAt, &worktreePath, &gitBranch, &repoPath)
	if err != nil {
		return nil, err
	}
	if mission.Valid {
		s.Mission = &mission.String
	}
	if claudeID.Valid {
		s.ClaudeSessionID = &claudeID.String
	}
	if worktreePath.Valid {
		s.WorktreePath = &worktreePath.String
	}
	if gitBranch.Valid {
		s.GitBranch = &gitBranch.String
	}
	if repoPath.Valid {
		s.RepoPath = &repoPath.String
	}
	s.Hidden = hiddenInt != 0
	return &s, nil
}

func deleteSession(ctx context.Context, id string) error {
	if database == nil {
		return nil
	}
	s, err := getSession(ctx, id)
	if err != nil {
		return err
	}
	// Kill the tmux session if it exists
	exec.Command("tmux", "kill-session", "-t", s.TmuxSession).Run()
	// Archive and delete the scrollback snapshot
	deleteSessionSnapshot(s.ID)
	recordSessionEvent(id, s.Name, "deleted", "")
	fireWebhook("session_deleted", map[string]interface{}{"id": id, "name": s.Name})
	_, err = database.ExecContext(ctx, "DELETE FROM managed_sessions WHERE id = ?", id)
	return err
}

func updateSessionStatus(ctx context.Context, id, status string) error {
	_, err := database.ExecContext(ctx,
		"UPDATE managed_sessions SET status = ?, updated_at = ? WHERE id = ?",
		status, time.Now().Unix(), id,
	)
	if err == nil {
		fireWebhook("session_status_changed", map[string]interface{}{"id": id, "status": status})
		if status == "stopped" {
			if s, e := getSession(ctx, id); e == nil {
				recordSessionEvent(id, s.Name, "stopped", "")
			}
		}
	}
	return err
}

func updateSessionMission(ctx context.Context, id, mission string) error {
	if database == nil {
		return nil
	}
	_, err := database.ExecContext(ctx,
		"UPDATE managed_sessions SET mission = ?, updated_at = ? WHERE id = ?",
		mission, time.Now().Unix(), id,
	)
	return err
}

func updateClaudeSessionID(ctx context.Context, id, claudeSessionID string) error {
	if database == nil {
		return nil
	}
	_, err := database.ExecContext(ctx,
		"UPDATE managed_sessions SET claude_session_id = ?, updated_at = ? WHERE id = ?",
		claudeSessionID, time.Now().Unix(), id,
	)
	return err
}

func renameSession(ctx context.Context, id, name string) error {
	if database == nil {
		return nil
	}
	_, err := database.ExecContext(ctx,
		"UPDATE managed_sessions SET name = ?, updated_at = ? WHERE id = ?",
		name, time.Now().Unix(), id,
	)
	if err != nil {
		return err
	}
	recordSessionEvent(id, name, "renamed", "")
	return nil
}

func updateSessionDirectory(ctx context.Context, id, directory string) error {
	if database == nil {
		return nil
	}
	_, err := database.ExecContext(ctx,
		"UPDATE managed_sessions SET directory = ?, updated_at = ? WHERE id = ?",
		directory, time.Now().Unix(), id,
	)
	return err
}

// updateSessionFields atomically updates any combination of directory and
// mission in a single SQL statement. Pass nil for fields to leave unchanged.
func updateSessionFields(ctx context.Context, id string, directory, mission *string) error {
	if database == nil {
		return nil
	}
	setClauses := []string{"updated_at = ?"}
	args := []interface{}{time.Now().Unix()}
	if directory != nil {
		setClauses = append(setClauses, "directory = ?")
		args = append(args, nullableString(*directory))
	}
	if mission != nil {
		setClauses = append(setClauses, "mission = ?")
		args = append(args, nullableString(*mission))
	}
	if len(setClauses) == 1 {
		return nil
	}
	args = append(args, id)
	query := "UPDATE managed_sessions SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	_, err := database.ExecContext(ctx, query, args...)
	return err
}

func updateSessionContext(ctx context.Context, id, contextID, contextName string) error {
	if database == nil {
		return nil
	}
	_, err := database.ExecContext(ctx,
		"UPDATE managed_sessions SET updated_at = ? WHERE id = ?",
		nullableString(contextID), nullableString(contextName), time.Now().Unix(), id,
	)
	return err
}

// refreshSessionStatuses checks each session's tmux and updates status accordingly.
func refreshSessionStatuses(ctx context.Context) {
	sessions, err := listSessions(ctx)
	if err != nil {
		return
	}
	for _, s := range sessions {
		alive := isTmuxAlive(s.TmuxSession)
		if alive && s.Status != "running" {
			updateSessionStatus(ctx, s.ID, "running")
		} else if !alive && s.Status == "running" {
			updateSessionStatus(ctx, s.ID, "stopped")
		}
	}
}

func isTmuxAlive(name string) bool {
	return exec.Command("tmux", "has-session", "-t", name).Run() == nil
}

func captureTerminal(tmuxName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx,
		"tmux", "capture-pane", "-p", "-S", "-300", "-t", tmuxName,
	).Output()
	if err != nil {
		return "", fmt.Errorf("capture-pane: %w", err)
	}
	return xansi.Strip(string(out)), nil
}

func injectToSession(tmuxName, text string) error {
	if out, err := exec.Command("tmux", "send-keys", "-t", tmuxName, "-l", "--", text).CombinedOutput(); err != nil {
		return fmt.Errorf("tmux send-keys: %v: %s", err, out)
	}
	if out, err := exec.Command("tmux", "send-keys", "-t", tmuxName, "Enter").CombinedOutput(); err != nil {
		return fmt.Errorf("tmux enter: %v: %s", err, out)
	}
	return nil
}

// compactWatcher polls a tmux pane for Claude Code's compact-conversation prompt
// and automatically declines it for the given duration. Used after --resume to
// prevent history loss when restarting for MCP reconnection or after a reboot.
func compactWatcher(tmuxSession string, duration time.Duration) {
	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		if !isTmuxAlive(tmuxSession) {
			return
		}
		out, err := captureTerminal(tmuxSession)
		if err != nil {
			continue
		}
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		start := len(lines) - 6
		if start < 0 {
			start = 0
		}
		for _, line := range lines[start:] {
			lower := strings.ToLower(strings.TrimSpace(line))
			// Match Claude Code's compact prompt: contains "compact" + looks like a Y/n question
			if strings.Contains(lower, "compact") && (strings.Contains(lower, "y/n") ||
				strings.Contains(lower, "yes/no") ||
				strings.Contains(lower, "would you like") ||
				strings.Contains(lower, "continue without") ||
				strings.HasSuffix(lower, "? ") ||
				strings.HasSuffix(lower, "?")) &&
				!strings.Contains(lower, "/compact") {
				exec.Command("tmux", "send-keys", "-t", tmuxSession, "n", "Enter").Run()
				log.Printf("compact-watcher: auto-declined compact prompt for %s", tmuxSession)
				return
			}
		}
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// syncTmuxSessions detects tmux sessions that were killed externally
// and updates their status. Called periodically from main.
func syncTmuxSessions() {
	ctx := context.Background()
	sessions, err := listSessions(ctx)
	if err != nil {
		log.Printf("session: sync error: %v", err)
		return
	}
	for _, s := range sessions {
		alive := isTmuxAlive(s.TmuxSession)
		if !alive && s.Status == "running" {
			updateSessionStatus(ctx, s.ID, "stopped")
		} else if alive && s.Status == "stopped" {
			updateSessionStatus(ctx, s.ID, "running")
		}
	}
}
