package miniapp

import (
	"github.com/gin-gonic/gin"

	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// GetRelationOptions 返回可关联的行程/攻略/搭子列表（仅已发布，供清单/记账本编辑页选择关联）
// @Summary 关联目标选项（行程/攻略/搭子）
// @Security BearerAuth
// @Tags 小程序-通用
// @Success 200 {object} response.Response{data=object{trips=[]object{id=string,title=string},guides=[]object{id=string,title=string},partners=[]object{id=string,title=string}}}
// @Router /api/v1/relations [get]
func GetRelationOptions(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	// 返回全部行程/攻略/搭子（status/isDraft=-1 不筛草稿，用户要求"全部"）
	trips, _, err := repository.ListUserTrips(userID, 1, 100, -1)
	if err != nil {
		response.Fail(c, 500, "获取行程列表失败")
		return
	}
	guides, _, err := repository.ListMyGuides(userID, 1, 100, -1)
	if err != nil {
		response.Fail(c, 500, "获取攻略列表失败")
		return
	}
	partners, _, err := repository.GetMyPartners(userID, 1, 100, -1)
	if err != nil {
		response.Fail(c, 500, "获取搭子列表失败")
		return
	}

	tripItems := make([]gin.H, 0, len(trips))
	for _, t := range trips {
		tripItems = append(tripItems, gin.H{"id": t.ID, "title": t.Title})
	}
	guideItems := make([]gin.H, 0, len(guides))
	for _, g := range guides {
		guideItems = append(guideItems, gin.H{"id": g.ID, "title": g.Title})
	}
	partnerItems := make([]gin.H, 0, len(partners))
	for _, p := range partners {
		partnerItems = append(partnerItems, gin.H{"id": p.ID, "title": p.Title})
	}
	response.Success(c, gin.H{
		"trips":    tripItems,
		"guides":   guideItems,
		"partners": partnerItems,
	})
}
