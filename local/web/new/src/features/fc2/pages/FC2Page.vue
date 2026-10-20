<template>
  <PageShell class="sc-events-page-shell fc2-page-shell" :title="pageTitle" :note="pageNote" eyebrow="">
    <div
      id="page-root"
      class="container-fluid sc-events-page-root next02-sc-pink-theme next02-sc-events-rewrite bucket-list-standard sh-bucket-list-page fc2-page fc2-list-page"
    >
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell sh-bucket-hero fc2-hero">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">FC2 / SH THREAD DATA</span>
            <h4>{{ pageTitle }}</h4>
            <p>{{ pageNote }}</p>
          </div>
          <div class="next02-sc-hero-actions">
            <a href="/shpool" class="btn btn-outline-secondary" target="_blank" rel="noopener noreferrer">SH 帖子池</a>
          </div>
        </section>

        <div
          class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card sh-bucket-table-card fc2-table-card"
        >
          <div class="card-body">
            <div
              class="bucket-list-controls sh-bucket-controls sh-bucket-toolbar fc2-filter-toolbar"
              aria-label="FC2 筛选栏"
            >
              <div class="bucket-list-control-main sh-bucket-toolbar-main">
                <div class="sh-bucket-toolbar-bottom">
                  <div class="sh-bucket-toolbar-filter-area">
                    <form
                      class="bucket-list-filter-row sh-bucket-filter-row fc2-filter-row"
                      method="get"
                      action="/fclist"
                    >
                      <input type="hidden" name="p" value="1" />
                      <input type="hidden" name="sort" :value="queryText('Sort') || 'post_time'" />
                      <input type="hidden" name="order" :value="queryText('Order') || 'desc'" />
                      <label class="bucket-filter-field fc2-keyword-field"
                        ><span>关键词</span
                        ><input
                          class="form-control"
                          name="keyword"
                          :value="queryText('Keyword')"
                          placeholder="番号 / 影片名称 / 出演信息"
                      /></label>
                      <label class="bucket-filter-field fc2-status-field"
                        ><span>数据状态</span
                        ><select class="form-select" name="status" :value="queryText('Status')">
                          <option value="">全部</option>
                          <option value="ready">ready</option>
                          <option value="conflict">conflict</option>
                        </select></label
                      >
                      <label class="bucket-filter-field is-page-size fc2-page-size-field"
                        ><span>每页</span
                        ><input
                          class="form-control"
                          type="number"
                          min="1"
                          name="ps"
                          :value="queryText('PageSize') || '50'"
                      /></label>
                      <div class="bucket-list-filter-actions">
                        <button class="btn bucket-list-apply" type="submit">筛选</button
                        ><a class="btn bucket-list-clear" href="/fclist">清空</a>
                      </div>
                    </form>
                  </div>
                  <div class="bucket-list-control-meta sh-bucket-toolbar-status">
                    <div class="sh-bucket-toolbar-metrics" aria-label="FC2 页面摘要">
                      <span :title="`总数 ${total}`" :aria-label="`总数 ${total}`"
                        ><strong>{{ total }}</strong></span
                      >
                      <span :title="`当前页 ${rows.length}`" :aria-label="`当前页 ${rows.length}`"
                        ><strong>{{ rows.length }}</strong></span
                      >
                    </div>
                    <PaginationBar v-if="paginationLinks.length" :links="paginationLinks" />
                  </div>
                </div>
              </div>
            </div>

            <div class="table-responsive fc2-table-wrap">
              <table
                class="table table-hover align-middle mb-0 sc-events-table sh-bucket-table fc2-table fc2-list-table"
              >
                <colgroup>
                  <col class="fc2-col-code" />
                  <col class="fc2-col-post-time" />
                  <col class="fc2-col-title" />
                  <col class="fc2-col-performer" />
                  <col class="fc2-col-play" />
                  <col class="fc2-col-download-time" />
                  <col class="fc2-col-sc-times" />
                  <col class="fc2-col-come-times" />
                  <col class="fc2-col-last-selected-sc-time" />
                </colgroup>
                <thead>
                  <tr>
                    <th>
                      <a class="fc2-sort-link" :href="sortHref('fc2_key')">影片标识{{ sortMark("fc2_key") }}</a>
                    </th>
                    <th>
                      <a class="fc2-sort-link" :href="sortHref('post_time')">发布时间{{ sortMark("post_time") }}</a>
                    </th>
                    <th>
                      <a class="fc2-sort-link" :href="sortHref('title')">影片名称{{ sortMark("title") }}</a>
                    </th>
                    <th>
                      <a class="fc2-sort-link" :href="sortHref('performer_text')"
                        >出演信息{{ sortMark("performer_text") }}</a
                      >
                    </th>
                    <th>
                      <a class="fc2-sort-link" :href="sortHref('play')">播放{{ sortMark("play") }}</a>
                    </th>
                    <th>
                      <a class="fc2-sort-link" :href="sortHref('birth_time')">下载时间{{ sortMark("birth_time") }}</a>
                    </th>
                    <th>
                      <a class="fc2-sort-link" :href="sortHref('sc_times')">SC 次数{{ sortMark("sc_times") }}</a>
                    </th>
                    <th>
                      <a class="fc2-sort-link" :href="sortHref('come_times')">Come 次数{{ sortMark("come_times") }}</a>
                    </th>
                    <th>
                      <a class="fc2-sort-link" :href="sortHref('last_selected_sc_time')"
                        >最近入选 SC{{ sortMark("last_selected_sc_time") }}</a
                      >
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in rows" :key="rowText(row, 'FC2Key')">
                    <td>
                      <a class="fc2-code-link" :href="detailHref(row)" target="_blank" rel="noopener noreferrer">{{
                        rowText(row, "FC2Key")
                      }}</a>
                    </td>
                    <td>
                      <a
                        v-if="rowText(row, 'MetadataThreadTID')"
                        class="fc2-thread-link"
                        :href="`/shthread/${encodeURIComponent(rowText(row, 'MetadataThreadTID'))}`"
                        target="_blank"
                        rel="noopener noreferrer"
                        ><ScDateTime :value="rowValue(row, 'PostTime')" unix /></a
                      ><ScDateTime v-else :value="rowValue(row, 'PostTime')" unix />
                    </td>
                    <td class="fc2-title-cell">
                      <a
                        :href="detailHref(row)"
                        :title="rowText(row, 'Title')"
                        target="_blank"
                        rel="noopener noreferrer"
                        >{{ rowText(row, "Title") || "-" }}</a
                      >
                    </td>
                    <td class="fc2-performer-cell" :title="rowText(row, 'PerformerText')">
                      {{ rowText(row, "PerformerText") || "-" }}
                    </td>
                    <td>
                      <a
                        v-if="rowValue(row, 'ShowPlayButton')"
                        class="btn btn-sm fc2-play-button"
                        :class="rowText(row, 'PlayButtonClass')"
                        :href="rowText(row, 'VideoURL')"
                        target="_blank"
                        rel="noopener noreferrer"
                        >{{ rowText(row, "PlayButtonText") || "Play" }}</a
                      ><span v-else>-</span>
                    </td>
                    <td>
                      <ScDateTime
                        v-if="rowValue(row, 'DownloadTime')"
                        :value="rowValue(row, 'DownloadTime')"
                        unix
                      /><span v-else>-</span>
                    </td>
                    <td>
                      <strong>{{ rowText(row, "ScTimes") || "0" }}</strong>
                    </td>
                    <td>
                      <strong>{{ rowText(row, "ComeTimes") || "0" }}</strong>
                    </td>
                    <td>
                      <ScDateTime
                        v-if="rowValue(row, 'LastSelectedScTime')"
                        :value="rowValue(row, 'LastSelectedScTime')"
                        unix
                      /><span v-else>-</span>
                    </td>
                  </tr>
                  <tr v-if="!rows.length">
                    <td colspan="9"><div class="sc-empty-state fc2-empty-state">当前筛选下没有 FC2 影片数据。</div></td>
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
import "@/features/fc2/styles/index.css";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/features/fc2/api";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import type { FC2Record } from "@/features/fc2/types";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<FC2Record>({});
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

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title") || "FC2影片列表");
const pageNote = computed(() => text(payload.value, "PageNote"));
const rows = computed(() => asArray<FC2Record>(pick(payload.value, "Rows", "rows")));
const query = computed(() => asRecord(pick(payload.value, "Query", "query")));
const total = computed(() => Number(pick(payload.value, "Total", "total") || 0));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")));
const pageLinks = computed(() => asArray<FC2Record>(pick(pageInfo.value, "Links", "links")));
const paginationLinks = computed(() => pageLinks.value.map((item) => ({ ...item, Href: pageHref(item) })));

function rowValue(row: FC2Record, key: string): unknown {
  return pick(row, key);
}
function rowText(row: FC2Record, key: string): string {
  return text(row, key);
}
function queryText(key: string): string {
  return text(query.value, key);
}
function detailHref(row: FC2Record): string {
  return `/fcdetail/${encodeURIComponent(rowText(row, "FC2Key"))}`;
}
function currentParams(): URLSearchParams {
  const params = new URLSearchParams();
  const pageSize = queryText("PageSize") || "50";
  const keyword = queryText("Keyword").trim();
  const status = queryText("Status").trim();
  const sort = queryText("Sort").trim();
  const order = queryText("Order").trim();
  if (pageSize !== "50") params.set("ps", pageSize);
  if (keyword) params.set("keyword", keyword);
  if (status) params.set("status", status);
  if (sort && sort !== "post_time") params.set("sort", sort);
  if (order && order !== "desc") params.set("order", order);
  return params;
}
function sortHref(field: string): string {
  const params = currentParams();
  params.set("sort", field);
  params.set("order", queryText("Sort") === field && queryText("Order") === "desc" ? "asc" : "desc");
  const queryString = params.toString();
  return `/fclist${queryString ? `?${queryString}` : ""}`;
}
function sortMark(field: string): string {
  if (queryText("Sort") !== field) return "";
  return queryText("Order") === "asc" ? " ↑" : " ↓";
}
function pageHref(item: FC2Record): string {
  const params = currentParams();
  const page = Number(pick(item, "Page", "page") || 1);
  if (page > 1) params.set("p", String(page));
  const queryString = params.toString();
  return `/fclist${queryString ? `?${queryString}` : ""}`;
}
</script>
