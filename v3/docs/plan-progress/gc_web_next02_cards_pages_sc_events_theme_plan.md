# gc-web-next02 Cards 页面 `/sc-events` 粉色主题改造计划

## 1. 计划目标

本计划用于将 `v3/gc-web-next02` 中 `cardsmediamowned` 和其他 Cards 相关页面，统一改造成当前 `/sc-events` 的粉色主题和紧凑页面规格。

改造以 `/cardsmediamowned` 作为 Cards 主页面样本，同时覆盖 Cards 聚合页和排名页。

改造目标：

- Cards 页面与 `/sc-events` 使用同一粉色主题。
- 顶部区域、筛选、排序、统计和分页使用统一布局语言。
- MovieCard 保留封面浏览优势和原有业务状态。
- 页面入口统一在新窗口打开。
- 不改变 Cards 页面现有 API、参数、字段、状态和持久化行为。

## 2. 规范来源

- `v3/docs/standard/gc_web_next02_frontend_style_standard.md`
- 当前 `/sc-events` 页面实际实现
- 当前 Cards 页面已有业务结构和交互

发生冲突时：

- 用户当前明确要求优先。
- 业务结构和参数语义保持现有实现。
- 视觉和交互默认服从前端样式规范。

## 3. 修改范围

### 3.1 页面文件

- `v3/gc-web-next02/src/pages/CardsPage.vue`
- `v3/gc-web-next02/src/pages/CardsAggPage.vue`
- `v3/gc-web-next02/src/pages/RankPage.vue`

### 3.2 共享组件

- `v3/gc-web-next02/src/components/MovieCard.vue`
- `v3/gc-web-next02/src/components/MovieGrid.vue`
- `v3/gc-web-next02/src/components/MovieCardFilterForm.vue`
- `v3/gc-web-next02/src/components/LinkStrip.vue`
- `v3/gc-web-next02/src/components/PaginationBar.vue`
- `v3/gc-web-next02/src/components/PageShell.vue` 只在确有共享需求时调整，禁止影响非 Cards 页面。

### 3.3 样式文件

- `v3/gc-web-next02/src/styles/index.css`

### 3.4 版本边界

- 只修改 `v3/gc-web-next02`。
- 禁止修改其他 next 版本。
- 禁止修改后端，除非后续用户明确提出新的业务功能。

## 4. 路由范围

### 4.1 Cards 主页面

| 状态 | 路由 | 页面组件 | 主要特点 |
|---|---|---|---|
| 待处理 | `/cards` | `CardsPage.vue` | 通用 Cards 入口 |
| 待处理 | `/cardstoday` | `CardsPage.vue` | 今日卡片、日期编辑 |
| 待处理 | `/cardsmediamowned` | `CardsPage.vue` | MediaOwned 筛选样本页 |
| 待处理 | `/cardshasrank` | `CardsPage.vue` | 排名筛选 |
| 待处理 | `/cardsneeddownload` | `CardsPage.vue` | 待下载状态 |
| 待处理 | `/cardsrandom` | `CardsPage.vue` | 随机查询 |

### 4.2 Cards 扩展页面

| 状态 | 路由 | 页面组件 | 主要特点 |
|---|---|---|---|
| 待处理 | `/cardsagg` | `CardsAggPage.vue` | 聚合统计与多组表格 |
| 待处理 | `/moviecarddayrank` | `RankPage.vue` | 日期榜单与 MovieGrid |
| 待处理 | `/moviecardperiodrank` | `RankPage.vue` | 周期榜单与 MovieGrid |

## 5. 明确排除项

- `/sc-events-cards` 使用独立方案：
  - `v3/docs/plan-progress/gc_web_next02_sc_events_cards_rewrite_plan.md`
- WDir 页面中的 MovieCard 网格默认不属于本计划主体范围。
- Movie 详情页中的相关卡片不属于本计划主体范围。
- 其他页面偶发使用 MovieCard 时，只在共享组件改动不可避免地生效且不破坏原页面时保留。

## 6. 样本页面

### 6.1 样本路由

- `/cardsmediamowned`

### 6.2 样本职责

先在 `/cardsmediamowned` 完成：

- 粉色主题作用域
- 紧凑顶部
- 排序与 Owned 控制栏
- Total、Current Page 和分页
- 筛选面板
- MovieGrid
- MovieCard
- 页面入口新窗口规则
- 响应式和 reduced-motion

样本完成后，再同步到其他 Cards 路由。

禁止在样本尚未稳定时对所有路由进行无差别批量覆盖。

## 7. 默认主题

| 用途 | 色值 |
|---|---|
| 主粉色 | `#d94f8a` |
| 深粉链接 | `#9f2f63` |
| 主文字 | `#4b2939` |
| 次要文字 | `#8b6577` |
| 主浅粉背景 | `#fff0f6` |
| 次浅粉背景 | `#fff5fa` |
| 卡片背景 | `#fffafd` |
| hover 背景 | `#fde8f2` |
| 主边框 | `#eed5e1` |
| 细分隔线 | `#f1dce6` |

业务状态色可以保留：

- 成功
- 警告
- 错误
- 已拥有
- 待下载
- 删除或不可用

状态色必须表达真实业务含义，禁止仅为视觉统一强制全部改成粉色。

## 8. 共享主题作用域

### 8.1 建议类名

```text
next02-pink-cards-page
next02-pink-cards-hero
next02-pink-cards-toolbar
next02-pink-cards-filter
next02-pink-movie-grid
next02-pink-movie-card
next02-pink-cards-metrics
next02-pink-cards-pagination
```

### 8.2 作用域原则

- Cards 公共视觉写入共享类。
- 路由特有行为继续使用现有 route class。
- 禁止给每个 Cards 路由复制整套相同 CSS。
- 禁止用无作用域全局选择器覆盖非 Cards 页面。
- 共享组件样式必须检查 WDir、Cast、Movie 等其他调用方。

## 9. 页面结构标准

Cards 主页面改造后的模块顺序：

1. 紧凑顶部区域
2. Cards 控制栏
3. 可展开筛选面板
4. MovieGrid
5. 日期编辑弹窗或其他页面特有弹窗
6. 空状态或错误状态

禁止新增重复统计卡、重复说明和装饰性大模块。

## 10. 紧凑顶部区域

### 10.1 内容

左侧：

- `MovieCard`
- 当前页面标题
- 简短说明或当前日期

右侧：

- 当前页面已有主要入口
- 页面特有操作

### 10.2 尺寸

| 项目 | 规格 |
|---|---|
| 上下内边距 | `9px` |
| 左右内边距 | `14px` |
| 标题字号 | `18px` |
| 说明字号 | `12px` |
| 默认布局 | 桌面单行 |

- 空间不足时说明文字单行省略。
- 窄屏允许换行。
- 禁止用大标题和大面积上下留白拉高页面。

## 11. Cards 控制栏

### 11.1 左侧区域

默认顺序：

1. 排序
2. WMedia 标识
3. Owned 状态
4. 筛选按钮
5. 页面特有控制

### 11.2 右侧区域

默认顺序：

1. `Total`
2. `Current Page`
3. 分页

### 11.3 尺寸与样式

- 排序按钮、Owned 按钮、筛选按钮、统计胶囊和分页按钮默认高度 `30px`。
- 普通按钮使用浅粉背景、粉色边框和深粉文字。
- 当前选中项使用主粉色背景和白色文字。
- hover 使用 `#fff0f6` 或 `#fde8f2`。
- 控制栏空间不足时允许换行。
- 窄屏下分页保持在统计信息之后。

## 12. 排序与 Owned

### 12.1 排序

- 保留现有排序字段、名称、Href 和 active 状态。
- 排序操作继续在当前窗口执行。
- 当前排序项使用主粉色实心状态。
- 其他排序项使用浅粉次按钮。

### 12.2 Owned

- 保留 `MAll`、`MOwned`、`MNotOwned` 的现有参数语义。
- `/cardsmediamowned` 不得改变默认 MediaOwned 条件。
- 当前 Owned 状态使用主粉色或明确业务状态色。
- Owned 切换继续在当前窗口执行。

## 13. 筛选面板

### 13.1 结构

- 保留 `MovieCardFilterForm.vue` 当前字段、参数和提交方式。
- 保留展开与收起。
- 保留筛选和清空。
- 保留页面特有隐藏参数。

### 13.2 样式

- 面板使用白色到浅粉背景。
- 边框使用 `#eed5e1`。
- 圆角 `10px`。
- 输入框、下拉框和按钮高度保持紧凑。
- focus 使用粉色边框和粉色 focus ring。
- 筛选按钮使用主粉色。
- 清空使用浅粉背景和深粉文字。

### 13.3 行为

- 筛选、清空和字段切换继续在当前窗口执行。
- 不修改现有查询参数名称和默认值。
- 不新增默认截断或结果上限。

## 14. MovieGrid

### 14.1 响应式列数

保持当前 Cards 网格适配逻辑，并统一视觉间距：

| 可用宽度 | 默认列数 |
|---|---:|
| 超宽屏 | 6 列 |
| 大屏 | 5 列 |
| 中屏 | 3 至 4 列 |
| 小屏 | 2 列 |
| 极窄屏 | 1 列 |

如当前页面业务明确使用不同列数，保留当前行为并只统一间距和卡片外观。

### 14.2 网格规则

- 卡片间距统一。
- 同一行卡片高度一致。
- 页面整体禁止横向滚动。
- 禁止为了统一视觉改变海报宽高比。

## 15. MovieCard 外观

### 15.1 卡片容器

- 背景使用白色到 `#fffafd` 渐变。
- 边框使用 `1px solid #eed5e1`。
- 圆角使用 `10px`。
- 阴影使用低透明粉色阴影。
- overflow 与海报圆角匹配。

### 15.2 hover

- 上浮 `2px`。
- 边框轻微增强为粉色。
- 阴影轻微增强。
- 动画时长 `150ms` 至 `180ms`。
- hover 不得改变卡片尺寸和网格位置。

### 15.3 海报

- 保留现有 2:3 比例。
- 保留图片来源、懒加载和解码方式。
- 保留空封面逻辑。
- 空封面使用浅粉背景和深粉文字。
- 禁止改变裁切逻辑。

## 16. MovieCard 信息层级

保持现有字段，不删除业务信息。

建议视觉层级：

1. 影片编号
2. 影片标题
3. 主要状态
4. 演员、类型和标签
5. 下载或 Media 状态
6. 排名、观看、SC 和 Come 信息
7. 外部站点入口

### 16.1 文本

- 影片编号使用深粉色和高字重。
- 标题使用主文字色。
- 次要信息使用 `#8b6577`。
- 长标题和长标签单行省略。
- 禁止单条长内容拉高整张卡片。

### 16.2 标签

- 普通标签使用浅粉背景和粉色边框。
- 业务状态标签保留真实状态色。
- 标签高度和间距保持紧凑。
- 禁止无业务意义地增加多种颜色。

### 16.3 数字

- 排名、观看人数、SC 次数和 Come 次数使用 `tabular-nums`。
- 主要数字可使用深粉色。
- 禁止用大尺寸统计块占用卡片高度。

## 17. 页面入口行为

### 17.1 新窗口打开

以下入口必须在新窗口打开：

- Movie 详情
- Cast 详情
- 导演、片商、标签等进入其他页面的入口
- 其他 Cards 页面入口
- 聚合页进入 Cards 的入口
- 外部站点入口

统一设置：

```html
target="_blank"
rel="noopener noreferrer"
```

### 17.2 当前窗口操作

以下操作继续在当前窗口执行：

- 筛选
- 排序
- Owned 切换
- 分页
- 日期切换
- 周期切换
- 类型和分类切换
- 筛选面板展开与收起
- 日期编辑弹窗

## 18. `/cardstoday` 特殊规则

- 保留日期编辑按钮。
- 保留日期编辑弹窗。
- 保留“今天”快捷按钮。
- 保留当前 item key、保存接口和状态反馈。
- 弹窗使用粉色主题、粉色 focus 和紧凑按钮。
- 禁止改变日期格式和保存 payload。

## 19. `/cardsmediamowned` 特殊规则

- 作为主样本页面优先改造。
- 保留 MediaOwned 默认语义。
- 保留 WMedia、MAll、MOwned 和 MNotOwned。
- 保留所有排序项。
- 保留筛选面板字段和分页。
- 检查卡片中的 Media 状态与播放入口。

## 20. `/cardshasrank` 特殊规则

- 保留排名筛选和排序参数。
- 排名信息保持明显，但使用克制的粉色强调。
- 禁止改变最高排名、在榜天数和排名日期语义。

## 21. `/cardsneeddownload` 特殊规则

- 保留待下载条件和状态。
- 待下载、已拥有和不可用状态继续使用真实状态色。
- 禁止因主题改造弱化关键下载状态。

## 22. `/cardsrandom` 特殊规则

- 保留随机查询行为。
- 保留当前随机参数和刷新语义。
- 视觉改造不得引入固定排序或缓存结果展示。

## 23. 排名页面

### 23.1 `/moviecarddayrank`

- 保留日期输入。
- 保留前一天、后一天和随机日期。
- 保留当前日期显示和 Total。
- 保留分页。
- 顶部、控制栏、按钮和 MovieGrid 改为粉色主题。

### 23.2 `/moviecardperiodrank`

- 保留周期类型。
- 保留分类。
- 保留周期 key、日期范围、上一期、下一期和最新入口。
- 保留分页。
- 顶部、控制栏、按钮和 MovieGrid 改为粉色主题。

### 23.3 排名页当前窗口行为

- 日期、周期、类型、分类和分页继续在当前窗口执行。
- 卡片进入详情继续在新窗口打开。

## 24. `/cardsagg` 特殊规则

### 24.1 页面顶部

- 使用同一紧凑粉色顶部。
- 保留 Cards 入口和筛选按钮。
- 保留 Total。

### 24.2 聚合表格

聚合表格按照当前 `/sc-events` 的列表规格改造：

| 项目 | 规格 |
|---|---|
| 表头高度 | `30px` |
| 数据行高度 | `28px` |
| 字号 | `13px` |
| 内边距 | `1px 10px 2px` |
| 表头背景 | `#fcecf4` |
| hover 背景 | `#fde8f2` |
| 边框 | `#f1dce6` |

- 每张聚合表按自身字段规划列宽。
- 禁止复制 `/sc-events` 的具体列宽比例。
- 长文本单行省略。
- 页面整体禁止横向滚动。

### 24.3 聚合业务

- 保留 Skipped 状态。
- 保留分组、时间线、分布和 Top Movie 数据。
- 保留当前过滤参数和跳过条件。
- 禁止修改聚合计算和后端限制。

## 25. 空状态、错误状态和加载状态

- 保留现有 LoadingState 和 ErrorPanel。
- 空状态使用浅粉背景、粉色虚线边框和次要文字。
- 错误状态必须展示真实错误。
- 禁止增加大面积插画和空白。
- 状态切换不得改变页面整体结构。

## 26. 动画

- 页面整体淡入 `280ms`。
- 顶部和控制栏允许轻量上浮。
- 卡片允许小幅顺序淡入。
- hover 控制在 `150ms` 至 `180ms`。
- 弹窗使用轻量淡入和小幅缩放。
- 必须提供 `prefers-reduced-motion: reduce` 覆盖。

## 27. 响应式规则

### 27.1 桌面端

- 顶部和控制栏优先单行。
- 排序、Owned 和筛选靠左。
- Total、Current Page 和分页靠右。
- MovieGrid 使用多列布局。

### 27.2 中等宽度

- 控制栏允许换行。
- 排序保持在分页之前。
- 页面整体不得横向滚动。

### 27.3 移动端

- 顶部允许换行。
- 控制栏纵向排列。
- 筛选控件占满可用宽度。
- MovieGrid 使用 1 至 2 列。
- 弹窗适配可用宽度。

## 28. 实施阶段

### 阶段一：记录 `/cardsmediamowned` 基准

- 记录字段、筛选、排序、Owned、分页和卡片入口。
- 记录当前 MovieGrid 和 MovieCard 结构。
- 记录其他页面对共享组件的使用。

### 阶段二：建立 Cards 公共主题

- 增加 Cards 页面独立粉色作用域。
- 建立顶部、控制栏、筛选、统计、分页和卡片变量。
- 确保不影响非 Cards 页面。

### 阶段三：改造 `/cardsmediamowned`

- 重写顶部。
- 重写控制栏。
- 重写筛选面板。
- 重写 MovieGrid 和 MovieCard。
- 补齐新窗口属性。

### 阶段四：同步 Cards 主路由

- `/cards`
- `/cardstoday`
- `/cardshasrank`
- `/cardsneeddownload`
- `/cardsrandom`

每个路由单独检查特殊行为。

### 阶段五：改造排名页面

- `/moviecarddayrank`
- `/moviecardperiodrank`

保留全部日期和周期控制。

### 阶段六：改造 `/cardsagg`

- 重写顶部和控制栏。
- 按列表规格改造聚合表格。
- 保留聚合分组和跳过状态。

### 阶段七：响应式和 reduced-motion

- 检查桌面、中屏和移动端。
- 检查网格列数和控制栏换行。
- 补齐 reduced-motion。

### 阶段八：全路由检查

- 检查全部 Cards 路由。
- 检查共享组件的其他调用方。
- 检查页面入口与当前窗口操作边界。

### 阶段九：构建与文档

- 更新相关进度记录。
- 执行 `git diff --check`。
- 执行 `npm run build`。

## 29. 单页检查模板

```md
### 页面：/cards-example

- 页面组件：
- 共享组件：
- 排序参数：
- Owned 参数：
- 筛选参数：
- 分页参数：
- 日期或周期控制：
- 页面入口：
- 特殊状态：
- 桌面布局：
- 移动端布局：
- 构建结果：
- 当前状态：待处理 / 代码完成 / 构建通过 / 交互检查通过
```

## 30. 验收矩阵

每个 Cards 路由必须记录：

| 项目 | 状态 |
|---|---|
| 粉色主题生效 | 待检查 |
| 顶部紧凑 | 待检查 |
| 排序靠左 | 待检查 |
| Owned 靠左 | 待检查 |
| Total 与 Current Page 靠右 | 待检查 |
| 分页最右 | 待检查 |
| 筛选功能正常 | 待检查 |
| MovieGrid 正常 | 待检查 |
| MovieCard 字段完整 | 待检查 |
| 页面入口新窗口打开 | 待检查 |
| 当前页操作留在当前窗口 | 待检查 |
| 页面无横向滚动 | 待检查 |
| 移动端正常 | 待检查 |
| 构建通过 | 待检查 |

## 31. 默认禁止项

- 禁止删除 MovieCard 字段。
- 禁止改变海报比例和裁切逻辑。
- 禁止改变 API 和查询参数。
- 禁止改变 Owned、Rank、NeedDownload 和 Random 语义。
- 禁止改变日期编辑和周期导航行为。
- 禁止为每个路由复制整套样式。
- 禁止无作用域修改 MovieCard 并误伤其他页面。
- 禁止新增重复统计和大面积留白。
- 禁止将业务状态全部改成粉色。
- 禁止修改其他 next 版本。
- 禁止遗漏新窗口安全属性。

## 32. 总体完成标准

只有以下条件全部满足，才允许记录本计划完成：

- `/cardsmediamowned` 样本完成。
- Cards 主页面全部完成。
- 日榜和周期榜全部完成。
- `/cardsagg` 完成。
- 所有页面使用同一粉色主题。
- 所有页面保留原筛选、排序、Owned、分页和特殊控制。
- MovieCard 字段和状态完整。
- 页面入口新窗口行为完整。
- 当前页面操作仍在当前窗口执行。
- 页面整体无横向滚动。
- 共享组件没有误伤范围外页面。
- reduced-motion 完整。
- `git diff --check` 通过。
- `npm run build` 通过。
