<template>
  <section class="next02-wdir-children" aria-label="Subdirectories">
    <div class="next02-wdir-children-head">
      <h2>{{ text(children, "Title") || "子目录" }}</h2>
      <a v-if="text(children, 'MoreURL')" :href="text(children, 'MoreURL')">更多</a>
    </div>
    <div v-if="items.length" class="next02-wdir-directory-grid">
      <a
        v-for="item in items"
        :key="itemKey(item)"
        :href="`/mediadir?id=${encodeURIComponent(itemId(item))}`"
        class="next02-wdir-directory-item"
      >
        <strong class="next02-wdir-directory-name" :title="`${itemName(item)} · ${typeLabel(item)}`">
          {{ itemName(item) }}
          <small class="badge rounded-pill ms-2" :class="directoryTypeBadgeClass(mediaType(item))">
            {{ typeLabel(item) }}
          </small>
        </strong>
        <span class="next02-wdir-directory-metric"
          ><small>媒体</small><b>{{ filmCount(item) }}</b></span
        >
        <span class="next02-wdir-directory-metric"
          ><small>容量</small><b>{{ totalSize(item) }}</b></span
        >
      </a>
    </div>
    <div v-else class="movie-empty next02-wdir-empty">暂无子目录。</div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { MediaRecord } from "@/features/media/types";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

const props = defineProps<{ children: MediaRecord }>();
const items = computed(() => asArray<MediaRecord>(props.children.Items || props.children.items));

function summary(item: MediaRecord): MediaRecord {
  return asRecord(pick(item, "Summary"));
}

function stat(item: MediaRecord): MediaRecord {
  return asRecord(asArray<MediaRecord>(pick(item, "Stats"))[0]);
}

function itemId(item: MediaRecord): string {
  return text(summary(item), "id", "Id");
}

function itemName(item: MediaRecord): string {
  return text(summary(item), "name", "Name");
}

function itemKey(item: MediaRecord): string {
  return itemId(item) || itemName(item);
}

function mediaType(item: MediaRecord): string {
  return text(stat(item), "media_type", "MediaType") || "empty";
}

function typeLabel(item: MediaRecord): string {
  return directoryTypeLabel(mediaType(item));
}

function filmCount(item: MediaRecord): string {
  return text(stat(item), "film_count", "FilmCount") || "0";
}

function totalSize(item: MediaRecord): string {
  return humanSize(pick(stat(item), "total_size", "TotalSize"));
}

function humanSize(input: unknown): string {
  const value = Number(input || 0);
  if (value >= 1024 ** 4) return `${(value / 1024 ** 4).toFixed(1)} TB`;
  if (value >= 1024 ** 3) return `${(value / 1024 ** 3).toFixed(1)} GB`;
  if (value >= 1024 ** 2) return `${(value / 1024 ** 2).toFixed(1)} MB`;
  if (value >= 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${value} B`;
}

function directoryTypeLabel(value: string): string {
  if (value === "jav") return "JAV";
  if (value === "fc2ppv") return "FC2";
  if (value === "mixed") return "混合";
  return "空";
}

function directoryTypeBadgeClass(value: string): string {
  if (value === "jav") return "text-bg-secondary";
  if (value === "fc2ppv") return "text-bg-warning";
  if (value === "mixed") return "text-bg-info";
  return "text-bg-light";
}
</script>
