---
classification: non-sensitive
type: product-architecture
project: LegalLLM-Training / BriefBox
created: '2026-05-23'
status: STUB — original lost to vault auto-sync; needs full re-write
recovered: '2026-05-23 — original deleted by vault auto-sync. Full text recoverable from prior session transcript at ~/.claude/projects/-home-sbarker-swarmops/1b7601ee-b048-44c6-95f5-e0a16cbd336d.jsonl'
tags:
- briefbox
- standalone
- macbook
- m4-max
- mlx
- product
- STUB
---

> **STUB — this doc's original content was lost on 2026-05-23 when the Obsidian vault auto-sync wiped the file. Below is what's recoverable from the Index description. Full text is in the prior session's JSONL transcript.**

# Standalone Local Briefbox App (2026-05-23) — STUB

## TL;DR (recovered from Index description)

**Standalone local Briefbox app for M4 Max MBP** — gpt-oss-120b Q4_K_M MoE via MLX as the primary, gpt-oss-20b as quick mode, BGE-M3 + BGE-reranker for retrieval, LanceDB vector store. ~65GB on disk, ~70GB peak RAM.

**Killer feature: client-data privacy by physics.** Nothing leaves the machine. Critical for barrister workflows where uploading documents would risk legal privilege.

**Two tiers:**
- Chambers flagship £5-7K MBP
- Lite tier on M4 Pro

**Ships after cloud v1 stabilises.**

## Recovered architecture sketch

- **Primary inference:** gpt-oss-120b Q4_K_M MoE via Apple MLX (~40GB weights, runs on 128GB M4 Max)
- **Quick mode:** gpt-oss-20b for fast queries that don't need the full 120B
- **Embeddings:** BGE-M3 (568M params; runs on CPU)
- **Reranker:** BGE-reranker-v2-m3 (568M params; CPU)
- **Vector store:** LanceDB (local, file-based, no daemon)
- **Disk:** ~65GB total (model weights + corpus index)
- **Peak RAM:** ~70GB (gpt-oss-120b loaded + retrieval stack + headroom)

## Strategic rationale

1. **Privacy by physics** — barristers can ingest privileged documents into a local Briefbox without ever touching a cloud API.
2. **No subscription** — one-off purchase + the cost of the MacBook itself.
3. **Offline** — works on the train, in chambers, anywhere.
4. **Owns the data** — chambers don't trust their own data to a third party.

## To restore the full original

```
~/.claude/projects/-home-sbarker-swarmops/1b7601ee-b048-44c6-95f5-e0a16cbd336d.jsonl
```
Look for Write tool calls targeting `Standalone-Local-App-2026-05-23.md`.

## Related

- [[Multi-Model-Architecture-2026-05-23]] — model-class decisions; the standalone app uses the same architecture
- [[Fine-Tune-Recipe-2026-05-23]] — the fine-tunes that ship inside the standalone app
- [[Research-gpt-oss-120b-capability-fit-2026-05-23]] — capability research informs whether 120B Q4 is enough for the workload
