package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// ListRoles 获取所有角色（分页）
// @Summary 角色列表
// @Security BearerAuth
// @Tags 后台-角色
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.Role,total=int}}
// @Router /api/v1/admin/roles [get]
func ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	roles, total, err := repository.ListRoles(page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取角色列表失败")
		return
	}
	response.Success(c, gin.H{"list": roles, "total": total})
}

// CreateRole 创建角色
// @Summary 创建角色
// @Security BearerAuth
// @Tags 后台-角色
// @Param body body object{name=string,description=string,permissions=string} true "角色信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/role [post]
func CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Permissions string `json:"permissions"` // JSON string
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	role := model.Role{
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	}
	if err := repository.CreateRole(&role); err != nil {
		response.Fail(c, 500, "创建失败")
		return
	}
	response.Success(c, role)
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Security BearerAuth
// @Tags 后台-角色
// @Param id path int true "角色ID"
// @Param body body object{name=string,description=string,permissions=string} true "角色信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/role/{id} [put]
func UpdateRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Permissions string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	role, err := repository.GetRoleByID(uint(id))
	if err != nil {
		response.Fail(c, 404, "角色不存在")
		return
	}
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Permissions != "" {
		role.Permissions = req.Permissions
	}
	if err := repository.UpdateRole(role); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Security BearerAuth
// @Tags 后台-角色
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/role/{id} [delete]
func DeleteRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := repository.DeleteRole(uint(id)); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}
