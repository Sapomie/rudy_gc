# DailyBest 远端服务部署

## 配置

从 `remote/` 项目目录执行命令时，运行配置位于 `api/config/server.yaml`，包含：

- 远端数据库实际连接信息。
- `0.0.0.0:5031` 服务监听地址。
- `Asia/Shanghai` 时区、每日 `08:30` 抓取和 `1h`失败恢复。
- SMTP 地址、端口、发件邮箱、授权码、收件邮箱、发件人名称和远端详情页公网地址。
- API启动后立即执行一次目标站探测，完成后按`180m`周期继续探测。
- FlareSolverr内部地址、持久会话名和页面请求超时。

API通过独立FlareSolverr容器访问源站。Cookie和User-Agent由持久Solver会话管理，不写入YAML、数据库或请求日志。Solver使用服务器host网络执行目标请求，只绑定Docker内部bridge网关`172.30.0.1:8191`。

本地配置位于 `local/api/config/config.yaml` 的 `DailyBestRemote`，只保存远端 API 地址。本地配置不包含远端数据库连接。

## 数据库迁移

```bash
cd remote
make migrate-api
```

远端迁移创建或升级：

- `d_dailybest_fetch_run`
- `d_dailybest_fetch_log`
- `d_dailybest_runtime_event`
- `d_dailybest_target_probe`
- `d_dailybest_movie`
- `d_dailybest_source_config`
- `c_rank`
- `c_rank_day_stat`

`c_rank_day_stat`直接保存页面使用的统计值、源排行抓取完成时间`fetched_on`、排行入库时间`synced_on`和计算时间；每日统计查询不再连接抓取任务表或排名明细表。旧数据迁移只回填新增时间字段，不覆盖已有统计值和原有时间字段。

旧快照表存在时，迁移会先把最新影片基本信息转入 `d_dailybest_movie`，再把每个日期、分类的最新快照转入 `c_rank`；校验排名数量后删除三张旧快照表。服务启动后只为缺少统计行的排行日期生成一次`c_rank_day_stat`，已有日期保持不变。迁移同时删除已经停用的 `c_rank_period_stat`、`c_rank_period_item` 和 `c_rank_period`。

本地进入 `local/` 后执行 `make migrate-api` 创建或升级 `d_dailybest_sync_rank`。

## 生产目录

Docker版远端服务与共享MySQL使用以下目录：

```text
/var/snap/docker/common/rudy_gc/remote/
├── releases/<Docker版本>/
├── config/server.yaml
└── backups/

/var/snap/docker/common/rudy_infra/mysql/
├── data/
├── config/mysql.env
└── backups/
```

- `remote-solver`使用固定摘要的FlareSolverr镜像，容器使用host网络且只绑定`172.30.0.1:8191`。
- `rudy_mysql_shared`作为服务器共享基础设施独立运行，通过外部网络`rudy_shared_infra`向各Docker服务提供`mysql`网络别名。
- `remote-api`监听`0.0.0.0:5031`，通过外部网络`rudy_shared_infra`连接共享MySQL。
- `remote-web`监听`0.0.0.0:5032`，容器内Nginx代理`/api/`和`/readyz`到`remote-api:5031`；`/healthz`只检查Web进程。
- 三个容器使用`restart: unless-stopped`，进程异常后由Docker重新拉起。
- Docker不配置健康检查；`/healthz`和`/readyz`保留给人工检查和运行状态页面。
- 容器日志单文件`20m`，保留`5`个。

数据库迁移使用API镜像内的独立迁移程序：

```bash
make docker-migrate
```

## 离线镜像交付

共享MySQL和Remote应用分别生成离线包。构建机可以访问镜像仓库，生产服务器只接收归档并执行`docker load`。

本机构建共享MySQL包：

```bash
make shared-mysql-bundle RELEASE_TAG=<发布版本>
```

本机构建Remote包：

```bash
cd remote
make docker-build RELEASE_TAG=<发布版本>
```

服务器先安装共享MySQL并恢复生产数据库，再安装Remote：

```bash
scripts/install-shared-mysql.sh <MySQL离线包目录> \
  /var/snap/docker/common/rudy_infra/mysql/config/mysql.env \
  /var/snap/docker/common/rudy_infra/mysql/data

scripts/install-remote-offline.sh \
  <Remote离线包目录> \
  /var/snap/docker/common/rudy_gc/remote/config/server.yaml
```

生产`server.yaml`设置为`root:1001`、权限`640`，供 API 镜像内 UID/GID `1001` 的`lighthouse`用户只读访问。安装脚本依次等待 Solver 和 API 就绪后再启动下游容器。

安装脚本逐项校验SHA-256、镜像存在性和AMD64架构。Compose镜像使用`localhost/`名称和`pull_policy: never`；所有启动命令附带`--pull never --no-build`，生产服务器不访问Docker Hub或GHCR。

## 运行和检查

```bash
cd remote
make docker-config REMOTE_ENV_FILE=<发布目录>/remote-release.env
make docker-up REMOTE_ENV_FILE=<发布目录>/remote-release.env
```

服务启动后检查：

```text
GET /healthz
GET /readyz
```

生产服务状态检查：

```bash
make docker-status
make docker-logs
```

生产配置的核心时间参数为：

```yaml
DailyFetchTime: "08:30"
FailureRetryDelay: "1h"
TargetProbe:
  Interval: "180m"
BrowserMonitorInterval: "10s"
Source:
  Browser:
    BaseURL: "http://172.30.0.1:8191"
    Session: "rudy_gc_dailybest"
```

每日`08:30`自动抓取首次失败时发送邮件，后续同一业务日期和榜单分类的恢复任务连续失败不重复发送；恢复任务成功后发送一次恢复邮件。手动抓取不触发邮件通知。邮件发送异常只写入服务日志，不改变抓取任务状态和失败恢复流程。

## 构建和测试

```bash
cd remote
make build-api
make test-api
make docker-build RELEASE_TAG=<发布版本>
```

## 独立运行验收

部署完成后，API、Web、Solver、共享MySQL、生产配置和持久数据全部位于新服务器。生效配置和发布文件不得包含旧服务器地址；新服务器不使用旧服务器的数据库、代理、Cookie、文件、端口、SSH隧道或常驻进程。本地项目只访问`8.134.149.91`。
