<template>
  <PageFrame title="Day Rank" eyebrow="Ranking" description="按日查看电影排名，支持日期跳转和分页。">
    <template #actions>
      <button class="rgc-button" :disabled="!response || response.prev_disabled" @click="goDate(response?.prev_date)">前一天</button>
      <button class="rgc-button" :disabled="!response" @click="goDate(response?.random_date)">随机</button>
      <button class="rgc-button" :disabled="!response || response.next_disabled" @click="goDate(response?.next_date)">后一天</button>
    </template>

    <form class="rgc-panel rgc-rank-tools" @submit.prevent="apply">
      <label>
        <span>Rank Date</span>
        <input v-model="dateInput" type="date" name="date" />
      </label>
      <label>
        <span>Page Size</span>
        <input v-model="pageSize" type="number" name="ps" />
      </label>
      <button class="rgc-button primary" type="submit">Go</button>
      <strong v-if="response">{{ response.rank_date }} · {{ response.total }} 条</strong>
    </form>

    <StateBlock v-if="loading" title="加载中" message="正在读取日榜。" />
    <StateBlock v-else-if="errorMessage" title="加载失败" :message="errorMessage" tone="error" />
    <template v-else-if="response">
      <div class="rgc-panel rgc-list-head">
        <PaginationControl :page="response.page" :page-size="response.page_size" :total="response.total" @go="goPage" />
      </div>
      <MovieGrid v-if="response.items.length > 0" :items="response.items" />
      <StateBlock v-else title="无结果" message="当前日期没有排名数据。" tone="empty" />
    </template>
  </PageFrame>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { getRankDay } from "@/api/gc";
import MovieGrid from "@/components/MovieGrid.vue";
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
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function apply() {
  router.push({ path: route.path, query: compact({ date: dateInput.value, ps: pageSize.value, p: 1 }) });
}

function goDate(date?: string) {
  if (!date) {
    return;
  }
  router.push({ path: route.path, query: { ...route.query, date, p: 1 } });
}

function goPage(page: number) {
  router.push({ path: route.path, query: { ...route.query, p: page } });
}

function compact(query: Record<string, string | number>) {
  return Object.fromEntries(Object.entries(query).filter(([, value]) => value !== ""));
}
</script>
