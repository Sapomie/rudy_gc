<template>
  <PageShell class="crawler-tasks-page-shell" :title="title" eyebrow="">
    <div id="crawlerTasksPage" class="container-fluid crawler-tasks-page-root next02-sc-pink-theme next02-crawler-tasks-rewrite">
      <LoadingState v-if="loading" />
      <ErrorPanel v-else-if="error" :message="error" />
      <template v-else>
        <div class="row g-0">
          <div class="col-12">
            <section class="page-title-box next02-crawler-tasks-hero next02-sc-movie-top-shell" aria-labelledby="crawlerTasksTitle">
              <div class="next02-crawler-tasks-hero-main">
                <div class="next02-crawler-tasks-kicker">Crawler Jobs</div>
                <h4 id="crawlerTasksTitle">{{ title }}</h4>
                <div class="next02-crawler-tasks-note">查看抓取任务状态、计数和任务入口。</div>
              </div>
              <div class="next02-crawler-tasks-hero-actions">
                <a :href="backHref" class="btn btn-outline-primary btn-sm">回到任务台</a>
              </div>
            </section>
          </div>
        </div>

        <div class="row g-0">
          <div class="col-12">
            <div class="card next02-crawler-tasks-panel next02-sc-movie-table-card">
              <div class="card-body">
                <div class="next02-crawler-tasks-panel-head">
                  <div class="next02-crawler-tasks-panel-title">
                    <h4 class="card-title mb-0">任务列表</h4>
                    <span id="taskCountBadge" class="badge bg-light text-dark">{{ rows.length }} 个</span>
                  </div>
                  <div class="next02-crawler-tasks-metrics" aria-label="抓取任务统计">
                    <span v-for="item in taskMetrics" :key="item.label" class="next02-crawler-tasks-metric">
                      <span>{{ item.label }}</span>
                      <strong>{{ item.value }}</strong>
                    </span>
                  </div>
                </div>
                <div class="table-responsive">
                  <table class="table table-hover align-middle mb-0 next02-crawler-tasks-table">
                    <colgroup>
                      <col style="width: 17%">
                      <col style="width: 17%">
                      <col style="width: 18%">
                      <col style="width: 7%">
                      <col style="width: 7%">
                      <col style="width: 10%">
                      <col style="width: 16%">
                      <col style="width: 8%">
                    </colgroup>
                    <thead>
                      <tr>
                        <th>任务ID</th>
                        <th>任务</th>
                        <th>当前状态</th>
                        <th>待处理</th>
                        <th>已处理</th>
                        <th>成功/失败</th>
                        <th>开始时间</th>
                        <th>操作</th>
                      </tr>
                    </thead>
                    <tbody id="taskTableBody">
                      <tr v-if="!rows.length">
                        <td colspan="8" class="next02-crawler-tasks-empty">暂无任务</td>
                      </tr>
                      <tr v-for="row in rows" v-else :key="jobID(row)">
                        <td class="next02-task-id" :title="jobID(row)">{{ jobID(row) }}</td>
                        <td :title="taskLabel(row)">{{ taskLabel(row) }}</td>
                        <td :title="statusText(row)"><span class="badge fetch-site-badge" :class="statusClass(row)">{{ statusText(row) }}</span></td>
                        <td class="next02-task-number">{{ text(row, "queued_count", "QueuedCount") || 0 }}</td>
                        <td class="next02-task-number">{{ text(row, "handled_count", "HandledCount") || 0 }}</td>
                        <td class="next02-task-number">{{ text(row, "success_count", "SuccessCount") || 0 }} / {{ text(row, "failed_count", "FailedCount") || 0 }}</td>
                        <td class="next02-task-time"><ScDateTime :value="startedAtValue(row)" :unix="startedAtIsUnix(row)" seconds /></td>
                        <td><a class="next02-task-link" data-job-action="back-to-task" :href="taskHref(row)">回到任务</a></td>
                      </tr>
                    </tbody>
                  </table>
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
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LoadingState from "@/components/LoadingState.vue";
import PageShell from "@/components/PageShell.vue";
import ScDateTime from "@/components/ScDateTime.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, pick, text } from "@/utils/format";

const route = useRoute();
const page = ref<AnyRecord>({});
const jobs = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
let jobsPoller: number | undefined;

watch(() => route.fullPath, load, { immediate: true });

onBeforeUnmount(() => {
  if (jobsPoller) {
    window.clearInterval(jobsPoller);
    jobsPoller = undefined;
  }
});

async function load() {
  loading.value = true;
  error.value = "";
  try {
    page.value = asRecord(await getJSON(apiPathForRoute(route)));
    jobs.value = asRecord(await getJSON("/api/crawler/jobs"));
    ensureJobsPolling();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(page.value, "Title") || "任务列表");
const backHref = computed(() => {
  const base = text(page.value, "TriggerPageURL") || "/dailybestfetch";
  const id = text(page.value, "JobID");
  return id ? `${base}?job_id=${encodeURIComponent(id)}` : base;
});
const rows = computed(() => asArray<AnyRecord>(pick(jobs.value, "jobs", "Jobs", "items", "Items")));
const runningRows = computed(() => rows.value.filter((row) => !Boolean(pick(row, "done", "Done"))).length);
const doneRows = computed(() => rows.value.filter((row) => Boolean(pick(row, "done", "Done"))).length);
const taskMetrics = computed(() => [
  { label: "任务总数", value: String(rows.value.length) },
  { label: "运行中", value: String(runningRows.value) },
  { label: "已完成", value: String(doneRows.value) }
]);

const taskLabels: Record<string, string> = {
  spider_daily_best: "DailyBest 抓取",
  spider_daily_best_sync: "DailyBest 同步",
  spider_seeds: "活跃 Seeds",
  spider_seed_by_name: "按名称抓取",
  spider_refresh_oldest_detail: "刷新最久详情",
  spider_download_cover: "图片抓取",
  spider_translate_title: "标题翻译",
  spider_post_process: "图片 + 标题翻译",
  spider_backfill_person: "person 回填",
  spider_backfill_rank_period: "周期排行回填",
  spider_backfill_fetch_site: "外站任务回填",
  spider_fetch_javbus_resources: "JavBus 资源抓取",
  spider_fetch_javbus_filtered_resources: "JavBus 列表筛选抓取",
  spider_fetch_sukebei_resources: "Sukebei 资源抓取",
  spider_fetch_sukebei_filtered_resources: "Sukebei 列表筛选抓取",
  spider_fetch_sehuatang_magnets: "色花堂磁力抓取",
  spider_rebuild_cast_rank: "演员 Rank 回填",
  spider_rebuild_actor_rank: "单演员 Rank",
  sc_rebuild_stats: "SC 统计回填",
  sc_move: "SC 影片移动",
  sc_add: "SC 新增"
};

const triggerPages: Record<string, string> = {
  spider_daily_best: "/dailybestfetch",
  spider_daily_best_sync: "/dailybestfetch",
  spider_seeds: "/seedfetch",
  spider_seed_by_name: "/seedfetch",
  spider_refresh_oldest_detail: "/seedfetch",
  spider_download_cover: "/postprocess",
  spider_translate_title: "/postprocess",
  spider_post_process: "/postprocess",
  spider_backfill_person: "/databackfill",
  spider_backfill_rank_period: "/databackfill",
  spider_backfill_fetch_site: "/databackfill",
  spider_fetch_javbus_resources: "/sitefetch",
  spider_fetch_javbus_filtered_resources: "/javbusfetch",
  spider_fetch_sukebei_resources: "/sitefetch",
  spider_fetch_sukebei_filtered_resources: "/sukebeifetch",
  spider_fetch_sehuatang_magnets: "/sehuatangfetch",
  spider_rebuild_cast_rank: "/databackfill",
  spider_rebuild_actor_rank: "/databackfill",
  sc_rebuild_stats: "/databackfill",
  sc_move: "/databackfill",
  sc_add: "/scadd"
};

function jobID(row: AnyRecord): string {
  return text(row, "job_id", "JobID", "id", "ID");
}

function taskType(row: AnyRecord): string {
  return text(row, "task_type", "TaskType");
}

function taskLabel(row: AnyRecord): string {
  const type = taskType(row);
  return taskLabels[type] || type || "-";
}

function stageLabel(stage: string): string {
  switch (stage) {
    case "job_started":
      return "任务已启动";
    case "detail_paused":
      return "详情抓取已暂停";
    case "detail_resumed":
      return "详情抓取已恢复";
    case "paused":
      return "任务已暂停";
    case "resumed":
      return "任务已继续";
    case "done":
      return "任务完成";
    case "failed":
      return "任务失败";
    default:
      return stage || "-";
  }
}

function statusText(row: AnyRecord): string {
  return text(row, "message", "Message") || stageLabel(text(row, "stage", "Stage", "status", "Status"));
}

function statusClass(row: AnyRecord): string {
  const stage = text(row, "stage", "Stage", "status", "Status");
  if (Boolean(pick(row, "done", "Done"))) {
    return "is-done";
  }
  if (stage.includes("failed")) {
    return "is-error";
  }
  if (stage.includes("paused")) {
    return "is-paused";
  }
  return "is-running";
}

function startedAtValue(row: AnyRecord): unknown {
  const existing = text(row, "started_at_text", "StartedAtText");
  return existing || pick(row, "started_at", "StartedAt");
}

function startedAtIsUnix(row: AnyRecord): boolean {
  return !text(row, "started_at_text", "StartedAtText");
}

function taskHref(row: AnyRecord): string {
  const base = triggerPages[taskType(row)] || text(page.value, "TriggerPageURL") || "/dailybestfetch";
  return `${base}?job_id=${encodeURIComponent(jobID(row))}`;
}

function ensureJobsPolling() {
  if (jobsPoller) {
    return;
  }
  jobsPoller = window.setInterval(async () => {
    try {
      jobs.value = asRecord(await getJSON("/api/crawler/jobs"));
    } catch {
      // Keep the current table visible when one refresh fails.
    }
  }, 5000);
}
</script>
