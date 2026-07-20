# 本地项目

本地项目承载现有业务 API 和前端。DailyBest 同步通过远端 HTTP API 获取快照，本地服务不直接访问远端数据库。

## 目录

- `api/`：本地 Go API、数据库迁移和 DailyBest 同步模块。
- `web/`：本地 Vue/Vite 前端。
- `make/`：本地项目的运行、验证和备份目标。

## 配置

本地 API 配置位于 `api/config/config.yaml`。`DailyBestRemote` 保存远端 API 地址和只读 Token；远端数据库连接只存在于远端项目配置中。

## 常用命令

```bash
make run-api
make run-web
make run-dev
make migrate-api
make verify
make stop-all
```

本地 API 与前端分别监听 `0.0.0.0:5013` 和 `0.0.0.0:5022`。
