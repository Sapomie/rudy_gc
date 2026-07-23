<template>
  <PageShell
    class="remote-list-page"
    eyebrow="MONITOR"
    title="目标站探测"
    note="查看真实 Chromium 对 DailyBest 目标页面的定时可访问性检查。"
  >
    <template #actions>
      <RouterLink class="button-secondary" to="/dailybestruntimestatus">查看运行状态</RouterLink>
      <button class="button-secondary" type="button" :disabled="loading" @click="loadProbes">刷新</button>
    </template>

    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error">
      <button class="button-secondary" type="button" @click="loadProbes">重新加载</button>
    </ErrorPanel>
    <section v-else class="panel remote-list-card">
      <header class="remote-list-toolbar">
        <form class="remote-list-filter remote-list-filter-wide" @submit.prevent="applyFilters">
          <label class="remote-list-filter-field">
            <span>状态</span>
            <select v-model="status">
              <option value="">全部</option>
              <option value="running">探测中</option>
              <option value="succeeded">成功</option>
              <option value="failed">失败</option>
              <option value="interrupted">已中断</option>
            </select>
          </label>
          <label class="remote-list-filter-field remote-list-date-field">
            <span>开始时间起</span>
            <input v-model="startedFrom" type="datetime-local" />
          </label>
          <label class="remote-list-filter-field remote-list-date-field">
            <span>开始时间止</span>
            <input v-model="startedTo" type="datetime-local" />
          </label>
          <label class="remote-list-filter-field remote-list-page-size-field">
            <span>每页</span>
            <input v-model.trim="pageSizeInput" type="number" min="1" />
          </label>
          <button class="remote-list-filter-submit" type="submit">查询</button>
          <button class="remote-list-filter-clear" type="button" @click="clearFilters">清空</button>
        </form>
        <div class="remote-list-metrics" aria-label="目标站探测摘要">
          <span
            ><strong>{{ pageInfo.total }}</strong> Total</span
          >
          <span
            ><strong>{{ probes.length }}</strong> Current Page</span
          >
        </div>
        <PaginationBar :page-info="pageInfo" />
      </header>

      <EmptyState v-if="!queried" title="尚未查询" message="设置筛选条件后点击查询，或点击刷新加载探测记录。" />
      <EmptyState v-else-if="!probes.length" title="没有匹配的探测记录" message="调整查询条件后重新查看。" />
      <div v-else class="data-table-wrap remote-list-table-wrap">
        <table class="data-table remote-list-table target-probes-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>请求 ID</th>
              <th>状态</th>
              <th>HTTP 状态码</th>
              <th>耗时</th>
              <th>浏览器状态</th>
              <th>目标地址</th>
              <th>最终地址</th>
              <th>错误类型</th>
              <th>错误信息</th>
              <th>开始时间</th>
              <th>完成时间</th>
              <th>创建时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="probe in probes" :key="probe.id">
              <td class="number-cell">#{{ probe.id }}</td>
              <td class="mono-value" :title="probe.request_id">{{ probe.request_id }}</td>
              <td><StatusBadge :label="formatProbeStatus(probe.status)" :tone="statusTone(probe.status)" /></td>
              <td class="number-cell">{{ probe.http_status || "—" }}</td>
              <td class="number-cell">{{ probe.duration_ms }} ms</td>
              <td>
                <StatusBadge
                  :label="formatRuntimeState(probe.browser_status)"
                  :tone="statusTone(probe.browser_status)"
                />
              </td>
              <td class="mono-value" :title="probe.target_url">{{ probe.target_url }}</td>
              <td class="mono-value" :title="probe.final_url || undefined">{{ probe.final_url || "—" }}</td>
              <td>{{ probe.error_type || "—" }}</td>
              <td class="error-value" :title="probe.error_message || undefined">{{ probe.error_message || "—" }}</td>
              <td class="number-cell">{{ formatDateTime(probe.started_at) }}</td>
              <td class="number-cell">{{ formatDateTime(probe.finished_at) }}</td>
              <td class="number-cell">{{ formatDateTime(probe.created_on) }}</td>
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
import { DEFAULT_LIST_PAGE_SIZE, listTargetProbes } from "@/features/dailybest/api";
import StatusBadge from "@/features/dailybest/components/StatusBadge.vue";
import type { PageInfo, TargetProbe, TargetProbeFilters } from "@/features/dailybest/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import { errorMessage, formatDateTime, formatProbeStatus, formatRuntimeState, statusTone } from "@/shared/utils/format";
import { compactPageQuery, pageInfoOrFallback, pageQueryFromRoute } from "@/shared/utils/pagination";

const route = useRoute();
const router = useRouter();
const probes = ref<TargetProbe[]>([]);
const pageInfo = ref<PageInfo>({ total: 0, page: 1, page_size: DEFAULT_LIST_PAGE_SIZE, page_total: 1 });
const status = ref("");
const startedFrom = ref("");
const startedTo = ref("");
const pageSizeInput = ref(String(DEFAULT_LIST_PAGE_SIZE));
const queried = ref(false);
const loading = ref(false);
const error = ref("");

function queryValue(name: string): string {
  const value = route.query[name];
  return Array.isArray(value) ? value[0] || "" : value || "";
}

function positiveInteger(value: string): number | undefined {
  const number = Number(value);
  return Number.isInteger(number) && number > 0 ? number : undefined;
}

function inputToUnix(value: string): number | undefined {
  if (!value) return undefined;
  const timestamp = new Date(value).getTime();
  return Number.isFinite(timestamp) ? Math.floor(timestamp / 1000) : undefined;
}

function unixToInput(value: string): string {
  const seconds = positiveInteger(value);
  if (!seconds) return "";
  const date = new Date(seconds * 1000);
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 16);
}

function filterQuery(): LocationQueryRaw {
  return {
    status: status.value || undefined,
    started_from: inputToUnix(startedFrom.value),
    started_to: inputToUnix(startedTo.value)
  };
}

function currentFilters(): TargetProbeFilters {
  const page = pageQueryFromRoute(route.query);
  return { ...filterQuery(), p: page.p, ps: page.ps } as TargetProbeFilters;
}

function syncFiltersFromRoute(): void {
  status.value = queryValue("status");
  startedFrom.value = unixToInput(queryValue("started_from"));
  startedTo.value = unixToInput(queryValue("started_to"));
  pageSizeInput.value = String(pageQueryFromRoute(route.query).ps);
}

async function applyFilters(): Promise<void> {
  const pageSize = positiveInteger(pageSizeInput.value) || DEFAULT_LIST_PAGE_SIZE;
  const currentPath = route.fullPath;
  queried.value = true;
  await router.replace({ path: route.path, query: compactPageQuery(filterQuery(), 1, pageSize) });
  if (route.fullPath === currentPath) await loadProbes();
}

async function clearFilters(): Promise<void> {
  const pageSize = pageQueryFromRoute(route.query).ps;
  queried.value = false;
  probes.value = [];
  pageInfo.value = { total: 0, page: 1, page_size: pageSize, page_total: 1 };
  error.value = "";
  await router.replace({ path: route.path, query: compactPageQuery({}, 1, pageSize) });
}

async function loadProbes(): Promise<void> {
  const page = pageQueryFromRoute(route.query);
  queried.value = true;
  loading.value = true;
  error.value = "";
  try {
    const response = await listTargetProbes(currentFilters());
    probes.value = response.probes;
    pageInfo.value = pageInfoOrFallback(response.page_info, page, response.probes.length);
  } catch (reason) {
    error.value = errorMessage(reason);
  } finally {
    loading.value = false;
  }
}

syncFiltersFromRoute();
watch(
  () => route.fullPath,
  () => {
    syncFiltersFromRoute();
    if (queried.value) void loadProbes();
  }
);
</script>
