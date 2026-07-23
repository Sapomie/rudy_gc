<template>
  <PageShell eyebrow="FETCH RUN" :title="`抓取任务 #${runIDLabel}`" note="查看一次抓取任务从创建到完成的完整执行信息。">
    <template #actions>
      <RouterLink class="button-secondary" to="/dailybestfetchrunlist">返回任务列表</RouterLink>
      <RouterLink class="button-secondary" :to="`/dailybestfetchloglist?fetch_run_id=${runID}`"
        >查看请求日志</RouterLink
      >
      <button class="button-secondary" type="button" :disabled="loading" @click="loadRun">刷新</button>
    </template>

    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error">
      <button class="button-secondary" type="button" @click="loadRun">重新加载</button>
    </ErrorPanel>
    <EmptyState v-else-if="!queried" title="尚未查询" message="点击刷新读取该抓取任务。" />
    <template v-else-if="run">
      <section class="summary-strip" aria-label="任务摘要">
        <article>
          <span>当前状态</span>
          <StatusBadge :label="formatRunStatus(run.status)" :tone="statusTone(run.status)" />
        </article>
        <article>
          <span>业务日期</span><strong>{{ run.business_date }}</strong>
        </article>
        <article>
          <span>抓取规模</span><strong>{{ run.page_count }} 页 / {{ run.item_count }} 条</strong>
        </article>
        <article>
          <span>触发方式</span><strong>{{ formatTriggerType(run.trigger_type) }}</strong>
        </article>
      </section>

      <section class="panel">
        <header class="panel-head">
          <div>
            <span class="panel-kicker">D_DAILYBEST_FETCH_RUN</span>
            <h2>任务运行信息</h2>
          </div>
          <StatusBadge :label="formatRunStatus(run.status)" :tone="statusTone(run.status)" />
        </header>
        <div class="detail-grid">
          <div>
            <span>任务 ID</span><strong>{{ run.id }}</strong>
          </div>
          <div>
            <span>榜单类型</span><strong>{{ formatCategory(run.category) }}</strong>
          </div>
          <div>
            <span>开始时间</span><strong>{{ formatDateTime(run.started_at) }}</strong>
          </div>
          <div>
            <span>完成时间</span><strong>{{ formatDateTime(run.finished_at) }}</strong>
          </div>
          <div>
            <span>创建时间</span><strong>{{ formatDateTime(run.created_at) }}</strong>
          </div>
          <div>
            <span>更新时间</span><strong>{{ formatDateTime(run.updated_at) }}</strong>
          </div>
          <div>
            <span>页面数量</span><strong>{{ run.page_count }}</strong>
          </div>
          <div>
            <span>条目数量</span><strong>{{ run.item_count }}</strong>
          </div>
          <div class="wide-row">
            <span>Checksum</span><strong class="mono-value">{{ run.checksum || "—" }}</strong>
          </div>
          <div class="wide-row">
            <span>错误信息</span>
            <strong :class="{ 'error-value': run.error_message }">{{ run.error_message || "—" }}</strong>
          </div>
        </div>
      </section>
    </template>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { getFetchRun } from "@/features/dailybest/api";
import StatusBadge from "@/features/dailybest/components/StatusBadge.vue";
import type { FetchRun } from "@/features/dailybest/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import {
  errorMessage,
  formatCategory,
  formatDateTime,
  formatRunStatus,
  formatTriggerType,
  statusTone
} from "@/shared/utils/format";

const route = useRoute();
const run = ref<FetchRun | null>(null);
const queried = ref(false);
const loading = ref(false);
const error = ref("");
const runID = computed(() => Number(route.params.id));
const runIDLabel = computed(() => (Number.isInteger(runID.value) && runID.value > 0 ? runID.value : "无效"));

async function loadRun(): Promise<void> {
  queried.value = true;
  if (!Number.isInteger(runID.value) || runID.value <= 0) {
    error.value = "任务 ID 必须为正整数";
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    run.value = await getFetchRun(runID.value);
  } catch (reason) {
    error.value = errorMessage(reason);
  } finally {
    loading.value = false;
  }
}

watch(
  () => route.params.id,
  () => {
    queried.value = false;
    run.value = null;
    error.value = "";
  }
);
</script>
