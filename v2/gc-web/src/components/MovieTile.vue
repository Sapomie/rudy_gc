<template>
  <article class="rgc-movie-card">
    <RouterLink class="rgc-poster-link" :to="detailHref" :aria-label="`查看 ${movie.movie_name} 详情`">
      <img
        v-if="movie.jacket_img && !imageFailed"
        :src="movie.jacket_img"
        :alt="`${movie.movie_name} 海报`"
        loading="lazy"
        decoding="async"
        @error="imageFailed = true"
      />
      <span v-else class="rgc-poster-fallback">{{ movie.movie_name }}</span>

      <span v-if="movie.come_times > 0 || movie.sc_times > 0" class="rgc-poster-badges top">
        <b v-if="movie.come_times > 0" class="come">❤ {{ movie.come_times }}</b>
        <b v-if="movie.sc_times > 0" class="sc">× {{ movie.sc_times }}</b>
      </span>

      <span class="rgc-poster-badges bottom">
        <b v-if="movie.highest_rank > 0" class="rank">#{{ movie.highest_rank }}</b>
        <b v-if="Number.isFinite(movie.score)" class="score">★ {{ movie.score }}</b>
      </span>
    </RouterLink>

    <div class="rgc-card-section title">
      <h2 :title="movie.title || movie.movie_name">
        <RouterLink :to="prefixHref">{{ movie.movie_name }}</RouterLink>
        <small v-if="Number.isFinite(movie.viewers_number_watched)">({{ movie.viewers_number_watched }})</small>
      </h2>
      <p v-if="movie.title" :title="movie.title">{{ movie.title }}</p>
    </div>

    <div v-if="mediaLineVisible" class="rgc-card-section media">
      <a v-if="mediaHref" :href="mediaHref" :class="['rgc-media-action', mediaTone]">MPlay</a>
      <span v-else class="rgc-media-action removed">MRemoved</span>
      <small v-if="movie.film_birth_date_w_media">{{ movie.film_birth_date_w_media }}</small>
    </div>

    <div class="rgc-card-section tags muted">
      <RouterLink v-for="genre in movie.genre || []" :key="`${movie.movie_jav_id}-${genre}`" :to="genreHref(genre)">
        {{ genre }}
      </RouterLink>
    </div>

    <div class="rgc-card-section people">
      <RouterLink
        v-for="cast in movie.cast || []"
        :key="`${movie.movie_jav_id}-${cast.person_id}-${cast.name}`"
        :to="castHref(cast)"
        :title="cast.name_show || cast.display_name || cast.name"
      >
        {{ cast.name_show || cast.display_name || cast.name }}
      </RouterLink>
      <RouterLink v-if="movie.director" :to="movie.director_href || `/cards?dn=${encodeURIComponent(movie.director)}`" :title="movie.director">
        {{ movie.director }}
      </RouterLink>
    </div>

    <div class="rgc-card-section links">
      <a v-if="movie.bus_url" :href="movie.bus_url" target="_blank" rel="noopener">Bus</a>
      <a v-if="movie.search_url" :href="movie.search_url" target="_blank" rel="noopener">Suk</a>
      <a v-if="movie.jav_url" :href="movie.jav_url" target="_blank" rel="noopener">Lib</a>
      <time v-if="movie.releasing_date">{{ movie.releasing_date }}</time>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";

import type { CastTag, MovieCard } from "@/types/api";

const props = defineProps<{
  movie: MovieCard;
}>();

const imageFailed = ref(false);
const detailHref = computed(() => props.movie.movie_href || `/movie/${encodeURIComponent(props.movie.movie_name)}`);
const prefixHref = computed(() => props.movie.prefix_href || `/cards?pn=${encodeURIComponent(props.movie.prefix || props.movie.movie_name.split("-")[0] || "")}`);
const mediaLineVisible = computed(() => props.movie.owned_w_media >= 4);
const mediaHref = computed(() => (props.movie.owned_w_media === 6 ? "" : props.movie.video_url_w_media));
const mediaTone = computed(() => (props.movie.owned_w_media === 5 ? "info" : "success"));

function genreHref(genre: string) {
  return `/cards?gn=${encodeURIComponent(genre)}`;
}

function castHref(cast: CastTag) {
  return cast.href || `/cards?cn=${encodeURIComponent(cast.name)}`;
}
</script>
