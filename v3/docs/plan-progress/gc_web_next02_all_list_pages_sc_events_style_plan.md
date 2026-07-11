# gc-web-next02 全部列表页面对齐 `/sc-events` 改造计划

## 1. 计划目标

本计划用于将 `v3/gc-web-next02` 中所有以列表、表格或结果清单为主要内容的页面，统一改造成当前 `/sc-events` 的默认规格样式。

默认规范来源：

- `v3/docs/standard/gc_web_next02_frontend_style_standard.md`
- 当前 `/sc-events` 页面实际实现

本计划只统一视觉体系、页面密度、列表格式和页面入口行为，不改变页面原有业务字段、筛选参数、排序参数、分页语义、按钮行为、API、数据结构和持久化逻辑。

## 2. 总体原则

- 当前 `/sc-events` 是所有 next02 列表页面的默认视觉基准。
- 所有页面使用同一粉色主题、紧凑布局、表格密度、交互反馈和新窗口规则。
- 每个页面必须根据自身字段内容单独规划列宽。
- 禁止机械复制 `/sc-events` 的具体列宽比例。
- 禁止使用一个通用数据表格覆盖不同页面的业务结构。
- 禁止删除、合并或重命名既有字段、按钮、Tab、筛选和下钻入口。
- 禁止修改其他 next 版本。
- 页面改造必须逐页完成，发现同类问题时也按页面逐个处理。

## 3. 基准页面

### 3.1 基准路由

- `/sc-events`

### 3.2 基准规格

| 项目 | 默认规格 |
|---|---|
| 主题 | 粉色主题 |
| 页面顶部 | 紧凑单行 |
| 筛选位置 | 列表控制栏最左侧 |
| 统计位置 | 分页左侧 |
| 分页位置 | 控制栏最右侧 |
| 表格布局 | `table-layout: fixed` |
| 表格宽度 | `100%` |
| 表头高度 | `30px` |
| 数据行高度 | `28px` |
| 字号 | `13px` |
| 单元格内边距 | `1px 10px 2px` |
| 长文本 | 单行省略 |
| 页面横向滚动 | 禁止 |
| 页面入口 | 新窗口打开 |
| 当前页操作 | 当前窗口执行 |

## 4. 适用范围定义

### 4.1 主列表页面

页面主要内容为表格或清单时，整个页面按本计划改造。

### 4.2 嵌入式列表页面

页面同时包含详情、表单、任务控制或卡片时，只改造其中的列表、表格、结果清单和对应控制栏，不改变其他业务区域。

### 4.3 不属于本计划的页面

- 纯卡片网格页面
- 纯表单页面
- 纯任务控制页面
- 纯详情展示页面
- 实时日志页面
- 没有列表、表格或结果清单的页面

这些页面只有在其内部存在明确列表区域时，才处理该列表区域。

## 5. 主列表页面矩阵

### 5.1 List 导航组

| 状态 | 路由 | 页面文件 | 主要列表 |
|---|---|---|---|
| 交互检查通过 | `/medias` | `MediasPage.vue` | Media 列表 |
| 待处理 | `/w-agg-events` | `WAggEventsPage.vue` | W 聚合事件列表 |
| 交互检查通过 | `/torrent-albums` | `TorrentAlbumsPage.vue` | 种子专辑列表 |
| 交互检查通过 | `/movie-albums` | `MovieAlbumsPage.vue` | 影片专辑列表 |
| 待处理 | `/d-seeds` | `DSeedsPage.vue` | DSeed 列表 |
| 交互检查通过 | `/casts` | `CastsPage.vue` | 演员列表 |
| 待处理 | `/e-items` | `EItemsPage.vue` | EItem 列表 |

### 5.2 记录与配置列表

| 状态 | 路由 | 页面文件 | 主要列表 |
|---|---|---|---|
| 待处理 | `/crawl-records` | `RecordsPage.vue` | 抓取记录列表 |
| 待处理 | `/records` | `RecordsPage.vue` | 抓取记录兼容入口 |
| 待处理 | `/fetch-sites` | `FetchSitesPage.vue` | FetchSite 配置列表 |
| 待处理 | `/crawler/tasks` | `CrawlerTasksPage.vue` | 抓取任务列表 |

### 5.3 聚合桶列表

| 状态 | 路由 | 页面文件 | 主要列表 |
|---|---|---|---|
| 交互检查通过 | `/movie-agg-all/release-bucket-list` | `ReleaseBucketListPage.vue` | Release Bucket 列表 |
| 交互检查通过 | `/w-media-agg/bucket-list` | `ReleaseBucketListPage.vue` | WMedia Bucket 列表 |

### 5.4 FetchSite 数据列表

| 状态 | 路由 | 页面文件 | 主要列表 |
|---|---|---|---|
| 待处理 | `/fetch-site-javbus-list` | `ListPage.vue` | JavBus 抓取列表 |
| 待处理 | `/fetch-site-sukebei-list` | `ListPage.vue` | Sukebei 抓取列表 |
| 交互检查通过 | `/fetch-site-sehuatang-list` | `ListPage.vue` | 色花堂磁力列表 |
| 待处理 | `/fetch-site-sehuatang-sh-list` | `ListPage.vue` | 色花堂 SH 帖子列表 |
| 待处理 | `/fetch-site-sehuatang-sh-detail/:threadTid` | `ListPage.vue` | 色花堂 SH 详情资源列表 |

## 6. 嵌入式列表区域矩阵

以下页面只改造列表、表格或结果区，不重写页面其他业务模块。

| 状态 | 路由 | 页面文件 | 改造区域 | 保留区域 |
|---|---|---|---|---|
| 待处理 | `/cast` | `CastPage.vue` | Cast 列表表格 | 原筛选和入口语义 |
| 待处理 | `/cast/:personId` | `CastPage.vue` | 页面内表格/清单 | 演员详情结构 |
| 交互检查通过 | `/movie/:movieName` | `MoviePage.vue` | Info、Rank、视频信息、资源列表等表格的主题与交互 | 海报、模块顺序、资源表格既有尺寸 |
| 待处理 | `/triggers/media` | `MediaIngestPage.vue` | 执行结果表格 | 表单和任务操作 |
| 待处理 | `/triggers/media-rescan` | `MediaRescanPage.vue` | 执行结果表格 | 表单和任务操作 |
| 待处理 | `/triggers/media-rollback` | `MediaRollbackPage.vue` | 执行结果表格 | 表单和任务操作 |
| 待处理 | `/triggers/sc-media` | `ScMediaMovePage.vue` | 计划和执行结果列表 | 两阶段任务操作 |

## 7. TriggerPage 结果列表矩阵

`TriggerPage.vue` 服务多个路由。只统一结果表格、事件清单和相关控制栏，任务表单和业务按钮保持原样。

| 状态 | 路由 | 结果区处理 |
|---|---|---|
| 待处理 | `/triggers/dailybest` | 任务结果和事件列表 |
| 待处理 | `/triggers/seeds` | 任务结果和事件列表 |
| 待处理 | `/triggers/post-process` | 任务结果和事件列表 |
| 待处理 | `/triggers/backfill` | 任务结果和事件列表 |
| 待处理 | `/triggers/fetch-site` | 预览和执行结果列表 |
| 待处理 | `/triggers/fetch-sehuatang` | 预览和执行结果列表 |
| 待处理 | `/triggers/fetch-sehuatang-sh` | 预览和执行结果列表 |
| 待处理 | `/triggers/fetch-sehuatang-sh-thread-list` | 帖子列表结果 |
| 待处理 | `/triggers/fetch-sehuatang-sh-thread-detail` | 详情结果列表 |
| 待处理 | `/triggers/fetch-site-javbus-filtered` | 筛选结果列表 |
| 待处理 | `/triggers/fetch-site-sukebei-filtered` | 筛选结果列表 |

## 8. 明确排除项

以下页面默认不纳入本次“列表规格改造”主体范围：

- `/cards`、`/cardstoday`、`/cardshasrank`、`/cardsmediamowned`、`/cardsneeddownload`、`/cardsrandom`
- `/moviecarddayrank`、`/moviecardperiodrank`
- `/sc-events-cards`
- `/sc-pick-smart-media`
- `/wdir` 的 MovieCard 网格和目录卡片
- `/crawler/detail-loop` 实时状态与日志
- 纯任务表单和纯详情展示模块

排除原因：这些页面的主要内容是卡片、目录、表单、状态或日志，不应强制改造成表格列表。

`/sc-events-cards` 使用独立方案：

- `v3/docs/plan-progress/gc_web_next02_sc_events_cards_rewrite_plan.md`

## 9. 公共样式体系

### 9.1 公共作用域

列表页面统一增加页面级独立类，禁止使用无作用域全局覆盖。

建议公共类：

```text
next02-pink-list-page
next02-pink-list-hero
next02-pink-list-toolbar
next02-pink-list-panel
next02-pink-list-table-wrap
next02-pink-list-table
next02-pink-list-metrics
next02-pink-list-pagination
```

页面仍需保留自身类名处理独有字段、列宽和业务状态。

### 9.2 默认颜色

| 用途 | 色值 |
|---|---|
| 主粉色 | `#d94f8a` |
| 深粉链接 | `#9f2f63` |
| 主文字 | `#4b2939` |
| 次要文字 | `#8b6577` |
| 主浅粉背景 | `#fff0f6` |
| 次浅粉背景 | `#fff5fa` |
| 表头背景 | `#fcecf4` |
| 行 hover | `#fde8f2` |
| 主边框 | `#eed5e1` |
| 表格分隔线 | `#f1dce6` |

## 10. 页面结构标准

### 10.1 紧凑顶部

- 页面标识、标题、说明和主要入口优先保持单行。
- 默认上下内边距 `9px`，左右内边距 `14px`。
- 标题默认字号 `18px`。
- 说明默认字号 `12px`，空间不足时单行省略。
- 禁止增加重复摘要模块和大面积留白。

### 10.2 列表控制栏

默认顺序：

1. 左侧筛选
2. 左侧排序或状态切换
3. 右侧统计胶囊
4. 最右侧分页

- 控件默认高度 `30px`。
- 控制栏空间不足时允许换行。
- 窄屏筛选控件可以纵向排列。

### 10.3 列表面板

- 背景使用白色到浅粉渐变。
- 面板圆角 `10px`。
- 内层表格容器圆角 `7px`。
- 使用低透明粉色阴影。
- 面板 hover 不得改变尺寸。

## 11. 表格标准

### 11.1 基础尺寸

| 项目 | 默认规格 |
|---|---|
| 宽度 | `100%` |
| 布局 | `table-layout: fixed` |
| 表头高度 | `30px` |
| 数据行高度 | `28px` |
| 字号 | `13px` |
| 内边距 | `1px 10px 2px` |
| 横向滚动 | 默认禁止 |

### 11.2 表头

- 背景使用 `#fcecf4`。
- 字重默认 `850`。
- 可排序表头显示指针光标。
- hover 使用 `#fde8f2`，底部显示 `2px` 主粉色线。

### 11.3 数据行

- 默认字重 `680`。
- 数字列使用 `tabular-nums` 和 `800` 字重。
- 行 hover 使用 `#fde8f2`。
- hover 不得改变行高。
- 状态色必须表达真实业务状态。

## 12. 列宽设计硬规则

- 每个页面必须先记录列名、典型内容和最长内容类型。
- 时间列必须能完整显示当前页面的时间格式。
- 两位数字、计数、页码和短状态列必须保持紧凑。
- ID、Hash、路径、名称、演员、标题、备注等长文本列优先获得空间。
- 所有列宽总和必须为 `100%`。
- 禁止自动等分全部列。
- 禁止机械复制 `/sc-events` 的列宽比例。
- 禁止使用不稳定的像素与百分比混合 `calc()`。
- 禁止使用会使目标列意外收缩的 `max-width: 0`。
- 长文本默认单行省略，并保留完整内容入口或 `title`。
- 页面整体不得出现横向滚动。
- 只有用户明确要求或业务字段确实无法在容器内表达时，才允许表格内部横向滚动。

## 13. 特殊页面边界

### 13.1 `/movie/:movieName`

- 资源表格保留既有高度、表头高度、行高、最小宽度和原列宽配置。
- 不得因统一主题改变海报尺寸、顶部布局、模块顺序或资源表格尺寸。
- 只统一颜色、边框、链接、hover、focus 和状态反馈。

### 13.2 `ListPage.vue`

- 一个组件服务多个路由，必须按 `ViewType` 或路由增加独立作用域。
- 每个路由保留自身字段、筛选、按钮、弹窗和任务入口。
- 禁止用一个列宽配置覆盖全部 ListPage 路由。

### 13.3 `TriggerPage.vue`

- 只统一结果列表和事件清单。
- 任务选择、表单字段、开始/暂停/继续/停止按钮保持原结构和行为。
- 空结果、运行中、失败和完成状态必须保留。

### 13.4 同组件多路由

- `/crawl-records` 与 `/records` 使用同一组件时，共享视觉实现。
- 两个 Release Bucket 路由使用同一组件时，保留各自标题、数据源和字段差异。
- 页面验收仍需按路由分别记录。

## 14. 页面入口行为

### 14.1 新窗口打开

所有进入其他页面的入口必须同时设置：

```html
target="_blank"
rel="noopener noreferrer"
```

适用范围包括：

- 详情页
- Movie 页面
- Cast 页面
- 任务页面
- 独立列表或卡片视图
- 其他路由入口

### 14.2 当前窗口操作

以下操作继续在当前窗口执行：

- 筛选
- 排序
- 分页
- 清空
- 当前页面 Tab 或状态切换

## 15. 空状态与错误状态

- 空状态放在列表面板内部。
- 使用浅粉背景、粉色虚线边框和次要文字。
- 空状态不得增加大面积插画和空白。
- 错误状态保留真实错误信息。
- 加载、空、错误、完成状态不得互相覆盖。

## 16. 动画与反馈

- 页面整体使用轻量淡入。
- 顶部和列表面板可使用小幅上浮。
- hover 动画控制在 `150ms` 至 `180ms`。
- 排序、分页、筛选和按钮必须有清晰反馈。
- 必须提供 `prefers-reduced-motion: reduce` 覆盖。

## 17. 实施阶段

### 阶段一：建立共享样式

- 提取粉色列表主题变量。
- 建立顶部、控制栏、面板、表格、统计和分页公共类。
- 确保公共类不会覆盖卡片页、详情页和其他 next 版本。

### 阶段二：改造 List 导航组

- `/medias`
- `/w-agg-events`
- `/torrent-albums`
- `/movie-albums`
- `/d-seeds`
- `/casts`
- `/e-items`

### 阶段三：改造记录、配置和桶列表

- `/crawl-records`
- `/records`
- `/fetch-sites`
- `/crawler/tasks`
- `/movie-agg-all/release-bucket-list`
- `/w-media-agg/bucket-list`

### 阶段四：改造 FetchSite 数据列表

- `/fetch-site-javbus-list`
- `/fetch-site-sukebei-list`
- `/fetch-site-sehuatang-list`
- `/fetch-site-sehuatang-sh-list`
- `/fetch-site-sehuatang-sh-detail/:threadTid`

### 阶段五：改造嵌入式列表

- Cast 页面列表区
- Movie 页面表格区
- Media 操作结果区
- SC Media 结果区

### 阶段六：改造 TriggerPage 结果区

- 按路由逐页处理结果列表和事件清单。
- 不改变任务表单和任务控制逻辑。

### 阶段七：全量检查

- 检查公共样式作用域。
- 检查各页面列宽。
- 检查筛选、排序和分页参数。
- 检查新窗口入口。
- 执行构建、类型检查和差异检查。

## 18. 单页实施步骤

每个页面固定执行：

1. 读取当前页面模板、数据字段、筛选、排序、分页和入口。
2. 记录全部列名、典型内容和长内容列。
3. 规划总和为 `100%` 的独立列宽。
4. 接入粉色主题和紧凑顶部。
5. 将筛选放左、统计和分页放右。
6. 对齐表头、行高、字号、边框和 hover。
7. 补齐长文本省略与完整内容入口。
8. 补齐页面入口新窗口属性。
9. 检查空状态、错误状态和移动端布局。
10. 执行类型检查、构建和差异检查。

## 19. 单页记录模板

```md
### 页面：/example

- 页面文件：
- 主要列表：
- 筛选参数：
- 排序参数：
- 分页参数：
- 详情入口：
- 列宽方案：
- 特殊尺寸保留项：
- 新窗口入口：
- 构建结果：
- 交互检查：
- 当前状态：待处理 / 代码完成 / 构建通过 / 交互检查通过
```

### 页面：`/casts`

- 页面文件：`v3/gc-web-next02/src/pages/CastsPage.vue`
- 主要列表：演员统计列表
- 筛选参数：`keyword`、`wowned_min`、`wowned_max`、`sc_times_min`、`sc_times_max`、`come_times_min`、`come_times_max`、`last_sc_from`、`last_sc_to`
- 排序参数：`sort`、`order`
- 分页参数：`p`、`ps`
- 详情入口：`/cast/:personId`
- 列宽方案：Name `17%`、Chinese `16%`、Age `5%`、Height `7%`、Movies `7%`、WOwned `8%`、SC Times `8%`、Come Times `9%`、Last SC `13%`、Highest Rank `10%`
- 特殊尺寸保留项：无
- 新窗口入口：演员详情链接已设置 `target="_blank"` 和 `rel="noopener noreferrer"`
- 构建结果：`make check-web-next02` 通过；`npm --prefix gc-web-next02 run build` 通过
- 交互检查：桌面端和移动端均无页面或表格横向溢出；表头 `30px`、数据行 `28px`、字号 `13px`；筛选和分页响应式换行正常；浏览器控制台无警告或错误
- 当前状态：交互检查通过

### 页面：`/medias`

- 页面文件：`v3/gc-web-next02/src/pages/MediasPage.vue`
- 主要列表：Media 基础数据、演员、播放信息和媒体探测操作列表
- 筛选参数：`keyword`、`size_min`、`size_max`、`height_min`、`height_max`、`duration_min`、`duration_max`、`bit_rate_min`、`bit_rate_max`、`frame_average_min`、`frame_average_max`、`self_make`、`has_mask`、`sc_times_min`、`sc_times_max`、`last_sc_from`、`last_sc_to`、`birth_from`、`birth_to`、`releasing_from`、`releasing_to`
- 排序参数：`sort`、`order`
- 分页参数：`p`、`ps`
- 详情入口：`/movie/:movieName`、`/cast/:personId` 和媒体播放地址
- 列宽方案：识别码 `9%`、演员 `10%`、大小 `5%`、高度 `4%`、时长 `5%`、码率 `6%`、平均帧率 `5.5%`、播放 `5%`、自制 `4%`、遮罩 `4%`、SC `3%`、最后 SC `10.5%`、下载时间 `10.5%`、发行时间 `10.5%`、操作 `8%`
- 特殊尺寸保留项：`scan_meta` 按钮保持 `24px` 紧凑高度，确保数据行维持 `28px`
- 新窗口入口：Movie、Cast 和播放链接已设置 `target="_blank"` 和 `rel="noopener noreferrer"`
- 构建结果：`make check-web-next02` 通过；`npm --prefix gc-web-next02 run build` 通过
- 交互检查：桌面端和移动端均无页面或表格横向溢出；基础筛选与时间筛选采用明确两行分组，时间筛选、操作按钮、统计和分页保持同一基线；三个日期列完整显示；表头 `30px`、数据行 `28px`、字号 `13px`；全部筛选参数保留；浏览器控制台无警告或错误；`scan_meta` 写操作链保持原实现
- 当前状态：交互检查通过

### 页面：`/torrent-albums`

- 页面文件：`v3/gc-web-next02/src/pages/TorrentAlbumsPage.vue`
- 主要列表：种子相册中的影片、来源、Info Hash、大小和时间信息列表
- 筛选参数：`album_name`、`keyword`、`source_type`、`info_hash`、`publish_from`、`publish_to`、`created_from`、`created_to`
- 排序参数：`sort`、`order`；支持影片、来源、发布日期和下载中时间排序
- 分页参数：`p`、`ps`
- 详情入口：`/movie/:movieName`
- 列宽方案：选择 `3%`、影片 `12%`、来源 `8%`、Info Hash `28%`、大小 `8%`、发布日期 `10%`、下载中时间 `15%`、操作 `16%`
- 特殊尺寸保留项：Info Hash 保持单行省略；复制 Hash、移除按钮使用 `23px` 高度，确保数据行维持 `28px`
- 新窗口入口：Movie 链接已设置 `target="_blank"` 和 `rel="noopener noreferrer"`
- 构建结果：`make check-web-next02` 通过；`npm run build` 通过，保留现有构建产物体积提示
- 交互检查：使用 `album_name=Media` 验证 1608 条数据列表，并验证默认“下载中”空相册；桌面端 `1280px` 和窄屏 `390px` 均无页面或表格横向溢出；表头 `30px`、数据行 `28px`、字号 `13px`；8 列固定列宽总和为 `100%`；快捷相册、筛选、排序、分页和批量工具栏结构正常；新增相册弹窗使用粉色主题；浏览器控制台无警告或错误
- 当前状态：交互检查通过

### 页面：`/movie-albums`

- 页面文件：`v3/gc-web-next02/src/pages/MovieAlbumsPage.vue`
- 主要列表：电影相册中的影片和加入时间列表
- 筛选参数：`album_name`、`keyword`
- 排序参数：`sort`、`order`；支持影片和加入时间排序
- 分页参数：`p`、`ps`
- 详情入口：`/movie/:movieName`
- 列宽方案：选择 `4%`、影片 `60%`、加入时间 `24%`、操作 `12%`
- 特殊尺寸保留项：保留待下载电影、待删除、v3、vx 快捷相册；保留统一删除和移除已下载条目的条件操作及确认流程
- 新窗口入口：Movie 链接已设置 `target="_blank"` 和 `rel="noopener noreferrer"`
- 构建结果：`make check-web-next02` 通过；`npm run build` 通过，保留现有构建产物体积提示
- 交互检查：使用默认“待下载电影”相册验证 382 条数据列表，并使用“待删除”相册验证统一删除按钮的显示和禁用状态；桌面端 `1280px` 和窄屏 `390px` 均无页面或表格横向溢出；表头 `30px`、数据行 `28px`、字号 `13px`；4 列固定列宽总和为 `100%`；筛选、排序、分页、条件操作和新增电影相册弹窗结构正常；浏览器控制台无警告或错误
- 当前状态：交互检查通过

### 页面：`/movie/:movieName`

- 页面文件：`v3/gc-web-next02/src/pages/MoviePage.vue`
- 主要列表：Info 表格、Rank 表格、WMedia 视频信息、JavBus/Sukebei/Sehuatang 资源列表和抓取状态表格
- 筛选参数：无服务端筛选参数；资源 Tab 和 Info/Rank Tab 保持当前页切换
- 排序参数：Rank 的 `date`、`rank` 本地排序；资源表格的 `size`、`date`、`downloads` 本地排序
- 分页参数：Rank 本地页码和每页数量
- 详情入口：MovieCard、Cast、SC、电影专辑、每日 Rank、外部资源和媒体播放地址
- 列宽方案：Info 标签列保留 `96px`；JavBus 保留 `38% / 72px / 96px / 70px / 240px / 132px`；Sukebei 保留 `42% / 72px / 104px / 74px / 240px / 132px`；Sehuatang 保留 `48% / 116px / 240px / 132px`
- 特殊尺寸保留项：海报、顶部布局、模块顺序、Info/Rank 尺寸、资源表格行高、内边距、列宽及内部横向滚动规则完整保留
- 新窗口入口：MovieCard、Cast、SC、专辑、Rank、外部资源和播放链接已设置 `target="_blank"` 和 `rel="noopener noreferrer"`
- 构建结果：`make check-web-next02` 通过；`npm --prefix gc-web-next02 run build` 通过
- 交互检查：使用 `/movie/ABF-359` 完成桌面端和移动端检查，并使用 `/movie/SNOS-081` 验证资源 Tab；Info、Rank、视频信息、资源列表和抓取状态区域使用粉色主题；资源 Tab 内容使用 `180ms` 淡入上移动画；匹配资源行、左侧强调条和 Hash 胶囊使用粉色状态；页面整体无横向滚动；资源表格保持内部横向滚动；浏览器控制台无警告或错误；业务写入按钮保持原操作链
- 当前状态：交互检查通过

### 页面：`/w-media-agg/bucket-list`

- 页面文件：`v3/gc-web-next02/src/pages/ReleaseBucketListPage.vue`
- 主要列表：WMedia 创建时间聚合桶落库列表
- 筛选参数：`scope_key`、`year`、`quarter`、`month`、`ps`
- 排序参数：`sort`、`dir`
- 分页参数：`p`、`ps`
- 详情入口：`/w-media-agg/birth` 聚合页、`/triggers/agg` 回填页和聚合桶下钻页
- 列宽方案：时间桶 `22%`、媒体数 `12%`、已删除数 `12%`、总大小 `17%`、字幕数 `12%`、最新下载时间 `25%`
- 特殊尺寸保留项：WMedia 最新下载时间保留年月日和时分完整显示空间
- 新窗口入口：聚合页、回填页和时间桶下钻链接已设置 `target="_blank"` 和 `rel="noopener noreferrer"`
- 构建结果：`make check-web-next02` 通过；`npm run build` 通过，保留现有构建产物体积提示
- 交互检查：桌面端 `1280px` 和窄屏 `390px` 均无页面或表格横向溢出；表头 `30px`、数据行 `28px`、字号 `13px`；6 列固定列宽总和为 `100%`；筛选、排序、层级和分页参数保持原语义；浏览器控制台无警告或错误
- 当前状态：交互检查通过

### 页面：`/movie-agg-all/release-bucket-list`

- 页面文件：`v3/gc-web-next02/src/pages/ReleaseBucketListPage.vue`
- 主要列表：影片上映日聚合桶落库列表
- 筛选参数：`agg_mode`、`scope_key`、`year`、`quarter`、`month`、`day`、`ps`
- 排序参数：`sort`、`dir`
- 分页参数：`p`、`ps`
- 详情入口：`/movie-agg-all/release` 聚合页、`/triggers/agg` 回填页和聚合桶下钻页
- 列宽方案：时间桶 `26%`、总数 `14%`、已拥有 `14%`、总大小 `20%`、最新上映时间 `26%`
- 特殊尺寸保留项：`owned + day` 模式的月份分组行继续使用交替浅粉背景
- 新窗口入口：聚合页、回填页和时间桶下钻链接已设置 `target="_blank"` 和 `rel="noopener noreferrer"`；全量/已拥有模式切换继续使用当前窗口
- 构建结果：`make check-web-next02` 通过；`npm run build` 通过，保留现有构建产物体积提示
- 交互检查：桌面端 `1280px` 和窄屏 `390px` 均无页面或表格横向溢出；表头 `30px`、数据行 `28px`、字号 `13px`；5 列固定列宽总和为 `100%`；实测 `agg_mode=owned&level=day&ps=24` 保留模式、层级、分页参数，24 行均应用月份分组；浏览器控制台无警告或错误
- 当前状态：交互检查通过

## 20. 验收标准

每个页面必须同时满足：

- 使用当前 `/sc-events` 粉色主题。
- 顶部区域紧凑，无重复摘要和无需求留白。
- 筛选靠左，统计和分页靠右。
- 表头 `30px`、数据行 `28px`、字号 `13px`。
- 列宽根据实际内容规划且总和为 `100%`。
- 短数字列紧凑，长文本列获得主要空间。
- 长文本单行省略。
- 页面整体无横向滚动。
- 页面入口在新窗口打开并包含安全属性。
- 筛选、排序、分页仍在当前窗口执行。
- 原字段、按钮、参数、API 和业务行为完整保留。
- `git diff --check` 通过。
- 前端类型检查和生产构建通过。

## 21. 总体完成条件

只有以下条件全部满足，才允许将本计划标记为完成：

- 主列表页面矩阵全部完成。
- FetchSite 数据列表全部完成。
- 嵌入式列表区域全部完成。
- TriggerPage 各路由结果区全部完成。
- 所有特殊页面边界得到保留。
- 所有页面构建通过。
- 所有页面入口行为符合规范。
- 不存在误伤卡片页、详情页和其他 next 版本的全局样式覆盖。
