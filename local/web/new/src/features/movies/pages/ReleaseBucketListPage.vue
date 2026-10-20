<template>
  <PageShell class="torrent-albums-page-shell release-bucket-page-shell" :title="pageTitle" :note="pageNote" eyebrow="">
    <div
      id="page-root"
      class="container-fluid torrent-albums-page release-bucket-page-root next02-bucket-rewrite bucket-list-standard next02-sc-pink-theme next02-sc-events-rewrite"
      :class="
        isWMediaBucket
          ? 'w-media-bucket-list-page is-wmedia-bucket'
          : 'movie-release-bucket-list-page is-release-bucket'
      "
    >
      <div v-if="loading" class="alert bucket-list-feedback is-loading">加载中...</div>
      <div v-else-if="error" class="alert bucket-list-feedback is-error">{{ error }}</div>
      <template v-else>
        <header class="bucket-list-hero">
          <div class="bucket-list-hero-copy">
            <span class="bucket-list-eyebrow">{{ isWMediaBucket ? "WMedia Birth" : "Movie Release" }}</span>
            <h2>{{ pageTitle }}</h2>
            <p>{{ pageNote || "筛选、排序和查看时间桶统计结果。" }}</p>
          </div>
          <div class="bucket-list-hero-actions">
            <a href="/aggbackfill" class="bucket-list-action" target="_blank" rel="noopener noreferrer">回填入口</a>
          </div>
        </header>

        <section class="bucket-list-panel next02-sc-movie-table-card">
          <div
            class="rank-stats-toolbar-scope rank-page-shell cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell"
          >
            <form
              class="sticky-toolbar cards-toolbar rank-sticky-toolbar next02-pink-cards-toolbar rank-stats-toolbar"
              method="get"
              :action="listPath"
            >
              <input v-if="aggMode === 'owned'" type="hidden" name="agg_mode" value="owned" />
              <input type="hidden" name="sort" :value="queryValue('Sort')" />
              <input type="hidden" name="dir" :value="queryValue('Dir')" />
              <input type="hidden" name="level" :value="queryValue('Level')" />

              <div class="cards-toolbar-sort rank-toolbar-sort bucket-toolbar-primary">
                <div class="bucket-toolbar-group bucket-toolbar-level-group">
                  <span class="soft-badge rank-meta-badge">层级</span>
                  <LinkStrip :items="levelItems" />
                </div>
                <div v-if="!isWMediaBucket" class="bucket-toolbar-group bucket-toolbar-mode-group">
                  <LinkStrip :items="modeSwitchItems" />
                </div>
                <div class="rank-date-form rank-stats-date-controls bucket-toolbar-group bucket-toolbar-filter-group">
                  <label class="rank-stats-filter-field">
                    <span>时间桶</span>
                    <input
                      class="rank-date-input rank-bucket-scope-input"
                      type="text"
                      name="scope_key"
                      :value="queryValue('ScopeKey')"
                      placeholder="如 2026-04、2026-Q02"
                    />
                  </label>
                  <label class="rank-stats-filter-field">
                    <span>Year</span>
                    <input
                      class="rank-date-input rank-bucket-year-input"
                      type="number"
                      name="year"
                      :value="queryValue('Year')"
                      placeholder="如 2026"
                    />
                  </label>
                  <label class="rank-stats-filter-field">
                    <span>Quarter</span>
                    <input
                      class="rank-date-input rank-bucket-short-input"
                      type="number"
                      name="quarter"
                      :value="queryValue('Quarter')"
                      placeholder="1-4"
                    />
                  </label>
                  <label class="rank-stats-filter-field">
                    <span>Month</span>
                    <input
                      class="rank-date-input rank-bucket-short-input"
                      type="number"
                      name="month"
                      :value="queryValue('Month')"
                      placeholder="1-12"
                    />
                  </label>
                  <label v-if="!isWMediaBucket" class="rank-stats-filter-field">
                    <span>Day</span>
                    <input
                      class="rank-date-input rank-bucket-short-input"
                      type="number"
                      name="day"
                      :value="queryValue('Day')"
                      placeholder="1-31"
                    />
                  </label>
                  <label class="rank-stats-filter-field">
                    <span>每页</span>
                    <input
                      id="releaseBucketPageSize"
                      class="rank-date-input rank-stats-page-size-input"
                      type="number"
                      step="1"
                      name="ps"
                      :value="queryValue('PageSize') || '24'"
                    />
                  </label>
                  <div class="rank-navigation-group">
                    <button class="button-secondary rank-toolbar-action" type="submit">筛选</button>
                    <a class="button-secondary rank-toolbar-action" :href="clearHref">清空</a>
                  </div>
                </div>
              </div>

              <div v-if="hasRows" class="toolbar-line cards-toolbar-bottom rank-toolbar-line">
                <div class="next02-pink-cards-toolbar-status">
                  <div class="next02-pink-cards-metrics" aria-label="时间桶摘要">
                    <span :title="`总数 ${total}`" :aria-label="`总数 ${total}`">
                      <strong>{{ total }}</strong>
                    </span>
                    <span :title="`当前页 ${rows.length}`" :aria-label="`当前页 ${rows.length}`">
                      <strong>{{ rows.length }}</strong>
                    </span>
                  </div>
                  <div class="cards-toolbar-pagination next02-pink-cards-pagination">
                    <PaginationBar :page-info="pageInfo" />
                  </div>
                </div>
              </div>
            </form>
          </div>

          <div class="table-responsive bucket-list-table-wrap">
            <table
              class="table table-hover align-middle mb-0 sc-events-table bucket-table"
              :class="isWMediaBucket ? 'is-wmedia-table' : 'is-release-table'"
            >
              <colgroup v-if="isWMediaBucket">
                <col class="bucket-col-name" />
                <col class="bucket-col-media" />
                <col class="bucket-col-size" />
                <col class="bucket-col-removed" />
                <col class="bucket-col-size" />
                <col class="bucket-col-size" />
                <col class="bucket-col-subtitle" />
                <col class="bucket-col-latest" />
              </colgroup>
              <colgroup v-else>
                <col class="bucket-col-name" />
                <col class="bucket-col-all" />
                <col class="bucket-col-owned" />
                <col class="bucket-col-removed" />
                <col class="bucket-col-size" />
                <col class="bucket-col-latest" />
              </colgroup>
              <thead>
                <tr>
                  <th class="sc-sortable-head">
                    <SortLink class="sc-sort-link" :link="headerLink('Bucket')" />
                  </th>
                  <th v-for="column in metricColumns" :key="column.header" class="sc-number sc-sortable-head">
                    <SortLink class="sc-sort-link" :link="headerLink(column.header)" />
                  </th>
                  <th v-if="isWMediaBucket" class="sc-number sc-sortable-head">
                    <SortLink class="sc-sort-link" :link="headerLink('Subtitle')" />
                  </th>
                  <th class="sc-sortable-head">
                    <SortLink class="sc-sort-link" :link="headerLink(latestHeaderKey)" />
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="!hasRows">
                  <td :colspan="tableColumnCount">
                    <div class="sc-empty-state">{{ emptyText }}</div>
                  </td>
                </tr>
                <tr
                  v-for="(row, index) in rows"
                  v-else
                  :key="rowKey(row)"
                  class="next02-sc-event-row"
                  :class="rowClass(row, index)"
                >
                  <td class="sc-content-cell" :title="rowText(row, 'BucketDisplay')">
                    <span>{{ rowText(row, "BucketDisplay") }}</span>
                  </td>
                  <td v-for="column in metricColumns" :key="column.header" class="sc-number">
                    {{ column.size ? humanSize(rowValue(row, column.value)) : rowText(row, column.value) }}
                  </td>
                  <td v-if="isWMediaBucket" class="sc-number">{{ rowText(row, "HasSubCount") }}</td>
                  <td class="sc-muted">
                    <ScDateTime v-if="isWMediaBucket" :value="rowValue(row, latestValueKey)" unix />
                    <span v-else>{{ formatUnixDate(rowValue(row, latestValueKey)) }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import "@/features/movies/styles/index.css";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/features/movies/api";
import LinkStrip from "@/shared/components/LinkStrip.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import type { PageInfo } from "@/shared/types/gc";
import type { MovieRecord } from "@/features/movies/types";
import { asArray, asRecord, normalizeHref, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<MovieRecord>({});
const loading = ref(false);
const error = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title") || "Movie Release Bucket");
const pageNote = computed(() => text(payload.value, "PageNote"));
const rows = computed(() => asArray<MovieRecord>(pick(payload.value, "Rows", "rows")));
const query = computed(() => asRecord(pick(payload.value, "Query", "query")));
const total = computed(() => text(payload.value, "Total", "total") || rows.value.length);
const hasRows = computed(() => rows.value.length > 0);
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const modeSwitchItems = computed(() => bucketLinkItems(pick(payload.value, "ModeSwitches")));
const levelItems = computed(() => bucketLinkItems(pick(payload.value, "LevelLinks")));
const headerLinks = computed(() => asRecord(pick(payload.value, "HeaderLinks")));
const aggMode = computed(() => text(payload.value, "AggMode"));
const isWMediaBucket = computed(
  () => text(payload.value, "ViewType") === "w-media-bucket-list" || route.path === "/bucketsbirth"
);
const listPath = computed(
  () => text(payload.value, "ListPath") || (isWMediaBucket.value ? "/bucketsbirth" : "/bucketsrelease")
);
const clearHref = computed(() => text(payload.value, "ClearHref") || listPath.value);
const metricColumns = computed(() =>
  isWMediaBucket.value
    ? [
        { header: "Added", value: "AddedCount", size: false },
        { header: "AddedSize", value: "AddedSizeBytes", size: true },
        { header: "Removed", value: "RemovedCount", size: false },
        { header: "RemovedSize", value: "RemovedSizeBytes", size: true },
        { header: "NetSize", value: "NetSizeBytes", size: true }
      ]
    : [
        { header: "CountAll", value: "CountAll", size: false },
        { header: "CountOwned", value: "CountOwned", size: false },
        { header: "Removed", value: "RemovedCount", size: false },
        { header: "Size", value: "SizeBytes", size: true }
      ]
);
const latestHeaderKey = computed(() => (isWMediaBucket.value ? "LatestBirth" : "LatestRelease"));
const latestValueKey = computed(() => (isWMediaBucket.value ? "LatestBirthTime" : "LatestReleasingDate"));
const tableColumnCount = computed(() => (isWMediaBucket.value ? 8 : 6));
const emptyText = computed(() =>
  isWMediaBucket.value
    ? "当前筛选下没有 `w_media_birth_bucket_stat` 数据。若刚建表，请先去聚合回填页执行一次“回填 WMedia 时间聚合”。"
    : "当前筛选下没有 `movie_release_bucket_stat` 数据。若刚建表，请先去聚合回填页执行一次“回填上映日时间聚合”。"
);
function queryValue(key: string): string {
  return text(query.value, key);
}

function rowValue(row: MovieRecord, ...keys: string[]): unknown {
  return pick(row, ...keys);
}

function rowText(row: MovieRecord, ...keys: string[]): string {
  return text(row, ...keys);
}

function rowKey(row: MovieRecord): string {
  return rowText(row, "BucketDisplay") || JSON.stringify(row);
}

function bucketMonth(row: MovieRecord): number {
  const match = rowText(row, "BucketDisplay").match(/^\d{4}-(?:M)?(\d{2})(?:-|$)/);
  return Number(match?.[1] || 0);
}

function rowClass(row: MovieRecord, index: number): Record<string, boolean> {
  const month = bucketMonth(row);
  const previousMonth = index > 0 ? bucketMonth(rows.value[index - 1]) : 0;
  return {
    "sc-running-month-neutral": month <= 0,
    "sc-running-month-light": month > 0 && month % 2 === 0,
    "sc-running-month-deep": month > 0 && month % 2 !== 0,
    "sc-running-month-start": month > 0 && (index === 0 || month !== previousMonth)
  };
}

function headerLink(key: string): MovieRecord {
  return asRecord(headerLinks.value[key]);
}

function linkLabel(item: MovieRecord): string {
  return text(item, "Label", "label", "Name", "name", "Title", "title");
}

function linkHref(item: MovieRecord): string {
  return normalizeHref(text(item, "Href", "href") || "#");
}

function linkActive(item: MovieRecord): boolean {
  return Boolean(item.Active || item.active);
}

function bucketLinkItems(value: unknown) {
  return asArray<MovieRecord>(value).map((item) => ({
    label: linkLabel(item),
    href: linkHref(item),
    active: linkActive(item)
  }));
}

function humanSize(input: unknown): string {
  const value = Number(input || 0);
  if (value < 0) return `-${humanSize(-value)}`;
  if (value >= 1024 ** 4) return `${(value / 1024 ** 4).toFixed(1)} TB`;
  if (value >= 1024 ** 3) return `${(value / 1024 ** 3).toFixed(1)} GB`;
  if (value >= 1024 ** 2) return `${(value / 1024 ** 2).toFixed(1)} MB`;
  if (value >= 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${value} B`;
}

function formatUnixDate(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}
</script>
