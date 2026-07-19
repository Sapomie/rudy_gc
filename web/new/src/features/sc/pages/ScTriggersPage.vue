<template>
  <PageShell class="sc-triggers-page-shell" :title="title" eyebrow="">
    <div id="page-root" class="container-fluid sc-triggers-page-root next02-sc-pink-theme next02-sc-trigger-rewrite">
      <div v-if="loading" class="alert alert-info py-2">加载中...</div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <template v-else>
        <section
          class="next02-sc-hero next02-sc-trigger-hero next02-sc-movie-top-shell"
          aria-labelledby="scTriggerTitle"
        >
          <div class="next02-sc-hero-copy">
            <span class="next02-sc-eyebrow">SC / IMPORT CONSOLE</span>
            <h4 id="scTriggerTitle">SC Preview &amp; Submit</h4>
            <p>扫描 SC 根目录，核对事件和影片状态，再提交后台任务。</p>
          </div>
          <div class="next02-sc-hero-actions">
            <a href="/databackfill" class="btn btn-outline-secondary">回填任务台</a>
            <a href="/crawltasklist" class="btn btn-outline-primary">查看任务列表</a>
          </div>
        </section>

        <div
          id="msgArea"
          class="alert fade show small"
          :class="messageOk ? 'alert-success' : 'alert-danger'"
          :style="{ display: message ? 'block' : 'none' }"
        >
          {{ message }}
        </div>

        <div class="card next02-sc-panel next02-sc-scan-panel">
          <div class="card-body">
            <div class="next02-sc-panel-head">
              <div>
                <span class="next02-sc-panel-kicker">STEP 01 / DISCOVER</span>
                <h4 class="card-title">扫描 SC 事件</h4>
              </div>
              <span class="next02-sc-status-dot" :class="{ 'is-busy': previewLoading }">{{
                previewLoading ? "扫描中" : "等待扫描"
              }}</span>
            </div>
            <p class="card-title-desc">
              扫描配置中的 SC 根目录，并要求其中仅有一个合法事件目录（例如：2026-03-17-20-00）。
            </p>
            <p class="text-muted small mb-3">
              SC Root: <code>{{ scRootDir }}</code>
            </p>
            <button
              id="btnScAddPreview"
              type="button"
              class="btn btn-outline-success"
              :disabled="previewLoading"
              @click="loadPreview"
            >
              {{ previewLoading ? "正在扫描..." : "Add Preview" }}
            </button>
          </div>
        </div>

        <div
          class="card mt-3 next02-sc-panel next02-sc-preview-panel"
          id="scPreviewPanel"
          :class="{ 'd-none': !previewVisible }"
        >
          <div class="card-body">
            <div class="next02-sc-panel-head">
              <div>
                <span class="next02-sc-panel-kicker">STEP 02 / REVIEW</span>
                <h4 class="card-title mb-0">核对并提交 SC 事件</h4>
              </div>
              <span class="next02-sc-status-dot is-ready">{{ previewName || "Preview" }}</span>
            </div>
            <form id="formScAddConfirm" @submit.prevent="submitAdd">
              <div
                id="scAddPreviewMsg"
                class="small mb-3"
                :class="
                  previewMsg ? (previewOk ? 'alert alert-success py-2 px-3' : 'alert alert-danger py-2 px-3') : ''
                "
              >
                {{ previewMsg }}
              </div>

              <div class="next02-sc-review-grid mb-3">
                <div class="next02-sc-review-field is-name">
                  <label class="form-label">事件名</label>
                  <input type="text" class="form-control" id="scPreviewName" :value="previewName" readonly />
                </div>
                <div class="next02-sc-review-field is-image">
                  <label class="form-label">图片</label>
                  <input type="text" class="form-control" id="scPreviewImage" :value="previewImageText" readonly />
                </div>
                <div class="next02-sc-review-field is-count">
                  <label class="form-label">影片数</label>
                  <input
                    type="text"
                    class="form-control"
                    id="scPreviewMovieCount"
                    :value="String(previewMovieCount)"
                    readonly
                  />
                </div>
                <div class="next02-sc-review-field is-selected-count">
                  <label class="form-label">SC 影片数</label>
                  <input
                    type="text"
                    class="form-control"
                    id="scPreviewSelectedMovieCount"
                    :value="String(selectedScMovies.length)"
                    readonly
                  />
                </div>
                <div class="next02-sc-review-field is-come-movie">
                  <label class="form-label">Come Movie</label>
                  <select class="form-select" id="scPreviewComeMovie" v-model="form.comeMovieRef">
                    <option value="">可为空</option>
                    <option v-for="movie in selectedScMovies" :key="movieRef(movie)" :value="movieRef(movie)">
                      {{ movie.movieName }}{{ movie.title ? ` · ${movie.title}` : "" }}
                    </option>
                  </select>
                </div>
                <div class="next02-sc-review-field is-movie-cast">
                  <label class="form-label">Movie Cast</label>
                  <select
                    class="form-select"
                    id="scPreviewMovieCast"
                    v-model="form.movieCast"
                    :disabled="comeMovieIsFC2"
                  >
                    <option value="">可为空</option>
                    <option v-for="cast in comeMovieCasts" :key="cast" :value="cast">{{ cast }}</option>
                  </select>
                </div>
                <div class="next02-sc-review-field is-kind">
                  <label class="form-label">种类</label>
                  <input type="text" class="form-control" id="scPreviewKind" v-model="form.kind" placeholder="可为空" />
                </div>
                <div class="next02-sc-review-field is-duration">
                  <label class="form-label">时长(分钟)</label>
                  <input
                    type="number"
                    class="form-control"
                    id="scPreviewDuration"
                    v-model.number="form.duration"
                    placeholder="分钟"
                  />
                </div>
                <div class="next02-sc-review-field is-fg">
                  <label class="form-label">FG</label>
                  <input type="text" class="form-control" id="scPreviewFg" v-model="form.fg" placeholder="可为空" />
                </div>
                <div class="next02-sc-review-field is-vessel">
                  <label class="form-label">Vessel</label>
                  <input
                    type="text"
                    class="form-control"
                    id="scPreviewVessel"
                    v-model="form.vessel"
                    placeholder="可为空"
                  />
                </div>
                <div class="next02-sc-review-field is-remarks">
                  <label class="form-label">Remarks</label>
                  <textarea
                    class="form-control"
                    id="scPreviewRemarks"
                    v-model="form.remarks"
                    rows="1"
                    placeholder="可为空"
                  ></textarea>
                </div>
                <div class="next02-sc-review-field is-movie-details">
                  <label class="form-label">影片明细</label>
                  <div class="sc-preview-card-surface cards-page-shell next02-pink-cards-page">
                    <div
                      class="sc-preview-movie-grid movie-grid-v2 next02-pink-movie-grid"
                      id="scPreviewMovieTableBody"
                    >
                      <div v-if="!previewMovies.length" class="movie-empty">暂无预览数据</div>
                      <div
                        v-for="(movie, index) in previewMovies"
                        :key="movieRef(movie) || index"
                        class="col movie-grid-col-v2"
                      >
                        <component :is="previewCardComponent(movie)" :movie="movie.card">
                          <template #actions>
                            <li class="list-group-item sc-preview-is-sc-buttons">
                              <div>
                                <input
                                  :id="`sc-preview-is-sc-${index}-1`"
                                  v-model.number="movie.isSc"
                                  type="radio"
                                  class="btn-check sc-preview-is-sc-input-1"
                                  :name="`sc-preview-is-sc-${index}`"
                                  :value="GLIST_IS_NOT_SC"
                                  autocomplete="off"
                                  :data-preview-movie-idx="index"
                                  :data-preview-is-sc="GLIST_IS_NOT_SC"
                                />
                                <label
                                  class="btn btn-outline-secondary sc-preview-is-sc-btn sc-preview-is-sc-btn-1"
                                  :for="`sc-preview-is-sc-${index}-1`"
                                  >不计入</label
                                >
                              </div>
                              <div>
                                <input
                                  :id="`sc-preview-is-sc-${index}-2`"
                                  v-model.number="movie.isSc"
                                  type="radio"
                                  class="btn-check sc-preview-is-sc-input-2"
                                  :name="`sc-preview-is-sc-${index}`"
                                  :value="GLIST_IS_SC"
                                  autocomplete="off"
                                  :data-preview-movie-idx="index"
                                  :data-preview-is-sc="GLIST_IS_SC"
                                />
                                <label
                                  class="btn btn-outline-secondary sc-preview-is-sc-btn sc-preview-is-sc-btn-2"
                                  :for="`sc-preview-is-sc-${index}-2`"
                                  >计入SC</label
                                >
                              </div>
                            </li>
                          </template>
                        </component>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="d-flex justify-content-end next02-sc-submit-row">
                <button type="submit" class="btn btn-success">确认提交</button>
              </div>
            </form>
          </div>
        </div>
      </template>
    </div>
  </PageShell>
</template>

<script setup lang="ts">
import "@/features/sc/styles/index.css";
import { computed, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPathForRoute, getJSON, sendJSON } from "@/features/sc/api";
import FC2MovieCard from "@/features/fc2/components/FC2MovieCard.vue";
import MovieCard from "@/features/movies/components/MovieCard.vue";
import PageShell from "@/shared/components/PageShell.vue";
import type { ScRecord } from "@/features/sc/types";
import { asArray, asRecord, pick, text } from "@/shared/utils/format";
import { normalizeScDuration, type ScDurationInput } from "@/shared/utils/scDuration";

type PreviewMovie = {
  movieType: string;
  movieName: string;
  movieJavId: string;
  fc2MovieId: number;
  movieKey: string;
  title: string;
  movieHref: string;
  jacketImg: string;
  casts: string[];
  card: ScRecord;
  isSc: number;
};

const GLIST_IS_NOT_SC = 1;
const GLIST_IS_SC = 2;
const MOVIE_TYPE_JAV = "jav";
const MOVIE_TYPE_FC2 = "fc2ppv";

const route = useRoute();
const payload = ref<ScRecord>({});
const loading = ref(false);
const error = ref("");
const message = ref("");
const messageOk = ref(true);
const previewMsg = ref("");
const previewOk = ref(true);
const previewLoading = ref(false);
const previewVisible = ref(false);
const previewData = ref<ScRecord>({});
const previewMovies = ref<PreviewMovie[]>([]);
const form = reactive<{
  comeMovieRef: string;
  movieCast: string;
  kind: string;
  duration: ScDurationInput;
  fg: string;
  vessel: string;
  remarks: string;
}>({
  comeMovieRef: "",
  movieCast: "",
  kind: "",
  duration: "",
  fg: "",
  vessel: "",
  remarks: ""
});

watch(() => route.fullPath, load, { immediate: true });
watch(
  () => form.comeMovieRef,
  () => {
    form.movieCast = "";
  }
);

async function load() {
  loading.value = true;
  error.value = "";
  message.value = "";
  try {
    payload.value = asRecord(await getJSON(apiPathForRoute(route)));
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

const title = computed(() => text(payload.value, "Title") || "SC Triggers");
const scRootDir = computed(() => text(payload.value, "ScRootDir"));
const previewName = computed(() => text(previewData.value, "scName"));
const previewMovieCount = computed(() => Number(pick(previewData.value, "movieCount") || 0));
const previewImageText = computed(() => {
  if (!Object.keys(previewData.value).length) return "";
  return pick(previewData.value, "imageFound") ? text(previewData.value, "imageName") || "已发现图片" : "无图片";
});
const selectedScMovies = computed(() => previewMovies.value.filter((movie) => movie.isSc === GLIST_IS_SC));
const selectedComeMovie = computed(() => previewMovies.value.find((item) => movieRef(item) === form.comeMovieRef));
const comeMovieIsFC2 = computed(() => selectedComeMovie.value?.movieType === MOVIE_TYPE_FC2);
const comeMovieCasts = computed(() => {
  return selectedComeMovie.value?.casts || [];
});
watch(selectedScMovies, (movies) => {
  if (form.comeMovieRef && !movies.some((movie) => movieRef(movie) === form.comeMovieRef)) {
    form.comeMovieRef = "";
  }
});

function resetForm() {
  form.comeMovieRef = "";
  form.movieCast = "";
  form.kind = "";
  form.duration = "";
  form.fg = "";
  form.vessel = "";
  form.remarks = "";
}

function normalizeMovies(raw: unknown): PreviewMovie[] {
  return asArray<ScRecord>(raw).map((movie) => {
    const movieType = text(movie, "movieType") || (text(movie, "movieJavId") ? MOVIE_TYPE_JAV : MOVIE_TYPE_FC2);
    const movieName = text(movie, "movieName");
    const movieKey = text(movie, "movieKey");
    const title = text(movie, "title");
    const jacketImg = text(movie, "jacketImg");
    const card = asRecord(pick(movie, "card"));
    return {
      movieType,
      movieName,
      movieJavId: text(movie, "movieJavId"),
      fc2MovieId: Number(pick(movie, "fc2MovieId") || 0),
      movieKey,
      title,
      movieHref: text(movie, "movieHref"),
      jacketImg,
      casts: asArray<string>(pick(movie, "casts")),
      card: Object.keys(card).length
        ? card
        : movieType === MOVIE_TYPE_FC2
          ? { FC2Key: movieKey || movieName, Title: title, CoverImageURL: jacketImg }
          : { name: movieName, title, jacket_img: jacketImg },
      isSc: GLIST_IS_SC
    };
  });
}

function previewCardComponent(movie: PreviewMovie) {
  return movie.movieType === MOVIE_TYPE_FC2 ? FC2MovieCard : MovieCard;
}

function movieRef(movie: PreviewMovie): string {
  if (movie.movieType === MOVIE_TYPE_FC2 && movie.fc2MovieId > 0) return `${MOVIE_TYPE_FC2}:${movie.fc2MovieId}`;
  if (movie.movieType === MOVIE_TYPE_JAV && movie.movieJavId) return `${MOVIE_TYPE_JAV}:${movie.movieJavId}`;
  return "";
}

async function loadPreview() {
  previewLoading.value = true;
  message.value = "";
  try {
    const data = asRecord(await sendJSON("/api/triggers/sc/add-preview", "POST", {}));
    previewData.value = data;
    previewMovies.value = normalizeMovies(pick(data, "movies"));
    previewVisible.value = true;
    previewMsg.value = "";
    resetForm();
  } catch (err) {
    message.value = `异常：${err instanceof Error ? err.message : String(err)}`;
    messageOk.value = false;
  } finally {
    previewLoading.value = false;
  }
}

function buildPayload() {
  const comeMovie = previewMovies.value.find((movie) => movieRef(movie) === form.comeMovieRef);
  return {
    comeMovieType: comeMovie?.movieType || "",
    comeMovieJavId: comeMovie?.movieJavId || "",
    comeFC2MovieId: comeMovie?.fc2MovieId || 0,
    movieCast: form.movieCast.trim(),
    kind: form.kind.trim(),
    duration: normalizeScDuration(form.duration),
    fg: form.fg.trim(),
    vessel: form.vessel.trim(),
    remarks: form.remarks.trim(),
    movies: previewMovies.value.map((movie) => ({
      movieType: movie.movieType,
      movieJavId: movie.movieJavId,
      fc2MovieId: movie.fc2MovieId,
      isSc: movie.isSc === GLIST_IS_SC ? GLIST_IS_SC : GLIST_IS_NOT_SC
    }))
  };
}

async function submitAdd() {
  if (!Object.keys(previewData.value).length) {
    previewMsg.value = "请先执行 Add Preview";
    previewOk.value = false;
    return;
  }
  let body: ReturnType<typeof buildPayload>;
  try {
    body = buildPayload();
  } catch (err) {
    previewMsg.value = err instanceof Error ? err.message : "提交参数有误";
    previewOk.value = false;
    return;
  }
  try {
    const data = asRecord(await sendJSON("/api/triggers/sc/add", "POST", body));
    previewMsg.value = "AddSc 任务已启动，正在跳转任务列表";
    previewOk.value = true;
    message.value = "AddSc 任务已启动";
    messageOk.value = true;
    const jobID = text(data, "job_id", "JobID");
    if (jobID) {
      window.setTimeout(() => {
        window.location.href = `/crawltasklist?job_id=${encodeURIComponent(jobID)}`;
      }, 300);
    }
  } catch (err) {
    previewMsg.value = `异常：${err instanceof Error ? err.message : String(err)}`;
    previewOk.value = false;
  }
}
</script>
