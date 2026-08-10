package miniapp

import (
	"strconv"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/internal/ws"
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
	var req struct {
		TargetType string `json:"targetType" binding:"required"`
		TargetID   string `json:"targetId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	userID := c.MustGet("userID").(string)
	// 已收藏则直接返回成功（幂等）
	if repository.IsFavorited(userID, req.TargetID, req.TargetType) {
		response.Success(c, nil)
		return
	}
	fav := model.Favorite{
		UserID:     userID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
	}
	if err := repository.AddFavorite(&fav); err != nil {
		response.Fail(c, 500, "收藏失败")
		return
	}
	// 通知内容作者（非本人收藏才通知）+ 实时推送，type=2 与点赞同属"点赞收藏"分类
	var authorID, content string
	switch req.TargetType {
	case "guide":
		if g, err := repository.GetGuideByID(req.TargetID); err == nil && g != nil {
			authorID = g.UserID
			content = "您的攻略收到一个收藏"
		}
	case "trip":
		if t, err := repository.GetTripByID(req.TargetID); err == nil && t != nil {
			authorID = t.UserID
			content = "您的行程收到一个收藏"
		}
	case "partner":
		if p, err := repository.GetPartnerByID(req.TargetID); err == nil && p != nil {
			authorID = p.UserID
			content = "您的搭子收到一个收藏"
		}
	}
	if authorID != "" && authorID != userID {
		notification := model.Notification{
			UserID:     authorID,
			FromUserID: userID,
			Type:       2,
			RelatedID:  req.TargetID,
			Content:    content,
		}
		if err := repository.CreateNotification(&notification); err == nil {
			ws.WsHub.PushToUser(authorID, map[string]interface{}{
				"action":  "new_notification",
				"type":    2, // 点赞收藏
				"content": notification.Content,
			})
		}
	}
	response.Success(c, fav)
}

// RemoveFavorite 取消收藏
// @Summary 取消收藏
// @Security BearerAuth
// @Tags 小程序-收藏
// @Param body body object{targetId=string,targetType=string} true "收藏目标信息"
// @Success 200 {object} response.Response
// @Router /api/v1/favorite/remove [post]
func RemoveFavorite(c *gin.Context) {
	var req struct {
		TargetType string `json:"targetType" binding:"required"`
		ID         string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	userID := c.MustGet("userID").(string)
	if err := repository.RemoveFavorite(userID, req.ID, req.TargetType); err != nil {
		response.Fail(c, 500, "取消失败")
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
	userID := c.MustGet("userID").(string)
	favs, total, err := repository.ListUserFavorites(userID, targetType, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": favs, "total": total})
}
