# SwarmOps v3 — Unified MCP Interface & Smart Pool Routing

## Overview

Rename all MCP tools from `rc_*` to `swo_*`, expose the warm pool via MCP, add model selection to session dispatch, and implement smart routing. This transforms SwarmOps from a legacy remote-code wrapper into a first-class AI task orchestration platform.

## Scope

Two codebases change:
1. **MCP Server** — `TKNet-MCPServer/servers/tkn-remote-code-auto/server.py` (rename to `tkn-swarmops`)
2. **SwarmOps Go binary** — `swarmops-rework/` (minor additions to REST API)

---

## Phase 1: Rename & Restructure MCP Server

### 1.1 Create new server directory (don't rename — tkn-remote-code-auto still serves its own instance)

The MCP stack has TWO deployments of the same code:
- `tkn-remote-code-auto` (port 8238) — controls the separate remote-code-auto instance
- `tkn-swarmops` (port 8243) — controls the NUC SwarmOps instance

They currently share `RemoteCodeAutoServer` via a `SERVER_DIRS` override in `main.py` (line 148).

Create a NEW server directory:
```
servers/tkn-swarmops/server.py   (new SwarmOpsServer class)
servers/tkn-swarmops/__init__.py (empty)
```

Copy `servers/tkn-remote-code-auto/server.py` as the starting point, then:
- Rename class: `RemoteCodeAutoServer` → `SwarmOpsServer`
- Rename all `rc_*` methods to `swo_*`
- Add new pool/session tools (Phase 2-5)
- Update `__init__` defaults: `server_name = "tkn-swarmops"`, `port = 8243`

Update `main.py`:
- Line 118: change `("RemoteCodeAutoServer", 8243, "TKN_RC_NUC")` → `("SwarmOpsServer", 8243, "TKN_SWO")`
- Line 148: REMOVE the `SERVER_DIRS` entry for `tkn-swarmops` (it now has its own directory)

Leave `tkn-remote-code-auto` completely untouched.

### 1.2 Rename all tools rc_* → swo_*

Every tool method gets renamed. The function names ARE the MCP tool names (the decorator reads them).

| Old name | New name | Category |
|----------|----------|----------|
| `rc_health_check` | `swo_health_check` | read |
| `rc_health` | `swo_dashboard_stats` | read |
| `rc_tmux_sessions` | `swo_tmux_sessions` | read |
| `rc_list_agents` | `swo_list_agents` | read |
| `rc_list_roots` | `swo_list_roots` | read |
| `rc_list_projects` | `swo_list_projects` | read |
| `rc_list_tasks` | `swo_list_tasks` | read |
| `rc_list_executions` | `swo_list_executions` | read |
| `rc_get_execution` | `swo_get_execution` | read |
| `rc_wait_for_execution` | `swo_wait_for_execution` | read |
| `rc_execution_progress` | `swo_execution_progress` | read |
| `rc_dashboard` | `swo_agent_dashboard` | read |
| `rc_git_status` | `swo_git_status` | read |
| `rc_git_diff` | `swo_git_diff` | read |
| `rc_git_branches` | `swo_git_branches` | read |
| `rc_run_task` | `swo_run_task` | write |
| `rc_send_input` | `swo_send_input` | write |
| `rc_create_project` | `swo_create_project` | write |
| `rc_create_task` | `swo_create_task` | write |
| `rc_git_add` | `swo_git_add` | write |
| `rc_git_commit` | `swo_git_commit` | write |
| `rc_accept_execution` | `swo_accept_execution` | dangerous |
| `rc_delete_execution` | `swo_delete_execution` | dangerous |
| `rc_git_push` | `swo_git_push` | dangerous |

Also update all docstrings and description strings that reference `rc_*` tool names.

### 1.3 Update Docker stack

In `TKNet-Docker-Stacks/mcp-swarmops/docker-compose.yml`, the `STANDALONE_SERVER` env var must change:

```yaml
- STANDALONE_SERVER=tkn-swarmops
```

This is already correct (the server entry point in main.py uses the server_name to find the right class). Verify the registration in `TKNet-MCPServer/main.py` maps `tkn-swarmops` to the new class.

### 1.4 Backward compatibility

Add temporary aliases so existing Hermes SOUL.md references to `rc_*` don't break during transition:

```python
# Temporary aliases — remove after Hermes SOUL.md is updated
rc_health_check = swo_health_check
rc_dashboard = swo_agent_dashboard
rc_run_task = swo_run_task
# ... etc
```

Remove these after SOUL.md is updated to use `swo_*` names.

---

## Phase 2: New Pool Tools (MCP wrappers for existing SwarmOps API)

The SwarmOps Go binary already exposes these endpoints. The MCP server just needs to wrap them.

### 2.1 swo_pool_status (read tool)

Wraps `GET /api/swarm/pool` on the SwarmOps Go server.

```python
@read_tool(description="Get warm pool health: available slots per model, cost, request counts.")
async def swo_pool_status(self) -> dict:
    """Get the current state of the warm Claude CLI session pool.

    Returns slot availability per model (haiku, sonnet, opus), total cost,
    request counts, and whether each slot is idle/busy/starting/dead.

    Use this before swo_pool_chat to check which models have available slots.

    Returns:
        {"enabled": bool, "models": {model: {slots: [...], available: int, ...}}, "total_cost_usd": float}
    """
    return await self._get("/api/swarm/pool")
```

HITL: auto-approved (read-only).

### 2.2 swo_pool_chat (write tool)

Wraps `POST /v1/chat/completions` on the SwarmOps Go server.

```python
@write_tool(
    description=(
        "Send a one-shot question to the warm Claude pool. "
        "Specify model: 'haiku' (fast/cheap), 'sonnet' (balanced), 'opus' (strongest). "
        "No session state — use swo_run_task for multi-step work."
    )
)
async def swo_pool_chat(
    self,
    message: str,
    model: str = "sonnet",
    system_prompt: str = "",
) -> dict:
    """Send a one-shot message to a warm Claude CLI session in the pool.

    The pool maintains warm Claude Code processes (with full MCP access).
    This is faster than swo_run_task for simple questions because it
    skips tmux session creation.

    Model selection guide:
        haiku  — fast, cheap. Lookups, summaries, simple questions.
        sonnet — balanced. Code review, moderate analysis, tool use.
        opus   — strongest. Complex multi-step reasoning, architecture.

    Args:
        message:       The question or task for the pool worker.
        model:         Model to use: "haiku", "sonnet", or "opus" (default: "sonnet").
        system_prompt: Optional system prompt prepended to the message.

    Returns:
        {"response": str, "model": str, "usage": {input_tokens, output_tokens},
         "cost_usd": float, "latency_ms": int}
    """
    messages = []
    if system_prompt:
        messages.append({"role": "system", "content": system_prompt})
    messages.append({"role": "user", "content": message})

    # Resolve short model names
    model_map = {"haiku": "claude-haiku-4-5", "sonnet": "claude-sonnet-4-6", "opus": "claude-opus-4-6"}
    resolved_model = model_map.get(model.lower(), model)

    body = {
        "model": resolved_model,
        "messages": messages,
        "stream": False,
    }

    # Use the pool API key if configured
    headers = {}
    pool_api_key = os.environ.get("POOL_API_KEY", "")
    if pool_api_key:
        headers["Authorization"] = f"Bearer {pool_api_key}"

    start = time.monotonic()
    self._ensure_client()
    response = await self._http.post("/v1/chat/completions", json=body, headers=headers)
    response.raise_for_status()
    elapsed_ms = int((time.monotonic() - start) * 1000)

    result = response.json()
    choice = result.get("choices", [{}])[0]
    usage = result.get("usage", {})

    return {
        "response": choice.get("message", {}).get("content", ""),
        "model": result.get("model", resolved_model),
        "usage": {
            "input_tokens": usage.get("prompt_tokens", 0),
            "output_tokens": usage.get("completion_tokens", 0),
        },
        "cost_usd": 0.0,  # Pool doesn't return cost in OAI format; enhancement for later
        "latency_ms": elapsed_ms,
    }
```

HITL: write (Telegram approval) because it runs Claude Code with tool access.

### 2.3 swo_pool_models (read tool)

Wraps `GET /v1/models`.

```python
@read_tool(description="List models available in the warm pool with context lengths.")
async def swo_pool_models(self) -> dict:
    """List models available in the SwarmOps warm pool.

    Returns:
        {"models": [{"id": str, "context_length": int}]}
    """
    return await self._get("/v1/models")
```

---

## Phase 3: New Session Tools (wrap SwarmOps v2 session API)

The Go binary has a newer session API at `/api/swarm/sessions` that's simpler than the legacy remote-code API. Expose these for lightweight session management.

### 3.1 swo_list_sessions (read tool)

```python
@read_tool(description="List all managed Claude Code sessions with status and working directory.")
async def swo_list_sessions(self) -> list:
    """List sessions managed by SwarmOps (not legacy remote-code).

    Returns:
        List of session objects with id, name, directory, status (running/stopped), tmux_session.
    """
    return await self._get("/api/swarm/sessions")
```

### 3.2 swo_create_session (write tool)

```python
@write_tool(
    description=(
        "Create a new Claude Code session in a named tmux session. "
        "Specify a working directory for the session. "
        "Use for long-running, interactive tasks that need multiple inputs."
    )
)
async def swo_create_session(
    self,
    name: str,
    directory: str = ".",
    context_id: str | None = None,
) -> dict:
    """Create a new managed Claude Code session.

    Spawns a tmux session, cd's into the directory, and launches `claude`.
    Use swo_send_session_input to interact with it.

    Args:
        name:       Session name (must be unique).
        directory:  Working directory for the session.
        context_id: Optional mcp-context ID to preload.

    Returns:
        Session object with id, name, tmux_session, directory, status.
    """
    body = {"name": name, "directory": directory}
    if context_id:
        body["context_id"] = context_id
    return await self._post("/api/swarm/sessions", body)
```

### 3.3 swo_session_terminal (read tool)

```python
@read_tool(description="Capture the terminal output of a managed session (last 300 lines).")
async def swo_session_terminal(self, session_id: str) -> dict:
    """Get the terminal output of a managed session.

    Args:
        session_id: Session ID (from swo_list_sessions).

    Returns:
        {"content": str} — the captured terminal text.
    """
    return await self._get(f"/api/swarm/sessions/{session_id}/terminal")
```

### 3.4 swo_send_session_input (write tool)

```python
@write_tool(description="Send text input to a managed session's tmux terminal.")
async def swo_send_session_input(self, session_id: str, input_text: str) -> dict:
    """Send input to a managed session.

    Args:
        session_id: Session ID (from swo_list_sessions).
        input_text: Text to type into the session.

    Returns:
        {"status": "ok"}
    """
    return await self._post(
        f"/api/swarm/sessions/{session_id}/input",
        {"input": input_text},
    )
```

### 3.5 swo_delete_session (dangerous tool)

```python
@dangerous_tool(
    warning="Kills the tmux session immediately. Any in-progress work is lost.",
    description="Delete a managed session and kill its tmux session.",
)
async def swo_delete_session(self, session_id: str) -> dict:
    """Delete a managed session.

    Args:
        session_id: Session ID (from swo_list_sessions).

    Returns:
        {"ok": True}
    """
    return await self._delete(f"/api/swarm/sessions/{session_id}")
```

### 3.6 swo_swarm_dashboard (read tool)

Wraps `/api/swarm/dashboard` — the combined view that includes both sessions and pool status.

```python
@read_tool(
    description=(
        "Combined SwarmOps dashboard: all sessions (running/stopped), "
        "pool status (slots per model), and aggregate stats."
    )
)
async def swo_swarm_dashboard(self) -> dict:
    """Get the combined SwarmOps dashboard.

    Returns sessions, running count, total count, and pool status in one call.
    Use this as the primary overview tool.

    Returns:
        {"sessions": [...], "running": int, "total": int, "pool": {...}}
    """
    return await self._get("/api/swarm/dashboard")
```

---

## Phase 4: Model Selection for Session Dispatch

### 4.1 Go change: spawn.go — accept model parameter

Current `spawnSession` launches `claude` with no model flag. Add a `model` parameter.

File: `swarmops-rework/spawn.go`

```go
func spawnSession(ctx context.Context, name, directory string, contextID, contextName *string, model string) (*Session, error) {
    s, err := createSession(ctx, name, directory, contextID, contextName, false)
    if err != nil {
        return nil, fmt.Errorf("create session: %w", err)
    }

    cmd := exec.Command("tmux", "new-session", "-d", "-s", s.TmuxSession, "-c", directory)
    if out, err := cmd.CombinedOutput(); err != nil {
        deleteSession(ctx, s.ID)
        return nil, fmt.Errorf("tmux new-session: %v: %s", err, out)
    }

    // Build claude command with optional model
    claudeCmd := "claude"
    if model != "" {
        claudeCmd = fmt.Sprintf("claude --model %s", model)
    }
    if err := exec.Command("tmux", "send-keys", "-t", s.TmuxSession, claudeCmd, "Enter").Run(); err != nil {
        log.Printf("spawn: failed to start claude in %s: %v", s.TmuxSession, err)
    }

    log.Printf("spawn: created session %q (tmux=%s, dir=%s, model=%s)", name, s.TmuxSession, directory, model)
    return s, nil
}
```

### 4.2 Go change: api_slim.go — accept model in session create

In `handleSwarmSessionsAPI`, the POST body should accept a `model` field:

```go
var req struct {
    Name      string  `json:"name"`
    Directory string  `json:"directory"`
    ContextID *string `json:"context_id"`
    Model     string  `json:"model"`  // NEW: "haiku", "sonnet", "opus", or full name
}
```

Pass `req.Model` to `spawnSession`.

### 4.3 MCP change: swo_create_session gets model param

Add `model: str = ""` parameter to `swo_create_session`:

```python
async def swo_create_session(
    self,
    name: str,
    directory: str = ".",
    model: str = "",
    context_id: str | None = None,
) -> dict:
```

Include in POST body:
```python
if model:
    body["model"] = model
```

### 4.4 Update callers in api_slim.go

The `handleTaskExecutionsAPI` POST handler (line 140-168) also calls `spawnSession` — update it to pass model from the request body too.

---

## Phase 5: SwarmOps Config API (expose via MCP)

### 5.1 swo_get_config (read tool)

```python
@read_tool(description="Get all SwarmOps configuration values with source (default/env/db/runtime).")
async def swo_get_config(self) -> list:
    """Get all SwarmOps configuration entries.

    Returns:
        List of config entries with key, value, source, description.
    """
    return await self._get("/api/swarm/config")
```

### 5.2 swo_set_config (dangerous tool)

```python
@dangerous_tool(
    warning="Changes SwarmOps configuration. Some changes require a restart to take effect.",
    description="Set a SwarmOps configuration value (pool size, models, timeouts, etc.).",
)
async def swo_set_config(self, key: str, value: str) -> dict:
    """Set a SwarmOps configuration value.

    Known keys:
        pool.enabled         — Enable/disable warm pool (true/false)
        pool.slots_per_model — Warm instances per model (1-10)
        pool.models          — Comma-separated model list
        pool.request_timeout_s — Per-request timeout in seconds
        pool.idle_recycle_min  — Recycle idle slots after N minutes

    Args:
        key:   Config key (e.g., "pool.slots_per_model").
        value: New value.

    Returns:
        Updated config entry.
    """
    return await self._post(f"/api/swarm/config/{key}", {"value": value, "changed_by": "mcp"})
```

Note: the Go handler uses PUT, not POST. The MCP tool should use `_put` (add a `_put` helper method to the server class).

---

## Phase 6: HITL Policy Update

Update the docstring at the top of server.py with the new tool names and categories:

```
HITL POLICY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Read (auto-approved):
    swo_health_check       — service health
    swo_dashboard_stats    — dashboard statistics
    swo_tmux_sessions      — tmux session list
    swo_list_agents        — agent discovery
    swo_list_roots         — project roots
    swo_list_projects      — project list
    swo_list_tasks         — task list
    swo_list_executions    — execution list
    swo_get_execution      — execution details
    swo_wait_for_execution — poll with milestones
    swo_execution_progress — quick status snapshot
    swo_agent_dashboard    — combined agent view
    swo_git_status         — git status
    swo_git_diff           — git diff
    swo_git_branches       — git branches
    swo_pool_status        — pool slot availability       [NEW]
    swo_pool_models        — available models              [NEW]
    swo_list_sessions      — managed session list          [NEW]
    swo_session_terminal   — session terminal capture      [NEW]
    swo_swarm_dashboard    — combined dashboard            [NEW]
    swo_get_config         — configuration values          [NEW]

Write (requires Telegram approval):
    swo_run_task           — start task execution
    swo_send_input         — send input to execution
    swo_create_project     — create project
    swo_create_task        — create task
    swo_git_add            — stage files
    swo_git_commit         — commit changes
    swo_pool_chat          — send message to pool          [NEW]
    swo_create_session     — create managed session        [NEW]
    swo_send_session_input — send input to session         [NEW]

Dangerous (requires approval + warning):
    swo_accept_execution   — mark execution completed
    swo_delete_execution   — kill execution
    swo_git_push           — push to remote
    swo_delete_session     — kill managed session          [NEW]
    swo_set_config         — change configuration          [NEW]
```

---

## Phase 7: Update Hermes SOUL.md

After all MCP tools are deployed, update the SwarmOps section in SOUL.md:

```yaml
### Key Tools

| Tool | When to use |
|------|-------------|
| `swo_swarm_dashboard` | Overview: sessions + pool status in one call |
| `swo_pool_chat` | Quick one-shot question (specify model: haiku/sonnet/opus) |
| `swo_pool_status` | Check available pool slots before dispatching |
| `swo_create_session` | Long-running task needing multiple interactions |
| `swo_run_task` | Dispatch to legacy project/task system |
| `swo_send_input` | Send follow-up to running execution |
| `swo_send_session_input` | Send follow-up to managed session |
| `swo_wait_for_execution` | Poll execution with milestones |
| `swo_session_terminal` | Read session terminal output |
| `swo_set_config` | Change pool size, models, timeouts |
```

### Routing guidance for SOUL.md

```
### Model Routing

Choose the cheapest model that can handle the task:

| Task type | Model | Why |
|-----------|-------|-----|
| Status checks, simple lookups | haiku | Fast, cheap, sufficient |
| Code review, moderate analysis | sonnet | Good balance |
| Complex multi-step, architecture | opus | Full reasoning power |
| File edits, git ops, deploys | opus (session) | Needs tool access + judgment |
```

---

## Implementation Order

1. Phase 1 (rename) — must go first, everything builds on it
2. Phase 2 (pool tools) — biggest value, unlocks model routing
3. Phase 4 (model in sessions) — Go change, small
4. Phase 3 (session tools) — MCP wrappers, straightforward
5. Phase 5 (config API) — nice to have
6. Phase 6 (HITL update) — documentation
7. Phase 7 (SOUL.md) — last, after deployment

## Files Changed

### TKNet-MCPServer repo
- `servers/tkn-remote-code-auto/server.py` → rename to `servers/tkn-swarmops/server.py`
- `main.py` — update server registration if needed
- Tests if any exist for tkn-remote-code-auto

### swarmops-rework repo
- `spawn.go` — add model parameter
- `api_slim.go` — accept model in session/execution create POST bodies

### TKNet-Docker-Stacks repo
- `mcp-swarmops/docker-compose.yml` — verify STANDALONE_SERVER env var
- Add `POOL_API_KEY` env var if pool auth is desired

### TKNet-Hermes-Hannah repo
- `SOUL.md` — update SwarmOps section with new tool names and routing guidance

## Testing

1. After MCP server rename: verify `swo_health_check` responds
2. After pool tools: call `swo_pool_status`, then `swo_pool_chat` with each model
3. After model sessions: create session with `model: "haiku"`, verify claude process uses haiku
4. After config tools: read config, set a value, verify it persists
5. End-to-end: Hermes conversation → swo_pool_chat(model="haiku") → verify haiku slot used

## Risk & Rollback

- Rename is the riskiest change (breaks existing tool calls). Mitigated by temporary aliases.
- Pool chat goes through existing /v1/chat/completions — battle-tested.
- Go changes are additive (new optional param) — backward compatible.
- If anything breaks, revert the MCP server deploy. SwarmOps Go binary is unaffected by MCP changes.
