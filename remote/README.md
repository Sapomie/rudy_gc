# 远端项目

远端项目负责按计划抓取当日排行、保存影片基本信息、结构化排名和派生统计，并通过 HTTP API 向本地项目提供同步数据。远端项目拥有独立的 API、前端、配置、迁移和 Make 入口。

## 目录

- `api/`：远端 Go API、抓取任务、排名存储、同步接口和数据库迁移。
- `web/`：远端 Vue/Vite 管理前端。
- `deploy/docker/`：API、Browser Solver和Web的Docker部署配置。
- `make/`：远端项目的运行与验证目标。

## 配置

远端API配置位于`api/config/server.yaml`，包含数据库、抓取和恢复时间、`180m`目标站探测以及FlareSolverr内部地址和会话配置。生产配置由服务器`/home/lighthouse/data/rudy_gc/remote/config/server.yaml`只读挂载。源站Cookie由服务器内Solver会话自动管理。

## 常用命令

```bash
make run-api
make run-web
make run-dev
make migrate-api
make verify
make stop-all
make docker-config
make docker-build
make docker-migrate
make docker-up
make docker-status
```

远端API与前端分别监听`0.0.0.0:5031`和`0.0.0.0:5032`。FlareSolverr只监听Docker内部网关`172.30.0.1:8191`。
