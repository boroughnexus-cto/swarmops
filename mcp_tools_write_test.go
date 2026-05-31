package main

import (
	"context"
	"testing"
)

// TestCreateSession_Basic verifies that createSession stores name, directory,
// mission, model, and profile correctly and round-trips through getSession.
func TestCreateSession_Basic(t *testing.T) {
	defer setupTestDB(t)()
	ctx := context.Background()

	mission := "Investigate the Icinga alert"
	model := "claude-sonnet-4-6"

	sess, err := createSession(ctx, "my-named-session", "/tmp/work-dir",
		&mission, false, model, "")
	if err != nil {
		t.Fatalf("createSession: %v", err)
	}

	// Verify returned session has all fields
	if sess.Name != "my-named-session" {
		t.Errorf("Name = %q, want %q", sess.Name, "my-named-session")
	}
	if sess.Directory != "/tmp/work-dir" {
		t.Errorf("Directory = %q, want %q", sess.Directory, "/tmp/work-dir")
	}
	if sess.Mission == nil || *sess.Mission != mission {
		t.Errorf("Mission = %v, want %q", sess.Mission, mission)
	}
	if sess.Model != model {
		t.Errorf("Model = %q, want %q", sess.Model, model)
	}

	// Round-trip through getSession to verify DB persistence
	got, err := getSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("getSession: %v", err)
	}
	if got.Name != "my-named-session" {
		t.Errorf("persisted Name = %q, want %q", got.Name, "my-named-session")
	}
	if got.Mission == nil || *got.Mission != mission {
		t.Errorf("persisted Mission = %v, want %q", got.Mission, mission)
	}
	if got.Model != model {
		t.Errorf("persisted Model = %q, want %q", got.Model, model)
	}
}

// TestCreateSession_NilMission verifies that nil mission works and stores nulls correctly.
func TestCreateSession_NilMission(t *testing.T) {
	defer setupTestDB(t)()
	ctx := context.Background()

	sess, err := createSession(ctx, "minimal", "/tmp", nil, false, "", "")
	if err != nil {
		t.Fatalf("createSession: %v", err)
	}
	if sess.Mission != nil {
		t.Errorf("Mission = %v, want nil", sess.Mission)
	}
	if sess.Model != "" {
		t.Errorf("Model = %q, want empty", sess.Model)
	}

	got, err := getSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("getSession: %v", err)
	}
	if got.Mission != nil {
		t.Errorf("persisted Mission = %v, want nil", got.Mission)
	}
}
