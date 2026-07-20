<template>
  <section class="page-shell">
    <header class="page-heading">
      <div>
        <p v-if="eyebrow" class="page-eyebrow">{{ eyebrow }}</p>
        <h1>{{ title }}</h1>
        <p v-if="note" class="page-note">{{ note }}</p>
      </div>
      <div v-if="$slots.actions" class="page-actions">
        <slot name="actions" />
      </div>
    </header>
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
