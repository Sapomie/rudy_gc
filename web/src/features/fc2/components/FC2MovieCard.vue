<template>
  <article class="card shadow-sm h-100 card-lift movie-card-v2 next02-pink-movie-card fc2-movie-card">
    <a class="d-block poster-link" :href="detailHref" :aria-label="`查看 ${fc2Key} 详情`" target="_blank" rel="noopener noreferrer">
      <div class="thumb-2x3">
        <img v-if="coverImageURL" :src="coverImageURL" :alt="`${fc2Key} 联系表封面`" loading="lazy" decoding="async">
        <div v-else class="thumb-empty fc2-card-cover-empty" aria-label="暂无联系表封面"></div>
        <div class="badge-top-left">
          <span v-if="comeTimes" class="badge-come">❤ {{ comeTimes }}</span>
          <span v-if="scTimes" class="badge-times">× {{ scTimes }}</span>
        </div>
      </div>
    </a>
    <ul class="list-group list-group-flush">
      <li class="list-group-item movie-card-title-block">
        <h6 class="mb-1 text-truncate" :title="title">
          <a :href="detailHref" class="stretched-link text-reset text-decoration-none movie-name" target="_blank" rel="noopener noreferrer">{{ fc2Key || "-" }}</a>
        </h6>
        <p v-if="title" class="mb-0 text-muted line-clamp-2 movie-subtitle" :title="title">{{ title }}</p>
      </li>
      <li v-if="showPlayButton" class="list-group-item py-2 movie-card-media-block">
        <div class="card-play-stack">
          <div class="card-play-row">
            <a :href="videoURL" class="btn btn-sm" :class="playButtonClass" target="_blank" rel="noopener noreferrer">{{ playButtonText }}</a>
            <span v-if="downloadTime" class="text-muted small card-play-date movie-card-download-date">下载 {{ formatUnix(downloadTime) }}</span>
          </div>
        </div>
      </li>
      <li class="list-group-item cast-list movie-card-tag-block fc2-card-performer-block">
        <span v-if="performerText" class="person-tag movie-card-tag-person movie-card-tag-cast" :title="performerText">{{ performerText }}</span>
      </li>
      <li class="list-group-item movie-card-sc-history fc2-card-sc-history">
        <span class="movie-card-sc-history-label">最近入选 SC</span>
        <span class="movie-card-sc-history-value">{{ lastSelectedScTime ? formatUnix(lastSelectedScTime) : "-" }}</span>
      </li>
      <li class="list-group-item d-flex align-items-center small card-link-row movie-card-footer">
        <span class="movie-card-source source-fc2">FC2</span>
        <span v-if="postTime" class="movie-card-tag-date ms-auto text-muted">发布 {{ formatUnix(postTime) }}</span>
      </li>
      <slot name="actions" />
    </ul>
  </article>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { AnyRecord } from "@/shared/types/gc";
import { num, pick, text } from "@/shared/utils/format";

const props = defineProps<{ movie: AnyRecord }>();

const fc2Key = computed(() => text(props.movie, "FC2Key"));
const title = computed(() => text(props.movie, "Title"));
const performerText = computed(() => text(props.movie, "PerformerText"));
const coverImageURL = computed(() => text(props.movie, "CoverImageURL"));
const detailHref = computed(() => `/fcdetail/${encodeURIComponent(fc2Key.value)}`);
const videoURL = computed(() => text(props.movie, "VideoURL"));
const showPlayButton = computed(() => Boolean(pick(props.movie, "ShowPlayButton")) && Boolean(videoURL.value));
const playButtonClass = computed(() => text(props.movie, "PlayButtonClass") || "btn-success");
const playButtonText = computed(() => text(props.movie, "PlayButtonText") || "Play");
const downloadTime = computed(() => num(props.movie, "DownloadTime"));
const scTimes = computed(() => num(props.movie, "ScTimes"));
const comeTimes = computed(() => num(props.movie, "ComeTimes"));
const lastSelectedScTime = computed(() => num(props.movie, "LastSelectedScTime"));
const postTime = computed(() => num(props.movie, "PostTime"));

function formatUnix(value: number): string {
  const date = new Date(value * 1000);
  if (!value || Number.isNaN(date.getTime())) return "-";
  const pad = (part: number) => String(part).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}
</script>
