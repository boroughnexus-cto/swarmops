package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// ─── Per-repo mutex registry ────────────────────────────────────────────────
// Serializes git worktree operations per repository root to avoid git index
// lock contention when multiple agents spawn from the same repo concurrently.

var (
	registryMu    sync.Mutex
	worktreeLocks = make(map[string]*sync.Mutex)
)

// repoMu returns the mutex for a canonicalized repoPath, creating one if needed.
func repoMu(repoPath string) *sync.Mutex {
	key := filepath.Clean(repoPath)
	registryMu.Lock()
	defer registryMu.Unlock()
	if m, ok := worktreeLocks[key]; ok {
		return m
	}
	m := &sync.Mutex{}
	worktreeLocks[key] = m
	return m
}

// ─── Worktree operations ─────────────────────────────────────────────────────

// createWorktree creates a git worktree at worktreePath on a new branch.
// The entire operation is serialized per repo to avoid git lock contention.
func createWorktree(repoPath, worktreePath, branch string) error {
	mu := repoMu(repoPath)
	mu.Lock()
	defer mu.Unlock()

	out, err := exec.Command(
		"git", "-C", repoPath, "worktree", "add", "-b", branch, "--", worktreePath,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %w: %s", err, out)
	}
	return nil
}

// removeWorktree removes a git worktree and optionally deletes the branch.
// If worktreePath no longer exists on disk, the removal step is skipped
// (idempotent). The full operation (stat check + remove + branch delete) is
// serialized under a single lock to avoid TOCTOU races.
// force=true should be used during teardown (agent may have uncommitted work);
// force=false is safe for rollback of a freshly-created empty worktree.
func removeWorktree(repoPath, worktreePath, branch string, force, deleteBranch bool) error {
	mu := repoMu(repoPath)
	mu.Lock()
	defer mu.Unlock()

	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		args := []string{"-C", repoPath, "worktree", "remove"}
		if force {
			args = append(args, "--force")
		}
		args = append(args, "--", worktreePath)
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("git worktree remove: %w: %s", err, out)
		}
	}

	if deleteBranch && branch != "" {
		if out, err := exec.Command("git", "-C", repoPath, "branch", "-d", branch).CombinedOutput(); err != nil {
			// Non-fatal: branch may have unmerged commits or may already be gone.
			// Log but do not propagate — the worktree is already removed.
			_ = out
		}
	}
	return nil
}

// ─── TASK.md ─────────────────────────────────────────────────────────────────

// writeTaskMD writes the task brief to TASK.md inside dir.
// Creates dir if it does not exist. Overwrites any existing TASK.md.
func writeTaskMD(dir, brief string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return os.WriteFile(filepath.Join(dir, "TASK.md"), []byte(brief), 0o644)
}

// ─── Path helpers ─────────────────────────────────────────────────────────────

// defaultWorktreeBase returns the default parent directory for worktrees,
// adjacent to the repo root: /home/user/myproject → /home/user/myproject-worktrees
func defaultWorktreeBase(repoPath string) string {
	return filepath.Clean(repoPath) + "-worktrees"
}

var slugRe = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// sanitizeSlug converts a session name into a safe branch-name component.
// Non-alphanumeric characters are replaced with dashes, result is capped at 40 chars.
func sanitizeSlug(name string) string {
	s := slugRe.ReplaceAllString(name, "-")
	if len(s) > 40 {
		s = s[:40]
	}
	if s == "" {
		s = "agent"
	}
	return s
}

// resolveWorktreePath builds a unique worktree path under base.
// Format: base/<sanitized-name>-<4-char-hex>
func resolveWorktreePath(base, sessionName string) string {
	return filepath.Join(base, sanitizeSlug(sessionName)+"-"+generateID()[:4])
}

// worktreeBranchName derives the git branch name from the worktree directory name.
// Format: agent/<dir-name>
func worktreeBranchName(worktreePath string) string {
	return "agent/" + filepath.Base(worktreePath)
}

// ─── Input validation ─────────────────────────────────────────────────────────

// validateRepoPath resolves repoPath to an absolute path, confirms it exists,
// and verifies it is a git repository. Returns the absolute path.
func validateRepoPath(repoPath string) (string, error) {
	abs, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("path does not exist: %s", abs)
	}
	out, err := exec.Command("git", "-C", abs, "rev-parse", "--git-dir").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %s", abs)
	}
	_ = out
	return abs, nil
}

// validateWorktreePath resolves worktreePath to an absolute path and rejects
// any path that contains ".." traversal components in the raw input.
func validateWorktreePath(worktreePath string) (string, error) {
	// Reject raw ".." traversal before Abs resolution so callers cannot
	// escape an intended subtree via relative-path tricks.
	for _, part := range strings.Split(filepath.ToSlash(worktreePath), "/") {
		if part == ".." {
			return "", fmt.Errorf("worktree_path must not contain '..' traversal")
		}
	}
	abs, err := filepath.Abs(worktreePath)
	if err != nil {
		return "", fmt.Errorf("invalid worktree path: %w", err)
	}
	return abs, nil
}
