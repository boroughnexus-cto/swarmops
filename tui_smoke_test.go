package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestSmoke_ProgramStartsAndQuits verifies that a tea.Program built from
// initialModel starts and shuts down cleanly. Uses a goroutine + p.Quit()
// to avoid any ESC-sequence parsing ambiguity or Init timing assumptions.
// p.Kill() is registered as cleanup to prevent goroutine leaks if the test times out.
func TestSmoke_ProgramStartsAndQuits(t *testing.T) {
	client := &fakeSwarmClient{}
	model := initialModel(client)

	p := tea.NewProgram(
		model,
		tea.WithoutRenderer(),
		tea.WithInput(bytes.NewReader(nil)),
		tea.WithOutput(io.Discard),
	)

	t.Cleanup(func() { p.Kill() })

	errCh := make(chan error, 1)
	go func() {
		_, err := p.Run()
		errCh <- err
	}()

	// p.Quit() is safe to call before Run() enters its event loop — the quit
	// message is buffered in the program's channel and processed on startup.
	time.Sleep(10 * time.Millisecond)
	p.Quit()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("program exited with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("program did not quit within 5 seconds")
	}
}

// TestSmoke_InitialModelState verifies that initialModel produces a valid
// zero state without requiring a running backend.
func TestSmoke_InitialModelState(t *testing.T) {
	client := &fakeSwarmClient{}
	m := initialModel(client)

	if m.mode != modePassthrough {
		t.Errorf("initial mode should be modePassthrough, got %d", m.mode)
	}
	if m.cursor != 0 {
		t.Errorf("initial cursor should be 0, got %d", m.cursor)
	}
	// items must be nil (not just empty) at init — rendering branches on this
	if m.items != nil {
		t.Errorf("initial items should be nil, got len=%d", len(m.items))
	}
}

// TestSmoke_ViewDoesNotPanic calls View on a zero-state model to confirm
// it handles the pre-WindowSizeMsg state gracefully.
func TestSmoke_ViewDoesNotPanic(t *testing.T) {
	client := &fakeSwarmClient{}
	m := initialModel(client)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("View() panicked on zero-state model: %v", r)
		}
	}()

	out := m.View()
	if out == "" {
		t.Error("View() returned empty string on zero-state model")
	}
}

// TestSmoke_ViewAfterResize verifies the normal render path: send a
// WindowSizeMsg then call View(). This is the state users actually see.
func TestSmoke_ViewAfterResize(t *testing.T) {
	client := &fakeSwarmClient{}
	m := initialModel(client)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	resized, ok := updated.(tuiModel)
	if !ok {
		t.Fatalf("Update(WindowSizeMsg) returned %T, want tuiModel", updated)
	}
	m = resized

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("View() panicked after WindowSizeMsg: %v", r)
		}
	}()

	out := m.View()
	if out == "" {
		t.Error("View() returned empty string after WindowSizeMsg")
	}
	if !strings.Contains(stripAnsi(out), "SwarmOps") {
		preview := out
		if len(preview) > 200 {
			preview = preview[:200]
		}
		t.Errorf("View() after resize should contain SwarmOps title, got: %q", preview)
	}
}
