<template>
  <div class="card-grid-size-control">
    <div class="rank-navigation-group" role="group" aria-label="每行卡片数快捷选项">
      <button
        v-for="count in presets"
        :key="count"
        class="button-secondary rank-toolbar-action"
        :class="{ active: columns === count }"
        type="button"
        :aria-pressed="columns === count"
        :disabled="disabled"
        @click="setColumns(count)"
      >
        {{ count }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useCardGridSize } from "@/shared/composables/useCardGridSize";

defineProps<{ disabled?: boolean }>();

const presets = [6, 8, 10];
const { columns, setColumns } = useCardGridSize();
</script>

<style scoped>
@layer business {
  :global(.next02-pink-cards-page .cards-toolbar:has(.card-grid-size-control) .cards-toolbar-sort) {
    overflow-x: auto;
    overflow-y: hidden;
  }

  .card-grid-size-control {
    display: inline-flex;
    min-width: max-content;
    flex: 0 0 auto;
    align-items: center;
    gap: 3px;
    margin: 0;
  }

  .card-grid-size-control .button-secondary:hover:not(:disabled):not(.active) {
    border-color: color-mix(in srgb, var(--primary) 30%, transparent);
    background: var(--pink-cards-soft, var(--primary-soft));
    color: var(--pink-cards-primary, var(--primary));
  }

  .card-grid-size-control .button-secondary.active {
    border-color: color-mix(in srgb, var(--primary) 32%, transparent);
    background: var(--pink-cards-primary, var(--primary));
    color: var(--panel);
  }

  .card-grid-size-control button:focus-visible {
    outline: 2px solid color-mix(in srgb, var(--primary) 55%, transparent);
    outline-offset: 2px;
  }

  .card-grid-size-control :disabled {
    opacity: 0.65;
  }
}
</style>
