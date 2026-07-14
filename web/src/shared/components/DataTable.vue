<template>
  <div class="table-card">
    <div v-if="title || total !== undefined" class="table-card-head">
      <h2 v-if="title">{{ title }}</h2>
      <span v-if="total !== undefined" class="soft-badge">Total {{ total }}</span>
    </div>
    <div class="table-scroll">
      <table class="data-table">
        <thead>
          <tr>
            <th v-for="column in columns" :key="column">{{ labelFor(column) }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="rows.length === 0">
            <td :colspan="columns.length || 1" class="empty-cell">暂无数据</td>
          </tr>
          <tr v-for="(row, index) in rows" :key="index">
            <td v-for="column in columns" :key="column">
              <a v-if="linkFor(row, column)" :href="linkFor(row, column) || '#'" class="table-link">
                {{ cellText(row, column) }}
              </a>
              <span v-else>{{ cellText(row, column) }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { AnyRecord } from "@/shared/types/gc";
import { asRecord, formatValue, text } from "@/shared/utils/format";

const props = defineProps<{
  rows: AnyRecord[];
  title?: string;
  total?: number | string;
  preferredColumns?: string[];
}>();

const columns = computed(() => {
  const seen = new Set<string>();
  const preferred = props.preferredColumns || [];
  for (const key of preferred) {
    seen.add(key);
  }
  for (const row of props.rows.slice(0, 12)) {
    for (const key of Object.keys(row)) {
      if (!key.toLowerCase().endsWith("href") && !key.toLowerCase().endsWith("url")) {
        seen.add(key);
      }
    }
  }
  return Array.from(seen).slice(0, 14);
});

function labelFor(column: string): string {
  return column.replace(/([a-z])([A-Z])/g, "$1 $2").replaceAll("_", " ");
}

function cellText(row: AnyRecord, column: string): string {
  return formatValue(asRecord(row)[column]);
}

function linkFor(row: AnyRecord, column: string): string {
  const record = asRecord(row);
  const direct = text(record, `${column}Href`, `${column}_href`, `${column}URL`, `${column}_url`);
  if (direct) {
    return direct;
  }
  if (/name|title|jav|movie/i.test(column)) {
    const movie = text(record, "movie_name", "MovieName", "name", "Name", "JavID", "jav_id");
    if (movie && column !== "SiteName") {
      return `/moviedetail/${encodeURIComponent(movie)}`;
    }
  }
  return "";
}
</script>
