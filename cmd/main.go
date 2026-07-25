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
	"travel-server/internal/cron"
	"travel-server/internal/handler/admin"
	"travel-server/internal/handler/common"
	"travel-server/internal/handler/miniapp"
	"travel-server/internal/middleware"
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

	// 启动定时任务（浏览历史清理、过期搭子自动关闭等）
	cron.Start()

	// Swagger 文档面板
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		// ==================== 公开接口（无需登录） ====================
		api.POST("/user/login", miniapp.UserLogin)                 // 微信登录
		api.POST("/admin/login", admin.AdminLogin)                 // 后台登录
		api.GET("/guide/feed", miniapp.GetGuideFeed)               // 攻略瀑布流
		api.GET("/nearby", miniapp.GetNearby)                      // 周边推荐
		api.GET("/nearby/recommend", miniapp.GetTopRecommend)      // TOP推荐
		api.GET("/weather", common.GetWeather)                     // 天气查询
		api.GET("/weather/qweather", common.GetQWeather)           // 天气查询（和风）
		api.GET("/comments", miniapp.GetComments)                  // 评论列表
		api.GET("/comment/replies", miniapp.GetReplies)            // 子回复列表
		api.GET("/regions/all", miniapp.GetAllRegions)             // 国内省/市列表
		api.GET("/regions/countries", miniapp.GetCountries)        // 境外国家列表
		api.GET("/destinations/search", miniapp.SearchDestination) // 目的地搜索
		api.GET("/partner/list", miniapp.GetPartnerList)           // 搭子列表（公共）

		// ==================== 小程序端（需 JWT 登录） ====================
		miniAuth := api.Group("", middleware.JWTAuth())
		{
			// ---------- 媒体/文件上传 ----------
			miniAuth.POST("/upload/single", common.UploadSingleImage) // 单图（新增）
			miniAuth.DELETE("/image/delete", common.DeleteImage)      // 删除单照片
			miniAuth.POST("/upload/batch", common.UploadImages)       // 批量上传(没用到)

			// ---------- 用户 ----------
			miniAuth.GET("/profile", miniapp.GetMyProfile)                   // 个人主页（攻略数/行程数/粉丝/关注）
			miniAuth.GET("/profile/:id", miniapp.GetUserProfile)             // 他人个人主页
			miniAuth.GET("/profile/:id/favorites", miniapp.GetUserFavorites) // 他人收藏列表
			miniAuth.GET("/profile/:id/feed", miniapp.GetUserFeed)           // 他人的攻略+行程流（按时间）
			miniAuth.GET("/user/info", miniapp.GetUserInfo)                  // 个人信息
			miniAuth.PUT("/user/profile", miniapp.UpdateProfile)             // 更新个人资料

			// ---------- 攻略 ----------
			miniAuth.POST("/guide", miniapp.CreateGuide)                       // 创建攻略
			miniAuth.GET("/my/guides", miniapp.GetMyGuides)                    // 我的攻略列表
			miniAuth.GET("/guide/:id", miniapp.GetGuideDetail)                 // 攻略详情
			miniAuth.PUT("/guide/:id", miniapp.UpdateGuide)                    // 更新攻略
			miniAuth.POST("/guide/:id/day", miniapp.CreateGuideDay)            // 新增一天
			miniAuth.DELETE("/guide/day/:id", miniapp.DeleteGuideDay)          // 删除一天
			miniAuth.POST("/guide/:id/like", miniapp.LikeGuide)                // 点赞攻略
			miniAuth.DELETE("/guide/:id/like", miniapp.UnlikeGuide)            // 取消点赞
			miniAuth.POST("/guide/day/:id/item", miniapp.CreateGuideDayItem)   // 新增行程项
			miniAuth.PUT("/guide/day/item/:id", miniapp.UpdateGuideDayItem)    // 更新行程项
			miniAuth.DELETE("/guide/day/item/:id", miniapp.DeleteGuideDayItem) // 删除行程项

			// ---------- 行程 ----------
			miniAuth.POST("/trip", miniapp.CreateTrip)                 // 手动创建行程
			miniAuth.GET("/my/trips", miniapp.GetMyTrips)              // 我的行程列表
			miniAuth.POST("/trip/ai-generate", miniapp.AIGenerateTrip) // AI智能生成行程
			miniAuth.GET("/trip/:id", miniapp.GetTrip)                 // 行程详情
			miniAuth.PUT("/trip/:id", miniapp.UpdateTrip)              // 更新行程
			miniAuth.POST("/trip/day", miniapp.AddTripDay)             // 添加行程日
			miniAuth.PUT("/trip/day/:id", miniapp.UpdateTripDay)       // 更新行程日
			miniAuth.DELETE("/trip/day/:id", miniapp.DeleteTripDay)    // 删除行程日
			miniAuth.POST("/trip/item", miniapp.AddTripItem)           // 添加行程项
			miniAuth.PUT("/trip/item/:id", miniapp.UpdateTripItem)     // 更新行程项
			miniAuth.DELETE("/trip/item/:id", miniapp.DeleteTripItem)  // 删除行程项
			miniAuth.POST("/trip/member", miniapp.InviteMember)        // 邀请同行者
			miniAuth.DELETE("/trip/member/:id", miniapp.RemoveMember)  // 移除同行者

			// ---------- 搭子 ----------
			miniAuth.POST("/partner", miniapp.CreatePartner)                    // 发布搭子
			miniAuth.GET("/my/partners", miniapp.GetMyPartners)                 // 我的搭子列表
			miniAuth.GET("/partner/:id", miniapp.GetPartnerDetail)              // 搭子详情
			miniAuth.POST("/partner/:id/apply", miniapp.ApplyPartner)           // 申请加入
			miniAuth.PUT("/partner/:id/application", miniapp.HandleApplication) // 处理申请
			miniAuth.PUT("/partner/:id/cancel", miniapp.CancelPartner)          // 取消搭子

			// ---------- 消息 ----------
			miniAuth.GET("/message/list", miniapp.GetMessageList)               // 消息列表
			miniAuth.POST("/message/send", miniapp.SendMessage)                 // 发送消息
			miniAuth.GET("/message/conversations", miniapp.GetConversationList) // 消息中心聊天会话列表

			// ---------- 通知 ----------
			miniAuth.GET("/notification/unread", miniapp.GetUnreadNotificationCounts) // 未读通知数量
			miniAuth.GET("/notification/list", miniapp.GetNotificationList)           // 通知列表（按type筛选）
			miniAuth.PUT("/notification/read/:id", miniapp.MarkNotificationRead)      // 标记单条已读
			miniAuth.PUT("/notification/read-all", miniapp.MarkAllNotificationsRead)  // 全部已读

			// ---------- 记账 ----------
			miniAuth.GET("/account/:tripId", miniapp.GetAccounts)     // 查询行程账本
			miniAuth.POST("/account", miniapp.AddAccount)             // 添加一笔账
			miniAuth.POST("/account/import", miniapp.ImportWechatPay) // 导入微信账单

			// ---------- 备忘清单 ----------
			miniAuth.GET("/checklist", miniapp.GetChecklists)                     // 清单列表
			miniAuth.GET("/checklist/categories", miniapp.GetChecklistCategories) // 分类列表
			miniAuth.GET("/checklist/:id", miniapp.GetChecklistDetail)            // 清单详情
			miniAuth.POST("/checklist", miniapp.CreateChecklist)                  // 创建清单
			miniAuth.PUT("/checklist/:id", miniapp.UpdateChecklist)               // 更新清单
			miniAuth.DELETE("/checklist/:id", miniapp.DeleteChecklist)            // 删除清单
			miniAuth.PUT("/checklist/:id/item", miniapp.UpdateChecklistItem)      // 更新清单条目

			// ---------- 足迹 ----------
			miniAuth.GET("/footprint", miniapp.GetFootprints)         // 足迹列表
			miniAuth.POST("/footprint/sync", miniapp.SyncFootprint)   // 同步足迹
			miniAuth.GET("/footprint/poster", miniapp.GeneratePoster) // 生成足迹海报

			// ---------- 收藏 ----------
			miniAuth.POST("/favorite", miniapp.AddFavorite)           // 添加收藏
			miniAuth.POST("/favorite/remove", miniapp.RemoveFavorite) // 取消收藏
			miniAuth.GET("/favorites", miniapp.GetFavorites)          // 收藏列表

			// ---------- 评论 ----------
			miniAuth.POST("/comment", miniapp.CreateComment)        // 发表评论
			miniAuth.DELETE("/comment/:id", miniapp.DeleteComment)  // 删除评论
			miniAuth.POST("/comment/:id/like", miniapp.LikeComment) // 点赞评论

			// ---------- 浏览历史 ----------
			miniAuth.POST("/browse/history", miniapp.AddBrowseHistory)           // 新增浏览记录
			miniAuth.GET("/browse/history", miniapp.GetBrowseHistory)            // 浏览历史列表
			miniAuth.DELETE("/browse/history/clear", miniapp.ClearBrowseHistory) // 清空历史
			miniAuth.DELETE("/browse/history/:id", miniapp.DeleteBrowseHistory)  // 删除单条

			// ---------- 关注 ----------
			miniAuth.POST("/follow/:id", miniapp.FollowUser)                    // 关注
			miniAuth.DELETE("/follow/:id", miniapp.UnfollowUser)                // 取消关注
			miniAuth.GET("/follow/following", miniapp.GetMyFollowingList)       // 我的关注
			miniAuth.GET("/follow/followers", miniapp.GetMyFollowerList)        // 我的粉丝
			miniAuth.GET("/follow/following/:id", miniapp.GetUserFollowingList) // 他人关注
			miniAuth.GET("/follow/followers/:id", miniapp.GetUserFollowerList)  // 他人粉丝
			miniAuth.GET("/follow/status/:id", miniapp.GetFollowStatus)         // 关系状态
			miniAuth.GET("/follow/counts", miniapp.GetMyFollowCounts)           // 我的总数
			miniAuth.GET("/follow/counts/:id", miniapp.GetUserFollowCounts)     // 他人总数
			miniAuth.DELETE("/follow/followers/:id", miniapp.RemoveFollower)    // 移除粉丝
			miniAuth.POST("/follow/block/:id", miniapp.BlockUser)               // 拉黑
			miniAuth.DELETE("/follow/block/:id", miniapp.UnblockUser)           // 解除拉黑
			miniAuth.GET("/follow/blacklist", miniapp.GetMyBlacklist)           // 我的黑名单
			miniAuth.GET("/follow/blocked/:id", miniapp.IsBlockedByUser)        // 校验被对方拉黑
		}

		// ==================== 后台管理（需管理员 JWT） ====================
		adminGroup := api.Group("/admin", middleware.AdminJWTAuth())
		{
			// ---------- 后台媒体/文件上传 ----------
			adminGroup.POST("/upload/single", common.UploadSingleAdminImage) // 单图（新增）
			adminGroup.POST("/upload/batch", common.UploadAdminImages)       // 批量上传(没用到)

			// ---------- 认证 ----------
			adminGroup.GET("/info", admin.GetAdminInfo) // 管理员信息

			// ---------- 仪表盘 ----------
			adminGroup.GET("/dashboard", admin.Dashboard) // 仪表盘

			// ---------- 后台管理员 ----------
			adminGroup.GET("/admin/users", admin.ListAdminUsers)        // 管理员列表
			adminGroup.POST("/admin/user", admin.CreateAdminUser)       // 创建管理员
			adminGroup.PUT("/admin/user/:id", admin.UpdateAdminUser)    // 更新管理员
			adminGroup.DELETE("/admin/user/:id", admin.DeleteAdminUser) // 删除管理员

			// ---------- 角色权限 ----------
			adminGroup.GET("/roles", admin.ListRoles)        // 角色列表
			adminGroup.POST("/role", admin.CreateRole)       // 创建角色
			adminGroup.PUT("/role/:id", admin.UpdateRole)    // 更新角色
			adminGroup.DELETE("/role/:id", admin.DeleteRole) // 删除角色

			// ---------- 小程序用户 ----------
			adminGroup.GET("/users", admin.ListUsers)              // 用户列表
			adminGroup.PUT("/user/:id/role", admin.UpdateUserRole) // 修改用户角色

			// ---------- 攻略内容 ----------
			adminGroup.GET("/guides", admin.ListGuides)                  // 攻略列表
			adminGroup.PUT("/guide/:id/status", admin.UpdateGuideStatus) // 更新攻略状态

			// ---------- 官方搭子 ----------
			adminGroup.POST("/partner", admin.CreatePartner) // 创建官方搭子
			adminGroup.GET("/partners", admin.ListPartners)  // 官方搭子列表

			// ---------- 推荐内容 ----------
			adminGroup.POST("/recommendation", admin.SaveRecommendation)  // 保存推荐
			adminGroup.GET("/recommendations", admin.ListRecommendations) // 推荐列表
		}
	}

	// WebSocket 协同编辑
	api.GET("/ws", common.WebSocketHandler)

	log.Println("Server running on :", config.AppConfig.ServerPort)
	r.Run(":" + config.AppConfig.ServerPort)
}
