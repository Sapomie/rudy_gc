<template>
  <PageShell
    class="movie-agg-page-shell sc-events-page-shell fetch-sites-page-shell"
    title="FetchSites 配置"
    eyebrow=""
  >
    <div
      id="page-root"
      class="container-fluid movie-agg-page-root sc-events-page-root fetch-sites-page-root next02-sc-pink-theme next02-sc-events-rewrite next02-fetch-sites-rewrite"
    >
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">FETCH SITES / CONFIG</span>
            <h4>FetchSites 配置</h4>
            <p>维护抓取站点配置，保存后当前 API 进程立即使用新配置。</p>
          </div>
          <div class="next02-sc-hero-actions">
            <div class="btn-group" role="group" aria-label="FetchSites 视图">
              <span class="btn active" aria-current="page">配置列表</span>
            </div>
          </div>
        </section>

        <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
          <div class="card-body">
            <div class="next02-sc-panel-head next02-fetch-sites-toolbar">
              <div class="next02-fetch-sites-toolbar-title">抓取站点配置</div>
              <div
                v-if="items.length"
                class="next02-sc-hero-metrics sc-event-list-metrics"
                aria-label="FetchSites 配置摘要"
              >
                <span
                  ><strong>{{ items.length }}</strong> Sites</span
                >
                <span
                  ><strong>{{ enabledCount }}</strong> Enabled</span
                >
                <span
                  ><strong>{{ persistedCount }}</strong> Persisted</span
                >
              </div>
            </div>
            <div class="table-responsive">
              <table class="table table-hover align-middle mb-0 sc-events-table fetch-sites-table">
                <colgroup>
                  <col class="fetch-sites-col-code" />
                  <col class="fetch-sites-col-name" />
                  <col class="fetch-sites-col-url" />
                  <col class="fetch-sites-col-agent" />
                  <col class="fetch-sites-col-proxy" />
                  <col class="fetch-sites-col-number" />
                  <col class="fetch-sites-col-number" />
                  <col class="fetch-sites-col-number" />
                  <col class="fetch-sites-col-retry" />
                  <col class="fetch-sites-col-status" />
                  <col class="fetch-sites-col-source" />
                  <col class="fetch-sites-col-action" />
                </colgroup>
                <thead>
                  <tr>
                    <th>SiteCode</th>
                    <th>SiteName</th>
                    <th>BaseURL</th>
                    <th>UserAgent</th>
                    <th>Proxy</th>
                    <th>Timeout</th>
                    <th>ReqSleep</th>
                    <th>RetrySleep</th>
                    <th>MaxRetry</th>
                    <th>Status</th>
                    <th>来源</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in items" :key="siteCode(item)" class="next02-sc-event-row">
                    <td class="mono" :title="siteCode(item)">{{ siteCode(item) }}</td>
                    <td :title="text(item, 'site_name')">{{ text(item, "site_name") || "-" }}</td>
                    <td class="mono" :title="text(item, 'base_url')">{{ text(item, "base_url") || "-" }}</td>
                    <td :title="text(item, 'user_agent')">{{ text(item, "user_agent") || "-" }}</td>
                    <td class="mono" :title="text(item, 'proxy')">{{ text(item, "proxy") || "-" }}</td>
                    <td>{{ text(item, "timeout_seconds") }}</td>
                    <td>{{ text(item, "request_sleep_ms") }}</td>
                    <td>{{ text(item, "retry_sleep_ms") }}</td>
                    <td>{{ text(item, "max_retry_times") }}</td>
                    <td>
                      <span :class="statusClass(item)">{{ statusText(item) }}</span>
                    </td>
                    <td>{{ pick(item, "persisted") ? "已入库" : "默认值" }}</td>
                    <td>
                      <button
                        type="button"
                        class="sc-event-link js-edit-fetch-site"
                        :data-site-code="siteCode(item)"
                        @click="openEdit(item)"
                      >
                        编辑
                      </button>
                    </td>
                  </tr>
                  <tr v-if="!items.length">
                    <td colspan="12"><div class="sc-empty-state">还没有可展示的抓取站点配置。</div></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div
          class="modal fade"
          id="fetchSiteEditModal"
          tabindex="-1"
          :aria-hidden="modalOpen ? undefined : 'true'"
          :aria-modal="modalOpen ? 'true' : undefined"
          :class="{ show: modalOpen }"
          role="dialog"
          @click.self="modalOpen = false"
        >
          <div class="modal-dialog modal-lg modal-dialog-centered">
            <div class="modal-content">
              <form id="fetchSiteEditForm" @submit.prevent="save">
                <div class="modal-header">
                  <div>
                    <span class="next02-fetch-sites-modal-eyebrow">FETCH SITE CONFIG</span>
                    <h5 class="modal-title">编辑 {{ text(edit, "site_name") || "FetchSite" }}</h5>
                  </div>
                  <button
                    type="button"
                    class="next02-fetch-sites-modal-close"
                    aria-label="关闭"
                    @click="modalOpen = false"
                  >
                    ×
                  </button>
                </div>
                <div class="modal-body">
                  <input type="hidden" id="edit_site_code" :value="text(edit, 'site_code')" />
                  <div class="row g-3">
                    <div class="col-md-4">
                      <label class="form-label" for="edit_site_code_text">SiteCode</label>
                      <input
                        type="text"
                        class="form-control"
                        id="edit_site_code_text"
                        :value="text(edit, 'site_code')"
                        disabled
                      />
                    </div>
                    <div class="col-md-4">
                      <label class="form-label" for="edit_site_name">SiteName</label>
                      <input type="text" class="form-control" id="edit_site_name" v-model="edit.site_name" required />
                    </div>
                    <div class="col-md-4">
                      <label class="form-label" for="edit_status">Status</label>
                      <select class="form-select" id="edit_status" v-model.number="edit.status">
                        <option :value="1">启用</option>
                        <option :value="0">停用</option>
                      </select>
                    </div>
                    <div class="col-12">
                      <label class="form-label" for="edit_base_url">BaseURL</label>
                      <input type="url" class="form-control mono" id="edit_base_url" v-model="edit.base_url" required />
                    </div>
                    <div class="col-12">
                      <label class="form-label" for="edit_user_agent">UserAgent</label>
                      <textarea class="form-control" id="edit_user_agent" v-model="edit.user_agent" rows="2"></textarea>
                    </div>
                    <div class="col-12">
                      <label class="form-label" for="edit_cookie">Cookie</label>
                      <textarea class="form-control mono" id="edit_cookie" v-model="edit.cookie" rows="3"></textarea>
                    </div>
                    <div class="col-12">
                      <label class="form-label" for="edit_proxy">Proxy</label>
                      <input type="text" class="form-control mono" id="edit_proxy" v-model="edit.proxy" />
                    </div>
                    <div class="col-md-3">
                      <label class="form-label" for="edit_timeout_seconds">TimeoutSeconds</label>
                      <input
                        type="number"
                        class="form-control"
                        id="edit_timeout_seconds"
                        v-model.number="edit.timeout_seconds"
                        min="1"
                        required
                      />
                    </div>
                    <div class="col-md-3">
                      <label class="form-label" for="edit_request_sleep_ms">RequestSleepMs</label>
                      <input
                        type="number"
                        class="form-control"
                        id="edit_request_sleep_ms"
                        v-model.number="edit.request_sleep_ms"
                        min="0"
                        required
                      />
                    </div>
                    <div class="col-md-3">
                      <label class="form-label" for="edit_retry_sleep_ms">RetrySleepMs</label>
                      <input
                        type="number"
                        class="form-control"
                        id="edit_retry_sleep_ms"
                        v-model.number="edit.retry_sleep_ms"
                        min="0"
                        required
                      />
                    </div>
                    <div class="col-md-3">
                      <label class="form-label" for="edit_max_retry_times">MaxRetryTimes</label>
                      <input
                        type="number"
                        class="form-control"
                        id="edit_max_retry_times"
                        v-model.number="edit.max_retry_times"
                        min="1"
                        required
                      />
                    </div>
                  </div>
                  <div
                    id="fetchSiteEditMsg"
                    class="small mt-3"
                    :class="messageClass"
                    :style="{ display: message ? 'block' : 'none' }"
                  >
                    {{ message }}
                  </div>
                </div>
                <div class="modal-footer">
                  <button type="button" class="btn btn-light" @click="modalOpen = false">取消</button>
                  <button type="submit" class="btn btn-primary">保存</button>
                </div>
              </form>
            </div>
          </div>
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import "@/features/crawler/styles/index.css";
import { computed, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/features/crawler/api";
import PageShell from "@/shared/components/PageShell.vue";
import type { CrawlerRecord } from "@/features/crawler/types";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<CrawlerRecord>({});
const loading = ref(false);
const error = ref("");
const message = ref("");
const messageOk = ref(true);
const modalOpen = ref(false);
const edit = reactive<CrawlerRecord>({});

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const items = computed(() => asArray<CrawlerRecord>(pick(payload.value, "items", "Items")));
const messageClass = computed(() => (messageOk.value ? "text-success" : "text-danger"));
const enabledCount = computed(() => items.value.filter((item) => Number(pick(item, "status") || 0) === 1).length);
const persistedCount = computed(() => items.value.filter((item) => Boolean(pick(item, "persisted"))).length);

function siteCode(item: CrawlerRecord): string {
  return text(item, "site_code");
}

function statusText(item: CrawlerRecord): string {
  return Number(pick(item, "status") || 0) === 1 ? "启用" : "停用";
}

function statusClass(item: CrawlerRecord): string {
  return Number(pick(item, "status") || 0) === 1 ? "fetch-sites-status-enabled" : "sc-muted";
}

function openEdit(item: CrawlerRecord) {
  Object.keys(edit).forEach((key) => delete edit[key]);
  Object.assign(edit, item);
  message.value = "";
  modalOpen.value = true;
}

async function save() {
  const code = text(edit, "site_code");
  if (!code) {
    message.value = "缺少 site_code";
    messageOk.value = false;
    return;
  }
  try {
    await sendJSON(`/api/fetch-sites/${encodeURIComponent(code)}`, "PUT", edit);
    message.value = "保存成功";
    messageOk.value = true;
    modalOpen.value = false;
    await load();
  } catch (err) {
    message.value = err instanceof Error ? err.message : String(err);
    messageOk.value = false;
  }
}
</script>
