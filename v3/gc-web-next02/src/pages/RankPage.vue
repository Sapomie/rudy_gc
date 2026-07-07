<template>
  <PageShell class="rank-page-shell cards-page-shell" :class="rankShellClass" :title="title" :note="rankNote" eyebrow="MovieCard Rank">
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <template v-else>
      <div v-if="!message" class="sticky-toolbar rank-sticky-toolbar">
        <div class="rank-sort-bar">
          <div v-if="isDayRank" class="rank-day-header">
            <div class="alert alert-success py-2 mb-0">
              <div class="rank-alert-stack">
                <div class="rank-alert-row">
                  <strong class="font-size-14">Rank Date</strong>
                  <span class="font-18 fw-semibold">{{ rankDate }}</span>
                </div>
                <div class="rank-alert-row">
                  <strong class="font-size-14">Total</strong>
                  <span class="font-18 fw-semibold">{{ total }}</span>
                </div>
              </div>
            </div>
            <div class="rank-day-actions">
              <div class="btn-group btn-group-sm" role="group" aria-label="Rank day navigation">
                <a :href="dayPrevHref" class="btn btn-outline-primary" :class="{ disabled: prevDisabled }" :aria-disabled="prevDisabled ? 'true' : 'false'">前一天</a>
                <a :href="dayNextHref" class="btn btn-outline-primary" :class="{ disabled: nextDisabled }" :aria-disabled="nextDisabled ? 'true' : 'false'">后一天</a>
                <a :href="dayRandHref" class="btn btn-outline-primary">随机</a>
              </div>
              <form class="rank-date-form" method="get" action="/moviecarddayrank">
                <input
                  class="form-control form-control-sm"
                  type="text"
                  name="date"
                  :value="rankDate"
                  placeholder="YYYY-MM-DD"
                  maxlength="10"
                  pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                  inputmode="numeric"
                >
                <button class="btn btn-sm btn-outline-primary" type="submit">Go</button>
              </form>
            </div>
          </div>

          <div v-else class="rank-period-header">
            <div class="alert alert-success py-2 mb-0">
              <div class="rank-period-alert">
                <div class="rank-period-meta">
                  <div class="rank-period-title-row">
                    <strong class="font-size-14">{{ periodTypeLabel }}</strong>
                    <span class="font-18 fw-semibold">{{ periodKey }}</span>
                    <span class="badge bg-light text-dark border">{{ categoryLabel }}</span>
                  </div>
                  <div class="small text-muted">
                    {{ rangeStart }} ~ {{ rangeEnd }} | Total {{ total }}
                  </div>
                </div>
                <div class="rank-period-switches">
                  <div class="btn-group btn-group-sm" role="group" aria-label="Period type switch">
                    <a
                      v-for="item in typeLinks"
                      :key="item.label"
                      :href="item.href || '#'"
                      class="btn"
                      :class="item.active ? 'btn-primary active disabled' : 'btn-outline-primary'"
                    >
                      {{ item.label }}
                    </a>
                  </div>
                  <div class="btn-group btn-group-sm" role="group" aria-label="Period category switch">
                    <a
                      v-for="item in categoryLinks"
                      :key="item.label"
                      :href="item.href || '#'"
                      class="btn"
                      :class="item.active ? 'btn-secondary active disabled' : 'btn-outline-secondary'"
                    >
                      {{ item.label }}
                    </a>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="toolbar-line rank-toolbar-line">
          <div class="owned-line">
            <div v-if="isPeriodRank" class="btn-group btn-group-sm" role="group" aria-label="Period navigation">
              <a :href="prevHref || '#'" class="btn btn-outline-primary" :class="{ disabled: prevDisabled }" :aria-disabled="prevDisabled ? 'true' : 'false'">上一期</a>
              <a :href="nextHref || '#'" class="btn btn-outline-primary" :class="{ disabled: nextDisabled }" :aria-disabled="nextDisabled ? 'true' : 'false'">下一期</a>
              <a :href="latestHref || '/moviecardperiodrank'" class="btn btn-outline-primary">最新</a>
            </div>
          </div>
          <PaginationBar :page-info="pageInfo" />
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
const dayPrevHref = computed(() => `/moviecarddayrank?date=${encodeURIComponent(text(payload.value, "prevDate"))}`);
const dayNextHref = computed(() => `/moviecarddayrank?date=${encodeURIComponent(text(payload.value, "nextDate"))}`);
const dayRandHref = computed(() => `/moviecarddayrank?date=${encodeURIComponent(text(payload.value, "randDate"))}`);
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
