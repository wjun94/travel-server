package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// ListAdminUsers 获取后台用户列表
// @Summary 后台用户列表
// @Security BearerAuth
// @Tags 后台-用户
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.AdminUser,total=int}}
// @Router /api/v1/admin/users [get]
func ListAdminUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	users, total, err := repository.ListAdminUsers(page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": users, "total": total})
}

// CreateAdminUser 创建后台用户
// @Summary 创建后台用户
// @Security BearerAuth
// @Tags 后台-用户
// @Param body body object{username=string,password=string,role_id=int} true "用户信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/user [post]
func CreateAdminUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		RoleID   uint   `json:"roleId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := model.AdminUser{
		Username:     req.Username,
		PasswordHash: string(hash),
		RoleID:       req.RoleID,
		Status:       1,
	}
	if err := repository.CreateAdminUser(&user); err != nil {
		response.Fail(c, 500, "创建失败")
		return
	}
	response.Success(c, user)
}

// UpdateAdminUser 修改后台用户（角色、状态、密码）
// @Summary 修改后台用户
// @Security BearerAuth
// @Tags 后台-用户
// @Param id path int true "用户ID"
// @Param body body object{role_id=int,status=int,password=string} true "修改参数"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/user/{id} [put]
func UpdateAdminUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		RoleID   uint   `json:"roleId"`
		Status   *int   `json:"status"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	user, err := repository.GetAdminUserByID(uint(id))
	if err != nil {
		response.Fail(c, 404, "用户不存在")
		return
	}
	if req.RoleID != 0 {
		user.RoleID = req.RoleID
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user.PasswordHash = string(hash)
	}
	if err := repository.UpdateAdminUser(user); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// DeleteAdminUser 删除后台用户
// @Summary 删除后台用户
// @Security BearerAuth
// @Tags 后台-用户
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/user/{id} [delete]
func DeleteAdminUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := repository.DeleteAdminUser(uint(id)); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}
