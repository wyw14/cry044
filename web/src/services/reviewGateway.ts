import type { DeliberationCase, FinalDecision, ReviewStatistics } from '../types/review'

type Success<T> = { data: T; request_id: string }
type Failure = { code?: string; message?: string; request_id?: string; field_errors?: Record<string, string> }
type ReviewCall = { resource: string; role?: string; method?: 'GET' | 'POST'; body?: unknown }

export class ReviewApiFault extends Error {
  constructor(public readonly code: string, message: string, public readonly requestId = '', public readonly fields: Record<string, string> = {}) { super(message) }
}

async function send<T>(call: ReviewCall): Promise<T> {
  const headers = new Headers({ Accept: 'application/json', 'Content-Type': 'application/json', 'X-Actor-ID': 'demo-coordinator', 'X-Actor-Role': call.role ?? 'coordinator' })
  const response = await fetch(`/api/v1${call.resource}`, { method: call.method ?? 'GET', headers, body: call.body === undefined ? undefined : JSON.stringify(call.body) })
  const document: Success<T> | Failure = await response.json()
  if (response.ok) return (document as Success<T>).data
  const failure = document as Failure
  throw new ReviewApiFault(failure.code ?? 'REVIEW_REQUEST_FAILED', failure.message ?? 'Review request failed', failure.request_id, failure.field_errors)
}

function batchPath(batchID: string, suffix: string) { return `/review-batches/${encodeURIComponent(batchID)}${suffix}` }

export const reviewGateway = {
  async deliberation(batchID: string, options: { vetoed?: boolean; page?: number } = {}) {
    const query = new URLSearchParams({ page: String(options.page ?? 1), page_size: '20', sort: 'material_id' })
    if (options.vetoed === true) query.append('vetoed', 'true')
    const board = await send<{ items: DeliberationCase[] }>({ resource: `${batchPath(batchID, '/deliberation')}?${query}` })
    return board.items
  },
  resolve(batchID: string, materialID: string, revision: number, input: { passed: boolean; conclusion: string; rationale: string }) {
    return send<FinalDecision>({ resource: batchPath(batchID, `/materials/${encodeURIComponent(materialID)}/final-decisions`), method: 'POST', body: { expected_revision: revision, ...input } })
  },
  statistics(window: { from?: string; to?: string } = {}) {
    const query = new URLSearchParams(); if (window.from) query.append('from', window.from); if (window.to) query.append('to', window.to)
    return send<ReviewStatistics>({ resource: `/review-statistics?${query}`, role: 'auditor' })
  },
}
