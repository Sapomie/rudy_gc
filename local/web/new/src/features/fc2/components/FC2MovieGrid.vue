<template>
  <div class="movie-grid-v2 next02-pink-movie-grid fc2-movie-card-grid" :style="gridStyle">
    <div
      v-for="(card, index) in cards"
      :key="cardKey(card, index)"
      class="col movie-grid-col-v2"
      :style="gridItemStyle"
    >
      <FC2MovieCard :movie="card" />
    </div>
    <div v-if="cards.length === 0" class="movie-empty" style="grid-column: 1 / -1">{{ emptyText }}</div>
  </div>
</template>

<script setup lang="ts">
import FC2MovieCard from "@/features/fc2/components/FC2MovieCard.vue";
import type { FC2Record } from "@/features/fc2/types";
import { text } from "@/shared/utils/format";
import { useCardGridSize } from "@/shared/composables/useCardGridSize";

const { gridStyle, gridItemStyle } = useCardGridSize();

withDefaults(
  defineProps<{
    cards: FC2Record[];
    emptyText?: string;
  }>(),
  {
    emptyText: "暂无 FC2 影片数据"
  }
);

function cardKey(card: FC2Record, index: number): string {
  return text(card, "FC2Key") || String(index);
}
</script>
