<template>
  <Teleport to="body">
    <div v-if="active" class="modal-backdrop-v2 cafo-edit-overlay" role="presentation" @click.self="requestClose">
      <form
        ref="dialog"
        class="modal-panel-v2 cafo-edit-panel"
        role="dialog"
        aria-modal="true"
        aria-labelledby="cafoEditTitle"
        tabindex="-1"
        @submit.prevent="submit"
        @keydown.esc="requestClose"
      >
        <div class="modal-header-v2">
          <div>
            <h2 id="cafoEditTitle">编辑 c_cafo 资料</h2>
            <p>{{ personName }}</p>
          </div>
          <button type="button" class="modal-close-v2" aria-label="关闭" :disabled="saving" @click="requestClose">
            ×
          </button>
        </div>

        <div class="modal-body-v2 cafo-edit-body">
          <div v-if="errorMessage" class="alert alert-danger py-2 mb-0" role="alert">{{ errorMessage }}</div>
          <div v-if="successMessage" class="alert alert-success py-2 mb-0" role="status">{{ successMessage }}</div>
          <div v-if="loading" class="cafo-edit-state" aria-live="polite">正在读取关联资料...</div>
          <div v-else-if="cafos.length === 0" class="cafo-edit-state">当前人物没有关联的 c_cafo 资料。</div>
          <template v-else>
            <label v-if="cafos.length > 1" class="cafo-edit-field is-wide">
              <span>关联资料</span>
              <select
                v-model.number="selectedCafoID"
                class="form-select"
                :disabled="saving"
                @change="selectCurrentCafo"
              >
                <option v-for="cafo in cafos" :key="cafo.id" :value="cafo.id">
                  #{{ cafo.id }} · {{ cafo.name }}{{ cafo.chinese ? ` · ${cafo.chinese}` : "" }}
                </option>
              </select>
            </label>

            <div class="cafo-edit-grid">
              <label class="cafo-edit-field">
                <span>Name</span>
                <input
                  ref="firstInput"
                  v-model="form.name"
                  class="form-control"
                  name="name"
                  type="text"
                  required
                  :disabled="saving"
                />
              </label>
              <label class="cafo-edit-field">
                <span>Chinese</span>
                <input v-model="form.chinese" class="form-control" name="chinese" type="text" :disabled="saving" />
              </label>
              <label class="cafo-edit-field is-wide">
                <span>Alias</span>
                <input v-model="form.alias" class="form-control" name="alias" type="text" :disabled="saving" />
              </label>
              <label class="cafo-edit-field">
                <span>Birth Day</span>
                <input v-model="form.birthDay" class="form-control" name="birthDay" type="date" :disabled="saving" />
              </label>
              <label class="cafo-edit-field">
                <span>Height</span>
                <input v-model="form.height" class="form-control" name="height" type="number" :disabled="saving" />
              </label>
              <label class="cafo-edit-field">
                <span>Cup</span>
                <input v-model="form.cup" class="form-control" name="cup" type="text" :disabled="saving" />
              </label>
              <label class="cafo-edit-field">
                <span>BWH</span>
                <input v-model="form.bwh" class="form-control" name="bwh" type="text" :disabled="saving" />
              </label>
              <label class="cafo-edit-field is-wide">
                <span>Avatar</span>
                <input v-model="form.avatar" class="form-control" name="avatar" type="text" :disabled="saving" />
              </label>
              <label class="cafo-edit-field is-wide">
                <span>人物简介</span>
                <textarea
                  v-model="form.biography"
                  class="form-control"
                  name="biography"
                  rows="5"
                  :disabled="saving"
                ></textarea>
              </label>
            </div>
          </template>
        </div>

        <div class="modal-footer-v2">
          <button type="button" class="button-secondary" :disabled="saving" @click="requestClose">取消</button>
          <button type="submit" class="button-primary" :disabled="loading || saving || cafos.length === 0">
            {{ saving ? "保存中..." : "保存" }}
          </button>
        </div>
      </form>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { nextTick, reactive, ref, watch } from "vue";
import type { CafoProfile, CafoWriteInput } from "@/features/movies/api";

const props = defineProps<{
  active: boolean;
  loading: boolean;
  saving: boolean;
  personName: string;
  cafos: CafoProfile[];
  errorMessage: string;
  successMessage: string;
}>();

const emit = defineEmits<{
  close: [];
  save: [cafoID: number, input: CafoWriteInput];
}>();

const dialog = ref<HTMLFormElement | null>(null);
const firstInput = ref<HTMLInputElement | null>(null);
const selectedCafoID = ref(0);
const form = reactive({
  name: "",
  alias: "",
  chinese: "",
  birthDay: "",
  height: "",
  cup: "",
  bwh: "",
  avatar: "",
  biography: ""
});

watch(
  () => props.active,
  (active) => {
    if (!active) return;
    applyCafo(props.cafos[0]);
    void nextTick(() => (firstInput.value || dialog.value)?.focus());
  }
);

watch(
  () => props.cafos,
  (cafos) => {
    if (!props.active) return;
    applyCafo(cafos.find((cafo) => cafo.id === selectedCafoID.value) || cafos[0]);
    void nextTick(() => (firstInput.value || dialog.value)?.focus());
  }
);

function selectCurrentCafo() {
  applyCafo(props.cafos.find((cafo) => cafo.id === selectedCafoID.value));
}

function applyCafo(cafo?: CafoProfile) {
  selectedCafoID.value = cafo?.id || 0;
  form.name = cafo?.name || "";
  form.alias = cafo?.alias || "";
  form.chinese = cafo?.chinese || "";
  form.birthDay = unixToDateInput(cafo?.birthDay || 0);
  form.height = cafo?.height ? String(cafo.height) : "";
  form.cup = cafo?.cup || "";
  form.bwh = cafo?.bwh || "";
  form.avatar = cafo?.avatar || "";
  form.biography = cafo?.biography || "";
}

function requestClose() {
  if (!props.saving) emit("close");
}

function submit() {
  if (selectedCafoID.value <= 0) return;
  emit("save", selectedCafoID.value, {
    name: form.name,
    alias: form.alias,
    chinese: form.chinese,
    birthDay: dateInputToUnix(form.birthDay),
    height: Number(form.height || 0),
    cup: form.cup,
    bwh: form.bwh,
    avatar: form.avatar,
    biography: form.biography
  });
}

function unixToDateInput(value: number): string {
  if (!Number.isFinite(value) || value === 0) return "";
  const date = new Date(value * 1000);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function dateInputToUnix(value: string): number {
  const parts = value.split("-").map(Number);
  if (parts.length !== 3 || parts.some((part) => !Number.isFinite(part))) return 0;
  return Math.floor(new Date(parts[0], parts[1] - 1, parts[2]).getTime() / 1000);
}
</script>

<style scoped>
.cafo-edit-overlay {
  padding: 20px;
}

.cafo-edit-panel {
  width: min(720px, 100%);
  max-height: calc(100vh - 40px);
}

.modal-header-v2 > div {
  min-width: 0;
}

.modal-header-v2 p {
  overflow: hidden;
  margin: 3px 0 0;
  color: var(--muted);
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cafo-edit-body {
  display: grid;
  gap: 12px;
  overflow-y: auto;
}

.cafo-edit-state {
  border: 1px dashed var(--line);
  border-radius: 8px;
  padding: 24px;
  color: var(--muted);
  text-align: center;
}

.cafo-edit-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.cafo-edit-field {
  display: grid;
  min-width: 0;
  gap: 5px;
}

.cafo-edit-field.is-wide {
  grid-column: 1 / -1;
}

.cafo-edit-field > span {
  color: var(--muted);
  font-size: 11px;
  font-weight: 850;
}

@media (max-width: 640px) {
  .cafo-edit-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .cafo-edit-field.is-wide {
    grid-column: auto;
  }
}

@media (prefers-reduced-motion: reduce) {
  .cafo-edit-panel,
  .cafo-edit-panel * {
    transition: none !important;
  }
}
</style>
