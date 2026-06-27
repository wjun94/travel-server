## 1. 系统概述
该后端服务采用**纯环境变量驱动**的配置管理方案。不依赖 YAML/JSON/TOML 等外部配置文件解析库（如 Viper），而是通过 Go 标准库 `os` 直接读取环境变量，并在应用启动时加载到全局单例结构体中。

## 2. 核心实现与文件
- **配置加载器**: `pkg/config/config.go`
  - 定义了 `Config` 结构体，涵盖数据库、Redis、第三方服务（微信、高德、DeepSeek、七牛云）等所有配置项。
  - 提供 `LoadConfig()` 函数，在 `main` 函数入口处调用，初始化全局变量 `AppConfig`。
  - 实现了 `getEnv`, `getEnvInt64` 等辅助函数，支持为每个配置项设置**默认值**（Fallback），确保在缺少环境变量时服务仍能启动（主要用于开发环境）。
- **环境变量模板**: `.env.example`, `.env.dev`, `.env.prod`
  - `.env.example`: 配置项模板，包含所有必需和可选的环境变量键名及注释。
  - `.env.dev` / `.env.prod`: 分别对应开发和生产环境的具体配置值。
- **入口逻辑**: `cmd/main.go`
  - 在 `main()` 函数第一行调用 `config.LoadConfig()`，确保后续数据库初始化和路由注册能获取到正确配置。

## 3. 架构与约定
- **全局单例模式**: 配置加载后存储在 `config.AppConfig` 全局指针变量中，项目中任何包均可通过 `config.AppConfig.Field` 访问配置，无需传递上下文或依赖注入。
- **环境隔离**: 通过 `ENV` 变量区分 `development` 和 `production` 环境。Docker Compose 部署时通过 `env_file: .env.prod` 指定生产环境配置。
- **敏感信息管理**: 
  - 密钥（如 `APPSECRET`, `QINIU_SECRET_KEY`, `DEEPSEEK_API_KEY`）直接存储在 `.env` 文件中。
  - 证书路径（如微信支付私钥 `WECHATPAYPRIVATEKEYPATH`）配置为相对路径（如 `./certs/apiclient_key.pem`），依赖文件系统挂载。
- **默认值策略**: 代码中硬编码了开发友好的默认值（如本地 MySQL `127.0.0.1:3306`，Redis `127.0.0.1:6379`），降低了本地开发环境的配置门槛。

## 4. 开发者规范
- **新增配置项**: 
  1. 在 `pkg/config/config.go` 的 `Config` 结构体中添加字段。
  2. 在 `LoadConfig()` 中使用 `getEnv("KEY", "default")` 进行映射。
  3. 在 `.env.example` 中添加对应的键名和说明。
- **安全约束**: 
  - **严禁**将 `.env.dev` 或 `.env.prod` 提交至版本控制系统（已在 `.gitignore` 中忽略）。
  - 生产环境部署时，应通过 CI/CD 流水线或容器编排工具（如 Kubernetes Secrets）注入环境变量，而非直接使用 `.env` 文件。
- **类型安全**: 对于非字符串类型（如 `int64`, `float64`），必须使用专用的 `getEnvInt64` 等函数进行解析，避免在业务代码中重复处理类型转换错误。