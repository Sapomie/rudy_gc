<template>
  <PageShell eyebrow="COLLECTOR" title="手动抓取" note="创建一次当前月榜抓取任务，任务会在后台完整抓取所有页面。">
    <section class="panel fetch-form-panel">
      <header class="panel-head">
        <h2>任务参数</h2>
        <StatusBadge label="月榜" tone="neutral" />
      </header>
      <form class="panel-body" @submit.prevent="submitFetch">
        <div class="form-grid">
          <div class="form-field">
            <label for="business-date">业务日期</label>
            <input id="business-date" v-model="businessDate" class="form-control" type="date" />
            <p class="form-hint">留空时由服务按当前业务日期计算；源站只接受当前榜单日期。</p>
          </div>
          <div class="form-field">
            <label for="category">榜单类型</label>
            <select id="category" class="form-select" disabled>
              <option>月榜（category=1）</option>
            </select>
            <p class="form-hint">当前远端服务只采集月榜。</p>
          </div>
        </div>
        <div class="form-actions">
          <button class="button-primary" type="submit" :disabled="submitting">
            {{ submitting ? "正在创建…" : "创建抓取任务" }}
          </button>
          <RouterLink class="button-secondary" to="/dailybestfetchrunlist">查看任务记录</RouterLink>
        </div>
      </form>
    </section>

    <ErrorPanel v-if="error" :message="error" />
    <section v-if="accepted" class="panel accepted-panel">
      <header class="panel-head">
        <h2>任务创建成功</h2>
        <StatusBadge :label="formatRunStatus(accepted.status)" :tone="statusTone(accepted.status)" />
      </header>
      <div class="key-value-list">
        <div>
          <span>任务 ID</span><strong>{{ accepted.run_id }}</strong>
        </div>
        <div>
          <span>返回信息</span><strong>{{ accepted.message }}</strong>
        </div>
        <RouterLink class="button-primary" to="/dailybestfetchrunlist">查看任务进度</RouterLink>
      </div>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { startFetch } from "@/features/dailybest/api";
import StatusBadge from "@/features/dailybest/components/StatusBadge.vue";
import type { FetchAcceptedResponse } from "@/features/dailybest/types";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import PageShell from "@/shared/components/PageShell.vue";
import { errorMessage, formatRunStatus, statusTone } from "@/shared/utils/format";

const businessDate = ref("");
const submitting = ref(false);
const error = ref("");
const accepted = ref<FetchAcceptedResponse | null>(null);

async function submitFetch(): Promise<void> {
  submitting.value = true;
  error.value = "";
  accepted.value = null;
  try {
    accepted.value = await startFetch({ business_date: businessDate.value, category: 1 });
  } catch (reason) {
    error.value = errorMessage(reason);
  } finally {
    submitting.value = false;
  }
}
</script>
