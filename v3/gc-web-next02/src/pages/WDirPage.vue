<template>
  <PageShell
    :class="['wdir-page-shell', 'cards-page-shell', 'next02-pink-cards-page', 'next02-pink-cards-page-shell', wdirVariantClass]"
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
                  <a v-if="index < breadcrumbs.length - 1" :href="`/mediadir?id=${crumbId(item)}`">{{ crumbName(item) }}</a>
                  <strong v-else>{{ crumbName(item) }}</strong>
                </template>
              </nav>
              <div class="next02-wdir-overview-metrics" aria-label="目录统计">
                <span v-for="item in statCards" :key="item.label"><small>{{ item.label }}</small><strong><ScDateTime v-if="item.dateTime" :value="item.value" unix /><template v-else>{{ item.value || "-" }}</template></strong></span>
              </div>
            </div>
            <ChildrenGrid :children="children" />
          </section>

          <div class="sticky-toolbar cards-toolbar wdir-movie-toolbar next02-pink-cards-toolbar" aria-label="WDir 影片控制栏">
            <div class="cards-toolbar-sort">
              <LinkStrip :items="sortButtons" />
            </div>
            <div class="toolbar-line cards-toolbar-bottom">
              <div class="owned-line cards-toolbar-filters">
                <LinkStrip :items="ownedButtons" />
                <a class="button-secondary wdir-recursive-button" :class="{ active: recursive }" :href="recursiveHref(true)">递归</a>
                <a class="button-secondary wdir-recursive-button" :class="{ active: !recursive }" :href="recursiveHref(false)">直属</a>
                <button type="button" class="button-secondary cards-filter-button" :class="{ active: filterOpen }" :aria-expanded="filterOpen" @click.stop.prevent="toggleFilter">筛选</button>
              </div>
              <div class="next02-pink-cards-toolbar-status">
                <div class="next02-pink-cards-metrics" aria-label="WDir 页面摘要">
                  <span><strong>{{ total }}</strong> Total</span>
                  <span><strong>{{ movies.length }}</strong> Current Page</span>
                </div>
                <div v-if="pageLinks.length" class="cards-toolbar-pagination next02-pink-cards-pagination">
                  <PaginationBar :page-info="pageInfo" />
                </div>
              </div>
            </div>
          </div>

          <div id="movieCardFilterCollapse" class="collapse next02-pink-cards-filter next02-wdir-filter" :class="{ show: filterOpen }">
            <div class="card">
              <div class="card-body">
              <form class="row g-2 align-items-end" method="get" :action="filterAction">
                <input v-if="filterValue('PageSize')" type="hidden" name="ps" :value="filterValue('PageSize')">
                <input v-if="filterValue('OrderBy')" type="hidden" name="od" :value="filterValue('OrderBy')">
                <div v-for="field in filterFields" :key="field.name" class="col-12 col-md-6 col-xl-2" :class="{ 'smart-pick-date-field': field.placeholder === 'YYYY-MM-DD' }">
                  <label class="form-label mb-1">{{ field.label }}</label>
                  <input class="form-control" :class="{ 'smart-pick-date-input': field.placeholder === 'YYYY-MM-DD' }" :name="field.name" :value="field.value" :placeholder="field.placeholder" :inputmode="field.placeholder === 'YYYY-MM-DD' ? 'numeric' : undefined" :maxlength="field.placeholder === 'YYYY-MM-DD' ? 10 : undefined" :pattern="field.placeholder === 'YYYY-MM-DD' ? '[0-9]{4}-[0-9]{2}-[0-9]{2}' : undefined">
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
            <MovieGrid v-if="movies.length" :cards="movies" />
            <div v-else class="movie-empty">这个目录下暂时没有媒体。</div>
          </section>
        </template>

        <template v-else>
          <section class="next02-wdir-overview next02-wdir-root-overview" aria-label="根目录概览">
            <div class="next02-wdir-overview-toolbar">
              <div class="next02-wdir-root-scope"><span>根目录</span><strong>完整目录树</strong></div>
              <div class="next02-wdir-root-status">
                <span class="next02-wdir-root-count"><strong>{{ total }}</strong> 个目录</span>
              <div v-if="pageLinks.length" class="cards-toolbar-pagination wdir-root-pagination next02-wdir-root-pagination next02-pink-cards-pagination">
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
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LinkStrip from "@/components/LinkStrip.vue";
import LoadingState from "@/components/LoadingState.vue";
import MovieGrid from "@/components/MovieGrid.vue";
import PageShell from "@/components/PageShell.vue";
import PaginationBar from "@/components/PaginationBar.vue";
import ScDateTime from "@/components/ScDateTime.vue";
import type { AnyRecord, PageInfo } from "@/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/utils/format";

const sortLabels: Record<string, string> = {
  ByDetailUpdate: "Detail更新",
  ByReleasing: "上映日期",
  ByViewer: "观看人数",
  ByRankDate: "上榜日期",
  ByDaysInRank: "在榜天数",
  ByHighestRank: "最高排名",
  ByBirthTime: "下载时间",
  ByLastScTime: "最后sc",
  ByScTimes: "sc次数",
  ByComeTimes: "co次数",
  ByCastAgeDesc: "年龄降序",
  ByCastAgeAsc: "年龄升序"
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
const wdirVariantClass = computed(() => isDetail.value ? "is-wdir-detail" : "is-wdir-root");
const children = computed(() => asRecord(pick(payload.value, "Children")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<AnyRecord>(pageInfo.value.Links || pageInfo.value.links));
const total = computed(() => text(payload.value, "Total", "total") || "0");
const detail = computed(() => asRecord(pick(payload.value, "Detail")));
const directory = computed(() => asRecord(pick(detail.value, "directory", "Directory")));
const directoryName = computed(() => text(directory.value, "name", "Name") || title.value);
const pageHeadingTitle = computed(() => isDetail.value ? directoryName.value : title.value);
const pageNote = computed(() => isDetail.value ? "目录路径、媒体统计、子目录与影片。" : "按完整目录树浏览 w_folder / w_media。");
const directoryId = computed(() => text(directory.value, "id", "Id"));
const breadcrumbs = computed(() => asArray<AnyRecord>(pick(detail.value, "breadcrumbs", "Breadcrumbs")));
const stats = computed(() => asRecord(pick(detail.value, "stats", "Stats")));
const recursive = computed(() => Boolean(payload.value.Recursive));
const query = computed(() => text(payload.value, "Query"));
const movies = computed(() => asArray<AnyRecord>(pick(payload.value, "movies", "Movies")));
const sortQuery = computed(() => asRecord(pick(payload.value, "SortQuery", "sortQuery")));
const ownedQuery = computed(() => asRecord(pick(payload.value, "ownedQuery", "OwnedQuery")));
const filter = computed(() => asRecord(pick(payload.value, "MovieCardFilter")));
const filterAction = computed(() => text(filter.value, "Action") || "/mediadir");
const filterClearHref = computed(() => text(filter.value, "ClearHref") || "/mediadir");
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
  { label: "媒体数", value: text(stats.value, "film_count", "FilmCount") || "0", dateTime: false },
  { label: "总大小", value: humanSize(pick(stats.value, "total_size", "TotalSize")), dateTime: false },
  { label: "最近下载时间", value: pick(stats.value, "last_film_birth", "LastFilmBirth"), dateTime: true },
  { label: "最近更新", value: pick(stats.value, "last_updated_on", "LastUpdatedOn"), dateTime: true }
]);
const sortButtons = computed(() => Object.entries(sortQuery.value).filter(([key]) => key !== "ByMediaBirthTime").map(([key, value]) => {
  const item = asRecord(value);
  return { key, label: sortLabels[key] || key, href: normalizeHref(text(item, "Href", "href")), active: Boolean(item.Active || item.active), disabled: Boolean(item.Active || item.active) };
}).filter((item) => item.href));
const ownedButtons = computed(() => [
  ["MediaAll", "All"],
  ["MediaOwned", "Owned"],
  ["MediaNotOwned", "NotOwned"]
].map(([key, label]) => {
  const item = asRecord(ownedQuery.value[key]);
  return { key, label, href: normalizeHref(text(item, "Href", "href")), active: Boolean(item.Active || item.active), disabled: Boolean(item.Active || item.active) };
}));

function filterValue(key: string): string {
  return text(filter.value, key);
}

function toggleFilter(): void {
  filterOpen.value = !filterOpen.value;
}

function recursiveHref(next: boolean): string {
  return `/mediadir?id=${encodeURIComponent(directoryId.value)}&recursive=${next ? "1" : "0"}${query.value || ""}`;
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

const ChildrenGrid = defineComponent({
  name: "WDirChildrenGrid",
  props: {
    children: { type: Object as PropType<AnyRecord>, required: true }
  },
  setup(props) {
    return () => {
      const items = asArray<AnyRecord>(props.children.Items || props.children.items);
      return h("section", { class: "next02-wdir-children", "aria-label": "Subdirectories" }, [
        h("div", { class: "next02-wdir-children-head" }, [
          h("h2", text(props.children, "Title") || "子目录"),
          text(props.children, "MoreURL") ? h("a", { href: text(props.children, "MoreURL") }, "更多") : null
        ]),
        items.length
          ? h("div", { class: "next02-wdir-directory-grid" }, items.map((item) => {
            const summary = asRecord(pick(item, "Summary"));
            const stat = asRecord(asArray<AnyRecord>(pick(item, "Stats"))[0]);
            const id = text(summary, "id", "Id");
            const name = text(summary, "name", "Name");
            return h("a", { href: `/mediadir?id=${encodeURIComponent(id)}`, class: "next02-wdir-directory-item", key: id || name }, [
              h("span", { class: "next02-wdir-folder-mark", "aria-hidden": "true" }),
              h("strong", { class: "next02-wdir-directory-name", title: name }, name),
              h("span", { class: "next02-wdir-directory-metric" }, [h("small", "媒体"), h("b", text(stat, "film_count", "FilmCount") || "0")]),
              h("span", { class: "next02-wdir-directory-metric" }, [h("small", "容量"), h("b", humanSize(pick(stat, "total_size", "TotalSize")))]),
              h("time", { class: "next02-wdir-directory-updated" }, ["更新 ", h(ScDateTime, { value: pick(summary, "updated_on", "UpdatedOn"), unix: true })])
            ]);
          }))
          : h("div", { class: "movie-empty next02-wdir-empty" }, "暂无子目录。")
      ]);
    };
  }
});

</script>
