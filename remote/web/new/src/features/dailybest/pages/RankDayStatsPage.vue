<template>
  <PageShell
    class="remote-list-page"
    eyebrow="DATA"
    title="每日统计"
    note="查询 c_rank_day_stat 中按排行日重建的变化统计。"
  >
    <template #actions
      ><button class="button-secondary" type="button" :disabled="loading" @click="loadRows">刷新</button></template
    >
    <LoadingState v-if="loading" /><ErrorPanel v-else-if="error" :message="error" />
    <section v-else class="panel remote-list-card">
      <header class="remote-list-toolbar">
        <form class="remote-list-filter remote-list-filter-wide" @submit.prevent="applyFilters">
          <label class="remote-list-filter-field remote-list-date-field"
            ><span>排行日期</span><input v-model="rankDateFilter" type="date" /></label
          ><label class="remote-list-filter-field remote-list-number-field"
            ><span>日序号</span><input v-model.trim="dayNumber" type="number" min="1" placeholder="全部" /></label
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
      <EmptyState v-if="!rows.length" title="没有匹配的每日统计" message="完成排名抓取后会自动生成。" />
      <div v-else class="data-table-wrap remote-list-table-wrap">
        <table class="data-table remote-list-table rank-day-stats-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>日期</th>
              <th>日序号</th>
              <th>上一排行日</th>
              <th>排行行数</th>
              <th>影片数</th>
              <th>连续</th>
              <th>首次</th>
              <th>回归</th>
              <th>掉榜</th>
              <th>上升</th>
              <th>下降</th>
              <th>不变</th>
              <th>Checksum</th>
              <th>计算时间</th>
              <th>创建时间</th>
              <th>更新时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>#{{ row.id }}</td>
              <td>{{ row.rank_date }}</td>
              <td>{{ row.day_number }}</td>
              <td>{{ row.prev_day_number || "—" }}</td>
              <td>{{ row.rank_row_count }}</td>
              <td>{{ row.movie_count }}</td>
              <td>{{ row.continued_movie_count }}</td>
              <td>{{ row.first_seen_movie_count }}</td>
              <td>{{ row.returning_movie_count }}</td>
              <td>{{ row.dropped_movie_count }}</td>
              <td>{{ row.rise_movie_count }}</td>
              <td>{{ row.fall_movie_count }}</td>
              <td>{{ row.unchanged_movie_count }}</td>
              <td class="checksum-cell" :title="row.source_checksum">{{ row.source_checksum }}</td>
              <td>{{ formatDateTime(row.calculated_on) }}</td>
              <td>{{ formatDateTime(row.created_on) }}</td>
              <td>{{ formatDateTime(row.updated_on) }}</td>
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
import { DEFAULT_LIST_PAGE_SIZE, listRankDayStats } from "@/features/dailybest/api";
import type { PageInfo, RankDayStat, RankDayStatFilters } from "@/features/dailybest/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import { errorMessage, formatDateTime } from "@/shared/utils/format";
import { compactPageQuery, pageInfoOrFallback, pageQueryFromRoute } from "@/shared/utils/pagination";
const route = useRoute(),
  router = useRouter(),
  rows = ref<RankDayStat[]>([]),
  pageInfo = ref<PageInfo>({ total: 0, page: 1, page_size: DEFAULT_LIST_PAGE_SIZE, page_total: 1 }),
  rankDateFilter = ref(""),
  dayNumber = ref(""),
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
  return { rank_date: rankDateFilter.value || undefined, day_number: pos(dayNumber.value) };
}
function current(): RankDayStatFilters {
  const p = pageQueryFromRoute(route.query);
  return { ...filters(), p: p.p, ps: p.ps } as RankDayStatFilters;
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
    const r = await listRankDayStats(current());
    rows.value = r.stats;
    pageInfo.value = pageInfoOrFallback(r.page_info, p, r.stats.length);
  } catch (e) {
    error.value = errorMessage(e);
  } finally {
    loading.value = false;
  }
}
watch(
  () => route.fullPath,
  () => {
    rankDateFilter.value = q("rank_date");
    dayNumber.value = q("day_number");
    pageSizeInput.value = String(pageQueryFromRoute(route.query).ps);
    void loadRows();
  },
  { immediate: true }
);
</script>
