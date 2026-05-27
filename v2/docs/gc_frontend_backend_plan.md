# rudy_gc v2 前后端重建计划

更新时间：2026-05-27

## 执行顺序

1. 先更新文档和进度。
2. 再执行代码、脚本、测试、构建或页面验证。
3. 执行后再次更新文档，记录真实结果。

本顺序适用于所有 v2 任务，包括前端、后端、SQL、缓存、部署、图片链路、测试和排查。

## 当前维护口径

- 后续只更新 `v2/` 功能。
- 历史系统不会继续更新，不再作为 v2 实现规则、复用来源或风格约束。
- 历史系统仅作为只读验收基准：功能结果、字段完整性、交互语义、API 输出、截图接近度。
- v2 实现必须属于 `v2/gc-api` 与 `v2/gc-web` 自己的代码体系。
- 禁止使用 `z_reference`、历史模板、历史 CSS、历史 JS、历史 class、历史静态资源、历史页面壳和历史交互代码。

## 前端重写总原则

- 当前 v2 前端全部重写。
- 使用全新的 Vue/Vite/TypeScript 前端设计。
- 页面功能必须覆盖既定功能范围，代码实现必须属于 v2 新前端。
- 不准为了快而做半套页面、假交互、假按钮、空实现。
- 每个页面写完后必须打开历史只读基准页面和 v2 页面，使用 MCP Chrome 截图观察对比。
- 页面验收必须按“截图对比 -> 记录不足点 -> 修改 -> 再截图对比”循环执行。
- 页面功能接近且视觉/交互相似度达到 90% 后，才允许进入合格候选。
- 构建通过、API 200、页面能打开、基础功能可点，都不能作为页面合格依据。

## 前端架构目标

状态：`in_progress`

当前状态：
- 用户已要求删除并重写前端。
- cards 相关 6 个页面已使用 v2 全新前端补齐并进入合格候选。
- movie detail、rank、generic、operation 等页面仍需按本文档继续清点、覆盖和验证。

目标结构：
- `src/api`：只封装 v2 API，不放页面逻辑。
- `src/types`：按 API 语义定义类型，不复用不同语义 DTO。
- `src/router`：覆盖 v2 页面路由。
- `src/layouts`：全新 v2 布局。
- `src/components`：全新组件体系。
- `src/pages`：cards、movie detail、rank、generic、notfound。
- `src/styles`：全新样式体系。

## 页面与 API 对照

| 页面 | 前端路由 | 后端 API | 状态 | 要求 |
| --- | --- | --- | --- | --- |
| Cards | `/cards` | `GET /api/gc/v2/cards` | in_progress | 筛选、排序、分页、卡片、图片全部对齐。 |
| Today | `/cardstoday` | `GET /api/gc/v2/cards?view=cardstoday` | in_progress | 今日卡片输出对齐。 |
| Has Rank | `/cardshasrank` | `GET /api/gc/v2/cards?view=cardshasrank` | in_progress | 排名相关筛选和排序对齐。 |
| Media Owned | `/cardsmediamowned` | `GET /api/gc/v2/cards?view=cardsmediamowned` | in_progress | 媒体状态、MPlay、图片对齐。 |
| Need Download | `/cardsneeddownload` | `GET /api/gc/v2/cards?view=cardsneeddownload` | in_progress | 下载状态筛选对齐。 |
| Random | `/cardsrandom` | `GET /api/gc/v2/cards?view=cardsrandom` | in_progress | 随机数量参数对齐。 |
| Movie Detail | `/movie/:movie` | `GET /api/gc/v2/movie/:movie` | in_progress | 使用 movie name 路由语义，详情字段完整。 |
| Day Rank | `/moviecarddayrank` | `GET /api/gc/v2/rank/day` | in_progress | 日期导航、随机、分页、卡片对齐。 |
| Period Rank | `/moviecardperiodrank` | `GET /api/gc/v2/rank/period` | in_progress | 周期筛选、分类、分页、周期元数据对齐。 |
| Generic Pages | 多路由 | `GET /api/gc/v2/pages/:key` | in_progress | 表格、筛选、表头排序、分页、空态对齐。 |
| Operation Pages | 多路由 | 待真实动作 API | blocked | 无真实 API 前不得渲染可用动作按钮。 |

## 后端 API 要求

状态：`in_progress`

- 现有基础 API 可以作为前端数据来源，但不能自动视为完成。
- 所有页面动作必须迁移到真实业务 API。
- 禁止恢复通用 `page-actions` 假动作入口。
- API 响应必须区分展示 ID、内部 ID、外部 ID、路由参数。
- `/movie/:movie` 页面参数使用电影番号 `a_movie.name`。
- 数据库关联和内部 API 可使用 `jav_id`，但页面链接不得用 `jav_id` 拼 `/movie/...`。
- 查询链路以 v2 自有模型、缓存和索引设计为准。

## 已完成候选记录

### Cards

状态：`in_progress`

当前结果：
- 6 个 cards 相关页面已使用 v2 新前端补齐：`/cards`、`/cardstoday`、`/cardshasrank`、`/cardsmediamowned`、`/cardsneeddownload`、`/cardsrandom`。
- 侧边栏、顶部快捷入口和页面内 view tabs 统一来自 `cardRoutes`。
- 页面调用统一通过 `GET /api/gc/v2/cards`，由当前路由 meta 写入对应 `view`。
- 已验证排序链接、顶部/底部分页、WMedia 快捷筛选、筛选面板、view 切换、电影详情链接、Bus/Suk/Lib 外链。
- MCP Chrome 已逐页打开历史只读基准与 v2，并保存截图到 `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/`。
- v2 浏览器复测：6 个页面均无破图、无 v2 控制台 warn/error；首屏数量分别为 18、18、18、18、18、6。
- cards 相关页面进入合格候选；v2 全量仍保持 `in_progress`。

后端结果：
- 已重建 cards service。
- 已接入 `MovieTypeCache`，cards 列表按 ID 优先读缓存。
- 已移除 cards 查询里的通用宽 JOIN、`COUNT(DISTINCT)`、无必要 `GROUP BY`、`MAX(...) ORDER BY` 残留实现。
- 已修正 `/cardsneeddownload` 的真实相册名为 `待下载电影`，接口 Total 恢复为 278。
- 已将 `/cardshasrank` 排序尾键对齐 `bm_minfo.idx_first_rank_name`，避免 filesort。
- 已将 `/cardsmediamowned` 改为 w_media 单源分页，并对齐 `idx_w_media_birth_name`。
- 已执行 `go test ./...`、前端 `npm run test`、前端 `npm run build` 和 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify`，均通过。

## 前端验收清单

每个页面必须记录：

- 路由可访问。
- 导航可进入。
- API 调用成功。
- 错误态可见。
- 空态可见。
- 筛选可用。
- 排序可用。
- 分页可用。
- 图片可显示。
- 内部链接语义正确。
- 外链正确。
- 截图与历史只读基准接近。
- 历史基准截图路径。
- v2 截图路径。
- 不足点清单。
- 修复轮次。
- 复查截图。
- 最终相似度判断，必须达到 90%。

List 页专项：

- 表头排序可点。
- 升降序可切换。
- 保留当前筛选参数。
- 当前列有激活态。
- 方向提示可见。

Movie Card 专项：

- `0` 分数和 `0` 观看数不能被当作空值隐藏。
- 图片路径不做静默替换。
- 筛选参数只在用户显式传入时生效。

## 当前执行记录：generic list pages

状态：`in_progress`

本轮一次性完成以下 19 个 v2 List 页面：
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

后端基础结果：
- 已重建 cards service。
- 已接入 `MovieTypeCache`，cards 列表按 ID 优先读缓存。
- 已移除 cards 查询里的通用宽 JOIN、`COUNT(DISTINCT)`、无必要 `GROUP BY`、`MAX(...) ORDER BY` 残留实现。
- 已修正 `/cardsneeddownload` 的真实相册名为 `待下载电影`，接口 Total 恢复为 278。
- 已将 `/cardshasrank` 排序尾键对齐 `bm_minfo.idx_first_rank_name`，避免 filesort。
- 已将 `/cardsmediamowned` 改为 w_media 单源分页，并对齐 `idx_w_media_birth_name`。
- 已执行 `go test ./...`、前端 `npm run test`、前端 `npm run build` 和 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify`，均通过。

设计规则：
- 这 19 个页面全部使用 v2 新前端 `generic` 列表承载页。
- 页面主体使用紧凑后台列表设计：页头、筛选工具栏、状态条、表格、分页。
- 筛选、Page Size、查询、清空合并到同一条工具栏。
- 表格为主体，分页贴近表格，不单独做大面板。
- 表头排序必须可点、可切换升降序、保留当前筛选参数、当前列有激活态和方向提示。
- 内部链接必须保持语义正确：电影链接使用 movie name，`/cast/:id` 使用 cast id，`/wdir` 使用目录 id。
- 没有真实动作 API 的按钮不渲染。
- 禁止 pending action、通用假 action、空实现、禁用按钮冒充完成。
- 禁止使用 `z_reference`、历史模板、历史 CSS、历史 JS、历史 class、历史静态资源。

计划修改文件：
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/router/index.ts`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/data/routes.ts`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/layouts/AppLayout.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/pages/generic/index.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/components/DataTable.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/styles/index.css`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/scripts/verify-routes.mjs`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/scripts/verify-api-contracts.mjs`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/scripts/verify-source.mjs`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/model/modelx/page_registry.go`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/model/modelx/page_repo.go`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/model/modelx/page_repo_test.go`

逐页验收记录字段固定为：

| 页面 | 路由可访问 | 导航可进入 | API 调用成功 | 错误态可见 | 空态可见 | 筛选可用 | 排序可用 | 分页可用 | 图片可显示 | 内部链接语义正确 | 外链正确 | 历史基准截图路径 | v2 截图路径 | 不足点清单 | 修复轮次 | 复查截图 | 最终相似度判断 | 状态 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `/casts` | 通过 | 通过 | 通过，Total 12696 | 通过，共用截图 `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-error-state.png` | 通过，共用截图 `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-empty-state.png` | 通过 | 通过 | 通过 | 无图片 | 通过，`/cast/{id}` | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-casts.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-casts.png` | 表头 ASC 箭头显示错误，已修复 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-casts.png` | 达到 90% | in_progress |
| `/cast/:id` | 通过，样例 `/cast/33` | 通过 `/casts` 进入 | 通过，Total 1 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过，id 参数 | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-cast-detail.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-cast-detail.png` | 无新增不足 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-cast-detail.png` | 达到 90% | in_progress |
| `/records` | 通过 | 通过 | 通过，Total 103 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过 | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-records.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-records.png` | 历史基准 `/records` 跳转到 `/crawl-records`，已记录 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-records.png` | 达到 90% | in_progress |
| `/crawl-records` | 通过 | 通过 | 通过，Total 103 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过 | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-crawl-records.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-crawl-records.png` | 无新增不足 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-crawl-records.png` | 达到 90% | in_progress |
| `/sc-events` | 通过 | 通过 | 通过，Total 124 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过，movie 链接使用 movie name | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-sc-events.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-sc-events.png` | 无真实 operation 链接已隐藏 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-sc-events.png` | 达到 90% | in_progress |
| `/medias` | 通过 | 通过 | 通过，Total 2771 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过，movie 链接使用 movie name | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-medias.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-medias.png` | 无真实 operation 链接已隐藏 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-medias.png` | 达到 90% | in_progress |
| `/w-agg-events` | 通过 | 通过 | 通过，Total 324 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过 | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-w-agg-events.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-w-agg-events.png` | 无新增不足 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-w-agg-events.png` | 达到 90% | in_progress |
| `/e-items` | 通过 | 通过 | 通过，Total 111385 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过，movie 链接使用 movie name | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-e-items.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-e-items.png` | 无新增不足 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-e-items.png` | 达到 90% | in_progress |
| `/torrent-albums` | 通过 | 通过 | 通过，Total 1040 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过，movie 链接使用 movie name | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-torrent-albums.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-torrent-albums.png` | 无新增不足 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-torrent-albums.png` | 达到 90% | in_progress |
| `/movie-albums` | 通过 | 通过 | 通过，Total 6939 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过，movie 链接使用 movie name | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-movie-albums.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-movie-albums.png` | 表头 ASC 箭头显示错误，已修复 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-movie-albums.png` | 达到 90% | in_progress |
| `/d-seeds` | 通过 | 通过 | 通过，Total 186 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过 | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-d-seeds.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-d-seeds.png` | 无新增不足 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-d-seeds.png` | 达到 90% | in_progress |
| `/fetch-site-javbus-list` | 通过 | 通过 | 通过，Total 111385 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过，movie 链接使用 movie name | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-fetch-site-javbus-list.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-fetch-site-javbus-list.png` | 无新增不足 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-fetch-site-javbus-list.png` | 达到 90% | in_progress |
| `/fetch-site-sukebei-list` | 通过 | 通过 | 通过，Total 111385 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过，movie 链接使用 movie name | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-fetch-site-sukebei-list.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-fetch-site-sukebei-list.png` | 无新增不足 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-fetch-site-sukebei-list.png` | 达到 90% | in_progress |
| `/fetch-site-sehuatang-list` | 通过 | 通过 | 通过，Total 962 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过，movie 链接使用 movie name | 通过，`thread_url` 渲染为外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-fetch-site-sehuatang-list.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-fetch-site-sehuatang-list.png` | 外链列补充真实链接渲染 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-fetch-site-sehuatang-list.png` | 达到 90% | in_progress |
| `/wdir` | 通过 | 通过 | 通过，Total 85 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过，目录 id 链接 | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-wdir.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-wdir.png` | 表头 ASC 箭头显示错误，已修复 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-wdir.png` | 达到 90% | in_progress |
| `/w-media-agg/birth` | 通过 | 通过 | 通过，Total 38 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过 | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-w-media-agg-birth.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-w-media-agg-birth.png` | 无真实 operation 链接已隐藏 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-w-media-agg-birth.png` | 达到 90% | in_progress |
| `/w-media-agg/bucket-list` | 通过 | 通过 | 通过，Total 38 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过 | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-w-media-agg-bucket-list.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-w-media-agg-bucket-list.png` | 无真实 operation 链接已隐藏 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-w-media-agg-bucket-list.png` | 达到 90% | in_progress |
| `/movie-agg-all/release` | 通过 | 通过 | 通过，Total 5445 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过 | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-movie-agg-all-release.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-movie-agg-all-release.png` | 无新增不足 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-movie-agg-all-release.png` | 达到 90% | in_progress |
| `/movie-agg-all/release-bucket-list` | 通过 | 通过 | 通过，Total 5445 | 通过，共用错误态 | 通过，共用空态 | 通过 | 通过 | 通过 | 无图片 | 通过 | 无外链 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-generic-movie-release-bucket-list.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-movie-release-bucket-list.png` | 无新增不足 | 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-generic-movie-release-bucket-list.png` | 达到 90% | in_progress |

验收动作：
- 逐页打开历史只读基准。
- 逐页打开 v2 页面。
- 保存历史基准截图和 v2 截图到 `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/`。
- 记录不足点清单。
- 按不足点修复并复查截图。
- 最终相似度判断必须达到 90%。
- 执行 `go test ./...`、`npm run test`、`npm run build` 和 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify`。
- 验证结束后关闭本轮临时启动的 v2 服务并确认 `2040`、`2041` 无监听。

本轮执行结果：
- 已接入 19 个 List 页面 v2 路由和侧边栏导航；`/cast/:id` 通过 `/casts` 表格链接进入。
- `generic` 页面已改为紧凑列表承载页：筛选工具栏、状态条、DataTable、表头排序、分页、空态、错误态。
- DataTable 已支持内部 `RouterLink` 与外部 `http/https` 链接。
- API 抽检 19 个页面均为 200，Total 已记录在逐页验收表。
- 浏览器交互验收：表头排序可点，升降序可切换并保留参数；分页可切到第 2 页；空态和错误态可见。
- 内部链接验收：movie 链接使用 movie name，cast 链接使用 id，wdir 链接使用目录 id。
- 外链验收：`/fetch-site-sehuatang-list` 的 `thread_url` 渲染为新标签外链。
- 控制台验收：正常页面无 warn/error；错误态截图中的 400 与 number input warning 为刻意触发错误态的验证证据。
- 全量验证：`go test ./...`、`npm run test`、`npm run build`、`make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify` 均通过。
- 端口清理：本轮临时启动的 v2 服务已关闭，`2040` Web 与 `2041` API 均无监听。

## 禁止项

- 禁止使用历史系统模板。
- 禁止使用历史系统 CSS。
- 禁止使用历史系统 JS。
- 禁止使用历史系统 class 命名体系。
- 禁止复制历史静态资源作为 v2 设计基础。
- 禁止把历史页面结构直接搬进 Vue。
- 禁止 pending action。
- 禁止通用假 action。
- 禁止空实现。
- 禁止假成功响应。
- 禁止禁用按钮冒充完成。
- 禁止因为构建通过、API 200、页面能打开就标完成。

## 验证命令记录

当前状态：`in_progress`

| 时间 | 命令 | 结果 | 说明 |
| --- | --- | --- | --- |
| 2026-05-27 | `npm install` | 通过 | 为新 v2 前端安装依赖并生成 lockfile。 |
| 2026-05-27 | `npm run test` | 通过 | `vue-tsc --noEmit`、`verify-source`、`verify-routes`、`verify-api-contracts` 均通过。 |
| 2026-05-27 | `npm run build` | 通过 | Vite 生产构建通过。 |

## 当前执行记录：movie detail MIDA-593

状态：`in_progress`

本轮只写：
- `/movie/:movie` v2 详情页。
- `GET /api/gc/v2/movie/:movie` 详情接口字段补齐。
- 固定只读验收样例：`MIDA-593`。

实现要求：
- 使用 v2 全新前端实现。
- 历史详情页只作为功能、字段、交互、截图和输出基准。
- 页面路由参数使用电影番号 `a_movie.name`。
- 页面链接必须使用后端返回的 name 语义链接，禁止用内部 `jav_id` 拼 `/movie/...`。
- 只渲染已有真实后端 API 支撑的动作；没有真实 API 的动作不渲染可用按钮。

功能对齐清单：
- 封面、番号、标题、中文标题、发行日期、maker、label、导演、prefix。
- 类型、演员、年龄展示、评分、观看数、rank 信息、need download 状态。
- media 信息、MPlay 链接、Bus/Suk/Lib/Sehuatang 等外链。
- SC/Come/最近 SC 相关信息。
- 加载态、错误态、空态、图片加载失败态。

验收动作：
- 打开历史只读基准 `/movie/MIDA-593`。
- 打开 v2 `/movie/MIDA-593`。
- MCP Chrome 截图观察并记录差异。
- 按差异修改并复查，直到达到本轮页面口径。
- 执行 `go test ./...`、`npm run test`、`npm run build` 和 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify`。
- 关闭本轮临时启动的 v2 服务并确认 2040/2041 无监听。

## 当前执行记录：rank day / rank period

状态：`in_progress`

本轮只写：
- `/moviecarddayrank` v2 页面。
- `/moviecardperiodrank` v2 页面。
- `GET /api/gc/v2/rank/day` 输出补齐。
- `GET /api/gc/v2/rank/period` 输出补齐。

实现要求：
- 使用 v2 全新前端实现。
- 历史 rank 页面只作为功能、字段、交互、截图和输出基准。
- 禁止读取或使用 `z_reference`。
- 禁止复用历史模板、CSS、JS、class、静态资源、页面壳和交互代码。
- 电影详情链接必须使用 movie name 或后端返回的 `movie_href`。
- 不渲染没有真实业务 API 支撑的动作按钮。

Day Rank 功能清单：
- 日期筛选。
- 前一天、后一天、随机、最新。
- 顶部和底部分页。
- 排名卡片、图片、电影详情链接。

Period Rank 功能清单：
- 周期类型、分类、周期 key。
- 前一周期、后一周期、随机、最新。
- 顶部和底部分页。
- 周期排名元数据。
- 排名卡片、图片、电影详情链接。

验收动作：
- 打开历史只读基准 `/moviecarddayrank`。
- 打开 v2 `/moviecarddayrank`。
- 打开历史只读基准 `/moviecardperiodrank`。
- 打开 v2 `/moviecardperiodrank`。
- MCP Chrome 截图观察并记录差异。
- 按差异修改并复查，直到达到本轮页面口径。
- 执行 `go test ./...`、`npm run test`、`npm run build` 和 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify`。
- 关闭本轮临时启动的 v2 服务并确认 2040/2041 无监听。

本轮执行结果：
- 后端补齐 `rank/day.latest_date`，用于“最新”动作。
- 后端补齐 `rank/period.prev_href`、`next_href`、`random_href`，用于周期导航。
- period 随机期数查询使用 `idx_period_range` 避免 filesort；本轮 EXPLAIN 结果为 `Using where; Backward index scan`。
- 前端新增 v2 自有 `/moviecarddayrank` 与 `/moviecardperiodrank` 页面，接入侧边栏和顶部快捷入口。
- 两个页面都使用 `MovieTile`、`PaginationControl`、`PageFrame`、`StateBlock` 和 v2 自有样式。
- 截图文件：
  - `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-moviecarddayrank.png`
  - `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-moviecarddayrank.png`
  - `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-moviecardperiodrank.png`
  - `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-moviecardperiodrank.png`
- 浏览器交互验收：
  - Day Rank 分页切到 `?p=2` 后首个排名为 `#19`。
  - Day Rank 前一天切到 `2026-05-25`。
  - Period Rank 分页切到 `?p=2` 后首个排名为 `#19`。
  - Period Rank 上一期切到 `Month 2026-04`。
- 控制台验收：v2 两页均无 warn/error。
- 全量验证：`make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify` 通过。
- 端口清理：`2040` Web 与 `2041` API 均已无监听。
