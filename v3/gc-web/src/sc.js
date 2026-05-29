const root = document.getElementById("page-root");
const msgArea = document.getElementById("msgArea");
const GLIST_IS_NOT_SC = 1;
const GLIST_IS_SC = 2;

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

function formatUnixHm(value) {
  const n = Number(value || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}

function humanDays(seconds) {
  const n = Number(seconds || 0);
  if (n <= 0) return "-";
  return `${(n / 86400).toFixed(2)} 天`;
}

function humanSize(bytes) {
  const value = Number(bytes || 0);
  if (!value) return "-";
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`;
  return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

function monthParityClass(value) {
  const n = Number(value || 0);
  if (n <= 0) return "sc-month-even";
  const month = new Date(n * 1000).getMonth() + 1;
  return month % 2 === 0 ? "sc-month-even" : "sc-month-odd";
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
  const path = window.location.pathname;
  if (path === "/sc-events") return "/api/sc-events";
  if (path === "/sc-events-cards") return "/api/sc-events-cards";
  if (path.startsWith("/sc-events/")) {
    return `/api/sc-events/${encodeURIComponent(decodeURIComponent(path.slice("/sc-events/".length)))}`;
  }
  if (path === "/sc-pick-smart-media") return "/api/sc-pick-smart-media";
  if (path === "/triggers/sc") return "/api/triggers/sc";
  if (path === "/triggers/media") return "/api/triggers/media";
  if (path === "/triggers/media-rescan") return "/api/triggers/media-rescan";
  if (path === "/triggers/media-rollback") return "/api/triggers/media-rollback";
  if (path === "/triggers/sc-media") return "/api/triggers/sc-media";
  return "/api/sc-events";
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

function renderHeader(data, subtitle = "") {
  return `<div class="d-flex align-items-center justify-content-between mb-3">
    <div>
      <h4 class="mb-1">${escapeHtml(data.PageTitle || data.Title || "")}</h4>
      ${subtitle || data.PageNote ? `<p class="text-muted mb-0">${escapeHtml(subtitle || data.PageNote)}</p>` : ""}
    </div>
  </div>`;
}

function renderSummary(data, label) {
  const items = Array.isArray(data.Items) ? data.Items : [];
  return `<div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
    <div class="sc-page-summary">
      <span class="sc-page-summary-label">${escapeHtml(label)}</span>
      <span class="sc-page-summary-value">${escapeHtml(data.Total ?? 0)}</span>
    </div>
    ${items.length ? `<div class="ms-xl-auto">${renderPagination(data.PageInfo || data.pageInfo)}</div>` : ""}
  </div>`;
}

function linkSort(link, label, button = false) {
  const suffix = link?.Active ? (link.Desc ? " ↓" : " ↑") : "";
  if (button) {
    return `<a href="${attr(link?.Href || "#")}" class="btn btn-sm btn-outline-primary ${link?.Active ? "active disabled" : ""}">${escapeHtml(label + suffix)}</a>`;
  }
  return `<a class="sc-sort-link ${link?.Active ? "active" : ""}" href="${attr(link?.Href || "#")}">${escapeHtml(label + suffix)}</a>`;
}

function linkCell(href, text, fallback = "-") {
  const value = String(text || "");
  if (!value) return `<span class="text-muted">${escapeHtml(fallback)}</span>`;
  if (!href) return escapeHtml(value);
  return `<a href="${attr(href)}">${escapeHtml(value)}</a>`;
}

function legacyPathValue(value) {
  return encodeURI(String(value || ""));
}

function eventDetailHref(name) {
  return `/sc-events/${legacyPathValue(name)}`;
}

function movieHref(name) {
  return name ? `/movie/${legacyPathValue(name)}` : "";
}

function formatUnixFull(value) {
  const n = Number(value || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  const hh = String(d.getHours()).padStart(2, "0");
  const mi = String(d.getMinutes()).padStart(2, "0");
  const ss = String(d.getSeconds()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}:${ss}`;
}

function renderScViewSwitch(viewQ) {
  return `<div class="btn-group" role="group" aria-label="SC event view switch">
    <a href="${attr(viewQ?.ListHref || "/sc-events")}" class="btn btn-outline-primary ${viewQ?.IsList ? "active disabled" : ""}">列表视图</a>
    <a href="${attr(viewQ?.CardHref || "/sc-events-cards")}" class="btn btn-outline-primary ${viewQ?.IsCard ? "active disabled" : ""}">卡片视图</a>
  </div>`;
}

function renderScEventList(data) {
  const sort = data.ScSortQuery || {};
  const rows = (data.Items || []).map((item) => {
    const event = item.Event || {};
    const cooldown = Number(event.Cooldown || 0);
    const duration = Number(event.DurationMinutes || 0);
    return `<tr class="${monthParityClass(event.ScTime)}">
      <td><a class="sc-event-link" href="${attr(eventDetailHref(event.Name))}" title="${attr(event.Name || "")}">
        <span class="sc-event-time"><span class="sc-event-date">${formatUnix(event.ScTime)}</span> <span class="sc-event-hour">${formatUnixHm(event.ScTime)}</span></span>
      </a></td>
      <td>${escapeHtml(item.TotalMovieCount)}</td>
      <td>${escapeHtml(item.ScMovieCount)}</td>
      <td>${linkCell(movieHref(event.ComeMovieName), event.ComeMovieName)}</td>
      <td>${cooldown > 0 ? humanDays(cooldown) : '<span class="text-muted">-</span>'}</td>
      <td class="sc-muted">${duration > 0 ? escapeHtml(duration) : '<span class="text-muted">-</span>'}</td>
      <td class="sc-muted">${event.Kind ? escapeHtml(event.Kind) : '<span class="text-muted">-</span>'}</td>
      <td class="sc-muted">${linkCell(item.MovieCastHref, event.MovieCast)}</td>
      <td class="sc-muted">${event.Vessel ? escapeHtml(event.Vessel) : '<span class="text-muted">-</span>'}</td>
      <td class="sc-muted">${event.Fg ? escapeHtml(event.Fg) : '<span class="text-muted">-</span>'}</td>
    </tr>`;
  });
  return `<div class="d-flex align-items-center justify-content-between mb-3">
      <div>
        <h4 class="mb-1">SC Events</h4>
        <p class="text-muted mb-0">浏览最近的 SC 事件，并进入详情页查看本次影片清单。</p>
      </div>
      <div class="d-flex flex-wrap gap-2">${renderScViewSwitch(data.ScViewQuery)}<a href="/triggers/backfill" class="btn btn-outline-secondary">回填任务台</a></div>
    </div>
    ${renderSummary(data, "Total Events")}
    <div class="card">
      <div class="card-body">
        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
            <thead><tr>
              <th>${linkSort(sort.ByTime, "时间")}</th>
              <th>总片数</th>
              <th>${linkSort(sort.ByMovieCnt, "有 SC")}</th>
              <th>${linkSort(sort.ByComeMovie, "Come 片")}</th>
              <th>${linkSort(sort.ByCooldown, "Cooldown")}</th>
              <th>${linkSort(sort.ByDuration, "时长(分钟)")}</th>
              <th>${linkSort(sort.ByKind, "种类")}</th>
              <th>${linkSort(sort.ByMovieCast, "演员")}</th>
              <th>${linkSort(sort.ByVessel, "Vessel")}</th>
              <th>${linkSort(sort.ByFg, "FG")}</th>
            </tr></thead>
            <tbody>${rows.length ? rows.join("") : '<tr><td colspan="10"><div class="sc-empty-state">还没有可展示的 SC 事件。</div></td></tr>'}</tbody>
          </table>
        </div>
      </div>
    </div>`;
}

function renderScEventCards(data) {
  const sort = data.ScSortQuery || {};
  const cards = (data.Items || []).map((item) => {
    const event = item.Event || {};
    const cover = item.HasComeMovieCover
      ? `<img src="${attr(item.ComeMovieJacketImg)}" alt="${attr(event.Name)} 封面" loading="lazy" decoding="async">`
      : `<div class="sc-event-cover-empty">
          <span class="sc-event-cover-kicker">SC EVENT</span>
          <strong><span class="sc-event-time"><span class="sc-event-date">${formatUnix(event.ScTime)}</span><span class="sc-event-hour">${formatUnixHm(event.ScTime)}</span></span></strong>
          <span>${escapeHtml(event.Name)}</span>
        </div>`;
    return `<div class="col">
      <article class="card shadow-sm h-100 card-lift sc-event-grid-card ${monthParityClass(event.ScTime)}">
        <a href="${attr(item.DetailHref || eventDetailHref(event.Name))}" class="text-decoration-none text-reset">
          <div class="thumb-2x3 sc-event-cover">
            ${cover}
            <div class="sc-event-cover-badge">${escapeHtml(event.MovieNumber)} 部</div>
          </div>
        </a>
        <div class="card-body sc-event-card-body">
          <div class="small mb-2"><span class="sc-event-time"><span class="sc-event-date">${formatUnix(event.ScTime)}</span> <span class="sc-event-hour">${formatUnixHm(event.ScTime)}</span></span></div>
          <h5 class="h6 mb-2"><a href="${attr(item.DetailHref || eventDetailHref(event.Name))}" class="text-decoration-none text-reset">${escapeHtml(event.Name)}</a></h5>
          <div class="mb-3"><div class="text-muted small text-uppercase mb-1">Come Movie</div>${linkCell(item.ComeMovieHref, item.ComeMovieName)}</div>
          <div class="sc-event-meta-grid">
            <span class="sc-pill">${event.Cooldown ? humanDays(event.Cooldown) : "Cooldown -"}</span>
            <span class="sc-pill">${event.Kind ? `种类 ${escapeHtml(event.Kind)}` : "种类 -"}</span>
            <span class="sc-pill">${event.MovieCast ? escapeHtml(event.MovieCast) : "演员 -"}</span>
            <span class="sc-pill">${event.Vessel ? `Vessel ${escapeHtml(event.Vessel)}` : "Vessel -"}</span>
            <span class="sc-pill">${event.Fg ? `FG ${escapeHtml(event.Fg)}` : "FG -"}</span>
          </div>
        </div>
      </article>
    </div>`;
  });
  return `<div class="d-flex align-items-center justify-content-between mb-3">
      <div>
        <h4 class="mb-1">SC Event Cards</h4>
        <p class="text-muted mb-0">以封面卡片的方式浏览 SC 事件，主图使用 come movie 的电影封面。</p>
      </div>
      <div class="d-flex flex-wrap gap-2">${renderScViewSwitch(data.ScViewQuery)}<a href="/triggers/backfill" class="btn btn-outline-secondary">回填任务台</a></div>
    </div>
    <div class="d-flex flex-column flex-xl-row align-items-xl-start justify-content-between gap-3 mb-3">
      <div class="d-flex align-items-center gap-3 flex-wrap">
        <div class="sc-page-summary"><span class="sc-page-summary-label">Total Events</span><span class="sc-page-summary-value">${escapeHtml(data.Total ?? 0)}</span></div>
        <div class="d-flex flex-wrap gap-2 sc-event-sort-bar">
          ${linkSort(sort.ByTime, "时间", true)}
          ${linkSort(sort.ByMovieCnt, "片数", true)}
          ${linkSort(sort.ByComeMovie, "Come 片", true)}
          ${linkSort(sort.ByCooldown, "Cooldown", true)}
          ${linkSort(sort.ByKind, "种类", true)}
          ${linkSort(sort.ByMovieCast, "演员", true)}
          ${linkSort(sort.ByVessel, "Vessel", true)}
          ${linkSort(sort.ByFg, "FG", true)}
        </div>
      </div>
      <div class="ms-xl-auto">${renderPagination(data.PageInfo || data.pageInfo)}</div>
    </div>
    ${cards.length ? `<div class="row row-cols-2 row-cols-sm-3 row-cols-lg-5 row-cols-xxl-6 g-3">${cards.join("")}</div>` : '<div class="sc-empty-state">还没有可展示的 SC 事件。</div>'}`;
}

function renderMoviePlayRow(movie) {
  if ((movie.owned_w_media || 0) < 4) {
    return "";
  }
  let play = "";
  if (movie.owned_w_media === 5) {
    play = `<a href="${attr(movie.video_url_w_media)}" class="btn btn-sm btn-info">MPlay</a>`;
  } else if (movie.owned_w_media === 4) {
    play = `<a href="${attr(movie.video_url_w_media)}" class="btn btn-sm btn-success">MPlay</a>`;
  } else if (movie.owned_w_media === 6) {
    play = '<span class="btn btn-sm btn-secondary disabled" aria-disabled="true">MRemoved</span>';
  }
  return `<li class="list-group-item py-2">
    <div class="card-play-stack"><div class="card-play-row">
      ${play}
      ${movie.film_birth_date_w_media ? `<span class="text-muted small card-play-date">${escapeHtml(movie.film_birth_date_w_media)}</span>` : ""}
    </div></div>
  </li>`;
}

function renderMovieCard(movie) {
  movie = movie || {};
  const genres = (movie.genre || []).map((genre) =>
    `<a href="/cards?gn=${encodeURIComponent(genre)}" class="badge badge-genre me-1 mb-1">${escapeHtml(genre)}</a>`
  ).join("");
  const casts = (movie.cast || []).map((cast) => {
    const nameShow = cast.NameShow || cast.DisplayName || cast.Name || "";
    if ((cast.PersonId || 0) > 0) {
      return `<a href="/cast/${attr(cast.PersonId)}" class="person-tag mr-2" title="${attr(nameShow)}">${escapeHtml(nameShow)}</a>`;
    }
    return `<a href="/cards?cn=${encodeURIComponent(cast.Name || "")}" class="person-tag mr-2" title="${attr(nameShow)}">${escapeHtml(nameShow)}</a>`;
  }).join("");
  const director = movie.director
    ? `<a href="/cards?dn=${encodeURIComponent(movie.director)}" class="person-tag ml-2" title="${attr(movie.director)}">${escapeHtml(movie.director)}</a>`
    : "";
  const scoreBadge = String(movie.score ?? "") !== "" ? `<span class="badge-score">★ ${escapeHtml(movie.score)}</span>` : "";
  const watched = String(movie.viewers_number_watched ?? "") !== "" ? `<small class="text-muted ml-2">(${escapeHtml(movie.viewers_number_watched)})</small>` : "";
  return `<div class="col">
    <article class="card shadow-sm h-100 card-lift">
      <a href="/movie/${attr(movie.name || "")}" class="d-block" aria-label="查看 ${attr(movie.name || "")} 详情">
        <div class="thumb-2x3">
          <img src="${attr(movie.jacket_img || "")}" alt="${attr(movie.name || "")} 海报" loading="lazy" decoding="async">
          <div class="badge-top-left">
            ${(movie.come_times || 0) > 0 ? `<span class="badge-come">❤ ${escapeHtml(movie.come_times)}</span>` : ""}
            ${(movie.sc_times || 0) > 0 ? ` <span class="badge-times">× ${escapeHtml(movie.sc_times)}</span>` : ""}
          </div>
          <div class="badge-bottom-right">
            ${(movie.highest_rank || 0) > 0 ? `<span class="badge-rank">#${escapeHtml(movie.highest_rank)}</span>` : ""}
            ${scoreBadge}
          </div>
        </div>
      </a>
      <ul class="list-group list-group-flush">
        <li class="list-group-item">
          <h6 class="mb-1 text-truncate" title="${attr(movie.title || "")}">
            <a href="/cards?pn=${encodeURIComponent(movie.prefix || "")}" class="stretched-link text-reset text-decoration-none">${escapeHtml(movie.name || "")}</a>
            ${watched}
          </h6>
          ${movie.title ? `<p class="mb-0 text-muted line-clamp-2" title="${attr(movie.title)}">${escapeHtml(movie.title)}</p>` : ""}
        </li>
        ${renderMoviePlayRow(movie)}
        <li class="list-group-item text-muted">${genres}</li>
        <li class="list-group-item">${casts}${director}</li>
        <li class="list-group-item d-flex align-items-center small">
          ${movie.bus_url ? `<a href="${attr(movie.bus_url)}" class="me-2" target="_blank" rel="noopener">Bus</a>` : ""}
          ${movie.search_url ? `<a href="${attr(movie.search_url)}" class="me-2" target="_blank" rel="noopener">Suk</a>` : ""}
          ${movie.jav_url ? `<a href="${attr(movie.jav_url)}" class="me-2" target="_blank" rel="noopener">Lib</a>` : ""}
          <span class="ms-auto text-muted">${escapeHtml(movie.releasing_date || "")}</span>
        </li>
      </ul>
    </article>
  </div>`;
}

function renderScEventEditModal(data, event) {
  const comeOptions = (data.EditComeMovieOptions || []).map((option) =>
    `<option value="${attr(option.MovieJavId || "")}" ${option.IsCome ? "selected" : ""}>${escapeHtml(option.MovieName || "")}</option>`
  ).join("");
  const currentCastOptions = (data.EditCurrentMovieCastNames || []).map((name) =>
    `<option value="${attr(name)}" ${event.MovieCast === name ? "selected" : ""}>${escapeHtml(name)}</option>`
  ).join("");
  const castSourcePool = (data.EditComeMovieOptions || []).map((option) => {
    const casts = (option.CastOptions || []).map((name) =>
      `<span class="js-sc-event-cast-item" data-cast-name="${attr(name)}"></span>`
    ).join("");
    return `<div class="js-sc-event-cast-source" data-movie-jav-id="${attr(option.MovieJavId || "")}">${casts}</div>`;
  }).join("");

  return `<div class="modal fade" id="scEventEditModal" tabindex="-1" aria-hidden="true" aria-labelledby="scEventEditModalTitle">
    <div class="modal-dialog modal-lg modal-dialog-centered">
      <div class="modal-content">
        <form id="scEventEditForm" data-update-url="/api/sc-events/${encodeURIComponent(event.Name || "")}">
          <div class="modal-header">
            <div>
              <h5 class="modal-title mb-1" id="scEventEditModalTitle">编辑 SC Event</h5>
              <div class="small text-muted">${escapeHtml(event.Name || "")}</div>
            </div>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="关闭"></button>
          </div>
          <div class="modal-body">
            <div class="row g-3">
              <div class="col-xl-3 col-lg-4 col-md-6">
                <label class="form-label mb-1" for="scEventEditName">事件名</label>
                <input class="form-control" id="scEventEditName" type="text" value="${attr(event.Name || "")}" readonly>
              </div>
              <div class="col-xl-3 col-lg-4 col-md-6">
                <label class="form-label mb-1" for="scEventEditImage">图片</label>
                <input class="form-control" id="scEventEditImage" type="text" value="${attr(data.EditImageName || "无图片")}" readonly>
              </div>
              <div class="col-xl-3 col-lg-4 col-md-6">
                <label class="form-label mb-1" for="scEventEditMovieCount">影片数</label>
                <input class="form-control" id="scEventEditMovieCount" type="text" value="${attr(data.EditMovieCount ?? 0)}" readonly>
              </div>
              <div class="col-xl-3 col-lg-4 col-md-6">
                <label class="form-label mb-1" for="scEventEditScMovieCount">SC 影片数</label>
                <input class="form-control" id="scEventEditScMovieCount" type="text" value="${attr(data.EditScMovieCount ?? 0)}" readonly>
              </div>
              <div class="col-xl-3 col-lg-4 col-md-6">
                <label class="form-label mb-1" for="scEventEditComeMovie">Come Movie</label>
                <select class="form-select" id="scEventEditComeMovie">
                  <option value="">可为空</option>
                  ${comeOptions}
                </select>
              </div>
              <div class="col-xl-3 col-lg-4 col-md-6">
                <label class="form-label mb-1" for="scEventEditMovieCast">Movie Cast</label>
                <select class="form-select" id="scEventEditMovieCast" data-selected-value="${attr(event.MovieCast || "")}">
                  <option value="">可为空</option>
                  ${currentCastOptions}
                </select>
              </div>
              <div class="col-xl-3 col-lg-4 col-md-6">
                <label class="form-label mb-1" for="scEventEditKind">种类</label>
                <input class="form-control" id="scEventEditKind" type="text" value="${attr(event.Kind || "")}" placeholder="可为空">
              </div>
              <div class="col-xl-3 col-lg-4 col-md-6">
                <label class="form-label mb-1" for="scEventEditDuration">时长(分钟)</label>
                <input class="form-control" id="scEventEditDuration" type="number" step="1" value="${attr(event.DurationMinutes || 0)}">
              </div>
              <div class="col-xl-3 col-lg-4 col-md-6">
                <label class="form-label mb-1" for="scEventEditFg">FG</label>
                <input class="form-control" id="scEventEditFg" type="text" value="${attr(event.Fg || "")}" placeholder="可为空">
              </div>
              <div class="col-xl-3 col-lg-4 col-md-6">
                <label class="form-label mb-1" for="scEventEditVessel">Vessel</label>
                <input class="form-control" id="scEventEditVessel" type="text" value="${attr(event.Vessel || "")}" placeholder="可为空">
              </div>
              <div class="col-12">
                <label class="form-label mb-1" for="scEventEditRemarks">Remarks</label>
                <textarea class="form-control" id="scEventEditRemarks" rows="3" placeholder="可为空">${escapeHtml(event.Remarks || "")}</textarea>
              </div>
            </div>
            <div class="d-none" id="scEventEditCastSourcePool">${castSourcePool}</div>
            <div class="alert d-none mt-3 mb-0" id="scEventEditMsg" role="alert"></div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-outline-secondary" data-bs-dismiss="modal">取消</button>
            <button type="submit" class="btn btn-primary" id="scEventEditSubmit">保存</button>
          </div>
        </form>
      </div>
    </div>
  </div>`;
}

function renderScDetailMetric(label, value) {
  return `<div class="sc-detail-info-item sc-metric-card">
    <div class="sc-metric-label">${escapeHtml(label)}</div>
    <div class="sc-metric-value">${value}</div>
  </div>`;
}

function renderScEventDetailInfo(event) {
  const imageBlock = event.ImagePath
    ? `<div class="col-12 col-xl-6">
        <div class="card h-100 sc-detail-side-card">
          <div class="card-body">
            <div class="text-muted small text-uppercase mb-2">SC Image</div>
            <div class="sc-detail-image-wrap">
              <img src="${attr(event.ImagePath)}" alt="${attr(event.Name || "")} 图片" class="sc-detail-image">
            </div>
          </div>
        </div>
      </div>`
    : "";
  const comeMovie = event.ComeMovieName
    ? `<a href="${attr(movieHref(event.ComeMovieName))}">${escapeHtml(event.ComeMovieName)}</a>`
    : "-";
  const leadCast = event.MovieCast
    ? `<a href="/cards?cn=${encodeURIComponent(event.MovieCast)}">${escapeHtml(event.MovieCast)}</a>`
    : "-";
  return `<div class="row g-3 mb-3">
    ${imageBlock}
    <div class="col-12 ${event.ImagePath ? "col-xl-6" : ""}">
      <div class="card h-100 sc-detail-side-card">
        <div class="card-body">
          <div class="sc-detail-info-grid">
            ${renderScDetailMetric("Sc Time", escapeHtml(formatUnixFull(event.ScTime)))}
            ${renderScDetailMetric("Movie Count", escapeHtml(event.MovieNumber ?? 0))}
            ${renderScDetailMetric("Cooldown", event.Cooldown > 0 ? escapeHtml(humanDays(event.Cooldown)) : "-")}
            ${renderScDetailMetric("Come Movie", comeMovie)}
            ${renderScDetailMetric("种类", escapeHtml(event.Kind || "-"))}
            ${renderScDetailMetric("Lead Cast", leadCast)}
            ${renderScDetailMetric("时长(分钟)", event.DurationMinutes > 0 ? `${escapeHtml(event.DurationMinutes)} 分钟` : "-")}
            ${renderScDetailMetric("FG", escapeHtml(event.Fg || "-"))}
            ${renderScDetailMetric("Vessel", escapeHtml(event.Vessel || "-"))}
            <div class="sc-detail-info-item sc-metric-card sc-detail-remarks">
              <div class="sc-metric-label">Remarks</div>
              <div class="sc-metric-value">${event.Remarks ? escapeHtml(event.Remarks) : '<span class="text-muted">暂无备注。</span>'}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>`;
}

function renderScEventFailedItems(items) {
  const rows = (items || []).map((item) => {
    const entry = item.Entry || {};
    return `<tr>
      <td>${escapeHtml(item.MovieName || "")}</td>
      <td>${escapeHtml(entry.MovieJavId || "")}</td>
      <td class="text-danger small">${escapeHtml(item.LoadError || "")}</td>
    </tr>`;
  }).join("");
  if (!rows) return "";
  return `<div class="card mt-3">
    <div class="card-body">
      <h5 class="mb-3">Skipped / Error Items</h5>
      <div class="table-responsive">
        <table class="table table-sm align-middle mb-0">
          <thead><tr><th>Movie</th><th>JavId</th><th>Error</th></tr></thead>
          <tbody>${rows}</tbody>
        </table>
      </div>
    </div>
  </div>`;
}

function renderScEventEmptyState(movieFilter) {
  if (movieFilter === "all") {
    return '<div class="sc-empty-state">本次 SC 还没有电影明细。</div>';
  }
  if (movieFilter === "nosc") {
    return '<div class="sc-empty-state">当前筛选下没有 is_sc=1 的电影。</div>';
  }
  return '<div class="sc-empty-state">当前筛选下没有 is_sc=2 的电影。</div>';
}

function renderScEventDetail(data) {
  const event = data.Event || {};
  const items = data.Items || [];
  const movieCards = items.map((item) => item.MovieType ? renderMovieCard(item.MovieType) : "").join("");
  const moviesBlock = items.length
    ? `<section aria-label="SC Movies">
        <div class="row row-cols-2 row-cols-sm-3 row-cols-lg-5 row-cols-xxl-6 g-3" id="movie-grid">${movieCards}</div>
      </section>
      ${renderScEventFailedItems(data.FailedItems)}`
    : renderScEventEmptyState(data.MovieFilter);
  return `<div class="d-flex align-items-center justify-content-between mb-3">
      <div>
        <h4 class="mb-1">${escapeHtml(event.Name || "")}</h4>
        <p class="text-muted mb-0">本次 SC 的基础信息、影片清单与快速操作入口。</p>
      </div>
      <div class="sc-action-row">
        <button type="button" class="btn btn-outline-primary" id="btnOpenScEventEditModal" data-bs-toggle="modal" data-bs-target="#scEventEditModal">编辑信息</button>
        <a href="/sc-events" class="btn btn-outline-secondary">返回列表</a>
        <a href="/triggers/sc-media?sc_name=${encodeURIComponent(event.Name || "")}" class="btn btn-outline-primary">移动 SC Media</a>
        <a href="/triggers/backfill" class="btn btn-outline-secondary">回填任务台</a>
      </div>
    </div>
    <div id="scEventPageMsg" class="alert d-none mb-3" role="alert"></div>
    ${renderScEventEditModal(data, event)}
    ${renderScEventDetailInfo(event)}
    <div class="d-flex align-items-center justify-content-between mb-3">
      <div class="d-flex align-items-center gap-3 flex-wrap">
        <h5 class="mb-0">Movies</h5>
        <div class="btn-group btn-group-sm" role="group" aria-label="电影筛选">
          <a href="${attr(data.MovieFilterAllHref)}" class="btn ${data.MovieFilter === "all" ? "btn-primary" : "btn-outline-primary"}">全部电影</a>
          <a href="${attr(data.MovieFilterScHref)}" class="btn ${data.MovieFilter === "sc" ? "btn-primary" : "btn-outline-primary"}">有 SC</a>
          <a href="${attr(data.MovieFilterNoScHref)}" class="btn ${data.MovieFilter === "nosc" ? "btn-primary" : "btn-outline-primary"}">无 SC</a>
        </div>
      </div>
      <span class="text-muted small">${escapeHtml(items.length)} items</span>
    </div>
    ${moviesBlock}`;
}

function renderScTriggers(data) {
  return `<div class="d-flex align-items-center justify-content-between mb-3">
      <div>
        <h4 class="mb-1">SC Triggers</h4>
        <p class="text-muted mb-0">在这里执行 SC 事件导入预览、填写和确认提交。</p>
      </div>
      <div class="d-flex gap-2">
        <a href="/triggers/backfill" class="btn btn-outline-secondary">回填任务台</a>
        <a href="/crawler/tasks" class="btn btn-outline-primary">查看任务列表</a>
      </div>
    </div>

    <div class="card">
      <div class="card-body">
        <h4 class="card-title">添加 SC 事件</h4>
        <p class="card-title-desc">扫描配置中的 SC 根目录，并要求其中仅有一个合法事件目录（例如：2026-03-17-20-00）。</p>
        <p class="text-muted small mb-3">SC Root: <code>${escapeHtml(data.ScRootDir || "")}</code></p>
        <button id="btnScAddPreview" type="button" class="btn btn-outline-success">Add Preview</button>
      </div>
    </div>

    <div class="card mt-3 d-none" id="scPreviewPanel">
      <div class="card-body">
        <h4 class="card-title mb-3">添加 SC 事件</h4>
        <form id="formScAddConfirm">
          <div id="scAddPreviewMsg" class="small mb-3"></div>

          <div class="row g-3 mb-3">
            <div class="col-xl-3 col-lg-4 col-md-6">
              <label class="form-label">事件名</label>
              <input type="text" class="form-control" id="scPreviewName" readonly>
            </div>
            <div class="col-xl-3 col-lg-4 col-md-6">
              <label class="form-label">图片</label>
              <input type="text" class="form-control" id="scPreviewImage" readonly>
            </div>
            <div class="col-xl-3 col-lg-4 col-md-6">
              <label class="form-label">影片数</label>
              <input type="text" class="form-control" id="scPreviewMovieCount" readonly>
            </div>
            <div class="col-xl-3 col-lg-4 col-md-6">
              <label class="form-label">SC 影片数</label>
              <input type="text" class="form-control" id="scPreviewSelectedMovieCount" readonly>
            </div>
            <div class="col-xl-3 col-lg-4 col-md-6">
              <label class="form-label">Come Movie</label>
              <select class="form-select" id="scPreviewComeMovie"></select>
            </div>
            <div class="col-xl-3 col-lg-4 col-md-6">
              <label class="form-label">Movie Cast</label>
              <select class="form-select" id="scPreviewMovieCast"></select>
            </div>
            <div class="col-xl-3 col-lg-4 col-md-6">
              <label class="form-label">种类</label>
              <input type="text" class="form-control" id="scPreviewKind" placeholder="可为空">
            </div>
            <div class="col-xl-3 col-lg-4 col-md-6">
              <label class="form-label">时长(分钟)</label>
              <input type="number" class="form-control" id="scPreviewDuration" placeholder="分钟">
            </div>
            <div class="col-xl-3 col-lg-4 col-md-6">
              <label class="form-label">FG</label>
              <input type="text" class="form-control" id="scPreviewFg" placeholder="可为空">
            </div>
            <div class="col-xl-3 col-lg-4 col-md-6">
              <label class="form-label">Vessel</label>
              <input type="text" class="form-control" id="scPreviewVessel" placeholder="可为空">
            </div>
            <div class="col-12">
              <label class="form-label">Remarks</label>
              <textarea class="form-control" id="scPreviewRemarks" rows="3" placeholder="可为空"></textarea>
            </div>
            <div class="col-12">
              <label class="form-label">影片明细</label>
              <div class="sc-preview-movie-grid" id="scPreviewMovieTableBody">
                <div class="text-muted">暂无预览数据</div>
              </div>
            </div>
          </div>

          <div class="d-flex justify-content-end">
            <button type="submit" class="btn btn-success">确认提交</button>
          </div>
        </form>
      </div>
    </div>`;
}

function renderScPickSmart(data) {
  const pageSource = String(data.PageSource || "wmedia").toLowerCase() === "wmedia" ? "wmedia" : "vfilm";
  const castOwnedScRatioPenaltyAlpha = pageSource === "wmedia" ? "0" : "0.7";
  const birthStartLabel = data.BirthStartLabel || "下载日期起 (mbs)";
  const birthEndLabel = data.BirthEndLabel || "下载日期止 (mbe)";
  const birthStartField = data.BirthStartField || "MediaBirthTimeStart";
  const birthEndField = data.BirthEndField || "MediaBirthTimeEnd";
  const dir1Label = data.Dir1Label || "MediaDir1 (md1)";
  const dir2Label = data.Dir2Label || "MediaDir2 (md2)";
  const dir3Label = data.Dir3Label || "MediaDir3 (md3)";
  const dir4Label = data.Dir4Label || "MediaDir4 (md4)";
  const dir1Field = data.Dir1Field || "MediaDir1";
  const dir2Field = data.Dir2Field || "MediaDir2";
  const dir3Field = data.Dir3Field || "MediaDir3";
  const dir4Field = data.Dir4Field || "MediaDir4";

  return `${renderHeader(data)}
    <div class="card">
      <div class="card-body">
        <p class="text-muted mb-2">按多组条件从本地已拥有影片里抽片，并叠加 Cast SC 阻断、Rank 配额和演员去重规则。</p>
        <p class="text-muted small mb-3">${escapeHtml(data.FixedHint || "")}</p>

        <form id="formPickSmart" class="needs-validation" novalidate data-source="${attr(pageSource)}">
          <div class="card border mb-3">
            <div class="card-body">
              <h6 class="mb-3">全局抽取策略</h6>
              <div class="row g-3 align-items-end">
                <div class="col-md-2">
                  <label for="pickCount" class="form-label">抽取数量</label>
                  <input type="number" min="1" class="form-control" id="pickCount" value="8" required>
                </div>
                <div class="col-md-2">
                  <label for="castLastScBlockDays" class="form-label">Cast LastSc 阻断天数</label>
                  <input type="number" min="0" class="form-control" id="castLastScBlockDays" value="60" required>
                </div>
                <div class="col-md-2">
                  <label for="lastScEventBlockDays" class="form-label">LastSc Event 阻断天数</label>
                  <input type="number" min="0" class="form-control" id="lastScEventBlockDays" value="30" required>
                </div>
                <div class="col-md-2">
                  <label for="rank20Min" class="form-label">Rank20 至少</label>
                  <input type="number" min="0" class="form-control" id="rank20Min" value="0" required>
                </div>
                <div class="col-md-2">
                  <label for="rank100Min" class="form-label">Rank100 至少</label>
                  <input type="number" min="0" class="form-control" id="rank100Min">
                </div>
                <div class="col-md-2">
                  <label for="rank500Min" class="form-label">Rank500 至少</label>
                  <input type="number" min="0" class="form-control" id="rank500Min">
                </div>
                <div class="col-md-2">
                  <label for="castLastScPenaltyAlpha" class="form-label">Cast LastSc 惩罚系数</label>
                  <input type="number" min="0" step="0.1" class="form-control" id="castLastScPenaltyAlpha" value="0.7" required>
                </div>
                <div class="col-md-2">
                  <label for="lastScEventPenaltyAlpha" class="form-label">LastSc Event 惩罚系数</label>
                  <input type="number" min="0" step="0.1" class="form-control" id="lastScEventPenaltyAlpha" value="0.7" required>
                </div>
                <div class="col-md-2">
                  <label for="castOwnedScRatioPenaltyAlpha" class="form-label">Cast SC Ratio 惩罚系数</label>
                  <input type="number" min="0" step="0.1" class="form-control" id="castOwnedScRatioPenaltyAlpha" value="${attr(castOwnedScRatioPenaltyAlpha)}" required>
                </div>
                <div class="col-md-2">
                  <label for="movieHasScPenaltyAlpha" class="form-label">Movie Has SC 惩罚系数</label>
                  <input type="number" min="0" step="0.1" class="form-control" id="movieHasScPenaltyAlpha" value="0.7" required>
                </div>
              </div>
            </div>
          </div>

          <div class="row g-3 align-items-center mb-3">
            <div class="col-auto">
              <button type="button" id="btnAddGroup" class="btn btn-outline-secondary">添加一组条件</button>
            </div>
          </div>

          <div id="pickGroups"></div>

          <div class="mt-3">
            <button type="submit" class="btn btn-primary">执行 Smart Pick + Copy</button>
            <button type="button" class="btn btn-outline-primary ms-2" id="btnPickOnly">仅 Smart Pick</button>
          </div>
        </form>
      </div>
    </div>

    <div class="card mt-3" id="copyCard" style="display: none;">
      <div class="card-body">
        <div class="d-flex align-items-center justify-content-between mb-2">
          <h5 class="mb-0">Copy 进度</h5>
          <button type="button" class="btn btn-sm btn-outline-danger" id="btnStopCopy">停止 Copy</button>
        </div>
        <div class="text-muted small mb-2" id="copyMeta"></div>
        <div class="progress mb-2" style="height: 6px;">
          <div class="progress-bar" role="progressbar" id="copyProgress" style="width: 0%;"></div>
        </div>
        <div class="small text-muted" id="copyCurrent"></div>
        <div class="small text-danger mt-1" id="copyError" style="display: none;"></div>
      </div>
    </div>

    <div class="card mt-3" id="pickInfoCard" style="display: none;">
      <div class="card-body">
        <div class="d-flex align-items-center justify-content-between flex-wrap gap-2 mb-3">
          <h5 class="mb-0">Pick 概览</h5>
          <span class="text-muted small" id="pickInfoSummary"></span>
        </div>
        <div class="row g-3" id="pickInfoGrid"></div>
      </div>
    </div>

    <div class="card mt-3" id="pickedCard" style="display: none;">
      <div class="card-body">
        <div class="d-flex align-items-center justify-content-between flex-wrap gap-2 mb-3">
          <h5 class="mb-0">抽取结果</h5>
          <div class="d-flex align-items-center flex-wrap gap-2">
            <div class="btn-group btn-group-sm flex-wrap" id="pickedSortBar" role="group" aria-label="抽取结果排序">
              <button type="button" class="btn btn-outline-primary" data-sort-base="picked" data-default-dir="desc">抽取顺序</button>
              <button type="button" class="btn btn-outline-primary active" data-sort-base="film_birth" data-default-dir="desc">下载时间</button>
              <button type="button" class="btn btn-outline-primary" data-sort-base="releasing" data-default-dir="desc">上映日期</button>
              <button type="button" class="btn btn-outline-primary" data-sort-base="sc" data-default-dir="desc">SC次数</button>
              <button type="button" class="btn btn-outline-primary" data-sort-base="come" data-default-dir="desc">CO次数</button>
              <button type="button" class="btn btn-outline-primary" data-sort-base="rank" data-default-dir="asc">最高排名</button>
              <button type="button" class="btn btn-outline-primary" data-sort-base="score" data-default-dir="desc">评分</button>
            </div>
            <span class="text-muted small" id="pickedCount"></span>
          </div>
        </div>
        <div class="row row-cols-2 row-cols-sm-3 row-cols-lg-5 g-3 picked-grid-8" id="pickedGrid"></div>
      </div>
    </div>

    <template id="pickGroupTpl">
      <div class="card border mb-3 pick-group">
        <div class="card-body">
          <div class="d-flex align-items-center justify-content-between mb-2">
            <h6 class="mb-0">条件组</h6>
            <button type="button" class="btn btn-sm btn-outline-danger btn-remove-group">移除</button>
          </div>

          <div class="row g-3">
            <div class="col-md-3">
              <label class="form-label">权重</label>
              <input type="number" min="1" class="form-control" data-role="weight" value="12" required>
            </div>
            <div class="col-md-3 d-flex align-items-end">
              <button type="button" class="btn btn-sm btn-outline-secondary w-100" data-role="adv-toggle">高级参数</button>
            </div>

            <div class="col-md-3">
              <label class="form-label">放映日期起 (rs)</label>
              <input type="text" class="form-control" data-req-field="ReleasingDateStart" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
            </div>
            <div class="col-md-3">
              <label class="form-label">放映日期止 (re)</label>
              <input type="text" class="form-control" data-req-field="ReleasingDateEnd" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
            </div>
            <div class="col-md-3">
              <label class="form-label">${escapeHtml(birthStartLabel)}</label>
              <input type="text" class="form-control" data-req-field="${attr(birthStartField)}" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
            </div>
            <div class="col-md-3">
              <label class="form-label">${escapeHtml(birthEndLabel)}</label>
              <input type="text" class="form-control" data-req-field="${attr(birthEndField)}" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
            </div>

            <div class="col-md-3">
              <label class="form-label">ScTimes 最小 (scmin)</label>
              <input type="number" class="form-control" data-req-field="ScTimesMin" data-type="int-opt" placeholder="可为空">
            </div>
            <div class="col-md-3">
              <label class="form-label">ScTimes 最大 (scmax)</label>
              <input type="number" class="form-control" data-req-field="ScTimesMax" data-type="int-opt" placeholder="可为空">
            </div>
            <div class="col-md-3">
              <label class="form-label">ComeTimes 最小 (comin)</label>
              <input type="number" class="form-control" data-req-field="ComeTimesMin" data-type="int-opt" placeholder="可为空">
            </div>
            <div class="col-md-3">
              <label class="form-label">ComeTimes 最大 (comax)</label>
              <input type="number" class="form-control" data-req-field="ComeTimesMax" data-type="int-opt" placeholder="可为空">
            </div>

            <div class="col-12 collapse" data-role="adv-panel">
              <div class="row g-3 mt-1">
                <div class="col-md-4">
                  <label class="form-label">演员 (cn)</label>
                  <input type="text" class="form-control" data-req-field="CastNames">
                </div>
                <div class="col-md-4">
                  <label class="form-label">类型 (gn)</label>
                  <input type="text" class="form-control" data-req-field="GenreNames">
                </div>
                <div class="col-md-4">
                  <label class="form-label">导演 (dn)</label>
                  <input type="text" class="form-control" data-req-field="DirectorName">
                </div>
                <div class="col-md-4">
                  <label class="form-label">前缀 (pn)</label>
                  <input type="text" class="form-control" data-req-field="PrefixName">
                </div>
                <div class="col-md-4">
                  <label class="form-label">制作商 (mn)</label>
                  <input type="text" class="form-control" data-req-field="MakerName">
                </div>
                <div class="col-md-4">
                  <label class="form-label">厂牌 (ln)</label>
                  <input type="text" class="form-control" data-req-field="LabelName">
                </div>
                <div class="col-md-4">
                  <label class="form-label">${escapeHtml(dir4Label)}</label>
                  <input type="text" class="form-control" data-req-field="${attr(dir4Field)}">
                </div>
                <div class="col-md-4">
                  <label class="form-label">关键字 (wd)</label>
                  <input type="text" class="form-control" data-req-field="Word" placeholder="可为空">
                </div>

                <div class="col-md-3">
                  <label class="form-label">评分最小 (smin)</label>
                  <input type="number" step="0.1" class="form-control" data-req-field="ScoreMin" data-type="float">
                </div>
                <div class="col-md-3">
                  <label class="form-label">评分最大 (smax)</label>
                  <input type="number" step="0.1" class="form-control" data-req-field="ScoreMax" data-type="float-opt">
                </div>
                <div class="col-md-3">
                  <label class="form-label">观看最小 (vwmin)</label>
                  <input type="number" class="form-control" data-req-field="ViewWatchedMin" data-type="int">
                </div>
                <div class="col-md-3">
                  <label class="form-label">观看最大 (vwmax)</label>
                  <input type="number" class="form-control" data-req-field="ViewWatchedMax" data-type="int-opt">
                </div>

                <div class="col-md-3">
                  <label class="form-label">最后评分起 (lsctmin)</label>
                  <input type="text" class="form-control" data-req-field="LastScTimeMin" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
                </div>
                <div class="col-md-3">
                  <label class="form-label">最后评分止 (lsctmax)</label>
                  <input type="text" class="form-control" data-req-field="LastScTimeMax" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}">
                </div>
                <div class="col-md-3">
                  <label class="form-label">${escapeHtml(dir1Label)}</label>
                  <input type="text" class="form-control" data-req-field="${attr(dir1Field)}">
                </div>
                <div class="col-md-3">
                  <label class="form-label">${escapeHtml(dir2Label)}</label>
                  <input type="text" class="form-control" data-req-field="${attr(dir2Field)}">
                </div>
                <div class="col-md-3">
                  <label class="form-label">${escapeHtml(dir3Label)}</label>
                  <input type="text" class="form-control" data-req-field="${attr(dir3Field)}">
                </div>
                <div class="col-md-3">
                  <label class="form-label">排序 (od)</label>
                  <input type="text" class="form-control" data-req-field="OrderBy" placeholder="sc/lsct/bt...">
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>`;
}

function renderMediaIngest(data) {
  const rootDirs = Array.isArray(data.RootDirs) ? data.RootDirs : [];
  const favoriteRows = Array.isArray(data.FavoriteRows) ? data.FavoriteRows : [];
  const rootList = rootDirs.length
    ? rootDirs.map((dir) => `<li><code>${escapeHtml(dir)}</code></li>`).join("")
    : '<li class="text-muted">未配置 media.rootDirs</li>';
  const favoriteList = favoriteRows.length
    ? favoriteRows.map((row) => {
      const movieName = row.MovieName || row.movie_name || "";
      const movieJavId = row.MovieJavID || row.MovieJavId || row.movie_jav_id || "";
      const label = movieName || movieJavId || "-";
      const sourceLabel = row.SourceLabel || row.source_label || "";
      const size = row.Size || row.size || "";
      return `<tr>
        <td class="text-nowrap">${escapeHtml(label)}</td>
        <td class="text-nowrap">${escapeHtml(sourceLabel)}</td>
        <td class="text-nowrap">${escapeHtml(size || "-")}</td>
      </tr>`;
    }).join("")
    : '<tr><td colspan="3" class="text-center text-muted">暂无下载中数据</td></tr>';
  const table = (id, title, columns) => `<div class="card">
    <div class="card-body">
      <h5 class="card-title">${escapeHtml(title)}</h5>
      <div class="table-responsive">
        <table class="table table-sm table-hover" id="${attr(id)}">
          <thead><tr>${columns.map((column) => `<th>${escapeHtml(column)}</th>`).join("")}</tr></thead>
          <tbody><tr><td colspan="${columns.length}" class="text-center text-muted">暂无数据</td></tr></tbody>
        </table>
      </div>
    </div>
  </div>`;

  return `<div id="mediaIngestPage">
    <div class="row mb-3">
      <div class="col-12 d-flex align-items-center justify-content-between">
        <div>
          <h4 class="mb-1">${escapeHtml(data.PageTitle || data.Title || "")}</h4>
          <p class="text-muted mb-0">${escapeHtml(data.PageNote || "")}</p>
        </div>
        <div class="d-flex gap-2">
          <a href="/torrent-albums?album_name=%E4%B8%8B%E8%BD%BD%E4%B8%AD" class="btn btn-outline-primary btn-sm">下载中列表</a>
          <a href="/triggers/media-rescan" class="btn btn-outline-secondary btn-sm">媒体重扫</a>
          <a href="/triggers/media-rollback" class="btn btn-outline-secondary btn-sm">名称回滚</a>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-lg-4">
        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">操作</h5>
            <div class="d-grid gap-2">
              <button type="button" id="btnPrecheck" class="btn btn-primary">第一段：预处理</button>
              <button type="button" id="btnCommit" class="btn btn-success" disabled>第二段：执行插入</button>
              <button type="button" id="btnReturn" class="btn btn-outline-secondary d-none">返回（清空本次预处理）</button>
            </div>
            <div class="mt-3 small text-muted">
              <div>部分失败时可选择：</div>
              <div>1. 仅插入通过项（点第二段）</div>
              <div>2. 返回后修正再重跑预处理</div>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">运行概览</h5>
            <div class="row g-2">
              <div class="col-4"><div class="overview-item"><div class="label">总数</div><div class="value" id="sumTotal">0</div></div></div>
              <div class="col-4"><div class="overview-item"><div class="label" id="sumPassLabel">通过</div><div class="value text-success" id="sumPass">0</div></div></div>
              <div class="col-4"><div class="overview-item"><div class="label" id="sumFailLabel">失败</div><div class="value text-danger" id="sumFail">0</div></div></div>
            </div>
            <div class="mt-3">
              <div class="text-muted small">状态</div>
              <div id="statusText" class="fw-semibold">等待操作</div>
            </div>
            <div class="mt-2 text-muted small" id="generatedAtText"></div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">Root 目录</h5>
            <ul class="mb-0 root-list">${rootList}</ul>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">下载中（tm_album_item）</h5>
            <div class="table-responsive">
              <table class="table table-sm table-hover align-middle mb-0">
                <thead><tr><th>影片</th><th>来源</th><th>大小</th></tr></thead>
                <tbody>${favoriteList}</tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <div class="col-lg-8">
        ${table("tblPrecheckPass", "预处理通过 Movie", ["影片", "源文件", "目标名", "目标目录", "Alias", "Hash"])}
        ${table("tblPrecheckFail", "预处理失败 Movie", ["影片", "源文件", "失败原因"])}
        ${table("tblPrecheckPreview", "插入前预览（通过项）", ["影片", "目标目录", "目标名", "Alias", "大小"])}
        <div class="card" id="commitSuccessSection">
          <div class="card-body">
            <h5 class="card-title">插入成功 Movie</h5>
            <div class="table-responsive">
              <table class="table table-sm table-hover" id="tblCommitSuccess">
                <thead><tr><th>影片</th><th>目标路径</th><th>Alias</th><th>Hash</th></tr></thead>
                <tbody><tr><td colspan="4" class="text-center text-muted">暂无数据</td></tr></tbody>
              </table>
            </div>
          </div>
        </div>
        <div class="card" id="commitFailSection">
          <div class="card-body">
            <h5 class="card-title">插入失败 Movie</h5>
            <div class="table-responsive">
              <table class="table table-sm table-hover" id="tblCommitFail">
                <thead><tr><th>影片</th><th>源文件</th><th>失败原因</th></tr></thead>
                <tbody><tr><td colspan="3" class="text-center text-muted">暂无数据</td></tr></tbody>
              </table>
            </div>
          </div>
        </div>
        <div class="card" id="commitPreviewSection">
          <div class="card-body">
            <h5 class="card-title">插入后预览（成功项）</h5>
            <div class="table-responsive">
              <table class="table table-sm table-hover" id="tblCommitPreview">
                <thead><tr><th>影片</th><th>目标路径</th><th>Alias</th><th>大小</th></tr></thead>
                <tbody><tr><td colspan="4" class="text-center text-muted">暂无数据</td></tr></tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>`;
}

function renderMediaRescan(data) {
  const rootOptions = Array.isArray(data.RootOptions) ? data.RootOptions : [];
  const renderScope = (root, scope) => {
    const scopeValue = scope.scope || scope.Scope || "";
    const label = scope.label || scope.Label || "";
    const baseDir = scope.base_dir || scope.BaseDir || "";
    const exists = Boolean(scope.exists ?? scope.Exists);
    const children = Array.isArray(scope.children || scope.Children) ? (scope.children || scope.Children) : [];
    const branchRows = children.length
      ? children.map((child) => {
        const name = child.name || child.Name || "";
        return `<label class="branch-checkbox-item">
          <input type="checkbox" class="form-check-input branch-checkbox" data-root="${attr(root)}" data-scope="${attr(scopeValue)}" data-branch="${attr(name)}">
          <code>${escapeHtml(name)}</code>
        </label>`;
      }).join("")
      : '<div class="text-muted small">暂无可选二级目录</div>';
    return `<div class="root-base-item">
      <label class="root-checkbox-item">
        <input type="checkbox" class="form-check-input root-checkbox" value="${attr(root)}" data-root="${attr(root)}" data-scope="${attr(scopeValue)}" checked>
        <span>
          <code>${escapeHtml(label)}</code>
          <span class="text-muted small ms-1">${exists ? `扫描基目录：${escapeHtml(baseDir)}` : "当前未挂载"}</span>
        </span>
      </label>
      <div class="root-branch-list">
        <div class="text-muted small mb-1">二级目录：不选则扫描整个 ${escapeHtml(label)}</div>
        ${branchRows}
      </div>
    </div>`;
  };
  const rootList = rootOptions.length
    ? rootOptions.map((option) => {
      const root = option.root || option.Root || "";
      const scopes = Array.isArray(option.scopes || option.Scopes) ? (option.scopes || option.Scopes) : [];
      return `<div class="root-scope-item" data-root="${attr(root)}">
        <div class="fw-semibold mb-2"><code>${escapeHtml(root)}</code></div>
        <div class="root-base-list">${scopes.map((scope) => renderScope(root, scope)).join("")}</div>
      </div>`;
    }).join("")
    : '<div class="text-muted">未配置 media.rootDirs</div>';
  const table = (id, title, columns) => `<div class="card">
    <div class="card-body">
      <h5 class="card-title">${escapeHtml(title)}</h5>
      <div class="table-responsive">
        <table class="table table-sm table-hover" id="${attr(id)}">
          <thead><tr>${columns.map((column) => `<th>${escapeHtml(column)}</th>`).join("")}</tr></thead>
          <tbody><tr><td colspan="${columns.length}" class="text-center text-muted">暂无数据</td></tr></tbody>
        </table>
      </div>
    </div>
  </div>`;

  return `<div id="mediaRescanPage">
    <div class="row mb-3">
      <div class="col-12 d-flex align-items-center justify-content-between">
        <div>
          <h4 class="mb-1">${escapeHtml(data.PageTitle || data.Title || "")}</h4>
          <p class="text-muted mb-0">${escapeHtml(data.PageNote || "")}</p>
        </div>
        <div class="d-flex gap-2">
          <a href="/triggers/media" class="btn btn-outline-primary btn-sm">媒体入库</a>
          <a href="/triggers/media-rollback" class="btn btn-outline-secondary btn-sm">名称回滚</a>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-lg-4">
        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">操作</h5>
            <div class="d-grid gap-2">
              <button type="button" id="btnSelectAll" class="btn btn-outline-secondary">全选</button>
              <button type="button" id="btnClearAll" class="btn btn-outline-secondary">清空</button>
              <button type="button" id="btnRescan" class="btn btn-primary">开始重扫</button>
            </div>
            <div class="mt-3 small text-muted">
              只会对本次勾选且当前存在的扫描范围做删除判断；基目录已选但未选二级目录时，表示扫描整个所选基目录。
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">扫描目录</h5>
            <div class="root-scope-list">${rootList}</div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">运行概览</h5>
            <div class="row g-2">
              <div class="col-6"><div class="overview-item"><div class="label">扫描文件</div><div class="value" id="sumTotalFiles">0</div></div></div>
              <div class="col-6"><div class="overview-item"><div class="label">匹配成功</div><div class="value" id="sumMatched">0</div></div></div>
              <div class="col-6"><div class="overview-item"><div class="label">位置变更</div><div class="value text-primary" id="sumMoved">0</div></div></div>
              <div class="col-6"><div class="overview-item"><div class="label">恢复记录</div><div class="value text-success" id="sumRestored">0</div></div></div>
              <div class="col-6"><div class="overview-item"><div class="label">标记删除</div><div class="value text-warning" id="sumRemoved">0</div></div></div>
              <div class="col-6"><div class="overview-item"><div class="label">未匹配</div><div class="value text-secondary" id="sumUnmatched">0</div></div></div>
              <div class="col-6"><div class="overview-item"><div class="label">跳过范围</div><div class="value text-muted" id="sumSkippedRoots">0</div></div></div>
              <div class="col-6"><div class="overview-item"><div class="label">错误</div><div class="value text-danger" id="sumErrors">0</div></div></div>
            </div>
            <div class="mt-3">
              <div class="text-muted small">状态</div>
              <div id="statusText" class="fw-semibold">等待操作</div>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">本次目录结果</h5>
            <div class="mb-2"><div class="text-muted small">已选择 root</div><ul class="root-list mb-0" id="selectedRootList"><li class="text-muted">暂无数据</li></ul></div>
            <div class="mb-2"><div class="text-muted small">实际扫描 root</div><ul class="root-list mb-0" id="scannedRootList"><li class="text-muted">暂无数据</li></ul></div>
            <div class="mb-2"><div class="text-muted small">已跳过 root</div><ul class="root-list mb-0" id="skippedRootList"><li class="text-muted">暂无数据</li></ul></div>
            <div class="mb-2"><div class="text-muted small">已选择范围</div><ul class="root-list mb-0" id="selectedTargetList"><li class="text-muted">暂无数据</li></ul></div>
            <div class="mb-2"><div class="text-muted small">实际扫描范围</div><ul class="root-list mb-0" id="scannedTargetList"><li class="text-muted">暂无数据</li></ul></div>
            <div><div class="text-muted small">已跳过范围</div><ul class="root-list mb-0" id="skippedTargetList"><li class="text-muted">暂无数据</li></ul></div>
          </div>
        </div>
      </div>

      <div class="col-lg-8">
        ${table("tblMoved", "位置变更", ["影片", "匹配方式", "旧路径", "新路径"])}
        ${table("tblRestored", "恢复记录", ["影片", "匹配方式", "旧路径", "当前路径"])}
        ${table("tblRemoved", "标记删除", ["影片", "原 root", "原路径"])}
        ${table("tblUnmatched", "未匹配文件", ["文件名", "root", "路径"])}
        ${table("tblErrors", "错误", ["文件/影片", "路径", "错误"])}
      </div>
    </div>
  </div>`;
}

function renderMediaRollback(data) {
  const rootDirs = Array.isArray(data.RootDirs) ? data.RootDirs : [];
  const rootList = rootDirs.length
    ? rootDirs.map((dir) => `<li><code>${escapeHtml(dir)}</code></li>`).join("")
    : '<li class="text-muted">未配置 media.rootDirs</li>';
  const table = (id, title, columns) => `<div class="card">
    <div class="card-body">
      <h5 class="card-title">${escapeHtml(title)}</h5>
      <div class="table-responsive">
        <table class="table table-sm table-hover" id="${attr(id)}">
          <thead><tr>${columns.map((column) => `<th>${escapeHtml(column)}</th>`).join("")}</tr></thead>
          <tbody><tr><td colspan="${columns.length}" class="text-center text-muted">暂无数据</td></tr></tbody>
        </table>
      </div>
    </div>
  </div>`;

  return `<div id="mediaRollbackPage">
    <div class="row mb-3">
      <div class="col-12 d-flex align-items-center justify-content-between">
        <div>
          <h4 class="mb-1">${escapeHtml(data.PageTitle || data.Title || "")}</h4>
          <p class="text-muted mb-0">${escapeHtml(data.PageNote || "")}</p>
        </div>
        <div class="d-flex gap-2">
          <a href="/triggers/media" class="btn btn-outline-primary btn-sm">返回入库页</a>
          <a href="/triggers/media-rescan" class="btn btn-outline-secondary btn-sm">媒体重扫</a>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-lg-4">
        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">操作</h5>
            <div class="d-grid gap-2">
              <button type="button" id="btnRollback" class="btn btn-primary">开始回滚名称</button>
            </div>
            <div class="mt-3 small text-muted">
              从每个 root 的 <code>001_process/004_rollback</code> 读取文件，按规则还原名称并移回 <code>001_process/001_ingest_new</code>。
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">运行概览</h5>
            <div class="row g-2">
              <div class="col-4"><div class="overview-item"><div class="label">总数</div><div class="value" id="sumTotal">0</div></div></div>
              <div class="col-4"><div class="overview-item"><div class="label">成功</div><div class="value text-success" id="sumSuccess">0</div></div></div>
              <div class="col-4"><div class="overview-item"><div class="label">失败</div><div class="value text-danger" id="sumFail">0</div></div></div>
            </div>
            <div class="mt-3">
              <div class="text-muted small">状态</div>
              <div id="statusText" class="fw-semibold">等待操作</div>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">Root 目录</h5>
            <ul class="mb-0 root-list">${rootList}</ul>
          </div>
        </div>
      </div>

      <div class="col-lg-8">
        ${table("tblRollbackSuccess", "回滚成功 Movie", ["影片", "源文件", "目标文件"])}
        ${table("tblRollbackFail", "回滚失败 Movie", ["影片", "源文件", "失败原因", "失败文件位置"])}
        ${table("tblRollbackPreview", "回滚后预览（成功项）", ["影片", "回滚后路径"])}
      </div>
    </div>
  </div>`;
}

function renderScMediaMove(data) {
  const mode = data.Mode === "all" ? "all" : "single";
  const scName = data.ScName || "";
  const rootDirs = Array.isArray(data.RootDirs) ? data.RootDirs : [];
  const rootList = rootDirs.length
    ? rootDirs.map((dir) => `<li><code>${escapeHtml(dir)}</code></li>`).join("")
    : '<li class="text-muted">未配置 media.rootDirs</li>';
  const navButton = data.ScDetailHref
    ? `<a href="${attr(data.ScDetailHref)}" class="btn btn-outline-primary btn-sm">返回当前 SC</a>`
    : '<a href="/sc-events" class="btn btn-outline-primary btn-sm">SC Events</a>';
  const table = (id, title, columns) => `<div class="card">
    <div class="card-body">
      <h5 class="card-title">${escapeHtml(title)}</h5>
      <div class="table-responsive">
        <table class="table table-sm table-hover" id="${attr(id)}">
          <thead><tr>${columns.map((column) => `<th>${escapeHtml(column)}</th>`).join("")}</tr></thead>
          <tbody><tr><td colspan="${columns.length}" class="text-center text-muted">暂无数据</td></tr></tbody>
        </table>
      </div>
    </div>
  </div>`;

  return `<div id="scMediaMovePage">
    <div class="row mb-3">
      <div class="col-12 d-flex align-items-center justify-content-between">
        <div>
          <h4 class="mb-1">${escapeHtml(data.PageTitle || data.Title || "")}</h4>
          <p class="text-muted mb-0">${escapeHtml(data.PageNote || "")}</p>
        </div>
        <div class="d-flex gap-2">
          ${navButton}
          <a href="/triggers/media" class="btn btn-outline-secondary btn-sm">媒体入库</a>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-lg-4">
        <div class="card">
          <div class="card-body">
            <label class="form-label d-block mb-2">执行范围</label>
            <div class="d-flex flex-wrap gap-3 mb-3">
              <div class="form-check">
                <input class="form-check-input" type="radio" name="scopeMode" id="scopeModeSingle" value="single" ${mode === "all" ? "" : "checked"}>
                <label class="form-check-label" for="scopeModeSingle">单个 SC</label>
              </div>
              <div class="form-check">
                <input class="form-check-input" type="radio" name="scopeMode" id="scopeModeAll" value="all" ${mode === "all" ? "checked" : ""}>
                <label class="form-check-label" for="scopeModeAll">全部符合条件</label>
              </div>
            </div>

            <div id="scNameGroup" ${mode === "all" ? 'class="d-none"' : ""}>
              <label class="form-label" for="scNameInput">SC 名称</label>
              <input type="text" class="form-control" id="scNameInput" value="${attr(scName)}" placeholder="例如 2026-04-01-22-30">
              <div class="form-text">可以从 SC 详情页进入，也可以手动输入后执行第一段预处理。</div>
            </div>
            <div id="allModeHint" class="form-text ${mode === "all" ? "" : "d-none"}">
              将汇总所有 SC 事件中 <code>is_sc=2</code> 的电影，按影片去重后统一预处理与移动。
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">操作</h5>
            <div class="d-grid gap-2">
              <button type="button" id="btnPrecheck" class="btn btn-primary">第一段：预处理</button>
              <button type="button" id="btnCommit" class="btn btn-success" disabled>第二段：执行移动</button>
              <button type="button" id="btnReturn" class="btn btn-outline-secondary d-none">返回（清空本次预处理）</button>
            </div>
            <div class="mt-3 small text-muted">
              <div>第一段只读取本次范围内涉及到的 WMedia 并预览目标目录。</div>
              <div>第二段只移动预处理通过项到同 root 的 <code>watched</code>。</div>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">运行概览</h5>
            <div class="row g-2">
              <div class="col-3"><div class="overview-item"><div class="label">总数</div><div class="value" id="sumTotal">0</div></div></div>
              <div class="col-3"><div class="overview-item"><div class="label">可移动</div><div class="value text-success" id="sumPass">0</div></div></div>
              <div class="col-3"><div class="overview-item"><div class="label">跳过</div><div class="value text-secondary" id="sumSkip">0</div></div></div>
              <div class="col-3"><div class="overview-item"><div class="label">失败</div><div class="value text-danger" id="sumFail">0</div></div></div>
            </div>
            <div class="mt-3">
              <div class="text-muted small">状态</div>
              <div id="statusText" class="fw-semibold">等待操作</div>
            </div>
            <div class="mt-2 text-muted small" id="generatedAtText"></div>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h5 class="card-title mb-3">Root 目录</h5>
            <ul class="mb-0 root-list">${rootList}</ul>
          </div>
        </div>
      </div>

      <div class="col-lg-8">
        ${table("tblPrecheckPass", "预处理通过 Media", ["SC Events", "影片", "源路径", "目标路径"])}
        ${table("tblPrecheckSkip", "预处理跳过 Media", ["SC Events", "影片", "源路径", "原因"])}
        ${table("tblPrecheckFail", "预处理失败 Media", ["SC Events", "影片", "源路径", "失败原因"])}
        ${table("tblCommitSuccess", "移动成功 Media", ["SC Events", "影片", "源路径", "目标路径"])}
        ${table("tblCommitFail", "移动失败 Media", ["SC Events", "影片", "源路径", "失败原因"])}
      </div>
    </div>
  </div>`;
}

function writeOutput(id, value) {
  const el = document.getElementById(id);
  if (el) el.textContent = typeof value === "string" ? value : JSON.stringify(value, null, 2);
}

function bindScTriggers() {
  const btnScAddPreview = document.getElementById("btnScAddPreview");
  const scPreviewPanel = document.getElementById("scPreviewPanel");
  const formScAddConfirm = document.getElementById("formScAddConfirm");
  const scPreviewMsg = document.getElementById("scAddPreviewMsg");
  const scPreviewName = document.getElementById("scPreviewName");
  const scPreviewImage = document.getElementById("scPreviewImage");
  const scPreviewMovieCount = document.getElementById("scPreviewMovieCount");
  const scPreviewSelectedMovieCount = document.getElementById("scPreviewSelectedMovieCount");
  const scPreviewMovieTableBody = document.getElementById("scPreviewMovieTableBody");
  const scPreviewComeMovie = document.getElementById("scPreviewComeMovie");
  const scPreviewMovieCast = document.getElementById("scPreviewMovieCast");
  const scPreviewKind = document.getElementById("scPreviewKind");
  const scPreviewDuration = document.getElementById("scPreviewDuration");
  const scPreviewFg = document.getElementById("scPreviewFg");
  const scPreviewVessel = document.getElementById("scPreviewVessel");
  const scPreviewRemarks = document.getElementById("scPreviewRemarks");
  let scPreviewData = null;

  function setScPreviewMsg(text, ok) {
    if (!scPreviewMsg) return;
    if (!text) {
      scPreviewMsg.className = "small mb-3";
      scPreviewMsg.textContent = "";
      return;
    }
    scPreviewMsg.className = `alert ${ok ? "alert-success" : "alert-danger"} small py-2 px-3 mb-3`;
    scPreviewMsg.textContent = text;
  }

  function normalizePreviewMovies(movies) {
    if (!Array.isArray(movies)) {
      return [];
    }
    return movies.map((movie) => ({
      movieName: movie && movie.movieName ? movie.movieName : "",
      movieJavId: movie && movie.movieJavId ? movie.movieJavId : "",
      movieHref: movie && movie.movieHref ? movie.movieHref : "",
      jacketImg: movie && movie.jacketImg ? movie.jacketImg : "",
      casts: Array.isArray(movie && movie.casts) ? movie.casts.slice() : [],
      isSc: GLIST_IS_SC
    }));
  }

  function getSelectedScMovies() {
    if (!scPreviewData || !Array.isArray(scPreviewData.movies)) {
      return [];
    }
    return scPreviewData.movies.filter((movie) => movie && movie.isSc === GLIST_IS_SC);
  }

  function findMovieByJavId(movieJavId) {
    if (!scPreviewData || !Array.isArray(scPreviewData.movies)) {
      return null;
    }
    return scPreviewData.movies.find((item) => item.movieJavId === movieJavId) || null;
  }

  function buildCastOptions(movieJavId) {
    if (!scPreviewMovieCast) return;
    scPreviewMovieCast.innerHTML = "";
    const movie = findMovieByJavId(movieJavId);

    const emptyOption = document.createElement("option");
    emptyOption.value = "";
    emptyOption.textContent = "可为空";
    scPreviewMovieCast.appendChild(emptyOption);

    if (!movie || !Array.isArray(movie.casts) || movie.casts.length === 0) {
      return;
    }
    movie.casts.forEach((name) => {
      const option = document.createElement("option");
      option.value = name;
      option.textContent = name;
      scPreviewMovieCast.appendChild(option);
    });
  }

  function syncSelectedMovieCount() {
    if (!scPreviewSelectedMovieCount) return;
    scPreviewSelectedMovieCount.value = String(getSelectedScMovies().length);
  }

  function buildIsScRadio(groupName, idx, value, checked) {
    const inputId = `${groupName}-${value}`;
    return `<div>
      <input type="radio"
             class="btn-check sc-preview-is-sc-input-${value}"
             name="${attr(groupName)}"
             value="${attr(value)}"
             ${checked ? "checked" : ""}
             autocomplete="off"
             id="${attr(inputId)}"
             data-preview-movie-idx="${attr(idx)}"
             data-preview-is-sc="${attr(value)}">
      <label class="btn btn-outline-secondary sc-preview-is-sc-btn sc-preview-is-sc-btn-${value}" for="${attr(inputId)}">${value === GLIST_IS_SC ? "计入SC" : "不计入"}</label>
    </div>`;
  }

  function renderMovieTable() {
    if (!scPreviewMovieTableBody) return;
    if (!scPreviewData || !Array.isArray(scPreviewData.movies) || scPreviewData.movies.length === 0) {
      scPreviewMovieTableBody.innerHTML = '<div class="text-muted">暂无预览数据</div>';
      return;
    }

    scPreviewMovieTableBody.innerHTML = scPreviewData.movies.map((movie, idx) => {
      const poster = movie.jacketImg
        ? `<img class="sc-preview-movie-poster" src="${attr(movie.jacketImg)}" alt="${attr((movie.movieName || "") + " 海报")}" loading="lazy" decoding="async">`
        : '<div class="sc-preview-movie-poster-empty"></div>';
      const posterWrap = movie.movieHref ? `<a href="${attr(movie.movieHref)}">${poster}</a>` : `<div>${poster}</div>`;
      const title = movie.movieHref
        ? `<a href="${attr(movie.movieHref)}" class="sc-preview-movie-link fw-semibold">${escapeHtml(movie.movieName || "")}</a>`
        : `<div class="sc-preview-movie-link fw-semibold">${escapeHtml(movie.movieName || "")}</div>`;
      const radioName = `sc-preview-is-sc-${idx}`;
      return `<div>
        <div class="sc-preview-movie-card">
          ${posterWrap}
          ${title}
          <div class="sc-preview-is-sc-buttons">
            ${buildIsScRadio(radioName, idx, GLIST_IS_NOT_SC, movie.isSc === GLIST_IS_NOT_SC)}
            ${buildIsScRadio(radioName, idx, GLIST_IS_SC, movie.isSc === GLIST_IS_SC)}
          </div>
        </div>
      </div>`;
    }).join("");

    scPreviewMovieTableBody.querySelectorAll("[data-preview-movie-idx]").forEach((input) => {
      input.addEventListener("change", () => {
        const targetIndex = parseInt(String(input.getAttribute("data-preview-movie-idx") || ""), 10);
        const value = Number(input.getAttribute("data-preview-is-sc") || 0);
        if (!Number.isFinite(targetIndex) || !scPreviewData || !Array.isArray(scPreviewData.movies)) {
          return;
        }
        const currentMovie = scPreviewData.movies[targetIndex];
        if (!currentMovie) {
          return;
        }
        currentMovie.isSc = value === GLIST_IS_SC ? GLIST_IS_SC : GLIST_IS_NOT_SC;
        rebuildComeMovieOptions();
      });
    });
  }

  function rebuildComeMovieOptions() {
    if (!scPreviewComeMovie) return;
    const selectedScMovies = getSelectedScMovies();
    const previousValue = (scPreviewComeMovie.value || "").trim();
    scPreviewComeMovie.innerHTML = "";

    const emptyOption = document.createElement("option");
    emptyOption.value = "";
    emptyOption.textContent = "可为空";
    scPreviewComeMovie.appendChild(emptyOption);

    selectedScMovies.forEach((movie) => {
      const option = document.createElement("option");
      option.value = movie.movieJavId;
      option.textContent = movie.movieName;
      if (movie.movieJavId === previousValue) {
        option.selected = true;
      }
      scPreviewComeMovie.appendChild(option);
    });

    if (![...scPreviewComeMovie.options].some((option) => option.selected)) {
      scPreviewComeMovie.options[0].selected = true;
    }
    buildCastOptions(scPreviewComeMovie.value);
    syncSelectedMovieCount();
  }

  function fillScPreview(data) {
    scPreviewData = {
      scName: data.scName || "",
      imageFound: !!data.imageFound,
      imageName: data.imageName || "",
      movieCount: data.movieCount || 0,
      movies: normalizePreviewMovies(data.movies)
    };
    if (scPreviewName) scPreviewName.value = scPreviewData.scName;
    if (scPreviewImage) scPreviewImage.value = scPreviewData.imageFound ? (scPreviewData.imageName || "已发现图片") : "无图片";
    if (scPreviewMovieCount) scPreviewMovieCount.value = String(scPreviewData.movieCount || 0);
    if (scPreviewKind) scPreviewKind.value = "";
    if (scPreviewDuration) scPreviewDuration.value = "";
    if (scPreviewFg) scPreviewFg.value = "";
    if (scPreviewVessel) scPreviewVessel.value = "";
    if (scPreviewRemarks) scPreviewRemarks.value = "";
    setScPreviewMsg("", true);

    renderMovieTable();
    rebuildComeMovieOptions();
  }

  function buildPayload() {
    const payload = {
      comeMovieJavId: (scPreviewComeMovie && scPreviewComeMovie.value || "").trim(),
      movieCast: (scPreviewMovieCast && scPreviewMovieCast.value || "").trim(),
      kind: (scPreviewKind && scPreviewKind.value || "").trim(),
      fg: (scPreviewFg && scPreviewFg.value || "").trim(),
      vessel: (scPreviewVessel && scPreviewVessel.value || "").trim(),
      remarks: (scPreviewRemarks && scPreviewRemarks.value || "").trim(),
      movies: (scPreviewData && Array.isArray(scPreviewData.movies) ? scPreviewData.movies : []).map((movie) => ({
        movieJavId: movie.movieJavId,
        isSc: movie.isSc === GLIST_IS_SC ? GLIST_IS_SC : GLIST_IS_NOT_SC
      }))
    };

    const rawDurationMinutes = (scPreviewDuration && scPreviewDuration.value || "").trim();
    if (rawDurationMinutes === "") {
      payload.duration = 0;
    } else {
      const parsedDurationMinutes = parseInt(rawDurationMinutes, 10);
      if (!Number.isFinite(parsedDurationMinutes)) {
        throw new Error("时长(分钟)必须是数字");
      }
      payload.duration = parsedDurationMinutes;
    }
    return payload;
  }

  scPreviewComeMovie?.addEventListener("change", () => {
    buildCastOptions(scPreviewComeMovie.value);
  });

  btnScAddPreview?.addEventListener("click", async () => {
    try {
      const data = await fetchJSON("/api/triggers/sc/add-preview", {method: "POST"});
      fillScPreview(data);
      if (scPreviewPanel) {
        scPreviewPanel.classList.remove("d-none");
        scPreviewPanel.scrollIntoView({behavior: "smooth", block: "start"});
      }
    } catch (err) {
      showMessage(`异常：${err.message}`, "danger");
    }
  });

  formScAddConfirm?.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!scPreviewData) {
      setScPreviewMsg("请先执行 Add Preview", false);
      return;
    }

    let payload;
    try {
      payload = buildPayload();
    } catch (err) {
      setScPreviewMsg(err.message || "提交参数有误", false);
      return;
    }

    try {
      const data = await fetchJSON("/api/triggers/sc/add", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(payload)
      });
      setScPreviewMsg("AddSc 任务已启动，正在跳转任务列表", true);
      showMessage("AddSc 任务已启动");
      if (data && data.job_id) {
        window.setTimeout(() => {
          window.location.href = `/crawler/tasks?job_id=${encodeURIComponent(data.job_id)}`;
        }, 300);
      }
    } catch (err) {
      setScPreviewMsg(`异常：${err.message}`, false);
    }
  });
}

function bindSmartPick() {
  const form = document.getElementById("formPickSmart");
  const groupWrap = document.getElementById("pickGroups");
  const groupTpl = document.getElementById("pickGroupTpl");
  const btnAdd = document.getElementById("btnAddGroup");
  const btnPickOnly = document.getElementById("btnPickOnly");
  const pickedCard = document.getElementById("pickedCard");
  const pickedGrid = document.getElementById("pickedGrid");
  const pickedCount = document.getElementById("pickedCount");
  const pickedSortBar = document.getElementById("pickedSortBar");
  const pickInfoCard = document.getElementById("pickInfoCard");
  const pickInfoSummary = document.getElementById("pickInfoSummary");
  const pickInfoGrid = document.getElementById("pickInfoGrid");
  const copyCard = document.getElementById("copyCard");
  const copyMeta = document.getElementById("copyMeta");
  const copyProgress = document.getElementById("copyProgress");
  const copyCurrent = document.getElementById("copyCurrent");
  const copyError = document.getElementById("copyError");
  const btnStopCopy = document.getElementById("btnStopCopy");
  if (!form || !groupWrap || !groupTpl || !btnAdd) return;

  const pageSource = String(form.getAttribute("data-source") || "vfilm").toLowerCase() === "wmedia" ? "wmedia" : "vfilm";
  let groupSeq = 0;
  let pickedMoviesCache = [];
  let pickedTotalSizeGb = "";
  let pickedSortBase = "film_birth";
  let pickedSortDir = "desc";
  let copyTimer = null;
  let copyStatusBootstrapped = false;
  let copyStartedInPage = false;

  function post(url, body) {
    return fetch(url, {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(body)
    });
  }

  function bindRemove(btn) {
    btn.addEventListener("click", () => {
      const group = btn.closest(".pick-group");
      if (!group) return;
      if (groupWrap.children.length <= 1) {
        showMessage("至少保留一组条件", "danger");
        return;
      }
      group.remove();
    });
  }

  function setupAdvanced(group) {
    const toggle = group.querySelector('[data-role="adv-toggle"]');
    const panel = group.querySelector('[data-role="adv-panel"]');
    if (!toggle || !panel) return;
    groupSeq += 1;
    const id = `smart-adv-panel-${groupSeq}`;
    panel.id = id;
    toggle.setAttribute("data-bs-toggle", "collapse");
    toggle.setAttribute("data-bs-target", `#${id}`);
    toggle.setAttribute("aria-controls", id);
    toggle.setAttribute("aria-expanded", "false");
  }

  function addGroup() {
    const node = document.importNode(groupTpl.content, true);
    const group = node.querySelector(".pick-group");
    const btnRemove = node.querySelector(".btn-remove-group");
    if (btnRemove) bindRemove(btnRemove);
    if (group) setupAdvanced(group);
    groupWrap.appendChild(group);
  }

  function parseNumber(raw, isFloat) {
    const value = isFloat ? Number.parseFloat(raw) : Number.parseInt(raw, 10);
    return Number.isFinite(value) ? value : null;
  }

  function numberOrZero(value) {
    return Number.isFinite(value) ? value : 0;
  }

  function parseDateValue(raw) {
    if (!raw) return 0;
    const ts = Date.parse(raw);
    return Number.isFinite(ts) ? ts : 0;
  }

  function rankValue(movie) {
    const value = Number.isFinite(movie?.highest_rank) ? movie.highest_rank : 0;
    return value > 0 ? value : Number.MAX_SAFE_INTEGER;
  }

  function compareText(a, b) {
    return String(a || "").localeCompare(String(b || ""), "zh-Hans-CN", {
      numeric: true,
      sensitivity: "base"
    });
  }

  function comparePickedOrder(a, b) {
    return numberOrZero(a?._pickedOrder) - numberOrZero(b?._pickedOrder);
  }

  function sortPickedMovies(list, sortKey) {
    const sorted = list.slice();
    switch (sortKey) {
      case "film_birth_desc":
        sorted.sort((a, b) => parseDateValue(b.film_birth_date) - parseDateValue(a.film_birth_date) || comparePickedOrder(a, b));
        break;
      case "film_birth_asc":
        sorted.sort((a, b) => parseDateValue(a.film_birth_date) - parseDateValue(b.film_birth_date) || comparePickedOrder(a, b));
        break;
      case "releasing_desc":
        sorted.sort((a, b) => parseDateValue(b.releasing_date) - parseDateValue(a.releasing_date) || comparePickedOrder(a, b));
        break;
      case "releasing_asc":
        sorted.sort((a, b) => parseDateValue(a.releasing_date) - parseDateValue(b.releasing_date) || comparePickedOrder(a, b));
        break;
      case "sc_desc":
        sorted.sort((a, b) => numberOrZero(b.sc_times) - numberOrZero(a.sc_times) || comparePickedOrder(a, b));
        break;
      case "sc_asc":
        sorted.sort((a, b) => numberOrZero(a.sc_times) - numberOrZero(b.sc_times) || comparePickedOrder(a, b));
        break;
      case "come_desc":
        sorted.sort((a, b) => numberOrZero(b.come_times) - numberOrZero(a.come_times) || comparePickedOrder(a, b));
        break;
      case "come_asc":
        sorted.sort((a, b) => numberOrZero(a.come_times) - numberOrZero(b.come_times) || comparePickedOrder(a, b));
        break;
      case "rank_best":
        sorted.sort((a, b) => rankValue(a) - rankValue(b) || comparePickedOrder(a, b));
        break;
      case "rank_worst":
        sorted.sort((a, b) => rankValue(b) - rankValue(a) || comparePickedOrder(a, b));
        break;
      case "score_desc":
        sorted.sort((a, b) => numberOrZero(b.score) - numberOrZero(a.score) || comparePickedOrder(a, b));
        break;
      case "score_asc":
        sorted.sort((a, b) => numberOrZero(a.score) - numberOrZero(b.score) || comparePickedOrder(a, b));
        break;
      case "name_asc":
        sorted.sort((a, b) => compareText(a.name, b.name) || comparePickedOrder(a, b));
        break;
      default:
        sorted.sort(comparePickedOrder);
        break;
    }
    return sorted;
  }

  function currentSortKey() {
    if (pickedSortBase === "picked") return "picked";
    if (pickedSortBase === "rank") return pickedSortDir === "desc" ? "rank_worst" : "rank_best";
    return `${pickedSortBase}_${pickedSortDir}`;
  }

  function updatePickedSortButtons() {
    if (!pickedSortBar) return;
    pickedSortBar.querySelectorAll("[data-sort-base]").forEach((btn) => {
      const base = btn.getAttribute("data-sort-base") || "";
      const label = btn.getAttribute("data-label") || btn.textContent.replace(/\s[↑↓]$/, "");
      btn.setAttribute("data-label", label);
      const active = base === pickedSortBase;
      btn.classList.toggle("active", active);
      if (active && base !== "picked") {
        btn.textContent = `${label} ${pickedSortDir === "asc" ? "↑" : "↓"}`;
      } else {
        btn.textContent = label;
      }
    });
  }

  function renderSmartPickedCard(movie) {
    if (!movie) return "";
    const name = movie.name || "";
    const title = movie.title || "";
    const jacket = movie.jacket_img || "";
    const prefix = movie.prefix || "";
    const releasingDate = movie.releasing_date || "";
    const filmBirthDate = movie.film_birth_date || "";
    const director = movie.director || "";
    const score = Number.isFinite(movie.score) ? movie.score : null;
    const viewersWatched = movie.viewers_number_watched !== undefined && movie.viewers_number_watched !== null ? movie.viewers_number_watched : "";
    const scTimes = Number.isFinite(movie.sc_times) ? movie.sc_times : 0;
    const comeTimes = Number.isFinite(movie.come_times) ? movie.come_times : 0;
    const highestRank = Number.isFinite(movie.highest_rank) ? movie.highest_rank : 0;
    const owned = Number.isFinite(movie.owned) ? movie.owned : 0;
    const busUrl = movie.bus_url || "";
    const searchUrl = movie.search_url || "";
    const javUrl = movie.jav_url || "";
    const videoUrl = movie.video_url || "";
    const genres = Array.isArray(movie.genre) ? movie.genre : [];
    const casts = Array.isArray(movie.cast) ? movie.cast : [];
    const badgeTop = `${comeTimes > 0 ? `<span class="badge-come">❤ ${comeTimes}</span>` : ""}${scTimes > 0 ? ` <span class="badge-times">× ${scTimes}</span>` : ""}`;
    const badgeBottom = highestRank > 0 || score !== null
      ? `<div class="badge-bottom-right">${highestRank > 0 ? `<span class="badge-rank">#${highestRank}</span>` : ""}${score !== null ? `<span class="badge-score">★ ${score}</span>` : ""}</div>`
      : "";
    const viewerInfo = score !== null && viewersWatched !== "" ? `<small class="text-muted ml-2">(${escapeHtml(viewersWatched)})</small>` : "";
    let ownedLine = "";
    if (owned >= 2) {
      let btn = "";
      if (owned === 5) {
        btn = `<a href="${attr(videoUrl)}" class="btn btn-sm btn-info me-2">Play</a>`;
      } else if (owned === 4) {
        btn = `<a href="${attr(videoUrl)}" class="btn btn-sm btn-success me-2">Play</a>`;
      } else if (owned === 6) {
        btn = `<a href="${attr(videoUrl)}" class="btn btn-sm btn-secondary me-2">ReMoved</a>`;
      }
      const filmDate = filmBirthDate ? `<span class="text-muted small">| ${escapeHtml(filmBirthDate)}</span>` : "";
      ownedLine = `<li class="list-group-item d-flex align-items-center">${btn}${filmDate}</li>`;
    }
    const genreHtml = genres.map((genre) =>
      `<a href="/cards?gn=${encodeURIComponent(genre)}" class="badge badge-genre me-1 mb-1">${escapeHtml(genre)}</a>`
    ).join("");
    const castHtml = casts.map((cast) => {
      const cname = cast?.name || "";
      const cnameShow = cast?.name_show || cname;
      return `<a href="/cards?cn=${encodeURIComponent(cname)}" class="person-tag mr-2" title="${attr(cnameShow)}">${escapeHtml(cnameShow)}</a>`;
    }).join("");
    const directorHtml = director
      ? `<a href="/cards?dn=${encodeURIComponent(director)}" class="person-tag ml-2" title="${attr(director)}">${escapeHtml(director)}</a>`
      : "";
    const links = `${busUrl ? `<a href="${attr(busUrl)}" class="me-2" target="_blank" rel="noopener">Bus</a>` : ""}${searchUrl ? `<a href="${attr(searchUrl)}" class="me-2" target="_blank" rel="noopener">Suk</a>` : ""}${javUrl ? `<a href="${attr(javUrl)}" class="me-2" target="_blank" rel="noopener">Lib</a>` : ""}`;

    return `<div class="col">
      <article class="card shadow-sm h-100 card-lift">
        <a href="/movie/${encodeURIComponent(name)}" class="d-block" aria-label="查看 ${attr(name)} 详情">
          <div class="thumb-2x3">
            <img src="${attr(jacket)}" alt="${attr(name)} 海报" loading="lazy" decoding="async">
            <div class="badge-top-left">${badgeTop}</div>
            ${badgeBottom}
          </div>
        </a>
        <ul class="list-group list-group-flush">
          <li class="list-group-item">
            <h6 class="mb-1 text-truncate" title="${attr(title)}">
              <a href="/cards?pn=${encodeURIComponent(prefix)}" class="stretched-link text-reset text-decoration-none">${escapeHtml(name)}</a>
              ${viewerInfo}
            </h6>
            ${title ? `<p class="mb-0 text-muted line-clamp-2" title="${attr(title)}">${escapeHtml(title)}</p>` : ""}
          </li>
          ${ownedLine}
          <li class="list-group-item text-muted">${genreHtml}</li>
          <li class="list-group-item">${castHtml}${directorHtml}</li>
          <li class="list-group-item d-flex align-items-center small">
            ${links}
            <span class="ms-auto text-muted">${escapeHtml(releasingDate)}</span>
          </li>
        </ul>
      </article>
    </div>`;
  }

  function drawPickedMovies(list, totalSizeGb, shouldScroll) {
    if (!pickedCard || !pickedGrid || !pickedCount) return;
    pickedGrid.innerHTML = list.map(renderSmartPickedCard).join("");
    pickedCount.textContent = `共 ${list.length} 部${totalSizeGb ? ` · ${totalSizeGb} GB` : ""}`;
    pickedCard.style.display = list.length > 0 ? "block" : "none";
    if (shouldScroll && list.length > 0) {
      pickedCard.scrollIntoView({behavior: "smooth", block: "start"});
    }
  }

  function applyPickedSort(shouldScroll) {
    drawPickedMovies(sortPickedMovies(pickedMoviesCache, currentSortKey()), pickedTotalSizeGb, shouldScroll);
    updatePickedSortButtons();
  }

  function renderMovies(movies, totalSizeGb) {
    const list = Array.isArray(movies) ? movies : [];
    pickedMoviesCache = list.map((movie, idx) => Object.assign({_pickedOrder: idx}, movie));
    pickedTotalSizeGb = totalSizeGb || "";
    applyPickedSort(true);
  }

  function infoNumber(info, key) {
    const value = info && Number(info[key]);
    return Number.isFinite(value) ? value : 0;
  }

  function clearPickInfo() {
    if (pickInfoCard) pickInfoCard.style.display = "none";
    if (pickInfoSummary) pickInfoSummary.textContent = "";
    if (pickInfoGrid) pickInfoGrid.innerHTML = "";
  }

  function renderPickInfo(info) {
    if (!pickInfoCard || !pickInfoSummary || !pickInfoGrid || !info) {
      clearPickInfo();
      return;
    }
    const totalSizeGb = info.total_size_gb || "";
    const rows = [
      ["原始候选电影", infoNumber(info, "raw_candidate_movie_count")],
      ["参与计算 Cast", infoNumber(info, "cast_count")],
      ["阻断后剩余 Cast", infoNumber(info, "after_block_cast_count")],
      ["阻断后剩余 Movie", infoNumber(info, "after_block_movie_count")],
      ["总大小(GB)", totalSizeGb || "-"],
      ["Cast LastSc 阻断 Cast", infoNumber(info, "cast_last_sc_blocked_cast_count")],
      ["LastSc Event 阻断 Cast", infoNumber(info, "last_sc_event_blocked_cast_count")],
      ["Cast LastSc 影响 Movie", infoNumber(info, "cast_last_sc_blocked_movie_count")],
      ["LastSc Event 影响 Movie", infoNumber(info, "last_sc_event_blocked_movie_count")],
      ["演员去重跳过 Movie", infoNumber(info, "selected_actor_skipped_movie_count")]
    ];
    pickInfoSummary.textContent = `阻断后 Movie ${infoNumber(info, "after_block_movie_count")} 部`;
    pickInfoGrid.innerHTML = rows.map((row) => `<div class="col-6 col-md-3 col-xl-2">
      <div class="border rounded p-2 h-100">
        <div class="text-muted small">${escapeHtml(row[0])}</div>
        <div class="fw-semibold">${escapeHtml(row[1])}</div>
      </div>
    </div>`).join("");
    pickInfoCard.style.display = "block";
  }

  function setCopyVisible(visible) {
    if (copyCard) copyCard.style.display = visible ? "block" : "none";
  }

  function resetCopyStatusUI() {
    setCopyVisible(false);
    if (copyProgress) {
      copyProgress.style.width = "0%";
      copyProgress.textContent = "";
    }
    if (copyMeta) copyMeta.textContent = "";
    if (copyCurrent) copyCurrent.textContent = "";
    if (copyError) {
      copyError.style.display = "none";
      copyError.textContent = "";
    }
    if (btnStopCopy) btnStopCopy.disabled = true;
  }

  function updateCopyStatus(status) {
    if (!copyCard || !copyMeta || !copyProgress || !copyCurrent || !copyError || !btnStopCopy) return;
    const st = status || {};
    const total = Number.isFinite(st.total) ? st.total : 0;
    const done = Number.isFinite(st.done) ? st.done : 0;
    const running = Boolean(st.running);
    const stopped = Boolean(st.stopped);
    const finishedAt = st.finished_at || 0;
    const current = st.current || "";
    const lastErr = st.last_error || "";

    if (!running && !stopped && !finishedAt && total === 0) {
      setCopyVisible(false);
      return;
    }

    setCopyVisible(true);
    const percent = total > 0 ? Math.min(100, Math.round((done / total) * 100)) : 0;
    copyProgress.style.width = `${percent}%`;
    copyProgress.textContent = percent > 0 ? `${percent}%` : "";

    let stateText = "等待中";
    if (running) stateText = "复制中";
    else if (stopped) stateText = "已停止";
    else if (finishedAt) stateText = "已完成";
    copyMeta.textContent = `${stateText} · ${done}/${total}`;
    copyCurrent.textContent = current ? `当前：${current}` : "";
    if (lastErr) {
      copyError.style.display = "block";
      copyError.textContent = lastErr;
    } else {
      copyError.style.display = "none";
      copyError.textContent = "";
    }
    btnStopCopy.disabled = !running;
  }

  function stopCopyPolling() {
    if (!copyTimer) return;
    window.clearInterval(copyTimer);
    copyTimer = null;
  }

  function fetchCopyStatus() {
    fetch("/api/triggers/sc/copy-status")
      .then((resp) => resp.json())
      .then((data) => {
        const st = data || {};
        if (!copyStatusBootstrapped) {
          copyStatusBootstrapped = true;
          if (!copyStartedInPage && !st.running) {
            resetCopyStatusUI();
            stopCopyPolling();
            return;
          }
        }
        updateCopyStatus(st);
        if (st.running) startCopyPolling();
        else stopCopyPolling();
      })
      .catch(() => {});
  }

  function startCopyPolling() {
    if (copyTimer) return;
    copyTimer = window.setInterval(fetchCopyStatus, 2000);
  }

  function buildReqFromGroup(group) {
    const req = {
      Page: 1,
      PageSize: 100000,
      MediaOwned: 3
    };
    group.querySelectorAll("[data-req-field]").forEach((input) => {
      const key = input.getAttribute("data-req-field");
      const type = input.getAttribute("data-type") || "string";
      const raw = String(input.value || "").trim();
      if (!key) return;
      if (type === "int") {
        const value = raw === "" ? 0 : parseNumber(raw, false);
        req[key] = value === null ? 0 : value;
        return;
      }
      if (type === "int-opt") {
        if (raw === "") return;
        const value = parseNumber(raw, false);
        if (value !== null) req[key] = value;
        return;
      }
      if (type === "float") {
        const value = raw === "" ? 0 : parseNumber(raw, true);
        req[key] = value === null ? 0 : value;
        return;
      }
      if (type === "float-opt") {
        if (raw === "") return;
        const value = parseNumber(raw, true);
        if (value !== null) req[key] = value;
        return;
      }
      if (raw !== "") req[key] = raw;
    });
    return req;
  }

  function buildOptions(n) {
    const castLastScBlockDays = parseNumber(document.getElementById("castLastScBlockDays").value.trim(), false);
    const lastScEventBlockDays = parseNumber(document.getElementById("lastScEventBlockDays").value.trim(), false);
    const rawRank20Min = document.getElementById("rank20Min").value.trim();
    const rawRank100Min = document.getElementById("rank100Min").value.trim();
    const rawRank500Min = document.getElementById("rank500Min").value.trim();
    const castLastScPenaltyAlpha = parseNumber(document.getElementById("castLastScPenaltyAlpha").value.trim(), true);
    const lastScEventPenaltyAlpha = parseNumber(document.getElementById("lastScEventPenaltyAlpha").value.trim(), true);
    const castOwnedScRatioPenaltyAlpha = parseNumber(document.getElementById("castOwnedScRatioPenaltyAlpha").value.trim(), true);
    const movieHasScPenaltyAlpha = parseNumber(document.getElementById("movieHasScPenaltyAlpha").value.trim(), true);
    const rank20Min = rawRank20Min === "" ? 0 : parseNumber(rawRank20Min, false);
    let rank100Min = rawRank100Min === "" ? null : parseNumber(rawRank100Min, false);
    let rank500Min = rawRank500Min === "" ? null : parseNumber(rawRank500Min, false);

    if (rank20Min !== null) {
      if (rank100Min === null) rank100Min = rank20Min;
      if (rank500Min === null) rank500Min = rank100Min;
    }
    if (castLastScBlockDays === null || lastScEventBlockDays === null || rank20Min === null || rank100Min === null || rank500Min === null ||
      castLastScPenaltyAlpha === null || lastScEventPenaltyAlpha === null || castOwnedScRatioPenaltyAlpha === null || movieHasScPenaltyAlpha === null) {
      showMessage("全局抽取策略参数不完整", "danger");
      return null;
    }
    if (rank20Min > rank100Min || rank100Min > rank500Min || rank500Min > n) {
      showMessage("需要满足 0 <= rank20 <= rank100 <= rank500 <= 抽取数量", "danger");
      return null;
    }
    return {
      castLastScBlockDays,
      lastScEventBlockDays,
      rank20Min,
      rank100Min,
      rank500Min,
      castLastScPenaltyAlpha,
      lastScEventPenaltyAlpha,
      castOwnedScRatioPenaltyAlpha,
      movieHasScPenaltyAlpha
    };
  }

  function buildPayload() {
    if (!form.checkValidity()) {
      form.classList.add("was-validated");
      return null;
    }
    const n = Number.parseInt(document.getElementById("pickCount").value.trim(), 10);
    if (!Number.isFinite(n) || n <= 0) {
      showMessage("请输入大于 0 的数量", "danger");
      return null;
    }
    const options = buildOptions(n);
    if (!options) return null;
    const groups = groupWrap.querySelectorAll(".pick-group");
    if (groups.length === 0) {
      showMessage("请添加至少一组条件", "danger");
      return null;
    }
    const reqs = [];
    for (const group of groups) {
      const weightEl = group.querySelector('[data-role="weight"]');
      const weight = weightEl ? Number.parseInt(weightEl.value.trim(), 10) : 0;
      if (!Number.isFinite(weight) || weight <= 0) {
        showMessage("权重需为正数", "danger");
        return null;
      }
      reqs.push({weight, req: buildReqFromGroup(group)});
    }
    return {pickN: n, source: pageSource, options, reqs};
  }

  function runPick(url, actionLabel) {
    const payload = buildPayload();
    if (!payload) return;
    clearPickInfo();
    post(url, payload)
      .then(async (resp) => {
        const data = await resp.json().catch(() => ({}));
        if (!resp.ok) {
          showMessage(data.error || `触发失败(${resp.status})`, "danger");
          return;
        }
        const picked = data.picked || 0;
        renderPickInfo(data.pick_info || null);
        renderMovies(data.movies || [], data.total_size_gb || "");
        if (data.copy_status) {
          copyStatusBootstrapped = true;
          copyStartedInPage = Boolean(data.copy_started || data.copy_status.running);
          updateCopyStatus(data.copy_status);
          if (data.copy_status.running) startCopyPolling();
        }
        if (data.copy_started === false && data.copy_status?.running) {
          showMessage("Copy 任务正在执行，本次不再启动", "danger");
          return;
        }
        showMessage(`${actionLabel || "执行完成"}，已抽取 ${picked} 部`);
      })
      .catch((err) => showMessage(`异常：${err}`, "danger"));
  }

  btnAdd.addEventListener("click", addGroup);
  addGroup();

  form.addEventListener("submit", (event) => {
    event.preventDefault();
    runPick("/api/triggers/sc/pick-smart-copy", "Smart Pick + Copy 完成");
  });

  btnPickOnly?.addEventListener("click", () => {
    runPick("/api/triggers/sc/pick-smart-only", "Smart Pick 完成");
  });

  btnStopCopy?.addEventListener("click", () => {
    post("/api/triggers/sc/copy-stop", {})
      .then((resp) => resp.json().catch(() => ({})))
      .then((data) => {
        copyStatusBootstrapped = true;
        updateCopyStatus(data.status || {});
        showMessage("已发送停止指令");
        stopCopyPolling();
      })
      .catch((err) => showMessage(`停止失败：${err}`, "danger"));
  });

  pickedSortBar?.addEventListener("click", (event) => {
    const btn = event.target.closest("[data-sort-base]");
    if (!btn) return;
    const base = btn.getAttribute("data-sort-base") || "picked";
    const defaultDir = btn.getAttribute("data-default-dir") || "desc";
    if (base === pickedSortBase) {
      if (base !== "picked") {
        pickedSortDir = pickedSortDir === "asc" ? "desc" : "asc";
      }
    } else {
      pickedSortBase = base;
      pickedSortDir = defaultDir;
    }
    applyPickedSort(false);
  });

  fetchCopyStatus();
}

function bindMediaIngest() {
  const page = document.getElementById("mediaIngestPage");
  if (!page) return;

  const API = {
    plan: "/api/triggers/media/plan",
    precheck: "/api/triggers/media/precheck",
    commit: "/api/triggers/media/commit",
    back: "/api/triggers/media/return"
  };
  const btnPrecheck = document.getElementById("btnPrecheck");
  const btnCommit = document.getElementById("btnCommit");
  const btnReturn = document.getElementById("btnReturn");
  const sumTotal = document.getElementById("sumTotal");
  const sumPassLabel = document.getElementById("sumPassLabel");
  const sumPass = document.getElementById("sumPass");
  const sumFailLabel = document.getElementById("sumFailLabel");
  const sumFail = document.getElementById("sumFail");
  const statusText = document.getElementById("statusText");
  const generatedAtText = document.getElementById("generatedAtText");
  let loading = false;

  function showMsg(text, ok) {
    showMessage(text, ok ? "success" : "danger");
  }

  function renderMovieLink(movieName) {
    const name = String(movieName || "").trim();
    if (!name) return "-";
    return `<a href="/movie/${encodeURIComponent(name)}">${escapeHtml(name)}</a>`;
  }

  function isCommitPhase(data) {
    return String((data && data.phase) || "").trim() === "commit";
  }

  function focusCommitResults(data) {
    const safe = data || {};
    const targetId = Array.isArray(safe.commit_fail) && safe.commit_fail.length
      ? "commitFailSection"
      : (Array.isArray(safe.commit_success) && safe.commit_success.length ? "commitSuccessSection" : "commitPreviewSection");
    const target = document.getElementById(targetId);
    if (target && typeof target.scrollIntoView === "function") {
      target.scrollIntoView({behavior: "smooth", block: "start"});
    }
  }

  function setLoading(next) {
    loading = Boolean(next);
    if (btnPrecheck) btnPrecheck.disabled = loading;
    if (btnCommit) {
      const canCommit = btnCommit.getAttribute("data-can-commit") === "1";
      btnCommit.disabled = loading || !canCommit;
    }
    if (btnReturn) btnReturn.disabled = loading;
  }

  function request(url, method) {
    return fetch(url, {
      method: method || "GET",
      headers: {"Content-Type": "application/json"}
    }).then(async (resp) => {
      const payload = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        const err = new Error((payload && payload.error) || `请求失败(${resp.status})`);
        err.result = payload && payload.result ? payload.result : null;
        throw err;
      }
      return payload || {};
    });
  }

  function renderRows(tableId, rows, columns) {
    const table = document.getElementById(tableId);
    const tbody = table?.querySelector("tbody");
    if (!tbody) return;
    const safeRows = Array.isArray(rows) ? rows : [];
    if (!safeRows.length) {
      tbody.innerHTML = `<tr><td colspan="${columns.length}" class="text-center text-muted">暂无数据</td></tr>`;
      return;
    }
    tbody.innerHTML = safeRows.map((row) => {
      const cells = columns.map((column) => `<td class="${column.cls || ""}">${column.render(row)}</td>`).join("");
      return `<tr>${cells}</tr>`;
    }).join("");
  }

  function renderState(data) {
    const safe = data || {};
    const precheck = safe.precheck || {total: 0, passed: 0, failed: 0};
    const commitPhase = isCommitPhase(safe);
    const total = commitPhase ? Number(safe.total || 0) : Number(precheck.total || 0);
    const success = commitPhase ? Number(safe.success || 0) : Number(precheck.passed || 0);
    const failed = commitPhase ? Number(safe.failed || 0) : Number(precheck.failed || 0);

    if (sumPassLabel) sumPassLabel.textContent = commitPhase ? "成功" : "通过";
    if (sumFailLabel) sumFailLabel.textContent = "失败";
    if (sumTotal) sumTotal.textContent = String(total);
    if (sumPass) sumPass.textContent = String(success);
    if (sumFail) sumFail.textContent = String(failed);
    if (statusText) statusText.textContent = safe.message || "等待操作";
    if (generatedAtText) {
      generatedAtText.textContent = commitPhase ? "" : (safe.generated_at ? `预处理时间：${formatUnixFull(safe.generated_at)}` : "");
    }

    if (btnCommit) {
      const canCommit = Boolean(safe.can_commit);
      btnCommit.setAttribute("data-can-commit", canCommit ? "1" : "0");
      btnCommit.disabled = loading || !canCommit;
      btnCommit.textContent = safe.partial_failed ? "第二段：仅插入通过项" : "第二段：执行插入";
    }

    if (btnReturn) {
      btnReturn.classList.toggle("d-none", !(safe.partial_failed && safe.has_plan));
    }

    renderRows("tblPrecheckPass", safe.precheck_pass, [
      {render: (row) => escapeHtml(row.movie_name || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.source_path || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.target_file_name || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.target_dir || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.alias || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.source_torrent_hash || "-")}
    ]);
    renderRows("tblPrecheckFail", safe.precheck_fail, [
      {render: (row) => renderMovieLink(row.movie_name)},
      {cls: "mono", render: (row) => escapeHtml(row.source_path || "-")},
      {render: (row) => `<span class="text-danger">${escapeHtml(row.error || "-")}</span>`}
    ]);
    renderRows("tblPrecheckPreview", safe.precheck_preview, [
      {render: (row) => escapeHtml(row.movie_name || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.target_dir || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.target_file_name || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.alias || "-")},
      {render: (row) => escapeHtml(humanSize(row.size))}
    ]);
    renderRows("tblCommitSuccess", safe.commit_success, [
      {render: (row) => escapeHtml(row.movie_name || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.target_path || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.alias || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.source_torrent_hash || "-")}
    ]);
    renderRows("tblCommitFail", safe.commit_fail, [
      {render: (row) => escapeHtml(row.movie_name || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.source_path || "-")},
      {render: (row) => `<span class="text-danger">${escapeHtml(row.error || "-")}</span>`}
    ]);
    renderRows("tblCommitPreview", safe.commit_preview, [
      {render: (row) => escapeHtml(row.movie_name || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.target_path || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.alias || "-")},
      {render: (row) => escapeHtml(humanSize(row.size))}
    ]);
  }

  function loadPlan() {
    setLoading(true);
    request(API.plan, "GET")
      .then(renderState)
      .catch((err) => {
        showMsg(err.message || "读取预处理状态失败", false);
        if (err.result) renderState(err.result);
      })
      .finally(() => setLoading(false));
  }

  btnPrecheck?.addEventListener("click", () => {
    setLoading(true);
    request(API.precheck, "POST")
      .then((data) => {
        renderState(data);
        showMsg(data.message || "预处理完成", true);
      })
      .catch((err) => {
        showMsg(err.message || "预处理失败", false);
        if (err.result) renderState(err.result);
      })
      .finally(() => setLoading(false));
  });

  btnCommit?.addEventListener("click", () => {
    if (btnCommit.disabled) return;
    setLoading(true);
    if (statusText) statusText.textContent = "正在执行第二段插入...";
    request(API.commit, "POST")
      .then((data) => {
        renderState(data);
        focusCommitResults(data);
        showMsg(data.message || "第二段执行完成", true);
      })
      .catch((err) => {
        showMsg(err.message || "第二段执行失败", false);
        if (err.result) {
          renderState(err.result);
          focusCommitResults(err.result);
        }
      })
      .finally(() => setLoading(false));
  });

  btnReturn?.addEventListener("click", () => {
    setLoading(true);
    request(API.back, "POST")
      .then((data) => {
        renderState(data);
        showMsg(data.message || "已返回", true);
      })
      .catch((err) => showMsg(err.message || "返回失败", false))
      .finally(() => setLoading(false));
  });

  loadPlan();
}

function bindMediaRollback() {
  const page = document.getElementById("mediaRollbackPage");
  if (!page) return;

  const API = {
    rollback: "/api/triggers/media/rollback"
  };
  const btnRollback = document.getElementById("btnRollback");
  const sumTotal = document.getElementById("sumTotal");
  const sumSuccess = document.getElementById("sumSuccess");
  const sumFail = document.getElementById("sumFail");
  const statusText = document.getElementById("statusText");
  let loading = false;

  function showMsg(text, ok) {
    showMessage(text, ok ? "success" : "danger");
  }

  function renderMovieLink(movieName) {
    const name = String(movieName || "").trim();
    if (!name) return "-";
    return `<a href="/movie/${encodeURIComponent(name)}">${escapeHtml(name)}</a>`;
  }

  function setLoading(next) {
    loading = Boolean(next);
    if (btnRollback) btnRollback.disabled = loading;
  }

  function requestRollback() {
    return fetch(API.rollback, {
      method: "POST",
      headers: {"Content-Type": "application/json"}
    }).then(async (resp) => {
      const payload = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        const err = new Error((payload && payload.error) || `请求失败(${resp.status})`);
        err.result = payload && payload.result ? payload.result : null;
        throw err;
      }
      return payload || {};
    });
  }

  function renderRows(tableId, rows, columns) {
    const table = document.getElementById(tableId);
    const tbody = table?.querySelector("tbody");
    if (!tbody) return;
    const safeRows = Array.isArray(rows) ? rows : [];
    if (!safeRows.length) {
      tbody.innerHTML = `<tr><td colspan="${columns.length}" class="text-center text-muted">暂无数据</td></tr>`;
      return;
    }
    tbody.innerHTML = safeRows.map((row) => {
      const cells = columns.map((column) => `<td class="${column.cls || ""}">${column.render(row)}</td>`).join("");
      return `<tr>${cells}</tr>`;
    }).join("");
  }

  function renderState(data) {
    const safe = data || {};
    if (sumTotal) sumTotal.textContent = String(safe.total || 0);
    if (sumSuccess) sumSuccess.textContent = String(safe.success || 0);
    if (sumFail) sumFail.textContent = String(safe.failed || 0);
    if (statusText) statusText.textContent = safe.message || "等待操作";

    renderRows("tblRollbackSuccess", safe.rollback_success, [
      {render: (row) => renderMovieLink(row.movie_name)},
      {cls: "mono", render: (row) => escapeHtml(row.source_path || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.target_path || "-")}
    ]);
    renderRows("tblRollbackFail", safe.rollback_fail, [
      {render: (row) => renderMovieLink(row.movie_name)},
      {cls: "mono", render: (row) => escapeHtml(row.source_path || "-")},
      {render: (row) => `<span class="text-danger">${escapeHtml(row.error || "-")}</span>`},
      {cls: "mono", render: (row) => escapeHtml(row.failed_path || "-")}
    ]);
    renderRows("tblRollbackPreview", safe.rollback_preview, [
      {render: (row) => renderMovieLink(row.movie_name)},
      {cls: "mono", render: (row) => escapeHtml(row.target_path || "-")}
    ]);
  }

  btnRollback?.addEventListener("click", () => {
    setLoading(true);
    requestRollback()
      .then((data) => {
        renderState(data);
        showMsg(data.message || "回滚完成", true);
      })
      .catch((err) => {
        showMsg(err.message || "回滚失败", false);
        if (err.result) renderState(err.result);
      })
      .finally(() => setLoading(false));
  });
}

function bindMediaRescan() {
  const page = document.getElementById("mediaRescanPage");
  if (!page) return;

  const API = {
    rescan: "/api/triggers/media/rescan"
  };
  const btnRescan = document.getElementById("btnRescan");
  const btnSelectAll = document.getElementById("btnSelectAll");
  const btnClearAll = document.getElementById("btnClearAll");
  const statusText = document.getElementById("statusText");
  const rootCheckboxes = Array.from(document.querySelectorAll(".root-checkbox"));
  const branchCheckboxes = Array.from(document.querySelectorAll(".branch-checkbox"));
  const rootScopeItems = Array.from(document.querySelectorAll(".root-scope-item"));
  const sumTotalFiles = document.getElementById("sumTotalFiles");
  const sumMatched = document.getElementById("sumMatched");
  const sumMoved = document.getElementById("sumMoved");
  const sumRestored = document.getElementById("sumRestored");
  const sumRemoved = document.getElementById("sumRemoved");
  const sumUnmatched = document.getElementById("sumUnmatched");
  const sumSkippedRoots = document.getElementById("sumSkippedRoots");
  const sumErrors = document.getElementById("sumErrors");
  const selectedRootList = document.getElementById("selectedRootList");
  const scannedRootList = document.getElementById("scannedRootList");
  const skippedRootList = document.getElementById("skippedRootList");
  const selectedTargetList = document.getElementById("selectedTargetList");
  const scannedTargetList = document.getElementById("scannedTargetList");
  const skippedTargetList = document.getElementById("skippedTargetList");
  let loading = false;

  function showMsg(text, ok) {
    showMessage(text, ok ? "success" : "danger");
  }

  function isScopeChecked(root, scope) {
    return rootCheckboxes.some((checkbox) => checkbox.dataset.root === root && checkbox.dataset.scope === scope && checkbox.checked);
  }

  function syncBranchStates() {
    rootScopeItems.forEach((item) => {
      const children = item.querySelectorAll(".branch-checkbox");
      children.forEach((checkbox) => {
        const enabled = isScopeChecked(checkbox.dataset.root, checkbox.dataset.scope);
        checkbox.disabled = loading || !enabled;
        const label = checkbox.closest(".branch-checkbox-item");
        if (label) label.classList.toggle("is-disabled", !enabled);
      });
    });
  }

  function setLoading(next) {
    loading = Boolean(next);
    if (btnRescan) btnRescan.disabled = loading;
    if (btnSelectAll) btnSelectAll.disabled = loading;
    if (btnClearAll) btnClearAll.disabled = loading;
    rootCheckboxes.forEach((checkbox) => {
      checkbox.disabled = loading;
    });
    branchCheckboxes.forEach((checkbox) => {
      checkbox.disabled = loading || !isScopeChecked(checkbox.dataset.root, checkbox.dataset.scope);
    });
  }

  function buildSelections() {
    return rootCheckboxes.filter((checkbox) => checkbox.checked).map((checkbox) => {
      const root = String(checkbox.dataset.root || checkbox.value || "").trim();
      const scope = String(checkbox.dataset.scope || "").trim();
      const branches = branchCheckboxes.filter((branchCheckbox) => {
        return branchCheckbox.dataset.root === root && branchCheckbox.dataset.scope === scope && branchCheckbox.checked;
      }).map((branchCheckbox) => String(branchCheckbox.dataset.branch || "").trim()).filter((value) => value !== "");
      return {root, scope, branches};
    }).filter((item) => item.root !== "" && item.scope !== "");
  }

  function requestRescan(selections) {
    return fetch(API.rescan, {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({selections})
    }).then(async (resp) => {
      const payload = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        const err = new Error((payload && payload.error) || `请求失败(${resp.status})`);
        err.result = payload && payload.result ? payload.result : null;
        throw err;
      }
      return payload || {};
    });
  }

  function renderList(target, items) {
    if (!target) return;
    const safeItems = Array.isArray(items) ? items : [];
    if (!safeItems.length) {
      target.innerHTML = '<li class="text-muted">暂无数据</li>';
      return;
    }
    target.innerHTML = safeItems.map((item) => `<li><code>${escapeHtml(item)}</code></li>`).join("");
  }

  function renderRows(tableId, rows, columns) {
    const table = document.getElementById(tableId);
    const tbody = table?.querySelector("tbody");
    if (!tbody) return;
    const safeRows = Array.isArray(rows) ? rows : [];
    if (!safeRows.length) {
      tbody.innerHTML = `<tr><td colspan="${columns.length}" class="text-center text-muted">暂无数据</td></tr>`;
      return;
    }
    tbody.innerHTML = safeRows.map((row) => {
      return `<tr>${columns.map((column) => `<td class="${column.cls || ""}">${column.render(row)}</td>`).join("")}</tr>`;
    }).join("");
  }

  function buildStatusText(data) {
    const safe = data || {};
    const scannedTargets = Array.isArray(safe.scanned_targets) ? safe.scanned_targets.length : 0;
    const skippedTargets = Array.isArray(safe.skipped_targets) ? safe.skipped_targets.length : 0;
    if (!safe.selected_roots || !safe.selected_roots.length) {
      return "等待操作";
    }
    if (scannedTargets === 0 && skippedTargets > 0) {
      return "所选范围当前都不存在，已全部跳过";
    }
    return `重扫完成：扫描 ${scannedTargets} 个范围，跳过 ${skippedTargets} 个范围`;
  }

  function renderState(data) {
    const safe = data || {};
    if (sumTotalFiles) sumTotalFiles.textContent = String(safe.total_files || 0);
    if (sumMatched) sumMatched.textContent = String(safe.matched || 0);
    if (sumMoved) sumMoved.textContent = String(safe.moved || 0);
    if (sumRestored) sumRestored.textContent = String(safe.restored || 0);
    if (sumRemoved) sumRemoved.textContent = String(safe.marked_removed || 0);
    if (sumUnmatched) sumUnmatched.textContent = String(safe.unmatched || 0);
    if (sumSkippedRoots) sumSkippedRoots.textContent = String((safe.skipped_targets || []).length);
    if (sumErrors) sumErrors.textContent = String(safe.errors || 0);
    if (statusText) statusText.textContent = buildStatusText(safe);

    renderList(selectedRootList, safe.selected_roots);
    renderList(scannedRootList, safe.scanned_roots);
    renderList(skippedRootList, safe.skipped_roots);
    renderList(selectedTargetList, safe.selected_targets);
    renderList(scannedTargetList, safe.scanned_targets);
    renderList(skippedTargetList, safe.skipped_targets);

    renderRows("tblMoved", safe.moved_items, [
      {render: (row) => escapeHtml(row.movie_name || row.movie_jav_id || row.file_name || "-")},
      {render: (row) => escapeHtml(row.match_by || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.previous_path || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.path || "-")}
    ]);
    renderRows("tblRestored", safe.restored_items, [
      {render: (row) => escapeHtml(row.movie_name || row.movie_jav_id || row.file_name || "-")},
      {render: (row) => escapeHtml(row.match_by || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.previous_path || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.path || "-")}
    ]);
    renderRows("tblRemoved", safe.removed_items, [
      {render: (row) => escapeHtml(row.movie_name || row.movie_jav_id || row.file_name || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.root_dir || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.path || "-")}
    ]);
    renderRows("tblUnmatched", safe.unmatched_items, [
      {render: (row) => escapeHtml(row.file_name || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.root_dir || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.path || "-")}
    ]);
    renderRows("tblErrors", safe.error_items, [
      {render: (row) => escapeHtml(row.movie_name || row.file_name || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.path || "-")},
      {render: (row) => `<span class="text-danger">${escapeHtml(row.error || "-")}</span>`}
    ]);
  }

  btnSelectAll?.addEventListener("click", () => {
    rootCheckboxes.forEach((checkbox) => {
      checkbox.checked = true;
    });
    branchCheckboxes.forEach((checkbox) => {
      checkbox.checked = false;
    });
    syncBranchStates();
  });

  btnClearAll?.addEventListener("click", () => {
    rootCheckboxes.forEach((checkbox) => {
      checkbox.checked = false;
    });
    branchCheckboxes.forEach((checkbox) => {
      checkbox.checked = false;
    });
    syncBranchStates();
  });

  rootCheckboxes.forEach((checkbox) => {
    checkbox.addEventListener("change", syncBranchStates);
  });

  btnRescan?.addEventListener("click", () => {
    const selections = buildSelections();
    if (!selections.length) {
      showMsg("至少选择一个扫描目录", false);
      return;
    }
    setLoading(true);
    if (statusText) statusText.textContent = "正在重扫...";
    requestRescan(selections)
      .then((data) => {
        renderState(data);
        showMsg("媒体重扫完成", true);
      })
      .catch((err) => {
        showMsg(err.message || "媒体重扫失败", false);
        if (err.result) renderState(err.result);
      })
      .finally(() => {
        setLoading(false);
        syncBranchStates();
      });
  });

  syncBranchStates();
}

function bindScMediaMove() {
  const page = document.getElementById("scMediaMovePage");
  if (!page) return;

  const API = {
    plan: "/api/triggers/sc-media/plan",
    precheck: "/api/triggers/sc-media/precheck",
    commit: "/api/triggers/sc-media/commit",
    back: "/api/triggers/sc-media/return"
  };
  const scopeModeInputs = Array.from(document.querySelectorAll('input[name="scopeMode"]'));
  const scNameGroup = document.getElementById("scNameGroup");
  const allModeHint = document.getElementById("allModeHint");
  const scNameInput = document.getElementById("scNameInput");
  const btnPrecheck = document.getElementById("btnPrecheck");
  const btnCommit = document.getElementById("btnCommit");
  const btnReturn = document.getElementById("btnReturn");
  const sumTotal = document.getElementById("sumTotal");
  const sumPass = document.getElementById("sumPass");
  const sumSkip = document.getElementById("sumSkip");
  const sumFail = document.getElementById("sumFail");
  const statusText = document.getElementById("statusText");
  const generatedAtText = document.getElementById("generatedAtText");
  let loading = false;

  function currentMode() {
    const checked = scopeModeInputs.find((input) => input?.checked);
    return checked ? String(checked.value || "").trim() : "single";
  }

  function currentScName() {
    return scNameInput ? String(scNameInput.value || "").trim() : "";
  }

  function renderMovieLink(movieName) {
    const name = String(movieName || "").trim();
    if (!name) return "-";
    return `<a href="/movie/${encodeURIComponent(name)}">${escapeHtml(name)}</a>`;
  }

  function renderScLinks(scNames) {
    const names = Array.isArray(scNames)
      ? scNames.map((name) => String(name || "").trim()).filter((name) => name !== "")
      : [];
    if (!names.length) return "-";
    return names.map((name) => `<a href="/sc-events/${encodeURIComponent(name)}">${escapeHtml(name)}</a>`).join("<br>");
  }

  function setLoading(next) {
    loading = Boolean(next);
    if (btnPrecheck) btnPrecheck.disabled = loading;
    if (btnCommit) {
      const canCommit = btnCommit.getAttribute("data-can-commit") === "1";
      btnCommit.disabled = loading || !canCommit;
    }
    if (btnReturn) btnReturn.disabled = loading;
    scopeModeInputs.forEach((input) => {
      input.disabled = loading;
    });
    if (scNameInput) scNameInput.disabled = loading;
  }

  function syncModeUi() {
    const mode = currentMode();
    const isAllMode = mode === "all";
    scNameGroup?.classList.toggle("d-none", isAllMode);
    allModeHint?.classList.toggle("d-none", !isAllMode);
  }

  function request(url, options) {
    return fetch(url, options).then(async (resp) => {
      const payload = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        const err = new Error(payload?.error || `请求失败(${resp.status})`);
        err.result = payload?.result || null;
        throw err;
      }
      return payload || {};
    });
  }

  function requestJSON(url, method, body) {
    return request(url, {
      method,
      headers: {"Content-Type": "application/json"},
      body: body ? JSON.stringify(body) : null
    });
  }

  function renderRows(tableId, rows, columns) {
    const table = document.getElementById(tableId);
    const tbody = table?.querySelector("tbody");
    if (!tbody) return;
    const safeRows = Array.isArray(rows) ? rows : [];
    if (!safeRows.length) {
      tbody.innerHTML = `<tr><td colspan="${columns.length}" class="text-center text-muted">暂无数据</td></tr>`;
      return;
    }
    tbody.innerHTML = safeRows.map((row) => {
      const cells = columns.map((column) => `<td class="${column.cls || ""}">${column.render(row)}</td>`).join("");
      return `<tr>${cells}</tr>`;
    }).join("");
  }

  function renderState(data) {
    const safe = data || {};
    if (sumTotal) sumTotal.textContent = String(safe.total || 0);
    if (sumPass) sumPass.textContent = String(safe.movable || 0);
    if (sumSkip) sumSkip.textContent = String(safe.skipped || 0);
    if (sumFail) sumFail.textContent = String(safe.failed || 0);
    if (statusText) statusText.textContent = safe.message || "等待操作";
    if (generatedAtText) {
      generatedAtText.textContent = safe.generated_at ? `预处理时间：${formatUnixFull(safe.generated_at)}` : "";
    }
    if (btnCommit) {
      const canCommit = Boolean(safe.can_commit);
      btnCommit.setAttribute("data-can-commit", canCommit ? "1" : "0");
      btnCommit.disabled = loading || !canCommit;
    }
    if (btnReturn) {
      btnReturn.classList.toggle("d-none", !safe.has_plan);
    }

    renderRows("tblPrecheckPass", safe.precheck_pass, [
      {render: (row) => renderScLinks(row.sc_names)},
      {render: (row) => renderMovieLink(row.movie_name)},
      {cls: "mono", render: (row) => escapeHtml(row.source_path || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.target_path || "-")}
    ]);
    renderRows("tblPrecheckSkip", safe.precheck_skip, [
      {render: (row) => renderScLinks(row.sc_names)},
      {render: (row) => renderMovieLink(row.movie_name)},
      {cls: "mono", render: (row) => escapeHtml(row.source_path || "-")},
      {render: (row) => `<span class="text-muted">${escapeHtml(row.error || "-")}</span>`}
    ]);
    renderRows("tblPrecheckFail", safe.precheck_fail, [
      {render: (row) => renderScLinks(row.sc_names)},
      {render: (row) => renderMovieLink(row.movie_name)},
      {cls: "mono", render: (row) => escapeHtml(row.source_path || "-")},
      {render: (row) => `<span class="text-danger">${escapeHtml(row.error || "-")}</span>`}
    ]);
    renderRows("tblCommitSuccess", safe.commit_success, [
      {render: (row) => renderScLinks(row.sc_names)},
      {render: (row) => renderMovieLink(row.movie_name)},
      {cls: "mono", render: (row) => escapeHtml(row.source_path || "-")},
      {cls: "mono", render: (row) => escapeHtml(row.target_path || "-")}
    ]);
    renderRows("tblCommitFail", safe.commit_fail, [
      {render: (row) => renderScLinks(row.sc_names)},
      {render: (row) => renderMovieLink(row.movie_name)},
      {cls: "mono", render: (row) => escapeHtml(row.source_path || "-")},
      {render: (row) => `<span class="text-danger">${escapeHtml(row.error || "-")}</span>`}
    ]);
  }

  function emptySingleState() {
    renderState({
      mode: "single",
      message: "请先输入 SC 名称",
      has_plan: false,
      can_commit: false,
      precheck_pass: [],
      precheck_skip: [],
      precheck_fail: [],
      commit_success: [],
      commit_fail: []
    });
  }

  function loadPlan() {
    const mode = currentMode();
    const scName = currentScName();
    if (mode === "single" && !scName) {
      emptySingleState();
      return;
    }
    const qs = new URLSearchParams({mode, sc_name: scName});
    setLoading(true);
    request(`${API.plan}?${qs}`, {method: "GET"})
      .then(renderState)
      .catch((err) => {
        showMessage(err.message || "读取预处理状态失败", "danger");
        if (err.result) renderState(err.result);
      })
      .finally(() => setLoading(false));
  }

  btnPrecheck?.addEventListener("click", () => {
    const mode = currentMode();
    const scName = currentScName();
    if (mode === "single" && !scName) {
      showMessage("请先输入 SC 名称", "danger");
      return;
    }
    setLoading(true);
    requestJSON(API.precheck, "POST", {mode, sc_name: scName})
      .then((data) => {
        renderState(data);
        showMessage(data.message || "预处理完成");
      })
      .catch((err) => {
        showMessage(err.message || "预处理失败", "danger");
        if (err.result) renderState(err.result);
      })
      .finally(() => setLoading(false));
  });

  btnCommit?.addEventListener("click", () => {
    if (btnCommit.disabled) return;
    const mode = currentMode();
    const scName = currentScName();
    if (mode === "single" && !scName) {
      showMessage("请先输入 SC 名称", "danger");
      return;
    }
    setLoading(true);
    requestJSON(API.commit, "POST", {mode, sc_name: scName})
      .then((data) => {
        renderState(data);
        showMessage(data.message || "第二段执行完成");
      })
      .catch((err) => {
        showMessage(err.message || "第二段执行失败", "danger");
        if (err.result) renderState(err.result);
      })
      .finally(() => setLoading(false));
  });

  btnReturn?.addEventListener("click", () => {
    const mode = currentMode();
    const scName = currentScName();
    if (mode === "single" && !scName) {
      showMessage("请先输入 SC 名称", "danger");
      return;
    }
    setLoading(true);
    requestJSON(API.back, "POST", {mode, sc_name: scName})
      .then((data) => {
        renderState(data);
        showMessage(data.message || "已返回");
      })
      .catch((err) => showMessage(err.message || "返回失败", "danger"))
      .finally(() => setLoading(false));
  });

  scopeModeInputs.forEach((input) => {
    input.addEventListener("change", () => {
      syncModeUi();
      loadPlan();
    });
  });

  syncModeUi();
  if (currentMode() === "all" || currentScName()) {
    loadPlan();
  }
}

function bindScEventDetail() {
  const pageMsgEl = document.getElementById("scEventPageMsg");
  const modalEl = document.getElementById("scEventEditModal");
  const formEl = document.getElementById("scEventEditForm");
  const submitBtn = document.getElementById("scEventEditSubmit");
  const formMsgEl = document.getElementById("scEventEditMsg");
  const comeMovieEl = document.getElementById("scEventEditComeMovie");
  const castSourcePoolEl = document.getElementById("scEventEditCastSourcePool");
  const kindEl = document.getElementById("scEventEditKind");
  const durationEl = document.getElementById("scEventEditDuration");
  const fgEl = document.getElementById("scEventEditFg");
  const vesselEl = document.getElementById("scEventEditVessel");
  const movieCastEl = document.getElementById("scEventEditMovieCast");
  const remarksEl = document.getElementById("scEventEditRemarks");

  if (!formEl || !submitBtn || !comeMovieEl || !castSourcePoolEl || !kindEl || !durationEl || !fgEl || !vesselEl || !movieCastEl || !remarksEl) {
    return;
  }

  function showInlineMsg(el, text, ok, extraClass = "") {
    if (!el) return;
    el.textContent = String(text || "");
    el.className = `alert ${extraClass} ${ok ? "alert-success" : "alert-danger"}`.trim();
  }

  function hideInlineMsg(el, extraClass = "") {
    if (!el) return;
    el.textContent = "";
    el.className = `alert d-none ${extraClass}`.trim();
  }

  function setSubmitting(submitting) {
    submitBtn.disabled = submitting;
    submitBtn.textContent = submitting ? "保存中..." : "保存";
  }

  function listCastNames(movieJavId) {
    const normalizedMovieJavId = String(movieJavId || "").trim();
    if (!normalizedMovieJavId) {
      return [];
    }
    const sourceEls = castSourcePoolEl.querySelectorAll(".js-sc-event-cast-source");
    for (const sourceEl of sourceEls) {
      if (!sourceEl || String(sourceEl.getAttribute("data-movie-jav-id") || "").trim() !== normalizedMovieJavId) {
        continue;
      }
      return Array.from(sourceEl.querySelectorAll(".js-sc-event-cast-item"))
        .map((itemEl) => String(itemEl.getAttribute("data-cast-name") || "").trim())
        .filter((name) => name !== "");
    }
    return [];
  }

  function appendMovieCastOption(value, selected) {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = value;
    option.selected = !!selected;
    movieCastEl.appendChild(option);
  }

  function rebuildMovieCastOptions(movieJavId, preferredValue) {
    const normalizedPreferredValue = String(preferredValue || "").trim();
    const castNames = listCastNames(movieJavId);
    movieCastEl.innerHTML = "";

    const emptyOption = document.createElement("option");
    emptyOption.value = "";
    emptyOption.textContent = "可为空";
    emptyOption.selected = normalizedPreferredValue === "";
    movieCastEl.appendChild(emptyOption);

    let matched = normalizedPreferredValue === "";
    castNames.forEach((name) => {
      const selected = normalizedPreferredValue !== "" && name === normalizedPreferredValue;
      if (selected) {
        matched = true;
      }
      appendMovieCastOption(name, selected);
    });

    if (!matched && normalizedPreferredValue !== "") {
      appendMovieCastOption(normalizedPreferredValue, true);
    }
  }

  function buildPayload() {
    return {
      comeMovieJavId: String(comeMovieEl.value || "").trim(),
      kind: String(kindEl.value || "").trim(),
      duration: Number(durationEl.value || 0),
      fg: String(fgEl.value || "").trim(),
      vessel: String(vesselEl.value || "").trim(),
      movieCast: String(movieCastEl.value || "").trim(),
      remarks: String(remarksEl.value || "").trim()
    };
  }

  comeMovieEl.addEventListener("change", () => {
    rebuildMovieCastOptions(comeMovieEl.value, "");
  });

  formEl.addEventListener("submit", async (event) => {
    event.preventDefault();
    hideInlineMsg(formMsgEl, "mt-3 mb-0");
    setSubmitting(true);

    try {
      const data = await fetchJSON(String(formEl.getAttribute("data-update-url") || ""), {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(buildPayload())
      });
      if (data && data.ok === false) {
        throw new Error(data.error || "保存失败");
      }
      showInlineMsg(pageMsgEl, data.message || "保存成功", true, "mb-3");
      if (window.bootstrap && modalEl) {
        window.bootstrap.Modal.getOrCreateInstance(modalEl).hide();
      }
      window.setTimeout(() => {
        window.location.reload();
      }, 250);
    } catch (err) {
      showInlineMsg(formMsgEl, err.message || "保存失败", false, "mt-3 mb-0");
      setSubmitting(false);
    }
  });

  rebuildMovieCastOptions(comeMovieEl.value, movieCastEl.getAttribute("data-selected-value") || movieCastEl.value || "");
}

function render(data) {
  if (data.ViewType === "sc-events") return renderScEventList(data);
  if (data.ViewType === "sc-event-cards") return renderScEventCards(data);
  if (data.ViewType === "sc-event-detail") return renderScEventDetail(data);
  if (data.ViewType === "sc-triggers") return renderScTriggers(data);
  if (data.ViewType === "sc-pick-smart-media") return renderScPickSmart(data);
  if (data.ViewType === "media-ingest") return renderMediaIngest(data);
  if (data.ViewType === "media-rescan") return renderMediaRescan(data);
  if (data.ViewType === "media-rollback") return renderMediaRollback(data);
  if (data.ViewType === "sc-media-move") return renderScMediaMove(data);
  return `<div class="alert alert-warning">未知页面类型：${escapeHtml(data.ViewType)}</div>`;
}

function bind(data) {
  if (data.ViewType === "sc-event-detail") bindScEventDetail();
  if (data.ViewType === "sc-triggers") bindScTriggers();
  if (data.ViewType === "sc-pick-smart-media") bindSmartPick();
  if (data.ViewType === "media-ingest") bindMediaIngest();
  if (data.ViewType === "media-rescan") bindMediaRescan();
  if (data.ViewType === "media-rollback") bindMediaRollback();
  if (data.ViewType === "sc-media-move") bindScMediaMove();
}

async function main() {
  try {
    const data = await fetchJSON(apiPathForCurrentPage() + window.location.search);
    document.title = `${data.Title || data.PageTitle || "SC"} - Rudy-GC`;
    root.innerHTML = render(data);
    bind(data);
  } catch (err) {
    root.innerHTML = `<div class="alert alert-danger">加载失败：${escapeHtml(err.message)}</div>`;
  }
}

main();
