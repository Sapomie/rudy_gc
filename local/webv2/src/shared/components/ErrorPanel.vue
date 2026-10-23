<template>
  <div class="error-panel state-line error" role="alert">
    <strong>加载失败</strong>
    <span>{{ message }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{ message: string }>();
</script>
