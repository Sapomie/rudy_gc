const apiPathByPage = new Map([
  ["/medias", "/api/medias"],
  ["/crawl-records", "/api/crawl-records"],
  ["/w-agg-events", "/api/w-agg-events"],
  ["/torrent-albums", "/api/torrent-albums"],
  ["/movie-albums", "/api/movie-albums"],
  ["/d-seeds", "/api/d-seeds"],
  ["/casts", "/api/casts"],
  ["/e-items", "/api/e-items"]
]);

const root = document.getElementById("page-root");
const msgArea = document.getElementById("msgArea");

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

function get(row, ...keys) {
  for (const key of keys) {
    if (row && Object.prototype.hasOwnProperty.call(row, key) && row[key] !== null && row[key] !== undefined) {
      return row[key];
    }
  }
  return "";
}

function formatUnix(value, withTime = false) {
  const n = Number(value || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  if (!withTime) return `${yyyy}-${mm}-${dd}`;
  const hh = String(d.getHours()).padStart(2, "0");
  const mi = String(d.getMinutes()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}`;
}

function formatUnixMinute(value) {
  return formatUnix(value, true);
}

function formatPositive(value) {
  const n = Number(value || 0);
  return n > 0 ? escapeHtml(value) : "-";
}

function formatAgeFromBirth(value) {
  const n = Number(value || 0);
  if (n <= 0) return "-";
  const birthday = new Date(n * 1000);
  const now = new Date();
  let age = now.getFullYear() - birthday.getFullYear();
  if (now.getMonth() < birthday.getMonth() || (now.getMonth() === birthday.getMonth() && now.getDate() < birthday.getDate())) {
    age -= 1;
  }
  return age >= 0 ? String(age) : "-";
}

function formatHeightCm(value) {
  const n = Number(value || 0);
  return n > 0 ? `${escapeHtml(value)} cm` : "-";
}

function formatRank(value) {
  const n = Number(value || 0);
  return n > 0 ? `#${escapeHtml(value)}` : "-";
}

function showMessage(message, type = "success") {
  msgArea.className = `alert alert-${type} fade show small py-2 px-3`;
  msgArea.textContent = message;
  msgArea.style.display = "block";
  window.setTimeout(() => {
    msgArea.style.display = "none";
  }, 2600);
}

function apiPathForCurrentPage() {
  return apiPathByPage.get(window.location.pathname) || "/api/medias";
}

async function fetchJSON(path, options) {
  const resp = await fetch(path, options);
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

function renderPagination(pageInfo) {
  const links = pageInfo && pageInfo.Links || [];
  if (!links.length) return "";
  const items = links.map((item) => {
    if (item.Ellipsis) {
      return `<li class="page-item disabled"><span class="page-link">...</span></li>`;
    }
    return `<li class="page-item ${item.Active ? "active" : ""} ${item.Disabled ? "disabled" : ""}">
      <a class="page-link" href="${attr(item.Href || "#")}" ${item.Disabled ? 'tabindex="-1" aria-disabled="true"' : ""}>${escapeHtml(item.Label)}</a>
    </li>`;
  }).join("");
  return `<nav aria-label="Page navigation"><ul class="pagination pagination-sm mb-0">${items}</ul></nav>`;
}

function renderPageHeader(data, subtitle) {
  return `<div class="d-flex align-items-center justify-content-between mb-3">
    <div>
      <h4 class="mb-1">${escapeHtml(data.PageTitle || data.Title || "")}</h4>
      ${subtitle || data.PageNote ? `<p class="text-muted mb-0">${escapeHtml(subtitle || data.PageNote)}</p>` : ""}
    </div>
  </div>`;
}

function renderSummary(data, label) {
  return `<div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
    <div class="sc-page-summary">
      <span class="sc-page-summary-label">${escapeHtml(label)}</span>
      <span class="sc-page-summary-value">${escapeHtml(data.Total ?? 0)}</span>
    </div>
    <div class="ms-xl-auto">${renderPagination(data.PageInfo || data.pageInfo)}</div>
  </div>`;
}

function input(name, label, value, type = "text", attrs = "", placeholder = "") {
  const placeholderAttr = placeholder ? ` placeholder="${attr(placeholder)}"` : "";
  return `<div class="col-12 col-md-6 col-xl-3">
    <label class="form-label mb-1">${escapeHtml(label)}</label>
    <input class="form-control" type="${attr(type)}" name="${attr(name)}" value="${attr(value || "")}"${placeholderAttr} ${attrs}>
  </div>`;
}

function rangeInput(label, minName, minValue, maxName, maxValue, type = "number", attrs = "", minPlaceholder = "Min", maxPlaceholder = "Max") {
  const minPlaceholderAttr = minPlaceholder ? ` placeholder="${attr(minPlaceholder)}"` : "";
  const maxPlaceholderAttr = maxPlaceholder ? ` placeholder="${attr(maxPlaceholder)}"` : "";
  return `<div class="col-12 col-md-6 col-xl-3">
    <label class="form-label mb-1">${escapeHtml(label)}</label>
    <div class="row g-2">
      <div class="col-6">
        <input class="form-control" type="${attr(type)}" name="${attr(minName)}" value="${attr(minValue || "")}"${minPlaceholderAttr} ${attrs}>
      </div>
      <div class="col-6">
        <input class="form-control" type="${attr(type)}" name="${attr(maxName)}" value="${attr(maxValue || "")}"${maxPlaceholderAttr} ${attrs}>
      </div>
    </div>
  </div>`;
}

function select(name, label, value, options) {
  const body = options.map(([v, text]) =>
    `<option value="${attr(v)}" ${String(value || "") === String(v) ? "selected" : ""}>${escapeHtml(text)}</option>`
  ).join("");
  return `<div class="col-12 col-md-6 col-xl-3">
    <label class="form-label mb-1">${escapeHtml(label)}</label>
    <select class="form-select" name="${attr(name)}">${body}</select>
  </div>`;
}

function hiddenSortFields(query) {
  return `${query?.PageSize ? `<input type="hidden" name="ps" value="${attr(query.PageSize)}">` : ""}
    ${query?.Sort ? `<input type="hidden" name="sort" value="${attr(query.Sort)}">` : ""}
    ${query?.Order ? `<input type="hidden" name="order" value="${attr(query.Order)}">` : ""}`;
}

function formCard(action, body, clearHref) {
  return `<div class="section-card card mb-3">
    <div class="card-body">
      <form class="row g-3 align-items-end" method="get" action="${attr(action)}">
        ${body}
        <div class="col-12 d-flex gap-2">
          <button class="btn btn-primary" type="submit">Apply</button>
          <a class="btn btn-outline-secondary" href="${attr(clearHref || action)}">Clear</a>
        </div>
      </form>
    </div>
  </div>`;
}

function linkSort(link, label) {
  const suffix = link?.Active ? (link.Desc ? " ↓" : " ↑") : "";
  return `<a class="sort-link ${link?.Active ? "active" : ""}" href="${attr(link?.Href || "#")}">${escapeHtml(label + suffix)}</a>`;
}

function castSortLink(link, label) {
  const suffix = link?.Active ? (link.Desc ? " ↓" : " ↑") : "";
  return `<a class="sc-sort-link sort-link ${link?.Active ? "active" : ""}" href="${attr(link?.Href || "#")}">${escapeHtml(label + suffix)}</a>`;
}

function wAggEventSortLink(link, label) {
  const indicator = link?.Active ? `<span class="w-agg-event-sort-indicator">${link.Desc ? "↓" : "↑"}</span>` : "";
  return `<a class="w-agg-event-sort-link ${link?.Active ? "is-active" : ""}" href="${attr(link?.Href || "#")}">${escapeHtml(label)}${indicator}</a>`;
}

function linkCell(href, text, fallback = "-") {
  const value = String(text || "");
  if (!value) return `<span class="text-muted">${escapeHtml(fallback)}</span>`;
  if (!href) return escapeHtml(value);
  return `<a href="${attr(href)}">${escapeHtml(value)}</a>`;
}

function table(headers, rows, emptyText = "暂无数据") {
  const header = headers.map((h) => `<th>${h}</th>`).join("");
  const body = rows.length
    ? rows.join("")
    : `<tr><td colspan="${headers.length}"><div class="sc-empty-state">${escapeHtml(emptyText)}</div></td></tr>`;
  return `<div class="section-card card">
    <div class="card-body">
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0 list-table">
          <thead><tr>${header}</tr></thead>
          <tbody>${body}</tbody>
        </table>
      </div>
    </div>
  </div>`;
}

function renderMedias(data) {
  const query = data.Query || {};
  const sort = data.FilmSortQuery || {};
  const form = formCard("/medias", `${hiddenSortFields(query)}
    ${input("keyword", "识别码搜索", query.Keyword, "text", "", "Movie Name")}
    ${rangeInput("大小范围(GB)", "size_min", query.SizeMin, "size_max", query.SizeMax, "number", 'min="0" step="0.01"')}
    ${rangeInput("高度范围", "height_min", query.HeightMin, "height_max", query.HeightMax, "number", 'min="0"')}
    ${rangeInput("时长范围(分钟)", "duration_min", query.DurationMin, "duration_max", query.DurationMax, "number", 'min="0"')}
    ${rangeInput("码率范围", "bit_rate_min", query.BitRateMin, "bit_rate_max", query.BitRateMax, "number", 'min="0"')}
    ${rangeInput("平均帧率范围", "frame_average_min", query.FrameAverageMin, "frame_average_max", query.FrameAverageMax, "number", 'min="0" step="0.01"')}
    ${rangeInput("SC 次数范围", "sc_times_min", query.ScTimesMin, "sc_times_max", query.ScTimesMax, "number", 'min="0"')}
    ${select("self_make", "自制", query.SelfMake, [["", "全部"], ["1", "否"], ["2", "是"]])}
    ${select("has_mask", "遮罩", query.HasMask, [["", "全部"], ["1", "有码"], ["2", "去码"], ["3", "无码"]])}
    ${rangeInput("最后 SC 时间", "last_sc_from", query.LastScFrom, "last_sc_to", query.LastScTo, "date", "", "", "")}
    ${rangeInput("下载时间", "birth_from", query.BirthFrom, "birth_to", query.BirthTo, "date", "", "", "")}
    ${rangeInput("发行时间", "releasing_from", query.ReleasingFrom, "releasing_to", query.ReleasingTo, "date", "", "", "")}`, "/medias");
  const rows = (data.Items || []).map((item) => {
    const casts = (item.Casts || []).map((cast) => linkCell(cast.Href, cast.Name)).join(" / ");
    return `<tr data-film-id="${attr(item.Id)}">
      <td>${linkCell(item.MovieHref, item.MovieName)}</td>
      <td class="wide">${casts || escapeHtml(item.CastName || "-")}</td>
      <td>${escapeHtml(item.SizeGB)}</td>
      <td><span data-role="film-height">${formatPositive(item.Height)}</span></td>
      <td><span data-role="film-duration">${formatPositive(item.DurationMinutes)}</span></td>
      <td><span data-role="film-bitrate">${formatPositive(item.BitRate)}</span></td>
      <td><span data-role="film-frame">${escapeHtml(item.FrameAverage)}</span></td>
      <td>${item.ShowPlayButton ? `<a class="btn btn-sm ${attr(item.PlayButtonClass || "btn-outline-primary")}" href="${attr(item.VideoURL)}">${escapeHtml(item.PlayButtonText || "MPlay")}</a>` : `<span class="text-muted">-</span>`}</td>
      <td>${escapeHtml(item.SelfMakeText)}</td>
      <td>${escapeHtml(item.HasMaskText)}</td>
      <td>${escapeHtml(item.ScTimes)}</td>
      <td>${formatUnix(item.LastScTime)}</td>
      <td>${formatUnix(item.BirthTime)}</td>
      <td>${formatUnix(item.ReleasingDate)}</td>
      <td>
        <button type="button" class="btn btn-sm btn-outline-primary js-film-probe" data-film-id="${attr(item.Id)}">scan_meta</button>
      </td>
    </tr>`;
  });
  return renderPageHeader(data, "按列表查看 w_media(source_type=2) 的基础数据，并复用 MovieType 缓存补充演员与播放信息。")
    + renderSummary(data, "Total Medias")
    + form
    + table([
      linkSort(sort.ByMovieName, "识别码"), "演员", linkSort(sort.BySize, "大小(GB)"), linkSort(sort.ByHeight, "高度"),
      linkSort(sort.ByDuration, "时长(分钟)"), linkSort(sort.ByBitRate, "码率"), linkSort(sort.ByFrameAverage, "平均帧率"),
      "播放", linkSort(sort.BySelfMake, "自制"), linkSort(sort.ByHasMask, "遮罩"), linkSort(sort.ByScTimes, "SC"),
      linkSort(sort.ByLastSc, "最后SC"), linkSort(sort.ByBirthTime, "下载时间"), linkSort(sort.ByReleasingDate, "发行时间"), "操作"
    ], rows);
}

function setRoleText(row, role, value) {
  const el = row && row.querySelector(`[data-role="${role}"]`);
  if (el) {
    el.textContent = value;
  }
}

function updateMediaProbeRow(row, result) {
  if (!row || !result) {
    return;
  }
  setRoleText(row, "film-height", Number(result.height || 0) > 0 ? String(result.height) : "-");
  setRoleText(row, "film-duration", Number(result.duration_minutes || 0) > 0 ? String(result.duration_minutes) : "-");
  setRoleText(row, "film-bitrate", Number(result.bit_rate || 0) > 0 ? String(result.bit_rate) : "-");
  setRoleText(row, "film-frame", result.frame_average || "-");
}

async function handleMediaProbe(btn) {
  const mediaId = btn && btn.getAttribute("data-film-id");
  if (!mediaId) {
    return;
  }
  const row = btn.closest("tr");
  const oldText = btn.textContent;
  btn.disabled = true;
  btn.textContent = "scanning...";
  try {
    const data = await fetchJSON(`/api/medias/${encodeURIComponent(mediaId)}/probe-meta`, {method: "POST"});
    if (!data || !data.ok) {
      throw new Error(data && data.error || "scan_meta failed");
    }
    updateMediaProbeRow(row, data.result);
    showMessage("scan_meta 完成");
  } catch (err) {
    showMessage(err && err.message ? err.message : "scan_meta 失败", "danger");
  } finally {
    btn.disabled = false;
    btn.textContent = oldText;
  }
}

function renderWAggEvents(data) {
  const q = data.Query || {};
  const links = data.HeaderLinks || {};
  const pageSizeOptions = [20, 50, 100, 200].map((size) =>
    `<option value="${size}" ${Number(q.PageSize || 0) === size ? "selected" : ""}>${size}</option>`
  ).join("");
  const pagination = renderPagination(data.PageInfo || data.pageInfo);
  const hasRows = Boolean((data.Rows || []).length);
  const summary = `<div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
    <div class="w-agg-event-summary">
      <span class="w-agg-event-summary-label">总行数</span>
      <span class="w-agg-event-summary-value">${escapeHtml(data.Total ?? 0)}</span>
    </div>
    ${hasRows && pagination ? `<div class="ms-xl-auto">${pagination}</div>` : ""}
  </div>`;
  const form = `<div class="section-card card mb-3">
    <div class="card-body">
      <form class="row g-3 align-items-end" method="get" action="/w-agg-events">
        ${q.Sort ? `<input type="hidden" name="sort" value="${attr(q.Sort)}">` : ""}
        ${q.Order ? `<input type="hidden" name="order" value="${attr(q.Order)}">` : ""}
        ${input("agg_key", "AggKey", q.AggKey, "text", "", "精确匹配")}
        ${input("flow_key", "FlowKey", q.FlowKey, "text", "", "精确匹配")}
        ${input("status", "Status", q.Status, "text", "", "精确匹配")}
        <div class="col-12 col-md-6 col-xl-1">
          <label class="form-label mb-1">每页</label>
          <select class="form-select" id="wAggEventPageSize" name="ps">${pageSizeOptions}</select>
        </div>
        <div class="col-12 col-xl-2 d-flex gap-2">
          <button class="btn btn-primary" type="submit">筛选</button>
          <a class="btn btn-outline-secondary" href="${attr(data.ClearHref || "/w-agg-events")}">清空</a>
        </div>
      </form>
    </div>
  </div>`;
  const rows = (data.Rows || []).map((row) => `<tr data-item-id="${attr(row.ID)}">
    <td class="text-nowrap">${formatUnix(row.StartedTime, true)}</td>
    <td class="text-nowrap w-agg-event-key">${escapeHtml(row.AggKey)}</td>
    <td class="text-nowrap w-agg-event-key">${escapeHtml(row.FlowKey)}</td>
    <td class="text-nowrap">${escapeHtml(row.ScopeCount)}</td>
    <td class="text-nowrap">${escapeHtml(row.BucketCount)}</td>
    <td class="text-nowrap">${escapeHtml(row.TopCount)}</td>
    <td class="text-nowrap">${formatUnix(row.FinishedTime, true)}</td>
    <td class="text-nowrap">${escapeHtml(row.DurationSecondsText)}</td>
  </tr>`);
  const tableBody = rows.length
    ? rows.join("")
    : `<tr><td colspan="8"><div class="w-agg-event-empty-state">当前筛选下没有 \`w_agg_event\` 数据。</div></td></tr>`;
  const tableHtml = `<div class="section-card card">
    <div class="card-body">
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0 w-agg-event-table">
          <thead><tr>
            <th class="text-nowrap">${wAggEventSortLink(links.StartedTime, "StartedTime")}</th>
            <th class="text-nowrap">${wAggEventSortLink(links.AggKey, "AggKey")}</th>
            <th class="text-nowrap">${wAggEventSortLink(links.FlowKey, "FlowKey")}</th>
            <th class="text-nowrap">${wAggEventSortLink(links.ScopeCount, "ScopeCount")}</th>
            <th class="text-nowrap">${wAggEventSortLink(links.BucketCount, "BucketCount")}</th>
            <th class="text-nowrap">${wAggEventSortLink(links.TopCount, "TopCount")}</th>
            <th class="text-nowrap">${wAggEventSortLink(links.FinishedTime, "FinishedTime")}</th>
            <th class="text-nowrap">${wAggEventSortLink(links.Duration, "Duration(秒)")}</th>
          </tr></thead>
          <tbody>${tableBody}</tbody>
        </table>
      </div>
    </div>
  </div>`;
  return `<div class="w-agg-event-list-page" id="wAggEventListPage">`
    + renderPageHeader(data)
    + summary
    + form
    + tableHtml
    + (hasRows && pagination ? `<div class="d-flex justify-content-end mt-3">${pagination}</div>` : "")
    + `</div>`;
}

function crawlRecordSortLink(link, label) {
  const indicator = link?.Active ? `<span class="crawl-record-sort-indicator">${link.Desc ? "↓" : "↑"}</span>` : "";
  return `<a class="crawl-record-sort-link ${link?.Active ? "is-active" : ""}" href="${attr(link?.Href || "#")}">${escapeHtml(label)}${indicator}</a>`;
}

function renderCrawlRecords(data) {
  const q = data.Query || {};
  const links = data.HeaderLinks || {};
  const rows = data.Rows || [];
  const pagination = renderPagination(data.PageInfo || data.pageInfo);
  const hasRows = Boolean(rows.length);
  const pageSizeOptions = [20, 50, 100, 200].map((size) =>
    `<option value="${size}" ${Number(q.PageSize || 0) === size ? "selected" : ""}>${size}</option>`
  ).join("");
  const typeLinks = (data.TypeLinks || []).map((link) =>
    `<a href="${attr(link.Href || "#")}" class="btn btn-sm ${link.Active ? "btn-primary" : "btn-outline-primary"}">${escapeHtml(link.Label || "")}</a>`
  ).join("");
  const summary = `<div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
    <div class="crawl-record-summary">
      <span class="crawl-record-summary-label">总行数</span>
      <span class="crawl-record-summary-value">${escapeHtml(data.Total ?? 0)}</span>
    </div>
    ${hasRows && pagination ? `<div class="ms-xl-auto">${pagination}</div>` : ""}
  </div>`;
  const form = `<div class="card mb-3">
    <div class="card-body">
      <form class="row g-3 align-items-end" method="get" action="/crawl-records">
        <input type="hidden" name="sort" value="${attr(q.Sort || "")}">
        <input type="hidden" name="order" value="${attr(q.Order || "")}">
        <input type="hidden" name="type" value="${attr(q.Type || "")}" id="crawlRecordTypeInput">

        <div class="col-12 col-xl-8">
          <label class="form-label mb-2">Type</label>
          <div class="d-flex flex-wrap gap-2 crawl-record-type-switch">${typeLinks}</div>
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">每页</label>
          <select class="form-select" id="crawlRecordPageSize" name="ps">${pageSizeOptions}</select>
        </div>
        <div class="col-12 col-xl-2 d-flex gap-2">
          <button class="btn btn-primary" type="submit">筛选</button>
          <a class="btn btn-outline-secondary" href="${attr(data.ClearHref || "/crawl-records")}">清空</a>
        </div>
      </form>
    </div>
  </div>`;
  const tableRows = rows.map((row) => `<tr>
    <td class="text-nowrap">${formatUnixMinute(row.StartedTime)}</td>
    <td class="text-nowrap crawl-record-type">${escapeHtml(row.Type)}</td>
    <td class="text-nowrap">${escapeHtml(row.DetailNumber)}</td>
    <td class="text-nowrap">${escapeHtml(row.DurationMinText)}</td>
    <td class="text-nowrap crawl-record-name">${escapeHtml(row.Name)}</td>
  </tr>`).join("");
  const tableBody = tableRows || `<tr>
    <td colspan="5"><div class="crawl-record-empty-state">当前筛选下没有 \`e_record\` 数据。</div></td>
  </tr>`;
  const tableHtml = `<div class="card">
    <div class="card-body">
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0 crawl-record-table">
          <thead><tr>
            <th class="text-nowrap">${crawlRecordSortLink(links.StartedTime, "StartedTime")}</th>
            <th class="text-nowrap">${crawlRecordSortLink(links.Type, "Type")}</th>
            <th class="text-nowrap">${crawlRecordSortLink(links.DetailNumber, "DetailNumber")}</th>
            <th class="text-nowrap">${crawlRecordSortLink(links.DurationMin, "Duration(min)")}</th>
            <th class="text-nowrap">${crawlRecordSortLink(links.Name, "Name")}</th>
          </tr></thead>
          <tbody>${tableBody}</tbody>
        </table>
      </div>
    </div>
  </div>`;
  return `<div class="crawl-record-list-page" id="crawlRecordListPage">`
    + renderPageHeader(data)
    + summary
    + form
    + tableHtml
    + (hasRows && pagination ? `<div class="d-flex justify-content-end mt-3">${pagination}</div>` : "")
    + `</div>`;
}

function albumOptions(albums, selectedName) {
  const options = (albums || []).map((album) => [album.Name, album.Name || album.ID]);
  if (!options.some(([name]) => String(name) === String(selectedName || ""))) {
    options.unshift([selectedName || "", selectedName || "默认"]);
  }
  return options;
}

function albumSortLink(link, label) {
  const suffix = link?.Active ? (link.Desc ? " ↓" : " ↑") : "";
  return `<a class="album-sort-link ${link?.Active ? "active" : ""}" href="${attr(link?.Href || "#")}">${escapeHtml(label + suffix)}</a>`;
}

function torrentAlbumSourceBadge(sourceType) {
  switch (String(sourceType || "")) {
    case "javbus_magnet":
      return '<span class="badge bg-primary-subtle text-primary">JavBus</span>';
    case "sukebei_torrent":
      return '<span class="badge bg-danger-subtle text-danger">Sukebei</span>';
    case "sehuatang_magnet":
      return '<span class="badge bg-warning-subtle text-warning">Sehuatang</span>';
    default:
      return '<span class="badge bg-secondary-subtle text-secondary">未知</span>';
  }
}

function renderTorrentAlbums(data) {
  const q = data.Query || {};
  const sort = data.SortQuery || {};
  const albums = data.Albums || [];
  const selectedAlbumName = data.SelectedAlbumName || q.AlbumName || "";
  const selectedAlbumID = Number(data.SelectedAlbumID || 0);
  const pagination = renderPagination(data.PageInfo || data.pageInfo);
  const hasRows = Boolean((data.Rows || []).length);
  const albumSelectOptions = albumOptions(albums, selectedAlbumName).map(([value, text]) =>
    `<option value="${attr(value)}" ${String(value || "") === String(selectedAlbumName || "") ? "selected" : ""}>${escapeHtml(text)}</option>`
  ).join("");
  const sourceOptions = [
    ["", "全部"],
    ["javbus_magnet", "JavBus 磁力"],
    ["sukebei_torrent", "Sukebei 种子"],
    ["sehuatang_magnet", "Sehuatang 磁力"]
  ].map(([value, text]) =>
    `<option value="${attr(value)}" ${String(q.SourceType || "") === value ? "selected" : ""}>${escapeHtml(text)}</option>`
  ).join("");
  const moveTargetOptions = albums
    .filter((album) => Number(album.ID || 0) > 0 && Number(album.ID || 0) !== selectedAlbumID)
    .map((album) => `<option value="${attr(album.ID)}">${escapeHtml(album.Name)}</option>`)
    .join("");
  const quickClass = (name, activeClass, outlineClass) => selectedAlbumName === name ? activeClass : outlineClass;
  const summary = `<div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
    <div class="sc-page-summary">
      <span class="sc-page-summary-label">相册</span>
      <span class="sc-page-summary-value">${escapeHtml(selectedAlbumName)}</span>
      <span class="sc-page-summary-label ms-3">总数</span>
      <span class="sc-page-summary-value">${escapeHtml(data.Total ?? 0)}</span>
    </div>
    <div class="d-flex align-items-center gap-2">
      <button type="button" class="btn btn-outline-primary btn-sm" id="btnOpenCreateTorrentAlbumModal">新增相册</button>
    </div>
    ${hasRows && pagination ? `<div class="ms-xl-auto">${pagination}</div>` : ""}
  </div>`;
  const form = `<div class="section-card card mb-3">
    <div class="card-body">
      <form class="row g-3 align-items-end" id="torrentAlbumsFilterForm" method="get" action="/torrent-albums">
        ${q.PageSize ? `<input type="hidden" name="ps" value="${attr(q.PageSize)}">` : ""}
        ${q.Sort ? `<input type="hidden" name="sort" value="${attr(q.Sort)}">` : ""}
        ${q.Order ? `<input type="hidden" name="order" value="${attr(q.Order)}">` : ""}
        <div class="col-12">
          <label class="form-label mb-1">快捷相册</label>
          <div class="d-flex flex-wrap gap-2 album-quick-filter-group">
            <button type="button" class="btn btn-sm ${quickClass("下载中", "btn-warning", "btn-outline-warning")} js-torrent-album-quick-filter" data-album-name="下载中">下载中</button>
            <button type="button" class="btn btn-sm ${quickClass("待下载", "btn-info", "btn-outline-info")} js-torrent-album-quick-filter" data-album-name="待下载">待下载</button>
            <button type="button" class="btn btn-sm ${quickClass("Media", "btn-primary", "btn-outline-primary")} js-torrent-album-quick-filter" data-album-name="Media">Media</button>
          </div>
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">相册</label>
          <select class="form-select" id="torrentAlbumsAlbumName" name="album_name">${albumSelectOptions}</select>
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">影片编号 / JavID</label>
          <input class="form-control" type="text" name="keyword" value="${attr(q.Keyword || "")}" placeholder="movie_name 或 movie_jav_id">
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">来源</label>
          <select class="form-select" name="source_type">${sourceOptions}</select>
        </div>
        <div class="col-12 col-md-6 col-xl-4">
          <label class="form-label mb-1">Info Hash</label>
          <input class="form-control" type="text" name="info_hash" value="${attr(q.InfoHash || "")}" placeholder="按 Info Hash 模糊筛选">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">发布日期开始</label>
          <input class="form-control" type="date" name="publish_from" value="${attr(q.PublishFrom || "")}">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">发布日期结束</label>
          <input class="form-control" type="date" name="publish_to" value="${attr(q.PublishTo || "")}">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">下载中时间开始</label>
          <input class="form-control" type="date" name="created_from" value="${attr(q.CreatedFrom || "")}">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">下载中时间结束</label>
          <input class="form-control" type="date" name="created_to" value="${attr(q.CreatedTo || "")}">
        </div>
        <div class="col-12 d-flex gap-2">
          <button class="btn btn-primary" type="submit">筛选</button>
          <a class="btn btn-outline-secondary" href="/torrent-albums">清空</a>
        </div>
      </form>
    </div>
  </div>`;
  const rows = (data.Rows || []).map((row) => `<tr>
    <td class="text-center">
      <input class="form-check-input js-torrent-album-item-check" type="checkbox" value="${attr(row.ID)}" aria-label="选择 ${attr(row.MovieName)}">
    </td>
    <td class="text-nowrap">${linkCell(`/movie/${encodeURIComponent(row.MovieName || "")}`, row.MovieName || row.MovieJavID)}</td>
    <td class="text-nowrap">${torrentAlbumSourceBadge(row.SourceType)}</td>
    <td class="album-hash-cell"><code class="album-hash">${escapeHtml(row.InfoHash)}</code></td>
    <td class="text-nowrap">${escapeHtml(row.Size || "-")}</td>
    <td class="text-nowrap">${formatUnix(row.PublishTime)}</td>
    <td class="text-nowrap">${formatUnix(row.CreatedOn, true)}</td>
    <td class="text-nowrap">
      <button type="button" class="btn btn-sm btn-outline-secondary js-copy-hash" data-hash="${attr(row.InfoHash)}">复制 Hash</button>
      <button type="button" class="btn btn-sm btn-outline-danger js-remove-torrent-album-item ms-1" data-item-id="${attr(row.ID)}">移除</button>
    </td>
  </tr>`);
  const tableBody = rows.length ? rows.join("") : `<tr><td colspan="8"><div class="sc-empty-state">当前筛选下没有相册资源数据。</div></td></tr>`;
  const tableHtml = `<div class="section-card card">
    <div class="card-body">
      <div class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <div class="small text-muted">已选 <span id="torrentAlbumBatchSelectedCount">0</span> 项</div>
        <div class="d-flex flex-wrap gap-2">
          <button type="button" class="btn btn-sm btn-outline-secondary" id="btnBatchCopyTorrentAlbumHashes" disabled>批量复制 Hash</button>
          <button type="button" class="btn btn-sm btn-outline-primary" id="btnBatchMoveTorrentAlbumItems" disabled>批量移动</button>
          <button type="button" class="btn btn-sm btn-outline-danger" id="btnBatchRemoveTorrentAlbumItems" disabled>批量移除</button>
        </div>
      </div>
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0">
          <thead><tr>
            <th class="text-center text-nowrap"><input class="form-check-input js-torrent-album-item-check-all" id="torrentAlbumItemCheckAll" type="checkbox" aria-label="全选当前页"></th>
            <th>${albumSortLink(sort.ByMovieName, "影片")}</th>
            <th>${albumSortLink(sort.BySourceType, "来源")}</th>
            <th>Info Hash</th>
            <th>大小</th>
            <th>${albumSortLink(sort.ByPublish, "发布日期")}</th>
            <th>${albumSortLink(sort.ByCreated, "下载中时间")}</th>
            <th class="text-nowrap">操作</th>
          </tr></thead>
          <tbody>${tableBody}</tbody>
        </table>
      </div>
    </div>
  </div>`;
  const modals = `<div class="modal fade" id="torrentAlbumBatchRemoveModal" tabindex="-1" aria-labelledby="torrentAlbumBatchRemoveModalLabel" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" id="torrentAlbumBatchRemoveModalLabel">批量移除确认</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body"><p class="mb-0">确认移除已选中的 <span id="torrentAlbumBatchRemoveCount">0</span> 条记录吗？</p></div>
        <div class="modal-footer">
          <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
          <button type="button" class="btn btn-danger" id="torrentAlbumBatchRemoveConfirmBtn">确认移除</button>
        </div>
      </div>
    </div>
  </div>
  <div class="modal fade" id="torrentAlbumBatchMoveModal" tabindex="-1" aria-labelledby="torrentAlbumBatchMoveModalLabel" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" id="torrentAlbumBatchMoveModalLabel">批量移动确认</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body">
          <label class="form-label mb-1" for="torrentAlbumBatchMoveTarget">目标相册</label>
          <select class="form-select" id="torrentAlbumBatchMoveTarget">
            <option value="">请选择目标相册</option>
            ${moveTargetOptions}
          </select>
          <p class="text-muted small mt-2 mb-0">将已选中的 <span id="torrentAlbumBatchMoveCount">0</span> 条记录移动到目标相册。</p>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
          <button type="button" class="btn btn-primary" id="torrentAlbumBatchMoveConfirmBtn">确认移动</button>
        </div>
      </div>
    </div>
  </div>
  <div class="modal fade" id="torrentAlbumCreateModal" tabindex="-1" aria-labelledby="torrentAlbumCreateModalLabel" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" id="torrentAlbumCreateModalLabel">新增相册</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body">
          <label class="form-label mb-1" for="torrentAlbumCreateName">相册名称</label>
          <input type="text" class="form-control" id="torrentAlbumCreateName" maxlength="128" placeholder="请输入相册名称">
          <p class="text-muted small mt-2 mb-0">创建成功后会自动切换到该相册并刷新列表。</p>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
          <button type="button" class="btn btn-primary" id="torrentAlbumCreateConfirmBtn">确认创建</button>
        </div>
      </div>
    </div>
  </div>`;
  return `<div id="torrentAlbumsPage" data-album-id="${attr(selectedAlbumID)}">`
    + renderPageHeader(data)
    + `<div id="torrentAlbumsMsg" class="alert d-none mb-3" role="alert"></div>`
    + summary
    + form
    + tableHtml
    + (hasRows && pagination ? `<div class="d-flex justify-content-end mt-3">${pagination}</div>` : "")
    + modals
    + `</div>`;
}

function renderMovieAlbums(data) {
  const q = data.Query || {};
  const sort = data.SortQuery || {};
  const albums = data.Albums || [];
  const selectedAlbumName = data.SelectedAlbumName || q.AlbumName || "";
  const selectedAlbumID = Number(data.SelectedAlbumID || 0);
  const pagination = renderPagination(data.PageInfo || data.pageInfo);
  const hasRows = Boolean((data.Rows || []).length);
  const albumSelectOptions = albumOptions(albums, selectedAlbumName).map(([value, text]) =>
    `<option value="${attr(value)}" ${String(value || "") === String(selectedAlbumName || "") ? "selected" : ""}>${escapeHtml(text)}</option>`
  ).join("");
  const quickClass = (name, activeClass, outlineClass) => selectedAlbumName === name ? activeClass : outlineClass;
  const canExecuteRemove = selectedAlbumName === "待删除";
  const canRemoveDownloaded = ["待下载电影", "v3", "vx"].includes(selectedAlbumName);
  const summary = `<div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
    <div class="sc-page-summary">
      <span class="sc-page-summary-label">相册</span>
      <span class="sc-page-summary-value">${escapeHtml(selectedAlbumName)}</span>
      <span class="sc-page-summary-label ms-3">总数</span>
      <span class="sc-page-summary-value">${escapeHtml(data.Total ?? 0)}</span>
    </div>
    <div class="d-flex align-items-center gap-2">
      <button type="button" class="btn btn-outline-primary btn-sm" id="btnOpenCreateMovieAlbumModal">新增电影相册</button>
    </div>
    ${hasRows && pagination ? `<div class="ms-xl-auto">${pagination}</div>` : ""}
  </div>`;
  const batchActions = canExecuteRemove || canRemoveDownloaded ? `<div class="d-flex flex-wrap gap-2">
      ${canExecuteRemove ? `<button type="button" class="btn btn-sm btn-danger" id="btnBatchExecuteMovieAlbumRemove" disabled>统一删除</button>` : ""}
      ${canRemoveDownloaded ? `<button type="button" class="btn btn-sm btn-outline-warning" id="btnRemoveDownloadedMovieAlbumItems">移除已下载条目</button>` : ""}
    </div>` : "";
  const form = `<div class="section-card card mb-3">
    <div class="card-body">
      <form class="row g-3 align-items-end" id="movieAlbumsFilterForm" method="get" action="/movie-albums">
        ${q.PageSize ? `<input type="hidden" name="ps" value="${attr(q.PageSize)}">` : ""}
        ${q.Sort ? `<input type="hidden" name="sort" value="${attr(q.Sort)}">` : ""}
        ${q.Order ? `<input type="hidden" name="order" value="${attr(q.Order)}">` : ""}
        <div class="col-12">
          <label class="form-label mb-1">快捷相册</label>
          <div class="d-flex flex-wrap gap-2 album-quick-filter-group">
            <button type="button" class="btn btn-sm ${quickClass("待下载电影", "btn-warning", "btn-outline-warning")} js-movie-album-quick-filter" data-album-name="待下载电影">待下载电影</button>
            <button type="button" class="btn btn-sm ${quickClass("待删除", "btn-danger", "btn-outline-danger")} js-movie-album-quick-filter" data-album-name="待删除">待删除</button>
            <button type="button" class="btn btn-sm ${quickClass("v3", "btn-primary", "btn-outline-primary")} js-movie-album-quick-filter" data-album-name="v3">v3</button>
            <button type="button" class="btn btn-sm ${quickClass("vx", "btn-primary", "btn-outline-primary")} js-movie-album-quick-filter" data-album-name="vx">vx</button>
          </div>
        </div>
        <div class="col-12 col-md-6 col-xl-4">
          <label class="form-label mb-1">相册</label>
          <select class="form-select" id="movieAlbumsAlbumName" name="album_name">${albumSelectOptions}</select>
        </div>
        <div class="col-12 col-md-6 col-xl-4">
          <label class="form-label mb-1">影片编号 / 标题</label>
          <input class="form-control" type="text" name="keyword" value="${attr(q.Keyword || "")}" placeholder="movie_name 或 movie_jav_id">
        </div>
        <div class="col-12 d-flex gap-2">
          <button class="btn btn-primary" type="submit">筛选</button>
          <a class="btn btn-outline-secondary" href="/movie-albums">清空</a>
        </div>
      </form>
    </div>
  </div>`;
  const rows = (data.Rows || []).map((row) => `<tr data-item-id="${attr(row.ID)}">
    <td class="text-center">
      <input class="form-check-input js-movie-album-item-check" type="checkbox" value="${attr(row.ID)}" aria-label="选择 ${attr(row.MovieName)}">
    </td>
    <td class="text-nowrap">${linkCell(`/movie/${encodeURIComponent(row.MovieName || row.MovieJavID || "")}`, row.MovieName || row.MovieJavID)}</td>
    <td class="text-nowrap">${formatUnix(row.CreatedOn, true)}</td>
    <td class="text-nowrap">
      <button type="button" class="btn btn-sm btn-outline-danger js-remove-movie-album-item" data-item-id="${attr(row.ID)}">移除</button>
    </td>
  </tr>`);
  const tableBody = rows.length ? rows.join("") : `<tr><td colspan="4"><div class="sc-empty-state">当前筛选下没有电影相册数据。</div></td></tr>`;
  const tableHtml = `<div class="section-card card">
    <div class="card-body">
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0">
          <thead><tr>
            <th class="text-center text-nowrap"><input class="form-check-input js-movie-album-item-check-all" id="movieAlbumItemCheckAll" type="checkbox" aria-label="全选当前页"></th>
            <th>${albumSortLink(sort.ByMovieName, "影片")}</th>
            <th>${albumSortLink(sort.ByCreated, "加入时间")}</th>
            <th class="text-nowrap">操作</th>
          </tr></thead>
          <tbody>${tableBody}</tbody>
        </table>
      </div>
    </div>
  </div>`;
  const modals = `<div class="modal fade" id="movieAlbumCreateModal" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">新增电影相册</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body">
          <label class="form-label mb-1" for="movieAlbumCreateName">相册名称</label>
          <input type="text" class="form-control" id="movieAlbumCreateName" maxlength="128" placeholder="请输入电影相册名称">
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
          <button type="button" class="btn btn-primary" id="movieAlbumCreateConfirmBtn">确认创建</button>
        </div>
      </div>
    </div>
  </div>
  <div class="modal fade" id="movieAlbumExecuteRemoveModal" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">统一删除确认</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body">
          确认执行已选中的 <span id="movieAlbumExecuteRemoveCount">0</span> 项吗？这会移动对应的 native w_media 到 \`001_process/005_removed\`。
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
          <button type="button" class="btn btn-danger" id="movieAlbumExecuteRemoveConfirmBtn">确认删除</button>
        </div>
      </div>
    </div>
  </div>
  <div class="modal fade" id="movieAlbumRemoveDownloadedModal" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">移除已下载条目确认</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body">
          <div class="mb-2">将要移除 <span class="fw-semibold" id="movieAlbumRemoveDownloadedCount">0</span> 条。</div>
          <div class="small text-muted mb-2">只会移除相册条目，不会删除媒体文件。</div>
          <div class="border rounded p-2" style="max-height: 280px; overflow-y: auto;">
            <div id="movieAlbumRemoveDownloadedEmpty" class="small text-muted">正在加载预览...</div>
            <div id="movieAlbumRemoveDownloadedList" class="list-group list-group-flush"></div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
          <button type="button" class="btn btn-warning" id="movieAlbumRemoveDownloadedConfirmBtn" disabled>确认移除</button>
        </div>
      </div>
    </div>
  </div>`;
  return `<div id="movieAlbumItemsPage" data-album-id="${attr(selectedAlbumID)}" data-album-name="${attr(selectedAlbumName)}">`
    + renderPageHeader(data)
    + `<div id="movieAlbumItemsMsg" class="alert d-none mb-3" role="alert"></div>`
    + summary
    + `<div class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <div class="small text-muted">已选 <span id="movieAlbumBatchSelectedCount">0</span> 项</div>
        ${batchActions}
      </div>`
    + form
    + tableHtml
    + (hasRows && pagination ? `<div class="d-flex justify-content-end mt-3">${pagination}</div>` : "")
    + modals
    + `</div>`;
}

function renderDSeeds(data) {
  const q = data.Query || {};
  const sort = data.DSeedSort || {};
  const items = data.Items || [];
  const pagination = renderPagination(data.PageInfo || data.pageInfo);
  const hasRows = items.length > 0;
  const pageSize = Number(q.PageSize || 0);
  const selectOptions = (value, options) => options.map(([optionValue, text]) =>
    `<option value="${attr(optionValue)}" ${String(value || "") === String(optionValue) ? "selected" : ""}>${escapeHtml(text)}</option>`
  ).join("");
  const dSeedSort = (link, label) => {
    const suffix = link?.Active ? (link.Desc ? " ↓" : " ↑") : "";
    return `<a class="d-seed-sort-link ${link?.Active ? "active" : ""}" href="${attr(link?.Href || "#")}">${escapeHtml(label + suffix)}</a>`;
  };
  const pageSizeOptions = [24, 50, 100].map((size) =>
    `<option value="${size}" ${pageSize === size ? "selected" : ""}>${size}</option>`
  ).join("");
  const header = `<div class="d-flex align-items-center justify-content-between mb-3">
    <div>
      <h4 class="mb-1">${escapeHtml(data.PageTitle || data.Title || "")}</h4>
      ${data.PageNote ? `<p class="text-muted mb-0">${escapeHtml(data.PageNote)}</p>` : ""}
    </div>
    <button type="button" class="btn btn-outline-primary btn-sm" id="btnOpenDSeedCreateModal">新增条目</button>
  </div>`;
  const summary = `<div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
    <div class="d-seed-page-summary">
      <span class="d-seed-page-summary-label">总数</span>
      <span class="d-seed-page-summary-value">${escapeHtml(data.Total ?? 0)}</span>
    </div>
    ${hasRows && pagination ? `<div class="ms-xl-auto">${pagination}</div>` : ""}
  </div>`;
  const form = `<div class="section-card card mb-3">
    <div class="card-body">
      <form class="row g-2 align-items-end" method="get" action="/d-seeds">
        ${q.Sort ? `<input type="hidden" name="sort" value="${attr(q.Sort)}">` : ""}
        ${q.Order ? `<input type="hidden" name="order" value="${attr(q.Order)}">` : ""}
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">查询名</label>
          <input class="form-control" type="text" name="name" value="${attr(q.Name || "")}" placeholder="支持模糊搜索">
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">启用状态</label>
          <select class="form-select" name="active">${selectOptions(q.Active, [["", "全部"], ["2", "启用"], ["1", "停用"]])}</select>
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">名称类型</label>
          <select class="form-select" name="name_type">${selectOptions(q.NameType, [["", "全部"], ["1", "Prefix"], ["2", "Label"]])}</select>
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">查询类型</label>
          <select class="form-select" name="search_type">${selectOptions(q.SearchType, [["", "全部"], ["1", "偏移量"], ["2", "起止页"]])}</select>
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">最后状态</label>
          <select class="form-select" name="last_status">${selectOptions(q.LastStatus, [["", "全部"], ["0", "空闲"], ["1", "成功"], ["2", "空页"], ["3", "错误"]])}</select>
        </div>
        <div class="col-12 col-md-6 col-xl-1">
          <label class="form-label mb-1">每页</label>
          <select class="form-select" name="ps">${pageSizeOptions}</select>
        </div>
        <div class="col-12 d-flex gap-2">
          <button class="btn btn-primary" type="submit">筛选</button>
          <a class="btn btn-outline-secondary" href="/d-seeds">清空</a>
        </div>
      </form>
    </div>
  </div>`;
  const rows = items.map((row) => `<tr>
    <td class="text-nowrap">${row.CardsHref ? `<a class="fw-semibold" href="${attr(row.CardsHref)}">${escapeHtml(row.Name)}</a>` : escapeHtml(row.Name)}</td>
    <td><span class="badge ${attr(row.ActiveClass)}">${escapeHtml(row.ActiveText)}</span></td>
    <td class="text-nowrap">${escapeHtml(row.NameTypeText)}</td>
    <td class="text-nowrap">${escapeHtml(row.SearchTypeText)}</td>
    <td class="text-nowrap">${escapeHtml(row.MovieTotal)}</td>
    <td class="text-nowrap">${row.MovieLatestReleasingMovieHref ? `<a href="${attr(row.MovieLatestReleasingMovieHref)}">${escapeHtml(row.MovieLatestReleasingMovieName)}</a>` : "-"}</td>
    <td class="text-nowrap">${formatUnix(row.MovieLatestReleasingDate)}</td>
    <td class="text-nowrap">${formatUnix(row.MovieLastAddedTime, true)}</td>
    <td class="text-nowrap">${escapeHtml(row.LastInsertCount)}</td>
    <td class="text-nowrap">${escapeHtml(row.PageNow)}</td>
    <td class="text-nowrap">${escapeHtml(row.Offset)} / ${escapeHtml(row.StartPage)}-${escapeHtml(row.EndPage)}</td>
    <td class="text-nowrap">${formatUnix(row.LastQueryTime, true)}</td>
    <td class="text-nowrap">
      <button type="button"
        class="btn btn-sm btn-outline-primary js-d-seed-edit"
        data-item-id="${attr(row.Id)}"
        data-name="${attr(row.Name)}"
        data-active="${attr(row.ActiveValue)}"
        data-search-type="${attr(row.SearchTypeValue)}"
        data-name-type="${attr(row.NameTypeValue)}"
        data-page-now="${attr(row.PageNow)}"
        data-offset="${attr(row.Offset)}"
        data-start-page="${attr(row.StartPage)}"
        data-end-page="${attr(row.EndPage)}"
        data-last-status="${attr(row.LastStatusValue)}"
        data-last-error="${attr(row.LastError)}">编辑</button>
    </td>
  </tr>`);
  const tableBody = rows.length ? rows.join("") : `<tr><td colspan="13"><div class="d-seed-empty-state">当前筛选下没有 d_seed 数据。</div></td></tr>`;
  const tableHtml = `<div class="section-card card">
    <div class="card-body">
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0">
          <thead><tr>
            <th>${dSeedSort(sort.ByName, "查询名")}</th>
            <th>${dSeedSort(sort.ByActive, "启用状态")}</th>
            <th>${dSeedSort(sort.ByNameType, "名称类型")}</th>
            <th>${dSeedSort(sort.BySearchType, "查询类型")}</th>
            <th class="text-nowrap">${dSeedSort(sort.ByMovieTotal, "电影总量")}</th>
            <th class="text-nowrap">${dSeedSort(sort.ByMovieLatestReleasingMovieName, "最新上映电影")}</th>
            <th class="text-nowrap">${dSeedSort(sort.ByMovieLatestReleasingDate, "最新上映日")}</th>
            <th class="text-nowrap">${dSeedSort(sort.ByMovieLastAddedTime, "上次新增")}</th>
            <th class="text-nowrap">${dSeedSort(sort.ByLastInsertCount, "本轮新增")}</th>
            <th class="text-nowrap">${dSeedSort(sort.ByPageNow, "当前页")}</th>
            <th class="text-nowrap">断点</th>
            <th class="text-nowrap">${dSeedSort(sort.ByLastQueryTime, "最后抓取")}</th>
            <th class="text-nowrap">操作</th>
          </tr></thead>
          <tbody>${tableBody}</tbody>
        </table>
      </div>
    </div>
  </div>`;
  const modal = `<div class="modal fade" id="dSeedEditModal" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered modal-lg">
      <div class="modal-content">
        <form id="dSeedEditForm">
          <div class="modal-header">
            <div>
              <h5 class="modal-title" id="dSeedEditModalTitle">新增 d_seed 条目</h5>
              <div class="small text-muted" id="dSeedEditModalMeta"></div>
            </div>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <input type="hidden" id="dSeedEditId">
            <div class="row g-3">
              <div class="col-12 col-md-6">
                <label class="form-label mb-1" for="dSeedEditName">查询名</label>
                <input class="form-control" id="dSeedEditName" type="text" maxlength="128">
              </div>
              <div class="col-12 col-md-3">
                <label class="form-label mb-1" for="dSeedEditActive">启用状态</label>
                <select class="form-select" id="dSeedEditActive">
                  <option value="2">启用</option>
                  <option value="1">停用</option>
                </select>
              </div>
              <div class="col-12 col-md-3">
                <label class="form-label mb-1" for="dSeedEditLastStatus">最后状态</label>
                <select class="form-select" id="dSeedEditLastStatus">
                  <option value="0">空闲</option>
                  <option value="1">成功</option>
                  <option value="2">空页</option>
                  <option value="3">错误</option>
                </select>
              </div>
              <div class="col-12 col-md-3">
                <label class="form-label mb-1" for="dSeedEditNameType">名称类型</label>
                <select class="form-select" id="dSeedEditNameType">
                  <option value="1">Prefix</option>
                  <option value="2">Label</option>
                </select>
              </div>
              <div class="col-12 col-md-3">
                <label class="form-label mb-1" for="dSeedEditSearchType">查询类型</label>
                <select class="form-select" id="dSeedEditSearchType">
                  <option value="1">偏移量</option>
                  <option value="2">起止页</option>
                </select>
              </div>
              <div class="col-12 col-md-3">
                <label class="form-label mb-1" for="dSeedEditPageNow">当前页</label>
                <input class="form-control" id="dSeedEditPageNow" type="number" min="0" step="1">
              </div>
              <div class="col-12 col-md-3" id="dSeedOffsetFieldWrap">
                <label class="form-label mb-1" for="dSeedEditOffset">偏移量</label>
                <input class="form-control" id="dSeedEditOffset" type="number" min="0" step="1">
              </div>
              <div class="col-12 col-md-3" id="dSeedStartPageFieldWrap">
                <label class="form-label mb-1" for="dSeedEditStartPage">起始页</label>
                <input class="form-control" id="dSeedEditStartPage" type="number" min="0" step="1">
              </div>
              <div class="col-12 col-md-3" id="dSeedEndPageFieldWrap">
                <label class="form-label mb-1" for="dSeedEditEndPage">结束页</label>
                <input class="form-control" id="dSeedEditEndPage" type="number" min="0" step="1">
              </div>
              <div class="col-12">
                <label class="form-label mb-1" for="dSeedEditLastError">最后错误</label>
                <textarea class="form-control" id="dSeedEditLastError" rows="3"></textarea>
              </div>
            </div>
            <div class="alert d-none mt-3 mb-0" id="dSeedEditMsg" role="alert"></div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
            <button type="submit" class="btn btn-primary" id="dSeedEditSubmit">保存</button>
          </div>
        </form>
      </div>
    </div>
  </div>`;
  return `<div id="dSeedListPage">`
    + header
    + `<div id="dSeedMsg" class="alert d-none mb-3" role="alert"></div>`
    + summary
    + form
    + tableHtml
    + (hasRows && pagination ? `<div class="d-flex justify-content-end mt-3">${pagination}</div>` : "")
    + modal
    + `</div>`;
}

function renderCasts(data) {
  const q = data.Query || {};
  const sort = data.CastSortQuery || {};
  const rows = (data.Items || []).map((row) => `<tr>
    <td><a class="sc-event-link" href="/cast/${attr(row.Id)}">${escapeHtml(row.Name || "")}</a></td>
    <td>${escapeHtml(row.Chinese || "-")}</td>
    <td>${formatAgeFromBirth(row.BirthDay)}</td>
    <td>${formatHeightCm(row.Height)}</td>
    <td>${escapeHtml(row.MovieNumber ?? 0)}</td>
    <td>${escapeHtml(row.OwnedWMediaNumber ?? 0)}</td>
    <td>${escapeHtml(row.ScTimes ?? 0)}</td>
    <td>${escapeHtml(row.ComeTimes ?? 0)}</td>
    <td>${formatUnix(row.LastScTime)}</td>
    <td>${formatRank(row.HighestRank)}</td>
  </tr>`);
  const hasRows = rows.length > 0;
  const pagination = renderPagination(data.PageInfo || data.pageInfo);
  const summary = `<div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
    <div class="sc-page-summary">
      <span class="sc-page-summary-label">Total Casts</span>
      <span class="sc-page-summary-value">${escapeHtml(data.Total ?? 0)}</span>
    </div>
    ${hasRows && pagination ? `<div class="ms-xl-auto">${pagination}</div>` : ""}
  </div>`;
  const form = `<div class="card mb-3">
    <div class="card-body">
      <form class="row g-3 align-items-end" method="get" action="/casts">
        ${hiddenSortFields(q)}
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1" for="casts-keyword">Name Search</label>
          <input id="casts-keyword" class="form-control" type="text" name="keyword" placeholder="Name / Chinese / Alias" value="${attr(q.Keyword || "")}">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">WOwned Range</label>
          <div class="row g-2">
            <div class="col-6">
              <input id="casts-owned-min" class="form-control" type="number" min="0" name="wowned_min" placeholder="Min" value="${attr(q.OwnedMin || "")}">
            </div>
            <div class="col-6">
              <input id="casts-owned-max" class="form-control" type="number" min="0" name="wowned_max" placeholder="Max" value="${attr(q.OwnedMax || "")}">
            </div>
          </div>
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">SC Times Range</label>
          <div class="row g-2">
            <div class="col-6">
              <input id="casts-sc-times-min" class="form-control" type="number" min="0" name="sc_times_min" placeholder="Min" value="${attr(q.ScTimesMin || "")}">
            </div>
            <div class="col-6">
              <input id="casts-sc-times-max" class="form-control" type="number" min="0" name="sc_times_max" placeholder="Max" value="${attr(q.ScTimesMax || "")}">
            </div>
          </div>
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">Come Times Range</label>
          <div class="row g-2">
            <div class="col-6">
              <input id="casts-come-times-min" class="form-control" type="number" min="0" name="come_times_min" placeholder="Min" value="${attr(q.ComeTimesMin || "")}">
            </div>
            <div class="col-6">
              <input id="casts-come-times-max" class="form-control" type="number" min="0" name="come_times_max" placeholder="Max" value="${attr(q.ComeTimesMax || "")}">
            </div>
          </div>
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">Last SC Range</label>
          <div class="row g-2">
            <div class="col-6">
              <input id="casts-last-sc-from" class="form-control" type="date" name="last_sc_from" value="${attr(q.LastScFrom || "")}">
            </div>
            <div class="col-6">
              <input id="casts-last-sc-to" class="form-control" type="date" name="last_sc_to" value="${attr(q.LastScTo || "")}">
            </div>
          </div>
        </div>
        <div class="col-12 d-flex gap-2">
          <button class="btn btn-primary" type="submit">Apply</button>
          <a class="btn btn-outline-secondary" href="/casts">Clear</a>
        </div>
      </form>
    </div>
  </div>`;
  const tableHtml = `<div class="card">
    <div class="card-body">
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0 list-table">
          <thead>
            <tr>
              <th>${castSortLink(sort.ByName, "Name")}</th>
              <th>${castSortLink(sort.ByChinese, "Chinese")}</th>
              <th>${castSortLink(sort.ByAge, "Age")}</th>
              <th>${castSortLink(sort.ByHeight, "Height")}</th>
              <th>${castSortLink(sort.ByMovieNumber, "Movies")}</th>
              <th>${castSortLink(sort.ByWOwned, "WOwned")}</th>
              <th>${castSortLink(sort.ByScTimes, "SC Times")}</th>
              <th>${castSortLink(sort.ByComeTimes, "Come Times")}</th>
              <th>${castSortLink(sort.ByLastSc, "Last SC")}</th>
              <th>${castSortLink(sort.ByHighestRank, "Highest Rank")}</th>
            </tr>
          </thead>
          <tbody>
            ${hasRows ? rows.join("") : `<tr><td colspan="10"><div class="sc-empty-state">还没有可展示的演员。</div></td></tr>`}
          </tbody>
        </table>
      </div>
    </div>
  </div>`;
  return renderPageHeader(data, "浏览演员统计，并进入详情页查看该演员的 SC 历史和代表影片。")
    + summary
    + form
    + tableHtml;
}

function renderEItems(data) {
  const q = data.Query || {};
  const sort = data.ItemSortQuery || {};
  const option = (value, label, current) => `<option value="${attr(value)}" ${String(current || "") === String(value) ? "selected" : ""}>${escapeHtml(label)}</option>`;
  const rows = (data.Items || []).map((row) => `<tr data-item-id="${attr(row.id)}">
    <td class="text-center">
      <input class="form-check-input js-item-check" type="checkbox" value="${attr(row.id)}" aria-label="选择 ${attr(row.jav_id)}">
    </td>
    <td class="text-nowrap">
      ${row.jav_url ? `<a href="${attr(row.jav_url)}" target="_blank" rel="noopener noreferrer">${escapeHtml(row.jav_id)}</a>` : escapeHtml(row.jav_id)}
    </td>
    <td><a href="/movie/${attr(row.name)}">${escapeHtml(row.name)}</a></td>
    <td class="js-item-has-detail">${escapeHtml(row.has_detail_text || "-")}</td>
    <td class="js-item-has-download-cover">${escapeHtml(row.has_download_cover_text || "-")}</td>
    <td class="js-item-has-chinese">${escapeHtml(row.has_chinese_text || "-")}</td>
    <td class="js-item-detail-need-scan">${escapeHtml(row.detail_need_scan_text || "-")}</td>
    <td class="text-nowrap">${formatUnix(row.detail_birth_time, true)}</td>
    <td class="text-nowrap">${formatUnix(row.last_query_detail_time, true)}</td>
    <td class="text-nowrap">
      <button type="button"
              class="btn btn-sm btn-outline-primary js-item-status-edit"
              data-item-id="${attr(row.id)}"
              data-item-jav-id="${attr(row.jav_id)}"
              data-item-name="${attr(row.name)}"
              data-has-detail="${attr(row.has_detail_value)}"
              data-has-download-cover="${attr(row.has_download_cover_value)}"
              data-has-chinese="${attr(row.has_chinese_value)}"
              data-detail-need-scan="${attr(row.detail_need_scan_value)}">编辑状态</button>
      <button type="button"
              class="btn btn-sm btn-outline-danger js-item-delete ms-1"
              data-item-id="${attr(row.id)}"
              data-item-jav-id="${attr(row.jav_id)}"
              data-item-name="${attr(row.name)}">删除</button>
    </td>
  </tr>`);
  const hasRows = rows.length > 0;
  const pagination = renderPagination(data.PageInfo || data.pageInfo);
  const summary = `<div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
    <div class="sc-page-summary">
      <span class="sc-page-summary-label">总数</span>
      <span class="sc-page-summary-value" id="itemTotalCount">${escapeHtml(data.Total ?? 0)}</span>
    </div>
    ${hasRows && pagination ? `<div class="ms-xl-auto">${pagination}</div>` : ""}
  </div>`;
  const form = `<div class="card mb-3">
    <div class="card-body">
      <form class="row g-2 align-items-end" method="get" action="/e-items">
        ${hiddenSortFields(q)}
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">JavLib 编号</label>
          <input class="form-control" type="text" name="jav_id" value="${attr(q.JavID || "")}" placeholder="JavLib 编号">
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">影片编号</label>
          <input class="form-control" type="text" name="name" value="${attr(q.Name || "")}" placeholder="影片编号">
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">详情状态</label>
          <select class="form-select" name="has_detail">
            ${option("", "全部", q.HasDetail)}
            ${option("none", "无详情", q.HasDetail)}
            ${option("ok", "已有详情", q.HasDetail)}
          </select>
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">封面状态</label>
          <select class="form-select" name="has_download_cover">
            ${option("", "全部", q.HasDownloadCover)}
            ${option("none", "无本地封面", q.HasDownloadCover)}
            ${option("ok", "已有本地封面", q.HasDownloadCover)}
            ${option("wrong", "错误封面链接", q.HasDownloadCover)}
          </select>
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">中文字幕</label>
          <select class="form-select" name="has_chinese">
            ${option("", "全部", q.HasChinese)}
            ${option("none", "无中文字幕", q.HasChinese)}
            ${option("ok", "有中文字幕", q.HasChinese)}
            ${option("error", "翻译错误", q.HasChinese)}
            ${option("sensitive", "含敏感词", q.HasChinese)}
          </select>
        </div>
        <div class="col-12 col-md-6 col-xl-2">
          <label class="form-label mb-1">详情扫描状态</label>
          <select class="form-select" name="detail_need_scan">
            ${option("", "全部", q.DetailNeedScan)}
            ${option("need", "需要扫描", q.DetailNeedScan)}
            ${option("no_need", "无需扫描", q.DetailNeedScan)}
            ${option("wrong", "内容错误", q.DetailNeedScan)}
          </select>
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">详情时间从</label>
          <input class="form-control" type="date" name="detail_birth_time_from" value="${attr(q.DetailBirthTimeFrom || "")}">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">详情时间到</label>
          <input class="form-control" type="date" name="detail_birth_time_to" value="${attr(q.DetailBirthTimeTo || "")}">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">最后查询时间从</label>
          <input class="form-control" type="date" name="last_query_detail_time_from" value="${attr(q.LastQueryDetailTimeFrom || "")}">
        </div>
        <div class="col-12 col-md-6 col-xl-3">
          <label class="form-label mb-1">最后查询时间到</label>
          <input class="form-control" type="date" name="last_query_detail_time_to" value="${attr(q.LastQueryDetailTimeTo || "")}">
        </div>
        <div class="col-12 d-flex gap-2">
          <button class="btn btn-primary" type="submit">筛选</button>
          <a class="btn btn-outline-secondary" href="/e-items">清空</a>
        </div>
      </form>
    </div>
  </div>`;
  const tableHtml = `<div class="card">
    <div class="card-body">
      <div class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <div class="small text-muted">已选 <span id="itemBatchSelectedCount">0</span> 项</div>
        <button type="button" class="btn btn-sm btn-outline-primary" id="btnBatchEditStatus" disabled>批量编辑状态</button>
      </div>
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0 list-table">
          <thead>
            <tr>
              <th class="text-center text-nowrap"><input class="form-check-input js-item-check-all" id="itemCheckAll" type="checkbox" aria-label="全选当前页"></th>
              <th>${castSortLink(sort.ByJavID, "JavLib 编号")}</th>
              <th>${castSortLink(sort.ByName, "影片编号")}</th>
              <th>${castSortLink(sort.ByHasDetail, "详情状态")}</th>
              <th>${castSortLink(sort.ByHasDownloadCover, "封面状态")}</th>
              <th>${castSortLink(sort.ByHasChinese, "中文字幕")}</th>
              <th>${castSortLink(sort.ByDetailNeedScan, "详情扫描状态")}</th>
              <th>${castSortLink(sort.ByDetailBirthTime, "详情时间")}</th>
              <th>${castSortLink(sort.ByLastQueryDetailTime, "最后查询时间")}</th>
              <th class="text-nowrap">操作</th>
            </tr>
          </thead>
          <tbody id="itemTableBody">
            ${hasRows ? rows.join("") : `<tr class="js-item-empty-row"><td colspan="10"><div class="sc-empty-state">当前筛选下没有 e_item 数据。</div></td></tr>`}
          </tbody>
        </table>
      </div>
    </div>
  </div>`;
  const singleModal = `<div class="modal fade" id="itemStatusModal" tabindex="-1" aria-labelledby="itemStatusModalLabel" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <form id="itemStatusForm">
          <input type="hidden" id="itemStatusItemId">
          <div class="modal-header">
            <div>
              <h5 class="modal-title" id="itemStatusModalLabel">编辑状态</h5>
              <div id="itemStatusMeta" class="small text-muted"></div>
            </div>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <div class="row g-3">
              <div class="col-12 col-md-6">
                <label for="itemStatusHasDetail" class="form-label">详情状态</label>
                <select id="itemStatusHasDetail" class="form-select" name="hasDetail">
                  <option value="none">无详情</option>
                  <option value="ok">已有详情</option>
                </select>
              </div>
              <div class="col-12 col-md-6">
                <label for="itemStatusHasDownloadCover" class="form-label">封面状态</label>
                <select id="itemStatusHasDownloadCover" class="form-select" name="hasDownloadCover">
                  <option value="none">无本地封面</option>
                  <option value="ok">已有本地封面</option>
                  <option value="wrong">错误封面链接</option>
                </select>
              </div>
              <div class="col-12 col-md-6">
                <label for="itemStatusHasChinese" class="form-label">中文字幕</label>
                <select id="itemStatusHasChinese" class="form-select" name="hasChinese">
                  <option value="none">无中文字幕</option>
                  <option value="ok">有中文字幕</option>
                  <option value="error">翻译错误</option>
                  <option value="sensitive">含敏感词</option>
                </select>
              </div>
              <div class="col-12 col-md-6">
                <label for="itemStatusDetailNeedScan" class="form-label">详情扫描状态</label>
                <select id="itemStatusDetailNeedScan" class="form-select" name="detailNeedScan">
                  <option value="need">需要扫描</option>
                  <option value="no_need">无需扫描</option>
                  <option value="wrong">内容错误</option>
                </select>
              </div>
            </div>
            <div id="itemStatusMsg" class="alert small py-2 px-3 mt-3 mb-0 d-none"></div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
            <button type="submit" class="btn btn-primary" id="itemStatusSubmit">保存</button>
          </div>
        </form>
      </div>
    </div>
  </div>`;
  const batchModal = `<div class="modal fade" id="itemBatchStatusModal" tabindex="-1" aria-labelledby="itemBatchStatusModalLabel" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <form id="itemBatchStatusForm">
          <div class="modal-header">
            <div>
              <h5 class="modal-title" id="itemBatchStatusModalLabel">批量编辑状态</h5>
              <div id="itemBatchStatusMeta" class="small text-muted"></div>
            </div>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <div class="row g-3">
              <div class="col-12 col-md-6">
                <label for="itemBatchHasDetail" class="form-label">详情状态</label>
                <select id="itemBatchHasDetail" class="form-select" name="hasDetail">
                  <option value="">不修改</option>
                  <option value="none">无详情</option>
                  <option value="ok">已有详情</option>
                </select>
              </div>
              <div class="col-12 col-md-6">
                <label for="itemBatchHasDownloadCover" class="form-label">封面状态</label>
                <select id="itemBatchHasDownloadCover" class="form-select" name="hasDownloadCover">
                  <option value="">不修改</option>
                  <option value="none">无本地封面</option>
                  <option value="ok">已有本地封面</option>
                  <option value="wrong">错误封面链接</option>
                </select>
              </div>
              <div class="col-12 col-md-6">
                <label for="itemBatchHasChinese" class="form-label">中文字幕</label>
                <select id="itemBatchHasChinese" class="form-select" name="hasChinese">
                  <option value="">不修改</option>
                  <option value="none">无中文字幕</option>
                  <option value="ok">有中文字幕</option>
                  <option value="error">翻译错误</option>
                  <option value="sensitive">含敏感词</option>
                </select>
              </div>
              <div class="col-12 col-md-6">
                <label for="itemBatchDetailNeedScan" class="form-label">详情扫描状态</label>
                <select id="itemBatchDetailNeedScan" class="form-select" name="detailNeedScan">
                  <option value="">不修改</option>
                  <option value="need">需要扫描</option>
                  <option value="no_need">无需扫描</option>
                  <option value="wrong">内容错误</option>
                </select>
              </div>
            </div>
            <div id="itemBatchStatusMsg" class="alert small py-2 px-3 mt-3 mb-0 d-none"></div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
            <button type="submit" class="btn btn-primary" id="itemBatchStatusSubmit">批量保存</button>
          </div>
        </form>
      </div>
    </div>
  </div>`;
  const deleteModal = `<div class="modal fade" id="itemDeleteModal" tabindex="-1" aria-labelledby="itemDeleteModalLabel" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <div>
            <h5 class="modal-title" id="itemDeleteModalLabel">删除电影</h5>
            <div id="itemDeleteMeta" class="small text-muted"></div>
          </div>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body">
          <p class="mb-0">会删除这部电影在所有相关表中的数据，并写入删除记录。</p>
          <div id="itemDeleteMsg" class="alert small py-2 px-3 mt-3 mb-0 d-none"></div>
          <input type="hidden" id="itemDeleteItemId">
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
          <button type="button" class="btn btn-danger" id="itemDeleteSubmit">确认删除</button>
        </div>
      </div>
    </div>
  </div>`;
  return renderPageHeader({Title: "EItem 列表"}, "浏览 e_item 列表，并按字段筛选条目。")
    + summary
    + form
    + tableHtml
    + singleModal
    + batchModal
    + deleteModal;
}

function bindDSeedEditor() {
  const msgEl = document.getElementById("dSeedMsg");
  const openCreateBtn = document.getElementById("btnOpenDSeedCreateModal");
  const modalEl = document.getElementById("dSeedEditModal");
  const form = document.getElementById("dSeedEditForm");
  if (!form) return;
  if (!window.bootstrap || !modalEl) return;

  const modal = window.bootstrap.Modal.getOrCreateInstance(modalEl);
  const idEl = document.getElementById("dSeedEditId");
  const titleEl = document.getElementById("dSeedEditModalTitle");
  const metaEl = document.getElementById("dSeedEditModalMeta");
  const formMsgEl = document.getElementById("dSeedEditMsg");
  const submitBtn = document.getElementById("dSeedEditSubmit");
  const nameEl = document.getElementById("dSeedEditName");
  const activeEl = document.getElementById("dSeedEditActive");
  const searchTypeEl = document.getElementById("dSeedEditSearchType");
  const nameTypeEl = document.getElementById("dSeedEditNameType");
  const pageNowEl = document.getElementById("dSeedEditPageNow");
  const offsetEl = document.getElementById("dSeedEditOffset");
  const startPageEl = document.getElementById("dSeedEditStartPage");
  const endPageEl = document.getElementById("dSeedEditEndPage");
  const offsetWrapEl = document.getElementById("dSeedOffsetFieldWrap");
  const startPageWrapEl = document.getElementById("dSeedStartPageFieldWrap");
  const endPageWrapEl = document.getElementById("dSeedEndPageFieldWrap");
  const lastStatusEl = document.getElementById("dSeedEditLastStatus");
  const lastErrorEl = document.getElementById("dSeedEditLastError");

  function showDSeedMessage(el, text, ok) {
    if (!el) return;
    el.textContent = String(text || "");
    el.className = `alert small py-2 px-3 mb-3 ${ok ? "alert-success" : "alert-danger"}`;
  }

  function hideDSeedMessage(el) {
    if (!el) return;
    el.textContent = "";
    el.className = "alert d-none mt-3 mb-0";
  }

  function setSubmitting(submitting) {
    if (!submitBtn) return;
    submitBtn.disabled = submitting;
    submitBtn.textContent = submitting ? "保存中..." : "保存";
  }

  function setValue(el, value) {
    if (el) {
      el.value = String(value == null ? "" : value);
    }
  }

  function setHidden(el, hidden) {
    if (!el) return;
    el.classList.toggle("d-none", Boolean(hidden));
  }

  function updateSearchTypeFields() {
    const searchType = String(searchTypeEl?.value || "1");
    const isOffset = searchType === "1";
    setHidden(offsetWrapEl, !isOffset);
    setHidden(startPageWrapEl, isOffset);
    setHidden(endPageWrapEl, isOffset);
  }

  function openCreateModal() {
    setValue(idEl, "");
    setValue(nameEl, "");
    setValue(activeEl, "2");
    setValue(searchTypeEl, "1");
    setValue(nameTypeEl, "1");
    setValue(pageNowEl, "0");
    setValue(offsetEl, "0");
    setValue(startPageEl, "0");
    setValue(endPageEl, "0");
    setValue(lastStatusEl, "0");
    setValue(lastErrorEl, "");
    if (titleEl) titleEl.textContent = "新增 d_seed 条目";
    if (metaEl) metaEl.textContent = "创建新的查询配置";
    hideDSeedMessage(formMsgEl);
    updateSearchTypeFields();
    setSubmitting(false);
    modal.show();
  }

  function openEditModal(button) {
    setValue(idEl, button.getAttribute("data-item-id"));
    setValue(nameEl, button.getAttribute("data-name"));
    setValue(activeEl, button.getAttribute("data-active"));
    setValue(searchTypeEl, button.getAttribute("data-search-type"));
    setValue(nameTypeEl, button.getAttribute("data-name-type"));
    setValue(pageNowEl, button.getAttribute("data-page-now"));
    setValue(offsetEl, button.getAttribute("data-offset"));
    setValue(startPageEl, button.getAttribute("data-start-page"));
    setValue(endPageEl, button.getAttribute("data-end-page"));
    setValue(lastStatusEl, button.getAttribute("data-last-status"));
    setValue(lastErrorEl, button.getAttribute("data-last-error"));
    if (titleEl) titleEl.textContent = "编辑 d_seed 条目";
    if (metaEl) metaEl.textContent = String(button.getAttribute("data-name") || "");
    hideDSeedMessage(formMsgEl);
    updateSearchTypeFields();
    setSubmitting(false);
    modal.show();
  }

  function buildPayload() {
    const searchType = Number(searchTypeEl?.value || 0);
    const isOffset = searchType === 1;
    return {
      name: String(nameEl?.value || "").trim(),
      active: Number(activeEl?.value || 0),
      search_type: searchType,
      name_type: Number(nameTypeEl?.value || 0),
      page_now: Number(pageNowEl?.value || 0),
      offset: isOffset ? Number(offsetEl?.value || 0) : 0,
      start_page: isOffset ? 0 : Number(startPageEl?.value || 0),
      end_page: isOffset ? 0 : Number(endPageEl?.value || 0),
      last_status: Number(lastStatusEl?.value || 0),
      last_error: String(lastErrorEl?.value || "").trim()
    };
  }

  openCreateBtn?.addEventListener("click", openCreateModal);
  searchTypeEl?.addEventListener("change", updateSearchTypeFields);
  document.addEventListener("click", (event) => {
    const editBtn = event.target.closest(".js-d-seed-edit");
    if (!editBtn) return;
    openEditModal(editBtn);
  });

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const id = String(idEl?.value || "").trim();
    const body = buildPayload();
    setSubmitting(true);
    hideDSeedMessage(formMsgEl);
    try {
      const path = id ? `/api/d-seeds/${encodeURIComponent(id)}` : "/api/d-seeds";
      const data = await fetchJSON(path, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(body)
      });
      showDSeedMessage(msgEl, data.message || "保存成功", true);
      modal.hide();
      window.setTimeout(() => window.location.reload(), 250);
    } catch (err) {
      showDSeedMessage(formMsgEl, err.message || "保存失败", false);
      setSubmitting(false);
    }
  });

  updateSearchTypeFields();
}

function bindTorrentAlbums() {
  const page = document.getElementById("torrentAlbumsPage");
  if (!page) return;

  const msgEl = document.getElementById("torrentAlbumsMsg");
  const checkAll = document.getElementById("torrentAlbumItemCheckAll");
  const filterForm = document.getElementById("torrentAlbumsFilterForm");
  const albumNameSelect = document.getElementById("torrentAlbumsAlbumName");
  const selectedCountEl = document.getElementById("torrentAlbumBatchSelectedCount");
  const batchCopyBtn = document.getElementById("btnBatchCopyTorrentAlbumHashes");
  const batchMoveBtn = document.getElementById("btnBatchMoveTorrentAlbumItems");
  const batchRemoveBtn = document.getElementById("btnBatchRemoveTorrentAlbumItems");
  const batchMoveModalEl = document.getElementById("torrentAlbumBatchMoveModal");
  const batchMoveTargetEl = document.getElementById("torrentAlbumBatchMoveTarget");
  const batchMoveCountEl = document.getElementById("torrentAlbumBatchMoveCount");
  const batchMoveConfirmBtn = document.getElementById("torrentAlbumBatchMoveConfirmBtn");
  const batchRemoveModalEl = document.getElementById("torrentAlbumBatchRemoveModal");
  const batchRemoveCountEl = document.getElementById("torrentAlbumBatchRemoveCount");
  const batchRemoveConfirmBtn = document.getElementById("torrentAlbumBatchRemoveConfirmBtn");
  const createBtn = document.getElementById("btnOpenCreateTorrentAlbumModal");
  const createModalEl = document.getElementById("torrentAlbumCreateModal");
  const createNameEl = document.getElementById("torrentAlbumCreateName");
  const createConfirmBtn = document.getElementById("torrentAlbumCreateConfirmBtn");
  const currentAlbumID = parsePositiveID(page.getAttribute("data-album-id"));
  const batchMoveModal = batchMoveModalEl && window.bootstrap?.Modal ? window.bootstrap.Modal.getOrCreateInstance(batchMoveModalEl) : null;
  const batchRemoveModal = batchRemoveModalEl && window.bootstrap?.Modal ? window.bootstrap.Modal.getOrCreateInstance(batchRemoveModalEl) : null;
  const createModal = createModalEl && window.bootstrap?.Modal ? window.bootstrap.Modal.getOrCreateInstance(createModalEl) : null;
  let pendingMoveIDs = [];
  let pendingRemoveIDs = [];

  function showAlbumMessage(message, ok = true) {
    if (!msgEl) return;
    msgEl.textContent = String(message || "");
    msgEl.className = `alert small py-2 px-3 mb-3 ${ok ? "alert-success" : "alert-danger"}`;
  }

  function parsePositiveID(value) {
    const id = Number.parseInt(String(value || "0"), 10);
    return Number.isInteger(id) && id > 0 ? id : 0;
  }

  function selectedChecks() {
    return Array.from(page.querySelectorAll(".js-torrent-album-item-check:checked"));
  }

  function selectedIDs() {
    return selectedChecks().map((checkbox) => parsePositiveID(checkbox.value)).filter((id) => id > 0);
  }

  function selectedHashes() {
    return selectedChecks().map((checkbox) => {
      const row = checkbox.closest("tr");
      const copyBtn = row && row.querySelector(".js-copy-hash");
      const code = row && row.querySelector(".album-hash");
      return String(copyBtn?.getAttribute("data-hash") || code?.textContent || "").trim();
    }).filter(Boolean);
  }

  function syncSelection() {
    const allChecks = Array.from(page.querySelectorAll(".js-torrent-album-item-check"));
    const ids = selectedIDs();
    if (selectedCountEl) selectedCountEl.textContent = String(ids.length);
    if (batchCopyBtn) batchCopyBtn.disabled = ids.length === 0;
    if (batchMoveBtn) batchMoveBtn.disabled = ids.length === 0 || currentAlbumID <= 0;
    if (batchRemoveBtn) batchRemoveBtn.disabled = ids.length === 0 || currentAlbumID <= 0;
    if (checkAll) {
      checkAll.checked = allChecks.length > 0 && ids.length === allChecks.length;
      checkAll.indeterminate = ids.length > 0 && ids.length < allChecks.length;
    }
  }

  function removeRowsByIDs(ids) {
    ids.forEach((id) => page.querySelector(`tr[data-item-id="${id}"]`)?.remove());
    syncSelection();
  }

  async function requestAlbumJSON(path, body) {
    const data = await fetchJSON(path, {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(body || {})
    });
    if (data && data.ok === false) {
      throw new Error(data.error || "请求失败");
    }
    return data;
  }

  async function copyText(value) {
    const text = String(value || "").trim();
    if (!text) throw new Error("缺少可复制内容");
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
      return;
    }
    const input = document.createElement("textarea");
    input.value = text;
    input.style.position = "fixed";
    input.style.left = "-9999px";
    document.body.appendChild(input);
    input.focus();
    input.select();
    const ok = document.execCommand("copy");
    input.remove();
    if (!ok) throw new Error("复制失败");
  }

  function upsertOption(selectEl, value, text) {
    if (!selectEl) return;
    const stringValue = String(value || "").trim();
    const stringText = String(text || "").trim();
    if (!stringValue || !stringText) return;
    const existing = Array.from(selectEl.options || []).find((option) => String(option.value || "").trim() === stringValue);
    if (existing) {
      existing.textContent = stringText;
      return;
    }
    const option = document.createElement("option");
    option.value = stringValue;
    option.textContent = stringText;
    selectEl.appendChild(option);
  }

  async function removeOne(itemID, button) {
    if (currentAlbumID <= 0) {
      showAlbumMessage("当前相册无效，无法移除", false);
      return;
    }
    if (button) button.disabled = true;
    try {
      const data = await requestAlbumJSON(`/api/albums/${currentAlbumID}/items/remove`, {item_id: itemID});
      if (data.removed) removeRowsByIDs([itemID]);
      showAlbumMessage(data.message || "移除完成", true);
    } catch (err) {
      showAlbumMessage(err.message || "移除失败", false);
    } finally {
      if (button) button.disabled = false;
    }
  }

  async function batchRemove(ids) {
    if (batchRemoveBtn) batchRemoveBtn.disabled = true;
    try {
      const data = await requestAlbumJSON(`/api/albums/${currentAlbumID}/items/batch-remove`, {item_ids: ids});
      const failed = new Set((data.failed_ids || []).map(parsePositiveID).filter((id) => id > 0));
      const removed = ids.filter((id) => !failed.has(id));
      removeRowsByIDs(removed);
      showAlbumMessage(failed.size ? `已移除 ${data.removed_count || removed.length} 条，失败 ${failed.size} 条` : `已批量移除 ${data.removed_count || removed.length} 条`, failed.size === 0);
    } catch (err) {
      showAlbumMessage(err.message || "批量移除失败", false);
    } finally {
      if (batchRemoveBtn) batchRemoveBtn.disabled = false;
    }
  }

  async function batchMove(ids, targetAlbumID) {
    if (batchMoveBtn) batchMoveBtn.disabled = true;
    try {
      const data = await requestAlbumJSON(`/api/albums/${currentAlbumID}/items/batch-move`, {item_ids: ids, target_album_id: targetAlbumID});
      const failed = new Set((data.failed_ids || []).map(parsePositiveID).filter((id) => id > 0));
      const moved = ids.filter((id) => !failed.has(id));
      removeRowsByIDs(moved);
      showAlbumMessage(failed.size ? `已移动 ${data.moved_count || moved.length} 条，失败 ${failed.size} 条` : `已批量移动 ${data.moved_count || moved.length} 条`, failed.size === 0);
    } catch (err) {
      showAlbumMessage(err.message || "批量移动失败", false);
    } finally {
      if (batchMoveBtn) batchMoveBtn.disabled = false;
    }
  }

  async function createAlbum(name) {
    const albumName = String(name || "").trim();
    if (!albumName) {
      showAlbumMessage("相册名称不能为空", false);
      return;
    }
    if (createConfirmBtn) createConfirmBtn.disabled = true;
    try {
      const data = await requestAlbumJSON("/api/albums", {name: albumName});
      upsertOption(albumNameSelect, data.album_name, data.album_name);
      upsertOption(batchMoveTargetEl, data.album_id, data.album_name);
      showAlbumMessage(data.message || "相册创建成功", true);
      createModal?.hide();
      if (albumNameSelect && filterForm) {
        albumNameSelect.value = data.album_name;
        filterForm.submit();
      }
    } catch (err) {
      showAlbumMessage(err.message || "新增相册失败", false);
      if (createConfirmBtn) createConfirmBtn.disabled = false;
    }
  }

  checkAll?.addEventListener("change", () => {
    page.querySelectorAll(".js-torrent-album-item-check").forEach((checkbox) => {
      checkbox.checked = checkAll.checked;
    });
    syncSelection();
  });
  batchCopyBtn?.addEventListener("click", async () => {
    try {
      const hashes = selectedHashes();
      if (!hashes.length) throw new Error("请先选择要复制的条目");
      await copyText(hashes.join("\n"));
      showAlbumMessage(`已复制 ${hashes.length} 条 Hash`, true);
    } catch (err) {
      showAlbumMessage(err.message || "复制失败", false);
    }
  });
  batchMoveBtn?.addEventListener("click", () => {
    pendingMoveIDs = selectedIDs();
    if (!pendingMoveIDs.length) {
      showAlbumMessage("请先选择要移动的条目", false);
      return;
    }
    if (batchMoveCountEl) batchMoveCountEl.textContent = String(pendingMoveIDs.length);
    if (batchMoveTargetEl) batchMoveTargetEl.value = "";
    batchMoveModal?.show();
  });
  batchRemoveBtn?.addEventListener("click", () => {
    pendingRemoveIDs = selectedIDs();
    if (!pendingRemoveIDs.length) {
      showAlbumMessage("请先选择要移除的条目", false);
      return;
    }
    if (batchRemoveCountEl) batchRemoveCountEl.textContent = String(pendingRemoveIDs.length);
    batchRemoveModal?.show();
  });
  batchMoveConfirmBtn?.addEventListener("click", () => {
    const targetAlbumID = parsePositiveID(batchMoveTargetEl?.value);
    if (targetAlbumID <= 0) {
      showAlbumMessage("请选择目标相册", false);
      return;
    }
    batchMoveConfirmBtn.disabled = true;
    batchMove(pendingMoveIDs.slice(), targetAlbumID);
    batchMoveModal?.hide();
  });
  batchRemoveConfirmBtn?.addEventListener("click", () => {
    batchRemoveConfirmBtn.disabled = true;
    batchRemove(pendingRemoveIDs.slice());
    batchRemoveModal?.hide();
  });
  batchMoveModalEl?.addEventListener("hidden.bs.modal", () => {
    pendingMoveIDs = [];
    if (batchMoveCountEl) batchMoveCountEl.textContent = "0";
    if (batchMoveConfirmBtn) batchMoveConfirmBtn.disabled = false;
  });
  batchRemoveModalEl?.addEventListener("hidden.bs.modal", () => {
    pendingRemoveIDs = [];
    if (batchRemoveCountEl) batchRemoveCountEl.textContent = "0";
    if (batchRemoveConfirmBtn) batchRemoveConfirmBtn.disabled = false;
  });
  createBtn?.addEventListener("click", () => {
    if (createNameEl) createNameEl.value = "";
    if (createConfirmBtn) createConfirmBtn.disabled = false;
    createModal?.show();
    window.setTimeout(() => createNameEl?.focus(), 50);
  });
  createConfirmBtn?.addEventListener("click", () => createAlbum(createNameEl?.value));
  createNameEl?.addEventListener("keydown", (event) => {
    if (event.key !== "Enter") return;
    event.preventDefault();
    createAlbum(createNameEl.value);
  });

  page.addEventListener("change", (event) => {
    if (event.target?.classList?.contains("js-torrent-album-item-check")) syncSelection();
  });
  page.addEventListener("click", async (event) => {
    const quickButton = event.target.closest(".js-torrent-album-quick-filter");
    if (quickButton) {
      const albumName = String(quickButton.getAttribute("data-album-name") || "").trim();
      if (albumName && albumNameSelect && filterForm) {
        albumNameSelect.value = albumName;
        filterForm.submit();
      }
      return;
    }
    const copyButton = event.target.closest(".js-copy-hash");
    if (copyButton) {
      try {
        await copyText(copyButton.getAttribute("data-hash"));
        showAlbumMessage("Info Hash 已复制", true);
      } catch (err) {
        showAlbumMessage(err.message || "复制失败", false);
      }
      return;
    }
    const removeButton = event.target.closest(".js-remove-torrent-album-item");
    if (removeButton) {
      removeOne(parsePositiveID(removeButton.getAttribute("data-item-id")), removeButton);
    }
  });

  syncSelection();
}

function bindMovieAlbums() {
  const page = document.getElementById("movieAlbumItemsPage");
  if (!page) return;

  const albumID = parsePositiveID(page.getAttribute("data-album-id"));
  const albumName = String(page.getAttribute("data-album-name") || "").trim();
  const msgEl = document.getElementById("movieAlbumItemsMsg");
  const filterForm = document.getElementById("movieAlbumsFilterForm");
  const albumNameSelect = document.getElementById("movieAlbumsAlbumName");
  const checkAll = document.getElementById("movieAlbumItemCheckAll");
  const selectedCountEl = document.getElementById("movieAlbumBatchSelectedCount");
  const batchExecuteBtn = document.getElementById("btnBatchExecuteMovieAlbumRemove");
  const executeModalEl = document.getElementById("movieAlbumExecuteRemoveModal");
  const executeCountEl = document.getElementById("movieAlbumExecuteRemoveCount");
  const executeConfirmBtn = document.getElementById("movieAlbumExecuteRemoveConfirmBtn");
  const createBtn = document.getElementById("btnOpenCreateMovieAlbumModal");
  const createModalEl = document.getElementById("movieAlbumCreateModal");
  const createNameEl = document.getElementById("movieAlbumCreateName");
  const createConfirmBtn = document.getElementById("movieAlbumCreateConfirmBtn");
  const removeDownloadedBtn = document.getElementById("btnRemoveDownloadedMovieAlbumItems");
  const removeDownloadedModalEl = document.getElementById("movieAlbumRemoveDownloadedModal");
  const removeDownloadedConfirmBtn = document.getElementById("movieAlbumRemoveDownloadedConfirmBtn");
  const removeDownloadedCountEl = document.getElementById("movieAlbumRemoveDownloadedCount");
  const removeDownloadedListEl = document.getElementById("movieAlbumRemoveDownloadedList");
  const removeDownloadedEmptyEl = document.getElementById("movieAlbumRemoveDownloadedEmpty");
  const executeModal = executeModalEl && window.bootstrap?.Modal ? window.bootstrap.Modal.getOrCreateInstance(executeModalEl) : null;
  const createModal = createModalEl && window.bootstrap?.Modal ? window.bootstrap.Modal.getOrCreateInstance(createModalEl) : null;
  const removeDownloadedModal = removeDownloadedModalEl && window.bootstrap?.Modal ? window.bootstrap.Modal.getOrCreateInstance(removeDownloadedModalEl) : null;
  let removeDownloadedPreviewItems = [];

  function parsePositiveID(value) {
    const id = Number.parseInt(String(value || "0"), 10);
    return Number.isInteger(id) && id > 0 ? id : 0;
  }

  function showMovieAlbumMessage(message, ok = true) {
    if (!msgEl) return;
    msgEl.textContent = String(message || "");
    msgEl.className = `alert small py-2 px-3 mb-3 ${ok ? "alert-success" : "alert-danger"}`;
  }

  async function requestMovieAlbumJSON(path, body) {
    const data = await fetchJSON(path, {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(body || {})
    });
    if (data && data.ok === false) {
      throw new Error(data.error || "请求失败");
    }
    return data;
  }

  function selectedIDs() {
    return Array.from(page.querySelectorAll(".js-movie-album-item-check:checked"))
      .map((checkbox) => parsePositiveID(checkbox.value))
      .filter((id) => id > 0);
  }

  function syncSelection() {
    const allChecks = Array.from(page.querySelectorAll(".js-movie-album-item-check"));
    const ids = selectedIDs();
    if (selectedCountEl) selectedCountEl.textContent = String(ids.length);
    if (batchExecuteBtn) batchExecuteBtn.disabled = !(albumName === "待删除" && ids.length > 0 && albumID > 0);
    if (checkAll) {
      checkAll.checked = allChecks.length > 0 && ids.length === allChecks.length;
      checkAll.indeterminate = ids.length > 0 && ids.length < allChecks.length;
    }
  }

  function removeRowsByIDs(ids) {
    ids.forEach((id) => page.querySelector(`tr[data-item-id="${id}"]`)?.remove());
    syncSelection();
  }

  function renderRemoveDownloadedPreview(items) {
    removeDownloadedPreviewItems = Array.isArray(items) ? items : [];
    if (removeDownloadedCountEl) {
      removeDownloadedCountEl.textContent = String(removeDownloadedPreviewItems.length);
    }
    if (removeDownloadedListEl) {
      removeDownloadedListEl.innerHTML = removeDownloadedPreviewItems.map((item) => {
        const name = String(item?.MovieName || "").trim();
        const javID = String(item?.MovieJavID || "").trim();
        const title = name || javID || "-";
        const meta = javID ? `<div class="small text-muted"><code>${escapeHtml(javID)}</code></div>` : "";
        return `<div class="list-group-item px-0 py-2 border-0 border-bottom">
          <div class="fw-semibold">${escapeHtml(title)}</div>
          ${meta}
        </div>`;
      }).join("");
    }
    if (removeDownloadedEmptyEl) {
      if (removeDownloadedPreviewItems.length > 0) {
        removeDownloadedEmptyEl.classList.add("d-none");
      } else {
        removeDownloadedEmptyEl.textContent = "没有可移除的已下载条目。";
        removeDownloadedEmptyEl.classList.remove("d-none");
      }
    }
    if (removeDownloadedConfirmBtn) {
      removeDownloadedConfirmBtn.disabled = removeDownloadedPreviewItems.length === 0;
    }
  }

  async function createMovieAlbum(name) {
    const album = String(name || "").trim();
    if (!album) {
      showMovieAlbumMessage("请输入电影相册名称", false);
      return;
    }
    if (createConfirmBtn) createConfirmBtn.disabled = true;
    try {
      const data = await requestMovieAlbumJSON("/api/movie-albums", {name: album});
      showMovieAlbumMessage(data.message || "电影相册创建成功", true);
      createModal?.hide();
      window.location.href = `/movie-albums?album_name=${encodeURIComponent(String(data.album_name || album))}`;
    } catch (err) {
      showMovieAlbumMessage(err.message || "电影相册创建失败", false);
      if (createConfirmBtn) createConfirmBtn.disabled = false;
    }
  }

  async function removeOne(itemID, button) {
    if (albumID <= 0) {
      showMovieAlbumMessage("当前电影相册无效", false);
      return;
    }
    if (itemID <= 0) {
      showMovieAlbumMessage("缺少相册条目", false);
      return;
    }
    if (button) button.disabled = true;
    try {
      const data = await requestMovieAlbumJSON(`/api/movie-albums/${albumID}/items/remove`, {item_id: itemID});
      if (data.removed) removeRowsByIDs([itemID]);
      showMovieAlbumMessage(data.message || "移除完成", true);
    } catch (err) {
      showMovieAlbumMessage(err.message || "移除失败", false);
    } finally {
      if (button) button.disabled = false;
    }
  }

  async function executeRemove(ids) {
    if (executeConfirmBtn) executeConfirmBtn.disabled = true;
    try {
      const data = await requestMovieAlbumJSON(`/api/movie-albums/${albumID}/execute-remove`, {item_ids: ids});
      showMovieAlbumMessage(data.message || "统一删除完成", true);
      executeModal?.hide();
      window.setTimeout(() => window.location.reload(), 500);
    } catch (err) {
      showMovieAlbumMessage(err.message || "统一删除失败", false);
      if (executeConfirmBtn) executeConfirmBtn.disabled = false;
    }
  }

  async function loadRemoveDownloadedPreview() {
    if (albumID <= 0) {
      showMovieAlbumMessage("当前电影相册无效", false);
      return;
    }
    if (removeDownloadedBtn) removeDownloadedBtn.disabled = true;
    if (removeDownloadedEmptyEl) {
      removeDownloadedEmptyEl.textContent = "正在加载预览...";
      removeDownloadedEmptyEl.classList.remove("d-none");
    }
    if (removeDownloadedListEl) removeDownloadedListEl.innerHTML = "";
    if (removeDownloadedCountEl) removeDownloadedCountEl.textContent = "0";
    if (removeDownloadedConfirmBtn) removeDownloadedConfirmBtn.disabled = true;
    try {
      const data = await requestMovieAlbumJSON(`/api/movie-albums/${albumID}/remove-downloaded-items-preview`, {});
      renderRemoveDownloadedPreview(data.items);
      if (!Array.isArray(data.items) || data.items.length === 0) {
        showMovieAlbumMessage("没有可移除的已下载条目", true);
        return;
      }
      removeDownloadedModal?.show();
    } catch (err) {
      showMovieAlbumMessage(err.message || "加载移除预览失败", false);
    } finally {
      if (removeDownloadedBtn) removeDownloadedBtn.disabled = false;
    }
  }

  async function removeDownloadedItems() {
    if (albumID <= 0) {
      showMovieAlbumMessage("当前电影相册无效", false);
      return;
    }
    if (removeDownloadedPreviewItems.length === 0) {
      showMovieAlbumMessage("没有可移除的已下载条目", false);
      return;
    }
    if (removeDownloadedConfirmBtn) removeDownloadedConfirmBtn.disabled = true;
    try {
      const data = await requestMovieAlbumJSON(`/api/movie-albums/${albumID}/remove-downloaded-items`, {});
      showMovieAlbumMessage(data.message || "已移除已下载条目", true);
      removeDownloadedModal?.hide();
      window.setTimeout(() => window.location.reload(), 500);
    } catch (err) {
      showMovieAlbumMessage(err.message || "移除已下载条目失败", false);
      if (removeDownloadedConfirmBtn) removeDownloadedConfirmBtn.disabled = false;
    }
  }

  checkAll?.addEventListener("change", () => {
    page.querySelectorAll(".js-movie-album-item-check").forEach((checkbox) => {
      checkbox.checked = checkAll.checked;
    });
    syncSelection();
  });
  batchExecuteBtn?.addEventListener("click", () => {
    const ids = selectedIDs();
    if (!ids.length) {
      showMovieAlbumMessage("请先选择待删除条目", false);
      return;
    }
    if (executeCountEl) executeCountEl.textContent = String(ids.length);
    executeModal?.show();
  });
  executeConfirmBtn?.addEventListener("click", () => {
    const ids = selectedIDs();
    if (!ids.length) {
      showMovieAlbumMessage("请先选择待删除条目", false);
      return;
    }
    executeRemove(ids);
  });
  executeModalEl?.addEventListener("hidden.bs.modal", () => {
    if (executeConfirmBtn) executeConfirmBtn.disabled = false;
  });
  createBtn?.addEventListener("click", () => {
    if (createNameEl) createNameEl.value = "";
    if (createConfirmBtn) createConfirmBtn.disabled = false;
    createModal?.show();
    window.setTimeout(() => createNameEl?.focus(), 50);
  });
  createConfirmBtn?.addEventListener("click", () => createMovieAlbum(createNameEl?.value));
  createNameEl?.addEventListener("keydown", (event) => {
    if (event.key !== "Enter") return;
    event.preventDefault();
    createMovieAlbum(createNameEl.value);
  });
  removeDownloadedBtn?.addEventListener("click", loadRemoveDownloadedPreview);
  removeDownloadedConfirmBtn?.addEventListener("click", removeDownloadedItems);
  removeDownloadedModalEl?.addEventListener("hidden.bs.modal", () => {
    if (removeDownloadedConfirmBtn) removeDownloadedConfirmBtn.disabled = false;
  });
  page.addEventListener("change", (event) => {
    if (event.target?.classList?.contains("js-movie-album-item-check")) syncSelection();
  });
  page.addEventListener("click", (event) => {
    const quickButton = event.target.closest(".js-movie-album-quick-filter");
    if (quickButton) {
      const nextAlbum = String(quickButton.getAttribute("data-album-name") || "").trim();
      if (nextAlbum && albumNameSelect && filterForm) {
        albumNameSelect.value = nextAlbum;
        filterForm.submit();
      } else if (nextAlbum) {
        window.location.href = `/movie-albums?album_name=${encodeURIComponent(nextAlbum)}`;
      }
      return;
    }
    const removeButton = event.target.closest(".js-remove-movie-album-item");
    if (removeButton) {
      removeOne(parsePositiveID(removeButton.getAttribute("data-item-id")), removeButton);
    }
  });

  syncSelection();
}

function bindEItems() {
  const tableBodyEl = document.getElementById("itemTableBody");
  const totalCountEl = document.getElementById("itemTotalCount");
  const singleModalEl = document.getElementById("itemStatusModal");
  const singleForm = document.getElementById("itemStatusForm");
  if (!tableBodyEl || !singleModalEl || !singleForm || !window.bootstrap) return;

  const singleModal = window.bootstrap.Modal.getOrCreateInstance(singleModalEl);
  const singleItemIdInput = document.getElementById("itemStatusItemId");
  const singleMetaEl = document.getElementById("itemStatusMeta");
  const singleMsgEl = document.getElementById("itemStatusMsg");
  const singleSubmitBtn = document.getElementById("itemStatusSubmit");
  const singleHasDetailEl = document.getElementById("itemStatusHasDetail");
  const singleHasDownloadCoverEl = document.getElementById("itemStatusHasDownloadCover");
  const singleHasChineseEl = document.getElementById("itemStatusHasChinese");
  const singleDetailNeedScanEl = document.getElementById("itemStatusDetailNeedScan");

  const batchModalEl = document.getElementById("itemBatchStatusModal");
  const batchForm = document.getElementById("itemBatchStatusForm");
  const batchModal = batchModalEl ? window.bootstrap.Modal.getOrCreateInstance(batchModalEl) : null;
  const batchMetaEl = document.getElementById("itemBatchStatusMeta");
  const batchMsgEl = document.getElementById("itemBatchStatusMsg");
  const batchSubmitBtn = document.getElementById("itemBatchStatusSubmit");
  const batchHasDetailEl = document.getElementById("itemBatchHasDetail");
  const batchHasDownloadCoverEl = document.getElementById("itemBatchHasDownloadCover");
  const batchHasChineseEl = document.getElementById("itemBatchHasChinese");
  const batchDetailNeedScanEl = document.getElementById("itemBatchDetailNeedScan");
  const batchEditBtn = document.getElementById("btnBatchEditStatus");
  const selectedCountEl = document.getElementById("itemBatchSelectedCount");
  const checkAllEl = document.getElementById("itemCheckAll");

  const deleteModalEl = document.getElementById("itemDeleteModal");
  const deleteModal = deleteModalEl ? window.bootstrap.Modal.getOrCreateInstance(deleteModalEl) : null;
  const deleteMetaEl = document.getElementById("itemDeleteMeta");
  const deleteMsgEl = document.getElementById("itemDeleteMsg");
  const deleteItemIdInput = document.getElementById("itemDeleteItemId");
  const deleteSubmitBtn = document.getElementById("itemDeleteSubmit");

  function showMsg(el, text, ok) {
    if (!el) return;
    el.textContent = String(text || "");
    el.className = `alert small py-2 px-3 mt-3 mb-0 ${ok ? "alert-success" : "alert-danger"}`;
  }

  function hideMsg(el) {
    if (!el) return;
    el.textContent = "";
    el.className = "alert small py-2 px-3 mt-3 mb-0 d-none";
  }

  function setSubmitting(button, submitting, idleText, submittingText) {
    if (!button) return;
    button.disabled = submitting;
    button.textContent = submitting ? submittingText : idleText;
  }

  function setSelectValue(selectEl, value) {
    if (!selectEl) return;
    const next = String(value || "");
    selectEl.value = Array.from(selectEl.options || []).some((option) => option.value === next)
      ? next
      : String(selectEl.options?.[0]?.value || "");
  }

  function getItemRows() {
    return Array.from(tableBodyEl.querySelectorAll("tr[data-item-id]"));
  }

  function getItemCheckboxes() {
    return Array.from(tableBodyEl.querySelectorAll(".js-item-check"));
  }

  function getSelectedIds() {
    return getItemCheckboxes()
      .filter((checkbox) => checkbox.checked)
      .map((checkbox) => Number.parseInt(String(checkbox.value || "0"), 10))
      .filter((id) => Number.isInteger(id) && id > 0);
  }

  function updateButtonDataset(button, item) {
    if (!button || !item) return;
    button.dataset.hasDetail = String(item.has_detail_value || "");
    button.dataset.hasDownloadCover = String(item.has_download_cover_value || "");
    button.dataset.hasChinese = String(item.has_chinese_value || "");
    button.dataset.detailNeedScan = String(item.detail_need_scan_value || "");
  }

  function updateRow(item) {
    if (!item) return;
    const row = tableBodyEl.querySelector(`tr[data-item-id="${String(item.id)}"]`);
    if (!row) return;
    const hasDetailCell = row.querySelector(".js-item-has-detail");
    const hasDownloadCoverCell = row.querySelector(".js-item-has-download-cover");
    const hasChineseCell = row.querySelector(".js-item-has-chinese");
    const detailNeedScanCell = row.querySelector(".js-item-detail-need-scan");
    const editBtn = row.querySelector(".js-item-status-edit");

    if (hasDetailCell) hasDetailCell.textContent = item.has_detail_text || "-";
    if (hasDownloadCoverCell) hasDownloadCoverCell.textContent = item.has_download_cover_text || "-";
    if (hasChineseCell) hasChineseCell.textContent = item.has_chinese_text || "-";
    if (detailNeedScanCell) detailNeedScanCell.textContent = item.detail_need_scan_text || "-";
    updateButtonDataset(editBtn, item);
  }

  function updateTotalCount(delta) {
    if (!totalCountEl) return;
    const current = Number.parseInt(String(totalCountEl.textContent || "0"), 10);
    if (!Number.isFinite(current)) return;
    totalCountEl.textContent = String(Math.max(0, current + delta));
  }

  function renderEmptyStateIfNeeded() {
    if (getItemRows().length > 0) return;
    tableBodyEl.innerHTML = `<tr class="js-item-empty-row"><td colspan="10"><div class="sc-empty-state">当前筛选下没有 e_item 数据。</div></td></tr>`;
  }

  function removeEmptyState() {
    tableBodyEl.querySelector(".js-item-empty-row")?.remove();
  }

  function syncSelectionState() {
    const checkboxes = getItemCheckboxes();
    const checkedCount = checkboxes.filter((checkbox) => checkbox.checked).length;
    if (selectedCountEl) selectedCountEl.textContent = String(checkedCount);
    if (batchEditBtn) batchEditBtn.disabled = checkedCount === 0;
    if (checkAllEl) {
      checkAllEl.checked = checkboxes.length > 0 && checkedCount === checkboxes.length;
      checkAllEl.indeterminate = checkedCount > 0 && checkedCount < checkboxes.length;
    }
    checkboxes.forEach((checkbox) => {
      const row = checkbox.closest("tr");
      if (row) row.classList.toggle("table-active", checkbox.checked);
    });
  }

  function clearSelection() {
    getItemCheckboxes().forEach((checkbox) => {
      checkbox.checked = false;
    });
    if (checkAllEl) {
      checkAllEl.checked = false;
      checkAllEl.indeterminate = false;
    }
    syncSelectionState();
  }

  function buildBatchPayload(ids) {
    const payload = {ids};
    let changedFieldCount = 0;
    if (batchHasDetailEl?.value) {
      payload.hasDetail = batchHasDetailEl.value;
      changedFieldCount += 1;
    }
    if (batchHasDownloadCoverEl?.value) {
      payload.hasDownloadCover = batchHasDownloadCoverEl.value;
      changedFieldCount += 1;
    }
    if (batchHasChineseEl?.value) {
      payload.hasChinese = batchHasChineseEl.value;
      changedFieldCount += 1;
    }
    if (batchDetailNeedScanEl?.value) {
      payload.detailNeedScan = batchDetailNeedScanEl.value;
      changedFieldCount += 1;
    }
    return {payload, changedFieldCount};
  }

  function removeRow(itemId) {
    const id = String(itemId || "");
    tableBodyEl.querySelector(`tr[data-item-id="${id}"]`)?.remove();
    updateTotalCount(-1);
    syncSelectionState();
    renderEmptyStateIfNeeded();
  }

  document.addEventListener("change", (event) => {
    const target = event.target;
    if (!(target instanceof Element)) return;
    if (target.classList.contains("js-item-check")) {
      syncSelectionState();
      return;
    }
    if (checkAllEl && target === checkAllEl) {
      getItemCheckboxes().forEach((checkbox) => {
        checkbox.checked = checkAllEl.checked;
      });
      syncSelectionState();
    }
  });

  document.addEventListener("click", (event) => {
    const target = event.target;
    if (!(target instanceof Element)) return;

    const editButton = target.closest(".js-item-status-edit");
    if (editButton) {
      if (singleItemIdInput) singleItemIdInput.value = editButton.getAttribute("data-item-id") || "";
      if (singleMetaEl) {
        singleMetaEl.textContent = `${editButton.getAttribute("data-item-jav-id") || "-"} / ${editButton.getAttribute("data-item-name") || "-"}`;
      }
      setSelectValue(singleHasDetailEl, editButton.getAttribute("data-has-detail"));
      setSelectValue(singleHasDownloadCoverEl, editButton.getAttribute("data-has-download-cover"));
      setSelectValue(singleHasChineseEl, editButton.getAttribute("data-has-chinese"));
      setSelectValue(singleDetailNeedScanEl, editButton.getAttribute("data-detail-need-scan"));
      hideMsg(singleMsgEl);
      singleModal.show();
      return;
    }

    const deleteButton = target.closest(".js-item-delete");
    if (deleteButton) {
      if (deleteItemIdInput) deleteItemIdInput.value = deleteButton.getAttribute("data-item-id") || "";
      if (deleteMetaEl) {
        deleteMetaEl.textContent = `${deleteButton.getAttribute("data-item-jav-id") || "-"} / ${deleteButton.getAttribute("data-item-name") || "-"}`;
      }
      hideMsg(deleteMsgEl);
      deleteModal?.show();
      return;
    }

    if (batchEditBtn && target === batchEditBtn) {
      const ids = getSelectedIds();
      if (!ids.length) return;
      if (batchMetaEl) batchMetaEl.textContent = `已选 ${ids.length} 项`;
      setSelectValue(batchHasDetailEl, "");
      setSelectValue(batchHasDownloadCoverEl, "");
      setSelectValue(batchHasChineseEl, "");
      setSelectValue(batchDetailNeedScanEl, "");
      hideMsg(batchMsgEl);
      batchModal?.show();
    }
  });

  singleModalEl.addEventListener("hidden.bs.modal", () => {
    hideMsg(singleMsgEl);
    setSubmitting(singleSubmitBtn, false, "保存", "保存中...");
  });
  batchModalEl?.addEventListener("hidden.bs.modal", () => {
    hideMsg(batchMsgEl);
    setSubmitting(batchSubmitBtn, false, "批量保存", "保存中...");
  });
  deleteModalEl?.addEventListener("hidden.bs.modal", () => {
    hideMsg(deleteMsgEl);
    setSubmitting(deleteSubmitBtn, false, "确认删除", "删除中...");
  });

  singleForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    hideMsg(singleMsgEl);
    const itemId = String(singleItemIdInput?.value || "").trim();
    if (!itemId) {
      showMsg(singleMsgEl, "缺少条目 id", false);
      return;
    }

    setSubmitting(singleSubmitBtn, true, "保存", "保存中...");
    try {
      const payload = await fetchJSON(`/api/e-items/${encodeURIComponent(itemId)}/status`, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({
          hasDetail: singleHasDetailEl?.value || "",
          hasDownloadCover: singleHasDownloadCoverEl?.value || "",
          hasChinese: singleHasChineseEl?.value || "",
          detailNeedScan: singleDetailNeedScanEl?.value || ""
        })
      });
      if (payload?.ok === false) throw new Error(payload.error || "保存失败");
      updateRow(payload.item);
      showMsg(singleMsgEl, "状态已更新", true);
      window.setTimeout(() => singleModal.hide(), 250);
    } catch (err) {
      showMsg(singleMsgEl, err.message || "保存失败", false);
    } finally {
      setSubmitting(singleSubmitBtn, false, "保存", "保存中...");
    }
  });

  batchForm?.addEventListener("submit", async (event) => {
    event.preventDefault();
    hideMsg(batchMsgEl);
    const ids = getSelectedIds();
    if (!ids.length) {
      showMsg(batchMsgEl, "请先选择条目", false);
      return;
    }

    const built = buildBatchPayload(ids);
    if (built.changedFieldCount === 0) {
      showMsg(batchMsgEl, "至少选择一个要修改的状态", false);
      return;
    }

    setSubmitting(batchSubmitBtn, true, "批量保存", "保存中...");
    try {
      const payload = await fetchJSON("/api/e-items/status/batch", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(built.payload)
      });
      if (payload?.ok === false) throw new Error(payload.error || "保存失败");
      if (Array.isArray(payload.items)) payload.items.forEach(updateRow);
      clearSelection();
      showMsg(batchMsgEl, "批量状态已更新", true);
      window.setTimeout(() => batchModal?.hide(), 250);
    } catch (err) {
      showMsg(batchMsgEl, err.message || "批量保存失败", false);
    } finally {
      setSubmitting(batchSubmitBtn, false, "批量保存", "保存中...");
    }
  });

  deleteSubmitBtn?.addEventListener("click", async () => {
    hideMsg(deleteMsgEl);
    const itemId = String(deleteItemIdInput?.value || "").trim();
    if (!itemId) {
      showMsg(deleteMsgEl, "缺少条目 id", false);
      return;
    }

    setSubmitting(deleteSubmitBtn, true, "确认删除", "删除中...");
    try {
      const payload = await fetchJSON(`/api/e-items/${encodeURIComponent(itemId)}`, {method: "DELETE"});
      if (payload?.ok === false) throw new Error(payload.error || "删除失败");
      removeEmptyState();
      removeRow(payload.id || itemId);
      showMsg(deleteMsgEl, "电影已删除", true);
      window.setTimeout(() => deleteModal?.hide(), 250);
    } catch (err) {
      showMsg(deleteMsgEl, err.message || "删除失败", false);
    } finally {
      setSubmitting(deleteSubmitBtn, false, "确认删除", "删除中...");
    }
  });

  syncSelectionState();
}

function render(data) {
  const view = data.ViewType;
  if (view === "medias") return renderMedias(data);
  if (view === "crawl-records") return renderCrawlRecords(data);
  if (view === "w-agg-events") return renderWAggEvents(data);
  if (view === "torrent-albums") return renderTorrentAlbums(data);
  if (view === "movie-albums") return renderMovieAlbums(data);
  if (view === "d-seeds") return renderDSeeds(data);
  if (view === "casts") return renderCasts(data);
  if (view === "e-items") return renderEItems(data);
  return `<div class="alert alert-warning">未知页面类型：${escapeHtml(view)}</div>`;
}

async function main() {
  try {
    const data = await fetchJSON(apiPathForCurrentPage() + window.location.search);
    document.title = `${data.Title || data.PageTitle || "List"} - Rudy-GC`;
    root.innerHTML = render(data);
    bindDSeedEditor();
    bindTorrentAlbums();
    bindMovieAlbums();
    bindEItems();
  } catch (err) {
    root.innerHTML = `<div class="alert alert-danger">加载失败：${escapeHtml(err.message)}</div>`;
  }
}

document.addEventListener("change", (event) => {
  const pageSize = event.target.closest("#crawlRecordPageSize");
  if (!pageSize) {
    return;
  }
  const form = pageSize.closest("form");
  if (form) {
    form.submit();
  }
});

document.addEventListener("click", (event) => {
  const btn = event.target.closest(".js-film-probe");
  if (!btn) {
    return;
  }
  event.preventDefault();
  handleMediaProbe(btn);
});

main();
