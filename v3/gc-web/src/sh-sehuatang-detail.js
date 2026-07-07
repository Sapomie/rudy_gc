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

function text(value, fallback = "-") {
  const current = String(value ?? "").trim();
  return current === "" ? fallback : current;
}

function formatUnix(sec) {
  const n = Number(sec || 0);
  if (n <= 0) {
    return "-";
  }
  const d = new Date(n * 1000);
  if (Number.isNaN(d.getTime())) {
    return "-";
  }
  const year = String(d.getFullYear());
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  const hour = String(d.getHours()).padStart(2, "0");
  const minute = String(d.getMinutes()).padStart(2, "0");
  const second = String(d.getSeconds()).padStart(2, "0");
  return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
}

function formatSHDetailStatus(status) {
  const current = String(status || "").trim();
  if (current === "pending") {
    return "待抓详情";
  }
  if (current === "fetched") {
    return "已抓详情";
  }
  if (current === "failed") {
    return "详情失败";
  }
  return text(current);
}

function lines(value) {
  return String(value || "")
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line !== "");
}

async function fetchJSON(path) {
  const resp = await fetch(path);
  const body = await resp.text();
  let data = {};
  try {
    data = body ? JSON.parse(body) : {};
  } catch {
    data = {error: body};
  }
  if (!resp.ok) {
    throw new Error(data.error || resp.statusText);
  }
  return data;
}

async function copyToClipboard(value) {
  const content = String(value || "").trim();
  if (!content) {
    return;
  }
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(content);
    return;
  }
  const textarea = document.createElement("textarea");
  textarea.value = content;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand("copy");
  textarea.remove();
}

function showMessage(message, level) {
  const box = document.getElementById("shDetailMsg");
  if (!box) {
    return;
  }
  box.className = `alert alert-${level || "info"} mb-3`;
  box.textContent = message;
  window.setTimeout(() => {
    box.classList.add("d-none");
  }, 2500);
}

function renderInfoCard(thread) {
  return `<div class="card mb-3">
    <div class="card-body">
      <div class="row g-3">
        <div class="col-12 col-md-4 col-xl-2">
          <div class="text-muted small">thread_tid</div>
          <div class="fw-semibold">${escapeHtml(thread.ThreadTID || "")}</div>
        </div>
        <div class="col-12 col-md-4 col-xl-2">
          <div class="text-muted small">来源版块</div>
          <div class="fw-semibold">${escapeHtml(text(thread.SourceForumID))}</div>
        </div>
        <div class="col-12 col-md-4 col-xl-2">
          <div class="text-muted small">来源页</div>
          <div class="fw-semibold">${thread.SourcePageURL
            ? `<a href="${attr(thread.SourcePageURL)}" target="_blank" rel="noopener noreferrer">${escapeHtml(text(thread.SourcePage))}</a>`
            : escapeHtml(text(thread.SourcePage))}</div>
        </div>
        <div class="col-12 col-md-4 col-xl-2">
          <div class="text-muted small">详情状态</div>
          <div class="fw-semibold">${escapeHtml(formatSHDetailStatus(thread.DetailStatus))}</div>
        </div>
        <div class="col-12 col-md-4 col-xl-2">
          <div class="text-muted small">作者</div>
          <div class="fw-semibold">${escapeHtml(text(thread.AuthorName))}</div>
        </div>
        <div class="col-12 col-md-4 col-xl-2">
          <div class="text-muted small">发布时间</div>
          <div class="fw-semibold">${escapeHtml(formatUnix(thread.PostTime))}</div>
        </div>
        <div class="col-12 col-md-4 col-xl-2">
          <div class="text-muted small">最近抓到</div>
          <div class="fw-semibold">${escapeHtml(formatUnix(thread.LastSeenTime))}</div>
        </div>
        <div class="col-12 col-md-4 col-xl-2">
          <div class="text-muted small">详情抓取时间</div>
          <div class="fw-semibold">${escapeHtml(formatUnix(thread.DetailFetchedOn))}</div>
        </div>
        <div class="col-12 col-md-4 col-xl-2">
          <div class="text-muted small">详情失败时间</div>
          <div class="fw-semibold">${escapeHtml(formatUnix(thread.DetailFailedOn))}</div>
        </div>
        <div class="col-12 col-md-4 col-xl-2">
          <div class="text-muted small">入库时间</div>
          <div class="fw-semibold">${escapeHtml(formatUnix(thread.CreatedOn))}</div>
        </div>
        <div class="col-12 col-md-4 col-xl-2">
          <div class="text-muted small">更新时间</div>
          <div class="fw-semibold">${escapeHtml(formatUnix(thread.UpdatedOn))}</div>
        </div>
        <div class="col-12">
          <div class="text-muted small">content_hash</div>
          <div class="font-monospace small">${escapeHtml(thread.ContentHash || "")}</div>
        </div>
        ${thread.DetailError ? `<div class="col-12">
          <div class="text-muted small">详情错误</div>
          <div class="small text-danger">${escapeHtml(thread.DetailError)}</div>
        </div>` : ""}
      </div>
    </div>
  </div>`;
}

function renderCopyList(title, items, copyAllLabel) {
  return `<div class="card mb-3">
    <div class="card-body">
      <div class="d-flex align-items-center justify-content-between gap-2 mb-3">
        <h5 class="card-title mb-0">${escapeHtml(title)}</h5>
        <button type="button" class="btn btn-sm btn-outline-secondary js-copy-value" data-copy-value="${attr(items.join("\n"))}" ${items.length ? "" : "disabled"}>${escapeHtml(copyAllLabel)}</button>
      </div>
      ${items.length
        ? `<div class="list-group list-group-flush">
          ${items.map((item) => `<div class="list-group-item px-0 d-flex align-items-start gap-2">
            <code class="flex-grow-1 text-break">${escapeHtml(item)}</code>
            <button type="button" class="btn btn-sm btn-outline-secondary js-copy-value" data-copy-value="${attr(item)}">复制</button>
          </div>`).join("")}
        </div>`
        : `<div class="text-muted small">暂无数据。</div>`}
    </div>
  </div>`;
}

function renderContent(thread) {
  return `<div class="card mb-3">
    <div class="card-body">
      <div class="d-flex align-items-center justify-content-between gap-2 mb-3">
        <h5 class="card-title mb-0">正文</h5>
        <button type="button" class="btn btn-sm btn-outline-secondary js-copy-value" data-copy-value="${attr(thread.RawContentText || "")}" ${thread.RawContentText ? "" : "disabled"}>复制正文</button>
      </div>
      <pre class="bg-light border rounded p-3 small mb-0">${escapeHtml(thread.RawContentText || "")}</pre>
    </div>
  </div>`;
}

function renderFetchItems(items) {
  const rows = Array.isArray(items) ? items : [];
  return `<div class="card">
    <div class="card-body">
      <h5 class="card-title mb-3">相关抓取记录</h5>
      <div class="table-responsive">
        <table class="table table-sm align-middle mb-0">
          <thead>
            <tr>
              <th>run_id</th>
              <th>模式</th>
              <th>来源版块</th>
              <th>来源页</th>
              <th>状态</th>
              <th>thread_row_id</th>
              <th>更新时间</th>
              <th>错误</th>
            </tr>
          </thead>
          <tbody>
            ${rows.length ? rows.map((item) => `<tr>
              <td>${escapeHtml(item.RunID || "")}</td>
              <td>${escapeHtml(item.FetchMode || "")}</td>
              <td>${escapeHtml(item.SourceForumID || "")}</td>
              <td>${item.SourcePageURL
                ? `<a href="${attr(item.SourcePageURL)}" target="_blank" rel="noopener noreferrer">${escapeHtml(item.SourcePage || "")}</a>`
                : escapeHtml(item.SourcePage || "")}</td>
              <td><span class="badge fetch-site-badge">${escapeHtml(item.Status || "")}</span></td>
              <td>${escapeHtml(item.ThreadRowID || "")}</td>
              <td class="text-nowrap">${escapeHtml(formatUnix(item.UpdatedOn))}</td>
              <td>${item.Error ? `<details><summary>错误</summary><div class="mt-2">${escapeHtml(item.Error)}</div></details>` : "-"}</td>
            </tr>`).join("") : `<tr><td colspan="8"><div class="fetch-site-empty-state">暂无相关抓取记录。</div></td></tr>`}
          </tbody>
        </table>
      </div>
    </div>
  </div>`;
}

function renderPage(data) {
  const thread = data.Thread || {};
  const magnets = lines(thread.MagnetText);
  const hashes = lines(thread.InfoHashText);
  return `<div id="shDetailPage">
    <div class="d-flex flex-column flex-xl-row align-items-xl-start justify-content-between gap-3 mb-3">
      <div>
        <h4 class="mb-1">${escapeHtml(thread.ThreadTitle || data.PageTitle || "Sehuatang SH 帖子详情")}</h4>
        <p class="text-muted mb-0">${escapeHtml(data.PageTitle || "")}</p>
      </div>
      <div class="d-flex flex-wrap gap-2">
        <a class="btn btn-outline-primary btn-sm" href="${attr(data.ListPageURL || "/fetch-site-sehuatang-sh-list")}">返回列表</a>
        ${thread.ThreadURL ? `<a class="btn btn-outline-secondary btn-sm" href="${attr(thread.ThreadURL)}" target="_blank" rel="noopener noreferrer">打开原帖</a>` : ""}
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.TaskPageURL || "/triggers/fetch-sehuatang-sh-thread-detail")}">详情补全</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.ListTaskPageURL || "/triggers/fetch-sehuatang-sh-thread-list")}">列表抓取</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.TasksPageURL || "/crawler/tasks")}">任务列表</a>
      </div>
    </div>
    <div id="shDetailMsg" class="alert d-none mb-3" role="alert"></div>
    ${renderInfoCard(thread)}
    ${renderContent(thread)}
    ${renderCopyList("magnet", magnets, "复制全部 magnet")}
    ${renderCopyList("info_hash", hashes, "复制全部 hash")}
    ${renderFetchItems(data.FetchItems)}
  </div>`;
}

function initPage() {
  const page = document.getElementById("shDetailPage");
  if (!page) {
    return;
  }
  page.addEventListener("click", async (event) => {
    const button = event.target.closest(".js-copy-value");
    if (!button) {
      return;
    }
    try {
      await copyToClipboard(button.getAttribute("data-copy-value") || "");
      showMessage("已复制", "success");
    } catch (err) {
      showMessage(`复制失败：${err.message || err}`, "danger");
    }
  });
}

async function main() {
  try {
    const data = await fetchJSON("/api" + window.location.pathname);
    document.title = `${data.Title || data.PageTitle || "Sehuatang SH Detail"} - Rudy-GC`;
    root.innerHTML = renderPage(data);
    initPage();
  } catch (err) {
    root.innerHTML = `<div class="alert alert-danger">加载失败：${escapeHtml(err.message)}</div>`;
  }
}

main();
