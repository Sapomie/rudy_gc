<template>
  <PageShell
    class="remote-list-page"
    eyebrow="MONITOR"
    title="运行状态"
    note="查看 Remote 自身组件和目标网站的独立运行状态。"
  >
    <template #actions>
      <RouterLink class="button-secondary" to="/dailybesttargetprobelist">查看目标站探测</RouterLink>
      <button class="button-secondary" type="button" :disabled="statusLoading" @click="loadStatus">刷新状态</button>
    </template>

    <LoadingState v-if="statusLoading" :lines="3" />
    <ErrorPanel v-else-if="statusError" :message="statusError">
      <button class="button-secondary" type="button" @click="loadStatus">重新加载</button>
    </ErrorPanel>
    <EmptyState v-else-if="!statusLoaded" title="尚未加载运行状态" message="点击刷新状态后读取 Remote 当前运行信息。" />
    <template v-else-if="runtimeStatus">
      <section class="overview-grid">
        <article class="stat-card">
          <span class="stat-label">API</span>
          <strong>{{ formatRuntimeState(runtimeStatus.api_status) }}</strong>
          <StatusBadge
            :label="formatRuntimeState(runtimeStatus.database_status)"
            :tone="statusTone(runtimeStatus.database_status)"
          />
        </article>
        <article class="stat-card">
          <span class="stat-label">Browser Solver</span>
          <strong>{{ formatRuntimeState(runtimeStatus.browser_status.state) }}</strong>
          <StatusBadge
            :label="formatRuntimeState(runtimeStatus.browser_status.state)"
            :tone="statusTone(runtimeStatus.browser_status.state)"
          />
        </article>
        <article class="stat-card">
          <span class="stat-label">目标网站</span>
          <strong>{{ formatRuntimeState(runtimeStatus.target_status) }}</strong>
          <StatusBadge
            :label="formatRuntimeState(runtimeStatus.target_status)"
            :tone="statusTone(runtimeStatus.target_status)"
          />
        </article>
        <article class="stat-card">
          <span class="stat-label">下一次探测</span>
          <strong>{{ formatDateTime(runtimeStatus.next_probe_at) }}</strong>
          <span class="stat-note">由服务端调度配置决定</span>
        </article>
      </section>
      <section class="panel">
        <header class="panel-head"><h2>运行详情</h2></header>
        <div class="detail-grid">
          <div>
            <span>Solver 版本</span><strong>{{ runtimeStatus.browser_status.version || "—" }}</strong>
          </div>
          <div>
            <span>最近状态检查</span
            ><strong>{{ formatDateTime(runtimeStatus.browser_status.last_heartbeat_at) }}</strong>
          </div>
          <div>
            <span>最近成功探测</span><strong>{{ formatDateTime(runtimeStatus.last_probe_success) }}</strong>
          </div>
          <div>
            <span>最近失败探测</span><strong>{{ formatDateTime(runtimeStatus.last_probe_failure) }}</strong>
          </div>
          <div class="wide-row">
            <span>User-Agent</span
            ><strong class="mono-value">{{ runtimeStatus.browser_status.user_agent || "—" }}</strong>
          </div>
          <div class="wide-row">
            <span>Solver 错误</span
            ><strong class="error-value">{{
              runtimeStatus.browser_error || runtimeStatus.browser_status.last_error || "—"
            }}</strong>
          </div>
        </div>
      </section>
    </template>

    <section class="panel remote-list-card runtime-events-card">
      <header class="remote-list-toolbar">
        <form class="remote-list-filter remote-list-filter-wide" @submit.prevent="applyEventFilters">
          <label class="remote-list-filter-field">
            <span>组件</span>
            <select v-model="eventComponent">
              <option value="">全部</option>
              <option value="api">API</option>
              <option value="browser_agent">Browser Solver</option>
              <option value="target">目标网站</option>
            </select>
          </label>
          <label class="remote-list-filter-field">
            <span>级别</span>
            <select v-model="eventLevel">
              <option value="">全部</option>
              <option value="info">信息</option>
              <option value="error">错误</option>
            </select>
          </label>
          <label class="remote-list-filter-field remote-list-page-size-field">
            <span>每页</span>
            <input v-model.trim="pageSizeInput" type="number" min="1" />
          </label>
          <button class="remote-list-filter-submit" type="submit">查询事件</button>
          <button class="remote-list-filter-clear" type="button" @click="clearEventFilters">清空</button>
        </form>
        <div class="remote-list-metrics" aria-label="运行事件摘要">
          <span
            ><strong>{{ pageInfo.total }}</strong> Total</span
          >
          <span
            ><strong>{{ events.length }}</strong> Current Page</span
          >
        </div>
        <PaginationBar :page-info="pageInfo" />
      </header>
      <LoadingState v-if="eventsLoading" />
      <ErrorPanel v-else-if="eventsError" :message="eventsError" />
      <EmptyState
        v-else-if="!eventsQueried"
        title="尚未查询运行事件"
        message="点击查询事件后读取数据库中的状态变化记录。"
      />
      <EmptyState v-else-if="!events.length" title="没有匹配的运行事件" message="调整查询条件后重新查看。" />
      <div v-else class="data-table-wrap remote-list-table-wrap">
        <table class="data-table remote-list-table runtime-events-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>组件</th>
              <th>事件</th>
              <th>级别</th>
              <th>消息</th>
              <th>任务 ID</th>
              <th>探测 ID</th>
              <th>发生时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="event in events" :key="event.id">
              <td class="number-cell">#{{ event.id }}</td>
              <td>{{ event.component }}</td>
              <td>{{ event.event_type }}</td>
              <td>
                <StatusBadge
                  :label="event.level === 'error' ? '错误' : '信息'"
                  :tone="event.level === 'error' ? 'danger' : 'info'"
                />
              </td>
              <td :class="{ 'error-value': event.level === 'error' }" :title="event.message">{{ event.message }}</td>
              <td class="number-cell">{{ event.fetch_run_id || "—" }}</td>
              <td class="number-cell">{{ event.target_probe_id || "—" }}</td>
              <td class="number-cell">{{ formatDateTime(event.occurred_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { useRoute, useRouter, type LocationQueryRaw } from "vue-router";
import { DEFAULT_LIST_PAGE_SIZE, listRuntimeEvents, loadRuntimeStatus } from "@/features/dailybest/api";
import StatusBadge from "@/features/dailybest/components/StatusBadge.vue";
import type { PageInfo, RuntimeEvent, RuntimeEventFilters, RuntimeStatus } from "@/features/dailybest/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import { errorMessage, formatDateTime, formatRuntimeState, statusTone } from "@/shared/utils/format";
import { compactPageQuery, pageInfoOrFallback, pageQueryFromRoute } from "@/shared/utils/pagination";

const route = useRoute();
const router = useRouter();
const runtimeStatus = ref<RuntimeStatus | null>(null);
const statusLoaded = ref(false);
const statusLoading = ref(false);
const statusError = ref("");
const events = ref<RuntimeEvent[]>([]);
const eventsQueried = ref(false);
const eventsLoading = ref(false);
const eventsError = ref("");
const eventComponent = ref("");
const eventLevel = ref("");
const pageSizeInput = ref(String(DEFAULT_LIST_PAGE_SIZE));
const pageInfo = ref<PageInfo>({ total: 0, page: 1, page_size: DEFAULT_LIST_PAGE_SIZE, page_total: 1 });

function queryValue(name: string): string {
  const value = route.query[name];
  return Array.isArray(value) ? value[0] || "" : value || "";
}

function positiveInteger(value: string): number | undefined {
  const number = Number(value);
  return Number.isInteger(number) && number > 0 ? number : undefined;
}

function eventFilterQuery(): LocationQueryRaw {
  return { component: eventComponent.value || undefined, level: eventLevel.value || undefined };
}

function currentEventFilters(): RuntimeEventFilters {
  const page = pageQueryFromRoute(route.query);
  return { ...eventFilterQuery(), p: page.p, ps: page.ps } as RuntimeEventFilters;
}

function syncEventFilters(): void {
  eventComponent.value = queryValue("component");
  eventLevel.value = queryValue("level");
  pageSizeInput.value = String(pageQueryFromRoute(route.query).ps);
}

async function loadStatus(): Promise<void> {
  statusLoaded.value = true;
  statusLoading.value = true;
  statusError.value = "";
  try {
    runtimeStatus.value = await loadRuntimeStatus();
  } catch (reason) {
    statusError.value = errorMessage(reason);
  } finally {
    statusLoading.value = false;
  }
}

async function applyEventFilters(): Promise<void> {
  const pageSize = positiveInteger(pageSizeInput.value) || DEFAULT_LIST_PAGE_SIZE;
  const currentPath = route.fullPath;
  eventsQueried.value = true;
  await router.replace({ path: route.path, query: compactPageQuery(eventFilterQuery(), 1, pageSize) });
  if (route.fullPath === currentPath) await loadEvents();
}

async function clearEventFilters(): Promise<void> {
  const pageSize = pageQueryFromRoute(route.query).ps;
  eventsQueried.value = false;
  events.value = [];
  pageInfo.value = { total: 0, page: 1, page_size: pageSize, page_total: 1 };
  eventsError.value = "";
  await router.replace({ path: route.path, query: compactPageQuery({}, 1, pageSize) });
}

async function loadEvents(): Promise<void> {
  const page = pageQueryFromRoute(route.query);
  eventsQueried.value = true;
  eventsLoading.value = true;
  eventsError.value = "";
  try {
    const response = await listRuntimeEvents(currentEventFilters());
    events.value = response.events;
    pageInfo.value = pageInfoOrFallback(response.page_info, page, response.events.length);
  } catch (reason) {
    eventsError.value = errorMessage(reason);
  } finally {
    eventsLoading.value = false;
  }
}

syncEventFilters();
watch(
  () => route.fullPath,
  () => {
    syncEventFilters();
    if (eventsQueried.value) void loadEvents();
  }
);
</script>
