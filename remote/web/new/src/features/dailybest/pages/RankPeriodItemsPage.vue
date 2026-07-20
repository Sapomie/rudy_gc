<template>
  <PageShell
    class="remote-list-page"
    eyebrow="DATA"
    title="周期明细"
    note="查询 c_rank_period_item 中各周期的影片名次、得分和变化。"
  >
    <template #actions
      ><button class="button-secondary" type="button" :disabled="loading" @click="loadRows">刷新</button></template
    >
    <LoadingState v-if="loading" /><ErrorPanel v-else-if="error" :message="error" />
    <section v-else class="panel remote-list-card">
      <header class="remote-list-toolbar">
        <form class="remote-list-filter remote-list-filter-wide" @submit.prevent="applyFilters">
          <label class="remote-list-filter-field remote-list-number-field"
            ><span>周期 ID</span><input v-model.trim="periodID" type="number" min="1" placeholder="全部" /></label
          ><label class="remote-list-filter-field"
            ><span>影片 JavID</span><input v-model.trim="movieJavID" type="search" placeholder="包含查询" /></label
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
      <EmptyState v-if="!rows.length" title="没有匹配的周期明细" message="选择周期或调整影片条件后重新查看。" />
      <div v-else class="data-table-wrap remote-list-table-wrap">
        <table class="data-table remote-list-table rank-period-items-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>周期 ID</th>
              <th>排名</th>
              <th>影片</th>
              <th>得分</th>
              <th>上榜天数</th>
              <th>取样天数</th>
              <th>首次日</th>
              <th>最后日</th>
              <th>最佳日排名</th>
              <th>样本最差排名</th>
              <th>上期排名</th>
              <th>排名变化</th>
              <th>历史峰值</th>
              <th>创建时间</th>
              <th>更新时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>#{{ row.id }}</td>
              <td>#{{ row.period_id }}</td>
              <td class="rank-cell">{{ row.rank_pos }}</td>
              <td class="dailybest-movie-table-cell">
                <MovieBasicCell :jav-id="row.movie_jav_id" :movie="row.movie" />
              </td>
              <td>{{ formatDecimal(row.score) }}</td>
              <td>{{ row.days_in_rank }}</td>
              <td>{{ row.used_pick_days }}</td>
              <td>{{ row.first_rank_day_number }} / {{ rankDate(row.first_rank_day_number) }}</td>
              <td>{{ row.last_rank_day_number }} / {{ rankDate(row.last_rank_day_number) }}</td>
              <td>{{ row.best_rank }}</td>
              <td>{{ row.worst_picked_rank }}</td>
              <td>{{ row.prev_rank || "—" }}</td>
              <td :class="rankChangeClass(row.rank_change)">{{ formatRankChange(row.rank_change) }}</td>
              <td>{{ row.peak_rank }}</td>
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
import { DEFAULT_LIST_PAGE_SIZE, listRankPeriodItems } from "@/features/dailybest/api";
import MovieBasicCell from "@/features/dailybest/components/MovieBasicCell.vue";
import type { PageInfo, RankPeriodItem, RankPeriodItemFilters } from "@/features/dailybest/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import {
  errorMessage,
  formatDateTime,
  formatDecimal,
  formatRankChange,
  rankChangeClass,
  rankDate
} from "@/shared/utils/format";
import { compactPageQuery, pageInfoOrFallback, pageQueryFromRoute } from "@/shared/utils/pagination";
const route = useRoute(),
  router = useRouter(),
  rows = ref<RankPeriodItem[]>([]),
  pageInfo = ref<PageInfo>({ total: 0, page: 1, page_size: DEFAULT_LIST_PAGE_SIZE, page_total: 1 }),
  periodID = ref(""),
  movieJavID = ref(""),
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
  return { period_id: pos(periodID.value), movie_jav_id: movieJavID.value || undefined };
}
function current(): RankPeriodItemFilters {
  const p = pageQueryFromRoute(route.query);
  return { ...filters(), p: p.p, ps: p.ps } as RankPeriodItemFilters;
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
    const r = await listRankPeriodItems(current());
    rows.value = r.items;
    pageInfo.value = pageInfoOrFallback(r.page_info, p, r.items.length);
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
    movieJavID.value = q("movie_jav_id");
    pageSizeInput.value = String(pageQueryFromRoute(route.query).ps);
    void loadRows();
  },
  { immediate: true }
);
</script>
