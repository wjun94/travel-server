package miniapp

import (
	"strconv"

	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetMyProfile 我的个人主页
// @Summary 我的个人主页
// @Security BearerAuth
// @Tags 小程序-用户
// @Success 200 {object} response.Response{data=repository.UserProfileStats}
// @Router /api/v1/profile [get]
func GetMyProfile(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	stats, err := repository.GetUserProfileStats(userID)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, stats)
}

// GetUserProfile 他人个人主页
// @Summary 他人个人主页
// @Security BearerAuth
// @Tags 小程序-用户
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response{data=repository.UserPublicProfile}
// @Router /api/v1/profile/{id} [get]
func GetUserProfile(c *gin.Context) {
	userID := c.Param("id")
	currentUserID := c.MustGet("userID").(string)

	profile, err := repository.GetUserPublicProfile(userID, currentUserID)
	if err != nil {
		response.Fail(c, 404, "用户不存在")
		return
	}
	response.Success(c, profile)
}

// GetUserFavorites 他人的收藏列表
// @Summary 他人的收藏列表
// @Security BearerAuth
// @Tags 小程序-用户
// @Param id path string true "用户ID"
// @Param target_type query string false "收藏类型(guide/trip)"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/v1/profile/{id}/favorites [get]
func GetUserFavorites(c *gin.Context) {
	userID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	targetType := c.Query("target_type")
	favs, total, err := repository.ListUserFavorites(userID, targetType, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": favs, "total": total})
}

// GetUserGuides 他人已发布的攻略列表
// @Summary 他人攻略列表
// @Security BearerAuth
// @Tags 小程序-用户
// @Param id path string true "用户ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/v1/profile/{id}/guides [get]
func GetUserGuides(c *gin.Context) {
	userID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	guides, total, err := repository.ListUserPublishedGuides(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": guides, "total": total})
}

// GetUserTrips 他人已公开的行程列表
// @Summary 他人行程列表
// @Security BearerAuth
// @Tags 小程序-用户
// @Param id path string true "用户ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/v1/profile/{id}/trips [get]
func GetUserTrips(c *gin.Context) {
	userID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	trips, total, err := repository.ListUserPublishedTrips(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": trips, "total": total})
}

// GetUserFeed 他人的公开内容流（攻略+行程合并，按时间倒序）
// @Summary 他人的公开内容流
// @Security BearerAuth
// @Tags 小程序-用户
// @Param id path string true "用户ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.FeedItem,total=int64}}
// @Router /api/v1/profile/{id}/feed [get]
func GetUserFeed(c *gin.Context) {
	userID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	items, total, err := repository.GetUserFeed(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": items, "total": total})
}
