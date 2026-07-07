<template>
  <PageShell class="torrent-albums-page-shell movie-albums-page-shell" title="MovieAlbums" note="" eyebrow="">
    <div id="page-root" class="container-fluid torrent-albums-page-root movie-albums-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <div
        v-else
        id="movieAlbumItemsPage"
        class="torrent-albums-page movie-albums-page"
        :data-album-id="selectedAlbumID"
        :data-album-name="selectedAlbumName"
      >
        <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
          <div>
            <h4 class="mb-1">{{ pageTitle }}</h4>
            <p v-if="pageNote" class="text-muted mb-0">{{ pageNote }}</p>
          </div>
        </div>

        <div id="movieAlbumItemsMsg" class="alert mb-3" :class="messageClass" role="alert">
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
            <button type="button" class="btn btn-outline-primary btn-sm" id="btnOpenCreateMovieAlbumModal" @click="openCreateModal">新增电影相册</button>
          </div>
          <div v-if="hasRows && pageLinks.length" class="ms-xl-auto">
            <MovieAlbumPagination :links="pageLinks" />
          </div>
        </div>

        <div class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
          <div class="small text-muted">已选 <span id="movieAlbumBatchSelectedCount">{{ selectedCount }}</span> 项</div>
          <div v-if="canExecuteRemove || canRemoveDownloaded" class="d-flex flex-wrap gap-2">
            <button
              v-if="canExecuteRemove"
              type="button"
              class="btn btn-sm btn-danger"
              id="btnBatchExecuteMovieAlbumRemove"
              :disabled="selectedCount === 0 || selectedAlbumID <= 0"
              @click="openExecuteModal"
            >统一删除</button>
            <button
              v-if="canRemoveDownloaded"
              type="button"
              class="btn btn-sm btn-outline-warning"
              id="btnRemoveDownloadedMovieAlbumItems"
              :disabled="loadingDownloaded"
              @click="loadRemoveDownloadedPreview"
            >移除已下载条目</button>
          </div>
        </div>

        <div class="section-card card mb-3">
          <div class="card-body">
            <form ref="filterForm" class="row g-3 align-items-end" id="movieAlbumsFilterForm" method="get" action="/movie-albums">
              <input v-if="queryValue('PageSize')" type="hidden" name="ps" :value="queryValue('PageSize')">
              <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')">
              <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')">

              <div class="col-12">
                <label class="form-label mb-1">快捷相册</label>
                <div class="d-flex flex-wrap gap-2 album-quick-filter-group">
                  <button
                    type="button"
                    class="btn btn-sm js-movie-album-quick-filter"
                    :class="selectedAlbumName === '待下载电影' ? 'btn-warning' : 'btn-outline-warning'"
                    data-album-name="待下载电影"
                    @click="quickFilter('待下载电影')"
                  >待下载电影</button>
                  <button
                    type="button"
                    class="btn btn-sm js-movie-album-quick-filter"
                    :class="selectedAlbumName === '待删除' ? 'btn-danger' : 'btn-outline-danger'"
                    data-album-name="待删除"
                    @click="quickFilter('待删除')"
                  >待删除</button>
                  <button
                    type="button"
                    class="btn btn-sm js-movie-album-quick-filter"
                    :class="selectedAlbumName === 'v3' ? 'btn-primary' : 'btn-outline-primary'"
                    data-album-name="v3"
                    @click="quickFilter('v3')"
                  >v3</button>
                  <button
                    type="button"
                    class="btn btn-sm js-movie-album-quick-filter"
                    :class="selectedAlbumName === 'vx' ? 'btn-primary' : 'btn-outline-primary'"
                    data-album-name="vx"
                    @click="quickFilter('vx')"
                  >vx</button>
                </div>
              </div>

              <div class="col-12 col-md-6 col-xl-4">
                <label class="form-label mb-1">相册</label>
                <select class="form-select" id="movieAlbumsAlbumName" name="album_name" :value="selectedAlbumName">
                  <option v-for="option in albumSelectOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
              </div>

              <div class="col-12 col-md-6 col-xl-4">
                <label class="form-label mb-1">影片编号 / 标题</label>
                <input class="form-control" type="text" name="keyword" :value="queryValue('Keyword')" placeholder="movie_name 或 movie_jav_id">
              </div>

              <div class="col-12 d-flex gap-2">
                <button class="btn btn-primary" type="submit">筛选</button>
                <a class="btn btn-outline-secondary" href="/movie-albums">清空</a>
              </div>
            </form>
          </div>
        </div>

        <div class="section-card card">
          <div class="card-body">
            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 movie-album-table">
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
                    <th><SortLink :link="sortLink('ByMovieName')" label="影片" /></th>
                    <th><SortLink :link="sortLink('ByCreated')" label="加入时间" /></th>
                    <th class="text-nowrap">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td colspan="4"><div class="sc-empty-state">当前筛选下没有电影相册数据。</div></td>
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
                    <td class="text-nowrap">
                      <a class="fw-semibold" :href="movieHref(row)">{{ value(row, "MovieName") || value(row, "MovieJavID") }}</a>
                    </td>
                    <td class="text-nowrap">{{ formatUnixMinute(value(row, "CreatedOn")) }}</td>
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
          </div>
        </div>

        <div v-if="hasRows && pageLinks.length" class="d-flex justify-content-end mt-3">
          <MovieAlbumPagination :links="pageLinks" />
        </div>

        <MovieAlbumModals
          :active-modal="activeModal"
          :selected-count="selectedCount"
          :create-name="createName"
          :downloaded-items="downloadedPreviewItems"
          :loading-downloaded="loadingDownloaded"
          @close="closeModal"
          @update-create-name="createName = $event"
          @confirm-create="createMovieAlbum"
          @confirm-execute="executeRemove"
          @confirm-remove-downloaded="removeDownloadedItems"
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
const loading = ref(false);
const error = ref("");
const messageText = ref("");
const messageOk = ref(true);
const activeModal = ref<"" | "create" | "execute" | "downloaded">("");
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

function queryValue(key: string): string {
  return text(query.value, key);
}

function value(row: AnyRecord, key: string): string {
  return text(row, key);
}

function rowKey(row: AnyRecord): string {
  return [value(row, "ID"), value(row, "MovieName"), value(row, "MovieJavID")].join(":");
}

function sortLink(key: string): AnyRecord {
  return asRecord(sortQuery.value[key]);
}

function movieHref(row: AnyRecord): string {
  return `/movie/${encodeURIComponent(value(row, "MovieName") || value(row, "MovieJavID"))}`;
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
  if (filterForm.value) {
    const select = filterForm.value.querySelector("#movieAlbumsAlbumName") as HTMLSelectElement | null;
    if (select) {
      select.value = albumName;
      filterForm.value.submit();
      return;
    }
  }
  const url = new URL("/movie-albums", window.location.origin);
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
    const url = new URL("/movie-albums", window.location.origin);
    url.searchParams.set("album_name", albumName);
    window.location.href = `${url.pathname}${url.search}`;
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "电影相册创建失败", false);
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
    showMessage(text(data, "message") || "统一删除完成");
    window.setTimeout(() => window.location.reload(), 500);
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
        class: ["album-sort-link", active ? "active" : ""],
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
    loadingDownloaded: { type: Boolean, required: true }
  },
  emits: ["close", "updateCreateName", "confirmCreate", "confirmExecute", "confirmRemoveDownloaded"],
  setup(props, { emit }) {
    const modalClass = (name: string) => ["modal", "fade", props.activeModal === name ? "show d-block" : ""];
    const ariaHidden = (name: string) => props.activeModal === name ? "false" : "true";
    const previewList = () => props.downloadedItems.length
      ? props.downloadedItems.map((item) => {
        const row = asRecord(item);
        const title = text(row, "MovieName") || text(row, "MovieJavID") || "-";
        const javID = text(row, "MovieJavID");
        return h("div", { class: "list-group-item px-0 py-2 border-0 border-bottom" }, [
          h("div", { class: "fw-semibold" }, title),
          javID ? h("div", { class: "small text-muted" }, [h("code", javID)]) : null
        ]);
      })
      : [h("div", { id: "movieAlbumRemoveDownloadedEmpty", class: "small text-muted" }, props.loadingDownloaded ? "正在加载预览..." : "没有可移除的已下载条目。")];
    return () => h("div", [
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
        h("div", { class: "modal-body" }, ["确认执行已选中的 ", h("span", { id: "movieAlbumExecuteRemoveCount" }, String(props.selectedCount)), " 项吗？这会移动对应的 native w_media 到 `001_process/005_removed`。"]),
        h("div", { class: "modal-footer" }, [
          h("button", { type: "button", class: "btn btn-light", onClick: () => emit("close") }, "取消"),
          h("button", { type: "button", class: "btn btn-danger", id: "movieAlbumExecuteRemoveConfirmBtn", onClick: () => emit("confirmExecute") }, "确认删除")
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
