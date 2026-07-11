<template>
  <PageShell
    class="next02-cast-detail-shell next02-cast-page cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell"
    :title="title"
    :note="note"
    eyebrow="Cast"
  >
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <template v-else>
      <div id="cast-detail-root" class="next02-cast-detail-root">
        <section class="next02-cast-overview" aria-label="人物概览">
          <div class="next02-cast-overview-toolbar">
            <div class="next02-cast-identity">
              <span>人物档案</span>
              <strong v-if="actorRawName && actorRawName !== actorName">{{ actorRawName }}</strong>
              <small v-if="actorAliasText">别名：{{ actorAliasText }}</small>
            </div>
            <div class="next02-cast-actions">
              <button
                id="btnRebuildActorRank"
                type="button"
                class="button-secondary next02-cast-action"
                :data-actor-name="actorRawName"
              >
                回填该演员 Rank
              </button>
              <a
                :href="cardsHref"
                class="button-primary next02-cast-action"
                target="_blank"
                rel="noopener noreferrer"
              >
                Cards
              </a>
            </div>
          </div>

          <div class="next02-cast-metrics" aria-label="人物统计">
            <template v-for="item in metricItems" :key="item.label">
              <button
                v-if="item.merge"
                type="button"
                class="next02-cast-metric next02-cast-merge-trigger"
                :class="{ active: mergeOpen }"
                :aria-expanded="mergeOpen"
                aria-controls="personMergeCollapse"
                @click="mergeOpen = !mergeOpen"
              >
                <small>{{ item.label }}</small>
                <strong>{{ item.value }}</strong>
              </button>
              <span v-else class="next02-cast-metric">
                <small>{{ item.label }}</small>
                <strong>{{ item.value }}</strong>
              </span>
            </template>
          </div>
        </section>

        <div id="castMsgArea" class="next02-cast-message" aria-live="polite"></div>

        <section class="next02-cast-section next02-cast-history" aria-labelledby="cast-history-title">
          <div class="next02-cast-section-head">
            <div>
              <h2 id="cast-history-title">SC 历史</h2>
              <p>最近参与的 SC 事件及关联影片。</p>
            </div>
            <div class="next02-cast-section-status">
              <span><strong>{{ recentEvents.length }}</strong> 条记录</span>
              <button
                v-if="recentEvents.length > 5"
                id="btnToggleScHistory"
                type="button"
                class="button-secondary next02-cast-history-toggle"
                :aria-expanded="historyExpanded"
                @click="historyExpanded = !historyExpanded"
              >
                {{ historyExpanded ? "收起" : "展开全部" }}
              </button>
            </div>
          </div>

          <div v-if="recentEvents.length" class="next02-cast-history-table-wrap">
            <table class="next02-cast-history-table">
              <colgroup>
                <col class="next02-cast-history-time">
                <col class="next02-cast-history-count">
                <col class="next02-cast-history-movies">
                <col class="next02-cast-history-cooldown">
              </colgroup>
              <thead>
                <tr>
                  <th>时间</th>
                  <th>影片数</th>
                  <th>Movies</th>
                  <th>Cooldown</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(event, index) in recentEvents"
                  v-show="historyExpanded || index < 5"
                  :key="`${eventHref(event)}-${index}`"
                >
                  <td>
                    <a :href="eventHref(event)" target="_blank" rel="noopener noreferrer">
                      {{ formatUnixMinute(eventTime(event)) }}
                    </a>
                  </td>
                  <td>{{ eventMovieCount(event) }}</td>
                  <td>
                    <div v-if="eventMovies(event).length" class="next02-cast-sc-movies">
                      <a
                        v-for="movie in eventMovies(event)"
                        :key="movieHref(movie) || movieName(movie)"
                        :href="movieHref(movie)"
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        {{ movieName(movie) }}<b v-if="movieCome(movie)">COME</b>
                      </a>
                    </div>
                    <span v-else class="next02-cast-empty-value">-</span>
                  </td>
                  <td>{{ humanDays(eventCooldown(event)) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else class="movie-empty next02-cast-history-empty">这个演员还没有关联的 SC 记录。</div>
        </section>

        <section
          v-if="personID > 0"
          v-show="mergeOpen"
          id="personMergeCollapse"
          class="next02-cast-merge-panel"
          aria-label="Person Merge"
        >
          <div class="next02-cast-merge-head">
            <div>
              <h2>Person Merge</h2>
              <p>当前人物作为保留项；搜索来源人物，预览后执行合并。</p>
            </div>
            <button type="button" class="button-secondary" @click="mergeOpen = false">收起</button>
          </div>
          <div class="next02-cast-merge-search">
            <input id="mergeSearchInput" type="text" placeholder="搜索 person 名称、中文名、alias 或 cast 名">
            <button id="btnMergeSearch" type="button" class="button-secondary">搜索候选</button>
          </div>
          <div class="next02-cast-merge-grid">
            <section>
              <h3>候选 person</h3>
              <div id="mergeCandidateList" class="merge-list"><div class="merge-list-empty">输入关键词后开始搜索。</div></div>
            </section>
            <section>
              <h3>待合并来源</h3>
              <div id="mergeSelectedList" class="merge-list"><div class="merge-list-empty">还没有选择来源 person。</div></div>
            </section>
          </div>
          <div class="next02-cast-merge-footer">
            <button id="btnPreviewMerge" type="button" class="button-primary" disabled>预览合并</button>
          </div>
          <div id="mergePreviewArea" class="merge-preview-grid" style="display:none;"></div>
        </section>

        <section class="next02-cast-representative" aria-labelledby="cast-movies-title">
          <div class="next02-cast-section-head next02-cast-movies-head">
            <div>
              <h2 id="cast-movies-title">代表影片</h2>
              <p>关联影片、库存状态、排序、筛选与分页。</p>
            </div>
          </div>

          <div class="sticky-toolbar cards-toolbar next02-pink-cards-toolbar next02-cast-movie-toolbar" aria-label="Cast 影片控制栏">
            <div class="cards-toolbar-sort">
              <LinkStrip :items="sortItems" />
            </div>
            <div class="toolbar-line cards-toolbar-bottom">
              <div class="owned-line cards-toolbar-filters">
                <LinkStrip :items="ownedItems" />
                <button
                  v-if="hasFilter"
                  type="button"
                  class="button-secondary cards-filter-button"
                  :class="{ active: filterOpen }"
                  :aria-pressed="filterOpen"
                  @click="filterOpen = !filterOpen"
                >
                  筛选
                </button>
              </div>
              <div class="next02-pink-cards-toolbar-status">
                <div class="next02-pink-cards-metrics" aria-label="Cast 影片摘要">
                  <span><strong>{{ total }}</strong> Total</span>
                  <span><strong>{{ movies.length }}</strong> Current Page</span>
                </div>
                <div class="cards-toolbar-pagination next02-pink-cards-pagination">
                  <PaginationBar :page-info="pageInfo" />
                </div>
              </div>
            </div>
          </div>

          <div v-show="filterOpen" class="next02-pink-cards-filter next02-cast-filter">
            <MovieCardFilterForm v-if="hasFilter" :filter="filter" />
          </div>

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
      </div>
    </template>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LinkStrip from "@/components/LinkStrip.vue";
import LoadingState from "@/components/LoadingState.vue";
import MovieCardFilterForm from "@/components/MovieCardFilterForm.vue";
import MovieGrid from "@/components/MovieGrid.vue";
import PageShell from "@/components/PageShell.vue";
import PaginationBar from "@/components/PaginationBar.vue";
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
  ["ByLastScTime", "最后sc"],
  ["ByScTimes", "sc次数"],
  ["ByComeTimes", "co次数"]
];

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const filterOpen = ref(false);
const historyExpanded = ref(false);
const mergeOpen = ref(false);

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  filterOpen.value = false;
  historyExpanded.value = false;
  mergeOpen.value = false;
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
const actorName = computed(() => text(actor.value, "Chinese", "chinese") || text(actor.value, "Name", "name") || "Cast");
const actorRawName = computed(() => text(actor.value, "Name", "name"));
const title = computed(() => actorName.value || text(payload.value, "Title") || "Cast");
const note = computed(() => "人物概览、SC 历史与代表影片。");
const personID = computed(() => Number(text(actor.value, "Id", "id") || 0));
const total = computed(() => text(payload.value, "total", "Total") || "0");
const movies = computed(() => asArray<AnyRecord>(pick(payload.value, "movies", "Movies")));
const filter = computed(() => asRecord(pick(payload.value, "MovieCardFilter")));
const hasFilter = computed(() => Object.keys(filter.value).length > 0);
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const recentEvents = computed(() => asArray<AnyRecord>(pick(page.value, "RecentEvents", "recentEvents")));
const actorAliases = computed(() => asArray<string>(pick(page.value, "ActorAliases", "actorAliases")));
const actorAliasText = computed(() => {
  const primaryNames = new Set([actorName.value, actorRawName.value].map((item) => item.trim()).filter(Boolean));
  return [...new Set(actorAliases.value.map((item) => String(item || "").trim()))]
    .filter((item) => item && !primaryNames.has(item))
    .join("、");
});
const sortQuery = computed(() => asRecord(pick(payload.value, "sortQuery", "SortQuery")));
const ownedQuery = computed(() => asRecord(pick(payload.value, "ownedQuery", "OwnedQuery")));
const cardsHref = computed(() => normalizeHref(text(page.value, "CardsHref", "cardsHref") || "/cards"));

const metricItems = computed(() => [
  { label: "影片数", value: text(actor.value, "MovieNumber", "movieNumber") || "0" },
  { label: "已拥有", value: text(actor.value, "OwnedWMediaNumber", "ownedWMediaNumber") || "0" },
  { label: "SC 次数", value: text(actor.value, "ScTimes", "scTimes") || "0" },
  { label: "COME 次数", value: text(actor.value, "ComeTimes", "comeTimes") || "0" },
  { label: "最后 SC", value: formatUnixDate(text(actor.value, "LastScTime", "lastScTime")) },
  { label: "最后 SC 事件", value: formatUnixDate(text(actor.value, "LastScEventTime", "lastScEventTime")) },
  { label: "最高排名", value: rankText(text(actor.value, "HighestRank", "highestRank")) },
  { label: "别名", value: String(actorAliases.value.length) },
  { label: "Person Merge", value: personID.value > 0 ? `#${personID.value}` : "-", merge: personID.value > 0 }
]);

const sortItems = computed(() => sortLabels.map(([key, label]) => {
  const item = asRecord(sortQuery.value[key]);
  return {
    label,
    href: normalizeHref(text(item, "Href") || route.fullPath),
    active: Boolean(item.Active)
  };
}));

const ownedItems = computed(() => [
  ["MediaAll", "All"],
  ["MediaOwned", "Owned"],
  ["MediaNotOwned", "NotOwned"]
].map(([key, label]) => {
  const item = asRecord(ownedQuery.value[key]);
  return {
    label,
    href: normalizeHref(text(item, "Href") || route.fullPath),
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
</script>
