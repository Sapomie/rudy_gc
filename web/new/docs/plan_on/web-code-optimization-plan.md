# Web 前端重建实施方案（已执行）

## 1. 文档信息

- 仓库：`/Users/gaojinwei/src/my/rudy_gc/web`
- 技术栈：Vue 3、TypeScript、Vue Router、Vite
- 执行日期：2026-07-16
- 执行状态：8 个阶段全部完成
- 进度与证据：[`web-code-optimization-progress.md`](./web-code-optimization-progress.md)

本文保留最终实施边界、落地结构和验收口径。阶段过程、命令结果、性能数据和页面验证记录集中写入进度文档。

## 2. 目标与完成口径

本次重建同时处理结构、冗余、类型、样式、加载性能和工程验证，完成口径如下：

1. 保留全部现有 URL、API 路径、HTTP 方法、参数语义和桌面端功能。
2. 删除确认零引用或已被新结构替代的文件。
3. 排序、分页、结果表格等通用能力集中到 shared。
4. 六个 feature 分别维护 API、类型、mapper、routes 和样式入口。
5. 大页面拆为路由页面、业务视图、展示组件和 composable。
6. 兼容型原始响应只停留在请求和 mapper 边界，页面优先使用 feature 类型。
7. 移除全量 `legacy.css`，保留实际使用规则并生成 feature CSS chunk。
8. 所有路由页面采用动态 import，降低首屏 JS 和开发态模块请求数。
9. 增加严格检查与关键纯函数、请求、路由、composable 测试。
10. 使用相同页面和环境对比重建前后的构建体积、资源数和浏览器导航指标。

## 3. 实施边界

### 3.1 修改范围

- `src/app/`
- `src/features/`
- `src/shared/`
- `vite.config.ts`
- `package.json`、`package-lock.json`
- ESLint、Prettier、Vitest 配置
- `docs/plan_on/`

### 3.2 保持项

- 根路由继续进入 `/listcardtoday`。
- 原有前端 URL 原样保留，不新增旧路由 alias 或兼容入口。
- 原有 API URL、HTTP 方法和业务 payload 保持。
- 服务监听地址保持 `0.0.0.0`。
- 保留已有响应式规则，本次只验收桌面端。
- 保留业务已有的数据数量规则，不新增上限、默认截断或静默裁剪。
- 后端、数据库结构和持久化语义不在本次修改范围。

### 3.3 验收安全边界

- 页面只执行加载、筛选、排序、分页、选择、弹窗开关和无副作用状态切换。
- 未在真实数据上执行删除、移动、覆盖或批量写入操作。
- 浏览器验证只使用连接 `127.0.0.1:9222` 的 Chrome MCP。
- Git 只更新暂存区，不自动提交、不自动推送。

## 4. 最终目录结构

```text
src/
├── app/
│   ├── layouts/
│   ├── router/
│   ├── styles/
│   │   ├── index.css
│   │   ├── tokens.css
│   │   ├── reset.css
│   │   ├── shell.css
│   │   └── navigation.css
│   └── navigation.ts
├── features/
│   ├── movies/
│   ├── crawler/
│   ├── sh/
│   ├── sc/
│   ├── media/
│   └── fc2/
│       ├── api.ts
│       ├── types.ts
│       ├── mappers.ts
│       ├── routes.ts
│       ├── pages/
│       ├── components/
│       ├── composables/
│       └── styles/
└── shared/
    ├── api/
    ├── components/
    │   ├── navigation/
    │   └── table/
    ├── router/
    ├── styles/
    └── utils/
```

各 feature 根据实际需要保留 components、composables 等子目录；`api.ts`、`types.ts`、`mappers.ts`、`routes.ts` 和样式入口已覆盖六个 feature。

## 5. 八阶段实施结果

### 阶段一：建立真实性能和功能基线

完成内容：

- 核对 Web Git 基线，实际基线 commit 为 `7297d57 初始化 Web 前端项目`。
- 执行重建前类型检查、测试和构建。
- 记录 7 组 API 直连/代理接口和 10 个关键页面。
- 记录构建体积、开发态资源数、DOMContentLoaded、load、LCP、TTFB 和 CLS。
- 记录关键页面的 DOM 结构、布局尺寸、控件状态、Console 和 Network 基线。

### 阶段二：清理冗余并建立共享能力

删除：

- `src/app/pages/HomePage.vue`
- `src/shared/components/DataTable.vue`

新增或统一：

- `src/shared/components/navigation/SortLink.vue`
- `src/shared/components/navigation/models.ts`
- `src/shared/components/PaginationBar.vue`
- `src/shared/components/table/ResultTable.vue`

排序、分页已覆盖 Movie、Crawler、SH、SC、Media、FC2 的重复实现；共享组件只处理展示合同，业务判断保留在 feature。

### 阶段三：建立 feature 数据边界

六个 feature 均建立：

- 具体 endpoint 和 HTTP 方法：`api.ts`
- 页面与业务模型：`types.ts`
- 原始字段兼容和转换：`mappers.ts`
- feature 路由集合：`routes.ts`

`AnyRecord` 使用从约 606 处收敛到 4 处，只保留在共享原始数据边界。通用请求的 JSON、文本和 HTTP 错误处理集中在 `shared/api/request.ts`。

### 阶段四：拆分大型页面

主要落地：

- Crawler List 拆为 JavBus、Sukebei、SH Pool、SH Thread 页面及 FetchSite View。
- Trigger 拆为表单、概览、任务选择、队列、运行状态、事件流、预览组件和 4 个 composable。
- Movie Detail 拆为详情 View、action、resource、album composable。
- Movie/Torrent Album 拆为各自 View，并共享选择和管理 composable。
- SC Events、Smart Pick、EItems、Media Rescan、FC2 Detail、聚合页面的主要业务视图已迁入 View/组件；其余中型页面按单一职责保留页面级实现。
- DSeed 编辑、EItem 状态、目录子项、重扫结果等局部职责拆成独立组件。

所有 `features/**/pages/*.vue` 已移除复杂内嵌 `defineComponent + h()` 业务组件；Trigger、Movie Detail、Album、SC Events、Smart Pick、EItems、Media Rescan、FC2 Detail 和聚合页面已形成页面组合层与业务 View/组件层。

### 阶段五：CSS 实质清理和按需加载

完成内容：

- 删除 39,009 行 `src/app/styles/legacy.css`。
- 将实际使用规则按 app、shared、movies、crawler、sh、sc、media、fc2 拆分。
- CSS 静态引用复核保留 5,427 条实际使用规则，删除 160 条零引用规则。
- app 只全局加载 token、reset、shell、navigation 和共享样式。
- 生产构建生成 6 个 feature CSS chunk，不再把全部页面样式塞入单一首屏 CSS。
- loading pulse 改为 opacity 合成动画，并保留 `prefers-reduced-motion`。

### 阶段六：路由模块化和懒加载

完成内容：

- `src/app/router/index.ts` 仅保留路由汇总、根跳转、全局 query 处理、标题和 404。
- 六个 feature 各自维护 `routes.ts`。
- 所有页面组件使用动态 import。
- 现有 URL、详情参数和 query 保持。
- 生产构建生成页面级异步 JS chunk，主入口不再包含全部页面实现。

### 阶段七：工程检查和测试

新增或完善：

- `vue-tsc --noEmit`
- 严格未使用变量/参数检查
- ESLint
- Prettier
- Vitest
- 生产构建

测试覆盖请求成功/失败解析、mapper、分页/排序模型、Album selection、Trigger form、任务时间和 feature routes；最终为 8 个测试文件、19 项测试。

### 阶段八：页面、网络和性能验收

完成内容：

- Chrome MCP 覆盖 61 个路由入口、动态参数样例、根路由和 404。
- 使用结构化 DOM、元素边界、计算样式和交互状态复核关键页面。
- 验证 Movie Album 选择和管理弹窗、SC Pick 条件增删、DailyBest SSE。
- DailyBest 页面保持 1 条 EventSource 连接。
- 7 组最终 API 直连和 Vite 代理请求均为 HTTP 200。
- Web `5022`、API `5013` 均监听全部接口；服务器局域网地址可访问 Web 和 API。
- 受控同环境对比中，主 JS、全局 CSS、资源数、DOMContentLoaded、load 和 LCP 均明显下降。

## 6. API 与路由合同

本次重建保持以下合同：

- 前端路径和 query 名称保持。
- feature router 只搬迁定义，不修改 route path。
- API endpoint、method、payload 字段和业务含义保持。
- 原始响应兼容由 mapper 接管，页面不改变后端返回合同。
- 根路由、详情直达、浏览器刷新和 404 均已实测。

## 7. 最终验证命令

```bash
npm run check
npm test
npm run build
```

三条命令均通过。详细构建体积、API 时间、Chrome 页面结果和基线对比见进度文档。

## 8. 最终状态

- 计划阶段：8 / 8 完成。
- 计划内剩余代码任务：无。
- 已知阻塞：无。
- 未验证项：移动端专用布局、真实破坏性业务写操作；两项均不属于本次验收范围。
- Git：全部本次变更已更新到暂存区，未创建 commit，未推送。
