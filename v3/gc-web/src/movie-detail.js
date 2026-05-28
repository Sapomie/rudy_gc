const root = document.getElementById("movie-detail-root");

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

function get(obj, ...keys) {
  if (!obj) return undefined;
  for (const key of keys) {
    if (Object.prototype.hasOwnProperty.call(obj, key)) {
      return obj[key];
    }
  }
  return undefined;
}

function text(obj, ...keys) {
  const value = get(obj, ...keys);
  return value ?? "";
}

function num(obj, ...keys) {
  const value = Number(get(obj, ...keys) ?? 0);
  return Number.isFinite(value) ? value : 0;
}

function list(value) {
  return Array.isArray(value) ? value : [];
}

function movieNameFromPath() {
  const prefix = "/movie/";
  const path = window.location.pathname;
  if (!path.startsWith(prefix)) return "";
  return decodeURIComponent(path.slice(prefix.length));
}

function request(url, options) {
  return fetch(url, options).then(async (response) => {
    const data = await response.json().catch(() => ({}));
    if (!response.ok || (data && data.ok === false)) {
      throw new Error((data && data.error) || `请求失败(${response.status})`);
    }
    return data;
  });
}

function renderTags(items, makeHref, makeText) {
  return list(items).map((item) => {
    const label = makeText ? makeText(item) : item;
    return `<a class="tag" href="${attr(makeHref(item))}">${escapeHtml(label)}</a>`;
  }).join("");
}

function selectedAlbums(albums) {
  return list(albums).filter((album) => !!get(album, "Selected", "selected"));
}

function renderSelectedAlbumBadges(albums) {
  return selectedAlbums(albums).map((album) => {
    const name = text(album, "Name", "name");
    return `<a href="/movie-albums?album_name=${encodeURIComponent(name)}" class="badge bg-primary-subtle text-primary text-decoration-none">${escapeHtml(name)}</a>`;
  }).join("");
}

function renderSelectedAlbumTags(albums) {
  const selected = selectedAlbums(albums);
  if (selected.length === 0) {
    return `<span class="text-muted">-</span>`;
  }
  return selected.map((album) => {
    const name = text(album, "Name", "name");
    return `<a class="tag" href="/movie-albums?album_name=${encodeURIComponent(name)}">${escapeHtml(name)}</a>`;
  }).join("");
}

function renderPlayButton(movie) {
  const owned = num(movie, "owned_w_media", "OwnedWMedia");
  const videoUrl = text(movie, "video_url_w_media", "VideoUrlWMedia");
  if (owned === 5) {
    return `<a href="${attr(videoUrl)}" class="btn btn-sm btn-info mr-1">MPlay</a>`;
  }
  if (owned === 4) {
    return `<a href="${attr(videoUrl)}" class="btn btn-sm btn-success mr-1">MPlay</a>`;
  }
  if (owned === 6) {
    return `<span class="btn btn-sm btn-secondary mr-1 disabled" aria-disabled="true">MRemoved</span>`;
  }
  return "";
}

function renderMovieInfo(data) {
  const movie = data.movieType || data.MovieType || {};
  const albums = data.MovieAlbums || data.movieAlbums || [];
  const casts = list(text(movie, "cast", "Cast")).map((cast) => {
    const nameShow = text(cast, "NameShow", "DisplayName", "Name", "name");
    const href = text(cast, "Url", "url") || `/cards?cn=${encodeURIComponent(text(cast, "Name", "name"))}`;
    return `<a class="tag" href="${attr(href)}" title="${attr(nameShow)}">${escapeHtml(nameShow)}</a>`;
  }).join("");
  const genres = renderTags(text(movie, "genre", "Genre"), (genre) => `/cards?gn=${encodeURIComponent(genre)}`);
  const sc = list(data.sc || data.SC);
  const scLinks = sc.length > 0
    ? sc.map((row) => `<a class="tag" href="${attr(text(row, "Href"))}">${escapeHtml(text(row, "ScName"))}</a>`).join("")
    : `<span class="text-muted">-</span>`;
  const needDownload = num(movie, "need_download", "NeedDownload");
  const javID = text(movie, "jav_id", "JavId");

  return `
    <table class="table table-bordered dt-responsive nowrap" style="width:100%">
      <tr class="table-light">
        <th>识别码</th>
        <td>
          <a href="/cards?pn=${encodeURIComponent(text(movie, "prefix", "Prefix"))}" class="mr-2">${escapeHtml(text(movie, "name", "Name"))}</a>
          ${renderPlayButton(movie)}
          <small class="text-muted">(${escapeHtml(text(movie, "encode_name", "EncodeName"))})</small>
        </td>
      </tr>
      <tr class="table-light"><th>发行时间</th><td>${escapeHtml(text(movie, "releasing_date", "ReleasingDate"))}</td></tr>
      <tr class="table-light">
        <th>长度</th>
        <td>
          ${escapeHtml(text(movie, "length", "Length"))}
          <button id="btnDownloadCover" class="btn btn-sm btn-outline-info ml-2" data-jav="${attr(javID)}">
            <i class="mdi mdi-image-outline"></i>
            <span class="ms-1">下载封面</span>
          </button>
        </td>
      </tr>
      <tr class="table-light"><th>制作商</th><td><a class="tag" href="/cards?mn=${encodeURIComponent(text(movie, "maker", "Maker"))}">${escapeHtml(text(movie, "maker", "Maker"))}</a></td></tr>
      <tr class="table-light"><th>发行商</th><td><a class="tag" href="/cards?ln=${encodeURIComponent(text(movie, "label", "Label"))}">${escapeHtml(text(movie, "label", "Label"))}</a></td></tr>
      <tr class="table-light"><th>导演</th><td><a class="tag" href="/cards?dn=${encodeURIComponent(text(movie, "director", "Director"))}" title="${attr(text(movie, "director", "Director"))}">${escapeHtml(text(movie, "director", "Director"))}</a></td></tr>
      <tr class="table-light"><th>分类</th><td>${genres}</td></tr>
      <tr class="table-light">
        <th>演员</th>
        <td>
          <div class="d-flex flex-wrap align-items-center gap-2">
            ${casts}
            <button type="button" class="btn btn-sm btn-outline-primary" data-bs-toggle="modal" data-bs-target="#addMovieCastModal" data-jav="${attr(javID)}">添加 Cast</button>
          </div>
        </td>
      </tr>
      <tr class="table-light">
        <th>评分</th>
        <td>
          ${escapeHtml(text(movie, "score", "Score"))}
          <small class="text-muted">（${escapeHtml(text(movie, "viewers_number_watched", "ViewersNumberWatched"))}）</small>
          <small class="text-muted ml-2">${escapeHtml(text(movie, "file_creating_date", "UpdateDate"))}</small>
          <button id="btn-downloadlater"
                  class="btn btn-sm ${needDownload === 2 ? "btn-warning" : "btn-outline-warning"} ms-2"
                  data-name="${attr(javID)}"
                  data-needdownload="${attr(needDownload)}"
                  title="${needDownload === 2 ? "取消稍后下载" : "添加稍后下载"}">
            <i class="mdi mdi-download${needDownload === 2 ? "" : "-outline"}"></i>
            <span class="ms-1">${needDownload === 2 ? "已添加" : "稍后下载"}</span>
          </button>
        </td>
      </tr>
      <tr class="table-light">
        <th>链接</th>
        <td>
          <a href="${attr(text(movie, "bus_url", "BusUrl"))}" class="mr-2" target="_blank" rel="noopener">Bus</a>
          <a href="${attr(text(movie, "search_url", "SearchUrl"))}" class="mr-2" target="_blank" rel="noopener">Suk</a>
          <a href="${attr(text(movie, "jav_url", "JavUrl"))}" target="_blank" rel="noopener">Lib</a>
        </td>
      </tr>
      <tr class="table-light"><th>RankInfo</th><td>${escapeHtml(text(movie, "first_ranking_day", "FirstRankingDay"))}</td></tr>
      <tr class="table-light"><th>Sc</th><td><div class="d-flex flex-wrap gap-2">${scLinks}</div></td></tr>
      <tr class="table-light"><th>电影相册</th><td>${renderSelectedAlbumTags(albums)}</td></tr>
    </table>

    <div class="modal fade" id="addMovieCastModal" tabindex="-1" aria-labelledby="addMovieCastModalLabel" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <form id="formAddMovieCast" data-jav="${attr(javID)}">
            <div class="modal-header">
              <h5 class="modal-title" id="addMovieCastModalLabel">添加已有演员</h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
            </div>
            <div class="modal-body">
              <div class="mb-3">
                <label for="movieCastName" class="form-label">演员名</label>
                <input type="text" class="form-control" id="movieCastName" placeholder="输入已存在演员名">
                <div class="form-text">只支持绑定已存在演员；不存在时会失败。</div>
              </div>
              <div id="movieCastMsg" class="small" style="display:none;"></div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
              <button type="submit" class="btn btn-primary">添加 Cast</button>
            </div>
          </form>
        </div>
      </div>
    </div>`;
}

function renderRankInfo(data) {
  return `
    <table id="moviedetail_rankinfo" class="table table-bordered dt-responsive nowrap" style="border-collapse: collapse; border-spacing: 0; width: 100%;">
      <thead><tr class="table-light"><th>Date</th><th>Pts</th></tr></thead>
      <tbody>
        ${list(data.rankInfos || data.RankInfos).map((row) => {
          const date = text(row, "Date");
          return `<tr><td><a href="/moviecarddayrank?date=${encodeURIComponent(date)}">${escapeHtml(date)}</a></td><td><span>${escapeHtml(text(row, "Rank"))}</span></td></tr>`;
        }).join("")}
      </tbody>
    </table>`;
}

function renderFilmInfo(data) {
  const movie = data.movieType || data.MovieType || {};
  const media = data.media || data.Media || data.MediaInfo;
  if (!media || num(movie, "owned_w_media", "OwnedWMedia") === 6) {
    return "";
  }
  const marked = !!(data.MovieDeleteAlbumMarked || data.movieDeleteAlbumMarked);
  return `
    <div class="row mt-3">
      <div class="col-12">
        <div class="card shadow-sm">
          <div class="card-body">
            <h5 class="card-title mb-3">视频信息</h5>
            <table class="table table-bordered dt-responsive nowrap mb-0" style="width:100%">
              <thead>
                <tr class="table-light">
                  <th>Source</th><th>TorrentHash</th><th>BirthTime</th><th>Size</th><th>Height</th><th>BitRate</th><th>时长(分钟)</th><th>Frame</th><th>SelfMake</th><th>Dir</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>WMedia</td>
                  <td class="text-monospace">${text(media, "SourceTorrentHash") ? escapeHtml(text(media, "SourceTorrentHash")) : `<span class="text-muted">-</span>`}</td>
                  <td>${escapeHtml(text(media, "BirthTime"))}</td>
                  <td>
                    <div class="d-flex flex-column gap-1">
                      <a href="${attr(text(media, "FilePath"))}">${Number(text(media, "Size") || 0).toFixed(2)}</a>
                      <button type="button" class="btn btn-sm ${marked ? "btn-danger" : "btn-outline-danger"} js-move-wmedia-removed" data-jav="${attr(text(movie, "jav_id", "JavId"))}">
                        ${marked ? "已在待删除" : "加入待删除"}
                      </button>
                    </div>
                  </td>
                  <td>${escapeHtml(text(media, "Height"))}</td>
                  <td>${Number(text(media, "BitRate") || 0).toFixed(0)}</td>
                  <td>${Number(text(media, "DurationMinutes") || 0).toFixed(2)}</td>
                  <td>${escapeHtml(text(media, "Frame"))}</td>
                  <td>${escapeHtml(text(media, "SelfMake"))}</td>
                  <td><a href="#" class="open-in-finder" data-path="${attr(text(media, "Directory"))}" title="在 Finder 打开">${escapeHtml(text(media, "Directory"))}</a></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>`;
}

function renderFavoriteButtons(row, sourceType) {
  const rowID = text(row, "RowID");
  const inDownload = !!get(row, "IsInDownload");
  const inPending = !!get(row, "IsInPending");
  return `
    <button type="button" class="btn btn-sm ${inDownload ? "btn-warning" : "btn-outline-warning"} js-add-favorite"
            data-source-type="${attr(sourceType)}" data-source-row-id="${attr(rowID)}" data-album-name="下载中" data-favorited="${inDownload ? "1" : "0"}"
            title="${inDownload ? "移出下载中" : "下载中"}">${inDownload ? "已在下载中" : "下载中"}</button>
    <button type="button" class="btn btn-sm ms-1 ${inPending ? "btn-info" : "btn-outline-info"} js-add-favorite"
            data-source-type="${attr(sourceType)}" data-source-row-id="${attr(rowID)}" data-album-name="待下载" data-favorited="${inPending ? "1" : "0"}"
            title="${inPending ? "移出待下载" : "待下载"}">${inPending ? "已在待下载" : "待下载"}</button>`;
}

function hashClass(mediaHash, rowHash) {
  return mediaHash && mediaHash === rowHash ? " fetch-site-hash-match" : "";
}

function renderJavbusTable(rows, mediaHash) {
  if (rows.length === 0) return `<div class="text-muted">暂无 JavBus 磁力数据</div>`;
  return `
    <div class="table-responsive">
      <table class="table table-sm table-hover align-middle mb-0" data-sort-table="javbus-magnets">
        <thead>
          <tr>
            <th>名称</th>
            <th><button type="button" class="btn btn-link btn-sm p-0 text-decoration-none js-sort-table" data-table-id="javbus-magnets" data-sort-key="size">大小</button></th>
            <th><button type="button" class="btn btn-link btn-sm p-0 text-decoration-none js-sort-table" data-table-id="javbus-magnets" data-sort-key="date">日期</button></th>
            <th>标签</th><th>Info Hash</th><th>相册</th>
          </tr>
        </thead>
        <tbody>
          ${rows.map((row) => {
            const infoHash = text(row, "InfoHash");
            return `<tr${get(row, "HasMatchedHash") ? ` class="fetch-site-row-match"` : ""}>
              <td>${escapeHtml(text(row, "MagnetName"))}</td>
              <td data-sort-key="size">${escapeHtml(text(row, "SizeText"))}</td>
              <td class="text-nowrap" data-sort-key="date">${escapeHtml(text(row, "ShareDate"))}</td>
              <td class="text-nowrap">${get(row, "HasSubtitle") ? `<span class="fetch-site-tag-subtitle">字幕</span>` : `<span class="text-muted">-</span>`}</td>
              <td class="text-monospace"><span class="${hashClass(mediaHash, infoHash).trim()}">${escapeHtml(infoHash)}</span></td>
              <td class="text-nowrap">${renderFavoriteButtons(row, "javbus_magnet")}</td>
            </tr>`;
          }).join("")}
        </tbody>
      </table>
    </div>`;
}

function renderSukebeiTable(rows, mediaHash) {
  if (rows.length === 0) return `<div class="text-muted">暂无 Sukebei 种子数据</div>`;
  return `
    <div class="fetch-site-detail-sukebei-wrap">
      <table class="table table-sm table-hover align-middle mb-0 fetch-site-detail-sukebei-table" data-sort-table="sukebei-torrents">
        <colgroup>
          <col style="width: 501px;"><col style="width: 110px;"><col style="width: 130px;"><col style="width: 90px;"><col style="width: 340px;"><col style="width: 90px;">
        </colgroup>
        <thead>
          <tr>
            <th class="fetch-site-col-title">标题</th>
            <th class="fetch-site-col-compact"><button type="button" class="btn btn-link btn-sm p-0 text-decoration-none js-sort-table" data-table-id="sukebei-torrents" data-sort-key="size">大小</button></th>
            <th class="fetch-site-col-compact"><button type="button" class="btn btn-link btn-sm p-0 text-decoration-none js-sort-table" data-table-id="sukebei-torrents" data-sort-key="date">日期</button></th>
            <th class="fetch-site-col-compact"><button type="button" class="btn btn-link btn-sm p-0 text-decoration-none js-sort-table" data-table-id="sukebei-torrents" data-sort-key="downloads">下载</button></th>
            <th class="fetch-site-col-hash">Info Hash</th><th class="fetch-site-col-compact">相册</th>
          </tr>
        </thead>
        <tbody>
          ${rows.map((row) => {
            const infoHash = text(row, "InfoHash");
            const title = text(row, "TorrentTitle");
            const viewURL = text(row, "ViewURL");
            return `<tr${get(row, "HasMatchedHash") ? ` class="fetch-site-row-match"` : ""}>
              <td class="fetch-site-col-title"><div class="fetch-site-title-content">${viewURL ? `<a href="${attr(viewURL)}" target="_blank" rel="noreferrer">${escapeHtml(title)}</a>` : escapeHtml(title)}</div></td>
              <td class="fetch-site-col-compact text-nowrap" data-sort-key="size">${escapeHtml(text(row, "SizeText"))}</td>
              <td class="fetch-site-col-compact text-nowrap" data-sort-key="date">${escapeHtml(text(row, "PublishTime"))}</td>
              <td class="fetch-site-col-compact text-nowrap" data-sort-key="downloads">${escapeHtml(text(row, "Completed"))}</td>
              <td class="fetch-site-col-hash text-monospace"><span class="fetch-site-hash-text${hashClass(mediaHash, infoHash)}" title="${attr(infoHash)}">${escapeHtml(infoHash)}</span></td>
              <td class="fetch-site-col-compact text-nowrap">${renderFavoriteButtons(row, "sukebei_torrent")}</td>
            </tr>`;
          }).join("")}
        </tbody>
      </table>
    </div>`;
}

function renderSehuatangTable(rows, mediaHash) {
  if (rows.length === 0) return `<div class="text-muted">暂无 Sehuatang 磁力数据</div>`;
  return `
    <div class="table-responsive">
      <table class="table table-sm table-hover align-middle mb-0" data-sort-table="sehuatang-magnets">
        <thead>
          <tr>
            <th>帖子标题</th>
            <th><button type="button" class="btn btn-link btn-sm p-0 text-decoration-none js-sort-table" data-table-id="sehuatang-magnets" data-sort-key="date">发帖时间</button></th>
            <th>Info Hash</th><th>相册</th>
          </tr>
        </thead>
        <tbody>
          ${rows.map((row) => {
            const infoHash = text(row, "InfoHash");
            const title = text(row, "ThreadTitle");
            const url = text(row, "ThreadURL");
            return `<tr${get(row, "HasMatchedHash") ? ` class="fetch-site-row-match"` : ""}>
              <td>${url ? `<a href="${attr(url)}" target="_blank" rel="noopener noreferrer">${escapeHtml(title)}</a>` : escapeHtml(title)}</td>
              <td class="text-nowrap" data-sort-key="date">${escapeHtml(text(row, "PostTime"))}</td>
              <td class="text-monospace"><span class="${hashClass(mediaHash, infoHash).trim()}">${escapeHtml(infoHash)}</span></td>
              <td class="text-nowrap">${renderFavoriteButtons(row, "sehuatang_magnet")}</td>
            </tr>`;
          }).join("")}
        </tbody>
      </table>
    </div>`;
}

function renderResources(data) {
  const media = data.media || data.Media || data.MediaInfo || {};
  const mediaHash = text(media, "SourceTorrentHash");
  const javbus = list(data.javbusMagnets);
  const sukebei = list(data.sukebeiRows);
  const sehuatang = list(data.sehuatangRows);
  return `
    <div id="fetchSiteListMsg" class="alert d-none mb-3" role="alert"></div>
    <div class="row">
      <div class="col-12 mb-3">
        <div class="card shadow-sm">
          <div class="card-body">
            <h5 class="card-title mb-3">资源列表</h5>
            <ul class="nav nav-tabs nav-tabs-custom mb-3" role="tablist">
              <li class="nav-item" role="presentation">
                <button class="nav-link active" id="tab-javbus-magnets" data-bs-toggle="tab" data-bs-target="#pane-javbus-magnets" type="button" role="tab" aria-controls="pane-javbus-magnets" aria-selected="true">
                  JavBus 磁力 ${javbus.length > 0 ? `<span class="badge bg-danger-subtle text-danger fs-6 px-2 py-1 ms-1">(${javbus.length})</span>` : ""}
                </button>
              </li>
              <li class="nav-item" role="presentation">
                <button class="nav-link" id="tab-sukebei-torrents" data-bs-toggle="tab" data-bs-target="#pane-sukebei-torrents" type="button" role="tab" aria-controls="pane-sukebei-torrents" aria-selected="false">
                  Sukebei 种子 ${sukebei.length > 0 ? `<span class="badge bg-danger-subtle text-danger fs-6 px-2 py-1 ms-1">(${sukebei.length})</span>` : ""}
                </button>
              </li>
              <li class="nav-item" role="presentation">
                <button class="nav-link" id="tab-sehuatang-magnets" data-bs-toggle="tab" data-bs-target="#pane-sehuatang-magnets" type="button" role="tab" aria-controls="pane-sehuatang-magnets" aria-selected="false">
                  Sehuatang 磁力 ${sehuatang.length > 0 ? `<span class="badge bg-danger-subtle text-danger fs-6 px-2 py-1 ms-1">(${sehuatang.length})</span>` : ""}
                </button>
              </li>
            </ul>
            <div class="tab-content">
              <div class="tab-pane fade show active" id="pane-javbus-magnets" role="tabpanel" aria-labelledby="tab-javbus-magnets">${renderJavbusTable(javbus, mediaHash)}</div>
              <div class="tab-pane fade" id="pane-sukebei-torrents" role="tabpanel" aria-labelledby="tab-sukebei-torrents">${renderSukebeiTable(sukebei, mediaHash)}</div>
              <div class="tab-pane fade" id="pane-sehuatang-magnets" role="tabpanel" aria-labelledby="tab-sehuatang-magnets">${renderSehuatangTable(sehuatang, mediaHash)}</div>
            </div>
          </div>
        </div>
      </div>
    </div>`;
}

function renderFetchStatus(status, title, triggerClass) {
  return `
    <div class="col-12 col-xl-6 mb-3">
      <div class="card shadow-sm h-100">
        <div class="card-body">
          <div class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
            <h5 class="card-title mb-0">${escapeHtml(title)}</h5>
            <button type="button" class="btn btn-sm btn-outline-primary ${triggerClass}">触发 ${title.replace(" 抓取状态", "")}</button>
          </div>
          ${status ? `
            <div class="table-responsive fetch-site-detail-sukebei-wrap">
              <table class="table table-sm table-bordered mb-0">
                <tbody>
                  <tr><th>状态</th><td>${escapeHtml(text(status, "FetchStatusText"))}</td></tr>
                  <tr><th>重试次数</th><td>${escapeHtml(text(status, "TryCount"))}</td></tr>
                  <tr><th>最后抓取</th><td>${escapeHtml(text(status, "LastFetchTime"))}</td></tr>
                  <tr><th>最后抓取距离现在</th><td>${escapeHtml(text(status, "LastFetchAgo"))}</td></tr>
                  <tr><th>Hash 数</th><td>${escapeHtml(text(status, "TorrentHashCount"))}</td></tr>
                  <tr><th>最新日期</th><td>${escapeHtml(text(status, "LatestPublishTime"))}</td></tr>
                  <tr><th>来源地址</th><td>${text(status, "SourceURL") ? `<a href="${attr(text(status, "SourceURL"))}" target="_blank" rel="noreferrer">${escapeHtml(text(status, "SourceURL"))}</a>` : "-"}</td></tr>
                  <tr><th>错误</th><td>${text(status, "LastError") ? escapeHtml(text(status, "LastError")) : "-"}</td></tr>
                </tbody>
              </table>
            </div>` : `<div class="text-muted">暂无 ${escapeHtml(title.replace(" 抓取状态", ""))} 抓取状态</div>`}
        </div>
      </div>
    </div>`;
}

function renderFetchStatuses(data) {
  return `
    <div class="row mt-3">
      <div class="col-12 mb-2 d-flex justify-content-end">
        <button type="button" class="btn btn-sm btn-primary js-trigger-both">同时触发 JavBus + Sukebei</button>
      </div>
      ${renderFetchStatus(data.javbusFetch, "JavBus 抓取状态", "js-trigger-javbus")}
      ${renderFetchStatus(data.sukebeiFetch, "Sukebei 抓取状态", "js-trigger-sukebei")}
    </div>`;
}

function renderMovieAlbumModal(albums) {
  return `
    <div class="modal fade" id="movieAlbumModal" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">加入电影相册</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <div id="movieAlbumMsg" class="alert d-none mb-3" role="alert"></div>
            <div class="mb-3">
              <label class="form-label mb-1" for="movieAlbumSelect">选择相册</label>
              <select class="form-select" id="movieAlbumSelect">
                ${list(albums).map((album) => {
                  const name = text(album, "Name", "name");
                  return `<option value="${attr(name)}">${escapeHtml(name)}${get(album, "Selected", "selected") ? "（已在相册）" : ""}</option>`;
                }).join("")}
              </select>
            </div>
            <div>
              <label class="form-label mb-1" for="movieAlbumNewName">新建相册</label>
              <div class="d-flex gap-2">
                <input type="text" class="form-control" id="movieAlbumNewName" maxlength="128" placeholder="输入新相册名称后创建">
                <button type="button" class="btn btn-outline-secondary" id="btnCreateMovieAlbumInline">创建</button>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
            <button type="button" class="btn btn-primary" id="btnAddMovieToAlbum">加入</button>
          </div>
        </div>
      </div>
    </div>`;
}

function renderModals(data) {
  return `
    <div class="modal fade" id="moveWMediaRemovedModal" tabindex="-1" aria-labelledby="moveWMediaRemovedModalLabel" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="moveWMediaRemovedModalLabel">加入待删除</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            将当前电影加入默认电影相册“待删除”。真正删除会在电影相册页统一执行。
            <div id="moveWMediaRemovedMsg" class="small mt-2" style="display:none;"></div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-light" data-bs-dismiss="modal">取消</button>
            <button type="button" class="btn btn-danger" id="btnMoveWMediaRemovedDo">加入待删除</button>
          </div>
        </div>
      </div>
    </div>
    ${renderMovieAlbumModal(data.MovieAlbums || data.movieAlbums || [])}`;
}

function renderPage(data) {
  const movie = data.movieType || data.MovieType || {};
  const albums = data.MovieAlbums || data.movieAlbums || [];
  const title = text(movie, "title", "Title") || text(movie, "name", "Name");
  const javID = text(movie, "jav_id", "JavId");
  const movieName = text(movie, "name", "Name");
  document.title = title ? `${title} - Rudy-GC` : "MovieDetail - Rudy-GC";
  root.innerHTML = `
    <div id="fetchSiteListPage" data-movie-jav-id="${attr(javID)}" data-movie-name="${attr(movieName)}">
      <div class="row">
        <div class="col-lg-7 mb-3">
          <div class="card shadow-sm">
            <div class="card-body pb-2">
              <h4 class="mb-1" title="${attr(title)}">${escapeHtml(title)}</h4>
              <div class="mt-2 d-flex flex-wrap align-items-center gap-2">
                <button type="button" class="btn btn-sm btn-outline-primary" id="btnOpenMovieAlbumModal">加入电影相册</button>
                ${renderSelectedAlbumBadges(albums)}
              </div>
            </div>
            <div class="thumb-2x2">
              <img src="${attr(text(movie, "jacket_img", "JacketImg"))}" alt="Card image cap" class="card-img-bottom img-fluid" style="object-fit: contain; width: 100%; height: auto;">
              <div class="badge-stack">
                ${num(movie, "come_times", "ComeTimes") > 0 ? `<span class="badge-come">❤ ${escapeHtml(text(movie, "come_times", "ComeTimes"))}</span>` : ""}
                ${num(movie, "sc_times", "ScTimes") > 0 ? `<span class="badge-times">× ${escapeHtml(text(movie, "sc_times", "ScTimes"))}</span>` : ""}
              </div>
              ${String(text(movie, "score", "Score")) !== "" ? `<span class="badge-score">★ ${escapeHtml(text(movie, "score", "Score"))}</span>` : ""}
            </div>
          </div>
        </div>
        <div class="col-lg-5">
          <div class="card shadow-sm">
            <div class="card-body">
              <ul class="nav nav-tabs nav-tabs-custom nav-justified" role="tablist">
                <li class="nav-item"><a class="nav-link active" data-bs-toggle="tab" href="#tab_primary" role="tab"><span class="d-block d-sm-none"><i class="far fa-envelope"></i></span><span class="d-none d-sm-block">Info</span></a></li>
                <li class="nav-item"><a class="nav-link" data-bs-toggle="tab" href="#tab_average" role="tab"><span class="d-block d-sm-none"><i class="fas fa-home"></i></span><span class="d-none d-sm-block">Rank</span></a></li>
              </ul>
              <div class="tab-content">
                <div class="tab-pane p-3 active" id="tab_primary" role="tabpanel">${renderMovieInfo(data)}</div>
                <div class="tab-pane p-3" id="tab_average" role="tabpanel">${renderRankInfo(data)}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
      ${renderFilmInfo(data)}
      ${renderResources(data)}
      ${renderFetchStatuses(data)}
    </div>
    ${renderModals(data)}`;
  initRankInfoTable();
}

function initRankInfoTable() {
  const jq = window.jQuery || window.$;
  if (!jq || !jq.fn || !jq.fn.DataTable) return;
  jq("#moviedetail_rankinfo").DataTable({
    aLengthMenu: [10, 50, 100],
    aaSorting: [[1, "asc"]]
  });
}

function showMsg(textValue, ok) {
  const msgEl = document.getElementById("fetchSiteListMsg");
  if (!msgEl) return;
  msgEl.textContent = String(textValue || "");
  msgEl.className = `alert small py-2 px-3 mb-3 ${ok ? "alert-success" : "alert-danger"}`;
}

function setButtonState(button, loading) {
  if (button) button.disabled = loading;
}

function setTriggerButtonsLoading(loading) {
  document.querySelectorAll(".js-trigger-javbus, .js-trigger-sukebei, .js-trigger-both").forEach((button) => setButtonState(button, loading));
}

function modalOf(id) {
  const el = document.getElementById(id);
  if (!el || !window.bootstrap || !window.bootstrap.Modal) return null;
  return window.bootstrap.Modal.getOrCreateInstance(el);
}

function parseSizeToBytes(value) {
  const raw = String(value || "").trim().toUpperCase();
  const match = raw.match(/([0-9]+(?:\.[0-9]+)?)\s*(B|KB|KIB|MB|MIB|GB|GIB|TB|TIB)/);
  if (!match) return 0;
  const size = parseFloat(match[1]);
  const units = {B: 1, KB: 1024, KIB: 1024, MB: 1024 ** 2, MIB: 1024 ** 2, GB: 1024 ** 3, GIB: 1024 ** 3, TB: 1024 ** 4, TIB: 1024 ** 4};
  return size * (units[match[2]] || 1);
}

function parseDateValue(value) {
  const ts = Date.parse(String(value || "").trim().replace(/\s+/g, "T"));
  return Number.isNaN(ts) ? 0 : ts;
}

function parseNumberValue(value) {
  const parsed = parseInt(String(value || "").replace(/[^\d-]/g, ""), 10);
  return Number.isNaN(parsed) ? 0 : parsed;
}

function extractSortValue(row, sortKey) {
  const cell = row.querySelector(`[data-sort-key="${sortKey}"]`);
  const value = cell ? cell.textContent : "";
  if (sortKey === "size") return parseSizeToBytes(value);
  if (sortKey === "date") return parseDateValue(value);
  if (sortKey === "downloads") return parseNumberValue(value);
  return String(value || "").trim().toLowerCase();
}

function sortTable(tableID, sortKey, button) {
  const table = document.querySelector(`[data-sort-table="${tableID}"]`);
  const tbody = table ? table.querySelector("tbody") : null;
  if (!tbody) return;
  const rows = Array.from(tbody.querySelectorAll("tr"));
  const currentKey = table.getAttribute("data-sort-current-key") || "";
  const currentOrder = table.getAttribute("data-sort-current-order") || "desc";
  const nextOrder = currentKey === sortKey && currentOrder === "desc" ? "asc" : "desc";

  rows.sort((left, right) => {
    const leftValue = extractSortValue(left, sortKey);
    const rightValue = extractSortValue(right, sortKey);
    if (leftValue === rightValue) return 0;
    if (nextOrder === "asc") return leftValue > rightValue ? 1 : -1;
    return leftValue < rightValue ? 1 : -1;
  });
  rows.forEach((row) => tbody.appendChild(row));
  table.setAttribute("data-sort-current-key", sortKey);
  table.setAttribute("data-sort-current-order", nextOrder);
  document.querySelectorAll(`.js-sort-table[data-table-id="${tableID}"]`).forEach((item) => {
    item.textContent = item.textContent.replace(/\s*[↑↓]$/, "");
  });
  if (button) {
    button.textContent = button.textContent.replace(/\s*[↑↓]$/, "") + (nextOrder === "asc" ? " ↑" : " ↓");
  }
}

function applyFavoriteButtonState(button, favorited, albumName) {
  const targetAlbumName = String(albumName || button.getAttribute("data-album-name") || "").trim() || "下载中";
  const pendingAlbum = targetAlbumName === "待下载";
  button.setAttribute("data-favorited", favorited ? "1" : "0");
  button.setAttribute("data-album-name", targetAlbumName);
  button.classList.remove("btn-warning", "btn-outline-warning", "btn-info", "btn-outline-info");
  button.classList.add(pendingAlbum ? (favorited ? "btn-info" : "btn-outline-info") : (favorited ? "btn-warning" : "btn-outline-warning"));
  button.textContent = favorited ? `已在${targetAlbumName}` : targetAlbumName;
  button.title = favorited ? `移出${targetAlbumName}` : targetAlbumName;
}

function readMovieTarget() {
  const page = document.getElementById("fetchSiteListPage");
  return {
    movieJavID: String(page ? page.getAttribute("data-movie-jav-id") : "").trim(),
    movieName: String(page ? page.getAttribute("data-movie-name") : "").trim()
  };
}

function switchAlbumFavorite(button) {
  const target = readMovieTarget();
  const sourceType = String(button.getAttribute("data-source-type") || "").trim();
  const sourceRowID = parseInt(String(button.getAttribute("data-source-row-id") || "0"), 10);
  const currentFavorited = String(button.getAttribute("data-favorited") || "").trim() === "1";
  const albumName = String(button.getAttribute("data-album-name") || "").trim() || "下载中";
  if (!target.movieJavID || !sourceType || !Number.isInteger(sourceRowID) || sourceRowID <= 0) {
    showMsg(`缺少${albumName}参数，无法执行`, false);
    return;
  }

  setButtonState(button, true);
  request(`/api/movie/${encodeURIComponent(target.movieJavID)}/album-item`, {
    method: currentFavorited ? "DELETE" : "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({
      movie_jav_id: target.movieJavID,
      source_type: sourceType,
      source_row_id: sourceRowID,
      album_name: albumName
    })
  }).then((data) => {
    const favorited = typeof data.favorited === "boolean" ? data.favorited : !currentFavorited;
    const responseAlbumName = data.album_name ? String(data.album_name).trim() : albumName;
    applyFavoriteButtonState(button, favorited, responseAlbumName);
    showMsg(data.message || "操作成功", true);
  }).catch((error) => {
    showMsg(error && error.message ? error.message : `${albumName}操作失败`, false);
  }).finally(() => {
    setButtonState(button, false);
  });
}

function startTask(taskType) {
  const target = readMovieTarget();
  if (!target.movieJavID || !target.movieName) {
    showMsg("缺少影片参数，无法触发任务", false);
    return;
  }
  setTriggerButtonsLoading(true);
  request("/api/crawler/jobs/start", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({task_type: taskType, movie_jav_id: target.movieJavID, movie_name: target.movieName})
  }).then(() => {
    const site = taskType === "spider_fetch_javbus_resources" ? "JavBus" : "Sukebei";
    showMsg(`已开启${site}抓取：${target.movieName}`, true);
  }).catch((error) => {
    showMsg(error && error.message ? error.message : "启动失败", false);
  }).finally(() => {
    setTriggerButtonsLoading(false);
  });
}

function startBothTasks() {
  const target = readMovieTarget();
  if (!target.movieJavID || !target.movieName) {
    showMsg("缺少影片参数，无法触发任务", false);
    return;
  }
  setTriggerButtonsLoading(true);
  Promise.allSettled([
    request("/api/crawler/jobs/start", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({task_type: "spider_fetch_javbus_resources", movie_jav_id: target.movieJavID, movie_name: target.movieName})
    }),
    request("/api/crawler/jobs/start", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({task_type: "spider_fetch_sukebei_resources", movie_jav_id: target.movieJavID, movie_name: target.movieName})
    })
  ]).then((results) => {
    const successSites = [];
    const failedSites = [];
    results.forEach((result, index) => {
      const site = index === 0 ? "JavBus" : "Sukebei";
      if (result.status === "fulfilled") {
        successSites.push(site);
      } else {
        failedSites.push(`${site}：${result.reason && result.reason.message ? result.reason.message : "启动失败"}`);
      }
    });
    if (successSites.length === 2) {
      showMsg(`已同时开启 JavBus + Sukebei 抓取：${target.movieName}`, true);
    } else if (successSites.length === 1) {
      showMsg(`部分成功：已开启${successSites[0]}；${failedSites.join("；")}`, false);
    } else {
      showMsg(`同时触发失败：${failedSites.join("；")}`, false);
    }
  }).finally(() => {
    setTriggerButtonsLoading(false);
  });
}

document.addEventListener("click", async (event) => {
  const sortButton = event.target.closest(".js-sort-table");
  if (sortButton) {
    sortTable(String(sortButton.getAttribute("data-table-id") || "").trim(), String(sortButton.getAttribute("data-sort-key") || "").trim(), sortButton);
    return;
  }

  const favoriteButton = event.target.closest(".js-add-favorite");
  if (favoriteButton) {
    switchAlbumFavorite(favoriteButton);
    return;
  }

  const btnLater = event.target.closest("#btn-downloadlater");
  if (btnLater) {
    const javID = btnLater.dataset.name;
    const current = btnLater.dataset.needdownload === "2";
    try {
      const data = await request(`/api/movie/${encodeURIComponent(javID)}/downloadlater`, {method: current ? "DELETE" : "POST"});
      const isAdded = data.needDownload === 2;
      btnLater.dataset.needdownload = isAdded ? "2" : "1";
      btnLater.classList.toggle("btn-warning", isAdded);
      btnLater.classList.toggle("btn-outline-warning", !isAdded);
      const icon = btnLater.querySelector("i");
      if (icon) {
        icon.classList.toggle("mdi-download-outline", !isAdded);
        icon.classList.toggle("mdi-download", isAdded);
      }
      const span = btnLater.querySelector("span");
      if (span) span.textContent = isAdded ? "已添加" : "稍后下载";
      btnLater.title = isAdded ? "取消稍后下载" : "添加稍后下载";
    } catch (error) {
      alert(error.message || "网络错误，请稍后重试");
    }
    return;
  }

  const btnCover = event.target.closest("#btnDownloadCover");
  if (btnCover) {
    const javID = String(btnCover.dataset.jav || "").trim();
    if (!javID) return;
    btnCover.disabled = true;
    btnCover.classList.add("disabled");
    try {
      await request(`/api/movie/${encodeURIComponent(javID)}/download-cover`, {method: "POST"});
      alert("封面下载完成");
    } catch (error) {
      alert(`下载失败：${error.message || "网络错误"}`);
    } finally {
      btnCover.disabled = false;
      btnCover.classList.remove("disabled");
    }
    return;
  }

  const moveButton = event.target.closest(".js-move-wmedia-removed");
  if (moveButton) {
    const modalEl = document.getElementById("moveWMediaRemovedModal");
    if (!modalEl) return;
    modalEl.dataset.jav = String(moveButton.dataset.jav || "").trim();
    const msgEl = modalEl.querySelector("#moveWMediaRemovedMsg");
    if (msgEl) {
      msgEl.style.display = "none";
      msgEl.textContent = "";
      msgEl.className = "small mt-2";
    }
    modalOf("moveWMediaRemovedModal")?.show();
    return;
  }

  const btnDoMove = event.target.closest("#btnMoveWMediaRemovedDo");
  if (btnDoMove) {
    const modalEl = document.getElementById("moveWMediaRemovedModal");
    const javID = modalEl ? String(modalEl.dataset.jav || "").trim() : "";
    if (!javID) return;
    btnDoMove.disabled = true;
    try {
      const data = await request(`/api/movie/${encodeURIComponent(javID)}/move-wmedia-removed`, {method: "POST"});
      const msgEl = modalEl.querySelector("#moveWMediaRemovedMsg");
      if (msgEl) {
        msgEl.textContent = data.message || "已加入待删除相册";
        msgEl.className = "small mt-2 text-success";
        msgEl.style.display = "block";
      }
      window.setTimeout(() => window.location.reload(), 400);
    } catch (error) {
      const msgEl = modalEl ? modalEl.querySelector("#moveWMediaRemovedMsg") : null;
      if (msgEl) {
        msgEl.textContent = error.message || "网络错误，请稍后重试";
        msgEl.className = "small mt-2 text-danger";
        msgEl.style.display = "block";
      }
    } finally {
      btnDoMove.disabled = false;
    }
    return;
  }

  if (event.target.closest(".js-trigger-javbus")) {
    startTask("spider_fetch_javbus_resources");
    return;
  }
  if (event.target.closest(".js-trigger-sukebei")) {
    startTask("spider_fetch_sukebei_resources");
    return;
  }
  if (event.target.closest(".js-trigger-both")) {
    startBothTasks();
    return;
  }

  if (event.target.closest("#btnOpenMovieAlbumModal")) {
    const msgEl = document.getElementById("movieAlbumMsg");
    if (msgEl) {
      msgEl.className = "alert d-none mb-3";
      msgEl.textContent = "";
    }
    modalOf("movieAlbumModal")?.show();
    return;
  }

  const createAlbumButton = event.target.closest("#btnCreateMovieAlbumInline");
  if (createAlbumButton) {
    const newNameEl = document.getElementById("movieAlbumNewName");
    const selectEl = document.getElementById("movieAlbumSelect");
    const msgEl = document.getElementById("movieAlbumMsg");
    const name = newNameEl ? String(newNameEl.value || "").trim() : "";
    if (!name) {
      if (msgEl) {
        msgEl.textContent = "请输入电影相册名称";
        msgEl.className = "alert mb-3 alert-danger";
      }
      return;
    }
    createAlbumButton.disabled = true;
    try {
      const data = await request("/api/movie-albums", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({name})
      });
      if (selectEl) {
        const exists = Array.from(selectEl.options).some((option) => String(option.value || "").trim() === (data.album_name || name));
        if (!exists) {
          const option = document.createElement("option");
          option.value = data.album_name || name;
          option.textContent = data.album_name || name;
          selectEl.appendChild(option);
        }
        selectEl.value = data.album_name || name;
      }
      if (newNameEl) newNameEl.value = "";
      if (msgEl) {
        msgEl.textContent = data.message || "电影相册创建成功";
        msgEl.className = "alert mb-3 alert-success";
      }
    } catch (error) {
      if (msgEl) {
        msgEl.textContent = error.message || "电影相册创建失败";
        msgEl.className = "alert mb-3 alert-danger";
      }
    } finally {
      createAlbumButton.disabled = false;
    }
    return;
  }

  const addAlbumButton = event.target.closest("#btnAddMovieToAlbum");
  if (addAlbumButton) {
    const selectEl = document.getElementById("movieAlbumSelect");
    const msgEl = document.getElementById("movieAlbumMsg");
    const target = readMovieTarget();
    const albumName = selectEl ? String(selectEl.value || "").trim() : "";
    if (!target.movieJavID || !albumName) {
      if (msgEl) {
        msgEl.textContent = !target.movieJavID ? "缺少影片编号" : "请选择电影相册";
        msgEl.className = "alert mb-3 alert-danger";
      }
      return;
    }
    addAlbumButton.disabled = true;
    try {
      const data = await request(`/api/movie/${encodeURIComponent(target.movieJavID)}/movie-album-item`, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({movie_jav_id: target.movieJavID, album_name: albumName})
      });
      if (msgEl) {
        msgEl.textContent = data.message || "已加入电影相册";
        msgEl.className = "alert mb-3 alert-success";
      }
      window.setTimeout(() => window.location.reload(), 500);
    } catch (error) {
      if (msgEl) {
        msgEl.textContent = error.message || "加入电影相册失败";
        msgEl.className = "alert mb-3 alert-danger";
      }
      addAlbumButton.disabled = false;
    }
  }
});

document.addEventListener("submit", async (event) => {
  const form = event.target.closest("#formAddMovieCast");
  if (!form) return;
  event.preventDefault();

  const javID = String(form.dataset.jav || "").trim();
  const input = form.querySelector("#movieCastName");
  const msg = form.querySelector("#movieCastMsg");
  const button = form.querySelector('button[type="submit"]');
  const name = input ? String(input.value || "").trim() : "";

  function showCastMsg(value, ok) {
    if (!msg) return;
    msg.textContent = value;
    msg.className = `${ok ? "text-success" : "text-danger"} small`;
    msg.style.display = "block";
  }

  if (!javID) {
    showCastMsg("缺少影片参数", false);
    return;
  }
  if (!name) {
    showCastMsg("演员名不能为空", false);
    return;
  }

  if (button) button.disabled = true;
  try {
    await request(`/api/movie/${encodeURIComponent(javID)}/add-cast`, {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({name})
    });
    showCastMsg("添加成功，正在刷新...", true);
    window.setTimeout(() => window.location.reload(), 300);
  } catch (error) {
    showCastMsg(error.message || "网络错误，请稍后重试", false);
  } finally {
    if (button) button.disabled = false;
  }
});

document.addEventListener("shown.bs.modal", (event) => {
  const modal = event.target.closest("#addMovieCastModal");
  if (!modal) return;
  const input = modal.querySelector("#movieCastName");
  const msg = modal.querySelector("#movieCastMsg");
  if (msg) {
    msg.style.display = "none";
    msg.textContent = "";
  }
  if (input) {
    input.value = "";
    input.focus();
  }
});

async function init() {
  const name = movieNameFromPath();
  if (!name) {
    root.innerHTML = `<div class="alert alert-danger">缺少影片参数</div>`;
    return;
  }
  try {
    const data = await request(`/api/movie/${encodeURIComponent(name)}`);
    renderPage(data);
  } catch (error) {
    root.innerHTML = `<div class="alert alert-danger">${escapeHtml(error.message || "加载失败")}</div>`;
  }
}

init();
