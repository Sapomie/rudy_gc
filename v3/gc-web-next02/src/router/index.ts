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
import SHMagnetPage from "@/pages/SHMagnetPage.vue";
import FC2Page from "@/pages/FC2Page.vue";
import FC2DetailPage from "@/pages/FC2DetailPage.vue";
import FC2CardsPage from "@/pages/FC2CardsPage.vue";
import WDirPage from "@/pages/WDirPage.vue";
import ScEventsPage from "@/pages/ScEventsPage.vue";
import ScPickSmartPage from "@/pages/ScPickSmartPage.vue";
import AggTriggersPage from "@/pages/AggTriggersPage.vue";
import ScTriggersPage from "@/pages/ScTriggersPage.vue";
import MediaIngestPage from "@/pages/MediaIngestPage.vue";
import FC2MediaIngestPage from "@/pages/FC2MediaIngestPage.vue";
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

const fc2CardRoutes = [
  { path: "/fclistcard", apiPath: "/fc2/cards", title: "FC2 Cards" },
  { path: "/fclistcardowned", apiPath: "/fc2/cards/owned", title: "FC2 Cards MediaOwned" },
  { path: "/fclistcardrandom", apiPath: "/fc2/cards/random", title: "FC2 Cards Random" }
].map(({ path, apiPath, title }) => ({ path, component: FC2CardsPage, meta: { title, apiPath } }));

const listRoutes = [
  { path: "/javbuslist", apiPath: "/fetch-site-javbus-list", title: "JavBus List" },
  { path: "/sukebeilist", apiPath: "/fetch-site-sukebei-list", title: "Sukebei List" },
  { path: "/sehuatanglist", apiPath: "/fetch-site-sehuatang-list", title: "Sehuatang List" }
].map(({ path, apiPath, title }) => ({ path, component: ListPage, meta: { title, apiPath } }));

const triggerRoutes = [
  { path: "/dailybestfetch", apiPath: "/triggers/dailybest", title: "DailyBest" },
  { path: "/seedfetch", apiPath: "/triggers/seeds", title: "Seeds" },
  { path: "/postprocess", apiPath: "/triggers/post-process", title: "PostProcess" },
  { path: "/databackfill", apiPath: "/triggers/backfill", title: "Backfill" },
  { path: "/sitefetch", apiPath: "/triggers/fetch-site", title: "FetchSite" },
  { path: "/sehuatangfetch", apiPath: "/triggers/fetch-sehuatang", title: "FetchSehuatang" },
  { path: "/javbusfetch", apiPath: "/triggers/fetch-site-javbus-filtered", title: "JavBus Filtered" },
  { path: "/sukebeifetch", apiPath: "/triggers/fetch-site-sukebei-filtered", title: "Sukebei Filtered" }
].map(({ path, apiPath, title }) => ({ path, component: TriggerPage, meta: { title, apiPath } }));

const routes: RouteRecordRaw[] = [
  { path: "/", redirect: "/listcardtoday" },
  ...cardRoutes,
  ...fc2CardRoutes,
  { path: "/listcardagg", component: CardsAggPage, meta: { title: "cardsagg", apiPath: "/cardsagg" } },
  { path: "/listcarddayrank", component: RankPage, meta: { title: "MovieCard DayRank", apiPath: "/moviecarddayrank" } },
  { path: "/listcardperiodrank", component: RankPage, meta: { title: "MovieCard PeriodRank", apiPath: "/moviecardperiodrank" } },
  { path: "/moviedetail/:movieName", component: MoviePage, meta: { title: "Movie Detail", apiPath: "/movie/:movieName" } },
  { path: "/castdetail/:personId", component: CastPage, meta: { title: "Cast", apiPath: "/cast/:personId" } },
  { path: "/aggrelease", component: MovieAggPage, meta: { title: "MovieAgg Release", apiPath: "/movie-agg-all/release" } },
  { path: "/bucketsrelease", component: ReleaseBucketListPage, meta: { title: "MovieAgg Release Bucket", apiPath: "/movie-agg-all/release-bucket-list" } },
  { path: "/bucketsbirth", component: ReleaseBucketListPage, meta: { title: "WMedia Bucket", apiPath: "/w-media-agg/bucket-list" } },
  { path: "/shbucketlist", component: SHBucketListPage, meta: { title: "SH 聚合桶", apiPath: "/sh-thread-agg/bucket-list" } },
  { path: "/aggbirth", component: MovieAggPage, meta: { title: "WMedia Birth", apiPath: "/w-media-agg/birth" } },
  { path: "/mediadir", component: WDirPage, meta: { title: "Media Directory", apiPath: "/wdir" } },
  { path: "/scevents", component: ScEventsPage, meta: { title: "SC Events", apiPath: "/sc-events" } },
  { path: "/sccards", component: ScEventsPage, meta: { title: "SC Event Cards", apiPath: "/sc-events-cards" } },
  { path: "/scevent/:eventName", component: ScEventsPage, meta: { title: "SC Event Detail", apiPath: "/sc-events/:eventName" } },
  { path: "/scpick", component: ScPickSmartPage, meta: { title: "SC Pick Smart Media", apiPath: "/sc-pick-smart-media" } },
  { path: "/aggbackfill", component: AggTriggersPage, meta: { title: "聚合回填", apiPath: "/triggers/agg" } },
  { path: "/scadd", component: ScTriggersPage, meta: { title: "SC Triggers", apiPath: "/triggers/sc" } },
  { path: "/mediaingest", component: MediaIngestPage, meta: { title: "媒体入库", apiPath: "/triggers/media" } },
  { path: "/fcingest", component: FC2MediaIngestPage, meta: { title: "FC2媒体入库", apiPath: "/triggers/fc2-media" } },
  { path: "/mediarescan", component: MediaRescanPage, meta: { title: "媒体重扫", apiPath: "/triggers/media-rescan" } },
  { path: "/mediarollback", component: MediaRollbackPage, meta: { title: "媒体回滚", apiPath: "/triggers/media-rollback" } },
  { path: "/scmove", component: ScMediaMovePage, meta: { title: "SC Media 移动", apiPath: "/triggers/sc-media" } },
  { path: "/crawlrecordlist", component: RecordsPage, meta: { title: "Crawl Records", apiPath: "/crawl-records" } },
  { path: "/medialist", component: MediasPage, meta: { title: "Media List", apiPath: "/medias" } },
  { path: "/aggeventlist", component: WAggEventsPage, meta: { title: "W Agg Events", apiPath: "/w-agg-events" } },
  { path: "/torrentalbumlist", component: TorrentAlbumsPage, meta: { title: "Torrent Albums", apiPath: "/torrent-albums" } },
  { path: "/moviealbumlist", component: MovieAlbumsPage, meta: { title: "Movie Albums", apiPath: "/movie-albums" } },
  { path: "/seedlist", component: DSeedsPage, meta: { title: "Seed Data", apiPath: "/d-seeds" } },
  { path: "/castlist", component: CastsPage, meta: { title: "Cast List", apiPath: "/casts" } },
  { path: "/eitemlist", component: EItemsPage, meta: { title: "EItems", apiPath: "/e-items" } },
  ...listRoutes,
  ...triggerRoutes,
  { path: "/shfetchall", component: TriggerPage, meta: { title: "SH 一体抓取", apiPath: "/triggers/fetch-sehuatang-sh" } },
  { path: "/shfetch", component: TriggerPage, meta: { title: "SH 列表建池", apiPath: "/triggers/fetch-sehuatang-sh-thread-list" } },
  { path: "/shfill", component: TriggerPage, meta: { title: "SH 详情补全", apiPath: "/triggers/fetch-sehuatang-sh-thread-detail" } },
  { path: "/shpool", component: ListPage, meta: { title: "SH 帖子池", apiPath: "/fetch-site-sehuatang-sh-list" } },
  { path: "/shthread/:threadTid", component: ListPage, meta: { title: "SH 帖子详情", apiPath: "/fetch-site-sehuatang-sh-detail/:threadTid" } },
  { path: "/shmagnetlist", component: SHMagnetPage, meta: { title: "SH 磁力", apiPath: "/sh-magnets" } },
  { path: "/fclist", component: FC2Page, meta: { title: "FC2影片", apiPath: "/fc2/movies" } },
  { path: "/fcdetail/:fc2Key", component: FC2DetailPage, meta: { title: "FC2影片详情", apiPath: "/fc2/movies/:fc2Key" } },
  { path: "/siteconfiglist", component: FetchSitesPage, meta: { title: "FetchSites", apiPath: "/fetch-sites" } },
  { path: "/crawltasklist", component: CrawlerTasksPage, meta: { title: "Crawler Tasks", apiPath: "/crawler/tasks" } },
  { path: "/detailfetch", component: CrawlerDetailLoopPage, meta: { title: "Detail Loop", apiPath: "/crawler/detail-loop/page" } },
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
