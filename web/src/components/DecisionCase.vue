<script setup lang="ts">
import type { DeliberationCase } from '../types/review'
defineProps<{ item: DeliberationCase }>()
defineEmits<{ open: [item: DeliberationCase] }>()
</script>
<template><article class="case-file" :data-alert="item.preliminary.vetoed || item.divergences.length > 0"><header><b>材料 {{ item.materialID }}</b><span>{{ item.submittedReviewers }}/{{ item.assignedReviewers }} 席已提交</span></header><div class="score"><strong>{{ item.preliminary.weightedScore.toFixed(1) }}</strong><small>聚合分</small></div><p v-if="item.preliminary.vetoed" class="veto">否决项已触发，分数不能覆盖</p><p v-else-if="item.divergences.length">{{ item.divergences.length }} 个条目需要议事说明</p><p v-else>评分一致，可进入最终结论</p><button class="ink-button" @click="$emit('open', item)">{{ item.decision ? '查看结论' : '形成结论' }}</button></article></template>
