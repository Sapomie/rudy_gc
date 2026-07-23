# DailyBest 远端 API

远端服务长期抓取当前月榜，将影片名称、番号、标题和封面写入 `d_dailybest_movie`，将结构化排名写入 `c_rank`，同步重建每日统计。原始页面只在一次抓取执行期间保留在内存中。

进程检查使用`GET /healthz`，数据库和Browser Solver就绪检查使用`GET /readyz`。这些接口供人工排查和运行状态页面使用。业务接口统一位于`/api/v1/dailybest`。

## 本地同步

```http
GET /api/v1/dailybest/rank-days?business_date_start=2026-07-01&business_date_end=2026-07-20&p=1&ps=24
GET /api/v1/dailybest/ranks?day_number=1297&category=1
```

`rank-days` 以天为行返回远端实际存在的排名日期，日期范围为闭区间，结束日期可省略。列表分页按天数计算；当天各分类的最新有效抓取批次放在 `categories`。日期与排名数来自 `c_rank`，页面数、抓取条目数、checksum 和完成时间来自 `d_dailybest_fetch_run`。

本地只为用户选中的日期调用 `ranks`，并校验日期序号、分类、数量与 checksum。查询单个日期元数据时使用 `day_number`，无需分页参数：

```http
GET /api/v1/dailybest/rank-days?day_number=1297
```

## 数据查询

```http
GET /api/v1/dailybest/ranks?business_date=2026-07-20&category=1&rank_pos=1&movie_jav_id=example&p=1&ps=24
GET /api/v1/dailybest/rank-day-stats?rank_date=2026-07-20&p=1&ps=24
```

筛选参数均可省略。设置 `p` 或 `ps` 时响应包含统一的 `page_info`；未设置时返回全部匹配结果。

排名明细中的 `movie` 返回影片名称、番号、标题和封面。

## 抓取任务

```http
POST /api/v1/dailybest/fetch
```

该接口不接收请求数据。服务自动计算当前业务日期并使用固定的月榜分类。请求成功返回 `202` 和抓取运行 ID。

```http
GET /api/v1/dailybest/fetch-runs?p=1&ps=24
GET /api/v1/dailybest/fetch-runs/{run_id}
```

抓取记录保存页面数、排名数、checksum、状态和错误信息。状态包括 `queued`、`running`、`succeeded`、`no_change` 和 `failed`。

## 请求日志

```http
GET /api/v1/dailybest/fetch-logs?fetch_run_id=12&page_number=3&status=failed&p=1&ps=24
```

请求日志按 ID 倒序返回。`fetch_run_id`、`page_number` 和 `status` 均可省略，状态包括 `requesting`、`succeeded` 和 `failed`。每条日志包含请求地址、HTTP 状态码、耗时、错误信息以及创建和更新时间。

## 运行监控

```http
GET /api/v1/dailybest/runtime-status
GET /api/v1/dailybest/runtime-events?p=1&ps=24
GET /api/v1/dailybest/target-probes?p=1&ps=24
```

API启动后立即通过FlareSolverr真实浏览器探测目标站一次，完成后每`180m`继续探测。每次探测写入`d_dailybest_target_probe`；API、Browser Solver和目标站状态变化写入`d_dailybest_runtime_event`。页面查询只读取运行状态和数据库记录，不会触发目标网站请求。

## 写入语义

- 完整抓取、解析和 checksum 计算成功后才进入发布。
- checksum 与当前日期、分类的排名一致时，更新影片基本信息，并将抓取记录标记为 `no_change`。
- 内容变化时，在同一事务中更新影片基本信息、替换对应 `c_rank`、重建每日统计，并将抓取记录标记为 `succeeded`。
- 发布失败时事务回滚，随后独立记录抓取失败信息。
