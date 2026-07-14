<template>
  <PageShell class="crawler-detail-loop-page-shell" :title="title" eyebrow="">
    <div id="crawlerDetailLoopPage" class="container-fluid crawler-detail-loop-page-root next02-sc-pink-theme next02-detail-loop-rewrite">
      <LoadingState v-if="loading" />
      <ErrorPanel v-else-if="error" :message="error" />
      <template v-else>
        <div class="row g-0">
          <div class="col-12">
            <section class="page-title-box next02-detail-loop-hero next02-sc-movie-top-shell" aria-labelledby="detailLoopTitle">
              <div class="next02-detail-loop-hero-main">
                <div class="next02-detail-loop-kicker">Detail Loop</div>
                <h4 id="detailLoopTitle">{{ title }}</h4>
                <div class="next02-detail-loop-note">详情抓取循环运行状态与实时日志。</div>
              </div>
              <div class="next02-detail-loop-hero-actions">
                <a href="/dailybestfetch" class="btn btn-outline-primary btn-sm">回到 DailyBest</a>
              </div>
            </section>
          </div>
        </div>

        <div class="row next02-detail-loop-workspace">
          <div class="col-lg-5 next02-detail-loop-control-column">
            <div class="card next02-detail-loop-panel">
              <div class="card-body">
                <div class="next02-detail-loop-panel-head">
                  <div>
                    <div class="next02-detail-loop-panel-eyebrow">Loop Control</div>
                    <h4 class="card-title mb-0">循环状态</h4>
                  </div>
                  <span id="runningBadge" class="badge" :class="runningBadgeClass">{{ runningText }}</span>
                </div>

                <div class="detail-loop-status-grid">
                  <div class="overview-item">
                    <div class="label">启动时间</div>
                    <div class="value detail-loop-text" id="startedAtText" :title="startedAt"><ScDateTime :value="startedAtValue" unix seconds /></div>
                  </div>
                  <div class="overview-item">
                    <div class="label">最近日志</div>
                    <div class="value detail-loop-text" id="lastEventAtText" :title="lastEventAt"><ScDateTime :value="lastEventAtValue" unix seconds /></div>
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

                <div class="detail-loop-latest">
                  <div class="detail-loop-latest-label">最近一条</div>
                  <div id="lastLineText" class="detail-loop-last-line">{{ lastLine || "-" }}</div>
                </div>

                <div class="detail-loop-actions">
                  <button type="button" class="btn btn-primary" id="startBtn" :disabled="running || actionPending" @click="start">开启 loop</button>
                  <button type="button" class="btn btn-outline-danger" id="stopBtn" :disabled="!running || actionPending" @click="stop">停止 loop</button>
                </div>

                <div id="statusMsg">{{ message || connectionMessage }}</div>
              </div>
            </div>
          </div>

          <div class="col-lg-7 next02-detail-loop-event-column">
            <div class="card event-card detail-loop-log-card next02-detail-loop-panel">
              <div class="card-body">
                <div class="next02-detail-loop-panel-head">
                  <div>
                    <div class="next02-detail-loop-panel-eyebrow">Event Stream</div>
                    <h4 class="card-title mb-0">实时日志</h4>
                  </div>
                  <div class="next02-detail-loop-event-tools">
                    <span class="next02-detail-loop-connection" :class="`is-${connectionState}`">
                      <span class="next02-detail-loop-live-dot" aria-hidden="true" />
                      {{ connectionText }}
                    </span>
                    <span id="eventCountText" class="badge bg-light text-dark">{{ rawEvents.length }} 条</span>
                    <label class="next02-detail-loop-raw-toggle" for="detailLoopDebugRaw">
                      <input id="detailLoopDebugRaw" v-model="showRawEvents" type="checkbox">
                      <span>显示原始 JSON</span>
                    </label>
                  </div>
                </div>
                <div class="result-box">
                  <pre id="resultBox"><template v-if="renderedEvents.length"><span
                    v-for="entry in renderedEvents"
                    :key="entry.key"
                    class="event-line"
                    :class="`log-${entry.level}`"
                  >{{ entry.content }}</span></template><span v-else class="event-empty">{{ connectionState === "connected" ? "等待详情抓取日志..." : "正在连接事件流..." }}</span></pre>
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
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LoadingState from "@/components/LoadingState.vue";
import PageShell from "@/components/PageShell.vue";
import ScDateTime from "@/components/ScDateTime.vue";
import type { AnyRecord } from "@/types/gc";
import { asRecord, pick, text } from "@/utils/format";

const route = useRoute();
const page = ref<AnyRecord>({});
const status = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const message = ref("");
const detailLoopEventHistorySize = 500;
const rawEvents = ref<AnyRecord[]>([]);
const showRawEvents = ref(false);
const connectionState = ref<"connecting" | "connected" | "reconnecting" | "closed">("connecting");
const actionPending = ref(false);
let eventSource: EventSource | undefined;
let eventReconnectTimer: number | undefined;
let statusPoller: number | undefined;

watch(() => route.fullPath, load, { immediate: true });

onBeforeUnmount(() => {
  closeEventStream();
  if (statusPoller) {
    window.clearInterval(statusPoller);
    statusPoller = undefined;
  }
});

async function load() {
  loading.value = true;
  error.value = "";
  try {
    page.value = asRecord(await getJSON(apiPathForRoute(route)));
    await refreshStatus();
    connectEventStream();
    ensureStatusPolling();
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
const startedAtValue = computed(() => pick(status.value, "started_at", "StartedAt"));
const lastEventAtValue = computed(() => pick(status.value, "last_event_at", "LastEventAt"));
const startedAt = computed(() => formatTimestamp(startedAtValue.value));
const lastEventAt = computed(() => formatTimestamp(lastEventAtValue.value));
const buffered = computed(() => text(status.value, "buffered", "Buffered") || "0");
const pausedText = computed(() => Boolean(pick(status.value, "paused", "Paused")) ? "是" : "否");
const lastLine = computed(() => text(status.value, "last_line", "LastLine"));
const connectionText = computed(() => ({
  connecting: "正在连接",
  connected: "实时连接",
  reconnecting: "正在重连",
  closed: "连接关闭"
})[connectionState.value]);
const connectionMessage = computed(() => connectionState.value === "connected" ? "事件流已连接" : connectionText.value);
const renderedEvents = computed(() => rawEvents.value.map((event, index) => ({
  key: `${eventSignature(event)}:${index}`,
  level: normalizeLevel(pick(event, "level", "Level"), text(event, "line", "Line")),
  content: showRawEvents.value ? JSON.stringify(event, null, 2) : formatEvent(event)
})));

async function start() {
  await call("/api/crawler/detail-loop/start");
}

async function stop() {
  await call("/api/crawler/detail-loop/stop");
}

async function call(path: string) {
  actionPending.value = true;
  try {
    const data = asRecord(await sendJSON(path, "POST", {}));
    message.value = text(data, "message", "Message") || "操作完成";
    await refreshStatus();
  } catch (err) {
    message.value = err instanceof Error ? err.message : String(err);
  } finally {
    actionPending.value = false;
  }
}

async function refreshStatus() {
  status.value = asRecord(await getJSON("/api/crawler/detail-loop"));
}

function ensureStatusPolling() {
  if (statusPoller) {
    return;
  }
  statusPoller = window.setInterval(async () => {
    try {
      await refreshStatus();
    } catch {
      // Keep the current snapshot visible when one refresh fails.
    }
  }, 5000);
}

function connectEventStream() {
  closeEventStream();
  connectionState.value = "connecting";
  const source = new EventSource("/api/crawler/detail-loop/stream");
  eventSource = source;
  source.onopen = () => {
    if (eventSource !== source) {
      return;
    }
    rawEvents.value = [];
    connectionState.value = "connected";
  };
  source.onmessage = (event) => {
    if (eventSource !== source) {
      return;
    }
    try {
      const incoming = asRecord(JSON.parse(event.data));
      rawEvents.value = [incoming, ...rawEvents.value].slice(0, detailLoopEventHistorySize);
      status.value = {
        ...status.value,
        last_event_at: pick(incoming, "at", "At"),
        last_line: text(incoming, "line", "Line") || text(incoming, "message", "Message")
      };
    } catch {
      message.value = "事件解析失败";
    }
  };
  source.onerror = () => {
    if (eventSource !== source) {
      return;
    }
    source.close();
    eventSource = undefined;
    connectionState.value = "reconnecting";
    scheduleEventReconnect();
  };
}

function closeEventStream() {
  if (eventSource) {
    eventSource.close();
    eventSource = undefined;
  }
  if (eventReconnectTimer) {
    window.clearTimeout(eventReconnectTimer);
    eventReconnectTimer = undefined;
  }
}

function scheduleEventReconnect() {
  if (eventReconnectTimer) {
    return;
  }
  eventReconnectTimer = window.setTimeout(() => {
    eventReconnectTimer = undefined;
    connectEventStream();
  }, 1500);
}

function formatTimestamp(value: unknown): string {
  const stamp = Number(value || 0);
  if (!stamp) {
    return "-";
  }
  const date = new Date(stamp * 1000);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  const pad = (part: number) => String(part).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

function eventTime(value: unknown): string {
  const stamp = Number(value || 0);
  const date = stamp > 0 ? new Date(stamp * 1000) : new Date();
  const pad = (part: number) => String(part).padStart(2, "0");
  return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

function normalizeLevel(value: unknown, line = ""): "info" | "warn" | "error" {
  const level = String(value || "").toLowerCase();
  if (level.includes("error") || /\sERROR\s/i.test(line)) {
    return "error";
  }
  if (level.includes("warn") || /\sWARN(?:ING)?\s/i.test(line)) {
    return "warn";
  }
  return "info";
}

function formatEvent(event: AnyRecord): string {
  const line = text(event, "line", "Line").trim();
  const content = line || text(event, "message", "Message") || "详情抓取事件";
  if (/^\[[^\]]+\]\s+(?:INFO|WARN|WARNING|ERROR)\s+/i.test(content)) {
    return content.replace(/^(\[[^\]]+\])\s+(?:INFO|WARN|WARNING|ERROR)\s+/i, "$1 ");
  }
  return `[${eventTime(pick(event, "at", "At"))}] ${content}`;
}

function eventSignature(event: AnyRecord): string {
  return `${Number(pick(event, "at", "At") || 0)}:${text(event, "level", "Level")}:${text(event, "line", "Line")}:${text(event, "message", "Message")}`;
}
</script>
