<script setup lang="ts">
import { computed, onMounted } from 'vue'
import SnapshotSeal from '../components/SnapshotSeal.vue'
import { useReviewRoom } from '../stores/reviewRoom'
const room = useReviewRoom()
onMounted(room.loadStatistics)
const maxReason = computed(() => Math.max(1, ...(room.statistics?.returnReasons.map(item => item.count) ?? [1])))
const versions = [
  { version: 4, status: 'published', name: '生产质量评议', criteria: 12, referenced: 8 },
  { version: 5, status: 'review', name: '生产质量评议', criteria: 14, referenced: 0 },
  { version: 2, status: 'deprecated', name: '供应商准入', criteria: 9, referenced: 21 },
]
</script>
<template><section><header class="chapter"><div><h1>审核标准谱系</h1><p>标准集、评分尺度、适用条件与历史引用保护</p></div><em>STANDARD / LINEAGE</em></header><div class="summary-grid"><div class="summary-card"><span>已发布版本</span><b>6</b></div><div class="summary-card"><span>历史快照引用</span><b>29</b></div><div class="summary-card"><span>平均审核周期</span><b>{{ room.statistics?.averageCycleHours.toFixed(1) ?? '—' }}h</b></div><div class="summary-card"><span>条目分歧</span><b>{{ room.statistics?.criteria.reduce((sum,item)=>sum+item.disagreementCount,0) ?? 0 }}</b></div></div><div class="split"><div class="ledger lineage"><article v-for="item in versions" :key="`${item.name}-${item.version}`"><SnapshotSeal :template="item.name" :version="item.version" :locked="item.status !== 'review'"/><div><strong>{{ item.criteria }} 个审核条目</strong><small>状态 {{ item.status }} · 历史引用 {{ item.referenced }}</small></div><button class="ink-button">查看谱系</button></article></div><aside class="paper"><h2>常见退回原因</h2><div v-for="item in room.statistics?.returnReasons ?? []" :key="item.reason" class="bar"><span>{{ item.reason }}</span><i :style="{width:`${item.count/maxReason*100}%`} "/><b>{{ item.count }}</b></div><p v-if="!room.statistics?.returnReasons.length">暂无统计数据</p></aside></div></section></template>
