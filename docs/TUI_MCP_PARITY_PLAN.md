# TUI / MCP Session Parity — Audit & Plan

**Status:** Part A (fragility fix) **implemented** in this PR; Parts B–D remain proposals.
**Author:** automated audit
**Scope:** Unify TUI-created and MCP-created SwarmOps sessions per Simon's 5 requirements.

> **Update:** The primary fragility identified below — five hand-rolled `claude`
> argv builders that could silently drift — has been **fixed** in this PR. All
> interactive paths now route through a single `interactiveClaudeArgs` helper, and
> the `--dangerously-skip-permissions` literal lives in exactly one `const`. Golden
> + invariant tests pin the behaviour. See **Part A** (marked ✅ DONE) for details.
> Parts B (TUI worktree spawn), C (nil-DB spawner guard), and D (docs) are still
> open proposals.

---

## TL;DR

**Most of the requested parity already exists in the current fork** — largely delivered by
commit `f6b0bac` *"feat(spawn): native claude sessions with Remote Control (remove happier)"* (#12)
and the `happier` removal in #13. The problem statement in `TASK.md` describes a **pre-#12
architecture**:

- The strings `state_source`, `external agent`, `isExternal`, etc. **do not exist anywhere in the
  current Go code** — the only matches are inside `TASK.md` itself. There is no longer an
  "external agent" rendering path or a `state_source` field to unify.
- TUI-created and MCP-created sessions already share **one** persistence store
  (`managed_sessions`), **one** creation function (`spawnSession`), and **one** sidebar rendering
  path. Both already get `--dangerously-skip-permissions` and Remote Control on by default.

So the remaining work is **not** a large re-plumbing. It is (a) a **consolidation refactor** to
remove the structural risk that the five hand-rolled `claude` command builders drift apart, plus
(b) two genuine, smaller asymmetries worth closing, and (c) documentation of what is intentionally
separate (the headless pool). Each requirement is mapped to its current state below.

---

## How sessions are actually created today

There are **two** session models in this codebase, and only one of them is the subject of this task:

### 1. Interactive tmux sessions (`managed_sessions`) — the unified path
Every interactive session — regardless of origin — is created by **`spawnSession`** in
[`spawn.go:75`](../spawn.go) and stored in the SQLite `managed_sessions` table by
**`createSession`** in [`session.go:90`](../session.go) with tmux name `sw-<id>`.

Callers of `spawnSession`:

| Origin | Entry point | Path to `spawnSession` |
|---|---|---|
| MCP `swop_run_task` | `mcp_tools_write.go:13` → `svc.RunTask` | `services.go:95` → `spawnSession` |
| MCP `swop_spawn_agent` | `mcp_tools_write.go:66` → `svc.SpawnAgent` | `services.go:115` → (worktree) → `spawnSession` |
| MCP `swop_smart_spawn` / REST smart-spawn | `api_slim.go:665` → `svc.SpawnAgent` | → `spawnSession` |
| **TUI "new session"** | `tui.go:1655` `m.spawner.Spawn` → `apiClient.Spawn` (`tui_client.go:49`) | `POST /api/swarm/sessions` → `api_slim.go:422` → `spawnSession` |
| REST `POST /api/swarm/sessions` | `api_slim.go:406` | `spawnSession` |

Key point: in `runTUIClient` (`main.go:235`) the TUI is a **separate process** talking to the
backend over HTTP. `initialModel` sets `spawner = api` whenever `api != nil`
([`tui.go:292`](../tui.go)), so the TUI's "new session" goes over HTTP to the backend and lands in
the same `spawnSession`. The in-process `defaultSpawner{}` branch is **test-only** (it would touch
a nil `database` in a client TUI).

`spawnSession` always applies, for every session:
```go
sessionCmd = append(sessionCmd, remoteControlArgs(name)...)                 // Remote Control ON
sessionCmd = append(sessionCmd, "--session-id", claudeUUID,
                                 "--dangerously-skip-permissions")           // skip-permissions
```
(`spawn.go:142–143`). `remoteControlArgs` (`spawn.go:54`) returns `--remote-control <name>` unless
the `SWARMOPS_DISABLE_REMOTE_CONTROL` kill-switch is set.

### 2. Headless pool workers (`claude -p` stream-json) — intentionally separate
`swarm_pool.go:332` builds a **different** kind of process: `claude -p --input-format stream-json
--output-format stream-json --no-session-persistence …`. These are ephemeral batch workers, not
tmux sessions; they render as `itemPoolSlot` in the sidebar (separate from `itemSession`) and are
**not** in `managed_sessions`. They cannot be driven by `swop_send_input` and have no Remote
Control (headless, non-interactive). **This is by design and out of scope** for TUI/MCP session
parity — but the plan notes it explicitly so it isn't mistaken for a divergence to "fix".

---

## Requirement-by-requirement audit

| # | Requirement | Current state | Gap? |
|---|---|---|---|
| 1 | TUI & MCP sessions identical (behavior/permissions/interaction) | Both → `spawnSession`; both stored in `managed_sessions`; both rendered identically in the sidebar (`tui.go:419`). No `state_source`/"external" concept exists. | **Largely met.** Residual: command construction is duplicated 5× (drift risk). |
| 2 | Both use `--dangerously-skip-permissions`, no exceptions | Present in **all five** interactive builders: `spawn.go:143`, `restore.go:93`, `resumeClaudeCmd` `tui.go:2532`, alt+S inline `tui.go:944`, and `StartSession`→`resumeClaudeCmd`. | **Met today**, but only by five copies agreeing. No single source of truth → structurally fragile. |
| 3 | Both visible in TUI sidebar | Sidebar is built from `api.listSessions()` → `GET /api/swarm/sessions` → `listSessions` → **all** `managed_sessions` (`tui.go:387`–`437`). MCP-spawned sessions already appear. | **Met.** |
| 4 | Both interactable via MCP (`swop_send_input`, `swop_stop`, …) | `SendInput`/`StopSession` look the session up by ID and act on its `TmuxSession` (`services.go:186`, `:203`). Works for any `managed_sessions` row irrespective of origin. | **Met.** |
| 5 | Remote Control ON by default | `remoteControlArgs` returns the flag for every interactive spawn/restore/resume unless explicitly disabled. | **Met.** |

---

## The genuine remaining work

### A. (Primary) Collapse the five `claude` command builders into one source of truth — ✅ DONE in this PR
**Problem:** Requirement #2 ("no exceptions") is currently satisfied by coincidence — five separate
functions each hand-append `--dangerously-skip-permissions` and the Remote Control flags. If a future
edit touches one and misses another, parity silently breaks. The duplicated builders are:

1. `spawn.go:141–149` — `spawnSession` (fresh create)
2. `restore.go:86–104` — `buildClaudeRestoreArgs` (post-reboot restore, with `--resume`)
3. `tui.go:2529–2540` — `resumeClaudeCmd` (TUI resume + `StartSession`)
4. `tui.go:941–945` — alt+S inline (fresh restart, no history)
5. *(headless pool, `swarm_pool.go:332` — separate model; leave as-is but add a shared constant for the flag literal)*

**Change:** Introduce one helper in `spawn.go` (next to `remoteControlArgs`), e.g.:

```go
// interactiveClaudeArgs builds the full `claude` argv for an interactive tmux
// session. resume="" → fresh --session-id; resume=<uuid> → --resume. This is the
// single source of truth for the flags every interactive session must carry
// (Remote Control + --dangerously-skip-permissions). Do not hand-roll elsewhere.
func interactiveClaudeArgs(opts interactiveSpawnOpts) []string { … }
```

Have `spawnSession`, `buildClaudeRestoreArgs`, `resumeClaudeCmd`, and the alt+S inline block all call
it. Define `const dangerouslySkipPermissions = "--dangerously-skip-permissions"` once and reference it
(including from the pool builder) so a grep for the literal lands in exactly one place.

**Why this is the right primary change:** it converts requirement #2 from "true today" into
"true by construction," and it removes the only real way the TUI and MCP paths can diverge going forward.

**As implemented (this PR):**
- `spawn.go` — added `const dangerouslySkipPermissions`, a `claudeFresh` / `claudeResume`
  mode enum, `interactiveClaudeOpts`, and `interactiveClaudeArgs`. `spawnSession` now calls it.
- `restore.go` `buildClaudeRestoreArgs`, `tui.go` `resumeClaudeCmd`, and the `tui.go` alt+S
  fresh-restart block now all call the shared helper.
- `swarm_pool.go` references the shared `const` for the flag literal (headless pool keeps its
  own command shape — it is not interactive — but can no longer disagree on the flag spelling).
- Peer-review refinements adopted: an explicit **mode enum** (not a `resume bool`); a
  `modelFlag` field with an explicit "model via ANTHROPIC_MODEL on restore" comment; and the
  **exact original argv ordering preserved** (`--session-id … --dangerously-skip-permissions`
  on fresh) so this is a behaviour-preserving refactor. `remoteControlArgs` was intentionally
  left unchanged (switching to `--remote-control=<name>` would be a gratuitous behaviour change).
- Tests: `TestInteractiveClaudeArgsGolden` pins exact argv for all fresh/resume permutations;
  `TestInteractiveClaudeArgsInvariants` asserts every interactive call site emits
  `--dangerously-skip-permissions` (always, even under the RC kill-switch) and Remote Control by
  default. Existing `TestResumeClaudeCmdNative` / `TestRemoteControlArgs` still pass unchanged.

### B. (Secondary) Let the TUI spawn worktree-backed agents (true #1 parity)
**Asymmetry:** MCP can create a **worktree-isolated** agent (`swop_spawn_agent` → `SpawnAgent`), but the
TUI's "new session" only hits `POST /api/swarm/sessions` → `spawnSession` (no worktree). So one capability
exists on MCP that the TUI cannot reach. If "TUI and MCP identical" is read strictly, this is the one
real behavioral gap.

**Change (optional, but recommended for strict parity):**
- Add `POST /api/swarm/agents` to `api_slim.go` that calls `globalServices.SpawnAgent` (repo_path, branch,
  task_brief, env_overrides).
- Add an `apiClient.SpawnAgent` method to `tui_client.go` and a TUI affordance (e.g. a key/flow that
  asks for repo + branch and spawns a worktree agent).
- Reuse the existing teardown (`swop_teardown_agent` / `TeardownAgent`) for cleanup from the TUI.

**Trade-off:** adds UI surface and a new endpoint. If Simon considers worktree-spawn an
MCP-/automation-only feature, skip B and instead document that the TUI intentionally creates
in-place (non-worktree) sessions.

### C. (Hardening) Make the in-process spawner path safe or remove it
`initialModel` falls back to `defaultSpawner{}` when `api == nil` (`tui.go:292–297`). In a real client
TUI that path would call `spawnSession` against a nil global `database` and panic. It is currently
only reached by tests, but it's a latent footgun that contradicts "one path."

**Change:** either (i) delete the `defaultSpawner` fallback and require a non-nil client (tests inject
`fakeSwarmClient`), or (ii) guard `spawnSession`/`createSession` with a clear error when `database == nil`.
Low effort, removes ambiguity about "which path runs."

### D. (Docs) Record what is intentionally not unified
Add a short note (in `README.md` or `SWARMOPS_V3_PLAN.md`) stating that **pool workers** (`claude -p`,
`itemPoolSlot`) are a deliberately separate, headless model — no tmux, no Remote Control, not in
`managed_sessions`, not driveable by `swop_send_input` — so future readers don't try to "fix" the
difference. Also document the `SWARMOPS_DISABLE_REMOTE_CONTROL` kill-switch as the one supported way
Remote Control is ever off.

---

## Files that would change

| File | Change | Part |
|---|---|---|
| `spawn.go` | Add `interactiveClaudeArgs` + `dangerouslySkipPermissions` const; refactor `spawnSession` to use them | A |
| `restore.go` | `buildClaudeRestoreArgs` calls shared helper | A |
| `tui.go` | `resumeClaudeCmd` + alt+S inline call shared helper | A |
| `swarm_pool.go` | Reference the shared flag const (no behavior change) | A |
| `api_slim.go` | New `POST /api/swarm/agents` → `SpawnAgent` | B |
| `tui_client.go` | `apiClient.SpawnAgent` + interface method | B |
| `tui.go` | TUI flow/keybinding to spawn a worktree agent | B |
| `tui.go` | Remove/guard `defaultSpawner` nil-DB fallback | C |
| `README.md` / `SWARMOPS_V3_PLAN.md` | Document pool-as-separate + RC kill-switch | D |
| tests (`spawn_test.go`, `tui_test.go`, `mcp_server_test.go`) | Cover shared builder; assert every interactive path emits skip-permissions + RC | A/B/C |

---

## Risks & trade-offs

- **Refactor regression (A):** the five builders have subtle differences (`--resume` vs `--session-id`,
  per-session `--strict-mcp-config`/`--mcp-config`, model via flag in spawn vs via `ANTHROPIC_MODEL` env in
  restore). The shared helper must take options for all of these or it will quietly drop a flag on one path.
  **Mitigation:** table-driven tests asserting exact argv per scenario (fresh, resume, restricted-MCP,
  LiteLLM-routed) before and after the refactor.
- **Behavioral surface (B):** adding TUI worktree-spawn introduces worktree lifecycle into the TUI
  (creation + teardown + branch hygiene). If teardown isn't wired, users could accumulate orphaned
  worktrees. **Mitigation:** ship B only with a teardown affordance, or defer B.
- **"Identical" vs "appropriate" (B):** strict identity would also imply the TUI can route via LiteLLM
  (`[gpt]`/`[dseek]`) and set `mcp_servers` subsets like `swop_spawn_agent` does. Confirm with Simon how
  far "identical" extends before building UI for every MCP arg.
- **Low risk for C/D:** guarding a nil DB and adding docs are safe.

---

## Suggested implementation order

1. **A — consolidation refactor** (highest value, lowest user-visible risk). Land the shared
   `interactiveClaudeArgs` + flag const + tests first; this locks requirement #2 structurally and is a
   prerequisite mindset for everything else.
2. **C — guard/remove `defaultSpawner`** (tiny, removes the only "second path" ambiguity).
3. **D — documentation** of pool separation + RC kill-switch (cheap, prevents future confusion).
4. **B — TUI worktree-spawn** *(only if Simon wants the TUI to reach `swop_spawn_agent`-style
   isolation; otherwise replace with a one-line doc that the TUI intentionally spawns in-place)*.

---

## Verification checklist (post-implementation)

- [x] `grep -rn '"--dangerously-skip-permissions"' --include=*.go` (non-test) returns exactly **one**
      site — the `const` in `spawn.go`. All other usages reference the const. *(Done — Part A.)*
- [ ] A session created via `swop_run_task`, via `swop_spawn_agent`, and via the TUI all show in the TUI
      sidebar with identical indicators and are all driveable by `swop_send_input` / `swop_stop_session`.
- [ ] Restart (alt+S), resume (`StartSession`), and post-reboot restore all emit Remote Control +
      skip-permissions (asserted by tests).
- [ ] Pool workers remain headless and clearly separate (no regression in `swarm_pool_test.go`).
