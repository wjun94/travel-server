package admin

import (
	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// CreateOfficialPartner 创建官方搭子团（后台发布）
// @Summary 创建官方搭子团
// @Security BearerAuth
// @Tags 后台-搭子
// @Param body body model.Partner true "官方搭子信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/official-partner [post]
func CreateOfficialPartner(c *gin.Context) {
	var p model.Partner
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	p.Type = 1   // 官方活动
	p.UserID = 0 // 系统发布
	if err := repository.CreatePartner(&p); err != nil {
		response.Fail(c, 500, "创建失败")
		return
	}
	response.Success(c, p)
}
