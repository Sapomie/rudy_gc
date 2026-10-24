<template>
  <button
    type="button"
    class="candidate-button"
    :class="[
      tone === 'primary' ? 'button-primary' : tone === 'danger' ? 'button-danger' : 'button-secondary',
      { 'candidate-button-icon-only': iconOnly }
    ]"
    :data-candidate="candidate"
    :data-tone="tone"
    :data-preview="preview"
    :disabled="disabled || loading"
    :aria-busy="loading || undefined"
    :aria-label="iconOnly ? label : undefined"
    :title="iconOnly ? label : undefined"
  >
    <span v-if="loading" class="candidate-button-spinner" aria-hidden="true" />
    <AppIcon v-else-if="icon" :name="icon" />
    <span v-if="!iconOnly">{{ label }}</span>
  </button>
</template>

<script setup lang="ts">
import AppIcon from "@/shared/components/AppIcon.vue";

withDefaults(
  defineProps<{
    label: string;
    candidate?: "a" | "b" | "c";
    tone?: "primary" | "neutral" | "danger";
    icon?: string;
    iconOnly?: boolean;
    disabled?: boolean;
    loading?: boolean;
    preview?: "hover" | "pressed" | "focus";
  }>(),
  {
    candidate: "a",
    tone: "neutral",
    icon: undefined,
    iconOnly: false,
    disabled: false,
    loading: false,
    preview: undefined
  }
);
</script>
