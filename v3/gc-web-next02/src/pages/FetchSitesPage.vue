<template>
  <PageShell class="fetch-sites-page-shell" title="FetchSites 配置" eyebrow="">
    <div id="fetch-sites-root" class="fetch-sites-page-root">
      <LoadingState v-if="loading" />
      <ErrorPanel v-else-if="error" :message="error" />
      <template v-else>
        <div class="row">
          <div class="col-12">
            <div class="page-title-box d-flex align-items-center justify-content-between">
              <div>
                <h4 class="mb-0">FetchSites 配置</h4>
                <div class="text-muted small mt-1">维护 t_fetch_site 的抓取站点配置；保存后当前 API 进程立即使用新配置。</div>
              </div>
              <div>
                <a class="btn btn-sm btn-outline-primary" href="/triggers/fetch-sehuatang">FetchSehuatang</a>
              </div>
            </div>
          </div>
        </div>

        <div class="card shadow-sm">
          <div class="card-body">
            <div class="table-responsive">
              <table class="table table-bordered table-hover fetch-sites-table mb-0">
                <thead class="table-light">
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
                  <tr v-if="!items.length">
                    <td colspan="12" class="text-center text-muted py-4">暂无数据</td>
                  </tr>
                  <tr v-for="item in items" v-else :key="siteCode(item)">
                    <td class="mono">{{ siteCode(item) }}</td>
                    <td>{{ text(item, "site_name") }}</td>
                    <td class="wide mono">{{ text(item, "base_url") }}</td>
                    <td class="wide">{{ text(item, "user_agent") || "-" }}</td>
                    <td class="wide mono">{{ text(item, "proxy") || "-" }}</td>
                    <td>{{ text(item, "timeout_seconds") }}</td>
                    <td>{{ text(item, "request_sleep_ms") }}</td>
                    <td>{{ text(item, "retry_sleep_ms") }}</td>
                    <td>{{ text(item, "max_retry_times") }}</td>
                    <td><span class="badge" :class="statusClass(item)">{{ statusText(item) }}</span></td>
                    <td>{{ pick(item, "persisted") ? "已入库" : "默认值" }}</td>
                    <td class="text-nowrap">
                      <button type="button" class="btn btn-sm btn-outline-primary js-edit-fetch-site" :data-site-code="siteCode(item)" @click="openEdit(item)">
                        <span class="ms-1">编辑</span>
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div class="modal fade" id="fetchSiteEditModal" tabindex="-1" aria-hidden="true" :class="{ show: modalOpen }">
          <div class="modal-dialog modal-lg modal-dialog-centered">
            <div class="modal-content">
              <form id="fetchSiteEditForm" @submit.prevent="save">
                <div class="modal-header">
                  <h5 class="modal-title">编辑 FetchSite</h5>
                  <button type="button" class="btn-close" aria-label="Close" @click="modalOpen = false"></button>
                </div>
                <div class="modal-body">
                  <input type="hidden" id="edit_site_code" :value="text(edit, 'site_code')">
                  <div class="row g-3">
                    <div class="col-md-4">
                      <label class="form-label" for="edit_site_code_text">SiteCode</label>
                      <input type="text" class="form-control" id="edit_site_code_text" :value="text(edit, 'site_code')" disabled>
                    </div>
                    <div class="col-md-4">
                      <label class="form-label" for="edit_site_name">SiteName</label>
                      <input type="text" class="form-control" id="edit_site_name" v-model="edit.site_name" required>
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
                      <input type="url" class="form-control mono" id="edit_base_url" v-model="edit.base_url" required>
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
                      <input type="text" class="form-control mono" id="edit_proxy" v-model="edit.proxy">
                    </div>
                    <div class="col-md-3">
                      <label class="form-label" for="edit_timeout_seconds">TimeoutSeconds</label>
                      <input type="number" class="form-control" id="edit_timeout_seconds" v-model.number="edit.timeout_seconds" min="1" required>
                    </div>
                    <div class="col-md-3">
                      <label class="form-label" for="edit_request_sleep_ms">RequestSleepMs</label>
                      <input type="number" class="form-control" id="edit_request_sleep_ms" v-model.number="edit.request_sleep_ms" min="0" required>
                    </div>
                    <div class="col-md-3">
                      <label class="form-label" for="edit_retry_sleep_ms">RetrySleepMs</label>
                      <input type="number" class="form-control" id="edit_retry_sleep_ms" v-model.number="edit.retry_sleep_ms" min="0" required>
                    </div>
                    <div class="col-md-3">
                      <label class="form-label" for="edit_max_retry_times">MaxRetryTimes</label>
                      <input type="number" class="form-control" id="edit_max_retry_times" v-model.number="edit.max_retry_times" min="1" required>
                    </div>
                  </div>
                  <div id="fetchSiteEditMsg" class="small mt-3" :class="messageClass" :style="{ display: message ? 'block' : 'none' }">{{ message }}</div>
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
import { computed, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LoadingState from "@/components/LoadingState.vue";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, pick, text } from "@/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const message = ref("");
const messageOk = ref(true);
const modalOpen = ref(false);
const edit = reactive<AnyRecord>({});

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

const items = computed(() => asArray<AnyRecord>(pick(payload.value, "items", "Items")));
const messageClass = computed(() => messageOk.value ? "text-success" : "text-danger");

function siteCode(item: AnyRecord): string {
  return text(item, "site_code");
}

function statusText(item: AnyRecord): string {
  return Number(pick(item, "status") || 0) === 1 ? "启用" : "停用";
}

function statusClass(item: AnyRecord): string {
  return Number(pick(item, "status") || 0) === 1 ? "bg-success" : "bg-secondary";
}

function openEdit(item: AnyRecord) {
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
