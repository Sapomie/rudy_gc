<template>
  <PageShell class="cast-detail-shell cast-detail-page cards-page-shell" :title="title" :note="note" eyebrow="Cast">
    <div id="cast-detail-root" class="container-fluid">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div class="cast-detail-header d-flex flex-column flex-lg-row justify-content-between align-items-lg-center mb-3 gap-2">
          <div>
            <h4 class="mb-1">{{ actorName }}</h4>
            <div class="text-muted">查看演员概览、最近参与的 SC 历史，以及代表影片。</div>
          </div>
          <div class="d-flex flex-wrap gap-2">
            <button id="btnRebuildActorRank" type="button" class="btn btn-outline-warning" :data-actor-name="actorRawName">回填该演员 Rank</button>
            <a :href="cardsHref" class="btn btn-outline-primary">Cards</a>
          </div>
        </div>

        <div id="castMsgArea" class="alert alert-success small py-2 px-3 mb-3"></div>

        <div class="row g-3 mb-3 align-items-start cast-summary-row">
          <div class="col-12 col-xl-6">
            <section class="cast-section-card">
              <div class="d-flex justify-content-between align-items-center mb-3">
                <h5 class="cast-section-title mb-0">SC 历史</h5>
                <span class="text-muted small">{{ recentEvents.length }} 条</span>
              </div>
              <div v-if="recentEvents.length" class="table-responsive">
                <table class="table table-hover align-middle mb-0">
                  <thead><tr><th>时间</th><th>影片数</th><th>Movies</th><th>Cooldown</th></tr></thead>
                  <tbody>
                    <tr v-for="(event, index) in recentEvents" :key="`${eventHref(event)}-${index}`" :class="{ 'sc-history-extra-row d-none': index >= 5 }">
                      <td><a :href="eventHref(event)">{{ formatUnixMinute(eventTime(event)) }}</a></td>
                      <td>{{ eventMovieCount(event) }}</td>
                      <td>
                        <div v-if="eventMovies(event).length" class="d-flex flex-wrap gap-2">
                          <a
                            v-for="movie in eventMovies(event)"
                            :key="movieHref(movie) || movieName(movie)"
                            :href="movieHref(movie)"
                            class="badge bg-light text-dark border text-decoration-none px-2 py-2"
                          >
                            {{ movieName(movie) }}<span v-if="movieCome(movie)" class="text-danger ms-1">COME</span>
                          </a>
                        </div>
                        <span v-else class="text-muted">-</span>
                      </td>
                      <td>{{ humanDays(eventCooldown(event)) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div v-else class="sc-empty-state">这个演员还没有关联的 SC 记录。</div>
              <div v-if="recentEvents.length > 5" class="mt-3">
                <button id="btnToggleScHistory" class="btn btn-sm btn-outline-secondary" type="button" aria-expanded="false" data-expand-text="展开全部" data-collapse-text="收起">展开全部</button>
              </div>
            </section>
          </div>

          <div class="col-12 col-xl-6">
            <section class="cast-section-card">
              <div class="d-flex justify-content-between align-items-center mb-3">
                <h5 class="cast-section-title mb-0">概览</h5>
              </div>
              <div class="row g-3">
                <div v-for="item in metricItems" :key="item.label" class="col-md-4">
                  <button
                    v-if="item.merge"
                    class="cast-summary-button"
                    type="button"
                    data-bs-toggle="collapse"
                    data-bs-target="#personMergeCollapse"
                    aria-expanded="false"
                    aria-controls="personMergeCollapse"
                  >
                    <div class="sc-metric-card">
                      <div class="sc-metric-label">{{ item.label }}</div>
                      <div class="sc-metric-value">{{ item.value }}</div>
                    </div>
                  </button>
                  <div v-else class="sc-metric-card">
                    <div class="sc-metric-label">{{ item.label }}</div>
                    <div class="sc-metric-value">{{ item.value }}</div>
                  </div>
                </div>
              </div>
            </section>
          </div>
        </div>

        <div v-if="personID > 0" id="personMergeCollapse" class="collapse mb-3">
          <div class="merge-inline-panel">
            <div class="small text-muted mb-3">当前详情页对应保留 person。搜索需要并入的 person，先预览，再执行合并。</div>
            <div class="input-group mb-3">
              <input id="mergeSearchInput" type="text" class="form-control" placeholder="搜索 person 名称、中文名、alias 或 cast 名">
              <button id="btnMergeSearch" type="button" class="btn btn-outline-primary">搜索候选</button>
            </div>
            <div class="mb-3">
              <div class="small text-muted mb-2">候选 person</div>
              <div id="mergeCandidateList" class="merge-list"><div class="merge-list-empty">输入关键词后开始搜索。</div></div>
            </div>
            <div class="mb-3">
              <div class="small text-muted mb-2">待合并来源</div>
              <div id="mergeSelectedList" class="merge-list"><div class="merge-list-empty">还没有选择来源 person。</div></div>
            </div>
            <div class="d-flex flex-wrap gap-2 mb-3">
              <button id="btnPreviewMerge" type="button" class="btn btn-primary" disabled>预览合并</button>
            </div>
            <div id="mergePreviewArea" class="merge-preview-grid" style="display:none;"></div>
          </div>
        </div>

        <section class="cast-section-card cast-representative-section">
          <div class="d-flex flex-column flex-lg-row justify-content-between align-items-lg-center gap-2 mb-3">
            <div>
              <h5 class="cast-section-title mb-0">代表影片</h5>
              <div class="text-muted small">当前模块与 Cards 同步，支持分页、排序、WOwned 与筛选。</div>
            </div>
            <div class="d-flex align-items-center">
              <nav v-if="pageLinks.length" class="mb-0" aria-label="Page navigation">
                <ul class="pagination pagination-sm mb-0">
                  <li v-for="(item, index) in pageLinks" :key="`${pageLabel(item)}-${index}`" class="page-item" :class="{ active: pageActive(item), disabled: pageDisabled(item) || pageEllipsis(item) }">
                    <span v-if="pageEllipsis(item)" class="page-link">...</span>
                    <a v-else class="page-link" :href="pageDisabled(item) ? undefined : normalizePageHref(pageHref(item))" :tabindex="pageDisabled(item) ? -1 : undefined" :aria-disabled="pageDisabled(item) ? 'true' : undefined">{{ pageLabel(item) }}</a>
                  </li>
                </ul>
              </nav>
            </div>
          </div>

          <div class="d-flex flex-wrap justify-content-start align-items-center gap-2 mb-2">
            <div class="btn-group btn-group-sm" role="group" aria-label="Sort Options">
              <a v-for="item in sortItems" :key="item.key" :href="item.href" class="btn btn-outline-primary" :class="{ 'active disabled': item.active }">{{ item.label }}</a>
            </div>
          </div>

          <div class="d-flex flex-column flex-md-row justify-content-between align-items-md-center gap-2 mb-3">
            <div class="d-flex flex-wrap align-items-center gap-2">
              <div class="d-flex flex-wrap align-items-center gap-3">
                <div class="d-flex flex-wrap align-items-center gap-2">
                  <span class="btn btn-sm btn-outline-dark disabled">WMedia</span>
                  <a v-for="item in ownedItems" :key="item.key" :href="item.href" class="btn btn-sm waves-effect waves-light" :class="[item.className, { 'active disabled': item.active }]" :aria-current="item.active ? 'true' : 'false'">{{ item.label }}</a>
                </div>
                <span class="btn btn-sm btn-primary waves-effect waves-light disabled">Total: {{ total }}</span>
              </div>
              <button v-if="hasFilter" class="btn btn-sm btn-outline-secondary" type="button" data-bs-toggle="collapse" data-bs-target="#movieCardFilterCollapse" aria-expanded="false" aria-controls="movieCardFilterCollapse">筛选</button>
            </div>
          </div>

          <MovieCardFilterForm v-if="hasFilter" :filter="filter" variant="cards-agg" />
          <MovieGrid :cards="movies" />
        </section>

        <div v-if="personID > 0" class="modal fade" id="personMergeConfirmModal" tabindex="-1" aria-labelledby="personMergeConfirmModalLabel" aria-hidden="true">
          <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
              <div class="modal-header">
                <h5 class="modal-title" id="personMergeConfirmModalLabel">执行 Person Merge</h5>
                <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
              </div>
              <div class="modal-body">
                <div class="small text-muted mb-2">这会把来源 person 下的全部 am_cast 迁到当前 person，并删除多余 person。</div>
                <div id="mergeConfirmSummary" class="small"></div>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
                <button type="button" class="btn btn-danger" id="btnExecuteMerge">确认执行</button>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import MovieCardFilterForm from "@/components/MovieCardFilterForm.vue";
import MovieGrid from "@/components/MovieGrid.vue";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord, PageInfo } from "@/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/utils/format";

const sortLabels = [
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

const page = computed(() => asRecord(pick(payload.value, "Page")));
const actor = computed(() => asRecord(pick(page.value, "Actor", "actor")));
const actorName = computed(() => text(actor.value, "Chinese", "chinese") || text(actor.value, "Name", "name") || "");
const actorRawName = computed(() => text(actor.value, "Name", "name"));
const title = computed(() => text(payload.value, "Title") || actorName.value || "Cast");
const note = computed(() => "");
const personID = computed(() => Number(text(actor.value, "Id", "id") || 0));
const total = computed(() => text(payload.value, "total", "Total") || "0");
const movies = computed(() => asArray<AnyRecord>(pick(payload.value, "movies", "Movies")));
const filter = computed(() => asRecord(pick(payload.value, "MovieCardFilter")));
const hasFilter = computed(() => Object.keys(filter.value).length > 0);
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => pageInfo.value.Links || []);
const recentEvents = computed(() => asArray<AnyRecord>(pick(page.value, "RecentEvents", "recentEvents")));
const actorAliases = computed(() => asArray<string>(pick(page.value, "ActorAliases", "actorAliases")));
const sortQuery = computed(() => asRecord(payload.value.sortQuery));
const ownedQuery = computed(() => asRecord(payload.value.ownedQuery));
const cardsHref = computed(() => text(page.value, "CardsHref", "cardsHref") || "/cards");

const metricItems = computed(() => [
  { label: "MOVIES", value: text(actor.value, "MovieNumber", "movieNumber") || "0" },
  { label: "WOWNED", value: text(actor.value, "OwnedWMediaNumber", "ownedWMediaNumber") || "0" },
  { label: "SC TIMES", value: text(actor.value, "ScTimes", "scTimes") || "0" },
  { label: "COME TIMES", value: text(actor.value, "ComeTimes", "comeTimes") || "0" },
  { label: "LAST SC", value: formatUnixDate(text(actor.value, "LastScTime", "lastScTime")) },
  { label: "LAST SC EVENT", value: formatUnixDate(text(actor.value, "LastScEventTime", "lastScEventTime")) },
  { label: "HIGHEST RANK", value: rankText(text(actor.value, "HighestRank", "highestRank")) },
  { label: "CAST ALIAS", value: String(actorAliases.value.length) },
  { label: "PERSON MERGE", value: personID.value > 0 ? `#${personID.value}` : "-", merge: personID.value > 0 }
]);

const sortItems = computed(() => sortLabels.map(([key, label]) => {
  const item = asRecord(sortQuery.value[key]);
  return {
    key,
    label,
    href: normalizeHref(text(item, "Href") || "/cards"),
    active: Boolean(item.Active)
  };
}));

const ownedItems = computed(() => [
  { key: "MediaAll", label: "MAll", className: "btn-primary" },
  { key: "MediaOwned", label: "MOwned", className: "btn-success" },
  { key: "MediaNotOwned", label: "MNotOwned", className: "btn-secondary" }
].map((meta) => {
  const item = asRecord(ownedQuery.value[meta.key]);
  return {
    ...meta,
    href: normalizeHref(text(item, "Href") || "/cards"),
    active: Boolean(item.Active)
  };
}));

function formatUnixDate(value: unknown): string {
  const n = Number(value || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function formatUnixMinute(value: unknown): string {
  const n = Number(value || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")} ${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}

function humanDays(value: unknown): string {
  const n = Number(value || 0);
  return n > 0 ? `${(n / 86400).toFixed(2)} 天` : "-";
}

function rankText(value: unknown): string {
  const n = Number(value || 0);
  return n > 0 ? `#${value}` : "-";
}

function eventHref(event: AnyRecord): string {
  return normalizeHref(text(event, "Href", "href"));
}

function eventTime(event: AnyRecord): string {
  return text(event, "ScTime", "scTime");
}

function eventCooldown(event: AnyRecord): string {
  return text(event, "Cooldown", "cooldown");
}

function eventMovieCount(event: AnyRecord): string {
  return text(event, "MovieCount", "movieCount") || "0";
}

function eventMovies(event: AnyRecord): AnyRecord[] {
  return asArray<AnyRecord>(pick(event, "Movies", "movies"));
}

function movieName(movie: AnyRecord): string {
  return text(movie, "Name", "name");
}

function movieHref(movie: AnyRecord): string {
  return normalizeHref(text(movie, "Href", "href"));
}

function movieCome(movie: AnyRecord): boolean {
  return Boolean(movie.IsCome || movie.isCome);
}

function pageLabel(item: AnyRecord): string {
  return text(item, "Label") || "...";
}

function pageHref(item: AnyRecord): string {
  return text(item, "Href");
}

function pageActive(item: AnyRecord): boolean {
  return Boolean(item.Active);
}

function pageDisabled(item: AnyRecord): boolean {
  return Boolean(item.Disabled);
}

function pageEllipsis(item: AnyRecord): boolean {
  return Boolean(item.Ellipsis);
}

function normalizePageHref(href: unknown): string {
  return normalizeHref(href);
}
</script>
