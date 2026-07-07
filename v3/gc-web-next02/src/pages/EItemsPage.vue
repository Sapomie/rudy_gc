<template>
  <PageShell class="torrent-albums-page-shell e-items-page-shell" title="EItem 列表" note="" eyebrow="">
    <div id="page-root" class="container-fluid torrent-albums-page e-items-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div class="d-flex align-items-center justify-content-between mb-3">
          <div>
            <h4 class="mb-1">EItem 列表</h4>
            <p class="text-muted mb-0">浏览 e_item 列表，并按字段筛选条目。</p>
          </div>
        </div>

        <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
          <div class="sc-page-summary">
            <span class="sc-page-summary-label">总数</span>
            <span class="sc-page-summary-value" id="itemTotalCount">{{ total }}</span>
          </div>
          <div v-if="hasRows && pageLinks.length" class="ms-xl-auto">
            <PageLinks :links="pageLinks" />
          </div>
        </div>

        <div class="card mb-3">
          <div class="card-body">
            <form class="row g-2 align-items-end" method="get" action="/e-items">
              <input v-if="queryValue('PageSize')" type="hidden" name="ps" :value="queryValue('PageSize')">
              <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')">
              <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')">

              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">JavLib 编号</label>
                <input class="form-control" type="text" name="jav_id" :value="queryValue('JavID')" placeholder="JavLib 编号">
              </div>
              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">影片编号</label>
                <input class="form-control" type="text" name="name" :value="queryValue('Name')" placeholder="影片编号">
              </div>
              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">详情状态</label>
                <select class="form-select" name="has_detail" :value="queryValue('HasDetail')">
                  <option value="">全部</option>
                  <option value="none">无详情</option>
                  <option value="ok">已有详情</option>
                </select>
              </div>
              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">封面状态</label>
                <select class="form-select" name="has_download_cover" :value="queryValue('HasDownloadCover')">
                  <option value="">全部</option>
                  <option value="none">无本地封面</option>
                  <option value="ok">已有本地封面</option>
                  <option value="wrong">错误封面链接</option>
                </select>
              </div>
              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">中文字幕</label>
                <select class="form-select" name="has_chinese" :value="queryValue('HasChinese')">
                  <option value="">全部</option>
                  <option value="none">无中文字幕</option>
                  <option value="ok">有中文字幕</option>
                  <option value="error">翻译错误</option>
                  <option value="sensitive">含敏感词</option>
                </select>
              </div>
              <div class="col-12 col-md-6 col-xl-2">
                <label class="form-label mb-1">详情扫描状态</label>
                <select class="form-select" name="detail_need_scan" :value="queryValue('DetailNeedScan')">
                  <option value="">全部</option>
                  <option value="need">需要扫描</option>
                  <option value="no_need">无需扫描</option>
                  <option value="wrong">内容错误</option>
                </select>
              </div>
              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">详情时间从</label>
                <input class="form-control" type="date" name="detail_birth_time_from" :value="queryValue('DetailBirthTimeFrom')">
              </div>
              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">详情时间到</label>
                <input class="form-control" type="date" name="detail_birth_time_to" :value="queryValue('DetailBirthTimeTo')">
              </div>
              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">最后查询时间从</label>
                <input class="form-control" type="date" name="last_query_detail_time_from" :value="queryValue('LastQueryDetailTimeFrom')">
              </div>
              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">最后查询时间到</label>
                <input class="form-control" type="date" name="last_query_detail_time_to" :value="queryValue('LastQueryDetailTimeTo')">
              </div>
              <div class="col-12 d-flex gap-2">
                <button class="btn btn-primary" type="submit">筛选</button>
                <a class="btn btn-outline-secondary" href="/e-items">清空</a>
              </div>
            </form>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <div class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
              <div class="small text-muted">已选 <span id="itemBatchSelectedCount">{{ selectedCount }}</span> 项</div>
              <button type="button" class="btn btn-sm btn-outline-primary" id="btnBatchEditStatus" :disabled="selectedCount === 0" @click="openBatchModal">批量编辑状态</button>
            </div>
            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 list-table e-items-table">
                <thead>
                  <tr>
                    <th class="text-center text-nowrap"><input class="form-check-input js-item-check-all" id="itemCheckAll" type="checkbox" aria-label="全选当前页" :checked="allSelected" :indeterminate="someSelected" @change="toggleAllFromEvent"></th>
                    <th><SortLink :link="sortLink('ByJavID')" label="JavLib 编号" /></th>
                    <th><SortLink :link="sortLink('ByName')" label="影片编号" /></th>
                    <th><SortLink :link="sortLink('ByHasDetail')" label="详情状态" /></th>
                    <th><SortLink :link="sortLink('ByHasDownloadCover')" label="封面状态" /></th>
                    <th><SortLink :link="sortLink('ByHasChinese')" label="中文字幕" /></th>
                    <th><SortLink :link="sortLink('ByDetailNeedScan')" label="详情扫描状态" /></th>
                    <th><SortLink :link="sortLink('ByDetailBirthTime')" label="详情时间" /></th>
                    <th><SortLink :link="sortLink('ByLastQueryDetailTime')" label="最后查询时间" /></th>
                    <th class="text-nowrap">操作</th>
                  </tr>
                </thead>
                <tbody id="itemTableBody">
                  <tr v-if="!hasRows" class="js-item-empty-row">
                    <td colspan="10"><div class="sc-empty-state">当前筛选下没有 e_item 数据。</div></td>
                  </tr>
                  <tr v-for="row in rows" v-else :key="rowId(row)" :data-item-id="rowId(row)" :class="{ 'table-active': isSelected(rowId(row)) }">
                    <td class="text-center"><input class="form-check-input js-item-check" type="checkbox" :value="rowId(row)" :aria-label="`选择 ${rowText(row, 'jav_id', 'JavId', 'JavID')}`" :checked="isSelected(rowId(row))" @change="toggleSelection(rowId(row), $event)"></td>
                    <td class="text-nowrap">
                      <a v-if="rowText(row, 'jav_url', 'JavUrl', 'JavURL')" :href="rowText(row, 'jav_url', 'JavUrl', 'JavURL')" target="_blank" rel="noopener noreferrer">{{ rowText(row, "jav_id", "JavId", "JavID") }}</a>
                      <template v-else>{{ rowText(row, "jav_id", "JavId", "JavID") }}</template>
                    </td>
                    <td><a :href="`/movie/${encodeURIComponent(rowText(row, 'name', 'Name'))}`">{{ rowText(row, "name", "Name") }}</a></td>
                    <td class="js-item-has-detail">{{ rowText(row, "has_detail_text", "HasDetailText") || "-" }}</td>
                    <td class="js-item-has-download-cover">{{ rowText(row, "has_download_cover_text", "HasDownloadCoverText") || "-" }}</td>
                    <td class="js-item-has-chinese">{{ rowText(row, "has_chinese_text", "HasChineseText") || "-" }}</td>
                    <td class="js-item-detail-need-scan">{{ rowText(row, "detail_need_scan_text", "DetailNeedScanText") || "-" }}</td>
                    <td class="text-nowrap">{{ formatUnixMinute(rowValue(row, "detail_birth_time", "DetailBirthTime")) }}</td>
                    <td class="text-nowrap">{{ formatUnixMinute(rowValue(row, "last_query_detail_time", "LastQueryDetailTime")) }}</td>
                    <td class="text-nowrap">
                      <button type="button" class="btn btn-sm btn-outline-primary js-item-status-edit" :data-item-id="rowId(row)" :data-item-jav-id="rowText(row, 'jav_id', 'JavId', 'JavID')" :data-item-name="rowText(row, 'name', 'Name')" :data-has-detail="rowText(row, 'has_detail_value', 'HasDetailValue')" :data-has-download-cover="rowText(row, 'has_download_cover_value', 'HasDownloadCoverValue')" :data-has-chinese="rowText(row, 'has_chinese_value', 'HasChineseValue')" :data-detail-need-scan="rowText(row, 'detail_need_scan_value', 'DetailNeedScanValue')" @click="openSingleModal(row)">编辑状态</button>
                      <button type="button" class="btn btn-sm btn-outline-danger js-item-delete ms-1" :data-item-id="rowId(row)" :data-item-jav-id="rowText(row, 'jav_id', 'JavId', 'JavID')" :data-item-name="rowText(row, 'name', 'Name')" @click="openDeleteModal(row)">删除</button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <ItemStatusModal
          id-prefix="itemStatus"
          title="编辑状态"
          submit-label="保存"
          submitting-label="保存中..."
          :show="singleModal.show"
          :meta="singleModal.meta"
          :message="singleModal.message"
          :message-ok="singleModal.messageOk"
          :form="singleModal.form"
          @close="closeSingleModal"
          @submit="submitSingleStatus"
          @update-form="updateSingleForm"
        />
        <ItemStatusModal
          id-prefix="itemBatch"
          title="批量编辑状态"
          submit-label="批量保存"
          submitting-label="保存中..."
          :show="batchModal.show"
          :meta="batchModal.meta"
          :message="batchModal.message"
          :message-ok="batchModal.messageOk"
          :form="batchModal.form"
          :batch="true"
          @close="closeBatchModal"
          @submit="submitBatchStatus"
          @update-form="updateBatchForm"
        />
        <div class="modal fade" id="itemDeleteModal" tabindex="-1" aria-labelledby="itemDeleteModalLabel" aria-hidden="true" :class="{ show: deleteModal.show }" :style="{ display: deleteModal.show ? 'block' : 'none' }">
          <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
              <div class="modal-header">
                <div>
                  <h5 class="modal-title" id="itemDeleteModalLabel">删除电影</h5>
                  <div id="itemDeleteMeta" class="small text-muted">{{ deleteModal.meta }}</div>
                </div>
                <button type="button" class="btn-close" aria-label="Close" @click="closeDeleteModal"></button>
              </div>
              <div class="modal-body">
                <p class="mb-0">会删除这部电影在所有相关表中的数据，并写入删除记录。</p>
                <div id="itemDeleteMsg" class="alert small py-2 px-3 mt-3 mb-0" :class="deleteModal.message ? (deleteModal.messageOk ? 'alert-success' : 'alert-danger') : 'd-none'">{{ deleteModal.message }}</div>
                <input type="hidden" id="itemDeleteItemId" :value="deleteModal.itemId">
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-light" @click="closeDeleteModal">取消</button>
                <button type="button" class="btn btn-danger" id="itemDeleteSubmit" :disabled="deleteModal.submitting" @click="submitDelete">{{ deleteModal.submitting ? "删除中..." : "确认删除" }}</button>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, reactive, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord, PageInfo } from "@/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/utils/format";

type StatusForm = {
  hasDetail: string;
  hasDownloadCover: string;
  hasChinese: string;
  detailNeedScan: string;
};

const route = useRoute();
const payload = ref<AnyRecord>({});
const rows = ref<AnyRecord[]>([]);
const loading = ref(false);
const error = ref("");
const selectedIds = ref<string[]>([]);

const emptyStatusForm = (): StatusForm => ({
  hasDetail: "",
  hasDownloadCover: "",
  hasChinese: "",
  detailNeedScan: ""
});

const singleModal = reactive({
  show: false,
  itemId: "",
  meta: "",
  message: "",
  messageOk: false,
  submitting: false,
  form: emptyStatusForm()
});
const batchModal = reactive({
  show: false,
  meta: "",
  message: "",
  messageOk: false,
  submitting: false,
  form: emptyStatusForm()
});
const deleteModal = reactive({
  show: false,
  itemId: "",
  meta: "",
  message: "",
  messageOk: false,
  submitting: false
});

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const data = asRecord(await getJSON(apiPathForRoute(route)));
    payload.value = data;
    rows.value = asArray<AnyRecord>(pick(data, "Items", "items", "Rows", "rows"));
    selectedIds.value = [];
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const query = computed(() => asRecord(pick(payload.value, "Query", "query")));
const total = computed(() => text(payload.value, "Total", "total") || rows.value.length);
const hasRows = computed(() => rows.value.length > 0);
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<AnyRecord>(pageInfo.value.Links || pageInfo.value.links));
const sortMap = computed(() => asRecord(pick(payload.value, "ItemSortQuery", "itemSortQuery")));
const selectedCount = computed(() => selectedIds.value.length);
const allSelected = computed(() => rows.value.length > 0 && selectedIds.value.length === rows.value.length);
const someSelected = computed(() => selectedIds.value.length > 0 && !allSelected.value);

function queryValue(key: string): string {
  return text(query.value, key);
}

function rowValue(row: AnyRecord, ...keys: string[]): unknown {
  return pick(row, ...keys);
}

function rowText(row: AnyRecord, ...keys: string[]): string {
  return text(row, ...keys);
}

function rowId(row: AnyRecord): string {
  return rowText(row, "id", "Id");
}

function sortLink(key: string): AnyRecord {
  return asRecord(sortMap.value[key]);
}

function formatUnixMinute(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  const hh = String(d.getHours()).padStart(2, "0");
  const mi = String(d.getMinutes()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}`;
}

function isSelected(id: string): boolean {
  return selectedIds.value.includes(id);
}

function toggleSelection(id: string, event: Event) {
  const checked = Boolean((event.target as HTMLInputElement | null)?.checked);
  if (checked && !selectedIds.value.includes(id)) {
    selectedIds.value = [...selectedIds.value, id];
  } else if (!checked) {
    selectedIds.value = selectedIds.value.filter((item) => item !== id);
  }
}

function toggleAllFromEvent(event: Event) {
  const checked = Boolean((event.target as HTMLInputElement | null)?.checked);
  selectedIds.value = checked ? rows.value.map(rowId).filter(Boolean) : [];
}

function formFromRow(row: AnyRecord): StatusForm {
  return {
    hasDetail: rowText(row, "has_detail_value", "HasDetailValue") || "none",
    hasDownloadCover: rowText(row, "has_download_cover_value", "HasDownloadCoverValue") || "none",
    hasChinese: rowText(row, "has_chinese_value", "HasChineseValue") || "none",
    detailNeedScan: rowText(row, "detail_need_scan_value", "DetailNeedScanValue") || "need"
  };
}

function updateSingleForm(next: StatusForm) {
  singleModal.form = { ...next };
}

function updateBatchForm(next: StatusForm) {
  batchModal.form = { ...next };
}

function openSingleModal(row: AnyRecord) {
  singleModal.itemId = rowId(row);
  singleModal.meta = `${rowText(row, "jav_id", "JavId", "JavID") || "-"} / ${rowText(row, "name", "Name") || "-"}`;
  singleModal.message = "";
  singleModal.messageOk = false;
  singleModal.form = formFromRow(row);
  singleModal.show = true;
}

function closeSingleModal() {
  singleModal.show = false;
  singleModal.message = "";
  singleModal.submitting = false;
}

function openBatchModal() {
  batchModal.meta = `已选 ${selectedIds.value.length} 项`;
  batchModal.message = "";
  batchModal.messageOk = false;
  batchModal.form = emptyStatusForm();
  batchModal.show = true;
}

function closeBatchModal() {
  batchModal.show = false;
  batchModal.message = "";
  batchModal.submitting = false;
}

function openDeleteModal(row: AnyRecord) {
  deleteModal.itemId = rowId(row);
  deleteModal.meta = `${rowText(row, "jav_id", "JavId", "JavID") || "-"} / ${rowText(row, "name", "Name") || "-"}`;
  deleteModal.message = "";
  deleteModal.messageOk = false;
  deleteModal.show = true;
}

function closeDeleteModal() {
  deleteModal.show = false;
  deleteModal.message = "";
  deleteModal.submitting = false;
}

function mergeUpdatedItem(item: unknown) {
  const next = asRecord(item);
  const id = text(next, "id", "Id");
  if (!id) return;
  rows.value = rows.value.map((row) => rowId(row) === id ? { ...row, ...next } : row);
}

async function submitSingleStatus() {
  if (!singleModal.itemId) {
    singleModal.message = "缺少条目 id";
    singleModal.messageOk = false;
    return;
  }
  singleModal.submitting = true;
  try {
    const data = asRecord(await sendJSON(`/api/e-items/${encodeURIComponent(singleModal.itemId)}/status`, "POST", singleModal.form));
    if (data.ok === false) throw new Error(text(data, "error") || "保存失败");
    mergeUpdatedItem(data.item);
    singleModal.message = "状态已更新";
    singleModal.messageOk = true;
    window.setTimeout(closeSingleModal, 250);
  } catch (err) {
    singleModal.message = err instanceof Error ? err.message : "保存失败";
    singleModal.messageOk = false;
  } finally {
    singleModal.submitting = false;
  }
}

async function submitBatchStatus() {
  if (selectedIds.value.length === 0) {
    batchModal.message = "请先选择条目";
    batchModal.messageOk = false;
    return;
  }
  const payloadData: AnyRecord = { ids: selectedIds.value.map((id) => Number(id)).filter((id) => Number.isFinite(id) && id > 0) };
  for (const [key, value] of Object.entries(batchModal.form)) {
    if (value) payloadData[key] = value;
  }
  if (Object.keys(payloadData).length === 1) {
    batchModal.message = "至少选择一个要修改的状态";
    batchModal.messageOk = false;
    return;
  }
  batchModal.submitting = true;
  try {
    const data = asRecord(await sendJSON("/api/e-items/status/batch", "POST", payloadData));
    if (data.ok === false) throw new Error(text(data, "error") || "批量保存失败");
    asArray<AnyRecord>(data.items).forEach(mergeUpdatedItem);
    selectedIds.value = [];
    batchModal.message = "批量状态已更新";
    batchModal.messageOk = true;
    window.setTimeout(closeBatchModal, 250);
  } catch (err) {
    batchModal.message = err instanceof Error ? err.message : "批量保存失败";
    batchModal.messageOk = false;
  } finally {
    batchModal.submitting = false;
  }
}

async function submitDelete() {
  if (!deleteModal.itemId) {
    deleteModal.message = "缺少条目 id";
    deleteModal.messageOk = false;
    return;
  }
  deleteModal.submitting = true;
  try {
    const data = asRecord(await sendJSON(`/api/e-items/${encodeURIComponent(deleteModal.itemId)}`, "DELETE"));
    if (data.ok === false) throw new Error(text(data, "error") || "删除失败");
    const removedId = text(data, "id", "Id") || deleteModal.itemId;
    rows.value = rows.value.filter((row) => rowId(row) !== removedId);
    selectedIds.value = selectedIds.value.filter((id) => id !== removedId);
    deleteModal.message = "电影已删除";
    deleteModal.messageOk = true;
    window.setTimeout(closeDeleteModal, 250);
  } catch (err) {
    deleteModal.message = err instanceof Error ? err.message : "删除失败";
    deleteModal.messageOk = false;
  } finally {
    deleteModal.submitting = false;
  }
}

const SortLink = defineComponent({
  name: "EItemsSortLink",
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
        class: ["sc-sort-link", { active }],
        href: normalizeHref(text(props.link, "Href", "href") || "#")
      }, `${props.label}${suffix}`);
    };
  }
});

const PageLinks = defineComponent({
  name: "EItemsPagination",
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

const ItemStatusModal = defineComponent({
  name: "ItemStatusModal",
  props: {
    idPrefix: { type: String, required: true },
    title: { type: String, required: true },
    submitLabel: { type: String, required: true },
    submittingLabel: { type: String, required: true },
    show: { type: Boolean, required: true },
    meta: { type: String, required: true },
    message: { type: String, required: true },
    messageOk: { type: Boolean, required: true },
    form: { type: Object as PropType<StatusForm>, required: true },
    batch: { type: Boolean, default: false }
  },
  emits: ["close", "submit", "update-form"],
  setup(props, { emit }) {
    const select = (id: string, label: string, key: keyof StatusForm, options: [string, string][]) => h("div", { class: "col-12 col-md-6" }, [
      h("label", { for: id, class: "form-label" }, label),
      h("select", {
        id,
        class: "form-select",
        name: key,
        value: props.form[key],
        onChange: (event: Event) => {
          emit("update-form", { ...props.form, [key]: (event.target as HTMLSelectElement).value });
        }
      }, options.map(([value, textValue]) => h("option", { value }, textValue)))
    ]);
    return () => h("div", {
      class: ["modal", "fade", { show: props.show }],
      id: `${props.idPrefix}${props.batch ? "StatusModal" : "Modal"}`,
      tabindex: "-1",
      "aria-hidden": props.show ? "false" : "true",
      style: { display: props.show ? "block" : "none" }
    }, [
      h("div", { class: "modal-dialog modal-dialog-centered" }, [
        h("div", { class: "modal-content" }, [
          h("form", {
            id: `${props.idPrefix}${props.batch ? "StatusForm" : "Form"}`,
            onSubmit: (event: Event) => {
              event.preventDefault();
              emit("submit");
            }
          }, [
            h("div", { class: "modal-header" }, [
              h("div", {}, [
                h("h5", { class: "modal-title", id: `${props.idPrefix}${props.batch ? "StatusModalLabel" : "ModalLabel"}` }, props.title),
                h("div", { class: "small text-muted", id: `${props.idPrefix}${props.batch ? "StatusMeta" : "Meta"}` }, props.meta)
              ]),
              h("button", { type: "button", class: "btn-close", "aria-label": "Close", onClick: () => emit("close") })
            ]),
            h("div", { class: "modal-body" }, [
              h("div", { class: "row g-3" }, [
                select(`${props.idPrefix}HasDetail`, "详情状态", "hasDetail", props.batch ? [["", "不修改"], ["none", "无详情"], ["ok", "已有详情"]] : [["none", "无详情"], ["ok", "已有详情"]]),
                select(`${props.idPrefix}HasDownloadCover`, "封面状态", "hasDownloadCover", props.batch ? [["", "不修改"], ["none", "无本地封面"], ["ok", "已有本地封面"], ["wrong", "错误封面链接"]] : [["none", "无本地封面"], ["ok", "已有本地封面"], ["wrong", "错误封面链接"]]),
                select(`${props.idPrefix}HasChinese`, "中文字幕", "hasChinese", props.batch ? [["", "不修改"], ["none", "无中文字幕"], ["ok", "有中文字幕"], ["error", "翻译错误"], ["sensitive", "含敏感词"]] : [["none", "无中文字幕"], ["ok", "有中文字幕"], ["error", "翻译错误"], ["sensitive", "含敏感词"]]),
                select(`${props.idPrefix}DetailNeedScan`, "详情扫描状态", "detailNeedScan", props.batch ? [["", "不修改"], ["need", "需要扫描"], ["no_need", "无需扫描"], ["wrong", "内容错误"]] : [["need", "需要扫描"], ["no_need", "无需扫描"], ["wrong", "内容错误"]])
              ]),
              h("div", {
                id: `${props.idPrefix}${props.batch ? "StatusMsg" : "Msg"}`,
                class: ["alert small py-2 px-3 mt-3 mb-0", props.message ? (props.messageOk ? "alert-success" : "alert-danger") : "d-none"]
              }, props.message)
            ]),
            h("div", { class: "modal-footer" }, [
              h("button", { type: "button", class: "btn btn-light", onClick: () => emit("close") }, "取消"),
              h("button", { type: "submit", class: "btn btn-primary", id: `${props.idPrefix}${props.batch ? "StatusSubmit" : "Submit"}` }, props.submitLabel || props.submittingLabel)
            ])
          ])
        ])
      ])
    ]);
  }
});
</script>
