<template>
  <div
    :id="modalId"
    class="e-items-modal-overlay"
    :class="{ show }"
    tabindex="-1"
    role="dialog"
    aria-modal="true"
    :aria-labelledby="labelId"
    :aria-hidden="show ? 'false' : 'true'"
  >
    <div class="e-items-modal-dialog e-items-status-dialog">
      <div class="e-items-modal-content">
        <form :id="`${idPrefix}${batch ? 'StatusForm' : 'Form'}`" @submit.prevent="$emit('submit')">
          <header class="e-items-modal-header">
            <div>
              <span class="e-items-modal-eyebrow">
                {{ batch ? "E ITEMS / BATCH STATUS" : "E ITEMS / STATUS" }}
              </span>
              <h5 :id="labelId">{{ title }}</h5>
              <p :id="`${idPrefix}${batch ? 'StatusMeta' : 'Meta'}`">{{ meta }}</p>
            </div>
            <button type="button" class="e-items-modal-close" aria-label="关闭" @click="$emit('close')">
              <span aria-hidden="true">×</span>
            </button>
          </header>

          <div class="e-items-modal-body">
            <div class="e-items-status-grid">
              <div v-for="field in fields" :key="field.key" class="e-items-status-field">
                <label :for="`${idPrefix}${field.idSuffix}`">{{ field.label }}</label>
                <select
                  class="e-items-modal-select"
                  :id="`${idPrefix}${field.idSuffix}`"
                  :name="field.key"
                  :value="form[field.key]"
                  @change="update(field.key, $event)"
                >
                  <option v-for="option in field.options" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
              </div>
            </div>
            <div
              :id="`${idPrefix}${batch ? 'StatusMsg' : 'Msg'}`"
              class="e-items-modal-message"
              :class="message ? (messageOk ? 'is-success' : 'is-error') : 'd-none'"
            >
              {{ message }}
            </div>
          </div>

          <footer class="e-items-modal-footer">
            <button type="button" class="e-items-modal-cancel" @click="$emit('close')">取消</button>
            <button
              :id="`${idPrefix}${batch ? 'StatusSubmit' : 'Submit'}`"
              type="submit"
              class="e-items-modal-submit"
              :disabled="submitting"
            >
              {{ submitting ? submittingLabel : submitLabel }}
            </button>
          </footer>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { EItemStatusForm } from "./models";

type FieldKey = keyof EItemStatusForm;
interface FieldOption {
  value: string;
  label: string;
}
interface FieldDefinition {
  key: FieldKey;
  idSuffix: string;
  label: string;
  options: FieldOption[];
}

const props = withDefaults(
  defineProps<{
    idPrefix: string;
    title: string;
    submitLabel: string;
    submittingLabel: string;
    show: boolean;
    meta: string;
    message: string;
    messageOk: boolean;
    form: EItemStatusForm;
    batch?: boolean;
    submitting?: boolean;
  }>(),
  { batch: false, submitting: false }
);

const emit = defineEmits<{
  close: [];
  submit: [];
  "update-form": [form: EItemStatusForm];
}>();

const modalId = computed(() => `${props.idPrefix}${props.batch ? "StatusModal" : "Modal"}`);
const labelId = computed(() => `${modalId.value}Label`);
const optional = (options: FieldOption[]) => (props.batch ? [{ value: "", label: "不修改" }, ...options] : options);
const fields = computed<FieldDefinition[]>(() => [
  {
    key: "hasDetail",
    idSuffix: "HasDetail",
    label: "详情状态",
    options: optional([
      { value: "none", label: "无详情" },
      { value: "ok", label: "已有详情" }
    ])
  },
  {
    key: "hasDownloadCover",
    idSuffix: "HasDownloadCover",
    label: "封面状态",
    options: optional([
      { value: "none", label: "无本地封面" },
      { value: "ok", label: "已有本地封面" },
      { value: "wrong", label: "错误封面链接" }
    ])
  },
  {
    key: "hasChinese",
    idSuffix: "HasChinese",
    label: "中文字幕",
    options: optional([
      { value: "none", label: "无中文字幕" },
      { value: "ok", label: "有中文字幕" },
      { value: "error", label: "翻译错误" },
      { value: "sensitive", label: "含敏感词" }
    ])
  },
  {
    key: "detailNeedScan",
    idSuffix: "DetailNeedScan",
    label: "详情扫描状态",
    options: optional([
      { value: "need", label: "需要扫描" },
      { value: "no_need", label: "无需扫描" },
      { value: "wrong", label: "内容错误" }
    ])
  }
]);

function update(key: FieldKey, event: Event) {
  emit("update-form", { ...props.form, [key]: (event.target as HTMLSelectElement).value });
}
</script>
