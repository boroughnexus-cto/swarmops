---
classification: non-sensitive
type: research-brief
project: LegalLLM-Training / BriefBox
created: '2026-05-23'
status: STUB — original lost to vault auto-sync; needs full re-write
recovered: '2026-05-23 — original deleted by vault auto-sync. Full text recoverable from prior session transcript at ~/.claude/projects/-home-sbarker-swarmops/1b7601ee-b048-44c6-95f5-e0a16cbd336d.jsonl'
tags:
- research
- gpt-oss
- capability
- briefbox
- legal-llm
- charlotte
- STUB
---

> **STUB — this doc's original content was lost on 2026-05-23 when the Obsidian vault auto-sync wiped the file. Below is the TL;DR from the Index description. Full text is recoverable from the prior session's JSONL transcript.**

# Research: gpt-oss-120b capability fit for Briefbox (2026-05-23) — STUB

## TL;DR (recovered from Index description)

**Peer-validated research brief on whether fine-tuned gpt-oss-120b can handle the real Briefbox workload.**

Defensible for v1 BUT only as a replaceable generation component. LegalBench is a weak signal (US-law biased). Charlotte's "missed authority" needs a pipeline trace before being classified. SFT alone is insufficient for OSCOLA — need metadata + prompting + validator + SFT layered. Multi-case synthesis is the major untested capability gap. Build a 25-50 task Charlotte harness BEFORE deciding model class.

## Recovered key findings

- **LegalBench 75.94% (rank 76/109)** for gpt-oss-120b — weak signal because LegalBench is heavily US-law biased.
- **Charlotte's complaints unpack into three different failure modes:**
  - Format / citation discipline (FIXABLE by SFT)
  - UK-specific case-law knowledge depth (HARDER — fine-tune training data is the bottleneck)
  - Pinpoint-paragraph extraction (mostly a retrieval/extraction problem, NOT generation)
  - Multi-case synthesis (deep legal-knowledge task; capability concern)
- **"Missed authority" complaints from production cannot be safely classified as generator vs retrieval vs corpus without a pipeline trace.** This motivated the [[Research-Action-Playbook-2026-05-23]] action 1.
- **Defensible for v1** as a replaceable generation component — architecture should not lock us into this specific base.

## To restore the full original

```
~/.claude/projects/-home-sbarker-swarmops/1b7601ee-b048-44c6-95f5-e0a16cbd336d.jsonl
```
Look for Write tool calls targeting `Research-gpt-oss-120b-capability-fit-2026-05-23.md`.

## Related

- [[Research-Action-Playbook-2026-05-23]] — the operational playbook derived from this brief
- [[Multi-Model-Architecture-2026-05-23]] — model-class decisions informed by this brief
- [[Fine-Tune-Recipe-2026-05-23]] — execution plan that follows from the brief's recommendations
- [[2026-05-23-missed-authority-EWHC-Ch-1031]] — the concrete incident whose root cause was identified per this brief's pipeline-trace recommendation
