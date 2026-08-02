package service

import (
	"errors"
	"log"

	"travel-server/internal/model"
	"travel-server/internal/repository"
)

// ApproveApplication 同意搭子申请，更新成员数并通知申请人
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
