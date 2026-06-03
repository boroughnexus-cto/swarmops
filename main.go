package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func init() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}
}

var database *sql.DB
var BuildCommit string // set via -ldflags="-X main.BuildCommit=$(git rev-parse --short HEAD)"

func main() {
	// Subcommand routing
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "tui":
			runTUIClient()
			return
		case "tui-restart":
			runTUIRestart()
			return
		case "redeploy":
			runRedeploy()
			return
		case "mcp":
			if len(os.Args) > 2 && os.Args[2] == "serve" {
				runMCPStdioMode()
				return
			}
			fmt.Fprintf(os.Stderr, "Usage: swarmops mcp serve\n")
			os.Exit(1)
		}
	}

	// TTY with no subcommand → TUI client
	if isTerminal() {
		runTUIClient()
		return
	}

	// Server mode: database, config, HTTP API, then pool (pool spawns are slow)
	database = initDatabase()
	defer database.Close()

	globalConfigService = newConfigService(database)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP server BEFORE pool so health checks pass during pool startup
	server := newHTTPServer()

	// Bind the port first so we fail fast if it is already taken
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatalf("Cannot bind %s: %v", server.Addr, err)
	}
	serverErr := make(chan error, 1)
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()
	log.Printf("SwarmOps server listening on %s", server.Addr)

	// Start + supervise the quota-proxy sidecar. Sessions route their
	// ANTHROPIC_BASE_URL through it (see spawn.go) to capture Anthropic
	// rate-limit headers; the watchdog restarts it on crash. Tied to ctx, so
	// shutdown (cancel) stops the child automatically.
	exePath, _ := os.Executable()
	superviseQuotaProxy(ctx, filepath.Join(filepath.Dir(exePath), "quota-proxy"), quotaProxyPort)

	// Pool init is slow (spawns 6 Claude CLI sessions) — runs after HTTP is up
	initPool(ctx)

	// Update services with pool reference (pool was nil when HTTP server started)
	if globalServices != nil && globalPool != nil {
		globalServices.pool = globalPool
	}

	// Session persistence: prune orphaned snapshots, then restore sessions async
	pruneOrphanedSnapshots(ctx)
	go restoreSessions(ctx)

	// Task bus: TTL expiry + deferred task re-queue
	startTTLSweeper(ctx)

	// Periodic session status sync + scrollback saves
	go func() {
		syncTicker := time.NewTicker(10 * time.Second)
		saveTicker := time.NewTicker(60 * time.Second)
		defer syncTicker.Stop()
		defer saveTicker.Stop()

		// Wait for restore to complete before starting saves
		select {
		case <-restoreComplete:
		case <-ctx.Done():
			return
		}

		for {
			select {
			case <-syncTicker.C:
				syncTmuxSessions()
			case <-saveTicker.C:
				saveAllScrollbacks(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sig:
		log.Printf("Received shutdown signal")
	case err := <-serverErr:
		log.Printf("HTTP server error: %v", err)
	}

	cancel() // also signals superviseQuotaProxy to stop the sidecar

	// Final scrollback save before shutdown
	log.Printf("Saving session scrollbacks before shutdown...")
	<-restoreComplete // ensure restore finished before saving
	shutdownSaveCtx, shutdownSaveCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownSaveCancel()
	saveAllScrollbacks(shutdownSaveCtx)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
}

// deployDirPath returns the canonical SwarmOps deploy checkout — the one place
// production is built and run from. Override with SWARMOPS_DEPLOY_DIR.
func deployDirPath() string {
	if d := strings.TrimSpace(os.Getenv("SWARMOPS_DEPLOY_DIR")); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "swarmops")
}

// deployScriptPath validates the deploy dir and returns the path to the single
// canonical deploy script (scripts/deploy.sh) inside it. Pure/testable: side
// effects limited to stat. Returns an error if the dir or script is missing.
func deployScriptPath(deployDir string) (string, error) {
	if strings.TrimSpace(deployDir) == "" {
		return "", fmt.Errorf("deploy dir is empty")
	}
	if fi, err := os.Stat(deployDir); err != nil || !fi.IsDir() {
		return "", fmt.Errorf("deploy dir %s not found", deployDir)
	}
	script := filepath.Join(deployDir, "scripts", "deploy.sh")
	if _, err := os.Stat(script); err != nil {
		return "", fmt.Errorf("deploy script %s not found (is the deploy dir a swarmops checkout?)", script)
	}
	return script, nil
}

// runRedeploy is the manual fallback for the GitHub Actions deploy job. It does
// NOT build from the current working tree — it delegates to the one sanctioned
// primitive (scripts/deploy.sh), which resets the deploy dir to origin/main,
// rebuilds both binaries, restarts, health-checks, and rolls back on failure.
// This guarantees `swarmops redeploy` can never ship a stale worktree build.
// Pass --force to deploy over a dirty deploy tree. Launches the TUI on success.
func runRedeploy() {
	args := os.Args[2:] // pass-through flags (e.g. --force)
	dir := deployDirPath()
	script, err := deployScriptPath(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "redeploy: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("redeploy: deploying origin/main into %s via %s\n", dir, script)
	cmd := exec.Command("bash", append([]string{script}, args...)...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "redeploy: deploy failed: %v\n", err)
		os.Exit(1)
	}
	exe := filepath.Join(dir, "swarmops")
	if _, err := os.Stat(exe); err == nil {
		fmt.Printf("redeploy: launching TUI from %s\n", exe)
		if err := syscall.Exec(exe, []string{exe, "tui"}, os.Environ()); err != nil {
			fmt.Fprintf(os.Stderr, "redeploy: exec %s: %v\n", exe, err)
			os.Exit(1)
		}
	}
}

// runTUIClient starts the TUI as an HTTP client against the backend.
func runTUIClient() {
	baseURL := os.Getenv("SWARM_API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	api := newAPIClient(baseURL)

	// Health check before launching TUI
	if err := api.healthCheck(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Start the backend with: systemctl --user start swarmops\n")
		os.Exit(1)
	}

	// Redirect logs so TUI alt-screen is clean
	f, err := os.OpenFile("swarmops.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(f)
		defer f.Close()
	}

	if err := runTUI(api); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}

// runTUIRestart kills all running swarmops tui processes and starts a fresh one.
// Uses exec to replace the current process so the new TUI gets a clean terminal.
func runTUIRestart() {
	// Kill existing TUI processes (but not ourselves)
	exec.Command("pkill", "-f", "swarmops tui").Run()
	time.Sleep(500 * time.Millisecond)

	// Find our own binary
	exe, err := os.Executable()
	if err != nil {
		exe = "swarmops"
	}

	fmt.Printf("Starting fresh TUI...\n")
	syscall.Exec(exe, []string{exe, "tui"}, os.Environ())
	// syscall.Exec replaces the current process; we never return here
}

// runMCPStdioMode runs the MCP server over stdin/stdout (for Claude Code stdio transport).
func runMCPStdioMode() {
	database = initDatabase()
	defer database.Close()

	globalConfigService = newConfigService(database)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initPool(ctx)

	enablePoolTools := os.Getenv("SWARMOPS_MCP_POOL_TOOLS") == "true"
	svc := &Services{db: database, pool: globalPool, config: globalConfigService}
	server := NewMCPServer(svc, enablePoolTools)

	log.Printf("SwarmOps MCP stdio server starting")
	runMCPStdio(server)
}

// globalServices is the shared service layer, initialized in server mode.
var globalServices *Services

// globalMCPServer is the MCP server instance, initialized in server mode.
var globalMCPServer *MCPServer

func newHTTPServer() *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","version":"2.0"}`))
	})

	mux.HandleFunc("/api/", handleAPI)

	// Quota proxy: transparently forward /api/quota → quota-proxy on quotaProxyPort
	mux.HandleFunc("/api/quota", func(w http.ResponseWriter, r *http.Request) {
		proxyURL := fmt.Sprintf("http://localhost:%d/quota", quotaProxyPort)
		req, _ := http.NewRequestWithContext(r.Context(), r.Method, proxyURL, r.Body)
		req.Header = r.Header.Clone()
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp == nil {
			http.Error(w, `{"error":"quota-proxy unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()
		for k, v := range resp.Header {
			if k == "Content-Type" || k == "Content-Length" {
				w.Header().Set(k, v[0])
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	// MCP server endpoint
	enablePoolTools := os.Getenv("SWARMOPS_MCP_POOL_TOOLS") == "true"
	globalServices = &Services{db: database, pool: globalPool, config: globalConfigService}
	globalMCPServer = NewMCPServer(globalServices, enablePoolTools)
	mux.HandleFunc("/mcp", handleMCPHTTP(globalMCPServer))

	// OpenAI-compatible pool API
	mux.HandleFunc("/v1/chat/completions", handlePoolChatCompletions)
	mux.HandleFunc("/v1/models", handlePoolListModels)
	mux.HandleFunc("/api/swarm/pool", handlePoolStatusAPI)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &http.Server{
		Addr:    ":" + port, // Binds 0.0.0.0 — pool API used by Hermes over Tailscale
		Handler: mux,
	}
}

func isTerminal() bool {
	for _, f := range []*os.File{os.Stdin, os.Stdout} {
		fi, err := f.Stat()
		if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
			return false
		}
	}
	return true
}
