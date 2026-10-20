<template>
  <div class="remote-error-state">
    <ErrorPanel :message="message" />
    <div v-if="$slots.default" class="inline-actions"><slot /></div>
  </div>
</template>

<script setup lang="ts">
import ErrorPanel from "@/shared/components/ErrorPanel.vue";

defineProps<{ message: string }>();
</script>
