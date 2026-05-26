CREATE TABLE IF NOT EXISTS known_repos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner TEXT NOT NULL,
    name TEXT NOT NULL,
    local_path TEXT,
    default_branch TEXT,
    description TEXT,
    claude_md_summary TEXT,
    last_scanned_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE(owner, name)
);

CREATE INDEX IF NOT EXISTS idx_known_repos_local_path ON known_repos(local_path);
