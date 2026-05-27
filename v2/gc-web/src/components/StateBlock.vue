<template>
  <div class="rgc-state" :class="`rgc-state-${tone}`">
    <strong>{{ title }}</strong>
    <span v-if="message">{{ message }}</span>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string;
    message?: string;
    tone?: "plain" | "error" | "empty";
  }>(),
  {
    tone: "plain",
  },
);
</script>
