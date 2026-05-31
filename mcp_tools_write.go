package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// registerWriteTools registers all write MCP tools and conditionally pool tools.
func registerWriteTools(reg *ToolRegistry, svc *Services, enablePoolTools bool) {
	// ─── swop_run_task ────────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_run_task",
			Description: "[WRITE] Create and start a new Claude Code session in a tmux window.",
			InputSchema: jsonSchema(map[string]interface{}{
				"name":         stringProp("Session name (auto-generated if empty)"),
				"directory":    stringProp("Working directory for the session (default: current directory)"),
				"mission":      stringProp("Optional mission statement (1-3 sentences describing session purpose)"),
				"context_id":   stringProp("Optional tkn-context ID to attach to the session"),
				"context_name": stringProp("Optional human-readable context label (display only — paired with context_id)"),
				"model":        stringProp("Optional model name or alias (e.g. 'sonnet', 'opus', 'claude-sonnet-4-6'). Defaults to system setting."),
				"profile":      stringProp("Optional happier backend profile (e.g. 'deepseek', 'openai', 'gemini'). Defaults to 'anthropic'. Takes effect on next restart if session already exists."),
				"task_brief":   stringProp("Optional task brief written to TASK.md in the working directory before the agent starts."),
				"env_overrides": map[string]interface{}{
					"type":        "object",
					"description": "Optional environment variable overrides injected into the session. Omit for the default Anthropic API. To route through LiteLLM, set ANTHROPIC_BASE_URL+ANTHROPIC_API_KEY (or use the TUI [gpt]/[dseek] picker entries which do this for you). Pick a backend with ANTHROPIC_MODEL: 'chatgptsub-gpt-5.5' → [gpt], 'or-deepseek-v4-pro' → [dseek]. Session names get the matching auto-prefix. Keys and values must be strings.",
					"additionalProperties": map[string]interface{}{"type": "string"},
				},
				"mcp_servers": map[string]interface{}{
					"type":        "array",
					"description": "Optional subset of MCP server names (from ~/.swarmops/mcp-config.json) to load in the session. Omit/empty = the default ~50-server catalogue (large; ~300k tokens of system+tools). Set to a small focused list (e.g. ['tkn-plane','tkn-komodo']) for [gpt]/[dseek] sessions to keep the system prompt tiny so DeepSeek/GPT-5.5 don't burn context. Implemented via --strict-mcp-config + a per-session config file.",
					"items":       map[string]interface{}{"type": "string"},
				},
			}, nil),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			name := getStringArg(args, "name", "")
			directory := getStringArg(args, "directory", "")
			model := getStringArg(args, "model", "")
			taskBrief := getStringArg(args, "task_brief", "")

			missionStr := getStringArg(args, "mission", "")
			var mission *string
			if missionStr != "" {
				mission = &missionStr
			}
			envOverrides := getStringMapArg(args, "env_overrides")
			envOverrides = stashMCPServers(args, envOverrides)

			// Auto-prepend [gpt] or [dseek] prefix when routing to a non-Anthropic backend.
			// Pick the prefix from ANTHROPIC_MODEL first (the explicit signal), falling
			// back to the legacy model arg for older callers.
			prefixModel := envOverrides["ANTHROPIC_MODEL"]
			if prefixModel == "" {
				prefixModel = model
			}
			name = autoPrefixSessionName(name, prefixModel, envOverrides)

			profile := getStringArg(args, "profile", "")
			return svc.RunTask(ctx, name, directory, mission, model, profile, taskBrief, envOverrides)
		},
	)

	// ─── swop_spawn_agent ─────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name: "swop_spawn_agent",
			Description: "[WRITE] Atomically create a git worktree and spawn a Claude Code agent inside it. " +
				"Writes an optional TASK.md brief before the agent starts. " +
				"Rolls back the worktree on any failure. " +
				"Use swop_teardown_agent to clean up when done.",
			InputSchema: jsonSchema(map[string]interface{}{
				"repo_path":     stringProp("Absolute path to the git repository root (required)."),
				"name":          stringProp("Agent session name (auto-generated if empty). Sessions routed via LiteLLM GPT should use a [gpt] prefix, e.g. '[gpt] my-task'."),
				"worktree_path": stringProp("Full path for the worktree directory (auto-generated under <repo>-worktrees/ if empty)."),
				"branch":        stringProp("Git branch name (defaults to agent/<slug>-<hex> derived from worktree path)."),
				"mission":       stringProp("Optional mission statement (1-3 sentences)."),
				"task_brief":    stringProp("Optional task brief written to TASK.md in the worktree before the agent starts."),
				"context_id":    stringProp("Optional tkn-context ID to attach to the session."),
				"context_name":  stringProp("Optional human-readable context label (display only)."),
				"model":         stringProp("Optional model name or alias (e.g. 'sonnet', 'opus'). Defaults to system setting."),
				"profile":       stringProp("Optional happier backend profile (e.g. 'deepseek', 'openai', 'gemini'). Defaults to 'anthropic'."),
				"env_overrides": map[string]interface{}{
					"type":        "object",
					"description": "Optional environment variable overrides injected into the session. Omit for the default Anthropic API. To route through LiteLLM, set ANTHROPIC_BASE_URL+ANTHROPIC_API_KEY (or use the TUI [gpt]/[dseek] picker entries which do this for you). Pick a backend with ANTHROPIC_MODEL: 'chatgptsub-gpt-5.5' → [gpt], 'or-deepseek-v4-pro' → [dseek]. Session names get the matching auto-prefix. Keys and values must be strings.",
					"additionalProperties": map[string]interface{}{"type": "string"},
				},
				"mcp_servers": map[string]interface{}{
					"type":        "array",
					"description": "Optional subset of MCP server names (from ~/.swarmops/mcp-config.json) to load in the session. Omit/empty = full catalogue (~50 servers, ~300k tokens). Use a small focused list for [gpt]/[dseek] agents to keep the system prompt tiny.",
					"items":       map[string]interface{}{"type": "string"},
				},
			}, []string{"repo_path"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			repoPathRaw := getStringArg(args, "repo_path", "")
			if repoPathRaw == "" {
				return nil, fmt.Errorf("repo_path is required")
			}

			// Validate repo_path at the system boundary.
			repoPath, err := validateRepoPath(repoPathRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid repo_path: %w", err)
			}

			// Validate worktree_path if provided.
			worktreePathRaw := getStringArg(args, "worktree_path", "")
			var worktreePath string
			if worktreePathRaw != "" {
				worktreePath, err = validateWorktreePath(worktreePathRaw)
				if err != nil {
					return nil, fmt.Errorf("invalid worktree_path: %w", err)
				}
			}

			name := getStringArg(args, "name", "")
			branch := getStringArg(args, "branch", "")
			model := getStringArg(args, "model", "")
			taskBrief := getStringArg(args, "task_brief", "")

			missionStr := getStringArg(args, "mission", "")
			var mission *string
			if missionStr != "" {
				mission = &missionStr
			}
			envOverrides := getStringMapArg(args, "env_overrides")
			envOverrides = stashMCPServers(args, envOverrides)

			// Auto-prepend [gpt] or [dseek] prefix when routing to a non-Anthropic backend.
			// Pick the prefix from ANTHROPIC_MODEL first (the explicit signal), falling
			// back to the legacy model arg for older callers.
			prefixModel := envOverrides["ANTHROPIC_MODEL"]
			if prefixModel == "" {
				prefixModel = model
			}
			name = autoPrefixSessionName(name, prefixModel, envOverrides)

			profile := getStringArg(args, "profile", "")
			return svc.SpawnAgent(ctx, name, repoPath, worktreePath, branch, mission, model, profile, taskBrief, envOverrides)
		},
	)

	// ─── swop_teardown_agent ──────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name: "swop_teardown_agent",
			Description: "[WRITE] Tear down an agent session: remove its git worktree, optionally delete the branch, " +
				"and delete the session record. Safe to call on plain swop_run_task sessions too (skips worktree step). " +
				"WARNING: delete_branch=true on an unmerged branch will discard uncommitted work.",
			InputSchema: jsonSchema(map[string]interface{}{
				"id":            stringProp("Session ID to tear down."),
				"delete_branch": map[string]interface{}{"type": "boolean", "description": "Delete the git branch after removing the worktree (default false). Only applicable to agent sessions."},
			}, []string{"id"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			id := getStringArg(args, "id", "")
			if id == "" {
				return nil, fmt.Errorf("id is required")
			}
			deleteBranch := false
			if v, ok := args["delete_branch"]; ok {
				if b, ok := v.(bool); ok {
					deleteBranch = b
				}
			}
			if err := svc.TeardownAgent(ctx, id, deleteBranch); err != nil {
				return nil, err
			}
			return map[string]string{"status": "torn_down"}, nil
		},
	)

	// ─── swop_send_input ──────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_send_input",
			Description: "[WRITE] Send text input to a running session's tmux terminal.",
			InputSchema: jsonSchema(map[string]interface{}{
				"session_id": stringProp("Session ID to send input to"),
				"input":      stringProp("Text to send to the session"),
			}, []string{"session_id", "input"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			sessionID := getStringArg(args, "session_id", "")
			input := getStringArg(args, "input", "")
			if sessionID == "" {
				return nil, fmt.Errorf("session_id is required")
			}
			if input == "" {
				return nil, fmt.Errorf("input is required")
			}
			if err := svc.SendInput(ctx, sessionID, input); err != nil {
				return nil, err
			}
			return map[string]string{"status": "ok"}, nil
		},
	)

	// ─── swop_rename_session ──────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_rename_session",
			Description: "[WRITE] Rename a session.",
			InputSchema: jsonSchema(map[string]interface{}{
				"id":   stringProp("Session ID"),
				"name": stringProp("New session name"),
			}, []string{"id", "name"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			id := getStringArg(args, "id", "")
			name := getStringArg(args, "name", "")
			if id == "" {
				return nil, fmt.Errorf("id is required")
			}
			if name == "" {
				return nil, fmt.Errorf("name is required")
			}
			if err := svc.RenameSession(ctx, id, name); err != nil {
				return nil, err
			}
			return map[string]string{"status": "ok", "name": name}, nil
		},
	)

	// ─── swop_update_session ──────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_update_session",
			Description: "[WRITE] Update editable session fields: name, mission, directory, profile, context_id/context_name. Only supplied (non-empty) fields are changed. Profile change takes effect on next restart — does not auto-restart the session.",
			InputSchema: jsonSchema(map[string]interface{}{
				"id":           stringProp("Session ID (required)"),
				"name":         stringProp("New session name (leave empty to keep current)"),
				"mission":      stringProp("New mission statement; pass empty string to clear"),
				"directory":    stringProp("New working directory (validated to exist; takes effect on next restart)"),
				"profile":      stringProp("New happier backend profile (e.g. 'deepseek', 'openai', 'gemini', '' for default anthropic). Takes effect on next restart."),
				"context_id":   stringProp("New context ID (empty to clear)"),
				"context_name": stringProp("New context display name (empty to clear)"),
			}, []string{"id"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			id := getStringArg(args, "id", "")
			if id == "" {
				return nil, fmt.Errorf("id is required")
			}
			notes := []string{}
			if name, ok := args["name"]; ok && name != "" {
				if err := svc.RenameSession(ctx, id, name.(string)); err != nil {
					return nil, fmt.Errorf("rename: %w", err)
				}
				notes = append(notes, "name updated")
			}
			if _, hasMission := args["mission"]; hasMission {
				mission := getStringArg(args, "mission", "")
				if err := svc.UpdateSessionMission(ctx, id, mission); err != nil {
					return nil, fmt.Errorf("mission: %w", err)
				}
				notes = append(notes, "mission updated")
			}
			if dir, ok := args["directory"]; ok && dir != "" {
				dirStr := dir.(string)
				if _, err := os.Stat(dirStr); err != nil {
					return nil, fmt.Errorf("directory %q does not exist: %w", dirStr, err)
				}
				if err := svc.UpdateSessionDirectory(ctx, id, dirStr); err != nil {
					return nil, fmt.Errorf("directory: %w", err)
				}
				notes = append(notes, "directory updated (takes effect on next restart)")
			}
			if _, hasProfile := args["profile"]; hasProfile {
				profile := getStringArg(args, "profile", "")
				if err := svc.UpdateSessionProfile(ctx, id, profile); err != nil {
					return nil, fmt.Errorf("profile: %w", err)
				}
				notes = append(notes, "profile updated (takes effect on next restart)")
			}
			if _, hasCtx := args["context_id"]; hasCtx {
				ctxID := getStringArg(args, "context_id", "")
				ctxName := getStringArg(args, "context_name", "")
				if err := svc.UpdateSessionContext(ctx, id, ctxID, ctxName); err != nil {
					return nil, fmt.Errorf("context: %w", err)
				}
				notes = append(notes, "context updated")
			}
			if len(notes) == 0 {
				return map[string]string{"status": "ok", "note": "no fields supplied"}, nil
			}
			return map[string]interface{}{"status": "ok", "updated": notes}, nil
		},
	)

	// ─── swop_set_mission ─────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_set_mission",
			Description: "[WRITE] Set or update the mission statement for a session (1-3 sentences).",
			InputSchema: jsonSchema(map[string]interface{}{
				"id":      stringProp("Session ID"),
				"mission": stringProp("Mission statement (empty string to clear)"),
			}, []string{"id", "mission"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			id := getStringArg(args, "id", "")
			mission := getStringArg(args, "mission", "")
			if id == "" {
				return nil, fmt.Errorf("id is required")
			}
			if err := svc.UpdateSessionMission(ctx, id, mission); err != nil {
				return nil, err
			}
			return map[string]string{"status": "ok"}, nil
		},
	)

	// ─── swop_stop_session ────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_stop_session",
			Description: "[WRITE] Send interrupt (Ctrl+C) to a running session to stop it.",
			InputSchema: jsonSchema(map[string]interface{}{
				"id": stringProp("Session ID"),
			}, []string{"id"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			id := getStringArg(args, "id", "")
			if id == "" {
				return nil, fmt.Errorf("id is required")
			}
			if err := svc.StopSession(ctx, id); err != nil {
				return nil, err
			}
			return map[string]string{"status": "ok"}, nil
		},
	)

	// ─── swop_start_session ───────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_start_session",
			Description: "[WRITE] Start or resume a stopped session (re-launches claude in its tmux window).",
			InputSchema: jsonSchema(map[string]interface{}{
				"id": stringProp("Session ID"),
			}, []string{"id"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			id := getStringArg(args, "id", "")
			if id == "" {
				return nil, fmt.Errorf("id is required")
			}
			if err := svc.StartSession(ctx, id); err != nil {
				return nil, err
			}
			return map[string]string{"status": "ok"}, nil
		},
	)

	// ─── swop_accept_execution ────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_accept_execution",
			Description: "[WRITE] Accept/acknowledge a completed execution (no-op, returns current state).",
			InputSchema: jsonSchema(map[string]interface{}{
				"id": stringProp("Execution/session ID"),
			}, []string{"id"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			id := getStringArg(args, "id", "")
			if id == "" {
				return nil, fmt.Errorf("id is required")
			}
			return svc.GetExecution(ctx, id)
		},
	)

	// ─── swop_delete_execution ────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_delete_execution",
			Description: "[WRITE] Delete a session and kill its tmux window.",
			InputSchema: jsonSchema(map[string]interface{}{
				"id": stringProp("Execution/session ID to delete"),
			}, []string{"id"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			id := getStringArg(args, "id", "")
			if id == "" {
				return nil, fmt.Errorf("id is required")
			}
			if err := svc.DeleteExecution(ctx, id); err != nil {
				return nil, err
			}
			return map[string]string{"status": "deleted"}, nil
		},
	)

	// ─── swop_wait_for_execution ──────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_wait_for_execution",
			Description: "[READ] Check execution status. Returns current state immediately — poll for updates.",
			InputSchema: jsonSchema(map[string]interface{}{
				"id": stringProp("Execution/session ID"),
			}, []string{"id"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			id := getStringArg(args, "id", "")
			if id == "" {
				return nil, fmt.Errorf("id is required")
			}
			return svc.ExecutionProgress(ctx, id)
		},
	)

	// ─── swop_git_add ─────────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_git_add",
			Description: "[WRITE] Stage files for git commit.",
			InputSchema: jsonSchema(map[string]interface{}{
				"path":  stringProp("Repository path (default: current directory)"),
				"files": arrayOfStringsProp("Files to stage (e.g. [\".\"] for all)"),
			}, []string{"files"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			path := getStringArg(args, "path", ".")
			files := getStringSliceArg(args, "files")
			if len(files) == 0 {
				return nil, fmt.Errorf("files is required")
			}
			if err := svc.GitAdd(path, files); err != nil {
				return nil, err
			}
			return map[string]string{"status": "ok"}, nil
		},
	)

	// ─── swop_git_commit ──────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_git_commit",
			Description: "[WRITE] Create a git commit with the given message.",
			InputSchema: jsonSchema(map[string]interface{}{
				"path":    stringProp("Repository path (default: current directory)"),
				"message": stringProp("Commit message"),
			}, []string{"message"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			path := getStringArg(args, "path", ".")
			message := getStringArg(args, "message", "")
			if message == "" {
				return nil, fmt.Errorf("message is required")
			}
			out, err := svc.GitCommit(path, message)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", err, out)
			}
			return map[string]string{"output": out}, nil
		},
	)

	// ─── swop_git_push ────────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_git_push",
			Description: "[WRITE] Push committed changes to remote.",
			InputSchema: jsonSchema(map[string]interface{}{
				"path": stringProp("Repository path (default: current directory)"),
			}, nil),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			path := getStringArg(args, "path", ".")
			out, err := svc.GitPush(path)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", err, out)
			}
			return map[string]string{"output": out}, nil
		},
	)

	// ─── swop_repos_refresh ───────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "swop_repos_refresh",
			Description: "[WRITE] Re-scan the local filesystem (under SWARMOPS_REPO_ROOTS, default ~/git-bnx) and GitHub (orgs in SWARMOPS_GITHUB_ORGS, default ThomkerNet+boroughnexus-cto) and upsert every discovered repo into the SwarmOps known_repos registry. Run after cloning a new repo or before invoking swop_newgoal on a fresh deploy.",
			InputSchema: jsonSchema(map[string]interface{}{}, nil),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			localCount, remoteCount, err := reposRefreshAll(ctx)
			result := map[string]interface{}{
				"local_upserts":  localCount,
				"remote_upserts": remoteCount,
			}
			if err != nil {
				result["error"] = err.Error()
			}
			return result, nil
		},
	)

	// ─── swop_newgoal ─────────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name: "swop_newgoal",
			Description: "[WRITE] Pick the right repo for a goal via the warm pool brain, then spawn a Claude Code session in that repo's working directory with the goal injected as the first prompt. Replaces 'manually choose directory + run swop_run_task' — eliminates context bloat from agents starting in the wrong tree. " +
				"If no repo matches confidently, returns the brain's suggestion without spawning so the caller can decide whether to clone something new.",
			InputSchema: jsonSchema(map[string]interface{}{
				"goal":        stringProp("The work to accomplish (1-3 sentences). The brain uses this to pick a repo and the spawned session sees it as the first user message."),
				"brain_model": stringProp("Optional model the dispatcher uses for the routing decision. Defaults to 'haiku' (fast, cheap, sufficient for ~30-repo catalogues). Pass 'sonnet' for trickier picks, or any LiteLLM-routed id (e.g. 'chatgptsub-gpt-5.5', 'or-deepseek-v4-pro') for a third-party brain — note: third-party brains require a [gpt]/[dseek] worker in the pool; default brains use the pool's claude models."),
				"backend":     stringProp("Optional inference backend for the spawned session. 'default' (or omit) = Anthropic; 'gpt' = LiteLLM-routed chatgptsub-gpt-5.5 (session auto-prefixed [gpt]); 'dseek' = LiteLLM-routed or-deepseek-v4-pro (session auto-prefixed [dseek]). The brain itself still uses brain_model — this only affects the spawned worker."),
				"mcp_servers": map[string]interface{}{
					"type":        "array",
					"description": "Optional subset of MCP server names to load in the spawned session. Recommended for backend='gpt'/'dseek' so the session boots with a tiny tool surface (~5k tokens) instead of the full ~50-server catalogue (~300k tokens). Example: ['tkn-plane','tkn-komodo','tkn-firecrawl'].",
					"items":       map[string]interface{}{"type": "string"},
				},
				"dry_run":     map[string]interface{}{"type": "boolean", "description": "If true, return the brain's decision without spawning a session. Default false."},
			}, []string{"goal"}),
		},
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			goal := strings.TrimSpace(getStringArg(args, "goal", ""))
			if goal == "" {
				return nil, fmt.Errorf("goal is required")
			}
			brainModel := getStringArg(args, "brain_model", "")
			if brainModel == "" {
				brainModel = brainDefaultModel
			}
			dryRun := getBoolArg(args, "dry_run", false)

			// Resolve the spawned-session backend. Empty / "default" → no
			// env overrides; "gpt" / "dseek" → LiteLLM-routed env vars from
			// backend.go, plus the matching model id so autoPrefixSessionName
			// adds the [gpt] / [dseek] prefix.
			backend := strings.ToLower(strings.TrimSpace(getStringArg(args, "backend", "")))
			var sessionEnv map[string]string
			var sessionModel string
			switch backend {
			case "", "default", "anthropic":
				// nothing — plain Anthropic
			case "gpt":
				sessionModel = litellmModelGPT55
				sessionEnv = litellmEnvOverrides(sessionModel)
			case "dseek", "deepseek":
				sessionModel = litellmModelDeepseek4
				sessionEnv = litellmEnvOverrides(sessionModel)
			default:
				return nil, fmt.Errorf("unknown backend %q (expected: default | gpt | dseek)", backend)
			}

			repos, err := listKnownRepos(ctx)
			if err != nil {
				return nil, fmt.Errorf("load repo registry: %w", err)
			}
			if len(repos) == 0 {
				return map[string]interface{}{
					"status":  "no-repos",
					"hint":    "registry is empty — run swop_repos_refresh first",
					"goal":    goal,
					"spawned": nil,
				}, nil
			}

			raw, err := brainAsk(ctx, svc, brainModel, brainSystemPrompt, brainUserPrompt(goal, repos))
			if err != nil {
				return nil, fmt.Errorf("brain dispatch: %w", err)
			}
			pick, err := brainParse(raw)
			if err != nil {
				return nil, fmt.Errorf("parse brain reply: %w", err)
			}

			// Resolve the pick against the catalogue.
			var matched *KnownRepo
			if pick.Pick != "" && strings.ToLower(pick.Pick) != "none" {
				for i := range repos {
					if repos[i].Slug() == pick.Pick {
						matched = &repos[i]
						break
					}
				}
			}

			result := map[string]interface{}{
				"goal":        goal,
				"brain_model": brainModel,
				"pick":        pick.Pick,
				"confidence":  pick.Confidence,
				"reasoning":   pick.Reasoning,
				"suggestions": pick.Suggestions,
			}

			switch {
			case matched == nil:
				// No-match or unknown slug — return suggestion, do not spawn.
				result["status"] = "no-match"
				result["hint"] = "no repo confidently matches — consider creating a new one or refining the goal"
				return result, nil

			case !matched.IsCloned():
				// Picked repo isn't cloned locally — surface the clone command.
				result["status"] = "uncloned-pick"
				result["matched"] = matched
				result["clone_command"] = fmt.Sprintf(
					"git clone https://github.com/%s/%s ~/git-bnx/%s",
					matched.Owner, matched.Name, matched.Name,
				)
				result["hint"] = "picked repo isn't cloned — clone it then call swop_newgoal again"
				return result, nil

			case dryRun:
				result["status"] = "dry-run"
				result["matched"] = matched
				return result, nil

			default:
				// Spawn a session in the picked repo, with the goal as the first message.
				name := sanitizeSessionName(goal)
				if name == "" {
					name = "newgoal"
				}
				// Auto-prefix [gpt] / [dseek] when routing through LiteLLM,
				// matching the TUI picker + swop_run_task behaviour. No-op for
				// the default backend (sessionEnv is nil).
				name = autoPrefixSessionName(name, sessionModel, sessionEnv)
				// Forward an optional mcp_servers subset so the spawned
				// session loads only those tools — crucial for [gpt]/[dseek]
				// to keep the system prompt under control.
				sessionEnv = stashMCPServers(args, sessionEnv)
				missionStr := goal
				sess, err := svc.RunTask(ctx, name, matched.LocalPath, &missionStr, sessionModel, "", goal, sessionEnv)
				if err != nil {
					return nil, fmt.Errorf("spawn session: %w", err)
				}
				result["backend"] = backend
				// Inject `/goal <goal>` asynchronously after Claude has had a
				// chance to start up and clear the workspace-trust prompt.
				// The slash command sets a session-scoped Stop hook so claude
				// can't end the conversation until the goal is met. TASK.md
				// (written by RunTask) still gives the agent the brief.
				go injectGoalAfterReady(sess.TmuxSession, goal)
				result["status"] = "spawned"
				result["matched"] = matched
				result["session"] = sess
				return result, nil
			}
		},
	)

	// ─── Pool tools (gated behind SWARMOPS_MCP_POOL_TOOLS) ──────────────

	if enablePoolTools {
		// ─── pool_status ────────────────────────────────────────────
		reg.Register(
			ToolDefinition{
				Name:        "pool_status",
				Description: "[READ] Get warm Claude CLI pool status (slots, models, costs, availability).",
				InputSchema: jsonSchema(map[string]interface{}{}, nil),
			},
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return svc.PoolStatus(), nil
			},
		)

		// ─── pool_chat ──────────────────────────────────────────────
		reg.Register(
			ToolDefinition{
				Name:        "pool_chat",
				Description: "[WRITE] Send a chat message through the warm Claude CLI pool. Returns the full response.",
				InputSchema: jsonSchema(map[string]interface{}{
					"model": stringProp("Model name or alias (e.g. 'haiku', 'sonnet', 'opus', 'claude-sonnet-4-6')"),
					"messages": map[string]interface{}{
						"type":        "array",
						"description": "OpenAI-format messages array [{role, content}, ...]",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"role":    map[string]interface{}{"type": "string"},
								"content": map[string]interface{}{"type": "string"},
							},
							"required": []string{"role", "content"},
						},
					},
				}, []string{"model", "messages"}),
			},
			poolChatHandler(svc),
		)
	}
}

// poolChatHandler creates the handler for the pool_chat tool.
func poolChatHandler(svc *Services) ToolHandler {
	return func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		if svc.pool == nil {
			return nil, fmt.Errorf("pool is not enabled")
		}

		modelName := getStringArg(args, "model", "")
		if modelName == "" {
			return nil, fmt.Errorf("model is required")
		}

		model, ok := resolveModel(modelName)
		if !ok {
			return nil, fmt.Errorf("unknown model: %s", modelName)
		}

		// Parse messages from args
		messagesRaw, ok := args["messages"]
		if !ok {
			return nil, fmt.Errorf("messages is required")
		}

		// Re-marshal and unmarshal to get typed messages
		messagesJSON, err := json.Marshal(messagesRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid messages: %w", err)
		}
		var messages []oaiMessage
		if err := json.Unmarshal(messagesJSON, &messages); err != nil {
			return nil, fmt.Errorf("invalid messages format: %w", err)
		}

		if len(messages) == 0 {
			return nil, fmt.Errorf("messages array is empty")
		}

		// Build prompt from messages
		prompt := messagesToPrompt(messages)

		// Acquire pool slot
		slot, err := svc.pool.Acquire(ctx, model)
		if err != nil {
			return nil, fmt.Errorf("pool exhausted for model %s: %w", model, err)
		}
		defer svc.pool.Release(slot)

		// Send query
		if err := slot.sendQuery(prompt); err != nil {
			slot.mu.Lock()
			slot.state = slotDead
			slot.mu.Unlock()
			return nil, fmt.Errorf("failed to send query: %w", err)
		}

		// Collect full response
		var fullText strings.Builder
		var resultEv poolEvent

		for {
			ev, err := readEventWithCtx(ctx, slot)
			if err != nil {
				slot.mu.Lock()
				slot.state = slotDead
				slot.mu.Unlock()
				return nil, fmt.Errorf("failed reading response: %w", err)
			}

			switch ev.Type {
			case "assistant":
				fullText.WriteString(extractAssistantText(ev))
			case "stream_event":
				fullText.WriteString(extractStreamText(ev))
			case "rate_limit_event":
				svc.pool.handleRateLimit(slot, ev)
			case "result":
				resultEv = ev
				goto done
			}
		}

	done:
		slot.mu.Lock()
		slot.errorCount = 0
		slot.totalCost += resultEv.CostUSD
		slot.totalRequests++
		slot.mu.Unlock()
		svc.pool.totalCost.Add(int64(resultEv.CostUSD * 1e6))

		if resultEv.IsError {
			action := classifyResultError(resultEv)
			if action == "disable" || action == "recycle" {
				slot.mu.Lock()
				slot.state = slotDead
				slot.mu.Unlock()
			}
			return nil, fmt.Errorf("model error: %s", resultEv.Result)
		}

		text := fullText.String()
		if text == "" {
			text = resultEv.Result
		}

		tokensIn, tokensOut := 0, 0
		if resultEv.Usage != nil {
			tokensIn = resultEv.Usage.InputTokens
			tokensOut = resultEv.Usage.OutputTokens
		}

		return map[string]interface{}{
			"response":         text,
			"model":            model,
			"tokens_in":        tokensIn,
			"tokens_out":       tokensOut,
			"cost_usd":         resultEv.CostUSD,
			"duration_ms":      resultEv.DurationMS,
			"duration_api_ms":  resultEv.DurationAPIMS,
		}, nil
	}
}

