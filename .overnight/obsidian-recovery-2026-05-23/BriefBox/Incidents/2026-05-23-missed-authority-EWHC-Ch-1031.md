---
classification: non-sensitive
type: incident-report
project: BriefBox
created: '2026-05-23'
incident-date: '2026-05-22'
reporter: Charlotte Thomas (Brick Court Chambers)
severity: high
status: root-cause-identified
recovered: '2026-05-23 — vault auto-sync deleted this file; reconstructed from conversation transcript'
tags:
- briefbox
- corpus-coverage
- crawler
- incident
- bailii
- charlotte
---

> **NON-SENSITIVE — incident report on missed-authority complaint. No client data, no PII. The cited case is public-domain BAILII content.**

# Incident: Missed authority `[2020] EWHC 1031 (Ch)` (De Sena v Notaro)

## Summary

On 2026-05-22 Charlotte Thomas asked Briefbox:

> "cases where an expert was criticised for taking an academic approach and not having real-world experience"

The model identified `[2015] EWHC 3294 (Pat)` as the only relevant case in corpus. Charlotte fed back via the in-app reporter that **`De Sena v Notaro [2020] EWHC 1031 (Ch) at [157]`** was a known relevant authority that should have been returned (https://www.bailii.org/ew/cases/EWHC/Ch/2020/1031.html).

This is a famous expert-witness criticism case in commercial-litigation circles, widely cited in practitioner journals as a "stark warning to expert witnesses".

## Root cause

**The case is not in the corpus.** Pipeline trace stopped at Step 1.

The crawler systematically omits **EWHC (Ch) decisions from 2019 and 2020 entirely.** Year-distribution of EWHC (Ch) judgments in `corpus_documents`:

| Year | Count |
|---|---|
| 2026 | 164 |
| 2025 | 554 |
| 2024 | 557 |
| 2023 | 575 |
| 2022 | 458 |
| 2021 | 113 |
| **2019-2020** | **0** |
| 2018 | 1 |
| 2004-2017 | ~6 total |

The crawler's effective coverage window starts at 2021 with sharp drop-offs on both sides. Five years of recent Chancery jurisprudence (2017-2020) — the vintage that includes many post-financial-crisis and pre-COVID principles still being argued — are essentially absent.

## Pipeline trace

| Step | Result |
|---|---|
| 1. Case in `corpus_documents`? | **NO** (only `[2025] EWHC 1031 (Ch)` returned for the EWHC Ch 1031 pattern; that's a different case) |
| 2. Chunking of paragraph 157? | N/A — case not in corpus |
| 3. Embedding retrieves it? | N/A — case not in corpus |
| 4. Reranker keeps it in top-N? | N/A — case not in corpus |
| 5. Model surfaces it from context? | N/A — case not in corpus |

The trace stopped at Step 1 because the root cause was found there.

## Other authorities likely also missing

| Year | Case | Why it matters |
|---|---|---|
| 2018 | Director of the Serious Fraud Office v ENRC [2018] EWCA Civ 2006 | Defining case on regulatory-investigation litigation privilege |
| 2018 | First Tower Trustees Ltd v CDS [2018] EWCA Civ 1396 | Reliance + entire-agreement clauses |
| 2019 | Pakistan International Airline Corp v Times Travel [2019] EWCA Civ 828 | Economic duress |
| 2019 | Sevilleja v Marex Financial [2018] EWCA Civ 1468 / [2020] UKSC 31 | Reflective loss / Foss v Harbottle (UKSC 2020 may be missing too) |
| 2020 | De Sena v Notaro [2020] EWHC 1031 (Ch) | Expert-witness criticism (this incident) |
| 2020 | Sky v SkyKick [2020] EWHC 990 (Ch) | Trade mark bad faith |
| Various 2015-2017 | Cavendish v Makdessi [2015] UKSC 67, Marex precursor, etc. | Foundational |

## Proposed fix (in priority order)

### 1. Immediate (this week)

**Extend crawler date range to 2015–2020 for EWHC (Ch) and EWCA (Civ).** Likely a one-line change to the crawler config — find the `crawler_sources` table or the schedule that bounds the date range.

```sql
SELECT id, name, base_url, year_range_start, year_range_end, status
FROM crawler_sources
WHERE name ILIKE '%bailii%' OR base_url ILIKE '%bailii%';
```

(The crawler subprocess currently has a separate schema-mismatch issue — `column "year" does not exist` — which surfaced when the briefbox container was restarted today as part of the langfuse fix. That needs fixing alongside this.)

### 2. Short term (next sprint)

**Backfill historical EWHC, EWCA, UKSC, UKHL coverage for 2000-2020.** A one-shot crawl of the BAILII corpus for these courts/years should add 50-100K cases to the existing 99K. Storage isn't a constraint (already on /tank with TB headroom).

### 3. Medium term (informed by the bake-off)

**Build a corpus-coverage benchmark** of 50-100 famous English commercial cases that any barrister-grade tool MUST be able to surface. Run as a regular health check.

## What we should NOT do (yet)

- **Don't fine-tune embeddings on legal pairs** to fix this incident specifically — that addresses a different failure mode (top-K retrieval ordering) which we now know is not the root cause here.
- **Don't rewrite the reranker** for the same reason.
- **Don't blame the model.** This is not a generator failure; the chunk wasn't in context.
- **Don't ship the fine-tuned model with a corpus gap** — the model can be perfect and still miss authorities that aren't in the index.

## What this incident shows about the broader architecture

Per [[Research-gpt-oss-120b-capability-fit-2026-05-23]] and the peer review consensus: **"missed authority" complaints from production cannot be safely classified as generator vs retrieval vs corpus without a pipeline trace.** This is the playbook in action — 20 minutes of SQL identified the root cause without touching the model, embeddings, or reranker.

It also reinforces the [[Multi-Model-Architecture-2026-05-23]] argument: **the highest-leverage investment is upstream of the generator**, not in the generator itself.

## Communication to Charlotte

> Charlotte — quick update on the missed-authority you flagged (De Sena v Notaro at [157]). Traced it: the case isn't in our corpus at all. The crawler started in 2021 with a five-year gap on either side, so 2020 Chancery decisions are entirely absent — De Sena is one of probably hundreds of similar gaps. Fix is in the crawler, not the model: extending coverage to 2015-2020 EWHC Ch is a small config change. I'll ping you when the backfill is done and you can re-run the same query to confirm.

## Plane tracking

- **BBOX-36** (`b53da030-…`) — Corpus coverage gap — 2019-2020 EWHC (Ch) decisions entirely missing. This incident's primary ticket.
- **BBOX-37** (`e3a6bc09-…`) — Crawler subprocess in restart loop on `UndefinedColumnError("year")` / `UndefinedColumnError("canonical_form")`. Surfaced when the briefbox container was rebuilt for the langfuse fix. **BBOX-37 blocks BBOX-36** — without a functional crawler, the corpus-extension fix can't actually pull data.
- Both filed 2026-05-23, both High severity, cross-linked via Plane comments.

## Ops log — 2026-05-23

- **Langfuse runtime fix landed**: `requirements.txt` in `/etc/komodo/stacks/briefbox/` had `litellm==1.55.4` but the proxy config wires `success_callback: ["langfuse"]` (BBOX-14). On a fresh container build `litellm` was crashing with `ModuleNotFoundError: No module named 'langfuse'` and supervisord was respawning indefinitely. Added `langfuse>=2.27.0,<3.0` and rebuilt — clean boot, `api: ok / litellm: ok`. Committed at `BoroughNexus/BriefBox@d92eed7`. Peer-validated via Opus 4.6 `quick_consult` before merge.
- **Crawler subprocess crash now visible**: with litellm no longer eating the restart loop, the crawler subprocess inside the same container is now visibly hitting the `column "year" does not exist` schema error. This is BBOX-37; out of scope for today's Action B.
- **Overnight seed-bank artefacts backed up to Unraid**: 86MB / 48 files in `/mnt/user/backups/briefbox/seed-bank-overnight-2026-05-23/`.

## Related

- [[Research-Action-Playbook-2026-05-23]] — defined this trace procedure (Action 1)
- [[Research-gpt-oss-120b-capability-fit-2026-05-23]] — research brief that identified this as the highest-priority diagnostic
- [[Harness-Peer-Review-2026-05-22]] — strategic framing: data is the asset; corpus IS part of the data
- [[BriefBox-HippoRAG]] — corpus + extraction pipeline (this is the layer that needs the fix)
- [[BriefBox]] — production stack
- [[Multi-Model-Architecture-2026-05-23]] — argues for retrieval-stack investment over generator fine-tune
