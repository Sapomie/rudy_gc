<template>
  <section
    class="smart-pick-type-options"
    :class="kind === 'jav' ? 'is-jav' : 'is-fc2'"
    :aria-label="kind === 'jav' ? 'JAV 抽取策略' : 'FC2 抽取策略'"
  >
    <h6>{{ kind === "jav" ? "JAV 抽取策略" : "FC2 抽取策略" }}</h6>
    <div class="row g-3 smart-pick-advanced-grid">
      <div v-for="field in fields" :key="field.key" class="col-md-6 smart-pick-field">
        <label class="form-label" :for="fieldID(field.key)">{{ field.label }}</label>
        <input
          :id="fieldID(field.key)"
          type="number"
          class="form-control"
          :min="field.min"
          :step="field.step"
          :value="valueOf(field.key)"
          :placeholder="field.optional ? '可为空' : ''"
          @input="update(field.key, ($event.target as HTMLInputElement).value, field.optional)"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { ScPickV2FC2Options, ScPickV2JAVOptions } from "@/features/sc/types";

type StrategyModel =
  (ScPickV2JAVOptions & { randomSeed?: number | null }) | (ScPickV2FC2Options & { randomSeed?: number | null });
type StrategyField = { key: string; label: string; min: number; step: number; optional?: boolean };

const props = defineProps<{ kind: "jav" | "fc2"; modelValue: StrategyModel }>();
const emit = defineEmits<{ "update:modelValue": [value: StrategyModel] }>();

const fields = computed<StrategyField[]>(() =>
  props.kind === "jav"
    ? [
        { key: "castParticipatedScBlockDays", label: "主演参与阻断天数", min: 0, step: 1 },
        { key: "castSelectedScBlockDays", label: "主演入选阻断天数", min: 0, step: 1 },
        { key: "rank20Min", label: "Rank20 累计保底", min: 0, step: 1 },
        { key: "rank100Min", label: "Rank100 累计保底", min: 0, step: 1 },
        { key: "rank500Min", label: "Rank500 累计保底", min: 0, step: 1 },
        { key: "castParticipatedScPenaltyAlpha", label: "主演参与惩罚", min: 0, step: 0.1 },
        { key: "castSelectedScPenaltyAlpha", label: "主演入选惩罚", min: 0, step: 0.1 },
        { key: "castOwnedScRatioPenaltyAlpha", label: "主演入选比例惩罚", min: 0, step: 0.1 },
        { key: "movieHasSelectedScPenaltyAlpha", label: "影片已入选惩罚", min: 0, step: 0.1 },
        { key: "randomSeed", label: "随机种子", min: 0, step: 1, optional: true }
      ]
    : [
        { key: "movieParticipatedScBlockDays", label: "最近参与 SC 阻断天数", min: 0, step: 1 },
        { key: "movieSelectedScBlockDays", label: "最近入选 SC 阻断天数", min: 0, step: 1 },
        { key: "movieSelectedScPenaltyAlpha", label: "最近入选时间惩罚", min: 0, step: 0.1 },
        { key: "movieHasSelectedScPenaltyAlpha", label: "历史入选次数惩罚", min: 0, step: 0.1 },
        { key: "randomSeed", label: "随机种子", min: 0, step: 1, optional: true }
      ]
);

function valueOf(key: string): number | "" {
  const value = (props.modelValue as unknown as Record<string, number | null | undefined>)[key];
  return value === null || value === undefined ? "" : value;
}

function fieldID(key: string): string {
  return `scpickv2-${props.kind}-${key}`;
}

function update(key: string, raw: string, optional = false) {
  const parsed = raw.trim() === "" && optional ? null : Number(raw);
  emit("update:modelValue", { ...props.modelValue, [key]: parsed } as StrategyModel);
}
</script>
