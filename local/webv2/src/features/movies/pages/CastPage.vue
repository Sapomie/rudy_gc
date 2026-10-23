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
              <a :href="cardsHref" class="button-primary next02-cast-action" target="_blank" rel="noopener noreferrer">
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
              <span
                ><strong>{{ recentEvents.length }}</strong> 条记录</span
              >
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
            <table class="table table-hover align-middle mb-0 sc-events-table next02-cast-history-table data-table">
              <colgroup>
                <col class="next02-cast-history-time" />
                <col class="next02-cast-history-count" />
                <col class="next02-cast-history-movies" />
                <col class="next02-cast-history-cooldown" />
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
                    <a class="sc-event-link" :href="eventHref(event)" target="_blank" rel="noopener noreferrer">
                      <span class="sc-event-time">
                        <span class="sc-event-date">{{ formatUnixDate(eventTime(event)) }}</span>
                        <small class="sc-event-hour">{{ formatUnixHm(eventTime(event)) }}</small>
                      </span>
                    </a>
                  </td>
                  <td class="sc-number sc-primary-value">{{ eventMovieCount(event) }}</td>
                  <td class="sc-content-cell">
                    <div v-if="eventMovies(event).length" class="next02-cast-sc-movies">
                      <a
                        v-for="movie in eventMovies(event)"
                        :key="movieHref(movie) || movieName(movie)"
                        :href="movieHref(movie)"
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        {{ movieName(movie) }}<b v-if="movieCome(movie)" class="next02-cast-sc-come">COME</b>
                      </a>
                    </div>
                    <span v-else class="next02-cast-empty-value">-</span>
                  </td>
                  <td class="sc-number">{{ humanDays(eventCooldown(event)) }}</td>
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
          <form class="next02-cast-merge-search" @submit.prevent="searchMergeCandidates">
            <input
              id="mergeSearchInput"
              v-model="mergeKeyword"
              type="text"
              placeholder="搜索 person 名称、中文名、alias 或 cast 名"
              :disabled="mergeSearching || mergeExecuting"
            />
            <button
              id="btnMergeSearch"
              type="submit"
              class="button-secondary"
              :disabled="mergeSearching || mergeExecuting || !mergeKeyword.trim()"
            >
              {{ mergeSearching ? "搜索中..." : "搜索候选" }}
            </button>
          </form>
          <div v-if="mergeMessage" class="next02-cast-merge-message" :class="{ 'is-ok': mergeMessageOK }" role="status">
            {{ mergeMessage }}
          </div>
          <div class="next02-cast-merge-grid">
            <section>
              <h3>候选 person</h3>
              <div id="mergeCandidateList" class="merge-list">
                <div v-if="mergeSearching" class="merge-list-empty">正在搜索候选 person...</div>
                <div v-else-if="!mergeSearchAttempted" class="merge-list-empty">输入关键词后开始搜索。</div>
                <div v-else-if="!mergeCandidates.length" class="merge-list-empty">没有找到可合并的 person。</div>
                <div
                  v-for="candidate in mergeCandidates"
                  v-else
                  :key="mergeCandidatePersonID(candidate)"
                  class="merge-list-item"
                >
                  <div class="merge-list-item-copy">
                    <strong>{{ mergeCandidateName(candidate) }}</strong>
                    <small>#{{ mergeCandidatePersonID(candidate) }}</small>
                    <small v-if="mergeCandidateAlias(candidate)">Alias：{{ mergeCandidateAlias(candidate) }}</small>
                    <small v-if="mergeCandidateCastNames(candidate)">
                      Cast：{{ mergeCandidateCastNames(candidate) }}
                    </small>
                  </div>
                  <button
                    type="button"
                    class="button-secondary"
                    :disabled="mergeSourceSelected(candidate) || mergeExecuting"
                    @click="addMergeSource(candidate)"
                  >
                    {{ mergeSourceSelected(candidate) ? "已选择" : "添加" }}
                  </button>
                </div>
              </div>
            </section>
            <section>
              <h3>待合并来源</h3>
              <div id="mergeSelectedList" class="merge-list">
                <div v-if="!selectedMergeSources.length" class="merge-list-empty">还没有选择来源 person。</div>
                <div
                  v-for="candidate in selectedMergeSources"
                  v-else
                  :key="mergeCandidatePersonID(candidate)"
                  class="merge-list-item"
                >
                  <div class="merge-list-item-copy">
                    <strong>{{ mergeCandidateName(candidate) }}</strong>
                    <small>#{{ mergeCandidatePersonID(candidate) }}</small>
                    <small v-if="mergeCandidateCastNames(candidate)">
                      Cast：{{ mergeCandidateCastNames(candidate) }}
                    </small>
                  </div>
                  <button
                    type="button"
                    class="button-secondary"
                    :disabled="mergeExecuting"
                    @click="removeMergeSource(candidate)"
                  >
                    移除
                  </button>
                </div>
              </div>
            </section>
          </div>
          <div class="next02-cast-merge-footer">
            <button
              id="btnPreviewMerge"
              type="button"
              class="button-primary"
              :disabled="!canPreviewMerge || mergePreviewing || mergeExecuting"
              @click="previewPersonMerge"
            >
              {{ mergePreviewing ? "预览中..." : "预览合并" }}
            </button>
          </div>
          <div v-if="mergePreview" id="mergePreviewArea" class="merge-preview-grid">
            <div>
              <small>保留 Person</small>
              <strong
                >#{{ mergeCandidatePersonID(mergePreview.keep) }} {{ mergeCandidateName(mergePreview.keep) }}</strong
              >
            </div>
            <div>
              <small>来源 Person</small>
              <strong>{{ mergePreviewSourceText }}</strong>
            </div>
            <div>
              <small>迁移 Cast</small>
              <strong>{{ mergePreview.moveCastNames.join("、") || "-" }}</strong>
            </div>
            <div>
              <small>影响影片</small>
              <strong>{{ mergePreview.affectedMovieCount }}</strong>
            </div>
            <button type="button" class="button-primary" :disabled="mergeExecuting" @click="openMergeConfirm">
              执行合并
            </button>
          </div>
        </section>

        <section class="next02-cast-representative" aria-labelledby="cast-movies-title">
          <div class="next02-cast-section-head next02-cast-movies-head">
            <div>
              <h2 id="cast-movies-title">代表影片</h2>
              <p>关联影片、库存状态、排序、筛选与分页。</p>
            </div>
          </div>

          <div
            class="sticky-toolbar cards-toolbar card-results-toolbar next02-pink-cards-toolbar next02-cast-movie-toolbar photo-results-toolbar"
            aria-label="Cast 影片控制栏"
          >
            <div class="cards-toolbar-sort sort-controls">
              <MovieCardSortStrip :query="sortQuery" :options="sortOptions" />
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
                <CardGridSizeControl />
                <div class="next02-pink-cards-metrics" aria-label="Cast 影片摘要">
                  <span :title="`总数 ${total}`" :aria-label="`总数 ${total}`"
                    ><strong>{{ total }}</strong></span
                  >
                  <span :title="`当前页 ${movies.length}`" :aria-label="`当前页 ${movies.length}`"
                    ><strong>{{ movies.length }}</strong></span
                  >
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

        <Teleport to="body">
          <div
            v-if="mergeConfirmOpen && mergePreview"
            id="personMergeConfirmModal"
            class="modal-backdrop-v2 next02-pink-cards-page next02-cast-merge-overlay"
            role="presentation"
            @click.self="closeMergeConfirm"
          >
            <section
              class="modal-panel-v2 next02-cast-merge-confirm"
              role="dialog"
              aria-modal="true"
              aria-labelledby="personMergeConfirmModalLabel"
            >
              <div class="modal-header-v2">
                <h2 id="personMergeConfirmModalLabel">执行 Person Merge</h2>
                <button
                  type="button"
                  class="modal-close-v2"
                  aria-label="关闭"
                  :disabled="mergeExecuting"
                  @click="closeMergeConfirm"
                >
                  ×
                </button>
              </div>
              <div class="modal-body-v2">
                <p>来源 Person 的全部 Cast 将迁移到当前保留 Person，随后删除来源 Person。</p>
                <dl id="mergeConfirmSummary" class="next02-cast-merge-confirm-summary">
                  <div>
                    <dt>保留</dt>
                    <dd>
                      #{{ mergeCandidatePersonID(mergePreview.keep) }} {{ mergeCandidateName(mergePreview.keep) }}
                    </dd>
                  </div>
                  <div>
                    <dt>删除来源</dt>
                    <dd>{{ mergePreviewSourceText }}</dd>
                  </div>
                  <div>
                    <dt>迁移 Cast</dt>
                    <dd>{{ mergePreview.moveCastNames.join("、") || "-" }}</dd>
                  </div>
                  <div>
                    <dt>影响影片</dt>
                    <dd>{{ mergePreview.affectedMovieCount }}</dd>
                  </div>
                </dl>
                <div v-if="mergeConfirmError" class="modal-message-v2" role="alert">{{ mergeConfirmError }}</div>
              </div>
              <div class="modal-footer-v2">
                <button type="button" class="button-secondary" :disabled="mergeExecuting" @click="closeMergeConfirm">
                  取消
                </button>
                <button
                  id="btnExecuteMerge"
                  type="button"
                  class="button-primary next02-cast-merge-danger"
                  :disabled="mergeExecuting"
                  @click="executePersonMerge"
                >
                  {{ mergeExecuting ? "合并中..." : "确认执行" }}
                </button>
              </div>
            </section>
          </div>
        </Teleport>
      </div>
    </template>
  </PageShell>
</template>

<script setup lang="ts">
import "@/features/movies/styles/index.css";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/features/movies/api";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LinkStrip from "@/shared/components/LinkStrip.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import MovieCardFilterForm from "@/features/movies/components/MovieCardFilterForm.vue";
import MovieGrid from "@/features/movies/components/MovieGrid.vue";
import CardGridSizeControl from "@/shared/components/CardGridSizeControl.vue";
import MovieCardSortStrip from "@/features/movies/components/MovieCardSortStrip.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import type { PageInfo } from "@/shared/types/gc";
import type { MovieRecord } from "@/features/movies/types";
import { asArray, asRecord, normalizeHref, pick, text } from "@/shared/utils/format";

interface PersonMergeCandidateView {
  person: MovieRecord;
  castNames: string[];
}

interface PersonMergePreviewView {
  keep: PersonMergeCandidateView;
  sources: PersonMergeCandidateView[];
  moveCastNames: string[];
  removePersonIds: number[];
  affectedMovieCount: number;
}

const sortOptions = [
  { key: "ByDetailUpdate", label: "Detail更新" },
  { key: "ByReleasing", label: "上映日期" },
  { key: "ByViewer", label: "观看人数" },
  { key: "ByRankDate", label: "上榜日期" },
  { key: "ByDaysInRank", label: "在榜天数" },
  { key: "ByHighestRank", label: "最高排名" },
  { key: "ByBirthTime", label: "下载时间" },
  { key: "ByLastSelectedScTime", label: "最近入选 SC" },
  { key: "ByScTimes", label: "sc次数" },
  { key: "ByComeTimes", label: "co次数" }
];

const route = useRoute();
const payload = ref<MovieRecord>({});
const loading = ref(false);
const error = ref("");
const filterOpen = ref(false);
const historyExpanded = ref(false);
const mergeOpen = ref(false);
const mergeKeyword = ref("");
const mergeCandidates = ref<PersonMergeCandidateView[]>([]);
const selectedMergeSources = ref<PersonMergeCandidateView[]>([]);
const mergeSearchAttempted = ref(false);
const mergeSearching = ref(false);
const mergePreviewing = ref(false);
const mergeExecuting = ref(false);
const mergeMessage = ref("");
const mergeMessageOK = ref(false);
const mergePreview = ref<PersonMergePreviewView | null>(null);
const mergeConfirmOpen = ref(false);
const mergeConfirmError = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  filterOpen.value = false;
  historyExpanded.value = false;
  mergeOpen.value = false;
  resetPersonMergeState();
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
const actorName = computed(
  () => text(actor.value, "Chinese", "chinese") || text(actor.value, "Name", "name") || "Cast"
);
const actorRawName = computed(() => text(actor.value, "Name", "name"));
const title = computed(() => actorName.value || text(payload.value, "Title") || "Cast");
const note = computed(() => "人物概览、SC 历史与代表影片。");
const personID = computed(() => Number(text(actor.value, "Id", "id") || 0));
const total = computed(() => text(payload.value, "total", "Total") || "0");
const movies = computed(() => asArray<MovieRecord>(pick(payload.value, "movies", "Movies")));
const filter = computed(() => asRecord(pick(payload.value, "MovieCardFilter")));
const hasFilter = computed(() => Object.keys(filter.value).length > 0);
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const recentEvents = computed(() => asArray<MovieRecord>(pick(page.value, "RecentEvents", "recentEvents")));
const actorAliases = computed(() => asArray<string>(pick(page.value, "ActorAliases", "actorAliases")));
const actorAliasText = computed(() => {
  const primaryNames = new Set([actorName.value, actorRawName.value].map((item) => item.trim()).filter(Boolean));
  return [...new Set(actorAliases.value.map((item) => String(item || "").trim()))]
    .filter((item) => item && !primaryNames.has(item))
    .join("、");
});
const sortQuery = computed(() => asRecord(pick(payload.value, "sortQuery", "SortQuery")));
const ownedQuery = computed(() => asRecord(pick(payload.value, "ownedQuery", "OwnedQuery")));
const cardsHref = computed(() => normalizeHref(text(page.value, "CardsHref", "cardsHref") || "/listcard"));
const selectedMergeSourceIDs = computed(
  () => new Set(selectedMergeSources.value.map((candidate) => mergeCandidatePersonID(candidate)))
);
const canPreviewMerge = computed(() => personID.value > 0 && selectedMergeSources.value.length > 0);
const mergePreviewSourceText = computed(() => {
  const preview = mergePreview.value;
  if (!preview) return "-";
  const sources = preview.sources.map(
    (candidate) => `#${mergeCandidatePersonID(candidate)} ${mergeCandidateName(candidate)}`
  );
  if (sources.length) return sources.join("、");
  return preview.removePersonIds.map((id) => `#${id}`).join("、") || "-";
});

const metricItems = computed(() => [
  { label: "影片数", value: text(actor.value, "MovieNumber", "movieNumber") || "0" },
  { label: "已拥有", value: text(actor.value, "OwnedWMediaNumber", "ownedWMediaNumber") || "0" },
  { label: "SC 次数", value: text(actor.value, "ScTimes", "scTimes") || "0" },
  { label: "COME 次数", value: text(actor.value, "ComeTimes", "comeTimes") || "0" },
  { label: "最近入选 SC", value: formatUnixDate(text(actor.value, "LastSelectedScTime", "lastSelectedScTime")) },
  {
    label: "最近参与 SC",
    value: formatUnixDate(text(actor.value, "LastParticipatedScTime", "lastParticipatedScTime"))
  },
  { label: "最高排名", value: rankText(text(actor.value, "HighestRank", "highestRank")) },
  { label: "别名", value: String(actorAliases.value.length) },
  { label: "Person Merge", value: personID.value > 0 ? `#${personID.value}` : "-", merge: personID.value > 0 }
]);

const ownedItems = computed(() =>
  [
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
  })
);

function resetPersonMergeState() {
  mergeKeyword.value = "";
  mergeCandidates.value = [];
  selectedMergeSources.value = [];
  mergeSearchAttempted.value = false;
  mergeSearching.value = false;
  mergePreviewing.value = false;
  mergeExecuting.value = false;
  mergeMessage.value = "";
  mergeMessageOK.value = false;
  mergePreview.value = null;
  mergeConfirmOpen.value = false;
  mergeConfirmError.value = "";
}

async function searchMergeCandidates() {
  const keyword = mergeKeyword.value.trim();
  if (!keyword || personID.value <= 0 || mergeSearching.value) return;

  mergeSearching.value = true;
  mergeSearchAttempted.value = true;
  mergeMessage.value = "";
  mergeMessageOK.value = false;
  try {
    const query = new URLSearchParams({ q: keyword });
    const data = asRecord(
      await getJSON(`/api/persons/${encodeURIComponent(String(personID.value))}/merge-candidates?${query.toString()}`)
    );
    mergeCandidates.value = parseMergeCandidates(pick(data, "candidates", "Candidates"));
  } catch (err) {
    mergeCandidates.value = [];
    mergeMessage.value = errorText(err);
  } finally {
    mergeSearching.value = false;
  }
}

function addMergeSource(candidate: PersonMergeCandidateView) {
  const candidateID = mergeCandidatePersonID(candidate);
  if (candidateID <= 0 || candidateID === personID.value || selectedMergeSourceIDs.value.has(candidateID)) return;
  selectedMergeSources.value = [...selectedMergeSources.value, candidate];
  invalidateMergePreview();
}

function removeMergeSource(candidate: PersonMergeCandidateView) {
  const candidateID = mergeCandidatePersonID(candidate);
  selectedMergeSources.value = selectedMergeSources.value.filter(
    (selected) => mergeCandidatePersonID(selected) !== candidateID
  );
  invalidateMergePreview();
}

function mergeSourceSelected(candidate: PersonMergeCandidateView): boolean {
  return selectedMergeSourceIDs.value.has(mergeCandidatePersonID(candidate));
}

function invalidateMergePreview() {
  mergePreview.value = null;
  mergeConfirmOpen.value = false;
  mergeConfirmError.value = "";
  mergeMessage.value = "";
  mergeMessageOK.value = false;
}

async function previewPersonMerge() {
  const sourcePersonIds = selectedMergeSources.value.map((candidate) => mergeCandidatePersonID(candidate));
  if (personID.value <= 0 || !sourcePersonIds.length || mergePreviewing.value) return;

  mergePreviewing.value = true;
  mergeMessage.value = "";
  mergeMessageOK.value = false;
  try {
    const data = asRecord(
      await sendJSON("/api/persons/merge-preview", "POST", {
        keepPersonId: personID.value,
        sourcePersonIds
      })
    );
    const preview = parseMergePreview(pick(data, "preview", "Preview"));
    if (!preview) throw new Error("Person Merge 预览响应无效");
    mergePreview.value = preview;
    mergeMessage.value = "合并预览已生成，请核对后执行。";
    mergeMessageOK.value = true;
    openMergeConfirm();
  } catch (err) {
    mergePreview.value = null;
    mergeMessage.value = errorText(err);
  } finally {
    mergePreviewing.value = false;
  }
}

function openMergeConfirm() {
  if (!mergePreview.value || mergeExecuting.value) return;
  mergeConfirmError.value = "";
  mergeConfirmOpen.value = true;
}

function closeMergeConfirm() {
  if (mergeExecuting.value) return;
  mergeConfirmOpen.value = false;
  mergeConfirmError.value = "";
}

async function executePersonMerge() {
  const preview = mergePreview.value;
  if (!preview || personID.value <= 0 || mergeExecuting.value) return;

  mergeExecuting.value = true;
  mergeConfirmError.value = "";
  try {
    const data = asRecord(
      await sendJSON("/api/persons/merge", "POST", {
        keepPersonId: personID.value,
        sourcePersonIds: preview.removePersonIds
      })
    );
    const result = asRecord(pick(data, "result", "Result"));
    const removedPersonIds = numberList(pick(result, "removedPersonIds", "RemovedPersonIds"));
    const affectedMovieCount = Number(pick(result, "affectedMovieCount", "AffectedMovieCount") || 0);
    mergeConfirmOpen.value = false;
    mergeExecuting.value = false;
    await load();
    mergeOpen.value = true;
    mergeMessage.value = `合并完成：移除 ${removedPersonIds.length} 个 Person，更新 ${affectedMovieCount} 部影片。`;
    mergeMessageOK.value = true;
  } catch (err) {
    mergeConfirmError.value = errorText(err);
  } finally {
    mergeExecuting.value = false;
  }
}

function parseMergeCandidates(value: unknown): PersonMergeCandidateView[] {
  const out: PersonMergeCandidateView[] = [];
  for (const raw of asArray(value)) {
    const candidate = parseMergeCandidate(raw);
    if (candidate) out.push(candidate);
  }
  return out;
}

function parseMergeCandidate(value: unknown): PersonMergeCandidateView | null {
  const row = asRecord(value);
  const person = asRecord(pick(row, "person", "Person"));
  const personIDValue = Number(pick(person, "Id", "id") || 0);
  if (personIDValue <= 0) return null;
  return {
    person,
    castNames: stringList(pick(row, "castNames", "CastNames"))
  };
}

function parseMergePreview(value: unknown): PersonMergePreviewView | null {
  const row = asRecord(value);
  const keep = parseMergeCandidate(pick(row, "keep", "Keep"));
  if (!keep) return null;
  return {
    keep,
    sources: parseMergeCandidates(pick(row, "sources", "Sources")),
    moveCastNames: stringList(pick(row, "moveCastNames", "MoveCastNames")),
    removePersonIds: numberList(pick(row, "removePersonIds", "RemovePersonIds")),
    affectedMovieCount: Number(pick(row, "affectedMovieCount", "AffectedMovieCount") || 0)
  };
}

function mergeCandidatePersonID(candidate: PersonMergeCandidateView): number {
  return Number(pick(candidate.person, "Id", "id") || 0);
}

function mergeCandidateName(candidate: PersonMergeCandidateView): string {
  return (
    text(candidate.person, "Chinese", "chinese") ||
    text(candidate.person, "Name", "name") ||
    `Person #${mergeCandidatePersonID(candidate)}`
  );
}

function mergeCandidateAlias(candidate: PersonMergeCandidateView): string {
  return text(candidate.person, "Alias", "alias");
}

function mergeCandidateCastNames(candidate: PersonMergeCandidateView): string {
  return candidate.castNames.join("、");
}

function stringList(value: unknown): string[] {
  return [
    ...new Set(
      asArray(value)
        .map((item) => String(item || "").trim())
        .filter(Boolean)
    )
  ];
}

function numberList(value: unknown): number[] {
  return [
    ...new Set(
      asArray(value)
        .map((item) => Number(item || 0))
        .filter((item) => item > 0)
    )
  ];
}

function errorText(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

function formatUnixDate(value: unknown): string {
  const n = Number(value || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function formatUnixHm(value: unknown): string {
  const n = Number(value || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}

function humanDays(value: unknown): string {
  const n = Number(value || 0);
  return n > 0 ? `${(n / 86400).toFixed(2)} 天` : "-";
}

function rankText(value: unknown): string {
  const n = Number(value || 0);
  return n > 0 ? `#${value}` : "-";
}

function eventHref(event: MovieRecord): string {
  return normalizeHref(text(event, "Href", "href"));
}

function eventTime(event: MovieRecord): string {
  return text(event, "ScTime", "scTime");
}

function eventCooldown(event: MovieRecord): string {
  return text(event, "Cooldown", "cooldown");
}

function eventMovieCount(event: MovieRecord): string {
  return text(event, "MovieCount", "movieCount") || "0";
}

function eventMovies(event: MovieRecord): MovieRecord[] {
  return asArray<MovieRecord>(pick(event, "Movies", "movies"));
}

function movieName(movie: MovieRecord): string {
  return text(movie, "Name", "name");
}

function movieHref(movie: MovieRecord): string {
  return normalizeHref(text(movie, "Href", "href"));
}

function movieCome(movie: MovieRecord): boolean {
  return Boolean(movie.IsCome || movie.isCome);
}
</script>
