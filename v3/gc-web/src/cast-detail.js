const sortLabels = [
  ["ByDetailUpdate", "Detail更新"],
  ["ByReleasing", "上映日期"],
  ["ByViewer", "观看人数"],
  ["ByRankDate", "上榜日期"],
  ["ByDaysInRank", "在榜天数"],
  ["ByHighestRank", "最高排名"],
  ["ByBirthTime", "下载时间"],
  ["ByMediaBirthTime", "Media下载时间"],
  ["ByLastScTime", "最后sc"],
  ["ByScTimes", "sc次数"],
  ["ByComeTimes", "co次数"]
];

const root = document.getElementById("cast-detail-root");

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

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function apiPathForCurrentPage() {
  if (window.location.pathname === "/cast") {
    return "/api/cast";
  }
  return `/api${window.location.pathname}`;
}

function linkClass(active, base) {
  return `${base || ""} ${active ? "active disabled" : ""}`.trim();
}

function actorDisplayName(actor) {
  return String(get(actor, "Chinese", "chinese") || get(actor, "Name", "name") || "");
}

function formatUnixDate(value) {
  const n = Number(value || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function formatUnixMinute(value) {
  const n = Number(value || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  const hh = String(d.getHours()).padStart(2, "0");
  const mi = String(d.getMinutes()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}`;
}

function humanDays(sec) {
  const n = Number(sec || 0);
  return n > 0 ? `${(n / 86400).toFixed(2)} 天` : "-";
}

function rankText(value) {
  const n = Number(value || 0);
  return n > 0 ? `#${escapeHtml(value)}` : "-";
}

function renderPagination(pageInfo) {
  const links = pageInfo && pageInfo.Links || [];
  const items = links.map((item) => {
    if (item.Ellipsis) {
      return `<li class="page-item disabled"><span class="page-link">...</span></li>`;
    }
    return `<li class="page-item ${item.Active ? "active" : ""} ${item.Disabled ? "disabled" : ""}">
      <a class="page-link" href="${attr(item.Href || "#")}" ${item.Disabled ? 'tabindex="-1" aria-disabled="true"' : ""}>${escapeHtml(item.Label)}</a>
    </li>`;
  }).join("");
  return `<nav class="mb-0" aria-label="Page navigation"><ul class="pagination pagination-sm mb-0">${items}</ul></nav>`;
}

function renderSortBar(sortQuery) {
  const items = sortLabels.map(([key, label]) => {
    const item = sortQuery && sortQuery[key] || {};
    return `<a href="${attr(item.Href || "/cards")}" class="${linkClass(item.Active, "btn btn-outline-primary")}">${escapeHtml(label)}</a>`;
  }).join("");
  return `<div class="btn-group btn-group-sm" role="group" aria-label="Sort Options">${items}</div>`;
}

function renderOwnedFilter(data) {
  const owned = data.ownedQuery || {};
  const all = owned.MediaAll || {};
  const yes = owned.MediaOwned || {};
  const no = owned.MediaNotOwned || {};
  return `<div class="d-flex flex-wrap align-items-center gap-3">
    <div class="d-flex flex-wrap align-items-center gap-2">
      <span class="btn btn-sm btn-outline-dark disabled">WMedia</span>
      <a href="${attr(all.Href || "/cards?mowned=1")}" class="${linkClass(all.Active, "btn btn-sm btn-primary waves-effect waves-light")}" aria-current="${all.Active ? "true" : "false"}">MAll</a>
      <a href="${attr(yes.Href || "/cards?mowned=3")}" class="${linkClass(yes.Active, "btn btn-sm btn-success waves-effect waves-light")}" aria-current="${yes.Active ? "true" : "false"}">MOwned</a>
      <a href="${attr(no.Href || "/cards?mowned=7")}" class="${linkClass(no.Active, "btn btn-sm btn-secondary waves-effect waves-light")}" aria-current="${no.Active ? "true" : "false"}">MNotOwned</a>
    </div>
    <span class="btn btn-sm btn-primary waves-effect waves-light disabled">Total: ${escapeHtml(data.total)}</span>
  </div>`;
}

function renderMovieCard(movie) {
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
  const playRow = renderPlayRow(movie);
  return `<div class="col">
    <article class="card shadow-sm h-100 card-lift">
      <a href="/movie/${attr(movie.name)}" class="d-block" aria-label="查看 ${attr(movie.name)} 详情">
        <div class="thumb-2x3">
          <img src="${attr(movie.jacket_img)}" alt="${attr(movie.name)} 海报" loading="lazy" decoding="async">
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
          <h6 class="mb-1 text-truncate" title="${attr(movie.title)}">
            <a href="/cards?pn=${encodeURIComponent(movie.prefix || "")}" class="stretched-link text-reset text-decoration-none">${escapeHtml(movie.name)}</a>
            ${watched}
          </h6>
          ${movie.title ? `<p class="mb-0 text-muted line-clamp-2" title="${attr(movie.title)}">${escapeHtml(movie.title)}</p>` : ""}
        </li>
        ${playRow}
        <li class="list-group-item text-muted">${genres}</li>
        <li class="list-group-item">${casts}${director}</li>
        <li class="list-group-item d-flex align-items-center small">
          ${movie.bus_url ? `<a href="${attr(movie.bus_url)}" class="me-2" target="_blank" rel="noopener">Bus</a>` : ""}
          ${movie.search_url ? `<a href="${attr(movie.search_url)}" class="me-2" target="_blank" rel="noopener">Suk</a>` : ""}
          ${movie.jav_url ? `<a href="${attr(movie.jav_url)}" class="me-2" target="_blank" rel="noopener">Lib</a>` : ""}
          <span class="ms-auto text-muted">${escapeHtml(movie.releasing_date)}</span>
        </li>
      </ul>
    </article>
  </div>`;
}

function renderPlayRow(movie) {
  if ((movie.owned_w_media || 0) < 4) {
    return "";
  }
  let play = "";
  if (movie.owned_w_media === 5) {
    play = `<a href="${attr(movie.video_url_w_media)}" class="btn btn-sm btn-info">MPlay</a>`;
  } else if (movie.owned_w_media === 4) {
    play = `<a href="${attr(movie.video_url_w_media)}" class="btn btn-sm btn-success">MPlay</a>`;
  } else if (movie.owned_w_media === 6) {
    play = `<span class="btn btn-sm btn-secondary disabled" aria-disabled="true">MRemoved</span>`;
  }
  return `<li class="list-group-item py-2">
    <div class="card-play-stack"><div class="card-play-row">
      ${play}
      ${movie.film_birth_date_w_media ? `<span class="text-muted small card-play-date">${escapeHtml(movie.film_birth_date_w_media)}</span>` : ""}
    </div></div>
  </li>`;
}

function renderFilter(data) {
  const filter = data.MovieCardFilter || {};
  const explicit = filter.Explicit || {};
  const input = (name, label, value, placeholder = "", type = "text", extra = "") => `<div class="col-12 col-md-6 col-xl-2">
    <label class="form-label mb-1">${escapeHtml(label)}</label>
    <input class="form-control" type="${attr(type)}" name="${attr(name)}" value="${attr(value || "")}" placeholder="${attr(placeholder)}" data-explicit="${explicit[name] ? "1" : "0"}" ${extra}>
  </div>`;
  return `<div id="movieCardFilterCollapse" class="collapse">
    <div class="card mb-3"><div class="card-body">
      <form id="movieCardFilterForm" class="row g-2 align-items-end" method="get" action="${attr(filter.Action || window.location.pathname)}">
        ${filter.PageSize ? `<input type="hidden" name="ps" value="${attr(filter.PageSize)}" data-explicit="${explicit.ps ? "1" : "0"}">` : ""}
        ${filter.OrderBy ? `<input type="hidden" name="od" value="${attr(filter.OrderBy)}" data-explicit="${explicit.od ? "1" : "0"}">` : ""}
        ${filter.RandomCount ? `<input type="hidden" name="n" value="${attr(filter.RandomCount)}" data-explicit="${explicit.n ? "1" : "0"}">` : ""}
        ${filter.PersonIds ? `<input type="hidden" name="pid" value="${attr(filter.PersonIds)}" data-explicit="${explicit.pid ? "1" : "0"}">` : ""}
        ${input("cn", "演员", filter.CastNames, "演员名，支持多值")}
        ${input("gn", "类型", filter.GenreNames, "类型名，支持多值")}
        ${input("dn", "导演", filter.DirectorName, "导演名")}
        ${input("pn", "前缀", filter.PrefixName, "前缀")}
        ${input("mn", "片商", filter.MakerName, "片商")}
        ${input("ln", "标签", filter.LabelName, "标签")}
        ${input("lj", "标签 JavID", filter.LabelJavID, "JavLib 标签 ID")}
        ${input("an", "相册名", filter.AlbumName, "相册名，支持多值")}
        ${input("wd", "关键词", filter.Word, "中文关键词")}
        ${renderSelect("mowned", "w_media 库存", filter.MediaOwned, explicit.mowned, [
          ["", "全部"], ["1", "全部电影"], ["2", "有 w_media 记录（含移除）"], ["3", "已拥有未移除"],
          ["4", "有字幕未移除"], ["5", "无字幕未移除"], ["6", "已移除"], ["7", "无 w_media 记录"]
        ])}
        ${renderSelect("nd", "下载状态", filter.NeedDownload, explicit.nd, [["", "全部"], ["1", "无需下载"], ["2", "需要下载"]])}
        ${input("drkmin", "在榜天数", filter.DaysInRankMin, "最小值", "number")}
        ${filter.HideDirs ? "" : `
          ${input("md1", "Media 目录 1", filter.MediaDir1, "目录名")}
          ${input("md2", "Media 目录 2", filter.MediaDir2, "目录名")}
          ${input("md3", "Media 目录 3", filter.MediaDir3, "目录名")}
          ${input("md4", "Media 目录 4", filter.MediaDir4, "目录名")}
        `}
        ${renderDatePair("上映日期", "rs", filter.ReleasingDateStart, "re", filter.ReleasingDateEnd, explicit)}
        ${renderDatePair("Media 下载时间", "mbs", filter.MediaBirthTimeStart, "mbe", filter.MediaBirthTimeEnd, explicit)}
        ${renderDatePair("上榜日期", "srds", filter.StartRankingDateStart, "srde", filter.StartRankingDateEnd, explicit)}
        ${renderDatePair("最后 SC 时间", "lsctmin", filter.LastScTimeMin, "lsctmax", filter.LastScTimeMax, explicit)}
        ${renderNumberPair("演员年龄", "cay", filter.CastAgeMin, "cao", filter.CastAgeMax, explicit, "0.1")}
        ${renderNumberPair("观看人数", "vwmin", filter.ViewWatchedMin, "vwmax", filter.ViewWatchedMax, explicit)}
        ${renderNumberPair("评分", "smin", filter.ScoreMin, "smax", filter.ScoreMax, explicit, "0.1")}
        ${renderNumberPair("SC 次数", "scmin", filter.ScTimesMin, "scmax", filter.ScTimesMax, explicit)}
        ${renderNumberPair("Come 次数", "comin", filter.ComeTimesMin, "comax", filter.ComeTimesMax, explicit)}
        <div class="col-12 d-flex gap-2">
          <button class="btn btn-primary" type="submit">筛选</button>
          <a class="btn btn-outline-secondary" href="${attr(filter.ClearHref || window.location.pathname)}">清空</a>
        </div>
      </form>
    </div></div>
  </div>`;
}

function renderSelect(name, label, value, explicit, options) {
  const opts = options.map(([optionValue, optionLabel]) =>
    `<option value="${attr(optionValue)}" ${String(value || "") === optionValue ? "selected" : ""}>${escapeHtml(optionLabel)}</option>`
  ).join("");
  return `<div class="col-12 col-md-6 col-xl-2">
    <label class="form-label mb-1">${escapeHtml(label)}</label>
    <select class="form-select" name="${attr(name)}" data-explicit="${explicit ? "1" : "0"}">${opts}</select>
  </div>`;
}

function renderDatePair(label, startName, startValue, endName, endValue, explicit) {
  return `<div class="col-12 col-xl-4">
    <label class="form-label mb-1">${escapeHtml(label)}</label>
    <div class="row g-2">
      <div class="col-6"><input class="form-control" type="text" name="${attr(startName)}" value="${attr(startValue || "")}" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}" data-explicit="${explicit[startName] ? "1" : "0"}"></div>
      <div class="col-6"><input class="form-control" type="text" name="${attr(endName)}" value="${attr(endValue || "")}" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}" data-explicit="${explicit[endName] ? "1" : "0"}"></div>
    </div>
  </div>`;
}

function renderNumberPair(label, startName, startValue, endName, endValue, explicit, step = "") {
  const stepAttr = step ? `step="${attr(step)}"` : "";
  return `<div class="col-12 col-xl-4">
    <label class="form-label mb-1">${escapeHtml(label)}</label>
    <div class="row g-2">
      <div class="col-6"><input class="form-control" type="number" ${stepAttr} name="${attr(startName)}" value="${attr(startValue || "")}" placeholder="最小值" data-explicit="${explicit[startName] ? "1" : "0"}"></div>
      <div class="col-6"><input class="form-control" type="number" ${stepAttr} name="${attr(endName)}" value="${attr(endValue || "")}" placeholder="最大值" data-explicit="${explicit[endName] ? "1" : "0"}"></div>
    </div>
  </div>`;
}

function renderScHistory(events) {
  const rows = asArray(events).map((event, index) => {
    const movies = asArray(get(event, "Movies", "movies"));
    const movieHtml = movies.length
      ? `<div class="d-flex flex-wrap gap-2">${movies.map((movie) => {
        const name = get(movie, "Name", "name");
        const come = get(movie, "IsCome", "isCome") ? `<span class="text-danger ms-1">COME</span>` : "";
        const href = get(movie, "Href", "href");
        if (href) {
          return `<a href="${attr(href)}" class="badge bg-light text-dark border text-decoration-none px-2 py-2">${escapeHtml(name)}${come}</a>`;
        }
        return `<span class="badge bg-light text-dark border px-2 py-2">${escapeHtml(name)}${come}</span>`;
      }).join("")}</div>`
      : `<span class="text-muted">-</span>`;
    return `<tr${index >= 5 ? ' class="sc-history-extra-row d-none"' : ""}>
      <td><a href="${attr(get(event, "Href", "href"))}">${formatUnixMinute(get(event, "ScTime", "scTime"))}</a></td>
      <td>${escapeHtml(get(event, "MovieCount", "movieCount") || 0)}</td>
      <td>${movieHtml}</td>
      <td>${humanDays(get(event, "Cooldown", "cooldown"))}</td>
    </tr>`;
  });
  const count = rows.length;
  const body = count
    ? `<div class="table-responsive">
        <table class="table table-hover align-middle mb-0">
          <thead><tr><th>时间</th><th>影片数</th><th>Movies</th><th>Cooldown</th></tr></thead>
          <tbody>${rows.join("")}</tbody>
        </table>
      </div>
      ${count > 5 ? `<div class="mt-3">
        <button id="btnToggleScHistory" class="btn btn-sm btn-outline-secondary" type="button" aria-expanded="false" data-expand-text="展开全部" data-collapse-text="收起">展开全部</button>
      </div>` : ""}`
    : `<div class="sc-empty-state">这个演员还没有关联的 SC 记录。</div>`;
  return `<section class="cast-section-card">
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h5 class="cast-section-title mb-0">SC 历史</h5>
      <span class="text-muted small">${escapeHtml(count)} 条</span>
    </div>
    ${body}
  </section>`;
}

function metric(label, value) {
  return `<div class="col-md-4">
    <div class="sc-metric-card">
      <div class="sc-metric-label">${escapeHtml(label)}</div>
      <div class="sc-metric-value">${value}</div>
    </div>
  </div>`;
}

function renderOverview(actor, aliases) {
  const personID = Number(get(actor, "Id", "id") || 0);
  const mergeMetric = personID > 0
    ? `<div class="col-md-4">
      <button class="cast-summary-button" type="button" data-bs-toggle="collapse" data-bs-target="#personMergeCollapse" aria-expanded="false" aria-controls="personMergeCollapse">
        <div class="sc-metric-card">
          <div class="sc-metric-label">Person Merge</div>
          <div class="sc-metric-value">#${escapeHtml(personID)}</div>
        </div>
      </button>
    </div>`
    : metric("Person Merge", "-");
  return `<section class="cast-section-card">
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h5 class="cast-section-title mb-0">概览</h5>
    </div>
    <div class="row g-3">
      ${metric("Movies", escapeHtml(get(actor, "MovieNumber", "movieNumber") || 0))}
      ${metric("WOwned", escapeHtml(get(actor, "OwnedWMediaNumber", "ownedWMediaNumber") || 0))}
      ${metric("SC Times", escapeHtml(get(actor, "ScTimes", "scTimes") || 0))}
      ${metric("Come Times", escapeHtml(get(actor, "ComeTimes", "comeTimes") || 0))}
      ${metric("Last SC", formatUnixDate(get(actor, "LastScTime", "lastScTime")))}
      ${metric("Last SC Event", formatUnixDate(get(actor, "LastScEventTime", "lastScEventTime")))}
      ${metric("Highest Rank", rankText(get(actor, "HighestRank", "highestRank")))}
      ${metric("Cast Alias", escapeHtml(asArray(aliases).length))}
      ${mergeMetric}
    </div>
  </section>`;
}

function renderPersonMerge(actor) {
  const personID = Number(get(actor, "Id", "id") || 0);
  if (personID <= 0) {
    return "";
  }
  return `<div id="personMergeCollapse" class="collapse mb-3">
    <div class="merge-inline-panel">
      <div class="small text-muted mb-3">当前详情页对应保留 person。搜索需要并入的 person，先预览，再执行合并。</div>
      <div class="input-group mb-3">
        <input id="mergeSearchInput" type="text" class="form-control" placeholder="搜索 person 名称、中文名、alias 或 cast 名">
        <button id="btnMergeSearch" type="button" class="btn btn-outline-primary">搜索候选</button>
      </div>
      <div class="mb-3">
        <div class="small text-muted mb-2">候选 person</div>
        <div id="mergeCandidateList" class="merge-list">
          <div class="merge-list-empty">输入关键词后开始搜索。</div>
        </div>
      </div>
      <div class="mb-3">
        <div class="small text-muted mb-2">待合并来源</div>
        <div id="mergeSelectedList" class="merge-list">
          <div class="merge-list-empty">还没有选择来源 person。</div>
        </div>
      </div>
      <div class="d-flex flex-wrap gap-2 mb-3">
        <button id="btnPreviewMerge" type="button" class="btn btn-primary" disabled>预览合并</button>
      </div>
      <div id="mergePreviewArea" class="merge-preview-grid" style="display:none;"></div>
    </div>
  </div>`;
}

function renderRepresentativeMovies(data) {
  const movies = asArray(data.movies);
  const pagination = renderPagination(data.PageInfo || data.pageInfo);
  return `<section class="cast-section-card">
    <div class="d-flex flex-column flex-lg-row justify-content-between align-items-lg-center gap-2 mb-3">
      <div>
        <h5 class="cast-section-title mb-0">代表影片</h5>
        <div class="text-muted small">当前模块与 Cards 同步，支持分页、排序、WOwned 与筛选。</div>
      </div>
      <div class="d-flex align-items-center">${pagination}</div>
    </div>
    <div class="d-flex flex-wrap justify-content-start align-items-center gap-2 mb-2">
      ${renderSortBar(data.sortQuery)}
    </div>
    <div class="d-flex flex-column flex-md-row justify-content-between align-items-md-center gap-2 mb-3">
      <div class="d-flex flex-wrap align-items-center gap-2">
        ${renderOwnedFilter(data)}
        ${data.MovieCardFilter ? `<button class="btn btn-sm btn-outline-secondary" type="button" data-bs-toggle="collapse" data-bs-target="#movieCardFilterCollapse" aria-expanded="false" aria-controls="movieCardFilterCollapse">筛选</button>` : ""}
      </div>
    </div>
    ${renderFilter(data)}
    ${movies.length ? `<div class="row row-cols-2 row-cols-sm-3 row-cols-lg-5 row-cols-xxl-6 g-3">${movies.map(renderMovieCard).join("")}</div>` : `<div class="sc-empty-state">暂时没有可展示的影片。</div>`}
  </section>`;
}

function renderMergeModal() {
  return `<div class="modal fade" id="personMergeConfirmModal" tabindex="-1" aria-labelledby="personMergeConfirmModalLabel" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" id="personMergeConfirmModalLabel">执行 Person Merge</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body">
          <div class="small text-muted mb-2">这会把来源 person 下的全部 am_cast 迁到当前 person，并删除多余 person。</div>
          <div id="mergeConfirmSummary" class="small"></div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
          <button type="button" class="btn btn-danger" id="btnExecuteMerge">确认执行</button>
        </div>
      </div>
    </div>
  </div>`;
}

function renderPage(data) {
  const page = data.Page || {};
  const actor = page.Actor || {};
  const personID = Number(get(actor, "Id", "id") || 0);
  return `<div class="d-flex flex-column flex-lg-row justify-content-between align-items-lg-center mb-3 gap-2">
    <div>
      <h4 class="mb-1">${escapeHtml(actorDisplayName(actor))}</h4>
      <div class="text-muted">查看演员概览、最近参与的 SC 历史，以及代表影片。</div>
    </div>
    <div class="d-flex flex-wrap gap-2">
      <button id="btnRebuildActorRank" type="button" class="btn btn-outline-warning" data-actor-name="${attr(get(actor, "Name", "name"))}">回填该演员 Rank</button>
      <a href="${attr(page.CardsHref || page.cardsHref || "/cards")}" class="btn btn-outline-primary">Cards</a>
    </div>
  </div>
  <div id="castMsgArea" class="alert alert-success small py-2 px-3 mb-3"></div>
  <div class="row g-3 mb-3 align-items-start">
    <div class="col-12 col-xl-6">${renderScHistory(page.RecentEvents || page.recentEvents)}</div>
    <div class="col-12 col-xl-6">${renderOverview(actor, page.ActorAliases || page.actorAliases)}</div>
  </div>
  ${renderPersonMerge(actor)}
  ${renderRepresentativeMovies(data)}
  ${personID > 0 ? renderMergeModal() : ""}`;
}

function initFilterForm() {
  const form = document.getElementById("movieCardFilterForm");
  if (!form) {
    return;
  }
  Array.from(form.elements).forEach((element) => {
    if (!shouldSkipElement(element)) {
      element.dataset.initialValue = normalizeValue(element);
    }
  });
  form.addEventListener("submit", (event) => {
    event.preventDefault();
    const action = form.getAttribute("action") || window.location.pathname;
    const target = new URL(action, window.location.origin);
    Array.from(form.elements).forEach((element) => {
      if (shouldSkipElement(element)) {
        return;
      }
      const name = element.name;
      const value = normalizeValue(element);
      const type = String(element.type || "").toLowerCase();
      const explicit = element.dataset.explicit === "1";
      const initialValue = element.dataset.initialValue || "";
      if ((type === "checkbox" || type === "radio") && !element.checked) {
        return;
      }
      if (type === "hidden") {
        if (explicit && value !== "") {
          target.searchParams.set(name, value);
        }
        return;
      }
      if (value === "") {
        return;
      }
      if (explicit || value !== initialValue) {
        target.searchParams.set(name, value);
      }
    });
    const query = target.searchParams.toString();
    window.location.assign(query ? `${target.pathname}?${query}` : target.pathname);
  });
}

function normalizeValue(element) {
  return String(element && element.value || "").trim();
}

function shouldSkipElement(element) {
  if (!element || !element.name || element.disabled) {
    return true;
  }
  const type = String(element.type || "").toLowerCase();
  return type === "submit" || type === "button" || type === "reset" || type === "fieldset";
}

function initCastActions(data) {
  const page = data.Page || {};
  const actor = page.Actor || {};
  const personId = Number(get(actor, "Id", "id") || 0);
  const msgArea = document.getElementById("castMsgArea");
  const btn = document.getElementById("btnRebuildActorRank");
  const mergeSearchInput = document.getElementById("mergeSearchInput");
  const btnMergeSearch = document.getElementById("btnMergeSearch");
  const mergeCandidateList = document.getElementById("mergeCandidateList");
  const mergeSelectedList = document.getElementById("mergeSelectedList");
  const btnPreviewMerge = document.getElementById("btnPreviewMerge");
  const mergePreviewArea = document.getElementById("mergePreviewArea");
  const mergeConfirmModal = document.getElementById("personMergeConfirmModal");
  const mergeConfirmSummary = document.getElementById("mergeConfirmSummary");
  const btnExecuteMerge = document.getElementById("btnExecuteMerge");
  const btnToggleScHistory = document.getElementById("btnToggleScHistory");
  const scHistoryExtraRows = Array.from(document.querySelectorAll(".sc-history-extra-row"));
  const selectedSources = new Map();
  let currentPreview = null;
  let scHistoryExpanded = false;

  function showMsg(text, ok = true) {
    if (!msgArea) return;
    msgArea.textContent = text;
    msgArea.className = `alert ${ok ? "alert-success" : "alert-danger"} small py-2 px-3 mb-3`;
    msgArea.style.display = "block";
  }

  function normalizePerson(person) {
    const row = person || {};
    return {
      id: Number(row.id || row.Id || 0),
      name: row.name || row.Name || "",
      chinese: row.chinese || row.Chinese || "",
      birthDay: row.birthDay || row.BirthDay || 0,
      movieNumber: row.movieNumber || row.MovieNumber || 0,
      ownedWMediaNumber: row.ownedWMediaNumber || row.OwnedWMediaNumber || 0,
      scTimes: row.scTimes || row.ScTimes || 0,
      comeTimes: row.comeTimes || row.ComeTimes || 0
    };
  }

  function personDisplayName(person) {
    const row = normalizePerson(person);
    return row.chinese || row.name || "(无名 person)";
  }

  function renderNameBadges(names) {
    if (!Array.isArray(names) || names.length === 0) {
      return `<span class="text-muted small">无 cast</span>`;
    }
    return names.map((name) => `<span class="cast-alias-badge">${escapeHtml(name)}</span>`).join("");
  }

  function renderCandidateRow(candidate, selected) {
    const person = normalizePerson(candidate && candidate.person ? candidate.person : candidate && candidate.Person);
    const castNames = Array.isArray(candidate && candidate.castNames) ? candidate.castNames : [];
    const primaryName = personDisplayName(person);
    const rawName = person.chinese && person.name ? `原名：${escapeHtml(person.name)}` : "";
    const birthDay = person.birthDay > 0 ? `生日：${formatUnixDate(person.birthDay)}` : "";
    const metrics = [
      `Movies ${person.movieNumber || 0}`,
      `WOwned ${person.ownedWMediaNumber || 0}`,
      `SC ${person.scTimes || 0}`,
      `COME ${person.comeTimes || 0}`
    ].join(" / ");

    return `<div class="merge-person-item${selected ? " is-selected" : ""}" data-person-id="${attr(person.id)}" data-person-name="${attr(person.name)}" data-person-chinese="${attr(person.chinese)}">
      <div class="merge-person-head">
        <div>
          <div class="merge-person-name">${escapeHtml(primaryName)} <span class="text-muted small">#${escapeHtml(person.id)}</span></div>
          <div class="merge-person-meta">
            ${rawName ? `<span>${rawName}</span>` : ""}
            ${birthDay ? `<span>${birthDay}</span>` : ""}
            <span>${escapeHtml(metrics)}</span>
          </div>
          <div class="cast-alias-wrap">${renderNameBadges(castNames)}</div>
        </div>
        <button type="button" class="btn btn-sm ${selected ? "btn-outline-secondary js-remove-source" : "btn-outline-primary js-add-source"}" data-person-id="${attr(person.id)}">${selected ? "移除" : "加入"}</button>
      </div>
    </div>`;
  }

  function renderCandidateList(candidates) {
    if (!mergeCandidateList) return;
    if (!Array.isArray(candidates) || candidates.length === 0) {
      mergeCandidateList.innerHTML = `<div class="merge-list-empty">没有找到可合并的候选 person。</div>`;
      return;
    }
    mergeCandidateList.innerHTML = candidates.map((candidate) => {
      const person = normalizePerson(candidate && candidate.person ? candidate.person : candidate && candidate.Person);
      return renderCandidateRow(candidate, selectedSources.has(person.id));
    }).join("");
  }

  function renderSelectedList() {
    if (!mergeSelectedList) return;
    const values = Array.from(selectedSources.values());
    if (values.length === 0) {
      mergeSelectedList.innerHTML = `<div class="merge-list-empty">还没有选择来源 person。</div>`;
      if (btnPreviewMerge) btnPreviewMerge.disabled = true;
      return;
    }
    mergeSelectedList.innerHTML = values.map((candidate) => renderCandidateRow(candidate, true)).join("");
    if (btnPreviewMerge) btnPreviewMerge.disabled = false;
  }

  function collectSourcePersonIds() {
    return Array.from(selectedSources.keys()).map((id) => Number(id)).filter((id) => Number.isFinite(id) && id > 0);
  }

  function request(url, options) {
    return fetch(url, options).then(async (resp) => {
      const body = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        throw new Error(body.error || `请求失败(${resp.status})`);
      }
      return body;
    });
  }

  function renderPreview(preview) {
    if (!mergePreviewArea) return;
    currentPreview = preview || null;
    if (!preview || !preview.keep) {
      mergePreviewArea.style.display = "none";
      mergePreviewArea.innerHTML = "";
      return;
    }

    const sourceHtml = (preview.sources || []).map((item) => {
      const person = normalizePerson(item.person || item.Person);
      return `<div class="merge-preview-box">
        <div class="merge-preview-label">来源 Person #${escapeHtml(person.id)}</div>
        <div class="merge-person-name">${escapeHtml(personDisplayName(person))}</div>
        <div class="cast-alias-wrap mt-2">${renderNameBadges(item.castNames || [])}</div>
      </div>`;
    }).join("");

    const keepPerson = normalizePerson(preview.keep.person || preview.keep.Person);
    mergePreviewArea.innerHTML = `<div class="merge-preview-box">
      <div class="merge-preview-label">保留 Person #${escapeHtml(keepPerson.id)}</div>
      <div class="merge-person-name">${escapeHtml(personDisplayName(keepPerson))}</div>
      <div class="cast-alias-wrap mt-2">${renderNameBadges(preview.keep.castNames || [])}</div>
    </div>
    ${sourceHtml}
    <div class="merge-preview-box">
      <div class="merge-preview-label">本次迁移</div>
      <div class="cast-alias-wrap mb-2">${renderNameBadges(preview.moveCastNames || [])}</div>
      <div class="small text-muted">删除 person：${escapeHtml((preview.removePersonIds || []).join(", ") || "-")}</div>
      <div class="small text-muted mt-1">受影响影片：${escapeHtml(preview.affectedMovieCount || 0)}</div>
      <div class="merge-preview-actions">
        <button type="button" class="btn btn-danger" id="btnOpenMergeConfirm">执行合并</button>
      </div>
    </div>`;
    mergePreviewArea.style.display = "grid";

    const btnOpenMergeConfirm = document.getElementById("btnOpenMergeConfirm");
    if (btnOpenMergeConfirm) {
      btnOpenMergeConfirm.addEventListener("click", () => {
        if (!mergeConfirmSummary || !mergeConfirmModal || !window.bootstrap || !window.bootstrap.Modal) return;
        mergeConfirmSummary.innerHTML = `<div>Keep Person #${escapeHtml(keepPerson.id)}</div>
          <div class="mt-1">删除 Person：${escapeHtml((preview.removePersonIds || []).join(", ") || "-")}</div>
          <div class="mt-1">迁移 Cast：${escapeHtml((preview.moveCastNames || []).join(", ") || "-")}</div>
          <div class="mt-1">受影响影片：${escapeHtml(preview.affectedMovieCount || 0)}</div>`;
        window.bootstrap.Modal.getOrCreateInstance(mergeConfirmModal).show();
      });
    }
  }

  function searchCandidates() {
    if (personId <= 0 || !mergeSearchInput) return;
    const keyword = (mergeSearchInput.value || "").trim();
    if (!keyword) {
      renderCandidateList([]);
      showMsg("请输入搜索关键词", false);
      return;
    }

    request(`/api/persons/${encodeURIComponent(personId)}/merge-candidates?q=${encodeURIComponent(keyword)}`, {method: "GET"})
      .then((body) => renderCandidateList(body.candidates || []))
      .catch((err) => showMsg(`搜索失败：${err.message}`, false));
  }

  if (btn) {
    btn.addEventListener("click", () => {
      const actorName = (btn.dataset.actorName || "").trim();
      if (!actorName) {
        showMsg("演员名为空，无法触发", false);
        return;
      }

      fetch("/api/triggers/cast/rebuild-rank", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({actorName})
      })
        .then((resp) => {
          if (resp.ok) {
            showMsg("该演员 Rank 回填已触发");
            return;
          }
          showMsg(`触发失败(${resp.status})`, false);
        })
        .catch((err) => showMsg(`异常：${err}`, false));
    });
  }

  if (btnMergeSearch) {
    btnMergeSearch.addEventListener("click", searchCandidates);
  }
  if (mergeSearchInput) {
    mergeSearchInput.addEventListener("keydown", (event) => {
      if (event.key === "Enter") {
        event.preventDefault();
        searchCandidates();
      }
    });
  }
  if (btnToggleScHistory && scHistoryExtraRows.length > 0) {
    btnToggleScHistory.addEventListener("click", () => {
      scHistoryExpanded = !scHistoryExpanded;
      scHistoryExtraRows.forEach((row) => row.classList.toggle("d-none", !scHistoryExpanded));
      btnToggleScHistory.textContent = scHistoryExpanded
        ? (btnToggleScHistory.dataset.collapseText || "收起")
        : (btnToggleScHistory.dataset.expandText || "展开全部");
      btnToggleScHistory.setAttribute("aria-expanded", scHistoryExpanded ? "true" : "false");
    });
  }
  if (mergeCandidateList) {
    mergeCandidateList.addEventListener("click", (event) => {
      const button = event.target.closest(".js-add-source, .js-remove-source");
      if (!button) return;
      const personIdValue = Number(button.dataset.personId || 0);
      if (!Number.isFinite(personIdValue) || personIdValue <= 0) return;
      const row = button.closest(".merge-person-item");
      if (!row) return;
      const castNames = Array.from(row.querySelectorAll(".cast-alias-badge")).map((el) => (el.textContent || "").trim()).filter(Boolean);
      const candidate = {
        person: {
          id: personIdValue,
          name: (row.dataset.personName || "").trim(),
          chinese: (row.dataset.personChinese || "").trim()
        },
        castNames
      };
      if (selectedSources.has(personIdValue)) {
        selectedSources.delete(personIdValue);
      } else {
        selectedSources.set(personIdValue, candidate);
      }
      renderSelectedList();
      currentPreview = null;
      renderPreview(null);
      searchCandidates();
    });
  }
  if (mergeSelectedList) {
    mergeSelectedList.addEventListener("click", (event) => {
      const button = event.target.closest(".js-remove-source");
      if (!button) return;
      const personIdValue = Number(button.dataset.personId || 0);
      if (!Number.isFinite(personIdValue) || personIdValue <= 0) return;
      selectedSources.delete(personIdValue);
      renderSelectedList();
      currentPreview = null;
      renderPreview(null);
      searchCandidates();
    });
  }
  if (btnPreviewMerge) {
    btnPreviewMerge.addEventListener("click", () => {
      const sourcePersonIds = collectSourcePersonIds();
      if (sourcePersonIds.length === 0) {
        showMsg("请先选择来源 person", false);
        return;
      }
      request("/api/persons/merge-preview", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({keepPersonId: personId, sourcePersonIds})
      })
        .then((body) => {
          renderPreview(body.preview);
          showMsg("合并预览已生成");
        })
        .catch((err) => showMsg(`预览失败：${err.message}`, false));
    });
  }
  if (btnExecuteMerge) {
    btnExecuteMerge.addEventListener("click", () => {
      if (!currentPreview) {
        showMsg("请先生成预览", false);
        return;
      }
      request("/api/persons/merge", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({keepPersonId: personId, sourcePersonIds: collectSourcePersonIds()})
      })
        .then((body) => {
          if (mergeConfirmModal && window.bootstrap && window.bootstrap.Modal) {
            window.bootstrap.Modal.getOrCreateInstance(mergeConfirmModal).hide();
          }
          showMsg("MergePerson 已完成，正在刷新页面");
          window.setTimeout(() => {
            window.location.href = `/cast/${encodeURIComponent(body.result && body.result.keepPersonId || personId)}`;
          }, 300);
        })
        .catch((err) => showMsg(`执行失败：${err.message}`, false));
    });
  }

  renderSelectedList();
}

async function main() {
  const response = await fetch(`${apiPathForCurrentPage()}${window.location.search}`);
  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || `请求失败(${response.status})`);
  }
  document.title = `${data.Title || "Cast"} - Rudy-GC`;
  root.innerHTML = renderPage(data);
  initFilterForm();
  initCastActions(data);
}

main().catch((err) => {
  root.innerHTML = `<div class="alert alert-danger">${escapeHtml(err.message)}</div>`;
});
