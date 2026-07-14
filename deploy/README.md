# Docker 部署

Docker 部署包含两个业务容器：

- `gc-api`：Go API 服务，Docker 内监听 `15013`。
- `gc-web`：运行 Vue/Vite 前端服务，Docker 内监听 `15012`，并代理 `/api/...` 到 `gc-api`。

基础服务使用已有 Docker 容器：

- MySQL：`rudy_infra_mysql_local`，宿主机端口 `127.0.0.1:3307`。
- Cache Redis：`rudy_infra_redis_local`，宿主机端口 `127.0.0.1:6389`。
- Biz Redis：`rudy_infra_redis_6399_local`，宿主机端口 `127.0.0.1:6399`。

## 数据库导入链路

服务镜像构建完成后，执行宿主机 MySQL 到 Docker MySQL 的导入：

```text
宿主机 MySQL 127.0.0.1:3306/rudy_gc
  -> mysqldump
  -> Docker MySQL rudy_infra_mysql_local/rudy_gc
```

导入脚本从 `rudy_infra_mysql_local` 容器里访问宿主机 MySQL：

```text
host.docker.internal:3306/rudy_gc
```

目标库写入同一个 Docker MySQL 容器内：

```text
127.0.0.1:3306/rudy_gc
```

目标数据库已有表时脚本停止，避免覆盖 Docker MySQL 现有数据。

## Docker 命令

Docker 拉取镜像和构建步骤默认使用本机代理：

```text
http://127.0.0.1:7897/
```

构建容器内部通过以下地址访问宿主代理：

```text
http://host.docker.internal:7897/
```

完整流程：

```bash
make -C /Users/gaojinwei/src/my/rudy_gc docker-rebuild
```

等价拆分流程：

```bash
make -C /Users/gaojinwei/src/my/rudy_gc docker-build
make -C /Users/gaojinwei/src/my/rudy_gc docker-import-db
make -C /Users/gaojinwei/src/my/rudy_gc docker-up
```

常用维护命令：

```bash
make -C /Users/gaojinwei/src/my/rudy_gc docker-ps
make -C /Users/gaojinwei/src/my/rudy_gc docker-logs
make -C /Users/gaojinwei/src/my/rudy_gc docker-down
```

## 运行配置

API Docker 配置文件：

```text
/Users/gaojinwei/src/my/rudy_gc/configs/docker/api.yaml
```

该文件保存 Docker 运行所需的完整数据库与 Redis 连接信息，由 Compose
只读挂载到 `/app/config/api.config.yaml`。

## 访问入口

```text
http://127.0.0.1:15012/cards
http://127.0.0.1:15012/api/health
```
