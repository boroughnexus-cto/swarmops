Run a peer code review of the current branch changes using the `tkn-aipeer` MCP server.

**Note:** `tkn-aipeer` is a TKN-stack internal MCP server (not a public tool). It must be mounted in your Claude Code session — check `list_tool_servers()` if uncertain. If unavailable, use `mcp__tkn-aipeer__peer_review` directly, or perform a manual review using the checklist in `.claude/agents/go-bubbletea-review.md`.

Steps:
1. Get the diff: `git diff main...HEAD` (or `git diff HEAD~1` if on main)
2. Identify changed files and their purpose
3. Call `mcp__tkn-aipeer__peer_review` with:
   - `code`: the diff
   - `context`: "SwarmOps Go/Bubbletea TUI — session manager for Claude Code"
   - `review_type`: "go" (or let it auto-detect)
4. Report findings grouped by severity: Critical → High → Medium → Low → Info

Focus areas for SwarmOps:
- Bubbletea model purity (no side effects in View, no time.Now() in View)
- Interface contract compliance (swarmClient, spawner interfaces)
- Test coverage for new key handlers or state transitions
- Concurrency safety (goroutines, channels in the poll loop)
- SQL injection risks in database.go / persist.go
- gofmt compliance
