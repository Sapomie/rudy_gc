<template>
  <PageShell
    class="next02-pink-list-page-shell sc-events-page-shell casts-page-shell"
    :title="pageTitle"
    note=""
    eyebrow=""
  >
    <div
      id="page-root"
      class="container-fluid next02-pink-list-page next02-sc-pink-theme next02-sc-events-rewrite casts-page-root"
    >
      <div v-if="loading" class="alert alert-info py-2 next02-pink-list-state">加载中...</div>
      <div v-else-if="error" class="alert alert-danger next02-pink-list-state">{{ error }}</div>
      <template v-else>
        <section class="next02-pink-list-hero casts-list-hero">
          <div class="next02-pink-list-hero-copy">
            <span class="next02-pink-list-eyebrow">CAST / DIRECTORY</span>
            <h4>{{ pageTitle }}</h4>
            <p>浏览演员统计，并进入详情页查看该演员的 SC 历史和代表影片。</p>
          </div>
        </section>

        <div
          class="card next02-pink-list-panel next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card casts-list-panel"
        >
          <div class="card-body">
            <div
              class="rank-stats-toolbar-scope rank-page-shell cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell standard-filter-toolbar casts-filter-toolbar"
            >
              <form
                class="sticky-toolbar cards-toolbar rank-sticky-toolbar next02-pink-cards-toolbar rank-stats-toolbar standard-filter-toolbar-form standard-filter-toolbar-form-two-row casts-filter-form"
                method="get"
                action="/castlist"
                id="castsFilterForm"
              >
                <input type="hidden" name="p" value="1" />
                <input v-if="queryValue('PageSize')" type="hidden" name="ps" :value="queryValue('PageSize')" />
                <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')" />
                <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')" />

                <div class="standard-filter-toolbar-row standard-filter-toolbar-row-primary">
                  <div class="cards-toolbar-sort rank-toolbar-sort standard-filter-toolbar-fields">
                    <div class="rank-toolbar-group standard-filter-group casts-filter-identity-group">
                      <label
                        class="rank-stats-filter-field standard-filter-field casts-filter-field casts-filter-keyword"
                        for="casts-keyword"
                      >
                        <span>Name</span>
                        <input
                          id="casts-keyword"
                          class="rank-date-input standard-filter-input"
                          type="text"
                          name="keyword"
                          placeholder="Name / Chinese / Alias"
                          :value="queryValue('Keyword')"
                        />
                      </label>
                    </div>

                    <div class="rank-toolbar-group standard-filter-group casts-filter-range-group">
                      <div class="rank-stats-filter-field standard-filter-field casts-filter-field">
                        <span>WOwned</span>
                        <span class="casts-filter-range standard-filter-range">
                          <input
                            id="casts-owned-min"
                            class="rank-date-input standard-filter-input standard-filter-range-input"
                            type="number"
                            min="0"
                            name="wowned_min"
                            placeholder="Min"
                            aria-label="WOwned minimum"
                            :value="queryValue('OwnedMin')"
                          />
                          <span aria-hidden="true">–</span>
                          <input
                            id="casts-owned-max"
                            class="rank-date-input standard-filter-input standard-filter-range-input"
                            type="number"
                            min="0"
                            name="wowned_max"
                            placeholder="Max"
                            aria-label="WOwned maximum"
                            :value="queryValue('OwnedMax')"
                          />
                        </span>
                      </div>
                      <div class="rank-stats-filter-field standard-filter-field casts-filter-field">
                        <span>SC Times</span>
                        <span class="casts-filter-range standard-filter-range">
                          <input
                            id="casts-sc-times-min"
                            class="rank-date-input standard-filter-input standard-filter-range-input"
                            type="number"
                            min="0"
                            name="sc_times_min"
                            placeholder="Min"
                            aria-label="SC Times minimum"
                            :value="queryValue('ScTimesMin')"
                          />
                          <span aria-hidden="true">–</span>
                          <input
                            id="casts-sc-times-max"
                            class="rank-date-input standard-filter-input standard-filter-range-input"
                            type="number"
                            min="0"
                            name="sc_times_max"
                            placeholder="Max"
                            aria-label="SC Times maximum"
                            :value="queryValue('ScTimesMax')"
                          />
                        </span>
                      </div>
                      <div class="rank-stats-filter-field standard-filter-field casts-filter-field">
                        <span>Come Times</span>
                        <span class="casts-filter-range standard-filter-range">
                          <input
                            id="casts-come-times-min"
                            class="rank-date-input standard-filter-input standard-filter-range-input"
                            type="number"
                            min="0"
                            name="come_times_min"
                            placeholder="Min"
                            aria-label="Come Times minimum"
                            :value="queryValue('ComeTimesMin')"
                          />
                          <span aria-hidden="true">–</span>
                          <input
                            id="casts-come-times-max"
                            class="rank-date-input standard-filter-input standard-filter-range-input"
                            type="number"
                            min="0"
                            name="come_times_max"
                            placeholder="Max"
                            aria-label="Come Times maximum"
                            :value="queryValue('ComeTimesMax')"
                          />
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="standard-filter-toolbar-row standard-filter-toolbar-row-secondary">
                  <div class="cards-toolbar-sort rank-toolbar-sort standard-filter-toolbar-fields">
                    <div class="rank-toolbar-group standard-filter-group casts-filter-date-group">
                      <div class="rank-stats-filter-field standard-filter-field casts-filter-field casts-filter-date">
                        <span>最近入选 SC</span>
                        <span class="casts-filter-range standard-filter-range">
                          <input
                            id="casts-last-selected-sc-from"
                            class="rank-date-input standard-filter-input standard-filter-date-input smart-pick-date-input"
                            type="text"
                            name="last_selected_sc_from"
                            aria-label="最近入选 SC 起始日期"
                            :value="queryValue('LastSelectedScFrom')"
                            placeholder="YYYY-MM-DD"
                            inputmode="numeric"
                            maxlength="10"
                            pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                          />
                          <span aria-hidden="true">–</span>
                          <input
                            id="casts-last-selected-sc-to"
                            class="rank-date-input standard-filter-input standard-filter-date-input smart-pick-date-input"
                            type="text"
                            name="last_selected_sc_to"
                            aria-label="最近入选 SC 结束日期"
                            :value="queryValue('LastSelectedScTo')"
                            placeholder="YYYY-MM-DD"
                            inputmode="numeric"
                            maxlength="10"
                            pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                          />
                        </span>
                      </div>
                    </div>
                  </div>

                  <div class="rank-navigation-group standard-filter-actions casts-filter-actions">
                    <button class="button-secondary rank-toolbar-action" type="submit">筛选</button>
                    <a class="button-secondary rank-toolbar-action" href="/castlist">清空</a>
                  </div>

                  <div
                    v-if="hasRows"
                    class="toolbar-line cards-toolbar-bottom rank-toolbar-line standard-filter-status-line"
                  >
                    <div class="owned-line cards-toolbar-filters rank-toolbar-filters"></div>
                    <div class="next02-pink-cards-toolbar-status">
                      <div class="next02-pink-cards-metrics" aria-label="演员列表统计">
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
                </div>
              </form>
            </div>

            <div class="table-responsive next02-pink-list-table-wrap">
              <table class="table table-hover align-middle mb-0 next02-pink-list-table sc-events-table casts-table">
                <colgroup>
                  <col class="casts-col-name" />
                  <col class="casts-col-chinese" />
                  <col class="casts-col-age" />
                  <col class="casts-col-height" />
                  <col class="casts-col-movies" />
                  <col class="casts-col-owned" />
                  <col class="casts-col-sc-times" />
                  <col class="casts-col-come-times" />
                  <col class="casts-col-last-selected-sc" />
                  <col class="casts-col-highest-rank" />
                  <col class="casts-col-actions" />
                </colgroup>
                <thead>
                  <tr>
                    <th class="next02-pink-list-sortable sc-sortable-head">
                      <SortLink
                        class="next02-pink-list-sort-link sc-sort-link"
                        :link="sortLink('ByName')"
                        label="Name"
                      />
                    </th>
                    <th class="next02-pink-list-sortable sc-sortable-head">
                      <SortLink
                        class="next02-pink-list-sort-link sc-sort-link"
                        :link="sortLink('ByChinese')"
                        label="Chinese"
                      />
                    </th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number">
                      <SortLink class="next02-pink-list-sort-link sc-sort-link" :link="sortLink('ByAge')" label="Age" />
                    </th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number">
                      <SortLink
                        class="next02-pink-list-sort-link sc-sort-link"
                        :link="sortLink('ByHeight')"
                        label="Height"
                      />
                    </th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number">
                      <SortLink
                        class="next02-pink-list-sort-link sc-sort-link"
                        :link="sortLink('ByMovieNumber')"
                        label="Movies"
                      />
                    </th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number">
                      <SortLink
                        class="next02-pink-list-sort-link sc-sort-link"
                        :link="sortLink('ByWOwned')"
                        label="WOwned"
                      />
                    </th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number">
                      <SortLink
                        class="next02-pink-list-sort-link sc-sort-link"
                        :link="sortLink('ByScTimes')"
                        label="SC Times"
                      />
                    </th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number">
                      <SortLink
                        class="next02-pink-list-sort-link sc-sort-link"
                        :link="sortLink('ByComeTimes')"
                        label="Come Times"
                      />
                    </th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number">
                      <SortLink
                        class="next02-pink-list-sort-link sc-sort-link"
                        :link="sortLink('ByLastSelectedSc')"
                        label="最近入选 SC"
                      />
                    </th>
                    <th class="next02-pink-list-sortable sc-sortable-head sc-number casts-number">
                      <SortLink
                        class="next02-pink-list-sort-link sc-sort-link"
                        :link="sortLink('ByHighestRank')"
                        label="Highest Rank"
                      />
                    </th>
                    <th class="sc-number casts-number">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!hasRows">
                    <td colspan="11"><div class="sc-empty-state">还没有可展示的演员。</div></td>
                  </tr>
                  <tr v-for="row in rows" v-else :key="rowKey(row)">
                    <td class="sc-content-cell casts-content-cell">
                      <a
                        class="next02-pink-list-link"
                        :href="`/castdetail/${value(row, 'Id')}`"
                        :title="value(row, 'Name')"
                        target="_blank"
                        rel="noopener noreferrer"
                        >{{ value(row, "Name") }}</a
                      >
                    </td>
                    <td class="sc-content-cell casts-content-cell" :title="value(row, 'Chinese') || undefined">
                      {{ value(row, "Chinese") || "-" }}
                    </td>
                    <td class="sc-number casts-number">{{ formatAgeFromBirth(value(row, "BirthDay")) }}</td>
                    <td class="sc-number casts-number">{{ formatHeightCm(value(row, "Height")) }}</td>
                    <td class="sc-number casts-number">{{ value(row, "MovieNumber") || 0 }}</td>
                    <td class="sc-number casts-number">{{ value(row, "OwnedWMediaNumber") || 0 }}</td>
                    <td class="sc-number casts-number">{{ value(row, "ScTimes") || 0 }}</td>
                    <td class="sc-number casts-number">{{ value(row, "ComeTimes") || 0 }}</td>
                    <td class="sc-number casts-number">{{ formatUnixDate(value(row, "LastSelectedScTime")) }}</td>
                    <td class="sc-number casts-number">{{ formatRank(value(row, "HighestRank")) }}</td>
                    <td class="sc-number casts-number casts-action-cell">
                      <button type="button" class="button-secondary casts-edit-button" @click="openCafoEditor(row)">
                        编辑资料
                      </button>
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
  <CafoEditModal
    :active="cafoModalOpen"
    :loading="cafoLoading"
    :saving="cafoSaving"
    :person-name="cafoPersonName"
    :cafos="cafoRows"
    :error-message="cafoError"
    :success-message="cafoMessage"
    @close="closeCafoEditor"
    @save="saveCafoProfile"
  />
</template>

<script setup lang="ts">
import "@/features/movies/styles/index.css";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import {
  apiPathForRoute,
  getJSON,
  listPersonCafos,
  updatePersonCafo,
  type CafoProfile,
  type CafoWriteInput
} from "@/features/movies/api";
import CafoEditModal from "@/features/movies/components/casts/CafoEditModal.vue";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import type { PageInfo } from "@/shared/types/gc";
import type { MovieRecord } from "@/features/movies/types";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<MovieRecord>({});
const rows = ref<MovieRecord[]>([]);
const loading = ref(false);
const error = ref("");
const cafoModalOpen = ref(false);
const cafoLoading = ref(false);
const cafoSaving = ref(false);
const cafoPersonID = ref(0);
const cafoPersonName = ref("");
const cafoRows = ref<CafoProfile[]>([]);
const cafoError = ref("");
const cafoMessage = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const data = asRecord(await getJSON(apiPathForRoute(route)));
    payload.value = data;
    rows.value = asArray<MovieRecord>(pick(data, "Items", "items", "Rows", "rows"));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title") || "Casts");
const query = computed(() => asRecord(pick(payload.value, "Query", "query")));
const total = computed(() => text(payload.value, "Total", "total") || rows.value.length);
const hasRows = computed(() => rows.value.length > 0);
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<MovieRecord>(pageInfo.value.Links || pageInfo.value.links));
const sortMap = computed(() => asRecord(pick(payload.value, "CastSortQuery", "castSortQuery")));

function queryValue(key: string): string {
  return text(query.value, key);
}

function value(row: MovieRecord, ...keys: string[]): string {
  return text(row, ...keys);
}

function rowKey(row: MovieRecord): string {
  return text(row, "Id", "id") || JSON.stringify(row);
}

async function openCafoEditor(row: MovieRecord) {
  const personID = Number(value(row, "Id", "id"));
  cafoPersonID.value = personID;
  cafoPersonName.value = value(row, "Name", "name");
  cafoRows.value = [];
  cafoError.value = "";
  cafoMessage.value = "";
  cafoModalOpen.value = true;
  cafoLoading.value = true;
  try {
    const linkedCafos = await listPersonCafos(personID);
    if (cafoModalOpen.value && cafoPersonID.value === personID) cafoRows.value = linkedCafos;
  } catch (err) {
    if (cafoModalOpen.value && cafoPersonID.value === personID) {
      cafoError.value = err instanceof Error ? err.message : String(err);
    }
  } finally {
    if (cafoPersonID.value === personID) cafoLoading.value = false;
  }
}

function closeCafoEditor() {
  if (!cafoSaving.value) cafoModalOpen.value = false;
}

async function saveCafoProfile(cafoID: number, input: CafoWriteInput) {
  cafoSaving.value = true;
  cafoError.value = "";
  cafoMessage.value = "";
  try {
    const response = await updatePersonCafo(cafoPersonID.value, cafoID, input);
    cafoRows.value = cafoRows.value.map((row) => (row.id === response.cafo.id ? response.cafo : row));
    rows.value = rows.value.map((row) => {
      if (Number(value(row, "Id", "id")) !== cafoPersonID.value) return row;
      return { ...row, ...response.person };
    });
    cafoPersonName.value = text(response.person, "Name", "name") || cafoPersonName.value;
    cafoMessage.value = response.message || "演员资料更新成功";
  } catch (err) {
    cafoError.value = err instanceof Error ? err.message : String(err);
  } finally {
    cafoSaving.value = false;
  }
}

function sortLink(key: string): MovieRecord {
  return asRecord(sortMap.value[key]);
}

function formatUnixDate(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function formatAgeFromBirth(input: unknown): string {
  const n = Number(input || 0);
  if (!Number.isFinite(n) || n === 0) return "-";
  const birthday = new Date(n * 1000);
  const now = new Date();
  let age = now.getFullYear() - birthday.getFullYear();
  if (
    now.getMonth() < birthday.getMonth() ||
    (now.getMonth() === birthday.getMonth() && now.getDate() < birthday.getDate())
  ) {
    age -= 1;
  }
  return age >= 0 ? String(age) : "-";
}

function formatHeightCm(input: unknown): string {
  const n = Number(input || 0);
  return n > 0 ? `${text({ input }, "input")} cm` : "-";
}

function formatRank(input: unknown): string {
  const n = Number(input || 0);
  return n > 0 ? `#${text({ input }, "input")}` : "-";
}
</script>
