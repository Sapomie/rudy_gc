# v3 legacy 拆分迁移记录

## 当前状态

- 已创建 `v3/` 项目骨架。
- 已将 `legacy/backend/` 机械复制到 `v3/gc-api/`。
- 已将 `legacy/backend/ui/` 机械复制到 `v3/gc-web/legacy-ui/`。
- 已将 `legacy/oldmodel/` 机械复制到 `v3/oldmodel/`，保持 `gc-api/oldmodel -> ../oldmodel` 软链依赖。
- 已将 `legacy/z_text/` 机械复制到 `v3/z_text/`，并补充 `gc-api/z_text -> ../z_text` 运行入口。
- 已将 `legacy/deploy/` 机械复制到 `v3/deploy/legacy-deploy/`。
- 当前后端 router 已切换为 API-only：`v3/gc-api/internal/router/router.go` 不再加载模板、不注册 HTML 页面路由、不提供后端 UI 静态资源。
- 当前前端 server 承担 UI 和静态资源入口：`/cards`、`/static/*`、`/src/*`、`/Volumes/*`、`/text/*` 均由 `v3/gc-web` 处理或代理。

## 迁移原则

- 每个页面先读取 legacy router、handler、service、model、SQL、模板和真实点击行为。
- 后端迁移只做页面 router 到 API 输出的拆分。
- 前端迁移只做 legacy 模板和静态脚本到独立前端的拆分。
- 页面业务逻辑、参数语义、默认值、排序、分页、筛选、链接、tab、下钻行为保持 legacy。
- 已拆分页面必须保留 legacy 菜单入口；尚未拆分页面不能回退到后端 HTML 页面。
- 后端 router 只允许注册 `/api/...`；UI 页面、前端静态资源和本地媒体文件访问全部归 `gc-web`。

## 页面迁移登记

| legacy 路由 | legacy 后端入口 | legacy 模板 | v3 API | v3 前端 | 状态 |
| --- | --- | --- | --- | --- | --- |
| `/cards` | `MovieHTMLHandler.ListMovieCardFull` | `page.list_movie_card` | `/api/cards` | `/cards` | done |
| `/cardstoday` | `MovieHTMLHandler.ListMovieCardToday` | `page.list_movie_card` | `/api/cardstoday` | `/cardstoday` | done |
| `/cardshasrank` | `MovieHTMLHandler.ListMovieCardHasRank` | `page.list_movie_card` | `/api/cardshasrank` | `/cardshasrank` | done |
| `/cardsmediamowned` | `MovieHTMLHandler.ListMovieCardMediaOwned` | `page.list_movie_card` | `/api/cardsmediamowned` | `/cardsmediamowned` | done |
| `/cardsneeddownload` | `MovieHTMLHandler.ListMovieCardNeedDownload` | `page.list_movie_card` | `/api/cardsneeddownload` | `/cardsneeddownload` | done |
| `/cardsrandom` | `MovieHTMLHandler.ListMovieCardFullRandom` | `page.list_movie_card` | `/api/cardsrandom` | `/cardsrandom` | done |
| `/moviecarddayrank` | `MovieHTMLHandler.ListMovieCardDayRank` | `page.list_movie_card_rank_day` | `/api/moviecarddayrank` | `/moviecarddayrank` | done |
| `/moviecardperiodrank` | `MovieHTMLHandler.ListMovieCardPeriodRank` | `page.list_movie_card_rank_period` | `/api/moviecardperiodrank` | `/moviecardperiodrank` | done |
| `/movie/:movie` | `MovieHTMLHandler.MovieDetail` | `page.movie_detail` | 待拆分 | 待拆分 | todo |
| `/wdir` | `WDirectoryHTMLHandler.RootList` | `page.wdir_list` / `page.wdir_detail` | `/api/wdir` | `/wdir` | done |
| `/w-media-agg/birth` | `WMediaAggHTMLHandler.BirthYears` | `page.w_media_agg_birth_time` | `/api/w-media-agg/birth` | `/w-media-agg/birth` | done |
| `/w-media-agg/bucket-list` | `WMediaAggHTMLHandler.BucketList` | `page.w_media_birth_bucket_list` | `/api/w-media-agg/bucket-list` | `/w-media-agg/bucket-list` | done |
| `/movie-agg-all/release` | `MovieAggHTMLHandler.MovieAggAllReleaseYears` | `page.movie_agg_all_time` | `/api/movie-agg-all/release` | `/movie-agg-all/release` | done |
| `/movie-agg-all/release-bucket-list` | `MovieAggHTMLHandler.MovieAggAllReleaseBucketList` | `page.movie_release_bucket_list` | `/api/movie-agg-all/release-bucket-list` | `/movie-agg-all/release-bucket-list` | done |
| `/triggers/agg` | `CrawlerPages.AggTriggersPage` | `page.agg_triggers` | `/api/triggers/agg` | `/triggers/agg` | done |
| `/triggers/dailybest` | `CrawlerPages.DailyBestPage` | `page.admin_triggers` | `/api/triggers/dailybest` | `/triggers/dailybest` | done |
| `/triggers/seeds` | `CrawlerPages.SeedsPage` | `page.admin_triggers` | `/api/triggers/seeds` | `/triggers/seeds` | done |
| `/triggers/post-process` | `CrawlerPages.PostProcessPage` | `page.admin_triggers` | `/api/triggers/post-process` | `/triggers/post-process` | done |
| `/triggers/backfill` | `CrawlerPages.BackfillPage` | `page.admin_triggers` | `/api/triggers/backfill` | `/triggers/backfill` | done |
| `/crawler/tasks` | `CrawlerPages.TasksPage` | `page.crawler_tasks` | `/api/crawler/tasks` | `/crawler/tasks` | done |
| `/crawler/detail-loop` | `CrawlerPages.DetailLoopPage` | `page.crawler_detail_loop` | `/api/crawler/detail-loop/page` | `/crawler/detail-loop` | done |
| `/medias` | `MovieHTMLHandler.MediaListPage` | `page.w_media_list` | `/api/medias` | `/medias` | done |
| `/w-agg-events` | `MovieHTMLHandler.WAggEventListPage` | `page.w_agg_event_list` | `/api/w-agg-events` | `/w-agg-events` | done |
| `/torrent-albums` | `MovieHTMLHandler.TorrentAlbumsPage` | `page.torrent_album_list` | `/api/torrent-albums` | `/torrent-albums` | done |
| `/movie-albums` | `MovieHTMLHandler.MovieAlbumItemsPage` | `page.movie_album_items_list` | `/api/movie-albums` | `/movie-albums` | done |
| `/d-seeds` | `CrawlerPages.DSeedListPage` | `page.d_seed_list` | `/api/d-seeds` | `/d-seeds` | done |
| `/casts` | `MovieHTMLHandler.CastListPage` | `page.cast_list` | `/api/casts` | `/casts` | done |
| `/e-items` | `MovieHTMLHandler.EItemListPage` | `page.e_item_list` | `/api/e-items` | `/e-items` | done |
| `/sc-events` | `MovieHTMLHandler.ListScEventsPage` | `page.sc_event_list` | `/api/sc-events` | `/sc-events` | done |
| `/sc-events-cards` | `MovieHTMLHandler.ListScEventsCardPage` | `page.sc_event_cards` | `/api/sc-events-cards` | `/sc-events-cards` | done |
| `/sc-events/:name` | `MovieHTMLHandler.ScEventDetailPage` | `page.sc_event_detail` | `/api/sc-events/:name` | `/sc-events/:name` | done |
| `/sc-pick-smart-media` | `MovieHTMLHandler.ScPickSmartMediaPage` | `page.sc_pick_smart` | `/api/sc-pick-smart-media` | `/sc-pick-smart-media` | done |
| `/triggers/sc` | `MovieHTMLHandler.ScTriggersPage` | `page.sc_triggers` | `/api/triggers/sc` | `/triggers/sc` | done |
| `/triggers/sc-media` | `CrawlerPages.ScMediaMovePage` | `page.sc_media_move` | `/api/triggers/sc-media` | `/triggers/sc-media` | done |

## 验收记录

- 当前只完成 v3 初始机械迁移骨架。
- `/cards` 家族已拆分：
  - v3 API：`/api/cards`、`/api/cardstoday`、`/api/cardshasrank`、`/api/cardsmediamowned`、`/api/cardsneeddownload`、`/api/cardsrandom`、`/api/moviecarddayrank`、`/api/moviecardperiodrank`。
  - v3 前端：`/cards`、`/cardstoday`、`/cardshasrank`、`/cardsmediamowned`、`/cardsneeddownload`、`/cardsrandom`、`/moviecarddayrank`、`/moviecardperiodrank`。
  - 后端继续调用 legacy 的 `parseMovieCardRequest`、`loadMovieCardPageData`、`movieSvc.ListMovieFull`。
  - `/cardsrandom` 继续调用 legacy 的 `movieSvc.ListMovieFullRandom` 和随机数量 `n` 逻辑。
  - `/moviecarddayrank` 继续调用 legacy 的 rank day 查询、日期归一、前一天/后一天/随机日期和分页逻辑。
  - `/moviecardperiodrank` 继续调用 legacy 的周期榜 `BuildRankPeriodPage`、周期切换、分类切换、上一期/下一期/最新和周期卡片数据结构。
  - API handler 内使用 legacy 前端路径构造分页、排序、筛选、榜单导航链接，避免生成 `/api/...` 链接。
  - 浏览器复查截图：`/Users/gaojinwei/src/my/rudy_gc/v3/tmp/cards/v3-cards-ps2.png`。
  - 本轮只确认 v3 `/cards?ps=2` 可渲染、API 200、图片请求 200、菜单存在、主要链接保持 `/cards` 语义。
  - 后端 UI router 删除后，`/api/cards` 仍由 API-only router 提供；图片资源改由 `gc-web` 读取本地 `/Volumes/*`。
- 其他页面 router 到 API、模板到独立前端的逐页拆分尚未执行。
- cards 家族 legacy/v3 截图逐项对比、后端输出逐字段对比、SQL/性能对比尚未执行，未标记验收过关。
- MovieAgg 菜单页已拆分：
  - v3 API：`/api/movie-agg-all/release`、`/api/movie-agg-all/release-bucket-list`、`/api/w-media-agg/birth`、`/api/w-media-agg/bucket-list`、`/api/triggers/agg`、`/api/wdir`。
  - v3 前端：`/movie-agg-all/release`、`/movie-agg-all/release-bucket-list`、`/w-media-agg/birth`、`/w-media-agg/bucket-list`、`/triggers/agg`、`/wdir`。
  - 后端继续调用 legacy 的 `moviereleaseagg`、`wmediaagg`、`wdir`、`movieSvc.ListMovieFull`、聚合回填 service。
  - API handler 内使用 legacy 前端路径构造分页、排序、层级、模式切换、目录下钻和聚合下钻链接，避免生成 `/api/...` 链接。
  - MovieAgg 菜单页 legacy/v3 截图逐项对比、后端输出逐字段对比、SQL/性能对比尚未执行，未标记验收过关。
- List 和 SC 菜单页已拆分：
  - v3 List API：`/api/medias`、`/api/w-agg-events`、`/api/torrent-albums`、`/api/movie-albums`、`/api/d-seeds`、`/api/casts`、`/api/e-items`。
  - v3 SC API：`/api/sc-events`、`/api/sc-events-cards`、`/api/sc-events/:name`、`/api/sc-pick-smart-media`、`/api/triggers/sc`、`/api/triggers/sc-media`。
  - v3 SC 动作 API：`/api/triggers/sc/add`、`/api/triggers/sc/add-preview`、`/api/triggers/sc/rebuild-stats`、`/api/triggers/sc/pick-smart-only`、`/api/triggers/sc/pick-smart-copy`、`/api/triggers/sc/copy-status`、`/api/triggers/sc/copy-stop`、`/api/triggers/sc-media/plan`、`/api/triggers/sc-media/precheck`、`/api/triggers/sc-media/commit`、`/api/triggers/sc-media/return`。
  - v3 前端：List 分组使用 `list.html`/`src/list.js`，SC 分组使用 `sc.html`/`src/sc.js`。
  - 后端继续沿用 legacy 的参数解析、默认值、筛选、排序、分页、数据组装、DSeed 创建/更新、SC 预览/挑片/移动接口语义。
  - API handler 内使用 legacy 前端路径构造分页、排序、视图切换、详情和下钻链接，避免生成 `/api/...` 链接。
  - List 和 SC 菜单页 legacy/v3 截图逐项对比、后端输出逐字段对比、SQL/性能对比尚未执行，未标记验收过关。
- DailyBest 任务页已拆分：
  - v3 API：`/api/triggers/dailybest`、`/api/crawler/jobs`、`/api/crawler/jobs/start`、`/api/crawler/jobs/stream`、`/api/crawler/jobs/:jobID/events`、`/api/crawler/jobs/:jobID/stream`、`/api/crawler/jobs/:jobID/pause`、`/api/crawler/jobs/:jobID/resume`、`/api/crawler/jobs/:jobID/stop`。
  - v3 前端：`/triggers/dailybest` 使用 `triggers.html`/`src/crawler-jobs.js`，继续复用 legacy `crawler_jobs.js` 与 `crawler_jobs.css`。
  - 后端继续沿用 legacy 的 `loop.StartTask`、`StartDailyBest`、`StartDailyBestSync`、任务列表、任务事件和暂停/继续/停止语义。
  - 页面数据保持 legacy `DailyBestPage` 的标题、按钮、默认任务、允许任务、阶段概览和 `job_id` 语义。
  - 本轮已通过 `go test ./internal/router/... ./internal/service/loop/...`、`npm run check`，并用临时 15012/15013 验证 `/triggers/dailybest` 页面、静态资源、API 请求和事件流均可加载。
- Seeds 任务页已拆分：
  - v3 API：`/api/triggers/seeds`，继续复用 `/api/crawler/jobs`、`/api/crawler/jobs/start`、`/api/crawler/jobs/stream`、`/api/crawler/jobs/:jobID/events`、`/api/crawler/jobs/:jobID/stream`、`/api/crawler/jobs/:jobID/pause`、`/api/crawler/jobs/:jobID/resume`、`/api/crawler/jobs/:jobID/stop`。
  - v3 前端：`/triggers/seeds` 使用 `triggers.html`/`src/crawler-jobs.js`，继续复用 legacy `crawler_jobs.js` 与 `crawler_jobs.css`。
  - 后端继续沿用 legacy 的 `loop.StartTask`、`StartSeeds`、`StartSeedByName`、`StartRefreshOldestDetail`、任务列表、任务事件和暂停/继续/停止语义。
  - 页面数据保持 legacy `SeedsPage` 的标题、按钮、默认任务、允许任务、`seeds_stages` 阶段概览和 `job_id` 语义。
- PostProcess / Backfill / DetailLoop / CrawlerTasks 已拆分：
  - v3 API：`/api/triggers/post-process`、`/api/triggers/backfill`、`/api/crawler/tasks`、`/api/crawler/detail-loop/page`。
  - v3 前端：`/triggers/post-process`、`/triggers/backfill` 继续使用 `triggers.html`/`src/crawler-jobs.js`；`/crawler/tasks` 使用 `crawler-tasks.html`/`src/crawler-tasks.js`；`/crawler/detail-loop` 使用 `crawler-detail-loop.html`/`src/crawler-detail-loop.js`。
  - 后端继续沿用 legacy 的 `loop.StartTask`、后处理、回填、任务列表、详情循环 API 与事件流语义。
  - 页面数据保持 legacy 的标题、按钮、默认任务、允许任务、概览模式、`job_id` 和 `apiBase` 语义。
