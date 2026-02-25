# EMS（救援物资管理系统）

一个面向应急救援场景的物资管理后台：包含用户与角色（RBAC）、仓储与库存、需求申请与调度、运输状态追踪、统计看板与地图可视化。

## 技术栈

### 前端（frontend）

- 框架：Vue 3 + TypeScript（Composition API / Script Setup）
- 构建：Vite
- UI：Element Plus（暗黑模式）
- 状态管理：Pinia
- 路由：vue-router（基于 meta.roles 的 RBAC 路由守卫）
- 网络：Axios（JWT Bearer Token 请求拦截、401 统一处理）
- 可视化：ECharts（统计图表）
- 地图：高德地图 JSAPI（运输轨迹可视化）

相关实现位置：

- 入口：[main.ts](file:///root/ems/frontend/src/main.ts)
- 路由与 RBAC：[router/index.ts](file:///root/ems/frontend/src/router/index.ts)
- 请求封装：[request.ts](file:///root/ems/frontend/src/api/request.ts)
- 登录态：[auth store](file:///root/ems/frontend/src/stores/auth.ts)
- 开发代理：[vite.config.ts](file:///root/ems/frontend/vite.config.ts)

### 后端（backend）

- 语言：Go（多模块：common + 各服务独立 go.mod）
- Web 框架：Gin（REST）
- RPC：gRPC（服务间调用）
- 配置：Viper（读取 YAML + 环境变量覆盖）
- ORM：GORM + PostgreSQL
- 接口定义与代码生成：
  - OpenAPI（oapi-codegen 生成 types/client/server 接口）
  - Protocol Buffers（protoc 生成 gRPC 代码）
- 部署：Docker + Docker Compose

默认配置文件（Docker 环境）：

- [global.yaml](file:///root/ems/backend/internal/common/config/global.yaml)

## 环境要求

- Node.js：^20.19.0 或 >=22.12.0（见 [package.json](file:///root/ems/frontend/package.json)）
- Go：1.24+（各后端服务 go.mod 指定 go 1.24.11）
- Docker + Docker Compose
- GNU Make（可选，但推荐用来统一入口）

## 项目结构

```text
ems/
  frontend/                 # Vue3 管理后台
  backend/                  # Go 多服务后端
  makefile                  # 常用命令（代码生成/构建/Docker）
  README.md
```

后端服务（REST + gRPC）：

- auth：认证服务（REST 8080 / gRPC 9000）
- user：用户服务（REST 8081 / gRPC 9001）
- warehouse：仓储服务（REST 8082 / gRPC 9002）
- scheduling：调度服务（REST 8083 / gRPC 9003）
- statistics：统计服务（REST 8084 / gRPC 9004）

## Quickstart（推荐：Docker 启动后端 + 本地启动前端）

### 1) 启动后端（Docker Compose）

在仓库根目录执行：

```bash
make docker-up
```

或直接执行：

```bash
docker compose -f backend/deploy/docker-compose.yaml up -d
```

默认会启动 PostgreSQL 与 5 个后端服务，并映射端口到本机：

- REST：8080-8084
- gRPC：9000-9004
- PostgreSQL：5432

### 2) 启动前端（本地）

```bash
cd frontend
npm install
npm run dev
```

前端开发环境通过 Vite 代理将 `/api/v1/**` 转发到本机的后端端口，配置见：

- [vite.config.ts](file:///root/ems/frontend/vite.config.ts)

## 工具与命令

### Makefile 常用目标

在根目录运行 `make help` 可查看所有命令说明：

```bash
make help
```

常用目标（见 [makefile](file:///root/ems/makefile)）：

- `make docker-up` / `make docker-down` / `make docker-restart`：启动/停止/重启 Docker 容器
- `make docker-build`：构建后端镜像
- `make gen`：生成 OpenAPI + gRPC 代码
- `make clean`：清理生成的代码
- `make build-bin`：编译所有后端服务二进制

### 代码生成（OpenAPI / Protobuf）

生成命令：

```bash
make gen
```

脚本位置：

- OpenAPI：[genopenapi.sh](file:///root/ems/backend/script/genopenapi.sh)
- Protobuf：[genproto.sh](file:///root/ems/backend/script/genproto.sh)

工具依赖（需自行安装并加入 PATH）：

- oapi-codegen
- protoc
- protoc-gen-go / protoc-gen-go-grpc

## 配置说明

### 后端配置加载

后端通过 Viper 读取 `global.yaml`，并支持用环境变量覆盖配置项（容器内通过挂载目录读取配置）：

- 配置目录挂载：`backend/internal/common/config`
- Docker Compose 已默认挂载该目录到每个服务容器中

注意：

- Docker 默认配置中各服务 host 使用容器名（例如 `auth/user/postgres`），如果改为本机直接运行服务，需要将 host 调整为 `localhost` 或对应地址。
- 仓库中包含地图 key 与加密 key 等敏感配置，建议在私有环境中替换为自己的值，并避免在公共仓库暴露。

### 前端请求与鉴权

- 请求统一走 Axios 实例：[request.ts](file:///root/ems/frontend/src/api/request.ts)
- JWT token 存放在 Pinia + LocalStorage：[auth store](file:///root/ems/frontend/src/stores/auth.ts)
- 路由基于 `meta.public` 与 `meta.roles` 做权限控制：[router/index.ts](file:///root/ems/frontend/src/router/index.ts)

## 开发建议

- 优先使用 `make help` 与 Makefile 统一入口管理生成/构建/部署，减少手动命令分散。
- 修改接口协议时，先更新 `backend/api` 下的 OpenAPI / proto，再运行 `make gen` 生成代码，最后实现 handler/service 层逻辑。
