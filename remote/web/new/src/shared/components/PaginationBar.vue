<template>
  <nav v-if="pageInfo.total" class="remote-list-pager" aria-label="分页">
    <ul class="pagination">
      <li
        v-for="item in items"
        :key="item.key"
        class="page-item"
        :class="{ active: item.active, disabled: item.disabled }"
      >
        <span v-if="item.ellipsis || item.disabled" class="page-link" :aria-disabled="item.disabled || undefined">
          {{ item.label }}
        </span>
        <RouterLink
          v-else
          class="page-link"
          :to="pageLocation(item.page || 1)"
          :aria-current="item.active ? 'page' : undefined"
        >
          {{ item.label }}
        </RouterLink>
      </li>
    </ul>
  </nav>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import type { PageInfo } from "@/features/dailybest/types";
import { buildPaginationItems, compactPageQuery } from "@/shared/utils/pagination";

const props = defineProps<{ pageInfo: PageInfo }>();
const route = useRoute();
const items = computed(() => buildPaginationItems(props.pageInfo));

function pageLocation(page: number) {
  return {
    path: route.path,
    query: compactPageQuery(route.query, page, props.pageInfo.page_size)
  };
}
</script>
