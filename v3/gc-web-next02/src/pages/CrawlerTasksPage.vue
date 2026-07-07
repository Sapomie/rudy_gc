<template>
  <PageShell class="crawler-tasks-page-shell" :title="title" eyebrow="">
    <div id="crawlerTasksPage" class="crawler-tasks-page-root">
      <LoadingState v-if="loading" />
      <ErrorPanel v-else-if="error" :message="error" />
      <template v-else>
        <div class="row">
          <div class="col-12">
            <div class="page-title-box d-flex align-items-center justify-content-between">
              <h4 class="mb-0">{{ title }}</h4>
              <a :href="backHref" class="btn btn-outline-primary btn-sm">回到任务台</a>
            </div>
          </div>
        </div>

        <div class="row">
          <div class="col-12">
            <div class="card">
              <div class="card-body">
                <div class="d-flex align-items-center justify-content-between mb-3">
                  <h4 class="card-title mb-0">任务列表</h4>
                  <span id="taskCountBadge" class="badge bg-light text-dark">{{ rows.length }} 个</span>
                </div>
                <div class="table-responsive">
                  <table class="table table-centered table-nowrap mb-0">
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
                        <td colspan="8" class="text-center text-muted py-4">暂无任务</td>
                      </tr>
                      <tr v-for="row in rows" v-else :key="jobID(row)">
                        <td class="text-nowrap">{{ jobID(row) }}</td>
                        <td>{{ text(row, "task_type", "TaskType") }}</td>
                        <td><span class="badge fetch-site-badge">{{ text(row, "status", "Status") }}</span></td>
                        <td>{{ text(row, "queued_count", "QueuedCount") || 0 }}</td>
                        <td>{{ text(row, "handled_count", "HandledCount") || 0 }}</td>
                        <td>{{ text(row, "success_count", "SuccessCount") || 0 }} / {{ text(row, "failed_count", "FailedCount") || 0 }}</td>
                        <td class="text-nowrap">{{ text(row, "started_at_text", "StartedAtText", "started_at", "StartedAt") || "-" }}</td>
                        <td class="text-nowrap"><button type="button" class="btn btn-sm btn-outline-secondary" disabled>查看</button></td>
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
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LoadingState from "@/components/LoadingState.vue";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, pick, text } from "@/utils/format";

const route = useRoute();
const page = ref<AnyRecord>({});
const jobs = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    page.value = asRecord(await getJSON(apiPathForRoute(route)));
    jobs.value = asRecord(await getJSON("/api/crawler/jobs"));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(page.value, "Title") || "任务列表");
const backHref = computed(() => {
  const base = text(page.value, "TriggerPageURL") || "/triggers/dailybest";
  const id = text(page.value, "JobID");
  return id ? `${base}?job_id=${encodeURIComponent(id)}` : base;
});
const rows = computed(() => asArray<AnyRecord>(pick(jobs.value, "jobs", "Jobs", "items", "Items")));

function jobID(row: AnyRecord): string {
  return text(row, "job_id", "JobID", "id", "ID");
}
</script>
