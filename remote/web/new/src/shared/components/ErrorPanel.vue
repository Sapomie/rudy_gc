<template>
  <div class="error-panel" role="alert">
    <strong>加载失败</strong>
    <span>{{ message }}</span>
    <div v-if="$slots.default" class="inline-actions"><slot /></div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ message: string }>();
</script>
