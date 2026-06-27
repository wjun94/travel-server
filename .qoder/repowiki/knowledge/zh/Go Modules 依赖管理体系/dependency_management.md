该项目采用标准的 **Go Modules** 进行依赖管理，配合 Docker 多阶段构建实现依赖的隔离与固化。

### 1. 核心配置
- **模块定义**：在 `go.mod` 中声明模块名为 `travel-server`，指定 Go 版本为 `1.26.3`。
- **依赖锁定**：通过 `go.sum` 文件精确记录所有直接和间接依赖的哈希值，确保构建的可重复性。
- **主要技术栈**：
    - Web 框架：`github.com/gin-gonic/gin`
    - ORM：`gorm.io/gorm` + `gorm.io/driver/mysql`
    - 缓存：`github.com/go-redis/redis/v8`
    - 认证：`github.com/golang-jwt/jwt/v5`
    - 文档：`github.com/swaggo/swag` 及其相关组件

### 2. 构建与部署策略
- **Docker 多阶段构建**：
    - **Builder 阶段**：基于 `golang:1.26.3-alpine` 镜像，先复制 `go.mod` 和 `go.sum` 执行 `go mod download` 以利用 Docker 层缓存加速构建，随后编译二进制文件。
    - **Runtime 阶段**：使用轻量级的 `alpine:3.19` 镜像，仅包含编译后的二进制文件和必要的静态资源，显著减小了最终镜像体积。
- **开发环境热重载**：使用 `.air.toml` 配置 `air` 工具，监听代码变动并自动重新编译 `./cmd` 目录下的入口文件，排除 `vendor` 和 `tmp` 目录以提升效率。

### 3. 开发者规范
- **依赖更新**：应通过 `go get <package>@<version>` 命令更新依赖，并同步提交 `go.mod` 和 `go.sum` 的变更。
- **版本一致性**：本地开发环境的 Go 版本应与 `go.mod` 中声明的版本（1.26.3）及 Dockerfile 中的基础镜像版本保持一致，以避免潜在的兼容性问题。
- **私有依赖**：目前未发现针对私有仓库的特殊配置（如 `GOPRIVATE`），所有依赖均从公共代理或源站获取。若引入内部库，需在 CI/CD 环境中配置相应的访问令牌或代理。