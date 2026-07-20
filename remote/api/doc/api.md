# DailyBest 远端 API

## 服务职责

`dailybest-remote` 长期运行并抓取当前 DailyBest 月榜。服务完整抓取所有页面、解析结构化排名、校验内容并发布不可变快照。本地 `rudy_gc` 只通过本 API 同步快照。

## 认证

- `/api/v1/dailybest/*` 使用只读 Bearer Token。
- `/api/v1/admin/dailybest/*` 使用管理 Bearer Token。
- `/healthz` 和 `/readyz` 用于服务检查。

## 同步接口

### 获取增量快照

```http
GET /api/v1/dailybest/snapshots?after_revision=0
Authorization: Bearer <read-token>
```

响应按 `revision` 升序返回所有已发布快照。首次同步使用 `after_revision=0`。

### 获取完整快照

```http
GET /api/v1/dailybest/snapshots/{revision}
Authorization: Bearer <read-token>
```

响应包含快照元数据、原始页面和结构化排名条目。本地会校验页面 checksum、页面数量、条目数量和快照 checksum。

### 获取指定日期最新快照

```http
GET /api/v1/dailybest/dates/{business_date}?category=1
Authorization: Bearer <read-token>
```

### 获取原始页面

```http
GET /api/v1/dailybest/snapshots/{revision}/pages
Authorization: Bearer <read-token>
```

## 运维接口

### 触发当前榜单抓取

```http
POST /api/v1/admin/dailybest/fetch
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "business_date": "2026-07-19",
  "category": 1
}
```

源站当前接口只提供当前榜单，服务会校验 `business_date`。请求成功返回 `202` 和抓取运行 ID。

### 查询运行记录

```http
GET /api/v1/admin/dailybest/fetch-runs
GET /api/v1/admin/dailybest/fetch-runs/{run_id}
Authorization: Bearer <admin-token>
```

## 快照语义

- 快照在完整抓取、解析和校验成功后通过事务发布。
- `dailybest_snapshot.id` 是对外 revision。
- 同一日期远端内容发生变化时会发布新 revision。
- checksum 相同的重复抓取记录为 `no_change`。
- 本地处理完成后才推进同步游标；中断后会重新拉取并幂等处理未完成 revision。
