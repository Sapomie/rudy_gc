# 远端项目

远端项目负责按计划抓取当日排行、保存影片基本信息、结构化排名和派生统计，并通过 HTTP API 向本地项目提供同步数据。远端项目拥有独立的 API、前端、配置、迁移和 Make 入口。

## 目录

- `api/`：远端 Go API、抓取任务、排名存储、同步接口和数据库迁移。
- `web/`：远端 Vue/Vite 管理前端。
- `deploy/docker/`：API、Browser Solver和Web的离线Docker部署配置。
- `make/`：远端项目的运行与验证目标。

## 配置

JAV 网站域名保存在远端数据库 `d_dailybest_source_config`，通过 Remote 的 `/dailybestsourceconfig` 或 Local 的 `/remotedailybestsourceconfig` 页面修改。两端页面读写同一远端配置，保存后下一轮抓取和探测生效，同一轮分页使用开始时的域名。YAML 的 `Source.JavAddress` 仅用于数据库配置缺失时首次导入；已有数据库配置会在重启时保留。探测地址由数据库域名生成，YAML 的 `TargetProbe` 只保留探测间隔。

远端API配置位于`api/config/server.yaml`，包含数据库、抓取和恢复时间、`180m`目标站探测、FlareSolverr内部地址和会话配置，以及每日抓取失败与恢复通知使用的SMTP配置和详情页公网地址。生产配置由服务器`/var/snap/docker/common/rudy_gc/remote/config/server.yaml`只读挂载。源站Cookie由服务器内Solver会话自动管理。

生产镜像统一在本机按`linux/amd64`构建并通过`docker save`生成离线包。服务器只执行镜像校验、`docker load`、数据库迁移和`docker compose up --pull never --no-build`，不访问Docker Hub或GHCR。MySQL作为服务器共享基础设施独立运行，Remote只通过外部网络`rudy_shared_infra`访问网络别名`mysql`。

## 常用命令

```bash
make run-api
make run-web
make run-dev
make migrate-api
make verify
make stop-all
make docker-config
make docker-build RELEASE_TAG=<发布版本>
make docker-migrate
make docker-up
make docker-status
```

远端API与前端分别监听`0.0.0.0:5031`和`0.0.0.0:5032`。FlareSolverr只监听Docker内部网关`172.30.0.1:8191`。
