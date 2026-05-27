# v2 页面重写完成度记录

更新时间：2026-05-28

## 硬规则

- 没有对比，没到90%不准验收过关！！！
- 截图保存不等于截图对比。
- API 200 不等于后端验收完成。
- `make verify` 通过不等于页面验收完成。
- 必须完成 legacy/v2 截图对比、后端输出对比、SQL/性能对比。
- 未实际执行的验收结论必须删除或改成未核验。

## 当前总状态

- 本批页面总数：21 个。
- 验收通过：3 个。
- 有无效旧实现记录：21 个。
- 待删除重写：18 个。
- 真实完成度：3/21。

## 本轮执行记录

- 执行时间：2026-05-27。
- 执行范围：`/casts`、`/cast/:id`、`/records`、`/crawl-records`、`/sc-events`、`/medias`、`/w-agg-events`、`/e-items`、`/torrent-albums`、`/movie-albums`、`/d-seeds`、`/fetch-site-javbus-list`、`/fetch-site-sukebei-list`、`/fetch-site-sehuatang-list`、`/wdir`、`/w-media-agg/birth`、`/w-media-agg/bucket-list`、`/movie-agg-all/release`、`/movie-agg-all/release-bucket-list`、`/moviecardperiodrank`、`/moviecarddayrank`。
- 当前状态：in_progress。
- 当前动作：先撤回旧无证据完成口径，逐页核对 legacy 后端链路、v2 后端链路、前端页面和截图对比证据。
- 验收口径：未完成 legacy/v2 截图对比、后端输出对比、SQL/性能对比前，本批页面保持未通过。

## 完成度表

| 页面 | 当前状态 | 旧实现处理 | 后端对比 | SQL性能对比 | 前端重写 | legacy截图 | v2截图 | 截图对比 | 相似度 | 验收状态 | 备注 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `/casts` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/cast/:id` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/records` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/crawl-records` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/sc-events` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/medias` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/w-agg-events` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/e-items` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/torrent-albums` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/movie-albums` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/d-seeds` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/fetch-site-javbus-list` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/fetch-site-sukebei-list` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/fetch-site-sehuatang-list` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/wdir` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/w-media-agg/birth` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/w-media-agg/bucket-list` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端对比、SQL/性能对比证据 |
| `/movie-agg-all/release` | in_progress | 已撤回并修正 | 已执行 | 已执行 | 已修正 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-agg-all-release/legacy-owned-root.png`、`legacy-owned-year-top-director.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-agg-all-release/v2-owned-root.png`、`v2-owned-year-top-director.png` | 已执行 | 达到 90% | 本轮行为通过 | 同页下钻 href 已修回 `/movie-agg-all/release?agg_mode=owned&year=2026`；year/quarter/month/day scope、Top tab 和下钻电影卡片已复查 |
| `/movie-agg-all/release-bucket-list` | in_progress | 已撤回 | 已执行 | 已执行 | 已修正 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-release-bucket-list/legacy-focus.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-release-bucket-list/v2-final.png` | 已执行 | 达到 90% | 本轮通过 | 可见列收敛为时间桶、总量、已拥有、总大小、最新上映时间；默认 40 条；SQL 命中 `idx_mrbs_level_sort`，无 filesort |
| `/moviecardperiodrank` | in_progress | 已撤回 | 已执行 | 已执行 | 已修正 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/moviecardperiodrank/legacy-focus.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/moviecardperiodrank/v2-final.png` | 已执行 | 达到 90% | 本轮通过 | 5 列卡片、周期筛选、分页、图片和 rank 元数据已复查；SQL 命中 `idx_period_range` 与 `idx_period_rank` |
| `/moviecarddayrank` | invalid | 待删除 | 未执行 | 未执行 | 待重写 | 待重新截 | 待重新截 | 未执行 | 未核验 | 不通过 | 之前记录无完整截图对比、后端输出对比证据不足 |

## 下一步重写顺序

1. 删除或撤回本批页面旧的无证据完成结论。
2. 逐页读取 legacy 路由、handler、service、model/SQL、页面输出。
3. 逐页读取 v2 当前路由、API、service、repo/model、SQL、前端页面。
4. 对比 method、path、参数、字段、数量、排序、筛选、分页、链接语义。
5. 对比 SQL 次数、执行计划、索引命中和耗时。
6. 后端不达标先修后端。
7. 前端按真实后端输出重写。
8. legacy `5010` 与 v2 `2040` 同路由同参数截图。
9. 逐页截图对比，记录差异。
10. 不到 90% 继续改，重新截图复查。
11. 按真实证据回写本完成度文档。

## 本轮两页验收记录

执行时间：2026-05-28 00:19。

### `/moviecardperiodrank`

- legacy 截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/moviecardperiodrank/legacy-focus.png`
- v2 复查截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/moviecardperiodrank/v2-final.png`
- 后端对比：legacy 与 v2 均为周期榜月榜默认口径；v2 API `GET /api/gc/v2/rank/period` 返回 200，`period_key=2026-05`，`category_label=MonthSource`，`total=300`，`page_size=18`。
- 前端对比：v2 已恢复 5 列卡片密度；保留 Week/Month/Quarter/Year、MonthSource/AllTimeSource、上一期、下一期、随机、最新、分页、图片、电影链接和 rank 元数据。
- SQL/性能：period 查询命中 `idx_period_range`，`Extra=Using where; Backward index scan`；item 查询命中 `idx_period_rank`，无 filesort；v2 API 实测约 43.36ms。
- 可见差异：v2 使用全新侧栏与顶部快捷入口；rank 元数据以紧凑条显示，和 legacy 卡片内文字块存在样式差异。
- 修复轮次：2。
- 相似度判断：达到 90%。
- 端口清理：已执行 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 stop-all`，`2040` 与 `2041` 无监听输出。

### `/movie-agg-all/release-bucket-list`

- legacy 截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-release-bucket-list/legacy-focus.png`
- v2 复查截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-release-bucket-list/v2-final.png`
- 后端对比：legacy 与 v2 默认均为上映时间月桶全量口径；v2 API `GET /api/gc/v2/pages/movie-release-bucket-list` 返回 200，`total=301`，`page_size=40`，默认倒序。
- 前端对比：v2 可见列已收敛为时间桶、All、Owned、总大小、最新上映时间；筛选包含层级、Mode、Scope、Year、Quarter、Month、Day；分页显示 1-40 / 301。
- SQL/性能：默认 count 命中 `idx_mrbs_level_sort`；默认 list 命中 `idx_mrbs_level_sort`，`Extra=NULL`，无 filesort；v2 API 实测约 2.45ms。
- 可见差异：v2 使用 select 控件承载层级与模式，legacy 使用按钮组；v2 全页截图不再重复 sticky 顶栏和侧栏。
- 修复轮次：2。
- 相似度判断：达到 90%。
- 端口清理：已执行 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 stop-all`，`2040` 与 `2041` 无监听输出。

## Aggregation 单页验收记录

执行时间：2026-05-28 01:13。

### `/movie-agg-all/release`

- 旧结论撤回：此前记录的“达到 90% / 本轮通过”无完整点击下钻和 Top tab 证据，已先改为 `invalid` 后重新修复。
- legacy 截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-agg-all-release/legacy-owned-root.png`、`legacy-owned-year.png`、`legacy-owned-year-top-director.png`、`legacy-owned-quarter.png`、`legacy-owned-month.png`、`legacy-owned-day.png`。
- v2 复查截图：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/page-rebuild-20260527/movie-agg-all-release/v2-owned-root.png`、`v2-owned-year.png`、`v2-owned-year-top-director.png`、`v2-owned-quarter.png`、`v2-owned-month.png`、`v2-owned-day.png`。
- 后端对比：legacy handler 读取 `agg_mode/year/quarter/month/day` 并走 `BuildReleaseView`；v2 已改为专用加载分支，root/year/quarter/month/day 分别输出 legacy 同语义 bucket、Top scope、`range_start/range_end`。API 实测 root owned `total=135`，year owned `total=128`，quarter owned `total=123`，month owned `total=131`，day owned `total=0`。
- 链接对比：legacy owned 根页首个年份 href 为 `/movie-agg-all/release?agg_mode=owned&year=2026`；v2 owned 根页首个年份 href 同为 `/movie-agg-all/release?agg_mode=owned&year=2026`。month/day bucket href 已按 legacy 去掉 `quarter` 参数。
- 前端对比：v2 年页恢复 `季度 / 月份` tab；Top 聚合默认折叠，展开后恢复 `演员 / 导演 / Label / Prefix` tab，四个 tab 已逐个点击；year/quarter/month/day 下钻均显示对应电影卡片或空态。
- SQL/性能：bucket 查询命中 `idx_mrbs_level_sort`；Top 查询命中 `uk_mrts_mode_scope_type_key` 或 `idx_mrts_scope_type_rank` 候选索引，小结果集排序与 legacy 同口径。API 热身后 root owned 约 2.42ms，year owned 约 4.35ms，year 电影卡片约 18.96ms。
- 可见差异：v2 保留全新 Vue 侧栏、顶部快捷入口、按钮和卡片样式；核心时间桶、Top tab、下钻参数、电影卡片行为与 legacy 对齐。
- 修复轮次：3。
- 相似度判断：达到 90%。
- 验证命令：`cd v2/gc-api && go test ./...` 通过；`cd v2/gc-web && npm run test` 通过；`cd v2/gc-web && npm run build` 通过。
- 端口清理：已执行 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 stop-all`；`lsof -iTCP:2040 -sTCP:LISTEN -n -P` 与 `lsof -iTCP:2041 -sTCP:LISTEN -n -P` 无监听输出。
