# Remote Docker 重写与迁移执行计划

## 1. 文档状态

- 编写日期：2026-07-22
- 目标服务器：`43.134.226.247`
- 目标部署目录：`/home/lighthouse/data/rudy_gc/remote/`
- 目标数据库：服务器现有 Docker MySQL 实例中的 `rudy_gc_collector`
- 计划状态：执行中；Docker运行链路已稳定，目标站Cloudflare仍拒绝服务器出口会话
- 本文档只定义执行步骤、验证标准和清理边界；实际代码重写、数据库迁移、部署和删除必须按阶段执行并逐项验收。

## 2. 改造目标

本次改造需要同时完成以下目标：

1. Remote API、Remote Web 和真实 Chromium 抓取环境全部改为 Docker 运行。
2. 服务器已有 Docker MySQL 实例继续承载 `rudy_gc_collector`，不创建新的 MySQL 容器。
3. 浏览器控制从 API 进程内的 `exec.Cmd` 生命周期判断，重写为独立 Browser Agent 容器和扩展长连接协议。
4. 浏览器状态持续可监控，Chromium、Xvfb、扩展或 Browser Agent 出现异常时能够自动恢复。
5. 按必填调度配置使用真实 Chromium 定时探测 DailyBest 目标页面，持续记录目标网站是否具备实际抓取条件。
6. 每天上海时间 `08:30` 执行 DailyBest 抓取。
7. `08:30` 抓取失败后，按照失败完成时间延后 1 小时创建恢复任务。
8. Remote API 重启后能够从数据库恢复当天尚未执行的重试计划。
9. 抓取成功后才原子发布影片、排名和日统计数据，失败时保留上一版完整数据。
10. Docker 版本验收完成后，删除服务器和仓库中的旧 Remote 非 Docker 部署实现。
11. 最终服务器只保留 Docker 版本、真实运行配置、浏览器 Profile Volume 和现有 MySQL 业务数据。

## 3. 范围与保留边界

### 3.1 本次重写范围

- `remote/api/`
- `remote/web/`
- `remote/deploy/`
- Remote 数据库迁移代码
- Remote 运行状态接口和页面
- Remote 抓取调度、目标网站定时探测、失败恢复、浏览器生命周期和运行日志
- `43.134.226.247` 上当前 Remote 非 Docker 部署

### 3.2 必须保留的内容

- 服务器已有 Docker MySQL 容器。
- `rudy_gc_collector` 数据库及现有业务数据。
- 现有数据库账号、密码和授权关系。
- 当前有效 Chromium Profile 中的 Cookie、`cf_clearance` 和浏览器状态。
- 已经发布的 Remote 排名、影片、抓取任务和请求日志。
- 当前仓库内与本次任务无关的用户修改。

### 3.3 最终需要删除的旧内容

- `rudy-gc-remote-api.service`
- `rudy-gc-remote-xvfb.service`
- Remote 专用旧 Nginx 站点配置和启用链接
- `/home/lighthouse/data/rudy_gc/remote/current`
- `/home/lighthouse/data/rudy_gc/remote/releases/` 下的旧 release
- 旧 Remote API 二进制
- 旧部署配置副本
- 旧扩展生成目录
- 旧 Profile 原目录（完成 Docker Volume 迁移并验证后）
- 旧 Remote 临时文件和项目专用日志
- 仓库中的旧 systemd、Xvfb、Nginx 部署文件
- 旧 `browser.go` 进程生命周期实现及对应测试

现有 MySQL 容器、`rudy_gc_collector` 数据库和迁移后的浏览器 Profile 不属于删除范围。

## 4. 目标系统架构

```text
访问者
  |
  v
remote-web :5032
  |  /api/*
  v
remote-api :5031
  |- DailyBest业务抓取
  |- DailyBest目标站定时探测
  |  Docker 内部 HTTP
  v
remote-browser :5033
  |- browser-agent
  |- Xvfb
  |- Chromium
  `- Rudy GC Chrome Extension

remote-api
  |
  `- existing-mysql-container:3306/rudy_gc_collector
```

### 4.1 容器职责

#### remote-web

- 构建并托管 Remote Web 静态资源。
- 对外监听 `0.0.0.0:5032`。
- 将 `/api/` 请求反向代理到 `remote-api:5031`。
- 提供Web进程检查接口。

#### remote-api

- 提供 Remote 业务 API。
- 执行数据库迁移。
- 管理 DailyBest 调度、抓取任务、目标站定时探测、失败恢复和数据发布。
- 通过 Docker 内部网络调用 Browser Agent。
- 记录抓取任务、页面请求日志、目标站探测记录和运行事件。
- 对外监听 `0.0.0.0:5031`。

#### remote-browser

- Browser Agent 作为容器主业务进程。
- Browser Agent 启动并监督 Xvfb 和 Chromium。
- Chromium 使用持久化 Profile Volume。
- Chrome 扩展与 Browser Agent 建立持久长连接。
- Browser Agent 接收 Remote API 的页面抓取请求。
- Browser Agent 对业务抓取和目标站探测使用同一条真实 Chromium 请求通道。
- 扩展使用 `chrome.tabs.create()` 创建页面标签。
- 页面完成或失败后返回状态、HTML 和错误，并关闭标签页。
- 监听 `0.0.0.0:5033`，端口只加入 Docker 内部网络，不映射公网。

#### 现有 MySQL 容器

- 继续存储 `rudy_gc_collector`。
- 通过现有外部 Docker 网络向 `remote-api` 提供连接。
- Remote Compose 不声明 MySQL service，不创建 MySQL Volume。

### 4.2 Docker 网络

计划使用两张网络：

1. `rudy-gc-remote-internal`
   - Remote Compose 创建。
   - `remote-web`、`remote-api`、`remote-browser`加入。
   - Browser Agent 端口只在该网络内可见。
2. 现有 MySQL 外部网络
   - 使用服务器核查得到的真实网络名称。
   - 在 Compose 中声明为 `external: true`。
   - 只有 `remote-api`加入。

数据库连接必须使用 MySQL 容器名或稳定网络别名，禁止使用容器动态 IP。

## 5. 服务器只读核查

执行任何部署或删除前，先完成以下只读核查并将结果记录在本计划的执行记录中。

### 5.1 Docker 环境

- Docker Engine 版本。
- Docker Compose 版本。
- 当前运行容器清单。
- MySQL 容器真实名称、镜像、运行状态和健康状态。
- MySQL 容器所属网络及网络别名。
- MySQL 数据持久化位置。
- 当前磁盘剩余空间。

建议核查命令：

```bash
docker version
docker compose version
docker ps --format '{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'
docker inspect <实际MySQL容器名>
docker network ls
docker system df
df -h
```

执行记录中必须替换并记录真实 MySQL 容器名和网络名，运行配置不得保留占位值。

### 5.2 MySQL 环境

- `rudy_gc_collector` 是否存在。
- 实际连接用户和密码是否可用。
- 用户是否具备迁移和业务读写权限。
- 当前表清单、数据量和索引状态。
- 当前字符集、时区和连接参数。
- 现有 Remote 数据的最新日期、最后抓取任务和最后请求日志。

### 5.3 旧 Remote 部署

- `rudy-gc-remote-api.service` 的真实路径、状态和启动参数。
- `rudy-gc-remote-xvfb.service` 的真实路径和状态。
- Remote API、Xvfb、Chromium 进程及其用户、PID 和 cgroup。
- Remote 专用 Nginx 配置文件和启用链接。
- `/home/lighthouse/data/rudy_gc/remote/` 下的目录、链接和文件清单。
- 当前生效 `server.yaml`。
- 当前 Chromium Profile 真实路径、属主、权限和大小。
- 当前 Docker 端口占用情况。

### 5.4 凭证核查

以下真实值必须在部署前完整确认并写入服务器实际配置：

- MySQL DSN。
- Browser Agent 内部访问凭证。
- Chromium Profile 路径。
- Remote API 与 Web 对外端口。
- Docker 外部 MySQL 网络名。
- noVNC 访问凭证（启用人工浏览器维护时）。

所有密码、令牌和连接信息必须为真实生效值，配置文件权限必须限制为部署用户可读。

## 6. 数据与运行状态备份

### 6.1 数据库备份

在首次迁移前，对 `rudy_gc_collector`执行完整逻辑备份，并验证备份可读取：

- 备份文件包含表结构、索引和数据。
- 备份文件记录生成时间和源 MySQL 容器信息。
- 备份文件保存在明确的临时回滚目录。
- 完成 Docker 版本最终验收后，按照清理阶段要求删除临时回滚文件。

### 6.2 Browser Profile备份

- 停止旧 Remote API 和旧 Chromium 写入前，记录 Profile 文件清单。
- 制作一次权限和软链接均可验证的 Profile 临时副本。
- 记录 Profile 原目录、属主、权限和总大小。
- 复制期间保证旧 Chromium 已停止，避免产生不一致的 SQLite 和 Cookie 文件。

### 6.3 配置与部署清单备份

- 保存当前生效 `server.yaml`的临时副本。
- 保存旧 systemd unit 和 Remote 专用 Nginx 配置的临时副本。
- 记录当前 `current`链接目标和 release 名称。
- 记录当前 API、Web 和 Chromium 版本信息。

临时回滚副本只用于本次切换，最终验收完成后必须清除。

## 7. 代码重写计划

### 7.1 Browser Agent二进制

在 `remote/api`同一个 Go module 中新增：

```text
remote/api/cmd/browser-agent/
```

Browser Agent负责：

- 加载独立配置。
- 启动和监督 Xvfb。
- 启动和监督 Chromium。
- 生成并加载 Chrome 扩展。
- 接收扩展长连接。
- 向 Remote API提供内部抓取接口。
- 维护浏览器状态机和请求状态。
- 暴露 `/healthz`、`/readyz`和内部状态接口。
- 输出 Chromium、Xvfb 和扩展生命周期日志。

### 7.2 Browser Agent分层目录

```text
remote/api/internal/config/
remote/api/internal/repo/browseragent/
remote/api/internal/router/handler/browseragent/
remote/api/internal/service/browseragent/
remote/api/internal/types/browseragent/
```

职责约束：

- `config`只负责配置结构、加载和完整性校验。
- `repo/browseragent`负责进程、扩展文件、Profile 和底层连接。
- `service/browseragent`负责状态机、请求编排、恢复和并发控制。
- `router/handler/browseragent`只负责内部 HTTP/WebSocket 协议。
- `types/browseragent`放无基础设施依赖的状态和协议结构。

### 7.3 删除旧浏览器实现

新实现通过测试并完成调用切换后，删除：

```text
remote/api/internal/repo/dailybest/browser.go
remote/api/internal/repo/dailybest/browser_test.go
```

同时删除以下旧行为：

- API直接启动并持有Chromium命令。
- `commandDone`作为浏览器健康状态。
- 每页执行一次`chromium --new-tab`。
- Chrome stdout/stderr写入`io.Discard`。
- API内部回环Browser Agent服务。

### 7.4 新浏览器协议

Browser Agent与扩展之间使用持久长连接。协议至少包含：

- Agent注册。
- Agent心跳。
- Agent状态变化。
- 创建页面请求。
- 页面请求接受。
- 页面完成。
- 页面失败。
- 取消请求。
- 关闭标签页。

每个页面请求必须携带唯一请求ID、目标URL和截止时间。Browser Agent必须保证：

- 同一请求只完成一次。
- 连接断开时所有挂起请求都得到明确失败结果。
- 页面完成后清理请求状态和标签页。
- Agent重连后不恢复已经失败的旧请求。

### 7.5 浏览器健康状态机

```text
starting
  -> ready
  -> failed

ready
  -> fetching
  -> disconnected

fetching
  -> ready
  -> disconnected
  -> failed

disconnected
  -> recovering

recovering
  -> ready
  -> fatal
```

状态判断依据：

- 扩展长连接状态。
- 心跳和请求响应。
- Chromium进程事件。
- Xvfb进程事件。
- 当前活动请求。

Chromium启动命令退出只记录进程事件；浏览器可用性以扩展连接和实际请求通道为准。

心跳周期、失联判定时间和启动等待时间必须作为必填配置提供，代码不设置隐藏默认值。

### 7.6 Remote API调用Browser Agent

Remote API中的DailyBest Source改为调用：

```text
http://remote-browser:5033
```

API客户端职责：

- 使用真实内部访问凭证。
- 发送目标URL和请求ID。
- 接收HTTP状态、HTML、耗时和错误类型。
- 将每次实际页面请求写入`d_dailybest_fetch_log`。
- Browser Agent连接失败时返回明确的可诊断错误。

### 7.7 DailyBest调度与一小时恢复

每天定时行为：

```text
Asia/Shanghai 08:30
  -> 检查当天成功和活动任务
  -> 创建scheduled任务
  -> 执行抓取
```

失败恢复行为：

```text
scheduled/recovery失败
  -> fetch_run记录failed和finished_at
  -> retry_at = finished_at + 1小时
  -> 到时再次检查当天成功和活动任务
  -> 创建trigger_type=recovery任务
```

恢复约束：

- `succeeded`或`no_change`结束当天恢复。
- `queued`或`running`阻止重复创建活动任务。
- Recovery再次失败后继续按照失败完成时间延后1小时。
- 当前业务日期变化后结束上一业务日恢复。
- Manual任务不进入scheduled/recovery自动恢复链。
- 服务启动时从数据库查询当天最后一次任务并恢复剩余等待时间。
- `FetchOnStart`改为当天状态协调，已有成功结果时跳过抓取。

### 7.8 数据发布

保持整批发布原则：

- 所有页面抓取完成后解析。
- 校验重复页面、重复JavID、重复名次、日期和checksum。
- 完整结果与当前checksum一致时记录`no_change`。
- 结果发生变化时在数据库事务中发布影片、排名和日统计。
- 页面抓取、解析或校验失败时不发布部分数据。

### 7.9 DailyBest目标站定时探测

Remote API增加独立的目标站探测调度。探测固定通过FlareSolverr真实浏览器访问：

```text
https://www.c97k.com/cn/vl_bestrated.php?&mode=&page=1
```

一次探测按以下顺序执行：

1. 等待FlareSolverr请求通道可用。
2. 生成唯一请求ID，并在真正发起请求前创建`running`探测记录。
3. 通过FlareSolverr持久会话打开目标页面并等待页面结果。
4. 记录HTTP状态、最终地址、耗时、浏览器状态和HTML检查结果。
5. 将数据库记录更新为`succeeded`或`failed`。

成功必须同时满足：

- 页面请求成功并取得HTTP 200。
- 最终地址仍属于配置的DailyBest目标站。
- 页面存在`#rightcolumn`。
- 页面存在`videothumblist`排行列表结构。
- 页面未命中Cloudflare验证页、访问拒绝页或其他非排行页面。
- FlareSolverr成功返回可解析HTML。

探测约束：

- 探测只验证目标网站的真实可抓取性，不解析和发布影片、`c_rank`或`c_rank_day_stat`。
- 每次实际访问目标网站都必须对应一条`d_dailybest_target_probe`记录。
- 探测失败只更新目标网站状态并记录错误，不将外部网站故障判定为本地容器故障。
- 实际DailyBest抓取与探测通过同一FlareSolverr客户端串行使用持久会话，避免并发修改同一会话状态。
- 探测到时若业务抓取正在执行，等待业务抓取结束后再执行，并记录真实开始时间和结束时间。
- API启动时将上一进程遗留的`running`探测记录更新为`interrupted`，立即执行一次真实探测；探测完成后按配置周期继续调度，避免重启后的监控盲区。
- 探测周期和时区必须由真实运行配置显式提供，代码不设置隐藏默认值。
- 运行状态页面的查询和刷新只读取数据库及本地运行状态，不触发目标网站探测。

## 8. 数据库调整计划

### 8.1 复用现有表

- `d_dailybest_fetch_run`
- `d_dailybest_fetch_log`
- `d_dailybest_movie`
- `c_rank`
- `c_rank_day_stat`

### 8.2 运行事件表

新增轻量表：

```text
d_dailybest_runtime_event
```

该表只记录状态变化，不记录每次心跳。需要覆盖：

- Browser Agent启动。
- Xvfb启动和退出。
- Chromium启动和退出。
- 扩展连接和失联。
- 浏览器恢复开始、成功和失败。
- API到Browser Agent连接异常和恢复。
- 容器级恢复事件。

字段设计必须保持最小化，至少包含组件、事件类型、级别、消息、发生时间，以及可选的关联抓取任务ID或目标站探测ID。

### 8.3 目标站探测表

新增：

```text
d_dailybest_target_probe
```

该表每行对应一次真实Chromium目标站探测，字段包含：

- `id`：主键。
- `request_id`：发送给Browser Agent的唯一请求ID，唯一索引。
- `status`：`running`、`succeeded`、`failed`或`interrupted`。
- `target_url`：实际请求地址。
- `final_url`：页面完成后的最终地址。
- `http_status`：目标页面HTTP状态。
- `duration_ms`：真实页面请求耗时。
- `browser_status`：请求完成时Browser Agent报告的浏览器状态。
- `error_type`：网络、Cloudflare、页面结构、浏览器或内部错误分类。
- `error_message`：可直接用于排查的错误信息。
- `started_at`：真正开始目标请求的时间。
- `finished_at`：请求完成、失败或中断的时间。

索引至少覆盖按开始时间倒序分页，以及按状态和开始时间查询。探测数据不设置默认保留期限、数量上限或自动截断。

### 8.4 抓取任务查询能力

ModelX需要提供：

- 查询指定业务日期和分类的成功任务。
- 查询指定业务日期和分类的活动任务。
- 查询指定业务日期和分类的最后失败任务。
- 查询需要恢复的一小时重试任务。
- 原子检查并创建任务，防止定时、恢复和手动任务重复排队。

### 8.5 迁移要求

- 所有迁移必须可重复执行。
- 迁移前确认数据库备份完成。
- 迁移后核对表结构、索引和已有数据量。
- 核对`d_dailybest_target_probe`能够完成开始记录、结果更新和分页查询。
- 迁移失败时API启动失败并输出明确错误。
- 迁移不得删除现有有效排名、影片、抓取任务和请求日志。

## 9. Docker构建与部署文件

目标目录：

```text
remote/deploy/docker/
├── compose.yaml
├── api.Dockerfile
├── browser.Dockerfile
├── web.Dockerfile
├── nginx.conf
└── README.md
```

### 9.1 api.Dockerfile

- 多阶段构建Remote API。
- 最终镜像只包含API二进制、CA证书和运行所需文件。
- 使用非root用户运行。
- 配置通过只读挂载注入。
- 提供API进程与就绪检查接口。

### 9.2 browser.Dockerfile

- 安装固定来源的Chromium、Xvfb和运行字体。
- 构建并安装Browser Agent二进制。
- 使用init进程正确回收子进程。
- 使用非root用户运行Browser Agent、Xvfb和Chromium。
- Profile固定挂载到持久化路径。
- Chrome、Xvfb和Agent输出进入容器标准输出和错误输出。
- 提供Browser Agent进程与就绪检查接口。
- 可选noVNC入口只绑定服务器回环地址，并通过SSH隧道访问。

### 9.3 web.Dockerfile

- 构建Remote Web。
- 最终镜像使用Nginx托管静态文件。
- `/api/`反向代理到`remote-api:5031`。
- 对外提供`5032`。
- 提供Web进程检查接口。

### 9.4 compose.yaml

必须包含：

- `remote-api`
- `remote-browser`
- `remote-web`
- 内部网络
- 现有MySQL外部网络
- Browser Profile持久化Volume
- 真实配置文件只读挂载
- 进程退出重启策略
- 容器重启策略
- 明确的依赖和就绪条件
- 资源及日志配置使用项目确认后的真实值

Compose中不得声明MySQL service。

## 10. 运行配置

服务器真实配置目录：

```text
/home/lighthouse/data/rudy_gc/remote/config/
```

API配置必须包含：

- 现有MySQL容器真实DSN。
- `0.0.0.0:5031`监听地址。
- `Asia/Shanghai`时区。
- `08:30`每日抓取时间。
- `1h`失败恢复延迟。
- Browser Agent Docker内部地址。
- Browser Agent真实访问凭证。
- DailyBest目标站探测地址。
- DailyBest目标站探测调度表达式。
- DailyBest目标站探测时区。

Browser Agent配置必须包含：

- `0.0.0.0:5033`监听地址。
- Chromium真实可执行文件路径。
- Profile Volume挂载路径。
- 扩展目录。
- Xvfb显示地址。
- 页面请求超时。
- 心跳周期。
- 失联判定时间。
- 启动等待时间。
- API内部访问凭证。

配置校验必须拒绝空密码、空令牌、空路径、占位值和不可达连接信息。

## 11. Browser Profile迁移

### 11.1 迁移步骤

1. 记录旧Profile路径、属主、权限和大小。
2. 停止旧Remote API和旧Chromium，确保Profile不再写入。
3. 创建Docker Browser Profile Volume。
4. 将Profile完整复制到Volume实际挂载目录。
5. 调整属主为Browser容器实际运行UID/GID。
6. 启动`remote-browser`。
7. 验证扩展建立长连接。
8. 验证Profile中的Cookie和`cf_clearance`可被Chromium读取。
9. 通过Browser Agent执行一次受控页面抓取。
10. 验证成功后保留Volume并进入API联调。

### 11.2 人工维护入口

noVNC入口只允许绑定服务器回环地址。访问方式：

```text
本机浏览器
  -> SSH隧道
  -> 服务器127.0.0.1上的noVNC
  -> 容器内真实Chromium
```

人工完成Cloudflare校验后，结果直接写入同一个Profile Volume。

## 12. 实时监控

### 12.1 健康接口

Remote API：

- `/healthz`：API进程存活。
- `/readyz`：数据库、Browser Agent和调度器可用。
- `/api/v1/dailybest/runtime-status`：当前运行状态。
- `/api/v1/dailybest/runtime-events`：运行事件列表。
- `/api/v1/dailybest/target-probes`：按条件分页查询目标站探测记录。

Browser Agent：

- `/healthz`：Agent进程存活。
- `/readyz`：Xvfb、Chromium和扩展长连接可用。

API运行状态必须分别返回：

- `browser_status`：Browser Agent、Xvfb、Chromium和扩展组成的本地浏览器状态。
- `target_status`：最近一次定时探测得到的目标网站状态。
- 最近一次成功探测时间。
- 最近一次失败探测时间和错误摘要。
- 下一次计划探测时间。

目标站探测失败不改变`/readyz`结果；`/readyz`只反映Remote自身依赖是否具备接收任务的能力。

### 12.2 目标网站定时监控

- Remote API按照必填调度配置创建探测任务。
- 探测请求通过Browser Agent中的真实Chromium和持久化Profile执行。
- 每次请求开始前先创建数据库记录，完成后更新结果。
- 目标网站访问失败时写入探测记录和运行事件，并在状态页面展示。
- 目标网站恢复后，下一次成功探测更新`target_status`并写入恢复事件。
- 探测失败等待下一次配置调度，不重启`remote-api`或`remote-browser`容器。
- 目标站探测结果不触发容器退出，避免Cloudflare、网络抖动或目标站维护引发容器重启循环。

### 12.3 状态页面

Remote Web增加运行状态页面，展示：

- API状态。
- 数据库状态。
- Browser Agent状态。
- Xvfb状态。
- Chromium状态。
- 扩展连接状态。
- 最近心跳。
- 当前活动请求。
- 最近抓取成功时间。
- 最近失败和恢复事件。
- 目标网站当前状态。
- 最近一次目标站探测时间、耗时、HTTP状态和最终地址。
- 最近一次目标站探测错误。
- 下一次目标站探测时间。
- 目标站探测记录筛选和分页列表。

页面进入时只展示页面结构；运行状态和探测记录由用户明确触发刷新或查询后读取。页面操作只查询Remote数据库及运行状态，禁止触发目标网站请求。

### 12.4 Docker监控

- `remote-api`和`remote-browser`保留就绪接口供人工和页面查询。
- `remote-web`保留静态站点进程检查接口。
- Docker重启策略负责容器级恢复。
- `docker compose ps`和`docker compose logs`能够直接查看当前状态和完整日志。
- Browser Agent持续监督Xvfb、x11vnc、noVNC、Chromium扩展连接，关键故障时退出进程并由Docker重启。
- 目标网站连接能力由独立定时探测检查并写入数据库。

## 13. 测试计划

### 13.1 Browser Agent单元测试

- 扩展注册和鉴权。
- 心跳更新。
- 页面请求创建和完成。
- 重复完成防护。
- 请求超时。
- 扩展断开时取消挂起请求。
- Chromium启动命令正常退出但扩展仍连接时保持ready。
- Chromium真实退出时状态转换。
- Xvfb退出时状态转换。

### 13.2 Scheduler单元测试

- 每天上海时间08:30创建scheduled任务。
- 当天已有成功任务时跳过。
- 当天已有活动任务时跳过。
- scheduled失败后1小时创建recovery任务。
- recovery失败后再次延后1小时。
- 成功或no_change后停止恢复。
- 业务日期变化后停止上一日恢复。
- API重启后恢复剩余等待时间。
- Manual任务不进入自动恢复链。
- API启动后立即创建目标站探测，完成后按配置周期继续创建。
- 业务抓取与探测同时到期时业务抓取优先。
- 业务抓取占用Chromium时探测等待并在空闲后执行。
- 探测失败不进入DailyBest业务抓取的一小时恢复链。

### 13.3 Repository和迁移测试

- 成功、活动、最后失败任务查询。
- 原子任务创建和并发去重。
- 运行事件写入和分页查询。
- 目标站探测开始记录、成功更新、失败更新和分页查询。
- API重启时将上一进程遗留的`running`探测更新为`interrupted`。
- 数据库迁移重复执行。
- 发布事务失败回滚。

### 13.4 Docker集成测试

- 三个容器进程正常启动并保持运行。
- API通过外部网络连接现有MySQL。
- Browser Agent启动Xvfb和Chromium。
- 扩展建立长连接。
- API通过Browser Agent完成一页抓取。
- 定时探测通过Browser Agent真实访问目标页面并写入数据库。
- 目标页面返回Cloudflare验证页时探测记录为失败。
- 目标站探测失败时三个容器保持运行，业务调度继续运行。
- Browser容器重启后Profile保留。
- Chromium进程终止后Browser容器恢复。
- Xvfb进程终止后Browser容器恢复。
- API容器重启后恢复重试计划。
- Web能够展示抓取、请求日志和运行状态。

### 13.5 长时间运行验证

- 保持容器持续运行并观察Browser Agent心跳。
- 在无抓取请求期间确认扩展连接持续有效。
- 到达下一次调度前确认Browser状态保持ready。
- 跨越多个探测调度周期，核对每次真实请求都有独立数据库记录。
- 受控阻断目标站连接，验证探测失败记录、状态展示和容器持续运行。
- 恢复目标站连接，验证后续探测成功并记录恢复事件。
- 验证一次真实08:30定时抓取。
- 受控制造一次失败并验证1小时后的recovery任务。

## 14. 构建与本地验证

代码完成后执行：

```text
Remote API和Browser Agent：
  gofmt
  go test ./...
  go vet ./...
  go build ./cmd/api
  go build ./cmd/browser-agent

Remote Web：
  类型检查
  单元测试
  生产构建

Docker：
  docker compose config
  构建全部镜像
  启动内部联调环境
  检查三个健康状态
```

任何失败必须在进入服务器切换阶段前修复。

## 15. 服务器部署步骤

### 阶段A：准备新版本

1. 完成服务器和MySQL只读核查。
2. 完成数据库、Profile和配置临时备份。
3. 创建Docker部署目录和真实配置。
4. 核对配置中全部凭证和连接信息。
5. 构建带明确版本标识的三个镜像。
6. 执行`docker compose config`验证最终配置。

### 阶段B：启动Browser容器

1. 停止旧Remote API和旧Chromium写入。
2. 创建并导入Browser Profile Volume。
3. 启动`remote-browser`。
4. 检查Xvfb、Chromium、扩展连接和健康接口。
5. 检查容器日志中不存在启动错误和空凭证。
6. 执行受控页面抓取并验证Cookie状态。

### 阶段C：启动API容器

1. 将`remote-api`加入现有MySQL外部网络。
2. 验证容器名能够解析到MySQL。
3. 验证真实DSN连接成功。
4. 启动`remote-api`并执行迁移。
5. 检查API、数据库、Browser Agent和调度器状态。
6. 验证抓取任务、请求日志、目标站探测和运行事件接口。
7. 确认API启动后立即完成首条目标站探测，并核对下一次探测时间为完成后一个配置周期。

### 阶段D：启动Web容器

1. 启动`remote-web`。
2. 验证`5032`页面。
3. 验证Web到API反向代理。
4. 验证抓取任务、请求日志、排名、目标站探测和运行状态页面。
5. 检查浏览器控制台和网络请求。
6. 验证进入页面时不触发目标网站请求，用户查询后只读取Remote API已有状态和数据库记录。

### 阶段E：业务验收

1. 手动创建一次抓取任务。
2. 验证全部有效页面和结束空页请求日志。
3. 验证fetch run最终状态。
4. 验证影片、`c_rank`和`c_rank_day_stat`。
5. 再次抓取并验证`no_change`。
6. 受控制造一次失败。
7. 验证失败记录、运行事件和一小时恢复任务。
8. 验证容器重启后的Profile和调度恢复。
9. 验证启动时真实探测成功、数据库记录完整，并核对周期调度时间。
10. 受控验证一次目标站探测失败，确认错误入库且容器不重启。
11. 恢复目标站访问后验证下一次探测记录成功和目标状态恢复。

所有业务验收完成后才能进入旧版本清理阶段。

## 16. 旧版本清理步骤

### 16.1 停止和注销systemd服务

1. 停止`rudy-gc-remote-api.service`。
2. 停止`rudy-gc-remote-xvfb.service`。
3. 禁用两个服务的开机启动。
4. 删除对应unit文件。
5. 执行systemd daemon reload。
6. 验证服务列表中不再存在旧Remote unit。

### 16.2 清理旧进程

- 核对旧API、Xvfb和Chromium进程。
- 只终止属于旧Remote部署的进程。
- 验证Docker容器中的新进程保持运行。
- 验证旧端口占用已经释放或由Docker版本接管。

### 16.3 清理旧Nginx配置

- 定位Remote专用旧配置文件。
- 删除Remote专用启用链接。
- 删除Remote专用旧配置。
- 检查Nginx配置语法。
- 重载Nginx时保留服务器其他站点。

### 16.4 清理旧文件

按服务器核查阶段记录的精确路径逐项删除：

- 旧`current`链接。
- 旧`releases`目录内容。
- 旧二进制。
- 旧配置副本。
- 旧扩展目录。
- 已完成迁移的旧Profile目录。
- 旧Remote临时文件和项目专用日志。

删除命令必须使用已经核对的精确路径，禁止对`/home/lighthouse/data/rudy_gc/remote/`根目录执行递归删除，因为该目录同时承载新Docker部署配置。

### 16.5 清理仓库旧实现

- 删除旧systemd unit。
- 删除旧Xvfb unit。
- 删除旧Remote Nginx配置。
- 删除旧浏览器生命周期代码和测试。
- 更新Remote README和部署文档为Docker版本。
- 更新Make入口，统一提供构建、测试、启动、停止、状态和日志命令。

### 16.6 清理临时回滚文件

Docker版本完成最终验收后：

- 删除本次切换产生的临时数据库备份。
- 删除旧Profile临时副本。
- 删除旧配置和unit临时副本。
- 删除无引用的旧Remote镜像。
- 保留正式MySQL数据、正式Browser Profile Volume和当前版本镜像。

## 17. 回滚方案

旧版本清理前保留一次完整回滚能力。

### 17.1 触发回滚的条件

- 新API无法稳定连接现有MySQL。
- Browser Agent无法稳定保持扩展连接。
- Profile迁移后Cookie不可用且无法恢复。
- 数据库迁移导致业务接口异常。
- Docker网络或端口切换无法满足服务访问。
- 抓取结果、日志或排名发布出现完整性问题。

### 17.2 回滚步骤

1. 停止新Remote Compose。
2. 恢复切换前数据库备份。
3. 恢复旧Profile目录。
4. 恢复旧生效配置、systemd unit和Remote专用Nginx配置。
5. 启动旧Xvfb和旧API。
6. 验证旧API、Web、数据库和一次抓取。
7. 保留新Docker日志用于问题定位。

旧版本正式删除后，本次临时回滚通道结束；后续版本回退使用Docker镜像、数据库正式备份和Profile Volume快照。

## 18. 最终验收标准

### 18.1 容器和网络

- `remote-api`进程正常且就绪接口可用。
- `remote-browser`进程正常且就绪接口可用。
- `remote-web`进程正常且静态站点可用。
- Remote Compose中不存在MySQL service。
- API通过现有MySQL外部网络连接真实MySQL容器。
- Browser Agent端口只在Docker内部网络可达。

### 18.2 浏览器可靠性

- Chromium启动器正常退出不会错误判定浏览器失效。
- 扩展长连接能够持续保持。
- Chromium真实退出能够被立即发现。
- Xvfb退出能够被立即发现。
- Browser容器能够自动恢复。
- Profile在容器重启后保持。
- Chrome和Xvfb退出原因能够从容器日志查看。

### 18.3 抓取和恢复

- 08:30定时任务能够执行。
- 每页请求都有`d_dailybest_fetch_log`记录。
- 抓取失败能够记录明确原因。
- 失败后一小时创建recovery任务。
- API重启能够恢复重试等待。
- 成功或`no_change`结束当天恢复。
- 抓取失败不发布部分排名。

### 18.4 目标网站监控

- 按真实配置的调度时间使用真实Chromium探测目标页面。
- API启动后立即执行一次目标站探测；页面启动只展示页面，不触发目标站访问。
- 每次目标站探测都有一条`d_dailybest_target_probe`记录。
- HTTP状态、最终地址、页面结构、Cloudflare页面和HTML返回均参与成功判断。
- `browser_status`和`target_status`独立展示。
- 目标站失败会记录错误和运行事件，Remote容器保持运行。
- 业务抓取与探测冲突时业务抓取优先，两个请求不会并发操作Profile。
- 探测不写入或修改影片、排名和日统计数据。
- 探测记录能够按条件分页查询，且不存在默认截断。

### 18.5 数据

- `rudy_gc_collector`现有数据完整。
- 数据库迁移成功。
- 影片、排名和日统计发布正确。
- 抓取任务、请求日志、目标站探测和运行事件可查询。
- 第二次相同抓取正确识别`no_change`。

### 18.6 页面和监控

- Remote Web可访问。
- 抓取任务、详情、请求日志、排名和日统计页面可用。
- 运行状态和运行事件页面可用。
- 目标网站状态和探测记录页面可用。
- 页面进入时只展示页面结构。
- 用户触发后能够读取当前运行状态和数据库中的探测记录。
- 页面查询和刷新不会触发目标网站请求。

### 18.7 旧版本清理

- 旧Remote API systemd服务已删除。
- 旧Xvfb systemd服务已删除。
- Remote专用旧Nginx配置已删除。
- 旧release、二进制、配置副本、扩展和Profile原目录已删除。
- 旧API、Xvfb和Chromium进程已清理。
- 仓库旧部署和浏览器生命周期代码已删除。
- 临时回滚文件已删除。
- 服务器只保留Docker版Remote和现有Docker MySQL业务数据。

## 19. 执行记录

实际执行时按以下格式记录每个阶段：

```text
阶段：
执行时间：
执行人：
目标主机：
实际文件或容器：
执行命令：
结果：
验证证据：
遗留问题：
回滚点：
```

任何阶段出现未通过项时，停止进入下一阶段，保留现场和日志，按照本计划的回滚边界处理。

### 19.1 实施基线核查

```text
阶段：实施基线核查
执行时间：2026-07-22 16:16 Asia/Shanghai
执行人：Codex
目标主机：本机、43.134.226.247
实际文件或容器：remote/api、remote/web/new、rudy_mysql_remote、旧 Remote systemd 服务和部署目录
执行命令：go test ./...、go vet ./...、go build；npm check/test/build；服务器 Docker、网络、systemd 和目录只读核查
结果：本地 API 与 Web 基线验证通过；服务器 Docker 和 MySQL 正常；旧 Remote API 与 Xvfb 服务仍在运行
验证证据：Go 全部包测试通过；Vitest 5 个测试文件、11 个用例通过；Vite 构建成功；rudy_mysql_remote 为 healthy
遗留问题：服务器没有 docker compose 子命令，现有 /usr/local/bin/docker-compose 版本为 1.28.6；新 Compose 文件必须兼容 Compose V1
回滚点：本阶段未修改服务器、数据库或运行状态
```

服务器基线明细：

- Docker Server：`20.10.5`。
- Compose：`docker-compose 1.28.6`。
- MySQL容器：`rudy_mysql_remote`，镜像`mysql:8.0`，状态`healthy`。
- 可复用外部网络候选：`rudy_remote_infra`。
- 旧服务：`rudy-gc-remote-api.service`和`rudy-gc-remote-xvfb.service`均为`active`。
- 旧部署根目录：`/home/lighthouse/data/rudy_gc/remote/`。
- 当前浏览器Profile：`browser-profile-chromium`；另有旧`browser-profile`需要在迁移验收后判定并清理。
- 当前release目录和`current`链接仍存在，进入最终清理阶段前保留。

### 19.2 Browser Agent、监控与恢复代码重写

```text
阶段：Browser Agent、监控与恢复代码重写
执行时间：2026-07-22 16:20-17:12 Asia/Shanghai
执行人：Codex
目标主机：本机
实际文件或容器：remote/api、remote/web/new、remote/deploy/docker、remote/make
执行命令：gofmt、go test -race ./...、go vet ./...、go build；npm check/test/build；docker-compose config
结果：独立 Browser Agent、目标站探测、一小时恢复、运行状态接口与页面、Docker 三容器定义已经完成并通过本地静态和自动化验证
验证证据：Go 全量竞态测试通过；API、Browser Agent、Migrate 构建成功；Web 7 个测试文件、14 个用例通过；Vite 构建成功；Compose 配置解析成功
遗留问题：服务器运行配置仍需写入新的必填调度、心跳、断连、启动和健康检查参数；真实Profile尚未迁移
回滚点：服务器、数据库和旧systemd服务尚未修改
```

本阶段完成内容：

- 新增独立`cmd/browser-agent`二进制。
- Browser Agent负责Xvfb和Chromium进程、扩展生成、长连接、心跳、串行业务请求、探测请求和故障状态。
- 删除API进程内Chromium启动、`commandDone`健康判断和每页执行`--new-tab`的旧实现。
- API通过带鉴权的Docker内部HTTP客户端调用Browser Agent。
- 业务抓取请求优先于目标站探测请求，两类请求不会并发操作同一个Profile。
- 新增`d_dailybest_runtime_event`和`d_dailybest_target_probe`的ModelG、ModelX及迁移。
- 每次目标站探测先创建数据库记录；数据库记录创建失败时阻止目标网站请求。
- 增加目标页面HTTP、最终地址、`#rightcolumn`、`videothumblist`和Cloudflare页面校验。
- 删除`FetchOnStart`，页面启动不会立即访问目标网站；API启动时只执行独立目标站探测，不创建排行抓取任务。
- 增加`finished_at + 1h`失败恢复协调器，API重启后从数据库恢复等待。
- 增加Browser Agent状态变化入库监控；只记录状态变化，不写心跳流水。
- 增加运行状态、运行事件和目标站探测查询接口。
- 增加运行状态和目标站探测页面，首次进入不请求接口，用户操作只读取Remote API已有状态和数据库记录。
- 增加兼容服务器`docker-compose 1.28.6`的`remote-browser`、`remote-api`和`remote-web`定义。
- Compose复用外部网络`rudy_remote_infra`中的MySQL别名`mysql`，未声明MySQL服务。

本阶段已修复问题：

- Dockerfile最初引用不存在的`shared/go/go.sum`，已删除错误复制项并重新验证。
- API内旧`FetchOnStart`仍会启动即抓取，已从配置结构、入口和调度器删除。
- Browser Agent失联监控原先可能在首次启动阶段提前使用断连时间终止，已调整为扩展首次就绪后再启动失联监控。
- 前端新增文件首次暂存命令使用了错误工作目录相对路径，已立即使用正确路径暂存并在格式化后再次更新暂存内容。
- 本机执行三镜像构建时，Docker Hub鉴权端点`auth.docker.io:443`连接超时，构建在读取基础镜像元数据阶段停止，尚未执行Dockerfile中的代码构建步骤；未修改本机或服务器代理，后续改由目标服务器验证基础镜像拉取与正式构建。

### 19.3 目标服务器镜像构建

```text
阶段：目标服务器镜像构建
执行时间：2026-07-22 17:18-17:32 Asia/Shanghai
执行人：Codex
目标主机：43.134.226.247
实际文件或容器：/home/lighthouse/data/rudy_gc/remote/releases/docker-20260722-171808、docker_remote-api:latest、docker_remote-browser:latest、docker_remote-web:latest
执行命令：上传受控构建上下文；docker-compose build remote-browser remote-api remote-web；docker images核查
结果：三类镜像全部构建成功；旧systemd服务、旧release、数据库和运行配置均未修改
验证证据：remote-browser镜像ce1caee468fa（874MB）；remote-api镜像b0af2670ab02（123MB）；remote-web镜像fb4121c5f670（48.4MB）
遗留问题：镜像入口、实际数据库迁移、Profile迁移、真实运行配置和三容器健康状态仍需验证
回滚点：删除本次新release与新镜像即可回到构建前状态；旧服务仍保持运行
```

本阶段使用的基础镜像均已在目标服务器成功拉取并完成构建，未修改目标服务器代理配置。构建上下文只包含`remote/`、`shared/go/`和根`.dockerignore`，未上传本地其他项目、Git目录或构建产物。

### 19.4 切换前备份与镜像入口核查

```text
阶段：切换前备份与镜像入口核查
执行时间：2026-07-22 17:35 Asia/Shanghai
执行人：Codex
目标主机：43.134.226.247
实际文件或容器：/home/lighthouse/data/rudy_gc/remote/backups/pre-docker-20260722-173527、rudy_mysql_remote、三类新镜像
执行命令：mysqldump单事务备份；gzip完整性校验；配置和旧服务文件备份；镜像-help入口测试；docker image inspect
结果：数据库、配置、systemd、Nginx和current链接回滚材料已生成；API、Browser Agent镜像入口可执行；三类镜像入口、命令和运行用户符合定义
验证证据：数据库压缩备份53KB且gzip -t通过；备份时c_rank=476、c_rank_day_stat=1、d_dailybest_fetch_log=58、d_dailybest_fetch_run=10、d_dailybest_movie=476；API和Browser Agent均正常输出命令帮助
遗留问题：当前服务器镜像尚未包含17:32后完成的连接身份、结果重投和遗留任务恢复修复，需要重新上传并构建；Profile一致性备份需在停止旧服务后完成
回滚点：/home/lighthouse/data/rudy_gc/remote/backups/pre-docker-20260722-173527
```

本阶段没有停止或重启旧Remote服务，没有执行数据库迁移，也没有修改代理。数据库备份使用MySQL容器已有真实凭证完成，凭证未写入执行文档。

启动前代码审计另外发现并修复：

- API重启前遗留的`queued`、`running`抓取任务和`requesting`页面日志原本会永久阻止自动恢复；现由同一事务统一标记失败，并写入运行事件。
- Browser Agent关闭HTTP前原本没有先结束扩展长连接，存在关闭等待；现先取消运行环境和关闭扩展连接，再关闭HTTP。
- 扩展心跳原本没有连接身份，旧Bootstrap页面可能掩盖当前长连接断开；现按连接ID验证心跳。
- 扩展提交页面结果遇到Agent瞬时连接失败时原本会直接丢失；现使用请求原有截止时间进行结果重投，不新增隐藏时限。
- 自动状态查询改为每种状态创建独立GORM查询，避免查询条件相互污染。

上述修复已重新执行`go test -race ./...`、`go vet ./...`和三个Go入口构建，全部通过。

### 19.5 迁移入口与页面初始化规则补齐

```text
阶段：迁移入口与页面初始化规则补齐
执行时间：2026-07-22 17:43-17:54 Asia/Shanghai
执行人：Codex
目标主机：本机
实际文件或容器：remote/deploy/docker/api.Dockerfile、remote/make/docker.mk、remote/web/new
执行命令：Compose配置解析；Make dry-run；前端全局初始化扫描；npm run check；npm test；npm run build
结果：API镜像增加同源构建的独立迁移二进制和docker-migrate入口；全部Remote页面首次进入均不再请求API；前端验证全部通过
验证证据：页面目录不存在onMounted和immediate首屏加载；Vitest 8个测试文件、19个用例通过；类型检查、未使用检查、ESLint、Prettier和Vite生产构建通过
遗留问题：更新后的API和Web镜像需要在目标服务器重建；真实运行配置仍待落地
回滚点：代码层可回退本阶段文件；服务器旧服务尚未切换
```

本阶段按项目内`FetchLogsPage`和`TargetProbesPage`的既有交互统一了以下页面：

- `OverviewPage`：点击刷新后读取健康状态、最新排行统计和最新任务。
- `FetchRunsPage`：点击刷新或提交每页数量后读取任务列表。
- `FetchRunDetailPage`：点击刷新后读取指定任务；路由ID变化时恢复未查询状态。
- `RanksPage`：点击筛选或刷新后读取排名；清空只恢复页面状态。
- `RankDayStatsPage`：点击筛选或刷新后读取每日统计；清空只恢复页面状态。

页面分页只会在本页已经由用户执行过查询后继续读取同类数据。页面操作仍只访问Remote API，不会触发目标网站探测。

目标服务器已在本阶段完成更新后镜像重建：

- `docker_remote-browser:latest`：`242d0170fc7b`，包含连接身份心跳、结果重投和安全关闭修复。
- `docker_remote-api:latest`：`fceae2737e15`，包含独立`/app/remote-migrate`和遗留任务恢复修复。
- `docker_remote-web:latest`：`d37741c3705d`，包含全部页面首次进入零请求修复。

三次构建均使用临时环境变量通过Compose V1的构建期配置解析，这些临时值没有写入文件、没有进入镜像，也不作为生产健康检查参数。服务器代理、旧systemd服务和现有端口仍未修改。

### 19.6 生产运行参数确认与noVNC接入

```text
阶段：生产运行参数确认与noVNC接入
执行时间：2026-07-22 18:24-18:30 Asia/Shanghai
执行人：Codex
目标主机：本机
实际文件或容器：remote/api、remote/web/new、remote/deploy/docker
执行命令：gofmt；go test -race ./...；go vet ./...；npm run check；npm test；npm run build；docker-compose config；Make dry-run
结果：删除Docker健康检查；目标站探测确定为180m；Browser运行参数、日志轮转和noVNC代码已落地；本地验证通过
验证证据：Go全量竞态测试和vet通过；Vitest 8个测试文件、19个用例通过；前端生产构建通过；Compose V1配置可解析
遗留问题：服务器需要重新构建包含noVNC组件的Browser镜像，生成真实noVNC凭证，写入真实生产配置并完成Profile迁移
回滚点：服务器旧服务仍在运行，pre-docker-20260722-173527备份保持可用
```

本阶段确认并采用：

- Compose不配置Docker健康检查；三个容器使用`restart: unless-stopped`。
- 目标站探测周期`180m`。
- Browser扩展心跳`10s`，断连判定`30s`，重连等待`2s`。
- Chromium启动等待`90s`，页面请求超时统一使用`Source.Timeout: 60`，页面轮询`250ms`。
- 容器`json-file`日志单文件`20m`，保留`5`个。
- noVNC启用，容器内监听`0.0.0.0:6080`，Docker只映射服务器`127.0.0.1:6080`。
- Browser Agent监督Xvfb、x11vnc、noVNC和扩展长连接；关键进程故障时主动退出并由Docker重启。

原`BrowserAgent.PageTimeout`只有配置校验且没有参与请求截止时间，已删除；页面截止时间只由实际生效的`Source.Timeout`控制。

### 19.7 noVNC生产镜像构建与切换前核查

```text
阶段：noVNC生产镜像构建与切换前核查
执行时间：2026-07-22 18:31-18:36 Asia/Shanghai
执行人：Codex
目标主机：43.134.226.247
实际文件或容器：docker_remote-browser:latest、docker_remote-api:latest、docker_remote-web:latest
执行命令：docker-compose build；docker images；临时容器内可执行文件和noVNC静态资源核查；systemd与监听端口核查
结果：三类最终镜像构建成功；Browser镜像包含Xvfb、Chromium、x11vnc、websockify和noVNC静态页面；旧服务及端口保持原状
验证证据：Browser镜像e326580c1139（1.12GB）；API镜像b2bb59bf7627（137MB）；Web镜像c43bc8a70da7（48.4MB）；旧API和Xvfb服务仍为active；6080未监听
遗留问题：真实生产配置和noVNC凭证待落地；Profile、数据库、API与Web尚未切换
回滚点：pre-docker-20260722-173527备份；旧systemd服务和Nginx仍在运行
```

本阶段未修改服务器代理、数据库、现有运行配置和旧服务状态。Browser镜像体积增加来自Chromium、X11、x11vnc、websockify及noVNC完整运行依赖，目标服务器当前磁盘空间仍需在容器启动前复核。

切换前隔离预检发现并处理中：

- 使用空临时Profile启动时，Xvfb、Chromium、x11vnc和noVNC均正常运行，但扩展未建立长连接，`/readyz`持续返回503。
- 检查实际Chromium命令行确认，Debian的`/usr/bin/chromium`启动脚本先注入空的`--load-extension=`，随后才出现项目扩展目录；Chromium 150没有加载项目扩展。
- 生产配置改为直接调用`/usr/lib/chromium/chromium`，绕过Debian启动脚本的空扩展参数。
- 继续检查发现新Profile首次启动时，直接导航扩展Bootstrap页和扩展注册之间存在竞态；改为由扩展后台在安装、浏览器启动和后台首次运行时主动确保Bootstrap页存在，Chromium初始页使用`about:blank`。
- 待使用相同隔离预检确认扩展回连后再进入Profile和服务切换。
- 本次预检使用临时空Profile，没有映射宿主端口、没有访问目标网站，预检容器已经删除。

修复后隔离预检已通过：

- 最终Browser镜像：`eeeb4e89cde0`。
- 使用临时空Profile启动约2秒后进入`ready`，扩展长连接已建立并产生心跳。
- Xvfb、Chromium、x11vnc和websockify进程均在运行，`/healthz`返回200，`/readyz`返回200，noVNC静态页面返回200。
- 预检没有映射任何宿主业务端口、没有访问目标网站，预检容器已在采集完成后删除。

### 19.8 旧Browser停机与Profile迁移

```text
阶段：旧Browser停机与Profile迁移
执行时间：2026-07-22 18:53-18:56 Asia/Shanghai
执行人：Codex
目标主机：43.134.226.247
实际文件或容器：旧Remote API/Xvfb systemd服务、browser-profile-chromium、pre-docker-20260722-173527、rudy_gc_remote_browser_profile
执行命令：systemctl stop；停机进程核查；cp -a；文件数量核对；Docker命名卷复制与属主核对
结果：旧API和Xvfb已停止；旧Chromium已退出；Profile停机一致性备份和Docker卷复制完成
验证证据：源Profile、备份和Docker卷均为968个文件、159MB；Docker卷根目录为1001:1001、权限700；5031/5033已释放，旧Nginx继续监听5032
遗留问题：新卷中需要清理旧Chromium遗留的三个Singleton运行时锁；真实Profile Browser容器尚未启动
回滚点：重新启动旧API/Xvfb systemd服务；Profile原目录和备份均保持不变
```

第一次停机脚本在读取`inactive`状态时受到`pipefail`影响，以systemd退出码3提前结束；当时只完成两个旧服务停止，未开始Profile备份。修正状态采集后从进程核查和备份步骤继续，全部复制结果已经逐项核对。

真实迁移Profile首次启动发现并处理中：

- Browser容器正常启动Xvfb、Chromium、x11vnc和noVNC，但扩展未连接；日志出现旧协议`POST /ready`并返回404。
- Profile内保留1.0.2版扩展目录和Service Worker缓存，新旧扩展使用相同固定ID，Chromium继续运行旧扩展缓存。
- 新扩展移除固定`key`，改由稳定容器路径`/data/browser-extension`生成独立ID；扩展后台启动时关闭旧固定ID的Bootstrap标签页，避免旧协议继续产生请求。
- Cookie、目标站存储和其他Profile数据不做修改；原始Profile和停机备份保持不变。

真实Profile修复后验收通过：

- 最终Browser镜像：`070cc65ed099`。
- 容器启动约2秒进入`ready`，扩展连接和心跳正常，容器重启计数为0。
- 新容器日志中的旧协议`POST /ready`数量为0。
- noVNC页面返回200，宿主机只在`127.0.0.1:6080`监听。
- API尚未启动，本阶段没有触发排行抓取或目标站探测。

API启动和clearance核查：

- 该阶段当时API、数据库和Browser Agent均进入`ready`，目标站探测记录保持0，下一次探测时间为启动后180分钟；最终方案已改为API启动后立即探测一次。
- 失败恢复协调器发现08:30失败任务已超过1小时，立即创建恢复任务`#11`；第1页请求记录`#59`以HTTP 403和Cloudflare验证失败结束，任务状态正确落为`failed`。
- 迁移Profile中存在一条`cf_clearance`，有效期到2027-07-21；旧配置没有独立User-Agent或Cookie字段。
- 服务器Debian Chromium持续停留在Cloudflare自动验证页；排查时曾错误使用本机Chrome验证目标页，该结果全部作废，未写入服务器配置、Profile或数据库。
- 项目规则已补充Remote Docker独立运行约束；后续Cloudflare验证、Cookie更新、抓取和验收只允许在目标服务器内完成。
- 扩展原先发现Cloudflare页面后约563ms即返回失败，没有给自动验证留出时间；现改为在完整60秒请求期限内持续等待，期限结束后才记录Cloudflare失败。
- 首次官方Google Chrome隔离预检因临时`/data`默认归root而无法创建Profile；修正预检tmpfs的UID、GID和权限后继续，生产命名卷不受该预检参数影响。
- Google Chrome 150预检中Xvfb、Chrome、x11vnc和noVNC均正常运行，但扩展在90秒内没有建立长连接。Chromium官方源码确认Chrome品牌构建从137起禁止`--load-extension`，150中相关Feature绕过已经移除，因此该发行版无法满足现有扩展通道。
- Browser镜像改用服务器下载的官方Chrome for Testing固定版本`151.0.7922.34`，下载包SHA256固定为`ae8736ac28bc69278551500f219fc749575648263c43ec5990749eff43b9fcf8`；该发行版面向自动化且保留扩展加载能力。
- 删除失效的`DisableLoadExtensionCommandLineSwitch`参数和`--no-sandbox`，Browser Agent和Chrome继续以UID 1001非root用户运行。
- 服务器Docker 20.10默认seccomp会阻止Chrome沙箱创建namespace，隔离测试报`No usable sandbox`；仅对独立Browser容器设置`seccomp=unconfined`，不增加Linux capabilities，Chrome进程和渲染进程继续启用自身沙箱。
- 仓库运行配置已移除本机Chrome、X11和Profile路径，统一使用Docker容器网络、容器内浏览器路径和命名卷路径，不引入本机UA、Cookie、Profile或浏览器依赖。

官方Chrome for Testing空Profile隔离预检已经通过：

- Browser镜像：`e3e5454f189c`，Chrome for Testing版本`151.0.7922.34`。
- 临时Profile使用UID 1001、GID 1001且权限`0700`的内存文件系统，没有挂载生产Profile。
- Browser Agent约2秒进入`ready`，扩展长连接建立，Xvfb、Chrome、x11vnc和websockify均在运行。
- `/healthz`、`/readyz`和noVNC静态页面均返回200；Chrome启动参数中没有`--no-sandbox`。
- 预检没有映射宿主业务端口、没有访问目标网站，临时容器已经删除。
- 一次预检脚本误用新镜像中不存在的`wget`，Browser实际已就绪但脚本持续误报；改用镜像内已安装的`curl`后采集到上述完整证据。

真实Profile首次切换发现并修复：

- 第一次启动被迁移卷中的`SingletonLock`、`SingletonSocket`和`SingletonCookie`阻止，Chrome明确记录Profile仍由旧容器占用并以退出码21结束。
- 启动失败后Docker重启复用容器可写层，Xvfb遗留的`/tmp/.X99-lock`和Socket又使后续启动连续失败。
- 停止Browser容器并确认无Browser进程后，只删除上述三个Profile运行锁；Cookie数据库、站点存储、扩展状态和其他Profile文件均未修改。
- 强制创建新Browser容器后，真实Profile约1秒恢复`ready`，API重新连通，容器重启计数为0。
- Browser Agent启动前现只清理独占Profile顶层的三个`Singleton*`运行文件，以及当前display对应的Xvfb锁和Socket；路径和display均经过校验。
- Browser运行进程改为独立进程组并广播退出状态；Browser Agent关闭或启动失败时，依次终止Chrome、websockify、x11vnc和Xvfb，避免运行锁重复遗留。
- 新增运行锁范围测试，确认只删除指定运行文件并保留Profile业务文件及其他Xvfb display文件；Go全量竞态测试和`go vet`通过。
- 第一次有序停止虽然2秒内以退出码0完成，但向整个Chrome进程组同时发送SIGTERM仍留下三个Singleton文件；进程回收进一步调整为先只向主进程发送SIGTERM并等待其完成清理，再终止仍存活的子进程组。
- 隔离空Profile验证显示Chrome主进程收到SIGINT后会完整退出并删除全部Singleton文件；最终回收信号改为SIGINT，等待主进程完成清理后再终止残余子进程组。该测试未映射宿主业务端口、未访问目标网站，临时容器已删除。
- 最终Browser镜像`0e89d866dc8c`已使用真实Profile完成两次受控启动和一次受控停止：两次均约2秒进入`ready`，停止耗时1秒、退出码0，停止后三个Singleton文件均不存在，重新启动后容器重启计数保持0。
- 最终真实Profile Browser、扩展长连接、Xvfb、x11vnc、noVNC和API当前均已恢复，API健康接口返回200。

真实抓取任务`#13`发现并处理中：

- 任务和第1页请求日志均已入库；任务在约2秒后失败，错误为等待页面结果时扩展长连接断开，Browser和API容器没有退出或重启。
- Browser日志显示Profile中多个Bootstrap页面约每3秒轮流建立连接；服务端原实现接受新连接时会关闭当前连接，因此抓取连接被另一个Bootstrap页面抢占。
- Browser Agent改为只接受一个活动扩展连接，重复连接返回409且不会替换当前连接。
- 扩展后台串行执行Bootstrap页面创建，启动时关闭重复Bootstrap页面并只保留一个；扩展版本提升到`2.0.1`，确保真实Profile重新加载。
- 增加重复连接保护测试，验证重复连接不会关闭当前连接且当前连接心跳继续有效。
- 首轮连接保护部署后活动连接和心跳已稳定，但仍有一个重复Bootstrap页面每约3秒收到409；Profile只读检查确认遗留固定ID`alpcadeipbfgpmihmhccdgngjagnpoop`和当前路径派生ID`npcdhgdmlceabmncnopennplhalfoell`同时注册，且均指向`/data/browser-extension`。
- 扩展版本提升到`2.0.2`；遗留固定ID启动时调用Chrome官方`management.uninstallSelf`自行注销，当前扩展不直接改写Profile JSON。
- Bootstrap连接增加Web Locks单例锁，同一扩展存在多个页面时只有一个页面能够访问Browser Agent连接接口。
- 继续观察发现唯一活动HTTP命令流仍会由Chrome页面在不同时间主动结束，流结束后页面能够重新连接；旧实现把流生命周期等同于扩展会话生命周期，使运行中的页面请求被错误取消。
- 扩展版本提升到`2.0.3`；Bootstrap保留服务端分配的`connection_id`，HTTP流结束后携带同一ID恢复命令流。
- Browser Agent将扩展会话和HTTP命令流拆开：命令流断开只标记暂离，不关闭会话和进行中的抓取；同一会话可恢复命令流，重复页面仍无法抢占。
- 扩展会话最终失效只由既有30秒无有效心跳判定，不新增超时参数。
- 真实抓取任务`#15`验证扩展会话在完整60秒页面期限内保持稳定，Browser容器重启计数为0；API请求仍在页面业务期限到达的同一时刻主动取消，导致Browser Agent的页面结果无法返回。
- API到Browser Agent的HTTP调用改为只继承上层取消信号；既有`Source.Timeout`继续作为下发给Browser Agent的页面业务截止时间，由Browser Agent负责在该期限内结束页面任务，消除业务截止与内部传输截止同刻竞争，不新增超时参数。
- 后续持续观察捕获到Browser容器重启：Bootstrap后台页的定时器心跳被Chrome节流，最后一次心跳30秒后会话被判定失效；扩展继续使用已失效的`connection_id`重连并持续收到409，恢复期限结束后Browser Agent主动退出，Docker按既有重启策略拉起容器。
- 扩展版本提升到`2.0.4`；心跳改由Browser Agent既有10秒命令流keepalive驱动，只有扩展实际读取keepalive后才回执心跳，避免依赖后台页定时器；服务端返回409时清除失效连接ID并建立新会话，使断连恢复形成闭环，不新增心跳或重试参数。
- 部署后的真实抓取任务`#16`在完整60秒内每10秒持续收到扩展心跳，Browser和API重启计数均为0；任务和第1页请求日志已入库，API不再出现传输context deadline错误。
- `#16`仍暴露Browser Agent计时器与内容脚本在同一业务截止点竞争：服务端先取消标签页，数据库只能得到通用deadline。扩展版本提升到`2.0.5`，增加`finish`最终检查命令；60秒只控制页面业务过程，到点后强制内容脚本立即分类并可靠投递真实页面结果，投递阶段复用既有30秒断连恢复期限，不增加页面时间和新参数。
- 真实抓取任务`#17`验证最终结果协议生效：第1页请求日志`#65`记录HTTP 403、Cloudflare验证失败和60053ms耗时，Browser保持`ready`且重启计数为0；进程、扩展连接、业务截止、错误分类和数据库日志链路均已稳定。
- 扩展版本提升到`2.0.6`；目标站改为单一持久活动标签，每次抓取复用并重新导航，清理同源重复标签；Cloudflare失败后保留验证页面继续运行，使服务器相同公网出口、Chrome User-Agent和持久Profile能够继续获取并保存clearance，后续自动重试、定时探测和noVNC维护复用同一会话。
- 任务`#18`至`#20`均保持扩展心跳和Browser零重启；Profile中的`cf_clearance`由服务器Chrome持续刷新，但noVNC实际检查发现任务结束后目标标签仍被关闭，迁移Profile内缓存的旧扩展后台脚本仍在响应旧`complete`消息并执行关闭标签逻辑。
- 扩展版本提升到`2.0.7`；内容脚本使用新的`browser_agent_result`消息，当前Bootstrap页面负责可靠写回Browser Agent，新旧后台脚本均不再参与业务结果投递。旧缓存无法识别新消息，因此不能再关闭目标标签，持久验证页与Cookie更新过程得到保留。
- 任务`#21`验证新结果通道生效：业务失败后目标标签继续存在，Browser保持`ready`且重启计数为0；noVNC确认目标页为无交互控件的Cloudflare自动验证转圈页。
- noVNC同时确认Chrome for Testing页面固定显示仅用于自动化测试的产品提示，目标站持续循环验证并反复刷新`cf_clearance`。Browser镜像增加项目此前已通过扩展连接预检的Debian标准Chromium，真实执行路径切换到`/usr/lib/chromium/chromium`进行服务器内A/B验证；原Chrome for Testing仍暂存于同一镜像作为本阶段回滚入口，验证完成后再确定最终保留项。
- 标准Chromium `150.0.7871.124`已使用同一Profile启动并通过扩展、Xvfb、noVNC和API就绪检查；切换前Profile已复制到独立回滚卷，源卷与回滚卷均为1017个文件、143988881字节。
- 标准Chromium的Cloudflare页面出现明确`Verify you are human`按钮；通过noVNC点击后返回`Incorrect device time`。宿主机与容器上海时间、UTC和epoch一致，RTC只相差约1秒且宿主机报告时钟已同步。
- 扩展版本提升到`2.0.8`；在同一条已记录的主文档请求上读取Cloudflare响应`Date`头和Chrome实际接收时间，将二者差值写入请求错误消息。该诊断不新增目标请求、不记录HTML或Cookie，用于确认外部响应时间与服务器时钟是否一致。
- 真实抓取任务`#26`记录目标响应时间与Chrome接收时间相差`88ms`；任务`#27`相差`618ms`。两次均保持Browser容器重启计数为0，宿主机、容器和目标响应的绝对时间一致，排除服务器实际时钟偏差。
- 任务`#27`在页面60秒有效期内通过noVNC两次触发服务器Chromium中的`Verify you are human`，Cloudflare仍重新发起验证并最终返回HTTP 403。当前抓取阻塞点收敛为Cloudflare对服务器Chromium会话的挑战判定；禁止通过修改服务器代理、使用本机浏览器或迁移本机Cookie规避。
- 手工连续创建任务`#24`和`#25`暴露同一业务日期可同时存在多个活动任务的数据库竞态；串行执行锁只控制执行阶段，第二个任务在队列中仍会提前消耗页面截止时间。修复目标为数据库直接保证同一业务日期和分类最多存在一个`queued`或`running`任务。
- 数据库增加生成字段`active_key`及唯一索引`uidx_dailybest_active_run`；两条并发手工请求实测只有任务`#28`创建并运行，另一条返回HTTP 409，数据库中只有一条活动记录。
- 任务`#28`在60秒业务截止后等待30秒仍没有收到最终页面结果；Browser与扩展心跳持续正常、容器零重启。noVNC检查确认超时取消后目标标签被扩展关闭，说明请求映射仍存在；代码检查定位为Cloudflare导航可能清除URL hash，内容脚本在`finish`阶段缺少请求标识后直接结束。
- 扩展版本提升到`2.0.9`；Bootstrap保留活动请求正文，并在`finish`命令中直接传递`request_id`和`deadline`，内容脚本仅把URL hash作为首选来源、把命令正文作为可靠回退。标签缺失、内容脚本不可用和标签创建失败统一通过既有结果重投通道提交。
- 服务器内隔离Google Chrome 150无扩展、无自动化参数A/B检查受到宿主虚拟CPU非法指令中断，目标标签显示`SIGILL`；禁用本地模型功能和V8 JIT后结果不变。生产Browser、Profile、代理及Cookie均未改动，该A/B结果不能用于判断Cloudflare是否放行。
- 扩展`2.0.9`生产任务`#30`和恢复任务`#31`均在约60033ms准确记录HTTP 403、Cloudflare分类和响应时钟，Browser零重启；`#28`出现的90秒无最终结果问题未再复现。
- 出口核查确认宿主机和Browser容器当前均使用`43.134.226.247`公网出口，容器内外均没有代理环境变量；本阶段没有修改服务器代理。
- 时区一致性核查发现Browser进程环境为`TZ=Asia/Shanghai`，但镜像内`/etc/localtime`和`/etc/timezone`仍指向UTC。绝对时间虽正确，系统时区与浏览器进程时区的双重信号可能触发Cloudflare设备时间指纹；Browser与API镜像改为在构建时固定系统时区为`Asia/Shanghai`。
- 新Browser镜像`0a740167e00f`已部署；容器内`/etc/localtime`、`/etc/timezone`、`date`和进程环境均统一为`Asia/Shanghai`，Browser保持`ready`且重启计数为0。
- 时区修正后的真实抓取任务`#33`在60026ms完成失败分类并落库：第1页HTTP 403，目标响应与Chrome接收时间相差`904ms`，Browser容器重启计数为0。系统时区不一致已经消除，但Cloudflare仍拒绝该服务器出口和浏览器会话。
- API时区镜像`6538b6648a69`已部署；API容器内系统时区和进程时区均为`Asia/Shanghai`，`/readyz`返回200，数据库及Browser Agent均为`ready`，API和Browser容器重启计数均为0。
- 服务器内使用官方Chrome for Testing `136.0.7103.113`进行临时空Profile隔离A/B；扩展正常连接，完整60秒请求最终仍为HTTP 403和Cloudflare分类，目标响应时钟差`513ms`。临时容器、Profile、安装包和noVNC隧道均已删除。
- Docker网络内外的`www.c97k.com`、Cloudflare STUN和Google STUN均可解析；Cloudflare STUN通过生产Browser相同Docker网络收到32字节有效响应，排除Docker DNS和UDP出口阻断。
- 使用标准Google Chrome `150.0.7871.181`、临时空Profile、无扩展且无自动化参数进行隔离A/B；通过官方Linux策略关闭Safe Browsing后未再发生SIGILL，Chrome持续运行且零重启，但目标页仍持续循环Cloudflare验证。该结果排除扩展、Chrome for Testing产品标识和SIGILL是持续403的根因；临时容器、策略文件和noVNC隧道均已删除。
- GORM运行配置忽略预期的RecordNotFound并关闭ANSI颜色；API镜像`54efc56a7710`部署后`/readyz`返回200，启动日志只保留服务监听记录和访问日志，不再输出预期空查询的错误噪声。
- Remote Web重新执行类型检查、未使用项检查、ESLint、格式检查、19个单元测试和生产构建，全部通过。候选容器首次进入运行概览仅加载静态资源，未发出API或目标站请求；用户点击刷新后才读取健康状态、每日统计和抓取任务。
- 候选验证发现Web Nginx的`/readyz`落入SPA fallback并产生假200；新增精确反向代理后，`/healthz`返回Web状态，`/readyz`真实返回Remote API的`{"status":"ready"}`。
- Docker Web镜像`e40b852a43b1`已接管公网5032；原宿主Nginx的Remote专用启用链接已注销，源配置和旧文件保留为当前回滚点，宿主Nginx的80端口默认站点不受影响。API、Browser和Web三个正式容器均为运行状态且重启计数为0，Browser Agent端口仍只在Docker网络内可达，noVNC仍只绑定服务器回环地址。
- A/B验证完成后删除Browser镜像中未使用的Chrome for Testing下载、校验、解压和安装步骤；最终Browser镜像`989b16c7ac36`只保留生产实际使用的Debian Chromium `150.0.7871.124`及noVNC运行依赖。同一Profile Volume切换后约3秒恢复`ready`，CfT目录不存在，Browser与API重启计数均为0。
- 失败补偿链完成一次生产实测：自动任务`#31`于22:03:10结束后，调度器在23:03:10创建`recovery`任务`#34`，间隔严格为1小时；任务在23:04:10结束，完整执行既有60秒页面期限。
- 任务`#34`只产生一条第1页请求日志`#80`：URL为DailyBest第1页、HTTP 403、耗时60029ms，错误准确归类为Cloudflare验证未通过，并记录目标响应时间、Chrome接收时间和`160ms`时钟差。任务和请求日志均已落库，没有重复请求日志。
- `#34`执行前后Browser扩展心跳持续为10秒，任务结束后Browser Agent从`fetching`恢复为`ready`；API和Browser容器重启计数均为0。下次失败补偿时间由`#34.finished_at + 1h`继续计算，补偿链不会因手工任务改变。
- 2026-07-22 23:05最终运行核验：Compose配置解析成功；Web `/healthz`返回`ok`，API `/readyz`返回`{"status":"ready"}`；数据库和Browser状态均为`ready`，Xvfb、x11vnc、noVNC、Chromium及扩展长连接全部在线；API、Browser、Web三个正式容器重启计数均为0。
- 正式端口核验：API监听`0.0.0.0:5031`，Web监听`0.0.0.0:5032`，noVNC只监听`127.0.0.1:6080`，Browser Agent的5033只存在于Docker网络。旧API和Xvfb systemd服务保持`inactive`，旧Remote Nginx启用链接保持注销。
- 当前唯一阻塞项是Cloudflare持续拒绝`43.134.226.247`出口上的服务器浏览器会话。服务器内不同Chromium/Chrome版本、空Profile、迁移Profile、有无扩展、系统时区、DNS和STUN出口均已分别实测，业务运行链路稳定且每次失败均可观测、可落库、可自动补偿，但当前尚无一次被目标站放行的Docker抓取成功结果。
- 按切换约束保留旧版本文件、Profile回滚卷和配置备份，暂不执行不可逆清理；待服务器浏览器会话获得目标站放行并完成一次全量成功、一次`no_change`及180分钟目标探测实测后，再执行最终旧版本删除。
- `#34`期间Chromium的WebRTC模块记录STUN域名解析错误；同期容器系统解析和直接UDP STUN测试均成功。为排除Docker内置DNS影响，使用最终Browser镜像、临时空Profile和显式`1.1.1.1`、`8.8.8.8`进行隔离A/B。
- 显式公共DNS的隔离请求仍在59318ms后返回HTTP 403和Cloudflare分类，目标响应时钟差`31ms`；Chromium WebRTC仍记录相同STUN解析错误，容器系统解析保持成功。由此确认更换Docker DNS既不能消除浏览器级日志，也不能使挑战通过，正式Compose不加入该无效变更。
- DNS A/B未挂载生产Profile、未连接生产API、未映射正式端口、未修改代理；临时容器和宿主回环测试端口`15033`均已删除。用于noVNC验收的本地SSH隧道和MCP Chrome noVNC标签页也已关闭，正式Remote运行不依赖本机进程。
- 持续稳定性排查确认Docker bridge会破坏Chromium WebRTC的STUN解析：同一最终Browser镜像在bridge网络中20秒记录54条`ERR_NAME_NOT_RESOLVED`，切换服务器host网络后15秒记录0条。显式公共DNS、关闭AsyncDns以及HostResolver映射均不能修复bridge中的P2P解析，故问题收敛到Docker bridge网络路径。
- 使用host网络完成60秒Browser Agent空Profile生产级A/B，STUN错误保持0，Browser Agent全程`ready`且重启0次；目标站仍返回HTTP 403，证明host网络修复了确定的浏览器网络缺陷，但Cloudflare仍拒绝服务器出口会话。
- 进一步完成官方Google Chrome 150、SwiftShader WebGL、无扩展、无CDP和Firefox ESR 140的host网络A/B；两种浏览器均持续停留在Cloudflare验证页。使用CDP把User-Agent、Client Hints、平台、语言、时区、分辨率和WebGL统一为此前凭证对应的Mac Chrome环境后仍为Cloudflare页面，浏览器品牌、内核、WebGL、自动化参数和常用前端指纹已分别排除。
- 服务器当前没有代理进程、代理环境、透明转发或IPv6出口；旧`cf_clearance`与原配套User-Agent直连测试已返回403。当前主要业务阻塞收敛为`43.134.226.247`公网出口受到目标Cloudflare策略拒绝。
- 正式网络修复开始落地：Browser容器改用host网络，Browser Agent只绑定`172.30.0.1:5033`，noVNC只绑定`127.0.0.1:6080`；Remote内部bridge固定为`172.30.0.0/16`，API通过内部网关主机映射访问Browser Agent。该调整不修改服务器代理，不开放Browser Agent或noVNC公网端口。

## 2026-07-23 FlareSolverr执行链重写

### 原Browser Agent切换问题

- 第一次host网络切换脚本错误地用Compose标签仍完整的重命名容器作为回滚容器；`docker-compose up`识别并重建了这些容器，失败处理随后删除了新容器，导致API和Browser Agent短暂缺失。
- 切换配置同时遗漏`NoVNC.VNCListenPort`，Browser Agent第一次创建后启动失败。补齐配置后重新创建API和Browser Agent，`/readyz`恢复200；Web在API缺失期间重启10次，API恢复后重新连通。
- 后续回滚只允许使用固定镜像、独立配置备份和明确的Compose文件，禁止再把带Compose标签的重命名容器当作回滚副本。

### Cloudflare处理方案验证

- 正式host网络任务`#36`仍在60033ms后返回Cloudflare HTTP 403，请求日志`#82`准确落库；Browser Agent全程心跳正常且容器零重启，host网络消除了STUN解析错误但没有解决挑战。
- 同一服务器、同一公网出口和host网络启动FlareSolverr`3.5.0`隔离容器。第1页在12999ms内自动通过Cloudflare，返回HTTP 200、完整`#rightcolumn`以及由服务器浏览器产生的`cf_clearance`。
- 持久会话连续请求第1至25页：第1页约12.3秒，第2至23页约0.7至1.1秒；前23页均返回有效排行，共458个影片节点，第24和25页为空页，符合现有遇空页结束协议。
- 隔离容器测试完成后立即删除，宿主8191端口不再监听；测试过程没有修改代理，没有使用本机Chrome、Cookie或Profile。

### 正式重写内容

- Compose增加固定摘要`sha256:139dfee1c6f89249c8d665d1333a42e8ec74ec0a86bc6bb1c8461e10d3a66a47`的`remote-solver`，使用host网络并只监听Docker内部网关`172.30.0.1:8191`。
- API改用持久FlareSolverr会话`rudy_gc_dailybest`，会话不存在时自动创建，Solver重启导致会话失效时自动重建；所有业务分页继续逐页写入`d_dailybest_fetch_log`。
- 删除自研Browser Agent二进制、扩展协议、Xvfb、x11vnc、noVNC、Profile处理代码和Browser Dockerfile；目标探测、手工抓取、定时抓取和失败补偿统一使用同一Solver客户端。
- 运行状态页面删除Xvfb、noVNC和扩展字段，改为显示Solver状态、版本、User-Agent、最近状态检查和错误。
- 本地后端全量`go test ./...`、`go vet ./...`和API构建通过；Remote Web类型检查、19个测试和生产构建通过。

### 最终验收结果

- 已部署正式Solver、API和Web容器，三个新容器均为运行状态且重启计数为0；Solver只监听`172.30.0.1:8191`。
- 正式任务`#38`成功：23个有效页面、458条排名、checksum为`49094c10f88aab77359fadaf985074e8f2da622de1390b6049cd9490e9dd2a2c`；第24页为空页并用于结束判定。24条请求日志全部HTTP 200，第1页12878ms自动通过挑战，后续页面复用会话。
- 同日任务`#39`正确标记为`no_change`：页数、条目数和checksum与`#38`一致，24条请求日志全部成功，第1页复用会话耗时842ms。
- 受控停止Solver超过10秒后，API`/readyz`返回503，运行事件写入`browser_solver_unavailable`；Solver启动后API恢复200，三个新容器重启计数保持0。
- Solver重启丢失运行会话后的任务`#40`再次标记为`no_change`，第1页自动重新通过Cloudflare，24条请求日志全部成功，无需人工写入Cookie或User-Agent。
- API重启后的任务`#41`再次标记为`no_change`：23个有效页面、458条排名和checksum均保持一致，24条请求日志全部HTTP 200；第1页在1569ms内复用Solver现有会话，证明API进程重启不会丢失抓取能力。
- 恢复测试发现`/readyz`把所有依赖故障误标为`database_unavailable`，已改为通用`service_unavailable`，避免把Solver故障错误归类为数据库故障。
- 目标站调度改为API启动后立即探测、完成后每`180m`继续探测；生产探测`#1`在838ms内返回HTTP 200，数据库记录和`target_available`运行事件均已写入，运行状态为`available`。
- 服务器旧Browser容器、旧Browser镜像、旧Profile卷、noVNC凭证、旧systemd单元、旧非Docker发布目录和旧回滚备份均已删除；当前只保留Docker发布目录、正式配置、数据库和API/Web/Solver三个正式容器。
- 删除所有零容器引用的旧Remote构建中间镜像后，服务器无标签镜像数量归零，根分区使用率从`87%`降至`56%`，可用空间从`7.4GiB`增加到`25GiB`；其他项目有标签镜像和容器保持不变。
- 自动状态查询曾把`manual`成功任务错误计入当天自动成功状态，导致旧`recovery #37`被手工任务`#38`至`#41`提前终止；查询已限定为只有`scheduled/recovery`成功才能结束自动恢复链，活动任务仍统一阻止并发创建。
- 修正部署后，`recovery #42`在`#37.finished_at + 1h`对应的`2026-07-23 02:06:10`准确创建并启动，02:06:23完成为`no_change`；23个有效页面、458条排名和完整checksum与成功基线一致，24条请求日志全部成功且全部HTTP 200，单页耗时469至889ms。
- `#42`完成后自动成功状态准确指向该任务，活动任务数和`id > 42`任务数均为0，证明成功后恢复链已停止，没有重复补偿。
- 生产连续监控期间API`/healthz`、API`/readyz`、Web`/healthz`和Web`/readyz`始终返回200，API、Web和Solver容器始终运行且重启计数为0；最终目标站状态为`available`。
- 08:30定时任务与失败后1小时补偿共用已经生产验证的`enqueueAutomatic -> execute -> FlareSolverr -> publish/no_change`执行链；本次已完成补偿调度的真实时序验收。

## 2026-07-24 `c_rank_day_stat`单表查询与生产部署

### 实现结果

- Local和Remote的`c_rank_day_stat`均增加`fetched_on`、`synced_on`，分别记录源排行抓取完成时间和排行同步入库时间。
- Local `/rankdaystats`列表、汇总和分页只查询`c_rank_day_stat`；Remote `/api/v1/dailybest/rank-day-stats`也只查询该表。
- 远端抓取、本地远端同步和本地直接抓取均在统计行首次生成时写入两个时间字段。
- 缺失统计补建只插入缺少的日期；已有统计行、`calculated_on`、`created_on`和`updated_on`不再覆盖。
- 历史迁移只对两个新增字段值为0的记录执行一次回填，其他统计字段和时间字段不参与更新。

### 验证与部署记录

```text
阶段：c_rank_day_stat单表查询与生产部署
执行时间：2026-07-24 01:01-01:16 Asia/Shanghai
执行人：Codex
目标主机：本机、43.134.226.247
实际文件或容器：local/api、local/web/new、remote/api、remote/web/new、shared/go、rudy_gc_remote_api、rudy_gc_remote_web、rudy_gc_remote_solver
执行命令：Go全量测试、vet和构建；前端测试、check和构建；本地数据库迁移；生产相关表备份；Docker镜像构建；生产迁移；Compose切换；API、数据库和运行状态核对
结果：Local和Remote均完成单表查询改造；本地及生产数据库新增字段完成回填；生产版本切换到docker-20260724-010758
验证证据：本地c_rank_day_stat共1279行，fetched_on和synced_on均1279行非零；生产3行均完成回填且迁移前全部既有字段逐项一致；生产迁移重复执行后完整API结果保持一致；API镜像26820fddff8c、Web镜像22da5a0dfdc1；三个容器运行且重启次数均为0；API与Web就绪接口均为200；启动探测#11返回HTTP 200、耗时917ms
遗留问题：无
回滚点：/home/lighthouse/data/rudy_gc/remote/backups/pre-rank-day-stat-docker-20260724-010758；rudy_gc_remote-api:rollback-docker-20260724-000910；rudy_gc_remote-web:rollback-docker-20260724-000910
```

部署上传时第一次`rsync`把共享模块放到了新发布目录根部`go/`；构建开始前已经移动到要求的`shared/go/`并核对`go.mod`和Compose文件，错误目录未进入镜像、未参与迁移，也未影响运行中的旧版本。

## 2026-07-28 每日抓取邮件通知生产部署

### 部署结果

- `remote/api`增加通用`Notifier`与SMTP STARTTLS实现；每日`08:30`自动抓取首次失败时发送告警，同一业务日期和榜单分类的恢复任务连续失败不重复发送，恢复成功后发送一次恢复邮件，手动抓取不触发通知。
- 生产配置保留服务器既有数据库、抓取和浏览器配置，只合并完整`Mail`配置；配置权限为`600`，详情页公网地址为`http://43.134.226.247:5032`。
- 发布源码排除`remote/api/config/server.yaml`，SMTP凭证只保存在生产配置文件及权限为`600`的部署前配置备份中。
- 新API镜像完成构建和幂等数据库迁移，仅重建`rudy_gc_remote_api`；Web、Solver和数据库运行状态保持不变。

### 验证与回滚记录

```text
阶段：每日抓取邮件通知生产部署
执行时间：2026-07-28 13:02-13:20 Asia/Shanghai
执行人：Codex
目标主机：43.134.226.247
发布版本：docker-20260728-130427
API镜像：f1ce1c30f0c8d254a71592b0263819850bafa75541e6a4866323b13d0f028d61
执行命令：上传独立发布目录；备份生产配置和旧API镜像；合并Mail配置；Compose校验；API镜像构建；生产迁移；仅重建remote-api；内外网健康检查；服务器内Go SMTP精确发信验证
结果：current已切换到docker-20260728-130427；API容器运行且重启次数为0；API与Web的healthz、readyz均为200；抓取任务API和任务详情页公网访问均为200；SMTP服务器接受服务器内同一Go实现提交的测试邮件
验证说明：容器内curl的SMTP认证协商返回Login denied；随后使用生产代码的Go SMTP实现、同一生产配置和API Docker网络精确验证成功，该curl结果未影响应用邮件链路
回滚配置：/home/lighthouse/data/rudy_gc/remote/backups/pre-mail-notify-docker-20260728-130427/server.yaml
回滚镜像：docker_remote-api:rollback-docker-20260728-130427
新镜像标签：docker_remote-api:docker-20260728-130427
遗留项：等待下一次真实08:30任务失败或恢复事件触发业务通知；不主动制造生产抓取失败
```
