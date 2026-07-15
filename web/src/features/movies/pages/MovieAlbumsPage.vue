<template>
  <PageShell class="torrent-albums-page-shell sc-events-page-shell movie-albums-page-shell" title="MovieAlbums" note="" eyebrow="">
    <div id="page-root" class="container-fluid torrent-albums-page-root movie-albums-page-root albums-list-standard movie-albums-list-standard next02-sc-pink-theme next02-sc-events-rewrite">
      <div v-if="loading" class="alert albums-list-feedback is-loading">加载中...</div>
      <div v-else-if="error" class="alert albums-list-feedback is-error">{{ error }}</div>
      <div
        v-else
        id="movieAlbumItemsPage"
        class="torrent-albums-page movie-albums-page"
        :data-album-id="selectedAlbumID"
        :data-album-name="selectedAlbumName"
      >
        <header class="albums-list-hero">
          <div class="albums-list-hero-copy">
            <span class="albums-list-eyebrow">Movie Albums</span>
            <h2>{{ pageTitle }}</h2>
            <p v-if="pageNote">{{ pageNote }}</p>
          </div>
          <div class="albums-list-hero-actions">
            <button type="button" class="btn albums-manage-button" id="btnOpenManageMovieAlbumsModal" @click="activeModal = 'manage'">管理相册</button>
            <button type="button" class="btn albums-create-button" id="btnOpenCreateMovieAlbumModal" @click="openCreateModal">新增电影相册</button>
          </div>
        </header>

        <div id="movieAlbumItemsMsg" class="alert albums-list-message" :class="messageClass" role="alert">
          {{ messageText }}
        </div>

        <section class="albums-list-panel next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
          <div class="albums-list-controls">
            <form ref="filterForm" class="albums-filter-form movie-albums-filter-form" id="movieAlbumsFilterForm" method="get" action="/moviealbumlist">
              <input v-if="queryValue('PageSize')" type="hidden" name="ps" :value="queryValue('PageSize')">
              <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')">
              <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')">

              <div class="albums-filter-row is-primary">
                <div class="albums-filter-field is-quick-filter">
                  <span>快捷相册</span>
                  <div class="album-quick-filter-group">
                  <button
                    type="button"
                    class="btn album-quick-button js-movie-album-quick-filter"
                    :class="{ 'is-active': selectedAlbumName === '待下载电影' }"
                    data-album-name="待下载电影"
                    @click="quickFilter('待下载电影')"
                  >待下载电影</button>
                  <button
                    type="button"
                    class="btn album-quick-button js-movie-album-quick-filter"
                    :class="{ 'is-active': selectedAlbumName === '待删除' }"
                    data-album-name="待删除"
                    @click="quickFilter('待删除')"
                  >待删除</button>
                  <button
                    type="button"
                    class="btn album-quick-button js-movie-album-quick-filter"
                    :class="{ 'is-active': selectedAlbumName === 'v3' }"
                    data-album-name="v3"
                    @click="quickFilter('v3')"
                  >v3</button>
                  <button
                    type="button"
                    class="btn album-quick-button js-movie-album-quick-filter"
                    :class="{ 'is-active': selectedAlbumName === 'vx' }"
                    data-album-name="vx"
                    @click="quickFilter('vx')"
                  >vx</button>
                  </div>
                </div>
                <label class="albums-filter-field is-album">
                  <span>相册</span>
                <select class="form-select" id="movieAlbumsAlbumName" name="album_name" :value="selectedAlbumName">
                  <option v-for="option in albumSelectOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
                </label>
                <label class="albums-filter-field is-keyword">
                  <span>影片编号 / 标题</span>
                <input class="form-control" type="text" name="keyword" :value="queryValue('Keyword')" placeholder="标题、JAV ID 或 FC2 Key">
                </label>
                <div class="albums-filter-field movie-albums-release-range">
                  <span>发布日期</span>
                  <div class="movie-albums-release-range-inputs">
                    <input class="form-control smart-pick-date-input" type="text" name="releasing_from" aria-label="发布开始日期" :value="queryValue('ReleasingFrom')" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
                    <span class="movie-albums-release-range-separator" aria-hidden="true">—</span>
                    <input class="form-control smart-pick-date-input" type="text" name="releasing_to" aria-label="发布结束日期" :value="queryValue('ReleasingTo')" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
                  </div>
                </div>
                <div class="albums-filter-actions">
                  <button class="btn albums-apply-button" type="submit">筛选</button>
                  <a class="btn albums-clear-button" href="/moviealbumlist">清空</a>
                </div>
              </div>
            </form>

            <div class="albums-list-meta">
              <div class="albums-list-summary">
                <span class="albums-list-summary-label">相册</span>
                <span class="albums-list-summary-value is-album-name">{{ selectedAlbumName }}</span>
                <span class="albums-list-summary-value">{{ total }}</span>
                <span class="albums-list-summary-label">Total Items</span>
              </div>
              <MovieAlbumPagination v-if="hasRows && pageLinks.length" :links="pageLinks" />
            </div>
          </div>

          <div class="albums-batch-toolbar">
            <div class="albums-selected-count">已选 <span id="movieAlbumBatchSelectedCount">{{ selectedCount }}</span> 项</div>
            <div class="albums-batch-actions">
              <button
                type="button"
                class="btn albums-batch-button"
                id="btnBatchCopyMovieAlbumItems"
                :disabled="selectedCount === 0 || selectedAlbumID <= 0 || !transferTargetOptions.length"
                @click="openTransferModal('copy')"
              >批量复制</button>
              <button
                type="button"
                class="btn albums-batch-button"
                id="btnBatchMoveMovieAlbumItems"
                :disabled="selectedCount === 0 || selectedAlbumID <= 0 || !transferTargetOptions.length"
                @click="openTransferModal('move')"
              >批量移动</button>
              <button
                v-if="canExecuteRemove"
                type="button"
                class="btn albums-batch-button is-danger"
                id="btnBatchExecuteMovieAlbumRemove"
                :disabled="selectedCount === 0 || selectedAlbumID <= 0"
                @click="openExecuteModal"
              >统一删除</button>
              <button
                v-if="canRemoveDownloaded"
                type="button"
                class="btn albums-batch-button"
                id="btnRemoveDownloadedMovieAlbumItems"
                :disabled="loadingDownloaded"
                @click="loadRemoveDownloadedPreview"
              >移除已下载条目</button>
            </div>
          </div>

          <div class="table-responsive albums-table-wrap">
              <table class="table table-hover align-middle mb-0 sc-events-table movie-album-table">
                <colgroup>
                  <col class="movie-album-col-select">
                  <col class="movie-album-col-movie">
                  <col class="movie-album-col-releasing">
                  <col class="movie-album-col-created">
                  <col class="movie-album-col-actions">
                </colgroup>
                <thead>
                  <tr>
                    <th class="text-center text-nowrap">
                      <input
                        class="form-check-input js-movie-album-item-check-all"
                        id="movieAlbumItemCheckAll"
                        type="checkbox"
                        aria-label="全选当前页"
                        :checked="allSelected"
                        :indeterminate="partialSelected"
                        @change="toggleAll(($event.target as HTMLInputElement).checked)"
                      >
                    </th>
                    <th class="sc-sortable-head"><SortLink :link="sortLink('ByMovieName')" label="影片" /></th>
                    <th class="sc-sortable-head sc-number"><SortLink :link="sortLink('ByReleasingDate')" label="发布日期" /></th>
                    <th class="sc-sortable-head"><SortLink :link="sortLink('ByCreated')" label="加入时间" /></th>
                    <th class="text-nowrap">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td colspan="5"><div class="sc-empty-state">当前筛选下没有电影相册数据。</div></td>
                  </tr>
                  <tr v-for="row in rows" v-else :key="rowKey(row)" :data-item-id="value(row, 'ID')">
                    <td class="text-center">
                      <input
                        class="form-check-input js-movie-album-item-check"
                        type="checkbox"
                        :value="value(row, 'ID')"
                        :aria-label="`选择 ${value(row, 'MovieName')}`"
                        :checked="isSelected(row)"
                        @change="toggleRow(row, ($event.target as HTMLInputElement).checked)"
                      >
                    </td>
                    <td class="sc-content-cell text-nowrap" :title="value(row, 'MovieName')">
                      <a class="album-entry-link" :href="movieHref(row)" target="_blank" rel="noopener noreferrer">{{ value(row, "MovieName") }}</a>
                    </td>
                    <td class="sc-number text-nowrap">{{ formatUnixDate(value(row, "ReleasingDate")) }}</td>
                    <td class="sc-number text-nowrap"><ScDateTime :value="value(row, 'CreatedOn')" unix /></td>
                    <td class="text-nowrap">
                      <button
                        type="button"
                        class="btn btn-sm btn-outline-danger js-remove-movie-album-item"
                        :data-item-id="value(row, 'ID')"
                        @click="removeOne(row)"
                      >移除</button>
                    </td>
                  </tr>
                </tbody>
              </table>
          </div>
        </section>

        <MovieAlbumModals
          :active-modal="activeModal"
          :selected-count="selectedCount"
          :create-name="createName"
          :downloaded-items="downloadedPreviewItems"
          :loading-downloaded="loadingDownloaded"
          :transfer-mode="transferMode"
          :transfer-target-options="transferTargetOptions"
          :album-options="manageAlbumOptions"
          :selected-album-id="selectedAlbumID"
          @close="closeModal"
          @update-create-name="createName = $event"
          @confirm-create="createMovieAlbum"
          @confirm-execute="executeRemove"
          @confirm-remove-downloaded="removeDownloadedItems"
          @confirm-transfer="transferMovieAlbumItems"
          @rename-album="renameManagedMovieAlbum"
          @delete-album="deleteManagedMovieAlbum"
        />
      </div>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/shared/api/request";
import PageShell from "@/shared/components/PageShell.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import type { AnyRecord, PageInfo } from "@/shared/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const rows = ref<AnyRecord[]>([]);
const selectedIDs = ref<Set<number>>(new Set());
const loading = ref(false);
const error = ref("");
const messageText = ref("");
const messageOk = ref(true);
const activeModal = ref<"" | "create" | "execute" | "downloaded" | "transfer" | "manage">("");
const transferMode = ref<"move" | "copy">("copy");
const createName = ref("");
const downloadedPreviewItems = ref<AnyRecord[]>([]);
const loadingDownloaded = ref(false);
const filterForm = ref<HTMLFormElement | null>(null);

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  messageText.value = "";
  selectedIDs.value = new Set();
  try {
    const data = asRecord(await getJSON(apiPathForRoute(route)));
    payload.value = data;
    rows.value = asArray<AnyRecord>(pick(data, "Rows", "rows"));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title"));
const pageNote = computed(() => text(payload.value, "PageNote"));
const query = computed(() => asRecord(pick(payload.value, "Query")));
const albums = computed(() => asArray<AnyRecord>(pick(payload.value, "Albums")));
const total = computed(() => text(payload.value, "Total", "total") || "0");
const selectedAlbumName = computed(() => text(payload.value, "SelectedAlbumName") || queryValue("AlbumName"));
const selectedAlbumID = computed(() => Number(text(payload.value, "SelectedAlbumID") || 0));
const hasRows = computed(() => rows.value.length > 0);
const sortQuery = computed(() => asRecord(pick(payload.value, "SortQuery")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<AnyRecord>(pageInfo.value.Links));
const selectedCount = computed(() => selectedIDs.value.size);
const allSelected = computed(() => rows.value.length > 0 && selectedIDs.value.size === rows.value.length);
const partialSelected = computed(() => selectedIDs.value.size > 0 && selectedIDs.value.size < rows.value.length);
const messageClass = computed(() => messageText.value ? (messageOk.value ? "alert-success small py-2 px-3" : "alert-danger small py-2 px-3") : "d-none");
const canExecuteRemove = computed(() => selectedAlbumName.value === "待删除");
const canRemoveDownloaded = computed(() => ["待下载电影", "v3", "vx"].includes(selectedAlbumName.value));
const albumSelectOptions = computed(() => {
  const options = albums.value.map((album) => ({ value: value(album, "Name"), label: value(album, "Name") || value(album, "ID") }));
  if (!options.some((item) => String(item.value) === String(selectedAlbumName.value || ""))) {
    options.unshift({ value: selectedAlbumName.value || "", label: selectedAlbumName.value || "默认" });
  }
  return options;
});
const transferTargetOptions = computed(() => albums.value
  .filter((album) => Number(value(album, "ID")) > 0 && Number(value(album, "ID")) !== selectedAlbumID.value)
  .map((album) => ({ value: value(album, "ID"), label: value(album, "Name") || value(album, "ID") })));
const manageAlbumOptions = computed(() => albums.value
  .filter((album) => Number(value(album, "ID")) > 0)
  .map((album) => ({
    value: value(album, "ID"),
    label: value(album, "Name") || value(album, "ID"),
    protected: pick(album, "Protected") === true
  })));

function queryValue(key: string): string {
  return text(query.value, key);
}

function value(row: AnyRecord, key: string): string {
  return text(row, key);
}

function rowKey(row: AnyRecord): string {
  return value(row, "ID");
}

function sortLink(key: string): AnyRecord {
  return asRecord(sortQuery.value[key]);
}

function movieHref(row: AnyRecord): string {
  if (movieType(row) === "fc2ppv") {
    return `/fcdetail/${encodeURIComponent(value(row, "FC2Key"))}`;
  }
  return `/moviedetail/${encodeURIComponent(value(row, "MovieName") || value(row, "MovieJavID"))}`;
}

function movieType(row: AnyRecord): "jav" | "fc2ppv" {
  return value(row, "MovieType") === "fc2ppv" || Boolean(value(row, "FC2Key")) ? "fc2ppv" : "jav";
}

function formatUnixDate(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function rowID(row: AnyRecord): number {
  return Number(value(row, "ID") || 0);
}

function isSelected(row: AnyRecord): boolean {
  return selectedIDs.value.has(rowID(row));
}

function toggleRow(row: AnyRecord, checked: boolean) {
  const next = new Set(selectedIDs.value);
  const id = rowID(row);
  if (checked) next.add(id);
  else next.delete(id);
  selectedIDs.value = next;
}

function toggleAll(checked: boolean) {
  selectedIDs.value = checked ? new Set(rows.value.map(rowID).filter((id) => id > 0)) : new Set();
}

function showMessage(message: string, ok = true) {
  messageText.value = message;
  messageOk.value = ok;
}

function quickFilter(albumName: string) {
  if (filterForm.value) {
    const select = filterForm.value.querySelector("#movieAlbumsAlbumName") as HTMLSelectElement | null;
    if (select) {
      select.value = albumName;
      filterForm.value.submit();
      return;
    }
  }
  const url = new URL("/moviealbumlist", window.location.origin);
  url.searchParams.set("album_name", albumName);
  window.location.href = `${url.pathname}${url.search}`;
}

function removeRowsByIDs(ids: number[]) {
  const idSet = new Set(ids);
  rows.value = rows.value.filter((row) => !idSet.has(rowID(row)));
  selectedIDs.value = new Set([...selectedIDs.value].filter((id) => !idSet.has(id)));
}

function closeModal() {
  activeModal.value = "";
}

function openCreateModal() {
  createName.value = "";
  activeModal.value = "create";
  window.setTimeout(() => document.getElementById("movieAlbumCreateName")?.focus(), 50);
}

function openExecuteModal() {
  if (!selectedIDs.value.size) {
    showMessage("请先选择待删除条目", false);
    return;
  }
  activeModal.value = "execute";
}

function openTransferModal(mode: "move" | "copy") {
  if (!selectedIDs.value.size) {
    showMessage("请先选择电影相册条目", false);
    return;
  }
  if (!transferTargetOptions.value.length) {
    showMessage("没有可用的目标电影相册", false);
    return;
  }
  transferMode.value = mode;
  activeModal.value = "transfer";
}

async function transferMovieAlbumItems(targetAlbumID: string) {
  const targetID = Number(targetAlbumID || 0);
  if (targetID <= 0) {
    showMessage("请选择目标电影相册", false);
    return;
  }
  const ids = [...selectedIDs.value];
  if (!ids.length) {
    showMessage("请先选择电影相册条目", false);
    return;
  }

  const mode = transferMode.value;
  closeModal();
  try {
    const data = asRecord(await sendJSON(`/api/movie-albums/${selectedAlbumID.value}/items/batch-${mode}`, "POST", {
      item_ids: ids,
      target_album_id: targetID
    }));
    const failedIDs = new Set(asArray<number>(pick(data, "failed_ids", "FailedIDs")).map(Number).filter((id) => id > 0));
    const processedCount = Number(pick(data, "processed_count") || 0);
    const skippedCount = Number(pick(data, "skipped_count") || 0);
    if (mode === "move") {
      await load();
    } else {
      selectedIDs.value = new Set();
    }
    const action = mode === "move" ? "移动" : "复制";
    const details = [`已${action} ${processedCount} 条`];
    if (skippedCount > 0) details.push(`跳过重复 ${skippedCount} 条`);
    if (failedIDs.size > 0) details.push(`失败 ${failedIDs.size} 条`);
    showMessage(details.join("，"), failedIDs.size === 0);
  } catch (err) {
    showMessage(err instanceof Error ? err.message : `批量${mode === "move" ? "移动" : "复制"}失败`, false);
  }
}

async function createMovieAlbum() {
  const name = createName.value.trim();
  if (!name) {
    showMessage("请输入电影相册名称", false);
    return;
  }
  try {
    const data = asRecord(await sendJSON("/api/movie-albums", "POST", { name }));
    closeModal();
    showMessage(text(data, "message") || "电影相册创建成功");
    const albumName = text(data, "album_name") || name;
    const url = new URL("/moviealbumlist", window.location.origin);
    url.searchParams.set("album_name", albumName);
    window.location.href = `${url.pathname}${url.search}`;
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "电影相册创建失败", false);
  }
}

async function renameManagedMovieAlbum(input: { id: string; name: string }) {
  const albumID = Number(input.id || 0);
  const name = String(input.name || "").trim();
  if (albumID <= 0 || !name) {
    showMessage("电影相册名称不能为空", false);
    return;
  }
  try {
    const data = asRecord(await sendJSON(`/api/movie-albums/${albumID}`, "PUT", { name }));
    closeModal();
    const updatedName = text(data, "album_name") || name;
    if (albumID === selectedAlbumID.value) {
      const url = new URL("/moviealbumlist", window.location.origin);
      url.searchParams.set("album_name", updatedName);
      window.location.href = `${url.pathname}${url.search}`;
      return;
    }
    window.location.reload();
  } catch (err) {
    closeModal();
    showMessage(err instanceof Error ? err.message : "编辑电影相册失败", false);
  }
}

async function deleteManagedMovieAlbum(albumIDValue: string) {
  const albumID = Number(albumIDValue || 0);
  if (albumID <= 0) {
    showMessage("电影相册 ID 无效", false);
    return;
  }
  try {
    await sendJSON(`/api/movie-albums/${albumID}`, "DELETE");
    closeModal();
    if (albumID === selectedAlbumID.value) {
      window.location.href = "/moviealbumlist";
      return;
    }
    window.location.reload();
  } catch (err) {
    closeModal();
    showMessage(err instanceof Error ? err.message : "删除电影相册失败", false);
  }
}

async function removeOne(row: AnyRecord) {
  const id = rowID(row);
  if (selectedAlbumID.value <= 0) {
    showMessage("当前电影相册无效", false);
    return;
  }
  if (id <= 0) {
    showMessage("缺少相册条目", false);
    return;
  }
  try {
    const data = asRecord(await sendJSON(`/api/movie-albums/${selectedAlbumID.value}/items/remove`, "POST", { item_id: id }));
    if (data.removed !== false) removeRowsByIDs([id]);
    showMessage(text(data, "message") || "移除完成");
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "移除失败", false);
  }
}

async function executeRemove() {
  const ids = [...selectedIDs.value];
  if (!ids.length) {
    showMessage("请先选择待删除条目", false);
    return;
  }
  try {
    const data = asRecord(await sendJSON(`/api/movie-albums/${selectedAlbumID.value}/execute-remove`, "POST", { item_ids: ids }));
    closeModal();
    const removedCount = Number(pick(data, "removed_count", "RemovedCount") || 0);
    const removedJAVCount = Number(pick(data, "removed_jav_count", "RemovedJAVCount") || 0);
    const removedFC2Count = Number(pick(data, "removed_fc2_count", "RemovedFC2Count") || 0);
    const failedCount = Number(pick(data, "failed_count", "FailedCount") || 0);
    await load();
    const details = [`统一删除完成：成功 ${removedCount} 条（JAV ${removedJAVCount} 条，FC2 ${removedFC2Count} 条）`];
    if (failedCount > 0) details.push(`失败 ${failedCount} 条，失败条目已保留`);
    showMessage(details.join("，"), failedCount === 0);
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "统一删除失败", false);
  }
}

async function loadRemoveDownloadedPreview() {
  if (selectedAlbumID.value <= 0) {
    showMessage("当前电影相册无效", false);
    return;
  }
  loadingDownloaded.value = true;
  downloadedPreviewItems.value = [];
  try {
    const data = asRecord(await sendJSON(`/api/movie-albums/${selectedAlbumID.value}/remove-downloaded-items-preview`, "POST", {}));
    const items = asArray<AnyRecord>(data.items);
    downloadedPreviewItems.value = items;
    if (!items.length) {
      showMessage("没有可移除的已下载条目");
      return;
    }
    activeModal.value = "downloaded";
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "加载移除预览失败", false);
  } finally {
    loadingDownloaded.value = false;
  }
}

async function removeDownloadedItems() {
  if (selectedAlbumID.value <= 0) {
    showMessage("当前电影相册无效", false);
    return;
  }
  if (!downloadedPreviewItems.value.length) {
    showMessage("没有可移除的已下载条目", false);
    return;
  }
  try {
    const data = asRecord(await sendJSON(`/api/movie-albums/${selectedAlbumID.value}/remove-downloaded-items`, "POST", {}));
    closeModal();
    showMessage(text(data, "message") || "已移除已下载条目");
    window.setTimeout(() => window.location.reload(), 500);
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "移除已下载条目失败", false);
  }
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
        class: ["album-sort-link", "sc-sort-link", active ? "active" : ""],
        href: normalizeHref(text(link, "Href") || "#")
      }, `${props.label}${active ? (desc ? " ↓" : " ↑") : ""}`);
    };
  }
});

const MovieAlbumPagination = defineComponent({
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

const MovieAlbumModals = defineComponent({
  props: {
    activeModal: { type: String, required: true },
    selectedCount: { type: Number, required: true },
    createName: { type: String, required: true },
    downloadedItems: { type: Array as PropType<AnyRecord[]>, required: true },
    loadingDownloaded: { type: Boolean, required: true },
    transferMode: { type: String as PropType<"move" | "copy">, required: true },
    transferTargetOptions: { type: Array as PropType<Array<{ value: string; label: string }>>, required: true },
    albumOptions: { type: Array as PropType<Array<{ value: string; label: string; protected: boolean }>>, required: true },
    selectedAlbumId: { type: Number, required: true }
  },
  emits: ["close", "updateCreateName", "confirmCreate", "confirmExecute", "confirmRemoveDownloaded", "confirmTransfer", "renameAlbum", "deleteAlbum"],
  setup(props, { emit }) {
    const editingAlbumID = ref("");
    const editingAlbumName = ref("");
    const deletingAlbumID = ref("");
    const deletingAlbumName = ref("");
    const modalClass = (name: string) => ["modal", "fade", props.activeModal === name ? "show d-block" : ""];
    const ariaHidden = (name: string) => props.activeModal === name ? "false" : "true";
    const resetManagement = () => {
      editingAlbumID.value = "";
      editingAlbumName.value = "";
      deletingAlbumID.value = "";
      deletingAlbumName.value = "";
    };
    const closeManagement = () => {
      resetManagement();
      emit("close");
    };
    const previewList = () => props.downloadedItems.length
      ? props.downloadedItems.map((item) => {
        const row = asRecord(item);
        const title = text(row, "MovieName") || text(row, "MovieJavID") || text(row, "FC2Key") || "-";
        const identity = text(row, "MovieJavID") || text(row, "FC2Key");
        return h("div", { class: "list-group-item px-0 py-2 border-0 border-bottom" }, [
          h("div", { class: "fw-semibold" }, title),
          identity ? h("div", { class: "small text-muted" }, [h("code", identity)]) : null
        ]);
      })
      : [h("div", { id: "movieAlbumRemoveDownloadedEmpty", class: "small text-muted" }, props.loadingDownloaded ? "正在加载预览..." : "没有可移除的已下载条目。")];
    return () => h("div", [
      h("div", {
        class: modalClass("manage"),
        id: "movieAlbumManageModal",
        tabindex: "-1",
        "aria-labelledby": "movieAlbumManageModalLabel",
        "aria-hidden": ariaHidden("manage")
      }, [h("div", { class: "modal-dialog modal-dialog-centered albums-manage-dialog" }, [h("div", { class: "modal-content" }, [
        h("div", { class: "modal-header" }, [
          h("h5", { class: "modal-title", id: "movieAlbumManageModalLabel" }, "管理电影相册"),
          h("button", { type: "button", class: "btn-close", "aria-label": "Close", onClick: closeManagement })
        ]),
        h("div", { class: "modal-body albums-manage-body" }, deletingAlbumID.value
          ? [
              h("div", { class: "albums-manage-delete-confirm" }, [
                h("h6", "删除电影相册确认"),
                h("p", ["确认删除电影相册“", h("strong", deletingAlbumName.value), "”吗？"]),
                h("p", { class: "text-muted small mb-0" }, "仅空电影相册可以删除；相册中存在任何影片记录时，删除操作会被拒绝。")
              ])
            ]
          : [
              h("p", { class: "text-muted small mb-2" }, "可编辑或删除自定义电影相册；业务相册由现有流程使用。"),
              h("div", { class: "albums-manage-list" }, props.albumOptions.map((option) => {
                const isEditing = editingAlbumID.value === option.value;
                const isSelected = Number(option.value) === props.selectedAlbumId;
                return h("div", { class: ["albums-manage-item", isSelected ? "is-selected" : ""], key: option.value }, [
                  h("div", { class: "albums-manage-item-main" }, isEditing
                    ? [h("input", {
                        class: "form-control albums-manage-name-input",
                        value: editingAlbumName.value,
                        maxlength: "128",
                        "aria-label": `编辑电影相册 ${option.label}`,
                        onInput: (event: Event) => { editingAlbumName.value = (event.target as HTMLInputElement).value; }
                      })]
                    : [
                        h("span", { class: "albums-manage-name" }, option.label),
                        option.protected ? h("span", { class: "albums-manage-system-badge" }, "业务相册") : null,
                        isSelected ? h("span", { class: "albums-manage-current-badge" }, "当前") : null
                      ]),
                  h("div", { class: "albums-manage-item-actions" }, isEditing
                    ? [
                        h("button", {
                          type: "button",
                          class: "btn btn-primary btn-sm",
                          disabled: !editingAlbumName.value.trim(),
                          onClick: () => emit("renameAlbum", { id: option.value, name: editingAlbumName.value })
                        }, "保存"),
                        h("button", {
                          type: "button",
                          class: "btn btn-light btn-sm",
                          onClick: () => { editingAlbumID.value = ""; editingAlbumName.value = ""; }
                        }, "取消")
                      ]
                    : [
                        h("button", {
                          type: "button",
                          class: "btn btn-outline-secondary btn-sm",
                          disabled: option.protected,
                          "aria-label": `编辑电影相册 ${option.label}`,
                          onClick: () => { editingAlbumID.value = option.value; editingAlbumName.value = option.label; }
                        }, "编辑"),
                        h("button", {
                          type: "button",
                          class: "btn btn-outline-danger btn-sm",
                          disabled: option.protected,
                          "aria-label": `删除电影相册 ${option.label}`,
                          onClick: () => { deletingAlbumID.value = option.value; deletingAlbumName.value = option.label; }
                        }, "删除")
                      ])
                ]);
              }))
            ]),
        h("div", { class: "modal-footer" }, [
          deletingAlbumID.value
            ? h("button", { type: "button", class: "btn btn-light", onClick: () => { deletingAlbumID.value = ""; deletingAlbumName.value = ""; } }, "返回管理")
            : null,
          deletingAlbumID.value
            ? h("button", { type: "button", class: "btn btn-danger", id: "movieAlbumDeleteConfirmBtn", onClick: () => emit("deleteAlbum", deletingAlbumID.value) }, "确认删除")
            : h("button", { type: "button", class: "btn btn-light", onClick: closeManagement }, "关闭")
        ])
      ])])]),
      h("div", {
        class: modalClass("create"),
        id: "movieAlbumCreateModal",
        tabindex: "-1",
        "aria-hidden": ariaHidden("create")
      }, [h("div", { class: "modal-dialog modal-dialog-centered" }, [h("div", { class: "modal-content" }, [
        h("div", { class: "modal-header" }, [
          h("h5", { class: "modal-title" }, "新增电影相册"),
          h("button", { type: "button", class: "btn-close", "aria-label": "Close", onClick: () => emit("close") })
        ]),
        h("div", { class: "modal-body" }, [
          h("label", { class: "form-label mb-1", for: "movieAlbumCreateName" }, "相册名称"),
          h("input", {
            type: "text",
            class: "form-control",
            id: "movieAlbumCreateName",
            maxlength: "128",
            placeholder: "请输入电影相册名称",
            value: props.createName,
            onInput: (event: Event) => emit("updateCreateName", (event.target as HTMLInputElement).value),
            onKeydown: (event: KeyboardEvent) => {
              if (event.key === "Enter") {
                event.preventDefault();
                emit("confirmCreate");
              }
            }
          })
        ]),
        h("div", { class: "modal-footer" }, [
          h("button", { type: "button", class: "btn btn-light", onClick: () => emit("close") }, "取消"),
          h("button", { type: "button", class: "btn btn-primary", id: "movieAlbumCreateConfirmBtn", onClick: () => emit("confirmCreate") }, "确认创建")
        ])
      ])])]),
      h("div", {
        class: modalClass("execute"),
        id: "movieAlbumExecuteRemoveModal",
        tabindex: "-1",
        "aria-hidden": ariaHidden("execute")
      }, [h("div", { class: "modal-dialog modal-dialog-centered" }, [h("div", { class: "modal-content" }, [
        h("div", { class: "modal-header" }, [
          h("h5", { class: "modal-title" }, "统一删除确认"),
          h("button", { type: "button", class: "btn-close", "aria-label": "Close", onClick: () => emit("close") })
        ]),
        h("div", { class: "modal-body" }, ["确认执行已选中的 ", h("span", { id: "movieAlbumExecuteRemoveCount" }, String(props.selectedCount)), " 项吗？这会按影片类型将对应的 w_media 或 fc2_media 移动到 `001_process/005_removed`。"]),
        h("div", { class: "modal-footer" }, [
          h("button", { type: "button", class: "btn btn-light", onClick: () => emit("close") }, "取消"),
          h("button", { type: "button", class: "btn btn-danger", id: "movieAlbumExecuteRemoveConfirmBtn", onClick: () => emit("confirmExecute") }, "确认删除")
        ])
      ])])]),
      h("div", {
        class: modalClass("transfer"),
        id: "movieAlbumTransferModal",
        tabindex: "-1",
        "aria-hidden": ariaHidden("transfer")
      }, [h("div", { class: "modal-dialog modal-dialog-centered" }, [h("div", { class: "modal-content" }, [
        h("div", { class: "modal-header" }, [
          h("h5", { class: "modal-title" }, props.transferMode === "move" ? "批量移动电影相册条目" : "批量复制电影相册条目"),
          h("button", { type: "button", class: "btn-close", "aria-label": "Close", onClick: () => emit("close") })
        ]),
        h("div", { class: "modal-body" }, [
          h("label", { class: "form-label mb-1", for: "movieAlbumTransferTarget" }, "目标电影相册"),
          h("select", { class: "form-select", id: "movieAlbumTransferTarget" }, [
            h("option", { value: "" }, "请选择目标电影相册"),
            ...props.transferTargetOptions.map((option) => h("option", { value: option.value }, option.label))
          ]),
          h("p", { class: "small text-muted mt-2 mb-0" }, props.transferMode === "move"
            ? `将已选中的 ${props.selectedCount} 条记录移动到目标相册，当前相册将移除这些条目。`
            : `将已选中的 ${props.selectedCount} 条记录复制到目标相册，当前相册继续保留这些条目。`)
        ]),
        h("div", { class: "modal-footer" }, [
          h("button", { type: "button", class: "btn btn-light", onClick: () => emit("close") }, "取消"),
          h("button", {
            type: "button",
            class: "btn btn-primary",
            id: "movieAlbumTransferConfirmBtn",
            onClick: () => {
              const target = document.getElementById("movieAlbumTransferTarget") as HTMLSelectElement | null;
              emit("confirmTransfer", target?.value || "");
            }
          }, props.transferMode === "move" ? "确认移动" : "确认复制")
        ])
      ])])]),
      h("div", {
        class: modalClass("downloaded"),
        id: "movieAlbumRemoveDownloadedModal",
        tabindex: "-1",
        "aria-hidden": ariaHidden("downloaded")
      }, [h("div", { class: "modal-dialog modal-dialog-centered" }, [h("div", { class: "modal-content" }, [
        h("div", { class: "modal-header" }, [
          h("h5", { class: "modal-title" }, "移除已下载条目确认"),
          h("button", { type: "button", class: "btn-close", "aria-label": "Close", onClick: () => emit("close") })
        ]),
        h("div", { class: "modal-body" }, [
          h("div", { class: "mb-2" }, ["将要移除 ", h("span", { class: "fw-semibold", id: "movieAlbumRemoveDownloadedCount" }, String(props.downloadedItems.length)), " 条。"]),
          h("div", { class: "small text-muted mb-2" }, "只会移除相册条目，不会删除媒体文件。"),
          h("div", { class: "border rounded p-2 remove-downloaded-preview" }, [
            h("div", { id: "movieAlbumRemoveDownloadedList", class: "list-group list-group-flush" }, previewList())
          ])
        ]),
        h("div", { class: "modal-footer" }, [
          h("button", { type: "button", class: "btn btn-light", onClick: () => emit("close") }, "取消"),
          h("button", {
            type: "button",
            class: "btn btn-warning",
            id: "movieAlbumRemoveDownloadedConfirmBtn",
            disabled: !props.downloadedItems.length,
            onClick: () => emit("confirmRemoveDownloaded")
          }, "确认移除")
        ])
      ])])]),
      props.activeModal ? h("div", { class: "modal-backdrop fade show", onClick: () => emit("close") }) : null
    ]);
  }
});
</script>
