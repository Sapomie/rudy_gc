# Remote Docker 运行说明

本目录定义`remote-solver`、`remote-api`和`remote-web`三个容器。MySQL继续使用服务器已有的`rudy_mysql_remote`，API通过外部网络`rudy_remote_infra`和别名`mysql`连接。

`remote-solver`使用固定摘要的FlareSolverr镜像和服务器host网络，自动处理Cloudflare浏览器挑战。服务只监听Docker内部bridge网关`172.30.0.1:8191`，不监听服务器公网地址。API使用持久会话`rudy_gc_dailybest`连续抓取全部分页，Cookie和User-Agent由Solver会话管理，不写入YAML、数据库或请求日志。

服务器使用`docker-compose 1.28.6`，`compose.yaml`固定使用兼容的Compose`2.4`格式。三个容器使用`restart: unless-stopped`，Compose不配置Docker健康检查。API和Web保留`/healthz`、`/readyz`供人工检查和运行状态页面使用。

每个容器的`json-file`日志单文件上限为`20m`，保留`5`个文件。真实服务配置固定只读挂载：

```text
/home/lighthouse/data/rudy_gc/remote/config/server.yaml
```

数据库迁移使用API镜像内与服务同源构建的独立二进制：

```bash
make docker-migrate
```

常用检查：

```bash
make docker-config
make docker-up
make docker-status
make docker-logs
```

部署和迁移进度记录在`remote/doc/docker_rewrite_plan.md`。
