# 共享 MySQL

共享 MySQL 作为服务器级基础设施独立部署，容器名为`rudy_mysql_shared`，持久数据位于`/var/snap/docker/common/rudy_infra/mysql/data`。所有业务 Compose 通过外部网络`rudy_shared_infra`连接，并使用网络别名`mysql`访问数据库。

服务器不构建或拉取镜像。MySQL 的 Linux AMD64 镜像在本机构建机拉取、重标记并通过`docker save`生成离线包；服务器校验后只执行`docker load`和`docker compose up --pull never --no-build`。

真实`MYSQL_ROOT_PASSWORD`只保存在服务器`/var/snap/docker/common/rudy_infra/mysql/config/mysql.env`，权限固定为`600`。每个业务使用独立数据库账号和数据库权限，业务配置不使用 MySQL root 账号。

共享网络由安装脚本独立创建，不归任何业务 Compose 所有。停止或删除 Remote Compose 不会停止 MySQL、删除共享网络或删除数据库数据。

本机构建离线包：

```bash
make shared-mysql-bundle RELEASE_TAG=<发布版本>
```

服务器安装：

```bash
scripts/install-shared-mysql.sh <离线包目录> \
  /var/snap/docker/common/rudy_infra/mysql/config/mysql.env \
  /var/snap/docker/common/rudy_infra/mysql/data
```
