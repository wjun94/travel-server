package miniapp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/database"
	"travel-server/pkg/response"
)

// MarkNotificationRead 标记单条通知为已读/未读
// @Summary 标记已读/未读
// @Security BearerAuth
// @Tags 小程序-通知
// @Param id path string true "通知ID"
// @Param isRead query int false "0未读 1已读（默认1）"
// @Success 200 {object} response.Response
// @Router /api/v1/notification/read/{id} [put]
func MarkNotificationRead(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(string)
	isRead, _ := strconv.Atoi(c.DefaultQuery("isRead", "1"))
	if err := repository.MarkNotificationRead(id, userID, isRead); err != nil {
		response.Fail(c, 404, err.Error())
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

// MarkTypeNotificationRead 按类型清空未读数（点击tab时调用）
// @Summary 按类型清空未读
// @Description 标记指定类型的所有通知为已读：1搭子申请 2点赞 3新增关注 4系统通知 5评论（新增评论/评论点赞）
// @Security BearerAuth
// @Tags 小程序-通知
// @Param type query int true "通知类型：1搭子申请 2点赞 3新增关注 4系统通知 5评论"
// @Success 200 {object} response.Response
// @Router /api/v1/notification/type-read [put]
func MarkTypeNotificationRead(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	// 兼容 query 和 body 两种传参方式
	notiType := 0
	if t := c.Query("type"); t != "" {
		notiType, _ = strconv.Atoi(t)
	} else {
		var req struct {
			Type int `json:"type"`
		}
		_ = c.ShouldBindJSON(&req)
		notiType = req.Type
	}
	if notiType < 1 || notiType > 5 {
		response.Fail(c, 400, "type参数无效")
		return
	}
	if err := repository.MarkTypeNotificationsRead(userID, notiType); err != nil {
		response.Fail(c, 500, "操作失败")
		return
	}
	response.Success(c, nil)
}

// ClearSystemNotifications 清空系统通知（type=4）
// @Summary 清空系统通知
// @Security BearerAuth
// @Tags 小程序-通知
// @Success 200 {object} response.Response
// @Router /api/v1/notification/system [delete]
func ClearSystemNotifications(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	if err := repository.DeleteSystemNotifications(userID); err != nil {
		response.Fail(c, 500, "清空失败")
		return
	}
	response.Success(c, nil)
}

// GetNotificationList 分页获取通知列表（type=0 全部，1搭子申请，2攻略点赞，3新增关注，4系统通知，5评论点赞）
// @Summary 通知列表
// @Security BearerAuth
// @Tags 小程序-通知
// @Param type query int false "通知类型：0全部 1搭子申请 2攻略点赞 3新增关注 4系统通知 5评论点赞"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]object{id=string,userId=string,fromUserId=string,type=int,relatedId=string,targetId=string,targetType=string,isRead=int,content=string,createdAt=string,fromUser=object{id=string,nickname=string,avatarUrl=string},commentContent=string,remark=string,status=int,reason=string},total=int64}}
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

	// 批量收集关联ID
	fromIDs := make(map[string]struct{})
	commentIDs := make(map[string]struct{})
	appIDs := make(map[string]struct{})
	for _, n := range list {
		if n.FromUserID != "" {
			fromIDs[n.FromUserID] = struct{}{}
		}
		// 即便 FromUserID 为空，也从相关记录中收集用户ID用于回退
		if n.RelatedID == "" {
			continue
		}
		switch n.Type {
		case 5:
			commentIDs[n.RelatedID] = struct{}{}
		case 1:
			appIDs[n.RelatedID] = struct{}{}
		case 3:
			// RelatedID 即关注者ID，加入用户查询集合
			fromIDs[n.RelatedID] = struct{}{}
		}
	}

	// 批量查用户
	userMap := make(map[string]model.User)
	if len(fromIDs) > 0 {
		ids := make([]string, 0, len(fromIDs))
		for id := range fromIDs {
			ids = append(ids, id)
		}
		var users []model.User
		database.DB.Where("id IN ?", ids).Find(&users)
		for i := range users {
			userMap[users[i].ID] = users[i]
		}
	}

	// 批量查评论（获取原内容、所属目标ID/类型、作者）
	type commentInfo struct {
		ID         string
		UserID     string
		Content    string
		TargetType string
		TargetID   string
	}
	commentMap := make(map[string]commentInfo)
	if len(commentIDs) > 0 {
		ids := make([]string, 0, len(commentIDs))
		for id := range commentIDs {
			ids = append(ids, id)
		}
		var comments []commentInfo
		database.DB.Model(&model.Comment{}).Select("id, user_id, content, target_type, target_id").Where("id IN ?", ids).Find(&comments)
		for _, c := range comments {
			commentMap[c.ID] = c
			// 评论作者也加入用户查询集合
			if c.UserID != "" {
				fromIDs[c.UserID] = struct{}{}
			}
		}
	}

	// 批量查搭子申请（获取所属搭子ID、申请人、申请备注、状态和拒绝理由）
	appMap := make(map[string]string)       // applicationID → partnerID
	appUserMap := make(map[string]string)   // applicationID → applicantID
	appRemarkMap := make(map[string]string) // applicationID → remark
	appStatusMap := make(map[string]int)    // applicationID → status(0待审核 1通过 2拒绝 3主动退出)
	appRejectMap := make(map[string]string) // applicationID → rejectReason
	if len(appIDs) > 0 {
		ids := make([]string, 0, len(appIDs))
		for id := range appIDs {
			ids = append(ids, id)
		}
		var apps []struct {
			ID           string
			PartnerID    string
			UserID       string
			Remark       string
			Status       int
			RejectReason string
		}
		database.DB.Model(&model.PartnerApplication{}).Select("id, partner_id, user_id, remark, status, reject_reason").Where("id IN ?", ids).Find(&apps)
		for _, a := range apps {
			appMap[a.ID] = a.PartnerID
			appUserMap[a.ID] = a.UserID
			appRemarkMap[a.ID] = a.Remark
			appStatusMap[a.ID] = a.Status
			appRejectMap[a.ID] = a.RejectReason
			if a.UserID != "" {
				fromIDs[a.UserID] = struct{}{}
			}
		}
	}

	// 重新查询一遍用户（回退ID已追加到 fromIDs）
	userMap = make(map[string]model.User)
	if len(fromIDs) > 0 {
		ids := make([]string, 0, len(fromIDs))
		for id := range fromIDs {
			ids = append(ids, id)
		}
		var users []model.User
		database.DB.Where("id IN ?", ids).Find(&users)
		for i := range users {
			userMap[users[i].ID] = users[i]
		}
	}

	// 构建富化响应
	type fromUserVO struct {
		ID        string `json:"id"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarUrl"`
	}
	type itemVO struct {
		ID             string      `json:"id"`
		UserID         string      `json:"userId"`                   // 通知接收者ID
		FromUserID     string      `json:"fromUserId"`               // 通知触发者ID
		Type           int         `json:"type"`                     // 1搭子申请 2攻略点赞 3新增关注 4系统通知 5评论点赞
		RelatedID      string      `json:"relatedId"`                // 原始关联单据ID
		TargetID       string      `json:"targetId"`                 // 用于前端跳转的目标ID
		TargetType     string      `json:"targetType"`               // 跳转目标类型：guide/trip/partner/user
		IsRead         int         `json:"isRead"`                   // 0未读 1已读
		Content        string      `json:"content"`                  // 通知摘要文字
		CreatedAt      string      `json:"createdAt"`                // 创建时间（ISO8601）
		FromUser       *fromUserVO `json:"fromUser"`                 // 触发者信息，null表示无触发者
		CommentContent string      `json:"commentContent,omitempty"` // type=5 时的原评论内容，已删除则显示"原评论已删除"
		Remark         string      `json:"remark,omitempty"`         // type=1 搭子申请的申请备注
		Status         int         `json:"status"`                   // type=1 搭子申请状态：0待审核 1通过 2拒绝 3主动退出（不用omitempty，保证待审核0也返回）
		Reason         string      `json:"reason,omitempty"`         // type=1 拒绝时的拒绝理由
	}
	items := make([]itemVO, 0, len(list))
	for _, n := range list {
		// 确定 fromUser 的真实用户ID（优先 FromUserID，再按类型回退）
		fromUserID := n.FromUserID
		if fromUserID == "" {
			switch n.Type {
			case 5:
				if ci, ok := commentMap[n.RelatedID]; ok {
					fromUserID = ci.UserID
				}
			case 3:
				fromUserID = n.RelatedID // RelatedID = followerID
			case 1:
				if uid, ok := appUserMap[n.RelatedID]; ok {
					fromUserID = uid
				}
			}
		}
		var fu *fromUserVO
		if u, ok := userMap[fromUserID]; ok {
			fu = &fromUserVO{
				ID:        u.ID,
				Nickname:  u.Nickname,
				AvatarURL: u.AvatarURL,
			}
		}

		// 计算跳转用的 targetId / targetType
		targetID := n.RelatedID
		targetType := ""
		switch n.Type {
		case 1:
			// 搭子申请 → 跳转到搭子详情页
			if pid, ok := appMap[n.RelatedID]; ok {
				targetID = pid
			}
			targetType = "partner"
		case 2:
			// 攻略点赞 → 跳转到攻略详情
			targetType = "guide"
		case 3:
			// 新增关注 → 跳转到用户主页
			targetType = "user"
		case 5:
			// 评论点赞 → 跳转到评论所属内容页
			if ci, ok := commentMap[n.RelatedID]; ok {
				targetID = ci.TargetID
				targetType = ci.TargetType
			}
		}

		cc := ""
		if n.Type == 5 {
			if ci, ok := commentMap[n.RelatedID]; ok {
				cc = ci.Content
			} else {
				cc = "原评论已删除"
			}
		}

		// 搭子申请的备注/状态/拒绝理由
		remark := ""
		appStatus := 0
		rejectReason := ""
		if n.Type == 1 {
			remark = appRemarkMap[n.RelatedID]
			appStatus = appStatusMap[n.RelatedID]
			rejectReason = appRejectMap[n.RelatedID]
		}

		items = append(items, itemVO{
			ID:             n.ID,
			UserID:         n.UserID,
			FromUserID:     n.FromUserID,
			Type:           n.Type,
			RelatedID:      n.RelatedID,
			TargetID:       targetID,
			TargetType:     targetType,
			IsRead:         n.IsRead,
			Content:        n.Content,
			CreatedAt:      n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			FromUser:       fu,
			CommentContent: cc,
			Remark:         remark,
			Status:         appStatus,
			Reason:         rejectReason,
		})
	}

	response.Success(c, gin.H{"list": items, "total": total})
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
