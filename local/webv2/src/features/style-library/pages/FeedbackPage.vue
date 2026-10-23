<template>
  <div class="style-library-view">
    <section class="style-library-section">
      <header class="style-library-section-head">
        <h2>状态与任务反馈</h2>
        <p>示例进度可启动、暂停、继续和重置</p>
      </header>
      <div class="style-library-section-body">
        <div class="style-library-actions">
          <StatusBadge v-for="state in semanticStates" :key="state.label" v-bind="state" />
        </div>
        <div class="style-library-progress-head">
          <strong>{{
            progress === 100 ? "示例已完成" : running ? "正在处理示例" : progress ? "示例已暂停" : "等待开始"
          }}</strong
          ><span>{{ progress }}%</span>
        </div>
        <div
          class="progress next02-pink-trigger-progress style-library-progress"
          :class="{ 'is-complete': progress === 100, 'is-paused': !running }"
        >
          <div
            class="progress-bar progress-bar-striped progress-bar-animated"
            role="progressbar"
            aria-label="示例任务进度"
            :aria-valuenow="progress"
            aria-valuemin="0"
            aria-valuemax="100"
            :style="{ width: `${progress}%` }"
          />
        </div>
        <div class="style-library-actions">
          <button type="button" class="button-primary" :disabled="running || progress === 100" @click="start">
            {{ progress ? "继续" : "开始演示" }}</button
          ><button type="button" class="button-secondary" :disabled="!running" @click="pause">暂停</button
          ><button type="button" class="button-secondary" :disabled="!progress && !running" @click="reset">重置</button
          ><span class="style-library-note" role="status">{{
            progress === 100 ? "全部示例处理完成" : "仅演示进度变化"
          }}</span>
        </div>
        <div class="job-event-stream style-library-console" role="log" aria-label="示例任务日志">
          <div v-for="(entry, index) in logs" :key="index" class="event-line" :class="entry.tone">
            <span>{{ String(index + 1).padStart(2, "0") }}</span> {{ entry.text }}
          </div>
        </div>
      </div>
    </section>
    <div class="style-library-columns">
      <section class="style-library-section">
        <header class="style-library-section-head">
          <h2>加载与空状态</h2>
          <p>明确当前内容状态</p>
        </header>
        <div class="style-library-section-body">
          <LoadingState layout="rows" :count="3" label="正在加载示例…" />
          <div class="state-line">暂无内容，调整筛选后重试。</div>
        </div>
      </section>
      <section class="style-library-section">
        <header class="style-library-section-head">
          <h2>提示与确认</h2>
          <p>成功、错误与危险操作</p>
        </header>
        <div class="style-library-section-body">
          <div class="state-line success" role="status">保存成功，示例设置已更新。</div>
          <template v-if="!retryDone"
            ><ErrorPanel message="示例连接失败，请重试。" /><button
              type="button"
              class="button-secondary"
              @click="retryDone = true"
            >
              重试示例
            </button></template
          >
          <div v-else class="state-line success" role="status">重试成功，示例内容已恢复。</div>
          <div class="style-library-actions">
            <button type="button" class="button-secondary" @click="openDialog(false)">普通确认</button
            ><button type="button" class="button-danger" @click="openDialog(true)">危险操作确认</button>
          </div>
          <p class="style-library-note" role="status">{{ result || "弹层支持取消、Esc 关闭与键盘焦点切换。" }}</p>
        </div>
      </section>
    </div>
    <dialog
      ref="dialog"
      class="modal-panel-v2 style-library-dialog"
      aria-labelledby="style-confirm-title"
      aria-describedby="style-confirm-description"
      @click.self="dialog?.close()"
    >
      <div class="modal-header-v2">
        <h2 id="style-confirm-title">{{ danger ? "移除示例记录" : "保存示例设置" }}</h2>
        <button type="button" class="modal-close-v2" aria-label="关闭确认" @click="dialog?.close()">×</button>
      </div>
      <div class="modal-body-v2">
        <p id="style-confirm-description">
          {{ danger ? "将演示移除所选记录的确认流程。" : "将演示保存当前设置的确认流程。" }}
        </p>
        <p class="style-library-note">操作仅更新本页演示结果。</p>
      </div>
      <div class="modal-footer-v2">
        <button type="button" class="button-secondary" autofocus @click="dialog?.close()">取消</button
        ><button type="button" :class="danger ? 'button-danger' : 'button-primary'" @click="confirm">
          {{ danger ? "移除示例" : "保存示例" }}
        </button>
      </div>
    </dialog>
  </div>
</template>
<script setup lang="ts">
import { onUnmounted, ref } from "vue";
import StatusBadge from "@/features/remote-dailybest/components/StatusBadge.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import { semanticStates } from "../catalog";
const progress = ref(0);
const running = ref(false);
const retryDone = ref(false);
const danger = ref(false);
const result = ref("");
const dialog = ref<HTMLDialogElement>();
const logs = ref([{ text: "示例已就绪，等待开始。", tone: "log-info" }]);
let timer: ReturnType<typeof setInterval> | undefined;
function stopTimer() {
  if (timer !== undefined) clearInterval(timer);
  timer = undefined;
  running.value = false;
}
function start() {
  if (running.value || progress.value === 100) return;
  running.value = true;
  logs.value.push({ text: progress.value ? "继续处理。" : "开始处理。", tone: "log-info" });
  timer = setInterval(() => {
    progress.value += 20;
    logs.value.push({ text: progress.value === 100 ? "处理完成。" : `已处理 ${progress.value}%。`, tone: "log-info" });
    if (progress.value === 100) stopTimer();
  }, 800);
}
function pause() {
  stopTimer();
  logs.value.push({ text: "已暂停，等待继续。", tone: "log-warn" });
}
function reset() {
  stopTimer();
  progress.value = 0;
  logs.value = [{ text: "示例已重置，等待开始。", tone: "log-info" }];
}
function openDialog(isDanger: boolean) {
  danger.value = isDanger;
  dialog.value?.showModal();
}
function confirm() {
  result.value = danger.value ? "已完成移除示例" : "已保存示例设置";
  dialog.value?.close();
}
onUnmounted(stopTimer);
</script>
