<template>
  <div class="style-library-view route-movie-detail">
    <div class="movie-detail-v2 movie-detail-page movie-detail-original-page">
      <div class="movie-detail-top-shell">
        <div class="movie-detail-titlebar">
          <div class="movie-detail-title-main"><h2 class="movie-detail-title">FRAME 01 · 详情示例</h2></div>
          <div class="movie-detail-title-meta">
            <span class="movie-title-meta-chip">DEMO-001</span><span class="movie-title-meta-chip">2026-10-23</span>
          </div>
        </div>
        <div class="row movie-original-top-row">
          <div class="col-lg-7 mb-3 mb-lg-0">
            <div class="card movie-section-card movie-poster-card">
              <button
                type="button"
                class="style-library-poster-button"
                aria-label="放大预览封面"
                @click="preview?.showModal()"
              >
                <div class="thumb-2x2 movie-poster-frame">
                  <img
                    :src="coverArtwork()"
                    alt="粉色几何示例封面"
                    class="card-img-bottom img-fluid movie-poster-image"
                  />
                  <div class="badge-stack"><span class="badge-come">❤ 2</span><span class="badge-times">× 3</span></div>
                  <span class="badge-score">★ 8.1</span>
                </div>
              </button>
            </div>
            <p class="style-library-note">点击封面查看完整图片</p>
          </div>
          <div class="col-lg-5">
            <div class="card movie-section-card movie-info-card">
              <div class="card-body">
                <ul
                  class="nav nav-tabs nav-tabs-custom nav-justified movie-detail-tabs"
                  role="tablist"
                  aria-label="详情内容"
                >
                  <li v-for="tab in ['info', 'rank']" :key="tab" class="nav-item" role="presentation">
                    <button
                      :id="`style-${tab}-tab`"
                      class="nav-link movie-detail-tab"
                      :class="{ active: activeTab === tab }"
                      type="button"
                      role="tab"
                      :aria-selected="activeTab === tab"
                      :aria-controls="`style-${tab}-panel`"
                      :tabindex="activeTab === tab ? 0 : -1"
                      @click="activeTab = tab"
                      @keydown.left.prevent="changeTab"
                      @keydown.right.prevent="changeTab"
                    >
                      {{ tab === "info" ? "Info" : "Rank" }}
                    </button>
                  </li>
                </ul>
                <div
                  v-show="activeTab === 'info'"
                  id="style-info-panel"
                  class="movie-info-pane"
                  role="tabpanel"
                  aria-labelledby="style-info-tab"
                >
                  <table class="table data-table movie-info-table">
                    <tbody>
                      <tr>
                        <th>识别码</th>
                        <td>
                          <div class="movie-info-code-line">
                            <span class="movie-info-code">DEMO-001</span
                            ><button
                              type="button"
                              class="btn btn-success movie-action-button movie-action-button-play"
                              @click="message = '播放按钮示例'"
                            >
                              Play
                            </button>
                          </div>
                        </td>
                      </tr>
                      <tr>
                        <th>发行时间</th>
                        <td><span class="movie-info-value movie-tag-date">上映 2026-10-01</span></td>
                      </tr>
                      <tr>
                        <th>长度</th>
                        <td>
                          <div class="movie-info-action-line">
                            <span class="movie-info-value">120 分钟</span
                            ><button
                              type="button"
                              class="btn btn-outline-info movie-action-button"
                              @click="preview?.showModal()"
                            >
                              查看封面
                            </button>
                          </div>
                        </td>
                      </tr>
                      <tr>
                        <th>制作商</th>
                        <td><span class="tag movie-tag-company">Rudy Studio</span></td>
                      </tr>
                      <tr>
                        <th>发行商</th>
                        <td><span class="tag movie-tag-company">Collection</span></td>
                      </tr>
                      <tr>
                        <th>导演</th>
                        <td><span class="tag movie-tag-person">示例作者</span></td>
                      </tr>
                      <tr>
                        <th>分类</th>
                        <td>
                          <div class="movie-info-tag-group">
                            <span
                              v-for="tag in ['收藏', '示例分类', '完整封面']"
                              :key="tag"
                              class="tag movie-tag-genre"
                              >{{ tag }}</span
                            >
                          </div>
                        </td>
                      </tr>
                      <tr>
                        <th>演员</th>
                        <td><span class="tag movie-tag-person">示例人物</span></td>
                      </tr>
                      <tr>
                        <th>评分</th>
                        <td>
                          <div class="movie-info-score-line">
                            <span class="movie-info-score">8.1</span><span class="tag movie-tag-count">128</span
                            ><button
                              class="btn movie-download-later-button movie-action-button"
                              :class="added ? 'btn-success' : 'btn-outline-warning'"
                              :aria-pressed="added"
                              type="button"
                              @click="added = !added"
                            >
                              {{ added ? "已添加" : "稍后下载" }}
                            </button>
                          </div>
                        </td>
                      </tr>
                      <tr>
                        <th>链接</th>
                        <td><RouterLink to="/styles/gallery" class="tag movie-tag-link">查看图库示例</RouterLink></td>
                      </tr>
                      <tr>
                        <th>RankInfo</th>
                        <td><span class="movie-info-value">2026-10-23 (12)</span></td>
                      </tr>
                      <tr>
                        <th>SC</th>
                        <td><span class="movie-info-empty">—</span></td>
                      </tr>
                      <tr>
                        <th>电影相册</th>
                        <td>
                          <div class="movie-info-album-cell">
                            <button
                              type="button"
                              class="btn movie-album-primary-action"
                              :class="album ? 'btn-success' : 'button-secondary'"
                              :aria-pressed="album"
                              @click="album = !album"
                            >
                              {{ album ? "已加入相册" : "加入电影相册" }}</button
                            ><span class="movie-info-empty">{{ album ? "示例相册" : "—" }}</span>
                          </div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <div
                  v-show="activeTab === 'rank'"
                  id="style-rank-panel"
                  class="movie-rank-panel"
                  role="tabpanel"
                  aria-labelledby="style-rank-tab"
                >
                  <table class="data-table table">
                    <thead>
                      <tr>
                        <th>日期</th>
                        <th>排名</th>
                        <th>变化</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="row in ranks" :key="row.date">
                        <td>{{ row.date }}</td>
                        <td>#{{ row.rank }}</td>
                        <td><StatusBadge :label="row.change" tone="good" /></td>
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
    <p class="style-library-note" role="status">
      {{ message || "详情信息区域沿用业务页尺寸；Info / Rank 可切换，下载与相册操作为本地演示。" }}
    </p>
    <dialog
      ref="preview"
      class="modal-panel-v2 style-library-dialog style-library-preview"
      aria-labelledby="style-preview-title"
      @click.self="preview?.close()"
    >
      <div class="modal-header-v2">
        <h2 id="style-preview-title">完整封面</h2>
        <button type="button" class="modal-close-v2" aria-label="关闭预览" @click="preview?.close()">×</button>
      </div>
      <div class="modal-body-v2"><img :src="coverArtwork()" alt="粉色几何示例封面完整预览" /></div>
      <div class="modal-footer-v2">
        <button type="button" class="button-secondary" @click="preview?.close()">关闭</button>
      </div>
    </dialog>
  </div>
</template>
<script setup lang="ts">
import { nextTick, ref } from "vue";
import StatusBadge from "@/features/remote-dailybest/components/StatusBadge.vue";
import { coverArtwork } from "../fixtures";
const activeTab = ref("info");
const added = ref(false);
const album = ref(false);
const message = ref("");
const preview = ref<HTMLDialogElement>();
const ranks = [
  { date: "2026-10-23", rank: 12, change: "↑ 4" },
  { date: "2026-10-22", rank: 16, change: "↑ 2" },
  { date: "2026-10-21", rank: 18, change: "首次入榜" }
];
async function changeTab() {
  activeTab.value = activeTab.value === "info" ? "rank" : "info";
  await nextTick();
  document.getElementById(`style-${activeTab.value}-tab`)?.focus();
}
</script>
