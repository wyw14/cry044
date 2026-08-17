<script setup lang="ts">
import { onMounted, ref } from 'vue'
import DecisionCase from '../components/DecisionCase.vue'
import SnapshotSeal from '../components/SnapshotSeal.vue'
import { reviewGateway } from '../services/reviewGateway'
import { useReviewRoom } from '../stores/reviewRoom'
import type { DeliberationCase } from '../types/review'
const room = useReviewRoom(); const selected = ref<DeliberationCase|null>(null); const rationale = ref(''); const conclusion = ref('退回补充证据'); const message = ref('')
onMounted(room.refresh)
async function resolve(){if(!selected.value)return;try{await reviewGateway.resolve(room.batchID,selected.value.materialID,room.revision,{passed:false,conclusion:conclusion.value,rationale:rationale.value});message.value='最终结论已写入批次';selected.value=null;await room.refresh()}catch(error){message.value=error instanceof Error?error.message:'结论保存失败'}}
</script>
<template><section><header class="chapter"><div><h1>分歧议事单</h1><p>多席位聚合、分歧解释与不可覆盖否决</p></div><SnapshotSeal template="生产质量评议" :version="4"/></header><div class="ledger"><div class="ledger-head"><label>批次<input v-model.trim="room.batchID" /></label><button class="ink-button" @click="room.refresh">刷新议事单</button><span v-if="room.loading">正在聚合各席意见…</span><span class="error">{{ room.error }}</span></div><div class="case-grid"><DecisionCase v-for="item in room.cases" :key="item.materialID" :item="item" @open="selected=$event"/></div></div><div v-if="selected" class="paper" style="margin-top:18px"><h2>材料 {{ selected.materialID }} · 形成最终结论</h2><div v-for="divergence in selected.divergences" :key="divergence.criterionID"><h3>{{ divergence.criterionID }} · 分差 {{ divergence.spread }}</h3><div class="positions"><div v-for="position in divergence.positions" :key="position.reviewerID" class="position"><b>{{ position.reviewerID }}</b><strong>{{ position.score }}</strong><span>{{ position.comment || '无补充意见' }}</span></div></div></div><p v-if="selected.preliminary.vetoed" class="veto">该材料触发否决项，只能形成不通过结论。</p><label class="field">结论<input v-model.trim="conclusion" /></label><label class="field">议事依据<textarea v-model.trim="rationale" rows="4" /></label><button class="ink-button red" @click="resolve">写入最终结论</button></div><p>{{ message }}</p></section></template>
