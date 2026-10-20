<template>
  <PageShell class="torrent-albums-page-shell d-seeds-page-shell" title="DSeeds" note="" eyebrow="">
    <div id="page-root" class="container-fluid torrent-albums-page-root d-seeds-page-root d-seeds-list-standard">
      <div v-if="loading" class="alert d-seeds-feedback is-loading">加载中...</div>
      <div v-else-if="error" class="alert d-seeds-feedback is-error">{{ error }}</div>
      <div v-else id="dSeedListPage" class="torrent-albums-page d-seeds-page">
        <header class="d-seeds-list-hero">
          <div class="d-seeds-list-hero-copy">
            <span class="d-seeds-list-eyebrow">DSeed</span>
            <h2>{{ pageTitle }}</h2>
            <p v-if="pageNote">{{ pageNote }}</p>
          </div>
          <button type="button" class="btn d-seeds-create-button" id="btnOpenDSeedCreateModal" @click="openCreateModal">
            新增条目
          </button>
        </header>

        <div id="dSeedMsg" class="alert d-seeds-message" :class="messageClass" role="alert">
          {{ messageText }}
        </div>

        <section class="d-seeds-list-panel">
          <div class="d-seeds-list-controls">
            <div
              class="rank-stats-toolbar-scope rank-page-shell cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell standard-filter-toolbar d-seeds-filter-toolbar"
            >
              <form
                class="sticky-toolbar cards-toolbar rank-sticky-toolbar next02-pink-cards-toolbar rank-stats-toolbar standard-filter-toolbar-form d-seeds-filter-row"
                id="dSeedFilterForm"
                method="get"
                action="/seedlist"
              >
                <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')" />
                <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')" />

                <div class="cards-toolbar-sort rank-toolbar-sort standard-filter-toolbar-fields">
                  <div class="rank-toolbar-group standard-filter-group d-seeds-filter-query-group">
                    <label class="rank-stats-filter-field standard-filter-field d-seeds-filter-field is-name">
                      <span>查询名</span>
                      <input
                        class="rank-date-input standard-filter-input form-control"
                        type="text"
                        name="name"
                        autocomplete="off"
                        :value="queryValue('Name')"
                        placeholder="支持模糊搜索"
                      />
                    </label>
                  </div>

                  <div class="rank-toolbar-group standard-filter-group d-seeds-filter-state-group">
                    <label class="rank-stats-filter-field standard-filter-field d-seeds-filter-field is-select">
                      <span>启用状态</span>
                      <select
                        class="rank-date-input standard-filter-input form-select"
                        name="active"
                        :value="queryValue('Active')"
                      >
                        <option value="">全部</option>
                        <option value="2">启用</option>
                        <option value="1">停用</option>
                      </select>
                    </label>
                    <label class="rank-stats-filter-field standard-filter-field d-seeds-filter-field is-select">
                      <span>名称类型</span>
                      <select
                        class="rank-date-input standard-filter-input form-select"
                        name="name_type"
                        :value="queryValue('NameType')"
                      >
                        <option value="">全部</option>
                        <option value="1">Prefix</option>
                        <option value="2">Label</option>
                      </select>
                    </label>
                    <label class="rank-stats-filter-field standard-filter-field d-seeds-filter-field is-select">
                      <span>查询类型</span>
                      <select
                        class="rank-date-input standard-filter-input form-select"
                        name="search_type"
                        :value="queryValue('SearchType')"
                      >
                        <option value="">全部</option>
                        <option value="1">偏移量</option>
                        <option value="2">起止页</option>
                      </select>
                    </label>
                    <label class="rank-stats-filter-field standard-filter-field d-seeds-filter-field is-select">
                      <span>最后状态</span>
                      <select
                        class="rank-date-input standard-filter-input form-select"
                        name="last_status"
                        :value="queryValue('LastStatus')"
                      >
                        <option value="">全部</option>
                        <option value="0">空闲</option>
                        <option value="1">成功</option>
                        <option value="2">空页</option>
                        <option value="3">错误</option>
                      </select>
                    </label>
                  </div>

                  <div class="rank-toolbar-group standard-filter-group d-seeds-filter-page-group">
                    <label class="rank-stats-filter-field standard-filter-field d-seeds-filter-field is-page-size">
                      <span>每页</span>
                      <input
                        class="rank-date-input standard-filter-input standard-filter-page-size-input form-control"
                        type="number"
                        step="1"
                        name="ps"
                        :value="String(queryValue('PageSize') || 50)"
                      />
                    </label>
                  </div>
                </div>

                <div class="rank-navigation-group standard-filter-actions d-seeds-filter-actions">
                  <button class="button-secondary rank-toolbar-action" type="submit">筛选</button>
                  <a class="button-secondary rank-toolbar-action" href="/seedlist">清空</a>
                </div>

                <div
                  v-if="hasRows"
                  class="toolbar-line cards-toolbar-bottom rank-toolbar-line standard-filter-status-line"
                >
                  <div class="owned-line cards-toolbar-filters rank-toolbar-filters"></div>
                  <div class="next02-pink-cards-toolbar-status">
                    <div class="next02-pink-cards-metrics" aria-label="种子查询摘要">
                      <span :title="`总数 ${total}`" :aria-label="`总数 ${total}`"
                        ><strong>{{ total }}</strong></span
                      >
                      <span :title="`当前页 ${rows.length}`" :aria-label="`当前页 ${rows.length}`"
                        ><strong>{{ rows.length }}</strong></span
                      >
                    </div>
                    <div v-if="pageLinks.length" class="cards-toolbar-pagination next02-pink-cards-pagination">
                      <PaginationBar :links="pageLinks" />
                    </div>
                  </div>
                </div>
              </form>
            </div>
          </div>

          <div class="d-seeds-table-wrap">
            <table class="table table-hover align-middle mb-0 d-seed-table">
              <colgroup>
                <col class="d-seed-col-name" />
                <col class="d-seed-col-active" />
                <col class="d-seed-col-name-type" />
                <col class="d-seed-col-search-type" />
                <col class="d-seed-col-movie-total" />
                <col class="d-seed-col-latest-movie" />
                <col class="d-seed-col-release-date" />
                <col class="d-seed-col-last-added" />
                <col class="d-seed-col-insert-count" />
                <col class="d-seed-col-page-now" />
                <col class="d-seed-col-breakpoint" />
                <col class="d-seed-col-last-query" />
                <col class="d-seed-col-actions" />
              </colgroup>
              <thead>
                <tr>
                  <th><SortLink class="d-seed-sort-link" :link="sortLink('ByName')" label="查询名" /></th>
                  <th><SortLink class="d-seed-sort-link" :link="sortLink('ByActive')" label="启用状态" /></th>
                  <th><SortLink class="d-seed-sort-link" :link="sortLink('ByNameType')" label="名称类型" /></th>
                  <th><SortLink class="d-seed-sort-link" :link="sortLink('BySearchType')" label="查询类型" /></th>
                  <th class="text-nowrap">
                    <SortLink class="d-seed-sort-link" :link="sortLink('ByMovieTotal')" label="电影总量" />
                  </th>
                  <th class="text-nowrap">
                    <SortLink
                      class="d-seed-sort-link"
                      :link="sortLink('ByMovieLatestReleasingMovieName')"
                      label="最新上映电影"
                    />
                  </th>
                  <th class="text-nowrap">
                    <SortLink
                      class="d-seed-sort-link"
                      :link="sortLink('ByMovieLatestReleasingDate')"
                      label="最新上映日"
                    />
                  </th>
                  <th class="text-nowrap">
                    <SortLink class="d-seed-sort-link" :link="sortLink('ByMovieLastAddedTime')" label="上次新增" />
                  </th>
                  <th class="text-nowrap">
                    <SortLink class="d-seed-sort-link" :link="sortLink('ByLastInsertCount')" label="本轮新增" />
                  </th>
                  <th class="text-nowrap">
                    <SortLink class="d-seed-sort-link" :link="sortLink('ByPageNow')" label="当前页" />
                  </th>
                  <th class="text-nowrap">断点</th>
                  <th class="text-nowrap">
                    <SortLink class="d-seed-sort-link" :link="sortLink('ByLastQueryTime')" label="最后抓取" />
                  </th>
                  <th class="text-nowrap">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="!hasRows">
                  <td colspan="13"><div class="d-seed-empty-state">当前筛选下没有 d_seed 数据。</div></td>
                </tr>
                <tr v-for="row in rows" v-else :key="rowKey(row)">
                  <td class="text-nowrap" :title="value(row, 'Name')">
                    <a
                      v-if="value(row, 'CardsHref')"
                      class="d-seed-entry-link"
                      :href="value(row, 'CardsHref')"
                      target="_blank"
                      rel="noopener noreferrer"
                      >{{ value(row, "Name") }}</a
                    >
                    <span v-else>{{ value(row, "Name") }}</span>
                  </td>
                  <td>
                    <span class="badge" :class="value(row, 'ActiveClass')">{{ value(row, "ActiveText") }}</span>
                  </td>
                  <td class="text-nowrap">{{ value(row, "NameTypeText") }}</td>
                  <td class="text-nowrap">{{ value(row, "SearchTypeText") }}</td>
                  <td class="text-nowrap">{{ value(row, "MovieTotal") }}</td>
                  <td class="text-nowrap" :title="value(row, 'MovieLatestReleasingMovieName')">
                    <a
                      v-if="value(row, 'MovieLatestReleasingMovieHref')"
                      class="d-seed-entry-link"
                      :href="value(row, 'MovieLatestReleasingMovieHref')"
                      target="_blank"
                      rel="noopener noreferrer"
                      >{{ value(row, "MovieLatestReleasingMovieName") }}</a
                    >
                    <span v-else>-</span>
                  </td>
                  <td class="text-nowrap">{{ formatUnixDate(value(row, "MovieLatestReleasingDate")) }}</td>
                  <td class="text-nowrap"><ScDateTime :value="value(row, 'MovieLastAddedTime')" unix /></td>
                  <td class="text-nowrap">{{ value(row, "LastInsertCount") }}</td>
                  <td class="text-nowrap">{{ value(row, "PageNow") }}</td>
                  <td class="text-nowrap">
                    {{ value(row, "Offset") }} / {{ value(row, "StartPage") }}-{{ value(row, "EndPage") }}
                  </td>
                  <td class="text-nowrap"><ScDateTime :value="value(row, 'LastQueryTime')" unix /></td>
                  <td class="text-nowrap">
                    <button
                      type="button"
                      class="btn btn-sm btn-outline-primary js-d-seed-edit"
                      :data-item-id="value(row, 'Id')"
                      :data-name="value(row, 'Name')"
                      :data-active="value(row, 'ActiveValue')"
                      :data-search-type="value(row, 'SearchTypeValue')"
                      :data-name-type="value(row, 'NameTypeValue')"
                      :data-page-now="value(row, 'PageNow')"
                      :data-offset="value(row, 'Offset')"
                      :data-start-page="value(row, 'StartPage')"
                      :data-end-page="value(row, 'EndPage')"
                      :data-last-status="value(row, 'LastStatusValue')"
                      :data-last-error="value(row, 'LastError')"
                      @click="openEditModal(row)"
                    >
                      编辑
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <DSeedEditModal
          :active="activeModal"
          :model="editModel"
          :submitting="submitting"
          @close="closeModal"
          @update="updateEditModel"
          @submit="submitDSeed"
        />
      </div>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import "@/features/crawler/styles/index.css";
import { computed, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/features/crawler/api";
import DSeedEditModal from "@/features/crawler/components/d-seeds/DSeedEditModal.vue";
import type { DSeedEditModel as EditModel } from "@/features/crawler/components/d-seeds/models";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import type { PageInfo } from "@/shared/types/gc";
import type { CrawlerRecord } from "@/features/crawler/types";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

const emptyModel = (): EditModel => ({
  id: "",
  name: "",
  active: "2",
  searchType: "1",
  nameType: "1",
  pageNow: "0",
  offset: "0",
  startPage: "0",
  endPage: "0",
  lastStatus: "0",
  lastError: ""
});

const route = useRoute();
const payload = ref<CrawlerRecord>({});
const rows = ref<CrawlerRecord[]>([]);
const loading = ref(false);
const error = ref("");
const messageText = ref("");
const messageOk = ref(true);
const activeModal = ref(false);
const submitting = ref(false);
const editModel = reactive<EditModel>(emptyModel());

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  messageText.value = "";
  try {
    const data = asRecord(await getJSON(apiPathForRoute(route)));
    payload.value = data;
    rows.value = asArray<CrawlerRecord>(pick(data, "Items", "items", "Rows", "rows"));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title"));
const pageNote = computed(() => text(payload.value, "PageNote"));
const total = computed(() => text(payload.value, "Total", "total") || "0");
const query = computed(() => asRecord(pick(payload.value, "Query")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<CrawlerRecord>(pageInfo.value.Links));
const sortQuery = computed(() => asRecord(pick(payload.value, "DSeedSort")));
const hasRows = computed(() => rows.value.length > 0);
const messageClass = computed(() =>
  messageText.value ? (messageOk.value ? "alert-success small py-2 px-3" : "alert-danger small py-2 px-3") : "d-none"
);

function queryValue(key: string): string {
  return text(query.value, key);
}

function value(row: CrawlerRecord, key: string): string {
  return text(row, key);
}

function rowKey(row: CrawlerRecord): string {
  return [value(row, "Id"), value(row, "Name")].join(":");
}

function sortLink(key: string): CrawlerRecord {
  return asRecord(sortQuery.value[key]);
}

function formatUnixDate(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function assignModel(next: Partial<EditModel>) {
  Object.assign(editModel, emptyModel(), next);
}

function showMessage(message: string, ok = true) {
  messageText.value = message;
  messageOk.value = ok;
}

function openCreateModal() {
  assignModel({});
  activeModal.value = true;
}

function openEditModal(row: CrawlerRecord) {
  assignModel({
    id: value(row, "Id"),
    name: value(row, "Name"),
    active: value(row, "ActiveValue"),
    searchType: value(row, "SearchTypeValue"),
    nameType: value(row, "NameTypeValue"),
    pageNow: value(row, "PageNow"),
    offset: value(row, "Offset"),
    startPage: value(row, "StartPage"),
    endPage: value(row, "EndPage"),
    lastStatus: value(row, "LastStatusValue"),
    lastError: value(row, "LastError")
  });
  activeModal.value = true;
}

function closeModal() {
  activeModal.value = false;
  submitting.value = false;
}

function updateEditModel(patch: Partial<EditModel>) {
  Object.assign(editModel, patch);
}

async function submitDSeed() {
  const isOffset = String(editModel.searchType || "1") === "1";
  const body = {
    name: editModel.name.trim(),
    active: Number(editModel.active || 0),
    search_type: Number(editModel.searchType || 0),
    name_type: Number(editModel.nameType || 0),
    page_now: Number(editModel.pageNow || 0),
    offset: isOffset ? Number(editModel.offset || 0) : 0,
    start_page: isOffset ? 0 : Number(editModel.startPage || 0),
    end_page: isOffset ? 0 : Number(editModel.endPage || 0),
    last_status: Number(editModel.lastStatus || 0),
    last_error: editModel.lastError.trim()
  };
  const id = editModel.id.trim();
  submitting.value = true;
  try {
    const data = asRecord(await sendJSON(id ? `/api/d-seeds/${encodeURIComponent(id)}` : "/api/d-seeds", "POST", body));
    showMessage(text(data, "message") || "保存成功");
    closeModal();
    window.setTimeout(() => window.location.reload(), 250);
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "保存失败", false);
  } finally {
    submitting.value = false;
  }
}
</script>
