<template>
  <PageShell
    class="movie-detail-shell fc2-page-shell fc2-detail-page-shell"
    :title="pageTitle"
    :note="pageNote"
    eyebrow="FC2 Movie Detail"
  >
    <div
      id="page-root"
      class="movie-detail-v2 movie-detail-page movie-detail-original-page fc2-page fc2-detail-page fc2-detail-movie-layout"
    >
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div class="movie-detail-top-shell fc2-detail-top-shell">
          <div class="movie-detail-titlebar fc2-detail-titlebar">
            <div class="movie-detail-title-main">
              <h4 class="mb-0 movie-detail-title" :title="movieText('Title')">
                {{ movieText("Title") || movieText("FC2Key") || pageTitle }}
              </h4>
            </div>
            <div class="movie-detail-title-meta">
              <span class="movie-title-meta-chip movie-title-meta-code">{{ movieText("FC2Key") || "-" }}</span>
              <span
                v-if="movieText('PerformerText')"
                class="movie-title-meta-chip movie-title-meta-cast"
                :title="movieText('PerformerText')"
                >{{ movieText("PerformerText") }}</span
              >
              <span class="movie-title-meta-chip movie-title-meta-mask">{{ movieText("MaskType") || "未知" }}</span>
              <span class="movie-title-meta-chip movie-title-meta-resource">磁力 {{ magnetResources.length }}</span>
              <span class="movie-title-meta-chip movie-title-meta-resource">帖子 {{ resources.length }}</span>
            </div>
          </div>

          <div class="row movie-original-top-row fc2-original-top-row">
            <div class="col-lg-7 mb-3 mb-lg-0">
              <div class="card shadow-sm movie-section-card movie-poster-card fc2-cover-card">
                <template v-if="primaryCoverImageID">
                  <a
                    class="movie-poster-frame fc2-detail-cover-frame"
                    :href="rowText(primaryCoverImage, 'ImageURL')"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <img
                      class="card-img-bottom img-fluid movie-poster-image fc2-detail-cover-image"
                      :src="rowText(primaryCoverImage, 'ImageURL')"
                      alt="5×5 详情联系表封面"
                    />
                    <div class="badge-stack">
                      <span class="badge-come">❤ {{ scStatText("ComeTimes") || "0" }}</span>
                      <span class="badge-times">× {{ scStatText("ScTimes") || "0" }}</span>
                    </div>
                    <span v-if="imageIsCover(primaryCoverImage)" class="badge-score fc2-cover-current-badge"
                      >当前封面</span
                    >
                  </a>
                  <div class="fc2-cover-toolbar">
                    <div class="fc2-cover-file">
                      <span>5×5 详情联系表封面</span>
                      <a
                        :href="rowText(primaryCoverImage, 'ImageURL')"
                        target="_blank"
                        rel="noopener noreferrer"
                        :title="rowText(primaryCoverImage, 'FilePath')"
                        >{{ imageFileName(primaryCoverImage) }}</a
                      >
                      <ScDateTime :value="rowValue(primaryCoverImage, 'CreatedOn')" unix />
                    </div>
                    <div class="fc2-capture-card-actions">
                      <button
                        v-if="imageCanBeCover(primaryCoverImage) && !imageIsCover(primaryCoverImage)"
                        type="button"
                        class="btn btn-sm btn-outline-primary"
                        :disabled="Boolean(coverBusyImageID) || Boolean(deleteBusyImageID)"
                        @click="setCoverImage(primaryCoverImage)"
                      >
                        {{ coverBusyImageID === primaryCoverImageID ? "设置中" : "设为封面" }}
                      </button>
                      <template v-if="deleteConfirmImageID === primaryCoverImageID">
                        <button
                          type="button"
                          class="btn btn-sm btn-danger"
                          :disabled="Boolean(coverBusyImageID) || deleteBusyImageID === primaryCoverImageID"
                          @click="deleteMediaImage(primaryCoverImage)"
                        >
                          {{ deleteBusyImageID === primaryCoverImageID ? "正在删除" : "确认删除" }}
                        </button>
                        <button
                          type="button"
                          class="btn btn-sm btn-outline-secondary"
                          :disabled="Boolean(coverBusyImageID) || deleteBusyImageID === primaryCoverImageID"
                          @click="cancelDeleteMediaImage(primaryCoverImage)"
                        >
                          取消
                        </button>
                      </template>
                      <button
                        v-else
                        type="button"
                        class="btn btn-sm btn-outline-danger"
                        :disabled="Boolean(coverBusyImageID) || Boolean(deleteBusyImageID)"
                        @click="requestDeleteMediaImage(primaryCoverImage)"
                      >
                        删除
                      </button>
                    </div>
                  </div>
                </template>
                <div v-else class="detail-poster-empty fc2-cover-empty">5×5 详情联系表封面尚未生成</div>
                <div v-if="captureFeedback || captureError" class="movie-action-feedback fc2-cover-feedback">
                  <span v-if="captureFeedback" class="movie-action-feedback-ok" role="status">{{
                    captureFeedback
                  }}</span>
                  <span v-if="captureError" class="movie-action-feedback-error" role="alert">{{ captureError }}</span>
                </div>
              </div>
            </div>

            <div class="col-lg-5">
              <div class="card shadow-sm movie-section-card movie-info-card fc2-info-card">
                <div class="card-body">
                  <ul class="nav nav-tabs nav-tabs-custom nav-justified movie-detail-tabs" role="tablist">
                    <li class="nav-item" role="presentation">
                      <button type="button" class="nav-link movie-detail-tab active" role="tab">Info</button>
                    </li>
                  </ul>

                  <div class="tab-content movie-detail-tab-content">
                    <div class="tab-pane p-3 active movie-info-pane" role="tabpanel">
                      <table
                        class="table table-bordered dt-responsive nowrap movie-info-table fc2-info-table"
                        style="width: 100%"
                      >
                        <tbody>
                          <tr class="table-light">
                            <th>识别码</th>
                            <td>
                              <div class="movie-info-code-line">
                                <span class="movie-info-code movie-tag-code">{{ movieText("FC2Key") || "-" }}</span
                                ><a
                                  v-if="mediaValue('ShowPlayButton')"
                                  class="btn btn-sm movie-action-button movie-action-button-play"
                                  :class="mediaText('PlayButtonClass')"
                                  :href="mediaText('VideoURL')"
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  >{{ mediaText("PlayButtonText") || "Play" }}</a
                                >
                              </div>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>FC2 编码</th>
                            <td>
                              <span class="movie-info-value">{{ movieText("Code") || "-" }}</span>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>影片名称</th>
                            <td>
                              <span class="fc2-info-title" :title="movieText('Title')">{{
                                movieText("Title") || "-"
                              }}</span>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>出演</th>
                            <td>
                              <span class="tag movie-tag-person">{{ movieText("PerformerText") || "-" }}</span>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>有码状态</th>
                            <td>
                              <span class="fc2-state-badge is-mask">{{ movieText("MaskType") || "未知" }}</span>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>数据状态</th>
                            <td>
                              <span class="fc2-state-badge" :class="statusClass(movieText('DataStatus'))">{{
                                movieText("DataStatus") || "-"
                              }}</span>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>发布时间</th>
                            <td>
                              <span class="movie-info-value movie-tag-date"
                                ><ScDateTime
                                  v-if="movieText('PostTime')"
                                  :value="movieText('PostTime')"
                                  unix
                                /><template v-else>-</template></span
                              >
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>资料来源</th>
                            <td>
                              <a
                                v-if="metadataResource"
                                class="movie-tag-link fc2-thread-link"
                                :href="`/shthread/${encodeURIComponent(rowText(metadataResource, 'ThreadTID'))}`"
                                >帖子 {{ rowText(metadataResource, "ThreadTID") }}</a
                              ><span v-else>-</span>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>媒体文件</th>
                            <td>
                              <span class="fc2-info-file" :title="mediaText('FileName')">{{
                                mediaText("FileName") || "-"
                              }}</span>
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
                                  id="btnOpenFC2MovieAlbumModal"
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
                                  >{{ album.name }}</a
                                >
                                <span v-if="!selectedAlbums.length" class="movie-info-empty">-</span>
                              </div>
                            </td>
                          </tr>
                          <tr class="table-light">
                            <th>导航</th>
                            <td>
                              <div class="movie-info-tag-group movie-info-links">
                                <a class="movie-tag-link" href="/fclist">FC2 列表</a
                                ><a class="movie-tag-link" href="/shpool" target="_blank" rel="noopener noreferrer"
                                  >SH 帖子池</a
                                >
                              </div>
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="hasMedia" class="row mt-3 movie-film-row">
          <div class="col-12">
            <div class="card shadow-sm movie-section-card movie-film-card fc2-film-card">
              <div class="card-body">
                <div class="movie-film-panel">
                  <div class="movie-film-head">
                    <h5 class="card-title movie-section-title">视频信息</h5>
                    <span class="movie-film-source">FC2Media</span>
                  </div>
                  <div class="movie-film-grid fc2-film-grid">
                    <div class="movie-film-item">
                      <span class="movie-film-label">Source</span><strong class="movie-film-value">FC2Media</strong>
                    </div>
                    <div class="movie-film-item movie-film-item-wide movie-film-hash">
                      <span class="movie-film-label">TorrentHash</span>
                      <div class="movie-film-value-line">
                        <strong class="text-monospace movie-film-value" :title="mediaText('SourceTorrentHash')">{{
                          mediaText("SourceTorrentHash") || "-"
                        }}</strong
                        ><button
                          v-if="mediaText('SourceTorrentHash')"
                          class="btn btn-sm btn-outline-secondary"
                          type="button"
                          @click="copyText(mediaText('SourceTorrentHash'), '来源 Hash')"
                        >
                          复制
                        </button>
                      </div>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">BirthTime</span
                      ><strong class="movie-film-value"><ScDateTime :value="mediaValue('BirthTime')" unix /></strong>
                    </div>
                    <div class="movie-film-item movie-film-size">
                      <span class="movie-film-label">Size</span>
                      <div class="movie-film-value-line">
                        <strong class="movie-film-value">{{ formatMediaSize(mediaValue("Size")) }}</strong>
                        <button
                          v-if="!mediaRemoved"
                          type="button"
                          class="btn btn-sm js-move-wmedia-removed"
                          :class="movieDeleteAlbumMarked ? 'btn-danger' : 'btn-outline-danger'"
                          :disabled="moveRemovedBusy"
                          @click="openMoveRemovedModal"
                        >
                          {{ movieDeleteAlbumMarked ? "已在待删除" : "加入待删除" }}
                        </button>
                      </div>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">Resolution</span
                      ><strong class="movie-film-value">{{ mediaResolution }}</strong>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">Duration</span
                      ><strong class="movie-film-value">{{ formatMediaDuration(mediaValue("Duration")) }}</strong>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">BitRate</span
                      ><strong class="movie-film-value">{{ formatMediaBitRate(mediaValue("BitRate")) }}</strong>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">Frame</span
                      ><strong class="movie-film-value">{{ formatMediaFrame(mediaValue("FrameAverage")) }}</strong>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">字幕</span
                      ><strong class="movie-film-value">{{ mediaText("HasSub") === "2" ? "有" : "无" }}</strong>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">自制</span
                      ><strong class="movie-film-value">{{ mediaText("SelfMake") === "2" ? "是" : "否" }}</strong>
                    </div>
                    <div class="movie-film-item">
                      <span class="movie-film-label">遮罩</span
                      ><strong class="movie-film-value">{{ mediaMaskText }}</strong>
                    </div>
                    <div class="movie-film-item movie-film-item-wide movie-film-dir">
                      <span class="movie-film-label">Dir</span>
                      <div class="movie-film-value-line">
                        <strong class="movie-film-value movie-film-path" :title="mediaPath">{{
                          mediaPath || "-"
                        }}</strong
                        ><button
                          v-if="mediaPath"
                          class="btn btn-sm btn-outline-secondary"
                          type="button"
                          @click="copyText(mediaPath, '媒体路径')"
                        >
                          复制
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="row mt-3 movie-resource-row fc2-magnet-resource-row">
          <div class="col-12 mb-3">
            <div class="card shadow-sm movie-section-card movie-resource-card fc2-resource-panel">
              <div class="card-body">
                <div class="movie-resource-heading-row fc2-resource-heading">
                  <h5 class="card-title mb-0 movie-section-title">资源列表</h5>
                  <div class="fc2-panel-actions">
                    <span v-if="resourceAlbumError" class="fc2-resource-album-error" role="alert">{{
                      resourceAlbumError
                    }}</span>
                    <span v-if="copyFeedback" class="fc2-copy-feedback" role="status">{{ copyFeedback }}</span>
                    <span class="movie-resource-count">{{ magnetResources.length }}</span>
                  </div>
                </div>
                <div class="table-responsive movie-resource-table-wrap fc2-table-wrap">
                  <table class="table table-hover align-middle mb-0 movie-resource-table fc2-table fc2-magnet-table">
                    <colgroup>
                      <col class="fc2-magnet-title" />
                      <col class="fc2-magnet-time" />
                      <col class="fc2-magnet-hash" />
                      <col class="fc2-magnet-albums" />
                    </colgroup>
                    <thead>
                      <tr>
                        <th>资源标题</th>
                        <th>发布时间</th>
                        <th>Info Hash / Magnet</th>
                        <th>相册</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="resource in magnetResources" :key="resourceRowID(resource)">
                        <td>
                          <div class="fc2-magnet-title-cell">
                            <a
                              v-if="rowText(resource, 'ThreadURL')"
                              class="fc2-magnet-title-link"
                              :href="rowText(resource, 'ThreadURL')"
                              target="_blank"
                              rel="noopener noreferrer"
                              :title="rowText(resource, 'ThreadTitle')"
                              >{{ rowText(resource, "ThreadTitle") || rowText(resource, "InfoHash") }}</a
                            ><span v-else :title="rowText(resource, 'ThreadTitle')">{{
                              rowText(resource, "ThreadTitle") || rowText(resource, "InfoHash") || "-"
                            }}</span
                            ><a
                              v-if="rowText(resource, 'ThreadTID')"
                              class="bucket-list-action"
                              :href="`/shthread/${encodeURIComponent(rowText(resource, 'ThreadTID'))}`"
                              >帖子详情</a
                            >
                          </div>
                        </td>
                        <td><ScDateTime :value="rowValue(resource, 'PostTime')" unix /></td>
                        <td>
                          <div class="fc2-copy-cell">
                            <span class="fc2-mono" :title="rowText(resource, 'InfoHash')">{{
                              rowText(resource, "InfoHash") || "-"
                            }}</span
                            ><button
                              v-if="rowText(resource, 'InfoHash')"
                              type="button"
                              class="fc2-copy-button"
                              @click="copyText(rowText(resource, 'InfoHash'), `Info Hash ${resourceRowID(resource)}`)"
                            >
                              Hash
                            </button>
                            <button
                              v-if="rowText(resource, 'MagnetURL')"
                              type="button"
                              class="fc2-copy-button"
                              @click="copyText(rowText(resource, 'MagnetURL'), `Magnet ${resourceRowID(resource)}`)"
                            >
                              Magnet
                            </button>
                          </div>
                        </td>
                        <td>
                          <div class="fc2-resource-album-actions">
                            <button
                              v-for="albumName in ['下载中', '待下载']"
                              :key="albumName"
                              type="button"
                              class="fc2-resource-album-state"
                              :class="resourceAlbumButtonClass(resource, albumName)"
                              :aria-pressed="isResourceInAlbum(resource, albumName) ? 'true' : 'false'"
                              :title="resourceAlbumButtonTitle(resource, albumName)"
                              :disabled="resourceAlbumBusyKey === resourceAlbumKey(resource, albumName)"
                              @click="switchResourceAlbum(resource, albumName)"
                            >
                              {{ resourceAlbumButtonText(resource, albumName) }}
                            </button>
                          </div>
                        </td>
                      </tr>
                      <tr v-if="!magnetResources.length">
                        <td colspan="4"><div class="sc-empty-state fc2-empty-state">当前影片没有磁力资源。</div></td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="row movie-resource-row">
          <div class="col-12 mb-3">
            <div class="card shadow-sm movie-section-card movie-resource-card fc2-resource-panel">
              <div class="card-body">
                <div class="movie-resource-heading-row fc2-resource-heading">
                  <h5 class="card-title mb-0 movie-section-title">帖子资源</h5>
                  <div class="fc2-panel-actions">
                    <span v-if="copyFeedback" class="fc2-copy-feedback" role="status">{{ copyFeedback }}</span>
                    <span class="movie-resource-count">{{ resources.length }}</span>
                    <button
                      class="bucket-list-action"
                      type="button"
                      :disabled="!allHashes.length"
                      @click="copyText(allHashes.join('\n'), '全部 Info Hash')"
                    >
                      复制全部 Hash
                    </button>
                    <button
                      class="bucket-list-action"
                      type="button"
                      :disabled="!allMagnets.length"
                      @click="copyText(allMagnets.join('\n'), '全部 Magnet')"
                    >
                      复制全部 Magnet
                    </button>
                  </div>
                </div>
                <div class="table-responsive movie-resource-table-wrap fc2-table-wrap">
                  <table class="table table-hover align-middle mb-0 movie-resource-table fc2-table fc2-resource-table">
                    <colgroup>
                      <col class="fc2-resource-time" />
                      <col class="fc2-resource-size" />
                      <col class="fc2-resource-hash" />
                      <col class="fc2-resource-magnet" />
                      <col class="fc2-resource-attachment" />
                      <col class="fc2-resource-action" />
                    </colgroup>
                    <thead>
                      <tr>
                        <th>发布时间</th>
                        <th>容量</th>
                        <th>Info Hash</th>
                        <th>Magnet</th>
                        <th>附件数</th>
                        <th>操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      <template v-for="resource in resources" :key="rowText(resource, 'ThreadTID')">
                        <tr>
                          <td><ScDateTime :value="rowValue(resource, 'PostTime')" unix /></td>
                          <td>{{ metaText(resource, "size_text") || "-" }}</td>
                          <td>
                            <div class="fc2-copy-cell">
                              <span class="fc2-mono" :title="rowText(resource, 'InfoHashText')">{{
                                compactLines(rowText(resource, "InfoHashText")) || "-"
                              }}</span
                              ><button
                                v-if="rowText(resource, 'InfoHashText')"
                                type="button"
                                class="fc2-copy-button"
                                @click="
                                  copyText(
                                    rowText(resource, 'InfoHashText'),
                                    `Info Hash ${rowText(resource, 'ThreadTID')}`
                                  )
                                "
                              >
                                复制
                              </button>
                            </div>
                          </td>
                          <td>
                            <div class="fc2-copy-cell">
                              <span class="fc2-mono" :title="rowText(resource, 'MagnetText')">{{
                                compactLines(rowText(resource, "MagnetText")) || "-"
                              }}</span
                              ><button
                                v-if="rowText(resource, 'MagnetText')"
                                type="button"
                                class="fc2-copy-button"
                                @click="
                                  copyText(rowText(resource, 'MagnetText'), `Magnet ${rowText(resource, 'ThreadTID')}`)
                                "
                              >
                                复制
                              </button>
                            </div>
                          </td>
                          <td>
                            <button
                              v-if="attachments(resource).length"
                              type="button"
                              class="fc2-attachment-toggle"
                              :aria-expanded="expandedThreadTID === rowText(resource, 'ThreadTID')"
                              @click="toggleAttachments(resource)"
                            >
                              {{ attachments(resource).length }}
                              {{ expandedThreadTID === rowText(resource, "ThreadTID") ? "收起" : "展开" }}</button
                            ><span v-else>0</span>
                          </td>
                          <td>
                            <div class="fc2-row-actions">
                              <a
                                class="bucket-list-action"
                                :href="`/shthread/${encodeURIComponent(rowText(resource, 'ThreadTID'))}`"
                                >帖子详情</a
                              ><a
                                class="bucket-list-action"
                                :href="rowText(resource, 'ThreadURL')"
                                target="_blank"
                                rel="noopener noreferrer"
                                >原帖</a
                              >
                            </div>
                          </td>
                        </tr>
                        <tr v-if="expandedThreadTID === rowText(resource, 'ThreadTID')" class="fc2-attachment-row">
                          <td colspan="6">
                            <table class="fc2-attachment-table">
                              <thead>
                                <tr>
                                  <th>文件名</th>
                                  <th>类型</th>
                                  <th>容量</th>
                                  <th>下载次数</th>
                                  <th>上传时间</th>
                                </tr>
                              </thead>
                              <tbody>
                                <tr
                                  v-for="(attachment, index) in attachments(resource)"
                                  :key="`${rowText(attachment, 'file_name')}-${index}`"
                                >
                                  <td :title="rowText(attachment, 'file_name')">
                                    {{ rowText(attachment, "file_name") }}
                                  </td>
                                  <td>{{ rowText(attachment, "media_type") || "-" }}</td>
                                  <td>{{ rowText(attachment, "size_text") || "-" }}</td>
                                  <td>{{ rowText(attachment, "download_count") || "0" }}</td>
                                  <td>{{ rowText(attachment, "upload_time_text") || "-" }}</td>
                                </tr>
                              </tbody>
                            </table>
                          </td>
                        </tr>
                      </template>
                      <tr v-if="!resources.length">
                        <td colspan="6"><div class="sc-empty-state fc2-empty-state">当前影片没有帖子资源。</div></td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div
          class="modal fade"
          id="fc2MovieAlbumModal"
          tabindex="-1"
          :class="{ show: showMovieAlbumModal }"
          :aria-hidden="showMovieAlbumModal ? 'false' : 'true'"
        >
          <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
              <div class="modal-header">
                <h5 class="modal-title">加入电影相册</h5>
                <button
                  type="button"
                  class="btn-close"
                  aria-label="Close"
                  @click="showMovieAlbumModal = false"
                ></button>
              </div>
              <div class="modal-body">
                <div
                  v-if="movieAlbumMessage"
                  id="fc2MovieAlbumMsg"
                  class="alert mb-3"
                  :class="movieAlbumMessageOk ? 'alert-success' : 'alert-danger'"
                  role="alert"
                >
                  {{ movieAlbumMessage }}
                </div>
                <div class="mb-3">
                  <label class="form-label mb-1" for="fc2MovieAlbumSelect">选择相册</label>
                  <select
                    v-model="selectedMovieAlbumName"
                    class="form-select"
                    id="fc2MovieAlbumSelect"
                    :disabled="movieAlbumBusy"
                  >
                    <option v-for="album in albumOptions" :key="album.name" :value="album.name">
                      {{ album.name }}{{ album.selected ? "（已在相册）" : "" }}
                    </option>
                  </select>
                </div>
                <div>
                  <label class="form-label mb-1" for="fc2MovieAlbumNewName">新建相册</label>
                  <div class="d-flex gap-2">
                    <input
                      v-model="newMovieAlbumName"
                      type="text"
                      class="form-control"
                      id="fc2MovieAlbumNewName"
                      maxlength="128"
                      placeholder="输入新相册名称后创建"
                      :disabled="movieAlbumBusy"
                    />
                    <button
                      type="button"
                      class="btn btn-outline-secondary"
                      id="btnCreateFC2MovieAlbumInline"
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
                  id="btnAddFC2MovieToAlbum"
                  :disabled="movieAlbumBusy || !selectedMovieAlbumName"
                  @click="addMovieToAlbum"
                >
                  {{ movieAlbumBusy ? "处理中" : "加入" }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <div
          class="modal fade"
          id="moveFC2MediaRemovedModal"
          tabindex="-1"
          aria-labelledby="moveFC2MediaRemovedModalLabel"
          :class="{ show: showMoveRemovedModal }"
          :aria-hidden="showMoveRemovedModal ? 'false' : 'true'"
        >
          <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
              <div class="modal-header">
                <h5 class="modal-title" id="moveFC2MediaRemovedModalLabel">加入待删除</h5>
                <button
                  type="button"
                  class="btn-close"
                  aria-label="Close"
                  @click="showMoveRemovedModal = false"
                ></button>
              </div>
              <div class="modal-body">
                将当前FC2影片加入默认电影相册“待删除”。真正删除会在电影相册页统一执行。
                <div
                  v-if="moveRemovedMessage"
                  class="small mt-2"
                  :class="moveRemovedMessageOk ? 'text-success' : 'text-danger'"
                >
                  {{ moveRemovedMessage }}
                </div>
              </div>
              <div class="modal-footer">
                <button
                  type="button"
                  class="btn btn-light"
                  :disabled="moveRemovedBusy"
                  @click="showMoveRemovedModal = false"
                >
                  取消
                </button>
                <button
                  type="button"
                  class="btn btn-danger"
                  id="btnMoveFC2MediaRemovedDo"
                  :disabled="moveRemovedBusy"
                  @click="addToDeleteAlbum"
                >
                  {{ moveRemovedBusy ? "处理中" : "加入待删除" }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/features/fc2/api";
import PageShell from "@/shared/components/PageShell.vue";
import ScDateTime from "@/shared/components/ScDateTime.vue";
import type { FC2Record } from "@/features/fc2/types";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";

const route = useRoute();
const payload = ref<FC2Record>({});
const loading = ref(false);
const error = ref("");
const expandedThreadTID = ref("");
const copyFeedback = ref("");
const captureFeedback = ref("");
const captureError = ref("");
const coverBusyImageID = ref("");
const deleteConfirmImageID = ref("");
const deleteBusyImageID = ref("");
const showMovieAlbumModal = ref(false);
const selectedMovieAlbumName = ref("");
const newMovieAlbumName = ref("");
const movieAlbumBusy = ref(false);
const movieAlbumMessage = ref("");
const movieAlbumMessageOk = ref(true);
const showMoveRemovedModal = ref(false);
const moveRemovedBusy = ref(false);
const moveRemovedMessage = ref("");
const moveRemovedMessageOk = ref(true);
const resourceAlbumBusyKey = ref("");
const resourceAlbumError = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  expandedThreadTID.value = "";
  copyFeedback.value = "";
  captureFeedback.value = "";
  captureError.value = "";
  coverBusyImageID.value = "";
  deleteConfirmImageID.value = "";
  deleteBusyImageID.value = "";
  showMovieAlbumModal.value = false;
  movieAlbumMessage.value = "";
  showMoveRemovedModal.value = false;
  moveRemovedBusy.value = false;
  moveRemovedMessage.value = "";
  resourceAlbumBusyKey.value = "";
  resourceAlbumError.value = "";
  try {
    await refreshPayload();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

async function refreshPayload(): Promise<void> {
  payload.value = asRecord(await getJSON(apiPathForRoute(route)));
}

const pageTitle = computed(() => text(payload.value, "PageTitle", "Title") || "FC2影片详情");
const pageNote = computed(() => text(payload.value, "PageNote"));
const movie = computed(() => asRecord(pick(payload.value, "Movie", "movie")));
const media = computed(() => asRecord(pick(payload.value, "Media", "media")));
const scStat = computed(() => asRecord(pick(payload.value, "ScStat", "scStat")));
const scRows = computed(() =>
  asArray<FC2Record>(pick(payload.value, "SC", "sc"))
    .map((row) => ({
      name: text(row, "ScName"),
      href: text(row, "Href")
    }))
    .filter((row) => row.name)
);
const hasMedia = computed(() => Boolean(mediaText("ID")));
const primaryCoverImage = computed<FC2Record>(() =>
  asRecord(pick(payload.value, "DetailCoverImage", "detailCoverImage"))
);
const primaryCoverImageID = computed(() => imageID(primaryCoverImage.value));
const magnetResources = computed(() => asArray<FC2Record>(pick(payload.value, "MagnetResources", "magnetResources")));
const resources = computed(() => asArray<FC2Record>(pick(payload.value, "Resources", "resources")));
const albums = computed(() => asArray<FC2Record>(pick(payload.value, "MovieAlbums", "movieAlbums")));
const albumOptions = computed(() =>
  albums.value
    .map((album) => ({
      name: text(album, "Name", "name"),
      selected: Boolean(pick(album, "Selected", "selected"))
    }))
    .filter((album) => album.name)
);
const selectedAlbums = computed(() => albumOptions.value.filter((album) => album.selected));
const movieDeleteAlbumMarked = computed(() =>
  albumOptions.value.some((album) => album.name === "待删除" && album.selected)
);
const mediaRemoved = computed(() => Number(mediaValue("IsRemoved") || 0) === 2);
const metadataResource = computed(() =>
  resources.value.find((resource) => rowText(resource, "ID") === movieText("MetadataThreadID"))
);
const allHashes = computed(() =>
  uniqueLines(resources.value.flatMap((resource) => lines(rowText(resource, "InfoHashText"))))
);
const allMagnets = computed(() =>
  uniqueLines(resources.value.flatMap((resource) => lines(rowText(resource, "MagnetText"))))
);
const mediaPath = computed(() =>
  [mediaText("FullDir").replace(/\/$/, ""), mediaText("FileName")].filter(Boolean).join("/")
);
const mediaResolution = computed(() => {
  const width = Number(mediaValue("Width") || 0);
  const height = Number(mediaValue("Height") || 0);
  return width > 0 && height > 0 ? `${width} × ${height}` : "-";
});
const mediaMaskText = computed(() =>
  mediaText("HasMask") === "2" ? "去码" : mediaText("HasMask") === "3" ? "无码" : "有码"
);

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

function rowValue(row: FC2Record, key: string): unknown {
  return pick(row, key);
}
function rowText(row: FC2Record, key: string): string {
  return text(row, key);
}
function resourceRowID(row: FC2Record): number {
  return Number(text(row, "RowID", "row_id", "Id", "ID", "id") || 0);
}
function resourceAlbumKey(row: FC2Record, albumName: string): string {
  return `${resourceRowID(row)}:${albumName}`;
}
function isResourceInAlbum(row: FC2Record, albumName: string): boolean {
  return albumName === "待下载"
    ? Boolean(pick(row, "IsInPending", "is_in_pending", "InPending"))
    : Boolean(pick(row, "IsInDownload", "is_in_download", "InDownload"));
}
function setResourceAlbumState(row: FC2Record, albumName: string, favorited: boolean): void {
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
function resourceAlbumButtonClass(row: FC2Record, albumName: string): string[] {
  return [
    albumName === "待下载" ? "is-pending" : "is-downloading",
    isResourceInAlbum(row, albumName) ? "is-selected" : "is-idle"
  ];
}
function resourceAlbumButtonText(row: FC2Record, albumName: string): string {
  return isResourceInAlbum(row, albumName) ? `已在${albumName}` : albumName;
}
function resourceAlbumButtonTitle(row: FC2Record, albumName: string): string {
  return isResourceInAlbum(row, albumName) ? `移出${albumName}` : `加入${albumName}`;
}
async function switchResourceAlbum(row: FC2Record, albumName: string): Promise<void> {
  const fc2Key = movieText("FC2Key");
  const sourceRowID = resourceRowID(row);
  const favorited = isResourceInAlbum(row, albumName);
  if (!fc2Key || !Number.isInteger(sourceRowID) || sourceRowID <= 0) {
    resourceAlbumError.value = `缺少${albumName}参数，无法执行`;
    return;
  }
  const busyKey = resourceAlbumKey(row, albumName);
  resourceAlbumBusyKey.value = busyKey;
  resourceAlbumError.value = "";
  try {
    const data = asRecord(
      await sendJSON(
        `/api/fc2/movies/${encodeURIComponent(fc2Key)}/resource-album-item`,
        favorited ? "DELETE" : "POST",
        { source_row_id: sourceRowID, album_name: albumName }
      )
    );
    setResourceAlbumState(row, albumName, typeof data.favorited === "boolean" ? data.favorited : !favorited);
  } catch (err) {
    resourceAlbumError.value = err instanceof Error ? err.message : `${albumName}状态更新失败`;
  } finally {
    resourceAlbumBusyKey.value = "";
  }
}
function movieText(key: string): string {
  return text(movie.value, key);
}
function mediaValue(key: string): unknown {
  return pick(media.value, key);
}
function mediaText(key: string): string {
  return text(media.value, key);
}
function scStatText(key: string): string {
  return text(scStat.value, key);
}
function resourceMeta(resource: FC2Record): FC2Record {
  return asRecord(pick(resource, "Meta", "meta"));
}
function metaText(resource: FC2Record, key: string): string {
  return text(resourceMeta(resource), key);
}
function attachments(resource: FC2Record): FC2Record[] {
  return asArray<FC2Record>(pick(resourceMeta(resource), "attachments", "Attachments"));
}
function lines(value: string): string[] {
  return value
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean);
}
function uniqueLines(values: string[]): string[] {
  return [...new Set(values)];
}
function compactLines(value: string): string {
  return lines(value).join(" · ");
}
function statusClass(status: string): string {
  return status === "conflict" ? "is-conflict" : status === "ready" ? "is-ready" : "is-neutral";
}
function imageTypeText(image: FC2Record): string {
  return rowText(image, "ImageType") === "detail_contact_sheet" ? "5×5 详情联系表封面" : "2×5 卡片联系表封面";
}
function imageFileName(image: FC2Record): string {
  return rowText(image, "FilePath").split(/[\\/]/).pop() || "打开原图";
}
function imageID(image: FC2Record): string {
  return rowText(image, "ID").trim();
}
function imageCanBeCover(image: FC2Record): boolean {
  return rowText(image, "ImageType") === "contact_sheet";
}
function imageIsCover(image: FC2Record): boolean {
  return pick(image, "IsCover") === true;
}
function formatMediaSize(value: unknown): string {
  const bytes = Number(value || 0);
  return bytes > 0 ? `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB` : "-";
}
function formatMediaDuration(value: unknown): string {
  const seconds = Number(value || 0);
  if (seconds <= 0) return "-";
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remaining = Math.floor(seconds % 60);
  return [hours, minutes, remaining].map((part) => String(part).padStart(2, "0")).join(":");
}
function formatMediaBitRate(value: unknown): string {
  const bitRate = Number(value || 0);
  return bitRate > 0 ? `${Math.round(bitRate / 1000)} kbps` : "-";
}
function formatMediaFrame(value: unknown): string {
  const frame = Number(value || 0);
  return frame > 0 ? frame.toFixed(2) : "-";
}
function toggleAttachments(resource: FC2Record) {
  const threadTID = rowText(resource, "ThreadTID");
  expandedThreadTID.value = expandedThreadTID.value === threadTID ? "" : threadTID;
}
function openMovieAlbumModal() {
  movieAlbumMessage.value = "";
  newMovieAlbumName.value = "";
  if (!selectedMovieAlbumName.value && albumOptions.value.length) {
    selectedMovieAlbumName.value = albumOptions.value[0].name;
  }
  showMovieAlbumModal.value = true;
}
function openMoveRemovedModal() {
  moveRemovedMessage.value = "";
  showMoveRemovedModal.value = true;
}
async function addToDeleteAlbum() {
  const fc2Key = movieText("FC2Key");
  if (!fc2Key) {
    moveRemovedMessage.value = "缺少FC2影片标识";
    moveRemovedMessageOk.value = false;
    return;
  }
  moveRemovedBusy.value = true;
  moveRemovedMessage.value = "";
  try {
    const data = asRecord(
      await sendJSON(`/api/fc2/movies/${encodeURIComponent(fc2Key)}/movie-album-item`, "POST", {
        fc2_key: fc2Key,
        album_name: "待删除"
      })
    );
    const albumName = text(data, "album_name", "AlbumName") || "待删除";
    upsertMovieAlbumOption(albumName, true);
    moveRemovedMessage.value = text(data, "message") || "已加入待删除相册";
    moveRemovedMessageOk.value = true;
    await refreshPayload();
    showMoveRemovedModal.value = false;
  } catch (err) {
    moveRemovedMessage.value = err instanceof Error ? err.message : "加入待删除相册失败";
    moveRemovedMessageOk.value = false;
  } finally {
    moveRemovedBusy.value = false;
  }
}
function upsertMovieAlbumOption(name: string, selected: boolean) {
  const nextPayload: FC2Record = { ...payload.value };
  const key = Array.isArray(nextPayload.MovieAlbums) ? "MovieAlbums" : "movieAlbums";
  const rows = asArray<FC2Record>(nextPayload[key]).map((album) => ({ ...album }));
  const found = rows.find((album) => text(album, "Name", "name") === name);
  if (found) {
    found.Name = name;
    found.name = name;
    found.Selected = selected || Boolean(pick(found, "Selected", "selected"));
    found.selected = found.Selected;
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
    movieAlbumMessage.value = err instanceof Error ? err.message : "电影相册创建失败";
    movieAlbumMessageOk.value = false;
  } finally {
    movieAlbumBusy.value = false;
  }
}
async function addMovieToAlbum() {
  const fc2Key = movieText("FC2Key");
  const albumName = selectedMovieAlbumName.value.trim();
  if (!fc2Key || !albumName) {
    movieAlbumMessage.value = !fc2Key ? "缺少FC2影片标识" : "请选择电影相册";
    movieAlbumMessageOk.value = false;
    return;
  }
  movieAlbumBusy.value = true;
  movieAlbumMessage.value = "";
  try {
    const data = asRecord(
      await sendJSON(`/api/fc2/movies/${encodeURIComponent(fc2Key)}/movie-album-item`, "POST", {
        fc2_key: fc2Key,
        album_name: albumName
      })
    );
    const responseAlbumName = text(data, "album_name", "AlbumName") || albumName;
    upsertMovieAlbumOption(responseAlbumName, true);
    movieAlbumMessage.value = text(data, "message") || "已加入电影相册";
    movieAlbumMessageOk.value = true;
    await refreshPayload();
    showMovieAlbumModal.value = false;
  } catch (err) {
    movieAlbumMessage.value = err instanceof Error ? err.message : "加入电影相册失败";
    movieAlbumMessageOk.value = false;
  } finally {
    movieAlbumBusy.value = false;
  }
}
async function setCoverImage(image: FC2Record) {
  const id = imageID(image);
  if (!id || !imageCanBeCover(image) || imageIsCover(image) || coverBusyImageID.value || deleteBusyImageID.value)
    return;
  coverBusyImageID.value = id;
  deleteConfirmImageID.value = "";
  captureFeedback.value = "";
  captureError.value = "";
  try {
    const fc2Key = movieText("FC2Key");
    const response = asRecord(
      await sendJSON(`/api/fc2/movies/${encodeURIComponent(fc2Key)}/images/${encodeURIComponent(id)}/cover`, "PUT")
    );
    if (pick(response, "ok") !== true) throw new Error("接口未确认封面设置结果");
    await refreshPayload();
    captureFeedback.value = "封面设置完成";
  } catch (err) {
    captureError.value = err instanceof Error ? err.message : String(err);
  } finally {
    coverBusyImageID.value = "";
  }
}
function requestDeleteMediaImage(image: FC2Record) {
  const id = imageID(image);
  if (!id || coverBusyImageID.value || deleteBusyImageID.value) return;
  deleteConfirmImageID.value = id;
  captureFeedback.value = "";
  captureError.value = "";
}
function cancelDeleteMediaImage(image: FC2Record) {
  const id = imageID(image);
  if (deleteBusyImageID.value === id) return;
  if (deleteConfirmImageID.value === id) deleteConfirmImageID.value = "";
}
async function deleteMediaImage(image: FC2Record) {
  const id = imageID(image);
  if (!id || deleteConfirmImageID.value !== id || coverBusyImageID.value || deleteBusyImageID.value) return;
  deleteBusyImageID.value = id;
  captureFeedback.value = "";
  captureError.value = "";
  try {
    const fc2Key = movieText("FC2Key");
    const response = asRecord(
      await sendJSON(`/api/fc2/movies/${encodeURIComponent(fc2Key)}/images/${encodeURIComponent(id)}`, "DELETE")
    );
    if (pick(response, "ok") !== true) throw new Error("接口未确认图片删除结果");
    await refreshPayload();
    deleteConfirmImageID.value = "";
    captureFeedback.value = `${imageTypeText(image)}已移动到删除目录`;
  } catch (err) {
    captureError.value = err instanceof Error ? err.message : String(err);
  } finally {
    deleteBusyImageID.value = "";
  }
}
async function copyText(value: string, label: string) {
  if (!value.trim()) return;
  try {
    await navigator.clipboard.writeText(value.trim());
    copyFeedback.value = `${label} 已复制`;
  } catch {
    const input = document.createElement("textarea");
    input.value = value.trim();
    input.setAttribute("readonly", "");
    input.style.position = "fixed";
    input.style.opacity = "0";
    document.body.appendChild(input);
    input.select();
    const copied = document.execCommand("copy");
    input.remove();
    copyFeedback.value = copied ? `${label} 已复制` : `${label} 复制失败`;
  }
}
</script>
