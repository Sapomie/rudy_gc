<template>
  <form class="rank-date-form rank-stats-date-controls card-grid-size-control" @submit.prevent="applyColumns">
    <label class="rank-stats-filter-field card-grid-size-label">
      <span>每行</span>
      <input
        v-model="draft"
        class="rank-date-input rank-stats-page-size-input card-grid-size-input"
        aria-label="每行卡片数"
        type="number"
        min="1"
        step="1"
        placeholder="默认"
        required
        :disabled="disabled"
      />
    </label>
    <button class="button-secondary rank-toolbar-action" type="submit" :disabled="disabled || !validDraft">设置</button>
    <div class="rank-navigation-group" role="group" aria-label="每行卡片数快捷选项">
      <button
        v-for="count in presets"
        :key="count"
        class="button-secondary rank-toolbar-action"
        :class="{ active: columns === count }"
        type="button"
        :aria-pressed="columns === count"
        :disabled="disabled"
        @click="selectColumns(count)"
      >
        {{ count }}
      </button>
    </div>
    <button
      class="button-secondary rank-toolbar-action"
      :class="{ active: columns === null }"
      type="button"
      :aria-pressed="columns === null"
      :disabled="disabled"
      @click="restoreDefault"
    >
      默认
    </button>
  </form>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { parseCardGridColumns, useCardGridSize } from "@/shared/composables/useCardGridSize";

defineProps<{ disabled?: boolean }>();

const presets = [6, 8, 10];
const { columns, setColumns, resetColumns } = useCardGridSize();
const draft = ref<number | string>("");
const validDraft = computed(() => parseCardGridColumns(draft.value) !== null);

watch(columns, (value) => (draft.value = value ?? ""), { immediate: true });

function applyColumns() {
  setColumns(draft.value);
}

function selectColumns(count: number) {
  draft.value = count;
  setColumns(count);
}

function restoreDefault() {
  draft.value = "";
  resetColumns();
}
</script>

<style scoped>
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

.card-grid-size-control .card-grid-size-label {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 3px;
  color: var(--pink-cards-muted, var(--muted));
  font-size: 10px;
  font-weight: 850;
  white-space: nowrap;
}

.card-grid-size-control .card-grid-size-input {
  width: 52px;
  height: 28px;
  min-height: 28px;
  border: 1px solid var(--pink-cards-line, var(--line));
  border-radius: 6px;
  padding: 2px 4px;
  background: var(--pink-cards-panel, var(--paper));
  color: var(--pink-cards-ink, var(--text));
  font-size: 10px;
  font-weight: 700;
}

.card-grid-size-control .button-secondary:hover:not(:disabled):not(.active) {
  border-color: rgba(217, 79, 138, 0.3);
  background: var(--pink-cards-soft, var(--primary-soft));
  color: var(--pink-cards-primary, var(--primary));
}

.card-grid-size-control .button-secondary.active {
  border-color: rgba(217, 79, 138, 0.32);
  background: var(--pink-cards-primary, var(--primary));
  color: #fff;
}

.card-grid-size-control :is(button, input):focus-visible {
  outline: 2px solid rgba(217, 79, 138, 0.55);
  outline-offset: 2px;
}

.card-grid-size-control :disabled {
  opacity: 0.65;
}
</style>
