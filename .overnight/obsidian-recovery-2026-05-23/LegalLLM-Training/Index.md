---
classification: non-sensitive
type: training-log
project: LegalLLM-Training
---

> **NON-SENSITIVE — system-prompt engineering work for an English civil procedure benchmark. No client data, no PII, no privileged content. All citations and case names are public-domain references; no advice is being given.**

# LegalLLM Training — System Prompt Iteration

> Iterative refinement of an English commercial litigation system prompt for use with [[Charlotte-Law-AI-LegalLLM-Setup]] and BriefBox legal extraction work.

**Date range:** 2026-04-30 → 2026-05-01
**Target model:** MiniMax M2.7 (FP8, on `molecule-air`)
**Comparator:** SaulLM-141B-Instruct (legal-specialist, also on `molecule-air`)
**Reviewers:** GPT-5.2 / GPT-5.3 Codex / Gemini 3.1 Pro via [[tkn-aipeer]]

## Purpose

Develop a reliable, reusable system prompt that improves a generalist LLM's performance on English civil procedure questions — specifically: defendant participation rights when no defence has been filed, debarring orders, and the effect of debarring orders. The benchmark question is a fraud-litigation procedural question framed for a legally-trained practitioner.

## Files

- [[Original-Question]] — the test question (parts a–c)
- [[Use-Case-Charlotte-Drafting-Tool]] — what the prompt is being optimised for
- [[OpenWebUI-Custom-Model-Architecture]] — production architecture: raw SGLang + OpenWebUI Custom Models
- [[Interface-Contract]] — deployment topology decisions (Stage B output of iteratecycle)
- [[Deployment-Runbook]] — **operator runbook for the deployed Minimax Commercial Law tool**
- [[Iteration-Log]] — score progression, key findings, peer-review synthesis
- [[Finetuning-Plan-2026-05-22]] — **peer-validated execution plan for SFT + DPO fine-tuning on gpt-oss-120b** (5 phases: eval harness → base-model bake-off → grounded SFT → DPO from corrective edits → runtime citation auditor)
- [[Data-Generation-Pipeline-2026-05-22]] — **detailed engineering plan for SFT/DPO data generation** using Minimax M2.7 (LiteLLM-routed primary generator) + Opus/Gemini/GPT-5.5 (validators + preferred-output sources) + six anti-hallucination negative-example techniques
- [[Pilot-Results-2026-05-22]] — **50-seed validation pilot results**: pipeline works (0 errors), 62% Gate 1 (8 pts short), Gate 4 PASS, frontier consensus caught audit blind spot on substantive proposition correctness; three calibration issues identified for pre-scale fix
- [[AI-Training-Module-Spec-2026-05-22]] — **Supabase-backed UI for Charlotte's gate-2 triage**: schema, RLS, API contract, ingestion script, deferred-items list. Branch `feature/ai-training-module` in BriefBox-UI.
- [[Harness-Peer-Review-2026-05-22]] — **3-model architecture review of the whole harness**: thesis is "good data system, wrong objective". Action list incl. new partial_knowledge stratum, frozen production-shaped eval pack, barrister_correction capture in triage UI, reduced volume targets, pairwise comparison.
- [[Multi-Model-Architecture-2026-05-23]] — **Architecture decision** on whether to fine-tune 4 specialised gpt-oss models (RAG / reasoning / reranking / embeddings) vs the current monolithic plan. TL;DR: yes to specialisation, no to gpt-oss for all four — embeddings need a dual-encoder (BGE-M3), reranker needs a cross-encoder (BGE-reranker-v2-m3). Recommend ship monolithic v1 first, plan modular v1.5. **(STUB — full text lost to vault auto-sync 2026-05-23; reconstructed from prior session transcript.)**
- [[Standalone-Local-App-2026-05-23]] — **Standalone local Briefbox app for M4 Max MBP**: gpt-oss-120b Q4_K_M MoE via MLX as the primary, gpt-oss-20b as quick mode, BGE-M3 + BGE-reranker for retrieval, LanceDB vector store. ~65GB on disk, ~70GB peak RAM. Killer feature: client-data privacy by physics. Two tiers (Chambers flagship £5-7K MBP, Lite on M4 Pro). Ships after cloud v1 stabilises. **(STUB — full text lost to vault auto-sync 2026-05-23.)**
- [[Research-gpt-oss-120b-capability-fit-2026-05-23]] — **Peer-validated research brief** on whether fine-tuned gpt-oss-120b can handle the real Briefbox workload. TL;DR: defensible for v1 BUT only as a replaceable generation component. LegalBench is a weak signal (US-law biased). Charlotte's "missed authority" needs a pipeline trace before being classified. SFT alone is insufficient for OSCOLA — need metadata + prompting + validator + SFT layered. Multi-case synthesis is the major untested capability gap. Build a 25-50 task Charlotte harness BEFORE deciding model class. **(STUB — full text lost to vault auto-sync 2026-05-23.)**
- [[Research-Action-Playbook-2026-05-23]] — **Operational playbook** for the 4 priority actions from the research brief: (1) trace the missed `[2020] EWHC Ch 1031` through corpus/embed/rerank/context/generation, (2) build the Charlotte harness to 25-50 tasks + run the bake-off, (3) add deterministic citation controls (metadata + validator + repair pass), (4) test multi-case synthesis on long context. Step-by-step + tools + done-when + 5-week sprint cadence. **(STUB — full text lost to vault auto-sync 2026-05-23.)**
- [[Fine-Tune-Recipe-2026-05-23]] — **4-track fine-tune execution plan** (Track A: gpt-oss-120b SFT+DPO via TRL+PEFT bf16 LoRA; Track B: gpt-oss-120b LoRA for structured extraction; Track C: BGE-M3 unified InfoNCE+self-distill via FlagEmbedding; Track D: BGE-reranker-v2-m3 listwise CE). Blackwell-specific: no FA3/FA4 on sm_120, bf16 LoRA preferred over QLoRA with 384 GB VRAM, FSDP2 over ZeRO-3, never LoRA the MoE router. Env live at `/data/train-venv` on molecule-air; smoke-tested 2026-05-23.
- [[Licensing-Considerations-2026-05-23]] — **Commercial licensing due-diligence** for the entire fine-tune stack with verbatim ToS quotes from OpenAI (Dec 2025), Anthropic (Jun 2025), Google (Feb/Mar 2026), GitHub Copilot (Mar 2026), and licence findings for gpt-oss-120b (Apache 2.0 ✅), BGE models (MIT/Apache ✅), Llama-4 (community licence with attribution+naming+AUP conditions), MiniMax-M2 (❌ "Modified MIT" actually prohibits commercial use), SaulLM (MIT). **Decision recorded:** Briefbox is personal-research-scope; seed bank approved for use. Reopen if scope changes to commercial offering.
- [[Fine-Tune-Execution-Next-Steps-2026-05-23]] — **Execution runbook for resuming the 4-track programme.** Three entry-point options (data prep first / smoke test first / both parallel) with step-by-step checklists, skeleton scripts (Harmony wrap, hard-negative mine, smoke-test SFT), inference bring-down/bring-up notes, smoke-test success criteria, and a resume-command snippet for the next session. Read this FIRST when picking up the project again.
- [[Triage-Rubric-2026-05-23]] — **The gold/silver/reject grading rubric for Charlotte at `/training` ([live in BriefBox-UI PR #21](https://github.com/BoroughNexus/BriefBox-UI/pull/21)).** Litmus test (her wording, 2026-05-23): *"Could I put my name on this if I were rushing? — correct, competent & useful even if not to personal taste/style/extreme level of detail."* Critical principle (her refinement, 2026-05-23): **framing trumps substance** — same superseded-test answer can be gold, silver, or reject depending on whether it flags the supersession and adds value on what current law is. 21 grey areas in 6 categories (wrong vintage / wrong authority / hierarchy confusion / wrong jurisdiction / missing piece / format). Maps onto existing `reasons` enum. Drives the in-app contextual crib sheet modal — auto-opens first visit per browser, on-demand thereafter via `?` icon next to the *Decision* label.
- [[Bench-Harness]] — the original (no-tools) harness script
- [[Tool-Architecture]] — v11+ tool-augmented grounding (mock now, BriefBox/Firecrawl later)
- [[SaulLM-Diagnostic-Notes]] — bring-up fixes for SaulLM-141B on SGLang/Blackwell
- `Prompts/v0.md` … `Prompts/v15.md` — each prompt version. **v15 = deployed**.
- `Deployment/` — the as-deployed artefacts (FastAPI app, systemd unit, Custom Model JSON, register script)
- `Functions/minimax_legal_tools.py` — the OpenWebUI Tool (LLM-callable wrapper)
- `Tests/test_legal_tools_svc.py` — 15-test pytest suite for the service
- `Tests/negative_cases_test.py` — 24-case leak regression test
- `Fixtures/legal_tools_mock.py` — fixture data (15 cases, 29 CPR entries)
- `Answers/answer-v6.md` — best generalist run (27/40)
- `Answers/answer-v8.md` — broader "what remains live" framing
- `Answers/answer-v10.md` + peer-review.md — whitelist approach (~34/40, zero fabrications)
- `Answers/answer-v11.md` + peer-review.md + transcript.json — tool-augmented (~32.5/40)
- `Answers/answer-v14.md` + transcript.json — **production-candidate output (~37/40)**

## Key Findings

1. **Generalist LLMs hallucinate citations confidently.** No amount of substantive prompt engineering eliminates this. **Solution that worked:** either embed a curated citation whitelist in the prompt with explicit `[verified]` / `[neutral citation to verify]` markers (v10), OR provide tools that verify each citation against an authoritative source before emission (v11+). Both deliver zero fabrications.
2. **Tool grounding ≠ proposition grounding.** v11 (tool-only) eliminated fabrication but introduced new substantive errors (rule-misapplication) that v10's substantive prompt prevented. The substantive framework in the prompt is doing real work; tools complement it but cannot replace it.
3. **Three procedural starting positions are essential framing.** No defence filed (no order); defence struck out under CPR 3.4; express debarring order in place. v6 first introduced this; v14 split Position A into three sub-aspects (positive case / testing claimant's evidence / general procedural rights) that finally cleared the persistent reviewer complaint.
4. **Burden of proof has two distinct stages.** Pre-judgment: claimant must satisfy court on evidence. Post-judgment: elements of cause of action established; only quantum and remedy specifics remain. Conflating these is the most common error.
5. **CPR Part 13 set-aside applies ONLY to Part 12 default judgments.** Not to CPR 3.5 judgments. Practitioner intuition often gets this wrong.
6. **A debarring order is procedurally gated, not ousted.** It cannot extinguish core CPR-conferred rights (3.9 / 3.1(7) / Part 52); it can only impose permission gates and timing constraints. v14's "procedurally gated, not ousted" framing was the unlock.
7. **Score progression v0 → v14: 17 → 37 (out of 40).** Plateau at 27 was when the use case + acceptance criterion were unspecified; once "Charlotte's drafting tool + zero fabrication" became the target, the next 5 iterations took us to 37.
8. **Layered architecture** — raw model in SGLang, routing in model-proxy, use-case packaging (system prompt + tools) in OpenWebUI Custom Models. The legal prompt belongs at the OpenWebUI layer, not in the inference stack. See [[OpenWebUI-Custom-Model-Architecture]].

## Score Progression (4 dimensions × 10 = 40 max)

| Version | Legal | Cite | Complete | Utility | Total | Note |
|---------|-------|------|----------|---------|-------|------|
| v0      | 4     | 3    | 6        | 4       | 17    | generic baseline |
| v1      | 6     | 5    | 6        | 6       | 23    | + CPR rules + scenarios |
| v2      | 6.5   | 6    | 7        | 7       | 26.5  | + 3-stage framework |
| v3      | 6.5   | 5.5  | 7.5      | 6.5     | 26    | qualified Denton/BPP |
| v4      | 6     | 5    | 6.5      | 6.5     | 24    | residual rights pass |
| v5      | 5     | 5    | 7        | 5       | 22    | regression — burden of proof error |
| **v6**  | **6** | **6**| **8**    | **7**   | **27**| **best generalist run** |
| v7      | 5     | 4    | 6        | 5       | 20    | regression — "cannot cross-examine" too absolute |
| v8      | 6.5   | 5    | 8        | 7       | 26.5  | broader "what remains live"; HRH Prince Abdulaziz citation flagged |
| v9      | tbd   | tbd  | tbd      | tbd     | tbd   | citation discipline; sub-rule discipline |
| v10     | 7.5   | 9.5  | 8.5      | 8.5     | 34    | whitelist embedded; first zero-fabrication |
| v11     | 7     | 9.5  | 8        | 8       | 32.5  | tool-augmented (lost substantive accuracy) |
| v12     | 7.5   | 9    | 8.5      | 8       | 33    | combined whitelist + tools |
| v13     | 8     | 9.5  | 9        | 8.5     | 35    | Position A reframed |
| v14 | 9.5 | 9.5 | 9 | 9 | 37 | iteratecycle predecessor |
| v14 (disclosure overfit-check) | 8 | 8 | 9 | 9 | 34 | not overfit — fixture gap explained the drop |
| **v15** | **9.5** | **9.5** | **9** | **9** | **~37** | **DEPLOYED — Minimax Commercial Law on molecule-air**; OMIT-entirely + tool-error handling + lookup budget |
| v15 (disclosure) | 8.5 | 9 | 9 | 9 | ~36 | fixture expansion lifted disclosure score |
| SaulLM-v15 | tbd | tbd  | tbd | tbd | tbd | deferred — SGLang inference-time issue |

## Related Notes

- [[Charlotte-Law-AI-LegalLLM-Setup]] — broader legal-AI infrastructure project
- [[Charlotte-Law-AI-Legal-Case-Law-Crawling]] — case-law data ingestion pipeline (precursor to RAG)
- [[Charlotte-Law-AI-Finetune-Commercialisation]] — fine-tune commercialisation thinking
- [[BriefBox]] — entity-extraction pipeline that uses the same MiniMax M2.7 backend
