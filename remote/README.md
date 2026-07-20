# 远端项目

远端项目负责按计划抓取当日排行、保存不可变快照，并通过带 Token 的 HTTP API 向本地项目提供同步数据。远端项目拥有独立的 API、前端、配置、迁移和 Make 入口。

## 目录

- `api/`：远端 Go API、抓取任务、快照存储、同步接口和数据库迁移。
- `web/`：远端 Vue/Vite 管理前端。
- `make/`：远端项目的运行与验证目标。

## 配置

远端 API 配置位于 `api/config/server.yaml`，包含远端数据库、API Token、抓取时间和源站访问配置。部署前必须核对该文件中的实际凭证和连接信息。

## 常用命令

```bash
make run-api
make run-web
make run-dev
make migrate-api
make verify
make stop-all
```

远端 API 与前端分别监听 `0.0.0.0:8891` 和 `0.0.0.0:8892`。
