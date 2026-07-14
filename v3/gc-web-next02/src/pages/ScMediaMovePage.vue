<template>
  <PageShell class="sc-media-move-page-shell" :title="title" eyebrow="">
    <div id="page-root" class="container-fluid sc-media-move-page-root next02-sc-pink-theme next02-sc-media-rewrite">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div id="scMediaMovePage">
          <section class="next02-sc-hero next02-sc-media-hero next02-sc-movie-top-shell" aria-labelledby="scMediaMoveTitle">
            <div class="next02-sc-hero-copy">
              <span class="next02-sc-eyebrow">SC / MEDIA PIPELINE</span>
              <h4 id="scMediaMoveTitle">{{ pageTitle }}</h4>
              <p>{{ pageNote }}</p>
            </div>
            <div class="next02-sc-hero-actions">
              <a :href="navHref" class="btn btn-outline-primary btn-sm">{{ navLabel }}</a>
              <a href="/mediaingest" class="btn btn-outline-secondary btn-sm">媒体入库</a>
            </div>
          </section>

          <div id="msgArea" class="alert fade show small" :class="messageOk ? 'alert-success' : 'alert-danger'" :style="{ display: message ? 'block' : 'none' }">
            {{ message }}
          </div>

          <div class="row next02-sc-media-workspace">
            <div class="col-lg-4 next02-sc-media-control-col">
              <div class="card next02-sc-panel next02-sc-scope-panel">
                <div class="card-body">
                  <div class="next02-sc-panel-head">
                    <div>
                      <span class="next02-sc-panel-kicker">SCOPE</span>
                      <h5 class="card-title mb-0">执行范围</h5>
                    </div>
                    <span class="next02-sc-status-dot">{{ mode === 'all' ? '全部 SC' : '单个 SC' }}</span>
                  </div>
                  <div class="d-flex flex-wrap gap-3 mb-3">
                    <div class="form-check">
                      <input v-model="mode" class="form-check-input" type="radio" name="scopeMode" id="scopeModeSingle" value="single" :disabled="actionLoading">
                      <label class="form-check-label" for="scopeModeSingle">单个 SC</label>
                    </div>
                    <div class="form-check">
                      <input v-model="mode" class="form-check-input" type="radio" name="scopeMode" id="scopeModeAll" value="all" :disabled="actionLoading">
                      <label class="form-check-label" for="scopeModeAll">全部符合条件</label>
                    </div>
                  </div>

                  <div id="scNameGroup" :class="{ 'd-none': mode === 'all' }">
                    <label class="form-label" for="scNameInput">SC 名称</label>
                    <input v-model="scName" type="text" class="form-control" id="scNameInput" placeholder="例如 2026-04-01-22-30" :disabled="actionLoading">
                    <div class="form-text">可以从 SC 详情页进入，也可以手动输入后执行第一段预处理。</div>
                  </div>
                  <div id="allModeHint" class="form-text" :class="{ 'd-none': mode !== 'all' }">
                    将汇总所有 SC 事件中 <code>is_sc=2</code> 的电影，按影片去重后统一预处理与移动。
                  </div>
                </div>
              </div>

              <div class="card next02-sc-panel next02-sc-action-panel">
                <div class="card-body">
                  <div class="next02-sc-panel-head">
                    <div><span class="next02-sc-panel-kicker">PIPELINE</span><h5 class="card-title mb-0">两段式操作</h5></div>
                  </div>
                  <div class="d-grid gap-2">
                    <button type="button" id="btnPrecheck" class="btn btn-primary" :disabled="actionLoading" @click="runPrecheck">第一段：预处理</button>
                    <button type="button" id="btnCommit" class="btn btn-success" :disabled="actionLoading || !canCommit" :data-can-commit="canCommit ? '1' : '0'" @click="runCommit">第二段：执行移动</button>
                    <button type="button" id="btnReturn" class="btn btn-outline-secondary" :class="{ 'd-none': !hasPlan }" :disabled="actionLoading" @click="runReturn">返回（清空本次预处理）</button>
                  </div>
                  <div class="mt-3 small text-muted">
                    <div>第一段只读取本次范围内涉及到的 WMedia 并预览目标目录。</div>
                    <div>第二段只移动预处理通过项到同 root 的 <code>watched</code>。</div>
                  </div>
                </div>
              </div>

              <div class="card next02-sc-panel next02-sc-overview-panel">
                <div class="card-body">
                  <div class="next02-sc-panel-head">
                    <div><span class="next02-sc-panel-kicker">STATUS</span><h5 class="card-title mb-0">运行概览</h5></div>
                    <span class="next02-sc-status-dot" :class="{ 'is-busy': actionLoading, 'is-ready': canCommit }">{{ actionLoading ? '处理中' : (canCommit ? '可提交' : '等待操作') }}</span>
                  </div>
                  <div class="row g-2">
                    <div class="col-3"><div class="overview-item"><div class="label">总数</div><div class="value" id="sumTotal">{{ num(state, "total") }}</div></div></div>
                    <div class="col-3"><div class="overview-item"><div class="label">可移动</div><div class="value text-success" id="sumPass">{{ num(state, "movable") }}</div></div></div>
                    <div class="col-3"><div class="overview-item"><div class="label">跳过</div><div class="value text-secondary" id="sumSkip">{{ num(state, "skipped") }}</div></div></div>
                    <div class="col-3"><div class="overview-item"><div class="label">失败</div><div class="value text-danger" id="sumFail">{{ num(state, "failed") }}</div></div></div>
                  </div>
                  <div class="mt-3">
                    <div class="text-muted small">状态</div>
                    <div id="statusText" class="fw-semibold">{{ text(state, "message") || "等待操作" }}</div>
                  </div>
                  <div v-if="generatedAt" class="mt-2 text-muted small" id="generatedAtText">预处理时间：<ScDateTime :value="generatedAt" /></div>
                </div>
              </div>

              <div class="card next02-sc-panel next02-sc-roots-panel">
                <div class="card-body">
                  <div class="next02-sc-panel-head"><div><span class="next02-sc-panel-kicker">STORAGE</span><h5 class="card-title mb-0">Root 目录</h5></div></div>
                  <ul class="mb-0 root-list">
                    <li v-for="dir in rootDirs" :key="dir"><code>{{ dir }}</code></li>
                    <li v-if="!rootDirs.length" class="text-muted">未配置 media.rootDirs</li>
                  </ul>
                </div>
              </div>
            </div>

            <div class="col-lg-8 next02-sc-media-result-col">
              <ResultTable v-for="table in tables" :key="table.id" :table="table" />
            </div>
          </div>
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import PageShell from "@/components/PageShell.vue";
import ScDateTime from "@/components/ScDateTime.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, formatValue, num, pick, text } from "@/utils/format";

type ResultColumn = { label: string; key: string; width: string; cls?: string };
type ResultTableDef = { id: string; title: string; columns: ResultColumn[]; rows: AnyRecord[] };

const route = useRoute();
const payload = ref<AnyRecord>({});
const state = ref<AnyRecord>({});
const loading = ref(false);
const actionLoading = ref(false);
const error = ref("");
const message = ref("");
const messageOk = ref(true);
const mode = ref<"single" | "all">("single");
const scName = ref("");

watch(() => route.fullPath, load, { immediate: true });
watch([mode, scName], () => {
  if (mode.value === "all" || scName.value.trim()) void loadPlan();
  else emptySingleState();
});

async function load() {
  loading.value = true;
  error.value = "";
  message.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
    mode.value = text(payload.value, "Mode") === "all" ? "all" : "single";
    scName.value = text(payload.value, "ScName");
    if (mode.value === "all" || scName.value.trim()) await loadPlan();
    else emptySingleState();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title") || "SC Media 移动");
const pageTitle = computed(() => text(payload.value, "PageTitle") || title.value);
const pageNote = computed(() => text(payload.value, "PageNote"));
const rootDirs = computed(() => asArray<string>(pick(payload.value, "RootDirs")));
const navHref = computed(() => text(payload.value, "ScDetailHref") || "/scevents");
const navLabel = computed(() => text(payload.value, "ScDetailHref") ? "返回当前 SC" : "SC Events");
const canCommit = computed(() => Boolean(pick(state.value, "can_commit")));
const hasPlan = computed(() => Boolean(pick(state.value, "has_plan")));
const generatedAt = computed(() => text(state.value, "generated_at"));
const tables = computed<ResultTableDef[]>(() => [
  table("tblPrecheckPass", "预处理通过 Media", ["SC Events", "影片", "源路径", "目标路径"], ["sc_names", "movie_name", "source_path", "target_path"], "precheck_pass"),
  table("tblPrecheckSkip", "预处理跳过 Media", ["SC Events", "影片", "源路径", "原因"], ["sc_names", "movie_name", "source_path", "error"], "precheck_skip"),
  table("tblPrecheckFail", "预处理失败 Media", ["SC Events", "影片", "源路径", "失败原因"], ["sc_names", "movie_name", "source_path", "error"], "precheck_fail"),
  table("tblCommitSuccess", "移动成功 Media", ["SC Events", "影片", "源路径", "目标路径"], ["sc_names", "movie_name", "source_path", "target_path"], "commit_success"),
  table("tblCommitFail", "移动失败 Media", ["SC Events", "影片", "源路径", "失败原因"], ["sc_names", "movie_name", "source_path", "error"], "commit_fail")
]);

function table(id: string, title: string, labels: string[], keys: string[], rowKey: string): ResultTableDef {
  const widths = ["18%", "18%", "32%", "32%"];
  return {
    id,
    title,
    columns: labels.map((label, index) => ({ label, key: keys[index], width: widths[index], cls: keys[index].includes("path") ? "mono" : "" })),
    rows: asArray<AnyRecord>(pick(state.value, rowKey))
  };
}

function emptySingleState() {
  state.value = {
    mode: "single",
    message: "等待操作",
    has_plan: false,
    can_commit: false,
    precheck_pass: [],
    precheck_skip: [],
    precheck_fail: [],
    commit_success: [],
    commit_fail: []
  };
}

function requestBody() {
  return { mode: mode.value, sc_name: scName.value.trim() };
}

async function loadPlan() {
  if (mode.value === "single" && !scName.value.trim()) {
    emptySingleState();
    return;
  }
  const qs = new URLSearchParams(requestBody()).toString();
  state.value = asRecord(await getJSON(`/api/triggers/sc-media/plan?${qs}`));
}

async function runAction(path: string, successMessage: string) {
  if (mode.value === "single" && !scName.value.trim()) {
    message.value = "请先输入 SC 名称";
    messageOk.value = false;
    return;
  }
  actionLoading.value = true;
  message.value = "";
  try {
    const data = asRecord(await sendJSON(path, "POST", requestBody()));
    state.value = data;
    message.value = text(data, "message") || successMessage;
    messageOk.value = true;
  } catch (err) {
    message.value = err instanceof Error ? err.message : "请求失败";
    messageOk.value = false;
  } finally {
    actionLoading.value = false;
  }
}

function runPrecheck() {
  void runAction("/api/triggers/sc-media/precheck", "预处理完成");
}

function runCommit() {
  void runAction("/api/triggers/sc-media/commit", "第二段执行完成");
}

function runReturn() {
  void runAction("/api/triggers/sc-media/return", "已返回");
}

function cellValue(row: AnyRecord, column: ResultColumn): string {
  const value = pick(row, column.key);
  return Array.isArray(value) ? value.map((item) => String(item)).join("\n") : formatValue(value);
}

const ResultTable = defineComponent({
  name: "ScMediaMoveResultTable",
  props: { table: { type: Object as PropType<ResultTableDef>, required: true } },
  setup(props) {
    return () => h("div", { class: "card next02-sc-panel next02-sc-result-panel" }, [
      h("div", { class: "card-body" }, [
        h("div", { class: "next02-sc-panel-head" }, [
          h("div", [h("span", { class: "next02-sc-panel-kicker" }, "RESULT"), h("h5", { class: "card-title mb-0" }, props.table.title)]),
          h("span", { class: "next02-sc-result-count" }, `${props.table.rows.length} 项`)
        ]),
        h("div", { class: "table-responsive" }, [
          h("table", { class: "table table-hover align-middle mb-0 next02-sc-media-table", id: props.table.id }, [
            h("colgroup", props.table.columns.map((column) => h("col", { style: { width: column.width } }))),
            h("thead", [h("tr", props.table.columns.map((column) => h("th", column.label)))]),
            h("tbody", props.table.rows.length
              ? props.table.rows.map((row) => h("tr", props.table.columns.map((column) => {
                  const value = cellValue(row, column);
                  return h("td", { class: column.cls || "", title: value }, value);
                })))
              : [h("tr", [h("td", { colspan: props.table.columns.length, class: "text-center text-muted" }, "暂无数据")])])
          ])
        ])
      ])
    ]);
  }
});
</script>
