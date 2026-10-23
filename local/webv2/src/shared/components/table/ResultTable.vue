<template>
  <div :class="cardClasses">
    <div class="card-body">
      <div :class="headClass">
        <div>
          <component :is="variant === 'sc' ? 'span' : 'div'" :class="eyebrowClass">RESULT</component>
          <h5 class="card-title mb-0">{{ table.title }}</h5>
        </div>
        <span :class="countClass">{{ table.rows.length }} {{ variant === "sc" ? "项" : "行" }}</span>
      </div>
      <div class="table-responsive table-scroll">
        <table :id="table.id" :class="tableClass">
          <colgroup>
            <col v-for="column in table.columns" :key="column.key" :style="{ width: column.width }" />
          </colgroup>
          <thead>
            <tr>
              <th v-for="column in table.columns" :key="column.key">{{ column.label }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, index) in table.rows" :key="rowKey(row, index)">
              <td
                v-for="column in table.columns"
                :key="column.key"
                :class="column.cls"
                :title="displayValue(row, column)"
              >
                {{ displayValue(row, column) }}
              </td>
            </tr>
            <tr v-if="!table.rows.length">
              <td :colspan="table.columns.length || 1" class="text-center text-muted">暂无数据</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { formatValue } from "@/shared/utils/format";

export interface ResultColumn {
  label: string;
  key: string;
  width?: string;
  cls?: string;
}
export interface ResultTableDefinition {
  id: string;
  title: string;
  columns: ResultColumn[];
  rows: Record<string, unknown>[];
}

const props = withDefaults(
  defineProps<{
    table: ResultTableDefinition;
    variant?: "media" | "sc";
    cardClass?: string;
    rowKey?: (row: Record<string, unknown>, index: number) => string | number;
    cellValue?: (row: Record<string, unknown>, column: ResultColumn) => string;
  }>(),
  {
    variant: "media",
    cardClass: "",
    rowKey: (_row: Record<string, unknown>, index: number) => index,
    cellValue: undefined
  }
);

const cardClasses = computed(() =>
  props.variant === "sc"
    ? ["card", "next02-sc-panel", "next02-sc-result-panel", props.cardClass]
    : ["card", "next02-media-result-card", props.cardClass]
);
const headClass = computed(() => (props.variant === "sc" ? "next02-sc-panel-head" : "next02-media-panel-head"));
const eyebrowClass = computed(() => (props.variant === "sc" ? "next02-sc-panel-kicker" : "next02-media-eyebrow"));
const countClass = computed(() => (props.variant === "sc" ? "next02-sc-result-count" : "next02-media-count-pill"));
const tableClass = computed(() =>
  props.variant === "sc"
    ? "data-table table table-hover align-middle mb-0 next02-sc-media-table"
    : "data-table table table-hover align-middle mb-0 next02-media-table"
);

function displayValue(row: Record<string, unknown>, column: ResultColumn): string {
  if (props.cellValue) return props.cellValue(row, column);
  const value = row[column.key];
  return Array.isArray(value) ? value.map(String).join("\n") : formatValue(value);
}
</script>
