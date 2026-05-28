const root = document.getElementById("page-root");

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
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

function renderPage(data) {
  return `<div id="crawlerDetailLoopPage">
    <div class="row">
      <div class="col-12">
        <div class="page-title-box d-flex align-items-center justify-content-between">
          <h4 class="mb-0">${escapeHtml(data.Title || "详情抓取循环")}</h4>
          <a href="/triggers/dailybest" class="btn btn-outline-primary btn-sm">回到 DailyBest</a>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-lg-5">
        <div class="card">
          <div class="card-body">
            <div class="d-flex align-items-center justify-content-between mb-3">
              <h4 class="card-title mb-0">循环状态</h4>
              <span id="runningBadge" class="badge bg-light text-dark">检测中</span>
            </div>

            <div class="detail-loop-status-grid">
              <div class="overview-item">
                <div class="label">启动时间</div>
                <div class="value detail-loop-text" id="startedAtText">-</div>
              </div>
              <div class="overview-item">
                <div class="label">最近日志</div>
                <div class="value detail-loop-text" id="lastEventAtText">-</div>
              </div>
              <div class="overview-item">
                <div class="label">缓冲数量</div>
                <div class="value" id="bufferedText">0</div>
              </div>
              <div class="overview-item">
                <div class="label">暂停状态</div>
                <div class="value" id="pausedText">否</div>
              </div>
            </div>

            <div class="mt-3">
              <div class="small text-muted mb-2">最近一条</div>
              <div id="lastLineText" class="detail-loop-last-line">-</div>
            </div>

            <div class="detail-loop-actions">
              <button type="button" class="btn btn-primary" id="startBtn">开启 loop</button>
              <button type="button" class="btn btn-outline-danger" id="stopBtn">停止 loop</button>
            </div>

            <div id="statusMsg" class="small text-muted mt-3">等待加载</div>
          </div>
        </div>
      </div>

      <div class="col-lg-7">
        <div class="card event-card detail-loop-log-card">
          <div class="card-body">
            <div class="d-flex align-items-center justify-content-between mb-3">
              <h4 class="card-title mb-0">实时日志</h4>
              <span id="eventCountText" class="badge bg-light text-dark">0 条</span>
            </div>
            <div class="result-box">
              <pre id="resultBox">等待连接...</pre>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>`;
}

function loadLegacyScript() {
  const script = document.createElement("script");
  script.src = "/static/customer/crawler_detail_loop.js";
  script.async = false;
  document.body.appendChild(script);
}

async function main() {
  try {
    const data = await fetchJSON("/api/crawler/detail-loop/page");
    document.title = `${data.Title || "详情抓取循环"} - Rudy-GC`;
    window.CRAWLER_DETAIL_LOOP_PAGE_DATA = {
      apiBase: data.ApiBase || "/api/crawler/detail-loop",
    };
    root.innerHTML = renderPage(data);
    loadLegacyScript();
  } catch (err) {
    root.innerHTML = `<div class="alert alert-danger">加载失败：${escapeHtml(err.message)}</div>`;
  }
}

main();
