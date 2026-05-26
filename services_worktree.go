package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// validateRepoPath cleans and validates that path is an absolute path to a git repository.
func validateRepoPath(repoPath string) (string, error) {
	path := filepath.Clean(repoPath)
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("repo_path must be absolute: %q", repoPath)
	}
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("repo_path does not exist: %q", path)
	}
	if err := exec.Command("git", "-C", path, "rev-parse", "--git-dir").Run(); err != nil {
		return "", fmt.Errorf("not a git repository: %q", path)
	}
	return path, nil
}

// validateWorktreePath cleans and validates that path is absolute.
// Does not require the path to exist (it will be created by createWorktree).
func validateWorktreePath(worktreePath string) (string, error) {
	path := filepath.Clean(worktreePath)
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("worktree_path must be absolute: %q", worktreePath)
	}
	return path, nil
}

// defaultWorktreeBase returns the default base directory for new worktrees within repoPath.
func defaultWorktreeBase(repoPath string) string {
	return filepath.Join(repoPath, ".worktrees")
}

// resolveWorktreePath returns the worktree path for the given name under base.
func resolveWorktreePath(base, name string) string {
	return filepath.Join(base, name)
}

// worktreeBranchName derives a git branch name from the worktree path.
func worktreeBranchName(worktreePath string) string {
	return "agent/" + filepath.Base(worktreePath)
}

// createWorktree creates a new git worktree at worktreePath on a new branch.
func createWorktree(repoPath, worktreePath, branch string) error {
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0o755); err != nil {
		return fmt.Errorf("create worktree parent dir: %w", err)
	}
	out, err := exec.Command("git", "-C", repoPath, "worktree", "add", "-b", branch, worktreePath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// removeWorktree removes a git worktree and optionally deletes its branch.
// force removes the worktree even if it has uncommitted changes.
// deleteBranch deletes the branch after the worktree is removed.
func removeWorktree(repoPath, worktreePath, branch string, force, deleteBranch bool) error {
	args := []string{"-C", repoPath, "worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree remove: %v: %s", err, strings.TrimSpace(string(out)))
	}
	if deleteBranch && branch != "" {
		// Best-effort; ignore errors (branch may already be deleted or not local).
		exec.Command("git", "-C", repoPath, "branch", "-d", branch).Run()
	}
	return nil
}

// writeTaskMD writes taskBrief to TASK.md in directory.
func writeTaskMD(directory, taskBrief string) error {
	content := "# Task\n\n" + taskBrief + "\n"
	return os.WriteFile(filepath.Join(directory, "TASK.md"), []byte(content), 0o644)
}
