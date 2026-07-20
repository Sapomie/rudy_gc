# DailyBest 远端服务部署

## 配置

从 `remote/` 项目目录执行命令时，运行配置位于 `api/config/server.yaml`，包含：

- 远端数据库实际连接信息。
- `0.0.0.0:5031` 服务监听地址。
- `Asia/Shanghai` 时区及每日抓取时间。
- 源站地址、Cookie、User-Agent 和请求超时。

本地配置位于 `local/api/config/config.yaml` 的 `DailyBestRemote`，只保存远端 API 地址。本地配置不包含远端数据库连接。

## 数据库迁移

```bash
cd remote
make migrate-api
```

远端迁移创建或升级：

- `d_dailybest_fetch_run`
- `d_dailybest_movie`
- `c_rank`
- `c_rank_day_stat`
- `c_rank_period`
- `c_rank_period_item`
- `c_rank_period_stat`

旧快照表存在时，迁移会先把最新影片基本信息转入 `d_dailybest_movie`，再把每个日期、分类的最新快照转入 `c_rank`；校验排名数量后删除三张旧快照表，并重建每日统计与周期排行。

本地进入 `local/` 后执行 `make migrate-api` 创建或升级 `d_dailybest_sync_rank`。

## 运行和检查

```bash
cd remote
make run-api
```

服务启动后检查：

```text
GET /healthz
GET /readyz
```

部署防火墙需要允许本地服务器访问远端 API 端口。远端数据库端口只供远端服务所在主机访问。

## 构建和测试

```bash
cd remote
make build-api
make test-api
```
