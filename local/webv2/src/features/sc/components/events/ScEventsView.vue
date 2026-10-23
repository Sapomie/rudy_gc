<template>
  <PageShell
    :class="[
      'movie-agg-page-shell',
      'sc-events-page-shell',
      {
        'next02-sc-event-detail-shell': viewType === 'sc-event-detail',
        'cards-page-shell': viewType === 'sc-event-detail',
        'next02-pink-cards-page': viewType === 'sc-event-detail',
        'next02-pink-cards-page-shell': viewType === 'sc-event-detail'
      }
    ]"
    :title="title"
    :note="pageNote"
    :eyebrow="viewType === 'sc-event-detail' ? 'SC Event' : ''"
  >
    <div
      id="page-root"
      class="movie-agg-page-root"
      :class="{
        'container-fluid sc-events-page-root': viewType !== 'sc-event-detail',
        'next02-sc-pink-theme': viewType === 'sc-events' || viewType === 'sc-event-cards',
        'next02-sc-events-rewrite': viewType === 'sc-events',
        'next02-sc-event-cards-rewrite': viewType === 'sc-event-cards',
        'next02-sc-event-detail-root': viewType === 'sc-event-detail'
      }"
    >
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <template v-if="viewType === 'sc-events'">
          <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell gc-page-head">
            <div class="next02-sc-hero-copy">
              <span class="next02-sc-eyebrow">SC / EVENT STREAM</span>
              <h1>SC Events</h1>
              <p>沿时间线浏览 SC 事件，并快速进入本次影片、Come Movie 与演员信息。</p>
            </div>
            <div class="next02-sc-hero-actions">
              <ScViewSwitch />
              <a href="/databackfill" class="btn btn-outline-secondary" target="_blank" rel="noopener noreferrer"
                >回填任务台</a
              >
            </div>
          </section>

          <div class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card">
            <div class="card-body">
              <div class="next02-sc-panel-head">
                <form class="sc-event-filter-form" method="get" action="/scevents">
                  <input type="hidden" name="p" value="1" />
                  <input v-if="routeQueryValue('sort')" type="hidden" name="sort" :value="routeQueryValue('sort')" />
                  <input v-if="routeQueryValue('order')" type="hidden" name="order" :value="routeQueryValue('order')" />
                  <label class="sc-event-filter-field">
                    <span>起始日期</span>
                    <input
                      class="smart-pick-date-input"
                      type="text"
                      name="start_date"
                      :value="routeQueryValue('start_date')"
                      placeholder="YYYY-MM-DD"
                      inputmode="numeric"
                      maxlength="10"
                      pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                    />
                  </label>
                  <label class="sc-event-filter-field">
                    <span>结束日期</span>
                    <input
                      class="smart-pick-date-input"
                      type="text"
                      name="end_date"
                      :value="routeQueryValue('end_date')"
                      placeholder="YYYY-MM-DD"
                      inputmode="numeric"
                      maxlength="10"
                      pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}"
                    />
                  </label>
                  <label class="sc-event-filter-field sc-event-option-filter-field">
                    <span>Kind</span>
                    <select name="kind_option_id" :value="routeQueryValue('kind_option_id')">
                      <option value="">全部</option>
                      <option v-for="option in scOptions.kind" :key="option.id" :value="option.id">
                        {{ option.optionName }}{{ option.isEnabled === 1 ? "" : "（已停用）" }}
                      </option>
                    </select>
                  </label>
                  <label class="sc-event-filter-field sc-event-option-filter-field">
                    <span>FG</span>
                    <select name="fg_option_id" :value="routeQueryValue('fg_option_id')">
                      <option value="">全部</option>
                      <option v-for="option in scOptions.fg" :key="option.id" :value="option.id">
                        {{ option.optionName }}{{ option.isEnabled === 1 ? "" : "（已停用）" }}
                      </option>
                    </select>
                  </label>
                  <label class="sc-event-filter-field sc-event-option-filter-field">
                    <span>Vessel</span>
                    <select name="vessel_option_id" :value="routeQueryValue('vessel_option_id')">
                      <option value="">全部</option>
                      <option v-for="option in scOptions.vessel" :key="option.id" :value="option.id">
                        {{ option.optionName }}{{ option.isEnabled === 1 ? "" : "（已停用）" }}
                      </option>
                    </select>
                  </label>
                  <label class="sc-event-filter-field sc-event-page-size-field">
                    <span>每页</span>
                    <input type="number" step="1" name="ps" :value="filterPageSize" />
                  </label>
                  <button type="submit" class="sc-event-filter-submit">筛选</button>
                  <a href="/scevents" class="sc-event-filter-clear">清空</a>
                </form>
                <div v-if="items.length" class="next02-sc-hero-metrics sc-event-list-metrics" aria-label="SC 事件摘要">
                  <span
                    ><strong>{{ total }}</strong> Events</span
                  >
                  <span
                    ><strong>{{ items.length }}</strong> Current Page</span
                  >
                </div>
                <div v-if="items.length" class="next02-sc-movie-pager">
                  <PaginationBar :links="pageLinks" />
                </div>
              </div>
              <div class="table-responsive table-scroll">
                <table class="table table-hover align-middle mb-0 sc-events-table data-table">
                  <colgroup class="sc-events-colgroup">
                    <col class="sc-events-col-time" />
                    <col class="sc-events-col-total" />
                    <col class="sc-events-col-count" />
                    <col class="sc-events-col-jav" />
                    <col class="sc-events-col-fc2" />
                    <col class="sc-events-col-movie" />
                    <col class="sc-events-col-cooldown" />
                    <col class="sc-events-col-duration" />
                    <col class="sc-events-col-kind" />
                    <col class="sc-events-col-cast" />
                    <col class="sc-events-col-vessel" />
                    <col class="sc-events-col-fg" />
                  </colgroup>
                  <thead>
                    <tr>
                      <th class="sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByTime')" label="时间" class="sc-sort-link" />
                      </th>
                      <th class="sc-number sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByTotalMovieCnt')" label="总片数" class="sc-sort-link" />
                      </th>
                      <th class="sc-number sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByMovieCnt')" label="有 SC" class="sc-sort-link" />
                      </th>
                      <th class="sc-number sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByJavMovieCnt')" label="JAV" class="sc-sort-link" />
                      </th>
                      <th class="sc-number sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByFC2MovieCnt')" label="FC2" class="sc-sort-link" />
                      </th>
                      <th class="sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByComeMovie')" label="Come 片" class="sc-sort-link" />
                      </th>
                      <th class="sc-number sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByCooldown')" label="Cooldown" class="sc-sort-link" />
                      </th>
                      <th class="sc-number sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByDuration')" label="时长" class="sc-sort-link" />
                      </th>
                      <th class="sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByKind')" label="种类" class="sc-sort-link" />
                      </th>
                      <th class="sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByMovieCast')" label="演员" class="sc-sort-link" />
                      </th>
                      <th class="sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByVessel')" label="Vessel" class="sc-sort-link" />
                      </th>
                      <th class="sc-sortable-head">
                        <SharedSortLink :link="sortItem('ByFg')" label="FG" class="sc-sort-link" />
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="(item, index) in items"
                      :key="eventName(eventOf(item))"
                      :class="eventRunningRowClass(item, index)"
                    >
                      <td>
                        <a
                          class="sc-event-link"
                          :href="eventDetailHref(eventName(eventOf(item)))"
                          :title="eventName(eventOf(item))"
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          <span class="sc-event-time"
                            ><span class="sc-event-date">{{ formatUnix(pick(eventOf(item), "ScTime")) }}</span
                            ><small class="sc-event-hour">{{
                              formatUnixHm(pick(eventOf(item), "ScTime"))
                            }}</small></span
                          >
                        </a>
                      </td>
                      <td class="sc-number sc-primary-value">{{ text(item, "TotalMovieCount") }}</td>
                      <td class="sc-number">{{ text(item, "ScMovieCount") }}</td>
                      <td class="sc-number">{{ text(item, "JavMovieCount") || "0" }}</td>
                      <td class="sc-number">{{ text(item, "FC2MovieCount") || "0" }}</td>
                      <td class="sc-content-cell">
                        <span v-if="text(eventOf(item), 'ComeMovieName')" class="sc-event-come-movie">
                          <span
                            :class="['sc-event-come-pill', movieSourceClass(text(item, 'ComeMovieType'))]"
                            :aria-label="`${movieSourceLabel(text(item, 'ComeMovieType'))} 影片：${text(eventOf(item), 'ComeMovieName')}`"
                          >
                            <LinkOrMuted
                              :href="text(item, 'ComeMovieHref')"
                              :value="text(eventOf(item), 'ComeMovieName')"
                              new-window
                            />
                          </span>
                        </span>
                        <span v-else class="text-muted">-</span>
                      </td>
                      <td class="sc-number">
                        <span v-if="num(eventOf(item), 'Cooldown') > 0">{{
                          humanDays(pick(eventOf(item), "Cooldown"))
                        }}</span
                        ><span v-else class="text-muted">-</span>
                      </td>
                      <td class="sc-number sc-muted">
                        <span v-if="num(eventOf(item), 'DurationMinutes') > 0">{{
                          text(eventOf(item), "DurationMinutes")
                        }}</span
                        ><span v-else class="text-muted">-</span>
                      </td>
                      <td class="sc-muted">
                        <span v-if="text(eventOf(item), 'Kind')">{{ text(eventOf(item), "Kind") }}</span
                        ><span v-else class="text-muted">-</span>
                      </td>
                      <td class="sc-muted sc-content-cell">
                        <LinkOrMuted
                          :href="text(item, 'MovieCastHref')"
                          :value="text(eventOf(item), 'MovieCast')"
                          new-window
                        />
                      </td>
                      <td class="sc-muted">
                        <span v-if="text(eventOf(item), 'Vessel')">{{ text(eventOf(item), "Vessel") }}</span
                        ><span v-else class="text-muted">-</span>
                      </td>
                      <td class="sc-muted sc-content-cell">
                        <span v-if="text(eventOf(item), 'Fg')">{{ text(eventOf(item), "Fg") }}</span
                        ><span v-else class="text-muted">-</span>
                      </td>
                    </tr>
                    <tr v-if="!items.length">
                      <td colspan="12"><div class="sc-empty-state">还没有可展示的 SC 事件。</div></td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </template>

        <template v-else-if="viewType === 'sc-event-cards'">
          <section class="cards-toolbar card-results-toolbar" aria-label="SC 事件卡片控制栏">
            <div class="cards-toolbar-sort" aria-label="排序方式">
              <div class="link-strip">
                <SharedSortLink
                  v-for="sort in cardSorts"
                  :key="sort.key"
                  :link="sortItem(sort.key)"
                  :label="sort.label"
                  class="strip-link"
                />
              </div>
            </div>
            <div class="next02-pink-cards-metrics" aria-label="SC 事件卡片摘要">
              <span :aria-label="`总数 ${total}`"
                ><strong>{{ total }}</strong></span
              >
              <span :aria-label="`当前页 ${items.length}`"
                ><strong>{{ items.length }}</strong></span
              >
            </div>
            <div class="cards-toolbar-filters">
              <ScViewSwitch />
              <a href="/databackfill" class="button-secondary" target="_blank" rel="noopener noreferrer">回填任务台</a>
            </div>
            <CardGridSizeControl />
            <div class="cards-toolbar-pagination">
              <PaginationBar :links="pageLinks" />
            </div>
          </section>

          <div v-if="items.length" class="sc-event-card-grid" :style="gridStyle">
            <div
              v-for="(item, index) in items"
              :key="eventName(eventOf(item))"
              class="sc-event-card-col"
              :style="gridItemStyle"
            >
              <article class="card h-100 sc-event-grid-card" :class="eventCardClass(item, index)">
                <a
                  :href="text(item, 'DetailHref') || eventDetailHref(eventName(eventOf(item)))"
                  class="sc-event-cover-link"
                  :title="eventName(eventOf(item))"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <div class="thumb-2x3 sc-event-cover">
                    <img
                      v-if="item.HasComeMovieCover"
                      :src="text(item, 'ComeMovieJacketImg')"
                      :alt="`${eventName(eventOf(item))} 封面`"
                      loading="lazy"
                      decoding="async"
                    />
                    <div v-else class="sc-event-cover-empty">
                      <span class="sc-event-cover-kicker">SC EVENT</span>
                      <strong
                        ><span class="sc-event-time"
                          ><span class="sc-event-date">{{ formatUnix(pick(eventOf(item), "ScTime")) }}</span
                          ><span class="sc-event-hour">{{ formatUnixHm(pick(eventOf(item), "ScTime")) }}</span></span
                        ></strong
                      >
                      <span>{{ eventName(eventOf(item)) }}</span>
                    </div>
                    <div class="sc-event-cover-badge">
                      {{ text(cardListItemOf(item), "ScMovieCount") || text(eventOf(item), "MovieNumber") || "0" }} SC
                    </div>
                  </div>
                </a>
                <div class="card-body sc-event-card-body">
                  <div class="sc-event-card-heading">
                    <a
                      :href="text(item, 'DetailHref') || eventDetailHref(eventName(eventOf(item)))"
                      class="sc-event-card-date-link"
                      :title="eventName(eventOf(item))"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <span class="sc-event-time">
                        <span class="sc-event-date">{{ formatUnix(pick(eventOf(item), "ScTime")) }}</span>
                        <span class="sc-event-hour">{{ formatUnixHm(pick(eventOf(item), "ScTime")) }}</span>
                      </span>
                    </a>
                  </div>
                  <div class="sc-event-card-primary">
                    <span class="sc-event-card-label">Come Movie</span>
                    <span class="sc-event-card-primary-link" :title="text(item, 'ComeMovieName')">
                      <LinkOrMuted
                        :href="text(item, 'ComeMovieHref')"
                        :value="text(item, 'ComeMovieName')"
                        new-window
                      />
                    </span>
                  </div>
                  <div class="sc-event-card-numbers" aria-label="影片数量">
                    <span
                      ><small>总片数</small
                      ><strong>{{ text(cardListItemOf(item), "TotalMovieCount") || "0" }}</strong></span
                    >
                    <span
                      ><small>SC 片数</small
                      ><strong>{{
                        text(cardListItemOf(item), "ScMovieCount") || text(eventOf(item), "MovieNumber") || "0"
                      }}</strong></span
                    >
                  </div>
                  <div class="sc-event-meta-grid">
                    <div class="sc-event-meta-row">
                      <span class="sc-event-meta-label">Cooldown</span>
                      <span class="sc-event-meta-value sc-event-meta-number">{{
                        num(eventOf(item), "Cooldown") ? humanDays(pick(eventOf(item), "Cooldown")) : "-"
                      }}</span>
                    </div>
                    <div class="sc-event-meta-row">
                      <span class="sc-event-meta-label">种类</span>
                      <span class="sc-event-meta-value" :title="text(eventOf(item), 'Kind')">{{
                        text(eventOf(item), "Kind") || "-"
                      }}</span>
                    </div>
                    <div class="sc-event-meta-row">
                      <span class="sc-event-meta-label">演员</span>
                      <span class="sc-event-meta-value" :title="text(eventOf(item), 'MovieCast')">
                        <LinkOrMuted
                          :href="text(cardListItemOf(item), 'MovieCastHref')"
                          :value="text(eventOf(item), 'MovieCast')"
                          new-window
                        />
                      </span>
                    </div>
                    <div class="sc-event-meta-row">
                      <span class="sc-event-meta-label">Vessel</span>
                      <span class="sc-event-meta-value" :title="text(eventOf(item), 'Vessel')">{{
                        text(eventOf(item), "Vessel") || "-"
                      }}</span>
                    </div>
                    <div class="sc-event-meta-row">
                      <span class="sc-event-meta-label">FG</span>
                      <span class="sc-event-meta-value" :title="text(eventOf(item), 'Fg')">{{
                        text(eventOf(item), "Fg") || "-"
                      }}</span>
                    </div>
                  </div>
                </div>
              </article>
            </div>
          </div>
          <div v-else class="sc-empty-state sc-event-card-empty">还没有可展示的 SC 事件。</div>
        </template>

        <template v-else-if="viewType === 'sc-event-detail'">
          <div class="next02-sc-detail-actionbar" aria-label="SC Event 操作栏">
            <span
              ><strong>{{ items.length }}</strong> 当前影片</span
            >
            <div class="next02-sc-detail-actions">
              <button type="button" class="button-primary" id="btnOpenScEventEditModal" @click="openEditModal">
                编辑信息
              </button>
              <button
                type="button"
                class="button-secondary danger"
                id="btnOpenScEventDeleteModal"
                :disabled="deleteSaving"
                @click="openDeleteModal"
              >
                删除事件
              </button>
              <a href="/scevents" class="button-secondary" target="_blank" rel="noopener noreferrer">返回列表</a>
              <a
                :href="`/scmove?sc_name=${encodeURIComponent(eventName(event))}`"
                class="button-secondary"
                target="_blank"
                rel="noopener noreferrer"
                >移动 SC Media</a
              >
              <a href="/databackfill" class="button-secondary" target="_blank" rel="noopener noreferrer">回填任务台</a>
            </div>
          </div>
          <div
            id="scEventPageMsg"
            class="next02-sc-detail-message"
            :class="{ 'text-danger': detailMessage && !detailMessageOK }"
            role="alert"
          >
            {{ detailMessage }}
          </div>
          <Teleport to="body">
            <div
              v-if="editModalOpen"
              class="modal-backdrop-v2 sc-event-edit-overlay"
              role="presentation"
              @click.self="closeEditModal"
            >
              <form
                id="scEventEditForm"
                class="modal-panel-v2 sc-event-edit-panel"
                role="dialog"
                aria-modal="true"
                aria-labelledby="scEventEditModalTitle"
                @submit.prevent="saveScEvent"
              >
                <div class="modal-header-v2">
                  <h2 id="scEventEditModalTitle">编辑 SC Event</h2>
                  <button
                    type="button"
                    class="modal-close-v2"
                    aria-label="关闭"
                    :disabled="editSaving"
                    @click="closeEditModal"
                  >
                    ×
                  </button>
                </div>
                <div class="modal-body-v2 sc-event-edit-body">
                  <div
                    v-if="detailMessage"
                    class="alert py-2 mb-0"
                    :class="detailMessageOK ? 'alert-success' : 'alert-danger'"
                    role="alert"
                  >
                    {{ detailMessage }}
                  </div>
                  <div class="sc-event-edit-grid">
                    <label class="sc-event-edit-field">
                      <span>Name</span>
                      <input id="scEventEditName" class="form-control" type="text" :value="eventName(event)" readonly />
                    </label>
                    <label class="sc-event-edit-field">
                      <span>Come Movie</span>
                      <select
                        id="scEventEditComeMovie"
                        class="form-select"
                        name="comeGListId"
                        :value="editForm.comeGListId"
                        @change="changeComeMovie"
                      >
                        <option :value="0">未选择</option>
                        <option
                          v-for="item in editComeMovieOptions"
                          :key="num(item, 'GListId')"
                          :value="num(item, 'GListId')"
                        >
                          {{ text(item, "MovieName") }} ({{ text(item, "MovieKey") }})
                        </option>
                      </select>
                    </label>
                    <label class="sc-event-edit-field">
                      <span>Lead Cast</span>
                      <select
                        id="scEventEditMovieCast"
                        v-model.number="editForm.movieCastPersonId"
                        class="form-select"
                        name="movieCastPersonId"
                        :disabled="editForm.comeGListId <= 0 || editComeMovieIsFC2"
                      >
                        <option :value="0">未选择</option>
                        <option
                          v-for="cast in editCastOptions"
                          :key="num(cast, 'personId', 'PersonID')"
                          :value="num(cast, 'personId', 'PersonID')"
                        >
                          {{ text(cast, "name", "Name") }}
                        </option>
                      </select>
                    </label>
                    <label class="sc-event-edit-field">
                      <span>Kind</span>
                      <select
                        id="scEventEditKind"
                        v-model.number="editForm.kindOptionId"
                        class="form-select"
                        name="kindOptionId"
                      >
                        <option :value="0">未选择</option>
                        <option v-for="option in editKindOptions" :key="option.id" :value="option.id">
                          {{ option.optionName }}
                        </option>
                      </select>
                    </label>
                    <label class="sc-event-edit-field">
                      <span>Duration</span>
                      <input
                        id="scEventEditDuration"
                        v-model="editForm.duration"
                        class="form-control"
                        name="duration"
                        type="number"
                      />
                    </label>
                    <label class="sc-event-edit-field">
                      <span>FG</span>
                      <select
                        id="scEventEditFg"
                        v-model.number="editForm.fgOptionId"
                        class="form-select"
                        name="fgOptionId"
                      >
                        <option :value="0">未选择</option>
                        <option v-for="option in editFGOptions" :key="option.id" :value="option.id">
                          {{ option.optionName }}
                        </option>
                      </select>
                    </label>
                    <label class="sc-event-edit-field">
                      <span>Vessel</span>
                      <select
                        id="scEventEditVessel"
                        v-model.number="editForm.vesselOptionId"
                        class="form-select"
                        name="vesselOptionId"
                      >
                        <option :value="0">未选择</option>
                        <option v-for="option in editVesselOptions" :key="option.id" :value="option.id">
                          {{ option.optionName }}
                        </option>
                      </select>
                    </label>
                    <label class="sc-event-edit-field is-wide">
                      <span>Remarks</span>
                      <textarea
                        id="scEventEditRemarks"
                        v-model="editForm.remarks"
                        class="form-control"
                        name="remarks"
                        rows="3"
                      ></textarea>
                    </label>
                    <div class="sc-event-edit-field is-wide">
                      <span>图片</span>
                      <div class="filter-actions sc-event-edit-image-actions">
                        <input
                          id="scEventEditImage"
                          ref="editImageInput"
                          type="file"
                          accept=".jpg,.jpeg,.png,.webp"
                          aria-label="SC 事件图片"
                          hidden
                          :disabled="editSaving"
                          @change="selectEditImage"
                        />
                        <button
                          type="button"
                          class="button-secondary"
                          :disabled="editSaving"
                          @click="editImageInput?.click()"
                        >
                          选择图片
                        </button>
                        <span class="sc-event-edit-image-name">{{
                          editImageFile?.name || text(payload, "EditImageName") || "尚无图片"
                        }}</span>
                        <button
                          v-if="editImageFile"
                          type="button"
                          class="button-secondary"
                          :disabled="editSaving"
                          @click="clearEditImage"
                        >
                          取消本次选图
                        </button>
                      </div>
                      <img
                        v-if="editImageSource && !editImageInvalid"
                        :src="editImageSource"
                        :alt="editImageFile ? '待保存图片预览' : '当前事件图片'"
                        class="next02-sc-detail-image sc-event-edit-image-preview"
                        @load="finishEditImagePreview"
                        @error="failEditImagePreview"
                      />
                    </div>
                  </div>
                </div>
                <div class="modal-footer-v2">
                  <button type="button" class="button-secondary" :disabled="editSaving" @click="closeEditModal">
                    取消
                  </button>
                  <button
                    type="submit"
                    class="button-primary"
                    :disabled="editSaving || editImageLoading || editImageInvalid"
                  >
                    {{ editSaving ? "保存中..." : "保存" }}
                  </button>
                </div>
              </form>
            </div>
          </Teleport>
          <Teleport to="body">
            <div
              v-if="deleteModalOpen"
              class="modal-backdrop-v2 sc-event-delete-overlay"
              role="presentation"
              @click.self="closeDeleteModal"
            >
              <section
                class="modal-panel-v2 sc-event-delete-panel"
                role="dialog"
                aria-modal="true"
                aria-labelledby="scEventDeleteModalTitle"
                aria-describedby="scEventDeleteDescription"
              >
                <div class="modal-header-v2">
                  <h2 id="scEventDeleteModalTitle">删除 SC Event</h2>
                  <button
                    type="button"
                    class="modal-close-v2"
                    aria-label="关闭"
                    :disabled="deleteSaving"
                    @click="closeDeleteModal"
                  >
                    ×
                  </button>
                </div>
                <div class="modal-body-v2">
                  <p id="scEventDeleteDescription" class="sc-event-delete-name">
                    确认删除事件“<strong>{{ eventName(event) }}</strong
                    >”吗？
                  </p>
                  <p class="sc-event-delete-warning">
                    将删除该事件及关联的 {{ text(payload, "TotalMovieCount") || "0" }} 条影片记录，并重新计算相关 SC
                    统计和下一场事件的 cooldown。SC Media 会继续保留，事件图片将移入删除归档目录。
                  </p>
                  <div v-if="deleteMessage" class="modal-message-v2" role="alert">{{ deleteMessage }}</div>
                </div>
                <div class="modal-footer-v2">
                  <button type="button" class="button-secondary" :disabled="deleteSaving" @click="closeDeleteModal">
                    取消
                  </button>
                  <button
                    type="button"
                    class="button-secondary danger sc-event-delete-confirm"
                    :disabled="deleteSaving"
                    @click="submitDeleteScEvent"
                  >
                    {{ deleteSaving ? "删除中..." : "确认删除" }}
                  </button>
                </div>
              </section>
            </div>
          </Teleport>

          <section class="next02-sc-detail-overview" aria-label="SC Event 概览">
            <div class="next02-sc-detail-panel next02-sc-detail-image-panel">
              <div class="next02-sc-detail-panel-head">
                <h2>SC Image</h2>
              </div>
              <div class="next02-sc-detail-image-wrap">
                <img
                  v-if="text(event, 'ImagePath')"
                  :src="text(event, 'ImagePath')"
                  :alt="`${eventName(event)} 图片`"
                  class="next02-sc-detail-image"
                />
                <div
                  v-else
                  class="next02-sc-detail-image-empty"
                  role="img"
                  :aria-label="`${eventName(event)} 暂无图片`"
                >
                  No Image
                </div>
              </div>
            </div>
            <div class="next02-sc-detail-panel next02-sc-detail-info-panel">
              <div class="next02-sc-detail-panel-head">
                <h2>事件信息</h2>
              </div>
              <div class="next02-sc-detail-info-grid">
                <MetricCard label="Sc Time" :value="formatUnixFull(pick(event, 'ScTime'))" date-time />
                <MetricCard label="总片数" :value="text(payload, 'TotalMovieCount') || '0'" />
                <MetricCard label="有 SC" :value="text(payload, 'ScMovieCount') || '0'" />
                <MetricCard label="无 SC" :value="text(payload, 'NoScMovieCount') || '0'" />
                <MetricCard label="JAV" :value="text(payload, 'JavMovieCount') || '0'" />
                <MetricCard label="FC2" :value="text(payload, 'FC2MovieCount') || '0'" />
                <MetricCard
                  label="Cooldown"
                  :value="num(event, 'Cooldown') > 0 ? humanDays(pick(event, 'Cooldown')) : '-'"
                />
                <MetricCard
                  label="Come Movie"
                  :value="text(event, 'ComeMovieName') || '-'"
                  :href="text(payload, 'ComeMovieHref')"
                />
                <MetricCard label="Come 类型" :value="movieSourceLabel(text(payload, 'ComeMovieType'))" />
                <MetricCard label="种类" :value="text(event, 'Kind') || '-'" />
                <MetricCard
                  label="Lead Cast"
                  :value="text(event, 'MovieCast') || '-'"
                  :href="text(event, 'MovieCast') ? `/listcard?cn=${encodeURIComponent(text(event, 'MovieCast'))}` : ''"
                />
                <MetricCard
                  label="时长(分钟)"
                  :value="num(event, 'DurationMinutes') > 0 ? `${text(event, 'DurationMinutes')} 分钟` : '-'"
                />
                <MetricCard label="FG" :value="text(event, 'Fg') || '-'" />
                <MetricCard label="Vessel" :value="text(event, 'Vessel') || '-'" />
                <div class="sc-detail-info-item sc-metric-card sc-detail-remarks">
                  <div class="sc-metric-label">Remarks</div>
                  <div class="sc-metric-value">
                    <span v-if="text(event, 'Remarks')">{{ text(event, "Remarks") }}</span
                    ><span v-else class="text-muted">暂无备注。</span>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section class="next02-sc-detail-movies" aria-labelledby="sc-detail-movies-title">
            <div class="next02-sc-detail-section-head">
              <div>
                <h2 id="sc-detail-movies-title">Movies</h2>
                <p>本次事件的影片及 SC 状态。</p>
              </div>
            </div>

            <div class="cards-toolbar card-results-toolbar" aria-label="电影筛选">
              <div class="cards-toolbar-sort"><span class="cards-toolbar-label">影片</span></div>
              <div class="next02-pink-cards-metrics" aria-label="电影摘要">
                <span :aria-label="`总数 ${text(payload, 'TotalMovieCount') || '0'}`"
                  ><strong>{{ text(payload, "TotalMovieCount") || "0" }}</strong></span
                >
                <span :aria-label="`当前页 ${items.length}`"
                  ><strong>{{ items.length }}</strong></span
                >
              </div>
              <div class="cards-toolbar-filters">
                <div class="link-strip">
                  <a
                    class="strip-link"
                    :href="text(payload, 'MovieFilterAllHref')"
                    :class="{ active: movieFilter === 'all' }"
                    >全部电影 {{ text(payload, "TotalMovieCount") || "0" }}</a
                  >
                  <a
                    class="strip-link"
                    :href="text(payload, 'MovieFilterScHref')"
                    :class="{ active: movieFilter === 'sc' }"
                    >有 SC {{ text(payload, "ScMovieCount") || "0" }}</a
                  >
                  <a
                    class="strip-link"
                    :href="text(payload, 'MovieFilterNoScHref')"
                    :class="{ active: movieFilter === 'nosc' }"
                    >无 SC {{ text(payload, "NoScMovieCount") || "0" }}</a
                  >
                  <a
                    class="strip-link"
                    :href="text(payload, 'MovieFilterJAVHref')"
                    :class="{ active: movieFilter === 'jav' }"
                    >JAV {{ text(payload, "JavMovieCount") || "0" }}</a
                  >
                  <a
                    class="strip-link"
                    :href="text(payload, 'MovieFilterFC2Href')"
                    :class="{ active: movieFilter === 'fc2' }"
                    >FC2 {{ text(payload, "FC2MovieCount") || "0" }}</a
                  >
                </div>
              </div>
              <CardGridSizeControl />
            </div>

            <section v-if="items.length" aria-label="SC Movies">
              <div
                class="movie-grid-v2 next02-pink-movie-grid next02-sc-detail-movie-grid"
                id="movie-grid"
                :style="gridStyle"
              >
                <div
                  v-for="(item, index) in items"
                  :key="movieKey(item, index)"
                  class="col movie-grid-col-v2"
                  :style="gridItemStyle"
                >
                  <component :is="eventMovieCardComponent(item)" :movie="eventMovieCardData(item)">
                    <template #actions>
                      <li
                        class="list-group-item movie-card-sc-status"
                        :class="eventMovieHasSc(item) ? 'has-sc-status' : 'no-sc-status'"
                        style="z-index: 2"
                      >
                        <div class="movie-card-sc-status-line">
                          <div class="movie-card-sc-badges">
                            <span
                              class="movie-card-source"
                              :class="eventMovieIsFC2(item) ? 'source-fc2' : 'source-jav'"
                              >{{ eventMovieIsFC2(item) ? "FC2" : "JAV" }}</span
                            >
                            <span class="movie-card-sc-state" :class="eventMovieHasSc(item) ? 'has-sc' : 'no-sc'">{{
                              eventMovieHasSc(item) ? "有 SC" : "无 SC"
                            }}</span>
                            <span v-if="eventMovieIsCome(item)" class="movie-card-sc-come">Come</span>
                          </div>
                          <div class="movie-card-sc-actions" role="group" aria-label="编辑影片 SC 状态">
                            <button
                              type="button"
                              class="js-sc-event-movie-status has-sc-action"
                              :class="{ active: eventMovieHasSc(item) }"
                              :data-is-sc="GLIST_IS_SC"
                              :disabled="eventMovieHasSc(item) || eventMovieStatusSaving(item)"
                              @click="updateMovieStatus(eventName(event), eventMovieGListID(item), GLIST_IS_SC)"
                            >
                              {{ eventMovieStatusSaving(item) ? "保存中" : "有 SC" }}
                            </button>
                            <button
                              type="button"
                              class="js-sc-event-movie-status no-sc-action"
                              :class="{ active: !eventMovieHasSc(item) }"
                              :data-is-sc="GLIST_IS_NOT_SC"
                              :disabled="!eventMovieHasSc(item) || eventMovieStatusSaving(item)"
                              @click="updateMovieStatus(eventName(event), eventMovieGListID(item), GLIST_IS_NOT_SC)"
                            >
                              {{ eventMovieStatusSaving(item) ? "保存中" : "无 SC" }}
                            </button>
                          </div>
                        </div>
                      </li>
                    </template>
                  </component>
                </div>
              </div>
            </section>
            <div v-else class="movie-empty next02-sc-detail-empty">{{ emptyDetailText }}</div>
          </section>

          <section v-if="failedItems.length" class="next02-sc-detail-failed" aria-labelledby="sc-detail-failed-title">
            <div class="next02-sc-detail-section-head">
              <div>
                <h2 id="sc-detail-failed-title">Skipped / Error Items</h2>
                <p>未能载入的影片条目。</p>
              </div>
            </div>
            <div class="next02-sc-detail-failed-table-wrap table-scroll">
              <table class="next02-sc-detail-failed-table data-table">
                <colgroup>
                  <col class="failed-movie" />
                  <col class="failed-javid" />
                  <col class="failed-error" />
                </colgroup>
                <thead>
                  <tr>
                    <th>Movie</th>
                    <th>JavId</th>
                    <th>Error</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(item, index) in failedItems" :key="index">
                    <td>{{ text(item, "MovieName") }}</td>
                    <td>{{ text(asRecord(item.Entry), "MovieJavId") || text(asRecord(item.FC2Movie), "FC2Key") }}</td>
                    <td>{{ text(item, "LoadError") }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </template>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onBeforeUnmount, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import {
  apiPathForRoute,
  deleteScEvent,
  getJSON,
  getScOptions,
  sendJSON,
  updateScEvent,
  type ScOptionGroups,
  type ScOptionItem
} from "@/features/sc/api";
import FC2MovieCard from "@/features/fc2/components/FC2MovieCard.vue";
import MovieCard from "@/features/movies/components/MovieCard.vue";
import PageShell from "@/shared/components/PageShell.vue";
import CardGridSizeControl from "@/shared/components/CardGridSizeControl.vue";
import { useCardGridSize } from "@/shared/composables/useCardGridSize";
import PaginationBar from "@/shared/components/PaginationBar.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import SharedSortLink from "@/shared/components/navigation/SortLink.vue";
import type { ScRecord } from "@/features/sc/types";
import { asArray, asRecord, normalizeHref, num, pick, text } from "@/shared/utils/format";

const GLIST_IS_NOT_SC = 1;
const GLIST_IS_SC = 2;
const cardSorts = [
  { key: "ByTime", label: "时间" },
  { key: "ByMovieCnt", label: "片数" },
  { key: "ByComeMovie", label: "Come 片" },
  { key: "ByCooldown", label: "Cooldown" },
  { key: "ByKind", label: "种类" },
  { key: "ByMovieCast", label: "演员" },
  { key: "ByVessel", label: "Vessel" },
  { key: "ByFg", label: "FG" }
];

const route = useRoute();
const { gridStyle, gridItemStyle } = useCardGridSize();
const payload = ref<ScRecord>({});
const cardListItems = ref<ScRecord[]>([]);
const loading = ref(false);
const error = ref("");
const detailMessage = ref("");
const detailMessageOK = ref(true);
const editSaving = ref(false);
const editModalOpen = ref(false);
const editImageInput = ref<HTMLInputElement | null>(null);
const editImageFile = ref<File | null>(null);
const editImagePreviewUrl = ref("");
const editImageLoading = ref(false);
const editImageInvalid = ref(false);
const deleteSaving = ref(false);
const deleteModalOpen = ref(false);
const deleteMessage = ref("");
const movieStatusSavingKey = ref("");
const scOptions = ref<ScOptionGroups>({ kind: [], fg: [], vessel: [] });
const editForm = reactive({
  comeGListId: 0,
  movieCastPersonId: 0,
  kindOptionId: 0,
  duration: "0",
  fgOptionId: 0,
  vesselOptionId: 0,
  remarks: ""
});

watch(() => route.fullPath, load, { immediate: true });
onBeforeUnmount(clearEditImage);

async function load() {
  editModalOpen.value = false;
  clearEditImage();
  loading.value = true;
  error.value = "";
  cardListItems.value = [];
  try {
    if (route.path === "/sccards") {
      const [pageData, listData, optionData] = await Promise.all([
        getJSON(apiPathForRoute(route)),
        getJSON(cardListSummaryApiPath()).catch(() => ({})),
        getScOptions(true)
      ]);
      payload.value = asRecord(pageData);
      cardListItems.value = asArray<ScRecord>(pick(asRecord(listData), "Items"));
      scOptions.value = optionData;
    } else if (route.name === "sc-events") {
      const [pageData, optionData] = await Promise.all([getJSON(apiPathForRoute(route)), getScOptions(true)]);
      payload.value = asRecord(pageData);
      scOptions.value = optionData;
    } else if (route.name === "sc-event-detail") {
      const [pageData, optionData] = await Promise.all([getJSON(apiPathForRoute(route)), getScOptions()]);
      payload.value = asRecord(pageData);
      scOptions.value = optionData;
    } else {
      payload.value = asRecord(await getJSON(apiPathForRoute(route)));
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title", "PageTitle") || "SC Events");
const pageNote = computed(
  () =>
    text(payload.value, "PageNote") ||
    (viewType.value === "sc-event-detail" ? "基础信息、影片清单与快速操作入口。" : "")
);
const viewType = computed(() => text(payload.value, "ViewType"));
const items = computed(() => asArray<ScRecord>(pick(payload.value, "Items")));
const failedItems = computed(() => asArray<ScRecord>(pick(payload.value, "FailedItems")));
const total = computed(() => text(payload.value, "Total") || String(items.value.length));
const pageInfo = computed(() => asRecord(pick(payload.value, "PageInfo", "pageInfo")));
const pageLinks = computed(() => asArray<ScRecord>(pageInfo.value.Links || pageInfo.value.links));
const event = computed(() => asRecord(pick(payload.value, "Event")));
const editImageSource = computed(() => editImagePreviewUrl.value || text(event.value, "ImagePath"));
const editComeMovieOptions = computed(() => asArray<ScRecord>(pick(payload.value, "EditComeMovieOptions")));
const selectedEditComeMovie = computed(() =>
  editComeMovieOptions.value.find((item) => num(item, "GListId") === editForm.comeGListId)
);
const editComeMovieIsFC2 = computed(() => text(selectedEditComeMovie.value || {}, "MovieType") === "fc2ppv");
const editCastOptions = computed(() => {
  const selected = selectedEditComeMovie.value;
  return selected ? asArray<ScRecord>(pick(selected, "CastOptions")) : [];
});
const editKindOptions = computed(() =>
  withCurrentScOption(scOptions.value.kind, editForm.kindOptionId, text(event.value, "Kind"), "kind")
);
const editFGOptions = computed(() =>
  withCurrentScOption(scOptions.value.fg, editForm.fgOptionId, text(event.value, "Fg"), "fg")
);
const editVesselOptions = computed(() =>
  withCurrentScOption(scOptions.value.vessel, editForm.vesselOptionId, text(event.value, "Vessel"), "vessel")
);
const movieFilter = computed(() => text(payload.value, "MovieFilter") || "sc");
const sortQuery = computed(() => asRecord(pick(payload.value, "ScSortQuery")));
const scViewQuery = computed(() => asRecord(pick(payload.value, "ScViewQuery")));
const cardListItemsByEventName = computed(() => {
  const index = new Map<string, ScRecord>();
  for (const item of cardListItems.value) {
    const name = eventName(eventOf(item));
    if (name) index.set(name, item);
  }
  return index;
});
const filterPageSize = computed(() => {
  return routeQueryValue("ps") || "24";
});
const emptyDetailText = computed(() => {
  if (movieFilter.value === "all") return "本次 SC 还没有电影明细。";
  if (movieFilter.value === "nosc") return "当前筛选下没有 is_sc=1 的电影。";
  if (movieFilter.value === "jav") return "当前事件没有 JAV 影片。";
  if (movieFilter.value === "fc2") return "当前事件没有 FC2 影片。";
  return "当前筛选下没有 is_sc=2 的电影。";
});

watch(event, syncEditForm, { immediate: true });

function eventOf(item: ScRecord): ScRecord {
  return asRecord(pick(item, "Event"));
}

function movieSourceLabel(movieType: string): string {
  if (movieType === "fc2ppv" || movieType === "fc2") return "FC2";
  if (movieType === "jav") return "JAV";
  return "-";
}

function movieSourceClass(movieType: string): string {
  if (movieType === "fc2ppv" || movieType === "fc2") return "source-fc2";
  if (movieType === "jav") return "source-jav";
  return "source-unknown";
}

function cardListSummaryApiPath(): string {
  const query = route.fullPath.split("?", 2)[1] || "";
  const params = new URLSearchParams(query);
  if (!params.has("ps")) params.set("ps", "18");
  const encoded = params.toString();
  return `/api/sc-events${encoded ? `?${encoded}` : ""}`;
}

function cardListItemOf(item: ScRecord): ScRecord {
  return cardListItemsByEventName.value.get(eventName(eventOf(item))) || {};
}

function routeQueryValue(key: string): string {
  const value = route.query[key];
  if (Array.isArray(value)) return value[0] ? String(value[0]) : "";
  return value == null ? "" : String(value);
}

function eventName(row: ScRecord): string {
  return text(row, "Name", "name");
}

function inputValue(event: Event): string {
  return (event.target as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement).value;
}

function withCurrentScOption(
  options: ScOptionItem[],
  currentID: number,
  currentName: string,
  optionType: ScOptionItem["optionType"]
): ScOptionItem[] {
  if (currentID <= 0 || options.some((option) => option.id === currentID)) return options;
  return [
    {
      id: currentID,
      optionType,
      optionName: `${currentName || `选项 #${currentID}`}（已停用）`,
      sortOrder: 0,
      isEnabled: 0,
      createdOn: 0,
      updatedOn: 0
    },
    ...options
  ];
}

function syncEditForm(): void {
  editForm.comeGListId = num(payload.value, "EditComeGListId");
  editForm.movieCastPersonId = num(event.value, "MovieCastPersonID", "movieCastPersonId");
  editForm.kindOptionId = num(event.value, "KindOptionID", "kindOptionId");
  editForm.duration = text(event.value, "DurationMinutes", "Duration") || "0";
  editForm.fgOptionId = num(event.value, "FGOptionID", "fgOptionId");
  editForm.vesselOptionId = num(event.value, "VesselOptionID", "vesselOptionId");
  editForm.remarks = text(event.value, "Remarks");
}

function openEditModal(): void {
  syncEditForm();
  clearEditImage();
  detailMessage.value = "";
  editModalOpen.value = true;
}

function closeEditModal(): void {
  if (editSaving.value) return;
  editModalOpen.value = false;
  clearEditImage();
}

function clearEditImage(): void {
  if (editImagePreviewUrl.value) URL.revokeObjectURL(editImagePreviewUrl.value);
  editImagePreviewUrl.value = "";
  editImageFile.value = null;
  editImageLoading.value = false;
  editImageInvalid.value = false;
  if (editImageInput.value) editImageInput.value.value = "";
}

function selectEditImage(inputEvent: Event): void {
  const file = (inputEvent.target as HTMLInputElement).files?.[0];
  if (!file) return;
  clearEditImage();
  detailMessage.value = "";
  detailMessageOK.value = true;
  editImageFile.value = file;
  if (!/\.(jpe?g|png|webp)$/i.test(file.name)) {
    editImageInvalid.value = true;
    detailMessageOK.value = false;
    detailMessage.value = "请选择 JPG、PNG 或 WEBP 图片";
    return;
  }
  editImageLoading.value = true;
  editImagePreviewUrl.value = URL.createObjectURL(file);
}

function finishEditImagePreview(imageEvent: Event): void {
  if ((imageEvent.target as HTMLImageElement).currentSrc === editImagePreviewUrl.value) {
    editImageLoading.value = false;
  }
}

function failEditImagePreview(imageEvent: Event): void {
  if (!editImageFile.value || (imageEvent.target as HTMLImageElement).currentSrc !== editImagePreviewUrl.value) return;
  editImageLoading.value = false;
  editImageInvalid.value = true;
  detailMessageOK.value = false;
  detailMessage.value = "图片无法读取，请重新选择";
}

function openDeleteModal(): void {
  deleteMessage.value = "";
  deleteModalOpen.value = true;
}

function closeDeleteModal(): void {
  if (deleteSaving.value) return;
  deleteModalOpen.value = false;
  deleteMessage.value = "";
}

function changeComeMovie(event: Event): void {
  editForm.comeGListId = Number(inputValue(event) || 0);
  const selected = editComeMovieOptions.value.find((item) => num(item, "GListId") === editForm.comeGListId);
  const castOptions = selected ? asArray<ScRecord>(pick(selected, "CastOptions")) : [];
  if (!castOptions.some((cast) => num(cast, "personId", "PersonID") === editForm.movieCastPersonId)) {
    editForm.movieCastPersonId = 0;
  }
}

async function saveScEvent(eventObject: Event): Promise<void> {
  eventObject.preventDefault();
  if (editSaving.value || editImageLoading.value || editImageInvalid.value) return;
  editSaving.value = true;
  detailMessage.value = "";
  try {
    const name = eventName(event.value);
    const result = asRecord(
      await updateScEvent(
        name,
        {
          comeGListId: editForm.comeGListId,
          movieCastPersonId: editForm.movieCastPersonId,
          kindOptionId: editForm.kindOptionId,
          duration: Number(editForm.duration || 0),
          fgOptionId: editForm.fgOptionId,
          vesselOptionId: editForm.vesselOptionId,
          remarks: editForm.remarks
        },
        editImageFile.value
      )
    );
    detailMessageOK.value = true;
    detailMessage.value = text(result, "message") || "SC Event 信息更新成功";
    payload.value = {
      ...payload.value,
      Event: asRecord(pick(result, "item")),
      EditImageName:
        text(asRecord(pick(result, "item")), "ImagePath")
          .split(/[\\/]/)
          .pop() || "",
      EditComeGListId: editForm.comeGListId
    };
    editModalOpen.value = false;
    clearEditImage();
  } catch (err) {
    detailMessageOK.value = false;
    detailMessage.value = err instanceof Error ? err.message : String(err);
  } finally {
    editSaving.value = false;
  }
}

async function submitDeleteScEvent(): Promise<void> {
  if (deleteSaving.value) return;
  const name = eventName(event.value);
  if (!name) {
    deleteMessage.value = "缺少 SC 事件名";
    return;
  }

  deleteSaving.value = true;
  deleteMessage.value = "";
  try {
    const result = asRecord(await deleteScEvent(name));
    if (pick(result, "deleted") !== true) throw new Error(text(result, "error") || "接口未确认删除结果");
    window.location.href = "/scevents";
  } catch (err) {
    deleteMessage.value = err instanceof Error ? err.message : String(err);
  } finally {
    deleteSaving.value = false;
  }
}

async function updateMovieStatus(scName: string, gListID: number, isSc: number): Promise<void> {
  const savingKey = `${scName}:${gListID}`;
  if (!scName || gListID <= 0 || movieStatusSavingKey.value) return;
  movieStatusSavingKey.value = savingKey;
  detailMessage.value = "";
  try {
    const result = asRecord(
      await sendJSON(`/api/sc-events/${encodeURIComponent(scName)}/movies/${gListID}/status`, "POST", { isSc })
    );
    detailMessageOK.value = true;
    detailMessage.value = text(result, "message") || "SC Event 影片状态更新成功";
    await load();
  } catch (err) {
    detailMessageOK.value = false;
    detailMessage.value = err instanceof Error ? err.message : String(err);
  } finally {
    movieStatusSavingKey.value = "";
  }
}

function eventDetailHref(name: string): string {
  return `/scevent/${encodeURI(name || "")}`;
}

function sortItem(key: string): ScRecord {
  return asRecord(sortQuery.value[key]);
}

function formatUnix(input: unknown, withTime = false): string {
  const n = Number(input || 0);
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

function formatUnixHm(input: unknown): string {
  const n = Number(input || 0);
  if (!n) return "-";
  const d = new Date(n * 1000);
  return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}

function formatUnixFull(input: unknown): string {
  const n = Number(input || 0);
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

function humanDays(input: unknown): string {
  const n = Number(input || 0);
  if (n <= 0) return "-";
  return `${(n / 86400).toFixed(2)} 天`;
}

function eventMonthKey(input: unknown): string {
  const n = Number(input || 0);
  if (n <= 0) return "";
  const date = new Date(n * 1000);
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`;
}

function eventRunningMonthClass(input: unknown): string {
  const month = eventMonthKey(input);
  if (!month) return "sc-running-month-neutral";
  const monthNumber = Number(month.slice(5, 7));
  return `sc-running-month-${(monthNumber % 3) + 1}`;
}

function eventRunningTableMonthClass(input: unknown): string {
  const month = eventMonthKey(input);
  if (!month) return "sc-running-month-neutral";
  const monthNumber = Number(month.slice(5, 7));
  return monthNumber % 2 === 0 ? "sc-running-month-light" : "sc-running-month-deep";
}

function isEventRunningMonthStart(item: ScRecord, index: number): boolean {
  if (index <= 0) return false;
  const currentMonth = eventMonthKey(pick(eventOf(item), "ScTime"));
  const previousMonth = eventMonthKey(pick(eventOf(items.value[index - 1]), "ScTime"));
  return Boolean(currentMonth && previousMonth && currentMonth !== previousMonth);
}

function eventRunningRowClass(item: ScRecord, index: number) {
  return [
    "next02-sc-event-row",
    eventRunningTableMonthClass(pick(eventOf(item), "ScTime")),
    { "sc-running-month-start": isEventRunningMonthStart(item, index) }
  ];
}

function eventCardClass(item: ScRecord, index: number) {
  return [
    eventRunningMonthClass(pick(eventOf(item), "ScTime")),
    { "sc-running-month-start": isEventRunningMonthStart(item, index) }
  ];
}

function movieKey(item: ScRecord, index: number): string {
  const entry = asRecord(pick(item, "Entry"));
  const gListID = num(entry, "Id");
  if (gListID > 0) return `g-list-${gListID}`;
  const fc2Movie = asRecord(pick(item, "FC2Movie"));
  if (num(fc2Movie, "ID") > 0) return `fc2-${num(fc2Movie, "ID")}`;
  const movie = asRecord(pick(item, "MovieType"));
  return text(movie, "name", "Name", "jav_id", "JavID") || String(index);
}

function eventMovieEntry(item: ScRecord): ScRecord {
  return asRecord(pick(item, "Entry"));
}

function eventMovieIsFC2(item: ScRecord): boolean {
  return text(eventMovieEntry(item), "MovieType") === "fc2ppv" || num(eventMovieEntry(item), "FC2MovieId") > 0;
}

function eventMovieCardComponent(item: ScRecord) {
  return eventMovieIsFC2(item) ? FC2MovieCard : MovieCard;
}

function eventMovieCardData(item: ScRecord): ScRecord {
  const card = asRecord(pick(item, "Card"));
  if (Object.keys(card).length) return card;
  return eventMovieIsFC2(item) ? asRecord(pick(item, "FC2Movie")) : asRecord(pick(item, "MovieType"));
}

function eventMovieGListID(item: ScRecord): number {
  return num(eventMovieEntry(item), "Id");
}

function eventMovieHasSc(item: ScRecord): boolean {
  return num(eventMovieEntry(item), "IsSc") === GLIST_IS_SC;
}

function eventMovieIsCome(item: ScRecord): boolean {
  return num(eventMovieEntry(item), "IsCome") === 2;
}

function eventMovieStatusSaving(item: ScRecord): boolean {
  return movieStatusSavingKey.value === `${eventName(event.value)}:${eventMovieGListID(item)}`;
}

const LinkOrMuted = defineComponent({
  name: "ScLinkOrMuted",
  props: {
    href: { type: String, default: "" },
    value: { type: String, default: "" },
    newWindow: { type: Boolean, default: false }
  },
  setup(props) {
    return () =>
      props.value
        ? props.href
          ? h(
              "a",
              {
                href: props.href,
                target: props.newWindow ? "_blank" : undefined,
                rel: props.newWindow ? "noopener noreferrer" : undefined
              },
              props.value
            )
          : h("span", props.value)
        : h("span", { class: "text-muted" }, "-");
  }
});

const ScViewSwitch = defineComponent({
  name: "ScViewSwitch",
  setup() {
    return () =>
      h("div", { class: "btn-group", role: "group", "aria-label": "SC event view switch" }, [
        h(
          "a",
          {
            href: normalizeHref(text(scViewQuery.value, "ListHref"), "/scevents"),
            class: ["btn", "btn-outline-primary", { "active disabled": Boolean(scViewQuery.value.IsList) }],
            target: "_blank",
            rel: "noopener noreferrer"
          },
          "列表视图"
        ),
        h(
          "a",
          {
            href: normalizeHref(text(scViewQuery.value, "CardHref"), "/sccards"),
            class: ["btn", "btn-outline-primary", { "active disabled": Boolean(scViewQuery.value.IsCard) }],
            target: "_blank",
            rel: "noopener noreferrer"
          },
          "卡片视图"
        )
      ]);
  }
});

const MetricCard = defineComponent({
  name: "ScMetricCard",
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
    href: { type: String, default: "" },
    dateTime: { type: Boolean, default: false }
  },
  setup(props) {
    return () =>
      h("div", { class: "sc-detail-info-item sc-metric-card" }, [
        h("div", { class: "sc-metric-label" }, props.label),
        h(
          "div",
          { class: "sc-metric-value" },
          props.dateTime
            ? [h(ScDateTime, { value: props.value })]
            : props.href && props.value !== "-"
              ? [h("a", { href: props.href, target: "_blank", rel: "noopener noreferrer" }, props.value)]
              : props.value
        )
      ]);
  }
});
</script>
