import type { RouteRecordRaw } from "vue-router";

export const dailyBestRoutes: RouteRecordRaw[] = [
  {
    path: "/dailybestoverview",
    name: "dailybest-overview",
    component: () => import("@/features/dailybest/pages/OverviewPage.vue"),
    meta: { title: "运行概览" }
  },
  {
    path: "/ranklist",
    name: "dailybest-rank-list",
    component: () => import("@/features/dailybest/pages/RanksPage.vue"),
    meta: { title: "排名明细" }
  },
  {
    path: "/rankdaystats",
    name: "dailybest-rank-day-stats",
    component: () => import("@/features/dailybest/pages/RankDayStatsPage.vue"),
    meta: { title: "每日统计" }
  },
  {
    path: "/rankperiodlist",
    name: "dailybest-rank-period-list",
    component: () => import("@/features/dailybest/pages/RankPeriodsPage.vue"),
    meta: { title: "周期排行" }
  },
  {
    path: "/rankperioditemlist",
    name: "dailybest-rank-period-item-list",
    component: () => import("@/features/dailybest/pages/RankPeriodItemsPage.vue"),
    meta: { title: "周期明细" }
  },
  {
    path: "/rankperiodstats",
    name: "dailybest-rank-period-stats",
    component: () => import("@/features/dailybest/pages/RankPeriodStatsPage.vue"),
    meta: { title: "周期统计" }
  },
  {
    path: "/dailybestfetchrunlist",
    name: "dailybest-fetch-run-list",
    component: () => import("@/features/dailybest/pages/FetchRunsPage.vue"),
    meta: { title: "抓取任务" }
  },
  {
    path: "/dailybestfetchrundetail/:id",
    name: "dailybest-fetch-run-detail",
    component: () => import("@/features/dailybest/pages/FetchRunDetailPage.vue"),
    meta: { title: "任务详情" }
  },
  {
    path: "/dailybestfetch",
    name: "dailybest-fetch",
    component: () => import("@/features/dailybest/pages/FetchPage.vue"),
    meta: { title: "手动抓取" }
  }
];
