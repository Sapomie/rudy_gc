<template>
  <PageShell class="next02-pink-list-page-shell medias-page-shell" :title="title" note="" eyebrow="">
    <div
      id="msgArea"
      class="alert fade show small py-2 px-3"
      :class="msgClass"
      :style="{ display: msgText ? 'block' : 'none' }"
    >
      {{ msgText }}
    </div>
    <div id="page-root" class="container-fluid next02-pink-list-page medias-page-root">
      <div v-if="loading" class="alert alert-info py-2 next02-pink-list-state">加载中...</div>
      <div v-else-if="error" class="alert alert-danger next02-pink-list-state">{{ error }}</div>
      <template v-else>
        <section class="next02-pink-list-hero medias-list-hero gc-page-head">
          <div class="next02-pink-list-hero-copy">
            <span class="next02-pink-list-eyebrow">MEDIA / LIBRARY</span>
            <h1>{{ title }}</h1>
            <p>按列表查看 w_media(source_type=2) 的基础数据，并复用 MovieType 缓存补充演员与播放信息。</p>
          </div>
        </section>

        <div class="card next02-pink-list-panel medias-list-panel">
          <div class="card-body">
            <div
              class="rank-stats-toolbar-scope rank-page-shell cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell standard-filter-toolbar medias-filter-toolbar"
            >
              <form
                class="sticky-toolbar cards-toolbar rank-sticky-toolbar next02-pink-cards-toolbar rank-stats-toolbar standard-filter-toolbar-form standard-filter-toolbar-form-two-row medias-filter-form photo-results-toolbar"
                method="get"
                action="/medialist"
                id="mediasFilterForm"
              >
                <input type="hidden" name="p" value="1" />
                <input v-if="queryValue('PageSize')" type="hidden" name="ps" :value="queryValue('PageSize')" />
                <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')" />
                <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')" />

                <div class="standard-filter-toolbar-row standard-filter-toolbar-row-primary">
                  <div class="cards-toolbar-sort rank-toolbar-sort standard-filter-toolbar-fields sort-controls">
                    <div class="rank-toolbar-group standard-filter-group medias-filter-identity-group">
                      <label
                        class="rank-stats-filter-field standard-filter-field medias-filter-field medias-filter-keyword"
                      >
                        <span>识别码搜索</span>
                        <input
                          class="rank-date-input standard-filter-input"
                          type="text"
                          name="keyword"
                          :value="queryValue('Keyword')"
                          placeholder="Movie Name"
                        />
                      </label>
                    </div>

                    <div class="rank-toolbar-group standard-filter-group medias-filter-number-group">
                      <div
                        v-for="field in numberRangeFields"
                        :key="field.label"
                        class="rank-stats-filter-field standard-filter-field medias-filter-field"
                      >
                        <span>{{ field.label }}</span>
                        <span class="medias-filter-range standard-filter-range">
                          <input
                            class="rank-date-input standard-filter-input standard-filter-range-input"
                            :type="field.type"
                            :name="field.minName"
                            :value="queryValue(field.minKey)"
                            :min="field.min"
                            :step="field.step"
                            :placeholder="field.minPlaceholder"
                            :aria-label="`${field.label} minimum`"
                          />
                          <span aria-hidden="true">–</span>
                          <input
                            class="rank-date-input standard-filter-input standard-filter-range-input"
                            :type="field.type"
                            :name="field.maxName"
                            :value="queryValue(field.maxKey)"
                            :min="field.min"
                            :step="field.step"
                            :placeholder="field.maxPlaceholder"
                            :aria-label="`${field.label} maximum`"
                          />
                        </span>
                      </div>
                    </div>

                    <div class="rank-toolbar-group standard-filter-group medias-filter-state-group">
                      <label class="rank-stats-filter-field standard-filter-field medias-filter-field">
                        <span>自制</span>
                        <select class="rank-date-input standard-filter-input standard-filter-select" name="self_make">
                          <option
                            v-for="option in selfMakeOptions"
                            :key="option.value"
                            :value="option.value"
                            :selected="queryValue('SelfMake') === option.value"
                          >
                            {{ option.label }}
                          </option>
                        </select>
                      </label>

                      <label class="rank-stats-filter-field standard-filter-field medias-filter-field">
                        <span>遮罩</span>
                        <select class="rank-date-input standard-filter-input standard-filter-select" name="has_mask">
                          <option
                            v-for="option in maskOptions"
                            :key="option.value"
                            :value="option.value"
                            :selected="queryValue('HasMask') === option.value"
                          >
                            {{ option.label }}
                          </option>
                        </select>
                      </label>
                    </div>
                  </div>
                </div>

                <div class="standard-filter-toolbar-row standard-filter-toolbar-row-secondary">
                  <div class="cards-toolbar-sort rank-toolbar-sort standard-filter-toolbar-fields sort-controls">
                    <div class="rank-toolbar-group standard-filter-group medias-filter-date-group">
                      <div
                        v-for="field in dateRangeFields"
                        :key="field.label"
                        class="rank-stats-filter-field standard-filter-field medias-filter-field medias-filter-date smart-pick-date-field"
                      >
                        <span>{{ field.label }}</span>
                        <span class="medias-filter-range standard-filter-range">
                          <input
                            class="rank-date-input standard-filter-input standard-filter-date-input smart-pick-date-input"
                            :type="field.type"
                            :name="field.minName"
                            :value="queryValue(field.minKey)"
                            :placeholder="field.minPlaceholder"
                            :aria-label="`${field.label} from`"
                            inputmode="numeric"
                            maxlength="10"
                            pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                          />
                          <span aria-hidden="true">–</span>
                          <input
                            class="rank-date-input standard-filter-input standard-filter-date-input smart-pick-date-input"
                            :type="field.type"
                            :name="field.maxName"
                            :value="queryValue(field.maxKey)"
                            :placeholder="field.maxPlaceholder"
                            :aria-label="`${field.label} to`"
                            inputmode="numeric"
                            maxlength="10"
                            pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                          />
                        </span>
                      </div>
                    </div>
                  </div>

                  <div class="rank-navigation-group standard-filter-actions medias-filter-actions">
                    <button class="button-secondary rank-toolbar-action" type="submit">筛选</button>
                    <a class="button-secondary rank-toolbar-action" href="/medialist">清空</a>
                  </div>

                  <div class="toolbar-line cards-toolbar-bottom rank-toolbar-line standard-filter-status-line">
                    <div class="owned-line cards-toolbar-filters rank-toolbar-filters"></div>
                    <div class="next02-pink-cards-toolbar-status">
                      <div class="next02-pink-cards-metrics" aria-label="媒体列表统计">
                        <span :title="`总数 ${total}`" :aria-label="`总数 ${total}`"
                          ><strong>{{ total }}</strong></span
                        >
                        <span :title="`当前页 ${items.length}`" :aria-label="`当前页 ${items.length}`"
                          ><strong>{{ items.length }}</strong></span
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

            <div class="next02-pink-list-table-wrap medias-list-table-wrap">
              <table class="table table-hover align-middle mb-0 next02-pink-list-table medias-table data-table">
                <colgroup>
                  <col class="medias-col-name" />
                  <col class="medias-col-cast" />
                  <col class="medias-col-size" />
                  <col class="medias-col-height" />
                  <col class="medias-col-duration" />
                  <col class="medias-col-bitrate" />
                  <col class="medias-col-frame" />
                  <col class="medias-col-play" />
                  <col class="medias-col-self-make" />
                  <col class="medias-col-mask" />
                  <col class="medias-col-sc" />
                  <col class="medias-col-last-selected-sc" />
                  <col class="medias-col-birth" />
                  <col class="medias-col-release" />
                  <col class="medias-col-action" />
                </colgroup>
                <thead>
                  <tr>
                    <th class="next02-pink-list-sortable">
                      <SortLink class="next02-pink-list-sort-link" :link="sortLink('ByMovieName')" label="识别码" />
                    </th>
                    <th>演员</th>
                    <th class="next02-pink-list-sortable medias-number">
                      <SortLink class="next02-pink-list-sort-link" :link="sortLink('BySize')" label="大小(GB)" />
                    </th>
                    <th class="next02-pink-list-sortable medias-number">
                      <SortLink class="next02-pink-list-sort-link" :link="sortLink('ByHeight')" label="高度" />
                    </th>
                    <th class="next02-pink-list-sortable medias-number">
                      <SortLink class="next02-pink-list-sort-link" :link="sortLink('ByDuration')" label="时长(分钟)" />
                    </th>
                    <th class="next02-pink-list-sortable medias-number">
                      <SortLink class="next02-pink-list-sort-link" :link="sortLink('ByBitRate')" label="码率" />
                    </th>
                    <th class="next02-pink-list-sortable medias-number">
                      <SortLink
                        class="next02-pink-list-sort-link"
                        :link="sortLink('ByFrameAverage')"
                        label="平均帧率"
                      />
                    </th>
                    <th>播放</th>
                    <th class="next02-pink-list-sortable">
                      <SortLink class="next02-pink-list-sort-link" :link="sortLink('BySelfMake')" label="自制" />
                    </th>
                    <th class="next02-pink-list-sortable">
                      <SortLink class="next02-pink-list-sort-link" :link="sortLink('ByHasMask')" label="遮罩" />
                    </th>
                    <th class="next02-pink-list-sortable medias-number">
                      <SortLink class="next02-pink-list-sort-link" :link="sortLink('ByScTimes')" label="SC" />
                    </th>
                    <th class="next02-pink-list-sortable medias-number">
                      <SortLink
                        class="next02-pink-list-sort-link"
                        :link="sortLink('ByLastSelectedSc')"
                        label="最近入选 SC"
                      />
                    </th>
                    <th class="next02-pink-list-sortable medias-number">
                      <SortLink class="next02-pink-list-sort-link" :link="sortLink('ByBirthTime')" label="下载时间" />
                    </th>
                    <th class="next02-pink-list-sortable medias-number">
                      <SortLink
                        class="next02-pink-list-sort-link"
                        :link="sortLink('ByReleasingDate')"
                        label="发行时间"
                      />
                    </th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!items.length">
                    <td colspan="15"><div class="sc-empty-state">暂无数据</div></td>
                  </tr>
                  <tr v-for="item in items" v-else :key="itemKey(item)" :data-film-id="value(item, 'Id')">
                    <td class="medias-content-cell">
                      <a
                        class="next02-pink-list-link"
                        :href="value(item, 'MovieHref')"
                        :title="value(item, 'MovieName')"
                        target="_blank"
                        rel="noopener noreferrer"
                        >{{ value(item, "MovieName") }}</a
                      >
                    </td>
                    <td class="medias-content-cell" :title="value(item, 'CastName') || undefined">
                      <template v-if="casts(item).length">
                        <template
                          v-for="(cast, index) in casts(item)"
                          :key="`${value(item, 'Id')}-${cast.Name || index}`"
                        >
                          <span v-if="index"> / </span
                          ><a
                            class="next02-pink-list-link"
                            :href="text(cast, 'Href')"
                            :title="text(cast, 'Name')"
                            target="_blank"
                            rel="noopener noreferrer"
                            >{{ text(cast, "Name") }}</a
                          >
                        </template>
                      </template>
                      <template v-else>{{ value(item, "CastName") || "-" }}</template>
                    </td>
                    <td class="medias-number">{{ value(item, "SizeGB") }}</td>
                    <td class="medias-number">
                      <span data-role="film-height">{{ probedValue(item, "height", "Height") }}</span>
                    </td>
                    <td class="medias-number">
                      <span data-role="film-duration">{{
                        probedValue(item, "duration_minutes", "DurationMinutes")
                      }}</span>
                    </td>
                    <td class="medias-number">
                      <span data-role="film-bitrate">{{ probedValue(item, "bit_rate", "BitRate") }}</span>
                    </td>
                    <td class="medias-number">
                      <span data-role="film-frame">{{ probedFrame(item) }}</span>
                    </td>
                    <td>
                      <a
                        v-if="showPlay(item)"
                        class="btn btn-sm medias-row-action"
                        :class="value(item, 'PlayButtonClass') || 'btn-outline-primary'"
                        :href="value(item, 'VideoURL')"
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        {{ value(item, "PlayButtonText") || "MPlay" }}
                      </a>
                      <span v-else class="text-muted">-</span>
                    </td>
                    <td>{{ value(item, "SelfMakeText") }}</td>
                    <td>{{ value(item, "HasMaskText") }}</td>
                    <td class="medias-number">{{ value(item, "ScTimes") }}</td>
                    <td class="medias-number">{{ formatUnixDate(value(item, "LastSelectedScTime")) }}</td>
                    <td class="medias-number">{{ formatUnixDate(value(item, "BirthTime")) }}</td>
                    <td class="medias-number">{{ formatUnixDate(value(item, "ReleasingDate")) }}</td>
                    <td>
                      <button
                        type="button"
                        class="btn btn-sm btn-outline-primary js-film-probe medias-row-action"
                        :data-film-id="value(item, 'Id')"
                        :disabled="probeLoading[value(item, 'Id')]"
                        @click="handleProbe(item)"
                      >
                        {{ probeLoading[value(item, "Id")] ? "scanning..." : "scan_meta" }}
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
</template>

<script setup lang="ts">
import "@/features/media/styles/index.css";
import { computed, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/features/media/api";
import PageShell from "@/shared/components/PageShell.vue";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import SortLink from "@/shared/components/navigation/SortLink.vue";
import type { PageInfo } from "@/shared/types/gc";
import type { MediaRecord } from "@/features/media/types";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<MediaRecord>({});
const loading = ref(false);
const error = ref("");
const msgText = ref("");
const msgType = ref("success");
const probeLoading = reactive<Record<string, boolean>>({});
const probeResults = reactive<Record<string, MediaRecord>>({});

const numberRangeFields = [
  {
    label: "大小范围(GB)",
    minName: "size_min",
    minKey: "SizeMin",
    maxName: "size_max",
    maxKey: "SizeMax",
    type: "number",
    min: "0",
    step: "0.01",
    minPlaceholder: "Min",
    maxPlaceholder: "Max"
  },
  {
    label: "高度范围",
    minName: "height_min",
    minKey: "HeightMin",
    maxName: "height_max",
    maxKey: "HeightMax",
    type: "number",
    min: "0",
    step: undefined,
    minPlaceholder: "Min",
    maxPlaceholder: "Max"
  },
  {
    label: "时长范围(分钟)",
    minName: "duration_min",
    minKey: "DurationMin",
    maxName: "duration_max",
    maxKey: "DurationMax",
    type: "number",
    min: "0",
    step: undefined,
    minPlaceholder: "Min",
    maxPlaceholder: "Max"
  },
  {
    label: "码率范围",
    minName: "bit_rate_min",
    minKey: "BitRateMin",
    maxName: "bit_rate_max",
    maxKey: "BitRateMax",
    type: "number",
    min: "0",
    step: undefined,
    minPlaceholder: "Min",
    maxPlaceholder: "Max"
  },
  {
    label: "平均帧率范围",
    minName: "frame_average_min",
    minKey: "FrameAverageMin",
    maxName: "frame_average_max",
    maxKey: "FrameAverageMax",
    type: "number",
    min: "0",
    step: "0.01",
    minPlaceholder: "Min",
    maxPlaceholder: "Max"
  },
  {
    label: "SC 次数范围",
    minName: "sc_times_min",
    minKey: "ScTimesMin",
    maxName: "sc_times_max",
    maxKey: "ScTimesMax",
    type: "number",
    min: "0",
    step: undefined,
    minPlaceholder: "Min",
    maxPlaceholder: "Max"
  }
];
const dateRangeFields = [
  {
    label: "最近入选 SC 时间",
    minName: "last_selected_sc_from",
    minKey: "LastSelectedScFrom",
    maxName: "last_selected_sc_to",
    maxKey: "LastSelectedScTo",
    type: "text",
    minPlaceholder: "YYYY-MM-DD",
    maxPlaceholder: "YYYY-MM-DD"
  },
  {
    label: "下载时间",
    minName: "birth_from",
    minKey: "BirthFrom",
    maxName: "birth_to",
    maxKey: "BirthTo",
    type: "text",
    minPlaceholder: "YYYY-MM-DD",
    maxPlaceholder: "YYYY-MM-DD"
  },
  {
    label: "发行时间",
    minName: "releasing_from",
    minKey: "ReleasingFrom",
    maxName: "releasing_to",
    maxKey: "ReleasingTo",
    type: "text",
    minPlaceholder: "YYYY-MM-DD",
    maxPlaceholder: "YYYY-MM-DD"
  }
];
const selfMakeOptions = [
  { value: "", label: "全部" },
  { value: "1", label: "否" },
  { value: "2", label: "是" }
];
const maskOptions = [
  { value: "", label: "全部" },
  { value: "1", label: "有码" },
  { value: "2", label: "去码" },
  { value: "3", label: "无码" }
];

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

const title = computed(() => text(payload.value, "PageTitle", "Title") || "Medias");
const total = computed(() => text(payload.value, "Total") || "0");
const query = computed(() => asRecord(pick(payload.value, "Query")));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")) as PageInfo);
const pageLinks = computed(() => asArray<MediaRecord>(pageInfo.value.Links));
const items = computed(() => asArray<MediaRecord>(pick(payload.value, "Items")));
const sortLinks = computed(() => asRecord(pick(payload.value, "FilmSortQuery")));
const msgClass = computed(() => (msgType.value === "danger" ? "alert-danger" : "alert-success"));

function queryValue(key: string): string {
  return text(query.value, key);
}

function value(row: MediaRecord, key: string): string {
  return text(row, key);
}

function itemKey(row: MediaRecord): string {
  return value(row, "Id") || value(row, "MovieName");
}

function casts(row: MediaRecord): MediaRecord[] {
  return asArray<MediaRecord>(pick(row, "Casts"));
}

function sortLink(key: string): MediaRecord {
  return asRecord(sortLinks.value[key]);
}

function showPlay(row: MediaRecord): boolean {
  return Boolean(row.ShowPlayButton || row.showPlayButton);
}

function formatUnixDate(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function formatPositive(input: unknown): string {
  const n = Number(input || 0);
  return n > 0 ? String(input) : "-";
}

function probedValue(row: MediaRecord, probeKey: string, itemKeyName: string): string {
  const id = value(row, "Id");
  if (id && probeResults[id] && probeResults[id][probeKey] !== undefined) {
    return formatPositive(probeResults[id][probeKey]);
  }
  return formatPositive(row[itemKeyName]);
}

function probedFrame(row: MediaRecord): string {
  const id = value(row, "Id");
  if (id && probeResults[id] && probeResults[id].frame_average !== undefined) {
    return text(probeResults[id], "frame_average") || "-";
  }
  return value(row, "FrameAverage") || "-";
}

function showMessage(message: string, type = "success") {
  msgText.value = message;
  msgType.value = type;
  window.setTimeout(() => {
    msgText.value = "";
  }, 2600);
}

async function handleProbe(row: MediaRecord) {
  const id = value(row, "Id");
  if (!id) return;
  probeLoading[id] = true;
  try {
    const data = asRecord(await sendJSON(`/api/medias/${encodeURIComponent(id)}/probe-meta`, "POST"));
    if (!data.ok) {
      throw new Error(text(data, "error") || "scan_meta failed");
    }
    probeResults[id] = asRecord(data.result);
    showMessage("scan_meta 完成");
  } catch (err) {
    showMessage(err instanceof Error ? err.message : "scan_meta 失败", "danger");
  } finally {
    probeLoading[id] = false;
  }
}
</script>
