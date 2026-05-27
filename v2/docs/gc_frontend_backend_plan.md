# rudy_gc v2 前后端重建计划

更新时间：2026-05-27

## 执行顺序

1. 先更新文档和进度。
2. 再执行代码、脚本、测试、构建或页面验证。
3. 执行后再次更新文档，记录真实结果。

本顺序适用于所有 v2 任务，包括前端、后端、SQL、缓存、部署、图片链路、测试和排查。

## 当前维护口径

- 后续只更新 `v2/` 功能。
- 历史系统仅作为只读验收基准：功能结果、字段完整性、交互语义、API 输出、截图接近度。
- v2 实现必须属于 `v2/gc-api` 与 `v2/gc-web` 自己的代码体系。
- 禁止使用 `z_reference`、历史模板、历史 CSS、历史 JS、历史 class、历史静态资源、历史页面壳和历史交互代码。

## 本轮范围

状态：`in_progress`

本轮只处理 21 个页面：

- `/casts`
- `/cast/:id`
- `/records`
- `/crawl-records`
- `/sc-events`
- `/medias`
- `/w-agg-events`
- `/e-items`
- `/torrent-albums`
- `/movie-albums`
- `/d-seeds`
- `/fetch-site-javbus-list`
- `/fetch-site-sukebei-list`
- `/fetch-site-sehuatang-list`
- `/wdir`
- `/w-media-agg/birth`
- `/w-media-agg/bucket-list`
- `/movie-agg-all/release`
- `/movie-agg-all/release-bucket-list`
- `/moviecardperiodrank`
- `/moviecarddayrank`

## 修改范围

- `/Users/gaojinwei/src/my/rudy_gc/v2/docs/v2_page_rebuild_completion.md`
- `/Users/gaojinwei/src/my/rudy_gc/v2/docs/gc_frontend_backend_plan.md`
- `/Users/gaojinwei/src/my/rudy_gc/v2/docs/v2_rebuild_todo.md`
- `/Users/gaojinwei/src/my/rudy_gc/v2/deploy/README.md`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/model/modelx/page_registry.go`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/model/modelx/page_repo.go`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/router/handler/page/handler.go`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/types/page.go`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/scripts/verify-api-contracts.mjs`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/components/DataTable.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/pages/generic/index.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/router/index.ts`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/styles/index.css`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/types/api.ts`

## 主要改动点

- 撤回 21 个页面旧的无证据完成口径，完成度文档只写真实证据。
- 通用列表后端补齐页面专属字段、默认筛选、排序别名、筛选规则、隐藏列和链接语义。
- `/records` 按 legacy 行为从前端重定向到 `/crawl-records`。
- `/cast/:id` 在演员基础信息外追加真实 cards 接口影片卡片。
- `/w-media-agg/birth`、`/movie-agg-all/release` 改为聚合总览形态，展示年份桶和 root Top 统计。
- DataTable 隐藏内部字段，仅渲染可见列，并格式化时间、大小和时长。
- 前端通用页补齐筛选、排序、分页、错误态、空态、聚合总览、外链和内部链接。
- 截图复查时取消 sticky 侧栏、顶栏、表头和卡片工具条，避免长截图重复拼接造成误判。
- `/fetch-site-javbus-list`、`/fetch-site-sukebei-list` 默认链路避免为了媒体出生时间全量宽 JOIN，媒体出生时间仅在列、排序、筛选需要时通过相关子查询读取。

## 验收口径

- legacy 固定为 `http://localhost:5010/...`。
- v2 页面固定为 `http://localhost:2040/...`。
- v2 API 固定为 `http://localhost:2041/api/gc/v2/...`。
- 每页必须记录 legacy 截图、v2 截图、截图差异、后端输出对比、SQL/性能对比、修复轮次、复查截图、真实相似度判断和端口清理结果。
- 截图保存只作为取证，不等于截图对比。
- API 200、页面能打开、构建通过、测试通过都不能替代页面验收。
- 未完成逐页截图对比、后端输出对比和 SQL/性能对比前，本批页面保持 `in_progress`。

## 当前取证

- 原始截图目录：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/`
- 复查截图目录：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/`
- 当前已存在 21 组 `legacy-review.png` 与 `v2-review.png`。
- 当前已存在 3 张复查汇总图：`compare-review-part-1.jpg`、`compare-review-part-2.jpg`、`compare-review-part-3.jpg`。
- `/moviecardperiodrank` 本轮最终 v2 截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/moviecardperiodrank/v2-final.png`。
- `/movie-agg-all/release-bucket-list` 本轮最终 v2 截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-release-bucket-list/v2-final.png`。
- `/movie-agg-all/release` 本轮最终 v2 截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-agg-all-release/v2-owned-root.png`、`v2-owned-year-top-director.png`。

## 当前风险

- `v2_page_rebuild_completion.md` 仍需回写每页真实证据。
- 21 页仍需逐页记录可见差异和真实相似度判断。
- 后端链路仍需回写 method、path、参数、字段、排序、筛选、分页、链接语义、SQL 和性能结果。
- `/wdir`、`/w-media-agg/birth` 和部分长表页仍需按复查图记录明显差异。
- 验证结束后必须关闭本轮临时 v2 服务，并记录 `2040`、`2041` 清理结果。

## 本轮两页结果

- `/moviecardperiodrank`：已恢复 5 列卡片密度，压缩 rank 元数据；API 200，`total=300`，SQL 命中 `idx_period_range` 与 `idx_period_rank`，无 filesort；本轮人工截图对比达到 90%。
- `/movie-agg-all/release-bucket-list`：可见列收敛为时间桶、总量、已拥有、总大小、最新上映时间；默认每页 40；API 200，`total=301`；SQL 命中 `idx_mrbs_level_sort`，无 filesort；本轮人工截图对比达到 90%。
- `/movie-agg-all/release`：旧 90% 结论已撤回并重新修复；年份 bucket href 已恢复同页 `/movie-agg-all/release?agg_mode=owned&year=2026`；year/quarter/month/day scope、Top `演员 / 导演 / Label / Prefix` tab、下钻电影卡片已截图点击复查；本轮人工截图对比达到 90%。
- 端口清理：已执行 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 stop-all`，`2040` 与 `2041` 无监听输出。

## 2026-05-28 Aggregation 本轮范围

状态：`in_progress`

本轮目标：
- 完成 Aggregation 菜单剩余页面：`/w-agg-events`、`/wdir`、`/w-media-agg/birth`、`/w-media-agg/bucket-list`、`/movie-agg-all/release`。
- `/movie-agg-all/release-bucket-list` 已在上一轮达到 90%，本轮只做回归复核。

本轮修改范围：
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/model/modelx/page_registry.go`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/model/modelx/page_repo.go`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/pages/generic/index.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/components/DataTable.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/styles/index.css`
- `/Users/gaojinwei/src/my/rudy_gc/v2/docs/v2_page_rebuild_completion.md`
- `/Users/gaojinwei/src/my/rudy_gc/v2/docs/gc_frontend_backend_plan.md`
- `/Users/gaojinwei/src/my/rudy_gc/v2/docs/v2_rebuild_todo.md`
- `/Users/gaojinwei/src/my/rudy_gc/v2/deploy/README.md`

本轮验收口径：
- 每页保存 legacy `5010` 与 v2 `2040` 同路由截图。
- 每页记录 API 输出、字段、排序、筛选、分页和链接语义。
- 每页记录 SQL 执行计划、索引命中和耗时。
- 未达到 90% 的页面继续修正并重新截图。
- 验证结束后关闭本轮临时 `2040`、`2041` 服务并记录端口清理。

## 下一步执行清单

1. 复跑后端测试、前端测试和构建。
2. 复跑 21 个页面的 legacy 页面、v2 页面、v2 API 状态码与耗时记录。
3. 按 21 页逐项回写完成度文档。
4. 对未达 90% 的页面继续修正并重新截图。
5. 验证结束后关闭临时 v2 服务，记录端口清理结果。
