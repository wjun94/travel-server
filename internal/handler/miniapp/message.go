package miniapp

import (
	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// GetMessageList 获取与指定用户的私聊记录
// @Summary 获取消息列表
// @Security BearerAuth
// @Tags 小程序-消息
// @Param targetUserId query string true "对方用户ID"
// @Success 200 {object} response.Response{data=[]model.Message}
// @Router /api/v1/message/list [get]
func GetMessageList(c *gin.Context) {
	uid := c.MustGet("userID").(string)
	targetID := c.Query("targetUserId")
	if targetID == "" {
		response.Fail(c, 400, "缺少targetUserId")
		return
	}
	msgs, err := repository.GetMessagesBetweenUsers(uid, targetID)
	if err != nil {
		response.Fail(c, 500, "获取消息失败")
		return
	}
	response.Success(c, msgs)
}

// SendMessage 发送私聊消息
// @Summary 发送消息
// @Security BearerAuth
// @Tags 小程序-消息
// @Param body body object{toUserId=string,content=string} true "消息内容"
// @Success 200 {object} response.Response
// @Router /api/v1/message/send [post]
func SendMessage(c *gin.Context) {
	var req struct {
		ToUserID string `json:"toUserId" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	msg := model.Message{
		FromUserID: c.MustGet("userID").(string),
		ToUserID:   req.ToUserID,
		Content:    req.Content,
		Type:       1, // 私聊
	}
	if err := repository.CreateMessage(&msg); err != nil {
		response.Fail(c, 500, "发送失败")
		return
	}
	response.Success(c, msg)
}

// GetConversationList 获取会话列表（消息中心）
// @Summary 会话列表
// @Security BearerAuth
// @Tags 小程序-消息
// @Success 200 {object} response.Response{data=[]repository.ConversationItem}
// @Router /api/v1/message/conversations [get]
func GetConversationList(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	list, err := repository.GetConversationList(userID)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, list)
}
