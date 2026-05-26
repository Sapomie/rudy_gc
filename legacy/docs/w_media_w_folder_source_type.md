# w_media / w_folder 来源类型说明

当前系统只保留新版原生媒体来源。

## 1. source_type

- `source_type = 2`
  含义：新版原生 `w_media` / `w_folder` 数据。

`source_type = 1` 是旧 `v_film` 来源，已进入清理阶段：

- 业务代码不得再新增 `source_type = 1` 的读取、写入、筛选、聚合和页面入口。
- 已迁出的电影应通过 `c_movie_album` / `c_movie_album_item` 保留人工相册关系。
- 存量 `source_type = 1` 数据清理后，不再作为业务查询来源。

## 2. w_media

`w_media` 保存原生媒体记录。

说明：

- 业务查询时必须明确使用原生来源。
- 电影与媒体的关联以 `movie_jav_id` 为内部稳定关联字段。
- 页面展示与跳转仍以 `a_movie.name` 为用户可见番号。

## 3. w_folder

`w_folder` 保存原生目录树。

说明：

- 当前唯一性按 `(parent_id, name, source_type)`、`(path, source_type)` 约束。
- `w_media.directory_id -> w_folder.id` 时，两侧必须同为原生来源。
- 如果某条 `w_media` 当前不属于受管理目录范围，则允许 `directory_id = 0`。

## 4. native 目录范围

对原生数据，目录树只管理以下两棵子树：

- `root_dir/media`
- `root_dir/watched`

因此：

- `w_folder` 只应包含 `media`、`watched` 及其祖先目录。
- `001_process`、`005_removed` 等路径不应进入 native 目录树。
- 如果某条 native `w_media` 不在 `media/watched` 下，则 `directory_id` 应为 `0`。

## 5. native 目录页

native 目录页入口：

- `/wdir`
  展示 native 根目录列表
- `/wdir/:id`
  展示某个 native 目录详情

这意味着：

- 不能依赖固定目录 id，例如 `/wdir/1`。
- 菜单和面包屑都必须走根列表入口，而不是写死某个目录 id。

## 6. 开发约束

后续开发必须遵守：

- 查询 `w_media` 时明确使用原生来源。
- 查询 `w_folder` 时明确使用原生来源。
- 目录筛选、目录聚合、目录详情都不得重新引入旧来源。
- 任何目录重建、迁移、回填逻辑都只维护原生来源。
- native 目录树重建时，只允许纳入 `media/watched` 两棵子树。
