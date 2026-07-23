<template>
  <PageShell
    class="remote-list-page"
    eyebrow="COLLECTOR"
    title="请求日志"
    note="查看 DailyBest 每一次网站请求的执行结果。"
  >
    <template #actions>
      <RouterLink class="button-secondary" to="/dailybestfetchrunlist">查看抓取任务</RouterLink>
      <button class="button-secondary" type="button" :disabled="loading" @click="loadLogs">刷新</button>
    </template>

    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error">
      <button class="button-secondary" type="button" @click="loadLogs">重新加载</button>
    </ErrorPanel>
    <section v-else class="panel remote-list-card">
      <header class="remote-list-toolbar">
        <form class="remote-list-filter remote-list-filter-wide" @submit.prevent="applyFilters">
          <label class="remote-list-filter-field remote-list-number-field">
            <span>任务 ID</span>
            <input v-model.trim="fetchRunID" type="number" min="1" placeholder="全部" />
          </label>
          <label class="remote-list-filter-field remote-list-number-field">
            <span>页面</span>
            <input v-model.trim="pageNumber" type="number" min="1" placeholder="全部" />
          </label>
          <label class="remote-list-filter-field">
            <span>状态</span>
            <select v-model="status">
              <option value="">全部</option>
              <option value="requesting">请求中</option>
              <option value="succeeded">成功</option>
              <option value="failed">失败</option>
            </select>
          </label>
          <label class="remote-list-filter-field remote-list-page-size-field">
            <span>每页</span>
            <input v-model.trim="pageSizeInput" type="number" min="1" />
          </label>
          <button class="remote-list-filter-submit" type="submit">查询</button>
          <button class="remote-list-filter-clear" type="button" @click="clearFilters">清空</button>
        </form>
        <div class="remote-list-metrics" aria-label="请求日志摘要">
          <span
            ><strong>{{ pageInfo.total }}</strong> Total</span
          >
          <span
            ><strong>{{ logs.length }}</strong> Current Page</span
          >
        </div>
        <PaginationBar :page-info="pageInfo" />
      </header>

      <EmptyState v-if="!queried" title="尚未查询" message="设置筛选条件后点击查询，或点击刷新加载请求日志。" />
      <EmptyState v-else-if="!logs.length" title="没有匹配的请求日志" message="调整查询条件后重新查看。" />
      <div v-else class="data-table-wrap remote-list-table-wrap">
        <table class="data-table remote-list-table fetch-runs-table fetch-logs-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>任务 ID</th>
              <th>页面</th>
              <th>请求次数</th>
              <th>状态</th>
              <th>HTTP 状态码</th>
              <th>耗时</th>
              <th>请求地址</th>
              <th>错误</th>
              <th>创建时间</th>
              <th>更新时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="logRow in logs" :key="logRow.id">
              <td class="number-cell">#{{ logRow.id }}</td>
              <td class="number-cell">
                <RouterLink class="remote-list-primary-link" :to="`/dailybestfetchrundetail/${logRow.fetch_run_id}`">
                  #{{ logRow.fetch_run_id }}
                </RouterLink>
              </td>
              <td class="number-cell">{{ logRow.page_number }}</td>
              <td class="number-cell">{{ logRow.attempt_number }}</td>
              <td>
                <StatusBadge :label="formatFetchLogStatus(logRow.status)" :tone="statusTone(logRow.status)" />
              </td>
              <td class="number-cell">{{ logRow.http_status || "—" }}</td>
              <td class="number-cell">{{ logRow.duration_ms }} ms</td>
              <td class="mono-value" :title="logRow.request_url">{{ logRow.request_url }}</td>
              <td class="error-value" :title="logRow.error_message || undefined">{{ logRow.error_message || "—" }}</td>
              <td class="number-cell">{{ formatDateTime(logRow.created_on) }}</td>
              <td class="number-cell">{{ formatDateTime(logRow.updated_on) }}</td>
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
import { DEFAULT_LIST_PAGE_SIZE, listFetchLogs } from "@/features/dailybest/api";
import StatusBadge from "@/features/dailybest/components/StatusBadge.vue";
import type { FetchLog, FetchLogFilters, PageInfo } from "@/features/dailybest/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import { errorMessage, formatDateTime, formatFetchLogStatus, statusTone } from "@/shared/utils/format";
import { compactPageQuery, pageInfoOrFallback, pageQueryFromRoute } from "@/shared/utils/pagination";

const route = useRoute();
const router = useRouter();
const logs = ref<FetchLog[]>([]);
const pageInfo = ref<PageInfo>({ total: 0, page: 1, page_size: DEFAULT_LIST_PAGE_SIZE, page_total: 1 });
const fetchRunID = ref("");
const pageNumber = ref("");
const status = ref("");
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

function filterQuery(): LocationQueryRaw {
  return {
    fetch_run_id: positiveInteger(fetchRunID.value),
    page_number: positiveInteger(pageNumber.value),
    status: status.value || undefined
  };
}

function currentFilters(): FetchLogFilters {
  const page = pageQueryFromRoute(route.query);
  return { ...filterQuery(), p: page.p, ps: page.ps } as FetchLogFilters;
}

function syncFiltersFromRoute(): void {
  fetchRunID.value = queryValue("fetch_run_id");
  pageNumber.value = queryValue("page_number");
  status.value = queryValue("status");
  pageSizeInput.value = String(pageQueryFromRoute(route.query).ps);
}

async function applyFilters(): Promise<void> {
  const pageSize = positiveInteger(pageSizeInput.value) || DEFAULT_LIST_PAGE_SIZE;
  const currentPath = route.fullPath;
  queried.value = true;
  await router.replace({ path: route.path, query: compactPageQuery(filterQuery(), 1, pageSize) });
  if (route.fullPath === currentPath) await loadLogs();
}

async function clearFilters(): Promise<void> {
  const pageSize = pageQueryFromRoute(route.query).ps;
  queried.value = false;
  logs.value = [];
  pageInfo.value = { total: 0, page: 1, page_size: pageSize, page_total: 1 };
  error.value = "";
  await router.replace({ path: route.path, query: compactPageQuery({}, 1, pageSize) });
}

async function loadLogs(): Promise<void> {
  const page = pageQueryFromRoute(route.query);
  queried.value = true;
  loading.value = true;
  error.value = "";
  try {
    const response = await listFetchLogs(currentFilters());
    logs.value = response.logs;
    pageInfo.value = pageInfoOrFallback(response.page_info, page, response.logs.length);
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
    if (queried.value) void loadLogs();
  }
);
</script>
