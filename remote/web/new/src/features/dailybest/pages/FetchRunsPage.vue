<template>
  <PageShell class="remote-list-page" eyebrow="COLLECTOR" title="抓取任务" note="查看定时任务和手动任务的执行结果。">
    <template #actions>
      <RouterLink class="button-primary" to="/dailybestfetch">新建抓取</RouterLink>
      <button class="button-secondary" type="button" :disabled="loading" @click="loadRuns">刷新</button>
    </template>

    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <section v-else class="panel remote-list-card">
      <header class="remote-list-toolbar">
        <form class="remote-list-filter" @submit.prevent="applyPageSize">
          <label class="remote-list-filter-field remote-list-page-size-field">
            <span>每页</span>
            <input v-model.trim="pageSizeInput" type="number" name="ps" min="1" />
          </label>
          <button class="remote-list-filter-submit" type="submit">应用</button>
        </form>
        <div class="remote-list-metrics" aria-label="抓取任务摘要">
          <span
            ><strong>{{ pageInfo.total }}</strong> Total</span
          >
          <span
            ><strong>{{ runs.length }}</strong> Current Page</span
          >
        </div>
        <PaginationBar :page-info="pageInfo" />
      </header>

      <EmptyState v-if="!runs.length" title="暂无抓取任务" message="创建任务后会在这里显示运行记录。">
        <RouterLink class="button-primary" to="/dailybestfetch">新建抓取</RouterLink>
      </EmptyState>
      <div v-else class="data-table-wrap remote-list-table-wrap">
        <table class="data-table remote-list-table fetch-runs-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>业务日期</th>
              <th>榜单</th>
              <th>触发方式</th>
              <th>状态</th>
              <th>开始时间</th>
              <th>完成时间</th>
              <th>规模</th>
              <th>Checksum</th>
              <th>创建时间</th>
              <th>更新时间</th>
              <th>错误</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="run in runs" :key="run.id">
              <td class="number-cell">
                <RouterLink class="remote-list-primary-link" :to="`/dailybestfetchrundetail/${run.id}`"
                  >#{{ run.id }}</RouterLink
                >
              </td>
              <td class="number-cell">{{ run.business_date }}</td>
              <td>{{ formatCategory(run.category) }}</td>
              <td>{{ formatTriggerType(run.trigger_type) }}</td>
              <td><StatusBadge :label="formatRunStatus(run.status)" :tone="statusTone(run.status)" /></td>
              <td class="number-cell">{{ formatDateTime(run.started_at) }}</td>
              <td class="number-cell">{{ formatDateTime(run.finished_at) }}</td>
              <td class="number-cell">{{ run.page_count }} 页 / {{ run.item_count }} 条</td>
              <td class="checksum-cell" :title="run.checksum">{{ run.checksum || "—" }}</td>
              <td class="number-cell">{{ formatDateTime(run.created_at) }}</td>
              <td class="number-cell">{{ formatDateTime(run.updated_at) }}</td>
              <td class="error-value" :title="run.error_message || undefined">{{ run.error_message || "—" }}</td>
              <td>
                <RouterLink class="table-action" :to="`/dailybestfetchrundetail/${run.id}`">详情</RouterLink>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { DEFAULT_LIST_PAGE_SIZE, listFetchRuns } from "@/features/dailybest/api";
import StatusBadge from "@/features/dailybest/components/StatusBadge.vue";
import type { FetchRun, PageInfo } from "@/features/dailybest/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import { compactPageQuery, pageInfoOrFallback, pageQueryFromRoute } from "@/shared/utils/pagination";
import {
  errorMessage,
  formatCategory,
  formatDateTime,
  formatRunStatus,
  formatTriggerType,
  statusTone
} from "@/shared/utils/format";

const route = useRoute();
const router = useRouter();
const runs = ref<FetchRun[]>([]);
const pageInfo = ref<PageInfo>({ total: 0, page: 1, page_size: DEFAULT_LIST_PAGE_SIZE, page_total: 1 });
const pageSizeInput = ref(String(DEFAULT_LIST_PAGE_SIZE));
const loading = ref(false);
const error = ref("");

async function applyPageSize(): Promise<void> {
  const parsed = Number(pageSizeInput.value);
  const pageSize = Number.isInteger(parsed) && parsed > 0 ? parsed : DEFAULT_LIST_PAGE_SIZE;
  const currentPath = route.fullPath;
  await router.replace({ path: route.path, query: compactPageQuery(route.query, 1, pageSize) });
  if (route.fullPath === currentPath) await loadRuns();
}

async function loadRuns(): Promise<void> {
  const page = pageQueryFromRoute(route.query);
  pageSizeInput.value = String(page.ps);
  loading.value = true;
  error.value = "";
  try {
    const response = await listFetchRuns(page);
    runs.value = response.runs;
    pageInfo.value = pageInfoOrFallback(response.page_info, page, response.runs.length);
  } catch (reason) {
    error.value = errorMessage(reason);
  } finally {
    loading.value = false;
  }
}

watch(() => route.fullPath, loadRuns, { immediate: true });
</script>
