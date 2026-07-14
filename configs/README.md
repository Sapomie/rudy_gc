# 运行配置

- `local/api.yaml`：本地 `make run-api` 与 `make migrate-api` 使用的完整配置。
- `docker/api.yaml`：Docker Compose 挂载到 API 容器的完整配置。

移动配置时必须同步检查数据库、Redis、远程数据源、媒体目录和抓取站点连接，
运行必需字段不得留空。
