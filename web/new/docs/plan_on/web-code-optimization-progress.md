# Web 前端重建执行进度与验收记录

## 1. 最终状态

- 对应方案：[`web-code-optimization-plan.md`](./web-code-optimization-plan.md)
- 仓库：`/Users/gaojinwei/src/my/rudy_gc/web`
- 执行日期：2026-07-16
- 当前状态：全部完成
- 已完成阶段：8 / 8
- 阻塞：无
- Git 基线：`7297d57 初始化 Web 前端项目`
- Git 交付：变更已暂存，未提交，未推送

## 2. 阶段进度

| 阶段 | 内容 | 状态 | 完成日期 | 核心证据 |
| --- | --- | --- | --- | --- |
| 一 | 真实性能和功能基线 | 完成 | 2026-07-16 | 初始构建、API/Chrome 与结构化页面基线 |
| 二 | 冗余清理和共享组件 | 完成 | 2026-07-16 | 删除 2 个零引用文件，统一排序/分页/结果表格 |
| 三 | Feature API、类型、mapper | 完成 | 2026-07-16 | 6 个 feature 数据边界，`AnyRecord` 606→4 |
| 四 | 大型页面拆分 | 完成 | 2026-07-16 | 路由页轻量化，View/组件/composable 拆分 |
| 五 | CSS 清理和按需加载 | 完成 | 2026-07-16 | 删除 legacy，5,427 条规则保留，160 条零引用删除 |
| 六 | 路由模块化和懒加载 | 完成 | 2026-07-16 | 6 个 feature routes，页面级异步 chunk |
| 七 | 工程检查和测试 | 完成 | 2026-07-16 | check 通过，8 个文件 19 项测试通过，build 通过 |
| 八 | 页面、网络和性能验收 | 完成 | 2026-07-16 | 61 个路由通过，Console 0，7 组 API 200 |

## 3. 代码结构结果

### 3.1 冗余清理

- `src/app/pages/HomePage.vue`：静态引用、路由引用和动态引用均为零，已删除。
- `src/shared/components/DataTable.vue`：静态引用、路由引用和动态引用均为零，已删除。
- `src/app/styles/legacy.css`：规则拆分并复核后删除。

### 3.2 共享能力

- `PaginationBar.vue`：兼容 pageInfo 和原始分页 links，支持 pill/bootstrap 两种现有视觉。
- `SortLink.vue`：统一 href、active、direction、disabled、label。
- `navigation/models.ts`：集中分页和排序展示模型。
- `ResultTable.vue`：按调用方显式列/行渲染，不推导或截断业务数据。

共享分页/排序已用于 WAggEvents、Medias、DSeeds、Records、Casts、EItems、SH Bucket、Release Bucket、FC2、SH Magnet、SC Events、Movie Albums 和 Torrent Albums。

### 3.3 Feature 边界

以下 feature 均有 `api.ts`、`types.ts`、`mappers.ts`、`routes.ts` 和样式入口：

- movies
- crawler
- sh
- sc
- media
- fc2

`AnyRecord` 从约 606 处降至 4 处，仅保留在共享动态响应边界。页面所需大小写兼容、可选字段和默认值由 mapper/feature helper 处理。

### 3.4 页面拆分

主要新增模块：

- Trigger：`TriggerForm`、`TriggerOverview`、`JobSelector`、`JobQueueColumn`、`JobRuntimePanel`、`JobEventStream`、两个 Preview 组件和 4 个 composable。
- Movie：聚合 View、Movie Detail View、Movie/Torrent Album View，以及 action/resource/album/selection/management composable。
- Crawler：FetchSite View、DSeed modal/number field、EItems View/status modal。
- Media：MediaRescan View、目录子项网格、重扫结果列表。
- SC：ScEvents View、ScPickSmart View。
- FC2：FC2 Detail View。

最终 `features/**/pages/*.vue` 中已无复杂内嵌 `defineComponent + h()` 业务组件。计划列出的核心大型页面已迁入 View、组件或 composable；其余中型页面继续按单一页面职责维护。

## 4. CSS 与构建结果

### 4.1 CSS 处理

- 旧 `legacy.css`：39,009 行，已删除。
- 静态规则复核：保留 5,427 条实际使用规则，删除 160 条零引用规则。
- 样式分配：app 负责 token/reset/shell/navigation，shared 负责 button/form/table/modal/pagination，六个 feature 各自加载业务样式。
- 最终生成 6 个 feature CSS chunk：58.74、93.24、109.79、116.04、119.75、138.14 kB。
- 最终全局 CSS：604.02 kB，gzip 73.89 kB。

### 4.2 生产构建

最终构建：

- Vite 6.4.3
- 转换模块：221
- 主入口 JS：125.23 kB，gzip 46.21 kB
- 页面和共享能力均生成独立异步 chunk
- 构建耗时：936 ms
- 无大于 500 kB 的 JS chunk 警告

## 5. 基线与最终对比

对比页面为 `/listcardtoday`，旧版和新版在相同 Chrome、相同视口、相同时间段和本机环境测试。

| 指标 | 重建前 | 重建后 | 变化 |
| --- | ---: | ---: | ---: |
| 主入口 JS | 737.88 kB | 125.23 kB | -83.0% |
| 主入口 JS gzip | 194.91 kB | 46.21 kB | -76.3% |
| 全局 CSS | 870.55 kB | 604.02 kB | -30.6% |
| 全局 CSS gzip | 107.39 kB | 73.89 kB | -31.2% |
| 开发态资源数 | 85 | 53 | -37.6% |
| 开发态 script 数 | 66 | 34 | -48.5% |
| stylesheet 数 | 2 | 2 | 持平 |
| 传输量 | 3,049,994 B | 3,040,394 B | -9,600 B |
| DOMContentLoaded | 129 ms | 70 ms | -45.7% |
| load | 176 ms | 108 ms | -38.6% |
| 受控热缓存 LCP 中位数 | 255 ms | 113 ms | -55.7% |
| TTFB | 4–7 ms | 4–5 ms | 稳定 |
| CLS | 0.00 | 0.0071 | Good，低于 0.1 |

说明：新版单次冷图片缓存样本出现 316–338 ms LCP；相同缓存条件下的三次中位数为 113 ms。CLS 的 0.0071 来自侧栏当前分组展开动画，Chrome 未给出可直接消除的 estimated saving，数值仍处于 Good 范围。其余首屏体积、资源和导航指标均明显改善。

## 6. API 验收

以下为最终同机实测，直连端口 `5013`，代理端口 `5022`：

| API | 直连 total | Vite 代理 total | 响应 |
| --- | ---: | ---: | --- |
| `/api/cardstoday` | 0.027421 s | 0.022631 s | 200，46,897 B |
| `/api/movie/OVVR-614` | 0.012643 s | 0.004573 s | 200，3,768 B |
| `/api/movie-albums` | 0.003838 s | 0.004396 s | 200，11,269 B |
| `/api/triggers/dailybest` | 0.000474 s | 0.001157 s | 200，953 B |
| `/api/fetch-site-sehuatang-sh-list` | 0.023110 s | 0.017069 s | 200，91,951 B |
| `/api/sc-events` | 0.045478 s | 0.039560 s | 200，16,053 B |
| `/api/fc2/movies/FC2PPV-4845383-1` | 0.002494 s | 0.002753 s | 200，2,109 B |

API 路径、HTTP 方法和 payload 合同未修改。抽查页面的 Vite 代理和 API 直连均成功。

## 7. Chrome MCP 页面验收

### 7.1 全路由

- 覆盖范围：Movie/Card、Crawler、SH、SC、Media、FC2。
- 覆盖数量：61 个当前路由入口、动态参数样例、根路由和 404。
- 页面主内容：全部出现。
- alert/error 面板：0。
- 最终 Console error/warn：0。
- 动态样例使用当前 API 返回的有效实体；SH Thread 使用 `/shthread/3630136`。

### 7.2 关键交互

- Movie Album：第一行选择数 `0 → 1 → 0`，批量按钮状态正确恢复；管理弹窗可打开和关闭。
- Torrent Album：页面、分页、排序和列表正常，控制台无警告或错误。
- SC Pick：新增 JAV 条件组 `2 → 3`，删除该组恢复 `3 → 2`。
- DailyBest：页面只建立 1 条 `/api/crawler/jobs/stream` EventSource，状态 200。
- SC Events：共享分页和 12 个排序入口均生成正确链接，控制台无警告或错误。

未执行真实删除、移动、写回、批量提交等会改变业务数据的动作。

### 7.3 结构化页面证据

页面证据由 Chrome MCP 读取并对照以下数据：

- 页面标题、路由、主体区域和导航状态。
- 表单、表格、卡片、图片、分页、弹窗和操作按钮数量。
- 关键元素的边界尺寸、计算样式、换行、溢出和裁切状态。
- 无副作用交互前后的文本、属性、选中项、禁用状态和 DOM 变化。
- Console error/warn、失败请求、API 状态和 EventSource 数量。

## 8. 服务与局域网访问

- API：TCP `*:5013`，监听全部接口。
- Vite：TCP `*:5022`，监听全部接口。
- 实际局域网 IP：`192.168.31.165`。
- 局域网访问 Web `/listcardtoday`：HTTP 200，total 0.0112 s。
- 局域网访问 API `/api/cardstoday`：HTTP 200，total 0.0210 s。

监听地址满足 `0.0.0.0` 约束，局域网实际 IP 访问已验证。

## 9. 工程命令结果

### `npm run check`

- `vue-tsc --noEmit`：通过。
- `vue-tsc --noEmit --noUnusedLocals --noUnusedParameters`：通过。
- ESLint：通过。
- Prettier check：通过。

### `npm test`

- 测试文件：8 / 8 通过。
- 测试项：19 / 19 通过。
- 最终耗时：164 ms。

### `npm run build`

- 类型检查：通过。
- Vite 生产构建：通过。
- 模块数：221。
- 最终构建耗时：936 ms。

## 10. 最终交付清单

- [x] 冗余文件删除
- [x] 共享分页、排序、结果表格
- [x] 六个 feature API/type/mapper/routes
- [x] 大页面和状态逻辑拆分
- [x] `AnyRecord` 收敛
- [x] legacy CSS 删除、规则复核和按需加载
- [x] 全页面路由懒加载
- [x] 类型、未使用、Lint、格式、单元测试和构建
- [x] 7 组 API 验证
- [x] 61 个路由 Chrome MCP 验证
- [x] 关键无副作用交互验证
- [x] 关键页面结构化 DOM、布局、交互、Console 和 Network 验证
- [x] 同环境性能对比
- [x] `0.0.0.0` 和局域网访问验证
- [x] 进度文档更新
- [x] Git 暂存区更新

计划内剩余事项：无。
