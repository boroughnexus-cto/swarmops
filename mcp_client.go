package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ─── MCP client for calling remote MCP servers from the TUI ─────────────────

// mcpToolCall makes a JSON-RPC tools/call to an MCP server via streamable-http.
// Returns the text content blocks from the tool result.
func mcpToolCall(serverURL, toolName string, args map[string]interface{}) ([]string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", serverURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// Parse SSE response: extract "data: " lines
	var jsonData []byte
	for _, line := range strings.Split(string(respBody), "\n") {
		if strings.HasPrefix(line, "data: ") {
			jsonData = []byte(strings.TrimPrefix(line, "data: "))
			break
		}
	}
	if jsonData == nil {
		// Try as plain JSON
		jsonData = respBody
	}

	var rpcResp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(jsonData, &rpcResp); err != nil {
		return nil, fmt.Errorf("parse MCP response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", rpcResp.Error.Message)
	}

	var texts []string
	for _, c := range rpcResp.Result.Content {
		if c.Type == "text" {
			texts = append(texts, c.Text)
		}
	}
	return texts, nil
}
