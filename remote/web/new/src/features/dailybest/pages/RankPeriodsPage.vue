<template>
  <PageShell
    class="remote-list-page"
    eyebrow="DATA"
    title="周期排行"
    note="查询 c_rank_period 中的周、月、季度和年度排行周期。"
  >
    <template #actions
      ><button class="button-secondary" type="button" :disabled="loading" @click="loadRows">刷新</button></template
    ><LoadingState v-if="loading" /><ErrorPanel v-else-if="error" :message="error" />
    <section v-else class="panel remote-list-card">
      <header class="remote-list-toolbar">
        <form class="remote-list-filter remote-list-filter-wide" @submit.prevent="applyFilters">
          <label class="remote-list-filter-field"
            ><span>周期类型</span
            ><select v-model="periodType">
              <option value="">全部</option>
              <option value="1">周</option>
              <option value="2">月</option>
              <option value="3">季度</option>
              <option value="4">年</option>
            </select></label
          ><label class="remote-list-filter-field"
            ><span>周期键</span><input v-model.trim="periodKey" type="search" placeholder="例如 2026-07" /></label
          ><label class="remote-list-filter-field remote-list-number-field"
            ><span>分类</span><input v-model.trim="category" type="number" min="1" placeholder="全部" /></label
          ><label class="remote-list-filter-field remote-list-page-size-field"
            ><span>每页</span><input v-model.trim="pageSizeInput" type="number" min="1" /></label
          ><button class="remote-list-filter-submit" type="submit">筛选</button
          ><button class="remote-list-filter-clear" type="button" @click="clearFilters">清空</button>
        </form>
        <div class="remote-list-metrics">
          <span
            ><strong>{{ pageInfo.total }}</strong> Total</span
          ><span
            ><strong>{{ rows.length }}</strong> Current Page</span
          >
        </div>
        <PaginationBar :page-info="pageInfo" />
      </header>
      <EmptyState v-if="!rows.length" title="没有匹配的周期排行" message="完成排名抓取后会自动生成。" />
      <div v-else class="data-table-wrap remote-list-table-wrap">
        <table class="data-table remote-list-table rank-periods-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>类型</th>
              <th>周期键</th>
              <th>分类</th>
              <th>开始日</th>
              <th>结束日</th>
              <th>取样天数</th>
              <th>Top N</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>更新时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>#{{ row.id }}</td>
              <td>{{ formatPeriodType(row.period_type) }}</td>
              <td>{{ row.period_key }}</td>
              <td>{{ formatCategory(row.category) }}</td>
              <td>{{ row.start_day_number }} / {{ rankDate(row.start_day_number) }}</td>
              <td>{{ row.end_day_number }} / {{ rankDate(row.end_day_number) }}</td>
              <td>{{ row.pick_days }}</td>
              <td>{{ row.top_n }}</td>
              <td>{{ formatPeriodStatus(row.status) }}</td>
              <td>{{ formatDateTime(row.created_on) }}</td>
              <td>{{ formatDateTime(row.updated_on) }}</td>
              <td>
                <RouterLink class="table-action" :to="`/rankperioditemlist?period_id=${row.id}`">查看明细</RouterLink>
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
import { useRoute, useRouter, type LocationQueryRaw } from "vue-router";
import { DEFAULT_LIST_PAGE_SIZE, listRankPeriods } from "@/features/dailybest/api";
import type { PageInfo, RankPeriod, RankPeriodFilters } from "@/features/dailybest/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import {
  errorMessage,
  formatCategory,
  formatDateTime,
  formatPeriodStatus,
  formatPeriodType,
  rankDate
} from "@/shared/utils/format";
import { compactPageQuery, pageInfoOrFallback, pageQueryFromRoute } from "@/shared/utils/pagination";
const route = useRoute(),
  router = useRouter(),
  rows = ref<RankPeriod[]>([]),
  pageInfo = ref<PageInfo>({ total: 0, page: 1, page_size: DEFAULT_LIST_PAGE_SIZE, page_total: 1 }),
  periodType = ref(""),
  periodKey = ref(""),
  category = ref(""),
  pageSizeInput = ref(String(DEFAULT_LIST_PAGE_SIZE)),
  loading = ref(false),
  error = ref("");
const q = (n: string) => {
    const v = route.query[n];
    return Array.isArray(v) ? v[0] || "" : v || "";
  },
  pos = (v: string) => {
    const n = Number(v);
    return Number.isInteger(n) && n > 0 ? n : undefined;
  };
function filters(): LocationQueryRaw {
  return {
    period_type: pos(periodType.value),
    period_key: periodKey.value || undefined,
    category: pos(category.value)
  };
}
function current(): RankPeriodFilters {
  const p = pageQueryFromRoute(route.query);
  return { ...filters(), p: p.p, ps: p.ps } as RankPeriodFilters;
}
async function applyFilters() {
  const size = pos(pageSizeInput.value) || DEFAULT_LIST_PAGE_SIZE,
    path = route.fullPath;
  await router.replace({ path: route.path, query: compactPageQuery(filters(), 1, size) });
  if (path === route.fullPath) await loadRows();
}
async function clearFilters() {
  const size = pageQueryFromRoute(route.query).ps,
    path = route.fullPath;
  await router.replace({ path: route.path, query: compactPageQuery({}, 1, size) });
  if (path === route.fullPath) await loadRows();
}
async function loadRows() {
  const p = pageQueryFromRoute(route.query);
  loading.value = true;
  error.value = "";
  try {
    const r = await listRankPeriods(current());
    rows.value = r.periods;
    pageInfo.value = pageInfoOrFallback(r.page_info, p, r.periods.length);
  } catch (e) {
    error.value = errorMessage(e);
  } finally {
    loading.value = false;
  }
}
watch(
  () => route.fullPath,
  () => {
    periodType.value = q("period_type");
    periodKey.value = q("period_key");
    category.value = q("category");
    pageSizeInput.value = String(pageQueryFromRoute(route.query).ps);
    void loadRows();
  },
  { immediate: true }
);
</script>
