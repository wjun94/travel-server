package service

import (
	"errors"
	"log"

	"travel-server/internal/model"
	"travel-server/internal/repository"
)

// ApproveApplication 同意搭子申请，更新成员数、自动创建/加入群聊并通知申请人
func ApproveApplication(appID string) error {
	app, err := repository.GetApplicationByID(appID)
	if err != nil {
		return err
	}
	if app.Status != 0 {
		return errors.New("申请已被处理")
	}
	// 更新申请状态为同意
	if err := repository.UpdateApplicationStatus(appID, 1); err != nil {
		return err
	}
	// 增加搭子当前人数
	partner, err := repository.GetPartnerByID(app.PartnerID)
	if err != nil {
		return err
	}
	if partner.CurrentMembers < partner.MaxMembers {
		partner.CurrentMembers++
		repository.UpdatePartner(partner)
	}
	// 自动创建群聊（不存在则创建并加入群主），并把申请人加入（失败不影响主流程）
	if err := ensurePartnerConversation(partner, app.UserID); err != nil {
		log.Printf("创建搭子群聊失败: %v", err)
	}
	// 通知申请人（通知失败不影响主流程）
	if err := repository.CreateNotification(&model.Notification{
		UserID:     app.UserID,
		FromUserID: partner.UserID,
		Type:       1,
		RelatedID:  appID,
		Content:    "您的搭子申请已通过",
	}); err != nil {
		log.Printf("创建搭子申请通过通知失败: %v", err)
	}
	return nil
}

// ensurePartnerConversation 确保搭子群聊存在，并把用户加入（群主在创建时加入）
func ensurePartnerConversation(partner *model.Partner, userID string) error {
	conv, err := repository.GetConversationByPartnerID(partner.ID)
	if err != nil {
		// 群聊不存在，创建（群主=搭子创建者，名称=搭子标题）
		conv = &model.Conversation{
			PartnerID: partner.ID,
			Name:      partner.Title,
			OwnerID:   partner.UserID,
		}
		if err := repository.CreateConversation(conv); err != nil {
			return err
		}
		// 群主加入
		if err := repository.AddConversationMember(conv.ID, partner.UserID); err != nil {
			return err
		}
	}
	// 申请人加入
	return repository.AddConversationMember(conv.ID, userID)
}

// RejectApplication 拒绝搭子申请，保存理由并通知申请人
func RejectApplication(appID, reason string) error {
	app, err := repository.GetApplicationByID(appID)
	if err != nil {
		return err
	}
	if app.Status != 0 {
		return errors.New("申请已被处理")
	}
	if err := repository.RejectApplication(appID, reason); err != nil {
		return err
	}
	// 通知申请人（通知失败不影响主流程）
	partner, err := repository.GetPartnerByID(app.PartnerID)
	if err != nil {
		return err
	}
	if err := repository.CreateNotification(&model.Notification{
		UserID:     app.UserID,
		FromUserID: partner.UserID,
		Type:       1,
		RelatedID:  appID,
		Content:    "您的搭子申请已被拒绝",
	}); err != nil {
		log.Printf("创建搭子申请拒绝通知失败: %v", err)
	}
	return nil
}
