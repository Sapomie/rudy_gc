# v3 legacy 拆分迁移记录

## 当前状态

- 已创建 `v3/` 项目骨架。
- 已将 `legacy/backend/` 机械复制到 `v3/gc-api/`。
- 已将 `legacy/backend/ui/` 机械复制到 `v3/gc-web/legacy-ui/`。
- 已将 `legacy/oldmodel/` 机械复制到 `v3/oldmodel/`，保持 `gc-api/oldmodel -> ../oldmodel` 软链依赖。
- 已将 `legacy/z_text/` 机械复制到 `v3/z_text/`，并补充 `gc-api/z_text -> ../z_text` 运行入口。
- 已将 `legacy/deploy/` 机械复制到 `v3/deploy/legacy-deploy/`。

## 迁移原则

- 每个页面先读取 legacy router、handler、service、model、SQL、模板和真实点击行为。
- 后端迁移只做页面 router 到 API 输出的拆分。
- 前端迁移只做 legacy 模板和静态脚本到独立前端的拆分。
- 页面业务逻辑、参数语义、默认值、排序、分页、筛选、链接、tab、下钻行为保持 legacy。
- 已拆分页面必须保留 legacy 菜单入口；尚未拆分页面的前端路由可临时代理到 v3 API 的 legacy HTML 页面，避免导航断链。

## 页面迁移登记

| legacy 路由 | legacy 后端入口 | legacy 模板 | v3 API | v3 前端 | 状态 |
| --- | --- | --- | --- | --- | --- |
| `/cards` | `MovieHTMLHandler.ListMovieCardFull` | `page.list_movie_card` | `/api/cards` | `/cards` | in_progress |
| `/cardstoday` | `MovieHTMLHandler.ListMovieCardToday` | `page.list_movie_card` | 待拆分 | 待拆分 | todo |
| `/cardshasrank` | `MovieHTMLHandler.ListMovieCardHasRank` | `page.list_movie_card` | 待拆分 | 待拆分 | todo |
| `/cardsmediamowned` | `MovieHTMLHandler.ListMovieCardMediaOwned` | `page.list_movie_card` | 待拆分 | 待拆分 | todo |
| `/cardsneeddownload` | `MovieHTMLHandler.ListMovieCardNeedDownload` | `page.list_movie_card` | 待拆分 | 待拆分 | todo |
| `/moviecarddayrank` | `MovieHTMLHandler.ListMovieCardDayRank` | `page.list_movie_card_rank_day` | 待拆分 | 待拆分 | todo |
| `/moviecardperiodrank` | `MovieHTMLHandler.ListMovieCardPeriodRank` | `page.list_movie_card_rank_period` | 待拆分 | 待拆分 | todo |
| `/movie/:movie` | `MovieHTMLHandler.MovieDetail` | `page.movie_detail` | 待拆分 | 待拆分 | todo |
| `/wdir` | `WDirectoryHTMLHandler.RootList` | `page.wdir_list` | 待拆分 | 待拆分 | todo |
| `/w-media-agg/birth` | `WMediaAggHTMLHandler.BirthYears` | `page.w_media_agg_birth_time` | 待拆分 | 待拆分 | todo |
| `/w-media-agg/bucket-list` | `WMediaAggHTMLHandler.BucketList` | `page.w_media_birth_bucket_list` | 待拆分 | 待拆分 | todo |
| `/movie-agg-all/release` | `MovieAggHTMLHandler.MovieAggAllReleaseYears` | `page.movie_agg_all_time` | 待拆分 | 待拆分 | todo |
| `/movie-agg-all/release-bucket-list` | `MovieAggHTMLHandler.MovieAggAllReleaseBucketList` | `page.movie_release_bucket_list` | 待拆分 | 待拆分 | todo |

## 验收记录

- 当前只完成 v3 初始机械迁移骨架。
- `/cards` 已开始拆分：
  - v3 API：`/api/cards`。
  - v3 前端：`/cards`。
  - 后端继续调用 legacy 的 `parseMovieCardRequest`、`loadMovieCardPageData`、`movieSvc.ListMovieFull`。
  - API handler 内使用 legacy 路径 `/cards` 构造分页、排序、筛选链接，避免生成 `/api/cards` 链接。
  - 浏览器复查截图：`/Users/gaojinwei/src/my/rudy_gc/v3/tmp/cards/v3-cards-ps2.png`。
  - 本轮只确认 v3 `/cards?ps=2` 可渲染、API 200、图片请求 200、菜单存在、主要链接保持 `/cards` 语义。
- 其他页面 router 到 API、模板到独立前端的逐页拆分尚未执行。
- legacy/v3 截图逐项对比、后端输出逐字段对比、SQL/性能对比尚未执行。
