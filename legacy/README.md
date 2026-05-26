# Legacy 目录说明

`legacy/backend` 是旧版 `rudy_gc` 单体根目录。

当前职责：

- `legacy/backend`：旧 Go 模块、旧 handler/service、旧模板、旧静态资源
- `legacy/deploy`：旧部署文件
- `legacy/docs`、`legacy/doc`：旧文档与笔记
- `legacy/make`：旧 make 模块
- `legacy/oldmodel`：旧模型定义
- `legacy/z_reference`：前端参考模板
- `legacy/z_text`：文本与草稿

当前边界：

- 旧运行面保留在 `legacy`
- 新前后端分离开发默认落到 `v2`
- `v2/gc-api` 当前仍通过兼容层依赖 `legacy/backend` 的部分 service 和 handler
