# DailyBest 远端 API

远端服务长期抓取当前月榜，将影片名称、番号、标题和封面写入 `d_dailybest_movie`，将结构化排名写入 `c_rank`，同步重建每日统计与周期排行。原始页面只在一次抓取执行期间保留在内存中。

健康检查使用 `GET /healthz`，数据库就绪检查使用 `GET /readyz`。业务接口统一位于 `/api/v1/dailybest`。

## 本地同步

```http
GET /api/v1/dailybest/rank-sync-manifests
GET /api/v1/dailybest/ranks?day_number=1297&category=1
```

清单按业务日期和分类返回当前有效批次，字段包含 `fetch_run_id`、`business_date`、`day_number`、页面数、排名数、checksum 和完成时间。本地按清单获取排名，并校验日期序号、分类、数量与 checksum。

## 数据查询

```http
GET /api/v1/dailybest/ranks?business_date=2026-07-20&category=1&rank_pos=1&movie_jav_id=example&p=1&ps=24
GET /api/v1/dailybest/rank-day-stats?rank_date=2026-07-20&p=1&ps=24
GET /api/v1/dailybest/rank-periods?period_type=2&period_key=2026-07&category=1&p=1&ps=24
GET /api/v1/dailybest/rank-period-items?period_id=1&movie_jav_id=example&p=1&ps=24
GET /api/v1/dailybest/rank-period-stats?period_id=1&p=1&ps=24
```

筛选参数均可省略。设置 `p` 或 `ps` 时响应包含统一的 `page_info`；未设置时返回全部匹配结果。

排名明细和周期明细中的 `movie` 返回影片名称、番号、标题和封面；周期统计中的 `rank_one_movie` 返回第一名影片的相同信息。

## 抓取任务

```http
POST /api/v1/dailybest/fetch
Content-Type: application/json

{
  "business_date": "2026-07-20",
  "category": 1
}
```

源站只提供当前榜单，服务会校验业务日期。请求成功返回 `202` 和抓取运行 ID。

```http
GET /api/v1/dailybest/fetch-runs?p=1&ps=24
GET /api/v1/dailybest/fetch-runs/{run_id}
```

抓取记录保存页面数、排名数、checksum、状态和错误信息。状态包括 `queued`、`running`、`succeeded`、`no_change` 和 `failed`。

## 写入语义

- 完整抓取、解析和 checksum 计算成功后才进入发布。
- checksum 与当前日期、分类的排名一致时，更新影片基本信息，并将抓取记录标记为 `no_change`。
- 内容变化时，在同一事务中更新影片基本信息、替换对应 `c_rank`、重建每日统计与周期排行，并将抓取记录标记为 `succeeded`。
- 发布失败时事务回滚，随后独立记录抓取失败信息。
