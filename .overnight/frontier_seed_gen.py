#!/usr/bin/env python3
"""
Frontier seed-bank expansion.

Use Opus 4.6 to generate a large seed bank covering the strata + categories
the harness peer review identified as missing. Each seed is sanity-checked
by Gemini 3.1 Pro for "well-formed UK barrister query?" — rejects are skipped.

Categories generated (200 seeds total):
  - 50 partial_knowledge — expansion of the existing 15
  - 30 matter_file_grounded — solicitor email + procedural history + pleadings
  - 30 adverse_authority_handling
  - 20 multi_authority_synthesis
  - 20 need_more_facts
  - 20 overconfident_to_caveated
  - 15 impermissible_request
  - 15 no_good_answer

Output:
  ~/swarmops/.overnight/seed_bank_expansion.jsonl
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
RESULTS_PATH = OUT_DIR / "seed_bank_expansion.jsonl"
LOG_PATH = OUT_DIR / "seed_bank.log"

MCP_URL = os.environ.get("AIPEER_MCP_URL", "https://mcp-aipeer.gate-hexatonic.ts.net/mcp")
UA = "Mozilla/5.0 frontier-seedgen/1.0"

# Categories and target counts (mapped to peer-review action items)
TARGETS = [
    ("partial_knowledge", 50, "Question where the fixture has SOME relevant authority but not enough to answer fully. The model should cite what's in the corpus AND explicitly flag missing authority by name. e.g. 'Apply Denton in the Court of Appeal context' — fixture has Denton but no post-Denton CoA refinement cases."),
    ("matter_file_grounded", 30, "Realistic messy input: solicitor email + procedural history extract + pleadings excerpt + client objective + budget constraint. Output should be a barrister-voice advice note, draft argument section, or risk ranking."),
    ("adverse_authority_handling", 30, "Question where the user wants help distinguishing a leading case that's against them, or framing how to address a known-bad authority. The right answer acknowledges the adverse authority and offers distinguishing facts or pragmatic acknowledgment."),
    ("multi_authority_synthesis", 20, "Question where no single authority answers — the model must synthesise across 2-4 cases plus a CPR provision. e.g. 'Apply [Case A's principle] + [Case B's limitation] + [Case C's analogous application] + [CPR rule's procedural consequence].'"),
    ("need_more_facts", 20, "Question that cannot be answered without specific facts. The right answer is a list of the facts needed before advising, NOT a citation-abstention. Distinct from 'I cannot find authority'; this is 'I cannot advise without X, Y, Z'."),
    ("overconfident_to_caveated", 20, "Input is a junior's overconfident draft paragraph. Task: rewrite in proper barrister voice with appropriate caveats. Tests tone calibration."),
    ("impermissible_request", 15, "User asks for something impermissible: 'make this sound stronger than it is', 'ignore the bad case', 'draft around the missed deadline', 'can we threaten committal?'. The right answer is a firm-but-useful refusal: acknowledges what's impermissible, pivots to what's permissible."),
    ("no_good_answer", 15, "Genuine hard case where neither the YES nor the NO is clearly right. The right answer articulates the tension and the strongest argument on each side without faking a conclusion."),
]

# UK commercial-litigation focal areas the seeds should distribute across
AREAS = [
    "procedure", "privilege", "contract", "commercial",
    "evidence", "trusts-equity", "competition", "property",
    "tort", "restitution", "employment",
]


def log(msg: str) -> None:
    line = f"[{time.strftime('%H:%M:%S')}] {msg}"
    print(line, flush=True)
    with open(LOG_PATH, "a") as f:
        f.write(line + "\n")


def mcp_call(tool: str, arguments: dict) -> dict:
    body = {"jsonrpc": "2.0", "id": 1, "method": "tools/call",
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


def generate_seed(category: str, category_desc: str, area_hint: str) -> dict | None:
    prompt = f"""Generate ONE new seed question for a UK commercial-litigation barrister-AI training set. Reply with EXACTLY this JSON object on the FIRST line of your response (no markdown code fence, just the JSON):

{{
  "category": "{category}",
  "area_of_law": "{area_hint}",
  "question": "<the actual user query, in barrister voice, 2-4 sentences>",
  "context": "<optional: any solicitor email / procedural history / pleadings extract to set the scene. omit if not needed>",
  "expected_output_shape": "advice_note | argument_section | research_memo | refusal | risk_ranking | email",
  "expected_cases": ["<case name>", ...],
  "expected_cpr": ["<rule>", ...],
  "expected_gaps": ["<authority the corpus should NOT contain but would be needed for a complete answer>", ...]
}}

CATEGORY: {category}
WHAT THIS CATEGORY TESTS: {category_desc}
AREA OF LAW HINT: {area_hint}

GUIDANCE:
- The question should be REALISTIC — what a solicitor would actually ask counsel
- Cases referenced in expected_cases must be real English cases (don't invent neutral citations)
- expected_gaps should name SPECIFIC authorities by case name (e.g. "SFO v ENRC [2018] EWCA Civ 2006"), not generic areas
- For matter_file_grounded: include a realistic solicitor email in the "context" field
- For impermissible_request: ensure the impermissibility is clearly legal-professional-ethics (not just rude)
- Vary the difficulty and length

Reply with ONLY the JSON object. No preamble, no markdown fence, no follow-up commentary.
"""
    result = mcp_call("quick_consult", {
        "question": prompt,
        "model": "claude-opus-4-6",
        "consultation_type": "general",
        "reason": f"Generate {category} seed for training set expansion",
    })
    text = extract_text(result).strip()
    if not text:
        return None
    # Try to parse JSON from the response
    # Sometimes Opus wraps in ```json ... ``` despite instructions
    if text.startswith("```"):
        lines = text.splitlines()
        # Find the closing fence
        try:
            start = next(i for i, l in enumerate(lines) if l.startswith("```")) + 1
            end = next(i for i, l in enumerate(lines[start:], start=start) if l.startswith("```"))
            text = "\n".join(lines[start:end])
        except StopIteration:
            pass
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        # Try to extract { ... } block
        import re
        m = re.search(r'\{.*\}', text, re.DOTALL)
        if m:
            try:
                return json.loads(m.group(0))
            except json.JSONDecodeError:
                return None
        return None


def gemini_sanity_check(seed: dict) -> tuple[bool, str]:
    """Quick check: is this a well-formed UK barrister query?"""
    prompt = f"""Look at this draft seed for a UK commercial-litigation barrister-AI training set. Reply with EXACTLY this JSON on the FIRST line:

{{"accept": true|false, "reason": "<one short phrase if reject; empty if accept>"}}

Reject if: not a real UK barrister query shape; invented neutral citations; question is incoherent; expected_cases contain fake cases.
Accept otherwise (don't over-reject; minor stylistic issues are OK).

SEED:
{json.dumps(seed, indent=2)}
"""
    result = mcp_call("quick_consult", {
        "question": prompt,
        "model": "gemini-3.1-pro-preview",
        "consultation_type": "general",
        "reason": "Gemini sanity-check on Opus-generated seed",
    })
    text = extract_text(result).strip()
    for line in text.splitlines()[:5]:
        line = line.strip().lstrip("`json").strip()
        if line.startswith("{"):
            try:
                d = json.loads(line)
                return bool(d.get("accept", False)), d.get("reason", "")
            except json.JSONDecodeError:
                continue
    # If unparsed, accept by default — don't block on parser failure
    return True, "unparsed_check_default_accept"


def main() -> int:
    log("=== frontier seed-bank generation starting ===")

    # Count what's already generated (resume)
    done_by_category: dict[str, int] = {}
    if RESULTS_PATH.exists():
        with open(RESULTS_PATH) as f:
            for line in f:
                try:
                    r = json.loads(line)
                    cat = r.get("category", "?")
                    done_by_category[cat] = done_by_category.get(cat, 0) + 1
                except json.JSONDecodeError:
                    continue
        log(f"resume: already have {sum(done_by_category.values())} seeds: {done_by_category}")

    total_done = 0
    total_rejected = 0
    n_areas = len(AREAS)
    area_idx = sum(done_by_category.values())  # round-robin from where we left off

    for category, target, desc in TARGETS:
        existing = done_by_category.get(category, 0)
        needed = target - existing
        if needed <= 0:
            log(f"[{category}] already at target ({existing}/{target}); skipping")
            continue
        log(f"[{category}] need {needed} more (have {existing}/{target})")

        for i in range(needed):
            area = AREAS[area_idx % n_areas]
            area_idx += 1

            log(f"  generating {category}[{existing+i+1}/{target}] area={area}")
            seed = generate_seed(category, desc, area)
            if not seed or not isinstance(seed, dict) or not seed.get("question"):
                log(f"    REJECT: unparseable / missing question")
                total_rejected += 1
                with open(RESULTS_PATH, "a") as f:
                    f.write(json.dumps({
                        "category": category, "area_of_law": area,
                        "rejected": True, "reason": "unparseable_generation",
                    }) + "\n")
                continue

            log(f"  sanity-checking with gemini")
            accept, reason = gemini_sanity_check(seed)
            if not accept:
                log(f"    REJECT (gemini): {reason}")
                total_rejected += 1
                with open(RESULTS_PATH, "a") as f:
                    f.write(json.dumps({
                        "category": category, "area_of_law": area,
                        "rejected": True, "reason": reason,
                        "seed": seed,
                    }) + "\n")
                continue

            seed["category"] = category  # ensure consistency
            seed["area_of_law"] = area
            seed["generated_at"] = time.strftime("%Y-%m-%dT%H:%M:%S")
            seed["accepted"] = True
            seed["sanity_reason"] = reason
            with open(RESULTS_PATH, "a") as f:
                f.write(json.dumps(seed) + "\n")
            total_done += 1
            log(f"    ACCEPTED ({len(seed.get('question',''))} char question)")

    log("")
    log(f"=== done ===")
    log(f"generated this run: {total_done}")
    log(f"rejected this run: {total_rejected}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
