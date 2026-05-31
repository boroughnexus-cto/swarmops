---
classification: non-sensitive
type: architecture-discussion
project: LegalLLM-Training / BriefBox
created: '2026-05-23'
status: STUB — original lost to vault auto-sync; needs full re-write
recovered: '2026-05-23 — original deleted by vault auto-sync; this stub reconstructed from Index description + partial Read in subsequent session. Full text recoverable from prior session transcript at ~/.claude/projects/-home-sbarker-swarmops/1b7601ee-b048-44c6-95f5-e0a16cbd336d.jsonl'
tags:
- briefbox
- fine-tuning
- architecture
- gpt-oss
- rag
- embeddings
- reranking
- charlotte
- STUB
---

> **STUB — this doc's original content was lost on 2026-05-23 when the Obsidian vault auto-sync (running from nuc-ubuntu-dev) wiped the file from the working tree before it could be committed. Below is what's recoverable from the Index description and a partial read in a later session. Full original is in the prior session's JSONL transcript and can be reconstructed when needed.**

# Multi-Model Architecture Decision (2026-05-23) — STUB

> Should we fine-tune four specialised gpt-oss models — one each for RAG, reasoning, reranking, and embeddings — instead of one general-purpose fine-tune? Or is the right answer "specialise, but only some of them, and not all from gpt-oss"?

## TL;DR (recovered from Index description)

**Yes to specialisation, no to gpt-oss for all four roles.** The architectural instinct (one model per pipeline role) is correct. The model-class assumption (use gpt-oss for everything) is wrong for embeddings and overkill for reranking.

The right shape is:

| Role | Right model class | Base model | Why this base |
|---|---|---|---|
| **Embeddings** | dual-encoder ~500M params | **BGE-M3** or jina-embeddings-v3 | Purpose-built contrastive-trained encoder; produces better embeddings than any decoder at 200× lower inference cost |
| **Reranking** | cross-encoder 500M-3B | **BGE-reranker-v2-m3** or mxbai-rerank-large | Pairwise relevance scoring is the task; cross-encoders dominate cross-encoder benchmarks |
| **Generation / RAG** | decoder 70-120B | **gpt-oss-120b** | This is the current plan; the only role where gpt-oss is the right base |
| **Runtime auditor** | small decoder 3-8B | **gpt-oss-20b** or Llama-3-8B (or Opus 4.6 via API) | Hallucination detection — could be fine-tuned or just call frontier API |

## Recovered key points

- **gpt-oss-120b is the wrong model class for embeddings.** It's a generative decoder MoE; embeddings need a fundamentally different architecture (dual-encoder, contrastive InfoNCE loss). Mining hidden states from a decoder gives inferior embeddings to a purpose-built encoder, at vastly higher cost. The MTEB leaderboard is dominated by ≤7B encoder-derived models. **Hard architectural "no".**
- **gpt-oss-120b is overkill for reranking.** Cross-encoders (500M-3B params) dominate the relevant benchmarks for pairwise relevance scoring. Using a 120B decoder to rerank is using a tractor to slice a tomato — 200× the inference cost, 50× the latency, no recall@k benefit.
- **Generation / RAG / reasoning is where gpt-oss IS the right base.** Decoder architecture matches the task; 120B-class capacity justified for legal reasoning. This is the current plan; don't fork it.

## Strategic trade-offs

**Pro splitting** (architecturally):
1. Each component is purpose-built for its task; one generalist will be worse than the sum of three specialists.
2. Cost-at-scale is dramatically lower (embeddings + reranking run on small GPU / CPU; generation is the only expensive step).
3. Each component has a clean evaluation metric: recall@k for embeddings, nDCG@k for reranker, the v15 rubric + production-eval-pack for generator.
4. The retrieval stack is where most legal-AI failures hide. Stanford HAI's Legal RAG Hallucinations study (2024) showed commercial legal RAG tools hallucinate at 17-33%; most failures originate in retrieval. **Investing in embeddings + reranker is probably higher leverage than investing more in the generator beyond v15.**
5. Modular swappability — can upgrade any one component independently.

**Recommendation: ship monolithic v1 first, plan modular v1.5.**

## To restore the full original

The complete content of this doc is in this conversation's predecessor session JSONL:
```
~/.claude/projects/-home-sbarker-swarmops/1b7601ee-b048-44c6-95f5-e0a16cbd336d.jsonl
```
Look for Write tool calls targeting `Multi-Model-Architecture-2026-05-23.md`.

## Related

- [[Fine-Tune-Recipe-2026-05-23]] — implements the model-class decisions from this doc
- [[Research-gpt-oss-120b-capability-fit-2026-05-23]] — the capability research that motivated this discussion
- [[Finetuning-Plan-2026-05-22]] — earlier high-level plan
