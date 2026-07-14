# rudy_gc v3

v3 是当前维护版本，使用 Go API 与 Vue 前端。

## 目录

- `gc-api/`：Go API 后端。
- `gc-web-next02/`：当前 Vue 前端。
- `oldmodel/`：来自 `legacy/oldmodel/` 的旧模型代码副本，保持 legacy 后端软链依赖。
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
- 前端：`make run-web-next02`（`make run-web` 为同一入口）
- API 与前端一起：`make run-all`
- 验证：`make verify`
- 停止端口：`make stop-all`
