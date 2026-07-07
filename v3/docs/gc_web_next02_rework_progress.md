# gc-web-next02 截图问题页面重做进度

## 本次重做结论
- 本次重做范围来自 2026-07-08 用户截图指出的问题页面。
- 上一轮把部分低相似度、按钮未实点、交互未接通页面错误记录为“已完成”，本次全部回收重做。
- 本文档只记录本次 next02 重做进度和规则。
- 当前写入范围固定为 `v3/gc-web-next02/`、`v3/tmp/gc-web-next02-page-match/` 和本文档。
- 当前版本固定为 `next02`；处理 next02 期间不修改 next03、next04、next05。

## 本次失败原因记录
- 截图对比没有形成“发现差异 -> 修改页面 -> 再截图验证”的闭环。
- 多个页面只完成 DOM 主体结构和快验截图，没有逐页肉眼修正到旧版布局。
- 多个按钮只是视觉占位，缺少旧版对应点击行为、API 调用、弹窗、任务触发、刷新和状态更新。
- 进度文档把 85%-94% 的页面误标为完成，导致完成状态失真。

## 新验收规则
- 页面整体视图框架相似度必须 `>= 90%`。
- 所有可见按钮必须逐个实点验证。
- 顶部 tab、侧边栏入口、主操作按钮、筛选按钮、排序按钮、分页、表格行操作、弹窗按钮、复制按钮、任务按钮都必须验证。
- 表单提交、筛选、排序、分页、弹窗、复制、任务触发、状态刷新、结果表更新必须和 `v3/gc-web` 旧版行为对齐。
- 只截图、只 DOM 通过、只 HTTP 200、只页面有内容，均不能记录为完成。
- 截图相似度达到 90% 但按钮未验证时，页面仍然不能完成。
- 按钮全部可用但整体框架相似度低于 90% 时，页面仍然不能完成。
- 每个页面完成时必须记录：旧版对照文件、next02 修改文件、截图相似度、按钮验证结论、剩余问题。
- 页面完成后才能在队列中标记 `[已完成]`。

## 旧版对照文件
- MovieAgg / WDir 页面：`v3/gc-web/src/movie-agg.js`
- SC 页面：`v3/gc-web/src/sc.js`
- 抓取任务页面：`v3/gc-web/src/crawler-jobs.js`、`v3/gc-web/src/crawler-tasks.js`、`v3/gc-web/src/crawler-detail-loop.js`
- FetchSite 页面：`v3/gc-web/src/fetch-sites.js`、`v3/gc-web/src/fetch-site-list.js`
- 色花堂 SH 页面：`v3/gc-web/src/crawler-jobs.js`、`v3/gc-web/src/fetch-site-list.js`、`v3/gc-web/src/sh-sehuatang-detail.js`
- Media 页面：`v3/gc-web/src/crawler-jobs.js`
- 侧边栏入口：`v3/gc-web/src/sidebar.js`
- 顶部导航入口：`v3/gc-web/src/top-nav.js`

## next02 主要修改文件
- `v3/gc-web-next02/src/pages/MovieAggPage.vue`
- `v3/gc-web-next02/src/pages/ReleaseBucketListPage.vue`
- `v3/gc-web-next02/src/pages/WDirPage.vue`
- `v3/gc-web-next02/src/pages/AggTriggersPage.vue`
- `v3/gc-web-next02/src/pages/ScEventsPage.vue`
- `v3/gc-web-next02/src/pages/ScPickSmartPage.vue`
- `v3/gc-web-next02/src/pages/ScTriggersPage.vue`
- `v3/gc-web-next02/src/pages/ScMediaMovePage.vue`
- `v3/gc-web-next02/src/pages/TriggerPage.vue`
- `v3/gc-web-next02/src/pages/CrawlerDetailLoopPage.vue`
- `v3/gc-web-next02/src/pages/CrawlerTasksPage.vue`
- `v3/gc-web-next02/src/pages/FetchSitesPage.vue`
- `v3/gc-web-next02/src/pages/ListPage.vue`
- `v3/gc-web-next02/src/pages/MediaIngestPage.vue`
- `v3/gc-web-next02/src/pages/MediaRescanPage.vue`
- `v3/gc-web-next02/src/pages/MediaRollbackPage.vue`
- `v3/gc-web-next02/src/styles/index.css`
- `v3/gc-web-next02/src/router/index.ts`
- `v3/tmp/gc-web-next02-page-match/capture-page-round.mjs`

## 本次重做页面队列
- [待重做] next02 /movie-agg-all/release
- [待重做] next02 /movie-agg-all/release-bucket-list
- [待重做] next02 /w-media-agg/bucket-list
- [待重做] next02 /w-media-agg/birth
- [待重做] next02 /triggers/agg
- [待重做] next02 /wdir
- [待重做] next02 /sc-events
- [待重做] next02 /sc-pick-smart-media
- [待重做] next02 /triggers/sc
- [待重做] next02 /triggers/sc-media
- [待重做] next02 /triggers/seeds
- [待重做] next02 /triggers/post-process
- [待重做] next02 /triggers/backfill
- [待重做] next02 /crawler/detail-loop
- [待重做] next02 /crawler/tasks
- [待重做] next02 /triggers/fetch-site
- [待重做] next02 /triggers/fetch-sehuatang
- [待重做] next02 /fetch-sites
- [待重做] next02 /fetch-site-javbus-list
- [待重做] next02 /fetch-site-sukebei-list
- [待重做] next02 /fetch-site-sehuatang-list
- [待重做] next02 /triggers/fetch-sehuatang-sh
- [待重做] next02 /triggers/fetch-sehuatang-sh-thread-list
- [待重做] next02 /triggers/fetch-sehuatang-sh-thread-detail
- [待重做] next02 /fetch-site-sehuatang-sh-list
- [待重做] next02 /triggers/media
- [待重做] next02 /triggers/media-rescan
- [待重做] next02 /triggers/media-rollback

## 当前总进度
- 本次重做总页数：28
- 已完成：0
- 待重做：28
- 当前页面：未开始
- 当前状态：规则文档已创建，等待按队列逐页重做代码、按钮和截图验证。

## 单页记录模板
```md
### next02 /path
- 状态：
- 旧版对照文件：
- next02 修改文件：
- 布局截图相似度：
- hover / 交互截图相似度：
- 按钮实点结果：
- 已修复问题：
- 剩余问题：
- 结论：
```
