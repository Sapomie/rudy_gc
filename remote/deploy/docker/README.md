# Remote Docker 运行说明

本目录定义`remote-solver`、`remote-api`和`remote-web`三个容器。共享MySQL由`deploy/shared/mysql/compose.yaml`独立管理，API通过外部网络`rudy_shared_infra`和别名`mysql`连接。停止或删除Remote Compose不会影响MySQL、共享网络或数据库数据。

`remote-solver`使用固定摘要的FlareSolverr镜像和服务器host网络，自动处理Cloudflare浏览器挑战。服务只监听Docker内部bridge网关`172.30.0.1:8191`，不监听服务器公网地址。API使用持久会话`rudy_gc_dailybest`连续抓取全部分页，Cookie和User-Agent由Solver会话管理，不写入YAML、数据库或请求日志。

三个容器的Linux AMD64镜像在本机构建并打入离线包。服务器加载的镜像使用`localhost/rudy_gc/*:<发布版本>`名称，Compose设置`pull_policy: never`，启动命令固定附带`--pull never --no-build`。服务器不访问Docker Hub或GHCR。

三个容器使用`restart: unless-stopped`，Compose不配置Docker健康检查。API和Web保留`/healthz`、`/readyz`供人工检查和运行状态页面使用。

每个容器的`json-file`日志单文件上限为`20m`，保留`5`个文件。真实服务配置固定只读挂载：

```text
/var/snap/docker/common/rudy_gc/remote/config/server.yaml
```

数据库迁移使用API镜像内与服务同源构建的独立二进制：

```bash
make docker-migrate
```

本机构建：

```bash
make docker-build RELEASE_TAG=<发布版本>
```

服务器加载和启动：

```bash
scripts/install-remote-offline.sh <离线包目录> /var/snap/docker/common/rudy_gc/remote/config/server.yaml
```

安装脚本固定按 Solver、API、Web 的顺序启动；Solver 可访问且 API 就绪后才启动下游服务，避免首次目标站探测和 Nginx 上游解析发生竞态。

发布包内的`nginx.conf`只读挂载到 Web 容器。Nginx通过固定的Remote bridge网关`172.30.0.1:5031`访问宿主机已发布的API端口，容器启动和请求转发均不依赖Docker内置DNS。

生产`server.yaml`由`root:1001`持有并设置为`640`，其中`1001`是 API 镜像内`lighthouse`用户的组 ID；文件只读挂载到容器。

常用检查：

```bash
make docker-status
make docker-logs
```

部署和迁移进度记录在`remote/doc/docker_rewrite_plan.md`。
