/*
 * 旅行搭子小程序后端服务入口
 * 初始化配置 → 数据库 → 注册路由 → 启动 HTTP 服务
 */
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	"travel-server/pkg/response"
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
	qiniu.InitQiniu()                 // 新增七牛云
	miniapp.BackfillGuideIsOverseas() // 存量攻略国内外标记回填（幂等）

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
		api.POST("/user/login", miniapp.UserLogin)                                     // 微信登录
		api.POST("/admin/login", admin.AdminLogin)                                     // 后台登录
		api.GET("/guide/feed", miniapp.GetGuideFeed)                                   // 攻略瀑布流
		api.GET("/nearby", miniapp.GetNearby)                                          // 周边推荐
		api.GET("/nearby/recommend", miniapp.GetTopRecommend)                          // TOP推荐
		api.GET("/weather", common.GetWeather)                                         // 天气查询
		api.GET("/weather/qweather", common.GetQWeather)                               // 天气查询（和风）
		api.GET("/comments", miniapp.GetComments)                                      // 评论列表
		api.GET("/comment/replies", miniapp.GetReplies)                                // 子回复列表
		api.GET("/regions/all", miniapp.GetAllRegions)                                 // 国内省/市列表
		api.GET("/regions/countries", miniapp.GetCountries)                            // 境外国家列表
		api.GET("/destinations/search", miniapp.SearchDestination)                     // 目的地搜索
		api.GET("/partner/list", middleware.OptionalJWTAuth(), miniapp.GetPartnerList) // 搭子列表（公共，可选JWT）

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
			miniAuth.GET("/user/info", miniapp.GetUserInfo)                  // 个人信息（含邀请码）
			miniAuth.PUT("/user/profile", miniapp.UpdateProfile)             // 更新个人资料
			miniAuth.POST("/bind/phone", miniapp.BindPhone)                  // 绑定微信手机号

			// ---------- AI ----------
			miniAuth.GET("/ai/quota", miniapp.GetAiQuota) // AI调用额度（行程/搭子今日剩余次数）

			// ---------- 攻略 ----------
			miniAuth.POST("/guide", miniapp.CreateGuide)                       // 创建攻略
			miniAuth.GET("/my/guides", miniapp.GetMyGuides)                    // 我的攻略列表（status可筛草稿）
			miniAuth.GET("/guide/:id", miniapp.GetGuideDetail)                 // 攻略详情
			miniAuth.PUT("/guide/:id", miniapp.UpdateGuide)                    // 更新攻略
			miniAuth.DELETE("/guide/:id", miniapp.DeleteGuide)                 // 删除攻略（仅作者）
			miniAuth.POST("/guide/:id/day", miniapp.CreateGuideDay)            // 新增一天
			miniAuth.DELETE("/guide/day/:id", miniapp.DeleteGuideDay)          // 删除一天
			miniAuth.POST("/guide/:id/like", miniapp.LikeGuide)                // 点赞攻略
			miniAuth.DELETE("/guide/:id/like", miniapp.UnlikeGuide)            // 取消点赞
			miniAuth.POST("/guide/day/:id/item", miniapp.CreateGuideDayItem)   // 新增行程项
			miniAuth.PUT("/guide/day/item/:id", miniapp.UpdateGuideDayItem)    // 更新行程项
			miniAuth.DELETE("/guide/day/item/:id", miniapp.DeleteGuideDayItem) // 删除行程项

			// ---------- 行程 ----------
			miniAuth.POST("/trip", miniapp.CreateTrip)                 // 手动创建行程
			miniAuth.GET("/my/trips", miniapp.GetMyTrips)              // 我的行程列表（status可筛草稿）
			miniAuth.POST("/trip/ai-generate", miniapp.AIGenerateTrip) // AI智能生成行程
			miniAuth.GET("/trip/:id", miniapp.GetTrip)                 // 行程详情
			miniAuth.PUT("/trip/:id", miniapp.UpdateTrip)              // 更新行程
			miniAuth.DELETE("/trip/:id", miniapp.DeleteTrip)           // 删除行程（仅作者）
			miniAuth.POST("/trip/day", miniapp.AddTripDay)             // 添加行程日
			miniAuth.DELETE("/trip/day/:id", miniapp.DeleteTripDay)    // 删除行程日
			miniAuth.POST("/trip/item", miniapp.AddTripItem)           // 添加行程项
			miniAuth.PUT("/trip/item/:id", miniapp.UpdateTripItem)     // 更新行程项
			miniAuth.DELETE("/trip/item/:id", miniapp.DeleteTripItem)  // 删除行程项
			miniAuth.POST("/trip/member", miniapp.InviteMember)        // 邀请同行者
			miniAuth.DELETE("/trip/member/:id", miniapp.RemoveMember)  // 移除同行者

			// ---------- 搭子 ----------
			miniAuth.POST("/partner", miniapp.CreatePartner)                    // 发布搭子
			miniAuth.POST("/partner/ai-generate", miniapp.AIGeneratePartner)    // AI智能生成搭子
			miniAuth.GET("/my/partners", miniapp.GetMyPartners)                 // 我的搭子列表（isDraft可筛草稿）
			miniAuth.GET("/my/joined-partners", miniapp.GetMyJoinedPartners)    // 我参与的搭子列表
			miniAuth.GET("/partner/:id", miniapp.GetPartnerDetail)              // 搭子详情
			miniAuth.PUT("/partner/:id", miniapp.UpdatePartner)                 // 更新搭子（编辑草稿）
			miniAuth.DELETE("/partner/:id", miniapp.DeletePartner)              // 删除搭子（仅作者）
			miniAuth.POST("/partner/:id/apply", miniapp.ApplyPartner)           // 申请加入
			miniAuth.PUT("/partner/:id/application", miniapp.HandleApplication) // 处理申请
			miniAuth.PUT("/partner/:id/cancel", miniapp.CancelPartner)          // 解散搭子（发起人）
			miniAuth.PUT("/partner/:id/leave", miniapp.LeavePartner)            // 退出搭子（成员）
			miniAuth.POST("/partner/:id/like", miniapp.LikePartner)             // 点赞搭子
			miniAuth.DELETE("/partner/:id/like", miniapp.UnlikePartner)         // 取消点赞搭子

			// ---------- 群聊 ----------
			miniAuth.GET("/conversation/list", miniapp.GetMyConversations)         // 我的群聊列表
			miniAuth.GET("/conversation/:id", miniapp.GetConversationDetail)       // 群聊详情（含成员）
			miniAuth.GET("/conversation/:id/messages", miniapp.GetGroupMessages)   // 群聊消息列表
			miniAuth.POST("/conversation/:id/message", miniapp.SendGroupMessage)   // 发送群聊消息
			miniAuth.PUT("/conversation/:id/kick", miniapp.KickConversationMember) // 踢出群成员（群主）

			// ---------- 消息 ----------
			miniAuth.GET("/message/list", miniapp.GetMessageList)              // 消息列表
			miniAuth.GET("/message/conversations", miniapp.GetMyConversations) // 会话列表（兼容旧路径）
			miniAuth.POST("/message/send", miniapp.SendMessage)                // 发送消息
			miniAuth.POST("/message/clear", miniapp.ClearChatHistory)          // 清空聊天记录（会话保留）
			miniAuth.DELETE("/message/session", miniapp.DeleteChatSession)     // 删除会话（列表不显示）

			// ---------- 通知 ----------
			miniAuth.GET("/notification/unread", miniapp.GetUnreadNotificationCounts) // 未读通知数量
			miniAuth.GET("/notification/list", miniapp.GetNotificationList)
			miniAuth.GET("/notification/:id", miniapp.GetNotificationDetail)          // 通知详情           // 通知列表（按type筛选）
			miniAuth.PUT("/notification/read/:id", miniapp.MarkNotificationRead)      // 标记单条已读
			miniAuth.PUT("/notification/type-read", miniapp.MarkTypeNotificationRead) // 按类型清空未读
			miniAuth.PUT("/notification/read-all", miniapp.MarkAllNotificationsRead)  // 全部已读
			miniAuth.DELETE("/notification/system", miniapp.ClearSystemNotifications) // 清空系统通知

			// ---------- 记账 ----------
			miniAuth.GET("/account/list", miniapp.GetAccountList)         // 账本明细（按目标）
			miniAuth.POST("/account", miniapp.AddAccount)                 // 添加一笔账
			miniAuth.POST("/account/book", miniapp.CreateAccountBook)     // 创建自主账本
			miniAuth.DELETE("/account/book", miniapp.DeleteAccountBook)   // 删除整本账本
			miniAuth.PUT("/account/:id", miniapp.UpdateAccount)           // 编辑一笔账
			miniAuth.DELETE("/account/:id", miniapp.DeleteAccount)        // 删除一笔账
			miniAuth.GET("/account/summary", miniapp.GetAccountSummary)   // 账本汇总
			miniAuth.GET("/account/overview", miniapp.GetAccountOverview) // 我的账本总览
			miniAuth.POST("/account/import", miniapp.ImportWechatPay)     // 导入微信账单

			// ---------- 投诉 ----------
			miniAuth.POST("/complaint", miniapp.SubmitComplaint)       // 提交投诉
			miniAuth.GET("/complaint/list", miniapp.ListComplaints)    // 我的投诉列表
			miniAuth.GET("/complaint/:id", miniapp.GetComplaintDetail) // 我的投诉详情

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
			adminGroup.GET("/guide/:id", admin.GetGuideDetail)           // 攻略详情

			// ---------- 行程审核 ----------
			adminGroup.GET("/trips", admin.ListTrips)                  // 行程列表
			adminGroup.PUT("/trip/:id/status", admin.UpdateTripStatus) // 审核行程（发布/下架/归档）
			adminGroup.GET("/trip/:id", admin.GetTripDetail)           // 行程详情

			// ---------- 搭子管理 ----------
			adminGroup.POST("/partner", admin.CreatePartner)                 // 创建官方搭子
			adminGroup.GET("/partners", admin.ListPartners)                  // 搭子列表
			adminGroup.PUT("/partner/:id/status", admin.UpdatePartnerStatus) // 审核搭子（下架/恢复）
			adminGroup.GET("/partner/:id", admin.GetPartnerDetail)           // 搭子详情

			// ---------- 推荐内容 ----------
			adminGroup.POST("/recommendation", admin.SaveRecommendation)   // 保存推荐
			adminGroup.GET("/recommendations", admin.ListRecommendations)  // 推荐列表
			adminGroup.GET("/complaints", admin.ListComplaints)            // 投诉列表
			adminGroup.PUT("/complaint/:id/status", admin.HandleComplaint) // 处理投诉
			adminGroup.DELETE("/complaint/:id", admin.DeleteComplaint)     // 删除投诉

			// ---------- 消息管理（系统通知） ----------
			adminGroup.POST("/sys-message", admin.CreateSysMessage)           // 发送系统消息（立即/定时）
			adminGroup.GET("/sys-messages", admin.ListSysMessages)            // 系统消息列表
			adminGroup.PUT("/sys-message/:id/cancel", admin.CancelSysMessage) // 取消定时发送
		}
	}

	// WebSocket 协同编辑
	api.GET("/ws", common.WebSocketHandler)

	// 静态文件服务（需放在 API 路由之后，避免覆盖接口路由）
	r.Static("/static", "./static") // 通过 /static/xxx 访问

	// 未匹配路由时，尝试从 static 目录直接读取文件（如 /kb406P2iUq.txt），否则返回 404
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.Contains(path, "..") {
			response.Fail(c, http.StatusNotFound, "页面不存在")
			return
		}
		fullPath := filepath.Join("./static", filepath.Clean(path))
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			c.File(fullPath)
			return
		}
		response.Fail(c, http.StatusNotFound, "页面不存在")
	})

	log.Println("Server running on :", config.AppConfig.ServerPort)
	r.Run(":" + config.AppConfig.ServerPort)
}
