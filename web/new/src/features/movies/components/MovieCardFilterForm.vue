<template>
  <div v-if="isCardsAggVariant && hasFilter" :id="collapseId" class="collapse" :class="{ show: open }">
    <div class="card mb-3">
      <div class="card-body">
        <form :id="formId" class="row g-2 align-items-end" method="get" :action="action" @submit.prevent="handleSubmit">
          <input
            v-for="item in hiddenFields"
            :key="item.name"
            type="hidden"
            :name="item.name"
            :value="item.value"
            :data-explicit="explicitFlag(item.name)"
            :data-initial-value="item.value"
          />

          <div v-for="field in textFields" :key="field.name" class="col-12 col-md-6 col-xl-2">
            <label class="form-label mb-1">{{ field.label }}</label>
            <input
              class="form-control"
              :type="field.type || 'text'"
              :name="field.name"
              :value="fieldValue(field.key)"
              :placeholder="field.placeholder || ''"
              :step="field.step"
              :data-explicit="explicitFlag(field.name)"
              :data-initial-value="fieldValue(field.key)"
            />
          </div>

          <div v-for="field in selectFields" :key="field.name" class="col-12 col-md-6 col-xl-2">
            <label class="form-label mb-1">{{ field.label }}</label>
            <select
              class="form-select"
              :name="field.name"
              :value="fieldValue(field.key)"
              :data-explicit="explicitFlag(field.name)"
              :data-initial-value="fieldValue(field.key)"
            >
              <option v-for="option in field.options" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </div>

          <div v-for="field in extraTextFields" :key="field.name" class="col-12 col-md-6 col-xl-2">
            <label class="form-label mb-1">{{ field.label }}</label>
            <input
              class="form-control"
              :type="field.type || 'text'"
              :name="field.name"
              :value="fieldValue(field.key)"
              :placeholder="field.placeholder || ''"
              :data-explicit="explicitFlag(field.name)"
              :data-initial-value="fieldValue(field.key)"
            />
          </div>

          <div v-for="pair in datePairs" :key="pair.label" class="col-12 col-xl-4 smart-pick-date-field">
            <label class="form-label mb-1">{{ pair.label }}</label>
            <div class="row g-2">
              <div class="col-6">
                <input
                  class="form-control smart-pick-date-input"
                  type="text"
                  :name="pair.startName"
                  :value="fieldValue(pair.startKey)"
                  placeholder="YYYY-MM-DD"
                  inputmode="numeric"
                  maxlength="10"
                  pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                  :data-explicit="explicitFlag(pair.startName)"
                  :data-initial-value="fieldValue(pair.startKey)"
                />
              </div>
              <div class="col-6">
                <input
                  class="form-control smart-pick-date-input"
                  type="text"
                  :name="pair.endName"
                  :value="fieldValue(pair.endKey)"
                  placeholder="YYYY-MM-DD"
                  inputmode="numeric"
                  maxlength="10"
                  pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                  :data-explicit="explicitFlag(pair.endName)"
                  :data-initial-value="fieldValue(pair.endKey)"
                />
              </div>
            </div>
          </div>

          <div v-for="pair in numberPairs" :key="pair.label" class="col-12 col-xl-4">
            <label class="form-label mb-1">{{ pair.label }}</label>
            <div class="row g-2">
              <div class="col-6">
                <input
                  class="form-control"
                  type="number"
                  :step="pair.step"
                  :name="pair.startName"
                  :value="fieldValue(pair.startKey)"
                  placeholder="最小值"
                  :data-explicit="explicitFlag(pair.startName)"
                  :data-initial-value="fieldValue(pair.startKey)"
                />
              </div>
              <div class="col-6">
                <input
                  class="form-control"
                  type="number"
                  :step="pair.step"
                  :name="pair.endName"
                  :value="fieldValue(pair.endKey)"
                  placeholder="最大值"
                  :data-explicit="explicitFlag(pair.endName)"
                  :data-initial-value="fieldValue(pair.endKey)"
                />
              </div>
            </div>
          </div>

          <div class="col-12 d-flex gap-2">
            <button class="btn btn-primary" type="submit">筛选</button>
            <a class="btn btn-outline-secondary" :href="clearHref">清空</a>
          </div>
        </form>
      </div>
    </div>
  </div>

  <form
    v-else-if="hasFilter"
    :id="formId"
    class="filter-panel"
    method="get"
    :action="action"
    @submit.prevent="handleSubmit"
  >
    <input
      v-for="item in hiddenFields"
      :key="item.name"
      type="hidden"
      :name="item.name"
      :value="item.value"
      :data-explicit="explicitFlag(item.name)"
      :data-initial-value="item.value"
    />

    <div class="filter-grid align-end">
      <label v-for="field in textFields" :key="field.name" class="filter-field">
        <span>{{ field.label }}</span>
        <input
          :type="field.type || 'text'"
          :name="field.name"
          :value="fieldValue(field.key)"
          :placeholder="field.placeholder || ''"
          :step="field.step"
          :data-explicit="explicitFlag(field.name)"
          :data-initial-value="fieldValue(field.key)"
        />
      </label>

      <label v-for="field in selectFields" :key="field.name" class="filter-field">
        <span>{{ field.label }}</span>
        <select
          :name="field.name"
          :value="fieldValue(field.key)"
          :data-explicit="explicitFlag(field.name)"
          :data-initial-value="fieldValue(field.key)"
        >
          <option v-for="option in field.options" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
      </label>

      <label v-for="field in extraTextFields" :key="field.name" class="filter-field">
        <span>{{ field.label }}</span>
        <input
          :type="field.type || 'text'"
          :name="field.name"
          :value="fieldValue(field.key)"
          :placeholder="field.placeholder || ''"
          :data-explicit="explicitFlag(field.name)"
          :data-initial-value="fieldValue(field.key)"
        />
      </label>

      <div v-for="pair in datePairs" :key="pair.label" class="filter-field wide smart-pick-date-field">
        <span>{{ pair.label }}</span>
        <div class="filter-pair">
          <input
            class="smart-pick-date-input"
            type="text"
            :name="pair.startName"
            :value="fieldValue(pair.startKey)"
            placeholder="YYYY-MM-DD"
            inputmode="numeric"
            maxlength="10"
            pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
            :data-explicit="explicitFlag(pair.startName)"
            :data-initial-value="fieldValue(pair.startKey)"
          />
          <input
            class="smart-pick-date-input"
            type="text"
            :name="pair.endName"
            :value="fieldValue(pair.endKey)"
            placeholder="YYYY-MM-DD"
            inputmode="numeric"
            maxlength="10"
            pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
            :data-explicit="explicitFlag(pair.endName)"
            :data-initial-value="fieldValue(pair.endKey)"
          />
        </div>
      </div>

      <div v-for="pair in numberPairs" :key="pair.label" class="filter-field wide">
        <span>{{ pair.label }}</span>
        <div class="filter-pair">
          <input
            type="number"
            :step="pair.step"
            :name="pair.startName"
            :value="fieldValue(pair.startKey)"
            placeholder="最小值"
            :data-explicit="explicitFlag(pair.startName)"
            :data-initial-value="fieldValue(pair.startKey)"
          />
          <input
            type="number"
            :step="pair.step"
            :name="pair.endName"
            :value="fieldValue(pair.endKey)"
            placeholder="最大值"
            :data-explicit="explicitFlag(pair.endName)"
            :data-initial-value="fieldValue(pair.endKey)"
          />
        </div>
      </div>
    </div>

    <div class="filter-actions">
      <button class="button-primary" type="submit">筛选</button>
      <a class="button-secondary" :href="clearHref">清空</a>
    </div>
  </form>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { MovieRecord } from "@/features/movies/types";
import { asRecord, text } from "@/shared/utils/format";

const props = defineProps<{
  filter?: MovieRecord;
  formId?: string;
  variant?: string;
  open?: boolean;
  collapseId?: string;
}>();

type FieldOption = {
  value: string;
  label: string;
};

type TextField = {
  name: string;
  key: string;
  label: string;
  placeholder?: string;
  type?: string;
  step?: string;
};

type SelectField = TextField & {
  options: FieldOption[];
};

type PairField = {
  label: string;
  startName: string;
  startKey: string;
  endName: string;
  endKey: string;
  step?: string;
};

const filter = computed(() => asRecord(props.filter));
const formId = computed(() => props.formId || "movieCardFilterForm");
const collapseId = computed(() => props.collapseId || "movieCardFilterCollapse");
const isCardsAggVariant = computed(() => props.variant === "cards-agg");
const explicit = computed(() => asRecord(filter.value.Explicit));
const hasFilter = computed(() => Object.keys(filter.value).length > 0);
const action = computed(() => text(filter.value, "Action") || "/listcard");
const clearHref = computed(() => text(filter.value, "ClearHref") || action.value);
const open = computed(() => Boolean(props.open));

const hiddenFields = computed(() =>
  [
    { name: "ps", value: text(filter.value, "PageSize") },
    { name: "od", value: text(filter.value, "OrderBy") },
    { name: "n", value: text(filter.value, "RandomCount") },
    { name: "pid", value: text(filter.value, "PersonIds") }
  ].filter((item) => item.value)
);

const textFields: TextField[] = [
  { name: "cn", key: "CastNames", label: "演员", placeholder: "演员名，支持多值" },
  { name: "gn", key: "GenreNames", label: "类型", placeholder: "类型名，支持多值" },
  { name: "dn", key: "DirectorName", label: "导演", placeholder: "导演名" },
  { name: "pn", key: "PrefixName", label: "前缀", placeholder: "前缀" },
  { name: "mn", key: "MakerName", label: "片商", placeholder: "片商" },
  { name: "ln", key: "LabelName", label: "标签", placeholder: "标签" },
  { name: "lj", key: "LabelJavID", label: "标签 JavID", placeholder: "JavLib 标签 ID" },
  { name: "an", key: "AlbumName", label: "相册名", placeholder: "相册名，支持多值" },
  { name: "wd", key: "Word", label: "关键词", placeholder: "中文关键词" }
];

const selectFields: SelectField[] = [
  {
    name: "mowned",
    key: "MediaOwned",
    label: "w_media 库存",
    options: [
      { value: "", label: "全部" },
      { value: "1", label: "全部电影" },
      { value: "2", label: "有 w_media 记录（含移除）" },
      { value: "3", label: "已拥有未移除" },
      { value: "4", label: "有字幕未移除" },
      { value: "5", label: "无字幕未移除" },
      { value: "6", label: "已移除" },
      { value: "7", label: "无 w_media 记录" }
    ]
  },
  {
    name: "nd",
    key: "NeedDownload",
    label: "下载状态",
    options: [
      { value: "", label: "全部" },
      { value: "1", label: "无需下载" },
      { value: "2", label: "需要下载" }
    ]
  }
];

const baseExtraTextFields: TextField[] = [
  { name: "drkmin", key: "DaysInRankMin", label: "在榜天数", placeholder: "最小值", type: "number" }
];

const dirFields: TextField[] = [
  { name: "md1", key: "MediaDir1", label: "Media 目录 1", placeholder: "目录名" },
  { name: "md2", key: "MediaDir2", label: "Media 目录 2", placeholder: "目录名" },
  { name: "md3", key: "MediaDir3", label: "Media 目录 3", placeholder: "目录名" },
  { name: "md4", key: "MediaDir4", label: "Media 目录 4", placeholder: "目录名" }
];

const extraTextFields = computed(() => {
  if (filter.value.HideDirs) {
    return baseExtraTextFields;
  }
  return [...baseExtraTextFields, ...dirFields];
});

const datePairs: PairField[] = [
  { label: "上映日期", startName: "rs", startKey: "ReleasingDateStart", endName: "re", endKey: "ReleasingDateEnd" },
  {
    label: "Media 下载时间",
    startName: "mbs",
    startKey: "MediaBirthTimeStart",
    endName: "mbe",
    endKey: "MediaBirthTimeEnd"
  },
  {
    label: "上榜日期",
    startName: "srds",
    startKey: "StartRankingDateStart",
    endName: "srde",
    endKey: "StartRankingDateEnd"
  },
  {
    label: "最近入选 SC 时间",
    startName: "lsstmin",
    startKey: "LastSelectedScTimeMin",
    endName: "lsstmax",
    endKey: "LastSelectedScTimeMax"
  }
];

const numberPairs: PairField[] = [
  { label: "演员年龄", startName: "cay", startKey: "CastAgeMin", endName: "cao", endKey: "CastAgeMax", step: "0.1" },
  { label: "观看人数", startName: "vwmin", startKey: "ViewWatchedMin", endName: "vwmax", endKey: "ViewWatchedMax" },
  { label: "评分", startName: "smin", startKey: "ScoreMin", endName: "smax", endKey: "ScoreMax", step: "0.1" },
  { label: "SC 次数", startName: "scmin", startKey: "ScTimesMin", endName: "scmax", endKey: "ScTimesMax" },
  { label: "Come 次数", startName: "comin", startKey: "ComeTimesMin", endName: "comax", endKey: "ComeTimesMax" }
];

function fieldValue(key: string): string {
  return text(filter.value, key);
}

function explicitFlag(name: string): string {
  return explicit.value[name] ? "1" : "0";
}

function normalizeValue(element: Element): string {
  return "value" in element ? String(element.value || "").trim() : "";
}

function shouldSkipElement(element: Element): boolean {
  const control = element as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement | HTMLButtonElement;
  if (!control.name || control.disabled) {
    return true;
  }
  const type = String("type" in control ? control.type : control.tagName).toLowerCase();
  return type === "submit" || type === "button" || type === "reset" || type === "fieldset";
}

function handleSubmit(event: Event) {
  const form = event.currentTarget as HTMLFormElement;
  const params = new URLSearchParams();
  Array.from(form.elements).forEach((element) => {
    if (shouldSkipElement(element)) {
      return;
    }
    const control = element as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement;
    const value = normalizeValue(control);
    const type = String("type" in control ? control.type : control.tagName).toLowerCase();
    const isExplicit = control.dataset.explicit === "1";
    const initialValue = control.dataset.initialValue || "";
    if ((type === "checkbox" || type === "radio") && "checked" in control && !control.checked) {
      return;
    }
    if (type === "hidden") {
      if (isExplicit && value !== "") {
        params.set(control.name, value);
      }
      return;
    }
    if (value === "") {
      return;
    }
    if (isExplicit || value !== initialValue) {
      params.set(control.name, value);
    }
  });
  const target = form.getAttribute("action") || window.location.pathname;
  const query = params.toString();
  window.location.assign(query ? `${target}?${query}` : target);
}
</script>
