<template>
  <div class="rgc-table-wrap">
    <table class="rgc-table">
      <thead>
        <tr>
          <th
            v-for="column in visibleColumns"
            :key="column.key"
            :class="{ sortable: column.sortable, active: column.key === orderBy }"
            @click="column.sortable && $emit('sort', column.key)"
          >
            <span>{{ column.label }}</span>
            <b v-if="column.key === orderBy">{{ order.toLowerCase() === "asc" ? "↑" : "↓" }}</b>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, rowIndex) in rows" :key="row.id || row.name || rowIndex">
          <td v-for="column in visibleColumns" :key="column.key">
            <RouterLink
              v-if="column.link && row[column.key] && fillLink(column.link, row) && !isExternalLink(fillLink(column.link, row))"
              :to="fillLink(column.link, row)"
              :class="cellClass(column.key)"
            >
              {{ formatCell(column.key, row[column.key]) }}
            </RouterLink>
            <a
              v-else-if="column.link && row[column.key] && fillLink(column.link, row)"
              :href="fillLink(column.link, row)"
              target="_blank"
              rel="noreferrer"
              :class="cellClass(column.key)"
            >
              {{ formatCell(column.key, row[column.key]) }}
            </a>
            <span v-else :class="cellClass(column.key)">{{ formatCell(column.key, row[column.key]) }}</span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

import type { PageColumn, PageRow } from "@/types/api";

const props = defineProps<{
  columns: PageColumn[];
  rows: PageRow[];
  orderBy: string;
  order: string;
}>();

defineEmits<{
  sort: [key: string];
}>();

const visibleColumns = computed(() => props.columns.filter((column) => !column.hidden));

function fillLink(pattern: string, row: PageRow) {
  const directMatch = pattern.match(/^\{([^}]+)\}$/);
  if (directMatch) {
    return row[directMatch[1]] || "";
  }
  return pattern.replace(/\{([^}]+)\}/g, (_, key: string) => encodeURIComponent(row[key] || ""));
}

function isExternalLink(target: string) {
  return /^https?:\/\//i.test(target);
}

function cellClass(key: string) {
  return {
    "rgc-table-cell-long": /file|dir|path|title|url|error|remarks|hash/i.test(key),
    "rgc-table-cell-number": /count|number|times|size|rank|height|width|duration|rate|page|total|id$/i.test(key),
  };
}

function formatCell(key: string, value: string | undefined) {
  if (value === undefined || value === "") {
    return "-";
  }
  if (/size_bytes/i.test(key)) {
    return formatBytes(Number(value));
  }
  if (/size$/i.test(key) && /^\d+$/.test(value)) {
    return formatBytes(Number(value));
  }
  if (/(time|date|birth|release|created|updated|last_sc)$/i.test(key) && /^\d{9,}$/.test(value)) {
    return formatUnix(Number(value), /date|birth|release/i.test(key));
  }
  if (/duration/i.test(key) && /^\d+$/.test(value)) {
    return formatDuration(Number(value));
  }
  return value;
}

function formatBytes(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return "-";
  }
  const units = ["B", "KB", "MB", "GB", "TB"];
  let current = value;
  let unit = 0;
  while (current >= 1024 && unit < units.length - 1) {
    current /= 1024;
    unit += 1;
  }
  return `${current.toFixed(current >= 10 ? 1 : 2)} ${units[unit]}`;
}

function formatUnix(value: number, dateOnly: boolean) {
  if (!Number.isFinite(value) || value <= 0) {
    return "-";
  }
  const date = new Date(value * 1000);
  const yyyy = date.getFullYear();
  const mm = String(date.getMonth() + 1).padStart(2, "0");
  const dd = String(date.getDate()).padStart(2, "0");
  if (dateOnly) {
    return `${yyyy}-${mm}-${dd}`;
  }
  const hh = String(date.getHours()).padStart(2, "0");
  const mi = String(date.getMinutes()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}`;
}

function formatDuration(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return "-";
  }
  if (value >= 3600) {
    return `${Math.floor(value / 3600)}h ${Math.floor((value % 3600) / 60)}m`;
  }
  if (value >= 60) {
    return `${Math.floor(value / 60)}m`;
  }
  return `${value}s`;
}
</script>
