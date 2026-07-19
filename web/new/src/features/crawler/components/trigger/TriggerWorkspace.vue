<template>
  <PageShell class="crawler-jobs-page-shell" :title="title" :note="note" eyebrow="">
    <div
      id="page-root"
      class="container-fluid crawler-jobs-page-root"
      :class="{ 'next02-seeds-page-root next02-pink-trigger-page-root': isPinkTriggerRewritePage }"
    >
      <LoadingState v-if="loading" />
      <ErrorPanel v-else-if="error" :message="error" />
      <template v-else>
        <div
          id="crawlerJobsPage"
          class="next02-rewrite-shell next02-trigger-rewrite"
          :class="{
            'next02-sh-command-page': isSehuatangSHPage,
            'next02-fetch-site-command-page': isFetchSiteCommandPage,
            'next02-dailybest-command-page': isDailyBestCommandPage,
            'next02-seeds-command-page': isSeedsCommandPage,
            'next02-seeds-rewrite next02-pink-trigger-rewrite': isPinkTriggerRewritePage,
            'next02-sc-pink-theme': isPinkTriggerRewritePage,
            'next02-post-process-command-page': isPostProcessCommandPage,
            'next02-backfill-command-page': isBackfillCommandPage,
            'next02-sh-detail-command-page': showSHDetailFields
          }"
        >
          <div class="row">
            <div class="col-12">
              <div
                class="page-title-box next02-command-hero d-flex align-items-center justify-content-between"
                :class="{
                  'next02-seeds-hero next02-pink-trigger-hero next02-sc-movie-top-shell': isPinkTriggerRewritePage
                }"
              >
                <div class="next02-command-hero-main" :class="{ 'next02-sc-hero-copy': isPinkTriggerRewritePage }">
                  <div
                    v-if="isCommandVisualPage"
                    class="next02-hero-kicker"
                    :class="{ 'next02-sc-eyebrow': isPinkTriggerRewritePage }"
                  >
                    {{ commandHeroKicker }}
                  </div>
                  <h4 class="mb-0">{{ title }}</h4>
                  <div v-if="note" class="text-muted small mt-1 next02-hero-note">{{ note }}</div>
                  <div v-if="isCommandVisualPage" class="next02-hero-metrics" aria-label="page summary">
                    <span v-for="item in commandHeroMetrics" :key="item.label" class="next02-hero-metric">
                      <strong>{{ item.value }}</strong>
                      <span>{{ item.label }}</span>
                    </span>
                  </div>
                </div>
                <div
                  class="task-link-buttons next02-command-actions"
                  :class="{
                    'next02-sc-hero-actions next02-seeds-hero-actions next02-pink-trigger-hero-actions':
                      isPinkTriggerRewritePage
                  }"
                >
                  <a
                    v-for="item in allHeaderLinks"
                    :key="item.href"
                    :href="item.href"
                    class="btn btn-outline-secondary btn-sm"
                    :target="isSeedsCommandPage ? '_blank' : undefined"
                    :rel="isSeedsCommandPage ? 'noopener noreferrer' : undefined"
                    >{{ item.label }}</a
                  >
                </div>
              </div>
            </div>
          </div>

          <div
            class="row next02-command-grid"
            :class="{
              'next02-seeds-workspace next02-pink-trigger-workspace': isPinkTriggerRewritePage,
              'next02-pink-trigger-workspace-form-heavy': isFormHeavyTriggerPage
            }"
          >
            <div
              class="col-lg-5"
              :class="{ 'next02-seeds-control-column next02-pink-trigger-control-column': isPinkTriggerRewritePage }"
            >
              <TriggerForm
                :pink="isPinkTriggerRewritePage"
                :command-visual="isCommandVisualPage"
                :eyebrow="commandTaskEyebrow"
                :title="taskPanelTitle"
                :status="currentTaskStatusLabel"
                :helper="taskHelperText"
                :message="message"
              >
                <form id="crawlerTaskForm" :class="{ 'is-hidden': hideTaskForm }" @submit.prevent="startDefault">
                  <input type="hidden" id="task_type" name="task_type" :value="selectedTaskType" />
                  <div class="form-grid next02-control-grid">
                    <div class="form-group compact-field grid-span-2 next02-shfetchall-task-field">
                      <label>触发任务</label>
                      <div
                        class="btn-group btn-group-toggle d-flex flex-wrap task-type-toggle"
                        :class="{
                          'next02-seeds-task-segmented next02-pink-trigger-task-segmented': isPinkTriggerRewritePage
                        }"
                        role="group"
                        aria-label="crawler task type"
                      >
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

                    <div
                      class="form-group compact-field grid-span-2"
                      id="nameFieldWrap"
                      :class="{ 'is-hidden': !showNameField }"
                    >
                      <label for="task_name">Seed 名称</label>
                      <input
                        id="task_name"
                        v-model="form.name"
                        name="name"
                        class="form-control"
                        type="text"
                        placeholder="输入 seed 名称"
                      />
                    </div>

                    <div
                      class="form-group compact-field grid-span-2"
                      id="numberFieldWrap"
                      :class="{ 'is-hidden': !showNumberField }"
                    >
                      <label for="task_number">刷新数量</label>
                      <input
                        id="task_number"
                        v-model.number="form.number"
                        name="number"
                        class="form-control"
                        type="number"
                        placeholder="输入刷新数量"
                      />
                    </div>

                    <div
                      class="form-group compact-field grid-span-2"
                      id="actorNameFieldWrap"
                      :class="{ 'is-hidden': !showActorNameField }"
                    >
                      <label for="task_actor_name">演员名称</label>
                      <input
                        id="task_actor_name"
                        v-model="form.actorName"
                        name="actor_name"
                        class="form-control"
                        type="text"
                        placeholder="输入演员名称"
                      />
                    </div>

                    <div
                      class="form-group compact-field grid-span-2"
                      id="translateStatusesFieldWrap"
                      :class="{ 'is-hidden': !showTranslateStatusesField }"
                    >
                      <label class="d-block mb-2">标题翻译状态（可多选）</label>
                      <input
                        v-for="status in translateStatuses"
                        :key="`translate-status-input-${status}`"
                        type="hidden"
                        name="statuses"
                        :value="status"
                      />
                      <div class="next02-translate-status-toggle" role="group" aria-label="标题翻译状态（可多选）">
                        <button
                          v-for="option in translateStatusOptions"
                          :key="option.value"
                          type="button"
                          class="next02-translate-status-button"
                          :class="{ 'is-active': translateStatuses.includes(option.value) }"
                          :aria-pressed="translateStatuses.includes(option.value)"
                          @click="toggleTranslateStatus(option.value)"
                        >
                          {{ option.label }}
                        </button>
                      </div>
                    </div>

                    <div
                      class="form-group compact-field grid-span-2 next02-auto-fetch-setting"
                      id="autoFetchSiteFieldWrap"
                      :class="{ 'is-hidden': !showAutoFetchSiteField }"
                    >
                      <label>详情后自动抓取外站</label>
                      <div class="form-check next02-auto-fetch-switch">
                        <input
                          id="task_auto_fetch_site"
                          name="auto_fetch_site"
                          class="form-check-input"
                          type="checkbox"
                          checked
                        />
                        <label class="form-check-label" for="task_auto_fetch_site">JavBus / Sukebei</label>
                      </div>
                    </div>

                    <div
                      class="form-group compact-field grid-span-2"
                      id="fetchSehuatangKeywordFieldWrap"
                      :class="{ 'is-hidden': !showKeywordField }"
                    >
                      <label for="task_sehuatang_keyword">标题关键词</label>
                      <input
                        id="task_sehuatang_keyword"
                        v-model="form.keyword"
                        name="keyword"
                        class="form-control"
                        type="text"
                        placeholder="可选，留空表示抓取全部普通帖"
                      />
                    </div>

                    <div
                      class="form-group compact-field grid-span-2"
                      id="fetchSehuatangSHModeFieldWrap"
                      :class="{ 'is-hidden': !showSHModeField }"
                    >
                      <label for="task_sehuatang_sh_fetch_mode">抓取模式</label>
                      <select
                        id="task_sehuatang_sh_fetch_mode"
                        v-model="form.fetchMode"
                        name="fetch_mode"
                        class="form-select"
                      >
                        <option value="search">搜索结果翻页</option>
                        <option value="forum">版块列表翻页</option>
                      </select>
                    </div>

                    <div
                      class="form-group compact-field grid-span-2"
                      id="fetchSehuatangSHForumsFieldWrap"
                      :class="{ 'is-hidden': !showSHForumsField }"
                    >
                      <label>目标版块</label>
                      <template v-if="isSHSingleForumPage">
                        <input
                          v-for="forumID in targetForumIDs"
                          :key="`target-forum-${forumID}`"
                          type="hidden"
                          name="target_forum_ids"
                          :value="forumID"
                        />
                        <div class="sh-forum-single-toggle" role="radiogroup" aria-label="目标版块">
                          <button
                            type="button"
                            class="sh-forum-toggle-button"
                            :class="{ 'is-active': allTargetForumsSelected }"
                            :aria-pressed="allTargetForumsSelected"
                            @click="selectAllTargetForums"
                          >
                            全部
                          </button>
                          <button
                            v-for="forum in shForums"
                            :key="forum.forumID"
                            type="button"
                            class="sh-forum-toggle-button"
                            :class="{
                              'is-active': targetForumIDs.length === 1 && targetForumIDs[0] === forum.forumID
                            }"
                            :aria-pressed="targetForumIDs.length === 1 && targetForumIDs[0] === forum.forumID"
                            @click="selectTargetForum(forum.forumID)"
                          >
                            {{ forum.forumName }}
                          </button>
                        </div>
                      </template>
                      <div v-else class="status-check-grid">
                        <div v-for="forum in shForums" :key="forum.forumID" class="form-check">
                          <input
                            :id="`task_sehuatang_sh_forum_${forum.forumID}`"
                            v-model="targetForumIDs"
                            name="target_forum_ids"
                            class="form-check-input"
                            type="checkbox"
                            :value="forum.forumID"
                          />
                          <label class="form-check-label" :for="`task_sehuatang_sh_forum_${forum.forumID}`">{{
                            forum.forumName
                          }}</label>
                        </div>
                      </div>
                    </div>

                    <div
                      class="form-group compact-field grid-span-2"
                      id="fetchSehuatangSHExistingPolicyFieldWrap"
                      :class="{ 'is-hidden': !showSHExistingPolicyField }"
                    >
                      <label>已有帖子处理</label>
                      <input type="hidden" name="existing_policy" :value="form.existingPolicy" />
                      <div class="sh-forum-single-toggle" role="radiogroup" aria-label="已有帖子处理">
                        <button
                          type="button"
                          class="sh-forum-toggle-button"
                          :class="{ 'is-active': form.existingPolicy === 'upsert' }"
                          :aria-pressed="form.existingPolicy === 'upsert'"
                          @click="form.existingPolicy = 'upsert'"
                        >
                          Upsert（更新已有）
                        </button>
                        <button
                          type="button"
                          class="sh-forum-toggle-button"
                          :class="{ 'is-active': form.existingPolicy === 'skip' }"
                          :aria-pressed="form.existingPolicy === 'skip'"
                          @click="form.existingPolicy = 'skip'"
                        >
                          跳过已有
                        </button>
                      </div>
                      <div class="text-muted small mt-1 next02-sh-existing-policy-note">
                        {{
                          form.existingPolicy === "skip"
                            ? "已有帖子不更新、不抓取详情，只处理本次发现的新帖子。"
                            : "更新已有帖子基础信息，并重新抓取当前页面全部帖子详情。"
                        }}
                      </div>
                    </div>

                    <div
                      class="form-group compact-field grid-span-2"
                      id="fetchSehuatangSHDetailStatusFieldWrap"
                      :class="{ 'is-hidden': !showSHDetailFields }"
                    >
                      <label class="d-block mb-2">详情状态</label>
                      <div class="status-check-grid sh-detail-status-toggle">
                        <div class="form-check">
                          <input
                            id="task_sehuatang_sh_detail_status_pending"
                            v-model="shDetailStatusValues"
                            name="detail_statuses"
                            class="form-check-input"
                            type="checkbox"
                            value="pending"
                          />
                          <label class="form-check-label" for="task_sehuatang_sh_detail_status_pending">待抓详情</label>
                        </div>
                        <div class="form-check">
                          <input
                            id="task_sehuatang_sh_detail_status_failed"
                            v-model="shDetailStatusValues"
                            name="detail_statuses"
                            class="form-check-input"
                            type="checkbox"
                            value="failed"
                          />
                          <label class="form-check-label" for="task_sehuatang_sh_detail_status_failed">详情失败</label>
                        </div>
                        <div class="form-check">
                          <input
                            id="task_sehuatang_sh_detail_status_fetched"
                            v-model="shDetailStatusValues"
                            name="detail_statuses"
                            class="form-check-input"
                            type="checkbox"
                            value="fetched"
                          />
                          <label class="form-check-label" for="task_sehuatang_sh_detail_status_fetched">已抓详情</label>
                        </div>
                      </div>
                    </div>

                    <div
                      class="form-group compact-field"
                      id="fetchSehuatangSHPostTimeFromFieldWrap"
                      :class="{ 'is-hidden': !showSHDetailFields }"
                    >
                      <label for="task_sehuatang_sh_post_time_from">发帖时间开始</label>
                      <input
                        id="task_sehuatang_sh_post_time_from"
                        v-model="form.postTimeFrom"
                        name="post_time_from"
                        class="form-control smart-pick-date-input"
                        type="text"
                        placeholder="YYYY-MM-DD"
                        inputmode="numeric"
                        maxlength="10"
                        pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                      />
                    </div>

                    <div
                      class="form-group compact-field"
                      id="fetchSehuatangSHPostTimeToFieldWrap"
                      :class="{ 'is-hidden': !showSHDetailFields }"
                    >
                      <label for="task_sehuatang_sh_post_time_to">发帖时间结束</label>
                      <input
                        id="task_sehuatang_sh_post_time_to"
                        v-model="form.postTimeTo"
                        name="post_time_to"
                        class="form-control smart-pick-date-input"
                        type="text"
                        placeholder="YYYY-MM-DD"
                        inputmode="numeric"
                        maxlength="10"
                        pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                      />
                    </div>

                    <div
                      class="form-group compact-field"
                      id="fetchSehuatangSHDetailLimitFieldWrap"
                      :class="{ 'is-hidden': !showSHDetailFields }"
                    >
                      <label for="task_sehuatang_sh_detail_limit">预览 limit</label>
                      <input
                        id="task_sehuatang_sh_detail_limit"
                        v-model.number="form.detailLimit"
                        name="limit"
                        class="form-control"
                        type="number"
                        min="1"
                        placeholder="留空不限制"
                      />
                    </div>

                    <div
                      class="form-group compact-field"
                      id="fetchSehuatangStartPageFieldWrap"
                      :class="{ 'is-hidden': !showPagingFields }"
                    >
                      <label for="task_sehuatang_start_page">起始页</label>
                      <input
                        id="task_sehuatang_start_page"
                        v-model.number="form.startPage"
                        name="start_page"
                        class="form-control"
                        type="number"
                        placeholder="输入起始页"
                      />
                    </div>

                    <div
                      class="form-group compact-field"
                      id="fetchSehuatangEndPageFieldWrap"
                      :class="{ 'is-hidden': !showPagingFields }"
                    >
                      <label for="task_sehuatang_end_page">结束页</label>
                      <input
                        id="task_sehuatang_end_page"
                        v-model.number="form.endPage"
                        name="end_page"
                        class="form-control"
                        type="number"
                        placeholder="输入结束页"
                      />
                    </div>
                  </div>

                  <div
                    v-if="showFetchSitePanel"
                    id="fetchSiteFilterPanel"
                    class="mt-3 next02-fetch-site-filter-panel"
                    @input="invalidateFetchSitePreview"
                    @change="invalidateFetchSitePreview"
                  >
                    <div class="card border bg-light mb-0 next02-fetch-site-filter-card">
                      <div class="card-body">
                        <div
                          class="d-flex align-items-center justify-content-between mb-3 next02-fetch-site-filter-head"
                        >
                          <div class="next02-fetch-site-filter-title">
                            <span>FILTER</span>
                            <h5 class="mb-0">筛选影片</h5>
                          </div>
                          <div class="d-flex align-items-center next02-fetch-site-filter-actions">
                            <button
                              type="button"
                              id="toggleFetchSiteFilterBtn"
                              class="btn btn-outline-primary btn-sm mr-2"
                              @click="toggleFetchSiteFilter"
                            >
                              {{ fetchSiteFilterOpen ? "收起筛选" : "展开筛选" }}
                            </button>
                            <button
                              type="button"
                              id="clearFetchSiteFilterBtn"
                              class="btn btn-outline-secondary btn-sm"
                              @click="clearMovieFilters"
                            >
                              清空筛选
                            </button>
                          </div>
                        </div>
                        <div
                          id="fetchSiteFilterWrap"
                          class="next02-fetch-site-filter-fields"
                          :class="{ 'is-hidden': !fetchSiteFilterOpen }"
                        >
                          <MovieCardFilterForm :filter="filter" />
                        </div>
                        <div
                          class="row g-2 align-items-end mt-2 next02-fetch-site-filter-section next02-fetch-site-time-row"
                        >
                          <div class="col-12 next02-fetch-site-section-heading">
                            <div>
                              <span>FETCH WINDOW</span>
                              <label class="form-label mb-0">抓取时间</label>
                            </div>
                            <small>小于设定天数的影片将被跳过</small>
                          </div>
                          <div class="col-12 col-md-6 col-xl-3">
                            <label class="form-label mb-1">最近抓取</label>
                            <input
                              type="number"
                              class="form-control"
                              name="last_fetch_duration_days"
                              value="5"
                              placeholder="例如 3"
                            />
                          </div>
                          <div class="col-12 col-md-6 col-xl-3">
                            <label class="form-label mb-1">最近成功</label>
                            <input
                              type="number"
                              class="form-control"
                              name="last_success_duration_days"
                              value="5"
                              placeholder="例如 3"
                            />
                          </div>
                        </div>
                        <div
                          class="row g-2 align-items-end mt-2 next02-fetch-site-filter-section next02-fetch-site-order-row"
                        >
                          <div class="col-12 next02-fetch-site-section-heading">
                            <div>
                              <span>ORDER</span>
                              <label class="form-label mb-0">排序规则</label>
                            </div>
                          </div>
                          <div class="col-12 next02-fetch-site-order-options">
                            <div
                              class="btn-group btn-group-toggle d-flex flex-wrap task-type-toggle"
                              role="group"
                              aria-label="fetch site movie order by"
                            >
                              <button
                                v-for="option in fetchSiteOrderOptions"
                                :key="option.value"
                                type="button"
                                class="btn btn-outline-primary flex-fill"
                                :class="{ active: fetchSiteOrderBy === option.value }"
                                :data-fetch-site-orderby="option.value"
                                @click="setFetchSiteOrderBy(option.value)"
                              >
                                {{ option.label }}
                              </button>
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

                  <div
                    class="d-flex align-items-center flex-wrap action-row next02-action-dock"
                    :class="{
                      'next02-seeds-action-dock next02-pink-trigger-action-dock': isPinkTriggerRewritePage,
                      'next02-fetch-site-action-dock': isFetchSiteCommandPage
                    }"
                  >
                    <button
                      v-if="showFetchSitePanel"
                      type="button"
                      id="previewFetchSiteBtn"
                      class="btn btn-outline-info mr-2"
                      :disabled="fetchSitePreviewLoading"
                      @click="previewFetchSite"
                    >
                      预览筛选结果
                    </button>
                    <button
                      v-if="showSHDetailFields"
                      type="button"
                      id="previewSHDetailBtn"
                      class="btn btn-outline-info mr-2"
                      :disabled="shDetailPreviewLoading"
                      @click="previewSHDetail"
                    >
                      筛选帖子预览
                    </button>
                    <button type="submit" id="submitBtn" class="btn btn-primary" :disabled="submitDisabled">
                      {{ showSHDetailFields || showFetchSitePanel ? "抓取预览结果" : "开始执行" }}
                    </button>
                    <button
                      type="button"
                      id="pauseBtn"
                      class="btn btn-outline-warning ml-2"
                      :disabled="pauseDisabled"
                      @click="controlSelected('pause')"
                    >
                      暂停
                    </button>
                    <button
                      type="button"
                      id="resumeBtn"
                      class="btn btn-outline-success ml-2"
                      :disabled="resumeDisabled"
                      @click="controlSelected('resume')"
                    >
                      继续
                    </button>
                    <button
                      type="button"
                      id="stopBtn"
                      class="btn btn-outline-danger ml-2"
                      :disabled="stopDisabled"
                      @click="controlSelected('stop')"
                    >
                      停止
                    </button>
                  </div>
                </form>
              </TriggerForm>

              <JobRuntimePanel
                :pink="isPinkTriggerRewritePage"
                :command-visual="isCommandVisualPage"
                :eyebrow="commandRuntimeEyebrow"
                :runtime-overview="isRuntimeContextOverview"
                :current-job="currentJob"
                :current-job-done="currentJobDone"
                :badge-text="runtimeBadgeText"
                :jobs="jobs"
                :selected-job-id="selectedJobID"
                :runtime-context="runtimeContext"
                :stage-mode="isStageOverviewPage"
                :stage-groups="stageOverviewGroups"
                :show-metrics="showRuntimeOverviewMetrics"
                :metrics="overviewItems"
                :show-queue="showFilteredQueuePanel"
                :active-items="filteredQueueActiveItems"
                :done-items="filteredQueueDoneItems"
                :show-progress="showRuntimeProgress"
                :progress-percent="currentProgressPercent"
                :status-class="runtimeStatusClass"
                :status-text="runtimeStatusText"
                :event-count="renderedEventEntries.length"
                @select-job="openJob($event)"
              />
            </div>

            <JobEventStream
              v-model:show-raw="showRawEvents"
              :pink="isPinkTriggerRewritePage"
              :command-visual="isCommandVisualPage"
              :eyebrow="commandEventEyebrow"
              :title="eventTitle"
              :entries="renderedEventEntries"
              :empty-text="emptyStateText"
            />
          </div>

          <FetchSitePreviewModal
            v-if="showFetchSitePanel"
            :open="fetchSitePreviewOpen"
            :loading="fetchSitePreviewLoading"
            :pink="isPinkTriggerRewritePage"
            :error="fetchSitePreviewError"
            :summary="fetchSitePreviewSummary"
            :groups="fetchSitePreviewMetricGroups"
            @hide="hideFetchSitePreview"
          />

          <SHDetailPreview
            v-if="showSHDetailFields"
            :open="shDetailPreviewOpen"
            :loading="shDetailPreviewLoading"
            :error="shDetailPreviewError"
            :summary="shDetailPreviewSummary"
            :pool-href="shDetailPreviewPoolHref"
            :groups="shDetailPreviewMetricGroups"
            @collapse="collapseSHDetailPreview"
          />
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/features/crawler/api";
import { useCrawlerJobs } from "@/features/crawler/composables/useCrawlerJobs";
import { useJobEventStream } from "@/features/crawler/composables/useJobEventStream";
import { useTriggerForm } from "@/features/crawler/composables/useTriggerForm";
import { useTriggerPreview } from "@/features/crawler/composables/useTriggerPreview";
import FetchSitePreviewModal from "@/features/crawler/components/trigger/FetchSitePreviewModal.vue";
import JobEventStream from "@/features/crawler/components/trigger/JobEventStream.vue";
import JobRuntimePanel from "@/features/crawler/components/trigger/JobRuntimePanel.vue";
import SHDetailPreview from "@/features/crawler/components/trigger/SHDetailPreview.vue";
import TriggerForm from "@/features/crawler/components/trigger/TriggerForm.vue";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import MovieCardFilterForm from "@/features/movies/components/MovieCardFilterForm.vue";
import PageShell from "@/shared/components/PageShell.vue";
import type { CrawlerRecord } from "@/features/crawler/types";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";
import { loadSHForumOptions, shForumName, shForumNames, type SHForumOption } from "@/shared/utils/shForums";
import { estimateRemainingSeconds, formatTaskEta } from "@/shared/utils/taskEta";

const route = useRoute();
const payload = ref<CrawlerRecord>({});
const shForums = ref<SHForumOption[]>([]);
const jobs = ref<CrawlerRecord[]>([]);
const loading = ref(false);
const error = ref("");
const message = ref("");
const selectedTaskType = ref("");
const selectedJobID = ref("");
const translateStatusOptions = [
  { value: "1", label: "无中文字幕(1)" },
  { value: "2", label: "有中文字幕(2)" },
  { value: "3", label: "翻译错误(3)" },
  { value: "4", label: "敏感词(4)" }
] as const;
const translateStatuses = ref<string[]>(["1"]);
const fetchSiteFilterOpen = ref(false);
const fetchSiteOrderBy = ref("rd");
const {
  open: fetchSitePreviewOpen,
  loading: fetchSitePreviewLoading,
  reset: resetFetchSitePreview
} = useTriggerPreview();
const fetchSitePreviewTargets = ref<CrawlerRecord[]>([]);
const fetchSitePreviewSignature = ref("");
const fetchSitePreviewAggregate = ref<CrawlerRecord>({});
const fetchSitePreviewError = ref("");
const fetchSitePreviewTaskType = ref("");
const targetForumIDs = ref<string[]>(["103"]);
const allTargetForumsSelected = computed(() => {
  const availableForumIDs = shForums.value.map((forum) => forum.forumID);
  const selected = new Set(targetForumIDs.value);
  return (
    availableForumIDs.length > 0 &&
    selected.size === availableForumIDs.length &&
    availableForumIDs.every((forumID) => selected.has(forumID))
  );
});
const shDetailStatusValues = ref<string[]>(["pending"]);
const {
  open: shDetailPreviewOpen,
  loading: shDetailPreviewLoading,
  error: shDetailPreviewError,
  reset: resetSHDetailPreview
} = useTriggerPreview();
const shDetailPreviewThreadIDs = ref<string[]>([]);
const shDetailPreviewAggregate = ref<CrawlerRecord>({});
const shDetailPreviewTotal = ref(0);
const shDetailPreviewSignature = ref("");
const shDetailPreviewPoolHref = ref("");
const jobEventHistorySize = 300;
const rawJobEvents = ref<CrawlerRecord[]>([]);
const rawJobEventKeys = new Set<string>();
const showRawEvents = ref(false);
const nowTick = ref(Date.now());
let elapsedTimer: number | undefined;
let eventReconnectTimer: number | undefined;
let lastEventID = 0;
const form = useTriggerForm();
const jobEventStream = useJobEventStream();
const { start: startJobsPolling, stop: stopJobsPolling } = useCrawlerJobs(() => loadJobs(true));

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
  { value: "lsst", label: "最近入选 SC" },
  { value: "sc", label: "sc次数" },
  { value: "co", label: "co次数" }
];

const movieFilterFieldNames = [
  "cn",
  "pid",
  "gn",
  "dn",
  "pn",
  "mn",
  "ln",
  "rs",
  "re",
  "mbs",
  "mbe",
  "cay",
  "cao",
  "srds",
  "srde",
  "drkmin",
  "nd",
  "wd",
  "mowned",
  "vwmin",
  "vwmax",
  "smin",
  "smax",
  "lsstmin",
  "lsstmax",
  "scmin",
  "scmax",
  "comin",
  "comax",
  "md1",
  "md2",
  "md3",
  "md4",
  "od",
  "order"
];

const fetchSiteDurationFieldNames = ["last_fetch_duration_days", "last_success_duration_days"];

watch(() => route.fullPath, load, { immediate: true });

onBeforeUnmount(() => {
  stopJobsPolling();
  if (elapsedTimer) {
    window.clearInterval(elapsedTimer);
    elapsedTimer = undefined;
  }
  closeEventStream();
});

async function load() {
  closeEventStream();
  clearJobEvents();
  loading.value = true;
  error.value = "";
  message.value = "";
  try {
    const needsSHForums = ["/shfetchall", "/shfetch", "/shfill"].includes(route.path);
    const [pagePayload, forumOptions] = await Promise.all([
      getJSON(apiPathForRoute(route)),
      needsSHForums ? loadSHForumOptions() : Promise.resolve([])
    ]);
    payload.value = asRecord(pagePayload);
    shForums.value = forumOptions;
    selectedTaskType.value = text(payload.value, "DefaultTaskType");
    const configuredJobID = normalizeJobID(text(payload.value, "JobID") || String(route.query.job_id || ""));
    const storedJobID = storedSelectedJobID();
    selectedJobID.value = configuredJobID || storedJobID;
    translateStatuses.value = ["1"];
    fetchSiteFilterOpen.value = false;
    fetchSiteOrderBy.value = text(asRecord(pick(payload.value, "MovieCardFilter")), "OrderBy") || "rd";
    hideFetchSitePreview();
    form.keyword = "";
    form.startPage = Number(text(payload.value, "FetchSehuatangStartPage") || 0);
    form.endPage = Number(text(payload.value, "FetchSehuatangEndPage") || 0);
    form.fetchMode = text(payload.value, "FetchSehuatangSHMode") === "forum" ? "forum" : "search";
    form.existingPolicy = "skip";
    targetForumIDs.value = initialTargetForumIDs();
    shDetailStatusValues.value = initialSHDetailStatusValues();
    form.number = 1000000;
    form.postTimeFrom = "";
    form.postTimeTo = "";
    form.detailLimit = undefined;
    hideSHDetailPreview();
    await loadJobs(true);
    ensureJobsPolling();
    ensureElapsedTimer();
    if (selectedJobID.value) {
      await openJob(selectedJobID.value, false);
    } else {
      connectEventStream();
    }
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
const taskButtons = computed(() => asArray<CrawlerRecord>(pick(payload.value, "TaskButtons")));
const hideTaskForm = computed(() => Boolean(pick(payload.value, "HideTaskForm")));
const headerLinks = computed(() =>
  asArray<CrawlerRecord>(pick(payload.value, "HeaderLinks")).map((item) => ({
    label: text(item, "Label"),
    href: text(item, "Href")
  }))
);
const allHeaderLinks = computed(() => {
  const links = [...headerLinks.value];
  const ensure = (href: string, label: string) => {
    if (!links.some((item) => item.href === href)) {
      links.push({ href, label });
    }
  };
  ensure("/crawltasklist", "任务列表");
  ensure("/scadd", "SC 预览与提交");
  return links;
});
const filter = computed(() => asRecord(pick(payload.value, "MovieCardFilter")));
const showMovieFilter = computed(() => Boolean(payload.value.ShowMovieFilters));
const emptyStateText = computed(() => text(payload.value, "EmptyStateText") || "等待任务触发");
const currentTaskType = computed(() => selectedTaskType.value || text(payload.value, "DefaultTaskType"));
const isFetchSiteTask = computed(() => fetchSiteTaskTypes.has(currentTaskType.value));
const isSehuatangSHTask = computed(() => currentTaskType.value === "spider_fetch_sehuatang_sh_threads");
const isSehuatangSHListTask = computed(() => currentTaskType.value === "spider_fetch_sehuatang_sh_thread_list");
const isSehuatangSHDetailTask = computed(() => currentTaskType.value === "spider_fetch_sehuatang_sh_thread_detail");
const isSehuatangSHPage = computed(
  () => isSehuatangSHTask.value || isSehuatangSHListTask.value || isSehuatangSHDetailTask.value
);
const isSHSingleForumPage = computed(() => ["/shfetchall", "/shfetch", "/shfill"].includes(route.path));
const isFetchSiteCommandPage = computed(() => route.path === "/sitefetch" && isFetchSiteTask.value);
const isDailyBestCommandPage = computed(() => route.path === "/dailybestfetch");
const isSeedsCommandPage = computed(() => route.path === "/seedfetch");
const isPostProcessCommandPage = computed(() => route.path === "/postprocess");
const isBackfillCommandPage = computed(() => route.path === "/databackfill");
const isFilteredFetchSitePage = computed(() => ["/javbusfetch", "/sukebeifetch"].includes(route.path));
const pinkTriggerPaths = new Set([
  "/dailybestfetch",
  "/seedfetch",
  "/postprocess",
  "/databackfill",
  "/sitefetch",
  "/shfetchall",
  "/shfetch",
  "/shfill",
  "/javbusfetch",
  "/sukebeifetch"
]);
const isPinkTriggerRewritePage = computed(() => pinkTriggerPaths.has(route.path));
const isFormHeavyTriggerPage = computed(() => isFetchSiteCommandPage.value || isSehuatangSHPage.value);
const isCommandVisualPage = computed(() => isPinkTriggerRewritePage.value);
const showFetchSitePanel = computed(() => isFetchSiteTask.value && showMovieFilter.value);
const showNameField = computed(() => currentTaskType.value === "spider_seed_by_name");
const showNumberField = computed(() => currentTaskType.value === "spider_refresh_oldest_detail");
const showActorNameField = computed(() => currentTaskType.value === "spider_rebuild_actor_rank");
const showSHModeField = computed(() => isSehuatangSHTask.value && route.path !== "/shfetchall");
const showSHExistingPolicyField = computed(() => isSehuatangSHTask.value && route.path === "/shfetchall");
const showSHForumsField = computed(
  () => isSehuatangSHTask.value || isSehuatangSHListTask.value || isSehuatangSHDetailTask.value
);
const showSHDetailFields = computed(() => isSehuatangSHDetailTask.value);
const showPagingFields = computed(() => isSehuatangSHTask.value || isSehuatangSHListTask.value);
const showKeywordField = computed(
  () => isSehuatangSHTask.value && route.path !== "/shfetchall" && form.fetchMode === "search"
);
const showAutoFetchSiteField = computed(() => autoFetchSiteTaskTypes.has(currentTaskType.value));
const showTranslateStatusesField = computed(
  () => currentTaskType.value === "spider_translate_title" || currentTaskType.value === "spider_post_process"
);
const overviewMode = computed(() => text(payload.value, "OverviewExtraMode"));
const isSHListOverview = computed(() => overviewMode.value === "sehuatang_list_progress");
const isSHDetailOverview = computed(() => overviewMode.value === "sehuatang_detail_progress");
const isSHRuntimeOverview = computed(() => isSHListOverview.value || isSHDetailOverview.value);
const isFetchSiteRuntimeOverview = computed(() => overviewMode.value === "fetch_site_target_progress");
const isRuntimeContextOverview = computed(() => isSHRuntimeOverview.value || isFetchSiteRuntimeOverview.value);
const isStageOverviewPage = computed(() =>
  ["dailybest_stages", "seeds_stages", "post_process_stages"].includes(overviewMode.value)
);
const allowedTaskTypes = computed(() =>
  asArray<string>(pick(payload.value, "AllowedTaskTypes"))
    .map((item) => String(item || "").trim())
    .filter(Boolean)
);
const selectedJobStorageKey = computed(() => text(payload.value, "StorageKey") || "crawler_jobs_selected_job");
const currentJob = computed(() => {
  const selected = selectedJobID.value.trim();
  if (selected) {
    const byID = jobs.value.find((item) => jobID(item) === selected);
    return byID;
  }
  const type = currentTaskType.value;
  const list = jobs.value.filter((item) => text(item, "task_type", "TaskType") === type);
  list.sort((a, b) => jobSortTs(b) - jobSortTs(a));
  return list[0];
});
const currentJobDone = computed(() => Boolean(pick(currentJob.value || {}, "done", "Done")));
const currentJobPaused = computed(() => Boolean(pick(currentJob.value || {}, "paused", "Paused")));
const showRuntimeOverviewMetrics = computed(() => !isRuntimeContextOverview.value || Boolean(currentJob.value));
const showRuntimeProgress = computed(() => !isRuntimeContextOverview.value || Boolean(currentJob.value));
const runtimeBadgeText = computed(() => {
  if (!isRuntimeContextOverview.value) {
    return "Live";
  }
  if (!currentJob.value) {
    return "待执行";
  }
  if (currentJobDone.value && text(currentJob.value, "stage", "Stage") === "failed") {
    return "失败";
  }
  if (currentJobDone.value) {
    return "已完成";
  }
  if (currentJobPaused.value) {
    return "已暂停";
  }
  return "抓取中";
});
const currentTaskStatusLabel = computed(() => {
  if (!currentJob.value) {
    return "Ready";
  }
  if (currentJobDone.value) {
    return "Done";
  }
  if (currentJobPaused.value) {
    return "Paused";
  }
  return "Running";
});
const pauseDisabled = computed(() => !currentJob.value || currentJobPaused.value || currentJobDone.value);
const resumeDisabled = computed(() => !currentJob.value || !currentJobPaused.value || currentJobDone.value);
const stopDisabled = computed(() => {
  if (currentJobDone.value) {
    return true;
  }
  return !currentJob.value && !selectedJobID.value.trim();
});
const currentProgressPercent = computed(() => {
  const job = currentJob.value || {};
  if (isSHRuntimeOverview.value) {
    const progress = sehuatangCurrentProgressCounts(job);
    return progress.total > 0 ? Math.max(0, Math.min(100, Math.round((progress.completed * 100) / progress.total))) : 0;
  }
  if (isFetchSiteRuntimeOverview.value) {
    const completed = jobStatNumber(job, "completed_items");
    const selected = jobStatNumber(job, "selected_items");
    return selected > 0 ? Math.max(0, Math.min(100, Math.round((completed * 100) / selected))) : 0;
  }
  if (["javbus_filtered_queue", "sukebei_filtered_queue"].includes(overviewMode.value)) {
    const pending = asArray<CrawlerRecord>(pick(job, "pending_items", "PendingItems")).length;
    const running = asArray<CrawlerRecord>(pick(job, "running_items", "RunningItems")).length;
    const done = asArray<CrawlerRecord>(pick(job, "done_items", "DoneItems")).length;
    const total = pending + running + done;
    return total > 0 ? Math.max(0, Math.min(100, Math.round((done * 100) / total))) : 0;
  }
  return progressPercent(job);
});
const runtimeStatusText = computed(() => {
  const job = currentJob.value;
  if (!job) {
    return selectedJobID.value ? "正在获取任务状态" : emptyStateText.value;
  }
  return text(job, "message", "Message") || stageLabel(text(job, "stage", "Stage"));
});
const runtimeStatusClass = computed(() => {
  if (!currentJob.value) {
    return "text-muted";
  }
  if (currentJobDone.value && text(currentJob.value, "stage", "Stage") === "failed") {
    return "err";
  }
  return currentJobDone.value ? "ok" : "text-muted";
});
const renderedEventEntries = computed(() =>
  rawJobEvents.value.flatMap((event, index) => {
    const entry = buildEventEntry(event, showRawEvents.value);
    return entry ? [{ ...entry, key: `${eventKey(event)}:${index}` }] : [];
  })
);
const showFilteredQueuePanel = computed(() =>
  ["javbus_filtered_queue", "sukebei_filtered_queue"].includes(overviewMode.value)
);
const filteredQueuePendingItems = computed(() =>
  asArray<CrawlerRecord>(pick(currentJob.value || {}, "pending_items", "PendingItems"))
);
const filteredQueueRunningItems = computed(() =>
  asArray<CrawlerRecord>(pick(currentJob.value || {}, "running_items", "RunningItems"))
);
const filteredQueueActiveItems = computed(() => [
  ...filteredQueueRunningItems.value,
  ...filteredQueuePendingItems.value
]);
const filteredQueueDoneItems = computed(() =>
  [...asArray<CrawlerRecord>(pick(currentJob.value || {}, "done_items", "DoneItems"))].reverse()
);
const fetchSitePreviewSummary = computed(() => {
  if (!fetchSitePreviewOpen.value) {
    return "当前还没有预览数据";
  }
  if (fetchSitePreviewLoading.value) {
    return "正在生成固定目标清单";
  }
  const matched = Number(pick(fetchSitePreviewAggregate.value, "matched_movies", "MatchedMovies") || 0);
  const selected = Number(pick(fetchSitePreviewAggregate.value, "selected_items", "SelectedItems") || 0);
  return `${taskLabelByType(fetchSitePreviewTaskType.value)} 命中 ${matched} 部影片，生成 ${selected} 个抓取项`;
});
const fetchSitePreviewMetricGroups = computed(() => {
  const aggregate = fetchSitePreviewAggregate.value;
  return [
    {
      label: "数据范围",
      items: [
        {
          label: "命中影片",
          value: String(Number(pick(aggregate, "matched_movies", "MatchedMovies") || 0)),
          className: ""
        },
        {
          label: "抓取项",
          value: String(Number(pick(aggregate, "selected_items", "SelectedItems") || 0)),
          className: "is-pending"
        }
      ]
    },
    {
      label: "来源分布",
      items: [
        {
          label: "JavBus",
          value: String(Number(pick(aggregate, "javbus_count", "JavbusCount") || 0)),
          className: "is-fetched"
        },
        {
          label: "Sukebei",
          value: String(Number(pick(aggregate, "sukebei_count", "SukebeiCount") || 0)),
          className: "is-forum"
        },
        {
          label: "双来源影片",
          value: String(Number(pick(aggregate, "both_count", "BothCount") || 0)),
          className: ""
        }
      ]
    },
    {
      label: "时间过滤",
      items: [
        {
          label: "最近抓取跳过",
          value: String(Number(pick(aggregate, "skipped_recent_fetch", "SkippedRecentFetch") || 0)),
          className: ""
        },
        {
          label: "最近成功跳过",
          value: String(Number(pick(aggregate, "skipped_recent_success", "SkippedRecentSuccess") || 0)),
          className: ""
        }
      ]
    }
  ];
});
const shDetailPreviewSummary = computed(() => {
  if (!shDetailPreviewOpen.value) {
    return "当前还没有预览数据";
  }
  if (shDetailPreviewLoading.value) {
    return "正在加载预览数据";
  }
  return `筛选命中 ${shDetailPreviewTotal.value} 条，本次抓取 ${shDetailPreviewThreadIDs.value.length} 条`;
});
const shDetailPreviewMetricGroups = computed(() => {
  const aggregate = shDetailPreviewAggregate.value;
  const forumCounts = asRecord(pick(aggregate, "forum_counts", "ForumCounts"));
  return [
    {
      label: "数据范围",
      items: [
        { label: "命中总数", value: String(shDetailPreviewTotal.value), className: "" },
        { label: "本次抓取", value: String(shDetailPreviewThreadIDs.value.length), className: "" }
      ]
    },
    {
      label: "详情状态",
      items: [
        {
          label: "待抓详情",
          value: String(Number(pick(aggregate, "pending_detail_count", "PendingDetailCount") || 0)),
          className: "is-pending"
        },
        {
          label: "已抓详情",
          value: String(Number(pick(aggregate, "fetched_detail_count", "FetchedDetailCount") || 0)),
          className: "is-fetched"
        },
        {
          label: "详情失败",
          value: String(Number(pick(aggregate, "failed_detail_count", "FailedDetailCount") || 0)),
          className: "is-failed"
        }
      ]
    },
    {
      label: "来源板块",
      items: shForums.value.map((forum) => ({
        label: forum.forumName,
        value: String(Number(pick(forumCounts, forum.forumID) || 0)),
        className: "is-forum"
      }))
    }
  ];
});
const shDetailFilterSignature = computed(() =>
  JSON.stringify({
    targetForumIDs: [...targetForumIDs.value].sort(),
    detailStatuses: [...shDetailStatusValues.value].sort(),
    postTimeFrom: form.postTimeFrom.trim(),
    postTimeTo: form.postTimeTo.trim(),
    limit: form.detailLimit ?? ""
  })
);
const shDetailSubmitDisabled = computed(() => {
  if (!showSHDetailFields.value) {
    return false;
  }
  return (
    shDetailPreviewLoading.value ||
    !shDetailPreviewThreadIDs.value.length ||
    shDetailPreviewSignature.value !== shDetailFilterSignature.value
  );
});
const fetchSiteSubmitDisabled = computed(() => {
  if (!showFetchSitePanel.value) {
    return false;
  }
  return (
    fetchSitePreviewLoading.value ||
    !fetchSitePreviewTargets.value.length ||
    !fetchSitePreviewSignature.value ||
    fetchSitePreviewTaskType.value !== currentTaskType.value
  );
});
const submitDisabled = computed(() =>
  showFetchSitePanel.value ? fetchSiteSubmitDisabled.value : shDetailSubmitDisabled.value
);

watch(shDetailFilterSignature, (signature) => {
  if (shDetailPreviewSignature.value && shDetailPreviewSignature.value !== signature) {
    hideSHDetailPreview();
  }
});
watch(selectedTaskType, (taskType, previousTaskType) => {
  if (previousTaskType && taskType !== previousTaskType) {
    hideFetchSitePreview();
  }
});
const shHeroMetrics = computed(() => [
  { label: "目标版块", value: shForumNames(shForums.value, targetForumIDs.value) },
  { label: "起止页", value: showPagingFields.value ? `${form.startPage || "-"}-${form.endPage || "-"}` : "-" },
  { label: "已有帖子", value: showSHExistingPolicyField.value ? shExistingPolicyLabel(form.existingPolicy) : "-" },
  { label: "当前任务", value: taskLabelByType(currentTaskType.value) }
]);
const shListRuntimeContext = computed(() => {
  const job = currentJob.value || {};
  const progress = sehuatangCurrentProgressCounts(job);
  const targetForumText = jobStatText(job, "target_forum_ids") || targetForumIDs.value.join(",");
  const currentForumID = jobStatText(job, "current_forum_id") || targetForumText.split(",")[0] || "";
  const currentPage = jobStatNumber(job, "current_page");
  const startPage = jobStatNumber(job, "start_page") || Number(form.startPage || 0);
  const endPage = jobStatNumber(job, "end_page") || Number(form.endPage || 0);
  return [
    { label: "当前阶段", value: currentPhaseLabel(job) },
    { label: "已有帖子", value: shExistingPolicyLabel(jobStatText(job, "existing_policy") || form.existingPolicy) },
    { label: "目标版块", value: shForumNames(shForums.value, targetForumText) },
    { label: "页码范围", value: startPage > 0 && endPage > 0 ? `${startPage}-${endPage}` : "-" },
    {
      label: "当前抓取",
      value: currentPage > 0 ? `${shForumName(shForums.value, currentForumID)} · 第 ${currentPage} 页` : "准备中"
    },
    { label: "运行时长", value: formatJobElapsed(job) },
    { label: "预计剩余时间", value: formatJobEstimatedRemaining(job, progress.completed, progress.total) }
  ];
});
const shDetailRuntimeContext = computed(() => {
  const job = currentJob.value || {};
  const targetForumText = jobStatText(job, "target_forum_ids") || targetForumIDs.value.join(",");
  const page = sehuatangPageCounts(job);
  const progress = sehuatangCurrentProgressCounts(job);
  const currentPageText = jobStatText(job, "current_page_text");
  const currentMatch = currentPageText.match(/^detail=(\d+)\/(\d+)(?:\s+tid=(\S+))?/);
  const currentDetail = currentMatch
    ? `${currentMatch[1]}/${currentMatch[2]}${currentMatch[3] ? ` · TID ${currentMatch[3]}` : ""}`
    : jobStatNumber(job, "current_page") > 0
      ? `第 ${jobStatNumber(job, "current_page")} 条`
      : "准备中";
  return [
    { label: "当前阶段", value: currentPhaseLabel(job) },
    { label: "目标版块", value: shForumNames(shForums.value, targetForumText) },
    { label: "选中帖子", value: String(page.total) },
    { label: "当前抓取", value: currentDetail },
    { label: "运行时长", value: formatJobElapsed(job) },
    { label: "预计剩余时间", value: formatJobEstimatedRemaining(job, progress.completed, progress.total) }
  ];
});
const shRuntimeContext = computed(() =>
  isSHDetailOverview.value ? shDetailRuntimeContext.value : shListRuntimeContext.value
);
const fetchSiteRuntimeContext = computed(() => {
  const job = currentJob.value || {};
  const completed = jobStatNumber(job, "completed_items");
  const selected = jobStatNumber(job, "selected_items");
  const currentSource = jobStatText(job, "current_source");
  const currentMovieName = jobStatText(job, "current_movie_name");
  const currentMovieJavID = jobStatText(job, "current_movie_jav_id");
  const currentTarget = [currentSource, currentMovieName || currentMovieJavID].filter(Boolean).join(" · ");
  return [
    { label: "当前阶段", value: currentPhaseLabel(job) },
    { label: "当前来源", value: currentSource || "准备中" },
    { label: "当前抓取", value: currentTarget || "准备中" },
    { label: "目标进度", value: `${completed}/${selected}` },
    { label: "运行时长", value: formatJobElapsed(job) },
    { label: "预计剩余时间", value: formatJobEstimatedRemaining(job, completed, selected) }
  ];
});
const runtimeContext = computed(() =>
  isFetchSiteRuntimeOverview.value ? fetchSiteRuntimeContext.value : shRuntimeContext.value
);
const fetchSiteHeroMetrics = computed(() => [
  { label: "当前任务", value: taskLabelByType(currentTaskType.value) },
  {
    label: "预览目标",
    value: String(Number(pick(fetchSitePreviewAggregate.value, "selected_items", "SelectedItems") || 0))
  },
  { label: "排序", value: fetchSiteOrderBy.value || "rd" },
  {
    label: "命中影片",
    value: String(Number(pick(fetchSitePreviewAggregate.value, "matched_movies", "MatchedMovies") || 0))
  }
]);
const dailybestHeroMetrics = computed(() => [
  { label: "当前任务", value: taskLabelByType(currentTaskType.value) },
  { label: "任务按钮", value: String(taskButtons.value.length || 0) },
  { label: "阶段面板", value: overviewMode.value === "dailybest_stages" ? "启用" : "默认" },
  { label: "任务状态", value: currentTaskStatusLabel.value }
]);
const seedsHeroMetrics = computed(() => [
  { label: "当前任务", value: taskLabelByType(currentTaskType.value) },
  { label: "Seed 名称", value: form.name.trim() || "-" },
  { label: "刷新数量", value: String(form.number || 0) },
  { label: "任务状态", value: currentTaskStatusLabel.value }
]);
const postProcessHeroMetrics = computed(() => [
  { label: "当前任务", value: taskLabelByType(currentTaskType.value) },
  { label: "状态数", value: String(translateStatuses.value.length || 0) },
  { label: "阶段面板", value: overviewMode.value === "post_process_stages" ? "启用" : "默认" },
  { label: "任务状态", value: currentTaskStatusLabel.value }
]);
const backfillHeroMetrics = computed(() => [
  { label: "当前任务", value: taskLabelByType(currentTaskType.value) },
  { label: "演员名称", value: form.actorName.trim() || "-" },
  { label: "任务按钮", value: String(taskButtons.value.length || 0) },
  { label: "任务状态", value: currentTaskStatusLabel.value }
]);
const commandHeroKicker = computed(() => {
  if (isDailyBestCommandPage.value) {
    return "DailyBest Command";
  }
  if (isSeedsCommandPage.value) {
    return "Seeds Command";
  }
  if (isPostProcessCommandPage.value) {
    return "PostProcess Command";
  }
  if (isBackfillCommandPage.value) {
    return "Backfill Command";
  }
  if (isFetchSiteCommandPage.value) {
    return "FetchSite Command";
  }
  if (isFilteredFetchSitePage.value) {
    return route.path.includes("javbus") ? "JavBus Filter Queue" : "Sukebei Filter Queue";
  }
  return "Sehuatang SH Command";
});
const commandTaskEyebrow = computed(() =>
  isDailyBestCommandPage.value
    ? "Rank Console"
    : isSeedsCommandPage.value
      ? "Seeds Console"
      : isPostProcessCommandPage.value
        ? "Post Console"
        : isBackfillCommandPage.value
          ? "Backfill Console"
          : isFetchSiteCommandPage.value
            ? "Fetch Console"
            : isFilteredFetchSitePage.value
              ? "Filter Queue"
              : "Task Console"
);
const commandRuntimeEyebrow = computed(() =>
  isDailyBestCommandPage.value
    ? "Rank Runtime"
    : isSeedsCommandPage.value
      ? "Seeds Runtime"
      : isPostProcessCommandPage.value
        ? "Post Runtime"
        : isBackfillCommandPage.value
          ? "Backfill Runtime"
          : isFetchSiteCommandPage.value
            ? "Site Runtime"
            : isFilteredFetchSitePage.value
              ? "Queue Runtime"
              : "Runtime"
);
const commandEventEyebrow = computed(() =>
  isDailyBestCommandPage.value
    ? "Rank Event Stream"
    : isSeedsCommandPage.value
      ? "Seeds Event Stream"
      : isPostProcessCommandPage.value
        ? "Post Event Stream"
        : isBackfillCommandPage.value
          ? "Backfill Event Stream"
          : isFetchSiteCommandPage.value
            ? "Fetch Event Stream"
            : isFilteredFetchSitePage.value
              ? "Queue Event Stream"
              : "Event Stream"
);
const commandHeroMetrics = computed(() => {
  if (isDailyBestCommandPage.value) {
    return dailybestHeroMetrics.value;
  }
  if (isSeedsCommandPage.value) {
    return seedsHeroMetrics.value;
  }
  if (isPostProcessCommandPage.value) {
    return postProcessHeroMetrics.value;
  }
  if (isBackfillCommandPage.value) {
    return backfillHeroMetrics.value;
  }
  if (isFetchSiteCommandPage.value) {
    return fetchSiteHeroMetrics.value;
  }
  return shHeroMetrics.value;
});
const overviewItems = computed(() => {
  const mode = overviewMode.value;
  if (mode === "fetch_site_target_progress") {
    const job = currentJob.value || {};
    return [
      {
        label: "目标进度",
        value: `${jobStatNumber(job, "completed_items")} / ${jobStatNumber(job, "selected_items")}`
      },
      {
        label: "总体成功 / 失败",
        value: `${jobStatNumber(job, "success_items")} / ${jobStatNumber(job, "failed_items")}`
      },
      { label: "JavBus 已处理", value: String(jobStatNumber(job, "javbus_completed")) },
      {
        label: "JavBus 成功 / 失败",
        value: `${jobStatNumber(job, "javbus_success")} / ${jobStatNumber(job, "javbus_failed")}`
      },
      { label: "Sukebei 已处理", value: String(jobStatNumber(job, "sukebei_completed")) },
      {
        label: "Sukebei 成功 / 失败",
        value: `${jobStatNumber(job, "sukebei_success")} / ${jobStatNumber(job, "sukebei_failed")}`
      }
    ];
  }
  if (mode === "sehuatang_detail_progress") {
    const job = currentJob.value || {};
    const page = sehuatangPageCounts(job);
    return [
      { label: "详情进度", value: `${page.completed} / ${page.total}` },
      { label: "选中帖子", value: String(page.total) },
      { label: "已处理", value: String(page.completed) },
      { label: "抓取成功", value: String(jobStatNumber(job, "fetched_count")) },
      { label: "抓取失败", value: String(jobStatNumber(job, "failed_count")) }
    ];
  }
  if (mode === "sehuatang_list_progress") {
    const job = currentJob.value || {};
    const page = sehuatangPageCounts(job);
    const listItems = [
      { label: "页面进度", value: `${page.completed} / ${page.total}` },
      {
        label: "页面成功 / 失败",
        value: `${jobStatNumber(job, "success_page_count")} / ${jobStatNumber(job, "failed_page_count")}`
      },
      { label: "页面帖子", value: String(jobStatNumber(job, "seen_count")) },
      { label: "去重帖子", value: String(jobStatNumber(job, "matched_count")) },
      { label: "新增帖子", value: String(jobStatNumber(job, "inserted_count")) },
      { label: "更新已有", value: String(jobStatNumber(job, "existing_count")) },
      { label: "跳过已有", value: String(jobStatNumber(job, "skipped_existing_count")) },
      { label: "入库失败", value: String(jobStatNumber(job, "persist_failed_count")) },
      { label: "跳过置顶", value: String(jobStatNumber(job, "skipped_top_count")) }
    ];
    if (!isSehuatangSHTask.value) {
      return listItems;
    }
    const detailTarget = hasJobStat(job, "detail_target_count")
      ? jobStatNumber(job, "detail_target_count")
      : jobStatNumber(job, "detail_total_page_count");
    return [
      { label: "当前阶段", value: currentPhaseLabel(job) },
      ...listItems,
      { label: "详情进度", value: phaseProgressText(job, "detail") },
      { label: "详情目标", value: String(detailTarget) },
      {
        label: "详情成功 / 失败",
        value: `${jobStatNumber(job, "detail_fetched_count")} / ${jobStatNumber(job, "detail_failed_count")}`
      }
    ];
  }
  if (mode === "javbus_filtered_queue" || mode === "sukebei_filtered_queue") {
    const site = mode === "javbus_filtered_queue" ? "JavBus" : "Sukebei";
    const pending = filteredQueuePendingItems.value.length;
    const running = filteredQueueRunningItems.value.length;
    const done = filteredQueueDoneItems.value.length;
    const job = currentJob.value || {};
    return [
      { label: `${site} 队列总数`, value: String(pending + running + done) },
      { label: "待入列", value: String(pending) },
      { label: "成功", value: String(Number(pick(job, "success_count", "SuccessCount") || 0)) },
      { label: "失败", value: String(Number(pick(job, "failed_count", "FailedCount") || 0)) }
    ];
  }
  if (mode === "dailybest_stages" || mode === "seeds_stages") {
    const firstLabel = mode === "seeds_stages" ? "种子抓取" : "榜单抓取";
    const firstPhase = mode === "seeds_stages" ? "seed" : "bestinv";
    const job = currentJob.value || {};
    return [
      { label: "当前阶段", value: currentPhaseLabel(job) },
      { label: "已运行时长", value: formatElapsed(pick(job, "started_at", "StartedAt")) },
      { label: `${firstLabel}进度`, value: phaseProgressText(job, firstPhase) },
      { label: `${firstLabel}正确/错误`, value: phaseResultText(job, firstPhase) },
      { label: "详情抓取进度", value: phaseProgressText(job, "detail") },
      { label: "详情抓取正确/错误", value: phaseResultText(job, "detail") },
      { label: "封面下载进度", value: phaseProgressText(job, "cover") },
      { label: "封面下载正确/错误", value: phaseResultText(job, "cover") },
      { label: "标题翻译进度", value: phaseProgressText(job, "translate") },
      { label: "标题翻译正确/错误", value: phaseResultText(job, "translate") },
      { label: "JavBus 抓取进度", value: phaseProgressText(job, "fetch_javbus") },
      { label: "JavBus 抓取正确/错误", value: phaseResultText(job, "fetch_javbus") },
      { label: "Sukebei 抓取进度", value: phaseProgressText(job, "fetch_sukebei") },
      { label: "Sukebei 抓取正确/错误", value: phaseResultText(job, "fetch_sukebei") }
    ];
  }
  if (mode === "post_process_stages") {
    const job = currentJob.value || {};
    return [
      { label: "当前阶段", value: currentPhaseLabel(job) },
      { label: "已运行时长", value: formatElapsed(pick(job, "started_at", "StartedAt")) },
      { label: "封面下载进度", value: phaseProgressText(job, "cover") },
      { label: "封面下载正确/错误", value: phaseResultText(job, "cover") },
      { label: "标题翻译进度", value: phaseProgressText(job, "translate") },
      { label: "标题翻译正确/错误", value: phaseResultText(job, "translate") }
    ];
  }
  const job = currentJob.value || {};
  return [
    {
      label: text(asRecord(pick(payload.value, "Labels")), "Extra") || "任务类型",
      value: currentJob.value ? taskLabelByType(text(job, "task_type", "TaskType")) : "-"
    },
    { label: text(asRecord(pick(payload.value, "Labels")), "Result") || "成功/失败", value: jobResultText(job) },
    {
      label: text(asRecord(pick(payload.value, "Labels")), "Elapsed") || "已运行时长",
      value: formatElapsed(pick(job, "started_at", "StartedAt"))
    }
  ];
});

const stageOverviewGroups = computed(() => {
  if (overviewMode.value === "post_process_stages") {
    const items = overviewItems.value;
    const definitions = [
      { label: "运行状态", itemLabels: ["当前阶段", "已运行时长"] },
      { label: "封面下载", itemLabels: ["封面下载进度", "封面下载正确/错误"] },
      { label: "标题翻译", itemLabels: ["标题翻译进度", "标题翻译正确/错误"] }
    ];
    return definitions
      .map((definition) => ({
        label: definition.label,
        items: definition.itemLabels.flatMap((label) => items.find((item) => item.label === label) || [])
      }))
      .filter((group) => group.items.length);
  }
  const firstStageLabel = overviewMode.value === "dailybest_stages" ? "榜单抓取" : "种子抓取";
  const definitions = [
    { label: "运行状态", itemLabels: ["当前阶段", "已运行时长"] },
    { label: firstStageLabel, itemLabels: [`${firstStageLabel}进度`, `${firstStageLabel}正确/错误`] },
    { label: "详情抓取", itemLabels: ["详情抓取进度", "详情抓取正确/错误"] },
    { label: "封面下载", itemLabels: ["封面下载进度", "封面下载正确/错误"] },
    { label: "标题翻译", itemLabels: ["标题翻译进度", "标题翻译正确/错误"] },
    { label: "JavBus 抓取", itemLabels: ["JavBus 抓取进度", "JavBus 抓取正确/错误"] },
    { label: "Sukebei 抓取", itemLabels: ["Sukebei 抓取进度", "Sukebei 抓取正确/错误"] }
  ];
  const items = overviewItems.value;
  const remaining = new Set(items);
  const groups = definitions
    .map((definition) => {
      const groupItems = definition.itemLabels.flatMap((label) => {
        const item = items.find((candidate) => candidate.label === label);
        if (!item) {
          return [];
        }
        remaining.delete(item);
        return [item];
      });
      return { label: definition.label, items: groupItems };
    })
    .filter((group) => group.items.length);

  if (remaining.size) {
    groups.push({ label: "其他指标", items: Array.from(remaining) });
  }
  return groups;
});

const stagePhaseLabels: Record<string, string> = {
  bestinv: "榜单抓取",
  seed: "种子抓取",
  detail: "详情抓取",
  fetch_javbus: "JavBus 抓取",
  fetch_sukebei: "Sukebei 抓取",
  cover: "封面下载",
  translate: "标题翻译"
};

const phaseByEventStage: Record<string, string> = {
  pipeline_pre: "seed",
  job_started: "seed",
  bestinv_prepare: "bestinv",
  bestinv_page_done: "bestinv",
  seed_queue_ready: "seed",
  seed_item_done: "seed",
  seed_page_done: "seed",
  detail_queue_ready: "detail",
  detail_item_done: "detail",
  detail_fetch_failed: "detail",
  detail_parse_failed: "detail",
  fetch_site_after_detail_prepare: "detail",
  fetch_javbus_queue_ready: "fetch_javbus",
  fetch_javbus_done: "fetch_javbus",
  fetch_javbus_failed: "fetch_javbus",
  fetch_javbus_all_done: "fetch_javbus",
  fetch_sukebei_queue_ready: "fetch_sukebei",
  fetch_sukebei_done: "fetch_sukebei",
  fetch_sukebei_failed: "fetch_sukebei",
  fetch_sukebei_all_done: "fetch_sukebei",
  cover_queue_ready: "cover",
  cover_download_done: "cover",
  cover_download_failed: "cover",
  translate_queue_ready: "translate",
  translate_done: "translate",
  translate_failed: "translate"
};

function phaseStat(job: CrawlerRecord, phase: string) {
  const phases = asRecord(pick(job, "phase_stats", "PhaseStats"));
  const stat = asRecord(phases[phase]);
  return {
    handled: Number(pick(stat, "handled_count", "HandledCount") || 0),
    total: Number(pick(stat, "total_count", "TotalCount") || 0),
    success: Number(pick(stat, "success_count", "SuccessCount") || 0),
    failed: Number(pick(stat, "failed_count", "FailedCount") || 0)
  };
}

function phaseProgressText(job: CrawlerRecord, phase: string): string {
  const stat = phaseStat(job, phase);
  return stat.total > 0 ? `${stat.handled}/${stat.total}` : "-";
}

function phaseResultText(job: CrawlerRecord, phase: string): string {
  const stat = phaseStat(job, phase);
  return `${stat.success} / ${stat.failed}`;
}

function jobStats(job: CrawlerRecord): CrawlerRecord {
  return asRecord(pick(job, "stats", "Stats"));
}

function jobStatsText(job: CrawlerRecord): CrawlerRecord {
  return asRecord(pick(job, "stats_text", "StatsText"));
}

function hasJobStat(job: CrawlerRecord, key: string): boolean {
  return Object.prototype.hasOwnProperty.call(jobStats(job), key);
}

function jobStatNumber(job: CrawlerRecord, key: string): number {
  return Number(jobStats(job)[key] || 0);
}

function jobStatText(job: CrawlerRecord, key: string): string {
  return String(jobStatsText(job)[key] || "").trim();
}

function shExistingPolicyLabel(policy: string): string {
  return String(policy || "")
    .trim()
    .toLowerCase() === "skip"
    ? "跳过已有"
    : "Upsert（更新已有）";
}

function sehuatangPageCounts(job: CrawlerRecord): { completed: number; total: number } {
  const listPhase = phaseStat(job, "list");
  const completed = jobStatNumber(job, "completed_page_count");
  const total = jobStatNumber(job, "total_page_count");
  if (total > 0) {
    return { completed, total };
  }
  return { completed: listPhase.handled, total: listPhase.total };
}

function sehuatangCurrentProgressCounts(job: CrawlerRecord): { completed: number; total: number } {
  const phaseKey = jobCurrentPhaseKey(job);
  if (phaseKey) {
    const phase = phaseStat(job, phaseKey);
    if (phase.total > 0) {
      return { completed: phase.handled, total: phase.total };
    }
  }

  if (phaseKey === "detail" || isSHDetailOverview.value) {
    const taskType = text(job, "task_type", "TaskType");
    const detailPrefix = taskType === "spider_fetch_sehuatang_sh_threads" ? "detail_" : "";
    const completed = jobStatNumber(job, `${detailPrefix}completed_page_count`);
    const total = jobStatNumber(job, `${detailPrefix}total_page_count`);
    if (total > 0) {
      return { completed, total };
    }
  }

  return sehuatangPageCounts(job);
}

function jobCurrentPhaseKey(job: CrawlerRecord): string {
  const stage = text(job, "stage", "Stage");
  const dailyBestInitialPhase =
    overviewMode.value === "dailybest_stages" && ["pipeline_pre", "job_started"].includes(stage) ? "bestinv" : "";
  return text(job, "current_phase_key", "CurrentPhaseKey") || dailyBestInitialPhase || phaseByEventStage[stage] || "";
}

function currentPhaseLabel(job: CrawlerRecord): string {
  const phase = jobCurrentPhaseKey(job);
  return phase ? stagePhaseLabels[phase] || phase : "-";
}

function formatElapsed(value: unknown): string {
  const currentTime = nowTick.value;
  const startedAt = Number(value || 0);
  if (!startedAt) {
    return "00:00";
  }
  const delta = Math.max(0, Math.floor(currentTime / 1000) - startedAt);
  const hours = Math.floor(delta / 3600);
  const minutes = Math.floor((delta % 3600) / 60);
  const seconds = delta % 60;
  if (hours > 0) {
    return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
  }
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

function formatJobElapsed(job: CrawlerRecord): string {
  const startedAt = Number(pick(job, "started_at", "StartedAt") || 0);
  if (!startedAt) {
    return "00:00";
  }
  const finishedAt = pick(job, "done", "Done") ? Number(pick(job, "at", "At") || 0) : 0;
  if (!finishedAt) {
    return formatElapsed(startedAt);
  }
  const delta = Math.max(0, finishedAt - startedAt);
  const hours = Math.floor(delta / 3600);
  const minutes = Math.floor((delta % 3600) / 60);
  const seconds = delta % 60;
  if (hours > 0) {
    return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
  }
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

function jobElapsedSeconds(job: CrawlerRecord): number {
  const startedAt = Number(pick(job, "started_at", "StartedAt") || 0);
  if (startedAt <= 0) {
    return 0;
  }
  const finishedAt = pick(job, "done", "Done") ? Number(pick(job, "at", "At") || 0) : Math.floor(nowTick.value / 1000);
  return Math.max(0, finishedAt - startedAt);
}

function formatJobEstimatedRemaining(job: CrawlerRecord, completed: number, total: number): string {
  if (pick(job, "done", "Done")) {
    return "00:00:00";
  }
  return formatTaskEta(estimateRemainingSeconds(jobElapsedSeconds(job), completed, total));
}

function progressPercent(job: CrawlerRecord): number {
  const handled = Number(pick(job, "handled_count", "HandledCount") || 0);
  const queued = Number(pick(job, "queued_count", "QueuedCount") || 0);
  const total = handled + queued;
  return total > 0 ? Math.max(0, Math.min(100, Math.round((handled * 100) / total))) : 0;
}

function stageLabel(stage: string): string {
  const labels: Record<string, string> = {
    job_started: "任务已启动",
    detail_paused: "详情抓取已暂停",
    detail_resumed: "详情抓取已恢复",
    paused: "任务已暂停",
    resumed: "任务已继续",
    done: "任务完成",
    failed: "任务失败"
  };
  return labels[stage] || stage || emptyStateText.value;
}

function normalizeLogLevel(value: unknown): "info" | "warn" | "error" {
  const level = String(value || "")
    .trim()
    .toLowerCase();
  if (level === "warning" || level === "warn") {
    return "warn";
  }
  return level === "error" ? "error" : "info";
}

function eventTimeText(value: unknown): string {
  const date = Number(value || 0) > 0 ? new Date(Number(value) * 1000) : new Date();
  return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}:${String(date.getSeconds()).padStart(2, "0")}`;
}

function escapeEventHTML(value: unknown): string {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function formatEventLineHTML(value: string): string {
  const parts = value.split(" | ");
  const linkMovieCodes = (source: string) =>
    escapeEventHTML(source).replace(/\b([A-Za-z][A-Za-z0-9]*-\d{2,})\b/g, (_match, code: string) => {
      return `<a href="/moviedetail/${encodeURIComponent(code)}">${escapeEventHTML(code)}</a>`;
    });
  if (parts.length < 5 || !/^\[[^\]]+\]$/.test(parts[0] || "")) {
    return linkMovieCodes(value);
  }
  const subjectID = String(parts[2] || "").trim();
  if (!/^\d+$/.test(subjectID) || !String(parts[3] || "").startsWith("new_items=")) {
    return linkMovieCodes(value);
  }
  const localHref = `/moviedetail/${encodeURIComponent(subjectID)}`;
  const doubanHref = `https://movie.douban.com/subject/${encodeURIComponent(subjectID)}/`;
  const suffix = parts.slice(3).map(escapeEventHTML).join(" | ");
  return `${escapeEventHTML(parts[0])} | <a href="${localHref}" target="_blank" rel="noopener noreferrer">${escapeEventHTML(parts[1])}</a> | <a href="${doubanHref}" target="_blank" rel="noopener noreferrer">${escapeEventHTML(subjectID)}</a> | ${suffix}`;
}

function buildEventEntry(
  event: CrawlerRecord,
  raw: boolean
): { level: "info" | "warn" | "error"; html: string } | null {
  const prefix = `[${eventTimeText(pick(event, "at", "At"))}]`;
  const level = normalizeLogLevel(pick(event, "level", "Level"));
  if (raw) {
    return { level, html: escapeEventHTML(`${prefix} ${JSON.stringify(event, null, 2)}`) };
  }
  const kind = text(event, "kind", "Kind");
  if (kind === "progress") {
    const stage = text(event, "stage", "Stage");
    const parts = [`${prefix} ${text(event, "message", "Message") || stageLabel(stage)}`];
    const handled = Number(pick(event, "handled_count", "HandledCount") || 0);
    const queued = Number(pick(event, "queued_count", "QueuedCount") || 0);
    const success = Number(pick(event, "success_count", "SuccessCount") || 0);
    const failed = Number(pick(event, "failed_count", "FailedCount") || 0);
    if (handled || queued) {
      parts.push(`handled=${handled}`, `queued=${queued}`);
    }
    if (success || failed) {
      parts.push(`success=${success}`, `failed=${failed}`);
    }
    return {
      level: stage === "failed" ? "error" : stage === "paused" ? "warn" : "info",
      html: formatEventLineHTML(parts.join(" | "))
    };
  }
  if (kind !== "log") {
    return null;
  }
  const line = text(event, "line", "Line").trim();
  const match = line.match(/^\[([^\]]+)\]\s+(INFO|WARN|WARNING|ERROR)\s+(.*)$/i);
  if (match) {
    return { level: normalizeLogLevel(match[2]), html: formatEventLineHTML(`[${match[1]}] ${match[3]}`) };
  }
  return { level, html: formatEventLineHTML(`${prefix} ${line || text(event, "message", "Message")}`) };
}

function eventKey(event: CrawlerRecord): string {
  const id = Number(pick(event, "id", "ID") || 0);
  return `${text(event, "job_id", "JobID")}:${id || `${text(event, "kind", "Kind")}:${Number(pick(event, "at", "At") || 0)}:${text(event, "line", "Line", "message", "Message")}`}`;
}

function initialTargetForumIDs(): string[] {
  const values = asArray<string>(pick(payload.value, "FetchSehuatangSHTargetForums"))
    .map((item) => String(item || "").trim())
    .filter(Boolean);
  const available = new Set(shForums.value.map((item) => item.forumID));
  const selected = values.filter((item) => available.has(item));
  const normalized = selected.length ? selected : shForums.value.map((item) => item.forumID);
  return normalized;
}

function selectAllTargetForums() {
  targetForumIDs.value = shForums.value.map((forum) => forum.forumID);
}

function selectTargetForum(forumID: string) {
  targetForumIDs.value = forumID ? [forumID] : [];
}

function initialSHDetailStatusValues(): string[] {
  const values = asArray<string>(pick(payload.value, "FetchSehuatangSHDetailStatuses"))
    .map((item) => String(item || "").trim())
    .filter(Boolean);
  if (values.length) {
    return values;
  }
  const fallback = text(payload.value, "FetchSehuatangSHDetailStatus");
  return ["failed", "fetched"].includes(fallback) ? [fallback] : ["pending"];
}

function taskType(task: CrawlerRecord): string {
  return text(task, "TaskType", "task_type");
}

function toggleTranslateStatus(status: string): void {
  const selected = new Set(translateStatuses.value);
  if (selected.has(status)) {
    selected.delete(status);
  } else {
    selected.add(status);
  }
  translateStatuses.value = translateStatusOptions.map((option) => option.value).filter((value) => selected.has(value));
}

function taskLabel(task: CrawlerRecord): string {
  return text(task, "Label", "label") || taskType(task);
}

function taskLabelByType(type: string): string {
  const matched = taskButtons.value.find((item) => taskType(item) === type);
  return matched ? taskLabel(matched) : type;
}

function jobID(job: CrawlerRecord): string {
  return text(job, "job_id", "JobID", "id", "ID");
}

function jobSortTs(job: CrawlerRecord): number {
  return Number(pick(job, "at", "At", "started_at", "StartedAt") || 0);
}

function jobResultText(job: CrawlerRecord): string {
  return `${Number(pick(job, "success_count", "SuccessCount") || 0)} / ${Number(pick(job, "failed_count", "FailedCount") || 0)}`;
}

function normalizeJobID(value: unknown): string {
  const id = String(value || "").trim();
  return id && /^[A-Za-z0-9_.:-]+$/.test(id) ? id : "";
}

function storedSelectedJobID(): string {
  try {
    const raw = window.localStorage.getItem(selectedJobStorageKey.value);
    const id = normalizeJobID(raw);
    if (raw && !id) {
      window.localStorage.removeItem(selectedJobStorageKey.value);
    }
    return id;
  } catch {
    return "";
  }
}

function persistSelectedJobID(id: string) {
  try {
    if (id) {
      window.localStorage.setItem(selectedJobStorageKey.value, id);
    } else {
      window.localStorage.removeItem(selectedJobStorageKey.value);
    }
  } catch {
    // localStorage may be unavailable in restricted browser contexts.
  }
}

function syncSelectedJobIDToURL(id: string) {
  const url = new URL(window.location.href);
  if (id) {
    url.searchParams.set("job_id", id);
  } else {
    url.searchParams.delete("job_id");
  }
  window.history.replaceState(window.history.state, "", `${url.pathname}${url.search}${url.hash}`);
}

function isVisibleTaskType(type: string): boolean {
  return !allowedTaskTypes.value.length || allowedTaskTypes.value.includes(type);
}

function ensureElapsedTimer() {
  if (elapsedTimer) {
    return;
  }
  elapsedTimer = window.setInterval(() => {
    nowTick.value = Date.now();
  }, 1000);
}

function clearJobEvents() {
  rawJobEvents.value = [];
  rawJobEventKeys.clear();
  lastEventID = 0;
}

function appendJobEvent(event: CrawlerRecord) {
  const type = text(event, "task_type", "TaskType");
  const selected = selectedJobID.value.trim();
  const id = text(event, "job_id", "JobID");
  if (!isVisibleTaskType(type) || (selected && id !== selected)) {
    return;
  }
  const key = eventKey(event);
  if (rawJobEventKeys.has(key)) {
    return;
  }
  rawJobEventKeys.add(key);
  const nextEvents = [event, ...rawJobEvents.value];
  if (nextEvents.length > jobEventHistorySize) {
    const removedEvents = nextEvents.splice(jobEventHistorySize);
    removedEvents.forEach((item) => rawJobEventKeys.delete(eventKey(item)));
  }
  rawJobEvents.value = nextEvents;
}

function mergeJobEvent(event: CrawlerRecord) {
  if (text(event, "kind", "Kind") !== "progress") {
    return;
  }
  const id = text(event, "job_id", "JobID");
  const type = text(event, "task_type", "TaskType");
  if (!id || !isVisibleTaskType(type)) {
    return;
  }
  const index = jobs.value.findIndex((item) => jobID(item) === id);
  const current = index >= 0 ? jobs.value[index] : {};
  const stage = text(event, "stage", "Stage") || text(current, "stage", "Stage");
  const next: CrawlerRecord = {
    ...current,
    ...event,
    job_id: id,
    task_type: type || text(current, "task_type", "TaskType"),
    stage,
    done: Boolean(pick(event, "done", "Done")),
    paused: stage === "paused" ? true : stage === "resumed" ? false : Boolean(pick(current, "paused", "Paused"))
  };
  const list = [...jobs.value];
  if (index >= 0) {
    list[index] = next;
  } else {
    list.push(next);
  }
  jobs.value = list;
}

function handleIncomingEvent(event: CrawlerRecord) {
  const id = Number(pick(event, "id", "ID") || 0);
  if (id > lastEventID) {
    lastEventID = id;
  }
  const type = text(event, "task_type", "TaskType");
  if (!isVisibleTaskType(type)) {
    return;
  }
  mergeJobEvent(event);
  appendJobEvent(event);
}

async function loadJobEvents(id: string, reportError: boolean): Promise<boolean> {
  try {
    const data = asRecord(await getJSON(`/api/crawler/jobs/${encodeURIComponent(id)}/events`));
    if (selectedJobID.value !== id) {
      return false;
    }
    const events = asArray<CrawlerRecord>(pick(data, "events", "Events"));
    const visible = events.filter(
      (event) => text(event, "job_id", "JobID") === id && isVisibleTaskType(text(event, "task_type", "TaskType"))
    );
    lastEventID = events.reduce((max, event) => Math.max(max, Number(pick(event, "id", "ID") || 0)), 0);
    rawJobEvents.value = [...visible].reverse().slice(0, jobEventHistorySize);
    rawJobEventKeys.clear();
    rawJobEvents.value.forEach((event) => rawJobEventKeys.add(eventKey(event)));
    visible.forEach(mergeJobEvent);
    return true;
  } catch (err) {
    if (reportError && selectedJobID.value === id) {
      message.value = err instanceof Error ? err.message : String(err);
    }
    return false;
  }
}

function closeEventStream() {
  jobEventStream.stop();
  if (eventReconnectTimer) {
    window.clearTimeout(eventReconnectTimer);
    eventReconnectTimer = undefined;
  }
}

function scheduleEventReconnect() {
  if (eventReconnectTimer) {
    return;
  }
  eventReconnectTimer = window.setTimeout(() => {
    eventReconnectTimer = undefined;
    connectEventStream();
  }, 1500);
}

function connectEventStream() {
  const selected = normalizeJobID(selectedJobID.value);
  const key = selected ? `job:${selected}` : "group";
  if (jobEventStream.activeKey.value === key) {
    return;
  }
  closeEventStream();
  const base = selected ? `/api/crawler/jobs/${encodeURIComponent(selected)}/stream` : "/api/crawler/jobs/stream";
  const url = lastEventID > 0 ? `${base}?last_event_id=${encodeURIComponent(String(lastEventID))}` : base;
  jobEventStream.connect(key, url, {
    onMessage(data) {
      try {
        handleIncomingEvent(asRecord(JSON.parse(data)));
      } catch {
        message.value = "事件解析失败";
      }
    },
    onError() {
      if (!selected || !currentJobDone.value) scheduleEventReconnect();
    }
  });
}

async function openJob(value: string, reportError = true) {
  const id = normalizeJobID(value);
  closeEventStream();
  clearJobEvents();
  selectedJobID.value = id;
  persistSelectedJobID(id);
  syncSelectedJobIDToURL(id);
  if (!id) {
    connectEventStream();
    return;
  }
  const loaded = await loadJobEvents(id, reportError);
  if (!loaded) {
    selectedJobID.value = "";
    persistSelectedJobID("");
    syncSelectedJobIDToURL("");
    clearJobEvents();
  }
  connectEventStream();
}

function ensureJobsPolling() {
  startJobsPolling();
}

async function loadJobs(silent = false) {
  try {
    const data = asRecord(await getJSON("/api/crawler/jobs"));
    jobs.value = asArray<CrawlerRecord>(pick(data, "jobs", "Jobs", "items", "Items"));
  } catch (err) {
    if (!silent) {
      message.value = err instanceof Error ? err.message : String(err);
    }
  }
}

function fetchSiteFilterForm(): HTMLFormElement | null {
  return document.querySelector("#fetchSiteFilterWrap form");
}

function namedField(name: string): HTMLInputElement | HTMLSelectElement | null {
  return document.querySelector(`#fetchSiteFilterPanel [name="${name}"]`);
}

function namedFieldValue(name: string): string {
  const field = namedField(name);
  return field ? String(field.value || "").trim() : "";
}

function setNamedFieldValue(name: string, value: string) {
  let field = namedField(name);
  const formEl = fetchSiteFilterForm();
  if (!field && formEl && name === "od") {
    const input = document.createElement("input");
    input.type = "hidden";
    input.name = name;
    formEl.appendChild(input);
    field = input;
  }
  if (field) {
    field.value = value;
  }
}

function appendMovieFilters(body: Record<string, unknown>) {
  movieFilterFieldNames.forEach((name) => {
    const value = namedFieldValue(name);
    if (value) {
      body[name] = value;
    }
  });
}

function appendFetchSiteDurationFields(body: Record<string, unknown>) {
  fetchSiteDurationFieldNames.forEach((name) => {
    const value = namedFieldValue(name) || "5";
    body[name] = value;
  });
}

function autoFetchSiteChecked(): boolean {
  const input = document.querySelector<HTMLInputElement>("#task_auto_fetch_site");
  return input ? input.checked : true;
}

function buildTaskPayload(type: string): Record<string, unknown> {
  const body: Record<string, unknown> = { task_type: type };
  if (type === "spider_seed_by_name") {
    body.name = form.name;
  }
  if (type === "spider_refresh_oldest_detail") {
    body.number = form.number;
  }
  if (type === "spider_rebuild_actor_rank") {
    body.actor_name = form.actorName;
  }
  if ((type === "spider_translate_title" || type === "spider_post_process") && translateStatuses.value.length) {
    body.statuses = translateStatuses.value;
  }
  if (fetchSiteTaskTypes.has(type)) {
    appendMovieFilters(body);
    appendFetchSiteDurationFields(body);
  }
  if (autoFetchSiteTaskTypes.has(type)) {
    body.auto_fetch_site = autoFetchSiteChecked() ? "1" : "0";
  }
  if (type === "spider_fetch_sehuatang_sh_threads") {
    if (!targetForumIDs.value.length) {
      throw new Error("请选择目标版块");
    }
    body.fetch_mode = "forum";
    body.existing_policy = form.existingPolicy;
    body.keyword = "";
    body.start_page = form.startPage;
    body.end_page = form.endPage;
    body.target_forum_ids = targetForumIDs.value;
  }
  if (type === "spider_fetch_sehuatang_sh_thread_list") {
    if (!targetForumIDs.value.length) {
      throw new Error("请选择目标版块");
    }
    body.start_page = form.startPage;
    body.end_page = form.endPage;
    body.target_forum_ids = targetForumIDs.value;
  }
  if (type === "spider_fetch_sehuatang_sh_thread_detail") {
    if (!shDetailPreviewSignature.value || shDetailPreviewSignature.value !== shDetailFilterSignature.value) {
      throw new Error("筛选条件已变化，请重新执行帖子预览");
    }
    const tids = shDetailPreviewThreadIDs.value.map((item) => item.trim()).filter(Boolean);
    if (!tids.length) {
      throw new Error("请先预览筛选结果，确认存在可抓取帖子");
    }
    body.thread_tids = tids;
    body.trigger_source = "shfill";
    body.target_forum_ids = targetForumIDs.value;
    body.detail_statuses = shDetailStatusValues.value;
    body.post_time_from = form.postTimeFrom.trim();
    body.post_time_to = form.postTimeTo.trim();
  }
  return body;
}

function toggleFetchSiteFilter() {
  if (!isFetchSiteTask.value) {
    return;
  }
  fetchSiteFilterOpen.value = !fetchSiteFilterOpen.value;
}

function setFetchSiteOrderBy(value: string) {
  fetchSiteOrderBy.value = value;
  setNamedFieldValue("od", value);
  hideFetchSitePreview();
}

function clearMovieFilters() {
  movieFilterFieldNames.forEach((name) => {
    setNamedFieldValue(name, name === "od" ? "rd" : "");
  });
  fetchSiteDurationFieldNames.forEach((name) => {
    setNamedFieldValue(name, "5");
  });
  fetchSiteOrderBy.value = "rd";
  hideFetchSitePreview();
}

function hideFetchSitePreview() {
  resetFetchSitePreview();
  fetchSitePreviewTargets.value = [];
  fetchSitePreviewSignature.value = "";
  fetchSitePreviewAggregate.value = {};
  fetchSitePreviewError.value = "";
  fetchSitePreviewTaskType.value = "";
}

function invalidateFetchSitePreview() {
  if (
    fetchSitePreviewOpen.value ||
    fetchSitePreviewLoading.value ||
    fetchSitePreviewTargets.value.length > 0 ||
    fetchSitePreviewSignature.value
  ) {
    hideFetchSitePreview();
  }
}

function hideSHDetailPreview() {
  resetSHDetailPreview();
  shDetailPreviewThreadIDs.value = [];
  shDetailPreviewAggregate.value = {};
  shDetailPreviewTotal.value = 0;
  shDetailPreviewError.value = "";
  shDetailPreviewSignature.value = "";
  shDetailPreviewPoolHref.value = "";
}

function collapseSHDetailPreview() {
  shDetailPreviewOpen.value = false;
}

function currentSHDetailFilterPayload(): Record<string, unknown> {
  if (!targetForumIDs.value.length) {
    throw new Error("请选择目标版块");
  }
  const body: Record<string, unknown> = {
    task_type: "spider_fetch_sehuatang_sh_thread_detail",
    target_forum_ids: targetForumIDs.value,
    detail_statuses: shDetailStatusValues.value,
    post_time_from: form.postTimeFrom,
    post_time_to: form.postTimeTo
  };
  if (form.detailLimit !== undefined && form.detailLimit !== null && String(form.detailLimit).trim() !== "") {
    const limit = Number(form.detailLimit);
    if (!Number.isFinite(limit) || limit <= 0) {
      throw new Error("请填写合法的 limit");
    }
    body.limit = Math.floor(limit);
  }
  return body;
}

function buildSHDetailPoolHref(body: Record<string, unknown>): string {
  const params = new URLSearchParams();
  params.set("p", "1");
  asArray<unknown>(body.target_forum_ids).forEach((value) => {
    const current = String(value || "").trim();
    if (current) {
      params.append("source_forum_ids", current);
    }
  });
  asArray<unknown>(body.detail_statuses).forEach((value) => {
    const current = String(value || "").trim();
    if (current) {
      params.append("detail_statuses", current);
    }
  });
  const postTimeFrom = String(body.post_time_from || "").trim();
  const postTimeTo = String(body.post_time_to || "").trim();
  if (postTimeFrom) {
    params.set("post_time_from", postTimeFrom);
  }
  if (postTimeTo) {
    params.set("post_time_to", postTimeTo);
  }
  return `/shpool?${params.toString()}`;
}

async function previewSHDetail() {
  let body: Record<string, unknown>;
  try {
    body = currentSHDetailFilterPayload();
  } catch (err) {
    message.value = err instanceof Error ? err.message : String(err);
    return;
  }
  message.value = "";
  shDetailPreviewLoading.value = true;
  shDetailPreviewOpen.value = true;
  shDetailPreviewError.value = "";
  shDetailPreviewSignature.value = "";
  shDetailPreviewThreadIDs.value = [];
  shDetailPreviewAggregate.value = {};
  shDetailPreviewTotal.value = 0;
  shDetailPreviewPoolHref.value = "";
  try {
    const data = asRecord(await sendJSON("/api/crawler/jobs/sehuatang-sh-thread-detail/preview", "POST", body));
    shDetailPreviewThreadIDs.value = Array.from(
      new Set(
        asArray<unknown>(pick(data, "thread_tids", "ThreadTIDs"))
          .map((item) => String(item || "").trim())
          .filter(Boolean)
      )
    );
    shDetailPreviewAggregate.value = asRecord(pick(data, "aggregate", "Aggregate"));
    shDetailPreviewTotal.value = Number(pick(data, "total", "Total") || 0);
    shDetailPreviewPoolHref.value = buildSHDetailPoolHref(body);
    shDetailPreviewSignature.value = shDetailFilterSignature.value;
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : String(err);
    shDetailPreviewError.value = errorMessage;
    message.value = errorMessage;
  } finally {
    shDetailPreviewLoading.value = false;
  }
}

async function previewFetchSite() {
  const type = currentTaskType.value;
  if (!fetchSiteTaskTypes.has(type)) {
    message.value = "当前任务不支持筛选预览";
    return;
  }
  let body: Record<string, unknown>;
  try {
    body = buildTaskPayload(type);
  } catch (err) {
    message.value = err instanceof Error ? err.message : String(err);
    return;
  }
  message.value = "";
  fetchSitePreviewLoading.value = true;
  fetchSitePreviewOpen.value = true;
  fetchSitePreviewTargets.value = [];
  fetchSitePreviewSignature.value = "";
  fetchSitePreviewAggregate.value = {};
  fetchSitePreviewError.value = "";
  fetchSitePreviewTaskType.value = type;
  try {
    const data = asRecord(await sendJSON("/api/crawler/jobs/fetch-site/preview", "POST", body));
    fetchSitePreviewTargets.value = asArray<CrawlerRecord>(pick(data, "targets", "Targets"));
    fetchSitePreviewSignature.value = text(data, "signature", "Signature");
    fetchSitePreviewAggregate.value = asRecord(pick(data, "aggregate", "Aggregate"));
    fetchSitePreviewTaskType.value = text(data, "task_type", "TaskType") || type;
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : String(err);
    fetchSitePreviewError.value = errorMessage;
    message.value = errorMessage;
  } finally {
    fetchSitePreviewLoading.value = false;
  }
}

async function startDefault() {
  await startTask({ TaskType: selectedTaskType.value || text(payload.value, "DefaultTaskType"), Label: "默认任务" });
}

async function startTask(task: CrawlerRecord) {
  const type = taskType(task);
  if (!type) {
    message.value = "缺少 task_type";
    return;
  }
  let body: Record<string, unknown>;
  if (fetchSiteTaskTypes.has(type) && showFetchSitePanel.value) {
    if (
      !fetchSitePreviewTargets.value.length ||
      !fetchSitePreviewSignature.value ||
      fetchSitePreviewTaskType.value !== type
    ) {
      message.value = "请先生成当前任务的目标清单";
      return;
    }
    body = { task_type: type, fetch_site_targets: fetchSitePreviewTargets.value };
  } else {
    try {
      body = buildTaskPayload(type);
    } catch (err) {
      message.value = err instanceof Error ? err.message : String(err);
      return;
    }
  }
  try {
    const data = asRecord(await sendJSON("/api/crawler/jobs/start", "POST", body));
    const id = normalizeJobID(text(data, "job_id", "JobID", "id", "ID"));
    message.value = text(data, "message", "Message") || `已启动 ${type}`;
    await loadJobs(true);
    await openJob(id);
  } catch (err) {
    message.value = err instanceof Error ? err.message : String(err);
  }
}

async function controlSelected(action: "pause" | "resume" | "stop") {
  const id = selectedJobID.value.trim() || jobID(currentJob.value || {});
  if (!id) {
    message.value = "当前没有可操作的任务";
    return;
  }
  try {
    const data = asRecord(await sendJSON(`/api/crawler/jobs/${encodeURIComponent(id)}/${action}`, "POST"));
    const label = action === "pause" ? "暂停" : action === "resume" ? "继续" : "停止";
    message.value = text(data, "message", "Message") || `已${label}任务`;
    await loadJobs(true);
  } catch (err) {
    message.value = err instanceof Error ? err.message : String(err);
  }
}
</script>
