---
classification: non-sensitive
type: action-playbook
project: LegalLLM-Training / BriefBox
created: '2026-05-23'
status: STUB — original lost to vault auto-sync; needs full re-write
recovered: '2026-05-23 — original deleted by vault auto-sync. Full text recoverable from prior session transcript at ~/.claude/projects/-home-sbarker-swarmops/1b7601ee-b048-44c6-95f5-e0a16cbd336d.jsonl'
tags:
- playbook
- briefbox
- charlotte
- fine-tuning
- sprint
- STUB
---

> **STUB — this doc's original content was lost on 2026-05-23 when the Obsidian vault auto-sync wiped the file. Below is what's recoverable from the Index description. Full text is in the prior session's JSONL transcript.**

# Research Action Playbook (2026-05-23) — STUB

## TL;DR (recovered from Index description)

**Operational playbook for the 4 priority actions from [[Research-gpt-oss-120b-capability-fit-2026-05-23]]:**

1. **Trace the missed `[2020] EWHC Ch 1031`** through corpus → embed → rerank → context → generation. Five-step trace methodology.
2. **Build the Charlotte harness to 25-50 tasks + run the bake-off.** Replaces "more seeds" thinking — the bottleneck is harness coverage, not seed volume.
3. **Add deterministic citation controls** — metadata + validator + repair pass at the generation layer.
4. **Test multi-case synthesis on long context** — the major untested capability gap from the research brief.

Step-by-step + tools + done-when + 5-week sprint cadence. Replaces "more seeds" thinking; bottleneck is harness, not data.

## Recovered details

- Action 1 (trace `[2020] EWHC Ch 1031`) was executed 2026-05-23 — see [[2026-05-23-missed-authority-EWHC-Ch-1031]]. Root cause: case not in corpus (BBOX-36).
- Action 2 (Charlotte harness) is the v1 evaluation strategy — see [[Production-Eval-Pack-2026-05-22]] for the existing 15-task starter pack.
- Action 3 (citation controls) leads into the v15 prompt + tool architecture; OSCOLA validator is the deterministic piece.
- Action 4 (multi-case synthesis) is gated on Tracks A1 + A2 in [[Fine-Tune-Recipe-2026-05-23]].

## To restore the full original

```
~/.claude/projects/-home-sbarker-swarmops/1b7601ee-b048-44c6-95f5-e0a16cbd336d.jsonl
```
Look for Write tool calls targeting `Research-Action-Playbook-2026-05-23.md`.

## Related

- [[Research-gpt-oss-120b-capability-fit-2026-05-23]] — the research brief this playbook operationalises
- [[Fine-Tune-Recipe-2026-05-23]] — the fine-tune that supports actions 3 + 4
- [[2026-05-23-missed-authority-EWHC-Ch-1031]] — action 1 executed
- [[Pilot-Results-2026-05-22]] — pilot whose Gate 2 measurement informs action 2
- [[Production-Eval-Pack-2026-05-22]] — the harness this playbook scales to 25-50 tasks
