---
classification: non-sensitive
type: rubric
project: LegalLLM-Training / BriefBox-UI
created: '2026-05-23'
status: implemented-pending-real-world-application
implementation-pr: https://github.com/BoroughNexus/BriefBox-UI/pull/21
recovered: '2026-05-23 — vault auto-sync deleted this file; reconstructed from conversation transcript'
tags:
- triage
- rubric
- charlotte
- briefbox-ui
- ai-training
- grading
---

> **NON-SENSITIVE — the gold/silver/reject grading rubric used at `/training` in BriefBox-UI. Authoritative source for the in-app crib sheet and for Charlotte's onboarding to the triage workflow. No client data; all examples are illustrative.**

# Triage Rubric — Gold · Silver · Reject (2026-05-23)

## The litmus test (one sentence)

> **"Could I put my name on this if I were rushing?"**

- **YES** → **GOLD**
- **Yes, after one specific fix** → **SILVER**
- **No, even with edits, this would mislead a recipient about current English law** → **REJECT**

Wording confirmed by Charlotte on 2026-05-23. The shift from *"would I"* to *"could I"* is deliberate: it frames the test as a **capability / competence floor** rather than a personal preference. The question isn't *"would I choose to sign this?"* — it's *"is this work of an acceptable professional standard, irrespective of whether I'd have phrased it that way?"*

The *rushing* qualifier matters. It encodes the production reality: the AI's value is time-saving, so the output must be safe enough that the barrister doesn't have to catch subtle errors to use it. "I could fix this in 10 minutes" is a silver answer. "This needs me to re-write from scratch" is a reject.

## Grade definitions

| Grade | Plain English | What you'd do with the output |
|---|---|---|
| **GOLD** | **Correct, competent and useful — even if not necessarily to your personal taste, style, or extreme level of detail.** Substantively right + correct current authority + appropriately framed + properly caveated. | Lift it straight into a brief, advice, pleading, or research note. No edits. |
| **SILVER** | One step from ship-ready. Right direction; needs one specific, named fix. Substance sound; polish incomplete. | Make the fix, then ship. The fix is mechanical (verify citation, add missed caveat, reframe a paragraph). |
| **REJECT** | Not safe even with edits. Wrong, misleading, or right-answer-to-wrong-question. Fixing it means writing a new answer. | Discard. Optionally provide the corrected version via the `barrister_correction` field — that becomes gold-tier training data. |

> **Gold guardrail (Charlotte, 2026-05-23):** *"Correct, competent and useful, even if not necessarily to personal taste/style/extreme level of detail."* Don't downgrade work to silver just because you'd have written it differently — downgrade only if the substance, framing, or completeness falls short of professional. This is the single most important guard against grade inflation toward silver.

## Critical principle: framing trumps substance

Refined per Charlotte's clarification (2026-05-23). The bright-line test is **not** *"is the cited test superseded?"* — it's **"would the recipient be misled about current English law?"**

The same substantive answer can be **gold, silver, or reject** depending on framing:

| Framing of a "right answer to a now-superseded test" | Grade |
|---|---|
| Presented **as if** it were current law | **REJECT** — recipient relies on bad law |
| Clearly flagged as superseded **and** used to contextualise older cases, explain doctrinal evolution, or answer a historical-facts question (e.g., the law as at the time of the contract / facts / events) | **GOLD** if useful + competent |
| Flagged as superseded but adds no value on *why the shift matters* or what current law now is | **SILVER** — right impulse, weak execution |
| User asked for the current-law answer and got a historical-test answer instead (however well-framed) | **REJECT** — didn't answer the question asked |

The principle generalises beyond superseded tests. Wherever the question is *"is X technically correct?"*, reframe it as *"would a recipient relying on this be misled?"* That is the test that protects professional reputation, not technical correctness in isolation.

Why this matters for the model:
- It teaches the model that **citing historical law is not the failure** — failing to flag that it's historical is the failure.
- It preserves the model's ability to be a useful research assistant for cases that turn on doctrinal evolution, pre-reform context, or the law as at the time of the events.
- It targets the actual harm: misleading the recipient.

## 21 grey areas — how to grade them

Grouped into six categories. Each entry is a situation that should make Charlotte pause; the recommended grade is the rubric's call, but the barrister's judgement always overrides.

### Category 1 — Right answer, wrong vintage

| # | Situation | Grade | Why |
|---|---|---|---|
| 1 | **Superseded leading case** — correct application of an overruled test (e.g., pre-reform *Wednesbury* unreasonableness). | **REJECT** if presented as current law; **SILVER** if flagged as superseded but adds no value on what current law now is; **GOLD** if clearly flagged and used to contextualise older cases, explain doctrinal evolution, or answer a historical-facts question. | See "Critical principle: framing trumps substance" above. The motivating example for this rubric. |
| 2 | **Statute amended** — correct citation of pre-amendment wording when post-amendment is now in force. | **REJECT** if the amendment changes the answer; **SILVER** if cosmetic and the principle holds. | Distinguishes substantive amendment from drafting tidy-ups. |
| 3 | **Pre-CPR procedural rules** — references RSC where CPR now governs. | **REJECT** | CPR is a complete code; RSC references are anachronistic in modern practice. |
| 4 | **Pre / post-Brexit EU law** — cites a CJEU decision as binding on English court post-2020. | **REJECT** if framed as binding; **SILVER** if explicitly flagged as "retained EU law" / "persuasive only" / "no longer binding". | Retained-EU-law nuance is real; framing is what matters. |

### Category 2 — Right principle, wrong authority

| # | Situation | Grade | Why |
|---|---|---|---|
| 5 | **Real case cited for a proposition it doesn't actually support** — the citation exists; the principle stated is wrong for that case. | **REJECT** | Recipient pulls the case and finds it doesn't say what's claimed. Trust-destroying. |
| 6 | **Hallucinated case** — plausible-looking citation that doesn't exist. | **REJECT** (zero tolerance) | The BSB takes a very dim view of fabricated authority. Single hallucinated case in a brief is professional-conduct territory. |
| 7 | **Misses the leading case** — answer is right but cites a minor authority when the famous one is expected. (E.g., negligence answer without *Donoghue v Stevenson*, constructive trust without *Westdeutsche*.) | **SILVER** if the cited authority is also good law and supports the proposition; **REJECT** if so off-piste that you lose confidence in the rest. | The leading case is what other practitioners will expect to see; missing it is bad shape but often not wrong. |
| 8 | **First-instance decision reversed on appeal** — correctly summarises the first-instance reasoning without noting the reversal. | **REJECT** | Same trust problem as #5 — recipient relies on the wrong outcome. |

### Category 3 — Court hierarchy / precedential weight confusion

| # | Situation | Grade | Why |
|---|---|---|---|
| 9 | **Obiter dressed as ratio** — famous obiter dictum presented as binding ratio. | **SILVER** if the proposition is sound and the answer is useful; **REJECT** if recipient would over-rely on precedential weight. | Whether it matters depends on whether the recipient is relying on the weight. |
| 10 | **Persuasive cited as binding** — Scottish, Northern Irish, Commonwealth, or US case framed as binding on English court. | **SILVER** | Often useful as comparative authority; framing fix is mechanical. |
| 11 | **Per incuriam not flagged** — case correctly summarised but decided per incuriam (without binding authority considered) and shouldn't be relied on. | **REJECT** | The whole point of the per incuriam doctrine is that the case isn't precedent. |
| 12 | **Lower-court endorsement cited as apex-settled** — Court of Appeal endorsement of first-instance reasoning treated as if Supreme Court has ruled. | **SILVER** | Often fine in practice; the caveat is one sentence. |

### Category 4 — Wrong jurisdiction / wrong branch of law

| # | Situation | Grade | Why |
|---|---|---|---|
| 13 | **US / foreign doctrine smuggled in** — e.g., imports US summary judgment standards into CPR Part 24, or US-style "discovery" into English disclosure. | **REJECT** | Foreign procedural concepts don't map; recipient applies the wrong test. |
| 14 | **Equity / common law conflation** — mixes equitable remedies with common-law remedies (e.g., calls common-law damages "specific performance", or treats equitable compensation as if it were tortious damages). | **REJECT** | Fundamental mis-framing; the recipient's analysis breaks. |
| 15 | **Civil vs criminal standard slippage** — applies "beyond reasonable doubt" to a civil claim; fails to flag the heightened evidentiary expectations for serious civil allegations (fraud, dishonesty). | **REJECT** if standard is wrong; **SILVER** if standard right but evidentiary nuance missed. | Standard-of-proof errors are a textbook professional failure. |

### Category 5 — Right answer, missing piece

| # | Situation | Grade | Why |
|---|---|---|---|
| 16 | **All required elements except one** — correctly states 3 of 4 elements of negligence; forgets causation. Or states all but one of the *American Cyanamid* limbs. | **SILVER** | Mechanical fix: add the missing element. The framing is right; the answer is incomplete. |
| 17 | **No application to facts** — correct statement of law, no analysis of how it applies to the user's question. | **SILVER** | Useful as starting material; the barrister will do the application. Drop a grade only if the user expressly asked for applied advice. |
| 18 | **Right test, missing caveat** — correct test stated, but fails to flag a contested area or a recent contrary authority. | **SILVER** if caveat is minor; **REJECT** if absence misleads about settled-ness of the law. | The line is whether the recipient would draw a different conclusion knowing the missing caveat. |
| 19 | **Overconfidence on contested law** — states as settled what's subject to a Court of Appeal disagreement or academic split. | **REJECT** if presented as settled; **SILVER** if answer is right on one side of the contest and acknowledges the other. | Pretending settled law on contested questions is the classic AI failure mode in legal QA. |

### Category 6 — Format / citation discipline

| # | Situation | Grade | Why |
|---|---|---|---|
| 20 | **Right substance, wrong OSCOLA format** — correct case and proposition, citation in wrong form (e.g., abbreviated "[2015] EWHC 3294 (Pat)" without party names or pinpoint when full OSCOLA expected). | **SILVER** | Mechanical fix. But persistent across the dataset = real signal for the model. |
| 21 | **Pinpoint paragraph wrong** — cites the case at [37] when the relevant para is [137]. | **SILVER** if the cited paragraph also supports the point (rare); **REJECT** if recipient goes to the wrong paragraph and finds nothing relevant. | Wrong pinpoint that misleads is functionally a hallucination. |

## How the rubric appears in the UI (Charlotte-facing)

1. **Per-screen reference link** in the triage form header: a small `?` icon next to "Decision" labelled *"Gold / Silver / Reject — what do these mean?"*.
2. **Modal crib sheet** opens on click. Top section: the one-sentence litmus test in large type. Middle: the three grade definitions. Bottom: a *contextual* slice of the 21 grey areas — only the categories relevant to the **areas-of-law tags** on the current sample (e.g., a `procedure` + `civil` sample shows Categories 1, 4, 6; a `contract` + `commercial` sample shows Categories 2, 3, 5).
3. **Dismissible per session** — Charlotte can close the modal and won't see it auto-open again, but the `?` icon stays available.
4. **First-time use** — the modal auto-opens the first time a new triager visits `/training/[id]`. Stored in `localStorage` (per-browser; no schema change needed).

## Why these 21 (and not more / fewer)

- They cover the failure modes seen across the 73,500-seed bank and Charlotte's first production query feedback (2026-05-22).
- They map cleanly onto the existing `reasons` enum in `training_triage_decisions` (citation_error, legal_error, tone_off, completeness, dangerous_helpfulness, format, other) — every grey area resolves to one or two reason tags.
- 21 is the cognitive ceiling for "list of grey areas a barrister can hold in working memory" — beyond that the rubric starts to function as case law that needs its own lookup. If we discover a new high-frequency failure mode in production, it gets a numbered slot here and the model retrains against the new label distribution.

## Mapping rubric grey areas → existing `reasons` enum

| Reason | Grey areas |
|---|---|
| `citation_error` | 5, 6, 8, 11, 20, 21 |
| `legal_error` | 1, 2, 3, 4, 13, 14, 15 |
| `tone_off` | (rarely triggered by these grey areas; more about voice/register) |
| `completeness` | 7, 16, 17, 18 |
| `dangerous_helpfulness` | 19 (overconfidence on contested law is the canonical case) |
| `format` | 20 |
| `other` | 9, 10, 12 (precedential-weight nuances don't have a dedicated reason yet — candidate for a `precedential_weight` reason in v2) |

## Open questions for Charlotte — status

- ✅ **Litmus phrasing:** answered 2026-05-23 — *"Could I put my name on this if I were rushing? (correct, competent & useful even if not to personal taste/style/extreme level of detail)"*. Folded into the litmus test sentence + the GOLD definition above.
- ✅ **Is REJECT right for "right answer to superseded test"?** — answered 2026-05-23: *"yes though it might depend... as long as you are clear that it has been superseded"*. Folded in as the new "framing trumps substance" principle and the multi-grade treatment of grey area #1.
- ✅ **Additional grey areas she hits in practice?** — sent full 21-area list with proposed grades on 2026-05-23. Charlotte's reply: *"Looks good in principle, will have to see how it is to apply in practice!"* Rubric locked provisionally; refine as real triage exposes patterns the 21 don't cover.
- ❓ **Should `precedential_weight` be a new reason tag in v2?** — our design decision, not for Charlotte. Hold open until we see how often grey areas 9–12 fire and whether triagers reach for `other` to flag them.

## Implementation (2026-05-23)

Live in BriefBox-UI on branch `feature/triage-crib-sheet` ([PR #21](https://github.com/BoroughNexus/briefbox-ui/pull/21)).

**Files:**
- `src/lib/training/triage-rubric.ts` — typed mirror of this doc. **Single source of truth at runtime.** Content edits MUST update both files together; structural-integrity tests catch drift.
- `src/components/training/TriageCribSheet.tsx` — Radix Dialog modal. Auto-opens on first visit (versioned localStorage flag `triage_crib_seen_v1` — bump the suffix when the rubric changes substantively to force a re-show). Triggered by a `?` icon next to the *Decision* label in `DecisionForm`.
- `src/lib/training/__tests__/triage-rubric.test.ts` — 22 unit tests including a prototype-key regression guard.
- `src/components/training/__tests__/TriageCribSheet.test.tsx` — 13 component tests covering auto-open lifecycle, contextual filtering, keyboard close, and private-mode resilience.

**Built via `/iteratecycle`** with `tkn-aipeer` peer review at both the plan stage (4 findings applied — split mapping out of rubric content, `AreaOfLaw` union type for compile-time completeness, removed `forceOpen` test prop, string-slug grey-area IDs) and the implementation stage (8 findings applied — including two HIGH prototype-chain bugs caught + regression tests). 213/213 tests pass.

**What's deliberately not implemented in v1:**
- No analytics on which categories Charlotte clicks through most (could inform v2 mapping refinements).
- No way to dismiss the modal permanently — by design; rubric changes should be visible. Use `localStorage.setItem('triage_crib_seen_v1', '1')` in browser console to mute it ad-hoc.
- No barrister-correction loop on the rubric itself (i.e., no "I'd grade this differently" button). Feedback for v2 comes via direct conversation, not in-app.

## Related

- [[AI-Training-Module-Spec-2026-05-22]] — the schema this rubric grades against
- [[Fine-Tune-Recipe-2026-05-23]] — the fine-tune that consumes the graded data
- [[Pilot-Results-2026-05-22]] — pilot whose Gate 2 measurement uses this rubric
- [[Harness-Peer-Review-2026-05-22]] — the peer review that motivated the `barrister_correction` field for upgrading reject→gold via correction capture
- BriefBox-UI repo: `~/git-bnx/BriefBox-UI/` — branch `feature/triage-crib-sheet`
