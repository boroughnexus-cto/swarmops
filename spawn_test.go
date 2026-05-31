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

	// With happier available, model is passed via ANTHROPIC_MODEL env var
	args := resumeClaudeCmd("", "")
	if !containsEnv(args, "ANTHROPIC_MODEL=sonnet") {
		t.Errorf("resumeClaudeCmd with empty model should set ANTHROPIC_MODEL=sonnet; got %v", args)
	}

	args = resumeClaudeCmd("happier-abc", "opus")
	if !containsEnv(args, "ANTHROPIC_MODEL=opus") {
		t.Errorf("resumeClaudeCmd should preserve explicit model via env; got %v", args)
	}
	// happier --yolo does not use --existing-session flag
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

// containsEnv returns true if args contains an env KEY=value pair.
func containsEnv(args []string, kv string) bool {
	for _, arg := range args {
		if arg == kv {
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
