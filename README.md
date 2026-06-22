# 旅行搭子小程序后端服务

基于 Gin + GORM + MySQL + Redis + DeepSeek 的旅游社交平台后端，为微信小程序提供 **攻略分享、行程规划、搭子组队、协同编辑、周边推荐、记账备忘、足迹海报** 等一站式服务，并配套后台管理系统。

---

## 🚀 功能特性

- **微信登录**：静默授权获取用户身份，JWT 认证。
- **瀑布流攻略**：双列图文展示，支持发布、审核、点赞分享。
- **智能行程**：对接 DeepSeek 大模型，一键生成个性化旅行计划；同时支持手动创建。
- **实时协同编辑**：基于 WebSocket，好友可同时编辑同一份行程，版本冲突乐观锁控制。
- **周边游推荐**：接入高德地图 API，按当前位置搜索景区、民宿，支持一键导航和天气查询。
- **搭子组队**：发布搭子需求，一键分享到微信群，申请/同意后自动建立私聊。
- **消息中心**：支持私聊、系统通知，新消息可通过微信订阅消息提醒。
- **旅行记账**：按行程管理收支，支持从微信支付账单快速导入。
- **备忘清单**：自定义待办事项，可勾选、置顶，提供官方模板。
- **足迹点亮**：根据旅行记录自动点亮城市，生成专属足迹海报。
- **后台管理**：仪表盘统计、用户角色管理、内容审核、官方搭子团/领队、TOP 推荐配置。

---

## 🛠 技术栈

| 类别     | 技术选型                     |
| -------- | ---------------------------- |
| 语言     | Go 1.20+                     |
| Web 框架 | Gin                          |
| 数据库   | MySQL (GORM) + Redis         |
| 认证     | JWT (golang-jwt)             |
| 实时通信 | Gorilla WebSocket            |
| AI 模型  | DeepSeek Chat API            |
| 地图服务 | 高德地图 Web API             |
| 配置管理 | 环境变量 (os.LookupEnv)      |
| 文档     | Swagger (swaggo/gin-swagger) |

---

## 📁 目录结构

```
travel-server/
├── cmd/
│   └── main.go                     # 服务入口，路由注册
├── internal/
│   ├── ai/
│   │   └── deepseek.go            # DeepSeek 聊天客户端
│   ├── handler/
│   │   ├── common.go              # 天气、WebSocket 公共处理
│   │   ├── miniapp/               # 小程序端接口
│   │   │   ├── user.go            # 登录、用户信息
│   │   │   ├── post.go            # 攻略瀑布流
│   │   │   ├── trip.go            # 行程创建/编辑/协同
│   │   │   ├── nearby.go          # 周边推荐
│   │   │   ├── partner.go         # 搭子组队
│   │   │   ├── message.go         # 私聊消息
│   │   │   ├── accounting.go      # 记账管理
│   │   │   ├── checklist.go       # 备忘清单
│   │   │   └── footprint.go       # 足迹与海报
│   │   └── admin/                 # 后台管理接口
│   │       ├── dashboard.go       # 数据看板
│   │       ├── user.go            # 用户管理
│   │       ├── post.go            # 内容审核
│   │       ├── partner.go         # 官方搭子
│   │       └── recommendation.go  # 推荐配置
│   ├── middleware/
│   │   ├── cors.go                # 跨域中间件
│   │   └── jwt.go                 # JWT + 管理员权限
│   ├── model/                     # GORM 数据模型
│   ├── repository/                # 数据库操作层
│   ├── service/                   # 业务逻辑层
│   └── ws/
│       └── hub.go                 # WebSocket 房间管理
├── pkg/
│   ├── config/
│   │   └── config.go              # 环境变量配置加载
│   ├── database/
│   │   ├── mysql.go               # MySQL 连接 & 自动迁移
│   │   └── redis.go               # Redis 连接
│   └── response/
│       └── response.go            # 统一 JSON 响应
├── docs/                          # Swagger 文档（自动生成）
├── .env.example                   # 环境变量模板
├── go.mod
├── go.sum
└── README.md
```

---

## ⚙️ 快速开始

### 1. 环境要求

- Go 1.20+
- MySQL 8.0+ (已创建数据库 `travel`)
- Redis 7.0+
- 微信小程序 AppID & AppSecret
- [高德地图 Web API Key](https://lbs.amap.com/)
- [DeepSeek API Key](https://platform.deepseek.com/)

### 2. 配置环境变量

复制 `.env.example` 并修改为真实值，或直接在系统中 export 环境变量：

```bash
export SERVER_PORT=8080

export DB_HOST=127.0.0.1
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password
export DB_NAME=travel

export REDIS_HOST=127.0.0.1
export REDIS_PORT=6379
export REDIS_PASSWORD=
export REDIS_DB=0

export APPID=wx1234567890
export APPSECRET=your_app_secret

export AMAP_KEY=your_amap_key
export DEEPSEEK_API_KEY=sk-xxxx
```

### 3. 安装依赖并运行

生成 Swag 文档

```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -o docs
```

```bash
cd travel-server
go mod tidy
swag init -g cmd/main.go -o docs --parseDependency --parseInternal
go run cmd/main.go
```

服务默认监听 `http://0.0.0.0:8080`。

### 4. 访问 Swagger 文档

启动后打开浏览器访问：

> http://localhost:8080/swagger/index.html

可在线调试所有 API。

---

## 📡 API 概览

### 小程序端接口

| 接口路径                        | 方法 | 说明              | 认证             |
| ------------------------------- | ---- | ----------------- | ---------------- |
| /api/v1/user/login              | POST | 微信登录          | ❌               |
| /api/v1/feed                    | GET  | 瀑布流攻略        | ❌               |
| /api/v1/nearby                  | GET  | 周边推荐          | ❌               |
| /api/v1/nearby/recommend        | GET  | 本周 TOP 推荐     | ❌               |
| /api/v1/weather                 | GET  | 天气查询          | ❌               |
| /api/v1/user/info               | GET  | 获取个人信息      | ✅               |
| /api/v1/user/profile            | PUT  | 更新个人资料      | ✅               |
| /api/v1/post                    | POST | 发布攻略          | ✅               |
| /api/v1/post/:id                | GET  | 攻略详情          | ✅               |
| /api/v1/trip                    | POST | 手动创建行程      | ✅               |
| /api/v1/trip/ai-generate        | POST | AI 生成行程       | ✅               |
| /api/v1/trip/:id                | GET  | 行程详情          | ✅               |
| /api/v1/trip/:id                | PUT  | 协同编辑行程      | ✅               |
| /api/v1/trip/:id/invite         | POST | 邀请协同编辑      | ✅               |
| /api/v1/partner                 | POST | 发布搭子          | ✅               |
| /api/v1/partner/list            | GET  | 搭子列表          | ✅               |
| /api/v1/partner/:id/apply       | POST | 申请加入搭子      | ✅               |
| /api/v1/partner/:id/application | PUT  | 处理申请          | ✅               |
| /api/v1/message/list            | GET  | 消息记录          | ✅               |
| /api/v1/message/send            | POST | 发送消息          | ✅               |
| /api/v1/account/:tripId         | GET  | 查看账本          | ✅               |
| /api/v1/account                 | POST | 添加记账          | ✅               |
| /api/v1/account/import          | POST | 导入微信支付账单  | ✅               |
| /api/v1/checklist               | GET  | 获取备忘清单      | ✅               |
| /api/v1/checklist               | POST | 创建清单          | ✅               |
| /api/v1/checklist/:id/item      | PUT  | 更新勾选状态      | ✅               |
| /api/v1/footprint               | GET  | 查看足迹          | ✅               |
| /api/v1/footprint/sync          | POST | 同步足迹          | ✅               |
| /api/v1/footprint/poster        | GET  | 生成足迹海报      | ✅               |
| /ws                             | WS   | 协同编辑+实时消息 | ✅ (query token) |

### 后台管理接口 (需 Admin 权限)

| 接口路径                       | 方法 | 说明           |
| ------------------------------ | ---- | -------------- |
| /api/v1/admin/dashboard        | GET  | 数据看板       |
| /api/v1/admin/users            | GET  | 用户列表       |
| /api/v1/admin/user/:id/role    | PUT  | 修改用户角色   |
| /api/v1/admin/posts            | GET  | 攻略列表       |
| /api/v1/admin/post/:id/status  | PUT  | 审核攻略状态   |
| /api/v1/admin/official-partner | POST | 发布官方搭子团 |
| /api/v1/admin/recommendation   | POST | 保存推荐内容   |
| /api/v1/admin/recommendations  | GET  | 推荐列表       |

> ✅ 表示需在 Header 中携带 `Authorization: Bearer <token>`

---

## 🔌 WebSocket 通信

连接地址：`ws://localhost:8080/ws?token=jwt_token`

### 协同编辑行程

```json
// 1. 加入行程房间
{"action": "join_trip", "trip_id": "1"}

// 2. 编辑行程（服务器将广播给房间内其他用户）
{
  "action": "edit_trip",
  "trip_id": "1",
  "daily_plans": {
    "days": [
      {
        "day": 1,
        "items": [
          { "time": "09:00", "name": "西湖", "type": "attraction", "duration": "2h" }
        ]
      }
    ]
  }
}
```

---

## 📦 部署

推荐使用 Docker 容器化部署：

```dockerfile
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o server cmd/main.go

FROM alpine:latest
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

构建并运行：

```bash
docker build -t travel-server .
docker run -d --env-file .env -p 8080:8080 travel-server
```

## Docker 部署

## 🧪 开发模式（热重载 + 实时日志）

为了方便本地开发，我们提供了独立的开发环境配置文件，支持**代码修改自动重启**，无需手动重新构建镜像。

### 使用步骤

1. **安装 Docker Desktop**（确保 docker-compose 插件可用）

2. **启动开发环境**

```bash
docker-compose -f docker-compose.dev.yml up -d --build
```

3. **查看实时日志**

```bash
docker-compose -f docker-compose.dev.yml logs -f backend
```

4. **修改代码**  
   编辑任意 `.go` 文件，保存后 Air 会自动检测变化 → 重新编译 → 重启服务。  
   无需重新构建镜像或重启容器。

5. **访问服务**  
   同生产模式：`http://localhost:8080/admin`

6. **重启开发环境**

```bash
docker-compose -f docker-compose.dev.yml restart
```

7. **停止但不删除数据**

```bash
docker-compose -f docker-compose.dev.yml down
```

8. **停止并彻底清空数据（重置环境）**

```bash
docker-compose -f docker-compose.dev.yml down -v
```

9. **启动开发环境**

```bash
docker-compose -f docker-compose.dev.yml up -d
```

10. **进入容器**

```bash
docker exec -it travel-backend-dev sh
```

---

## 📝 许可证

MIT License

---

**享受你的旅程！** ✈️  
如果遇到问题，欢迎提交 Issue。
