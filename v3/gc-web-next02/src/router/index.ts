import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import CardsPage from "@/pages/CardsPage.vue";
import CardsAggPage from "@/pages/CardsAggPage.vue";
import RankPage from "@/pages/RankPage.vue";
import MoviePage from "@/pages/MoviePage.vue";
import CastPage from "@/pages/CastPage.vue";
import RecordsPage from "@/pages/RecordsPage.vue";
import MediasPage from "@/pages/MediasPage.vue";
import WAggEventsPage from "@/pages/WAggEventsPage.vue";
import TorrentAlbumsPage from "@/pages/TorrentAlbumsPage.vue";
import MovieAlbumsPage from "@/pages/MovieAlbumsPage.vue";
import DSeedsPage from "@/pages/DSeedsPage.vue";
import CastsPage from "@/pages/CastsPage.vue";
import EItemsPage from "@/pages/EItemsPage.vue";
import ListPage from "@/pages/ListPage.vue";
import MovieAggPage from "@/pages/MovieAggPage.vue";
import ReleaseBucketListPage from "@/pages/ReleaseBucketListPage.vue";
import SHBucketListPage from "@/pages/SHBucketListPage.vue";
import WDirPage from "@/pages/WDirPage.vue";
import ScEventsPage from "@/pages/ScEventsPage.vue";
import ScPickSmartPage from "@/pages/ScPickSmartPage.vue";
import AggTriggersPage from "@/pages/AggTriggersPage.vue";
import ScTriggersPage from "@/pages/ScTriggersPage.vue";
import MediaIngestPage from "@/pages/MediaIngestPage.vue";
import MediaRescanPage from "@/pages/MediaRescanPage.vue";
import MediaRollbackPage from "@/pages/MediaRollbackPage.vue";
import ScMediaMovePage from "@/pages/ScMediaMovePage.vue";
import TriggerPage from "@/pages/TriggerPage.vue";
import FetchSitesPage from "@/pages/FetchSitesPage.vue";
import CrawlerTasksPage from "@/pages/CrawlerTasksPage.vue";
import CrawlerDetailLoopPage from "@/pages/CrawlerDetailLoopPage.vue";
import NotFoundPage from "@/pages/NotFoundPage.vue";
import { compactRouteQuery, type CompactQuery } from "@/utils/query";

const cardRoutes = [
  { path: "/listcard", apiPath: "/cards", title: "Cards" },
  { path: "/listcardtoday", apiPath: "/cardstoday", title: "Cards Today" },
  { path: "/listcardrank", apiPath: "/cardshasrank", title: "Cards HasRank" },
  { path: "/listcardowned", apiPath: "/cardsmediamowned", title: "Cards MediaOwned" },
  { path: "/listcardneed", apiPath: "/cardsneeddownload", title: "Cards NeedDownload" },
  { path: "/listcardrandom", apiPath: "/cardsrandom", title: "Cards Random" }
].map(({ path, apiPath, title }) => ({ path, component: CardsPage, meta: { title, apiPath } }));

const listRoutes = [
  "/fetch-site-javbus-list",
  "/fetch-site-sukebei-list",
  "/fetch-site-sehuatang-list",
].map((path) => ({ path, component: ListPage, meta: { title: path.replace("/:eventName", "").replace("/:threadTid", "").slice(1) || "List" } }));

const triggerRoutes = [
  "/triggers/dailybest",
  "/triggers/seeds",
  "/triggers/post-process",
  "/triggers/backfill",
  "/triggers/fetch-site",
  "/triggers/fetch-sehuatang",
  "/triggers/fetch-site-javbus-filtered",
  "/triggers/fetch-site-sukebei-filtered"
].map((path) => ({ path, component: TriggerPage, meta: { title: path.slice(1) } }));

const routes: RouteRecordRaw[] = [
  { path: "/", redirect: "/listcardtoday" },
  ...cardRoutes,
  { path: "/listcardagg", component: CardsAggPage, meta: { title: "cardsagg", apiPath: "/cardsagg" } },
  { path: "/listcarddayrank", component: RankPage, meta: { title: "MovieCard DayRank", apiPath: "/moviecarddayrank" } },
  { path: "/listcardperiodrank", component: RankPage, meta: { title: "MovieCard PeriodRank", apiPath: "/moviecardperiodrank" } },
  { path: "/moviedetail/:movieName", component: MoviePage, meta: { title: "Movie Detail", apiPath: "/movie/:movieName" } },
  { path: "/cast", component: CastPage, meta: { title: "Cast" } },
  { path: "/castdetail/:personId", component: CastPage, meta: { title: "Cast", apiPath: "/cast/:personId" } },
  { path: "/aggrelease", component: MovieAggPage, meta: { title: "MovieAgg Release", apiPath: "/movie-agg-all/release" } },
  { path: "/bucketsrelease", component: ReleaseBucketListPage, meta: { title: "MovieAgg Release Bucket", apiPath: "/movie-agg-all/release-bucket-list" } },
  { path: "/bucketsbirth", component: ReleaseBucketListPage, meta: { title: "WMedia Bucket", apiPath: "/w-media-agg/bucket-list" } },
  { path: "/bucketssh", component: SHBucketListPage, meta: { title: "SH 聚合桶", apiPath: "/sh-thread-agg/bucket-list" } },
  { path: "/aggbirth", component: MovieAggPage, meta: { title: "WMedia Birth", apiPath: "/w-media-agg/birth" } },
  { path: "/mediadir", component: WDirPage, meta: { title: "Media Directory", apiPath: "/wdir" } },
  { path: "/sc-events", component: ScEventsPage, meta: { title: "SC Events" } },
  { path: "/sc-events-cards", component: ScEventsPage, meta: { title: "SC Event Cards" } },
  { path: "/sc-events/:eventName", component: ScEventsPage, meta: { title: "SC Event Detail" } },
  { path: "/sc-pick-smart-media", component: ScPickSmartPage, meta: { title: "SC Pick Smart Media" } },
  { path: "/aggbackfill", component: AggTriggersPage, meta: { title: "聚合回填", apiPath: "/triggers/agg" } },
  { path: "/triggers/sc", component: ScTriggersPage, meta: { title: "SC Triggers" } },
  { path: "/triggers/media", component: MediaIngestPage, meta: { title: "媒体入库" } },
  { path: "/triggers/media-rescan", component: MediaRescanPage, meta: { title: "媒体重扫" } },
  { path: "/triggers/media-rollback", component: MediaRollbackPage, meta: { title: "媒体回滚" } },
  { path: "/triggers/sc-media", component: ScMediaMovePage, meta: { title: "SC Media 移动" } },
  { path: "/crawl-records", component: RecordsPage, meta: { title: "crawl-records" } },
  { path: "/records", component: RecordsPage, meta: { title: "records", apiPath: "/crawl-records" } },
  { path: "/medias", component: MediasPage, meta: { title: "medias" } },
  { path: "/w-agg-events", component: WAggEventsPage, meta: { title: "w-agg-events" } },
  { path: "/torrent-albums", component: TorrentAlbumsPage, meta: { title: "torrent-albums" } },
  { path: "/movie-albums", component: MovieAlbumsPage, meta: { title: "movie-albums" } },
  { path: "/d-seeds", component: DSeedsPage, meta: { title: "d-seeds" } },
  { path: "/casts", component: CastsPage, meta: { title: "casts" } },
  { path: "/e-items", component: EItemsPage, meta: { title: "e-items" } },
  ...listRoutes,
  ...triggerRoutes,
  { path: "/shfetchall", component: TriggerPage, meta: { title: "SH 一体抓取", apiPath: "/triggers/fetch-sehuatang-sh" } },
  { path: "/shfetch", component: TriggerPage, meta: { title: "SH 列表建池", apiPath: "/triggers/fetch-sehuatang-sh-thread-list" } },
  { path: "/shfill", component: TriggerPage, meta: { title: "SH 详情补全", apiPath: "/triggers/fetch-sehuatang-sh-thread-detail" } },
  { path: "/shpool", component: ListPage, meta: { title: "SH 帖子池", apiPath: "/fetch-site-sehuatang-sh-list" } },
  { path: "/shthread/:threadTid", component: ListPage, meta: { title: "SH 帖子详情", apiPath: "/fetch-site-sehuatang-sh-detail/:threadTid" } },
  { path: "/fetch-sites", component: FetchSitesPage, meta: { title: "FetchSites" } },
  { path: "/crawler/tasks", component: CrawlerTasksPage, meta: { title: "Crawler Tasks" } },
  { path: "/crawler/detail-loop", component: CrawlerDetailLoopPage, meta: { title: "Detail Loop", apiPath: "/crawler/detail-loop/page" } },
  { path: "/:pathMatch(.*)*", component: NotFoundPage, meta: { title: "Not Found" } }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

function querySignature(query: Record<string, unknown>): string {
  const params = new URLSearchParams();
  Object.keys(query).sort().forEach((key) => {
    const rawValues = Array.isArray(query[key]) ? query[key] as unknown[] : [query[key]];
    rawValues.forEach((value) => params.append(key, value === undefined || value === null ? "" : String(value)));
  });
  return params.toString();
}

router.beforeEach((to) => {
  const compacted = compactRouteQuery(to.path, to.query);
  if (querySignature(to.query) === querySignature(compacted as CompactQuery)) {
    return true;
  }
  return {
    path: to.path,
    query: compacted,
    hash: to.hash,
    replace: true
  };
});

router.afterEach((to) => {
  document.title = `${String(to.meta.title || "Rudy-GC")} - Next02`;
});

export default router;
