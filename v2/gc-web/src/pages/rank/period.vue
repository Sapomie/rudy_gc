<template>
  <PageFrame class="rgc-period-page" title="Period Rank" eyebrow="Ranking">
    <form class="rgc-panel rgc-rank-tools rgc-rank-compact rgc-rank-compact-period" @submit.prevent="apply">
      <div class="rgc-rank-current rgc-period-summary">
        <div class="rgc-period-title">
          <strong>{{ response ? `${response.period_type_label} ${response.period_key}` : "最新周期榜" }}</strong>
          <span v-if="response">{{ response.category_label }} · {{ response.range_start }} ~ {{ response.range_end }} · {{ response.total }} movies</span>
          <span v-else>读取中</span>
        </div>
      </div>

      <div class="rgc-rank-fields">
        <label>
          <span>Type</span>
          <select v-model="periodType">
            <option value="week">Week</option>
            <option value="month">Month</option>
            <option value="quarter">Quarter</option>
            <option value="year">Year</option>
          </select>
        </label>
        <label>
          <span>Category</span>
          <select v-model="category">
            <option value="1">MonthSource</option>
            <option value="2">AllTimeSource</option>
          </select>
        </label>
        <label class="wide">
          <span>Key</span>
          <input v-model.trim="periodKey" type="text" name="key" placeholder="留空取最新" />
        </label>
        <label>
          <span>Page Size</span>
          <input v-model="pageSize" type="number" name="ps" />
        </label>
      </div>

      <div v-if="response" class="rgc-inline-nav rgc-rank-segments">
        <RouterLink v-for="item in response.type_links" :key="item.href" :class="{ active: item.active }" :to="item.href">{{ item.label }}</RouterLink>
        <RouterLink v-for="item in response.category_links" :key="item.href" :class="{ active: item.active }" :to="item.href">{{ item.label }}</RouterLink>
      </div>

      <div class="rgc-rank-actions">
        <RouterLink v-if="response && response.prev_href" class="rgc-button compact" :to="response.prev_href">上一期 {{ response.prev_key }}</RouterLink>
        <span v-else class="rgc-button compact disabled">上一期</span>
        <RouterLink v-if="response && response.next_href" class="rgc-button compact" :to="response.next_href">下一期 {{ response.next_key }}</RouterLink>
        <span v-else class="rgc-button compact disabled">下一期</span>
        <RouterLink v-if="response && response.random_href" class="rgc-button compact" :to="response.random_href">随机</RouterLink>
        <RouterLink v-if="response" class="rgc-button compact primary" :to="response.latest_href || '/moviecardperiodrank'">最新</RouterLink>
        <button class="rgc-button compact primary" type="submit">筛选</button>
        <button class="rgc-button compact" type="button" @click="clear">清空</button>
      </div>
    </form>

    <StateBlock v-if="loading" title="加载中" message="正在读取周期榜。" />
    <StateBlock v-else-if="errorMessage" title="加载失败" :message="errorMessage" tone="error" />
    <template v-else-if="response">
      <div class="rgc-list-head rgc-rank-listbar">
        <span>{{ rankWindowText() }}</span>
        <PaginationControl :page="response.page" :page-size="response.page_size" :total="response.total" @go="goPage" />
      </div>

      <div v-if="response.items.length > 0" class="rgc-period-grid">
        <article v-for="item in response.items" :key="`${item.rank_pos}-${item.movie.movie_jav_id}`" class="rgc-period-card">
          <span class="rgc-rank-position">#{{ item.rank_pos }}</span>
          <MovieTile :movie="item.movie" />
          <div class="rgc-period-meta">
            <b :class="`tone-${item.rank_change_class}`">{{ item.rank_change_text }}</b>
            <small>上期 {{ item.prev_rank_text }} · 峰值 #{{ item.peak_rank }}</small>
            <small>最佳 #{{ item.best_rank }} · 最差 #{{ item.worst_picked_rank }}</small>
            <small>上榜 {{ item.days_in_rank }} 天 · 取样 {{ item.used_pick_days }} 天</small>
            <small>Score {{ item.score.toFixed(2) }}</small>
          </div>
        </article>
      </div>
      <StateBlock v-else title="无结果" message="当前周期没有排名数据。" tone="empty" />

      <div class="rgc-list-foot rgc-rank-listbar">
        <PaginationControl :page="response.page" :page-size="response.page_size" :total="response.total" @go="goPage" />
      </div>
    </template>
  </PageFrame>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { getRankPeriod } from "@/api/gc";
import MovieTile from "@/components/MovieTile.vue";
import PageFrame from "@/components/PageFrame.vue";
import PaginationControl from "@/components/PaginationControl.vue";
import StateBlock from "@/components/StateBlock.vue";
import type { RankPeriodResponse } from "@/types/api";

const route = useRoute();
const router = useRouter();
const response = ref<RankPeriodResponse | null>(null);
const loading = ref(false);
const errorMessage = ref("");
const periodType = ref("month");
const category = ref("1");
const periodKey = ref("");
const pageSize = ref("");

watch(() => route.fullPath, load);
onMounted(load);

async function load() {
  loading.value = true;
  errorMessage.value = "";
  periodType.value = String(route.query.type ?? "month");
  category.value = String(route.query.category ?? "1");
  periodKey.value = String(route.query.key ?? "");
  pageSize.value = String(route.query.ps ?? "");
  try {
    response.value = await getRankPeriod({ ...route.query, type: periodType.value, category: category.value });
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function apply() {
  router.push({
    path: route.path,
    query: compact({ type: periodType.value, category: category.value, key: periodKey.value, ps: pageSize.value, p: 1 }),
  });
}

function clear() {
  router.push({ path: route.path, query: {} });
}

function goPage(page: number) {
  router.push({ path: route.path, query: compact({ ...route.query, p: page }) });
}

function rankWindowText() {
  if (!response.value || response.value.total === 0) {
    return "0 movies";
  }
  const start = (response.value.page - 1) * response.value.page_size + 1;
  const end = Math.min(response.value.page * response.value.page_size, response.value.total);
  return `#${start}-${end} · ${response.value.total} movies`;
}

function compact(query: Record<string, unknown>) {
  return Object.fromEntries(Object.entries(query).filter(([, value]) => value !== "" && value !== undefined && value !== null));
}
</script>
