<template>
  <PageShell
    class="fetch-site-list-page-shell"
    :class="{ 'sc-events-page-shell': viewType === 'fetch-site-sehuatang-list' }"
    :title="title"
    eyebrow=""
  >
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <div
      v-else
      id="page-root"
      class="fetch-site-list-page-root"
      :class="listRewriteClass"
      @click="handlePageClick"
      @submit="handlePageSubmit"
      v-html="pageHTML"
    ></div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LoadingState from "@/components/LoadingState.vue";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, pick, text as readText } from "@/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const otherAlbumTarget = ref<{ movieJavID: string; sourceType: string; sourceRowID: number } | null>(null);

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value.PageTitle || payload.value.Title || route.meta.title || "List"));
const viewType = computed(() => text(payload.value.ViewType));
const listRewriteClass = computed(() => {
  if (viewType.value === "fetch-site-javbus-list") {
    return "next02-fetch-site-javbus-list-rewrite";
  }
  if (viewType.value === "fetch-site-sukebei-list") {
    return "next02-fetch-site-sukebei-list-rewrite";
  }
  if (viewType.value === "fetch-site-sehuatang-list") {
    return "next02-fetch-site-sehuatang-list-rewrite container-fluid next02-sc-pink-theme next02-sc-events-rewrite";
  }
  return "";
});
const pageHTML = computed(() => {
  if (viewType.value === "fetch-site-javbus-list") {
    return renderFetchSitePage(payload.value, "javbus");
  }
  if (viewType.value === "fetch-site-sukebei-list") {
    return renderFetchSitePage(payload.value, "sukebei");
  }
  if (viewType.value === "fetch-site-sehuatang-list") {
    return renderSehuatangPage(payload.value);
  }
  if (viewType.value === "fetch-site-sehuatang-sh-list") {
    return renderSHPage(payload.value);
  }
  if (viewType.value === "fetch-site-sehuatang-sh-detail") {
    return renderSHDetailPage(payload.value);
  }
  return renderFallbackTable(payload.value);
});

function escapeHtml(value: unknown): string {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function attr(value: unknown): string {
  return escapeHtml(value);
}

function showMsg(message: string, ok: boolean) {
  const msgEl = document.getElementById("fetchSiteListMsg")
    || document.getElementById("sehuatangMagnetListMsg")
    || document.getElementById("sehuatangSHListMsg")
    || document.getElementById("shDetailMsg");
  if (!msgEl) {
    return;
  }
  msgEl.textContent = message;
  msgEl.className = `alert small py-2 px-3 mb-3 ${ok ? "alert-success" : "alert-danger"}`;
}

function setTriggerButtonsLoading(loading: boolean) {
  document.querySelectorAll<HTMLButtonElement>(".js-trigger-javbus, .js-trigger-sukebei, .js-trigger-both, .js-trigger-filtered-fetch-site, .js-trigger-filtered-sukebei").forEach((button) => {
    button.disabled = loading;
  });
}

function text(value: unknown, fallback = "-"): string {
  const current = String(value ?? "").trim();
  return current === "" ? fallback : current;
}

function valueOf(value: unknown, fallback: unknown = ""): unknown {
  return value === undefined || value === null ? fallback : value;
}

function formatUnix(sec: unknown, mode = ""): string {
  const n = Number(sec || 0);
  if (!Number.isFinite(n) || n <= 0) {
    return "-";
  }
  const date = new Date(n * 1000);
  const y = String(date.getFullYear());
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  const h = String(date.getHours()).padStart(2, "0");
  const min = String(date.getMinutes()).padStart(2, "0");
  const s = String(date.getSeconds()).padStart(2, "0");
  if (mode === "date") {
    return `${y}-${m}-${d}`;
  }
  if (mode === "hm") {
    return `${h}:${min}`;
  }
  return `${y}-${m}-${d} ${h}:${min}:${s}`;
}

function pageLinks(pageInfo: unknown, label: string): string {
  const links = asArray<AnyRecord>(pick(asRecord(pageInfo), "Links"));
  if (!links.length) {
    return "";
  }
  return `<nav class="mb-0" aria-label="${attr(label)} Page navigation">
    <ul class="pagination pagination-sm mb-0">
      ${links.map((link) => {
        if (pick(link, "Ellipsis")) {
          return `<li class="page-item disabled"><span class="page-link">...</span></li>`;
        }
        return `<li class="page-item ${pick(link, "Active") ? "active" : ""} ${pick(link, "Disabled") ? "disabled" : ""}">
          <a class="page-link" href="${attr(readText(link, "Href") || "#")}" ${pick(link, "Disabled") ? `tabindex="-1" aria-disabled="true"` : ""}>${escapeHtml(readText(link, "Label"))}</a>
        </li>`;
      }).join("")}
    </ul>
  </nav>`;
}

function sortLabel(link: unknown, label: string): string {
  const row = asRecord(link);
  const arrow = pick(row, "Active") ? (pick(row, "Desc") ? " ↓" : " ↑") : "";
  return `<a class="fetch-site-sort-link sc-sort-link ${pick(row, "Active") ? "active" : ""}" href="${attr(readText(row, "Href") || "#")}">${escapeHtml(label)}${arrow}</a>`;
}

function statusChecked(data: AnyRecord, value: number): string {
  const selected = asRecord(pick(data, "StatusSelected"));
  return selected[String(value)] ? "checked" : "";
}

function renderFetchSiteStatusGroup(data: AnyRecord, siteKey: string): string {
  const statusID = (value: number) => `${siteKey}-status-${value}`;
  return `<div class="fetch-site-status-group" data-status-group>
    <button type="button" class="btn btn-sm btn-outline-secondary fetch-site-status-clear ${pick(data, "StatusAllActive") ? "active" : ""}" data-status-clear>全部</button>
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

function triggerSortOptions(query: AnyRecord): string {
  const current = text(pick(query, "TriggerSortKey"), "");
  const options = [
    ["last_fetch_time:asc", "最后抓取时间升序"],
    ["last_fetch_time:desc", "最后抓取时间降序"],
    ["release_date:asc", "发行时间升序"],
    ["release_date:desc", "发行时间降序"],
    ["latest_publish_time:asc", "最新日期升序"],
    ["latest_publish_time:desc", "最新日期降序"],
    ["movie_name:asc", "影片名升序"],
    ["movie_name:desc", "影片名降序"]
  ];
  return options.map(([value, label]) => {
    const selected = (current === "" && value === "last_fetch_time:asc") || current === value;
    return `<option value="${attr(value)}" ${selected ? "selected" : ""}>${escapeHtml(label)}</option>`;
  }).join("");
}

function renderOwnedFilter(data: AnyRecord): string {
  const owned = asRecord(pick(data, "ownedQuery"));
  const total = valueOf(pick(data, "total"), pick(data, "Total") || 0);
  const link = (item: unknown, label: string, className: string) => {
    const current = asRecord(item);
    return `<a href="${attr(readText(current, "Href") || "#")}"
      class="btn btn-sm ${className} waves-effect waves-light ${pick(current, "Active") ? "active disabled" : ""}"
      aria-current="${pick(current, "Active") ? "true" : "false"}">${escapeHtml(label)}</a>`;
  };
  return `<div class="d-flex flex-wrap align-items-center gap-3">
    <div class="d-flex flex-wrap align-items-center gap-2">
      <span class="btn btn-sm btn-outline-dark disabled">WMedia</span>
      ${link(pick(owned, "MediaAll"), "MAll", "btn-primary")}
      ${link(pick(owned, "MediaOwned"), "MOwned", "btn-success")}
      ${link(pick(owned, "MediaNotOwned"), "MNotOwned", "btn-secondary")}
    </div>
    <span class="btn btn-sm btn-primary waves-effect waves-light disabled">Total: ${escapeHtml(total)}</span>
  </div>`;
}

function renderFetchSiteFilterForm(data: AnyRecord, site: "javbus" | "sukebei"): string {
  const q = asRecord(pick(data, "Query"));
  const action = site === "javbus" ? "/fetch-site-javbus-list" : "/fetch-site-sukebei-list";
  const formID = site === "javbus" ? "fetchSiteJavbusFilterForm" : "fetchSiteSukebeiFilterForm";
  const siteLabel = site === "javbus" ? "JavBus" : "Sukebei";
  const triggerClass = site === "javbus" ? "js-trigger-filtered-fetch-site" : "js-trigger-filtered-sukebei";
  return `<div class="card mb-3">
    <div class="card-body">
      <form class="row g-3 align-items-end" method="get" action="${attr(action)}" id="${attr(formID)}" ${site === "javbus" ? `data-fetch-site-filter-form="1"` : ""}>
        <input type="hidden" name="ps" value="${attr(pick(q, "PageSize") || "")}">
        ${pick(q, "Sort") ? `<input type="hidden" name="sort" value="${attr(pick(q, "Sort"))}">` : ""}
        ${pick(q, "Order") ? `<input type="hidden" name="order" value="${attr(pick(q, "Order"))}">` : ""}
        ${pick(q, "MediaOwned") ? `<input type="hidden" name="mowned" value="${attr(pick(q, "MediaOwned"))}">` : ""}
        <div class="col-12">
          <div class="fetch-site-filter-grid fetch-site-filter-grid-5 fetch-site-sukebei-filter-top">
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">${escapeHtml(siteLabel)} 影片编号 / JavID</label>
              <input class="form-control" type="text" name="keyword" value="${attr(pick(q, "Keyword") || "")}" placeholder="movie_name 或 movie_jav_id">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">${escapeHtml(siteLabel)} 状态</label>
              ${renderFetchSiteStatusGroup(data, site)}
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">最后抓取开始</label>
              <input class="form-control" type="text" name="last_fetch_from" value="${attr(pick(q, "LastFetchFrom") || "")}">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">最后抓取结束</label>
              <input class="form-control" type="text" name="last_fetch_to" value="${attr(pick(q, "LastFetchTo") || "")}">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">触发抓取顺序</label>
              <select class="form-select" name="trigger_sort_key">${triggerSortOptions(q)}</select>
            </div>
          </div>
        </div>
        <div class="col-12">
          <div class="fetch-site-filter-grid fetch-site-filter-grid-6">
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">发行时间开始</label>
              <input class="form-control" type="text" name="release_date_from" value="${attr(pick(q, "ReleaseDateFrom") || "")}">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">发行时间结束</label>
              <input class="form-control" type="text" name="release_date_to" value="${attr(pick(q, "ReleaseDateTo") || "")}">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">WMedia 下载时间开始</label>
              <input class="form-control" type="text" name="media_birth_from" value="${attr(pick(q, "MediaBirthFrom") || "")}">
            </div>
            <div class="fetch-site-filter-item">
              <label class="form-label mb-1">WMedia 下载时间结束</label>
              <input class="form-control" type="text" name="media_birth_to" value="${attr(pick(q, "MediaBirthTo") || "")}">
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

function renderWMediaCell(row: AnyRecord): string {
  const owned = Number(pick(row, "OwnedWMedia") || 0);
  const birth = pick(row, "FilmBirthDateWMedia") ? `<span class="text-muted small ms-1">${escapeHtml(pick(row, "FilmBirthDateWMedia"))}</span>` : "";
  if (owned === 5) {
    return `<a href="${attr(pick(row, "VideoURLWMedia") || "#")}" class="btn btn-sm btn-info me-2">MPlay</a>${birth}`;
  }
  if (owned === 4) {
    return `<a href="${attr(pick(row, "VideoURLWMedia") || "#")}" class="btn btn-sm btn-success me-2">MPlay</a>${birth}`;
  }
  if (owned === 6) {
    return `<span class="btn btn-sm btn-secondary me-2 disabled" aria-disabled="true">MRemoved</span>${birth}`;
  }
  return `-${birth}`;
}

function renderFetchSiteRow(row: AnyRecord, site: "javbus" | "sukebei"): string {
  const triggerClass = site === "javbus" ? "js-trigger-javbus" : "js-trigger-sukebei";
  const triggerLabel = site === "javbus" ? "触发 JavBus" : "触发 Sukebei";
  const triggerButtonClass = site === "javbus" ? "btn-outline-primary" : "btn-outline-danger";
  const movieName = readText(row, "MovieName");
  const errorHTML = readText(row, "LastError")
    ? `<details class="fetch-site-error-details"><summary>有错误</summary><div class="mt-2">${escapeHtml(readText(row, "LastError"))}</div></details>`
    : `<span class="text-muted">无错误</span>`;
  return `<tr data-movie-jav-id="${attr(readText(row, "MovieJavID"))}" data-movie-name="${attr(movieName)}">
    <td class="text-nowrap"><div class="fw-semibold"><a href="/movie/${attr(encodeURIComponent(movieName))}">${escapeHtml(movieName)}</a></div></td>
    <td class="text-nowrap">${escapeHtml(readText(row, "ReleaseDateText"))}</td>
    <td><span class="badge fetch-site-badge">${escapeHtml(readText(row, "FetchStatusText"))}</span></td>
    <td>${escapeHtml(valueOf(pick(row, "TryCount"), 0))}</td>
    <td class="text-nowrap">${escapeHtml(readText(row, "LastFetchText"))}</td>
    <td>${escapeHtml(valueOf(pick(row, "TorrentHashCount"), 0))}</td>
    <td class="text-nowrap">${escapeHtml(readText(row, "LatestPublishText"))}</td>
    <td class="text-nowrap">${renderWMediaCell(row)}</td>
    <td class="fetch-site-error-cell">${errorHTML}</td>
    <td class="text-nowrap"><button type="button" class="btn btn-sm ${triggerButtonClass} ${triggerClass}">${escapeHtml(triggerLabel)}</button></td>
  </tr>`;
}

function renderFetchSiteTable(data: AnyRecord, site: "javbus" | "sukebei"): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  const sort = asRecord(pick(data, "SortQuery"));
  const siteLabel = site === "javbus" ? "JavBus" : "Sukebei";
  return `<div class="card">
    <div class="card-body">
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0">
          <thead>
            <tr>
              <th>${sortLabel(pick(sort, "ByMovieName"), "影片")}</th>
              <th>${sortLabel(pick(sort, "ByReleaseDate"), "发行时间")}</th>
              <th>${sortLabel(pick(sort, "ByFetchStatus"), "状态")}</th>
              <th>重试</th>
              <th>${sortLabel(pick(sort, "ByLastFetchTime"), "最后抓取")}</th>
              <th>${sortLabel(pick(sort, "ByHashCount"), "Hash 数")}</th>
              <th>${sortLabel(pick(sort, "ByLatestPublish"), "最新日期")}</th>
              <th>${sortLabel(pick(sort, "ByMediaBirthTime"), "WMedia 下载时间")}</th>
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

function renderFetchSitePage(data: AnyRecord, site: "javbus" | "sukebei"): string {
  const siteLabel = site === "javbus" ? "JavBus" : "Sukebei";
  const peerLabel = site === "javbus" ? "Sukebei" : "JavBus";
  const ownURL = site === "javbus" ? "/fetch-site-javbus-list" : "/fetch-site-sukebei-list";
  const formID = site === "javbus" ? "fetchSiteJavbusFilterForm" : "fetchSiteSukebeiFilterForm";
  const filteredTaskType = site === "javbus" ? pick(data, "FilteredJavbusTaskType") : pick(data, "FilteredSukebeiTaskType");
  const filteredAttr = site === "javbus"
    ? `data-filtered-task-type="${attr(filteredTaskType || "")}" data-filter-form-id="${attr(formID)}"`
    : `data-filtered-sukebei-task-type="${attr(filteredTaskType || "")}"`;
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  return `<div id="fetchSiteListPage" data-task-page-url="${attr(pick(data, "TaskPageURL") || "/triggers/fetch-site")}" ${filteredAttr}>
    <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
      <div>
        <h4 class="mb-1">${escapeHtml(pick(data, "PageTitle") || `${siteLabel} 抓取列表`)}</h4>
        <p class="text-muted mb-0">${escapeHtml(pick(data, "PageNote") || "")}</p>
      </div>
      <div class="d-flex flex-wrap gap-2">
        <a class="btn btn-outline-primary btn-sm" href="${attr(ownURL)}">${escapeHtml(siteLabel)} 列表</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "PeerPageURL") || "#")}">${escapeHtml(peerLabel)} 列表</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "TaskPageURL") || "#")}">抓取任务页</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "TasksPageURL") || "/crawler/tasks")}">任务列表</a>
      </div>
    </div>
    <div id="fetchSiteListMsg" class="alert d-none mb-3" role="alert"></div>
    <div class="d-flex flex-column flex-xl-row align-items-xl-center justify-content-between gap-3 mb-3">
      <div class="d-flex flex-wrap gap-3">
        <div class="fetch-site-summary-card"><span class="fetch-site-summary-label">${escapeHtml(siteLabel)} 总任务数</span><span class="fetch-site-summary-value">${escapeHtml(valueOf(pick(data, "Total"), 0))}</span></div>
        <div class="fetch-site-summary-card"><span class="fetch-site-summary-label">成功数目</span><span class="fetch-site-summary-value">${escapeHtml(valueOf(pick(data, "SuccessCount"), 0))}</span></div>
        <div class="fetch-site-summary-card"><span class="fetch-site-summary-label">待抓取数目</span><span class="fetch-site-summary-value">${escapeHtml(valueOf(pick(data, "PendingCount"), 0))}</span></div>
        <div class="fetch-site-summary-card"><span class="fetch-site-summary-label">失败数目</span><span class="fetch-site-summary-value">${escapeHtml(valueOf(pick(data, "FailedCount"), 0))}</span></div>
      </div>
      ${rows.length ? pageLinks(pick(data, "PageInfo"), siteLabel) : ""}
    </div>
    ${renderFetchSiteFilterForm(data, site)}
    ${renderFetchSiteTable(data, site)}
  </div>`;
}

function renderQuickTags(data: AnyRecord): string {
  return asArray<AnyRecord>(pick(data, "QuickTags")).map((item) =>
    `<a class="btn btn-sm ${pick(item, "Active") ? "btn-primary" : "btn-outline-primary"}" href="${attr(readText(item, "Href") || "#")}">${escapeHtml(readText(item, "Label"))}</a>`
  ).join("");
}

function renderSehuatangFilterForm(data: AnyRecord): string {
  const q = asRecord(pick(data, "Query"));
  return `<form class="sc-event-filter-form sehuatang-filter-form" id="sehuatangMagnetFilterForm" method="get" action="/fetch-site-sehuatang-list">
        <input type="hidden" name="ps" value="${attr(pick(q, "PageSize") || "")}">
        ${pick(q, "Sort") ? `<input type="hidden" name="sort" value="${attr(pick(q, "Sort"))}">` : ""}
        ${pick(q, "Order") ? `<input type="hidden" name="order" value="${attr(pick(q, "Order"))}">` : ""}
        <div class="sehuatang-filter-cluster sehuatang-filter-query" aria-label="检索条件">
          <label class="sc-event-filter-field sehuatang-filter-field is-keyword" for="sehuatangMagnetKeywordInput">
            <span>关键词</span>
            <input class="form-control" id="sehuatangMagnetKeywordInput" type="text" name="keyword" value="${attr(pick(q, "Keyword") || "")}" placeholder="影片名 / 编号 / 帖子标题">
          </label>
          <label class="sc-event-filter-field sehuatang-filter-field is-movie-id" for="sehuatangMagnetMovieIDInput">
            <span>影片编号</span>
            <input class="form-control" id="sehuatangMagnetMovieIDInput" type="text" name="movie_jav_id" value="${attr(pick(q, "MovieJavID") || "")}" placeholder="JavLib 编号">
          </label>
          <label class="sc-event-filter-field sehuatang-filter-field is-hash" for="sehuatangMagnetHashInput">
            <span>Info Hash</span>
            <input class="form-control" id="sehuatangMagnetHashInput" type="text" name="info_hash" value="${attr(pick(q, "InfoHash") || "")}" placeholder="BTIH">
          </label>
          <label class="sc-event-filter-field sehuatang-filter-field is-tag" for="sehuatangMagnetTagInput">
            <span>标签</span>
            <input class="form-control" id="sehuatangMagnetTagInput" type="text" name="tag" value="${attr(pick(q, "Tag") || "")}" placeholder="标签">
          </label>
        </div>
        <div class="sehuatang-filter-cluster sehuatang-filter-date-range" aria-label="发帖日期范围">
          <span class="sehuatang-filter-cluster-label">发帖日期</span>
          <label class="sc-event-filter-field sehuatang-filter-field is-date" for="sehuatangMagnetPostFromInput">
            <span>开始</span>
            <input class="form-control" id="sehuatangMagnetPostFromInput" type="date" name="post_time_from" value="${attr(pick(q, "PostTimeFrom") || "")}">
          </label>
          <span class="sehuatang-filter-range-separator" aria-hidden="true">—</span>
          <label class="sc-event-filter-field sehuatang-filter-field is-date" for="sehuatangMagnetPostToInput">
            <span>结束</span>
            <input class="form-control" id="sehuatangMagnetPostToInput" type="date" name="post_time_to" value="${attr(pick(q, "PostTimeTo") || "")}">
          </label>
        </div>
        <div class="sehuatang-filter-cluster sehuatang-filter-quick-actions" aria-label="快捷筛选与操作">
          <div class="sc-event-filter-field sehuatang-filter-field is-quick-tag">
            <span>快捷标签</span>
            <div class="album-quick-filter-group">${renderQuickTags(data)}</div>
          </div>
          <div class="sehuatang-filter-actions">
            <button class="sc-event-filter-submit" type="submit">筛选</button>
            <a class="sc-event-filter-clear" href="/fetch-site-sehuatang-list">清空</a>
          </div>
        </div>
      </form>`;
}

function renderFavoriteButton(row: AnyRecord, albumName: string, active: boolean, className: string, outlineClass: string): string {
  const canFavorite = Boolean(pick(row, "CanFavorite"));
  return `<button type="button"
    class="btn btn-sm ${active ? className : outlineClass} js-add-favorite"
    data-source-type="sehuatang_magnet"
    data-source-row-id="${attr(pick(row, "Id") || 0)}"
    data-favorited="${active ? "1" : "0"}"
    data-album-name="${attr(albumName)}"
    ${canFavorite ? "" : `disabled title="缺少影片标识，暂不可加入相册"`}>${escapeHtml(albumName)}</button>`;
}

function sehuatangTimeParts(value: unknown): { date: string; time: string } {
  const formatted = formatUnix(value);
  if (formatted === "-") {
    return { date: "-", time: "" };
  }
  const [date, time = ""] = formatted.split(" ", 2);
  return { date, time };
}

function sehuatangMonthKey(row: AnyRecord): string {
  const date = formatUnix(pick(row, "PostTime"), "date");
  return date === "-" ? "" : date.slice(0, 7);
}

function sehuatangMonthClass(row: AnyRecord): string {
  const month = sehuatangMonthKey(row);
  if (!month) {
    return "sc-running-month-neutral";
  }
  const monthNumber = Number(month.slice(5, 7));
  return `sc-running-month-${(monthNumber % 3) + 1}`;
}

function isSehuatangMonthStart(rows: AnyRecord[], index: number): boolean {
  if (index <= 0) {
    return false;
  }
  const currentMonth = sehuatangMonthKey(rows[index]);
  const previousMonth = sehuatangMonthKey(rows[index - 1]);
  return Boolean(currentMonth && previousMonth && currentMonth !== previousMonth);
}

function renderSehuatangTime(value: unknown): string {
  const { date, time } = sehuatangTimeParts(value);
  return `<span class="sc-event-time"><span class="sc-event-date">${escapeHtml(date)}</span>${time ? `<small class="sc-event-hour">${escapeHtml(time)}</small>` : ""}</span>`;
}

function renderSehuatangRow(row: AnyRecord, index: number, rows: AnyRecord[]): string {
  const threadTitle = readText(row, "ThreadTitle");
  const titleHTML = readText(row, "ThreadURL")
    ? `<a class="fetch-site-title-ellipsis" href="${attr(readText(row, "ThreadURL"))}" target="_blank" rel="noopener noreferrer" title="${attr(threadTitle)}">${escapeHtml(threadTitle)}</a>`
    : `<span class="fetch-site-title-ellipsis" title="${attr(threadTitle)}">${escapeHtml(threadTitle)}</span>`;
  const movieName = readText(row, "MovieName");
  const movieHTML = movieName ? `<a class="sehuatang-movie-link" href="/movie/${attr(encodeURIComponent(movieName))}" target="_blank" rel="noopener noreferrer" title="${attr(movieName)}">${escapeHtml(movieName)}</a>` : "-";
  const monthStartClass = isSehuatangMonthStart(rows, index) ? " sc-running-month-start" : "";
  return `<tr class="js-sehuatang-row next02-sc-event-row ${sehuatangMonthClass(row)}${monthStartClass}"
    data-movie-route-key="${attr(readText(row, "MovieRouteKey"))}"
    data-movie-jav-id="${attr(readText(row, "MovieJavID"))}"
    data-movie-name="${attr(movieName)}"
    data-post-day="${attr(formatUnix(pick(row, "PostTime"), "date"))}">
    <td>${renderSehuatangTime(pick(row, "PostTime"))}</td>
    <td class="sc-content-cell">${titleHTML}</td>
    <td class="sc-content-cell">${movieHTML}</td>
    <td class="sehuatang-album-cell">
      <div class="fetch-site-album-inline">
        ${renderFavoriteButton(row, "下载中", Boolean(pick(row, "IsInDownload")), "btn-warning", "btn-outline-warning")}
        ${renderFavoriteButton(row, "待下载", Boolean(pick(row, "IsInPending")), "btn-info", "btn-outline-info")}
        <button type="button" class="btn btn-sm btn-outline-secondary js-open-other-album-modal" data-source-type="sehuatang_magnet" data-source-row-id="${attr(pick(row, "Id") || 0)}" ${pick(row, "CanFavorite") ? "" : `disabled title="缺少影片标识，暂不可加入相册"`}>其他相册</button>
      </div>
    </td>
    <td class="sc-content-cell sehuatang-hash-cell"><span class="fetch-site-hash-text" title="${attr(readText(row, "InfoHash"))}">${escapeHtml(readText(row, "InfoHash"))}</span></td>
    <td>${renderSehuatangTime(pick(row, "LastSeenTime"))}</td>
  </tr>`;
}

function renderSehuatangTable(data: AnyRecord): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  const sort = asRecord(pick(data, "SortQuery"));
  return `<div class="table-responsive">
        <table class="table table-hover align-middle mb-0 sc-events-table sehuatang-magnet-table">
          <colgroup>
            <col class="sehuatang-col-post-time">
            <col class="sehuatang-col-title">
            <col class="sehuatang-col-movie">
            <col class="sehuatang-col-album">
            <col class="sehuatang-col-hash">
            <col class="sehuatang-col-last-seen">
          </colgroup>
          <thead>
            <tr>
              <th class="sc-sortable-head">${sortLabel(pick(sort, "ByPostTime"), "发帖时间")}</th>
              <th>帖子标题</th>
              <th class="sc-sortable-head">${sortLabel(pick(sort, "ByMovieName"), "影片名")}</th>
              <th>相册</th>
              <th>info_hash</th>
              <th class="sc-sortable-head">${sortLabel(pick(sort, "ByLastSeenTime"), "最近抓到")}</th>
            </tr>
          </thead>
          <tbody>${rows.length ? rows.map((row, index) => renderSehuatangRow(row, index, rows)).join("") : `<tr><td colspan="6"><div class="sc-empty-state">当前筛选下没有 Sehuatang 磁力数据。</div></td></tr>`}</tbody>
        </table>
      </div>`;
}

function renderOtherAlbumModal(data: AnyRecord): string {
  const albums = asArray<string>(pick(data, "Albums"));
  return `<div class="modal fade" id="sehuatangOtherAlbumModal" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered"><div class="modal-content">
      <div class="modal-header"><h5 class="modal-title">选择其他相册</h5><button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="关闭"></button></div>
      <div class="modal-body">${albums.length ? `<div class="fetch-site-other-album-list" id="sehuatangOtherAlbumList">${albums.map((name) => `<button type="button" class="btn btn-outline-secondary js-other-album-option" data-album-name="${attr(name)}">${escapeHtml(name)}</button>`).join("")}</div>` : `<div class="text-muted small">当前没有可选的其他相册。</div>`}</div>
    </div></div>
  </div>`;
}

function renderDateModal(): string {
  return `<div class="modal fade sehuatang-date-modal" id="sehuatangDateModal" tabindex="-1" role="dialog" aria-modal="true" aria-labelledby="sehuatangDateModalTitle" aria-hidden="true">
    <div class="modal-dialog sehuatang-date-dialog"><div class="modal-content sehuatang-date-content">
      <form class="sehuatang-date-form" id="sehuatangDateForm">
        <header class="sehuatang-date-header">
          <div class="sehuatang-date-heading">
            <span class="sehuatang-date-eyebrow">SEHUATANG</span>
            <h5 id="sehuatangDateModalTitle">修改日期</h5>
          </div>
          <button type="button" class="sehuatang-date-close" data-bs-dismiss="modal" aria-label="关闭"><span aria-hidden="true">×</span></button>
        </header>
        <div class="sehuatang-date-body">
          <input type="hidden" id="sehuatangDateItemKey" value="">
          <label class="sehuatang-date-label" for="sehuatangDateInput"><span>日期</span><small>YYYY-MM-DD</small></label>
          <div class="sehuatang-date-control-row">
            <input id="sehuatangDateInput" type="text" value="" placeholder="YYYY-MM-DD" inputmode="numeric" autocomplete="off">
            <button type="button" class="sehuatang-date-today" id="sehuatangDateTodayBtn">今天</button>
          </div>
          <div class="alert d-none mb-0 sehuatang-date-message" id="sehuatangDateMsg" role="alert"></div>
        </div>
        <footer class="sehuatang-date-footer">
          <button type="button" class="sehuatang-date-cancel" data-bs-dismiss="modal">取消</button>
          <button type="submit" class="sehuatang-date-save" id="sehuatangDateSubmit">保存日期</button>
        </footer>
      </form>
    </div></div>
  </div>`;
}

function renderSehuatangPage(data: AnyRecord): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  return `<div id="sehuatangMagnetListPage" class="sehuatang-list-content">
    <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell">
      <div class="next02-sc-hero-copy">
        <span class="next02-sc-eyebrow">SEHUATANG / MAGNET POOL</span>
        <h4>${escapeHtml(pick(data, "PageTitle") || "Sehuatang Magnet 列表")}</h4>
        <p>${escapeHtml(pick(data, "PageNote") || "")}</p>
      </div>
      <div class="next02-sc-hero-actions">
        <div class="btn-group" role="group" aria-label="Sehuatang 页面操作">
          <button type="button" class="btn btn-outline-primary" id="btnEditShtTime" data-item-key="sht_time" data-item-label="SHT时间" data-item-value="${attr(pick(data, "ShtTimeValue") || "")}">SHT时间：${escapeHtml(pick(data, "ShtTimeDisplay") || "未设置")}</button>
          <a class="btn btn-outline-secondary" href="${attr(pick(data, "TaskPageURL") || "/triggers/fetch-sehuatang")}" target="_blank" rel="noopener noreferrer">抓取任务页</a>
          <a class="btn btn-outline-secondary" href="${attr(pick(data, "TasksPageURL") || "/crawler/tasks")}" target="_blank" rel="noopener noreferrer">任务列表</a>
        </div>
      </div>
    </section>
    <div id="sehuatangMagnetListMsg" class="alert d-none mb-3" role="alert"></div>
    <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
      <div class="card-body">
        <div class="next02-sc-panel-head">
          ${renderSehuatangFilterForm(data)}
          ${rows.length ? `<div class="next02-sc-hero-metrics sc-event-list-metrics" aria-label="Sehuatang 磁力摘要"><span><strong>${escapeHtml(valueOf(pick(data, "Total"), 0))}</strong> Records</span><span><strong>${escapeHtml(rows.length)}</strong> Current Page</span></div>` : ""}
          ${rows.length ? `<div class="next02-sc-movie-pager">${pageLinks(pick(data, "PageInfo"), "Sehuatang")}</div>` : ""}
        </div>
        ${renderSehuatangTable(data)}
      </div>
    </div>
  </div>
  ${renderOtherAlbumModal(data)}
  ${renderDateModal()}`;
}

function lineCount(value: unknown): number {
  const current = String(value || "").trim();
  return current ? current.split(/\r?\n/).filter((line) => line.trim() !== "").length : 0;
}

function renderOptions(current: unknown, options: [string, string][]): string {
  const value = String(current || "");
  return options.map(([optionValue, label]) =>
    `<option value="${attr(optionValue)}" ${value === optionValue ? "selected" : ""}>${escapeHtml(label)}</option>`
  ).join("");
}

function renderSHFilterForm(data: AnyRecord): string {
  const q = asRecord(pick(data, "Query"));
  return `<div class="card mb-3 next02-sh-filter-panel">
    <div class="card-body">
      <div class="next02-sh-panel-head">
        <div>
          <div class="next02-sh-eyebrow">Filter</div>
          <h5 class="card-title mb-0">筛选帖子池</h5>
        </div>
      </div>
      <form class="row g-3 align-items-end" method="get" action="/fetch-site-sehuatang-sh-list">
        <input type="hidden" name="ps" value="${attr(pick(q, "PageSize") || "")}">
        ${pick(q, "Sort") ? `<input type="hidden" name="sort" value="${attr(pick(q, "Sort"))}">` : ""}
        ${pick(q, "Order") ? `<input type="hidden" name="order" value="${attr(pick(q, "Order"))}">` : ""}
        <div class="col-12 col-md-6 col-xl-3"><label class="form-label mb-1">关键词</label><input class="form-control" type="text" name="keyword" value="${attr(pick(q, "Keyword") || "")}" placeholder="标题 / 正文 / thread_tid"></div>
        <div class="col-12 col-md-6 col-xl-2"><label class="form-label mb-1">thread_tid</label><input class="form-control" type="text" name="thread_tid" value="${attr(pick(q, "ThreadTID") || "")}"></div>
        <div class="col-12 col-md-6 col-xl-2"><label class="form-label mb-1">作者</label><input class="form-control" type="text" name="author_name" value="${attr(pick(q, "AuthorName") || "")}"></div>
        <div class="col-12 col-md-6 col-xl-3"><label class="form-label mb-1">info_hash</label><input class="form-control" type="text" name="info_hash" value="${attr(pick(q, "InfoHash") || "")}"></div>
        <div class="col-12 col-md-6 col-xl-2"><label class="form-label mb-1">来源版块</label><select class="form-select" name="source_forum_id">${renderOptions(pick(q, "SourceForumID"), [["", "全部"], ["36", "亚洲无码原创"], ["103", "高清中文字幕"]])}</select></div>
        <div class="col-12 col-md-6 col-xl-2"><label class="form-label mb-1">详情状态</label><select class="form-select" name="detail_status">${renderOptions(pick(q, "DetailStatus"), [["", "全部"], ["pending", "待抓详情"], ["failed", "详情失败"], ["fetched", "已抓详情"]])}</select></div>
        <div class="col-12 col-md-6 col-xl-2"><label class="form-label mb-1">magnet</label><select class="form-select" name="has_magnet">${renderOptions(pick(q, "HasMagnet"), [["", "全部"], ["1", "有 magnet"], ["0", "无 magnet"]])}</select></div>
        <div class="col-12 col-md-6 col-xl-3"><label class="form-label mb-1">发帖时间开始</label><input class="form-control" type="text" name="post_time_from" value="${attr(pick(q, "PostTimeFrom") || "")}" placeholder="YYYY-MM-DD"></div>
        <div class="col-12 col-md-6 col-xl-3"><label class="form-label mb-1">发帖时间结束</label><input class="form-control" type="text" name="post_time_to" value="${attr(pick(q, "PostTimeTo") || "")}" placeholder="YYYY-MM-DD"></div>
        <div class="col-12 col-md-6 col-xl-3"><label class="form-label mb-1">最近抓到开始</label><input class="form-control" type="text" name="last_seen_time_from" value="${attr(pick(q, "LastSeenTimeFrom") || "")}" placeholder="YYYY-MM-DD"></div>
        <div class="col-12 col-md-6 col-xl-3"><label class="form-label mb-1">最近抓到结束</label><input class="form-control" type="text" name="last_seen_time_to" value="${attr(pick(q, "LastSeenTimeTo") || "")}" placeholder="YYYY-MM-DD"></div>
        <div class="col-12 d-flex gap-2 next02-sh-filter-actions"><button class="btn btn-primary" type="submit">筛选</button><a class="btn btn-outline-secondary" href="/fetch-site-sehuatang-sh-list">清空</a></div>
      </form>
    </div>
  </div>`;
}

function renderSHThreadRow(row: AnyRecord, index: number): string {
  const title = readText(row, "ThreadTitle");
  const titleHTML = readText(row, "ThreadURL")
    ? `<a class="fetch-site-title-ellipsis" href="${attr(readText(row, "ThreadURL"))}" target="_blank" rel="noopener noreferrer" title="${attr(title)}">${escapeHtml(title)}</a>`
    : `<span class="fetch-site-title-ellipsis" title="${attr(title)}">${escapeHtml(title)}</span>`;
  const magnetCount = lineCount(pick(row, "MagnetText"));
  const hashCount = lineCount(pick(row, "InfoHashText"));
  const status = readText(row, "DetailStatus");
  return `<tr>
    <td class="text-nowrap">${escapeHtml(readText(row, "ThreadTID"))}</td>
    <td>${titleHTML}</td>
    <td class="text-nowrap">${escapeHtml(readText(row, "AuthorName"))}</td>
    <td class="text-nowrap">${escapeHtml(formatUnix(pick(row, "PostTime")))}</td>
    <td class="text-nowrap"><span class="next02-sh-forum-pill">${escapeHtml(readText(row, "SourceForumID"))}</span></td>
    <td class="text-nowrap"><span class="next02-sh-status-pill" data-status="${attr(status)}">${escapeHtml(status)}</span></td>
    <td class="text-nowrap"><span class="next02-sh-count-pill">${escapeHtml(magnetCount)}</span></td>
    <td class="text-nowrap"><span class="next02-sh-count-pill">${escapeHtml(hashCount)}</span></td>
    <td class="text-nowrap">${escapeHtml(formatUnix(pick(row, "LastSeenTime")))}</td>
    <td class="text-nowrap">${escapeHtml(formatUnix(pick(row, "CreatedOn")))}</td>
    <td class="text-nowrap">${escapeHtml(formatUnix(pick(row, "UpdatedOn")))}</td>
    <td class="text-nowrap"><div class="btn-group btn-group-sm" role="group">
      <button type="button" class="btn btn-outline-primary js-sh-view" data-index="${attr(index)}">内容</button>
      <a class="btn btn-outline-primary" href="/fetch-site-sehuatang-sh-detail/${attr(encodeURIComponent(readText(row, "ThreadTID")))}">详情</a>
      <button type="button" class="btn btn-outline-secondary js-sh-copy" data-index="${attr(index)}" data-field="MagnetText" ${magnetCount ? "" : "disabled"}>magnet</button>
      <button type="button" class="btn btn-outline-secondary js-sh-copy" data-index="${attr(index)}" data-field="InfoHashText" ${hashCount ? "" : "disabled"}>hash</button>
    </div></td>
  </tr>`;
}

function renderSHThreadTable(data: AnyRecord): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  const sort = asRecord(pick(data, "SortQuery"));
  return `<div class="card mb-3 next02-sh-table-panel">
    <div class="card-body">
      <div class="next02-sh-panel-head">
        <div>
          <div class="next02-sh-eyebrow">Thread Pool</div>
          <h5 class="card-title mb-0">帖子数据</h5>
        </div>
        <div class="next02-sh-table-meta">${escapeHtml(rows.length)} 行当前页</div>
      </div>
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0">
          <thead><tr>
            <th>${sortLabel(pick(sort, "ByThreadTID"), "thread_tid")}</th>
            <th>帖子标题</th>
            <th>作者</th>
            <th>${sortLabel(pick(sort, "ByPostTime"), "发布时间")}</th>
            <th>版块</th>
            <th>详情状态</th>
            <th>magnet</th>
            <th>hash</th>
            <th>${sortLabel(pick(sort, "ByLastSeenTime"), "最近抓到")}</th>
            <th>${sortLabel(pick(sort, "ByCreatedOn"), "入库时间")}</th>
            <th>${sortLabel(pick(sort, "ByUpdatedOn"), "更新时间")}</th>
            <th class="text-nowrap">操作</th>
          </tr></thead>
          <tbody>${rows.length ? rows.map(renderSHThreadRow).join("") : `<tr><td colspan="12"><div class="fetch-site-empty-state">当前筛选下没有 Sehuatang SH 帖子数据。</div></td></tr>`}</tbody>
        </table>
      </div>
    </div>
  </div>`;
}

function renderSHRuns(data: AnyRecord): string {
  const runs = asArray<AnyRecord>(pick(data, "RecentRuns"));
  return `<div class="card next02-sh-runs-panel">
    <div class="card-body">
      <div class="next02-sh-panel-head">
        <div>
          <div class="next02-sh-eyebrow">Runs</div>
          <h5 class="card-title mb-0">最近抓取记录</h5>
        </div>
        <div class="next02-sh-table-meta">${escapeHtml(runs.length)} 条</div>
      </div>
      <div class="table-responsive">
        <table class="table table-sm align-middle mb-0">
          <thead><tr><th>run_id</th><th>模式</th><th>关键词</th><th>版块</th><th>状态</th><th>发现</th><th>命中</th><th>跳过版块</th><th>跳过已存在</th><th>入库</th><th>失败</th><th>开始</th><th>结束</th><th>错误</th></tr></thead>
          <tbody>${runs.length ? runs.map((run) => `<tr>
            <td>${escapeHtml(pick(run, "Id") || "")}</td>
            <td>${escapeHtml(readText(run, "FetchMode"))}</td>
            <td>${escapeHtml(readText(run, "Keyword"))}</td>
            <td>${escapeHtml(readText(run, "TargetForumIDs"))}</td>
            <td><span class="badge fetch-site-badge">${escapeHtml(readText(run, "Status"))}</span></td>
            <td>${escapeHtml(valueOf(pick(run, "SeenCount"), 0))}</td>
            <td>${escapeHtml(valueOf(pick(run, "MatchedCount"), 0))}</td>
            <td>${escapeHtml(valueOf(pick(run, "SkippedForumCount"), 0))}</td>
            <td>${escapeHtml(valueOf(pick(run, "SkippedExistingCount"), 0))}</td>
            <td>${escapeHtml(valueOf(pick(run, "FetchedCount"), 0))}</td>
            <td>${escapeHtml(valueOf(pick(run, "FailedCount"), 0))}</td>
            <td class="text-nowrap">${escapeHtml(formatUnix(pick(run, "StartedOn")))}</td>
            <td class="text-nowrap">${escapeHtml(formatUnix(pick(run, "FinishedOn")))}</td>
            <td>${readText(run, "LastError") ? `<details><summary>错误</summary><div class="mt-2">${escapeHtml(readText(run, "LastError"))}</div></details>` : "-"}</td>
          </tr>`).join("") : `<tr><td colspan="14"><div class="fetch-site-empty-state">暂无抓取记录。</div></td></tr>`}</tbody>
        </table>
      </div>
    </div>
  </div>`;
}

function renderSHDetailModal(): string {
  return `<div class="modal fade" id="shThreadDetailModal" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-xl modal-dialog-scrollable"><div class="modal-content">
      <div class="modal-header"><h5 class="modal-title" id="shThreadDetailTitle">帖子内容</h5><button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="关闭"></button></div>
      <div class="modal-body"><div class="mb-3 small text-muted" id="shThreadDetailMeta"></div><h6>正文</h6><pre class="bg-light border rounded p-3 small" id="shThreadDetailContent"></pre><h6>magnet</h6><pre class="bg-light border rounded p-3 small" id="shThreadDetailMagnet"></pre><h6>info_hash</h6><pre class="bg-light border rounded p-3 small mb-0" id="shThreadDetailHash"></pre></div>
    </div></div>
  </div>`;
}

function renderSHPage(data: AnyRecord): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  const total = valueOf(pick(data, "Total"), 0);
  const magnetTotal = rows.reduce((sum, row) => sum + lineCount(pick(row, "MagnetText")), 0);
  const hashTotal = rows.reduce((sum, row) => sum + lineCount(pick(row, "InfoHashText")), 0);
  return `<div id="sehuatangSHListPage" class="next02-sh-list-rewrite">
    <div class="next02-sh-list-hero">
      <div class="next02-sh-list-hero-main">
        <div class="next02-sh-eyebrow">Sehuatang SH Pool</div>
        <h4 class="mb-1">${escapeHtml(pick(data, "PageTitle") || "Sehuatang SH 帖子列表")}</h4>
        <p class="text-muted mb-0">${escapeHtml(pick(data, "PageNote") || "")}</p>
        <div class="next02-sh-kpi-strip">
          <span><strong>${escapeHtml(total)}</strong><em>总数</em></span>
          <span><strong>${escapeHtml(rows.length)}</strong><em>当前页</em></span>
          <span><strong>${escapeHtml(magnetTotal)}</strong><em>magnet</em></span>
          <span><strong>${escapeHtml(hashTotal)}</strong><em>hash</em></span>
        </div>
      </div>
      <div class="next02-sh-list-actions">
        <a class="btn btn-outline-primary btn-sm" href="/fetch-site-sehuatang-sh-list">SehuatangSH 列表</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "OldListURL") || "/fetch-site-sehuatang-list")}">旧磁力列表</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "TaskPageURL") || "/triggers/fetch-sehuatang-sh-thread-list")}">列表抓取</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "DetailTaskPageURL") || "/triggers/fetch-sehuatang-sh-thread-detail")}">详情补全</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "TasksPageURL") || "/crawler/tasks")}">任务列表</a>
      </div>
    </div>
    <div id="sehuatangSHListMsg" class="alert d-none mb-3" role="alert"></div>
    <div class="next02-sh-list-toolbar">
      <div class="fetch-site-summary-card"><span class="fetch-site-summary-label">SH 帖子总数</span><span class="fetch-site-summary-value">${escapeHtml(total)}</span></div>
      ${rows.length ? pageLinks(pick(data, "PageInfo"), "SehuatangSH") : ""}
    </div>
    ${renderSHFilterForm(data)}
    ${renderSHThreadTable(data)}
    ${renderSHRuns(data)}
  </div>
  ${renderSHDetailModal()}`;
}

function formatSHDetailStatus(status: unknown): string {
  const current = String(status || "").trim();
  if (current === "pending") return "待抓详情";
  if (current === "fetched") return "已抓详情";
  if (current === "failed") return "详情失败";
  return text(current);
}

function renderInfoCard(thread: AnyRecord): string {
  const item = (label: string, value: string) => `<div class="col-12 col-md-4 col-xl-2"><div class="text-muted small">${escapeHtml(label)}</div><div class="fw-semibold">${value}</div></div>`;
  return `<div class="card mb-3">
    <div class="card-body">
      <div class="row g-3">
        ${item("thread_tid", escapeHtml(readText(thread, "ThreadTID")))}
        ${item("来源版块", escapeHtml(readText(thread, "SourceForumID")))}
        ${item("来源页", readText(thread, "SourcePageURL") ? `<a href="${attr(readText(thread, "SourcePageURL"))}" target="_blank" rel="noopener noreferrer">${escapeHtml(readText(thread, "SourcePage"))}</a>` : escapeHtml(readText(thread, "SourcePage")))}
        ${item("详情状态", escapeHtml(formatSHDetailStatus(pick(thread, "DetailStatus"))))}
        ${item("作者", escapeHtml(readText(thread, "AuthorName")))}
        ${item("发布时间", escapeHtml(formatUnix(pick(thread, "PostTime"))))}
        ${item("最近抓到", escapeHtml(formatUnix(pick(thread, "LastSeenTime"))))}
        ${item("详情抓取时间", escapeHtml(formatUnix(pick(thread, "DetailFetchedOn"))))}
        ${item("详情失败时间", escapeHtml(formatUnix(pick(thread, "DetailFailedOn"))))}
        ${item("入库时间", escapeHtml(formatUnix(pick(thread, "CreatedOn"))))}
        ${item("更新时间", escapeHtml(formatUnix(pick(thread, "UpdatedOn"))))}
        <div class="col-12"><div class="text-muted small">content_hash</div><div class="font-monospace small">${escapeHtml(readText(thread, "ContentHash"))}</div></div>
        ${readText(thread, "DetailError") ? `<div class="col-12"><div class="text-muted small">详情错误</div><div class="small text-danger">${escapeHtml(readText(thread, "DetailError"))}</div></div>` : ""}
      </div>
    </div>
  </div>`;
}

function splitLines(value: unknown): string[] {
  return String(value || "").split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
}

function renderCopyList(title: string, items: string[], copyAllLabel: string): string {
  return `<div class="card mb-3">
    <div class="card-body">
      <div class="d-flex align-items-center justify-content-between gap-2 mb-3">
        <h5 class="card-title mb-0">${escapeHtml(title)}</h5>
        <button type="button" class="btn btn-sm btn-outline-secondary js-copy-value" data-copy-value="${attr(items.join("\n"))}" ${items.length ? "" : "disabled"}>${escapeHtml(copyAllLabel)}</button>
      </div>
      ${items.length ? `<div class="list-group list-group-flush">${items.map((item) => `<div class="list-group-item px-0 d-flex align-items-start gap-2"><code class="flex-grow-1 text-break">${escapeHtml(item)}</code><button type="button" class="btn btn-sm btn-outline-secondary js-copy-value" data-copy-value="${attr(item)}">复制</button></div>`).join("")}</div>` : `<div class="text-muted small">暂无数据。</div>`}
    </div>
  </div>`;
}

function renderContent(thread: AnyRecord): string {
  return `<div class="card mb-3">
    <div class="card-body">
      <div class="d-flex align-items-center justify-content-between gap-2 mb-3">
        <h5 class="card-title mb-0">正文</h5>
        <button type="button" class="btn btn-sm btn-outline-secondary js-copy-value" data-copy-value="${attr(readText(thread, "RawContentText"))}" ${readText(thread, "RawContentText") ? "" : "disabled"}>复制正文</button>
      </div>
      <pre class="bg-light border rounded p-3 small mb-0">${escapeHtml(readText(thread, "RawContentText"))}</pre>
    </div>
  </div>`;
}

function renderFetchItems(items: unknown): string {
  const rows = asArray<AnyRecord>(items);
  return `<div class="card">
    <div class="card-body">
      <h5 class="card-title mb-3">相关抓取记录</h5>
      <div class="table-responsive">
        <table class="table table-sm align-middle mb-0">
          <thead><tr><th>run_id</th><th>模式</th><th>来源版块</th><th>来源页</th><th>状态</th><th>thread_row_id</th><th>更新时间</th><th>错误</th></tr></thead>
          <tbody>${rows.length ? rows.map((item) => `<tr>
            <td>${escapeHtml(pick(item, "RunID") || "")}</td>
            <td>${escapeHtml(readText(item, "FetchMode"))}</td>
            <td>${escapeHtml(readText(item, "SourceForumID"))}</td>
            <td>${readText(item, "SourcePageURL") ? `<a href="${attr(readText(item, "SourcePageURL"))}" target="_blank" rel="noopener noreferrer">${escapeHtml(readText(item, "SourcePage"))}</a>` : escapeHtml(readText(item, "SourcePage"))}</td>
            <td><span class="badge fetch-site-badge">${escapeHtml(readText(item, "Status"))}</span></td>
            <td>${escapeHtml(pick(item, "ThreadRowID") || "")}</td>
            <td class="text-nowrap">${escapeHtml(formatUnix(pick(item, "UpdatedOn")))}</td>
            <td>${readText(item, "Error") ? `<details><summary>错误</summary><div class="mt-2">${escapeHtml(readText(item, "Error"))}</div></details>` : "-"}</td>
          </tr>`).join("") : `<tr><td colspan="8"><div class="fetch-site-empty-state">暂无相关抓取记录。</div></td></tr>`}</tbody>
        </table>
      </div>
    </div>
  </div>`;
}

function renderSHDetailPage(data: AnyRecord): string {
  const thread = asRecord(pick(data, "Thread"));
  const magnets = splitLines(pick(thread, "MagnetText"));
  const hashes = splitLines(pick(thread, "InfoHashText"));
  return `<div id="shDetailPage">
    <div class="d-flex flex-column flex-xl-row align-items-xl-start justify-content-between gap-3 mb-3">
      <div>
        <h4 class="mb-1">${escapeHtml(readText(thread, "ThreadTitle") || pick(data, "PageTitle") || "Sehuatang SH 帖子详情")}</h4>
        <p class="text-muted mb-0">${escapeHtml(pick(data, "PageTitle") || "")}</p>
      </div>
      <div class="d-flex flex-wrap gap-2">
        <a class="btn btn-outline-primary btn-sm" href="${attr(pick(data, "ListPageURL") || "/fetch-site-sehuatang-sh-list")}">返回列表</a>
        ${readText(thread, "ThreadURL") ? `<a class="btn btn-outline-secondary btn-sm" href="${attr(readText(thread, "ThreadURL"))}" target="_blank" rel="noopener noreferrer">打开原帖</a>` : ""}
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "TaskPageURL") || "/triggers/fetch-sehuatang-sh-thread-detail")}">详情补全</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "ListTaskPageURL") || "/triggers/fetch-sehuatang-sh-thread-list")}">列表抓取</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "TasksPageURL") || "/crawler/tasks")}">任务列表</a>
      </div>
    </div>
    <div id="shDetailMsg" class="alert d-none mb-3" role="alert"></div>
    ${renderInfoCard(thread)}
    ${renderContent(thread)}
    ${renderCopyList("magnet", magnets, "复制全部 magnet")}
    ${renderCopyList("info_hash", hashes, "复制全部 hash")}
    ${renderFetchItems(pick(data, "FetchItems"))}
  </div>`;
}

function renderFallbackTable(data: AnyRecord): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows", "Items", "items"));
  const first = rows[0] || {};
  const headers = Object.keys(first);
  return `<div class="card"><div class="card-body"><div class="table-responsive"><table class="table table-hover align-middle mb-0">
    <thead><tr>${headers.map((key) => `<th>${escapeHtml(key)}</th>`).join("")}</tr></thead>
    <tbody>${rows.map((row) => `<tr>${headers.map((key) => `<td>${escapeHtml(readText(row, key))}</td>`).join("")}</tr>`).join("")}</tbody>
  </table></div></div></div>`;
}

function readMovieTarget(button: Element): { movieJavID: string; movieName: string } {
  const row = button.closest("tr[data-movie-jav-id]");
  if (row) {
    return {
      movieJavID: String(row.getAttribute("data-movie-jav-id") || row.getAttribute("data-movie-route-key") || row.getAttribute("data-movie-name") || "").trim(),
      movieName: String(row.getAttribute("data-movie-name") || "").trim()
    };
  }
  const container = button.closest("[data-movie-jav-id][data-movie-name]");
  return {
    movieJavID: String(container?.getAttribute("data-movie-jav-id") || container?.getAttribute("data-movie-route-key") || container?.getAttribute("data-movie-name") || "").trim(),
    movieName: String(container?.getAttribute("data-movie-name") || "").trim()
  };
}

async function startSiteTask(taskType: string, movieJavID: string, movieName: string) {
  setTriggerButtonsLoading(true);
  try {
    await sendJSON("/api/crawler/jobs/start", "POST", {
      task_type: taskType,
      movie_jav_id: movieJavID,
      movie_name: movieName
    });
    showMsg(`已开启${taskType === "spider_fetch_javbus_resources" ? "JavBus" : "Sukebei"}抓取：${movieName}`, true);
  } catch (err) {
    showMsg(err instanceof Error ? err.message : "启动失败", false);
  } finally {
    setTriggerButtonsLoading(false);
  }
}

function buildFilteredTaskPayload(button: Element): Record<string, unknown> {
  const page = document.getElementById("fetchSiteListPage");
  const taskType = String(page?.getAttribute("data-filtered-task-type") || page?.getAttribute("data-filtered-sukebei-task-type") || "").trim();
  if (!taskType) {
    throw new Error("缺少筛选任务类型，无法触发");
  }
  const formID = String(page?.getAttribute("data-filter-form-id") || "").trim();
  const form = (formID ? document.getElementById(formID) : null)
    || button.closest(".card")?.querySelector("form")
    || document.querySelector("form[data-fetch-site-filter-form]");
  if (!(form instanceof HTMLFormElement)) {
    throw new Error("缺少筛选表单，无法触发");
  }
  const formData = new FormData(form);
  const payload: Record<string, unknown> = { task_type: taskType };
  [
    "keyword", "last_fetch_from", "last_fetch_to",
    "release_date_from", "release_date_to",
    "film_birth_from", "film_birth_to",
    "media_birth_from", "media_birth_to",
    "owned", "mowned", "sort", "order"
  ].forEach((key) => {
    const value = String(formData.get(key) || "").trim();
    if (value) {
      payload[key] = value;
    }
  });
  const statuses = formData.getAll("statuses").map((value) => String(value || "").trim()).filter(Boolean);
  if (statuses.length) {
    payload.statuses = statuses;
  }
  const triggerSortKey = String(formData.get("trigger_sort_key") || "").trim();
  if (triggerSortKey) {
    const [triggerSort, triggerOrder] = triggerSortKey.split(":");
    if (triggerSort) {
      payload.trigger_sort = triggerSort;
    }
    if (triggerOrder) {
      payload.trigger_order = triggerOrder;
    }
  }
  return payload;
}

async function startFilteredTask(button: Element) {
  let body: Record<string, unknown>;
  try {
    body = buildFilteredTaskPayload(button);
  } catch (err) {
    showMsg(err instanceof Error ? err.message : "构造筛选任务失败", false);
    return;
  }
  setTriggerButtonsLoading(true);
  try {
    const data = asRecord(await sendJSON("/api/crawler/jobs/start", "POST", body));
    const jobID = readText(data, "job_id", "JobID");
    const page = document.getElementById("fetchSiteListPage");
    const taskPageURL = String(page?.getAttribute("data-task-page-url") || "/triggers/fetch-site").trim();
    if (jobID) {
      window.location.href = `${taskPageURL}?job_id=${encodeURIComponent(jobID)}`;
      return;
    }
    showMsg("任务已创建，但缺少 job_id", false);
  } catch (err) {
    showMsg(err instanceof Error ? err.message : "启动失败", false);
  } finally {
    setTriggerButtonsLoading(false);
  }
}

function refreshStatusClearButton(form: Element | null) {
  const clearButton = document.querySelector("[data-status-clear]");
  if (!clearButton) {
    return;
  }
  const checkedCount = form?.querySelectorAll("input[name='statuses']:checked").length || 0;
  clearButton.classList.toggle("active", checkedCount === 0);
}

function closeModal(modal: HTMLElement | null) {
  if (!modal) {
    return;
  }
  const activeElement = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  if (activeElement && modal.contains(activeElement)) {
    const returnFocus = modal.id === "sehuatangDateModal"
      ? document.getElementById("btnEditShtTime") as HTMLElement | null
      : null;
    if (returnFocus) {
      returnFocus.focus();
    } else {
      activeElement.blur();
    }
  }
  modal.classList.remove("show");
  modal.style.display = "none";
  modal.setAttribute("aria-hidden", "true");
}

function openModal(modal: HTMLElement | null) {
  if (!modal) {
    return;
  }
  modal.classList.add("show");
  modal.style.display = "block";
  modal.removeAttribute("aria-hidden");
}

function showInlineMsg(el: HTMLElement | null, message: string, ok: boolean) {
  if (!el) {
    return;
  }
  el.textContent = message;
  el.className = `alert mb-0 ${ok ? "alert-success" : "alert-danger"}`;
}

function hideInlineMsg(el: HTMLElement | null) {
  if (!el) {
    return;
  }
  el.textContent = "";
  el.className = "alert d-none mb-0";
}

function normalizeDateInputValue(raw: unknown): string {
  const digits = String(raw || "").replace(/\D/g, "").slice(0, 8);
  if (digits.length <= 4) {
    return digits;
  }
  if (digits.length <= 6) {
    return `${digits.slice(0, 4)}-${digits.slice(4)}`;
  }
  return `${digits.slice(0, 4)}-${digits.slice(4, 6)}-${digits.slice(6)}`;
}

function isValidDateValue(raw: unknown): boolean {
  return /^[0-9]{4}-[0-9]{2}-[0-9]{2}$/.test(String(raw || "").trim());
}

function todayDateString(): string {
  const now = new Date();
  const year = String(now.getFullYear());
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function openSehuatangDateModal(button: Element) {
  const modal = document.getElementById("sehuatangDateModal");
  const title = document.getElementById("sehuatangDateModalTitle");
  const keyInput = document.getElementById("sehuatangDateItemKey") as HTMLInputElement | null;
  const dateInput = document.getElementById("sehuatangDateInput") as HTMLInputElement | null;
  const dateMsg = document.getElementById("sehuatangDateMsg");
  const label = String(button.getAttribute("data-item-label") || "日期").trim();
  const key = String(button.getAttribute("data-item-key") || "").trim();
  const value = String(button.getAttribute("data-item-value") || "").trim();
  if (title) {
    title.textContent = `修改${label}`;
  }
  if (keyInput) {
    keyInput.value = key;
  }
  if (dateInput) {
    dateInput.value = value;
  }
  hideInlineMsg(dateMsg);
  openModal(modal);
}

async function submitSehuatangDate(form: HTMLFormElement) {
  const keyInput = form.querySelector<HTMLInputElement>("#sehuatangDateItemKey");
  const dateInput = form.querySelector<HTMLInputElement>("#sehuatangDateInput");
  const submitButton = form.querySelector<HTMLButtonElement>("#sehuatangDateSubmit");
  const dateMsg = form.querySelector<HTMLElement>("#sehuatangDateMsg");
  const itemKey = String(keyInput?.value || "").trim();
  const itemValue = normalizeDateInputValue(dateInput?.value || "");
  if (dateInput) {
    dateInput.value = itemValue;
  }
  if (!itemKey) {
    showInlineMsg(dateMsg, "缺少日期 key", false);
    return;
  }
  if (!itemValue) {
    showInlineMsg(dateMsg, "请选择日期", false);
    return;
  }
  if (!isValidDateValue(itemValue)) {
    showInlineMsg(dateMsg, "日期格式必须为 YYYY-MM-DD", false);
    return;
  }
  if (submitButton) {
    submitButton.disabled = true;
  }
  try {
    const data = asRecord(await sendJSON("/api/w-kv/date", "POST", {
      item_key: itemKey,
      item_value: itemValue
    }));
    const savedValue = readText(data, "item_value", "ItemValue") || itemValue;
    const editButton = document.getElementById("btnEditShtTime");
    if (editButton) {
      editButton.setAttribute("data-item-value", savedValue);
      editButton.textContent = `SHT时间：${savedValue}`;
    }
    showMsg("SHT时间已更新", true);
    hideInlineMsg(dateMsg);
    closeModal(document.getElementById("sehuatangDateModal"));
  } catch (err) {
    showInlineMsg(dateMsg, err instanceof Error ? err.message : "保存失败", false);
  } finally {
    if (submitButton) {
      submitButton.disabled = false;
    }
  }
}

function favoriteButtonTarget(button: Element, albumName?: string): { movieJavID: string; sourceType: string; sourceRowID: number; favorited: boolean; albumName: string } {
  const target = readMovieTarget(button);
  return {
    movieJavID: target.movieJavID || String(button.getAttribute("data-movie-jav-id") || "").trim(),
    sourceType: String(button.getAttribute("data-source-type") || "").trim(),
    sourceRowID: Number(button.getAttribute("data-source-row-id") || 0),
    favorited: String(button.getAttribute("data-favorited") || "") === "1",
    albumName: albumName || String(button.getAttribute("data-album-name") || "").trim() || "下载中"
  };
}

async function switchFavorite(button: Element, albumName?: string) {
  const target = favoriteButtonTarget(button, albumName);
  if (!target.movieJavID || !target.sourceType || !target.sourceRowID) {
    showMsg(`缺少${target.albumName}参数，无法执行`, false);
    return;
  }
  const htmlButton = button instanceof HTMLButtonElement ? button : null;
  if (htmlButton) {
    htmlButton.disabled = true;
  }
  try {
    const data = asRecord(await sendJSON(`/api/movie/${encodeURIComponent(target.movieJavID)}/album-item`, target.favorited ? "DELETE" : "POST", {
      movie_jav_id: target.movieJavID,
      source_type: target.sourceType,
      source_row_id: target.sourceRowID,
      album_name: target.albumName
    }));
    const favorited = typeof pick(data, "favorited", "Favorited") === "boolean" ? Boolean(pick(data, "favorited", "Favorited")) : !target.favorited;
    const nextAlbum = readText(data, "album_name", "AlbumName") || target.albumName;
    button.setAttribute("data-favorited", favorited ? "1" : "0");
    button.setAttribute("data-album-name", nextAlbum);
    button.textContent = favorited ? `已在${nextAlbum}` : nextAlbum;
    showMsg(readText(data, "message", "Message") || "操作成功", true);
  } catch (err) {
    showMsg(err instanceof Error ? err.message : `${target.albumName}操作失败`, false);
  } finally {
    if (htmlButton) {
      htmlButton.disabled = false;
    }
  }
}

async function copyToClipboard(value: string) {
  const content = String(value || "").trim();
  if (!content) {
    showMsg("没有可复制的内容", false);
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

async function copyValue(value: string) {
  try {
    await copyToClipboard(value);
    showMsg("已复制", true);
  } catch (err) {
    showMsg(err instanceof Error ? `复制失败：${err.message}` : "复制失败", false);
  }
}

function currentSHRows(): AnyRecord[] {
  return asArray<AnyRecord>(pick(payload.value, "Rows"));
}

function openSHThreadModal(index: number) {
  const row = currentSHRows()[index] || {};
  const title = document.getElementById("shThreadDetailTitle");
  const meta = document.getElementById("shThreadDetailMeta");
  const content = document.getElementById("shThreadDetailContent");
  const magnet = document.getElementById("shThreadDetailMagnet");
  const hash = document.getElementById("shThreadDetailHash");
  if (title) {
    title.textContent = readText(row, "ThreadTitle") || "帖子内容";
  }
  if (meta) {
    meta.textContent = [
      readText(row, "ThreadTID") ? `tid=${readText(row, "ThreadTID")}` : "",
      readText(row, "AuthorName") ? `作者=${readText(row, "AuthorName")}` : "",
      pick(row, "PostTime") ? `发布时间=${formatUnix(pick(row, "PostTime"))}` : "",
      readText(row, "ThreadURL") ? `原帖=${readText(row, "ThreadURL")}` : ""
    ].filter(Boolean).join(" | ");
  }
  if (content) {
    content.textContent = readText(row, "RawContentText", "raw_content_text");
  }
  if (magnet) {
    magnet.textContent = readText(row, "MagnetText", "magnet_text");
  }
  if (hash) {
    hash.textContent = readText(row, "InfoHashText", "info_hash_text");
  }
  openModal(document.getElementById("shThreadDetailModal"));
}

function handlePageClick(event: MouseEvent) {
  const target = event.target instanceof Element ? event.target : null;
  if (!target) {
    return;
  }
  const dismiss = target.closest("[data-bs-dismiss='modal']");
  if (dismiss) {
    closeModal(dismiss.closest(".modal") as HTMLElement | null);
    return;
  }
  const clearStatusButton = target.closest("[data-status-clear]");
  if (clearStatusButton) {
    const form = clearStatusButton.closest("form");
    form?.querySelectorAll<HTMLInputElement>("input[name='statuses']").forEach((input) => {
      input.checked = false;
    });
    refreshStatusClearButton(form);
    return;
  }
  const filteredButton = target.closest(".js-trigger-filtered-fetch-site, .js-trigger-filtered-sukebei");
  if (filteredButton) {
    void startFilteredTask(filteredButton);
    return;
  }
  const openDateButton = target.closest("#btnEditShtTime");
  if (openDateButton) {
    openSehuatangDateModal(openDateButton);
    return;
  }
  const todayButton = target.closest("#sehuatangDateTodayBtn");
  if (todayButton) {
    const input = document.getElementById("sehuatangDateInput") as HTMLInputElement | null;
    if (input) {
      input.value = todayDateString();
    }
    hideInlineMsg(document.getElementById("sehuatangDateMsg"));
    return;
  }
  const javbusButton = target.closest(".js-trigger-javbus");
  const sukebeiButton = target.closest(".js-trigger-sukebei");
  if (javbusButton || sukebeiButton) {
    const button = javbusButton || sukebeiButton;
    const movie = readMovieTarget(button);
    if (!movie.movieJavID || !movie.movieName) {
      showMsg("缺少影片参数，无法触发任务", false);
      return;
    }
    void startSiteTask(javbusButton ? "spider_fetch_javbus_resources" : "spider_fetch_sukebei_resources", movie.movieJavID, movie.movieName);
    return;
  }
  const favoriteButton = target.closest(".js-add-favorite");
  if (favoriteButton) {
    void switchFavorite(favoriteButton);
    return;
  }
  const shViewButton = target.closest(".js-sh-view");
  if (shViewButton) {
    openSHThreadModal(Number(shViewButton.getAttribute("data-index") || -1));
    return;
  }
  const shCopyButton = target.closest(".js-sh-copy");
  if (shCopyButton) {
    const row = currentSHRows()[Number(shCopyButton.getAttribute("data-index") || -1)] || {};
    const field = String(shCopyButton.getAttribute("data-field") || "").trim();
    void copyValue(readText(row, field));
    return;
  }
  const copyButton = target.closest(".js-copy-value");
  if (copyButton) {
    void copyValue(String(copyButton.getAttribute("data-copy-value") || ""));
    return;
  }
  const otherAlbumButton = target.closest(".js-open-other-album-modal");
  if (otherAlbumButton) {
    const movie = readMovieTarget(otherAlbumButton);
    otherAlbumTarget.value = {
      movieJavID: movie.movieJavID,
      sourceType: String(otherAlbumButton.getAttribute("data-source-type") || "").trim(),
      sourceRowID: Number(otherAlbumButton.getAttribute("data-source-row-id") || 0)
    };
    openModal(document.getElementById("sehuatangOtherAlbumModal"));
    return;
  }
  const albumOption = target.closest(".js-other-album-option");
  if (albumOption && otherAlbumTarget.value) {
    const albumName = String(albumOption.getAttribute("data-album-name") || "").trim();
    const fakeButton = document.createElement("button");
    fakeButton.setAttribute("data-source-type", otherAlbumTarget.value.sourceType);
    fakeButton.setAttribute("data-source-row-id", String(otherAlbumTarget.value.sourceRowID));
    fakeButton.setAttribute("data-album-name", albumName);
    fakeButton.setAttribute("data-movie-jav-id", otherAlbumTarget.value.movieJavID);
    void switchFavorite(fakeButton, albumName);
    closeModal(document.getElementById("sehuatangOtherAlbumModal"));
  }
}

function handlePageSubmit(event: SubmitEvent) {
  const form = event.target instanceof HTMLFormElement ? event.target : null;
  if (!form || form.id !== "sehuatangDateForm") {
    return;
  }
  event.preventDefault();
  void submitSehuatangDate(form);
}
</script>
