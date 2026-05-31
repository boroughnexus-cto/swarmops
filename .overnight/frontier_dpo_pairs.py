#!/usr/bin/env python3
"""
Frontier-corrected DPO pairs generator.

For each ingested queue item where the frontier consensus said anything
worse than "majority gold", build a DPO pair:
  - rejected = Minimax M2.7 output (the existing one in training_queue_items.payload)
  - chosen   = Opus 4.6 "what Charlotte would lift" version, critiqued by
               GPT-5.5, revised by Opus, then 3-model peer-graded.

Full chain saved per item — gives Charlotte (or future analyst) the full
audit trail of why this particular "chosen" was selected.

Outputs:
  ~/swarmops/.overnight/dpo_pairs_results.jsonl   (resumable; append-mode)
  ~/swarmops/.overnight/dpo_pairs_summary.json    (final aggregate)
"""
from __future__ import annotations
import json
import os
import sys
import time
import urllib.request
from pathlib import Path

OUT_DIR = Path.home() / "swarmops" / ".overnight"
OUT_DIR.mkdir(exist_ok=True)
RESULTS_PATH = OUT_DIR / "dpo_pairs_results.jsonl"
SUMMARY_PATH = OUT_DIR / "dpo_pairs_summary.json"
LOG_PATH = OUT_DIR / "dpo_pairs.log"

SUPABASE_URL = os.environ.get("SUPABASE_URL", "https://auth.briefbox.work")
SERVICE_ROLE = os.environ.get("SUPABASE_SERVICE_ROLE_KEY")
MCP_URL = os.environ.get("AIPEER_MCP_URL", "https://mcp-aipeer.gate-hexatonic.ts.net/mcp")
UA = "Mozilla/5.0 frontier-dpo/1.0"

if not SERVICE_ROLE:
    sys.exit("ERROR: SUPABASE_SERVICE_ROLE_KEY required")

FRONTIER_RESULTS = OUT_DIR / "frontier_consensus_results.jsonl"
if not FRONTIER_RESULTS.exists():
    sys.exit(f"ERROR: {FRONTIER_RESULTS} not found — run frontier_consensus.py first")


def log(msg: str) -> None:
    line = f"[{time.strftime('%H:%M:%S')}] {msg}"
    print(line, flush=True)
    with open(LOG_PATH, "a") as f:
        f.write(line + "\n")


def fetch_queue_items() -> dict[int, dict]:
    req = urllib.request.Request(
        f"{SUPABASE_URL}/rest/v1/training_queue_items?status=eq.active&select=*",
        headers={
            "apikey": SERVICE_ROLE,
            "Authorization": f"Bearer {SERVICE_ROLE}",
            "Accept": "application/json",
            "User-Agent": UA,
        },
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        items = json.loads(resp.read())
    return {item["id"]: item for item in items}


def load_frontier_results() -> dict[int, dict]:
    out = {}
    with open(FRONTIER_RESULTS) as f:
        for line in f:
            try:
                r = json.loads(line)
                if "queue_item_id" in r:
                    out[r["queue_item_id"]] = r
            except json.JSONDecodeError:
                continue
    return out


def mcp_call(tool: str, arguments: dict, call_id: int = 1) -> dict:
    body = {"jsonrpc": "2.0", "id": call_id, "method": "tools/call",
            "params": {"name": tool, "arguments": arguments}}
    req = urllib.request.Request(
        MCP_URL, method="POST", data=json.dumps(body).encode(),
        headers={"Content-Type": "application/json",
                 "Accept": "application/json, text/event-stream",
                 "User-Agent": UA},
    )
    try:
        with urllib.request.urlopen(req, timeout=300) as resp:
            raw = resp.read().decode()
        for line in raw.splitlines():
            if line.startswith("data: "):
                obj = json.loads(line[6:])
                if "result" in obj: return obj["result"]
                if "error" in obj: return {"_mcp_error": obj["error"]}
        return {"_mcp_error": "no data line"}
    except Exception as e:  # noqa: BLE001
        return {"_mcp_error": f"{type(e).__name__}: {e}"}


def extract_text(result: dict) -> str:
    """Pull the text content from a quick_consult or peer_consult response."""
    if result.get("_mcp_error"):
        return ""
    try:
        text = result["content"][0]["text"]
        # quick_consult returns a JSON dict; peer_consult returns a JSON dict with results[]
        parsed = json.loads(text)
        # quick_consult shape: result.content
        if "result" in parsed and isinstance(parsed["result"], dict):
            return parsed["result"].get("content", "") or ""
        # peer_consult shape: results: list
        if "results" in parsed:
            # concatenate all model contents
            return "\n\n--- next model ---\n\n".join(
                r.get("content", "") for r in parsed.get("results", [])
            )
        return text
    except (KeyError, IndexError, json.JSONDecodeError):
        return ""


def opus_correct(question: str, retrieved_context: list, tool_calls: list, minimax_output: str,
                 frontier_complaints: list[str]) -> str:
    """Opus drafts a "what Charlotte would lift" version."""
    ctx_summary = "\n".join(
        f"- {c.get('doc_id', '?')}: {c.get('text', '')[:200]}"
        for c in (retrieved_context or [])[:5]
    )
    tools_summary = ", ".join(set(tc.get("name", "?") for tc in (tool_calls or [])))
    complaints = "\n".join(f"- {c[:200]}" for c in frontier_complaints[:5])
    prompt = f"""You are a senior UK commercial-litigation barrister. The Briefbox AI model produced the output below for a barrister's research query. Frontier reviewers flagged issues with it.

Your job: rewrite the answer as YOU would write it for a junior to lift into a brief. Be concise, properly caveated, legally accurate. Use the same authorities the model had available (listed in retrieval context). If a relevant authority is missing from the corpus, explicitly say so rather than inventing — partial-but-honest is the goal, not confident-but-fabricated.

Reply with ONLY the corrected answer in barrister voice (markdown OK). No preamble, no meta-commentary about what was wrong.

---
QUESTION:
{question}

---
RETRIEVED CONTEXT (model had this available):
{ctx_summary}

TOOLS THE MODEL CALLED: {tools_summary}

---
MODEL'S ORIGINAL OUTPUT:
{minimax_output[:6000]}

---
FRONTIER REVIEWER COMPLAINTS (concrete issues to address):
{complaints}
"""
    result = mcp_call("quick_consult", {
        "question": prompt,
        "model": "claude-opus-4-6",
        "consultation_type": "general",
        "reason": "Generate Charlotte-style corrected answer for DPO pair (chosen half)",
    })
    return extract_text(result).strip()


def gpt55_critique(question: str, opus_draft: str) -> str:
    """GPT-5.5 critiques the Opus draft — concrete edit list, not score."""
    prompt = f"""You are a senior UK barrister reviewing a junior's draft. Identify the 3-5 most important concrete edits needed before this answer could be lifted into a brief. Focus on legal correctness, completeness, and tone. Be specific (line/paragraph, what to change to what).

Reply with ONLY the edit list as a numbered set of concrete edits. No preamble, no praise, no overall verdict.

QUESTION:
{question}

DRAFT:
{opus_draft[:6000]}
"""
    result = mcp_call("quick_consult", {
        "question": prompt,
        "model": "github_copilot/gpt-5.5",
        "consultation_type": "general",
        "reason": "GPT-5.5 critique pass on Opus draft for DPO pair",
    })
    return extract_text(result).strip()


def opus_revise(question: str, draft: str, critique: str) -> str:
    """Opus revises based on GPT-5.5 critique."""
    prompt = f"""You wrote the draft below in response to the question. A senior barrister has produced a concrete edit list. Apply ALL the edits and return the revised version.

Reply with ONLY the revised answer in barrister voice. No preamble.

QUESTION:
{question}

YOUR DRAFT:
{draft[:6000]}

EDITS REQUIRED:
{critique[:3000]}
"""
    result = mcp_call("quick_consult", {
        "question": prompt,
        "model": "claude-opus-4-6",
        "consultation_type": "general",
        "reason": "Opus revision after GPT-5.5 critique for DPO pair",
    })
    return extract_text(result).strip()


def peer_grade(question: str, minimax_output: str, opus_revised: str) -> dict:
    """3-model frontier consensus on the pair."""
    prompt = f"""Two candidate answers to the same legal question. Pick which you would lift into a brief (A or B), or whether neither is acceptable.

Reply on the FIRST line with JSON exactly:
{{"preferred": "a"|"b"|"both_bad"|"both_good", "confidence": "strong"|"mild", "primary_reason": "<short phrase>"}}

Then 1-2 sentences of justification.

QUESTION:
{question}

ANSWER A (Minimax M2.7 output):
{minimax_output[:5000]}

ANSWER B (Opus draft, GPT-5.5-critiqued, Opus-revised):
{opus_revised[:5000]}
"""
    result = mcp_call("peer_consult", {
        "question": "Pick the answer you'd lift into a brief; JSON-first-line as specified.",
        "code_snippet": prompt,
        "models": ["github_copilot/gpt-5.5", "gemini-3.1-pro-preview", "claude-opus-4-6"],
        "use_llm_routing": False,
        "synthesize": False,
        "reason": "3-model grade of Minimax-vs-Opus DPO pair",
    })
    return result


def extract_complaints(frontier_record: dict) -> list[str]:
    """Pull each model's primary_reason into a list of concrete complaints."""
    out = []
    for name, d in (frontier_record.get("per_model") or {}).items():
        if d.get("decision") in ("silver", "reject"):
            reason = d.get("primary_reason", "")
            if reason and "unparsed" not in str(reason).lower():
                out.append(f"{name}: {reason}")
    return out


def main() -> int:
    log("=== frontier DPO-pairs job starting ===")
    items = fetch_queue_items()
    frontier = load_frontier_results()
    log(f"queue items: {len(items)} ; frontier records: {len(frontier)}")

    # Targets: items where ≥1 frontier model said silver/reject (most of them)
    # AND where we have at least one parseable complaint to feed Opus.
    targets = []
    for qid, item in items.items():
        fr = frontier.get(qid)
        if not fr or "per_model" not in fr:
            continue
        decisions = fr.get("decisions", [])
        # Skip if all 3 already said gold (no correction needed)
        if decisions.count("gold") == 3:
            continue
        complaints = extract_complaints(fr)
        if not complaints:
            continue  # nothing to instruct Opus with
        targets.append((qid, item, fr, complaints))
    log(f"targets: {len(targets)} items need correction")

    # Resume support
    done_ids: set[int] = set()
    if RESULTS_PATH.exists():
        with open(RESULTS_PATH) as f:
            for line in f:
                try: done_ids.add(json.loads(line).get("queue_item_id"))
                except json.JSONDecodeError: pass
        log(f"resume: skipping {len(done_ids)} already-completed items")

    n_done = 0
    n_errors = 0
    for idx, (qid, item, fr, complaints) in enumerate(targets):
        if qid in done_ids:
            continue
        log(f"[{idx+1}/{len(targets)}] item {qid} ({item.get('stratum','?')})")
        payload = item.get("payload") or {}
        question = item.get("question", "")
        retrieval_context = payload.get("retrieval_context") or []
        tool_calls = payload.get("tool_calls") or []
        minimax_output = payload.get("output", "") or ""

        try:
            log(f"  step 1: opus draft")
            draft = opus_correct(question, retrieval_context, tool_calls, minimax_output, complaints)
            if not draft or len(draft) < 100:
                raise RuntimeError(f"opus draft too short: {len(draft)} chars")

            log(f"  step 2: gpt5.5 critique")
            critique = gpt55_critique(question, draft)
            if not critique:
                raise RuntimeError("empty critique")

            log(f"  step 3: opus revise")
            revised = opus_revise(question, draft, critique)
            if not revised or len(revised) < 100:
                raise RuntimeError(f"revised too short: {len(revised)} chars")

            log(f"  step 4: peer grade")
            grade_result = peer_grade(question, minimax_output, revised)
            grade_text = extract_text(grade_result) if not grade_result.get("_mcp_error") else f"MCP_ERR: {grade_result['_mcp_error']}"

            row = {
                "queue_item_id": qid,
                "stratum": item.get("stratum"),
                "question": question,
                "minimax_output": minimax_output,
                "opus_draft": draft,
                "gpt55_critique": critique,
                "opus_revised": revised,
                "peer_grade_raw": grade_text,
                "frontier_complaints_summary": complaints[:5],
                "audit_bucket_at_ingest": fr.get("audit_bucket"),
                "frontier_decisions_at_ingest": fr.get("decisions"),
                "generated_at": time.strftime("%Y-%m-%dT%H:%M:%S"),
            }
            with open(RESULTS_PATH, "a") as f:
                f.write(json.dumps(row) + "\n")
            n_done += 1
            log(f"  ok ({len(revised)} char revision)")
        except Exception as e:  # noqa: BLE001
            n_errors += 1
            log(f"  ERROR: {type(e).__name__}: {e}")
            with open(RESULTS_PATH, "a") as f:
                f.write(json.dumps({
                    "queue_item_id": qid,
                    "error": f"{type(e).__name__}: {e}",
                    "generated_at": time.strftime("%Y-%m-%dT%H:%M:%S"),
                }) + "\n")

    log("")
    log(f"=== done ===")
    log(f"processed: {n_done}/{len(targets) - len(done_ids)}")
    log(f"errors: {n_errors}")

    summary = {
        "n_targets": len(targets),
        "n_completed_this_run": n_done,
        "n_errors": n_errors,
        "generated_at": time.strftime("%Y-%m-%dT%H:%M:%S"),
    }
    SUMMARY_PATH.write_text(json.dumps(summary, indent=2))
    log(f"summary → {SUMMARY_PATH}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
