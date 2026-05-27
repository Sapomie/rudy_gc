<template>
  <nav class="rgc-pager" aria-label="分页">
    <button type="button" :disabled="page <= 1" @click="emitGo(1)">Start</button>
    <button type="button" :disabled="page <= 1" @click="emitGo(page - 1)">Prev</button>
    <button
      v-for="item in pageItems"
      :key="item.key"
      type="button"
      :class="{ active: item.page === page, ghost: item.ellipsis }"
      :disabled="item.ellipsis"
      @click="!item.ellipsis && emitGo(item.page)"
    >
      {{ item.label }}
    </button>
    <button type="button" :disabled="page >= totalPages" @click="emitGo(page + 1)">Next</button>
    <button type="button" :disabled="page >= totalPages" @click="emitGo(totalPages)">End</button>
  </nav>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  page: number;
  pageSize: number;
  total: number;
}>();

const emit = defineEmits<{
  go: [page: number];
}>();

interface PageItem {
  key: string;
  label: string;
  page: number;
  ellipsis: boolean;
}

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / Math.max(1, props.pageSize))));
const pageItems = computed<PageItem[]>(() => {
  const result: PageItem[] = [];
  const left = Math.max(1, props.page - 2);
  const right = Math.min(totalPages.value, props.page + 2);

  if (left > 1) {
    result.push(makeItem(1));
    if (left > 2) {
      result.push({ key: "gap-left", label: "...", page: 0, ellipsis: true });
    }
  }

  for (let page = left; page <= right; page += 1) {
    result.push(makeItem(page));
  }

  if (right < totalPages.value) {
    if (right < totalPages.value - 1) {
      result.push({ key: "gap-right", label: "...", page: 0, ellipsis: true });
    }
    result.push(makeItem(totalPages.value));
  }

  return result;
});

function makeItem(page: number): PageItem {
  return { key: `page-${page}`, label: String(page), page, ellipsis: false };
}

function emitGo(nextPage: number) {
  emit("go", Math.min(Math.max(1, nextPage), totalPages.value));
}
</script>
