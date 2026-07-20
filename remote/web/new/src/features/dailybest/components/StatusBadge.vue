<template>
  <span class="status-badge" :data-tone="tone">{{ label }}</span>
</template>

<script setup lang="ts">
defineProps<{
  label: string;
  tone?: string;
}>();
</script>
