# rudy_gc

rudy_gc 使用 Go API 与 Vue 前端，当前代码直接维护在仓库根目录。

## 目录

- `api/`：Go API、数据库模型、业务服务和迁移入口。
- `web/`：Vue/Vite 前端。
- `docs/`：架构、规范和历史迁移文档。
- `make/`：Make 目标实现。

## 本地运行

- 后端：`make run-api`
- 前端：`make run-web`
- API 与前端一起：`make run-dev`
- 验证：`make verify`
- 停止端口：`make stop-all`

默认 API 配置为 `api/cmd/api/config.yaml`，该文件包含实际运行所需的完整连接信息。
