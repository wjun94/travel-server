package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// CreatePartner 创建官方搭子团（后台发布）
// @Summary 创建官方搭子团
// @Security BearerAuth
// @Tags 后台-搭子
// @Param body body model.Partner true "官方搭子信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/partner [post]
func CreatePartner(c *gin.Context) {
	var p model.Partner
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	p.Type = 1    // 官方活动
	p.UserID = "" // 系统发布
	if err := repository.CreatePartner(&p); err != nil {
		response.Fail(c, 500, "创建失败")
		return
	}
	response.Success(c, p)
}

// ListPartners 获取官方搭子团列表（分页，支持目的地/状态筛选）
// @Summary 官方搭子团列表
// @Description 获取所有官方创建的搭子团，支持分页及目的地、状态筛选
// @Security BearerAuth
// @Tags 后台-搭子
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param destination query string false "目的地关键词"
// @Param status query int false "状态(0招募中 1满员 2取消 3已过期，-1或不传为全部)" default(-1)
// @Success 200 {object} response.Response{data=object{list=[]model.Partner,total=int}}
// @Router /api/v1/admin/partners [get]
func ListPartners(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	destination := c.Query("destination")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	partners, total, err := repository.GetPartners(page, pageSize, destination, status)
	if err != nil {
		response.Fail(c, 500, "获取官方搭子列表失败")
		return
	}
	response.Success(c, gin.H{
		"list":  partners,
		"total": total,
	})
}
