<template>
  <label v-if="jobs.length" class="next02-job-selector">
    <span class="sr-only">选择任务实例</span>
    <select class="form-select form-select-sm" :value="selectedJobId" @change="selectJob">
      <option value="">选择任务实例</option>
      <option v-for="job in jobs" :key="jobId(job)" :value="jobId(job)">
        {{ jobLabel(job) }}
      </option>
    </select>
  </label>
</template>

<script setup lang="ts">
import type { CrawlerRecord } from "@/features/crawler/types";
import { pick, text } from "@/shared/utils/format";

defineProps<{
  jobs: CrawlerRecord[];
  selectedJobId: string;
}>();

const emit = defineEmits<{ select: [jobId: string] }>();

function jobId(job: CrawlerRecord): string {
  return String(pick(job, "id", "ID", "job_id", "JobID") || "");
}

function jobLabel(job: CrawlerRecord): string {
  const task = text(job, "task_type", "TaskType") || "任务";
  const stage = text(job, "stage", "Stage") || "等待";
  return `${task} · ${stage} · ${jobId(job)}`;
}

function selectJob(event: Event) {
  emit("select", (event.target as HTMLSelectElement).value);
}
</script>
