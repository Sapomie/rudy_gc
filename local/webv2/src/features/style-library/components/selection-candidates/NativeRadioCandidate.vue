<template>
  <article class="native-radio-candidate">
    <header class="native-radio-head">
      <span>01</span>
      <div>
        <h3>原生单选行</h3>
        <p>保留标准单选语义，以圆环和实心圆点表达状态。</p>
      </div>
    </header>

    <fieldset class="native-radio-fieldset">
      <legend>目标版块</legend>
      <div class="native-radio-options">
        <label v-for="option in forumOptions" :key="option.value" :class="{ disabled: option.disabled }">
          <input
            v-model="selectedForum"
            type="radio"
            name="native-radio-forum"
            :value="option.value"
            :disabled="option.disabled"
          />
          <span>{{ option.label }}</span>
        </label>
      </div>
    </fieldset>

    <fieldset class="native-radio-fieldset">
      <legend>已有帖子处理</legend>
      <div class="native-radio-options">
        <label v-for="option in policyOptions" :key="option.value">
          <input v-model="selectedPolicy" type="radio" name="native-radio-policy" :value="option.value" />
          <span>{{ option.label }}</span>
        </label>
      </div>
    </fieldset>

    <footer>{{ selectedForumLabel }} · {{ selectedPolicyLabel }}</footer>
  </article>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";

const forumOptions = [
  { value: "all", label: "全部" },
  { value: "asia-original", label: "亚洲视频原创" },
  { value: "chinese-subtitle", label: "高清中文字幕" },
  { value: "other", label: "其他版块", disabled: true }
] as const;

const policyOptions = [
  { value: "upsert", label: "更新已有" },
  { value: "skip", label: "跳过已有" }
] as const;

const selectedForum = ref("all");
const selectedPolicy = ref("upsert");
const selectedForumLabel = computed(
  () => forumOptions.find((option) => option.value === selectedForum.value)?.label || "未选择版块"
);
const selectedPolicyLabel = computed(
  () => policyOptions.find((option) => option.value === selectedPolicy.value)?.label || "未选择处理方式"
);
</script>

<style scoped>
.native-radio-candidate {
  display: grid;
  align-content: start;
  gap: 15px;
  min-width: 0;
  padding: 16px;
  border: 1px solid var(--line);
  border-radius: var(--card-radius);
  background: var(--panel);
}

.native-radio-head {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.native-radio-head > span {
  color: var(--primary);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  line-height: 20px;
}

.native-radio-head h3,
.native-radio-head p {
  margin: 0;
}

.native-radio-head h3 {
  color: var(--text);
  font-size: 14px;
  line-height: 20px;
}

.native-radio-head p {
  margin-top: 2px;
  color: var(--muted);
  font-size: 11px;
  line-height: 17px;
}

.native-radio-fieldset {
  min-width: 0;
  margin: 0;
  border: 0;
  padding: 0;
}

.native-radio-fieldset legend {
  margin-bottom: 7px;
  color: var(--muted);
  font-size: 11px;
  font-weight: 600;
  line-height: 16px;
}

.native-radio-options {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 18px;
}

.native-radio-options label {
  display: inline-flex;
  min-height: 24px;
  align-items: center;
  gap: 7px;
  color: var(--text);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  cursor: pointer;
}

.native-radio-options input {
  appearance: none;
  box-sizing: border-box;
  width: 15px;
  height: 15px;
  flex: 0 0 15px;
  margin: 0;
  border: 1px solid var(--line-strong);
  border-radius: 50%;
  background: var(--panel);
  cursor: pointer;
  transition:
    border-color 140ms ease,
    box-shadow 140ms ease;
}

.native-radio-options input:checked {
  border: 4px solid var(--primary);
}

.native-radio-options label:hover:not(.disabled) input {
  border-color: var(--primary-hover);
}

.native-radio-options input:focus-visible {
  outline: 2px solid var(--primary);
  outline-offset: 2px;
}

.native-radio-options label:has(input:checked) span {
  color: var(--primary-hover);
  font-weight: 650;
}

.native-radio-options label.disabled {
  color: var(--placeholder);
  cursor: not-allowed;
  opacity: 0.55;
}

.native-radio-candidate footer {
  padding-top: 10px;
  border-top: 1px solid var(--soft);
  color: var(--muted);
  font-size: 11px;
  line-height: 16px;
}

@media (prefers-reduced-motion: reduce) {
  .native-radio-options input {
    transition: none;
  }
}
</style>
