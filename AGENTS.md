# 仓库指南

## 项目结构与模块组织
- `cmd/` 存放入口程序，例如 `cmd/api`（主服务）和 `cmd/db_migrate`。
- `internal/` 存放核心业务代码（config、domain、repo、transport、observability）。
- `pkg/` 放置可在 `internal/` 外复用的共享包。
- `data/` 包含生成的模型和本地数据资源（见 `data/modelx`）。
- `ui/` 包含服务端渲染模板与静态资源（`ui/templates`、`ui/static`）。
- `deploy/` 存放 Docker 构建资源；`make/` 存放拆分的 Make 目标。
- `oldmodel/` 与 `z_text/` 为遗留/笔记目录，除非必须否则不要新增工作。

## 构建、测试与开发命令
- `go run ./cmd/api` 本地启动 API 服务。
- `make -f make/dev.mk run` 通过 Make 启动同一入口。
- `make -f make/model.mk gen-model` 使用 `goctl` 从 MySQL 生成模型。
- `make -f make/migrate.mk auto-migrate` 执行数据库迁移。
- `make -f make/docker.mk docker-build` 使用 `deploy/docker/api.Dockerfile` 构建镜像。
- `make -f make/cache.mk clear-cache` 清理匹配 `cache:rudyGc:*` 的 Redis 键。

## 编码风格与命名规范
- Go 代码遵循 `gofmt`，修改 `.go` 文件后需执行 `gofmt`。
- 包名简短小写，与目录名一致。
- 导出标识符使用 `PascalCase`，非导出使用 `camelCase`。
- 模板位于 `ui/templates/pages`，共享片段在 `ui/templates/partials`。
- 静态资源按类型放在 `ui/static/` 下（如 `css`、`js`、`image`）。

## 数据分层与缓存规则
- 仅 `data/modelx` 允许写 SQL；所有 SQL 写在 `*_model.go` 的自定义方法中。
- `internal/infra` 的 repo 实现只调用 modelx 方法，不写 SQL。
- `internal/repo` 仅定义接口，不写 SQL，也不直接使用 modelx。
- `internal/domain` 业务层只调用 repo，禁止直接使用 `SqlConn` 或写 SQL。
- 读取：`FindOne/FindOneByX` 使用 modelx 的缓存查询。
- 更新：必须调用 `*_model_gen.go` 的 `Update`/`Delete` 以确保清缓存。
- 自定义查询/聚合：在 `*_model.go` 中新增方法，并使用 `QueryRowNoCacheCtx/QueryRowsNoCacheCtx`。
- 避免无谓清缓存：repo 层应先对比旧值，无变化则跳过 `Update`。

## 测试指南
- 当前无 `_test.go` 文件；新增测试应与被测包同目录。
- 新增测试时使用 `go test ./...` 进行包级测试。
- 测试命名用 `TestXxx`，多场景使用表驱动。

## 提交与 PR 规范
- 近期提交使用 `feat`、`refactor` 等前缀，也有 `feat(ui)` 这种带 scope 的形式。
- 优先使用 `type(scope): 简短描述` 的祈使语气（如 `feat(api): add movie cache warmup`）。
- PR 需包含清晰描述、关联 issue 链接，UI 变更需附截图。
- 说明任何特殊环境（DB/Redis）以及验证用的命令。

## 配置说明
- 默认 DB/Redis 配置在 `make/common.mk`（如 `DB_URL`、`REDIS_HOST`）。
- 可用环境变量或 `make` 参数覆盖：`make -f make/model.mk gen-model DB_URL=...`。
