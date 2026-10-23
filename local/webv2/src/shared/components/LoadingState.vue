<template>
  <div class="loading-state" :class="`loading-${layout}`" aria-busy="true">
    <div class="state-line" role="status"><span class="loading-spinner" aria-hidden="true" />{{ label }}</div>
    <div class="loading-placeholders" aria-hidden="true">
      <div v-for="index in count" :key="index" class="loading-placeholder"><i /><span /><span /></div>
    </div>
  </div>
</template>
<script setup lang="ts">
withDefaults(defineProps<{ layout?: "cards" | "rows" | "detail" | "summary"; count?: number; label?: string }>(), {
  layout: "cards",
  count: 12,
  label: "加载中..."
});
</script>
