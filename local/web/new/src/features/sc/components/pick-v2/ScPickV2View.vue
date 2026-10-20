<template>
  <PageShell
    class="sc-events-page-shell sc-pick-page-shell cards-page-shell next02-pink-cards-page next02-pink-cards-page-shell"
    title="SC Pick"
    eyebrow=""
  >
    <div class="container-fluid sc-pick-page-root next02-sc-pink-theme next02-smart-pick-page-rewrite">
      <section class="next02-sc-movie-top-shell next02-smart-pick-hero">
        <div class="next02-sc-hero-copy">
          <span class="next02-sc-eyebrow">SC / PICK</span>
          <h4>SC Pick</h4>
          <p>先抽取并预览结果，再将当前结果作为独立任务复制。</p>
        </div>
        <div class="next02-sc-hero-actions">
          <a href="/scevents" class="btn btn-outline-secondary" target="_blank" rel="noopener noreferrer">SC Events</a>
        </div>
      </section>

      <div v-if="message" class="next02-sc-feedback is-success" role="status">
        <span class="smart-pick-feedback-message">{{ message }}</span>
      </div>
      <div v-if="error" class="next02-sc-feedback is-error" role="alert">
        <span class="smart-pick-feedback-message">{{ error }}</span>
      </div>

      <ScPickForm :loading="picking || copyActive" @submit="handleSubmit" @invalid="handleInvalid" />
      <ScCopyJobPanel v-if="copyStatus" :status="copyStatus" :active="copyActive" @stop="stopCopy" />
      <ScPickResults
        v-if="preview"
        :preview="preview"
        :copy-starting="copyStarting"
        :copy-disabled="copyStarting || copyActive || Boolean(copyStatus)"
        @copy="handleCopy"
      />
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import ScCopyJobPanel from "@/features/sc/components/pick-v2/ScCopyJobPanel.vue";
import ScPickForm from "@/features/sc/components/pick-v2/ScPickForm.vue";
import ScPickResults from "@/features/sc/components/pick-v2/ScPickResults.vue";
import { useScPickV2Run } from "@/features/sc/composables/useScPickV2Run";
import type { ScPickV2Request } from "@/features/sc/types";
import PageShell from "@/shared/components/PageShell.vue";

const {
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
} = useScPickV2Run();

onMounted(() => void resumeCopy());

function handleSubmit(request: ScPickV2Request) {
  if (copyActive.value) {
    handleInvalid("当前复制任务仍在运行，请等待任务结束或停止后再执行新的抽取");
    return;
  }
  void runPick(request).catch(() => undefined);
}

function handleInvalid(reason: string) {
  message.value = "";
  error.value = reason;
}

function handleCopy() {
  void startCopy().catch(() => undefined);
}
</script>
