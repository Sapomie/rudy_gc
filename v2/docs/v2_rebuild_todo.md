# rudy_gc v2 重建 TODO

更新时间：2026-05-27

## 最高执行口径

- v2 全量验收合格前，禁止停止在阶段性进度。
- 阶段性构建通过、局部测试通过、单页可打开、API 返回 200、路由存在、文档写完，都不能作为完成理由。
- 所有页面、API、真实动作链路、图片链路、SQL/性能、输出对比、截图对比、测试、Make、部署文档、端口清理全部通过后，才允许标记最终完成。
- 未完成全量验收前，状态只能使用 `in_progress`、`blocked`、`invalid`、`todo`，禁止写 `done`、`completed`、`finished` 等最终完成词。

## 页面截图循环验收硬规则

- 每次写完一个 v2 页面，必须打开旧版本页面和 v2 页面。
- 必须使用 MCP Chrome 对旧版本页面和 v2 页面分别截图，并人工观察差异。
- 页面验收必须循环执行：截图对比 -> 记录不足点 -> 修改 -> 再截图对比。
- 只要功能、字段、布局、交互、图片、分页、筛选、排序、按钮、链接、空态、错误态仍有明显差异，就必须继续修改。
- 页面功能接近且视觉/交互相似度达到 90% 后，才允许进入合格候选。
- 未达到 90% 相似度前，页面状态只能保持 `in_progress` 或 `invalid`。
- 禁止用随便写一个页面、基础页面壳、局部功能可用、构建通过、API 有返回、页面能打开、单次截图通过来结束页面任务。
- 每个页面必须记录：旧版截图路径、v2 截图路径、不足点清单、修复轮次、复查截图、最终相似度判断。

## 文档先行硬规则

- 任何 v2 实现任务必须先更新本文档和 `/Users/gaojinwei/src/my/rudy_gc/v2/docs/gc_frontend_backend_plan.md`，再写代码、跑脚本、测试、构建或页面验证。
- 涉及运行、端口、容器、部署、静态资源、图片代理时，必须同步更新 `/Users/gaojinwei/src/my/rudy_gc/v2/deploy/README.md`。
- 每轮代码前，文档必须记录：本轮目标、修改范围、当前状态、风险、验收口径、下一步执行清单。
- 每轮代码后，文档必须记录：实际完成、失败原因、剩余缺口、验证结果、端口清理状态。
- 禁止用聊天记录、口头说明、临时计划、测试输出替代文档进度。

## 前端重写硬规则

- 当前 v2 前端按“全部重写”处理，使用全新 Vue/Vite/TypeScript 前端设计。
- 禁止使用旧项目前端的模板、CSS、JS、class、静态资源、布局结构、页面壳、交互代码。
- 禁止从旧项目前端复制或改写模板片段、样式片段、脚本片段。
- 旧版本只能作为功能结果、字段完整性、交互语义、API 输出、截图接近度的验收基准。
- 页面视觉可以接近之前版本，但实现必须是 v2 新前端自己的组件、样式、路由和状态管理。
- 当前被中断生成的 v2 前端临时文件不能视为完成态；后续必须按本文档重新审查、覆盖、补齐和验证。

## 必须对齐的功能范围

### Movie Cards

状态：`in_progress`

必须覆盖：
- `/cards`
- `/cardstoday`
- `/cardshasrank`
- `/cardsmediamowned`
- `/cardsneeddownload`
- `/cardsrandom`
- 筛选参数：演员、person id、类型、导演、prefix、maker、label、本地 label、album、发行时间、媒体时间、年龄、上榜时间、上榜天数、下载状态、媒体状态、观看数、评分、SC、Come、媒体目录、关键词。
- 排序参数：发行日期、详情更新、年龄、观看数、上榜日期、最高排名、上榜天数、媒体时间、SC、Come、最近 SC。
- 分页：顶部和底部都必须可用。
- 卡片字段：封面、番号、标题、评分、观看数、排名、Come、SC、MPlay、类型、演员、导演、外链、发行日期。
- 图片：v2 前端、API 静态代理、本地路径、容器路径都必须显示。

验收：
- 与之前版本逐页截图对比。
- API 输出数量、排序、字段完整性对比。
- 图片请求全部成功。
- 性能不得明显低于之前版本。

### Movie Detail

状态：`in_progress`

必须覆盖：
- `/movie/:movie`
- 路由参数使用电影番号 `a_movie.name`。
- 页面链接必须使用后端返回的 `movie_href` 或 name 语义，禁止用内部 `jav_id` 拼页面 URL。
- 展示字段：封面、番号、标题、发行、maker、label、导演、类型、演员、评分、观看数、need download、媒体信息、rank info、SC events、JavBus、Sukebei、Sehuatang。
- 真实动作：只有后端已有真实业务 API 时才渲染可用按钮。
- 禁止用假按钮、禁用按钮、pending action、空实现、假成功响应冒充完成。

验收：
- 截图对比。
- 详情 API 字段对比。
- movie name 与 jav_id 语义专项检查。
- 所有外链和内部链接语义正确。

### Rank

状态：`in_progress`

必须覆盖：
- `/moviecarddayrank`
- `/moviecardperiodrank`
- 日期/周期筛选。
- 前后跳转、随机、最新。
- 分页。
- 日榜卡片展示。
- 周期榜卡片展示和周期排名元数据。

验收：
- day rank 与 period rank 截图对比。
- API 输出数量、排序、字段完整性对比。
- 周期导航参数对比。

### Generic List / Operation Pages

状态：`in_progress`

必须覆盖：
- casts、records、crawl-records、sc-events、medias、w-agg-events、e-items、torrent-albums、movie-albums、d-seeds。
- fetch-site-javbus-list、fetch-site-sukebei-list、fetch-site-sehuatang-list。
- wdir、w-media-agg、movie release agg。
- crawler、triggers、media、SC、agg 等操作入口。

硬规则：
- List 页必须有筛选、表头排序、分页、空态、链接列、状态信息。
- Operation 页没有真实业务 API 前，只能展示状态和说明，禁止渲染假动作按钮。
- 所有真实按钮必须调用真实业务 API。

验收：
- 每个页面路由可访问。
- 导航可进入。
- 筛选、排序、分页可用。
- 表头排序可点、可切换升降序、保留筛选参数、有激活态和方向提示。

## API 与性能验收

状态：`in_progress`

每个 API 必须对比：
- method
- path
- query/body 参数
- 输出字段
- 输出数量
- 排序结果
- SQL 链路
- SQL 次数
- 执行计划
- 耗时

性能硬规则：
- 能走缓存的唯一查询必须走缓存方法。
- 写查询前先评估索引、扫描行数、排序代价、回表次数。
- 不得新增隐式阈值、上限、下限、硬过滤。
- 数值筛选必须区分未传值和值为 0，0 是合法业务值。
- 出现 filesort、临时表、无意义 IN/EXISTS、重复扫表时必须继续优化。

## 图片链路验收

状态：`in_progress`

必须覆盖：
- API 静态路由。
- Vite 开发代理。
- 生产 Web 代理。
- Compose 只读挂载。
- 本地真实文件路径。
- 容器运行口径。

验收：
- Cards 首屏图片全部成功。
- Movie Detail 图片成功。
- Rank 卡片图片成功。
- 容器环境图片路径成功。

## 当前真实进度记录

| 模块 | 状态 | 当前记录 |
| --- | --- | --- |
| 文档 | in_progress | 已写入用户要求和验收口径；本轮新增页面截图循环验收硬规则和 90% 相似度口径。 |
| 前端 | in_progress | cards 相关 6 个页面已用 v2 全新前端补齐并进入合格候选：`/cards`、`/cardstoday`、`/cardshasrank`、`/cardsmediamowned`、`/cardsneeddownload`、`/cardsrandom`；v2 全量仍保持 `in_progress`。 |
| 后端 API | in_progress | cards API 已按专用查询路径和缓存链路修复并完成接口级性能对比；movie/rank/pages 与真实动作链路仍需逐项迁移和验收。 |
| 图片链路 | in_progress | 需要重新按开发、API、部署、容器全链路验收。 |
| 部署 | in_progress | 需要随前端重写同步构建、代理、端口、容器说明。 |
| 测试/Make | in_progress | 已执行 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify`，包含后端 `go test ./...`、前端 `npm run test`、路由/API/source 契约检查和生产构建，均通过；v2 全量仍保持 `in_progress`。 |
| 端口清理 | in_progress | 本轮临时启动的 v2 `2040` Web 与 `2041` API 已通过 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 stop-all` 关闭，`lsof` 确认两个端口无监听。 |

## 本轮执行范围：补齐 cards 相关页面

状态：`in_progress`

本轮目标：
- 补齐 `/cardstoday`、`/cardshasrank`、`/cardsmediamowned`、`/cardsneeddownload`、`/cardsrandom`。
- 使用 v2 全新前端实现，禁止复用旧项目前端模板、CSS、JS、class、静态资源。
- 每个页面接入对应 cards view 参数。

本轮验收口径：
- cards 相关页面功能和性能至少等于旧项目，优先更好；凡是比旧项目差的实现一律不合格。

本轮执行清单：
- 读取旧项目 cards 相关页面入口和当前 v2 前端现状。
- 补齐 cards 相关路由、页面绑定、API 参数、样式和契约检查。
- 执行必要测试、构建和页面对比。
- 回写本文档和 `/Users/gaojinwei/src/my/rudy_gc/v2/docs/gc_frontend_backend_plan.md` 的真实结果。

本轮实际结果：
- 已补齐 cards 相关 6 个 v2 前端路由：`/cards`、`/cardstoday`、`/cardshasrank`、`/cardsmediamowned`、`/cardsneeddownload`、`/cardsrandom`。
- `cardRoutes` 已成为侧边栏、顶部快捷入口和页面内 view tabs 的同一数据源，每个路由向 `GET /api/gc/v2/cards` 传入对应 `view`。
- 页面保留 v2 全新组件、样式与 class，旧项目前端只作为截图和字段/功能基准。
- 已验证排序、分页、WMedia 快捷筛选、筛选面板、view 切换、电影详情链接、Bus/Suk/Lib 外链存在性。
- 已使用 MCP Chrome 对 legacy 与 v2 的 6 个 cards 页面逐页截图；v2 首屏图片 0 破图、控制台 0 个 v2 warn/error。
- 截图路径：
  - legacy：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-cards.png`、`legacy-cardstoday.png`、`legacy-cardshasrank.png`、`legacy-cardsmediamowned.png`、`legacy-cardsneeddownload.png`、`legacy-cardsrandom.png`
  - v2：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-cards.png`、`v2-cardstoday.png`、`v2-cardshasrank.png`、`v2-cardsmediamowned.png`、`v2-cardsneeddownload.png`、`v2-cardsrandom.png`

本轮页面验收记录：
- `/cards`：Total 111385，首屏 18 卡，18 图完成加载。
- `/cardstoday`：Total 110982，首屏 18 卡，18 图完成加载。
- `/cardshasrank`：Total 16646，首屏 18 卡，18 图完成加载。
- `/cardsmediamowned`：Total 846，首屏 18 卡，18 图完成加载。
- `/cardsneeddownload`：Total 278，首屏 18 卡，18 图完成加载。
- `/cardsrandom`：Total 111385，首屏 6 卡，6 图完成加载。
- 当前 cards 相关页面进入合格候选；v2 全量仍保持 `in_progress`。

## 本轮执行范围：cards 后端查询重写

状态：`in_progress`

本轮目标：
- 只重写 cards 相关后端查询链路。
- 旧 cards 查询链路中偏离旧项目性能路径的实现全部视为 `invalid`。
- 以旧项目 cards 查询策略为基准重建 v2 cards model/service/handler/types/cache 相关代码。
- 禁止自己新造复杂低效 SQL。

本轮验收口径：
- cards 相关后端查询性能必须至少等于旧项目，优先更好；凡是比旧项目慢的实现一律不合格。

本轮执行清单：
- 读取旧项目 cards 的 handler、service、model 查询和业务缓存链路。
- 读取当前 v2 cards 后端现状。
- 按旧项目高性能链路重建 v2 cards 查询。
- 执行必要编译、测试和性能对比。
- 回写本文档和 `/Users/gaojinwei/src/my/rudy_gc/v2/docs/gc_frontend_backend_plan.md` 的真实结果。

本轮实际结果：
- 已恢复 v2 后端编译所需的 modelx 与非 cards service 骨架。
- 已重建 cards service。
- cards service 已接入 `MovieTypeCache` 等价业务缓存，先按 ID 命中缓存，未命中再读取 DB 并写回缓存。
- cards 查询入口已切到旧项目同型链路：按来源表查 ID、需要时取交集、再按排序来源分页。
- 默认 `/cards` 走 `a_movie` 单源路径。
- `od=sc` 走 `w_media + g_sc_stat` 专用路径。
- 卡片详情已停止使用宽 JOIN 一次性拼装，改为按 ID 小查询组装，配合业务缓存复用。
- 已清理 cards 查询里残留的 `COUNT(DISTINCT)`、无必要 `GROUP BY`、`MAX(...) ORDER BY` 旧实现文本。
- 已执行 `go test ./...`，通过。
- 已执行只读 SQL 耗时抽样：默认 cards 核心 SQL 约 0.08/0.04/0.03s，`od=sc` 专用 SQL 约 0.01/0.01/0.01s。
- 本轮后续启动 v2 API/Web 做接口和页面实测，收尾已关闭 2040/2041 并确认端口无监听。

本轮接口级实测结果：
- 5 次耗时中位数对比：
  - `/cards`：v2 0.0115s，legacy 0.0129s。
  - `/cardstoday`：v2 0.0129s，legacy 0.0148s。
  - `/cardshasrank`：v2 0.0039s，legacy 0.0059s。
  - `/cardsmediamowned`：v2 0.0032s，legacy 0.0146s。
  - `/cardsneeddownload`：v2 0.0050s，legacy 0.0064s。
  - `/cardsrandom`：v2 0.2238s，legacy 0.2885s。
- `/cardsneeddownload` 常量已修正为真实相册名 `待下载电影`，Total 从 0 恢复为 278。
- `/cardshasrank` 排序已对齐 `bm_minfo.idx_first_rank_name`，EXPLAIN 为 `Using where; Backward index scan`，无 filesort。
- `/cardsmediamowned` 已改为 w_media 单源分页并对齐 `idx_w_media_birth_name`，EXPLAIN 为 `Using where`，无 filesort。

剩余缺口：
- cards 后端查询已达到本轮性能口径；v2 全量 API、真实动作和非 cards 页面仍保持 `in_progress`。

## 本轮执行范围：只做 /cards

状态：`in_progress`

当前覆盖说明：
- 本历史小节记录的是上一轮只做 `/cards` 时的失败口径。
- 本轮 cards 相关页面补齐后，`/cards` 已被纳入 6 个 cards 页面统一验收并进入合格候选。
- v2 全量仍保持 `in_progress`，禁止把 `/cards` 合格候选当作 v2 最终完成。

本轮目标：
- 只实现 `/cards` 页面。
- 调用 `GET /api/gc/v2/cards?view=cards`。
- 完成 `/cards` 的全新 v2 前端设计、筛选、排序、分页、电影卡片、图片展示、链接语义、旧版功能完整对齐和新旧截图循环验收。
- `/cards` 写完、验证、回写文档后停止。

本轮不做：
- `/cardstoday`
- `/cardshasrank`
- `/cardsmediamowned`
- `/cardsneeddownload`
- `/cardsrandom`
- movie detail
- rank day / rank period
- generic pages
- 任何没有真实 API 的动作按钮

验收口径：
- `/cards` 源码必须是有效 Vue SFC。
- 前端必须使用全新 v2 设计，禁止旧项目前端模板、CSS、JS、class、静态资源。
- 旧版 `/cards` 只能作为功能、字段、交互和截图验收基准，禁止复用旧前端实现。
- 筛选、排序、分页参数必须透传到 `/api/gc/v2/cards`。
- 卡片字段必须包含封面、番号、标题、评分、观看数、排名、Come、SC、MPlay、类型、演员、导演、外链、发行日期。
- `0` 分数和 `0` 观看数必须可显示，不能当作空值隐藏。
- 电影详情链接必须使用后端返回的 `movie_href`。
- 写完后必须打开旧版 `/cards` 和 v2 `/cards`，使用 MCP Chrome 截图观察对比。
- 必须记录不足点清单，并按“修改 -> 截图复查”循环修复。
- `/cards` 功能接近且视觉/交互相似度达到 90% 后，才允许进入合格候选。
- 构建通过、API 200、页面能打开、基础筛选分页能用，都不能作为 `/cards` 合格依据。
- 验证完成后必须回写本文档和 `/Users/gaojinwei/src/my/rudy_gc/v2/docs/gc_frontend_backend_plan.md`。

本轮实际结果：
- `/cards` 路由已收敛为本轮唯一 cards 页面入口，根路径重定向到 `/cards`。
- 已实现 `/cards` 的全新 v2 布局、筛选表单、排序、分页、卡片网格、图片展示、电影链接、外链和错误/空态。
- 已固定请求 `GET /api/gc/v2/cards?view=cards`，本轮不暴露其它 cards 变体入口。
- 已使用后端返回的 `movie_href` 作为电影详情链接。
- 已用 `Number.isFinite` 保留 `0` 分数和 `0` 观看数显示。
- 已新增 `/cards` 范围验证脚本，禁止本轮恢复假 `page-actions`。
- 已执行 `npm install` 生成前端依赖锁文件。
- 已执行 `npm run test`，通过。
- 已执行 `npm run build`，通过。
- 上一轮历史记录曾判定 `/cards` 不合格；本轮 cards 相关 6 页面补齐后已被覆盖。
- 当前 `/cards` 已纳入 cards 相关页面统一截图、功能、交互、接口和性能验收，并进入合格候选。
- v2 全量仍保持 `in_progress`，禁止把 `/cards` 合格候选当作 v2 最终完成。

## /cards 截图循环验收记录

状态：`in_progress`

| 轮次 | 旧版截图 | v2 截图 | 不足点 | 处理结果 |
| --- | --- | --- | --- | --- |
| 0 | 未生成 | 未生成 | 历史轮次仅有最小可运行页面壳，旧版功能完整度、字段展示、交互、布局、图片链路、分页位置、筛选排序和链接语义均未完成新旧截图对比。 | 判定不合格，必须重做后进入截图循环。 |
| 1 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-legacy-round1.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-v2-round1.png` | cards 工具栏、WMedia、分页、卡片字段已接近；顶部栏仍显示 v2 breadcrumb/单个 Cards 快捷入口，左侧菜单只有 Cards，与旧版首屏菜单结构差距明显。 | 继续修改 v2 layout 的 header/sidebar 可见结构，再截图复查。 |
| 2 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-legacy-round1.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-v2-round2.png` | 顶部栏和左侧菜单已接近；排序栏额外显示年龄降序/年龄升序两个旧版首屏没有的按钮；活动态颜色仍偏黑，旧版为紫色系。 | 继续收敛排序栏和活动态颜色，再截图复查。 |
| 3 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-legacy-final.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/cards-v2-final.png` | 顶部栏、左侧菜单、工具栏和卡片内容已接近；v2 首屏卡片宽度偏大，旧版同宽视口约可见 5 张完整卡，v2 约 4 张多一点。 | 继续调整 v2 自有 CSS 网格断点和卡片密度，再截图复查。 |
| 4 | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-cards.png` | `/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-cards.png` | 本轮 `/cards` 已随 6 个 cards 相关页面统一复测：Total 111385、首屏 18 卡、18 图完成加载、排序/分页/筛选/view 切换/外链可用。 | 进入合格候选；v2 全量仍保持 `in_progress`。 |

## 本轮 /cards 重做执行记录

状态：`in_progress`

本轮开始时间：2026-05-27

本轮目标：
- 只重做 `/cards` 页面，不扩展到其它页面。
- 旧版 `/cards` 只作为功能、字段、交互、截图和输出基准，禁止复用旧前端模板、CSS、JS、class、静态资源。
- 补齐 `/cards` 筛选、排序、顶部/底部分页、卡片字段、图片、内部链接、外链、加载态、空态、错误态、URL query 同步。
- 写完后打开旧版 `/cards` 和 v2 `/cards`，使用 MCP Chrome 截图观察。
- 记录每轮不足点，继续修改并复查，直到功能接近且视觉/交互相似度达到 90%。

本轮计划修改范围：
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/layouts/AppLayout.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/pages/cards/index.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/components/MovieTile.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/components/MovieGrid.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/components/PaginationControl.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/api/gc.ts`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/types/api.ts`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/styles/index.css`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/scripts/verify-api-contracts.mjs`

本轮验收口径：
- `npm run test` 通过。
- `npm run build` 通过。
- 旧版 `/cards` 与 v2 `/cards` 均可打开。
- MCP Chrome 生成旧版截图和 v2 截图。
- 首屏图片请求成功。
- 不足点清单回写本文档。
- 达到 90% 相似度后，`/cards` 仍只能进入合格候选；v2 全量完成前不得标最终完成。

## 下一步执行清单

1. 继续 `/cards` 时，必须先按本文档补全旧版功能对齐清单。
2. 重新实现 `/cards` 完整功能。
3. 打开旧版 `/cards` 与 v2 `/cards`，使用 MCP Chrome 截图对比。
4. 记录不足点并反复修改，直到功能接近且视觉/交互相似度达到 90%。

## 本轮执行范围：movie detail MIDA-593

状态：`in_progress`

本轮目标：
- 重写 v2 电影详情页 `/movie/:movie`。
- 固定对比样例：legacy `http://localhost:5010/movie/MIDA-593` 与 v2 `http://localhost:2040/movie/MIDA-593`。
- 路由参数必须使用电影番号 `a_movie.name`，禁止使用内部 `jav_id` 拼页面 URL。
- 旧版详情页只作为功能、字段、交互、截图和输出基准，禁止复用旧前端模板、CSS、JS、class、静态资源。

本轮计划修改范围：
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/router/handler/movie`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/service/movie`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/model/modelx`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/types`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/pages/movie/index.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/api/gc.ts`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/types/api.ts`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/styles/index.css`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/scripts/verify-api-contracts.mjs`

本轮验收口径：
- legacy `/movie/MIDA-593` 与 v2 `/movie/MIDA-593` 必须逐页打开并截图对比。
- 详情字段、图片、外链、内部链接、媒体信息、rank 信息、SC 信息和真实动作可见性必须接近旧版。
- 没有真实业务 API 的动作禁止渲染成可用按钮。
- 图片请求必须成功，v2 控制台不得有 warn/error。
- 后端详情 API 查询链路必须至少达到旧版同等性能，优先更好。
- 写完后执行后端测试、前端测试、构建、`make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify`。
- 验证结束必须关闭本轮临时启动的 2040/2041 端口进程。
- 最终仍只能进入合格候选；v2 全量完成前保持 `in_progress`。
