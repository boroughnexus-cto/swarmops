package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// BrainPick is the structured decision returned by the rc_newgoal dispatcher.
// `Pick` is either an "owner/name" slug from the known_repos catalogue or the
// sentinel "none" — the latter signals that no existing repo fits and the
// caller should propose creating a new one (or pick a closer-fitting candidate
// from `Suggestions`).
type BrainPick struct {
	Pick        string   `json:"pick"`                  // "owner/name" or "none"
	Confidence  string   `json:"confidence"`            // "high" | "medium" | "low"
	Reasoning   string   `json:"reasoning"`             // 1-3 sentence justification
	Suggestions []string `json:"suggestions,omitempty"` // up to 3 close-but-not-picked slugs
}

// brainDefaultModel is the model used when the caller omits brain_model in
// rc_newgoal. Haiku is fast (~1-2s end-to-end) and quite good enough to pick
// among ~30 well-described repos; callers can override per-call (sonnet for
// trickier picks, chatgptsub-gpt-5.5 for cross-checking, etc.).
const brainDefaultModel = "haiku"

// brainAsk sends a single non-streaming chat completion through the warm pool
// and returns the assistant text. Mirrors poolChatHandler's plumbing (Acquire
// → sendQuery → drain events → Release) without going through the HTTP layer.
//
// Returns an error if the pool is disabled, the model is unknown, the slot
// errors out, or the model itself returns an error result.
func brainAsk(ctx context.Context, svc *Services, modelName, systemPrompt, userPrompt string) (string, error) {
	if svc.pool == nil {
		return "", fmt.Errorf("pool is not enabled — set POOL_ENABLED=true and configure POOL_MODELS")
	}
	model, ok := resolveModel(modelName)
	if !ok {
		return "", fmt.Errorf("unknown model: %s", modelName)
	}

	sysJSON, _ := json.Marshal(systemPrompt)
	userJSON, _ := json.Marshal(userPrompt)
	messages := []oaiMessage{
		{Role: "system", Content: sysJSON},
		{Role: "user", Content: userJSON},
	}
	prompt := messagesToPrompt(messages)

	slot, err := svc.pool.Acquire(ctx, model)
	if err != nil {
		return "", fmt.Errorf("pool exhausted for model %s: %w", model, err)
	}
	defer svc.pool.Release(slot)

	if err := slot.sendQuery(prompt); err != nil {
		slot.mu.Lock()
		slot.state = slotDead
		slot.mu.Unlock()
		return "", fmt.Errorf("send query: %w", err)
	}

	var fullText strings.Builder
	var resultEv poolEvent
	for {
		ev, err := readEventWithCtx(ctx, slot)
		if err != nil {
			slot.mu.Lock()
			slot.state = slotDead
			slot.mu.Unlock()
			return "", fmt.Errorf("read event: %w", err)
		}
		switch ev.Type {
		case "assistant":
			fullText.WriteString(extractAssistantText(ev))
		case "stream_event":
			fullText.WriteString(extractStreamText(ev))
		case "rate_limit_event":
			svc.pool.handleRateLimit(slot, ev)
		case "result":
			resultEv = ev
			goto done
		}
	}
done:
	slot.mu.Lock()
	slot.errorCount = 0
	slot.totalCost += resultEv.CostUSD
	slot.totalRequests++
	slot.mu.Unlock()
	svc.pool.totalCost.Add(int64(resultEv.CostUSD * 1e6))

	if resultEv.IsError {
		return "", fmt.Errorf("model error: %s", resultEv.Result)
	}
	text := fullText.String()
	if text == "" {
		text = resultEv.Result
	}
	return text, nil
}

// brainSystemPrompt is the constant system message that frames every brain
// invocation. Defines the available repos, the output schema, and the hard
// rules: pick from the list or "none"; reply with JSON only.
const brainSystemPrompt = `You are SwarmOps' repo-routing brain. Given a user goal and a catalogue of repositories, pick the single repo whose scope best matches the goal so a Claude Code session can be spawned in the correct working directory.

Output STRICT JSON only, no prose before or after, matching this schema:

{
  "pick": "<owner/name from catalogue, or the literal string none>",
  "confidence": "<high|medium|low>",
  "reasoning": "<one short sentence>",
  "suggestions": ["<owner/name>", "..."]
}

Rules:
- "pick" MUST be either an exact "owner/name" slug from the catalogue or the literal string "none".
- Use "none" when no repo's scope is a credible match — DO NOT force a pick.
- Prefer repos that are already cloned (is_cloned: true) when the match is comparable.
- "suggestions" lists up to 3 runner-up slugs (omit or empty when you picked one with high confidence).
- "confidence" should be "high" only when the goal is unambiguously in the picked repo's scope.
- Reply with the JSON object alone — no markdown fences, no commentary.`

// brainUserPrompt renders the goal + repos catalogue into the user-turn body.
// Format is plain text + JSON catalogue to keep the prompt cheap; the model
// reads JSON well enough without a structured-tools setup.
func brainUserPrompt(goal string, repos []KnownRepo) string {
	type repoForBrain struct {
		Slug          string `json:"slug"`
		IsCloned      bool   `json:"is_cloned"`
		Description   string `json:"description,omitempty"`
		ClaudeSummary string `json:"claude_md,omitempty"`
	}
	catalogue := make([]repoForBrain, 0, len(repos))
	for _, r := range repos {
		catalogue = append(catalogue, repoForBrain{
			Slug:          r.Slug(),
			IsCloned:      r.IsCloned(),
			Description:   r.Description,
			ClaudeSummary: r.ClaudeMDSummary,
		})
	}
	body, _ := json.MarshalIndent(catalogue, "", "  ")
	var sb strings.Builder
	sb.WriteString("GOAL: ")
	sb.WriteString(strings.TrimSpace(goal))
	sb.WriteString("\n\nREPOS:\n")
	sb.Write(body)
	sb.WriteString("\n\nReturn JSON.")
	return sb.String()
}

// brainParse extracts a BrainPick from the model's raw text reply. Handles a
// few common formatting quirks: plain JSON, JSON wrapped in ```fences```, or
// JSON with leading/trailing prose. Returns an error if no parseable object
// is found.
func brainParse(raw string) (BrainPick, error) {
	var pick BrainPick
	// Strip optional ```json … ``` fence.
	fenced := regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)\\s*```").FindStringSubmatch(raw)
	if len(fenced) == 2 {
		raw = fenced[1]
	}
	// Find first { … last } substring — tolerates leading commentary.
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return pick, fmt.Errorf("no JSON object found in brain reply: %q", truncateForError(raw))
	}
	candidate := raw[start : end+1]
	if err := json.Unmarshal([]byte(candidate), &pick); err != nil {
		return pick, fmt.Errorf("parse brain JSON: %w (raw: %q)", err, truncateForError(candidate))
	}
	pick.Pick = strings.TrimSpace(pick.Pick)
	pick.Confidence = strings.ToLower(strings.TrimSpace(pick.Confidence))
	return pick, nil
}

func truncateForError(s string) string {
	const max = 200
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
