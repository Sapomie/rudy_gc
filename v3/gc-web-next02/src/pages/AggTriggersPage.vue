<template>
  <PageShell class="movie-agg-page-shell agg-triggers-page-shell" :title="title" :note="pageNote" eyebrow="">
    <div id="page-root" class="container-fluid movie-agg-page-root agg-triggers-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div
          v-if="message"
          id="msgArea"
          class="alert fade show small py-2 px-3"
          :class="messageOk ? 'alert-success' : 'alert-danger'"
        >
          {{ message }}
        </div>
        <div v-else id="msgArea" class="alert fade small py-2 px-3 agg-trigger-msg-hidden"></div>

        <div id="aggTriggersPage">
          <div class="row mb-3">
            <div class="col-12 d-flex align-items-center justify-content-between">
              <div>
                <h4 class="mb-1">{{ pageTitle }}</h4>
                <p class="text-muted mb-0">{{ pageNote }}</p>
              </div>
              <div class="d-flex gap-2">
                <a href="/w-media-agg/birth" class="btn btn-outline-primary btn-sm">WMedia 聚合页</a>
                <a href="/movie-agg-all/release" class="btn btn-outline-secondary btn-sm">上映日聚合页</a>
              </div>
            </div>
          </div>

          <div class="row g-3">
            <div v-for="card in cards" :key="card.prefix" class="col-xl-6">
              <div class="card h-100">
                <div class="card-body">
                  <div class="d-flex align-items-center justify-content-between gap-2 mb-3">
                    <h5 class="card-title mb-0">{{ card.title }}</h5>
                    <button type="button" :id="card.buttonId" class="btn btn-outline-primary btn-sm" :disabled="running" @click="runBackfill(card)">
                      执行回填
                    </button>
                  </div>
                  <div class="text-muted small mb-3">{{ card.note }}</div>
                  <div class="row g-2">
                    <div v-for="metric in card.metrics" :key="metric.key" class="col-6">
                      <div class="overview-item">
                        <div class="label">{{ metric.label }}</div>
                        <div class="value" :id="`${card.prefix}${metric.key}`">{{ metricValue(card.prefix, metric.key) }}</div>
                      </div>
                    </div>
                  </div>
                  <div class="mt-3">
                    <div class="text-muted small">状态</div>
                    <div :id="`${card.prefix}BackfillStatusText`" class="fw-semibold">{{ statusFor(card.prefix) }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asRecord, num, pick, text } from "@/utils/format";

type Metric = {
  key: string;
  label: string;
};

type BackfillCard = {
  title: string;
  note: string;
  prefix: "agg" | "releaseAgg";
  buttonId: string;
  url: string;
  loadingText: string;
  doneMessage: string;
  metrics: Metric[];
};

const metricMap: Record<string, string> = {
  ClearedBucketRows: "cleared_bucket_rows",
  ClearedTopRows: "cleared_top_rows",
  ClearedDirtyRows: "cleared_dirty_rows",
  DirtyDays: "dirty_days",
  DirtyMonths: "dirty_months",
  YearBuckets: "year_buckets",
  QuarterBuckets: "quarter_buckets",
  MonthBuckets: "month_buckets",
  DayBuckets: "day_buckets",
  TopRows: "top_rows",
  ElapsedMs: "elapsed_ms"
};

const commonMetrics: Metric[] = [
  { key: "ClearedBucketRows", label: "清空桶行" },
  { key: "ClearedTopRows", label: "清空 Top 行" },
  { key: "ClearedDirtyRows", label: "清空脏标记" },
  { key: "YearBuckets", label: "年桶" },
  { key: "QuarterBuckets", label: "季桶" },
  { key: "MonthBuckets", label: "月桶" },
  { key: "DayBuckets", label: "日桶" },
  { key: "TopRows", label: "Top 行" },
  { key: "ElapsedMs", label: "耗时" }
];

const cards: BackfillCard[] = [
  {
    title: "WMedia 时间聚合回填",
    note: "新建 `w_media` 时间聚合表后，在这里执行一次回填，旧数据才会写入 `/w-media-agg/birth`。",
    prefix: "agg",
    buttonId: "btnBackfillWMediaAgg",
    url: "/api/agg/w-media/backfill",
    loadingText: "正在回填旧的 WMedia 时间聚合数据...",
    doneMessage: "WMedia 时间聚合回填完成",
    metrics: [
      ...commonMetrics.slice(0, 3),
      { key: "DirtyDays", label: "脏日期数" },
      ...commonMetrics.slice(3)
    ]
  },
  {
    title: "上映日时间聚合回填",
    note: "新建 `movie_release_*` 表后，在这里执行一次回填，旧数据才会写入 `/movie-agg-all/release`。",
    prefix: "releaseAgg",
    buttonId: "btnBackfillMovieReleaseAgg",
    url: "/api/agg/movie-release/backfill",
    loadingText: "正在回填...",
    doneMessage: "上映日时间聚合回填完成",
    metrics: [
      ...commonMetrics.slice(0, 3),
      { key: "DirtyMonths", label: "脏月份数" },
      ...commonMetrics.slice(3)
    ]
  }
];

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const message = ref("");
const messageOk = ref(true);
const running = ref(false);
const results = reactive<Record<BackfillCard["prefix"], AnyRecord>>({
  agg: {},
  releaseAgg: {}
});
const statusText = reactive<Record<BackfillCard["prefix"], string>>({
  agg: "",
  releaseAgg: ""
});

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

const title = computed(() => text(payload.value, "Title") || "聚合回填");
const pageTitle = computed(() => text(payload.value, "PageTitle") || title.value);
const pageNote = computed(() => text(payload.value, "PageNote"));

function metricValue(prefix: BackfillCard["prefix"], key: string): string {
  const value = num(results[prefix], metricMap[key] || key);
  return key === "ElapsedMs" ? `${value} ms` : String(value);
}

function statusFor(prefix: BackfillCard["prefix"]): string {
  return statusText[prefix] || "尚未执行回填";
}

function doneStatus(result: AnyRecord): string {
  return `回填完成：日桶 ${num(result, "day_buckets")}，Top 行 ${num(result, "top_rows")}`;
}

async function runBackfill(card: BackfillCard) {
  running.value = true;
  message.value = "";
  statusText[card.prefix] = card.loadingText;
  try {
    const data = asRecord(await sendJSON(card.url, "POST", {}));
    const result = asRecord(pick(data, "result", "Result"));
    results[card.prefix] = result;
    statusText[card.prefix] = doneStatus(result);
    message.value = text(data, "message", "Message") || card.doneMessage;
    messageOk.value = true;
  } catch (err) {
    statusText[card.prefix] = "回填失败，请查看提示信息";
    message.value = err instanceof Error ? err.message : "回填失败";
    messageOk.value = false;
  } finally {
    running.value = false;
  }
}
</script>
