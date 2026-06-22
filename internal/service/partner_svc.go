package service

import (
	"errors"

	"travel-server/internal/repository"
)

// ApproveApplication 同意搭子申请，更新成员数
func ApproveApplication(appID uint) error {
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
	return nil
}
