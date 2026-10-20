<template>
  <div class="card border pick-group smart-pick-group-panel">
    <div class="card-body">
      <div class="smart-pick-group-head">
        <h6>{{ kind === "jav" ? "JAV" : "FC2" }} 条件组 {{ index + 1 }}</h6>
        <button
          type="button"
          class="btn btn-sm btn-outline-danger btn-remove-group smart-pick-remove-group"
          :disabled="!canRemove"
          @click="$emit('remove')"
        >
          移除
        </button>
      </div>

      <div class="row g-3 smart-pick-group-grid">
        <div class="col-md-3 smart-pick-field">
          <label class="form-label" :for="fieldID('weight')">权重</label>
          <input
            :id="fieldID('weight')"
            class="form-control"
            type="number"
            min="1"
            step="1"
            :value="modelValue.weight"
            @input="updateWeight(($event.target as HTMLInputElement).value)"
          />
        </div>
        <FieldControl
          v-for="field in basicFields"
          :key="field.key"
          class="col-md-3"
          :control-id="fieldID(field.key)"
          :field="field"
          :value="valueOf(field.key)"
          @update="updateFilter(field, $event)"
        />
        <div class="col-md-3 d-flex align-items-end smart-pick-advanced-toggle-wrap">
          <button
            type="button"
            class="btn btn-sm btn-outline-secondary w-100 smart-pick-advanced-toggle"
            :aria-controls="advancedPanelID"
            :aria-expanded="modelValue.expanded ? 'true' : 'false'"
            @click="toggleExpanded"
          >
            高级参数
          </button>
        </div>
        <div
          :id="advancedPanelID"
          class="col-12 collapse smart-pick-advanced-panel"
          :class="{ show: modelValue.expanded }"
        >
          <div class="row g-3 smart-pick-advanced-grid">
            <FieldControl
              v-for="field in advancedFields"
              :key="field.key"
              class="col-md-4"
              :control-id="fieldID(field.key)"
              :field="field"
              :value="valueOf(field.key)"
              @update="updateFilter(field, $event)"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, type PropType } from "vue";
import type { ScPickV2FilterValue, ScPickV2Group } from "@/features/sc/types";

type FieldOption = { value: string; label: string };
type FieldDefinition = {
  key: string;
  label: string;
  type?: "text" | "number";
  step?: number;
  placeholder?: string;
  options?: FieldOption[];
};

const props = defineProps<{
  kind: "jav" | "fc2";
  index: number;
  modelValue: ScPickV2Group;
  canRemove: boolean;
}>();
const emit = defineEmits<{ "update:modelValue": [value: ScPickV2Group]; remove: [] }>();

const basicFields = computed<FieldDefinition[]>(() =>
  props.kind === "jav"
    ? [
        { key: "ReleasingDateStart", label: "上映日期起", placeholder: "YYYY-MM-DD" },
        { key: "ReleasingDateEnd", label: "上映日期止", placeholder: "YYYY-MM-DD" }
      ]
    : [
        { key: "PostTimeStart", label: "发布日期起", placeholder: "YYYY-MM-DD" },
        { key: "PostTimeEnd", label: "发布日期止", placeholder: "YYYY-MM-DD" }
      ]
);

const javAdvancedFields: FieldDefinition[] = [
  { key: "MediaBirthTimeStart", label: "下载日期起", placeholder: "YYYY-MM-DD" },
  { key: "MediaBirthTimeEnd", label: "下载日期止", placeholder: "YYYY-MM-DD" },
  { key: "ScTimesMin", label: "SC 次数最小", type: "number", step: 1 },
  { key: "ScTimesMax", label: "SC 次数最大", type: "number", step: 1 },
  { key: "ComeTimesMin", label: "Come 次数最小", type: "number", step: 1 },
  { key: "ComeTimesMax", label: "Come 次数最大", type: "number", step: 1 },
  { key: "CastNames", label: "演员" },
  { key: "GenreNames", label: "类型" },
  { key: "DirectorName", label: "导演" },
  { key: "PrefixName", label: "前缀" },
  { key: "MakerName", label: "制作商" },
  { key: "LabelName", label: "厂牌" },
  { key: "MediaDir1", label: "媒体目录 1" },
  { key: "MediaDir2", label: "媒体目录 2" },
  { key: "MediaDir3", label: "媒体目录 3" },
  { key: "MediaDir4", label: "媒体目录 4" },
  { key: "Word", label: "关键字" },
  { key: "ScoreMin", label: "评分最小", type: "number", step: 0.1 },
  { key: "ScoreMax", label: "评分最大", type: "number", step: 0.1 },
  { key: "HighestRankMin", label: "最高排名最小", type: "number", step: 1 },
  { key: "HighestRankMax", label: "最高排名最大", type: "number", step: 1 },
  { key: "ViewWatchedMin", label: "观看数最小", type: "number", step: 1 },
  { key: "ViewWatchedMax", label: "观看数最大", type: "number", step: 1 },
  { key: "LastSelectedScTimeMin", label: "最近入选 SC 起", placeholder: "YYYY-MM-DD" },
  { key: "LastSelectedScTimeMax", label: "最近入选 SC 止", placeholder: "YYYY-MM-DD" }
];

const fc2AdvancedFields: FieldDefinition[] = [
  { key: "MediaBirthTimeStart", label: "下载日期起", placeholder: "YYYY-MM-DD" },
  { key: "MediaBirthTimeEnd", label: "下载日期止", placeholder: "YYYY-MM-DD" },
  { key: "ScTimesMin", label: "SC 次数最小", type: "number", step: 1 },
  { key: "ScTimesMax", label: "SC 次数最大", type: "number", step: 1 },
  { key: "ComeTimesMin", label: "Come 次数最小", type: "number", step: 1 },
  { key: "ComeTimesMax", label: "Come 次数最大", type: "number", step: 1 },
  { key: "LastSelectedScTimeMin", label: "最近入选 SC 起", placeholder: "YYYY-MM-DD" },
  { key: "LastSelectedScTimeMax", label: "最近入选 SC 止", placeholder: "YYYY-MM-DD" },
  { key: "FC2Key", label: "FC2 Key" },
  { key: "Code", label: "FC2 Code" },
  { key: "Title", label: "标题" },
  { key: "PerformerText", label: "表演者" },
  { key: "MaskType", label: "遮罩类型" },
  { key: "Word", label: "综合关键字" },
  { key: "MediaDir1", label: "媒体目录 1" },
  { key: "MediaDir2", label: "媒体目录 2" },
  { key: "MediaDir3", label: "媒体目录 3" },
  { key: "MediaDir4", label: "媒体目录 4" },
  {
    key: "HasSub",
    label: "字幕",
    options: [
      { value: "", label: "全部" },
      { value: "1", label: "无字幕" },
      { value: "2", label: "有字幕" }
    ]
  },
  {
    key: "SelfMake",
    label: "自制",
    options: [
      { value: "", label: "全部" },
      { value: "1", label: "非自制" },
      { value: "2", label: "自制" }
    ]
  },
  {
    key: "HasMask",
    label: "去码状态",
    options: [
      { value: "", label: "全部" },
      { value: "1", label: "未去码" },
      { value: "2", label: "已去码" },
      { value: "3", label: "无码" }
    ]
  }
];

const advancedFields = computed(() => (props.kind === "jav" ? javAdvancedFields : fc2AdvancedFields));
const advancedPanelID = computed(() => `scpickv2-${props.kind}-${props.modelValue.id}-advanced`);

function fieldID(key: string): string {
  return `scpickv2-${props.kind}-${props.modelValue.id}-${key.toLowerCase()}`;
}

function valueOf(key: string): ScPickV2FilterValue {
  return props.modelValue.filter[key] ?? "";
}

function updateWeight(raw: string) {
  emit("update:modelValue", { ...props.modelValue, weight: Number(raw) });
}

function updateFilter(field: FieldDefinition, raw: string) {
  let value: ScPickV2FilterValue = raw;
  if (field.type === "number" || field.options) value = raw === "" ? null : Number(raw);
  emit("update:modelValue", {
    ...props.modelValue,
    filter: { ...props.modelValue.filter, [field.key]: value }
  });
}

function toggleExpanded() {
  emit("update:modelValue", { ...props.modelValue, expanded: !props.modelValue.expanded });
}

const FieldControl = defineComponent({
  props: {
    controlId: { type: String, required: true },
    field: { type: Object as PropType<FieldDefinition>, required: true },
    value: { type: [String, Number] as PropType<ScPickV2FilterValue>, default: "" }
  },
  emits: ["update"],
  setup(fieldProps, { emit: fieldEmit }) {
    return () =>
      h("div", { class: "smart-pick-field" }, [
        h("label", { class: "form-label", for: fieldProps.controlId }, fieldProps.field.label),
        fieldProps.field.options
          ? h(
              "select",
              {
                id: fieldProps.controlId,
                class: "form-select",
                value: fieldProps.value ?? "",
                onChange: (event: Event) => fieldEmit("update", (event.target as HTMLSelectElement).value)
              },
              fieldProps.field.options.map((option) => h("option", { value: option.value }, option.label))
            )
          : h("input", {
              id: fieldProps.controlId,
              class: "form-control",
              type: fieldProps.field.type || "text",
              min: fieldProps.field.type === "number" ? 0 : undefined,
              step: fieldProps.field.step,
              value: fieldProps.value ?? "",
              placeholder: fieldProps.field.placeholder,
              inputmode: fieldProps.field.placeholder === "YYYY-MM-DD" ? "numeric" : undefined,
              maxlength: fieldProps.field.placeholder === "YYYY-MM-DD" ? 10 : undefined,
              pattern: fieldProps.field.placeholder === "YYYY-MM-DD" ? "[0-9]{4}-[0-9]{2}-[0-9]{2}" : undefined,
              onInput: (event: Event) => fieldEmit("update", (event.target as HTMLInputElement).value)
            })
      ]);
  }
});
</script>
