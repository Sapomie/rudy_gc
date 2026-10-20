# API 现状清单

## 1. 清单说明

本清单依据 2026-07-20 当前代码生成，用于描述事实和安排后续验证。路由来源为 `internal/router/router.go`，调用链来源为对应 Handler、Service 和 modelx/repo。

当前清单中的路由统计：

| 方法 | 数量 |
| --- | ---: |
| GET | 73 |
| POST | 68 |
| PUT | 5 |
| DELETE | 7 |
| 合计 | 153 |

## 2. 路由清单

### 2.1 健康、配置与通用操作

| 方法 | 路径 | Handler |
| --- | --- | --- |
| GET | `/api/health` | Router 内联处理 |
| GET | `/api/fetch-sites` | `fetchSiteConfigAPI.List` |
| PUT | `/api/fetch-sites/:site_code` | `fetchSiteConfigAPI.Update` |
| POST | `/api/w-kv/date` | `wKvAPI.UpsertDate` |

### 2.2 电影卡片、详情与人物

| 方法 | 路径 | Handler |
| --- | --- | --- |
| GET | `/api/cards` | `cardsAPI.ListMovieCardFull` |
| GET | `/api/cardstoday` | `cardsAPI.ListMovieCardToday` |
| GET | `/api/cardshasrank` | `cardsAPI.ListMovieCardHasRank` |
| GET | `/api/cardsmediamowned` | `cardsAPI.ListMovieCardMediaOwned` |
| GET | `/api/cardsneeddownload` | `cardsAPI.ListMovieCardNeedDownload` |
| GET | `/api/cardsrandom` | `cardsAPI.ListMovieCardFullRandom` |
| GET | `/api/cardsagg` | `cardsAPI.ListMovieCardAgg` |
| GET | `/api/moviecarddayrank` | `cardsAPI.ListMovieCardDayRank` |
| GET | `/api/moviecardperiodrank` | `cardsAPI.ListMovieCardPeriodRank` |
| GET | `/api/movie/:movie` | `movieDetailAPI.Detail` |
| POST | `/api/movie/:movie/downloadlater` | `movieDetailAPI.AddToDownloadLater` |
| DELETE | `/api/movie/:movie/downloadlater` | `movieDetailAPI.RemoveFromDownloadLater` |
| POST | `/api/movie/:movie/download-cover` | `movieDetailAPI.DownloadCoverNow` |
| POST | `/api/movie/:movie/move-wmedia-removed` | `movieDetailAPI.MoveWMediaToRemoved` |
| PUT | `/api/movie/:movie/torrent-hash` | `movieDetailAPI.UpdateTorrentHash` |
| POST | `/api/movie/:movie/add-cast` | `movieDetailAPI.AddCast` |
| POST | `/api/movie/:movie/album-item` | `movieDetailAPI.AddFetchResourceToAlbum` |
| DELETE | `/api/movie/:movie/album-item` | `movieDetailAPI.RemoveFetchResourceFromAlbum` |
| POST | `/api/movie/:movie/movie-album-item` | `movieDetailAPI.AddMovieToMovieAlbum` |
| POST | `/api/movie-albums` | `movieDetailAPI.CreateMovieAlbum` |
| POST | `/api/open-finder` | `movieDetailAPI.OpenFinder` |
| GET | `/api/casts` | `listPagesAPI.CastList` |
| GET | `/api/cast` | `castPagesAPI.Detail` |
| GET | `/api/cast/:id` | `castPagesAPI.Detail` |
| GET | `/api/persons/:id/merge-candidates` | `personAPI.MergeCandidates` |
| POST | `/api/persons/merge-preview` | `personAPI.MergePreview` |
| POST | `/api/persons/merge` | `personAPI.Merge` |

### 2.3 聚合、目录与媒体查询

| 方法 | 路径 | Handler |
| --- | --- | --- |
| GET | `/api/movie-agg-all/release-bucket-list` | `movieAggAPI.MovieAggAllReleaseBucketList` |
| GET | `/api/w-media-agg/bucket-list` | `wMediaAggAPI.BucketList` |
| GET | `/api/sh-thread-agg/bucket-list` | `shThreadAggAPI.BucketList` |
| GET | `/api/wdir` | `wDirAPI.RootList` |
| GET | `/api/triggers/agg` | `aggTriggerAPI.Page` |
| POST | `/api/agg/w-media/backfill` | `aggTriggerAPI.BackfillWMediaAgg` |
| POST | `/api/agg/movie-release/backfill` | `aggTriggerAPI.BackfillMovieReleaseAgg` |
| POST | `/api/agg/sh-thread/backfill` | `aggTriggerAPI.BackfillSHThreadAgg` |
| GET | `/api/medias` | `listPagesAPI.MediaList` |
| POST | `/api/medias/:id/probe-meta` | `listPagesAPI.ProbeMediaMeta` |
| GET | `/api/crawl-records` | `listPagesAPI.CrawlRecordList` |
| GET | `/api/w-agg-events` | `listPagesAPI.WAggEventList` |

### 2.4 Torrent、电影相册与抓取项

| 方法 | 路径 | Handler |
| --- | --- | --- |
| GET | `/api/torrent-albums` | `listPagesAPI.TorrentAlbums` |
| POST | `/api/albums` | `listPagesAPI.CreateTorrentAlbum` |
| PUT | `/api/albums/:albumID` | `listPagesAPI.UpdateTorrentAlbum` |
| DELETE | `/api/albums/:albumID` | `listPagesAPI.DeleteTorrentAlbum` |
| POST | `/api/albums/:albumID/items/remove` | `listPagesAPI.RemoveTorrentAlbumItem` |
| POST | `/api/albums/:albumID/items/batch-remove` | `listPagesAPI.BatchRemoveTorrentAlbumItems` |
| POST | `/api/albums/:albumID/items/batch-move` | `listPagesAPI.BatchMoveTorrentAlbumItems` |
| GET | `/api/movie-albums` | `listPagesAPI.MovieAlbumItems` |
| PUT | `/api/movie-albums/:albumID` | `listPagesAPI.UpdateMovieAlbum` |
| DELETE | `/api/movie-albums/:albumID` | `listPagesAPI.DeleteMovieAlbum` |
| POST | `/api/movie-albums/:albumID/items/remove` | `listPagesAPI.RemoveMovieAlbumItem` |
| POST | `/api/movie-albums/:albumID/items/batch-move` | `listPagesAPI.BatchMoveMovieAlbumItems` |
| POST | `/api/movie-albums/:albumID/items/batch-copy` | `listPagesAPI.BatchCopyMovieAlbumItems` |
| POST | `/api/movie-albums/:albumID/execute-remove` | `listPagesAPI.ExecuteMovieAlbumRemove` |
| POST | `/api/movie-albums/:albumID/remove-downloaded-items-preview` | `listPagesAPI.PreviewRemoveDownloadedMovieAlbumItems` |
| POST | `/api/movie-albums/:albumID/remove-downloaded-items` | `listPagesAPI.RemoveDownloadedMovieAlbumItems` |
| GET | `/api/d-seeds` | `dSeedAPI.List` |
| POST | `/api/d-seeds` | `dSeedAPI.Create` |
| POST | `/api/d-seeds/:id` | `dSeedAPI.Update` |
| GET | `/api/e-items` | `listPagesAPI.EItemList` |
| POST | `/api/e-items/status/batch` | `eItemAPI.UpdateStatusBatch` |
| POST | `/api/e-items/:id/status` | `eItemAPI.UpdateStatus` |
| DELETE | `/api/e-items/:id` | `eItemAPI.Delete` |

### 2.5 FC2 与色花堂

| 方法 | 路径 | Handler |
| --- | --- | --- |
| POST | `/api/sh-threads/batch-delete` | `shThreadAPI.BatchDelete` |
| GET | `/api/sh-magnets` | `shMagnetAPI.List` |
| POST | `/api/sh-magnets/batch-add-album` | `shMagnetAPI.BatchAddToAlbum` |
| GET | `/api/fc2/movies` | `fc2API.List` |
| GET | `/api/fc2/movies/:fc2_key` | `fc2API.Detail` |
| POST | `/api/fc2/movies/:fc2_key/resource-album-item` | `fc2API.AddResourceToAlbum` |
| DELETE | `/api/fc2/movies/:fc2_key/resource-album-item` | `fc2API.RemoveResourceFromAlbum` |
| POST | `/api/fc2/movies/:fc2_key/movie-album-item` | `fc2API.AddMovieToAlbum` |
| GET | `/api/fc2/cards` | `fc2CardsAPI.List` |
| GET | `/api/fc2/cards/owned` | `fc2CardsAPI.ListOwned` |
| GET | `/api/fc2/cards/random` | `fc2CardsAPI.ListRandom` |
| PUT | `/api/fc2/movies/:fc2_key/images/:image_id/cover` | `fc2API.SetCoverImage` |
| DELETE | `/api/fc2/movies/:fc2_key/images/:image_id` | `fc2API.DeleteMediaImage` |
| GET | `/api/sh-forums` | `shForumAPI.List` |

### 2.6 SC 事件与智能选择

| 方法 | 路径 | Handler |
| --- | --- | --- |
| GET | `/api/sc-events` | `scPagesAPI.ListScEvents` |
| GET | `/api/sc-events-cards` | `scPagesAPI.ListScEventCards` |
| GET | `/api/sc-events/:name` | `scPagesAPI.ScEventDetail` |
| POST | `/api/sc-events/:name` | `scEventAPI.Update` |
| POST | `/api/sc-events/:name/movies/:g_list_id/status` | `scEventAPI.UpdateMovieStatus` |
| GET | `/api/sc-pick-smart-media` | `scPagesAPI.ScPickSmartMediaPage` |
| GET | `/api/triggers/sc` | `scPagesAPI.ScTriggersPage` |
| POST | `/api/triggers/cast/rebuild-rank` | `scTriggerAPI.RebuildActorRank` |
| POST | `/api/triggers/sc/add` | `scTriggerAPI.Add` |
| POST | `/api/triggers/sc/add-preview` | `scTriggerAPI.AddPreview` |
| POST | `/api/triggers/sc/rebuild-stats` | `scTriggerAPI.RebuildStats` |
| POST | `/api/triggers/sc/pick-smart-only` | `scTriggerAPI.PickSmartOnly` |
| POST | `/api/triggers/sc/pick-smart-copy` | `scTriggerAPI.PickSmartCopy` |
| GET | `/api/triggers/sc/copy-status` | `scTriggerAPI.CopyStatus` |
| POST | `/api/triggers/sc/copy-stop` | `scTriggerAPI.CopyStop` |

### 2.7 媒体计划与文件操作

| 方法 | 路径 | Handler |
| --- | --- | --- |
| GET | `/api/triggers/media` | `scPagesAPI.MediaPage` |
| GET | `/api/triggers/fc2-media` | `scPagesAPI.FC2MediaPage` |
| GET | `/api/triggers/media-rescan` | `scPagesAPI.MediaRescanPage` |
| GET | `/api/triggers/media-rollback` | `scPagesAPI.MediaRollbackPage` |
| GET | `/api/triggers/media/plan` | `mediaTriggerAPI.Plan` |
| POST | `/api/triggers/media/precheck` | `mediaTriggerAPI.Precheck` |
| POST | `/api/triggers/media/commit` | `mediaTriggerAPI.Commit` |
| POST | `/api/triggers/media/return` | `mediaTriggerAPI.Return` |
| POST | `/api/triggers/media/rescan` | `mediaTriggerAPI.Rescan` |
| POST | `/api/triggers/media/rollback` | `mediaTriggerAPI.Rollback` |
| GET | `/api/triggers/fc2-media/plan` | `fc2MediaTriggerAPI.Plan` |
| POST | `/api/triggers/fc2-media/precheck` | `fc2MediaTriggerAPI.Precheck` |
| POST | `/api/triggers/fc2-media/commit` | `fc2MediaTriggerAPI.Commit` |
| POST | `/api/triggers/fc2-media/return` | `fc2MediaTriggerAPI.Return` |
| POST | `/api/triggers/fc2-media/default-covers` | `fc2MediaTriggerAPI.BackfillDefaultCovers` |
| GET | `/api/triggers/sc-media` | `scPagesAPI.ScMediaMovePage` |
| GET | `/api/triggers/sc-media/plan` | `scMediaTriggerAPI.Plan` |
| POST | `/api/triggers/sc-media/precheck` | `scMediaTriggerAPI.Precheck` |
| POST | `/api/triggers/sc-media/commit` | `scMediaTriggerAPI.Commit` |
| POST | `/api/triggers/sc-media/return` | `scMediaTriggerAPI.Return` |

### 2.8 抓取页面与兼容触发入口

| 方法 | 路径 | Handler |
| --- | --- | --- |
| GET | `/api/triggers/dailybest` | `crawlerPagesAPI.DailyBestPage` |
| GET | `/api/triggers/seeds` | `crawlerPagesAPI.SeedsPage` |
| GET | `/api/triggers/post-process` | `crawlerPagesAPI.PostProcessPage` |
| GET | `/api/triggers/backfill` | `crawlerPagesAPI.BackfillPage` |
| GET | `/api/triggers/fetch-site` | `crawlerPagesAPI.FetchSitePage` |
| GET | `/api/triggers/fetch-sehuatang-sh` | `crawlerPagesAPI.FetchSehuatangSHPage` |
| GET | `/api/triggers/fetch-sehuatang-sh-thread-list` | `crawlerPagesAPI.FetchSehuatangSHThreadListPage` |
| GET | `/api/triggers/fetch-sehuatang-sh-thread-detail` | `crawlerPagesAPI.FetchSehuatangSHThreadDetailPage` |
| GET | `/api/triggers/fetch-site-javbus-filtered` | `crawlerPagesAPI.FetchSiteJavbusFilteredPage` |
| GET | `/api/triggers/fetch-site-sukebei-filtered` | `crawlerPagesAPI.FetchSiteSukebeiFilteredPage` |
| GET | `/api/fetch-site-javbus-list` | `crawlerPagesAPI.FetchSiteJavbusListPageMain` |
| GET | `/api/fetch-site-sukebei-list` | `crawlerPagesAPI.FetchSiteSukebeiListPageMain` |
| GET | `/api/fetch-site-sehuatang-sh-list` | `crawlerPagesAPI.FetchSiteSehuatangSHListPageMain` |
| GET | `/api/fetch-site-sehuatang-sh-detail/:thread_tid` | `crawlerPagesAPI.FetchSiteSehuatangSHDetailPageMain` |
| POST | `/api/triggers/daily-best` | `crawlerAPI.DailyBest` |
| GET | `/api/dailybest/rank-days` | `dailyBestRankDayAPI.List` |
| POST | `/api/dailybest/rank-days/sync` | `dailyBestRankDayAPI.Sync` |
| POST | `/api/triggers/rebuild-cast-rank` | `crawlerAPI.RebuildCastRank` |
| POST | `/api/triggers/seeds` | `crawlerAPI.Seeds` |
| POST | `/api/triggers/seed-by-name` | `crawlerAPI.SeedByName` |
| POST | `/api/triggers/refresh-oldest-detail` | `crawlerAPI.RefreshOldestDetail` |

### 2.9 Crawler 任务与 SSE

| 方法 | 路径 | Handler |
| --- | --- | --- |
| POST | `/api/crawler/jobs/start` | `crawlerAPI.Start` |
| POST | `/api/crawler/jobs/fetch-site/preview` | `crawlerAPI.PreviewFetchSite` |
| POST | `/api/crawler/jobs/sehuatang-sh-thread-detail/preview` | `crawlerAPI.PreviewSehuatangSHThreadDetails` |
| GET | `/api/crawler/tasks` | `crawlerPagesAPI.TasksPage` |
| GET | `/api/crawler/jobs` | `crawlerAPI.ListJobs` |
| GET | `/api/crawler/jobs/stream` | `crawlerAPI.StreamAll` |
| GET | `/api/crawler/detail-loop/page` | `crawlerPagesAPI.DetailLoopPage` |
| GET | `/api/crawler/detail-loop` | `crawlerAPI.GetDetailLoop` |
| POST | `/api/crawler/detail-loop/start` | `crawlerAPI.StartDetailLoop` |
| POST | `/api/crawler/detail-loop/stop` | `crawlerAPI.StopDetailLoop` |
| GET | `/api/crawler/detail-loop/stream` | `crawlerAPI.StreamDetailLoop` |
| GET | `/api/crawler/jobs/:jobID` | `crawlerAPI.GetJob` |
| GET | `/api/crawler/jobs/:jobID/events` | `crawlerAPI.GetJobEvents` |
| GET | `/api/crawler/jobs/:jobID/stream` | `crawlerAPI.Stream` |
| POST | `/api/crawler/jobs/:jobID/pause` | `crawlerAPI.Pause` |
| POST | `/api/crawler/jobs/:jobID/resume` | `crawlerAPI.Resume` |
| POST | `/api/crawler/jobs/:jobID/stop` | `crawlerAPI.Stop` |

## 3. 主要调用链清单

| 领域 | 入口 | 主要链路 | 数据来源 | 已核实事实 | 待验证风险 |
| --- | --- | --- | --- | --- | --- |
| 电影卡片 | `/api/cards*` | `MovieCardAPI -> movie.Service.ListMovieFull -> MovieListRepoSqlx -> GetMovieType` | MySQL、MovieType Redis、Model Cache | 当前页逐项调用 `GetMovieType` | 缓存异常时 SQL 是否随页内项目放大 |
| 电影详情 | `/api/movie/:movie` | `MovieDetailAPI -> movie.Service -> 多 Model` | MySQL、Redis、文件信息 | 详情组合多个领域数据 | 重复读取、缓存一致性和关联缺失成本 |
| 复杂电影筛选 | `/api/cards*` | `MovieListRepoSqlx` 候选生成、集合运算、排序分页 | 多个单表与现有子查询/连接查询 | `movie_list_query.go` 2083 行 | 大候选集合、深分页和排序成本 |
| 电影聚合 | `/api/cardsagg`、`/api/movie-agg-all/*` | movie Service、moviereleaseagg Service | 源表、聚合结果表 | 同时存在在线聚合与结果表 | 在线扫描和回填一致性 |
| 目录与媒体 | `/api/wdir`、`/api/medias` | wdir/media Service -> WFolder/WMedia Model | MySQL、文件系统 | 页面包含统计与媒体元数据 | 子树统计、文件探测和列表查询成本 |
| 相册 | `/api/torrent-albums`、`/api/movie-albums` | ListPagesAPI -> movie Service -> Album Models | MySQL、Model Cache | 同时支持单项和批量操作 | 批量循环、缓存失效和事务边界 |
| FC2 | `/api/fc2/*` | FC2 Handler -> media/sc Service -> FC2 Models | MySQL、文件系统 | 列表、详情、媒体图片和统计组合 | 排序、详情补全和图片文件一致性 |
| 色花堂 | `/api/sh-*` | SH Handler -> fetchsehuatang/movie Service | MySQL、外部站点 | 磁力、帖子、论坛和抓取运行数据关联 | 列表关联、批量写入和远端失败 |
| SC | `/api/sc-*` | SC Handler -> sc Service -> GList/GSc/Person/Media Models | MySQL、文件系统 | 包含事件、演员、媒体和复制任务 | 候选计算、复制 IO 和写一致性 |
| 聚合回填 | `/api/agg/*` | Trigger -> 聚合 Service | MySQL | 通过显式入口执行回填 | 对在线请求和连接池的影响 |
| 媒体计划 | `/api/triggers/*-media/*` | Handler -> media Service | MySQL、多个媒体目录、计划文件 | 预检与提交分离 | 源变化、冲突和部分失败恢复 |
| Crawler | `/api/crawler/jobs*` | CrawlerAPI -> FetchLoopService -> managed jobs | 内存、MySQL、外部站点 | 支持状态控制、事件和 SSE | 重复任务、资源竞争、退出与订阅释放 |

## 4. 缓存清单

| 缓存 | 配置/实现 | 当前用途 | 已核实行为 | 待验证项 |
| --- | --- | --- | --- | --- |
| go-zero Model Cache | `Config.Cache`、生成 Model 的 `CachedConn` | 主键和唯一键查询缓存 | 多个生成 Model 已接入 | 实际命中率、写失效覆盖、异常回退 |
| MovieType 业务缓存 | `BizRedis`、`MovieTypeCache` | 完整电影卡片对象 | 默认配置含前缀、版本和 TTL；读取失败后构建完整对象 | 热/冷/异常三种成本、批量能力、序列化兼容 |
| Crawler 内存状态 | `managedProgressJobManager` | 任务快照、事件历史、订阅 | 进程内保存并通过 SSE 推送 | 生命周期、容量变化、重启语义和资源释放 |
| 聚合结果表 | release、media、SH bucket/top 等 Model | 页面读取预计算统计 | 已存在回填入口和结果读取 | 数据新鲜度、增量/全量一致性、回填影响 |

## 5. 存量 SQL `JOIN` 清单

静态扫描在以下生产文件中发现 SQL `JOIN`。该清单只说明代码存在，不代表调用频率或性能结论：

- `internal/model/modelx/moviex/a_movie_model.go`
- `internal/model/modelx/moviex/am_cast_model.go`
- `internal/model/modelx/moviex/am_director_model.go`
- `internal/model/modelx/moviex/am_genre_model.go`
- `internal/model/modelx/moviex/am_label_model.go`
- `internal/model/modelx/moviex/am_maker_model.go`
- `internal/model/modelx/moviex/am_prefix_model.go`
- `internal/model/modelx/moviex/c_person_model.go`
- `internal/model/modelx/moviex/c_person_sc_model.go`
- `internal/model/modelx/moviex/c_rank_period_item_model.go`
- `internal/model/modelx/moviex/movie_list_query.go`
- `internal/model/modelx/moviex/sh_magnet_model.go`
- `internal/model/modelx/moviex/w_media_model.go`
- `internal/model/modelx/spiderx/d_seed_model.go`
- `internal/model/modelg/migrations.go`

迁移时需要区分：

- 在线请求 SQL。
- 后台回填 SQL。
- Schema 迁移和历史数据修复 SQL。
- 仅用于测试或注释的文本。

每条在线链路先记录输入、输出、执行计划和耗时，再设计单表候选查询与 Go 集合运算。Schema 中的存量迁移需要单独评估是否仍会在现有数据库执行。

## 6. 错误与契约现状

Handler 当前同时使用多种错误响应形态，包括：

- `{"error": "..."}`
- `{"ok": false, "error": "..."}`
- 直接返回底层错误文本
- Crawler 专用错误写入函数

已核实存在 400、404、500 等状态映射，但同类操作的字段和错误文本不完全一致。后续需要按领域建立契约测试，先记录当前前端依赖，再决定统一范围。

## 7. 任务与 SSE 现状

已核实能力：

- 创建任务并返回 Job ID。
- 查询全部任务、运行中任务、单任务和事件历史。
- 暂停、恢复和停止任务。
- 订阅单任务、全部任务和详情循环事件。
- SSE 初始历史、实时事件和心跳。
- 详情循环单独启动和停止。

待验证：

- 相同业务参数的重复任务行为。
- 服务退出时所有任务和 Goroutine 的收敛。
- 客户端断开后订阅和 Ticker 的释放。
- 任务运行期间数据库连接、外部请求和磁盘 IO 对普通 API 的影响。
- 任务历史和事件在长时间运行下的内存变化。

## 8. 测试与验证现状

| 项目 | 当前结果 |
| --- | --- |
| Go 测试文件 | 34 个 |
| Go Benchmark | 0 个 |
| Router 测试 | 覆盖 DailyBest、Seeds 和 FC2 特定路由约束 |
| 全路由快照 | 未建立 |
| Handler 测试 | 覆盖部分 FC2、SH、聚合和触发接口 |
| Repository 测试 | 覆盖部分 SQL 构造与事务行为 |
| Service 测试 | 覆盖部分媒体、SC、相册、SH 和目录逻辑 |
| 真实接口性能基线 | 未建立 |
| 真实数据库执行计划清单 | 未建立 |
| 缓存异常测试 | 未建立完整矩阵 |
| 2026-07-16 全量测试 | `go test -count=1 ./...` 通过 |
| 2026-07-16 静态检查 | `go vet ./...` 通过 |
| 2026-07-16 构建 | `go build ./cmd/api` 通过 |

## 9. 基线采集队列

以下队列仅表示需要采集，不代表已经确认性能问题：

| 范围 | 采集重点 | 进入优化的依据 |
| --- | --- | --- |
| 电影卡片页面 | SQL 随页大小变化、MovieType 缓存三种状态、响应组装 | 用户等待时间和查询放大证据 |
| 电影详情 | 多 Model 读取、缓存重复、关联缺失 | 请求耗时与重复读取证据 |
| 相册页面 | 候选集合、排序、批量操作 | SQL、候选规模和事务证据 |
| FC2 列表与详情 | 排序、媒体补全、图片信息 | SQL、响应大小和页面耗时 |
| SC 事件 | 事件、演员、媒体关联 | 候选规模、SQL 和组装成本 |
| 媒体与目录 | 目录统计、媒体分页、文件探测 | 数据库与文件 IO 证据 |
| 聚合页面 | 在线计算与结果表读取 | 源表扫描和聚合读取对比 |
| Crawler | 轮询、SSE、任务资源使用 | 请求频率、连接和资源竞争证据 |

清单在路由、调用链、缓存或数据访问方式变化后必须同步更新。
