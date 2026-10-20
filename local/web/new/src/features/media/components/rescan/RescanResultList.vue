<template>
  <div class="next02-rescan-result-list">
    <div class="next02-rescan-result-title">{{ title }}</div>
    <ul :id="id" class="root-list mb-0">
      <li v-for="item in items" :key="item">
        <code>{{ item }}</code>
      </li>
      <li v-if="!items.length" class="text-muted">暂无数据</li>
    </ul>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  id: string;
  title: string;
  items: string[];
}>();
</script>
