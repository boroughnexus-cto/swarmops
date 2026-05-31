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
  ~/swarmops/.overnight/seed_bank_expansion_minimax.jsonl
  ~/swarmops/.overnight/seed_bank_bulk_minimax.log

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
RESULTS_PATH = OUT_DIR / "seed_bank_expansion_minimax.jsonl"
LOG_PATH = OUT_DIR / "seed_bank_bulk_minimax.log"

MINIMAX_URL = "https://macmini-hermes-simon.gate-hexatonic.ts.net:8645/v1/chat/completions"
UA = "Mozilla/5.0 frontier-seedgen-bulk-minimax/1.0"

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

N_WORKERS = 2  # macmini Minimax is a single instance — 2 workers max useful
BATCH_SIZE = 5
MAX_RETRIES = 2

write_lock = threading.Lock()
log_lock = threading.Lock()


def log(msg: str) -> None:
    line = f"[{time.strftime('%H:%M:%S')}] {msg}"
    with log_lock:
        print(line, flush=True)
        with open(LOG_PATH, "a") as f:
            f.write(line + "\n")


import re

_THINK_RE = re.compile(r"<think>.*?</think>\s*", re.DOTALL)


def minimax_call(prompt: str, max_tokens: int = 12000) -> dict:
    """Direct OpenAI-compatible call to Minimax M2.7 on macmini via hermes proxy.

    Note: Minimax M2.7 is a "thinking" model — it emits <think>...</think>
    blocks before the real answer. Output is much larger than non-thinking
    models; we use a high max_tokens (12K) to ensure the JSON array isn't
    truncated.
    """
    body = {
        "model": "minimax-m2.7",
        "messages": [{"role": "user", "content": prompt}],
        "max_tokens": max_tokens,
        "temperature": 0.7,
    }
    req = urllib.request.Request(
        MINIMAX_URL, method="POST", data=json.dumps(body).encode(),
        headers={"Content-Type": "application/json", "User-Agent": UA},
    )
    try:
        with urllib.request.urlopen(req, timeout=240) as resp:
            return {"ok": True, "data": json.loads(resp.read())}
    except Exception as e:  # noqa: BLE001
        return {"ok": False, "_err": f"{type(e).__name__}: {e}"}


def extract_text(result: dict) -> str:
    """Extract assistant text, stripping Minimax's <think>...</think> blocks."""
    if not result.get("ok"):
        return ""
    try:
        text = result["data"]["choices"][0]["message"]["content"] or ""
        # Strip thinking blocks — they confuse JSON parsing downstream.
        return _THINK_RE.sub("", text).strip()
    except (KeyError, IndexError, TypeError):
        return ""


def generate_batch(category: str, category_desc: str, area_hint: str, n: int) -> list[dict]:
    """Ask Minimax to produce N seeds at once, in JSON array form."""
    prompt = f"""You are generating training data for a UK barrister-AI. Generate exactly {n} REALISTIC seed questions in JSON array form.

CATEGORY: {category}
WHAT THIS CATEGORY TESTS: {category_desc}
AREA-OF-LAW: {area_hint}

CRITICAL: Fill in real legal content. Do NOT output the angle-bracket placeholders literally. Use real UK case names and neutral citations.

EXAMPLE OUTPUT (this is the FORMAT you must follow — but vary the content):
[
  {{"category":"{category}","area_of_law":"procedure","question":"Our client missed the disclosure deadline by 2 weeks because of a partner's illness. Opponent wants debarring. Advise on relief from sanctions and the likelihood of success at stage 2 of Denton.","expected_output_shape":"advice_note","expected_cases":["Denton v TH White Ltd [2014] EWCA Civ 906","Mitchell v News Group [2013] EWCA Civ 1537"],"expected_cpr":["3.9"],"expected_gaps":["Hallam Estates Ltd v Baker [2014] EWCA Civ 661 on post-Denton illness exception"]}}
]

Now generate {n} DIVERSE seeds in the {category} category for area {area_hint}. Vary the sub-issues. Output ONLY the JSON array — no preamble, no markdown fence, no commentary. Start with `[` and end with `]`.
"""
    # Minimax M2.7 on macmini — direct OpenAI-compatible HTTP. Free local
    # compute, separate from frontier rate-limit pools. Quality similar to
    # frontier on simple generation tasks (no tool-call chain here).
    result = minimax_call(prompt)
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
