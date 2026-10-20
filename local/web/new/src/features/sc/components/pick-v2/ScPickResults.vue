<template>
  <section class="smart-pick-results smart-pick-mixed-results" aria-labelledby="scpickv2ResultsTitle">
    <div class="card next02-sc-panel smart-pick-results-panel">
      <div class="card-body">
        <div class="next02-sc-panel-head smart-pick-results-head">
          <div>
            <span class="next02-sc-panel-kicker">MIXED PICKED MOVIES</span>
            <h5 id="scpickv2ResultsTitle">混合抽取预览</h5>
          </div>
          <div class="scpickv2-results-actions">
            <span class="next02-sc-result-count">共 {{ preview.picked }} 部 · {{ preview.totalSizeGB }} GB</span>
            <button type="button" class="btn btn-primary" :disabled="copyDisabled" @click="$emit('copy')">
              {{ copyStarting ? "正在创建复制任务..." : "复制当前结果" }}
            </button>
          </div>
        </div>
        <div class="scpickv2-preview-meta" :title="preview.previewID">预览 ID：{{ preview.previewID }}</div>
      </div>
    </div>

    <section v-if="preview.jav.movies.length" class="smart-pick-result-section is-jav">
      <div class="smart-pick-result-section-head">
        <div>
          <span>JAV</span>
          <h5>JAV 结果</h5>
        </div>
        <div class="smart-pick-sort-tools">
          <div class="btn-group btn-group-sm flex-wrap" role="group" aria-label="JAV 结果排序">
            <button
              v-for="option in javSortOptions"
              :key="option.key"
              type="button"
              class="btn btn-outline-primary"
              :class="{ active: javSort === option.key }"
              @click="selectJAVSort(option.key)"
            >
              {{ option.label
              }}{{ javSort === option.key && option.key !== "picked" ? (javAscending ? " ↑" : " ↓") : "" }}
            </button>
          </div>
          <span class="next02-sc-result-count">{{ preview.jav.picked }} 部 · {{ preview.jav.totalSizeGB }} GB</span>
        </div>
      </div>
      <section class="smart-pick-cards-section" aria-label="JAV cards">
        <MovieGrid :cards="sortedJAV" />
      </section>
    </section>

    <section v-if="preview.fc2.movies.length" class="smart-pick-result-section is-fc2">
      <div class="smart-pick-result-section-head">
        <div>
          <span>FC2</span>
          <h5>FC2 结果</h5>
        </div>
        <div class="smart-pick-sort-tools">
          <div class="btn-group btn-group-sm flex-wrap" role="group" aria-label="FC2 结果排序">
            <button
              v-for="option in fc2SortOptions"
              :key="option.key"
              type="button"
              class="btn btn-outline-primary"
              :class="{ active: fc2Sort === option.key }"
              @click="selectFC2Sort(option.key)"
            >
              {{ option.label
              }}{{ fc2Sort === option.key && option.key !== "picked" ? (fc2Ascending ? " ↑" : " ↓") : "" }}
            </button>
          </div>
          <span class="next02-sc-result-count">{{ preview.fc2.picked }} 部 · {{ preview.fc2.totalSizeGB }} GB</span>
        </div>
      </div>
      <section class="smart-pick-cards-section" aria-label="FC2 cards">
        <FC2MovieGrid :cards="sortedFC2" />
      </section>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import FC2MovieGrid from "@/features/fc2/components/FC2MovieGrid.vue";
import MovieGrid from "@/features/movies/components/MovieGrid.vue";
import type { ScPickV2Response } from "@/features/sc/types";
import { num } from "@/shared/utils/format";

const props = defineProps<{ preview: ScPickV2Response; copyStarting: boolean; copyDisabled: boolean }>();
defineEmits<{ copy: [] }>();

const javSort = ref("release");
const javAscending = ref(false);
const fc2Sort = ref("picked");
const fc2Ascending = ref(true);
const javSortOptions = [
  { key: "picked", label: "抽取顺序" },
  { key: "download", label: "下载日期" },
  { key: "release", label: "上映日期" },
  { key: "sc", label: "SC 次数" },
  { key: "come", label: "Come 次数" }
];
const fc2SortOptions = [
  { key: "picked", label: "抽取顺序" },
  { key: "download", label: "下载日期" },
  { key: "post", label: "发布日期" },
  { key: "sc", label: "SC 次数" },
  { key: "come", label: "Come 次数" }
];

const sortedJAV = computed(() => {
  const rows = props.preview.jav.movies.slice();
  const direction = javAscending.value ? 1 : -1;
  return rows.sort((left, right) => {
    const order = left.picked_order - right.picked_order;
    if (javSort.value === "picked") return order;
    if (javSort.value === "download")
      return (dateValue(left.film_birth_date_w_media) - dateValue(right.film_birth_date_w_media)) * direction || order;
    if (javSort.value === "release")
      return (dateValue(left.releasing_date) - dateValue(right.releasing_date)) * direction || order;
    if (javSort.value === "sc") return (num(left, "sc_times") - num(right, "sc_times")) * direction || order;
    return (num(left, "come_times") - num(right, "come_times")) * direction || order;
  });
});

const sortedFC2 = computed(() => {
  const rows = props.preview.fc2.movies.slice();
  const direction = fc2Ascending.value ? 1 : -1;
  return rows.sort((left, right) => {
    const order = num(left, "picked_order") - num(right, "picked_order");
    if (fc2Sort.value === "picked") return order;
    if (fc2Sort.value === "download")
      return (num(left, "DownloadTime") - num(right, "DownloadTime")) * direction || order;
    if (fc2Sort.value === "post") return (num(left, "PostTime") - num(right, "PostTime")) * direction || order;
    if (fc2Sort.value === "sc") return (num(left, "ScTimes") - num(right, "ScTimes")) * direction || order;
    return (num(left, "ComeTimes") - num(right, "ComeTimes")) * direction || order;
  });
});

function selectJAVSort(key: string) {
  if (javSort.value === key && key !== "picked") javAscending.value = !javAscending.value;
  else {
    javSort.value = key;
    javAscending.value = key === "picked";
  }
}

function selectFC2Sort(key: string) {
  if (fc2Sort.value === key && key !== "picked") fc2Ascending.value = !fc2Ascending.value;
  else {
    fc2Sort.value = key;
    fc2Ascending.value = key === "picked";
  }
}

function dateValue(value: string): number {
  return Date.parse(value) || 0;
}
</script>
