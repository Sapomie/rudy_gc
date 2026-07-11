# rudy_gc v3

v3 是 legacy 项目的前后端分离拆分迁移。

## 目录

- `gc-api/`：来自 `legacy/backend/` 的后端代码副本。页面迁移时只允许把 legacy router 页面入口改成 API 输出。
- `gc-web/legacy-ui/`：来自 `legacy/backend/ui/` 的模板、静态资源和页面脚本副本。前端拆分必须从这里继续迁移。
- `oldmodel/`：来自 `legacy/oldmodel/` 的旧模型代码副本，保持 legacy 后端软链依赖。
- `z_text/`：来自 `legacy/z_text/` 的文本资源副本，供 legacy 静态文本入口继续解析。
- `docs/`：v3 拆分迁移记录。
- `deploy/`：v3 部署记录；`deploy/legacy-deploy/` 是 legacy deploy 的机械副本。

## 硬规则

- 所有页面的底层后端逻辑必须保持 legacy 原逻辑。
- 参数解析、默认值、筛选、排序、分页、SQL、数据组装、跳转语义必须保持 legacy 原样。
- 页面形态、字段、按钮、tab、链接、下钻参数和可见行为必须保持 legacy 原样。
- 迁移只能做 router 到 API、模板到独立前端的拆分。
- 禁止重写业务逻辑，禁止凭理解优化。

## 本地运行

- 后端：`make run-api`
- 前端：`make run-web`
- API 与两个前端一起：`make run-all`（同时启动 legacy `gc-web` 和 `gc-web-next02`）
- 验证：`make verify`
- 停止端口：`make stop-all`（同时停止 API、legacy `gc-web` 和 `gc-web-next02`）
