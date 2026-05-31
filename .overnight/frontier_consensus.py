#!/usr/bin/env python3
"""
Frontier consensus on all ingested training_queue_items via tkn-aipeer HTTP MCP.

For each item:
  - Fetch the question + payload (model output + tool calls + audit verdict)
  - Build a triage prompt
  - Call tkn-aipeer.peer_consult (3 frontier models: Opus 4.6, Gemini 3.1 Pro, GPT-5.5)
    asking each to grade GOLD/SILVER/REJECT
  - Persist results to JSONL (resumable)

Outputs:
  ~/swarmops/.overnight/frontier_consensus_results.jsonl
  ~/swarmops/.overnight/frontier_consensus_summary.json
"""
from __future__ import annotations
import json
import os
import re
import sys
import time
import urllib.request
from pathlib import Path

OUT_DIR = Path.home() / "swarmops" / ".overnight"
OUT_DIR.mkdir(exist_ok=True)
RESULTS_PATH = OUT_DIR / "frontier_consensus_results.jsonl"
SUMMARY_PATH = OUT_DIR / "frontier_consensus_summary.json"
LOG_PATH = OUT_DIR / "frontier_consensus.log"

SUPABASE_URL = os.environ.get("SUPABASE_URL", "https://auth.briefbox.work")
SERVICE_ROLE = os.environ.get("SUPABASE_SERVICE_ROLE_KEY")
MCP_URL = os.environ.get("AIPEER_MCP_URL", "https://mcp-aipeer.gate-hexatonic.ts.net/mcp")

if not SERVICE_ROLE:
    sys.exit("ERROR: SUPABASE_SERVICE_ROLE_KEY required")


def log(msg: str) -> None:
    line = f"[{time.strftime('%H:%M:%S')}] {msg}"
    print(line, flush=True)
    with open(LOG_PATH, "a") as f:
        f.write(line + "\n")


UA = "Mozilla/5.0 frontier-consensus/1.0"


def fetch_queue_items() -> list[dict]:
    req = urllib.request.Request(
        f"{SUPABASE_URL}/rest/v1/training_queue_items?status=eq.active&select=*&order=id.asc",
        headers={
            "apikey": SERVICE_ROLE,
            "Authorization": f"Bearer {SERVICE_ROLE}",
            "Accept": "application/json",
            "User-Agent": UA,  # Cloudflare bot-filter blocks Python-urllib default
        },
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        return json.loads(resp.read())


def mcp_call(tool: str, arguments: dict, call_id: int = 1) -> dict:
    """Call an MCP tool via HTTP. Server uses streamable-HTTP / SSE."""
    body = {
        "jsonrpc": "2.0",
        "id": call_id,
        "method": "tools/call",
        "params": {"name": tool, "arguments": arguments},
    }
    req = urllib.request.Request(
        MCP_URL,
        method="POST",
        headers={
            "Content-Type": "application/json",
            "Accept": "application/json, text/event-stream",
            "User-Agent": UA,
        },
        data=json.dumps(body).encode(),
    )
    try:
        with urllib.request.urlopen(req, timeout=180) as resp:
            raw = resp.read().decode()
        # Streamable-HTTP returns SSE lines: "event: message\ndata: {…}\n"
        for line in raw.splitlines():
            if line.startswith("data: "):
                obj = json.loads(line[6:])
                if "result" in obj:
                    return obj["result"]
                if "error" in obj:
                    return {"_mcp_error": obj["error"]}
        return {"_mcp_error": "no data line"}
    except Exception as e:  # noqa: BLE001
        return {"_mcp_error": f"{type(e).__name__}: {e}"}


def build_prompt(item: dict) -> str:
    payload = item.get("payload", {}) or {}
    question = item.get("question", "")
    output = (payload.get("output") or "")[:8000]  # cap
    tool_calls = payload.get("tool_calls") or []
    audit = payload.get("audit") or {}
    return f"""You are a senior UK barrister reviewing an AI-generated legal-research output for inclusion in a fine-tuning training dataset.

Grade this single sample on the GOLD/SILVER/REJECT scale:
- **GOLD** = would lift directly into a brief; legally correct, well-cited, useful, properly caveated
- **SILVER** = close but needs editing before lifting (citation issues, tone off, completeness gap)
- **REJECT** = wrong / fabricated / misapplied / unsafe — should not appear in training data

Reply with EXACTLY this JSON on the FIRST line of your response:
{{"decision": "gold"|"silver"|"reject", "primary_reason": "<short phrase>", "confidence": "strong"|"mild"|"unsure"}}

Then 1-3 sentences of justification. No preamble before the JSON.

---
QUESTION:
{question}

---
MODEL OUTPUT:
{output}

---
TOOL CALLS MADE: {len(tool_calls)}
AUDITOR VERDICT (informational): {audit.get('bucket_strict', 'UNKNOWN')} / score {audit.get('scores', {}).get('total', 0)}"""


JSON_RE = re.compile(r'\{[^{}]*"decision"\s*:\s*"(gold|silver|reject)"[^{}]*\}', re.IGNORECASE)


def parse_decision(text: str) -> dict:
    if not text:
        return {"decision": "empty", "primary_reason": "", "confidence": "unsure"}
    # Try first-line strict
    for line in text.strip().splitlines()[:5]:
        s = line.strip().lstrip("`json").lstrip()
        if s.startswith("{"):
            try:
                d = json.loads(s)
                if "decision" in d:
                    d["decision"] = d["decision"].lower()
                    return d
            except json.JSONDecodeError:
                pass
    # Fallback: regex over whole text
    m = JSON_RE.search(text)
    if m:
        try:
            d = json.loads(m.group(0))
            d["decision"] = d.get("decision", "unparsed").lower()
            return d
        except json.JSONDecodeError:
            pass
    return {"decision": "unparsed", "primary_reason": text[:120], "confidence": "unsure"}


def main() -> int:
    log("=== frontier consensus job starting ===")
    items = fetch_queue_items()
    log(f"fetched {len(items)} active queue items")

    seen_ids: set[int] = set()
    if RESULTS_PATH.exists():
        with open(RESULTS_PATH) as f:
            for line in f:
                try:
                    seen_ids.add(json.loads(line).get("queue_item_id"))
                except json.JSONDecodeError:
                    continue
        log(f"resume: skipping {len(seen_ids)} already-processed items")

    matrix_counts: dict[str, int] = {}
    consensus_gold = majority_gold = 0
    n_processed = 0

    for i, item in enumerate(items):
        if item["id"] in seen_ids:
            continue
        n_processed += 1
        log(f"[{i+1}/{len(items)}] item {item['id']} stratum={item.get('stratum', '?')} — consulting")
        prompt = build_prompt(item)
        # Explicit model list — overrides the consultation_type auto-routing.
        # Default routing picked Claude Sonnet 4.6 as the third slot and
        # Sonnet only followed the JSON-first-line format on 9/44 items
        # (it interpreted the format instruction as code-reference metadata
        # rather than the actual task). Opus 4.6 follows it cleanly.
        result = mcp_call("peer_consult", {
            "question": "Grade this single legal-AI sample as GOLD/SILVER/REJECT for fine-tuning training inclusion. Reply with JSON on the first line as specified.",
            "code_snippet": prompt,
            "models": ["github_copilot/gpt-5.5", "gemini-3.1-pro-preview", "claude-opus-4-6"],
            "use_llm_routing": False,
            "synthesize": False,
            "reason": f"Frontier consensus on training_queue_items.id={item['id']} for Charlotte's gate-2 pre-calibration",
        })

        if result.get("_mcp_error"):
            log(f"  ERROR: {result['_mcp_error']}")
            with open(RESULTS_PATH, "a") as f:
                f.write(json.dumps({
                    "queue_item_id": item["id"],
                    "stratum": item.get("stratum"),
                    "audit_bucket": item.get("payload", {}).get("audit", {}).get("bucket_strict"),
                    "error": result["_mcp_error"],
                }) + "\n")
            continue

        # peer_consult returns content[0]["text"] containing JSON of results
        try:
            content_str = result["content"][0]["text"]
            parsed = json.loads(content_str)
            model_results = parsed.get("results", [])
        except (KeyError, IndexError, json.JSONDecodeError) as e:
            log(f"  PARSE ERROR: {e}")
            with open(RESULTS_PATH, "a") as f:
                f.write(json.dumps({
                    "queue_item_id": item["id"],
                    "error": f"parse: {e}",
                    "raw": str(result)[:500],
                }) + "\n")
            continue

        per_model = {}
        decisions: list[str] = []
        for r in model_results:
            name = r.get("model_name", "?")
            content = r.get("content", "")
            d = parse_decision(content)
            per_model[name] = d
            decisions.append(d.get("decision", "unknown"))

        gold_count = decisions.count("gold")
        consensus = ("consensus_gold" if gold_count == 3
                     else "majority_gold" if gold_count >= 2
                     else "non_gold")
        if gold_count == 3:
            consensus_gold += 1
        if gold_count >= 2:
            majority_gold += 1

        audit_bucket = (item.get("payload") or {}).get("audit", {}).get("bucket_strict", "?")
        for d in decisions:
            key = f"audit={audit_bucket}__frontier={d}"
            matrix_counts[key] = matrix_counts.get(key, 0) + 1

        row = {
            "queue_item_id": item["id"],
            "stratum": item.get("stratum"),
            "question_excerpt": item.get("question", "")[:100],
            "audit_bucket": audit_bucket,
            "audit_score": (item.get("payload") or {}).get("audit", {}).get("scores", {}).get("total", 0),
            "per_model": per_model,
            "consensus": consensus,
            "decisions": decisions,
        }
        with open(RESULTS_PATH, "a") as f:
            f.write(json.dumps(row) + "\n")
        log(f"  → {decisions} = {consensus}")

    log("")
    log(f"=== done ===")
    log(f"processed: {n_processed}")
    log(f"consensus gold (all 3): {consensus_gold}")
    log(f"majority gold (≥2):     {majority_gold}")

    summary = {
        "n_items": len(items),
        "n_processed_this_run": n_processed,
        "n_consensus_gold": consensus_gold,
        "n_majority_gold": majority_gold,
        "matrix_counts": matrix_counts,
        "generated_at": time.strftime("%Y-%m-%dT%H:%M:%S"),
    }
    SUMMARY_PATH.write_text(json.dumps(summary, indent=2))
    log(f"summary → {SUMMARY_PATH}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
