<script setup lang="ts">defineProps<{ template: string; version: number; locked?: boolean }>()</script>
<template><span class="snapshot-seal"><b>{{ template }} · V{{ version }}</b><small>{{ locked === false ? '草稿' : '批次快照已封存' }}</small></span></template>
