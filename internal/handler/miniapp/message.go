package miniapp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// GetMessageList 获取与指定用户的私聊记录
// @Summary 获取消息列表
// @Security BearerAuth
// @Tags 小程序-消息
// @Param target_user_id query int true "对方用户ID"
// @Success 200 {object} response.Response{data=[]model.Message}
// @Router /api/v1/message/list [get]
func GetMessageList(c *gin.Context) {
	uid := c.GetUint("userID")
	targetIDStr := c.Query("targetUserId")
	if targetIDStr == "" {
		response.Fail(c, 400, "缺少targetUserId")
		return
	}
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	msgs, err := repository.GetMessagesBetweenUsers(uid, uint(targetID))
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
// @Param body body object{to_user_id=int,content=string} true "消息内容"
// @Success 200 {object} response.Response
// @Router /api/v1/message/send [post]
func SendMessage(c *gin.Context) {
	var req struct {
		ToUserID uint   `json:"toUserId"`
		Content  string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	msg := model.Message{
		FromUserID: c.GetUint("userID"),
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
