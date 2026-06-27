package admin

import (
	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// SaveRecommendation 保存/更新推荐内容
// @Summary 保存推荐
// @Security BearerAuth
// @Tags 后台-推荐
// @Param body body model.Recommendation true "推荐内容"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/recommendation [post]
func SaveRecommendation(c *gin.Context) {
	var rec model.Recommendation
	if err := c.ShouldBindJSON(&rec); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	rec.ID = "" // 强制新建，由 BeforeCreate 生成 ID
	if err := repository.CreateRecommendation(&rec); err != nil {
		response.Fail(c, 500, "保存失败")
		return
	}
	response.Success(c, rec)
}

// ListRecommendations 获取推荐列表
// @Summary 推荐列表
// @Security BearerAuth
// @Tags 后台-推荐
// @Success 200 {object} response.Response{data=[]model.Recommendation}
// @Router /api/v1/admin/recommendations [get]
func ListRecommendations(c *gin.Context) {
	recs, _ := repository.GetRecommendations("")
	response.Success(c, recs)
}
