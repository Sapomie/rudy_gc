<template>
  <div class="empty-panel sc-empty-state">
    <strong>{{ title }}</strong>
    <p v-if="message">{{ message }}</p>
    <div v-if="$slots.default" class="inline-actions"><slot /></div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string; message?: string }>();
</script>
