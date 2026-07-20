<template>
  <PageShell
    class="remote-list-page"
    eyebrow="DATA"
    title="周期统计"
    note="查询 c_rank_period_stat 中各周期的数据质量与榜单变化统计。"
  >
    <template #actions
      ><button class="button-secondary" type="button" :disabled="loading" @click="loadRows">刷新</button></template
    ><LoadingState v-if="loading" /><ErrorPanel v-else-if="error" :message="error" />
    <section v-else class="panel remote-list-card">
      <header class="remote-list-toolbar">
        <form class="remote-list-filter remote-list-filter-wide" @submit.prevent="applyFilters">
          <label class="remote-list-filter-field remote-list-number-field"
            ><span>周期 ID</span><input v-model.trim="periodID" type="number" min="1" placeholder="全部" /></label
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
      <EmptyState v-if="!rows.length" title="没有匹配的周期统计" message="完成周期排行计算后会自动生成。" />
      <div v-else class="data-table-wrap remote-list-table-wrap">
        <table class="data-table remote-list-table rank-period-stats-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>周期 ID</th>
              <th>上期 ID</th>
              <th>源行数</th>
              <th>源天数</th>
              <th>源影片数</th>
              <th>缺失天数</th>
              <th>榜单影片数</th>
              <th>连续</th>
              <th>首次</th>
              <th>回归</th>
              <th>退出</th>
              <th>上升</th>
              <th>下降</th>
              <th>不变</th>
              <th>创新高</th>
              <th>第一名</th>
              <th>平均得分</th>
              <th>平均上榜天数</th>
              <th>源开始时间</th>
              <th>源结束时间</th>
              <th>Checksum</th>
              <th>计算时间</th>
              <th>创建时间</th>
              <th>更新时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>#{{ row.id }}</td>
              <td>
                <RouterLink :to="`/rankperioditemlist?period_id=${row.period_id}`">#{{ row.period_id }}</RouterLink>
              </td>
              <td>{{ row.prev_period_id ? `#${row.prev_period_id}` : "—" }}</td>
              <td>{{ row.source_rank_row_count }}</td>
              <td>{{ row.source_rank_day_count }}</td>
              <td>{{ row.source_movie_count }}</td>
              <td>{{ row.missing_rank_day_count }}</td>
              <td>{{ row.period_movie_count }}</td>
              <td>{{ row.continued_movie_count }}</td>
              <td>{{ row.first_seen_movie_count }}</td>
              <td>{{ row.returning_movie_count }}</td>
              <td>{{ row.dropped_movie_count }}</td>
              <td>{{ row.rise_movie_count }}</td>
              <td>{{ row.fall_movie_count }}</td>
              <td>{{ row.unchanged_movie_count }}</td>
              <td>{{ row.peak_movie_count }}</td>
              <td class="dailybest-movie-table-cell">
                <MovieBasicCell :jav-id="row.rank_one_movie_jav_id" :movie="row.rank_one_movie" />
              </td>
              <td>{{ formatDecimal(row.average_score) }}</td>
              <td>{{ formatDecimal(row.average_days_in_rank) }}</td>
              <td>{{ formatDateTime(row.source_fetched_start_on) }}</td>
              <td>{{ formatDateTime(row.source_fetched_end_on) }}</td>
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
import { DEFAULT_LIST_PAGE_SIZE, listRankPeriodStats } from "@/features/dailybest/api";
import MovieBasicCell from "@/features/dailybest/components/MovieBasicCell.vue";
import type { PageInfo, RankPeriodStat, RankPeriodStatFilters } from "@/features/dailybest/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import { errorMessage, formatDateTime, formatDecimal } from "@/shared/utils/format";
import { compactPageQuery, pageInfoOrFallback, pageQueryFromRoute } from "@/shared/utils/pagination";
const route = useRoute(),
  router = useRouter(),
  rows = ref<RankPeriodStat[]>([]),
  pageInfo = ref<PageInfo>({ total: 0, page: 1, page_size: DEFAULT_LIST_PAGE_SIZE, page_total: 1 }),
  periodID = ref(""),
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
  return { period_id: pos(periodID.value) };
}
function current(): RankPeriodStatFilters {
  const p = pageQueryFromRoute(route.query);
  return { ...filters(), p: p.p, ps: p.ps } as RankPeriodStatFilters;
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
    const r = await listRankPeriodStats(current());
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
    periodID.value = q("period_id");
    pageSizeInput.value = String(pageQueryFromRoute(route.query).ps);
    void loadRows();
  },
  { immediate: true }
);
</script>
