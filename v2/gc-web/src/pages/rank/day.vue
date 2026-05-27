<template>
  <PageFrame title="Day Rank" eyebrow="Ranking">
    <form class="rgc-panel rgc-rank-tools rgc-rank-compact" @submit.prevent="apply">
      <div class="rgc-rank-current">
        <strong>{{ response?.rank_date || dateInput || "最新日榜" }}</strong>
        <span>{{ response ? `${response.total} movies` : "读取中" }}</span>
      </div>

      <div class="rgc-rank-fields">
        <label>
          <span>Rank Date</span>
          <input v-model="dateInput" type="date" name="date" />
        </label>
        <label>
          <span>Page Size</span>
          <input v-model="pageSize" type="number" name="ps" />
        </label>
      </div>

      <div class="rgc-rank-actions">
        <button class="rgc-button compact" type="button" :disabled="!response || response.prev_disabled" @click="goDate(response?.prev_date)">前一天</button>
        <button class="rgc-button compact" type="button" :disabled="!response || response.next_disabled" @click="goDate(response?.next_date)">后一天</button>
        <button class="rgc-button compact" type="button" :disabled="!response" @click="goDate(response?.random_date)">随机</button>
        <button class="rgc-button compact primary" type="button" :disabled="!response" @click="goDate(response?.latest_date)">最新</button>
        <button class="rgc-button compact primary" type="submit">Go</button>
        <button class="rgc-button compact" type="button" @click="clear">清空</button>
      </div>
    </form>

    <StateBlock v-if="loading" title="加载中" message="正在读取日榜。" />
    <StateBlock v-else-if="errorMessage" title="加载失败" :message="errorMessage" tone="error" />
    <template v-else-if="response">
      <div class="rgc-list-head rgc-rank-listbar">
        <span>{{ rankWindowText() }}</span>
        <PaginationControl :page="response.page" :page-size="response.page_size" :total="response.total" @go="goPage" />
      </div>

      <div v-if="response.items.length > 0" class="rgc-movie-grid rgc-rank-grid">
        <article v-for="(movie, index) in response.items" :key="movie.movie_jav_id || movie.movie_name" class="rgc-rank-card-wrap">
          <span class="rgc-rank-position">#{{ rankNumber(index) }}</span>
          <MovieTile :movie="movie" />
        </article>
      </div>
      <StateBlock v-else title="无结果" message="当前日期没有排名数据。" tone="empty" />

      <div class="rgc-list-foot rgc-rank-listbar">
        <PaginationControl :page="response.page" :page-size="response.page_size" :total="response.total" @go="goPage" />
      </div>
    </template>
  </PageFrame>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { getRankDay } from "@/api/gc";
import MovieTile from "@/components/MovieTile.vue";
import PageFrame from "@/components/PageFrame.vue";
import PaginationControl from "@/components/PaginationControl.vue";
import StateBlock from "@/components/StateBlock.vue";
import type { RankDayResponse } from "@/types/api";

const route = useRoute();
const router = useRouter();
const response = ref<RankDayResponse | null>(null);
const loading = ref(false);
const errorMessage = ref("");
const dateInput = ref("");
const pageSize = ref("");

watch(() => route.fullPath, load);
onMounted(load);

async function load() {
  loading.value = true;
  errorMessage.value = "";
  dateInput.value = String(route.query.date ?? "");
  pageSize.value = String(route.query.ps ?? "");
  try {
    response.value = await getRankDay({ ...route.query });
    if (!dateInput.value) {
      dateInput.value = response.value.rank_date;
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function apply() {
  router.push({ path: route.path, query: compact({ date: dateInput.value, ps: pageSize.value, p: 1 }) });
}

function clear() {
  router.push({ path: route.path, query: {} });
}

function goDate(date?: string) {
  if (!date) {
    return;
  }
  router.push({ path: route.path, query: compact({ ...route.query, date, p: 1 }) });
}

function goPage(page: number) {
  router.push({ path: route.path, query: compact({ ...route.query, p: page }) });
}

function rankNumber(index: number) {
  if (!response.value) {
    return index + 1;
  }
  return (response.value.page - 1) * response.value.page_size + index + 1;
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
