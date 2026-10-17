# 本地项目

本地项目承载现有业务 API 和前端。DailyBest 日期页面通过远端 HTTP API 分页获取已有天数，用户选择缺失日期后再获取对应排名，本地服务不直接访问远端数据库。

## 目录

- `api/`：本地 Go API、数据库迁移和 DailyBest 同步模块。
- `web/`：本地 Vue/Vite 前端。
- `make/`：本地项目的运行、验证和备份目标。

## 配置

本地 API 配置位于 `api/config/config.yaml`。`DailyBestRemote` 保存远端 API 地址和可选的 HTTP 代理地址；`Proxy`为空时直连 Remote，配置完整 HTTP/HTTPS 地址时才经过代理。远端数据库连接只存在于远端项目配置中。

Local 直接抓取 Javlib 的域名和请求参数保存在本地 API 数据库的 `t_fetch_site` 中，`site_code` 为 `javlib`，通过前端 `/siteconfiglist` 与其他站点统一编辑。保存后当前 API 进程立即更新地址、UserAgent、Cookie、代理、超时和启用状态。Javlib 分页及重试流程继续使用现有策略。

`Fetcher.JavAddress`、`Fetcher.UserAgent`、`Fetcher.Cookie`、`Fetcher.Timeout` 仅作为 Javlib 首次入库的来源；已有数据库配置优先。API 启动和数据库迁移会补齐缺失的 Javlib 配置行。仅初始化这一行时，在 `local/api/` 执行 `go run ./cmd/migrate -f config/config.yaml -javlib-only`，该命令保留已有站点配置。

## 常用命令

以下命令在 `local/` 目录执行：

```bash
make run-api
make run-web
make run-dev
make migrate-api
make verify
make stop-all
```

从工作区根目录同时启动或停止本地 API 和前端时，使用：

```bash
make run-local
make stop-local
```

本地 API 与前端分别监听 `0.0.0.0:5021` 和 `0.0.0.0:5022`。
