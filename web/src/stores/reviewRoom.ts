import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { reviewGateway } from '../services/reviewGateway'
import type { DeliberationCase, ReviewStatistics } from '../types/review'

export const useReviewRoom = defineStore('review-room', () => {
  const batchID = ref('batch-1')
  const revision = ref(2)
  const cases = ref<DeliberationCase[]>([])
  const statistics = ref<ReviewStatistics | null>(null)
  const loading = ref(false)
  const error = ref('')
  const unresolved = computed(() => cases.value.filter(item => !item.decision && (item.preliminary.vetoed || item.divergences.length > 0)))
  const quorumMissing = computed(() => cases.value.filter(item => !item.quorumMet))
  async function refresh() { loading.value = true; error.value = ''; try { cases.value = await reviewGateway.deliberation(batchID.value) } catch (cause) { error.value = cause instanceof Error ? cause.message : '议事单加载失败' } finally { loading.value = false } }
  async function loadStatistics() { try { statistics.value = await reviewGateway.statistics() } catch (cause) { error.value = cause instanceof Error ? cause.message : '统计加载失败' } }
  return { batchID, revision, cases, statistics, loading, error, unresolved, quorumMissing, refresh, loadStatistics }
})
