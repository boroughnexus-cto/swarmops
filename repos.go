package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// KnownRepo is one entry in the SwarmOps repo registry. The registry feeds
// swop_newgoal so the brain can pick the right working directory for a goal
// instead of every session starting in $HOME and re-discovering everything.
type KnownRepo struct {
	ID              int64  `json:"id"`
	Owner           string `json:"owner"`
	Name            string `json:"name"`
	LocalPath       string `json:"local_path,omitempty"` // empty if not cloned locally
	DefaultBranch   string `json:"default_branch,omitempty"`
	Description     string `json:"description,omitempty"`
	ClaudeMDSummary string `json:"claude_md_summary,omitempty"`
	LastScannedAt   int64  `json:"last_scanned_at"`
}

// Slug returns "owner/name" — the canonical id used in brain prompts.
func (r *KnownRepo) Slug() string { return r.Owner + "/" + r.Name }

// IsCloned reports whether the repo has a working tree on disk.
func (r *KnownRepo) IsCloned() bool { return r.LocalPath != "" }

// repoDiscoveryRoots returns the filesystem roots scanned for local clones.
// Default ~/git-bnx; override via SWARMOPS_REPO_ROOTS=path1:path2.
func repoDiscoveryRoots() []string {
	if v := strings.TrimSpace(os.Getenv("SWARMOPS_REPO_ROOTS")); v != "" {
		out := []string{}
		for _, p := range strings.Split(v, ":") {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		return []string{filepath.Join(home, "git-bnx")}
	}
	return nil
}

// repoDiscoveryOrgs returns the GitHub orgs / users queried via `gh repo list`.
// Configurable via SWARMOPS_GITHUB_ORGS=org1,org2; defaults match the homelab
// orgs in active use (ThomkerNet for personal / TKN, boroughnexus-cto for BNX).
func repoDiscoveryOrgs() []string {
	if v := strings.TrimSpace(os.Getenv("SWARMOPS_GITHUB_ORGS")); v != "" {
		out := []string{}
		for _, o := range strings.Split(v, ",") {
			if o = strings.TrimSpace(o); o != "" {
				out = append(out, o)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return []string{"ThomkerNet", "boroughnexus-cto"}
}

// listKnownRepos returns every entry, ordered by owner/name. Used by the
// swop_newgoal brain prompt and the swop_list_repos read tool.
func listKnownRepos(ctx context.Context) ([]KnownRepo, error) {
	rows, err := database.QueryContext(ctx,
		`SELECT id, owner, name, COALESCE(local_path,''), COALESCE(default_branch,''),
		        COALESCE(description,''), COALESCE(claude_md_summary,''), last_scanned_at
		 FROM known_repos ORDER BY owner, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []KnownRepo
	for rows.Next() {
		var r KnownRepo
		if err := rows.Scan(&r.ID, &r.Owner, &r.Name, &r.LocalPath, &r.DefaultBranch,
			&r.Description, &r.ClaudeMDSummary, &r.LastScannedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// upsertKnownRepo inserts or updates a known_repos row keyed by (owner, name).
// last_scanned_at is set to the current unix time.
func upsertKnownRepo(ctx context.Context, r *KnownRepo) error {
	r.LastScannedAt = time.Now().Unix()
	_, err := database.ExecContext(ctx, `
		INSERT INTO known_repos (owner, name, local_path, default_branch, description, claude_md_summary, last_scanned_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(owner, name) DO UPDATE SET
		    local_path = COALESCE(NULLIF(excluded.local_path, ''), known_repos.local_path),
		    default_branch = COALESCE(NULLIF(excluded.default_branch, ''), known_repos.default_branch),
		    description = COALESCE(NULLIF(excluded.description, ''), known_repos.description),
		    claude_md_summary = COALESCE(NULLIF(excluded.claude_md_summary, ''), known_repos.claude_md_summary),
		    last_scanned_at = excluded.last_scanned_at
	`, r.Owner, r.Name, r.LocalPath, r.DefaultBranch, r.Description, r.ClaudeMDSummary, r.LastScannedAt)
	return err
}

// reposRefreshAll re-runs both discovery passes (local filesystem + remote
// GitHub via gh CLI) and upserts every result. Returns (localFound, remoteFound, err).
// Errors from a single source are logged but don't abort the other source — a
// missing gh CLI shouldn't prevent local discovery from working.
func reposRefreshAll(ctx context.Context) (int, int, error) {
	localCount, localErr := discoverLocalRepos(ctx)
	if localErr != nil {
		log.Printf("repos: local discovery: %v", localErr)
	}
	remoteCount, remoteErr := discoverGitHubRepos(ctx)
	if remoteErr != nil {
		log.Printf("repos: github discovery: %v", remoteErr)
	}
	if localErr != nil && remoteErr != nil {
		return localCount, remoteCount, fmt.Errorf("local: %v; github: %v", localErr, remoteErr)
	}
	return localCount, remoteCount, nil
}

// githubURLRegex matches the owner/name parts of an https or ssh GitHub URL.
//
//	https://github.com/owner/repo[.git]
//	git@github.com:owner/repo[.git]
var githubURLRegex = regexp.MustCompile(`(?:github\.com[:/])([^/]+)/([^/]+?)(?:\.git)?$`)

// discoverLocalRepos walks each configured root looking for .git directories,
// extracts the GitHub owner/repo from `git remote get-url origin`, reads the
// description (first README.md paragraph) and CLAUDE.md summary, and upserts.
// Caps the walk depth so deep node_modules trees don't dominate the scan.
func discoverLocalRepos(ctx context.Context) (int, error) {
	roots := repoDiscoveryRoots()
	if len(roots) == 0 {
		return 0, nil
	}
	count := 0
	for _, root := range roots {
		if _, err := os.Stat(root); err != nil {
			continue
		}
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if walkErr != nil {
				return nil // skip unreadable subtrees
			}
			if !d.IsDir() {
				return nil
			}
			// Limit walk depth: each `/` past root adds a level. Stop at depth 5.
			rel := strings.TrimPrefix(path, root)
			if strings.Count(rel, string(os.PathSeparator)) > 5 {
				return filepath.SkipDir
			}
			// Skip common irrelevant dirs.
			base := d.Name()
			if base == "node_modules" || base == ".venv" || base == "venv" ||
				base == "target" || base == ".cache" || base == "dist" ||
				base == "build" || base == "__pycache__" {
				return filepath.SkipDir
			}
			gitDir := filepath.Join(path, ".git")
			info, err := os.Stat(gitDir)
			if err != nil || !info.IsDir() {
				return nil
			}
			// Found a repo root. Don't descend further into it.
			defer func() {}() // SkipDir below
			r := scanLocalRepo(path)
			if r != nil {
				if err := upsertKnownRepo(ctx, r); err != nil {
					log.Printf("repos: upsert %s: %v", path, err)
				} else {
					count++
				}
			}
			return filepath.SkipDir
		})
		if err != nil {
			return count, err
		}
	}
	return count, nil
}

// scanLocalRepo reads a single repo at path and returns its KnownRepo record,
// or nil if the repo doesn't have a GitHub origin we can parse.
func scanLocalRepo(path string) *KnownRepo {
	originURL, err := runGit(path, "remote", "get-url", "origin")
	if err != nil {
		return nil
	}
	originURL = strings.TrimSpace(originURL)
	m := githubURLRegex.FindStringSubmatch(originURL)
	if m == nil {
		return nil
	}
	owner, name := m[1], m[2]
	defaultBranch, _ := runGit(path, "symbolic-ref", "--short", "HEAD")
	defaultBranch = strings.TrimSpace(defaultBranch)
	return &KnownRepo{
		Owner:           owner,
		Name:            name,
		LocalPath:       path,
		DefaultBranch:   defaultBranch,
		Description:     readReadmeSummary(path),
		ClaudeMDSummary: readClaudeMDSummary(path),
	}
}

// runGit runs `git -C dir <args...>` and returns stdout, suppressing stderr
// from polluting our logs. Non-zero exits surface as an error.
func runGit(dir string, args ...string) (string, error) {
	full := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", full...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// readReadmeSummary extracts the first non-empty paragraph from README.md.
// Returns "" if the file is missing or unreadable. Capped at 400 chars so
// big READMEs don't bloat the registry / brain prompt.
func readReadmeSummary(path string) string {
	for _, candidate := range []string{"README.md", "README", "readme.md"} {
		p := filepath.Join(path, candidate)
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		var lines []string
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" && len(lines) > 0 {
				break
			}
			// Skip leading markdown heading marker; we want the prose summary,
			// not the title.
			if len(lines) == 0 && strings.HasPrefix(line, "#") {
				continue
			}
			if line != "" {
				lines = append(lines, line)
			}
		}
		s := strings.Join(lines, " ")
		if len(s) > 400 {
			s = s[:400] + "…"
		}
		return s
	}
	return ""
}

// readClaudeMDSummary returns an excerpt from the repo's CLAUDE.md, biased
// toward the section that describes the repo's purpose / context boundary.
// Returns "" if no CLAUDE.md is present.
func readClaudeMDSummary(path string) string {
	p := filepath.Join(path, "CLAUDE.md")
	data, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	text := string(data)
	// Prefer a "Context Boundary" or "## What Lives Here" section if present,
	// since those typically describe scope succinctly.
	for _, anchor := range []string{
		"## Context Boundary", "## What Lives Here", "# Context",
	} {
		if i := strings.Index(text, anchor); i >= 0 {
			tail := text[i:]
			// Stop at the next H2 to keep the excerpt focused.
			if end := strings.Index(tail[len(anchor):], "\n## "); end > 0 {
				tail = tail[:len(anchor)+end]
			}
			if len(tail) > 600 {
				tail = tail[:600] + "…"
			}
			return strings.TrimSpace(tail)
		}
	}
	// Fall back to the first 500 chars.
	if len(text) > 500 {
		text = text[:500] + "…"
	}
	return strings.TrimSpace(text)
}

// ghRepoListItem matches the JSON shape returned by `gh repo list --json …`.
type ghRepoListItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
	DefaultBranchRef struct {
		Name string `json:"name"`
	} `json:"defaultBranchRef"`
}

// discoverGitHubRepos queries each configured org via gh CLI and upserts every
// repo into the registry. Repos already known with a local_path keep their
// path; new repos are inserted with local_path = "" (uncloned).
func discoverGitHubRepos(ctx context.Context) (int, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return 0, fmt.Errorf("gh CLI not available: %w", err)
	}
	orgs := repoDiscoveryOrgs()
	count := 0
	for _, org := range orgs {
		cmd := exec.CommandContext(ctx, "gh", "repo", "list", org,
			"--limit", "200",
			"--json", "name,description,owner,defaultBranchRef")
		out, err := cmd.Output()
		if err != nil {
			log.Printf("repos: gh repo list %s: %v", org, err)
			continue
		}
		var items []ghRepoListItem
		if err := json.Unmarshal(out, &items); err != nil {
			log.Printf("repos: parse gh output for %s: %v", org, err)
			continue
		}
		for _, it := range items {
			r := KnownRepo{
				Owner:         it.Owner.Login,
				Name:          it.Name,
				DefaultBranch: it.DefaultBranchRef.Name,
				Description:   it.Description,
			}
			if err := upsertKnownRepo(ctx, &r); err != nil {
				log.Printf("repos: upsert %s/%s: %v", r.Owner, r.Name, err)
				continue
			}
			count++
		}
	}
	return count, nil
}
