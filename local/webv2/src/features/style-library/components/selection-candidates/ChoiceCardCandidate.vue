<template>
  <article class="choice-card-candidate">
    <header class="choice-card-head">
      <div class="choice-card-title">
        <span>05</span>
        <div>
          <h3>说明型选择卡</h3>
          <p>用标题、说明和状态徽标承载需要理解后再选择的配置。</p>
        </div>
      </div>
      <output>{{ chosenForum }} · {{ chosenPolicy }}</output>
    </header>

    <fieldset class="choice-card-fieldset">
      <legend>目标版块</legend>
      <div class="choice-card-forum-grid">
        <template v-for="option in forumOptions" :key="option.value">
          <input
            :id="`choice-card-forum-${option.value}`"
            v-model="forum"
            type="radio"
            name="choice-card-forum"
            :value="option.value"
            :disabled="option.disabled"
          />
          <label :for="`choice-card-forum-${option.value}`">
            <span class="choice-card-code">{{ option.code }}</span>
            <span class="choice-card-copy">
              <strong>{{ option.label }}</strong>
              <small>{{ option.description }}</small>
            </span>
            <span class="choice-card-badge">{{ forum === option.value ? "已选" : "选择" }}</span>
          </label>
        </template>
      </div>
    </fieldset>

    <fieldset class="choice-card-fieldset">
      <legend>已有帖子处理</legend>
      <div class="choice-card-policy-grid">
        <template v-for="option in policyOptions" :key="option.value">
          <input
            :id="`choice-card-policy-${option.value}`"
            v-model="policy"
            type="radio"
            name="choice-card-policy"
            :value="option.value"
          />
          <label :for="`choice-card-policy-${option.value}`">
            <span class="choice-card-copy">
              <strong>{{ option.label }}</strong>
              <small>{{ option.description }}</small>
            </span>
            <span class="choice-card-corner" aria-hidden="true">✓</span>
          </label>
        </template>
      </div>
    </fieldset>
  </article>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";

const forumOptions = [
  { value: "all", code: "ALL", label: "全部", description: "覆盖当前任务允许的全部版块。" },
  { value: "asia-original", code: "AO", label: "亚洲视频原创", description: "只处理原创内容版块。" },
  { value: "chinese-subtitle", code: "CS", label: "高清中文字幕", description: "只处理中文字幕版块。" },
  { value: "other", code: "ETC", label: "其他版块", description: "当前任务暂不可使用。", disabled: true }
] as const;

const policyOptions = [
  { value: "upsert", label: "更新已有", description: "刷新基础信息，并重新进入详情抓取。" },
  { value: "skip", label: "跳过已有", description: "保留现有记录，只处理新发现的帖子。" }
] as const;

const forum = ref("asia-original");
const policy = ref("upsert");
const chosenForum = computed(() => forumOptions.find((option) => option.value === forum.value)?.label || "未选择版块");
const chosenPolicy = computed(
  () => policyOptions.find((option) => option.value === policy.value)?.label || "未选择策略"
);
</script>

<style scoped>
.choice-card-candidate {
  display: grid;
  gap: 16px;
  min-width: 0;
  padding: 16px;
  border: 1px solid var(--line);
  border-radius: var(--card-radius);
  background: var(--panel-soft);
}

.choice-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.choice-card-title {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
}

.choice-card-title > span {
  display: grid;
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  place-items: center;
  border: 1px solid var(--text);
  border-radius: 4px;
  background: var(--text);
  color: var(--panel);
  font-size: 11px;
  font-weight: 750;
}

.choice-card-title h3,
.choice-card-title p {
  margin: 0;
}

.choice-card-title h3 {
  color: var(--text);
  font-size: 14px;
  line-height: 20px;
}

.choice-card-title p {
  margin-top: 2px;
  color: var(--muted);
  font-size: 10px;
  line-height: 16px;
}

.choice-card-head output {
  flex: 0 0 auto;
  border: 1px solid var(--primary-border);
  border-radius: 4px;
  padding: 4px 8px;
  background: var(--primary-soft);
  color: var(--primary-hover);
  font-size: 10px;
  font-weight: 650;
  line-height: 16px;
}

.choice-card-fieldset {
  min-width: 0;
  margin: 0;
  border: 0;
  padding: 0;
}

.choice-card-fieldset legend {
  margin-bottom: 7px;
  color: var(--muted);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.05em;
  line-height: 16px;
}

.choice-card-fieldset input {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  opacity: 0;
}

.choice-card-forum-grid,
.choice-card-policy-grid {
  display: grid;
  gap: 8px;
}

.choice-card-forum-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.choice-card-policy-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.choice-card-fieldset label {
  position: relative;
  display: flex;
  min-width: 0;
  min-height: 58px;
  align-items: center;
  gap: 9px;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 6px;
  padding: 8px 9px;
  background: var(--panel);
  cursor: pointer;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease,
    transform 150ms ease;
}

.choice-card-code {
  display: grid;
  width: 32px;
  height: 32px;
  flex: 0 0 32px;
  place-items: center;
  border-radius: 4px;
  background: var(--soft);
  color: var(--muted);
  font-size: 8px;
  font-weight: 750;
  letter-spacing: 0.04em;
}

.choice-card-copy {
  display: grid;
  min-width: 0;
  gap: 2px;
}

.choice-card-copy strong {
  color: var(--text);
  font-size: 11px;
  font-weight: 650;
  line-height: 16px;
}

.choice-card-copy small {
  overflow: hidden;
  color: var(--muted);
  font-size: 9px;
  line-height: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.choice-card-badge {
  margin-left: auto;
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 1px 6px;
  color: var(--placeholder);
  font-size: 8px;
  font-weight: 700;
  line-height: 14px;
}

.choice-card-corner {
  position: absolute;
  top: 0;
  right: 0;
  display: grid;
  width: 22px;
  height: 22px;
  place-items: center;
  border-radius: 0 0 0 8px;
  background: var(--line);
  color: transparent;
  font-size: 10px;
}

.choice-card-fieldset label:hover {
  border-color: var(--primary-border);
  transform: translateY(-1px);
}

.choice-card-fieldset input:checked + label {
  border-color: var(--primary);
  box-shadow: inset 3px 0 0 var(--primary);
}

.choice-card-fieldset input:checked + label .choice-card-code {
  background: var(--primary-soft);
  color: var(--primary-hover);
}

.choice-card-fieldset input:checked + label .choice-card-badge {
  border-color: var(--primary-border);
  background: var(--primary-soft);
  color: var(--primary-hover);
}

.choice-card-fieldset input:checked + label .choice-card-corner {
  background: var(--primary);
  color: var(--on-primary);
}

.choice-card-fieldset input:focus-visible + label {
  outline: 2px solid var(--primary);
  outline-offset: 2px;
}

.choice-card-fieldset input:disabled + label {
  cursor: not-allowed;
  opacity: 0.42;
}

@media (prefers-reduced-motion: reduce) {
  .choice-card-fieldset label {
    transition: none;
  }
}
</style>
