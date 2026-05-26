# rudy_gc v2 部署说明

## 目录说明

- `docker-compose.yml`：本地容器编排入口，包含 `gc-api`、`gc-web`，并通过配置连接既有 MySQL/Redis。
- `gc-api.Dockerfile`：后端镜像构建文件。
- `gc-web.Dockerfile`：前端镜像构建文件。
- `gc-api.config.example.yaml`：后端容器配置模板。
- `README.md`：部署入口、配置和验证记录。

## 本地配置

首次部署前，将配置模板复制为真实运行配置：

```bash
cp deploy/gc-api.config.example.yaml deploy/gc-api.config.yaml
```

需要核对并写入真实值：

- `DataSource`：MySQL 连接信息，容器内默认主机为 `mysql:3306`。
- `Cache`：go-zero 缓存 Redis 配置，容器内默认主机为 `redis:6379`。
- `BizRedis`：业务 Redis 配置，容器内默认主机为 `redis:6379`。
- `Fetcher.Cookie`、`Fetcher.Proxy`：按实际抓取环境配置。
- `Film`、`Media`：按实际本机挂载路径配置。
- `/Volumes/Expansion`、`/Volumes/Getea`、`/Volumes/movie-un`、`/Volumes/T7/data`：Compose 已按 legacy 图片和媒体路径只读挂载到 `gc-api` 容器；这些路径缺失时，容器内图片与媒体静态访问会失败。

如需覆盖镜像名或标签，可复制 secrets 示例：

```bash
cp make/local.secrets.example.mk make/local.secrets.mk
```

然后修改：

- `GC_API_IMAGE`
- `GC_API_TAG`
- `GC_WEB_IMAGE`
- `GC_WEB_TAG`
- `DOCKER_USERNAME`
- `DOCKER_PASSWORD`

## Make 入口

所有命令都可以从 `v2` 根目录执行：

```bash
make run-api
make run-web
make run-all
make test-api
make test-web
make verify
make docker-config
make docker-build
make docker-up
make docker-down
```

也支持从 `v2/make` 目录直接按叶子 makefile 调用：

```bash
make -f gc-api.mk run-api
make -f gc-web.mk run-web
make -f run.mk verify
make -f deploy.mk docker-config
```

## 端口

- `gc-web`：默认 `2040`
- `gc-api`：默认 `2041`
- `mysql`：默认映射 `3306`
- `redis`：默认映射 `6379`

## 验证清单

- `make test-api`：运行后端单元测试。
- `make test-web`：运行前端单元测试。
- `make build-api`：编译后端包。
- `make build-web`：构建前端产物。
- `make verify`：串行执行后端测试、前端测试、后端构建、前端构建。
- `make docker-config`：检查 Compose 配置能否正常解析。
- 浏览器验证：启动 `make run-all` 后打开 `http://127.0.0.1:2040`，逐页截图对比 legacy 页面，检查导航、筛选、排序、分页、图片、按钮、详情页和 API 性能。

## 完成度记录

- API 拆分：`in_progress`，已具备 `gc-api` 独立 Go 模块、配置、基础路由和测试入口；legacy 真实动作 API 仍需逐项迁移。
- Web 拆分：`in_progress`，已具备 `gc-web` 独立 Vue/Vite 模块、路由覆盖和测试入口；通用页不能作为 legacy 页面完成态。
- Make 拆分：已完成根 Makefile 与 `make/*.mk` 叶子入口，支持从 `v2` 和 `v2/make` 两种目录执行。
- Deploy 拆分：`in_progress`，已完成 Dockerfile、Compose、配置模板、部署 Make 入口和 legacy `/Volumes/...` 只读挂载；仍需随功能迁移补齐实际验收记录。
- 验证入口：已完成 API/Web 测试、构建、统一 verify、Compose 配置检查入口。
