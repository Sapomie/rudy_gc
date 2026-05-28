# v3 deploy

`legacy-deploy/` 是从 `legacy/deploy/` 机械复制来的部署来源。

v3 部署只允许围绕前后端分离做必要拆分：

- 后端从 legacy router 页面入口拆成 API 服务。
- 前端从 legacy 模板和静态资源拆成独立前端服务。
- `gc-api` 的 router 只提供 `/api/...`，不提供 HTML 页面、不加载模板、不提供 UI 静态资源。
- `gc-web` 提供页面、legacy 静态资源、本地媒体文件读取和 API 代理。
- 运行配置中的数据库、Redis、外部路径和业务凭证必须保持真实有效，不得用空值或占位值交付。

## 本地端口

- `gc-api` 默认监听 `:5013`，避免占用 legacy 的 `:5010`。
- `gc-web` 默认监听 `:5012`，通过 `V3_API_ORIGIN` 代理到 v3 API。

## Make 命令

- `make -C /Users/gaojinwei/src/my/rudy_gc/v3 paths`：查看路径和端口。
- `make -C /Users/gaojinwei/src/my/rudy_gc/v3 run-api`：启动 v3 API。
- `make -C /Users/gaojinwei/src/my/rudy_gc/v3 run-web`：启动 v3 web。
- `make -C /Users/gaojinwei/src/my/rudy_gc/v3 run-all`：同时启动 v3 API 和 v3 web。
- `make -C /Users/gaojinwei/src/my/rudy_gc/v3 stop-all`：停止 v3 API 和 v3 web 端口。
- `make -C /Users/gaojinwei/src/my/rudy_gc/v3 verify`：执行后端测试和前端检查。

子 make 文件也必须支持单独执行：

- `/usr/bin/make -f /Users/gaojinwei/src/my/rudy_gc/v3/make/run.mk -C /Users/gaojinwei/src/my/rudy_gc/v3/make run-all`
- `/usr/bin/make -f /Users/gaojinwei/src/my/rudy_gc/v3/make/run.mk -C /Users/gaojinwei/src/my/rudy_gc/v3/make stop-all`
- `/usr/bin/make -f /Users/gaojinwei/src/my/rudy_gc/v3/make/gc-api.mk -C /Users/gaojinwei/src/my/rudy_gc/v3/make test-api`
- `/usr/bin/make -f /Users/gaojinwei/src/my/rudy_gc/v3/make/gc-web.mk -C /Users/gaojinwei/src/my/rudy_gc/v3/make check-web`
- `/usr/bin/make -f /Users/gaojinwei/src/my/rudy_gc/v3/make/verify.mk -C /Users/gaojinwei/src/my/rudy_gc/v3/make verify`
