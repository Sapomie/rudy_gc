# DailyBest 远端服务部署

## 配置

从 `remote/` 项目目录执行命令时，运行配置位于 `api/config/server.yaml`，包含：

- 远端数据库实际连接信息。
- `0.0.0.0:8891` 服务监听地址。
- 只读 Token 和管理 Token。
- `Asia/Shanghai` 时区及每日抓取时间。
- 源站地址、Cookie、User-Agent 和请求超时。

本地配置位于 `local/api/config/api.yaml` 的 `DailyBestRemote`，只保存远端 API 地址和只读 Token。本地配置不包含远端数据库连接。

## 数据库迁移

```bash
cd remote
make migrate-api
```

远端迁移只创建：

- `dailybest_fetch_run`
- `dailybest_snapshot`
- `dailybest_snapshot_item`
- `dailybest_snapshot_page`

本地进入 `local/` 后执行 `make migrate-api` 创建 `d_dailybest_sync_snapshot`。

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
