<template>
  <PageShell class="sc-events-page-shell sh-bucket-page-shell" :title="pageTitle" :note="pageNote" eyebrow="">
    <div
      id="page-root"
      class="container-fluid sc-events-page-root next02-sc-pink-theme next02-sc-events-rewrite bucket-list-standard sh-bucket-list-page"
    >
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell sh-bucket-hero">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">SH / TIME BUCKETS</span>
            <h4>{{ pageTitle }}</h4>
            <p>{{ pageNote }}</p>
          </div>
          <div class="next02-sc-hero-actions">
            <a href="/shpool" class="btn btn-outline-secondary" target="_blank" rel="noopener noreferrer">SH 帖子池</a>
            <a href="/aggbackfill" class="btn btn-outline-secondary" target="_blank" rel="noopener noreferrer"
              >聚合回填</a
            >
          </div>
        </section>

        <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card sh-bucket-table-card">
          <div class="card-body">
            <div class="bucket-list-controls sh-bucket-controls sh-bucket-toolbar" aria-label="SH 聚合桶筛选栏">
              <div class="bucket-list-control-main sh-bucket-toolbar-main">
                <div class="sh-bucket-toolbar-bottom">
                  <div class="sh-bucket-toolbar-filter-area">
                    <div class="bucket-list-switch-group bucket-level-switch">
                      <span>层级</span
                      ><a
                        v-for="item in levelLinks"
                        :key="linkKey(item)"
                        :href="linkHref(item)"
                        class="bucket-list-switch"
                        :class="{ 'is-active': linkActive(item) }"
                        >{{ linkLabel(item) }}</a
                      >
                    </div>
                    <div class="bucket-list-switch-group bucket-scope-switch">
                      <span>聚合范围</span
                      ><a
                        v-for="item in scopeLinks"
                        :key="item.value"
                        :href="scopeHref(item.value)"
                        class="bucket-list-switch"
                        :class="{ 'is-active': scopeActive(item.value) }"
                        >{{ item.label }}</a
                      >
                    </div>
                    <form class="bucket-list-filter-row sh-bucket-filter-row" method="get" :action="listPath">
                      <input type="hidden" name="p" value="1" />
                      <input type="hidden" name="sort" :value="queryValue('Sort')" />
                      <input type="hidden" name="dir" :value="queryValue('Dir')" />
                      <input type="hidden" name="level" :value="queryValue('Level')" />
                      <input type="hidden" name="agg_scope" :value="queryValue('AggScope') || 'all'" />
                      <label class="bucket-filter-field is-date"
                        ><span>起始时间</span
                        ><input
                          class="form-control smart-pick-date-input"
                          type="text"
                          name="time_from"
                          :value="queryValue('TimeFrom')"
                          placeholder="YYYY-MM-DD"
                          inputmode="numeric"
                          maxlength="10"
                          pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                      /></label>
                      <label class="bucket-filter-field is-date"
                        ><span>结束时间</span
                        ><input
                          class="form-control smart-pick-date-input"
                          type="text"
                          name="time_to"
                          :value="queryValue('TimeTo')"
                          placeholder="YYYY-MM-DD"
                          inputmode="numeric"
                          maxlength="10"
                          pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                      /></label>
                      <label class="bucket-filter-field is-page-size"
                        ><span>每页</span
                        ><input class="form-control" type="number" name="ps" :value="queryValue('PageSize') || '100'"
                      /></label>
                      <div class="bucket-list-filter-actions">
                        <button class="btn bucket-list-apply" type="submit">筛选</button
                        ><a class="btn bucket-list-clear" :href="clearHref">清空</a>
                      </div>
                    </form>
                  </div>
                  <div class="bucket-list-control-meta sh-bucket-toolbar-status">
                    <div class="sh-bucket-toolbar-metrics" aria-label="SH 聚合桶页面摘要">
                      <span :title="`总数 ${total}`" :aria-label="`总数 ${total}`"
                        ><strong>{{ total }}</strong></span
                      >
                      <span :title="`当前页 ${rows.length}`" :aria-label="`当前页 ${rows.length}`"
                        ><strong>{{ rows.length }}</strong></span
                      >
                    </div>
                    <PaginationBar v-if="pageLinks.length" :links="pageLinks" />
                  </div>
                </div>
              </div>
            </div>

            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 sc-events-table sh-bucket-table">
                <colgroup>
                  <col class="sh-bucket-col-time" />
                  <col class="sh-bucket-col-count" />
                  <col class="sh-bucket-col-count" />
                  <col class="sh-bucket-col-count" />
                  <col class="sh-bucket-col-count" />
                  <col class="sh-bucket-col-resource" />
                  <col class="sh-bucket-col-resource" />
                  <col class="sh-bucket-col-latest" />
                </colgroup>
                <thead>
                  <tr>
                    <th class="sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="headerLink('Bucket')" indicator-class="ms-1" />
                    </th>
                    <th class="sc-number sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="headerLink('All')" indicator-class="ms-1" />
                    </th>
                    <th class="sc-number sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="headerLink('Pending')" indicator-class="ms-1" />
                    </th>
                    <th class="sc-number sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="headerLink('Fetched')" indicator-class="ms-1" />
                    </th>
                    <th class="sc-number sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="headerLink('Failed')" indicator-class="ms-1" />
                    </th>
                    <th class="sc-number sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="headerLink('Magnet')" indicator-class="ms-1" />
                    </th>
                    <th class="sc-number sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="headerLink('InfoHash')" indicator-class="ms-1" />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="headerLink('Latest')" indicator-class="ms-1" />
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, index) in rows" :key="rowKey(row)" :class="rowClass(row, index)">
                    <td class="sc-content-cell">
                      <a
                        class="sc-event-link"
                        :href="rowText(row, 'ViewHref')"
                        target="_blank"
                        rel="noopener noreferrer"
                        >{{ rowText(row, "BucketDisplay") }}</a
                      >
                    </td>
                    <td class="sc-number sc-primary-value">{{ rowText(row, "CountAll") }}</td>
                    <td class="sc-number">{{ rowText(row, "CountPending") }}</td>
                    <td class="sc-number">{{ rowText(row, "CountFetched") }}</td>
                    <td class="sc-number">{{ rowText(row, "CountFailed") }}</td>
                    <td class="sc-number">{{ rowText(row, "CountWithMagnet") }}</td>
                    <td class="sc-number">{{ rowText(row, "CountWithInfoHash") }}</td>
                    <td class="sc-muted"><ScDateTime :value="rowValue(row, 'LatestPostTime')" unix /></td>
                  </tr>
                  <tr v-if="!rows.length">
                    <td colspan="8">
                      <div class="sc-empty-state">
                        当前筛选下没有 SH 帖子时间桶数据，请先在聚合回填页执行 SH 聚合回填。
                      </div>
                    </td>
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
import "@/features/sh/styles/index.css";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/features/sh/api";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import type { PageInfo } from "@/shared/types/gc";
import type { SHRecord } from "@/features/sh/types";
import { asArray, asRecord, normalizeHref, pick, text } from "@/shared/utils/format";
import { loadSHForumOptions, type SHForumOption } from "@/shared/utils/shForums";

const route = useRoute();
const payload = ref<SHRecord>({});
const shForums = ref<SHForumOption[]>([]);
const loading = ref(false);
const error = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const [pagePayload, forumOptions] = await Promise.all([getJSON(apiPathForRoute(route)), loadSHForumOptions()]);
    payload.value = asRecord(pagePayload);
    shForums.value = forumOptions;
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title") || "SH 帖子时间桶落库列表");
const pageNote = computed(() => text(payload.value, "PageNote"));
const rows = computed(() => asArray<SHRecord>(pick(payload.value, "Rows", "rows")));
const query = computed(() => asRecord(pick(payload.value, "Query", "query")));
const total = computed(() => Number(pick(payload.value, "Total", "total") || 0));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<SHRecord>(pageInfo.value.Links || pageInfo.value.links));
const levelLinks = computed(() => asArray<SHRecord>(pick(payload.value, "LevelLinks")));
const scopeLinks = computed(() => [
  { value: "all", label: "全部帖子" },
  ...shForums.value.map((forum) => ({ value: `forum:${forum.forumID}`, label: forum.forumName }))
]);
const headerLinks = computed(() => asRecord(pick(payload.value, "HeaderLinks")));
const listPath = computed(() => text(payload.value, "ListPath") || "/shbucketlist");
const clearHref = computed(() => text(payload.value, "ClearHref") || "/shbucketlist");

function queryValue(key: string): string {
  return text(query.value, key);
}
function rowValue(row: SHRecord, key: string): unknown {
  return pick(row, key);
}
function rowText(row: SHRecord, key: string): string {
  return text(row, key);
}
function rowKey(row: SHRecord): string {
  return `${rowText(row, "AggScope")}-${rowText(row, "BucketDisplay")}`;
}
function headerLink(key: string): SHRecord {
  return asRecord(headerLinks.value[key]);
}
function linkLabel(item: SHRecord): string {
  return text(item, "Label", "label");
}
function linkHref(item: SHRecord): string {
  return normalizeHref(text(item, "Href", "href") || "#");
}
function linkActive(item: SHRecord): boolean {
  return Boolean(item.Active || item.active);
}
function linkKey(item: SHRecord): string {
  return `${linkLabel(item)}-${linkHref(item)}`;
}

function scopeActive(value: string): boolean {
  return (queryValue("AggScope") || "all") === value;
}

function scopeHref(value: string): string {
  const params = new URLSearchParams();
  for (const [key, rawValue] of Object.entries(route.query)) {
    const values = Array.isArray(rawValue) ? rawValue : [rawValue];
    for (const item of values) {
      if (item != null) params.append(key, String(item));
    }
  }
  params.set("agg_scope", value);
  params.set("p", "1");
  return `${listPath.value}?${params.toString()}`;
}

function rowClass(row: SHRecord, index: number): Record<string, boolean> {
  const timestamp = Number(rowValue(row, "BucketStart") || 0);
  const month = timestamp > 0 ? new Date(timestamp * 1000).getMonth() : -1;
  const previous = index > 0 ? Number(rowValue(rows.value[index - 1], "BucketStart") || 0) : 0;
  const previousMonth = previous > 0 ? new Date(previous * 1000).getMonth() : -2;
  return {
    "sc-running-month-neutral": month < 0,
    [`sc-running-month-${(month % 3) + 1}`]: month >= 0,
    "sc-running-month-start": month !== previousMonth
  };
}
</script>
