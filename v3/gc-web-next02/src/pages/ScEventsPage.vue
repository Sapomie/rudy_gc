<template>
  <PageShell class="movie-agg-page-shell sc-events-page-shell" :title="title" :note="pageNote" eyebrow="">
    <div id="page-root" class="container-fluid movie-agg-page-root sc-events-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <template v-if="viewType === 'sc-events'">
          <div class="d-flex align-items-center justify-content-between mb-3 sc-page-head">
            <div>
              <h4 class="mb-1">SC Events</h4>
              <p class="text-muted mb-0">浏览最近的 SC 事件，并进入详情页查看本次影片清单。</p>
            </div>
            <div class="d-flex flex-wrap gap-2">
              <ScViewSwitch />
              <a href="/triggers/backfill" class="btn btn-outline-secondary">回填任务台</a>
            </div>
          </div>

          <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
            <div class="sc-page-summary">
              <span class="sc-page-summary-label">Total Events</span>
              <span class="sc-page-summary-value">{{ total }}</span>
            </div>
            <div v-if="items.length" class="ms-xl-auto">
              <PageLinks :links="pageLinks" />
            </div>
          </div>

          <div class="card">
            <div class="card-body">
              <div class="table-responsive">
                <table class="table table-hover align-middle mb-0 sc-events-table">
                  <thead>
                    <tr>
                      <th><SortLink sort-key="ByTime" label="时间" /></th>
                      <th>总片数</th>
                      <th><SortLink sort-key="ByMovieCnt" label="有 SC" /></th>
                      <th><SortLink sort-key="ByComeMovie" label="Come 片" /></th>
                      <th><SortLink sort-key="ByCooldown" label="Cooldown" /></th>
                      <th><SortLink sort-key="ByDuration" label="时长(分钟)" /></th>
                      <th><SortLink sort-key="ByKind" label="种类" /></th>
                      <th><SortLink sort-key="ByMovieCast" label="演员" /></th>
                      <th><SortLink sort-key="ByVessel" label="Vessel" /></th>
                      <th><SortLink sort-key="ByFg" label="FG" /></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in items" :key="eventName(eventOf(item))" :class="monthParityClass(pick(eventOf(item), 'ScTime'))">
                      <td>
                        <a class="sc-event-link" :href="eventDetailHref(eventName(eventOf(item)))" :title="eventName(eventOf(item))">
                          <span class="sc-event-time"><span class="sc-event-date">{{ formatUnix(pick(eventOf(item), "ScTime")) }}</span> <span class="sc-event-hour">{{ formatUnixHm(pick(eventOf(item), "ScTime")) }}</span></span>
                        </a>
                      </td>
                      <td>{{ text(item, "TotalMovieCount") }}</td>
                      <td>{{ text(item, "ScMovieCount") }}</td>
                      <td><LinkOrMuted :href="movieHref(text(eventOf(item), 'ComeMovieName'))" :value="text(eventOf(item), 'ComeMovieName')" /></td>
                      <td><span v-if="num(eventOf(item), 'Cooldown') > 0">{{ humanDays(pick(eventOf(item), "Cooldown")) }}</span><span v-else class="text-muted">-</span></td>
                      <td class="sc-muted"><span v-if="num(eventOf(item), 'DurationMinutes') > 0">{{ text(eventOf(item), "DurationMinutes") }}</span><span v-else class="text-muted">-</span></td>
                      <td class="sc-muted"><span v-if="text(eventOf(item), 'Kind')">{{ text(eventOf(item), "Kind") }}</span><span v-else class="text-muted">-</span></td>
                      <td class="sc-muted"><LinkOrMuted :href="text(item, 'MovieCastHref')" :value="text(eventOf(item), 'MovieCast')" /></td>
                      <td class="sc-muted"><span v-if="text(eventOf(item), 'Vessel')">{{ text(eventOf(item), "Vessel") }}</span><span v-else class="text-muted">-</span></td>
                      <td class="sc-muted"><span v-if="text(eventOf(item), 'Fg')">{{ text(eventOf(item), "Fg") }}</span><span v-else class="text-muted">-</span></td>
                    </tr>
                    <tr v-if="!items.length">
                      <td colspan="10"><div class="sc-empty-state">还没有可展示的 SC 事件。</div></td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </template>

        <template v-else-if="viewType === 'sc-event-cards'">
          <div class="d-flex align-items-center justify-content-between mb-3 sc-page-head">
            <div>
              <h4 class="mb-1">SC Event Cards</h4>
              <p class="text-muted mb-0">以封面卡片的方式浏览 SC 事件，主图使用 come movie 的电影封面。</p>
            </div>
            <div class="d-flex flex-wrap gap-2">
              <ScViewSwitch />
              <a href="/triggers/backfill" class="btn btn-outline-secondary">回填任务台</a>
            </div>
          </div>

          <div class="d-flex flex-column flex-xl-row align-items-xl-start justify-content-between gap-3 mb-3">
            <div class="d-flex align-items-center gap-3 flex-wrap">
              <div class="sc-page-summary">
                <span class="sc-page-summary-label">Total Events</span>
                <span class="sc-page-summary-value">{{ total }}</span>
              </div>
              <div class="d-flex flex-wrap gap-2 sc-event-sort-bar">
                <SortLink sort-key="ByTime" label="时间" button />
                <SortLink sort-key="ByMovieCnt" label="片数" button />
                <SortLink sort-key="ByComeMovie" label="Come 片" button />
                <SortLink sort-key="ByCooldown" label="Cooldown" button />
                <SortLink sort-key="ByKind" label="种类" button />
                <SortLink sort-key="ByMovieCast" label="演员" button />
                <SortLink sort-key="ByVessel" label="Vessel" button />
                <SortLink sort-key="ByFg" label="FG" button />
              </div>
            </div>
            <div class="ms-xl-auto">
              <PageLinks :links="pageLinks" />
            </div>
          </div>

          <div v-if="items.length" class="row row-cols-2 row-cols-sm-3 row-cols-lg-5 row-cols-xxl-6 g-3 sc-event-card-grid">
            <div v-for="item in items" :key="eventName(eventOf(item))" class="col">
              <article class="card shadow-sm h-100 card-lift sc-event-grid-card" :class="monthParityClass(pick(eventOf(item), 'ScTime'))">
                <a :href="text(item, 'DetailHref') || eventDetailHref(eventName(eventOf(item)))" class="text-decoration-none text-reset">
                  <div class="thumb-2x3 sc-event-cover">
                    <img v-if="item.HasComeMovieCover" :src="text(item, 'ComeMovieJacketImg')" :alt="`${eventName(eventOf(item))} 封面`" loading="lazy" decoding="async">
                    <div v-else class="sc-event-cover-empty">
                      <span class="sc-event-cover-kicker">SC EVENT</span>
                      <strong><span class="sc-event-time"><span class="sc-event-date">{{ formatUnix(pick(eventOf(item), "ScTime")) }}</span><span class="sc-event-hour">{{ formatUnixHm(pick(eventOf(item), "ScTime")) }}</span></span></strong>
                      <span>{{ eventName(eventOf(item)) }}</span>
                    </div>
                    <div class="sc-event-cover-badge">{{ text(eventOf(item), "MovieNumber") }} 部</div>
                  </div>
                </a>
                <div class="card-body sc-event-card-body">
                  <div class="small mb-2"><span class="sc-event-time"><span class="sc-event-date">{{ formatUnix(pick(eventOf(item), "ScTime")) }}</span> <span class="sc-event-hour">{{ formatUnixHm(pick(eventOf(item), "ScTime")) }}</span></span></div>
                  <h5 class="h6 mb-2"><a :href="text(item, 'DetailHref') || eventDetailHref(eventName(eventOf(item)))" class="text-decoration-none text-reset">{{ eventName(eventOf(item)) }}</a></h5>
                  <div class="mb-3">
                    <div class="text-muted small text-uppercase mb-1">Come Movie</div>
                    <LinkOrMuted :href="text(item, 'ComeMovieHref')" :value="text(item, 'ComeMovieName')" />
                  </div>
                  <div class="sc-event-meta-grid">
                    <span class="sc-pill">{{ num(eventOf(item), "Cooldown") ? humanDays(pick(eventOf(item), "Cooldown")) : "Cooldown -" }}</span>
                    <span class="sc-pill">{{ text(eventOf(item), "Kind") ? `种类 ${text(eventOf(item), "Kind")}` : "种类 -" }}</span>
                    <span class="sc-pill">{{ text(eventOf(item), "MovieCast") || "演员 -" }}</span>
                    <span class="sc-pill">{{ text(eventOf(item), "Vessel") ? `Vessel ${text(eventOf(item), "Vessel")}` : "Vessel -" }}</span>
                    <span class="sc-pill">{{ text(eventOf(item), "Fg") ? `FG ${text(eventOf(item), "Fg")}` : "FG -" }}</span>
                  </div>
                </div>
              </article>
            </div>
          </div>
          <div v-else class="sc-empty-state">还没有可展示的 SC 事件。</div>
        </template>

        <template v-else-if="viewType === 'sc-event-detail'">
          <div class="d-flex align-items-center justify-content-between mb-3 sc-page-head">
            <div>
              <h4 class="mb-1">{{ eventName(event) }}</h4>
              <p class="text-muted mb-0">本次 SC 的基础信息、影片清单与快速操作入口。</p>
            </div>
            <div class="sc-action-row">
              <button type="button" class="btn btn-outline-primary" id="btnOpenScEventEditModal" data-bs-toggle="modal" data-bs-target="#scEventEditModal">编辑信息</button>
              <a href="/sc-events" class="btn btn-outline-secondary">返回列表</a>
              <a :href="`/triggers/sc-media?sc_name=${encodeURIComponent(eventName(event))}`" class="btn btn-outline-primary">移动 SC Media</a>
              <a href="/triggers/backfill" class="btn btn-outline-secondary">回填任务台</a>
            </div>
          </div>
          <div id="scEventPageMsg" class="alert d-none mb-3" role="alert"></div>
          <ScEventEditModal />

          <div class="row g-3 mb-3">
            <div v-if="text(event, 'ImagePath')" class="col-12 col-xl-6">
              <div class="card h-100 sc-detail-side-card">
                <div class="card-body">
                  <div class="text-muted small text-uppercase mb-2">SC Image</div>
                  <div class="sc-detail-image-wrap">
                    <img :src="text(event, 'ImagePath')" :alt="`${eventName(event)} 图片`" class="sc-detail-image">
                  </div>
                </div>
              </div>
            </div>
            <div class="col-12" :class="{ 'col-xl-6': text(event, 'ImagePath') }">
              <div class="card h-100 sc-detail-side-card">
                <div class="card-body">
                  <div class="sc-detail-info-grid">
                    <MetricCard label="Sc Time" :value="formatUnixFull(pick(event, 'ScTime'))" />
                    <MetricCard label="Movie Count" :value="text(event, 'MovieNumber') || '0'" />
                    <MetricCard label="Cooldown" :value="num(event, 'Cooldown') > 0 ? humanDays(pick(event, 'Cooldown')) : '-'" />
                    <MetricCard label="Come Movie" :value="text(event, 'ComeMovieName') || '-'" :href="movieHref(text(event, 'ComeMovieName'))" />
                    <MetricCard label="种类" :value="text(event, 'Kind') || '-'" />
                    <MetricCard label="Lead Cast" :value="text(event, 'MovieCast') || '-'" :href="text(event, 'MovieCast') ? `/cards?cn=${encodeURIComponent(text(event, 'MovieCast'))}` : ''" />
                    <MetricCard label="时长(分钟)" :value="num(event, 'DurationMinutes') > 0 ? `${text(event, 'DurationMinutes')} 分钟` : '-'" />
                    <MetricCard label="FG" :value="text(event, 'Fg') || '-'" />
                    <MetricCard label="Vessel" :value="text(event, 'Vessel') || '-'" />
                    <div class="sc-detail-info-item sc-metric-card sc-detail-remarks">
                      <div class="sc-metric-label">Remarks</div>
                      <div class="sc-metric-value"><span v-if="text(event, 'Remarks')">{{ text(event, "Remarks") }}</span><span v-else class="text-muted">暂无备注。</span></div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="d-flex align-items-center justify-content-between mb-3">
            <div class="d-flex align-items-center gap-3 flex-wrap">
              <h5 class="mb-0">Movies</h5>
              <div class="btn-group btn-group-sm" role="group" aria-label="电影筛选">
                <a :href="text(payload, 'MovieFilterAllHref')" class="btn" :class="movieFilter === 'all' ? 'btn-primary' : 'btn-outline-primary'">全部电影</a>
                <a :href="text(payload, 'MovieFilterScHref')" class="btn" :class="movieFilter === 'sc' ? 'btn-primary' : 'btn-outline-primary'">有 SC</a>
                <a :href="text(payload, 'MovieFilterNoScHref')" class="btn" :class="movieFilter === 'nosc' ? 'btn-primary' : 'btn-outline-primary'">无 SC</a>
              </div>
            </div>
            <span class="text-muted small">{{ items.length }} items</span>
          </div>

          <section v-if="items.length" aria-label="SC Movies">
            <div class="movie-grid-v2 sc-movie-grid" id="movie-grid">
              <div v-for="(item, index) in items" :key="movieKey(item, index)" class="col movie-grid-col-v2">
                <ScMovieCard :item="item" :event-name="eventName(event)" />
              </div>
            </div>
          </section>
          <div v-else class="sc-empty-state">{{ emptyDetailText }}</div>

          <div v-if="failedItems.length" class="card mt-3">
            <div class="card-body">
              <h5 class="mb-3">Skipped / Error Items</h5>
              <div class="table-responsive">
                <table class="table table-sm align-middle mb-0">
                  <thead><tr><th>Movie</th><th>JavId</th><th>Error</th></tr></thead>
                  <tbody>
                    <tr v-for="(item, index) in failedItems" :key="index">
                      <td>{{ text(item, "MovieName") }}</td>
                      <td>{{ text(asRecord(item.Entry), "MovieJavId") }}</td>
                      <td class="text-danger small">{{ text(item, "LoadError") }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </template>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, normalizeHref, num, pick, text } from "@/utils/format";

const GLIST_IS_NOT_SC = 1;
const GLIST_IS_SC = 2;

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

const title = computed(() => text(payload.value, "Title", "PageTitle") || "SC Events");
const pageNote = computed(() => text(payload.value, "PageNote"));
const viewType = computed(() => text(payload.value, "ViewType"));
const items = computed(() => asArray<AnyRecord>(pick(payload.value, "Items")));
const failedItems = computed(() => asArray<AnyRecord>(pick(payload.value, "FailedItems")));
const total = computed(() => text(payload.value, "Total") || String(items.value.length));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")));
const pageLinks = computed(() => asArray<AnyRecord>(pageInfo.value.Links || pageInfo.value.links));
const event = computed(() => asRecord(pick(payload.value, "Event")));
const movieFilter = computed(() => text(payload.value, "MovieFilter") || "sc");
const sortQuery = computed(() => asRecord(pick(payload.value, "ScSortQuery")));
const scViewQuery = computed(() => asRecord(pick(payload.value, "ScViewQuery")));
const emptyDetailText = computed(() => {
  if (movieFilter.value === "all") return "本次 SC 还没有电影明细。";
  if (movieFilter.value === "nosc") return "当前筛选下没有 is_sc=1 的电影。";
  return "当前筛选下没有 is_sc=2 的电影。";
});

function eventOf(item: AnyRecord): AnyRecord {
  return asRecord(pick(item, "Event"));
}

function eventName(row: AnyRecord): string {
  return text(row, "Name", "name");
}

function eventDetailHref(name: string): string {
  return `/sc-events/${encodeURI(name || "")}`;
}

function movieHref(name: string): string {
  return name ? `/movie/${encodeURI(name)}` : "";
}

function sortItem(key: string): AnyRecord {
  return asRecord(sortQuery.value[key]);
}

function sortLabel(key: string, label: string): string {
  const item = sortItem(key);
  if (!item.Active) return label;
  return `${label}${item.Desc ? " ↓" : " ↑"}`;
}

function formatUnix(input: unknown, withTime = false): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  if (!withTime) return `${yyyy}-${mm}-${dd}`;
  const hh = String(d.getHours()).padStart(2, "0");
  const mi = String(d.getMinutes()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}`;
}

function formatUnixHm(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}

function formatUnixFull(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  const hh = String(d.getHours()).padStart(2, "0");
  const mi = String(d.getMinutes()).padStart(2, "0");
  const ss = String(d.getSeconds()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}:${ss}`;
}

function humanDays(input: unknown): string {
  const n = Number(input || 0);
  if (n <= 0) return "-";
  return `${(n / 86400).toFixed(2)} 天`;
}

function monthParityClass(input: unknown): string {
  const n = Number(input || 0);
  if (n <= 0) return "sc-month-even";
  const month = new Date(n * 1000).getMonth() + 1;
  return month % 2 === 0 ? "sc-month-even" : "sc-month-odd";
}

function movieKey(item: AnyRecord, index: number): string {
  const movie = asRecord(pick(item, "MovieType"));
  return text(movie, "name", "Name", "jav_id", "JavID") || String(index);
}

function humanMovieSize(input: unknown): string {
  const value = Number(input || 0);
  if (value >= 1024 ** 4) return `${(value / 1024 ** 4).toFixed(1)} TB`;
  if (value >= 1024 ** 3) return `${(value / 1024 ** 3).toFixed(1)} GB`;
  if (value >= 1024 ** 2) return `${(value / 1024 ** 2).toFixed(1)} MB`;
  if (value >= 1024) return `${(value / 1024).toFixed(1)} KB`;
  return value ? `${value} B` : "";
}

const PageLinks = defineComponent({
  name: "ScPageLinks",
  props: {
    links: { type: Array as PropType<AnyRecord[]>, required: true }
  },
  setup(props) {
    return () => props.links.length ? h("nav", { "aria-label": "Page navigation" }, [
      h("ul", { class: "pagination pagination-sm mb-0" }, props.links.map((item, index) => {
        if (item.Ellipsis || item.ellipsis) {
          return h("li", { class: "page-item disabled", key: `ellipsis-${index}` }, [h("span", { class: "page-link" }, "...")]);
        }
        const disabled = Boolean(item.Disabled || item.disabled);
        const active = Boolean(item.Active || item.active);
        return h("li", { class: ["page-item", { active, disabled }], key: `${text(item, "Label", "label")}-${index}` }, [
          h("a", {
            class: "page-link",
            href: disabled ? "#" : normalizeHref(text(item, "Href", "href")),
            tabindex: disabled ? "-1" : undefined,
            "aria-disabled": disabled ? "true" : undefined
          }, text(item, "Label", "label"))
        ]);
      }))
    ]) : null;
  }
});

const LinkOrMuted = defineComponent({
  name: "ScLinkOrMuted",
  props: {
    href: { type: String, default: "" },
    value: { type: String, default: "" }
  },
  setup(props) {
    return () => props.value
      ? (props.href ? h("a", { href: props.href }, props.value) : h("span", props.value))
      : h("span", { class: "text-muted" }, "-");
  }
});

const SortLink = defineComponent({
  name: "ScSortLink",
  props: {
    sortKey: { type: String, required: true },
    label: { type: String, required: true },
    button: { type: Boolean, default: false }
  },
  setup(props) {
    return () => {
      const item = sortItem(props.sortKey);
      const cls = props.button
        ? ["btn", "btn-sm", "btn-outline-primary", { "active disabled": Boolean(item.Active) }]
        : ["sc-sort-link", { active: Boolean(item.Active) }];
      return h("a", { href: normalizeHref(text(item, "Href")), class: cls }, sortLabel(props.sortKey, props.label));
    };
  }
});

const ScViewSwitch = defineComponent({
  name: "ScViewSwitch",
  setup() {
    return () => h("div", { class: "btn-group", role: "group", "aria-label": "SC event view switch" }, [
      h("a", {
        href: normalizeHref(text(scViewQuery.value, "ListHref"), "/sc-events"),
        class: ["btn", "btn-outline-primary", { "active disabled": Boolean(scViewQuery.value.IsList) }]
      }, "列表视图"),
      h("a", {
        href: normalizeHref(text(scViewQuery.value, "CardHref"), "/sc-events-cards"),
        class: ["btn", "btn-outline-primary", { "active disabled": Boolean(scViewQuery.value.IsCard) }]
      }, "卡片视图")
    ]);
  }
});

const MetricCard = defineComponent({
  name: "ScMetricCard",
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
    href: { type: String, default: "" }
  },
  setup(props) {
    return () => h("div", { class: "sc-detail-info-item sc-metric-card" }, [
      h("div", { class: "sc-metric-label" }, props.label),
      h("div", { class: "sc-metric-value" }, props.href && props.value !== "-" ? [h("a", { href: props.href }, props.value)] : props.value)
    ]);
  }
});

const ScEventEditModal = defineComponent({
  name: "ScEventEditModal",
  setup() {
    return () => h("div", { class: "modal fade", id: "scEventEditModal", tabindex: "-1", "aria-hidden": "true" }, [
      h("div", { class: "modal-dialog modal-lg modal-dialog-scrollable" }, [
        h("div", { class: "modal-content" }, [
          h("div", { class: "modal-header" }, [
            h("h5", { class: "modal-title" }, "编辑 SC Event"),
            h("button", { type: "button", class: "btn-close", "data-bs-dismiss": "modal", "aria-label": "Close" })
          ]),
          h("div", { class: "modal-body" }, [
            h("form", { id: "scEventEditForm", "data-update-url": `/api/sc-events/${encodeURIComponent(eventName(event.value))}` }, [
              h("div", { class: "row g-3" }, [
                h("div", { class: "col-12 col-md-6" }, [h("label", { class: "form-label" }, "Name"), h("input", { class: "form-control", id: "scEventEditName", type: "text", value: eventName(event.value), readonly: true })]),
                h("div", { class: "col-12 col-md-6" }, [h("label", { class: "form-label" }, "Come Movie"), h("input", { class: "form-control", id: "scEventEditComeMovie", type: "text", value: text(event.value, "ComeMovieName") })]),
                h("div", { class: "col-12 col-md-6" }, [h("label", { class: "form-label" }, "Lead Cast"), h("input", { class: "form-control", id: "scEventEditMovieCast", type: "text", value: text(event.value, "MovieCast") })]),
                h("div", { class: "col-12 col-md-6" }, [h("label", { class: "form-label" }, "Kind"), h("input", { class: "form-control", id: "scEventEditKind", type: "text", value: text(event.value, "Kind") })]),
                h("div", { class: "col-12 col-md-6" }, [h("label", { class: "form-label" }, "Duration"), h("input", { class: "form-control", id: "scEventEditDuration", type: "number", value: text(event.value, "DurationMinutes") })]),
                h("div", { class: "col-12 col-md-6" }, [h("label", { class: "form-label" }, "FG"), h("input", { class: "form-control", id: "scEventEditFg", type: "text", value: text(event.value, "Fg") })]),
                h("div", { class: "col-12" }, [h("label", { class: "form-label" }, "Remarks"), h("textarea", { class: "form-control", id: "scEventEditRemarks", rows: "3", value: text(event.value, "Remarks") })])
              ])
            ])
          ]),
          h("div", { class: "modal-footer" }, [
            h("button", { type: "button", class: "btn btn-outline-secondary", "data-bs-dismiss": "modal" }, "取消"),
            h("button", { type: "submit", form: "scEventEditForm", class: "btn btn-primary" }, "保存")
          ])
        ])
      ])
    ]);
  }
});

const ScMovieCard = defineComponent({
  name: "ScMovieCard",
  props: {
    item: { type: Object as PropType<AnyRecord>, required: true },
    eventName: { type: String, required: true }
  },
  setup(props) {
    const movie = () => asRecord(pick(props.item, "MovieType"));
    const entry = () => asRecord(pick(props.item, "Entry"));
    const movieName = () => text(movie(), "name", "Name", "jav_id", "JavID");
    const movieTitle = () => text(movie(), "title", "Title");
    const statusControls = () => {
      const movieJavID = text(entry(), "MovieJavId");
      const isSc = num(entry(), "IsSc") === GLIST_IS_SC;
      const isCome = num(entry(), "IsCome") === 2;
      if (!movieJavID || !props.eventName) return null;
      const apiURL = `/api/sc-events/${encodeURIComponent(props.eventName)}/movies/${encodeURIComponent(movieJavID)}/status`;
      return h("li", { class: "list-group-item position-relative", style: "z-index:2;" }, [
        h("div", { class: "d-flex align-items-center justify-content-between gap-2 flex-wrap" }, [
          h("div", { class: "d-flex align-items-center gap-2 flex-wrap" }, [
            h("span", { class: ["badge", isSc ? "bg-success" : "bg-secondary"] }, isSc ? "有 SC" : "无 SC"),
            isCome ? h("span", { class: "badge bg-danger" }, "Come") : null
          ]),
          h("div", { class: "btn-group btn-group-sm", role: "group", "aria-label": "编辑影片 SC 状态" }, [
            h("button", { type: "button", class: ["btn", isSc ? "btn-success" : "btn-outline-success", "js-sc-event-movie-status"], "data-url": apiURL, "data-is-sc": GLIST_IS_SC, disabled: isSc || undefined }, "设为有 SC"),
            h("button", { type: "button", class: ["btn", isSc ? "btn-outline-secondary" : "btn-secondary", "js-sc-event-movie-status"], "data-url": apiURL, "data-is-sc": GLIST_IS_NOT_SC, disabled: !isSc || undefined }, "设为无 SC")
          ])
        ])
      ]);
    };
    const moviePlayRow = () => {
      const ownedMedia = num(movie(), "owned_w_media", "OwnedWMedia");
      if (ownedMedia < 4) return null;
      const videoUrl = text(movie(), "video_url_w_media", "VideoUrlWMedia");
      const mediaBirth = text(movie(), "film_birth_date_w_media", "FilmBirthDateWMedia");
      const play = ownedMedia === 6
        ? h("span", { class: "btn btn-sm btn-secondary disabled", "aria-disabled": "true" }, "MRemoved")
        : h("a", { href: videoUrl, class: ["btn", "btn-sm", ownedMedia === 5 ? "btn-info" : "btn-success"] }, "MPlay");
      return h("li", { class: "list-group-item py-2" }, [
        h("div", { class: "card-play-stack" }, [h("div", { class: "card-play-row" }, [
          play,
          mediaBirth ? h("span", { class: "text-muted small card-play-date" }, mediaBirth) : null
        ])])
      ]);
    };
    const genreLinks = () => asArray<string>(pick(movie(), "genre", "Genre")).map((genre) => h("a", { href: `/cards?gn=${encodeURIComponent(genre)}`, class: "badge badge-genre me-1 mb-1" }, String(genre)));
    const castLinks = () => asArray<AnyRecord>(pick(movie(), "cast", "Cast")).map((cast) => {
      const label = text(cast, "NameShow", "DisplayName", "Name", "name");
      const personID = text(cast, "PersonId", "person_id", "Id", "id");
      return h("a", { href: personID && personID !== "0" ? `/cast/${encodeURIComponent(personID)}` : `/cards?cn=${encodeURIComponent(label)}`, class: "person-tag mr-2", title: label }, label);
    });
    return () => h("article", { class: "card shadow-sm h-100 card-lift movie-card-v2" }, [
      h("a", { href: movieHref(movieName()), class: "d-block poster-link", "aria-label": `查看 ${movieName()} 详情` }, [
        h("div", { class: "thumb-2x3" }, [
          text(movie(), "jacket_img", "JacketImg") ? h("img", { src: text(movie(), "jacket_img", "JacketImg"), alt: `${movieName()} 海报`, loading: "lazy", decoding: "async" }) : h("div", { class: "thumb-empty" }, "No Image"),
          h("div", { class: "badge-top-left" }, [
            num(movie(), "come_times", "ComeTimes") ? h("span", { class: "badge-come" }, `❤ ${num(movie(), "come_times", "ComeTimes")}`) : null,
            num(movie(), "sc_times", "ScTimes") ? h("span", { class: "badge-times" }, `× ${num(movie(), "sc_times", "ScTimes")}`) : null
          ]),
          h("div", { class: "badge-bottom-right" }, [
            num(movie(), "highest_rank", "HighestRank") ? h("span", { class: "badge-rank" }, `#${num(movie(), "highest_rank", "HighestRank")}`) : null,
            text(movie(), "score", "Score") ? h("span", { class: "badge-score" }, `★ ${text(movie(), "score", "Score")}`) : null
          ])
        ])
      ]),
      h("ul", { class: "list-group list-group-flush" }, [
        statusControls(),
        h("li", { class: "list-group-item" }, [
          h("h6", { class: "mb-1 text-truncate", title: movieTitle() }, [
            h("a", { href: `/cards?pn=${encodeURIComponent(text(movie(), "prefix", "Prefix"))}`, class: "stretched-link text-reset text-decoration-none movie-name" }, movieName() || "-"),
            text(movie(), "viewers_number_watched", "ViewersNumberWatched") ? h("small", { class: "text-muted ml-2" }, `(${text(movie(), "viewers_number_watched", "ViewersNumberWatched")})`) : null
          ]),
          movieTitle() ? h("p", { class: "mb-0 text-muted line-clamp-2", title: movieTitle() }, movieTitle()) : null
        ]),
        moviePlayRow(),
        h("li", { class: "list-group-item text-muted" }, genreLinks()),
        h("li", { class: "list-group-item" }, [
          ...castLinks(),
          text(movie(), "director", "Director") ? h("a", { href: `/cards?dn=${encodeURIComponent(text(movie(), "director", "Director"))}`, class: "person-tag ml-2", title: text(movie(), "director", "Director") }, text(movie(), "director", "Director")) : null
        ]),
        h("li", { class: "list-group-item d-flex align-items-center small card-link-row" }, [
          text(movie(), "bus_url", "BusUrl") ? h("a", { href: text(movie(), "bus_url", "BusUrl"), class: "me-2", target: "_blank", rel: "noopener" }, "Bus") : null,
          text(movie(), "search_url", "SearchUrl") ? h("a", { href: text(movie(), "search_url", "SearchUrl"), class: "me-2", target: "_blank", rel: "noopener" }, "Suk") : null,
          text(movie(), "jav_url", "JavUrl") ? h("a", { href: text(movie(), "jav_url", "JavUrl"), class: "me-2", target: "_blank", rel: "noopener" }, "Lib") : null,
          h("span", { class: "ms-auto text-muted" }, text(movie(), "releasing_date", "ReleasingDate"))
        ])
      ].filter(Boolean))
    ]);
  }
});
</script>
