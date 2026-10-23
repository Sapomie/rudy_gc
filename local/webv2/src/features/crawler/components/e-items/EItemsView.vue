<template>
  <PageShell class="e-items-page-shell sc-events-page-shell" title="EItem 列表" note="" eyebrow="">
    <div id="page-root" class="container-fluid e-items-page-root next02-sc-pink-theme next02-sc-events-rewrite">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell e-items-hero gc-page-head">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">E ITEMS / STATUS POOL</span>
            <h1>EItem 列表</h1>
            <p>浏览 e_item 数据，并按影片标识、业务状态和时间范围筛选条目。</p>
          </div>
        </section>

        <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card e-items-panel">
          <div class="card-body">
            <div class="next02-sc-panel-head standard-filter-panel-head">
              <div
                class="rank-stats-toolbar-scope rank-page-shell cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell standard-filter-toolbar e-items-filter-toolbar"
              >
                <form
                  class="sticky-toolbar cards-toolbar rank-sticky-toolbar next02-pink-cards-toolbar rank-stats-toolbar standard-filter-toolbar-form standard-filter-toolbar-form-two-row e-items-filter-form photo-results-toolbar"
                  method="get"
                  action="/eitemlist"
                >
                  <input type="hidden" name="p" value="1" />
                  <input v-if="queryValue('PageSize')" type="hidden" name="ps" :value="queryValue('PageSize')" />
                  <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')" />
                  <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')" />

                  <div class="standard-filter-toolbar-row standard-filter-toolbar-row-primary">
                    <div class="cards-toolbar-sort rank-toolbar-sort standard-filter-toolbar-fields sort-controls">
                      <div class="rank-toolbar-group standard-filter-group e-items-filter-identity-group">
                        <label
                          class="rank-stats-filter-field standard-filter-field sc-event-filter-field e-items-filter-field is-jav-id"
                        >
                          <span>JavLib 编号</span>
                          <input
                            class="rank-date-input standard-filter-input"
                            type="text"
                            name="jav_id"
                            autocomplete="off"
                            :value="queryValue('JavID')"
                            placeholder="JavLib 编号"
                          />
                        </label>
                        <label
                          class="rank-stats-filter-field standard-filter-field sc-event-filter-field e-items-filter-field is-name"
                        >
                          <span>影片编号</span>
                          <input
                            class="rank-date-input standard-filter-input"
                            type="text"
                            name="name"
                            autocomplete="off"
                            :value="queryValue('Name')"
                            placeholder="影片编号"
                          />
                        </label>
                      </div>

                      <div class="rank-toolbar-group standard-filter-group e-items-filter-state-group">
                        <label
                          class="rank-stats-filter-field standard-filter-field sc-event-filter-field e-items-filter-field"
                        >
                          <span>详情状态</span>
                          <select
                            class="rank-date-input standard-filter-input"
                            name="has_detail"
                            :value="queryValue('HasDetail')"
                          >
                            <option value="">全部</option>
                            <option value="none">无详情</option>
                            <option value="ok">已有详情</option>
                          </select>
                        </label>
                        <label
                          class="rank-stats-filter-field standard-filter-field sc-event-filter-field e-items-filter-field is-cover"
                        >
                          <span>封面状态</span>
                          <select
                            class="rank-date-input standard-filter-input"
                            name="has_download_cover"
                            :value="queryValue('HasDownloadCover')"
                          >
                            <option value="">全部</option>
                            <option value="none">无本地封面</option>
                            <option value="ok">已有本地封面</option>
                            <option value="wrong">错误封面链接</option>
                          </select>
                        </label>
                        <label
                          class="rank-stats-filter-field standard-filter-field sc-event-filter-field e-items-filter-field is-chinese"
                        >
                          <span>中文字幕</span>
                          <select
                            class="rank-date-input standard-filter-input"
                            name="has_chinese"
                            :value="queryValue('HasChinese')"
                          >
                            <option value="">全部</option>
                            <option value="none">无中文字幕</option>
                            <option value="ok">有中文字幕</option>
                            <option value="error">翻译错误</option>
                            <option value="sensitive">含敏感词</option>
                          </select>
                        </label>
                        <label
                          class="rank-stats-filter-field standard-filter-field sc-event-filter-field e-items-filter-field is-scan"
                        >
                          <span>详情扫描</span>
                          <select
                            class="rank-date-input standard-filter-input"
                            name="detail_need_scan"
                            :value="queryValue('DetailNeedScan')"
                          >
                            <option value="">全部</option>
                            <option value="need">需要扫描</option>
                            <option value="no_need">无需扫描</option>
                            <option value="wrong">内容错误</option>
                          </select>
                        </label>
                      </div>
                    </div>
                  </div>

                  <div class="standard-filter-toolbar-row standard-filter-toolbar-row-secondary">
                    <div class="cards-toolbar-sort rank-toolbar-sort standard-filter-toolbar-fields sort-controls">
                      <div class="rank-toolbar-group standard-filter-group e-items-filter-detail-date-group">
                        <label
                          class="rank-stats-filter-field standard-filter-field sc-event-filter-field e-items-filter-field"
                        >
                          <span>详情从</span>
                          <input
                            class="rank-date-input standard-filter-input standard-filter-date-input smart-pick-date-input"
                            type="text"
                            name="detail_birth_time_from"
                            :value="queryValue('DetailBirthTimeFrom')"
                            placeholder="YYYY-MM-DD"
                            inputmode="numeric"
                            maxlength="10"
                            pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                          />
                        </label>
                        <label
                          class="rank-stats-filter-field standard-filter-field sc-event-filter-field e-items-filter-field"
                        >
                          <span>详情到</span>
                          <input
                            class="rank-date-input standard-filter-input standard-filter-date-input smart-pick-date-input"
                            type="text"
                            name="detail_birth_time_to"
                            :value="queryValue('DetailBirthTimeTo')"
                            placeholder="YYYY-MM-DD"
                            inputmode="numeric"
                            maxlength="10"
                            pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                          />
                        </label>
                      </div>

                      <div class="rank-toolbar-group standard-filter-group e-items-filter-query-date-group">
                        <label
                          class="rank-stats-filter-field standard-filter-field sc-event-filter-field e-items-filter-field"
                        >
                          <span>查询从</span>
                          <input
                            class="rank-date-input standard-filter-input standard-filter-date-input smart-pick-date-input"
                            type="text"
                            name="last_query_detail_time_from"
                            :value="queryValue('LastQueryDetailTimeFrom')"
                            placeholder="YYYY-MM-DD"
                            inputmode="numeric"
                            maxlength="10"
                            pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                          />
                        </label>
                        <label
                          class="rank-stats-filter-field standard-filter-field sc-event-filter-field e-items-filter-field"
                        >
                          <span>查询到</span>
                          <input
                            class="rank-date-input standard-filter-input standard-filter-date-input smart-pick-date-input"
                            type="text"
                            name="last_query_detail_time_to"
                            :value="queryValue('LastQueryDetailTimeTo')"
                            placeholder="YYYY-MM-DD"
                            inputmode="numeric"
                            maxlength="10"
                            pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                          />
                        </label>
                      </div>
                    </div>

                    <div class="rank-navigation-group standard-filter-actions e-items-filter-actions">
                      <button class="button-secondary rank-toolbar-action" type="submit">筛选</button>
                      <a class="button-secondary rank-toolbar-action" href="/eitemlist">清空</a>
                    </div>

                    <div
                      v-if="hasRows"
                      class="toolbar-line cards-toolbar-bottom rank-toolbar-line standard-filter-status-line"
                    >
                      <div class="owned-line cards-toolbar-filters rank-toolbar-filters"></div>
                      <div class="next02-pink-cards-toolbar-status">
                        <div class="next02-pink-cards-metrics" aria-label="EItem 摘要">
                          <span :title="`总数 ${total}`" :aria-label="`总数 ${total}`"
                            ><strong id="itemTotalCount">{{ total }}</strong></span
                          >
                          <span :title="`已选 ${selectedCount}`" :aria-label="`已选 ${selectedCount}`"
                            ><strong id="itemBatchSelectedCount">{{ selectedCount }}</strong></span
                          >
                        </div>
                        <button
                          type="button"
                          class="button-secondary rank-toolbar-action standard-filter-secondary-action e-items-batch-action"
                          id="btnBatchEditStatus"
                          :disabled="selectedCount === 0"
                          @click="openBatchModal"
                        >
                          批量编辑状态
                        </button>
                        <div v-if="pageLinks.length" class="cards-toolbar-pagination next02-pink-cards-pagination">
                          <PaginationBar :links="pageLinks" />
                        </div>
                      </div>
                    </div>
                  </div>
                </form>
              </div>
            </div>

            <div class="table-responsive table-scroll">
              <table class="table table-hover align-middle mb-0 sc-events-table e-items-table data-table">
                <colgroup>
                  <col class="e-items-col-select" />
                  <col class="e-items-col-jav-id" />
                  <col class="e-items-col-name" />
                  <col class="e-items-col-detail" />
                  <col class="e-items-col-cover" />
                  <col class="e-items-col-chinese" />
                  <col class="e-items-col-scan" />
                  <col class="e-items-col-detail-time" />
                  <col class="e-items-col-query-time" />
                  <col class="e-items-col-actions" />
                </colgroup>
                <thead>
                  <tr>
                    <th class="e-items-select-head">
                      <input
                        class="e-items-check js-item-check-all"
                        id="itemCheckAll"
                        type="checkbox"
                        aria-label="全选当前页"
                        :checked="allSelected"
                        :indeterminate="someSelected"
                        @change="toggleAllFromEvent"
                      />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="sortLink('ByJavID')" label="JavLib 编号" />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="sortLink('ByName')" label="影片编号" />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="sortLink('ByHasDetail')" label="详情状态" />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="sortLink('ByHasDownloadCover')" label="封面状态" />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="sortLink('ByHasChinese')" label="中文字幕" />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="sortLink('ByDetailNeedScan')" label="详情扫描状态" />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="sortLink('ByDetailBirthTime')" label="详情时间" />
                    </th>
                    <th class="sc-sortable-head">
                      <SortLink class="sc-sort-link" :link="sortLink('ByLastQueryDetailTime')" label="最后查询时间" />
                    </th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody id="itemTableBody">
                  <template v-if="hasRows">
                    <tr
                      v-for="(row, index) in rows"
                      :key="rowId(row)"
                      :data-item-id="rowId(row)"
                      :class="eItemRowClass(row, index)"
                    >
                      <td>
                        <input
                          class="e-items-check js-item-check"
                          type="checkbox"
                          :value="rowId(row)"
                          :aria-label="'选择 ' + rowText(row, 'jav_id', 'JavId', 'JavID')"
                          :checked="isSelected(rowId(row))"
                          @change="toggleSelection(rowId(row), $event)"
                        />
                      </td>
                      <td class="sc-content-cell">
                        <a
                          v-if="rowText(row, 'jav_url', 'JavUrl', 'JavURL')"
                          :href="rowText(row, 'jav_url', 'JavUrl', 'JavURL')"
                          target="_blank"
                          rel="noopener noreferrer"
                          :title="rowText(row, 'jav_id', 'JavId', 'JavID')"
                          >{{ rowText(row, "jav_id", "JavId", "JavID") }}</a
                        >
                        <span v-else :title="rowText(row, 'jav_id', 'JavId', 'JavID')">{{
                          rowText(row, "jav_id", "JavId", "JavID")
                        }}</span>
                      </td>
                      <td class="sc-content-cell">
                        <a
                          :href="'/moviedetail/' + encodeURIComponent(rowText(row, 'name', 'Name'))"
                          target="_blank"
                          rel="noopener noreferrer"
                          :title="rowText(row, 'name', 'Name')"
                          >{{ rowText(row, "name", "Name") }}</a
                        >
                      </td>
                      <td>
                        <span
                          class="e-items-status-text"
                          :title="rowText(row, 'has_detail_text', 'HasDetailText') || '-'"
                          >{{ rowText(row, "has_detail_text", "HasDetailText") || "-" }}</span
                        >
                      </td>
                      <td>
                        <span
                          class="e-items-status-text"
                          :title="rowText(row, 'has_download_cover_text', 'HasDownloadCoverText') || '-'"
                          >{{ rowText(row, "has_download_cover_text", "HasDownloadCoverText") || "-" }}</span
                        >
                      </td>
                      <td>
                        <span
                          class="e-items-status-text"
                          :title="rowText(row, 'has_chinese_text', 'HasChineseText') || '-'"
                          >{{ rowText(row, "has_chinese_text", "HasChineseText") || "-" }}</span
                        >
                      </td>
                      <td>
                        <span
                          class="e-items-status-text"
                          :title="rowText(row, 'detail_need_scan_text', 'DetailNeedScanText') || '-'"
                          >{{ rowText(row, "detail_need_scan_text", "DetailNeedScanText") || "-" }}</span
                        >
                      </td>
                      <td>
                        <span class="sc-event-time"
                          ><span class="sc-event-date">{{
                            formatUnixDate(rowValue(row, "detail_birth_time", "DetailBirthTime"))
                          }}</span
                          ><small
                            v-if="formatUnixHm(rowValue(row, 'detail_birth_time', 'DetailBirthTime'))"
                            class="sc-event-hour"
                            >{{ formatUnixHm(rowValue(row, "detail_birth_time", "DetailBirthTime")) }}</small
                          ></span
                        >
                      </td>
                      <td>
                        <span class="sc-event-time"
                          ><span class="sc-event-date">{{
                            formatUnixDate(rowValue(row, "last_query_detail_time", "LastQueryDetailTime"))
                          }}</span
                          ><small
                            v-if="formatUnixHm(rowValue(row, 'last_query_detail_time', 'LastQueryDetailTime'))"
                            class="sc-event-hour"
                            >{{ formatUnixHm(rowValue(row, "last_query_detail_time", "LastQueryDetailTime")) }}</small
                          ></span
                        >
                      </td>
                      <td>
                        <div class="e-items-row-actions">
                          <button
                            type="button"
                            class="e-items-row-action js-item-status-edit"
                            :data-item-id="rowId(row)"
                            @click="openSingleModal(row)"
                          >
                            编辑状态
                          </button>
                          <button
                            type="button"
                            class="e-items-row-action is-danger js-item-delete"
                            :data-item-id="rowId(row)"
                            @click="openDeleteModal(row)"
                          >
                            删除
                          </button>
                        </div>
                      </td>
                    </tr>
                  </template>
                  <tr v-else class="js-item-empty-row">
                    <td colspan="10"><div class="sc-empty-state">当前筛选下没有 e_item 数据。</div></td>
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
          :submitting="singleModal.submitting"
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
          :submitting="batchModal.submitting"
          @close="closeBatchModal"
          @submit="submitBatchStatus"
          @update-form="updateBatchForm"
        />
        <div
          class="e-items-modal-overlay"
          id="itemDeleteModal"
          tabindex="-1"
          role="dialog"
          aria-modal="true"
          aria-labelledby="itemDeleteModalLabel"
          :aria-hidden="deleteModal.show ? 'false' : 'true'"
          :class="{ show: deleteModal.show }"
        >
          <div class="e-items-modal-dialog e-items-delete-dialog">
            <div class="e-items-modal-content">
              <header class="e-items-modal-header">
                <div>
                  <span class="e-items-modal-eyebrow">E ITEMS / DELETE</span>
                  <h5 id="itemDeleteModalLabel">删除电影</h5>
                  <p id="itemDeleteMeta">{{ deleteModal.meta }}</p>
                </div>
                <button type="button" class="e-items-modal-close" aria-label="关闭" @click="closeDeleteModal">
                  <span aria-hidden="true">×</span>
                </button>
              </header>
              <div class="e-items-modal-body">
                <p class="e-items-delete-warning">会删除这部电影在所有相关表中的数据，并写入删除记录。</p>
                <div
                  id="itemDeleteMsg"
                  class="e-items-modal-message"
                  :class="[deleteModal.message ? (deleteModal.messageOk ? 'is-success' : 'is-error') : 'd-none']"
                >
                  {{ deleteModal.message }}
                </div>
                <input type="hidden" id="itemDeleteItemId" :value="deleteModal.itemId" />
              </div>
              <footer class="e-items-modal-footer">
                <button type="button" class="e-items-modal-cancel" @click="closeDeleteModal">取消</button>
                <button
                  type="button"
                  class="e-items-modal-submit is-danger"
                  id="itemDeleteSubmit"
                  :disabled="deleteModal.submitting"
                  @click="submitDelete"
                >
                  {{ deleteModal.submitting ? "删除中..." : "确认删除" }}
                </button>
              </footer>
            </div>
          </div>
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/features/crawler/api";
import ItemStatusModal from "@/features/crawler/components/e-items/ItemStatusModal.vue";
import type { EItemStatusForm as StatusForm } from "@/features/crawler/components/e-items/models";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import type { PageInfo } from "@/shared/types/gc";
import type { CrawlerRecord } from "@/features/crawler/types";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<CrawlerRecord>({});
const rows = ref<CrawlerRecord[]>([]);
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
    rows.value = asArray<CrawlerRecord>(pick(data, "Items", "items", "Rows", "rows"));
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
const pageLinks = computed(() => asArray<CrawlerRecord>(pageInfo.value.Links || pageInfo.value.links));
const sortMap = computed(() => asRecord(pick(payload.value, "ItemSortQuery", "itemSortQuery")));
const selectedCount = computed(() => selectedIds.value.length);
const allSelected = computed(() => rows.value.length > 0 && selectedIds.value.length === rows.value.length);
const someSelected = computed(() => selectedIds.value.length > 0 && !allSelected.value);

function queryValue(key: string): string {
  return text(query.value, key);
}

function rowValue(row: CrawlerRecord, ...keys: string[]): unknown {
  return pick(row, ...keys);
}

function rowText(row: CrawlerRecord, ...keys: string[]): string {
  return text(row, ...keys);
}

function rowId(row: CrawlerRecord): string {
  return rowText(row, "id", "Id");
}

function sortLink(key: string): CrawlerRecord {
  return asRecord(sortMap.value[key]);
}

function dateParts(input: unknown): { date: string; time: string } {
  const n = Number(input || 0);
  if (!n) return { date: "-", time: "" };
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  const hh = String(d.getHours()).padStart(2, "0");
  const mi = String(d.getMinutes()).padStart(2, "0");
  return { date: `${yyyy}-${mm}-${dd}`, time: `${hh}:${mi}` };
}

function formatUnixDate(input: unknown): string {
  return dateParts(input).date;
}

function formatUnixHm(input: unknown): string {
  return dateParts(input).time;
}

function eItemGroupTime(row: CrawlerRecord): unknown {
  return (
    rowValue(row, "last_query_detail_time", "LastQueryDetailTime") ||
    rowValue(row, "detail_birth_time", "DetailBirthTime")
  );
}

function eItemMonthKey(row: CrawlerRecord): string {
  const date = formatUnixDate(eItemGroupTime(row));
  return date === "-" ? "" : date.slice(0, 7);
}

function eItemMonthClass(row: CrawlerRecord): string {
  const month = eItemMonthKey(row);
  if (!month) return "sc-running-month-neutral";
  const monthNumber = Number(month.slice(5, 7));
  return `sc-running-month-${(monthNumber % 3) + 1}`;
}

function eItemRowClass(row: CrawlerRecord, index: number) {
  const currentMonth = eItemMonthKey(row);
  const previousMonth = index > 0 ? eItemMonthKey(rows.value[index - 1]) : "";
  return [
    "next02-sc-event-row",
    eItemMonthClass(row),
    {
      "sc-running-month-start": Boolean(index > 0 && currentMonth && previousMonth && currentMonth !== previousMonth),
      "is-selected": isSelected(rowId(row))
    }
  ];
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

function formFromRow(row: CrawlerRecord): StatusForm {
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

function openSingleModal(row: CrawlerRecord) {
  singleModal.itemId = rowId(row);
  singleModal.meta = `${rowText(row, "jav_id", "JavId", "JavID") || "-"} / ${rowText(row, "name", "Name") || "-"}`;
  singleModal.message = "";
  singleModal.messageOk = false;
  singleModal.form = formFromRow(row);
  singleModal.show = true;
}

function closeSingleModal() {
  releaseModalFocus();
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
  releaseModalFocus();
  batchModal.show = false;
  batchModal.message = "";
  batchModal.submitting = false;
}

function openDeleteModal(row: CrawlerRecord) {
  deleteModal.itemId = rowId(row);
  deleteModal.meta = `${rowText(row, "jav_id", "JavId", "JavID") || "-"} / ${rowText(row, "name", "Name") || "-"}`;
  deleteModal.message = "";
  deleteModal.messageOk = false;
  deleteModal.show = true;
}

function closeDeleteModal() {
  releaseModalFocus();
  deleteModal.show = false;
  deleteModal.message = "";
  deleteModal.submitting = false;
}

function releaseModalFocus() {
  const activeElement = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  if (activeElement?.closest(".e-items-modal-overlay")) {
    activeElement.blur();
  }
}

function mergeUpdatedItem(item: unknown) {
  const next = asRecord(item);
  const id = text(next, "id", "Id");
  if (!id) return;
  rows.value = rows.value.map((row) => (rowId(row) === id ? { ...row, ...next } : row));
}

async function submitSingleStatus() {
  if (!singleModal.itemId) {
    singleModal.message = "缺少条目 id";
    singleModal.messageOk = false;
    return;
  }
  singleModal.submitting = true;
  try {
    const data = asRecord(
      await sendJSON(`/api/e-items/${encodeURIComponent(singleModal.itemId)}/status`, "POST", singleModal.form)
    );
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
  const payloadData: CrawlerRecord = {
    ids: selectedIds.value.map((id) => Number(id)).filter((id) => Number.isFinite(id) && id > 0)
  };
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
    asArray<CrawlerRecord>(data.items).forEach(mergeUpdatedItem);
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
</script>
