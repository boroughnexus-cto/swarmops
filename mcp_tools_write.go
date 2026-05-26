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
	// ─── rc_run_task ────────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_run_task",
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
			ctxIDStr := getStringArg(args, "context_id", "")
			var contextID *string
			if ctxIDStr != "" {
				contextID = &ctxIDStr
			}
			ctxNameStr := getStringArg(args, "context_name", "")
			var contextName *string
			if ctxNameStr != "" {
				contextName = &ctxNameStr
			}
			envOverrides := getStringMapArg(args, "env_overrides")

			// Auto-prepend [gpt] or [dseek] prefix when routing to a non-Anthropic backend.
			// Pick the prefix from ANTHROPIC_MODEL first (the explicit signal), falling
			// back to the legacy model arg for older callers.
			prefixModel := envOverrides["ANTHROPIC_MODEL"]
			if prefixModel == "" {
				prefixModel = model
			}
			name = autoPrefixSessionName(name, prefixModel, envOverrides)

			profile := getStringArg(args, "profile", "")
			return svc.RunTask(ctx, name, directory, contextID, contextName, mission, model, profile, taskBrief, envOverrides)
		},
	)

	// ─── rc_spawn_agent ─────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name: "rc_spawn_agent",
			Description: "[WRITE] Atomically create a git worktree and spawn a Claude Code agent inside it. " +
				"Writes an optional TASK.md brief before the agent starts. " +
				"Rolls back the worktree on any failure. " +
				"Use rc_teardown_agent to clean up when done.",
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
			ctxIDStr := getStringArg(args, "context_id", "")
			var contextID *string
			if ctxIDStr != "" {
				contextID = &ctxIDStr
			}
			ctxNameStr := getStringArg(args, "context_name", "")
			var contextName *string
			if ctxNameStr != "" {
				contextName = &ctxNameStr
			}
			envOverrides := getStringMapArg(args, "env_overrides")

			// Auto-prepend [gpt] or [dseek] prefix when routing to a non-Anthropic backend.
			// Pick the prefix from ANTHROPIC_MODEL first (the explicit signal), falling
			// back to the legacy model arg for older callers.
			prefixModel := envOverrides["ANTHROPIC_MODEL"]
			if prefixModel == "" {
				prefixModel = model
			}
			name = autoPrefixSessionName(name, prefixModel, envOverrides)

			profile := getStringArg(args, "profile", "")
			return svc.SpawnAgent(ctx, name, repoPath, worktreePath, branch, contextID, contextName, mission, model, profile, taskBrief, envOverrides)
		},
	)

	// ─── rc_teardown_agent ──────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name: "rc_teardown_agent",
			Description: "[WRITE] Tear down an agent session: remove its git worktree, optionally delete the branch, " +
				"and delete the session record. Safe to call on plain rc_run_task sessions too (skips worktree step). " +
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

	// ─── rc_send_input ──────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_send_input",
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

	// ─── rc_rename_session ──────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_rename_session",
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

	// ─── rc_update_session ──────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_update_session",
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

	// ─── rc_set_mission ─────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_set_mission",
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

	// ─── rc_stop_session ────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_stop_session",
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

	// ─── rc_start_session ───────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_start_session",
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

	// ─── rc_accept_execution ────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_accept_execution",
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

	// ─── rc_delete_execution ────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_delete_execution",
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

	// ─── rc_wait_for_execution ──────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_wait_for_execution",
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

	// ─── rc_git_add ─────────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_git_add",
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

	// ─── rc_git_commit ──────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_git_commit",
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

	// ─── rc_git_push ────────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_git_push",
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

	// ─── rc_repos_refresh ───────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name:        "rc_repos_refresh",
			Description: "[WRITE] Re-scan the local filesystem (under SWARMOPS_REPO_ROOTS, default ~/git-bnx) and GitHub (orgs in SWARMOPS_GITHUB_ORGS, default ThomkerNet+boroughnexus-cto) and upsert every discovered repo into the SwarmOps known_repos registry. Run after cloning a new repo or before invoking rc_newgoal on a fresh deploy.",
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

	// ─── rc_newgoal ─────────────────────────────────────────────────────
	reg.Register(
		ToolDefinition{
			Name: "rc_newgoal",
			Description: "[WRITE] Pick the right repo for a goal via the warm pool brain, then spawn a Claude Code session in that repo's working directory with the goal injected as the first prompt. Replaces 'manually choose directory + run rc_run_task' — eliminates context bloat from agents starting in the wrong tree. " +
				"If no repo matches confidently, returns the brain's suggestion without spawning so the caller can decide whether to clone something new.",
			InputSchema: jsonSchema(map[string]interface{}{
				"goal":        stringProp("The work to accomplish (1-3 sentences). The brain uses this to pick a repo and the spawned session sees it as the first user message."),
				"brain_model": stringProp("Optional model the dispatcher uses for the routing decision. Defaults to 'haiku' (fast, cheap, sufficient for ~30-repo catalogues). Pass 'sonnet' for trickier picks, or any LiteLLM-routed id (e.g. 'chatgptsub-gpt-5.5', 'or-deepseek-v4-pro') for a third-party brain — note: third-party brains require a [gpt]/[dseek] worker in the pool; default brains use the pool's claude models."),
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

			repos, err := listKnownRepos(ctx)
			if err != nil {
				return nil, fmt.Errorf("load repo registry: %w", err)
			}
			if len(repos) == 0 {
				return map[string]interface{}{
					"status":  "no-repos",
					"hint":    "registry is empty — run rc_repos_refresh first",
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
				result["hint"] = "picked repo isn't cloned — clone it then call rc_newgoal again"
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
				missionStr := goal
				sess, err := svc.RunTask(ctx, name, matched.LocalPath, nil, nil, &missionStr, "", "", goal, nil)
				if err != nil {
					return nil, fmt.Errorf("spawn session: %w", err)
				}
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

