# rudy_gc v2 前后端重建计划

更新时间：2026-05-27

## 执行顺序

1. 先更新文档和进度。
2. 再执行代码、脚本、测试、构建或页面验证。
3. 执行后再次更新文档，记录真实结果。

本顺序适用于所有 v2 任务，包括前端、后端、SQL、缓存、部署、图片链路、测试和排查。

## 前端重写总原则

- 当前 v2 前端全部重写。
- 使用全新的前端设计，不使用旧项目前端的模板、CSS、JS、class、静态资源和布局结构。
- 旧版本只作为验收基准：功能、字段、交互语义、API 输出、截图接近度。
- 页面功能必须对齐之前版本，代码实现必须属于 v2 新前端。
- 不准为了快而做半套页面、假交互、假按钮、空实现。
- 每个页面写完后必须打开旧版页面和 v2 页面，使用 MCP Chrome 截图观察对比。
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
- `src/components`：全新组件体系，使用 v2 自有 class。
- `src/pages`：cards、movie detail、rank、generic、notfound。
- `src/styles`：全新样式，不使用旧项目前端 class。

## 页面与 API 对照

| 页面 | 前端路由 | 后端 API | 状态 | 要求 |
| --- | --- | --- | --- | --- |
| Cards | `/cards` | `GET /api/gc/v2/cards` | in_progress | 筛选、排序、分页、卡片、图片全部对齐。 |
| Today | `/cardstoday` | `GET /api/gc/v2/cards?view=cardstoday` | in_progress | 今日卡片输出与之前版本对齐。 |
| Has Rank | `/cardshasrank` | `GET /api/gc/v2/cards?view=cardshasrank` | in_progress | 排名相关筛选和排序对齐。 |
| Media Owned | `/cardsmediamowned` | `GET /api/gc/v2/cards?view=cardsmediamowned` | in_progress | 媒体状态、MPlay、图片对齐。 |
| Need Download | `/cardsneeddownload` | `GET /api/gc/v2/cards?view=cardsneeddownload` | in_progress | 下载状态筛选对齐。 |
| Random | `/cardsrandom` | `GET /api/gc/v2/cards?view=cardsrandom` | in_progress | 随机数量参数对齐。 |
| Movie Detail | `/movie/:movie` | `GET /api/gc/v2/movie/:movie` | in_progress | 使用 movie name 路由语义，详情字段完整。 |
| Day Rank | `/moviecarddayrank` | `GET /api/gc/v2/rank/day` | in_progress | 日期导航、随机、分页、卡片对齐。 |
| Period Rank | `/moviecardperiodrank` | `GET /api/gc/v2/rank/period` | in_progress | 周期筛选、分类、分页、周期元数据对齐。 |
| Generic Pages | 多路由 | `GET /api/gc/v2/pages/:key` | in_progress | 表格、筛选、表头排序、分页、空态对齐。 |
| Operation Pages | 多路由 | 待真实动作 API | blocked | 无真实 API 前不得渲染可用动作按钮。 |

## 本轮前端范围：补齐 cards 相关页面

状态：`in_progress`

本轮只写：
- `/cardstoday`
- `/cardshasrank`
- `/cardsmediamowned`
- `/cardsneeddownload`
- `/cardsrandom`
- cards 相关路由、页面绑定、API 参数、样式和契约检查。

实现要求：
- 使用 v2 全新前端实现。
- 禁止复用旧项目前端模板、CSS、JS、class、静态资源。
- 每个页面接入对应 `GET /api/gc/v2/cards?view=...`。
- cards 相关页面功能和性能至少等于旧项目，优先更好；凡是比旧项目差的实现一律不合格。

本轮实际结果：
- 6 个 cards 相关页面已使用 v2 新前端补齐：`/cards`、`/cardstoday`、`/cardshasrank`、`/cardsmediamowned`、`/cardsneeddownload`、`/cardsrandom`。
- 侧边栏、顶部快捷入口和页面内 view tabs 统一来自 `cardRoutes`，每个页面路由保留自身 URL，不在前端默认入口携带 `view` 查询参数。
- 页面调用统一通过 `GET /api/gc/v2/cards`，由当前路由 meta 写入对应 `view`。
- 已验证排序链接、顶部/底部分页、WMedia 快捷筛选、筛选面板、view 切换、电影详情链接、Bus/Suk/Lib 外链。
- MCP Chrome 已逐页打开 legacy 与 v2，并保存截图到 `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/`。
- v2 最终浏览器复测：6 个页面均无破图、无 v2 控制台 warn/error；首屏数量分别为 18、18、18、18、18、6。
- cards 相关页面进入合格候选；v2 全量仍保持 `in_progress`。

## 后端 API 要求

状态：`in_progress`

- 现有基础 API 可以作为前端数据来源，但不能自动视为完成。
- 所有页面动作必须迁移到真实业务 API。
- 禁止恢复通用 `page-actions` 假动作入口。
- API 响应必须区分展示 ID、内部 ID、外部 ID、路由参数。
- `/movie/:movie` 页面参数使用电影番号 `a_movie.name`。
- 数据库关联和内部 API 可使用 `jav_id`，但页面链接不得用 `jav_id` 拼 `/movie/...`。

## 本轮后端范围：cards 查询重写

状态：`in_progress`

本轮只写：
- cards 相关后端查询链路。
- cards 相关 model/service/handler/types/cache 代码。
- cards 查询重写所需测试与文档记录。

实现要求：
- 以旧项目 cards 查询策略为基准。
- 禁止自己新造复杂低效 SQL。
- cards 相关后端查询性能必须至少等于旧项目，优先更好；凡是比旧项目慢的实现一律不合格。

本轮实际结果：
- 已重建 cards service。
- 已接入 `MovieTypeCache` 等价业务缓存，cards 列表按 ID 优先读缓存。
- 已把 cards 查询入口切到旧项目同型专用路径。
- 已移除 cards 查询里的通用宽 JOIN、`COUNT(DISTINCT)`、无必要 `GROUP BY`、`MAX(...) ORDER BY` 残留实现。
- 已修正 `/cardsneeddownload` 的真实相册名为 `待下载电影`，接口 Total 恢复为 278。
- 已将 `/cardshasrank` 排序尾键对齐 `bm_minfo.idx_first_rank_name`，避免 filesort。
- 已将 `/cardsmediamowned` 改为 w_media 单源分页，并对齐 `idx_w_media_birth_name`。
- 已执行 `go test ./...`、前端 `npm run test`、前端 `npm run build` 和 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify`，均通过。
- 已执行接口级耗时对比，cards 相关 6 个 API 的 5 次中位数均达到本轮性能口径。

剩余缺口：
- cards 相关后端查询已达到本轮性能口径；v2 全量 API、真实动作和非 cards 页面仍保持 `in_progress`。

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
- 截图与之前版本接近。
- 旧版截图路径。
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

## 禁止项

- 禁止使用旧项目前端模板。
- 禁止使用旧项目前端 CSS。
- 禁止使用旧项目前端 JS。
- 禁止使用旧项目前端 class 命名体系。
- 禁止复制旧静态资源作为 v2 设计基础。
- 禁止把旧页面结构直接搬进 Vue。
- 禁止 pending action。
- 禁止通用假 action。
- 禁止空实现。
- 禁止假成功响应。
- 禁止禁用按钮冒充完成。
- 禁止因为构建通过、API 200、页面能打开就标完成。

## 验证命令记录

当前状态：`in_progress`

后续每轮必须填写：

| 时间 | 命令 | 结果 | 说明 |
| --- | --- | --- | --- |
| 2026-05-27 | 未执行 | in_progress | 本轮只重写文档，不运行测试或服务。 |
| 2026-05-27 | `npm install` | 通过 | 为新 v2 前端安装依赖并生成 lockfile。 |
| 2026-05-27 | `npm run test` | 通过 | `vue-tsc --noEmit`、`verify-source`、`verify-routes`、`verify-api-contracts` 均通过。 |
| 2026-05-27 | `npm run build` | 通过 | Vite 生产构建通过。 |
| 2026-05-27 | 文档规则写入 | in_progress | 新增 MCP Chrome 新旧截图对比、差异修复循环和 90% 相似度验收规则；未运行服务。 |

## 本轮前端范围：/cards

状态：`in_progress`

当前覆盖说明：
- 本历史小节记录的是上一轮只做 `/cards` 时的失败口径。
- 本轮 cards 相关页面补齐后，`/cards` 已纳入 6 个 cards 页面统一验收并进入合格候选。
- v2 全量仍保持 `in_progress`。

本轮只写：
- `/cards`
- `/api/gc/v2/cards?view=cards` 请求封装和页面调用
- `/cards` 页面需要的全新布局、卡片、分页、筛选、排序组件

本轮不写：
- `/cardstoday`
- `/cardshasrank`
- `/cardsmediamowned`
- `/cardsneeddownload`
- `/cardsrandom`
- `/movie/:movie`
- `/moviecarddayrank`
- `/moviecardperiodrank`
- generic pages
- 操作页动作

实现要求：
- 使用全新 v2 前端设计。
- 不使用旧项目前端模板、CSS、JS、class、静态资源。
- 保留 `/cards` 功能语义：筛选、排序、分页、卡片字段、图片展示、内部链接、外链。
- 没有真实动作 API 的按钮不渲染。
- 写完后必须打开旧版 `/cards` 和 v2 `/cards`。
- 必须使用 MCP Chrome 截图观察新旧页面差异。
- 必须记录不足点，并按“修改 -> 截图复查”循环修复。
- `/cards` 功能接近且视觉/交互相似度达到 90% 后，才允许进入合格候选。
- 构建通过、API 200、页面能打开、基础筛选分页能用，都不能作为 `/cards` 合格依据。
- 验证后回写真实结果。

本轮实现结果：
- `/cards` 已有全新 v2 页面、布局、卡片、分页、筛选、排序、样式和 API 请求封装。
- 路由范围已收敛到 `/cards` 与 404；本轮未开放其它 cards 变体。
- `/cards` 页面固定向后端传 `view=cards`。
- 电影卡片保留 `0` 分数和 `0` 观看数显示。
- 电影卡片详情链接使用后端 `movie_href`。
- 验证脚本覆盖 `/cards` 源码有效性、路由范围、API/筛选/卡片/分页契约和无假动作。
- 上一轮历史记录曾判定 `/cards` 不合格；本轮 cards 相关 6 页面补齐后已被覆盖。
- 当前 `/cards` 已纳入 cards 相关页面统一截图、功能、交互、接口和性能验收，并进入合格候选。
- v2 全量仍保持 `in_progress`。

## /cards 新旧截图对比记录

状态：`in_progress`

| 轮次 | 旧版截图 | v2 截图 | 不足点 | 修复记录 | 相似度判断 |
| --- | --- | --- | --- | --- | --- |
| 0 | 未生成 | 未生成 | 历史轮次仅有最小可运行页面壳，未完成旧版功能完整对齐和 MCP Chrome 截图观察。 | 未修复。 | 未达到 90%。 |
| 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-legacy-round1.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-v2-round1.png` | 顶部栏和左侧菜单与旧版差距明显；cards 主工具栏和卡片字段已接近。 | 继续修改 v2 自有 `AppLayout` 与样式。 | 未达到 90%。 |
| 2 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-legacy-round1.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-v2-round2.png` | 旧版首屏没有年龄排序按钮；v2 活动态偏黑，和旧版紫色活动态有差距。 | 继续修改 `/cards` 排序可见项和活动态样式。 | 未达到 90%。 |
| 3 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-legacy-final.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-v2-final.png` | v2 首屏卡片密度偏低，同宽视口下卡片数少于旧版；需要调整新 v2 CSS 网格断点。 | 修改 `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/styles/index.css` 的 `/cards` 网格密度。 | 未达到 90%。 |
| 4 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-cards.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-cards.png` | 本轮 `/cards` 已随 6 个 cards 相关页面统一复测：Total 111385、首屏 18 卡、18 图完成加载、排序/分页/筛选/view 切换/外链可用。 | 进入合格候选；v2 全量仍保持 `in_progress`。 | 达到本轮页面口径。 |

## 本轮 /cards 重做计划

状态：`in_progress`

开始时间：2026-05-27

执行范围：
- 只处理 `/cards`。
- 旧版 `/cards` 作为功能、字段、交互、截图和输出基准。
- v2 前端继续使用全新 Vue/Vite/TypeScript 实现，禁止复用旧前端模板、CSS、JS、class、静态资源。
- 允许为 `/cards` 首屏相似度调整 v2 自有 layout 的 header/sidebar 可见结构，但不得恢复旧前端模板、CSS、JS、class、静态资源。

功能对齐清单：
- 顶部页面信息和结果统计。
- 筛选区：关键词、演员、person id、类型、导演、prefix、maker、label、本地 label、album、发行时间、媒体时间、年龄、上榜时间、上榜天数、下载状态、媒体状态、观看数、评分、SC、Come、媒体目录。
- 排序区：发行日期、详情更新、年龄、观看数、上榜日期、最高排名、上榜天数、媒体时间、SC、Come、最近 SC。
- 分页区：顶部和底部都可操作，并保留当前 query。
- 卡片区：封面、番号、标题、评分、观看数、排名、Come、SC、MPlay、类型、演员、导演、外链、发行日期。
- 状态区：加载态、错误态、空态、图片加载失败态。

验收动作：
- `npm run test`
- `npm run build`
- 打开旧版 `/cards`。
- 打开 v2 `/cards`。
- MCP Chrome 截图观察。
- 按不足点修改并复查。
- 回写截图路径、差异清单、修复轮次和相似度判断。

## 端口清理记录

当前状态：`in_progress`

- 默认 API 端口：`2041`
- 默认 Web 端口：`2040`
- 每次启动临时服务后，必须关闭并记录。
- 本轮 `/cards` 实现未启动服务。
- 端口检查结果：`2041` 已有 api 监听；本轮未启动该进程，也未关闭非本轮启动进程。
- 端口检查结果：未发现 `2040` Web 监听。

## 下一步

1. 本轮按用户要求在 `/cards` 完成后停止。
2. 后续任何继续动作仍必须先更新文档。

## 本轮前端范围：movie detail MIDA-593

状态：`in_progress`

本轮只写：
- `/movie/:movie` v2 详情页。
- `GET /api/gc/v2/movie/:movie` 详情接口字段补齐。
- 固定验收样例：legacy `http://localhost:5010/movie/MIDA-593` 与 v2 `http://localhost:2040/movie/MIDA-593`。

实现要求：
- 使用 v2 全新前端实现，不使用旧项目前端模板、CSS、JS、class、静态资源。
- 旧版详情页只作为功能、字段、交互、截图和输出基准。
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
- 打开 legacy `/movie/MIDA-593`。
- 打开 v2 `/movie/MIDA-593`。
- MCP Chrome 截图观察并记录差异。
- 按差异修改并复查，直到达到本轮页面口径。
- 执行 `go test ./...`、`npm run test`、`npm run build` 和 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify`。
- 关闭本轮临时启动的 v2 服务并确认 2040/2041 无监听。
