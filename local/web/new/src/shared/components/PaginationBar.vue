<template>
  <nav v-if="items.length" class="app-pagination" aria-label="分页">
    <template v-for="item in items" :key="item.key">
      <span
        v-if="item.disabled"
        class="app-pagination-item"
        :class="{ 'is-active': item.active, 'is-disabled': item.disabled, 'is-ellipsis': item.ellipsis }"
        :aria-current="item.active ? 'page' : undefined"
        aria-disabled="true"
      >
        {{ item.label }}
      </span>
      <a
        v-else
        :href="item.href"
        class="app-pagination-item"
        :class="{ 'is-active': item.active }"
        :aria-current="item.active ? 'page' : undefined"
      >
        {{ item.label }}
      </a>
    </template>
  </nav>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { PageInfo } from "@/shared/types/gc";
import { paginationItems, paginationItemsFromLinks } from "@/shared/components/navigation/models";

const props = withDefaults(
  defineProps<{
    pageInfo?: PageInfo;
    links?: Record<string, unknown>[];
  }>(),
  { pageInfo: undefined, links: undefined }
);
const items = computed(() => (props.links ? paginationItemsFromLinks(props.links) : paginationItems(props.pageInfo)));
</script>
