<template>
  <article class="segmented-choice-candidate">
    <header class="segmented-choice-head">
      <div>
        <span>02 · SEGMENTED</span>
        <h3>连体分段器</h3>
      </div>
      <p>整组共用一条轨道，当前值以实色分段显示。</p>
    </header>

    <section class="segmented-choice-field" aria-labelledby="segmented-forum-label">
      <h4 id="segmented-forum-label">目标版块</h4>
      <div
        class="segmented-choice-track segmented-choice-track-forum"
        role="radiogroup"
        aria-label="连体分段器目标版块"
      >
        <button
          v-for="option in forumOptions"
          :key="option.value"
          type="button"
          :class="{ active: selectedForum === option.value }"
          role="radio"
          :aria-checked="selectedForum === option.value"
          :disabled="option.disabled"
          @click="selectedForum = option.value"
        >
          {{ option.label }}
        </button>
      </div>
    </section>

    <section class="segmented-choice-field" aria-labelledby="segmented-policy-label">
      <h4 id="segmented-policy-label">已有帖子处理</h4>
      <div
        class="segmented-choice-track segmented-choice-track-policy"
        role="radiogroup"
        aria-label="连体分段器已有帖子处理"
      >
        <button
          v-for="option in policyOptions"
          :key="option.value"
          type="button"
          :class="{ active: selectedPolicy === option.value }"
          role="radio"
          :aria-checked="selectedPolicy === option.value"
          @click="selectedPolicy = option.value"
        >
          {{ option.label }}
        </button>
      </div>
    </section>

    <div class="segmented-choice-summary">
      <span>当前配置</span>
      <strong>{{ selectedForumLabel }} / {{ selectedPolicyLabel }}</strong>
    </div>
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

const selectedForum = ref("asia-original");
const selectedPolicy = ref("upsert");
const selectedForumLabel = computed(
  () => forumOptions.find((option) => option.value === selectedForum.value)?.label || "未选择版块"
);
const selectedPolicyLabel = computed(
  () => policyOptions.find((option) => option.value === selectedPolicy.value)?.label || "未选择处理方式"
);
</script>

<style scoped>
.segmented-choice-candidate {
  display: grid;
  align-content: start;
  gap: 14px;
  min-width: 0;
  padding: 16px;
  border: 1px solid var(--primary-border);
  border-radius: 10px;
  background: linear-gradient(145deg, var(--panel) 0%, var(--primary-soft) 160%);
  box-shadow: inset 0 1px 0 var(--panel);
}

.segmented-choice-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
}

.segmented-choice-head span {
  display: block;
  color: var(--primary);
  font-size: 9px;
  font-weight: 750;
  letter-spacing: 0.1em;
  line-height: 14px;
}

.segmented-choice-head h3,
.segmented-choice-head p {
  margin: 0;
}

.segmented-choice-head h3 {
  margin-top: 2px;
  color: var(--text);
  font-size: 15px;
  line-height: 20px;
}

.segmented-choice-head p {
  max-width: 230px;
  color: var(--muted);
  font-size: 10px;
  line-height: 16px;
  text-align: right;
}

.segmented-choice-field {
  display: grid;
  gap: 5px;
}

.segmented-choice-field h4 {
  margin: 0;
  color: var(--muted);
  font-size: 10px;
  font-weight: 650;
  line-height: 16px;
}

.segmented-choice-track {
  display: grid;
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 6px;
  padding: 2px;
  background: var(--soft);
}

.segmented-choice-track-forum {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.segmented-choice-track-policy {
  width: min(100%, 330px);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.segmented-choice-track button {
  min-width: 0;
  min-height: 30px;
  margin: 0;
  border: 0;
  border-radius: 4px;
  padding: 5px 8px;
  background: transparent;
  color: var(--muted);
  font-size: 11px;
  font-weight: 600;
  line-height: 18px;
  white-space: nowrap;
  cursor: pointer;
  transition:
    background-color 150ms ease,
    color 150ms ease,
    box-shadow 150ms ease;
}

.segmented-choice-track button:hover:not(:disabled):not(.active) {
  color: var(--primary-hover);
}

.segmented-choice-track button.active {
  background: var(--primary);
  color: var(--on-primary);
  box-shadow: 0 2px 6px color-mix(in srgb, var(--primary) 24%, transparent);
}

.segmented-choice-track button:active:not(:disabled) {
  background: var(--primary-hover);
}

.segmented-choice-track button:focus-visible {
  position: relative;
  z-index: 1;
  outline: 2px solid var(--primary);
  outline-offset: -2px;
}

.segmented-choice-track button:disabled {
  color: var(--placeholder);
  cursor: not-allowed;
  opacity: 0.45;
}

.segmented-choice-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding-top: 10px;
  border-top: 1px dashed var(--primary-border);
  color: var(--muted);
  font-size: 10px;
  line-height: 16px;
}

.segmented-choice-summary strong {
  color: var(--primary-hover);
  font-size: 11px;
}

@media (prefers-reduced-motion: reduce) {
  .segmented-choice-track button {
    transition: none;
  }
}
</style>
