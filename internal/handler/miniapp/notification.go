package miniapp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// MarkNotificationRead 标记单条通知为已读
// @Summary 标记已读
// @Security BearerAuth
// @Tags 小程序-通知
// @Param id path string true "通知ID"
// @Success 200 {object} response.Response
// @Router /api/v1/notification/read/{id} [put]
func MarkNotificationRead(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(string)
	if err := repository.MarkNotificationRead(id, userID); err != nil {
		response.Fail(c, 500, "操作失败")
		return
	}
	response.Success(c, nil)
}

// MarkAllNotificationsRead 标记所有通知为已读
// @Summary 全部已读
// @Security BearerAuth
// @Tags 小程序-通知
// @Success 200 {object} response.Response
// @Router /api/v1/notification/read-all [put]
func MarkAllNotificationsRead(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	if err := repository.MarkAllNotificationsRead(userID); err != nil {
		response.Fail(c, 500, "操作失败")
		return
	}
	response.Success(c, nil)
}

// GetNotificationList 分页获取通知列表（type=0 全部，1搭子申请，2点赞，3新增关注，4系统通知，5评论）
// @Summary 通知列表
// @Security BearerAuth
// @Tags 小程序-通知
// @Param type query int false "通知类型：0全部 1搭子申请 2点赞 3新增关注 4系统通知 5评论"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.Notification,total=int64}}
// @Router /api/v1/notification/list [get]
func GetNotificationList(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	notiType, _ := strconv.Atoi(c.DefaultQuery("type", "0"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	list, total, err := repository.ListNotifications(userID, notiType, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// GetUnreadNotificationCounts 获取所有类型的未读通知数量
// @Summary 未读通知数量
// @Security BearerAuth
// @Tags 小程序-通知
// @Success 200 {object} response.Response{data=object{partnerApplyCount=int64,commentCount=int64,likeCount=int64,followCount=int64,systemNotifyCount=int64}}
// @Router /api/v1/notification/unread [get]
func GetUnreadNotificationCounts(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	partnerApplyCount, likeCount, followCount, commentCount, systemNotifyCount, err := repository.GetUnreadCounts(userID)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{
		"partnerApplyCount": partnerApplyCount,
		"commentCount":      commentCount,
		"likeCount":         likeCount,
		"followCount":       followCount,
		"systemNotifyCount": systemNotifyCount,
	})
}
