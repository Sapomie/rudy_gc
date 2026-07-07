<template>
  <PageShell class="torrent-albums-page-shell casts-page-shell" title="Casts" note="" eyebrow="">
    <div id="page-root" class="container-fluid torrent-albums-page casts-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div class="d-flex align-items-center justify-content-between mb-3">
          <div>
            <h4 class="mb-1">{{ pageTitle }}</h4>
            <p class="text-muted mb-0">浏览演员统计，并进入详情页查看该演员的 SC 历史和代表影片。</p>
          </div>
        </div>

        <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
          <div class="sc-page-summary">
            <span class="sc-page-summary-label">Total Casts</span>
            <span class="sc-page-summary-value">{{ total }}</span>
          </div>
          <div v-if="hasRows && pageLinks.length" class="ms-xl-auto">
            <CastsPagination :links="pageLinks" />
          </div>
        </div>

        <div class="card mb-3">
          <div class="card-body">
            <form class="row g-3 align-items-end" method="get" action="/casts" id="castsFilterForm">
              <input v-if="queryValue('PageSize')" type="hidden" name="ps" :value="queryValue('PageSize')">
              <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')">
              <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')">

              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1" for="casts-keyword">Name Search</label>
                <input id="casts-keyword" class="form-control" type="text" name="keyword" placeholder="Name / Chinese / Alias" :value="queryValue('Keyword')">
              </div>
              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">WOwned Range</label>
                <div class="row g-2">
                  <div class="col-6">
                    <input id="casts-owned-min" class="form-control" type="number" min="0" name="wowned_min" placeholder="Min" :value="queryValue('OwnedMin')">
                  </div>
                  <div class="col-6">
                    <input id="casts-owned-max" class="form-control" type="number" min="0" name="wowned_max" placeholder="Max" :value="queryValue('OwnedMax')">
                  </div>
                </div>
              </div>
              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">SC Times Range</label>
                <div class="row g-2">
                  <div class="col-6">
                    <input id="casts-sc-times-min" class="form-control" type="number" min="0" name="sc_times_min" placeholder="Min" :value="queryValue('ScTimesMin')">
                  </div>
                  <div class="col-6">
                    <input id="casts-sc-times-max" class="form-control" type="number" min="0" name="sc_times_max" placeholder="Max" :value="queryValue('ScTimesMax')">
                  </div>
                </div>
              </div>
              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">Come Times Range</label>
                <div class="row g-2">
                  <div class="col-6">
                    <input id="casts-come-times-min" class="form-control" type="number" min="0" name="come_times_min" placeholder="Min" :value="queryValue('ComeTimesMin')">
                  </div>
                  <div class="col-6">
                    <input id="casts-come-times-max" class="form-control" type="number" min="0" name="come_times_max" placeholder="Max" :value="queryValue('ComeTimesMax')">
                  </div>
                </div>
              </div>
              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">Last SC Range</label>
                <div class="row g-2">
                  <div class="col-6">
                    <input id="casts-last-sc-from" class="form-control" type="date" name="last_sc_from" :value="queryValue('LastScFrom')">
                  </div>
                  <div class="col-6">
                    <input id="casts-last-sc-to" class="form-control" type="date" name="last_sc_to" :value="queryValue('LastScTo')">
                  </div>
                </div>
              </div>
              <div class="col-12 d-flex gap-2">
                <button class="btn btn-primary" type="submit">Apply</button>
                <a class="btn btn-outline-secondary" href="/casts">Clear</a>
              </div>
            </form>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 list-table casts-table">
                <thead>
                  <tr>
                    <th><SortLink :link="sortLink('ByName')" label="Name" /></th>
                    <th><SortLink :link="sortLink('ByChinese')" label="Chinese" /></th>
                    <th><SortLink :link="sortLink('ByAge')" label="Age" /></th>
                    <th><SortLink :link="sortLink('ByHeight')" label="Height" /></th>
                    <th><SortLink :link="sortLink('ByMovieNumber')" label="Movies" /></th>
                    <th><SortLink :link="sortLink('ByWOwned')" label="WOwned" /></th>
                    <th><SortLink :link="sortLink('ByScTimes')" label="SC Times" /></th>
                    <th><SortLink :link="sortLink('ByComeTimes')" label="Come Times" /></th>
                    <th><SortLink :link="sortLink('ByLastSc')" label="Last SC" /></th>
                    <th><SortLink :link="sortLink('ByHighestRank')" label="Highest Rank" /></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td colspan="10"><div class="sc-empty-state">还没有可展示的演员。</div></td>
                  </tr>
                  <tr v-for="row in rows" v-else :key="rowKey(row)">
                    <td><a class="sc-event-link" :href="`/cast/${value(row, 'Id')}`">{{ value(row, "Name") }}</a></td>
                    <td>{{ value(row, "Chinese") || "-" }}</td>
                    <td>{{ formatAgeFromBirth(value(row, "BirthDay")) }}</td>
                    <td>{{ formatHeightCm(value(row, "Height")) }}</td>
                    <td>{{ value(row, "MovieNumber") || 0 }}</td>
                    <td>{{ value(row, "OwnedWMediaNumber") || 0 }}</td>
                    <td>{{ value(row, "ScTimes") || 0 }}</td>
                    <td>{{ value(row, "ComeTimes") || 0 }}</td>
                    <td class="text-nowrap">{{ formatUnixDate(value(row, "LastScTime")) }}</td>
                    <td>{{ formatRank(value(row, "HighestRank")) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </template>
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
const rows = ref<AnyRecord[]>([]);
const loading = ref(false);
const error = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const data = asRecord(await getJSON(apiPathForRoute(route)));
    payload.value = data;
    rows.value = asArray<AnyRecord>(pick(data, "Items", "items", "Rows", "rows"));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title") || "Casts");
const query = computed(() => asRecord(pick(payload.value, "Query", "query")));
const total = computed(() => text(payload.value, "Total", "total") || rows.value.length);
const hasRows = computed(() => rows.value.length > 0);
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<AnyRecord>(pageInfo.value.Links || pageInfo.value.links));
const sortMap = computed(() => asRecord(pick(payload.value, "CastSortQuery", "castSortQuery")));

function queryValue(key: string): string {
  return text(query.value, key);
}

function value(row: AnyRecord, key: string): string {
  return text(row, key);
}

function rowKey(row: AnyRecord): string {
  return text(row, "Id", "id") || JSON.stringify(row);
}

function sortLink(key: string): AnyRecord {
  return asRecord(sortMap.value[key]);
}

function formatUnixDate(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function formatAgeFromBirth(input: unknown): string {
  const n = Number(input || 0);
  if (n <= 0) return "-";
  const birthday = new Date(n * 1000);
  const now = new Date();
  let age = now.getFullYear() - birthday.getFullYear();
  if (now.getMonth() < birthday.getMonth() || (now.getMonth() === birthday.getMonth() && now.getDate() < birthday.getDate())) {
    age -= 1;
  }
  return age >= 0 ? String(age) : "-";
}

function formatHeightCm(input: unknown): string {
  const n = Number(input || 0);
  return n > 0 ? `${text({ input }, "input")} cm` : "-";
}

function formatRank(input: unknown): string {
  const n = Number(input || 0);
  return n > 0 ? `#${text({ input }, "input")}` : "-";
}

const SortLink = defineComponent({
  name: "CastsSortLink",
  props: {
    link: { type: Object as PropType<AnyRecord>, required: true },
    label: { type: String, required: true }
  },
  setup(props) {
    return () => {
      const active = Boolean(props.link.Active || props.link.active);
      const desc = Boolean(props.link.Desc || props.link.desc);
      const suffix = active ? (desc ? " ↓" : " ↑") : "";
      return h("a", {
        class: ["sc-sort-link", "sort-link", { active }],
        href: normalizeHref(text(props.link, "Href", "href") || "#")
      }, `${props.label}${suffix}`);
    };
  }
});

const CastsPagination = defineComponent({
  name: "CastsPagination",
  props: {
    links: { type: Array as PropType<AnyRecord[]>, required: true }
  },
  setup(props) {
    return () => h("nav", { "aria-label": "Page navigation" }, [
      h("ul", { class: "pagination pagination-sm mb-0" }, props.links.map((item, index) => {
        if (item.Ellipsis || item.ellipsis) {
          return h("li", { class: "page-item disabled", key: `ellipsis-${index}` }, [
            h("span", { class: "page-link" }, "...")
          ]);
        }
        const disabled = Boolean(item.Disabled || item.disabled);
        const active = Boolean(item.Active || item.active);
        return h("li", { class: ["page-item", { active, disabled }], key: `${item.Label || item.label}-${index}` }, [
          h("a", {
            class: "page-link",
            href: disabled ? "#" : normalizeHref(text(item, "Href", "href") || "#"),
            tabindex: disabled ? "-1" : undefined,
            "aria-disabled": disabled ? "true" : undefined
          }, text(item, "Label", "label"))
        ]);
      }))
    ]);
  }
});
</script>
