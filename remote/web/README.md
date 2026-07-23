# Rudy-GC Remote Web

远端采集服务的独立 Vue 管理端。页面样式与 local 前端保持一致，包含运行概览、抓取任务、请求日志、排名明细、每日统计和手动抓取；排名明细同时展示影片名称、番号、标题和封面。请求日志页仅在用户主动查询、刷新或切换分页时加载数据。

开发服务器监听 `0.0.0.0:5032`，并将 `/healthz`、`/readyz` 和 `/api` 请求代理到远端 API。

```bash
npm install --prefix new
npm run dev
```

验证命令：

```bash
npm run check
npm test
npm run build
```
