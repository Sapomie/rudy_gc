<template>
  <PageShell class="records-page-shell sc-events-page-shell" title="Records" note="" eyebrow="">
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <div v-else id="page-root" class="records-page-root next02-sc-pink-theme next02-sc-events-rewrite">
      <div id="crawlRecordListPage" class="crawl-record-list-page next02-records-rewrite">
        <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell gc-page-head">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">CRAWL / RECORD POOL</span>
            <h1>{{ pageTitle || "CrawlRecords" }}</h1>
            <p>{{ pageNote || "抓取记录、耗时和类型筛选追踪。" }}</p>
          </div>
        </section>

        <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
          <div class="card-body">
            <div
              class="rank-stats-toolbar-scope rank-page-shell cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell standard-filter-toolbar crawl-record-filter-toolbar"
            >
              <form
                class="sticky-toolbar cards-toolbar rank-sticky-toolbar next02-pink-cards-toolbar rank-stats-toolbar standard-filter-toolbar-form crawl-record-filter-form photo-results-toolbar"
                method="get"
                action="/crawlrecordlist"
              >
                <input type="hidden" name="sort" :value="queryValue('Sort')" />
                <input type="hidden" name="order" :value="queryValue('Order')" />
                <input id="crawlRecordTypeInput" type="hidden" name="type" :value="queryValue('Type')" />

                <div class="cards-toolbar-sort rank-toolbar-sort standard-filter-toolbar-fields sort-controls">
                  <div class="rank-toolbar-group standard-filter-group crawl-record-filter-types" aria-label="记录类型">
                    <span class="soft-badge rank-meta-badge">Type</span>
                    <LinkStrip :items="typeLinks" />
                  </div>

                  <div class="rank-toolbar-group standard-filter-group crawl-record-filter-options">
                    <label
                      class="rank-stats-filter-field standard-filter-field crawl-record-page-size-field"
                      for="crawlRecordPageSize"
                    >
                      <span>每页</span>
                      <input
                        id="crawlRecordPageSize"
                        class="rank-date-input standard-filter-input standard-filter-page-size-input"
                        type="number"
                        step="1"
                        name="ps"
                        :value="queryValue('PageSize') || '20'"
                      />
                    </label>
                  </div>
                </div>

                <div class="rank-navigation-group standard-filter-actions crawl-record-filter-actions">
                  <button class="button-secondary rank-toolbar-action" type="submit">筛选</button>
                  <a class="button-secondary rank-toolbar-action" :href="clearHref">清空</a>
                </div>

                <div
                  v-if="hasRows"
                  class="toolbar-line cards-toolbar-bottom rank-toolbar-line standard-filter-status-line"
                >
                  <div class="owned-line cards-toolbar-filters rank-toolbar-filters"></div>
                  <div class="next02-pink-cards-toolbar-status">
                    <div class="next02-pink-cards-metrics" aria-label="抓取记录摘要">
                      <span :title="`总数 ${total}`" :aria-label="`总数 ${total}`"
                        ><strong>{{ total }}</strong></span
                      >
                      <span :title="`当前页 ${rows.length}`" :aria-label="`当前页 ${rows.length}`"
                        ><strong>{{ rows.length }}</strong></span
                      >
                    </div>
                    <div v-if="pageLinks.length" class="cards-toolbar-pagination next02-pink-cards-pagination">
                      <PaginationBar :links="pageLinks" />
                    </div>
                  </div>
                </div>
              </form>
            </div>

            <div class="table-responsive table-scroll">
              <table class="table table-hover align-middle mb-0 sc-events-table crawl-record-table data-table">
                <colgroup>
                  <col class="crawl-record-col-started" />
                  <col class="crawl-record-col-type" />
                  <col class="crawl-record-col-detail" />
                  <col class="crawl-record-col-duration" />
                  <col class="crawl-record-col-name" />
                </colgroup>
                <thead>
                  <tr>
                    <th class="sc-sortable-head">
                      <SortLink
                        class="crawl-record-sort-link sc-sort-link"
                        :link="headerLink('StartedTime')"
                        label="StartedTime"
                        indicator-class="crawl-record-sort-indicator"
                      />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink
                        class="crawl-record-sort-link sc-sort-link"
                        :link="headerLink('Type')"
                        label="Type"
                        indicator-class="crawl-record-sort-indicator"
                      />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink
                        class="crawl-record-sort-link sc-sort-link"
                        :link="headerLink('DetailNumber')"
                        label="DetailNumber"
                        indicator-class="crawl-record-sort-indicator"
                      />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink
                        class="crawl-record-sort-link sc-sort-link"
                        :link="headerLink('DurationMin')"
                        label="Duration(min)"
                        indicator-class="crawl-record-sort-indicator"
                      />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink
                        class="crawl-record-sort-link sc-sort-link"
                        :link="headerLink('Name')"
                        label="Name"
                        indicator-class="crawl-record-sort-indicator"
                      />
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td colspan="5">
                      <div class="sc-empty-state crawl-record-empty-state">当前筛选下没有 `e_record` 数据。</div>
                    </td>
                  </tr>
                  <tr
                    v-for="(row, index) in rows"
                    v-else
                    :key="rowKey(row, index)"
                    class="next02-sc-event-row"
                    :class="[recordDayClass(index), { 'crawl-record-day-start': isRecordDayStart(index) }]"
                  >
                    <td :title="formatUnixMinute(value(row, 'StartedTime'))">
                      <span class="sc-event-time">
                        <span class="sc-event-date">{{ recordTimeParts(value(row, "StartedTime")).date }}</span>
                        <small v-if="recordTimeParts(value(row, 'StartedTime')).time" class="sc-event-hour">{{
                          recordTimeParts(value(row, "StartedTime")).time
                        }}</small>
                      </span>
                    </td>
                    <td class="sc-content-cell">
                      <span class="crawl-record-type" :title="value(row, 'Type')">{{ value(row, "Type") }}</span>
                    </td>
                    <td class="sc-number sc-primary-value">{{ value(row, "DetailNumber") }}</td>
                    <td class="sc-number">{{ value(row, "DurationMinText") }}</td>
                    <td class="sc-content-cell">
                      <span class="crawl-record-name" :title="value(row, 'Name')">{{ value(row, "Name") }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import "@/features/crawler/styles/index.css";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/features/crawler/api";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LinkStrip from "@/shared/components/LinkStrip.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import type { PageInfo } from "@/shared/types/gc";
import type { CrawlerRecord } from "@/features/crawler/types";
import { asArray, asRecord, normalizeHref, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<CrawlerRecord>({});
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

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title"));
const pageNote = computed(() => text(payload.value, "PageNote"));
const total = computed(() => text(payload.value, "Total", "total") || "0");
const query = computed(() => asRecord(pick(payload.value, "Query")));
const rows = computed(() => asArray<CrawlerRecord>(pick(payload.value, "Rows", "rows")));
const hasRows = computed(() => rows.value.length > 0);
const typeLinks = computed(() =>
  asArray<CrawlerRecord>(pick(payload.value, "TypeLinks")).map((item) => ({
    label: text(item, "Label", "label"),
    href: normalizeHref(text(item, "Href", "href") || "#"),
    active: Boolean(item.Active || item.active)
  }))
);
const headerLinks = computed(() => asRecord(pick(payload.value, "HeaderLinks")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<CrawlerRecord>(pageInfo.value.Links));
const clearHref = computed(() => normalizeHref(text(payload.value, "ClearHref") || "/crawlrecordlist"));

function queryValue(key: string): string {
  return text(query.value, key);
}

function headerLink(key: string): CrawlerRecord {
  return asRecord(headerLinks.value[key]);
}

function value(row: CrawlerRecord, key: string): string {
  return text(row, key);
}

function rowKey(row: CrawlerRecord, index: number): string {
  return [value(row, "StartedTime"), value(row, "Type"), value(row, "Name"), index].join(":");
}

function formatUnixMinute(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")} ${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}

function recordTimeParts(input: unknown): { date: string; time: string } {
  const formatted = formatUnixMinute(input);
  if (formatted === "-") {
    return { date: "-", time: "" };
  }
  const [date, time = ""] = formatted.split(" ", 2);
  return { date, time };
}

function recordDayKey(row: CrawlerRecord): string {
  return recordTimeParts(value(row, "StartedTime")).date;
}

function recordDayClass(index: number): string {
  const day = recordDayKey(rows.value[index]);
  if (day === "-") {
    return "crawl-record-day-neutral";
  }
  let groupIndex = 0;
  for (let rowIndex = 1; rowIndex <= index; rowIndex += 1) {
    const currentDay = recordDayKey(rows.value[rowIndex]);
    const previousDay = recordDayKey(rows.value[rowIndex - 1]);
    if (currentDay !== "-" && previousDay !== "-" && currentDay !== previousDay) {
      groupIndex += 1;
    }
  }
  return `crawl-record-day-group-${(groupIndex % 2) + 1}`;
}

function isRecordDayStart(index: number): boolean {
  if (index <= 0) {
    return false;
  }
  const currentDay = recordDayKey(rows.value[index]);
  const previousDay = recordDayKey(rows.value[index - 1]);
  return currentDay !== "-" && previousDay !== "-" && currentDay !== previousDay;
}
</script>
