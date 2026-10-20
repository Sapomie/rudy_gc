<template>
  <div class="row next02-sh-detail-preview-row">
    <div class="col-12">
      <div class="next02-sh-detail-preview-standard next02-sc-events-rewrite">
        <div
          id="shDetailPreviewPanel"
          class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card"
          :class="{ 'is-hidden': !open }"
        >
          <div class="card-body">
            <div class="next02-sc-panel-head next02-sh-detail-preview-toolbar">
              <div>
                <span class="next02-sc-panel-kicker">FILTER SUMMARY</span>
                <h5 class="mb-0">筛选结果预览</h5>
                <div id="shDetailPreviewSummary" class="next02-sh-detail-preview-summary">{{ summary }}</div>
              </div>
              <div class="next02-sh-detail-preview-controls">
                <div class="next02-sh-detail-preview-actions">
                  <a
                    v-if="poolHref"
                    class="btn btn-outline-secondary btn-sm"
                    :href="poolHref"
                    target="_blank"
                    rel="noopener noreferrer"
                    >在帖子池查看</a
                  >
                  <button
                    id="hideSHDetailPreviewBtn"
                    type="button"
                    class="btn btn-outline-secondary btn-sm"
                    @click="$emit('collapse')"
                  >
                    收起
                  </button>
                </div>
              </div>
            </div>
            <div v-if="error" class="alert alert-danger next02-sh-detail-preview-error" role="alert">{{ error }}</div>
            <div v-if="loading" class="sc-empty-state">正在统计筛选结果...</div>
            <div v-else class="next02-sh-detail-preview-groups" aria-label="帖子筛选聚合数据">
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
  error: string;
  summary: string;
  poolHref: string;
  groups: TriggerMetricGroup[];
}>();

defineEmits<{ collapse: [] }>();
</script>
