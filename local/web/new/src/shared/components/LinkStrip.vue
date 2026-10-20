<template>
  <div v-if="items.length" class="link-strip">
    <a
      v-for="(item, index) in items"
      :key="`${item.label}-${index}`"
      :href="item.href || '#'"
      class="strip-link"
      :class="{ active: item.active, disabled: item.disabled }"
    >
      {{ item.label }}
    </a>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  items: { label: string; href?: string; active?: boolean; disabled?: boolean }[];
}>();
</script>
