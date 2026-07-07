<template>
  <PageShell class="medias-page-shell" title="Medias" note="" eyebrow="">
    <div id="msgArea" class="alert fade show small py-2 px-3" :class="msgClass" :style="{ display: msgText ? 'block' : 'none' }">
      {{ msgText }}
    </div>
    <div id="page-root" class="container-fluid medias-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div class="d-flex align-items-center justify-content-between mb-3">
          <div>
            <h4 class="mb-1">{{ title }}</h4>
            <p class="text-muted mb-0">按列表查看 w_media(source_type=2) 的基础数据，并复用 MovieType 缓存补充演员与播放信息。</p>
          </div>
        </div>

        <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
          <div class="sc-page-summary">
            <span class="sc-page-summary-label">Total Medias</span>
            <span class="sc-page-summary-value">{{ total }}</span>
          </div>
          <div v-if="pageLinks.length" class="ms-xl-auto">
            <MediaPagination :links="pageLinks" />
          </div>
        </div>

        <div class="section-card card mb-3">
          <div class="card-body">
            <form class="row g-3 align-items-end" method="get" action="/medias">
              <input v-if="queryValue('PageSize')" type="hidden" name="ps" :value="queryValue('PageSize')">
              <input v-if="queryValue('Sort')" type="hidden" name="sort" :value="queryValue('Sort')">
              <input v-if="queryValue('Order')" type="hidden" name="order" :value="queryValue('Order')">

              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">识别码搜索</label>
                <input class="form-control" type="text" name="keyword" :value="queryValue('Keyword')" placeholder="Movie Name">
              </div>

              <div v-for="field in numberRangeFields" :key="field.label" class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">{{ field.label }}</label>
                <div class="row g-2">
                  <div class="col-6">
                    <input class="form-control" :type="field.type" :name="field.minName" :value="queryValue(field.minKey)" :min="field.min" :step="field.step" :placeholder="field.minPlaceholder">
                  </div>
                  <div class="col-6">
                    <input class="form-control" :type="field.type" :name="field.maxName" :value="queryValue(field.maxKey)" :min="field.min" :step="field.step" :placeholder="field.maxPlaceholder">
                  </div>
                </div>
              </div>

              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">自制</label>
                <select class="form-select" name="self_make">
                  <option v-for="option in selfMakeOptions" :key="option.value" :value="option.value" :selected="queryValue('SelfMake') === option.value">{{ option.label }}</option>
                </select>
              </div>

              <div class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">遮罩</label>
                <select class="form-select" name="has_mask">
                  <option v-for="option in maskOptions" :key="option.value" :value="option.value" :selected="queryValue('HasMask') === option.value">{{ option.label }}</option>
                </select>
              </div>

              <div v-for="field in dateRangeFields" :key="field.label" class="col-12 col-md-6 col-xl-3">
                <label class="form-label mb-1">{{ field.label }}</label>
                <div class="row g-2">
                  <div class="col-6">
                    <input class="form-control" :type="field.type" :name="field.minName" :value="queryValue(field.minKey)" :placeholder="field.minPlaceholder">
                  </div>
                  <div class="col-6">
                    <input class="form-control" :type="field.type" :name="field.maxName" :value="queryValue(field.maxKey)" :placeholder="field.maxPlaceholder">
                  </div>
                </div>
              </div>

              <div class="col-12 d-flex gap-2">
                <button class="btn btn-primary" type="submit">Apply</button>
                <a class="btn btn-outline-secondary" href="/medias">Clear</a>
              </div>
            </form>
          </div>
        </div>

        <div class="section-card card">
          <div class="card-body">
            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 list-table medias-table">
                <thead>
                  <tr>
                    <th><SortLink :link="sortLink('ByMovieName')" label="识别码" /></th>
                    <th>演员</th>
                    <th><SortLink :link="sortLink('BySize')" label="大小(GB)" /></th>
                    <th><SortLink :link="sortLink('ByHeight')" label="高度" /></th>
                    <th><SortLink :link="sortLink('ByDuration')" label="时长(分钟)" /></th>
                    <th><SortLink :link="sortLink('ByBitRate')" label="码率" /></th>
                    <th><SortLink :link="sortLink('ByFrameAverage')" label="平均帧率" /></th>
                    <th>播放</th>
                    <th><SortLink :link="sortLink('BySelfMake')" label="自制" /></th>
                    <th><SortLink :link="sortLink('ByHasMask')" label="遮罩" /></th>
                    <th><SortLink :link="sortLink('ByScTimes')" label="SC" /></th>
                    <th><SortLink :link="sortLink('ByLastSc')" label="最后SC" /></th>
                    <th><SortLink :link="sortLink('ByBirthTime')" label="下载时间" /></th>
                    <th><SortLink :link="sortLink('ByReleasingDate')" label="发行时间" /></th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!items.length">
                    <td colspan="15"><div class="sc-empty-state">暂无数据</div></td>
                  </tr>
                  <tr v-for="item in items" v-else :key="itemKey(item)" :data-film-id="value(item, 'Id')">
                    <td><a :href="value(item, 'MovieHref')">{{ value(item, "MovieName") }}</a></td>
                    <td class="wide">
                      <template v-if="casts(item).length">
                        <template v-for="(cast, index) in casts(item)" :key="`${value(item, 'Id')}-${cast.Name || index}`">
                          <span v-if="index"> / </span><a :href="text(cast, 'Href')">{{ text(cast, "Name") }}</a>
                        </template>
                      </template>
                      <template v-else>{{ value(item, "CastName") || "-" }}</template>
                    </td>
                    <td>{{ value(item, "SizeGB") }}</td>
                    <td><span data-role="film-height">{{ probedValue(item, "height", "Height") }}</span></td>
                    <td><span data-role="film-duration">{{ probedValue(item, "duration_minutes", "DurationMinutes") }}</span></td>
                    <td><span data-role="film-bitrate">{{ probedValue(item, "bit_rate", "BitRate") }}</span></td>
                    <td><span data-role="film-frame">{{ probedFrame(item) }}</span></td>
                    <td>
                      <a v-if="showPlay(item)" class="btn btn-sm" :class="value(item, 'PlayButtonClass') || 'btn-outline-primary'" :href="value(item, 'VideoURL')">
                        {{ value(item, "PlayButtonText") || "MPlay" }}
                      </a>
                      <span v-else class="text-muted">-</span>
                    </td>
                    <td>{{ value(item, "SelfMakeText") }}</td>
                    <td>{{ value(item, "HasMaskText") }}</td>
                    <td>{{ value(item, "ScTimes") }}</td>
                    <td>{{ formatUnixDate(value(item, "LastScTime")) }}</td>
                    <td>{{ formatUnixDate(value(item, "BirthTime")) }}</td>
                    <td>{{ formatUnixDate(value(item, "ReleasingDate")) }}</td>
                    <td>
                      <button type="button" class="btn btn-sm btn-outline-primary js-film-probe" :data-film-id="value(item, 'Id')" :disabled="probeLoading[value(item, 'Id')]" @click="handleProbe(item)">
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
import { computed, defineComponent, h, reactive, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord, PageInfo } from "@/types/gc";
import { asArray, asRecord, normalizeHref, pick, text } from "@/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const msgText = ref("");
const msgType = ref("success");
const probeLoading = reactive<Record<string, boolean>>({});
const probeResults = reactive<Record<string, AnyRecord>>({});

const numberRangeFields = [
  { label: "大小范围(GB)", minName: "size_min", minKey: "SizeMin", maxName: "size_max", maxKey: "SizeMax", type: "number", min: "0", step: "0.01", minPlaceholder: "Min", maxPlaceholder: "Max" },
  { label: "高度范围", minName: "height_min", minKey: "HeightMin", maxName: "height_max", maxKey: "HeightMax", type: "number", min: "0", step: undefined, minPlaceholder: "Min", maxPlaceholder: "Max" },
  { label: "时长范围(分钟)", minName: "duration_min", minKey: "DurationMin", maxName: "duration_max", maxKey: "DurationMax", type: "number", min: "0", step: undefined, minPlaceholder: "Min", maxPlaceholder: "Max" },
  { label: "码率范围", minName: "bit_rate_min", minKey: "BitRateMin", maxName: "bit_rate_max", maxKey: "BitRateMax", type: "number", min: "0", step: undefined, minPlaceholder: "Min", maxPlaceholder: "Max" },
  { label: "平均帧率范围", minName: "frame_average_min", minKey: "FrameAverageMin", maxName: "frame_average_max", maxKey: "FrameAverageMax", type: "number", min: "0", step: "0.01", minPlaceholder: "Min", maxPlaceholder: "Max" },
  { label: "SC 次数范围", minName: "sc_times_min", minKey: "ScTimesMin", maxName: "sc_times_max", maxKey: "ScTimesMax", type: "number", min: "0", step: undefined, minPlaceholder: "Min", maxPlaceholder: "Max" }
];
const dateRangeFields = [
  { label: "最后 SC 时间", minName: "last_sc_from", minKey: "LastScFrom", maxName: "last_sc_to", maxKey: "LastScTo", type: "date", minPlaceholder: "", maxPlaceholder: "" },
  { label: "下载时间", minName: "birth_from", minKey: "BirthFrom", maxName: "birth_to", maxKey: "BirthTo", type: "date", minPlaceholder: "", maxPlaceholder: "" },
  { label: "发行时间", minName: "releasing_from", minKey: "ReleasingFrom", maxName: "releasing_to", maxKey: "ReleasingTo", type: "date", minPlaceholder: "", maxPlaceholder: "" }
];
const selfMakeOptions = [{ value: "", label: "全部" }, { value: "1", label: "否" }, { value: "2", label: "是" }];
const maskOptions = [{ value: "", label: "全部" }, { value: "1", label: "有码" }, { value: "2", label: "去码" }, { value: "3", label: "无码" }];

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
const pageLinks = computed(() => asArray<AnyRecord>(pageInfo.value.Links));
const items = computed(() => asArray<AnyRecord>(pick(payload.value, "Items")));
const sortLinks = computed(() => asRecord(pick(payload.value, "FilmSortQuery")));
const msgClass = computed(() => msgType.value === "danger" ? "alert-danger" : "alert-success");

function queryValue(key: string): string {
  return text(query.value, key);
}

function value(row: AnyRecord, key: string): string {
  return text(row, key);
}

function itemKey(row: AnyRecord): string {
  return value(row, "Id") || value(row, "MovieName");
}

function casts(row: AnyRecord): AnyRecord[] {
  return asArray<AnyRecord>(pick(row, "Casts"));
}

function sortLink(key: string): AnyRecord {
  return asRecord(sortLinks.value[key]);
}

function showPlay(row: AnyRecord): boolean {
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

function probedValue(row: AnyRecord, probeKey: string, itemKeyName: string): string {
  const id = value(row, "Id");
  if (id && probeResults[id] && probeResults[id][probeKey] !== undefined) {
    return formatPositive(probeResults[id][probeKey]);
  }
  return formatPositive(row[itemKeyName]);
}

function probedFrame(row: AnyRecord): string {
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

async function handleProbe(row: AnyRecord) {
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

const SortLink = defineComponent({
  props: {
    link: { type: Object, required: true },
    label: { type: String, required: true }
  },
  setup(props) {
    return () => {
      const link = asRecord(props.link);
      const active = Boolean(link.Active);
      const desc = Boolean(link.Desc);
      return h("a", {
        class: ["sort-link", active ? "active" : ""],
        href: normalizeHref(text(link, "Href") || "#")
      }, `${props.label}${active ? (desc ? " ↓" : " ↑") : ""}`);
    };
  }
});

const MediaPagination = defineComponent({
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
        return h("li", { key: `${label}-${index}`, class: ["page-item", active ? "active" : "", disabled ? "disabled" : ""] }, ellipsis
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
</script>
