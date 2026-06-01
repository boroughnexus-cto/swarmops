package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// ─── TUI state sharing (TUI process → API server) ─────────────────────────────

var tuiState struct {
	sync.RWMutex
	json string // latest rendered View() output
}

var pendingTUIKey struct {
	sync.Mutex
	key string // pending key to inject; "" = none
}

// TUIStateHandler accepts state POSTs from the TUI process and stores them.
func handleTUIStateAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"POST only"}`, http.StatusMethodNotAllowed)
		return
	}
	var state struct {
		Rendered  string `json:"rendered"`
		State     string `json:"state"` // JSON snapshot of key fields
		Timestamp int64  `json:"timestamp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"bad json: %v"}`, err), http.StatusBadRequest)
		return
	}
	tuiState.Lock()
	tuiState.json = state.Rendered
	tuiState.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// handleTUIKeyAPI handles both GET /tui/key (TUI polls) and POST /tui/key (external injects).
func handleTUIKeyAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		pendingTUIKey.Lock()
		k := pendingTUIKey.key
		pendingTUIKey.key = "" // consume
		pendingTUIKey.Unlock()
		if k == "" {
			w.Write([]byte(`{}`))
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"key": k})
		return
	}
	if r.Method == "POST" {
		var req struct{ Key string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Key == "" {
			http.Error(w, `{"error":"key field required"}`, http.StatusBadRequest)
			return
		}
		pendingTUIKey.Lock()
		pendingTUIKey.key = req.Key
		pendingTUIKey.Unlock()
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Error(w, `{"error":"GET or POST only"}`, http.StatusMethodNotAllowed)
}

// TUIStateGet returns the latest TUI state snapshot (GET).
func handleTUIStateGetAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, `{"error":"GET only"}`, http.StatusMethodNotAllowed)
		return
	}
	tuiState.RLock()
	rendered := tuiState.json
	tuiState.RUnlock()
	if rendered == "" {
		http.Error(w, `{"error":"TUI state not yet pushed — is the TUI running?"}`, http.StatusServiceUnavailable)
		return
	}
	// Decode the stored state JSON to return it directly
	json.NewEncoder(w).Encode(map[string]string{"rendered": rendered})
}

// handleAPI is the main REST API router. Maintains URL contract for the
// tkn-swarmops MCP server wrapper (and the historical tkn-remote-code-*
// names it replaced).
func handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/")
	pathParts := strings.Split(path, "/")
	if len(pathParts) == 0 {
		http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	switch pathParts[0] {
	case "dashboard":
		handleDashboardAPI(w, r, ctx)
	case "agents":
		// MCP: swop_list_agents → returns sessions as "agents"
		handleSessionsAsAgentsAPI(w, r, ctx)
	case "task-executions":
		// MCP: swop_list_executions, swop_run_task, swop_send_input
		handleTaskExecutionsAPI(w, r, ctx, pathParts[1:])
	case "tmux-sessions":
		handleTmuxSessionsAPI(w, r)
	case "git":
		handleGitAPI(w, r, pathParts[1:])
	case "roots":
		handleRootsAPI(w, r)
	case "projects":
		handleProjectsAPI(w, r)
	case "swarm":
		handleSwarmSubAPI(w, r, ctx, pathParts[1:])
	case "tui":
		// TUI state push: POST /api/tui/state
		// TUI key poll/inject: GET|POST /api/tui/key
		// TUI view snapshot: GET /api/tui/view
		if len(pathParts) < 2 {
			http.Error(w, `{"error":"tui endpoint required"}`, http.StatusBadRequest)
			return
		}
		switch pathParts[1] {
		case "state":
			handleTUIStateAPI(w, r)
		case "key":
			handleTUIKeyAPI(w, r)
		case "view":
			handleTUIStateGetAPI(w, r)
		default:
			http.Error(w, `{"error":"tui/state, tui/key, or tui/view required"}`, http.StatusBadRequest)
		}
		return
	default:
		http.Error(w, `{"error":"unknown endpoint"}`, http.StatusNotFound)
	}
}

// ─── Dashboard ───────────────────────────────────────────────────────────────

type DashboardStats struct {
	ActiveSessions           int           `json:"active_sessions"`
	Projects                 int           `json:"projects"`
	TaskExecutions           int           `json:"task_executions"`
	Agents                   int           `json:"agents"`
	GitChangesAwaitingReview []interface{} `json:"git_changes_awaiting_review"`
	AgentsWaitingForInput    []interface{} `json:"agents_waiting_for_input"`
	RemotePorts              []interface{} `json:"remote_ports"`
	DirectoryDevServers      []interface{} `json:"directory_dev_servers"`
}

func handleDashboardAPI(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	stats, _ := globalServices.Dashboard(ctx)
	_ = json.NewEncoder(w).Encode(stats)
}

// ─── Sessions as Agents (MCP compat) ────────────────────────────────────────

func handleSessionsAsAgentsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	agents, _ := globalServices.ListAgents(ctx)
	_ = json.NewEncoder(w).Encode(agents)
}

// ─── Task Executions (MCP compat) ───────────────────────────────────────────

func handleTaskExecutionsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case http.MethodGet:
		// swop_list_executions → return sessions as executions
		sessions, _ := listSessions(ctx)
		_ = json.NewEncoder(w).Encode(sessions)

	case http.MethodPost:
		// swop_run_task → create a new session
		if len(pathParts) > 0 {
			// Handle sub-resources like /:id/input
			if len(pathParts) >= 2 && pathParts[1] == "input" {
				handleSessionInput(w, r, ctx, pathParts[0])
				return
			}
		}
		var req struct {
			Name      string `json:"name"`
			Directory string `json:"directory"`
			Model     string `json:"model"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Name == "" {
			req.Name = "session-" + generateID()
		}
		if req.Directory == "" {
			req.Directory = "."
		}
		s, err := spawnSession(ctx, req.Name, req.Directory, nil, req.Model, "", nil)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(s)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleSessionInput(w http.ResponseWriter, r *http.Request, ctx context.Context, sessionID string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Input string `json:"input"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	s, err := getSession(ctx, sessionID)
	if err != nil {
		http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
		return
	}
	if err := injectToSession(s.TmuxSession, req.Input); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ─── Tmux Sessions ──────────────────────────────────────────────────────────

type TmuxSessionInfo struct {
	Name    string `json:"name"`
	Created string `json:"created"`
	Preview string `json:"preview"`
}

func handleTmuxSessionsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sessions, _ := globalServices.TmuxSessions()
	json.NewEncoder(w).Encode(sessions)
}

// ─── Git ─────────────────────────────────────────────────────────────────────

func handleGitAPI(w http.ResponseWriter, r *http.Request, pathParts []string) {
	if len(pathParts) == 0 {
		http.Error(w, `{"error":"missing git subcommand"}`, http.StatusBadRequest)
		return
	}

	dir := r.URL.Query().Get("path")
	if dir == "" {
		dir = "."
	}

	switch pathParts[0] {
	case "status":
		status, err := globalServices.GitStatus(dir)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(status)

	case "branches":
		branches, err := globalServices.GitBranches(dir, false)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"branches": branches})

	case "diff":
		diff, _ := globalServices.GitDiff(dir, false)
		json.NewEncoder(w).Encode(map[string]string{"diff": diff})

	case "log":
		gitLog, _ := globalServices.GitLog(dir)
		json.NewEncoder(w).Encode(map[string]string{"log": gitLog})

	default:
		http.Error(w, `{"error":"unknown git subcommand"}`, http.StatusBadRequest)
	}
}

func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ─── Roots ───────────────────────────────────────────────────────────────────

func handleRootsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	roots, _ := globalServices.ListRoots(context.Background())
	json.NewEncoder(w).Encode(roots)
}

// ─── Projects (minimal stub) ────────────────────────────────────────────────

func handleProjectsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	json.NewEncoder(w).Encode([]interface{}{})
}

// ─── Swarm sub-API ──────────────────────────────────────────────────────────

func handleSwarmSubAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	if len(pathParts) == 0 {
		http.Error(w, `{"error":"missing swarm subpath"}`, http.StatusNotFound)
		return
	}

	switch pathParts[0] {
	case "config":
		handleConfigAPI(w, r)
	case "pool":
		handlePoolStatusAPI(w, r)
	case "sessions":
		handleSwarmSessionsAPI(w, r, ctx, pathParts[1:])
	case "dashboard":
		handleSwarmDashboardAPI(w, r, ctx)
	case "tasks":
		handleGlobalTasksAPI(w, r, ctx)
	case "audit":
		handleSwarmAuditAPI(w, r, ctx)
	case "smart-spawn":
		handleSmartSpawnAPI(w, r, ctx)
	default:
		http.Error(w, `{"error":"unknown swarm endpoint"}`, http.StatusNotFound)
	}
}

func handleSwarmAuditAPI(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	limit := 100
	events, err := listAuditEvents(ctx, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if events == nil {
		events = []ManagedSessionEvent{}
	}
	json.NewEncoder(w).Encode(events)
}

func handleSwarmSessionsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	if len(pathParts) == 0 {
		switch r.Method {
		case http.MethodGet:
			refreshSessionStatuses(ctx)
			sessions, err := listSessions(ctx)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
				return
			}
			if sessions == nil {
				sessions = []Session{}
			}
			json.NewEncoder(w).Encode(sessions)

		case http.MethodPost:
			var req struct {
				Name         string            `json:"name"`
				Directory    string            `json:"directory"`
				Mission      *string           `json:"mission"`
				Model        string            `json:"model"`
				Profile      string            `json:"profile"`
				EnvOverrides map[string]string `json:"env_overrides"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == "" {
				http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
				return
			}
			if req.Directory == "" {
				req.Directory = "."
			}
			s, err := spawnSession(ctx, req.Name, req.Directory, req.Mission, req.Model, req.Profile, req.EnvOverrides)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(s)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	sessionID := pathParts[0]
	subPath := pathParts[1:]

	if len(subPath) == 0 {
		switch r.Method {
		case http.MethodGet:
			s, err := getSession(ctx, sessionID)
			if err != nil {
				http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(s)

		case http.MethodDelete:
			if err := deleteSession(ctx, sessionID); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodPatch:
			var body struct {
				Name      string  `json:"name"`
				Mission   *string `json:"mission"`
				Profile   *string `json:"profile"`
				Directory *string `json:"directory"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
				return
			}
			if body.Name != "" {
				if err := renameSession(ctx, sessionID, body.Name); err != nil {
					http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
					return
				}
			}
			if body.Mission != nil || body.Profile != nil || body.Directory != nil {
				if err := updateSessionFields(ctx, sessionID, body.Profile, body.Directory, body.Mission); err != nil {
					http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
					return
				}
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	switch subPath[0] {
	case "terminal":
		s, err := getSession(ctx, sessionID)
		if err != nil {
			http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
			return
		}
		content, err := captureTerminal(s.TmuxSession)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"content": content})

	case "input":
		handleSessionInput(w, r, ctx, sessionID)

	case "external-events":
		handleExternalEvents(w, r, ctx, sessionID)

	case "tasks":
		handleSessionTasksAPI(w, r, ctx, sessionID, subPath[1:])

	default:
		http.Error(w, `{"error":"unknown session endpoint"}`, http.StatusNotFound)
	}
}


// handleExternalEvents handles POST /api/swarm/sessions/:id/external-events.
// Injects content into the session terminal. Requires Bearer token auth when
// n8n.events_token is configured.
func handleExternalEvents(w http.ResponseWriter, r *http.Request, ctx context.Context, sessionID string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Bearer token auth (only enforced when token is configured)
	if globalConfigService != nil {
		token := globalConfigService.GetString("n8n.events_token", "")
		if token != "" {
			if r.Header.Get("Authorization") != "Bearer "+token {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
		}
	}

	s, err := getSession(ctx, sessionID)
	if err != nil {
		http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
		return
	}

	var req struct {
		Event   string `json:"event"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		http.Error(w, `{"error":"content required"}`, http.StatusBadRequest)
		return
	}

	if err := injectToSession(s.TmuxSession, req.Content); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
}

func handleSwarmDashboardAPI(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	dashboard, _ := globalServices.SwarmDashboard(ctx)
	json.NewEncoder(w).Encode(dashboard)
}

// handleSmartSpawnAPI handles POST /api/swarm/smart-spawn.
// With dry_run=true it returns the brain's routing pick without spawning.
// With dry_run=false (or repo_slug provided) it creates a worktree + session.
func handleSmartSpawnAPI(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"POST only"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Goal     string `json:"goal"`
		RepoSlug string `json:"repo_slug"`
		DryRun   bool   `json:"dry_run"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"bad JSON"}`, http.StatusBadRequest)
		return
	}
	if req.Goal == "" {
		http.Error(w, `{"error":"goal required"}`, http.StatusBadRequest)
		return
	}

	if globalServices == nil {
		http.Error(w, `{"error":"services not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	repos, err := listKnownRepos(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	rawReply, err := brainAsk(ctx, globalServices, brainDefaultModel, brainSystemPrompt, brainUserPrompt(req.Goal, repos))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}
	pick, err := brainParse(rawReply)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	result := smartSpawnResult{Pick: pick}

	if req.DryRun {
		json.NewEncoder(w).Encode(result)
		return
	}

	// Determine which repo slug to use: explicit override, then brain's pick
	repoSlug := req.RepoSlug
	if repoSlug == "" {
		repoSlug = pick.Pick
	}

	if repoSlug == "" || repoSlug == "none" {
		// No matching repo: spawn a plain session in HOME
		home, _ := os.UserHomeDir()
		name := sanitizeSessionName(req.Goal)
		mission := req.Goal
		s, err := spawnSession(ctx, name, home, &mission, "", "", nil)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Session = s
		}
		json.NewEncoder(w).Encode(result)
		return
	}

	// Find the repo in the registry
	var targetRepo *KnownRepo
	for i := range repos {
		if repos[i].Slug() == repoSlug {
			targetRepo = &repos[i]
			break
		}
	}
	if targetRepo == nil {
		result.Error = fmt.Sprintf("repo not found in registry: %s", repoSlug)
		json.NewEncoder(w).Encode(result)
		return
	}
	if !targetRepo.IsCloned() {
		result.Error = fmt.Sprintf("repo not cloned locally: %s", repoSlug)
		json.NewEncoder(w).Encode(result)
		return
	}

	// Spawn agent in a new worktree inside the repo
	name := "smart-" + sanitizeSlug(req.Goal)
	if len(name) > 30 {
		name = name[:30]
	}
	mission := req.Goal
	s, err := globalServices.SpawnAgent(ctx, name, targetRepo.LocalPath, "", "", &mission, "", "", req.Goal, nil)
	if err != nil {
		result.Error = err.Error()
	} else {
		result.Session = s
	}

	json.NewEncoder(w).Encode(result)
}

func init() {
	// Silence unused import warnings
	_ = log.Printf
}
