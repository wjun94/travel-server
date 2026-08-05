package repository

import (
	"errors"
	"time"

	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// GetAccountsByTarget 获取某目标的记账记录（按消费时间倒序）
func GetAccountsByTarget(targetType, targetID, userID string) ([]model.Accounting, error) {
	var accounts []model.Accounting
	err := database.DB.Where("target_type = ? AND target_id = ? AND user_id = ?", targetType, targetID, userID).
		Order("consumed_at desc, created_at desc").Find(&accounts).Error
	return accounts, err
}

// CreateAccount 添加记账条目
func CreateAccount(acc *model.Accounting) error {
	return database.DB.Create(acc).Error
}

// DeleteAccount 删除记账条目（仅本人）
func DeleteAccount(id, userID string) error {
	res := database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Accounting{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrAccountNotFound
	}
	return nil
}

// UpdateAccount 编辑记账条目（仅本人，可改分类/金额/备注/消费时间）
func UpdateAccount(id, userID string, updates map[string]interface{}) error {
	res := database.DB.Model(&model.Accounting{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrAccountNotFound
	}
	return nil
}

// AccountSummary 账本汇总
type AccountSummary struct {
	TotalAmount  float64            `json:"totalAmount"`  // 总支出
	Count        int64              `json:"count"`        // 总笔数
	CategoryStat map[string]float64 `json:"categoryStat"` // 各分类金额
}

// GetAccountSummary 获取某目标的账本汇总
func GetAccountSummary(targetType, targetID, userID string) (AccountSummary, error) {
	var summary AccountSummary
	var rows []struct {
		Category string
		Sum      float64
		Count    int64
	}
	err := database.DB.Model(&model.Accounting{}).
		Select("category, SUM(amount) as sum, COUNT(*) as count").
		Where("target_type = ? AND target_id = ? AND user_id = ?", targetType, targetID, userID).
		Group("category").Scan(&rows).Error
	if err != nil {
		return summary, err
	}
	summary.CategoryStat = make(map[string]float64, len(rows))
	for _, r := range rows {
		summary.TotalAmount += r.Sum
		summary.Count += r.Count
		summary.CategoryStat[r.Category] = r.Sum
	}
	return summary, nil
}

// AccountOverviewItem 账本总览项（按目标聚合）
type AccountOverviewItem struct {
	TargetType  string    `json:"targetType"`  // 绑定类型
	TargetID    string    `json:"targetId"`    // 绑定目标ID
	TargetName  string    `json:"targetName"`  // 目标名称（行程标题/攻略标题/搭子标题）
	TotalAmount float64   `json:"totalAmount"` // 总支出
	Count       int64     `json:"count"`       // 总笔数
	LastTime    time.Time `json:"lastTime"`    // 最后记账时间
}

// GetAccountOverview 获取我的账本总览（按目标聚合，按最后记账时间倒序）
func GetAccountOverview(userID string) ([]AccountOverviewItem, error) {
	type aggRow struct {
		TargetType  string
		TargetID    string
		TargetName  string
		TotalAmount float64
		Count       int64
		LastTime    time.Time
	}
	var aggs []aggRow
	err := database.DB.Model(&model.Accounting{}).
		Select("target_type, target_id, MAX(target_name) as target_name, SUM(amount) as total_amount, COUNT(*) as count, MAX(consumed_at) as last_time").
		Where("user_id = ? AND target_type IN (?) AND target_id != ''", userID, []string{"trip", "guide", "partner", "custom"}).
		Group("target_type, target_id").Scan(&aggs).Error
	if err != nil {
		return nil, err
	}

	// 批量取绑定目标名称（自主账本直接用存储的账本名）
	tripIDs, guideIDs, partnerIDs := []string{}, []string{}, []string{}
	for _, a := range aggs {
		switch a.TargetType {
		case "trip":
			tripIDs = append(tripIDs, a.TargetID)
		case "guide":
			guideIDs = append(guideIDs, a.TargetID)
		case "partner":
			partnerIDs = append(partnerIDs, a.TargetID)
		}
	}
	nameMap := make(map[string]string, len(aggs))
	for _, a := range aggs {
		if a.TargetType == "custom" && a.TargetName != "" {
			nameMap["custom:"+a.TargetID] = a.TargetName
		}
	}
	if len(tripIDs) > 0 {
		var trips []model.Trip
		database.DB.Select("id, title").Where("id IN ?", tripIDs).Find(&trips)
		for _, t := range trips {
			nameMap["trip:"+t.ID] = t.Title
		}
	}
	if len(guideIDs) > 0 {
		var guides []model.Guide
		database.DB.Select("id, title").Where("id IN ?", guideIDs).Find(&guides)
		for _, g := range guides {
			nameMap["guide:"+g.ID] = g.Title
		}
	}
	if len(partnerIDs) > 0 {
		var partners []model.Partner
		database.DB.Select("id, title").Where("id IN ?", partnerIDs).Find(&partners)
		for _, p := range partners {
			nameMap["partner:"+p.ID] = p.Title
		}
	}

	items := make([]AccountOverviewItem, 0, len(aggs))
	for _, a := range aggs {
		items = append(items, AccountOverviewItem{
			TargetType:  a.TargetType,
			TargetID:    a.TargetID,
			TargetName:  nameMap[a.TargetType+":"+a.TargetID],
			TotalAmount: a.TotalAmount,
			Count:       a.Count,
			LastTime:    a.LastTime,
		})
	}
	// 按最后记账时间倒序
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].LastTime.After(items[j-1].LastTime); j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
	return items, nil
}

// DeleteAccountBook 删除整本账本（该目标下的所有记账条目，仅本人）。
// 账本下无记录时视为已删除，返回成功（幂等删除）。
func DeleteAccountBook(targetType, targetID, userID string) error {
	res := database.DB.Where("target_type = ? AND target_id = ? AND user_id = ?", targetType, targetID, userID).
		Delete(&model.Accounting{})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// ErrAccountNotFound 记账条目不存在
var ErrAccountNotFound = errors.New("记账条目不存在")
