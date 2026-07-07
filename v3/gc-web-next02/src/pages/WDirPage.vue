<template>
  <PageShell class="movie-agg-page-shell wdir-page-shell" :title="title" note="按完整目录树浏览 `w_folder / w_media`。" eyebrow="">
    <div id="page-root" class="container-fluid movie-agg-page-root wdir-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <template v-if="isDetail">
          <section class="section-card wdir-detail-head">
            <nav aria-label="breadcrumb">
              <ol class="breadcrumb mb-1">
                <li class="breadcrumb-item"><a href="/wdir">根目录</a></li>
                <li v-for="(item, index) in breadcrumbs" :key="crumbKey(item, index)" class="breadcrumb-item" :class="{ active: index === breadcrumbs.length - 1 }">
                  <a v-if="index < breadcrumbs.length - 1" :href="`/wdir?id=${crumbId(item)}`">{{ crumbName(item) }}</a>
                  <span v-else>{{ crumbName(item) }}</span>
                </li>
              </ol>
            </nav>
            <div class="d-flex flex-column flex-md-row justify-content-between align-items-start align-items-md-center gap-2">
              <h3 class="mb-0">{{ directoryName }}</h3>
              <div class="btn-group">
                <a class="btn btn-sm btn-outline-secondary" :class="{ active: recursive }" :href="recursiveHref(true)">递归</a>
                <a class="btn btn-sm btn-outline-secondary" :class="{ active: !recursive }" :href="recursiveHref(false)">直属</a>
              </div>
            </div>
            <div class="row g-2 mt-2">
              <div v-for="item in statCards" :key="item.label" class="col-6 col-md-3">
                <div class="card shadow-sm h-100">
                  <div class="card-body">
                    <div class="text-muted small">{{ item.label }}</div>
                    <div class="h6 mb-0">{{ item.value }}</div>
                  </div>
                </div>
              </div>
            </div>
            <ChildrenGrid :children="children" />
          </section>

          <section class="section-card wdir-toolbar">
            <div class="d-flex flex-wrap justify-content-start align-items-center mb-2">
              <a v-for="item in sortButtons" :key="item.key" :href="item.href" class="btn btn-outline-primary" :class="{ 'active disabled': item.active }">{{ item.label }}</a>
            </div>
            <div class="d-flex flex-column flex-md-row justify-content-between align-items-md-center gap-2">
              <div class="d-flex flex-wrap align-items-center gap-2">
                <a v-for="item in ownedButtons" :key="item.key" :href="item.href" class="btn btn-sm" :class="[item.className, { 'active disabled': item.active }]">{{ item.label }}</a>
                <button type="button" class="btn btn-sm btn-outline-secondary" @click="filterOpen = !filterOpen">筛选</button>
              </div>
              <PageLinks v-if="pageLinks.length" :links="pageLinks" />
            </div>
          </section>

          <section v-if="filterOpen" class="card mb-3">
            <div class="card-body">
              <form class="row g-2 align-items-end" method="get" :action="filterAction">
                <input v-if="filterValue('PageSize')" type="hidden" name="ps" :value="filterValue('PageSize')">
                <input v-if="filterValue('OrderBy')" type="hidden" name="od" :value="filterValue('OrderBy')">
                <div v-for="field in filterFields" :key="field.name" class="col-12 col-md-6 col-xl-2">
                  <label class="form-label mb-1">{{ field.label }}</label>
                  <input class="form-control" :name="field.name" :value="field.value" :placeholder="field.placeholder">
                </div>
                <div class="col-12 d-flex gap-2">
                  <button class="btn btn-primary" type="submit">筛选</button>
                  <a class="btn btn-outline-secondary" :href="filterClearHref">清空</a>
                </div>
              </form>
            </div>
          </section>

          <section class="section-card">
            <MovieGrid v-if="movies.length" :cards="movies" />
            <div v-else class="alert alert-light border mb-0">这个目录下暂时没有媒体。</div>
          </section>
        </template>

        <template v-else>
          <section class="section-card">
            <div class="d-flex flex-column flex-md-row justify-content-between align-items-start align-items-md-center gap-3 mb-3">
              <div>
                <h3 class="mb-1">媒体目录</h3>
                <p class="text-muted mb-0">按完整目录树浏览 `w_folder / w_media`。</p>
              </div>
              <div class="d-flex flex-wrap align-items-center gap-2">
                <span class="btn btn-sm btn-outline-secondary disabled">目录数 {{ total }}</span>
                <PageLinks v-if="pageLinks.length" :links="pageLinks" />
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
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import MovieGrid from "@/components/MovieGrid.vue";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/utils/format";

const sortLabels: Record<string, string> = {
  ByDetailUpdate: "Detail更新",
  ByReleasing: "上映日期",
  ByViewer: "观看人数",
  ByRankDate: "上榜日期",
  ByDaysInRank: "在榜天数",
  ByHighestRank: "最高排名",
  ByBirthTime: "下载时间",
  ByMediaBirthTime: "Media下载时间",
  ByLastScTime: "最后sc",
  ByScTimes: "sc次数",
  ByComeTimes: "co次数"
};

const route = useRoute();
const payload = ref<AnyRecord>({});
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
const children = computed(() => asRecord(pick(payload.value, "Children")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")));
const pageLinks = computed(() => asArray<AnyRecord>(pageInfo.value.Links || pageInfo.value.links));
const total = computed(() => text(payload.value, "Total", "total") || "0");
const detail = computed(() => asRecord(pick(payload.value, "Detail")));
const directory = computed(() => asRecord(pick(detail.value, "directory", "Directory")));
const directoryName = computed(() => text(directory.value, "name", "Name") || title.value);
const directoryId = computed(() => text(directory.value, "id", "Id"));
const breadcrumbs = computed(() => asArray<AnyRecord>(pick(detail.value, "breadcrumbs", "Breadcrumbs")));
const stats = computed(() => asRecord(pick(detail.value, "stats", "Stats")));
const recursive = computed(() => Boolean(payload.value.Recursive));
const query = computed(() => text(payload.value, "Query"));
const movies = computed(() => asArray<AnyRecord>(pick(payload.value, "movies", "Movies")));
const sortQuery = computed(() => asRecord(pick(payload.value, "SortQuery", "sortQuery")));
const ownedQuery = computed(() => asRecord(pick(payload.value, "ownedQuery", "OwnedQuery")));
const filter = computed(() => asRecord(pick(payload.value, "MovieCardFilter")));
const filterAction = computed(() => text(filter.value, "Action") || "/wdir");
const filterClearHref = computed(() => text(filter.value, "ClearHref") || "/wdir");
const filterFields = computed(() => [
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
].map(([name, label, key, placeholder]) => ({ name, label, value: filterValue(key), placeholder })));
const statCards = computed(() => [
  { label: "媒体数", value: text(stats.value, "film_count", "FilmCount") || "0" },
  { label: "总大小", value: humanSize(pick(stats.value, "total_size", "TotalSize")) },
  { label: "最近下载时间", value: formatUnix(pick(stats.value, "last_film_birth", "LastFilmBirth"), true) },
  { label: "最近更新", value: formatUnix(pick(stats.value, "last_updated_on", "LastUpdatedOn"), true) }
]);
const sortButtons = computed(() => Object.entries(sortQuery.value).map(([key, value]) => {
  const item = asRecord(value);
  return { key, label: sortLabels[key] || key, href: normalizeHref(text(item, "Href", "href")), active: Boolean(item.Active || item.active) };
}).filter((item) => item.href));
const ownedButtons = computed(() => [
  ["MediaAll", "MAll", "btn-primary"],
  ["MediaOwned", "MOwned", "btn-success"],
  ["MediaNotOwned", "MNotOwned", "btn-secondary"]
].map(([key, label, className]) => {
  const item = asRecord(ownedQuery.value[key]);
  return { key, label, className, href: normalizeHref(text(item, "Href", "href")), active: Boolean(item.Active || item.active) };
}));

function filterValue(key: string): string {
  return text(filter.value, key);
}

function recursiveHref(next: boolean): string {
  return `/wdir?id=${encodeURIComponent(directoryId.value)}&recursive=${next ? "1" : "0"}${query.value || ""}`;
}

function crumbName(item: AnyRecord): string {
  return text(item, "name", "Name");
}

function crumbId(item: AnyRecord): string {
  return encodeURIComponent(text(item, "id", "Id"));
}

function crumbKey(item: AnyRecord, index: number): string {
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

function formatUnix(input: unknown, withTime = false): string {
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

const ChildrenGrid = defineComponent({
  name: "WDirChildrenGrid",
  props: {
    children: { type: Object as PropType<AnyRecord>, required: true }
  },
  setup(props) {
    return () => {
      const items = asArray<AnyRecord>(props.children.Items || props.children.items);
      return h("section", { class: "mt-3", "aria-label": "Subdirectories" }, [
        h("div", { class: "d-flex align-items-center justify-content-between mb-2" }, [
          h("h5", { class: "mb-0" }, text(props.children, "Title") || "子目录"),
          text(props.children, "MoreURL") ? h("a", { class: "small", href: text(props.children, "MoreURL") }, "更多") : null
        ]),
        items.length
          ? h("div", { class: "row row-cols-2 row-cols-sm-3 row-cols-lg-5 row-cols-xxl-6 g-3" }, items.map((item) => {
            const summary = asRecord(pick(item, "Summary"));
            const stat = asRecord(asArray<AnyRecord>(pick(item, "Stats"))[0]);
            const id = text(summary, "id", "Id");
            const name = text(summary, "name", "Name");
            return h("div", { class: "col", key: id || name }, [
              h("a", { href: `/wdir?id=${encodeURIComponent(id)}`, class: "text-decoration-none" }, [
                h("div", { class: "card h-100 shadow-sm" }, [
                  h("div", { class: "card-body py-2" }, [
                    h("div", { class: "d-flex align-items-start" }, [
                      h("div", { class: "flex-grow-1" }, [
                        h("div", { class: "fw-semibold text-truncate", title: name }, name),
                        h("div", { class: "text-muted small" }, `${text(stat, "film_count", "FilmCount") || "0"} 个媒体 · ${humanSize(pick(stat, "total_size", "TotalSize"))}`)
                      ])
                    ])
                  ]),
                  h("div", { class: "card-footer py-1 text-muted small" }, `更新：${formatUnix(pick(summary, "updated_on", "UpdatedOn"), true)}`)
                ])
              ])
            ]);
          }))
          : h("div", { class: "alert alert-light border mb-0" }, "暂无子目录。")
      ]);
    };
  }
});

const PageLinks = defineComponent({
  name: "WDirPagination",
  props: {
    links: { type: Array as PropType<AnyRecord[]>, required: true }
  },
  setup(props) {
    return () => h("nav", { "aria-label": "Page navigation" }, [
      h("ul", { class: "pagination pagination-sm mb-0" }, props.links.map((item, index) => {
        if (item.Ellipsis || item.ellipsis) {
          return h("li", { class: "page-item disabled", key: `ellipsis-${index}` }, [h("span", { class: "page-link" }, "...")]);
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
