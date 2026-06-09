package main

import (
	"reflect"
	"strings"
	"testing"
)

// TestInteractiveClaudeArgsGolden pins the exact argv produced by the shared
// builder for the fresh and resume modes, so any change to flag content or
// ordering is deliberate (and shows up as a test diff). The pre-refactor
// hand-rolled builders produced exactly these sequences.
func TestInteractiveClaudeArgsGolden(t *testing.T) {
	t.Setenv("SWARMOPS_DISABLE_REMOTE_CONTROL", "") // force Remote Control on
	uuid := generateUUID()

	tests := []struct {
		name string
		opts interactiveClaudeOpts
		want []string
	}{
		{
			name: "fresh spawn with model + restricted mcp",
			opts: interactiveClaudeOpts{name: "task-a", mode: claudeFresh, sessionID: uuid, modelFlag: "opus", mcpConfig: "/cfg.json"},
			want: []string{"claude", "--remote-control", "task-a", "--session-id", uuid, dangerouslySkipPermissions, "--strict-mcp-config", "--mcp-config", "/cfg.json", "--model", "opus"},
		},
		{
			name: "fresh spawn no model no mcp (default model path)",
			opts: interactiveClaudeOpts{name: "task-b", mode: claudeFresh, sessionID: uuid},
			want: []string{"claude", "--remote-control", "task-b", "--session-id", uuid, dangerouslySkipPermissions},
		},
		{
			name: "resume restore: mcp, model via env (modelFlag empty)",
			opts: interactiveClaudeOpts{name: "task-c", mode: claudeResume, sessionID: uuid, mcpConfig: "/cfg.json"},
			want: []string{"claude", "--remote-control", "task-c", dangerouslySkipPermissions, "--resume", uuid, "--strict-mcp-config", "--mcp-config", "/cfg.json"},
		},
		{
			name: "resume TUI: model flag, no mcp",
			opts: interactiveClaudeOpts{name: "task-d", mode: claudeResume, sessionID: uuid, modelFlag: "sonnet"},
			want: []string{"claude", "--remote-control", "task-d", dangerouslySkipPermissions, "--resume", uuid, "--model", "sonnet"},
		},
		{
			name: "resume with legacy non-UUID id omits --resume",
			opts: interactiveClaudeOpts{name: "task-e", mode: claudeResume, sessionID: "legacy-abc", modelFlag: "opus"},
			want: []string{"claude", "--remote-control", "task-e", dangerouslySkipPermissions, "--model", "opus"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := interactiveClaudeArgs(tc.opts)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("argv mismatch\n got: %v\nwant: %v", got, tc.want)
			}
		})
	}
}

// TestInteractiveClaudeArgsInvariants is the structural guarantee behind the
// TUI/MCP parity requirement: EVERY interactive path emits
// --dangerously-skip-permissions, and Remote Control is on unless the documented
// SWARMOPS_DISABLE_REMOTE_CONTROL kill-switch is set (the one allowed exception).
func TestInteractiveClaudeArgsInvariants(t *testing.T) {
	uuid := generateUUID()
	// Mirrors every real call site: spawn (fresh), alt+S restart (fresh+forced
	// model), restore (resume+mcp, env model), TUI resume / StartSession (resume).
	callSites := []interactiveClaudeOpts{
		{name: "spawn", mode: claudeFresh, sessionID: uuid, modelFlag: "sonnet", mcpConfig: "/c.json"},
		{name: "altS", mode: claudeFresh, sessionID: uuid, modelFlag: "sonnet"},
		{name: "restore", mode: claudeResume, sessionID: uuid, mcpConfig: "/c.json"},
		{name: "resume", mode: claudeResume, sessionID: uuid, modelFlag: "opus"},
	}

	t.Run("skip-permissions always present", func(t *testing.T) {
		t.Setenv("SWARMOPS_DISABLE_REMOTE_CONTROL", "")
		for _, o := range callSites {
			args := interactiveClaudeArgs(o)
			if !hasFlag(args, dangerouslySkipPermissions) {
				t.Errorf("%s: missing %s; got %v", o.name, dangerouslySkipPermissions, args)
			}
			if !hasFlag(args, "--remote-control") {
				t.Errorf("%s: Remote Control must be on by default; got %v", o.name, args)
			}
		}
	})

	t.Run("kill-switch disables only Remote Control, never skip-permissions", func(t *testing.T) {
		t.Setenv("SWARMOPS_DISABLE_REMOTE_CONTROL", "1")
		for _, o := range callSites {
			args := interactiveClaudeArgs(o)
			if hasFlag(args, "--remote-control") {
				t.Errorf("%s: kill-switch should drop Remote Control; got %v", o.name, args)
			}
			if !hasFlag(args, dangerouslySkipPermissions) {
				t.Errorf("%s: skip-permissions must survive the RC kill-switch; got %v", o.name, args)
			}
		}
	})
}

func TestEffectiveModel(t *testing.T) {
	t.Setenv("SWARMOPS_DEFAULT_MODEL", "")

	if got := effectiveModel(""); got != "sonnet" {
		t.Errorf("effectiveModel(\"\") = %q, want %q", got, "sonnet")
	}
	if got := effectiveModel("opus"); got != "opus" {
		t.Errorf("effectiveModel(\"opus\") = %q, want %q", got, "opus")
	}

	t.Setenv("SWARMOPS_DEFAULT_MODEL", "haiku")
	if got := effectiveModel(""); got != "haiku" {
		t.Errorf("effectiveModel(\"\") with env override = %q, want %q", got, "haiku")
	}
	// Explicit model still wins over env override.
	if got := effectiveModel("opus"); got != "opus" {
		t.Errorf("effectiveModel(\"opus\") with env override = %q, want %q (explicit must win)", got, "opus")
	}
}

func TestResumeClaudeCmdNative(t *testing.T) {
	// Empty/invalid id + empty model → bare native claude, no --resume, no --model.
	// Remote Control is always enabled.
	args := resumeClaudeCmd("", "", "")
	if len(args) == 0 || args[0] != "claude" {
		t.Fatalf("resumeClaudeCmd should invoke claude; got %v", args)
	}
	if !hasFlag(args, "--remote-control") {
		t.Errorf("resumeClaudeCmd should always enable --remote-control; got %v", args)
	}
	if hasFlag(args, "--resume") {
		t.Errorf("resumeClaudeCmd(\"\",\"\",\"\") should not pass --resume; got %v", args)
	}

	// A valid UUID id is resumed via --resume; explicit model via --model; named RC.
	id := generateUUID()
	args = resumeClaudeCmd(id, "opus", "my-session")
	if !containsPair(args, "--resume", id) {
		t.Errorf("resumeClaudeCmd(uuid, ...) should pass --resume %s; got %v", id, args)
	}
	if !containsPair(args, "--model", "opus") {
		t.Errorf("resumeClaudeCmd(..., opus) should pass --model opus; got %v", args)
	}
	if !containsPair(args, "--remote-control", "my-session") {
		t.Errorf("resumeClaudeCmd should name remote control after the session; got %v", args)
	}

	// A non-UUID id (legacy) starts a fresh conversation — no --resume.
	args = resumeClaudeCmd("legacy-abc", "opus", "n")
	if hasFlag(args, "--resume") {
		t.Errorf("resumeClaudeCmd with non-UUID id should not pass --resume; got %v", args)
	}
}

func TestRemoteControlArgs(t *testing.T) {
	t.Setenv("SWARMOPS_DISABLE_REMOTE_CONTROL", "") // force enabled regardless of ambient env
	if got := remoteControlArgs("smart-foo"); len(got) != 2 || got[0] != "--remote-control" || got[1] != "smart-foo" {
		t.Errorf("named: got %v", got)
	}
	// Empty or dash-leading names fall back to the bare flag (avoid swallowing next flag).
	for _, n := range []string{"", "  ", "-x"} {
		if got := remoteControlArgs(n); len(got) != 1 || got[0] != "--remote-control" {
			t.Errorf("bare fallback for %q: got %v", n, got)
		}
	}
}

// containsPair returns true if args contains flag immediately followed by value.
func containsPair(args []string, flag, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

// hasFlag returns true if args contains the given flag.
func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

// Sanity: defaultSessionModelFallback should be a non-empty Claude alias.
func TestDefaultModelFallbackIsValidAlias(t *testing.T) {
	if defaultSessionModelFallback == "" {
		t.Fatal("defaultSessionModelFallback must not be empty")
	}
	// Known claude aliases as of 2026-05.
	valid := []string{"haiku", "sonnet", "opus"}
	for _, v := range valid {
		if strings.EqualFold(defaultSessionModelFallback, v) {
			return
		}
	}
	t.Errorf("defaultSessionModelFallback = %q, want one of %v", defaultSessionModelFallback, valid)
}
