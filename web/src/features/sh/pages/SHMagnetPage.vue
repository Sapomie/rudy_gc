<template>
  <PageShell class="sc-events-page-shell" :title="pageTitle" :note="pageNote" eyebrow="">
    <div class="next02-sh-list-rewrite container-fluid next02-sc-pink-theme next02-sc-events-rewrite bucket-list-standard sh-bucket-list-page">
      <div v-if="loading" class="sc-empty-state">正在加载色花堂磁力数据...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <div v-else id="shMagnetListPage" class="sehuatang-list-content">
        <section class="next02-sc-hero next02-sc-events-hero next02-sh-list-hero next02-sc-movie-top-shell">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">色花堂 / 磁力资源</span>
            <h4>{{ pageTitle }}</h4>
            <p>{{ pageNote }}</p>
          </div>
          <div class="next02-sc-hero-actions next02-sh-list-actions">
            <div class="btn-group" role="group" aria-label="色花堂磁力页面操作">
              <a class="btn btn-outline-secondary" href="/shpool">SH 帖子池</a>
              <a class="btn btn-outline-secondary" href="/shfill" target="_blank" rel="noopener noreferrer">详情补全</a>
              <a class="btn btn-outline-secondary" href="/crawltasklist" target="_blank" rel="noopener noreferrer">任务列表</a>
            </div>
          </div>
        </section>

        <div v-if="feedback" class="alert mb-3" :class="feedbackOK ? 'alert-success' : 'alert-danger'" role="status">{{ feedback }}</div>

        <div class="card next02-sc-panel next02-sc-events-panel next02-sh-list-panel next02-sc-movie-table-card">
          <div class="card-body">
            <div class="next02-sc-panel-head next02-sh-list-control">
              <form class="next02-sh-filter-form bucket-list-controls sh-bucket-controls sh-bucket-toolbar next02-sh-filter-toolbar" method="get" action="/shmagnetlist">
                <input type="hidden" name="p" value="1">
                <input type="hidden" name="sort" :value="queryText('Sort')">
                <input type="hidden" name="order" :value="queryText('Order')">
                <input type="hidden" name="movie_type" :value="queryText('MovieType')">
                <input type="hidden" name="download_state" :value="queryText('DownloadState')">
                <input type="hidden" name="source_forum_id" :value="queryText('SourceForumID')">
                <div class="bucket-list-control-main sh-bucket-toolbar-main next02-sh-filter-toolbar-main">
                  <div class="next02-sh-filter-toolbar-top">
                    <div class="sh-bucket-toolbar-filter-area next02-sh-filter-query" aria-label="检索条件">
                      <label class="next02-sh-filter-field is-keyword"><span>关键词</span><input class="form-control" name="keyword" :value="queryText('Keyword')" placeholder="影片 / 帖子 / TID"></label>
                      <label class="next02-sh-filter-field is-hash"><span>影片编号</span><input class="form-control" name="movie_jav_id" :value="queryText('MovieJavID')" placeholder="JavLib 编号"></label>
                      <label class="next02-sh-filter-field is-hash"><span>Info Hash</span><input class="form-control" name="info_hash" :value="queryText('InfoHash')" placeholder="BTIH"></label>
                      <label class="next02-sh-filter-field next02-sh-page-size-field"><span>每页</span><input class="form-control" type="number" step="1" name="ps" :value="queryText('PageSize') || '50'" inputmode="numeric"></label>
                      <div class="next02-sh-filter-cluster next02-sh-filter-dates" aria-label="发帖日期范围">
                        <span class="next02-sh-filter-cluster-label">发帖日期</span>
                        <label class="next02-sh-filter-field smart-pick-date-field"><span>开始</span><input class="form-control smart-pick-date-input" type="text" name="post_time_from" :value="queryText('PostTimeFrom')" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"></label>
                        <span class="next02-sh-filter-range-separator" aria-hidden="true">—</span>
                        <label class="next02-sh-filter-field smart-pick-date-field"><span>结束</span><input class="form-control smart-pick-date-input" type="text" name="post_time_to" :value="queryText('PostTimeTo')" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"></label>
                      </div>
                      <div class="next02-sh-filter-actions bucket-list-filter-actions"><button class="sc-event-filter-submit bucket-list-apply" type="submit">筛选</button><a class="sc-event-filter-clear bucket-list-clear" href="/shmagnetlist">清空</a></div>
                    </div>
                  </div>
                  <div class="next02-sh-filter-toolbar-bottom">
                    <div class="sh-bucket-toolbar-filter-area next02-sh-filter-secondary">
                      <div class="next02-sh-filter-cluster next02-sh-filter-choice">
                        <span class="next02-sh-filter-cluster-label">影片类型</span>
                        <div class="next02-sh-filter-choice-buttons">
                          <a class="next02-sh-filter-choice-all" :class="{ active: !queryText('MovieType') }" :href="filterHref({ movie_type: '', p: 1 })">全部</a>
                          <a class="next02-sh-filter-choice-button" :class="{ active: queryText('MovieType') === 'jav' }" :href="filterHref({ movie_type: 'jav', p: 1 })">JAV</a>
                          <a class="next02-sh-filter-choice-button" :class="{ active: queryText('MovieType') === 'fc2ppv' }" :href="filterHref({ movie_type: 'fc2ppv', p: 1 })">FC2PPV</a>
                        </div>
                      </div>
                      <div class="next02-sh-filter-cluster next02-sh-filter-choice">
                        <span class="next02-sh-filter-cluster-label">下载状态</span>
                        <div class="next02-sh-filter-choice-buttons">
                          <a class="next02-sh-filter-choice-all" :class="{ active: !queryText('DownloadState') }" :href="filterHref({ download_state: '', p: 1 })">全部</a>
                          <a class="next02-sh-filter-choice-button" :class="{ active: queryText('DownloadState') === '0' }" :href="filterHref({ download_state: '0', p: 1 })">未下载</a>
                          <a class="next02-sh-filter-choice-button" :class="{ active: queryText('DownloadState') === '1' }" :href="filterHref({ download_state: '1', p: 1 })">已下载</a>
                          <a class="next02-sh-filter-choice-button" :class="{ active: queryText('DownloadState') === '2' }" :href="filterHref({ download_state: '2', p: 1 })">已下载（已移除）</a>
                        </div>
                      </div>
                      <div class="next02-sh-filter-cluster next02-sh-filter-choice">
                        <span class="next02-sh-filter-cluster-label">来源版块</span>
                        <div class="next02-sh-filter-choice-buttons">
                          <a class="next02-sh-filter-choice-all" :class="{ active: !queryText('SourceForumID') }" :href="filterHref({ source_forum_id: '', p: 1 })">全部</a>
                          <a
                            v-for="forum in forums"
                            :key="forumText(forum, 'forum_id', 'ForumID')"
                            class="next02-sh-filter-choice-button"
                            :class="{ active: queryText('SourceForumID') === forumText(forum, 'forum_id', 'ForumID') }"
                            :href="filterHref({ source_forum_id: forumText(forum, 'forum_id', 'ForumID'), p: 1 })"
                          >{{ forumText(forum, "forum_name", "ForumName") }}</a>
                        </div>
                      </div>
                    </div>
                    <div class="bucket-list-control-meta sh-bucket-toolbar-status next02-sh-filter-toolbar-status">
                      <div class="sh-bucket-toolbar-metrics next02-sh-toolbar-metrics" aria-label="色花堂磁力摘要">
                        <span :title="`总数 ${total}`" :aria-label="`总数 ${total}`"><strong>{{ total }}</strong></span>
                        <span :title="`当前页 ${rows.length}`" :aria-label="`当前页 ${rows.length}`"><strong>{{ rows.length }}</strong></span>
                      </div>
                      <PageLinks v-if="pageLinks.length" :links="pageLinks" />
                    </div>
                  </div>
                </div>
              </form>
            </div>

            <div class="sh-magnet-selection-bar" aria-label="SH 磁力批量操作">
              <div class="sh-magnet-selection-count">已选 <strong>{{ checkedCount }}</strong> 项</div>
              <button type="button" class="btn btn-sm btn-primary" @click="showBatchDialog">批量加入相册</button>
            </div>

            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 sc-events-table sh-bucket-table sh-magnet-bucket-table">
                <colgroup>
                  <col class="sh-magnet-bucket-col-select">
                  <col class="sh-magnet-bucket-col-post">
                  <col class="sh-magnet-bucket-col-title">
                  <col class="sh-magnet-bucket-col-movie">
                  <col class="sh-magnet-bucket-col-download">
                  <col class="sh-magnet-bucket-col-album">
                  <col class="sh-magnet-bucket-col-hash">
                  <col class="sh-magnet-bucket-col-updated">
                </colgroup>
                <thead>
                  <tr>
                    <th class="text-center text-nowrap">
                      <input
                        class="form-check-input"
                        type="checkbox"
                        aria-label="全选当前页"
                        :checked="currentPageFullyChecked"
                        :indeterminate="currentPagePartiallyChecked"
                        @change="setCurrentPageChecked(($event.target as HTMLInputElement).checked)"
                      >
                    </th>
                    <th class="sc-sortable-head"><a class="sc-sort-link" :class="{ active: queryText('Sort') === 'post_time' }" :href="sortHref('post_time')">发帖时间{{ sortMark('post_time') }}</a></th>
                    <th>帖子标题</th>
                    <th class="sc-sortable-head"><a class="sc-sort-link" :class="{ active: queryText('Sort') === 'movie_name' }" :href="sortHref('movie_name')">影片名{{ sortMark('movie_name') }}</a></th>
                    <th>下载状态</th>
                    <th>相册</th>
                    <th>Info Hash</th>
                    <th class="sc-sortable-head"><a class="sc-sort-link" :class="{ active: queryText('Sort') === 'updated_on' }" :href="sortHref('updated_on')">最近抓到{{ sortMark('updated_on') }}</a></th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(row, index) in rows"
                    :key="rowNumber(row, 'Id')"
                    :class="bucketRowClass(index)"
                  >
                    <td class="text-center">
                      <input
                        class="form-check-input"
                        type="checkbox"
                        :aria-label="`选择 ${rowText(row, 'MovieName') || rowText(row, 'ThreadTitle') || rowNumber(row, 'Id')}`"
                        :checked="rowIsChecked(row)"
                        @change="setRowChecked(row, ($event.target as HTMLInputElement).checked)"
                      >
                    </td>
                    <td class="sc-muted"><ScDateTime :value="rowValue(row, 'PostTime')" unix /></td>
                    <td class="sc-content-cell">
                      <a v-if="rowText(row, 'ThreadURL')" class="fetch-site-title-ellipsis" :href="rowText(row, 'ThreadURL')" target="_blank" rel="noopener noreferrer" :title="rowText(row, 'ThreadTitle')">{{ rowText(row, "ThreadTitle") || rowText(row, "ThreadTID") }}</a>
                      <span v-else class="fetch-site-title-ellipsis" :title="rowText(row, 'ThreadTitle')">{{ rowText(row, "ThreadTitle") || rowText(row, "ThreadTID") || "-" }}</span>
                    </td>
                    <td class="sc-content-cell">
                      <a v-if="movieHref(row)" class="sehuatang-movie-link" :href="movieHref(row)" target="_blank" rel="noopener noreferrer" :title="rowText(row, 'MovieName')">{{ rowText(row, "MovieName") }}</a>
                      <span v-else>{{ rowText(row, "MovieName") || "-" }}</span>
                    </td>
                    <td class="sh-magnet-download-cell">
                      <span class="sh-magnet-download-state" :class="downloadStateClass(row)">{{ downloadStateText(row) }}</span>
                    </td>
                    <td class="sehuatang-album-cell">
                      <div class="fetch-site-album-inline">
                        <button type="button" class="btn btn-sm" :class="rowFlag(row, 'IsInDownload') ? 'btn-warning' : 'btn-outline-warning'" :disabled="albumBusy(row, '下载中') || !rowFlag(row, 'CanFavorite')" @click="switchAlbum(row, '下载中')">{{ rowFlag(row, "IsInDownload") ? "已在下载中" : "下载中" }}</button>
                        <button type="button" class="btn btn-sm" :class="rowFlag(row, 'IsInPending') ? 'btn-info' : 'btn-outline-info'" :disabled="albumBusy(row, '待下载') || !rowFlag(row, 'CanFavorite')" @click="switchAlbum(row, '待下载')">{{ rowFlag(row, "IsInPending") ? "已在待下载" : "待下载" }}</button>
                        <button type="button" class="btn btn-sm btn-outline-secondary" :disabled="!rowFlag(row, 'CanFavorite')" @click="openOtherAlbums(row)">其他相册</button>
                      </div>
                    </td>
                    <td class="sc-content-cell"><span class="fetch-site-hash-text" :title="rowText(row, 'InfoHash')">{{ rowText(row, "InfoHash") || "-" }}</span></td>
                    <td class="sc-muted"><ScDateTime :value="rowValue(row, 'UpdatedOn')" unix /></td>
                  </tr>
                  <tr v-if="!rows.length"><td colspan="8"><div class="sc-empty-state">当前筛选下没有色花堂磁力数据。</div></td></tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>

    <template v-if="otherAlbumRow">
      <div class="modal fade show d-block" tabindex="-1" role="dialog" aria-modal="true" @click.self="closeOtherAlbums">
        <div class="modal-dialog modal-dialog-centered"><div class="modal-content">
          <div class="modal-header"><h5 class="modal-title">选择其他相册</h5><button type="button" class="btn-close" aria-label="关闭" @click="closeOtherAlbums"></button></div>
          <div class="modal-body">
            <div v-if="albums.length" class="fetch-site-other-album-list">
              <button v-for="album in albums" :key="album" type="button" class="btn btn-outline-secondary" :disabled="otherAlbumBusy" @click="addOtherAlbum(album)">{{ album }}</button>
            </div>
            <div v-else class="text-muted small">当前没有可选的其他相册。</div>
          </div>
        </div></div>
      </div>
      <div class="modal-backdrop fade show"></div>
    </template>

    <Teleport to="body">
      <div v-if="batchDialogVisible" class="sh-magnet-batch-overlay" role="presentation" @click.self="hideBatchDialog">
        <section class="sh-magnet-batch-dialog" role="dialog" aria-modal="true" aria-labelledby="shMagnetBatchDialogTitle">
          <header class="sh-magnet-batch-dialog-header">
            <div>
              <span class="sh-magnet-batch-dialog-eyebrow">SH MAGNET / BATCH</span>
              <h5 id="shMagnetBatchDialogTitle">批量加入相册</h5>
            </div>
            <button type="button" class="sh-magnet-batch-dialog-close" aria-label="关闭" :disabled="batchSubmitting" @click="hideBatchDialog">×</button>
          </header>
          <div class="sh-magnet-batch-dialog-body">
            <div class="sh-magnet-batch-dialog-summary">已选择 <strong>{{ checkedCount }}</strong> 条磁力资源</div>
            <label class="form-label" for="shMagnetBatchAlbumSelect">目标相册</label>
            <select id="shMagnetBatchAlbumSelect" v-model="chosenAlbumName" class="form-select" :disabled="batchSubmitting">
              <option v-for="album in availableBatchAlbums" :key="album" :value="album">{{ album }}</option>
            </select>
            <p>已有条目保持不变；新增成功后会从当前选择中移除。</p>
          </div>
          <footer class="sh-magnet-batch-dialog-footer">
            <button type="button" class="btn btn-outline-secondary" :disabled="batchSubmitting" @click="hideBatchDialog">取消</button>
            <button type="button" class="btn btn-primary" :disabled="batchSubmitting || !chosenAlbumName || checkedCount === 0" @click="addCheckedRowsToAlbum">{{ batchSubmitting ? "正在加入..." : "确认加入" }}</button>
          </footer>
        </section>
      </div>
    </Teleport>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/shared/api/request";
import PageShell from "@/shared/components/PageShell.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import type { AnyRecord } from "@/shared/types/gc";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const feedback = ref("");
const feedbackOK = ref(true);
const albumBusyKey = ref("");
const otherAlbumRow = ref<AnyRecord | null>(null);
const otherAlbumBusy = ref(false);
const checkedRowIDs = ref<Set<number>>(new Set());
const batchDialogVisible = ref(false);
const batchSubmitting = ref(false);
const chosenAlbumName = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  feedback.value = "";
  checkedRowIDs.value = new Set();
  batchDialogVisible.value = false;
  chosenAlbumName.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title") || "色花堂磁力列表");
const pageNote = computed(() => text(payload.value, "PageNote"));
const rows = computed(() => asArray<AnyRecord>(pick(payload.value, "Rows", "rows")));
const forums = computed(() => asArray<AnyRecord>(pick(payload.value, "Forums", "forums")));
const albums = computed(() => asArray<string>(pick(payload.value, "Albums", "albums")));
const query = computed(() => asRecord(pick(payload.value, "Query", "query")));
const total = computed(() => Number(pick(payload.value, "Total", "total") || 0));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")));
const pageLinks = computed(() => asArray<AnyRecord>(pick(pageInfo.value, "Links", "links")));
const checkedCount = computed(() => checkedRowIDs.value.size);
const currentPageFullyChecked = computed(() => rows.value.length > 0 && rows.value.every((row) => checkedRowIDs.value.has(rowNumber(row, "Id"))));
const currentPagePartiallyChecked = computed(() => checkedCount.value > 0 && !currentPageFullyChecked.value);
const availableBatchAlbums = computed(() => Array.from(new Set(["下载中", "待下载", ...albums.value].filter(Boolean))));

function rowValue(row: AnyRecord, key: string): unknown { return pick(row, key); }
function rowText(row: AnyRecord, key: string): string { return text(row, key); }
function rowNumber(row: AnyRecord, key: string): number { return Number(pick(row, key) || 0); }
function rowFlag(row: AnyRecord, key: string): boolean { return Boolean(pick(row, key)); }
function forumText(forum: AnyRecord, ...keys: string[]): string { return text(forum, ...keys); }
function queryText(key: string): string { return text(query.value, key); }

function rowIsChecked(row: AnyRecord): boolean {
  return checkedRowIDs.value.has(rowNumber(row, "Id"));
}

function setRowChecked(row: AnyRecord, checked: boolean) {
  const id = rowNumber(row, "Id");
  if (id <= 0) return;
  const next = new Set(checkedRowIDs.value);
  if (checked) next.add(id);
  else next.delete(id);
  checkedRowIDs.value = next;
}

function setCurrentPageChecked(checked: boolean) {
  checkedRowIDs.value = checked
    ? new Set(rows.value.map((row) => rowNumber(row, "Id")).filter((id) => id > 0))
    : new Set();
}

function currentParams(): URLSearchParams {
  const params = new URLSearchParams();
  const values: Array<[string, string]> = [
    ["p", queryText("Page")], ["ps", queryText("PageSize")], ["sort", queryText("Sort")], ["order", queryText("Order")],
    ["keyword", queryText("Keyword")], ["movie_jav_id", queryText("MovieJavID")], ["info_hash", queryText("InfoHash")],
    ["movie_type", queryText("MovieType")], ["download_state", queryText("DownloadState")], ["source_forum_id", queryText("SourceForumID")],
    ["post_time_from", queryText("PostTimeFrom")], ["post_time_to", queryText("PostTimeTo")]
  ];
  for (const [key, value] of values) {
    if (value && !(key === "p" && value === "1") && !(key === "ps" && value === "50")) params.set(key, value);
  }
  return params;
}

function filterHref(overrides: Record<string, string | number>): string {
  const params = currentParams();
  for (const [key, raw] of Object.entries(overrides)) {
    const value = String(raw);
    if (!value || (key === "p" && value === "1")) params.delete(key);
    else params.set(key, value);
  }
  const encoded = params.toString();
  return `/shmagnetlist${encoded ? `?${encoded}` : ""}`;
}

function pageHref(item: AnyRecord): string {
  return filterHref({ p: Number(pick(item, "Page", "page") || 1) });
}

function sortHref(field: string): string {
  const active = queryText("Sort") === field;
  const nextOrder = active && queryText("Order") === "desc" ? "asc" : "desc";
  return filterHref({ sort: field, order: nextOrder, p: 1 });
}

function sortMark(field: string): string {
  if (queryText("Sort") !== field) return "";
  return queryText("Order") === "asc" ? " ↑" : " ↓";
}

function bucketDayKey(index: number): string {
  const timestamp = Number(rowValue(rows.value[index], "PostTime") || 0);
  if (!Number.isFinite(timestamp) || timestamp <= 0) return "";
  const date = new Date(timestamp * 1000);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (part: number) => String(part).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

function bucketRowClass(index: number): Record<string, boolean> {
  const day = bucketDayKey(index);
  if (!day) return { "sehuatang-day-neutral": true };

  let groupIndex = 0;
  for (let rowIndex = 1; rowIndex <= index; rowIndex += 1) {
    const currentDay = bucketDayKey(rowIndex);
    const previousDay = bucketDayKey(rowIndex - 1);
    if (currentDay && previousDay && currentDay !== previousDay) {
      groupIndex += 1;
    }
  }
  const previousDay = index > 0 ? bucketDayKey(index - 1) : "";
  return {
    [`sehuatang-day-group-${(groupIndex % 2) + 1}`]: true,
    "sehuatang-day-start": Boolean(previousDay && day !== previousDay)
  };
}

function movieHref(row: AnyRecord): string {
  const movieName = rowText(row, "MovieName");
  if (!movieName) return "";
  if (rowText(row, "MovieType") === "fc2ppv") {
    const fc2Key = rowText(row, "FC2Key");
    return fc2Key ? `/fcdetail/${encodeURIComponent(fc2Key)}` : "";
  }
  return `/moviedetail/${encodeURIComponent(movieName)}`;
}

function downloadState(row: AnyRecord): number {
  return rowNumber(row, "DownloadState");
}

function downloadStateText(row: AnyRecord): string {
  if (downloadState(row) === 1) return "已下载";
  if (downloadState(row) === 2) return "已下载（已移除）";
  return "未下载";
}

function downloadStateClass(row: AnyRecord): string {
  if (downloadState(row) === 1) return "is-downloaded";
  if (downloadState(row) === 2) return "is-removed";
  return "is-missing";
}

function albumKey(row: AnyRecord, albumName: string): string { return `${rowNumber(row, "Id")}:${albumName}`; }
function albumBusy(row: AnyRecord, albumName: string): boolean { return albumBusyKey.value === albumKey(row, albumName); }
function albumFlag(albumName: string): string { return albumName === "待下载" ? "IsInPending" : "IsInDownload"; }

async function switchAlbum(row: AnyRecord, albumName: string) {
  const routeKey = rowText(row, "MovieRouteKey") || rowText(row, "InfoHash");
  const sourceRowID = rowNumber(row, "Id");
  if (!routeKey || sourceRowID <= 0) return;
  const flag = albumFlag(albumName);
  const active = rowFlag(row, flag);
  albumBusyKey.value = albumKey(row, albumName);
  feedback.value = "";
  try {
    const result = asRecord(await sendJSON(`/api/movie/${encodeURIComponent(routeKey)}/album-item`, active ? "DELETE" : "POST", {
      movie_jav_id: rowText(row, "MovieJavId"), source_type: "sh_magnet", source_row_id: sourceRowID, album_name: albumName
    }));
    row[flag] = Boolean(pick(result, "favorited"));
    feedbackOK.value = true;
    feedback.value = text(result, "message") || `${albumName}操作完成`;
  } catch (err) {
    feedbackOK.value = false;
    feedback.value = err instanceof Error ? err.message : String(err);
  } finally {
    albumBusyKey.value = "";
  }
}

function openOtherAlbums(row: AnyRecord) { otherAlbumRow.value = row; }
function closeOtherAlbums() { if (!otherAlbumBusy.value) otherAlbumRow.value = null; }

async function addOtherAlbum(albumName: string) {
  const row = otherAlbumRow.value;
  if (!row) return;
  const routeKey = rowText(row, "MovieRouteKey") || rowText(row, "InfoHash");
  const sourceRowID = rowNumber(row, "Id");
  if (!routeKey || sourceRowID <= 0) return;
  otherAlbumBusy.value = true;
  feedback.value = "";
  try {
    const result = asRecord(await sendJSON(`/api/movie/${encodeURIComponent(routeKey)}/album-item`, "POST", {
      movie_jav_id: rowText(row, "MovieJavId"), source_type: "sh_magnet", source_row_id: sourceRowID, album_name: albumName
    }));
    feedbackOK.value = true;
    feedback.value = text(result, "message") || `已加入相册：${albumName}`;
    otherAlbumRow.value = null;
  } catch (err) {
    feedbackOK.value = false;
    feedback.value = err instanceof Error ? err.message : String(err);
  } finally {
    otherAlbumBusy.value = false;
  }
}

function showBatchDialog() {
  if (checkedCount.value === 0) {
    feedbackOK.value = false;
    feedback.value = "请先勾选要加入相册的条目";
    return;
  }
  feedback.value = "";
  chosenAlbumName.value = availableBatchAlbums.value[0] || "";
  batchDialogVisible.value = true;
}

function hideBatchDialog() {
  if (batchSubmitting.value) return;
  batchDialogVisible.value = false;
}

async function addCheckedRowsToAlbum() {
  const albumName = chosenAlbumName.value.trim();
  const sourceRowIDs = [...checkedRowIDs.value];
  if (!albumName || !sourceRowIDs.length) return;
  batchSubmitting.value = true;
  feedback.value = "";
  try {
    const result = asRecord(await sendJSON("/api/sh-magnets/batch-add-album", "POST", {
      album_name: albumName,
      source_row_ids: sourceRowIDs
    }));
    const addedIDs = asArray<number>(pick(result, "added_ids")).map(Number);
    const existingIDs = asArray<number>(pick(result, "existing_ids")).map(Number);
    const successfulIDs = new Set([...addedIDs, ...existingIDs].filter((id) => id > 0));
    if (albumName === "下载中" || albumName === "待下载") {
      const flag = albumFlag(albumName);
      for (const row of rows.value) {
        if (successfulIDs.has(rowNumber(row, "Id"))) row[flag] = true;
      }
    }
    const nextChecked = new Set(checkedRowIDs.value);
    for (const id of successfulIDs) nextChecked.delete(id);
    checkedRowIDs.value = nextChecked;
    const addedCount = Number(pick(result, "added_count") || addedIDs.length);
    const existingCount = Number(pick(result, "existing_count") || existingIDs.length);
    const failedCount = Number(pick(result, "failed_count") || 0);
    feedbackOK.value = failedCount === 0;
    feedback.value = failedCount > 0
      ? `已新增 ${addedCount} 项，已有 ${existingCount} 项，失败 ${failedCount} 项`
      : `已批量加入相册“${albumName}”：新增 ${addedCount} 项，已有 ${existingCount} 项`;
    batchDialogVisible.value = false;
  } catch (err) {
    feedbackOK.value = false;
    feedback.value = err instanceof Error ? err.message : "批量加入相册失败";
  } finally {
    batchSubmitting.value = false;
  }
}

const PageLinks = defineComponent({
  props: { links: { type: Array as PropType<AnyRecord[]>, required: true } },
  setup(props) {
    return () => h("nav", { "aria-label": "色花堂磁力分页" }, [
      h("ul", { class: "pagination pagination-sm mb-0" }, props.links.map((item, index) => {
        if (item.Ellipsis || item.ellipsis) return h("li", { class: "page-item disabled", key: `ellipsis-${index}` }, [h("span", { class: "page-link" }, "...")]);
        const disabled = Boolean(item.Disabled || item.disabled);
        return h("li", { class: ["page-item", { active: Boolean(item.Active || item.active), disabled }], key: `${text(item, "Label", "label")}-${index}` }, [
          h("a", { class: "page-link", href: disabled ? "#" : pageHref(item), "aria-disabled": disabled ? "true" : undefined }, text(item, "Label", "label"))
        ]);
      }))
    ]);
  }
});
</script>
