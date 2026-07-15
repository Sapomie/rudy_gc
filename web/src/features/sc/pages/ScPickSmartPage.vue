<template>
  <PageShell class="sc-events-page-shell sc-pick-page-shell cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell" :title="title" eyebrow="">
    <div id="page-root" class="container-fluid sc-pick-page-root next02-sc-pink-theme next02-smart-pick-page-rewrite">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <section class="next02-sc-movie-top-shell next02-smart-pick-hero">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">SC / SMART PICK LAB</span>
            <h4>{{ pageTitle }}</h4>
            <p>{{ effectivePageNote }}</p>
          </div>
          <div class="next02-sc-hero-actions">
            <a href="/scevents" class="btn btn-outline-secondary" target="_blank" rel="noopener noreferrer">SC Events</a>
            <a href="/crawltasklist" class="btn btn-outline-primary" target="_blank" rel="noopener noreferrer">任务列表</a>
          </div>
        </section>

        <div v-if="message" class="next02-sc-feedback" :class="messageOk ? 'is-success' : 'is-error'" role="status">
          <span class="smart-pick-feedback-message">{{ message }}</span>
          <span v-if="messageOk && pickInfoRows.length" class="smart-pick-feedback-summary">
            <span v-for="row in pickInfoRows" :key="row.label" class="smart-pick-feedback-stat"><span>{{ row.label }}</span><strong>{{ row.value }}</strong></span>
          </span>
        </div>

        <div class="card next02-sc-panel next02-sc-movie-table-card smart-pick-config-panel">
          <div class="card-body">
            <form ref="formEl" id="formPickSmart" class="needs-validation" novalidate :data-source="pageSource" data-movie-type="mixed" @submit.prevent="runPick(true)">
              <div class="smart-pick-strategy-row">
                <div class="card border smart-pick-global-panel">
                  <div class="card-body">
                    <div class="smart-pick-section-title"><span>01</span><h6>混合抽取数量</h6></div>
                    <div class="row g-3 align-items-end smart-pick-global-grid">
                      <FieldInput v-for="field in globalBasicFields" :key="field.id" col-class="col-md-2" v-bind="field" />
                      <div class="col-md-2 d-flex align-items-end smart-pick-advanced-toggle-wrap">
                        <button
                          type="button"
                          class="btn btn-sm btn-outline-secondary w-100 smart-pick-advanced-toggle"
                          data-role="global-adv-toggle"
                          data-bs-toggle="collapse"
                          data-bs-target="#smartGlobalAdvancedPanel"
                          aria-controls="smartGlobalAdvancedPanel"
                          :aria-expanded="globalAdvancedExpanded ? 'true' : 'false'"
                          @click="globalAdvancedExpanded = !globalAdvancedExpanded"
                        >高级参数</button>
                      </div>
                      <div id="smartGlobalAdvancedPanel" class="col-12 collapse smart-pick-advanced-panel" :class="{ show: globalAdvancedExpanded }" data-role="global-adv-panel">
                        <div class="smart-pick-mixed-options-grid">
                          <section class="smart-pick-type-options is-jav" aria-label="JAV 抽取策略">
                            <h6>JAV 抽取策略</h6>
                            <div class="row g-3 smart-pick-advanced-grid">
                              <FieldInput v-for="field in javGlobalAdvancedFields" :key="field.id" col-class="col-md-6" v-bind="field" />
                            </div>
                          </section>
                          <section class="smart-pick-type-options is-fc2" aria-label="FC2 抽取策略">
                            <h6>FC2 抽取策略</h6>
                            <div class="row g-3 smart-pick-advanced-grid">
                              <FieldInput v-for="field in fc2GlobalAdvancedFields" :key="field.id" col-class="col-md-6" v-bind="field" />
                            </div>
                          </section>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <section class="smart-pick-type-groups is-jav" aria-labelledby="javPickGroupsTitle">
                  <div class="smart-pick-type-groups-head">
                    <div><span>JAV</span><h6 id="javPickGroupsTitle">JAV 条件</h6></div>
                    <button type="button" id="btnAddJAVGroup" class="btn btn-outline-secondary" :disabled="actionLoading" @click="addGroup('jav')">＋ 添加 JAV 条件</button>
                  </div>
                  <div id="javPickGroups">
                    <PickGroup
                      v-for="group in javGroups"
                      :key="group.id"
                      :payload="payload"
                      movie-type="jav"
                      :group-id="group.id"
                      :can-remove="javGroups.length > 1"
                      :expanded="group.expanded"
                      @remove="removeGroup('jav', group.id)"
                      @toggle="toggleGroup('jav', group.id)"
                    />
                  </div>
                </section>

                <section class="smart-pick-type-groups is-fc2" aria-labelledby="fc2PickGroupsTitle">
                  <div class="smart-pick-type-groups-head">
                    <div><span>FC2</span><h6 id="fc2PickGroupsTitle">FC2 条件</h6></div>
                    <button type="button" id="btnAddFC2Group" class="btn btn-outline-secondary" :disabled="actionLoading" @click="addGroup('fc2ppv')">＋ 添加 FC2 条件</button>
                  </div>
                  <div id="fc2PickGroups">
                    <PickGroup
                      v-for="group in fc2Groups"
                      :key="group.id"
                      :payload="payload"
                      movie-type="fc2ppv"
                      :group-id="group.id"
                      :can-remove="fc2Groups.length > 1"
                      :expanded="group.expanded"
                      @remove="removeGroup('fc2ppv', group.id)"
                      @toggle="toggleGroup('fc2ppv', group.id)"
                    />
                  </div>
                </section>
              </div>

              <div class="smart-pick-actions">
                <button type="submit" class="btn btn-primary" :disabled="actionLoading">{{ actionLoading ? "执行中..." : "执行 Smart Pick + Copy" }}</button>
                <button type="button" class="btn btn-outline-primary" id="btnPickOnly" :disabled="actionLoading" @click="runPick(false)">仅 Smart Pick</button>
              </div>
            </form>
          </div>
        </div>

        <div v-if="copyVisible" class="card next02-sc-panel smart-pick-copy-panel" id="copyCard">
          <div class="card-body">
            <div class="next02-sc-panel-head">
              <div><span class="next02-sc-panel-kicker">COPY STATUS</span><h5>Copy 进度</h5></div>
              <button type="button" class="btn btn-sm btn-outline-danger" id="btnStopCopy" :disabled="!copyRunning" @click="stopCopy">停止 Copy</button>
            </div>
            <div class="smart-pick-copy-status-row">
              <span class="smart-pick-copy-meta" id="copyMeta">{{ copyMeta }}</span>
              <span class="smart-pick-copy-percent">{{ copyPercent }}%</span>
            </div>
            <div class="progress smart-pick-progress">
              <div class="progress-bar" role="progressbar" id="copyProgress" :style="{ width: `${copyPercent}%` }" :aria-valuenow="copyPercent" :aria-valuetext="`${copyPercent}%`" aria-valuemin="0" aria-valuemax="100"></div>
            </div>
            <div class="smart-pick-copy-current" id="copyCurrent" :title="copyCurrent">{{ copyCurrent }}</div>
            <div v-if="copyError" class="smart-pick-copy-error" id="copyError">{{ copyError }}</div>
          </div>
        </div>

        <div v-if="totalPickedCount" class="smart-pick-results smart-pick-mixed-results" id="pickedCard">
          <div class="card next02-sc-panel smart-pick-results-panel">
            <div class="card-body">
              <div class="next02-sc-panel-head smart-pick-results-head">
                <div><span class="next02-sc-panel-kicker">MIXED PICKED MOVIES</span><h5>混合抽取结果</h5></div>
                <span class="next02-sc-result-count" id="pickedCount">共 {{ totalPickedCount }} 部<span v-if="pickedTotalSizeGb"> · {{ pickedTotalSizeGb }} GB</span></span>
              </div>
            </div>
          </div>

          <section v-if="pickedJAVMovies.length" class="smart-pick-result-section is-jav" aria-label="JAV 抽取结果">
            <div class="smart-pick-result-section-head">
              <div><span>JAV</span><h5>JAV 结果</h5></div>
              <div class="smart-pick-sort-tools">
                <div class="btn-group btn-group-sm flex-wrap" role="group" aria-label="JAV 结果排序">
                  <button v-for="item in javSortButtons" :key="item.key" type="button" class="btn btn-outline-primary" :class="{ active: javSortBase === item.key }" @click="selectSort('jav', item)">{{ sortButtonLabel('jav', item) }}</button>
                </div>
                <span class="next02-sc-result-count">{{ pickedJAVMovies.length }} 部<span v-if="pickedJAVTotalSizeGb"> · {{ pickedJAVTotalSizeGb }} GB</span></span>
              </div>
            </div>
            <section class="smart-pick-cards-section" aria-label="JAV cards">
              <MovieGrid :cards="sortedPickedJAVMovies" />
            </section>
          </section>

          <section v-if="pickedFC2Movies.length" class="smart-pick-result-section is-fc2" aria-label="FC2 抽取结果">
            <div class="smart-pick-result-section-head">
              <div><span>FC2</span><h5>FC2 结果</h5></div>
              <div class="smart-pick-sort-tools">
                <div class="btn-group btn-group-sm flex-wrap" role="group" aria-label="FC2 结果排序">
                  <button v-for="item in fc2SortButtons" :key="item.key" type="button" class="btn btn-outline-primary" :class="{ active: fc2SortBase === item.key }" @click="selectSort('fc2ppv', item)">{{ sortButtonLabel('fc2ppv', item) }}</button>
                </div>
                <span class="next02-sc-result-count">{{ pickedFC2Movies.length }} 部<span v-if="pickedFC2TotalSizeGb"> · {{ pickedFC2TotalSizeGb }} GB</span></span>
              </div>
            </div>
            <FC2MovieGrid :cards="sortedPickedFC2Movies" aria-label="FC2 cards" />
          </section>
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onBeforeUnmount, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/shared/api/request";
import FC2MovieGrid from "@/features/fc2/components/FC2MovieGrid.vue";
import MovieGrid from "@/features/movies/components/MovieGrid.vue";
import PageShell from "@/shared/components/PageShell.vue";
import type { AnyRecord } from "@/shared/types/gc";
import { asArray, asRecord, num, pick, text } from "@/shared/utils/format";

type FieldConfig = {
  id?: string;
  label: string;
  type?: string;
  min?: string;
  step?: string;
  value?: string;
  placeholder?: string;
  reqField?: string;
  dataType?: string;
  role?: string;
  required?: boolean;
  options?: Array<{ value: string; label: string }>;
};

type PickGroupState = {
  id: number;
  expanded: boolean;
};

type SortButton = { key: string; label: string; dir: "asc" | "desc" };
type MovieType = "jav" | "fc2ppv";

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const javGroups = ref<PickGroupState[]>([{ id: 1, expanded: false }]);
const fc2Groups = ref<PickGroupState[]>([{ id: 1, expanded: false }]);
const globalAdvancedExpanded = ref(false);
const nextJAVGroupId = ref(2);
const nextFC2GroupId = ref(2);
const formEl = ref<HTMLFormElement | null>(null);
const actionLoading = ref(false);
const message = ref("");
const messageOk = ref(true);
const pickedJAVMovies = ref<AnyRecord[]>([]);
const pickedFC2Movies = ref<AnyRecord[]>([]);
const pickedTotalSizeGb = ref("");
const pickedJAVTotalSizeGb = ref("");
const pickedFC2TotalSizeGb = ref("");
const javPickInfo = ref<AnyRecord>({});
const fc2PickInfo = ref<AnyRecord>({});
const copyStatus = ref<AnyRecord>({});
const copyStartedInPage = ref(false);
const javSortBase = ref("releasing");
const javSortDir = ref<"asc" | "desc">("desc");
const fc2SortBase = ref("owned");
const fc2SortDir = ref<"asc" | "desc">("desc");
let copyTimer: number | undefined;

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
    copyStartedInPage.value = false;
    void fetchCopyStatus();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title") || "SC Pick Smart Media");
const pageTitle = computed(() => text(payload.value, "PageTitle") || title.value);
const pageNote = computed(() => text(payload.value, "PageNote"));
const pageSource = computed(() => (text(payload.value, "PageSource").toLowerCase() === "wmedia" ? "wmedia" : "vfilm"));
const castOwnedScRatioPenaltyAlpha = computed(() => pageSource.value === "wmedia" ? "0" : "0.7");
const effectivePageNote = computed(() => pageNote.value || "分别配置 JAV 与 FC2 数量和条件，一次完成混合抽片与 Copy。");

const globalBasicFields = computed<FieldConfig[]>(() => [
  { id: "javPickCount", label: "JAV 数量", type: "number", min: "0", value: "32", placeholder: "输入数量", required: true },
  { id: "fc2PickCount", label: "FC2 数量", type: "number", min: "0", value: "2", placeholder: "输入数量", required: true }
]);
const javGlobalAdvancedFields = computed<FieldConfig[]>(() => [
  { id: "castParticipatedScBlockDays", label: "主演参与阻断天", type: "number", min: "0", value: "30", required: true },
  { id: "castSelectedScBlockDays", label: "主演入选阻断天", type: "number", min: "0", value: "60", required: true },
  { id: "rank20Min", label: "Rank20 至少", type: "number", min: "0", value: "0", required: true },
  { id: "rank100Min", label: "Rank100 至少", type: "number", min: "0" },
  { id: "rank500Min", label: "Rank500 至少", type: "number", min: "0" },
  { id: "castParticipatedScPenaltyAlpha", label: "主演参与惩罚", type: "number", min: "0", step: "0.1", value: "0.7", required: true },
  { id: "castSelectedScPenaltyAlpha", label: "主演入选惩罚", type: "number", min: "0", step: "0.1", value: "0.7", required: true },
  { id: "castOwnedScRatioPenaltyAlpha", label: "主演入选比例惩罚", type: "number", min: "0", step: "0.1", value: castOwnedScRatioPenaltyAlpha.value, required: true },
  { id: "movieHasSelectedScPenaltyAlpha", label: "影片已入选惩罚", type: "number", min: "0", step: "0.1", value: "0.7", required: true }
]);
const fc2GlobalAdvancedFields: FieldConfig[] = [
  { id: "fc2MovieParticipatedScBlockDays", label: "影片参与阻断天", type: "number", min: "0", value: "30", required: true },
  { id: "fc2MovieSelectedScBlockDays", label: "影片入选阻断天", type: "number", min: "0", value: "30", required: true },
  { id: "fc2MovieSelectedScPenaltyAlpha", label: "影片入选惩罚", type: "number", min: "0", step: "0.1", value: "0.7", required: true },
  { id: "fc2MovieHasSelectedScPenaltyAlpha", label: "影片已入选惩罚", type: "number", min: "0", step: "0.1", value: "0.7", required: true }
];

const javSortButtons: SortButton[] = [
  { key: "picked", label: "抽取顺序", dir: "desc" },
  { key: "film_birth", label: "下载时间", dir: "desc" },
  { key: "releasing", label: "上映日期", dir: "desc" },
  { key: "sc", label: "SC次数", dir: "desc" },
  { key: "come", label: "CO次数", dir: "desc" },
  { key: "rank", label: "最高排名", dir: "asc" },
  { key: "score", label: "评分", dir: "desc" }
];
const fc2SortButtons: SortButton[] = [
  { key: "owned", label: "Owned 排列", dir: "desc" },
  { key: "picked", label: "抽取顺序", dir: "desc" },
  { key: "film_birth", label: "下载时间", dir: "desc" },
  { key: "post", label: "发布时间", dir: "desc" },
  { key: "sc", label: "SC次数", dir: "desc" },
  { key: "come", label: "CO次数", dir: "desc" }
];

const copyRunning = computed(() => Boolean(pick(copyStatus.value, "running")));
const copyTotal = computed(() => num(copyStatus.value, "total"));
const copyDone = computed(() => num(copyStatus.value, "done"));
const copyVisible = computed(() => copyRunning.value || Boolean(pick(copyStatus.value, "stopped")) || Boolean(pick(copyStatus.value, "finished_at")) || copyTotal.value > 0);
const copyPercent = computed(() => copyTotal.value > 0 ? Math.min(100, Math.round((copyDone.value / copyTotal.value) * 100)) : 0);
const copyMeta = computed(() => {
  const state = copyRunning.value ? "复制中" : (pick(copyStatus.value, "stopped") ? "已停止" : (pick(copyStatus.value, "finished_at") ? "已完成" : "等待中"));
  return `${state} · ${copyDone.value}/${copyTotal.value}`;
});
const copyCurrent = computed(() => text(copyStatus.value, "current") ? `当前：${text(copyStatus.value, "current")}` : "");
const copyError = computed(() => text(copyStatus.value, "last_error"));
const totalPickedCount = computed(() => pickedJAVMovies.value.length + pickedFC2Movies.value.length);
const pickInfoRows = computed(() => {
  if (!Object.keys(javPickInfo.value).length && !Object.keys(fc2PickInfo.value).length) return [];
  return [
    { label: "JAV 候选", value: num(javPickInfo.value, "raw_candidate_movie_count") },
    { label: "JAV 抽取", value: pickedJAVMovies.value.length },
    { label: "FC2 候选", value: num(fc2PickInfo.value, "raw_candidate_movie_count") },
    { label: "FC2 抽取", value: pickedFC2Movies.value.length }
  ];
});

const sortedPickedJAVMovies = computed(() => {
  const list = pickedJAVMovies.value.slice();
  const dir = javSortDir.value === "asc" ? 1 : -1;
  const order = (row: AnyRecord) => num(row, "_pickedOrder");
  const numeric = (row: AnyRecord, ...keys: string[]) => Number(pick(row, ...keys) || 0);
  const dated = (row: AnyRecord, ...keys: string[]) => Date.parse(text(row, ...keys)) || 0;
  const compare = (a: AnyRecord, b: AnyRecord) => {
    if (javSortBase.value === "picked") return order(a) - order(b);
    if (javSortBase.value === "film_birth") return (dated(a, "film_birth_date", "FilmBirthDate") - dated(b, "film_birth_date", "FilmBirthDate")) * dir || order(a) - order(b);
    if (javSortBase.value === "releasing") return (dated(a, "releasing_date", "ReleasingDate") - dated(b, "releasing_date", "ReleasingDate")) * dir || order(a) - order(b);
    if (javSortBase.value === "sc") return (numeric(a, "sc_times", "ScTimes") - numeric(b, "sc_times", "ScTimes")) * dir || order(a) - order(b);
    if (javSortBase.value === "come") return (numeric(a, "come_times", "ComeTimes") - numeric(b, "come_times", "ComeTimes")) * dir || order(a) - order(b);
    if (javSortBase.value === "score") return (numeric(a, "score", "Score") - numeric(b, "score", "Score")) * dir || order(a) - order(b);
    const aRank = numeric(a, "highest_rank", "HighestRank") || Number.MAX_SAFE_INTEGER;
    const bRank = numeric(b, "highest_rank", "HighestRank") || Number.MAX_SAFE_INTEGER;
    return (aRank - bRank) * dir || order(a) - order(b);
  };
  return list.sort(compare);
});

const sortedPickedFC2Movies = computed(() => {
  const list = pickedFC2Movies.value.slice();
  const dir = fc2SortDir.value === "asc" ? 1 : -1;
  const order = (row: AnyRecord) => num(row, "_pickedOrder");
  const numeric = (row: AnyRecord, ...keys: string[]) => Number(pick(row, ...keys) || 0);
  const compareOwned = (a: AnyRecord, b: AnyRecord) =>
    numeric(b, "PostTime") - numeric(a, "PostTime") ||
    numeric(b, "PPVNo") - numeric(a, "PPVNo") ||
    numeric(a, "CodeSeq") - numeric(b, "CodeSeq") ||
    numeric(b, "ID") - numeric(a, "ID") ||
    order(a) - order(b);
  return list.sort((a, b) => {
    if (fc2SortBase.value === "owned") return compareOwned(a, b);
    if (fc2SortBase.value === "picked") return order(a) - order(b);
    if (fc2SortBase.value === "film_birth") return (numeric(a, "DownloadTime") - numeric(b, "DownloadTime")) * dir || order(a) - order(b);
    if (fc2SortBase.value === "post") return (numeric(a, "PostTime") - numeric(b, "PostTime")) * dir || order(a) - order(b);
    if (fc2SortBase.value === "sc") return (numeric(a, "ScTimes") - numeric(b, "ScTimes")) * dir || order(a) - order(b);
    return (numeric(a, "ComeTimes") - numeric(b, "ComeTimes")) * dir || order(a) - order(b);
  });
});

function addGroup(movieType: MovieType) {
  if (movieType === "jav") {
    javGroups.value.push({ id: nextJAVGroupId.value, expanded: false });
    nextJAVGroupId.value += 1;
    return;
  }
  fc2Groups.value.push({ id: nextFC2GroupId.value, expanded: false });
  nextFC2GroupId.value += 1;
}

function removeGroup(movieType: MovieType, id: number) {
  const target = movieType === "jav" ? javGroups : fc2Groups;
  if (target.value.length <= 1) return;
  target.value = target.value.filter((group) => group.id !== id);
}

function toggleGroup(movieType: MovieType, id: number) {
  const target = movieType === "jav" ? javGroups : fc2Groups;
  target.value = target.value.map((group) => group.id === id ? { ...group, expanded: !group.expanded } : group);
}

function inputValue(id: string): string {
  return String((document.getElementById(id) as HTMLInputElement | null)?.value || "").trim();
}

function parseNumber(raw: string, float = false): number | null {
  const value = float ? Number.parseFloat(raw) : Number.parseInt(raw, 10);
  return Number.isFinite(value) ? value : null;
}

function showMessage(value: string, ok = true) {
  message.value = value;
  messageOk.value = ok;
}

function buildReqFromGroup(group: Element, movieType: MovieType): AnyRecord {
  const req: AnyRecord = movieType === "jav" ? { Page: 1, MediaOwned: 3 } : {};
  group.querySelectorAll<HTMLInputElement | HTMLSelectElement>("[data-req-field]").forEach((input) => {
    const key = input.dataset.reqField || "";
    const type = input.dataset.type || "string";
    const raw = input.value.trim();
    if (!key) return;
    if (type === "int" || type === "float") {
      req[key] = parseNumber(raw, type === "float") ?? 0;
      return;
    }
    if (type === "int-opt" || type === "float-opt") {
      if (!raw) return;
      const value = parseNumber(raw, type === "float-opt");
      if (value !== null) req[key] = value;
      return;
    }
    if (raw) req[key] = raw;
  });
  return req;
}

function buildFC2Options(): AnyRecord | null {
  const movieParticipatedScBlockDays = parseNumber(inputValue("fc2MovieParticipatedScBlockDays"));
  const movieSelectedScBlockDays = parseNumber(inputValue("fc2MovieSelectedScBlockDays"));
  const movieSelectedScPenaltyAlpha = parseNumber(inputValue("fc2MovieSelectedScPenaltyAlpha"), true);
  const movieHasSelectedScPenaltyAlpha = parseNumber(inputValue("fc2MovieHasSelectedScPenaltyAlpha"), true);
  if ([movieParticipatedScBlockDays, movieSelectedScBlockDays, movieSelectedScPenaltyAlpha, movieHasSelectedScPenaltyAlpha].some((value) => value === null)) {
    showMessage("FC2 全局抽取策略参数不完整", false);
    return null;
  }
  return { movieParticipatedScBlockDays, movieSelectedScBlockDays, movieSelectedScPenaltyAlpha, movieHasSelectedScPenaltyAlpha };
}

function buildJAVOptions(n: number): AnyRecord | null {
  const castParticipatedScBlockDays = parseNumber(inputValue("castParticipatedScBlockDays"));
  const castSelectedScBlockDays = parseNumber(inputValue("castSelectedScBlockDays"));
  const rank20Min = parseNumber(inputValue("rank20Min") || "0");
  let rank100Min = inputValue("rank100Min") ? parseNumber(inputValue("rank100Min")) : rank20Min;
  let rank500Min = inputValue("rank500Min") ? parseNumber(inputValue("rank500Min")) : rank100Min;
  const castParticipatedScPenaltyAlpha = parseNumber(inputValue("castParticipatedScPenaltyAlpha"), true);
  const castSelectedScPenaltyAlpha = parseNumber(inputValue("castSelectedScPenaltyAlpha"), true);
  const castOwnedScRatioPenaltyAlphaValue = parseNumber(inputValue("castOwnedScRatioPenaltyAlpha"), true);
  const movieHasSelectedScPenaltyAlpha = parseNumber(inputValue("movieHasSelectedScPenaltyAlpha"), true);
  if ([castParticipatedScBlockDays, castSelectedScBlockDays, rank20Min, rank100Min, rank500Min, castParticipatedScPenaltyAlpha, castSelectedScPenaltyAlpha, castOwnedScRatioPenaltyAlphaValue, movieHasSelectedScPenaltyAlpha].some((value) => value === null)) {
    showMessage("全局抽取策略参数不完整", false);
    return null;
  }
  rank100Min = rank100Min as number;
  rank500Min = rank500Min as number;
  if ((rank20Min as number) > rank100Min || rank100Min > rank500Min || rank500Min > n) {
    showMessage("需要满足 0 <= rank20 <= rank100 <= rank500 <= 抽取数量", false);
    return null;
  }
  return {
    castParticipatedScBlockDays,
    castSelectedScBlockDays,
    rank20Min,
    rank100Min,
    rank500Min,
    castParticipatedScPenaltyAlpha,
    castSelectedScPenaltyAlpha,
    castOwnedScRatioPenaltyAlpha: castOwnedScRatioPenaltyAlphaValue,
    movieHasSelectedScPenaltyAlpha
  };
}

function buildGroupRequests(containerID: string, movieType: MovieType): AnyRecord[] | null {
  const groupElements = Array.from(document.querySelectorAll(`#${containerID} .pick-group`));
  if (!groupElements.length) {
    showMessage(`${movieType === "jav" ? "JAV" : "FC2"} 至少需要一组条件`, false);
    return null;
  }
  const reqs: AnyRecord[] = [];
  for (const group of groupElements) {
    const weight = parseNumber(String(group.querySelector<HTMLInputElement>('[data-role="weight"]')?.value || ""));
    if (weight === null || weight <= 0) {
      showMessage(`${movieType === "jav" ? "JAV" : "FC2"} 条件权重需为正数`, false);
      return null;
    }
    const groupReq = buildReqFromGroup(group, movieType);
    reqs.push(movieType === "jav" ? { weight, req: groupReq } : { weight, fc2Req: groupReq });
  }
  return reqs;
}

function buildPayload(): AnyRecord | null {
  if (!formEl.value?.checkValidity()) {
    formEl.value?.classList.add("was-validated");
    showMessage("请检查必填参数", false);
    return null;
  }
  const javPickN = parseNumber(inputValue("javPickCount"));
  const fc2PickN = parseNumber(inputValue("fc2PickCount"));
  if (javPickN === null || fc2PickN === null || javPickN < 0 || fc2PickN < 0) {
    showMessage("JAV 与 FC2 数量需要填写非负整数", false);
    return null;
  }
  if (javPickN === 0 && fc2PickN === 0) {
    showMessage("JAV 与 FC2 至少填写一个大于 0 的数量", false);
    return null;
  }
  const javOptions = javPickN > 0 ? buildJAVOptions(javPickN) : {};
  if (javOptions === null) return null;
  const fc2Options = fc2PickN > 0 ? buildFC2Options() : {};
  if (fc2Options === null) return null;
  const javReqs = javPickN > 0 ? buildGroupRequests("javPickGroups", "jav") : [];
  if (javReqs === null) return null;
  const fc2Reqs = fc2PickN > 0 ? buildGroupRequests("fc2PickGroups", "fc2ppv") : [];
  if (fc2Reqs === null) return null;
  return {
    source: pageSource.value,
    jav: { pickN: javPickN, options: javOptions, reqs: javReqs },
    fc2: { pickN: fc2PickN, options: fc2Options, reqs: fc2Reqs }
  };
}

async function runPick(withCopy: boolean) {
  const body = buildPayload();
  if (!body) return;
  actionLoading.value = true;
  message.value = "";
  javPickInfo.value = {};
  fc2PickInfo.value = {};
  try {
    const data = asRecord(await sendJSON(withCopy ? "/api/triggers/sc/pick-smart-copy" : "/api/triggers/sc/pick-smart-only", "POST", body));
    const javData = asRecord(pick(data, "jav", "JAV"));
    const fc2Data = asRecord(pick(data, "fc2", "FC2"));
    javPickInfo.value = asRecord(pick(javData, "pick_info", "PickInfo"));
    fc2PickInfo.value = asRecord(pick(fc2Data, "pick_info", "PickInfo"));
    pickedJAVMovies.value = asArray<AnyRecord>(pick(javData, "movies", "Movies")).map((movie, index) => ({ ...movie, _pickedOrder: index }));
    pickedFC2Movies.value = asArray<AnyRecord>(pick(fc2Data, "movies", "Movies")).map((movie, index) => ({ ...movie, _pickedOrder: index }));
    pickedTotalSizeGb.value = text(data, "total_size_gb", "TotalSizeGB");
    pickedJAVTotalSizeGb.value = text(javData, "total_size_gb", "TotalSizeGB");
    pickedFC2TotalSizeGb.value = text(fc2Data, "total_size_gb", "TotalSizeGB");
    const nextCopyStatus = asRecord(pick(data, "copy_status", "CopyStatus"));
    if (Object.keys(nextCopyStatus).length) {
      copyStatus.value = nextCopyStatus;
      copyStartedInPage.value = Boolean(pick(data, "copy_started", "CopyStarted")) || Boolean(pick(nextCopyStatus, "running"));
      if (copyRunning.value) startCopyPolling();
    }
    if (pick(data, "copy_started", "CopyStarted") === false && copyRunning.value) {
      showMessage("Copy 任务正在执行，本次不再启动", false);
      return;
    }
    showMessage(`${withCopy ? "Smart Pick + Copy" : "Smart Pick"} 完成：JAV ${pickedJAVMovies.value.length} 部，FC2 ${pickedFC2Movies.value.length} 部，共 ${num(data, "picked", "Picked")} 部`);
  } catch (err) {
    showMessage(`异常：${err instanceof Error ? err.message : String(err)}`, false);
  } finally {
    actionLoading.value = false;
  }
}

async function fetchCopyStatus() {
  try {
    const data = asRecord(await getJSON("/api/triggers/sc/copy-status"));
    if (!copyStartedInPage.value && !pick(data, "running")) {
      copyStatus.value = {};
      stopCopyPolling();
      return;
    }
    copyStatus.value = data;
    if (copyRunning.value) startCopyPolling();
    else stopCopyPolling();
  } catch {
    stopCopyPolling();
  }
}

function startCopyPolling() {
  if (copyTimer !== undefined) return;
  copyTimer = window.setInterval(() => void fetchCopyStatus(), 2000);
}

function stopCopyPolling() {
  if (copyTimer === undefined) return;
  window.clearInterval(copyTimer);
  copyTimer = undefined;
}

async function stopCopy() {
  try {
    const data = asRecord(await sendJSON("/api/triggers/sc/copy-stop", "POST", {}));
    copyStatus.value = asRecord(pick(data, "status", "Status"));
    showMessage("已发送停止指令");
    stopCopyPolling();
  } catch (err) {
    showMessage(`停止失败：${err instanceof Error ? err.message : String(err)}`, false);
  }
}

function selectSort(movieType: MovieType, item: SortButton) {
  const base = movieType === "jav" ? javSortBase : fc2SortBase;
  const direction = movieType === "jav" ? javSortDir : fc2SortDir;
  const fixedOrder = item.key === "picked" || item.key === "owned";
  if (base.value === item.key && !fixedOrder) {
    direction.value = direction.value === "asc" ? "desc" : "asc";
  } else {
    base.value = item.key;
    direction.value = item.dir;
  }
}

function sortButtonLabel(movieType: MovieType, item: SortButton): string {
  const base = movieType === "jav" ? javSortBase.value : fc2SortBase.value;
  const direction = movieType === "jav" ? javSortDir.value : fc2SortDir.value;
  if (base !== item.key || item.key === "picked" || item.key === "owned") return item.label;
  return `${item.label} ${direction === "asc" ? "↑" : "↓"}`;
}

onBeforeUnmount(stopCopyPolling);

function dynamicLabel(row: AnyRecord, key: string, fallback: string): string {
  return text(row, key) || fallback;
}

function dynamicField(row: AnyRecord, key: string, fallback: string): string {
  return text(row, key) || fallback;
}

const FieldInput = defineComponent({
  name: "SmartPickFieldInput",
  props: {
    colClass: { type: String, default: "" },
    id: { type: String, required: true },
    label: { type: String, required: true },
    type: { type: String, default: "text" },
    min: { type: String, default: "" },
    step: { type: String, default: "" },
    value: { type: String, default: "" },
    placeholder: { type: String, default: "" },
    reqField: { type: String, default: "" },
    dataType: { type: String, default: "" },
    role: { type: String, default: "" },
    required: { type: Boolean, default: false }
  },
  setup(props) {
    return () => h("div", { class: [props.colClass || "col-md-3", "smart-pick-field"] }, [
      h("label", { class: "form-label", for: props.id }, props.label),
      h("input", {
        type: props.type,
        min: props.min || undefined,
        step: props.step || undefined,
        class: "form-control",
        id: props.id,
        value: props.value,
        placeholder: props.placeholder || undefined,
        "data-req-field": props.reqField || undefined,
        "data-type": props.dataType || undefined,
        "data-role": props.role || undefined,
        required: props.required || undefined,
        inputmode: props.placeholder === "YYYY-MM-DD" ? "numeric" : undefined,
        maxlength: props.placeholder === "YYYY-MM-DD" ? "10" : undefined,
        pattern: props.placeholder === "YYYY-MM-DD" ? "[0-9]{4}-[0-9]{2}-[0-9]{2}" : undefined
      })
    ]);
  }
});

const FieldSelect = defineComponent({
  name: "SmartPickFieldSelect",
  props: {
    colClass: { type: String, default: "" },
    id: { type: String, required: true },
    label: { type: String, required: true },
    reqField: { type: String, required: true },
    dataType: { type: String, default: "int-opt" },
    options: { type: Array as PropType<Array<{ value: string; label: string }>>, required: true }
  },
  setup(props) {
    return () => h("div", { class: [props.colClass || "col-md-3", "smart-pick-field"] }, [
      h("label", { class: "form-label", for: props.id }, props.label),
      h("select", {
        class: "form-select",
        id: props.id,
        "data-req-field": props.reqField,
        "data-type": props.dataType
      }, props.options.map((option) => h("option", { key: option.value, value: option.value }, option.label)))
    ]);
  }
});

const PickGroup = defineComponent({
  name: "SmartPickGroup",
  props: {
    payload: { type: Object as PropType<AnyRecord>, required: true },
    movieType: { type: String as PropType<"jav" | "fc2ppv">, required: true },
    groupId: { type: Number, required: true },
    canRemove: { type: Boolean, default: true },
    expanded: { type: Boolean, default: false }
  },
  emits: ["remove", "toggle"],
  setup(props, { emit }) {
    const typeKey = () => props.movieType === "jav" ? "jav" : "fc2";
    const fieldId = (name: string) => `smart-${typeKey()}-${props.groupId}-${name}`;
    const advPanelId = () => `smart-${typeKey()}-adv-panel-${props.groupId}`;
    const basicFields = (): FieldConfig[] => props.movieType === "jav" ? [
      { id: fieldId("rs"), label: "放映日期起 (rs)", reqField: "ReleasingDateStart", placeholder: "YYYY-MM-DD" },
      { id: fieldId("re"), label: "放映日期止 (re)", reqField: "ReleasingDateEnd", placeholder: "YYYY-MM-DD" }
    ] : [
      { id: fieldId("pts"), label: "发布时间起", reqField: "PostTimeStart", placeholder: "YYYY-MM-DD" },
      { id: fieldId("pte"), label: "发布时间止", reqField: "PostTimeEnd", placeholder: "YYYY-MM-DD" }
    ];
    const javAdvancedFields = (): FieldConfig[] => [
      { reqField: dynamicField(props.payload, "BirthStartField", "MediaBirthTimeStart"), label: dynamicLabel(props.payload, "BirthStartLabel", "下载日期起 (mbs)"), placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: dynamicField(props.payload, "BirthEndField", "MediaBirthTimeEnd"), label: dynamicLabel(props.payload, "BirthEndLabel", "下载日期止 (mbe)"), placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: "ScTimesMin", label: "ScTimes 最小 (scmin)", type: "number", dataType: "int-opt", placeholder: "可为空", colClass: "col-md-3" },
      { reqField: "ScTimesMax", label: "ScTimes 最大 (scmax)", type: "number", dataType: "int-opt", placeholder: "可为空", colClass: "col-md-3" },
      { reqField: "ComeTimesMin", label: "ComeTimes 最小 (comin)", type: "number", dataType: "int-opt", placeholder: "可为空", colClass: "col-md-3" },
      { reqField: "ComeTimesMax", label: "ComeTimes 最大 (comax)", type: "number", dataType: "int-opt", placeholder: "可为空", colClass: "col-md-3" },
      { reqField: "CastNames", label: "演员 (cn)" },
      { reqField: "GenreNames", label: "类型 (gn)" },
      { reqField: "DirectorName", label: "导演 (dn)" },
      { reqField: "PrefixName", label: "前缀 (pn)" },
      { reqField: "MakerName", label: "制作商 (mn)" },
      { reqField: "LabelName", label: "厂牌 (ln)" },
      { reqField: dynamicField(props.payload, "Dir4Field", "MediaDir4"), label: dynamicLabel(props.payload, "Dir4Label", "MediaDir4 (md4)") },
      { reqField: "Word", label: "关键字 (wd)", placeholder: "可为空" },
      { reqField: "ScoreMin", label: "评分最小 (smin)", type: "number", step: "0.1", dataType: "float", colClass: "col-md-3" },
      { reqField: "ScoreMax", label: "评分最大 (smax)", type: "number", step: "0.1", dataType: "float-opt", colClass: "col-md-3" },
      { reqField: "HighestRankMin", label: "最高排名最小", type: "number", dataType: "int-opt", colClass: "col-md-3" },
      { reqField: "HighestRankMax", label: "最高排名最大", type: "number", dataType: "int-opt", colClass: "col-md-3" },
      { reqField: "ViewWatchedMin", label: "观看最小 (vwmin)", type: "number", dataType: "int", colClass: "col-md-3" },
      { reqField: "ViewWatchedMax", label: "观看最大 (vwmax)", type: "number", dataType: "int-opt", colClass: "col-md-3" },
      { reqField: "LastSelectedScTimeMin", label: "最近入选 SC 起 (lsstmin)", placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: "LastSelectedScTimeMax", label: "最近入选 SC 止 (lsstmax)", placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: dynamicField(props.payload, "Dir1Field", "MediaDir1"), label: dynamicLabel(props.payload, "Dir1Label", "MediaDir1 (md1)"), colClass: "col-md-3" },
      { reqField: dynamicField(props.payload, "Dir2Field", "MediaDir2"), label: dynamicLabel(props.payload, "Dir2Label", "MediaDir2 (md2)"), colClass: "col-md-3" },
      { reqField: dynamicField(props.payload, "Dir3Field", "MediaDir3"), label: dynamicLabel(props.payload, "Dir3Label", "MediaDir3 (md3)"), colClass: "col-md-3" },
      { reqField: "OrderBy", label: "排序 (od)", placeholder: "sc/lsst/bt...", colClass: "col-md-3" }
    ];
    const fc2AdvancedFields = (): FieldConfig[] => [
      { reqField: "MediaBirthTimeStart", label: "下载日期起", placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: "MediaBirthTimeEnd", label: "下载日期止", placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: "ScTimesMin", label: "SC 次数最小", type: "number", min: "0", dataType: "int-opt", placeholder: "可为空", colClass: "col-md-3" },
      { reqField: "ScTimesMax", label: "SC 次数最大", type: "number", min: "0", dataType: "int-opt", placeholder: "可为空", colClass: "col-md-3" },
      { reqField: "ComeTimesMin", label: "Come 次数最小", type: "number", min: "0", dataType: "int-opt", placeholder: "可为空", colClass: "col-md-3" },
      { reqField: "ComeTimesMax", label: "Come 次数最大", type: "number", min: "0", dataType: "int-opt", placeholder: "可为空", colClass: "col-md-3" },
      { reqField: "LastSelectedScTimeMin", label: "最近入选 SC 起", placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: "LastSelectedScTimeMax", label: "最近入选 SC 止", placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: "FC2Key", label: "FC2 Key", placeholder: "支持包含匹配" },
      { reqField: "Code", label: "FC2 Code", placeholder: "支持包含匹配" },
      { reqField: "Title", label: "标题", placeholder: "支持包含匹配" },
      { reqField: "PerformerText", label: "表演者", placeholder: "支持包含匹配" },
      { reqField: "MaskType", label: "影片遮罩类型", placeholder: "精确匹配" },
      { reqField: "Word", label: "综合关键字", placeholder: "编号/标题/表演者" },
      { reqField: "MediaDir1", label: "媒体目录 1" },
      { reqField: "MediaDir2", label: "媒体目录 2" },
      { reqField: "MediaDir3", label: "媒体目录 3" },
      { reqField: "MediaDir4", label: "媒体目录 4" },
      { reqField: "HasSub", label: "字幕", dataType: "int-opt", options: [{ value: "", label: "全部" }, { value: "1", label: "无字幕" }, { value: "2", label: "有字幕" }] },
      { reqField: "SelfMake", label: "自制", dataType: "int-opt", options: [{ value: "", label: "全部" }, { value: "1", label: "非自制" }, { value: "2", label: "自制" }] },
      { reqField: "HasMask", label: "去码状态", dataType: "int-opt", options: [{ value: "", label: "全部" }, { value: "1", label: "未去码" }, { value: "2", label: "已去码" }, { value: "3", label: "无码" }] }
    ];
    const advFields = () => props.movieType === "jav" ? javAdvancedFields() : fc2AdvancedFields();
    return () => h("div", { class: "card border pick-group smart-pick-group-panel" }, [
      h("div", { class: "card-body" }, [
        h("div", { class: "smart-pick-group-head" }, [
          h("h6", `${props.movieType === "jav" ? "JAV" : "FC2"} 条件组 ${props.groupId}`),
          h("button", {
            type: "button",
            class: "btn btn-sm btn-outline-danger btn-remove-group smart-pick-remove-group",
            disabled: !props.canRemove,
            onClick: () => emit("remove")
          }, "移除")
        ]),
        h("div", { class: "row g-3 smart-pick-group-grid" }, [
          h(FieldInput, { colClass: "col-md-3", id: fieldId("weight"), label: "权重", type: "number", min: "1", value: "12", required: true, role: "weight" }),
          ...basicFields().map((field) => h(FieldInput, { colClass: "col-md-3", ...field })),
          h("div", { class: "col-md-3 d-flex align-items-end smart-pick-advanced-toggle-wrap" }, [
            h("button", {
              type: "button",
              class: "btn btn-sm btn-outline-secondary w-100 smart-pick-advanced-toggle",
              "data-role": "adv-toggle",
              "data-bs-toggle": "collapse",
              "data-bs-target": `#${advPanelId()}`,
              "aria-controls": advPanelId(),
              "aria-expanded": props.expanded ? "true" : "false",
              onClick: () => emit("toggle")
            }, "高级参数")
          ]),
          h("div", { class: `col-12 collapse smart-pick-advanced-panel${props.expanded ? " show" : ""}`, id: advPanelId(), "data-role": "adv-panel" }, [
            h("div", { class: "row g-3 smart-pick-advanced-grid" }, [
              ...advFields().map((field) => field.options
                ? h(FieldSelect, {
                    colClass: field.colClass || "col-md-4",
                    id: fieldId((field.reqField || "field").toLowerCase()),
                    label: field.label,
                    reqField: field.reqField || "",
                    dataType: field.dataType || "int-opt",
                    options: field.options
                  })
                : h(FieldInput, {
                    colClass: field.colClass || "col-md-4",
                    id: fieldId((field.reqField || "field").toLowerCase()),
                    label: field.label,
                    type: field.type || "text",
                    min: field.min || "",
                    step: field.step || "",
                    reqField: field.reqField || "",
                    dataType: field.dataType || "",
                    placeholder: field.placeholder || ""
                  }))
            ])
          ])
        ])
      ])
    ]);
  }
});
</script>
