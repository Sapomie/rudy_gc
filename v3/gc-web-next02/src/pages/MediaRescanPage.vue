<template>
  <PageShell class="media-ingest-page-shell media-rescan-page-shell" :title="title" eyebrow="">
    <div id="page-root" class="container-fluid media-ingest-page-root media-rescan-page-root next02-sc-pink-theme">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div id="mediaRescanPage" class="next02-media-rewrite next02-rescan-rewrite">
          <div class="row g-0">
            <div class="col-12">
              <section class="next02-media-hero next02-sc-movie-top-shell" aria-labelledby="mediaRescanTitle">
                <div class="next02-media-hero-main">
                  <div class="next02-media-eyebrow">Media Rescan</div>
                  <h4 id="mediaRescanTitle">{{ pageTitle }}</h4>
                  <p>{{ pageNote }}</p>
                </div>
                <div class="next02-media-hero-actions">
                  <a href="/mediaingest" class="btn btn-outline-primary btn-sm">媒体入库</a>
                  <a href="/mediarollback" class="btn btn-outline-secondary btn-sm">名称回滚</a>
                </div>
              </section>
            </div>
          </div>

          <div id="msgArea" class="alert fade show small" :class="messageOk ? 'alert-success' : 'alert-danger'" :style="{ display: message ? 'block' : 'none' }">
            {{ message }}
          </div>

          <div class="row next02-media-grid">
            <div class="col-lg-4 next02-media-control-col">
              <div class="card next02-media-action-card next02-rescan-action-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Selection</div>
                      <h5 class="card-title mb-0">操作</h5>
                    </div>
                    <span class="next02-media-live" :class="{ 'is-running': actionLoading }">{{ actionLoading ? "执行中" : "就绪" }}</span>
                  </div>
                  <div class="next02-media-action-stack next02-rescan-action-stack">
                    <button type="button" id="btnSelectAll" class="btn btn-outline-secondary" :disabled="actionLoading" @click="selectAll">全选</button>
                    <button type="button" id="btnClearAll" class="btn btn-outline-secondary" :disabled="actionLoading" @click="clearAll">清空</button>
                    <button type="button" id="btnRescan" class="btn btn-primary" :disabled="actionLoading" @click="runRescan">开始重扫</button>
                  </div>
                  <div class="next02-media-help">
                    只会对本次勾选且当前存在的扫描范围做删除判断；基目录已选但未选二级目录时，表示扫描整个所选基目录。
                  </div>
                </div>
              </div>

              <div class="card next02-media-root-card next02-rescan-scope-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Roots</div>
                      <h5 class="card-title mb-0">扫描目录</h5>
                    </div>
                    <span class="next02-media-count-pill">{{ selectedScopes.length }} 个范围</span>
                  </div>
                  <div class="root-scope-list">
                    <div v-if="!rootOptions.length" class="text-muted">未配置 media.rootDirs</div>
                    <div v-for="option in rootOptions" :key="rootOf(option)" class="root-scope-item" :data-root="rootOf(option)">
                      <div class="fw-semibold mb-2"><code>{{ rootOf(option) }}</code></div>
                      <div class="root-base-list">
                        <div v-for="scope in scopesOf(option)" :key="scopeKey(option, scope)" class="root-base-item">
                          <label class="root-checkbox-item">
                            <input
                              v-model="selectedScopes"
                              type="checkbox"
                              class="form-check-input root-checkbox"
                              :value="scopeKey(option, scope)"
                              :data-root="rootOf(option)"
                              :data-scope="scopeValue(scope)"
                              :disabled="actionLoading"
                            >
                            <span>
                              <code>{{ scopeLabel(scope) }}</code>
                              <span class="text-muted small ms-1">{{ scopeExists(scope) ? `扫描基目录：${scopeBaseDir(scope)}` : "当前未挂载" }}</span>
                            </span>
                          </label>
                          <div class="root-branch-list">
                            <div class="text-muted small mb-1">二级目录：不选则扫描整个 {{ scopeLabel(scope) }}</div>
                            <template v-if="childrenOf(scope).length">
                              <label
                                v-for="child in childrenOf(scope)"
                                :key="branchKey(option, scope, child)"
                                class="branch-checkbox-item"
                                :class="{ 'is-disabled': !isScopeSelected(option, scope) }"
                              >
                                <input
                                  v-model="selectedBranches"
                                  type="checkbox"
                                  class="form-check-input branch-checkbox"
                                  :value="branchKey(option, scope, child)"
                                  :data-root="rootOf(option)"
                                  :data-scope="scopeValue(scope)"
                                  :data-branch="childName(child)"
                                  :disabled="actionLoading || !isScopeSelected(option, scope)"
                                >
                                <code>{{ childName(child) }}</code>
                              </label>
                            </template>
                            <div v-else class="text-muted small">暂无可选二级目录</div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="card next02-media-overview-card next02-rescan-overview-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Overview</div>
                      <h5 class="card-title mb-0">运行概览</h5>
                    </div>
                  </div>
                  <div class="row g-2">
                    <div v-for="item in summaryItems" :key="item.id" class="col-6">
                      <div class="overview-item">
                        <div class="label">{{ item.label }}</div>
                        <div class="value" :class="item.className" :id="item.id">{{ item.value }}</div>
                      </div>
                    </div>
                  </div>
                  <div class="mt-3">
                    <div class="text-muted small">状态</div>
                    <div id="statusText" class="fw-semibold">{{ statusText }}</div>
                  </div>
                </div>
              </div>

              <div class="card next02-media-root-card next02-rescan-list-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Scope Result</div>
                      <h5 class="card-title mb-0">本次目录结果</h5>
                    </div>
                  </div>
                  <ResultList id="selectedRootList" title="已选择 root" :items="listValue('selected_roots')" />
                  <ResultList id="scannedRootList" title="实际扫描 root" :items="listValue('scanned_roots')" />
                  <ResultList id="skippedRootList" title="已跳过 root" :items="listValue('skipped_roots')" />
                  <ResultList id="selectedTargetList" title="已选择范围" :items="listValue('selected_targets')" />
                  <ResultList id="scannedTargetList" title="实际扫描范围" :items="listValue('scanned_targets')" />
                  <ResultList id="skippedTargetList" title="已跳过范围" :items="listValue('skipped_targets')" />
                </div>
              </div>
            </div>

            <div class="col-lg-8 next02-media-result-col">
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
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, formatValue, num, pick, text } from "@/utils/format";

type ResultColumn = {
  label: string;
  key: string;
  width: string;
  cls?: string;
};

type ResultTableDef = {
  id: string;
  title: string;
  columns: ResultColumn[];
  rows: AnyRecord[];
};

const route = useRoute();
const payload = ref<AnyRecord>({});
const result = ref<AnyRecord>({});
const loading = ref(false);
const actionLoading = ref(false);
const error = ref("");
const message = ref("");
const messageOk = ref(true);
const selectedScopes = ref<string[]>([]);
const selectedBranches = ref<string[]>([]);

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  message.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
    initializeSelections();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title") || "媒体重扫");
const pageTitle = computed(() => text(payload.value, "PageTitle") || title.value);
const pageNote = computed(() => text(payload.value, "PageNote"));
const rootOptions = computed(() => asArray<AnyRecord>(pick(payload.value, "RootOptions")));
const statusText = computed(() => buildStatusText(result.value));
const summaryItems = computed(() => [
  { id: "sumTotalFiles", label: "扫描文件", value: num(result.value, "total_files"), className: "" },
  { id: "sumMatched", label: "匹配成功", value: num(result.value, "matched"), className: "" },
  { id: "sumMoved", label: "位置变更", value: num(result.value, "moved"), className: "text-primary" },
  { id: "sumRestored", label: "恢复记录", value: num(result.value, "restored"), className: "text-success" },
  { id: "sumRemoved", label: "标记删除", value: num(result.value, "marked_removed"), className: "text-warning" },
  { id: "sumUnmatched", label: "未匹配", value: num(result.value, "unmatched"), className: "text-secondary" },
  { id: "sumSkippedRoots", label: "跳过范围", value: listValue("skipped_targets").length, className: "text-muted" },
  { id: "sumErrors", label: "错误", value: num(result.value, "errors"), className: "text-danger" }
]);
const tables = computed<ResultTableDef[]>(() => [
  {
    id: "tblMoved",
    title: "位置变更",
    columns: [
      { label: "影片", key: "movie_name", width: "16%" },
      { label: "匹配方式", key: "match_by", width: "16%" },
      { label: "旧路径", key: "previous_path", width: "34%", cls: "mono" },
      { label: "新路径", key: "path", width: "34%", cls: "mono" }
    ],
    rows: asArray<AnyRecord>(pick(result.value, "moved_items"))
  },
  {
    id: "tblRestored",
    title: "恢复记录",
    columns: [
      { label: "影片", key: "movie_name", width: "16%" },
      { label: "匹配方式", key: "match_by", width: "16%" },
      { label: "旧路径", key: "previous_path", width: "34%", cls: "mono" },
      { label: "当前路径", key: "path", width: "34%", cls: "mono" }
    ],
    rows: asArray<AnyRecord>(pick(result.value, "restored_items"))
  },
  {
    id: "tblRemoved",
    title: "标记删除",
    columns: [
      { label: "影片", key: "movie_name", width: "16%" },
      { label: "原 root", key: "root_dir", width: "26%", cls: "mono" },
      { label: "原路径", key: "path", width: "58%", cls: "mono" }
    ],
    rows: asArray<AnyRecord>(pick(result.value, "removed_items"))
  },
  {
    id: "tblUnmatched",
    title: "未匹配文件",
    columns: [
      { label: "文件名", key: "file_name", width: "20%" },
      { label: "root", key: "root_dir", width: "26%", cls: "mono" },
      { label: "路径", key: "path", width: "54%", cls: "mono" }
    ],
    rows: asArray<AnyRecord>(pick(result.value, "unmatched_items"))
  },
  {
    id: "tblErrors",
    title: "错误",
    columns: [
      { label: "文件/影片", key: "movie_name", width: "18%" },
      { label: "路径", key: "path", width: "42%", cls: "mono" },
      { label: "错误", key: "error", width: "40%" }
    ],
    rows: asArray<AnyRecord>(pick(result.value, "error_items"))
  }
]);

function rootOf(option: AnyRecord): string {
  return text(option, "root", "Root");
}

function scopesOf(option: AnyRecord): AnyRecord[] {
  return asArray<AnyRecord>(pick(option, "scopes", "Scopes"));
}

function scopeValue(scope: AnyRecord): string {
  return text(scope, "scope", "Scope");
}

function scopeLabel(scope: AnyRecord): string {
  return text(scope, "label", "Label");
}

function scopeBaseDir(scope: AnyRecord): string {
  return text(scope, "base_dir", "BaseDir");
}

function scopeExists(scope: AnyRecord): boolean {
  return Boolean(pick(scope, "exists", "Exists"));
}

function childrenOf(scope: AnyRecord): AnyRecord[] {
  return asArray<AnyRecord>(pick(scope, "children", "Children"));
}

function childName(child: AnyRecord): string {
  return text(child, "name", "Name");
}

function scopeKey(option: AnyRecord, scope: AnyRecord): string {
  return `${rootOf(option)}||${scopeValue(scope)}`;
}

function branchKey(option: AnyRecord, scope: AnyRecord, child: AnyRecord): string {
  return `${scopeKey(option, scope)}||${childName(child)}`;
}

function isScopeSelected(option: AnyRecord, scope: AnyRecord): boolean {
  return selectedScopes.value.includes(scopeKey(option, scope));
}

function allScopeKeys(): string[] {
  return rootOptions.value.flatMap((option) => scopesOf(option).map((scope) => scopeKey(option, scope)));
}

function allBranchKeys(): string[] {
  return rootOptions.value.flatMap((option) => scopesOf(option).flatMap((scope) => childrenOf(scope).map((child) => branchKey(option, scope, child))));
}

function initializeSelections() {
  selectedScopes.value = allScopeKeys();
  selectedBranches.value = [];
}

function selectAll() {
  selectedScopes.value = allScopeKeys();
  selectedBranches.value = [];
}

function clearAll() {
  selectedScopes.value = [];
  selectedBranches.value = [];
}

function buildSelections() {
  return selectedScopes.value.map((key) => {
    const [root, scope] = key.split("||");
    const branches = selectedBranches.value
      .filter((branch) => branch.startsWith(`${key}||`))
      .map((branch) => branch.split("||").slice(2).join("||"))
      .filter((name) => name !== "");
    return { root, scope, branches };
  }).filter((item) => item.root && item.scope);
}

async function runRescan() {
  const selections = buildSelections();
  if (!selections.length) {
    message.value = "至少选择一个扫描目录";
    messageOk.value = false;
    return;
  }
  actionLoading.value = true;
  message.value = "";
  try {
    const data = asRecord(await sendJSON("/api/triggers/media/rescan", "POST", { selections }));
    result.value = data;
    message.value = "媒体重扫完成";
    messageOk.value = true;
  } catch (err) {
    message.value = err instanceof Error ? err.message : "媒体重扫失败";
    messageOk.value = false;
  } finally {
    actionLoading.value = false;
  }
}

function listValue(key: string): string[] {
  return asArray<unknown>(pick(result.value, key)).map((item) => String(item));
}

function buildStatusText(row: AnyRecord): string {
  const selectedRoots = listValue("selected_roots");
  const scannedTargets = listValue("scanned_targets");
  const skippedTargets = listValue("skipped_targets");
  if (!selectedRoots.length) return "等待操作";
  if (!scannedTargets.length && skippedTargets.length > 0) return "所选范围当前都不存在，已全部跳过";
  return `重扫完成：扫描 ${scannedTargets.length} 个范围，跳过 ${skippedTargets.length} 个范围`;
}

function cellValue(row: AnyRecord, column: ResultColumn): string {
  const value = pick(row, column.key) ?? pick(row, "movie_jav_id") ?? pick(row, "file_name");
  return formatValue(value);
}

const ResultList = defineComponent({
  name: "MediaRescanResultList",
  props: {
    id: { type: String, required: true },
    title: { type: String, required: true },
    items: { type: Array as PropType<string[]>, required: true }
  },
  setup(props) {
    return () => h("div", { class: "next02-rescan-result-list" }, [
      h("div", { class: "next02-rescan-result-title" }, props.title),
      h("ul", { class: "root-list mb-0", id: props.id }, props.items.length
        ? props.items.map((item) => h("li", [h("code", item)]))
        : [h("li", { class: "text-muted" }, "暂无数据")])
    ]);
  }
});

const ResultTable = defineComponent({
  name: "MediaRescanResultTable",
  props: {
    table: { type: Object as PropType<ResultTableDef>, required: true }
  },
  setup(props) {
    return () => h("div", { class: "card next02-media-result-card next02-rescan-result-card" }, [
      h("div", { class: "card-body" }, [
        h("div", { class: "next02-media-panel-head" }, [
          h("div", [
            h("div", { class: "next02-media-eyebrow" }, "Result"),
            h("h5", { class: "card-title mb-0" }, props.table.title)
          ]),
          h("span", { class: "next02-media-count-pill" }, `${props.table.rows.length} 行`)
        ]),
        h("div", { class: "table-responsive" }, [
          h("table", { class: "table table-hover align-middle mb-0 next02-media-table", id: props.table.id }, [
            h("colgroup", props.table.columns.map((column) => h("col", { style: { width: column.width } }))),
            h("thead", [
              h("tr", props.table.columns.map((column) => h("th", column.label)))
            ]),
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
