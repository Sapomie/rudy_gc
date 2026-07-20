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
```

`make verify` 只执行测试、检查和构建，不启动服务。
