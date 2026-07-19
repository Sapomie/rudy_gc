<template>
  <nav v-if="items.length" :class="{ 'pagination-bar': variant === 'pill' }" aria-label="分页">
    <ul v-if="variant === 'bootstrap'" class="pagination pagination-sm mb-0">
      <li
        v-for="item in items"
        :key="item.key"
        class="page-item"
        :class="{ active: item.active, disabled: item.disabled }"
      >
        <span v-if="item.ellipsis" class="page-link">...</span>
        <a
          v-else
          class="page-link"
          :href="item.href"
          :tabindex="item.disabled ? -1 : undefined"
          :aria-current="item.active ? 'page' : undefined"
          :aria-disabled="item.disabled || undefined"
        >
          {{ item.label }}
        </a>
      </li>
    </ul>
    <template v-else>
      <a
        v-for="item in items"
        :key="item.key"
        :href="item.href"
        class="page-pill"
        :class="{ active: item.active, disabled: item.disabled }"
        :aria-current="item.active ? 'page' : undefined"
        :aria-disabled="item.disabled || undefined"
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
    variant?: "pill" | "bootstrap";
  }>(),
  { pageInfo: undefined, links: undefined, variant: "pill" }
);
const items = computed(() => (props.links ? paginationItemsFromLinks(props.links) : paginationItems(props.pageInfo)));
</script>
