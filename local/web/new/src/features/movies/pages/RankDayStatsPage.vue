<template>
  <PageShell class="sc-events-page-shell" title="Rank Day Stats" note="" eyebrow="">
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <div
      v-else
      id="page-root"
      class="container-fluid sc-events-page-root next02-sc-pink-theme next02-sc-events-rewrite"
    >
      <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell">
        <div class="next02-sc-hero-copy">
          <span class="next02-sc-eyebrow">RANK / DAY STATS</span>
          <h4>{{ pageTitle }}</h4>
          <p>{{ pageNote }}</p>
        </div>
        <div class="next02-sc-hero-actions">
          <div class="btn-group" role="group" aria-label="排行统计视图">
            <a class="btn btn-outline-primary active disabled" href="/rankdaystats" aria-current="page">日统计</a>
            <a class="btn btn-outline-primary" href="/rankperiodstats">周期统计</a>
          </div>
          <a href="/listcarddayrank" class="btn btn-outline-secondary" target="_blank" rel="noopener noreferrer"
            >日排行</a
          >
        </div>
      </section>

      <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
        <div class="card-body">
          <div
            class="rank-stats-toolbar-scope rank-page-shell cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell"
          >
            <form
              class="sticky-toolbar cards-toolbar rank-sticky-toolbar next02-pink-cards-toolbar rank-stats-toolbar"
              method="get"
              action="/rankdaystats"
            >
              <input type="hidden" name="p" value="1" />
              <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')" />
              <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')" />
              <div class="cards-toolbar-sort rank-toolbar-sort">
                <div class="rank-date-form rank-stats-date-controls">
                  <label class="rank-stats-filter-field">
                    <span>起始日期</span>
                    <input
                      class="rank-date-input smart-pick-date-input"
                      type="text"
                      name="start_date"
                      :value="queryValue('StartDate')"
                      placeholder="YYYY-MM-DD"
                      inputmode="numeric"
                      maxlength="10"
                      pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                    />
                  </label>
                  <label class="rank-stats-filter-field">
                    <span>结束日期</span>
                    <input
                      class="rank-date-input smart-pick-date-input"
                      type="text"
                      name="end_date"
                      :value="queryValue('EndDate')"
                      placeholder="YYYY-MM-DD"
                      inputmode="numeric"
                      maxlength="10"
                      pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                    />
                  </label>
                  <label class="rank-stats-filter-field">
                    <span>每页</span>
                    <input
                      class="rank-date-input rank-stats-page-size-input"
                      type="number"
                      step="1"
                      name="ps"
                      :value="queryValue('PageSize')"
                    />
                  </label>
                  <div class="rank-navigation-group">
                    <button type="submit" class="button-secondary rank-toolbar-action">筛选</button>
                    <a :href="clearHref" class="button-secondary rank-toolbar-action">清空</a>
                  </div>
                </div>
              </div>

              <div v-if="hasRows" class="toolbar-line cards-toolbar-bottom rank-toolbar-line">
                <div class="owned-line cards-toolbar-filters rank-toolbar-filters"></div>
                <div class="next02-pink-cards-toolbar-status">
                  <div class="next02-pink-cards-metrics" aria-label="日排行统计摘要">
                    <span :title="`总数 ${formatCount(total)}`" :aria-label="`总数 ${formatCount(total)}`">
                      <strong>{{ formatCount(total) }}</strong>
                    </span>
                    <span :title="`当前页 ${rows.length}`" :aria-label="`当前页 ${rows.length}`">
                      <strong>{{ rows.length }}</strong>
                    </span>
                  </div>
                  <div class="cards-toolbar-pagination next02-pink-cards-pagination">
                    <PaginationBar :page-info="pageInfo" />
                  </div>
                </div>
              </div>
            </form>
          </div>

          <div class="table-responsive">
            <table class="table table-hover align-middle mb-0 sc-events-table">
              <thead>
                <tr>
                  <th class="sc-sortable-head"><SortLink class="sc-sort-link" :link="headerLink('DayNumber')" /></th>
                  <th class="sc-number sc-sortable-head">
                    <SortLink class="sc-sort-link" :link="headerLink('RankRowCount')" />
                  </th>
                  <th class="sc-number sc-sortable-head">
                    <SortLink class="sc-sort-link" :link="headerLink('MovieCount')" />
                  </th>
                  <th class="sc-number sc-sortable-head">
                    <SortLink class="sc-sort-link" :link="headerLink('ContinuedMovieCount')" />
                  </th>
                  <th class="sc-number sc-sortable-head">
                    <SortLink class="sc-sort-link" :link="headerLink('FirstSeenMovieCount')" />
                  </th>
                  <th class="sc-number sc-sortable-head">
                    <SortLink class="sc-sort-link" :link="headerLink('ReturningMovieCount')" />
                  </th>
                  <th class="sc-number sc-sortable-head">
                    <SortLink class="sc-sort-link" :link="headerLink('DroppedMovieCount')" />
                  </th>
                  <th>上期</th>
                  <th>抓取时间</th>
                  <th>同步时间</th>
                  <th class="sc-sortable-head">
                    <SortLink class="sc-sort-link" :link="headerLink('CalculatedOn')" />
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="!hasRows">
                  <td colspan="11"><div class="sc-empty-state">当前筛选下没有 `c_rank_day_stat` 数据。</div></td>
                </tr>
                <tr
                  v-for="(row, index) in rows"
                  v-else
                  :key="value(row, 'DayNumber')"
                  class="next02-sc-event-row"
                  :class="dayRowClass(row, index)"
                >
                  <td class="sc-content-cell">
                    <a
                      class="sc-event-link"
                      :href="`/listcarddayrank?date=${encodeURIComponent(value(row, 'RankDate'))}`"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <span class="sc-event-date">{{ value(row, "RankDate") }}</span>
                    </a>
                  </td>
                  <td class="sc-number sc-primary-value">{{ count(row, "RankRowCount") }}</td>
                  <td class="sc-number">{{ count(row, "MovieCount") }}</td>
                  <td class="sc-number">{{ count(row, "ContinuedMovieCount") }}</td>
                  <td class="sc-number">
                    <a class="sc-event-link" :href="firstSeenMoviesHref(row)" target="_blank" rel="noopener noreferrer">
                      {{ count(row, "FirstSeenMovieCount") }}
                    </a>
                  </td>
                  <td class="sc-number">{{ count(row, "ReturningMovieCount") }}</td>
                  <td class="sc-number">{{ count(row, "DroppedMovieCount") }}</td>
                  <td class="sc-content-cell">
                    <span class="sc-event-date">{{ previousRankDate(row) }}</span>
                  </td>
                  <td class="sc-muted"><ScDateTime :value="pick(row, 'FetchedOn')" unix /></td>
                  <td class="sc-muted"><ScDateTime :value="pick(row, 'SyncedOn')" unix /></td>
                  <td class="sc-muted"><ScDateTime :value="pick(row, 'CalculatedOn')" unix /></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import "@/features/movies/styles/index.css";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/features/movies/api";
import { formatRankStatCount, rankDateFromDayNumber } from "@/features/movies/components/stats/rankStats";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import type { AnyRecord, PageInfo } from "@/shared/types/gc";
import { asArray, asRecord, normalizeHref, num, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title") || "日排行统计");
const pageNote = computed(() => text(payload.value, "PageNote") || "直接查看 c_rank_day_stat 落库记录。");
const rows = computed(() => asArray<AnyRecord>(pick(payload.value, "Rows", "rows")));
const hasRows = computed(() => rows.value.length > 0);
const total = computed(() => num(payload.value, "Total", "total"));
const query = computed(() => asRecord(pick(payload.value, "Query", "query")));
const headerLinks = computed(() => asRecord(pick(payload.value, "HeaderLinks")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const clearHref = computed(() => normalizeHref(text(payload.value, "ClearHref") || "/rankdaystats"));

function queryValue(key: string): string {
  return text(query.value, key);
}

function headerLink(key: string): AnyRecord {
  return asRecord(headerLinks.value[key]);
}

function value(row: AnyRecord, key: string): string {
  return text(row, key);
}

function formatCount(value: unknown): string {
  return formatRankStatCount(value);
}

function count(row: AnyRecord, key: string): string {
  return formatCount(num(row, key));
}

function previousRankDate(row: AnyRecord): string {
  return rankDateFromDayNumber(num(row, "PrevDayNumber"));
}

function firstSeenMoviesHref(row: AnyRecord): string {
  const rankDate = value(row, "RankDate");
  const params = new URLSearchParams({ srds: rankDate, srde: rankDate });
  return `/listcardrank?${params.toString()}`;
}

function dayMonth(row: AnyRecord): string {
  return value(row, "RankDate").slice(0, 7);
}

function dayRowClass(row: AnyRecord, index: number): Record<string, boolean> {
  const month = dayMonth(row);
  const monthNumber = Number(month.slice(5, 7));
  const previousMonth = index > 0 ? dayMonth(rows.value[index - 1]) : "";
  return {
    "sc-running-month-neutral": !month,
    "sc-running-month-light": Boolean(month) && monthNumber % 2 === 0,
    "sc-running-month-deep": Boolean(month) && monthNumber % 2 !== 0,
    "sc-running-month-start": Boolean(month) && month !== previousMonth
  };
}
</script>
