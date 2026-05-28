const root = document.getElementById("page-root");
const triggerPageAPIs = {
  "/triggers/dailybest": "/api/triggers/dailybest",
  "/triggers/seeds": "/api/triggers/seeds",
  "/triggers/post-process": "/api/triggers/post-process",
  "/triggers/backfill": "/api/triggers/backfill",
  "/triggers/fetch-site": "/api/triggers/fetch-site",
  "/triggers/fetch-sehuatang": "/api/triggers/fetch-sehuatang",
  "/triggers/fetch-site-javbus-filtered": "/api/triggers/fetch-site-javbus-filtered",
  "/triggers/fetch-site-sukebei-filtered": "/api/triggers/fetch-site-sukebei-filtered",
};

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

function apiPathForCurrentPage() {
  return triggerPageAPIs[window.location.pathname] || "/api/triggers/dailybest";
}

function syncHeaderNav(pathname) {
  document.querySelectorAll("[data-trigger-nav]").forEach((link) => {
    const active = link.getAttribute("data-trigger-nav") === pathname;
    link.classList.toggle("btn-primary", active);
    link.classList.toggle("btn-outline-primary", !active);
  });
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

function renderTaskButtons(data) {
  const defaultTaskType = String(data.DefaultTaskType || "");
  return (data.TaskButtons || []).map((button) => {
    const taskType = String(button.TaskType || "");
    const active = taskType === defaultTaskType ? " active" : "";
    return `<button type="button" class="btn btn-outline-primary flex-fill${active}" data-task-type="${attr(taskType)}">${escapeHtml(button.Label || taskType)}</button>`;
  }).join("");
}

function renderHeaderLinks(data) {
  const customLinks = (data.HeaderLinks || []).map((link) =>
    `<a href="${attr(link.Href || "#")}" class="btn btn-outline-primary btn-sm mr-2">${escapeHtml(link.Label || "")}</a>`
  ).join("");
  return `${customLinks}
    <a href="/crawler/tasks" class="btn btn-outline-primary btn-sm">任务列表</a>
    <a href="/triggers/sc" class="btn btn-outline-primary btn-sm ml-2">SC 预览与提交</a>`;
}

function explicitFlag(explicit, name) {
  return explicit && explicit[name] ? "1" : "0";
}

function renderInput(name, label, value, placeholder = "", type = "text", explicit = {}, extra = "") {
  return `<div class="col-12 col-md-6 col-xl-2">
    <label class="form-label mb-1">${escapeHtml(label)}</label>
    <input class="form-control" type="${attr(type)}" name="${attr(name)}" value="${attr(value || "")}" placeholder="${attr(placeholder)}" data-explicit="${explicitFlag(explicit, name)}" ${extra}>
  </div>`;
}

function renderSelect(name, label, value, explicit, options) {
  const currentValue = String(value || "");
  const opts = options.map(([optionValue, optionLabel]) =>
    `<option value="${attr(optionValue)}" ${currentValue === optionValue ? "selected" : ""}>${escapeHtml(optionLabel)}</option>`
  ).join("");
  return `<div class="col-12 col-md-6 col-xl-2">
    <label class="form-label mb-1">${escapeHtml(label)}</label>
    <select class="form-select" name="${attr(name)}" data-explicit="${explicitFlag(explicit, name)}">${opts}</select>
  </div>`;
}

function renderDatePair(label, startName, startValue, endName, endValue, explicit) {
  return `<div class="col-12 col-xl-4">
    <label class="form-label mb-1">${escapeHtml(label)}</label>
    <div class="row g-2">
      <div class="col-6"><input class="form-control" type="text" name="${attr(startName)}" value="${attr(startValue || "")}" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}" data-explicit="${explicitFlag(explicit, startName)}"></div>
      <div class="col-6"><input class="form-control" type="text" name="${attr(endName)}" value="${attr(endValue || "")}" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}" data-explicit="${explicitFlag(explicit, endName)}"></div>
    </div>
  </div>`;
}

function renderNumberPair(label, startName, startValue, endName, endValue, explicit, step = "") {
  const stepAttr = step ? `step="${attr(step)}"` : "";
  return `<div class="col-12 col-xl-4">
    <label class="form-label mb-1">${escapeHtml(label)}</label>
    <div class="row g-2">
      <div class="col-6"><input class="form-control" type="number" ${stepAttr} name="${attr(startName)}" value="${attr(startValue || "")}" placeholder="最小值" data-explicit="${explicitFlag(explicit, startName)}"></div>
      <div class="col-6"><input class="form-control" type="number" ${stepAttr} name="${attr(endName)}" value="${attr(endValue || "")}" placeholder="最大值" data-explicit="${explicitFlag(explicit, endName)}"></div>
    </div>
  </div>`;
}

function renderMovieCardFilterFields(data) {
  const filter = data.MovieCardFilter || {};
  const explicit = filter.Explicit || {};
  return `${filter.PageSize ? `<input type="hidden" name="ps" value="${attr(filter.PageSize)}" data-explicit="${explicitFlag(explicit, "ps")}">` : ""}
    ${filter.OrderBy ? `<input type="hidden" name="od" value="${attr(filter.OrderBy)}" data-explicit="${explicitFlag(explicit, "od")}">` : ""}
    ${filter.RandomCount ? `<input type="hidden" name="n" value="${attr(filter.RandomCount)}" data-explicit="${explicitFlag(explicit, "n")}">` : ""}
    ${filter.PersonIds ? `<input type="hidden" name="pid" value="${attr(filter.PersonIds)}" data-explicit="${explicitFlag(explicit, "pid")}">` : ""}
    ${renderInput("cn", "演员", filter.CastNames, "演员名，支持多值", "text", explicit)}
    ${renderInput("gn", "类型", filter.GenreNames, "类型名，支持多值", "text", explicit)}
    ${renderInput("dn", "导演", filter.DirectorName, "导演名", "text", explicit)}
    ${renderInput("pn", "前缀", filter.PrefixName, "前缀", "text", explicit)}
    ${renderInput("mn", "片商", filter.MakerName, "片商", "text", explicit)}
    ${renderInput("ln", "标签", filter.LabelName, "标签", "text", explicit)}
    ${renderInput("lj", "标签 JavID", filter.LabelJavID, "JavLib 标签 ID", "text", explicit)}
    ${renderInput("an", "相册名", filter.AlbumName, "相册名，支持多值", "text", explicit)}
    ${renderInput("wd", "关键词", filter.Word, "中文关键词", "text", explicit)}
    ${renderSelect("mowned", "w_media 库存", filter.MediaOwned, explicit, [
      ["", "全部"], ["1", "全部电影"], ["2", "有 w_media 记录（含移除）"], ["3", "已拥有未移除"],
      ["4", "有字幕未移除"], ["5", "无字幕未移除"], ["6", "已移除"], ["7", "无 w_media 记录"],
    ])}
    ${renderSelect("nd", "下载状态", filter.NeedDownload, explicit, [["", "全部"], ["1", "无需下载"], ["2", "需要下载"]])}
    ${renderInput("drkmin", "在榜天数", filter.DaysInRankMin, "最小值", "number", explicit)}
    ${filter.HideDirs ? "" : `
      ${renderInput("md1", "Media 目录 1", filter.MediaDir1, "目录名", "text", explicit)}
      ${renderInput("md2", "Media 目录 2", filter.MediaDir2, "目录名", "text", explicit)}
      ${renderInput("md3", "Media 目录 3", filter.MediaDir3, "目录名", "text", explicit)}
      ${renderInput("md4", "Media 目录 4", filter.MediaDir4, "目录名", "text", explicit)}
    `}
    ${renderDatePair("上映日期", "rs", filter.ReleasingDateStart, "re", filter.ReleasingDateEnd, explicit)}
    ${renderDatePair("Media 下载时间", "mbs", filter.MediaBirthTimeStart, "mbe", filter.MediaBirthTimeEnd, explicit)}
    ${renderDatePair("上榜日期", "srds", filter.StartRankingDateStart, "srde", filter.StartRankingDateEnd, explicit)}
    ${renderDatePair("最后 SC 时间", "lsctmin", filter.LastScTimeMin, "lsctmax", filter.LastScTimeMax, explicit)}
    ${renderNumberPair("演员年龄", "cay", filter.CastAgeMin, "cao", filter.CastAgeMax, explicit, "0.1")}
    ${renderNumberPair("观看人数", "vwmin", filter.ViewWatchedMin, "vwmax", filter.ViewWatchedMax, explicit)}
    ${renderNumberPair("评分", "smin", filter.ScoreMin, "smax", filter.ScoreMax, explicit, "0.1")}
    ${renderNumberPair("SC 次数", "scmin", filter.ScTimesMin, "scmax", filter.ScTimesMax, explicit)}
    ${renderNumberPair("Come 次数", "comin", filter.ComeTimesMin, "comax", filter.ComeTimesMax, explicit)}`;
}

function renderFetchSiteOrderButtons(data) {
  const filter = data.MovieCardFilter || {};
  const current = String(filter.OrderBy || "");
  const options = [
    ["du", "Detail更新"],
    ["rd", "上映日期"],
    ["vw", "观看人数"],
    ["rk", "上榜日期"],
    ["drk", "在榜天数"],
    ["hrk", "最高排名"],
    ["bt", "下载时间"],
    ["mbt", "Media下载时间"],
    ["lsct", "最后sc"],
    ["sc", "sc次数"],
    ["co", "co次数"],
  ];
  return options.map(([value, label]) =>
    `<button type="button" class="btn btn-outline-primary flex-fill${current === value ? " active" : ""}" data-fetch-site-orderby="${attr(value)}">${escapeHtml(label)}</button>`
  ).join("");
}

function renderFetchSiteFilterPanel(data) {
  if (!data.ShowMovieFilters) {
    return "";
  }
  const filter = data.MovieCardFilter || {};
  const explicit = filter.Explicit || {};
  return `<div id="fetchSiteFilterPanel" class="is-hidden mt-3">
    <div class="card border bg-light mb-0">
      <div class="card-body">
        <div class="d-flex align-items-center justify-content-between mb-3">
          <h5 class="mb-0">筛选影片</h5>
          <div class="d-flex align-items-center">
            <button type="button" id="toggleFetchSiteFilterBtn" class="btn btn-outline-primary btn-sm mr-2">展开筛选</button>
            <button type="button" id="clearFetchSiteFilterBtn" class="btn btn-outline-secondary btn-sm">清空筛选</button>
          </div>
        </div>
        <div id="fetchSiteFilterWrap" class="is-hidden">
          <div class="row g-2 align-items-end">
            ${renderMovieCardFilterFields(data)}
          </div>
        </div>
        <div class="row g-2 align-items-end mt-2">
          <div class="col-12">
            <label class="form-label mb-1">抓取时间过滤（天）</label>
            <div class="text-muted small mb-1">小于设定天数的影片将被跳过。</div>
          </div>
          <div class="col-12 col-md-6 col-xl-3">
            <label class="form-label mb-1">最近抓取</label>
            <input type="number" class="form-control" name="last_fetch_duration_days" value="5" placeholder="例如 3">
          </div>
          <div class="col-12 col-md-6 col-xl-3">
            <label class="form-label mb-1">最近成功</label>
            <input type="number" class="form-control" name="last_success_duration_days" value="5" placeholder="例如 3">
          </div>
        </div>
        <div class="row g-2 align-items-end mt-2">
          <div class="col-12">
            <label class="form-label mb-1">排序规则</label>
            <div class="btn-group btn-group-toggle d-flex flex-wrap task-type-toggle" role="group" aria-label="fetch site movie order by">
              ${renderFetchSiteOrderButtons(data)}
            </div>
          </div>
          <div class="col-12 col-md-6 col-xl-2">
            <label class="form-label mb-1">排序顺序</label>
            <select class="form-select" name="order" data-explicit="${explicitFlag(explicit, "order")}">
              <option value="" ${String(filter.Order || "") === "" ? "selected" : ""}>默认</option>
              <option value="asc" ${String(filter.Order || "") === "asc" ? "selected" : ""}>升序</option>
              <option value="desc" ${String(filter.Order || "") === "desc" ? "selected" : ""}>降序</option>
            </select>
          </div>
        </div>
      </div>
    </div>
  </div>`;
}

function renderFetchSitePreviewPanel(data) {
  if (!data.ShowMovieFilters) {
    return "";
  }
  return `<div class="row">
    <div class="col-12">
      <div class="card is-hidden" id="fetchSitePreviewPanel">
        <div class="card-body">
          <div class="d-flex align-items-center justify-content-between mb-3">
            <div>
              <h4 class="card-title mb-0">筛选结果预览</h4>
              <div class="text-muted small mt-1" id="fetchSitePreviewSummary">当前还没有预览数据</div>
            </div>
            <div class="d-flex align-items-center">
              <span class="badge bg-light text-dark mr-2" id="fetchSitePreviewMetaText">共 0 条</span>
              <button type="button" class="btn btn-outline-secondary btn-sm" id="hideFetchSitePreviewBtn">收起</button>
            </div>
          </div>
          <div class="table-responsive">
            <table class="table table-sm table-hover align-middle mb-0">
              <thead>
                <tr>
                  <th style="width: 88px;">序号</th>
                  <th>影片</th>
                  <th style="width: 240px;">JavID</th>
                  <th style="width: 180px;">来源</th>
                </tr>
              </thead>
              <tbody id="fetchSitePreviewTableBody">
                <tr>
                  <td colspan="4" class="text-center text-muted py-4">等待预览</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="d-flex align-items-center justify-content-between mt-3">
            <div class="text-muted small" id="fetchSitePreviewPageText">第 0/0 页</div>
            <div>
              <button type="button" class="btn btn-outline-secondary btn-sm mr-2" id="fetchSitePreviewPrevBtn" disabled>上一页</button>
              <button type="button" class="btn btn-outline-secondary btn-sm" id="fetchSitePreviewNextBtn" disabled>下一页</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>`;
}

function renderStageOverview(data) {
  if (data.OverviewExtraMode === "fetch_site_summary") {
    return `<div class="row">
      <div class="col-md-6 mb-2">
        <div class="overview-item">
          <div class="label">JavBus 进度（总/完成）</div>
          <div class="value" id="fetchSiteJavbusProgressText">0 / 0</div>
        </div>
      </div>
      <div class="col-md-6 mb-2">
        <div class="overview-item">
          <div class="label">JavBus 结果（成功/失败）</div>
          <div class="value" id="fetchSiteJavbusResultText">0 / 0</div>
        </div>
      </div>
      <div class="col-md-6 mb-2">
        <div class="overview-item">
          <div class="label">Sukebei 进度（总/完成）</div>
          <div class="value" id="fetchSiteSukebeiProgressText">0 / 0</div>
        </div>
      </div>
      <div class="col-md-6 mb-2">
        <div class="overview-item">
          <div class="label">Sukebei 结果（成功/失败）</div>
          <div class="value" id="fetchSiteSukebeiResultText">0 / 0</div>
        </div>
      </div>
    </div>`;
  }

  if (data.OverviewExtraMode === "sehuatang_progress") {
    return `<div class="row">
      <div class="col-md-4 mb-2">
        <div class="overview-item">
          <div class="label">抓取进度（总/完成）</div>
          <div class="value" id="sehuatangProgressText">0 / 0</div>
        </div>
      </div>
      <div class="col-md-4 mb-2">
        <div class="overview-item">
          <div class="label">成功/失败</div>
          <div class="value" id="resultText">0 / 0</div>
        </div>
      </div>
      <div class="col-md-4 mb-2">
        <div class="overview-item">
          <div class="label">已运行时长</div>
          <div class="value" id="elapsedText">00:00</div>
        </div>
      </div>
    </div>`;
  }

  if (data.OverviewExtraMode === "javbus_filtered_queue" || data.OverviewExtraMode === "sukebei_filtered_queue") {
    const siteLabel = data.OverviewExtraMode === "javbus_filtered_queue" ? "JavBus" : "Sukebei";
    return `<div class="row">
      <div class="col-md-3 col-6 mb-2">
        <div class="overview-item">
          <div class="label">${escapeHtml(siteLabel)} 队列总数</div>
          <div class="value" id="filteredQueueTotalText">0</div>
        </div>
      </div>
      <div class="col-md-3 col-6 mb-2">
        <div class="overview-item">
          <div class="label">待入列</div>
          <div class="value" id="filteredQueuePendingText">0</div>
        </div>
      </div>
      <div class="col-md-3 col-6 mb-2">
        <div class="overview-item">
          <div class="label">成功</div>
          <div class="value" id="filteredQueueSuccessText">0</div>
        </div>
      </div>
      <div class="col-md-3 col-6 mb-2">
        <div class="overview-item">
          <div class="label">失败</div>
          <div class="value" id="filteredQueueFailedText">0</div>
        </div>
      </div>
      <div class="col-12 mt-2">
        <div class="queue-panel">
          <div class="queue-column">
            <div class="queue-title">待入列 / 处理中</div>
            <div class="queue-list" id="filteredQueuePendingList">
              <div class="queue-empty">等待任务启动</div>
            </div>
          </div>
          <div class="queue-column">
            <div class="queue-title">已完成</div>
            <div class="queue-list" id="filteredQueueDoneList">
              <div class="queue-empty">等待任务启动</div>
            </div>
          </div>
        </div>
      </div>
    </div>`;
  }

  if (data.OverviewExtraMode === "post_process_stages") {
    return `<div class="row dailybest-stage-row">
      <div class="col-12 mb-2">
        <div class="overview-item">
          <div class="label">当前阶段</div>
          <div class="value" id="currentStageText">-</div>
          <div class="subvalue" id="elapsedText">已运行时长 00:00</div>
        </div>
      </div>
    </div>
    <div class="row dailybest-stage-row">
      <div class="col-md-6 mb-2">
        <div class="overview-item overview-stage-card" id="dailyBestBestinvProgressCard" data-phase="cover">
          <div class="label">封面下载进度</div>
          <div class="value" id="dailyBestBestinvProgressText">-</div>
        </div>
      </div>
      <div class="col-md-6 mb-2">
        <div class="overview-item overview-stage-card" id="dailyBestBestinvResultCard" data-phase="cover">
          <div class="label">封面下载正确/错误</div>
          <div class="value" id="dailyBestBestinvResultText">0 / 0</div>
        </div>
      </div>
    </div>
    <div class="row dailybest-stage-row">
      <div class="col-md-6 mb-2">
        <div class="overview-item overview-stage-card" id="dailyBestDetailProgressCard" data-phase="translate">
          <div class="label">标题翻译进度</div>
          <div class="value" id="dailyBestDetailProgressText">-</div>
        </div>
      </div>
      <div class="col-md-6 mb-2">
        <div class="overview-item overview-stage-card" id="dailyBestDetailResultCard" data-phase="translate">
          <div class="label">标题翻译正确/错误</div>
          <div class="value" id="dailyBestDetailResultText">0 / 0</div>
        </div>
      </div>
    </div>`;
  }

  if (data.OverviewExtraMode !== "dailybest_stages" && data.OverviewExtraMode !== "seeds_stages") {
    const labels = data.Labels || {};
    return `<div class="row">
      <div class="col-md-4 mb-2">
        <div class="overview-item">
          <div class="label">${escapeHtml(labels.Extra || "任务类型")}</div>
          <div class="value" id="detailLoopText">-</div>
        </div>
      </div>
      <div class="col-md-4 mb-2">
        <div class="overview-item">
          <div class="label">${escapeHtml(labels.Result || "成功/失败")}</div>
          <div class="value" id="resultText">0 / 0</div>
        </div>
      </div>
      <div class="col-md-4 mb-2">
        <div class="overview-item">
          <div class="label">${escapeHtml(labels.Elapsed || "已运行时长")}</div>
          <div class="value" id="elapsedText">00:00</div>
        </div>
      </div>
    </div>`;
  }

  const isSeedsStages = data.OverviewExtraMode === "seeds_stages";
  const firstPhase = isSeedsStages ? "seed" : "bestinv";
  const firstLabel = isSeedsStages ? "种子抓取" : "榜单抓取";

  return `<div class="row dailybest-stage-row">
    <div class="col-12 mb-2">
      <div class="overview-item">
        <div class="label">当前阶段</div>
        <div class="value" id="currentStageText">-</div>
        <div class="subvalue" id="elapsedText">已运行时长 00:00</div>
      </div>
    </div>
  </div>
  <div class="row dailybest-stage-row">
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestBestinvProgressCard" data-phase="${attr(firstPhase)}">
        <div class="label">${escapeHtml(firstLabel)}进度</div>
        <div class="value" id="dailyBestBestinvProgressText">-</div>
      </div>
    </div>
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestBestinvResultCard" data-phase="${attr(firstPhase)}">
        <div class="label">${escapeHtml(firstLabel)}正确/错误</div>
        <div class="value" id="dailyBestBestinvResultText">0 / 0</div>
      </div>
    </div>
  </div>
  <div class="row dailybest-stage-row">
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestDetailProgressCard" data-phase="detail">
        <div class="label">详情抓取进度</div>
        <div class="value" id="dailyBestDetailProgressText">-</div>
      </div>
    </div>
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestDetailResultCard" data-phase="detail">
        <div class="label">详情抓取正确/错误</div>
        <div class="value" id="dailyBestDetailResultText">0 / 0</div>
      </div>
    </div>
  </div>
  <div class="row dailybest-stage-row">
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestCoverProgressCard" data-phase="cover">
        <div class="label">封面下载进度</div>
        <div class="value" id="dailyBestCoverProgressText">-</div>
      </div>
    </div>
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestCoverResultCard" data-phase="cover">
        <div class="label">封面下载正确/错误</div>
        <div class="value" id="dailyBestCoverResultText">0 / 0</div>
      </div>
    </div>
  </div>
  <div class="row dailybest-stage-row">
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestTranslateProgressCard" data-phase="translate">
        <div class="label">标题翻译进度</div>
        <div class="value" id="dailyBestTranslateProgressText">-</div>
      </div>
    </div>
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestTranslateResultCard" data-phase="translate">
        <div class="label">标题翻译正确/错误</div>
        <div class="value" id="dailyBestTranslateResultText">0 / 0</div>
      </div>
    </div>
  </div>
  <div class="row dailybest-stage-row">
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestFetchJavbusProgressCard" data-phase="fetch_javbus">
        <div class="label">JavBus 抓取进度</div>
        <div class="value" id="dailyBestFetchJavbusProgressText">-</div>
      </div>
    </div>
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestFetchJavbusResultCard" data-phase="fetch_javbus">
        <div class="label">JavBus 抓取正确/错误</div>
        <div class="value" id="dailyBestFetchJavbusResultText">0 / 0</div>
      </div>
    </div>
  </div>
  <div class="row dailybest-stage-row">
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestFetchSukebeiProgressCard" data-phase="fetch_sukebei">
        <div class="label">Sukebei 抓取进度</div>
        <div class="value" id="dailyBestFetchSukebeiProgressText">-</div>
      </div>
    </div>
    <div class="col-md-6 mb-2">
      <div class="overview-item overview-stage-card" id="dailyBestFetchSukebeiResultCard" data-phase="fetch_sukebei">
        <div class="label">Sukebei 抓取正确/错误</div>
        <div class="value" id="dailyBestFetchSukebeiResultText">0 / 0</div>
      </div>
    </div>
  </div>`;
}

function renderPage(data) {
  return `<div id="crawlerJobsPage">
    <div class="row">
      <div class="col-12">
        <div class="page-title-box d-flex align-items-center justify-content-between">
          <div>
            <h4 class="mb-0">${escapeHtml(data.PageTitle || data.Title || "后台任务控制台")}</h4>
            ${data.PageNote ? `<div class="text-muted small mt-1">${escapeHtml(data.PageNote)}</div>` : ""}
          </div>
          <div class="task-link-buttons">${renderHeaderLinks(data)}</div>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-lg-5">
        <div class="card">
          <div class="card-body">
            <h4 class="card-title mb-3">${escapeHtml(data.TaskPanelTitle || "任务触发与实时进度")}</h4>
            ${data.TaskHelperText ? `<div class="alert alert-info small py-2 px-3 mb-3">${escapeHtml(data.TaskHelperText)}</div>` : ""}
            <form id="crawlerTaskForm" class="${data.HideTaskForm ? "is-hidden" : ""}">
              <input type="hidden" id="task_type" name="task_type" value="${attr(data.DefaultTaskType || "")}">
              <div class="form-grid">
                <div class="form-group compact-field grid-span-2">
                  <label>触发任务</label>
                  <div class="btn-group btn-group-toggle d-flex flex-wrap task-type-toggle" role="group" aria-label="crawler task type">
                    ${renderTaskButtons(data)}
                  </div>
                </div>

                <div class="form-group compact-field grid-span-2 is-hidden" id="nameFieldWrap">
                  <label for="task_name">Seed 名称</label>
                  <input id="task_name" name="name" class="form-control" type="text" placeholder="输入 seed 名称">
                </div>

                <div class="form-group compact-field grid-span-2 is-hidden" id="numberFieldWrap">
                  <label for="task_number">刷新数量</label>
                  <input id="task_number" name="number" class="form-control" type="number" value="1000000" placeholder="输入刷新数量">
                </div>

                <div class="form-group compact-field grid-span-2 is-hidden" id="actorNameFieldWrap">
                  <label for="task_actor_name">演员名称</label>
                  <input id="task_actor_name" name="actor_name" class="form-control" type="text" placeholder="输入演员名称">
                </div>

                <div class="form-group compact-field grid-span-2 is-hidden" id="translateStatusesFieldWrap">
                  <label class="d-block mb-2">标题翻译状态（可多选）</label>
                  <div class="d-flex flex-wrap gap-3">
                    <div class="form-check">
                      <input id="translate_status_1" name="statuses" class="form-check-input" type="checkbox" value="1" checked>
                      <label class="form-check-label" for="translate_status_1">无中文字幕(1)</label>
                    </div>
                    <div class="form-check">
                      <input id="translate_status_2" name="statuses" class="form-check-input" type="checkbox" value="2">
                      <label class="form-check-label" for="translate_status_2">有中文字幕(2)</label>
                    </div>
                    <div class="form-check">
                      <input id="translate_status_3" name="statuses" class="form-check-input" type="checkbox" value="3">
                      <label class="form-check-label" for="translate_status_3">翻译错误(3)</label>
                    </div>
                    <div class="form-check">
                      <input id="translate_status_4" name="statuses" class="form-check-input" type="checkbox" value="4">
                      <label class="form-check-label" for="translate_status_4">敏感词(4)</label>
                    </div>
                  </div>
                </div>

                <div class="form-group compact-field grid-span-2 is-hidden" id="autoFetchSiteFieldWrap">
                  <label class="d-block mb-2">详情后自动抓取外站</label>
                  <div class="form-check">
                    <input id="task_auto_fetch_site" name="auto_fetch_site" class="form-check-input" type="checkbox" checked>
                    <label class="form-check-label" for="task_auto_fetch_site">detail 解析完成后自动抓取 JavBus / Sukebei</label>
                  </div>
                </div>

                <div class="form-group compact-field grid-span-2 is-hidden" id="fetchSehuatangListURLFieldWrap">
                  <label for="task_sehuatang_list_url">列表页 URL</label>
                  <input id="task_sehuatang_list_url" name="list_url" class="form-control" type="text" value="${attr(data.FetchSehuatangList || "")}" placeholder="输入 色花堂列表页 URL">
                </div>

                <div class="form-group compact-field grid-span-2 is-hidden" id="fetchSehuatangKeywordFieldWrap">
                  <label for="task_sehuatang_keyword">标题关键词</label>
                  <input id="task_sehuatang_keyword" name="keyword" class="form-control" type="text" value="${attr(data.FetchSehuatangKey || "")}" placeholder="可选，留空表示抓取全部普通帖">
                </div>

                <div class="form-group compact-field is-hidden" id="fetchSehuatangStartPageFieldWrap">
                  <label for="task_sehuatang_start_page">起始页</label>
                  <input id="task_sehuatang_start_page" name="start_page" class="form-control" type="number" value="${attr(data.FetchSehuatangStartPage || "")}" placeholder="输入起始页">
                </div>

                <div class="form-group compact-field is-hidden" id="fetchSehuatangEndPageFieldWrap">
                  <label for="task_sehuatang_end_page">结束页</label>
                  <input id="task_sehuatang_end_page" name="end_page" class="form-control" type="number" value="${attr(data.FetchSehuatangEndPage || "")}" placeholder="输入结束页">
                </div>

                <div class="form-group compact-field grid-span-2 is-hidden" id="fetchSehuatangPersistModeFieldWrap">
                  <label for="task_sehuatang_persist_mode">入库模式</label>
                  <select id="task_sehuatang_persist_mode" name="persist_mode" class="form-select">
                    <option value="upsert_all">upsert 所有帖子</option>
                    <option value="skip_existing">跳过数据库中已存在的帖子</option>
                  </select>
                </div>
              </div>

              ${renderFetchSiteFilterPanel(data)}

              <div class="d-flex align-items-center flex-wrap action-row">
                ${data.ShowMovieFilters ? `<button type="button" id="previewFetchSiteBtn" class="btn btn-outline-info mr-2">预览筛选结果</button>` : ""}
                <button type="submit" id="submitBtn" class="btn btn-primary">开始执行</button>
                <button type="button" id="pauseBtn" class="btn btn-outline-warning ml-2" disabled>暂停</button>
                <button type="button" id="resumeBtn" class="btn btn-outline-success ml-2" disabled>继续</button>
                <button type="button" id="stopBtn" class="btn btn-outline-danger ml-2" disabled>停止</button>
              </div>
            </form>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <h4 class="card-title mb-3">运行概览</h4>
            <div class="crawl-overview">
              ${renderStageOverview(data)}
              <div class="progress mt-3">
                <div id="progressBar" class="progress-bar progress-bar-striped progress-bar-animated" role="progressbar" style="width: 0%" aria-valuemin="0" aria-valuemax="100" aria-valuenow="0">0%</div>
              </div>
              <div class="mt-3 d-flex align-items-center justify-content-between">
                <span id="statusMsg" class="text-muted">${escapeHtml(data.EmptyStateText || "")}</span>
                <span class="badge bg-light text-dark" id="eventCountText">0 条</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="col-lg-7">
        <div class="card event-card">
          <div class="card-body">
            <div class="d-flex align-items-center justify-content-between mb-3">
              <h4 class="card-title mb-0">${escapeHtml(data.EventTitle || "事件流")}</h4>
              <div class="custom-control custom-checkbox">
                <input type="checkbox" class="custom-control-input" id="debugRaw">
                <label class="custom-control-label" for="debugRaw">显示原始 JSON</label>
              </div>
            </div>
            <div class="result-box">
              <pre id="resultBox">${escapeHtml(data.EmptyStateText || "")}</pre>
            </div>
          </div>
        </div>
      </div>
    </div>
    ${renderFetchSitePreviewPanel(data)}
  </div>`;
}

function loadLegacyConsole(data) {
  window.CRAWLER_JOBS_PAGE_DATA = {
    initialJobId: data.JobID || "",
    storageKey: data.StorageKey || "",
    tasksPageUrl: "/crawler/tasks",
    scPreviewUrl: "/triggers/sc",
    defaultTaskType: data.DefaultTaskType || "",
    overviewExtraMode: data.OverviewExtraMode || "",
    emptyStateText: data.EmptyStateText || "",
    fetchSehuatangDefaults: {
      listUrl: data.FetchSehuatangList || "",
      keyword: data.FetchSehuatangKey || "",
      startPage: data.FetchSehuatangStartPage || 0,
      endPage: data.FetchSehuatangEndPage || 0,
      persistMode: data.FetchSehuatangPersistMode || "",
    },
    allowedTaskTypes: Array.isArray(data.AllowedTaskTypes) ? data.AllowedTaskTypes : [],
  };

  if (typeof window.createCrawlerJobConsole === "function") {
    window.createCrawlerJobConsole(window.CRAWLER_JOBS_PAGE_DATA);
    return;
  }

  const script = document.createElement("script");
  script.src = "/static/customer/crawler_jobs.js";
  script.async = false;
  document.body.appendChild(script);
}

async function main() {
  try {
    syncHeaderNav(window.location.pathname);
    const data = await fetchJSON(apiPathForCurrentPage() + window.location.search);
    document.title = `${data.Title || data.PageTitle || "Triggers"} - Rudy-GC`;
    root.innerHTML = renderPage(data);
    loadLegacyConsole(data);
  } catch (err) {
    root.innerHTML = `<div class="alert alert-danger">加载失败：${escapeHtml(err.message)}</div>`;
  }
}

main();
