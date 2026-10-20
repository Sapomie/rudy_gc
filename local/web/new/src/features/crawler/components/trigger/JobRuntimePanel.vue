<template>
  <div
    class="card next02-panel next02-metric-card"
    :class="{
      'next02-seeds-panel next02-seeds-runtime-panel next02-pink-trigger-panel next02-pink-trigger-runtime-panel': pink
    }"
  >
    <div class="card-body">
      <div class="next02-panel-head" :class="{ 'next02-seeds-panel-head next02-pink-trigger-panel-head': pink }">
        <div>
          <div v-if="commandVisual" class="next02-panel-eyebrow">{{ eyebrow }}</div>
          <h4 class="card-title mb-0">运行概览</h4>
        </div>
        <div class="d-flex align-items-center gap-2">
          <JobSelector :jobs="jobs" :selected-job-id="selectedJobId" @select="$emit('select-job', $event)" />
          <span
            v-if="commandVisual"
            class="next02-live-dot"
            :class="{
              'is-idle': runtimeOverview && !currentJob,
              'is-complete': runtimeOverview && currentJobDone,
              'is-failed': runtimeOverview && currentJobDone && text(currentJob || {}, 'stage', 'Stage') === 'failed'
            }"
            >{{ badgeText }}</span
          >
        </div>
      </div>

      <TriggerOverview
        :show-runtime-context="runtimeOverview && Boolean(currentJob)"
        :runtime-context="runtimeContext"
        :stage-mode="stageMode"
        :stage-groups="stageGroups"
        :show-metrics="showMetrics"
        :items="metrics"
      >
        <div v-if="showQueue" class="queue-panel mt-2">
          <JobQueueColumn
            id="filteredQueuePendingList"
            title="待入列 / 处理中"
            :items="activeItems"
            :empty-text="selectedJobId ? '当前任务没有待处理影片' : '等待任务启动'"
          />
          <JobQueueColumn
            id="filteredQueueDoneList"
            title="已完成"
            :items="doneItems"
            :empty-text="selectedJobId ? '当前还没有处理完成的影片' : '等待任务启动'"
            show-completed-at
          />
        </div>
        <div
          v-if="showProgress"
          class="progress mt-3"
          :class="{ 'next02-seeds-progress next02-pink-trigger-progress': pink, 'is-complete': currentJobDone }"
        >
          <div
            id="progressBar"
            class="progress-bar progress-bar-striped progress-bar-animated"
            role="progressbar"
            :style="{ width: `${progressPercent}%` }"
            aria-valuemin="0"
            aria-valuemax="100"
            :aria-valuenow="progressPercent"
          >
            {{ progressPercent }}%
          </div>
        </div>
        <div
          class="mt-3 d-flex align-items-center justify-content-between"
          :class="{ 'next02-seeds-runtime-footer next02-pink-trigger-runtime-footer': pink }"
        >
          <span id="statusMsg" :class="statusClass">{{ statusText }}</span>
          <span id="eventCountText" class="badge bg-light text-dark">事件 {{ eventCount }} 条</span>
        </div>
      </TriggerOverview>
    </div>
  </div>
</template>

<script setup lang="ts">
import JobQueueColumn from "./JobQueueColumn.vue";
import JobSelector from "./JobSelector.vue";
import TriggerOverview from "./TriggerOverview.vue";
import type { TriggerMetric, TriggerMetricGroup } from "./models";
import type { CrawlerRecord } from "@/features/crawler/types";
import { text } from "@/shared/utils/format";

defineProps<{
  pink: boolean;
  commandVisual: boolean;
  eyebrow: string;
  runtimeOverview: boolean;
  currentJob?: CrawlerRecord;
  currentJobDone: boolean;
  badgeText: string;
  jobs: CrawlerRecord[];
  selectedJobId: string;
  runtimeContext: TriggerMetric[];
  stageMode: boolean;
  stageGroups: TriggerMetricGroup[];
  showMetrics: boolean;
  metrics: TriggerMetric[];
  showQueue: boolean;
  activeItems: CrawlerRecord[];
  doneItems: CrawlerRecord[];
  showProgress: boolean;
  progressPercent: number;
  statusClass: string;
  statusText: string;
  eventCount: number;
}>();

defineEmits<{ "select-job": [jobId: string] }>();
</script>
