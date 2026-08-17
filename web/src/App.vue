<script setup lang="ts">
const navigation = [
  { to: '/', mark: '库', label: '标准谱系' },
  { to: '/compare', mark: '比', label: '模板对比' },
  { to: '/work', mark: '审', label: '评议席位' },
  { to: '/disagreements', mark: '议', label: '分歧议事' },
  { to: '/history', mark: '史', label: '批次回溯' },
  { to: '/approval', mark: '发', label: '发布审批' },
  { to: '/transfer', mark: '入', label: '受控交换' },
]
</script>
<template>
  <div class="review-hall">
    <header class="topline">
      <div class="seal">评</div><div><strong>审核标准议事厅</strong><small>可复用模板 · 多席位评议 · 不可覆盖否决</small></div>
      <div class="offline">● 离线工作区</div>
    </header>
    <nav class="folio"><RouterLink v-for="item in navigation" :key="item.to" :to="item.to"><i>{{ item.mark }}</i><span>{{ item.label }}</span></RouterLink></nav>
    <main class="desk"><RouterView /></main>
  </div>
</template>
