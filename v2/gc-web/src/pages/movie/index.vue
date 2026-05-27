<template>
  <PageFrame :title="titleText" eyebrow="Movie Detail" :description="descriptionText">
    <template #actions>
      <RouterLink class="rgc-button" to="/cards">Cards</RouterLink>
      <a v-if="movie?.bus_url" class="rgc-button" :href="movie.bus_url" target="_blank" rel="noopener">Bus</a>
      <a v-if="movie?.search_url" class="rgc-button" :href="movie.search_url" target="_blank" rel="noopener">Suk</a>
      <a v-if="movie?.jav_url" class="rgc-button" :href="movie.jav_url" target="_blank" rel="noopener">Lib</a>
    </template>

    <StateBlock v-if="loading" title="加载中" message="正在读取电影详情。" />
    <StateBlock v-else-if="errorMessage" title="加载失败" :message="errorMessage" tone="error" />

    <template v-else-if="detail && movie">
      <section class="rgc-detail-hero">
        <article class="rgc-detail-poster-panel">
          <div class="rgc-detail-titlebar">
            <h2 :title="movie.title">{{ movie.title || movie.movie_name }}</h2>
            <div class="rgc-detail-album-line">
              <RouterLink v-for="album in selectedMovieAlbums" :key="album.id || album.name" :to="`/movie-albums?album_name=${encodeURIComponent(album.name)}`">
                {{ album.name }}
              </RouterLink>
              <span v-if="selectedMovieAlbums.length === 0">-</span>
            </div>
          </div>

          <div class="rgc-detail-poster">
            <img v-if="movie.jacket_img && !imageFailed" :src="movie.jacket_img" :alt="`${movie.movie_name} 海报`" @error="imageFailed = true" />
            <span v-else>{{ movie.movie_name }}</span>
            <div class="rgc-detail-badge-stack">
              <b v-if="movie.come_times > 0" class="come">❤ {{ movie.come_times }}</b>
              <b v-if="movie.sc_times > 0" class="sc">× {{ movie.sc_times }}</b>
            </div>
            <b v-if="Number.isFinite(movie.score)" class="rgc-detail-score">★ {{ movie.score }}</b>
          </div>
        </article>

        <article class="rgc-detail-info-panel">
          <nav class="rgc-detail-tabs" aria-label="详情切换">
            <button type="button" :class="{ active: infoTab === 'info' }" @click="infoTab = 'info'">Info</button>
            <button type="button" :class="{ active: infoTab === 'rank' }" @click="infoTab = 'rank'">Rank</button>
          </nav>

          <div v-if="infoTab === 'info'" class="rgc-detail-table">
            <div>
              <span>识别码</span>
              <div class="rgc-detail-inline">
                <RouterLink :to="movie.prefix_href || `/cards?pn=${encodeURIComponent(movie.prefix)}`">{{ movie.movie_name }}</RouterLink>
                <a v-if="mediaHref" :href="mediaHref" :class="['rgc-detail-chip', mediaTone]">MPlay</a>
                <span v-else-if="movie.owned_w_media === 6" class="rgc-detail-chip muted">MRemoved</span>
                <small v-if="movie.encode_name">({{ movie.encode_name }})</small>
              </div>
            </div>
            <div><span>发行时间</span><strong>{{ movie.releasing_date || "-" }}</strong></div>
            <div><span>长度</span><strong>{{ movie.length || "-" }}</strong></div>
            <div>
              <span>制作商</span>
              <RouterLink :to="movie.maker_href || `/cards?mn=${encodeURIComponent(movie.maker)}`">{{ movie.maker || "-" }}</RouterLink>
            </div>
            <div>
              <span>发行商</span>
              <RouterLink :to="movie.label_href || `/cards?ln=${encodeURIComponent(movie.label)}`">{{ movie.label || "-" }}</RouterLink>
            </div>
            <div>
              <span>导演</span>
              <RouterLink :to="movie.director_href || `/cards?dn=${encodeURIComponent(movie.director)}`">{{ movie.director || "-" }}</RouterLink>
            </div>
            <div>
              <span>分类</span>
              <div class="rgc-detail-tags">
                <RouterLink v-for="genre in movie.genre || []" :key="genre" :to="`/cards?gn=${encodeURIComponent(genre)}`">{{ genre }}</RouterLink>
                <small v-if="(movie.genre || []).length === 0">-</small>
              </div>
            </div>
            <div>
              <span>演员</span>
              <div class="rgc-detail-tags">
                <RouterLink v-for="cast in movie.cast || []" :key="`${cast.person_id}-${cast.name}`" :to="cast.href || `/cards?cn=${encodeURIComponent(cast.name)}`">
                  {{ cast.name_show || cast.display_name || cast.name }}
                </RouterLink>
                <small v-if="(movie.cast || []).length === 0">-</small>
              </div>
            </div>
            <div>
              <span>评分</span>
              <strong>★ {{ movie.score }} <small>({{ movie.viewers_number_watched }}) {{ movie.update_date }}</small></strong>
            </div>
            <div>
              <span>链接</span>
              <div class="rgc-detail-inline">
                <a v-if="movie.bus_url" :href="movie.bus_url" target="_blank" rel="noopener">Bus</a>
                <a v-if="movie.search_url" :href="movie.search_url" target="_blank" rel="noopener">Suk</a>
                <a v-if="movie.jav_url" :href="movie.jav_url" target="_blank" rel="noopener">Lib</a>
              </div>
            </div>
            <div><span>RankInfo</span><strong>{{ movie.first_ranking_day || "-" }}</strong></div>
            <div>
              <span>Sc</span>
              <div class="rgc-detail-tags">
                <RouterLink v-for="event in detail.sc_events || []" :key="`${event.sc_name}-${event.sc_time}`" :to="event.href">{{ event.sc_name }}</RouterLink>
                <small v-if="(detail.sc_events || []).length === 0">-</small>
              </div>
            </div>
            <div>
              <span>电影相册</span>
              <div class="rgc-detail-tags">
                <RouterLink v-for="album in selectedMovieAlbums" :key="`info-${album.id || album.name}`" :to="`/movie-albums?album_name=${encodeURIComponent(album.name)}`">
                  {{ album.name }}
                </RouterLink>
                <small v-if="selectedMovieAlbums.length === 0">-</small>
              </div>
            </div>
          </div>

          <div v-else class="rgc-detail-rank-wrap">
            <table class="rgc-detail-data-table">
              <thead><tr><th>Date</th><th>Pts</th></tr></thead>
              <tbody>
                <tr v-for="rank in detail.rank_infos || []" :key="rank.date">
                  <td><RouterLink :to="rank.href">{{ rank.date }}</RouterLink></td>
                  <td>{{ rank.rank }}</td>
                </tr>
                <tr v-if="(detail.rank_infos || []).length === 0"><td colspan="2">暂无排名</td></tr>
              </tbody>
            </table>
          </div>
        </article>
      </section>

      <section v-if="detail.media && movie.owned_w_media !== 6" class="rgc-detail-panel">
        <div class="rgc-detail-section-head">
          <h2>视频信息</h2>
          <span v-if="detail.marked_delete" class="rgc-detail-status danger">已在待删除</span>
        </div>
        <div class="rgc-detail-table-scroll">
          <table class="rgc-detail-data-table">
            <thead>
              <tr>
                <th>Source</th>
                <th>TorrentHash</th>
                <th>BirthTime</th>
                <th>Size</th>
                <th>Height</th>
                <th>BitRate</th>
                <th>时长(分钟)</th>
                <th>Frame</th>
                <th>SelfMake</th>
                <th>Dir</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>WMedia</td>
                <td><span class="rgc-detail-hash">{{ detail.media.source_torrent_hash || "-" }}</span></td>
                <td>{{ detail.media.birth_time }}</td>
                <td><a :href="detail.media.file_path">{{ formatFixed(detail.media.size, 2) }}</a></td>
                <td>{{ detail.media.height }}</td>
                <td>{{ formatFixed(detail.media.bit_rate, 0) }}</td>
                <td>{{ formatFixed(detail.media.duration_minutes, 2) }}</td>
                <td>{{ detail.media.frame }}</td>
                <td>{{ detail.media.self_make }}</td>
                <td><a :href="detail.media.directory">{{ detail.media.directory }}</a></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="rgc-detail-panel">
        <div class="rgc-detail-section-head">
          <h2>资源列表</h2>
          <nav class="rgc-detail-tabs compact" aria-label="资源切换">
            <button v-for="tab in resourceTabs" :key="tab.key" type="button" :class="{ active: activeResourceTab === tab.key }" @click="activeResourceTab = tab.key">
              {{ tab.label }} <small>({{ tab.count }})</small>
            </button>
          </nav>
        </div>

        <div class="rgc-detail-table-scroll">
          <table v-if="activeResourceTab === 'javbus'" class="rgc-detail-data-table">
            <thead>
              <tr>
                <th>名称</th>
                <th><button type="button" :class="sortClass('javbus', 'size')" @click="setSort('javbus', 'size')">大小 {{ sortMark("javbus", "size") }}</button></th>
                <th><button type="button" :class="sortClass('javbus', 'date')" @click="setSort('javbus', 'date')">日期 {{ sortMark("javbus", "date") }}</button></th>
                <th>标签</th>
                <th>Info Hash</th>
                <th>相册</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in sortedJavbus" :key="row.row_id" :class="{ 'is-match': row.has_matched_hash }">
                <td>{{ row.magnet_name }}</td>
                <td>{{ row.size_text }}</td>
                <td>{{ row.share_date }}</td>
                <td><span v-if="row.has_subtitle" class="rgc-detail-status info">字幕</span><span v-else>-</span></td>
                <td><span :class="hashClass(row.info_hash)">{{ row.info_hash }}</span></td>
                <td><ResourceStatus :download="row.is_in_download" :pending="row.is_in_pending" /></td>
              </tr>
              <tr v-if="sortedJavbus.length === 0"><td colspan="6">暂无 JavBus 磁力数据</td></tr>
            </tbody>
          </table>

          <table v-else-if="activeResourceTab === 'sukebei'" class="rgc-detail-data-table">
            <thead>
              <tr>
                <th>标题</th>
                <th><button type="button" :class="sortClass('sukebei', 'size')" @click="setSort('sukebei', 'size')">大小 {{ sortMark("sukebei", "size") }}</button></th>
                <th><button type="button" :class="sortClass('sukebei', 'date')" @click="setSort('sukebei', 'date')">日期 {{ sortMark("sukebei", "date") }}</button></th>
                <th><button type="button" :class="sortClass('sukebei', 'downloads')" @click="setSort('sukebei', 'downloads')">下载 {{ sortMark("sukebei", "downloads") }}</button></th>
                <th>Info Hash</th>
                <th>相册</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in sortedSukebei" :key="row.row_id" :class="{ 'is-match': row.has_matched_hash }">
                <td><a v-if="row.view_url" :href="row.view_url" target="_blank" rel="noopener">{{ row.torrent_title }}</a><span v-else>{{ row.torrent_title }}</span></td>
                <td>{{ row.size_text }}</td>
                <td>{{ row.publish_time }}</td>
                <td>{{ row.completed }}</td>
                <td><span :class="hashClass(row.info_hash)">{{ row.info_hash }}</span></td>
                <td><ResourceStatus :download="row.is_in_download" :pending="row.is_in_pending" /></td>
              </tr>
              <tr v-if="sortedSukebei.length === 0"><td colspan="6">暂无 Sukebei 种子数据</td></tr>
            </tbody>
          </table>

          <table v-else class="rgc-detail-data-table">
            <thead>
              <tr>
                <th>帖子标题</th>
                <th><button type="button" :class="sortClass('sehuatang', 'date')" @click="setSort('sehuatang', 'date')">发帖时间 {{ sortMark("sehuatang", "date") }}</button></th>
                <th>Info Hash</th>
                <th>相册</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in sortedSehuatang" :key="row.row_id" :class="{ 'is-match': row.has_matched_hash }">
                <td><a v-if="row.thread_url" :href="row.thread_url" target="_blank" rel="noopener">{{ row.thread_title }}</a><span v-else>{{ row.thread_title }}</span></td>
                <td>{{ row.post_time }}</td>
                <td><span :class="hashClass(row.info_hash)">{{ row.info_hash }}</span></td>
                <td><ResourceStatus :download="row.is_in_download" :pending="row.is_in_pending" /></td>
              </tr>
              <tr v-if="sortedSehuatang.length === 0"><td colspan="4">暂无 Sehuatang 磁力数据</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="rgc-detail-fetch-grid">
        <FetchStatusCard title="JavBus 抓取状态" :status="detail.javbus_fetch" />
        <FetchStatusCard title="Sukebei 抓取状态" :status="detail.sukebei_fetch" />
      </section>
    </template>
  </PageFrame>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

import { getMovieDetail } from "@/api/gc";
import PageFrame from "@/components/PageFrame.vue";
import StateBlock from "@/components/StateBlock.vue";
import type {
  MovieDetailResponse,
  MovieFetchSiteStatus,
  MovieJavbusMagnet,
  MovieSehuatangMagnet,
  MovieSukebeiTorrent,
} from "@/types/api";

type ResourceTab = "javbus" | "sukebei" | "sehuatang";
type SortKey = "size" | "date" | "downloads";

const route = useRoute();
const detail = ref<MovieDetailResponse | null>(null);
const loading = ref(false);
const errorMessage = ref("");
const imageFailed = ref(false);
const infoTab = ref<"info" | "rank">("info");
const activeResourceTab = ref<ResourceTab>("javbus");
const sortState = ref<{ tab: ResourceTab; key: SortKey; order: "asc" | "desc" }>({ tab: "javbus", key: "date", order: "desc" });

const movieParam = computed(() => String(route.params.movie || ""));
const movie = computed(() => detail.value?.movie || null);
const titleText = computed(() => movie.value?.title || movie.value?.movie_name || movieParam.value);
const descriptionText = computed(() => (movie.value ? `${movie.value.movie_name} · ${movie.value.releasing_date || "-"}` : "电影详情"));
const selectedMovieAlbums = computed(() => (detail.value?.movie_albums || []).filter((album) => album.selected));
const mediaHref = computed(() => {
  if (!movie.value || movie.value.owned_w_media === 6) {
    return "";
  }
  return movie.value.video_url_w_media || "";
});
const mediaTone = computed(() => (movie.value?.owned_w_media === 5 ? "info" : "success"));
const resourceTabs = computed(() => [
  { key: "javbus" as const, label: "JavBus 磁力", count: detail.value?.javbus_magnets?.length || 0 },
  { key: "sukebei" as const, label: "Sukebei 种子", count: detail.value?.sukebei_rows?.length || 0 },
  { key: "sehuatang" as const, label: "Sehuatang 磁力", count: detail.value?.sehuatang_rows?.length || 0 },
]);
const sortedJavbus = computed(() => sortRows(detail.value?.javbus_magnets || [], "javbus"));
const sortedSukebei = computed(() => sortRows(detail.value?.sukebei_rows || [], "sukebei"));
const sortedSehuatang = computed(() => sortRows(detail.value?.sehuatang_rows || [], "sehuatang"));

watch(() => route.fullPath, load);
onMounted(load);

async function load() {
  loading.value = true;
  errorMessage.value = "";
  imageFailed.value = false;
  try {
    detail.value = await getMovieDetail(movieParam.value);
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function setSort(tab: ResourceTab, key: SortKey) {
  const current = sortState.value;
  if (current.tab === tab && current.key === key) {
    sortState.value = { tab, key, order: current.order === "desc" ? "asc" : "desc" };
    return;
  }
  sortState.value = { tab, key, order: "desc" };
}

function sortClass(tab: ResourceTab, key: SortKey) {
  return ["rgc-detail-sort", { active: sortState.value.tab === tab && sortState.value.key === key }];
}

function sortMark(tab: ResourceTab, key: SortKey) {
  if (sortState.value.tab !== tab || sortState.value.key !== key) {
    return "";
  }
  return sortState.value.order === "desc" ? "↓" : "↑";
}

function sortRows<T extends MovieJavbusMagnet | MovieSukebeiTorrent | MovieSehuatangMagnet>(rows: T[], tab: ResourceTab): T[] {
  const state = sortState.value;
  if (state.tab !== tab) {
    return [...rows];
  }
  const direction = state.order === "desc" ? -1 : 1;
  return [...rows].sort((left, right) => compareResource(left, right, state.key) * direction);
}

function compareResource(left: MovieJavbusMagnet | MovieSukebeiTorrent | MovieSehuatangMagnet, right: MovieJavbusMagnet | MovieSukebeiTorrent | MovieSehuatangMagnet, key: SortKey) {
  if (key === "size") {
    return parseSize(resourceSize(left)) - parseSize(resourceSize(right));
  }
  if (key === "downloads") {
    return resourceCompleted(left) - resourceCompleted(right);
  }
  return resourceDate(left).localeCompare(resourceDate(right));
}

function resourceSize(row: MovieJavbusMagnet | MovieSukebeiTorrent | MovieSehuatangMagnet) {
  return "size_text" in row ? row.size_text : "";
}

function resourceCompleted(row: MovieJavbusMagnet | MovieSukebeiTorrent | MovieSehuatangMagnet) {
  return "completed" in row ? row.completed : 0;
}

function resourceDate(row: MovieJavbusMagnet | MovieSukebeiTorrent | MovieSehuatangMagnet) {
  if ("share_date" in row) {
    return row.share_date || "";
  }
  if ("publish_time" in row) {
    return row.publish_time || "";
  }
  return row.post_time || "";
}

function parseSize(raw: string) {
  const match = raw.match(/^([\d.]+)\s*(TB|GB|MB|KB|B)$/i);
  if (!match) {
    return 0;
  }
  const value = Number(match[1]);
  const unit = match[2].toUpperCase();
  const multipliers: Record<string, number> = { B: 1, KB: 1024, MB: 1024 ** 2, GB: 1024 ** 3, TB: 1024 ** 4 };
  return Number.isFinite(value) ? value * multipliers[unit] : 0;
}

function hashClass(hash: string) {
  const mediaHash = detail.value?.media?.source_torrent_hash || "";
  return ["rgc-detail-hash", { match: mediaHash !== "" && mediaHash.toUpperCase() === (hash || "").toUpperCase() }];
}

function formatFixed(value: number, digits: number) {
  return Number.isFinite(value) ? value.toFixed(digits) : "-";
}

const ResourceStatus = defineComponent({
  props: {
    download: { type: Boolean, required: true },
    pending: { type: Boolean, required: true },
  },
  setup(props) {
    return () =>
      h("div", { class: "rgc-detail-resource-status" }, [
        h("span", { class: ["download", { active: props.download }] }, props.download ? "已在下载中" : "下载中"),
        h("span", { class: ["pending", { active: props.pending }] }, props.pending ? "已在待下载" : "待下载"),
      ]);
  },
});

const FetchStatusCard = defineComponent({
  props: {
    title: { type: String, required: true },
    status: { type: Object as () => MovieFetchSiteStatus | null, default: null },
  },
  setup(props) {
    return () =>
      h("article", { class: "rgc-detail-panel rgc-detail-fetch-card" }, [
        h("h2", props.title),
        props.status
          ? h("dl", [
              statusRow("状态", props.status.fetch_status_text),
              statusRow("重试次数", String(props.status.try_count)),
              statusRow("最后抓取", props.status.last_fetch_time),
              statusRow("最后抓取距离现在", props.status.last_fetch_ago),
              statusRow("Hash 数", String(props.status.torrent_hash_count)),
              statusRow("最新日期", props.status.latest_publish_time),
              statusLinkRow("来源地址", props.status.source_url),
              statusRow("错误", props.status.last_error || "-"),
            ])
          : h("p", { class: "rgc-detail-empty" }, "暂无抓取状态"),
      ]);
  },
});

function statusRow(label: string, value: string) {
  return h("div", [h("dt", label), h("dd", value || "-")]);
}

function statusLinkRow(label: string, value: string) {
  return h("div", [h("dt", label), h("dd", value ? h("a", { href: value, target: "_blank", rel: "noopener" }, value) : "-")]);
}
</script>
