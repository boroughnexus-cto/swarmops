/**
 * Pure transform functions used by the corrections-to-JSONL ETL.
 *
 * Lives in src/lib/training/ rather than scripts/lib/ so the existing vitest
 * config picks them up without extending the `include` glob. These functions
 * are pure (no I/O, no side effects, no Node-only deps) and tree-shake out of
 * the Next.js client bundle — verified by an ESLint no-restricted-imports
 * rule preventing src/app/** and src/components/** from depending on this
 * module.
 *
 * If anything in this file ever needs to grow a Node-only dep (fs / path /
 * @supabase/supabase-js / etc.), move the whole module to scripts/lib/ and
 * extend vitest config.
 */

export type CorrectionScope = 'full' | 'section' | 'annotation' | 'inline' | null

/** Possible scope-applied labels — concrete strings only, used as object keys
 *  for the scopeStats counter in the ETL script. */
export type ScopeApplied =
  | 'full'
  | 'section'
  | 'annotation'
  | 'inline'
  | 'heuristic_section'
  | 'heuristic_full'
  | 'heuristic_inline'

/**
 * Render the gpt-oss-120b Harmony chat template manually. We don't apply the
 * tokenizer here — the Track A training script does that at load time via
 * `AutoTokenizer.apply_chat_template`. This produces the canonical text form
 * so the downstream training script can re-tokenise without surprises.
 *
 * SFT loss masks to the `final` channel only, per Fine-Tune-Recipe-2026-05-23.
 */
export function harmonyWrap(
  systemPrompt: string,
  user: string,
  assistantFinal: string,
): string {
  return (
    `<|start|>system<|message|>${systemPrompt}<|end|>` +
    `<|start|>user<|message|>${user}<|end|>` +
    `<|start|>assistant<|channel|>final<|message|>${assistantFinal}<|return|>`
  )
}

/** Same template but with no assistant content — used for the DPO `prompt`
 *  field where chosen/rejected are emitted separately. */
export function harmonyPromptOnly(systemPrompt: string, user: string): string {
  return (
    `<|start|>system<|message|>${systemPrompt}<|end|>` +
    `<|start|>user<|message|>${user}<|end|>` +
    `<|start|>assistant<|channel|>final<|message|>`
  )
}

/**
 * Replace the section of `modelOutput` whose markdown header matches the
 * first header found in `correction`. Returns an appended-correction result
 * if no header matches.
 *
 * Section boundary: same-or-higher-level header marks the end. For example,
 * a level-2 (`##`) header section ends at the next `#` or `##`, not at the
 * next `###`.
 */
export function mergeBySectionHeader(modelOutput: string, correction: string): string {
  const headerMatch = correction.match(/^(#{1,6})\s+(.+)$/m)
  if (!headerMatch) {
    return `${modelOutput.trimEnd()}\n\n${correction.trim()}\n`
  }
  const [, hashes, title] = headerMatch
  const level = hashes!.length
  const escapedTitle = title!.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

  // End-of-input anchor: JavaScript regex has no \z. Use `$(?![\s\S])` —
  // `$` in multiline mode is end-of-line; the negative lookahead asserts no
  // character follows, which only holds at true end-of-input. (Phase-4 fix
  // after the special-regex-chars test surfaced the bug — without it,
  // sections that ran to end-of-document weren't matched at all.)
  const sectionRegex = new RegExp(
    `^#{${level}}\\s+${escapedTitle}[ \\t]*$[\\s\\S]*?` +
      `(?=^#{1,${level}}\\s|$(?![\\s\\S]))`,
    'gm',
  )
  const replaced = modelOutput.replace(sectionRegex, () => correction.trim())
  // If the regex didn't match anything, fall back to appending.
  if (replaced === modelOutput) {
    return `${modelOutput.trimEnd()}\n\n${correction.trim()}\n`
  }
  return replaced
}

/**
 * Resolve the `chosen` text for a silver/reject row given the original model
 * output and the barrister's correction. Behaviour depends on correction_scope:
 *
 *   full        → correction replaces the whole output
 *   section     → correction is spliced into the output by header match
 *   annotation  → correction IS the full output with the barrister's edits;
 *                  treat as full replacement
 *   inline      → correction is appended to the output
 *   NULL        → heuristic: header-match if title found in output;
 *                  else full-replace if correction is similar size;
 *                  else inline append
 */
export function applyCorrection(
  modelOutput: string,
  correction: string,
  scope: CorrectionScope,
): { chosen: string; scopeApplied: ScopeApplied } {
  switch (scope) {
    case 'full':
    case 'annotation':
      return { chosen: correction, scopeApplied: scope }
    case 'section':
      return { chosen: mergeBySectionHeader(modelOutput, correction), scopeApplied: 'section' }
    case 'inline':
      return {
        chosen: `${modelOutput.trimEnd()}\n\n${correction.trim()}\n`,
        scopeApplied: 'inline',
      }
    case null: {
      const headerMatch = correction.match(/^(#{1,6})\s+(.+)$/m)
      if (headerMatch) {
        const escapedTitle = headerMatch[2]!.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
        const titleInOutput = new RegExp(`^#{1,6}\\s+${escapedTitle}`, 'm')
        if (titleInOutput.test(modelOutput)) {
          return {
            chosen: mergeBySectionHeader(modelOutput, correction),
            scopeApplied: 'heuristic_section',
          }
        }
      }
      // No matching header. If correction length is close to model output,
      // treat as full replacement; otherwise inline append.
      if (correction.length > modelOutput.length * 0.5) {
        return { chosen: correction, scopeApplied: 'heuristic_full' }
      }
      return {
        chosen: `${modelOutput.trimEnd()}\n\n${correction.trim()}\n`,
        scopeApplied: 'heuristic_inline',
      }
    }
  }
}
