<template>
  <div class="movie-grid-v2 next02-pink-movie-grid">
    <div v-for="(item, index) in cards" :key="cardKey(item, index)" class="col movie-grid-col-v2">
      <MovieCard :movie="movieOf(item)" :rank="rankOf(item)" :card="item" :mode="mode" />
    </div>
    <div v-if="cards.length === 0" class="movie-empty">暂无影片数据</div>
  </div>
</template>

<script setup lang="ts">
import type { AnyRecord } from "@/types/gc";
import { asRecord, text } from "@/utils/format";
import MovieCard from "@/components/MovieCard.vue";

defineProps<{ cards: AnyRecord[]; mode?: string }>();

function movieOf(item: AnyRecord): AnyRecord {
  return asRecord(item.Movie || item.movie || item);
}

function rankOf(item: AnyRecord): AnyRecord {
  return asRecord(item.Item || item.item || item);
}

function cardKey(item: AnyRecord, index: number): string {
  const movie = movieOf(item);
  return text(movie, "name", "Name", "jav_id", "JavID") || String(index);
}
</script>
