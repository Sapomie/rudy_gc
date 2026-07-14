<template>
  <PageShell class="sc-events-page-shell sc-pick-page-shell cards-page-shell next02-pink-cards-page" :title="title" eyebrow="">
    <div id="page-root" class="container-fluid sc-pick-page-root next02-sc-pink-theme next02-smart-pick-page-rewrite">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <section class="next02-sc-movie-top-shell next02-smart-pick-hero">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">SC / SMART PICK LAB</span>
            <h4>{{ pageTitle }}</h4>
            <p>{{ pageNote || "组合多组条件、Rank 配额和 SC 阻断策略，完成抽片与 Copy。" }}</p>
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
            <form ref="formEl" id="formPickSmart" class="needs-validation" novalidate :data-source="pageSource" @submit.prevent="runPick(true)">
              <div class="smart-pick-strategy-row">
                <div class="card border smart-pick-global-panel">
                  <div class="card-body">
                    <div class="smart-pick-section-title"><span>01</span><h6>全局抽取策略</h6></div>
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
                        <div class="row g-3 smart-pick-advanced-grid">
                          <FieldInput v-for="field in globalAdvancedFields" :key="field.id" col-class="col-md-3" v-bind="field" />
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="smart-pick-groups-column">
                  <div class="smart-pick-groups-head">
                    <div class="smart-pick-section-title"><span>02</span><h6>条件组</h6></div>
                    <button type="button" id="btnAddGroup" class="btn btn-outline-secondary" :disabled="actionLoading" @click="addGroup">＋ 添加一组条件</button>
                  </div>

                  <div id="pickGroups">
                    <PickGroup
                      v-for="group in groups"
                      :key="group.id"
                      :payload="payload"
                      :group-id="group.id"
                      :can-remove="groups.length > 1"
                      :expanded="group.expanded"
                      @remove="removeGroup(group.id)"
                      @toggle="toggleGroup(group.id)"
                    />
                  </div>
                </div>
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
            <div class="text-muted small mb-2" id="copyMeta">{{ copyMeta }}</div>
            <div class="progress mb-2 smart-pick-progress">
              <div class="progress-bar" role="progressbar" id="copyProgress" :style="{ width: `${copyPercent}%` }" :aria-valuenow="copyPercent" aria-valuemin="0" aria-valuemax="100">{{ copyPercent ? `${copyPercent}%` : '' }}</div>
            </div>
            <div class="small text-muted" id="copyCurrent">{{ copyCurrent }}</div>
            <div v-if="copyError" class="small text-danger mt-1" id="copyError">{{ copyError }}</div>
          </div>
        </div>

        <div v-if="pickedMovies.length" class="smart-pick-results" id="pickedCard">
          <div class="card next02-sc-panel smart-pick-results-panel">
            <div class="card-body">
              <div class="next02-sc-panel-head smart-pick-results-head">
                <div><span class="next02-sc-panel-kicker">PICKED MOVIES</span><h5>抽取结果</h5></div>
                <div class="smart-pick-sort-tools">
                  <div class="btn-group btn-group-sm flex-wrap" id="pickedSortBar" role="group" aria-label="抽取结果排序">
                    <button v-for="item in pickedSortButtons" :key="item.key" type="button" class="btn btn-outline-primary" :class="{ active: pickedSortBase === item.key }" :data-sort-base="item.key" :data-default-dir="item.dir" @click="selectSort(item)">{{ sortButtonLabel(item) }}</button>
                  </div>
                  <span class="next02-sc-result-count" id="pickedCount">共 {{ pickedMovies.length }} 部<span v-if="pickedTotalSizeGb"> · {{ pickedTotalSizeGb }} GB</span></span>
                </div>
              </div>
            </div>
          </div>
          <section class="smart-pick-cards-section" aria-label="Picked movies">
            <MovieGrid :cards="sortedPickedMovies" />
          </section>
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onBeforeUnmount, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import MovieGrid from "@/components/MovieGrid.vue";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, num, pick, text } from "@/utils/format";

type FieldConfig = {
  id: string;
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
};

type PickGroupState = {
  id: number;
  expanded: boolean;
};

type SortButton = { key: string; label: string; dir: "asc" | "desc" };

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const groups = ref<PickGroupState[]>([{ id: 1, expanded: false }]);
const globalAdvancedExpanded = ref(false);
const nextGroupId = ref(2);
const formEl = ref<HTMLFormElement | null>(null);
const actionLoading = ref(false);
const message = ref("");
const messageOk = ref(true);
const pickedMovies = ref<AnyRecord[]>([]);
const pickedTotalSizeGb = ref("");
const pickInfo = ref<AnyRecord>({});
const copyStatus = ref<AnyRecord>({});
const copyStartedInPage = ref(false);
const pickedSortBase = ref("releasing");
const pickedSortDir = ref<"asc" | "desc">("desc");
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

const globalFields = computed<FieldConfig[]>(() => [
  { id: "pickCount", label: "抽取数量", type: "number", min: "1", value: "32", required: true },
  { id: "castLastScBlockDays", label: "Cast LastSc 阻断天数", type: "number", min: "0", value: "60", required: true },
  { id: "lastScEventBlockDays", label: "LastSc Event 阻断天数", type: "number", min: "0", value: "30", required: true },
  { id: "rank20Min", label: "Rank20 至少", type: "number", min: "0", value: "0", required: true },
  { id: "rank100Min", label: "Rank100 至少", type: "number", min: "0" },
  { id: "rank500Min", label: "Rank500 至少", type: "number", min: "0" },
  { id: "castLastScPenaltyAlpha", label: "Cast LastSc 惩罚系数", type: "number", min: "0", step: "0.1", value: "0.7", required: true },
  { id: "lastScEventPenaltyAlpha", label: "LastSc Event 惩罚系数", type: "number", min: "0", step: "0.1", value: "0.7", required: true },
  { id: "castOwnedScRatioPenaltyAlpha", label: "Cast SC Ratio 惩罚系数", type: "number", min: "0", step: "0.1", value: castOwnedScRatioPenaltyAlpha.value, required: true },
  { id: "movieHasScPenaltyAlpha", label: "Movie Has SC 惩罚系数", type: "number", min: "0", step: "0.1", value: "0.7", required: true }
]);
const globalBasicFields = computed(() => globalFields.value.slice(0, 1));
const globalAdvancedFields = computed(() => globalFields.value.slice(1));

const pickedSortButtons: SortButton[] = [
  { key: "picked", label: "抽取顺序", dir: "desc" },
  { key: "film_birth", label: "下载时间", dir: "desc" },
  { key: "releasing", label: "上映日期", dir: "desc" },
  { key: "sc", label: "SC次数", dir: "desc" },
  { key: "come", label: "CO次数", dir: "desc" },
  { key: "rank", label: "最高排名", dir: "asc" },
  { key: "score", label: "评分", dir: "desc" }
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
const pickInfoAfterMovies = computed(() => num(pickInfo.value, "after_block_movie_count"));
const pickInfoAfterCasts = computed(() => num(pickInfo.value, "after_block_cast_count"));
const pickInfoRows = computed(() => Object.keys(pickInfo.value).length ? [
  { label: "原始候选电影", value: num(pickInfo.value, "raw_candidate_movie_count") },
  { label: "阻断后剩余电影", value: pickInfoAfterMovies.value },
  { label: "参与计算演员", value: num(pickInfo.value, "cast_count") },
  { label: "阻断后剩余演员", value: pickInfoAfterCasts.value }
] : []);

const sortedPickedMovies = computed(() => {
  const list = pickedMovies.value.slice();
  const dir = pickedSortDir.value === "asc" ? 1 : -1;
  const order = (row: AnyRecord) => num(row, "_pickedOrder");
  const numeric = (row: AnyRecord, ...keys: string[]) => Number(pick(row, ...keys) || 0);
  const dated = (row: AnyRecord, ...keys: string[]) => Date.parse(text(row, ...keys)) || 0;
  const compare = (a: AnyRecord, b: AnyRecord) => {
    if (pickedSortBase.value === "picked") return order(a) - order(b);
    if (pickedSortBase.value === "film_birth") return (dated(a, "film_birth_date", "FilmBirthDate") - dated(b, "film_birth_date", "FilmBirthDate")) * dir || order(a) - order(b);
    if (pickedSortBase.value === "releasing") return (dated(a, "releasing_date", "ReleasingDate") - dated(b, "releasing_date", "ReleasingDate")) * dir || order(a) - order(b);
    if (pickedSortBase.value === "sc") return (numeric(a, "sc_times", "ScTimes") - numeric(b, "sc_times", "ScTimes")) * dir || order(a) - order(b);
    if (pickedSortBase.value === "come") return (numeric(a, "come_times", "ComeTimes") - numeric(b, "come_times", "ComeTimes")) * dir || order(a) - order(b);
    if (pickedSortBase.value === "score") return (numeric(a, "score", "Score") - numeric(b, "score", "Score")) * dir || order(a) - order(b);
    const aRank = numeric(a, "highest_rank", "HighestRank") || Number.MAX_SAFE_INTEGER;
    const bRank = numeric(b, "highest_rank", "HighestRank") || Number.MAX_SAFE_INTEGER;
    return (aRank - bRank) * dir || order(a) - order(b);
  };
  return list.sort(compare);
});

function addGroup() {
  groups.value.push({ id: nextGroupId.value, expanded: false });
  nextGroupId.value += 1;
}

function removeGroup(id: number) {
  if (groups.value.length <= 1) return;
  groups.value = groups.value.filter((group) => group.id !== id);
}

function toggleGroup(id: number) {
  groups.value = groups.value.map((group) => group.id === id ? { ...group, expanded: !group.expanded } : group);
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

function buildReqFromGroup(group: Element): AnyRecord {
  const req: AnyRecord = { Page: 1, PageSize: 100000, MediaOwned: 3 };
  group.querySelectorAll<HTMLInputElement>("[data-req-field]").forEach((input) => {
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

function buildOptions(n: number): AnyRecord | null {
  const castLastScBlockDays = parseNumber(inputValue("castLastScBlockDays"));
  const lastScEventBlockDays = parseNumber(inputValue("lastScEventBlockDays"));
  const rank20Min = parseNumber(inputValue("rank20Min") || "0");
  let rank100Min = inputValue("rank100Min") ? parseNumber(inputValue("rank100Min")) : rank20Min;
  let rank500Min = inputValue("rank500Min") ? parseNumber(inputValue("rank500Min")) : rank100Min;
  const castLastScPenaltyAlpha = parseNumber(inputValue("castLastScPenaltyAlpha"), true);
  const lastScEventPenaltyAlpha = parseNumber(inputValue("lastScEventPenaltyAlpha"), true);
  const castOwnedScRatioPenaltyAlphaValue = parseNumber(inputValue("castOwnedScRatioPenaltyAlpha"), true);
  const movieHasScPenaltyAlpha = parseNumber(inputValue("movieHasScPenaltyAlpha"), true);
  if ([castLastScBlockDays, lastScEventBlockDays, rank20Min, rank100Min, rank500Min, castLastScPenaltyAlpha, lastScEventPenaltyAlpha, castOwnedScRatioPenaltyAlphaValue, movieHasScPenaltyAlpha].some((value) => value === null)) {
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
    castLastScBlockDays,
    lastScEventBlockDays,
    rank20Min,
    rank100Min,
    rank500Min,
    castLastScPenaltyAlpha,
    lastScEventPenaltyAlpha,
    castOwnedScRatioPenaltyAlpha: castOwnedScRatioPenaltyAlphaValue,
    movieHasScPenaltyAlpha
  };
}

function buildPayload(): AnyRecord | null {
  if (!formEl.value?.checkValidity()) {
    formEl.value?.classList.add("was-validated");
    showMessage("请检查必填参数", false);
    return null;
  }
  const pickN = parseNumber(inputValue("pickCount"));
  if (pickN === null || pickN <= 0) {
    showMessage("请输入大于 0 的数量", false);
    return null;
  }
  const options = buildOptions(pickN);
  if (!options) return null;
  const groupElements = Array.from(document.querySelectorAll("#pickGroups .pick-group"));
  if (!groupElements.length) {
    showMessage("请添加至少一组条件", false);
    return null;
  }
  const reqs: AnyRecord[] = [];
  for (const group of groupElements) {
    const weight = parseNumber(String(group.querySelector<HTMLInputElement>('[data-role="weight"]')?.value || ""));
    if (weight === null || weight <= 0) {
      showMessage("权重需为正数", false);
      return null;
    }
    reqs.push({ weight, req: buildReqFromGroup(group) });
  }
  return { pickN, source: pageSource.value, options, reqs };
}

async function runPick(withCopy: boolean) {
  const body = buildPayload();
  if (!body) return;
  actionLoading.value = true;
  message.value = "";
  pickInfo.value = {};
  try {
    const data = asRecord(await sendJSON(withCopy ? "/api/triggers/sc/pick-smart-copy" : "/api/triggers/sc/pick-smart-only", "POST", body));
    pickInfo.value = asRecord(pick(data, "pick_info", "PickInfo"));
    pickedMovies.value = asArray<AnyRecord>(pick(data, "movies", "Movies")).map((movie, index) => ({ ...movie, _pickedOrder: index }));
    pickedTotalSizeGb.value = text(data, "total_size_gb", "TotalSizeGB");
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
    showMessage(`${withCopy ? "Smart Pick + Copy" : "Smart Pick"} 完成，已抽取 ${num(data, "picked", "Picked")} 部`);
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

function selectSort(item: SortButton) {
  if (pickedSortBase.value === item.key && item.key !== "picked") {
    pickedSortDir.value = pickedSortDir.value === "asc" ? "desc" : "asc";
  } else {
    pickedSortBase.value = item.key;
    pickedSortDir.value = item.dir;
  }
}

function sortButtonLabel(item: SortButton): string {
  if (pickedSortBase.value !== item.key || item.key === "picked") return item.label;
  return `${item.label} ${pickedSortDir.value === "asc" ? "↑" : "↓"}`;
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

const PickGroup = defineComponent({
  name: "SmartPickGroup",
  props: {
    payload: { type: Object as PropType<AnyRecord>, required: true },
    groupId: { type: Number, required: true },
    canRemove: { type: Boolean, default: true },
    expanded: { type: Boolean, default: false }
  },
  emits: ["remove", "toggle"],
  setup(props, { emit }) {
    const fieldId = (name: string) => `smart-${props.groupId}-${name}`;
    const advPanelId = () => `smart-adv-panel-${props.groupId}`;
    const basicFields = () => [
      { id: fieldId("rs"), label: "放映日期起 (rs)", reqField: "ReleasingDateStart", placeholder: "YYYY-MM-DD" },
      { id: fieldId("re"), label: "放映日期止 (re)", reqField: "ReleasingDateEnd", placeholder: "YYYY-MM-DD" }
    ];
    const advFields = () => [
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
      { reqField: "LastScTimeMin", label: "最后评分起 (lsctmin)", placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: "LastScTimeMax", label: "最后评分止 (lsctmax)", placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: dynamicField(props.payload, "Dir1Field", "MediaDir1"), label: dynamicLabel(props.payload, "Dir1Label", "MediaDir1 (md1)"), colClass: "col-md-3" },
      { reqField: dynamicField(props.payload, "Dir2Field", "MediaDir2"), label: dynamicLabel(props.payload, "Dir2Label", "MediaDir2 (md2)"), colClass: "col-md-3" },
      { reqField: dynamicField(props.payload, "Dir3Field", "MediaDir3"), label: dynamicLabel(props.payload, "Dir3Label", "MediaDir3 (md3)"), colClass: "col-md-3" },
      { reqField: "OrderBy", label: "排序 (od)", placeholder: "sc/lsct/bt...", colClass: "col-md-3" }
    ];
    return () => h("div", { class: "card border pick-group smart-pick-group-panel" }, [
      h("div", { class: "card-body" }, [
        h("div", { class: "smart-pick-group-head" }, [
          h("h6", `条件组 ${props.groupId}`),
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
              ...advFields().map((field) => h(FieldInput, {
                colClass: field.colClass || "col-md-4",
                id: fieldId(field.reqField.toLowerCase()),
                label: field.label,
                type: field.type || "text",
                step: field.step || "",
                reqField: field.reqField,
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
