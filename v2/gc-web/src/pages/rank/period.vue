<template>
  <PageFrame title="Period Rank" eyebrow="Ranking" description="按周、月、季、年查看周期榜。">
    <template #actions>
      <RouterLink v-if="response" class="rgc-button" :to="response.latest_href || '/moviecardperiodrank'">最新</RouterLink>
    </template>

    <form class="rgc-panel rgc-rank-tools" @submit.prevent="apply">
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
        <input v-model="category" type="number" name="category" />
      </label>
      <label>
        <span>Key</span>
        <input v-model.trim="periodKey" type="text" name="key" placeholder="留空取最新" />
      </label>
      <label>
        <span>Page Size</span>
        <input v-model="pageSize" type="number" name="ps" />
      </label>
      <button class="rgc-button primary" type="submit">筛选</button>
    </form>

    <StateBlock v-if="loading" title="加载中" message="正在读取周期榜。" />
    <StateBlock v-else-if="errorMessage" title="加载失败" :message="errorMessage" tone="error" />
    <template v-else-if="response">
      <div class="rgc-panel rgc-period-summary">
        <div>
          <strong>{{ response.period_type_label }} {{ response.period_key }}</strong>
          <span>{{ response.category_label }} · {{ response.range_start }} ~ {{ response.range_end }} · {{ response.total }} 条</span>
        </div>
        <div class="rgc-inline-nav">
          <RouterLink v-for="item in response.type_links" :key="item.href" :class="{ active: item.active }" :to="item.href">{{ item.label }}</RouterLink>
          <RouterLink v-for="item in response.category_links" :key="item.href" :class="{ active: item.active }" :to="item.href">{{ item.label }}</RouterLink>
        </div>
      </div>

      <div class="rgc-panel rgc-list-head">
        <PaginationControl :page="response.page" :page-size="response.page_size" :total="response.total" @go="goPage" />
      </div>

      <div v-if="response.items.length > 0" class="rgc-period-grid">
        <article v-for="item in response.items" :key="`${item.rank_pos}-${item.movie.movie_jav_id}`" class="rgc-period-card">
          <MovieTile :movie="item.movie" />
          <div class="rgc-period-meta">
            <b>#{{ item.rank_pos }}</b>
            <span :class="`tone-${item.rank_change_class}`">{{ item.rank_change_text }}</span>
            <small>上期 {{ item.prev_rank_text }}</small>
            <small>峰值 #{{ item.peak_rank }}</small>
            <small>最佳日 #{{ item.best_rank }}</small>
            <small>上榜 {{ item.days_in_rank }} 天 · 取样 {{ item.used_pick_days }} 天</small>
            <small>Score {{ item.score.toFixed(2) }}</small>
          </div>
        </article>
      </div>
      <StateBlock v-else title="无结果" message="当前周期没有排名数据。" tone="empty" />
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

function goPage(page: number) {
  router.push({ path: route.path, query: { ...route.query, p: page } });
}

function compact(query: Record<string, string | number>) {
  return Object.fromEntries(Object.entries(query).filter(([, value]) => value !== ""));
}
</script>
