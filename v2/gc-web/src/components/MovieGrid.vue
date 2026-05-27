<template>
  <div class="rgc-movie-grid">
    <MovieTile v-for="movie in items" :key="movie.movie_jav_id || movie.movie_name" :movie="movie" />
  </div>
</template>

<script setup lang="ts">
import MovieTile from "@/components/MovieTile.vue";
import type { MovieCard } from "@/types/api";

defineProps<{
  items: MovieCard[];
}>();
</script>
