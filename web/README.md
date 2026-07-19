# Web 前端

当前前端代码位于 `new/`，默认使用 5022 端口，并监听 `0.0.0.0`。

开发服务器默认代理到 `http://127.0.0.1:5013`。

## 安装

```bash
npm --prefix new ci
```

## 启动

```bash
npm run dev
```

也可以在项目根目录执行：

```bash
make run-web
```

## 检查和构建

```bash
npm run check
npm test
npm run build
```

对应的项目根目录命令为：

```bash
make check-web
make test-web
make build-web
```

`make run-all` 同时启动 API 和当前前端。
