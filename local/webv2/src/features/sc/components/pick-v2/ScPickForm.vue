<template>
  <div class="card next02-sc-panel next02-sc-movie-table-card smart-pick-config-panel">
    <div class="card-body">
      <form class="needs-validation" novalidate @submit.prevent="submit">
        <div class="smart-pick-strategy-row">
          <div class="card border smart-pick-global-panel">
            <div class="card-body">
              <div class="smart-pick-section-title">
                <span>01</span>
                <h6>混合抽取数量</h6>
              </div>
              <div class="row g-3 align-items-end smart-pick-global-grid">
                <div class="col-md-2 smart-pick-field">
                  <label class="form-label" for="scpickv2JavPickN">JAV 数量</label>
                  <input
                    id="scpickv2JavPickN"
                    v-model.number="javPickN"
                    type="number"
                    min="0"
                    step="1"
                    class="form-control"
                  />
                </div>
                <div class="col-md-2 smart-pick-field">
                  <label class="form-label" for="scpickv2Fc2PickN">FC2 数量</label>
                  <input
                    id="scpickv2Fc2PickN"
                    v-model.number="fc2PickN"
                    type="number"
                    min="0"
                    step="1"
                    class="form-control"
                  />
                </div>
                <div class="col-md-2 d-flex align-items-end smart-pick-advanced-toggle-wrap">
                  <button
                    type="button"
                    class="btn btn-sm btn-outline-secondary w-100 smart-pick-advanced-toggle"
                    aria-controls="scpickv2GlobalAdvancedPanel"
                    :aria-expanded="globalAdvancedExpanded ? 'true' : 'false'"
                    @click="globalAdvancedExpanded = !globalAdvancedExpanded"
                  >
                    高级参数
                  </button>
                </div>
                <div
                  id="scpickv2GlobalAdvancedPanel"
                  class="col-12 collapse smart-pick-advanced-panel"
                  :class="{ show: globalAdvancedExpanded }"
                >
                  <div class="smart-pick-mixed-options-grid">
                    <ScPickStrategyFields kind="jav" :model-value="javOptions" @update:model-value="applyJAVOptions" />
                    <ScPickStrategyFields kind="fc2" :model-value="fc2Options" @update:model-value="applyFC2Options" />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <section class="smart-pick-type-groups is-jav" aria-labelledby="scpickv2JavGroups">
            <div class="smart-pick-type-groups-head">
              <div>
                <span>JAV</span>
                <h6 id="scpickv2JavGroups">JAV 条件</h6>
              </div>
              <button type="button" class="btn btn-outline-secondary" :disabled="loading" @click="addGroup('jav')">
                ＋ 添加 JAV 条件
              </button>
            </div>
            <div>
              <ScPickGroupEditor
                v-for="(group, index) in javGroups"
                :key="group.id"
                kind="jav"
                :index="index"
                :model-value="group"
                :can-remove="javGroups.length > 1"
                @update:model-value="updateGroup('jav', $event)"
                @remove="removeGroup('jav', group.id)"
              />
            </div>
          </section>

          <section class="smart-pick-type-groups is-fc2" aria-labelledby="scpickv2Fc2Groups">
            <div class="smart-pick-type-groups-head">
              <div>
                <span>FC2</span>
                <h6 id="scpickv2Fc2Groups">FC2 条件</h6>
              </div>
              <button type="button" class="btn btn-outline-secondary" :disabled="loading" @click="addGroup('fc2')">
                ＋ 添加 FC2 条件
              </button>
            </div>
            <div>
              <ScPickGroupEditor
                v-for="(group, index) in fc2Groups"
                :key="group.id"
                kind="fc2"
                :index="index"
                :model-value="group"
                :can-remove="fc2Groups.length > 1"
                @update:model-value="updateGroup('fc2', $event)"
                @remove="removeGroup('fc2', group.id)"
              />
            </div>
          </section>
        </div>

        <div class="smart-pick-actions">
          <button type="submit" class="btn btn-primary" :disabled="loading">
            {{ loading ? "正在抽取..." : "执行抽取预览" }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import ScPickGroupEditor from "@/features/sc/components/pick-v2/ScPickGroupEditor.vue";
import ScPickStrategyFields from "@/features/sc/components/pick-v2/ScPickStrategyFields.vue";
import { useScPickV2Form } from "@/features/sc/composables/useScPickV2Form";
import type { ScPickV2FC2Options, ScPickV2JAVOptions, ScPickV2Request } from "@/features/sc/types";

defineProps<{ loading: boolean }>();
const emit = defineEmits<{ submit: [request: ScPickV2Request]; invalid: [message: string] }>();
const globalAdvancedExpanded = ref(false);
const {
  javPickN,
  fc2PickN,
  javOptions,
  fc2Options,
  javGroups,
  fc2Groups,
  addGroup,
  updateGroup,
  removeGroup,
  buildRequest
} = useScPickV2Form();

function applyJAVOptions(value: ScPickV2JAVOptions & { randomSeed?: number | null }) {
  Object.assign(javOptions, value);
}

function applyFC2Options(value: ScPickV2FC2Options & { randomSeed?: number | null }) {
  Object.assign(fc2Options, value);
}

function submit() {
  try {
    emit("submit", buildRequest());
  } catch (reason) {
    emit("invalid", reason instanceof Error ? reason.message : String(reason));
  }
}
</script>
