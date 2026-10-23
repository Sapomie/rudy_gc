<template>
  <section class="page-shell" :aria-label="title">
    <div v-if="$slots.actions" class="page-actions page-toolbar" role="group" aria-label="页面操作">
      <slot name="actions" />
    </div>
    <slot />
  </section>
</template>

<script setup lang="ts">
defineProps<{
  title: string;
  note?: string;
  eyebrow?: string;
}>();
</script>
