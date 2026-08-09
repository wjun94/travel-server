package repository

import (
	"time"

	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
)

// CreatePartner 发布搭子
func CreatePartner(p *model.Partner) error {
	return database.DB.Create(p).Error
}

// CountTodayAIPartners 统计用户今日AI生成的搭子数
func CountTodayAIPartners(userID string) (int64, error) {
	var count int64
	start := time.Now().Truncate(24 * time.Hour)
	err := database.DB.Model(&model.Partner{}).
		Where("user_id = ? AND is_ai = 1 AND created_at >= ?", userID, start).
		Count(&count).Error
	return count, err
}

// GetPartnerList 分页获取搭子列表，支持关键词搜索（标题/目的地/简述/标签）
// 只返回可申请加入的搭子（0招募中/1满员可候补），已解散/已过期/已下架不可见
func GetPartnerList(page, pageSize int, keyword string) ([]model.Partner, int64, error) {
	var list []model.Partner
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Partner{}).Where("status IN (0, 1)") // 可申请加入：招募中/满员（满员可申请候补）
	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR destination LIKE ? OR `desc` LIKE ? OR tags LIKE ?", kw, kw, kw, kw)
	}
	query.Count(&total)
	err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error
	return list, total, err
}

// GetPartnerByID 根据 ID 获取搭子
func GetPartnerByID(id string) (*model.Partner, error) {
	var p model.Partner
	err := database.DB.First(&p, "id = ?", id).Error
	return &p, err
}

// UpdatePartner 更新搭子信息
func UpdatePartner(p *model.Partner) error {
	return database.DB.Save(p).Error
}

// CreateApplication 创建搭子申请
func CreateApplication(app *model.PartnerApplication) error {
	return database.DB.Create(app).Error
}

// GetApplicationByID 获取申请详情
func GetApplicationByID(id string) (*model.PartnerApplication, error) {
	var app model.PartnerApplication
	err := database.DB.First(&app, "id = ?", id).Error
	return &app, err
}

// UpdateApplicationStatus 修改申请状态
func UpdateApplicationStatus(id string, status int) error {
	return database.DB.Model(&model.PartnerApplication{}).Where("id = ?", id).Update("status", status).Error
}

// UpdatePartnerStatus 修改搭子状态（后台审核：0恢复招募 1满员 2已取消 3已过期 4已下架）
func UpdatePartnerStatus(id string, status int) error {
	return database.DB.Model(&model.Partner{}).Where("id = ?", id).Update("status", status).Error
}

// RejectApplication 拒绝申请并保存拒绝理由
func RejectApplication(id, reason string) error {
	return database.DB.Model(&model.PartnerApplication{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": 2, "reject_reason": reason}).Error
}

// GetMyApplication 获取当前用户对指定搭子的最新一条申请记录（含状态和拒绝理由）
func GetMyApplication(userID, partnerID string) (*model.PartnerApplication, error) {
	var app model.PartnerApplication
	err := database.DB.Where("user_id = ? AND partner_id = ?", userID, partnerID).
		Order("created_at DESC").First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// GetUserAppliedPartnerIDs 获取当前用户已申请过的搭子ID集合（待审核+通过）
func GetUserAppliedPartnerIDs(userID string, partnerIDs []string) (map[string]bool, error) {
	var apps []model.PartnerApplication
	result := make(map[string]bool, len(partnerIDs))
	if len(partnerIDs) == 0 {
		return result, nil
	}
	if err := database.DB.Model(&model.PartnerApplication{}).
		Select("partner_id").
		Where("user_id = ? AND partner_id IN ? AND status IN (0, 1)", userID, partnerIDs).
		Find(&apps).Error; err != nil {
		return nil, err
	}
	for _, app := range apps {
		result[app.PartnerID] = true
	}
	return result, nil
}

// GetPartners 搭子列表（分页，支持目的地/状态/类型筛选；type -1 全部 0用户 1官方）
func GetPartners(page, pageSize int, destination string, status int, ptype int) ([]model.Partner, int64, error) {
	var list []model.Partner
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Partner{})
	if ptype >= 0 {
		query = query.Where("type = ?", ptype)
	}
	if destination != "" {
		query = query.Where("destination LIKE ?", "%"+destination+"%")
	}
	// status 仅筛选有效值（-1 或未传表示全部）
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error
	return list, total, err
}

// GetMyPartners 我发布的搭子列表（isDraft>=0 时按草稿状态筛选，-1 返回全部）
func GetMyPartners(userID string, page, pageSize, isDraft int) ([]model.Partner, int64, error) {
	var list []model.Partner
	var total int64
	offset := (page - 1) * pageSize
	query := database.DB.Model(&model.Partner{}).Where("user_id = ?", userID)
	if isDraft >= 0 {
		query = query.Where("is_draft = ?", isDraft)
	}
	query.Count(&total)
	listQuery := database.DB.Where("user_id = ?", userID)
	if isDraft >= 0 {
		listQuery = listQuery.Where("is_draft = ?", isDraft)
	}
	err := listQuery.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error
	return list, total, err
}

// GetJoinedPartners 我参与的搭子列表（申请已通过且搭子已发布，按发布时间倒序）
func GetJoinedPartners(userID string, page, pageSize int) ([]model.Partner, int64, error) {
	var list []model.Partner
	var total int64
	query := database.DB.Model(&model.Partner{}).
		Joins("JOIN partner_applications pa ON pa.partner_id = partners.id").
		Where("pa.user_id = ? AND pa.status = 1 AND partners.is_draft = 0", userID)
	query.Count(&total)
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("partners.created_at desc").Find(&list).Error
	return list, total, err
}

// CountJoinedPartners 统计我参与的搭子数（申请已通过且搭子已发布）
func CountJoinedPartners(userID string) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Partner{}).
		Joins("JOIN partner_applications pa ON pa.partner_id = partners.id").
		Where("pa.user_id = ? AND pa.status = 1 AND partners.is_draft = 0", userID).
		Count(&count).Error
	return count, err
}

// DeletePartnerCascade 删除搭子（级联删除申请记录）
func DeletePartnerCascade(id string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("partner_id = ?", id).Delete(&model.PartnerApplication{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.Partner{}).Error
	})
}

// CancelPartner 发起人解散搭子（状态置为2，保存解散原因），并拒绝所有待审核申请
func CancelPartner(id, userID, reason string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 拒绝待审核申请
		if err := tx.Model(&model.PartnerApplication{}).
			Where("partner_id = ? AND status = 0", id).
			Update("status", 2).Error; err != nil {
			return err
		}
		// 更新搭子状态 + 解散原因
		return tx.Model(&model.Partner{}).
			Where("id = ? AND user_id = ?", id, userID).
			Updates(map[string]interface{}{"status": 2, "cancel_reason": reason}).Error
	})
}

// GetPartnerMemberIDs 获取搭子已加入成员（申请已通过）的用户ID列表
func GetPartnerMemberIDs(partnerID string) ([]string, error) {
	var ids []string
	err := database.DB.Model(&model.PartnerApplication{}).
		Where("partner_id = ? AND status = 1", partnerID).
		Pluck("user_id", &ids).Error
	return ids, err
}

// GetEarliestPendingApplication 获取搭子最早的待审核申请
func GetEarliestPendingApplication(partnerID string) (*model.PartnerApplication, error) {
	var app model.PartnerApplication
	err := database.DB.Where("partner_id = ? AND status = 0", partnerID).
		Order("created_at asc").First(&app).Error
	return &app, err
}

// GetApplicationByPartnerAndUser 获取用户在该搭子的已通过申请（判断是否已加入）
func GetApplicationByPartnerAndUser(partnerID, userID string) (*model.PartnerApplication, error) {
	var app model.PartnerApplication
	err := database.DB.Where("partner_id = ? AND user_id = ? AND status = 1", partnerID, userID).
		First(&app).Error
	return &app, err
}

// LeavePartner 成员退出搭子（申请置为3主动退出，人数减一；原满员则恢复招募中）
func LeavePartner(partnerID, userID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 申请状态置为主动退出
		if err := tx.Model(&model.PartnerApplication{}).
			Where("partner_id = ? AND user_id = ? AND status = 1", partnerID, userID).
			Update("status", 3).Error; err != nil {
			return err
		}
		// 已通过人数减一
		if err := tx.Model(&model.Partner{}).
			Where("id = ? AND current_members > 0", partnerID).
			UpdateColumn("current_members", gorm.Expr("current_members - 1")).Error; err != nil {
			return err
		}
		// 满员状态下有人退出 → 恢复招募中（有候补时由调用方自动补位）
		return tx.Model(&model.Partner{}).
			Where("id = ? AND status = 1 AND current_members < max_members", partnerID).
			Update("status", 0).Error
	})
}

// AutoCloseExpiredPartners 系统自动关闭过期未满员搭子（状态置为3），返回受影响条数
// 条件：status=0、StartDate < now、MinMembers > 0、CurrentMembers < MinMembers
func AutoCloseExpiredPartners() int64 {
	now := time.Now()
	result := database.DB.Model(&model.Partner{}).
		Where("status = 0 AND start_date IS NOT NULL AND start_date < ? AND min_members > 0 AND current_members < min_members", now).
		Update("status", 3)
	return result.RowsAffected
}

// LikePartner 点赞搭子（点赞独立记录在 partner_likes 表，与收藏分离）
func LikePartner(userID, partnerID string) error {
	var count int64
	database.DB.Model(&model.PartnerLike{}).
		Where("user_id = ? AND partner_id = ?", userID, partnerID).
		Count(&count)
	if count > 0 {
		return nil
	}
	tx := database.DB.Begin()
	if err := tx.Create(&model.PartnerLike{
		UserID:    userID,
		PartnerID: partnerID,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&model.Partner{}).
		Where("id = ?", partnerID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// UnlikePartner 取消点赞搭子
func UnlikePartner(userID, partnerID string) error {
	tx := database.DB.Begin()
	r := tx.Where("user_id = ? AND partner_id = ?", userID, partnerID).
		Delete(&model.PartnerLike{})
	if r.Error != nil {
		tx.Rollback()
		return r.Error
	}
	if r.RowsAffected == 0 {
		tx.Rollback()
		return nil
	}
	if err := tx.Model(&model.Partner{}).
		Where("id = ?", partnerID).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// IsPartnerLiked 当前用户是否已点赞该搭子
func IsPartnerLiked(userID, partnerID string) bool {
	var count int64
	database.DB.Model(&model.PartnerLike{}).
		Where("user_id = ? AND partner_id = ?", userID, partnerID).
		Count(&count)
	return count > 0
}

// IncrementPartnerViewCount 增加搭子浏览量
func IncrementPartnerViewCount(id string) error {
	return database.DB.Model(&model.Partner{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// GetPartnerCommentCounts 批量查询多个搭子的评论数
func GetPartnerCommentCounts(partnerIDs []string) map[string]int64 {
	if len(partnerIDs) == 0 {
		return nil
	}
	type result struct {
		TargetID string
		Count    int64
	}
	var rows []result
	database.DB.Model(&model.Comment{}).
		Select("target_id, COUNT(*) as count").
		Where("target_type = ? AND target_id IN ?", "partner", partnerIDs).
		Group("target_id").
		Find(&rows)
	m := make(map[string]int64, len(rows))
	for _, r := range rows {
		m[r.TargetID] = r.Count
	}
	return m
}
