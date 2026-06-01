package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ─── Messages ────────────────────────────────────────────────────────────────

// goalBrainResultMsg carries the brain's routing decision back to the TUI Update loop.
type goalBrainResultMsg struct {
	goal string
	pick BrainPick
	err  string
}

// goalSpawnDoneMsg carries the result of a smart spawn (worktree + session or plain session).
type goalSpawnDoneMsg struct {
	sessionName string
	err         string
}

// ─── API response shape ──────────────────────────────────────────────────────

// smartSpawnResult is the JSON body returned by POST /api/swarm/smart-spawn.
type smartSpawnResult struct {
	Pick    BrainPick `json:"pick"`
	Session *Session  `json:"session,omitempty"`
	Error   string    `json:"error,omitempty"`
}

// ─── Tea commands ────────────────────────────────────────────────────────────

// goalBrainCmd calls the backend smart-spawn endpoint in dry-run mode and
// returns a goalBrainResultMsg when the brain has made its routing decision.
func goalBrainCmd(goal string, api swarmClient) tea.Cmd {
	return func() tea.Msg {
		if api == nil {
			return goalBrainResultMsg{goal: goal, err: "no backend connection"}
		}
		result, err := api.smartSpawn(goal, "", true)
		if err != nil {
			return goalBrainResultMsg{goal: goal, err: err.Error()}
		}
		if result.Error != "" {
			return goalBrainResultMsg{goal: goal, err: result.Error}
		}
		return goalBrainResultMsg{goal: goal, pick: result.Pick}
	}
}

// goalSpawnCmd calls the backend to create a worktree + session (or a plain
// session when repoSlug is empty) and returns a goalSpawnDoneMsg when done.
func goalSpawnCmd(goal, repoSlug string, api swarmClient) tea.Cmd {
	return func() tea.Msg {
		if api == nil {
			return goalSpawnDoneMsg{err: "no backend connection"}
		}
		result, err := api.smartSpawn(goal, repoSlug, false)
		if err != nil {
			return goalSpawnDoneMsg{err: err.Error()}
		}
		if result.Error != "" {
			return goalSpawnDoneMsg{err: result.Error}
		}
		name := ""
		if result.Session != nil {
			name = result.Session.Name
		}
		return goalSpawnDoneMsg{sessionName: name}
	}
}

// ─── View ────────────────────────────────────────────────────────────────────

// renderGoalConfirm renders the multi-line status area for modeGoalConfirm.
func renderGoalConfirm(m tuiModel) string {
	pick := m.goalPick
	var lines []string

	// Pick + confidence header
	pickLabel := pick.Pick
	if pickLabel == "" || pickLabel == "none" {
		pickLabel = "none (no matching repo)"
	}
	lines = append(lines, fmt.Sprintf("Brain: %s  confidence: %s", pickLabel, pick.Confidence))

	// One-line reasoning
	if pick.Reasoning != "" {
		lines = append(lines, dimStyle.Render(pick.Reasoning))
	}

	// Suggestions when pick is none
	if (pick.Pick == "" || pick.Pick == "none") && len(pick.Suggestions) > 0 {
		lines = append(lines, dimStyle.Render("Suggestions: "+strings.Join(pick.Suggestions, ", ")))
	}

	// Yes / No options
	opts := []string{"Yes — spawn session", "No — cancel"}
	var optParts []string
	for i, opt := range opts {
		if i == m.goalConfirmCursor {
			optParts = append(optParts, selectedStyle.Render("> "+opt))
		} else {
			optParts = append(optParts, dimStyle.Render("  "+opt))
		}
	}
	lines = append(lines, strings.Join(optParts, "  ")+dimStyle.Render("  (↑/↓ y/n Enter Esc)"))

	return strings.Join(lines, "\n")
}
