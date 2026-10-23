<template>
  <div class="col-12 col-md-3" :id="wrapperId">
    <label class="form-label mb-1" :for="id">{{ label }}</label>
    <input
      :id="id"
      class="form-control"
      type="number"
      min="0"
      step="1"
      :value="value"
      @input="$emit('update', field, ($event.target as HTMLInputElement).value)"
    />
  </div>
</template>

<script setup lang="ts">
import type { DSeedEditModel } from "./models";

withDefaults(
  defineProps<{
    id: string;
    label: string;
    field: keyof DSeedEditModel;
    value: string;
    wrapperId?: string;
  }>(),
  { wrapperId: undefined }
);

defineEmits<{
  update: [field: keyof DSeedEditModel, value: string];
}>();
</script>
