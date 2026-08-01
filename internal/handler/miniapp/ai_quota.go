package miniapp

import (
	"github.com/gin-gonic/gin"

	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// GetAiQuota 查询今日AI调用额度
// @Summary AI调用额度
// @Description 每日基础1次，邀请好友成功1人可额外+1次；返回今日剩余次数
// @Security BearerAuth
// @Tags 小程序-AI
// @Success 200 {object} response.Response{data=object{trip=object{used=int,total=int,remain=int},partner=object{used=int,total=int,remain=int}}}
// @Router /api/v1/ai/quota [get]
func GetAiQuota(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	// 今日邀请成功奖励次数
	inviteCount, _ := repository.CountTodayInviteSuccess(userID)
	bonus := int(inviteCount)

	tripUsed, _ := repository.CountTodayAITrips(userID)
	tripTotal := 1 + bonus
	tripRemain := tripTotal - int(tripUsed)
	if tripRemain < 0 {
		tripRemain = 0
	}

	partnerUsed, _ := repository.CountTodayAIPartners(userID)
	partnerTotal := 1 + bonus
	partnerRemain := partnerTotal - int(partnerUsed)
	if partnerRemain < 0 {
		partnerRemain = 0
	}

	response.Success(c, gin.H{
		"trip": gin.H{
			"used":   tripUsed,
			"total":  tripTotal,
			"remain": tripRemain,
		},
		"partner": gin.H{
			"used":   partnerUsed,
			"total":  partnerTotal,
			"remain": partnerRemain,
		},
	})
}
