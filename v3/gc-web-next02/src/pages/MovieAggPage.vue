<template>
  <PageShell
    :class="['torrent-albums-page-shell', 'movie-agg-page-shell', 'cards-page-shell', 'next02-pink-cards-page', 'next02-pink-cards-page-shell', movieAggVariantClass]"
    :title="title"
    :note="note"
    eyebrow="MovieAgg"
  >
    <div id="page-root" :class="['container-fluid torrent-albums-page movie-agg-page-root next02-movieagg-rewrite', movieAggVariantClass]">
      <LoadingState v-if="loading" />
      <ErrorPanel v-else-if="error" :message="error" />
      <template v-else>
        <template v-if="isReleaseView">
          <section class="section-card movie-agg-release-section next02-rewrite-card next02-movieagg-section next02-movieagg-timeline">
            <div class="next02-movieagg-timeline-toolbar">
              <div class="next02-movieagg-timeline-heading">
                <h2>时间聚合</h2>
                <BreadcrumbLine :items="breadcrumbs" />
              </div>
              <div class="next02-movieagg-timeline-controls">
                <div v-if="hasYearTabs" class="next02-movieagg-segmented" role="tablist" aria-label="时间粒度">
                  <button v-if="bucketsQAll.length" type="button" role="tab" :aria-selected="releaseBucketTabActive('quarters')" :class="{ active: releaseBucketTabActive('quarters') }" @click="selectReleaseBucketTab('quarters')">季度</button>
                  <button v-if="bucketsMAll.length" type="button" role="tab" :aria-selected="releaseBucketTabActive('months')" :class="{ active: releaseBucketTabActive('months') }" @click="selectReleaseBucketTab('months')">月份</button>
                </div>
                <div class="next02-movieagg-mode-switches">
                  <a v-for="item in modeSwitches" :key="linkKey(item)" :href="linkHref(item)" :class="{ active: linkActive(item) }">{{ linkLabel(item) }}</a>
                </div>
              </div>
            </div>
            <template v-if="hasYearTabs">
              <BucketRows v-if="bucketsQAll.length" v-show="releaseBucketTabActive('quarters')" :items="bucketsQAll" grid-class="bucket-grid-quarter" />
              <BucketRows v-if="bucketsMAll.length" v-show="releaseBucketTabActive('months')" :items="bucketsMAll" grid-class="bucket-grid-month" />
            </template>
            <BucketRows v-else-if="bucketsAll.length" :items="bucketsAll" grid-class="bucket-grid-default" />
            <div v-else class="movie-empty next02-movieagg-empty">暂无聚合数据。</div>
          </section>

          <section v-if="topGroups.length" class="section-card movie-agg-top-section next02-rewrite-card next02-movieagg-section next02-movieagg-top">
            <div class="next02-movieagg-top-toolbar">
              <div class="next02-movieagg-section-title">
                <h2>Top 聚合</h2>
                <span v-if="aggModeLabel" class="next02-movieagg-mode-badge">{{ aggModeLabel }}</span>
              </div>
              <div class="next02-movieagg-top-tabs" role="tablist" aria-label="Top 聚合类型">
                <button v-for="(group, index) in topGroups" :key="group.type" type="button" role="tab" :aria-selected="topGroupActive(group, index)" :class="{ active: topGroupActive(group, index) }" @click="selectTopGroup(group)">{{ group.label }}</button>
              </div>
              <button type="button" class="button-secondary next02-movieagg-expand" :aria-expanded="topExpanded" @click="topExpanded = !topExpanded">{{ topExpanded ? "收起" : "展开" }}</button>
            </div>
            <div class="collapse" :class="{ show: topExpanded }" id="topAggCollapse">
              <div v-for="(group, index) in topGroups" v-show="topGroupActive(group, index)" :key="`${group.type}-pane`" class="next02-movieagg-top-pane" role="tabpanel">
                <div class="top-stat-grid">
                  <a v-for="(item, itemIndex) in group.items" :key="topKey(group.type, item, itemIndex)" :href="topHref(group.type, item)" class="top-stat-card" :title="topName(item)" target="_blank" rel="noopener noreferrer">
                    <span class="top-stat-rank">#{{ String(itemIndex + 1).padStart(2, "0") }}</span>
                    <strong class="top-stat-name">{{ topName(item) }}</strong>
                    <span class="top-stat-metric"><small>总</small><b>{{ topCountAll(item) }}</b></span>
                    <span class="top-stat-metric"><small>拥有</small><b>{{ topCountOwned(item) }}</b></span>
                  </a>
                </div>
              </div>
            </div>
          </section>

          <section v-if="moviesProvided" class="movie-agg-cards-section movie-agg-movies-section">
            <MovieToolbar :sort-query="sortQuery" :owned-query="ownedQuery" :total-label="totalLabel" :current-page-count="movies.length" :page-links="pageLinks" />
            <MovieGrid v-if="movies.length" :cards="movies" />
            <div v-else class="movie-empty next02-movieagg-empty">暂无影片。</div>
          </section>
        </template>

        <template v-else-if="isWMediaBirth">
          <div class="w-media-agg-page">
            <section class="section-card next02-rewrite-card next02-movieagg-section next02-movieagg-timeline">
              <div class="next02-movieagg-timeline-toolbar">
                <div class="next02-movieagg-timeline-heading">
                  <h2>下载时间聚合</h2>
                  <BreadcrumbLine :items="breadcrumbs" />
                </div>
                <div v-if="hasWMediaYearTabs" class="next02-movieagg-timeline-controls">
                  <div class="next02-movieagg-segmented" role="tablist" aria-label="时间粒度">
                    <button v-if="bucketsQ.length" type="button" role="tab" :aria-selected="wMediaBucketTabActive('quarters')" :class="{ active: wMediaBucketTabActive('quarters') }" @click="selectWMediaBucketTab('quarters')">季度</button>
                    <button v-if="bucketsM.length" type="button" role="tab" :aria-selected="wMediaBucketTabActive('months')" :class="{ active: wMediaBucketTabActive('months') }" @click="selectWMediaBucketTab('months')">月份</button>
                  </div>
                </div>
              </div>
              <template v-if="hasWMediaYearTabs">
                <BucketRows v-if="bucketsQ.length" v-show="wMediaBucketTabActive('quarters')" :items="bucketsQ" grid-class="bucket-grid-quarter" />
                <BucketRows v-if="bucketsM.length" v-show="wMediaBucketTabActive('months')" :items="bucketsM" grid-class="bucket-grid-month" />
              </template>
              <template v-else-if="hasWMediaBuckets">
                <BucketRows v-if="bucketsY.length" title="年份" :items="bucketsY" grid-class="bucket-grid-default" />
                <BucketRows v-if="bucketsQ.length" title="季度" :items="bucketsQ" grid-class="bucket-grid-quarter" />
                <BucketRows v-if="bucketsM.length" title="月份" :items="bucketsM" grid-class="bucket-grid-month" />
                <BucketRows v-if="bucketsD.length" title="日期" :items="bucketsD" grid-class="bucket-grid-month" />
              </template>
              <div v-else class="movie-empty next02-movieagg-empty">
                当前范围暂无可展示的时间桶。<a href="/triggers/agg" class="alert-link" target="_blank" rel="noopener noreferrer">如为新建统计表，请先到聚合回填页执行 WMedia 时间聚合回填。</a>
              </div>
            </section>

            <section v-if="topGroups.length" class="section-card movie-agg-top-section next02-rewrite-card next02-movieagg-section next02-movieagg-top">
              <div class="next02-movieagg-top-toolbar">
                <div class="next02-movieagg-section-title">
                  <h2>Top 聚合</h2>
                </div>
                <div class="next02-movieagg-top-tabs" role="tablist" aria-label="Top 聚合类型">
                  <button v-for="(group, index) in topGroups" :key="group.type" type="button" role="tab" :aria-selected="topGroupActive(group, index)" :class="{ active: topGroupActive(group, index) }" @click="selectTopGroup(group)">{{ group.label }}</button>
                </div>
                <button type="button" class="button-secondary next02-movieagg-expand" :aria-expanded="topExpanded" @click="topExpanded = !topExpanded">{{ topExpanded ? "收起" : "展开" }}</button>
              </div>
              <div class="collapse" :class="{ show: topExpanded }" id="topAggCollapse">
                <div v-for="(group, index) in topGroups" v-show="topGroupActive(group, index)" :key="`${group.type}-pane`" class="next02-movieagg-top-pane" role="tabpanel">
                  <div class="top-stat-grid">
                    <a v-for="(item, itemIndex) in group.items" :key="topKey(group.type, item, itemIndex)" :href="topHref(group.type, item)" class="top-stat-card" :title="topName(item)" target="_blank" rel="noopener noreferrer">
                      <span class="top-stat-rank">#{{ String(itemIndex + 1).padStart(2, "0") }}</span>
                      <strong class="top-stat-name">{{ topName(item) }}</strong>
                      <span class="top-stat-metric"><small>出现</small><b>{{ topCountAll(item) }}</b></span>
                      <span class="top-stat-metric"><small>容量</small><b>{{ topSize(item) }}</b></span>
                    </a>
                  </div>
                </div>
              </div>
            </section>

            <section class="movie-agg-cards-section">
              <MovieToolbar :sort-query="sortQuery" :owned-query="ownedQuery" :total-label="totalLabel" :current-page-count="movies.length" :page-links="pageLinks" />
              <MovieGrid v-if="movies.length" :cards="movies" />
              <div v-else class="movie-empty next02-movieagg-empty">当前范围暂无媒体卡片。</div>
            </section>
          </div>
        </template>

        <template v-else>
          <section class="section-card next02-rewrite-card">
            <BreadcrumbLine :items="breadcrumbs" />
            <h4 class="section-title">{{ title }}</h4>
            <BucketRows v-for="section in fallbackBucketSections" :key="section.key" :title="section.title" :items="section.items" grid-class="bucket-grid bucket-grid-default" />
          </section>
          <section v-if="topGroups.length" class="section-card movie-agg-top-section next02-rewrite-card">
            <div class="d-flex align-items-center justify-content-start mb-2">
              <h4 class="section-title mb-0 me-3">Top 聚合</h4>
              <button type="button" class="btn btn-outline-info waves-effect waves-light btn-sm" @click="topExpanded = !topExpanded">{{ topExpanded ? "收起" : "展开" }}</button>
            </div>
            <div class="collapse" :class="{ show: topExpanded }" id="topAggCollapse">
              <ul class="nav nav-tabs nav-tabs-custom nav-justified" role="tablist">
                <li v-for="(group, index) in topGroups" :key="group.type" class="nav-item">
                  <a class="nav-link" :class="{ active: topGroupActive(group, index) }" :href="`#tab-top-${group.type}`" role="tab" @click.prevent="selectTopGroup(group)">{{ group.label }}</a>
                </li>
              </ul>
              <div class="tab-content pt-3">
                <div v-for="(group, index) in topGroups" :key="`${group.type}-pane`" class="tab-pane p-2" :class="{ active: topGroupActive(group, index) }" :id="`tab-top-${group.type}`" role="tabpanel">
                  <div class="top-stat-grid">
                    <a v-for="(item, itemIndex) in group.items" :key="topKey(group.type, item, itemIndex)" :href="topHref(group.type, item)" class="top-stat-card text-decoration-none" :title="topName(item)">
                      <div class="card shadow-sm">
                        <div class="card-body p-2 text-center">
                          <div class="mb-1">
                            <span class="badge bg-secondary text-white fw-normal badge-slim">#{{ itemIndex + 1 }}</span>
                            <span class="text-primary fw-semibold text-truncate">{{ topName(item) }}</span>
                          </div>
                          <div v-if="isWMediaBirth" class="d-flex flex-column align-items-center gap-1">
                            <span class="badge bg-light text-muted border badge-slim">出现：<span class="text-dark fw-semibold">{{ topCountAll(item) }}</span></span>
                            <span class="badge bg-light text-muted border badge-slim">体积：<span class="text-dark fw-semibold">{{ topSize(item) }}</span></span>
                          </div>
                          <div v-else class="d-flex flex-column align-items-center gap-1">
                            <span class="badge bg-light text-muted border badge-slim">总：<span class="text-dark fw-semibold">{{ topCountAll(item) }}</span></span>
                            <span class="badge bg-light text-muted border badge-slim">拥有：<span class="text-dark fw-semibold">{{ topCountOwned(item) }}</span></span>
                          </div>
                        </div>
                      </div>
                    </a>
                  </div>
                </div>
              </div>
            </div>
          </section>
          <section v-if="movies.length" class="movie-agg-cards-section">
            <MovieToolbar :sort-query="sortQuery" :owned-query="ownedQuery" :total-label="totalLabel" :current-page-count="movies.length" :page-links="pageLinks" />
            <MovieGrid :cards="movies" />
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
import LoadingState from "@/components/LoadingState.vue";
import MovieGrid from "@/components/MovieGrid.vue";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/utils/format";

type TopGroup = {
  label: string;
  type: string;
  items: AnyRecord[];
};

const sortLabels: [string, string][] = [
  ["ByDetailUpdate", "Detail更新"],
  ["ByReleasing", "上映日期"],
  ["ByViewer", "观看人数"],
  ["ByRankDate", "上榜日期"],
  ["ByDaysInRank", "在榜天数"],
  ["ByHighestRank", "最高排名"],
  ["ByBirthTime", "下载时间"],
  ["ByLastScTime", "最后sc"],
  ["ByScTimes", "sc次数"],
  ["ByComeTimes", "co次数"]
];

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const topExpanded = ref(false);
const activeReleaseBucketTab = ref<"quarters" | "months">("quarters");
const activeWMediaBucketTab = ref<"quarters" | "months">("quarters");
const activeTopType = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
    topExpanded.value = false;
    activeReleaseBucketTab.value = "quarters";
    activeWMediaBucketTab.value = "quarters";
    activeTopType.value = "";
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title") || String(route.meta.title || "MovieAgg"));
const aggModeLabel = computed(() => text(payload.value, "AggModeLabel"));
const note = computed(() => [aggModeLabel.value, text(payload.value, "RangeStart"), text(payload.value, "RangeEnd")].filter(Boolean).join(" · ") || "按时间桶、Top 维度和媒体状态查看聚合结果。");
const isReleaseView = computed(() => text(payload.value, "ViewType") === "movie-agg-release" || route.path === "/movie-agg-all/release");
const isWMediaBirth = computed(() => text(payload.value, "ViewType") === "w-media-agg-birth" || route.path === "/w-media-agg/birth");
const movieAggVariantClass = computed(() => isReleaseView.value ? "is-release-view" : (isWMediaBirth.value ? "is-wmedia-birth-view" : "is-generic-view"));
const breadcrumbs = computed(() => asArray<AnyRecord>(pick(payload.value, "Breadcrumbs")));
const modeSwitches = computed(() => asArray<AnyRecord>(pick(payload.value, "ModeSwitches")));
const bucketsAll = computed(() => asArray<AnyRecord>(pick(payload.value, "BucketsAll")));
const bucketsQAll = computed(() => asArray<AnyRecord>(pick(payload.value, "BucketsQAll")));
const bucketsMAll = computed(() => asArray<AnyRecord>(pick(payload.value, "BucketsMAll")));
const bucketsY = computed(() => asArray<AnyRecord>(pick(payload.value, "BucketsY")));
const bucketsQ = computed(() => asArray<AnyRecord>(pick(payload.value, "BucketsQ")));
const bucketsM = computed(() => asArray<AnyRecord>(pick(payload.value, "BucketsM")));
const bucketsD = computed(() => asArray<AnyRecord>(pick(payload.value, "BucketsD")));
const hasYearTabs = computed(() => bucketsQAll.value.length > 0 || bucketsMAll.value.length > 0);
const hasWMediaYearTabs = computed(() => text(payload.value, "Level") === "year" && (bucketsQ.value.length > 0 || bucketsM.value.length > 0));
const hasWMediaBuckets = computed(() => bucketsY.value.length > 0 || bucketsQ.value.length > 0 || bucketsM.value.length > 0 || bucketsD.value.length > 0);
const movies = computed(() => asArray<AnyRecord>(pick(payload.value, "movies", "Movies")));
const moviesProvided = computed(() => Object.prototype.hasOwnProperty.call(payload.value, "Movies") || Object.prototype.hasOwnProperty.call(payload.value, "movies"));
const sortQuery = computed(() => asRecord(pick(payload.value, "SortQuery", "sortQuery")));
const ownedQuery = computed(() => asRecord(pick(payload.value, "ownedQuery", "OwnedQuery")));
const totalLabel = computed(() => text(payload.value, "TotalLabel", "total", "Total"));
const pageLinks = computed(() => asArray<AnyRecord>(asRecord(pick(payload.value, "PageInfo", "pageInfo")).Links || asRecord(pick(payload.value, "PageInfo", "pageInfo")).links));
const topGroups = computed<TopGroup[]>(() => [
  { label: "演员", type: "cast", items: asArray<AnyRecord>(pick(payload.value, "TopCastsAll", "TopCasts")) },
  { label: "导演", type: "director", items: asArray<AnyRecord>(pick(payload.value, "TopDirectorsAll", "TopDirectors")) },
  { label: "Label", type: "label", items: asArray<AnyRecord>(pick(payload.value, "TopLabelsAll", "TopLabels")) },
  { label: "Prefix", type: "prefix", items: asArray<AnyRecord>(pick(payload.value, "TopPrefixesAll", "TopPrefixes")) }
].filter((group) => group.items.length > 0));

watch(topGroups, (groups) => {
  if (!groups.length) {
    activeTopType.value = "";
    return;
  }
  if (!groups.some((group) => group.type === activeTopType.value)) {
    activeTopType.value = groups[0].type;
  }
});
const fallbackBucketSections = computed(() => {
  const keys: [string, string][] = [
    ["BucketsY", "年份"],
    ["BucketsQ", "季度"],
    ["BucketsM", "月份"],
    ["BucketsD", "日期"],
    ["BucketsAll", "时间桶"]
  ];
  return keys.map(([key, title]) => ({ key, title, items: asArray<AnyRecord>(payload.value[key]) }))
    .filter((section) => section.items.length > 0);
});

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

function selectReleaseBucketTab(tab: "quarters" | "months"): void {
  activeReleaseBucketTab.value = tab;
}

function releaseBucketTabActive(tab: "quarters" | "months"): boolean {
  if (tab === "quarters") {
    return bucketsQAll.value.length > 0 && activeReleaseBucketTab.value === "quarters";
  }
  return bucketsMAll.value.length > 0 && (bucketsQAll.value.length === 0 || activeReleaseBucketTab.value === "months");
}

function selectWMediaBucketTab(tab: "quarters" | "months"): void {
  activeWMediaBucketTab.value = tab;
}

function wMediaBucketTabActive(tab: "quarters" | "months"): boolean {
  if (tab === "quarters") {
    return bucketsQ.value.length > 0 && activeWMediaBucketTab.value === "quarters";
  }
  return bucketsM.value.length > 0 && (bucketsQ.value.length === 0 || activeWMediaBucketTab.value === "months");
}

function selectTopGroup(group: TopGroup): void {
  activeTopType.value = group.type;
}

function topGroupActive(group: TopGroup, index: number): boolean {
  return activeTopType.value ? activeTopType.value === group.type : index === 0;
}

function bucketLabel(item: AnyRecord): string {
  return text(item, "Label", "label", "BucketDisplay", "ScopeKey", "Name", "name");
}

function bucketHref(item: AnyRecord): string {
  return normalizeHref(text(item, "Href", "href", "ViewHref") || "#");
}

function bucketCountAll(item: AnyRecord): string {
  return text(item, "CountAll", "Count", "MovieCount") || "0";
}

function bucketCountOwned(item: AnyRecord): string {
  return text(item, "CountOwned", "OwnedCount") || "0";
}

function bucketSizeGb(item: AnyRecord): string {
  return Number(pick(item, "SizeAllGB", "SizeGB") || 0).toFixed(1);
}

function topName(item: AnyRecord): string {
  return text(item, "Name", "name", "Label", "label");
}

function topCountAll(item: AnyRecord): string {
  return text(item, "CountAll", "Count") || "0";
}

function topCountOwned(item: AnyRecord): string {
  return text(item, "CountOwned", "OwnedCount") || "0";
}

function humanSize(input: unknown): string {
  const value = Number(input || 0);
  if (value >= 1024 ** 4) return `${(value / 1024 ** 4).toFixed(1)} TB`;
  if (value >= 1024 ** 3) return `${(value / 1024 ** 3).toFixed(1)} GB`;
  if (value >= 1024 ** 2) return `${(value / 1024 ** 2).toFixed(1)} MB`;
  if (value >= 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${value} B`;
}

function topSize(item: AnyRecord): string {
  return humanSize(pick(item, "SizeBytes"));
}

function topKey(type: string, item: AnyRecord, index: number): string {
  return `${type}-${topName(item)}-${index}`;
}

function topHref(type: string, item: AnyRecord): string {
  const params = new URLSearchParams();
  if (type === "cast" && Number(pick(item, "PersonId", "AggID") || 0) > 0) {
    params.set("pid", String(pick(item, "PersonId", "AggID")));
  } else if (type === "cast") {
    params.set("cn", topName(item));
  } else if (type === "director") {
    params.set("dn", topName(item));
  } else if (type === "label") {
    params.set("ln", topName(item));
  } else {
    params.set("pn", topName(item));
  }
  params.set("mowned", isWMediaBirth.value ? "3" : (text(payload.value, "AggMode") === "owned" ? "3" : "1"));
  const start = text(payload.value, "RangeStart");
  const end = text(payload.value, "RangeEnd");
  if (start) params.set(isWMediaBirth.value ? "mbs" : "rs", start);
  if (end) params.set(isWMediaBirth.value ? "mbe" : "re", end);
  return `/cards?${params.toString()}`;
}

const BreadcrumbLine = defineComponent({
  name: "MovieAggBreadcrumbLine",
  props: {
    items: { type: Array as PropType<AnyRecord[]>, required: true }
  },
  setup(props) {
    return () => h("div", { class: "bucket-breadcrumbs" }, props.items.map((item, index) => [
      index > 0 ? h("span", { class: "bucket-breadcrumb-separator", "aria-hidden": "true" }, "›") : null,
      h("a", { href: linkHref(item) }, linkLabel(item))
    ]).flat());
  }
});

const BucketRows = defineComponent({
  name: "MovieAggBucketRows",
  props: {
    title: { type: String, default: "" },
    items: { type: Array as PropType<AnyRecord[]>, required: true },
    gridClass: { type: String, required: true }
  },
  setup(props, { attrs }) {
    return () => h("section", { ...attrs, class: ["bucket-section", attrs.class] }, [
      props.title ? h("div", { class: "bucket-section-title" }, props.title) : null,
      h("div", {
        class: ["next02-movieagg-bucket-strip", props.gridClass]
      }, props.items.map((item) => {
        const mediaBucket = pick(item, "Count") !== undefined;
        return h("a", { href: bucketHref(item), class: "next02-movieagg-bucket-item", key: bucketHref(item) }, [
          h("strong", { class: "next02-movieagg-bucket-label" }, bucketLabel(item)),
          h("span", { class: "next02-movieagg-bucket-metric" }, [
            h("small", mediaBucket ? "影片" : "总数"),
            h("b", bucketCountAll(item))
          ]),
          mediaBucket ? null : h("span", { class: "next02-movieagg-bucket-metric" }, [
            h("small", "拥有"),
            h("b", bucketCountOwned(item))
          ]),
          h("span", { class: "next02-movieagg-bucket-metric next02-movieagg-bucket-size" }, [
            h("small", "容量"),
            h("b", `${bucketSizeGb(item)} GB`)
          ])
        ]);
      }))
    ]);
  }
});

const MovieToolbar = defineComponent({
  name: "MovieAggToolbar",
  props: {
    sortQuery: { type: Object as PropType<AnyRecord>, required: true },
    ownedQuery: { type: Object as PropType<AnyRecord>, required: true },
    totalLabel: { type: String, required: true },
    currentPageCount: { type: Number, required: true },
    pageLinks: { type: Array as PropType<AnyRecord[]>, required: true }
  },
  setup(props) {
    const sortItem = (key: string, label: string) => {
      const item = asRecord(props.sortQuery[key]);
      return {
        label,
        href: linkHref(item),
        active: linkActive(item),
        disabled: false
      };
    };
    const ownedItem = (key: string, label: string) => {
      const item = asRecord(props.ownedQuery[key]);
      return {
        label,
        href: linkHref(item),
        active: linkActive(item),
        disabled: false
      };
    };
    const linkStrip = (items: { label: string; href?: string; active?: boolean; disabled?: boolean }[]) => items.length
      ? h("div", { class: "link-strip" }, items.map((item, index) => h("a", {
        key: `${item.label}-${index}`,
        href: item.href || "#",
        class: ["strip-link", { active: item.active, disabled: item.disabled }]
      }, item.label)))
      : null;
    const pager = () => props.pageLinks.length ? h("nav", { class: "pagination-bar", "aria-label": "分页" }, props.pageLinks.map((item, index) => {
        if (item.Ellipsis || item.ellipsis) {
          return h("a", { class: "page-pill disabled", key: `ellipsis-${index}` }, "...");
        }
        const disabled = Boolean(item.Disabled || item.disabled);
        const active = Boolean(item.Active || item.active);
        return h("a", {
          class: ["page-pill", { active, disabled }],
          href: disabled ? undefined : normalizeHref(text(item, "Href", "href") || "#"),
          key: `${text(item, "Label", "label")}-${index}`,
          "aria-disabled": disabled ? "true" : undefined
        }, text(item, "Label", "label"));
      })) : null;
    return () => h("div", { class: "sticky-toolbar cards-toolbar movie-agg-toolbar next02-pink-cards-toolbar", "aria-label": "MovieAgg 影片控制栏" }, [
      h("div", { class: "cards-toolbar-sort" }, [
        linkStrip(sortLabels.map(([key, label]) => sortItem(key, label)))
      ]),
      h("div", { class: "toolbar-line cards-toolbar-bottom" }, [
        h("div", { class: "owned-line cards-toolbar-filters" }, [
          linkStrip([
            ownedItem("MediaAll", "All"),
            ownedItem("MediaOwned", "Owned"),
            ownedItem("MediaNotOwned", "NotOwned")
          ])
        ]),
        h("div", { class: "next02-pink-cards-toolbar-status" }, [
          h("div", { class: "next02-pink-cards-metrics", "aria-label": "MovieAgg 页面摘要" }, [
            h("span", [h("strong", props.totalLabel || "0"), " Total"]),
            h("span", [h("strong", String(props.currentPageCount)), " Current Page"])
          ]),
          props.pageLinks.length ? h("div", { class: "cards-toolbar-pagination next02-pink-cards-pagination" }, [pager()]) : null
        ])
      ])
    ]);
  }
});
</script>
