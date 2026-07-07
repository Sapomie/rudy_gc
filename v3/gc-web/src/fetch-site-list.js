const root = document.getElementById("page-root");

const pageAPIs = {
  "/fetch-site-javbus-list": "/api/fetch-site-javbus-list",
  "/fetch-site-sukebei-list": "/api/fetch-site-sukebei-list",
  "/fetch-site-sehuatang-list": "/api/fetch-site-sehuatang-list",
  "/fetch-site-sehuatang-sh-list": "/api/fetch-site-sehuatang-sh-list",
};

let shThreadRows = [];

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

function valueOf(value, fallback = "") {
  return value === undefined || value === null ? fallback : value;
}

function text(value, fallback = "-") {
  const current = String(valueOf(value, "")).trim();
  return current === "" ? fallback : current;
}

function boolClass(ok, yes, no) {
  return ok ? yes : no;
}

function formatUnix(sec, mode) {
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
  if (mode === "date") {
    return `${year}-${month}-${day}`;
  }
  if (mode === "hm") {
    return `${hour}:${minute}`;
  }
  return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
}

function apiPathForCurrentPage() {
  return pageAPIs[window.location.pathname] || "/api/fetch-site-javbus-list";
}

function syncHeaderNav(pathname) {
  document.querySelectorAll("[data-fetch-site-nav]").forEach((link) => {
    const active = link.getAttribute("data-fetch-site-nav") === pathname;
    link.classList.toggle("btn-primary", active);
    link.classList.toggle("btn-outline-primary", !active);
  });
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

function loadScriptOnce(src, id) {
  if (document.getElementById(id)) {
    return;
  }
  const script = document.createElement("script");
  script.id = id;
  script.src = src;
  script.async = false;
  document.body.appendChild(script);
}

function renderPageLinks(pageInfo, label) {
  const links = Array.isArray(pageInfo?.Links) ? pageInfo.Links : [];
  if (!links.length) {
    return "";
  }
  return `<nav class="mb-0" aria-label="${attr(label)} Page navigation">
    <ul class="pagination pagination-sm mb-0">
      ${links.map((link) => {
        if (link.Ellipsis) {
          return `<li class="page-item disabled"><span class="page-link">...</span></li>`;
        }
        return `<li class="page-item ${link.Active ? "active" : ""} ${link.Disabled ? "disabled" : ""}">
          <a class="page-link" href="${attr(link.Href || "#")}" ${link.Disabled ? `tabindex="-1" aria-disabled="true"` : ""}>${escapeHtml(link.Label || "")}</a>
        </li>`;
      }).join("")}
    </ul>
  </nav>`;
}

function sortLabel(link, label) {
  const current = link || {};
  const arrow = current.Active ? (current.Desc ? " ↓" : " ↑") : "";
  return `<a class="fetch-site-sort-link ${current.Active ? "active" : ""}" href="${attr(current.Href || "#")}">${escapeHtml(label)}${arrow}</a>`;
}

function renderOwnedFilter(data) {
  const owned = data.ownedQuery || {};
  const total = valueOf(data.total, data.Total || 0);
  const link = (item, label, className) => {
    const current = item || {};
    return `<a href="${attr(current.Href || "#")}"
      class="btn btn-sm ${className} waves-effect waves-light ${current.Active ? "active disabled" : ""}"
      aria-current="${current.Active ? "true" : "false"}">${escapeHtml(label)}</a>`;
  };
  return `<div class="d-flex flex-wrap align-items-center gap-3">
    <div class="d-flex flex-wrap align-items-center gap-2">
      <span class="btn btn-sm btn-outline-dark disabled">WMedia</span>
      ${link(owned.MediaAll, "MAll", "btn-primary")}
      ${link(owned.MediaOwned, "MOwned", "btn-success")}
      ${link(owned.MediaNotOwned, "MNotOwned", "btn-secondary")}
    </div>
    <span class="btn btn-sm btn-primary waves-effect waves-light disabled">Total: ${escapeHtml(total)}</span>
  </div>`;
}

function statusChecked(data, value) {
  const selected = data.StatusSelected || {};
  return selected[String(value)] ? "checked" : "";
}

function renderFetchSiteStatusGroup(data, siteKey) {
  const allActive = data.StatusAllActive ? "active" : "";
  const statusID = (value) => `${siteKey}-status-${value}`;
  return `<div class="fetch-site-status-group" data-status-group>
    <button type="button" class="btn btn-sm btn-outline-secondary fetch-site-status-clear ${allActive}" data-status-clear>全部</button>
    <input type="checkbox" class="btn-check" id="${attr(statusID(1))}" name="statuses" value="1" autocomplete="off" ${statusChecked(data, 1)}>
    <label class="btn btn-sm btn-outline-primary" for="${attr(statusID(1))}">待抓取</label>
    <input type="checkbox" class="btn-check" id="${attr(statusID(2))}" name="statuses" value="2" autocomplete="off" ${statusChecked(data, 2)}>
    <label class="btn btn-sm btn-outline-info" for="${attr(statusID(2))}">抓取中</label>
    <input type="checkbox" class="btn-check" id="${attr(statusID(3))}" name="statuses" value="3" autocomplete="off" ${statusChecked(data, 3)}>
    <label class="btn btn-sm btn-outline-success" for="${attr(statusID(3))}">成功</label>
    <input type="checkbox" class="btn-check" id="${attr(statusID(4))}" name="statuses" value="4" autocomplete="off" ${statusChecked(data, 4)}>
    <label class="btn btn-sm btn-outline-danger" for="${attr(statusID(4))}">失败</label>
  </div>`;
}

function renderTriggerSortOptions(query) {
  const current = String(query.TriggerSortKey || "");
  const options = [
    ["last_fetch_time:asc", "最后抓取时间升序"],
    ["last_fetch_time:desc", "最后抓取时间降序"],
    ["release_date:asc", "发行时间升序"],
    ["release_date:desc", "发行时间降序"],
    ["latest_publish_time:asc", "最新日期升序"],
    ["latest_publish_time:desc", "最新日期降序"],
    ["movie_name:asc", "影片名升序"],
    ["movie_name:desc", "影片名降序"],
  ];
  return options.map(([value, label]) => {
    const selected = (current === "" && value === "last_fetch_time:asc") || current === value;
    return `<option value="${attr(value)}" ${selected ? "selected" : ""}>${escapeHtml(label)}</option>`;
  }).join("");
}

function renderFetchSiteFilterForm(data, site) {
  const q = data.Query || {};
  const action = site === "javbus" ? "/fetch-site-javbus-list" : "/fetch-site-sukebei-list";
  const formID = site === "javbus" ? "fetchSiteJavbusFilterForm" : "fetchSiteSukebeiFilterForm";
  const siteLabel = site === "javbus" ? "JavBus" : "Sukebei";
  const triggerClass = site === "javbus" ? "js-trigger-filtered-fetch-site" : "js-trigger-filtered-sukebei";
  return `<div class="card mb-3">
    <div class="card-body">
      <form class="row g-3 align-items-end" method="get" action="${attr(action)}" id="${attr(formID)}" ${site === "javbus" ? `data-fetch-site-filter-form="1"` : ""}>
        <input type="hidden" name="ps" value="${attr(q.PageSize || "")}">
        ${q.Sort ? `<input type="hidden" name="sort" value="${attr(q.Sort)}">` : ""}
        ${q.Order ? `<input type="hidden" name="order" value="${attr(q.Order)}">` : ""}
        ${q.MediaOwned ? `<input type="hidden" name="mowned" value="${attr(q.MediaOwned)}">` : ""}
        <div class="col-12">
          <div class="fetch-site-filter-grid fetch-site-filter-grid-5 fetch-site-sukebei-filter-top">
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">${escapeHtml(siteLabel)} 影片编号 / JavID</label>
              <input class="form-control" type="text" name="keyword" value="${attr(q.Keyword || "")}" placeholder="movie_name 或 movie_jav_id">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">${escapeHtml(siteLabel)} 状态</label>
              ${renderFetchSiteStatusGroup(data, site)}
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">最后抓取开始</label>
              <input class="form-control" type="text" name="last_fetch_from" value="${attr(q.LastFetchFrom || "")}">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">最后抓取结束</label>
              <input class="form-control" type="text" name="last_fetch_to" value="${attr(q.LastFetchTo || "")}">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">触发抓取顺序</label>
              <select class="form-select" name="trigger_sort_key">${renderTriggerSortOptions(q)}</select>
            </div>
          </div>
        </div>
        <div class="col-12">
          <div class="fetch-site-filter-grid fetch-site-filter-grid-6">
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">发行时间开始</label>
              <input class="form-control" type="text" name="release_date_from" value="${attr(q.ReleaseDateFrom || "")}">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">发行时间结束</label>
              <input class="form-control" type="text" name="release_date_to" value="${attr(q.ReleaseDateTo || "")}">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">WMedia 下载时间开始</label>
              <input class="form-control" type="text" name="media_birth_from" value="${attr(q.MediaBirthFrom || "")}">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">WMedia 下载时间结束</label>
              <input class="form-control" type="text" name="media_birth_to" value="${attr(q.MediaBirthTo || "")}">
            </div>
          </div>
        </div>
        <div class="col-12 d-flex gap-2">
          <button class="btn btn-primary" type="submit">筛选</button>
          <a class="btn btn-outline-secondary" href="${attr(action)}">清空</a>
          <button class="btn btn-outline-danger ${attr(triggerClass)}" type="button">触发筛选结果抓取</button>
        </div>
      </form>
      <div class="mt-3">${renderOwnedFilter(data)}</div>
    </div>
  </div>`;
}

function renderWMediaCell(row) {
  const owned = Number(row.OwnedWMedia || 0);
  const birth = row.FilmBirthDateWMedia
    ? `<span class="text-muted small ms-1">${escapeHtml(row.FilmBirthDateWMedia)}</span>`
    : "";
  if (owned === 5) {
    return `<a href="${attr(row.VideoURLWMedia || "#")}" class="btn btn-sm btn-info me-2">MPlay</a>${birth}`;
  }
  if (owned === 4) {
    return `<a href="${attr(row.VideoURLWMedia || "#")}" class="btn btn-sm btn-success me-2">MPlay</a>${birth}`;
  }
  if (owned === 6) {
    return `<span class="btn btn-sm btn-secondary me-2 disabled" aria-disabled="true">MRemoved</span>${birth}`;
  }
  return `-${birth}`;
}

function renderFetchSiteRow(row, site) {
  const triggerClass = site === "javbus" ? "js-trigger-javbus" : "js-trigger-sukebei";
  const triggerLabel = site === "javbus" ? "触发 JavBus" : "触发 Sukebei";
  const triggerButtonClass = site === "javbus" ? "btn-outline-primary" : "btn-outline-danger";
  const errorHTML = row.LastError
    ? `<details class="fetch-site-error-details"><summary>有错误</summary><div class="mt-2">${escapeHtml(row.LastError)}</div></details>`
    : `<span class="text-muted">无错误</span>`;
  return `<tr data-movie-jav-id="${attr(row.MovieJavID || "")}" data-movie-name="${attr(row.MovieName || "")}">
    <td class="text-nowrap"><div class="fw-semibold"><a href="/movie/${attr(encodeURIComponent(row.MovieName || ""))}">${escapeHtml(row.MovieName || "")}</a></div></td>
    <td class="text-nowrap">${escapeHtml(text(row.ReleaseDateText))}</td>
    <td><span class="badge fetch-site-badge">${escapeHtml(text(row.FetchStatusText))}</span></td>
    <td>${escapeHtml(valueOf(row.TryCount, 0))}</td>
    <td class="text-nowrap">${escapeHtml(text(row.LastFetchText))}</td>
    <td>${escapeHtml(valueOf(row.TorrentHashCount, 0))}</td>
    <td class="text-nowrap">${escapeHtml(text(row.LatestPublishText))}</td>
    <td class="text-nowrap">${renderWMediaCell(row)}</td>
    <td class="fetch-site-error-cell">${errorHTML}</td>
    <td class="text-nowrap"><button type="button" class="btn btn-sm ${triggerButtonClass} ${triggerClass}">${escapeHtml(triggerLabel)}</button></td>
  </tr>`;
}

function renderFetchSiteTable(data, site) {
  const rows = Array.isArray(data.Rows) ? data.Rows : [];
  const sort = data.SortQuery || {};
  const siteLabel = site === "javbus" ? "JavBus" : "Sukebei";
  return `<div class="card">
    <div class="card-body">
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0">
          <thead>
            <tr>
              <th>${sortLabel(sort.ByMovieName, "影片")}</th>
              <th>${sortLabel(sort.ByReleaseDate, "发行时间")}</th>
              <th>${sortLabel(sort.ByFetchStatus, "状态")}</th>
              <th>重试</th>
              <th>${sortLabel(sort.ByLastFetchTime, "最后抓取")}</th>
              <th>${sortLabel(sort.ByHashCount, "Hash 数")}</th>
              <th>${sortLabel(sort.ByLatestPublish, "最新日期")}</th>
              <th>${sortLabel(sort.ByMediaBirthTime, "WMedia 下载时间")}</th>
              <th>错误</th>
              <th class="text-nowrap">操作</th>
            </tr>
          </thead>
          <tbody>
            ${rows.length
              ? rows.map((row) => renderFetchSiteRow(row, site)).join("")
              : `<tr><td colspan="11"><div class="fetch-site-empty-state">当前筛选下没有 ${escapeHtml(siteLabel)} 抓取任务。</div></td></tr>`}
          </tbody>
        </table>
      </div>
    </div>
  </div>`;
}

function renderFetchSitePage(data, site) {
  const siteLabel = site === "javbus" ? "JavBus" : "Sukebei";
  const peerLabel = site === "javbus" ? "Sukebei" : "JavBus";
  const ownURL = site === "javbus" ? "/fetch-site-javbus-list" : "/fetch-site-sukebei-list";
  const filteredTaskType = site === "javbus" ? data.FilteredJavbusTaskType : data.FilteredSukebeiTaskType;
  const formID = site === "javbus" ? "fetchSiteJavbusFilterForm" : "fetchSiteSukebeiFilterForm";
  const filteredAttr = site === "javbus"
    ? `data-filtered-task-type="${attr(filteredTaskType || "")}" data-filter-form-id="${attr(formID)}"`
    : `data-filtered-sukebei-task-type="${attr(filteredTaskType || "")}"`;
  return `<div id="fetchSiteListPage" data-task-page-url="${attr(data.TaskPageURL || "/triggers/fetch-site")}" ${filteredAttr}>
    <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
      <div>
        <h4 class="mb-1">${escapeHtml(data.PageTitle || `${siteLabel} 抓取列表`)}</h4>
        <p class="text-muted mb-0">${escapeHtml(data.PageNote || "")}</p>
      </div>
      <div class="d-flex flex-wrap gap-2">
        <a class="btn btn-outline-primary btn-sm" href="${attr(ownURL)}">${escapeHtml(siteLabel)} 列表</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.PeerPageURL || "#")}">${escapeHtml(peerLabel)} 列表</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.TaskPageURL || "#")}">抓取任务页</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.TasksPageURL || "/crawler/tasks")}">任务列表</a>
      </div>
    </div>
    <div id="fetchSiteListMsg" class="alert d-none mb-3" role="alert"></div>

    <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
      <div class="d-flex flex-wrap gap-3">
        <div class="fetch-site-summary-card">
          <span class="fetch-site-summary-label">${escapeHtml(siteLabel)} 总任务数</span>
          <span class="fetch-site-summary-value">${escapeHtml(valueOf(data.Total, 0))}</span>
        </div>
        <div class="fetch-site-summary-card">
          <span class="fetch-site-summary-label">成功数目</span>
          <span class="fetch-site-summary-value">${escapeHtml(valueOf(data.SuccessCount, 0))}</span>
        </div>
        <div class="fetch-site-summary-card">
          <span class="fetch-site-summary-label">待抓取数目</span>
          <span class="fetch-site-summary-value">${escapeHtml(valueOf(data.PendingCount, 0))}</span>
        </div>
        <div class="fetch-site-summary-card">
          <span class="fetch-site-summary-label">失败数目</span>
          <span class="fetch-site-summary-value">${escapeHtml(valueOf(data.FailedCount, 0))}</span>
        </div>
      </div>
      ${(Array.isArray(data.Rows) && data.Rows.length) ? renderPageLinks(data.PageInfo, siteLabel) : ""}
    </div>

    ${renderFetchSiteFilterForm(data, site)}
    ${renderFetchSiteTable(data, site)}
  </div>`;
}

function renderQuickTags(data) {
  const tags = Array.isArray(data.QuickTags) ? data.QuickTags : [];
  return tags.map((item) =>
    `<a class="btn btn-sm ${item.Active ? "btn-primary" : "btn-outline-primary"}" href="${attr(item.Href || "#")}">${escapeHtml(item.Label || "")}</a>`
  ).join("");
}

function renderSehuatangFilterForm(data) {
  const q = data.Query || {};
  return `<div class="card mb-3">
    <div class="card-body">
      <form class="row g-3 align-items-end" id="sehuatangMagnetFilterForm" method="get" action="/fetch-site-sehuatang-list">
        <input type="hidden" name="ps" value="${attr(q.PageSize || "")}">
        ${q.Sort ? `<input type="hidden" name="sort" value="${attr(q.Sort)}">` : ""}
        ${q.Order ? `<input type="hidden" name="order" value="${attr(q.Order)}">` : ""}
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">关键词</label>
          <input class="form-control" type="text" name="keyword" value="${attr(q.Keyword || "")}" placeholder="movie_name / movie_jav_id / thread_title">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">movie_jav_id</label>
          <input class="form-control" type="text" name="movie_jav_id" value="${attr(q.MovieJavID || "")}" placeholder="JavLib 编号">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">info_hash</label>
          <input class="form-control" type="text" name="info_hash" value="${attr(q.InfoHash || "")}" placeholder="BTIH">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">tag</label>
          <input class="form-control" id="sehuatangMagnetTagInput" type="text" name="tag" value="${attr(q.Tag || "")}" placeholder="标签">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">发帖时间开始</label>
          <input class="form-control" type="text" name="post_time_from" value="${attr(q.PostTimeFrom || "")}" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">发帖时间结束</label>
          <input class="form-control" type="text" name="post_time_to" value="${attr(q.PostTimeTo || "")}" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
        </div>
        <div class="col-12">
          <label class="form-label mb-1">快捷 tag</label>
          <div class="d-flex flex-wrap gap-2 album-quick-filter-group">${renderQuickTags(data)}</div>
        </div>
        <div class="col-12 d-flex gap-2">
          <button class="btn btn-primary" type="submit">筛选</button>
          <a class="btn btn-outline-secondary" href="/fetch-site-sehuatang-list">清空</a>
        </div>
      </form>
    </div>
  </div>`;
}

function renderFavoriteButton(row, albumName, isInAlbum, className, outlineClass) {
  const canFavorite = !!row.CanFavorite;
  return `<button type="button"
    class="btn btn-sm ${isInAlbum ? className : outlineClass} js-add-favorite"
    data-source-type="sehuatang_magnet"
    data-source-row-id="${attr(row.Id || 0)}"
    data-favorited="${isInAlbum ? "1" : "0"}"
    data-album-name="${attr(albumName)}"
    ${canFavorite ? "" : `disabled title="缺少影片标识，暂不可加入相册"`}>${escapeHtml(albumName)}</button>`;
}

function renderSehuatangRow(row) {
  const postDay = formatUnix(row.PostTime, "date");
  const postHM = formatUnix(row.PostTime, "hm");
  const threadTitle = text(row.ThreadTitle, "");
  const titleHTML = row.ThreadURL
    ? `<a class="fetch-site-title-ellipsis" href="${attr(row.ThreadURL)}" target="_blank" rel="noopener noreferrer" title="${attr(threadTitle)}">${escapeHtml(threadTitle)}</a>`
    : `<span class="fetch-site-title-ellipsis" title="${attr(threadTitle)}">${escapeHtml(threadTitle)}</span>`;
  const movieHTML = row.MovieName
    ? `<a href="/movie/${attr(encodeURIComponent(row.MovieName))}">${escapeHtml(row.MovieName)}</a>`
    : "-";
  return `<tr class="js-sehuatang-row"
    data-movie-route-key="${attr(row.MovieRouteKey || "")}"
    data-movie-jav-id="${attr(row.MovieJavID || "")}"
    data-movie-name="${attr(row.MovieName || "")}"
    data-post-day="${attr(postDay)}">
    <td class="text-nowrap">
      <div class="fetch-site-time-cell">
        <span class="fetch-site-date">${escapeHtml(postDay)}</span>
        <span class="fetch-site-clock">${escapeHtml(postHM)}</span>
      </div>
    </td>
    <td>${titleHTML}</td>
    <td class="text-nowrap">${movieHTML}</td>
    <td class="text-nowrap">
      <div class="fetch-site-album-inline">
        ${renderFavoriteButton(row, "下载中", !!row.IsInDownload, "btn-warning", "btn-outline-warning")}
        ${renderFavoriteButton(row, "待下载", !!row.IsInPending, "btn-info", "btn-outline-info")}
        <button type="button"
          class="btn btn-sm btn-outline-secondary js-open-other-album-modal"
          data-source-type="sehuatang_magnet"
          data-source-row-id="${attr(row.Id || 0)}"
          ${row.CanFavorite ? "" : `disabled title="缺少影片标识，暂不可加入相册"`}>其他相册</button>
      </div>
    </td>
    <td class="text-nowrap"><span class="fetch-site-hash-text" title="${attr(row.InfoHash || "")}">${escapeHtml(row.InfoHash || "")}</span></td>
    <td class="text-nowrap">${escapeHtml(formatUnix(row.LastSeenTime))}</td>
  </tr>`;
}

function renderSehuatangTable(data) {
  const rows = Array.isArray(data.Rows) ? data.Rows : [];
  const sort = data.SortQuery || {};
  return `<div class="card">
    <div class="card-body">
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0">
          <thead>
            <tr>
              <th>${sortLabel(sort.ByPostTime, "发帖时间")}</th>
              <th>帖子标题</th>
              <th>${sortLabel(sort.ByMovieName, "影片名")}</th>
              <th class="text-nowrap">相册</th>
              <th>info_hash</th>
              <th>${sortLabel(sort.ByLastSeenTime, "最近抓到")}</th>
            </tr>
          </thead>
          <tbody>
            ${rows.length
              ? rows.map(renderSehuatangRow).join("")
              : `<tr><td colspan="6"><div class="fetch-site-empty-state">当前筛选下没有 Sehuatang 磁力数据。</div></td></tr>`}
          </tbody>
        </table>
      </div>
    </div>
  </div>`;
}

function renderOtherAlbumModal(albums) {
  const items = Array.isArray(albums) ? albums : [];
  return `<div class="modal fade" id="sehuatangOtherAlbumModal" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">选择其他相册</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="关闭"></button>
        </div>
        <div class="modal-body">
          ${items.length
            ? `<div class="fetch-site-other-album-list" id="sehuatangOtherAlbumList">
                ${items.map((name) => `<button type="button" class="btn btn-outline-secondary js-other-album-option" data-album-name="${attr(name)}">${escapeHtml(name)}</button>`).join("")}
              </div>`
            : `<div class="text-muted small">当前没有可选的其他相册。</div>`}
        </div>
      </div>
    </div>
  </div>`;
}

function renderDateModal() {
  return `<div class="modal fade" id="sehuatangDateModal" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <form id="sehuatangDateForm">
          <div class="modal-header">
            <h5 class="modal-title" id="sehuatangDateModalTitle">修改日期</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="关闭"></button>
          </div>
          <div class="modal-body">
            <input type="hidden" id="sehuatangDateItemKey" value="">
            <div class="mb-3">
              <label class="form-label mb-1" for="sehuatangDateInput">日期</label>
              <input class="form-control"
                id="sehuatangDateInput"
                type="text"
                value=""
                placeholder="YYYY-MM-DD"
                inputmode="numeric"
                maxlength="10"
                pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
              <div class="mt-2">
                <button type="button" class="btn btn-sm btn-outline-secondary" id="sehuatangDateTodayBtn">今天</button>
              </div>
            </div>
            <div class="alert d-none mb-0" id="sehuatangDateMsg" role="alert"></div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
            <button type="submit" class="btn btn-primary" id="sehuatangDateSubmit">保存</button>
          </div>
        </form>
      </div>
    </div>
  </div>`;
}

function renderSehuatangPage(data) {
  return `<div id="sehuatangMagnetListPage">
    <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
      <div>
        <h4 class="mb-1">${escapeHtml(data.PageTitle || "Sehuatang Magnet 列表")}</h4>
        <p class="text-muted mb-0">${escapeHtml(data.PageNote || "")}</p>
      </div>
      <div class="d-flex flex-wrap gap-2">
        <button type="button"
          class="btn btn-outline-primary btn-sm"
          id="btnEditShtTime"
          data-item-key="sht_time"
          data-item-label="SHT时间"
          data-item-value="${attr(data.ShtTimeValue || "")}">SHT时间：${escapeHtml(data.ShtTimeDisplay || "未设置")}</button>
        <a class="btn btn-outline-primary btn-sm" href="/fetch-site-sehuatang-list">Sehuatang 列表</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.TaskPageURL || "/triggers/fetch-sehuatang")}">抓取任务页</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.TasksPageURL || "/crawler/tasks")}">任务列表</a>
      </div>
    </div>
    <div id="sehuatangMagnetListMsg" class="alert d-none mb-3" role="alert"></div>

    <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
      <div class="fetch-site-summary-card">
        <span class="fetch-site-summary-label">Sehuatang 总记录数</span>
        <span class="fetch-site-summary-value">${escapeHtml(valueOf(data.Total, 0))}</span>
      </div>
      ${(Array.isArray(data.Rows) && data.Rows.length) ? renderPageLinks(data.PageInfo, "Sehuatang") : ""}
    </div>

    ${renderSehuatangFilterForm(data)}
    ${renderSehuatangTable(data)}
  </div>
  ${renderOtherAlbumModal(data.Albums)}
  ${renderDateModal()}`;
}

function lineCount(value) {
  const textValue = String(value || "").trim();
  if (!textValue) {
    return 0;
  }
  return textValue.split(/\r?\n/).filter((line) => line.trim() !== "").length;
}

function renderSHHasMagnetOptions(q) {
  const current = String(q.HasMagnet || "");
  const options = [
    ["", "全部"],
    ["1", "有 magnet"],
    ["0", "无 magnet"],
  ];
  return options.map(([value, label]) =>
    `<option value="${attr(value)}" ${current === value ? "selected" : ""}>${escapeHtml(label)}</option>`
  ).join("");
}

function renderSHSourceForumOptions(q) {
  const current = String(q.SourceForumID || "");
  const options = [
    ["", "全部"],
    ["36", "亚洲无码原创"],
    ["103", "高清中文字幕"],
  ];
  return options.map(([value, label]) =>
    `<option value="${attr(value)}" ${current === value ? "selected" : ""}>${escapeHtml(label)}</option>`
  ).join("");
}

function renderSHDetailStatusOptions(q) {
  const current = String(q.DetailStatus || "");
  const options = [
    ["", "全部"],
    ["pending", "待抓详情"],
    ["failed", "详情失败"],
    ["fetched", "已抓详情"],
  ];
  return options.map(([value, label]) =>
    `<option value="${attr(value)}" ${current === value ? "selected" : ""}>${escapeHtml(label)}</option>`
  ).join("");
}

function renderSHFilterForm(data) {
  const q = data.Query || {};
  return `<div class="card mb-3">
    <div class="card-body">
      <form class="row g-3 align-items-end" method="get" action="/fetch-site-sehuatang-sh-list">
        <input type="hidden" name="ps" value="${attr(q.PageSize || "")}">
        ${q.Sort ? `<input type="hidden" name="sort" value="${attr(q.Sort)}">` : ""}
        ${q.Order ? `<input type="hidden" name="order" value="${attr(q.Order)}">` : ""}
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">关键词</label>
          <input class="form-control" type="text" name="keyword" value="${attr(q.Keyword || "")}" placeholder="标题 / 正文 / thread_tid">
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">thread_tid</label>
          <input class="form-control" type="text" name="thread_tid" value="${attr(q.ThreadTID || "")}">
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">作者</label>
          <input class="form-control" type="text" name="author_name" value="${attr(q.AuthorName || "")}">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">info_hash</label>
          <input class="form-control" type="text" name="info_hash" value="${attr(q.InfoHash || "")}">
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">来源版块</label>
          <select class="form-select" name="source_forum_id">${renderSHSourceForumOptions(q)}</select>
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">详情状态</label>
          <select class="form-select" name="detail_status">${renderSHDetailStatusOptions(q)}</select>
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">magnet</label>
          <select class="form-select" name="has_magnet">${renderSHHasMagnetOptions(q)}</select>
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">发帖时间开始</label>
          <input class="form-control" type="text" name="post_time_from" value="${attr(q.PostTimeFrom || "")}" placeholder="YYYY-MM-DD">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">发帖时间结束</label>
          <input class="form-control" type="text" name="post_time_to" value="${attr(q.PostTimeTo || "")}" placeholder="YYYY-MM-DD">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">最近抓到开始</label>
          <input class="form-control" type="text" name="last_seen_time_from" value="${attr(q.LastSeenTimeFrom || "")}" placeholder="YYYY-MM-DD">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">最近抓到结束</label>
          <input class="form-control" type="text" name="last_seen_time_to" value="${attr(q.LastSeenTimeTo || "")}" placeholder="YYYY-MM-DD">
        </div>
        <div class="col-12 d-flex gap-2">
          <button class="btn btn-primary" type="submit">筛选</button>
          <a class="btn btn-outline-secondary" href="/fetch-site-sehuatang-sh-list">清空</a>
        </div>
      </form>
    </div>
  </div>`;
}

function renderSHThreadRow(row, index) {
  const title = text(row.ThreadTitle, "");
  const titleHTML = row.ThreadURL
    ? `<a class="fetch-site-title-ellipsis" href="${attr(row.ThreadURL)}" target="_blank" rel="noopener noreferrer" title="${attr(title)}">${escapeHtml(title)}</a>`
    : `<span class="fetch-site-title-ellipsis" title="${attr(title)}">${escapeHtml(title)}</span>`;
  const magnetCount = lineCount(row.MagnetText);
  const hashCount = lineCount(row.InfoHashText);
  return `<tr>
    <td class="text-nowrap">${escapeHtml(row.ThreadTID || "")}</td>
    <td>${titleHTML}</td>
    <td class="text-nowrap">${escapeHtml(text(row.AuthorName))}</td>
    <td class="text-nowrap">${escapeHtml(formatUnix(row.PostTime))}</td>
    <td class="text-nowrap">${escapeHtml(row.SourceForumID || "")}</td>
    <td class="text-nowrap">${escapeHtml(row.DetailStatus || "")}</td>
    <td class="text-nowrap">${escapeHtml(magnetCount)}</td>
    <td class="text-nowrap">${escapeHtml(hashCount)}</td>
    <td class="text-nowrap">${escapeHtml(formatUnix(row.LastSeenTime))}</td>
    <td class="text-nowrap">${escapeHtml(formatUnix(row.CreatedOn))}</td>
    <td class="text-nowrap">${escapeHtml(formatUnix(row.UpdatedOn))}</td>
    <td class="text-nowrap">
      <div class="btn-group btn-group-sm" role="group">
        <button type="button" class="btn btn-outline-primary js-sh-view" data-index="${attr(index)}">内容</button>
        <a class="btn btn-outline-primary" href="/fetch-site-sehuatang-sh-detail/${attr(encodeURIComponent(row.ThreadTID || ""))}">详情</a>
        <button type="button" class="btn btn-outline-secondary js-sh-copy" data-index="${attr(index)}" data-field="MagnetText" ${magnetCount ? "" : "disabled"}>magnet</button>
        <button type="button" class="btn btn-outline-secondary js-sh-copy" data-index="${attr(index)}" data-field="InfoHashText" ${hashCount ? "" : "disabled"}>hash</button>
      </div>
    </td>
  </tr>`;
}

function renderSHThreadTable(data) {
  const rows = Array.isArray(data.Rows) ? data.Rows : [];
  const sort = data.SortQuery || {};
  return `<div class="card mb-3">
    <div class="card-body">
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0">
          <thead>
            <tr>
              <th>${sortLabel(sort.ByThreadTID, "thread_tid")}</th>
              <th>帖子标题</th>
              <th>作者</th>
              <th>${sortLabel(sort.ByPostTime, "发布时间")}</th>
              <th>版块</th>
              <th>详情状态</th>
              <th>magnet</th>
              <th>hash</th>
              <th>${sortLabel(sort.ByLastSeenTime, "最近抓到")}</th>
              <th>${sortLabel(sort.ByCreatedOn, "入库时间")}</th>
              <th>${sortLabel(sort.ByUpdatedOn, "更新时间")}</th>
              <th class="text-nowrap">操作</th>
            </tr>
          </thead>
          <tbody>
            ${rows.length
              ? rows.map(renderSHThreadRow).join("")
              : `<tr><td colspan="12"><div class="fetch-site-empty-state">当前筛选下没有 Sehuatang SH 帖子数据。</div></td></tr>`}
          </tbody>
        </table>
      </div>
    </div>
  </div>`;
}

function renderSHRuns(data) {
  const runs = Array.isArray(data.RecentRuns) ? data.RecentRuns : [];
  return `<div class="card">
    <div class="card-body">
      <h5 class="card-title mb-3">最近抓取记录</h5>
      <div class="table-responsive">
        <table class="table table-sm align-middle mb-0">
          <thead>
            <tr>
              <th>run_id</th>
              <th>模式</th>
              <th>关键词</th>
              <th>版块</th>
              <th>状态</th>
              <th>发现</th>
              <th>命中</th>
              <th>跳过版块</th>
              <th>跳过已存在</th>
              <th>入库</th>
              <th>失败</th>
              <th>开始</th>
              <th>结束</th>
              <th>错误</th>
            </tr>
          </thead>
          <tbody>
            ${runs.length ? runs.map((run) => `<tr>
              <td>${escapeHtml(run.Id || "")}</td>
              <td>${escapeHtml(run.FetchMode || "")}</td>
              <td>${escapeHtml(run.Keyword || "")}</td>
              <td>${escapeHtml(run.TargetForumIDs || "")}</td>
              <td><span class="badge fetch-site-badge">${escapeHtml(run.Status || "")}</span></td>
              <td>${escapeHtml(valueOf(run.SeenCount, 0))}</td>
              <td>${escapeHtml(valueOf(run.MatchedCount, 0))}</td>
              <td>${escapeHtml(valueOf(run.SkippedForumCount, 0))}</td>
              <td>${escapeHtml(valueOf(run.SkippedExistingCount, 0))}</td>
              <td>${escapeHtml(valueOf(run.FetchedCount, 0))}</td>
              <td>${escapeHtml(valueOf(run.FailedCount, 0))}</td>
              <td class="text-nowrap">${escapeHtml(formatUnix(run.StartedOn))}</td>
              <td class="text-nowrap">${escapeHtml(formatUnix(run.FinishedOn))}</td>
              <td>${run.LastError ? `<details><summary>错误</summary><div class="mt-2">${escapeHtml(run.LastError)}</div></details>` : "-"}</td>
            </tr>`).join("") : `<tr><td colspan="14"><div class="fetch-site-empty-state">暂无抓取记录。</div></td></tr>`}
          </tbody>
        </table>
      </div>
    </div>
  </div>`;
}

function renderSHDetailModal() {
  return `<div class="modal fade" id="shThreadDetailModal" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-xl modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" id="shThreadDetailTitle">帖子内容</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="关闭"></button>
        </div>
        <div class="modal-body">
          <div class="mb-3 small text-muted" id="shThreadDetailMeta"></div>
          <h6>正文</h6>
          <pre class="bg-light border rounded p-3 small" id="shThreadDetailContent"></pre>
          <h6>magnet</h6>
          <pre class="bg-light border rounded p-3 small" id="shThreadDetailMagnet"></pre>
          <h6>info_hash</h6>
          <pre class="bg-light border rounded p-3 small mb-0" id="shThreadDetailHash"></pre>
        </div>
      </div>
    </div>
  </div>`;
}

function renderSHPage(data) {
  shThreadRows = Array.isArray(data.Rows) ? data.Rows : [];
  return `<div id="sehuatangSHListPage">
    <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
      <div>
        <h4 class="mb-1">${escapeHtml(data.PageTitle || "Sehuatang SH 帖子列表")}</h4>
        <p class="text-muted mb-0">${escapeHtml(data.PageNote || "")}</p>
      </div>
      <div class="d-flex flex-wrap gap-2">
        <a class="btn btn-outline-primary btn-sm" href="/fetch-site-sehuatang-sh-list">SehuatangSH 列表</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.OldListURL || "/fetch-site-sehuatang-list")}">旧磁力列表</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.TaskPageURL || "/triggers/fetch-sehuatang-sh-thread-list")}">列表抓取</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.DetailTaskPageURL || "/triggers/fetch-sehuatang-sh-thread-detail")}">详情补全</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(data.TasksPageURL || "/crawler/tasks")}">任务列表</a>
      </div>
    </div>
    <div id="sehuatangSHListMsg" class="alert d-none mb-3" role="alert"></div>
    <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
      <div class="fetch-site-summary-card">
        <span class="fetch-site-summary-label">SH 帖子总数</span>
        <span class="fetch-site-summary-value">${escapeHtml(valueOf(data.Total, 0))}</span>
      </div>
      ${shThreadRows.length ? renderPageLinks(data.PageInfo, "SehuatangSH") : ""}
    </div>
    ${renderSHFilterForm(data)}
    ${renderSHThreadTable(data)}
    ${renderSHRuns(data)}
  </div>
  ${renderSHDetailModal()}`;
}

function showSHMessage(message, level) {
  const box = document.getElementById("sehuatangSHListMsg");
  if (!box) {
    return;
  }
  box.className = `alert alert-${level || "info"} mb-3`;
  box.textContent = message;
  window.setTimeout(() => {
    box.classList.add("d-none");
  }, 2500);
}

async function copyToClipboard(textValue) {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(textValue);
    return;
  }
  const textarea = document.createElement("textarea");
  textarea.value = textValue;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand("copy");
  textarea.remove();
}

function initSHPage() {
  const page = document.getElementById("sehuatangSHListPage");
  if (!page) {
    return;
  }
  const modalEl = document.getElementById("shThreadDetailModal");
  const modal = modalEl && window.bootstrap && window.bootstrap.Modal
    ? new window.bootstrap.Modal(modalEl)
    : null;
  page.addEventListener("click", async (event) => {
    const copyButton = event.target.closest(".js-sh-copy");
    if (copyButton) {
      const index = Number(copyButton.getAttribute("data-index") || -1);
      const field = copyButton.getAttribute("data-field") || "";
      const row = shThreadRows[index] || {};
      const value = String(row[field] || "").trim();
      if (!value) {
        showSHMessage("没有可复制的内容", "warning");
        return;
      }
      try {
        await copyToClipboard(value);
        showSHMessage("已复制", "success");
      } catch (err) {
        showSHMessage(`复制失败：${err.message || err}`, "danger");
      }
      return;
    }
    const viewButton = event.target.closest(".js-sh-view");
    if (viewButton) {
      const index = Number(viewButton.getAttribute("data-index") || -1);
      const row = shThreadRows[index] || {};
      document.getElementById("shThreadDetailTitle").textContent = row.ThreadTitle || "帖子内容";
      document.getElementById("shThreadDetailMeta").textContent = [
        row.ThreadTID ? `tid=${row.ThreadTID}` : "",
        row.AuthorName ? `作者=${row.AuthorName}` : "",
        row.PostTime ? `发布时间=${formatUnix(row.PostTime)}` : "",
        row.ThreadURL ? `原帖=${row.ThreadURL}` : "",
      ].filter(Boolean).join(" | ");
      document.getElementById("shThreadDetailContent").textContent = row.RawContentText || "";
      document.getElementById("shThreadDetailMagnet").textContent = row.MagnetText || "";
      document.getElementById("shThreadDetailHash").textContent = row.InfoHashText || "";
      if (modal) {
        modal.show();
      }
    }
  });
}

async function main() {
  try {
    syncHeaderNav(window.location.pathname);
    const data = await fetchJSON(apiPathForCurrentPage() + window.location.search);
    document.title = `${data.Title || data.PageTitle || "FetchSite List"} - Rudy-GC`;
    if (data.ViewType === "fetch-site-sehuatang-sh-list") {
      root.innerHTML = renderSHPage(data);
      initSHPage();
      return;
    }
    if (data.ViewType === "fetch-site-sehuatang-list") {
      root.innerHTML = renderSehuatangPage(data);
      loadScriptOnce("/static/customer/sehuatang_magnet_list.js", "legacy-sehuatang-magnet-list-js");
      return;
    }
    const site = data.ViewType === "fetch-site-sukebei-list" ? "sukebei" : "javbus";
    root.innerHTML = renderFetchSitePage(data, site);
    loadScriptOnce("/static/customer/fetch_site_list.js", "legacy-fetch-site-list-js");
  } catch (err) {
    root.innerHTML = `<div class="alert alert-danger">加载失败：${escapeHtml(err.message)}</div>`;
  }
}

main();
