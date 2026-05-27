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
| `/wdir` | `WDirectoryHTMLHandler.RootList` | `page.wdir_list` | 待拆分 | 待拆分 | todo |
| `/w-media-agg/birth` | `WMediaAggHTMLHandler.BirthYears` | `page.w_media_agg_birth_time` | 待拆分 | 待拆分 | todo |
| `/w-media-agg/bucket-list` | `WMediaAggHTMLHandler.BucketList` | `page.w_media_birth_bucket_list` | 待拆分 | 待拆分 | todo |
| `/movie-agg-all/release` | `MovieAggHTMLHandler.MovieAggAllReleaseYears` | `page.movie_agg_all_time` | 待拆分 | 待拆分 | todo |
| `/movie-agg-all/release-bucket-list` | `MovieAggHTMLHandler.MovieAggAllReleaseBucketList` | `page.movie_release_bucket_list` | 待拆分 | 待拆分 | todo |

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
