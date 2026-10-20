<template>
  <PageShell
    :class="['cards-page-shell', 'next02-pink-cards-page', 'next02-pink-cards-page-shell', pageClass]"
    :title="title"
    :note="pageNote"
    eyebrow="FC2 MovieCard"
  >
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <template v-else>
      <div class="sticky-toolbar cards-toolbar next02-pink-cards-toolbar" aria-label="FC2 Cards 控制栏">
        <div class="cards-toolbar-sort">
          <MovieCardSortStrip :query="sortQuery" :options="sortOptions" />
        </div>
        <div class="toolbar-line cards-toolbar-bottom">
          <div class="owned-line cards-toolbar-filters">
            <LinkStrip :items="cardLinks" />
          </div>
          <div class="next02-pink-cards-toolbar-status">
            <CardGridSizeControl />
            <div class="next02-pink-cards-metrics" aria-label="FC2 Cards 页面摘要">
              <span :title="`总数 ${total}`" :aria-label="`总数 ${total}`"
                ><strong>{{ total }}</strong></span
              >
              <span :title="`当前页 ${cards.length}`" :aria-label="`当前页 ${cards.length}`"
                ><strong>{{ cards.length }}</strong></span
              >
            </div>
            <div class="cards-toolbar-pagination next02-pink-cards-pagination">
              <PaginationBar :page-info="pageInfo" />
            </div>
          </div>
        </div>
      </div>

      <FC2MovieGrid :cards="cards" />
    </template>
  </PageShell>
</template>

<script setup lang="ts">
import "@/features/fc2/styles/index.css";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/features/fc2/api";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import CardGridSizeControl from "@/shared/components/CardGridSizeControl.vue";
import FC2MovieGrid from "@/features/fc2/components/FC2MovieGrid.vue";
import LinkStrip from "@/shared/components/LinkStrip.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import MovieCardSortStrip from "@/features/movies/components/MovieCardSortStrip.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import type { PageInfo } from "@/shared/types/gc";
import type { FC2Record } from "@/features/fc2/types";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<FC2Record>({});
const loading = ref(false);
const error = ref("");

const sortOptions = [
  { key: "ByPostTime", label: "发布时间" },
  { key: "ByFC2Key", label: "影片标识" },
  { key: "ByLastSelectedScTime", label: "最近入选 SC" },
  { key: "ByBirthTime", label: "下载时间" },
  { key: "ByScTimes", label: "SC 次数" }
];

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

const pageTitles: Record<string, string> = {
  "/fclistcard": "FC2 Cards",
  "/fclistcardowned": "FC2 Cards MediaOwned",
  "/fclistcardrandom": "FC2 Cards Random"
};
const title = computed(() => pageTitles[route.path] || text(payload.value, "Title") || "FC2 Cards");
const pageClass = computed(() => `cards-route-${route.path.replace(/^\/+/, "")}`);
const cards = computed(() => asArray<FC2Record>(pick(payload.value, "movies", "Movies")));
const total = computed(() => text(payload.value, "total", "Total") || "0");
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const sortQuery = computed(() => asRecord(pick(payload.value, "SortQuery", "sortQuery")));
const pageNote = computed(() =>
  route.path === "/fclistcardrandom"
    ? "随机展示 FC2 影片，卡片使用默认联系表封面。"
    : "FC2 影片卡片使用默认联系表封面。"
);
const cardLinks = computed(() => [
  { label: "All", href: "/fclistcard", active: route.path === "/fclistcard" },
  { label: "MediaOwned", href: "/fclistcardowned", active: route.path === "/fclistcardowned" },
  { label: "Random", href: "/fclistcardrandom", active: route.path === "/fclistcardrandom" }
]);
</script>
