<template>
  <PageFrame :title="response?.title || String(route.meta.title || 'Page')" eyebrow="Data Page" :description="response?.description || ''">
    <template #actions>
      <RouterLink v-for="link in response?.links || []" :key="link.href" class="rgc-button" :to="link.href">{{ link.label }}</RouterLink>
    </template>

    <form class="rgc-panel rgc-generic-form" @submit.prevent="apply">
      <label v-for="filter in activeFilters" :key="filter.key">
        <span>{{ filter.label }}</span>
        <select v-if="filter.type === 'select'" v-model="filterValues[filter.key]">
          <option v-for="option in filter.options || []" :key="option.value" :value="option.value">{{ option.label }}</option>
        </select>
        <input v-else v-model="filterValues[filter.key]" :placeholder="filter.placeholder" />
      </label>
      <label>
        <span>Page Size</span>
        <input v-model="pageSize" type="number" />
      </label>
      <button class="rgc-button primary" type="submit">查询</button>
      <button class="rgc-button" type="button" @click="clear">清空</button>
    </form>

    <StateBlock v-if="loading" title="加载中" message="正在读取页面数据。" />
    <StateBlock v-else-if="errorMessage" title="加载失败" :message="errorMessage" tone="error" />
    <template v-else-if="response">
      <div class="rgc-panel rgc-meta-strip">
        <span>{{ response.kind }}</span>
        <span>{{ response.total }} rows</span>
        <span v-for="stat in response.stats" :key="stat.label">{{ stat.label }} {{ stat.value }}</span>
      </div>
      <template v-if="response.kind === 'operation'">
        <StateBlock title="操作页入口已加载" message="当前 v2 后端尚未暴露真实动作接口，前端不渲染假按钮。" tone="empty" />
      </template>
      <template v-else>
        <DataTable
          v-if="response.rows.length > 0"
          :columns="response.columns"
          :rows="response.rows"
          :order-by="response.order_by"
          :order="response.order"
          @sort="sortBy"
        />
        <StateBlock v-else title="无数据" message="当前筛选条件没有行。" tone="empty" />
        <div class="rgc-panel rgc-list-foot">
          <PaginationControl :page="response.page" :page-size="response.page_size" :total="response.total" @go="goPage" />
        </div>
      </template>
    </template>
  </PageFrame>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { getGenericPage } from "@/api/gc";
import DataTable from "@/components/DataTable.vue";
import PageFrame from "@/components/PageFrame.vue";
import PaginationControl from "@/components/PaginationControl.vue";
import StateBlock from "@/components/StateBlock.vue";
import type { PageFilter, PageListResponse } from "@/types/api";

const route = useRoute();
const router = useRouter();
const response = ref<PageListResponse | null>(null);
const loading = ref(false);
const errorMessage = ref("");
const filterValues = reactive<Record<string, string>>({});
const pageSize = ref("");

const pageKey = computed(() => String(route.meta.pageKey || ""));
const activeFilters = computed<PageFilter[]>(() => response.value?.filters || [{ key: "q", label: "关键词", type: "text" }]);

watch(() => route.fullPath, load);
onMounted(load);

async function load() {
  loading.value = true;
  errorMessage.value = "";
  pageSize.value = String(route.query.ps ?? "");
  try {
    const params = buildRouteParams();
    const data = await getGenericPage(pageKey.value, params);
    response.value = data;
    syncFilters(data.filters);
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function buildRouteParams() {
  const params: Record<string, unknown> = { ...route.query };
  if (route.params.id) {
    params.id = route.params.id;
  }
  if (route.params.name) {
    params.q = route.params.name;
  }
  return params;
}

function syncFilters(filters: PageFilter[]) {
  filters.forEach((filter) => {
    filterValues[filter.key] = String(route.query[filter.key] ?? filter.value ?? "");
  });
}

function apply() {
  const query: Record<string, string | number> = { p: 1 };
  Object.entries(filterValues).forEach(([key, value]) => {
    if (value !== "") {
      query[key] = value;
    }
  });
  if (pageSize.value !== "") {
    query.ps = pageSize.value;
  }
  router.push({ path: route.path, query });
}

function clear() {
  router.push({ path: route.path, query: {} });
}

function sortBy(key: string) {
  const nextOrder = response.value?.order_by === key && response.value?.order === "asc" ? "desc" : "asc";
  router.push({ path: route.path, query: { ...route.query, od: key, order: nextOrder, p: 1 } });
}

function goPage(page: number) {
  router.push({ path: route.path, query: { ...route.query, p: page } });
}
</script>
