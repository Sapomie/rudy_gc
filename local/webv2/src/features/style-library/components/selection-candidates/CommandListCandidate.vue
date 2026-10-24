<template>
  <article class="command-list-candidate">
    <header class="command-list-head">
      <div>
        <small>CHOICE SYSTEM 04</small>
        <h3>纵向命令清单</h3>
      </div>
      <p>逐行扫描，右侧状态直接说明当前选择。</p>
    </header>

    <div class="command-list-columns">
      <section aria-labelledby="command-forum-title">
        <h4 id="command-forum-title">目标版块</h4>
        <ol class="command-list-options">
          <li v-for="(option, index) in forumOptions" :key="option.value">
            <button
              type="button"
              :class="{ current: forum === option.value }"
              role="radio"
              :aria-checked="forum === option.value"
              :disabled="option.disabled"
              @click="forum = option.value"
            >
              <span class="command-list-index">{{ String(index + 1).padStart(2, "0") }}</span>
              <span class="command-list-copy">
                <strong>{{ option.label }}</strong>
                <small>{{ option.note }}</small>
              </span>
              <span class="command-list-state">{{ forum === option.value ? "当前" : "选择" }}</span>
            </button>
          </li>
        </ol>
      </section>

      <section aria-labelledby="command-policy-title">
        <h4 id="command-policy-title">已有帖子处理</h4>
        <ol class="command-list-options command-list-policy">
          <li v-for="(option, index) in policyOptions" :key="option.value">
            <button
              type="button"
              :class="{ current: policy === option.value }"
              role="radio"
              :aria-checked="policy === option.value"
              @click="policy = option.value"
            >
              <span class="command-list-index">P{{ index + 1 }}</span>
              <span class="command-list-copy">
                <strong>{{ option.label }}</strong>
                <small>{{ option.note }}</small>
              </span>
              <span class="command-list-state">{{ policy === option.value ? "当前" : "选择" }}</span>
            </button>
          </li>
        </ol>
        <div class="command-list-output">
          <span>配置结果</span>
          <strong>{{ resultText }}</strong>
        </div>
      </section>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";

const forumOptions = [
  { value: "all", label: "全部", note: "覆盖全部目标版块" },
  { value: "asia-original", label: "亚洲视频原创", note: "仅抓取原创内容" },
  { value: "chinese-subtitle", label: "高清中文字幕", note: "仅抓取中文字幕内容" },
  { value: "other", label: "其他版块", note: "当前不可选择", disabled: true }
] as const;

const policyOptions = [
  { value: "upsert", label: "更新已有", note: "覆盖基础信息并重新抓取详情" },
  { value: "skip", label: "跳过已有", note: "保留现有数据，只处理新帖子" }
] as const;

const forum = ref("all");
const policy = ref("skip");
const resultText = computed(() => {
  const forumLabel = forumOptions.find((option) => option.value === forum.value)?.label || "未选择版块";
  const policyLabel = policyOptions.find((option) => option.value === policy.value)?.label || "未选择策略";
  return `${forumLabel} / ${policyLabel}`;
});
</script>

<style scoped>
.command-list-candidate {
  display: grid;
  gap: 15px;
  min-width: 0;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--text) 16%, var(--line));
  border-radius: 4px;
  background: var(--panel);
}

.command-list-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
  padding: 13px 15px;
  background: var(--text);
  color: var(--panel);
}

.command-list-head small {
  color: color-mix(in srgb, var(--panel) 68%, transparent);
  font-size: 8px;
  font-weight: 750;
  letter-spacing: 0.14em;
  line-height: 12px;
}

.command-list-head h3,
.command-list-head p {
  margin: 0;
}

.command-list-head h3 {
  margin-top: 2px;
  font-size: 14px;
  line-height: 20px;
}

.command-list-head p {
  color: color-mix(in srgb, var(--panel) 72%, transparent);
  font-size: 10px;
  line-height: 16px;
  text-align: right;
}

.command-list-columns {
  display: grid;
  grid-template-columns: 1.15fr 0.85fr;
  gap: 18px;
  padding: 0 15px 15px;
}

.command-list-columns section {
  min-width: 0;
}

.command-list-columns h4 {
  margin: 0 0 6px;
  color: var(--muted);
  font-size: 9px;
  font-weight: 750;
  letter-spacing: 0.08em;
  line-height: 16px;
  text-transform: uppercase;
}

.command-list-options {
  display: grid;
  margin: 0;
  border-top: 1px solid var(--line);
  padding: 0;
  list-style: none;
}

.command-list-options li {
  min-width: 0;
  border-bottom: 1px solid var(--line);
}

.command-list-options button {
  position: relative;
  display: grid;
  width: 100%;
  min-width: 0;
  grid-template-columns: 28px minmax(0, 1fr) 36px;
  align-items: center;
  gap: 7px;
  margin: 0;
  border: 0;
  padding: 7px 4px;
  background: transparent;
  color: var(--muted);
  text-align: left;
  cursor: pointer;
}

.command-list-options button::before {
  content: "";
  position: absolute;
  inset: 5px auto 5px -4px;
  width: 2px;
  border-radius: 2px;
  background: transparent;
}

.command-list-index {
  color: var(--placeholder);
  font-size: 9px;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  line-height: 16px;
}

.command-list-copy {
  display: grid;
  min-width: 0;
}

.command-list-copy strong {
  overflow: hidden;
  color: var(--text);
  font-size: 11px;
  font-weight: 650;
  line-height: 16px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.command-list-copy small {
  overflow: hidden;
  color: var(--muted);
  font-size: 9px;
  line-height: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.command-list-state {
  justify-self: end;
  color: var(--placeholder);
  font-size: 9px;
  font-weight: 650;
  line-height: 16px;
}

.command-list-options button:hover:not(:disabled):not(.current) {
  background: var(--panel-soft);
}

.command-list-options button.current {
  background: var(--primary-soft);
}

.command-list-options button.current::before {
  background: var(--primary);
}

.command-list-options button.current .command-list-index,
.command-list-options button.current .command-list-state {
  color: var(--primary-hover);
}

.command-list-options button:focus-visible {
  z-index: 1;
  outline: 2px solid var(--primary);
  outline-offset: -2px;
}

.command-list-options button:disabled {
  cursor: not-allowed;
  opacity: 0.42;
}

.command-list-output {
  display: grid;
  gap: 2px;
  margin-top: 12px;
  border-left: 2px solid var(--primary);
  padding: 6px 9px;
  background: var(--panel-soft);
}

.command-list-output span {
  color: var(--muted);
  font-size: 9px;
  line-height: 14px;
}

.command-list-output strong {
  color: var(--text);
  font-size: 11px;
  line-height: 17px;
}
</style>
