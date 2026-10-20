# Web 前端

当前前端代码位于 `new/`，默认使用 5022 端口，并监听 `0.0.0.0`。

开发服务器的 `/api` 默认代理到本地 API `http://127.0.0.1:5021`，`/remote-api`
默认直连 Remote API `http://175.178.245.71:5031`。只有显式设置`REMOTE_API_PROXY`
时才通过对应的 HTTP/HTTPS 代理转发。

“远端 DailyBest → 站点配置”（`/remotedailybestsourceconfig`）通过现有 `/remote-api` 通道读取和修改远端数据库中的 JAV 域名，与 Remote 的站点配置页面共用同一配置。保存后下一轮远端抓取和探测使用新域名。

## 安装

```bash
npm --prefix new ci
```

## 启动

```bash
npm run dev
```

需要切换 Remote API 地址时：

```bash
REMOTE_API_ORIGIN=http://remote-api-host:5031 npm run dev
```

需要显式使用代理时，再设置`REMOTE_API_PROXY=http://proxy-host:port`。

也可以在项目根目录执行：

```bash
make run-web
```

## 检查和构建

```bash
npm run check
npm test
npm run build
```

对应的 `local/` 目录命令为：

```bash
make check-web
make test-web
make build-web
```

在 `local/` 目录执行 `make run-all`，或在工作区根目录执行 `make run-local`，可同时启动本地 API 和当前前端。
