# SwarmOps Changelog

All notable changes are documented here.

## [Unreleased]

### Fixed
- **Sidebar header rendering**: Gap calculation used `lipgloss.Width(dimStyle.Render(ts))` but compared it against `sidebarInnerWidth` using byte-length math. Replaced with plain `fmt.Sprintf("%-8s %8s", "SwarmOps", ts)` — explicit 22-char fixed-width layout that avoids any lipgloss ANSI-width ambiguity.
- **`-X main.BuildCommit` was silently ignored**: `BuildCommit` variable was never declared in `main.go`, so the ldflags injection was a no-op. Added `var BuildCommit string` declaration.

### Added
- **Version display in TUI**: Sidebar header now shows `v<commit>` below the time (e.g. `vff90628`). Only visible when the binary was built with the ldflags injection.
- **Ctrl+\[ debug hotkey** (modePassthrough only): Writes per-line char counts and model dimensions to `/tmp/sidebar-debug.txt`.
- **WindowSizeMsg debug logging**: Every resize event now logs `w`, `h`, `contentWidth`, `sidebarWidth` to the server log for tracing width-flip issues.
- **renderSidebar LOW h logging**: Logs when `m.h < 20` for height-anomaly detection.

### Changed
- **Sidebar width now adjustable** (Shift+Alt+←/→): `sidebarWidth` range 18–40, persisted for the session lifetime. Affects both the sidebar rendering and `contentWidth = m.w - (sidebarWidth + 2)`.
- **resizeTmuxSessions guards**: Now skips sessions with attached tmux clients (client size takes precedence) and skips calls where dimensions haven't changed since last call (avoids redundant `tmux resize-window` that was causing the SIGWINCH width-flip).
- **sidebarStyle() method**: Changed from global `var sidebarStyle` to `m.sidebarStyle()` method, enabling dynamic width without recreating lipgloss styles each render.

## [2.0] — 2026-05-??

Initial MCP server + TUI release. Session lifecycle management, Plane/Livy/HApp integrations, warm pool, audit log, Icinga alerts.
