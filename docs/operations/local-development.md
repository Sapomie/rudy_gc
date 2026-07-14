# 本地开发

在仓库根目录执行：

```bash
make paths
make run-api
make run-web
make verify
```

默认端口为 API `5013`、Web `5022`，所有开发服务监听 `0.0.0.0`。
本地 API 读取 `configs/local/api.yaml`。
