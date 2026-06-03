package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// quotaProxyPort is the localhost port the quota-proxy sidecar listens on.
// Sessions route ANTHROPIC_BASE_URL through it (see spawn.go) so we can capture
// Anthropic rate-limit headers for usage tracking. If this port has no listener,
// every session pointed at it fails with ConnectionRefused — hence the watchdog.
const quotaProxyPort = 8082

// superviseQuotaProxy launches the quota-proxy sidecar binary and keeps it alive
// for the lifetime of ctx. If the proxy exits or crashes, it is restarted with
// capped exponential backoff; a process that stayed healthy resets the backoff.
// Supervision runs in a background goroutine, so this returns immediately.
//
// When ctx is cancelled (server shutdown) the child is killed via CommandContext
// and the goroutine exits. Missing binary is logged loudly but non-fatal: the
// server still runs, only usage tracking and proxy-routed sessions are affected.
func superviseQuotaProxy(ctx context.Context, path string, port int) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("WARNING: quota-proxy binary not found at %s — usage tracking disabled and any session with ANTHROPIC_BASE_URL=http://localhost:%d will fail (ConnectionRefused)", path, port)
		return
	}

	go func() {
		const (
			minBackoff    = time.Second
			maxBackoff    = 30 * time.Second
			healthyUptime = time.Minute // ran at least this long => reset backoff
		)
		backoff := minBackoff

		for {
			if ctx.Err() != nil {
				return
			}

			cmd := exec.CommandContext(ctx, path, fmt.Sprintf("--port=%d", port))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = os.Environ()

			start := time.Now()
			if err := cmd.Start(); err != nil {
				log.Printf("quota-proxy: start failed: %v — retrying in %s", err, backoff)
				if !sleepCtx(ctx, backoff) {
					return
				}
				backoff = nextBackoff(backoff, maxBackoff)
				continue
			}

			if waitForProxyHealthy(ctx, port) {
				log.Printf("quota-proxy started on port %d (ANTHROPIC_BASE_URL=http://localhost:%d for sessions)", port, port)
			}

			err := cmd.Wait()
			if ctx.Err() != nil {
				// Shutdown: CommandContext already signalled the child.
				log.Printf("quota-proxy stopped")
				return
			}

			ran := time.Since(start)
			log.Printf("quota-proxy exited after %s (%v) — restarting in %s", ran.Round(time.Millisecond), err, backoff)
			if !sleepCtx(ctx, backoff) {
				return
			}
			if ran >= healthyUptime {
				backoff = minBackoff
			} else {
				backoff = nextBackoff(backoff, maxBackoff)
			}
		}
	}()
}

// waitForProxyHealthy polls the sidecar's /health endpoint until it answers or a
// short window elapses. Returns true once the proxy is accepting connections.
func waitForProxyHealthy(ctx context.Context, port int) bool {
	url := fmt.Sprintf("http://localhost:%d/health", port)
	for i := 0; i < 40; i++ {
		if ctx.Err() != nil {
			return false
		}
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// nextBackoff doubles cur, capped at max.
func nextBackoff(cur, max time.Duration) time.Duration {
	next := cur * 2
	if next > max {
		return max
	}
	return next
}

// sleepCtx sleeps for d, returning false if ctx is cancelled first.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
