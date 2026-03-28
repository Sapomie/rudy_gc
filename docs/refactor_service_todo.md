# service 直连 modelx 重构 TODO

## 目标
- 将项目整体改造成 `config -> dep -> router -> handler -> service -> modelx`
- 删除 `internal/infra`
- 删除 `internal/repo`
- 逐步迁空并删除旧业务实现目录

## 进度

### 阶段 1：基础骨架
- [x] 新建 `internal/config/load.go`
- [x] 新建 `internal/dep/dep.go`
- [x] 新建 `internal/router/router.go`
- [x] 新建 `internal/router/handler`
- [x] `cmd/api/main.go` 切到 `config -> dep -> router`

### 阶段 2：movie 域迁移
- [x] 新建 `internal/service/movie`
- [x] `movie` 主链路改为 `service -> modelx`
- [x] `MovieType` 聚合迁到 `internal/service/movie`
- [x] `movie agg / detail / rank day / need download` 入口迁到 `internal/service/movie`
- [x] 复杂列表查询迁到 `internal/model/modelx/moviex/movie_list_query.go`
- [x] 删除旧 `internal/domain/movie`
- [x] 删除旧 movie-list 的 `repo/infra`

### 阶段 3：sc 域迁移
- [x] 新建 `internal/service/sc` 兼容入口
- [x] handler / loop 对外入口改依赖 `internal/service/sc`
- [x] 将 `internal/domain/sc` 的真实实现迁入 `internal/service/sc`
- [ ] 将 `sc` 内部读写从 repo 改为直接使用 modelx
- [x] 删除旧 `internal/domain/sc`

### 阶段 4：vfilm 域迁移
- [x] 新建 `internal/service/vfilm`
- [x] 迁移目录扫描
- [x] 迁移影片处理
- [x] 迁移重命名与目录详情
- [x] 删除旧 `internal/domain/vfilm`

### 阶段 5：spider 域迁移
- [x] 新建 `internal/service/spider`
- [x] 迁移详情抓取
- [x] 迁移 inventory / daily best / sync best
- [x] 迁移 cover / title / rank 更新
- [x] 删除旧 `internal/domain/spider/logic`

### 阶段 6：handler 迁移
- [x] 将 `internal/transport/http/api` 迁到 `internal/router/handler`
- [x] 将 `internal/transport/http/html` 迁到 `internal/router/handler`
- [x] `internal/router/router.go` 只绑定新 handler
- [x] 删除旧 `internal/transport/http`

### 阶段 7：deps 去 repo/infra 化
- [x] `svc.Deps` 补充 modelx 句柄
- [x] 删除 `Deps` 中剩余 repo 字段
- [x] 删除 `NewDeps` 中 repo / infra 构造
- [x] 所有 service 改为只依赖 modelx / cache / channels / fetcher

### 阶段 8：最终清理
- [x] 删除 `internal/infra`
- [x] 删除 `internal/repo`
- [x] 将剩余 `internal/domain/loop` 迁到 `internal/service/loop`
- [x] 将剩余 `internal/domain/spider/fetcher` / `types` 迁到 `internal/service/spider`
- [x] 全局检索并清理旧 import / 旧方法 / 未使用代码
- [x] `gofmt`
- [x] `go test ./...`

## 当前进行中
- 已完成主链路与剩余 domain 代码迁移，正在做最后一轮验证
