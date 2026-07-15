# Web 前端代码优化方案

## 1. 文档信息

- 目标目录：`web/`
- 当前前端：Vue 3、Vue Router、TypeScript、Vite
- 方案日期：2026-07-15
- 优化性质：保持现有功能、路由、接口和桌面端页面效果的渐进式重构
- 实施原则：先建立验证基线，再清理明确冗余，随后拆分职责、收紧类型、迁移样式，最后完成路由与工程化收敛

## 2. 优化目标

本次优化围绕可维护性、可验证性和模块边界展开，目标如下：

1. 保留现有 `app / features / shared` 顶层结构，明确每层职责。
2. 删除已经确认没有引用的页面和组件，减少历史代码干扰。
3. 收敛分页、排序、结果表格和相册管理等重复实现。
4. 将大型页面中的展示、状态管理、接口调用和事件流拆分到独立模块。
5. 将后端响应兼容逻辑集中在 feature 的 API 与 mapper 层，逐步减少页面中的 `AnyRecord`。
6. 将单一超大全局样式文件迁移为基础样式、共享样式和 feature 样式。
7. 将路由配置按 feature 拆分，并使用页面级懒加载。
8. 补充对公共组件、composable、接口映射和关键页面行为的测试。
9. 保持当前桌面端页面、接口参数、业务语义和持久化行为不变。

## 3. 本次方案边界

### 3.1 包含范围

- `web/src/app/`
- `web/src/features/`
- `web/src/shared/`
- `web/src/app/styles/legacy.css`
- `web/src/app/router/index.ts`
- `web/src/app/navigation.ts`
- `web/package.json` 中与检查、测试和构建相关的脚本
- 与前端优化直接相关的测试文件和说明文档

### 3.2 保持不变的内容

- 现有前端路由路径。
- 现有 API 路径、请求方法、参数含义和返回业务语义。
- 现有桌面端页面功能和关键交互。
- 现有服务监听地址 `0.0.0.0`。
- 现有响应式代码；本方案不增加移动端专用布局。
- 现有业务上限和截断规则，除非后续任务明确要求调整。

### 3.3 不在本方案中顺带处理的内容

- 后端接口和数据库结构调整。
- 页面视觉风格重新设计。
- 路由改名或兼容入口新增。
- 业务流程重做。
- 与当前前端优化无关的仓库文件。

## 4. 已验证的当前状态

### 4.1 目录结构

当前 `web/src` 已经按以下顶层结构组织：

```text
src/
├── app/
├── features/
└── shared/
```

该方向可以继续保留。主要问题集中在 feature 内部职责聚合、公共能力重复实现和全局样式持续叠加。

### 4.2 代码体量

检查到的主要大文件如下：

| 文件 | 当前行数 | 主要职责聚合 |
| --- | ---: | --- |
| `src/app/styles/legacy.css` | 38,957 | 全局基础、布局、组件、各 feature 页面、多轮视觉覆盖 |
| `features/crawler/pages/TriggerPage.vue` | 2,237 | 多种任务表单、轮询、SSE、预览、运行状态、事件展示 |
| `features/crawler/pages/ListPage.vue` | 1,491 | JavBus、Sukebei、SH 列表、SH 详情、字符串模板、DOM 事件代理 |
| `features/movies/pages/MoviePage.vue` | 1,359 | 影片信息、资源列表、排序、任务触发、相册操作和多个弹窗 |
| `features/sc/pages/ScEventsPage.vue` | 925 | 列表、卡片、详情、编辑、分页和内嵌渲染组件 |
| `features/movies/pages/MovieAlbumsPage.vue` | 882 | 相册筛选、选择、批量操作、管理弹窗和分页 |
| `features/sc/pages/ScPickSmartPage.vue` | 806 | 条件配置、字段组件、分组渲染和提交逻辑 |
| `features/movies/pages/TorrentAlbumsPage.vue` | 797 | 相册筛选、选择、批量操作、管理弹窗和分页 |

### 4.3 明确的零引用文件

以下文件经过全项目引用检查后为零引用：

1. `src/app/pages/HomePage.vue`
   - 根路由当前直接跳转到 `/listcardtoday`。
   - 路由配置和其他组件均未引用该页面。
2. `src/shared/components/DataTable.vue`
   - 当前没有任何页面导入该组件。
   - 组件内部存在从前 12 行推导字段、最多展示 14 列的历史截断逻辑。

执行清理前仍需再次运行引用检查，确认实施期间没有新增引用。

### 4.4 重复实现

静态检查发现以下内嵌组件重复：

- `SortLink`：9 处。
- `PageLinks`：6 处。
- `ResultTable`：5 处。
- 页面专用 Pagination：多处。

其中 `MovieAlbumsPage.vue` 和 `TorrentAlbumsPage.vue` 内的 `SortLink`、Pagination 代码已经确认相同。项目同时已有 `src/shared/components/PaginationBar.vue`，具备收敛基础。

### 4.5 数据类型

- `src/shared/types/gc.ts` 主要通过 `AnyRecord = Record<string, unknown>` 提供宽泛结构。
- `AnyRecord` 或同类宽泛记录类型在项目内约有 620 处引用。
- 页面大量使用 `pick()`、`text()` 和大小写兼容字段读取后端数据。
- `tsconfig.app.json` 已启用 `strict`，但宽泛记录类型降低了严格模式对业务字段的保护效果。

### 4.6 路由与导航

- `src/app/router/index.ts` 集中维护所有 feature 路由。
- 页面组件当前以同步 import 方式进入路由文件。
- `router/index.ts` 与 `navigation.ts` 分别维护部分相同路径信息。
- 全局 query 压缩和页面标题逻辑集中在 router 中，适合继续保留为 app 级能力。

### 4.7 样式状态

- `src/app/styles/index.css` 当前只导入 `legacy.css`。
- `legacy.css` 约占前端源码总行数的三分之二。
- 文件内包含 153 个媒体查询和 599 个 `!important`。
- 同一页面存在多轮追加式视觉重写，后置规则持续覆盖早期规则。
- 除 `ScDateTime.vue` 外，页面样式主要依赖全局 CSS。
- 静态扫描发现一批只在 CSS 中出现的 class 候选；动态路由 class、字符串模板和 `v-html` 使这些候选必须经过运行时页面覆盖验证后才能删除。

### 4.8 已执行检查

以下命令已经实际执行并通过：

```text
npm run check
npm test
vue-tsc --noEmit --noUnusedLocals --noUnusedParameters
```

当前测试结果为 8 项通过，测试主要覆盖 shared utility。正式构建和完整桌面端页面交互需要在每个实施阶段执行。

## 5. 目标目录结构

目标结构保留现有顶层分层，并让 feature 自己拥有路由、接口、类型、组件、状态逻辑和样式：

```text
web/src/
├── app/
│   ├── App.vue
│   ├── layouts/
│   │   └── AppLayout.vue
│   ├── navigation.ts
│   ├── pages/
│   │   └── NotFoundPage.vue
│   ├── router/
│   │   └── index.ts
│   └── styles/
│       ├── index.css
│       ├── tokens.css
│       ├── reset.css
│       └── shell.css
├── shared/
│   ├── api/
│   │   └── request.ts
│   ├── components/
│   │   ├── feedback/
│   │   ├── layout/
│   │   ├── navigation/
│   │   └── table/
│   ├── composables/
│   ├── styles/
│   │   ├── buttons.css
│   │   ├── forms.css
│   │   ├── modals.css
│   │   ├── pagination.css
│   │   └── tables.css
│   ├── types/
│   └── utils/
└── features/
    ├── movies/
    │   ├── api.ts
    │   ├── types.ts
    │   ├── mappers.ts
    │   ├── routes.ts
    │   ├── components/
    │   ├── composables/
    │   ├── pages/
    │   └── styles.css
    ├── crawler/
    │   └── 同上
    ├── media/
    │   └── 同上
    ├── sc/
    │   └── 同上
    ├── sh/
    │   └── 同上
    └── fc2/
        └── 同上
```

目录按实际内容渐进建立。单个 feature 内容较少时继续使用根级 `api.ts`、`types.ts` 和 `routes.ts`，避免提前创建空目录和过度分层。

## 6. 分阶段实施方案

### 阶段一：建立重构验证基线

#### 目标

为后续纯结构调整建立可重复验证的基准，避免将历史行为变化混入重构。

#### 检查对象

- 当前 Git 状态和用户已有修改。
- 关键路由与对应 API 路径。
- 桌面端关键页面初始截图和控制台状态。
- 当前类型检查、测试和构建结果。

#### 关键页面验证集合

- 卡片：`/listcardtoday`、`/listcardagg`、排名页面。
- 影片：影片详情、演员详情、影片相册、种子相册。
- Crawler：任务列表、抓取触发、JavBus/Sukebei 列表、DetailLoop。
- SH：一体抓取、列表建池、详情补全、帖子池、帖子详情、磁力、聚合桶。
- SC：事件列表、事件卡片、事件详情、Smart Pick、触发和媒体移动。
- Media：媒体列表、目录、入库、重扫、回滚、聚合事件。
- FC2：列表、详情、Cards 和媒体入库。

#### 验证方式

```text
npm run check
npm test
npm run build
```

页面验证固定使用项目约定的 Chrome MCP，检查：

- 页面可访问。
- 关键筛选、排序、分页和弹窗可操作。
- 路由跳转后的 query 与当前行为一致。
- 控制台没有新增错误。
- 接口请求路径、方法和载荷保持一致。

#### 产出

- 基线检查记录。
- 关键页面验证清单。
- 当前失败项与历史问题清单；历史问题与本次重构问题分别记录。

### 阶段二：清理明确冗余

#### 修改文件

- 删除 `src/app/pages/HomePage.vue`。
- 删除 `src/shared/components/DataTable.vue`。
- 修改 `src/app/styles/legacy.css`，清理只服务于 HomePage 和 DataTable 的确认无引用样式。

#### 主要改动

1. 再次确认 `HomePage` 和 `DataTable` 全项目零引用。
2. 删除两个文件。
3. 对以下 HomePage 专属 class 做精确引用检查后清理：
   - `home-grid`
   - `home-section`
   - `home-links`
4. 对以下 DataTable 专属 class 做精确引用检查后清理：
   - `table-card`
   - `table-card-head`
   - `data-table`
   - `empty-cell`
   - `table-link`
5. 同名片段或复合 class 仍被其他页面使用时保留对应规则。

#### 风险

- `legacy.css` 中部分通用 class 可能被字符串模板或动态内容使用。
- CSS 清理需要结合静态搜索和关键页面运行时验证，不能只依赖文件引用关系。

#### 验收标准

- 两个零引用 Vue 文件完成清理。
- 路由行为保持不变。
- 相关样式删除后关键页面无视觉回归。
- 类型检查、测试、构建和关键页面验证通过。

### 阶段三：统一排序与分页组件

#### 新增文件

- `src/shared/components/navigation/SortLink.vue`
- 根据实际类型归属新增或调整 shared pagination 类型文件。

#### 修改文件

- `src/shared/components/PaginationBar.vue`
- `features/media/pages/WAggEventsPage.vue`
- `features/media/pages/MediasPage.vue`
- `features/movies/pages/CastsPage.vue`
- `features/movies/pages/TorrentAlbumsPage.vue`
- `features/movies/pages/MovieAlbumsPage.vue`
- `features/sc/pages/ScEventsPage.vue`
- `features/crawler/pages/EItemsPage.vue`
- `features/crawler/pages/RecordsPage.vue`
- `features/crawler/pages/DSeedsPage.vue`
- 包含 `PageLinks` 的 FC2、SH、SC、Crawler、Movies 页面。

#### 组件职责

`SortLink.vue`：

- 接收明确的排序链接模型。
- 展示激活状态和升降序标记。
- 允许页面通过 class 或 slot 保留现有外观。
- 统一 href 规范化和无效链接处理。

`PaginationBar.vue`：

- 统一普通页、当前页、禁用页和省略号语义。
- 兼容当前后端返回的字段结构，并在 mapper 层逐步统一字段大小写。
- 保持分页位于页面右上角并贴右显示。
- 提供页面现有样式所需的 class 扩展能力。

#### 实施方式

1. 先以 MovieAlbums 和 TorrentAlbums 的完全重复实现建立公共组件。
2. 再按 feature 替换其他排序和分页实现。
3. 每替换一个 feature，立即验证对应列表页面。
4. 删除页面内的 `defineComponent + h()` 重复实现和不再使用的 import。

#### 验收标准

- 相同排序行为由同一公共组件承担。
- 相同分页行为由同一公共组件承担。
- 页面内不再保留同构的 `SortLink`、`PageLinks` 和 Pagination 实现。
- 分页位置、激活状态、禁用状态和 query 跳转保持一致。

### 阶段四：合并 Album 公共逻辑

#### 新增文件

```text
features/movies/components/albums/
├── AlbumFilterBar.vue
├── AlbumSelectionToolbar.vue
├── AlbumManageModal.vue
└── AlbumTableShell.vue

features/movies/composables/
├── useAlbumSelection.ts
└── useAlbumManagement.ts
```

#### 修改文件

- `features/movies/pages/MovieAlbumsPage.vue`
- `features/movies/pages/TorrentAlbumsPage.vue`
- 新增或调整 `features/movies/api.ts`
- 新增或调整 `features/movies/types.ts`

#### 公共逻辑范围

- 行选择、全选和半选状态。
- 相册创建、改名和删除弹窗状态。
- 相册筛选选项组织。
- 排序和分页展示。
- 通用成功、失败消息状态。

#### 保留在页面内的业务逻辑

- 影片相册与种子相册不同的行结构。
- 各自的批量操作和 API 路径。
- 下载完成条目清理、种子 Hash 复制等专属行为。
- 各自的业务按钮、列和操作条件。

#### 验收标准

- 两个页面不再复制相同的排序、分页、选择和管理弹窗实现。
- 页面仍保留清晰的业务差异。
- 创建、改名、删除、移动、复制和批量操作结果与当前一致。

### 阶段五：拆分 Crawler 列表页面

#### 当前问题

`features/crawler/pages/ListPage.vue` 当前同时承担四种页面渲染，通过 `v-html` 输出字符串模板，并通过根节点事件代理和 `document.querySelector` 处理交互。该结构降低了模板类型检查、组件复用和局部测试能力。

#### 目标文件

```text
features/crawler/pages/
├── JavbusListPage.vue
└── SukebeiListPage.vue

features/crawler/components/fetch-site/
├── FetchSiteFilterBar.vue
├── FetchSiteResultTable.vue
├── FetchSiteSummary.vue
└── FetchSiteActions.vue

features/sh/pages/
├── SHPoolPage.vue
└── SHThreadDetailPage.vue

features/sh/components/thread/
├── SHThreadFilterBar.vue
├── SHThreadTable.vue
├── SHThreadSummary.vue
└── SHThreadActions.vue
```

#### 路由归属调整

- `/javbuslist` 使用 `JavbusListPage.vue`。
- `/sukebeilist` 使用 `SukebeiListPage.vue`。
- `/shpool` 使用 `SHPoolPage.vue`。
- `/shthread/:threadTid` 使用 `SHThreadDetailPage.vue`。

所有路由路径和 API 路径保持不变。

#### 主要改动

1. 将字符串模板转换为 Vue template。
2. 将根节点事件代理转换为组件事件和明确的 handler。
3. 将 DOM 查询和 class 修改转换为响应式状态。
4. 将 SH forum 选项逻辑放入 SH feature。
5. 将 JavBus/Sukebei 共享结构留在 crawler 的 fetch-site components 中。
6. 删除完成迁移后的 `ListPage.vue`。

#### 风险

- 当前字符串模板中包含大量已有 class 和 data attribute，迁移时需要原样保留接口所依赖的字段和页面样式 class。
- 动态按钮状态、批量触发和筛选 query 是重点回归区域。

#### 验收标准

- 列表页面不再通过 `v-html` 渲染业务页面主体。
- 页面不再依赖全局 DOM 查询完成主要交互。
- 四个现有路由及其筛选、分页、触发和详情行为保持一致。

### 阶段六：拆分 TriggerPage

#### 目标结构

```text
features/crawler/
├── pages/
│   └── TriggerPage.vue
├── components/trigger/
│   ├── TriggerForm.vue
│   ├── TriggerOverview.vue
│   ├── JobSelector.vue
│   ├── JobRuntimePanel.vue
│   ├── JobEventStream.vue
│   ├── FetchSitePreviewModal.vue
│   └── SHDetailPreview.vue
├── composables/
│   ├── useCrawlerJobs.ts
│   ├── useJobEventStream.ts
│   ├── useTriggerForm.ts
│   └── useTriggerPreview.ts
├── api.ts
├── mappers.ts
└── types.ts
```

#### 职责划分

`TriggerPage.vue`：

- 根据 route meta 选择页面配置。
- 组合任务表单、状态面板、预览和事件流。
- 管理页面级错误和加载状态。

`useCrawlerJobs.ts`：

- 加载任务列表。
- 管理当前任务选择。
- 管理轮询生命周期。
- 提供暂停、恢复和停止任务能力。

`useJobEventStream.ts`：

- 建立和关闭 EventSource。
- 管理重连状态和最后事件 ID。
- 解析并去重事件。
- 保持当前事件历史规则，不在结构重构中改变保留逻辑。

`useTriggerForm.ts`：

- 根据任务类型生成表单状态和提交 payload。
- 维护 SH、FetchSite、DailyBest、Seeds、PostProcess 和 Backfill 的差异。

`useTriggerPreview.ts`：

- 管理 FetchSite 和 SH 详情预览。
- 管理预览加载、分页、汇总和关闭状态。

#### 实施顺序

1. 先提取无 UI 的事件格式化、任务映射和 payload 构造函数，并补充测试。
2. 提取 EventSource 和轮询生命周期。
3. 提取预览状态。
4. 提取展示组件。
5. 最后简化页面模板和 route 分支。

#### 验收标准

- 页面只承担组合和路由级状态。
- SSE 连接、重连、去重、选择任务和定时器清理保持正确。
- 各任务类型提交 payload 与当前一致。
- 暂停、恢复、停止和预览行为通过页面验证。

### 阶段七：拆分 MoviePage 与其他大型页面

#### MoviePage 目标结构

```text
features/movies/components/movie-detail/
├── MovieHero.vue
├── MovieInfoPanel.vue
├── MovieStatusPanel.vue
├── MovieResourceTable.vue
├── MovieRankTable.vue
├── MovieFetchStatusCard.vue
├── MovieAlbumModal.vue
└── MovieActionModals.vue

features/movies/composables/
├── useMovieActions.ts
├── useMovieResources.ts
└── useMovieAlbums.ts
```

#### MoviePage 拆分内容

- 影片基础信息和媒体信息展示。
- 排名排序与分页。
- JavBus、Sukebei、SH 资源表格。
- 稍后下载、封面下载、演员、Hash 和媒体移动操作。
- 资源收藏相册操作。
- 抓取任务状态卡片。

当前页面内的 `ResourceTable` 和 `FetchStatusCard` 改为独立 `.vue` 文件，减少 render function 与父页面闭包状态耦合。

#### 其他大型页面

完成前三个核心页面后，再按相同原则处理：

- `ScEventsPage.vue`
  - 列表、卡片和详情视图拆成独立组件。
  - `PageLinks`、`SortLink`、ViewSwitch、MetricCard 使用公共或 feature 组件。
- `ScPickSmartPage.vue`
  - 字段输入、选择器和条件组改为独立 `.vue` 组件。
- `FC2DetailPage.vue`
  - 详情信息、媒体信息、资源和操作模块独立。
- Media 入库、重扫、回滚和 SC Media Move
  - 比较各 `ResultTable` 数据合同。
  - 相同合同提取 feature 组件；字段和行为不同的表格保持独立。

#### 验收标准

- 页面组件聚焦页面组合。
- 子组件通过 props、emits 和明确类型通信。
- 页面内不再保留复杂的 `defineComponent + h()` 业务组件。
- 所有详情操作、表格排序、弹窗和接口结果保持一致。

### 阶段八：建立 feature API、类型和 mapper 边界

#### 目标

逐步将页面从宽泛后端响应结构转换为明确的前端业务模型。

#### 文件职责

`shared/api/request.ts`：

- HTTP 请求。
- JSON 解析。
- 状态码与统一错误。
- 通用请求选项。

`feature/api.ts`：

- 定义具体接口函数。
- 定义方法和 payload 类型。
- 调用 mapper 返回规范化模型。

`feature/types.ts`：

- 页面模型。
- 列表行、详情、分页、排序和操作 payload 类型。
- 任务、事件、相册、媒体等业务类型。

`feature/mappers.ts`：

- 兼容后端大小写字段。
- 处理可选字段和缺省值。
- 将原始响应转换为前端统一模型。
- 对必要字段执行明确校验。

#### 推进顺序

1. MovieAlbums、TorrentAlbums。
2. Crawler Job、Job Event 和 Trigger 配置。
3. Movie Detail、Movie Resource 和 Movie Action。
4. SC Events、Smart Pick。
5. Media、SH、FC2 其他页面。

#### AnyRecord 收敛规则

- API 原始响应边界允许暂时使用宽泛记录。
- mapper 输出必须是明确类型。
- 页面和展示组件优先只使用 mapper 输出。
- 无法稳定定义的数据需要在类型中显式标记可选字段，避免继续向整个页面扩散 `AnyRecord`。

#### 验收标准

- 新拆分模块拥有明确的输入、输出和事件类型。
- 页面不再反复读取同一字段的多种大小写形式。
- 字段兼容集中在 mapper。
- TypeScript 能够发现主要业务字段拼写和类型错误。

### 阶段九：拆分 legacy.css

#### 目标结构

```text
src/app/styles/
├── index.css
├── tokens.css
├── reset.css
└── shell.css

src/shared/styles/
├── buttons.css
├── forms.css
├── modals.css
├── pagination.css
└── tables.css

src/features/movies/styles.css
src/features/crawler/styles.css
src/features/media/styles.css
src/features/sc/styles.css
src/features/sh/styles.css
src/features/fc2/styles.css
```

#### 迁移原则

1. 第一轮只移动规则，不改选择器、属性和值。
2. `index.css` 按 `legacy.css` 当前先后顺序导入拆分文件，保持级联顺序。
3. 基础变量、reset 和 app shell 优先拆分。
4. 公共按钮、表单、分页、表格和弹窗随后拆分。
5. 各页面规则按 feature 迁移。
6. 所有规则完成物理迁移并验证后，再开始合并重复规则。
7. 删除规则前同时检查静态引用、动态 class、字符串模板和实际页面覆盖。

#### 样式归属

- `app/styles`：全局 token、reset、应用壳和导航。
- `shared/styles`：多个 feature 真实复用的视觉组件。
- `feature/styles.css`：只服务于对应 feature 的页面和组件。
- 单个组件完全独有且边界清晰的样式可以迁入组件 scoped style。

#### 清理顺序

1. 明确零引用页面和组件的专属样式。
2. 同一选择器在相同作用域中的重复声明。
3. 已被后置规则完整覆盖的历史块。
4. 通过共同 class 表达的重复视觉规则。
5. `!important` 依赖和过长路由前缀选择器。

#### 风险控制

- 每次只迁移一个职责明确的样式组并验证相关页面。
- 保留现有桌面端布局和响应式代码。
- 不在样式拆分阶段调整页面 DOM 结构。
- CSS 删除必须有运行时页面验证证据。

#### 验收标准

- `legacy.css` 中的规则完成有归属的迁移，最终移除该过渡文件。
- app、shared 和 feature 样式边界清晰。
- 关键页面视觉和交互状态保持一致。
- hover、active、focus、disabled 和 `prefers-reduced-motion` 行为保持完整。

### 阶段十：路由模块化与懒加载

#### 新增文件

```text
features/movies/routes.ts
features/crawler/routes.ts
features/media/routes.ts
features/sc/routes.ts
features/sh/routes.ts
features/fc2/routes.ts
```

#### 修改文件

- `src/app/router/index.ts`
- `src/app/navigation.ts`
- `src/env.d.ts` 或新增 Vue Router meta 类型声明文件。

#### 主要改动

1. 每个 feature 导出自己的 `RouteRecordRaw[]`。
2. 页面组件改为动态 import，实现页面级懒加载。
3. `app/router/index.ts` 只汇总 feature routes，并保留：
   - 根路由跳转。
   - query 压缩守卫。
   - document title。
   - 404 路由。
4. 为路由补充稳定的 route name。
5. 为 `meta.title` 和 `meta.apiPath` 建立类型声明。
6. 导航使用 route name 或共享路由常量，减少路径字符串重复。

#### 路由约束

- 保持所有现有路径。
- 保持所有现有 API path meta。
- 不新增旧路由别名、重定向或兼容入口。
- 根路由继续跳转到 `/listcardtoday`。

#### 验收标准

- 每个 feature 的路由与实现同目录维护。
- app router 保持轻量。
- 首次加载不再同步包含所有页面组件。
- 导航、直接输入 URL、刷新详情页和 query 跳转全部正常。

### 阶段十一：工程化检查与测试补充

#### package scripts

在现有脚本基础上补充并统一：

- 类型检查。
- 单元测试。
- lint。
- 格式检查。
- 构建。

具体工具和依赖在实施阶段根据当前 Vue、TypeScript 和 Node 环境确定，避免引入与现有运行环境冲突的配置。

#### 优先测试内容

1. `shared/api/request.ts`
   - 成功响应。
   - 非 JSON 错误。
   - JSON 错误。
   - HTTP 状态码。
2. SortLink 与 PaginationBar。
3. Album selection 与 management composable。
4. Feature mapper 字段兼容。
5. Trigger payload 构造。
6. Crawler Job 状态映射。
7. Job Event 去重、格式化和生命周期。
8. Route query 压缩和 feature route meta。

#### 页面验证

- 只验证项目要求的桌面端页面。
- 检查页面加载、空状态、错误状态、筛选、排序、分页、弹窗和提交结果。
- 检查控制台错误和关键网络请求。
- 对样式迁移页面保存迁移前后截图进行比对。

#### 验收标准

- `npm run check` 通过。
- `npm test` 通过。
- `npm run build` 通过。
- 新增 lint 和格式检查通过。
- 关键桌面端页面和操作验证通过。

## 7. 分阶段文件变更清单

| 阶段 | 新增 | 修改 | 删除 |
| --- | --- | --- | --- |
| 基线 | 验证记录 | 无业务代码修改 | 无 |
| 冗余清理 | 无 | `legacy.css` | `HomePage.vue`、`DataTable.vue` |
| 排序分页 | `SortLink.vue`、必要类型 | `PaginationBar.vue`、相关列表页面 | 页面内重复组件 |
| Album | Album components、composables、types/api | 两个 Album 页面 | 页面内重复实现 |
| Crawler List | 四个业务页面、列表组件 | feature routes | `ListPage.vue` |
| Trigger | trigger components、composables、types/api/mappers | `TriggerPage.vue` | 页面内嵌复杂模块 |
| Movie Detail | movie detail components、composables | `MoviePage.vue` | 页面内 render components |
| 类型/API | feature `api.ts/types.ts/mappers.ts` | 各页面和组件 | 被替代的宽泛读取逻辑 |
| 样式 | app/shared/feature 样式文件 | `index.css`、相关组件 | `legacy.css` 过渡文件 |
| 路由 | feature `routes.ts` | app router、navigation、meta 类型 | 集中式同步 imports |
| 工程化 | 测试和配置文件 | `package.json` | 重复测试入口按实际检查处理 |

## 8. 风险与控制措施

### 8.1 CSS 级联回归

风险：后置样式覆盖前置样式，物理拆分可能改变优先级。

控制：

- 第一轮严格保持导入顺序。
- 样式迁移和 DOM 重构分阶段实施。
- 通过关键页面截图、交互状态和控制台进行验证。

### 8.2 动态页面行为丢失

风险：`ListPage.vue` 的字符串模板和事件代理隐含部分交互合同。

控制：

- 迁移前逐项记录按钮、data attribute、事件和接口。
- 转换为 Vue 组件时保留原 class、query 和 payload。
- 四个路由逐页验证。

### 8.3 SSE 和轮询生命周期错误

风险：TriggerPage 拆分后出现重复连接、定时器未清理或任务选择不同步。

控制：

- 先为纯函数和生命周期逻辑补测试。
- composable 明确暴露 start、stop、reload 和 dispose。
- 页面卸载和路由切换时验证资源清理。

### 8.4 类型收紧改变兼容行为

风险：后端存在多种字段大小写，直接修改页面读取方式可能遗漏字段。

控制：

- mapper 同时覆盖当前代码读取的所有已知字段。
- 使用当前真实接口响应验证 mapper。
- 类型迁移按 feature 推进，不全项目一次替换。

### 8.5 公共组件过度抽象

风险：不同页面业务差异被强行放入大量条件参数。

控制：

- 只抽取已经确认同构的结构和行为。
- 业务表格优先保留 feature 归属。
- 公共组件通过 slot 提供展示扩展，避免加入业务判断。

### 8.6 用户已有修改冲突

风险：实施期间工作区可能存在其他修改。

控制：

- 每个阶段开始前检查 Git 状态。
- 只修改阶段清单内文件。
- 不覆盖、不回滚无关修改。
- 发现重叠修改时先核对差异再继续。

## 9. 验证矩阵

| 修改类型 | 必做检查 |
| --- | --- |
| 删除零引用文件 | 全项目引用搜索、类型检查、构建、根路由验证 |
| 公共组件替换 | 对应列表页排序、分页、禁用态、当前页验证 |
| Album 合并 | 筛选、选择、全选、创建、改名、删除、移动、复制和批量操作 |
| ListPage 拆分 | 四个路由、筛选、分页、触发、详情和 query |
| TriggerPage 拆分 | 各任务提交、任务选择、轮询、SSE、暂停、恢复、停止和预览 |
| MoviePage 拆分 | 详情加载、资源排序、抓取、相册、Hash、演员、封面和媒体操作 |
| 类型/API 迁移 | 真实接口响应、错误状态、可选字段和 mapper 测试 |
| CSS 迁移 | 页面截图、布局、表格、弹窗、交互状态和 reduced motion |
| 路由模块化 | 导航、直接访问、刷新、详情参数、query 和 404 |

## 10. 每阶段交付格式

每个阶段完成后统一交付：

1. 实际修改文件列表。
2. 主要结构调整和保持不变的业务行为。
3. 实际执行的检查命令及结果。
4. 实际验证的页面、接口和交互。
5. 未验证项及原因。
6. 仍存在的历史问题。
7. 下一阶段依赖，但不在当前阶段顺手修改。

## 11. 完成标准

整个优化计划完成时，应满足：

- 零引用页面和组件已经清理。
- 相同的排序、分页和相册管理能力不再散落重复实现。
- `ListPage.vue` 的业务字符串模板和主要 DOM 事件代理已转换为 Vue 组件。
- `TriggerPage.vue`、`MoviePage.vue` 和其他大型页面职责已经拆分。
- feature 拥有自己的 API、类型、mapper、路由和样式。
- 页面业务代码主要使用明确类型，字段兼容集中在接口边界。
- `legacy.css` 完成迁移并移除。
- app router 只保留全局路由职责，页面按需加载。
- 类型检查、测试、lint、格式检查和构建全部通过。
- 关键桌面端页面、控制台和网络请求完成实际验证。
- 路由、接口、参数、业务语义和页面功能与优化前保持一致。

## 12. 推荐执行顺序

1. 建立验证基线。
2. 清理 `HomePage.vue`、`DataTable.vue` 及确认无引用样式。
3. 统一 SortLink 和 PaginationBar。
4. 合并 MovieAlbums 与 TorrentAlbums 公共能力。
5. 拆分 Crawler ListPage。
6. 拆分 TriggerPage。
7. 拆分 MoviePage、ScEventsPage 和其他大型页面。
8. 随页面拆分逐步建立 feature API、types 和 mappers。
9. 在 DOM 结构稳定后迁移并清理 legacy.css。
10. 完成 feature routes、懒加载和导航收敛。
11. 补齐工程化检查、单元测试和完整桌面端验证。

该顺序优先处理确定性高的冗余和重复代码，再处理状态复杂页面，最后收敛全局样式和路由，便于逐阶段验证并控制回归范围。
