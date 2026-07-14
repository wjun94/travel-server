package miniapp

import (
	"strconv"

	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// FollowUser 关注用户
// @Summary 关注用户
// @Security BearerAuth
// @Tags 小程序-关注
// @Param id path string true "被关注者ID"
// @Success 200 {object} response.Response
// @Router /api/v1/follow/{id} [post]
func FollowUser(c *gin.Context) {
	userID := c.Param("id")
	followerID := c.MustGet("userID").(string)
	if err := repository.FollowUser(userID, followerID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

// UnfollowUser 取消关注
// @Summary 取消关注
// @Security BearerAuth
// @Tags 小程序-关注
// @Param id path string true "被关注者ID"
// @Success 200 {object} response.Response
// @Router /api/v1/follow/{id} [delete]
func UnfollowUser(c *gin.Context) {
	userID := c.Param("id")
	followerID := c.MustGet("userID").(string)
	if err := repository.UnfollowUser(userID, followerID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

// GetMyFollowingList 我的关注列表
// @Summary 我的关注列表
// @Security BearerAuth
// @Tags 小程序-关注
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]repository.FollowItem,total=int}}
// @Router /api/v1/follow/following [get]
func GetMyFollowingList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	userID := c.MustGet("userID").(string)
	list, total, err := repository.GetMyFollowingList(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// GetMyFollowerList 我的粉丝列表
// @Summary 我的粉丝列表
// @Security BearerAuth
// @Tags 小程序-关注
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]repository.FollowItem,total=int}}
// @Router /api/v1/follow/followers [get]
func GetMyFollowerList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	userID := c.MustGet("userID").(string)
	list, total, err := repository.GetMyFollowerList(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// GetUserFollowingList 他人的关注列表
// @Summary 他人的关注列表
// @Security BearerAuth
// @Tags 小程序-关注
// @Param id path string true "用户ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]repository.FollowItem,total=int}}
// @Router /api/v1/follow/following/{id} [get]
func GetUserFollowingList(c *gin.Context) {
	targetUserID := c.Param("id")
	currentUserID := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := repository.GetUserFollowingList(targetUserID, currentUserID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// GetUserFollowerList 他人的粉丝列表
// @Summary 他人的粉丝列表
// @Security BearerAuth
// @Tags 小程序-关注
// @Param id path string true "用户ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]repository.FollowItem,total=int}}
// @Router /api/v1/follow/followers/{id} [get]
func GetUserFollowerList(c *gin.Context) {
	targetUserID := c.Param("id")
	currentUserID := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := repository.GetUserFollowerList(targetUserID, currentUserID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// GetFollowStatus 与指定用户的关系状态
// @Summary 与指定用户的关系状态
// @Security BearerAuth
// @Tags 小程序-关注
// @Param id path string true "目标用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/follow/status/{id} [get]
func GetFollowStatus(c *gin.Context) {
	targetUserID := c.Param("id")
	currentUserID := c.MustGet("userID").(string)
	status, err := repository.GetFollowStatus(currentUserID, targetUserID)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"status": status})
}

// GetMyFollowCounts 我的关注/粉丝总数
// @Summary 我的关注/粉丝总数
// @Security BearerAuth
// @Tags 小程序-关注
// @Success 200 {object} response.Response
// @Router /api/v1/follow/counts [get]
func GetMyFollowCounts(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	followCount, followerCount, err := repository.GetMyCounts(userID)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{
		"followCount":   followCount,
		"followerCount": followerCount,
	})
}

// GetUserFollowCounts 他人的关注/粉丝总数
// @Summary 他人的关注/粉丝总数
// @Security BearerAuth
// @Tags 小程序-关注
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/follow/counts/{id} [get]
func GetUserFollowCounts(c *gin.Context) {
	targetUserID := c.Param("id")
	currentUserID := c.MustGet("userID").(string)
	followCount, followerCount, err := repository.GetUserCounts(targetUserID, currentUserID)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{
		"followCount":   followCount,
		"followerCount": followerCount,
	})
}

// RemoveFollower 移除粉丝
// @Summary 移除粉丝
// @Security BearerAuth
// @Tags 小程序-关注
// @Param id path string true "粉丝用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/follow/followers/{id} [delete]
func RemoveFollower(c *gin.Context) {
	followerID := c.Param("id")
	userID := c.MustGet("userID").(string)
	if err := repository.RemoveFollower(userID, followerID); err != nil {
		response.Fail(c, 400, "移除失败")
		return
	}
	response.Success(c, nil)
}

// BlockUser 拉黑用户
// @Summary 拉黑用户
// @Security BearerAuth
// @Tags 小程序-关注
// @Param id path string true "被拉黑用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/follow/block/{id} [post]
func BlockUser(c *gin.Context) {
	blockedUserID := c.Param("id")
	userID := c.MustGet("userID").(string)
	if err := repository.BlockUser(userID, blockedUserID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

// UnblockUser 解除拉黑
// @Summary 解除拉黑
// @Security BearerAuth
// @Tags 小程序-关注
// @Param id path string true "被拉黑用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/follow/block/{id} [delete]
func UnblockUser(c *gin.Context) {
	blockedUserID := c.Param("id")
	userID := c.MustGet("userID").(string)
	if err := repository.UnblockUser(userID, blockedUserID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

// GetMyBlacklist 我的拉黑名单
// @Summary 我的拉黑名单
// @Security BearerAuth
// @Tags 小程序-关注
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]repository.BlacklistItem,total=int}}
// @Router /api/v1/follow/blacklist [get]
func GetMyBlacklist(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	userID := c.MustGet("userID").(string)
	list, total, err := repository.GetMyBlacklist(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// IsBlockedByUser 校验是否被对方拉黑
// @Summary 校验是否被对方拉黑
// @Security BearerAuth
// @Tags 小程序-关注
// @Param id path string true "目标用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/follow/blocked/{id} [get]
func IsBlockedByUser(c *gin.Context) {
	targetUserID := c.Param("id")
	currentUserID := c.MustGet("userID").(string)
	blocked, err := repository.IsBlockedByUser(currentUserID, targetUserID)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"blocked": blocked})
}
