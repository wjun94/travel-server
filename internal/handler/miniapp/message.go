package miniapp

import (
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/database"
	"travel-server/pkg/response"
)

// MessageVO 消息视图对象（含发送者信息）
type MessageVO struct {
	ID         string    `json:"id"`
	FromUserID string    `json:"fromUserId"`
	ToUserID   string    `json:"toUserId"`
	Content    string    `json:"content"`
	Type       int       `json:"type"`
	IsRead     int       `json:"isRead"`
	CreatedAt  time.Time `json:"createdAt"`
	AvatarURL  string    `json:"avatarUrl"`
	Nickname   string    `json:"nickname"`
}

// GetMessageList 获取与指定用户的私聊记录
// @Summary 获取消息列表
// @Security BearerAuth
// @Tags 小程序-消息
// @Param targetUserId query string true "对方用户ID"
// @Success 200 {object} response.Response{data=[]MessageVO}
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

	// 标记对方发给当前用户的未读消息为已读
	_ = repository.MarkMessagesAsRead(targetID, uid)

	// 收集所有发送者ID
	userIDs := make([]string, 0)
	seen := make(map[string]bool)
	for _, msg := range msgs {
		if !seen[msg.FromUserID] {
			seen[msg.FromUserID] = true
			userIDs = append(userIDs, msg.FromUserID)
		}
	}

	// 批量查询用户信息
	var users []model.User
	database.DB.Select("id, nickname, avatar_url").Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[string]model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// 构造带用户信息的响应
	result := make([]MessageVO, 0, len(msgs))
	for _, msg := range msgs {
		u := userMap[msg.FromUserID]
		result = append(result, MessageVO{
			ID:         msg.ID,
			FromUserID: msg.FromUserID,
			ToUserID:   msg.ToUserID,
			Content:    msg.Content,
			Type:       msg.Type,
			IsRead:     msg.IsRead,
			CreatedAt:  msg.CreatedAt,
			AvatarURL:  u.AvatarURL,
			Nickname:   u.Nickname,
		})
	}

	response.Success(c, result)
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

	// 查询最新系统通知
	var latestNoti model.Notification
	database.DB.Where("user_id = ? AND type = 4", userID).
		Order("created_at desc").First(&latestNoti)
	var unreadSys int64
	database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND type = 4 AND is_read = 0", userID).Count(&unreadSys)

	// 在列表首部插入系统通知会话项（无通知记录时不显示）
	result := make([]repository.ConversationItem, 0, len(list)+1)
	if latestNoti.ID != "" {
		result = append(result, repository.ConversationItem{
			UserID:      "",
			Nickname:    "系统通知",
			AvatarURL:   "travel/img/notify.png",
			LastContent: latestNoti.Content,
			LastTime:    latestNoti.CreatedAt,
			UnreadCount: unreadSys,
		})
	}
	result = append(result, list...)

	response.Success(c, result)
}
