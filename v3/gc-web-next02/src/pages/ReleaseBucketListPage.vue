<template>
  <PageShell class="torrent-albums-page-shell release-bucket-page-shell" :title="pageTitle" :note="pageNote" eyebrow="">
    <div id="page-root" class="container-fluid torrent-albums-page release-bucket-page-root" :class="isWMediaBucket ? 'w-media-bucket-list-page' : 'movie-release-bucket-list-page'">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div class="d-flex flex-wrap align-items-center gap-2 mb-3">
          <div class="d-flex flex-wrap gap-2">
            <a v-for="item in modeSwitches" :key="linkKey(item)" :href="linkHref(item)" class="btn btn-sm" :class="linkActive(item) ? 'btn-primary' : 'btn-outline-primary'">{{ linkLabel(item) }}</a>
            <a :href="aggPageHref" class="btn btn-outline-primary btn-sm">返回聚合页</a>
            <a href="/triggers/agg" class="btn btn-outline-secondary btn-sm">去回填入口</a>
          </div>
        </div>

        <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
          <div class="bucket-list-summary">
            <span class="bucket-list-summary-label">总行数</span>
            <span class="bucket-list-summary-value">{{ total }}</span>
          </div>
          <div v-if="hasRows && pageLinks.length" class="ms-xl-auto">
            <PageLinks :links="pageLinks" />
          </div>
        </div>

        <div class="card mb-3">
          <div class="card-body">
            <form class="row g-2 align-items-end" method="get" :action="listPath">
              <input v-if="aggMode === 'owned'" type="hidden" name="agg_mode" value="owned">
              <input type="hidden" name="sort" :value="queryValue('Sort')">
              <input type="hidden" name="dir" :value="queryValue('Dir')">
              <input type="hidden" name="level" :value="queryValue('Level')">
              <div class="col-12 col-md-6 col-xl-4">
                <label class="form-label mb-1">时间桶关键字</label>
                <input class="form-control" type="text" name="scope_key" :value="queryValue('ScopeKey')" placeholder="如 2026-04、2026-Q02">
              </div>
              <div class="col-12 col-md-4 col-xl-2">
                <label class="form-label mb-1">Year</label>
                <input class="form-control" type="number" name="year" :value="queryValue('Year')" placeholder="如 2026">
              </div>
              <div class="col-12 col-md-4 col-xl-1">
                <label class="form-label mb-1">Quarter</label>
                <input class="form-control" type="number" name="quarter" :value="queryValue('Quarter')" placeholder="1-4">
              </div>
              <div class="col-12 col-md-4 col-xl-1">
                <label class="form-label mb-1">Month</label>
                <input class="form-control" type="number" name="month" :value="queryValue('Month')" placeholder="1-12">
              </div>
              <div v-if="!isWMediaBucket" class="col-12 col-md-4 col-xl-1">
                <label class="form-label mb-1">Day</label>
                <input class="form-control" type="number" name="day" :value="queryValue('Day')" placeholder="1-31">
              </div>
              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">每页条数</label>
                <select class="form-select" id="releaseBucketPageSize" name="ps" :value="queryValue('PageSize') || '24'">
                  <option value="24">24</option>
                  <option value="50">50</option>
                  <option value="100">100</option>
                  <option value="200">200</option>
                </select>
              </div>
              <div class="col-12 col-xl-2 d-flex gap-2">
                <button class="btn btn-primary" type="submit">筛选</button>
                <a class="btn btn-outline-secondary" :href="clearHref">清空</a>
              </div>
            </form>
          </div>
        </div>

        <div class="card mb-3">
          <div class="card-body">
            <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3">
              <div class="d-flex flex-wrap align-items-center gap-2">
                <span class="text-muted small me-1">排序</span>
                <a v-for="item in sortLinks" :key="linkKey(item)" :href="linkHref(item)" class="btn btn-sm" :class="linkActive(item) ? 'btn-primary' : 'btn-outline-primary'">{{ linkLabel(item) }}</a>
              </div>
              <div class="d-flex flex-wrap align-items-center justify-content-xl-end gap-2 bucket-level-switch">
                <span class="text-muted small me-1">层级</span>
                <a v-for="item in levelLinks" :key="linkKey(item)" :href="linkHref(item)" class="btn btn-sm" :class="linkActive(item) ? 'btn-primary' : 'btn-outline-primary'">{{ linkLabel(item) }}</a>
              </div>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 bucket-table">
                <thead>
                  <tr>
                    <th class="text-nowrap"><HeaderLink :link="headerLink('Bucket')" /></th>
                    <th v-for="key in metricHeaderKeys" :key="key" class="text-nowrap"><HeaderLink :link="headerLink(key)" /></th>
                    <th class="text-nowrap"><HeaderLink :link="headerLink('Size')" /></th>
                    <th v-if="isWMediaBucket" class="text-nowrap"><HeaderLink :link="headerLink('Subtitle')" /></th>
                    <th class="text-nowrap"><HeaderLink :link="headerLink(latestHeaderKey)" /></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td :colspan="tableColumnCount"><div class="bucket-empty-state">{{ emptyText }}</div></td>
                  </tr>
                  <tr v-for="row in rows" v-else :key="rowKey(row)" :class="rowClass(row)">
                    <td class="text-nowrap">
                      <a v-if="rowText(row, 'ViewHref')" class="bucket-name-link" :href="rowText(row, 'ViewHref')">{{ rowText(row, "BucketDisplay") }}</a>
                      <span v-else class="bucket-name">{{ rowText(row, "BucketDisplay") }}</span>
                    </td>
                    <td v-for="key in metricValueKeys" :key="key" class="text-nowrap">{{ rowText(row, key) }}</td>
                    <td class="text-nowrap">{{ humanSize(rowValue(row, "SizeBytes")) }}</td>
                    <td v-if="isWMediaBucket" class="text-nowrap">{{ rowText(row, "HasSubCount") }}</td>
                    <td class="text-nowrap">{{ formatUnixDate(rowValue(row, latestValueKey), isWMediaBucket) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div v-if="hasRows && pageLinks.length" class="d-flex justify-content-end mt-3">
          <PageLinks :links="pageLinks" />
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord, PageInfo } from "@/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
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
const rows = computed(() => asArray<AnyRecord>(pick(payload.value, "Rows", "rows")));
const query = computed(() => asRecord(pick(payload.value, "Query", "query")));
const total = computed(() => text(payload.value, "Total", "total") || rows.value.length);
const hasRows = computed(() => rows.value.length > 0);
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<AnyRecord>(pageInfo.value.Links || pageInfo.value.links));
const modeSwitches = computed(() => asArray<AnyRecord>(pick(payload.value, "ModeSwitches")));
const sortLinks = computed(() => asArray<AnyRecord>(pick(payload.value, "SortLinks")));
const levelLinks = computed(() => asArray<AnyRecord>(pick(payload.value, "LevelLinks")));
const headerLinks = computed(() => asRecord(pick(payload.value, "HeaderLinks")));
const aggMode = computed(() => text(payload.value, "AggMode"));
const isWMediaBucket = computed(() => text(payload.value, "ViewType") === "w-media-bucket-list" || route.path === "/w-media-agg/bucket-list");
const listPath = computed(() => text(payload.value, "ListPath") || (isWMediaBucket.value ? "/w-media-agg/bucket-list" : "/movie-agg-all/release-bucket-list"));
const clearHref = computed(() => text(payload.value, "ClearHref") || listPath.value);
const aggPageHref = computed(() => text(payload.value, "AggPageHref") || (isWMediaBucket.value ? "/w-media-agg/birth" : "/movie-agg-all/release"));
const metricHeaderKeys = computed(() => isWMediaBucket.value ? ["Media", "Removed"] : ["CountAll", "CountOwned"]);
const metricValueKeys = computed(() => isWMediaBucket.value ? ["MediaCount", "RemovedCount"] : ["CountAll", "CountOwned"]);
const latestHeaderKey = computed(() => isWMediaBucket.value ? "LatestBirth" : "LatestRelease");
const latestValueKey = computed(() => isWMediaBucket.value ? "LatestBirthTime" : "LatestReleasingDate");
const tableColumnCount = computed(() => isWMediaBucket.value ? 6 : 5);
const emptyText = computed(() => isWMediaBucket.value
  ? "当前筛选下没有 `w_media_birth_bucket_stat` 数据。若刚建表，请先去聚合回填页执行一次“回填 WMedia 时间聚合”。"
  : "当前筛选下没有 `movie_release_bucket_stat` 数据。若刚建表，请先去聚合回填页执行一次“回填上映日时间聚合”。");

function queryValue(key: string): string {
  return text(query.value, key);
}

function rowValue(row: AnyRecord, ...keys: string[]): unknown {
  return pick(row, ...keys);
}

function rowText(row: AnyRecord, ...keys: string[]): string {
  return text(row, ...keys);
}

function rowKey(row: AnyRecord): string {
  return rowText(row, "ViewHref", "BucketDisplay") || JSON.stringify(row);
}

function rowClass(row: AnyRecord): string {
  if (aggMode.value !== "owned" || queryValue("Level") !== "day") return "";
  const n = Number(rowValue(row, "LatestReleasingDate") || 0);
  if (!n) return "";
  const month = new Date(n * 1000).getMonth() + 1;
  return month % 2 === 0 ? "bucket-month-even" : "bucket-month-odd";
}

function headerLink(key: string): AnyRecord {
  return asRecord(headerLinks.value[key]);
}

function linkLabel(item: AnyRecord): string {
  return text(item, "Label", "label", "Name", "name", "Title", "title");
}

function linkHref(item: AnyRecord): string {
  return normalizeHref(text(item, "Href", "href") || "#");
}

function linkActive(item: AnyRecord): boolean {
  return Boolean(item.Active || item.active);
}

function linkKey(item: AnyRecord): string {
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

function formatUnixDate(input: unknown, withTime = false): string {
  const n = Number(input || 0);
  if (!n) return "";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  if (!withTime) return `${yyyy}-${mm}-${dd}`;
  const hh = String(d.getHours()).padStart(2, "0");
  const mi = String(d.getMinutes()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}`;
}

const HeaderLink = defineComponent({
  name: "ReleaseBucketHeaderLink",
  props: {
    link: { type: Object as PropType<AnyRecord>, required: true }
  },
  setup(props) {
    return () => {
      const active = Boolean(props.link.Active || props.link.active);
      const desc = Boolean(props.link.Desc || props.link.desc);
      return h("a", {
        class: ["bucket-sort-link", { "is-active": active }],
        href: normalizeHref(text(props.link, "Href", "href") || "#")
      }, [
        text(props.link, "Label", "label"),
        active ? h("span", { class: "bucket-sort-indicator" }, desc ? "↓" : "↑") : null
      ]);
    };
  }
});

const PageLinks = defineComponent({
  name: "ReleaseBucketPagination",
  props: {
    links: { type: Array as PropType<AnyRecord[]>, required: true }
  },
  setup(props) {
    return () => h("nav", { "aria-label": "Page navigation" }, [
      h("ul", { class: "pagination pagination-sm mb-0" }, props.links.map((item, index) => {
        if (item.Ellipsis || item.ellipsis) {
          return h("li", { class: "page-item disabled", key: `ellipsis-${index}` }, [
            h("span", { class: "page-link" }, "...")
          ]);
        }
        const disabled = Boolean(item.Disabled || item.disabled);
        const active = Boolean(item.Active || item.active);
        return h("li", { class: ["page-item", { active, disabled }], key: `${item.Label || item.label}-${index}` }, [
          h("a", {
            class: "page-link",
            href: disabled ? "#" : normalizeHref(text(item, "Href", "href") || "#"),
            tabindex: disabled ? "-1" : undefined,
            "aria-disabled": disabled ? "true" : undefined
          }, text(item, "Label", "label"))
        ]);
      }))
    ]);
  }
});
</script>
