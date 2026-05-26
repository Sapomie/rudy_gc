.PHONY: help legacy-help v2-help \
	v2-paths v2-run-api v2-test-api v2-build-api \
	v2-install-web v2-run-web v2-build-web \
	v2-docker-build v2-docker-up v2-docker-down v2-docker-logs

help:
	@echo "Project targets:"
	@echo "  legacy-help      查看 legacy 入口"
	@echo "  v2-help          查看 v2 入口"
	@echo "  v2-run-api       启动 v2 gc-api"
	@echo "  v2-test-api      运行 v2 gc-api 测试"
	@echo "  v2-build-api     构建 v2 gc-api"
	@echo "  v2-install-web   安装 v2 gc-web 依赖"
	@echo "  v2-run-web       启动 v2 gc-web 开发服务"
	@echo "  v2-build-web     构建 v2 gc-web"
	@echo "  v2-docker-build  构建 v2 docker 镜像"
	@echo "  v2-docker-up     启动 v2 docker compose"
	@echo "  v2-docker-down   停止 v2 docker compose"
	@echo "  v2-docker-logs   查看 v2 docker compose 日志"

legacy-help:
	@$(MAKE) -C legacy help

v2-help:
	@$(MAKE) -C v2 help

v2-paths:
	@$(MAKE) -C v2 paths

v2-run-api:
	@$(MAKE) -C v2 run-api

v2-test-api:
	@$(MAKE) -C v2 test-api

v2-build-api:
	@$(MAKE) -C v2 build-api

v2-install-web:
	@$(MAKE) -C v2 install-web

v2-run-web:
	@$(MAKE) -C v2 run-web

v2-build-web:
	@$(MAKE) -C v2 build-web

v2-docker-build:
	@$(MAKE) -C v2 docker-build

v2-docker-up:
	@$(MAKE) -C v2 docker-up

v2-docker-down:
	@$(MAKE) -C v2 docker-down

v2-docker-logs:
	@$(MAKE) -C v2 docker-logs
