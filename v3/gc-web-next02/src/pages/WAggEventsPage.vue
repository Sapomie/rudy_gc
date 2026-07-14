<template>
  <PageShell class="w-agg-events-page-shell sc-events-page-shell" title="WAggEvents" note="" eyebrow="">
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <div v-else id="page-root" class="w-agg-events-page-root next02-sc-pink-theme next02-sc-events-rewrite">
      <div id="wAggEventListPage" class="w-agg-event-list-page next02-w-agg-events-rewrite">
        <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">W AGG / EVENT POOL</span>
            <h4>{{ pageTitle || "WAggEvent 列表" }}</h4>
            <p>{{ pageNote || "直接查看 w_agg_event 落库记录。" }}</p>
          </div>
        </section>

        <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
          <div class="card-body">
            <div class="next02-sc-panel-head">
              <form class="sc-event-filter-form w-agg-event-filter-form" method="get" action="/aggeventlist">
                <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')">
                <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')">

                <div class="w-agg-event-filter-cluster w-agg-event-filter-query" aria-label="检索条件">
                  <label class="sc-event-filter-field w-agg-event-filter-field is-agg-key">
                    <span>AggKey</span>
                    <input class="form-control" type="text" name="agg_key" :value="queryValue('AggKey')" placeholder="精确匹配">
                  </label>
                  <label class="sc-event-filter-field w-agg-event-filter-field is-flow-key">
                    <span>FlowKey</span>
                    <input class="form-control" type="text" name="flow_key" :value="queryValue('FlowKey')" placeholder="精确匹配">
                  </label>
                  <label class="sc-event-filter-field w-agg-event-filter-field is-status">
                    <span>Status</span>
                    <input class="form-control" type="text" name="status" :value="queryValue('Status')" placeholder="精确匹配">
                  </label>
                </div>

                <div class="w-agg-event-filter-cluster w-agg-event-filter-options">
                  <label class="sc-event-filter-field w-agg-event-page-size-field" for="wAggEventPageSize">
                    <span>每页</span>
                    <input id="wAggEventPageSize" class="form-control" type="number" step="1" name="ps" :value="queryValue('PageSize') || '20'">
                  </label>

                  <div class="w-agg-event-filter-actions">
                    <button class="sc-event-filter-submit" type="submit">筛选</button>
                    <a class="sc-event-filter-clear" :href="clearHref">清空</a>
                  </div>
                </div>
              </form>

              <div v-if="hasRows" class="next02-sc-hero-metrics sc-event-list-metrics" aria-label="W 聚合事件摘要">
                <span><strong>{{ total }}</strong> Records</span>
                <span><strong>{{ rows.length }}</strong> Current Page</span>
              </div>

              <div v-if="hasRows && pageLinks.length" class="next02-sc-movie-pager">
                <WAggEventPagination :links="pageLinks" />
              </div>
            </div>

            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 sc-events-table w-agg-event-table">
                <colgroup>
                  <col class="w-agg-event-col-started">
                  <col class="w-agg-event-col-agg-key">
                  <col class="w-agg-event-col-flow-key">
                  <col class="w-agg-event-col-scope">
                  <col class="w-agg-event-col-bucket">
                  <col class="w-agg-event-col-top">
                  <col class="w-agg-event-col-finished">
                  <col class="w-agg-event-col-duration">
                </colgroup>
                <thead>
                  <tr>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('StartedTime')" label="StartedTime" /></th>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('AggKey')" label="AggKey" /></th>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('FlowKey')" label="FlowKey" /></th>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('ScopeCount')" label="ScopeCount" /></th>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('BucketCount')" label="BucketCount" /></th>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('TopCount')" label="TopCount" /></th>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('FinishedTime')" label="FinishedTime" /></th>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('Duration')" label="Duration(秒)" /></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td colspan="8"><div class="sc-empty-state w-agg-event-empty-state">当前筛选下没有 `w_agg_event` 数据。</div></td>
                  </tr>
                  <tr
                    v-for="(row, index) in rows"
                    v-else
                    :key="rowKey(row, index)"
                    class="next02-sc-event-row"
                    :class="[eventDayClass(index), { 'w-agg-event-day-start': isEventDayStart(index) }]"
                    :data-item-id="value(row, 'ID')"
                  >
                    <td :title="formatUnixMinute(value(row, 'StartedTime'))">
                      <span class="sc-event-time">
                        <span class="sc-event-date">{{ eventTimeParts(value(row, "StartedTime")).date }}</span>
                        <small v-if="eventTimeParts(value(row, 'StartedTime')).time" class="sc-event-hour">{{ eventTimeParts(value(row, "StartedTime")).time }}</small>
                      </span>
                    </td>
                    <td class="sc-content-cell"><span class="w-agg-event-key is-agg-key" :title="value(row, 'AggKey')">{{ value(row, "AggKey") }}</span></td>
                    <td class="sc-content-cell"><span class="w-agg-event-key is-flow-key" :title="value(row, 'FlowKey')">{{ value(row, "FlowKey") }}</span></td>
                    <td class="sc-number">{{ value(row, "ScopeCount") }}</td>
                    <td class="sc-number">{{ value(row, "BucketCount") }}</td>
                    <td class="sc-number sc-primary-value">{{ value(row, "TopCount") }}</td>
                    <td :title="formatUnixMinute(value(row, 'FinishedTime'))">
                      <span class="sc-event-time">
                        <span class="sc-event-date">{{ eventTimeParts(value(row, "FinishedTime")).date }}</span>
                        <small v-if="eventTimeParts(value(row, 'FinishedTime')).time" class="sc-event-hour">{{ eventTimeParts(value(row, "FinishedTime")).time }}</small>
                      </span>
                    </td>
                    <td class="sc-number">{{ value(row, "DurationSecondsText") }}</td>
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
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LoadingState from "@/components/LoadingState.vue";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord, PageInfo } from "@/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/utils/format";

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

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title"));
const pageNote = computed(() => text(payload.value, "PageNote"));
const total = computed(() => text(payload.value, "Total", "total") || "0");
const query = computed(() => asRecord(pick(payload.value, "Query")));
const rows = computed(() => asArray<AnyRecord>(pick(payload.value, "Rows", "rows")));
const hasRows = computed(() => rows.value.length > 0);
const headerLinks = computed(() => asRecord(pick(payload.value, "HeaderLinks")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<AnyRecord>(pageInfo.value.Links));
const clearHref = computed(() => normalizeHref(text(payload.value, "ClearHref") || "/aggeventlist"));

function queryValue(key: string): string {
  return text(query.value, key);
}

function headerLink(key: string): AnyRecord {
  return asRecord(headerLinks.value[key]);
}

function value(row: AnyRecord, key: string): string {
  return text(row, key);
}

function rowKey(row: AnyRecord, index: number): string {
  return [
    value(row, "StartedTime"),
    value(row, "AggKey"),
    value(row, "FlowKey"),
    value(row, "ScopeCount"),
    value(row, "BucketCount"),
    index
  ].join(":");
}

function formatUnixMinute(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")} ${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}

function eventTimeParts(input: unknown): { date: string; time: string } {
  const formatted = formatUnixMinute(input);
  if (formatted === "-") {
    return { date: "-", time: "" };
  }
  const [date, time = ""] = formatted.split(" ", 2);
  return { date, time };
}

function eventDayKey(row: AnyRecord): string {
  return eventTimeParts(value(row, "StartedTime")).date;
}

function eventDayClass(index: number): string {
  const day = eventDayKey(rows.value[index]);
  if (day === "-") {
    return "w-agg-event-day-neutral";
  }
  let groupIndex = 0;
  for (let rowIndex = 1; rowIndex <= index; rowIndex += 1) {
    const currentDay = eventDayKey(rows.value[rowIndex]);
    const previousDay = eventDayKey(rows.value[rowIndex - 1]);
    if (currentDay !== "-" && previousDay !== "-" && currentDay !== previousDay) {
      groupIndex += 1;
    }
  }
  return `w-agg-event-day-group-${(groupIndex % 2) + 1}`;
}

function isEventDayStart(index: number): boolean {
  if (index <= 0) {
    return false;
  }
  const currentDay = eventDayKey(rows.value[index]);
  const previousDay = eventDayKey(rows.value[index - 1]);
  return currentDay !== "-" && previousDay !== "-" && currentDay !== previousDay;
}

const SortLink = defineComponent({
  props: {
    link: { type: Object as PropType<AnyRecord>, required: true },
    label: { type: String, required: true }
  },
  setup(props) {
    return () => {
      const link = asRecord(props.link);
      const active = Boolean(link.Active);
      const desc = Boolean(link.Desc);
      return h("a", {
        class: ["w-agg-event-sort-link", "sc-sort-link", active ? "active" : ""],
        href: normalizeHref(text(link, "Href") || "#")
      }, [
        props.label,
        active ? h("span", { class: "w-agg-event-sort-indicator" }, desc ? "↓" : "↑") : null
      ]);
    };
  }
});

const WAggEventPagination = defineComponent({
  props: {
    links: { type: Array as PropType<AnyRecord[]>, required: true }
  },
  setup(props) {
    return () => h("nav", { "aria-label": "Page navigation" }, [
      h("ul", { class: "pagination pagination-sm mb-0" }, props.links.map((item, index) => {
        const row = asRecord(item);
        const ellipsis = Boolean(row.Ellipsis);
        const disabled = Boolean(row.Disabled) || ellipsis;
        const active = Boolean(row.Active);
        const label = text(row, "Label") || "...";
        return h("li", {
          key: `${label}-${index}`,
          class: ["page-item", active ? "active" : "", disabled ? "disabled" : ""]
        }, ellipsis
          ? h("span", { class: "page-link" }, "...")
          : h("a", {
            class: "page-link",
            href: disabled ? undefined : normalizeHref(text(row, "Href") || "#"),
            tabindex: disabled ? -1 : undefined,
            "aria-disabled": disabled ? "true" : undefined
          }, label));
      }))
    ]);
  }
});
</script>
