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

func TestResumeClaudeCmdAlwaysPassesModel(t *testing.T) {
	t.Setenv("SWARMOPS_DEFAULT_MODEL", "")

	args := resumeClaudeCmd("", "")
	if !containsPair(args, "--model", "sonnet") {
		t.Errorf("resumeClaudeCmd with empty model should include --model sonnet; got %v", args)
	}

	args = resumeClaudeCmd("happier-abc", "opus")
	if !containsPair(args, "--model", "opus") {
		t.Errorf("resumeClaudeCmd should preserve explicit model; got %v", args)
	}
	if !containsPair(args, "--existing-session", "happier-abc") {
		t.Errorf("resumeClaudeCmd should pass --existing-session for non-UUID id; got %v", args)
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
