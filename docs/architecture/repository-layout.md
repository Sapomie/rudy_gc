# 仓库目录结构

- `api/`：Go API 应用。
  - `cmd/`：API 与数据库迁移命令。
  - `internal/transport/http/`：HTTP 路由与处理器。
  - `internal/service/`：业务服务。
  - `internal/repository/`：数据库访问实现。
  - `internal/schema/`：数据库表结构和既有迁移实现。
  - `internal/migration/`：迁移命令统一入口。
  - `internal/platform/`：配置、数据库、日志和基础设施适配。
  - `internal/legacy/`：仍被当前业务调用的历史实现。
- `web/`：Vue/Vite 前端。
  - `src/app/`：应用入口、路由、布局和全局样式入口。
  - `src/features/`：按业务域组织的页面与组件。
  - `src/shared/`：跨业务复用的 API、组件、类型和工具。
- `configs/`：本地与 Docker 的完整运行配置。
- `deploy/`：部署编排与镜像构建。
- `docs/`：架构、规范、运维和迁移历史。
- `make/`：根 Makefile 引入的目标实现。
