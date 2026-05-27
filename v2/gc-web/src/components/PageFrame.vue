<template>
  <section class="rgc-frame">
    <div class="rgc-frame-head">
      <div>
        <p v-if="eyebrow" class="rgc-eyebrow">{{ eyebrow }}</p>
        <h1>{{ title }}</h1>
        <p v-if="description" class="rgc-description">{{ description }}</p>
      </div>
      <div v-if="$slots.actions" class="rgc-frame-actions">
        <slot name="actions" />
      </div>
    </div>
    <slot />
  </section>
</template>

<script setup lang="ts">
defineProps<{
  title: string;
  eyebrow?: string;
  description?: string;
}>();
</script>
