# EMS 项目

救援物资管理系统，前端基于 Vue3 + Vite，后端为 Go 多服务（REST + gRPC），并使用 PostgreSQL。

## Quickstart

### 依赖准备

- Node.js: ^20.19.0 或 >=22.12.0
- Go: 1.20+
- Docker + Docker Compose

### 启动后端（Docker）

在仓库根目录执行：

```bash
make build-bin

make docker-up

或者
cd backend/deploy
docker compose up -d

```

默认端口：

- Auth REST: 8080
- User REST: 8081
- Warehouse REST: 8082
- Scheduling REST: 8083
- Statistics REST: 8084
- PostgreSQL: 5432
- gRPC: 9000-9004

后端配置文件：

- [global.yaml](file:///root/ems/backend/internal/common/config/global.yaml)

如需修改数据库连接、服务端口或高德地图 Key，可直接修改该文件。

### 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端开发代理已配置到后端 REST 端口，详见：

- [vite.config.ts](file:///root/ems/frontend/vite.config.ts)

### 可选：生成接口代码

OpenAPI 与 gRPC 代码生成命令：

```bash
make gen
```

对应脚本：

- [genopenapi.sh](file:///root/ems/backend/script/genopenapi.sh)
- [genproto.sh](file:///root/ems/backend/script/genproto.sh)
