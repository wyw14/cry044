<script setup lang="ts">
const rows = [
  { item: '变量 ${product_line}', left: '固定“总装”', right: '按项目参数解析', kind: 'changed' },
  { item: '安全合规', left: '必审 · 权重 20%', right: '必审 · 否决项', kind: 'changed' },
  { item: '证据可追溯', left: '无', right: '可选 · 权重 10%', kind: 'added' },
  { item: '包装外观', left: '可选 · 权重 5%', right: '已移除', kind: 'removed' },
]
</script>
<template><section><header class="chapter"><div><h1>模板克隆与差异册</h1><p>条目、变量、权重、尺度和否决语义逐项比较</p></div><em>TEMPLATE / DIFF</em></header><div class="ledger"><div class="ledger-head"><SnapshotSeal template="生产质量评议" :version="4"/><span>对比</span><SnapshotSeal template="生产质量评议" :version="5" :locked="false"/></div><div class="lineage"><article v-for="row in rows" :key="row.item"><span class="tag" :class="{red:row.kind==='removed'}">{{ row.kind }}</span><div><strong>{{ row.item }}</strong><p><del>{{ row.left }}</del> → {{ row.right }}</p></div><button class="ink-button">查看条件</button></article></div></div></section></template>
