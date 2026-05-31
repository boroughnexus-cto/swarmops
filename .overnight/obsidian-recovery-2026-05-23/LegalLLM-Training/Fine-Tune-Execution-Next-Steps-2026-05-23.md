---
classification: non-sensitive
type: execution-runbook
project: LegalLLM-Training / BriefBox
created: '2026-05-23'
status: parked-ready-to-resume
recovered: '2026-05-23 — vault auto-sync deleted this file; reconstructed from conversation transcript'
tags:
- fine-tuning
- runbook
- briefbox
- molecule-air
- next-session
---

> **NON-SENSITIVE — execution runbook for resuming the 4-track fine-tune programme. Picks up where [[Fine-Tune-Recipe-2026-05-23]] (the recipe doc) leaves off. Read this first when coming back to the project.**

# Fine-Tune Execution — Next Steps (parked 2026-05-23)

## Where we are right now

| Item | Status |
|---|---|
| Hardware audited (4× RTX PRO 6000 Blackwell, 384 GB VRAM, 251 GB RAM, sm_120) | ✅ Done |
| Training venv at `/data/train-venv` on molecule-air (torch 2.9.1+cu128, transformers 5.9.0, trl 1.4.0, peft 0.19.1, FlagEmbedding 1.4.0, kernels + liger-kernel) | ✅ Done, smoke-imported |
| Models cached: gpt-oss-120b, BGE-M3, BGE-reranker-v2-m3 | ✅ Done |
| 73,500 seed bank backed up to Unraid + local at `~/swarmops/.overnight/` | ✅ Done |
| Recipe doc: per-track LoRA configs, hyperparameters, eval gates | ✅ `Fine-Tune-Recipe-2026-05-23` |
| Licensing review + research-scope decision | ✅ `Licensing-Considerations-2026-05-23` — seed bank approved for use under research-project scope |
| **Training data prepared as JSONL** | ❌ Not started |
| **Smoke-test of training rig (transformers 5.x + TRL 1.4 + FSDP2 on Blackwell)** | ❌ Not started |
| **First production training run** | ❌ Not started |

## Three entry-point options

### Option 1 — Data prep first (no production downtime)

**Build all four training-data JSONLs before touching the training rig.** No GPU contention, no inference downtime, ~2-4 hours of pure CPU/IO work. Output: 4 ready-to-train JSONLs at `/data2/training/data/`. Risk: if the training rig later turns out to be broken (transformers 5.x or TRL 1.4 API breakage we haven't validated), the data work is still useful but we lose the smoke-test-first signal.

### Option 2 — Smoke-test the rig first (needs ~1h production downtime)

**Stop the 3× gpt-oss-120b inference replicas, run a 100-seed Track A LoRA smoke test on 4× Blackwell.** Validates that transformers 5.x + TRL 1.4 + FSDP2 + bf16 LoRA actually trains end-to-end on these specific cards before sinking time into data prep we may not be able to use. ~1h total wall-clock including inference bring-down + bring-up. Charlotte needs to be off-tool for the window.

### Option 3 — Both in parallel

**Kick off Track A Harmony wrap on swarmops (CPU-only, ~30 min)** while the smoke test runs on molecule-air. Most efficient use of wall-clock but requires production downtime + parallel monitoring of two workstreams.

**Recommended:** Option 3 if the inference-downtime window is available; Option 1 if not.

## Execution checklist — Option 1 (Data prep)

### Step 1A — Harmony-wrap the 73k seeds for Track A SFT

The seed bank lives at `~/swarmops/.overnight/seed_bank_expansion{,_gpt,_sonnet}.jsonl` (the Minimax `_minimax` file stays quarantined for confabulation reasons). Each row needs to be wrapped in gpt-oss-120b's Harmony chat format with the loss masked to the `<|channel|>final` segment only.

Skeleton (write to `~/swarmops/.overnight/scripts/harmony_wrap.py`):

```python
import json
from pathlib import Path
from transformers import AutoTokenizer

SRC = [
    Path("~/swarmops/.overnight/seed_bank_expansion.jsonl").expanduser(),         # Gemini
    Path("~/swarmops/.overnight/seed_bank_expansion_gpt.jsonl").expanduser(),     # GPT-5.5
    Path("~/swarmops/.overnight/seed_bank_expansion_sonnet.jsonl").expanduser(),  # Sonnet
]
DST = Path("/data2/training/data/sft_legal_v1.jsonl")  # on molecule-air after rsync
TOK = AutoTokenizer.from_pretrained("/data3/models/gpt-oss-120b", trust_remote_code=True)

SYSTEM = (
    "You are a UK barrister's research assistant. Answer English commercial-litigation "
    "questions with OSCOLA-format citations. When you cannot answer, say so and explain why."
)

def wrap(seed):
    user = seed["question"]
    assistant = seed["answer"] if "answer" in seed else seed.get("model_answer", "")
    msgs = [
        {"role": "system", "content": SYSTEM},
        {"role": "user", "content": user},
        {"role": "assistant", "content": assistant},
    ]
    text = TOK.apply_chat_template(msgs, tokenize=False, add_generation_prompt=False)
    return {"text": text, "meta": {"category": seed.get("category"), "source": seed.get("_source")}}

with DST.open("w") as out:
    for src in SRC:
        for line in src.open():
            seed = json.loads(line)
            seed["_source"] = src.stem
            out.write(json.dumps(wrap(seed)) + "\n")
```

**Validation step before running:** `head -1 ~/swarmops/.overnight/seed_bank_expansion.jsonl | jq .` — field names may differ; check `question` vs `prompt`, `answer` vs `ideal_response`, etc.

**Rsync to molecule-air:** `rsync -avh /data2/training/data/sft_legal_v1.jsonl rbadmin@100.84.236.70:/data2/training/data/`.

### Step 1B — Build Track C/D triples (embedding + reranker)

Both tracks share data shape `{query, pos:[...], neg:[...]}`. One pipeline, two outputs.

**Query sources:**
1. **Real Charlotte queries** from the Briefbox `feedback_reports` Supabase table — highest-signal queries. Weight 5× in the training mix.
2. **Synthetic queries** generated from `corpus_documents.headnote` via self-hosted gpt-oss-120b (already deployed on port 30002/3/4). Target ~30k synthetic queries for Track C, ~20k for Track D.

**Hard negative mining:**

```bash
source /data/train-venv/bin/activate
python -m FlagEmbedding.scripts.hn_mine \
  --input_file /data2/training/data/qd_pairs_raw.jsonl \
  --output_file /data2/training/data/embed_pairs_v1.jsonl \
  --range_for_sampling 2-200 \
  --negative_number 7 \
  --use_gpu_for_searching \
  --embedder_name_or_path /data/models/bge-m3
```

Track D uses the same JSONL (FlagEmbedding's reranker trainer accepts the same shape with `knowledge_distillation=False`).

### Step 1C — Build Track B extraction pairs

For each of ~20,000 random `corpus_documents` rows, run gpt-oss-120b twice with different extraction prompts and keep the pair where both runs agree on `neutral_citation` and `parties`. Estimated cost: 20k judgments × 2 generations × ~3000 tokens out × 3 cards ≈ 8-12 hours of gpt-oss-120b inference time.

## Execution checklist — Option 2 (Smoke test first)

### Step 2A — Stop production inference

**TODO before running:** confirm the management system. Check whether the SGLang processes are started by systemd, supervisord, manually via `nohup`, or via Komodo. **Until that's verified, do not kill processes directly.**

Services to stop:
- 3× gpt-oss-120b SGLang on ports 30002, 30003, 30004
- 1× MiniMax-M2.7 SGLang on port 30000 (tp=4)
- Keep model-proxy at port 8000 running

### Step 2B — Tell Charlotte the inference is down for the window

Slack/Telegram/whatever channel she watches. Set expectation: ~60-90 min outage.

### Step 2C — Prepare smoke-test data

```bash
head -100 /data2/training/data/sft_legal_v1.jsonl > /data2/training/data/sft_smoke_v1.jsonl
```

### Step 2D — Write the SFT training script

Skeleton at `/data2/training/scripts/sft_smoke.py`:

```python
# accelerate launch --use_fsdp --fsdp_version 2 --num_processes 4 sft_smoke.py
import torch
from datasets import load_dataset
from transformers import AutoModelForCausalLM, AutoTokenizer
from trl import SFTTrainer, SFTConfig
from peft import LoraConfig, TaskType

MODEL = "/data3/models/gpt-oss-120b"
DATA = "/data2/training/data/sft_smoke_v1.jsonl"
OUT = "/data3/training-outputs/A-smoke-2026-05-XX/"

tok = AutoTokenizer.from_pretrained(MODEL, trust_remote_code=True)
ds = load_dataset("json", data_files=DATA, split="train")
model = AutoModelForCausalLM.from_pretrained(
    MODEL, torch_dtype=torch.bfloat16,
    attn_implementation="sdpa", trust_remote_code=True,
)
lora = LoraConfig(
    r=32, lora_alpha=64,
    target_modules=["q_proj","k_proj","v_proj","o_proj","gate_proj","up_proj","down_proj"],
    lora_dropout=0.0, bias="none", task_type=TaskType.CAUSAL_LM,
)
cfg = SFTConfig(
    output_dir=OUT, num_train_epochs=1,
    per_device_train_batch_size=1, gradient_accumulation_steps=4,
    learning_rate=2e-5, lr_scheduler_type="cosine", warmup_ratio=0.03,
    bf16=True, gradient_checkpointing=True,
    logging_steps=5, save_steps=50, max_seq_length=4096,
    packing=True, dataset_text_field="text",
    report_to="wandb", run_name="A-smoke-2026-05-XX",
)
trainer = SFTTrainer(model=model, args=cfg, train_dataset=ds, peft_config=lora, processing_class=tok)
trainer.train()
trainer.save_model(OUT + "adapter/")
```

**Success criteria:** run completes in <60 min, loss decreasing, no OOM, no breaking-change errors, saved adapter loads cleanly.

### Step 2E — Bring production inference back up

Reverse of 2A. Spot-check that `model-proxy` is routing again.

## Once smoke test is green: actual production execution order

Per [[Fine-Tune-Recipe-2026-05-23]]:

1. **Track C (Embeddings, BGE-M3)** — fastest, lowest risk (~4-8h)
2. **Track D (Reranker, BGE-reranker-v2-m3)** — same hardware (~3-5h)
3. **Track A1 (gpt-oss-120b SFT)** — overnight, full 72k seeds, 2 epochs (~24-48h)
4. **Track B (gpt-oss-120b extraction)** (~24-36h)
5. **Track A2 (gpt-oss-120b DPO)** (~12-24h)

## Open pre-execution questions

| Question | Resolution path |
|---|---|
| Are the SGLang processes managed by systemd, supervisord, or manual? | `systemctl list-units \| grep sgl` first; if nothing, check `/etc/supervisor/conf.d/`; if nothing, ask rbadmin |
| Do we have the BBOX-37 crawler fix done before Track B? | Track B doesn't strictly need the crawler — reads `corpus_documents` directly. But corpus coverage gap (BBOX-36) limits Track B's training distribution. |
| W&B Pro/Enterprise seat in place? | Needed before any `report_to="wandb"` run. Check `wandb login` status on molecule-air. |
| Where exactly does the model-proxy point — does stopping a backend cause user-visible errors? | Test with one replica stopped before stopping all three |
| What does the actual seed schema look like — does the skeleton script need adjusting? | `head -1 ~/swarmops/.overnight/seed_bank_expansion.jsonl \| jq .` |

## Resume command — copy-paste this into next session

> "Resume the Briefbox fine-tune. Read `Projects/LegalLLM-Training/Fine-Tune-Execution-Next-Steps-2026-05-23.md`. We left off having parked at the entry-point decision. Pick Option [1 / 2 / 3] and execute."

## Related notes

- [[Fine-Tune-Recipe-2026-05-23]] — the recipe (LoRA configs, hyperparameters, eval gates) this runbook drives
- [[Licensing-Considerations-2026-05-23]] — licensing decision: research-scope, seeds approved
- [[Multi-Model-Architecture-2026-05-23]] — model-class decisions
- [[Research-Action-Playbook-2026-05-23]] — the broader 5-week sprint plan
- [[Pilot-Results-2026-05-22]] — earlier pilot whose seed schema we're now operationalising
