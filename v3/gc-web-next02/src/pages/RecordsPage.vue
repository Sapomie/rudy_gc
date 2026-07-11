<template>
  <PageShell class="records-page-shell" title="Records" note="" eyebrow="">
    <div id="page-root" class="container-fluid records-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <div v-else id="crawlRecordListPage" class="crawl-record-list-page next02-records-rewrite">
        <div class="next02-records-hero">
          <div class="next02-records-hero-main">
            <div class="next02-records-kicker">Crawl Trace</div>
            <h4 class="mb-1">{{ pageTitle || "CrawlRecords" }}</h4>
            <p class="text-muted mb-0">{{ pageNote || "抓取记录、耗时和类型筛选追踪。" }}</p>
          </div>
        </div>

        <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3 next02-records-toolbar">
          <div class="crawl-record-summary">
            <span class="crawl-record-summary-label">总行数</span>
            <span class="crawl-record-summary-value">{{ total }}</span>
          </div>
          <div v-if="hasRows && pageLinks.length" class="ms-xl-auto">
            <RecordsPagination :links="pageLinks" />
          </div>
        </div>

        <div class="card mb-3">
          <div class="card-body">
            <form class="row g-3 align-items-end" method="get" action="/crawl-records">
              <input type="hidden" name="sort" :value="queryValue('Sort')">
              <input type="hidden" name="order" :value="queryValue('Order')">
              <input id="crawlRecordTypeInput" type="hidden" name="type" :value="queryValue('Type')">

              <div class="col-12 col-xl-8">
                <label class="form-label mb-2">Type</label>
                <div class="d-flex flex-wrap gap-2 crawl-record-type-switch">
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

              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">每页</label>
                <select id="crawlRecordPageSize" class="form-select" name="ps">
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

        <div class="card">
          <div class="card-body">
            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 crawl-record-table">
                <thead>
                  <tr>
                    <th class="text-nowrap"><SortLink :link="headerLink('StartedTime')" label="StartedTime" /></th>
                    <th class="text-nowrap"><SortLink :link="headerLink('Type')" label="Type" /></th>
                    <th class="text-nowrap"><SortLink :link="headerLink('DetailNumber')" label="DetailNumber" /></th>
                    <th class="text-nowrap"><SortLink :link="headerLink('DurationMin')" label="Duration(min)" /></th>
                    <th class="text-nowrap"><SortLink :link="headerLink('Name')" label="Name" /></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td colspan="5"><div class="crawl-record-empty-state">当前筛选下没有 `e_record` 数据。</div></td>
                  </tr>
                  <tr v-for="row in rows" v-else :key="rowKey(row)">
                    <td class="text-nowrap">{{ formatUnixMinute(value(row, "StartedTime")) }}</td>
                    <td class="text-nowrap crawl-record-type">{{ value(row, "Type") }}</td>
                    <td class="text-nowrap">{{ value(row, "DetailNumber") }}</td>
                    <td class="text-nowrap">{{ value(row, "DurationMinText") }}</td>
                    <td class="text-nowrap crawl-record-name">{{ value(row, "Name") }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div v-if="hasRows && pageLinks.length" class="d-flex justify-content-end mt-3">
          <RecordsPagination :links="pageLinks" />
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
const typeLinks = computed(() => asArray<AnyRecord>(pick(payload.value, "TypeLinks")));
const headerLinks = computed(() => asRecord(pick(payload.value, "HeaderLinks")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<AnyRecord>(pageInfo.value.Links));
const clearHref = computed(() => normalizeHref(text(payload.value, "ClearHref") || "/crawl-records"));

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
  return [value(row, "StartedTime"), value(row, "Type"), value(row, "Name")].join(":");
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
        class: ["crawl-record-sort-link", active ? "is-active" : ""],
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
