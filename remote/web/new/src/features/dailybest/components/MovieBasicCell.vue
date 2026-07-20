<template>
  <div class="dailybest-movie-cell">
    <a
      v-if="info.cover_url && !coverFailed"
      class="dailybest-movie-cover"
      :href="info.cover_url"
      target="_blank"
      rel="noreferrer"
      :title="`${displayCode} 封面`"
    >
      <img
        :src="info.cover_url"
        :alt="`${displayCode} 封面`"
        loading="lazy"
        decoding="async"
        @error="coverFailed = true"
      />
    </a>
    <span v-else class="dailybest-movie-cover dailybest-movie-cover-empty" aria-label="暂无封面">NO IMAGE</span>
    <span class="dailybest-movie-copy">
      <strong :title="displayCode">{{ displayCode }}</strong>
      <span :title="info.title || info.movie_name || javId">{{ info.title || info.movie_name || "—" }}</span>
      <small :title="detailLine">{{ detailLine }}</small>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { MovieBasic } from "@/features/dailybest/types";

const props = defineProps<{ javId: string; movie?: MovieBasic }>();
const coverFailed = ref(false);
const info = computed<MovieBasic>(() => props.movie ?? { movie_name: "", movie_code: "", title: "", cover_url: "" });
const displayCode = computed(() => info.value.movie_code || info.value.movie_name || props.javId || "—");
const detailLine = computed(() => {
  const values = [info.value.movie_name, props.javId].filter(
    (value, index, source) => value && source.indexOf(value) === index
  );
  return values.join(" · ") || "—";
});

watch(
  () => props.movie?.cover_url,
  () => {
    coverFailed.value = false;
  }
);
</script>
