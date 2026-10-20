<template>
  <PageShell
    class="remote-dailybest-page remote-list-page"
    eyebrow="COLLECTOR"
    title="站点配置"
    note="维护远端 DailyBest 使用的 JAV 网站域名，保存后下一轮抓取和探测使用新配置。"
  >
    <template #actions>
      <button class="button-secondary" type="button" :disabled="loading || saving" @click="loadConfig">刷新</button>
    </template>
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="loadError" :message="loadError">
      <button class="button-secondary" type="button" @click="loadConfig">重新加载</button>
    </ErrorPanel>
    <template v-else-if="config">
      <section class="panel fetch-form-panel">
        <header class="panel-head">
          <h2>JAV 网站</h2>
          <StatusBadge :label="config.jav_address ? '已配置' : '待配置'" :tone="config.jav_address ? 'good' : 'warn'" />
        </header>
        <form class="panel-body" @submit.prevent="saveConfig">
          <label class="remote-list-filter-field source-config-field">
            <span>JAV 域名</span>
            <input
              v-model="address"
              type="text"
              autocomplete="off"
              spellcheck="false"
              required
              :disabled="saving"
              aria-describedby="jav-address-hint"
            />
          </label>
          <p id="jav-address-hint" class="form-hint">
            填写域名，省略协议、端口和路径。同一轮分页抓取保持使用开始时的域名。
          </p>
          <div class="form-actions">
            <button class="button-primary" type="submit" :disabled="saving || !address.trim()">
              {{ saving ? "正在保存…" : "保存配置" }}
            </button>
            <span v-if="saved" role="status"><StatusBadge label="保存成功" tone="good" /></span>
          </div>
        </form>
      </section>
      <ErrorPanel v-if="saveError" :message="saveError" />
      <section class="panel fetch-form-panel">
        <header class="panel-head"><h2>当前配置</h2></header>
        <div class="key-value-list">
          <div>
            <span>当前域名</span><strong>{{ config.jav_address || "待配置" }}</strong>
          </div>
          <div>
            <span>更新时间</span><strong>{{ formatDateTime(config.updated_on) }}</strong>
          </div>
          <div class="wide-row">
            <span>目标站探测地址</span
            ><strong class="mono-value wrap-cell">{{ config.target_probe_url || "保存域名后生成" }}</strong>
          </div>
        </div>
      </section>
    </template>
  </PageShell>
</template>

<script setup lang="ts">
import StatusBadge from "@/features/remote-dailybest/components/StatusBadge.vue";
import { useSourceConfig } from "@/features/remote-dailybest/useSourceConfig";
import ErrorPanel from "@/features/remote-dailybest/components/RemoteErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import { formatDateTime } from "@/features/remote-dailybest/format";

const { config, address, loading, saving, loadError, saveError, saved, loadConfig, saveConfig } = useSourceConfig();
void loadConfig();
</script>
