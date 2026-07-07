<template>
  <PageShell class="torrent-albums-page-shell movie-agg-page-shell" :title="title" :note="note" eyebrow="">
    <div id="page-root" class="container-fluid torrent-albums-page movie-agg-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <template v-if="isReleaseView">
          <section class="section-card movie-agg-release-section">
            <div class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-2">
              <div class="d-flex flex-wrap gap-2">
                <a v-for="item in modeSwitches" :key="linkKey(item)" :href="linkHref(item)" class="btn btn-sm" :class="linkActive(item) ? 'btn-primary' : 'btn-outline-primary'">{{ linkLabel(item) }}</a>
              </div>
            </div>
            <BreadcrumbLine :items="breadcrumbs" />
            <h4 class="section-title">时间聚合</h4>
            <template v-if="hasYearTabs">
              <ul class="nav nav-tabs nav-tabs-custom nav-justified" role="tablist">
                <li v-if="bucketsQAll.length" class="nav-item"><a class="nav-link active" data-bs-toggle="tab" href="#tab-year-quarters" role="tab">季度</a></li>
                <li v-if="bucketsMAll.length" class="nav-item"><a class="nav-link" :class="{ active: !bucketsQAll.length }" data-bs-toggle="tab" href="#tab-year-months" role="tab">月份</a></li>
              </ul>
              <div class="tab-content pt-3">
                <BucketRows v-if="bucketsQAll.length" id="tab-year-quarters" class="tab-pane p-2 active" :items="bucketsQAll" grid-class="bucket-grid bucket-grid-quarter" />
                <BucketRows v-if="bucketsMAll.length" id="tab-year-months" class="tab-pane p-2" :class="{ active: !bucketsQAll.length }" :items="bucketsMAll" grid-class="bucket-grid bucket-grid-month" />
              </div>
            </template>
            <BucketRows v-else-if="bucketsAll.length" :items="bucketsAll" grid-class="bucket-grid bucket-grid-default" />
            <div v-else class="alert alert-info mb-0">暂无聚合数据。</div>
          </section>

          <section v-if="topGroups.length" class="section-card movie-agg-top-section">
            <div class="d-flex align-items-center justify-content-start mb-2">
              <h4 class="section-title mb-0 me-3">Top 聚合（{{ aggModeLabel }}）</h4>
              <button type="button" class="btn btn-outline-info waves-effect waves-light btn-sm" @click="topExpanded = !topExpanded">展开</button>
            </div>
            <div class="collapse" :class="{ show: topExpanded }" id="topAggCollapse">
              <ul class="nav nav-tabs nav-tabs-custom nav-justified" role="tablist">
                <li v-for="(group, index) in topGroups" :key="group.type" class="nav-item">
                  <a class="nav-link" :class="{ active: index === 0 }" data-bs-toggle="tab" :href="`#tab-top-${group.type}`" role="tab">{{ group.label }}</a>
                </li>
              </ul>
              <div class="tab-content pt-3">
                <div v-for="(group, index) in topGroups" :key="`${group.type}-pane`" class="tab-pane p-2" :class="{ active: index === 0 }" :id="`tab-top-${group.type}`" role="tabpanel">
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

          <section v-if="moviesProvided" class="section-card movie-agg-movies-section">
            <MovieToolbar :sort-query="sortQuery" :owned-query="ownedQuery" :total-label="totalLabel" :page-links="pageLinks" />
            <MovieGrid v-if="movies.length" :cards="movies" />
            <div v-else class="row"><div class="col-12"><div class="alert alert-info mb-0">暂无影片。</div></div></div>
          </section>
        </template>

        <template v-else>
          <section class="section-card">
            <BreadcrumbLine :items="breadcrumbs" />
            <h4 class="section-title">{{ title }}</h4>
            <BucketRows v-for="section in fallbackBucketSections" :key="section.key" :title="section.title" :items="section.items" grid-class="bucket-grid bucket-grid-default" />
          </section>
          <section v-if="topGroups.length" class="section-card movie-agg-top-section">
            <div class="d-flex align-items-center justify-content-start mb-2">
              <h4 class="section-title mb-0 me-3">Top 聚合</h4>
              <button type="button" class="btn btn-outline-info waves-effect waves-light btn-sm" @click="topExpanded = !topExpanded">展开</button>
            </div>
            <div class="collapse" :class="{ show: topExpanded }" id="topAggCollapse">
              <ul class="nav nav-tabs nav-tabs-custom nav-justified" role="tablist">
                <li v-for="(group, index) in topGroups" :key="group.type" class="nav-item">
                  <a class="nav-link" :class="{ active: index === 0 }" data-bs-toggle="tab" :href="`#tab-top-${group.type}`" role="tab">{{ group.label }}</a>
                </li>
              </ul>
              <div class="tab-content pt-3">
                <div v-for="(group, index) in topGroups" :key="`${group.type}-pane`" class="tab-pane p-2" :class="{ active: index === 0 }" :id="`tab-top-${group.type}`" role="tabpanel">
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
          <section v-if="movies.length" class="section-card">
            <MovieToolbar :sort-query="sortQuery" :owned-query="ownedQuery" :total-label="totalLabel" :page-links="pageLinks" />
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
  ["ByMediaBirthTime", "Media下载时间"],
  ["ByLastScTime", "最后sc"],
  ["ByScTimes", "sc次数"],
  ["ByComeTimes", "co次数"]
];

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const topExpanded = ref(false);

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
    topExpanded.value = false;
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title") || String(route.meta.title || "MovieAgg"));
const aggModeLabel = computed(() => text(payload.value, "AggModeLabel"));
const note = computed(() => [aggModeLabel.value, text(payload.value, "RangeStart"), text(payload.value, "RangeEnd")].filter(Boolean).join(" · "));
const isReleaseView = computed(() => text(payload.value, "ViewType") === "movie-agg-release" || route.path === "/movie-agg-all/release");
const isWMediaBirth = computed(() => text(payload.value, "ViewType") === "w-media-agg-birth" || route.path === "/w-media-agg/birth");
const breadcrumbs = computed(() => asArray<AnyRecord>(pick(payload.value, "Breadcrumbs")));
const modeSwitches = computed(() => asArray<AnyRecord>(pick(payload.value, "ModeSwitches")));
const bucketsAll = computed(() => asArray<AnyRecord>(pick(payload.value, "BucketsAll")));
const bucketsQAll = computed(() => asArray<AnyRecord>(pick(payload.value, "BucketsQAll")));
const bucketsMAll = computed(() => asArray<AnyRecord>(pick(payload.value, "BucketsMAll")));
const hasYearTabs = computed(() => bucketsQAll.value.length > 0 || bucketsMAll.value.length > 0);
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
    return () => h("div", { class: "bucket-breadcrumbs mb-2" }, props.items.map((item, index) => [
      index > 0 ? h("span", { class: "text-muted mx-1" }, "/") : null,
      h("a", { href: linkHref(item), class: "text-muted text-decoration-none me-1" }, linkLabel(item))
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
      h("div", { class: ["row", "g-3", props.gridClass] }, props.items.map((item) => h("div", { class: "col", key: bucketHref(item) }, [
        h("a", { href: bucketHref(item), class: "text-decoration-none" }, [
          h("div", { class: "card bucket-card h-100 shadow-sm" }, [
            h("div", { class: "card-body py-2" }, [
              h("div", { class: "fw-semibold" }, bucketLabel(item)),
              h("div", { class: "text-muted small" }, pick(item, "Count") !== undefined
                ? [
                  h("span", { class: "count" }, `${bucketCountAll(item)} 部`),
                  "，",
                  h("span", { class: "sizegb" }, `${bucketSizeGb(item)} GB`)
                ]
                : [
                  h("span", { class: "count" }, `总数：${bucketCountAll(item)}`),
                  "，",
                  h("span", { class: "count" }, `已拥有：${bucketCountOwned(item)}`),
                  h("br"),
                  h("span", { class: "sizegb" }, `${bucketSizeGb(item)} GB`)
                ])
            ])
          ])
        ])
      ])))
    ]);
  }
});

const MovieToolbar = defineComponent({
  name: "MovieAggToolbar",
  props: {
    sortQuery: { type: Object as PropType<AnyRecord>, required: true },
    ownedQuery: { type: Object as PropType<AnyRecord>, required: true },
    totalLabel: { type: String, required: true },
    pageLinks: { type: Array as PropType<AnyRecord[]>, required: true }
  },
  setup(props) {
    const sortButton = (key: string, label: string) => {
      const item = asRecord(props.sortQuery[key]);
      return h("a", {
        href: linkHref(item),
        class: ["btn", "btn-outline-primary", { "active disabled": linkActive(item) }]
      }, label);
    };
    const ownedButton = (key: string, label: string, cls: string) => {
      const item = asRecord(props.ownedQuery[key]);
      return h("a", {
        href: linkHref(item),
        class: ["btn", "btn-sm", cls, "waves-effect", "waves-light", { "active disabled": linkActive(item) }]
      }, label);
    };
    const pager = () => props.pageLinks.length ? h("nav", { class: "mb-0", "aria-label": "Page navigation" }, [
      h("ul", { class: "pagination pagination-sm mb-0" }, props.pageLinks.map((item, index) => {
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
    ]) : null;
    return () => h("div", { class: "position-sticky pb-2 pt-2 mb-3 border-bottom movie-agg-toolbar" }, [
      h("div", { class: "container-fluid px-0" }, [
        h("div", { class: "d-flex flex-wrap justify-content-start align-items-center mb-2" }, [
          h("div", { class: "btn-group btn-group-sm flex-wrap", role: "group", "aria-label": "Sort Options" }, sortLabels.map(([key, label]) => sortButton(key, label)))
        ]),
        h("div", { class: "d-flex flex-column flex-md-row justify-content-between align-items-md-center gap-2" }, [
          h("div", { class: "d-flex flex-wrap align-items-center gap-2" }, [
            h("span", { class: "btn btn-sm btn-outline-dark disabled" }, "WMedia"),
            ownedButton("MediaAll", "MAll", "btn-primary"),
            ownedButton("MediaOwned", "MOwned", "btn-success"),
            ownedButton("MediaNotOwned", "MNotOwned", "btn-secondary"),
            h("span", { class: "btn btn-sm btn-primary waves-effect waves-light disabled" }, `Total: ${props.totalLabel || "0"}`)
          ]),
          h("div", { class: "d-flex align-items-center ms-md-3" }, [pager()])
        ])
      ])
    ]);
  }
});
</script>
