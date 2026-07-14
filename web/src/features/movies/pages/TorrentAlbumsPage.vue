<template>
  <PageShell class="torrent-albums-page-shell sc-events-page-shell" title="TorrentAlbums" note="" eyebrow="">
    <div id="page-root" class="container-fluid torrent-albums-page-root albums-list-standard torrent-albums-list-standard next02-sc-pink-theme next02-sc-events-rewrite">
      <div v-if="loading" class="alert albums-list-feedback is-loading">加载中...</div>
      <div v-else-if="error" class="alert albums-list-feedback is-error">{{ error }}</div>
      <div v-else id="torrentAlbumsPage" class="torrent-albums-page" :data-album-id="selectedAlbumID">
        <header class="albums-list-hero">
          <div class="albums-list-hero-copy">
            <span class="albums-list-eyebrow">Torrent Albums</span>
            <h2>{{ pageTitle }}</h2>
            <p v-if="pageNote">{{ pageNote }}</p>
          </div>
          <div class="albums-list-hero-actions">
            <button type="button" class="btn albums-manage-button" id="btnOpenManageTorrentAlbumsModal" @click="openModal('manage')">管理相册</button>
            <button type="button" class="btn albums-create-button" id="btnOpenCreateTorrentAlbumModal" @click="openModal('create')">新增相册</button>
          </div>
        </header>

        <div id="torrentAlbumsMsg" class="alert albums-list-message" :class="messageClass" role="alert">
          {{ messageText }}
        </div>

        <section class="albums-list-panel next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
          <div class="albums-list-controls">
            <form ref="filterForm" class="albums-filter-form torrent-albums-filter-form" id="torrentAlbumsFilterForm" method="get" action="/torrentalbumlist">
              <input v-if="queryValue('PageSize')" type="hidden" name="ps" :value="queryValue('PageSize')">
              <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')">
              <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')">

              <div class="albums-filter-row is-primary">
                <div class="albums-filter-field is-quick-filter">
                  <span>快捷相册</span>
                  <div class="album-quick-filter-group">
                  <button
                    type="button"
                    class="btn album-quick-button js-torrent-album-quick-filter"
                    :class="{ 'is-active': selectedAlbumName === '下载中' }"
                    data-album-name="下载中"
                    @click="quickFilter('下载中')"
                  >下载中</button>
                  <button
                    type="button"
                    class="btn album-quick-button js-torrent-album-quick-filter"
                    :class="{ 'is-active': selectedAlbumName === '待下载' }"
                    data-album-name="待下载"
                    @click="quickFilter('待下载')"
                  >待下载</button>
                  <button
                    type="button"
                    class="btn album-quick-button js-torrent-album-quick-filter"
                    :class="{ 'is-active': selectedAlbumName === 'Media' }"
                    data-album-name="Media"
                    @click="quickFilter('Media')"
                  >Media</button>
                  </div>
                </div>
                <label class="albums-filter-field is-album">
                  <span>相册</span>
                <select class="form-select" id="torrentAlbumsAlbumName" name="album_name" :value="selectedAlbumName">
                  <option v-for="option in albumSelectOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
                </label>
                <label class="albums-filter-field is-keyword">
                  <span>影片编号</span>
                <input class="form-control" type="text" name="keyword" :value="queryValue('Keyword')" placeholder="movie_name 或 movie_jav_id">
                </label>
                <label class="albums-filter-field is-source">
                  <span>来源</span>
                <select class="form-select" name="source_type" :value="queryValue('SourceType')">
                  <option v-for="option in sourceOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
                </label>
                <label class="albums-filter-field is-hash">
                  <span>Info Hash</span>
                <input class="form-control" type="text" name="info_hash" :value="queryValue('InfoHash')" placeholder="按 Info Hash 模糊筛选">
                </label>
              </div>

              <div class="albums-filter-row is-secondary">
                <label class="albums-filter-field is-date">
                  <span>发布开始</span>
                <input class="form-control smart-pick-date-input" type="text" name="publish_from" :value="queryValue('PublishFrom')" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
                </label>
                <label class="albums-filter-field is-date">
                  <span>发布结束</span>
                <input class="form-control smart-pick-date-input" type="text" name="publish_to" :value="queryValue('PublishTo')" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
                </label>
                <label class="albums-filter-field is-date">
                  <span>下载开始</span>
                <input class="form-control smart-pick-date-input" type="text" name="created_from" :value="queryValue('CreatedFrom')" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
                </label>
                <label class="albums-filter-field is-date">
                  <span>下载结束</span>
                <input class="form-control smart-pick-date-input" type="text" name="created_to" :value="queryValue('CreatedTo')" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
                </label>
                <div class="albums-filter-actions">
                  <button class="btn albums-apply-button" type="submit">筛选</button>
                  <a class="btn albums-clear-button" href="/torrentalbumlist">清空</a>
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
              <TorrentAlbumPagination v-if="hasRows && pageLinks.length" :links="pageLinks" />
            </div>
          </div>

          <div class="albums-batch-toolbar">
            <div class="albums-selected-count">已选 <span id="torrentAlbumBatchSelectedCount">{{ selectedCount }}</span> 项</div>
            <div class="albums-batch-actions">
              <button type="button" class="btn albums-batch-button" id="btnBatchCopyTorrentAlbumHashes" :disabled="selectedCount === 0" @click="batchCopyHashes">批量复制 Hash</button>
              <button type="button" class="btn albums-batch-button" id="btnBatchMoveTorrentAlbumItems" :disabled="selectedCount === 0 || selectedAlbumID <= 0" @click="openBatchMove">批量移动</button>
              <button type="button" class="btn albums-batch-button is-danger" id="btnBatchRemoveTorrentAlbumItems" :disabled="selectedCount === 0 || selectedAlbumID <= 0" @click="openBatchRemove">批量移除</button>
            </div>
          </div>

          <div class="table-responsive albums-table-wrap">
              <table class="table table-hover align-middle mb-0 sc-events-table torrent-album-table">
                <colgroup>
                  <col class="torrent-album-col-select">
                  <col class="torrent-album-col-movie">
                  <col class="torrent-album-col-source">
                  <col class="torrent-album-col-hash">
                  <col class="torrent-album-col-size">
                  <col class="torrent-album-col-publish">
                  <col class="torrent-album-col-created">
                  <col class="torrent-album-col-actions">
                </colgroup>
                <thead>
                  <tr>
                    <th class="text-center text-nowrap">
                      <input
                        class="form-check-input js-torrent-album-item-check-all"
                        id="torrentAlbumItemCheckAll"
                        type="checkbox"
                        aria-label="全选当前页"
                        :checked="allSelected"
                        :indeterminate="partialSelected"
                        @change="toggleAll(($event.target as HTMLInputElement).checked)"
                      >
                    </th>
                    <th class="sc-sortable-head"><SortLink :link="sortLink('ByMovieName')" label="影片" /></th>
                    <th class="sc-sortable-head"><SortLink :link="sortLink('BySourceType')" label="来源" /></th>
                    <th>Info Hash</th>
                    <th class="sc-number">大小</th>
                    <th class="sc-sortable-head sc-number"><SortLink :link="sortLink('ByPublish')" label="发布日期" /></th>
                    <th class="sc-sortable-head sc-number"><SortLink :link="sortLink('ByCreated')" label="下载中时间" /></th>
                    <th class="text-nowrap">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td colspan="8"><div class="sc-empty-state">当前筛选下没有相册资源数据。</div></td>
                  </tr>
                  <tr v-for="row in rows" v-else :key="rowKey(row)" :data-item-id="value(row, 'ID')">
                    <td class="text-center">
                      <input
                        class="form-check-input js-torrent-album-item-check"
                        type="checkbox"
                        :value="value(row, 'ID')"
                        :aria-label="`选择 ${value(row, 'MovieName')}`"
                        :checked="isSelected(row)"
                        @change="toggleRow(row, ($event.target as HTMLInputElement).checked)"
                      >
                    </td>
                    <td class="sc-content-cell text-nowrap" :title="value(row, 'MovieName') || value(row, 'MovieJavID')">
                      <a class="album-entry-link" :href="movieHref(row)" target="_blank" rel="noopener noreferrer">{{ value(row, "MovieName") || value(row, "MovieJavID") }}</a>
                    </td>
                    <td class="text-nowrap"><span class="badge" :class="sourceBadgeClass(row)">{{ sourceBadgeText(row) }}</span></td>
                    <td class="album-hash-cell"><code class="album-hash">{{ value(row, "InfoHash") }}</code></td>
                    <td class="sc-number text-nowrap">{{ value(row, "Size") || "-" }}</td>
                    <td class="sc-number text-nowrap">{{ formatUnixDate(value(row, "PublishTime")) }}</td>
                    <td class="sc-number text-nowrap"><ScDateTime :value="value(row, 'CreatedOn')" unix /></td>
                    <td class="text-nowrap">
                      <button type="button" class="btn btn-sm btn-outline-secondary js-copy-hash" :data-hash="value(row, 'InfoHash')" @click="copyHash(value(row, 'InfoHash'))">复制 Hash</button>
                      <button type="button" class="btn btn-sm btn-outline-danger js-remove-torrent-album-item ms-1" :data-item-id="value(row, 'ID')" @click="removeOne(row)">移除</button>
                    </td>
                  </tr>
                </tbody>
              </table>
          </div>
        </section>

        <TorrentAlbumModals
          :active-modal="activeModal"
          :move-target-options="moveTargetOptions"
          :album-options="manageAlbumOptions"
          :selected-album-id="selectedAlbumID"
          :pending-count="pendingIDs.length"
          :create-name="createName"
          @close="closeModal"
          @confirm-remove="batchRemove"
          @confirm-move="batchMove"
          @update-create-name="createName = $event"
          @confirm-create="createAlbum"
          @rename-album="renameManagedAlbum"
          @delete-album="deleteManagedAlbum"
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
const pendingIDs = ref<number[]>([]);
const loading = ref(false);
const error = ref("");
const messageText = ref("");
const messageOk = ref(true);
const activeModal = ref<"" | "remove" | "move" | "create" | "manage">("");
const createName = ref("");
const filterForm = ref<HTMLFormElement | null>(null);

const sourceOptions = [
  { value: "", label: "全部" },
  { value: "javbus_magnet", label: "JavBus 磁力" },
  { value: "sukebei_torrent", label: "Sukebei 种子" },
  { value: "sh_magnet", label: "色花堂磁力" }
];

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
const albumSelectOptions = computed(() => {
  const options = albums.value.map((album) => ({ value: value(album, "Name"), label: value(album, "Name") || value(album, "ID") }));
  if (!options.some((item) => String(item.value) === String(selectedAlbumName.value || ""))) {
    options.unshift({ value: selectedAlbumName.value || "", label: selectedAlbumName.value || "默认" });
  }
  return options;
});
const moveTargetOptions = computed(() => albums.value
  .filter((album) => Number(value(album, "ID")) > 0 && Number(value(album, "ID")) !== selectedAlbumID.value)
  .map((album) => ({ value: value(album, "ID"), label: value(album, "Name") })));
const manageAlbumOptions = computed(() => albums.value
  .filter((album) => Number(value(album, "ID")) > 0)
  .map((album) => ({
    value: value(album, "ID"),
    label: value(album, "Name"),
    protected: pick(album, "Protected") === true
  })));

function queryValue(key: string): string {
  return text(query.value, key);
}

function value(row: AnyRecord, key: string): string {
  return text(row, key);
}

function rowKey(row: AnyRecord): string {
  return [value(row, "ID"), value(row, "MovieName"), value(row, "InfoHash")].join(":");
}

function sortLink(key: string): AnyRecord {
  return asRecord(sortQuery.value[key]);
}

function movieHref(row: AnyRecord): string {
  return `/moviedetail/${encodeURIComponent(value(row, "MovieName") || value(row, "MovieJavID"))}`;
}

function sourceBadgeText(row: AnyRecord): string {
  const sourceType = value(row, "SourceType");
  if (sourceType === "javbus_magnet") return "JavBus";
  if (sourceType === "sukebei_torrent") return "Sukebei";
  if (sourceType === "sh_magnet") return "色花堂";
  return "未知";
}

function sourceBadgeClass(row: AnyRecord): string {
  const sourceType = value(row, "SourceType");
  if (sourceType === "javbus_magnet") return "bg-primary-subtle text-primary";
  if (sourceType === "sukebei_torrent") return "bg-danger-subtle text-danger";
  if (sourceType === "sh_magnet") return "bg-warning-subtle text-warning";
  return "bg-secondary-subtle text-secondary";
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
  const url = new URL("/torrentalbumlist", window.location.origin);
  url.searchParams.set("album_name", albumName);
  window.location.href = `${url.pathname}${url.search}`;
}

async function copyText(raw: string) {
  const value = String(raw || "").trim();
  if (!value) throw new Error("缺少可复制内容");
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(value);
    return;
  }
  const input = document.createElement("textarea");
  input.value = value;
  input.style.position = "fixed";
  input.style.left = "-9999px";
  document.body.appendChild(input);
  input.focus();
  input.select();
  const ok = document.execCommand("copy");
  input.remove();
  if (!ok) throw new Error("复制失败");
}

async function copyHash(hash: string) {
  try {
    await copyText(hash);
    showMessage("Info Hash 已复制");
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "复制失败", false);
  }
}

async function batchCopyHashes() {
  try {
    const hashes = rows.value.filter((row) => selectedIDs.value.has(rowID(row))).map((row) => value(row, "InfoHash")).filter(Boolean);
    if (!hashes.length) throw new Error("请先选择要复制的条目");
    await copyText(hashes.join("\n"));
    showMessage(`已复制 ${hashes.length} 条 Hash`);
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "复制失败", false);
  }
}

async function removeOne(row: AnyRecord) {
  const id = rowID(row);
  if (selectedAlbumID.value <= 0 || id <= 0) {
    showMessage("当前相册无效，无法移除", false);
    return;
  }
  try {
    const data = asRecord(await sendJSON(`/api/albums/${selectedAlbumID.value}/items/remove`, "POST", { item_id: id }));
    if (data.removed) removeRowsByIDs([id]);
    showMessage(text(data, "message") || "移除完成");
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "移除失败", false);
  }
}

function removeRowsByIDs(ids: number[]) {
  const idSet = new Set(ids);
  rows.value = rows.value.filter((row) => !idSet.has(rowID(row)));
  selectedIDs.value = new Set([...selectedIDs.value].filter((id) => !idSet.has(id)));
}

function openModal(name: "remove" | "move" | "create" | "manage") {
  activeModal.value = name;
}

function closeModal() {
  activeModal.value = "";
  pendingIDs.value = [];
}

function openBatchMove() {
  pendingIDs.value = [...selectedIDs.value];
  if (!pendingIDs.value.length) {
    showMessage("请先选择要移动的条目", false);
    return;
  }
  openModal("move");
}

function openBatchRemove() {
  pendingIDs.value = [...selectedIDs.value];
  if (!pendingIDs.value.length) {
    showMessage("请先选择要移除的条目", false);
    return;
  }
  openModal("remove");
}

async function batchRemove() {
  const ids = pendingIDs.value.slice();
  closeModal();
  try {
    const data = asRecord(await sendJSON(`/api/albums/${selectedAlbumID.value}/items/batch-remove`, "POST", { item_ids: ids }));
    const failed = new Set(asArray<number>(data.failed_ids).map(Number).filter((id) => id > 0));
    const removed = ids.filter((id) => !failed.has(id));
    removeRowsByIDs(removed);
    showMessage(failed.size ? `已移除 ${text(data, "removed_count") || removed.length} 条，失败 ${failed.size} 条` : `已批量移除 ${text(data, "removed_count") || removed.length} 条`, failed.size === 0);
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "批量移除失败", false);
  }
}

async function batchMove(targetAlbumID: string) {
  const targetID = Number(targetAlbumID || 0);
  if (targetID <= 0) {
    showMessage("请选择目标相册", false);
    return;
  }
  const ids = pendingIDs.value.slice();
  closeModal();
  try {
    const data = asRecord(await sendJSON(`/api/albums/${selectedAlbumID.value}/items/batch-move`, "POST", { item_ids: ids, target_album_id: targetID }));
    const failed = new Set(asArray<number>(data.failed_ids).map(Number).filter((id) => id > 0));
    const moved = ids.filter((id) => !failed.has(id));
    removeRowsByIDs(moved);
    showMessage(failed.size ? `已移动 ${text(data, "moved_count") || moved.length} 条，失败 ${failed.size} 条` : `已批量移动 ${text(data, "moved_count") || moved.length} 条`, failed.size === 0);
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "批量移动失败", false);
  }
}

async function createAlbum() {
  const name = createName.value.trim();
  if (!name) {
    showMessage("相册名称不能为空", false);
    return;
  }
  try {
    const data = asRecord(await sendJSON("/api/albums", "POST", { name }));
    closeModal();
    showMessage(text(data, "message") || "相册创建成功");
    const albumName = text(data, "album_name") || name;
    const url = new URL("/torrentalbumlist", window.location.origin);
    url.searchParams.set("album_name", albumName);
    window.location.href = `${url.pathname}${url.search}`;
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "新增相册失败", false);
  }
}

async function renameManagedAlbum(input: { id: string; name: string }) {
  const albumID = Number(input.id || 0);
  const name = String(input.name || "").trim();
  if (albumID <= 0 || !name) {
    showMessage("相册名称不能为空", false);
    return;
  }
  try {
    const data = asRecord(await sendJSON(`/api/albums/${albumID}`, "PUT", { name }));
    closeModal();
    const updatedName = text(data, "album_name") || name;
    if (albumID === selectedAlbumID.value) {
      const url = new URL("/torrentalbumlist", window.location.origin);
      url.searchParams.set("album_name", updatedName);
      window.location.href = `${url.pathname}${url.search}`;
      return;
    }
    window.location.reload();
  } catch (err) {
    closeModal();
    showMessage(err instanceof Error ? err.message : "编辑相册失败", false);
  }
}

async function deleteManagedAlbum(albumIDValue: string) {
  const albumID = Number(albumIDValue || 0);
  if (albumID <= 0) {
    showMessage("相册 ID 无效", false);
    return;
  }
  try {
    await sendJSON(`/api/albums/${albumID}`, "DELETE");
    closeModal();
    if (albumID === selectedAlbumID.value) {
      window.location.href = "/torrentalbumlist";
      return;
    }
    window.location.reload();
  } catch (err) {
    closeModal();
    showMessage(err instanceof Error ? err.message : "删除相册失败", false);
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

const TorrentAlbumPagination = defineComponent({
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

const TorrentAlbumModals = defineComponent({
  props: {
    activeModal: { type: String, required: true },
    moveTargetOptions: { type: Array as PropType<{ value: string; label: string }[]>, required: true },
    albumOptions: { type: Array as PropType<{ value: string; label: string; protected: boolean }[]>, required: true },
    selectedAlbumId: { type: Number, required: true },
    pendingCount: { type: Number, required: true },
    createName: { type: String, required: true }
  },
  emits: ["close", "confirmRemove", "confirmMove", "updateCreateName", "confirmCreate", "renameAlbum", "deleteAlbum"],
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
    return () => h("div", [
      h("div", {
        class: modalClass("manage"),
        id: "torrentAlbumManageModal",
        tabindex: "-1",
        "aria-labelledby": "torrentAlbumManageModalLabel",
        "aria-hidden": ariaHidden("manage")
      }, [h("div", { class: "modal-dialog modal-dialog-centered albums-manage-dialog" }, [h("div", { class: "modal-content" }, [
        h("div", { class: "modal-header" }, [
          h("h5", { class: "modal-title", id: "torrentAlbumManageModalLabel" }, "管理相册"),
          h("button", { type: "button", class: "btn-close", "aria-label": "Close", onClick: closeManagement })
        ]),
        h("div", { class: "modal-body albums-manage-body" }, deletingAlbumID.value
          ? [
              h("div", { class: "albums-manage-delete-confirm" }, [
                h("h6", "删除相册确认"),
                h("p", ["确认删除相册“", h("strong", deletingAlbumName.value), "”吗？"]),
                h("p", { class: "text-muted small mb-0" }, "仅空相册可以删除；相册中存在任何收纳记录时，删除操作会被拒绝。")
              ])
            ]
          : [
              h("p", { class: "text-muted small mb-2" }, "可编辑或删除自定义相册；系统相册由现有业务流程使用。"),
              h("div", { class: "albums-manage-list" }, props.albumOptions.map((option) => {
                const isEditing = editingAlbumID.value === option.value;
                const isSelected = Number(option.value) === props.selectedAlbumId;
                return h("div", { class: ["albums-manage-item", isSelected ? "is-selected" : ""], key: option.value }, [
                  h("div", { class: "albums-manage-item-main" }, isEditing
                    ? [h("input", {
                        class: "form-control albums-manage-name-input",
                        value: editingAlbumName.value,
                        maxlength: "128",
                        "aria-label": `编辑相册 ${option.label}`,
                        onInput: (event: Event) => { editingAlbumName.value = (event.target as HTMLInputElement).value; }
                      })]
                    : [
                        h("span", { class: "albums-manage-name" }, option.label),
                        option.protected ? h("span", { class: "albums-manage-system-badge" }, "系统相册") : null,
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
                          "aria-label": `编辑相册 ${option.label}`,
                          onClick: () => { editingAlbumID.value = option.value; editingAlbumName.value = option.label; }
                        }, "编辑"),
                        h("button", {
                          type: "button",
                          class: "btn btn-outline-danger btn-sm",
                          disabled: option.protected,
                          "aria-label": `删除相册 ${option.label}`,
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
            ? h("button", { type: "button", class: "btn btn-danger", id: "torrentAlbumDeleteConfirmBtn", onClick: () => emit("deleteAlbum", deletingAlbumID.value) }, "确认删除")
            : h("button", { type: "button", class: "btn btn-light", onClick: closeManagement }, "关闭")
        ])
      ])])]),
      h("div", {
        class: modalClass("remove"),
        id: "torrentAlbumBatchRemoveModal",
        tabindex: "-1",
        "aria-labelledby": "torrentAlbumBatchRemoveModalLabel",
        "aria-hidden": ariaHidden("remove")
      }, [h("div", { class: "modal-dialog modal-dialog-centered" }, [h("div", { class: "modal-content" }, [
        h("div", { class: "modal-header" }, [
          h("h5", { class: "modal-title", id: "torrentAlbumBatchRemoveModalLabel" }, "批量移除确认"),
          h("button", { type: "button", class: "btn-close", "aria-label": "Close", onClick: () => emit("close") })
        ]),
        h("div", { class: "modal-body" }, [h("p", { class: "mb-0" }, ["确认移除已选中的 ", h("span", { id: "torrentAlbumBatchRemoveCount" }, String(props.pendingCount)), " 条记录吗？"])]),
        h("div", { class: "modal-footer" }, [
          h("button", { type: "button", class: "btn btn-light", onClick: () => emit("close") }, "取消"),
          h("button", { type: "button", class: "btn btn-danger", id: "torrentAlbumBatchRemoveConfirmBtn", onClick: () => emit("confirmRemove") }, "确认移除")
        ])
      ])])]),
      h("div", {
        class: modalClass("move"),
        id: "torrentAlbumBatchMoveModal",
        tabindex: "-1",
        "aria-labelledby": "torrentAlbumBatchMoveModalLabel",
        "aria-hidden": ariaHidden("move")
      }, [h("div", { class: "modal-dialog modal-dialog-centered" }, [h("div", { class: "modal-content" }, [
        h("div", { class: "modal-header" }, [
          h("h5", { class: "modal-title", id: "torrentAlbumBatchMoveModalLabel" }, "批量移动确认"),
          h("button", { type: "button", class: "btn-close", "aria-label": "Close", onClick: () => emit("close") })
        ]),
        h("div", { class: "modal-body" }, [
          h("label", { class: "form-label mb-1", for: "torrentAlbumBatchMoveTarget" }, "目标相册"),
          h("select", { class: "form-select", id: "torrentAlbumBatchMoveTarget" }, [
            h("option", { value: "" }, "请选择目标相册"),
            ...props.moveTargetOptions.map((option) => h("option", { value: option.value }, option.label))
          ]),
          h("p", { class: "text-muted small mt-2 mb-0" }, ["将已选中的 ", h("span", { id: "torrentAlbumBatchMoveCount" }, String(props.pendingCount)), " 条记录移动到目标相册。"])
        ]),
        h("div", { class: "modal-footer" }, [
          h("button", { type: "button", class: "btn btn-light", onClick: () => emit("close") }, "取消"),
          h("button", { type: "button", class: "btn btn-primary", id: "torrentAlbumBatchMoveConfirmBtn", onClick: () => {
            const el = document.getElementById("torrentAlbumBatchMoveTarget") as HTMLSelectElement | null;
            emit("confirmMove", el?.value || "");
          } }, "确认移动")
        ])
      ])])]),
      h("div", {
        class: modalClass("create"),
        id: "torrentAlbumCreateModal",
        tabindex: "-1",
        "aria-labelledby": "torrentAlbumCreateModalLabel",
        "aria-hidden": ariaHidden("create")
      }, [h("div", { class: "modal-dialog modal-dialog-centered" }, [h("div", { class: "modal-content" }, [
        h("div", { class: "modal-header" }, [
          h("h5", { class: "modal-title", id: "torrentAlbumCreateModalLabel" }, "新增相册"),
          h("button", { type: "button", class: "btn-close", "aria-label": "Close", onClick: () => emit("close") })
        ]),
        h("div", { class: "modal-body" }, [
          h("label", { class: "form-label mb-1", for: "torrentAlbumCreateName" }, "相册名称"),
          h("input", {
            type: "text",
            class: "form-control",
            id: "torrentAlbumCreateName",
            maxlength: "128",
            placeholder: "请输入相册名称",
            value: props.createName,
            onInput: (event: Event) => emit("updateCreateName", (event.target as HTMLInputElement).value),
            onKeydown: (event: KeyboardEvent) => {
              if (event.key !== "Enter") return;
              event.preventDefault();
              emit("confirmCreate");
            }
          }),
          h("p", { class: "text-muted small mt-2 mb-0" }, "创建成功后会自动切换到该相册并刷新列表。")
        ]),
        h("div", { class: "modal-footer" }, [
          h("button", { type: "button", class: "btn btn-light", onClick: () => emit("close") }, "取消"),
          h("button", { type: "button", class: "btn btn-primary", id: "torrentAlbumCreateConfirmBtn", onClick: () => emit("confirmCreate") }, "确认创建")
        ])
      ])])]),
      props.activeModal ? h("div", { class: "modal-backdrop fade show" }) : null
    ]);
  }
});
</script>
