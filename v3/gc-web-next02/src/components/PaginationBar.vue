<template>
  <nav v-if="links.length" class="pagination-bar" aria-label="分页">
    <a
      v-for="(item, index) in links"
      :key="`${item.Label || 'ellipsis'}-${index}`"
      :href="item.Ellipsis || item.Disabled ? undefined : normalizeHref(item.Href)"
      class="page-pill"
      :class="{ active: item.Active, disabled: item.Disabled || item.Ellipsis }"
    >
      {{ item.Ellipsis ? "..." : item.Label }}
    </a>
  </nav>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { PageInfo } from "@/types/gc";
import { normalizeHref } from "@/utils/format";

const props = defineProps<{ pageInfo?: PageInfo }>();
const links = computed(() => props.pageInfo?.Links || []);
</script>
