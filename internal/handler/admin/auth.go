package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"travel-server/internal/middleware"
	"travel-server/internal/repository"
	"travel-server/internal/service"
	"travel-server/pkg/response"
)

// AdminLogin 后台管理员登录
// @Summary 管理员登录
// @Description 使用用户名和密码登录后台管理系统，返回 Admin JWT Token 及基本用户信息
// @Tags 后台-认证
// @Accept json
// @Produce json
// @Param body body object{username=string,password=string} true "登录凭证"
// @Success 200 {object} response.Response{data=object{token=string,user=object{id=int,username=string,role=string}}}
// @Router /api/v1/admin/login [post]
func AdminLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}

	user, err := service.AdminLogin(req.Username, req.Password)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 使用 Admin 专用 token 生成
	token, err := middleware.GenerateAdminToken(user.ID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "生成token失败")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role.Name,
		},
	})
}

// GetAdminInfo 获取当前登录的管理员信息
// @Summary 获取管理员信息
// @Description 根据 Admin JWT Token 获取当前管理员的详细信息和角色权限
// @Security BearerAuth
// @Tags 后台-认证
// @Produce json
// @Success 200 {object} response.Response{data=object{id=int,username=string,role=object{name=string,permissions=string}}}
// @Router /api/v1/admin/info [get]
func GetAdminInfo(c *gin.Context) {
	adminID, _ := c.Get("adminUserID") // AdminOnly 中间件注入
	user, err := repository.GetAdminUserByID(adminID.(uint))
	if err != nil {
		response.Fail(c, 404, "用户不存在")
		return
	}
	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role": gin.H{
			"name":        user.Role.Name,
			"permissions": user.Role.Permissions,
		},
	})
}
