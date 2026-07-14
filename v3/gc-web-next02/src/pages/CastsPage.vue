<template>
  <PageShell class="next02-pink-list-page-shell sc-events-page-shell casts-page-shell" :title="pageTitle" note="" eyebrow="">
    <div id="page-root" class="container-fluid next02-pink-list-page next02-sc-pink-theme next02-sc-events-rewrite casts-page-root">
      <div v-if="loading" class="alert alert-info py-2 next02-pink-list-state">加载中...</div>
      <div v-else-if="error" class="alert alert-danger next02-pink-list-state">{{ error }}</div>
      <template v-else>
        <section class="next02-pink-list-hero casts-list-hero">
          <div class="next02-pink-list-hero-copy">
            <span class="next02-pink-list-eyebrow">CAST / DIRECTORY</span>
            <h4>{{ pageTitle }}</h4>
            <p>浏览演员统计，并进入详情页查看该演员的 SC 历史和代表影片。</p>
          </div>
        </section>

        <div class="card next02-pink-list-panel next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card casts-list-panel">
          <div class="card-body">
            <div class="next02-pink-list-toolbar casts-list-toolbar">
              <form class="casts-filter-form" method="get" action="/castlist" id="castsFilterForm">
                <input type="hidden" name="p" value="1">
                <input v-if="queryValue('PageSize')" type="hidden" name="ps" :value="queryValue('PageSize')">
                <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')">
                <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')">

                <label class="casts-filter-field casts-filter-keyword" for="casts-keyword">
                  <span>Name</span>
                  <input id="casts-keyword" type="text" name="keyword" placeholder="Name / Chinese / Alias" :value="queryValue('Keyword')">
                </label>
                <div class="casts-filter-field">
                  <span>WOwned</span>
                  <span class="casts-filter-range">
                    <input id="casts-owned-min" type="number" min="0" name="wowned_min" placeholder="Min" aria-label="WOwned minimum" :value="queryValue('OwnedMin')">
                    <span aria-hidden="true">–</span>
                    <input id="casts-owned-max" type="number" min="0" name="wowned_max" placeholder="Max" aria-label="WOwned maximum" :value="queryValue('OwnedMax')">
                  </span>
                </div>
                <div class="casts-filter-field">
                  <span>SC Times</span>
                  <span class="casts-filter-range">
                    <input id="casts-sc-times-min" type="number" min="0" name="sc_times_min" placeholder="Min" aria-label="SC Times minimum" :value="queryValue('ScTimesMin')">
                    <span aria-hidden="true">–</span>
                    <input id="casts-sc-times-max" type="number" min="0" name="sc_times_max" placeholder="Max" aria-label="SC Times maximum" :value="queryValue('ScTimesMax')">
                  </span>
                </div>
                <div class="casts-filter-field">
                  <span>Come Times</span>
                  <span class="casts-filter-range">
                    <input id="casts-come-times-min" type="number" min="0" name="come_times_min" placeholder="Min" aria-label="Come Times minimum" :value="queryValue('ComeTimesMin')">
                    <span aria-hidden="true">–</span>
                    <input id="casts-come-times-max" type="number" min="0" name="come_times_max" placeholder="Max" aria-label="Come Times maximum" :value="queryValue('ComeTimesMax')">
                  </span>
                </div>
                <div class="casts-filter-field casts-filter-date">
                  <span>Last SC</span>
                  <span class="casts-filter-range">
                    <input id="casts-last-sc-from" class="smart-pick-date-input" type="text" name="last_sc_from" aria-label="Last SC from" :value="queryValue('LastScFrom')" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
                    <span aria-hidden="true">–</span>
                    <input id="casts-last-sc-to" class="smart-pick-date-input" type="text" name="last_sc_to" aria-label="Last SC to" :value="queryValue('LastScTo')" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
                  </span>
                </div>
                <button class="next02-pink-list-submit" type="submit">Apply</button>
                <a class="next02-pink-list-clear" href="/castlist">Clear</a>
              </form>

              <div class="casts-list-controls">
                <div class="next02-pink-list-metrics" aria-label="演员列表统计">
                  <span><strong>{{ total }}</strong> Total Casts</span>
                </div>
                <div v-if="hasRows && pageLinks.length" class="next02-pink-list-pagination">
                  <CastsPagination :links="pageLinks" />
                </div>
              </div>
            </div>

            <div class="table-responsive next02-pink-list-table-wrap">
              <table class="table table-hover align-middle mb-0 next02-pink-list-table sc-events-table casts-table">
                <colgroup>
                  <col class="casts-col-name">
                  <col class="casts-col-chinese">
                  <col class="casts-col-age">
                  <col class="casts-col-height">
                  <col class="casts-col-movies">
                  <col class="casts-col-owned">
                  <col class="casts-col-sc-times">
                  <col class="casts-col-come-times">
                  <col class="casts-col-last-sc">
                  <col class="casts-col-highest-rank">
                </colgroup>
                <thead>
                  <tr>
                    <th class="next02-pink-list-sortable sc-sortable-head"><SortLink :link="sortLink('ByName')" label="Name" /></th>
                    <th class="next02-pink-list-sortable sc-sortable-head"><SortLink :link="sortLink('ByChinese')" label="Chinese" /></th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number"><SortLink :link="sortLink('ByAge')" label="Age" /></th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number"><SortLink :link="sortLink('ByHeight')" label="Height" /></th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number"><SortLink :link="sortLink('ByMovieNumber')" label="Movies" /></th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number"><SortLink :link="sortLink('ByWOwned')" label="WOwned" /></th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number"><SortLink :link="sortLink('ByScTimes')" label="SC Times" /></th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number"><SortLink :link="sortLink('ByComeTimes')" label="Come Times" /></th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number"><SortLink :link="sortLink('ByLastSc')" label="Last SC" /></th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number"><SortLink :link="sortLink('ByHighestRank')" label="Highest Rank" /></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td colspan="10"><div class="sc-empty-state">还没有可展示的演员。</div></td>
                  </tr>
                  <tr v-for="row in rows" v-else :key="rowKey(row)">
                    <td class="sc-content-cell casts-content-cell"><a class="next02-pink-list-link" :href="`/castdetail/${value(row, 'Id')}`" :title="value(row, 'Name')" target="_blank" rel="noopener noreferrer">{{ value(row, "Name") }}</a></td>
                    <td class="sc-content-cell casts-content-cell" :title="value(row, 'Chinese') || undefined">{{ value(row, "Chinese") || "-" }}</td>
                    <td class="sc-number casts-number">{{ formatAgeFromBirth(value(row, "BirthDay")) }}</td>
                    <td class="sc-number casts-number">{{ formatHeightCm(value(row, "Height")) }}</td>
                    <td class="sc-number casts-number">{{ value(row, "MovieNumber") || 0 }}</td>
                    <td class="sc-number casts-number">{{ value(row, "OwnedWMediaNumber") || 0 }}</td>
                    <td class="sc-number casts-number">{{ value(row, "ScTimes") || 0 }}</td>
                    <td class="sc-number casts-number">{{ value(row, "ComeTimes") || 0 }}</td>
                    <td class="sc-number casts-number">{{ formatUnixDate(value(row, "LastScTime")) }}</td>
                    <td class="sc-number casts-number">{{ formatRank(value(row, "HighestRank")) }}</td>
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
        class: ["next02-pink-list-sort-link", "sc-sort-link", "sort-link", { active }],
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
