package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// swarmClient is the interface the TUI uses to talk to the SwarmOps backend.
// The concrete implementation is apiClient (HTTP). Tests use fakeSwarmClient.
type swarmClient interface {
	Spawn(ctx context.Context, name, dir string, mission *string, model, profile string, envOverrides map[string]string) (*Session, error)
	listSessions() ([]Session, error)
	deleteSession(id string) error
	renameSession(id, name string) error
	poolStatus() (map[string]interface{}, error)
	getConfig(key string) (string, error)
	setMission(id, mission string) error
	updateSessionProfile(id, profile string) error
	updateSessionDirectory(id, directory string) error
	listAuditEvents(limit int) ([]ManagedSessionEvent, error)
	healthCheck() error
	quota() (*QuotaData, error)
	// Smart session creation: brain-routed spawn
	smartSpawn(goal, repoSlug string, dryRun bool) (*smartSpawnResult, error)
	// TUI ↔ API state sharing for agentic integration and testing
	pollTUIKey() (key string, hasKey bool) // non-blocking GET /api/tui/key
	pushTUIState(rendered string)          // async POST /api/tui/state
}

// apiClient is an HTTP client for the SwarmOps backend API.
// Used by the TUI in client mode instead of direct in-process calls.
type apiClient struct {
	baseURL string
	http    *http.Client
}

func newAPIClient(baseURL string) *apiClient {
	return &apiClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Spawn implements the Spawner interface via the HTTP API.
func (c *apiClient) Spawn(ctx context.Context, name, dir string, mission *string, model, profile string, envOverrides map[string]string) (*Session, error) {
	body := map[string]interface{}{
		"name":      name,
		"directory": dir,
	}
	if mission != nil {
		body["mission"] = *mission
	}
	if model != "" {
		body["model"] = model
	}
	if profile != "" {
		body["profile"] = profile
	}
	if len(envOverrides) > 0 {
		body["env_overrides"] = envOverrides
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/swarm/sessions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API %d: %s", resp.StatusCode, string(respBody))
	}

	var s Session
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &s, nil
}

func (c *apiClient) listSessions() ([]Session, error) {
	resp, err := c.http.Get(c.baseURL + "/api/swarm/sessions")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sessions []Session
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (c *apiClient) deleteSession(id string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/swarm/sessions/"+id, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("API %d", resp.StatusCode)
	}
	return nil
}

func (c *apiClient) renameSession(id, name string) error {
	data, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest("PATCH", c.baseURL+"/api/swarm/sessions/"+id, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("API %d", resp.StatusCode)
	}
	return nil
}

// poolStatus returns the pool status as a map, handling JSON number type
// conversion (JSON numbers unmarshal as float64, not int64).
func (c *apiClient) poolStatus() (map[string]interface{}, error) {
	resp, err := c.http.Get(c.baseURL + "/api/swarm/pool")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}
	return status, nil
}

func (c *apiClient) getConfig(key string) (string, error) {
	resp, err := c.http.Get(c.baseURL + "/api/swarm/config/" + key)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var entry struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return "", err
	}
	return entry.Value, nil
}

func (c *apiClient) setMission(id, mission string) error {
	data, _ := json.Marshal(map[string]interface{}{"mission": mission})
	req, err := http.NewRequest("PATCH", c.baseURL+"/api/swarm/sessions/"+id, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("API %d", resp.StatusCode)
	}
	return nil
}

func (c *apiClient) updateSessionProfile(id, profile string) error {
	data, _ := json.Marshal(map[string]string{"profile": profile})
	req, err := http.NewRequest("PATCH", c.baseURL+"/api/swarm/sessions/"+id, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("API %d", resp.StatusCode)
	}
	return nil
}

func (c *apiClient) updateSessionDirectory(id, directory string) error {
	data, _ := json.Marshal(map[string]string{"directory": directory})
	req, err := http.NewRequest("PATCH", c.baseURL+"/api/swarm/sessions/"+id, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("API %d", resp.StatusCode)
	}
	return nil
}

func (c *apiClient) listAuditEvents(limit int) ([]ManagedSessionEvent, error) {
	resp, err := c.http.Get(fmt.Sprintf("%s/api/swarm/audit?limit=%d", c.baseURL, limit))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var events []ManagedSessionEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}
	return events, nil
}

// healthCheck verifies the backend is reachable.
func (c *apiClient) healthCheck() error {
	resp, err := c.http.Get(c.baseURL + "/")
	if err != nil {
		return fmt.Errorf("cannot reach SwarmOps backend at %s: %w", c.baseURL, err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("backend returned %d", resp.StatusCode)
	}
	return nil
}

// QuotaData represents the usage quota data from the Anthropic rate-limit headers.
type QuotaData struct {
	CapturedAt    int64       `json:"captured_at"`
	OverallStatus string      `json:"overall_status"`
	BindingWindow string      `json:"binding_window"`
	Session5h     *WindowData `json:"session_5h,omitempty"`
	Weekly7d      *WindowData `json:"weekly_7d,omitempty"`
}

// WindowData holds per-window utilization data.
type WindowData struct {
	Utilization float64 `json:"utilization"`
	PercentLeft float64 `json:"percent_left"`
	Status      string  `json:"status"`
	ResetEpoch  int64   `json:"reset_epoch"`
}

// quota fetches the current usage quota from the backend.
func (c *apiClient) quota() (*QuotaData, error) {
	resp, err := c.http.Get(c.baseURL + "/api/quota")
	if err != nil {
		return nil, fmt.Errorf("quota fetch failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("quota endpoint returned %d", resp.StatusCode)
	}
	var q QuotaData
	if err := json.NewDecoder(resp.Body).Decode(&q); err != nil {
		return nil, fmt.Errorf("quota parse failed: %w", err)
	}
	return &q, nil
}

// smartSpawn calls POST /api/swarm/smart-spawn to route a goal to a repo and
// optionally spawn a session. With dryRun=true it returns the brain's pick
// without spawning; with dryRun=false it also creates the session.
func (c *apiClient) smartSpawn(goal, repoSlug string, dryRun bool) (*smartSpawnResult, error) {
	body, err := json.Marshal(map[string]interface{}{
		"goal":      goal,
		"repo_slug": repoSlug,
		"dry_run":   dryRun,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	// Brain calls can take a few seconds; use a longer timeout than the default 10s.
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", c.baseURL+"/api/swarm/smart-spawn", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API %d: %s", resp.StatusCode, string(respBody))
	}

	var result smartSpawnResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &result, nil
}

// pollTUIKey non-blocking-polls for a pending key from the API server.
// Returns ("", false) if no key is queued.
func (c *apiClient) pollTUIKey() (key string, hasKey bool) {
	type keyResp struct{ Key string }
	type emptyResp struct{}

	req, err := http.NewRequest("GET", c.baseURL+"/api/tui/key", nil)
	if err != nil {
		return "", false
	}
	req.Header.Set("Content-Type", "application/json")

	// Short timeout so polling doesn't stall the Update loop
	client := &http.Client{Timeout: 100 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode == 503 {
		// TUI not connected yet
		return "", false
	}
	if resp.StatusCode >= 400 {
		return "", false
	}
	var kr keyResp
	if err := json.NewDecoder(resp.Body).Decode(&kr); err != nil {
		// Empty body is fine (no key queued)
		return "", false
	}
	if kr.Key == "" {
		return "", false
	}
	return kr.Key, true
}

// pushTUIState POSTs the rendered TUI output to the API server for exposure
// via GET /api/tui/view and the swop_tui_state MCP tool. Runs asynchronously
// so it never blocks the render cycle.
func (c *apiClient) pushTUIState(rendered string) {
	body, _ := json.Marshal(map[string]interface{}{
		"rendered":  rendered,
		"timestamp": time.Now().Unix(),
	})
	go func() {
		req, err := http.NewRequest("POST", c.baseURL+"/api/tui/state", bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return
		}
		resp.Body.Close()
	}()
}
