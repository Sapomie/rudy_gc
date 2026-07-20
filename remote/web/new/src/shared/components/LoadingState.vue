<template>
  <div class="loading-state" aria-label="正在加载">
    <div v-for="line in lines" :key="line" class="loading-line" :class="{ short: line === lines }"></div>
    <div class="loading-grid">
      <span v-for="item in 12" :key="item"></span>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ lines?: number }>(), { lines: 4 });
</script>
