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
            <b v-if="column.key === orderBy">{{ order === "asc" ? "↑" : "↓" }}</b>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, rowIndex) in rows" :key="row.id || row.name || rowIndex">
          <td v-for="column in columns" :key="column.key">
            <RouterLink v-if="column.link && row[column.key]" :to="fillLink(column.link, row)">
              {{ row[column.key] }}
            </RouterLink>
            <span v-else>{{ row[column.key] }}</span>
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
  return pattern.replace(/\{([^}]+)\}/g, (_, key: string) => encodeURIComponent(row[key] || ""));
}
</script>
