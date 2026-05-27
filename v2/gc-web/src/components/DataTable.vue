<template>
  <div class="rgc-table-wrap">
    <table class="rgc-table">
      <thead>
        <tr>
          <th
            v-for="column in columns"
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
          <td v-for="column in columns" :key="column.key">
            <RouterLink
              v-if="column.link && row[column.key] && !isExternalLink(fillLink(column.link, row))"
              :to="fillLink(column.link, row)"
              :class="cellClass(column.key)"
            >
              {{ row[column.key] }}
            </RouterLink>
            <a
              v-else-if="column.link && row[column.key]"
              :href="fillLink(column.link, row)"
              target="_blank"
              rel="noreferrer"
              :class="cellClass(column.key)"
            >
              {{ row[column.key] }}
            </a>
            <span v-else :class="cellClass(column.key)">{{ row[column.key] }}</span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import type { PageColumn, PageRow } from "@/types/api";

defineProps<{
  columns: PageColumn[];
  rows: PageRow[];
  orderBy: string;
  order: string;
}>();

defineEmits<{
  sort: [key: string];
}>();

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
  };
}
</script>
