<template>
  <span class="sc-event-time" :title="fullText">
    <span class="sc-event-date">{{ parts.date }}</span>
    <small v-if="parts.time" class="sc-event-hour">{{ parts.time }}</small>
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    value: unknown;
    unix?: boolean;
    seconds?: boolean;
  }>(),
  {
    unix: false,
    seconds: false
  }
);

const parts = computed(() => (props.unix ? unixParts(props.value, props.seconds) : textParts(props.value)));
const fullText = computed(() => [parts.value.date, parts.value.time].filter(Boolean).join(" "));

function unixParts(value: unknown, seconds: boolean): { date: string; time: string } {
  const stamp = Number(value || 0);
  if (!stamp) return { date: "-", time: "" };
  const date = new Date(stamp * 1000);
  if (Number.isNaN(date.getTime())) return { date: "-", time: "" };
  const pad = (part: number) => String(part).padStart(2, "0");
  return {
    date: `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`,
    time: `${pad(date.getHours())}:${pad(date.getMinutes())}${seconds ? `:${pad(date.getSeconds())}` : ""}`
  };
}

function textParts(value: unknown): { date: string; time: string } {
  const raw = String(value ?? "").trim();
  if (!raw || raw === "0") return { date: "-", time: "" };
  const matched = raw.match(/^(\d{4}[-/]\d{2}[-/]\d{2})(?:[T\s]+(\d{2}:\d{2}(?::\d{2})?))?/);
  if (!matched) return { date: raw, time: "" };
  return { date: matched[1].replaceAll("/", "-"), time: matched[2] || "" };
}
</script>

<style scoped>
@layer business {
  .sc-event-time {
    display: inline-flex;
    min-width: 0;
    align-items: baseline;
    justify-content: center;
    gap: 5px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .sc-event-date {
    flex: 0 0 auto;
    color: var(--text);
    font-size: 13px;
    font-weight: 600;
    line-height: 1.15;
  }

  .sc-event-hour {
    flex: 0 0 auto;
    color: var(--primary);
    font-size: 11px;
    font-weight: 600;
    line-height: 1.15;
  }
}
</style>
