/*
 * 旅行搭子小程序后端服务入口
 * 初始化配置 → 数据库 → 注册路由 → 启动 HTTP 服务
 */
package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "travel-server/docs"
	"travel-server/internal/handler/admin"
	"travel-server/internal/handler/common"
	"travel-server/internal/handler/miniapp"
	"travel-server/internal/middleware"
	"travel-server/internal/repository"
	"travel-server/pkg/config"
	"travel-server/pkg/database"
	"travel-server/pkg/qiniu"
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
	qiniu.InitQiniu() // 新增七牛云

	// 创建 Gin 引擎 & 全局中间件
	r := gin.Default()
	r.Use(middleware.Cors())

	// 启动定时清理浏览历史（每天凌晨清理30天前的记录）
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
			if next.Before(now) {
				next = next.Add(24 * time.Hour)
			}
			time.Sleep(next.Sub(now))
			repository.CleanupBrowseHistory(30)
			log.Println("浏览历史清理完成")
		}
	}()

	// Swagger 文档面板
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		// ==================== 公开接口（无需登录） ====================
		api.POST("/user/login", miniapp.UserLogin)                 // 微信登录
		api.POST("/admin/login", admin.AdminLogin)                 // 后台登录
		api.GET("/guides", miniapp.GetGuideFeed)                   // 攻略瀑布流
		api.GET("/nearby", miniapp.GetNearby)                      // 周边推荐
		api.GET("/nearby/recommend", miniapp.GetTopRecommend)      // TOP推荐
		api.GET("/weather", common.GetWeather)                     // 天气查询
		api.GET("/weather/qweather", common.GetQWeather)           // 天气查询（和风）
		api.GET("/comments", miniapp.GetComments)                  // 评论列表
		api.GET("/comment/replies", miniapp.GetReplies)            // 子回复列表
		api.GET("/regions/domestic", miniapp.GetDomesticRegions)   // 国内省/市列表
		api.GET("/regions/countries", miniapp.GetCountries)        // 境外国家列表
		api.GET("/destinations/search", miniapp.SearchDestination) // 目的地搜索

		// ==================== 小程序端（需 JWT 登录） ====================
		miniAuth := api.Group("", middleware.JWTAuth())
		{
			// ---------- 媒体/文件上传 ----------
			miniAuth.POST("/upload/single", common.UploadSingleImage) // 单图（新增）
			miniAuth.DELETE("/image/delete", common.DeleteImage)      // 删除单照片
			miniAuth.POST("/upload/batch", common.UploadImages)       // 批量上传(没用到)

			// ---------- 用户 ----------
			miniAuth.GET("/profile", miniapp.GetMyProfile) // 个人主页（攻略数/行程数/粉丝/关注）
			miniAuth.GET("/user/info", miniapp.GetUserInfo)
			miniAuth.PUT("/user/profile", miniapp.UpdateProfile)

			// ---------- 攻略 ----------
			miniAuth.POST("/guide", miniapp.CreateGuide)
			miniAuth.GET("/guide/:id", miniapp.GetGuideDetail)
			miniAuth.PUT("/guide/:id", miniapp.UpdateGuide)
			miniAuth.POST("/guide/:id/day", miniapp.CreateGuideDay)            // 新增一天
			miniAuth.DELETE("/guide/day/:id", miniapp.DeleteGuideDay)          // 删除一天
			miniAuth.POST("/guide/:id/like", miniapp.LikeGuide)                // 点赞攻略
			miniAuth.DELETE("/guide/:id/like", miniapp.UnlikeGuide)            // 取消点赞
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
			miniAuth.GET("/checklist/categories", miniapp.GetChecklistCategories)
			miniAuth.GET("/checklist/:id", miniapp.GetChecklistDetail)
			miniAuth.POST("/checklist", miniapp.CreateChecklist)
			miniAuth.PUT("/checklist/:id", miniapp.UpdateChecklist)
			miniAuth.DELETE("/checklist/:id", miniapp.DeleteChecklist)
			miniAuth.PUT("/checklist/:id/item", miniapp.UpdateChecklistItem)

			// ---------- 足迹 ----------
			miniAuth.GET("/footprint", miniapp.GetFootprints)
			miniAuth.POST("/footprint/sync", miniapp.SyncFootprint)
			miniAuth.GET("/footprint/poster", miniapp.GeneratePoster)

			// ---------- 收藏 ----------
			miniAuth.POST("/favorite", miniapp.AddFavorite)
			miniAuth.POST("/favorite/remove", miniapp.RemoveFavorite)
			miniAuth.GET("/favorites", miniapp.GetFavorites)

			// ---------- 评论 ----------
			miniAuth.POST("/comment", miniapp.CreateComment)
			miniAuth.DELETE("/comment/:id", miniapp.DeleteComment)
			miniAuth.POST("/comment/:id/like", miniapp.LikeComment)

			// ---------- 浏览历史 ----------
			miniAuth.POST("/browse/history", miniapp.AddBrowseHistory)
			miniAuth.GET("/browse/history", miniapp.GetBrowseHistory)
			miniAuth.DELETE("/browse/history/clear", miniapp.ClearBrowseHistory)
			miniAuth.DELETE("/browse/history/:id", miniapp.DeleteBrowseHistory)

			// ---------- 关注 ----------
			miniAuth.POST("/follow/:id", miniapp.FollowUser)                    // 1.关注
			miniAuth.DELETE("/follow/:id", miniapp.UnfollowUser)                // 2.取消关注
			miniAuth.GET("/follow/following", miniapp.GetMyFollowingList)       // 3.我的关注
			miniAuth.GET("/follow/followers", miniapp.GetMyFollowerList)        // 4.我的粉丝
			miniAuth.GET("/follow/following/:id", miniapp.GetUserFollowingList) // 5.他人关注
			miniAuth.GET("/follow/followers/:id", miniapp.GetUserFollowerList)  // 6.他人粉丝
			miniAuth.GET("/follow/status/:id", miniapp.GetFollowStatus)         // 7.关系状态
			miniAuth.GET("/follow/counts", miniapp.GetMyFollowCounts)           // 8.我的总数
			miniAuth.GET("/follow/counts/:id", miniapp.GetUserFollowCounts)     // 9.他人总数
			miniAuth.DELETE("/follow/followers/:id", miniapp.RemoveFollower)    // 10.移除粉丝
			miniAuth.POST("/follow/block/:id", miniapp.BlockUser)               // 11.拉黑
			miniAuth.DELETE("/follow/block/:id", miniapp.UnblockUser)           // 12.解除拉黑
			miniAuth.GET("/follow/blacklist", miniapp.GetMyBlacklist)           // 13.我的黑名单
			miniAuth.GET("/follow/blocked/:id", miniapp.IsBlockedByUser)        // 14.校验被对方拉黑
		}

		// ==================== 后台管理（需管理员 JWT） ====================
		adminGroup := api.Group("/admin", middleware.AdminJWTAuth())
		{
			// ---------- 后台媒体/文件上传 ----------
			adminGroup.POST("/upload/single", common.UploadSingleAdminImage) // 单图（新增）
			adminGroup.POST("/upload/batch", common.UploadAdminImages)       // 批量上传(没用到)

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
	api.GET("/ws", common.WebSocketHandler)

	log.Println("Server running on :", config.AppConfig.ServerPort)
	r.Run(":" + config.AppConfig.ServerPort)
}
