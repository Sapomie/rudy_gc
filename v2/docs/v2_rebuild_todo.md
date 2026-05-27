# rudy_gc v2 重建 TODO

更新时间：2026-05-27

## 当前维护口径

- 后续只更新 `v2/` 功能。
- 历史系统仅作为只读验收基准，用于核对功能结果、字段完整性、交互语义、API 输出和截图接近度。
- v2 的实现必须来自 `v2/gc-api`、`v2/gc-web`、`v2/docs`、`v2/deploy` 自身。
- 禁止使用 `z_reference`、历史模板、历史 CSS、历史 JS、历史 class、历史静态资源、历史页面壳和历史交互代码。

## 最高执行口径

- v2 全量验收合格前，状态只能使用 `in_progress`、`blocked`、`invalid`、`todo`。
- 阶段性构建通过、局部测试通过、单页可打开、API 返回 200、路由存在、文档写完，都不能作为完成理由。
- 每个页面必须记录历史基准截图路径、v2 截图路径、不足点清单、修复轮次、复查截图、后端输出对比、SQL/性能对比和真实相似度判断。

## 文档先行硬规则

- 任何 v2 实现任务必须先更新本文档和 `/Users/gaojinwei/src/my/rudy_gc/v2/docs/gc_frontend_backend_plan.md`。
- 涉及运行、端口、容器、部署、静态资源、图片代理时，必须同步更新 `/Users/gaojinwei/src/my/rudy_gc/v2/deploy/README.md`。
- 每轮代码前，文档记录目标、范围、状态、风险、验收口径和执行清单。
- 每轮代码后，文档记录实际完成、失败原因、剩余缺口、验证结果和端口清理状态。

## 本轮 TODO：21 个页面验收闭环

状态：`in_progress`

本轮页面：

| 页面 | 当前任务 |
| --- | --- |
| `/casts` | 回写后端、SQL、截图差异和相似度证据 |
| `/cast/:id` | 回写演员信息与 cards 影片卡片对比证据 |
| `/records` | 回写 legacy 跳转与 v2 重定向证据 |
| `/crawl-records` | 回写列表字段、筛选、排序、分页证据 |
| `/sc-events` | 回写事件列表和电影链接语义证据 |
| `/medias` | 回写媒体列表筛选、排序、SQL 性能证据 |
| `/w-agg-events` | 回写聚合事件列表证据 |
| `/e-items` | 回写抓取 item 列表和性能证据 |
| `/torrent-albums` | 回写种子相册列表证据 |
| `/movie-albums` | 回写电影相册列表证据 |
| `/d-seeds` | 回写下载种子列表证据 |
| `/fetch-site-javbus-list` | 回写默认链路性能修正证据 |
| `/fetch-site-sukebei-list` | 回写默认链路性能修正证据 |
| `/fetch-site-sehuatang-list` | 回写外链列和列表对比证据 |
| `/wdir` | 回写目录树默认口径和差异证据 |
| `/w-media-agg/birth` | 回写下载时间聚合总览差异证据 |
| `/w-media-agg/bucket-list` | 回写下载时间桶列表证据 |
| `/movie-agg-all/release` | 本轮已按 legacy 同页下钻、Top tab 和电影卡片重新修复，端口已清理 |
| `/movie-agg-all/release-bucket-list` | 本轮证据已回写，继续随全量验收复查 |
| `/moviecardperiodrank` | 本轮证据已回写，继续随全量验收复查 |
| `/moviecarddayrank` | 按 rank 专页回写截图、API、分页、图片和 SQL 证据 |

## 本轮已取证

- 复查截图目录：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/`
- 已存在 21 组 legacy/v2 复查截图。
- 已存在 3 张复查汇总图：`compare-review-part-1.jpg`、`compare-review-part-2.jpg`、`compare-review-part-3.jpg`。
- `/moviecardperiodrank` 最终复查截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/moviecardperiodrank/v2-final.png`。
- `/movie-agg-all/release-bucket-list` 最终复查截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-release-bucket-list/v2-final.png`。
- `/movie-agg-all/release` 本轮复查截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-agg-all-release/v2-owned-root.png`、`v2-owned-year-top-director.png`、`v2-owned-quarter.png`、`v2-owned-month.png`、`v2-owned-day.png`。

## 本轮待执行

1. `/movie-agg-all/release` 旧完成结论已撤回并重新修复：年份 bucket 跳转恢复同页 `/movie-agg-all/release?agg_mode=owned&year=2026`，year/quarter/month/day scope、Top tab、下钻电影卡片已复查。
2. Aggregation 剩余页面待继续：`/w-agg-events`、`/wdir`、`/w-media-agg/birth`、`/w-media-agg/bucket-list`。
3. `/movie-agg-all/release-bucket-list` 做回归复核。
4. 复跑剩余 Aggregation 页面状态码、耗时和 SQL 执行计划。
5. 回写 `/Users/gaojinwei/src/my/rudy_gc/v2/docs/v2_page_rebuild_completion.md`。
6. 未达 90% 的页面继续修正、重启服务、重新截图、重新记录。
7. 验证结束后关闭本轮临时 v2 服务并记录端口清理。

## 当前风险

- 页面截图已经保存，逐页差异记录仍需补齐。
- `/movie-agg-all/release` 后端输出、SQL 和性能证据已按同页下钻、Top tab、年/季/月/日 scope 重新补齐；其余 Aggregation 页面仍需补齐。
- 端口清理状态：已执行 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 stop-all`；`2040` 与 `2041` 无监听输出。
