package repository

import (
	"errors"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// GetConversationByPartnerID 根据搭子ID查询群聊
func GetConversationByPartnerID(partnerID string) (*model.Conversation, error) {
	var conv model.Conversation
	err := database.DB.Where("partner_id = ?", partnerID).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// GetConversationByID 根据ID查询群聊
func GetConversationByID(id string) (*model.Conversation, error) {
	var conv model.Conversation
	err := database.DB.First(&conv, "id = ?", id).Error
	return &conv, err
}

// CreateConversation 创建群聊
func CreateConversation(conv *model.Conversation) error {
	return database.DB.Create(conv).Error
}

// AddConversationMember 添加群成员（已存在则忽略，幂等）
func AddConversationMember(convID, userID string) error {
	var cnt int64
	database.DB.Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).Count(&cnt)
	if cnt > 0 {
		return nil
	}
	return database.DB.Create(&model.ConversationMember{
		ConversationID: convID,
		UserID:         userID,
		JoinedAt:       time.Now(),
	}).Error
}

// IsConversationMember 判断用户是否群成员（含群主，不含被踢成员）
func IsConversationMember(convID, userID string) bool {
	var cnt int64
	database.DB.Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).Count(&cnt)
	return cnt > 0
}

// ConversationMemberVO 群成员视图（含用户信息）
type ConversationMemberVO struct {
	UserID    string    `json:"userId"`    // 成员用户ID
	Nickname  string    `json:"nickname"`  // 昵称
	AvatarURL string    `json:"avatarUrl"` // 头像
	JoinedAt  time.Time `json:"joinedAt"`  // 加入时间
}

// GetConversationMembers 获取群成员列表（按加入时间正序，不含被踢成员）
func GetConversationMembers(convID string) ([]ConversationMemberVO, error) {
	var list []ConversationMemberVO
	err := database.DB.Table("conversation_members cm").
		Select("cm.user_id, u.nickname, u.avatar_url, cm.joined_at").
		Joins("LEFT JOIN users u ON u.id = cm.user_id").
		Where("cm.conversation_id = ? AND cm.deleted_at IS NULL", convID).
		Order("cm.joined_at asc").
		Scan(&list).Error
	return list, err
}

// KickConversationMember 踢出群成员（软删成员记录）
func KickConversationMember(convID, userID string) error {
	return database.DB.Where("conversation_id = ? AND user_id = ?", convID, userID).
		Delete(&model.ConversationMember{}).Error
}

// DissolveConversation 解散群聊（软删群聊与全部成员，解散后所有人均无法再进入/发言；仅群主可调用）
func DissolveConversation(convID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", convID).Delete(&model.Conversation{}).Error; err != nil {
			return err
		}
		return tx.Where("conversation_id = ?", convID).Delete(&model.ConversationMember{}).Error
	})
}

// ChatItemVO 统一会话列表项（系统消息/私聊/群聊）
type ChatItemVO struct {
	ID          string    `json:"id"`          // 会话ID：私聊=对方用户ID，群聊=群聊ID，系统=system
	Type        string    `json:"type"`        // user私聊 group群聊 system系统消息
	Name        string    `json:"name"`        // 显示名：昵称/群名/公告
	AvatarURL   string    `json:"avatarUrl"`   // 头像（私聊）
	PartnerID   string    `json:"partnerId"`   // 群聊关联搭子ID
	MemberCount int64     `json:"memberCount"` // 群聊成员数
	LastContent string    `json:"lastContent"` // 最后一条消息内容
	LastTime    time.Time `json:"lastTime"`    // 最后消息时间
	UnreadCount int64     `json:"unreadCount"` // 未读数
}

// GetChatList 获取统一会话列表（系统消息+私聊+群聊，按最后消息时间倒序，无消息的排最后）
func GetChatList(userID string) ([]ChatItemVO, error) {
	items := make([]ChatItemVO, 0, 16)

	// 1. 私聊会话：基于 chat_sessions 表（清空后会话保留，删除后列表消失）
	var sessions []model.ChatSession
	database.DB.Where("user_id = ?", userID).Find(&sessions)

	// 兼容老数据：从消息表聚合出所有会话对方，缺失的自动补建会话
	type msgRow struct {
		FromUserID string
		ToUserID   string
		Content    string
		CreatedAt  time.Time
	}
	var msgs []msgRow
	database.DB.Model(&model.Message{}).
		Select("from_user_id, to_user_id, content, created_at").
		Where("type = 1 AND ((from_user_id = ? AND deleted_by_sender = 0) OR (to_user_id = ? AND deleted_by_receiver = 0))", userID, userID).
		Order("created_at desc").
		Find(&msgs)

	peerSet := make(map[string]bool)
	for _, s := range sessions {
		peerSet[s.PeerID] = true
	}
	for _, m := range msgs {
		otherID := m.ToUserID
		if m.ToUserID == userID {
			otherID = m.FromUserID
		}
		if !peerSet[otherID] {
			peerSet[otherID] = true
			_ = UpsertChatSession(userID, otherID) // 自动补建，失败忽略下次再补
		}
	}

	if len(peerSet) > 0 {
		// 批量查对方用户信息
		otherIDs := make([]string, 0, len(peerSet))
		for oid := range peerSet {
			otherIDs = append(otherIDs, oid)
		}
		var users []model.User
		database.DB.Select("id, nickname, avatar_url").Where("id IN ?", otherIDs).Find(&users)
		userMap := make(map[string]model.User)
		for _, u := range users {
			userMap[u.ID] = u
		}
		for _, oid := range otherIDs {
			u := userMap[oid]
			// 最后一条消息（未被当前用户删除）
			var last msgRow
			database.DB.Model(&model.Message{}).
				Select("content, created_at").
				Where("type = 1 AND ((from_user_id = ? AND to_user_id = ? AND deleted_by_sender = 0) OR (from_user_id = ? AND to_user_id = ? AND deleted_by_receiver = 0))", oid, userID, userID, oid).
				Order("created_at desc").
				First(&last)
			// 未读消息数（对方发给当前用户且未读）
			var unread int64
			database.DB.Model(&model.Message{}).
				Where("from_user_id = ? AND to_user_id = ? AND is_read = 0 AND type = 1 AND deleted_by_receiver = 0", oid, userID).
				Count(&unread)
			items = append(items, ChatItemVO{
				ID:          oid,
				Type:        "user",
				Name:        u.Nickname,
				AvatarURL:   u.AvatarURL,
				LastContent: last.Content,
				LastTime:    last.CreatedAt,
				UnreadCount: unread,
			})
		}
	}

	// 2. 群聊会话（群聊名称取关联搭子的标题，最后消息取自 conversation_messages 表）
	var convs []ChatItemVO
	// 注意：Select 子查询中的 userID 直接拼接（GORM 不支持 Select 内 ? 参数绑定；userID 为服务端生成的雪花ID，无注入风险）
	err := database.DB.Table("conversations c").
		Select("c.id, COALESCE(p.title, c.name) AS name, c.partner_id, " +
			"(SELECT COUNT(*) FROM conversation_members cm WHERE cm.conversation_id = c.id AND cm.deleted_at IS NULL) as member_count, " +
			"(SELECT content FROM conversation_messages cm3 WHERE cm3.conversation_id = c.id ORDER BY cm3.created_at DESC LIMIT 1) as last_content, " +
			"(SELECT created_at FROM conversation_messages cm3 WHERE cm3.conversation_id = c.id ORDER BY cm3.created_at DESC LIMIT 1) as last_time, " +
			"COALESCE((SELECT COUNT(*) FROM conversation_messages cm4 WHERE cm4.conversation_id = c.id AND cm4.from_user_id != '" + userID + "' " +
			"AND cm4.created_at > COALESCE((SELECT cr.last_read_at FROM conversation_reads cr WHERE cr.conversation_id = c.id AND cr.user_id = '" + userID + "'), '2000-01-01 00:00:00')), 0) as unread_count").
		Joins("LEFT JOIN partners p ON p.id = c.partner_id").
		Joins("JOIN conversation_members cm2 ON cm2.conversation_id = c.id AND cm2.user_id = '" + userID + "' AND cm2.deleted_at IS NULL").
		Where("c.deleted_at IS NULL").
		Scan(&convs).Error
	if err != nil {
		return nil, err
	}
	for i := range convs {
		convs[i].Type = "group"
		if convs[i].Name != "" {
			convs[i].Name += "群聊"
		}
		items = append(items, convs[i])
	}

	// 3. 系统消息（最新一条通知，无通知记录时不显示）
	var latestNoti model.Notification
	database.DB.Where("user_id = ? AND type = 4", userID).
		Order("created_at desc").First(&latestNoti)
	if latestNoti.ID != "" {
		var unreadSys int64
		database.DB.Model(&model.Notification{}).
			Where("user_id = ? AND type = 4 AND is_read = 0", userID).Count(&unreadSys)
		items = append(items, ChatItemVO{
			ID:          "system",
			Type:        "system",
			Name:        "公告",
			LastContent: latestNoti.Content,
			LastTime:    latestNoti.CreatedAt,
			UnreadCount: unreadSys,
		})
	}

	// 按最后消息时间倒序（无消息的群聊 last_time 为零值，自然排最后）
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].LastTime.After(items[j].LastTime)
	})
	return items, nil
}

// CreateConversationMessage 发送群聊消息
func CreateConversationMessage(msg *model.ConversationMessage) error {
	return database.DB.Create(msg).Error
}

// ConversationMessageVO 群聊消息视图（含发送者信息）
type ConversationMessageVO struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversationId"` // 群聊ID
	FromUserID     string    `json:"fromUserId"`     // 发送者ID
	Nickname       string    `json:"nickname"`       // 发送者昵称
	AvatarURL      string    `json:"avatarUrl"`      // 发送者头像
	Content        string    `json:"content"`        // 消息内容
	CreatedAt      time.Time `json:"createdAt"`      // 发送时间
}

// MarkConversationRead 标记群聊已读（记录该会话最新消息时间作为已读游标）
func MarkConversationRead(convID, userID string) error {
	var lastMsg struct{ CreatedAt time.Time }
	if err := database.DB.Model(&model.ConversationMessage{}).
		Where("conversation_id = ?", convID).
		Order("created_at desc").First(&lastMsg).Error; err != nil {
		return nil // 无消息无需标记
	}
	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "conversation_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_read_at"}),
	}).Create(&model.ConversationRead{
		ConversationID: convID,
		UserID:         userID,
		LastReadAt:     lastMsg.CreatedAt,
	}).Error
}

// GetConversationMessages 分页获取群聊消息（最新一页，按时间正序返回）；拉取消息即标记该会话已读
func GetConversationMessages(convID, userID string, page, pageSize int) ([]ConversationMessageVO, int64, error) {
	// 拉取消息视为已读：记录已读游标，会话列表未读数随之清零
	_ = MarkConversationRead(convID, userID)
	var total int64
	database.DB.Model(&model.ConversationMessage{}).
		Where("conversation_id = ?", convID).Count(&total)
	offset := (page - 1) * pageSize
	var list []ConversationMessageVO
	err := database.DB.Table("conversation_messages cm").
		Select("cm.id, cm.conversation_id, cm.from_user_id, u.nickname, u.avatar_url, cm.content, cm.created_at").
		Joins("LEFT JOIN users u ON u.id = cm.from_user_id").
		Where("cm.conversation_id = ?", convID).
		Order("cm.created_at desc").
		Offset(offset).Limit(pageSize).
		Scan(&list).Error
	// 倒序取页后反转成正序，保证聊天记录从旧到新展示
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, total, err
}

// ErrConversationNotFound 群聊不存在
var ErrConversationNotFound = errors.New("群聊不存在")
