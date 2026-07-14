<template>
  <PageShell class="rank-page-shell cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell" :class="rankShellClass" :title="title" :note="rankNote" eyebrow="MovieCard Rank">
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <template v-else>
      <div v-if="!message" class="sticky-toolbar cards-toolbar rank-sticky-toolbar next02-pink-cards-toolbar">
        <div class="cards-toolbar-sort rank-toolbar-sort">
          <template v-if="isDayRank">
            <span class="soft-badge rank-meta-badge">Rank Date</span>
            <span class="total-pill rank-date-pill">{{ rankDate || "-" }}</span>
            <form class="rank-date-form" method="get" action="/listcarddayrank">
              <input
                class="rank-date-input smart-pick-date-input"
                type="text"
                name="date"
                :value="rankDate"
                placeholder="YYYY-MM-DD"
                maxlength="10"
                pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                inputmode="numeric"
              >
              <button class="button-secondary rank-toolbar-action" type="submit">Go</button>
            </form>
          </template>

          <template v-else>
            <span class="soft-badge rank-meta-badge">{{ periodTypeLabel }}</span>
            <span class="total-pill rank-period-key">{{ periodKey || "-" }}</span>
            <span class="soft-badge rank-category-badge">{{ categoryLabel || "-" }}</span>
            <LinkStrip :items="typeLinks" />
            <LinkStrip :items="categoryLinks" />
          </template>
        </div>

        <div class="toolbar-line cards-toolbar-bottom rank-toolbar-line">
          <div class="owned-line cards-toolbar-filters rank-toolbar-filters">
            <template v-if="isDayRank">
              <div class="rank-navigation-group">
                <a :href="dayPrevHref" class="button-secondary rank-toolbar-action" :class="{ disabled: prevDisabled }" :aria-disabled="prevDisabled ? 'true' : 'false'">前一天</a>
                <a :href="dayNextHref" class="button-secondary rank-toolbar-action" :class="{ disabled: nextDisabled }" :aria-disabled="nextDisabled ? 'true' : 'false'">后一天</a>
                <a :href="dayRandHref" class="button-secondary rank-toolbar-action">随机</a>
              </div>
            </template>

            <template v-else>
              <span class="soft-badge rank-range-badge">{{ rangeStart || "-" }} ~ {{ rangeEnd || "-" }}</span>
              <div class="rank-navigation-group">
                <a :href="prevHref || '#'" class="button-secondary rank-toolbar-action" :class="{ disabled: prevDisabled }" :aria-disabled="prevDisabled ? 'true' : 'false'">上一期</a>
                <a :href="nextHref || '#'" class="button-secondary rank-toolbar-action" :class="{ disabled: nextDisabled }" :aria-disabled="nextDisabled ? 'true' : 'false'">下一期</a>
                <a :href="latestHref || '/listcardperiodrank'" class="button-secondary rank-toolbar-action">最新</a>
              </div>
            </template>
          </div>
          <div class="next02-pink-cards-toolbar-status">
            <div class="next02-pink-cards-metrics" aria-label="排名页面摘要">
              <span :title="`总数 ${total}`" :aria-label="`总数 ${total}`"><strong>{{ total }}</strong></span>
              <span :title="`当前页 ${cards.length}`" :aria-label="`当前页 ${cards.length}`"><strong>{{ cards.length }}</strong></span>
            </div>
            <div class="cards-toolbar-pagination next02-pink-cards-pagination">
              <PaginationBar :page-info="pageInfo" />
            </div>
          </div>
        </div>
      </div>

      <div v-if="message" class="empty-block">{{ message }}</div>
      <MovieGrid v-else :cards="cards" :mode="viewType" />
    </template>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LinkStrip from "@/components/LinkStrip.vue";
import LoadingState from "@/components/LoadingState.vue";
import MovieGrid from "@/components/MovieGrid.vue";
import PageShell from "@/components/PageShell.vue";
import PaginationBar from "@/components/PaginationBar.vue";
import type { AnyRecord, PageInfo } from "@/types/gc";
import { asArray, asRecord, pick, text } from "@/utils/format";

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

const title = computed(() => text(payload.value, "Title") || String(route.meta.title || "Rank"));
const viewType = computed(() => text(payload.value, "ViewType"));
const isDayRank = computed(() => viewType.value === "rank-day");
const isPeriodRank = computed(() => viewType.value === "rank-period");
const message = computed(() => text(payload.value, "message", "Message"));
const cards = computed(() => asArray<AnyRecord>(pick(payload.value, "cards", "movies", "Cards", "Movies")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const total = computed(() => text(payload.value, "total", "Total") || "0");
const rankDate = computed(() => text(payload.value, "rankDate"));
const periodTypeLabel = computed(() => text(payload.value, "periodTypeLabel"));
const categoryLabel = computed(() => text(payload.value, "categoryLabel"));
const periodKey = computed(() => text(payload.value, "periodKey"));
const rangeStart = computed(() => text(payload.value, "rangeStart"));
const rangeEnd = computed(() => text(payload.value, "rangeEnd"));
const prevDisabled = computed(() => Boolean(payload.value.prevDisabled));
const nextDisabled = computed(() => Boolean(payload.value.nextDisabled));
const prevHref = computed(() => text(payload.value, "prevHref"));
const nextHref = computed(() => text(payload.value, "nextHref"));
const latestHref = computed(() => text(payload.value, "latestHref"));
const dayPrevHref = computed(() => `/listcarddayrank?date=${encodeURIComponent(text(payload.value, "prevDate"))}`);
const dayNextHref = computed(() => `/listcarddayrank?date=${encodeURIComponent(text(payload.value, "nextDate"))}`);
const dayRandHref = computed(() => `/listcarddayrank?date=${encodeURIComponent(text(payload.value, "randDate"))}`);
const rankNote = computed(() => {
  if (isDayRank.value) {
    return `日期 ${rankDate.value || "-"}`;
  }
  return [periodTypeLabel.value, categoryLabel.value, periodKey.value].filter(Boolean).join(" · ");
});
const rankShellClass = computed(() => isDayRank.value ? "cards-route-dayrank" : "cards-route-periodrank");
const typeLinks = computed(() => flattenLinks(pick(payload.value, "typeLinks", "TypeLinks")));
const categoryLinks = computed(() => flattenLinks(pick(payload.value, "categoryLinks", "CategoryLinks")));

function flattenLinks(value: unknown) {
  if (Array.isArray(value)) {
    return value.map(linkItem);
  }
  return Object.values(asRecord(value)).map(linkItem);
}

function linkItem(value: unknown) {
  const row = asRecord(value);
  return {
    label: text(row, "Label", "label", "Name", "name") || text(row, "Title", "title"),
    href: text(row, "Href", "href"),
    active: Boolean(row.Active || row.active),
    disabled: Boolean(row.Disabled || row.disabled)
  };
}
</script>
