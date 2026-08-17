<script setup lang="ts">
import { computed, ref } from 'vue'
import SnapshotSeal from '../components/SnapshotSeal.vue'
const scores = ref([{id:'quality',title:'制造一致性',required:true,value:0,evidence:''},{id:'trace',title:'证据可追溯',required:false,value:0,evidence:''},{id:'safety',title:'安全合规（否决）',required:true,value:0,evidence:''}])
const opinion = ref('')
const complete = computed(() => scores.value.filter(item => item.required).every(item => item.value > 0 && item.evidence.trim()))
</script>
<template><section><header class="chapter"><div><h1>评议席位 · R-02</h1><p>分配材料、逐条评分、意见与本地证据</p></div><SnapshotSeal template="生产质量评议" :version="4"/></header><div class="split"><div class="paper"><h2>待审材料 M-104</h2><p>新能源壳体首件验证包</p><div class="print-sheet">校验和：3c9b…108e<br/>本地附件：drawing.pdf、inspection.jpg<br/>分配截止：今天 17:00</div></div><form class="paper"><h2>按快照条目评分</h2><div v-for="item in scores" :key="item.id" class="field"><label>{{ item.title }} {{ item.required ? '＊' : '' }}</label><select v-model.number="item.value"><option :value="0">未评分</option><option v-for="score in 5" :key="score" :value="score">{{ score }}</option></select><input v-model.trim="item.evidence" placeholder="本地证据编号" /></div><label class="field">评审意见<textarea v-model.trim="opinion" rows="4" /></label><button class="ink-button red" :disabled="!complete">提交本席意见</button></form></div></section></template>
