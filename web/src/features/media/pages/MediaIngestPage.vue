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
                  <div class="next02-media-eyebrow">Media Intake</div>
                  <h4 id="mediaIngestTitle">{{ pageTitle }}</h4>
                  <p>{{ pageNote }}</p>
                </div>
                <div class="next02-media-hero-actions">
                  <a href="/torrentalbumlist?album_name=%E4%B8%8B%E8%BD%BD%E4%B8%AD" class="btn btn-outline-primary btn-sm">下载中列表</a>
                  <a href="/mediarescan" class="btn btn-outline-secondary btn-sm">媒体重扫</a>
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
              <div class="card next02-media-action-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Pipeline</div>
                      <h5 class="card-title mb-0">操作</h5>
                    </div>
                    <span class="next02-media-live" :class="{ 'is-running': actionLoading }">{{ actionLoading ? "执行中" : "就绪" }}</span>
                  </div>
                  <div class="next02-media-action-stack">
                    <button type="button" id="btnPrecheck" class="btn btn-primary" :disabled="actionLoading" @click="runPrecheck">第一段：预处理</button>
                    <button type="button" id="btnCommit" class="btn btn-success" :disabled="actionLoading || !canCommit" :data-can-commit="canCommit ? '1' : '0'" @click="runCommit">{{ commitButtonText }}</button>
                    <button type="button" id="btnReturn" class="btn btn-outline-secondary" :class="{ 'd-none': !showReturn }" :disabled="actionLoading" @click="runReturn">返回（清空本次预处理）</button>
                  </div>
                  <div class="next02-media-help">
                    <div>部分失败时可选择：</div>
                    <div>1. 仅插入通过项（点第二段）</div>
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
                    <div class="col-4"><div class="overview-item"><div class="label">总数</div><div class="value" id="sumTotal">{{ summary.total }}</div></div></div>
                    <div class="col-4"><div class="overview-item"><div class="label" id="sumPassLabel">{{ summary.passLabel }}</div><div class="value text-success" id="sumPass">{{ summary.success }}</div></div></div>
                    <div class="col-4"><div class="overview-item"><div class="label" id="sumFailLabel">失败</div><div class="value text-danger" id="sumFail">{{ summary.failed }}</div></div></div>
                  </div>
                  <div class="mt-3">
                    <div class="text-muted small">状态</div>
                    <div id="statusText" class="fw-semibold">{{ statusText }}</div>
                  </div>
                  <div v-if="generatedAt" class="mt-2 text-muted small" id="generatedAtText">预处理时间：<ScDateTime :value="generatedAt" /></div>
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
                    <li v-for="dir in rootDirs" :key="dir"><code>{{ dir }}</code></li>
                    <li v-if="!rootDirs.length" class="text-muted">未配置 media.rootDirs</li>
                  </ul>
                </div>
              </div>

              <div class="card next02-media-favorite-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Albums</div>
                      <h5 class="card-title mb-0">下载中（tm_album_item）</h5>
                    </div>
                    <span class="next02-media-count-pill">{{ favoriteRows.length }} 条</span>
                  </div>
                  <div class="table-responsive">
                    <table class="table table-hover align-middle mb-0 next02-media-table next02-media-favorite-table">
                      <colgroup><col style="width: 50%"><col style="width: 25%"><col style="width: 25%"></colgroup>
                      <thead><tr><th>影片</th><th>来源</th><th>大小</th></tr></thead>
                      <tbody>
                        <tr v-for="(row, index) in favoriteRows" :key="favoriteKey(row, index)">
                          <td :title="favoriteLabel(row)">{{ favoriteLabel(row) }}</td>
                          <td :title="text(row, 'SourceLabel', 'source_label')">{{ text(row, "SourceLabel", "source_label") }}</td>
                          <td :title="text(row, 'Size', 'size') || '-'">{{ text(row, "Size", "size") || "-" }}</td>
                        </tr>
                        <tr v-if="!favoriteRows.length"><td colspan="3" class="text-center text-muted">暂无下载中数据</td></tr>
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
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/shared/api/request";
import PageShell from "@/shared/components/PageShell.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import type { AnyRecord } from "@/shared/types/gc";
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
  rows: AnyRecord[];
};

const route = useRoute();
const payload = ref<AnyRecord>({});
const plan = ref<AnyRecord>({});
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
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title") || "媒体入库");
const pageTitle = computed(() => text(payload.value, "PageTitle") || title.value);
const pageNote = computed(() => text(payload.value, "PageNote"));
const rootDirs = computed(() => asArray<string>(pick(payload.value, "RootDirs")));
const favoriteRows = computed(() => asArray<AnyRecord>(pick(payload.value, "FavoriteRows")));
const commitPhase = computed(() => text(plan.value, "phase") === "commit");
const precheckSummary = computed(() => asRecord(pick(plan.value, "precheck")));
const canCommit = computed(() => Boolean(pick(plan.value, "can_commit")));
const showReturn = computed(() => Boolean(pick(plan.value, "partial_failed")) && Boolean(pick(plan.value, "has_plan")));
const commitButtonText = computed(() => Boolean(pick(plan.value, "partial_failed")) ? "第二段：仅插入通过项" : "第二段：执行插入");
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
      { label: "影片", key: "movie_name", width: "14%" },
      { label: "源文件", key: "source_path", width: "26%", cls: "mono" },
      { label: "目标名", key: "target_file_name", width: "18%", cls: "mono" },
      { label: "目标目录", key: "target_dir", width: "18%", cls: "mono" },
      { label: "Alias", key: "alias", width: "10%", cls: "mono" },
      { label: "Hash", key: "source_torrent_hash", width: "14%", cls: "mono" }
    ],
    rows: asArray<AnyRecord>(pick(plan.value, "precheck_pass"))
  },
  {
    id: "tblPrecheckFail",
    title: "预处理失败 Movie",
    columns: [
      { label: "影片", key: "movie_name", width: "16%" },
      { label: "源文件", key: "source_path", width: "42%", cls: "mono" },
      { label: "失败原因", key: "error", width: "42%" }
    ],
    rows: asArray<AnyRecord>(pick(plan.value, "precheck_fail"))
  },
  {
    id: "tblPrecheckPreview",
    title: "插入前预览（通过项）",
    columns: [
      { label: "影片", key: "movie_name", width: "16%" },
      { label: "目标目录", key: "target_dir", width: "26%", cls: "mono" },
      { label: "目标名", key: "target_file_name", width: "26%", cls: "mono" },
      { label: "Alias", key: "alias", width: "16%", cls: "mono" },
      { label: "大小", key: "size", width: "16%" }
    ],
    rows: asArray<AnyRecord>(pick(plan.value, "precheck_preview"))
  },
  {
    id: "tblCommitSuccess",
    title: "插入成功 Movie",
    columns: [
      { label: "影片", key: "movie_name", width: "16%" },
      { label: "目标路径", key: "target_path", width: "40%", cls: "mono" },
      { label: "Alias", key: "alias", width: "18%", cls: "mono" },
      { label: "Hash", key: "source_torrent_hash", width: "26%", cls: "mono" }
    ],
    rows: asArray<AnyRecord>(pick(plan.value, "commit_success"))
  },
  {
    id: "tblCommitFail",
    title: "插入失败 Movie",
    columns: [
      { label: "影片", key: "movie_name", width: "16%" },
      { label: "源文件", key: "source_path", width: "42%", cls: "mono" },
      { label: "失败原因", key: "error", width: "42%" }
    ],
    rows: asArray<AnyRecord>(pick(plan.value, "commit_fail"))
  },
  {
    id: "tblCommitPreview",
    title: "插入后预览（成功项）",
    columns: [
      { label: "影片", key: "movie_name", width: "16%" },
      { label: "目标路径", key: "target_path", width: "46%", cls: "mono" },
      { label: "Alias", key: "alias", width: "18%", cls: "mono" },
      { label: "大小", key: "size", width: "20%" }
    ],
    rows: asArray<AnyRecord>(pick(plan.value, "commit_preview"))
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
  void runAction("/api/triggers/media/precheck", "预处理完成");
}

function runCommit() {
  void runAction("/api/triggers/media/commit", "第二段执行完成");
}

function runReturn() {
  void runAction("/api/triggers/media/return", "已返回");
}

function favoriteLabel(row: AnyRecord): string {
  return text(row, "MovieName") || text(row, "MovieJavID", "MovieJavId") || "-";
}

function favoriteKey(row: AnyRecord, index: number): string {
  return `${favoriteLabel(row)}-${index}`;
}

function cellValue(row: AnyRecord, column: ResultColumn): string {
  return formatValue(pick(row, column.key));
}

const ResultTable = defineComponent({
  name: "MediaIngestResultTable",
  props: {
    table: { type: Object as PropType<ResultTableDef>, required: true }
  },
  setup(props) {
    return () => h("div", { class: "card next02-media-result-card", id: props.table.id.replace("tbl", "").replace(/^./, (char) => `${char.toLowerCase()}`) + "Section" }, [
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
