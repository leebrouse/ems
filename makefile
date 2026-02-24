# ----------------------
# REST 和 gRPC 代码生成相关
# ----------------------

.PHONY: help gen-openapi gen-proto clean 

help: ## 显示帮助信息
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

gen-openapi: ## 生成 OpenAPI 代码 (从 YAML 生成 Go 类型和服务器接口)
	@echo "生成 OpenAPI 代码..."
	./backend/script/genopenapi.sh

gen-proto: ## 生成 gRPC 代码 (从 .proto 文件生成 Go gRPC 代码)
	@echo "生成 gRPC 代码..."
	./backend/script/genproto.sh

gen: gen-openapi gen-proto ## 生成所有代码 (OpenAPI + gRPC)

clean:  ## 清理生成的文件
	@echo "清理生成的代码文件..."
	rm -rf ./backend/internal/common/genopenapi/
	rm -rf ./backend/internal/common/genproto/

# ----------------------
# 服务编译相关
# ----------------------

.PHONY: build-auth build-user build-warehouse build-scheduling build-statistics build-bin clean-bin

build-auth: ## 编译 auth 服务
	@echo "编译 auth 服务..."
	cd backend/internal/auth && go build -o auth main.go

build-user: ## 编译 user 服务
	@echo "编译 user 服务..."
	cd backend/internal/user && go build -o user main.go

build-warehouse: ## 编译 warehouse 服务
	@echo "编译 warehouse 服务..."
	cd backend/internal/warehouse && go build -o warehouse main.go

build-scheduling: ## 编译 scheduling 服务
	@echo "编译 scheduling 服务..."
	cd backend/internal/scheduling && go build -o scheduling main.go

build-statistics: ## 编译 statistics 服务
	@echo "编译 statistics 服务..."
	cd backend/internal/statistics && go build -o statistics main.go

build-bin: build-auth build-user build-warehouse build-scheduling build-statistics ## 编译所有服务二进制

clean-bin: ## 清理编译后的二进制文件
	@echo "清理编译文件..."
	rm -f backend/internal/auth/auth
	rm -f backend/internal/user/user
	rm -f backend/internal/warehouse/warehouse
	rm -f backend/internal/scheduling/scheduling
	rm -f backend/internal/statistics/statistics

# ----------------------
# Docker 部署相关
# ----------------------

.PHONY: docker-up docker-down docker-build docker-restart

docker-up: ## 启动所有 Docker 容器
	@echo "启动 Docker 容器..."
	docker compose -f backend/deploy/docker-compose.yaml up -d

docker-down: ## 停止并移除 Docker 容器
	@echo "停止 Docker 容器..."
	docker compose -f backend/deploy/docker-compose.yaml down

docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker compose -f backend/deploy/docker-compose.yaml build

docker-restart: docker-down docker-up ## 重启 Docker 容器