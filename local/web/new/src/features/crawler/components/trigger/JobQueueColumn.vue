<template>
  <div class="queue-column">
    <div class="queue-title">{{ title }}</div>
    <div :id="id" class="queue-list">
      <div v-if="!items.length" class="queue-empty">{{ emptyText }}</div>
      <div v-for="(item, index) in items" v-else :key="itemKey(item, index)" class="queue-item">
        <div class="queue-item-top">
          <div class="queue-item-left">
            <span class="queue-item-index">{{ itemIndex(item, index) }}</span>
            <a
              v-if="movieName(item)"
              class="queue-item-name"
              :href="`/moviedetail/${encodeURIComponent(movieName(item))}`"
            >
              {{ movieName(item) }}
            </a>
            <span v-else class="queue-item-name">-</span>
          </div>
          <span class="queue-badge" :class="badgeClass(item)">{{ badgeText(item) }}</span>
        </div>
        <div v-if="showCompletedAt && completedAt(item)" class="queue-item-meta">
          完成时间：<ScDateTime :value="completedAt(item)" unix seconds />
        </div>
        <div v-if="errorText(item)" class="queue-item-error">{{ errorText(item) }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import ScDateTime from "@/shared/components/ScDateTime.vue";
import type { CrawlerRecord } from "@/features/crawler/types";
import { pick, text } from "@/shared/utils/format";

withDefaults(
  defineProps<{
    id: string;
    title: string;
    items: CrawlerRecord[];
    emptyText: string;
    showCompletedAt?: boolean;
  }>(),
  { showCompletedAt: false }
);

function movieName(item: CrawlerRecord): string {
  return text(item, "movie_name", "MovieName");
}

function errorText(item: CrawlerRecord): string {
  return text(item, "error", "Error");
}

function completedAt(item: CrawlerRecord): number {
  return Number(pick(item, "completed_at", "CompletedAt") || 0);
}

function itemIndex(item: CrawlerRecord, index: number): string {
  const sequence = Number(pick(item, "seq", "Seq") || 0);
  return String(sequence > 0 ? sequence : index + 1).padStart(2, "0");
}

function itemKey(item: CrawlerRecord, index: number): string {
  return `${movieName(item)}:${text(item, "movie_jav_id", "MovieJavID")}:${text(item, "state", "State")}:${itemIndex(item, index)}`;
}

function badgeClass(item: CrawlerRecord): string {
  const state = text(item, "state", "State");
  if (state === "running") return "is-running";
  if (state === "success") return "is-success";
  if (state === "failed") return "is-failed";
  return "is-pending";
}

function badgeText(item: CrawlerRecord): string {
  const state = text(item, "state", "State");
  if (state === "running") return "处理中";
  if (state === "success") return "成功";
  if (state === "failed") return "失败";
  return "待入列";
}
</script>
