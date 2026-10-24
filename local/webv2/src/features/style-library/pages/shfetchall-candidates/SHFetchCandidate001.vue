<template>
  <main class="candidate-001-page" data-review-candidate="001">
    <section class="candidate-001-review-head" aria-labelledby="candidate-001-review-title">
      <div>
        <span class="candidate-001-review-number">方案 001</span>
        <h1 id="candidate-001-review-title">参数清单</h1>
      </div>
      <p>固定标签轴 · 三行配置 · 独立任务控制</p>
    </section>

    <section class="candidate-001-workspace" aria-labelledby="candidate-001-workspace-title">
      <header class="candidate-001-workspace-head">
        <div class="candidate-001-workspace-name">
          <span class="candidate-001-kicker">SH FETCH</span>
          <h2 id="candidate-001-workspace-title">一体抓取</h2>
        </div>
        <output class="candidate-001-status" :data-state="taskState" aria-live="polite">
          <span aria-hidden="true" />{{ taskStateLabel }}
        </output>
      </header>

      <form class="candidate-001-form" @submit.prevent="startTask">
        <section class="candidate-001-row candidate-001-range-row" aria-labelledby="candidate-001-range-label">
          <div class="candidate-001-row-title">
            <span class="candidate-001-row-index" aria-hidden="true">01</span>
            <strong id="candidate-001-range-label">页码范围</strong>
          </div>
          <div class="candidate-001-range-fields">
            <label for="candidate-001-start-page">
              <span>起始页</span>
              <input
                id="candidate-001-start-page"
                v-model.number="startPage"
                name="start_page"
                type="number"
                inputmode="numeric"
                min="1"
              />
            </label>
            <span class="candidate-001-range-connector" aria-hidden="true">
              <i />
              至
              <i />
            </span>
            <label for="candidate-001-end-page">
              <span>结束页</span>
              <input
                id="candidate-001-end-page"
                v-model.number="endPage"
                name="end_page"
                type="number"
                inputmode="numeric"
                min="1"
              />
            </label>
          </div>
        </section>

        <fieldset class="candidate-001-row candidate-001-forum-row">
          <legend class="candidate-001-row-title">
            <span class="candidate-001-row-index" aria-hidden="true">02</span>
            <strong>目标版块</strong>
          </legend>
          <div class="candidate-001-forum-options" role="radiogroup" aria-label="目标版块">
            <button
              v-for="forum in forums"
              :key="forum.code"
              type="button"
              class="candidate-001-choice"
              :class="{ 'is-selected': targetForum === forum.code }"
              role="radio"
              :aria-checked="targetForum === forum.code"
              @click="targetForum = forum.code"
            >
              <span class="candidate-001-choice-dot" aria-hidden="true" />
              {{ forum.name }}
            </button>
          </div>
        </fieldset>

        <fieldset class="candidate-001-row candidate-001-policy-row">
          <legend class="candidate-001-row-title">
            <span class="candidate-001-row-index" aria-hidden="true">03</span>
            <strong>已有帖子处理</strong>
          </legend>
          <div class="candidate-001-policy-options" role="radiogroup" aria-label="已有帖子处理">
            <button
              type="button"
              class="candidate-001-choice"
              :class="{ 'is-selected': existingPolicy === 'upsert' }"
              role="radio"
              :aria-checked="existingPolicy === 'upsert'"
              @click="existingPolicy = 'upsert'"
            >
              <span class="candidate-001-choice-dot" aria-hidden="true" />
              更新已有
            </button>
            <button
              type="button"
              class="candidate-001-choice"
              :class="{ 'is-selected': existingPolicy === 'skip' }"
              role="radio"
              :aria-checked="existingPolicy === 'skip'"
              @click="existingPolicy = 'skip'"
            >
              <span class="candidate-001-choice-dot" aria-hidden="true" />
              跳过已有
            </button>
          </div>
        </fieldset>

        <footer class="candidate-001-command-bar">
          <div class="candidate-001-primary-command">
            <button
              class="candidate-001-command candidate-001-command-start"
              type="submit"
              :disabled="taskState === 'running' || taskState === 'paused'"
            >
              <span aria-hidden="true">▶</span>
              开始
            </button>
          </div>
          <div class="candidate-001-secondary-commands">
            <button
              class="candidate-001-command candidate-001-command-pause"
              type="button"
              :disabled="taskState !== 'running'"
              @click="pauseTask"
            >
              暂停
            </button>
            <button
              class="candidate-001-command candidate-001-command-resume"
              type="button"
              :disabled="taskState !== 'paused'"
              @click="resumeTask"
            >
              继续
            </button>
            <button
              class="candidate-001-command candidate-001-command-stop"
              type="button"
              :disabled="taskState !== 'running' && taskState !== 'paused'"
              @click="stopTask"
            >
              停止
            </button>
          </div>
        </footer>
      </form>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";

type Candidate001TaskState = "idle" | "running" | "paused" | "stopped";
type Candidate001ForumCode = "all" | "asia-original" | "chinese-subtitle";

const startPage = ref(1);
const endPage = ref(1);
const targetForum = ref<Candidate001ForumCode>("chinese-subtitle");
const existingPolicy = ref<"upsert" | "skip">("skip");
const taskState = ref<Candidate001TaskState>("idle");

const forums: Array<{ code: Candidate001ForumCode; name: string }> = [
  { code: "all", name: "全部" },
  { code: "asia-original", name: "亚洲视频原创" },
  { code: "chinese-subtitle", name: "高清中文字幕" }
];

const taskStateLabel = computed(() => {
  const labels: Record<Candidate001TaskState, string> = {
    idle: "待执行",
    running: "执行中",
    paused: "已暂停",
    stopped: "已停止"
  };
  return labels[taskState.value];
});

function startTask() {
  taskState.value = "running";
}

function pauseTask() {
  taskState.value = "paused";
}

function resumeTask() {
  taskState.value = "running";
}

function stopTask() {
  taskState.value = "stopped";
}
</script>

<style scoped>
.candidate-001-page {
  --candidate-001-ink: var(--text, #40373c);
  --candidate-001-muted: var(--muted, #7e6e77);
  --candidate-001-line: var(--line, #eadce3);
  --candidate-001-soft: var(--panel-soft, #fbf7f9);
  --candidate-001-panel: var(--panel, #fff);
  --candidate-001-accent: var(--primary, #c64a7b);
  --candidate-001-accent-deep: var(--primary-hover, #a93765);
  --candidate-001-accent-soft: var(--primary-soft, #fff0f6);
  --candidate-001-accent-line: var(--primary-border, #efb9ce);
  --candidate-001-danger: var(--danger, #d85c70);
  --candidate-001-danger-soft: var(--danger-soft, #fff0f2);
  display: grid;
  gap: 12px;
  min-width: 0;
}

.candidate-001-review-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
  padding: 4px 2px 10px;
  border-bottom: 1px solid var(--candidate-001-line);
}

.candidate-001-review-head > div {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.candidate-001-review-number {
  color: var(--candidate-001-accent-deep);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.candidate-001-review-head h1 {
  margin: 0;
  color: var(--candidate-001-ink);
  font-size: 15px;
  font-weight: 650;
  line-height: 22px;
}

.candidate-001-review-head p {
  margin: 0;
  color: var(--candidate-001-muted);
  font-size: 11px;
  line-height: 18px;
}

.candidate-001-workspace {
  overflow: hidden;
  border: 1px solid var(--candidate-001-line);
  border-radius: 10px;
  background: var(--candidate-001-panel);
  box-shadow: 0 1px 0 rgb(92 52 70 / 3%);
}

.candidate-001-workspace-head {
  display: flex;
  min-height: 46px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 8px 14px;
  border-bottom: 1px solid var(--candidate-001-line);
  background: var(--candidate-001-soft);
}

.candidate-001-workspace-name {
  display: flex;
  align-items: baseline;
  gap: 9px;
}

.candidate-001-kicker {
  color: var(--candidate-001-accent-deep);
  font-size: 9px;
  font-weight: 750;
  letter-spacing: 0.12em;
}

.candidate-001-workspace-name h2 {
  margin: 0;
  color: var(--candidate-001-ink);
  font-size: 13px;
  font-weight: 650;
  line-height: 20px;
}

.candidate-001-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--candidate-001-muted);
  font-size: 10px;
  font-weight: 600;
  line-height: 18px;
}

.candidate-001-status > span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--candidate-001-muted);
  box-shadow: 0 0 0 3px rgb(126 110 119 / 10%);
}

.candidate-001-status[data-state="running"] {
  color: var(--candidate-001-accent-deep);
}

.candidate-001-status[data-state="running"] > span {
  background: var(--candidate-001-accent);
  box-shadow: 0 0 0 3px var(--candidate-001-accent-soft);
}

.candidate-001-status[data-state="paused"] > span {
  background: var(--warn, #c18b38);
}

.candidate-001-status[data-state="stopped"] > span {
  background: var(--candidate-001-danger);
}

.candidate-001-form {
  display: grid;
}

.candidate-001-row {
  display: grid;
  grid-template-columns: 154px minmax(0, 1fr);
  min-width: 0;
  align-items: center;
  margin: 0;
  border: 0;
  border-bottom: 1px solid var(--candidate-001-line);
  padding: 0;
}

.candidate-001-row-title {
  box-sizing: border-box;
  display: flex;
  min-width: 0;
  height: 100%;
  align-items: center;
  gap: 10px;
  margin: 0;
  padding: 12px 14px;
  border-right: 1px solid var(--candidate-001-line);
  background: var(--candidate-001-soft);
}

.candidate-001-row-title strong {
  color: var(--candidate-001-ink);
  font-size: 11px;
  font-weight: 650;
  line-height: 18px;
  white-space: nowrap;
}

.candidate-001-row-index {
  display: grid;
  flex: 0 0 22px;
  width: 22px;
  height: 22px;
  place-items: center;
  border: 1px solid var(--candidate-001-accent-line);
  border-radius: 50%;
  background: var(--candidate-001-panel);
  color: var(--candidate-001-accent-deep);
  font-size: 8px;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
}

.candidate-001-range-fields {
  display: grid;
  grid-template-columns: minmax(132px, 180px) 72px minmax(132px, 180px);
  align-items: end;
  gap: 8px;
  padding: 10px 14px;
}

.candidate-001-range-fields label {
  display: grid;
  gap: 3px;
}

.candidate-001-range-fields label > span {
  color: var(--candidate-001-muted);
  font-size: 9px;
  font-weight: 550;
  line-height: 15px;
}

.candidate-001-range-fields input {
  box-sizing: border-box;
  width: 100%;
  height: 32px;
  border: 1px solid var(--candidate-001-line);
  border-radius: 6px;
  padding: 5px 9px;
  background: var(--candidate-001-panel);
  color: var(--candidate-001-ink);
  font: inherit;
  font-size: 12px;
  line-height: 20px;
  outline: 0;
  transition:
    border-color 140ms ease,
    box-shadow 140ms ease;
}

.candidate-001-range-fields input:hover {
  border-color: var(--candidate-001-accent-line);
}

.candidate-001-range-fields input:focus-visible {
  border-color: var(--candidate-001-accent);
  box-shadow: 0 0 0 3px var(--candidate-001-accent-soft);
}

.candidate-001-range-connector {
  display: flex;
  height: 32px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: var(--candidate-001-muted);
  font-size: 9px;
  line-height: 15px;
}

.candidate-001-range-connector i {
  width: 16px;
  height: 1px;
  background: var(--candidate-001-line);
}

.candidate-001-forum-options,
.candidate-001-policy-options {
  display: flex;
  min-width: 0;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  padding: 12px 14px;
}

.candidate-001-choice {
  box-sizing: border-box;
  display: inline-flex;
  min-height: 30px;
  align-items: center;
  gap: 6px;
  margin: 0;
  border: 1px solid transparent;
  border-radius: 6px;
  padding: 4px 9px;
  background: transparent;
  color: var(--candidate-001-muted);
  font: inherit;
  font-size: 11px;
  font-weight: 550;
  line-height: 20px;
  white-space: nowrap;
  cursor: pointer;
  transition:
    border-color 140ms ease,
    background-color 140ms ease,
    color 140ms ease;
}

.candidate-001-choice-dot {
  box-sizing: border-box;
  width: 11px;
  height: 11px;
  border: 1px solid var(--candidate-001-line);
  border-radius: 50%;
  background: var(--candidate-001-panel);
}

.candidate-001-choice:hover {
  border-color: var(--candidate-001-line);
  background: var(--candidate-001-soft);
  color: var(--candidate-001-ink);
}

.candidate-001-choice.is-selected {
  border-color: var(--candidate-001-accent-line);
  background: var(--candidate-001-accent-soft);
  color: var(--candidate-001-accent-deep);
}

.candidate-001-choice.is-selected .candidate-001-choice-dot {
  border: 3px solid var(--candidate-001-accent);
}

.candidate-001-choice:active {
  transform: translateY(1px);
}

.candidate-001-choice:focus-visible {
  outline: 2px solid var(--candidate-001-accent);
  outline-offset: 2px;
}

.candidate-001-command-bar {
  display: flex;
  min-height: 48px;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 14px;
  background: var(--candidate-001-soft);
}

.candidate-001-primary-command,
.candidate-001-secondary-commands {
  display: flex;
  align-items: center;
  gap: 6px;
}

.candidate-001-command {
  box-sizing: border-box;
  display: inline-flex;
  height: 30px;
  min-width: 58px;
  align-items: center;
  justify-content: center;
  gap: 5px;
  margin: 0;
  border: 1px solid var(--candidate-001-line);
  border-radius: 6px;
  padding: 4px 10px;
  background: var(--candidate-001-panel);
  color: var(--candidate-001-ink);
  font: inherit;
  font-size: 11px;
  font-weight: 600;
  line-height: 20px;
  cursor: pointer;
  transition:
    border-color 140ms ease,
    background-color 140ms ease,
    color 140ms ease,
    transform 140ms ease;
}

.candidate-001-command:hover:not(:disabled) {
  border-color: var(--candidate-001-accent-line);
  background: var(--candidate-001-accent-soft);
  color: var(--candidate-001-accent-deep);
}

.candidate-001-command:active:not(:disabled) {
  transform: translateY(1px);
}

.candidate-001-command:focus-visible {
  outline: 2px solid var(--candidate-001-accent);
  outline-offset: 2px;
}

.candidate-001-command:disabled {
  opacity: 0.42;
  cursor: not-allowed;
}

.candidate-001-command-start {
  min-width: 72px;
  border-color: var(--candidate-001-accent);
  background: var(--candidate-001-accent);
  color: #fff;
}

.candidate-001-command-start:hover:not(:disabled) {
  border-color: var(--candidate-001-accent-deep);
  background: var(--candidate-001-accent-deep);
  color: #fff;
}

.candidate-001-command-stop {
  border-color: color-mix(in srgb, var(--candidate-001-danger) 52%, var(--candidate-001-line));
  color: var(--candidate-001-danger);
}

.candidate-001-command-stop:hover:not(:disabled) {
  border-color: var(--candidate-001-danger);
  background: var(--candidate-001-danger-soft);
  color: var(--candidate-001-danger);
}

@media (prefers-reduced-motion: reduce) {
  .candidate-001-page :is(input, button) {
    transition: none;
  }
}
</style>
