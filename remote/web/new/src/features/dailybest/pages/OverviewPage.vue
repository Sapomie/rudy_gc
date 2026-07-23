<template>
  <PageShell eyebrow="REMOTE SERVICE" title="运行概览" note="查看远端采集服务、排名数据和抓取任务的当前状态。">
    <template #actions
      ><button class="button-secondary" type="button" :disabled="loading" @click="loadOverview">刷新</button></template
    >
    <section class="overview-grid">
      <article class="stat-card">
        <span class="stat-label">服务状态</span><strong>{{ healthLabel }}</strong
        ><StatusBadge
          :label="healthLabel"
          :tone="health === 'ready' ? 'good' : health === 'loading' ? 'info' : 'danger'"
        />
      </article>
      <article class="stat-card">
        <span class="stat-label">最新排行日期</span><strong>{{ latestDay?.rank_date ?? "—" }}</strong
        ><span class="stat-note">日序号 {{ latestDay?.day_number ?? "—" }}</span>
      </article>
      <article class="stat-card">
        <span class="stat-label">最新排行规模</span><strong>{{ latestDay?.rank_row_count ?? "—" }}</strong
        ><span class="stat-note">{{ latestDay ? `${latestDay.movie_count} 部影片` : "等待排行" }}</span>
      </article>
      <article class="stat-card">
        <span class="stat-label">最新抓取任务</span
        ><strong>{{ latestRun ? formatRunStatus(latestRun.status) : "—" }}</strong
        ><span class="stat-note">{{ latestRun ? formatDateTime(latestRun.updated_at) : "等待任务" }}</span>
      </article>
    </section>
    <LoadingState v-if="loading" :lines="3" />
    <div v-else class="overview-columns">
      <section class="panel">
        <header class="panel-head">
          <h2>最新每日统计</h2>
          <RouterLink class="button-secondary" to="/rankdaystats">查看全部</RouterLink>
        </header>
        <ErrorPanel v-if="rankError" :message="rankError" />
        <EmptyState v-else-if="!queried" title="尚未查询" message="点击刷新读取最新每日统计。" />
        <div v-else-if="latestDay" class="key-value-list">
          <div>
            <span>排行日期</span><strong>{{ latestDay.rank_date }}</strong>
          </div>
          <div>
            <span>日序号</span><strong>{{ latestDay.day_number }}</strong>
          </div>
          <div>
            <span>排行行数</span><strong>{{ latestDay.rank_row_count }}</strong>
          </div>
          <div>
            <span>影片数</span><strong>{{ latestDay.movie_count }}</strong>
          </div>
          <div>
            <span>首次上榜</span><strong>{{ latestDay.first_seen_movie_count }}</strong>
          </div>
          <div>
            <span>连续上榜</span><strong>{{ latestDay.continued_movie_count }}</strong>
          </div>
          <div class="wide-row">
            <span>Checksum</span><strong class="mono-value">{{ latestDay.source_checksum }}</strong>
          </div>
          <RouterLink class="button-primary" :to="`/ranklist?day_number=${latestDay.day_number}`"
            >查看当日排名</RouterLink
          >
        </div>
        <EmptyState v-else title="暂无排名" message="完成首次抓取后会生成排名及统计。" />
      </section>
      <section class="panel">
        <header class="panel-head">
          <h2>最新抓取任务</h2>
          <RouterLink class="button-secondary" to="/dailybestfetchrunlist">查看全部</RouterLink>
        </header>
        <ErrorPanel v-if="runError" :message="runError" />
        <EmptyState v-else-if="!queried" title="尚未查询" message="点击刷新读取最新抓取任务。" />
        <div v-else-if="latestRun" class="key-value-list">
          <div>
            <span>任务 ID</span
            ><RouterLink :to="`/dailybestfetchrundetail/${latestRun.id}`">#{{ latestRun.id }}</RouterLink>
          </div>
          <div>
            <span>状态</span
            ><StatusBadge :label="formatRunStatus(latestRun.status)" :tone="statusTone(latestRun.status)" />
          </div>
          <div>
            <span>触发方式</span><strong>{{ formatTriggerType(latestRun.trigger_type) }}</strong>
          </div>
          <div>
            <span>业务日期</span><strong>{{ latestRun.business_date }}</strong>
          </div>
          <div>
            <span>数据规模</span><strong>{{ latestRun.page_count }} 页 / {{ latestRun.item_count }} 条</strong>
          </div>
          <div>
            <span>更新时间</span><strong>{{ formatDateTime(latestRun.updated_at) }}</strong>
          </div>
          <div class="wide-row">
            <span>Checksum</span><strong class="mono-value">{{ latestRun.checksum || "—" }}</strong>
          </div>
          <div v-if="latestRun.error_message" class="wide-row error-value">
            <span>错误</span><strong>{{ latestRun.error_message }}</strong>
          </div>
          <RouterLink class="button-primary" :to="`/dailybestfetchrundetail/${latestRun.id}`">查看任务详情</RouterLink>
        </div>
        <EmptyState v-else title="暂无抓取任务" message="可以从手动抓取页面创建第一条任务。"
          ><RouterLink class="button-primary" to="/dailybestfetch">新建抓取</RouterLink></EmptyState
        >
      </section>
    </div>
  </PageShell>
</template>
<script setup lang="ts">
import { computed, ref } from "vue";
import { listFetchRuns, listRankDayStats, loadHealthState } from "@/features/dailybest/api";
import StatusBadge from "@/features/dailybest/components/StatusBadge.vue";
import type { FetchRun, HealthState, RankDayStat } from "@/features/dailybest/types";
import EmptyState from "@/shared/components/EmptyState.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import { errorMessage, formatDateTime, formatRunStatus, formatTriggerType, statusTone } from "@/shared/utils/format";
const health = ref<HealthState>("loading"),
  latestDay = ref<RankDayStat | null>(null),
  latestRun = ref<FetchRun | null>(null),
  rankError = ref(""),
  runError = ref(""),
  queried = ref(false),
  loading = ref(false);
const healthLabel = computed(() =>
  queried.value ? { loading: "检查中", ready: "运行正常", unavailable: "服务不可用" }[health.value] : "尚未查询"
);
async function loadOverview() {
  queried.value = true;
  loading.value = true;
  rankError.value = "";
  runError.value = "";
  await Promise.all([
    loadHealthState()
      .then((v) => (health.value = v))
      .catch(() => (health.value = "unavailable")),
    listRankDayStats({ p: 1, ps: 1 })
      .then((r) => (latestDay.value = r.stats[0] || null))
      .catch((e) => (rankError.value = errorMessage(e))),
    listFetchRuns({ p: 1, ps: 1 })
      .then((r) => (latestRun.value = r.runs[0] || null))
      .catch((e) => (runError.value = errorMessage(e)))
  ]);
  loading.value = false;
}
</script>
