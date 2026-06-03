# SwarmOps

Terminal session manager for Claude Code. Manage multiple Claude Code sessions from a single TUI with an OpenAI-compatible API for programmatic access.

## Features

- **TUI** — Left sidebar lists sessions, right pane shows live terminal output. Navigate with `ctrl+a`/`ctrl+z`, type to inject commands.
- **Session management** — Spawn Claude Code in named tmux sessions with optional context from mcp-context.
- **OpenAI API** — Warm pool of Claude CLI processes serving `/v1/chat/completions` for tool integrations.
- **REST API** — Consumed by the tkn-swarmops MCP server for remote control.

## Quick Start

```bash
# Server mode (API on :8080)
make run

# TUI mode
./swarmops tui
```

## TUI Keybindings

| Key | Action |
|-----|--------|
| `ctrl+a` / `up` | Move cursor up |
| `ctrl+z` / `down` | Move cursor down |
| `Enter` | Focus input (type to send to session) |
| `Esc` | Back to sidebar |
| `n` | New session |
| `d` | Delete session |
| `q` | Quit |

## Configuration

Pool settings are stored in SQLite (`system_config` table) and configurable via the API:

| Key | Default | Description |
|-----|---------|-------------|
| `pool.enabled` | `false` | Enable warm pool + /v1/ API |
| `pool.models` | `claude-haiku-4-5,claude-sonnet-4-6,claude-opus-4-6` | Models to pool |
| `pool.slots_per_model` | `2` | Warm instances per model |
| `pool.api_key` | (empty) | Bearer token for /v1/ |

Environment variables (`POOL_ENABLED`, `POOL_MODELS`, etc.) override defaults.

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/dashboard/stats` | Session count + pool status |
| `GET /api/tmux-sessions` | List tmux sessions |
| `GET /api/git/status?path=.` | Git status for a directory |
| `GET /api/git/branches?path=.` | Git branches |
| `POST /api/swarm/sessions` | Create a new session |
| `GET /api/swarm/sessions` | List sessions |
| `DELETE /api/swarm/sessions/:id` | Delete a session |
| `GET /api/swarm/sessions/:id/terminal` | Capture terminal output |
| `POST /api/swarm/sessions/:id/input` | Inject text into session |
| `GET /api/swarm/config` | List config |
| `PUT /api/swarm/config/:key` | Set config value |
| `POST /v1/chat/completions` | OpenAI-compatible chat |
| `GET /v1/models` | List available models |

## Development

```bash
make build      # Build both binaries (swarmops + quota-proxy), atomically
make test       # Run tests
make ci         # Merge gate: gofmt-check + vet + race tests
make fmt        # gofmt the tree
make clean      # Remove binaries + test DBs + deploy artifacts
```

## Deployment

There is exactly **one** way SwarmOps reaches production, and it is CI-gated.
Do **not** build from a worktree and copy a binary into `~/swarmops` — that is
how stale builds used to silently drop merged features.

- **Canonical branch:** `main`. It is branch-protected: changes land only via PR
  with the `ci` check green and the branch up to date. So `origin/main` is always
  gated-green.
- **The deploy primitive:** [`scripts/deploy.sh`](scripts/deploy.sh) is the sole
  deploy path. It `flock`s, hard-resets `~/swarmops` to the target commit,
  rebuilds both binaries atomically, restarts the service, health-checks, and
  **rolls back** to the previous commit + binaries if the health check fails.
- **Automatic:** on every push to `main`, the `deploy` job (GitHub Actions,
  self-hosted `nuc-ubuntu-dev` runner) runs `scripts/deploy.sh` pinned to the
  exact CI-validated commit (`DEPLOY_SHA=github.sha`).
- **Manual fallback** (runner offline, or redeploy current `main`):

  ```bash
  swarmops redeploy            # delegates to scripts/deploy.sh (deploys origin/main)
  swarmops redeploy --force    # also when ~/swarmops has uncommitted tracked edits
  make deploy                  # equivalent
  ```

  `make restart` only rebuilds + restarts the *current checkout in place* — it
  does not sync to `origin/main`, so it is for local iteration, not production.

## Architecture

```
main.go           — Entry point (server or TUI mode)
session.go         — Session CRUD + tmux capture/inject
spawn.go           — tmux session creation + claude launch
tui.go             — Bubbletea TUI
api_slim.go        — REST API
swarm_config.go    — Config service (system_config table)
swarm_pool.go      — Warm Claude CLI session pool
swarm_openai.go    — OpenAI-compatible API handlers
database.go        — SQLite migrations
```
