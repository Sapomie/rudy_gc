# rudy_gc

rudy_gc 工作区包含相互隔离的本地项目和远端项目。两个项目各自拥有 Go API、Vue 前端、运行配置和 Make 入口，只共享稳定的 DailyBest API 协议与校验算法。

## 目录

- `local/`：本地项目，通过远端 API 同步 DailyBest 数据，并承载现有业务功能。
- `remote/`：远端项目，定时抓取当日排行、保存快照并提供同步 API。
- `shared/go/`：两个 Go API 共用的稳定协议模块，不包含数据库、配置或业务实现。
- `make/`：工作区级 Make 目标。
- `go.work`：组合本地 API、远端 API 和共享协议模块。

每个项目的完整命令和配置位置分别见 `local/README.md` 与 `remote/README.md`。

## 工作区命令

```bash
make local-help
make remote-help
make run-local
make run-remote
make stop-local
make stop-remote
make verify
make shared-mysql-bundle RELEASE_TAG=<发布版本>
```

`make verify` 只执行测试、检查和构建，不启动服务。

生产Remote使用离线镜像交付。共享MySQL单独生成和部署离线包，Remote API、Web和Solver使用独立离线包；服务器仅加载镜像并以禁止拉取、禁止构建模式启动。

## 本机快捷命令

本机 zsh 提供以下快捷命令：

```bash
start_rgc
bg_rgc
stop_rgc
kill_rgc
```

- `start_rgc`：在前台启动本地 API 和前端。
- `bg_rgc`：在后台启动本地 API 和前端，并将日志写入 `$HOME/tmp/rudy_gc_run_all.log`。
- `stop_rgc`：正常终止后台进程并停止本地 API 和前端。
- `kill_rgc`：强制终止后台进程并停止本地 API 和前端。

这些快捷命令固定通过工作区根目录的公开目标 `make run-local` 和 `make stop-local` 运行。`make/workspace.mk` 负责将公开目标转发到 `local/` 内部目标；以后调整 `local/` 内部启动目标时，应同步更新该转发关系，避免本机快捷命令直接依赖内部目标。

`bg_rgc` 在创建后台进程和 PID 文件前会通过 dry-run 检查 `make run-local`；公开目标缺失时直接返回错误，避免输出错误的启动成功提示。
