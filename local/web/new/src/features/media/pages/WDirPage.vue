<template>
  <PageShell
    :class="[
      'wdir-page-shell',
      'cards-page-shell',
      'next02-pink-cards-page',
      'next02-pink-cards-page-shell',
      wdirVariantClass
    ]"
    :title="pageHeadingTitle"
    :note="pageNote"
    eyebrow="WDirectory"
  >
    <div id="page-root" :class="['container-fluid wdir-page-root next02-wdir-rewrite', wdirVariantClass]">
      <LoadingState v-if="loading" />
      <ErrorPanel v-else-if="error" :message="error" />
      <template v-else>
        <template v-if="isDetail">
          <section class="next02-wdir-overview" aria-label="目录概览">
            <div class="next02-wdir-overview-toolbar">
              <nav class="next02-wdir-breadcrumbs" aria-label="breadcrumb">
                <a href="/mediadir">根目录</a>
                <template v-for="(item, index) in breadcrumbs" :key="crumbKey(item, index)">
                  <span aria-hidden="true">›</span>
                  <a v-if="index < breadcrumbs.length - 1" :href="`/mediadir?id=${crumbId(item)}`">{{
                    crumbName(item)
                  }}</a>
                  <strong v-else>{{ crumbName(item) }}</strong>
                </template>
              </nav>
              <div class="next02-wdir-overview-metrics" aria-label="目录统计">
                <span v-for="item in statCards" :key="item.label"
                  ><small>{{ item.label }}</small
                  ><strong
                    ><ScDateTime v-if="item.dateTime" :value="item.value" unix /><template v-else>{{
                      item.value || "-"
                    }}</template></strong
                  ></span
                >
              </div>
            </div>
            <ChildrenGrid :children="children" />
          </section>

          <div
            v-if="hasMovieDirectory"
            class="sticky-toolbar cards-toolbar wdir-movie-toolbar next02-pink-cards-toolbar"
            aria-label="WDir 影片控制栏"
          >
            <div v-if="isJAVDirectory" class="cards-toolbar-sort">
              <MovieCardSortStrip :query="sortQuery" :options="sortOptions" />
            </div>
            <div class="toolbar-line cards-toolbar-bottom">
              <div class="owned-line cards-toolbar-filters">
                <LinkStrip v-if="isJAVDirectory" :items="ownedButtons" />
                <a
                  class="button-secondary wdir-recursive-button"
                  :class="{ active: recursive }"
                  :href="recursiveHref(true)"
                  >递归</a
                >
                <a
                  class="button-secondary wdir-recursive-button"
                  :class="{ active: !recursive }"
                  :href="recursiveHref(false)"
                  >直属</a
                >
                <button
                  v-if="isJAVDirectory"
                  type="button"
                  class="button-secondary cards-filter-button"
                  :class="{ active: filterOpen }"
                  :aria-expanded="filterOpen"
                  @click.stop.prevent="toggleFilter"
                >
                  筛选
                </button>
              </div>
              <div class="next02-pink-cards-toolbar-status">
                <div class="next02-pink-cards-metrics" aria-label="WDir 页面摘要">
                  <span
                    ><strong>{{ total }}</strong> Total</span
                  >
                  <span
                    ><strong>{{ movies.length }}</strong> Current Page</span
                  >
                </div>
                <div v-if="pageLinks.length" class="cards-toolbar-pagination next02-pink-cards-pagination">
                  <PaginationBar :page-info="pageInfo" />
                </div>
              </div>
            </div>
          </div>

          <div
            v-if="isJAVDirectory"
            id="movieCardFilterCollapse"
            class="collapse next02-pink-cards-filter next02-wdir-filter"
            :class="{ show: filterOpen }"
          >
            <div class="card">
              <div class="card-body">
                <form class="row g-2 align-items-end" method="get" :action="filterAction">
                  <input v-if="filterValue('PageSize')" type="hidden" name="ps" :value="filterValue('PageSize')" />
                  <input v-if="filterValue('OrderBy')" type="hidden" name="od" :value="filterValue('OrderBy')" />
                  <input v-if="filterValue('Order')" type="hidden" name="order" :value="filterValue('Order')" />
                  <div
                    v-for="field in filterFields"
                    :key="field.name"
                    class="col-12 col-md-6 col-xl-2"
                    :class="{ 'smart-pick-date-field': field.placeholder === 'YYYY-MM-DD' }"
                  >
                    <label class="form-label mb-1">{{ field.label }}</label>
                    <input
                      class="form-control"
                      :class="{ 'smart-pick-date-input': field.placeholder === 'YYYY-MM-DD' }"
                      :name="field.name"
                      :value="field.value"
                      :placeholder="field.placeholder"
                      :inputmode="field.placeholder === 'YYYY-MM-DD' ? 'numeric' : undefined"
                      :maxlength="field.placeholder === 'YYYY-MM-DD' ? 10 : undefined"
                      :pattern="field.placeholder === 'YYYY-MM-DD' ? '[0-9]{4}-[0-9]{2}-[0-9]{2}' : undefined"
                    />
                  </div>
                  <div class="col-12 d-flex gap-2">
                    <button class="btn btn-primary" type="submit">筛选</button>
                    <a class="btn btn-outline-secondary" :href="filterClearHref">清空</a>
                  </div>
                </form>
              </div>
            </div>
          </div>

          <section class="next02-wdir-movies" aria-label="Movies">
            <MovieGrid v-if="isJAVDirectory && movies.length" :cards="movies" />
            <FC2MovieGrid v-else-if="isFC2Directory" :cards="movies" empty-text="这个目录下暂时没有 FC2 媒体。" />
            <div v-else-if="isMixedDirectory" class="movie-empty">请选择 JAV 或 FC2 子目录。</div>
            <div v-else class="movie-empty">这个目录下暂时没有媒体。</div>
          </section>
        </template>

        <template v-else>
          <section class="next02-wdir-overview next02-wdir-root-overview" aria-label="根目录概览">
            <div class="next02-wdir-overview-toolbar">
              <div class="next02-wdir-root-scope"><span>根目录</span><strong>完整目录树</strong></div>
              <div class="next02-wdir-root-status">
                <span class="next02-wdir-root-count"
                  ><strong>{{ total }}</strong> 个目录</span
                >
                <div
                  v-if="pageLinks.length"
                  class="cards-toolbar-pagination wdir-root-pagination next02-wdir-root-pagination next02-pink-cards-pagination"
                >
                  <PaginationBar :page-info="pageInfo" />
                </div>
              </div>
            </div>
            <ChildrenGrid :children="children" />
          </section>
        </template>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import "@/features/media/styles/index.css";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/features/media/api";
import ChildrenGrid from "@/features/media/components/directory/WDirChildrenGrid.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LinkStrip from "@/shared/components/LinkStrip.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import FC2MovieGrid from "@/features/fc2/components/FC2MovieGrid.vue";
import MovieGrid from "@/features/movies/components/MovieGrid.vue";
import MovieCardSortStrip from "@/features/movies/components/MovieCardSortStrip.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import type { PageInfo } from "@/shared/types/gc";
import type { MediaRecord } from "@/features/media/types";
import { asArray, asRecord, normalizeHref, pick, text } from "@/shared/utils/format";

const sortOptions = [
  { key: "ByDetailUpdate", label: "Detail更新" },
  { key: "ByReleasing", label: "上映日期" },
  { key: "ByViewer", label: "观看人数" },
  { key: "ByRankDate", label: "上榜日期" },
  { key: "ByDaysInRank", label: "在榜天数" },
  { key: "ByHighestRank", label: "最高排名" },
  { key: "ByBirthTime", label: "下载时间" },
  { key: "ByLastSelectedScTime", label: "最近入选 SC" },
  { key: "ByScTimes", label: "sc次数" },
  { key: "ByComeTimes", label: "co次数" },
  { key: "ByCastAgeDesc", label: "演员年龄" }
];

const route = useRoute();
const payload = ref<MediaRecord>({});
const loading = ref(false);
const error = ref("");
const filterOpen = ref(false);

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
    filterOpen.value = false;
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title") || "媒体目录");
const isDetail = computed(() => text(payload.value, "ViewType") === "wdir-detail");
const wdirVariantClass = computed(() => (isDetail.value ? "is-wdir-detail" : "is-wdir-root"));
const children = computed(() => asRecord(pick(payload.value, "Children")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<MediaRecord>(pageInfo.value.Links || pageInfo.value.links));
const total = computed(() => text(payload.value, "Total", "total") || "0");
const detail = computed(() => asRecord(pick(payload.value, "Detail")));
const directory = computed(() => asRecord(pick(detail.value, "directory", "Directory")));
const directoryName = computed(() => text(directory.value, "name", "Name") || title.value);
const pageHeadingTitle = computed(() => (isDetail.value ? directoryName.value : title.value));
const pageNote = computed(() =>
  isDetail.value ? "目录路径、JAV/FC2 媒体统计、子目录与影片。" : "按完整目录树浏览 w_folder / w_media / fc2_media。"
);
const directoryId = computed(() => text(directory.value, "id", "Id"));
const breadcrumbs = computed(() => asArray<MediaRecord>(pick(detail.value, "breadcrumbs", "Breadcrumbs")));
const stats = computed(() => asRecord(pick(detail.value, "stats", "Stats")));
const mediaType = computed(
  () => text(payload.value, "MediaType") || text(stats.value, "media_type", "MediaType") || "empty"
);
const isJAVDirectory = computed(() => mediaType.value === "jav");
const isFC2Directory = computed(() => mediaType.value === "fc2ppv");
const isMixedDirectory = computed(() => mediaType.value === "mixed");
const hasMovieDirectory = computed(() => isJAVDirectory.value || isFC2Directory.value);
const recursive = computed(() => Boolean(payload.value.Recursive));
const query = computed(() => text(payload.value, "Query"));
const movies = computed(() => asArray<MediaRecord>(pick(payload.value, "movies", "Movies")));
const sortQuery = computed(() => asRecord(pick(payload.value, "SortQuery", "sortQuery")));
const ownedQuery = computed(() => asRecord(pick(payload.value, "ownedQuery", "OwnedQuery")));
const filter = computed(() => asRecord(pick(payload.value, "MovieCardFilter")));
const filterAction = computed(() => text(filter.value, "Action") || "/mediadir");
const filterClearHref = computed(() => text(filter.value, "ClearHref") || "/mediadir");
const filterFields = computed(() =>
  [
    ["cn", "演员", "CastNames", "演员名，支持多值"],
    ["gn", "类型", "GenreNames", "类型名，支持多值"],
    ["dn", "导演", "DirectorName", "导演名"],
    ["pn", "前缀", "PrefixName", "前缀"],
    ["mn", "片商", "MakerName", "片商"],
    ["ln", "标签", "LabelName", "标签"],
    ["wd", "关键词", "Word", "中文关键词"],
    ["rs", "上映开始", "ReleasingDateStart", "YYYY-MM-DD"],
    ["re", "上映结束", "ReleasingDateEnd", "YYYY-MM-DD"],
    ["mbs", "Media 下载开始", "MediaBirthTimeStart", "YYYY-MM-DD"],
    ["mbe", "Media 下载结束", "MediaBirthTimeEnd", "YYYY-MM-DD"]
  ].map(([name, label, key, placeholder]) => ({ name, label, value: filterValue(key), placeholder }))
);
const statCards = computed(() => [
  { label: "媒体数", value: text(stats.value, "film_count", "FilmCount") || "0", dateTime: false },
  { label: "总大小", value: humanSize(pick(stats.value, "total_size", "TotalSize")), dateTime: false },
  { label: "最近下载时间", value: pick(stats.value, "last_film_birth", "LastFilmBirth"), dateTime: true },
  { label: "最近更新", value: pick(stats.value, "last_updated_on", "LastUpdatedOn"), dateTime: true }
]);
const ownedButtons = computed(() =>
  [
    ["MediaAll", "All"],
    ["MediaOwned", "Owned"],
    ["MediaNotOwned", "NotOwned"]
  ].map(([key, label]) => {
    const item = asRecord(ownedQuery.value[key]);
    return {
      key,
      label,
      href: normalizeHref(text(item, "Href", "href")),
      active: Boolean(item.Active || item.active),
      disabled: Boolean(item.Active || item.active)
    };
  })
);

function filterValue(key: string): string {
  return text(filter.value, key);
}

function toggleFilter(): void {
  filterOpen.value = !filterOpen.value;
}

function recursiveHref(next: boolean): string {
  return `/mediadir?id=${encodeURIComponent(directoryId.value)}&recursive=${next ? "1" : "0"}${query.value || ""}`;
}

function crumbName(item: MediaRecord): string {
  return text(item, "name", "Name");
}

function crumbId(item: MediaRecord): string {
  return encodeURIComponent(text(item, "id", "Id"));
}

function crumbKey(item: MediaRecord, index: number): string {
  return `${crumbId(item)}-${index}`;
}

function humanSize(input: unknown): string {
  const value = Number(input || 0);
  if (value >= 1024 ** 4) return `${(value / 1024 ** 4).toFixed(1)} TB`;
  if (value >= 1024 ** 3) return `${(value / 1024 ** 3).toFixed(1)} GB`;
  if (value >= 1024 ** 2) return `${(value / 1024 ** 2).toFixed(1)} MB`;
  if (value >= 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${value} B`;
}
</script>
