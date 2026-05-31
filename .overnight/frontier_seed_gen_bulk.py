#!/usr/bin/env python3
"""
Bulk parallel seed-bank expansion — target 6000 seeds.

Strategy:
- 6 parallel worker threads (urllib + threading; safe for HTTP)
- Each task generates 10 seeds in ONE Opus call (batched)
- Skips per-seed Gemini sanity check — spot-check later
- Writes to the SAME file as the wrapper's Phase B so resume is shared
- Resumable on restart (counts existing entries by category)

Output:
  ~/swarmops/.overnight/seed_bank_expansion.jsonl
  ~/swarmops/.overnight/seed_bank_bulk.log

Targets total = 6000 across 8 categories.
At ~10 seeds per batch + 6 workers + ~30s per batch:
  expected throughput ~120 seeds/min = ~7200 seeds/hour
  6000 seeds → ~50 min wall time (with overhead, 60-90 min realistic)
Cost estimate:
  600 batches × ~$0.30 per Opus batch ≈ $180
"""
from __future__ import annotations
import json
import os
import sys
import threading
import time
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

OUT_DIR = Path.home() / "swarmops" / ".overnight"
OUT_DIR.mkdir(exist_ok=True)
RESULTS_PATH = OUT_DIR / "seed_bank_expansion.jsonl"
LOG_PATH = OUT_DIR / "seed_bank_bulk.log"

MCP_URL = os.environ.get("AIPEER_MCP_URL", "https://mcp-aipeer.gate-hexatonic.ts.net/mcp")
UA = "Mozilla/5.0 frontier-seedgen-bulk/1.0"

# Target distribution (6000 total)
TARGETS = [
    ("partial_knowledge", 15000,
     "Question where the fixture has SOME relevant authority but not enough to answer fully. Model should cite what's available AND explicitly name authorities it could NOT find in the corpus (by case name + neutral citation). e.g. 'Apply Denton in Court of Appeal context' — fixture has Denton but no post-Denton CoA cases like Mitchell-reaffirmed-by-Bedford."),
    ("matter_file_grounded", 9000,
     "Realistic messy input bundle: solicitor email + procedural history extract + pleadings excerpt + opponent's argument + client objective + budget constraint. Output should be a barrister-voice advice note, draft argument section, or risk ranking — not generic legal Q&A."),
    ("adverse_authority_handling", 9000,
     "Question where the user wants help distinguishing a leading case that's against them, or framing how to address a known-bad authority. The right answer acknowledges the adverse authority and offers distinguishing facts or pragmatic acknowledgment."),
    ("multi_authority_synthesis", 6000,
     "Question where no single authority answers — the model must synthesise across 2-4 cases plus a CPR provision. e.g. 'Apply Case A's principle + Case B's limitation + Case C's analogous facts + CPR procedural consequence.'"),
    ("need_more_facts", 6000,
     "Question that cannot be answered without specific facts. Right answer: list of facts needed before advising, NOT a citation-abstention. Distinct from 'I cannot find authority' — this is 'I cannot advise without X, Y, Z'."),
    ("overconfident_to_caveated", 6000,
     "Input is a junior's overconfident draft paragraph (provided in the question body). Task: rewrite in proper barrister voice with appropriate caveats. Tests tone calibration."),
    ("impermissible_request", 4500,
     "User asks for something impermissible: 'make this stronger than it is', 'ignore the bad case', 'draft around the missed deadline', 'can we threaten committal?'. The right answer is a firm-but-useful refusal: acknowledges what's impermissible, pivots to what IS permissible."),
    ("no_good_answer", 4500,
     "Hard case where neither yes nor no is clearly right. Right answer articulates the tension and the strongest argument on each side without faking a conclusion."),
]

AREAS = [
    "procedure", "privilege", "contract", "commercial",
    "evidence", "trusts-equity", "competition", "property",
    "tort", "restitution", "employment", "public-admin", "ip",
]

N_WORKERS = 4  # Gemini takes more concurrency than Opus
BATCH_SIZE = 5  # Gemini batch sweet-spot ~20s for 5
MAX_RETRIES = 2

write_lock = threading.Lock()
log_lock = threading.Lock()


def log(msg: str) -> None:
    line = f"[{time.strftime('%H:%M:%S')}] {msg}"
    with log_lock:
        print(line, flush=True)
        with open(LOG_PATH, "a") as f:
            f.write(line + "\n")


def mcp_call(arguments: dict) -> dict:
    body = {"jsonrpc": "2.0", "id": 1, "method": "tools/call",
            "params": {"name": "quick_consult", "arguments": arguments}}
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
    if result.get("_mcp_error"):
        return ""
    try:
        text = result["content"][0]["text"]
        parsed = json.loads(text)
        if "result" in parsed and isinstance(parsed["result"], dict):
            return parsed["result"].get("content", "") or ""
        return text
    except (KeyError, IndexError, json.JSONDecodeError):
        return ""


def generate_batch(category: str, category_desc: str, area_hint: str, n: int) -> list[dict]:
    """Ask Opus to produce N seeds at once, in JSON array form."""
    prompt = f"""Generate {n} new, DIVERSE seed questions for a UK commercial-litigation barrister-AI training set. Reply with EXACTLY one JSON array of {n} objects on the first lines of your response (no markdown code fence, just the JSON). Each object has this shape:

{{"category":"{category}","area_of_law":"<one of: {', '.join(AREAS)}>","question":"<barrister-voice query, 2-4 sentences>","context":"<optional solicitor-email or matter-file bundle, omit if not needed>","expected_output_shape":"advice_note|argument_section|research_memo|refusal|risk_ranking|email","expected_cases":["<real UK case name>", ...],"expected_cpr":["<rule>", ...],"expected_gaps":["<real authority name + neutral citation that would be needed for complete answer but probably not in corpus>", ...]}}

CATEGORY: {category}
WHAT THIS CATEGORY TESTS: {category_desc}
AREA-OF-LAW HINT (vary across the {n} seeds, but ground in commercial-litigation adjacent areas): {area_hint}

CRITICAL CONSTRAINTS:
- Questions must be DIVERSE — vary the legal sub-area, fact patterns, output shape, and authority focus
- Cases in expected_cases MUST be real English/UK cases with correct neutral citations (don't invent)
- Cases in expected_gaps must be real (e.g. "Norwich Pharmacal Co v C&E [1974] AC 133", "SFO v ENRC [2018] EWCA Civ 2006")
- For matter_file_grounded: include realistic context (solicitor email tone, named parties, dates, sum-at-stake)
- For impermissible_request: ensure the impermissibility is clear legal-professional-ethics not just rudeness
- For overconfident_to_caveated: put the junior's overconfident draft in the "context" field as the input to rewrite

Reply with ONLY the JSON array. Each object on its own logical block. No markdown fence, no preamble, no commentary between objects.
"""
    # Gemini 3.1 Pro instead of Opus — Opus is rate-limited under sustained
    # parallel load (Anthropic API throttling). Gemini via GitHub Copilot is
    # a separate rate-limit budget and produces ~5-item batches in ~20s.
    # Cost: free tier through Copilot.
    result = mcp_call({
        "question": prompt,
        "model": "gemini-3.1-pro-preview",
        "consultation_type": "general",
        "reason": f"Bulk seed generation: {category} batch of {n}",
    })
    text = extract_text(result).strip()
    if not text:
        return []
    # Strip markdown fence if present
    if text.startswith("```"):
        lines = text.splitlines()
        try:
            start = next(i for i, l in enumerate(lines) if l.startswith("```")) + 1
            end = next(i for i, l in enumerate(lines[start:], start=start) if l.startswith("```"))
            text = "\n".join(lines[start:end])
        except StopIteration:
            pass
    # Try parsing as array
    try:
        arr = json.loads(text)
        if isinstance(arr, list):
            return [s for s in arr if isinstance(s, dict) and s.get("question")]
        # Sometimes Opus returns a single object instead of array
        if isinstance(arr, dict) and arr.get("question"):
            return [arr]
    except json.JSONDecodeError:
        pass
    # Fallback: regex out object literals
    import re
    objs = []
    for m in re.finditer(r'\{[^{}]*"question"[^{}]*\}', text, re.DOTALL):
        try:
            obj = json.loads(m.group(0))
            if isinstance(obj, dict) and obj.get("question"):
                objs.append(obj)
        except json.JSONDecodeError:
            continue
    return objs


def write_seeds(seeds: list[dict], category: str) -> int:
    """Append seeds to RESULTS_PATH under a write lock. Returns count written."""
    n = 0
    with write_lock:
        with open(RESULTS_PATH, "a") as f:
            for s in seeds:
                s["category"] = category
                s["accepted"] = True
                s["generated_at"] = time.strftime("%Y-%m-%dT%H:%M:%S")
                f.write(json.dumps(s) + "\n")
                n += 1
    return n


def count_existing() -> dict[str, int]:
    counts: dict[str, int] = {}
    if not RESULTS_PATH.exists():
        return counts
    with open(RESULTS_PATH) as f:
        for line in f:
            try:
                r = json.loads(line)
                if r.get("accepted") and r.get("category"):
                    counts[r["category"]] = counts.get(r["category"], 0) + 1
            except json.JSONDecodeError:
                continue
    return counts


def worker_task(task: tuple) -> tuple[str, int, str | None]:
    """One task = one batch of N seeds in one Opus call.
    Retries once with 20s backoff on empty results (rate-limit recovery).
    Returns (category, n_written, err_or_none).
    """
    category, desc, area, n = task
    for attempt in range(MAX_RETRIES + 1):
        seeds = generate_batch(category, desc, area, n)
        if seeds:
            n_written = write_seeds(seeds, category)
            if attempt > 0:
                log(f"  +{n_written} [{category}] (succeeded on retry #{attempt})")
            return (category, n_written, None)
        if attempt < MAX_RETRIES:
            time.sleep(20)  # rate-limit backoff
    return (category, 0, "empty_or_unparseable_after_retry")


def main() -> int:
    log(f"=== bulk seed generation starting (target=60000, workers={N_WORKERS}, batch={BATCH_SIZE}) ===")
    existing = count_existing()
    log(f"existing accepted seeds: {sum(existing.values())} | {existing}")

    # Build the task list — flat list of (category, desc, area, n) tuples
    tasks: list[tuple] = []
    for category, target, desc in TARGETS:
        already = existing.get(category, 0)
        need = target - already
        if need <= 0:
            continue
        # Number of batches needed (round up)
        n_batches = (need + BATCH_SIZE - 1) // BATCH_SIZE
        for i in range(n_batches):
            this_batch = min(BATCH_SIZE, need - i * BATCH_SIZE)
            # Round-robin area across the batches for this category
            area = AREAS[(already // BATCH_SIZE + i) % len(AREAS)]
            tasks.append((category, desc, area, this_batch))

    log(f"prepared {len(tasks)} batches across categories")

    n_completed = 0
    n_failed = 0
    total_seeds_written = 0
    t0 = time.time()

    with ThreadPoolExecutor(max_workers=N_WORKERS) as pool:
        futures = {pool.submit(worker_task, task): task for task in tasks}
        for fut in as_completed(futures):
            task = futures[fut]
            try:
                category, n_written, err = fut.result()
                if err:
                    n_failed += 1
                    log(f"  FAIL [{category}]: {err}")
                else:
                    n_completed += 1
                    total_seeds_written += n_written
                    elapsed = time.time() - t0
                    rate = total_seeds_written / max(elapsed, 1) * 60  # seeds/min
                    log(f"  +{n_written} [{category}] ; total={total_seeds_written} ; rate={rate:.0f}/min ; {n_completed}/{len(tasks)} batches")
            except Exception as e:  # noqa: BLE001
                n_failed += 1
                log(f"  EXCEPTION on {task[0]}: {type(e).__name__}: {e}")

    log("")
    log(f"=== done ===")
    log(f"total seeds written this run: {total_seeds_written}")
    log(f"batches completed: {n_completed}")
    log(f"batches failed:    {n_failed}")
    log(f"elapsed: {(time.time() - t0)/60:.1f} min")

    final = count_existing()
    log(f"final per-category: {final}")
    log(f"GRAND TOTAL: {sum(final.values())}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
