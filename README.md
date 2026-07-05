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
│   │   └── deepseek.go            # DeepSeek 大模型客户端
│   ├── handler/
│   │   ├── common.go              # 天气查询 / WebSocket
│   │   ├── miniapp/               # 小程序端接口
│   │   │   ├── user.go            # 登录、个人信息
│   │   │   ├── guide.go           # 攻略 CRUD + 板块管理
│   │   │   ├── trip.go            # 行程 CRUD + AI 生成
│   │   │   ├── partner.go         # 搭子组队
│   │   │   ├── comment.go         # 评论与点赞
│   │   │   ├── favorite.go        # 收藏管理
│   │   │   ├── message.go         # 私聊消息
│   │   │   ├── accounting.go      # 记账管理
│   │   │   ├── checklist.go       # 备忘清单
│   │   │   ├── nearby.go          # 周边推荐
│   │   │   └── footprint.go       # 足迹与海报
│   │   └── admin/                 # 后台管理接口
│   │       ├── auth.go            # 登录认证
│   │       ├── dashboard.go       # 数据看板
│   │       ├── admin_users.go     # 管理员账号 CRUD
│   │       ├── user.go            # 小程序用户管理
│   │       ├── role.go            # 角色权限管理
│   │       ├── guide.go           # 攻略审核
│   │       ├── partner.go         # 官方搭子
│   │       └── recommendation.go  # TOP 推荐配置
│   ├── middleware/
│   │   ├── cors.go                # 跨域中间件
│   │   └── jwt.go                 # JWT 认证（小程序 + 管理员双通道）
│   ├── model/                     # GORM 数据模型（16 文件）
│   ├── repository/                # 数据库操作层（12 文件）
│   ├── service/                   # 业务逻辑层
│   └── ws/
│       └── hub.go                 # WebSocket 房间管理
├── pkg/
│   ├── config/
│   │   └── config.go              # 环境变量加载
│   ├── database/
│   │   ├── mysql.go               # MySQL 连接 & 雪花 ID 初始化
│   │   └── redis.go               # Redis 连接
│   ├── response/
│   │   └── response.go            # 统一 JSON 响应格式
│   └── snowflake/
│       └── snowflake.go           # 雪花算法分布式 ID 生成
├── docs/                          # Swagger 文档（swag init 自动生成）
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

### 公开接口（无需登录）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | /api/v1/user/login | 微信静默登录 |
| POST | /api/v1/admin/login | 后台管理员登录 |
| GET  | /api/v1/feed | 攻略瀑布流 |
| GET  | /api/v1/nearby | 周边推荐 |
| GET  | /api/v1/nearby/recommend | 本周 TOP 推荐 |
| GET  | /api/v1/weather | 高德天气查询 |
| GET  | /api/v1/weather/qweather | 和风天气查询 |
| GET  | /api/v1/comments | 评论列表 |
| GET  | /api/v1/comment/replies | 子回复列表 |

### 小程序端（需 JWT）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| **用户** | | |
| GET  | /api/v1/user/info | 个人信息 |
| PUT  | /api/v1/user/profile | 更新资料 |
| **攻略** | | |
| POST | /api/v1/guide | 创建攻略 |
| GET  | /api/v1/guide/:id | 攻略详情 |
| PUT  | /api/v1/guide/:id | 编辑攻略 |
| POST | /api/v1/guide/section | 添加板块 |
| PUT  | /api/v1/guide/section/:id | 更新板块 |
| DELETE | /api/v1/guide/section/:id | 删除板块 |
| PUT  | /api/v1/guide/sections/reorder | 板块排序 |
| **行程** | | |
| POST | /api/v1/trip | 创建行程 |
| POST | /api/v1/trip/ai-generate | AI 生成行程 |
| GET  | /api/v1/trip/:id | 行程详情 |
| PUT  | /api/v1/trip/:id | 编辑行程 |
| POST | /api/v1/trip/day | 添加日 |
| PUT  | /api/v1/trip/day/:id | 编辑日 |
| DELETE | /api/v1/trip/day/:id | 删除日 |
| POST | /api/v1/trip/item | 添加行程项 |
| PUT  | /api/v1/trip/item/:id | 编辑行程项 |
| DELETE | /api/v1/trip/item/:id | 删除行程项 |
| POST | /api/v1/trip/member | 邀请成员 |
| DELETE | /api/v1/trip/member/:id | 移除成员 |
| **搭子** | | |
| POST | /api/v1/partner | 发布搭子 |
| GET  | /api/v1/partner/list | 搭子列表 |
| POST | /api/v1/partner/:id/apply | 申请加入 |
| PUT  | /api/v1/partner/:id/application | 处理申请 |
| **消息** | | |
| GET  | /api/v1/message/list | 消息记录 |
| POST | /api/v1/message/send | 发送消息 |
| **记账** | | |
| GET  | /api/v1/account/:tripId | 查看账本 |
| POST | /api/v1/account | 添加记账 |
| POST | /api/v1/account/import | 导入微信支付账单 |
| **备忘清单** | | |
| GET  | /api/v1/checklist | 获取清单 |
| POST | /api/v1/checklist | 创建清单 |
| PUT  | /api/v1/checklist/:id/item | 更新勾选 |
| **足迹** | | |
| GET  | /api/v1/footprint | 查看足迹 |
| POST | /api/v1/footprint/sync | 同步足迹 |
| GET  | /api/v1/footprint/poster | 生成海报 |
| **收藏** | | |
| POST | /api/v1/favorite | 添加收藏 |
| DELETE | /api/v1/favorite/:id | 取消收藏 |
| GET  | /api/v1/favorites | 收藏列表 |
| **评论** | | |
| POST | /api/v1/comment | 发表评论 |
| POST | /api/v1/comment/:id/like | 点赞评论 |

### 后台管理（需 Admin JWT）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| **认证 & 仪表盘** | | |
| GET  | /api/v1/admin/info | 管理员信息 |
| GET  | /api/v1/admin/dashboard | 数据看板 |
| **管理员账号** | | |
| GET  | /api/v1/admin/admin/users | 管理员列表 |
| POST | /api/v1/admin/admin/user | 创建管理员 |
| PUT  | /api/v1/admin/admin/user/:id | 编辑管理员 |
| DELETE | /api/v1/admin/admin/user/:id | 删除管理员 |
| **角色权限** | | |
| GET  | /api/v1/admin/roles | 角色列表 |
| POST | /api/v1/admin/role | 创建角色 |
| PUT  | /api/v1/admin/role/:id | 编辑角色 |
| DELETE | /api/v1/admin/role/:id | 删除角色 |
| **小程序用户** | | |
| GET  | /api/v1/admin/users | 用户列表 |
| PUT  | /api/v1/admin/user/:id/role | 修改用户角色 |
| **攻略内容** | | |
| GET  | /api/v1/admin/guides | 攻略列表 |
| PUT  | /api/v1/admin/guide/:id/status | 审核攻略 |
| **官方搭子** | | |
| POST | /api/v1/admin/partner | 创建官方搭子 |
| GET  | /api/v1/admin/partners | 搭子列表 |
| **推荐内容** | | |
| POST | /api/v1/admin/recommendation | 保存推荐 |
| GET  | /api/v1/admin/recommendations | 推荐列表 |

> ✅ 所有认证接口需携带 Header `Authorization: Bearer <token>`
> API 完整文档请启动服务后访问 `http://localhost:8080/swagger/index.html`

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

## 🧪 本地开发（热重载）

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
