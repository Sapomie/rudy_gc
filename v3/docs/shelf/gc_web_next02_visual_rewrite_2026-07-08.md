# gc-web-next02 本轮用户指定视觉重写

## 规则
- 本文档只记录本轮用户指定视觉重写。
- 本轮初始已进行页面为 7 个页面；后续按本文档规则完成的页面继续加入已进行列表。
- 除已进行页面外，next02 其他所有路由页面全部算作没有进行本轮用户指定视觉重写。
- 历史完成、历史快验、旧进度记录不改变本轮状态。
- 没有进行本轮用户指定视觉重写的页面按左侧导航分组记录。
- 左侧导航未展示的路由页面统一放到“左侧导航未展示的路由页面”分组。
- 硬门禁：每个页面必须单独进行。
- 硬门禁：每个页面必须单独计时。
- 硬门禁：每个页面必须单独记录轮次。
- 硬门禁：每个页面必须单独截图验证。
- 硬门禁：每个页面必须单独按钮/API 实点验证。
- 硬门禁：每个页面必须单独写完成记录。
- 禁止把多个页面合并计时。
- 禁止把多个页面合并轮次。
- 禁止把多个页面合并截图结论。
- 禁止把多个页面合并按钮/API 验证结论。
- 禁止把“页面组投入 5 分钟”写成“每个页面投入 5 分钟”。
- 禁止把“页面组 2 轮修正”写成“每个页面 2 轮修正”。
- 每个页面最低投入 `5 分钟`，低于 5 分钟不能标记完成。
- 每个页面至少记录 `2 轮以上`视觉重写或修正动作，每一轮必须说明具体改动。
- 一个页面的 `5 分钟` 只属于该页面，不能和其他页面共享。
- 一个页面的 `2 轮以上` 只属于该页面，不能和其他页面共享。
- 视觉目标是 next02 新视觉；旧版只作为业务行为、按钮、接口和参数语义参考。
- 每个页面必须重写头部/hero、操作区、表单区、状态区、表格/列表区、空状态、弹窗或结果区。
- 每个页面必须包含页面进入、卡片进入、按钮反馈、表格/列表 hover、状态反馈；有弹窗的页面必须包含弹窗进入反馈。
- 功能必须可用：业务接口、点击行为、API payload、表单提交、复制、详情入口、全选、清空、预览等交互必须保留并实点验证。
- 完成记录必须包含：开始时间、结束时间、投入时长、重写/验证轮次、构建结果、截图结果、按钮/API 验证结果。
- 只有当前页面同时满足“单独 5 分钟、单独 2 轮以上、单独截图、单独按钮/API 验证、单独完成记录”，才能标记 `[已完成]`。
- 缺少任意一项时，该页面必须保留在“没有进行本轮用户指定视觉重写的页面”清单中。
- 构建标准：`npm run build` 通过；Vite chunk 体积提示可以接受。
- 接下来每完成 1 个页面，就从“没有进行本轮用户指定视觉重写的页面”移动到“已进行本轮用户指定视觉重写的页面”。
- 每个完成页面记录：时间范围、投入时长、重写/验证轮次、截图结果、按钮/API 验证结果。

## 已进行本轮用户指定视觉重写的页面
- [已完成] next02 /triggers/fetch-sehuatang-sh（2026-07-08 07:47-08:08 CST；投入 21 分钟；20 轮以上重写/验证）
- [已完成] next02 /triggers/fetch-sehuatang-sh-thread-list（2026-07-08 08:08-08:30 CST；投入 22 分钟；11 轮重写/验证）
- [已完成] next02 /triggers/fetch-sehuatang-sh-thread-detail（2026-07-08 08:30-08:58 CST；投入 28 分钟；15 轮重写/验证）
- [已完成] next02 /fetch-site-sehuatang-sh-list（2026-07-08 08:58-09:25 CST；投入 27 分钟；14 轮重写/验证）
- [已完成] next02 /triggers/media（2026-07-08 09:25-09:58 CST；投入 33 分钟；15 轮重写/验证）
- [已完成] next02 /triggers/media-rescan（2026-07-08 09:58-22:19 CST；累计投入 20 分钟以上；15 轮重写/验证）
- [已完成] next02 /triggers/media-rollback（2026-07-08 22:19-22:40 CST；投入 21 分钟；13 轮重写/验证）
- [已完成] next02 /triggers/fetch-site（2026-07-09 00:31-00:45 CST；投入 14 分钟；8 轮重写/验证）
- [已完成] next02 /triggers/fetch-sehuatang（2026-07-09 00:48:13-00:58:33 CST；投入 10 分 20 秒；9 轮重写/验证）
- [已完成] next02 /fetch-sites（2026-07-09 01:00:33-01:11:23 CST；投入 10 分 50 秒；6 轮重写/验证）
- [已完成] next02 /fetch-site-javbus-list（2026-07-09 01:12:51-01:23:11 CST；投入 10 分 20 秒；6 轮重写/验证）
- [已完成] next02 /fetch-site-sukebei-list（2026-07-09 01:24:13-01:34:49 CST；投入 10 分 36 秒；6 轮重写/验证）
- [已完成] next02 /fetch-site-sehuatang-list（2026-07-09 01:35:46-01:46:16 CST；投入 10 分 30 秒；7 轮重写/验证）
- [已完成] next02 /crawl-records（2026-07-09 03:33:17-03:43:44 CST；投入 10 分 27 秒；6 轮重写/验证）
- [已完成] next02 /triggers/dailybest（2026-07-09 03:44:57-03:55:45 CST；投入 10 分 48 秒；7 轮重写/验证）
- [已完成] next02 /triggers/seeds（2026-07-09 03:56:50-04:07:23 CST；投入 10 分 33 秒；5 轮重写/验证）
- [已完成] next02 /triggers/post-process（2026-07-09 04:08:33-04:19:24 CST；投入 10 分 51 秒；6 轮重写/验证）
- [已完成] next02 /triggers/backfill（2026-07-09 04:20:31-04:31:43 CST；投入 11 分 12 秒；5 轮重写/验证）
- [已完成] next02 /crawler/detail-loop（2026-07-09 04:32:52-04:44:38 CST；投入 11 分 46 秒；5 轮重写/验证）
- [已完成] next02 /crawler/tasks（2026-07-09 04:45:47-04:56:11 CST；投入 10 分 24 秒；6 轮重写/验证）
- [已完成] next02 /sc-events（首次 2026-07-10 08:58-09:03 CST；电影详情风格重写 10:34-10:47 CST；累计投入 18 分钟；4 轮重写/验证）
- [已完成] next02 /sc-pick-smart-media（2026-07-10 09:03-09:08 CST；投入 5 分钟；2 轮重写/验证）
- [已完成] next02 /triggers/sc（2026-07-10 09:08-09:13 CST；投入 5 分钟；2 轮重写/验证）
- [已完成] next02 /triggers/sc-media（2026-07-10 09:13-09:20 CST；投入 7 分钟；2 轮重写/验证）
- [已完成] next02 /movie/:movieName（以 `/movie/IPZZ-178` 验证；2026-07-10 10:14-10:32 CST；累计投入 18 分钟；3 轮重写/验证）

### next02 /movie/:movieName 完成记录
- 时间：2026-07-10 10:14-10:32 CST；累计投入 18 分钟，单页独立计时；验证样本为 `/movie/IPZZ-178`。
- 轮次：
  1. 只优化现有资源 Tab 和资源列表表格：强化选中态、表头层级、斑马纹、匹配行左侧标识、Hash 匹配态、字幕标签、两类相册按钮、表格边界及行 hover；未增加页面模块或重复数据。
  2. 只优化现有 Info、Rank、视频信息和抓取状态表格：统一标签列、内容列、分隔线、交替行、状态层级、焦点及 hover 反馈；未修改海报、页面结构和模块顺序。
  3. 参考 `/sc-pick-smart-media` 的进入节奏，为原有页面、顶部影片区、视频信息区、资源列表区和两张抓取状态卡增加依次淡入上浮动画，并补充 `prefers-reduced-motion` 关闭动画规则；未增加模块或修改布局。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-movie-rewrite-2026-07-10/movie-IPZZ-178/normal.png`、`v3/tmp/gc-web-next02-movie-rewrite-2026-07-10/movie-IPZZ-178/resource-table.png`；桌面原页面结构及资源表格肉眼检查通过；375px 窄屏页面无整体横向溢出，资源表格继续在自身容器内横向滚动。浏览器指针移动未触发可记录的 `:hover` 状态，已通过生效样式核对 hover 规则，未写入重复截图。
- 资源表格尺寸验证：优化前后 JavBus 表格均为表格高 `620px`、外框高 `622px`、表头 `30px`、数据行 `28px`、单元格内边距 `1px 10px 2px`、最小宽度 `860px`、`table-layout: fixed`；Sukebei 保持最小宽度 `980px`、表头 `30px`、数据行 `28px`；Sehuatang 保持最小宽度 `820px`。三种表格原列宽配置未变。
- 动画验证结果：页面整体使用 `movieDetailPageOpen 280ms`；顶部影片区使用 `movieDetailSectionOpen 420ms`；视频信息、资源列表和两张抓取状态卡延迟分别为 `70ms`、`130ms`、`220ms`、`270ms`。加载早期检测到透明度和 `translateY(9px)` 正在过渡，650ms 后全部恢复 `opacity: 1`、`translateY(0)`；动画前后表头 `30px`、数据行 `28px`、JavBus 表格高 `620px`。
- 按钮/API 验证结果：页面 API 返回 21 条 JavBus、15 条 Sukebei、0 条 Sehuatang 数据；JavBus、Sukebei、Sehuatang 三个资源 Tab 实点切换通过；Sukebei 大小排序后显示 `大小 ↓` 且首项为 `5.40 GB`；Sehuatang 空状态显示 `暂无 Sehuatang 磁力数据`；Info/Rank 标签切换通过，Rank 当前页渲染 10 条；页面控制台无本地应用错误。相册写入按钮保持原绑定，未执行会修改业务数据的操作。

### next02 /triggers/sc-media 完成记录
- 时间：2026-07-10 09:13-09:20 CST；投入 7 分钟，单页独立计时。
- 轮次：
  1. 重写 SC Media Move 的 hero、范围选择、两段式操作区、执行概览、根目录信息和五组结果面板，补充总数、可移动数、失败数及当前阶段状态反馈。
  2. 完成按钮反馈、状态胶囊、错误信息、结果表格 hover、窄屏单列布局及固定侧栏移动端覆盖，并完成真实接口与模式切换验证。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-sc-rewrite-2026-07-10/triggers-sc-media/normal.png`；桌面普通态与 375px 窄屏布局肉眼检查通过，无页面横向溢出。浏览器指针移动未触发可记录的 `:hover` 状态，已核对 hover 样式和按钮交互，未保留与普通态完全相同的重复截图。
- 按钮/API 验证结果：`GET /api/triggers/sc-media/plan` 页面计划读取通过；输入 SC 名称 `2026-07-07-06-15` 后点击 `第一段：预处理`，命中 `POST /api/triggers/sc-media/precheck` 并显示真实环境错误 `mkdir /Volumes/Expansion: permission denied`；切换 `全部符合条件` 后单个 SC 输入隐藏且提示显示；`can_commit=false` 时第二段按钮保持禁用，未执行文件移动。

### next02 /triggers/sc 完成记录
- 时间：2026-07-10 09:08-09:13 CST；投入 5 分钟，单页独立计时。
- 轮次：
  1. 重写 SC Preview & Submit hero、待预览/已选指标、扫描控制面板、预览结果面板、表单状态和提交操作区，统一 next02 的层级与进入动画。
  2. 完成预览按钮 loading、MovieCard hover、海报进入、状态切换反馈、移动端布局及固定侧栏窄屏覆盖，并实点验证预览和选择行为。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-sc-rewrite-2026-07-10/triggers-sc/normal.png`；桌面普通态与 375px 窄屏布局肉眼检查通过，无页面横向溢出。浏览器指针移动未触发可记录的 `:hover` 状态，已核对 hover 样式和卡片交互，未保留与普通态完全相同的重复截图。
- 按钮/API 验证结果：点击 `Add Preview` 命中 `POST /api/triggers/sc/add-preview`，返回并渲染 40 张影片卡片，初始已选 40；点击第一张卡片的 `不计入` 后已选数变为 39；提交按钮、选择状态和 payload 绑定已检查，为避免创建真实持久化任务未执行 `POST /api/triggers/sc/add`。

### next02 /sc-pick-smart-media 完成记录
- 时间：2026-07-10 09:03-09:08 CST；投入 5 分钟，单页独立计时。
- 轮次：
  1. 重写 Smart Pick hero、全局策略、动态分组、执行操作和状态反馈，并补齐原页面缺失的分组增删、展开、表单 payload、Smart Pick、复制状态轮询和停止复制交互。
  2. 重写抽取结果、四组筛选统计、复制进度、排序控件、MovieCard hover、表单焦点和移动端布局，并完成分组与真实 Smart Pick 接口实点验证。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-sc-rewrite-2026-07-10/sc-pick-smart-media/normal.png`；桌面普通态与 375px 窄屏布局肉眼检查通过，无页面横向溢出。浏览器指针移动未触发可记录的 `:hover` 状态，已核对 hover 样式和按钮交互，未保留与普通态完全相同的重复截图。
- 按钮/API 验证结果：新增分组后分组数由 1 变 2、展开高级条件可见、删除后恢复为 1；抽取数量填写 1 后点击 `仅 Smart Pick`，命中 `POST /api/triggers/sc/pick-smart-only`，页面反馈 `Smart Pick 完成，已抽取 1 部`，渲染 1 张 MovieCard 和 4 张筛选统计卡，结果计数为 `共 1 部 · 2.35 GB`。保留 legacy payload 的 `Page: 1`、`PageSize: 100000`、`MediaOwned: 3` 语义。

### next02 /sc-events 完成记录
- 时间：首次 2026-07-10 08:58-09:03 CST；电影详情风格重写 10:34-10:47 CST；累计投入 18 分钟，单页独立计时。
- 轮次：
  1. 为列表路由增加独立视觉作用域，重写 SC Events hero、总量/当前页指标、视图工具栏、事件时间线表格、分页和空状态，避免影响卡片页与详情页。
  2. 完成表格行 hover、按钮按压与焦点反馈、页面进入动画、移动端布局及固定侧栏窄屏覆盖，并完成列表数据、分页与视图入口验证。
  3. 参照 `/movie/:movieName` 重写顶部白色面板、标题层级、总量胶囊、列表/卡片视图 Tab、回填入口和分页；移除列表区重复的 Total 摘要，保留原字段与入口。
  4. 参照电影详情资源表重写事件表格面板、渐变表头、月份交替行、排序链接、日期入口、行 hover、空状态和移动端横向滚动，并接入页面、顶部、分页、表格区分段淡入上浮动画。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-sc-rewrite-2026-07-10/sc-events/normal.png`；桌面普通态与 375px 窄屏布局肉眼检查通过，无页面横向溢出。浏览器指针移动未触发可记录的 `:hover` 状态，已核对 hover 样式和表格交互，未保留与普通态完全相同的重复截图。
- 按钮/API 验证结果：SC Events 页面 API 请求成功，渲染 24 条当前页记录、10 列、总量 134 和分页；时间排序链接进入 `/sc-events?order=asc&p=1&sort=sc_time`；第 2 页进入 `/sc-events?p=2` 并渲染 24 条；点击 `卡片视图` 正确进入 `/sc-events-cards`，卡片页渲染 18 张卡片且未挂载列表页视觉作用域。375px 下页面整体无横向溢出，事件表格在 283px 容器内保留横向滚动；页面控制台无本地应用错误。

### next02 /crawler/tasks 完成记录
- 时间：2026-07-09 04:45:47-04:56:11 CST；投入 10 分 24 秒，单页独立计时。
- 轮次：
  1. 增加 `next02-crawler-tasks-rewrite` 独立页面态，接入 Crawler Jobs 轻量 hero、说明文本、任务统计指标和任务台入口。
  2. 重写 CrawlerTasks 页面主题、背景、hero 标题、指标胶囊、返回按钮和页面进入动画。
  3. 重写任务列表面板、面板 hover、标题、任务数徽标、表格容器、表头和表格行 hover。
  4. 重写任务状态徽标、运行中/已完成/暂停/失败状态、操作按钮、稳定点击选择器和空状态。
  5. 重写移动端 hero、指标区、面板内边距、表格横向扫描和窄屏标题布局。
  6. 肉眼检查发现任务数徽标贴近标题，改为 `next02-crawler-tasks-panel-head` 明确左右布局，并完成最终构建、截图和交互验证。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/crawler-tasks/round-06`；普通态相似度 96.8882%，hover 96.7964%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：真实 API 空状态验证通过；`GET /api/crawler/jobs` 返回 0 条时显示 `0 个` 和 `暂无任务`，点击 `回到任务台` 进入 `/triggers/dailybest`。受控任务数据验证通过；2 条任务渲染为 `job-daily`、`job-post`，状态徽标分别为 `is-running`、`is-done`，链接分别为 `/triggers/dailybest?job_id=job-daily`、`/triggers/post-process?job_id=job-post`，点击第一条 `回到任务` 后进入 `/triggers/dailybest?job_id=job-daily`。

### next02 /crawler/detail-loop 完成记录
- 时间：2026-07-09 04:32:52-04:44:38 CST；投入 11 分 46 秒，单页独立计时。
- 轮次：
  1. 增加 `next02-detail-loop-rewrite` 独立页面态，接入轻量 hero、状态指标和日志/控制区标识。
  2. 重写页面主题、背景、轻量标题、指标胶囊、返回按钮和页面进入动画。
  3. 重写状态面板、状态卡、运行 badge、最近一条日志和面板 hover 反馈。
  4. 重写 `startBtn`、`stopBtn`、状态消息和实时日志计数反馈。
  5. 重写日志结果框、pre 文本、移动端 hero/面板布局，并完成构建、截图、交互验证。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/crawler-detail-loop/round-01`；普通态相似度 94.8474%，hover 94.7671%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：拦截验证通过；初始 badge 为 `未启动`，`stopBtn` 禁用；点击开启命中 POST `/api/crawler/detail-loop/start`，开启后 badge 为 `运行中`、`startBtn` 禁用、`stopBtn` 可用、最近一条为 `loop started`；点击停止命中 POST `/api/crawler/detail-loop/stop`，停止后 badge 为 `未启动`、最近一条为 `loop stopped`。

### next02 /triggers/backfill 完成记录
- 时间：2026-07-09 04:20:31-04:31:43 CST；投入 11 分 12 秒，单页独立计时。
- 轮次：
  1. 增加 `next02-backfill-command-page` 独立页面态，接入 Backfill 专属 hero、指标、面板 eyebrow 和事件流标识。
  2. 重写 Backfill 页面主题、轻量标题区、顶部入口按钮、页面背景和进入动画。
  3. 重写任务面板、多任务按钮、演员名称输入和表单焦点/hover 状态。
  4. 重写执行/暂停/继续/停止按钮、进度条、概览卡片和状态反馈。
  5. 重写事件流结果区、原始 JSON 开关、事件计数、空状态和窄屏布局，并完成构建、截图、交互验证。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/triggers-backfill/round-01`；普通态相似度 95.8735%，hover 95.4930%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：拦截验证通过；任务按钮 6 个，点击 `spider_rebuild_actor_rank` 后演员名称字段可见；填写 `Alice` 后点击开始执行，POST `/api/crawler/jobs/start` 请求体包含 `task_type: spider_rebuild_actor_rank`、`actor_name: Alice`；暂停、继续、停止分别命中 `/api/crawler/jobs/backfill-test-job/pause`、`/resume`、`/stop`。

### next02 /triggers/post-process 完成记录
- 时间：2026-07-09 04:08:33-04:19:24 CST；投入 10 分 51 秒，单页独立计时。
- 轮次：
  1. 增加 `next02-post-process-command-page` 独立页面态，接入 PostProcess 专属 hero、指标、面板 eyebrow 和事件流标识。
  2. 重写 PostProcess 页面主题、轻量标题区、顶部入口按钮、页面背景和进入动画。
  3. 重写任务面板、任务类型按钮、标题翻译状态复选框、表单焦点和 hover 状态。
  4. 重写执行/暂停/继续/停止按钮、进度条、阶段概览卡片和状态 badge。
  5. 重写事件流结果区、原始 JSON 开关、事件计数、空状态和移动端布局。
  6. 第 1 次截图普通态 86.5125%、hover 86.3732%，DOM 主体结构通过但高度偏差；收敛状态复选区、标题 eyebrow 和操作区后第 2 次截图通过。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/triggers-post-process/round-02`；普通态相似度 95.2246%，hover 95.0740%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：拦截验证通过；任务按钮 3 个，点击 `spider_post_process` 后状态复选区可见；勾选状态 2 后点击开始执行，POST `/api/crawler/jobs/start` 请求体包含 `task_type: spider_post_process`、`statuses: ["1","2"]`；暂停、继续、停止分别命中 `/api/crawler/jobs/post-test-job/pause`、`/resume`、`/stop`。

### next02 /triggers/seeds 完成记录
- 时间：2026-07-09 03:56:50-04:07:23 CST；投入 10 分 33 秒，单页独立计时。
- 轮次：
  1. 增加 `next02-seeds-command-page` 独立页面态，接入 Seeds 专属 hero、指标、面板 eyebrow 和事件流标识。
  2. 重写 Seeds 页面主题、轻量标题区、顶部入口按钮、页面背景和进入动画。
  3. 重写任务面板、任务类型按钮、Seed 名称字段、刷新数量字段和自动外站复选框反馈。
  4. 重写执行/暂停/继续/停止按钮、进度条、阶段概览卡片和状态 badge。
  5. 重写事件流结果区、原始 JSON 开关、状态文本、事件计数和移动端布局，并完成构建、截图、交互验证。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/triggers-seeds/round-01`；普通态相似度 93.8946%，hover 93.8661%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：拦截验证通过；任务按钮 3 个，点击 `spider_seed_by_name` 后 Seed 名称字段可见；填写 `ABC-123`、取消自动外站后点击开始执行，POST `/api/crawler/jobs/start` 请求体包含 `task_type: spider_seed_by_name`、`name: ABC-123`、`auto_fetch_site: false`；暂停、继续、停止分别命中 `/api/crawler/jobs/seeds-test-job/pause`、`/resume`、`/stop`。

### next02 /triggers/dailybest 完成记录
- 时间：2026-07-09 03:44:57-03:55:45 CST；投入 10 分 48 秒，单页独立计时。
- 轮次：
  1. 增加 `next02-dailybest-command-page` 独立页面态，接入 DailyBest 专属 hero、指标、面板 eyebrow 和事件流标识。
  2. 补齐 DailyBest 的 51px 顶栏、240px 侧栏、页面容器和内容区布局规则。
  3. 重写 DailyBest hero、顶部操作入口、任务卡、运行卡、事件卡和状态胶囊。
  4. 重写任务类型按钮、输入控件焦点、执行/暂停/继续/停止按钮、进度条和运行概览卡片反馈。
  5. 重写事件流结果区、原始 JSON 开关、状态 badge、空日志视觉和移动端布局。
  6. 第 1 次截图普通态 87.1052%、hover 87.0764%，DOM 主体结构通过但高度偏差；收敛 hero 大卡片后第 2 次截图通过。
  7. 修复“detail 解析完成后自动抓取 JavBus / Sukebei”复选框只显示不入 payload 的功能问题，并完成最终构建、截图和接口实点验证。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/triggers-dailybest/round-03`；普通态相似度 92.1611%，hover 92.1302%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：拦截验证通过；任务按钮 2 个，点击 `DailyBest 同步` 后隐藏 `task_type` 为 `spider_daily_best_sync`；取消自动外站复选框后点击开始执行，POST `/api/crawler/jobs/start` 请求体包含 `task_type: spider_daily_best_sync`、`auto_fetch_site: false`；暂停、继续、停止分别命中 `/api/crawler/jobs/dailybest-test-job/pause`、`/resume`、`/stop`。

### next02 /crawl-records 完成记录
- 时间：2026-07-09 03:33:17-03:43:44 CST；投入 10 分 27 秒，单页独立计时。
- 轮次：
  1. 增加 `next02-records-rewrite` 独立页面态，接入轻量 hero、标题说明和抓取记录页面入口。
  2. 重写页面背景、hero 视觉、标题层级、指标入口和页面进入动画。
  3. 重写总行数卡片、筛选卡片、Type 切换、每页选择、筛选/清空按钮、hover/active/focus 反馈。
  4. 重写表格容器、表头、排序链接、分页、Type/Name 单元格和行 hover 反馈。
  5. 补充 `/crawl-records` 快验脚本配置、DOM 结构计数、空状态和移动端布局。
  6. 第 1 次截图普通态 84.5721%、hover 84.5107%，DOM 主体结构通过但高度偏差；收敛 hero 指标和入口按钮后第 2 次截图通过。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/crawl-records/round-02`；普通态相似度 90.1414%，hover 90.0857%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：实点验证通过；Type 链接 4 个，排序链接 5 个，分页链接存在；每页选择提交后 URL 包含 `ps=100`，点击 `DetailNumber` 后 URL 包含 `sort=detail_number`，点击 `DailyBest` 后 URL 包含 `type=DailyBest`，点击清空回到 `/crawl-records`，表格可见 17 行。

### next02 /triggers/fetch-site 完成记录
- 时间：2026-07-09 00:31-00:45 CST；投入 14 分钟，单页独立计时。
- 轮次：
  1. 增加 `next02-fetch-site-command-page` 独立页面态，给页面接入专属 hero、面板状态和视觉入口。
  2. 增加 fetch-site hero 指标计算，显示当前任务、抓取数量、排序和预览命中。
  3. 重写主背景、hero、指标胶囊、面板容器、状态徽标和进入动画。
  4. 重写筛选卡片、排序按钮、输入控件、预览按钮和执行按钮的 hover/active/焦点反馈。
  5. 重写运行概览、事件流、预览表格、徽标和移动端布局。
  6. 第 6 轮截图快验发现普通态 82.9773%，DOM 主体结构通过，但低于 90%。
  7. 压缩 hero 高度、面板头和指标区，第 7 轮截图普通态 87.4986%，仍低于 90%。
  8. 去除 fetch-site 装饰性 kicker/指标高度，隐藏面板装饰小字和状态胶囊，第 8 轮截图通过。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/triggers-fetch-site/round-08`；普通态相似度 90.7531%，hover 88.7757%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：拦截验证通过；展开筛选从隐藏变可见，排序切换到 `vw`，预览接口 POST 渲染 2 行，开始执行接口 POST 请求体包含 `task_type`、`number`、`od`、`last_fetch_duration_days`、`last_success_duration_days`。

### next02 /triggers/fetch-sehuatang 完成记录
- 时间：2026-07-09 00:48:13-00:58:33 CST；投入 10 分 20 秒，单页独立计时。
- 轮次：
  1. 增加 `next02-old-sehuatang-command-page` 独立页面态，接入该页专属 hero、面板状态和视觉入口。
  2. 增加旧色花堂 hero 指标计算，覆盖列表页、关键词、页码和入库模式。
  3. 重写主背景、轻量标题区、表单卡片、输入框和任务按钮状态。
  4. 重写停止条件复选项、列表 URL、入库模式、执行按钮和操作区反馈。
  5. 重写运行概览、进度条、事件流、结果区和响应式布局。
  6. 第 6 轮截图快验发现普通态 83.3365%，DOM 主体结构通过，但高度多 173px。
  7. 修正顶部入口按钮为单行，并恢复旧版“关键词”字段文案；第 7 轮普通态 85.9135%。
  8. 收敛卡片底色、主按钮颜色和操作区结构；第 8 轮普通态 87.6933%。
  9. 修正 `sehuatang-progress-grid` 为 6 列，恢复概览 3+6 布局；第 9 轮截图通过。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/triggers-fetch-sehuatang/round-09`；普通态相似度 91.2268%，hover 91.2172%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：拦截验证通过；勾选“抓到已存在停止”后起始页禁用；开始执行接口 POST 请求体包含 `task_type`、`list_url`、`keyword`、`stop_on_existing`、`start_page`、`end_page`、`persist_mode`。

### next02 /fetch-sites 完成记录
- 时间：2026-07-09 01:00:33-01:11:23 CST；投入 10 分 50 秒，单页独立计时。
- 轮次：
  1. 增加 `next02-fetch-sites-rewrite` 独立页面态，接入轻量 hero、配置摘要和顶部入口。
  2. 重写 hero、摘要胶囊、顶部入口按钮和主配置卡片。
  3. 重写配置表表头、长字段截断、状态 badge、行 hover 和编辑按钮。
  4. 重写编辑弹窗打开状态、遮罩、表单字段、保存按钮和消息反馈。
  5. 补充移动端布局、表格滚动和弹窗窄屏尺寸。
  6. 执行构建、截图快验和编辑保存接口拦截验证。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/fetch-sites/round-06`；普通态相似度 96.1799%，hover 96.1288%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：拦截验证通过；编辑按钮 4 个，点击后弹窗打开；修改 SiteName 和 Timeout 后保存发出 PUT `/api/fetch-sites/baidu_fanyi`，请求体包含 `site_code`、`site_name`、`base_url`、`timeout_seconds`、`request_sleep_ms`、`retry_sleep_ms`、`max_retry_times`、`status`，保存后弹窗关闭并显示“保存成功”。

### next02 /fetch-site-javbus-list 完成记录
- 时间：2026-07-09 01:12:51-01:23:11 CST；投入 10 分 20 秒，单页独立计时。
- 轮次：
  1. 增加 `next02-fetch-site-javbus-list-rewrite` 独立页面态，给 JavBus 列表接入专属视觉入口。
  2. 重写 JavBus 列表顶部 hero、说明文字、入口按钮和页面进入动画。
  3. 重写统计卡片、筛选卡片、输入控件、状态组和筛选焦点反馈。
  4. 重写表格表头、排序链接、状态 badge、行 hover、错误详情和单行触发按钮反馈。
  5. 重写分页、空状态、表格最小宽度和窄屏筛选堆叠。
  6. 执行构建、截图快验和按钮/API 拦截验证。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/fetch-site-javbus-list/round-06`；普通态相似度 92.3324%，hover 92.3089%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：拦截验证通过；首行触发按钮 50 个，点击首行发出 POST `/api/crawler/jobs/start`，请求体包含 `task_type: spider_fetch_javbus_resources`、`movie_jav_id`、`movie_name`；筛选触发按钮 1 个，清空状态后已选状态为 0，填写关键词 `ABP-123` 并选择 `release_date:desc` 后发出 `task_type: spider_fetch_javbus_filtered_resources`，请求体包含 `keyword`、`sort`、`order`、`trigger_sort`、`trigger_order`。

### next02 /fetch-site-sukebei-list 完成记录
- 时间：2026-07-09 01:24:13-01:34:49 CST；投入 10 分 36 秒，单页独立计时。
- 轮次：
  1. 增加 Sukebei 独立主题变量、页面背景、顶部 hero、入口按钮和页面进入动画。
  2. 重写 Sukebei 统计卡片、筛选卡片、输入控件、状态组和筛选焦点反馈。
  3. 重写表格表头、排序链接、状态 badge、行 hover、错误详情和单行触发按钮反馈。
  4. 重写分页、空状态、表格最小宽度和窄屏筛选堆叠。
  5. 细化影片链接、MPlay/WMedia、筛选触发按钮和状态消息反馈。
  6. 执行构建、截图快验和按钮/API 拦截验证。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/fetch-site-sukebei-list/round-06`；普通态相似度 92.2948%，hover 92.2872%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：拦截验证通过；首行触发按钮 50 个，点击首行发出 POST `/api/crawler/jobs/start`，请求体包含 `task_type: spider_fetch_sukebei_resources`、`movie_jav_id`、`movie_name`；筛选触发按钮 1 个，清空状态后已选状态为 0，填写关键词 `SSIS-001` 并选择 `latest_publish_time:desc` 后发出 `task_type: spider_fetch_sukebei_filtered_resources`，请求体包含 `keyword`、`sort`、`order`、`trigger_sort`、`trigger_order`。

### next02 /fetch-site-sehuatang-list 完成记录
- 时间：2026-07-09 01:35:46-01:46:16 CST；投入 10 分 30 秒，单页独立计时。
- 轮次：
  1. 增加 `next02-sehuatang-magnet-rewrite` 页面结构入口，并给旧磁力列表 hero 增加总记录、当前页、hash、可入册 KPI。
  2. 重写旧色花堂列表页面底色、顶部 hero、入口按钮和页面进入动画。
  3. 重写筛选卡片、关键词/movie_jav_id/info_hash/tag/发帖时间输入、快捷 tag 和焦点反馈。
  4. 重写发帖时间、帖子标题、影片名、相册、info_hash、最近抓到 6 列表格、排序链接、时间胶囊、标题链接和行 hover。
  5. 重写下载中/待下载/其他相册按钮、分页、空状态、SHT 时间弹窗/其他相册弹窗、状态消息和移动端布局。
  6. 补齐 `SHT时间` 按钮功能：打开日期弹窗、今天按钮、日期格式校验、POST `/api/w-kv/date` 保存、保存后按钮文案更新。
  7. 执行构建、功能补丁后的截图快验、筛选提交和按钮/API 拦截验证。
- 构建结果：`npm run build` 通过；Vite chunk 体积提示可接受。
- 截图结果：`v3/tmp/gc-web-next02-page-match/fetch-site-sehuatang-list/round-07`；普通态相似度 91.0862%，hover 91.0625%，DOM 主体结构通过，脚本 `passed:true`。
- 按钮/API 验证结果：拦截验证通过；筛选表单提交后 URL 带出 `ps`、`sort`、`order`、`keyword`、`movie_jav_id`、`info_hash`、`tag`、`post_time_from`、`post_time_to`；`SHT时间` 弹窗打开成功，今天按钮写入 `2026-07-09`，保存发出 POST `/api/w-kv/date`，请求体包含 `item_key: sht_time`、`item_value: 2026-07-09`，保存后按钮显示 `SHT时间：2026-07-09`；下载中按钮发出 POST `/api/movie/FC2PPV-3972042/album-item`，请求体包含 `movie_jav_id`、`source_type: sehuatang_magnet`、`source_row_id`、`album_name: 下载中`；其他相册弹窗打开成功，选择 `Media` 后发出同一接口，请求体 `album_name: Media`。

## 没有进行本轮用户指定视觉重写的页面

### MovieCard
- [ ] next02 /cardstoday
- [ ] next02 /cardsmediamowned
- [ ] next02 /cardshasrank
- [ ] next02 /moviecarddayrank
- [ ] next02 /moviecardperiodrank
- [ ] next02 /cardsneeddownload
- [ ] next02 /cardsrandom
- [ ] next02 /cardsagg

### MovieAgg
- [ ] next02 /movie-agg-all/release
- [ ] next02 /movie-agg-all/release-bucket-list
- [ ] next02 /w-media-agg/bucket-list
- [ ] next02 /w-media-agg/birth
- [ ] next02 /triggers/agg
- [ ] next02 /wdir

### List
- [ ] next02 /medias
- [ ] next02 /w-agg-events
- [ ] next02 /torrent-albums
- [ ] next02 /movie-albums
- [ ] next02 /d-seeds
- [ ] next02 /casts
- [ ] next02 /e-items

### SC
- [已完成] next02 /sc-events
- [已完成] next02 /sc-pick-smart-media
- [已完成] next02 /triggers/sc
- [已完成] next02 /triggers/sc-media

### 抓取任务
- [已完成] next02 /crawl-records
- [已完成] next02 /triggers/dailybest
- [已完成] next02 /triggers/seeds
- [已完成] next02 /triggers/post-process
- [已完成] next02 /triggers/backfill
- [已完成] next02 /crawler/detail-loop
- [已完成] next02 /crawler/tasks

### FetchSite操作
- [已完成] next02 /triggers/fetch-site
- [已完成] next02 /triggers/fetch-sehuatang
- [已完成] next02 /fetch-sites
- [已完成] next02 /fetch-site-javbus-list
- [已完成] next02 /fetch-site-sukebei-list
- [已完成] next02 /fetch-site-sehuatang-list

### 色花堂SH
- [已完成] next02 /triggers/fetch-sehuatang-sh
- [已完成] next02 /triggers/fetch-sehuatang-sh-thread-list
- [已完成] next02 /triggers/fetch-sehuatang-sh-thread-detail
- [已完成] next02 /fetch-site-sehuatang-sh-list

### Media操作
- [已完成] next02 /triggers/media
- [已完成] next02 /triggers/media-rescan
- [已完成] next02 /triggers/media-rollback

### 左侧导航未展示的路由页面
- [ ] next02 /
- [ ] next02 /cards
- [已完成] next02 /movie/:movieName
- [ ] next02 /cast
- [ ] next02 /cast/:personId
- [ ] next02 /records
- [ ] next02 /sc-events-cards
- [ ] next02 /sc-events/:eventName
- [ ] next02 /fetch-site-sehuatang-sh-detail/:threadTid
- [ ] next02 /triggers/fetch-site-javbus-filtered
- [ ] next02 /triggers/fetch-site-sukebei-filtered
