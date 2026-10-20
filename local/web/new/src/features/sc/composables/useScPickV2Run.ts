import { computed, onBeforeUnmount, ref } from "vue";
import { getScPickV2CopyStatus, runScPickV2, startScPickV2Copy, stopScPickV2Copy } from "@/features/sc/api";
import type { ScCopyV2Status, ScPickV2Request, ScPickV2Response } from "@/features/sc/types";

const storedJobKey = "scpickv2-copy-job";

export function useScPickV2Run() {
  const preview = ref<ScPickV2Response | null>(null);
  const copyStatus = ref<ScCopyV2Status | null>(null);
  const picking = ref(false);
  const copyStarting = ref(false);
  const message = ref("");
  const error = ref("");
  let pollTimer: number | undefined;

  const copyActive = computed(() => copyStatus.value?.state === "running" || copyStatus.value?.state === "stopping");

  async function runPick(request: ScPickV2Request) {
    stopPolling();
    picking.value = true;
    message.value = "";
    error.value = "";
    copyStatus.value = null;
    try {
      preview.value = await runScPickV2(request);
      message.value = `抽取完成：JAV ${preview.value.jav.picked} 部，FC2 ${preview.value.fc2.picked} 部`;
    } catch (reason) {
      preview.value = null;
      error.value = reason instanceof Error ? reason.message : String(reason);
      throw reason;
    } finally {
      picking.value = false;
    }
  }

  async function startCopy() {
    if (!preview.value?.previewID) throw new Error("当前没有可复制的抽取结果");
    copyStarting.value = true;
    error.value = "";
    try {
      copyStatus.value = await startScPickV2Copy(preview.value.previewID);
      sessionStorage.setItem(storedJobKey, copyStatus.value.jobID);
      message.value = `复制任务已启动：${copyStatus.value.jobID}`;
      schedulePoll();
    } catch (reason) {
      error.value = reason instanceof Error ? reason.message : String(reason);
      throw reason;
    } finally {
      copyStarting.value = false;
    }
  }

  async function stopCopy() {
    if (!copyStatus.value?.jobID) return;
    error.value = "";
    try {
      copyStatus.value = await stopScPickV2Copy(copyStatus.value.jobID);
      message.value = "已发送停止指令，正在等待当前文件退出";
      schedulePoll();
    } catch (reason) {
      error.value = reason instanceof Error ? reason.message : String(reason);
    }
  }

  async function resumeCopy() {
    const jobID = sessionStorage.getItem(storedJobKey);
    if (!jobID) return;
    try {
      copyStatus.value = await getScPickV2CopyStatus(jobID);
      if (copyActive.value) schedulePoll();
    } catch {
      sessionStorage.removeItem(storedJobKey);
    }
  }

  function schedulePoll() {
    stopPolling();
    if (!copyStatus.value?.jobID || !copyActive.value) return;
    pollTimer = window.setTimeout(() => void pollCopy(), 2000);
  }

  async function pollCopy() {
    if (!copyStatus.value?.jobID) return;
    try {
      copyStatus.value = await getScPickV2CopyStatus(copyStatus.value.jobID);
      if (!copyActive.value) {
        message.value = copyStatus.value.state === "completed" ? "复制完成" : `复制结束：${copyStatus.value.state}`;
      }
    } catch (reason) {
      error.value = reason instanceof Error ? reason.message : String(reason);
    } finally {
      schedulePoll();
    }
  }

  function stopPolling() {
    if (pollTimer === undefined) return;
    window.clearTimeout(pollTimer);
    pollTimer = undefined;
  }

  onBeforeUnmount(stopPolling);

  return {
    preview,
    copyStatus,
    picking,
    copyStarting,
    copyActive,
    message,
    error,
    runPick,
    startCopy,
    stopCopy,
    resumeCopy
  };
}
