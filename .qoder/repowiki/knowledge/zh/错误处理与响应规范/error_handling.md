该后端服务采用基于 Gin 框架的**手动错误检查与统一响应封装**模式，未引入专门的错误管理库（如 `pkg/errors`）或全局错误中间件。

### 1. 核心机制：统一响应结构
- **响应包**：`pkg/response` 定义了标准的 JSON 响应结构 `Response`，包含 `code`（0 为成功，非 0 为失败）、`msg`（提示信息）和 `data`（业务数据）。
- **辅助函数**：
  - `response.Success(c, data)`：返回 HTTP 200 及标准成功格式。
  - `response.Fail(c, httpCode, msg)`：返回指定的 HTTP 状态码及错误信息，`code` 固定为 1。

### 2. 错误传播策略
- **Service 层**：使用标准库 `errors.New` 创建业务错误（如“用户名或密码错误”），或直接透传底层错误（如 GORM 错误）。部分逻辑利用 `errors.Is` 判断特定错误（如 `gorm.ErrRecordNotFound`）以执行自动注册等分支逻辑。
- **Handler 层**：负责捕获 Service 或 Repository 层的错误，并将其转换为面向前端的 HTTP 响应。通常通过 `if err != nil` 检查后调用 `response.Fail`。
- **Middleware 层**：在 JWT 认证等中间件中，遇到解析失败或权限不足时，直接调用 `c.AbortWithStatusJSON` 中断请求并返回特定的 HTTP 状态码（如 401 Unauthorized）。

### 3. 关键约定与局限
- **HTTP 状态码映射**：错误处理强依赖 HTTP 状态码来区分错误类型（400 参数错误、401 未授权、404 不存在、500 服务器内部错误）。
- **错误信息暴露**：部分 Handler（如 `internal/handler/admin/auth.go`）直接将底层错误信息 `err.Error()` 返回给前端，可能存在敏感信息泄露风险，建议在生产环境中进行脱敏或映射为通用提示。
- **缺乏全局恢复**：代码中未发现 `panic/recover` 机制，若发生未捕获的 panic，将由 Gin 默认行为处理或直接导致进程崩溃。
- **错误码标准化**：目前 `response.Fail` 中的 `Code` 字段硬编码为 1，缺乏细粒度的业务错误码定义（如 10001 表示余额不足），前端难以通过 `code` 进行精确的逻辑判断。