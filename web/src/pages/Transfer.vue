<script setup lang="ts">
import { ref } from 'vue'
const fileName=ref('');const status=ref('等待本地文件');function select(){fileName.value='supplier-standard-v2.json';status.value='格式通过：9 个条目、2 个变量、1 个否决项'}
</script>
<template><section><header class="chapter"><div><h1>受控模板交换</h1><p>本地导入、结构校验、版本比较和离线导出</p></div><em>CONTROLLED / EXCHANGE</em></header><div class="split"><div class="paper"><h2>导入为独立草稿</h2><div class="drop-zone"><p>{{ fileName || '只接受不超过 2MiB 的 JSON 模板' }}</p><button class="ink-button" @click="select">选择本地文件</button></div><p>{{ status }}</p><button class="ink-button red" :disabled="!fileName">导入并生成 revision 1</button></div><div class="paper"><h2>离线导出</h2><p>导出包含模板变量、条目、尺度、适用条件与版本信息，不包含审计人的敏感字段。</p><label class="field">模板版本<select><option>生产质量评议 V4</option><option>供应商准入 V2</option></select></label><button class="ink-button">生成本地 JSON</button></div></div></section></template>
