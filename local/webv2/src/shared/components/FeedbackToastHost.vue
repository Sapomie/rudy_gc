<template>
  <Teleport to="body">
    <TransitionGroup
      name="feedback-toast"
      tag="div"
      class="feedback-toast-root"
      aria-label="操作通知"
      aria-live="polite"
      aria-relevant="additions"
    >
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="floating-feedback"
        :class="toast.tone"
        :role="toast.tone === 'error' ? 'alert' : 'status'"
      >
        <span>{{ toast.message }}</span>
        <a v-if="toast.action" class="floating-feedback-action" :href="toast.action.href">
          {{ toast.action.label }}
        </a>
        <button type="button" aria-label="关闭通知" @click="dismiss(toast.id)">×</button>
      </div>
    </TransitionGroup>
  </Teleport>
</template>

<script setup lang="ts">
import { useFeedbackToast } from "@/shared/composables/useFeedbackToast";

const { toasts, dismiss } = useFeedbackToast();
</script>
