<template>
  <article class="pill-choice-candidate">
    <header class="pill-choice-head">
      <div class="pill-choice-number">03</div>
      <div>
        <h3>标签胶囊组</h3>
        <p>选项按内容自然换行，以勾选标记强化当前值。</p>
      </div>
    </header>

    <div class="pill-choice-field">
      <span class="pill-choice-label">目标版块</span>
      <div class="pill-choice-options" role="radiogroup" aria-label="标签胶囊组目标版块">
        <button
          v-for="option in forumOptions"
          :key="option.value"
          type="button"
          :class="{ selected: forum === option.value }"
          role="radio"
          :aria-checked="forum === option.value"
          :disabled="option.disabled"
          @click="forum = option.value"
        >
          <span class="pill-choice-check" aria-hidden="true">{{ forum === option.value ? "✓" : "" }}</span>
          {{ option.label }}
        </button>
      </div>
    </div>

    <div class="pill-choice-field pill-choice-policy">
      <span class="pill-choice-label">已有帖子处理</span>
      <div class="pill-choice-options" role="radiogroup" aria-label="标签胶囊组已有帖子处理">
        <button
          v-for="option in policyOptions"
          :key="option.value"
          type="button"
          :class="{ selected: policy === option.value }"
          role="radio"
          :aria-checked="policy === option.value"
          @click="policy = option.value"
        >
          <span class="pill-choice-check" aria-hidden="true">{{ policy === option.value ? "✓" : "" }}</span>
          {{ option.label }}
        </button>
      </div>
    </div>

    <aside>
      <span>已选择</span>
      <b>{{ currentForum }}</b>
      <i aria-hidden="true">→</i>
      <b>{{ currentPolicy }}</b>
    </aside>
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

const forum = ref("chinese-subtitle");
const policy = ref("skip");
const currentForum = computed(() => forumOptions.find((option) => option.value === forum.value)?.label || "未选择");
const currentPolicy = computed(() => policyOptions.find((option) => option.value === policy.value)?.label || "未选择");
</script>

<style scoped>
.pill-choice-candidate {
  position: relative;
  display: grid;
  align-content: start;
  gap: 16px;
  min-width: 0;
  padding: 17px 16px 16px;
  border: 1px solid var(--line);
  border-radius: 14px;
  background: var(--panel);
}

.pill-choice-candidate::before {
  content: "";
  position: absolute;
  inset: 0 16px auto;
  height: 3px;
  border-radius: 0 0 3px 3px;
  background: linear-gradient(90deg, var(--primary), var(--primary-border));
}

.pill-choice-head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.pill-choice-number {
  display: grid;
  width: 32px;
  height: 32px;
  flex: 0 0 32px;
  place-items: center;
  border-radius: 50%;
  background: var(--primary-soft);
  color: var(--primary-hover);
  font-size: 11px;
  font-weight: 750;
}

.pill-choice-head h3,
.pill-choice-head p {
  margin: 0;
}

.pill-choice-head h3 {
  color: var(--text);
  font-size: 14px;
  line-height: 20px;
}

.pill-choice-head p {
  margin-top: 1px;
  color: var(--muted);
  font-size: 10px;
  line-height: 16px;
}

.pill-choice-field {
  display: grid;
  grid-template-columns: 100px minmax(0, 1fr);
  align-items: start;
  gap: 10px;
}

.pill-choice-label {
  padding-top: 5px;
  color: var(--muted);
  font-size: 11px;
  font-weight: 650;
  line-height: 18px;
}

.pill-choice-options {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  gap: 7px;
}

.pill-choice-options button {
  display: inline-flex;
  min-height: 30px;
  align-items: center;
  gap: 6px;
  margin: 0;
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 4px 11px 4px 7px;
  background: var(--panel-soft);
  color: var(--muted);
  font-size: 11px;
  font-weight: 600;
  line-height: 20px;
  cursor: pointer;
  transition:
    border-color 150ms ease,
    background-color 150ms ease,
    color 150ms ease,
    transform 150ms ease;
}

.pill-choice-check {
  display: grid;
  width: 16px;
  height: 16px;
  flex: 0 0 16px;
  place-items: center;
  border: 1px solid var(--line-strong);
  border-radius: 50%;
  background: var(--panel);
  color: transparent;
  font-size: 10px;
  line-height: 1;
}

.pill-choice-options button:hover:not(:disabled):not(.selected) {
  border-color: var(--primary-border);
  color: var(--primary-hover);
  transform: translateY(-1px);
}

.pill-choice-options button.selected {
  border-color: var(--primary);
  background: var(--primary-soft);
  color: var(--primary-hover);
}

.pill-choice-options button.selected .pill-choice-check {
  border-color: var(--primary);
  background: var(--primary);
  color: var(--on-primary);
}

.pill-choice-options button:active:not(:disabled) {
  transform: translateY(1px);
}

.pill-choice-options button:focus-visible {
  outline: 2px solid var(--primary);
  outline-offset: 2px;
}

.pill-choice-options button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.pill-choice-candidate aside {
  display: flex;
  min-height: 32px;
  align-items: center;
  gap: 7px;
  border-radius: 6px;
  padding: 6px 10px;
  background: var(--soft);
  color: var(--muted);
  font-size: 10px;
  line-height: 18px;
}

.pill-choice-candidate aside b {
  color: var(--text);
  font-size: 11px;
}

.pill-choice-candidate aside i {
  color: var(--primary);
  font-style: normal;
}

@media (prefers-reduced-motion: reduce) {
  .pill-choice-options button {
    transition: none;
  }
}
</style>
