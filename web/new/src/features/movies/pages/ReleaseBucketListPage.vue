<template>
  <PageShell class="torrent-albums-page-shell release-bucket-page-shell" :title="pageTitle" :note="pageNote" eyebrow="">
    <div
      id="page-root"
      class="container-fluid torrent-albums-page release-bucket-page-root next02-bucket-rewrite bucket-list-standard"
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
            <p>{{ pageNote || "筛选、排序和下钻时间桶聚合结果。" }}</p>
          </div>
          <div class="bucket-list-hero-actions">
            <template v-if="!isWMediaBucket">
              <a
                v-for="item in modeSwitches"
                :key="linkKey(item)"
                :href="linkHref(item)"
                class="bucket-list-action"
                :class="{ 'is-active': linkActive(item) }"
                >{{ linkLabel(item) }}</a
              >
            </template>
            <a :href="aggPageHref" class="bucket-list-action" target="_blank" rel="noopener noreferrer">返回聚合页</a>
            <a href="/aggbackfill" class="bucket-list-action" target="_blank" rel="noopener noreferrer">回填入口</a>
          </div>
        </header>

        <section class="bucket-list-panel">
          <div class="bucket-list-controls">
            <div class="bucket-list-control-main">
              <form class="bucket-list-filter-row" method="get" :action="listPath">
                <input v-if="aggMode === 'owned'" type="hidden" name="agg_mode" value="owned" />
                <input type="hidden" name="sort" :value="queryValue('Sort')" />
                <input type="hidden" name="dir" :value="queryValue('Dir')" />
                <input type="hidden" name="level" :value="queryValue('Level')" />
                <label class="bucket-filter-field is-scope">
                  <span>时间桶</span>
                  <input
                    class="form-control"
                    type="text"
                    name="scope_key"
                    :value="queryValue('ScopeKey')"
                    placeholder="如 2026-04、2026-Q02"
                  />
                </label>
                <label class="bucket-filter-field is-year">
                  <span>Year</span>
                  <input
                    class="form-control"
                    type="number"
                    name="year"
                    :value="queryValue('Year')"
                    placeholder="如 2026"
                  />
                </label>
                <label class="bucket-filter-field is-short">
                  <span>Quarter</span>
                  <input
                    class="form-control"
                    type="number"
                    name="quarter"
                    :value="queryValue('Quarter')"
                    placeholder="1-4"
                  />
                </label>
                <label class="bucket-filter-field is-short">
                  <span>Month</span>
                  <input
                    class="form-control"
                    type="number"
                    name="month"
                    :value="queryValue('Month')"
                    placeholder="1-12"
                  />
                </label>
                <label v-if="!isWMediaBucket" class="bucket-filter-field is-short">
                  <span>Day</span>
                  <input class="form-control" type="number" name="day" :value="queryValue('Day')" placeholder="1-31" />
                </label>
                <label class="bucket-filter-field is-page-size">
                  <span>每页</span>
                  <input
                    class="form-control"
                    id="releaseBucketPageSize"
                    type="number"
                    step="1"
                    name="ps"
                    :value="queryValue('PageSize') || '24'"
                  />
                </label>
                <div class="bucket-list-filter-actions">
                  <button class="btn bucket-list-apply" type="submit">筛选</button>
                  <a class="btn bucket-list-clear" :href="clearHref">清空</a>
                </div>
              </form>

              <div class="bucket-list-switch-row">
                <div class="bucket-list-switch-group">
                  <span>排序</span>
                  <a
                    v-for="item in sortLinks"
                    :key="linkKey(item)"
                    :href="linkHref(item)"
                    class="bucket-list-switch"
                    :class="{ 'is-active': linkActive(item) }"
                    >{{ linkLabel(item) }}</a
                  >
                </div>
                <div class="bucket-list-switch-group bucket-level-switch">
                  <span>层级</span>
                  <a
                    v-for="item in levelLinks"
                    :key="linkKey(item)"
                    :href="linkHref(item)"
                    class="bucket-list-switch"
                    :class="{ 'is-active': linkActive(item) }"
                    >{{ linkLabel(item) }}</a
                  >
                </div>
              </div>
            </div>

            <div class="bucket-list-control-meta">
              <div class="bucket-list-summary">
                <span class="bucket-list-summary-value">{{ total }}</span>
                <span class="bucket-list-summary-label">Total Buckets</span>
              </div>
              <PaginationBar v-if="hasRows && pageLinks.length" :links="pageLinks" variant="bootstrap" />
            </div>
          </div>

          <div class="bucket-list-table-wrap">
            <table
              class="table table-hover align-middle mb-0 bucket-table"
              :class="isWMediaBucket ? 'is-wmedia-table' : 'is-release-table'"
            >
              <colgroup v-if="isWMediaBucket">
                <col class="bucket-col-name" />
                <col class="bucket-col-media" />
                <col class="bucket-col-removed" />
                <col class="bucket-col-size" />
                <col class="bucket-col-subtitle" />
                <col class="bucket-col-latest" />
              </colgroup>
              <colgroup v-else>
                <col class="bucket-col-name" />
                <col class="bucket-col-all" />
                <col class="bucket-col-owned" />
                <col class="bucket-col-size" />
                <col class="bucket-col-latest" />
              </colgroup>
              <thead>
                <tr>
                  <th class="text-nowrap">
                    <SortLink
                      class="bucket-sort-link"
                      active-class="is-active"
                      indicator-class="bucket-sort-indicator"
                      :link="headerLink('Bucket')"
                    />
                  </th>
                  <th v-for="key in metricHeaderKeys" :key="key" class="text-nowrap">
                    <SortLink
                      class="bucket-sort-link"
                      active-class="is-active"
                      indicator-class="bucket-sort-indicator"
                      :link="headerLink(key)"
                    />
                  </th>
                  <th class="text-nowrap">
                    <SortLink
                      class="bucket-sort-link"
                      active-class="is-active"
                      indicator-class="bucket-sort-indicator"
                      :link="headerLink('Size')"
                    />
                  </th>
                  <th v-if="isWMediaBucket" class="text-nowrap">
                    <SortLink
                      class="bucket-sort-link"
                      active-class="is-active"
                      indicator-class="bucket-sort-indicator"
                      :link="headerLink('Subtitle')"
                    />
                  </th>
                  <th class="text-nowrap">
                    <SortLink
                      class="bucket-sort-link"
                      active-class="is-active"
                      indicator-class="bucket-sort-indicator"
                      :link="headerLink(latestHeaderKey)"
                    />
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="!hasRows">
                  <td :colspan="tableColumnCount">
                    <div class="bucket-empty-state">{{ emptyText }}</div>
                  </td>
                </tr>
                <tr v-for="row in rows" v-else :key="rowKey(row)" :class="rowClass(row)">
                  <td class="text-nowrap" :title="rowText(row, 'BucketDisplay')">
                    <a
                      v-if="rowText(row, 'ViewHref')"
                      class="bucket-name-link"
                      :href="rowText(row, 'ViewHref')"
                      target="_blank"
                      rel="noopener noreferrer"
                      >{{ rowText(row, "BucketDisplay") }}</a
                    >
                    <span v-else class="bucket-name">{{ rowText(row, "BucketDisplay") }}</span>
                  </td>
                  <td v-for="key in metricValueKeys" :key="key" class="text-nowrap">{{ rowText(row, key) }}</td>
                  <td class="text-nowrap">{{ humanSize(rowValue(row, "SizeBytes")) }}</td>
                  <td v-if="isWMediaBucket" class="text-nowrap">{{ rowText(row, "HasSubCount") }}</td>
                  <td class="text-nowrap">
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
const pageLinks = computed(() => asArray<MovieRecord>(pageInfo.value.Links || pageInfo.value.links));
const modeSwitches = computed(() => asArray<MovieRecord>(pick(payload.value, "ModeSwitches")));
const sortLinks = computed(() => asArray<MovieRecord>(pick(payload.value, "SortLinks")));
const levelLinks = computed(() => asArray<MovieRecord>(pick(payload.value, "LevelLinks")));
const headerLinks = computed(() => asRecord(pick(payload.value, "HeaderLinks")));
const aggMode = computed(() => text(payload.value, "AggMode"));
const isWMediaBucket = computed(
  () => text(payload.value, "ViewType") === "w-media-bucket-list" || route.path === "/bucketsbirth"
);
const listPath = computed(
  () => text(payload.value, "ListPath") || (isWMediaBucket.value ? "/bucketsbirth" : "/bucketsrelease")
);
const clearHref = computed(() => text(payload.value, "ClearHref") || listPath.value);
const aggPageHref = computed(
  () => text(payload.value, "AggPageHref") || (isWMediaBucket.value ? "/aggbirth" : "/aggrelease")
);
const metricHeaderKeys = computed(() => (isWMediaBucket.value ? ["Media", "Removed"] : ["CountAll", "CountOwned"]));
const metricValueKeys = computed(() =>
  isWMediaBucket.value ? ["MediaCount", "RemovedCount"] : ["CountAll", "CountOwned"]
);
const latestHeaderKey = computed(() => (isWMediaBucket.value ? "LatestBirth" : "LatestRelease"));
const latestValueKey = computed(() => (isWMediaBucket.value ? "LatestBirthTime" : "LatestReleasingDate"));
const tableColumnCount = computed(() => (isWMediaBucket.value ? 6 : 5));
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
  return rowText(row, "ViewHref", "BucketDisplay") || JSON.stringify(row);
}

function rowClass(row: MovieRecord): string {
  if (aggMode.value !== "owned" || queryValue("Level") !== "day") return "";
  const n = Number(rowValue(row, "LatestReleasingDate") || 0);
  if (!n) return "";
  const month = new Date(n * 1000).getMonth() + 1;
  return month % 2 === 0 ? "bucket-month-even" : "bucket-month-odd";
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

function linkKey(item: MovieRecord): string {
  return `${linkLabel(item)}-${linkHref(item)}`;
}

function humanSize(input: unknown): string {
  const value = Number(input || 0);
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
