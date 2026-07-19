import { defineFeatureRoutes } from "@/shared/router/routes";

const cardRoutes = [
  { path: "/listcard", name: "movie-cards", apiPath: "/cards", title: "Cards" },
  { path: "/listcardtoday", name: "movie-cards-today", apiPath: "/cardstoday", title: "Cards Today" },
  { path: "/listcardrank", name: "movie-cards-ranked", apiPath: "/cardshasrank", title: "Cards HasRank" },
  { path: "/listcardowned", name: "movie-cards-owned", apiPath: "/cardsmediamowned", title: "Cards MediaOwned" },
  { path: "/listcardneed", name: "movie-cards-needed", apiPath: "/cardsneeddownload", title: "Cards NeedDownload" },
  { path: "/listcardrandom", name: "movie-cards-random", apiPath: "/cardsrandom", title: "Cards Random" }
].map((route) => ({ ...route, component: () => import("./pages/CardsPage.vue") }));

export const movieRoutes = defineFeatureRoutes([
  ...cardRoutes,
  {
    path: "/listcardagg",
    name: "movie-cards-aggregate",
    component: () => import("./pages/CardsAggPage.vue"),
    title: "cardsagg",
    apiPath: "/cardsagg"
  },
  {
    path: "/listcarddayrank",
    name: "movie-cards-day-rank",
    component: () => import("./pages/RankPage.vue"),
    title: "MovieCard DayRank",
    apiPath: "/moviecarddayrank"
  },
  {
    path: "/listcardperiodrank",
    name: "movie-cards-period-rank",
    component: () => import("./pages/RankPage.vue"),
    title: "MovieCard PeriodRank",
    apiPath: "/moviecardperiodrank"
  },
  {
    path: "/moviedetail/:movieName",
    name: "movie-detail",
    component: () => import("./pages/MoviePage.vue"),
    title: "Movie Detail",
    apiPath: "/movie/:movieName"
  },
  {
    path: "/castdetail/:personId",
    name: "cast-detail",
    component: () => import("./pages/CastPage.vue"),
    title: "Cast",
    apiPath: "/cast/:personId"
  },
  {
    path: "/castlist",
    name: "casts",
    component: () => import("./pages/CastsPage.vue"),
    title: "Cast List",
    apiPath: "/casts"
  },
  {
    path: "/aggrelease",
    name: "movie-aggregate-release",
    component: () => import("./pages/MovieAggPage.vue"),
    title: "MovieAgg Release",
    apiPath: "/movie-agg-all/release"
  },
  {
    path: "/aggbirth",
    name: "movie-aggregate-birth",
    component: () => import("./pages/MovieAggPage.vue"),
    title: "WMedia Birth",
    apiPath: "/w-media-agg/birth"
  },
  {
    path: "/bucketsrelease",
    name: "movie-release-buckets",
    component: () => import("./pages/ReleaseBucketListPage.vue"),
    title: "MovieAgg Release Bucket",
    apiPath: "/movie-agg-all/release-bucket-list"
  },
  {
    path: "/bucketsbirth",
    name: "movie-birth-buckets",
    component: () => import("./pages/ReleaseBucketListPage.vue"),
    title: "WMedia Bucket",
    apiPath: "/w-media-agg/bucket-list"
  },
  {
    path: "/torrentalbumlist",
    name: "torrent-albums",
    component: () => import("./pages/TorrentAlbumsPage.vue"),
    title: "Torrent Albums",
    apiPath: "/torrent-albums"
  },
  {
    path: "/moviealbumlist",
    name: "movie-albums",
    component: () => import("./pages/MovieAlbumsPage.vue"),
    title: "Movie Albums",
    apiPath: "/movie-albums"
  }
]);
