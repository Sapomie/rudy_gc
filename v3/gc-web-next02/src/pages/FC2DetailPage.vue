<template>
  <PageShell class="movie-detail-shell fc2-page-shell fc2-detail-page-shell" :title="pageTitle" :note="pageNote" eyebrow="FC2 Movie Detail">
    <div id="page-root" class="movie-detail-v2 movie-detail-page movie-detail-original-page fc2-page fc2-detail-page fc2-detail-movie-layout">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <div class="movie-detail-top-shell fc2-detail-top-shell">
          <div class="movie-detail-titlebar fc2-detail-titlebar">
            <div class="movie-detail-title-main">
              <h4 class="mb-0 movie-detail-title" :title="movieText('Title')">{{ movieText("Title") || movieText("FC2Key") || pageTitle }}</h4>
            </div>
            <div class="movie-detail-title-meta">
              <span class="movie-title-meta-chip movie-title-meta-code">{{ movieText("FC2Key") || "-" }}</span>
              <span v-if="movieText('PerformerText')" class="movie-title-meta-chip movie-title-meta-cast" :title="movieText('PerformerText')">{{ movieText("PerformerText") }}</span>
              <span class="movie-title-meta-chip movie-title-meta-mask">{{ movieText("MaskType") || "未知" }}</span>
              <span class="movie-title-meta-chip movie-title-meta-resource">资源 {{ resources.length }}</span>
            </div>
          </div>

          <div class="row movie-original-top-row fc2-original-top-row">
            <div class="col-lg-7 mb-3 mb-lg-0">
              <div class="card shadow-sm movie-section-card movie-poster-card fc2-cover-card">
                <template v-if="primaryCoverImageID">
                  <a class="movie-poster-frame fc2-detail-cover-frame" :href="rowText(primaryCoverImage, 'ImageURL')" target="_blank" rel="noopener noreferrer">
                    <img class="card-img-bottom img-fluid movie-poster-image fc2-detail-cover-image" :src="rowText(primaryCoverImage, 'ImageURL')" alt="默认联系表封面">
                    <div class="badge-stack">
                      <span class="badge-come">❤ {{ scStatText("ComeTimes") || "0" }}</span>
                      <span class="badge-times">× {{ scStatText("ScTimes") || "0" }}</span>
                    </div>
                    <span v-if="imageIsCover(primaryCoverImage)" class="badge-score fc2-cover-current-badge">当前封面</span>
                  </a>
                  <div class="fc2-cover-toolbar">
                    <div class="fc2-cover-file">
                      <span>默认联系表封面</span>
                      <a :href="rowText(primaryCoverImage, 'ImageURL')" target="_blank" rel="noopener noreferrer" :title="rowText(primaryCoverImage, 'FilePath')">{{ imageFileName(primaryCoverImage) }}</a>
                      <ScDateTime :value="rowValue(primaryCoverImage, 'CreatedOn')" unix />
                    </div>
                    <div class="fc2-capture-card-actions">
                      <button
                        v-if="imageCanBeCover(primaryCoverImage) && !imageIsCover(primaryCoverImage)"
                        type="button"
                        class="btn btn-sm btn-outline-primary"
                        :disabled="Boolean(coverBusyImageID) || Boolean(deleteBusyImageID)"
                        @click="setCoverImage(primaryCoverImage)"
                      >{{ coverBusyImageID === primaryCoverImageID ? "设置中" : "设为封面" }}</button>
                      <template v-if="deleteConfirmImageID === primaryCoverImageID">
                        <button type="button" class="btn btn-sm btn-danger" :disabled="Boolean(coverBusyImageID) || deleteBusyImageID === primaryCoverImageID" @click="deleteMediaImage(primaryCoverImage)">{{ deleteBusyImageID === primaryCoverImageID ? "正在删除" : "确认删除" }}</button>
                        <button type="button" class="btn btn-sm btn-outline-secondary" :disabled="Boolean(coverBusyImageID) || deleteBusyImageID === primaryCoverImageID" @click="cancelDeleteMediaImage(primaryCoverImage)">取消</button>
                      </template>
                      <button v-else type="button" class="btn btn-sm btn-outline-danger" :disabled="Boolean(coverBusyImageID) || Boolean(deleteBusyImageID)" @click="requestDeleteMediaImage(primaryCoverImage)">删除</button>
                    </div>
                  </div>
                </template>
                <div v-else class="detail-poster-empty fc2-cover-empty">默认联系表封面尚未生成</div>
                <div v-if="captureFeedback || captureError" class="movie-action-feedback fc2-cover-feedback">
                  <span v-if="captureFeedback" class="movie-action-feedback-ok" role="status">{{ captureFeedback }}</span>
                  <span v-if="captureError" class="movie-action-feedback-error" role="alert">{{ captureError }}</span>
                </div>
              </div>
            </div>

            <div class="col-lg-5">
              <div class="card shadow-sm movie-section-card movie-info-card fc2-info-card">
                <div class="card-body">
                  <ul class="nav nav-tabs nav-tabs-custom nav-justified movie-detail-tabs" role="tablist">
                    <li class="nav-item" role="presentation"><button type="button" class="nav-link movie-detail-tab" :class="{ active: activeDetailTab === 'info' }" role="tab" @click="activeDetailTab = 'info'">Info</button></li>
                    <li class="nav-item" role="presentation"><button type="button" class="nav-link movie-detail-tab" :class="{ active: activeDetailTab === 'sc' }" role="tab" @click="activeDetailTab = 'sc'">SC</button></li>
                  </ul>

                  <div class="tab-content movie-detail-tab-content">
                    <div v-show="activeDetailTab === 'info'" class="tab-pane p-3 active movie-info-pane" role="tabpanel">
                      <table class="table table-bordered dt-responsive nowrap movie-info-table fc2-info-table" style="width:100%">
                        <tbody>
                          <tr class="table-light">
                            <th>识别码</th>
                            <td><div class="movie-info-code-line"><span class="movie-info-code movie-tag-code">{{ movieText("FC2Key") || "-" }}</span><a v-if="mediaValue('ShowPlayButton')" class="btn btn-sm movie-action-button movie-action-button-play" :class="mediaText('PlayButtonClass')" :href="mediaText('VideoURL')" target="_blank" rel="noopener noreferrer">{{ mediaText("PlayButtonText") || "Play" }}</a></div></td>
                          </tr>
                          <tr class="table-light"><th>FC2 编码</th><td><span class="movie-info-value">{{ movieText("Code") || "-" }}</span></td></tr>
                          <tr class="table-light"><th>影片名称</th><td><span class="fc2-info-title" :title="movieText('Title')">{{ movieText("Title") || "-" }}</span></td></tr>
                          <tr class="table-light"><th>出演</th><td><span class="tag movie-tag-person">{{ movieText("PerformerText") || "-" }}</span></td></tr>
                          <tr class="table-light"><th>有码状态</th><td><span class="fc2-state-badge is-mask">{{ movieText("MaskType") || "未知" }}</span></td></tr>
                          <tr class="table-light"><th>数据状态</th><td><span class="fc2-state-badge" :class="statusClass(movieText('DataStatus'))">{{ movieText("DataStatus") || "-" }}</span></td></tr>
                          <tr class="table-light"><th>发布时间</th><td><span class="movie-info-value movie-tag-date"><ScDateTime v-if="movieText('PostTime')" :value="movieText('PostTime')" unix /><template v-else>-</template></span></td></tr>
                          <tr class="table-light"><th>资料来源</th><td><a v-if="metadataResource" class="movie-tag-link fc2-thread-link" :href="`/shthread/${encodeURIComponent(rowText(metadataResource, 'ThreadTID'))}`">帖子 {{ rowText(metadataResource, "ThreadTID") }}</a><span v-else>-</span></td></tr>
                          <tr class="table-light"><th>媒体文件</th><td><span class="fc2-info-file" :title="mediaText('FileName')">{{ mediaText("FileName") || "-" }}</span></td></tr>
                          <tr class="table-light"><th>导航</th><td><div class="movie-info-tag-group movie-info-links"><a class="movie-tag-link" href="/fclist">FC2 列表</a><a class="movie-tag-link" href="/shpool" target="_blank" rel="noopener noreferrer">SH 帖子池</a></div></td></tr>
                        </tbody>
                      </table>
                    </div>

                    <div v-show="activeDetailTab === 'sc'" class="tab-pane p-3 active movie-rank-panel fc2-sc-panel" role="tabpanel">
                      <div class="fc2-sc-detail-grid">
                        <div><span>SC 次数</span><strong>{{ scStatText("ScTimes") || "0" }}</strong></div>
                        <div><span>Come 次数</span><strong>{{ scStatText("ComeTimes") || "0" }}</strong></div>
                        <div><span>最近 SC</span><strong><ScDateTime v-if="scStatValue('LastScTime')" :value="scStatValue('LastScTime')" unix /><template v-else>-</template></strong></div>
                        <div><span>最近事件</span><strong><a v-if="lastScEventHref" :href="lastScEventHref">{{ lastScEventName }}</a><template v-else>-</template></strong></div>
                      </div>
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
                    <div class="movie-film-item"><span class="movie-film-label">Source</span><strong class="movie-film-value">FC2Media</strong></div>
                    <div class="movie-film-item movie-film-item-wide movie-film-hash"><span class="movie-film-label">TorrentHash</span><div class="movie-film-value-line"><strong class="text-monospace movie-film-value" :title="mediaText('SourceTorrentHash')">{{ mediaText("SourceTorrentHash") || "-" }}</strong><button v-if="mediaText('SourceTorrentHash')" class="btn btn-sm btn-outline-secondary" type="button" @click="copyText(mediaText('SourceTorrentHash'), '来源 Hash')">复制</button></div></div>
                    <div class="movie-film-item"><span class="movie-film-label">BirthTime</span><strong class="movie-film-value"><ScDateTime :value="mediaValue('BirthTime')" unix /></strong></div>
                    <div class="movie-film-item"><span class="movie-film-label">Size</span><strong class="movie-film-value">{{ formatMediaSize(mediaValue("Size")) }}</strong></div>
                    <div class="movie-film-item"><span class="movie-film-label">Resolution</span><strong class="movie-film-value">{{ mediaResolution }}</strong></div>
                    <div class="movie-film-item"><span class="movie-film-label">Duration</span><strong class="movie-film-value">{{ formatMediaDuration(mediaValue("Duration")) }}</strong></div>
                    <div class="movie-film-item"><span class="movie-film-label">BitRate</span><strong class="movie-film-value">{{ formatMediaBitRate(mediaValue("BitRate")) }}</strong></div>
                    <div class="movie-film-item"><span class="movie-film-label">Frame</span><strong class="movie-film-value">{{ formatMediaFrame(mediaValue("FrameAverage")) }}</strong></div>
                    <div class="movie-film-item"><span class="movie-film-label">字幕</span><strong class="movie-film-value">{{ mediaText("HasSub") === "2" ? "有" : "无" }}</strong></div>
                    <div class="movie-film-item"><span class="movie-film-label">自制</span><strong class="movie-film-value">{{ mediaText("SelfMake") === "2" ? "是" : "否" }}</strong></div>
                    <div class="movie-film-item"><span class="movie-film-label">遮罩</span><strong class="movie-film-value">{{ mediaMaskText }}</strong></div>
                    <div class="movie-film-item movie-film-item-wide movie-film-dir"><span class="movie-film-label">Dir</span><div class="movie-film-value-line"><strong class="movie-film-value movie-film-path" :title="mediaPath">{{ mediaPath || "-" }}</strong><button v-if="mediaPath" class="btn btn-sm btn-outline-secondary" type="button" @click="copyText(mediaPath, '媒体路径')">复制</button></div></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="row mt-3 movie-resource-row">
          <div class="col-12 mb-3">
            <div class="card shadow-sm movie-section-card movie-resource-card fc2-resource-panel">
              <div class="card-body">
                <div class="movie-resource-heading-row fc2-resource-heading">
                  <h5 class="card-title mb-0 movie-section-title">帖子资源</h5>
                  <div class="fc2-panel-actions">
                    <span v-if="copyFeedback" class="fc2-copy-feedback" role="status">{{ copyFeedback }}</span>
                    <span class="movie-resource-count">{{ resources.length }}</span>
                    <button class="bucket-list-action" type="button" :disabled="!allHashes.length" @click="copyText(allHashes.join('\n'), '全部 Info Hash')">复制全部 Hash</button>
                    <button class="bucket-list-action" type="button" :disabled="!allMagnets.length" @click="copyText(allMagnets.join('\n'), '全部 Magnet')">复制全部 Magnet</button>
                  </div>
                </div>
                <div class="table-responsive movie-resource-table-wrap fc2-table-wrap">
                  <table class="table table-hover align-middle mb-0 movie-resource-table fc2-table fc2-resource-table">
                    <colgroup><col class="fc2-resource-time"><col class="fc2-resource-size"><col class="fc2-resource-hash"><col class="fc2-resource-magnet"><col class="fc2-resource-attachment"><col class="fc2-resource-action"></colgroup>
                    <thead><tr><th>发布时间</th><th>容量</th><th>Info Hash</th><th>Magnet</th><th>附件数</th><th>操作</th></tr></thead>
                    <tbody>
                      <template v-for="resource in resources" :key="rowText(resource, 'ThreadTID')">
                        <tr>
                          <td><ScDateTime :value="rowValue(resource, 'PostTime')" unix /></td>
                          <td>{{ metaText(resource, "size_text") || "-" }}</td>
                          <td><div class="fc2-copy-cell"><span class="fc2-mono" :title="rowText(resource, 'InfoHashText')">{{ compactLines(rowText(resource, "InfoHashText")) || "-" }}</span><button v-if="rowText(resource, 'InfoHashText')" type="button" class="fc2-copy-button" @click="copyText(rowText(resource, 'InfoHashText'), `Info Hash ${rowText(resource, 'ThreadTID')}`)">复制</button></div></td>
                          <td><div class="fc2-copy-cell"><span class="fc2-mono" :title="rowText(resource, 'MagnetText')">{{ compactLines(rowText(resource, "MagnetText")) || "-" }}</span><button v-if="rowText(resource, 'MagnetText')" type="button" class="fc2-copy-button" @click="copyText(rowText(resource, 'MagnetText'), `Magnet ${rowText(resource, 'ThreadTID')}`)">复制</button></div></td>
                          <td><button v-if="attachments(resource).length" type="button" class="fc2-attachment-toggle" :aria-expanded="expandedThreadTID === rowText(resource, 'ThreadTID')" @click="toggleAttachments(resource)">{{ attachments(resource).length }} {{ expandedThreadTID === rowText(resource, "ThreadTID") ? "收起" : "展开" }}</button><span v-else>0</span></td>
                          <td><div class="fc2-row-actions"><a class="bucket-list-action" :href="`/shthread/${encodeURIComponent(rowText(resource, 'ThreadTID'))}`">帖子详情</a><a class="bucket-list-action" :href="rowText(resource, 'ThreadURL')" target="_blank" rel="noopener noreferrer">原帖</a></div></td>
                        </tr>
                        <tr v-if="expandedThreadTID === rowText(resource, 'ThreadTID')" class="fc2-attachment-row">
                          <td colspan="6"><table class="fc2-attachment-table"><thead><tr><th>文件名</th><th>类型</th><th>容量</th><th>下载次数</th><th>上传时间</th></tr></thead><tbody><tr v-for="(attachment, index) in attachments(resource)" :key="`${rowText(attachment, 'file_name')}-${index}`"><td :title="rowText(attachment, 'file_name')">{{ rowText(attachment, "file_name") }}</td><td>{{ rowText(attachment, "media_type") || "-" }}</td><td>{{ rowText(attachment, "size_text") || "-" }}</td><td>{{ rowText(attachment, "download_count") || "0" }}</td><td>{{ rowText(attachment, "upload_time_text") || "-" }}</td></tr></tbody></table></td>
                        </tr>
                      </template>
                      <tr v-if="!resources.length"><td colspan="6"><div class="sc-empty-state fc2-empty-state">当前影片没有帖子资源。</div></td></tr>
                    </tbody>
                  </table>
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
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/api/request";
import PageShell from "@/components/PageShell.vue";
import ScDateTime from "@/components/ScDateTime.vue";
import type { AnyRecord } from "@/types/gc";
import { asArray, asRecord, pick, text } from "@/utils/format";

const route = useRoute();
const payload = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const activeDetailTab = ref("info");
const expandedThreadTID = ref("");
const copyFeedback = ref("");
const captureFeedback = ref("");
const captureError = ref("");
const coverBusyImageID = ref("");
const deleteConfirmImageID = ref("");
const deleteBusyImageID = ref("");

watch(() => route.fullPath, load, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";
  activeDetailTab.value = "info";
  expandedThreadTID.value = "";
  copyFeedback.value = "";
  captureFeedback.value = "";
  captureError.value = "";
  coverBusyImageID.value = "";
  deleteConfirmImageID.value = "";
  deleteBusyImageID.value = "";
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
const lastScEventName = computed(() => text(payload.value, "LastScEventName"));
const lastScEventHref = computed(() => text(payload.value, "LastScEventHref"));
const hasMedia = computed(() => Boolean(mediaText("ID")));
const mediaImages = computed(() => asArray<AnyRecord>(pick(payload.value, "MediaImages", "mediaImages")));
const primaryCoverImage = computed<AnyRecord>(() => mediaImages.value.find((image) => imageIsCover(image)) || mediaImages.value[0] || {});
const primaryCoverImageID = computed(() => imageID(primaryCoverImage.value));
const resources = computed(() => asArray<AnyRecord>(pick(payload.value, "Resources", "resources")));
const metadataResource = computed(() => resources.value.find((resource) => rowText(resource, "ID") === movieText("MetadataThreadID")));
const allHashes = computed(() => uniqueLines(resources.value.flatMap((resource) => lines(rowText(resource, "InfoHashText")))));
const allMagnets = computed(() => uniqueLines(resources.value.flatMap((resource) => lines(rowText(resource, "MagnetText")))));
const mediaPath = computed(() => [mediaText("FullDir").replace(/\/$/, ""), mediaText("FileName")].filter(Boolean).join("/"));
const mediaResolution = computed(() => {
  const width = Number(mediaValue("Width") || 0);
  const height = Number(mediaValue("Height") || 0);
  return width > 0 && height > 0 ? `${width} × ${height}` : "-";
});
const mediaMaskText = computed(() => mediaText("HasMask") === "2" ? "去码" : mediaText("HasMask") === "3" ? "无码" : "有码");

function rowValue(row: AnyRecord, key: string): unknown { return pick(row, key); }
function rowText(row: AnyRecord, key: string): string { return text(row, key); }
function movieText(key: string): string { return text(movie.value, key); }
function mediaValue(key: string): unknown { return pick(media.value, key); }
function mediaText(key: string): string { return text(media.value, key); }
function scStatValue(key: string): unknown { return pick(scStat.value, key); }
function scStatText(key: string): string { return text(scStat.value, key); }
function resourceMeta(resource: AnyRecord): AnyRecord { return asRecord(pick(resource, "Meta", "meta")); }
function metaText(resource: AnyRecord, key: string): string { return text(resourceMeta(resource), key); }
function attachments(resource: AnyRecord): AnyRecord[] { return asArray<AnyRecord>(pick(resourceMeta(resource), "attachments", "Attachments")); }
function lines(value: string): string[] { return value.split(/\r?\n/).map((item) => item.trim()).filter(Boolean); }
function uniqueLines(values: string[]): string[] { return [...new Set(values)]; }
function compactLines(value: string): string { return lines(value).join(" · "); }
function statusClass(status: string): string { return status === "conflict" ? "is-conflict" : status === "ready" ? "is-ready" : "is-neutral"; }
function mediaStateClass(value: string, activeValue: string): string { return value === activeValue ? "is-ready" : "is-neutral"; }
function imageTypeText(_image: AnyRecord): string { return "联系表封面"; }
function imageFileName(image: AnyRecord): string { return rowText(image, "FilePath").split(/[\\/]/).pop() || "打开原图"; }
function imageID(image: AnyRecord): string { return rowText(image, "ID").trim(); }
function imageCanBeCover(image: AnyRecord): boolean {
  return rowText(image, "ImageType") === "contact_sheet";
}
function imageIsCover(image: AnyRecord): boolean { return pick(image, "IsCover") === true; }
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
function toggleAttachments(resource: AnyRecord) {
  const threadTID = rowText(resource, "ThreadTID");
  expandedThreadTID.value = expandedThreadTID.value === threadTID ? "" : threadTID;
}
async function setCoverImage(image: AnyRecord) {
  const id = imageID(image);
  if (!id || !imageCanBeCover(image) || imageIsCover(image) || coverBusyImageID.value || deleteBusyImageID.value) return;
  coverBusyImageID.value = id;
  deleteConfirmImageID.value = "";
  captureFeedback.value = "";
  captureError.value = "";
  try {
    const fc2Key = movieText("FC2Key");
    const response = asRecord(await sendJSON(`/api/fc2/movies/${encodeURIComponent(fc2Key)}/images/${encodeURIComponent(id)}/cover`, "PUT"));
    if (pick(response, "ok") !== true) throw new Error("接口未确认封面设置结果");
    await refreshPayload();
    captureFeedback.value = "封面设置完成";
  } catch (err) {
    captureError.value = err instanceof Error ? err.message : String(err);
  } finally {
    coverBusyImageID.value = "";
  }
}
function requestDeleteMediaImage(image: AnyRecord) {
  const id = imageID(image);
  if (!id || coverBusyImageID.value || deleteBusyImageID.value) return;
  deleteConfirmImageID.value = id;
  captureFeedback.value = "";
  captureError.value = "";
}
function cancelDeleteMediaImage(image: AnyRecord) {
  const id = imageID(image);
  if (deleteBusyImageID.value === id) return;
  if (deleteConfirmImageID.value === id) deleteConfirmImageID.value = "";
}
async function deleteMediaImage(image: AnyRecord) {
  const id = imageID(image);
  if (!id || deleteConfirmImageID.value !== id || coverBusyImageID.value || deleteBusyImageID.value) return;
  deleteBusyImageID.value = id;
  captureFeedback.value = "";
  captureError.value = "";
  try {
    const fc2Key = movieText("FC2Key");
    const response = asRecord(await sendJSON(`/api/fc2/movies/${encodeURIComponent(fc2Key)}/images/${encodeURIComponent(id)}`, "DELETE"));
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
