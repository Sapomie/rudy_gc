# FC2 接入 SC / GList 完整实施方案

## 一、最终结论

- `w_media`、`fc2_media` 都保持媒体表职责，不增加 SC 字段。
- `g_sc_stat` 独立保存 JAV 影片 SC 统计；新增 `fc2_sc_stat` 独立保存 FC2 影片 SC 统计。
- `fc2_movie` 的 SC 统计字段数量为 0。
- `g_list` 增加影片类型和 FC2 影片标识，支持 JAV、FC2 混合事件。
- 所有本次新增和改造的 SC/FC2 查询采用单表查询、批量查询、Go 内存归并，SQL 中禁止 `JOIN` 和关联子查询。
- SC 新增、Come 片修改、影片状态修改必须事务落库。
- FC2 主资料、媒体状态、SC 写入、状态修改和全量回填都会同步更新 `fc2_sc_stat`。

## 二、代码现状与问题

### 2.1 WMedia 的 SC 字段来源

`internal/model/modelg/w_media.go` 中的 `w_media` 物理表当前没有 SC 字段。

`internal/model/modelx/moviex/w_media_model.go` 中的 `NativeMediaListRow` 返回了：

- `sc_times`
- `come_times`
- `last_sc_time`

这三个字段当前通过 `w_media LEFT JOIN g_sc_stat` 取得。零 JOIN 改造后仍由 `g_sc_stat` 提供，查询改为分别读取 `w_media`、`g_sc_stat`，再在 Go 中合并，`w_media` 表结构保持不变。

### 2.2 GScStat 的实际职责

`g_sc_stat` 当前保存：

- `sc_event_times`
- `sc_times`
- `come_times`
- `last_sc_time`
- `releasing_date`
- `media_birth_time`

其中 `sc_event_times` 是 JAV 影片完整统计的一部分；`releasing_date`、`media_birth_time` 是现有列表排序快照。

### 2.3 FC2 缺少独立 SC 统计表

`fc2_movie`、`fc2_media` 当前都没有 SC 字段，也没有与 `g_sc_stat` 对应的 FC2 统计表。

FC2 需要新增独立 `fc2_sc_stat`，保持影片主数据、媒体数据、SC 统计三类职责分离。

### 2.4 GList 当前只能表示 JAV

`g_list` 当前只有 `movie_jav_id`，所有新增、状态修改、Come Movie 修改、详情加载逻辑都把它作为唯一影片标识，无法可靠表示 FC2。

### 2.5 SC 新增存在部分落库风险

`internal/service/sc/ad__add_sc.go` 当前执行顺序为：

1. 循环逐条写入 `g_list`。
2. 复制事件图片。
3. 写入 `g_sc`。
4. 更新 SC 统计。

任意中间步骤失败都可能留下部分 `g_list` 或孤立图片，需要改成事务写入并补充文件清理。

### 2.6 FC2 当前存在关联子查询

`internal/model/modelx/moviex/fc2_movie_model.go` 当前存在：

- 查询 `metadata_thread_tid` 的关联子查询。
- 按播放状态排序的 `EXISTS` 关联子查询。
- 按下载时间排序的标量关联子查询。

这些查询全部需要消除。

## 三、数据库结构方案

### 3.1 g_list

增加字段：

```sql
movie_type varchar(16) NOT NULL DEFAULT 'jav'
fc2_movie_id bigint NOT NULL DEFAULT 0
```

字段规则：

| movie_type | movie_jav_id | fc2_movie_id |
|---|---|---:|
| `jav` | 必须有值 | 0 |
| `fc2ppv` | 空字符串 | 必须大于 0 |

增加索引：

```sql
INDEX idx_g_list_sc_name (sc_name, is_sc, is_come, id)
INDEX idx_g_list_jav_sc (movie_type, movie_jav_id, is_sc, sc_name)
INDEX idx_g_list_fc2_sc (movie_type, fc2_movie_id, is_sc, sc_name)
UNIQUE INDEX uk_g_list_event_movie
    (sc_name, movie_type, movie_jav_id, fc2_movie_id)
```

历史数据使用单表更新：

```sql
UPDATE g_list
SET movie_type = 'jav', fc2_movie_id = 0
WHERE TRIM(movie_jav_id) <> '';
```

创建唯一索引前先执行重复数据检查。检查发现重复时终止迁移并输出重复键，不执行静默删除或合并。

### 3.2 w_media

`w_media` 不增加 SC 字段，继续只保存媒体数据。

WMedia 普通列表继续单表查询 `w_media`。涉及 SC 展示、筛选或排序时：

1. 单表查询 `g_sc_stat`，使用其中的统计字段和排序快照确定影片范围与顺序。
2. 根据 `movie_jav_id` 单表批量查询 `w_media`。
3. Go 中按 `movie_jav_id` 合并，并保持 `g_sc_stat` 返回顺序。

`g_sc_stat.media_birth_time` 必须在 JAV 媒体入库、移动、移除、恢复时同步更新，使 SC 筛选和媒体存在性判断不依赖 JOIN。

`ListScMissingNativeMedia` 直接查询 `g_sc_stat.media_birth_time=0`，并由媒体写入流程保证该快照准确。

### 3.3 fc2_media

`fc2_media` 不增加 SC 字段，继续只保存 FC2 媒体、文件路径、视频参数和来源信息。

FC2 SC 统计统一写入 `fc2_sc_stat`。FC2 页面需要同时展示媒体和 SC 信息时，分别批量查询两个表后在 Go 中合并。

### 3.4 fc2_sc_stat

新增表：

```sql
CREATE TABLE fc2_sc_stat (
    id bigint NOT NULL AUTO_INCREMENT,
    fc2_movie_id bigint NOT NULL,
    fc2_key varchar(64) NOT NULL DEFAULT '',
    movie_name varchar(500) NOT NULL DEFAULT '',
    sc_event_times mediumint NOT NULL DEFAULT 0,
    sc_times mediumint NOT NULL DEFAULT 0,
    come_times mediumint NOT NULL DEFAULT 0,
    last_sc_time bigint NOT NULL DEFAULT 0,
    post_time bigint NOT NULL DEFAULT 0,
    media_birth_time bigint NOT NULL DEFAULT 0,
    created_on bigint NOT NULL DEFAULT 0,
    updated_on bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_fc2_sc_stat_movie_id (fc2_movie_id),
    UNIQUE INDEX uk_fc2_sc_stat_key (fc2_key),
    INDEX idx_fc2ss_sc_last_name (sc_times DESC, last_sc_time DESC, movie_name DESC),
    INDEX idx_fc2ss_come_last_name (come_times DESC, last_sc_time DESC, movie_name DESC),
    INDEX idx_fc2ss_last_name (last_sc_time DESC, movie_name DESC),
    INDEX idx_fc2ss_post_name (post_time DESC, movie_name DESC),
    INDEX idx_fc2ss_media_birth_name (media_birth_time DESC, movie_name DESC)
);
```

字段职责：

- `sc_event_times`：该 FC2 影片出现在 SC 事件中的总次数，包含 `is_sc=1`。
- `sc_times`：该 FC2 影片 `is_sc=2` 的事件次数。
- `come_times`：上述事件中 `is_come=2` 的次数。
- `last_sc_time`：`is_sc=2` 事件最大的 `g_sc.sc_time`。
- `post_time`：对应 `fc2_movie.post_time` 的排序快照。
- `media_birth_time`：对应有效 `fc2_media.birth_time` 的排序快照；无有效媒体时为 0。

`last_sc_event_time` 继续属于演员、人物统计，不进入 FC2 影片统计。

### 3.5 fc2_movie

FC2 的 SC 统计字段数量为 0。

为了消除现有播放状态和下载时间排序的关联子查询，增加两个列表排序快照字段：

```sql
media_present tinyint NOT NULL DEFAULT 0
media_birth_time bigint NOT NULL DEFAULT 0
```

这两个字段属于 FC2 列表排序快照，不属于 SC 统计。

`metadata_thread_tid` 改成批量单表读取 `sh_thread` 后在 Go 中合并，不再通过关联子查询生成。

## 四、数据库迁移与回填

### 4.1 迁移文件

修改：

- `internal/model/modelg/g_list.go`
- 新增 `internal/model/modelg/fc2_sc_stat.go`
- `internal/model/modelg/fc2_movie.go`
- `internal/model/modelg/a__init.go`

迁移必须幂等，使用现有 `hasTable`、`hasColumn`、`hasIndex` 检查方式。

### 4.2 迁移顺序

1. 增加 `g_list.movie_type`、`g_list.fc2_movie_id`。
2. 单表回填历史 `g_list` 为 `jav`。
3. 检查事件影片重复数据。
4. 创建 `g_list` 查询索引和唯一索引。
5. 创建 `fc2_sc_stat` 及其唯一索引、排序索引。
6. 增加 `fc2_movie.media_present`、`media_birth_time`。
7. 通过 Go 回填任务重建 `g_sc_stat`、`fc2_sc_stat` 和 FC2 媒体排序快照。

迁移 SQL 禁止 `JOIN`。

### 4.3 FC2 媒体快照回填

回填过程：

1. 单表执行 `UPDATE fc2_movie SET media_present=0, media_birth_time=0`。
2. 流式读取 `fc2_media`。
3. 按 `fc2_movie_id` 单表更新对应 `fc2_movie`。

不设置新的行数上限、默认截断或静默裁剪。

### 4.4 SC 统计回填

统一回填流程：

1. 单表读取 `g_list`。
2. 收集全部 `sc_name`。
3. 单表批量读取 `g_sc`。
4. 在 Go 中按影片类型和影片标识聚合。
5. JAV 更新 `g_sc_stat`。
6. FC2 更新或插入 `fc2_sc_stat`。

聚合结果为空时明确写入 0，保证历史脏统计能够清零。

## 五、SC Preview 识别流程

### 5.1 批量识别顺序

扫描目录后先收集全部视频文件，再按以下顺序识别：

1. 按完整文件名批量查询 `fc2_media.file_name`。
2. 完整文件名未命中的文件，解析编码前缀得到 FC2 Key，再批量查询 `fc2_media.fc2_key`。
3. 剩余文件继续执行 JAV `w_media.movie_name` 识别。
4. 同一文件出现多种有效归属时直接报错，不自动猜测。
5. FC2 Preview 分别批量读取 `fc2_movie`、`fc2_media`，在 Go 中合并。

FC2 实际落盘文件名示例：

```text
15122252531-2711719-1_sub_self_nomsk.mp4
```

编码前缀包含 FC2 前缀、番号和 `code_seq`，解析结果对应 `FC2PPV-2711719-1`。

### 5.2 Preview 数据结构

```json
{
  "movieType": "jav | fc2ppv",
  "movieJavId": "",
  "fc2MovieId": 0,
  "movieKey": "",
  "movieName": "",
  "title": "",
  "movieHref": "",
  "jacketImg": "",
  "casts": [],
  "isSc": 2
}
```

FC2 Preview 规则：

- `movieHref=/fc2/{FC2Key}`。
- 当前没有结构化演员关系，`casts=[]`。
- FC2 选作 Come Movie 时，Lead Cast 必须为空。

### 5.3 已发现的历史阈值

`internal/service/sc/ad__add_sc.go` 当前要求 MP4 文件 `size >= 20000`。

本方案保留该历史行为并增加测试，不新增任何阈值、上限或默认截断。

## 六、SC 落库流程

### 6.1 新增事件

1. 后端重新扫描目录，禁止直接相信前端 Preview。
2. 验证影片归属、Come Movie、SC 状态和 Lead Cast。
3. 将事件图片准备到临时目标。
4. 在同一个数据库事务内：
   - 插入一条 `g_sc`。
   - 批量插入全部 `g_list`。
   - 校验 Come Movie 最多一条。
   - 更新 `g_sc.movie_number`。
   - 更新 `g_sc.come_movie_name`。
5. 事务失败时回滚全部数据库记录，并清理已准备图片。
6. 事务成功后统一清除 `g_sc`、`g_list` 相关缓存。
7. 按影片类型执行统计重建：
   - JAV：更新 `g_sc_stat`、演员和人物统计。
   - FC2：更新 `fc2_sc_stat`。

当前逐条调用 `glUpsert` 的新增方式移除，新增专用 `ScEventWriteModel` 负责事务和缓存失效。

### 6.2 AddSc 请求结构

每个影片显式携带：

```json
{
  "movieType": "jav | fc2ppv",
  "movieJavId": "",
  "fc2MovieId": 0,
  "isSc": 2
}
```

Come Movie 同样使用类型化引用。后端重新扫描后校验请求中的影片引用确实存在于本次 Preview。

## 七、事件修改流程

### 7.1 修改影片 SC 状态

路由改为：

```text
POST /api/sc-events/:name/movies/:g_list_id/status
```

使用 `g_list.id` 定位，消除 JAV ID 与 FC2 ID 的歧义。

事务内执行：

1. 锁定目标 `g_list`。
2. 校验该记录属于当前事件。
3. 更新 `is_sc`。
4. 取消 SC 时同步清理该行 `is_come`。
5. 必要时清理 `g_sc.come_movie_name` 和 `movie_cast`。
6. 重新计算并写入 `g_sc.movie_number`。

事务提交后根据 `movie_type` 重建对应统计。

### 7.2 修改 Come Movie

请求字段由 `comeMovieJavId` 改成 `comeGListId`。

事务内统一更新同一事件全部 `g_list.is_come`，并更新 `g_sc`。

FC2 Come Movie 的 Lead Cast 必须为空；JAV Come Movie 继续按当前影片演员列表验证。

## 八、统计更新时机

统计更新必须覆盖以下时机：

1. SC 事件新增成功后。
2. `is_sc` 修改成功后。
3. Come Movie 修改成功后。
4. JAV 媒体入库后。
5. FC2 媒体入库后。
6. FC2 媒体删除后重新入库时。
7. FC2 主资料的标题、发布时间更新后。
8. 全量 SC 回填任务执行时。

FC2 媒体入库成功后立即执行：

```text
rebuildFC2ScStatsByMovieIDs(fc2MovieID)
syncFC2ScStatMediaSnapshot(fc2MovieID)
syncFC2MovieMediaSortSnapshot(fc2MovieID)
```

FC2 媒体移除时将 `fc2_sc_stat.media_birth_time` 更新为 0；重新入库时写入新的 `birth_time`。FC2 主资料更新时同步更新已有 `fc2_sc_stat.movie_name`、`post_time`，全程使用按 `fc2_movie_id` 的单表更新。

`RebuildAllScStats` 改为读取全部存在事件关系的影片，包括 `is_sc=1` 的记录，保证已经全部取消 SC 的影片统计能够清零。

## 九、缓存方案

### 9.1 JAV

- `g_sc_stat` 保留当前按 `id`、`movie_jav_id` 的缓存。
- 更新 `g_sc_stat` 后继续失效 MovieType 缓存。
- JAV 媒体写入流程同步更新 `g_sc_stat.media_birth_time`。

### 9.2 FC2

新增 `FC2ScStatModel`，使用与 `GScStatModel` 一致的缓存模式：

- 按 `id` 缓存。
- 按 `fc2_movie_id` 建立唯一索引缓存。
- 按 `fc2_key` 建立唯一索引缓存。
- Insert、Update、Delete 清除全部相关缓存键。
- FC2 页面列表通过一次 `ListByMovieIDs` 单表批量查询读取统计，避免逐行缓存查询。

`FC2MovieModel`、`FC2MediaModel` 保持现有单表批量查询模式；本方案不依赖把 SC 数据写进两个模型的缓存对象。

## 十、严格零 JOIN 查询方案

### 10.1 WMedia

- WMedia 与 `g_sc_stat` 分别单表查询，在 Go 中按 `movie_jav_id` 合并。
- SC 筛选和排序由 `g_sc_stat` 完成，媒体详情由 `w_media` 批量补齐。
- `g_sc_stat.media_birth_time` 作为媒体存在和下载时间快照，媒体写入链路负责同步。

### 10.2 FC2 列表

- `fc2_movie` 负责主列表分页和排序。
- `media_present`、`media_birth_time` 负责播放状态和下载时间排序。
- 当前页的 `fc2_media` 通过 `ListByMovieIDs` 单表批量查询。
- 当前页的 `fc2_sc_stat` 通过 `ListByMovieIDs` 单表批量查询。
- `metadata_thread_tid` 通过 `sh_thread.ListByIDs` 单表批量查询。
- 所有结果在 Handler 中使用 map 合并。

### 10.3 SC 统计

- 查询 `g_list`。
- 查询 `g_sc`。
- Go 聚合。
- JAV 更新 `g_sc_stat`。
- FC2 更新 `fc2_sc_stat`。

### 10.4 验收规则

- 新增和修改的 SC/FC2 SQL 中禁止出现 `JOIN`。
- 禁止通过 `EXISTS (SELECT ...)`、标量子查询或视图隐藏关联查询。
- 单表批量查询优先于逐行查询。
- 单表结果统一在 Go 中归并。

## 十一、其他调用链隔离

### 11.1 SC Media Move

`internal/service/media/60_sc_media_move.go` 当前只支持 `w_media`。

查询 `g_list` 时显式限定 `movie_type='jav'`，避免 FC2 记录进入 WMedia 移动流程。FC2 媒体继续保留在 FC2 独立目录体系。

### 11.2 JAV 影片删除

`internal/service/movie/20_movie_delete_relations.go` 查询和删除 `g_list` 时显式限定 `movie_type='jav'`，保证 JAV 删除流程不会处理 FC2 记录。

### 11.3 演员与人物统计

FC2 当前没有结构化演员关系。FC2 SC 写入不会进入 `am_cast`、`c_person` 统计流程。

## 十二、前端页面方案

### 12.1 ScTriggersPage.vue

- Preview 卡片增加 `JAV` / `FC2` 来源标记。
- 使用类型化影片标识。
- Come Movie 使用类型化引用。
- FC2 链接进入 `/fc2/{FC2Key}`。
- FC2 的 Lead Cast 为空并禁用。
- 提交 payload 携带 `movieType`、`movieJavId`、`fc2MovieId`。

### 12.2 ScEventsPage.vue

- 影片状态按钮使用 `g_list.id`。
- Come Movie 下拉框使用 `g_list.id`。
- SC 事件详情增加 FC2 卡片。
- JAV 卡片继续使用现有 MovieType。
- FC2 卡片展示 FC2 Key、标题、出演文本、播放按钮、下载时间和三个 SC 统计。
- Come Movie 链接按来源生成 `/moviedetail/...` 或 `/fc2/...`。

### 12.3 FC2Page.vue

增加：

- SC 次数。
- Come 次数。
- 最近 SC 时间。

后端分别批量查询 `fc2_media`、`fc2_sc_stat`，在 Go 中按 `fc2_movie_id` 合并后返回页面。

### 12.4 FC2DetailPage.vue

- 媒体资料区域增加三个 SC 指标。
- 增加进入最近 SC Event 的入口。
- 最近事件名通过独立 `g_list` 单表查询获得。
- 三个 SC 指标读取 `fc2_sc_stat`，媒体参数继续读取 `fc2_media`。

## 十三、修改文件清单

### 后端模型与迁移

- `v3/gc-api/internal/model/modelg/g_list.go`
- 新增 `v3/gc-api/internal/model/modelg/fc2_sc_stat.go`
- `v3/gc-api/internal/model/modelg/fc2_movie.go`
- `v3/gc-api/internal/model/modelg/a__init.go`
- `v3/gc-api/internal/model/modelx/moviex/g_list_model.go`
- `v3/gc-api/internal/model/modelx/moviex/g_list_model_gen.go`
- `v3/gc-api/internal/model/modelx/moviex/w_media_model.go`
- `v3/gc-api/internal/model/modelx/moviex/fc2_media_model.go`
- `v3/gc-api/internal/model/modelx/moviex/fc2_movie_model.go`
- 新增 `v3/gc-api/internal/model/modelx/moviex/fc2_sc_stat_model.go`
- 新增 `v3/gc-api/internal/model/modelx/moviex/fc2_sc_stat_model_gen.go`
- 新增 `v3/gc-api/internal/model/modelx/moviex/sc_event_write_model.go`
- `v3/gc-api/internal/svc/deps.go`

### 后端业务

- `v3/gc-api/internal/types/glist.go`
- `v3/gc-api/internal/types/sc_event.go`
- `v3/gc-api/internal/service/sc/ad__add_sc.go`
- `v3/gc-api/internal/service/sc/10_update_event.go`
- `v3/gc-api/internal/service/sc/20_store.go`
- `v3/gc-api/internal/service/sc/event_pages.go`
- `v3/gc-api/internal/service/sc/uc__update_sc_movie_cast_info.go`
- `v3/gc-api/internal/service/loop/10_jobs.go`
- `v3/gc-api/internal/service/media/73_fc2_ingest_commit.go`
- `v3/gc-api/internal/service/media/60_sc_media_move.go`
- `v3/gc-api/internal/service/movie/20_movie_delete_relations.go`

### 后端接口

- `v3/gc-api/internal/router/handler/trigger_sc_api.go`
- `v3/gc-api/internal/router/handler/sc_event_api.go`
- `v3/gc-api/internal/router/handler/sc_pages_api.go`
- `v3/gc-api/internal/router/handler/fc2_api.go`
- `v3/gc-api/internal/router/router.go`

### 前端

- `v3/gc-web-next02/src/pages/ScTriggersPage.vue`
- `v3/gc-web-next02/src/pages/ScEventsPage.vue`
- `v3/gc-web-next02/src/pages/FC2Page.vue`
- `v3/gc-web-next02/src/pages/FC2DetailPage.vue`
- `v3/gc-web-next02/src/styles/index.css`

### 测试

- `v3/gc-api/internal/service/sc/ad__add_sc_test.go`
- 新增 SC 事务写入测试。
- 新增 FC2 SC 聚合与回填测试。
- 新增 GList 类型隔离测试。
- `v3/gc-api/internal/router/handler/fc2_api_test.go`
- 新增零 JOIN SQL 构建测试。
- 前端补充 FC2 Preview、事件状态修改和 FC2 SC 展示测试。

## 十四、验证标准

### 14.1 后端验证

- 混合 JAV、FC2 的 Preview 能正确识别。
- FC2 完整文件名和编码前缀两种识别方式均通过。
- 同一文件归属冲突时返回明确错误。
- SC 新增任一步失败时，`g_sc` 与 `g_list` 均无部分数据。
- JAV 与 FC2 状态修改均通过 `g_list.id`。
- FC2 Come Movie 可保存，Lead Cast 保持为空。
- `fc2_sc_stat` 四个 SC 统计字段在新增、修改、重新入库、全量回填后保持一致。
- 已全部取消 SC 的影片统计能够归零。
- WMedia 的 SC 信息来自 `g_sc_stat`；FC2 的 SC 信息来自 `fc2_sc_stat`。
- `w_media`、`fc2_media` 数据库结构中没有新增 SC 字段。
- FC2 列表原有三个关联子查询全部清除。
- 相关 SQL 测试明确断言不存在 `JOIN` 和关联子查询。
- 执行 `make test-api` 和 `make build-api`。

### 14.2 前端验证

- 执行 `make check-web-next02`。
- 使用实际页面反复验证：
  - `/scadd` Preview。
  - 混合 JAV/FC2 提交。
  - `/sc-events/:name` 详情。
  - Come Movie 修改。
  - JAV/FC2 影片状态切换。
  - `/fclist` 三个 SC 字段展示。
  - `/fc2/:fc2_key` 三个 SC 字段展示。
- 检查网络请求、接口响应、页面交互和浏览器控制台错误。
- 至少完整执行两轮新增、修改、刷新后的页面回归验证。

## 十五、实施顺序

1. 完成数据库字段、索引和幂等迁移。
2. 完成 `fc2_sc_stat` 模型、批量单表查询和缓存失效。
3. 消除 WMedia 和 FC2 列表中的 SC/媒体关联查询，改为多次单表查询后 Go 合并。
4. 完成 GList 类型化模型和事务写模型。
5. 完成混合 JAV/FC2 Preview。
6. 完成 SC 新增事务流程。
7. 完成状态修改和 Come Movie 修改事务流程。
8. 完成 JAV、FC2 统计同步和全量回填。
9. 隔离 SC Media Move、JAV 删除和演员统计调用链。
10. 完成四个前端页面。
11. 完成后端测试、构建和前端检查。
12. 使用实际页面反复执行完整回归验证。
