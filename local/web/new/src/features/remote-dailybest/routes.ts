import type { RouteRecordRaw } from "vue-router";

export const remoteDailyBestRoutes: RouteRecordRaw[] = [
  {
    path: "/remotedailybestsourceconfig",
    name: "remote-dailybest-source-config",
    component: () => import("@/features/remote-dailybest/pages/SourceConfigPage.vue"),
    meta: { title: "远端站点配置", apiPath: "/api/v1/dailybest/source-config" }
  },
  {
    path: "/remotedailybestrankdaystats",
    name: "remote-dailybest-rank-day-stats",
    component: () => import("@/features/remote-dailybest/pages/RankDayStatsPage.vue"),
    meta: { title: "远端每日统计", apiPath: "/api/v1/dailybest/rank-day-stats" }
  },
  {
    path: "/remotedailybestfetchrunlist",
    name: "remote-dailybest-fetch-run-list",
    component: () => import("@/features/remote-dailybest/pages/FetchRunsPage.vue"),
    meta: { title: "远端抓取任务", apiPath: "/api/v1/dailybest/fetch-runs" }
  },
  {
    path: "/remotedailybestfetchrundetail/:id",
    name: "remote-dailybest-fetch-run-detail",
    component: () => import("@/features/remote-dailybest/pages/FetchRunDetailPage.vue"),
    meta: { title: "远端任务详情", apiPath: "/api/v1/dailybest/fetch-runs/:id" }
  },
  {
    path: "/remotedailybestfetchloglist",
    name: "remote-dailybest-fetch-log-list",
    component: () => import("@/features/remote-dailybest/pages/FetchLogsPage.vue"),
    meta: { title: "远端请求日志", apiPath: "/api/v1/dailybest/fetch-logs" }
  },
  {
    path: "/remotedailybestfetch",
    name: "remote-dailybest-fetch",
    component: () => import("@/features/remote-dailybest/pages/FetchPage.vue"),
    meta: { title: "远端手动抓取", apiPath: "/api/v1/dailybest/fetch" }
  },
  {
    path: "/remotedailybesttargetprobelist",
    name: "remote-dailybest-target-probe-list",
    component: () => import("@/features/remote-dailybest/pages/TargetProbesPage.vue"),
    meta: { title: "远端目标站探测", apiPath: "/api/v1/dailybest/target-probes" }
  }
];
