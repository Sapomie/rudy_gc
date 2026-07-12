<template>
  <PageShell
    :class="[
      'movie-agg-page-shell',
      'sc-events-page-shell',
      {
        'next02-sc-event-detail-shell': viewType === 'sc-event-detail',
        'cards-page-shell': viewType === 'sc-event-detail',
        'next02-pink-cards-page': viewType === 'sc-event-detail',
        'next02-pink-cards-page-shell': viewType === 'sc-event-detail'
      }
    ]"
    :title="title"
    :note="pageNote"
    :eyebrow="viewType === 'sc-event-detail' ? 'SC Event' : ''"
  >
    <div
      id="page-root"
      class="movie-agg-page-root"
      :class="{
        'container-fluid sc-events-page-root': viewType !== 'sc-event-detail',
        'next02-sc-pink-theme': viewType === 'sc-events' || viewType === 'sc-event-cards',
        'next02-sc-events-rewrite': viewType === 'sc-events',
        'next02-sc-event-cards-rewrite': viewType === 'sc-event-cards',
        'next02-sc-event-detail-root': viewType === 'sc-event-detail'
      }"
    >
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <template v-if="viewType === 'sc-events'">
          <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell">
            <div class="next02-sc-hero-copy">
              <span class="next02-sc-eyebrow">SC / EVENT STREAM</span>
              <h4>SC Events</h4>
              <p>沿时间线浏览 SC 事件，并快速进入本次影片、Come Movie 与演员信息。</p>
            </div>
            <div class="next02-sc-hero-actions">
              <ScViewSwitch />
              <a href="/triggers/backfill" class="btn btn-outline-secondary" target="_blank" rel="noopener noreferrer">回填任务台</a>
            </div>
          </section>

          <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
            <div class="card-body">
              <div class="next02-sc-panel-head">
                <form class="sc-event-filter-form" method="get" action="/sc-events">
                  <input type="hidden" name="p" value="1">
                  <input v-if="routeQueryValue('sort')" type="hidden" name="sort" :value="routeQueryValue('sort')">
                  <input v-if="routeQueryValue('order')" type="hidden" name="order" :value="routeQueryValue('order')">
                  <label class="sc-event-filter-field">
                    <span>起始日期</span>
                    <input type="date" name="start_date" :value="routeQueryValue('start_date')">
                  </label>
                  <label class="sc-event-filter-field">
                    <span>结束日期</span>
                    <input type="date" name="end_date" :value="routeQueryValue('end_date')">
                  </label>
                  <label class="sc-event-filter-field sc-event-page-size-field">
                    <span>每页</span>
                    <select name="ps" :value="filterPageSize">
                      <option v-for="size in pageSizeOptions" :key="size" :value="size">{{ size }}</option>
                    </select>
                  </label>
                  <button type="submit" class="sc-event-filter-submit">筛选</button>
                  <a href="/sc-events" class="sc-event-filter-clear">清空</a>
                </form>
                <div v-if="items.length" class="next02-sc-hero-metrics sc-event-list-metrics" aria-label="SC 事件摘要">
                  <span><strong>{{ total }}</strong> Events</span>
                  <span><strong>{{ items.length }}</strong> Current Page</span>
                </div>
                <div v-if="items.length" class="next02-sc-movie-pager">
                  <PageLinks :links="pageLinks" />
                </div>
              </div>
              <div class="table-responsive">
                <table class="table table-hover align-middle mb-0 sc-events-table">
                  <colgroup class="sc-events-colgroup">
                    <col class="sc-events-col-time">
                    <col class="sc-events-col-total">
                    <col class="sc-events-col-count">
                    <col class="sc-events-col-movie">
                    <col class="sc-events-col-cooldown">
                    <col class="sc-events-col-duration">
                    <col class="sc-events-col-kind">
                    <col class="sc-events-col-cast">
                    <col class="sc-events-col-vessel">
                    <col class="sc-events-col-fg">
                  </colgroup>
                  <thead>
                    <tr>
                      <th class="sc-sortable-head"><SortLink sort-key="ByTime" label="时间" /></th>
                      <th class="sc-number sc-sortable-head"><SortLink sort-key="ByTotalMovieCnt" label="总片数" /></th>
                      <th class="sc-number sc-sortable-head"><SortLink sort-key="ByMovieCnt" label="有 SC" /></th>
                      <th class="sc-sortable-head"><SortLink sort-key="ByComeMovie" label="Come 片" /></th>
                      <th class="sc-number sc-sortable-head"><SortLink sort-key="ByCooldown" label="Cooldown" /></th>
                      <th class="sc-number sc-sortable-head"><SortLink sort-key="ByDuration" label="时长" /></th>
                      <th class="sc-sortable-head"><SortLink sort-key="ByKind" label="种类" /></th>
                      <th class="sc-sortable-head"><SortLink sort-key="ByMovieCast" label="演员" /></th>
                      <th class="sc-sortable-head"><SortLink sort-key="ByVessel" label="Vessel" /></th>
                      <th class="sc-sortable-head"><SortLink sort-key="ByFg" label="FG" /></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(item, index) in items" :key="eventName(eventOf(item))" :class="eventRunningRowClass(item, index)">
                      <td>
                        <a class="sc-event-link" :href="eventDetailHref(eventName(eventOf(item)))" :title="eventName(eventOf(item))" target="_blank" rel="noopener noreferrer">
                          <span class="sc-event-time"><span class="sc-event-date">{{ formatUnix(pick(eventOf(item), "ScTime")) }}</span><small class="sc-event-hour">{{ formatUnixHm(pick(eventOf(item), "ScTime")) }}</small></span>
                        </a>
                      </td>
                      <td class="sc-number sc-primary-value">{{ text(item, "TotalMovieCount") }}</td>
                      <td class="sc-number">{{ text(item, "ScMovieCount") }}</td>
                      <td class="sc-content-cell"><LinkOrMuted :href="movieHref(text(eventOf(item), 'ComeMovieName'))" :value="text(eventOf(item), 'ComeMovieName')" new-window /></td>
                      <td class="sc-number"><span v-if="num(eventOf(item), 'Cooldown') > 0">{{ humanDays(pick(eventOf(item), "Cooldown")) }}</span><span v-else class="text-muted">-</span></td>
                      <td class="sc-number sc-muted"><span v-if="num(eventOf(item), 'DurationMinutes') > 0">{{ text(eventOf(item), "DurationMinutes") }}</span><span v-else class="text-muted">-</span></td>
                      <td class="sc-muted"><span v-if="text(eventOf(item), 'Kind')">{{ text(eventOf(item), "Kind") }}</span><span v-else class="text-muted">-</span></td>
                      <td class="sc-muted sc-content-cell"><LinkOrMuted :href="text(item, 'MovieCastHref')" :value="text(eventOf(item), 'MovieCast')" new-window /></td>
                      <td class="sc-muted"><span v-if="text(eventOf(item), 'Vessel')">{{ text(eventOf(item), "Vessel") }}</span><span v-else class="text-muted">-</span></td>
                      <td class="sc-muted sc-content-cell"><span v-if="text(eventOf(item), 'Fg')">{{ text(eventOf(item), "Fg") }}</span><span v-else class="text-muted">-</span></td>
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
          <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell next02-sc-cards-hero">
            <div class="next02-sc-hero-copy">
              <span class="next02-sc-eyebrow">SC / EVENT STREAM</span>
              <h4>SC Event Cards</h4>
              <p>以封面卡片浏览 SC 事件，并快速查看 Come Movie、数量与事件元数据。</p>
            </div>
            <div class="next02-sc-hero-actions">
              <ScViewSwitch />
              <a href="/triggers/backfill" class="btn btn-outline-secondary" target="_blank" rel="noopener noreferrer">回填任务台</a>
            </div>
          </section>

          <section class="next02-sc-card-controls" aria-label="SC 事件卡片控制栏">
            <div class="sc-event-sort-bar" aria-label="排序方式">
              <SortLink sort-key="ByTime" label="时间" button />
              <SortLink sort-key="ByMovieCnt" label="片数" button />
              <SortLink sort-key="ByComeMovie" label="Come 片" button />
              <SortLink sort-key="ByCooldown" label="Cooldown" button />
              <SortLink sort-key="ByKind" label="种类" button />
              <SortLink sort-key="ByMovieCast" label="演员" button />
              <SortLink sort-key="ByVessel" label="Vessel" button />
              <SortLink sort-key="ByFg" label="FG" button />
            </div>
            <div v-if="items.length" class="next02-sc-card-control-status">
              <div class="next02-sc-hero-metrics sc-event-card-metrics" aria-label="SC 事件卡片摘要">
                <span><strong>{{ total }}</strong> Events</span>
                <span><strong>{{ items.length }}</strong> Current Page</span>
              </div>
              <div class="next02-sc-movie-pager">
                <PageLinks :links="pageLinks" />
              </div>
            </div>
          </section>

          <div v-if="items.length" class="sc-event-card-grid">
            <div v-for="(item, index) in items" :key="eventName(eventOf(item))" class="sc-event-card-col">
              <article class="card h-100 sc-event-grid-card" :class="eventCardClass(item, index)">
                <a
                  :href="text(item, 'DetailHref') || eventDetailHref(eventName(eventOf(item)))"
                  class="sc-event-cover-link"
                  :title="eventName(eventOf(item))"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <div class="thumb-2x3 sc-event-cover">
                    <img v-if="item.HasComeMovieCover" :src="text(item, 'ComeMovieJacketImg')" :alt="`${eventName(eventOf(item))} 封面`" loading="lazy" decoding="async">
                    <div v-else class="sc-event-cover-empty">
                      <span class="sc-event-cover-kicker">SC EVENT</span>
                      <strong><span class="sc-event-time"><span class="sc-event-date">{{ formatUnix(pick(eventOf(item), "ScTime")) }}</span><span class="sc-event-hour">{{ formatUnixHm(pick(eventOf(item), "ScTime")) }}</span></span></strong>
                      <span>{{ eventName(eventOf(item)) }}</span>
                    </div>
                    <div class="sc-event-cover-badge">{{ text(cardListItemOf(item), "ScMovieCount") || text(eventOf(item), "MovieNumber") || "0" }} SC</div>
                  </div>
                </a>
                <div class="card-body sc-event-card-body">
                  <div class="sc-event-card-heading">
                    <a
                      :href="text(item, 'DetailHref') || eventDetailHref(eventName(eventOf(item)))"
                      class="sc-event-card-date-link"
                      :title="eventName(eventOf(item))"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <span class="sc-event-time">
                        <span class="sc-event-date">{{ formatUnix(pick(eventOf(item), "ScTime")) }}</span>
                        <span class="sc-event-hour">{{ formatUnixHm(pick(eventOf(item), "ScTime")) }}</span>
                      </span>
                    </a>
                  </div>
                  <div class="sc-event-card-primary">
                    <span class="sc-event-card-label">Come Movie</span>
                    <span class="sc-event-card-primary-link" :title="text(item, 'ComeMovieName')">
                      <LinkOrMuted :href="text(item, 'ComeMovieHref')" :value="text(item, 'ComeMovieName')" new-window />
                    </span>
                  </div>
                  <div class="sc-event-card-numbers" aria-label="影片数量">
                    <span><small>总片数</small><strong>{{ text(cardListItemOf(item), "TotalMovieCount") || "0" }}</strong></span>
                    <span><small>SC 片数</small><strong>{{ text(cardListItemOf(item), "ScMovieCount") || text(eventOf(item), "MovieNumber") || "0" }}</strong></span>
                  </div>
                  <div class="sc-event-meta-grid">
                    <div class="sc-event-meta-row">
                      <span class="sc-event-meta-label">Cooldown</span>
                      <span class="sc-event-meta-value sc-event-meta-number">{{ num(eventOf(item), "Cooldown") ? humanDays(pick(eventOf(item), "Cooldown")) : "-" }}</span>
                    </div>
                    <div class="sc-event-meta-row">
                      <span class="sc-event-meta-label">种类</span>
                      <span class="sc-event-meta-value" :title="text(eventOf(item), 'Kind')">{{ text(eventOf(item), "Kind") || "-" }}</span>
                    </div>
                    <div class="sc-event-meta-row">
                      <span class="sc-event-meta-label">演员</span>
                      <span class="sc-event-meta-value" :title="text(eventOf(item), 'MovieCast')">
                        <LinkOrMuted :href="text(cardListItemOf(item), 'MovieCastHref')" :value="text(eventOf(item), 'MovieCast')" new-window />
                      </span>
                    </div>
                    <div class="sc-event-meta-row">
                      <span class="sc-event-meta-label">Vessel</span>
                      <span class="sc-event-meta-value" :title="text(eventOf(item), 'Vessel')">{{ text(eventOf(item), "Vessel") || "-" }}</span>
                    </div>
                    <div class="sc-event-meta-row">
                      <span class="sc-event-meta-label">FG</span>
                      <span class="sc-event-meta-value" :title="text(eventOf(item), 'Fg')">{{ text(eventOf(item), "Fg") || "-" }}</span>
                    </div>
                  </div>
                </div>
              </article>
            </div>
          </div>
          <div v-else class="sc-empty-state sc-event-card-empty">还没有可展示的 SC 事件。</div>
        </template>

        <template v-else-if="viewType === 'sc-event-detail'">
          <div class="next02-sc-detail-actionbar" aria-label="SC Event 操作栏">
            <span><strong>{{ items.length }}</strong> 当前影片</span>
            <div class="next02-sc-detail-actions">
              <button type="button" class="button-primary" id="btnOpenScEventEditModal" data-bs-toggle="modal" data-bs-target="#scEventEditModal">编辑信息</button>
              <a href="/sc-events" class="button-secondary" target="_blank" rel="noopener noreferrer">返回列表</a>
              <a :href="`/triggers/sc-media?sc_name=${encodeURIComponent(eventName(event))}`" class="button-secondary" target="_blank" rel="noopener noreferrer">移动 SC Media</a>
              <a href="/triggers/backfill" class="button-secondary" target="_blank" rel="noopener noreferrer">回填任务台</a>
            </div>
          </div>
          <div id="scEventPageMsg" class="next02-sc-detail-message" role="alert"></div>
          <ScEventEditModal />

          <section class="next02-sc-detail-overview" :class="{ 'without-image': !text(event, 'ImagePath') }" aria-label="SC Event 概览">
            <div v-if="text(event, 'ImagePath')" class="next02-sc-detail-panel next02-sc-detail-image-panel">
              <div class="next02-sc-detail-panel-head">
                <h2>SC Image</h2>
              </div>
              <div class="next02-sc-detail-image-wrap">
                <img :src="text(event, 'ImagePath')" :alt="`${eventName(event)} 图片`" class="next02-sc-detail-image">
              </div>
            </div>
            <div class="next02-sc-detail-panel next02-sc-detail-info-panel">
              <div class="next02-sc-detail-panel-head">
                <h2>事件信息</h2>
              </div>
              <div class="next02-sc-detail-info-grid">
                <MetricCard label="Sc Time" :value="formatUnixFull(pick(event, 'ScTime'))" />
                <MetricCard label="Movie Count" :value="text(event, 'MovieNumber') || '0'" />
                <MetricCard label="Cooldown" :value="num(event, 'Cooldown') > 0 ? humanDays(pick(event, 'Cooldown')) : '-'" />
                <MetricCard label="Come Movie" :value="text(event, 'ComeMovieName') || '-'" :href="movieHref(text(event, 'ComeMovieName'))" />
                <MetricCard label="种类" :value="text(event, 'Kind') || '-'" />
                <MetricCard label="Lead Cast" :value="text(event, 'MovieCast') || '-'" :href="text(event, 'MovieCast') ? `/listcard?cn=${encodeURIComponent(text(event, 'MovieCast'))}` : ''" />
                <MetricCard label="时长(分钟)" :value="num(event, 'DurationMinutes') > 0 ? `${text(event, 'DurationMinutes')} 分钟` : '-'" />
                <MetricCard label="FG" :value="text(event, 'Fg') || '-'" />
                <MetricCard label="Vessel" :value="text(event, 'Vessel') || '-'" />
                <div class="sc-detail-info-item sc-metric-card sc-detail-remarks">
                  <div class="sc-metric-label">Remarks</div>
                  <div class="sc-metric-value"><span v-if="text(event, 'Remarks')">{{ text(event, "Remarks") }}</span><span v-else class="text-muted">暂无备注。</span></div>
                </div>
              </div>
            </div>
          </section>

          <section class="next02-sc-detail-movies" aria-labelledby="sc-detail-movies-title">
            <div class="next02-sc-detail-section-head">
              <div>
                <h2 id="sc-detail-movies-title">Movies</h2>
                <p>本次事件的影片及 SC 状态。</p>
              </div>
            </div>

            <div class="next02-sc-detail-movie-toolbar" aria-label="电影筛选">
              <div class="next02-sc-detail-filter-strip">
                <a :href="text(payload, 'MovieFilterAllHref')" :class="{ active: movieFilter === 'all' }">全部电影</a>
                <a :href="text(payload, 'MovieFilterScHref')" :class="{ active: movieFilter === 'sc' }">有 SC</a>
                <a :href="text(payload, 'MovieFilterNoScHref')" :class="{ active: movieFilter === 'nosc' }">无 SC</a>
              </div>
              <div class="next02-pink-cards-metrics" aria-label="电影摘要">
                <span><strong>{{ items.length }}</strong> Current Page</span>
                <span><strong>{{ text(event, 'MovieNumber') || '0' }}</strong> Total</span>
              </div>
            </div>

            <section v-if="items.length" aria-label="SC Movies">
              <div class="movie-grid-v2 next02-pink-movie-grid next02-sc-detail-movie-grid" id="movie-grid">
                <div v-for="(item, index) in items" :key="movieKey(item, index)" class="col movie-grid-col-v2">
                  <ScMovieCard :item="item" :event-name="eventName(event)" />
                </div>
              </div>
            </section>
            <div v-else class="movie-empty next02-sc-detail-empty">{{ emptyDetailText }}</div>
          </section>

          <section v-if="failedItems.length" class="next02-sc-detail-failed" aria-labelledby="sc-detail-failed-title">
            <div class="next02-sc-detail-section-head">
              <div>
                <h2 id="sc-detail-failed-title">Skipped / Error Items</h2>
                <p>未能载入的影片条目。</p>
              </div>
            </div>
            <div class="next02-sc-detail-failed-table-wrap">
              <table class="next02-sc-detail-failed-table">
                <colgroup><col class="failed-movie"><col class="failed-javid"><col class="failed-error"></colgroup>
                <thead><tr><th>Movie</th><th>JavId</th><th>Error</th></tr></thead>
                <tbody>
                  <tr v-for="(item, index) in failedItems" :key="index">
                    <td>{{ text(item, "MovieName") }}</td>
                    <td>{{ text(asRecord(item.Entry), "MovieJavId") }}</td>
                    <td>{{ text(item, "LoadError") }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
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
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, normalizeHref, num, pick, text } from "@/utils/format";

const GLIST_IS_NOT_SC = 1;
const GLIST_IS_SC = 2;

const route = useRoute();
const payload = ref<AnyRecord>({});
const cardListItems = ref<AnyRecord[]>([]);
const loading = ref(false);
const error = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  cardListItems.value = [];
  try {
    if (route.path === "/sc-events-cards") {
      const [pageData, listData] = await Promise.all([
        getJSON(apiPathForRoute(route)),
        getJSON(cardListSummaryApiPath()).catch(() => ({})),
      ]);
      payload.value = asRecord(pageData);
      cardListItems.value = asArray<AnyRecord>(pick(asRecord(listData), "Items"));
    } else {
      payload.value = asRecord(await getJSON(apiPathForRoute(route)));
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title", "PageTitle") || "SC Events");
const pageNote = computed(() => text(payload.value, "PageNote") || (viewType.value === "sc-event-detail" ? "基础信息、影片清单与快速操作入口。" : ""));
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
const cardListItemsByEventName = computed(() => {
  const index = new Map<string, AnyRecord>();
  for (const item of cardListItems.value) {
    const name = eventName(eventOf(item));
    if (name) index.set(name, item);
  }
  return index;
});
const pageSizeOptions = [24, 50, 100, 200];
const filterPageSize = computed(() => {
  const value = Number(routeQueryValue("ps"));
  return pageSizeOptions.includes(value) ? value : 24;
});
const emptyDetailText = computed(() => {
  if (movieFilter.value === "all") return "本次 SC 还没有电影明细。";
  if (movieFilter.value === "nosc") return "当前筛选下没有 is_sc=1 的电影。";
  return "当前筛选下没有 is_sc=2 的电影。";
});

function eventOf(item: AnyRecord): AnyRecord {
  return asRecord(pick(item, "Event"));
}

function cardListSummaryApiPath(): string {
  const query = route.fullPath.split("?", 2)[1] || "";
  const params = new URLSearchParams(query);
  if (!params.has("ps")) params.set("ps", "18");
  const encoded = params.toString();
  return `/api/sc-events${encoded ? `?${encoded}` : ""}`;
}

function cardListItemOf(item: AnyRecord): AnyRecord {
  return cardListItemsByEventName.value.get(eventName(eventOf(item))) || {};
}

function routeQueryValue(key: string): string {
  const value = route.query[key];
  if (Array.isArray(value)) return value[0] ? String(value[0]) : "";
  return value == null ? "" : String(value);
}

function eventName(row: AnyRecord): string {
  return text(row, "Name", "name");
}

function eventDetailHref(name: string): string {
  return `/sc-events/${encodeURI(name || "")}`;
}

function movieHref(name: string): string {
  return name ? `/moviedetail/${encodeURI(name)}` : "";
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

function eventMonthKey(input: unknown): string {
  const n = Number(input || 0);
  if (n <= 0) return "";
  const date = new Date(n * 1000);
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`;
}

function eventRunningMonthClass(input: unknown): string {
  const month = eventMonthKey(input);
  if (!month) return "sc-running-month-neutral";
  const monthNumber = Number(month.slice(5, 7));
  return `sc-running-month-${(monthNumber % 3) + 1}`;
}

function eventRunningTableMonthClass(input: unknown): string {
  const month = eventMonthKey(input);
  if (!month) return "sc-running-month-neutral";
  const monthNumber = Number(month.slice(5, 7));
  return monthNumber % 2 === 0 ? "sc-running-month-light" : "sc-running-month-deep";
}

function isEventRunningMonthStart(item: AnyRecord, index: number): boolean {
  if (index <= 0) return false;
  const currentMonth = eventMonthKey(pick(eventOf(item), "ScTime"));
  const previousMonth = eventMonthKey(pick(eventOf(items.value[index - 1]), "ScTime"));
  return Boolean(currentMonth && previousMonth && currentMonth !== previousMonth);
}

function eventRunningRowClass(item: AnyRecord, index: number) {
  return [
    "next02-sc-event-row",
    eventRunningTableMonthClass(pick(eventOf(item), "ScTime")),
    { "sc-running-month-start": isEventRunningMonthStart(item, index) },
  ];
}

function eventCardClass(item: AnyRecord, index: number) {
  return [
    eventRunningMonthClass(pick(eventOf(item), "ScTime")),
    { "sc-running-month-start": isEventRunningMonthStart(item, index) },
  ];
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
    value: { type: String, default: "" },
    newWindow: { type: Boolean, default: false }
  },
  setup(props) {
    return () => props.value
      ? (props.href ? h("a", {
          href: props.href,
          target: props.newWindow ? "_blank" : undefined,
          rel: props.newWindow ? "noopener noreferrer" : undefined
        }, props.value) : h("span", props.value))
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
        class: ["btn", "btn-outline-primary", { "active disabled": Boolean(scViewQuery.value.IsList) }],
        target: "_blank",
        rel: "noopener noreferrer"
      }, "列表视图"),
      h("a", {
        href: normalizeHref(text(scViewQuery.value, "CardHref"), "/sc-events-cards"),
        class: ["btn", "btn-outline-primary", { "active disabled": Boolean(scViewQuery.value.IsCard) }],
        target: "_blank",
        rel: "noopener noreferrer"
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
      h("div", { class: "sc-metric-value" }, props.href && props.value !== "-" ? [h("a", { href: props.href, target: "_blank", rel: "noopener noreferrer" }, props.value)] : props.value)
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
      return h("li", { class: ["list-group-item", "movie-card-sc-status", isSc ? "has-sc-status" : "no-sc-status"], style: "z-index:2;" }, [
        h("div", { class: "movie-card-sc-status-line" }, [
          h("div", { class: "movie-card-sc-badges" }, [
            h("span", { class: ["movie-card-sc-state", isSc ? "has-sc" : "no-sc"] }, isSc ? "有 SC" : "无 SC"),
            isCome ? h("span", { class: "movie-card-sc-come" }, "Come") : null
          ]),
          h("div", { class: "movie-card-sc-actions", role: "group", "aria-label": "编辑影片 SC 状态" }, [
            h("button", { type: "button", class: ["js-sc-event-movie-status", "has-sc-action", { active: isSc }], "data-url": apiURL, "data-is-sc": GLIST_IS_SC, disabled: isSc || undefined }, "有 SC"),
            h("button", { type: "button", class: ["js-sc-event-movie-status", "no-sc-action", { active: !isSc }], "data-url": apiURL, "data-is-sc": GLIST_IS_NOT_SC, disabled: !isSc || undefined }, "无 SC")
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
        ? h("span", { class: "btn btn-sm btn-secondary disabled", "aria-disabled": "true" }, "Removed")
        : h("a", { href: videoUrl, class: ["btn", "btn-sm", ownedMedia === 5 ? "btn-info" : "btn-success"], target: "_blank", rel: "noopener noreferrer" }, "Play");
      return h("li", { class: "list-group-item py-2 movie-card-media-block" }, [
        h("div", { class: "card-play-stack" }, [h("div", { class: "card-play-row" }, [
          play,
          mediaBirth ? h("span", { class: "text-muted small card-play-date movie-card-download-date" }, `下载 ${mediaBirth}`) : null
        ])])
      ]);
    };
    const genreLinks = () => asArray<string>(pick(movie(), "genre", "Genre")).map((genre) => h("a", { href: `/listcard?gn=${encodeURIComponent(genre)}`, class: "badge badge-genre movie-card-tag-genre me-1 mb-1", target: "_blank", rel: "noopener noreferrer" }, String(genre)));
    const castLinks = () => asArray<AnyRecord>(pick(movie(), "cast", "Cast")).map((cast) => {
      const label = text(cast, "NameShow", "DisplayName", "Name", "name");
      const personID = text(cast, "PersonId", "person_id", "Id", "id");
      return h("a", { href: personID && personID !== "0" ? `/castdetail/${encodeURIComponent(personID)}` : `/listcard?cn=${encodeURIComponent(label)}`, class: "person-tag movie-card-tag-person movie-card-tag-cast mr-2", title: label, target: "_blank", rel: "noopener noreferrer" }, label);
    });
    return () => h("article", { class: "card shadow-sm h-100 card-lift movie-card-v2 next02-pink-movie-card" }, [
      h("a", { href: movieHref(movieName()), class: "d-block poster-link", "aria-label": `查看 ${movieName()} 详情`, target: "_blank", rel: "noopener noreferrer" }, [
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
        h("li", { class: "list-group-item movie-card-title-block" }, [
          h("h6", { class: "mb-1 text-truncate", title: movieTitle() }, [
            h("a", { href: `/listcard?pn=${encodeURIComponent(text(movie(), "prefix", "Prefix"))}`, class: "stretched-link text-reset text-decoration-none movie-name", target: "_blank", rel: "noopener noreferrer" }, movieName() || "-"),
            text(movie(), "viewers_number_watched", "ViewersNumberWatched") ? h("small", { class: "text-muted ml-2" }, `(${text(movie(), "viewers_number_watched", "ViewersNumberWatched")})`) : null
          ]),
          movieTitle() ? h("p", { class: "mb-0 text-muted line-clamp-2 movie-subtitle", title: movieTitle() }, movieTitle()) : null
        ]),
        moviePlayRow(),
        h("li", { class: "list-group-item text-muted genre-list movie-card-tag-block" }, genreLinks()),
        h("li", { class: "list-group-item cast-list movie-card-tag-block" }, [
          ...castLinks(),
          text(movie(), "director", "Director") ? h("a", { href: `/listcard?dn=${encodeURIComponent(text(movie(), "director", "Director"))}`, class: "person-tag movie-card-tag-person movie-card-tag-director ml-2", title: text(movie(), "director", "Director"), target: "_blank", rel: "noopener noreferrer" }, text(movie(), "director", "Director")) : null
        ]),
        h("li", { class: "list-group-item d-flex align-items-center small card-link-row movie-card-footer" }, [
          text(movie(), "bus_url", "BusUrl") ? h("a", { href: text(movie(), "bus_url", "BusUrl"), class: "movie-card-tag-link me-2", target: "_blank", rel: "noopener noreferrer" }, "Bus") : null,
          text(movie(), "search_url", "SearchUrl") ? h("a", { href: text(movie(), "search_url", "SearchUrl"), class: "movie-card-tag-link me-2", target: "_blank", rel: "noopener noreferrer" }, "Suk") : null,
          text(movie(), "jav_url", "JavUrl") ? h("a", { href: text(movie(), "jav_url", "JavUrl"), class: "movie-card-tag-link me-2", target: "_blank", rel: "noopener noreferrer" }, "Lib") : null,
          h("span", { class: "movie-card-tag-date ms-auto text-muted" }, `上映 ${text(movie(), "releasing_date", "ReleasingDate")}`)
        ])
      ].filter(Boolean))
    ]);
  }
});
</script>
