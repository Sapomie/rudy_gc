# w_media / w_folder 来源类型说明

当前系统中，`w_media` 与 `w_folder` 都按 `source_type` 区分新旧两套数据。

## 1. w_media

`w_media` 同时保存两类媒体记录：

- `source_type = 1`
  含义：旧 `v_film` 迁移过来的 legacy 媒体
- `source_type = 2`
  含义：新版原生 `w_media` 媒体

说明：

- 同一个 `movie_jav_id`，允许同时存在两条记录：
  - 一条 `source_type = 1`
  - 一条 `source_type = 2`
- 当前唯一性按 `(movie_jav_id, source_type)`、`(movie_name, source_type)`、`(file_name, source_type)` 约束。
- 业务查询时必须明确来源类型，不能默认混查。

## 2. w_folder

`w_folder` 同时保存两套目录树：

- `source_type = 1`
  含义：旧 `v_film` 对应的 legacy 目录树
- `source_type = 2`
  含义：新版 `w_media` 对应的 native 目录树

说明：

- legacy 目录树与 native 目录树彼此隔离。
- 当前唯一性按 `(parent_id, name, source_type)`、`(path, source_type)` 约束。
- 不能把 `source_type = 1` 的媒体挂到 `source_type = 2` 的目录上。
- 不能把 `source_type = 2` 的媒体挂到 `source_type = 1` 的目录上。

## 3. 关联规则

`w_media` 与 `w_folder` 的关联必须同时满足：

- `w_media.directory_id -> w_folder.id`
- `w_media.source_type = w_folder.source_type`

也就是说：

- legacy 媒体只关联 legacy 目录
- native 媒体只关联 native 目录

如果某条 `w_media` 当前不属于受管理目录范围，则允许：

- `directory_id = 0`

## 4. native 目录范围

对 `source_type = 2` 的 native 数据，目录树只管理以下两棵子树：

- `root_dir/media`
- `root_dir/watched`

因此：

- `w_folder(source_type = 2)` 只应包含 `media`、`watched` 及其祖先目录。
- `001_process`、`005_removed` 等路径不应进入 native 目录树。
- 如果某条 native `w_media` 不在 `media/watched` 下，则 `directory_id` 应为 `0`。

## 5. legacy 目录语义

legacy 目录现在不再是假定“只有一个根目录”。

当前语义是：

- `/dir`
  展示全部 legacy 根目录列表
- `/dir/:id`
  展示某个 legacy 目录详情
- `/dir/root`
  仅作为兼容入口，跳转到 `/dir`

这意味着：

- legacy 目录树允许多根
- 不能再依赖固定目录 id，例如 `/dir/1`

## 6. native 目录语义

native 目录页与 legacy 目录页分开：

- `/wdir`
  展示 native 根目录列表
- `/wdir/:id`
  展示某个 native 目录详情

这意味着：

- 不能再依赖固定目录 id，例如 `/wdir/1`
- 菜单和面包屑都必须走根列表入口，而不是写死某个目录 id

## 7. 开发约束

后续开发必须遵守：

- 查询 `w_media` 时明确 `source_type`
- 查询 `w_folder` 时明确 `source_type`
- 目录筛选、目录聚合、目录详情都不能混用新旧来源
- 任何目录重建、迁移、回填逻辑都必须按 `source_type` 分开处理
- native 目录树重建时，只允许纳入 `media/watched` 两棵子树

## 8. 迁移后的理解方式

现在不能再把系统理解成：

- 旧视频在 `v_film`
- 新视频在 `w_media`

正确理解应该是：

- 旧视频在 `w_media(source_type = 1)`
- 新视频在 `w_media(source_type = 2)`
- 旧目录在 `w_folder(source_type = 1)`
- 新目录在 `w_folder(source_type = 2)`

也就是说，表已经统一，但业务语义仍然分成新旧两套来源。
