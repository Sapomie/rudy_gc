<template>
  <div class="row next02-sh-detail-preview-row">
    <div class="col-12">
      <div class="next02-sh-detail-preview-standard next02-sc-events-rewrite">
        <div
          id="fetchSitePreviewPanel"
          class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card"
          :class="{ 'is-hidden': !open, 'next02-pink-trigger-panel next02-pink-trigger-preview-panel': pink }"
        >
          <div class="card-body">
            <div class="next02-sc-panel-head next02-sh-detail-preview-toolbar">
              <div>
                <span class="next02-sc-panel-kicker">FILTER RESULT</span>
                <h5 class="mb-0">筛选结果预览</h5>
                <div id="fetchSitePreviewSummary" class="next02-sh-detail-preview-summary">{{ summary }}</div>
              </div>
              <div class="next02-sh-detail-preview-controls">
                <button
                  id="hideFetchSitePreviewBtn"
                  type="button"
                  class="btn btn-outline-secondary btn-sm"
                  @click="$emit('hide')"
                >
                  收起
                </button>
              </div>
            </div>
            <div v-if="error" class="alert alert-danger next02-sh-detail-preview-error" role="alert">{{ error }}</div>
            <div v-if="loading" class="sc-empty-state">正在生成目标清单...</div>
            <div v-else class="next02-sh-detail-preview-groups" aria-label="站点抓取筛选聚合数据">
              <section v-for="group in groups" :key="group.label" class="next02-sh-detail-preview-group">
                <div class="next02-sh-detail-preview-group-label">{{ group.label }}</div>
                <div class="next02-sh-detail-preview-group-items">
                  <div
                    v-for="item in group.items"
                    :key="item.label"
                    class="next02-sh-detail-preview-metric"
                    :class="item.className"
                  >
                    <span>{{ item.label }}</span>
                    <strong>{{ item.value }}</strong>
                  </div>
                </div>
              </section>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TriggerMetricGroup } from "./models";

defineProps<{
  open: boolean;
  loading: boolean;
  pink: boolean;
  error: string;
  summary: string;
  groups: TriggerMetricGroup[];
}>();

defineEmits<{ hide: [] }>();
</script>
