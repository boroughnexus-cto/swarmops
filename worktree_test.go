package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// gitExec runs a git command inside dir for test setup.
func gitExec(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

// initGitRepo creates a git repo with a single initial commit in dir.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitExec(t, dir, "init")
	gitExec(t, dir, "config", "user.email", "test@test.com")
	gitExec(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	gitExec(t, dir, "add", ".")
	gitExec(t, dir, "commit", "-m", "init")
}

// ─── sanitizeSlug ────────────────────────────────────────────────────────────

func TestSanitizeSlug(t *testing.T) {
	cases := []struct{ in, want string }{
		{"simple", "simple"},
		{"hello world", "hello-world"},
		{"SWM-58/worktree", "SWM-58-worktree"},
		{"", "agent"},
		{"a b!c@d#", "a-b-c-d-"},
	}
	for _, c := range cases {
		got := sanitizeSlug(c.in)
		if got != c.want {
			t.Errorf("sanitizeSlug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSanitizeSlug_MaxLength(t *testing.T) {
	long := strings.Repeat("a", 60)
	got := sanitizeSlug(long)
	if len(got) > 40 {
		t.Errorf("slug too long: %d chars", len(got))
	}
}

// ─── writeTaskMD ─────────────────────────────────────────────────────────────

func TestWriteTaskMD(t *testing.T) {
	dir := t.TempDir()
	brief := "# Task\nDo the thing.\n"
	if err := writeTaskMD(dir, brief); err != nil {
		t.Fatalf("writeTaskMD: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "TASK.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != brief {
		t.Errorf("TASK.md content mismatch: got %q, want %q", got, brief)
	}

	// Idempotent overwrite.
	updated := "# Updated\n"
	if err := writeTaskMD(dir, updated); err != nil {
		t.Fatalf("writeTaskMD overwrite: %v", err)
	}
	got2, _ := os.ReadFile(filepath.Join(dir, "TASK.md"))
	if string(got2) != updated {
		t.Errorf("TASK.md overwrite mismatch: got %q, want %q", got2, updated)
	}
}

func TestWriteTaskMD_CreatesDir(t *testing.T) {
	base := t.TempDir()
	nested := filepath.Join(base, "a", "b", "c")
	if err := writeTaskMD(nested, "hi"); err != nil {
		t.Fatalf("writeTaskMD in non-existent dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(nested, "TASK.md")); err != nil {
		t.Errorf("TASK.md not created: %v", err)
	}
}

// ─── defaultWorktreeBase / resolveWorktreePath ────────────────────────────────

func TestDefaultWorktreeBase(t *testing.T) {
	got := defaultWorktreeBase("/home/user/myproject")
	want := "/home/user/myproject-worktrees"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveWorktreePath(t *testing.T) {
	base := "/home/user/myproject-worktrees"
	path := resolveWorktreePath(base, "my session name")
	if !strings.HasPrefix(path, base) {
		t.Errorf("path %q does not start with base %q", path, base)
	}
	dir := filepath.Base(path)
	if dir == "" || dir == "." {
		t.Error("empty dir component")
	}
}

// ─── validateRepoPath / validateWorktreePath ──────────────────────────────────

func TestValidateRepoPath_NotGit(t *testing.T) {
	dir := t.TempDir()
	_, err := validateRepoPath(dir)
	if err == nil {
		t.Error("expected error for non-git directory, got nil")
	}
}

func TestValidateRepoPath_IsGit(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	got, err := validateRepoPath(dir)
	if err != nil {
		t.Fatalf("validateRepoPath on real git repo: %v", err)
	}
	if got == "" {
		t.Error("expected non-empty absolute path")
	}
}

func TestValidateWorktreePath_TraversalRejection(t *testing.T) {
	_, err := validateWorktreePath("../../etc/passwd")
	if err == nil {
		t.Error("expected error for path traversal, got nil")
	}
}

func TestValidateWorktreePath_Valid(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "worktree")
	got, err := validateWorktreePath(target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("expected non-empty path")
	}
}

// ─── createWorktree / removeWorktree ─────────────────────────────────────────

func TestCreateAndRemoveWorktree(t *testing.T) {
	repoDir := t.TempDir()
	initGitRepo(t, repoDir)

	worktreeDir := filepath.Join(t.TempDir(), "wt1")
	branch := "agent/test-wt1"

	if err := createWorktree(repoDir, worktreeDir, branch); err != nil {
		t.Fatalf("createWorktree: %v", err)
	}
	if _, err := os.Stat(worktreeDir); err != nil {
		t.Fatalf("worktree dir not created: %v", err)
	}

	if err := removeWorktree(repoDir, worktreeDir, branch, false, true); err != nil {
		t.Fatalf("removeWorktree: %v", err)
	}
	if _, err := os.Stat(worktreeDir); !os.IsNotExist(err) {
		t.Error("worktree dir still exists after removal")
	}
}

func TestRemoveWorktree_Idempotent(t *testing.T) {
	repoDir := t.TempDir()
	initGitRepo(t, repoDir)

	worktreeDir := filepath.Join(t.TempDir(), "wt-idem")
	branch := "agent/test-idem"

	if err := createWorktree(repoDir, worktreeDir, branch); err != nil {
		t.Fatalf("createWorktree: %v", err)
	}
	if err := removeWorktree(repoDir, worktreeDir, branch, false, false); err != nil {
		t.Fatalf("first remove: %v", err)
	}
	// Second remove should be a no-op (path already gone).
	if err := removeWorktree(repoDir, worktreeDir, branch, false, false); err != nil {
		t.Fatalf("second remove (idempotent): %v", err)
	}
}

func TestCreateWorktree_Concurrent(t *testing.T) {
	repoDir := t.TempDir()
	initGitRepo(t, repoDir)

	const n = 4
	errs := make([]error, n)
	var wg sync.WaitGroup
	dirs := make([]string, n)
	for i := range n {
		dirs[i] = filepath.Join(t.TempDir(), fmt.Sprintf("wt%d", i))
	}
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			branch := fmt.Sprintf("agent/concurrent-%d", i)
			errs[i] = createWorktree(repoDir, dirs[i], branch)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: %v", i, err)
		}
	}
}

// ─── repoMu canonicalization ──────────────────────────────────────────────────

func TestRepoMu_TrailingSlashSameKey(t *testing.T) {
	mu1 := repoMu("/home/user/repo")
	mu2 := repoMu("/home/user/repo/")
	if mu1 != mu2 {
		t.Error("trailing slash produced different mutex — canonicalization broken")
	}
}
