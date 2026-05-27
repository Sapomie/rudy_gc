# rudy_gc v2 重建 TODO

更新时间：2026-05-27

## 当前维护口径

- 后续只更新 `v2/` 功能。
- 历史系统不会继续更新，历史代码和历史前端资源不得作为 v2 实现规则、复用来源或风格约束。
- 历史系统仅作为只读验收基准，用于核对功能结果、字段完整性、交互语义、API 输出和截图接近度。
- v2 的实现必须来自 `v2/gc-api`、`v2/gc-web`、`v2/docs`、`v2/deploy` 自身。
- 禁止使用 `z_reference`、历史模板、历史 CSS、历史 JS、历史 class、历史静态资源、历史页面壳和历史交互代码。

## 最高执行口径

- v2 全量验收合格前，禁止停止在阶段性进度。
- 阶段性构建通过、局部测试通过、单页可打开、API 返回 200、路由存在、文档写完，都不能作为完成理由。
- 所有 v2 页面、v2 API、真实动作链路、图片链路、SQL/性能、输出对比、截图对比、测试、Make、部署文档、端口清理全部通过后，才允许标记最终完成。
- 未完成全量验收前，状态只能使用 `in_progress`、`blocked`、`invalid`、`todo`，禁止写 `done`、`completed`、`finished` 等最终完成词。

## 页面截图循环验收

- 每次写完一个 v2 页面，必须打开历史只读基准页面和 v2 页面。
- 必须使用 MCP Chrome 分别截图，并人工观察差异。
- 页面验收必须循环执行：截图对比 -> 记录不足点 -> 修改 -> 再截图对比。
- 只要功能、字段、布局、交互、图片、分页、筛选、排序、按钮、链接、空态、错误态仍有明显差异，就必须继续修改。
- 页面功能接近且视觉/交互相似度达到 90% 后，才允许进入合格候选。
- 每个页面必须记录：历史基准截图路径、v2 截图路径、不足点清单、修复轮次、复查截图、最终相似度判断。

## 文档先行硬规则

- 任何 v2 实现任务必须先更新本文档和 `/Users/gaojinwei/src/my/rudy_gc/v2/docs/gc_frontend_backend_plan.md`，再写代码、跑脚本、测试、构建或页面验证。
- 涉及运行、端口、容器、部署、静态资源、图片代理时，必须同步更新 `/Users/gaojinwei/src/my/rudy_gc/v2/deploy/README.md`。
- 每轮代码前，文档必须记录：本轮目标、修改范围、当前状态、风险、验收口径、下一步执行清单。
- 每轮代码后，文档必须记录：实际完成、失败原因、剩余缺口、验证结果、端口清理状态。
- 禁止用聊天记录、口头说明、临时计划、测试输出替代文档进度。

## v2 前端硬规则

- 当前 v2 前端按“全部重写”处理，使用全新 Vue/Vite/TypeScript 前端设计。
- 只允许使用 v2 自己的组件、样式、路由、状态管理和 API 封装。
- 禁止使用历史系统的模板、CSS、JS、class、静态资源、布局结构、页面壳、交互代码。
- 禁止从历史系统复制或改写模板片段、样式片段、脚本片段。
- 页面视觉可以接近历史只读基准，但实现必须是 v2 新前端自己的代码。
- 不准为了快而做半套页面、假交互、假按钮、空实现、假成功响应或禁用按钮冒充完成。

## 必须覆盖的功能范围

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

当前记录：
- 6 个 cards 相关页面已进入合格候选：`/cards`、`/cardstoday`、`/cardshasrank`、`/cardsmediamowned`、`/cardsneeddownload`、`/cardsrandom`。
- cards API 已改为 v2 自有专用查询链路并接入 `MovieTypeCache`。
- 已执行 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify` 并通过。
- v2 全量仍保持 `in_progress`。

### Movie Detail

状态：`in_progress`

必须覆盖：
- `/movie/:movie`
- 路由参数使用电影番号 `a_movie.name`。
- 页面链接必须使用后端返回的 `movie_href` 或 name 语义，禁止用内部 `jav_id` 拼页面 URL。
- 展示字段：封面、番号、标题、发行、maker、label、导演、类型、演员、评分、观看数、need download、媒体信息、rank info、SC events、JavBus、Sukebei、Sehuatang。
- 真实动作：只有后端已有真实业务 API 时才渲染可用按钮。

当前记录：
- 已进入本轮实现与截图差异修复阶段。
- 固定只读验收样例为 `MIDA-593`。
- v2 全量仍保持 `in_progress`。

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

## API 与性能验收

状态：`in_progress`

每个 API 必须核对：
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
| 文档 | in_progress | 已清理历史系统实现规则，保留历史系统只读验收基准口径。 |
| 前端 | in_progress | cards 相关 6 个页面已用 v2 全新前端补齐并进入合格候选；movie detail、rank、generic、operation 仍需继续覆盖和验证。 |
| 后端 API | in_progress | cards API 已按 v2 专用查询路径和缓存链路修复；movie/rank/pages 与真实动作链路仍需逐项迁移和验收。 |
| 图片链路 | in_progress | 需要重新按开发、API、部署、容器全链路验收。 |
| 部署 | in_progress | 需要随前端重写同步构建、代理、端口、容器说明。 |
| 测试/Make | in_progress | 已有 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify` 通过记录；后续每轮仍需重新验证。 |
| 端口清理 | in_progress | 每轮临时启动的 v2 `2040` Web 与 `2041` API 都必须在验证结束后关闭。 |

## 下一步执行清单

1. 当前执行 rank day / rank period 页面建立。
2. 所有新实现只使用 v2 自有代码体系。
3. 历史系统只作为只读截图、字段和输出基准。

## 本轮执行范围：rank day / rank period

状态：`in_progress`

本轮目标：
- 建立 v2 `/moviecarddayrank` 页面。
- 建立 v2 `/moviecardperiodrank` 页面。
- 两个页面都使用 v2 自有 Vue/Vite/TypeScript 前端、v2 API、v2 类型和 v2 样式体系。
- 历史系统只作为只读验收基准，禁止读取或复用历史前端实现。

本轮计划修改范围：
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/types/rank.go`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/service/rank`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-api/internal/model/modelx/movie_read_repo.go`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/router/index.ts`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/data/routes.ts`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/api/gc.ts`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/types/api.ts`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/pages/rank/day.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/pages/rank/period.vue`
- `/Users/gaojinwei/src/my/rudy_gc/v2/gc-web/src/styles/index.css`
- 必要时更新 v2 verify 脚本与 `/Users/gaojinwei/src/my/rudy_gc/v2/deploy/README.md`

Day Rank 功能清单：
- 日期筛选。
- 前一天、后一天、随机、最新。
- 顶部和底部分页。
- 排名卡片展示。
- 图片展示。
- 电影详情链接使用 movie name / `movie_href`。

Period Rank 功能清单：
- 周期类型。
- 分类。
- 周期 key。
- 前一周期、后一周期、随机、最新。
- 顶部和底部分页。
- 周期排名元数据。
- 排名卡片展示。
- 图片展示。
- 电影详情链接使用 movie name / `movie_href`。

本轮验收口径：
- 历史只读基准 `/moviecarddayrank` 与 v2 `/moviecarddayrank` 截图对比。
- 历史只读基准 `/moviecardperiodrank` 与 v2 `/moviecardperiodrank` 截图对比。
- API 输出数量、排序、字段完整性、导航参数和分页参数符合页面要求。
- 查询链路核对索引、排序、分页和执行计划风险，不新增隐式阈值或硬过滤。
- v2 控制台无 warn/error，首屏图片加载成功。
- 执行 `cd v2/gc-api && go test ./...`。
- 执行 `cd v2/gc-web && npm run test`。
- 执行 `cd v2/gc-web && npm run build`。
- 执行 `make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify`。
- 若本轮启动 2040/2041 临时服务，验证结束必须关闭并记录端口清理结果。

本轮验收结果：
- `/moviecarddayrank` 已接入 v2 Vue 页面、v2 API、v2 路由与导航入口。
- `/moviecardperiodrank` 已接入 v2 Vue 页面、v2 API、v2 路由与导航入口。
- 历史只读截图已保存：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-moviecarddayrank.png`、`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/legacy-moviecardperiodrank.png`。
- v2 截图已保存：`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-moviecarddayrank.png`、`/Users/gaojinwei/src/my/rudy_gc/v2/tmp/v2-moviecardperiodrank.png`。
- Day Rank 交互验收：分页到第 2 页后首个排名为 `#19`；前一天切换到 `2026-05-25`。
- Period Rank 交互验收：分页到第 2 页后首个排名为 `#19`；上一期切换到 `Month 2026-04`。
- 浏览器控制台验收：v2 两个页面无 warn/error。
- 网络验收：rank period 交互请求均为 200。
- 查询验收：随机 period 查询已走 `idx_period_range`，EXPLAIN 为 `Using where; Backward index scan`，无 `Using filesort`。
- 验证命令：`make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify` 通过。
- 端口清理：`2040` Web 与 `2041` API 均已无监听。
- 本轮实现只使用 v2 自有前端、类型、API、样式和契约脚本。
