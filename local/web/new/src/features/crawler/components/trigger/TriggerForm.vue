<template>
  <div class="card next02-panel next02-task-card" :class="{ 'next02-seeds-panel next02-pink-trigger-panel': pink }">
    <div class="card-body">
      <div class="next02-panel-head" :class="{ 'next02-seeds-panel-head next02-pink-trigger-panel-head': pink }">
        <div>
          <div v-if="commandVisual" class="next02-panel-eyebrow">{{ eyebrow }}</div>
          <h4 class="card-title mb-0">{{ title }}</h4>
        </div>
        <span v-if="commandVisual" class="next02-panel-status">{{ status }}</span>
      </div>
      <div v-if="helper" class="alert alert-info small py-2 px-3 mb-3 next02-helper-alert">{{ helper }}</div>
      <slot></slot>
      <p v-if="message" class="task-message">{{ message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  pink: boolean;
  commandVisual: boolean;
  eyebrow: string;
  title: string;
  status: string;
  helper: string;
  message: string;
}>();
</script>
