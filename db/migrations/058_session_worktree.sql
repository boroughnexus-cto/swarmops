ALTER TABLE managed_sessions ADD COLUMN worktree_path TEXT;
ALTER TABLE managed_sessions ADD COLUMN git_branch TEXT;
ALTER TABLE managed_sessions ADD COLUMN repo_path TEXT;
