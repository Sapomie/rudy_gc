const apiPathByPage = new Map([
  ["/movie-agg-all/release", "/api/movie-agg-all/release"],
  ["/movie-agg-all/release-bucket-list", "/api/movie-agg-all/release-bucket-list"],
  ["/w-media-agg/bucket-list", "/api/w-media-agg/bucket-list"],
  ["/w-media-agg/birth", "/api/w-media-agg/birth"],
  ["/triggers/agg", "/api/triggers/agg"],
  ["/wdir", "/api/wdir"]
]);

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

function humanSize(bytes) {
  const value = Number(bytes || 0);
  if (value >= 1024 ** 4) return `${(value / 1024 ** 4).toFixed(1)} TB`;
  if (value >= 1024 ** 3) return `${(value / 1024 ** 3).toFixed(1)} GB`;
  if (value >= 1024 ** 2) return `${(value / 1024 ** 2).toFixed(1)} MB`;
  if (value >= 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${value} B`;
}

function formatUnix(value, withTime = false) {
  const n = Number(value || 0);
  if (!n) return "";
  const d = new Date(n * 1000);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  if (!withTime) return `${yyyy}-${mm}-${dd}`;
  const hh = String(d.getHours()).padStart(2, "0");
  const mi = String(d.getMinutes()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}`;
}

function apiPathForCurrentPage() {
  return apiPathByPage.get(window.location.pathname) || "/api/movie-agg-all/release";
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
    return `<a href="${attr(item.Href || "#")}" class="${linkClass(item.Active, "btn btn-outline-primary")}">${escapeHtml(label)}</a>`;
  }).join("");
  return `<div class="btn-group btn-group-sm flex-wrap" role="group" aria-label="Sort Options">${items}</div>`;
}

function renderOwnedFilter(data) {
  const owned = data.ownedQuery || {};
  const all = owned.MediaAll || {};
  const yes = owned.MediaOwned || {};
  const no = owned.MediaNotOwned || {};
  return `
    <span class="btn btn-sm btn-outline-dark disabled">WMedia</span>
    <a href="${attr(all.Href || "#")}" class="${linkClass(all.Active, "btn btn-sm btn-primary waves-effect waves-light")}">MAll</a>
    <a href="${attr(yes.Href || "#")}" class="${linkClass(yes.Active, "btn btn-sm btn-success waves-effect waves-light")}">MOwned</a>
    <a href="${attr(no.Href || "#")}" class="${linkClass(no.Active, "btn btn-sm btn-secondary waves-effect waves-light")}">MNotOwned</a>
    <span class="btn btn-sm btn-primary waves-effect waves-light disabled">Total: ${escapeHtml(data.total ?? data.Total ?? 0)}</span>`;
}

function renderPlayRow(movie) {
  if ((movie.owned_w_media || 0) < 4) return "";
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
          ${renderPlayRow(movie)}
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

function renderBreadcrumbs(items) {
  if (!items || !items.length) return "";
  return `<div class="page-subtitle mb-2">${items.map((item) =>
    item.Href
      ? `<a href="${attr(item.Href)}" class="text-muted text-decoration-none me-1">${escapeHtml(item.Title)}</a> /`
      : `<span class="text-dark">${escapeHtml(item.Title)}</span>`
  ).join(" ")}</div>`;
}

function renderModeSwitches(items) {
  return `<div class="d-flex flex-wrap gap-2">${(items || []).map((item) =>
    `<a href="${attr(item.Href)}" class="btn btn-sm ${item.Active ? "btn-primary" : "btn-outline-primary"}">${escapeHtml(item.Label)}</a>`
  ).join("")}</div>`;
}

function renderReleaseBucket(bucket) {
  return `<div class="col">
    <a href="${attr(bucket.Href)}" class="text-decoration-none">
      <div class="card bucket-card h-100 shadow-sm">
        <div class="card-body py-2">
          <div class="fw-semibold">${escapeHtml(bucket.Label)}</div>
          <div class="text-muted small">
            <span class="count">总数：${escapeHtml(bucket.CountAll)}</span>，
            <span class="count">已拥有：${escapeHtml(bucket.CountOwned)}</span><br>
            <span class="sizegb">${Number(bucket.SizeAllGB || 0).toFixed(1)} GB</span>
          </div>
        </div>
      </div>
    </a>
  </div>`;
}

function renderWMediaBucket(bucket) {
  return `<div class="col">
    <a href="${attr(bucket.Href)}" class="text-decoration-none">
      <div class="card bucket-card h-100 shadow-sm">
        <div class="card-body py-2">
          <div class="fw-semibold">${escapeHtml(bucket.Label)}</div>
          <div class="text-muted small">
            <span class="count">${escapeHtml(bucket.Count)}</span> 部，
            <span class="sizegb">${Number(bucket.SizeGB || 0).toFixed(1)} GB</span>
          </div>
        </div>
      </div>
    </a>
  </div>`;
}

function renderBucketRows(title, buckets, renderBucket, cols = "row-cols-2 row-cols-sm-3 row-cols-lg-6") {
  if (!buckets || !buckets.length) return "";
  return `<section class="bucket-section" aria-label="${attr(title)}">
    ${title ? `<div class="bucket-section-title">${escapeHtml(title)}</div>` : ""}
    <div class="row ${cols} g-3">${buckets.map(renderBucket).join("")}</div>
  </section>`;
}

function topHref(type, item, rangeStart, rangeEnd, options) {
  const params = new URLSearchParams();
  if (type === "cast" && Number(item.PersonId || item.AggID || 0) > 0) {
    params.set("pid", String(item.PersonId || item.AggID));
  } else if (type === "cast") {
    params.set("cn", item.Name || "");
  } else if (type === "director") {
    params.set("dn", item.Name || "");
  } else if (type === "label") {
    params.set("ln", item.Name || "");
  } else {
    params.set("pn", item.Name || "");
  }
  if (options.mediaOwned) params.set("mowned", options.mediaOwned);
  if (rangeStart) params.set(options.startKey, rangeStart);
  if (rangeEnd) params.set(options.endKey, rangeEnd);
  return `/cards?${params.toString()}`;
}

function renderTopStats(title, items, type, data, options) {
  if (!items || !items.length) return "";
  return `<div class="tab-pane p-2 active" role="tabpanel">
    <div class="${options.gridClass || "d-flex flex-wrap gap-2 justify-content-start"}">
      ${items.map((item, index) => {
        const href = topHref(type, item, data.RangeStart, data.RangeEnd, options);
        if (options.wmedia) {
          return `<a href="${attr(href)}" class="top-stat-card text-decoration-none" title="${attr(item.Name)}">
            <div class="card shadow-sm h-100"><div class="card-body p-2 text-center">
              <div class="top-stat-rank">#${index + 1}</div>
              <div class="top-stat-name text-primary">${escapeHtml(item.Name)}</div>
              <div class="top-stat-meta">出现：${escapeHtml(item.Count)}</div>
              <div class="top-stat-meta">体积：${humanSize(item.SizeBytes)}</div>
            </div></div>
          </a>`;
        }
        return `<a href="${attr(href)}" class="text-decoration-none" style="flex:0 0 calc(10% - 8px); min-width:120px;">
          <div class="card shadow-sm" style="background-color:#f8f9fa;"><div class="card-body p-2 text-center">
            <div class="mb-1">
              <span class="badge bg-secondary text-white fw-normal badge-slim">#${index + 1}</span>
              <span class="text-primary fw-semibold text-truncate">${escapeHtml(item.Name)}</span>
            </div>
            <div class="d-flex flex-column align-items-center gap-1">
              <span class="badge bg-light text-muted border badge-slim">总：<span class="text-dark fw-semibold">${escapeHtml(item.CountAll)}</span></span>
              <span class="badge bg-light text-muted border badge-slim">拥有：<span class="text-dark fw-semibold">${escapeHtml(item.CountOwned)}</span></span>
            </div>
          </div></div>
        </a>`;
      }).join("")}
    </div>
  </div>`;
}

function renderTopSection(data, groups, label, options) {
  const available = groups.filter((g) => g.items && g.items.length);
  if (!available.length) return "";
  return `<div class="section-card">
    <div class="d-flex align-items-center justify-content-start mb-2">
      <h4 class="section-title mb-0 me-3">${escapeHtml(label)}</h4>
      <button type="button" class="btn btn-outline-info waves-effect waves-light btn-sm" data-bs-toggle="collapse" data-bs-target="#topAggCollapse">展开</button>
    </div>
    <div class="collapse" id="topAggCollapse">
      <ul class="nav nav-tabs nav-tabs-custom nav-justified" role="tablist">
        ${available.map((g, i) => `<li class="nav-item"><a class="nav-link ${i === 0 ? "active" : ""}" data-bs-toggle="tab" href="#tab-top-${g.type}" role="tab">${escapeHtml(g.label)}</a></li>`).join("")}
      </ul>
      <div class="tab-content pt-3">
        ${available.map((g, i) => `<div class="tab-pane p-2 ${i === 0 ? "active" : ""}" id="tab-top-${g.type}" role="tabpanel">${renderTopStats(g.label, g.items, g.type, data, options)}</div>`).join("")}
      </div>
    </div>
  </div>`;
}

function renderMovieToolbar(data) {
  return `<div class="position-sticky pb-2 pt-2 mb-3 border-bottom" style="top:0; z-index:1030; background:rgba(255,255,255,.9); backdrop-filter: blur(6px);">
    <div class="container-fluid px-0">
      <div class="d-flex flex-wrap justify-content-start align-items-center mb-2">${data.SortQuery ? renderSortBar(data.SortQuery) : ""}</div>
      <div class="d-flex flex-column flex-md-row justify-content-between align-items-md-center gap-2">
        <div class="d-flex flex-wrap align-items-center gap-2">${data.ownedQuery ? renderOwnedFilter(data) : ""}</div>
        <div class="d-flex align-items-center ms-md-3">${data.PageInfo ? renderPagination(data.PageInfo) : ""}</div>
      </div>
    </div>
  </div>`;
}

function renderMovieGrid(movies, emptyText) {
  if (!movies || !movies.length) {
    return `<div class="row"><div class="col-12"><div class="alert alert-info mb-0">${escapeHtml(emptyText)}</div></div></div>`;
  }
  return `<div class="row row-cols-2 row-cols-sm-3 row-cols-lg-5 row-cols-xxl-6 g-3" id="movie-grid">${movies.map(renderMovieCard).join("")}</div>`;
}

function renderMovieAggRelease(data) {
  const hasYearTabs = (data.BucketsQAll && data.BucketsQAll.length) || (data.BucketsMAll && data.BucketsMAll.length);
  const bucketHtml = hasYearTabs
    ? `<ul class="nav nav-tabs nav-tabs-custom nav-justified" role="tablist">
        ${data.BucketsQAll?.length ? `<li class="nav-item"><a class="nav-link active" data-bs-toggle="tab" href="#tab-year-quarters" role="tab">季度</a></li>` : ""}
        ${data.BucketsMAll?.length ? `<li class="nav-item"><a class="nav-link ${data.BucketsQAll?.length ? "" : "active"}" data-bs-toggle="tab" href="#tab-year-months" role="tab">月份</a></li>` : ""}
      </ul>
      <div class="tab-content pt-3">
        ${data.BucketsQAll?.length ? `<div class="tab-pane p-2 active" id="tab-year-quarters" role="tabpanel">${renderBucketRows("", data.BucketsQAll, renderReleaseBucket, "row-cols-2 row-cols-sm-2 row-cols-lg-4")}</div>` : ""}
        ${data.BucketsMAll?.length ? `<div class="tab-pane p-2 ${data.BucketsQAll?.length ? "" : "active"}" id="tab-year-months" role="tabpanel">${renderBucketRows("", data.BucketsMAll, renderReleaseBucket, "row-cols-3 row-cols-sm-4 row-cols-lg-6")}</div>` : ""}
      </div>`
    : (data.BucketsAll?.length ? renderBucketRows("", data.BucketsAll, renderReleaseBucket) : `<div class="alert alert-info mb-0">暂无聚合数据。</div>`);

  return `
    <div class="section-card">
      <div class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-2">${renderModeSwitches(data.ModeSwitches)}</div>
      ${renderBreadcrumbs(data.Breadcrumbs)}
      <h4 class="section-title">时间聚合</h4>
      ${bucketHtml}
    </div>
    ${renderTopSection(data, [
      { label: "演员", type: "cast", items: data.TopCastsAll },
      { label: "导演", type: "director", items: data.TopDirectorsAll },
      { label: "Label", type: "label", items: data.TopLabelsAll },
      { label: "Prefix", type: "prefix", items: data.TopPrefixesAll }
    ], `Top 聚合（${data.AggModeLabel || ""}）`, { mediaOwned: data.AggMode === "owned" ? "3" : "1", startKey: "rs", endKey: "re" })}
    ${data.Movies !== undefined ? `<div class="section-card">${renderMovieToolbar(data)}${renderMovieGrid(data.movies || data.Movies, "暂无影片。")}</div>` : ""}`;
}

function renderWMediaAggBirth(data) {
  const yearTabs = data.Level === "year" && ((data.BucketsQ && data.BucketsQ.length) || (data.BucketsM && data.BucketsM.length));
  const buckets = yearTabs
    ? `<ul class="nav nav-tabs nav-tabs-custom nav-justified" role="tablist">
        ${data.BucketsQ?.length ? `<li class="nav-item"><a class="nav-link active" data-bs-toggle="tab" href="#tab-wmagg-quarters" role="tab">季度</a></li>` : ""}
        ${data.BucketsM?.length ? `<li class="nav-item"><a class="nav-link ${data.BucketsQ?.length ? "" : "active"}" data-bs-toggle="tab" href="#tab-wmagg-months" role="tab">月份</a></li>` : ""}
      </ul>
      <div class="tab-content pt-3">
        ${data.BucketsQ?.length ? `<div class="tab-pane p-2 active" id="tab-wmagg-quarters" role="tabpanel">${renderBucketRows("", data.BucketsQ, renderWMediaBucket, "row-cols-2 row-cols-sm-2 row-cols-lg-4")}</div>` : ""}
        ${data.BucketsM?.length ? `<div class="tab-pane p-2 ${data.BucketsQ?.length ? "" : "active"}" id="tab-wmagg-months" role="tabpanel">${renderBucketRows("", data.BucketsM, renderWMediaBucket, "row-cols-3 row-cols-sm-4 row-cols-lg-6")}</div>` : ""}
      </div>`
    : [
      renderBucketRows("年份", data.BucketsY, renderWMediaBucket),
      renderBucketRows("季度", data.BucketsQ, renderWMediaBucket, "row-cols-2 row-cols-sm-2 row-cols-lg-4"),
      renderBucketRows("月份", data.BucketsM, renderWMediaBucket, "row-cols-3 row-cols-sm-4 row-cols-lg-6"),
      renderBucketRows("日期", data.BucketsD, renderWMediaBucket, "row-cols-3 row-cols-sm-4 row-cols-lg-6")
    ].join("") || `<div class="alert alert-info mb-0">当前范围暂无可展示的时间桶。<a href="/triggers/agg" class="alert-link">如为新建统计表，请先到聚合回填页执行 WMedia 时间聚合回填。</a></div>`;

  return `
    <div class="w-media-agg-page">
      <div class="section-card">${renderBreadcrumbs(data.Breadcrumbs)}<h4 class="section-title">${escapeHtml(data.Title)}</h4>${buckets}</div>
      ${renderTopSection(data, [
        { label: "演员", type: "cast", items: data.TopCasts },
        { label: "导演", type: "director", items: data.TopDirectors },
        { label: "Label", type: "label", items: data.TopLabels },
        { label: "Prefix", type: "prefix", items: data.TopPrefixes }
      ], "Top 聚合", { mediaOwned: "3", startKey: "mbs", endKey: "mbe", wmedia: true, gridClass: "top-stat-grid" })}
      <div class="section-card">${renderMovieToolbar(data)}${renderMovieGrid(data.movies || data.Movies, "当前范围暂无媒体卡片。")}</div>
    </div>`;
}

function pageSizeOptions(current) {
  return [24, 50, 100, 200].map((n) => `<option value="${n}" ${Number(current) === n ? "selected" : ""}>${n}</option>`).join("");
}

function renderSortButtonLinks(items) {
  return (items || []).map((item) =>
    `<a href="${attr(item.Href)}" class="btn btn-sm ${item.Active ? "btn-primary" : "btn-outline-primary"}">${escapeHtml(item.Label)}</a>`
  ).join("");
}

function renderHeaderLink(item) {
  if (!item) return "";
  return `<a class="bucket-sort-link ${item.Active ? "is-active" : ""}" href="${attr(item.Href)}">${escapeHtml(item.Label)}${item.Active ? `<span class="bucket-sort-indicator">${item.Desc ? "↓" : "↑"}</span>` : ""}</a>`;
}

function renderReleaseBucketList(data) {
  const q = data.Query || {};
  const rows = data.Rows || [];
  return `<div class="movie-release-bucket-list-page" id="movieReleaseBucketListPage">
    <div class="d-flex flex-wrap align-items-center gap-2 mb-3">
      ${renderModeSwitches(data.ModeSwitches)}
      <a href="${attr(data.AggPageHref)}" class="btn btn-outline-primary btn-sm">返回聚合页</a>
      <a href="/triggers/agg" class="btn btn-outline-secondary btn-sm">去回填入口</a>
    </div>
    ${renderBucketSummary(data.Total, data.PageInfo, rows.length)}
    ${renderBucketFilter(data.ListPath || "/movie-agg-all/release-bucket-list", q, data.ClearHref, true, data.AggMode)}
    ${renderBucketControls(data.SortLinks, data.LevelLinks)}
    <div class="card"><div class="card-body"><div class="table-responsive">
      <table class="table table-hover align-middle mb-0 bucket-table">
        <thead><tr>
          <th>${renderHeaderLink(data.HeaderLinks?.Bucket)}</th>
          <th>${renderHeaderLink(data.HeaderLinks?.CountAll)}</th>
          <th>${renderHeaderLink(data.HeaderLinks?.CountOwned)}</th>
          <th>${renderHeaderLink(data.HeaderLinks?.Size)}</th>
          <th>${renderHeaderLink(data.HeaderLinks?.LatestRelease)}</th>
        </tr></thead>
        <tbody>${rows.length ? rows.map((row) => `<tr>
          <td class="text-nowrap">${row.ViewHref ? `<a class="bucket-name-link" href="${attr(row.ViewHref)}">${escapeHtml(row.BucketDisplay)}</a>` : `<span class="bucket-name">${escapeHtml(row.BucketDisplay)}</span>`}</td>
          <td class="text-nowrap">${escapeHtml(row.CountAll)}</td>
          <td class="text-nowrap">${escapeHtml(row.CountOwned)}</td>
          <td class="text-nowrap">${humanSize(row.SizeBytes)}</td>
          <td class="text-nowrap">${formatUnix(row.LatestReleasingDate)}</td>
        </tr>`).join("") : `<tr><td colspan="5"><div class="bucket-empty-state">当前筛选下没有 \`movie_release_bucket_stat\` 数据。若刚建表，请先去 <a href="/triggers/agg">聚合回填页</a> 执行一次“回填上映日时间聚合”。</div></td></tr>`}</tbody>
      </table>
    </div></div></div>
    ${rows.length ? `<div class="d-flex justify-content-end mt-3">${renderPagination(data.PageInfo)}</div>` : ""}
  </div>`;
}

function renderWMediaBucketList(data) {
  const q = data.Query || {};
  const rows = data.Rows || [];
  return `<div class="w-media-bucket-list-page" id="wMediaBirthBucketListPage">
    <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
      <div><h4 class="mb-1">${escapeHtml(data.PageTitle)}</h4><p class="text-muted mb-0">${escapeHtml(data.PageNote)}</p></div>
      <div class="d-flex gap-2"><a href="${attr(data.AggPageHref)}" class="btn btn-outline-primary btn-sm">返回聚合页</a><a href="/triggers/agg" class="btn btn-outline-secondary btn-sm">去回填入口</a></div>
    </div>
    ${renderBucketSummary(data.Total, data.PageInfo, rows.length)}
    ${renderBucketFilter("/w-media-agg/bucket-list", q, data.ClearHref, false)}
    ${renderBucketControls(data.SortLinks, data.LevelLinks)}
    <div class="card"><div class="card-body"><div class="table-responsive">
      <table class="table table-hover align-middle mb-0 bucket-table">
        <thead><tr>
          <th>${renderHeaderLink(data.HeaderLinks?.Bucket)}</th>
          <th>${renderHeaderLink(data.HeaderLinks?.Media)}</th>
          <th>${renderHeaderLink(data.HeaderLinks?.Removed)}</th>
          <th>${renderHeaderLink(data.HeaderLinks?.Size)}</th>
          <th>${renderHeaderLink(data.HeaderLinks?.Subtitle)}</th>
          <th>${renderHeaderLink(data.HeaderLinks?.LatestBirth)}</th>
        </tr></thead>
        <tbody>${rows.length ? rows.map((row) => `<tr>
          <td class="text-nowrap">${row.ViewHref ? `<a class="bucket-name-link" href="${attr(row.ViewHref)}">${escapeHtml(row.BucketDisplay)}</a>` : `<span class="bucket-name">${escapeHtml(row.BucketDisplay)}</span>`}</td>
          <td class="text-nowrap">${escapeHtml(row.MediaCount)}</td>
          <td class="text-nowrap">${escapeHtml(row.RemovedCount)}</td>
          <td class="text-nowrap">${humanSize(row.SizeBytes)}</td>
          <td class="text-nowrap">${escapeHtml(row.HasSubCount)}</td>
          <td class="text-nowrap">${formatUnix(row.LatestBirthTime, true)}</td>
        </tr>`).join("") : `<tr><td colspan="6"><div class="bucket-empty-state">当前筛选下没有 \`w_media_birth_bucket_stat\` 数据。若刚建表，请先去 <a href="/triggers/agg">聚合回填页</a> 执行一次“回填 WMedia 时间聚合”。</div></td></tr>`}</tbody>
      </table>
    </div></div></div>
    ${rows.length ? `<div class="d-flex justify-content-end mt-3">${renderPagination(data.PageInfo)}</div>` : ""}
  </div>`;
}

function renderBucketSummary(total, pageInfo, hasRows) {
  return `<div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
    <div class="bucket-list-summary"><span class="bucket-list-summary-label">总行数</span><span class="bucket-list-summary-value">${escapeHtml(total)}</span></div>
    ${hasRows ? `<div class="ms-xl-auto">${renderPagination(pageInfo)}</div>` : ""}
  </div>`;
}

function renderBucketFilter(action, q, clearHref, hasDay, aggMode) {
  return `<div class="card mb-3"><div class="card-body">
    <form class="row g-2 align-items-end" method="get" action="${attr(action)}">
      ${aggMode === "owned" ? `<input type="hidden" name="agg_mode" value="owned">` : ""}
      <input type="hidden" name="sort" value="${attr(q.Sort || "")}">
      <input type="hidden" name="dir" value="${attr(q.Dir || "")}">
      <input type="hidden" name="level" value="${attr(q.Level || "")}">
      <div class="col-12 col-md-6 col-xl-4"><label class="form-label mb-1">时间桶关键字</label><input class="form-control" type="text" name="scope_key" value="${attr(q.ScopeKey || "")}" placeholder="如 2026-04、2026-Q02"></div>
      <div class="col-12 col-md-4 col-xl-2"><label class="form-label mb-1">Year</label><input class="form-control" type="number" name="year" value="${attr(q.Year || "")}" placeholder="如 2026"></div>
      <div class="col-12 col-md-4 col-xl-1"><label class="form-label mb-1">Quarter</label><input class="form-control" type="number" name="quarter" value="${attr(q.Quarter || "")}" placeholder="1-4"></div>
      <div class="col-12 col-md-4 col-xl-1"><label class="form-label mb-1">Month</label><input class="form-control" type="number" name="month" value="${attr(q.Month || "")}" placeholder="1-12"></div>
      ${hasDay ? `<div class="col-12 col-md-4 col-xl-1"><label class="form-label mb-1">Day</label><input class="form-control" type="number" name="day" value="${attr(q.Day || "")}" placeholder="1-31"></div>` : ""}
      <div class="col-12 col-md-6 col-xl-2"><label class="form-label mb-1">每页条数</label><select class="form-select" name="ps">${pageSizeOptions(q.PageSize)}</select></div>
      <div class="col-12 col-xl-2 d-flex gap-2"><button class="btn btn-primary" type="submit">筛选</button><a class="btn btn-outline-secondary" href="${attr(clearHref)}">清空</a></div>
    </form>
  </div></div>`;
}

function renderBucketControls(sortLinks, levelLinks) {
  return `<div class="card mb-3"><div class="card-body">
    <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3">
      <div class="d-flex flex-wrap align-items-center gap-2"><span class="text-muted small me-1">排序</span>${renderSortButtonLinks(sortLinks)}</div>
      <div class="d-flex flex-wrap align-items-center justify-content-xl-end gap-2 bucket-level-switch"><span class="text-muted small me-1">层级</span>${renderSortButtonLinks(levelLinks)}</div>
    </div>
  </div></div>`;
}

function renderAggTriggers(data) {
  return `<div id="aggTriggersPage">
    <div class="row mb-3"><div class="col-12 d-flex align-items-center justify-content-between">
      <div><h4 class="mb-1">${escapeHtml(data.PageTitle)}</h4><p class="text-muted mb-0">${escapeHtml(data.PageNote)}</p></div>
      <div class="d-flex gap-2"><a href="/w-media-agg/birth" class="btn btn-outline-primary btn-sm">WMedia 聚合页</a><a href="/movie-agg-all/release" class="btn btn-outline-secondary btn-sm">上映日聚合页</a></div>
    </div></div>
    <div class="row g-3">
      ${renderTriggerCard("WMedia 时间聚合回填", "新建 `w_media` 时间聚合表后，在这里执行一次回填，旧数据才会写入 `/w-media-agg/birth`。", "agg", "btnBackfillWMediaAgg")}
      ${renderTriggerCard("上映日时间聚合回填", "新建 `movie_release_*` 表后，在这里执行一次回填，旧数据才会写入 `/movie-agg-all/release`。", "releaseAgg", "btnBackfillMovieReleaseAgg")}
    </div>
  </div>`;
}

function renderTriggerCard(title, note, prefix, buttonId) {
  const labels = prefix === "agg"
    ? [["ClearedBucketRows", "清空桶行"], ["ClearedTopRows", "清空 Top 行"], ["ClearedDirtyRows", "清空脏标记"], ["DirtyDays", "脏日期数"], ["YearBuckets", "年桶"], ["QuarterBuckets", "季桶"], ["MonthBuckets", "月桶"], ["DayBuckets", "日桶"], ["TopRows", "Top 行"], ["ElapsedMs", "耗时"]]
    : [["ClearedBucketRows", "清空桶行"], ["ClearedTopRows", "清空 Top 行"], ["ClearedDirtyRows", "清空脏标记"], ["DirtyMonths", "脏月份数"], ["YearBuckets", "年桶"], ["QuarterBuckets", "季桶"], ["MonthBuckets", "月桶"], ["DayBuckets", "日桶"], ["TopRows", "Top 行"], ["ElapsedMs", "耗时"]];
  return `<div class="col-xl-6"><div class="card h-100"><div class="card-body">
    <div class="d-flex align-items-center justify-content-between gap-2 mb-3"><h5 class="card-title mb-0">${escapeHtml(title)}</h5><button type="button" id="${buttonId}" class="btn btn-outline-primary btn-sm">执行回填</button></div>
    <div class="text-muted small mb-3">${escapeHtml(note)}</div>
    <div class="row g-2">${labels.map(([key, label]) => `<div class="col-6"><div class="overview-item"><div class="label">${label}</div><div class="value" id="${prefix}${key}">0</div></div></div>`).join("")}</div>
    <div class="mt-3"><div class="text-muted small">状态</div><div id="${prefix}BackfillStatusText" class="fw-semibold">尚未执行回填</div></div>
  </div></div></div>`;
}

function initAggTriggers() {
  const msgArea = document.getElementById("msgArea");
  let loading = false;
  let msgTimer = null;
  const showMsg = (text, ok) => {
    clearTimeout(msgTimer);
    msgArea.textContent = text;
    msgArea.className = `alert ${ok ? "alert-success" : "alert-danger"} fade show small py-2 px-3`;
    msgArea.style.display = "block";
    msgTimer = setTimeout(() => { msgArea.style.display = "none"; }, 3000);
  };
  const setLoading = (next) => {
    loading = !!next;
    document.getElementById("btnBackfillWMediaAgg").disabled = loading;
    document.getElementById("btnBackfillMovieReleaseAgg").disabled = loading;
  };
  const request = (url) => fetch(url, { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" }).then(async (resp) => {
    const payload = await resp.json().catch(() => ({}));
    const result = payload && payload.result ? payload.result : null;
    if (!resp.ok) {
      const err = new Error((payload && payload.error) || `请求失败(${resp.status})`);
      err.result = result && result.result ? result.result : result;
      throw err;
    }
    return result && result.result ? result.result : result;
  });
  const renderState = (prefix, data, dirtyKey) => {
    const safe = data || {};
    const map = {
      ClearedBucketRows: "cleared_bucket_rows",
      ClearedTopRows: "cleared_top_rows",
      ClearedDirtyRows: "cleared_dirty_rows",
      DirtyDays: "dirty_days",
      DirtyMonths: "dirty_months",
      YearBuckets: "year_buckets",
      QuarterBuckets: "quarter_buckets",
      MonthBuckets: "month_buckets",
      DayBuckets: "day_buckets",
      TopRows: "top_rows",
      ElapsedMs: "elapsed_ms"
    };
    for (const [suffix, key] of Object.entries(map)) {
      const el = document.getElementById(`${prefix}${suffix}`);
      if (el) el.textContent = suffix === "ElapsedMs" ? `${safe[key] || 0} ms` : String(safe[key] || 0);
    }
    const status = document.getElementById(`${prefix}BackfillStatusText`);
    if (status) status.textContent = Object.keys(safe).length ? `回填完成：日桶 ${safe.day_buckets || 0}，Top 行 ${safe.top_rows || 0}` : "尚未执行回填";
    void dirtyKey;
  };
  document.getElementById("btnBackfillWMediaAgg").addEventListener("click", () => {
    setLoading(true);
    document.getElementById("aggBackfillStatusText").textContent = "正在回填旧的 WMedia 时间聚合数据...";
    request("/api/agg/w-media/backfill").then((data) => {
      renderState("agg", data, "dirty_days");
      showMsg("WMedia 时间聚合回填完成", true);
    }).catch((err) => {
      showMsg(err.message || "WMedia 时间聚合回填失败", false);
      if (err.result) renderState("agg", err.result, "dirty_days");
      document.getElementById("aggBackfillStatusText").textContent = "回填失败，请查看提示信息";
    }).finally(() => setLoading(false));
  });
  document.getElementById("btnBackfillMovieReleaseAgg").addEventListener("click", () => {
    setLoading(true);
    document.getElementById("releaseAggBackfillStatusText").textContent = "正在回填...";
    request("/api/agg/movie-release/backfill").then((data) => {
      renderState("releaseAgg", data, "dirty_months");
      showMsg("上映日时间聚合回填完成", true);
    }).catch((err) => {
      showMsg(err.message || "上映日时间聚合回填失败", false);
      if (err.result) renderState("releaseAgg", err.result, "dirty_months");
    }).finally(() => setLoading(false));
  });
}

function renderChildren(children) {
  const items = children && children.Items || [];
  return `<section class="mt-3" aria-label="Subdirectories">
    <div class="d-flex align-items-center justify-content-between mb-2"><h5 class="mb-0">${escapeHtml(children?.Title || "子目录")}</h5>${children?.MoreURL ? `<a class="small" href="${attr(children.MoreURL)}">更多</a>` : ""}</div>
    ${items.length ? `<div class="row row-cols-2 row-cols-sm-3 row-cols-lg-5 row-cols-xxl-6 g-3">${items.map((item) => {
      const stat = item.Stats && item.Stats[0] || {};
      return `<div class="col"><a href="/wdir?id=${attr(item.Summary?.id || item.Summary?.Id)}" class="text-decoration-none">
        <div class="card h-100 shadow-sm"><div class="card-body py-2"><div class="d-flex align-items-start"><div class="flex-grow-1">
          <div class="fw-semibold text-truncate" title="${attr(item.Summary?.name || item.Summary?.Name)}">${escapeHtml(item.Summary?.name || item.Summary?.Name)}</div>
          <div class="text-muted small">${escapeHtml(stat.film_count || stat.FilmCount || 0)} 个媒体 · ${humanSize(stat.total_size || stat.TotalSize || 0)}</div>
        </div></div></div><div class="card-footer py-1 text-muted small">更新：${formatUnix(item.Summary?.updated_on || item.Summary?.UpdatedOn, true)}</div></div>
      </a></div>`;
    }).join("")}</div>` : `<div class="alert alert-light border mb-0">暂无子目录。</div>`}
  </section>`;
}

function renderWDirList(data) {
  return `<div class="d-flex flex-column flex-md-row justify-content-between align-items-start align-items-md-center gap-3 mb-3">
    <div><h3 class="mb-1">媒体目录</h3><p class="text-muted mb-0">按完整目录树浏览 \`w_folder / w_media\`。</p></div>
    <div class="d-flex flex-wrap align-items-center gap-2"><span class="btn btn-sm btn-outline-secondary disabled">目录数 ${escapeHtml(data.Total)}</span>${renderPagination(data.PageInfo)}</div>
  </div>${renderChildren(data.Children)}`;
}

function renderWDirBreadcrumbs(detail) {
  const crumbs = detail?.breadcrumbs || detail?.Breadcrumbs || [];
  return `<nav aria-label="breadcrumb"><ol class="breadcrumb mb-1">
    <li class="breadcrumb-item"><a href="/wdir">根目录</a></li>
    ${crumbs.map((b, i) => i < crumbs.length - 1
      ? `<li class="breadcrumb-item"><a href="/wdir?id=${attr(b.id || b.Id)}">${escapeHtml(b.name || b.Name)}</a></li>`
      : `<li class="breadcrumb-item active" aria-current="page">${escapeHtml(b.name || b.Name)}</li>`).join("")}
  </ol></nav>`;
}

function renderWDirStats(detail) {
  const stats = detail?.stats || detail?.Stats;
  if (!stats) return "";
  return `<div class="row g-2">
    <div class="col-6 col-md-3"><div class="card shadow-sm h-100"><div class="card-body"><div class="text-muted small">媒体数</div><div class="h4 mb-0">${escapeHtml(stats.film_count || stats.FilmCount || 0)}</div></div></div></div>
    <div class="col-6 col-md-3"><div class="card shadow-sm h-100"><div class="card-body"><div class="text-muted small">总大小</div><div class="h4 mb-0">${humanSize(stats.total_size || stats.TotalSize || 0)}</div></div></div></div>
    <div class="col-6 col-md-3"><div class="card shadow-sm h-100"><div class="card-body"><div class="text-muted small">最近下载时间</div><div class="h6 mb-0">${formatUnix(stats.last_film_birth || stats.LastFilmBirth, true)}</div></div></div></div>
    <div class="col-6 col-md-3"><div class="card shadow-sm h-100"><div class="card-body"><div class="text-muted small">最近更新</div><div class="h6 mb-0">${formatUnix(stats.last_updated_on || stats.LastUpdatedOn, true)}</div></div></div></div>
  </div>`;
}

function renderFilter(data) {
  const filter = data.MovieCardFilter || {};
  const input = (name, label, value, placeholder = "", type = "text") => `<div class="col-12 col-md-6 col-xl-2"><label class="form-label mb-1">${escapeHtml(label)}</label><input class="form-control" type="${attr(type)}" name="${attr(name)}" value="${attr(value || "")}" placeholder="${attr(placeholder)}"></div>`;
  return `<div id="movieCardFilterCollapse" class="collapse"><div class="card mb-3"><div class="card-body">
    <form class="row g-2 align-items-end" method="get" action="${attr(filter.Action || "/wdir")}">
      ${filter.PageSize ? `<input type="hidden" name="ps" value="${attr(filter.PageSize)}">` : ""}
      ${filter.OrderBy ? `<input type="hidden" name="od" value="${attr(filter.OrderBy)}">` : ""}
      ${input("cn", "演员", filter.CastNames, "演员名，支持多值")}
      ${input("gn", "类型", filter.GenreNames, "类型名，支持多值")}
      ${input("dn", "导演", filter.DirectorName, "导演名")}
      ${input("pn", "前缀", filter.PrefixName, "前缀")}
      ${input("mn", "片商", filter.MakerName, "片商")}
      ${input("ln", "标签", filter.LabelName, "标签")}
      ${input("wd", "关键词", filter.Word, "中文关键词")}
      ${input("rs", "上映开始", filter.ReleasingDateStart, "YYYY-MM-DD")}
      ${input("re", "上映结束", filter.ReleasingDateEnd, "YYYY-MM-DD")}
      ${input("mbs", "Media 下载开始", filter.MediaBirthTimeStart, "YYYY-MM-DD")}
      ${input("mbe", "Media 下载结束", filter.MediaBirthTimeEnd, "YYYY-MM-DD")}
      <div class="col-12 d-flex gap-2"><button class="btn btn-primary" type="submit">筛选</button><a class="btn btn-outline-secondary" href="${attr(filter.ClearHref || "/wdir")}">清空</a></div>
    </form>
  </div></div></div>`;
}

function renderWDirDetail(data) {
  const dir = data.Detail?.directory || data.Detail?.Directory || {};
  const id = dir.id || dir.Id;
  return `<div class="position-sticky pb-2 pt-2 mb-3 border-bottom" style="top:0; z-index:1030; background:rgba(255,255,255,.9); backdrop-filter: blur(6px);">
    <div class="container-fluid px-0">
      <div class="d-flex flex-column flex-md-row justify-content-between align-items-start align-items-md-center gap-2">
        <div class="flex-grow-1">${renderWDirBreadcrumbs(data.Detail)}<h3 class="mb-0">${escapeHtml(dir.name || dir.Name)}</h3></div>
        <div class="btn-group">
          <a class="btn btn-sm btn-outline-secondary ${data.Recursive ? "active" : ""}" href="/wdir?id=${attr(id)}&recursive=1${attr(data.Query || "")}">递归</a>
          <a class="btn btn-sm btn-outline-secondary ${!data.Recursive ? "active" : ""}" href="/wdir?id=${attr(id)}&recursive=0${attr(data.Query || "")}">直属</a>
        </div>
      </div>
      <div class="mt-2">${renderWDirStats(data.Detail)}</div>
      ${renderChildren(data.Children)}
    </div>
  </div>
  <div class="position-sticky pb-2 pt-2 mb-3 border-bottom" style="top:0; z-index:1029; background:rgba(255,255,255,.9); backdrop-filter: blur(6px);">
    <div class="container-fluid px-0">
      <div class="d-flex flex-wrap justify-content-start align-items-center mb-2">${renderSortBar(data.SortQuery)}</div>
      <div class="d-flex flex-column flex-md-row justify-content-between align-items-md-center gap-2">
        <div class="d-flex flex-wrap align-items-center gap-2">${renderOwnedFilter(data)}<button class="btn btn-sm btn-outline-secondary" type="button" data-bs-toggle="collapse" data-bs-target="#movieCardFilterCollapse">筛选</button></div>
        <div class="d-flex align-items-center ms-md-3">${renderPagination(data.PageInfo)}</div>
      </div>
    </div>
  </div>
  ${renderFilter(data)}
  <section aria-label="Movies">${renderMovieGrid(data.movies || data.Movies, "这个目录下暂时没有媒体。")}</section>`;
}

function renderPage(data) {
  if (data.redirect) {
    window.location.assign(data.redirect);
    return;
  }
  document.title = `${data.Title || "MovieAgg"} - Rudy-GC`;
  const root = document.getElementById("page-root");
  if (data.ViewType === "movie-agg-release") root.innerHTML = renderMovieAggRelease(data);
  else if (data.ViewType === "movie-release-bucket-list") root.innerHTML = renderReleaseBucketList(data);
  else if (data.ViewType === "w-media-agg-birth") root.innerHTML = renderWMediaAggBirth(data);
  else if (data.ViewType === "w-media-bucket-list") root.innerHTML = renderWMediaBucketList(data);
  else if (data.ViewType === "agg-triggers") {
    root.innerHTML = renderAggTriggers(data);
    initAggTriggers();
  } else if (data.ViewType === "wdir-list") root.innerHTML = renderWDirList(data);
  else if (data.ViewType === "wdir-detail") root.innerHTML = renderWDirDetail(data);
  else root.innerHTML = `<div class="alert alert-info">暂无数据。</div>`;
}

async function main() {
  const response = await fetch(`${apiPathForCurrentPage()}${window.location.search}`);
  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || `请求失败(${response.status})`);
  }
  renderPage(data);
}

main().catch((err) => {
  document.getElementById("page-root").innerHTML = `<div class="alert alert-danger">${escapeHtml(err.message)}</div>`;
});
