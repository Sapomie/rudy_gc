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

const apiPathByPage = new Map([
  ["/cards", "/api/cards"],
  ["/cardstoday", "/api/cardstoday"],
  ["/cardshasrank", "/api/cardshasrank"],
  ["/cardsmediamowned", "/api/cardsmediamowned"],
  ["/cardsneeddownload", "/api/cardsneeddownload"],
  ["/cardsrandom", "/api/cardsrandom"],
  ["/moviecarddayrank", "/api/moviecarddayrank"],
  ["/moviecardperiodrank", "/api/moviecardperiodrank"]
]);

function apiPathForCurrentPage() {
  return apiPathByPage.get(window.location.pathname) || "/api/cards";
}

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

function linkClass(active, base) {
  return `${base || ""} ${active ? "active disabled" : ""}`.trim();
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
  return `
    <div class="d-flex flex-wrap align-items-center gap-3">
      <div class="d-flex flex-wrap align-items-center gap-2">
        <span class="btn btn-sm btn-outline-dark disabled">WMedia</span>
        <a href="${attr(all.Href || "/cards?mowned=1")}" class="${linkClass(all.Active, "btn btn-sm btn-primary waves-effect waves-light")}" aria-current="${all.Active ? "true" : "false"}">MAll</a>
        <a href="${attr(yes.Href || "/cards?mowned=3")}" class="${linkClass(yes.Active, "btn btn-sm btn-success waves-effect waves-light")}" aria-current="${yes.Active ? "true" : "false"}">MOwned</a>
        <a href="${attr(no.Href || "/cards?mowned=7")}" class="${linkClass(no.Active, "btn btn-sm btn-secondary waves-effect waves-light")}" aria-current="${no.Active ? "true" : "false"}">MNotOwned</a>
      </div>
      <span class="btn btn-sm btn-primary waves-effect waves-light disabled">Total: ${escapeHtml(data.total)}</span>
      <button class="btn btn-sm btn-outline-secondary" type="button" data-bs-toggle="collapse" data-bs-target="#movieCardFilterCollapse" aria-expanded="false" aria-controls="movieCardFilterCollapse">筛选</button>
    </div>`;
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
  return `
    <div class="col">
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

function renderPeriodPlayRow(movie) {
  if ((movie.owned_w_media || 0) < 2) {
    return "";
  }
  let play = "";
  if (movie.owned_w_media === 5) {
    play = `<a href="${attr(movie.video_url_w_media)}" class="btn btn-sm btn-info me-2">MPlay</a>`;
  } else if (movie.owned_w_media === 4) {
    play = `<a href="${attr(movie.video_url_w_media)}" class="btn btn-sm btn-success me-2">MPlay</a>`;
  } else if (movie.owned_w_media === 6) {
    play = `<span class="btn btn-sm btn-secondary me-2 disabled" aria-disabled="true">MRemoved</span>`;
  }
  return `<li class="list-group-item d-flex align-items-center">
    ${play}
    ${movie.film_birth_date_w_media ? `<span class="text-muted small ms-2">| M ${escapeHtml(movie.film_birth_date_w_media)}</span>` : ""}
  </li>`;
}

function renderPeriodMovieCard(card) {
  const movie = card.Movie || card.movie || {};
  const item = card.Item || {};
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
  const watched = String(movie.score ?? "") !== "" && String(movie.viewers_number_watched ?? "") !== ""
    ? `<small class="text-muted ml-2">(${escapeHtml(movie.viewers_number_watched)})</small>`
    : "";
  const itemScore = Number(item.Score || 0).toFixed(2);
  return `
    <div class="col">
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
          <li class="list-group-item">
            <div class="d-flex align-items-center justify-content-between gap-2 flex-wrap">
              <span class="badge bg-dark">周期 #${escapeHtml(item.RankPos)}</span>
              <span class="badge ${attr(card.RankChangeClass || "bg-secondary")}">${escapeHtml(card.RankChangeText || "")}</span>
            </div>
            <div class="small text-muted mt-2">
              上期 ${escapeHtml(card.PrevRankText || "--")} · 峰值 #${escapeHtml(item.PeakRank)} · 最佳日 #${escapeHtml(item.BestRank)}
            </div>
            <div class="small text-muted mt-1">
              上榜 ${escapeHtml(item.DaysInRank)} 天 · 取样 ${escapeHtml(item.UsedPickDays)} 天 · 分数 ${escapeHtml(itemScore)}
            </div>
          </li>
          ${renderPeriodPlayRow(movie)}
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

function renderFilter(data) {
  const filter = data.MovieCardFilter || {};
  const explicit = filter.Explicit || {};
  const input = (name, label, value, placeholder = "", type = "text", extra = "") => `
    <div class="col-12 col-md-6 col-xl-2">
      <label class="form-label mb-1">${escapeHtml(label)}</label>
      <input class="form-control" type="${attr(type)}" name="${attr(name)}" value="${attr(value || "")}" placeholder="${attr(placeholder)}" data-explicit="${explicit[name] ? "1" : "0"}" ${extra}>
    </div>`;
  return `
    <div id="movieCardFilterCollapse" class="collapse">
      <div class="card mb-3"><div class="card-body">
        <form id="movieCardFilterForm" class="row g-2 align-items-end" method="get" action="${attr(filter.Action || "/cards")}">
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
            <a class="btn btn-outline-secondary" href="${attr(filter.ClearHref || "/cards")}">清空</a>
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
    const params = new URLSearchParams();
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
          params.set(name, value);
        }
        return;
      }
      if (value === "") {
        return;
      }
      if (explicit || value !== initialValue) {
        params.set(name, value);
      }
    });
    const action = form.getAttribute("action") || window.location.pathname;
    const query = params.toString();
    window.location.assign(query ? `${action}?${query}` : action);
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

function renderRankDayHeader(data) {
  const prevHref = `/moviecarddayrank?date=${encodeURIComponent(data.prevDate || "")}`;
  const nextHref = `/moviecarddayrank?date=${encodeURIComponent(data.nextDate || "")}`;
  const randHref = `/moviecarddayrank?date=${encodeURIComponent(data.randDate || "")}`;
  return `<div class="w-100 d-flex flex-column flex-md-row align-items-start align-items-md-center justify-content-between gap-2">
    <div class="alert alert-success py-2 mb-0">
      <div class="d-flex flex-column gap-2">
        <div class="d-flex align-items-center gap-2">
          <strong class="font-size-14">Rank Date</strong>
          <span class="font-18 fw-semibold">${escapeHtml(data.rankDate || "")}</span>
        </div>
        <div class="d-flex align-items-center gap-2">
          <strong class="font-size-14">Total</strong>
          <span class="font-18 fw-semibold">${escapeHtml(data.total)}</span>
        </div>
      </div>
    </div>
    <div class="d-flex flex-column flex-md-row gap-2 align-items-start align-items-md-center">
      <div class="btn-group btn-group-sm" role="group" aria-label="Rank day navigation">
        <a href="${attr(prevHref)}" class="btn btn-outline-primary ${data.prevDisabled ? "disabled" : ""}" aria-disabled="${data.prevDisabled ? "true" : "false"}">前一天</a>
        <a href="${attr(nextHref)}" class="btn btn-outline-primary ${data.nextDisabled ? "disabled" : ""}" aria-disabled="${data.nextDisabled ? "true" : "false"}">后一天</a>
        <a href="${attr(randHref)}" class="btn btn-outline-primary">随机</a>
      </div>
      <form class="d-flex gap-2" method="get" action="/moviecarddayrank">
        <input type="text" class="form-control form-control-sm" name="date" value="${attr(data.rankDate || "")}"
               placeholder="YYYY-MM-DD" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}" inputmode="numeric">
        <button class="btn btn-sm btn-outline-primary" type="submit">Go</button>
      </form>
    </div>
  </div>`;
}

function renderRankPeriodHeader(data) {
  const renderLinks = (links, activeClass, inactiveClass) => (links || []).map((item) =>
    `<a href="${attr(item.Href || "#")}" class="btn ${item.Active ? `${activeClass} active disabled` : inactiveClass}">${escapeHtml(item.Label)}</a>`
  ).join("");
  return `<div class="w-100">
    <div class="alert alert-success py-2 mb-0">
      <div class="d-flex flex-column flex-lg-row align-items-lg-center justify-content-between gap-2">
        <div class="d-flex flex-column gap-1">
          <div class="d-flex align-items-center gap-2 flex-wrap">
            <strong class="font-size-14">${escapeHtml(data.periodTypeLabel || "")}</strong>
            <span class="font-18 fw-semibold">${escapeHtml(data.periodKey || "")}</span>
            <span class="badge bg-light text-dark border">${escapeHtml(data.categoryLabel || "")}</span>
          </div>
          <div class="small text-muted">
            ${escapeHtml(data.rangeStart || "")} ~ ${escapeHtml(data.rangeEnd || "")} | Total ${escapeHtml(data.total)}
          </div>
        </div>
        <div class="d-flex flex-wrap gap-2">
          <div class="btn-group btn-group-sm" role="group" aria-label="Period type switch">
            ${renderLinks(data.typeLinks, "btn-primary", "btn-outline-primary")}
          </div>
          <div class="btn-group btn-group-sm" role="group" aria-label="Period category switch">
            ${renderLinks(data.categoryLinks, "btn-secondary", "btn-outline-secondary")}
          </div>
        </div>
      </div>
    </div>
  </div>`;
}

function renderRankPeriodNav(data) {
  return `<div class="btn-group btn-group-sm" role="group" aria-label="Period navigation">
    <a href="${attr(data.prevHref || "#")}" class="btn btn-outline-primary ${data.prevDisabled ? "disabled" : ""}" aria-disabled="${data.prevDisabled ? "true" : "false"}">上一期</a>
    <a href="${attr(data.nextHref || "#")}" class="btn btn-outline-primary ${data.nextDisabled ? "disabled" : ""}" aria-disabled="${data.nextDisabled ? "true" : "false"}">下一期</a>
    <a href="${attr(data.latestHref || "/moviecardperiodrank")}" class="btn btn-outline-primary">最新</a>
  </div>`;
}

function renderMessage(data) {
  document.getElementById("sort-bar").innerHTML = "";
  document.getElementById("owned-filter").innerHTML = "";
  document.getElementById("pagination-top").innerHTML = "";
  document.getElementById("movie-card-filter").innerHTML = "";
  document.getElementById("movie-grid").innerHTML = `<div class="col-12"><div class="alert alert-info">${escapeHtml(data.message || "暂无数据")}</div></div>`;
}

function renderCardsPage(data) {
  document.getElementById("sort-bar").innerHTML = renderSortBar(data.sortQuery);
  document.getElementById("owned-filter").innerHTML = renderOwnedFilter(data);
  document.getElementById("pagination-top").innerHTML = renderPagination(data.PageInfo);
  document.getElementById("movie-card-filter").innerHTML = renderFilter(data);
  document.getElementById("movie-grid").innerHTML = (data.movies || []).map(renderMovieCard).join("");
  initFilterForm();
}

function renderRankDayPage(data) {
  if (data.message) {
    renderMessage(data);
    return;
  }
  document.getElementById("sort-bar").innerHTML = renderRankDayHeader(data);
  document.getElementById("owned-filter").innerHTML = "";
  document.getElementById("pagination-top").innerHTML = renderPagination(data.PageInfo);
  document.getElementById("movie-card-filter").innerHTML = "";
  document.getElementById("movie-grid").innerHTML = (data.movies || []).map(renderMovieCard).join("");
}

function renderRankPeriodPage(data) {
  if (data.message) {
    renderMessage(data);
    return;
  }
  document.getElementById("sort-bar").innerHTML = renderRankPeriodHeader(data);
  document.getElementById("owned-filter").innerHTML = renderRankPeriodNav(data);
  document.getElementById("pagination-top").innerHTML = renderPagination(data.PageInfo);
  document.getElementById("movie-card-filter").innerHTML = "";
  document.getElementById("movie-grid").innerHTML = (data.cards || []).map(renderPeriodMovieCard).join("");
}

async function main() {
  const response = await fetch(`${apiPathForCurrentPage()}${window.location.search}`);
  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || `请求失败(${response.status})`);
  }
  document.title = `${data.Title || "MovieCard"} - Rudy-GC`;
  if (data.ViewType === "rank-day") {
    renderRankDayPage(data);
    return;
  }
  if (data.ViewType === "rank-period") {
    renderRankPeriodPage(data);
    return;
  }
  renderCardsPage(data);
}

main().catch((err) => {
  document.getElementById("movie-grid").innerHTML = `<div class="col-12"><div class="alert alert-danger">${escapeHtml(err.message)}</div></div>`;
});
