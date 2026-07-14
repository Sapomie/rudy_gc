<template>
  <PageShell class="records-page-shell sc-events-page-shell" title="Records" note="" eyebrow="">
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <div v-else id="page-root" class="records-page-root next02-sc-pink-theme next02-sc-events-rewrite">
      <div id="crawlRecordListPage" class="crawl-record-list-page next02-records-rewrite">
        <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">CRAWL / RECORD POOL</span>
            <h4>{{ pageTitle || "CrawlRecords" }}</h4>
            <p>{{ pageNote || "抓取记录、耗时和类型筛选追踪。" }}</p>
          </div>
        </section>

        <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
          <div class="card-body">
            <div class="next02-sc-panel-head">
              <form class="sc-event-filter-form crawl-record-filter-form" method="get" action="/crawlrecordlist">
                <input type="hidden" name="sort" :value="queryValue('Sort')">
                <input type="hidden" name="order" :value="queryValue('Order')">
                <input id="crawlRecordTypeInput" type="hidden" name="type" :value="queryValue('Type')">

                <div class="crawl-record-filter-cluster crawl-record-filter-types" aria-label="记录类型">
                  <span class="crawl-record-filter-cluster-label">Type</span>
                  <div class="crawl-record-type-switch">
                    <a
                      v-for="item in typeLinks"
                      :key="linkLabel(item)"
                      :href="linkHref(item)"
                      class="btn btn-sm"
                      :class="linkActive(item) ? 'btn-primary' : 'btn-outline-primary'"
                    >
                      {{ linkLabel(item) }}
                    </a>
                  </div>
                </div>

                <div class="crawl-record-filter-cluster crawl-record-filter-options">
                  <label class="sc-event-filter-field crawl-record-page-size-field" for="crawlRecordPageSize">
                    <span>每页</span>
                    <input id="crawlRecordPageSize" class="form-control" type="number" step="1" name="ps" :value="queryValue('PageSize') || '20'">
                  </label>

                  <div class="crawl-record-filter-actions">
                    <button class="sc-event-filter-submit" type="submit">筛选</button>
                    <a class="sc-event-filter-clear" :href="clearHref">清空</a>
                  </div>
                </div>
              </form>

              <div v-if="hasRows" class="next02-sc-hero-metrics sc-event-list-metrics" aria-label="抓取记录摘要">
                <span><strong>{{ total }}</strong> Records</span>
                <span><strong>{{ rows.length }}</strong> Current Page</span>
              </div>

              <div v-if="hasRows && pageLinks.length" class="next02-sc-movie-pager">
                <RecordsPagination :links="pageLinks" />
              </div>
            </div>

            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 sc-events-table crawl-record-table">
                <colgroup>
                  <col class="crawl-record-col-started">
                  <col class="crawl-record-col-type">
                  <col class="crawl-record-col-detail">
                  <col class="crawl-record-col-duration">
                  <col class="crawl-record-col-name">
                </colgroup>
                <thead>
                  <tr>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('StartedTime')" label="StartedTime" /></th>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('Type')" label="Type" /></th>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('DetailNumber')" label="DetailNumber" /></th>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('DurationMin')" label="Duration(min)" /></th>
                    <th class="sc-sortable-head"><SortLink :link="headerLink('Name')" label="Name" /></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td colspan="5"><div class="sc-empty-state crawl-record-empty-state">当前筛选下没有 `e_record` 数据。</div></td>
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
                        <small v-if="recordTimeParts(value(row, 'StartedTime')).time" class="sc-event-hour">{{ recordTimeParts(value(row, "StartedTime")).time }}</small>
                      </span>
                    </td>
                    <td class="sc-content-cell"><span class="crawl-record-type" :title="value(row, 'Type')">{{ value(row, "Type") }}</span></td>
                    <td class="sc-number sc-primary-value">{{ value(row, "DetailNumber") }}</td>
                    <td class="sc-number">{{ value(row, "DurationMinText") }}</td>
                    <td class="sc-content-cell"><span class="crawl-record-name" :title="value(row, 'Name')">{{ value(row, "Name") }}</span></td>
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
import { apiPathForRoute, getJSON } from "@/shared/api/request";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import type { AnyRecord, PageInfo } from "@/shared/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/shared/utils/format";

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
const typeLinks = computed(() => asArray<AnyRecord>(pick(payload.value, "TypeLinks")));
const headerLinks = computed(() => asRecord(pick(payload.value, "HeaderLinks")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<AnyRecord>(pageInfo.value.Links));
const clearHref = computed(() => normalizeHref(text(payload.value, "ClearHref") || "/crawlrecordlist"));

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
  return [value(row, "StartedTime"), value(row, "Type"), value(row, "Name"), index].join(":");
}

function linkLabel(item: AnyRecord): string {
  return text(item, "Label", "label");
}

function linkHref(item: AnyRecord): string {
  return normalizeHref(text(item, "Href", "href") || "#");
}

function linkActive(item: AnyRecord): boolean {
  return Boolean(item.Active || item.active);
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

function recordDayKey(row: AnyRecord): string {
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

const SortLink = defineComponent({
  props: {
    link: { type: Object, required: true },
    label: { type: String, required: true }
  },
  setup(props) {
    return () => {
      const link = asRecord(props.link);
      const active = Boolean(link.Active);
      const desc = Boolean(link.Desc);
      return h("a", {
        class: ["crawl-record-sort-link", "sc-sort-link", active ? "active" : ""],
        href: normalizeHref(text(link, "Href") || "#")
      }, [
        props.label,
        active ? h("span", { class: "crawl-record-sort-indicator" }, desc ? "↓" : "↑") : null
      ]);
    };
  }
});

const RecordsPagination = defineComponent({
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
