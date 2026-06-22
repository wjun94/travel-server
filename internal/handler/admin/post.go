package admin

import (
	"strconv"

	_ "travel-server/internal/model" // swagger 类型引用需要
	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListPosts 后台攻略列表（含审核状态）
// @Summary 攻略列表
// @Security BearerAuth
// @Tags 后台-内容
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.Post,total=int}}
// @Router /api/v1/admin/posts [get]
func ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	posts, total, err := repository.ListPosts(page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取攻略列表失败")
		return
	}
	response.Success(c, gin.H{"list": posts, "total": total})
}

// UpdatePostStatus 审核攻略（发布/下架）
// @Summary 审核攻略
// @Security BearerAuth
// @Tags 后台-内容
// @Param id path int true "攻略ID"
// @Param body body object{status=int} true "状态(1已发布 2下架)"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/post/{id}/status [put]
func UpdatePostStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.Status != 1 && req.Status != 2 {
		response.Fail(c, 400, "状态值无效")
		return
	}
	if err := repository.UpdatePostStatus(uint(id), req.Status); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}
