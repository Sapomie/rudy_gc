<template>
  <PageShell class="crawler-detail-loop-page-shell" :title="title" eyebrow="">
    <div id="crawlerDetailLoopPage" class="crawler-detail-loop-page-root">
      <LoadingState v-if="loading" />
      <ErrorPanel v-else-if="error" :message="error" />
      <template v-else>
        <div class="row">
          <div class="col-12">
            <div class="page-title-box d-flex align-items-center justify-content-between">
              <h4 class="mb-0">{{ title }}</h4>
              <a href="/triggers/dailybest" class="btn btn-outline-primary btn-sm">回到 DailyBest</a>
            </div>
          </div>
        </div>

        <div class="row">
          <div class="col-lg-5">
            <div class="card">
              <div class="card-body">
                <div class="d-flex align-items-center justify-content-between mb-3">
                  <h4 class="card-title mb-0">循环状态</h4>
                  <span id="runningBadge" class="badge" :class="runningBadgeClass">{{ runningText }}</span>
                </div>

                <div class="detail-loop-status-grid">
                  <div class="overview-item">
                    <div class="label">启动时间</div>
                    <div class="value detail-loop-text" id="startedAtText">{{ startedAt }}</div>
                  </div>
                  <div class="overview-item">
                    <div class="label">最近日志</div>
                    <div class="value detail-loop-text" id="lastEventAtText">{{ lastEventAt }}</div>
                  </div>
                  <div class="overview-item">
                    <div class="label">缓冲数量</div>
                    <div class="value" id="bufferedText">{{ buffered }}</div>
                  </div>
                  <div class="overview-item">
                    <div class="label">暂停状态</div>
                    <div class="value" id="pausedText">{{ pausedText }}</div>
                  </div>
                </div>

                <div class="mt-3">
                  <div class="small text-muted mb-2">最近一条</div>
                  <div id="lastLineText" class="detail-loop-last-line">{{ lastLine || "-" }}</div>
                </div>

                <div class="detail-loop-actions">
                  <button type="button" class="btn btn-primary" id="startBtn" @click="start">开启 loop</button>
                  <button type="button" class="btn btn-outline-danger" id="stopBtn" @click="stop">停止 loop</button>
                </div>

                <div id="statusMsg" class="small text-muted mt-3">{{ message || "等待加载" }}</div>
              </div>
            </div>
          </div>

          <div class="col-lg-7">
            <div class="card event-card detail-loop-log-card">
              <div class="card-body">
                <div class="d-flex align-items-center justify-content-between mb-3">
                  <h4 class="card-title mb-0">实时日志</h4>
                  <span id="eventCountText" class="badge bg-light text-dark">0 条</span>
                </div>
                <div class="result-box">
                  <pre id="resultBox">{{ lastLine || "等待连接..." }}</pre>
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
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LoadingState from "@/components/LoadingState.vue";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asRecord, pick, text } from "@/utils/format";

const route = useRoute();
const page = ref<AnyRecord>({});
const status = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const message = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    page.value = asRecord(await getJSON(apiPathForRoute(route)));
    status.value = asRecord(await getJSON("/api/crawler/detail-loop"));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(page.value, "Title") || "详情抓取循环");
const running = computed(() => Boolean(pick(status.value, "running", "Running")));
const runningText = computed(() => running.value ? "运行中" : "未启动");
const runningBadgeClass = computed(() => running.value ? "badge-running" : "badge-stopped");
const startedAt = computed(() => text(status.value, "started_at", "StartedAt") || "-");
const lastEventAt = computed(() => text(status.value, "last_event_at", "LastEventAt") || "-");
const buffered = computed(() => text(status.value, "buffered", "Buffered") || "0");
const pausedText = computed(() => Boolean(pick(status.value, "paused", "Paused")) ? "是" : "否");
const lastLine = computed(() => text(status.value, "last_line", "LastLine"));

async function start() {
  await call("/api/crawler/detail-loop/start");
}

async function stop() {
  await call("/api/crawler/detail-loop/stop");
}

async function call(path: string) {
  try {
    const data = asRecord(await sendJSON(path, "POST", {}));
    message.value = text(data, "message", "Message") || "操作完成";
    status.value = asRecord(await getJSON("/api/crawler/detail-loop"));
  } catch (err) {
    message.value = err instanceof Error ? err.message : String(err);
  }
}
</script>
