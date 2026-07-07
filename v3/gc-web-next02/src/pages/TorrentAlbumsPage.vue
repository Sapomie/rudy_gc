<template>
  <PageShell class="torrent-albums-page-shell" title="TorrentAlbums" note="" eyebrow="">
    <div id="page-root" class="container-fluid torrent-albums-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <div v-else id="torrentAlbumsPage" class="torrent-albums-page" :data-album-id="selectedAlbumID">
        <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
          <div>
            <h4 class="mb-1">{{ pageTitle }}</h4>
            <p v-if="pageNote" class="text-muted mb-0">{{ pageNote }}</p>
          </div>
        </div>

        <div id="torrentAlbumsMsg" class="alert mb-3" :class="messageClass" role="alert">
          {{ messageText }}
        </div>

        <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
          <div class="sc-page-summary">
            <span class="sc-page-summary-label">相册</span>
            <span class="sc-page-summary-value">{{ selectedAlbumName }}</span>
            <span class="sc-page-summary-label ms-3">总数</span>
            <span class="sc-page-summary-value">{{ total }}</span>
          </div>
          <div class="d-flex align-items-center gap-2">
            <button type="button" class="btn btn-outline-primary btn-sm" id="btnOpenCreateTorrentAlbumModal" @click="openModal('create')">新增相册</button>
          </div>
          <div v-if="hasRows && pageLinks.length" class="ms-xl-auto">
            <TorrentAlbumPagination :links="pageLinks" />
          </div>
        </div>

        <div class="section-card card mb-3">
          <div class="card-body">
            <form ref="filterForm" class="row g-3 align-items-end" id="torrentAlbumsFilterForm" method="get" action="/torrent-albums">
              <input v-if="queryValue('PageSize')" type="hidden" name="ps" :value="queryValue('PageSize')">
              <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')">
              <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')">

              <div class="col-12">
                <label class="form-label mb-1">快捷相册</label>
                <div class="d-flex flex-wrap gap-2 album-quick-filter-group">
                  <button
                    type="button"
                    class="btn btn-sm js-torrent-album-quick-filter"
                    :class="selectedAlbumName === '下载中' ? 'btn-warning' : 'btn-outline-warning'"
                    data-album-name="下载中"
                    @click="quickFilter('下载中')"
                  >下载中</button>
                  <button
                    type="button"
                    class="btn btn-sm js-torrent-album-quick-filter"
                    :class="selectedAlbumName === '待下载' ? 'btn-info' : 'btn-outline-info'"
                    data-album-name="待下载"
                    @click="quickFilter('待下载')"
                  >待下载</button>
                  <button
                    type="button"
                    class="btn btn-sm js-torrent-album-quick-filter"
                    :class="selectedAlbumName === 'Media' ? 'btn-primary' : 'btn-outline-primary'"
                    data-album-name="Media"
                    @click="quickFilter('Media')"
                  >Media</button>
                </div>
              </div>

              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">相册</label>
                <select class="form-select" id="torrentAlbumsAlbumName" name="album_name" :value="selectedAlbumName">
                  <option v-for="option in albumSelectOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
              </div>

              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">影片编号 / JavID</label>
                <input class="form-control" type="text" name="keyword" :value="queryValue('Keyword')" placeholder="movie_name 或 movie_jav_id">
              </div>

              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">来源</label>
                <select class="form-select" name="source_type" :value="queryValue('SourceType')">
                  <option v-for="option in sourceOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
              </div>

              <div class="col-12 col-md-6 col-xl-4">
                <label class="form-label mb-1">Info Hash</label>
                <input class="form-control" type="text" name="info_hash" :value="queryValue('InfoHash')" placeholder="按 Info Hash 模糊筛选">
              </div>

              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">发布日期开始</label>
                <input class="form-control" type="date" name="publish_from" :value="queryValue('PublishFrom')">
              </div>

              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">发布日期结束</label>
                <input class="form-control" type="date" name="publish_to" :value="queryValue('PublishTo')">
              </div>

              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">下载中时间开始</label>
                <input class="form-control" type="date" name="created_from" :value="queryValue('CreatedFrom')">
              </div>

              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">下载中时间结束</label>
                <input class="form-control" type="date" name="created_to" :value="queryValue('CreatedTo')">
              </div>

              <div class="col-12 d-flex gap-2">
                <button class="btn btn-primary" type="submit">筛选</button>
                <a class="btn btn-outline-secondary" href="/torrent-albums">清空</a>
              </div>
            </form>
          </div>
        </div>

        <div class="section-card card">
          <div class="card-body">
            <div class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
              <div class="small text-muted">已选 <span id="torrentAlbumBatchSelectedCount">{{ selectedCount }}</span> 项</div>
              <div class="d-flex flex-wrap gap-2">
                <button type="button" class="btn btn-sm btn-outline-secondary" id="btnBatchCopyTorrentAlbumHashes" :disabled="selectedCount === 0" @click="batchCopyHashes">批量复制 Hash</button>
                <button type="button" class="btn btn-sm btn-outline-primary" id="btnBatchMoveTorrentAlbumItems" :disabled="selectedCount === 0 || selectedAlbumID <= 0" @click="openBatchMove">批量移动</button>
                <button type="button" class="btn btn-sm btn-outline-danger" id="btnBatchRemoveTorrentAlbumItems" :disabled="selectedCount === 0 || selectedAlbumID <= 0" @click="openBatchRemove">批量移除</button>
              </div>
            </div>

            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 torrent-album-table">
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
                    <th><SortLink :link="sortLink('ByMovieName')" label="影片" /></th>
                    <th><SortLink :link="sortLink('BySourceType')" label="来源" /></th>
                    <th>Info Hash</th>
                    <th>大小</th>
                    <th><SortLink :link="sortLink('ByPublish')" label="发布日期" /></th>
                    <th><SortLink :link="sortLink('ByCreated')" label="下载中时间" /></th>
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
                    <td class="text-nowrap">
                      <a class="fw-semibold" :href="movieHref(row)">{{ value(row, "MovieName") || value(row, "MovieJavID") }}</a>
                    </td>
                    <td class="text-nowrap"><span class="badge" :class="sourceBadgeClass(row)">{{ sourceBadgeText(row) }}</span></td>
                    <td class="album-hash-cell"><code class="album-hash">{{ value(row, "InfoHash") }}</code></td>
                    <td class="text-nowrap">{{ value(row, "Size") || "-" }}</td>
                    <td class="text-nowrap">{{ formatUnixDate(value(row, "PublishTime")) }}</td>
                    <td class="text-nowrap">{{ formatUnixMinute(value(row, "CreatedOn")) }}</td>
                    <td class="text-nowrap">
                      <button type="button" class="btn btn-sm btn-outline-secondary js-copy-hash" :data-hash="value(row, 'InfoHash')" @click="copyHash(value(row, 'InfoHash'))">复制 Hash</button>
                      <button type="button" class="btn btn-sm btn-outline-danger js-remove-torrent-album-item ms-1" :data-item-id="value(row, 'ID')" @click="removeOne(row)">移除</button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div v-if="hasRows && pageLinks.length" class="d-flex justify-content-end mt-3">
          <TorrentAlbumPagination :links="pageLinks" />
        </div>

        <TorrentAlbumModals
          :active-modal="activeModal"
          :move-target-options="moveTargetOptions"
          :pending-count="pendingIDs.length"
          :create-name="createName"
          @close="closeModal"
          @confirm-remove="batchRemove"
          @confirm-move="batchMove"
          @update-create-name="createName = $event"
          @confirm-create="createAlbum"
        />
      </div>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord, PageInfo } from "@/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const rows = ref<AnyRecord[]>([]);
const selectedIDs = ref<Set<number>>(new Set());
const pendingIDs = ref<number[]>([]);
const loading = ref(false);
const error = ref("");
const messageText = ref("");
const messageOk = ref(true);
const activeModal = ref<"" | "remove" | "move" | "create">("");
const createName = ref("");
const filterForm = ref<HTMLFormElement | null>(null);

const sourceOptions = [
  { value: "", label: "全部" },
  { value: "javbus_magnet", label: "JavBus 磁力" },
  { value: "sukebei_torrent", label: "Sukebei 种子" },
  { value: "sehuatang_magnet", label: "Sehuatang 磁力" }
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
  return `/movie/${encodeURIComponent(value(row, "MovieName") || value(row, "MovieJavID"))}`;
}

function sourceBadgeText(row: AnyRecord): string {
  const sourceType = value(row, "SourceType");
  if (sourceType === "javbus_magnet") return "JavBus";
  if (sourceType === "sukebei_torrent") return "Sukebei";
  if (sourceType === "sehuatang_magnet") return "Sehuatang";
  return "未知";
}

function sourceBadgeClass(row: AnyRecord): string {
  const sourceType = value(row, "SourceType");
  if (sourceType === "javbus_magnet") return "bg-primary-subtle text-primary";
  if (sourceType === "sukebei_torrent") return "bg-danger-subtle text-danger";
  if (sourceType === "sehuatang_magnet") return "bg-warning-subtle text-warning";
  return "bg-secondary-subtle text-secondary";
}

function formatUnixDate(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function formatUnixMinute(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")} ${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
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
  const url = new URL("/torrent-albums", window.location.origin);
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

function openModal(name: "remove" | "move" | "create") {
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
    const url = new URL("/torrent-albums", window.location.origin);
    url.searchParams.set("album_name", albumName);
    window.location.href = `${url.pathname}${url.search}`;
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "新增相册失败", false);
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
        class: ["album-sort-link", active ? "active" : ""],
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
    pendingCount: { type: Number, required: true },
    createName: { type: String, required: true }
  },
  emits: ["close", "confirmRemove", "confirmMove", "updateCreateName", "confirmCreate"],
  setup(props, { emit }) {
    const modalClass = (name: string) => ["modal", "fade", props.activeModal === name ? "show d-block" : ""];
    const ariaHidden = (name: string) => props.activeModal === name ? "false" : "true";
    return () => h("div", [
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
