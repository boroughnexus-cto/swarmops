import { describe, expect, it } from 'vitest'
import {
  applyCorrection,
  harmonyPromptOnly,
  harmonyWrap,
  mergeBySectionHeader,
} from '../etl-pure'

const SYSTEM = 'You are a UK barrister assistant.'

const FOUR_SECTION_OUTPUT = `## Overview
Some overview text.

## Stage 1
First-stage text with details.

## Stage 2
Second-stage text with the original (slightly off) explanation.

## Stage 3
Third-stage text.`

describe('harmonyWrap', () => {
  it('emits the canonical system + user + assistant-final sequence ending in return', () => {
    const out = harmonyWrap(SYSTEM, 'What is X?', 'X is Y.')
    expect(out).toBe(
      '<|start|>system<|message|>You are a UK barrister assistant.<|end|>' +
        '<|start|>user<|message|>What is X?<|end|>' +
        '<|start|>assistant<|channel|>final<|message|>X is Y.<|return|>',
    )
  })

  it('handles multi-line system + user + assistant content without escaping', () => {
    const out = harmonyWrap('sys\nline2', 'q\nline2', 'a\nline2')
    expect(out).toContain('<|message|>sys\nline2<|end|>')
    expect(out).toContain('<|message|>q\nline2<|end|>')
    expect(out).toContain('<|message|>a\nline2<|return|>')
  })
})

describe('harmonyPromptOnly', () => {
  it('emits prompt that ends at the final-channel message-open, no return token', () => {
    const out = harmonyPromptOnly(SYSTEM, 'q')
    expect(out.endsWith('<|return|>')).toBe(false)
    expect(out.endsWith('<|message|>')).toBe(true)
    expect(out).toBe(
      '<|start|>system<|message|>You are a UK barrister assistant.<|end|>' +
        '<|start|>user<|message|>q<|end|>' +
        '<|start|>assistant<|channel|>final<|message|>',
    )
  })
})

describe('mergeBySectionHeader', () => {
  it('replaces the matching section, preserving siblings', () => {
    const correction = `## Stage 2
Replaced text with proper citations.`
    const result = mergeBySectionHeader(FOUR_SECTION_OUTPUT, correction)
    expect(result).toContain('## Overview')
    expect(result).toContain('## Stage 1')
    expect(result).toContain('Replaced text with proper citations.')
    expect(result).not.toContain('Second-stage text with the original')
    expect(result).toContain('## Stage 3')
  })

  it('appends correction when no header in the correction text', () => {
    const correction = 'Just a paragraph with no header markup.'
    const result = mergeBySectionHeader(FOUR_SECTION_OUTPUT, correction)
    expect(result.endsWith('Just a paragraph with no header markup.\n')).toBe(true)
    // Original output preserved in full
    expect(result).toContain('## Overview')
    expect(result).toContain('## Stage 3')
  })

  it('appends correction when correction header does not match any section', () => {
    const correction = `## Stage 42
This section does not exist in the output.`
    const result = mergeBySectionHeader(FOUR_SECTION_OUTPUT, correction)
    expect(result).toContain('## Stage 42')
    expect(result).toContain('Second-stage text with the original')
  })

  it('escapes special regex characters in header titles', () => {
    const output = `## Stage (1.0) — first
First-stage text.

## Stage (2.0) — second
Second-stage text.`
    const correction = `## Stage (2.0) — second
Replaced.`
    const result = mergeBySectionHeader(output, correction)
    expect(result).toContain('First-stage text.')
    expect(result).toContain('Replaced.')
    expect(result).not.toContain('Second-stage text.')
  })

  it('stops the section boundary at the next same-or-higher level header', () => {
    const output = `## Outer
Outer body.

### Inner
Inner body.

## Next
Next body.`
    const correction = `## Outer
Replaced outer + inner.`
    const result = mergeBySectionHeader(output, correction)
    expect(result).toContain('Replaced outer + inner.')
    expect(result).toContain('## Next')
    expect(result).not.toContain('Outer body.')
  })
})

describe('applyCorrection', () => {
  it('scope=full returns the correction verbatim', () => {
    const { chosen, scopeApplied } = applyCorrection(FOUR_SECTION_OUTPUT, 'replacement', 'full')
    expect(chosen).toBe('replacement')
    expect(scopeApplied).toBe('full')
  })

  it('scope=annotation also returns the correction verbatim', () => {
    const { chosen, scopeApplied } = applyCorrection(
      FOUR_SECTION_OUTPUT,
      'annotated full output',
      'annotation',
    )
    expect(chosen).toBe('annotated full output')
    expect(scopeApplied).toBe('annotation')
  })

  it('scope=section invokes header-match merge', () => {
    const { chosen, scopeApplied } = applyCorrection(
      FOUR_SECTION_OUTPUT,
      `## Stage 2\nReplaced.`,
      'section',
    )
    expect(scopeApplied).toBe('section')
    expect(chosen).toContain('Replaced.')
    expect(chosen).toContain('## Stage 1')
  })

  it('scope=inline appends with double-newline separator', () => {
    const { chosen, scopeApplied } = applyCorrection(
      FOUR_SECTION_OUTPUT,
      'A new fragment to append.',
      'inline',
    )
    expect(scopeApplied).toBe('inline')
    expect(chosen.endsWith('A new fragment to append.\n')).toBe(true)
    expect(chosen).toContain('## Overview')
  })

  it('NULL scope + matching header → heuristic_section', () => {
    const { chosen, scopeApplied } = applyCorrection(
      FOUR_SECTION_OUTPUT,
      `## Stage 2\nReplaced.`,
      null,
    )
    expect(scopeApplied).toBe('heuristic_section')
    expect(chosen).toContain('Replaced.')
    expect(chosen).toContain('## Stage 1')
  })

  it('NULL scope + no matching header + large correction → heuristic_full', () => {
    const big = `# Brand new answer\n` + 'x'.repeat(FOUR_SECTION_OUTPUT.length)
    const { chosen, scopeApplied } = applyCorrection(FOUR_SECTION_OUTPUT, big, null)
    expect(scopeApplied).toBe('heuristic_full')
    expect(chosen).toBe(big)
  })

  it('NULL scope + no matching header + small correction → heuristic_inline', () => {
    const tiny = 'A small fragment.'
    const { chosen, scopeApplied } = applyCorrection(FOUR_SECTION_OUTPUT, tiny, null)
    expect(scopeApplied).toBe('heuristic_inline')
    expect(chosen).toContain('A small fragment.')
    expect(chosen).toContain('## Overview')
  })

  it('NULL scope + small correction with a header that matches → heuristic_section (not heuristic_inline)', () => {
    // Charlotte item-3-style case: short correction but its header matches one in the output.
    const correction = `## Stage 1\nRevised stage 1.`
    const { chosen, scopeApplied } = applyCorrection(FOUR_SECTION_OUTPUT, correction, null)
    expect(scopeApplied).toBe('heuristic_section')
    expect(chosen).toContain('Revised stage 1.')
    expect(chosen).not.toContain('First-stage text with details.')
  })
})
