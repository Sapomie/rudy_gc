const root = document.getElementById("fetch-sites-root");

let fetchSiteItems = [];

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function attr(value) {
  return escapeHtml(value);
}

function request(url, options) {
  return fetch(url, options).then(async (response) => {
    const data = await response.json().catch(() => ({}));
    if (!response.ok || data.ok === false) {
      throw new Error(data.error || `请求失败(${response.status})`);
    }
    return data;
  });
}

function showMsg(message, ok = true) {
  const el = document.getElementById("fetchSiteMsg");
  if (!el) return;
  el.textContent = message;
  el.className = `alert ${ok ? "alert-success" : "alert-danger"} fade show small py-2 px-3`;
  el.style.display = "block";
  window.clearTimeout(showMsg.timer);
  showMsg.timer = window.setTimeout(() => {
    el.style.display = "none";
  }, 2600);
}

function formatUnix(value) {
  const n = Number(value || 0);
  if (!Number.isFinite(n) || n <= 0) return "-";
  return new Date(n * 1000).toLocaleString();
}

function statusBadge(value) {
  const enabled = Number(value || 0) === 1;
  return `<span class="badge ${enabled ? "bg-success" : "bg-secondary"}">${enabled ? "启用" : "停用"}</span>`;
}

function renderRows(items) {
  if (!Array.isArray(items) || items.length === 0) {
    return `<tr><td colspan="12" class="text-center text-muted py-4">暂无数据</td></tr>`;
  }
  return items.map((item) => `
    <tr>
      <td class="mono">${escapeHtml(item.site_code)}</td>
      <td>${escapeHtml(item.site_name)}</td>
      <td class="wide mono">${escapeHtml(item.base_url)}</td>
      <td class="wide">${escapeHtml(item.user_agent || "-")}</td>
      <td class="wide mono">${escapeHtml(item.proxy || "-")}</td>
      <td>${escapeHtml(item.timeout_seconds)}</td>
      <td>${escapeHtml(item.request_sleep_ms)}</td>
      <td>${escapeHtml(item.retry_sleep_ms)}</td>
      <td>${escapeHtml(item.max_retry_times)}</td>
      <td>${statusBadge(item.status)}</td>
      <td>${item.persisted ? "已入库" : "默认值"}</td>
      <td class="text-nowrap">
        <button type="button" class="btn btn-sm btn-outline-primary js-edit-fetch-site" data-site-code="${attr(item.site_code)}">
          <i class="mdi mdi-pencil"></i>
          <span class="ms-1">编辑</span>
        </button>
      </td>
    </tr>
  `).join("");
}

function renderPage(items) {
  root.innerHTML = `
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
            <tbody>${renderRows(items)}</tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="modal fade" id="fetchSiteEditModal" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-lg modal-dialog-centered">
        <div class="modal-content">
          <form id="fetchSiteEditForm">
            <div class="modal-header">
              <h5 class="modal-title">编辑 FetchSite</h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
            </div>
            <div class="modal-body">
              <input type="hidden" id="edit_site_code">
              <div class="row g-3">
                <div class="col-md-4">
                  <label class="form-label" for="edit_site_code_text">SiteCode</label>
                  <input type="text" class="form-control" id="edit_site_code_text" disabled>
                </div>
                <div class="col-md-4">
                  <label class="form-label" for="edit_site_name">SiteName</label>
                  <input type="text" class="form-control" id="edit_site_name" required>
                </div>
                <div class="col-md-4">
                  <label class="form-label" for="edit_status">Status</label>
                  <select class="form-select" id="edit_status">
                    <option value="1">启用</option>
                    <option value="0">停用</option>
                  </select>
                </div>
                <div class="col-12">
                  <label class="form-label" for="edit_base_url">BaseURL</label>
                  <input type="url" class="form-control mono" id="edit_base_url" required>
                </div>
                <div class="col-12">
                  <label class="form-label" for="edit_user_agent">UserAgent</label>
                  <textarea class="form-control" id="edit_user_agent" rows="2"></textarea>
                </div>
                <div class="col-12">
                  <label class="form-label" for="edit_cookie">Cookie</label>
                  <textarea class="form-control mono" id="edit_cookie" rows="3"></textarea>
                </div>
                <div class="col-12">
                  <label class="form-label" for="edit_proxy">Proxy</label>
                  <input type="text" class="form-control mono" id="edit_proxy">
                </div>
                <div class="col-md-3">
                  <label class="form-label" for="edit_timeout_seconds">TimeoutSeconds</label>
                  <input type="number" class="form-control" id="edit_timeout_seconds" min="1" required>
                </div>
                <div class="col-md-3">
                  <label class="form-label" for="edit_request_sleep_ms">RequestSleepMs</label>
                  <input type="number" class="form-control" id="edit_request_sleep_ms" min="0" required>
                </div>
                <div class="col-md-3">
                  <label class="form-label" for="edit_retry_sleep_ms">RetrySleepMs</label>
                  <input type="number" class="form-control" id="edit_retry_sleep_ms" min="0" required>
                </div>
                <div class="col-md-3">
                  <label class="form-label" for="edit_max_retry_times">MaxRetryTimes</label>
                  <input type="number" class="form-control" id="edit_max_retry_times" min="1" required>
                </div>
              </div>
              <div id="fetchSiteEditMsg" class="small mt-3" style="display:none;"></div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
              <button type="submit" class="btn btn-primary">保存</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  `;
}

function modalOf(id) {
  const el = document.getElementById(id);
  if (!el || !window.bootstrap || !window.bootstrap.Modal) return null;
  return window.bootstrap.Modal.getOrCreateInstance(el);
}

function findItem(siteCode) {
  return fetchSiteItems.find((item) => String(item.site_code || "") === siteCode);
}

function setValue(id, value) {
  const el = document.getElementById(id);
  if (el) el.value = value ?? "";
}

function openEditModal(item) {
  if (!item) return;
  setValue("edit_site_code", item.site_code);
  setValue("edit_site_code_text", item.site_code);
  setValue("edit_site_name", item.site_name);
  setValue("edit_status", String(item.status ?? 1));
  setValue("edit_base_url", item.base_url);
  setValue("edit_user_agent", item.user_agent);
  setValue("edit_cookie", item.cookie);
  setValue("edit_proxy", item.proxy);
  setValue("edit_timeout_seconds", item.timeout_seconds);
  setValue("edit_request_sleep_ms", item.request_sleep_ms);
  setValue("edit_retry_sleep_ms", item.retry_sleep_ms);
  setValue("edit_max_retry_times", item.max_retry_times);
  const msg = document.getElementById("fetchSiteEditMsg");
  if (msg) {
    msg.style.display = "none";
    msg.textContent = "";
    msg.className = "small mt-3";
  }
  modalOf("fetchSiteEditModal")?.show();
}

function numberValue(id) {
  const el = document.getElementById(id);
  return Number(el ? el.value : 0);
}

function payloadFromForm() {
  return {
    site_name: String(document.getElementById("edit_site_name")?.value || "").trim(),
    base_url: String(document.getElementById("edit_base_url")?.value || "").trim(),
    user_agent: String(document.getElementById("edit_user_agent")?.value || "").trim(),
    cookie: String(document.getElementById("edit_cookie")?.value || "").trim(),
    proxy: String(document.getElementById("edit_proxy")?.value || "").trim(),
    timeout_seconds: numberValue("edit_timeout_seconds"),
    request_sleep_ms: numberValue("edit_request_sleep_ms"),
    retry_sleep_ms: numberValue("edit_retry_sleep_ms"),
    max_retry_times: numberValue("edit_max_retry_times"),
    status: numberValue("edit_status")
  };
}

function showEditMsg(message, ok = true) {
  const msg = document.getElementById("fetchSiteEditMsg");
  if (!msg) return;
  msg.textContent = message;
  msg.className = `${ok ? "text-success" : "text-danger"} small mt-3`;
  msg.style.display = "block";
}

async function load() {
  const data = await request("/api/fetch-sites");
  fetchSiteItems = Array.isArray(data.items) ? data.items : [];
  renderPage(fetchSiteItems);
}

document.addEventListener("click", (event) => {
  const editButton = event.target.closest(".js-edit-fetch-site");
  if (!editButton) return;
  const item = findItem(String(editButton.dataset.siteCode || ""));
  openEditModal(item);
});

document.addEventListener("submit", async (event) => {
  const form = event.target.closest("#fetchSiteEditForm");
  if (!form) return;
  event.preventDefault();

  const siteCode = String(document.getElementById("edit_site_code")?.value || "").trim();
  const submitButton = form.querySelector('button[type="submit"]');
  if (!siteCode) {
    showEditMsg("缺少 site_code", false);
    return;
  }
  if (submitButton) submitButton.disabled = true;
  try {
    await request(`/api/fetch-sites/${encodeURIComponent(siteCode)}`, {
      method: "PUT",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(payloadFromForm())
    });
    showEditMsg("保存成功", true);
    showMsg("FetchSite 配置已保存");
    modalOf("fetchSiteEditModal")?.hide();
    await load();
  } catch (error) {
    showEditMsg(error.message || "保存失败", false);
  } finally {
    if (submitButton) submitButton.disabled = false;
  }
});

load().catch((error) => {
  root.innerHTML = `<div class="alert alert-danger">${escapeHtml(error.message || "加载失败")}</div>`;
});
