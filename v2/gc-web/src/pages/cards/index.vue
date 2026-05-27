<template>
  <section class="rgc-cards-page" aria-label="Movie Cards">
    <h1 class="rgc-visually-hidden">{{ response?.title || currentRouteTitle }}</h1>

    <div class="rgc-cards-toolbar" aria-label="Cards tools">
      <div class="rgc-toolbar-row">
        <nav class="rgc-view-tabs" aria-label="Cards pages">
          <RouterLink
            v-for="item in viewLinks"
            :key="item.key"
            :to="item.to"
            :class="{ active: item.active }"
            :aria-current="item.active ? 'page' : undefined"
          >
            {{ item.title }}
          </RouterLink>
        </nav>
      </div>

      <div class="rgc-toolbar-row">
        <nav class="rgc-sort-strip" aria-label="排序">
          <RouterLink
            v-for="item in sortLinks"
            :key="item.value"
            :to="item.to"
            :class="{ active: item.active }"
            :aria-current="item.active ? 'page' : undefined"
          >
            {{ item.label }}
          </RouterLink>
        </nav>
      </div>

      <div class="rgc-toolbar-row split">
        <div class="rgc-owned-bar" aria-label="WMedia 快捷筛选">
          <span class="rgc-tool-label">WMedia</span>
          <RouterLink
            v-for="item in ownedLinks"
            :key="item.value"
            :to="item.to"
            :class="['rgc-owned-link', item.tone, { active: item.active }]"
            :aria-current="item.active ? 'page' : undefined"
          >
            {{ item.label }}
          </RouterLink>
          <span class="rgc-total-pill">Total: {{ totalText }}</span>
          <button class="rgc-button compact" type="button" @click="filterOpen = !filterOpen">
            {{ filterOpen ? "收起" : "筛选" }}
          </button>
          <button v-if="activeFilterCount > 0" class="rgc-button compact subtle" type="button" @click="clearFilters">
            清空 {{ activeFilterCount }}
          </button>
        </div>

        <PaginationControl
          v-if="response"
          :page="response.page"
          :page-size="response.page_size"
          :total="response.total"
          @go="goPage"
        />
      </div>
    </div>

    <form v-show="filterOpen" class="rgc-filter-panel" @submit.prevent="applyFilters">
      <div class="rgc-filter-grid">
        <label v-for="field in filterFields" :key="field.key" :class="{ wide: field.wide }">
          <span>{{ field.label }}</span>
          <select v-if="field.options" v-model="filters[field.key]" :name="field.key">
            <option v-for="option in field.options" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
          <input
            v-else
            v-model="filters[field.key]"
            :name="field.key"
            :type="field.type"
            :step="field.step"
            :placeholder="field.placeholder"
            :inputmode="field.dateLike ? 'numeric' : undefined"
            :maxlength="field.dateLike ? 10 : undefined"
            :pattern="field.dateLike ? '[0-9]{4}-[0-9]{2}-[0-9]{2}' : undefined"
          />
        </label>
      </div>

      <div class="rgc-filter-actions">
        <button class="rgc-button primary" type="submit">筛选</button>
        <button class="rgc-button" type="button" @click="clearFilters">清空</button>
        <span v-if="activeFilterLabels.length > 0" class="rgc-filter-summary">
          {{ activeFilterLabels.join(" / ") }}
        </span>
      </div>
    </form>

    <StateBlock v-if="loading" title="加载中" message="正在读取电影卡片。" />
    <StateBlock v-else-if="errorMessage" title="加载失败" :message="errorMessage" tone="error" />
    <template v-else-if="response">
      <MovieGrid v-if="response.items.length > 0" :items="response.items" />
      <StateBlock v-else title="无结果" message="当前筛选条件下没有电影。" tone="empty" />
      <div class="rgc-bottom-pager">
        <PaginationControl :page="response.page" :page-size="response.page_size" :total="response.total" @go="goPage" />
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { listCards } from "@/api/gc";
import MovieGrid from "@/components/MovieGrid.vue";
import PaginationControl from "@/components/PaginationControl.vue";
import StateBlock from "@/components/StateBlock.vue";
import { cardRoutes } from "@/data/routes";
import type { CardsListResponse } from "@/types/api";

interface SelectOption {
  value: string;
  label: string;
}

interface FilterField {
  key: string;
  label: string;
  type: "text" | "number";
  placeholder?: string;
  step?: string;
  dateLike?: boolean;
  wide?: boolean;
  options?: SelectOption[];
}

interface QueryOptions {
  keepPaging?: boolean;
}

const route = useRoute();
const router = useRouter();
const response = ref<CardsListResponse | null>(null);
const loading = ref(false);
const errorMessage = ref("");
const filterOpen = ref(false);

const filterFields: FilterField[] = [
  { key: "cn", label: "演员", type: "text", placeholder: "演员名，支持多值" },
  { key: "gn", label: "类型", type: "text", placeholder: "类型名，支持多值" },
  { key: "dn", label: "导演", type: "text", placeholder: "导演名" },
  { key: "pn", label: "前缀", type: "text", placeholder: "Prefix" },
  { key: "mn", label: "片商", type: "text", placeholder: "Maker" },
  { key: "ln", label: "标签", type: "text", placeholder: "本地 Label" },
  { key: "lj", label: "标签 JavID", type: "text", placeholder: "JavLib 标签 ID" },
  { key: "an", label: "相册名", type: "text", placeholder: "相册名，支持多值" },
  { key: "wd", label: "关键词", type: "text", placeholder: "中文关键词" },
  { key: "pid", label: "Person ID", type: "text", placeholder: "内部筛选 ID" },
  {
    key: "mowned",
    label: "w_media 库存",
    type: "text",
    options: [
      { value: "", label: "全部" },
      { value: "1", label: "全部电影" },
      { value: "2", label: "有 w_media 记录（含移除）" },
      { value: "3", label: "已拥有未移除" },
      { value: "4", label: "有字幕未移除" },
      { value: "5", label: "无字幕未移除" },
      { value: "6", label: "已移除" },
      { value: "7", label: "无 w_media 记录" },
    ],
  },
  {
    key: "nd",
    label: "下载状态",
    type: "text",
    options: [
      { value: "", label: "全部" },
      { value: "1", label: "无需下载" },
      { value: "2", label: "需要下载" },
    ],
  },
  { key: "drkmin", label: "在榜天数", type: "number", placeholder: "最小值" },
  { key: "md1", label: "Media 目录 1", type: "text", placeholder: "目录名" },
  { key: "md2", label: "Media 目录 2", type: "text", placeholder: "目录名" },
  { key: "md3", label: "Media 目录 3", type: "text", placeholder: "目录名" },
  { key: "md4", label: "Media 目录 4", type: "text", placeholder: "目录名" },
  { key: "rs", label: "上映开始", type: "text", dateLike: true },
  { key: "re", label: "上映结束", type: "text", dateLike: true },
  { key: "mbs", label: "Media 开始", type: "text", dateLike: true },
  { key: "mbe", label: "Media 结束", type: "text", dateLike: true },
  { key: "srds", label: "上榜开始", type: "text", dateLike: true },
  { key: "srde", label: "上榜结束", type: "text", dateLike: true },
  { key: "lsctmin", label: "最后 SC 开始", type: "text", dateLike: true },
  { key: "lsctmax", label: "最后 SC 结束", type: "text", dateLike: true },
  { key: "cay", label: "年龄最小", type: "number", step: "0.1", placeholder: "最小值" },
  { key: "cao", label: "年龄最大", type: "number", step: "0.1", placeholder: "最大值" },
  { key: "vwmin", label: "观看最小", type: "number", placeholder: "最小值" },
  { key: "vwmax", label: "观看最大", type: "number", placeholder: "最大值" },
  { key: "smin", label: "评分最小", type: "number", step: "0.1", placeholder: "最小值" },
  { key: "smax", label: "评分最大", type: "number", step: "0.1", placeholder: "最大值" },
  { key: "scmin", label: "SC 最小", type: "number", placeholder: "最小值" },
  { key: "scmax", label: "SC 最大", type: "number", placeholder: "最大值" },
  { key: "comin", label: "Come 最小", type: "number", placeholder: "最小值" },
  { key: "comax", label: "Come 最大", type: "number", placeholder: "最大值" },
  { key: "n", label: "随机数量", type: "number", placeholder: "6" },
  { key: "ps", label: "Page Size", type: "number", placeholder: "18" },
];

const sortOptions = [
  { value: "du", label: "Detail更新" },
  { value: "rd", label: "上映日期" },
  { value: "vw", label: "观看人数" },
  { value: "rk", label: "上榜日期" },
  { value: "drk", label: "在榜天数" },
  { value: "hrk", label: "最高排名" },
  { value: "bt", label: "下载时间" },
  { value: "mbt", label: "Media下载时间" },
  { value: "lsct", label: "最后sc" },
  { value: "sc", label: "sc次数" },
  { value: "co", label: "co次数" },
];

const filters = reactive<Record<string, string>>(
  Object.fromEntries(filterFields.map((field) => [field.key, ""])),
);

const currentView = computed(() => String(route.meta.cardView || "cards"));
const currentPath = computed(() => route.path || "/cards");
const currentRouteTitle = computed(() => cardRoutes.find((item) => item.key === currentView.value)?.title || "MovieCard");
const currentOrderBy = computed(() => response.value?.order_by || queryString("od") || "rd");
const totalText = computed(() => response.value?.total ?? "...");
const activeFilterCount = computed(() => filterFields.filter((field) => queryString(field.key) !== "").length);
const activeFilterLabels = computed(() =>
  filterFields
    .filter((field) => queryString(field.key) !== "")
    .slice(0, 5)
    .map((field) => `${field.label}: ${queryString(field.key)}`),
);

const viewLinks = computed(() =>
  cardRoutes.map((item) => ({
    ...item,
    active: currentView.value === item.key,
    to: { path: item.path, query: routeQueryObject({ keepPaging: false }) },
  })),
);

const sortLinks = computed(() =>
  sortOptions.map((item) => ({
    ...item,
    active: currentOrderBy.value === item.value,
    to: { path: currentPath.value, query: makeQuery({ od: item.value, p: "1" }) },
  })),
);

const ownedLinks = computed(() => {
  const current = queryString("mowned") || (currentView.value === "cardsmediamowned" ? "3" : "1");
  return [
    { value: "1", label: "MAll", tone: "primary" },
    { value: "3", label: "MOwned", tone: "success" },
    { value: "7", label: "MNotOwned", tone: "muted" },
  ].map((item) => ({
    ...item,
    active: current === item.value,
    to: { path: currentPath.value, query: makeQuery({ mowned: item.value, p: "1" }) },
  }));
});

watch(() => route.fullPath, load);
onMounted(load);

async function load() {
  loading.value = true;
  errorMessage.value = "";
  syncFormFromRoute();
  try {
    response.value = await listCards({ ...route.query, view: currentView.value });
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function syncFormFromRoute() {
  filterFields.forEach((field) => {
    filters[field.key] = queryString(field.key);
  });
}

function applyFilters() {
  const query: Record<string, string> = {};
  filterFields.forEach((field) => {
    const value = filters[field.key].trim();
    if (value !== "") {
      query[field.key] = value;
    }
  });
  for (const key of ["od", "order"]) {
    const value = queryString(key);
    if (value !== "") {
      query[key] = value;
    }
  }
  query.p = "1";
  router.push({ path: currentPath.value, query });
}

function clearFilters() {
  router.push({ path: currentPath.value, query: {} });
}

function goPage(page: number) {
  router.push({ path: currentPath.value, query: makeQuery({ p: String(page) }) });
}

function queryString(key: string) {
  const value = route.query[key];
  if (Array.isArray(value)) {
    return value[0] == null ? "" : String(value[0]);
  }
  return value == null ? "" : String(value);
}

function routeQueryObject(options: QueryOptions = {}) {
  const query: Record<string, string> = {};
  Object.entries(route.query).forEach(([key, value]) => {
    if (key === "view" || (options.keepPaging === false && key === "p")) {
      return;
    }
    if (Array.isArray(value)) {
      if (value[0] != null && String(value[0]) !== "") {
        query[key] = String(value[0]);
      }
      return;
    }
    if (value != null && String(value) !== "") {
      query[key] = String(value);
    }
  });
  return query;
}

function makeQuery(updates: Record<string, string>) {
  const query = routeQueryObject();
  Object.entries(updates).forEach(([key, value]) => {
    if (value === "") {
      delete query[key];
      return;
    }
    query[key] = value;
  });
  return query;
}
</script>
