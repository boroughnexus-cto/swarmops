#!/usr/bin/env bash
# Sequential overnight: Phase A (DPO pairs) then Phase B (seed bank).
# Logs to ~/swarmops/.overnight/overnight_runner.log.

set -uo pipefail
cd ~/swarmops/.overnight

export SUPABASE_SERVICE_ROLE_KEY='eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic2VydmljZV9yb2xlIiwiaXNzIjoic3VwYWJhc2UiLCJpYXQiOjE3NzkzNTgwNTQsImV4cCI6MTkzNzAzODA1NH0.GawmTPyyleR4m3bCY24sW5z4tRjzTyIwCV97ZtIubjI'
export SUPABASE_URL='https://auth.briefbox.work'

echo "[$(date -Iseconds)] === overnight job starting ==="

echo "[$(date -Iseconds)] PHASE A: frontier DPO pairs"
python3 frontier_dpo_pairs.py
echo "[$(date -Iseconds)] Phase A exit: $?"

echo ""
echo "[$(date -Iseconds)] PHASE B: frontier seed bank"
python3 frontier_seed_gen.py
echo "[$(date -Iseconds)] Phase B exit: $?"

echo "[$(date -Iseconds)] === overnight job complete ==="
echo ""
echo "Artefacts:"
echo "  $(wc -l dpo_pairs_results.jsonl 2>/dev/null) — DPO pairs"
echo "  $(wc -l seed_bank_expansion.jsonl 2>/dev/null) — new seeds"
