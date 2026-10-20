<template>
  <section class="card next02-sc-panel smart-pick-copy-panel" aria-labelledby="scpickv2CopyTitle">
    <div class="card-body">
      <div class="next02-sc-panel-head">
        <div>
          <span class="next02-sc-panel-kicker">COPY STATUS</span>
          <h5 id="scpickv2CopyTitle">Copy 进度</h5>
        </div>
        <button type="button" class="btn btn-sm btn-outline-danger" :disabled="!active" @click="$emit('stop')">
          {{ status.state === "stopping" ? "正在停止..." : "停止当前任务" }}
        </button>
      </div>

      <div class="smart-pick-copy-status-row">
        <span class="smart-pick-copy-meta">{{ copyMeta }}</span>
        <span class="smart-pick-copy-percent">{{ percent }}%</span>
      </div>
      <div class="progress smart-pick-progress">
        <div
          class="progress-bar"
          :class="{ 'bg-danger': status.state === 'failed', 'bg-warning': status.state === 'stopped' }"
          role="progressbar"
          :style="{ width: `${percent}%` }"
          :aria-valuenow="percent"
          :aria-valuetext="`${percent}%`"
          aria-valuemin="0"
          aria-valuemax="100"
        ></div>
      </div>
      <div class="smart-pick-copy-current" :title="copyCurrent">
        {{ copyCurrent }}
      </div>
      <div v-if="status.state === 'failed'" class="smart-pick-copy-error">复制结束，存在失败文件</div>

      <details v-if="status.failures.length" class="scpickv2-copy-failures" open>
        <summary>{{ status.failures.length }} 个失败文件</summary>
        <ul>
          <li v-for="(failure, index) in status.failures" :key="`${failure.fileName}-${index}`">
            <strong>{{ failure.fileName || "未知文件" }}</strong>
            <span>{{ failure.error }}</span>
          </li>
        </ul>
      </details>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { ScCopyV2Status } from "@/features/sc/types";

const props = defineProps<{ status: ScCopyV2Status; active: boolean }>();
defineEmits<{ stop: [] }>();

const percent = computed(() => {
  if (props.status.totalBytes > 0)
    return Math.min(100, Math.round((props.status.copiedBytes / props.status.totalBytes) * 100));
  if (props.status.total > 0) return Math.min(100, Math.round((props.status.done / props.status.total) * 100));
  return 0;
});
const stateLabel = computed(() => {
  const labels: Record<ScCopyV2Status["state"], string> = {
    running: "复制中",
    stopping: "正在停止",
    completed: "复制完成",
    failed: "复制结束，存在失败",
    stopped: "已停止"
  };
  return labels[props.status.state];
});
const copyMeta = computed(
  () =>
    `${stateLabel.value} · ${props.status.done}/${props.status.total} · 成功 ${props.status.success} · 失败 ${props.status.failed} · 跳过 ${props.status.skipped}`
);
const copyCurrent = computed(() => {
  const progress = `${formatBytes(props.status.copiedBytes)} / ${formatBytes(props.status.totalBytes)}`;
  const current = props.status.current ? ` · 当前：${props.status.current}` : "";
  return `${progress}${current} · 任务 ID：${props.status.jobID}`;
});

function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = value;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex += 1;
  }
  return `${size.toFixed(unitIndex === 0 ? 0 : 2)} ${units[unitIndex]}`;
}
</script>
