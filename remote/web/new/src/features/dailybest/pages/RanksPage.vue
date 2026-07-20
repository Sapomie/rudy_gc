<template>
  <PageShell class="remote-list-page" eyebrow="DATA" title="排名明细" note="查询 c_rank 中按业务日期保存的原始排名。">
    <template #actions
      ><button class="button-secondary" type="button" :disabled="loading" @click="loadRows">刷新</button></template
    >
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error"
      ><button class="button-secondary" type="button" @click="loadRows">重新加载</button></ErrorPanel
    >
    <section v-else class="panel remote-list-card">
      <header class="remote-list-toolbar">
        <form class="remote-list-filter remote-list-filter-wide" @submit.prevent="applyFilters">
          <label class="remote-list-filter-field remote-list-date-field"
            ><span>业务日期</span><input v-model="businessDate" type="date"
          /></label>
          <label class="remote-list-filter-field remote-list-number-field"
            ><span>日序号</span><input v-model.trim="dayNumber" type="number" min="1" placeholder="全部"
          /></label>
          <label class="remote-list-filter-field remote-list-number-field"
            ><span>分类</span><input v-model.trim="category" type="number" min="1" placeholder="全部"
          /></label>
          <label class="remote-list-filter-field remote-list-rank-field"
            ><span>排名</span><input v-model.trim="rankPos" type="number" min="1" placeholder="全部"
          /></label>
          <label class="remote-list-filter-field"
            ><span>影片 JavID</span><input v-model.trim="movieJavID" type="search" placeholder="包含查询"
          /></label>
          <label class="remote-list-filter-field remote-list-page-size-field"
            ><span>每页</span><input v-model.trim="pageSizeInput" type="number" min="1"
          /></label>
          <button class="remote-list-filter-submit" type="submit">筛选</button
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
      <EmptyState v-if="!rows.length" title="没有匹配的排名记录" message="调整查询条件后重新查看。" />
      <div v-else class="data-table-wrap remote-list-table-wrap">
        <table class="data-table remote-list-table ranks-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>排名键</th>
              <th>业务日期</th>
              <th>日序号</th>
              <th>分类</th>
              <th>排名</th>
              <th>影片</th>
              <th>创建时间</th>
              <th>更新时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td class="number-cell">#{{ row.id }}</td>
              <td class="mono-value" :title="row.rank_key">{{ row.rank_key }}</td>
              <td>{{ rankDate(row.day_number) }}</td>
              <td class="number-cell">{{ row.day_number }}</td>
              <td>{{ formatCategory(row.category) }}</td>
              <td class="rank-cell">{{ row.rank_pos }}</td>
              <td class="dailybest-movie-table-cell">
                <MovieBasicCell :jav-id="row.movie_jav_id" :movie="row.movie" />
              </td>
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
import { DEFAULT_LIST_PAGE_SIZE, listRanks } from "@/features/dailybest/api";
import MovieBasicCell from "@/features/dailybest/components/MovieBasicCell.vue";
import type { PageInfo, Rank, RankFilters } from "@/features/dailybest/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import { errorMessage, formatCategory, formatDateTime, rankDate } from "@/shared/utils/format";
import { compactPageQuery, pageInfoOrFallback, pageQueryFromRoute } from "@/shared/utils/pagination";
const route = useRoute(),
  router = useRouter();
const rows = ref<Rank[]>([]),
  pageInfo = ref<PageInfo>({ total: 0, page: 1, page_size: DEFAULT_LIST_PAGE_SIZE, page_total: 1 });
const businessDate = ref(""),
  dayNumber = ref(""),
  category = ref(""),
  rankPos = ref(""),
  movieJavID = ref(""),
  pageSizeInput = ref(String(DEFAULT_LIST_PAGE_SIZE)),
  loading = ref(false),
  error = ref("");
const q = (name: string) => {
  const value = route.query[name];
  return Array.isArray(value) ? value[0] || "" : value || "";
};
const pos = (value: string) => {
  const n = Number(value);
  return Number.isInteger(n) && n > 0 ? n : undefined;
};
function sync() {
  businessDate.value = q("business_date");
  dayNumber.value = q("day_number");
  category.value = q("category");
  rankPos.value = q("rank_pos");
  movieJavID.value = q("movie_jav_id");
  pageSizeInput.value = String(pageQueryFromRoute(route.query).ps);
}
function filters(): LocationQueryRaw {
  return {
    business_date: businessDate.value || undefined,
    day_number: pos(dayNumber.value),
    category: pos(category.value),
    rank_pos: pos(rankPos.value),
    movie_jav_id: movieJavID.value || undefined
  };
}
function current(): RankFilters {
  const page = pageQueryFromRoute(route.query);
  return { ...filters(), p: page.p, ps: page.ps } as RankFilters;
}
async function applyFilters() {
  const size = pos(pageSizeInput.value) || DEFAULT_LIST_PAGE_SIZE,
    currentPath = route.fullPath;
  await router.replace({ path: route.path, query: compactPageQuery(filters(), 1, size) });
  if (route.fullPath === currentPath) await loadRows();
}
async function clearFilters() {
  const size = pageQueryFromRoute(route.query).ps,
    currentPath = route.fullPath;
  await router.replace({ path: route.path, query: compactPageQuery({}, 1, size) });
  if (route.fullPath === currentPath) await loadRows();
}
async function loadRows() {
  const page = pageQueryFromRoute(route.query);
  loading.value = true;
  error.value = "";
  try {
    const response = await listRanks(current());
    rows.value = response.ranks;
    pageInfo.value = pageInfoOrFallback(response.page_info, page, response.ranks.length);
  } catch (reason) {
    error.value = errorMessage(reason);
  } finally {
    loading.value = false;
  }
}
watch(
  () => route.fullPath,
  () => {
    sync();
    void loadRows();
  },
  { immediate: true }
);
</script>
