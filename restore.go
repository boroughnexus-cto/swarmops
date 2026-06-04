package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var restoreComplete = make(chan struct{})

func restoreSessions(ctx context.Context) {
	defer close(restoreComplete)
	sessions, err := listSessions(ctx)
	if err != nil {
		log.Printf("restore: list sessions: %v", err)
		return
	}
	restored := 0
	for i, s := range sessions {
		if s.Hidden || isTmuxAlive(s.TmuxSession) {
			continue
		}
		dir := s.Directory
		if _, err := os.Stat(dir); err != nil {
			log.Printf("restore: %s dir %s missing", s.Name, dir)
			if h, err := os.UserHomeDir(); err == nil {
				dir = h
			} else {
				dir = "/tmp"
			}
		}
		updateSessionStatus(ctx, s.ID, "restoring")
		cArgs := buildClaudeRestoreArgs(ctx, &s)
		args := []string{"new-session", "-d", "-s", s.TmuxSession, "-c", dir, "-x", "200", "-y", "50"}
		for k, v := range restoreEnvFor(&s) {
			args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
		}
		args = append(args, "--")
		args = append(args, cArgs...)
		if out, err := exec.Command("tmux", args...).CombinedOutput(); err != nil {
			log.Printf("restore: tmux for %s: %v: %s", s.Name, err, out)
			updateSessionStatus(ctx, s.ID, "stopped")
			continue
		}
		if isTmuxAlive(s.TmuxSession) {
			updateSessionStatus(ctx, s.ID, "running")
			restored++
			// Inject orientation prompt after Claude has had time to initialise.
			// Stagger by session index so restores don't all fire at the same instant.
			delay := 12*time.Second + time.Duration(i)*3*time.Second
			go injectRestorePrompt(s, delay)
			go compactWatcher(s.TmuxSession, delay+90*time.Second)
		} else {
			updateSessionStatus(ctx, s.ID, "stopped")
		}
	}
	log.Printf("restore: restored %d/%d sessions", restored, len(sessions))
}

// injectRestorePrompt waits for Claude to finish initialising then sends a
// brief orientation message so the session resumes active work rather than
// sitting idle at the prompt.
func injectRestorePrompt(s Session, delay time.Duration) {
	time.Sleep(delay)
	if !isTmuxAlive(s.TmuxSession) {
		return
	}
	var parts []string
	parts = append(parts, fmt.Sprintf("You have been automatically restored after a system restart. Your session name is %q.", s.Name))
	if s.Mission != nil && strings.TrimSpace(*s.Mission) != "" {
		parts = append(parts, fmt.Sprintf("Your mission is: %s", *s.Mission))
	}
	parts = append(parts, "Please review your prior context and conversation history, then continue your work.")
	prompt := strings.Join(parts, " ")
	if err := injectToSession(s.TmuxSession, prompt); err != nil {
		log.Printf("restore: inject prompt for %s: %v", s.Name, err)
	} else {
		log.Printf("restore: injected orientation prompt for %s", s.Name)
	}
}

func buildClaudeRestoreArgs(ctx context.Context, s *Session) []string {
	// Resume the prior conversation by its persisted session id. Sessions created
	// in native mode have a UUID --session-id; resume them with --resume. Legacy
	// sessions with a non-UUID id start a fresh conversation — their history is
	// still on disk and can be reopened with /resume in-session.
	args := []string{"claude"}
	args = append(args, remoteControlArgs(s.Name)...)
	args = append(args, "--dangerously-skip-permissions")
	if s.ClaudeSessionID != nil && isValidUUID(*s.ClaudeSessionID) {
		args = append(args, "--resume", *s.ClaudeSessionID)
	}
	// Re-apply the per-session MCP subset if one was set at spawn, so a restored
	// session doesn't come back with the full ~50-server catalogue.
	if p := restrictedMCPConfigPath(s.ID); p != "" {
		args = append(args, "--strict-mcp-config", "--mcp-config", p)
	}
	// Model is passed via ANTHROPIC_MODEL in restoreEnvFor.
	return args
}

// restoreEnvFor returns the env vars that should be re-injected when a
// persisted session is restored after a SwarmOps restart. Always sets
// ANTHROPIC_MODEL; for sessions whose names mark them as LiteLLM-routed
// ([gpt] or [dseek] prefix), also restores the LiteLLM endpoint env so the
// resumed claude process targets the same backend it was created with.
func restoreEnvFor(s *Session) map[string]string {
	env := map[string]string{"ANTHROPIC_MODEL": effectiveModel(s.Model)}
	switch {
	case strings.HasPrefix(s.Name, "[gpt]"), strings.HasPrefix(s.Name, "[dseek]"):
		for k, v := range litellmEnvOverrides("") {
			env[k] = v
		}
		// Re-apply the explicit model the session was created with so the
		// LiteLLM backend stays consistent across restarts. effectiveModel
		// returns the configured fallback when s.Model is empty.
		env["ANTHROPIC_MODEL"] = effectiveModel(s.Model)
	}
	return env
}
