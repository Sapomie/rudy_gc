<template>
  <div class="col-lg-7" :class="{ 'next02-seeds-event-column next02-pink-trigger-event-column': pink }">
    <div
      class="card event-card next02-panel next02-event-card"
      :class="{
        'next02-seeds-panel next02-seeds-event-panel next02-pink-trigger-panel next02-pink-trigger-event-panel': pink
      }"
    >
      <div class="card-body">
        <div
          class="d-flex align-items-center justify-content-between mb-3 next02-panel-head"
          :class="{ 'next02-seeds-panel-head next02-pink-trigger-panel-head': pink }"
        >
          <div>
            <div v-if="commandVisual" class="next02-panel-eyebrow">{{ eyebrow }}</div>
            <h4 class="card-title mb-0">{{ title }}</h4>
          </div>
          <div
            class="custom-control custom-checkbox"
            :class="{ 'next02-seeds-raw-toggle next02-pink-trigger-raw-toggle': pink }"
          >
            <input
              id="debugRaw"
              :checked="showRaw"
              type="checkbox"
              class="custom-control-input"
              @change="$emit('update:showRaw', ($event.target as HTMLInputElement).checked)"
            />
            <label class="custom-control-label" for="debugRaw">显示原始 JSON</label>
          </div>
        </div>
        <div class="result-box" :class="{ 'next02-seeds-result-box next02-pink-trigger-result-box': pink }">
          <pre id="resultBox"><template v-if="entries.length"><span
            v-for="entry in entries"
            :key="entry.key"
            class="event-line"
            :class="`log-${entry.level}`"
            v-html="entry.html"
          /></template><span v-else class="event-empty">{{ emptyText }}</span></pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TriggerEventEntry } from "./models";

defineProps<{
  pink: boolean;
  commandVisual: boolean;
  eyebrow: string;
  title: string;
  showRaw: boolean;
  entries: TriggerEventEntry[];
  emptyText: string;
}>();

defineEmits<{ "update:showRaw": [value: boolean] }>();
</script>
