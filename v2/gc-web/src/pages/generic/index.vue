<template>
  <PageFrame :class="pageFrameClass" :title="pageTitle" :eyebrow="response?.group || 'List'" :description="response?.description || ''">
    <template #actions v-if="navLinks.length > 0">
      <RouterLink v-for="link in navLinks" :key="link.href" class="rgc-button compact" :to="link.href">{{ link.label }}</RouterLink>
    </template>

    <section :class="listShellClass">
      <section v-if="isMovieReleaseAgg" class="rgc-release-agg-hero">
        <div class="rgc-release-agg-tabs">
          <button class="rgc-button compact" :class="{ primary: releaseMode === 'all' }" type="button" @click="setReleaseMode('all')">全量</button>
          <button class="rgc-button compact" :class="{ primary: releaseMode === 'owned' }" type="button" @click="setReleaseMode('owned')">已拥有</button>
        </div>
        <RouterLink class="rgc-button compact" :to="releaseBucketListHref">去明细入口</RouterLink>
        <div class="rgc-release-agg-title">
          <span>上映日（{{ releaseModeLabel }}）</span>
          <b>时间聚合</b>
        </div>
        <nav v-if="releaseBreadcrumbs.length > 0" class="rgc-release-breadcrumbs">
          <template v-for="(item, index) in releaseBreadcrumbs" :key="`${item.label}-${index}`">
            <RouterLink v-if="item.href" :to="item.href">{{ item.label }}</RouterLink>
            <span v-else>{{ item.label }}</span>
            <small v-if="index < releaseBreadcrumbs.length - 1">/</small>
          </template>
        </nav>
      </section>

      <form v-else class="rgc-list-toolbar" @submit.prevent="apply">
        <div class="rgc-list-toolbar-main">
          <strong>{{ response?.title || route.meta.title || "List" }}</strong>
          <span>{{ toolbarSummary }}</span>
        </div>

        <div class="rgc-list-filter-grid">
          <label v-for="filter in formFilters" :key="filter.key" :class="{ compact: filter.type === 'select' }">
            <span>{{ filter.label }}</span>
            <select v-if="filter.type === 'select'" v-model="filterValues[filter.key]">
              <option v-for="option in filter.options || []" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
            <input
              v-else
              v-model="filterValues[filter.key]"
              :type="inputType(filter.type)"
              :placeholder="filter.placeholder"
              :inputmode="filter.type === 'number' ? 'decimal' : undefined"
            />
          </label>
          <label class="compact">
            <span>每页</span>
            <input v-model="pageSize" type="number" inputmode="numeric" />
          </label>
        </div>

        <div class="rgc-list-toolbar-actions">
          <button class="rgc-button compact primary" type="submit">筛选</button>
          <button class="rgc-button compact" type="button" @click="clear">清空</button>
        </div>
      </form>

      <StateBlock v-if="loading" title="加载中" message="正在读取页面数据。" />
      <StateBlock v-else-if="errorMessage" title="加载失败" :message="errorMessage" tone="error" />
      <template v-else-if="response">
        <div v-if="!isCastDetail && !isMovieReleaseAgg" class="rgc-list-controlbar">
          <div class="rgc-list-counts">
            <b>{{ pageWindowText }}</b>
            <span>排序 {{ response.order_by || "-" }} {{ response.order || "" }}</span>
          </div>
          <PaginationControl :page="response.page" :page-size="response.page_size" :total="response.total" @go="goPage" />
        </div>

        <section v-if="isAggOverview" :class="['rgc-agg-overview', { 'release': isMovieReleaseAgg }]">
          <template v-if="isMovieReleaseAgg">
            <section class="rgc-release-agg-panel">
              <div class="rgc-release-agg-panel-head">
                <h2>时间聚合</h2>
                <span>{{ activeReleaseBucketSection?.rows.length || 0 }}</span>
              </div>
              <div v-if="releaseBucketSections.length > 1" class="rgc-release-tabbar">
                <button
                  v-for="section in releaseBucketSections"
                  :key="section.key"
                  type="button"
                  :class="{ active: activeReleaseBucketKey === section.key }"
                  @click="activeReleaseBucketKey = section.key"
                >
                  {{ section.label }}
                </button>
              </div>
              <div v-if="activeReleaseBucketSection" :class="releaseBucketGridClass">
                <RouterLink v-for="row in activeReleaseBucketSection.rows" :key="row.id || row.label" class="rgc-release-year-card" :to="row.href || route.path">
                  <strong>{{ row.label || "-" }}</strong>
                  <span>总数：{{ row.count_all || 0 }}，已拥有：{{ row.count_owned || 0 }}</span>
                  <small>{{ formatBytes(Number(row.size_bytes || 0)) }}</small>
                </RouterLink>
              </div>
              <StateBlock v-else title="暂无聚合数据" message="当前上映日范围没有时间桶。" tone="empty" />
            </section>

            <section v-if="releaseTopSections.length > 0" class="rgc-release-agg-panel">
              <div class="rgc-release-agg-panel-head">
                <h2>Top 聚合（{{ releaseModeLabel }}）</h2>
                <button class="rgc-button compact" type="button" @click="releaseTopOpen = !releaseTopOpen">
                  {{ releaseTopOpen ? "收起" : "展开" }}
                </button>
              </div>
              <div v-if="releaseTopOpen" class="rgc-release-top-tabs">
                <div class="rgc-release-tabbar">
                  <button
                    v-for="section in releaseTopSections"
                    :key="section.key"
                    type="button"
                    :class="{ active: activeReleaseTopKey === section.key }"
                    @click="activeReleaseTopKey = section.key"
                  >
                    {{ section.label }}
                  </button>
                </div>
                <div v-if="activeReleaseTopSection" class="rgc-release-top-grid">
                  <template v-for="(row, index) in activeReleaseTopSection.rows" :key="row.id || row.label">
                    <RouterLink v-if="row.href" class="rgc-release-top-card" :to="row.href">
                      <span><b>#{{ row.rank_no || index + 1 }}</b>{{ row.label || "-" }}</span>
                      <small>总：{{ row.count_all || 0 }}</small>
                      <small>拥有：{{ row.count_owned || 0 }}</small>
                    </RouterLink>
                    <div v-else class="rgc-release-top-card">
                      <span><b>#{{ row.rank_no || index + 1 }}</b>{{ row.label || "-" }}</span>
                      <small>总：{{ row.count_all || 0 }}</small>
                      <small>拥有：{{ row.count_owned || 0 }}</small>
                    </div>
                  </template>
                </div>
              </div>
            </section>
          </template>

          <template v-else>
            <section v-for="section in aggSections" :key="section.key" class="rgc-agg-section">
              <div class="rgc-agg-section-head">
                <h2>{{ section.label }}</h2>
                <span>{{ section.rows.length }}</span>
              </div>
              <div :class="['rgc-agg-grid', { top: section.key !== 'Year' }]">
                <template v-for="row in section.rows" :key="row.id || row.label">
                  <RouterLink v-if="row.href" class="rgc-agg-tile" :to="row.href">
                    <strong>{{ row.label || "-" }}</strong>
                    <span>{{ aggPrimaryMetric(row) }}</span>
                    <small>{{ aggSecondaryMetric(row) }}</small>
                  </RouterLink>
                  <div v-else class="rgc-agg-tile">
                    <strong>{{ row.label || "-" }}</strong>
                    <span>{{ aggPrimaryMetric(row) }}</span>
                    <small>{{ aggSecondaryMetric(row) }}</small>
                  </div>
                </template>
              </div>
            </section>
          </template>

          <section v-if="isWMediaBirth" class="rgc-cast-card-section">
            <div class="rgc-list-controlbar">
              <div class="rgc-list-counts">
                <b>{{ aggCardWindowText }}</b>
                <span>Movie Cards</span>
              </div>
              <PaginationControl v-if="aggCards" :page="aggCards.page" :page-size="aggCards.page_size" :total="aggCards.total" @go="goAggCardPage" />
            </div>
            <StateBlock v-if="aggCardsLoading" title="加载中" message="正在读取聚合影片。" />
            <StateBlock v-else-if="aggCardsError" title="加载失败" :message="aggCardsError" tone="error" />
            <MovieGrid v-else-if="aggCards && aggCards.items.length > 0" :items="aggCards.items" />
            <StateBlock v-else title="无影片" message="当前聚合条件没有影片。" tone="empty" />
          </section>

          <section v-if="isMovieReleaseAgg && releaseShowsMovies" class="rgc-cast-card-section">
            <div class="rgc-list-controlbar">
              <div class="rgc-list-counts">
                <b>{{ aggCardWindowText }}</b>
                <span>Movie Cards</span>
              </div>
              <PaginationControl v-if="aggCards" :page="aggCards.page" :page-size="aggCards.page_size" :total="aggCards.total" @go="goAggCardPage" />
            </div>
            <StateBlock v-if="aggCardsLoading" title="加载中" message="正在读取上映日影片。" />
            <StateBlock v-else-if="aggCardsError" title="加载失败" :message="aggCardsError" tone="error" />
            <MovieGrid v-else-if="aggCards && aggCards.items.length > 0" :items="aggCards.items" />
            <StateBlock v-else title="无影片" message="当前上映日范围没有影片。" tone="empty" />
          </section>
        </section>

        <DataTable
          v-else-if="response.rows.length > 0"
          :columns="response.columns"
          :rows="response.rows"
          :order-by="response.order_by"
          :order="response.order"
          @sort="sortBy"
        />
        <StateBlock v-else title="无数据" message="当前筛选条件没有行。" tone="empty" />

        <section v-if="isCastDetail" class="rgc-cast-card-section">
          <div class="rgc-list-controlbar">
            <div class="rgc-list-counts">
              <b>{{ castCardWindowText }}</b>
              <span>Movie Cards</span>
            </div>
            <PaginationControl
              v-if="castCards"
              :page="castCards.page"
              :page-size="castCards.page_size"
              :total="castCards.total"
              @go="goCastCardPage"
            />
          </div>
          <StateBlock v-if="castCardsLoading" title="加载中" message="正在读取演员影片。" />
          <StateBlock v-else-if="castCardsError" title="加载失败" :message="castCardsError" tone="error" />
          <MovieGrid v-else-if="castCards && castCards.items.length > 0" :items="castCards.items" />
          <StateBlock v-else title="无影片" message="当前演员没有匹配影片。" tone="empty" />
        </section>

        <div v-else-if="!isAggOverview" class="rgc-list-controlbar bottom">
          <span>{{ pageWindowText }}</span>
          <PaginationControl :page="response.page" :page-size="response.page_size" :total="response.total" @go="goPage" />
        </div>
      </template>
    </section>
  </PageFrame>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { getGenericPage, listCards } from "@/api/gc";
import DataTable from "@/components/DataTable.vue";
import MovieGrid from "@/components/MovieGrid.vue";
import PageFrame from "@/components/PageFrame.vue";
import PaginationControl from "@/components/PaginationControl.vue";
import StateBlock from "@/components/StateBlock.vue";
import type { CardsListResponse, PageFilter, PageLink, PageListResponse } from "@/types/api";

const route = useRoute();
const router = useRouter();
const response = ref<PageListResponse | null>(null);
const castCards = ref<CardsListResponse | null>(null);
const aggCards = ref<CardsListResponse | null>(null);
const loading = ref(false);
const castCardsLoading = ref(false);
const aggCardsLoading = ref(false);
const errorMessage = ref("");
const castCardsError = ref("");
const aggCardsError = ref("");
const filterValues = reactive<Record<string, string>>({});
const pageSize = ref("");

const pageKey = computed(() => String(route.meta.pageKey || ""));
const isCastDetail = computed(() => pageKey.value === "cast-detail");
const isWMediaBirth = computed(() => pageKey.value === "w-media-agg-birth");
const isMovieReleaseAgg = computed(() => pageKey.value === "movie-agg-all-release");
const isMovieReleaseBucket = computed(() => pageKey.value === "movie-release-bucket-list");
const isAggOverview = computed(() => isWMediaBirth.value || isMovieReleaseAgg.value);
const pageTitle = computed(() => String(response.value?.title || route.meta.title || "Page"));
const activeFilters = computed<PageFilter[]>(() => response.value?.filters || [{ key: "q", label: "关键词", type: "text" }]);
const formFilters = computed<PageFilter[]>(() => {
  if (isMovieReleaseBucket.value) {
    return activeFilters.value.filter((filter) => filter.key !== "q");
  }
  return activeFilters.value;
});
const pageFrameClass = computed(() => ({ "rgc-release-bucket-page": isMovieReleaseBucket.value, "rgc-release-agg-page": isMovieReleaseAgg.value }));
const listShellClass = computed(() => ["rgc-list-shell", { "rgc-release-bucket-shell": isMovieReleaseBucket.value, "rgc-release-agg-shell": isMovieReleaseAgg.value }]);
const navLinks = computed<PageLink[]>(() => response.value?.links || []);
const releaseTopOpen = ref(false);
const activeReleaseTopKey = ref("");
const activeReleaseBucketKey = ref("");
const releaseMode = computed(() => String(route.query.agg_mode || filterValues.agg_mode || "all") === "owned" ? "owned" : "all");
const releaseModeLabel = computed(() => (releaseMode.value === "owned" ? "已拥有" : "全量"));
const toolbarSummary = computed(() => {
  if (!response.value) {
    return "等待数据";
  }
  return `${response.value.total} rows · page ${response.value.page}`;
});
const pageWindowText = computed(() => {
  if (!response.value || response.value.total === 0) {
    return "0 rows";
  }
  const start = (response.value.page - 1) * response.value.page_size + 1;
  const end = Math.min(response.value.total, response.value.page * response.value.page_size);
  return `${start}-${end} / ${response.value.total}`;
});
const castCardWindowText = computed(() => {
  if (!castCards.value || castCards.value.total === 0) {
    return "0 movies";
  }
  const start = (castCards.value.page - 1) * castCards.value.page_size + 1;
  const end = Math.min(castCards.value.total, castCards.value.page * castCards.value.page_size);
  return `${start}-${end} / ${castCards.value.total}`;
});
const aggCardWindowText = computed(() => {
  if (!aggCards.value || aggCards.value.total === 0) {
    return "0 movies";
  }
  const start = (aggCards.value.page - 1) * aggCards.value.page_size + 1;
  const end = Math.min(aggCards.value.total, aggCards.value.page * aggCards.value.page_size);
  return `${start}-${end} / ${aggCards.value.total}`;
});
const aggSections = computed(() => {
  const groups = new Map<string, PageListResponse["rows"]>();
  for (const row of response.value?.rows || []) {
    const key = row.section || "Other";
    groups.set(key, [...(groups.get(key) || []), row]);
  }
  return Array.from(groups.entries()).map(([key, rows]) => ({ key, label: aggSectionLabel(key), rows }));
});
const releaseBucketSections = computed(() => aggSections.value.filter((section) => ["Year", "Quarter", "Month", "Day"].includes(section.key)));
const releaseTopSections = computed(() => aggSections.value.filter((section) => section.key.startsWith("Top ")));
const activeReleaseBucketSection = computed(() => releaseBucketSections.value.find((section) => section.key === activeReleaseBucketKey.value) || releaseBucketSections.value[0] || null);
const activeReleaseTopSection = computed(() => releaseTopSections.value.find((section) => section.key === activeReleaseTopKey.value) || releaseTopSections.value[0] || null);
const releaseBucketGridClass = computed(() => ["rgc-release-year-grid", `level-${activeReleaseBucketSection.value?.key.toLowerCase() || "empty"}`]);
const releaseLevel = computed(() => statValue("level") || inferReleaseLevelFromQuery());
const releaseRangeStart = computed(() => statValue("range_start"));
const releaseRangeEnd = computed(() => statValue("range_end"));
const releaseShowsMovies = computed(() => releaseLevel.value !== "" && releaseLevel.value !== "root");
const releaseBucketListHref = computed(() => (releaseMode.value === "owned" ? "/movie-agg-all/release-bucket-list?agg_mode=owned" : "/movie-agg-all/release-bucket-list"));
const releaseBreadcrumbs = computed(() => buildReleaseBreadcrumbs());

watch(() => route.fullPath, load);
watch(releaseBucketSections, (sections) => {
  if (!sections.some((section) => section.key === activeReleaseBucketKey.value)) {
    activeReleaseBucketKey.value = sections[0]?.key || "";
  }
}, { immediate: true });
watch(releaseTopSections, (sections) => {
  if (!sections.some((section) => section.key === activeReleaseTopKey.value)) {
    activeReleaseTopKey.value = sections[0]?.key || "";
  }
}, { immediate: true });
onMounted(load);

async function load() {
  loading.value = true;
  errorMessage.value = "";
  pageSize.value = String(route.query.ps ?? "");
  try {
    const data = await getGenericPage(pageKey.value, buildRouteParams());
    response.value = data;
    syncFilters(data.filters);
    await loadCastCards();
    await loadAggCards();
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function buildRouteParams() {
  const params: Record<string, unknown> = { ...route.query };
  if (isCastDetail.value || isWMediaBirth.value) {
    params.p = "1";
  }
  if (route.params.id) {
    params.id = route.params.id;
  }
  return params;
}

async function loadAggCards() {
  if (!isWMediaBirth.value && !(isMovieReleaseAgg.value && releaseShowsMovies.value)) {
    aggCards.value = null;
    aggCardsError.value = "";
    aggCardsLoading.value = false;
    return;
  }
  aggCardsLoading.value = true;
  aggCardsError.value = "";
  try {
    if (isMovieReleaseAgg.value) {
      aggCards.value = await listCards(buildReleaseCardParams());
    } else {
      aggCards.value = await listCards({
        ...route.query,
        view: "cardsmediamowned",
        mowned: route.query.mowned || "3",
        od: route.query.od || "mbt",
        ps: route.query.ps || "18",
      });
    }
  } catch (error) {
    aggCardsError.value = error instanceof Error ? error.message : "加载失败";
  } finally {
    aggCardsLoading.value = false;
  }
}

async function loadCastCards() {
  if (!isCastDetail.value || !route.params.id) {
    castCards.value = null;
    castCardsError.value = "";
    castCardsLoading.value = false;
    return;
  }
  castCardsLoading.value = true;
  castCardsError.value = "";
  try {
    castCards.value = await listCards({
      ...route.query,
      pid: String(route.params.id),
      view: "cards",
      ps: route.query.ps || "18",
    });
  } catch (error) {
    castCardsError.value = error instanceof Error ? error.message : "加载失败";
  } finally {
    castCardsLoading.value = false;
  }
}

function syncFilters(filters: PageFilter[]) {
  const activeKeys = new Set(filters.map((filter) => filter.key));
  Object.keys(filterValues).forEach((key) => {
    if (!activeKeys.has(key)) {
      delete filterValues[key];
    }
  });
  filters.forEach((filter) => {
    if (filter.key === "id" && route.params.id) {
      filterValues[filter.key] = String(route.params.id);
      return;
    }
    filterValues[filter.key] = String(route.query[filter.key] ?? filter.value ?? "");
  });
}

function inputType(type: string) {
  if (type === "date" || type === "number") {
    return type;
  }
  return "text";
}

function apply() {
  const query: Record<string, string | number> = { p: 1 };
  formFilters.value.forEach((filter) => {
    const value = filterValues[filter.key] || "";
    if (value !== "") {
      query[filter.key] = value;
    }
  });
  if (pageSize.value !== "") {
    query.ps = pageSize.value;
  }
  router.push({ path: route.path, query });
}

function clear() {
  router.push({ path: route.path, query: {} });
}

function sortBy(key: string) {
  const currentOrder = String(response.value?.order || "").toLowerCase();
  const nextOrder = response.value?.order_by === key && currentOrder === "asc" ? "desc" : "asc";
  router.push({ path: route.path, query: compact({ ...route.query, sort: key, order: nextOrder, p: 1 }) });
}

function goPage(page: number) {
  router.push({ path: route.path, query: compact({ ...route.query, p: page }) });
}

function goCastCardPage(page: number) {
  router.push({ path: route.path, query: compact({ ...route.query, p: page }) });
}

function goAggCardPage(page: number) {
  router.push({ path: route.path, query: compact({ ...route.query, p: page }) });
}

function setReleaseMode(mode: "all" | "owned") {
  releaseTopOpen.value = false;
  const query = compact({ ...route.query, agg_mode: mode === "owned" ? "owned" : "all", p: 1 });
  router.push({ path: route.path, query });
}

function statValue(label: string) {
  return response.value?.stats.find((item) => item.label === label)?.value || "";
}

function inferReleaseLevelFromQuery() {
  if (!isMovieReleaseAgg.value) {
    return "";
  }
  if (!route.query.year) {
    return "root";
  }
  if (route.query.day) {
    return "day";
  }
  if (route.query.month) {
    return "month";
  }
  if (route.query.quarter) {
    return "quarter";
  }
  return "year";
}

function buildReleaseCardParams() {
  return compact({
    ...route.query,
    view: "cards",
    mowned: releaseMode.value === "owned" ? "3" : "1",
    rs: releaseRangeStart.value,
    re: releaseRangeEnd.value,
    od: route.query.od || "rd",
    ps: route.query.ps || "18",
  });
}

function buildReleaseBreadcrumbs() {
  if (!isMovieReleaseAgg.value) {
    return [];
  }
  const year = Number(route.query.year || 0);
  const quarter = Number(route.query.quarter || 0);
  const month = Number(route.query.month || 0);
  const day = Number(route.query.day || 0);
  const rootHref = buildReleaseHref({});
  const crumbs: Array<{ label: string; href: string }> = [{ label: `上映日（${releaseModeLabel.value}）`, href: year > 0 ? rootHref : "" }];
  if (year <= 0) {
    return crumbs;
  }
  crumbs.push({ label: `${year} 年`, href: quarter > 0 || month > 0 || day > 0 ? buildReleaseHref({ year }) : "" });
  const quarterValue = quarter || (month > 0 ? Math.floor((month - 1) / 3) + 1 : 0);
  if (day > 0) {
    crumbs.push({ label: `Q${quarterValue} 季`, href: buildReleaseHref({ year, quarter: quarterValue }) });
    crumbs.push({ label: `${String(month).padStart(2, "0")} 月`, href: buildReleaseHref({ year, month }) });
    crumbs.push({ label: `${String(day).padStart(2, "0")} 日`, href: "" });
    return crumbs;
  }
  if (month > 0) {
    crumbs.push({ label: `Q${quarterValue} 季`, href: buildReleaseHref({ year, quarter: quarterValue }) });
    crumbs.push({ label: `${String(month).padStart(2, "0")} 月`, href: "" });
    return crumbs;
  }
  if (quarter > 0) {
    crumbs.push({ label: `Q${quarter} 季`, href: "" });
  }
  return crumbs;
}

function buildReleaseHref(parts: { year?: number; quarter?: number; month?: number; day?: number }) {
  const query = compact({
    agg_mode: releaseMode.value === "owned" ? "owned" : "",
    year: parts.year || "",
    quarter: parts.quarter || "",
    month: parts.month || "",
    day: parts.day || "",
  });
  const params = new URLSearchParams(query as Record<string, string>);
  const encoded = params.toString();
  return encoded ? `/movie-agg-all/release?${encoded}` : "/movie-agg-all/release";
}

function aggSectionLabel(key: string) {
  const labels: Record<string, string> = {
    Year: "年份",
    Quarter: "季度",
    Month: "月份",
    Day: "日期",
    "Top Cast": "演员",
    "Top Director": "导演",
    "Top Label": "Label",
    "Top Prefix": "Prefix",
  };
  return labels[key] || key;
}

function aggPrimaryMetric(row: Record<string, string>) {
  if (row.media_count) {
    return `${row.media_count} media`;
  }
  if (row.count_all) {
    return `${row.count_all} all`;
  }
  return "-";
}

function aggSecondaryMetric(row: Record<string, string>) {
  const items: string[] = [];
  if (row.count_owned) {
    items.push(`${row.count_owned} owned`);
  }
  if (row.size_bytes && row.size_bytes !== "0") {
    items.push(formatBytes(Number(row.size_bytes)));
  }
  if (row.latest_time && row.latest_time !== "0") {
    items.push(formatUnixDate(Number(row.latest_time)));
  }
  return items.join(" · ") || "-";
}

function formatBytes(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return "-";
  }
  const units = ["B", "KB", "MB", "GB", "TB"];
  let current = value;
  let unit = 0;
  while (current >= 1024 && unit < units.length - 1) {
    current /= 1024;
    unit += 1;
  }
  return `${current.toFixed(current >= 10 ? 1 : 2)} ${units[unit]}`;
}

function formatUnixDate(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return "-";
  }
  const date = new Date(value * 1000);
  const yyyy = date.getFullYear();
  const mm = String(date.getMonth() + 1).padStart(2, "0");
  const dd = String(date.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function compact(query: Record<string, unknown>) {
  return Object.fromEntries(Object.entries(query).filter(([, value]) => value !== "" && value !== undefined && value !== null));
}
</script>
