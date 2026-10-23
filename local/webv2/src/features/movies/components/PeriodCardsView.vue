<template>
  <PageShell
    class="cards-page-shell rank-page-shell next02-pink-cards-page next02-pink-cards-page-shell cards-period-page"
    :title="title"
    :note="note"
  >
    <template #actions>
      <a class="next02-pink-cards-hero-link" href="/aggbackfill">回填入口</a>
    </template>

    <div class="cards-period-layout" :class="{ 'is-collapsed': directoryCollapsed }" :aria-busy="loading">
      <aside
        v-show="!directoryCollapsed"
        id="cards-period-directory"
        class="table-card cards-period-directory"
        aria-label="时间周期"
      >
        <header class="table-card-head cards-period-directory-head rank-stats-toolbar-scope">
          <div class="cards-period-directory-title">
            <h2>时间周期</h2>
            <span v-if="payload && !error">{{ payload.PeriodTotal.toLocaleString() }} 期</span>
          </div>
          <form
            class="rank-date-form rank-stats-date-controls cards-period-page-size"
            @submit.prevent="changePageSize(true)"
          >
            <label class="rank-stats-filter-field">
              <span>每页</span>
              <input
                v-model="periodPageSize"
                class="rank-date-input rank-stats-page-size-input"
                aria-label="每页周期数"
                type="number"
                min="1"
                step="1"
                required
                :disabled="loading"
              />
            </label>
            <button class="button-secondary rank-toolbar-action" type="submit" :disabled="loading">设置</button>
          </form>
        </header>
        <div class="cards-period-type-control">
          <LinkStrip :items="periodTypeItems" />
        </div>
        <div class="cards-period-directory-body" :class="{ 'has-pagination': hasPeriodPagination }">
          <PaginationBar
            v-if="hasPeriodPagination && payload"
            class="cards-period-directory-pager"
            :page-info="payload.PeriodPageInfo"
            aria-label="周期分页"
          />
          <div class="table-scroll cards-period-table-wrap">
            <div v-if="loading && !payload" class="alert bucket-list-feedback is-loading">加载周期中...</div>
            <div v-else-if="error" class="movie-empty cards-period-directory-empty">周期加载失败</div>
            <table v-else class="data-table cards-period-table" :class="{ 'is-release-table': !isBirth }">
              <thead>
                <tr>
                  <th><SortLink :link="periodSortLink('time')" label="周期" /></th>
                  <th class="sc-number">
                    <SortLink :link="periodSortLink('count')" :label="isBirth ? '媒体数' : '影片数'" />
                  </th>
                  <th v-if="!isBirth" class="sc-number">
                    <SortLink :link="periodSortLink('owned')" label="已拥有" />
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="!payload?.Periods.length">
                  <td :colspan="isBirth ? 2 : 3">
                    <div class="movie-empty cards-period-directory-empty">暂无周期数据</div>
                  </td>
                </tr>
                <tr
                  v-for="period in payload?.Periods || []"
                  :key="period.Key"
                  :class="{ 'is-selected': period.Key === payload?.SelectedPeriod?.Key }"
                >
                  <td>
                    <a
                      :href="selectCardPeriod(basis, query, period.Key)"
                      :aria-current="period.Key === payload?.SelectedPeriod?.Key ? 'true' : undefined"
                      :title="`${period.RangeStart} 至 ${period.RangeEnd}`"
                      >{{ period.Label }}</a
                    >
                  </td>
                  <td class="sc-number">
                    <a
                      class="cards-period-count-link"
                      :href="periodCountHref(period)"
                      target="_blank"
                      rel="noopener noreferrer"
                      :title="`在新窗口查看 ${period.Label} ${isBirth ? '媒体' : '影片'}`"
                      >{{ period.Count.toLocaleString() }}</a
                    >
                  </td>
                  <td v-if="!isBirth" class="sc-number">
                    <a
                      class="cards-period-count-link"
                      :href="periodOwnedCountHref(period)"
                      target="_blank"
                      rel="noopener noreferrer"
                      :title="`在新窗口查看 ${period.Label} 已拥有影片`"
                      >{{ period.CountOwned?.toLocaleString() ?? "—" }}</a
                    >
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </aside>

      <button
        class="icon-button cards-period-directory-toggle"
        type="button"
        :aria-expanded="!directoryCollapsed"
        aria-controls="cards-period-directory"
        :aria-label="directoryCollapsed ? '展开时间周期' : '收起时间周期'"
        :title="directoryCollapsed ? '展开时间周期' : '收起时间周期'"
        @click="directoryCollapsed = !directoryCollapsed"
      >
        <span aria-hidden="true">›</span>
      </button>

      <section class="cards-period-results" aria-label="周期影片">
        <header class="cards-period-results-head">
          <div class="cards-period-results-heading">
            <small>{{ isBirth ? "下载周期" : "上映周期" }}</small>
            <div class="cards-period-results-title">
              <h2>{{ payload?.SelectedPeriod?.Label || "请选择周期" }}</h2>
              <span v-if="payload?.SelectedPeriod" class="cards-period-results-range">
                {{ payload.SelectedPeriod.RangeStart }} 至 {{ payload.SelectedPeriod.RangeEnd }}
              </span>
            </div>
          </div>
          <span v-if="payload && !error" class="cards-period-results-count"
            ><strong>{{ payload.total.toLocaleString() }}</strong> 部影片</span
          >
        </header>

        <div
          v-if="payload?.SelectedPeriod && !error"
          class="sticky-toolbar cards-toolbar card-results-toolbar next02-pink-cards-toolbar cards-period-toolbar rank-stats-toolbar-scope photo-results-toolbar"
          aria-label="周期卡片控制栏"
        >
          <div class="cards-toolbar-sort sort-controls">
            <MovieCardSortStrip :query="payload.sortQuery" :options="sortOptions" />
          </div>
          <div class="owned-line cards-toolbar-filters">
            <LinkStrip v-if="!isBirth" :items="ownedLinks" />
            <form
              class="rank-date-form rank-stats-date-controls cards-period-page-size"
              @submit.prevent="changePageSize(false)"
            >
              <label class="rank-stats-filter-field">
                <span>每页</span>
                <input
                  v-model="cardPageSize"
                  class="rank-date-input rank-stats-page-size-input"
                  aria-label="每页卡片数"
                  type="number"
                  min="1"
                  step="1"
                  required
                  :disabled="loading"
                />
              </label>
              <button class="button-secondary rank-toolbar-action" type="submit" :disabled="loading">设置</button>
            </form>
          </div>
          <div class="next02-pink-cards-toolbar-status">
            <CardGridSizeControl :disabled="loading" />
            <div class="next02-pink-cards-metrics" aria-label="周期卡片摘要">
              <span :aria-label="`总数 ${payload.total}`"
                ><strong>{{ payload.total }}</strong></span
              >
              <span :aria-label="`当前页 ${payload.movies.length}`"
                ><strong>{{ payload.movies.length }}</strong></span
              >
            </div>
            <div class="cards-toolbar-pagination next02-pink-cards-pagination">
              <PaginationBar :page-info="payload.PageInfo" aria-label="卡片分页" />
            </div>
          </div>
        </div>
        <LoadingState v-if="loading" />
        <ErrorPanel v-else-if="error" :message="error" />
        <template v-else-if="payload">
          <MovieGrid :cards="payload.movies" />
          <PaginationBar
            v-if="payload.total > 0"
            class="cards-period-bottom-pager"
            :page-info="payload.PageInfo"
            aria-label="卡片底部分页"
          />
        </template>
      </section>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import "@/features/movies/styles/index.css";
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { apiPathForRoute, getJSON } from "@/features/movies/api";
import MovieGrid from "@/features/movies/components/MovieGrid.vue";
import MovieCardSortStrip from "@/features/movies/components/MovieCardSortStrip.vue";
import { movieCardSortOptions as sortOptions } from "@/features/movies/utils/movieCardSortOptions";
import {
  periodListCardHref,
  periodCardsHref,
  selectCardPeriod,
  sortCardPeriods,
  switchCardPeriodType,
  type CardPeriodBasis,
  type CardPeriodSort,
  type PeriodCardsResponse
} from "@/features/movies/utils/periodCards";
import PageShell from "@/shared/components/PageShell.vue";
import CardGridSizeControl from "@/shared/components/CardGridSizeControl.vue";
import LinkStrip from "@/shared/components/LinkStrip.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import { asRecord, pick, text } from "@/shared/utils/format";

const props = defineProps<{ basis: CardPeriodBasis; title: string; note: string }>();
const route = useRoute();
const router = useRouter();
const payload = ref<PeriodCardsResponse | null>(null);
const loading = ref(false);
const error = ref("");
const directoryCollapsed = ref(false);
const periodPageSize = ref<number | string>(18);
const cardPageSize = ref<number | string>(18);
let requestID = 0;

watch(() => route.fullPath, load, { immediate: true });
onBeforeUnmount(() => {
  requestID++;
});

async function load() {
  const current = ++requestID;
  loading.value = true;
  error.value = "";
  try {
    const response = await getJSON<PeriodCardsResponse>(apiPathForRoute(route));
    if (current !== requestID) return;
    payload.value = response;
    periodPageSize.value = response.PeriodPageInfo.PageSize || 18;
    cardPageSize.value = response.PageInfo.PageSize || 18;
  } catch (err) {
    if (current === requestID) error.value = err instanceof Error ? err.message : String(err);
  } finally {
    if (current === requestID) loading.value = false;
  }
}

const query = computed<PeriodCardsResponse["Query"]>(() => {
  if (payload.value && !loading.value && !error.value) return payload.value.Query;
  const values: PeriodCardsResponse["Query"] = {};
  for (const [key, value] of Object.entries(route.query || {})) {
    values[key] = (Array.isArray(value) ? value : [value]).filter((item) => item != null).map(String);
  }
  return values;
});
const isBirth = computed(() => props.basis === "birth");
const ownedLinks = computed(() => {
  const ownedQuery = asRecord(pick(payload.value, "ownedQuery"));
  const labels: Record<string, string> = {
    MediaAll: "All",
    MediaOwned: "Owned",
    MediaNotOwned: "NotOwned"
  };
  return Object.entries(labels).map(([key, label]) => {
    const item = asRecord(ownedQuery[key]);
    return {
      label,
      href: text(item, "Href") || route.fullPath,
      active: Boolean(item.Active),
      disabled: loading.value
    };
  });
});
const hasPeriodPagination = computed(() =>
  Boolean(payload.value && !error.value && payload.value.PeriodTotal > 0 && payload.value.PeriodPageInfo.Links?.length)
);
const periodTypeItems = computed(() =>
  [
    { value: "year", label: "年" },
    { value: "quarter", label: "季度" },
    { value: "month", label: "月" },
    { value: "day", label: "日" }
  ].map((option) => ({
    label: option.label,
    href: switchCardPeriodType(props.basis, query.value, option.value),
    active: (query.value.period_type?.[0] || "day") === option.value,
    disabled: loading.value
  }))
);
function periodSortLink(sort: CardPeriodSort) {
  return {
    Href: sortCardPeriods(props.basis, query.value, sort),
    Active: (query.value.period_sort?.[0] || "time") === sort,
    Desc: (query.value.period_order?.[0] || "desc") === "desc",
    Disabled: loading.value
  };
}

function periodCountHref(period: PeriodCardsResponse["Periods"][number]): string {
  return periodListCardHref(props.basis, period);
}

function periodOwnedCountHref(period: PeriodCardsResponse["Periods"][number]): string {
  return periodListCardHref(props.basis, period, true);
}

function changePageSize(periods: boolean) {
  const size = Number(periods ? periodPageSize.value : cardPageSize.value);
  if (loading.value || !Number.isSafeInteger(size) || size <= 0) return;
  const changes: Record<string, string | number | null> = periods
    ? { period_ps: size, period_p: 1, key: null, p: 1 }
    : { ps: size, p: 1 };
  void router.push(periodCardsHref(props.basis, query.value, changes));
}
</script>
