---
classification: non-sensitive
type: training-plan
project: LegalLLM-Training / BriefBox
created: '2026-05-23'
status: ready-to-execute
hardware: molecule-air (4x RTX PRO 6000 Blackwell + 1x 2080 Ti, 251GB RAM)
recovered: '2026-05-23 — vault auto-sync deleted this file; reconstructed from conversation transcript'
tags:
- fine-tuning
- gpt-oss
- bge-m3
- bge-reranker
- briefbox
- molecule-air
- blackwell
---

> **NON-SENSITIVE — fine-tuning execution plan. No client data, no PII. All seeds are synthetic/public-corpus derived. The recipe targets the Briefbox legal-AI stack for Charlotte at Brick Court Chambers.**

# Fine-Tune Recipe (4-track) — 2026-05-23

> The end-to-end plan for fine-tuning the four trainable components of the Briefbox pipeline on **molecule-air**. Replaces the model-class decisions in [[Multi-Model-Architecture-2026-05-23]] with concrete commands, hyperparameters, and execution gates. All four tracks share one Python env (`/data/train-venv`) and one wandb project.

## Scope at a glance

| Track | Component | Base | Method | Data | VRAM | Wall-clock |
|---|---|---|---|---|---|---|
| **A** | Generation / RAG | `openai/gpt-oss-120b` | LoRA-bf16 SFT → DPO | 73k SFT seeds + 5–10k DPO pairs | ~250 GB sharded 4× | 48–72 h (SFT) + 12–24 h (DPO) |
| **B** | Extraction | `openai/gpt-oss-120b` | LoRA-bf16 SFT | 15–25k judgment→structured pairs | shares Track A infra | 24–36 h |
| **C** | Embeddings | `BAAI/bge-m3` | Unified InfoNCE + self-distill | 10–50k (q, pos, neg×7) | <40 GB across 4 cards | 4–8 h |
| **D** | Reranking | `BAAI/bge-reranker-v2-m3` | Listwise CE on triples | 10–30k (q, pos, neg×7) | <40 GB across 4 cards | 4–8 h |

**All four assume:** the 4× RTX PRO 6000 Blackwell cards are **dedicated to training** for the duration of the job. The 3× gpt-oss-120b inference replicas and the MiniMax-M2.7 server must be stopped before each Track A/B run. Tracks C/D are small enough to share with inference if we use only 2 cards.

## Hardware reality

| Resource | Value | Notes |
|---|---|---|
| GPUs (training-capable) | 4× RTX PRO 6000 Blackwell Max-Q, 96 GB each = **384 GB VRAM** | sm_120 / cc (12, 0) |
| GPU (small) | 1× RTX 2080 Ti, 11 GB | Suitable for BGE inference only; too small for any training |
| CPU | 2× AMD EPYC 9334 (32C/64T each = 64C/128T) | Plenty for data preprocessing |
| RAM | 251 GB | bf16 LoRA needs ~210 GB sharded across GPUs; CPU offload not required |
| Storage | `/data` 1.8 TB (537 GB free) · `/data2` 1.7 TB (873 GB free) · `/data3` 469 GB NVMe (408 GB free) | Use `/data3` for hot checkpoints (fastest), `/data2` for training data + finished adapters |

## Blackwell-specific gotchas (read first)

1. **Flash-Attention 3/4 do NOT run on sm_120 RTX PRO 6000.** FA3 is sm_80–9.0 only excluding 8.6/8.9; FA4 requires SM100 datacenter Blackwell (TMEM) — neither covers SM120 workstation Blackwell. **Use PyTorch SDPA** (`attn_implementation="sdpa"`) or the `kernels-community/vllm-flash-attn3` wheel from the `kernels` package (Blackwell-built).
2. **bitsandbytes works but has a documented energy/perf penalty on Blackwell.** With 384 GB VRAM we don't need the memory savings → **prefer bf16 LoRA over QLoRA** for Tracks A and B.
3. **The proven baseline for these exact cards is `torch 2.9.1+cu128`** (the same wheel `/data/sglang-venv` uses to serve gpt-oss-120b + MiniMax-M2.7 right now). Do NOT use Axolotl as it force-downgrades to torch 2.8.0; write training scripts directly with TRL+PEFT instead.
4. **gpt-oss-120b is Harmony-format only.** The TRL default chat-template mask can mask the `analysis` reasoning channel incorrectly — verify SFT loss is computed on the `<|channel|>final` segment and `<|return|>` stop token. Without this you train the model to drift off-format.

## Shared infrastructure (live on molecule-air now)

### Python env

```
/data/train-venv/                    # Python 3.12.3
  torch                  2.9.1+cu128 (bf16 native on sm_120)
  transformers           5.9.0       (gpt-oss model_type='gpt_oss' confirmed loads)
  trl                    1.4.0       (SFTTrainer + DPOTrainer)
  peft                   0.19.1      (LoraConfig)
  accelerate             1.13.0      (FSDP2 launcher)
  bitsandbytes           0.49.2      (kept for optional QLoRA fallback)
  datasets               4.8.5
  FlagEmbedding          1.4.0       (BAAI's official BGE-M3 + reranker trainer)
  sentence-transformers  5.5.1       (alternative path for BGE)
  kernels                0.14.1      (gives access to vllm-flash-attn3 Blackwell wheel)
  liger-kernel           0.8.0       (fused MoE kernels — speedup for Track A/B)
  hf_transfer            0.1.9       (faster HF downloads)
  wandb / tensorboard / sentencepiece / protobuf
```

**Activation:** `source /data/train-venv/bin/activate`.

**Smoke test (already verified 2026-05-23 10:33 BST):** transformers loads `gpt_oss` model_type with 128 experts / 4 active / hidden 2880 / 36 layers / vocab 201088 / EOS `<|return|>`. FlagEmbedding imports `BGEM3FlagModel` and `FlagReranker`. peft `LoraConfig(r=32, alpha=64, target_modules=['q_proj','k_proj','v_proj','o_proj','gate_proj','up_proj','down_proj'])` constructs without error. trl `SFTTrainer` and `DPOTrainer` import.

### Disk layout convention

```
/data2/training/
  data/                   # JSONL seed banks (mirror of ~/swarmops/.overnight/ for training)
    sft_legal_v1.jsonl
    dpo_legal_v1.jsonl
    extraction_v1.jsonl
    embed_pairs_v1.jsonl
    rerank_triples_v1.jsonl
  runs/                   # per-run output dirs (one per wandb run)
    A-sft-gptoss120b-r32-2026-05-25/
    A-dpo-gptoss120b-2026-05-27/
    B-extract-gptoss120b-r16-2026-05-29/
    C-bge-m3-2026-05-30/
    D-bge-rerank-v2-m3-2026-05-30/

/data3/training-outputs/  # Hot checkpoints (fastest NVMe) — periodically rsync to /data2/training/runs
```

### Models on disk

```
/data3/models/gpt-oss-120b                          # ✓ already cached
/data/models/bge-m3                                 # ✓ already cached  
~/.cache/huggingface/hub/models--BAAI--bge-reranker-v2-m3  # ✓ pulled 2026-05-23 (this session)
```

### wandb project

`briefbox-ft` — one project, four runs prefixed `A-…`, `B-…`, `C-…`, `D-…` plus a date stamp so reruns are diffable.

## Track A — Generation / RAG (gpt-oss-120b · SFT → DPO)

> **Goal:** Charlotte-grade barrister voice with OSCOLA citations, abstention calibration ("honest-incomplete"), and consistent multi-authority synthesis. Replaces the v15 prompt+tool stack at the OpenWebUI layer.

### Phase A1 — SFT on the 73k seed bank

**Data:** ~72k usable seeds (excluding the quarantined Minimax M2.7 batch) wrapped in Harmony format. The wrapper applies the `<|start|>system…<|end|><|start|>user…<|end|><|start|>assistant<|channel|>final<|message|>…<|return|>` template. The `analysis` (reasoning) channel is omitted from the training target — we train barrister voice on the final-answer channel only, preserving the model's native reasoning behaviour from pre-training.

**LoRA configuration:**
```python
from peft import LoraConfig, TaskType
LoraConfig(
    r=32,
    lora_alpha=64,
    target_modules=["q_proj","k_proj","v_proj","o_proj",
                    "gate_proj","up_proj","down_proj"],  # NEVER the router
    lora_dropout=0.05,
    bias="none",
    task_type=TaskType.CAUSAL_LM,
)
```

The seven targets cover the attention block (q/k/v/o) and the FFN-expert block (gate/up/down). The MoE router weights (`router.weight` / `experts.gate_weight`) are deliberately excluded — no published recipe LoRAs them and adapting them causes expert-assignment collapse.

**Training hyperparameters:**

| Param | Value | Rationale |
|---|---|---|
| precision | bf16 | Blackwell native; skips bnb perf cliff |
| sharding | FSDP2 (`accelerate launch --use_fsdp --fsdp_version 2`) | FSDP2 + MoE is the proven path; ZeRO-3 has known instability with MoE expert sharding |
| attention | `attn_implementation="sdpa"` (or `kernels-community/vllm-flash-attn3` if `kernels` resolves it on Blackwell) | FA3/FA4 don't support sm_120 |
| learning rate | 2e-5 (constant warmup → cosine) | Standard bf16 LoRA on 120B |
| warmup ratio | 0.03 | |
| batch (per device) | 1 | Sample-packed |
| grad accumulation | 4 (effective batch 16 across 4 cards) | |
| sequence length | 4096 | Charlotte queries + retrieved chunks rarely exceed this |
| sample packing | True | Critical for throughput at seq_len 4096 |
| epochs | 2 (configurable; check eval at end of epoch 1) | At ~72k seeds × 4k tokens = ~290M tokens, 2 epochs is the sweet spot before barrister-voice overfit |
| optimizer | AdamW (8-bit not needed at 384 GB) | |
| weight decay | 0.0 (LoRA) | |
| save_steps | 500 | |
| eval_steps | 250 | Tracked: train loss, eval loss on a held-out 500-seed slice, OSCOLA-citation rate on a 50-prompt sanity set |
| liger-kernel | enabled (drops MoE forward/backward to fused kernels) | ~20-30% throughput on gpt-oss MoE |

**VRAM budget:** bf16 base 120B ≈ 240 GB sharded 4× = ~60 GB/card. LoRA params + optim state + activations ≈ another 15-20 GB/card. **Total per card ≈ 75-80 GB**, leaving 16-20 GB headroom — comfortable. QLoRA NF4 (~65 GB single-card) is the fallback if FSDP fails.

**Wall-clock estimate:** 290M tokens × 2 epochs at ~6k tokens/sec/card on bf16 LoRA + SDPA → ~24-48 h.

**Eval gates (must pass to promote):**
- Eval loss ≤ start eval loss − 15%
- OSCOLA-citation rate on 50-prompt sanity set ≥ 90% (start ≈ 35% per Pilot-Results-2026-05-22)
- Zero hallucinated case names in 50-prompt sanity set
- The frozen production-eval-pack: per-task win rate vs v15 prompt ≥ 50% (Gate 4)

### Phase A2 — DPO on corrective edits

**Data:** ~5-10k `(prompt, chosen, rejected)` triples from:
- The 16 corrected-frontier-consensus DPO pairs from `dpo_pairs_results.jsonl`
- Future Charlotte triage decisions captured via the AI-Training-Module UI (`barrister_correction` capture flow)
- The "abstention preferred over confident-wrong" pairs from the seed bank (`overconfident_to_caveated` category, 5,995 seeds; preferred = the caveated variant, rejected = the original overconfident)

**Crucial: chosen AND rejected must both be wrapped in Harmony format.** If the rejected variant is not Harmony-wrapped, you train the model to leave Harmony for the rejected pattern — catastrophic for format consistency.

**Trainer:** `trl.DPOTrainer` on top of the **SFT-LoRA-merged** Track A1 checkpoint. The reference model is the same merged checkpoint loaded in 4-bit (NF4) — fine here even with the bnb Blackwell perf penalty, because ref-model forwards are pure inference.

**Hyperparameters:**

| Param | Value | Rationale |
|---|---|---|
| beta | 0.07 | Mid-range; lower = more drift from SFT, higher = barely moves |
| LoRA r/alpha | 16 / 32 | Half of SFT — DPO needs less capacity to move the policy |
| target_modules | same 7 as A1 | |
| learning rate | 5e-6 | DPO is 4-10× more sensitive than SFT; start low |
| batch | 1/device, grad_accum 8 (effective 32) | |
| max_length | 4096 | |
| epochs | 1 | DPO with >1 epoch usually overfits |
| precision | bf16 | |

## Track B — Extraction (gpt-oss-120b · LoRA SFT)

> **Goal:** Structured extraction from raw EWHC judgments. Outputs: `{neutral_citation, parties, judge, court, division, date, headnote_paragraphs, key_passages, citations_referenced}`. Replaces the regex+heuristic crawler chunking pipeline.

### Data

15-25k `(judgment_text, structured_extract_json)` pairs. Bootstrap: use Opus 4.6 + Gemini 3.1 Pro to extract structured payloads from 25k random `corpus_documents`. Filter to pairs where both models agree on `neutral_citation` and `parties`. Augment with hand-labelled 200-300 pairs for edge cases (consolidated appeals, multi-respondent, dissenting judgments, ex tempore rulings).

### LoRA configuration

```python
LoraConfig(
    r=16, lora_alpha=32,
    target_modules=["q_proj","k_proj","v_proj","o_proj",
                    "gate_proj","up_proj","down_proj"],
    lora_dropout=0.0, bias="none", task_type=TaskType.CAUSAL_LM,
)
```

### Training hyperparameters

Same env as Track A. Differences: sequence length 8192 (full judgments), batch 1, grad accumulation 8 (effective 32), epochs 1, hard `<|return|>` stop after closing `}` of JSON. F1 on 9 extracted fields against hand-labelled 300-sample test set as eval metric.

## Track C — Embeddings (BAAI/bge-m3 · Unified fine-tune)

> **Goal:** Lift retrieval recall@k on Charlotte's commercial-litigation query distribution. Current recall is the bottleneck per the missed-authority incident — once BBOX-36 backfills the corpus, the embeddings model needs to surface the new content reliably.

### Training command

```bash
source /data/train-venv/bin/activate
torchrun --nproc_per_node 4 \
  -m FlagEmbedding.finetune.embedder.encoder_only.m3 \
  --model_name_or_path /data/models/bge-m3 \
  --train_data /data2/training/data/embed_pairs_v1.jsonl \
  --output_dir /data3/training-outputs/C-bge-m3-$(date +%Y-%m-%d)/ \
  --train_group_size 8 \
  --query_max_len 512 --passage_max_len 512 \
  --per_device_train_batch_size 2 \
  --learning_rate 1e-5 \
  --num_train_epochs 2 \
  --temperature 0.02 \
  --knowledge_distillation True \
  --unified_finetuning True \
  --use_self_distill True \
  --self_distill_start_step 0 \
  --kd_loss_type m3_kd_loss \
  --bf16 \
  --logging_steps 50 --save_steps 500 \
  --gradient_checkpointing
```

**Volume target:** 30k queries.

## Track D — Reranking (BAAI/bge-reranker-v2-m3 · Listwise CE)

### Training command

```bash
source /data/train-venv/bin/activate
torchrun --nproc_per_node 4 \
  -m FlagEmbedding.finetune.reranker.encoder_only.base \
  --model_name_or_path ~/.cache/huggingface/hub/models--BAAI--bge-reranker-v2-m3/snapshots/* \
  --train_data /data2/training/data/rerank_triples_v1.jsonl \
  --output_dir /data3/training-outputs/D-bge-rerank-v2-m3-$(date +%Y-%m-%d)/ \
  --train_group_size 8 \
  --query_max_len 512 --passage_max_len 512 \
  --pad_to_multiple_of 8 \
  --knowledge_distillation False \
  --learning_rate 6e-5 \
  --bf16 \
  --num_train_epochs 2 \
  --per_device_train_batch_size 2 \
  --gradient_accumulation_steps 1 \
  --warmup_ratio 0.1 \
  --weight_decay 0.01 \
  --gradient_checkpointing
```

**Use bf16 not fp16** — Blackwell does bf16 natively. **Volume target:** 20k triples.

## Execution order

1. **Track C (Embeddings)** first — independent of generator, smallest, lowest risk; ETA same day.
2. **Track D (Reranker)** second — same hardware, no conflict. ETA same day.
3. **Track A1 (SFT)** third — the big one. ETA ~48 h with eval at end-of-epoch-1.
4. **Track B (Extraction)** fourth. ETA ~30 h.
5. **Track A2 (DPO)** fifth — needs Track A1's merged checkpoint. ETA ~18 h.

**Total wall-clock (sequential):** ~6 days first-try. Realistic with debug loops: ~10 days.

## What we are NOT doing (and why)

- **No QLoRA on Track A/B.** With 384 GB VRAM, bf16 LoRA is faster and avoids the bnb Blackwell perf cliff.
- **No router LoRA.** Risk of expert-assignment collapse outweighs any conceivable upside.
- **No Axolotl.** Force-downgrades torch and loses Blackwell support that works.
- **No DeepSpeed ZeRO-3.** FSDP2 is the proven MoE path; ZeRO-3 + MoE has expert-sharding instabilities.
- **No fine-tuned embeddings on Track A's data.** Architecture decision stands: BGE-M3 is the right base for embeddings.
- **No training during Charlotte's working hours** (08:00–18:00 BST) until the new gpt-oss-120b adapter is qualified.

## Source citations

- [openai/gpt-oss-120b model card](https://huggingface.co/openai/gpt-oss-120b)
- [Unsloth gpt-oss tutorial](https://unsloth.ai/docs/models/gpt-oss-how-to-run-and-fine-tune/tutorial-how-to-fine-tune-gpt-oss)
- [Axolotl gpt-oss examples](https://github.com/axolotl-ai-cloud/axolotl/tree/main/examples/gpt-oss)
- [Dao-AILab/flash-attention#1987](https://github.com/Dao-AILab/flash-attention/issues/1987) — FA3 Blackwell non-support
- [bitsandbytes#1851](https://github.com/bitsandbytes-foundation/bitsandbytes/issues/1851) — Blackwell NF4 perf penalty
- [FlagEmbedding BGE-M3 fine-tune README](https://github.com/FlagOpen/FlagEmbedding/blob/master/examples/finetune/embedder/README.md)
- [Chen et al. 2024 — BGE-M3](https://arxiv.org/html/2402.03216v3)

## Related notes

- [[Multi-Model-Architecture-2026-05-23]] — architectural decision this plan implements
- [[Research-gpt-oss-120b-capability-fit-2026-05-23]] — capability research that motivated the fine-tune
- [[Research-Action-Playbook-2026-05-23]] — the playbook this plan supplies the "how" for action 2
- [[Finetuning-Plan-2026-05-22]] — earlier high-level plan; this doc is its concrete execution layer
- [[Pilot-Results-2026-05-22]] — pilot eval that defines the success bar
- [[2026-05-23-missed-authority-EWHC-Ch-1031]] — Charlotte's incident that motivates Track C/D corpus-coverage urgency
- [[Fine-Tune-Execution-Next-Steps-2026-05-23]] — execution runbook
- [[Licensing-Considerations-2026-05-23]] — licensing review (training data + base models)
