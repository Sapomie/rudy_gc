<template>
  <PageShell class="sc-events-page-shell sc-pick-page-shell" :title="title" eyebrow="">
    <div id="page-root" class="container-fluid sc-events-page-root sc-pick-page-root">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div class="d-flex align-items-center justify-content-between mb-3 sc-page-head">
          <div>
            <h4 class="mb-1">{{ pageTitle }}</h4>
            <p class="text-muted mb-0">{{ pageNote }}</p>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <p class="text-muted mb-2">按多组条件从本地已拥有影片里抽片，并叠加 Cast SC 阻断、Rank 配额和演员去重规则。</p>
            <p class="text-muted small mb-3">{{ fixedHint }}</p>

            <form id="formPickSmart" class="needs-validation" novalidate :data-source="pageSource">
              <div class="card border mb-3">
                <div class="card-body">
                  <h6 class="mb-3">全局抽取策略</h6>
                  <div class="row g-3 align-items-end">
                    <FieldInput v-for="field in globalFields" :key="field.id" col-class="col-md-2" v-bind="field" />
                  </div>
                </div>
              </div>

              <div class="row g-3 align-items-center mb-3">
                <div class="col-auto">
                  <button type="button" id="btnAddGroup" class="btn btn-outline-secondary" @click="addGroup">添加一组条件</button>
                </div>
              </div>

              <div id="pickGroups">
                <PickGroup
                  v-for="group in groups"
                  :key="group.id"
                  :payload="payload"
                  :group-id="group.id"
                  :can-remove="groups.length > 1"
                  :expanded="group.expanded"
                  @remove="removeGroup(group.id)"
                  @toggle="toggleGroup(group.id)"
                />
              </div>

              <div class="mt-3">
                <button type="submit" class="btn btn-primary">执行 Smart Pick + Copy</button>
                <button type="button" class="btn btn-outline-primary ms-2" id="btnPickOnly">仅 Smart Pick</button>
              </div>
            </form>
          </div>
        </div>

        <div class="card mt-3" id="copyCard" style="display: none;">
          <div class="card-body">
            <div class="d-flex align-items-center justify-content-between mb-2">
              <h5 class="mb-0">Copy 进度</h5>
              <button type="button" class="btn btn-sm btn-outline-danger" id="btnStopCopy">停止 Copy</button>
            </div>
            <div class="text-muted small mb-2" id="copyMeta"></div>
            <div class="progress mb-2" style="height: 6px;">
              <div class="progress-bar" role="progressbar" id="copyProgress" style="width: 0%;"></div>
            </div>
            <div class="small text-muted" id="copyCurrent"></div>
            <div class="small text-danger mt-1" id="copyError" style="display: none;"></div>
          </div>
        </div>

        <div class="card mt-3" id="pickInfoCard" style="display: none;">
          <div class="card-body">
            <div class="d-flex align-items-center justify-content-between flex-wrap gap-2 mb-3">
              <h5 class="mb-0">Pick 概览</h5>
              <span class="text-muted small" id="pickInfoSummary"></span>
            </div>
            <div class="row g-3" id="pickInfoGrid"></div>
          </div>
        </div>

        <div class="card mt-3" id="pickedCard" style="display: none;">
          <div class="card-body">
            <div class="d-flex align-items-center justify-content-between flex-wrap gap-2 mb-3">
              <h5 class="mb-0">抽取结果</h5>
              <div class="d-flex align-items-center flex-wrap gap-2">
                <div class="btn-group btn-group-sm flex-wrap" id="pickedSortBar" role="group" aria-label="抽取结果排序">
                  <button v-for="item in pickedSortButtons" :key="item.key" type="button" class="btn btn-outline-primary" :class="{ active: item.active }" :data-sort-base="item.key" :data-default-dir="item.dir">{{ item.label }}</button>
                </div>
                <span class="text-muted small" id="pickedCount"></span>
              </div>
            </div>
            <div class="row row-cols-2 row-cols-sm-3 row-cols-lg-5 g-3 picked-grid-8" id="pickedGrid"></div>
          </div>
        </div>

        <template id="pickGroupTpl">
          <PickGroup :payload="payload" />
        </template>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch, type PropType } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON } from "@/api/request";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asRecord, text } from "@/utils/format";

type FieldConfig = {
  id: string;
  label: string;
  type?: string;
  min?: string;
  step?: string;
  value?: string;
  placeholder?: string;
  reqField?: string;
  dataType?: string;
  role?: string;
  required?: boolean;
};

type PickGroupState = {
  id: number;
  expanded: boolean;
};

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const groups = ref<PickGroupState[]>([{ id: 1, expanded: false }]);
const nextGroupId = ref(2);

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

const title = computed(() => text(payload.value, "Title") || "SC Pick Smart Media");
const pageTitle = computed(() => text(payload.value, "PageTitle") || title.value);
const pageNote = computed(() => text(payload.value, "PageNote"));
const pageSource = computed(() => (text(payload.value, "PageSource").toLowerCase() === "wmedia" ? "wmedia" : "vfilm"));
const fixedHint = computed(() => text(payload.value, "FixedHint"));
const castOwnedScRatioPenaltyAlpha = computed(() => pageSource.value === "wmedia" ? "0" : "0.7");

const globalFields = computed<FieldConfig[]>(() => [
  { id: "pickCount", label: "抽取数量", type: "number", min: "1", value: "32", required: true },
  { id: "castLastScBlockDays", label: "Cast LastSc 阻断天数", type: "number", min: "0", value: "60", required: true },
  { id: "lastScEventBlockDays", label: "LastSc Event 阻断天数", type: "number", min: "0", value: "30", required: true },
  { id: "rank20Min", label: "Rank20 至少", type: "number", min: "0", value: "0", required: true },
  { id: "rank100Min", label: "Rank100 至少", type: "number", min: "0" },
  { id: "rank500Min", label: "Rank500 至少", type: "number", min: "0" },
  { id: "castLastScPenaltyAlpha", label: "Cast LastSc 惩罚系数", type: "number", min: "0", step: "0.1", value: "0.7", required: true },
  { id: "lastScEventPenaltyAlpha", label: "LastSc Event 惩罚系数", type: "number", min: "0", step: "0.1", value: "0.7", required: true },
  { id: "castOwnedScRatioPenaltyAlpha", label: "Cast SC Ratio 惩罚系数", type: "number", min: "0", step: "0.1", value: castOwnedScRatioPenaltyAlpha.value, required: true },
  { id: "movieHasScPenaltyAlpha", label: "Movie Has SC 惩罚系数", type: "number", min: "0", step: "0.1", value: "0.7", required: true }
]);

const pickedSortButtons = [
  { key: "picked", label: "抽取顺序", dir: "desc", active: false },
  { key: "film_birth", label: "下载时间", dir: "desc", active: false },
  { key: "releasing", label: "上映日期", dir: "desc", active: true },
  { key: "sc", label: "SC次数", dir: "desc", active: false },
  { key: "come", label: "CO次数", dir: "desc", active: false },
  { key: "rank", label: "最高排名", dir: "asc", active: false },
  { key: "score", label: "评分", dir: "desc", active: false }
];

function addGroup() {
  groups.value.push({ id: nextGroupId.value, expanded: false });
  nextGroupId.value += 1;
}

function removeGroup(id: number) {
  if (groups.value.length <= 1) return;
  groups.value = groups.value.filter((group) => group.id !== id);
}

function toggleGroup(id: number) {
  groups.value = groups.value.map((group) => group.id === id ? { ...group, expanded: !group.expanded } : group);
}

function dynamicLabel(row: AnyRecord, key: string, fallback: string): string {
  return text(row, key) || fallback;
}

function dynamicField(row: AnyRecord, key: string, fallback: string): string {
  return text(row, key) || fallback;
}

const FieldInput = defineComponent({
  name: "SmartPickFieldInput",
  props: {
    colClass: { type: String, default: "" },
    id: { type: String, required: true },
    label: { type: String, required: true },
    type: { type: String, default: "text" },
    min: { type: String, default: "" },
    step: { type: String, default: "" },
    value: { type: String, default: "" },
    placeholder: { type: String, default: "" },
    reqField: { type: String, default: "" },
    dataType: { type: String, default: "" },
    role: { type: String, default: "" },
    required: { type: Boolean, default: false }
  },
  setup(props) {
    return () => h("div", { class: props.colClass || "col-md-3" }, [
      h("label", { class: "form-label", for: props.id }, props.label),
      h("input", {
        type: props.type,
        min: props.min || undefined,
        step: props.step || undefined,
        class: "form-control",
        id: props.id,
        value: props.value,
        placeholder: props.placeholder || undefined,
        "data-req-field": props.reqField || undefined,
        "data-type": props.dataType || undefined,
        "data-role": props.role || undefined,
        required: props.required || undefined,
        inputmode: props.placeholder === "YYYY-MM-DD" ? "numeric" : undefined,
        maxlength: props.placeholder === "YYYY-MM-DD" ? "10" : undefined,
        pattern: props.placeholder === "YYYY-MM-DD" ? "[0-9]{4}-[0-9]{2}-[0-9]{2}" : undefined
      })
    ]);
  }
});

const PickGroup = defineComponent({
  name: "SmartPickGroup",
  props: {
    payload: { type: Object as PropType<AnyRecord>, required: true },
    groupId: { type: Number, required: true },
    canRemove: { type: Boolean, default: true },
    expanded: { type: Boolean, default: false }
  },
  emits: ["remove", "toggle"],
  setup(props, { emit }) {
    const fieldId = (name: string) => `smart-${props.groupId}-${name}`;
    const advPanelId = () => `smart-adv-panel-${props.groupId}`;
    const simpleFields = () => [
      { id: fieldId("rs"), label: "放映日期起 (rs)", reqField: "ReleasingDateStart", placeholder: "YYYY-MM-DD" },
      { id: fieldId("re"), label: "放映日期止 (re)", reqField: "ReleasingDateEnd", placeholder: "YYYY-MM-DD" },
      { id: fieldId("birth-start"), label: dynamicLabel(props.payload, "BirthStartLabel", "下载日期起 (mbs)"), reqField: dynamicField(props.payload, "BirthStartField", "MediaBirthTimeStart"), placeholder: "YYYY-MM-DD" },
      { id: fieldId("birth-end"), label: dynamicLabel(props.payload, "BirthEndLabel", "下载日期止 (mbe)"), reqField: dynamicField(props.payload, "BirthEndField", "MediaBirthTimeEnd"), placeholder: "YYYY-MM-DD" },
      { id: fieldId("scmin"), label: "ScTimes 最小 (scmin)", type: "number", reqField: "ScTimesMin", dataType: "int-opt", placeholder: "可为空" },
      { id: fieldId("scmax"), label: "ScTimes 最大 (scmax)", type: "number", reqField: "ScTimesMax", dataType: "int-opt", placeholder: "可为空" },
      { id: fieldId("comin"), label: "ComeTimes 最小 (comin)", type: "number", reqField: "ComeTimesMin", dataType: "int-opt", placeholder: "可为空" },
      { id: fieldId("comax"), label: "ComeTimes 最大 (comax)", type: "number", reqField: "ComeTimesMax", dataType: "int-opt", placeholder: "可为空" }
    ];
    const advFields = () => [
      { reqField: "CastNames", label: "演员 (cn)" },
      { reqField: "GenreNames", label: "类型 (gn)" },
      { reqField: "DirectorName", label: "导演 (dn)" },
      { reqField: "PrefixName", label: "前缀 (pn)" },
      { reqField: "MakerName", label: "制作商 (mn)" },
      { reqField: "LabelName", label: "厂牌 (ln)" },
      { reqField: dynamicField(props.payload, "Dir4Field", "MediaDir4"), label: dynamicLabel(props.payload, "Dir4Label", "MediaDir4 (md4)") },
      { reqField: "Word", label: "关键字 (wd)", placeholder: "可为空" },
      { reqField: "ScoreMin", label: "评分最小 (smin)", type: "number", step: "0.1", dataType: "float", colClass: "col-md-3" },
      { reqField: "ScoreMax", label: "评分最大 (smax)", type: "number", step: "0.1", dataType: "float-opt", colClass: "col-md-3" },
      { reqField: "HighestRankMin", label: "最高排名最小", type: "number", dataType: "int-opt", colClass: "col-md-3" },
      { reqField: "HighestRankMax", label: "最高排名最大", type: "number", dataType: "int-opt", colClass: "col-md-3" },
      { reqField: "ViewWatchedMin", label: "观看最小 (vwmin)", type: "number", dataType: "int", colClass: "col-md-3" },
      { reqField: "ViewWatchedMax", label: "观看最大 (vwmax)", type: "number", dataType: "int-opt", colClass: "col-md-3" },
      { reqField: "LastScTimeMin", label: "最后评分起 (lsctmin)", placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: "LastScTimeMax", label: "最后评分止 (lsctmax)", placeholder: "YYYY-MM-DD", colClass: "col-md-3" },
      { reqField: dynamicField(props.payload, "Dir1Field", "MediaDir1"), label: dynamicLabel(props.payload, "Dir1Label", "MediaDir1 (md1)"), colClass: "col-md-3" },
      { reqField: dynamicField(props.payload, "Dir2Field", "MediaDir2"), label: dynamicLabel(props.payload, "Dir2Label", "MediaDir2 (md2)"), colClass: "col-md-3" },
      { reqField: dynamicField(props.payload, "Dir3Field", "MediaDir3"), label: dynamicLabel(props.payload, "Dir3Label", "MediaDir3 (md3)"), colClass: "col-md-3" },
      { reqField: "OrderBy", label: "排序 (od)", placeholder: "sc/lsct/bt...", colClass: "col-md-3" }
    ];
    return () => h("div", { class: "card border mb-3 pick-group" }, [
      h("div", { class: "card-body" }, [
        h("div", { class: "d-flex align-items-center justify-content-between mb-2" }, [
          h("h6", { class: "mb-0" }, "条件组"),
          h("button", {
            type: "button",
            class: "btn btn-sm btn-outline-danger btn-remove-group",
            disabled: !props.canRemove,
            onClick: () => emit("remove")
          }, "移除")
        ]),
        h("div", { class: "row g-3" }, [
          h(FieldInput, { colClass: "col-md-3", id: fieldId("weight"), label: "权重", type: "number", min: "1", value: "12", required: true, role: "weight" }),
          h("div", { class: "col-md-3 d-flex align-items-end" }, [
            h("button", {
              type: "button",
              class: "btn btn-sm btn-outline-secondary w-100",
              "data-role": "adv-toggle",
              "data-bs-toggle": "collapse",
              "data-bs-target": `#${advPanelId()}`,
              "aria-controls": advPanelId(),
              "aria-expanded": props.expanded ? "true" : "false",
              onClick: () => emit("toggle")
            }, "高级参数")
          ]),
          ...simpleFields().map((field) => h(FieldInput, { colClass: "col-md-3", ...field })),
          h("div", { class: `col-12 collapse${props.expanded ? " show" : ""}`, id: advPanelId(), "data-role": "adv-panel" }, [
            h("div", { class: "row g-3 mt-1" }, [
              ...advFields().map((field) => h(FieldInput, {
                colClass: field.colClass || "col-md-4",
                id: fieldId(field.reqField.toLowerCase()),
                label: field.label,
                type: field.type || "text",
                step: field.step || "",
                reqField: field.reqField,
                dataType: field.dataType || "",
                placeholder: field.placeholder || ""
              }))
            ])
          ])
        ])
      ])
    ]);
  }
});
</script>
