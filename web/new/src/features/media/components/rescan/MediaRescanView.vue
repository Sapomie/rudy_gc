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
                  <div class="next02-media-eyebrow">Media Audit</div>
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

          <div
            id="msgArea"
            class="alert fade show small"
            :class="messageOk ? 'alert-success' : 'alert-danger'"
            :style="{ display: message ? 'block' : 'none' }"
          >
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
                    <span class="next02-media-live" :class="{ 'is-running': actionLoading }">{{
                      actionLoading ? "执行中" : "就绪"
                    }}</span>
                  </div>
                  <div class="next02-media-action-stack next02-rescan-action-stack">
                    <button
                      type="button"
                      id="btnSelectAll"
                      class="btn btn-outline-secondary"
                      :disabled="actionLoading"
                      @click="selectAll"
                    >
                      全选
                    </button>
                    <button
                      type="button"
                      id="btnClearAll"
                      class="btn btn-outline-secondary"
                      :disabled="actionLoading"
                      @click="clearAll"
                    >
                      清空
                    </button>
                    <button
                      type="button"
                      id="btnAudit"
                      class="btn btn-primary"
                      :disabled="actionLoading"
                      @click="runRescan"
                    >
                      开始检查
                    </button>
                  </div>
                  <div class="next02-media-help">
                    检查过程只读取数据库和磁盘；基目录已选但未选二级目录时，表示检查整个所选基目录。
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
                    <div v-if="!rootOptions.length" class="text-muted">未配置 JAV 或 FC2 媒体 root</div>
                    <div
                      v-for="option in rootOptions"
                      :key="optionKey(option)"
                      class="root-scope-item"
                      :data-root="rootOf(option)"
                      :data-media-type="mediaTypeOf(option)"
                    >
                      <div class="fw-semibold mb-2">
                        <span class="badge text-bg-secondary me-2">{{ mediaTypeLabel(option) }}</span>
                        <code>{{ rootOf(option) }}</code>
                      </div>
                      <div class="root-base-list">
                        <div v-for="scope in scopesOf(option)" :key="scopeKey(option, scope)" class="root-base-item">
                          <label class="root-checkbox-item">
                            <input
                              v-model="selectedScopes"
                              type="checkbox"
                              class="form-check-input root-checkbox"
                              :value="scopeKey(option, scope)"
                              :data-media-type="mediaTypeOf(option)"
                              :data-root="rootOf(option)"
                              :data-scope="scopeValue(scope)"
                              :disabled="actionLoading"
                            />
                            <span>
                              <code>{{ scopeLabel(scope) }}</code>
                              <span class="text-muted small ms-1">{{
                                scopeExists(scope) ? `扫描基目录：${scopeBaseDir(scope)}` : "当前未挂载"
                              }}</span>
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
                                  :data-media-type="mediaTypeOf(option)"
                                  :data-root="rootOf(option)"
                                  :data-scope="scopeValue(scope)"
                                  :data-branch="childName(child)"
                                  :disabled="actionLoading || !isScopeSelected(option, scope)"
                                />
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
                  <div class="row g-2 mt-1">
                    <div v-for="item in typeSummaryItems" :key="item.id" class="col-12">
                      <div class="overview-item">
                        <div class="label">{{ item.label }}</div>
                        <div class="small mt-1">
                          活动 {{ num(item.stats, "active_records") }} · 磁盘 {{ num(item.stats, "disk_files") }} · 正常
                          {{ num(item.stats, "healthy") }} · 异常 {{ typeAnomalyCount(item.stats) }}
                        </div>
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
              <ResultTable
                v-for="table in tables"
                :key="table.id"
                :table="table"
                card-class="next02-rescan-result-card"
                :cell-value="cellValue"
              />
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
import { apiPathForRoute, getJSON, sendJSON } from "@/features/media/api";
import ResultList from "@/features/media/components/rescan/RescanResultList.vue";
import PageShell from "@/shared/components/PageShell.vue";
import ResultTable from "@/shared/components/table/ResultTable.vue";
import type { MediaRecord } from "@/features/media/types";
import { asArray, asRecord, formatValue, num, pick, text } from "@/shared/utils/format";

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
  rows: MediaRecord[];
};

const route = useRoute();
const payload = ref<MediaRecord>({});
const result = ref<MediaRecord>({});
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

const title = computed(() => text(payload.value, "Title") || "媒体一致性检查");
const pageTitle = computed(() => text(payload.value, "PageTitle") || title.value);
const pageNote = computed(() => text(payload.value, "PageNote"));
const rootOptions = computed(() => asArray<MediaRecord>(pick(payload.value, "RootOptions")));
const statusText = computed(() => buildStatusText());
const summaryItems = computed(() => [
  { id: "sumActiveRecords", label: "活动记录", value: num(result.value, "active_records"), className: "" },
  { id: "sumDiskFiles", label: "磁盘文件", value: num(result.value, "disk_files"), className: "" },
  { id: "sumHealthy", label: "正常关联", value: num(result.value, "healthy"), className: "text-success" },
  { id: "sumMissing", label: "文件缺失", value: num(result.value, "missing"), className: "text-danger" },
  { id: "sumUntracked", label: "未关联", value: num(result.value, "untracked"), className: "text-warning" },
  { id: "sumSizeMismatch", label: "大小异常", value: num(result.value, "size_mismatch"), className: "text-warning" },
  { id: "sumDuplicateActive", label: "重复活动", value: num(result.value, "duplicate_active"), className: "text-danger" },
  { id: "sumSkippedRoots", label: "跳过范围", value: listValue("skipped_targets").length, className: "text-muted" },
  { id: "sumErrors", label: "错误", value: num(result.value, "errors"), className: "text-danger" }
]);
const typeSummaryItems = computed(() => [
  { id: "jav", label: "JAV", stats: asRecord(pick(result.value, "jav_stats")) },
  { id: "fc2", label: "FC2", stats: asRecord(pick(result.value, "fc2_stats")) }
]);
const tables = computed<ResultTableDef[]>(() => [
  {
    id: "tblMissing",
    title: "数据库记录对应文件缺失",
    columns: [
      { label: "类型", key: "media_type", width: "8%" },
      { label: "影片标识", key: "identity", width: "20%" },
      { label: "数据库大小", key: "database_size", width: "14%" },
      { label: "路径", key: "path", width: "58%", cls: "mono" }
    ],
    rows: asArray<MediaRecord>(pick(result.value, "missing_items"))
  },
  {
    id: "tblUntracked",
    title: "磁盘未关联文件",
    columns: [
      { label: "类型", key: "media_type", width: "8%" },
      { label: "文件名", key: "file_name", width: "22%" },
      { label: "文件大小", key: "disk_size", width: "16%" },
      { label: "路径", key: "path", width: "54%", cls: "mono" }
    ],
    rows: asArray<MediaRecord>(pick(result.value, "untracked_items"))
  },
  {
    id: "tblSizeMismatch",
    title: "文件大小不一致",
    columns: [
      { label: "类型", key: "media_type", width: "8%" },
      { label: "影片标识", key: "identity", width: "18%" },
      { label: "数据库大小", key: "database_size", width: "16%" },
      { label: "磁盘大小", key: "disk_size", width: "16%" },
      { label: "路径", key: "path", width: "42%", cls: "mono" }
    ],
    rows: asArray<MediaRecord>(pick(result.value, "size_mismatch_items"))
  },
  {
    id: "tblDuplicateActive",
    title: "重复活动记录",
    columns: [
      { label: "类型", key: "media_type", width: "8%" },
      { label: "重复键", key: "duplicate_key", width: "22%" },
      { label: "记录 ID", key: "record_id", width: "12%" },
      { label: "影片标识", key: "identity", width: "18%" },
      { label: "路径", key: "path", width: "40%", cls: "mono" }
    ],
    rows: asArray<MediaRecord>(pick(result.value, "duplicate_items"))
  },
  {
    id: "tblErrors",
    title: "错误",
    columns: [
      { label: "类型", key: "media_type", width: "8%" },
      { label: "文件/影片", key: "identity", width: "18%" },
      { label: "路径", key: "path", width: "38%", cls: "mono" },
      { label: "错误", key: "error", width: "36%" }
    ],
    rows: asArray<MediaRecord>(pick(result.value, "error_items"))
  }
]);

function rootOf(option: MediaRecord): string {
  return text(option, "root", "Root");
}

function mediaTypeOf(option: MediaRecord): string {
  return text(option, "media_type", "MediaType").toLowerCase();
}

function mediaTypeLabel(option: MediaRecord): string {
  return mediaTypeOf(option).toUpperCase();
}

function optionKey(option: MediaRecord): string {
  return `${mediaTypeOf(option)}||${rootOf(option)}`;
}

function scopesOf(option: MediaRecord): MediaRecord[] {
  return asArray<MediaRecord>(pick(option, "scopes", "Scopes"));
}

function scopeValue(scope: MediaRecord): string {
  return text(scope, "scope", "Scope");
}

function scopeLabel(scope: MediaRecord): string {
  return text(scope, "label", "Label");
}

function scopeBaseDir(scope: MediaRecord): string {
  return text(scope, "base_dir", "BaseDir");
}

function scopeExists(scope: MediaRecord): boolean {
  return Boolean(pick(scope, "exists", "Exists"));
}

function childrenOf(scope: MediaRecord): MediaRecord[] {
  return asArray<MediaRecord>(pick(scope, "children", "Children"));
}

function childName(child: MediaRecord): string {
  return text(child, "name", "Name");
}

function scopeKey(option: MediaRecord, scope: MediaRecord): string {
  return `${optionKey(option)}||${scopeValue(scope)}`;
}

function branchKey(option: MediaRecord, scope: MediaRecord, child: MediaRecord): string {
  return `${scopeKey(option, scope)}||${childName(child)}`;
}

function isScopeSelected(option: MediaRecord, scope: MediaRecord): boolean {
  return selectedScopes.value.includes(scopeKey(option, scope));
}

function allScopeKeys(): string[] {
  return rootOptions.value.flatMap((option) => scopesOf(option).map((scope) => scopeKey(option, scope)));
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
  return selectedScopes.value
    .map((key) => {
      const [media_type, root, scope] = key.split("||");
      const branches = selectedBranches.value
        .filter((branch) => branch.startsWith(`${key}||`))
        .map((branch) => branch.split("||").slice(3).join("||"))
        .filter((name) => name !== "");
      return { media_type, root, scope, branches };
    })
    .filter((item) => item.media_type && item.root && item.scope);
}

function typeAnomalyCount(stats: MediaRecord): number {
  return (
    num(stats, "missing") +
    num(stats, "untracked") +
    num(stats, "size_mismatch") +
    num(stats, "duplicate_active") +
    num(stats, "errors")
  );
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
    const anomalyCount =
      num(data, "missing") +
      num(data, "untracked") +
      num(data, "size_mismatch") +
      num(data, "duplicate_active") +
      num(data, "errors");
    message.value = anomalyCount ? `检查完成，发现 ${anomalyCount} 项异常` : "检查完成，媒体数据一致";
    messageOk.value = true;
  } catch (err) {
    message.value = err instanceof Error ? err.message : "媒体一致性检查失败";
    messageOk.value = false;
  } finally {
    actionLoading.value = false;
  }
}

function listValue(key: string): string[] {
  return asArray<unknown>(pick(result.value, key)).map((item) => String(item));
}

function buildStatusText(): string {
  const selectedRoots = listValue("selected_roots");
  const scannedTargets = listValue("scanned_targets");
  const skippedTargets = listValue("skipped_targets");
  if (!selectedRoots.length) return "等待操作";
  if (!scannedTargets.length && skippedTargets.length > 0) return "所选范围当前都不存在，已全部跳过";
  const anomalyCount =
    num(result.value, "missing") +
    num(result.value, "untracked") +
    num(result.value, "size_mismatch") +
    num(result.value, "duplicate_active") +
    num(result.value, "errors");
  return `检查完成：检查 ${scannedTargets.length} 个范围，跳过 ${skippedTargets.length} 个范围，发现 ${anomalyCount} 项异常`;
}

function cellValue(row: MediaRecord, column: ResultColumn): string {
  const value = pick(row, column.key) ?? pick(row, "movie_jav_id") ?? pick(row, "file_name");
  return formatValue(value);
}
</script>
