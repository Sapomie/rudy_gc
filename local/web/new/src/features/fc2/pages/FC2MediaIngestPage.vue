<template>
  <PageShell class="media-ingest-page-shell" :title="title" eyebrow="">
    <div id="page-root" class="container-fluid media-ingest-page-root next02-sc-pink-theme">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div id="mediaIngestPage" class="next02-media-rewrite">
          <div class="row g-0">
            <div class="col-12">
              <section class="next02-media-hero next02-sc-movie-top-shell" aria-labelledby="mediaIngestTitle">
                <div class="next02-media-hero-main">
                  <div class="next02-media-eyebrow">FC2 Media Intake</div>
                  <h4 id="mediaIngestTitle">{{ pageTitle }}</h4>
                  <p>{{ pageNote }}</p>
                </div>
                <div class="next02-media-hero-actions">
                  <a href="/fclist" class="btn btn-outline-primary btn-sm">FC2影片</a>
                  <a href="/shmagnetlist?movie_type=fc2ppv" class="btn btn-outline-secondary btn-sm">SH磁力</a>
                  <a href="/mediaingest" class="btn btn-outline-secondary btn-sm">WMedia入库</a>
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
              <div class="card next02-media-action-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Pipeline</div>
                      <h5 class="card-title mb-0">操作</h5>
                    </div>
                    <span class="next02-media-live" :class="{ 'is-running': actionLoading }">{{
                      actionLoading ? "执行中" : "就绪"
                    }}</span>
                  </div>
                  <div class="next02-media-action-stack">
                    <button
                      type="button"
                      id="btnPrecheck"
                      class="btn btn-primary"
                      :disabled="actionLoading"
                      @click="runPrecheck"
                    >
                      第一段：预处理
                    </button>
                    <button
                      type="button"
                      id="btnCommit"
                      class="btn btn-success"
                      :disabled="actionLoading || !canCommit"
                      :data-can-commit="canCommit ? '1' : '0'"
                      @click="runCommit"
                    >
                      {{ commitButtonText }}
                    </button>
                    <button
                      type="button"
                      id="btnReturn"
                      class="btn btn-outline-secondary"
                      :class="{ 'd-none': !showReturn }"
                      :disabled="actionLoading"
                      @click="runReturn"
                    >
                      返回（清空本次预处理）
                    </button>
                  </div>
                  <div class="next02-media-help">
                    <div>部分失败时可选择：</div>
                    <div>1. 仅入库通过项（点第二段）</div>
                    <div>2. 返回后修正再重跑预处理</div>
                  </div>
                </div>
              </div>

              <div class="card next02-media-overview-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Overview</div>
                      <h5 class="card-title mb-0">运行概览</h5>
                    </div>
                  </div>
                  <div class="row g-2">
                    <div class="col-4">
                      <div class="overview-item">
                        <div class="label">总数</div>
                        <div class="value" id="sumTotal">{{ summary.total }}</div>
                      </div>
                    </div>
                    <div class="col-4">
                      <div class="overview-item">
                        <div class="label" id="sumPassLabel">{{ summary.passLabel }}</div>
                        <div class="value text-success" id="sumPass">{{ summary.success }}</div>
                      </div>
                    </div>
                    <div class="col-4">
                      <div class="overview-item">
                        <div class="label" id="sumFailLabel">失败</div>
                        <div class="value text-danger" id="sumFail">{{ summary.failed }}</div>
                      </div>
                    </div>
                  </div>
                  <div class="mt-3">
                    <div class="text-muted small">状态</div>
                    <div id="statusText" class="fw-semibold">{{ statusText }}</div>
                  </div>
                  <div v-if="generatedAt" class="mt-2 text-muted small" id="generatedAtText">
                    预处理时间：<ScDateTime :value="generatedAt" unix />
                  </div>
                </div>
              </div>

              <div class="card next02-media-root-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Roots</div>
                      <h5 class="card-title mb-0">Root 目录</h5>
                    </div>
                    <span class="next02-media-count-pill">{{ rootDirs.length }} 个</span>
                  </div>
                  <ul class="mb-0 root-list">
                    <li v-for="dir in rootDirs" :key="dir">
                      <code>{{ dir }}</code>
                    </li>
                    <li v-if="!rootDirs.length" class="text-muted">未配置 FC2Media.RootDirs</li>
                  </ul>
                </div>
              </div>

              <div class="card next02-media-favorite-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Sources</div>
                      <h5 class="card-title mb-0">来源磁力（sh_magnet）</h5>
                    </div>
                    <span class="next02-media-count-pill">{{ favoriteRows.length }} 条</span>
                  </div>
                  <div class="table-responsive">
                    <table class="table table-hover align-middle mb-0 next02-media-table next02-media-favorite-table">
                      <colgroup>
                        <col style="width: 50%" />
                        <col style="width: 25%" />
                        <col style="width: 25%" />
                      </colgroup>
                      <thead>
                        <tr>
                          <th>FC2番号</th>
                          <th>来源帖子</th>
                          <th>Info Hash</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(row, index) in favoriteRows" :key="favoriteKey(row, index)">
                          <td :title="text(row, 'MovieName')">{{ text(row, "MovieName") || "-" }}</td>
                          <td :title="text(row, 'SourceRowID')">{{ text(row, "SourceRowID") || "-" }}</td>
                          <td class="mono" :title="text(row, 'InfoHash')">{{ text(row, "InfoHash") || "-" }}</td>
                        </tr>
                        <tr v-if="!favoriteRows.length">
                          <td colspan="3" class="text-center text-muted">暂无下载中来源</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
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
import "@/features/fc2/styles/index.css";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/features/fc2/api";
import PageShell from "@/shared/components/PageShell.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import ResultTable from "@/shared/components/table/ResultTable.vue";
import type { FC2Record } from "@/features/fc2/types";
import { asArray, asRecord, num, pick, text } from "@/shared/utils/format";

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
  rows: FC2Record[];
};

const route = useRoute();
const payload = ref<FC2Record>({});
const plan = ref<FC2Record>({});
const loading = ref(false);
const actionLoading = ref(false);
const error = ref("");
const message = ref("");
const messageOk = ref(true);

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  message.value = "";
  plan.value = {};
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title") || "FC2媒体入库");
const pageTitle = computed(() => text(payload.value, "PageTitle") || title.value);
const pageNote = computed(() => text(payload.value, "PageNote"));
const rootDirs = computed(() => asArray<string>(pick(payload.value, "RootDirs")));
const favoriteRows = computed(() => asArray<FC2Record>(pick(payload.value, "FavoriteRows")));
const commitPhase = computed(() => text(plan.value, "phase") === "commit");
const precheckSummary = computed(() => asRecord(pick(plan.value, "precheck")));
const canCommit = computed(() => Boolean(pick(plan.value, "can_commit")));
const showReturn = computed(() => Boolean(pick(plan.value, "partial_failed")) && Boolean(pick(plan.value, "has_plan")));
const replacementCount = computed(() => num(plan.value, "replace"));
const precheckPassRows = computed(() => asArray<FC2Record>(pick(plan.value, "precheck_pass")));
const commitButtonText = computed(() => {
  if (pick(plan.value, "partial_failed")) {
    return "第二段：仅执行通过项";
  }
  return replacementCount.value ? "第二段：执行新增 / 替换" : "第二段：执行入库";
});
const statusText = computed(() => text(plan.value, "message") || "等待操作");
const generatedAt = computed(() => {
  if (commitPhase.value) {
    return "";
  }
  const value = text(plan.value, "generated_at");
  return value && value !== "0" ? value : "";
});
const summary = computed(() => ({
  total: commitPhase.value ? num(plan.value, "total") : num(precheckSummary.value, "total"),
  success: commitPhase.value ? num(plan.value, "success") : num(precheckSummary.value, "passed"),
  failed: commitPhase.value ? num(plan.value, "failed") : num(precheckSummary.value, "failed"),
  passLabel: commitPhase.value ? "成功" : "通过"
}));
const tables = computed<ResultTableDef[]>(() => [
  {
    id: "tblPrecheckPass",
    title: "预处理通过 Movie",
    columns: [
      { label: "FC2影片标识", key: "fc2_key", width: "12%" },
      { label: "处理方式", key: "action_label", width: "12%" },
      { label: "源文件", key: "source_path", width: "18%", cls: "mono" },
      { label: "旧文件", key: "existing_path", width: "18%", cls: "mono" },
      { label: "目标名", key: "target_file_name", width: "14%", cls: "mono" },
      { label: "目标目录", key: "target_dir", width: "14%", cls: "mono" },
      { label: "提醒", key: "notice", width: "12%" }
    ],
    rows: precheckPassRows.value
  },
  {
    id: "tblPrecheckFail",
    title: "预处理失败 Movie",
    columns: [
      { label: "FC2影片标识", key: "fc2_key", width: "16%" },
      { label: "源文件", key: "source_path", width: "42%", cls: "mono" },
      { label: "失败原因", key: "error", width: "42%" }
    ],
    rows: asArray<FC2Record>(pick(plan.value, "precheck_fail"))
  },
  {
    id: "tblPrecheckPreview",
    title: "入库前预览（通过项）",
    columns: [
      { label: "FC2影片标识", key: "fc2_key", width: "12%" },
      { label: "处理方式", key: "action_label", width: "14%" },
      { label: "旧文件", key: "existing_path", width: "22%", cls: "mono" },
      { label: "目标目录", key: "target_dir", width: "22%", cls: "mono" },
      { label: "目标名", key: "target_file_name", width: "18%", cls: "mono" },
      { label: "大小", key: "size", width: "12%" }
    ],
    rows: asArray<FC2Record>(pick(plan.value, "precheck_preview"))
  },
  {
    id: "tblCommitSuccess",
    title: "入库成功 Movie",
    columns: [
      { label: "FC2影片标识", key: "fc2_key", width: "12%" },
      { label: "处理方式", key: "action_label", width: "14%" },
      { label: "目标路径", key: "target_path", width: "28%", cls: "mono" },
      { label: "旧文件移除位置", key: "removed_path", width: "28%", cls: "mono" },
      { label: "提醒", key: "notice", width: "18%" }
    ],
    rows: asArray<FC2Record>(pick(plan.value, "commit_success"))
  },
  {
    id: "tblCommitFail",
    title: "入库失败 Movie",
    columns: [
      { label: "FC2影片标识", key: "fc2_key", width: "12%" },
      { label: "处理方式", key: "action_label", width: "16%" },
      { label: "源文件", key: "source_path", width: "32%", cls: "mono" },
      { label: "失败原因", key: "error", width: "40%" }
    ],
    rows: asArray<FC2Record>(pick(plan.value, "commit_fail"))
  },
  {
    id: "tblCommitPreview",
    title: "入库后预览（成功项）",
    columns: [
      { label: "FC2影片标识", key: "fc2_key", width: "12%" },
      { label: "处理方式", key: "action_label", width: "16%" },
      { label: "目标路径", key: "target_path", width: "32%", cls: "mono" },
      { label: "旧文件移除位置", key: "removed_path", width: "28%", cls: "mono" },
      { label: "大小", key: "size", width: "12%" }
    ],
    rows: asArray<FC2Record>(pick(plan.value, "commit_preview"))
  }
]);

async function runAction(path: string, successMessage: string) {
  actionLoading.value = true;
  message.value = "";
  try {
    const data = asRecord(await sendJSON(path, "POST", {}));
    plan.value = data;
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
  void runAction("/api/triggers/fc2-media/precheck", "预处理完成");
}

function runCommit() {
  void runAction("/api/triggers/fc2-media/commit", "第二段执行完成");
}

function runReturn() {
  void runAction("/api/triggers/fc2-media/return", "已返回");
}

function favoriteKey(row: FC2Record, index: number): string {
  return `${text(row, "SourceType")}-${text(row, "SourceRowID")}-${index}`;
}
</script>
