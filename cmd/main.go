/*
 * 旅行搭子小程序后端服务入口
 * 初始化配置 → 数据库 → 注册路由 → 启动 HTTP 服务
 */
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "travel-server/docs"
	"travel-server/internal/handler/admin"
	"travel-server/internal/handler/common"
	"travel-server/internal/handler/miniapp"
	"travel-server/internal/middleware"
	"travel-server/pkg/config"
	"travel-server/pkg/database"
)

// @title           旅行搭子小程序 API
// @version         1.0
// @description     微信小程序旅游社交平台后端接口，提供攻略、行程、搭子、记账、周边游等功能
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 启动初始化
	config.LoadConfig()
	database.InitMySQL()
	database.InitRedis()

	// 创建 Gin 引擎 & 全局中间件
	r := gin.Default()
	r.Use(middleware.Cors())

	// Swagger 文档面板
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		// ==================== 公开接口（无需登录） ====================
		api.POST("/user/login", miniapp.UserLogin)            // 微信登录
		api.POST("/admin/login", admin.AdminLogin)            // 后台登录
		api.GET("/guides", miniapp.GetGuideFeed)              // 攻略瀑布流
		api.GET("/nearby", miniapp.GetNearby)                 // 周边推荐
		api.GET("/nearby/recommend", miniapp.GetTopRecommend) // TOP推荐
		api.GET("/weather", common.GetWeather)                // 天气查询
		api.GET("/weather/qweather", common.GetQWeather)      // 天气查询（和风）
		api.GET("/comments", miniapp.GetComments)             // 评论列表
		api.GET("/comment/replies", miniapp.GetReplies)       // 子回复列表

		// ==================== 小程序端（需 JWT 登录） ====================
		miniAuth := api.Group("", middleware.JWTAuth())
		{
			// ---------- 用户 ----------
			miniAuth.GET("/user/info", miniapp.GetUserInfo)
			miniAuth.PUT("/user/profile", miniapp.UpdateProfile)

			// ---------- 攻略 ----------
			miniAuth.POST("/guide", miniapp.CreateGuide)
			miniAuth.GET("/guide/:id", miniapp.GetGuideDetail)
			miniAuth.PUT("/guide/:id", miniapp.UpdateGuide)
			miniAuth.POST("/guide/:id/day", miniapp.CreateGuideDay)            // 新增一天
			miniAuth.DELETE("/guide/day/:id", miniapp.DeleteGuideDay)          // 删除一天
			miniAuth.POST("/guide/day/:id/item", miniapp.CreateGuideDayItem)   // 新增行程项
			miniAuth.PUT("/guide/day/item/:id", miniapp.UpdateGuideDayItem)    // 更新行程项
			miniAuth.DELETE("/guide/day/item/:id", miniapp.DeleteGuideDayItem) // 删除行程项

			// ---------- 行程 ----------
			miniAuth.POST("/trip", miniapp.CreateTrip)
			miniAuth.POST("/trip/ai-generate", miniapp.AIGenerateTrip)
			miniAuth.GET("/trip/:id", miniapp.GetTrip)
			miniAuth.PUT("/trip/:id", miniapp.UpdateTrip)
			miniAuth.POST("/trip/day", miniapp.AddTripDay)
			miniAuth.PUT("/trip/day/:id", miniapp.UpdateTripDay)
			miniAuth.DELETE("/trip/day/:id", miniapp.DeleteTripDay)
			miniAuth.POST("/trip/item", miniapp.AddTripItem)
			miniAuth.PUT("/trip/item/:id", miniapp.UpdateTripItem)
			miniAuth.DELETE("/trip/item/:id", miniapp.DeleteTripItem)
			miniAuth.POST("/trip/member", miniapp.InviteMember)
			miniAuth.DELETE("/trip/member/:id", miniapp.RemoveMember)

			// ---------- 搭子 ----------
			miniAuth.POST("/partner", miniapp.CreatePartner)
			miniAuth.GET("/partner/list", miniapp.GetPartnerList)
			miniAuth.POST("/partner/:id/apply", miniapp.ApplyPartner)
			miniAuth.PUT("/partner/:id/application", miniapp.HandleApplication)

			// ---------- 消息 ----------
			miniAuth.GET("/message/list", miniapp.GetMessageList)
			miniAuth.POST("/message/send", miniapp.SendMessage)

			// ---------- 记账 ----------
			miniAuth.GET("/account/:tripId", miniapp.GetAccounts)
			miniAuth.POST("/account", miniapp.AddAccount)
			miniAuth.POST("/account/import", miniapp.ImportWechatPay)

			// ---------- 备忘清单 ----------
			miniAuth.GET("/checklist", miniapp.GetChecklists)
			miniAuth.POST("/checklist", miniapp.CreateChecklist)
			miniAuth.PUT("/checklist/:id/item", miniapp.UpdateChecklistItem)

			// ---------- 足迹 ----------
			miniAuth.GET("/footprint", miniapp.GetFootprints)
			miniAuth.POST("/footprint/sync", miniapp.SyncFootprint)
			miniAuth.GET("/footprint/poster", miniapp.GeneratePoster)

			// ---------- 收藏 ----------
			miniAuth.POST("/favorite", miniapp.AddFavorite)
			miniAuth.DELETE("/favorite/:id", miniapp.RemoveFavorite)
			miniAuth.GET("/favorites", miniapp.GetFavorites)

			// ---------- 评论 ----------
			miniAuth.POST("/comment", miniapp.CreateComment)
			miniAuth.POST("/comment/:id/like", miniapp.LikeComment)
		}

		// ==================== 后台管理（需管理员 JWT） ====================
		adminGroup := api.Group("/admin", middleware.AdminJWTAuth())
		{
			// ---------- 认证 ----------
			adminGroup.GET("/info", admin.GetAdminInfo)

			// ---------- 仪表盘 ----------
			adminGroup.GET("/dashboard", admin.Dashboard)

			// ---------- 后台管理员 ----------
			adminGroup.GET("/admin/users", admin.ListAdminUsers)
			adminGroup.POST("/admin/user", admin.CreateAdminUser)
			adminGroup.PUT("/admin/user/:id", admin.UpdateAdminUser)
			adminGroup.DELETE("/admin/user/:id", admin.DeleteAdminUser)

			// ---------- 角色权限 ----------
			adminGroup.GET("/roles", admin.ListRoles)
			adminGroup.POST("/role", admin.CreateRole)
			adminGroup.PUT("/role/:id", admin.UpdateRole)
			adminGroup.DELETE("/role/:id", admin.DeleteRole)

			// ---------- 小程序用户 ----------
			adminGroup.GET("/users", admin.ListUsers)
			adminGroup.PUT("/user/:id/role", admin.UpdateUserRole)

			// ---------- 攻略内容 ----------
			adminGroup.GET("/guides", admin.ListGuides)
			adminGroup.PUT("/guide/:id/status", admin.UpdateGuideStatus)

			// ---------- 官方搭子 ----------
			adminGroup.POST("/partner", admin.CreatePartner)
			adminGroup.GET("/partners", admin.ListPartners)

			// ---------- 推荐内容 ----------
			adminGroup.POST("/recommendation", admin.SaveRecommendation)
			adminGroup.GET("/recommendations", admin.ListRecommendations)
		}
	}

	// WebSocket 协同编辑
	r.GET("/ws", common.WebSocketHandler)

	log.Println("Server running on :", config.AppConfig.ServerPort)
	r.Run(":" + config.AppConfig.ServerPort)
}
