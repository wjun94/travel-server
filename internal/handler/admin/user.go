package admin

import (
	"strconv"

	_ "travel-server/internal/model" // swagger 类型引用需要
	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListUsers 用户列表（分页，支持昵称/手机号关键词筛选）
// @Summary 用户列表
// @Security BearerAuth
// @Tags 后台-用户
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "昵称/手机号关键词"
// @Success 200 {object} response.Response{data=object{list=[]model.User,total=int}}
// @Router /api/v1/admin/users [get]
func ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	users, total, err := repository.ListUsers(page, pageSize, keyword)
	if err != nil {
		response.Fail(c, 500, "获取用户列表失败")
		return
	}
	response.Success(c, gin.H{"list": users, "total": total})
}

// UpdateUserRole 修改用户角色（普通/领队/管理员）
// @Summary 更新用户角色
// @Security BearerAuth
// @Tags 后台-用户
// @Param id path string true "用户ID"
// @Param body body object{role=int} true "角色(0普通 1领队 2管理员)"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/user/{id}/role [put]
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Role int `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.Role < 0 || req.Role > 2 {
		response.Fail(c, 400, "无效角色值")
		return
	}
	if err := repository.UpdateUserRole(id, req.Role); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}
