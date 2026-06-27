package miniapp

import (
	"strconv"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// AddFavorite 添加收藏
// @Summary 添加收藏
// @Security BearerAuth
// @Tags 小程序-收藏
// @Param body body model.Favorite true "收藏信息"
// @Success 200 {object} response.Response
// @Router /api/v1/favorite [post]
func AddFavorite(c *gin.Context) {
	var fav model.Favorite
	if err := c.ShouldBindJSON(&fav); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	fav.UserID = c.GetUint("userID")
	if err := repository.AddFavorite(&fav); err != nil {
		response.Fail(c, 500, "收藏失败")
		return
	}
	response.Success(c, fav)
}

// RemoveFavorite 取消收藏
// @Summary 取消收藏
// @Security BearerAuth
// @Tags 小程序-收藏
// @Param id path int true "收藏ID"
// @Param target_type query string true "收藏类型"
// @Success 200 {object} response.Response
// @Router /api/v1/favorite/{id} [delete]
func RemoveFavorite(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	targetType := c.Query("target_type")
	userID := c.GetUint("userID")
	if err := repository.RemoveFavorite(userID, uint(id), targetType); err != nil {
		response.Fail(c, 500, "取消收藏失败")
		return
	}
	response.Success(c, nil)
}

// GetFavorites 用户收藏列表
// @Summary 收藏列表
// @Security BearerAuth
// @Tags 小程序-收藏
// @Param target_type query string false "收藏类型(guide/trip)"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/v1/favorites [get]
func GetFavorites(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	targetType := c.Query("target_type")
	userID := c.GetUint("userID")
	favs, total, err := repository.ListUserFavorites(userID, targetType, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": favs, "total": total})
}
