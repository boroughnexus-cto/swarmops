-- 2026-05-23: Tighten ETL eligibility + add rejected-signal stream
--
-- After Charlotte's item-11 REJECT decision (100-char correction on a
-- 5000-char model output, 1.3% length ratio) we discovered the prior
-- eligibility predicate was too permissive: any silver/reject + any non-empty
-- correction got routed through the DPO-triple path. Fragmentary corrections
-- produced janky DPO `chosen` examples via the heuristic_inline append fallback.
--
-- This migration:
--
--   * Tightens `eligible_for_training` so fragmentary corrections fall off.
--     Requires either explicit non-fragmentary scope OR length proportion
--     >= 10% of the model output.
--   * Adds `eligible_as_rejected_signal` for silver/reject rows that DON'T
--     qualify as full triples — these are still useful as the rejected side
--     of a DPO pair (the model output is bad even if we lack a chosen).
--     Mutually exclusive with `eligible_for_training` (per peer review:
--     prevents implicit routing contract in TypeScript ETL).
--   * Wraps length expressions with COALESCE for NULL safety on missing
--     model outputs (per peer review).
--   * Exposes correction_length + model_output_length as debug columns so
--     callers can see why the heuristic classified a row the way it did.
--
-- Threshold (0.10 = 10%) empirically validated against Charlotte's first
-- four triage decisions:
--   item 1  (no correction)         : 0%    → rejected_signal_only
--   item 3  (section-scoped, silver): 14.3% → eligible_for_training ✓
--   item 6  (gold)                  : N/A   → eligible_for_training (gold path)
--   item 11 (fragmentary, reject)   : 1.3%  → rejected_signal_only ✓
--
-- Threshold is tunable; revisit after ≥30 corrections.
--
-- Migration is `CREATE OR REPLACE VIEW`. To prevent silent regression if
-- this PR merges before PR #25 (which introduces the `correction_scope`
-- column this view references), we assert the column exists up-front.

BEGIN;

----------------------------------------------------------------------------
-- 1. Prerequisite guard: assert PR #25's correction_scope column exists.
--    Fails loudly if applied before the prior migration. (Per peer-review:
--    CREATE OR REPLACE VIEW would otherwise succeed and the column would
--    look missing only at the ETL script's downstream consumer.)
----------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'training_triage_decisions'
      AND column_name = 'correction_scope'
  ) THEN
    RAISE EXCEPTION
      'training_triage_decisions.correction_scope column is missing — '
      'apply migration 2026_05_23_correction_scope_and_etl_view.sql first '
      '(PR #25). This migration depends on the column existing.';
  END IF;
END$$;

----------------------------------------------------------------------------
-- 2. Replace the view from PR #25 with the tightened version.
--    Two eligibility flags (mutually exclusive) + a couple of debug columns
--    so SQL analysts can audit the classification.
----------------------------------------------------------------------------

CREATE OR REPLACE VIEW public.training_eligible_for_etl
WITH (security_invoker = true)
AS
WITH base AS (
  SELECT
    d.*,
    q.pilot_run_id,
    q.sample_id,
    q.stratum,
    q.question,
    q.payload                                              AS queue_payload,
    q.areas_of_law                                         AS queue_areas_of_law,
    q.status                                               AS queue_status,
    COALESCE(length(d.barrister_correction), 0)            AS correction_length,
    COALESCE(length(q.payload->>'output'), 0)              AS model_output_length
  FROM public.training_triage_decisions d
  JOIN public.training_queue_items q
    ON q.id = d.queue_item_id
),
classified AS (
  SELECT
    *,
    -- Tightened: gold is always eligible; silver/reject only if there's a
    -- substantial correction OR an explicit non-fragmentary scope.
    CASE
      WHEN queue_status <> 'active' THEN false
      WHEN decision = 'gold' THEN true
      WHEN decision IN ('silver','reject')
           AND correction_length > 0
           AND (
                correction_scope IN ('full','section','annotation')
                OR
                -- Length-proportion heuristic (10%; tunable).
                -- GREATEST(model_output_length, 1) avoids division-by-zero;
                -- COALESCE in `base` neutralises NULL outputs.
                correction_length::float
                  >= 0.10 * GREATEST(model_output_length, 1)
              )
      THEN true
      ELSE false
    END AS eligible_for_training
  FROM base
)
SELECT
  id                                                         AS decision_id,
  queue_item_id,
  user_email,
  decision,
  reasons,
  areas_of_law                                               AS decision_areas_of_law,
  note,
  barrister_correction,
  correction_scope,
  created_at                                                 AS decision_created_at,
  updated_at                                                 AS decision_updated_at,
  pilot_run_id,
  sample_id,
  stratum,
  question,
  queue_payload,
  queue_areas_of_law,
  queue_status,
  correction_length,
  model_output_length,
  eligible_for_training,
  -- Mutually exclusive with eligible_for_training: silver/reject rows that
  -- did NOT qualify as full triples are still rejected-signal candidates.
  -- The ETL script consumes both flags independently — no implicit routing
  -- contract via TypeScript if/else ordering.
  CASE
    WHEN queue_status <> 'active' THEN false
    WHEN decision IN ('silver','reject') AND NOT eligible_for_training THEN true
    ELSE false
  END                                                        AS eligible_as_rejected_signal
FROM classified;

COMMENT ON VIEW public.training_eligible_for_etl IS
$$Joins barrister decisions with their source queue items + computes two
mutually-exclusive eligibility flags:

  eligible_for_training        — gold rows, OR silver/reject with a
                                  substantial correction (scope in
                                  full/section/annotation, OR length >= 10%
                                  of model output).
  eligible_as_rejected_signal  — silver/reject rows that did NOT qualify as
                                  full training pairs. Useful as the
                                  rejected-side of a DPO pair where the
                                  chosen side comes from synthesis (future
                                  enhancement) or is held back from training.

The flags do NOT overlap by construction. A row qualifies for AT MOST ONE
training route. Both flags compose with `queue_status = 'active'`.

security_invoker = true so the view inherits the caller's RLS — barristers
see only their own decisions; admins see all. The ETL script authenticates
with the service-role key and therefore bypasses RLS entirely.

Threshold (0.10) is tunable. Revisit after ≥30 corrections; currently
empirically calibrated against Charlotte's first 4 triage decisions where
item 3 sits at 14.3% (must qualify) and item 11 sits at 1.3% (must not).$$;

-- Re-grant (CREATE OR REPLACE preserves grants but re-stating is harmless).
GRANT SELECT ON public.training_eligible_for_etl TO authenticated;

COMMIT;
