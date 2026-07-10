package miniapp

import (
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
