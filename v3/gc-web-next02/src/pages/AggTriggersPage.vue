<template>
  <PageShell class="agg-triggers-page-shell" :title="title" eyebrow="">
    <div id="page-root" class="container-fluid agg-triggers-page-root next02-sc-pink-theme next02-agg-trigger-rewrite">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <section class="next02-agg-trigger-hero next02-sc-movie-top-shell" aria-labelledby="aggTriggerTitle">
          <div class="next02-agg-trigger-hero-main">
            <div class="next02-rewrite-eyebrow">Agg Triggers / Backfill Console</div>
            <h4 id="aggTriggerTitle">{{ pageTitle }}</h4>
            <p>{{ pageNote || "执行 WMedia、上映日与 SH 帖子聚合回填，刷新时间桶和 Top 数据。" }}</p>
          </div>
          <div class="next02-agg-trigger-hero-actions">
            <a href="/aggbirth" class="btn btn-outline-primary btn-sm">WMedia 聚合页</a>
            <a href="/aggrelease" class="btn btn-outline-secondary btn-sm">上映日聚合页</a>
            <a href="/bucketsrelease" class="btn btn-outline-secondary btn-sm">上映桶表</a>
            <a href="/bucketssh" class="btn btn-outline-secondary btn-sm">SH 桶表</a>
          </div>
        </section>

        <div
          v-if="message"
          id="msgArea"
          class="alert fade show small py-2 px-3 next02-agg-message"
          :class="messageOk ? 'alert-success' : 'alert-danger'"
        >
          {{ message }}
        </div>
        <div v-else id="msgArea" class="alert fade small py-2 px-3 agg-trigger-msg-hidden"></div>

        <div id="aggTriggersPage">
          <div class="next02-agg-trigger-grid">
            <div v-for="card in cards" :key="card.prefix" class="next02-agg-job-column">
              <div class="card h-100 next02-rewrite-card next02-agg-job-card">
                <div class="card-body">
                  <div class="next02-agg-panel-head">
                    <div>
                      <div class="next02-agg-panel-eyebrow">Backfill Task</div>
                      <h5 class="card-title mb-0">{{ card.title }}</h5>
                    </div>
                    <button type="button" :id="card.buttonId" class="btn btn-outline-primary btn-sm" :disabled="running" @click="runBackfill(card)">
                      执行回填
                    </button>
                  </div>
                  <div class="next02-agg-card-note">{{ card.note }}</div>
                  <div class="next02-agg-metric-grid">
                    <div v-for="metric in card.metrics" :key="metric.key">
                      <div class="overview-item">
                        <div class="label">{{ metric.label }}</div>
                        <div class="value" :id="`${card.prefix}${metric.key}`">{{ metricValue(card.prefix, metric.key) }}</div>
                      </div>
                    </div>
                  </div>
                  <div class="next02-agg-status-block">
                    <div class="next02-agg-status-label">状态</div>
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
import { computed, onUnmounted, reactive, ref, watch } from "vue";
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
  prefix: "agg" | "releaseAgg" | "shAgg";
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
  WeekBuckets: "week_buckets",
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
    note: "新建 `w_media` 时间聚合表后，在这里执行一次回填，旧数据才会写入 `/aggbirth`。",
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
    note: "新建 `movie_release_*` 表后，在这里执行一次回填，旧数据才会写入 `/aggrelease`。",
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
  },
  {
    title: "SH 帖子时间聚合回填",
    note: "按帖子发布时间回填全量及各论坛的日、ISO 周、月、季度和年聚合桶。",
    prefix: "shAgg",
    buttonId: "btnBackfillSHThreadAgg",
    url: "/api/agg/sh-thread/backfill",
    loadingText: "正在回填 SH 帖子时间聚合数据...",
    doneMessage: "SH 帖子时间聚合回填完成",
    metrics: [
      { key: "ClearedBucketRows", label: "清空桶行" },
      { key: "DirtyDays", label: "帖子日期数" },
      { key: "YearBuckets", label: "年桶" },
      { key: "QuarterBuckets", label: "季桶" },
      { key: "MonthBuckets", label: "月桶" },
      { key: "WeekBuckets", label: "ISO 周桶" },
      { key: "DayBuckets", label: "日桶" },
      { key: "ElapsedMs", label: "耗时" }
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
let messageTimer: number | undefined;
const results = reactive<Record<BackfillCard["prefix"], AnyRecord>>({
  agg: {},
  releaseAgg: {},
  shAgg: {}
});
const statusText = reactive<Record<BackfillCard["prefix"], string>>({
  agg: "",
  releaseAgg: "",
  shAgg: ""
});

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  message.value = "";
  clearMessageTimer();
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

function doneStatus(card: BackfillCard, result: AnyRecord): string {
  if (card.prefix === "shAgg") {
    return `回填完成：日桶 ${num(result, "day_buckets")}，ISO 周桶 ${num(result, "week_buckets")}`;
  }
  return `回填完成：日桶 ${num(result, "day_buckets")}，Top 行 ${num(result, "top_rows")}`;
}

function clearMessageTimer(): void {
  if (messageTimer !== undefined) {
    window.clearTimeout(messageTimer);
    messageTimer = undefined;
  }
}

function showMessage(value: string, ok: boolean): void {
  clearMessageTimer();
  message.value = value;
  messageOk.value = ok;
  messageTimer = window.setTimeout(() => {
    message.value = "";
    messageTimer = undefined;
  }, 3000);
}

function responseResult(data: AnyRecord): AnyRecord {
  const first = asRecord(pick(data, "result", "Result"));
  const nested = asRecord(pick(first, "result", "Result"));
  return Object.keys(nested).length ? nested : first;
}

async function runBackfill(card: BackfillCard) {
  running.value = true;
  message.value = "";
  clearMessageTimer();
  statusText[card.prefix] = card.loadingText;
  try {
    const data = asRecord(await sendJSON(card.url, "POST", {}));
    const result = responseResult(data);
    results[card.prefix] = result;
    statusText[card.prefix] = doneStatus(card, result);
    showMessage(text(data, "message", "Message") || card.doneMessage, true);
  } catch (err) {
    statusText[card.prefix] = "回填失败，请查看提示信息";
    showMessage(err instanceof Error ? err.message : "回填失败", false);
  } finally {
    running.value = false;
  }
}

onUnmounted(clearMessageTimer);
</script>
