<template>
  <PageShell class="w-agg-events-page-shell" title="WAggEvents" note="" eyebrow="">
    <div id="page-root" class="container-fluid w-agg-events-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <div v-else id="wAggEventListPage" class="w-agg-event-list-page">
        <div class="d-flex align-items-center justify-content-between mb-3">
          <div>
            <h4 class="mb-1">{{ pageTitle }}</h4>
            <p v-if="pageNote" class="text-muted mb-0">{{ pageNote }}</p>
          </div>
        </div>

        <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
          <div class="w-agg-event-summary">
            <span class="w-agg-event-summary-label">总行数</span>
            <span class="w-agg-event-summary-value">{{ total }}</span>
          </div>
          <div v-if="hasRows && pageLinks.length" class="ms-xl-auto">
            <WAggEventPagination :links="pageLinks" />
          </div>
        </div>

        <div class="section-card card mb-3">
          <div class="card-body">
            <form class="row g-3 align-items-end" method="get" action="/w-agg-events">
              <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')">
              <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')">

              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">AggKey</label>
                <input class="form-control" type="text" name="agg_key" :value="queryValue('AggKey')" placeholder="精确匹配">
              </div>

              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">FlowKey</label>
                <input class="form-control" type="text" name="flow_key" :value="queryValue('FlowKey')" placeholder="精确匹配">
              </div>

              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">Status</label>
                <input class="form-control" type="text" name="status" :value="queryValue('Status')" placeholder="精确匹配">
              </div>

              <div class="col-12 col-md-6 col-xl-1">
                <label class="form-label mb-1">每页</label>
                <select id="wAggEventPageSize" class="form-select" name="ps">
                  <option v-for="size in pageSizeOptions" :key="size" :value="size" :selected="Number(queryValue('PageSize')) === size">
                    {{ size }}
                  </option>
                </select>
              </div>

              <div class="col-12 col-xl-2 d-flex gap-2">
                <button class="btn btn-primary" type="submit">筛选</button>
                <a class="btn btn-outline-secondary" :href="clearHref">清空</a>
              </div>
            </form>
          </div>
        </div>

        <div class="section-card card">
          <div class="card-body">
            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 w-agg-event-table">
                <thead>
                  <tr>
                    <th class="text-nowrap"><SortLink :link="headerLink('StartedTime')" label="StartedTime" /></th>
                    <th class="text-nowrap"><SortLink :link="headerLink('AggKey')" label="AggKey" /></th>
                    <th class="text-nowrap"><SortLink :link="headerLink('FlowKey')" label="FlowKey" /></th>
                    <th class="text-nowrap"><SortLink :link="headerLink('ScopeCount')" label="ScopeCount" /></th>
                    <th class="text-nowrap"><SortLink :link="headerLink('BucketCount')" label="BucketCount" /></th>
                    <th class="text-nowrap"><SortLink :link="headerLink('TopCount')" label="TopCount" /></th>
                    <th class="text-nowrap"><SortLink :link="headerLink('FinishedTime')" label="FinishedTime" /></th>
                    <th class="text-nowrap"><SortLink :link="headerLink('Duration')" label="Duration(秒)" /></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td colspan="8"><div class="w-agg-event-empty-state">当前筛选下没有 `w_agg_event` 数据。</div></td>
                  </tr>
                  <tr v-for="row in rows" v-else :key="rowKey(row)" :data-item-id="value(row, 'ID')">
                    <td class="text-nowrap">{{ formatUnixMinute(value(row, "StartedTime")) }}</td>
                    <td class="text-nowrap w-agg-event-key">{{ value(row, "AggKey") }}</td>
                    <td class="text-nowrap w-agg-event-key">{{ value(row, "FlowKey") }}</td>
                    <td class="text-nowrap">{{ value(row, "ScopeCount") }}</td>
                    <td class="text-nowrap">{{ value(row, "BucketCount") }}</td>
                    <td class="text-nowrap">{{ value(row, "TopCount") }}</td>
                    <td class="text-nowrap">{{ formatUnixMinute(value(row, "FinishedTime")) }}</td>
                    <td class="text-nowrap">{{ value(row, "DurationSecondsText") }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div v-if="hasRows && pageLinks.length" class="d-flex justify-content-end mt-3">
          <WAggEventPagination :links="pageLinks" />
        </div>
      </div>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord, PageInfo } from "@/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const pageSizeOptions = [20, 50, 100, 200];

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
const clearHref = computed(() => normalizeHref(text(payload.value, "ClearHref") || "/w-agg-events"));

function queryValue(key: string): string {
  return text(query.value, key);
}

function headerLink(key: string): AnyRecord {
  return asRecord(headerLinks.value[key]);
}

function value(row: AnyRecord, key: string): string {
  return text(row, key);
}

function rowKey(row: AnyRecord): string {
  return [
    value(row, "StartedTime"),
    value(row, "AggKey"),
    value(row, "FlowKey"),
    value(row, "ScopeCount"),
    value(row, "BucketCount")
  ].join(":");
}

function formatUnixMinute(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")} ${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
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
        class: ["w-agg-event-sort-link", active ? "is-active" : ""],
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
