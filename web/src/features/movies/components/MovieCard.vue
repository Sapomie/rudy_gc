<template>
  <article class="card shadow-sm h-100 card-lift movie-card-v2 next02-pink-movie-card">
    <a class="d-block poster-link" :href="detailHref" :aria-label="`查看 ${name} 详情`" target="_blank" rel="noopener noreferrer">
      <div class="thumb-2x3">
        <img v-if="poster" :src="poster" :alt="`${name} 海报`" loading="lazy" decoding="async">
        <div v-else class="thumb-empty">No Image</div>
        <div class="badge-top-left">
          <span v-if="comeTimes" class="badge-come">❤ {{ comeTimes }}</span>
          <span v-if="scTimes" class="badge-times">× {{ scTimes }}</span>
        </div>
        <div class="badge-bottom-right">
          <span v-if="highestRank" class="badge-rank">#{{ highestRank }}</span>
          <span v-if="score" class="badge-score movie-card-tag-score">★ {{ score }}</span>
        </div>
      </div>
    </a>
    <ul class="list-group list-group-flush">
      <li class="list-group-item movie-card-title-block">
        <h6 class="mb-1 text-truncate" :title="title">
          <a :href="prefixHref" class="stretched-link text-reset text-decoration-none movie-name" target="_blank" rel="noopener noreferrer">{{ name || "-" }}</a>
          {{ " " }}
          <small v-if="watched" class="text-muted ml-2 watch-count">({{ watched }})</small>
        </h6>
        <p v-if="title" class="mb-0 text-muted line-clamp-2 movie-subtitle" :title="title">{{ title }}</p>
      </li>
      <li v-if="periodRankVisible" class="list-group-item period-rank-block movie-card-period-block">
        <div class="period-rank-badges">
          <span class="badge bg-dark">周期 #{{ rankPos }}</span>
          <span class="badge" :class="rankChangeClass">{{ rankChangeText }}</span>
        </div>
        <div class="small text-muted mt-2">
          上期 {{ prevRankText || "--" }} · 峰值 #{{ peakRank }} · 最佳日 #{{ bestRank }}
        </div>
        <div class="small text-muted mt-1">
          上榜 {{ daysInRank }} 天 · 取样 {{ usedPickDays }} 天 · 分数 {{ itemScore }}
        </div>
      </li>
      <li v-if="mediaRowVisible && periodMode" class="list-group-item d-flex align-items-center period-play-row movie-card-media-block">
        <a v-if="mediaPlayable" :href="videoUrl" class="btn btn-sm me-2" :class="mediaButtonClass" target="_blank" rel="noopener noreferrer">{{ mediaLabel }}</a>
        <span v-else-if="mediaLabel" class="btn btn-sm btn-secondary me-2 disabled" aria-disabled="true">{{ mediaLabel }}</span>
        <span v-if="mediaBirth" class="text-muted small ms-2 movie-card-download-date">下载 {{ mediaBirth }}</span>
      </li>
      <li v-else-if="mediaRowVisible" class="list-group-item py-2 movie-card-media-block">
        <div class="card-play-stack">
          <div class="card-play-row">
            <a v-if="mediaPlayable" :href="videoUrl" class="btn btn-sm" :class="mediaButtonClass" target="_blank" rel="noopener noreferrer">{{ mediaLabel }}</a>
            <span v-else class="btn btn-sm btn-secondary disabled" aria-disabled="true">{{ mediaLabel }}</span>
            <span v-if="mediaBirth" class="text-muted small card-play-date movie-card-download-date">下载 {{ mediaBirth }}</span>
          </div>
        </div>
      </li>
      <li class="list-group-item text-muted genre-list movie-card-tag-block">
        <a v-for="genre in genres" :key="genre" class="badge badge-genre movie-card-tag-genre me-1 mb-1" :href="`/listcard?gn=${encodeURIComponent(genre)}`" target="_blank" rel="noopener noreferrer">
          {{ genre }}
        </a>
      </li>
      <li class="list-group-item cast-list movie-card-tag-block">
        <a v-for="cast in casts" :key="cast.key" class="person-tag movie-card-tag-person movie-card-tag-cast mr-2" :href="cast.href" :title="cast.label" target="_blank" rel="noopener noreferrer">
          {{ cast.label }}
        </a>
        <a v-if="director" class="person-tag movie-card-tag-person movie-card-tag-director ml-2" :href="`/listcard?dn=${encodeURIComponent(director)}`" :title="director" target="_blank" rel="noopener noreferrer">
          {{ director }}
        </a>
      </li>
      <li class="list-group-item d-flex align-items-center small card-link-row movie-card-footer">
        <a v-if="busUrl" :href="busUrl" class="movie-card-tag-link me-2" target="_blank" rel="noopener noreferrer">Bus</a>
        <a v-if="searchUrl" :href="searchUrl" class="movie-card-tag-link me-2" target="_blank" rel="noopener noreferrer">Suk</a>
        <a v-if="javUrl" :href="javUrl" class="movie-card-tag-link me-2" target="_blank" rel="noopener noreferrer">Lib</a>
        <span v-if="releaseDate" class="movie-card-tag-date ms-auto text-muted">上映 {{ releaseDate }}</span>
      </li>
    </ul>
  </article>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { AnyRecord } from "@/shared/types/gc";
import { asArray, asRecord, num, pick, text } from "@/shared/utils/format";

const props = defineProps<{
  movie: AnyRecord;
  rank?: AnyRecord;
  card?: AnyRecord;
  mode?: string;
}>();

const movie = computed(() => asRecord(props.movie));
const rank = computed(() => asRecord(props.rank));
const card = computed(() => asRecord(props.card));
const name = computed(() => text(movie.value, "name", "Name", "jav_id", "JavID"));
const title = computed(() => text(movie.value, "title", "Title"));
const poster = computed(() => text(movie.value, "jacket_img", "JacketImg", "cover", "Cover"));
const prefix = computed(() => text(movie.value, "prefix", "Prefix"));
const score = computed(() => text(movie.value, "score", "Score"));
const watched = computed(() => text(movie.value, "viewers_number_watched", "ViewersNumberWatched"));
const releaseDate = computed(() => text(movie.value, "releasing_date", "ReleasingDate"));
const director = computed(() => text(movie.value, "director", "Director"));
const detailHref = computed(() => `/moviedetail/${encodeURIComponent(name.value)}`);
const prefixHref = computed(() => `/listcard?pn=${encodeURIComponent(prefix.value)}`);
const busUrl = computed(() => text(movie.value, "bus_url", "BusUrl"));
const searchUrl = computed(() => text(movie.value, "search_url", "SearchUrl"));
const javUrl = computed(() => text(movie.value, "jav_url", "JavUrl"));
const comeTimes = computed(() => num(movie.value, "come_times", "ComeTimes"));
const scTimes = computed(() => num(movie.value, "sc_times", "ScTimes"));
const highestRank = computed(() => num(movie.value, "highest_rank", "HighestRank"));
const mediaBirth = computed(() => text(movie.value, "film_birth_date_w_media", "FilmBirthDateWMedia"));
const videoUrl = computed(() => text(movie.value, "video_url_w_media", "VideoUrlWMedia"));
const ownedMedia = computed(() => num(movie.value, "owned_w_media", "OwnedWMedia"));
const periodMode = computed(() => props.mode === "rank-period" || Boolean(card.value.Item || card.value.item));
const mediaLabel = computed(() => {
  if (ownedMedia.value === 4 || ownedMedia.value === 5) {
    return "Play";
  }
  if (ownedMedia.value === 6) {
    return "Removed";
  }
  return "";
});
const mediaButtonClass = computed(() => ownedMedia.value === 5 ? "btn-info" : "btn-success");
const mediaPlayable = computed(() => Boolean(videoUrl.value) && (ownedMedia.value === 4 || ownedMedia.value === 5));
const mediaRowVisible = computed(() => periodMode.value ? ownedMedia.value >= 2 : Boolean(mediaLabel.value));
const genres = computed(() => asArray<string>(pick(movie.value, "genre", "Genre")).map(String).filter(Boolean));
const casts = computed(() => asArray<AnyRecord>(pick(movie.value, "cast", "Cast")).map((item, index) => {
  const record = asRecord(item);
  const label = text(record, "NameShow", "DisplayName", "Name", "name");
  const personID = text(record, "PersonId", "person_id", "Id", "id");
  return {
    key: `${label}-${index}`,
    label,
    href: personID && personID !== "0" ? `/castdetail/${encodeURIComponent(personID)}` : `/listcard?cn=${encodeURIComponent(label)}`
  };
}).filter((item) => item.label));
const rankPos = computed(() => text(rank.value, "RankPos", "rank_pos"));
const peakRank = computed(() => text(rank.value, "PeakRank", "peak_rank") || "--");
const bestRank = computed(() => text(rank.value, "BestRank", "best_rank") || "--");
const daysInRank = computed(() => text(rank.value, "DaysInRank", "days_in_rank") || "0");
const usedPickDays = computed(() => text(rank.value, "UsedPickDays", "used_pick_days") || "0");
const prevRankText = computed(() => text(card.value, "PrevRankText", "prev_rank_text"));
const rankChangeText = computed(() => text(card.value, "RankChangeText", "rank_change_text"));
const rankChangeClass = computed(() => text(card.value, "RankChangeClass", "rank_change_class") || "bg-secondary");
const itemScore = computed(() => {
  const value = text(rank.value, "Score", "score");
  return value ? Number(value).toFixed(2) : "0.00";
});
const periodRankVisible = computed(() => periodMode.value && Boolean(rankPos.value || daysInRank.value || rankChangeText.value));
</script>
