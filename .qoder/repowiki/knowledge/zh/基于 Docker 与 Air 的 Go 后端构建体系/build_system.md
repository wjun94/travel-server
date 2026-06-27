## 1. 构建系统概览
该项目采用 **Go Modules** 进行依赖管理，使用 **Docker** 和 **Docker Compose** 作为核心的构建、打包及部署工具。开发环境引入了 **Air** 实现代码热重载（Hot Reload），生产环境则通过多阶段构建（Multi-stage Build）生成轻量级镜像。

## 2. 核心构建流程
### 2.1 依赖与文档生成
- **依赖管理**：通过 `go.mod` 锁定版本，当前 Go 版本要求为 `1.26.3`。
- **API 文档**：在构建过程中自动集成 `swag` 工具，根据 `cmd/main.go` 中的注释生成 Swagger 文档（`docs/` 目录）。

### 2.2 生产环境构建 (Dockerfile)
采用两阶段构建策略以减小镜像体积：
1. **Builder 阶段**：基于 `golang:1.26.3-alpine`，安装 `swag`，下载依赖并编译二进制文件 `photo-print-backend`。
2. **Runtime 阶段**：基于 `alpine:3.19`，仅复制编译后的二进制文件和必要的静态资源（如 `static/` 目录）。

### 2.3 开发环境构建 (Dockerfile.dev)
- 基于 `golang:1.26.3-alpine`。
- 预装 `air` 和 `swag`。
- 启动时先执行 `swag init` 更新文档，随后运行 `air` 监听文件变化并自动重启服务。

## 3. 编排与环境隔离
项目通过两套 Docker Compose 配置实现环境隔离：
- **生产环境 (`docker-compose.yml`)**：
  - 使用 `Dockerfile` 构建。
  - 加载 `.env.prod` 环境变量。
  - 容器命名：`travel-server`, `travel-db`, `travel-redis`。
  - 依赖健康检查：确保 MySQL 和 Redis 就绪后再启动后端。
- **开发环境 (`docker-compose.dev.yml`)**：
  - 使用 `Dockerfile.dev` 构建。
  - 加载 `.env.dev` 环境变量。
  - 容器命名增加 `-dev` 后缀，数据卷独立（如 `mysql_data_dev`），避免污染生产数据。

## 4. 开发者规范
- **本地开发**：建议直接使用 `docker-compose -f docker-compose.dev.yml up` 启动全套服务，利用 Air 实现修改代码即生效。
- **端口映射**：后端服务默认映射至宿主机 `8082` 端口，MySQL 映射至 `3307`，Redis 映射至 `6379`。
- **文档维护**：每次修改 API 接口注释后，若未通过 Docker 启动，需手动运行 `swag init -g cmd/main.go -o docs` 以保持文档同步。