import { describe, expect, it } from 'vitest'
import type { DeliberationCase } from '../types/review'
const needsChair = (item: DeliberationCase) => !item.decision && (item.preliminary.vetoed || item.divergences.length > 0)
describe('deliberation queue', () => {
  it('keeps a vetoed high-score material in the chair queue', () => {
    const item = { decision: undefined, divergences: [], preliminary: { vetoed: true, passed: false, weightedScore: 100 } } as unknown as DeliberationCase
    expect(needsChair(item)).toBe(true)
  })
})
