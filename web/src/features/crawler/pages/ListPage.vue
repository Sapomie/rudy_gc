<template>
  <PageShell
    class="fetch-site-list-page-shell"
    :class="{ 'sc-events-page-shell': viewType === 'fetch-site-javbus-list' || viewType === 'fetch-site-sukebei-list' || viewType === 'fetch-site-sehuatang-sh-list' || viewType === 'fetch-site-sehuatang-sh-detail' }"
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
      @change="handlePageChange"
      v-html="pageHTML"
    ></div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/shared/api/request";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import type { AnyRecord } from "@/shared/types/gc";
import { asArray, asRecord, normalizeHref, pick, text as readText } from "@/shared/utils/format";
import { loadSHForumOptions, shForumFilterOptions, shForumName, shForumNames, type SHForumOption } from "@/shared/utils/shForums";

const route = useRoute();
const payload = ref<AnyRecord>({});
const shForums = ref<SHForumOption[]>([]);
const loading = ref(false);
const error = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const needsSHForums = route.path === "/shpool" || route.path.startsWith("/shthread/");
    const [pagePayload, forumOptions] = await Promise.all([
      getJSON(apiPathForRoute(route)),
      needsSHForums ? loadSHForumOptions() : Promise.resolve([])
    ]);
    payload.value = asRecord(pagePayload);
    shForums.value = forumOptions;
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
    return "next02-fetch-site-javbus-list-rewrite container-fluid next02-sc-pink-theme next02-sc-events-rewrite";
  }
  if (viewType.value === "fetch-site-sukebei-list") {
    return "next02-fetch-site-sukebei-list-rewrite container-fluid next02-sc-pink-theme next02-sc-events-rewrite";
  }
  if (viewType.value === "fetch-site-sehuatang-sh-list") {
    return "next02-fetch-site-sehuatang-sh-list-rewrite container-fluid next02-sc-pink-theme next02-sc-events-rewrite bucket-list-standard sh-bucket-list-page";
  }
  if (viewType.value === "fetch-site-sehuatang-sh-detail") {
    return "next02-fetch-site-sehuatang-sh-detail-rewrite container-fluid next02-sc-pink-theme next02-sc-events-rewrite";
  }
  return "";
});
const pageHTML = computed(() => {
  if (viewType.value === "fetch-site-javbus-list") {
    return renderJavbusPage(payload.value);
  }
  if (viewType.value === "fetch-site-sukebei-list") {
    return renderSukebeiPage(payload.value);
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
          <a class="page-link" href="${attr(normalizeHref(readText(link, "Href") || "#"))}" ${pick(link, "Disabled") ? `tabindex="-1" aria-disabled="true"` : ""}>${escapeHtml(readText(link, "Label"))}</a>
        </li>`;
      }).join("")}
    </ul>
  </nav>`;
}

function sortLabel(link: unknown, label: string): string {
  const row = asRecord(link);
  const arrow = pick(row, "Active") ? (pick(row, "Desc") ? " ↓" : " ↑") : "";
  return `<a class="fetch-site-sort-link sc-sort-link ${pick(row, "Active") ? "active" : ""}" href="${attr(normalizeHref(readText(row, "Href") || "#"))}">${escapeHtml(label)}${arrow}</a>`;
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

function renderJavbusOwnedFilter(data: AnyRecord): string {
  const owned = asRecord(pick(data, "ownedQuery"));
  const link = (item: unknown, label: string) => {
    const current = asRecord(item);
    return `<a href="${attr(normalizeHref(readText(current, "Href") || "#"))}" class="btn btn-sm ${pick(current, "Active") ? "btn-primary active disabled" : "btn-outline-primary"}" aria-current="${pick(current, "Active") ? "true" : "false"}">${escapeHtml(label)}</a>`;
  };
  return `<div class="album-quick-filter-group" aria-label="WMedia 筛选">
    ${link(pick(owned, "MediaAll"), "MAll")}
    ${link(pick(owned, "MediaOwned"), "MOwned")}
    ${link(pick(owned, "MediaNotOwned"), "MNotOwned")}
  </div>`;
}

function renderJavbusFilterForm(data: AnyRecord): string {
  const q = asRecord(pick(data, "Query"));
  return `<form class="sc-event-filter-form sehuatang-filter-form fetch-pool-filter-form" method="get" action="/javbuslist" id="fetchSiteJavbusFilterForm" data-fetch-site-filter-form="1">
    <input type="hidden" name="ps" value="${attr(pick(q, "PageSize") || "")}">
    ${pick(q, "Sort") ? `<input type="hidden" name="sort" value="${attr(pick(q, "Sort"))}">` : ""}
    ${pick(q, "Order") ? `<input type="hidden" name="order" value="${attr(pick(q, "Order"))}">` : ""}
    ${pick(q, "MediaOwned") ? `<input type="hidden" name="mowned" value="${attr(pick(q, "MediaOwned"))}">` : ""}
    <div class="sehuatang-filter-cluster sehuatang-filter-query" aria-label="检索条件">
      <label class="sc-event-filter-field sehuatang-filter-field is-keyword" for="javbusKeywordInput">
        <span>影片编号 / JavID</span>
        <input class="form-control" id="javbusKeywordInput" type="text" name="keyword" value="${attr(pick(q, "Keyword") || "")}" placeholder="movie_name 或 movie_jav_id">
      </label>
      <div class="sc-event-filter-field sehuatang-filter-field is-fetch-pool-status">
        <span>JavBus 状态</span>
        <div class="album-quick-filter-group fetch-site-status-group">${renderFetchSiteStatusGroup(data, "javbus")}</div>
      </div>
      <label class="sc-event-filter-field sehuatang-filter-field is-fetch-pool-sort" for="javbusTriggerSortSelect">
        <span>触发抓取顺序</span>
        <select class="form-select" id="javbusTriggerSortSelect" name="trigger_sort_key">${triggerSortOptions(q)}</select>
      </label>
    </div>
    <div class="sehuatang-filter-cluster sehuatang-filter-date-range" aria-label="最后抓取日期范围">
      <span class="sehuatang-filter-cluster-label">最后抓取</span>
      <label class="sc-event-filter-field sehuatang-filter-field is-date" for="javbusLastFetchFromInput"><span>开始</span><input class="form-control" id="javbusLastFetchFromInput" type="text" name="last_fetch_from" value="${attr(pick(q, "LastFetchFrom") || "")}"></label>
      <span class="sehuatang-filter-range-separator" aria-hidden="true">—</span>
      <label class="sc-event-filter-field sehuatang-filter-field is-date" for="javbusLastFetchToInput"><span>结束</span><input class="form-control" id="javbusLastFetchToInput" type="text" name="last_fetch_to" value="${attr(pick(q, "LastFetchTo") || "")}"></label>
    </div>
    <div class="sehuatang-filter-cluster sehuatang-filter-quick-actions" aria-label="快捷筛选与操作">
      <div class="sc-event-filter-field sehuatang-filter-field is-quick-tag">
        <span>WMedia</span>
        ${renderJavbusOwnedFilter(data)}
      </div>
      <div class="sehuatang-filter-actions">
        <button class="sc-event-filter-submit" type="submit">筛选</button>
        <a class="sc-event-filter-clear" href="/javbuslist">清空</a>
      </div>
    </div>
    <div class="fetch-pool-filter-secondary">
      <div class="sehuatang-filter-cluster sehuatang-filter-date-range" aria-label="发行日期范围">
        <span class="sehuatang-filter-cluster-label">发行时间</span>
        <label class="sc-event-filter-field sehuatang-filter-field is-date" for="javbusReleaseFromInput"><span>开始</span><input class="form-control" id="javbusReleaseFromInput" type="text" name="release_date_from" value="${attr(pick(q, "ReleaseDateFrom") || "")}"></label>
        <span class="sehuatang-filter-range-separator" aria-hidden="true">—</span>
        <label class="sc-event-filter-field sehuatang-filter-field is-date" for="javbusReleaseToInput"><span>结束</span><input class="form-control" id="javbusReleaseToInput" type="text" name="release_date_to" value="${attr(pick(q, "ReleaseDateTo") || "")}"></label>
      </div>
      <div class="sehuatang-filter-cluster sehuatang-filter-date-range" aria-label="WMedia 下载日期范围">
        <span class="sehuatang-filter-cluster-label">WMedia 下载</span>
        <label class="sc-event-filter-field sehuatang-filter-field is-date" for="javbusMediaBirthFromInput"><span>开始</span><input class="form-control" id="javbusMediaBirthFromInput" type="text" name="media_birth_from" value="${attr(pick(q, "MediaBirthFrom") || "")}"></label>
        <span class="sehuatang-filter-range-separator" aria-hidden="true">—</span>
        <label class="sc-event-filter-field sehuatang-filter-field is-date" for="javbusMediaBirthToInput"><span>结束</span><input class="form-control" id="javbusMediaBirthToInput" type="text" name="media_birth_to" value="${attr(pick(q, "MediaBirthTo") || "")}"></label>
      </div>
    </div>
  </form>`;
}

function renderJavbusTime(value: unknown): string {
  const current = text(value);
  if (current === "-") {
    return "-";
  }
  const [date, time = ""] = current.split(" ", 2);
  return `<span class="sc-event-time"><span class="sc-event-date">${escapeHtml(date)}</span>${time ? `<span class="sc-event-hour">${escapeHtml(time)}</span>` : ""}</span>`;
}

function javbusDayKey(row: AnyRecord): string {
  const current = readText(row, "LastFetchText");
  return current === "-" ? "" : current.slice(0, 10);
}

function javbusDayClass(rows: AnyRecord[], index: number): string {
  const day = javbusDayKey(rows[index]);
  if (!day) {
    return "sehuatang-day-neutral";
  }
  let groupIndex = 0;
  for (let rowIndex = 1; rowIndex <= index; rowIndex += 1) {
    const currentDay = javbusDayKey(rows[rowIndex]);
    const previousDay = javbusDayKey(rows[rowIndex - 1]);
    if (currentDay && previousDay && currentDay !== previousDay) {
      groupIndex += 1;
    }
  }
  return `sehuatang-day-group-${(groupIndex % 2) + 1}`;
}

function isJavbusDayStart(rows: AnyRecord[], index: number): boolean {
  if (index <= 0) {
    return false;
  }
  const currentDay = javbusDayKey(rows[index]);
  const previousDay = javbusDayKey(rows[index - 1]);
  return Boolean(currentDay && previousDay && currentDay !== previousDay);
}

function renderJavbusRow(row: AnyRecord, index: number, rows: AnyRecord[]): string {
  const movieName = readText(row, "MovieName");
  const lastFetchText = readText(row, "LastFetchText");
  const dayStartClass = isJavbusDayStart(rows, index) ? " sehuatang-day-start" : "";
  const errorHTML = readText(row, "LastError")
    ? `<details class="fetch-site-error-details"><summary>有错误</summary><div>${escapeHtml(readText(row, "LastError"))}</div></details>`
    : `<span class="javbus-no-error">无错误</span>`;
  return `<tr class="next02-sc-event-row ${javbusDayClass(rows, index)}${dayStartClass}" data-movie-jav-id="${attr(readText(row, "MovieJavID"))}" data-movie-name="${attr(movieName)}" data-fetch-day="${attr(javbusDayKey(row))}">
    <td class="sc-content-cell"><a class="sc-event-link" href="/moviedetail/${attr(encodeURIComponent(movieName))}" title="${attr(movieName)}">${escapeHtml(movieName)}</a></td>
    <td title="${attr(readText(row, "ReleaseDateText"))}">${renderJavbusTime(readText(row, "ReleaseDateText"))}</td>
    <td><div class="fetch-site-album-inline"><span class="btn btn-sm btn-outline-secondary disabled">${escapeHtml(readText(row, "FetchStatusText"))}</span></div></td>
    <td class="sc-number">${escapeHtml(valueOf(pick(row, "TryCount"), 0))}</td>
    <td title="${attr(lastFetchText)}">${renderJavbusTime(lastFetchText)}</td>
    <td class="sc-number sc-primary-value">${escapeHtml(valueOf(pick(row, "TorrentHashCount"), 0))}</td>
    <td title="${attr(readText(row, "LatestPublishText"))}">${renderJavbusTime(readText(row, "LatestPublishText"))}</td>
    <td class="fetch-pool-wmedia-cell"><div class="fetch-site-album-inline">${renderWMediaCell(row)}</div></td>
    <td class="fetch-site-error-cell">${errorHTML}</td>
    <td><div class="fetch-site-album-inline"><button type="button" class="btn btn-sm btn-outline-secondary js-trigger-javbus">触发 JavBus</button></div></td>
  </tr>`;
}

function renderJavbusTable(data: AnyRecord): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  const sort = asRecord(pick(data, "SortQuery"));
  return `<div class="table-responsive">
    <table class="table table-hover align-middle mb-0 sc-events-table sehuatang-magnet-table fetch-pool-task-table">
      <colgroup>
        <col class="fetch-pool-col-movie"><col class="fetch-pool-col-release"><col class="fetch-pool-col-status"><col class="fetch-pool-col-retry"><col class="fetch-pool-col-last-fetch">
        <col class="fetch-pool-col-hash"><col class="fetch-pool-col-latest"><col class="fetch-pool-col-wmedia"><col class="fetch-pool-col-error"><col class="fetch-pool-col-action">
      </colgroup>
      <thead><tr>
        <th class="sc-sortable-head">${sortLabel(pick(sort, "ByMovieName"), "影片")}</th>
        <th class="sc-sortable-head">${sortLabel(pick(sort, "ByReleaseDate"), "发行时间")}</th>
        <th class="sc-sortable-head">${sortLabel(pick(sort, "ByFetchStatus"), "状态")}</th>
        <th>重试</th>
        <th class="sc-sortable-head">${sortLabel(pick(sort, "ByLastFetchTime"), "最后抓取")}</th>
        <th class="sc-sortable-head">${sortLabel(pick(sort, "ByHashCount"), "Hash 数")}</th>
        <th class="sc-sortable-head">${sortLabel(pick(sort, "ByLatestPublish"), "最新日期")}</th>
        <th class="sc-sortable-head">${sortLabel(pick(sort, "ByMediaBirthTime"), "WMedia 下载时间")}</th>
        <th>错误</th>
        <th>操作</th>
      </tr></thead>
      <tbody>${rows.length ? rows.map((row, index) => renderJavbusRow(row, index, rows)).join("") : `<tr><td colspan="10"><div class="sc-empty-state">当前筛选下没有 JavBus 抓取任务。</div></td></tr>`}</tbody>
    </table>
  </div>`;
}

function renderJavbusPage(data: AnyRecord): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  return `<div id="javbusListPage" class="sehuatang-list-content javbus-list-content" data-task-page-url="${attr(pick(data, "TaskPageURL") || "/sitefetch")}" data-filtered-task-type="${attr(pick(data, "FilteredJavbusTaskType") || "")}" data-filter-form-id="fetchSiteJavbusFilterForm">
    <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell">
      <div class="next02-sc-hero-copy">
        <span class="next02-sc-eyebrow">JAVBUS / FETCH POOL</span>
        <h4>${escapeHtml(pick(data, "PageTitle") || "JavBus 抓取列表")}</h4>
        <p>${escapeHtml(pick(data, "PageNote") || "")}</p>
      </div>
      <div class="next02-sc-hero-actions">
        <div class="btn-group" role="group" aria-label="JavBus 页面操作">
          <a class="btn btn-outline-secondary" href="${attr(pick(data, "PeerPageURL") || "#")}">Sukebei 列表</a>
          <button class="btn btn-outline-secondary js-trigger-filtered-fetch-site" type="button">触发筛选抓取</button>
          <a class="btn btn-outline-secondary" href="${attr(pick(data, "TaskPageURL") || "/sitefetch")}">抓取任务页</a>
          <a class="btn btn-outline-secondary" href="${attr(pick(data, "TasksPageURL") || "/crawltasklist")}">任务列表</a>
        </div>
      </div>
    </section>
    <div id="fetchSiteListMsg" class="alert d-none mb-3" role="alert"></div>
    <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
      <div class="card-body">
        <div class="next02-sc-panel-head">
          ${renderJavbusFilterForm(data)}
          ${rows.length ? `<div class="next02-sc-hero-metrics sc-event-list-metrics" aria-label="JavBus 抓取摘要"><span><strong>${escapeHtml(valueOf(pick(data, "Total"), 0))}</strong> Records</span><span><strong>${escapeHtml(rows.length)}</strong> Current Page</span></div>` : ""}
          ${rows.length ? `<div class="next02-sc-movie-pager">${pageLinks(pick(data, "PageInfo"), "JavBus")}</div>` : ""}
        </div>
        ${renderJavbusTable(data)}
      </div>
    </div>
  </div>`;
}

function renderSukebeiFilterForm(data: AnyRecord): string {
  const q = asRecord(pick(data, "Query"));
  return `<form class="sc-event-filter-form sehuatang-filter-form fetch-pool-filter-form" method="get" action="/sukebeilist" id="fetchSiteSukebeiFilterForm" data-fetch-site-filter-form="1">
    <input type="hidden" name="ps" value="${attr(pick(q, "PageSize") || "")}">
    ${pick(q, "Sort") ? `<input type="hidden" name="sort" value="${attr(pick(q, "Sort"))}">` : ""}
    ${pick(q, "Order") ? `<input type="hidden" name="order" value="${attr(pick(q, "Order"))}">` : ""}
    ${pick(q, "MediaOwned") ? `<input type="hidden" name="mowned" value="${attr(pick(q, "MediaOwned"))}">` : ""}
    <div class="sehuatang-filter-cluster sehuatang-filter-query" aria-label="检索条件">
      <label class="sc-event-filter-field sehuatang-filter-field is-keyword" for="sukebeiKeywordInput">
        <span>影片编号 / JavID</span>
        <input class="form-control" id="sukebeiKeywordInput" type="text" name="keyword" value="${attr(pick(q, "Keyword") || "")}" placeholder="movie_name 或 movie_jav_id">
      </label>
      <div class="sc-event-filter-field sehuatang-filter-field is-fetch-pool-status">
        <span>Sukebei 状态</span>
        <div class="album-quick-filter-group fetch-site-status-group">${renderFetchSiteStatusGroup(data, "sukebei")}</div>
      </div>
      <label class="sc-event-filter-field sehuatang-filter-field is-fetch-pool-sort" for="sukebeiTriggerSortSelect">
        <span>触发抓取顺序</span>
        <select class="form-select" id="sukebeiTriggerSortSelect" name="trigger_sort_key">${triggerSortOptions(q)}</select>
      </label>
    </div>
    <div class="sehuatang-filter-cluster sehuatang-filter-date-range" aria-label="最后抓取日期范围">
      <span class="sehuatang-filter-cluster-label">最后抓取</span>
      <label class="sc-event-filter-field sehuatang-filter-field is-date" for="sukebeiLastFetchFromInput"><span>开始</span><input class="form-control" id="sukebeiLastFetchFromInput" type="text" name="last_fetch_from" value="${attr(pick(q, "LastFetchFrom") || "")}"></label>
      <span class="sehuatang-filter-range-separator" aria-hidden="true">—</span>
      <label class="sc-event-filter-field sehuatang-filter-field is-date" for="sukebeiLastFetchToInput"><span>结束</span><input class="form-control" id="sukebeiLastFetchToInput" type="text" name="last_fetch_to" value="${attr(pick(q, "LastFetchTo") || "")}"></label>
    </div>
    <div class="sehuatang-filter-cluster sehuatang-filter-quick-actions" aria-label="快捷筛选与操作">
      <div class="sc-event-filter-field sehuatang-filter-field is-quick-tag">
        <span>WMedia</span>
        ${renderSukebeiOwnedFilter(data)}
      </div>
      <div class="sehuatang-filter-actions">
        <button class="sc-event-filter-submit" type="submit">筛选</button>
        <a class="sc-event-filter-clear" href="/sukebeilist">清空</a>
      </div>
    </div>
    <div class="fetch-pool-filter-secondary">
      <div class="sehuatang-filter-cluster sehuatang-filter-date-range" aria-label="发行日期范围">
        <span class="sehuatang-filter-cluster-label">发行时间</span>
        <label class="sc-event-filter-field sehuatang-filter-field is-date" for="sukebeiReleaseFromInput"><span>开始</span><input class="form-control" id="sukebeiReleaseFromInput" type="text" name="release_date_from" value="${attr(pick(q, "ReleaseDateFrom") || "")}"></label>
        <span class="sehuatang-filter-range-separator" aria-hidden="true">—</span>
        <label class="sc-event-filter-field sehuatang-filter-field is-date" for="sukebeiReleaseToInput"><span>结束</span><input class="form-control" id="sukebeiReleaseToInput" type="text" name="release_date_to" value="${attr(pick(q, "ReleaseDateTo") || "")}"></label>
      </div>
      <div class="sehuatang-filter-cluster sehuatang-filter-date-range" aria-label="WMedia 下载日期范围">
        <span class="sehuatang-filter-cluster-label">WMedia 下载</span>
        <label class="sc-event-filter-field sehuatang-filter-field is-date" for="sukebeiMediaBirthFromInput"><span>开始</span><input class="form-control" id="sukebeiMediaBirthFromInput" type="text" name="media_birth_from" value="${attr(pick(q, "MediaBirthFrom") || "")}"></label>
        <span class="sehuatang-filter-range-separator" aria-hidden="true">—</span>
        <label class="sc-event-filter-field sehuatang-filter-field is-date" for="sukebeiMediaBirthToInput"><span>结束</span><input class="form-control" id="sukebeiMediaBirthToInput" type="text" name="media_birth_to" value="${attr(pick(q, "MediaBirthTo") || "")}"></label>
      </div>
    </div>
  </form>`;
}

function renderSukebeiOwnedFilter(data: AnyRecord): string {
  const owned = asRecord(pick(data, "ownedQuery"));
  const link = (item: unknown, label: string) => {
    const current = asRecord(item);
    return `<a href="${attr(normalizeHref(readText(current, "Href") || "#"))}" class="btn btn-sm ${pick(current, "Active") ? "btn-primary active disabled" : "btn-outline-primary"}" aria-current="${pick(current, "Active") ? "true" : "false"}">${escapeHtml(label)}</a>`;
  };
  return `<div class="album-quick-filter-group" aria-label="WMedia 筛选">
    ${link(pick(owned, "MediaAll"), "MAll")}
    ${link(pick(owned, "MediaOwned"), "MOwned")}
    ${link(pick(owned, "MediaNotOwned"), "MNotOwned")}
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

function renderSukebeiTime(value: unknown): string {
  const current = text(value);
  if (current === "-") {
    return "-";
  }
  const [date, time = ""] = current.split(" ", 2);
  return `<span class="sc-event-time"><span class="sc-event-date">${escapeHtml(date)}</span>${time ? `<span class="sc-event-hour">${escapeHtml(time)}</span>` : ""}</span>`;
}

function sukebeiDayKey(row: AnyRecord): string {
  const current = readText(row, "LastFetchText");
  return current === "-" ? "" : current.slice(0, 10);
}

function sukebeiDayClass(rows: AnyRecord[], index: number): string {
  const day = sukebeiDayKey(rows[index]);
  if (!day) {
    return "sehuatang-day-neutral";
  }
  let groupIndex = 0;
  for (let rowIndex = 1; rowIndex <= index; rowIndex += 1) {
    const currentDay = sukebeiDayKey(rows[rowIndex]);
    const previousDay = sukebeiDayKey(rows[rowIndex - 1]);
    if (currentDay && previousDay && currentDay !== previousDay) {
      groupIndex += 1;
    }
  }
  return `sehuatang-day-group-${(groupIndex % 2) + 1}`;
}

function isSukebeiDayStart(rows: AnyRecord[], index: number): boolean {
  if (index <= 0) {
    return false;
  }
  const currentDay = sukebeiDayKey(rows[index]);
  const previousDay = sukebeiDayKey(rows[index - 1]);
  return Boolean(currentDay && previousDay && currentDay !== previousDay);
}

function renderSukebeiRow(row: AnyRecord, index: number, rows: AnyRecord[]): string {
  const movieName = readText(row, "MovieName");
  const lastFetchText = readText(row, "LastFetchText");
  const dayStartClass = isSukebeiDayStart(rows, index) ? " sehuatang-day-start" : "";
  const errorHTML = readText(row, "LastError")
    ? `<details class="fetch-site-error-details"><summary>有错误</summary><div>${escapeHtml(readText(row, "LastError"))}</div></details>`
    : `<span class="sukebei-no-error">无错误</span>`;
  return `<tr class="next02-sc-event-row ${sukebeiDayClass(rows, index)}${dayStartClass}" data-movie-jav-id="${attr(readText(row, "MovieJavID"))}" data-movie-name="${attr(movieName)}" data-fetch-day="${attr(sukebeiDayKey(row))}">
    <td class="sc-content-cell"><a class="sc-event-link" href="/moviedetail/${attr(encodeURIComponent(movieName))}" title="${attr(movieName)}">${escapeHtml(movieName)}</a></td>
    <td title="${attr(readText(row, "ReleaseDateText"))}">${renderSukebeiTime(readText(row, "ReleaseDateText"))}</td>
    <td><div class="fetch-site-album-inline"><span class="btn btn-sm btn-outline-secondary disabled">${escapeHtml(readText(row, "FetchStatusText"))}</span></div></td>
    <td class="sc-number">${escapeHtml(valueOf(pick(row, "TryCount"), 0))}</td>
    <td title="${attr(lastFetchText)}">${renderSukebeiTime(lastFetchText)}</td>
    <td class="sc-number sc-primary-value">${escapeHtml(valueOf(pick(row, "TorrentHashCount"), 0))}</td>
    <td title="${attr(readText(row, "LatestPublishText"))}">${renderSukebeiTime(readText(row, "LatestPublishText"))}</td>
    <td class="fetch-pool-wmedia-cell"><div class="fetch-site-album-inline">${renderWMediaCell(row)}</div></td>
    <td class="fetch-site-error-cell">${errorHTML}</td>
    <td><div class="fetch-site-album-inline"><button type="button" class="btn btn-sm btn-outline-secondary js-trigger-sukebei">触发 Sukebei</button></div></td>
  </tr>`;
}

function renderSukebeiTable(data: AnyRecord): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  const sort = asRecord(pick(data, "SortQuery"));
  return `<div class="table-responsive">
        <table class="table table-hover align-middle mb-0 sc-events-table sehuatang-magnet-table fetch-pool-task-table">
          <colgroup>
            <col class="fetch-pool-col-movie"><col class="fetch-pool-col-release"><col class="fetch-pool-col-status"><col class="fetch-pool-col-retry"><col class="fetch-pool-col-last-fetch">
            <col class="fetch-pool-col-hash"><col class="fetch-pool-col-latest"><col class="fetch-pool-col-wmedia"><col class="fetch-pool-col-error"><col class="fetch-pool-col-action">
          </colgroup>
          <thead>
            <tr>
              <th class="sc-sortable-head">${sortLabel(pick(sort, "ByMovieName"), "影片")}</th>
              <th class="sc-sortable-head">${sortLabel(pick(sort, "ByReleaseDate"), "发行时间")}</th>
              <th class="sc-sortable-head">${sortLabel(pick(sort, "ByFetchStatus"), "状态")}</th>
              <th>重试</th>
              <th class="sc-sortable-head">${sortLabel(pick(sort, "ByLastFetchTime"), "最后抓取")}</th>
              <th class="sc-sortable-head">${sortLabel(pick(sort, "ByHashCount"), "Hash 数")}</th>
              <th class="sc-sortable-head">${sortLabel(pick(sort, "ByLatestPublish"), "最新日期")}</th>
              <th class="sc-sortable-head">${sortLabel(pick(sort, "ByMediaBirthTime"), "WMedia 下载时间")}</th>
              <th>错误</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            ${rows.length
              ? rows.map((row, index) => renderSukebeiRow(row, index, rows)).join("")
              : `<tr><td colspan="10"><div class="sc-empty-state">当前筛选下没有 Sukebei 抓取任务。</div></td></tr>`}
          </tbody>
        </table>
      </div>`;
}

function renderSukebeiPage(data: AnyRecord): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  return `<div id="sukebeiListPage" class="sehuatang-list-content sukebei-list-content" data-task-page-url="${attr(pick(data, "TaskPageURL") || "/sitefetch")}" data-filtered-sukebei-task-type="${attr(pick(data, "FilteredSukebeiTaskType") || "")}" data-filter-form-id="fetchSiteSukebeiFilterForm">
    <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell">
      <div class="next02-sc-hero-copy">
        <span class="next02-sc-eyebrow">SUKEBEI / TORRENT POOL</span>
        <h4>${escapeHtml(pick(data, "PageTitle") || "Sukebei 抓取列表")}</h4>
        <p>${escapeHtml(pick(data, "PageNote") || "")}</p>
      </div>
      <div class="next02-sc-hero-actions">
        <div class="btn-group" role="group" aria-label="Sukebei 页面操作">
          <a class="btn btn-outline-secondary" href="${attr(pick(data, "PeerPageURL") || "#")}">JavBus 列表</a>
          <button class="btn btn-outline-secondary js-trigger-filtered-sukebei" type="button">触发筛选抓取</button>
          <a class="btn btn-outline-secondary" href="${attr(pick(data, "TaskPageURL") || "#")}">抓取任务页</a>
          <a class="btn btn-outline-secondary" href="${attr(pick(data, "TasksPageURL") || "/crawltasklist")}">任务列表</a>
        </div>
      </div>
    </section>
    <div id="fetchSiteListMsg" class="alert d-none mb-3" role="alert"></div>
    <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
      <div class="card-body">
        <div class="next02-sc-panel-head">
          ${renderSukebeiFilterForm(data)}
          ${rows.length ? `<div class="next02-sc-hero-metrics sc-event-list-metrics" aria-label="Sukebei 抓取摘要"><span><strong>${escapeHtml(valueOf(pick(data, "Total"), 0))}</strong> Records</span><span><strong>${escapeHtml(rows.length)}</strong> Current Page</span></div>` : ""}
          ${rows.length ? `<div class="next02-sc-movie-pager">${pageLinks(pick(data, "PageInfo"), "Sukebei")}</div>` : ""}
        </div>
        ${renderSukebeiTable(data)}
      </div>
    </div>
  </div>`;
}

function renderSehuatangTime(value: unknown): string {
  const formatted = formatUnix(value);
  if (formatted === "-") {
    return `<span class="sc-event-time"><span class="sc-event-date">-</span></span>`;
  }
  const [date, time = ""] = formatted.split(" ", 2);
  return `<span class="sc-event-time"><span class="sc-event-date">${escapeHtml(date)}</span>${time ? `<small class="sc-event-hour">${escapeHtml(time)}</small>` : ""}</span>`;
}

function sehuatangDayKey(row: AnyRecord): string {
  const date = formatUnix(pick(row, "PostTime"), "date");
  return date === "-" ? "" : date;
}

function sehuatangDayClass(rows: AnyRecord[], index: number): string {
  const day = sehuatangDayKey(rows[index]);
  if (!day) {
    return "sehuatang-day-neutral";
  }

  let groupIndex = 0;
  for (let rowIndex = 1; rowIndex <= index; rowIndex += 1) {
    const currentDay = sehuatangDayKey(rows[rowIndex]);
    const previousDay = sehuatangDayKey(rows[rowIndex - 1]);
    if (currentDay && previousDay && currentDay !== previousDay) {
      groupIndex += 1;
    }
  }
  return `sehuatang-day-group-${(groupIndex % 2) + 1}`;
}

function isSehuatangDayStart(rows: AnyRecord[], index: number): boolean {
  if (index <= 0) {
    return false;
  }
  const currentDay = sehuatangDayKey(rows[index]);
  const previousDay = sehuatangDayKey(rows[index - 1]);
  return Boolean(currentDay && previousDay && currentDay !== previousDay);
}

function lineCount(value: unknown): number {
  const current = String(value || "").trim();
  return current ? current.split(/\r?\n/).filter((line) => line.trim() !== "").length : 0;
}

function shMultiQueryValues(q: AnyRecord, pluralKey: string, legacyKey: string): string[] {
  const values = asArray<unknown>(pick(q, pluralKey)).map((value) => String(value || "").trim()).filter(Boolean);
  const legacy = String(pick(q, legacyKey) || "").trim();
  if (legacy) {
    values.push(legacy);
  }
  return Array.from(new Set(values));
}

function shQuickFilterHref(name: string, value: string): string {
  const legacyNames: Record<string, string> = {
    movie_types: "movie_type",
    source_forum_ids: "source_forum_id",
    detail_statuses: "detail_status"
  };
  const params = new URLSearchParams();
  Object.entries(route.query).forEach(([key, rawValue]) => {
    const values = Array.isArray(rawValue) ? rawValue : [rawValue];
    values.forEach((item) => {
      if (item !== null && item !== undefined) {
        params.append(key, String(item));
      }
    });
  });
  params.delete(name);
  params.delete(legacyNames[name] || "");
  params.set("p", "1");
  if (value) {
    params.set(name, value);
  }
  return `/shpool?${params.toString()}`;
}

function renderSHMultiFilter(label: string, name: string, selectedValues: string[], options: [string, string][]): string {
  const selected = new Set(selectedValues);
  return `<div class="next02-sh-filter-cluster next02-sh-filter-choice" aria-label="${attr(label)}">
    <span class="next02-sh-filter-cluster-label">${escapeHtml(label)}</span>
    <div class="next02-sh-filter-choice-buttons">
      <a class="next02-sh-filter-choice-all ${selected.size === 0 ? "active" : ""}" href="${attr(shQuickFilterHref(name, ""))}">全部</a>
      ${options.map(([value, optionLabel]) => `<a class="next02-sh-filter-choice-button ${selected.has(value) ? "active" : ""}" href="${attr(shQuickFilterHref(name, value))}">${escapeHtml(optionLabel)}</a>`).join("")}
    </div>
  </div>`;
}

function renderSHFilterForm(data: AnyRecord): string {
  const q = asRecord(pick(data, "Query"));
  const sourceForumIDs = shMultiQueryValues(q, "SourceForumIDs", "SourceForumID");
  const detailStatuses = shMultiQueryValues(q, "DetailStatuses", "DetailStatus");
  const movieTypes = shMultiQueryValues(q, "MovieTypes", "MovieType");
  const pageSize = Number(pick(q, "PageSize") || 50);
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  const total = valueOf(pick(data, "Total"), 0);
  const aggregate = asRecord(pick(data, "Aggregate"));
  return `<form class="next02-sh-filter-form bucket-list-controls sh-bucket-controls sh-bucket-toolbar next02-sh-filter-toolbar" id="sehuatangSHFilterForm" method="get" action="/shpool">
        ${pick(q, "Sort") ? `<input type="hidden" name="sort" value="${attr(pick(q, "Sort"))}">` : ""}
        ${pick(q, "Order") ? `<input type="hidden" name="order" value="${attr(pick(q, "Order"))}">` : ""}
        <div class="bucket-list-control-main sh-bucket-toolbar-main next02-sh-filter-toolbar-main">
          <div class="next02-sh-filter-toolbar-top">
            <div class="sh-bucket-toolbar-filter-area next02-sh-filter-query" aria-label="检索条件">
              <label class="next02-sh-filter-field is-keyword" for="sehuatangSHKeywordInput"><span>关键词</span><input class="form-control" id="sehuatangSHKeywordInput" type="text" name="keyword" value="${attr(pick(q, "Keyword") || "")}" placeholder="帖子标题 / 正文"></label>
              <label class="next02-sh-filter-field is-hash" for="sehuatangSHHashInput"><span>info_hash</span><input class="form-control" id="sehuatangSHHashInput" type="text" name="info_hash" value="${attr(pick(q, "InfoHash") || "")}"></label>
              <label class="next02-sh-filter-field next02-sh-page-size-field" for="sehuatangSHPageSize"><span>每页</span><input class="form-control" id="sehuatangSHPageSize" type="number" step="1" name="ps" value="${attr(pageSize)}" inputmode="numeric"></label>
              <div class="next02-sh-filter-cluster next02-sh-filter-dates" aria-label="发帖日期范围">
                <span class="next02-sh-filter-cluster-label">发帖日期</span>
                <label class="next02-sh-filter-field smart-pick-date-field" for="sehuatangSHPostFromInput"><span>开始</span><input class="form-control smart-pick-date-input" id="sehuatangSHPostFromInput" type="text" name="post_time_from" value="${attr(pick(q, "PostTimeFrom") || "")}" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"></label>
                <span class="next02-sh-filter-range-separator" aria-hidden="true">—</span>
                <label class="next02-sh-filter-field smart-pick-date-field" for="sehuatangSHPostToInput"><span>结束</span><input class="form-control smart-pick-date-input" id="sehuatangSHPostToInput" type="text" name="post_time_to" value="${attr(pick(q, "PostTimeTo") || "")}" placeholder="YYYY-MM-DD" inputmode="numeric" maxlength="10" pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"></label>
              </div>
              <div class="next02-sh-filter-actions bucket-list-filter-actions">
                <button class="sc-event-filter-submit bucket-list-apply" type="submit">筛选</button>
                <a class="sc-event-filter-clear bucket-list-clear" href="/shpool">清空</a>
              </div>
            </div>
          </div>
          <div class="next02-sh-filter-toolbar-bottom next02-shpool-filter-toolbar-bottom">
            <div class="sh-bucket-toolbar-filter-area next02-sh-filter-secondary">
              ${renderSHMultiFilter("影片类型", "movie_types", movieTypes, [["jav", "JAV"], ["fc2ppv", "FC2PPV"]])}
              ${renderSHMultiFilter("来源版块", "source_forum_ids", sourceForumIDs, shForumFilterOptions(shForums.value))}
              ${renderSHMultiFilter("详情状态", "detail_statuses", detailStatuses, [["pending", "待抓详情"], ["failed", "详情失败"], ["fetched", "已抓详情"]])}
            </div>
            <div class="bucket-list-control-meta sh-bucket-toolbar-status next02-sh-filter-toolbar-status next02-shpool-pagination">
              ${rows.length ? pageLinks(pick(data, "PageInfo"), "SehuatangSH") : ""}
            </div>
          </div>
          <div class="next02-shpool-metrics-row">
            <div class="sh-bucket-toolbar-metrics next02-sh-toolbar-metrics" aria-label="Sehuatang SH 聚合摘要">
              <div class="next02-sh-metric-group">
                <span class="next02-sh-metric-group-label">数据范围</span>
                ${renderSHMetric(total, "总数")}
                ${renderSHMetric(rows.length, "当前页")}
              </div>
              <div class="next02-sh-metric-group">
                <span class="next02-sh-metric-group-label">详情状态</span>
                ${renderSHMetric(valueOf(pick(aggregate, "PendingDetailCount"), 0), "待抓详情", "is-pending")}
                ${renderSHMetric(valueOf(pick(aggregate, "FetchedDetailCount"), 0), "已抓详情", "is-fetched")}
                ${renderSHMetric(valueOf(pick(aggregate, "FailedDetailCount"), 0), "详情失败", "is-failed")}
              </div>
              <div class="next02-sh-metric-group">
                <span class="next02-sh-metric-group-label">来源板块</span>
                ${renderSHForumMetrics(aggregate)}
              </div>
            </div>
          </div>
        </div>
      </form>`;
}

function renderSHThreadRow(row: AnyRecord, index: number, rows: AnyRecord[]): string {
  const title = readText(row, "ThreadTitle");
  const threadTID = readText(row, "ThreadTID");
  const movieName = readText(row, "MovieName");
  const movieType = readText(row, "MovieType");
  const fc2Key = readText(row, "FC2Key");
  const threadURL = readText(row, "ThreadURL");
  const titleHTML = threadTID
    ? `<a class="fetch-site-title-ellipsis" href="/shthread/${attr(encodeURIComponent(threadTID))}" target="_blank" rel="noopener noreferrer" title="${attr(title)}">${escapeHtml(title)}</a>`
    : `<span class="fetch-site-title-ellipsis" title="${attr(title)}">${escapeHtml(title)}</span>`;
  const originalPostHTML = threadURL
    ? `<a class="btn btn-outline-primary" href="${attr(threadURL)}" target="_blank" rel="noopener noreferrer">原帖</a>`
    : `<button type="button" class="btn btn-outline-primary" disabled title="缺少原帖地址">原帖</button>`;
  const movieNameHTML = movieType === "jav" && movieName
    ? `<a class="next02-sh-movie-name" href="/moviedetail/${attr(encodeURIComponent(movieName))}" target="_blank" rel="noopener noreferrer">${escapeHtml(movieName)}</a>`
    : movieType === "fc2ppv" && fc2Key
      ? `<a class="next02-sh-movie-name" href="/fcdetail/${attr(encodeURIComponent(fc2Key))}" target="_blank" rel="noopener noreferrer">${escapeHtml(fc2Key)}</a>`
      : `<span class="next02-sh-movie-name">${escapeHtml(movieName || "-")}</span>`;
  const magnetCount = lineCount(pick(row, "MagnetText"));
  const status = readText(row, "DetailStatus");
  const forumID = readText(row, "SourceForumID");
  const dayStartClass = isSehuatangDayStart(rows, index) ? " sehuatang-day-start" : "";
  return `<tr class="next02-sc-event-row ${sehuatangDayClass(rows, index)}${dayStartClass}" data-post-day="${attr(formatUnix(pick(row, "PostTime"), "date"))}">
    <td class="next02-sh-select-cell"><input class="form-check-input js-sh-detail-select" type="checkbox" value="${attr(threadTID)}" aria-label="选择帖子 ${attr(threadTID)}" ${threadTID ? "" : "disabled"}></td>
    <td class="next02-sh-number-cell" title="${attr(threadTID)}">${escapeHtml(threadTID)}</td>
    <td class="next02-sh-movie-name-cell" title="${attr(fc2Key || movieName)}">${movieNameHTML}</td>
    <td class="sc-content-cell">${titleHTML}</td>
    <td>${renderSehuatangTime(pick(row, "PostTime"))}</td>
    <td><span class="next02-sh-forum-pill" data-forum-id="${attr(forumID)}">${escapeHtml(formatSHForumName(forumID))}</span></td>
    <td><span class="next02-sh-status-pill" data-status="${attr(status)}">${escapeHtml(status)}</span></td>
    <td title="${attr(formatUnix(pick(row, "LastSeenTime")))}">${renderSehuatangTime(pick(row, "LastSeenTime"))}</td>
    <td class="next02-sh-actions-cell"><div class="btn-group btn-group-sm" role="group" aria-label="帖子操作">
      <button type="button" class="btn btn-outline-primary js-sh-detail-start" data-thread-tid="${attr(threadTID)}" title="抓取详情" ${threadTID ? "" : "disabled"}>抓取</button>
      ${originalPostHTML}
      <button type="button" class="btn btn-outline-secondary js-sh-copy" data-index="${attr(index)}" data-field="MagnetText" title="复制 magnet" ${magnetCount ? "" : "disabled"}>M</button>
      <button type="button" class="btn btn-outline-danger js-sh-delete" data-thread-tid="${attr(threadTID)}" title="删除帖子" ${threadTID ? "" : "disabled"}>删除</button>
    </div></td>
  </tr>`;
}

function renderSHThreadTable(data: AnyRecord): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  const sort = asRecord(pick(data, "SortQuery"));
  return `<div class="table-responsive">
        <table class="table table-hover align-middle mb-0 sc-events-table next02-sh-thread-table">
          <colgroup>
            <col class="next02-sh-col-select"><col class="next02-sh-col-tid"><col class="next02-sh-col-movie"><col class="next02-sh-col-title"><col class="next02-sh-col-post">
            <col class="next02-sh-col-forum"><col class="next02-sh-col-status">
            <col class="next02-sh-col-seen"><col class="next02-sh-col-actions">
          </colgroup>
          <thead><tr>
            <th>选择</th>
            <th class="sc-sortable-head">${sortLabel(pick(sort, "ByThreadTID"), "thread_tid")}</th>
            <th class="sc-sortable-head">${sortLabel(pick(sort, "ByMovieName"), "影片名")}</th>
            <th>帖子标题</th>
            <th class="sc-sortable-head">${sortLabel(pick(sort, "ByPostTime"), "发布时间")}</th>
            <th>版块</th>
            <th class="sc-sortable-head">${sortLabel(pick(sort, "ByDetailStatus"), "详情状态")}</th>
            <th class="sc-sortable-head">${sortLabel(pick(sort, "ByLastSeenTime"), "最近抓到")}</th>
            <th>操作</th>
          </tr></thead>
          <tbody>${rows.length ? rows.map((row, index) => renderSHThreadRow(row, index, rows)).join("") : `<tr><td colspan="9"><div class="fetch-site-empty-state">当前筛选下没有 Sehuatang SH 帖子数据。</div></td></tr>`}</tbody>
        </table>
      </div>`;
}

function renderSHDetailBatchActions(rows: AnyRecord[]): string {
  if (!rows.length) {
    return "";
  }
  return `<div class="next02-sh-detail-batch-bar" aria-label="详情抓取批量操作">
    <div class="next02-sh-detail-batch-actions">
      <button type="button" class="btn btn-outline-secondary btn-sm js-sh-detail-select-all">全选当前页</button>
      <button type="button" class="btn btn-outline-secondary btn-sm js-sh-detail-clear-selection" disabled>取消选择</button>
      <button type="button" class="btn btn-primary btn-sm js-sh-detail-start-selected" disabled>抓取选中详情</button>
      <button type="button" class="btn btn-outline-danger btn-sm js-sh-delete-selected" disabled>删除选中</button>
    </div>
    <div class="next02-sh-detail-selection-count">已选择 <strong id="shDetailSelectedCount">0</strong> 条</div>
  </div>`;
}

function renderSHRuns(data: AnyRecord): string {
  const runs = asArray<AnyRecord>(pick(data, "RecentRuns"));
  return `<div class="card next02-sc-panel next02-sh-runs-panel">
    <div class="card-body">
      <div class="next02-sh-runs-head">
        <div>
          <div class="next02-sh-section-kicker">RUNS / RECENT</div>
          <h5 class="card-title mb-0">最近抓取记录</h5>
        </div>
        <div class="next02-sh-table-meta">${escapeHtml(runs.length)} 条</div>
      </div>
      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0 next02-sh-runs-table">
          <colgroup>
            <col class="next02-sh-run-id"><col class="next02-sh-run-mode"><col class="next02-sh-run-keyword"><col class="next02-sh-run-forum">
            <col class="next02-sh-run-status"><col class="next02-sh-run-count"><col class="next02-sh-run-count"><col class="next02-sh-run-count">
            <col class="next02-sh-run-count"><col class="next02-sh-run-count"><col class="next02-sh-run-count"><col class="next02-sh-run-time">
            <col class="next02-sh-run-time"><col class="next02-sh-run-error">
          </colgroup>
          <thead><tr><th>run_id</th><th>模式</th><th>关键词</th><th>版块</th><th>状态</th><th>发现</th><th>命中</th><th>跳过版块</th><th>跳过已存在</th><th>入库</th><th>失败</th><th>开始</th><th>结束</th><th>错误</th></tr></thead>
          <tbody>${runs.length ? runs.map((run) => `<tr>
            <td>${escapeHtml(pick(run, "Id") || "")}</td>
            <td>${escapeHtml(readText(run, "FetchMode"))}</td>
            <td>${escapeHtml(readText(run, "Keyword"))}</td>
            <td>${escapeHtml(shForumNames(shForums.value, readText(run, "TargetForumIDs")))}</td>
            <td><span class="badge fetch-site-badge">${escapeHtml(readText(run, "Status"))}</span></td>
            <td>${escapeHtml(valueOf(pick(run, "SeenCount"), 0))}</td>
            <td>${escapeHtml(valueOf(pick(run, "MatchedCount"), 0))}</td>
            <td>${escapeHtml(valueOf(pick(run, "SkippedForumCount"), 0))}</td>
            <td>${escapeHtml(valueOf(pick(run, "SkippedExistingCount"), 0))}</td>
            <td>${escapeHtml(valueOf(pick(run, "FetchedCount"), 0))}</td>
            <td>${escapeHtml(valueOf(pick(run, "FailedCount"), 0))}</td>
            <td title="${attr(formatUnix(pick(run, "StartedOn")))}">${renderSehuatangTime(pick(run, "StartedOn"))}</td>
            <td title="${attr(formatUnix(pick(run, "FinishedOn")))}">${renderSehuatangTime(pick(run, "FinishedOn"))}</td>
            <td>${readText(run, "LastError") ? `<details><summary>错误</summary><div class="mt-2">${escapeHtml(readText(run, "LastError"))}</div></details>` : "-"}</td>
          </tr>`).join("") : `<tr><td colspan="14"><div class="fetch-site-empty-state">暂无抓取记录。</div></td></tr>`}</tbody>
        </table>
      </div>
    </div>
  </div>`;
}

function formatSHForumName(value: unknown): string {
  return shForumName(shForums.value, value);
}

function renderSHMetric(value: unknown, label: string, className = ""): string {
  return `<span class="next02-sh-metric ${attr(className)}" title="${attr(`${label} ${valueOf(value, 0)}`)}" aria-label="${attr(`${label} ${valueOf(value, 0)}`)}"><span class="next02-sh-metric-label">${escapeHtml(label)}</span><strong>${escapeHtml(valueOf(value, 0))}</strong></span>`;
}

function renderSHForumMetrics(aggregate: AnyRecord): string {
  return shForums.value.map((forum) => renderSHMetric(valueOf(pick(aggregate, `Forum${forum.forumID}Count`), 0), forum.forumName, "is-forum")).join("");
}

function renderSHDeleteModal(): string {
  return `<div class="modal fade next02-sh-delete-modal" id="shThreadDeleteModal" tabindex="-1" aria-hidden="true" aria-labelledby="shThreadDeleteTitle" aria-describedby="shThreadDeleteDescription">
    <div class="modal-dialog modal-dialog-centered"><div class="modal-content">
      <div class="modal-header"><h5 class="modal-title" id="shThreadDeleteTitle">确认删除帖子</h5><button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="关闭"></button></div>
      <div class="modal-body">
        <p id="shThreadDeleteDescription">将从帖子池删除选中的帖子。</p>
        <div class="next02-sh-delete-warning">抓取历史会保留，相关时间桶聚合数据将同步刷新。</div>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-outline-secondary" data-bs-dismiss="modal">取消</button>
        <button type="button" class="btn btn-danger js-sh-delete-confirm">确认删除</button>
      </div>
    </div></div>
  </div>`;
}

function renderSHPage(data: AnyRecord): string {
  const rows = asArray<AnyRecord>(pick(data, "Rows"));
  return `<div id="sehuatangSHListPage" class="next02-sh-list-rewrite">
    <section class="next02-sc-hero next02-sc-events-hero next02-sh-list-hero next02-sc-movie-top-shell">
      <div class="next02-sc-hero-copy">
        <span class="next02-sc-eyebrow">SEHUATANG / SH THREAD POOL</span>
        <h4>${escapeHtml(pick(data, "PageTitle") || "Sehuatang SH 帖子列表")}</h4>
        <p>${escapeHtml(pick(data, "PageNote") || "")}</p>
      </div>
      <div class="next02-sc-hero-actions next02-sh-list-actions">
        <a class="btn btn-outline-primary btn-sm" href="/shpool">SehuatangSH 列表</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "TaskPageURL") || "/shfetch")}" target="_blank" rel="noopener noreferrer">列表抓取</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "DetailTaskPageURL") || "/shfill")}" target="_blank" rel="noopener noreferrer">详情补全</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "TasksPageURL") || "/crawltasklist")}" target="_blank" rel="noopener noreferrer">任务列表</a>
      </div>
    </section>
    <div id="sehuatangSHListMsg" class="alert d-none mb-3" role="alert"></div>
    <div class="card next02-sc-panel next02-sc-events-panel next02-sh-list-panel next02-sc-movie-table-card">
      <div class="card-body">
        <div class="next02-sc-panel-head next02-sh-list-control">
          ${renderSHFilterForm(data)}
        </div>
        ${renderSHDetailBatchActions(rows)}
        ${renderSHThreadTable(data)}
      </div>
    </div>
    ${renderSHRuns(data)}
  </div>
  ${renderSHDeleteModal()}`;
}

function formatSHDetailStatus(status: unknown): string {
  const current = String(status || "").trim();
  if (current === "pending") return "待抓详情";
  if (current === "fetched") return "已抓详情";
  if (current === "failed") return "详情失败";
  return text(current);
}

function renderSHDetailInfo(thread: AnyRecord): string {
  const status = readText(thread, "DetailStatus");
  const item = (label: string, value: string, className = "") => `<div class="next02-sh-detail-info-item${className ? ` ${className}` : ""}">
    <span class="next02-sh-detail-info-label">${escapeHtml(label)}</span>
    <div class="next02-sh-detail-info-value">${value || "-"}</div>
  </div>`;
  return `<section class="card next02-sc-panel next02-sc-events-panel next02-sh-detail-panel next02-sh-detail-info-panel">
    <div class="card-body">
      <div class="next02-sh-detail-section-head">
        <div><span class="next02-sh-section-kicker">THREAD METADATA</span><h5>帖子信息</h5></div>
        <span class="next02-sh-detail-table-meta">${escapeHtml(readText(thread, "ThreadTID"))}</span>
      </div>
      <div class="next02-sh-detail-info-grid">
        ${item("thread_tid", `<span class="next02-sh-detail-code-pill">${escapeHtml(readText(thread, "ThreadTID"))}</span>`)}
        ${item("来源版块", `<span class="next02-sh-detail-forum-pill">${escapeHtml(formatSHForumName(readText(thread, "SourceForumID")))}</span>`)}
        ${item("来源页", readText(thread, "SourcePageURL") ? `<a class="next02-sh-detail-link-pill" href="${attr(readText(thread, "SourcePageURL"))}" target="_blank" rel="noopener noreferrer">${escapeHtml(readText(thread, "SourcePage"))}</a>` : escapeHtml(readText(thread, "SourcePage")))}
        ${item("详情状态", `<span class="next02-sh-detail-status-pill" data-status="${attr(status)}">${escapeHtml(formatSHDetailStatus(status))}</span>`)}
        ${item("作者", escapeHtml(readText(thread, "AuthorName")))}
        ${item("发布时间", renderSehuatangTime(pick(thread, "PostTime")))}
        ${item("初始获取时间", renderSehuatangTime(pick(thread, "FirstSeenTime")))}
        ${item("最近抓到", renderSehuatangTime(pick(thread, "LastSeenTime")))}
        ${item("详情抓取时间", renderSehuatangTime(pick(thread, "DetailFetchedOn")))}
        ${readText(thread, "DetailError") ? item("详情错误", `<span class="next02-sh-detail-error">${escapeHtml(readText(thread, "DetailError"))}</span>`, "is-wide") : ""}
      </div>
    </div>
  </section>`;
}

function splitLines(value: unknown): string[] {
  return String(value || "").split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
}

function renderSHDetailCopyList(title: string, items: string[], copyAllLabel: string, kind: "magnet" | "hash"): string {
  return `<section class="card next02-sc-panel next02-sc-events-panel next02-sh-detail-panel next02-sh-detail-resource-panel is-${kind}">
    <div class="card-body">
      <div class="next02-sh-detail-section-head">
        <div><span class="next02-sh-section-kicker">${kind === "magnet" ? "MAGNET RESOURCES" : "INFO HASH RESOURCES"}</span><h5>${escapeHtml(title)}</h5></div>
        <div class="next02-sh-detail-section-actions">
          <span class="next02-sh-detail-count-pill">${escapeHtml(items.length)}</span>
          <button type="button" class="btn btn-sm btn-outline-secondary js-copy-value" data-copy-value="${attr(items.join("\n"))}" ${items.length ? "" : "disabled"}>${escapeHtml(copyAllLabel)}</button>
        </div>
      </div>
      ${items.length ? `<div class="next02-sh-detail-resource-list">${items.map((item) => `<div class="next02-sh-detail-resource-item"><code title="${attr(item)}">${escapeHtml(item)}</code><button type="button" class="btn btn-sm btn-outline-secondary js-copy-value" data-copy-value="${attr(item)}">复制</button></div>`).join("")}</div>` : `<div class="next02-sh-detail-empty">暂无${escapeHtml(title)}数据。</div>`}
    </div>
  </section>`;
}

function renderSHDetailContent(thread: AnyRecord): string {
  const content = readText(thread, "RawContentText");
  return `<section class="card next02-sc-panel next02-sc-events-panel next02-sh-detail-panel next02-sh-detail-content-panel">
    <div class="card-body">
      <div class="next02-sh-detail-section-head">
        <div><span class="next02-sh-section-kicker">THREAD CONTENT</span><h5>正文</h5></div>
        <button type="button" class="btn btn-sm btn-outline-secondary js-copy-value" data-copy-value="${attr(content)}" ${content ? "" : "disabled"}>复制正文</button>
      </div>
      ${content ? `<pre class="next02-sh-detail-content">${escapeHtml(content)}</pre>` : `<div class="next02-sh-detail-empty">当前帖子尚未抓取正文。</div>`}
    </div>
  </section>`;
}

function renderSHDetailFetchItems(items: unknown): string {
  const rows = asArray<AnyRecord>(items);
  return `<section class="card next02-sc-panel next02-sc-events-panel next02-sh-detail-panel next02-sh-detail-runs-panel">
    <div class="card-body">
      <div class="next02-sh-detail-section-head">
        <div><span class="next02-sh-section-kicker">FETCH HISTORY</span><h5>相关抓取记录</h5></div>
        <span class="next02-sh-detail-table-meta">${escapeHtml(rows.length)} Records</span>
      </div>
      <div class="table-responsive">
        <table class="table table-sm align-middle mb-0 next02-sh-detail-runs-table">
          <colgroup><col class="next02-sh-detail-run-id"><col class="next02-sh-detail-run-mode"><col class="next02-sh-detail-run-mode"><col class="next02-sh-detail-run-forum"><col class="next02-sh-detail-run-page"><col class="next02-sh-detail-run-status"><col class="next02-sh-detail-run-row"><col class="next02-sh-detail-run-time"><col class="next02-sh-detail-run-error"></colgroup>
          <thead><tr><th>run_id</th><th>运行来源</th><th>模式</th><th>来源版块</th><th>来源页</th><th>状态</th><th>thread_row_id</th><th>更新时间</th><th>错误</th></tr></thead>
          <tbody>${rows.length ? rows.map((item) => `<tr>
            <td>${escapeHtml(pick(item, "RunID") || "")}</td>
            <td>${escapeHtml(readText(item, "RunSource"))}</td>
            <td>${escapeHtml(readText(item, "FetchMode"))}</td>
            <td>${escapeHtml(formatSHForumName(readText(item, "SourceForumID")))}</td>
            <td>${readText(item, "SourcePageURL") ? `<a href="${attr(readText(item, "SourcePageURL"))}" target="_blank" rel="noopener noreferrer">${escapeHtml(readText(item, "SourcePage"))}</a>` : escapeHtml(readText(item, "SourcePage"))}</td>
            <td><span class="next02-sh-detail-run-status-pill" data-status="${attr(readText(item, "Status"))}">${escapeHtml(readText(item, "Status"))}</span></td>
            <td>${escapeHtml(pick(item, "ThreadRowID") || "")}</td>
            <td class="text-nowrap">${renderSehuatangTime(pick(item, "UpdatedOn"))}</td>
            <td class="next02-sh-detail-run-error-cell">${readText(item, "Error") ? `<details><summary>错误</summary><div class="mt-2">${escapeHtml(readText(item, "Error"))}</div></details>` : "-"}</td>
          </tr>`).join("") : `<tr><td colspan="9"><div class="fetch-site-empty-state">暂无相关抓取记录。</div></td></tr>`}</tbody>
        </table>
      </div>
    </div>
  </section>`;
}

function renderSHDetailPage(data: AnyRecord): string {
  const thread = asRecord(pick(data, "Thread"));
  const magnets = splitLines(pick(thread, "MagnetText"));
  const hashes = splitLines(pick(thread, "InfoHashText"));
  const status = readText(thread, "DetailStatus");
  const title = readText(thread, "ThreadTitle") || pick(data, "PageTitle") || "Sehuatang SH 帖子详情";
  return `<div id="shDetailPage" class="next02-sh-detail-rewrite">
    <section class="next02-sc-hero next02-sc-events-hero next02-sh-detail-hero next02-sc-movie-top-shell">
      <div class="next02-sc-hero-copy next02-sh-detail-hero-copy">
        <span class="next02-sc-eyebrow">SEHUATANG / SH THREAD DETAIL</span>
        <h4 title="${attr(title)}">${escapeHtml(title)}</h4>
      </div>
      <div class="next02-sh-detail-hero-meta" aria-label="帖子摘要">
        <span>Thread <strong>${escapeHtml(readText(thread, "ThreadTID"))}</strong></span>
        <span>版块 <strong>${escapeHtml(formatSHForumName(readText(thread, "SourceForumID")))}</strong></span>
        <span class="is-status" data-status="${attr(status)}">${escapeHtml(formatSHDetailStatus(status))}</span>
        <span>${renderSehuatangTime(pick(thread, "PostTime"))}</span>
      </div>
      <div class="next02-sc-hero-actions next02-sh-detail-actions">
        <a class="btn btn-outline-primary btn-sm" href="${attr(pick(data, "ListPageURL") || "/shpool")}">返回列表</a>
        ${readText(thread, "ThreadURL") ? `<a class="btn btn-outline-secondary btn-sm" href="${attr(readText(thread, "ThreadURL"))}" target="_blank" rel="noopener noreferrer">打开原帖</a>` : ""}
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "TaskPageURL") || "/shfill")}">详情补全</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "ListTaskPageURL") || "/shfetch")}">列表抓取</a>
        <a class="btn btn-outline-secondary btn-sm" href="${attr(pick(data, "TasksPageURL") || "/crawltasklist")}">任务列表</a>
      </div>
    </section>
    <div id="shDetailMsg" class="alert d-none" role="alert"></div>
    ${renderSHDetailInfo(thread)}
    ${renderSHDetailContent(thread)}
    <div class="next02-sh-detail-resource-grid">
      ${renderSHDetailCopyList("Magnet", magnets, "复制全部 Magnet", "magnet")}
      ${renderSHDetailCopyList("Info Hash", hashes, "复制全部 Hash", "hash")}
    </div>
    ${renderSHDetailFetchItems(pick(data, "FetchItems"))}
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
  const page = document.getElementById("javbusListPage") || document.getElementById("sukebeiListPage") || document.getElementById("fetchSiteListPage");
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
    const page = document.getElementById("javbusListPage") || document.getElementById("sukebeiListPage") || document.getElementById("fetchSiteListPage");
    const taskPageURL = String(page?.getAttribute("data-task-page-url") || "/sitefetch").trim();
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

let shDetailTaskStarting = false;
let shDeleteStarting = false;
let pendingSHDeleteThreadTIDs: string[] = [];

function isSHPoolActionStarting(): boolean {
  return shDetailTaskStarting || shDeleteStarting;
}

function shDetailSelectionInputs(): HTMLInputElement[] {
  return Array.from(document.querySelectorAll<HTMLInputElement>(".js-sh-detail-select"));
}

function selectedSHDetailThreadTIDs(): string[] {
  return Array.from(new Set(shDetailSelectionInputs()
    .filter((input) => input.checked)
    .map((input) => input.value.trim())
    .filter(Boolean)));
}

function refreshSHDetailSelection() {
  const inputs = shDetailSelectionInputs();
  const selectedCount = inputs.filter((input) => input.checked).length;
  const actionStarting = isSHPoolActionStarting();
  const count = document.getElementById("shDetailSelectedCount");
  if (count) {
    count.textContent = String(selectedCount);
  }
  inputs.forEach((input) => {
    input.closest("tr")?.classList.toggle("is-selected", input.checked);
  });
  document.querySelectorAll<HTMLButtonElement>(".js-sh-detail-select-all").forEach((button) => {
    button.disabled = actionStarting || inputs.length === 0 || selectedCount === inputs.length;
  });
  document.querySelectorAll<HTMLButtonElement>(".js-sh-detail-clear-selection, .js-sh-detail-start-selected, .js-sh-delete-selected").forEach((button) => {
    button.disabled = actionStarting || selectedCount === 0;
  });
}

function refreshSHPoolActionState() {
  const actionStarting = isSHPoolActionStarting();
  shDetailSelectionInputs().forEach((input) => {
    input.disabled = actionStarting || !input.value.trim();
  });
  document.querySelectorAll<HTMLButtonElement>(".js-sh-detail-start, .js-sh-delete").forEach((button) => {
    button.disabled = actionStarting || !String(button.getAttribute("data-thread-tid") || "").trim();
  });
  document.querySelectorAll<HTMLButtonElement>(".js-sh-delete-confirm").forEach((button) => {
    button.disabled = actionStarting || pendingSHDeleteThreadTIDs.length === 0;
  });
  refreshSHDetailSelection();
}

function setSHDetailTaskStarting(starting: boolean) {
  shDetailTaskStarting = starting;
  refreshSHPoolActionState();
}

function setSHDeleteStarting(starting: boolean) {
  shDeleteStarting = starting;
  refreshSHPoolActionState();
}

function openSHDeleteModal(threadTIDs: string[]) {
  pendingSHDeleteThreadTIDs = Array.from(new Set(threadTIDs.map((item) => item.trim()).filter(Boolean)));
  if (!pendingSHDeleteThreadTIDs.length) {
    showMsg("请选择要删除的帖子", false);
    return;
  }
  const description = document.getElementById("shThreadDeleteDescription");
  if (description) {
    description.textContent = pendingSHDeleteThreadTIDs.length === 1
      ? `将从帖子池删除帖子 ${pendingSHDeleteThreadTIDs[0]}。`
      : `将从帖子池删除选中的 ${pendingSHDeleteThreadTIDs.length} 条帖子。`;
  }
  refreshSHPoolActionState();
  openModal(document.getElementById("shThreadDeleteModal"));
}

async function deleteSHThreads() {
  const threadTIDs = [...pendingSHDeleteThreadTIDs];
  if (!threadTIDs.length || isSHPoolActionStarting()) {
    return;
  }
  setSHDeleteStarting(true);
  closeModal(document.getElementById("shThreadDeleteModal"));
  try {
    const data = asRecord(await sendJSON("/api/sh-threads/batch-delete", "POST", { thread_tids: threadTIDs }));
    const deletedCount = Number(pick(data, "deleted_count") || 0);
    showMsg(`已删除 ${deletedCount} 条帖子，正在刷新列表`, true);
    window.location.reload();
  } catch (err) {
    showMsg(err instanceof Error ? err.message : "删除帖子失败", false);
  } finally {
    pendingSHDeleteThreadTIDs = [];
    setSHDeleteStarting(false);
  }
}

async function startSHDetailTask(threadTIDs: string[]) {
  const tids = Array.from(new Set(threadTIDs.map((item) => item.trim()).filter(Boolean)));
  if (!tids.length) {
    showMsg("请选择要抓取详情的帖子", false);
    return;
  }
  const progressWindow = window.open("/shfill", "_blank");
  if (!progressWindow) {
    showMsg("新页面被浏览器拦截，请允许本站打开新页面", false);
    return;
  }
  progressWindow.opener = null;
  setSHDetailTaskStarting(true);
  try {
    const query = asRecord(pick(payload.value, "Query"));
    const data = asRecord(await sendJSON("/api/crawler/jobs/start", "POST", {
      task_type: "spider_fetch_sehuatang_sh_thread_detail",
      trigger_source: "shpool",
      thread_tids: tids,
      target_forum_ids: shMultiQueryValues(query, "SourceForumIDs", "SourceForumID"),
      detail_statuses: shMultiQueryValues(query, "DetailStatuses", "DetailStatus"),
      movie_types: shMultiQueryValues(query, "MovieTypes", "MovieType"),
      keyword: String(pick(query, "Keyword") || "").trim(),
      post_time_from: String(pick(query, "PostTimeFrom") || "").trim(),
      post_time_to: String(pick(query, "PostTimeTo") || "").trim()
    }));
    const jobID = readText(data, "job_id", "JobID");
    if (!jobID) {
      progressWindow.close();
      showMsg("详情抓取任务已创建，但缺少 job_id", false);
      return;
    }
    if (progressWindow.closed) {
      showMsg(`详情抓取任务已创建（${tids.length} 条），新页面已关闭`, true);
      return;
    }
    progressWindow.location.replace(`/shfill?job_id=${encodeURIComponent(jobID)}`);
    showMsg(`详情抓取任务已创建，已在新页面打开（${tids.length} 条）`, true);
  } catch (err) {
    progressWindow.close();
    showMsg(err instanceof Error ? err.message : "详情抓取任务启动失败", false);
  } finally {
    setSHDetailTaskStarting(false);
  }
}

function handlePageChange(event: Event) {
  const target = event.target instanceof Element ? event.target : null;
  if (target?.closest(".js-sh-detail-select")) {
    refreshSHDetailSelection();
  }
}

function closeModal(modal: HTMLElement | null) {
  if (!modal) {
    return;
  }
  const activeElement = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  if (activeElement && modal.contains(activeElement)) {
    activeElement.blur();
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
  const selectAllSHDetailsButton = target.closest(".js-sh-detail-select-all");
  if (selectAllSHDetailsButton) {
    shDetailSelectionInputs().forEach((input) => {
      input.checked = true;
    });
    refreshSHDetailSelection();
    return;
  }
  const clearSHDetailsButton = target.closest(".js-sh-detail-clear-selection");
  if (clearSHDetailsButton) {
    shDetailSelectionInputs().forEach((input) => {
      input.checked = false;
    });
    refreshSHDetailSelection();
    return;
  }
  const startSelectedSHDetailsButton = target.closest(".js-sh-detail-start-selected");
  if (startSelectedSHDetailsButton) {
    void startSHDetailTask(selectedSHDetailThreadTIDs());
    return;
  }
  const deleteSelectedSHThreadsButton = target.closest(".js-sh-delete-selected");
  if (deleteSelectedSHThreadsButton) {
    openSHDeleteModal(selectedSHDetailThreadTIDs());
    return;
  }
  const deleteSHThreadButton = target.closest(".js-sh-delete");
  if (deleteSHThreadButton) {
    openSHDeleteModal([String(deleteSHThreadButton.getAttribute("data-thread-tid") || "")]);
    return;
  }
  const confirmDeleteSHThreadsButton = target.closest(".js-sh-delete-confirm");
  if (confirmDeleteSHThreadsButton) {
    void deleteSHThreads();
    return;
  }
  const startSHDetailButton = target.closest(".js-sh-detail-start");
  if (startSHDetailButton) {
    void startSHDetailTask([String(startSHDetailButton.getAttribute("data-thread-tid") || "")]);
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
}
</script>
