<template>
  <PageShell class="movie-detail-shell" :title="pageTitle" :note="subtitle" eyebrow="Movie Detail">
    <LoadingState v-if="loading" />
    <ErrorPanel v-else-if="error" :message="error" />
    <template v-else>
      <div
        id="fetchSiteListPage"
        class="movie-detail-v2 movie-detail-page movie-detail-original-page"
        :data-movie-jav-id="javID"
        :data-movie-name="movieName"
      >
        <div class="movie-detail-top-shell">
          <div class="movie-detail-titlebar">
            <div class="movie-detail-title-main">
              <h4 class="mb-0 movie-detail-title" :title="titleText">{{ titleText || movieName }}</h4>
            </div>
            <div class="movie-detail-title-meta">
              <span v-if="movieName" class="movie-title-meta-chip movie-title-meta-code">{{ movieName }}</span>
              <span v-if="releaseDate" class="movie-title-meta-chip movie-title-meta-date">上映 {{ releaseDate }}</span>
              <span v-if="score" class="movie-title-meta-chip movie-title-meta-score">{{ score }}</span>
              <span v-if="castSummary" class="movie-title-meta-chip movie-title-meta-cast" :title="castSummary">{{
                castSummary
              }}</span>
            </div>
          </div>

          <div class="row movie-original-top-row">
            <div class="col-lg-7 mb-3 mb-lg-0">
              <div class="card shadow-sm movie-section-card movie-poster-card">
                <div class="thumb-2x2 movie-poster-frame">
                  <img
                    v-if="poster"
                    :src="poster"
                    alt="Card image cap"
                    class="card-img-bottom img-fluid movie-poster-image"
                  />
                  <div v-else class="card-img-bottom img-fluid detail-poster-empty">No Image</div>
                  <div class="badge-stack">
                    <span v-if="comeTimes" class="badge-come">❤ {{ comeTimes }}</span>
                    <span v-if="scTimes" class="badge-times">× {{ scTimes }}</span>
                  </div>
                  <span v-if="score" class="badge-score">★ {{ score }}</span>
                </div>
              </div>
            </div>

            <div class="col-lg-5">
              <div class="card shadow-sm movie-section-card movie-info-card">
                <div class="card-body">
                  <ul class="nav nav-tabs nav-tabs-custom nav-justified movie-detail-tabs" role="tablist">
                    <li class="nav-item" role="presentation">
                      <button
                        type="button"
                        class="nav-link movie-detail-tab"
                        :class="{ active: activeDetailTab === 'info' }"
                        role="tab"
                        @click="activeDetailTab = 'info'"
                      >
                        Info
                      </button>
                    </li>
                    <li class="nav-item" role="presentation">
                      <button
                        type="button"
                        class="nav-link movie-detail-tab"
                        :class="{ active: activeDetailTab === 'rank' }"
                        role="tab"
                        @click="activeDetailTab = 'rank'"
                      >
                        Rank
                      </button>
                    </li>
                  </ul>

                  <div class="tab-content movie-detail-tab-content">
                    <div
                      v-show="activeDetailTab === 'info'"
                      class="tab-pane p-3 active movie-info-pane"
                      role="tabpanel"
                    >
                      <table class="table table-bordered dt-responsive nowrap movie-info-table" style="width: 100%">
                        <tbody>
                          <tr class="table-light movie-info-row movie-info-row-code">
                            <th>识别码</th>
                            <td>
                              <div class="movie-info-code-line">
                                <a
                                  :href="`/listcard?pn=${encodeURIComponent(prefix)}`"
                                  class="movie-info-code movie-tag-code"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >{{ movieName }}</a
                                >
                                <a
                                  v-if="mediaPlayable"
                                  :href="videoUrl"
                                  class="btn btn-sm movie-action-button movie-action-button-play"
                                  :class="mediaButtonClass"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >Play</a
                                >
                                <span
                                  v-else-if="mediaLabel"
                                  class="btn btn-sm btn-secondary disabled movie-action-button movie-action-button-play"
                                  aria-disabled="true"
                                  >{{ mediaLabel }}</span
                                >
                                <small v-if="encodeName" class="movie-info-subtle">({{ encodeName }})</small>
                              </div>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>发行时间</th>
                            <td>
                              <span class="movie-info-value movie-tag-date movie-release-date">{{
                                releaseDate ? `上映 ${releaseDate}` : "-"
                              }}</span>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>长度</th>
                            <td>
                              <div class="movie-info-action-line">
                                <span class="movie-info-value">{{ lengthText || "-" }}</span>
                                <button
                                  id="btnDownloadCover"
                                  type="button"
                                  class="btn btn-sm btn-outline-info movie-action-button"
                                  :data-jav="javID"
                                  :disabled="coverBusy"
                                  @click="downloadCoverNow"
                                >
                                  <span>{{ coverBusy ? "下载中" : "下载封面" }}</span>
                                </button>
                              </div>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>制作商</th>
                            <td>
                              <div class="movie-info-tag-group">
                                <a
                                  v-if="maker"
                                  class="tag movie-tag-company"
                                  :href="`/listcard?mn=${encodeURIComponent(maker)}`"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >{{ maker }}</a
                                ><span v-else class="movie-info-empty">-</span>
                              </div>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>发行商</th>
                            <td>
                              <div class="movie-info-tag-group">
                                <a
                                  v-if="label"
                                  class="tag movie-tag-company"
                                  :href="`/listcard?ln=${encodeURIComponent(label)}`"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >{{ label }}</a
                                ><span v-else class="movie-info-empty">-</span>
                              </div>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>导演</th>
                            <td>
                              <div class="movie-info-tag-group">
                                <a
                                  v-if="director"
                                  class="tag movie-tag-person movie-tag-director"
                                  :href="`/listcard?dn=${encodeURIComponent(director)}`"
                                  :title="director"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >{{ director }}</a
                                ><span v-else class="movie-info-empty">-</span>
                              </div>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>分类</th>
                            <td>
                              <div class="movie-info-tag-group">
                                <a
                                  v-for="genre in genres"
                                  :key="genre"
                                  class="tag movie-tag-genre"
                                  :href="`/listcard?gn=${encodeURIComponent(genre)}`"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >{{ genre }}</a
                                >
                                <span v-if="!genres.length" class="movie-info-empty">-</span>
                              </div>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>演员</th>
                            <td>
                              <div class="movie-info-tag-group movie-info-cast-line">
                                <a
                                  v-for="cast in casts"
                                  :key="cast.href + cast.label"
                                  class="tag movie-tag-person movie-tag-cast"
                                  :href="cast.href"
                                  :title="cast.label"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >{{ cast.label }}</a
                                >
                                <button
                                  type="button"
                                  class="btn btn-sm btn-outline-primary movie-inline-action"
                                  data-bs-toggle="modal"
                                  data-bs-target="#addMovieCastModal"
                                  :data-jav="javID"
                                  @click="openCastModal"
                                >
                                  添加 Cast
                                </button>
                              </div>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>评分</th>
                            <td>
                              <div class="movie-info-score-line">
                                <span class="movie-info-score movie-tag-score">{{ score || "-" }}</span>
                                <span class="movie-info-pill-soft movie-tag-count">{{ watched || "-" }}</span>
                                <span class="movie-info-date movie-tag-date movie-update-date">{{
                                  updateDate || "-"
                                }}</span>
                                <button
                                  id="btn-downloadlater"
                                  type="button"
                                  class="btn btn-sm movie-action-button movie-download-later-button"
                                  :class="downloadLaterActive ? 'btn-warning' : 'btn-outline-warning'"
                                  :data-name="javID"
                                  :data-needdownload="needDownload"
                                  :title="downloadLaterTitle"
                                  :disabled="downloadLaterBusy"
                                  @click="toggleDownloadLater"
                                >
                                  <span>{{ downloadLaterButtonText }}</span>
                                </button>
                              </div>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>链接</th>
                            <td class="movie-link-strip">
                              <div class="movie-info-tag-group movie-info-links">
                                <a
                                  v-if="busUrl"
                                  class="movie-tag-link"
                                  :href="busUrl"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >Bus</a
                                >
                                <a
                                  v-if="searchUrl"
                                  class="movie-tag-link"
                                  :href="searchUrl"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >Suk</a
                                >
                                <a
                                  v-if="javUrl"
                                  class="movie-tag-link"
                                  :href="javUrl"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >Lib</a
                                >
                                <span v-if="!busUrl && !searchUrl && !javUrl" class="movie-info-empty">-</span>
                              </div>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>RankInfo</th>
                            <td>
                              <span class="movie-info-value">{{ firstRankingDay || "-" }}</span>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>Sc</th>
                            <td>
                              <div v-if="scRows.length" class="movie-info-tag-group">
                                <a
                                  v-for="row in scRows"
                                  :key="row.href + row.name"
                                  class="tag movie-tag-event"
                                  :href="row.href"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >{{ row.name }}</a
                                >
                              </div>
                              <span v-else class="movie-info-empty">-</span>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>电影相册</th>
                            <td>
                              <div class="movie-info-album-cell">
                                <button
                                  type="button"
                                  class="btn btn-sm btn-outline-primary movie-album-primary-action"
                                  id="btnOpenMovieAlbumModal"
                                  @click="openMovieAlbumModal"
                                >
                                  加入电影相册
                                </button>
                                <a
                                  v-for="album in selectedAlbums"
                                  :key="album.name"
                                  class="tag movie-album-badge movie-tag-album"
                                  :href="`/moviealbumlist?album_name=${encodeURIComponent(album.name)}`"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                >
                                  {{ album.name }}
                                </a>
                                <span v-if="!selectedAlbums.length" class="movie-info-empty">-</span>
                              </div>
                            </td>
                          </tr>
                        </tbody>
                      </table>

                      <div
                        v-if="downloadLaterMessage || downloadLaterError || coverMessage || coverError"
                        class="movie-action-feedback"
                      >
                        <span v-if="downloadLaterMessage" class="movie-action-feedback-ok">{{
                          downloadLaterMessage
                        }}</span>
                        <span v-if="downloadLaterError" class="movie-action-feedback-error">{{
                          downloadLaterError
                        }}</span>
                        <span v-if="coverMessage" class="movie-action-feedback-ok">{{ coverMessage }}</span>
                        <span v-if="coverError" class="movie-action-feedback-error">{{ coverError }}</span>
                      </div>
                    </div>

                    <div
                      v-show="activeDetailTab === 'rank'"
                      class="tab-pane p-3 active movie-rank-panel"
                      role="tabpanel"
                    >
                      <div v-if="rankRows.length" class="movie-rank-toolbar">
                        <label class="movie-rank-size">
                          <span>每页</span>
                          <input
                            v-model.number="rankPageSize"
                            class="form-control form-control-sm"
                            type="number"
                            step="1"
                            @change="rankPage = 1"
                          />
                        </label>
                        <div class="movie-rank-pager">
                          <button
                            type="button"
                            class="btn btn-sm btn-outline-secondary"
                            :disabled="rankPage <= 1"
                            @click="rankPage -= 1"
                          >
                            上一页
                          </button>
                          <span>{{ rankPage }} / {{ rankTotalPages }}</span>
                          <button
                            type="button"
                            class="btn btn-sm btn-outline-secondary"
                            :disabled="rankPage >= rankTotalPages"
                            @click="rankPage += 1"
                          >
                            下一页
                          </button>
                        </div>
                      </div>
                      <div v-if="rankRows.length" class="table-responsive">
                        <table
                          id="moviedetail_rankinfo"
                          class="table table-bordered dt-responsive nowrap mb-0 movie-rank-table"
                          style="width: 100%"
                        >
                          <thead>
                            <tr class="table-light">
                              <th
                                :aria-sort="
                                  rankSortKey === 'date'
                                    ? rankSortOrder === 'asc'
                                      ? 'ascending'
                                      : 'descending'
                                    : 'none'
                                "
                              >
                                <button
                                  type="button"
                                  class="movie-rank-sort-button"
                                  data-rank-sort-key="date"
                                  :class="{ active: rankSortKey === 'date' }"
                                  @click="sortRankTable('date')"
                                >
                                  <span>Date</span>
                                  <span class="movie-rank-sort-mark">{{ rankSortMark("date") }}</span>
                                </button>
                              </th>
                              <th
                                :aria-sort="
                                  rankSortKey === 'rank'
                                    ? rankSortOrder === 'asc'
                                      ? 'ascending'
                                      : 'descending'
                                    : 'none'
                                "
                              >
                                <button
                                  type="button"
                                  class="movie-rank-sort-button"
                                  data-rank-sort-key="rank"
                                  :class="{ active: rankSortKey === 'rank' }"
                                  @click="sortRankTable('rank')"
                                >
                                  <span>Pts</span>
                                  <span class="movie-rank-sort-mark">{{ rankSortMark("rank") }}</span>
                                </button>
                              </th>
                            </tr>
                          </thead>
                          <tbody>
                            <tr v-for="row in pagedRankRows" :key="row.date">
                              <td>
                                <a
                                  :href="`/listcarddayrank?date=${encodeURIComponent(row.date)}`"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >{{ row.date }}</a
                                >
                              </td>
                              <td>
                                <span>{{ row.rank }}</span>
                              </td>
                            </tr>
                          </tbody>
                        </table>
                      </div>
                      <div v-else class="text-muted">暂无 Rank 数据</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="showFilmInfo" class="row mt-3 movie-film-row">
          <div class="col-12">
            <div class="card shadow-sm movie-section-card movie-film-card">
              <div class="card-body">
                <div class="movie-film-panel">
                  <div class="movie-film-head">
                    <h5 class="card-title movie-section-title">视频信息</h5>
                    <span class="movie-film-source">WMedia</span>
                  </div>
                  <div class="movie-film-grid">
                    <div class="movie-film-item">
                      <span class="movie-film-label">Source</span>
                      <strong class="movie-film-value">WMedia</strong>
                    </div>
                    <div class="movie-film-item movie-film-item-wide movie-film-hash">
                      <span class="movie-film-label">TorrentHash</span>
                      <div class="movie-film-value-line">
                        <strong class="text-monospace movie-film-value" :title="mediaTorrentHash || '-'">{{
                          mediaTorrentHash || "-"
                        }}</strong>
                        <button
                          type="button"
                          class="btn btn-sm btn-outline-secondary js-edit-torrent-hash"
                          :data-jav="javID"
                          :data-hash="mediaTorrentHash"
                          @click="openHashModal"
                        >
                          修改
                        </button>
                      </div>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">BirthTime</span>
                      <strong class="movie-film-value movie-detail-download-date">{{
                        mediaText("BirthTime") ? `下载 ${mediaText("BirthTime")}` : "-"
                      }}</strong>
                    </div>
                    <div class="movie-film-item movie-film-size">
                      <span class="movie-film-label">Size</span>
                      <div class="movie-film-value-line">
                        <a
                          class="movie-film-size-link"
                          :href="mediaText('FilePath')"
                          target="_blank"
                          rel="noopener noreferrer"
                          >{{ mediaSize }}</a
                        >
                        <button
                          type="button"
                          class="btn btn-sm js-move-wmedia-removed"
                          :class="movieDeleteAlbumMarked ? 'btn-danger' : 'btn-outline-danger'"
                          :data-jav="javID"
                          @click="openMoveRemovedModal"
                        >
                          {{ movieDeleteAlbumMarked ? "已在待删除" : "加入待删除" }}
                        </button>
                      </div>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">Height</span>
                      <strong class="movie-film-value">{{ mediaText("Height") || "-" }}</strong>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">BitRate</span>
                      <strong class="movie-film-value">{{ mediaBitRate }}</strong>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">时长(分钟)</span>
                      <strong class="movie-film-value">{{ mediaDuration }}</strong>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">Frame</span>
                      <strong class="movie-film-value">{{ mediaText("Frame") || "-" }}</strong>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">SelfMake</span>
                      <strong class="movie-film-value">{{ mediaText("SelfMake") || "-" }}</strong>
                    </div>
                    <div class="movie-film-item movie-film-item-wide movie-film-dir">
                      <span class="movie-film-label">Dir</span>
                      <a
                        href="#"
                        class="open-in-finder movie-film-path"
                        :data-path="mediaText('Directory')"
                        :title="mediaText('Directory') || '-'"
                        >{{ mediaText("Directory") || "-" }}</a
                      >
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="row movie-resource-row">
          <div class="col-12 mb-3">
            <div class="card shadow-sm movie-section-card movie-resource-card">
              <div class="card-body">
                <div class="movie-resource-heading-row">
                  <h5 class="card-title mb-0 movie-section-title">资源列表</h5>
                  <div v-if="resourceAlbumError" class="movie-resource-album-error" role="alert">
                    {{ resourceAlbumError }}
                  </div>
                </div>
                <ul
                  class="nav nav-tabs nav-tabs-custom mb-3 movie-resource-tabs movie-resource-tabs-compact"
                  role="tablist"
                >
                  <li class="nav-item" role="presentation">
                    <button
                      class="nav-link"
                      :class="{ active: activeResourceTab === 'javbus' }"
                      id="tab-javbus-magnets"
                      type="button"
                      @click="activeResourceTab = 'javbus'"
                    >
                      <span class="movie-resource-tab-text">JavBus 磁力</span>
                      <span v-if="javbusRows.length" class="movie-resource-count">{{ javbusRows.length }}</span>
                    </button>
                  </li>
                  <li class="nav-item" role="presentation">
                    <button
                      class="nav-link"
                      :class="{ active: activeResourceTab === 'sukebei' }"
                      id="tab-sukebei-torrents"
                      type="button"
                      @click="activeResourceTab = 'sukebei'"
                    >
                      <span class="movie-resource-tab-text">Sukebei 种子</span>
                      <span v-if="sukebeiRows.length" class="movie-resource-count">{{ sukebeiRows.length }}</span>
                    </button>
                  </li>
                  <li class="nav-item" role="presentation">
                    <button
                      class="nav-link"
                      :class="{ active: activeResourceTab === 'sehuatang' }"
                      id="tab-sehuatang-magnets"
                      type="button"
                      @click="activeResourceTab = 'sehuatang'"
                    >
                      <span class="movie-resource-tab-text">色花堂磁力</span>
                      <span v-if="sehuatangRows.length" class="movie-resource-count">{{ sehuatangRows.length }}</span>
                    </button>
                  </li>
                </ul>
                <div class="tab-content">
                  <div
                    class="tab-pane fade"
                    :class="{ 'show active': activeResourceTab === 'javbus' }"
                    id="pane-javbus-magnets"
                    role="tabpanel"
                  >
                    <div v-if="!javbusRows.length" class="text-muted">暂无 JavBus 磁力数据</div>
                    <ResourceTable v-else type="javbus_magnet" :rows="javbusRows" :media-hash="mediaTorrentHash" />
                  </div>
                  <div
                    class="tab-pane fade"
                    :class="{ 'show active': activeResourceTab === 'sukebei' }"
                    id="pane-sukebei-torrents"
                    role="tabpanel"
                  >
                    <div v-if="!sukebeiRows.length" class="text-muted">暂无 Sukebei 种子数据</div>
                    <ResourceTable v-else type="sukebei_torrent" :rows="sukebeiRows" :media-hash="mediaTorrentHash" />
                  </div>
                  <div
                    class="tab-pane fade"
                    :class="{ 'show active': activeResourceTab === 'sehuatang' }"
                    id="pane-sehuatang-magnets"
                    role="tabpanel"
                  >
                    <div v-if="!sehuatangRows.length" class="text-muted">暂无色花堂磁力数据</div>
                    <ResourceTable v-else type="sh_magnet" :rows="sehuatangRows" :media-hash="mediaTorrentHash" />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="row mt-3 movie-fetch-row">
          <div class="col-12 mb-2 d-flex align-items-center justify-content-end gap-2 movie-fetch-toolbar">
            <div
              v-if="fetchTaskMessage"
              class="movie-fetch-notice"
              :class="fetchTaskMessageOk ? 'is-success' : 'is-error'"
              role="status"
            >
              {{ fetchTaskMessage }}
            </div>
            <button
              type="button"
              class="btn btn-sm btn-primary js-trigger-both"
              :disabled="fetchTaskBusy"
              @click="startBothFetchTasks"
            >
              {{ fetchTaskBusy ? "触发中" : "同时触发 JavBus + Sukebei" }}
            </button>
          </div>
          <FetchStatusCard
            title="JavBus 抓取状态"
            :status="javbusFetch"
            trigger-class="js-trigger-javbus"
            task-type="spider_fetch_javbus_resources"
            :loading="fetchTaskBusy"
            @trigger="startFetchTask"
          />
          <FetchStatusCard
            title="Sukebei 抓取状态"
            :status="sukebeiFetch"
            trigger-class="js-trigger-sukebei"
            task-type="spider_fetch_sukebei_resources"
            :loading="fetchTaskBusy"
            @trigger="startFetchTask"
          />
        </div>
      </div>

      <div
        class="modal fade"
        id="addMovieCastModal"
        tabindex="-1"
        aria-labelledby="addMovieCastModalLabel"
        :class="{ show: showCastModal }"
        :aria-hidden="showCastModal ? 'false' : 'true'"
      >
        <div class="modal-dialog modal-dialog-centered">
          <div class="modal-content">
            <form id="formAddMovieCast" :data-jav="javID" @submit.prevent="addMovieCast">
              <div class="modal-header">
                <h5 class="modal-title" id="addMovieCastModalLabel">添加已有演员</h5>
                <button type="button" class="btn-close" aria-label="Close" @click="showCastModal = false"></button>
              </div>
              <div class="modal-body">
                <div class="mb-3">
                  <label for="movieCastName" class="form-label">演员名</label>
                  <input
                    v-model="castName"
                    type="text"
                    class="form-control"
                    id="movieCastName"
                    placeholder="输入已存在演员名"
                    :disabled="castBusy"
                  />
                  <div class="form-text">只支持绑定已存在演员；不存在时会失败。</div>
                </div>
                <div
                  v-if="castMessage"
                  id="movieCastMsg"
                  class="small"
                  :class="castMessageOk ? 'text-success' : 'text-danger'"
                >
                  {{ castMessage }}
                </div>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-light" @click="showCastModal = false">取消</button>
                <button type="submit" class="btn btn-primary" :disabled="castBusy">
                  {{ castBusy ? "添加中" : "添加 Cast" }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <div
        class="modal fade"
        id="updateTorrentHashModal"
        tabindex="-1"
        aria-labelledby="updateTorrentHashModalLabel"
        :class="{ show: showHashModal }"
        :aria-hidden="showHashModal ? 'false' : 'true'"
      >
        <div class="modal-dialog modal-dialog-centered">
          <div class="modal-content">
            <form id="formUpdateTorrentHash" @submit.prevent="updateTorrentHash">
              <div class="modal-header">
                <h5 class="modal-title" id="updateTorrentHashModalLabel">修改 TorrentHash</h5>
                <button type="button" class="btn-close" aria-label="Close" @click="showHashModal = false"></button>
              </div>
              <div class="modal-body">
                <div class="mb-3">
                  <label for="movieTorrentHashInput" class="form-label">TorrentHash</label>
                  <input
                    v-model="torrentHashInput"
                    type="text"
                    class="form-control text-monospace"
                    id="movieTorrentHashInput"
                    placeholder="40 位 info_hash，可留空"
                    :disabled="hashBusy"
                  />
                </div>
                <div
                  v-if="hashMessage"
                  id="movieTorrentHashMsg"
                  class="small"
                  :class="hashMessageOk ? 'text-success' : 'text-danger'"
                >
                  {{ hashMessage }}
                </div>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-light" @click="showHashModal = false">取消</button>
                <button type="submit" class="btn btn-primary" :disabled="hashBusy">
                  {{ hashBusy ? "保存中" : "保存" }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <div
        class="modal fade"
        id="moveWMediaRemovedModal"
        tabindex="-1"
        aria-labelledby="moveWMediaRemovedModalLabel"
        :class="{ show: showMoveRemovedModal }"
        :aria-hidden="showMoveRemovedModal ? 'false' : 'true'"
      >
        <div class="modal-dialog modal-dialog-centered">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title" id="moveWMediaRemovedModalLabel">加入待删除</h5>
              <button type="button" class="btn-close" aria-label="Close" @click="showMoveRemovedModal = false"></button>
            </div>
            <div class="modal-body">
              将当前电影加入默认电影相册“待删除”。真正删除会在电影相册页统一执行。
              <div
                v-if="moveRemovedMessage"
                id="moveWMediaRemovedMsg"
                class="small mt-2"
                :class="moveRemovedMessageOk ? 'text-success' : 'text-danger'"
              >
                {{ moveRemovedMessage }}
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-light" @click="showMoveRemovedModal = false">取消</button>
              <button
                type="button"
                class="btn btn-danger"
                id="btnMoveWMediaRemovedDo"
                :disabled="moveRemovedBusy"
                @click="moveWMediaRemoved"
              >
                {{ moveRemovedBusy ? "处理中" : "加入待删除" }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <div
        class="modal fade"
        id="movieAlbumModal"
        tabindex="-1"
        :class="{ show: showMovieAlbumModal }"
        :aria-hidden="showMovieAlbumModal ? 'false' : 'true'"
      >
        <div class="modal-dialog modal-dialog-centered">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">加入电影相册</h5>
              <button type="button" class="btn-close" aria-label="Close" @click="showMovieAlbumModal = false"></button>
            </div>
            <div class="modal-body">
              <div
                v-if="movieAlbumMessage"
                id="movieAlbumMsg"
                class="alert mb-3"
                :class="movieAlbumMessageOk ? 'alert-success' : 'alert-danger'"
                role="alert"
              >
                {{ movieAlbumMessage }}
              </div>
              <div class="mb-3">
                <label class="form-label mb-1" for="movieAlbumSelect">选择相册</label>
                <select
                  v-model="selectedMovieAlbumName"
                  class="form-select"
                  id="movieAlbumSelect"
                  :disabled="movieAlbumBusy"
                >
                  <option v-for="album in albumOptions" :key="album.name" :value="album.name">
                    {{ album.name }}{{ album.selected ? "（已在相册）" : "" }}
                  </option>
                </select>
              </div>
              <div>
                <label class="form-label mb-1" for="movieAlbumNewName">新建相册</label>
                <div class="d-flex gap-2">
                  <input
                    v-model="newMovieAlbumName"
                    type="text"
                    class="form-control"
                    id="movieAlbumNewName"
                    maxlength="128"
                    placeholder="输入新相册名称后创建"
                    :disabled="movieAlbumBusy"
                  />
                  <button
                    type="button"
                    class="btn btn-outline-secondary"
                    id="btnCreateMovieAlbumInline"
                    :disabled="movieAlbumBusy"
                    @click="createMovieAlbumInline"
                  >
                    {{ movieAlbumBusy ? "处理中" : "创建" }}
                  </button>
                </div>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-light" @click="showMovieAlbumModal = false">取消</button>
              <button
                type="button"
                class="btn btn-primary"
                id="btnAddMovieToAlbum"
                :disabled="movieAlbumBusy"
                @click="addMovieToAlbum"
              >
                {{ movieAlbumBusy ? "处理中" : "加入" }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/features/movies/api";
import { useMovieActions } from "@/features/movies/composables/useMovieActions";
import { useMovieAlbums } from "@/features/movies/composables/useMovieAlbums";
import { useMovieResources } from "@/features/movies/composables/useMovieResources";
import ErrorPanel from "@/shared/components/ErrorPanel.vue";
import LoadingState from "@/shared/components/LoadingState.vue";
import PageShell from "@/shared/components/PageShell.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import type { MovieRecord } from "@/features/movies/types";
import { asArray, asRecord, num, pick, text } from "@/shared/utils/format";

type SortOrder = "asc" | "desc";
type RankSortKey = "date" | "rank";

const route = useRoute();
const payload = ref<MovieRecord>({});
const loading = ref(false);
const error = ref("");
const activeDetailTab = ref("info");
const { activeResourceTab, resourceSorts, reset: resetMovieResources } = useMovieResources();
const {
  showMovieAlbumModal,
  selectedMovieAlbumName,
  newMovieAlbumName,
  movieAlbumBusy,
  movieAlbumMessage,
  movieAlbumMessageOk,
  reset: resetMovieAlbums
} = useMovieAlbums();
const showCastModal = ref(false);
const showHashModal = ref(false);
const showMoveRemovedModal = ref(false);
const downloadLaterBusy = ref(false);
const downloadLaterMessage = ref("");
const downloadLaterError = ref("");
const {
  busy: coverBusy,
  message: coverMessage,
  error: coverError,
  reset: resetCoverAction,
  execute: executeCoverAction
} = useMovieActions();
const fetchTaskBusy = ref(false);
const fetchTaskMessage = ref("");
const fetchTaskMessageOk = ref(true);
const resourceAlbumError = ref("");
const favoriteBusyKey = ref("");
const rankPage = ref(1);
const rankPageSize = ref(10);
const rankSortKey = ref<RankSortKey>("date");
const rankSortOrder = ref<SortOrder>("desc");
const castName = ref("");
const castBusy = ref(false);
const castMessage = ref("");
const castMessageOk = ref(true);
const torrentHashInput = ref("");
const hashBusy = ref(false);
const hashMessage = ref("");
const hashMessageOk = ref(true);
const moveRemovedBusy = ref(false);
const moveRemovedMessage = ref("");
const moveRemovedMessageOk = ref(true);

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  activeDetailTab.value = "info";
  resetMovieResources();
  resetMovieAlbums();
  downloadLaterMessage.value = "";
  downloadLaterError.value = "";
  resetCoverAction();
  fetchTaskMessage.value = "";
  resourceAlbumError.value = "";
  rankPage.value = 1;
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const movie = computed(() => asRecord(pick(payload.value, "movieType", "MovieType", "movie", "Movie")));
const media = computed(() => asRecord(pick(payload.value, "media", "Media", "MediaInfo")));
const pageTitle = computed(() => titleText.value || movieName.value || String(route.params.movieName || "Movie"));
const movieName = computed(() => text(movie.value, "name", "Name", "jav_id", "JavID"));
const javID = computed(() => text(movie.value, "jav_id", "JavId") || movieName.value);
const titleText = computed(() => text(movie.value, "title", "Title"));
const subtitle = computed(() => [releaseDate.value, maker.value, label.value].filter(Boolean).join(" · "));
const poster = computed(() => text(movie.value, "jacket_img", "JacketImg"));
const prefix = computed(() => text(movie.value, "prefix", "Prefix"));
const encodeName = computed(() => text(movie.value, "encode_name", "EncodeName"));
const releaseDate = computed(() => text(movie.value, "releasing_date", "ReleasingDate"));
const updateDate = computed(() => text(movie.value, "file_creating_date", "UpdateDate"));
const lengthText = computed(() => text(movie.value, "length", "Length"));
const maker = computed(() => text(movie.value, "maker", "Maker"));
const label = computed(() => text(movie.value, "label", "Label"));
const director = computed(() => text(movie.value, "director", "Director"));
const score = computed(() => text(movie.value, "score", "Score"));
const watched = computed(() => text(movie.value, "viewers_number_watched", "ViewersNumberWatched"));
const firstRankingDay = computed(() => text(movie.value, "first_ranking_day", "FirstRankingDay"));
const busUrl = computed(() => text(movie.value, "bus_url", "BusUrl"));
const searchUrl = computed(() => text(movie.value, "search_url", "SearchUrl"));
const javUrl = computed(() => text(movie.value, "jav_url", "JavUrl"));
const videoUrl = computed(() => text(movie.value, "video_url_w_media", "VideoUrlWMedia"));
const comeTimes = computed(() => num(movie.value, "come_times", "ComeTimes"));
const scTimes = computed(() => num(movie.value, "sc_times", "ScTimes"));
const needDownload = computed(() => num(movie.value, "need_download", "NeedDownload"));
const downloadLaterActive = computed(() => needDownload.value === 2);
const downloadLaterTitle = computed(() => (downloadLaterActive.value ? "取消稍后下载" : "添加稍后下载"));
const downloadLaterButtonText = computed(() => {
  if (downloadLaterBusy.value) {
    return "处理中";
  }
  return downloadLaterActive.value ? "已添加" : "稍后下载";
});
const ownedMedia = computed(() => num(movie.value, "owned_w_media", "OwnedWMedia"));
const mediaLabel = computed(() =>
  ownedMedia.value === 6 ? "Removed" : ownedMedia.value === 4 || ownedMedia.value === 5 ? "Play" : ""
);
const mediaButtonClass = computed(() => (ownedMedia.value === 5 ? "btn-info" : "btn-success"));
const mediaPlayable = computed(() => Boolean(videoUrl.value) && (ownedMedia.value === 4 || ownedMedia.value === 5));
const genres = computed(() =>
  asArray<string>(pick(movie.value, "genre", "Genre"))
    .map(String)
    .filter(Boolean)
);
const casts = computed(() =>
  asArray<MovieRecord>(pick(movie.value, "cast", "Cast"))
    .map((item) => {
      const row = asRecord(item);
      const labelText = text(row, "NameShow", "DisplayName", "Name", "name");
      const href =
        text(row, "Url", "url") || `/listcard?cn=${encodeURIComponent(text(row, "Name", "name") || labelText)}`;
      return { label: labelText, href };
    })
    .filter((item) => item.label)
);
const castSummary = computed(() => casts.value.map((cast) => cast.label).join("、"));
const scRows = computed(() =>
  asArray<MovieRecord>(pick(payload.value, "sc", "SC"))
    .map((row) => ({
      name: text(row, "ScName"),
      href: text(row, "Href")
    }))
    .filter((row) => row.name)
);
const rankRows = computed(() =>
  asArray<MovieRecord>(pick(payload.value, "rankInfos", "RankInfos")).map((row) => ({
    date: text(row, "Date"),
    rank: text(row, "Rank")
  }))
);
const sortedRankRows = computed(() => {
  const order = rankSortOrder.value === "asc" ? 1 : -1;
  return [...rankRows.value].sort((left, right) => {
    const result =
      rankSortKey.value === "rank" ? rankValue(left.rank) - rankValue(right.rank) : left.date.localeCompare(right.date);
    return result === 0 ? left.date.localeCompare(right.date) : result * order;
  });
});
const rankTotalPages = computed(() => Math.max(1, Math.ceil(sortedRankRows.value.length / rankPageSize.value)));
const pagedRankRows = computed(() => {
  const page = Math.min(Math.max(rankPage.value, 1), rankTotalPages.value);
  const start = (page - 1) * rankPageSize.value;
  return sortedRankRows.value.slice(start, start + rankPageSize.value);
});
const albums = computed(() => asArray<MovieRecord>(pick(payload.value, "MovieAlbums", "movieAlbums")));
const albumOptions = computed(() =>
  albums.value
    .map((album) => ({
      name: text(album, "Name", "name"),
      selected: Boolean(album.Selected || album.selected)
    }))
    .filter((album) => album.name)
);
const selectedAlbums = computed(() => albumOptions.value.filter((album) => album.selected));
const javbusRows = computed(() => asArray<MovieRecord>(payload.value.javbusMagnets));
const sukebeiRows = computed(() => asArray<MovieRecord>(payload.value.sukebeiRows));
const sehuatangRows = computed(() => asArray<MovieRecord>(payload.value.sehuatangRows));
const javbusFetch = computed(() => asRecord(payload.value.javbusFetch));
const sukebeiFetch = computed(() => asRecord(payload.value.sukebeiFetch));
const movieDeleteAlbumMarked = computed(() =>
  Boolean(payload.value.MovieDeleteAlbumMarked || payload.value.movieDeleteAlbumMarked)
);
const showFilmInfo = computed(() => Object.keys(media.value).length > 0 && ownedMedia.value !== 6);
const mediaTorrentHash = computed(() => text(media.value, "SourceTorrentHash"));
const mediaSize = computed(() => Number(text(media.value, "Size") || 0).toFixed(2));
const mediaBitRate = computed(() => Number(text(media.value, "BitRate") || 0).toFixed(0));
const mediaDuration = computed(() => Number(text(media.value, "DurationMinutes") || 0).toFixed(2));

watch([rankTotalPages, rankPageSize], () => {
  if (rankPage.value > rankTotalPages.value) {
    rankPage.value = rankTotalPages.value;
  }
  if (rankPage.value < 1) {
    rankPage.value = 1;
  }
});

watch(
  albumOptions,
  (options) => {
    if (!options.length) {
      selectedMovieAlbumName.value = "";
      return;
    }
    if (!options.some((album) => album.name === selectedMovieAlbumName.value)) {
      selectedMovieAlbumName.value = options[0].name;
    }
  },
  { immediate: true }
);

function mediaText(key: string) {
  return text(media.value, key);
}

function rankValue(value: string) {
  const numeric = Number.parseFloat(value);
  return Number.isFinite(numeric) ? numeric : 0;
}

function sortRankTable(key: RankSortKey) {
  if (rankSortKey.value === key) {
    rankSortOrder.value = rankSortOrder.value === "asc" ? "desc" : "asc";
  } else {
    rankSortKey.value = key;
    rankSortOrder.value = "desc";
  }
  rankPage.value = 1;
}

function rankSortMark(key: RankSortKey) {
  if (rankSortKey.value !== key) {
    return "";
  }
  return rankSortOrder.value === "asc" ? "↑" : "↓";
}

function errText(err: unknown, fallback = "网络错误，请稍后重试") {
  return err instanceof Error ? err.message : String(err || fallback);
}

function showFetchTaskMessage(value: string, ok: boolean) {
  fetchTaskMessage.value = value;
  fetchTaskMessageOk.value = ok;
}

function updatePayloadRecord(keys: string[], updates: MovieRecord) {
  const nextPayload: MovieRecord = { ...payload.value };
  const targetKey =
    keys.find((key) => {
      const value = nextPayload[key];
      return Boolean(value) && typeof value === "object" && !Array.isArray(value);
    }) || keys[0];
  nextPayload[targetKey] = {
    ...asRecord(nextPayload[targetKey]),
    ...updates
  };
  payload.value = nextPayload;
}

function updateMovieNeedDownload(nextNeedDownload: number) {
  const nextPayload: MovieRecord = { ...payload.value };
  const movieKeys = ["movieType", "MovieType", "movie", "Movie"];
  const movieKey =
    movieKeys.find((key) => {
      const value = nextPayload[key];
      return Boolean(value) && typeof value === "object" && !Array.isArray(value);
    }) || "movieType";
  nextPayload[movieKey] = {
    ...asRecord(nextPayload[movieKey]),
    need_download: nextNeedDownload,
    NeedDownload: nextNeedDownload
  };
  payload.value = nextPayload;
}

async function toggleDownloadLater() {
  if (!javID.value || downloadLaterBusy.value) {
    return;
  }
  const current = downloadLaterActive.value;
  downloadLaterBusy.value = true;
  downloadLaterMessage.value = "";
  downloadLaterError.value = "";
  coverMessage.value = "";
  coverError.value = "";
  try {
    const data = asRecord(
      await sendJSON(`/api/movie/${encodeURIComponent(javID.value)}/downloadlater`, current ? "DELETE" : "POST")
    );
    const nextNeedDownload = num(data, "needDownload", "need_download", "NeedDownload") || (current ? 1 : 2);
    updateMovieNeedDownload(nextNeedDownload);
    downloadLaterMessage.value = nextNeedDownload === 2 ? "已加入稍后下载" : "已取消稍后下载";
  } catch (err) {
    downloadLaterError.value = err instanceof Error ? err.message : String(err);
  } finally {
    downloadLaterBusy.value = false;
  }
}

async function downloadCoverNow() {
  if (!javID.value || coverBusy.value) {
    return;
  }
  downloadLaterMessage.value = "";
  downloadLaterError.value = "";
  await executeCoverAction(
    () => sendJSON(`/api/movie/${encodeURIComponent(javID.value)}/download-cover`, "POST"),
    "封面下载完成"
  );
}

function openCastModal() {
  castName.value = "";
  castMessage.value = "";
  showCastModal.value = true;
}

async function addMovieCast() {
  const name = castName.value.trim();
  if (!javID.value) {
    castMessage.value = "缺少影片参数";
    castMessageOk.value = false;
    return;
  }
  if (!name) {
    castMessage.value = "演员名不能为空";
    castMessageOk.value = false;
    return;
  }
  castBusy.value = true;
  castMessage.value = "";
  try {
    await sendJSON(`/api/movie/${encodeURIComponent(javID.value)}/add-cast`, "POST", { name });
    castMessage.value = "添加成功";
    castMessageOk.value = true;
    showCastModal.value = false;
    await load();
  } catch (err) {
    castMessage.value = errText(err);
    castMessageOk.value = false;
  } finally {
    castBusy.value = false;
  }
}

function openHashModal() {
  torrentHashInput.value = mediaTorrentHash.value;
  hashMessage.value = "";
  showHashModal.value = true;
}

async function updateTorrentHash() {
  if (!javID.value) {
    hashMessage.value = "缺少影片参数";
    hashMessageOk.value = false;
    return;
  }
  hashBusy.value = true;
  hashMessage.value = "";
  try {
    const data = asRecord(
      await sendJSON(`/api/movie/${encodeURIComponent(javID.value)}/torrent-hash`, "PUT", {
        torrent_hash: torrentHashInput.value.trim()
      })
    );
    updatePayloadRecord(["media", "Media", "MediaInfo"], {
      SourceTorrentHash: text(data, "torrent_hash", "TorrentHash") || torrentHashInput.value.trim().toUpperCase()
    });
    hashMessage.value = "保存成功";
    hashMessageOk.value = true;
    showHashModal.value = false;
  } catch (err) {
    hashMessage.value = errText(err);
    hashMessageOk.value = false;
  } finally {
    hashBusy.value = false;
  }
}

function openMoveRemovedModal() {
  moveRemovedMessage.value = "";
  showMoveRemovedModal.value = true;
}

async function moveWMediaRemoved() {
  if (!javID.value) {
    moveRemovedMessage.value = "缺少影片参数";
    moveRemovedMessageOk.value = false;
    return;
  }
  moveRemovedBusy.value = true;
  moveRemovedMessage.value = "";
  try {
    const data = asRecord(await sendJSON(`/api/movie/${encodeURIComponent(javID.value)}/move-wmedia-removed`, "POST"));
    moveRemovedMessage.value = text(data, "message") || "已加入待删除相册";
    moveRemovedMessageOk.value = true;
    payload.value = { ...payload.value, MovieDeleteAlbumMarked: true, movieDeleteAlbumMarked: true };
    await load();
    showMoveRemovedModal.value = false;
  } catch (err) {
    moveRemovedMessage.value = errText(err);
    moveRemovedMessageOk.value = false;
  } finally {
    moveRemovedBusy.value = false;
  }
}

function openMovieAlbumModal() {
  movieAlbumMessage.value = "";
  newMovieAlbumName.value = "";
  if (!selectedMovieAlbumName.value && albumOptions.value.length) {
    selectedMovieAlbumName.value = albumOptions.value[0].name;
  }
  showMovieAlbumModal.value = true;
}

function upsertMovieAlbumOption(name: string, selected: boolean) {
  const nextPayload: MovieRecord = { ...payload.value };
  const key = Array.isArray(nextPayload.MovieAlbums) ? "MovieAlbums" : "movieAlbums";
  const rows = asArray<MovieRecord>(nextPayload[key]).map((album) => ({ ...album }));
  const found = rows.find((album) => text(album, "Name", "name") === name);
  if (found) {
    found.Name = name;
    found.name = name;
    found.Selected = selected || Boolean(found.Selected || found.selected);
    found.selected = selected || Boolean(found.Selected || found.selected);
  } else {
    rows.push({ Name: name, name, Selected: selected, selected });
  }
  nextPayload[key] = rows;
  payload.value = nextPayload;
}

async function createMovieAlbumInline() {
  const name = newMovieAlbumName.value.trim();
  if (!name) {
    movieAlbumMessage.value = "请输入电影相册名称";
    movieAlbumMessageOk.value = false;
    return;
  }
  movieAlbumBusy.value = true;
  movieAlbumMessage.value = "";
  try {
    const data = asRecord(await sendJSON("/api/movie-albums", "POST", { name }));
    const albumName = text(data, "album_name", "AlbumName") || name;
    upsertMovieAlbumOption(albumName, false);
    selectedMovieAlbumName.value = albumName;
    newMovieAlbumName.value = "";
    movieAlbumMessage.value = text(data, "message") || "电影相册创建成功";
    movieAlbumMessageOk.value = true;
  } catch (err) {
    movieAlbumMessage.value = errText(err, "电影相册创建失败");
    movieAlbumMessageOk.value = false;
  } finally {
    movieAlbumBusy.value = false;
  }
}

async function addMovieToAlbum() {
  const albumName = selectedMovieAlbumName.value.trim();
  if (!javID.value) {
    movieAlbumMessage.value = "缺少影片编号";
    movieAlbumMessageOk.value = false;
    return;
  }
  if (!albumName) {
    movieAlbumMessage.value = "请选择电影相册";
    movieAlbumMessageOk.value = false;
    return;
  }
  movieAlbumBusy.value = true;
  movieAlbumMessage.value = "";
  try {
    const data = asRecord(
      await sendJSON(`/api/movie/${encodeURIComponent(javID.value)}/movie-album-item`, "POST", {
        movie_jav_id: javID.value,
        album_name: albumName
      })
    );
    const responseAlbumName = text(data, "album_name", "AlbumName") || albumName;
    upsertMovieAlbumOption(responseAlbumName, true);
    movieAlbumMessage.value = text(data, "message") || "已加入电影相册";
    movieAlbumMessageOk.value = true;
    await load();
    showMovieAlbumModal.value = false;
  } catch (err) {
    movieAlbumMessage.value = errText(err, "加入电影相册失败");
    movieAlbumMessageOk.value = false;
  } finally {
    movieAlbumBusy.value = false;
  }
}

function taskSiteName(taskType: string) {
  return taskType === "spider_fetch_javbus_resources" ? "JavBus" : "Sukebei";
}

async function startFetchTask(taskType: string) {
  if (!javID.value || !movieName.value) {
    showFetchTaskMessage("缺少影片参数，无法触发任务", false);
    return;
  }
  fetchTaskBusy.value = true;
  fetchTaskMessage.value = "";
  try {
    await sendJSON("/api/crawler/jobs/start", "POST", {
      task_type: taskType,
      movie_jav_id: javID.value,
      movie_name: movieName.value
    });
    showFetchTaskMessage(`已开启${taskSiteName(taskType)}抓取：${movieName.value}`, true);
  } catch (err) {
    showFetchTaskMessage(errText(err, "启动失败"), false);
  } finally {
    fetchTaskBusy.value = false;
  }
}

async function startBothFetchTasks() {
  if (!javID.value || !movieName.value) {
    showFetchTaskMessage("缺少影片参数，无法触发任务", false);
    return;
  }
  fetchTaskBusy.value = true;
  fetchTaskMessage.value = "";
  const tasks = ["spider_fetch_javbus_resources", "spider_fetch_sukebei_resources"];
  try {
    const results = await Promise.allSettled(
      tasks.map((taskType) =>
        sendJSON("/api/crawler/jobs/start", "POST", {
          task_type: taskType,
          movie_jav_id: javID.value,
          movie_name: movieName.value
        })
      )
    );
    const successSites: string[] = [];
    const failedSites: string[] = [];
    results.forEach((result, index) => {
      const site = index === 0 ? "JavBus" : "Sukebei";
      if (result.status === "fulfilled") {
        successSites.push(site);
      } else {
        failedSites.push(`${site}：${errText(result.reason, "启动失败")}`);
      }
    });
    if (successSites.length === 2) {
      showFetchTaskMessage(`已同时开启 JavBus + Sukebei 抓取：${movieName.value}`, true);
    } else if (successSites.length === 1) {
      showFetchTaskMessage(`部分成功：已开启${successSites[0]}；${failedSites.join("；")}`, false);
    } else {
      showFetchTaskMessage(`同时触发失败：${failedSites.join("；")}`, false);
    }
  } finally {
    fetchTaskBusy.value = false;
  }
}

function resourceRowID(row: MovieRecord) {
  return Number(text(row, "RowID", "row_id", "ID", "id") || 0);
}

function normalizeResourceInfoHash(value: string) {
  return value.trim().toUpperCase();
}

function resourceHashMatchesMedia(infoHash: string, mediaHash: string) {
  const normalizedInfoHash = normalizeResourceInfoHash(infoHash);
  const normalizedMediaHash = normalizeResourceInfoHash(mediaHash);
  return Boolean(normalizedInfoHash && normalizedMediaHash && normalizedInfoHash === normalizedMediaHash);
}

function favoriteKey(row: MovieRecord, sourceType: string, albumName: string) {
  return `${sourceType}:${resourceRowID(row)}:${albumName}`;
}

function isResourceFavorited(row: MovieRecord, albumName: string) {
  return albumName === "待下载"
    ? Boolean(pick(row, "IsInPending", "is_in_pending", "InPending"))
    : Boolean(pick(row, "IsInDownload", "is_in_download", "InDownload"));
}

function setResourceFavorited(row: MovieRecord, albumName: string, favorited: boolean) {
  if (albumName === "待下载") {
    row.IsInPending = favorited;
    row.is_in_pending = favorited;
    row.InPending = favorited;
    return;
  }
  row.IsInDownload = favorited;
  row.is_in_download = favorited;
  row.InDownload = favorited;
}

function favoriteButtonClass(row: MovieRecord, albumName: string) {
  const favorited = isResourceFavorited(row, albumName);
  const albumClass = albumName === "待下载" ? "movie-resource-album-pending" : "movie-resource-album-downloading";
  const stateClass = favorited ? "is-selected" : "is-idle";
  const bootstrapClass =
    albumName === "待下载"
      ? favorited
        ? "btn-info"
        : "btn-outline-info"
      : favorited
        ? "btn-warning"
        : "btn-outline-warning";
  return `movie-resource-album-state ${albumClass} ${stateClass} ${bootstrapClass}`;
}

function favoriteButtonText(row: MovieRecord, albumName: string) {
  return isResourceFavorited(row, albumName) ? `已在${albumName}` : albumName;
}

function favoriteButtonTitle(row: MovieRecord, albumName: string) {
  return isResourceFavorited(row, albumName) ? `移出${albumName}` : albumName;
}

async function switchFetchResourceAlbum(row: MovieRecord, sourceType: string, albumName: string) {
  const sourceRowID = resourceRowID(row);
  const currentFavorited = isResourceFavorited(row, albumName);
  if (!javID.value || !sourceType || !Number.isInteger(sourceRowID) || sourceRowID <= 0) {
    resourceAlbumError.value = `缺少${albumName}参数，无法执行`;
    return;
  }
  const key = favoriteKey(row, sourceType, albumName);
  resourceAlbumError.value = "";
  favoriteBusyKey.value = key;
  try {
    const data = asRecord(
      await sendJSON(`/api/movie/${encodeURIComponent(javID.value)}/album-item`, currentFavorited ? "DELETE" : "POST", {
        movie_jav_id: javID.value,
        source_type: sourceType,
        source_row_id: sourceRowID,
        album_name: albumName
      })
    );
    const favorited = typeof data.favorited === "boolean" ? data.favorited : !currentFavorited;
    const responseAlbumName = text(data, "album_name", "AlbumName") || albumName;
    setResourceFavorited(row, responseAlbumName, favorited);
  } catch (err) {
    resourceAlbumError.value = errText(err, `${albumName}操作失败`);
  } finally {
    favoriteBusyKey.value = "";
  }
}

function parseSizeToBytes(value: string) {
  const match = value
    .trim()
    .toUpperCase()
    .match(/([0-9]+(?:\.[0-9]+)?)\s*(B|KB|KIB|MB|MIB|GB|GIB|TB|TIB)/);
  if (!match) {
    return 0;
  }
  const units: Record<string, number> = {
    B: 1,
    KB: 1024,
    KIB: 1024,
    MB: 1024 ** 2,
    MIB: 1024 ** 2,
    GB: 1024 ** 3,
    GIB: 1024 ** 3,
    TB: 1024 ** 4,
    TIB: 1024 ** 4
  };
  return Number(match[1]) * (units[match[2]] || 1);
}

function formatResourceSizeGB(value: string) {
  const source = value.trim();
  if (!source || source === "-") {
    return "-";
  }
  const bytes = parseSizeToBytes(source);
  if (!bytes && !/[0-9]/.test(source)) {
    return "-";
  }
  return `${(bytes / 1024 ** 3).toFixed(2)} GB`;
}

function parseDateValue(value: string) {
  const ts = Date.parse(value.trim().replace(/\s+/g, "T"));
  return Number.isNaN(ts) ? 0 : ts;
}

function parseNumberValue(value: string) {
  const parsed = Number.parseInt(value.replace(/[^\d-]/g, ""), 10);
  return Number.isNaN(parsed) ? 0 : parsed;
}

function resourceSortValue(row: MovieRecord, type: string, key: string) {
  if (key === "size") {
    return parseSizeToBytes(text(row, "SizeText"));
  }
  if (key === "date") {
    return parseDateValue(
      text(row, type === "javbus_magnet" ? "ShareDate" : type === "sukebei_torrent" ? "PublishTime" : "PostTime")
    );
  }
  if (key === "downloads") {
    return parseNumberValue(text(row, "Completed"));
  }
  return text(row, "MagnetName", "TorrentTitle", "ThreadTitle").toLowerCase();
}

function sortResourceTable(type: string, key: string) {
  const current = resourceSorts.value[type];
  const order = current?.key === key && current.order === "desc" ? "asc" : "desc";
  resourceSorts.value = {
    ...resourceSorts.value,
    [type]: { key, order }
  };
}

function resourceSortMark(type: string, key: string) {
  const current = resourceSorts.value[type];
  if (!current || current.key !== key) {
    return "";
  }
  return current.order === "asc" ? " ↑" : " ↓";
}

function sortedResourceRows(rows: MovieRecord[], type: string) {
  const current = resourceSorts.value[type];
  if (!current) {
    return rows;
  }
  return [...rows].sort((left, right) => {
    const leftValue = resourceSortValue(asRecord(left), type, current.key);
    const rightValue = resourceSortValue(asRecord(right), type, current.key);
    if (leftValue === rightValue) {
      return 0;
    }
    if (current.order === "asc") {
      return leftValue > rightValue ? 1 : -1;
    }
    return leftValue < rightValue ? 1 : -1;
  });
}

const ResourceTable = defineComponent({
  props: {
    rows: { type: Array<MovieRecord>, required: true },
    type: { type: String, required: true },
    mediaHash: { type: String, default: "" }
  },
  setup(props) {
    const sortButton = (label: string, key: string) =>
      h(
        "button",
        {
          type: "button",
          class: "btn btn-link btn-sm p-0 text-decoration-none js-sort-table movie-resource-sort",
          "data-table-id": props.type,
          "data-sort-key": key,
          onClick: () => sortResourceTable(props.type, key)
        },
        `${label}${resourceSortMark(props.type, key)}`
      );

    const favoriteButtons = (record: MovieRecord) => [
      h(
        "button",
        {
          type: "button",
          class: `btn btn-sm movie-resource-action ${favoriteButtonClass(record, "下载中")} js-add-favorite`,
          "data-source-type": props.type,
          "data-source-row-id": String(resourceRowID(record)),
          "data-album-name": "下载中",
          "data-favorited": isResourceFavorited(record, "下载中") ? "1" : "0",
          "aria-pressed": isResourceFavorited(record, "下载中") ? "true" : "false",
          title: favoriteButtonTitle(record, "下载中"),
          disabled: favoriteBusyKey.value === favoriteKey(record, props.type, "下载中"),
          onClick: () => switchFetchResourceAlbum(record, props.type, "下载中")
        },
        favoriteButtonText(record, "下载中")
      ),
      h(
        "button",
        {
          type: "button",
          class: `btn btn-sm ms-1 movie-resource-action ${favoriteButtonClass(record, "待下载")} js-add-favorite`,
          "data-source-type": props.type,
          "data-source-row-id": String(resourceRowID(record)),
          "data-album-name": "待下载",
          "data-favorited": isResourceFavorited(record, "待下载") ? "1" : "0",
          "aria-pressed": isResourceFavorited(record, "待下载") ? "true" : "false",
          title: favoriteButtonTitle(record, "待下载"),
          disabled: favoriteBusyKey.value === favoriteKey(record, props.type, "待下载"),
          onClick: () => switchFetchResourceAlbum(record, props.type, "待下载")
        },
        favoriteButtonText(record, "待下载")
      )
    ];

    const hashCell = (record: MovieRecord) => {
      const infoHash = text(record, "InfoHash");
      const matched = resourceHashMatchesMedia(infoHash, props.mediaHash);
      return h("td", { class: "movie-resource-col-hash text-monospace" }, [
        h(
          "span",
          { class: `movie-resource-hash ${matched ? "fetch-site-hash-match" : ""}`, title: infoHash },
          infoHash || "-"
        )
      ]);
    };

    const colgroup = () => {
      const widths =
        props.type === "sukebei_torrent"
          ? ["42%", "72px", "104px", "74px", "240px", "132px"]
          : props.type === "sh_magnet"
            ? ["48%", "116px", "240px", "132px"]
            : ["38%", "72px", "96px", "70px", "240px", "132px"];
      return h(
        "colgroup",
        widths.map((width) => h("col", { style: { width } }))
      );
    };

    const renderRows = () =>
      sortedResourceRows(props.rows as MovieRecord[], props.type).map((row) => {
        const record = asRecord(row);
        const infoHash = text(record, "InfoHash");
        const matched = Boolean(pick(record, "HasMatchedHash")) || resourceHashMatchesMedia(infoHash, props.mediaHash);
        if (props.type === "sukebei_torrent") {
          const title = text(record, "TorrentTitle");
          const url = text(record, "ViewURL");
          return h("tr", { class: matched ? "fetch-site-row-match" : "" }, [
            h("td", { class: "movie-resource-title-cell fetch-site-col-title" }, [
              url
                ? h(
                    "a",
                    {
                      class: "movie-resource-title-link",
                      href: url,
                      target: "_blank",
                      rel: "noopener noreferrer",
                      title
                    },
                    title
                  )
                : h("span", { class: "movie-resource-title-text", title }, title)
            ]),
            h(
              "td",
              { class: "movie-resource-number-cell fetch-site-col-compact text-nowrap", "data-sort-key": "size" },
              formatResourceSizeGB(text(record, "SizeText"))
            ),
            h(
              "td",
              { class: "movie-resource-number-cell fetch-site-col-compact text-nowrap", "data-sort-key": "date" },
              [h(ScDateTime, { value: text(record, "PublishTime") })]
            ),
            h(
              "td",
              { class: "movie-resource-number-cell fetch-site-col-compact text-nowrap", "data-sort-key": "downloads" },
              text(record, "Completed") || "-"
            ),
            hashCell(record),
            h(
              "td",
              { class: "movie-resource-actions-cell fetch-site-col-compact text-nowrap" },
              favoriteButtons(record)
            )
          ]);
        }
        if (props.type === "sh_magnet") {
          const title = text(record, "ThreadTitle");
          const url = text(record, "ThreadURL");
          return h("tr", { class: matched ? "fetch-site-row-match" : "" }, [
            h("td", { class: "movie-resource-title-cell" }, [
              url
                ? h(
                    "a",
                    {
                      class: "movie-resource-title-link",
                      href: url,
                      target: "_blank",
                      rel: "noopener noreferrer",
                      title
                    },
                    title
                  )
                : h("span", { class: "movie-resource-title-text", title }, title)
            ]),
            h("td", { class: "movie-resource-number-cell text-nowrap", "data-sort-key": "date" }, [
              h(ScDateTime, { value: text(record, "PostTime") })
            ]),
            hashCell(record),
            h("td", { class: "movie-resource-actions-cell text-nowrap" }, favoriteButtons(record))
          ]);
        }
        return h("tr", { class: matched ? "fetch-site-row-match" : "" }, [
          h(
            "td",
            { class: "movie-resource-title-cell" },
            h(
              "span",
              { class: "movie-resource-title-text", title: text(record, "MagnetName") || "-" },
              text(record, "MagnetName") || "-"
            )
          ),
          h(
            "td",
            { class: "movie-resource-number-cell", "data-sort-key": "size" },
            formatResourceSizeGB(text(record, "SizeText"))
          ),
          h(
            "td",
            { class: "movie-resource-number-cell text-nowrap", "data-sort-key": "date" },
            text(record, "ShareDate") || "-"
          ),
          h(
            "td",
            { class: "movie-resource-number-cell text-nowrap" },
            pick(record, "HasSubtitle")
              ? h("span", { class: "fetch-site-tag-subtitle" }, "字幕")
              : h("span", { class: "text-muted" }, "-")
          ),
          hashCell(record),
          h("td", { class: "movie-resource-actions-cell text-nowrap" }, favoriteButtons(record))
        ]);
      });

    const renderHead = () => {
      if (props.type === "sukebei_torrent") {
        return h("tr", [
          h("th", { class: "movie-resource-title-head fetch-site-col-title" }, "标题"),
          h("th", { class: "movie-resource-number-head fetch-site-col-compact" }, sortButton("大小", "size")),
          h("th", { class: "movie-resource-number-head fetch-site-col-compact" }, sortButton("日期", "date")),
          h("th", { class: "movie-resource-number-head fetch-site-col-compact" }, sortButton("下载", "downloads")),
          h("th", { class: "movie-resource-hash-head fetch-site-col-hash" }, "Info Hash"),
          h("th", { class: "movie-resource-actions-head fetch-site-col-compact" }, "相册")
        ]);
      }
      if (props.type === "sh_magnet") {
        return h("tr", [
          h("th", { class: "movie-resource-title-head" }, "帖子标题"),
          h("th", { class: "movie-resource-number-head" }, sortButton("发帖时间", "date")),
          h("th", { class: "movie-resource-hash-head" }, "Info Hash"),
          h("th", { class: "movie-resource-actions-head" }, "相册")
        ]);
      }
      return h("tr", [
        h("th", { class: "movie-resource-title-head" }, "名称"),
        h("th", { class: "movie-resource-number-head" }, sortButton("大小", "size")),
        h("th", { class: "movie-resource-number-head" }, sortButton("日期", "date")),
        h("th", { class: "movie-resource-number-head" }, "标签"),
        h("th", { class: "movie-resource-hash-head" }, "Info Hash"),
        h("th", { class: "movie-resource-actions-head" }, "相册")
      ]);
    };

    return () => {
      const tableTypeClass = props.type.replace(/_/g, "-");
      return h(
        "div",
        {
          class: `movie-resource-table-wrap movie-resource-table-wrap-${tableTypeClass} ${props.type === "sukebei_torrent" ? "fetch-site-detail-sukebei-wrap" : "table-responsive"}`
        },
        [
          h(
            "table",
            {
              class: `table table-sm table-hover align-middle mb-0 movie-resource-table movie-resource-table-${tableTypeClass} ${props.type === "sukebei_torrent" ? "fetch-site-detail-sukebei-table" : ""}`,
              "data-sort-table": props.type
            },
            [colgroup(), h("thead", [renderHead()]), h("tbody", renderRows())]
          )
        ]
      );
    };
  }
});

const FetchStatusCard = defineComponent({
  props: {
    title: { type: String, required: true },
    status: { type: Object as () => MovieRecord, required: true },
    triggerClass: { type: String, required: true },
    taskType: { type: String, required: true },
    loading: { type: Boolean, default: false }
  },
  emits: ["trigger"],
  setup(props, { emit }) {
    const row = (labelText: string, value: string, emptyText = "-", valueClass = "", dateTime = false) => {
      const displayValue = value || emptyText;
      return h("tr", [
        h("th", { class: "movie-fetch-status-label", title: labelText }, labelText),
        h(
          "td",
          { class: `movie-fetch-status-value ${valueClass}`.trim(), title: displayValue },
          dateTime ? [h(ScDateTime, { value: displayValue })] : displayValue
        )
      ]);
    };
    return () =>
      h("div", { class: "col-12 col-xl-6 mb-3" }, [
        h("div", { class: "card shadow-sm h-100 movie-section-card movie-fetch-status-card" }, [
          h("div", { class: "card-body" }, [
            h("div", { class: "d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3" }, [
              h("h5", { class: "card-title mb-0" }, props.title),
              h(
                "button",
                {
                  type: "button",
                  class: `btn btn-sm btn-outline-primary ${props.triggerClass}`,
                  disabled: props.loading,
                  onClick: () => emit("trigger", props.taskType)
                },
                props.loading ? "触发中" : `触发 ${props.title.replace(" 抓取状态", "")}`
              )
            ]),
            Object.keys(props.status || {}).length
              ? h("div", { class: "movie-resource-table-wrap movie-fetch-status-table-wrap" }, [
                  h(
                    "table",
                    {
                      class:
                        "table table-sm table-hover align-middle mb-0 movie-resource-table movie-fetch-status-table"
                    },
                    [
                      h("tbody", [
                        row("状态", text(props.status, "FetchStatusText")),
                        row("重试次数", text(props.status, "TryCount")),
                        row("最后抓取", text(props.status, "LastFetchTime"), "-", "", true),
                        row("最后抓取距离现在", text(props.status, "LastFetchAgo")),
                        row("Hash 数", text(props.status, "TorrentHashCount")),
                        row("最新日期", text(props.status, "LatestPublishTime"), "", "", true),
                        row("来源地址", text(props.status, "SourceURL"), "-", "movie-fetch-status-long"),
                        row("错误", text(props.status, "LastError"), "-", "movie-fetch-status-long")
                      ])
                    ]
                  )
                ])
              : h("div", { class: "text-muted" }, `暂无 ${props.title.replace(" 抓取状态", "")} 抓取状态`)
          ])
        ])
      ]);
  }
});
</script>
