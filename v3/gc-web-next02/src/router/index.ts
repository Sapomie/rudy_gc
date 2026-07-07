import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import HomePage from "@/pages/HomePage.vue";
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

const cardRoutes = [
  "/cards",
  "/cardstoday",
  "/cardshasrank",
  "/cardsmediamowned",
  "/cardsneeddownload",
  "/cardsrandom"
].map((path) => ({ path, component: CardsPage, meta: { title: path.slice(1) } }));

const listRoutes = [
  "/fetch-site-javbus-list",
  "/fetch-site-sukebei-list",
  "/fetch-site-sehuatang-list",
  "/fetch-site-sehuatang-sh-list",
  "/fetch-site-sehuatang-sh-detail/:threadTid"
].map((path) => ({ path, component: ListPage, meta: { title: path.replace("/:eventName", "").replace("/:threadTid", "").slice(1) || "List" } }));

const triggerRoutes = [
  "/triggers/dailybest",
  "/triggers/seeds",
  "/triggers/post-process",
  "/triggers/backfill",
  "/triggers/fetch-site",
  "/triggers/fetch-sehuatang",
  "/triggers/fetch-sehuatang-sh",
  "/triggers/fetch-sehuatang-sh-thread-list",
  "/triggers/fetch-sehuatang-sh-thread-detail",
  "/triggers/fetch-site-javbus-filtered",
  "/triggers/fetch-site-sukebei-filtered"
].map((path) => ({ path, component: TriggerPage, meta: { title: path.slice(1) } }));

const routes: RouteRecordRaw[] = [
  { path: "/", component: HomePage, meta: { title: "Dashboard" } },
  ...cardRoutes,
  { path: "/cardsagg", component: CardsAggPage, meta: { title: "cardsagg" } },
  { path: "/moviecarddayrank", component: RankPage, meta: { title: "MovieCard DayRank" } },
  { path: "/moviecardperiodrank", component: RankPage, meta: { title: "MovieCard PeriodRank" } },
  { path: "/movie/:movieName", component: MoviePage, meta: { title: "Movie Detail" } },
  { path: "/cast", component: CastPage, meta: { title: "Cast" } },
  { path: "/cast/:personId", component: CastPage, meta: { title: "Cast" } },
  { path: "/movie-agg-all/release", component: MovieAggPage, meta: { title: "MovieAgg Release" } },
  { path: "/movie-agg-all/release-bucket-list", component: ReleaseBucketListPage, meta: { title: "MovieAgg Release Bucket" } },
  { path: "/w-media-agg/bucket-list", component: ReleaseBucketListPage, meta: { title: "WMedia Bucket" } },
  { path: "/w-media-agg/birth", component: MovieAggPage, meta: { title: "WMedia Birth" } },
  { path: "/wdir", component: WDirPage, meta: { title: "WDirectory" } },
  { path: "/sc-events", component: ScEventsPage, meta: { title: "SC Events" } },
  { path: "/sc-events-cards", component: ScEventsPage, meta: { title: "SC Event Cards" } },
  { path: "/sc-events/:eventName", component: ScEventsPage, meta: { title: "SC Event Detail" } },
  { path: "/sc-pick-smart-media", component: ScPickSmartPage, meta: { title: "SC Pick Smart Media" } },
  { path: "/triggers/agg", component: AggTriggersPage, meta: { title: "聚合回填" } },
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
  { path: "/fetch-sites", component: FetchSitesPage, meta: { title: "FetchSites" } },
  { path: "/crawler/tasks", component: CrawlerTasksPage, meta: { title: "Crawler Tasks" } },
  { path: "/crawler/detail-loop", component: CrawlerDetailLoopPage, meta: { title: "Detail Loop", apiPath: "/crawler/detail-loop/page" } },
  { path: "/:pathMatch(.*)*", component: NotFoundPage, meta: { title: "Not Found" } }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.afterEach((to) => {
  document.title = `${String(to.meta.title || "Rudy-GC")} - Next02`;
});

export default router;
