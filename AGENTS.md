没有对比，没到90%不准验收过关！！！
没有对比，没到90%不准验收过关！！！
# rudy_gc v2 仓库指南

## 当前维护范围
- 后续功能开发、修复、验证和文档更新只面向 `v2/`。
- `legacy/`、根目录旧 `cmd/`、旧 `internal/`、旧 `ui/`、旧 `make/`、`oldmodel/`、`z_text/` 作为历史资料和 legacy 行为基准保留。
- 历史系统在 v2 页面重建中是功能语义、后端逻辑、参数语义、跳转逻辑、字段完整性、API 输出、截图接近度的强制基准。
- 无法保证优化正确时，必须按 legacy 行为、legacy 后端语义、legacy 跳转参数和 legacy SQL 口径原样迁移；宁可保持 legacy 逻辑，也不得凭猜测改写。
- 历史系统的模板、CSS、JS、class、静态资源、布局结构、页面壳、交互代码、Make 目标和实现风格都不得约束 v2。
- 禁止读取或使用 `z_reference` 作为 v2 前端参考源。

## v2 项目结构
- `v2/gc-api/`：v2 后端 API。
- `v2/gc-web/`：v2 Vue/Vite/TypeScript 前端。
- `v2/docs/`：v2 重建计划、进度和验收记录。
- `v2/deploy/`：v2 部署、容器、代理和运行配置。
- `v2/tmp/`：v2 验收截图和临时产物。

## v2 常用命令
- 后端测试：`cd v2/gc-api && go test ./...`
- 前端测试：`cd v2/gc-web && npm run test`
- 前端构建：`cd v2/gc-web && npm run build`
- 全量验证：`make -C /Users/gaojinwei/src/my/rudy_gc/v2 verify`
- 停止 v2 临时服务：`make -C /Users/gaojinwei/src/my/rudy_gc/v2 stop-all`

## 协作与执行
- 使用中文回复。
- 用户提问时先回答问题，再执行其他动作。
- 用户问“为什么 / 原因 / why”时，先说明当时判断依据、前提、取舍、风险和具体错点。
- 任何写操作前必须先给出明确方案：修改文件 + 主要改动点。
- 只有用户最新一条消息完整等于 `确认` 时，才允许执行写操作。
- 用户给出新的明确指令时，必须立即切换到最新指令；被用户打断、停止、否定或中止的任务不得继续执行或自行恢复。
- 禁止主动要求用户发送固定字样作为执行前提。
- 自身写错、判断错、实现偏差、漏查或误改时，必须先说“对不起”，再说明自身具体错点和修复动作。

## Git 与变更说明
- 不得回滚用户现有改动，除非用户明确要求。
- 用户要求 `git message`、`commit message`、提交说明时，输出必须全部使用中文。
- 当用户要求 `git message` 时，固定检查顺序：
  1. `git diff --stat --cached`
  2. `git diff --stat`
  3. 若都为空，再 `git status -sb`
- 回答 `git message` 相关问题时，必须明确说明暂存区和未暂存区检查结果。

## 安全默认策略
- 默认禁止破坏性命令：
  - `rm -rf`
  - `git reset --hard`
  - `git checkout -- <file>`
  - DROP/TRUNCATE 类 SQL
- 如确需执行，先说明风险，再获取用户明确授权。
