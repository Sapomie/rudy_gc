<template>
  <PageShell class="movie-agg-page-shell sc-events-page-shell media-ingest-page-shell media-rollback-page-shell" :title="title" eyebrow="">
    <div id="page-root" class="container-fluid movie-agg-page-root sc-events-page-root media-ingest-page-root media-rollback-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div id="msgArea" class="alert fade show small py-2 px-3" :class="messageOk ? 'alert-success' : 'alert-danger'" :style="{ display: message ? 'block' : 'none' }">
          {{ message }}
        </div>

        <div id="mediaRollbackPage" class="next02-media-rewrite next02-rollback-rewrite">
          <div class="row mb-3">
            <div class="col-12">
              <div class="next02-media-hero">
                <div class="next02-media-hero-main">
                  <div class="next02-media-eyebrow">Media Rollback</div>
                <h4 class="mb-1">{{ pageTitle }}</h4>
                <p class="text-muted mb-0">{{ pageNote }}</p>
                  <div class="next02-media-kpi-strip">
                    <span v-for="item in heroStats" :key="item.label">
                      <strong>{{ item.value }}</strong>
                      <em>{{ item.label }}</em>
                    </span>
                  </div>
              </div>
                <div class="next02-media-hero-actions">
                <a href="/triggers/media" class="btn btn-outline-primary btn-sm">返回入库页</a>
                <a href="/triggers/media-rescan" class="btn btn-outline-secondary btn-sm">媒体重扫</a>
                </div>
              </div>
            </div>
          </div>

          <div class="row next02-media-grid">
            <div class="col-lg-4 next02-media-control-col">
              <div class="card next02-media-action-card next02-rollback-action-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Rollback</div>
                      <h5 class="card-title mb-0">操作</h5>
                    </div>
                    <span class="next02-media-live">Ready</span>
                  </div>
                  <div class="next02-media-action-stack next02-rollback-action-stack">
                    <button type="button" id="btnRollback" class="btn btn-primary" :disabled="actionLoading" @click="runRollback">开始回滚名称</button>
                  </div>
                  <div class="next02-media-help">
                    从每个 root 的 <code>001_process/004_rollback</code> 读取文件，按规则还原名称并移回 <code>001_process/001_ingest_new</code>。
                  </div>
                </div>
              </div>

              <div class="card next02-media-overview-card next02-rollback-overview-card">
                <div class="card-body">
                  <div class="next02-media-panel-head">
                    <div>
                      <div class="next02-media-eyebrow">Overview</div>
                      <h5 class="card-title mb-0">运行概览</h5>
                    </div>
                  </div>
                  <div class="row g-2">
                    <div class="col-4"><div class="overview-item"><div class="label">总数</div><div class="value" id="sumTotal">{{ num(result, "total") }}</div></div></div>
                    <div class="col-4"><div class="overview-item"><div class="label">成功</div><div class="value text-success" id="sumSuccess">{{ num(result, "success") }}</div></div></div>
                    <div class="col-4"><div class="overview-item"><div class="label">失败</div><div class="value text-danger" id="sumFail">{{ num(result, "failed") }}</div></div></div>
                  </div>
                  <div class="mt-3">
                    <div class="text-muted small">状态</div>
                    <div id="statusText" class="fw-semibold">{{ text(result, "message") || "等待操作" }}</div>
                  </div>
                </div>
              </div>

              <div class="card next02-media-root-card next02-rollback-root-card">
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

const title = computed(() => text(payload.value, "Title") || "媒体回滚");
const pageTitle = computed(() => text(payload.value, "PageTitle") || title.value);
const pageNote = computed(() => text(payload.value, "PageNote"));
const rootDirs = computed(() => asArray<string>(pick(payload.value, "RootDirs")));
const heroStats = computed(() => [
  { label: "Root", value: rootDirs.value.length },
  { label: "总数", value: num(result.value, "total") },
  { label: "成功", value: num(result.value, "success") },
  { label: "失败", value: num(result.value, "failed") }
]);
const tables = computed<ResultTableDef[]>(() => [
  {
    id: "tblRollbackSuccess",
    title: "回滚成功 Movie",
    columns: [
      { label: "影片", key: "movie_name" },
      { label: "源文件", key: "source_path", cls: "mono" },
      { label: "目标文件", key: "target_path", cls: "mono" }
    ],
    rows: asArray<AnyRecord>(pick(result.value, "rollback_success"))
  },
  {
    id: "tblRollbackFail",
    title: "回滚失败 Movie",
    columns: [
      { label: "影片", key: "movie_name" },
      { label: "源文件", key: "source_path", cls: "mono" },
      { label: "失败原因", key: "error" },
      { label: "失败文件位置", key: "failed_path", cls: "mono" }
    ],
    rows: asArray<AnyRecord>(pick(result.value, "rollback_fail"))
  },
  {
    id: "tblRollbackPreview",
    title: "回滚后预览（成功项）",
    columns: [
      { label: "影片", key: "movie_name" },
      { label: "回滚后路径", key: "target_path", cls: "mono" }
    ],
    rows: asArray<AnyRecord>(pick(result.value, "rollback_preview"))
  }
]);

async function runRollback() {
  actionLoading.value = true;
  message.value = "";
  try {
    const data = asRecord(await sendJSON("/api/triggers/media/rollback", "POST", {}));
    result.value = data;
    message.value = text(data, "message") || "回滚完成";
    messageOk.value = true;
  } catch (err) {
    message.value = err instanceof Error ? err.message : "回滚失败";
    messageOk.value = false;
  } finally {
    actionLoading.value = false;
  }
}

function cellValue(row: AnyRecord, column: ResultColumn): string {
  return formatValue(pick(row, column.key));
}

const ResultTable = defineComponent({
  name: "MediaRollbackResultTable",
  props: {
    table: { type: Object as PropType<ResultTableDef>, required: true }
  },
  setup(props) {
    return () => h("div", { class: "card next02-media-result-card next02-rollback-result-card" }, [
      h("div", { class: "card-body" }, [
        h("div", { class: "next02-media-panel-head" }, [
          h("div", [
            h("div", { class: "next02-media-eyebrow" }, "Result"),
            h("h5", { class: "card-title mb-0" }, props.table.title)
          ]),
          h("span", { class: "next02-media-count-pill" }, `${props.table.rows.length} 行`)
        ]),
        h("div", { class: "table-responsive" }, [
          h("table", { class: "table table-sm table-hover", id: props.table.id }, [
            h("thead", [
              h("tr", props.table.columns.map((column) => h("th", column.label)))
            ]),
            h("tbody", props.table.rows.length
              ? props.table.rows.map((row) => h("tr", props.table.columns.map((column) => h("td", { class: column.cls || "" }, cellValue(row, column)))))
              : [h("tr", [h("td", { colspan: props.table.columns.length, class: "text-center text-muted" }, "暂无数据")])])
          ])
        ])
      ])
    ]);
  }
});
</script>
