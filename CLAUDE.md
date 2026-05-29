# SwarmOps — Claude Code Project Guide

SwarmOps is a Go + Bubbletea TUI that manages Claude Code sessions via tmux.
It exposes a JSON/HTTP API, an MCP server, and a terminal UI client.

---

## Architecture at a Glance

| Layer | Files | Purpose |
|-------|-------|---------|
| **TUI** | `tui.go`, `tui_client.go`, `tui_popups.go` | Bubbletea model + rendering |
| **API** | `api_slim.go`, `api_task_bus.go` | HTTP handlers |
| **MCP** | `mcp_server.go`, `mcp_tools_*.go` | MCP protocol layer |
| **Backend** | `backend.go`, `session.go`, `spawn.go` | Session lifecycle |
| **Persistence** | `database.go`, `persist.go` | SQLite via modernc |
| **Pool** | `swarm_pool.go`, `swarm_openai.go` | API-mode session pool |

Entry point: `main.go` — subcommand `tui` runs `runTUIClient()` (HTTP client connecting to the backend on :8080), no subcommand starts the server.

---

## Key Constraints

- **All Go code must be gofmt-clean** before committing. Run `gofmt -l .` and fix any output.
- **Tests run without a live backend** — the TUI uses `fakeSwarmClient` / `mockSpawner` interfaces. Never make tests depend on a running swarmops instance or tmux.
- **No test helpers in production code** — test-only helpers live in `_test.go` files only.
- **SQLite only** — no other DB dependencies.
- **No external process calls in tests** — mock tmux/exec interactions via interfaces.
- **`scratch/` is excluded from builds** (build tag `ignore` on those files).

---

## Working on the TUI

The TUI is a standard Bubbletea `Model/Update/View` pattern:

- `tuiModel` in `tui.go` is the main model struct.
- Modes (`modePassthrough`, `modeNewName`, `modeNewDir`, etc.) control key routing.
- Use `newTestModel(items)` in tests to create a headless model with fixed 80×24 dimensions.
- Use `fakeSwarmClient` for key-handler tests that exercise the client API.
- Golden files live in `testdata/` — regenerate with `UPDATE_GOLDEN=1 go test ./...`.

### Bubbletea rules
- `View()` must be pure/deterministic for a given model state (no time.Now() in View).
- Side effects belong in `Init()` / `Update()` as `tea.Cmd` returns, never in `View()`.
- Styles are defined at package level in `tui.go`; don't redefine inline.
- Use `lipgloss` for layout/color — no raw ANSI escapes.

---

## Running & Testing

```bash
make build          # compile binary
make test           # run all tests (no -race)
make test-race      # run with race detector
make vet            # go vet
make ci             # vet + test-race (pre-merge gate)
make status         # show running process + systemd health
make logs           # tail journalctl for the swarmops service
make restart        # build + systemctl --user restart + health check
```

### Slash commands (project-local)

| Command | Purpose |
|---------|---------|
| `/swarm-ci` | Run full CI gate (vet + test-race) |
| `/swarm-restart` | Build, stop port 8080, restart service |
| `/swarm-tmux` | Show active tmux sessions for swarmops |
| `/swarm-review` | Peer code review via tkn-aipeer |
| `/swarm-test-tui` | Run TUI-specific tests with verbose output |

---

## Deployment

SwarmOps runs as a systemd user service on `nuc-ubuntu-dev`.
The binary is at `~/swarmops/swarmops` (not in this repo's path).

Use `make restart` after building — it builds, kills port 8080, restarts the service, and health-checks.

Do **not** run `make restart` during active TUI sessions (it kills the backend the TUI is connected to).

---

## Prerequisites

These tools must be present for the full dev workflow:

| Tool | Required for |
|------|-------------|
| `python3 ≥ 3.8` | Branch-guard hook in `.claude/settings.local.json` |
| `tmux` | Session management (core functionality) |
| `systemctl --user` | `make restart / status / logs` (Linux/systemd only) |
| `journalctl` | `make logs` (Linux/systemd only) |

The `make status`, `make logs`, and `make restart` targets are **nuc-ubuntu-dev specific** — they rely on the `swarmops.service` systemd unit running on that host. They will not work on macOS or in containers.

`make fmt`, `make test`, `make ci`, and all `go` commands are fully portable.

### Hooks (local-only setup)

The branch-guard and gofmt hooks live in `.claude/settings.local.json`, which is globally gitignored (each developer maintains their own). To set up the branch guard locally, add to `.claude/settings.local.json`:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{"type": "command", "command": "python3 -c \"...<see branch in repo>...\"]
    }]
  }
}
```

### Formatting debt

13 pre-existing `.go` files have formatting debt (not addressed in any single branch to avoid noise). Run `make fmt` to bulk-format. The pre-commit branch-guard hook does **not** auto-format files — format explicitly with `make fmt` before committing.

## Branch Policy

- **Never commit directly to `main`** — always work on `agent/<slug>` branches.
- The pre-commit hook runs `gofmt -l .` and blocks if any files are not formatted.
- Merge via PR after `/swarm-review` passes.

---

## Test File Map

| File | What it tests |
|------|--------------|
| `tui_test.go` | View contracts, key handling, golden files |
| `tui_integration_test.go` | Multi-step state transitions |
| `tui_step_test.go` | Step-through wizard flows |
| `tui_popups_test.go` | Popup mode rendering & key handling |
| `tui_fuzz_test.go` | Fuzz targets for key handling |
| `mcp_server_test.go` | MCP protocol layer |
| `persist_test.go` | SQLite persistence |
| `task_bus_test.go` | Task bus message routing |
| `swarm_pool_test.go` | Pool slot lifecycle |
| `cache_probe_test.go` | Cache probe logic |

Integration tests that need a live DB use `t.Setenv("SWARMOPS_DB", ...)` to get an isolated temp DB.
