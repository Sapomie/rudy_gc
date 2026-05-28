const root = document.getElementById("page-root");

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

async function fetchJSON(path) {
  const resp = await fetch(path);
  const text = await resp.text();
  let data = {};
  try {
    data = text ? JSON.parse(text) : {};
  } catch {
    data = {error: text};
  }
  if (!resp.ok) {
    throw new Error(data.error || resp.statusText);
  }
  return data;
}

function backHref(data) {
  const base = String(data.TriggerPageURL || "/triggers/dailybest");
  const jobID = String(data.JobID || "").trim();
  if (!jobID) {
    return base;
  }
  return `${base}?job_id=${encodeURIComponent(jobID)}`;
}

function renderPage(data) {
  return `<div id="crawlerTasksPage">
    <div class="row">
      <div class="col-12">
        <div class="page-title-box d-flex align-items-center justify-content-between">
          <h4 class="mb-0">${escapeHtml(data.Title || "任务列表")}</h4>
          <a href="${attr(backHref(data))}" class="btn btn-outline-primary btn-sm">回到任务台</a>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-12">
        <div class="card">
          <div class="card-body">
            <div class="d-flex align-items-center justify-content-between mb-3">
              <h4 class="card-title mb-0">任务列表</h4>
              <span id="taskCountBadge" class="badge bg-light text-dark">0 个</span>
            </div>
            <div class="table-responsive">
              <table class="table table-centered table-nowrap mb-0">
                <thead>
                  <tr>
                    <th>任务ID</th>
                    <th>任务</th>
                    <th>当前状态</th>
                    <th>待处理</th>
                    <th>已处理</th>
                    <th>成功/失败</th>
                    <th>开始时间</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody id="taskTableBody">
                  <tr>
                    <td colspan="8" class="text-center text-muted py-4">加载中...</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>`;
}

function loadLegacyScript() {
  const script = document.createElement("script");
  script.src = "/static/customer/crawler_tasks.js";
  script.async = false;
  document.body.appendChild(script);
}

async function main() {
  try {
    const data = await fetchJSON(`/api/crawler/tasks${window.location.search}`);
    document.title = `${data.Title || "任务列表"} - Rudy-GC`;
    window.CRAWLER_TASKS_PAGE_DATA = {
      initialJobId: data.JobID || "",
      triggerPageUrl: data.TriggerPageURL || "/triggers/dailybest",
    };
    root.innerHTML = renderPage(data);
    loadLegacyScript();
  } catch (err) {
    root.innerHTML = `<div class="alert alert-danger">加载失败：${escapeHtml(err.message)}</div>`;
  }
}

main();
