.PHONY: help paths run-api stop-api test-api build-api install-web run-web stop-web test-web build-web preview-web run-all stop-all test-all build-all verify deploy-help docker-config docker-build docker-up docker-down docker-logs deploy-ps deploy-logs

help:
	@echo "V2 targets:"
	@echo "  paths         查看 v2 目录"
	@echo "  run-api       启动 gc-api"
	@echo "  stop-api      停止 gc-api 本地端口"
	@echo "  test-api      运行 gc-api 测试"
	@echo "  build-api     构建 gc-api"
	@echo "  install-web   安装 gc-web 依赖"
	@echo "  run-web       启动 gc-web 开发服务"
	@echo "  stop-web      停止 gc-web 本地端口"
	@echo "  test-web      运行 gc-web 测试"
	@echo "  build-web     构建 gc-web"
	@echo "  preview-web   预览 gc-web 构建产物"
	@echo "  run-all       同时启动 gc-api 与 gc-web"
	@echo "  stop-all      停止 gc-api 与 gc-web 本地端口"
	@echo "  test-all      运行 API 和 Web 测试"
	@echo "  build-all     构建 API 和 Web"
	@echo "  verify        运行 test-all 与 build-all"
	@echo "  deploy-help   查看部署目标"
	@echo "  docker-config 检查 compose 配置"
	@echo "  docker-build  构建 v2 docker 镜像"
	@echo "  docker-up     启动 v2 docker"
	@echo "  docker-down   停止 v2 docker"
	@echo "  docker-logs   查看 v2 docker 日志"
	@echo "  deploy-ps     查看 v2 docker 服务"

include make/common.mk
include make/gc-api.mk
include make/gc-web.mk
include make/run.mk
include make/deploy.mk
