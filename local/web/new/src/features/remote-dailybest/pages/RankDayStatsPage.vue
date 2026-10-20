<template>
  <PageShell
    class="remote-dailybest-page remote-list-page"
    eyebrow="DATA"
    title="每日统计"
    note="查询 c_rank_day_stat 中按排行日首次生成的变化统计。"
  >
    <template #actions
      ><button class="button-secondary" type="button" :disabled="loading" @click="loadRows">刷新</button></template
    >
    <LoadingState v-if="loading" /><ErrorPanel v-else-if="error" :message="error" />
    <section v-else class="panel remote-list-card">
      <header class="remote-list-toolbar">
        <form class="remote-list-filter remote-list-filter-wide" @submit.prevent="applyFilters">
          <label class="remote-list-filter-field remote-list-date-field"
            ><span>排行日期</span><input v-model="rankDateFilter" type="date" name="rank_date" /></label
          ><label class="remote-list-filter-field remote-list-number-field"
            ><span>日序号</span
            ><input v-model.trim="dayNumber" type="number" name="day_number" min="1" placeholder="全部" /></label
          ><label class="remote-list-filter-field remote-list-page-size-field"
            ><span>每页</span><input v-model.trim="pageSizeInput" type="number" name="ps" min="1" /></label
          ><button class="remote-list-filter-submit" type="submit">筛选</button
          ><button class="remote-list-filter-clear" type="button" @click="clearFilters">清空</button>
        </form>
        <div class="remote-list-metrics">
          <span
            ><strong>{{ pageInfo.total }}</strong> Total</span
          ><span
            ><strong>{{ rows.length }}</strong> Current Page</span
          ><span
            ><strong>{{ needsSyncCount }}</strong> 待同步</span
          ><span
            ><strong>{{ selectedCount }}</strong> 已选</span
          >
        </div>
        <button
          class="button-secondary remote-list-sync-action"
          type="button"
          :disabled="selectedCount === 0 || submitting"
          @click="syncSelected"
        >
          {{ submitting ? "正在提交" : "同步选中日期" }}
        </button>
        <PaginationBar :page-info="pageInfo" />
      </header>
      <div
        v-if="feedback"
        class="remote-list-feedback"
        :class="{ 'is-error': feedbackError }"
        :role="feedbackError ? 'alert' : 'status'"
      >
        <span>{{ feedback }}</span>
        <a v-if="jobHref" :href="jobHref" class="remote-list-primary-link">查看任务</a>
      </div>
      <span v-if="comparisonError" class="remote-list-inline-error" role="alert">
        本地排名状态读取失败：{{ comparisonError }}
      </span>
      <EmptyState v-if="!rows.length" title="没有匹配的每日统计" message="完成排名抓取后会自动生成。" />
      <div v-else class="data-table-wrap remote-list-table-wrap">
        <table class="data-table remote-list-table rank-day-stats-table">
          <colgroup>
            <col class="rank-day-stats-col-id" />
            <col class="rank-day-stats-col-date" />
            <col class="rank-day-stats-col-rank-count" />
            <col class="rank-day-stats-col-rank-count" />
            <col class="rank-day-stats-col-count" />
            <col class="rank-day-stats-col-count" />
            <col class="rank-day-stats-col-count" />
            <col class="rank-day-stats-col-count" />
            <col class="rank-day-stats-col-count" />
            <col class="rank-day-stats-col-status" />
            <col class="rank-day-stats-col-time" />
            <col class="rank-day-stats-col-time" />
            <col class="rank-day-stats-col-time" />
          </colgroup>
          <thead>
            <tr>
              <th>ID</th>
              <th>
                <label class="rank-day-stats-selection">
                  <input
                    type="checkbox"
                    aria-label="全选当前页待同步日期"
                    :checked="allSelected"
                    :indeterminate="someSelected"
                    :disabled="selectableRows.length === 0"
                    @change="toggleAll"
                  />
                  <span>日期</span>
                </label>
              </th>
              <th>排行行数</th>
              <th>本地排名</th>
              <th>影片数</th>
              <th>连续</th>
              <th>首次</th>
              <th>回归</th>
              <th>掉榜</th>
              <th>状态</th>
              <th>抓取时间</th>
              <th>同步时间</th>
              <th>计算时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.day_number">
              <td>#{{ row.id }}</td>
              <td>
                <label class="rank-day-stats-selection" :title="`日序号 ${row.day_number}`">
                  <input
                    type="checkbox"
                    :aria-label="`选择 ${row.rank_date} 排名`"
                    :checked="isSelected(row)"
                    :disabled="!rowNeedsSync(row)"
                    @change="toggleRow(row, $event)"
                  />
                  <span>{{ row.rank_date }}</span>
                </label>
              </td>
              <td>{{ row.rank_row_count }}</td>
              <td>{{ localRankCount(row) }}</td>
              <td>{{ row.movie_count }}</td>
              <td>{{ row.continued_movie_count }}</td>
              <td>{{ row.first_seen_movie_count }}</td>
              <td>{{ row.returning_movie_count }}</td>
              <td>{{ row.dropped_movie_count }}</td>
              <td>
                <StatusBadge :label="syncStatusLabel(row)" :tone="syncStatusTone(row)" />
              </td>
              <td><ScDateTime :value="row.fetched_at" unix seconds /></td>
              <td><ScDateTime :value="row.synced_at" unix seconds /></td>
              <td><ScDateTime :value="row.calculated_on" unix seconds /></td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </PageShell>
</template>
<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter, type LocationQueryRaw } from "vue-router";
import {
  DEFAULT_LIST_PAGE_SIZE,
  listLocalRankDayComparisons,
  listRankDayStats,
  syncLocalRankDays
} from "@/features/remote-dailybest/api";
import type { PageInfo, RankDayComparisonFilters, RankDayStatFilters } from "@/features/remote-dailybest/types";
import {
  collectRankDaySyncTargets,
  mergeRankDayStats,
  rankDayNeedsSync,
  type RankDayStatsRow
} from "@/features/remote-dailybest/rank-day-stats";
import EmptyState from "@/features/remote-dailybest/components/EmptyState.vue";
import ErrorPanel from "@/features/remote-dailybest/components/RemoteErrorPanel.vue";
import StatusBadge from "@/features/remote-dailybest/components/StatusBadge.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import PaginationBar from "@/features/remote-dailybest/components/RemotePaginationBar.vue";
import { errorMessage, formatRankDaySyncStatus, rankDaySyncStatusTone } from "@/features/remote-dailybest/format";
import { compactPageQuery, pageInfoOrFallback, pageQueryFromRoute } from "@/features/remote-dailybest/pagination";

const route = useRoute(),
  router = useRouter(),
  rows = ref<RankDayStatsRow[]>([]),
  pageInfo = ref<PageInfo>({ total: 0, page: 1, page_size: DEFAULT_LIST_PAGE_SIZE, page_total: 1 }),
  rankDateFilter = ref(""),
  dayNumber = ref(""),
  pageSizeInput = ref(String(DEFAULT_LIST_PAGE_SIZE)),
  loading = ref(false),
  error = ref(""),
  comparisonError = ref(""),
  submitting = ref(false),
  feedback = ref(""),
  feedbackError = ref(false),
  jobID = ref(""),
  selectedDayNumbers = ref<number[]>([]);
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
function comparisonFilters(value: RankDayStatFilters): RankDayComparisonFilters {
  return {
    p: value.p,
    ps: value.ps,
    business_date_start: value.rank_date,
    business_date_end: value.rank_date,
    day_number: value.day_number
  };
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
  if (path === route.fullPath) {
    syncFiltersFromRoute();
    await loadRows();
  }
}
async function loadRows() {
  const p = pageQueryFromRoute(route.query);
  const requestedFilters = current();
  loading.value = true;
  error.value = "";
  comparisonError.value = "";
  feedback.value = "";
  feedbackError.value = false;
  jobID.value = "";
  selectedDayNumbers.value = [];
  try {
    const [statsResponse, comparisonResult] = await Promise.all([
      listRankDayStats(requestedFilters),
      listLocalRankDayComparisons(comparisonFilters(requestedFilters)).then(
        (response) => ({ response, error: "" }),
        (reason: unknown) => ({ response: null, error: errorMessage(reason) })
      )
    ]);
    comparisonError.value = comparisonResult.error;
    rows.value = mergeRankDayStats(statsResponse.stats, comparisonResult.response?.Days || []);
    pageInfo.value = pageInfoOrFallback(statsResponse.page_info, p, statsResponse.stats.length);
  } catch (e) {
    rows.value = [];
    error.value = errorMessage(e);
  } finally {
    loading.value = false;
  }
}
function syncFiltersFromRoute() {
  rankDateFilter.value = q("rank_date");
  dayNumber.value = q("day_number");
  pageSizeInput.value = String(pageQueryFromRoute(route.query).ps);
}
syncFiltersFromRoute();
watch(
  () => route.fullPath,
  () => {
    syncFiltersFromRoute();
    void loadRows();
  },
  { immediate: true }
);

const selectableRows = computed(() => rows.value.filter(rowNeedsSync));
const needsSyncCount = computed(() => selectableRows.value.length);
const selectedCount = computed(() => selectedDayNumbers.value.length);
const allSelected = computed(
  () => selectableRows.value.length > 0 && selectableRows.value.every((row) => isSelected(row))
);
const someSelected = computed(() => selectedCount.value > 0 && !allSelected.value);
const jobHref = computed(() => (jobID.value ? `/dailybestfetch?job_id=${encodeURIComponent(jobID.value)}` : ""));

function rowNeedsSync(row: RankDayStatsRow): boolean {
  return rankDayNeedsSync(row);
}

function isSelected(row: RankDayStatsRow): boolean {
  return selectedDayNumbers.value.includes(row.day_number);
}

function toggleRow(row: RankDayStatsRow, event: Event): void {
  const selected = Boolean((event.target as HTMLInputElement | null)?.checked);
  if (selected && rowNeedsSync(row) && !selectedDayNumbers.value.includes(row.day_number)) {
    selectedDayNumbers.value = [...selectedDayNumbers.value, row.day_number];
  } else if (!selected) {
    selectedDayNumbers.value = selectedDayNumbers.value.filter((value) => value !== row.day_number);
  }
}

function toggleAll(event: Event): void {
  const selected = Boolean((event.target as HTMLInputElement | null)?.checked);
  selectedDayNumbers.value = selected ? selectableRows.value.map((row) => row.day_number) : [];
}

function selectedTargets() {
  return collectRankDaySyncTargets(rows.value, selectedDayNumbers.value);
}

async function syncSelected(): Promise<void> {
  const targets = selectedTargets();
  if (!targets.length) return;
  const selectedDays = selectedCount.value;
  submitting.value = true;
  feedback.value = "";
  feedbackError.value = false;
  jobID.value = "";
  try {
    const response = await syncLocalRankDays(targets);
    jobID.value = response.job_id;
    feedback.value = `已提交 ${selectedDays} 天的同步任务`;
    selectedDayNumbers.value = [];
  } catch (reason) {
    feedback.value = errorMessage(reason) || "同步任务提交失败";
    feedbackError.value = true;
  } finally {
    submitting.value = false;
  }
}

function localRankCount(row: RankDayStatsRow): number | string {
  return row.comparison?.local_rank_count ?? "—";
}

function syncStatusLabel(row: RankDayStatsRow): string {
  return row.comparison ? formatRankDaySyncStatus(row.comparison.status) : "—";
}

function syncStatusTone(row: RankDayStatsRow): string {
  return row.comparison ? rankDaySyncStatusTone(row.comparison.status) : "neutral";
}
</script>
