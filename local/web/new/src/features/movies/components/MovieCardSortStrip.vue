<template>
  <div v-if="items.length" class="link-strip">
    <SortLink v-for="item in items" :key="item.key" class="strip-link" :link="item.link" :label="item.label" />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { MovieRecord } from "@/features/movies/types";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import { asRecord, text } from "@/shared/utils/format";

const props = defineProps<{
  query?: MovieRecord;
  options: Array<{ key: string; label: string }>;
}>();

const items = computed(() =>
  props.options
    .map((option) => ({
      ...option,
      link: asRecord(props.query?.[option.key])
    }))
    .filter((item) => text(item.link, "Href", "href"))
);
</script>
