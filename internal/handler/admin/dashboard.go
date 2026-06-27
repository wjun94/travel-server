package admin

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// Dashboard 管理端首页统计
// @Summary 仪表盘统计
// @Security BearerAuth
// @Tags 后台-仪表盘
// @Success 200 {object} response.Response{data=object{user_count=int,guide_count=int,partner_count=int}}
// @Router /api/v1/admin/dashboard [get]
func Dashboard(c *gin.Context) {
	var userCount, guideCount, partnerCount int64
	database.DB.Model(&model.User{}).Count(&userCount)
	database.DB.Model(&model.Guide{}).Count(&guideCount)
	database.DB.Model(&model.Partner{}).Count(&partnerCount)
	response.Success(c, gin.H{
		"userCount":    userCount,
		"guideCount":   guideCount,
		"partnerCount": partnerCount,
	})
}
