<template>
  <div class="crawl-overview">
    <div v-if="showRuntimeContext && runtimeContext.length" class="sh-list-runtime-context">
      <span v-for="item in runtimeContext" :key="item.label">
        <small>{{ item.label }}</small>
        <strong>{{ item.value }}</strong>
      </span>
    </div>
    <div v-if="stageMode" class="next02-seeds-stage-groups">
      <section v-for="group in stageGroups" :key="group.label" class="next02-seeds-stage-group">
        <div class="next02-seeds-stage-label">{{ group.label }}</div>
        <div class="next02-seeds-stage-grid">
          <div v-for="item in group.items" :key="item.label" class="overview-item next02-seeds-metric-unit">
            <div class="label">{{ item.label }}</div>
            <div class="value">{{ item.value }}</div>
          </div>
        </div>
      </section>
    </div>
    <div v-else-if="showMetrics" class="status-grid" :class="{ 'sehuatang-list-progress-grid': showRuntimeContext }">
      <div v-for="item in items" :key="item.label" class="overview-item">
        <div class="label">{{ item.label }}</div>
        <div class="value">{{ item.value }}</div>
      </div>
    </div>
    <slot></slot>
  </div>
</template>

<script setup lang="ts">
import type { TriggerMetric, TriggerMetricGroup } from "./models";

defineProps<{
  showRuntimeContext: boolean;
  runtimeContext: TriggerMetric[];
  stageMode: boolean;
  stageGroups: TriggerMetricGroup[];
  showMetrics: boolean;
  items: TriggerMetric[];
}>();
</script>
