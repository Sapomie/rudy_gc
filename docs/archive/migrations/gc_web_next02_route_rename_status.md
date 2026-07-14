# gc-web-next02 前端路由改名状态

更新时间：2026-07-14

## 记录范围

- 当前前端：`v3/gc-web-next02/`
- 路由定义：`v3/gc-web-next02/src/router/index.ts`
- 本文只记录前端公开路由的命名状态。
- 后端 `/api` 路径继续沿用现有接口地址，通过前端路由的 `meta.apiPath` 接线。

## 命名规则

1. 普通页面使用一级路由。
2. 详情页面最多使用 `/:id` 二级结构。
3. 同一业务组使用统一前缀。
4. 避免 `/triggers/...`、`/crawler/...` 以及多个连字符组成的长路径。
5. 列表页面和详情页面使用不同的路由名称。
6. 路由改名后必须删除旧路由，严禁保留重定向、别名或兼容入口。

## 已完成：ListCard 路由组

| 页面功能 | 旧前端路由 | 当前正式路由 | 后端 API | 状态 |
|---|---|---|---|---|
| 全部卡片 | `/cards` | `/listcard` | `/api/cards` | 已完成 |
| 今日卡片 | `/cardstoday` | `/listcardtoday` | `/api/cardstoday` | 已完成 |
| 有排行卡片 | `/cardshasrank` | `/listcardrank` | `/api/cardshasrank` | 已完成 |
| 已有媒体卡片 | `/cardsmediamowned` | `/listcardowned` | `/api/cardsmediamowned` | 已完成 |
| 待下载卡片 | `/cardsneeddownload` | `/listcardneed` | `/api/cardsneeddownload` | 已完成 |
| 随机卡片 | `/cardsrandom` | `/listcardrandom` | `/api/cardsrandom` | 已完成 |
| 卡片聚合 | `/cardsagg` | `/listcardagg` | `/api/cardsagg` | 已完成 |
| 日排行 | `/moviecarddayrank` | `/listcarddayrank` | `/api/moviecarddayrank` | 已完成 |
| 周期排行 | `/moviecardperiodrank` | `/listcardperiodrank` | `/api/moviecardperiodrank` | 已完成 |

已同步导航、筛选表单、聚合跳转、影片标签、SC 标签、排行翻页、路由判断、样式兼容类和后端生成的前端链接。

## 已完成：SH 路由组

| 页面功能 | 旧前端路由 | 当前正式路由 | 后端 API | 状态 |
|---|---|---|---|---|
| SH 一体抓取 | `/sh` | `/shfetchall` | `/api/triggers/fetch-sehuatang-sh` | 已完成 |
| SH 列表抓取并建池 | `/shlist` | `/shfetch` | `/api/triggers/fetch-sehuatang-sh-thread-list` | 已完成 |
| SH 筛选并补全详情 | `/shdetail` | `/shfill` | `/api/triggers/fetch-sehuatang-sh-thread-detail` | 已完成 |
| SH 帖子池 | `/fetch-site-sehuatang-sh-list` | `/shpool` | `/api/fetch-site-sehuatang-sh-list` | 已完成 |
| SH 帖子详情 | `/fetch-site-sehuatang-sh-detail/:threadTid` | `/shthread/:threadTid` | `/api/fetch-site-sehuatang-sh-detail/:threadTid` | 已完成 |

所有 ListCard 和 SH 旧前端路由均已从路由表删除。

## 已完成：首页

| 默认路由 | 目标路由 | 状态 |
|---|---|---|
| `/` | `/listcardtoday` | 已完成 |

## 已完成：Movie 与 Cast

| 当前正式路由 | 候选新路由 | 状态 | 备注 |
|---|---|---|---|
| `/movie/:movieName` | `/moviedetail/:movieName` | 已完成 | 影片详情页 |
| `/cast` | 删除 | 已完成 | `name` 查询入口已迁移至 `/listcard?cn=演员名称` |
| `/cast/:personId` | `/castdetail/:personId` | 已完成 | 演员详情页 |

## 已完成：MovieAgg

| 当前正式路由 | 候选新路由 | 状态 |
|---|---|---|
| `/triggers/agg` | `/aggbackfill` | 已完成 |
| `/movie-agg-all/release` | `/aggrelease` | 已完成 |
| `/movie-agg-all/release-bucket-list` | `/bucketsrelease` | 已完成 |
| `/w-media-agg/bucket-list` | `/bucketsbirth` | 已完成 |
| `/w-media-agg/birth` | `/aggbirth` | 已完成 |
| `/wdir` | `/mediadir` | 已完成 |

## 已完成：SC

| 当前正式路由 | 候选新路由 | 状态 |
|---|---|---|
| `/sc-events` | `/scevents` | 已完成 |
| `/sc-events-cards` | `/sccards` | 已完成 |
| `/sc-events/:eventName` | `/scevent/:eventName` | 已完成 |
| `/sc-pick-smart-media` | `/scpick` | 已完成 |
| `/triggers/sc` | `/scadd` | 已完成 |
| `/triggers/sc-media` | `/scmove` | 已完成 |

## 已完成：SH 新页面

| 当前正式路由 | 候选新路由 | 状态 | 备注 |
|---|---|---|---|
| `/bucketssh` | `/shbucketlist` | 已完成 | SH 聚合桶列表 |
| `/shmagnets` | `/shmagnetlist` | 已完成 | SH 磁力列表 |

## 已完成：FC2

| 当前正式路由 | 候选新路由 | 状态 | 备注 |
|---|---|---|---|
| `/fclist` | `/fclist`（保留） | 已完成 | FC2 影片列表 |
| `/fc2/:fc2Key` | `/fcdetail/:fc2Key` | 已完成 | FC2 影片详情 |
| `/triggers/fc2-media` | `/fcingest` | 已完成 | FC2 媒体入库 |

## 已完成：抓取任务

| 当前正式路由 | 候选新路由 | 状态 |
|---|---|---|
| `/triggers/dailybest` | `/dailybestfetch` | 已完成 |
| `/triggers/seeds` | `/seedfetch` | 已完成 |
| `/triggers/post-process` | `/postprocess` | 已完成 |
| `/triggers/backfill` | `/databackfill` | 已完成 |
| `/crawl-records` | `/crawlrecordlist` | 已完成 |
| `/records` | 删除 | 已完成 |
| `/crawler/tasks` | `/crawltasklist` | 已完成 |
| `/crawler/detail-loop` | `/detailfetch` | 已完成 |

## 已完成：FetchSite

| 当前正式路由 | 候选新路由 | 状态 |
|---|---|---|
| `/triggers/fetch-site` | `/sitefetch` | 已完成 |
| `/triggers/fetch-site-javbus-filtered` | `/javbusfetch` | 已完成 |
| `/triggers/fetch-site-sukebei-filtered` | `/sukebeifetch` | 已完成 |
| `/fetch-sites` | `/siteconfiglist` | 已完成 |
| `/fetch-site-javbus-list` | `/javbuslist` | 已完成 |
| `/fetch-site-sukebei-list` | `/sukebeilist` | 已完成 |

## 已完成：Media

| 当前正式路由 | 候选新路由 | 状态 |
|---|---|---|
| `/triggers/media` | `/mediaingest` | 已完成 |
| `/triggers/media-rescan` | `/mediarescan` | 已完成 |
| `/triggers/media-rollback` | `/mediarollback` | 已完成 |
| `/medias` | `/medialist` | 已完成 |

## 已完成：数据列表

| 当前正式路由 | 候选新路由 | 状态 |
|---|---|---|
| `/w-agg-events` | `/aggeventlist` | 已完成 |
| `/torrent-albums` | `/torrentalbumlist` | 已完成 |
| `/movie-albums` | `/moviealbumlist` | 已完成 |
| `/d-seeds` | `/seedlist` | 已完成 |
| `/casts` | `/castlist` | 已完成 |
| `/e-items` | `/eitemlist` | 已完成 |

## 每组路由实施检查项

- 更新 `src/router/index.ts` 的正式路由并彻底删除旧地址。
- 使用 `meta.apiPath` 保持后端 API 地址不变。
- 更新侧边栏与顶部导航。
- 更新页面组件和共享组件中的内部链接。
- 更新筛选表单、分页、排序和清空地址。
- 更新基于 `route.path` 的页面功能判断。
- 更新或兼容路由样式类。
- 更新后端响应中仅供前端跳转使用的 URL。
- 扫描旧地址，确认前端路由表、导航和页面链接中没有残留。
- 执行 next02 构建、相关 Go 测试和 `git diff --check`。
