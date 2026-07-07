<template>
  <PageShell class="crawler-jobs-page-shell" :title="title" :note="note" eyebrow="">
    <div id="page-root" class="container-fluid crawler-jobs-page-root">
      <LoadingState v-if="loading" />
      <ErrorPanel v-else-if="error" :message="error" />
      <template v-else>
        <div id="crawlerJobsPage">
          <div class="row">
            <div class="col-12">
              <div class="page-title-box d-flex align-items-center justify-content-between">
                <div>
                  <h4 class="mb-0">{{ title }}</h4>
                  <div v-if="note" class="text-muted small mt-1">{{ note }}</div>
                </div>
                <div class="task-link-buttons">
                  <a v-for="item in allHeaderLinks" :key="item.href" :href="item.href" class="btn btn-outline-secondary btn-sm">{{ item.label }}</a>
                </div>
              </div>
            </div>
          </div>

          <div class="row">
            <div class="col-lg-5">
              <div class="card">
                <div class="card-body">
                  <h4 class="card-title mb-3">{{ taskPanelTitle }}</h4>
                  <div v-if="taskHelperText" class="alert alert-info small py-2 px-3 mb-3">{{ taskHelperText }}</div>
                  <form id="crawlerTaskForm" :class="{ 'is-hidden': hideTaskForm }" @submit.prevent="startDefault">
                    <input type="hidden" id="task_type" name="task_type" :value="selectedTaskType">
                    <div class="form-grid">
                      <div class="form-group compact-field grid-span-2">
                        <label>触发任务</label>
                        <div class="btn-group btn-group-toggle d-flex flex-wrap task-type-toggle" role="group" aria-label="crawler task type">
                          <button
                            v-for="task in taskButtons"
                            :key="taskType(task)"
                            type="button"
                            class="btn btn-outline-primary"
                            :class="{ active: selectedTaskType === taskType(task) }"
                            :data-task-type="taskType(task)"
                            @click="selectedTaskType = taskType(task)"
                          >
                            {{ taskLabel(task) }}
                          </button>
                        </div>
                      </div>

                      <div class="form-group compact-field grid-span-2" id="nameFieldWrap" :class="{ 'is-hidden': !showNameField }">
                        <label for="task_name">Seed 名称</label>
                        <input id="task_name" v-model="form.name" name="name" class="form-control" type="text" placeholder="输入 seed 名称">
                      </div>

                      <div class="form-group compact-field grid-span-2" id="numberFieldWrap" :class="{ 'is-hidden': !showNumberField }">
                        <label for="task_number">刷新数量</label>
                        <input id="task_number" v-model.number="form.number" name="number" class="form-control" type="number" placeholder="输入刷新数量">
                      </div>

                      <div class="form-group compact-field grid-span-2" id="actorNameFieldWrap" :class="{ 'is-hidden': !showActorNameField }">
                        <label for="task_actor_name">演员名称</label>
                        <input id="task_actor_name" v-model="form.actorName" name="actor_name" class="form-control" type="text" placeholder="输入演员名称">
                      </div>

                      <div class="form-group compact-field grid-span-2" id="translateStatusesFieldWrap" :class="{ 'is-hidden': !showTranslateStatusesField }">
                        <label class="d-block mb-2">标题翻译状态（可多选）</label>
                        <div class="status-check-grid">
                          <div class="form-check">
                            <input id="translate_status_1" name="translate_statuses" class="form-check-input" type="checkbox" value="1" checked>
                            <label class="form-check-label" for="translate_status_1">无中文字幕(1)</label>
                          </div>
                          <div class="form-check">
                            <input id="translate_status_2" name="translate_statuses" class="form-check-input" type="checkbox" value="2">
                            <label class="form-check-label" for="translate_status_2">有中文字幕(2)</label>
                          </div>
                          <div class="form-check">
                            <input id="translate_status_3" name="translate_statuses" class="form-check-input" type="checkbox" value="3">
                            <label class="form-check-label" for="translate_status_3">翻译错误(3)</label>
                          </div>
                          <div class="form-check">
                            <input id="translate_status_4" name="translate_statuses" class="form-check-input" type="checkbox" value="4">
                            <label class="form-check-label" for="translate_status_4">敏感词(4)</label>
                          </div>
                        </div>
                      </div>

                      <div class="form-group compact-field grid-span-2" id="autoFetchSiteFieldWrap" :class="{ 'is-hidden': !showAutoFetchSiteField }">
                        <label class="d-block mb-2">详情后自动抓取外站</label>
                        <div class="form-check">
                          <input id="task_auto_fetch_site" name="auto_fetch_site" class="form-check-input" type="checkbox" checked>
                          <label class="form-check-label" for="task_auto_fetch_site">detail 解析完成后自动抓取 JavBus / Sukebei</label>
                        </div>
                      </div>

                      <div class="form-group compact-field grid-span-2" id="fetchSehuatangListURLFieldWrap" :class="{ 'is-hidden': !showListUrlField }">
                        <label for="task_sehuatang_list_url">列表页 URL</label>
                        <input id="task_sehuatang_list_url" v-model="form.listURL" name="list_url" class="form-control" type="text" placeholder="输入 色花堂列表页 URL">
                      </div>

                      <div class="form-group compact-field grid-span-2" id="fetchSehuatangKeywordFieldWrap" :class="{ 'is-hidden': !showKeywordField }">
                        <label for="task_sehuatang_keyword">标题关键词</label>
                        <input id="task_sehuatang_keyword" v-model="form.keyword" name="keyword" class="form-control" type="text" placeholder="可选，留空表示抓取全部普通帖">
                      </div>

                      <div class="form-group compact-field grid-span-2" id="fetchSehuatangSHModeFieldWrap" :class="{ 'is-hidden': !showSHModeField }">
                        <label for="task_sehuatang_sh_fetch_mode">抓取模式</label>
                        <select id="task_sehuatang_sh_fetch_mode" v-model="form.fetchMode" name="fetch_mode" class="form-select">
                          <option value="search">搜索结果翻页</option>
                          <option value="forum">版块列表翻页</option>
                        </select>
                      </div>

                      <div class="form-group compact-field grid-span-2" id="fetchSehuatangSHForumsFieldWrap" :class="{ 'is-hidden': !showSHForumsField }">
                        <label class="d-block mb-2">目标版块</label>
                        <div class="status-check-grid">
                          <div class="form-check">
                            <input id="task_sehuatang_sh_forum_36" name="target_forum_ids" class="form-check-input" type="checkbox" value="36" :checked="targetForumChecked('36')">
                            <label class="form-check-label" for="task_sehuatang_sh_forum_36">亚洲无码原创</label>
                          </div>
                          <div class="form-check">
                            <input id="task_sehuatang_sh_forum_103" name="target_forum_ids" class="form-check-input" type="checkbox" value="103" :checked="targetForumChecked('103')">
                            <label class="form-check-label" for="task_sehuatang_sh_forum_103">高清中文字幕</label>
                          </div>
                        </div>
                      </div>

                      <div class="form-group compact-field grid-span-2" id="fetchSehuatangSHDetailStatusFieldWrap" :class="{ 'is-hidden': !showSHDetailFields }">
                        <label class="d-block mb-2">详情状态</label>
                        <div class="status-check-grid">
                          <div class="form-check">
                            <input id="task_sehuatang_sh_detail_status_pending" name="detail_statuses" class="form-check-input" type="checkbox" value="pending" :checked="detailStatusChecked('pending')">
                            <label class="form-check-label" for="task_sehuatang_sh_detail_status_pending">待抓详情</label>
                          </div>
                          <div class="form-check">
                            <input id="task_sehuatang_sh_detail_status_failed" name="detail_statuses" class="form-check-input" type="checkbox" value="failed" :checked="detailStatusChecked('failed')">
                            <label class="form-check-label" for="task_sehuatang_sh_detail_status_failed">详情失败</label>
                          </div>
                          <div class="form-check">
                            <input id="task_sehuatang_sh_detail_status_fetched" name="detail_statuses" class="form-check-input" type="checkbox" value="fetched" :checked="detailStatusChecked('fetched')">
                            <label class="form-check-label" for="task_sehuatang_sh_detail_status_fetched">已抓详情</label>
                          </div>
                        </div>
                      </div>

                      <div class="form-group compact-field grid-span-2" id="fetchSehuatangSHDetailTitleFieldWrap" :class="{ 'is-hidden': !showSHDetailFields }">
                        <label for="task_sehuatang_sh_detail_keyword">帖子标题筛选</label>
                        <input id="task_sehuatang_sh_detail_keyword" v-model="form.detailKeyword" name="keyword" class="form-control" type="text" placeholder="按标题筛选待抓详情帖子">
                      </div>

                      <div class="form-group compact-field" id="fetchSehuatangSHPostTimeFromFieldWrap" :class="{ 'is-hidden': !showSHDetailFields }">
                        <label for="task_sehuatang_sh_post_time_from">发帖时间开始</label>
                        <input id="task_sehuatang_sh_post_time_from" v-model="form.postTimeFrom" name="post_time_from" class="form-control" type="text" placeholder="YYYY-MM-DD">
                      </div>

                      <div class="form-group compact-field" id="fetchSehuatangSHPostTimeToFieldWrap" :class="{ 'is-hidden': !showSHDetailFields }">
                        <label for="task_sehuatang_sh_post_time_to">发帖时间结束</label>
                        <input id="task_sehuatang_sh_post_time_to" v-model="form.postTimeTo" name="post_time_to" class="form-control" type="text" placeholder="YYYY-MM-DD">
                      </div>

                      <div class="form-group compact-field" id="fetchSehuatangSHDetailLimitFieldWrap" :class="{ 'is-hidden': !showSHDetailFields }">
                        <label for="task_sehuatang_sh_detail_limit">预览 limit</label>
                        <input id="task_sehuatang_sh_detail_limit" v-model.number="form.detailLimit" name="limit" class="form-control" type="number" min="1" placeholder="留空不限制">
                      </div>

                      <div class="form-group compact-field grid-span-2" id="fetchSehuatangStopOnExistingFieldWrap" :class="{ 'is-hidden': !showOldSehuatangFields }">
                        <label class="d-block mb-2">抓取范围</label>
                        <div class="form-check">
                          <input id="task_sehuatang_stop_on_existing" name="stop_on_existing" class="form-check-input" type="checkbox" value="1" :checked="Boolean(pick(payload, 'FetchSehuatangStopOnExisting'))">
                          <label class="form-check-label" for="task_sehuatang_stop_on_existing">从第一页开始，抓到已存在帖子停止</label>
                        </div>
                      </div>

                      <div class="form-group compact-field" id="fetchSehuatangStartPageFieldWrap" :class="{ 'is-hidden': !showPagingFields }">
                        <label for="task_sehuatang_start_page">起始页</label>
                        <input id="task_sehuatang_start_page" v-model.number="form.startPage" name="start_page" class="form-control" type="number" placeholder="输入起始页">
                      </div>

                      <div class="form-group compact-field" id="fetchSehuatangEndPageFieldWrap" :class="{ 'is-hidden': !showPagingFields }">
                        <label for="task_sehuatang_end_page">结束页</label>
                        <input id="task_sehuatang_end_page" v-model.number="form.endPage" name="end_page" class="form-control" type="number" placeholder="输入结束页">
                      </div>

                      <div class="form-group compact-field grid-span-2" id="fetchSehuatangPersistModeFieldWrap" :class="{ 'is-hidden': !showOldSehuatangFields }">
                        <label for="task_sehuatang_persist_mode">入库模式</label>
                        <select id="task_sehuatang_persist_mode" v-model="form.persistMode" name="persist_mode" class="form-select">
                          <option value="upsert_all">upsert 所有帖子</option>
                          <option value="skip_existing">跳过数据库中已存在的帖子</option>
                        </select>
                      </div>
                    </div>

                    <div v-if="showFetchSitePanel" id="fetchSiteFilterPanel" class="mt-3">
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
                            <MovieCardFilterForm :filter="filter" />
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
                                <button v-for="option in fetchSiteOrderOptions" :key="option.value" type="button" class="btn btn-outline-primary flex-fill" :class="{ active: text(filter, 'OrderBy') === option.value }" :data-fetch-site-orderby="option.value">{{ option.label }}</button>
                              </div>
                            </div>
                            <div class="col-12 col-md-6 col-xl-2">
                              <label class="form-label mb-1">排序顺序</label>
                              <select class="form-select" name="order" :value="text(filter, 'Order')">
                                <option value="">默认</option>
                                <option value="asc">升序</option>
                                <option value="desc">降序</option>
                              </select>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>

                    <div class="d-flex align-items-center flex-wrap action-row">
                      <button v-if="showFetchSitePanel" type="button" id="previewFetchSiteBtn" class="btn btn-outline-info mr-2">预览筛选结果</button>
                      <button v-if="showSHDetailFields" type="button" id="previewSHDetailBtn" class="btn btn-outline-info mr-2">筛选帖子预览</button>
                      <button type="submit" id="submitBtn" class="btn btn-primary">{{ showSHDetailFields ? "抓取选中详情" : "开始执行" }}</button>
                      <button type="button" id="pauseBtn" class="btn btn-outline-warning ml-2" disabled>暂停</button>
                      <button type="button" id="resumeBtn" class="btn btn-outline-success ml-2" disabled>继续</button>
                      <button type="button" id="stopBtn" class="btn btn-outline-danger ml-2" disabled>停止</button>
                    </div>
                  </form>
                  <p v-if="message" class="task-message">{{ message }}</p>
                </div>
              </div>

              <div class="card">
                <div class="card-body">
                  <h4 class="card-title mb-3">运行概览</h4>
                  <div class="crawl-overview">
                    <div class="status-grid" :class="{ 'sehuatang-progress-grid': overviewMode === 'sehuatang_progress' }">
                      <div v-for="item in overviewItems" :key="item.label" class="overview-item">
                        <div class="label">{{ item.label }}</div>
                        <div class="value">{{ item.value }}</div>
                      </div>
                    </div>
                    <div v-if="showFilteredQueuePanel" class="queue-panel mt-2">
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
                    <div class="progress mt-3">
                      <div id="progressBar" class="progress-bar progress-bar-striped progress-bar-animated" role="progressbar" style="width: 0%" aria-valuemin="0" aria-valuemax="100" aria-valuenow="0">0%</div>
                    </div>
                    <div class="mt-3 d-flex align-items-center justify-content-between">
                      <span id="statusMsg" class="text-muted">{{ emptyStateText }}</span>
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
                    <h4 class="card-title mb-0">{{ eventTitle }}</h4>
                    <div class="custom-control custom-checkbox">
                      <input type="checkbox" class="custom-control-input" id="debugRaw">
                      <label class="custom-control-label" for="debugRaw">显示原始 JSON</label>
                    </div>
                  </div>
                  <div class="result-box">
                    <pre id="resultBox">{{ emptyStateText }}</pre>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="showFetchSitePanel" class="row">
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
          </div>

          <div v-if="showSHDetailFields" class="row">
            <div class="col-12">
              <div class="card is-hidden" id="shDetailPreviewPanel">
                <div class="card-body">
                  <div class="d-flex align-items-center justify-content-between mb-3">
                    <div>
                      <h4 class="card-title mb-1">帖子预览</h4>
                      <div class="text-muted small" id="shDetailPreviewSummary">当前还没有预览数据</div>
                    </div>
                    <div class="d-flex align-items-center">
                      <span class="badge bg-light text-dark mr-2" id="shDetailPreviewMetaText">共 0 条，已选 0 条</span>
                      <button type="button" id="selectAllSHDetailPreviewBtn" class="btn btn-outline-secondary btn-sm mr-2">全选</button>
                      <button type="button" id="clearSHDetailPreviewBtn" class="btn btn-outline-secondary btn-sm mr-2">取消选择</button>
                      <button type="button" id="hideSHDetailPreviewBtn" class="btn btn-outline-secondary btn-sm">收起</button>
                    </div>
                  </div>
                  <div class="table-responsive">
                    <table class="table table-sm table-hover align-middle mb-0">
                      <thead>
                        <tr>
                          <th class="text-nowrap">选择</th>
                          <th class="text-nowrap">TID</th>
                          <th>标题</th>
                          <th class="text-nowrap">版块</th>
                          <th class="text-nowrap">发布时间</th>
                          <th class="text-nowrap">详情状态</th>
                          <th class="text-nowrap">作者</th>
                        </tr>
                      </thead>
                      <tbody id="shDetailPreviewTableBody">
                        <tr><td colspan="7" class="text-center text-muted py-4">等待预览</td></tr>
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import ErrorPanel from "@/components/ErrorPanel.vue";
import LoadingState from "@/components/LoadingState.vue";
import MovieCardFilterForm from "@/components/MovieCardFilterForm.vue";
import PageShell from "@/components/PageShell.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, pick, text } from "@/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const message = ref("");
const selectedTaskType = ref("");
const form = reactive({
  name: "",
  number: 1000000,
  actorName: "",
  startPage: 0,
  endPage: 0,
  keyword: "",
  listURL: "",
  fetchMode: "search",
  persistMode: "skip_existing",
  detailKeyword: "",
  postTimeFrom: "",
  postTimeTo: "",
  detailLimit: undefined as number | undefined
});

const fetchSiteTaskTypes = new Set([
  "spider_fetch_javbus_resources",
  "spider_fetch_sukebei_resources",
  "spider_fetch_site_both_resources",
  "spider_fetch_javbus_filtered_resources",
  "spider_fetch_sukebei_filtered_resources"
]);

const autoFetchSiteTaskTypes = new Set([
  "spider_daily_best",
  "spider_daily_best_sync",
  "spider_seeds",
  "spider_seed_by_name"
]);

const fetchSiteOrderOptions = [
  { value: "du", label: "Detail更新" },
  { value: "rd", label: "上映日期" },
  { value: "vw", label: "观看人数" },
  { value: "rk", label: "上榜日期" },
  { value: "drk", label: "在榜天数" },
  { value: "hrk", label: "最高排名" },
  { value: "bt", label: "下载时间" },
  { value: "mbt", label: "Media下载时间" },
  { value: "lsct", label: "最后sc" },
  { value: "sc", label: "sc次数" },
  { value: "co", label: "co次数" }
];

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  message.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
    selectedTaskType.value = text(payload.value, "DefaultTaskType");
    form.keyword = text(payload.value, "FetchSehuatangKey");
    form.listURL = text(payload.value, "FetchSehuatangList");
    form.startPage = Number(text(payload.value, "FetchSehuatangStartPage") || 0);
    form.endPage = Number(text(payload.value, "FetchSehuatangEndPage") || 0);
    form.fetchMode = text(payload.value, "FetchSehuatangSHMode") === "forum" ? "forum" : "search";
    form.persistMode = text(payload.value, "FetchSehuatangPersistMode") === "upsert_all" ? "upsert_all" : "skip_existing";
    form.number = 1000000;
    form.detailKeyword = "";
    form.postTimeFrom = "";
    form.postTimeTo = "";
    form.detailLimit = undefined;
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "PageTitle", "Title") || String(route.meta.title || "Trigger"));
const note = computed(() => text(payload.value, "PageNote"));
const taskPanelTitle = computed(() => text(payload.value, "TaskPanelTitle") || "任务触发");
const eventTitle = computed(() => text(payload.value, "EventTitle") || "事件流");
const taskHelperText = computed(() => text(payload.value, "TaskHelperText"));
const taskButtons = computed(() => asArray<AnyRecord>(pick(payload.value, "TaskButtons")));
const hideTaskForm = computed(() => Boolean(pick(payload.value, "HideTaskForm")));
const headerLinks = computed(() => asArray<AnyRecord>(pick(payload.value, "HeaderLinks")).map((item) => ({
  label: text(item, "Label"),
  href: text(item, "Href")
})));
const allHeaderLinks = computed(() => {
  const links = [...headerLinks.value];
  const ensure = (href: string, label: string) => {
    if (!links.some((item) => item.href === href)) {
      links.push({ href, label });
    }
  };
  ensure("/crawler/tasks", "任务列表");
  ensure("/triggers/sc", "SC 预览与提交");
  return links;
});
const filter = computed(() => asRecord(pick(payload.value, "MovieCardFilter")));
const showMovieFilter = computed(() => Boolean(payload.value.ShowMovieFilters));
const emptyStateText = computed(() => text(payload.value, "EmptyStateText") || "等待任务触发");
const currentTaskType = computed(() => selectedTaskType.value || text(payload.value, "DefaultTaskType"));
const isFetchSiteTask = computed(() => fetchSiteTaskTypes.has(currentTaskType.value));
const isOldSehuatangTask = computed(() => currentTaskType.value === "spider_fetch_sehuatang_magnets");
const isSehuatangSHTask = computed(() => currentTaskType.value === "spider_fetch_sehuatang_sh_threads");
const isSehuatangSHListTask = computed(() => currentTaskType.value === "spider_fetch_sehuatang_sh_thread_list");
const isSehuatangSHDetailTask = computed(() => currentTaskType.value === "spider_fetch_sehuatang_sh_thread_detail");
const showFetchSitePanel = computed(() => isFetchSiteTask.value && showMovieFilter.value);
const showNameField = computed(() => currentTaskType.value === "spider_seed_by_name");
const showNumberField = computed(() => [
  "spider_refresh_oldest_detail",
  "spider_fetch_javbus_resources",
  "spider_fetch_sukebei_resources",
  "spider_fetch_site_both_resources"
].includes(currentTaskType.value));
const showActorNameField = computed(() => currentTaskType.value === "spider_rebuild_actor_rank");
const showOldSehuatangFields = computed(() => isOldSehuatangTask.value);
const showSHModeField = computed(() => isSehuatangSHTask.value);
const showSHForumsField = computed(() => isSehuatangSHTask.value || isSehuatangSHListTask.value || isSehuatangSHDetailTask.value);
const showSHDetailFields = computed(() => isSehuatangSHDetailTask.value);
const showPagingFields = computed(() => isOldSehuatangTask.value || isSehuatangSHTask.value || isSehuatangSHListTask.value);
const showKeywordField = computed(() => isOldSehuatangTask.value || (isSehuatangSHTask.value && form.fetchMode === "search"));
const showListUrlField = computed(() => isOldSehuatangTask.value);
const showAutoFetchSiteField = computed(() => autoFetchSiteTaskTypes.has(currentTaskType.value));
const showTranslateStatusesField = computed(() => currentTaskType.value === "spider_translate_title" || currentTaskType.value === "spider_post_process");
const overviewMode = computed(() => text(payload.value, "OverviewExtraMode"));
const showFilteredQueuePanel = computed(() => ["javbus_filtered_queue", "sukebei_filtered_queue"].includes(overviewMode.value));
const overviewItems = computed(() => {
  const mode = overviewMode.value;
  if (mode === "fetch_site_summary") {
    return [
      { label: "JavBus 进度（总/完成）", value: "0 / 0" },
      { label: "JavBus 结果（成功/失败）", value: "0 / 0" },
      { label: "Sukebei 进度（总/完成）", value: "0 / 0" },
      { label: "Sukebei 结果（成功/失败）", value: "0 / 0" }
    ];
  }
  if (mode === "sehuatang_progress") {
    return [
      { label: "页面进度（完成/总）", value: "0 / 0" },
      { label: "当前页面", value: "-" },
      { label: "已运行时长", value: "00:00" },
      { label: "发现帖子", value: "0" },
      { label: "目标命中", value: "0" },
      { label: "跳过版块", value: "0" },
      { label: "跳过已存在", value: "0" },
      { label: "入库", value: "0" },
      { label: "失败", value: "0" }
    ];
  }
  if (mode === "javbus_filtered_queue" || mode === "sukebei_filtered_queue") {
    const site = mode === "javbus_filtered_queue" ? "JavBus" : "Sukebei";
    return [
      { label: `${site} 队列总数`, value: "0" },
      { label: "待入列", value: "0" },
      { label: "成功", value: "0" },
      { label: "失败", value: "0" }
    ];
  }
  if (mode === "dailybest_stages" || mode === "seeds_stages") {
    const firstLabel = mode === "seeds_stages" ? "种子抓取" : "榜单抓取";
    return [
      { label: "当前阶段", value: "-" },
      { label: "已运行时长", value: "00:00" },
      { label: `${firstLabel}进度`, value: "-" },
      { label: `${firstLabel}正确/错误`, value: "0 / 0" },
      { label: "详情抓取进度", value: "-" },
      { label: "详情抓取正确/错误", value: "0 / 0" },
      { label: "封面下载进度", value: "-" },
      { label: "封面下载正确/错误", value: "0 / 0" },
      { label: "标题翻译进度", value: "-" },
      { label: "标题翻译正确/错误", value: "0 / 0" },
      { label: "JavBus 抓取进度", value: "-" },
      { label: "JavBus 抓取正确/错误", value: "0 / 0" },
      { label: "Sukebei 抓取进度", value: "-" },
      { label: "Sukebei 抓取正确/错误", value: "0 / 0" }
    ];
  }
  if (mode === "post_process_stages") {
    return [
      { label: "当前阶段", value: "- / 已运行时长 00:00" },
      { label: "封面下载进度", value: "-" },
      { label: "封面下载正确/错误", value: "0 / 0" },
      { label: "标题翻译进度", value: "-" },
      { label: "标题翻译正确/错误", value: "0 / 0" }
    ];
  }
  return [
    { label: text(asRecord(pick(payload.value, "Labels")), "Extra") || "任务类型", value: "-" },
    { label: text(asRecord(pick(payload.value, "Labels")), "Result") || "成功/失败", value: "0 / 0" },
    { label: text(asRecord(pick(payload.value, "Labels")), "Elapsed") || "已运行时长", value: "00:00" }
  ];
});

const targetForums = computed(() => {
  const values = asArray<string>(pick(payload.value, "FetchSehuatangSHTargetForums")).map((item) => String(item || "").trim()).filter(Boolean);
  return values.length ? values : ["36", "103"];
});

const shDetailStatuses = computed(() => {
  const values = asArray<string>(pick(payload.value, "FetchSehuatangSHDetailStatuses")).map((item) => String(item || "").trim()).filter(Boolean);
  if (values.length) {
    return values;
  }
  const fallback = text(payload.value, "FetchSehuatangSHDetailStatus");
  return ["failed", "fetched"].includes(fallback) ? [fallback] : ["pending"];
});

function targetForumChecked(id: string): boolean {
  return targetForums.value.includes(id);
}

function detailStatusChecked(status: string): boolean {
  return shDetailStatuses.value.includes(status);
}

function taskType(task: AnyRecord): string {
  return text(task, "TaskType", "task_type");
}

function taskLabel(task: AnyRecord): string {
  return text(task, "Label", "label") || taskType(task);
}

async function startDefault() {
  await startTask({ TaskType: selectedTaskType.value || text(payload.value, "DefaultTaskType"), Label: "默认任务" });
}

async function startTask(task: AnyRecord) {
  const type = taskType(task);
  if (!type) {
    message.value = "缺少 task_type";
    return;
  }
  const body = {
    task_type: type,
    name: form.name,
    number: form.number,
    actor_name: form.actorName,
    start_page: form.startPage,
    end_page: form.endPage,
    keyword: form.keyword,
    list_url: form.listURL,
    fetch_mode: form.fetchMode,
    persist_mode: form.persistMode,
    post_time_from: form.postTimeFrom,
    post_time_to: form.postTimeTo,
    limit: form.detailLimit
  };
  try {
    const data = asRecord(await sendJSON("/api/crawler/jobs/start", "POST", body));
    message.value = text(data, "message", "Message") || `已启动 ${type}`;
  } catch (err) {
    message.value = err instanceof Error ? err.message : String(err);
  }
}
</script>
