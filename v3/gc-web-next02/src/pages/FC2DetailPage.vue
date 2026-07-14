<template>
  <PageShell class="sc-events-page-shell fc2-page-shell fc2-detail-page-shell" :title="pageTitle" :note="pageNote" eyebrow="">
    <div id="page-root" class="container-fluid sc-events-page-root next02-sc-pink-theme next02-sc-events-rewrite bucket-list-standard sh-bucket-list-page fc2-page fc2-detail-page">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <section class="next02-sc-hero next02-sc-events-hero next02-sc-movie-top-shell sh-bucket-hero fc2-hero fc2-detail-hero">
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">FC2 / MOVIE DETAIL</span>
            <h4>{{ movieText("FC2Key") || pageTitle }}</h4>
            <p>{{ movieText("Title") || pageNote }}</p>
            <div class="fc2-hero-facts">
              <span :title="movieText('PerformerText')">出演：{{ movieText("PerformerText") || "-" }}</span>
              <span class="fc2-state-badge" :class="statusClass(movieText('DataStatus'))">{{ movieText("DataStatus") || "-" }}</span>
              <span>资源 {{ resources.length }}</span>
            </div>
          </div>
          <div class="next02-sc-hero-actions">
            <a class="btn btn-outline-secondary" href="/fclist">返回 FC2 列表</a>
            <a class="btn btn-outline-secondary" href="/shpool" target="_blank" rel="noopener noreferrer">SH 帖子池</a>
          </div>
        </section>

        <section id="fc2-metadata" class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card fc2-detail-panel fc2-metadata-panel">
          <div class="card-body">
            <div class="fc2-panel-heading">
              <div><span class="next02-sc-eyebrow">MOVIE PROFILE</span><h5>影片资料</h5></div>
              <div class="sh-bucket-toolbar-metrics" aria-label="影片资料摘要">
                <span :title="`资源数 ${resources.length}`"><strong>{{ resources.length }}</strong></span>
              </div>
            </div>
            <div class="fc2-sc-summary" aria-label="SC 统计">
              <span><small>SC 次数</small><strong>{{ scStatText("ScTimes") || "0" }}</strong></span>
              <span><small>Come 次数</small><strong>{{ scStatText("ComeTimes") || "0" }}</strong></span>
              <span><small>最近 SC</small><strong><ScDateTime v-if="scStatValue('LastScTime')" :value="scStatValue('LastScTime')" unix /><template v-else>-</template></strong></span>
              <span><small>最近事件</small><strong><a v-if="lastScEventHref" :href="lastScEventHref">{{ lastScEventName }}</a><template v-else>-</template></strong></span>
            </div>
            <div class="table-responsive fc2-table-wrap">
              <table class="table align-middle mb-0 sc-events-table sh-bucket-table fc2-table fc2-profile-table">
                <thead><tr><th>影片标识</th><th>影片名称</th><th>出演信息</th><th>有码状态</th><th>数据状态</th><th>资料来源帖子</th></tr></thead>
                <tbody><tr>
                  <td><strong>{{ movieText("FC2Key") || "-" }}</strong></td>
                  <td class="fc2-title-cell" :title="movieText('Title')">{{ movieText("Title") || "-" }}</td>
                  <td class="fc2-performer-cell" :title="movieText('PerformerText')">{{ movieText("PerformerText") || "-" }}</td>
                  <td><span class="fc2-state-badge is-mask">{{ movieText("MaskType") || "未知" }}</span></td>
                  <td><span class="fc2-state-badge" :class="statusClass(movieText('DataStatus'))">{{ movieText("DataStatus") || "-" }}</span></td>
                  <td><a v-if="metadataResource" class="fc2-thread-link" :href="`/shthread/${encodeURIComponent(rowText(metadataResource, 'ThreadTID'))}`">{{ rowText(metadataResource, "ThreadTID") }}</a><span v-else>-</span></td>
                </tr></tbody>
              </table>
            </div>
          </div>
        </section>

        <section v-if="hasMedia" id="fc2-media" class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card fc2-detail-panel fc2-media-panel">
          <div class="card-body">
            <div class="fc2-panel-heading">
              <div><span class="next02-sc-eyebrow">MEDIA PROFILE</span><h5>媒体资料</h5></div>
              <div class="fc2-panel-actions">
                <button v-if="mediaText('SourceTorrentHash')" class="bucket-list-action" type="button" @click="copyText(mediaText('SourceTorrentHash'), '来源 Hash')">复制 Hash</button>
                <button v-if="mediaPath" class="bucket-list-action" type="button" @click="copyText(mediaPath, '媒体路径')">复制路径</button>
              </div>
            </div>
            <div class="table-responsive fc2-table-wrap">
              <table class="table align-middle mb-0 sc-events-table sh-bucket-table fc2-table fc2-media-table">
                <colgroup><col class="fc2-media-play"><col class="fc2-media-file"><col class="fc2-media-time"><col class="fc2-media-size"><col class="fc2-media-resolution"><col class="fc2-media-duration"><col class="fc2-media-bitrate"><col class="fc2-media-frame"><col class="fc2-media-state"><col class="fc2-media-state"><col class="fc2-media-state"></colgroup>
                <thead><tr><th>播放</th><th>文件名</th><th>下载时间</th><th>大小</th><th>分辨率</th><th>时长</th><th>码率</th><th>平均帧率</th><th>字幕</th><th>自制</th><th>遮罩</th></tr></thead>
                <tbody><tr>
                  <td><a v-if="mediaValue('ShowPlayButton')" class="btn btn-sm fc2-play-button" :class="mediaText('PlayButtonClass')" :href="mediaText('VideoURL')" target="_blank" rel="noopener noreferrer">{{ mediaText("PlayButtonText") || "Play" }}</a><span v-else>-</span></td>
                  <td class="fc2-media-file-cell" :title="mediaText('FileName')">{{ mediaText("FileName") || "-" }}</td>
                  <td><ScDateTime :value="mediaValue('BirthTime')" unix /></td>
                  <td>{{ formatMediaSize(mediaValue("Size")) }}</td>
                  <td>{{ mediaResolution }}</td>
                  <td>{{ formatMediaDuration(mediaValue("Duration")) }}</td>
                  <td>{{ formatMediaBitRate(mediaValue("BitRate")) }}</td>
                  <td>{{ formatMediaFrame(mediaValue("FrameAverage")) }}</td>
                  <td><span class="fc2-state-badge" :class="mediaStateClass(mediaText('HasSub'), '2')">{{ mediaText("HasSub") === "2" ? "有" : "无" }}</span></td>
                  <td><span class="fc2-state-badge" :class="mediaStateClass(mediaText('SelfMake'), '2')">{{ mediaText("SelfMake") === "2" ? "是" : "否" }}</span></td>
                  <td><span class="fc2-state-badge is-mask">{{ mediaMaskText }}</span></td>
                </tr></tbody>
              </table>
            </div>
            <div class="fc2-media-source-row">
              <span><strong>来源 Hash</strong><code :title="mediaText('SourceTorrentHash')">{{ mediaText("SourceTorrentHash") || "-" }}</code></span>
              <span><strong>媒体路径</strong><code :title="mediaPath">{{ mediaPath || "-" }}</code></span>
            </div>
          </div>
        </section>

        <section v-if="hasMedia" id="fc2-media-images" class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card fc2-detail-panel fc2-capture-panel">
          <div class="card-body">
            <div class="fc2-panel-heading">
              <div><span class="next02-sc-eyebrow">VIDEO CAPTURE</span><h5>视频图片</h5></div>
              <div class="fc2-panel-actions">
                <span v-if="captureFeedback" class="fc2-copy-feedback" role="status">{{ captureFeedback }}</span>
              </div>
            </div>

            <div class="fc2-capture-forms">
              <form class="fc2-capture-form" @submit.prevent="generateScreenshot">
                <div class="fc2-capture-form-copy">
                  <strong>生成截图</strong>
                  <span>输入视频时间点，输出一张 PNG 图片。</span>
                </div>
                <label class="fc2-capture-field">
                  <span>截图时间（秒）</span>
                  <input v-model="screenshotTime" class="form-control" type="number" min="0" step="any" required :disabled="captureBusy !== ''">
                </label>
                <button class="btn bucket-list-apply" type="submit" :disabled="captureBusy !== '' || !screenshotFormReady">
                  {{ captureBusy === "screenshot" ? "正在生成..." : "生成截图" }}
                </button>
              </form>

              <form class="fc2-capture-form" @submit.prevent="generateContactSheet">
                <div class="fc2-capture-form-copy">
                  <strong>生成联系表</strong>
                  <span>按完整视频时长均匀抽帧并组成一张 PNG 图片。</span>
                </div>
                <div class="fc2-contact-sheet-fields">
                  <label class="fc2-capture-field"><span>列数</span><input v-model="contactColumns" class="form-control" type="number" min="1" step="1" required :disabled="captureBusy !== ''"></label>
                  <label class="fc2-capture-field"><span>行数</span><input v-model="contactRows" class="form-control" type="number" min="1" step="1" required :disabled="captureBusy !== ''"></label>
                  <label class="fc2-capture-field"><span>单格宽度</span><input v-model="contactCellWidth" class="form-control" type="number" min="1" step="1" required :disabled="captureBusy !== ''"></label>
                </div>
                <button class="btn bucket-list-apply" type="submit" :disabled="captureBusy !== '' || !contactFormReady">
                  {{ captureBusy === "contact_sheet" ? "正在生成..." : "生成联系表" }}
                </button>
              </form>
            </div>

            <div v-if="captureError" class="alert alert-danger fc2-capture-alert" role="alert">{{ captureError }}</div>

            <div v-if="mediaImages.length" class="fc2-capture-gallery">
              <article v-for="image in mediaImages" :key="rowText(image, 'ID')" class="fc2-capture-card" :class="{ 'is-cover': imageIsCover(image) }">
                <a class="fc2-capture-preview" :href="rowText(image, 'ImageURL')" target="_blank" rel="noopener noreferrer">
                  <img :src="rowText(image, 'ImageURL')" :alt="imageTypeText(image)" loading="lazy">
                </a>
                <div class="fc2-capture-card-meta">
                  <span class="fc2-state-badge" :class="rowText(image, 'ImageType') === 'contact_sheet' ? 'is-mask' : 'is-ready'">{{ imageTypeText(image) }}</span>
                  <span v-if="imageIsCover(image)" class="fc2-cover-badge">当前封面</span>
                  <ScDateTime :value="rowValue(image, 'CreatedOn')" unix />
                  <a :href="rowText(image, 'ImageURL')" target="_blank" rel="noopener noreferrer" :title="rowText(image, 'FilePath')">{{ imageFileName(image) }}</a>
                  <div class="fc2-capture-card-actions">
                    <button
                      v-if="imageIsScreenshot(image) && !imageIsCover(image)"
                      type="button"
                      class="btn btn-sm btn-outline-primary"
                      :disabled="Boolean(coverBusyImageID) || Boolean(deleteBusyImageID) || captureBusy !== ''"
                      @click="setCoverImage(image)"
                    >
                      {{ coverBusyImageID === imageID(image) ? "设置中" : "设为封面" }}
                    </button>
                    <template v-if="deleteConfirmImageID === imageID(image)">
                      <button type="button" class="btn btn-sm btn-danger" :disabled="Boolean(coverBusyImageID) || deleteBusyImageID === imageID(image)" @click="deleteMediaImage(image)">
                        {{ deleteBusyImageID === imageID(image) ? "正在删除" : "确认删除" }}
                      </button>
                      <button type="button" class="btn btn-sm btn-outline-secondary" :disabled="Boolean(coverBusyImageID) || deleteBusyImageID === imageID(image)" @click="cancelDeleteMediaImage(image)">取消</button>
                    </template>
                    <button v-else type="button" class="btn btn-sm btn-outline-danger" :disabled="Boolean(coverBusyImageID) || Boolean(deleteBusyImageID) || captureBusy !== ''" @click="requestDeleteMediaImage(image)">删除</button>
                  </div>
                </div>
              </article>
            </div>
            <div v-else class="sc-empty-state fc2-empty-state">当前媒体还没有生成图片。</div>
          </div>
        </section>

        <section class="card next02-sc-panel next02-sc-events-panel next02-sc-movie-table-card fc2-detail-panel fc2-resource-panel">
          <div class="card-body">
            <div class="fc2-panel-heading">
              <div><span class="next02-sc-eyebrow">SH THREAD RESOURCES</span><h5>帖子资源</h5></div>
              <div class="fc2-panel-actions">
                <span v-if="copyFeedback" class="fc2-copy-feedback" role="status">{{ copyFeedback }}</span>
                <button class="bucket-list-action" type="button" :disabled="!allHashes.length" @click="copyText(allHashes.join('\n'), '全部 Info Hash')">复制全部 Hash</button>
                <button class="bucket-list-action" type="button" :disabled="!allMagnets.length" @click="copyText(allMagnets.join('\n'), '全部 Magnet')">复制全部 Magnet</button>
              </div>
            </div>
            <div class="table-responsive fc2-table-wrap">
              <table class="table table-hover align-middle mb-0 sc-events-table sh-bucket-table fc2-table fc2-resource-table">
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
                      <td colspan="6">
                        <table class="fc2-attachment-table">
                          <thead><tr><th>文件名</th><th>类型</th><th>容量</th><th>下载次数</th><th>上传时间</th></tr></thead>
                          <tbody><tr v-for="(attachment, index) in attachments(resource)" :key="`${rowText(attachment, 'file_name')}-${index}`"><td :title="rowText(attachment, 'file_name')">{{ rowText(attachment, "file_name") }}</td><td>{{ rowText(attachment, "media_type") || "-" }}</td><td>{{ rowText(attachment, "size_text") || "-" }}</td><td>{{ rowText(attachment, "download_count") || "0" }}</td><td>{{ rowText(attachment, "upload_time_text") || "-" }}</td></tr></tbody>
                        </table>
                      </td>
                    </tr>
                  </template>
                  <tr v-if="!resources.length"><td colspan="6"><div class="sc-empty-state fc2-empty-state">当前影片没有帖子资源。</div></td></tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>
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
const expandedThreadTID = ref("");
const copyFeedback = ref("");
type NumericInput = string | number;

const screenshotTime = ref<NumericInput>("");
const contactColumns = ref<NumericInput>("");
const contactRows = ref<NumericInput>("");
const contactCellWidth = ref<NumericInput>("");
const captureBusy = ref("");
const captureFeedback = ref("");
const captureError = ref("");
const coverBusyImageID = ref("");
const deleteConfirmImageID = ref("");
const deleteBusyImageID = ref("");

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
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
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
const screenshotFormReady = computed(() => {
  const timeSeconds = Number(screenshotTime.value);
  return hasInputValue(screenshotTime.value) && Number.isFinite(timeSeconds) && timeSeconds >= 0;
});
const contactFormReady = computed(() => [contactColumns.value, contactRows.value, contactCellWidth.value].every((value) => {
  const numberValue = Number(value);
  return hasInputValue(value) && Number.isInteger(numberValue) && numberValue > 0;
}));

function rowValue(row: AnyRecord, key: string): unknown { return pick(row, key); }
function hasInputValue(value: NumericInput): boolean { return String(value).trim() !== ""; }
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
function imageTypeText(image: AnyRecord): string { return rowText(image, "ImageType") === "contact_sheet" ? "联系表" : "截图"; }
function imageFileName(image: AnyRecord): string { return rowText(image, "FilePath").split(/[\\/]/).pop() || "打开原图"; }
function imageID(image: AnyRecord): string { return rowText(image, "ID").trim(); }
function imageIsScreenshot(image: AnyRecord): boolean { return rowText(image, "ImageType") === "screenshot"; }
function imageIsCover(image: AnyRecord): boolean {
  const id = imageID(image);
  return id !== "" && (pick(image, "IsCover") === true || id === movieText("CoverImageID"));
}
function applyCoverState(coverImageID: string, images: AnyRecord[] = mediaImages.value): void {
  const normalizedCoverImageID = coverImageID.trim();
  const numericCoverImageID = Number(normalizedCoverImageID);
  payload.value = {
    ...payload.value,
    Movie: {
      ...movie.value,
      CoverImageID: Number.isSafeInteger(numericCoverImageID) && numericCoverImageID > 0 ? numericCoverImageID : 0,
    },
    MediaImages: images.map((item) => ({
      ...item,
      IsCover: normalizedCoverImageID !== "" && normalizedCoverImageID !== "0" && imageID(item) === normalizedCoverImageID,
    })),
  };
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
function toggleAttachments(resource: AnyRecord) {
  const threadTID = rowText(resource, "ThreadTID");
  expandedThreadTID.value = expandedThreadTID.value === threadTID ? "" : threadTID;
}
async function generateScreenshot() {
  const timeSeconds = Number(screenshotTime.value);
  if (!hasInputValue(screenshotTime.value) || !Number.isFinite(timeSeconds) || timeSeconds < 0) {
    captureError.value = "请输入有效的截图时间。";
    return;
  }
  await generateMediaImage("screenshot", { time_seconds: timeSeconds });
}
async function generateContactSheet() {
  const columns = Number(contactColumns.value);
  const rows = Number(contactRows.value);
  const cellWidth = Number(contactCellWidth.value);
  if (![columns, rows, cellWidth].every((value) => Number.isInteger(value) && value > 0)) {
    captureError.value = "列数、行数和单格宽度必须是正整数。";
    return;
  }
  await generateMediaImage("contact_sheet", { columns, rows, cell_width: cellWidth });
}
async function generateMediaImage(type: "screenshot" | "contact_sheet", request: AnyRecord) {
  captureBusy.value = type;
  captureFeedback.value = "";
  captureError.value = "";
  const fc2Key = movieText("FC2Key");
  const endpoint = type === "screenshot" ? "screenshot" : "contact-sheet";
  try {
    const response = asRecord(await sendJSON(`/api/fc2/movies/${encodeURIComponent(fc2Key)}/${endpoint}`, "POST", request));
    const image = asRecord(pick(response, "image", "Image"));
    if (!rowText(image, "ID")) throw new Error("接口未返回图片记录");
    const nextImages = [image, ...mediaImages.value];
    if (type === "screenshot") {
      applyCoverState(text(response, "cover_image_id") || imageID(image), nextImages);
      captureFeedback.value = "截图生成完成并已设为封面";
    } else {
      payload.value = { ...payload.value, MediaImages: nextImages };
      captureFeedback.value = "联系表生成完成";
    }
  } catch (err) {
    captureError.value = err instanceof Error ? err.message : String(err);
  } finally {
    captureBusy.value = "";
  }
}
async function setCoverImage(image: AnyRecord) {
  const id = imageID(image);
  if (!id || !imageIsScreenshot(image) || imageIsCover(image) || coverBusyImageID.value || deleteBusyImageID.value || captureBusy.value) return;
  coverBusyImageID.value = id;
  deleteConfirmImageID.value = "";
  captureFeedback.value = "";
  captureError.value = "";
  try {
    const fc2Key = movieText("FC2Key");
    const response = asRecord(await sendJSON(`/api/fc2/movies/${encodeURIComponent(fc2Key)}/images/${encodeURIComponent(id)}/cover`, "PUT"));
    if (pick(response, "ok") !== true) throw new Error("接口未确认封面设置结果");
    applyCoverState(text(response, "cover_image_id") || id);
    captureFeedback.value = "封面设置完成";
  } catch (err) {
    captureError.value = err instanceof Error ? err.message : String(err);
  } finally {
    coverBusyImageID.value = "";
  }
}
function requestDeleteMediaImage(image: AnyRecord) {
  const id = imageID(image);
  if (!id || coverBusyImageID.value || deleteBusyImageID.value || captureBusy.value) return;
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
    applyCoverState(text(response, "cover_image_id"), mediaImages.value.filter((item) => imageID(item) !== id));
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
