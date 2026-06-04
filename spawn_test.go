package main

import (
	"strings"
	"testing"
)

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
	// Known happier/claude aliases as of 2026-05.
	valid := []string{"haiku", "sonnet", "opus"}
	for _, v := range valid {
		if strings.EqualFold(defaultSessionModelFallback, v) {
			return
		}
	}
	t.Errorf("defaultSessionModelFallback = %q, want one of %v", defaultSessionModelFallback, valid)
}
